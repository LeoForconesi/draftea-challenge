package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"draftea-challenge/internal/adapters/messaging/rabbitmq"
	"draftea-challenge/internal/platform/config"
	"draftea-challenge/internal/platform/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("consumer exited with error: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zapLogger, err := logger.New(logger.Config{Level: cfg.Logger.Level, Development: cfg.Logger.Development})
	if err != nil {
		return err
	}
	defer func() { _ = zapLogger.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rabbitCfg := rabbitmq.Config{
		URL:                   cfg.Rabbit.URL,
		Exchange:              cfg.Rabbit.Exchange,
		MetricsQueue:          cfg.Rabbit.MetricsQueue,
		AuditQueue:            cfg.Rabbit.AuditQueue,
		PublishConfirmTimeout: cfg.Rabbit.PublishConfirmTimeout,
	}

	metricsConsumer, metricsCleanup, err := connectConsumerWithRetry(
		ctx,
		rabbitCfg,
		cfg.Rabbit.MetricsQueue,
		zapLogger,
		cfg.Rabbit.RelayMaxRetries,
		cfg.Rabbit.RelayInitialBackoff,
		cfg.Rabbit.RelayMaxBackoff,
	)
	if err != nil {
		return err
	}
	defer func() { _ = metricsCleanup() }()

	auditConsumer, auditCleanup, err := connectConsumerWithRetry(
		ctx,
		rabbitCfg,
		cfg.Rabbit.AuditQueue,
		zapLogger,
		cfg.Rabbit.RelayMaxRetries,
		cfg.Rabbit.RelayInitialBackoff,
		cfg.Rabbit.RelayMaxBackoff,
	)
	if err != nil {
		return err
	}
	defer func() { _ = auditCleanup() }()

	go func() {
		err := metricsConsumer.Start(ctx, func(ctx context.Context, msg amqp.Delivery) error {
			zapLogger.Info("metrics event", zap.String("routing_key", msg.RoutingKey), zap.ByteString("body", msg.Body))
			return nil
		})
		if err != nil {
			zapLogger.Error("metrics consumer stopped", zap.Error(err))
		}
	}()

	go func() {
		err := auditConsumer.Start(ctx, func(ctx context.Context, msg amqp.Delivery) error {
			zapLogger.Info("audit event", zap.String("routing_key", msg.RoutingKey), zap.ByteString("body", msg.Body))
			return nil
		})
		if err != nil {
			zapLogger.Error("audit consumer stopped", zap.Error(err))
		}
	}()

	<-ctx.Done()
	zapLogger.Info("consumers shutting down")
	return nil
}

func connectConsumerWithRetry(
	ctx context.Context,
	cfg rabbitmq.Config,
	queue string,
	log *zap.Logger,
	maxRetries int,
	initialBackoff time.Duration,
	maxBackoff time.Duration,
) (*rabbitmq.Consumer, func() error, error) {
	attempts := maxRetries + 1
	if attempts <= 0 {
		attempts = 1
	}
	backoff := initialBackoff
	if backoff <= 0 {
		backoff = 200 * time.Millisecond
	}
	if maxBackoff <= 0 {
		maxBackoff = 2 * time.Second
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		consumer, cleanup, err := rabbitmq.NewConsumer(cfg, queue, log)
		if err == nil {
			return consumer, cleanup, nil
		}
		lastErr = err
		log.Warn("rabbitmq consumer connect failed", zap.Error(err), zap.String("queue", queue))
		if i == attempts-1 {
			break
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
		next := backoff * 2
		if next > maxBackoff {
			next = maxBackoff
		}
		backoff = next
	}
	return nil, nil, lastErr
}
