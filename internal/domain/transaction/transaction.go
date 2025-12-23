package transaction

import (
	"draftea-challenge/internal/domain/errors"
	"time"

	"github.com/google/uuid"
)

// Transaction representa una transacción inmutable en el ledger.
type Transaction struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Type              Type      `json:"type"`
	Amount            int64     `json:"amount"` // en minor units
	Currency          string    `json:"currency"`
	Status            Status    `json:"status"`
	ProviderID        uuid.UUID `json:"provider_id,omitempty"`
	ExternalReference string    `json:"external_reference,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Type define el tipo de transacción.
type Type string

const (
	TypePayment Type = "PAYMENT"
	TypeRefund  Type = "REFUND"
)

// Status define el estado de la transacción.
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusDeclined Status = "DECLINED"
	StatusFailed   Status = "FAILED"
)

// NewTransaction crea una nueva transacción.
func NewTransaction(userID uuid.UUID, txType Type, amount int64, currency string, providerID uuid.UUID, externalRef string) (*Transaction, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id cannot be nil", nil)
	}
	if amount <= 0 {
		return nil, errors.NewValidationError("amount must be positive", map[string]interface{}{"amount": amount})
	}
	if currency == "" {
		return nil, errors.NewValidationError("currency cannot be empty", nil)
	}
	now := time.Now()
	return &Transaction{
		ID:                uuid.New(),
		UserID:            userID,
		Type:              txType,
		Amount:            amount,
		Currency:          currency,
		Status:            StatusPending,
		ProviderID:        providerID,
		ExternalReference: externalRef,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// UpdateStatus actualiza el estado de la transacción (solo para cambios válidos).
func (t *Transaction) UpdateStatus(newStatus Status) error {
	validTransitions := map[Status][]Status{
		StatusPending:  {StatusApproved, StatusDeclined, StatusFailed},
		StatusApproved: {},
		StatusDeclined: {},
		StatusFailed:   {},
	}
	if !contains(validTransitions[t.Status], newStatus) {
		return errors.NewValidationError("invalid status transition", map[string]interface{}{
			"current": t.Status,
			"new":     newStatus,
		})
	}
	t.Status = newStatus
	t.UpdatedAt = time.Now()
	return nil
}

// contains verifica si un slice contiene un elemento.
func contains(slice []Status, item Status) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
