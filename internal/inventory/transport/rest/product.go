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
	Create(ctx context.Context, warehouseID int64, name string, quantity int64) (*model.Product, error)
	Get(ctx context.Context, id int64) (*model.Product, error)
	ListByWarehouse(ctx context.Context, warehouseID int64) ([]model.Product, error)
	Delete(ctx context.Context, id int64) error
	AdjustQuantity(ctx context.Context, id int64, quantity int64) error
	Reserve(ctx context.Context, id int64, quantity int64) error
	Release(ctx context.Context, id int64, quantity int64) error
	CancelReservation(ctx context.Context, id int64, quantity int64) error
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

// Create godoc
// @Summary      Создать новый товар
// @Description  Создает товар с указанием ID склада, названия и начального количества
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        input body dto.CreateProductRequest true "Данные товара"
// @Success      201  {object} dto.ProductResponse
// @Failure      400  {object} dto.ErrorResponse "Невалидный JSON или бизнес-ошибки (пустое имя, отрицательное количество и т.д.)"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /products [post]
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

// Get godoc
// @Summary      Получить товар по ID
// @Description  Возвращает информацию о товаре и его остатках по ID
// @Tags         products
// @Produce      json
// @Param        id   path     int  true  "ID товара"
// @Success      200  {object} dto.ProductResponse
// @Failure      400  {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      404  {object} dto.ErrorResponse "Товар не найден"
// @Failure      500  {object} dto.ErrorResponse
// @Router       /products/{id} [get]
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.get"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	product, err := h.service.Get(r.Context(), int64(productID))
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

// ListByWarehouse godoc
// @Summary      Получить список товаров на складе
// @Description  Возвращает все товары, привязанные к конкретному складу
// @Tags         products
// @Produce      json
// @Param        id   path     int  true  "ID склада"
// @Success      200  {array}  dto.ProductResponse
// @Failure      400  {object} dto.ErrorResponse "Неверный формат ID склада"
// @Failure      500  {object} dto.ErrorResponse
// @Router       /warehouses/{id}/products [get]
func (h *ProductHandler) ListByWarehouse(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.list_by_warehouse"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	warehouseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse warehouse id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid warehouse ID format")
		return
	}

	log = log.With(slog.Int("warehouse_id", int(warehouseID)))

	products, err := h.service.ListByWarehouse(r.Context(), int64(warehouseID))
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

// Delete godoc
// @Summary      Удалить товар
// @Description  Удаляет запись о товаре по его ID
// @Tags         products
// @Param        id   path int true "ID товара"
// @Success      204  "No Content"
// @Failure      400  {object} dto.ErrorResponse
// @Failure      404  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.delete"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	err = h.service.Delete(r.Context(), int64(productID))
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

// AdjustQuantity godoc
// @Summary      Скорректировать остаток товара
// @Description  Увеличивает или уменьшает остаток товара на складе на переданную дельту
// @Tags         products
// @Accept       json
// @Param        id    path int                              true "ID товара"
// @Param        input body dto.AdjustProductQuantityRequest true "Дельта изменения остатка (+/-)"
// @Success      204   "No Content"
// @Failure      400   {object} dto.ErrorResponse "Некорректный ID или тело запроса"
// @Failure      404   {object} dto.ErrorResponse "Товар не найден"
// @Failure      409   {object} dto.ErrorResponse "Конфликт: недостаточно товара на складе"
// @Failure      500   {object} dto.ErrorResponse
// @Router       /products/{id}/adjust [patch]
func (h *ProductHandler) AdjustQuantity(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.adjust_quantity"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Warn("parse product id", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid product ID format")
		return
	}

	log = log.With(slog.Int("product_id", int(productID)))

	var req dto.AdjustProductQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.AdjustQuantity(r.Context(), int64(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reserve godoc
// @Summary      Зарезервировать товар
// @Description  Резервирует указанное количество товара
// @Tags         products
// @Accept       json
// @Param        id    path int                        true "ID товара"
// @Param        input body dto.ReserveProductRequest  true "Количество для резерва"
// @Success      204   "No Content"
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      409   {object} dto.ErrorResponse "Недостаточно свободного товара для резервирования"
// @Failure      500   {object} dto.ErrorResponse
// @Router       /products/{id}/reserve [patch]
func (h *ProductHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.reserve"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
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

	err = h.service.Reserve(r.Context(), int64(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Списать зарезервированный товар
// @Description  Списывает товар со склада после успешной отгрузки
// @Tags         products
// @Accept       json
// @Param        id    path int                        true "ID товара"
// @Param        input body dto.ReleaseProductRequest  true "Количество для списания"
// @Success      204   "No Content"
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      409   {object} dto.ErrorResponse "Попытка списать больше, чем зарезервировано"
// @Failure      500   {object} dto.ErrorResponse
// @Router       /products/{id}/release [patch]
func (h *ProductHandler) Release(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.release"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
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

	err = h.service.Release(r.Context(), int64(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CancelReservation godoc
// @Summary      Отменить резерв товара
// @Description  Снимает резерв с товара
// @Tags         products
// @Accept       json
// @Param        id    path int                              true "ID товара"
// @Param        input body dto.CancelReservationRequest     true "Количество для отмены брони"
// @Success      204   "No Content"
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      409   {object} dto.ErrorResponse "Попытка отменить больше резерва, чем существует"
// @Failure      500   {object} dto.ErrorResponse
// @Router       /products/{id}/cancel-reservation [patch]
func (h *ProductHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	const op = "rest.product.cancel_reservation"
	log := logger.FromContext(r.Context(), h.log).With(slog.String("op", op))

	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
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

	err = h.service.CancelReservation(r.Context(), int64(productID), req.Quantity)
	if err != nil {
		h.handleStockError(w, log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
