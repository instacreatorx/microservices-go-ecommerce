package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ecommerce/payment/internal/domain"
	"github.com/ecommerce/payment/internal/repository"
	"github.com/ecommerce/pkg/broker"
	"github.com/google/uuid"
)

type PaymentUsecase struct {
	repo   repository.PaymentRepository
	broker *broker.RabbitMQ
}

func NewPaymentUsecase(repo repository.PaymentRepository, b *broker.RabbitMQ) *PaymentUsecase {
	return &PaymentUsecase{
		repo:   repo,
		broker: b,
	}
}

func (uc *PaymentUsecase) Process(ctx context.Context, req *domain.ProcessPaymentRequest) (*domain.Payment, error) {
	existing, _ := uc.repo.GetByOrderID(ctx, req.OrderID)
	if existing != nil {
		return nil, errors.New("payment already exists for this order")
	}

	txID := uuid.New().String()

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        domain.PaymentCompleted,
		Method:        req.Method,
		TransactionID: txID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	event := map[string]interface{}{
		"type":     "payment.completed",
		"order_id": req.OrderID,
		"status":   "completed",
	}
	body, _ := json.Marshal(event)
	if err := uc.broker.Publish("ecommerce", "payment.completed", body); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUsecase) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *PaymentUsecase) Refund(ctx context.Context, id string) (*domain.Payment, error) {
	payment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if payment.Status != domain.PaymentCompleted {
		return nil, errors.New("payment cannot be refunded")
	}

	if err := uc.repo.UpdateStatus(ctx, id, domain.PaymentRefunded); err != nil {
		return nil, err
	}
	payment.Status = domain.PaymentRefunded

	event := map[string]interface{}{
		"type":     "payment.refunded",
		"order_id": payment.OrderID,
	}
	body, _ := json.Marshal(event)
	uc.broker.Publish("ecommerce", "payment.refunded", body)

	return payment, nil
}
