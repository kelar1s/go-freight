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
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
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
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, model.ErrInvalidQuantity):
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
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
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

func (h *ProductHandler) ReserveProduct(w http.ResponseWriter, r *http.Request) {
	const op = "handler.ReserveProduct"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.ReserveProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.ReserveProduct(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrNotEnoughQuantity):
			log.Warn("not enough quantity", slog.String("error", err.Error()))
			WriteError(w, http.StatusConflict, err.Error())
		case errors.Is(err, model.ErrInvalidQuantity):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to reserve product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) ReleaseProduct(w http.ResponseWriter, r *http.Request) {
	const op = "handler.ReleaseProduct"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.ReleaseProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.ReleaseProduct(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrNotEnoughQuantity):
			log.Warn("not enough quantity", slog.String("error", err.Error()))
			WriteError(w, http.StatusConflict, err.Error())
		case errors.Is(err, model.ErrInvalidQuantity):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to release product", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	const op = "handler.CancelReservation"

	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		log.Warn("failed to parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.CancelReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.CancelReservation(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrProductNotFound):
			log.Warn("product not found")
			WriteError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, model.ErrInvalidProductID):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrNotEnoughQuantity):
			log.Warn("not enough quantity", slog.String("error", err.Error()))
			WriteError(w, http.StatusConflict, err.Error())
		case errors.Is(err, model.ErrInvalidQuantity):
			log.Warn("invalid input", slog.String("error", err.Error()))
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to cancel reservation", slog.String("error", err.Error()))
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
