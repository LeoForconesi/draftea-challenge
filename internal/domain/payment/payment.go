package payment

import (
	"github.com/google/uuid"
	"draftea-challenge/internal/domain/errors"
)

// Payment representa una solicitud de pago de servicios.
type Payment struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	ProviderID        uuid.UUID `json:"provider_id"`
	ExternalReference string    `json:"external_reference"`
	Amount            int64     `json:"amount"` // en minor units
	Currency          string    `json:"currency"`
}

// NewPayment crea una nueva solicitud de pago.
func NewPayment(userID uuid.UUID, providerID uuid.UUID, externalRef string, amount int64, currency string) (*Payment, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id cannot be nil", nil)
	}
	if providerID == uuid.Nil {
		return nil, errors.NewValidationError("provider_id cannot be nil", nil)
	}
	if externalRef == "" {
		return nil, errors.NewValidationError("external_reference cannot be empty", nil)
	}
	if amount <= 0 {
		return nil, errors.NewValidationError("amount must be positive", map[string]interface{}{"amount": amount})
	}
	if currency == "" {
		return nil, errors.NewValidationError("currency cannot be empty", nil)
	}
	return &Payment{
		ID:                uuid.New(),
		UserID:            userID,
		ProviderID:        providerID,
		ExternalReference: externalRef,
		Amount:            amount,
		Currency:          currency,
	}, nil
}
