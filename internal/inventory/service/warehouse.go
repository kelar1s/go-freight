package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/kelar1s/go-freight/internal/inventory/model"
)

//go:generate mockery --name=WarehouseRepo --output=./mocks --outpkg=mocks --with-expecter=true
type WarehouseRepo interface {
	Create(ctx context.Context, warehouse *model.Warehouse) error
	Get(ctx context.Context, id int64) (*model.Warehouse, error)
	List(ctx context.Context) ([]model.Warehouse, error)
	Update(ctx context.Context, warehouse *model.Warehouse) error
	Delete(ctx context.Context, id int64) error
}

type WarehouseService struct {
	repo WarehouseRepo
}

func NewWarehouseService(repo WarehouseRepo) *WarehouseService {
	return &WarehouseService{
		repo: repo,
	}
}

func (s *WarehouseService) Create(ctx context.Context, name string, location string) (*model.Warehouse, error) {
	const op = "inventory.service.warehouse.create"

	name = strings.TrimSpace(name)
	location = strings.TrimSpace(location)
	if name == "" {
		return nil, fmt.Errorf("%s: %w", op, model.ErrEmptyWarehouseName)
	}
	if location == "" {
		return nil, fmt.Errorf("%s: %w", op, model.ErrEmptyWarehouseLocation)
	}
	warehouse := &model.Warehouse{
		Name:     name,
		Location: location,
	}
	if err := s.repo.Create(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return warehouse, nil
}

func (s *WarehouseService) Get(ctx context.Context, id int64) (*model.Warehouse, error) {
	const op = "inventory.service.warehouse.get"

	if id <= 0 {
		return nil, fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	warehouse, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return warehouse, nil
}

func (s *WarehouseService) List(ctx context.Context) ([]model.Warehouse, error) {
	const op = "inventory.service.warehouse.list"

	warehouses, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return warehouses, nil
}

func (s *WarehouseService) Update(ctx context.Context, id int64, name, location string) error {
	const op = "inventory.service.warehouse.update"

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
	warehouse := &model.Warehouse{
		ID:       id,
		Name:     name,
		Location: location,
	}
	err := s.repo.Update(ctx, warehouse)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *WarehouseService) Delete(ctx context.Context, id int64) error {
	const op = "inventory.service.warehouse.delete"

	if id <= 0 {
		return fmt.Errorf("%s: %w", op, model.ErrInvalidWarehouseID)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
