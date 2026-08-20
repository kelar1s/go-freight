package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/repository/pg"
)

type WarehouseRepo struct {
	db *pg.Queries
}

func NewWarehouseRepo(db *pg.Queries) *WarehouseRepo {
	return &WarehouseRepo{
		db: db,
	}
}

func (r *WarehouseRepo) Create(ctx context.Context, warehouse *model.Warehouse) error {
	const op = "inventory.repository.warehouse.create"

	pgWarehouse, err := r.db.CreateWarehouse(ctx, pg.CreateWarehouseParams{
		Name:     warehouse.Name,
		Location: warehouse.Location,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	warehouse.CreatedAt = pgWarehouse.CreatedAt
	warehouse.ID = pgWarehouse.ID
	return nil
}

func (r *WarehouseRepo) Get(ctx context.Context, id int32) (*model.Warehouse, error) {
	const op = "inventory.repository.warehouse.get"

	pgWarehouse, err := r.db.GetWarehouse(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, model.ErrWarehouseNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &model.Warehouse{
		ID:        pgWarehouse.ID,
		Name:      pgWarehouse.Name,
		Location:  pgWarehouse.Location,
		CreatedAt: pgWarehouse.CreatedAt,
	}, nil
}

func (r *WarehouseRepo) List(ctx context.Context) ([]model.Warehouse, error) {
	const op = "inventory.repository.warehouse.list"

	pgListWarehouses, err := r.db.ListWarehouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	listWarehouses := make([]model.Warehouse, len(pgListWarehouses))
	for idx, val := range pgListWarehouses {
		listWarehouses[idx] = model.Warehouse{
			ID:        val.ID,
			Name:      val.Name,
			Location:  val.Location,
			CreatedAt: val.CreatedAt,
		}
	}
	return listWarehouses, nil
}

func (r *WarehouseRepo) Update(ctx context.Context, warehouse *model.Warehouse) error {
	const op = "inventory.repository.warehouse.update"

	_, err := r.db.UpdateWarehouse(ctx, pg.UpdateWarehouseParams{
		ID:       warehouse.ID,
		Name:     warehouse.Name,
		Location: warehouse.Location,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrWarehouseNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *WarehouseRepo) Delete(ctx context.Context, id int32) error {
	const op = "inventory.repository.warehouse.delete"

	_, err := r.db.DeleteWarehouse(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrWarehouseNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
