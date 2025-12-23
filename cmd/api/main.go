package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"draftea-challenge/internal/platform/config"
	"draftea-challenge/internal/platform/logger"
	"draftea-challenge/internal/platform/server"

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

	zapLogger, err := logger.New(logger.Config{
		Level:       cfg.Logger.Level,
		Development: cfg.Logger.Development,
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = zapLogger.Sync()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	srv := server.New(cfg.App.HTTPAddr, mux, cfg.App.ShutdownTimeout)

	zapLogger.Info("api listening", zap.String("addr", cfg.App.HTTPAddr), zap.String("env", cfg.App.Env))
	return srv.Run(ctx)
}
