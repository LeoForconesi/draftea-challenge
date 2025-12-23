package builders

import (
	"draftea-challenge/internal/application/payments"

	"github.com/google/uuid"
)

// PaymentRequestBuilder builds valid payment requests for tests.
type PaymentRequestBuilder struct {
	req payments.ProcessPaymentRequest
}

func NewPaymentRequestBuilder() *PaymentRequestBuilder {
	return &PaymentRequestBuilder{
		req: payments.ProcessPaymentRequest{
			UserID:            uuid.New(),
			ProviderID:        uuid.New(),
			ExternalReference: "invoice-123",
			Amount:            1000,
			Currency:          "USD",
			IdempotencyKey:    "idem-123",
		},
	}
}

func (b *PaymentRequestBuilder) WithUserID(id uuid.UUID) *PaymentRequestBuilder {
	b.req.UserID = id
	return b
}

func (b *PaymentRequestBuilder) WithProviderID(id uuid.UUID) *PaymentRequestBuilder {
	b.req.ProviderID = id
	return b
}

func (b *PaymentRequestBuilder) WithExternalReference(ref string) *PaymentRequestBuilder {
	b.req.ExternalReference = ref
	return b
}

func (b *PaymentRequestBuilder) WithAmount(amount int64) *PaymentRequestBuilder {
	b.req.Amount = amount
	return b
}

func (b *PaymentRequestBuilder) WithCurrency(currency string) *PaymentRequestBuilder {
	b.req.Currency = currency
	return b
}

func (b *PaymentRequestBuilder) WithIdempotencyKey(key string) *PaymentRequestBuilder {
	b.req.IdempotencyKey = key
	return b
}

func (b *PaymentRequestBuilder) Build() *payments.ProcessPaymentRequest {
	copyReq := b.req
	return &copyReq
}
