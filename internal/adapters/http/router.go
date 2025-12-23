package httpapi

import (
	"time"

	"draftea-challenge/internal/adapters/http/handlers"
	"draftea-challenge/internal/adapters/http/middleware"

	"github.com/gin-contrib/cors"
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
		cors.New(cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "OPTIONS"},
			AllowHeaders: []string{
				"Content-Type",
				"Idempotency-Key",
				"X-API-Key",
				"X-Request-ID",
			},
			ExposeHeaders: []string{"X-Request-ID"},
			MaxAge:        12 * time.Hour,
		}),
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.APIKeyAuth(deps.APIKey),
		middleware.Timeout(deps.RequestTimeout),
	)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/wallets", deps.WalletHandler.ListWallets)
	router.POST("/wallets", deps.WalletHandler.CreateWallet)

	walletsGroup := router.Group("/wallets/:user_id")
	walletsGroup.POST("/payments", deps.PaymentHandler.CreatePayment)
	walletsGroup.GET("/balance", deps.WalletHandler.GetBalance)
	walletsGroup.GET("/transactions", deps.WalletHandler.ListTransactions)
	walletsGroup.POST("/top-up", deps.WalletHandler.TopUp)

	return router
}
