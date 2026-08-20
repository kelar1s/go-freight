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

//go:generate mockery --name=WarehouseService --output=./mocks --outpkg=mocks --with-expecter=true
type WarehouseService interface {
	Create(ctx context.Context, name string, location string) (*model.Warehouse, error)
	Get(ctx context.Context, id int64) (*model.Warehouse, error)
	List(ctx context.Context) ([]model.Warehouse, error)
	Update(ctx context.Context, id int64, name string, location string) error
	Delete(ctx context.Context, id int64) error
}

type WarehouseHandler struct {
	service WarehouseService
	log     *slog.Logger
}

func NewWarehouseHandler(service WarehouseService, log *slog.Logger) *WarehouseHandler {
	return &WarehouseHandler{
		service: service,
		log:     log,
	}
}

func (h *WarehouseHandler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "rest.warehouse.create"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	var req dto.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	warehouse, err := h.service.Create(r.Context(), req.Name, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmptyWarehouseName), errors.Is(err, model.ErrEmptyWarehouseLocation):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("create warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	res := dto.ToWarehouseResponse(warehouse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("encode response", slog.String("error", err.Error()))
	}
}

func (h *WarehouseHandler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "rest.warehouse.get"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	warehouse, err := h.service.Get(r.Context(), int64(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrWarehouseNotFound):
			log.Warn("warehouse not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("get warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	res := dto.ToWarehouseResponse(warehouse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("encode response", slog.String("error", err.Error()))
	}
}

func (h *WarehouseHandler) List(w http.ResponseWriter, r *http.Request) {
	const op = "rest.warehouse.list"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouses, err := h.service.List(r.Context())
	if err != nil {
		log.Error("list warehouses", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	res := dto.ToWarehouseResponseList(warehouses)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("encode response", slog.String("error", err.Error()))
	}
}

func (h *WarehouseHandler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "rest.warehouse.update"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	var req dto.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.Update(r.Context(), int64(warehouseID), req.Name, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmptyWarehouseName), errors.Is(err, model.ErrEmptyWarehouseLocation):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrWarehouseNotFound):
			log.Warn("warehouse not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("update warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WarehouseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "rest.warehouse.delete"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	err = h.service.Delete(r.Context(), int64(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrWarehouseNotFound):
			log.Warn("warehouse not found")
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			log.Error("delete warehouse", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
