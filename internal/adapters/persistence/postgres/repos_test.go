package postgres

import (
	"context"
	"testing"
	"time"

	domainerrors "draftea-challenge/internal/domain/errors"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateBalanceCreatesRowWithWalletID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&WalletModel{}, &WalletBalanceModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	userID := uuid.New()
	walletID := uuid.New()
	wm := WalletModel{ID: walletID.String(), UserID: userID.String(), CreatedAt: time.Now()}
	if err := db.Create(&wm).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}

	repo := NewPostgresPersistence(db)
	if err := repo.UpdateBalance(context.Background(), userID, "USD", 100); err != nil {
		t.Fatalf("update balance: %v", err)
	}

	var bal WalletBalanceModel
	if err := db.Where("user_id = ? AND currency = ?", userID.String(), "USD").First(&bal).Error; err != nil {
		t.Fatalf("fetch balance: %v", err)
	}
	if bal.WalletID != walletID.String() {
		t.Fatalf("expected wallet_id %s, got %s", walletID.String(), bal.WalletID)
	}
}

func TestUpdateBalanceMissingWallet(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&WalletModel{}, &WalletBalanceModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewPostgresPersistence(db)
	missingUser := uuid.New()
	err = repo.UpdateBalance(context.Background(), missingUser, "USD", 100)
	if err == nil {
		t.Fatalf("expected error")
	}
	if domErr, ok := err.(domainerrors.Error); !ok || domErr.Code != domainerrors.CodeNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}
