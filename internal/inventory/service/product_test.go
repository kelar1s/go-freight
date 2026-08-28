package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/inventory/service/mocks"
)

var (
	mockTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	mockTTL  = 5 * time.Minute
	errRepo  = errors.New("repository error")
	errCache = errors.New("cache miss")
)

func TestProductService_Create(t *testing.T) {
	type TestCase struct {
		name           string
		warehouseID    int64
		inputName      string
		quantity       int64
		mockSetup      func(r *mocks.ProductRepo, c *mocks.Cache)
		expectedResult *model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:        "Success",
			warehouseID: 1,
			inputName:   " Box ",
			quantity:    10,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				r.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *model.Product) bool {
					return p.WarehouseID == 1 && p.Name == "Box" && p.Quantity == 10
				})).Run(func(ctx context.Context, p *model.Product) {
					p.ID = 1
					p.CreatedAt = mockTime
				}).Return(nil).Once()
			},
			expectedResult: &model.Product{
				ProductMeta:  model.ProductMeta{ID: 1, WarehouseID: 1, Name: "Box", CreatedAt: mockTime},
				ProductStock: model.ProductStock{Quantity: 10, Reserved: 0},
			},
			expectedError: nil,
		},
		{
			name:           "Error - Invalid WH ID",
			warehouseID:    0,
			inputName:      "Box",
			quantity:       10,
			mockSetup:      func(r *mocks.ProductRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidWarehouseID,
		},
		{
			name:           "Error - Negative Quantity",
			warehouseID:    1,
			inputName:      "Box",
			quantity:       -5,
			mockSetup:      func(r *mocks.ProductRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidQuantity,
		},
		{
			name:           "Error - Empty Name",
			warehouseID:    1,
			inputName:      "   ",
			quantity:       10,
			mockSetup:      func(r *mocks.ProductRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrEmptyProductName,
		},
		{
			name:        "Error - Repo Error",
			warehouseID: 1,
			inputName:   "Box",
			quantity:    10,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				r.EXPECT().Create(mock.Anything, mock.Anything).Return(errRepo).Once()
			},
			expectedResult: nil,
			expectedError:  errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			res, err := svc.Create(context.Background(), tc.warehouseID, tc.inputName, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestProductService_Get(t *testing.T) {
	type TestCase struct {
		name           string
		inputID        int64
		mockSetup      func(r *mocks.ProductRepo, c *mocks.Cache)
		expectedResult *model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:    "Success - Cache Miss",
			inputID: 1,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				c.EXPECT().Get(mock.Anything, "product:meta:1", mock.Anything).Return(errCache).Twice()
				meta := &model.ProductMeta{ID: 1, WarehouseID: 1, Name: "Box", CreatedAt: mockTime}
				r.EXPECT().GetMeta(mock.Anything, int64(1)).Return(meta, nil).Once()
				c.EXPECT().Set(mock.Anything, "product:meta:1", meta, mockTTL).Return(nil).Once()
				stock := &model.ProductStock{Quantity: 10, Reserved: 2}
				r.EXPECT().GetStock(mock.Anything, int64(1)).Return(stock, nil).Once()
			},
			expectedResult: &model.Product{
				ProductMeta:  model.ProductMeta{ID: 1, WarehouseID: 1, Name: "Box", CreatedAt: mockTime},
				ProductStock: model.ProductStock{Quantity: 10, Reserved: 2},
			},
			expectedError: nil,
		},
		{
			name:    "Success - Cache Hit",
			inputID: 2,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				c.EXPECT().Get(mock.Anything, "product:meta:2", mock.Anything).Run(func(ctx context.Context, key string, dest interface{}) {
					*dest.(**model.ProductMeta) = &model.ProductMeta{ID: 2, WarehouseID: 1, Name: "Pallet", CreatedAt: mockTime}
				}).Return(nil).Once()
				stock := &model.ProductStock{Quantity: 50, Reserved: 0}
				r.EXPECT().GetStock(mock.Anything, int64(2)).Return(stock, nil).Once()
			},
			expectedResult: &model.Product{
				ProductMeta:  model.ProductMeta{ID: 2, WarehouseID: 1, Name: "Pallet", CreatedAt: mockTime},
				ProductStock: model.ProductStock{Quantity: 50, Reserved: 0},
			},
			expectedError: nil,
		},
		{
			name:           "Error - Invalid ID",
			inputID:        0,
			mockSetup:      func(r *mocks.ProductRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidProductID,
		},
		{
			name:    "Error - GetMeta Repo Error",
			inputID: 1,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				c.EXPECT().Get(mock.Anything, "product:meta:1", mock.Anything).Return(errCache).Twice()
				r.EXPECT().GetMeta(mock.Anything, int64(1)).Return(nil, errRepo).Once()
			},
			expectedResult: nil,
			expectedError:  errRepo,
		},
		{
			name:    "Error - GetStock Repo Error",
			inputID: 1,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				c.EXPECT().Get(mock.Anything, "product:meta:1", mock.Anything).Run(func(ctx context.Context, key string, dest interface{}) {
					*dest.(**model.ProductMeta) = &model.ProductMeta{ID: 1, WarehouseID: 1, Name: "Box", CreatedAt: mockTime}
				}).Return(nil).Once()
				r.EXPECT().GetStock(mock.Anything, int64(1)).Return(nil, errRepo).Once()
			},
			expectedResult: nil,
			expectedError:  errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			res, err := svc.Get(context.Background(), tc.inputID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestProductService_ListByWarehouse(t *testing.T) {
	type TestCase struct {
		name           string
		warehouseID    int64
		mockSetup      func(r *mocks.ProductRepo)
		expectedResult []model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:        "Success",
			warehouseID: 1,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().ListByWarehouse(mock.Anything, int64(1)).Return([]model.Product{
					{ProductMeta: model.ProductMeta{ID: 1, Name: "A"}},
				}, nil).Once()
			},
			expectedResult: []model.Product{{ProductMeta: model.ProductMeta{ID: 1, Name: "A"}}},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid WH ID",
			warehouseID:    0,
			mockSetup:      func(r *mocks.ProductRepo) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidWarehouseID,
		},
		{
			name:        "Error - Repo Error",
			warehouseID: 1,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().ListByWarehouse(mock.Anything, int64(1)).Return(nil, errRepo).Once()
			},
			expectedResult: nil,
			expectedError:  errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			res, err := svc.ListByWarehouse(context.Background(), tc.warehouseID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestProductService_Delete(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		mockSetup     func(r *mocks.ProductRepo, c *mocks.Cache)
		expectedError error
	}

	tests := []TestCase{
		{
			name:    "Success",
			inputID: 1,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				r.EXPECT().Delete(mock.Anything, int64(1)).Return(nil).Once()
				c.EXPECT().Delete(mock.Anything, "product:meta:1").Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			mockSetup:     func(r *mocks.ProductRepo, c *mocks.Cache) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:    "Error - Repo Error",
			inputID: 1,
			mockSetup: func(r *mocks.ProductRepo, c *mocks.Cache) {
				r.EXPECT().Delete(mock.Anything, int64(1)).Return(errRepo).Once()
			},
			expectedError: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			err := svc.Delete(context.Background(), tc.inputID)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestProductService_AdjustQuantity(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		quantity      int64
		mockSetup     func(r *mocks.ProductRepo)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 10,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().AdjustQuantity(mock.Anything, int64(1), int64(10)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       -1,
			quantity:      10,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Zero Quantity",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Error",
			inputID:  1,
			quantity: -100,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().AdjustQuantity(mock.Anything, int64(1), int64(-100)).Return(errRepo).Once()
			},
			expectedError: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			err := svc.AdjustQuantity(context.Background(), tc.inputID, tc.quantity)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestProductService_Reserve(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		quantity      int64
		mockSetup     func(r *mocks.ProductRepo)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().Reserve(mock.Anything, int64(1), int64(5)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      5,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Error",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().Reserve(mock.Anything, int64(1), int64(5)).Return(errRepo).Once()
			},
			expectedError: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			err := svc.Reserve(context.Background(), tc.inputID, tc.quantity)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestProductService_Release(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		quantity      int64
		mockSetup     func(r *mocks.ProductRepo)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().Release(mock.Anything, int64(1), int64(5)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      5,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Error",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().Release(mock.Anything, int64(1), int64(5)).Return(errRepo).Once()
			},
			expectedError: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			err := svc.Release(context.Background(), tc.inputID, tc.quantity)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestProductService_CancelReservation(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		quantity      int64
		mockSetup     func(r *mocks.ProductRepo)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().CancelReservation(mock.Anything, int64(1), int64(5)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      5,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Error",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().CancelReservation(mock.Anything, int64(1), int64(5)).Return(errRepo).Once()
			},
			expectedError: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo)

			svc := service.NewProductService(mockRepo, mockCache, mockTTL)
			err := svc.CancelReservation(context.Background(), tc.inputID, tc.quantity)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
