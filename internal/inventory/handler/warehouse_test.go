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
	"github.com/kelar1s/go-freight/internal/inventory/handler"
	"github.com/kelar1s/go-freight/internal/inventory/handler/mocks"
	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateWarehouse(t *testing.T) {
	type testCase struct {
		name           string
		requestBody    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	mockTime := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

	tests := []testCase{
		{
			name:        "Success",
			requestBody: `{"name": "Main Warehouse", "location": "Moscow"}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateWarehouse", mock.Anything, "Main Warehouse", "Moscow").
					Return(model.Warehouse{ID: 1, Name: "Main Warehouse", Location: "Moscow", CreatedAt: mockTime}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id": 1, "name": "Main Warehouse", "location": "Moscow", "created_at": "2026-04-11T12:00:00Z"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			requestBody:    `{"name": "Main Warehouse, "location": "Moscow"}`,
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "invalid request body"}`,
		},
		{
			name:        "Bad Request - Empty Name",
			requestBody: `{"name": "", "location": "Moscow"}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateWarehouse", mock.Anything, "", "Moscow").
					Return(model.Warehouse{}, model.ErrEmptyWarehouseName).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "warehouse name cannot be empty"}`,
		},
		{
			name:        "Bad Request - Empty Location",
			requestBody: `{"name": "Main", "location": ""}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateWarehouse", mock.Anything, "Main", "").
					Return(model.Warehouse{}, model.ErrEmptyWarehouseLocation).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "warehouse location cannot be empty"}`,
		},
		{
			name:        "Internal Server Error - DB",
			requestBody: `{"name": "Main", "location": "Moscow"}`,
			mockSetup: func(s *mocks.Service) {
				s.On("CreateWarehouse", mock.Anything, "Main", "Moscow").
					Return(model.Warehouse{}, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": "internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Post("/api/v1/warehouses", h.CreateWarehouse)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouses", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_DeleteWarehouse(t *testing.T) {
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
				s.On("DeleteWarehouse", mock.Anything, int32(1)).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:        "Bad Request - Invalid ID",
			warehouseID: "999",
			mockSetup: func(s *mocks.Service) {
				s.On("DeleteWarehouse", mock.Anything, int32(999)).Return(model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Internal Server Error - DB",
			warehouseID: "1",
			mockSetup: func(s *mocks.Service) {
				s.On("DeleteWarehouse", mock.Anything, int32(1)).Return(errors.New("db error")).Once()
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
			r.Delete("/api/v1/warehouses/{id}", h.DeleteWarehouse)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/warehouses/"+tc.warehouseID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_GetWarehouse(t *testing.T) {
	type testCase struct {
		name           string
		warehouseID    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	mockTime := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

	tests := []testCase{
		{
			name:        "Success",
			warehouseID: "1",
			mockSetup: func(s *mocks.Service) {
				s.On("GetWarehouse", mock.Anything, int32(1)).
					Return(model.Warehouse{ID: 1, Name: "A", Location: "B", CreatedAt: mockTime}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"A","location":"B","created_at":"2026-04-11T12:00:00Z"}`,
		},
		{
			name:        "Not Found",
			warehouseID: "99",
			mockSetup: func(s *mocks.Service) {
				s.On("GetWarehouse", mock.Anything, int32(99)).Return(model.Warehouse{}, model.ErrWarehouseNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"warehouse not found"}`,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:        "Bad Request - Invalid ID",
			warehouseID: "123",
			mockSetup: func(s *mocks.Service) {
				s.On("GetWarehouse", mock.Anything, int32(123)).Return(model.Warehouse{}, model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Internal Server Error - DB",
			warehouseID: "123",
			mockSetup: func(s *mocks.Service) {
				s.On("GetWarehouse", mock.Anything, int32(123)).Return(model.Warehouse{}, errors.New("db error")).Once()
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
			r.Get("/api/v1/warehouses/{id}", h.GetWarehouse)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/"+tc.warehouseID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_ListWarehouses(t *testing.T) {
	type testCase struct {
		name           string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name: "Success",
			mockSetup: func(s *mocks.Service) {
				s.On("ListWarehouses", mock.Anything).Return([]model.Warehouse{{ID: 1, Name: "W1", Location: "L1"}}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"name":"W1","location":"L1","created_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name: "Internal Server Error - DB",
			mockSetup: func(s *mocks.Service) {
				s.On("ListWarehouses", mock.Anything).Return(nil, errors.New("db error")).Once()
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
			r.Get("/api/v1/warehouses", h.ListWarehouses)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_UpdateWarehouse(t *testing.T) {
	type testCase struct {
		name           string
		warehouseID    string
		requestBody    string
		mockSetup      func(s *mocks.Service)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:        "Success",
			warehouseID: "1",
			requestBody: `{"name":"New","location":"Loc"}`,
			mockSetup: func(s *mocks.Service) {
				s.On("UpdateWarehouse", mock.Anything, int32(1), "New", "Loc").Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			warehouseID:    "1",
			requestBody:    `{"name": "New Name, "location": "New Location"}`,
			mockSetup:      func(s *mocks.Service) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "invalid request body"}`,
		},
		{
			name:        "Bad Request - Empty Name",
			warehouseID: "1",
			requestBody: `{"name":"","location":"Loc"}`,
			mockSetup: func(s *mocks.Service) {
				s.On("UpdateWarehouse", mock.Anything, int32(1), "", "Loc").Return(model.ErrEmptyWarehouseName).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"warehouse name cannot be empty"}`,
		},
		{
			name:        "Bad Request - Empty Location",
			warehouseID: "1",
			requestBody: `{"name":"Name","location":""}`,
			mockSetup: func(s *mocks.Service) {
				s.On("UpdateWarehouse", mock.Anything, int32(1), "Name", "").Return(model.ErrEmptyWarehouseLocation).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"warehouse location cannot be empty"}`,
		},
		{
			name:        "Internal Server Error - DB",
			warehouseID: "1",
			requestBody: `{"name": "Main", "location": "Loc"}`,
			mockSetup: func(s *mocks.Service) {
				s.On("UpdateWarehouse", mock.Anything, int32(1), "Main", "Loc").Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": "internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewService(t)
			tc.mockSetup(mockSvc)
			h := handler.NewProductHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Put("/api/v1/warehouses/{id}", h.UpdateWarehouse)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/warehouses/"+tc.warehouseID, bytes.NewBufferString(tc.requestBody))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}
