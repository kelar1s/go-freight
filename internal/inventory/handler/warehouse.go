package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kelar1s/go-freight/internal/inventory/handler/dto"
	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/server/middleware/logger"
)

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
