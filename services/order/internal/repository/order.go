package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce/order/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id, userID string) (*domain.Order, error)
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
}

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	orderQuery := `INSERT INTO orders (id, user_id, total, status, shipping_address, created_at, updated_at)
	               VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, orderQuery,
		order.ID, order.UserID, order.Total, order.Status,
		order.ShippingAddress, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (id, order_id, product_id, product_name, price, quantity, subtotal)
	              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemQuery,
			item.ID, item.OrderID, item.ProductID, item.ProductName,
			item.Price, item.Quantity, item.Subtotal)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id, userID string) (*domain.Order, error) {
	orderQuery := `SELECT id, user_id, total, status, shipping_address, created_at, updated_at
	               FROM orders WHERE id = $1 AND user_id = $2`
	row := r.pool.QueryRow(ctx, orderQuery, id, userID)

	order := &domain.Order{}
	err := row.Scan(&order.ID, &order.UserID, &order.Total, &order.Status,
		&order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	itemsQuery := `SELECT id, order_id, product_id, product_name, price, quantity, subtotal
	               FROM order_items WHERE order_id = $1`
	rows, err := r.pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item := domain.OrderItem{}
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Price, &item.Quantity, &item.Subtotal); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	query := `SELECT id, user_id, total, status, shipping_address, created_at, updated_at
	          FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status,
			&o.ShippingAddress, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}
