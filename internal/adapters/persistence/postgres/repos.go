package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	appoutbox "draftea-challenge/internal/application/outbox"
	"draftea-challenge/internal/application/payments"
	"draftea-challenge/internal/application/wallets"
	domainerrors "draftea-challenge/internal/domain/errors"
	domaintx "draftea-challenge/internal/domain/transaction"
	domainwallet "draftea-challenge/internal/domain/wallet"
)

type PostgresPersistence struct {
	db *gorm.DB
}

func NewPostgresPersistence(db *gorm.DB) *PostgresPersistence {
	return &PostgresPersistence{db: db}
}

// WalletRepository
func (p *PostgresPersistence) GetWallet(ctx context.Context, userID uuid.UUID) (*domainwallet.Wallet, error) {
	var w WalletModel
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID.String()).First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewNotFoundError("wallet not found")
		}
		return nil, err
	}
	// Load balances
	var balances []WalletBalanceModel
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID.String()).Find(&balances).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64)
	for _, b := range balances {
		m[b.Currency] = b.CurrentBalance
	}
	return &domainwallet.Wallet{
		ID:       uuid.MustParse(w.ID),
		UserID:   uuid.MustParse(w.UserID),
		Balances: m,
		Name:     w.Name,
	}, nil
}

func (p *PostgresPersistence) CreateWallet(ctx context.Context, w *domainwallet.Wallet) error {
	wm := WalletModel{ID: w.ID.String(), UserID: w.UserID.String(), Name: w.Name, CreatedAt: time.Now()}
	if err := p.db.WithContext(ctx).Create(&wm).Error; err != nil {
		return err
	}
	// create balances rows
	for cur, bal := range w.Balances {
		bm := WalletBalanceModel{ID: uuid.NewString(), WalletID: wm.ID, UserID: wm.UserID, Currency: cur, CurrentBalance: bal, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := p.db.WithContext(ctx).Create(&bm).Error; err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresPersistence) UpdateBalance(ctx context.Context, userID uuid.UUID, currency string, newBalance int64) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var bal WalletBalanceModel
		// Lock the specific row for update
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND currency = ?", userID.String(), currency).First(&bal).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				var walletRow WalletModel
				if err := tx.Where("user_id = ?", userID.String()).First(&walletRow).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return domainerrors.NewNotFoundError("wallet not found")
					}
					return err
				}
				// create row with the existing wallet_id
				b := WalletBalanceModel{
					ID:             uuid.NewString(),
					WalletID:       walletRow.ID,
					UserID:         userID.String(),
					Currency:       currency,
					CurrentBalance: newBalance,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}
				if err := tx.Create(&b).Error; err != nil {
					return err
				}
				return nil
			}
			return err
		}
		bal.CurrentBalance = newBalance
		bal.UpdatedAt = time.Now()
		if err := tx.Save(&bal).Error; err != nil {
			return err
		}
		return nil
	})
}

// Ensure PostgresPersistence implements WalletRepository interface
var _ wallets.WalletRepository = (*PostgresPersistence)(nil)

func (p *PostgresPersistence) ListWallets(ctx context.Context, limit, offset int) ([]*domainwallet.Wallet, int, error) {
	var total int64
	if err := p.db.WithContext(ctx).Model(&WalletModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []WalletModel
	if err := p.db.WithContext(ctx).Order("created_at").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []*domainwallet.Wallet{}, int(total), nil
	}

	userIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
	}

	var balances []WalletBalanceModel
	if err := p.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&balances).Error; err != nil {
		return nil, 0, err
	}

	balMap := make(map[string]map[string]int64)
	for _, b := range balances {
		if _, ok := balMap[b.UserID]; !ok {
			balMap[b.UserID] = make(map[string]int64)
		}
		balMap[b.UserID][b.Currency] = b.CurrentBalance
	}

	out := make([]*domainwallet.Wallet, 0, len(rows))
	for _, r := range rows {
		balances := balMap[r.UserID]
		if balances == nil {
			balances = make(map[string]int64)
		}
		out = append(out, &domainwallet.Wallet{
			ID:       uuid.MustParse(r.ID),
			UserID:   uuid.MustParse(r.UserID),
			Balances: balances,
			Name:     r.Name,
		})
	}

	return out, int(total), nil
}

// PaymentRepository & IdempotencyRepo & Outbox
func (p *PostgresPersistence) CreateTransaction(ctx context.Context, txDomain *domaintx.Transaction) error {
	var walletRow WalletModel
	if err := p.db.WithContext(ctx).Where("user_id = ?", txDomain.UserID.String()).First(&walletRow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerrors.NewNotFoundError("wallet not found")
		}
		return err
	}
	m := TransactionModel{
		ID:                txDomain.ID.String(),
		WalletID:          walletRow.ID,
		UserID:            txDomain.UserID.String(),
		Type:              string(txDomain.Type),
		Amount:            txDomain.Amount,
		Currency:          txDomain.Currency,
		Status:            string(txDomain.Status),
		ProviderID:        txDomain.ProviderID.String(),
		ExternalReference: txDomain.ExternalReference,
		CreatedAt:         txDomain.CreatedAt,
		UpdatedAt:         txDomain.UpdatedAt,
	}
	return p.db.WithContext(ctx).Create(&m).Error
}

func (p *PostgresPersistence) UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status domaintx.Status) error {
	return p.db.WithContext(ctx).Model(&TransactionModel{}).Where("id = ?", txID.String()).Updates(map[string]interface{}{"status": string(status), "updated_at": time.Now()}).Error
}

func (p *PostgresPersistence) GetTransactionByID(ctx context.Context, txID uuid.UUID) (*domaintx.Transaction, error) {
	var m TransactionModel
	if err := p.db.WithContext(ctx).Where("id = ?", txID.String()).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewNotFoundError("transaction not found")
		}
		return nil, err
	}
	return &domaintx.Transaction{
		ID:                uuid.MustParse(m.ID),
		UserID:            uuid.MustParse(m.UserID),
		Type:              domaintx.Type(m.Type),
		Amount:            m.Amount,
		Currency:          m.Currency,
		Status:            domaintx.Status(m.Status),
		ProviderID:        uuid.MustParse(m.ProviderID),
		ExternalReference: m.ExternalReference,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}, nil
}

func (p *PostgresPersistence) ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domaintx.Transaction, error) {
	var rows []TransactionModel
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID.String()).Order("created_at desc").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domaintx.Transaction, 0, len(rows))
	for _, r := range rows {
		t := &domaintx.Transaction{
			ID:       uuid.MustParse(r.ID),
			UserID:   uuid.MustParse(r.UserID),
			Type:     domaintx.Type(r.Type),
			Amount:   r.Amount,
			Currency: r.Currency,
			Status:   domaintx.Status(r.Status),
			ProviderID: func() uuid.UUID {
				if r.ProviderID == "" {
					return uuid.Nil
				}
				return uuid.MustParse(r.ProviderID)
			}(),
			ExternalReference: r.ExternalReference,
			CreatedAt:         r.CreatedAt,
			UpdatedAt:         r.UpdatedAt,
		}
		out = append(out, t)
	}
	return out, nil
}

// Idempotency
func (p *PostgresPersistence) GetIdempotencyRecord(ctx context.Context, userID uuid.UUID, key string) (*payments.IdempotencyRecord, error) {
	var m IdempotencyModel
	if err := p.db.WithContext(ctx).Where("user_id = ? AND key = ?", userID.String(), key).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	rid, _ := uuid.Parse(m.RequestID)
	return &payments.IdempotencyRecord{UserID: uuid.MustParse(m.UserID), Key: m.Key, RequestID: rid, Response: m.Response, CreatedAt: m.CreatedAt}, nil
}

func (p *PostgresPersistence) CreateIdempotencyRecord(ctx context.Context, record *payments.IdempotencyRecord) error {
	createdAt := record.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	m := IdempotencyModel{
		ID:        uuid.NewString(),
		UserID:    record.UserID.String(),
		Key:       record.Key,
		RequestID: record.RequestID.String(),
		Response:  record.Response,
		CreatedAt: createdAt,
	}
	return p.db.WithContext(ctx).Create(&m).Error
}

// Outbox
func (p *PostgresPersistence) CreateEvent(ctx context.Context, event *appoutbox.OutboxEvent) error {
	m := OutboxModel{ID: event.ID.String(), EventType: event.EventType, Payload: event.Payload, CreatedAt: event.CreatedAt}
	if event.SentAt != nil {
		m.SentAt = event.SentAt
	}
	return p.db.WithContext(ctx).Create(&m).Error
}

func (p *PostgresPersistence) GetPendingEvents(ctx context.Context, limit int) ([]*appoutbox.OutboxEvent, error) {
	var rows []OutboxModel
	if err := p.db.WithContext(ctx).Where("sent_at IS NULL").Order("created_at").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*appoutbox.OutboxEvent, 0, len(rows))
	for _, r := range rows {
		out = append(out, &appoutbox.OutboxEvent{ID: uuid.MustParse(r.ID), EventType: r.EventType, Payload: r.Payload, CreatedAt: r.CreatedAt, SentAt: r.SentAt})
	}
	return out, nil
}

func (p *PostgresPersistence) MarkEventAsSent(ctx context.Context, eventID uuid.UUID) error {
	return p.db.WithContext(ctx).Model(&OutboxModel{}).Where("id = ?", eventID.String()).Update("sent_at", time.Now()).Error
}

// Compile-time interface checks
var _ payments.PaymentRepository = (*PostgresPersistence)(nil)
var _ payments.IdempotencyRepository = (*PostgresPersistence)(nil)
var _ appoutbox.OutboxRepository = (*PostgresPersistence)(nil)
