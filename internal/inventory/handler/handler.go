package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kelar1s/go-freight/internal/inventory/handler/dto"
	"github.com/kelar1s/go-freight/internal/inventory/model"
)

//go:generate mockery --name=Service --output=./mocks --outpkg=mocks --with-expecter=true
type Service interface {
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

type ProductHandler struct {
	log     *slog.Logger
	service Service
}

func NewProductHandler(service Service, log *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		log:     log,
	}
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := dto.ErrorResponse{
		Error: message,
	}
	_ = json.NewEncoder(w).Encode(err)
}
