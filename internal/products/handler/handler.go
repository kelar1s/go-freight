package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kelar1s/go-freight/internal/products/handler/dto"
	"github.com/kelar1s/go-freight/internal/products/model"
	"github.com/kelar1s/go-freight/internal/server/middleware/logger"
)

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

func (h *ProductHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "handler.CreateWarehouse"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	var req dto.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	warehouse, err := h.service.CreateWarehouse(r.Context(), req.Name, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmptyWarehouseName):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrEmptyWarehouseLocation):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to create warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToWarehouseResponse(warehouse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "handler.DeleteWarehouse"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	err = h.service.DeleteWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to delete warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "handler.GetWarehouse"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	warehouse, err := h.service.GetWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrWarehouseNotFound):
			log.Warn("warehouse not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("failed to get warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToWarehouseResponse(warehouse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	const op = "handler.ListWarehouses"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouses, err := h.service.ListWarehouses(r.Context())
	if err != nil {
		log.Error("failed to get warehouses", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	res := dto.ToWarehouseResponseList(warehouses)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "handler.UpdateWarehouse"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	var req dto.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.UpdateWarehouse(r.Context(), int32(warehouseID), req.Name, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmptyWarehouseName):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrEmptyWarehouseLocation):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to update warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	const op = "handler.CreateProduct"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	product, err := h.service.CreateProduct(r.Context(), req.WarehouseID, req.Name, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrInvalidQuantity):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrEmptyProductName):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to create product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToProductResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	const op = "handler.DeleteProduct"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	err = h.service.DeleteProduct(r.Context(), int32(productID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to delete product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	const op = "handler.GetProduct"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	product, err := h.service.GetProduct(r.Context(), int32(productID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("failed to get product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToProductResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) ListProductsByWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "handler.ListProductsByWarehouse"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	products, err := h.service.ListProductsByWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to get products by warehouse id", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToProductResponseList(products)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("failed to encode response", slog.String("error", err.Error()))
		return
	}
}

func (h *ProductHandler) SetProductQuantity(w http.ResponseWriter, r *http.Request) {
	const op = "handler.SetProductQuantity"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.SetProductQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.SetProductQuantity(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to set product quantity", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) AddProductQuantity(w http.ResponseWriter, r *http.Request) {
	const op = "handler.AddProductQuantity"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.SetProductQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.AddProductQuantity(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrNotEnoughQuantity):
			log.Warn("not enough quantity", slog.String("error", err.Error()))
			WriteError(w, http.StatusConflict, err.Error())
		default:
			log.Error("failed to add product quantity", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := dto.ErrorResponse{
		Error: message,
	}
	json.NewEncoder(w).Encode(err)
}
