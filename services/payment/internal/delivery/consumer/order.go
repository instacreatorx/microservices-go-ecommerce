package consumer

import (
	"context"
	"encoding/json"

	"github.com/ecommerce/payment/internal/domain"
	"github.com/ecommerce/payment/internal/usecase"
	"github.com/ecommerce/pkg/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type OrderConsumer struct {
	uc     *usecase.PaymentUsecase
	broker *broker.RabbitMQ
	logger *zap.Logger
}

func NewOrderConsumer(uc *usecase.PaymentUsecase, b *broker.RabbitMQ, logger *zap.Logger) *OrderConsumer {
	return &OrderConsumer{
		uc:     uc,
		broker: b,
		logger: logger,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) {
	if err := c.broker.DeclareExchange("ecommerce", "topic"); err != nil {
		c.logger.Fatal("failed to declare exchange", zap.Error(err))
	}

	queue, err := c.broker.DeclareQueue("payment.order", nil)
	if err != nil {
		c.logger.Fatal("failed to declare queue", zap.Error(err))
	}

	if err := c.broker.BindQueue(queue.Name, "order.*", "ecommerce"); err != nil {
		c.logger.Fatal("failed to bind queue", zap.Error(err))
	}

	msgs, err := c.broker.Consume(queue.Name)
	if err != nil {
		c.logger.Fatal("failed to consume", zap.Error(err))
	}

	c.logger.Info("starting order consumer")
	forever := make(chan struct{})

	go func() {
		for msg := range msgs {
			c.processMessage(ctx, msg)
		}
	}()

	<-forever
}

func (c *OrderConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var event struct {
		Type    string  `json:"type"`
		OrderID string  `json:"order_id"`
		UserID  string  `json:"user_id"`
		Total   float64 `json:"total"`
	}

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("failed to unmarshal order event", zap.Error(err))
		msg.Nack(false, false)
		return
	}

	switch event.Type {
	case "order.created":
		req := &domain.ProcessPaymentRequest{
			OrderID:  event.OrderID,
			UserID:   event.UserID,
			Amount:   event.Total,
			Currency: "USD",
			Method:   "card",
		}

		if _, err := c.uc.Process(ctx, req); err != nil {
			c.logger.Error("failed to process payment", zap.Error(err))

			event := map[string]interface{}{
				"type":     "payment.failed",
				"order_id": event.OrderID,
				"status":   "failed",
			}
			body, _ := json.Marshal(event)
			c.broker.Publish("ecommerce", "payment.failed", body)
		}

		msg.Ack(false)
	}

	msg.Ack(false)
}
