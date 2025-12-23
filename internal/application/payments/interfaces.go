package payments

import (
	"context"
	"draftea-challenge/internal/domain/payment"
	"draftea-challenge/internal/domain/transaction"
	"time"

	"github.com/google/uuid"
)

// PaymentRepository define la interfaz para acceder a datos de pagos y transacciones.
type PaymentRepository interface {
	CreateTransaction(ctx context.Context, tx *transaction.Transaction) error
	UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status transaction.Status) error
	GetTransactionByID(ctx context.Context, txID uuid.UUID) (*transaction.Transaction, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error)
}

// PaymentGateway define la interfaz para interactuar con la pasarela de pago externa.
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, p *payment.Payment) (string, error) // retorna status o error
}

// IdempotencyRepository define la interfaz para manejar claves de idempotencia.
type IdempotencyRepository interface {
	GetIdempotencyRecord(ctx context.Context, userID uuid.UUID, key string) (*IdempotencyRecord, error)
	CreateIdempotencyRecord(ctx context.Context, record *IdempotencyRecord) error
}

// IdempotencyRecord representa un registro de idempotencia.
type IdempotencyRecord struct {
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	RequestID uuid.UUID `json:"request_id"` // ID de la transacción o pago
	Response  string    `json:"response"`   // JSON de la respuesta original
	CreatedAt time.Time `json:"created_at"`
}
