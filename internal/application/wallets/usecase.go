package wallets

import (
	"context"
	"draftea-challenge/internal/domain/errors"
	"draftea-challenge/internal/domain/transaction"

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

// isNotFoundError verifica si es error de no encontrado.
func isNotFoundError(err error) bool {
	if domErr, ok := err.(errors.Error); ok && domErr.Code == errors.CodeNotFound {
		return true
	}
	return false
}
