package middleware

import (
	"draftea-challenge/internal/adapters/http/presenter"
	"draftea-challenge/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery recovers from panics and logs the error.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error("panic recovered", zap.Any("error", recovered), zap.String("request_id", GetRequestID(c)))
		presenter.WriteError(c, errors.NewInternalError("internal server error"))
		c.Abort()
	})
}
