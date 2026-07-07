package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce/payment/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error
}

type PostgresPaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentRepository(pool *pgxpool.Pool) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{pool: pool}
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `INSERT INTO payments (id, order_id, user_id, amount, currency, status, method, transaction_id, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.pool.Exec(ctx, query,
		payment.ID, payment.OrderID, payment.UserID, payment.Amount,
		payment.Currency, payment.Status, payment.Method,
		payment.TransactionID, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `SELECT id, order_id, user_id, amount, currency, status, method, transaction_id, created_at, updated_at
	          FROM payments WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	p := &domain.Payment{}
	err := row.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency,
		&p.Status, &p.Method, &p.TransactionID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}

func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, user_id, amount, currency, status, method, transaction_id, created_at, updated_at
	          FROM payments WHERE order_id = $1`
	row := r.pool.QueryRow(ctx, query, orderID)

	p := &domain.Payment{}
	err := row.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency,
		&p.Status, &p.Method, &p.TransactionID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("get payment by order: %w", err)
	}
	return p, nil
}

func (r *PostgresPaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, time.Now(), id)
	return err
}
