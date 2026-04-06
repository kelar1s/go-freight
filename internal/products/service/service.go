package service

import (
	"context"
	"strings"

	"github.com/kelar1s/go-freight/internal/products/model"
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
	name = strings.TrimSpace(name)
	location = strings.TrimSpace(location)
	if name == "" {
		return model.Warehouse{}, model.ErrEmptyWarehouseName
	}
	if location == "" {
		return model.Warehouse{}, model.ErrEmptyWarehouseLocation
	}
	warehouse, err := ps.repo.CreateWarehouse(ctx, name, location)
	if err != nil {
		return model.Warehouse{}, err
	}
	return warehouse, nil
}

func (ps *ProductService) DeleteWarehouse(ctx context.Context, id int32) error {
	if id <= 0 {
		return model.ErrInvalidWarehouseID
	}
	err := ps.repo.DeleteWarehouse(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (ps *ProductService) GetWarehouse(ctx context.Context, id int32) (model.Warehouse, error) {
	if id <= 0 {
		return model.Warehouse{}, model.ErrInvalidWarehouseID
	}
	warehouse, err := ps.repo.GetWarehouse(ctx, id)
	if err != nil {
		return model.Warehouse{}, err
	}
	return warehouse, nil
}

func (ps *ProductService) ListWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	warehouses, err := ps.repo.ListWarehouses(ctx)
	if err != nil {
		return nil, err
	}
	return warehouses, nil
}

func (ps *ProductService) UpdateWarehouse(ctx context.Context, id int32, name string, location string) error {
	if id <= 0 {
		return model.ErrInvalidWarehouseID
	}
	name = strings.TrimSpace(name)
	location = strings.TrimSpace(location)
	if name == "" {
		return model.ErrEmptyWarehouseName
	}
	if location == "" {
		return model.ErrEmptyWarehouseLocation
	}
	err := ps.repo.UpdateWarehouse(ctx, id, name, location)
	if err != nil {
		return err
	}
	return nil
}

func (ps *ProductService) CreateProduct(ctx context.Context, warehouseID int32, name string, quantity int32) (model.Product, error) {
	if warehouseID <= 0 {
		return model.Product{}, model.ErrInvalidWarehouseID
	}
	if quantity < 0 {
		return model.Product{}, model.ErrInvalidQuantity
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Product{}, model.ErrEmptyProductName
	}
	product, err := ps.repo.CreateProduct(ctx, warehouseID, name, quantity)
	if err != nil {
		return model.Product{}, err
	}
	return product, nil
}

func (ps *ProductService) DeleteProduct(ctx context.Context, id int32) error {
	if id <= 0 {
		return model.ErrInvalidProductID
	}
	err := ps.repo.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (ps *ProductService) GetProduct(ctx context.Context, id int32) (model.Product, error) {
	if id <= 0 {
		return model.Product{}, model.ErrInvalidProductID
	}
	product, err := ps.repo.GetProduct(ctx, id)
	if err != nil {
		return model.Product{}, err
	}
	return product, nil
}

func (ps *ProductService) ListProductsByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error) {
	if warehouseID <= 0 {
		return nil, model.ErrInvalidWarehouseID
	}
	products, err := ps.repo.ListProductsByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (ps *ProductService) SetProductQuantity(ctx context.Context, id int32, quantity int32) error {
	if id <= 0 {
		return model.ErrInvalidProductID
	}
	if quantity < 0 {
		return model.ErrInvalidQuantity
	}
	err := ps.repo.SetProductQuantity(ctx, id, quantity)
	if err != nil {
		return err
	}
	return nil
}

func (ps *ProductService) AddProductQuantity(ctx context.Context, id int32, quantity int32) error {
	if id <= 0 {
		return model.ErrInvalidProductID
	}
	err := ps.repo.AddProductQuantity(ctx, id, quantity)
	if err != nil {
		return err
	}
	return nil
}
