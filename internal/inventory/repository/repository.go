package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/repository/pg"
	"github.com/lib/pq"
)

type ProductRepository struct {
	db *pg.Queries
}

func NewProductRepository(db *pg.Queries) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (pr *ProductRepository) CreateWarehouse(ctx context.Context, name string, location string) (model.Warehouse, error) {
	const op = "repository.postgres.CreateWarehouse"

	pgWarehouse, err := pr.db.CreateWarehouse(ctx, pg.CreateWarehouseParams{Name: name, Location: location})
	if err != nil {
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, err)
	}
	return model.Warehouse{
		ID:        pgWarehouse.ID,
		Name:      pgWarehouse.Name,
		Location:  pgWarehouse.Location,
		CreatedAt: pgWarehouse.CreatedAt,
	}, nil
}

func (pr *ProductRepository) DeleteWarehouse(ctx context.Context, id int32) error {
	const op = "repository.postgres.DeleteWarehouse"

	_, err := pr.db.DeleteWarehouse(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrWarehouseNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) GetWarehouse(ctx context.Context, id int32) (model.Warehouse, error) {
	const op = "repository.postgres.GetWarehouse"

	pgWarehouse, err := pr.db.GetWarehouse(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Warehouse{}, fmt.Errorf("%s: %w", op, model.ErrWarehouseNotFound)
		}
		return model.Warehouse{}, fmt.Errorf("%s: %w", op, err)
	}
	return model.Warehouse{
		ID:        pgWarehouse.ID,
		Name:      pgWarehouse.Name,
		Location:  pgWarehouse.Location,
		CreatedAt: pgWarehouse.CreatedAt,
	}, nil
}

func (pr *ProductRepository) ListWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	const op = "repository.postgres.ListWarehouses"

	pgListWarehouses, err := pr.db.ListWarehouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	listWarehouses := make([]model.Warehouse, len(pgListWarehouses))
	for id, val := range pgListWarehouses {
		listWarehouses[id] = model.Warehouse{
			ID:        val.ID,
			Name:      val.Name,
			Location:  val.Location,
			CreatedAt: val.CreatedAt,
		}
	}
	return listWarehouses, nil
}

func (pr *ProductRepository) UpdateWarehouse(ctx context.Context, id int32, name string, location string) error {
	const op = "repository.postgres.UpdateWarehouse"

	_, err := pr.db.UpdateWarehouse(ctx, pg.UpdateWarehouseParams{
		ID:       id,
		Name:     name,
		Location: location,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrWarehouseNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) CreateProduct(ctx context.Context, warehouseID int32, name string, quantity int32) (model.Product, error) {
	const op = "repository.postgres.CreateProduct"

	pgProduct, err := pr.db.CreateProduct(ctx, pg.CreateProductParams{
		WarehouseID: warehouseID,
		Name:        name,
		Quantity:    quantity,
	})
	if err != nil {
		return model.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	return model.Product{
		ID:          pgProduct.ID,
		WarehouseID: pgProduct.WarehouseID,
		Name:        pgProduct.Name,
		Quantity:    pgProduct.Quantity,
		Reserved:    pgProduct.Reserved,
		CreatedAt:   pgProduct.CreatedAt,
	}, nil
}

func (pr *ProductRepository) DeleteProduct(ctx context.Context, id int32) error {
	const op = "repository.postgres.DeleteProduct"

	_, err := pr.db.DeleteProduct(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) GetProduct(ctx context.Context, id int32) (model.Product, error) {
	const op = "repository.postgres.GetProduct"

	pgProduct, err := pr.db.GetProduct(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return model.Product{}, fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		default:
			return model.Product{}, fmt.Errorf("%s: %w", op, err)
		}
	}
	return model.Product{
		ID:          pgProduct.ID,
		WarehouseID: pgProduct.WarehouseID,
		Name:        pgProduct.Name,
		Quantity:    pgProduct.Quantity,
		Reserved:    pgProduct.Reserved,
		CreatedAt:   pgProduct.CreatedAt,
	}, nil
}

func (pr *ProductRepository) ListProductsByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error) {
	const op = "repository.postgres.ListProductsByWarehouse"

	pgListProducts, err := pr.db.ListProductsByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	listProducts := make([]model.Product, len(pgListProducts))
	for ind, val := range pgListProducts {
		listProducts[ind] = model.Product{
			ID:          val.ID,
			WarehouseID: val.WarehouseID,
			Name:        val.Name,
			Quantity:    val.Quantity,
			Reserved:    val.Reserved,
			CreatedAt:   val.CreatedAt,
		}
	}
	return listProducts, nil
}

func (pr *ProductRepository) SetProductQuantity(ctx context.Context, id int32, quantity int32) error {
	const op = "repository.postgres.SetProductQuantity"

	_, err := pr.db.SetProductQuantity(ctx, pg.SetProductQuantityParams{
		ID:       id,
		Quantity: quantity,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) AddProductQuantity(ctx context.Context, id int32, quantity int32) error {
	const op = "repository.postgres.AddProductQuantity"

	_, err := pr.db.AddProductQuantity(ctx, pg.AddProductQuantityParams{
		ID:       id,
		Quantity: quantity,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23514" {
				return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) ReserveProduct(ctx context.Context, id int32, quantity int32) error {
	const op = "repository.postgres.ReserveProduct"

	_, err := pr.db.ReserveProduct(ctx, pg.ReserveProductParams{
		ID:       id,
		Reserved: quantity,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23514" {
				return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) ReleaseProduct(ctx context.Context, id int32, quantity int32) error {
	const op = "repository.postgres.ReleaseProduct"

	_, err := pr.db.ReleaseProduct(ctx, pg.ReleaseProductParams{
		ID:       id,
		Quantity: quantity,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23514" {
				return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pr *ProductRepository) CancelReservation(ctx context.Context, id int32, quantity int32) error {
	const op = "repository.postgres.CancelReservation"

	_, err := pr.db.CancelReservation(ctx, pg.CancelReservationParams{
		ID:       id,
		Reserved: quantity,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23514" {
				return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
