package relay

import (
	"context"
	"fmt"
	"sync"
	"time"

	"draftea-challenge/internal/application/outbox"
	"draftea-challenge/internal/application/ports"
)

// Relay publishes pending outbox events to the message broker.
type Relay struct {
	repo        outbox.OutboxRepository
	publisher   outbox.MessagePublisher
	clock       ports.Clock
	batchSize   int
	maxInFlight int
	retries     int
	backoff     backoffConfig
}

type backoffConfig struct {
	initial time.Duration
	max     time.Duration
}

// Config configures relay behavior.
type Config struct {
	BatchSize      int
	MaxInFlight    int
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// NewRelay creates a new outbox relay.
func NewRelay(repo outbox.OutboxRepository, publisher outbox.MessagePublisher, clock ports.Clock, cfg Config) *Relay {
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	maxInFlight := cfg.MaxInFlight
	if maxInFlight <= 0 {
		maxInFlight = 10
	}
	retries := cfg.MaxRetries
	if retries < 0 {
		retries = 0
	}
	initialBackoff := cfg.InitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = 200 * time.Millisecond
	}
	maxBackoff := cfg.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 2 * time.Second
	}
	return &Relay{
		repo:        repo,
		publisher:   publisher,
		clock:       clock,
		batchSize:   batchSize,
		maxInFlight: maxInFlight,
		retries:     retries,
		backoff: backoffConfig{
			initial: initialBackoff,
			max:     maxBackoff,
		},
	}
}

// ProcessOnce publishes a single batch of pending events.
func (r *Relay) ProcessOnce(ctx context.Context) error {
	events, err := r.repo.GetPendingEvents(ctx, r.batchSize)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	sem := make(chan struct{}, r.maxInFlight)
	var wg sync.WaitGroup
	errCh := make(chan error, len(events))

	for _, event := range events {
		wg.Add(1)
		sem <- struct{}{}
		go func(ev *outbox.OutboxEvent) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := r.publishWithRetry(ctx, ev); err != nil {
				errCh <- err
				return
			}
			now := r.clock.Now()
			ev.SentAt = &now
			if err := r.repo.MarkEventAsSent(ctx, ev.ID); err != nil {
				errCh <- fmt.Errorf("mark outbox event %s as sent: %w", ev.ID, err)
			}
		}(event)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Relay) publishWithRetry(ctx context.Context, event *outbox.OutboxEvent) error {
	var lastErr error
	backoff := r.backoff.initial

	for attempt := 0; attempt <= r.retries; attempt++ {
		if err := r.publisher.Publish(ctx, "payments.events", event.EventType, []byte(event.Payload)); err != nil {
			lastErr = err
		} else {
			return nil
		}

		if attempt == r.retries {
			break
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
		backoff = nextBackoff(backoff, r.backoff.max)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("failed to publish outbox event %s", event.ID)
	}
	return fmt.Errorf("publish outbox event %s: %w", event.ID, lastErr)
}

func nextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}
	return next
}
