package wallets

import (
	"context"
	"draftea-challenge/internal/domain/errors"
	"draftea-challenge/internal/domain/transaction"
	"draftea-challenge/internal/domain/wallet"

	"github.com/google/uuid"
)

// GetBalanceService obtiene el saldo de una wallet.
type GetBalanceService struct {
	walletRepo WalletRepository
}

// NewGetBalanceService crea una nueva instancia.
func NewGetBalanceService(walletRepo WalletRepository) *GetBalanceService {
	return &GetBalanceService{walletRepo: walletRepo}
}

// GetBalanceResponse representa la respuesta de saldo.
type GetBalanceResponse struct {
	UserID   uuid.UUID        `json:"user_id"`
	Balances map[string]int64 `json:"balances"`
	Name     string           `json:"name,omitempty"`
}

// GetBalance obtiene el saldo del usuario.
func (s *GetBalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (*GetBalanceResponse, error) {
	w, err := s.walletRepo.GetWallet(ctx, userID)
	if err != nil {
		if isNotFoundError(err) {
			return &GetBalanceResponse{UserID: userID, Balances: make(map[string]int64)}, nil
		}
		return nil, err
	}
	return &GetBalanceResponse{
		UserID:   w.UserID,
		Balances: w.Balances,
		Name:     w.Name,
	}, nil
}

// GetTransactionsService obtiene el historial de transacciones.
type GetTransactionsService struct {
	paymentRepo interface {
		ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error)
	}
}

// NewGetTransactionsService crea una nueva instancia.
func NewGetTransactionsService(paymentRepo interface {
	ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error)
}) *GetTransactionsService {
	return &GetTransactionsService{paymentRepo: paymentRepo}
}

// GetTransactionsRequest representa la solicitud.
type GetTransactionsRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// GetTransactionsResponse representa la respuesta.
type GetTransactionsResponse struct {
	Transactions []*transaction.Transaction `json:"transactions"`
	Total        int                        `json:"total"`
}

// GetTransactions obtiene las transacciones del usuario.
func (s *GetTransactionsService) GetTransactions(ctx context.Context, req *GetTransactionsRequest) (*GetTransactionsResponse, error) {
	txs, err := s.paymentRepo.ListTransactions(ctx, req.UserID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	// Nota: Para total, asumir que el repo lo maneja o simplificar
	return &GetTransactionsResponse{
		Transactions: txs,
		Total:        len(txs), // Simplificado
	}, nil
}

// TopUpService credits balances for testing purposes.
type TopUpService struct {
	walletRepo WalletRepository
	txRepo     interface {
		CreateTransaction(ctx context.Context, tx *transaction.Transaction) error
		UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status transaction.Status) error
	}
}

// NewTopUpService creates a new top-up service.
func NewTopUpService(walletRepo WalletRepository, txRepo interface {
	CreateTransaction(ctx context.Context, tx *transaction.Transaction) error
	UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status transaction.Status) error
}) *TopUpService {
	return &TopUpService{walletRepo: walletRepo, txRepo: txRepo}
}

// TopUpRequest represents a top-up request.
type TopUpRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Amount   int64     `json:"amount"`
	Currency string    `json:"currency"`
}

// TopUpResponse represents a top-up response.
type TopUpResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Balance       int64     `json:"balance"`
}

// TopUp credits the wallet balance and records a transaction.
func (s *TopUpService) TopUp(ctx context.Context, req *TopUpRequest) (*TopUpResponse, error) {
	if req.Amount <= 0 {
		return nil, errors.NewValidationError("amount must be positive", map[string]interface{}{"amount": req.Amount})
	}
	if req.Currency == "" {
		return nil, errors.NewValidationError("currency cannot be empty", nil)
	}

	w, err := s.walletRepo.GetWallet(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if err := w.Credit(req.Currency, req.Amount); err != nil {
		return nil, err
	}
	if err := s.walletRepo.UpdateBalance(ctx, req.UserID, req.Currency, w.GetBalance(req.Currency)); err != nil {
		return nil, err
	}

	tx, err := transaction.NewTransaction(req.UserID, transaction.TypeTopUp, req.Amount, req.Currency, uuid.Nil, "top-up")
	if err != nil {
		return nil, err
	}
	if err := tx.UpdateStatus(transaction.StatusApproved); err != nil {
		return nil, err
	}
	if err := s.txRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}
	if err := s.txRepo.UpdateTransactionStatus(ctx, tx.ID, transaction.StatusApproved); err != nil {
		return nil, err
	}

	return &TopUpResponse{
		TransactionID: tx.ID,
		Balance:       w.GetBalance(req.Currency),
	}, nil
}

// ListWalletsService lists wallets for testing visibility.
type ListWalletsService struct {
	walletRepo WalletRepository
}

// NewListWalletsService creates a new list service.
func NewListWalletsService(walletRepo WalletRepository) *ListWalletsService {
	return &ListWalletsService{walletRepo: walletRepo}
}

// WalletSummary represents a wallet entry.
type WalletSummary struct {
	UserID   uuid.UUID        `json:"user_id"`
	Balances map[string]int64 `json:"balances"`
}

// ListWalletsRequest represents the list request.
type ListWalletsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ListWalletsResponse represents the list response.
type ListWalletsResponse struct {
	Wallets []WalletSummary `json:"wallets"`
	Total   int             `json:"total"`
}

// ListWallets returns wallets with balances.
func (s *ListWalletsService) ListWallets(ctx context.Context, req *ListWalletsRequest) (*ListWalletsResponse, error) {
	walletsList, total, err := s.walletRepo.ListWallets(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	out := make([]WalletSummary, 0, len(walletsList))
	for _, w := range walletsList {
		out = append(out, WalletSummary{
			UserID:   w.UserID,
			Balances: w.Balances,
		})
	}
	return &ListWalletsResponse{
		Wallets: out,
		Total:   total,
	}, nil
}

// CreateWalletService creates a wallet for a user.
type CreateWalletService struct {
	walletRepo WalletRepository
}

// NewCreateWalletService creates a new create wallet service.
func NewCreateWalletService(walletRepo WalletRepository) *CreateWalletService {
	return &CreateWalletService{walletRepo: walletRepo}
}

// CreateWalletRequest represents a create wallet request.
type CreateWalletRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

// CreateWalletResponse represents the create response.
type CreateWalletResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name,omitempty"`
}

// CreateWallet creates a new wallet if it does not exist.
func (s *CreateWalletService) CreateWallet(ctx context.Context, req *CreateWalletRequest) (*CreateWalletResponse, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.NewValidationError("user_id cannot be nil", nil)
	}
	if len(req.Name) > 20 {
		return nil, errors.NewValidationError("name must be at most 20 characters", map[string]interface{}{"name": req.Name})
	}

	if _, err := s.walletRepo.GetWallet(ctx, req.UserID); err == nil {
		return nil, errors.NewValidationError("wallet already exists", map[string]interface{}{"user_id": req.UserID.String()})
	} else if !isNotFoundError(err) {
		return nil, err
	}

	w, err := wallet.NewWalletWithName(req.UserID, req.Name)
	if err != nil {
		return nil, err
	}
	if err := s.walletRepo.CreateWallet(ctx, w); err != nil {
		return nil, err
	}

	return &CreateWalletResponse{
		UserID: w.UserID,
		Name:   w.Name,
	}, nil
}

// isNotFoundError verifica si es error de no encontrado.
func isNotFoundError(err error) bool {
	if domErr, ok := err.(errors.Error); ok && domErr.Code == errors.CodeNotFound {
		return true
	}
	return false
}
