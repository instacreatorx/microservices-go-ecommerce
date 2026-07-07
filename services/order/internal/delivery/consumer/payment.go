package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ecommerce/order/internal/domain"
	"github.com/ecommerce/order/internal/usecase"
	"github.com/ecommerce/pkg/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type PaymentConsumer struct {
	uc     *usecase.OrderUsecase
	broker *broker.RabbitMQ
	logger *zap.Logger
}

func NewPaymentConsumer(uc *usecase.OrderUsecase, b *broker.RabbitMQ, logger *zap.Logger) *PaymentConsumer {
	return &PaymentConsumer{
		uc:     uc,
		broker: b,
		logger: logger,
	}
}

func (c *PaymentConsumer) Start(ctx context.Context) {
	if err := c.broker.DeclareExchange("ecommerce", "topic"); err != nil {
		c.logger.Fatal("failed to declare exchange", zap.Error(err))
	}

	queue, err := c.broker.DeclareQueue("order.payment", nil)
	if err != nil {
		c.logger.Fatal("failed to declare queue", zap.Error(err))
	}

	if err := c.broker.BindQueue(queue.Name, "payment.*", "ecommerce"); err != nil {
		c.logger.Fatal("failed to bind queue", zap.Error(err))
	}

	msgs, err := c.broker.Consume(queue.Name)
	if err != nil {
		c.logger.Fatal("failed to consume", zap.Error(err))
	}

	c.logger.Info("starting payment consumer")
	forever := make(chan struct{})

	go func() {
		for msg := range msgs {
			c.processMessage(ctx, msg)
		}
	}()

	<-forever
}

func (c *PaymentConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var event struct {
		Type    string `json:"type"`
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("failed to unmarshal payment event", zap.Error(err))
		msg.Nack(false, false)
		return
	}

	switch event.Type {
	case "payment.completed":
		if err := c.uc.UpdateStatus(ctx, event.OrderID, domain.StatusProcessing); err != nil {
			c.logger.Error("failed to update order status", zap.Error(err))
			msg.Nack(false, true)
			return
		}
		c.logger.Info("order status updated to processing", zap.String("order_id", event.OrderID))

	case "payment.failed":
		if err := c.uc.UpdateStatus(ctx, event.OrderID, domain.StatusFailed); err != nil {
			c.logger.Error("failed to update order status", zap.Error(err))
			msg.Nack(false, true)
			return
		}
		c.logger.Info("order status updated to failed", zap.String("order_id", event.OrderID))
	}

	msg.Ack(false)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
