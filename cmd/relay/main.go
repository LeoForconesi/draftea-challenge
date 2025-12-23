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

	rabbitCfg := rabbitmq.Config{
		URL:                   cfg.Rabbit.URL,
		Exchange:              cfg.Rabbit.Exchange,
		MetricsQueue:          cfg.Rabbit.MetricsQueue,
		AuditQueue:            cfg.Rabbit.AuditQueue,
		PublishConfirmTimeout: cfg.Rabbit.PublishConfirmTimeout,
	}

	publisher, publisherCleanup, err := rabbitmq.NewPublisher(rabbitCfg, zapLogger)
	if err != nil {
		return err
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	zapLogger.Info("outbox relay started")
	for {
		select {
		case <-ctx.Done():
			zapLogger.Info("outbox relay shutting down")
			return nil
		case <-ticker.C:
			if err := relayWorker.ProcessOnce(ctx); err != nil {
				zapLogger.Error("outbox relay error", zap.Error(err))
			}
		}
	}
}
