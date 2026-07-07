package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce/product/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	List(ctx context.Context, offset, limit int, category string) ([]*domain.Product, int, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, quantity int) error
}

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `INSERT INTO products (id, name, description, price, stock, category, image_url, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query,
		product.ID, product.Name, product.Description, product.Price,
		product.Stock, product.Category, product.ImageURL,
		product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
	          FROM products WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	product := &domain.Product{}
	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.Category, &product.ImageURL,
		&product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get product by id: %w", err)
	}
	return product, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, offset, limit int, category string) ([]*domain.Product, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM products"
	args := []interface{}{}

	if category != "" {
		countQuery += " WHERE category = $1"
		args = append(args, category)
	}
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	query := `SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
	          FROM products`
	queryArgs := []interface{}{}

	if category != "" {
		query += " WHERE category = $1"
		queryArgs = append(queryArgs, category)
	}
	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		p := &domain.Product{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price,
			&p.Stock, &p.Category, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `UPDATE products SET name=$1, description=$2, price=$3, category=$4,
	          image_url=$5, updated_at=$6 WHERE id=$7`
	_, err := r.pool.Exec(ctx, query,
		product.Name, product.Description, product.Price, product.Category,
		product.ImageURL, time.Now(), product.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) UpdateStock(ctx context.Context, id string, quantity int) error {
	query := `UPDATE products SET stock = stock + $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, quantity, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update stock: %w", err)
	}
	return nil
}
