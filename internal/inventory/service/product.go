package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/kelar1s/go-freight/internal/inventory/model"
)

//go:generate mockery --name=ProductRepo --output=./mocks --outpkg=mocks --with-expecter=true
type ProductRepo interface {
	Create(ctx context.Context, product *model.Product) error
	Get(ctx context.Context, id int32) (*model.Product, error)
	ListByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error)
	Delete(ctx context.Context, id int32) error
	AddQuantity(ctx context.Context, id int32, quantity int32) error
	Reserve(ctx context.Context, id int32, quantity int32) error
	Release(ctx context.Context, id int32, quantity int32) error
	CancelReservation(ctx context.Context, id int32, quantity int32) error
}

type ProductService struct {
	repo ProductRepo
}

func NewProductService(repo ProductRepo) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) Create(ctx context.Context, warehouseID int32, name string, quantity int32) (*model.Product, error) {
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
		WarehouseID: warehouseID,
		Name:        name,
		Quantity:    quantity,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return product, nil
}

func (s *ProductService) Get(ctx context.Context, id int32) (*model.Product, error) {
	const op = "inventory.service.product.get"

	if id <= 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	product, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return product, nil
}

func (s *ProductService) ListByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error) {
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

func (s *ProductService) Delete(ctx context.Context, id int32) error {
	const op = "inventory.service.product.delete"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ProductService) AddQuantity(ctx context.Context, id int32, quantity int32) error {
	const op = "inventory.service.product.add_quantity"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	// Если quantity == 0, запрос в БД делать бессмысленно
	if quantity == 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}

	if err := s.repo.AddQuantity(ctx, id, quantity); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ProductService) Reserve(ctx context.Context, id int32, quantity int32) error {
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

func (s *ProductService) Release(ctx context.Context, id int32, quantity int32) error {
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

func (s *ProductService) CancelReservation(ctx context.Context, id int32, quantity int32) error {
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
