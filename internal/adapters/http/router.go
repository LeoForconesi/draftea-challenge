package httpapi

import (
	"time"

	"draftea-challenge/internal/adapters/http/handlers"
	"draftea-challenge/internal/adapters/http/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouterDeps defines dependencies needed to build the router.
type RouterDeps struct {
	Logger         *zap.Logger
	APIKey         string
	RequestTimeout time.Duration
	PaymentHandler *handlers.PaymentHandler
	WalletHandler  *handlers.WalletHandler
}

// NewRouter builds the Gin engine with middleware and routes.
func NewRouter(deps RouterDeps) *gin.Engine {
	router := gin.New()

	router.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.APIKeyAuth(deps.APIKey),
		middleware.Timeout(deps.RequestTimeout),
	)

	walletsGroup := router.Group("/wallets/:user_id")
	walletsGroup.POST("/payments", deps.PaymentHandler.CreatePayment)
	walletsGroup.GET("/balance", deps.WalletHandler.GetBalance)
	walletsGroup.GET("/transactions", deps.WalletHandler.ListTransactions)

	return router
}
