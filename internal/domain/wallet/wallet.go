package wallet

import (
	"draftea-challenge/internal/domain/errors"

	"github.com/google/uuid"
)

// Wallet representa la entidad de billetera de un usuario, con balances por moneda.
type Wallet struct {
	ID       uuid.UUID        `json:"id"`
	UserID   uuid.UUID        `json:"user_id"`
	Balances map[string]int64 `json:"balances"` // currency -> balance in minor units
	Name     string           `json:"name,omitempty"`
}

// Balance es un value object para representar un saldo en una moneda específica.
type Balance struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"` // en unidades menores (e.g., centavos para USD)
}

// NewWallet crea una nueva wallet para un usuario.
func NewWallet(userID uuid.UUID) (*Wallet, error) {
	return NewWalletWithName(userID, "")
}

// NewWalletWithName crea una wallet con un nombre opcional.
func NewWalletWithName(userID uuid.UUID, name string) (*Wallet, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id cannot be nil", nil)
	}
	if len(name) > 20 {
		return nil, errors.NewValidationError("name must be at most 20 characters", map[string]interface{}{"name": name})
	}
	return &Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Balances: make(map[string]int64),
		Name:     name,
	}, nil
}

// GetBalance devuelve el saldo para una moneda específica.
func (w *Wallet) GetBalance(currency string) int64 {
	return w.Balances[currency]
}

// SetBalance establece el saldo para una moneda (usado internamente, con validación).
func (w *Wallet) SetBalance(currency string, amount int64) error {
	if amount < 0 {
		return errors.NewValidationError("balance cannot be negative", map[string]interface{}{"currency": currency, "amount": amount})
	}
	w.Balances[currency] = amount
	return nil
}

// Debit debita un monto de una moneda (valida fondos suficientes).
func (w *Wallet) Debit(currency string, amount int64) error {
	if amount <= 0 {
		return errors.NewValidationError("debit amount must be positive", map[string]interface{}{"amount": amount})
	}
	current := w.GetBalance(currency)
	if current < amount {
		return errors.NewInsufficientFundsError("insufficient funds", map[string]interface{}{
			"currency": currency,
			"current":  current,
			"required": amount,
		})
	}
	w.Balances[currency] = current - amount
	return nil
}

// Credit acredita un monto a una moneda.
func (w *Wallet) Credit(currency string, amount int64) error {
	if amount <= 0 {
		return errors.NewValidationError("credit amount must be positive", map[string]interface{}{"amount": amount})
	}
	w.Balances[currency] += amount
	return nil
}
