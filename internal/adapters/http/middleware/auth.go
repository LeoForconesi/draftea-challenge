package middleware

import (
	"draftea-challenge/internal/adapters/http/presenter"
	"draftea-challenge/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

const apiKeyHeader = "X-API-Key"

// APIKeyAuth enforces a static API key if configured.
func APIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.Next()
			return
		}

		if c.GetHeader(apiKeyHeader) != expectedKey {
			presenter.WriteError(c, errors.NewUnauthorizedError("invalid api key"))
			c.Abort()
			return
		}
		c.Next()
	}
}
