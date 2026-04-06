package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kelar1s/go-freight/internal/products/handler/dto"
	"github.com/kelar1s/go-freight/internal/products/model"
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
	service Service
}

func NewProductHandler(service Service) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	warehouse, err := h.service.CreateWarehouse(r.Context(), req.Name, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmptyWarehouseName):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrEmptyWarehouseLocation):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToWarehouseResponse(warehouse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error: failed to encode response: %v", err) // todo: logger
		return
	}
}

func (h *ProductHandler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}
	err = h.service.DeleteWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}
	warehouse, err := h.service.GetWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrWarehouseNotFound):
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToWarehouseResponse(warehouse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error: failed to encode response: %v", err) // todo: logger
		return
	}
}

func (h *ProductHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	warehouses, err := h.service.ListWarehouses(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	res := dto.ToWarehouseResponseList(warehouses)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error: failed to encode response: %v", err) // todo: logger
		return
	}
}

func (h *ProductHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}
	var req dto.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.UpdateWarehouse(r.Context(), int32(warehouseID), req.Name, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmptyWarehouseName):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrEmptyWarehouseLocation):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	product, err := h.service.CreateProduct(r.Context(), req.WarehouseID, req.Name, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrInvalidQuantity):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrEmptyProductName):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToProductResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error: failed to encode response: %v", err) // todo: logger
		return
	}
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}
	err = h.service.DeleteProduct(r.Context(), int32(productID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}
	product, err := h.service.GetProduct(r.Context(), int32(productID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrProductNotFound):
			WriteError(w, http.StatusNotFound, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToProductResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error: failed to encode response: %v", err) // todo: logger
		return
	}
}

func (h *ProductHandler) ListProductsByWarehouse(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}
	products, err := h.service.ListProductsByWarehouse(r.Context(), int32(warehouseID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidWarehouseID):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	res := dto.ToProductResponseList(products)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error: failed to encode response: %v", err) // todo: logger
		return
	}
}

func (h *ProductHandler) SetProductQuantity(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}
	var req dto.SetProductQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.SetProductQuantity(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) AddProductQuantity(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}
	var req dto.SetProductQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = h.service.AddProductQuantity(r.Context(), int32(productID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidProductID):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrNotEnoughQuantity):
			WriteError(w, http.StatusConflict, err.Error())
		default:
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
