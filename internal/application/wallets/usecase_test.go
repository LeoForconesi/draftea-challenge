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
	wallet       *wallet.Wallet
	err          error
	updateCalls  int
	updatedValue int64
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
	m.updateCalls++
	m.updatedValue = newBalance
	return nil
}

func (m *mockWalletRepo) ListWallets(ctx context.Context, limit, offset int) ([]*wallet.Wallet, int, error) {
	return nil, 0, nil
}

type mockPaymentRepo struct {
	transactions []*transaction.Transaction
	createdTxs   []*transaction.Transaction
	statuses     []transaction.Status
}

func (m *mockPaymentRepo) ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	return m.transactions, nil
}

func (m *mockPaymentRepo) CreateTransaction(ctx context.Context, tx *transaction.Transaction) error {
	m.createdTxs = append(m.createdTxs, tx)
	return nil
}

func (m *mockPaymentRepo) UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status transaction.Status) error {
	m.statuses = append(m.statuses, status)
	return nil
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
	if resp.Name != "" {
		t.Fatalf("expected empty name")
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

func TestTopUpCreatesBalance(t *testing.T) {
	userID := uuid.New()
	w, _ := wallet.NewWallet(userID)
	_ = w.SetBalance("USD", 100)
	repo := &mockWalletRepo{wallet: w}
	txRepo := &mockPaymentRepo{}

	svc := NewTopUpService(repo, txRepo)
	resp, err := svc.TopUp(context.Background(), &TopUpRequest{UserID: userID, Amount: 1000, Currency: "USD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Balance != 1100 {
		t.Fatalf("expected balance 1100, got %d", resp.Balance)
	}
	if len(txRepo.createdTxs) != 1 {
		t.Fatalf("expected transaction created")
	}
}

func TestListWallets(t *testing.T) {
	userID := uuid.New()
	w, _ := wallet.NewWallet(userID)
	_ = w.SetBalance("USD", 100)

	repo := &mockListWalletRepo{wallets: []*wallet.Wallet{w}, total: 1}
	svc := NewListWalletsService(repo)
	resp, err := svc.ListWallets(context.Background(), &ListWalletsRequest{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Wallets) != 1 {
		t.Fatalf("expected 1 wallet")
	}
}

type mockListWalletRepo struct {
	wallets []*wallet.Wallet
	total   int
}

func (m *mockListWalletRepo) GetWallet(ctx context.Context, userID uuid.UUID) (*wallet.Wallet, error) {
	return nil, errors.NewNotFoundError("wallet not found")
}

func (m *mockListWalletRepo) CreateWallet(ctx context.Context, w *wallet.Wallet) error {
	return nil
}

func (m *mockListWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, currency string, newBalance int64) error {
	return nil
}

func (m *mockListWalletRepo) ListWallets(ctx context.Context, limit, offset int) ([]*wallet.Wallet, int, error) {
	return m.wallets, m.total, nil
}
