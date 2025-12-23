package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OutboxRepository define la interfaz para manejar eventos de outbox.
type OutboxRepository interface {
	CreateEvent(ctx context.Context, event *OutboxEvent) error
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)
	MarkEventAsSent(ctx context.Context, eventID uuid.UUID) error
}

// MessagePublisher define la interfaz para publicar mensajes a RabbitMQ.
type MessagePublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// OutboxEvent representa un evento en la tabla de outbox.
type OutboxEvent struct {
	ID        uuid.UUID  `json:"id"`
	EventType string     `json:"event_type"` // e.g., "payment.created"
	Payload   string     `json:"payload"`    // JSON del evento
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
}
