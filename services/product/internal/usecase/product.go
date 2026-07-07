package usecase

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ecommerce/product/internal/domain"
	"github.com/ecommerce/product/internal/repository"
	"github.com/google/uuid"
)

type ProductUsecase struct {
	repo repository.ProductRepository
}

func NewProductUsecase(repo repository.ProductRepository) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

func (uc *ProductUsecase) Create(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	if req.Name == "" || req.Price <= 0 {
		return nil, errors.New("invalid product data")
	}

	product := &domain.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *ProductUsecase) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *ProductUsecase) List(ctx context.Context, page, pageSize int, category string) ([]*domain.Product, int, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	products, total, err := uc.repo.List(ctx, offset, pageSize, category)
	if err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return products, total, totalPages, nil
}

func (uc *ProductUsecase) Update(ctx context.Context, req *domain.UpdateProductRequest) (*domain.Product, error) {
	product, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	product.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *ProductUsecase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *ProductUsecase) UpdateStock(ctx context.Context, id string, quantity int) (*domain.Product, error) {
	if err := uc.repo.UpdateStock(ctx, id, quantity); err != nil {
		return nil, err
	}
	return uc.repo.GetByID(ctx, id)
}
