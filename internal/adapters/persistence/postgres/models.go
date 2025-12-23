package postgres

import (
	"time"

	"gorm.io/gorm"
)

type WalletModel struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	UserID    string `gorm:"type:varchar(36);uniqueIndex"`
	Name      string `gorm:"type:char(20)"`
	CreatedAt time.Time
}

type WalletBalanceModel struct {
	ID             string `gorm:"primaryKey;type:varchar(36)"`
	WalletID       string `gorm:"type:varchar(36);index"`
	UserID         string `gorm:"type:varchar(36);index"`
	Currency       string `gorm:"type:varchar(8);index"`
	CurrentBalance int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TransactionModel struct {
	ID                string `gorm:"primaryKey;type:varchar(36)"`
	UserID            string `gorm:"type:varchar(36);index"`
	Type              string `gorm:"type:varchar(32)"`
	Amount            int64
	Currency          string `gorm:"type:varchar(8)"`
	Status            string `gorm:"type:varchar(32);index"`
	ProviderID        string `gorm:"type:varchar(36);index"`
	ExternalReference string `gorm:"type:text"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type IdempotencyModel struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	UserID    string `gorm:"type:varchar(36);index"`
	Key       string `gorm:"type:text;index"`
	RequestID string `gorm:"type:varchar(36)"`
	Response  string `gorm:"type:jsonb"`
	CreatedAt time.Time
}

type OutboxModel struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	EventType string `gorm:"type:varchar(128)"`
	Payload   string `gorm:"type:jsonb"`
	CreatedAt time.Time
	SentAt    *time.Time
}

// Ensure GORM recognizes table names (optional)
func (WalletModel) TableName() string        { return "wallets" }
func (WalletBalanceModel) TableName() string { return "wallet_balances" }
func (TransactionModel) TableName() string   { return "transactions" }
func (IdempotencyModel) TableName() string   { return "idempotency_records" }
func (OutboxModel) TableName() string        { return "outbox" }

// AutoMigrate helper
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&WalletModel{}, &WalletBalanceModel{}, &TransactionModel{}, &IdempotencyModel{}, &OutboxModel{})
}
