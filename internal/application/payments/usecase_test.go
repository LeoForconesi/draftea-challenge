package payments

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"draftea-challenge/internal/application/outbox"
	"draftea-challenge/internal/domain/errors"
	"draftea-challenge/internal/domain/payment"
	"draftea-challenge/internal/domain/transaction"
	"draftea-challenge/internal/domain/wallet"

	"github.com/google/uuid"
)

type mockPaymentRepo struct {
	createdTxs []*transaction.Transaction
	updates    []transaction.Status
}

func (m *mockPaymentRepo) CreateTransaction(ctx context.Context, tx *transaction.Transaction) error {
	m.createdTxs = append(m.createdTxs, tx)
	return nil
}

func (m *mockPaymentRepo) UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status transaction.Status) error {
	m.updates = append(m.updates, status)
	return nil
}

func (m *mockPaymentRepo) GetTransactionByID(ctx context.Context, txID uuid.UUID) (*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockPaymentRepo) ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	return nil, nil
}

type mockWalletRepo struct {
	wallet       *wallet.Wallet
	getErr       error
	created      bool
	balanceCalls []int64
}

func (m *mockWalletRepo) GetWallet(ctx context.Context, userID uuid.UUID) (*wallet.Wallet, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.wallet, nil
}

func (m *mockWalletRepo) CreateWallet(ctx context.Context, w *wallet.Wallet) error {
	m.created = true
	m.wallet = w
	return nil
}

func (m *mockWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, currency string, newBalance int64) error {
	m.balanceCalls = append(m.balanceCalls, newBalance)
	return nil
}

func (m *mockWalletRepo) ListWallets(ctx context.Context, limit, offset int) ([]*wallet.Wallet, int, error) {
	return nil, 0, nil
}

type mockGateway struct {
	status string
	err    error
	calls  int
}

func (m *mockGateway) ProcessPayment(ctx context.Context, p *payment.Payment) (string, error) {
	m.calls++
	return m.status, m.err
}

type mockIdempotencyRepo struct {
	record      *IdempotencyRecord
	created     *IdempotencyRecord
	createCalls int
}

func (m *mockIdempotencyRepo) GetIdempotencyRecord(ctx context.Context, userID uuid.UUID, key string) (*IdempotencyRecord, error) {
	return m.record, nil
}

func (m *mockIdempotencyRepo) CreateIdempotencyRecord(ctx context.Context, record *IdempotencyRecord) error {
	m.created = record
	m.createCalls++
	return nil
}

type mockOutboxRepo struct {
	events []*outbox.OutboxEvent
}

func (m *mockOutboxRepo) CreateEvent(ctx context.Context, event *outbox.OutboxEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockOutboxRepo) GetPendingEvents(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
	return nil, nil
}

func (m *mockOutboxRepo) MarkEventAsSent(ctx context.Context, eventID uuid.UUID) error {
	return nil
}

type fixedIDGen struct {
	id uuid.UUID
}

func (f fixedIDGen) New() uuid.UUID {
	if f.id != uuid.Nil {
		return f.id
	}
	return uuid.New()
}

type fixedClock struct {
	t time.Time
}

func (f fixedClock) Now() time.Time {
	return f.t
}

func TestProcessPayment_HappyPath(t *testing.T) {
	userID := uuid.New()
	w, err := wallet.NewWallet(userID)
	if err != nil {
		t.Fatalf("wallet init: %v", err)
	}
	if err := w.SetBalance("USD", 1000); err != nil {
		t.Fatalf("set balance: %v", err)
	}

	payRepo := &mockPaymentRepo{}
	walletRepo := &mockWalletRepo{wallet: w}
	gateway := &mockGateway{status: "approved"}
	idemRepo := &mockIdempotencyRepo{}
	outboxRepo := &mockOutboxRepo{}
	clock := fixedClock{t: time.Now()}

	svc := NewPaymentService(payRepo, walletRepo, gateway, idemRepo, outboxRepo, fixedIDGen{}, clock)

	resp, err := svc.ProcessPayment(context.Background(), &ProcessPaymentRequest{
		UserID:            userID,
		ProviderID:        uuid.New(),
		ExternalReference: "ref-1",
		Amount:            500,
		Currency:          "USD",
		IdempotencyKey:    "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != string(transaction.StatusApproved) {
		t.Fatalf("expected approved status, got %s", resp.Status)
	}
	if gateway.calls != 1 {
		t.Fatalf("expected gateway call, got %d", gateway.calls)
	}
	if len(outboxRepo.events) != 2 {
		t.Fatalf("expected 2 outbox events, got %d", len(outboxRepo.events))
	}
	if outboxRepo.events[0].EventType != "payment.created" {
		t.Fatalf("expected payment.created event, got %s", outboxRepo.events[0].EventType)
	}
	if outboxRepo.events[1].EventType != "payment.completed" {
		t.Fatalf("expected payment.completed event, got %s", outboxRepo.events[1].EventType)
	}
}

func TestProcessPayment_InsufficientFunds(t *testing.T) {
	userID := uuid.New()
	w, _ := wallet.NewWallet(userID)
	_ = w.SetBalance("USD", 100)

	payRepo := &mockPaymentRepo{}
	walletRepo := &mockWalletRepo{wallet: w}
	gateway := &mockGateway{status: "approved"}
	idemRepo := &mockIdempotencyRepo{}
	outboxRepo := &mockOutboxRepo{}

	svc := NewPaymentService(payRepo, walletRepo, gateway, idemRepo, outboxRepo, fixedIDGen{}, fixedClock{})

	_, err := svc.ProcessPayment(context.Background(), &ProcessPaymentRequest{
		UserID:            userID,
		ProviderID:        uuid.New(),
		ExternalReference: "ref-1",
		Amount:            500,
		Currency:          "USD",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if domErr, ok := err.(errors.Error); !ok || domErr.Code != errors.CodeInsufficientFunds {
		t.Fatalf("expected insufficient funds error, got %v", err)
	}
	if gateway.calls != 0 {
		t.Fatalf("expected gateway not called")
	}
	if len(payRepo.createdTxs) != 0 {
		t.Fatalf("expected no transaction created")
	}
}

func TestProcessPayment_GatewayTimeout(t *testing.T) {
	userID := uuid.New()
	w, _ := wallet.NewWallet(userID)
	_ = w.SetBalance("USD", 1000)

	payRepo := &mockPaymentRepo{}
	walletRepo := &mockWalletRepo{wallet: w}
	gateway := &mockGateway{err: errors.NewGatewayTimeoutError("timeout")}
	idemRepo := &mockIdempotencyRepo{}
	outboxRepo := &mockOutboxRepo{}

	svc := NewPaymentService(payRepo, walletRepo, gateway, idemRepo, outboxRepo, fixedIDGen{}, fixedClock{t: time.Now()})

	_, err := svc.ProcessPayment(context.Background(), &ProcessPaymentRequest{
		UserID:            userID,
		ProviderID:        uuid.New(),
		ExternalReference: "ref-1",
		Amount:            500,
		Currency:          "USD",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if domErr, ok := err.(errors.Error); !ok || domErr.Code != errors.CodeGatewayTimeout {
		t.Fatalf("expected gateway timeout error, got %v", err)
	}
	if len(outboxRepo.events) == 0 {
		t.Fatalf("expected outbox events")
	}
	last := outboxRepo.events[len(outboxRepo.events)-1]
	if last.EventType != "payment.failed" {
		t.Fatalf("expected payment.failed event, got %s", last.EventType)
	}
}

func TestProcessPayment_IdempotencyHit(t *testing.T) {
	userID := uuid.New()
	recordResp := ProcessPaymentResponse{TransactionID: uuid.New(), Status: string(transaction.StatusApproved)}
	payload, _ := json.Marshal(recordResp)

	payRepo := &mockPaymentRepo{}
	walletRepo := &mockWalletRepo{}
	gateway := &mockGateway{status: "approved"}
	idemRepo := &mockIdempotencyRepo{record: &IdempotencyRecord{UserID: userID, Key: "idem-1", Response: string(payload)}}
	outboxRepo := &mockOutboxRepo{}

	svc := NewPaymentService(payRepo, walletRepo, gateway, idemRepo, outboxRepo, fixedIDGen{}, fixedClock{})

	resp, err := svc.ProcessPayment(context.Background(), &ProcessPaymentRequest{
		UserID:            userID,
		ProviderID:        uuid.New(),
		ExternalReference: "ref-1",
		Amount:            500,
		Currency:          "USD",
		IdempotencyKey:    "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TransactionID != recordResp.TransactionID {
		t.Fatalf("expected idempotent transaction id")
	}
	if gateway.calls != 0 {
		t.Fatalf("expected gateway not called")
	}
}

func TestProcessPayment_DeclinedCreatesRefundEvent(t *testing.T) {
	userID := uuid.New()
	w, _ := wallet.NewWallet(userID)
	_ = w.SetBalance("USD", 1000)

	payRepo := &mockPaymentRepo{}
	walletRepo := &mockWalletRepo{wallet: w}
	gateway := &mockGateway{status: "declined"}
	idemRepo := &mockIdempotencyRepo{}
	outboxRepo := &mockOutboxRepo{}

	svc := NewPaymentService(payRepo, walletRepo, gateway, idemRepo, outboxRepo, fixedIDGen{}, fixedClock{t: time.Now()})

	_, err := svc.ProcessPayment(context.Background(), &ProcessPaymentRequest{
		UserID:            userID,
		ProviderID:        uuid.New(),
		ExternalReference: "ref-1",
		Amount:            500,
		Currency:          "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundRefund := false
	for _, event := range outboxRepo.events {
		if event.EventType == "refund.created" {
			foundRefund = true
		}
	}
	if !foundRefund {
		t.Fatalf("expected refund.created event")
	}
	if last := outboxRepo.events[len(outboxRepo.events)-1]; !strings.HasPrefix(last.EventType, "payment.") {
		t.Fatalf("expected payment event, got %s", last.EventType)
	}
}
