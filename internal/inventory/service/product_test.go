package service_test

import (
	"context"
	"testing"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/inventory/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductService_Create(t *testing.T) {
	type TestCase struct {
		name           string
		warehouseID    int32
		inputName      string
		quantity       int32
		mockSetup      func(r *mocks.ProductRepo)
		expectedResult *model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:        "Success",
			warehouseID: 1,
			inputName:   " Box ",
			quantity:    10,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *model.Product) bool {
					return p.WarehouseID == 1 && p.Name == "Box" && p.Quantity == 10
				})).Run(func(ctx context.Context, p *model.Product) {
					p.ID = 1
					p.CreatedAt = mockTime
				}).Return(nil).Once()
			},

			expectedResult: &model.Product{ID: 1, WarehouseID: 1, Name: "Box", Quantity: 10, CreatedAt: mockTime},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid WH ID",
			warehouseID:    0,
			inputName:      "Box",
			quantity:       10,
			mockSetup:      func(r *mocks.ProductRepo) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidWarehouseID,
		},
		{
			name:           "Error - Negative Quantity",
			warehouseID:    1,
			inputName:      "Box",
			quantity:       -5,
			mockSetup:      func(r *mocks.ProductRepo) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidQuantity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			tc.mockSetup(mockRepo)
			svc := service.NewProductService(mockRepo)

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
		inputID        int32
		mockSetup      func(r *mocks.ProductRepo)
		expectedResult *model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:    "Success",
			inputID: 1,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().Get(mock.Anything, int32(1)).Return(&model.Product{ID: 1, Name: "Box"}, nil).Once()
			},
			expectedResult: &model.Product{ID: 1, Name: "Box"},
			expectedError:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			tc.mockSetup(mockRepo)
			svc := service.NewProductService(mockRepo)

			res, err := svc.Get(context.Background(), tc.inputID)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}

func TestProductService_AddQuantity(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.ProductRepo)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 10,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().AddQuantity(mock.Anything, int32(1), int32(10)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Zero Quantity",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Not Enough",
			inputID:  1,
			quantity: -100,
			mockSetup: func(r *mocks.ProductRepo) {
				r.EXPECT().AddQuantity(mock.Anything, int32(1), int32(-100)).Return(model.ErrNotEnoughQuantity).Once()
			},
			expectedError: model.ErrNotEnoughQuantity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			tc.mockSetup(mockRepo)
			svc := service.NewProductService(mockRepo)

			err := svc.AddQuantity(context.Background(), tc.inputID, tc.quantity)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestProductService_Reserve(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.ProductRepo)
		expectedError error
	}

	tests := []TestCase{
		{
			name:          "Success",
			inputID:       1,
			quantity:      5,
			mockSetup:     func(r *mocks.ProductRepo) { r.EXPECT().Reserve(mock.Anything, int32(1), int32(5)).Return(nil).Once() },
			expectedError: nil,
		},
		{
			name:          "Error - Negative Quantity",
			inputID:       1,
			quantity:      -5,
			mockSetup:     func(r *mocks.ProductRepo) {},
			expectedError: model.ErrInvalidQuantity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepo(t)
			tc.mockSetup(mockRepo)
			svc := service.NewProductService(mockRepo)

			err := svc.Reserve(context.Background(), tc.inputID, tc.quantity)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
