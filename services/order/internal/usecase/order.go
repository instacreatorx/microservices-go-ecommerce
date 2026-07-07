package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/ecommerce/order/internal/domain"
	"github.com/ecommerce/order/internal/repository"
	"github.com/ecommerce/pkg/broker"
	"github.com/google/uuid"
)

type OrderUsecase struct {
	repo    repository.OrderRepository
	broker  *broker.RabbitMQ
}

func NewOrderUsecase(repo repository.OrderRepository, b *broker.RabbitMQ) *OrderUsecase {
	return &OrderUsecase{
		repo:   repo,
		broker: b,
	}
}

func (uc *OrderUsecase) Create(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	orderID := uuid.New().String()
	var items []domain.OrderItem
	var total float64

	for _, item := range req.Items {
		subtotal := item.Price * float64(item.Quantity)
		total += subtotal
		items = append(items, domain.OrderItem{
			ID:          uuid.New().String(),
			OrderID:     orderID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	order := &domain.Order{
		ID:              orderID,
		UserID:          req.UserID,
		Items:           items,
		Total:           total,
		Status:          domain.StatusPending,
		ShippingAddress: req.ShippingAddress,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	event := map[string]interface{}{
		"type":    "order.created",
		"order_id": orderID,
		"user_id":  req.UserID,
		"total":    total,
	}
	body, _ := json.Marshal(event)
	if err := uc.broker.Publish("ecommerce", "order.created", body); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUsecase) GetByID(ctx context.Context, id, userID string) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id, userID)
}

func (uc *OrderUsecase) List(ctx context.Context, userID string, page, pageSize int) ([]*domain.Order, int, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	orders, total, err := uc.repo.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return orders, total, totalPages, nil
}

func (uc *OrderUsecase) Cancel(ctx context.Context, id, userID string) error {
	order, err := uc.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if order.Status == domain.StatusCancelled || order.Status == domain.StatusDelivered {
		return errors.New("order cannot be cancelled")
	}

	if err := uc.repo.UpdateStatus(ctx, id, domain.StatusCancelled); err != nil {
		return err
	}

	event := map[string]interface{}{
		"type":    "order.cancelled",
		"order_id": id,
		"user_id":  userID,
	}
	body, _ := json.Marshal(event)
	uc.broker.Publish("ecommerce", "order.cancelled", body)

	return nil
}

func (uc *OrderUsecase) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}
