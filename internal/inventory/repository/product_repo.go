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

type ProductRepo struct {
	db *pg.Queries
}

func NewProductRepo(db *pg.Queries) *ProductRepo {
	return &ProductRepo{
		db: db,
	}
}

func (r *ProductRepo) Create(ctx context.Context, product *model.Product) error {
	const op = "inventory.repository.product.create"
	pgProduct, err := r.db.CreateProduct(ctx, pg.CreateProductParams{
		WarehouseID: product.WarehouseID,
		Name:        product.Name,
		Quantity:    product.Quantity,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	product.CreatedAt = pgProduct.CreatedAt
	product.ID = pgProduct.ID

	return nil
}

func (r *ProductRepo) Get(ctx context.Context, id int64) (*model.Product, error) {
	const op = "inventory.repository.product.get"

	pgProduct, err := r.db.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &model.Product{
		ID:          pgProduct.ID,
		WarehouseID: pgProduct.WarehouseID,
		Name:        pgProduct.Name,
		Quantity:    pgProduct.Quantity,
		Reserved:    pgProduct.Reserved,
		CreatedAt:   pgProduct.CreatedAt,
	}, nil
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	const op = "inventory.repository.product.delete"

	_, err := r.db.DeleteProduct(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *ProductRepo) ListByWarehouse(ctx context.Context, id int64) ([]model.Product, error) {
	const op = "inventory.repository.product.list_by_warehouse"

	pgListProducts, err := r.db.ListProductsByWarehouse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	listProducts := make([]model.Product, len(pgListProducts))
	for idx, val := range pgListProducts {
		listProducts[idx] = model.Product{
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

func (r *ProductRepo) AdjustQuantity(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.adjust_quantity"

	_, err := r.db.AdjustProductQuantity(ctx, pg.AdjustProductQuantityParams{
		ID:       id,
		Quantity: quantity,
	})
	return handleStockError(op, err)
}

func (r *ProductRepo) Reserve(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.reserve"

	_, err := r.db.ReserveProduct(ctx, pg.ReserveProductParams{
		ID:       id,
		Reserved: quantity,
	})
	return handleStockError(op, err)
}

func (r *ProductRepo) Release(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.release"

	_, err := r.db.ReleaseProduct(ctx, pg.ReleaseProductParams{
		ID:       id,
		Quantity: quantity,
	})
	return handleStockError(op, err)
}

func (r *ProductRepo) CancelReservation(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.cancel_reservation"

	_, err := r.db.CancelReservation(ctx, pg.CancelReservationParams{
		ID:       id,
		Reserved: quantity,
	})
	return handleStockError(op, err)
}

func handleStockError(op string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
	}
	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23514" {
			return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}
