package relay

import (
	"context"
	"fmt"

	"draftea-challenge/internal/application/outbox"
	"draftea-challenge/internal/application/ports"
)

// Relay publishes pending outbox events to the message broker.
type Relay struct {
	repo      outbox.OutboxRepository
	publisher outbox.MessagePublisher
	clock     ports.Clock
	batchSize int
}

// NewRelay creates a new outbox relay.
func NewRelay(repo outbox.OutboxRepository, publisher outbox.MessagePublisher, clock ports.Clock, batchSize int) *Relay {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Relay{
		repo:      repo,
		publisher: publisher,
		clock:     clock,
		batchSize: batchSize,
	}
}

// ProcessOnce publishes a single batch of pending events.
func (r *Relay) ProcessOnce(ctx context.Context) error {
	events, err := r.repo.GetPendingEvents(ctx, r.batchSize)
	if err != nil {
		return err
	}
	for _, event := range events {
		routingKey := event.EventType
		if err := r.publisher.Publish(ctx, "payments.events", routingKey, []byte(event.Payload)); err != nil {
			return fmt.Errorf("publish outbox event %s: %w", event.ID, err)
		}
		now := r.clock.Now()
		event.SentAt = &now
		if err := r.repo.MarkEventAsSent(ctx, event.ID); err != nil {
			return fmt.Errorf("mark outbox event %s as sent: %w", event.ID, err)
		}
	}
	return nil
}
