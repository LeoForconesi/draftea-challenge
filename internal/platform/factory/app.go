package factory

import (
	"draftea-challenge/internal/adapters/gateway/httpclient"
	"draftea-challenge/internal/adapters/http"
	"draftea-challenge/internal/adapters/http/handlers"
	"draftea-challenge/internal/adapters/persistence/postgres"
	"draftea-challenge/internal/application/payments"
	"draftea-challenge/internal/application/wallets"
	"draftea-challenge/internal/platform/clock"
	"draftea-challenge/internal/platform/config"
	"draftea-challenge/internal/platform/db"
	"draftea-challenge/internal/platform/idgen"
	"draftea-challenge/internal/platform/logger"
	"draftea-challenge/internal/platform/server"

	"go.uber.org/zap"
)

// App bundles runtime components.
type App struct {
	Server  *server.Server
	Logger  *zap.Logger
	Cleanup func() error
}

// Build wires the application with concrete dependencies.
func Build(cfg config.Config) (*App, error) {
	zapLogger, err := logger.New(logger.Config{
		Level:       cfg.Logger.Level,
		Development: cfg.Logger.Development,
	})
	if err != nil {
		return nil, err
	}

	dbConn, dbCleanup, err := db.NewPostgres(cfg.DB, zapLogger)
	if err != nil {
		_ = zapLogger.Sync()
		return nil, err
	}

	persistence := postgres.NewPostgresPersistence(dbConn)
	gateway := httpclient.New(httpclient.Config{
		BaseURL:                cfg.Gateway.URL,
		Timeout:                cfg.Gateway.Timeout,
		MaxRetries:             cfg.Gateway.MaxRetries,
		RetryInitialBackoff:    cfg.Gateway.RetryInitialBackoff,
		RetryMaxBackoff:        cfg.Gateway.RetryMaxBackoff,
		CircuitBreakerFailures: cfg.Gateway.CircuitBreakerFailures,
		CircuitBreakerCooldown: cfg.Gateway.CircuitBreakerCooldown,
		MaxInFlight:            cfg.Gateway.MaxInFlight,
	})

	paymentService := payments.NewPaymentService(
		persistence,
		persistence,
		gateway,
		persistence,
		persistence,
		idgen.UUIDGenerator{},
		clock.SystemClock{},
	)
	balanceService := wallets.NewGetBalanceService(persistence)
	transactionsService := wallets.NewGetTransactionsService(persistence)
	topUpService := wallets.NewTopUpService(persistence, persistence)
	listService := wallets.NewListWalletsService(persistence)
	createWalletService := wallets.NewCreateWalletService(persistence)

	paymentHandler := handlers.NewPaymentHandler(paymentService)
	walletHandler := handlers.NewWalletHandler(balanceService, transactionsService, topUpService, listService, createWalletService)

	router := httpapi.NewRouter(httpapi.RouterDeps{
		Logger:         zapLogger,
		APIKey:         cfg.App.APIKey,
		RequestTimeout: cfg.App.RequestTimeout,
		PaymentHandler: paymentHandler,
		WalletHandler:  walletHandler,
	})

	srv := server.New(cfg.App.HTTPAddr, router, cfg.App.ShutdownTimeout)

	cleanup := func() error {
		var firstErr error
		if err := dbCleanup(); err != nil {
			firstErr = err
		}
		if err := zapLogger.Sync(); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	return &App{
		Server:  srv,
		Logger:  zapLogger,
		Cleanup: cleanup,
	}, nil
}
