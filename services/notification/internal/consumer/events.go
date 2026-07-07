package consumer

import (
	"context"
	"encoding/json"

	"github.com/ecommerce/pkg/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type EventConsumer struct {
	broker *broker.RabbitMQ
	logger *zap.Logger
}

func NewEventConsumer(b *broker.RabbitMQ, logger *zap.Logger) *EventConsumer {
	return &EventConsumer{
		broker: b,
		logger: logger,
	}
}

func (c *EventConsumer) Start(ctx context.Context) {
	if err := c.broker.DeclareExchange("ecommerce", "topic"); err != nil {
		c.logger.Fatal("failed to declare exchange", zap.Error(err))
	}

	queue, err := c.broker.DeclareQueue("notification.all", nil)
	if err != nil {
		c.logger.Fatal("failed to declare queue", zap.Error(err))
	}

	if err := c.broker.BindQueue(queue.Name, "#", "ecommerce"); err != nil {
		c.logger.Fatal("failed to bind queue", zap.Error(err))
	}

	msgs, err := c.broker.Consume(queue.Name)
	if err != nil {
		c.logger.Fatal("failed to consume", zap.Error(err))
	}

	c.logger.Info("starting notification consumer")
	forever := make(chan struct{})

	go func() {
		for msg := range msgs {
			c.processMessage(ctx, msg)
		}
	}()

	<-forever
}

func (c *EventConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("failed to unmarshal event", zap.Error(err))
		msg.Nack(false, false)
		return
	}

	eventType, _ := event["type"].(string)

	switch eventType {
	case "order.created":
		c.logger.Info("sending order confirmation email",
			zap.String("order_id", toString(event["order_id"])),
		)

	case "order.shipped":
		c.logger.Info("sending shipping notification email",
			zap.String("order_id", toString(event["order_id"])),
		)

	case "payment.completed":
		c.logger.Info("sending payment confirmation email",
			zap.String("order_id", toString(event["order_id"])),
		)

	case "payment.failed":
		c.logger.Info("sending payment failure email",
			zap.String("order_id", toString(event["order_id"])),
		)

	case "user.registered":
		c.logger.Info("sending welcome email",
			zap.String("user_id", toString(event["user_id"])),
		)

	default:
		c.logger.Info("unknown event type", zap.String("type", eventType))
	}

	msg.Ack(false)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}
