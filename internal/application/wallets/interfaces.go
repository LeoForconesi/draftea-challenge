package wallets

import (
	"context"
	"draftea-challenge/internal/domain/wallet"
	"time"

	"github.com/google/uuid"
)

// WalletRepository define la interfaz para acceder a datos de wallets.
type WalletRepository interface {
	GetWallet(ctx context.Context, userID uuid.UUID) (*wallet.Wallet, error)
	CreateWallet(ctx context.Context, w *wallet.Wallet) error
	UpdateBalance(ctx context.Context, userID uuid.UUID, currency string, newBalance int64) error
	ListWallets(ctx context.Context, limit, offset int) ([]*wallet.Wallet, int, error)
}

// Clock define la interfaz para obtener el tiempo actual (para testabilidad).
type Clock interface {
	Now() time.Time
}
