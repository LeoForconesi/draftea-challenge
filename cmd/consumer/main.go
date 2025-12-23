package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	metricsConsumer, metricsCleanup, err := rabbitmq.NewConsumer(rabbitCfg, cfg.Rabbit.MetricsQueue, zapLogger)
	if err != nil {
		return err
	}
	defer func() { _ = metricsCleanup() }()

	auditConsumer, auditCleanup, err := rabbitmq.NewConsumer(rabbitCfg, cfg.Rabbit.AuditQueue, zapLogger)
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
