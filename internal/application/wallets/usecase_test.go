package wallets

import (
	"context"
	"testing"

	"draftea-challenge/internal/domain/errors"
	"draftea-challenge/internal/domain/transaction"
	"draftea-challenge/internal/domain/wallet"

	"github.com/google/uuid"
)

type mockWalletRepo struct {
	wallet *wallet.Wallet
	err    error
}

func (m *mockWalletRepo) GetWallet(ctx context.Context, userID uuid.UUID) (*wallet.Wallet, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.wallet, nil
}

func (m *mockWalletRepo) CreateWallet(ctx context.Context, w *wallet.Wallet) error {
	return nil
}

func (m *mockWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, currency string, newBalance int64) error {
	return nil
}

type mockPaymentRepo struct {
	transactions []*transaction.Transaction
}

func (m *mockPaymentRepo) ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	return m.transactions, nil
}

func TestGetBalance_NotFoundReturnsEmpty(t *testing.T) {
	userID := uuid.New()
	repo := &mockWalletRepo{err: errors.NewNotFoundError("wallet not found")}

	svc := NewGetBalanceService(repo)
	resp, err := svc.GetBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserID != userID {
		t.Fatalf("unexpected user id")
	}
	if len(resp.Balances) != 0 {
		t.Fatalf("expected empty balances")
	}
}

func TestGetTransactions(t *testing.T) {
	userID := uuid.New()
	tx, _ := transaction.NewTransaction(userID, transaction.TypePayment, 100, "USD", uuid.New(), "ref")

	repo := &mockPaymentRepo{transactions: []*transaction.Transaction{tx}}
	svc := NewGetTransactionsService(repo)

	resp, err := svc.GetTransactions(context.Background(), &GetTransactionsRequest{UserID: userID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Transactions) != 1 {
		t.Fatalf("expected 1 transaction")
	}
}
