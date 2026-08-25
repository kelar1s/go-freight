package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/repository/pg"
)

type ProductRepo struct {
	db *sql.DB
	q  *pg.Queries
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{
		db: db,
		q:  pg.New(db),
	}
}

func (r *ProductRepo) Create(ctx context.Context, product *model.Product) error {
	const op = "inventory.repository.product.create"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", op, err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.q.WithTx(tx)

	pgMeta, err := qtx.CreateProductMeta(ctx, pg.CreateProductMetaParams{
		WarehouseID: product.WarehouseID,
		Name:        product.Name,
	})
	if err != nil {
		return fmt.Errorf("%s: create meta: %w", op, err)
	}

	err = qtx.CreateProductStock(ctx, pg.CreateProductStockParams{
		ProductID: pgMeta.ID,
		Quantity:  product.Quantity,
	})
	if err != nil {
		return fmt.Errorf("%s: create stock: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit tx: %w", op, err)
	}

	product.ID = pgMeta.ID
	product.CreatedAt = pgMeta.CreatedAt
	product.Reserved = 0

	return nil
}

func (r *ProductRepo) GetMeta(ctx context.Context, id int64) (*model.ProductMeta, error) {
	const op = "inventory.repository.product.get_meta"

	pgMeta, err := r.q.GetProductMeta(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &model.ProductMeta{
		ID:          pgMeta.ID,
		WarehouseID: pgMeta.WarehouseID,
		Name:        pgMeta.Name,
		CreatedAt:   pgMeta.CreatedAt,
	}, nil
}

func (r *ProductRepo) GetStock(ctx context.Context, id int64) (*model.ProductStock, error) {
	const op = "inventory.repository.product.get_stock"

	pgStock, err := r.q.GetProductStock(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, model.ErrProductNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &model.ProductStock{
		Quantity: pgStock.Quantity,
		Reserved: pgStock.Reserved,
	}, nil
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	const op = "inventory.repository.product.delete"

	_, err := r.q.DeleteProduct(ctx, id)
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

	pgListProducts, err := r.q.ListProductsByWarehouse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	listProducts := make([]model.Product, len(pgListProducts))
	for idx, val := range pgListProducts {
		listProducts[idx] = model.Product{
			ProductMeta: model.ProductMeta{
				ID:          val.ID,
				WarehouseID: val.WarehouseID,
				Name:        val.Name,
				CreatedAt:   val.CreatedAt,
			},
			ProductStock: model.ProductStock{
				Quantity: val.Quantity,
				Reserved: val.Reserved,
			},
		}
	}
	return listProducts, nil
}

func (r *ProductRepo) AdjustQuantity(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.adjust_quantity"

	_, err := r.q.AdjustProductQuantity(ctx, pg.AdjustProductQuantityParams{
		ProductID: id,
		Quantity:  quantity,
	})
	return handleStockError(op, err)
}

func (r *ProductRepo) Reserve(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.reserve"

	_, err := r.q.ReserveProduct(ctx, pg.ReserveProductParams{
		ProductID: id,
		Reserved:  quantity,
	})
	return handleStockError(op, err)
}

func (r *ProductRepo) Release(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.release"

	_, err := r.q.ReleaseProduct(ctx, pg.ReleaseProductParams{
		ProductID: id,
		Quantity:  quantity,
	})
	return handleStockError(op, err)
}

func (r *ProductRepo) CancelReservation(ctx context.Context, id int64, quantity int64) error {
	const op = "inventory.repository.product.cancel_reservation"

	_, err := r.q.CancelReservation(ctx, pg.CancelReservationParams{
		ProductID: id,
		Reserved:  quantity,
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
	if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgxErr.Code == "23514" {
			return fmt.Errorf("%s: %w", op, model.ErrNotEnoughQuantity)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}
