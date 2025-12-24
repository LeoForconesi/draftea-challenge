package wallet

import (
	"testing"

	"draftea-challenge/internal/domain/errors"

	"github.com/google/uuid"
)

func TestWalletDebitInsufficientFunds(t *testing.T) {
	w, err := NewWallet(uuid.New())
	if err != nil {
		t.Fatalf("new wallet: %v", err)
	}
	if err := w.SetBalance("USD", 100); err != nil {
		t.Fatalf("set balance: %v", err)
	}

	err = w.Debit("USD", 200)
	if err == nil {
		t.Fatalf("expected error")
	}
	if domErr, ok := err.(errors.Error); !ok || domErr.Code != errors.CodeInsufficientFunds {
		t.Fatalf("expected insufficient funds error, got %v", err)
	}
}

func TestWalletCredit(t *testing.T) {
	w, err := NewWallet(uuid.New())
	if err != nil {
		t.Fatalf("new wallet: %v", err)
	}
	if err := w.Credit("USD", 500); err != nil {
		t.Fatalf("credit: %v", err)
	}
	if got := w.GetBalance("USD"); got != 500 {
		t.Fatalf("expected balance 500, got %d", got)
	}
}
