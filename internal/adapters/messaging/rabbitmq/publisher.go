package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Publisher publishes messages to RabbitMQ.
type Publisher struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	confirmations  <-chan amqp.Confirmation
	exchange       string
	confirmTimeout time.Duration
	log            *zap.Logger
}

// Config configures RabbitMQ connections and topology.
type Config struct {
	URL                   string
	Exchange              string
	MetricsQueue          string
	AuditQueue            string
	PublishConfirmTimeout time.Duration
}

// NewPublisher creates a publisher and declares exchange/queues.
func NewPublisher(cfg Config, log *zap.Logger) (*Publisher, func() error, error) {
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

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	confirmations := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	publisher := &Publisher{
		conn:           conn,
		channel:        ch,
		confirmations:  confirmations,
		exchange:       cfg.Exchange,
		confirmTimeout: cfg.PublishConfirmTimeout,
		log:            log,
	}

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

	return publisher, cleanup, nil
}

// Publish sends a message to the exchange with the given routing key.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	if exchange == "" {
		exchange = p.exchange
	}

	if err := p.channel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		return err
	}

	if p.confirmTimeout <= 0 {
		return nil
	}

	select {
	case confirm := <-p.confirmations:
		if !confirm.Ack {
			return fmt.Errorf("publish not acknowledged")
		}
		return nil
	case <-time.After(p.confirmTimeout):
		return fmt.Errorf("publish confirm timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func setupTopology(ch *amqp.Channel, cfg Config) error {
	if err := ch.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	if cfg.MetricsQueue != "" {
		if _, err := ch.QueueDeclare(cfg.MetricsQueue, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(cfg.MetricsQueue, "payment.*", cfg.Exchange, false, nil); err != nil {
			return err
		}
	}

	if cfg.AuditQueue != "" {
		if _, err := ch.QueueDeclare(cfg.AuditQueue, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(cfg.AuditQueue, "payment.*", cfg.Exchange, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(cfg.AuditQueue, "refund.*", cfg.Exchange, false, nil); err != nil {
			return err
		}
	}

	return nil
}
