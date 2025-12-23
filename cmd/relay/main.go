package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"draftea-challenge/internal/adapters/messaging/rabbitmq"
	"draftea-challenge/internal/adapters/persistence/postgres"
	"draftea-challenge/internal/application/outbox/relay"
	"draftea-challenge/internal/platform/clock"
	"draftea-challenge/internal/platform/config"
	"draftea-challenge/internal/platform/db"
	"draftea-challenge/internal/platform/logger"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("relay exited with error: %v", err)
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

	dbConn, dbCleanup, err := db.NewPostgres(cfg.DB, zapLogger)
	if err != nil {
		return err
	}
	defer func() { _ = dbCleanup() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rabbitCfg := rabbitmq.Config{
		URL:                   cfg.Rabbit.URL,
		Exchange:              cfg.Rabbit.Exchange,
		MetricsQueue:          cfg.Rabbit.MetricsQueue,
		AuditQueue:            cfg.Rabbit.AuditQueue,
		PublishConfirmTimeout: cfg.Rabbit.PublishConfirmTimeout,
	}

	publisher, publisherCleanup, err := rabbitmq.NewPublisher(rabbitCfg, zapLogger)
	if err != nil {
		publisher, publisherCleanup, err = connectPublisherWithRetry(
			ctx,
			rabbitCfg,
			zapLogger,
			cfg.Rabbit.RelayMaxRetries,
			cfg.Rabbit.RelayInitialBackoff,
			cfg.Rabbit.RelayMaxBackoff,
		)
		if err != nil {
			return err
		}
	}
	defer func() { _ = publisherCleanup() }()

	outboxRepo := postgres.NewPostgresPersistence(dbConn)

	relayWorker := relay.NewRelay(outboxRepo, publisher, clock.SystemClock{}, relay.Config{
		BatchSize:      cfg.Rabbit.RelayBatchSize,
		MaxInFlight:    cfg.Rabbit.RelayMaxInFlight,
		MaxRetries:     cfg.Rabbit.RelayMaxRetries,
		InitialBackoff: cfg.Rabbit.RelayInitialBackoff,
		MaxBackoff:     cfg.Rabbit.RelayMaxBackoff,
	})

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	zapLogger.Info("outbox relay started")
	for {
		select {
		case <-ctx.Done():
			zapLogger.Info("outbox relay shutting down")
			return nil
		case <-ticker.C:
			pending, err := outboxRepo.GetPendingEvents(ctx, cfg.Rabbit.RelayBatchSize)
			if err != nil {
				zapLogger.Error("outbox relay fetch error", zap.Error(err))
				continue
			}
			zapLogger.Debug("outbox relay tick", zap.Int("pending", len(pending)))
			if len(pending) == 0 {
				continue
			}
			if err := relayWorker.ProcessOnce(ctx); err != nil {
				zapLogger.Error("outbox relay error", zap.Error(err))
				continue
			}
			zapLogger.Debug("outbox relay processed batch", zap.Int("count", len(pending)))
		}
	}
}

func connectPublisherWithRetry(
	ctx context.Context,
	cfg rabbitmq.Config,
	log *zap.Logger,
	maxRetries int,
	initialBackoff time.Duration,
	maxBackoff time.Duration,
) (*rabbitmq.Publisher, func() error, error) {
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
		publisher, cleanup, err := rabbitmq.NewPublisher(cfg, log)
		if err == nil {
			return publisher, cleanup, nil
		}
		lastErr = err
		log.Warn("rabbitmq publisher connect failed", zap.Error(err))
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
