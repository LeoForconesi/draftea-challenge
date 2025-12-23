package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"draftea-challenge/internal/platform/config"
	"draftea-challenge/internal/platform/factory"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("api exited with error: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	app, err := factory.Build(cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = app.Cleanup()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.Logger.Info("api listening", zap.String("addr", cfg.App.HTTPAddr), zap.String("env", cfg.App.Env))
	return app.Server.Run(ctx)
}
