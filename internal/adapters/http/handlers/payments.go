package handlers

import (
	"context"
	"net/http"

	"draftea-challenge/internal/adapters/http/presenter"
	"draftea-challenge/internal/application/payments"
	"draftea-challenge/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentHandler handles payment endpoints.
type PaymentHandler struct {
	service interface {
		ProcessPayment(ctx context.Context, req *payments.ProcessPaymentRequest) (*payments.ProcessPaymentResponse, error)
	}
}

// NewPaymentHandler creates a PaymentHandler.
func NewPaymentHandler(service interface {
	ProcessPayment(ctx context.Context, req *payments.ProcessPaymentRequest) (*payments.ProcessPaymentResponse, error)
}) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type paymentRequest struct {
	ProviderID        string `json:"provider_id"`
	ExternalReference string `json:"external_reference"`
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
}

// CreatePayment handles POST /wallets/{user_id}/payments.
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid user_id", map[string]interface{}{"user_id": c.Param("user_id")}))
		return
	}

	var body paymentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid request body", map[string]interface{}{"error": err.Error()}))
		return
	}

	details := make(map[string]interface{})

	providerID, err := uuid.Parse(body.ProviderID)
	if err != nil {
		details["provider_id"] = body.ProviderID
	}
	if body.ExternalReference == "" {
		details["external_reference"] = "required"
	}
	if body.Amount <= 0 {
		details["amount"] = body.Amount
	}
	if body.Currency == "" {
		details["currency"] = "required"
	}
	if len(details) > 0 {
		presenter.WriteError(c, errors.NewValidationError("invalid payment request", details))
		return
	}

	req := &payments.ProcessPaymentRequest{
		UserID:            userID,
		ProviderID:        providerID,
		ExternalReference: body.ExternalReference,
		Amount:            body.Amount,
		Currency:          body.Currency,
		IdempotencyKey:    c.GetHeader("Idempotency-Key"),
	}

	resp, err := h.service.ProcessPayment(c.Request.Context(), req)
	if err != nil {
		presenter.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
