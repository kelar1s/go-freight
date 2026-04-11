package handler_test

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

	"github.com/kelar1s/go-freight/internal/inventory/handler"
	"github.com/kelar1s/go-freight/internal/inventory/handler/mocks"
	"github.com/kelar1s/go-freight/internal/inventory/model"
)

func TestHandler_CreateProduct(t *testing.T) {
	type testCase struct {
		name           string
		requestBody    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:        "Success",
			requestBody: `{"warehouse_id":1,"name":"Box","quantity":10}`,
			mockSetup: func(s *mocks.Service) {
				p := model.Product{ID: 1, WarehouseID: 1, Name: "Box", Quantity: 10, CreatedAt: time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)}
				s.On("CreateProduct", mock.Anything, int32(1), "Box", int32(10)).Return(p, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"warehouse_id":1,"name":"Box","quantity":10,"created_at":"2026-04-11T12:00:00Z"}`,
		},
		{
			name:        "Bad Request - Invalid JSON",
			requestBody: `{"warehouse_id":1,"name":"Box,"quantity":10}`,
			mockSetup: func(s *mocks.Service) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
		{
			name:        "Bad Request - Invalid Warehouse ID",
			requestBody: `{"warehouse_id":1,"name":"Box","quantity":10}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateProduct", mock.Anything, int32(1), "Box", int32(10)).Return(model.Product{}, model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Bad Request - Empty Name",
			requestBody: `{"warehouse_id":1,"name":"","quantity":10}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateProduct", mock.Anything, int32(1), "", int32(10)).Return(model.Product{}, model.ErrEmptyProductName).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"product name cannot be empty"}`,
		},
		{
			name:        "Bad Request - Invalid Quantity",
			requestBody: `{"warehouse_id":1,"name":"Box","quantity":-5}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateProduct", mock.Anything, int32(1), "Box", int32(-5)).Return(model.Product{}, model.ErrInvalidQuantity).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product quantity"}`,
		},
		{
			name:        "Internal Server Error - DB",
			requestBody: `{"warehouse_id":1,"name":"Box","quantity":10}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateProduct", mock.Anything, int32(1), "Box", int32(10)).Return(model.Product{}, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Post("/api/v1/products", h.CreateProduct)

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

func TestHandler_DeleteProduct(t *testing.T) {
	type testCase struct {
		name           string
		productID      string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:      "Success",
			productID: "1",
			mockSetup: func(s *mocks.Service) {
				s.On("DeleteProduct", mock.Anything, int32(1)).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			productID:      "abc",
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product ID format"}`,
		},
		{
			name:      "Bad Request - Invalid ID",
			productID: "99",
			mockSetup: func(s *mocks.Service) {
				s.On("DeleteProduct", mock.Anything, int32(99)).Return(model.ErrInvalidProductID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product id"}`,
		},
		{
			name:      "Internal Server Error - DB",
			productID: "1",
			mockSetup: func(s *mocks.Service) {
				s.On("DeleteProduct", mock.Anything, int32(1)).Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Delete("/api/v1/products/{id}", h.DeleteProduct)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/"+tc.productID, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_GetProduct(t *testing.T) {
	type testCase struct {
		name           string
		productID      string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:      "Success",
			productID: "1",
			mockSetup: func(s *mocks.Service) {
				p := model.Product{ID: 1, Name: "A", Quantity: 5, CreatedAt: time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)}
				s.On("GetProduct", mock.Anything, int32(1)).Return(p, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"warehouse_id":0,"name":"A","quantity":5,"created_at":"2026-04-11T12:00:00Z"}`,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			productID:      "abc",
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product ID format"}`,
		},
		{
			name:      "Bad Request - Invalid ID",
			productID: "2",
			mockSetup: func(s *mocks.Service) {
				s.On("GetProduct", mock.Anything, int32(2)).Return(model.Product{}, model.ErrInvalidProductID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product id"}`,
		},
		{
			name:      "Not Found",
			productID: "99",
			mockSetup: func(s *mocks.Service) {
				s.On("GetProduct", mock.Anything, int32(99)).Return(model.Product{}, model.ErrProductNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"product not found"}`,
		},
		{
			name:      "Internal Server Error - DB",
			productID: "1",
			mockSetup: func(s *mocks.Service) {
				s.On("GetProduct", mock.Anything, int32(1)).Return(model.Product{}, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Get("/api/v1/products/{id}", h.GetProduct)

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

func TestHandler_ListProductsByWarehouse(t *testing.T) {
	type testCase struct {
		name           string
		warehouseID    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:        "Success",
			warehouseID: "1",
			mockSetup: func(s *mocks.Service) {
				pList := []model.Product{{ID: 1, WarehouseID: 1, Name: "P1", Quantity: 10}}
				s.On("ListProductsByWarehouse", mock.Anything, int32(1)).Return(pList, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"warehouse_id":1,"name":"P1","quantity":10,"created_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:        "Bad Request - Invalid Warehouse ID",
			warehouseID: "99",
			mockSetup: func(s *mocks.Service) {
				s.On("ListProductsByWarehouse", mock.Anything, int32(99)).Return(nil, model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Internal Server Error - DB",
			warehouseID: "1",
			mockSetup: func(s *mocks.Service) {
				s.On("ListProductsByWarehouse", mock.Anything, int32(1)).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Get("/api/v1/warehouses/{id}/products", h.ListProductsByWarehouse)

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

func TestHandler_SetProductQuantity(t *testing.T) {
	type testCase struct {
		name           string
		productID      string
		requestBody    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:        "Success",
			productID:   "1",
			requestBody: `{"quantity":50}`,
			mockSetup: func(s *mocks.Service) {
				s.On("SetProductQuantity", mock.Anything, int32(1), int32(50)).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			productID:      "abc",
			requestBody:    `{"quantity":50}`,
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product ID format"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			productID:      "1",
			requestBody:    `{"quantity":}`,
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
		{
			name:        "Bad Request - Invalid Product ID",
			productID:   "1",
			requestBody: `{"quantity":50}`,
			mockSetup: func(s *mocks.Service) {
				s.On("SetProductQuantity", mock.Anything, int32(1), int32(50)).Return(model.ErrInvalidProductID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product id"}`,
		},
		{
			name:        "Internal Server Error - DB",
			productID:   "1",
			requestBody: `{"quantity":50}`,
			mockSetup: func(s *mocks.Service) {
				s.On("SetProductQuantity", mock.Anything, int32(1), int32(50)).Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Patch("/api/v1/products/{id}/quantity", h.SetProductQuantity)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+tc.productID+"/quantity", bytes.NewBufferString(tc.requestBody))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_AddProductQuantity(t *testing.T) {
	type testCase struct {
		name           string
		productID      string
		requestBody    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:        "Success",
			productID:   "1",
			requestBody: `{"quantity":-5}`,
			mockSetup: func(s *mocks.Service) {
				s.On("AddProductQuantity", mock.Anything, int32(1), int32(-5)).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			productID:      "abc",
			requestBody:    `{"quantity":5}`,
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product ID format"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			productID:      "1",
			requestBody:    `{"quantity":"lots"}`,
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
		{
			name:        "Bad Request - Invalid Product ID",
			productID:   "1",
			requestBody: `{"quantity":5}`,
			mockSetup: func(s *mocks.Service) {
				s.On("AddProductQuantity", mock.Anything, int32(1), int32(5)).Return(model.ErrInvalidProductID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid product id"}`,
		},
		{
			name:        "Conflict - Not Enough Quantity",
			productID:   "1",
			requestBody: `{"quantity":-100}`,
			mockSetup: func(s *mocks.Service) {
				s.On("AddProductQuantity", mock.Anything, int32(1), int32(-100)).Return(model.ErrNotEnoughQuantity).Once()
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"not enough quantity"}`,
		},
		{
			name:        "Internal Server Error - DB",
			productID:   "1",
			requestBody: `{"quantity":5}`,
			mockSetup: func(s *mocks.Service) {
				s.On("AddProductQuantity", mock.Anything, int32(1), int32(5)).Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Patch("/api/v1/products/{id}/add", h.AddProductQuantity)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+tc.productID+"/add", bytes.NewBufferString(tc.requestBody))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}
