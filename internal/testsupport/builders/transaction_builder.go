package builders

import (
	"draftea-challenge/internal/domain/transaction"

	"github.com/google/uuid"
)

// TransactionBuilder builds transactions for tests.
type TransactionBuilder struct {
	userID            uuid.UUID
	txType            transaction.Type
	amount            int64
	currency          string
	providerID        uuid.UUID
	externalReference string
	status            transaction.Status
	statusSet         bool
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		userID:            uuid.New(),
		txType:            transaction.TypePayment,
		amount:            1000,
		currency:          "USD",
		providerID:        uuid.New(),
		externalReference: "invoice-123",
	}
}

func (b *TransactionBuilder) WithUserID(id uuid.UUID) *TransactionBuilder {
	b.userID = id
	return b
}

func (b *TransactionBuilder) WithType(txType transaction.Type) *TransactionBuilder {
	b.txType = txType
	return b
}

func (b *TransactionBuilder) WithAmount(amount int64) *TransactionBuilder {
	b.amount = amount
	return b
}

func (b *TransactionBuilder) WithCurrency(currency string) *TransactionBuilder {
	b.currency = currency
	return b
}

func (b *TransactionBuilder) WithProviderID(id uuid.UUID) *TransactionBuilder {
	b.providerID = id
	return b
}

func (b *TransactionBuilder) WithExternalReference(ref string) *TransactionBuilder {
	b.externalReference = ref
	return b
}

func (b *TransactionBuilder) WithStatus(status transaction.Status) *TransactionBuilder {
	b.status = status
	b.statusSet = true
	return b
}

func (b *TransactionBuilder) Build() (*transaction.Transaction, error) {
	tx, err := transaction.NewTransaction(b.userID, b.txType, b.amount, b.currency, b.providerID, b.externalReference)
	if err != nil {
		return nil, err
	}
	if b.statusSet {
		if err := tx.UpdateStatus(b.status); err != nil {
			return nil, err
		}
	}
	return tx, nil
}
