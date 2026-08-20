package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/transport/rest/dto"
	"github.com/kelar1s/go-freight/internal/pkg/middleware/logger"
)

//go:generate mockery --name=ProductService --output=./mocks --outpkg=mocks --with-expecter=true
type ProductService interface {
	Create(ctx context.Context, warehouseID int32, name string, quantity int32) (*model.Product, error)
	Get(ctx context.Context, id int32) (*model.Product, error)
	ListByWarehouse(ctx context.Context, warehouseID int32) ([]model.Product, error)
	Delete(ctx context.Context, id int32) error
	AddQuantity(ctx context.Context, id int32, quantity int32) error
	Reserve(ctx context.Context, id int32, quantity int32) error
	Release(ctx context.Context, id int32, quantity int32) error
	CancelReservation(ctx context.Context, id int32, quantity int32) error
}

type ProductHandler struct {
	service ProductService
	log     *slog.Logger
}

func NewProductHandler(service ProductService, log *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		log:     log,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.create"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.service.Create(r.Context(), req.WarehouseID, req.Name, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID), errors.Is(err, model.ErrInvalidQuantity), errors.Is(err, model.ErrEmptyProductName):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("create product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	res := dto.ToProductResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.get"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	product, err := h.service.Get(r.Context(), int32(productID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("get product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	res := dto.ToProductResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) ListByWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.list_by_warehouse"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	products, err := h.service.ListByWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("list products by warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	res := dto.ToProductResponseList(products)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("encode response", slog.String("error", err.Error()))
	}
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.delete"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	err = h.service.Delete(r.Context(), int32(productID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("delete product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) AddQuantity(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.add_quantity"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.SetProductQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.AddQuantity(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.reserve"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.ReserveProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.Reserve(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) Release(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.release"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.ReleaseProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.Release(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.cancel_reservation"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.CancelReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.CancelReservation(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleStockError выносит общую логику обработки ошибок для операций со складом, чтобы не дублировать код в каждом методе
func (h *ProductHandler) handleStockError(w http.ResponseWriter, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, model.ErrProductNotFound):
		log.Warn("product not found")
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, model.ErrInvalidProductID), errors.Is(err, model.ErrInvalidQuantity):
		log.Warn("invalid input", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrNotEnoughQuantity):
		log.Warn("not enough quantity", slog.String("error", err.Error()))
		WriteError(w, http.StatusConflict, err.Error())
	default:
		log.Error("stock operation failed", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
