package presenter

import (
	"net/http"

	"draftea-challenge/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// ErrorResponse wraps domain errors in the API response shape.
type ErrorResponse struct {
	Error errors.Error `json:"error"`
}

// WriteError writes a consistent JSON error response.
func WriteError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if domErr, ok := err.(errors.Error); ok {
		c.JSON(statusFor(domErr.Code), ErrorResponse{Error: domErr})
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: errors.NewInternalError("unexpected error"),
	})
}

func statusFor(code string) int {
	switch code {
	case errors.CodeValidationError:
		return http.StatusBadRequest
	case errors.CodeUnauthorized:
		return http.StatusUnauthorized
	case errors.CodeNotFound:
		return http.StatusNotFound
	case errors.CodeInsufficientFunds:
		return http.StatusConflict
	case errors.CodeGatewayTimeout:
		return http.StatusGatewayTimeout
	case errors.CodeGatewayError:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
