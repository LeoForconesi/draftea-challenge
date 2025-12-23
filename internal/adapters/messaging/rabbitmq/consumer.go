package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Consumer consumes messages from a queue.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	log     *zap.Logger
}

// NewConsumer creates a consumer for the given queue.
func NewConsumer(cfg Config, queue string, log *zap.Logger) (*Consumer, func() error, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	if err := setupTopology(ch, cfg); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	consumer := &Consumer{conn: conn, channel: ch, queue: queue, log: log}
	cleanup := func() error {
		if err := ch.Close(); err != nil {
			log.Warn("failed to close rabbitmq channel", zap.Error(err))
		}
		if err := conn.Close(); err != nil {
			log.Warn("failed to close rabbitmq connection", zap.Error(err))
			return err
		}
		return nil
	}

	return consumer, cleanup, nil
}

// Start begins consuming messages with the handler.
func (c *Consumer) Start(ctx context.Context, handler func(context.Context, amqp.Delivery) error) error {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := handler(ctx, msg); err != nil {
				c.log.Warn("consumer handler error", zap.Error(err))
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}
