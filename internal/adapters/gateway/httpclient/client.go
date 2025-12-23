package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"draftea-challenge/internal/domain/errors"
	"draftea-challenge/internal/domain/payment"
)

// Client implements the payment gateway using HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
	retries    int
	backoff    backoffConfig
	breaker    *circuitBreaker
	semaphore  chan struct{}
}

type backoffConfig struct {
	initial time.Duration
	max     time.Duration
}

// Config defines gateway client resilience settings.
type Config struct {
	BaseURL                string
	Timeout                time.Duration
	MaxRetries             int
	RetryInitialBackoff    time.Duration
	RetryMaxBackoff        time.Duration
	CircuitBreakerFailures int
	CircuitBreakerCooldown time.Duration
	MaxInFlight            int
}

// New creates a new gateway client.
func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	initialBackoff := cfg.RetryInitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = 200 * time.Millisecond
	}
	maxBackoff := cfg.RetryMaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 2 * time.Second
	}
	cbFailures := cfg.CircuitBreakerFailures
	if cbFailures <= 0 {
		cbFailures = 5
	}
	cbCooldown := cfg.CircuitBreakerCooldown
	if cbCooldown <= 0 {
		cbCooldown = 10 * time.Second
	}
	maxInFlight := cfg.MaxInFlight
	if maxInFlight <= 0 {
		maxInFlight = 20
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retries: maxRetries,
		backoff: backoffConfig{
			initial: initialBackoff,
			max:     maxBackoff,
		},
		breaker:   newCircuitBreaker(cbFailures, cbCooldown),
		semaphore: make(chan struct{}, maxInFlight),
	}
}

type gatewayRequest struct {
	ProviderID        string `json:"provider_id"`
	ExternalReference string `json:"external_reference"`
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
}

type gatewayResponse struct {
	Status string `json:"status"`
}

// ProcessPayment calls the external gateway.
func (c *Client) ProcessPayment(ctx context.Context, p *payment.Payment) (string, error) {
	if !c.breaker.allow() {
		return "", errors.NewGatewayError("gateway circuit breaker open")
	}

	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-ctx.Done():
		return "", errors.NewGatewayTimeoutError("gateway timeout")
	}

	payload := gatewayRequest{
		ProviderID:        p.ProviderID.String(),
		ExternalReference: p.ExternalReference,
		Amount:            p.Amount,
		Currency:          p.Currency,
	}

	var lastErr error
	backoff := c.backoff.initial

	for attempt := 0; attempt <= c.retries; attempt++ {
		body, err := json.Marshal(payload)
		if err != nil {
			return "", errors.NewInternalError("failed to encode gateway request")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/pay", c.baseURL), bytes.NewReader(body))
		if err != nil {
			return "", errors.NewInternalError("failed to create gateway request")
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = errors.NewGatewayTimeoutError("gateway timeout")
		} else {
			defer resp.Body.Close()
			var out gatewayResponse
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				lastErr = errors.NewGatewayError("invalid gateway response")
			} else {
				switch resp.StatusCode {
				case http.StatusOK:
					c.breaker.success()
					return out.Status, nil
				case http.StatusBadRequest:
					c.breaker.success()
					return "declined", nil
				case http.StatusGatewayTimeout:
					lastErr = errors.NewGatewayTimeoutError("gateway timeout")
				default:
					lastErr = errors.NewGatewayError("gateway error")
				}
			}
		}

		c.breaker.failure()

		if attempt == c.retries {
			break
		}
		if !isRetryable(lastErr) {
			break
		}

		wait := jitter(backoff)
		if wait > 0 {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return "", errors.NewGatewayTimeoutError("gateway timeout")
			}
		}
		backoff = nextBackoff(backoff, c.backoff.max)
	}

	if lastErr == nil {
		lastErr = errors.NewGatewayError("gateway error")
	}
	return "", lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if domErr, ok := err.(errors.Error); ok {
		return domErr.Code == errors.CodeGatewayTimeout || domErr.Code == errors.CodeGatewayError
	}
	return false
}

func nextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}
	return next
}

func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	// +/- 20% jitter.
	delta := float64(d) * 0.2
	return time.Duration(float64(d) + (rand.Float64()*2-1)*delta)
}

type circuitBreaker struct {
	mu        sync.Mutex
	failures  int
	threshold int
	openUntil time.Time
	cooldown  time.Duration
}

func newCircuitBreaker(threshold int, cooldown time.Duration) *circuitBreaker {
	return &circuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
	}
}

func (c *circuitBreaker) allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.openUntil) {
		return false
	}
	return true
}

func (c *circuitBreaker) success() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
}

func (c *circuitBreaker) failure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.failures >= c.threshold {
		c.openUntil = time.Now().Add(c.cooldown)
		c.failures = 0
	}
}
