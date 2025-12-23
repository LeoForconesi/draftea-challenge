package payments

import (
	"context"
	"draftea-challenge/internal/application/outbox"
	"draftea-challenge/internal/application/ports"
	"draftea-challenge/internal/application/wallets"
	"draftea-challenge/internal/domain/errors"
	"draftea-challenge/internal/domain/payment"
	"draftea-challenge/internal/domain/transaction"
	"draftea-challenge/internal/domain/wallet"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// PaymentService orquesta el flujo de pagos.
type PaymentService struct {
	paymentRepo     PaymentRepository
	walletRepo      wallets.WalletRepository
	gateway         PaymentGateway
	idempotencyRepo IdempotencyRepository
	outboxRepo      outbox.OutboxRepository
	idGen           ports.IDGenerator
	clock           ports.Clock
}

// NewPaymentService crea una nueva instancia de PaymentService.
func NewPaymentService(
	paymentRepo PaymentRepository,
	walletRepo wallets.WalletRepository,
	gateway PaymentGateway,
	idempotencyRepo IdempotencyRepository,
	outboxRepo outbox.OutboxRepository,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		walletRepo:      walletRepo,
		gateway:         gateway,
		idempotencyRepo: idempotencyRepo,
		outboxRepo:      outboxRepo,
		idGen:           idGen,
		clock:           clock,
	}
}

// ProcessPaymentRequest representa la solicitud de pago.
type ProcessPaymentRequest struct {
	UserID            uuid.UUID `json:"user_id"`
	ProviderID        uuid.UUID `json:"provider_id"`
	ExternalReference string    `json:"external_reference"`
	Amount            int64     `json:"amount"`
	Currency          string    `json:"currency"`
	IdempotencyKey    string    `json:"idempotency_key"`
}

// ProcessPaymentResponse representa la respuesta.
type ProcessPaymentResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Status        string    `json:"status"`
}

// ProcessPayment ejecuta el flujo de pago con idempotencia.
func (s *PaymentService) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
	// Validar idempotencia
	if req.IdempotencyKey != "" {
		record, err := s.idempotencyRepo.GetIdempotencyRecord(ctx, req.UserID, req.IdempotencyKey)
		if err != nil && !isNotFoundError(err) {
			return nil, err
		}
		if record != nil {
			// Retornar respuesta original
			var resp ProcessPaymentResponse
			if err := json.Unmarshal([]byte(record.Response), &resp); err != nil {
				return nil, errors.NewInternalError("failed to unmarshal idempotency response")
			}
			return &resp, nil
		}
	}

	// Crear entidad Payment
	p, err := payment.NewPayment(req.UserID, req.ProviderID, req.ExternalReference, req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	// Iniciar transacción DB para lock y debit
	// Nota: Asumimos que el repo maneja transacciones; en implementación real, usar tx.Begin()
	w, err := s.walletRepo.GetWallet(ctx, req.UserID)
	if err != nil {
		if isNotFoundError(err) {
			w, _ = wallet.NewWallet(req.UserID)
			if err := s.walletRepo.CreateWallet(ctx, w); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Debit (con lock implícito en repo)
	if err := w.Debit(req.Currency, req.Amount); err != nil {
		return nil, err
	}
	if err := s.walletRepo.UpdateBalance(ctx, req.UserID, req.Currency, w.GetBalance(req.Currency)); err != nil {
		return nil, err
	}

	// Crear transacción
	tx, err := transaction.NewTransaction(req.UserID, transaction.TypePayment, req.Amount, req.Currency, req.ProviderID, req.ExternalReference)
	if err != nil {
		return nil, err
	}
	if err := s.paymentRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	// Llamar a gateway
	status, err := s.gateway.ProcessPayment(ctx, p)
	if err != nil {
		// Gateway error: refund interno
		if refundErr := s.refundInternal(ctx, tx, w); refundErr != nil {
			// Log error, pero continuar
		}
		tx.UpdateStatus(transaction.StatusFailed)
		s.paymentRepo.UpdateTransactionStatus(ctx, tx.ID, transaction.StatusFailed)
		return nil, errors.NewGatewayError("gateway processing failed")
	}

	// Finalizar basado en status
	switch status {
	case "approved":
		tx.UpdateStatus(transaction.StatusApproved)
	case "declined":
		tx.UpdateStatus(transaction.StatusDeclined)
		// Refund
		s.refundInternal(ctx, tx, w)
	default:
		tx.UpdateStatus(transaction.StatusFailed)
		s.refundInternal(ctx, tx, w)
	}
	s.paymentRepo.UpdateTransactionStatus(ctx, tx.ID, tx.Status)

	// Crear evento outbox
	event := &outbox.OutboxEvent{
		ID:        s.idGen.New(),
		EventType: fmt.Sprintf("payment.%s", status),
		Payload:   fmt.Sprintf(`{"transaction_id":"%s","status":"%s"}`, tx.ID, status),
		CreatedAt: s.clock.Now(),
	}
	s.outboxRepo.CreateEvent(ctx, event)

	// Guardar idempotencia
	resp := &ProcessPaymentResponse{
		TransactionID: tx.ID,
		Status:        string(tx.Status),
	}
	respJSON, _ := json.Marshal(resp)
	record := &IdempotencyRecord{
		UserID:    req.UserID,
		Key:       req.IdempotencyKey,
		RequestID: tx.ID,
		Response:  string(respJSON),
		CreatedAt: s.clock.Now(),
	}
	s.idempotencyRepo.CreateIdempotencyRecord(ctx, record)

	return resp, nil
}

// refundInternal realiza un reembolso interno.
func (s *PaymentService) refundInternal(ctx context.Context, tx *transaction.Transaction, w *wallet.Wallet) error {
	w.Credit(tx.Currency, tx.Amount)
	s.walletRepo.UpdateBalance(ctx, tx.UserID, tx.Currency, w.GetBalance(tx.Currency))
	refundTx, _ := transaction.NewTransaction(tx.UserID, transaction.TypeRefund, tx.Amount, tx.Currency, tx.ProviderID, tx.ExternalReference)
	s.paymentRepo.CreateTransaction(ctx, refundTx)
	refundTx.UpdateStatus(transaction.StatusApproved)
	s.paymentRepo.UpdateTransactionStatus(ctx, refundTx.ID, transaction.StatusApproved)
	return nil
}

// isNotFoundError verifica si es error de no encontrado.
func isNotFoundError(err error) bool {
	if domErr, ok := err.(errors.Error); ok && domErr.Code == errors.CodeNotFound {
		return true
	}
	return false
}
