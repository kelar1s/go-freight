package rest_test

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/transport/rest"
	"github.com/kelar1s/go-freight/internal/inventory/transport/rest/mocks"
)

var mockTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

func TestProductHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(s *mocks.ProductService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			requestBody: `{"warehouse_id":1,"name":"Box","quantity":10}`,
			mockSetup: func(s *mocks.ProductService) {
				p := &model.Product{
					ProductMeta:  model.ProductMeta{ID: 1, WarehouseID: 1, Name: "Box", CreatedAt: mockTime},
					ProductStock: model.ProductStock{Quantity: 10, Reserved: 0},
				}
				s.On("Create", mock.Anything, int64(1), "Box", int64(10)).Return(p, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"warehouse_id":1,"name":"Box","quantity":10,"reserved":0,"created_at":"2026-04-11T12:00:00Z"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			requestBody:    `{"warehouse_id":1,"name":"Box,}`,
			mockSetup:      func(s *mocks.ProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
		{
			name:        "Bad Request - Validation Error",
			requestBody: `{"warehouse_id":0,"name":"Box","quantity":10}`,
			mockSetup: func(s *mocks.ProductService) {
				s.On("Create", mock.Anything, int64(0), "Box", int64(10)).Return(nil, model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Internal Server Error",
			requestBody: `{"warehouse_id":1,"name":"Box","quantity":10}`,
			mockSetup: func(s *mocks.ProductService) {
				s.On("Create", mock.Anything, int64(1), "Box", int64(10)).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewProductService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Post("/api/v1/products", h.Create)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBufferString(tc.requestBody))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestProductHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		mockSetup      func(s *mocks.ProductService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success",
			productID: "1",
			mockSetup: func(s *mocks.ProductService) {
				p := &model.Product{
					ProductMeta:  model.ProductMeta{ID: 1, WarehouseID: 1, Name: "A", CreatedAt: mockTime},
					ProductStock: model.ProductStock{Quantity: 5, Reserved: 2},
				}
				s.On("Get", mock.Anything, int64(1)).Return(p, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"warehouse_id":1,"name":"A","quantity":5,"reserved":2,"created_at":"2026-04-11T12:00:00Z"}`,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			productID:      "abc",
			mockSetup:      func(s *mocks.ProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product ID format"}`,
		},
		{
			name:      "Bad Request - Invalid ID Logic",
			productID: "0",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Get", mock.Anything, int64(0)).Return(nil, model.ErrInvalidProductID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product id"}`,
		},
		{
			name:      "Not Found",
			productID: "99",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Get", mock.Anything, int64(99)).Return(nil, model.ErrProductNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"product not found"}`,
		},
		{
			name:      "Internal Server Error",
			productID: "1",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Get", mock.Anything, int64(1)).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewProductService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Get("/api/v1/products/{id}", h.Get)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+tc.productID, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestProductHandler_ListByWarehouse(t *testing.T) {
	tests := []struct {
		name           string
		warehouseID    string
		mockSetup      func(s *mocks.ProductService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			warehouseID: "1",
			mockSetup: func(s *mocks.ProductService) {
				pList := []model.Product{
					{
						ProductMeta:  model.ProductMeta{ID: 1, WarehouseID: 1, Name: "P1", CreatedAt: mockTime},
						ProductStock: model.ProductStock{Quantity: 10, Reserved: 0},
					},
				}
				s.On("ListByWarehouse", mock.Anything, int64(1)).Return(pList, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"warehouse_id":1,"name":"P1","quantity":10,"reserved":0,"created_at":"2026-04-11T12:00:00Z"}]`,
		},
		{
			name:           "Bad Request - Invalid Param",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.ProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:        "Bad Request - Invalid ID Logic",
			warehouseID: "0",
			mockSetup: func(s *mocks.ProductService) {
				s.On("ListByWarehouse", mock.Anything, int64(0)).Return(nil, model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Internal Server Error",
			warehouseID: "1",
			mockSetup: func(s *mocks.ProductService) {
				s.On("ListByWarehouse", mock.Anything, int64(1)).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewProductService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Get("/api/v1/warehouses/{id}/products", h.ListByWarehouse)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/"+tc.warehouseID+"/products", nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestProductHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		mockSetup      func(s *mocks.ProductService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success",
			productID: "1",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Delete", mock.Anything, int64(1)).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid Param",
			productID:      "abc",
			mockSetup:      func(s *mocks.ProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Bad Request - Invalid ID Logic",
			productID: "0",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Delete", mock.Anything, int64(0)).Return(model.ErrInvalidProductID).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Not Found",
			productID: "99",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Delete", mock.Anything, int64(99)).Return(model.ErrProductNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Internal Server Error",
			productID: "1",
			mockSetup: func(s *mocks.ProductService) {
				s.On("Delete", mock.Anything, int64(1)).Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewProductService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Delete("/api/v1/products/{id}", h.Delete)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/"+tc.productID, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestProductHandler_StockOperations(t *testing.T) {
	operations := []struct {
		methodName string
		route      string
		handlerFn  func(h *rest.ProductHandler) http.HandlerFunc
	}{
		{"AdjustQuantity", "/adjust", func(h *rest.ProductHandler) http.HandlerFunc { return h.AdjustQuantity }},
		{"Reserve", "/reserve", func(h *rest.ProductHandler) http.HandlerFunc { return h.Reserve }},
		{"Release", "/release", func(h *rest.ProductHandler) http.HandlerFunc { return h.Release }},
		{"CancelReservation", "/cancel-reservation", func(h *rest.ProductHandler) http.HandlerFunc { return h.CancelReservation }},
	}

	for _, op := range operations {
		t.Run(op.methodName, func(t *testing.T) {
			tests := []struct {
				name           string
				productID      string
				requestBody    string
				mockSetup      func(s *mocks.ProductService)
				expectedStatus int
				expectedBody   string
			}{
				{
					name:        "Success",
					productID:   "1",
					requestBody: `{"quantity":5}`,
					mockSetup: func(s *mocks.ProductService) {
						s.On(op.methodName, mock.Anything, int64(1), int64(5)).Return(nil).Once()
					},
					expectedStatus: http.StatusNoContent,
				},
				{
					name:           "Bad Request - Invalid ID Format",
					productID:      "abc",
					requestBody:    `{"quantity":5}`,
					mockSetup:      func(s *mocks.ProductService) {},
					expectedStatus: http.StatusBadRequest,
					expectedBody:   `{"error":"invalid product ID format"}`,
				},
				{
					name:           "Bad Request - Invalid JSON",
					productID:      "1",
					requestBody:    `{"quantity":"lots"}`,
					mockSetup:      func(s *mocks.ProductService) {},
					expectedStatus: http.StatusBadRequest,
					expectedBody:   `{"error":"invalid request body"}`,
				},
				{
					name:        "Not Found",
					productID:   "1",
					requestBody: `{"quantity":5}`,
					mockSetup: func(s *mocks.ProductService) {
						s.On(op.methodName, mock.Anything, int64(1), int64(5)).Return(model.ErrProductNotFound).Once()
					},
					expectedStatus: http.StatusNotFound,
					expectedBody:   `{"error":"product not found"}`,
				},
				{
					name:        "Bad Request - Invalid Input Logic",
					productID:   "1",
					requestBody: `{"quantity":-5}`,
					mockSetup: func(s *mocks.ProductService) {
						s.On(op.methodName, mock.Anything, int64(1), int64(-5)).Return(model.ErrInvalidQuantity).Once()
					},
					expectedStatus: http.StatusBadRequest,
					expectedBody:   `{"error":"invalid product quantity"}`,
				},
				{
					name:        "Conflict - Not Enough Quantity",
					productID:   "1",
					requestBody: `{"quantity":100}`,
					mockSetup: func(s *mocks.ProductService) {
						s.On(op.methodName, mock.Anything, int64(1), int64(100)).Return(model.ErrNotEnoughQuantity).Once()
					},
					expectedStatus: http.StatusConflict,
					expectedBody:   `{"error":"not enough quantity"}`,
				},
				{
					name:        "Internal Server Error",
					productID:   "1",
					requestBody: `{"quantity":5}`,
					mockSetup: func(s *mocks.ProductService) {
						s.On(op.methodName, mock.Anything, int64(1), int64(5)).Return(errors.New("db error")).Once()
					},
					expectedStatus: http.StatusInternalServerError,
					expectedBody:   `{"error":"internal server error"}`,
				},
			}

			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					mockSvc := mocks.NewProductService(t)
					tc.mockSetup(mockSvc)
					h := rest.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

					r := chi.NewRouter()

					r.Patch("/api/v1/products/{id}"+op.route, op.handlerFn(h))

					req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+tc.productID+op.route, bytes.NewBufferString(tc.requestBody))
					rr := httptest.NewRecorder()
					r.ServeHTTP(rr, req)

					assert.Equal(t, tc.expectedStatus, rr.Code)
					if tc.expectedBody != "" {
						assert.JSONEq(t, tc.expectedBody, rr.Body.String())
					}
				})
			}
		})
	}
}
