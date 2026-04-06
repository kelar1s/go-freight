package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kelar1s/go-freight/internal/products/model"
	"github.com/kelar1s/go-freight/internal/products/repository/pg"
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
	pgWarehouse, err := pr.db.CreateWarehouse(ctx, pg.CreateWarehouseParams{Name: name, Location: location})
	if err != nil {
		return model.Warehouse{}, err
	}
	return model.Warehouse{
		ID:        pgWarehouse.ID,
		Name:      pgWarehouse.Name,
		Location:  pgWarehouse.Location,
		CreatedAt: pgWarehouse.CreatedAt,
	}, nil
}

func (pr *ProductRepository) DeleteWarehouse(ctx context.Context, id int32) error {
	return pr.db.DeleteWarehouse(ctx, id)
}

func (pr *ProductRepository) GetWarehouse(ctx context.Context, id int32) (model.Warehouse, error) {
	pgWarehouse, err := pr.db.GetWarehouse(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return model.Warehouse{}, model.ErrWarehouseNotFound
		default:
			return model.Warehouse{}, err
		}
	}
	return model.Warehouse{
		ID:        pgWarehouse.ID,
		Name:      pgWarehouse.Name,
		Location:  pgWarehouse.Location,
		CreatedAt: pgWarehouse.CreatedAt,
	}, nil
}

func (pr *ProductRepository) ListWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	pgListWarehouses, err := pr.db.ListWarehouses(ctx)
	if err != nil {
		return nil, err
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
	return pr.db.UpdateWarehouse(ctx, pg.UpdateWarehouseParams{
		ID:       id,
		Name:     name,
		Location: location,
	})
}

func (pr *ProductRepository) CreateProduct(ctx context.Context, warehouseID int32, name string, quantity int32) (model.Product, error) {
	pgProduct, err := pr.db.CreateProduct(ctx, pg.CreateProductParams{
		WarehouseID: warehouseID,
		Name:        name,
		Quantity:    quantity,
	})
	if err != nil {
		return model.Product{}, err
	}
	return model.Product{
		ID:          pgProduct.ID,
		WarehouseID: pgProduct.WarehouseID,
		Name:        pgProduct.Name,
		Quantity:    pgProduct.Quantity,
		CreatedAt:   pgProduct.CreatedAt,
	}, nil
}

func (pr *ProductRepository) DeleteProduct(ctx context.Context, id int32) error {
	return pr.db.DeleteProduct(ctx, id)
}

func (pr *ProductRepository) GetProduct(ctx context.Context, id int32) (model.Product, error) {
	pgProduct, err := pr.db.GetProduct(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return model.Product{}, model.ErrProductNotFound
		default:
			return model.Product{}, err
		}
	}
	return model.Product{
		ID:          pgProduct.ID,
		WarehouseID: pgProduct.WarehouseID,
		Name:        pgProduct.Name,
		Quantity:    pgProduct.Quantity,
		CreatedAt:   pgProduct.CreatedAt,
	}, nil
}

func (pr *ProductRepository) ListProductsByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error) {
	pgListProducts, err := pr.db.ListProductsByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	listProducts := make([]model.Product, len(pgListProducts))
	for ind, val := range pgListProducts {
		listProducts[ind] = model.Product{
			ID:          val.ID,
			WarehouseID: val.WarehouseID,
			Name:        val.Name,
			Quantity:    val.Quantity,
			CreatedAt:   val.CreatedAt,
		}
	}
	return listProducts, nil
}

func (pr *ProductRepository) SetProductQuantity(ctx context.Context, id int32, quantity int32) error {
	return pr.db.SetProductQuantity(ctx, pg.SetProductQuantityParams{
		ID:       id,
		Quantity: quantity,
	})

}

func (pr *ProductRepository) AddProductQuantity(ctx context.Context, id int32, quantity int32) error {
	err := pr.db.AddProductQuantity(ctx, pg.AddProductQuantityParams{
		ID:       id,
		Quantity: quantity,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23514" {
				return model.ErrNotEnoughQuantity
			}
		}
		return err
	}
	return nil
}
