package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kelar1s/go-freight/internal/inventory/model"
)

//go:generate mockery --name=Cache --output=./mocks --outpkg=mocks --with-expecter=true
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

//go:generate mockery --name=ProductRepo --output=./mocks --outpkg=mocks --with-expecter=true
type ProductRepo interface {
	Create(ctx context.Context, product *model.Product) error
	GetMeta(ctx context.Context, id int64) (*model.ProductMeta, error)
	GetStock(ctx context.Context, id int64) (*model.ProductStock, error)
	ListByWarehouse(ctx context.Context, warehouseID int64) ([]model.Product, error)
	Delete(ctx context.Context, id int64) error
	AdjustQuantity(ctx context.Context, id int64, quantity int64) error
	Reserve(ctx context.Context, id int64, quantity int64) error
	Release(ctx context.Context, id int64, quantity int64) error
	CancelReservation(ctx context.Context, id int64, quantity int64) error
}

type ProductService struct {
	repo     ProductRepo
	cache    Cache
	cacheTTL time.Duration
}

func NewProductService(repo ProductRepo, cache Cache, cacheTTL time.Duration) *ProductService {
	return &ProductService{
		repo:     repo,
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

func productMetaCacheKey(id int64) string {
	return fmt.Sprintf("product:meta:%d", id)
}

func (s *ProductService) Create(ctx context.Context, warehouseID int64, name string, quantity int64) (*model.Product, error) {
	const op = "inventory.service.product.create"

	if warehouseID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	if quantity < 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%s: %w", op, model.ErrEmptyProductName)
	}

	product := &model.Product{
		ProductMeta: model.ProductMeta{
			WarehouseID: warehouseID,
			Name:        name,
		},
		ProductStock: model.ProductStock{
			Quantity: quantity,
			Reserved: 0,
		},
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return product, nil
}

func (s *ProductService) Get(ctx context.Context, id int64) (*model.Product, error) {
	const op = "inventory.service.product.get"

	if id <= 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}

	key := productMetaCacheKey(id)
	var meta *model.ProductMeta

	err := s.cache.Get(ctx, key, &meta)
	if err != nil {
		meta, err = s.repo.GetMeta(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		_ = s.cache.Set(ctx, key, meta, s.cacheTTL)
	}

	stock, err := s.repo.GetStock(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &model.Product{
		ProductMeta:  *meta,
		ProductStock: *stock,
	}, nil
}

func (s *ProductService) ListByWarehouse(ctx context.Context, warehouseID int64) ([]model.Product, error) {
	const op = "inventory.service.product.list_by_warehouse"

	if warehouseID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}

	products, err := s.repo.ListByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	const op = "inventory.service.product.delete"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_ = s.cache.Delete(ctx, productMetaCacheKey(id))

	return nil
}

func (s *ProductService) AdjustQuantity(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.service.product.adjust_quantity"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	if quantity == 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}

	if err := s.repo.AdjustQuantity(ctx, id, quantity); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ProductService) Reserve(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.service.product.reserve"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	if quantity <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}

	if err := s.repo.Reserve(ctx, id, quantity); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ProductService) Release(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.service.product.release"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	if quantity <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}

	if err := s.repo.Release(ctx, id, quantity); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ProductService) CancelReservation(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.service.product.cancel_reservation"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	if quantity <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}

	if err := s.repo.CancelReservation(ctx, id, quantity); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
