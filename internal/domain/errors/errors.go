package errors

import "fmt"

// Error representa un error de dominio con código tipado, mensaje y detalles opcionales.
type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implementa la interfaz error.
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Códigos de error de dominio (tipados).
const (
	CodeValidationError    = "VALIDATION_ERROR"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeNotFound           = "NOT_FOUND"
	CodeInsufficientFunds  = "INSUFFICIENT_FUNDS"
	CodeGatewayTimeout     = "GATEWAY_TIMEOUT"
	CodeGatewayError       = "GATEWAY_ERROR"
	CodeInternal           = "INTERNAL"
)

// Funciones constructoras para errores comunes.
func NewValidationError(message string, details map[string]interface{}) Error {
	return Error{
		Code:    CodeValidationError,
		Message: message,
		Details: details,
	}
}

func NewUnauthorizedError(message string) Error {
	return Error{
		Code:    CodeUnauthorized,
		Message: message,
	}
}

func NewNotFoundError(message string) Error {
	return Error{
		Code:    CodeNotFound,
		Message: message,
	}
}

func NewInsufficientFundsError(message string, details map[string]interface{}) Error {
	return Error{
		Code:    CodeInsufficientFunds,
		Message: message,
		Details: details,
	}
}

func NewGatewayTimeoutError(message string) Error {
	return Error{
		Code:    CodeGatewayTimeout,
		Message: message,
	}
}

func NewGatewayError(message string) Error {
	return Error{
		Code:    CodeGatewayError,
		Message: message,
	}
}

func NewInternalError(message string) Error {
	return Error{
		Code:    CodeInternal,
		Message: message,
	}
}
