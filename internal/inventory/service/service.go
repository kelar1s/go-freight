package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/kelar1s/go-freight/internal/inventory/model"
)

type Repository interface {
	CreateWarehouse(ctx context.Context, name string, location string) (model.Warehouse, error)
	DeleteWarehouse(ctx context.Context, id int32) error
	GetWarehouse(ctx context.Context, id int32) (model.Warehouse, error)
	ListWarehouses(ctx context.Context) ([]model.Warehouse, error)
	UpdateWarehouse(ctx context.Context, id int32, name string, location string) error

	CreateProduct(ctx context.Context, warehouseID int32, name string, quantity int32) (model.Product, error)
	DeleteProduct(ctx context.Context, id int32) error
	GetProduct(ctx context.Context, id int32) (model.Product, error)
	ListProductsByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error)
	SetProductQuantity(ctx context.Context, id int32, quantity int32) error
	AddProductQuantity(ctx context.Context, id int32, quantity int32) error
}

type ProductService struct {
	repo Repository
}

func NewProductService(repo Repository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (ps *ProductService) CreateWarehouse(ctx context.Context, name string, location string) (model.Warehouse, error) {
	const op = "service.ProductService.CreateWarehouse"

	name = strings.TrimSpace(name)
	location = strings.TrimSpace(location)
	if name == "" {
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, model.ErrEmptyWarehouseName)
	}
	if location == "" {
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, model.ErrEmptyWarehouseLocation)
	}
	warehouse, err := ps.repo.CreateWarehouse(ctx, name, location)
	if err != nil {
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, err)
	}
	return warehouse, nil
}

func (ps *ProductService) DeleteWarehouse(ctx context.Context, id int32) error {
	const op = "service.ProductService.DeleteWarehouse"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	err := ps.repo.DeleteWarehouse(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ps *ProductService) GetWarehouse(ctx context.Context, id int32) (model.Warehouse, error) {
	const op = "service.ProductService.GetWarehouse"

	if id <= 0 {
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	warehouse, err := ps.repo.GetWarehouse(ctx, id)
	if err != nil {
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, err)
	}
	return warehouse, nil
}

func (ps *ProductService) ListWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	const op = "service.ProductService.ListWarehouses"

	warehouses, err := ps.repo.ListWarehouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return warehouses, nil
}

func (ps *ProductService) UpdateWarehouse(ctx context.Context, id int32, name string, location string) error {
	const op = "service.ProductService.UpdateWarehouse"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	name = strings.TrimSpace(name)
	location = strings.TrimSpace(location)
	if name == "" {
		return fmt.Errorf("%s: %w", op, model.ErrEmptyWarehouseName)
	}
	if location == "" {
		return fmt.Errorf("%s: %w", op, model.ErrEmptyWarehouseLocation)
	}
	err := ps.repo.UpdateWarehouse(ctx, id, name, location)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ps *ProductService) CreateProduct(ctx context.Context, warehouseID int32, name string, quantity int32) (model.Product, error) {
	const op = "service.ProductService.CreateProduct"

	if warehouseID <= 0 {
		return model.Product{}, fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	if quantity < 0 {
		return model.Product{}, fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Product{}, fmt.Errorf("%s: %w", op, model.ErrEmptyProductName)
	}
	product, err := ps.repo.CreateProduct(ctx, warehouseID, name, quantity)
	if err != nil {
		return model.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	return product, nil
}

func (ps *ProductService) DeleteProduct(ctx context.Context, id int32) error {
	const op = "service.ProductService.DeleteProduct"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	err := ps.repo.DeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ps *ProductService) GetProduct(ctx context.Context, id int32) (model.Product, error) {
	const op = "service.ProductService.GetProduct"

	if id <= 0 {
		return model.Product{}, fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	product, err := ps.repo.GetProduct(ctx, id)
	if err != nil {
		return model.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	return product, nil
}

func (ps *ProductService) ListProductsByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error) {
	const op = "service.ProductService.ListProductsByWarehouse"

	if warehouseID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	products, err := ps.repo.ListProductsByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return products, nil
}

func (ps *ProductService) SetProductQuantity(ctx context.Context, id int32, quantity int32) error {
	const op = "service.ProductService.SetProductQuantity"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	if quantity < 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidQuantity)
	}
	err := ps.repo.SetProductQuantity(ctx, id, quantity)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ps *ProductService) AddProductQuantity(ctx context.Context, id int32, quantity int32) error {
	const op = "service.ProductService.AddProductQuantity"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidProductID)
	}
	err := ps.repo.AddProductQuantity(ctx, id, quantity)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
