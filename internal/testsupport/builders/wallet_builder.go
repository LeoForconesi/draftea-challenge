package builders

import (
	"draftea-challenge/internal/domain/wallet"

	"github.com/google/uuid"
)

// WalletBuilder builds wallets with preset balances for tests.
type WalletBuilder struct {
	userID   uuid.UUID
	balances map[string]int64
}

func NewWalletBuilder() *WalletBuilder {
	return &WalletBuilder{
		userID:   uuid.New(),
		balances: make(map[string]int64),
	}
}

func (b *WalletBuilder) WithUserID(id uuid.UUID) *WalletBuilder {
	b.userID = id
	return b
}

func (b *WalletBuilder) WithBalance(currency string, amount int64) *WalletBuilder {
	b.balances[currency] = amount
	return b
}

func (b *WalletBuilder) Build() (*wallet.Wallet, error) {
	w, err := wallet.NewWallet(b.userID)
	if err != nil {
		return nil, err
	}
	for currency, amount := range b.balances {
		if err := w.SetBalance(currency, amount); err != nil {
			return nil, err
		}
	}
	return w, nil
}
