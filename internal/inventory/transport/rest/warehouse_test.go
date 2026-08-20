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

func TestWarehouseHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(s *mocks.WarehouseService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			requestBody: `{"name": "Main", "location": "Moscow"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Create", mock.Anything, "Main", "Moscow").
					Return(&model.Warehouse{ID: 1, Name: "Main", Location: "Moscow", CreatedAt: mockTime}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id": 1, "name": "Main", "location": "Moscow", "created_at": "2026-04-11T12:00:00Z"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			requestBody:    `{"name": "Main, "location": "Moscow"}`,
			mockSetup:      func(s *mocks.WarehouseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "invalid request body"}`,
		},
		{
			name:        "Bad Request - Empty Name",
			requestBody: `{"name": "", "location": "Moscow"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Create", mock.Anything, "", "Moscow").Return(nil, model.ErrEmptyWarehouseName).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "warehouse name cannot be empty"}`,
		},
		{
			name:        "Bad Request - Empty Location",
			requestBody: `{"name": "Main", "location": ""}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Create", mock.Anything, "Main", "").Return(nil, model.ErrEmptyWarehouseLocation).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "warehouse location cannot be empty"}`,
		},
		{
			name:        "Internal Server Error",
			requestBody: `{"name": "Main", "location": "Moscow"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Create", mock.Anything, "Main", "Moscow").Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": "internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewWarehouseService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewWarehouseHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Post("/api/v1/warehouses", h.Create)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouses", bytes.NewBufferString(tc.requestBody))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestWarehouseHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		warehouseID    string
		mockSetup      func(s *mocks.WarehouseService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			warehouseID: "1",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Get", mock.Anything, int32(1)).
					Return(&model.Warehouse{ID: 1, Name: "A", Location: "B", CreatedAt: mockTime}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"A","location":"B","created_at":"2026-04-11T12:00:00Z"}`,
		},
		{
			name:           "Bad Request - Invalid Param format",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.WarehouseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:        "Bad Request - Invalid ID logic",
			warehouseID: "0",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Get", mock.Anything, int32(0)).Return(nil, model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Not Found",
			warehouseID: "99",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Get", mock.Anything, int32(99)).Return(nil, model.ErrWarehouseNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"warehouse not found"}`,
		},
		{
			name:        "Internal Server Error",
			warehouseID: "1",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Get", mock.Anything, int32(1)).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewWarehouseService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewWarehouseHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Get("/api/v1/warehouses/{id}", h.Get)

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

func TestWarehouseHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(s *mocks.WarehouseService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("List", mock.Anything).Return([]model.Warehouse{{ID: 1, Name: "W1", Location: "L1", CreatedAt: mockTime}}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"name":"W1","location":"L1","created_at":"2026-04-11T12:00:00Z"}]`,
		},
		{
			name: "Internal Server Error",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("List", mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewWarehouseService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewWarehouseHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Get("/api/v1/warehouses", h.List)

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

func TestWarehouseHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		warehouseID    string
		requestBody    string
		mockSetup      func(s *mocks.WarehouseService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			warehouseID: "1",
			requestBody: `{"name":"New","location":"Loc"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Update", mock.Anything, int32(1), "New", "Loc").Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			warehouseID:    "abc",
			requestBody:    `{"name":"New","location":"Loc"}`,
			mockSetup:      func(s *mocks.WarehouseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			warehouseID:    "1",
			requestBody:    `{"name":"New,}`,
			mockSetup:      func(s *mocks.WarehouseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
		{
			name:        "Bad Request - Empty Name",
			warehouseID: "1",
			requestBody: `{"name":"","location":"Loc"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Update", mock.Anything, int32(1), "", "Loc").Return(model.ErrEmptyWarehouseName).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"warehouse name cannot be empty"}`,
		},
		{
			name:        "Not Found",
			warehouseID: "1",
			requestBody: `{"name":"New","location":"Loc"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Update", mock.Anything, int32(1), "New", "Loc").Return(model.ErrWarehouseNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"warehouse not found"}`,
		},
		{
			name:        "Internal Server Error",
			warehouseID: "1",
			requestBody: `{"name":"New","location":"Loc"}`,
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Update", mock.Anything, int32(1), "New", "Loc").Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewWarehouseService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewWarehouseHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Put("/api/v1/warehouses/{id}", h.Update)

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

func TestWarehouseHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		warehouseID    string
		mockSetup      func(s *mocks.WarehouseService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			warehouseID: "1",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Delete", mock.Anything, int32(1)).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Bad Request - Invalid ID Format",
			warehouseID:    "abc",
			mockSetup:      func(s *mocks.WarehouseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse ID format"}`,
		},
		{
			name:        "Bad Request - Invalid ID logic",
			warehouseID: "0",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Delete", mock.Anything, int32(0)).Return(model.ErrInvalidWarehouseID).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid warehouse id"}`,
		},
		{
			name:        "Not Found",
			warehouseID: "99",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Delete", mock.Anything, int32(99)).Return(model.ErrWarehouseNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"warehouse not found"}`,
		},
		{
			name:        "Internal Server Error",
			warehouseID: "1",
			mockSetup: func(s *mocks.WarehouseService) {
				s.On("Delete", mock.Anything, int32(1)).Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewWarehouseService(t)
			tc.mockSetup(mockSvc)
			h := rest.NewWarehouseHandler(mockSvc, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := chi.NewRouter()
			r.Delete("/api/v1/warehouses/{id}", h.Delete)

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
