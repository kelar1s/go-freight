package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/inventory/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
var errRepoExplosion = errors.New("repo explosion")

func TestService_CreateWarehouse(t *testing.T) {
	type TestCase struct {
		name           string
		inputName      string
		inputLocation  string
		mockSetup      func(r *mocks.Repository)
		expectedResult model.Warehouse
		expectedError  error
	}

	tests := []TestCase{
		{
			name:          "Success with trim",
			inputName:     "  Central Moscow  ",
			inputLocation: "  Russia Moscow  ",
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CreateWarehouse(mock.Anything, "Central Moscow", "Russia Moscow").
					Return(model.Warehouse{ID: 1, Name: "Central Moscow", Location: "Russia Moscow", CreatedAt: mockTime}, nil).Once()
			},
			expectedResult: model.Warehouse{ID: 1, Name: "Central Moscow", Location: "Russia Moscow", CreatedAt: mockTime},
			expectedError:  nil,
		},
		{
			name:           "Error - Empty Name",
			inputName:      "   ",
			inputLocation:  "Moscow",
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Warehouse{},
			expectedError:  model.ErrEmptyWarehouseName,
		},
		{
			name:           "Error - Empty Location",
			inputName:      "Main",
			inputLocation:  "   ",
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Warehouse{},
			expectedError:  model.ErrEmptyWarehouseLocation,
		},
		{
			name:          "Error - Repo Failure",
			inputName:     "Main",
			inputLocation: "Moscow",
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CreateWarehouse(mock.Anything, "Main", "Moscow").Return(model.Warehouse{}, errRepoExplosion).Once()
			},
			expectedResult: model.Warehouse{},
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			res, err := svc.CreateWarehouse(context.Background(), tc.inputName, tc.inputLocation)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestService_DeleteWarehouse(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:          "Success",
			inputID:       1,
			mockSetup:     func(r *mocks.Repository) { r.EXPECT().DeleteWarehouse(mock.Anything, int32(1)).Return(nil).Once() },
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidWarehouseID,
		},
		{
			name:    "Error - Repo Failure",
			inputID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().DeleteWarehouse(mock.Anything, int32(1)).Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.DeleteWarehouse(context.Background(), tc.inputID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetWarehouse(t *testing.T) {
	type TestCase struct {
		name           string
		inputID        int32
		mockSetup      func(r *mocks.Repository)
		expectedResult model.Warehouse
		expectedError  error
	}

	tests := []TestCase{
		{
			name:    "Success",
			inputID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().GetWarehouse(mock.Anything, int32(1)).Return(model.Warehouse{ID: 1, Name: "Main"}, nil).Once()
			},
			expectedResult: model.Warehouse{ID: 1, Name: "Main"},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid ID",
			inputID:        -5,
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Warehouse{},
			expectedError:  model.ErrInvalidWarehouseID,
		},
		{
			name:    "Error - Repo Failure",
			inputID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().GetWarehouse(mock.Anything, int32(1)).Return(model.Warehouse{}, errRepoExplosion).Once()
			},
			expectedResult: model.Warehouse{},
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			res, err := svc.GetWarehouse(context.Background(), tc.inputID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestService_ListWarehouses(t *testing.T) {
	type TestCase struct {
		name           string
		mockSetup      func(r *mocks.Repository)
		expectedResult []model.Warehouse
		expectedError  error
	}

	tests := []TestCase{
		{
			name: "Success",
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ListWarehouses(mock.Anything).Return([]model.Warehouse{{ID: 1, Name: "Main"}}, nil).Once()
			},
			expectedResult: []model.Warehouse{{ID: 1, Name: "Main"}},
			expectedError:  nil,
		},
		{
			name: "Error - Repo Failure",
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ListWarehouses(mock.Anything).Return(nil, errRepoExplosion).Once()
			},
			expectedResult: nil,
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			res, err := svc.ListWarehouses(context.Background())

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

func TestService_UpdateWarehouse(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		inputName     string
		inputLocation string
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:          "Success with trim",
			inputID:       1,
			inputName:     " New ",
			inputLocation: " Loc ",
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().UpdateWarehouse(mock.Anything, int32(1), "New", "Loc").Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			inputName:     "New",
			inputLocation: "Loc",
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidWarehouseID,
		},
		{
			name:          "Error - Empty Name",
			inputID:       1,
			inputName:     "  ",
			inputLocation: "Loc",
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrEmptyWarehouseName,
		},
		{
			name:          "Error - Empty Location",
			inputID:       1,
			inputName:     "New",
			inputLocation: "  ",
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrEmptyWarehouseLocation,
		},
		{
			name:          "Error - Repo Failure",
			inputID:       1,
			inputName:     "New",
			inputLocation: "Loc",
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().UpdateWarehouse(mock.Anything, int32(1), "New", "Loc").Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.UpdateWarehouse(context.Background(), tc.inputID, tc.inputName, tc.inputLocation)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CreateProduct(t *testing.T) {
	type TestCase struct {
		name           string
		warehouseID    int32
		inputName      string
		quantity       int32
		mockSetup      func(r *mocks.Repository)
		expectedResult model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:        "Success",
			warehouseID: 1,
			inputName:   " Box ",
			quantity:    10,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CreateProduct(mock.Anything, int32(1), "Box", int32(10)).
					Return(model.Product{ID: 1, WarehouseID: 1, Name: "Box", Quantity: 10, Reserved: 0}, nil).Once()
			},
			expectedResult: model.Product{ID: 1, WarehouseID: 1, Name: "Box", Quantity: 10, Reserved: 0},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid WH ID",
			warehouseID:    0,
			inputName:      "Box",
			quantity:       10,
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Product{},
			expectedError:  model.ErrInvalidWarehouseID,
		},
		{
			name:           "Error - Invalid Quantity",
			warehouseID:    1,
			inputName:      "Box",
			quantity:       -5,
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Product{},
			expectedError:  model.ErrInvalidQuantity,
		},
		{
			name:           "Error - Empty Name",
			warehouseID:    1,
			inputName:      "  ",
			quantity:       10,
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Product{},
			expectedError:  model.ErrEmptyProductName,
		},
		{
			name:        "Error - Repo Failure",
			warehouseID: 1,
			inputName:   "Box",
			quantity:    10,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CreateProduct(mock.Anything, int32(1), "Box", int32(10)).Return(model.Product{}, errRepoExplosion).Once()
			},
			expectedResult: model.Product{},
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			res, err := svc.CreateProduct(context.Background(), tc.warehouseID, tc.inputName, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestService_DeleteProduct(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:          "Success",
			inputID:       1,
			mockSetup:     func(r *mocks.Repository) { r.EXPECT().DeleteProduct(mock.Anything, int32(1)).Return(nil).Once() },
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       -1,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:    "Error - Repo Failure",
			inputID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().DeleteProduct(mock.Anything, int32(1)).Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.DeleteProduct(context.Background(), tc.inputID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetProduct(t *testing.T) {
	type TestCase struct {
		name           string
		inputID        int32
		mockSetup      func(r *mocks.Repository)
		expectedResult model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:    "Success",
			inputID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().GetProduct(mock.Anything, int32(1)).Return(model.Product{ID: 1, Name: "Box", Reserved: 0}, nil).Once()
			},
			expectedResult: model.Product{ID: 1, Name: "Box", Reserved: 0},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid ID",
			inputID:        0,
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: model.Product{},
			expectedError:  model.ErrInvalidProductID,
		},
		{
			name:    "Error - Repo Failure",
			inputID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().GetProduct(mock.Anything, int32(1)).Return(model.Product{}, errRepoExplosion).Once()
			},
			expectedResult: model.Product{},
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			res, err := svc.GetProduct(context.Background(), tc.inputID)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestService_ListProductsByWarehouse(t *testing.T) {
	type TestCase struct {
		name           string
		warehouseID    int32
		mockSetup      func(r *mocks.Repository)
		expectedResult []model.Product
		expectedError  error
	}

	tests := []TestCase{
		{
			name:        "Success",
			warehouseID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ListProductsByWarehouse(mock.Anything, int32(1)).Return([]model.Product{{ID: 1, Name: "Box", Reserved: 0}}, nil).Once()
			},
			expectedResult: []model.Product{{ID: 1, Name: "Box", Reserved: 0}},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid WH ID",
			warehouseID:    0,
			mockSetup:      func(r *mocks.Repository) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidWarehouseID,
		},
		{
			name:        "Error - Repo Failure",
			warehouseID: 1,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ListProductsByWarehouse(mock.Anything, int32(1)).Return(nil, errRepoExplosion).Once()
			},
			expectedResult: nil,
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			res, err := svc.ListProductsByWarehouse(context.Background(), tc.warehouseID)

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

func TestService_SetProductQuantity(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 50,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().SetProductQuantity(mock.Anything, int32(1), int32(50)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      50,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity",
			inputID:       1,
			quantity:      -5,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Failure",
			inputID:  1,
			quantity: 50,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().SetProductQuantity(mock.Anything, int32(1), int32(50)).Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.SetProductQuantity(context.Background(), tc.inputID, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_AddProductQuantity(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success (Positive)",
			inputID:  1,
			quantity: 10,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().AddProductQuantity(mock.Anything, int32(1), int32(10)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:     "Success (Negative)",
			inputID:  1,
			quantity: -5,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().AddProductQuantity(mock.Anything, int32(1), int32(-5)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      10,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:     "Error - Repo Failure (Not Enough Quantity)",
			inputID:  1,
			quantity: -100,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().AddProductQuantity(mock.Anything, int32(1), int32(-100)).Return(model.ErrNotEnoughQuantity).Once()
			},
			expectedError: model.ErrNotEnoughQuantity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.AddProductQuantity(context.Background(), tc.inputID, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ReserveProduct(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ReserveProduct(mock.Anything, int32(1), int32(5)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      5,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity (Zero)",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:          "Error - Invalid Quantity (Negative)",
			inputID:       1,
			quantity:      -5,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Failure (Not Enough Quantity)",
			inputID:  1,
			quantity: 100,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ReserveProduct(mock.Anything, int32(1), int32(100)).Return(model.ErrNotEnoughQuantity).Once()
			},
			expectedError: model.ErrNotEnoughQuantity,
		},
		{
			name:     "Error - Repo Failure",
			inputID:  1,
			quantity: 5,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ReserveProduct(mock.Anything, int32(1), int32(5)).Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.ReserveProduct(context.Background(), tc.inputID, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ReleaseProduct(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 3,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ReleaseProduct(mock.Anything, int32(1), int32(3)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      3,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity (Zero)",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:          "Error - Invalid Quantity (Negative)",
			inputID:       1,
			quantity:      -3,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Failure (Not Enough Reserved)",
			inputID:  1,
			quantity: 100,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ReleaseProduct(mock.Anything, int32(1), int32(100)).Return(model.ErrNotEnoughQuantity).Once()
			},
			expectedError: model.ErrNotEnoughQuantity,
		},
		{
			name:     "Error - Repo Failure",
			inputID:  1,
			quantity: 3,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().ReleaseProduct(mock.Anything, int32(1), int32(3)).Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.ReleaseProduct(context.Background(), tc.inputID, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CancelReservation(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int32
		quantity      int32
		mockSetup     func(r *mocks.Repository)
		expectedError error
	}

	tests := []TestCase{
		{
			name:     "Success",
			inputID:  1,
			quantity: 2,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CancelReservation(mock.Anything, int32(1), int32(2)).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			quantity:      2,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidProductID,
		},
		{
			name:          "Error - Invalid Quantity (Zero)",
			inputID:       1,
			quantity:      0,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:          "Error - Invalid Quantity (Negative)",
			inputID:       1,
			quantity:      -2,
			mockSetup:     func(r *mocks.Repository) {},
			expectedError: model.ErrInvalidQuantity,
		},
		{
			name:     "Error - Repo Failure (Not Enough Reserved)",
			inputID:  1,
			quantity: 100,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CancelReservation(mock.Anything, int32(1), int32(100)).Return(model.ErrNotEnoughQuantity).Once()
			},
			expectedError: model.ErrNotEnoughQuantity,
		},
		{
			name:     "Error - Repo Failure",
			inputID:  1,
			quantity: 2,
			mockSetup: func(r *mocks.Repository) {
				r.EXPECT().CancelReservation(mock.Anything, int32(1), int32(2)).Return(errRepoExplosion).Once()
			},
			expectedError: errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			tc.mockSetup(mockRepo)
			svc := service.NewInventoryService(mockRepo)

			err := svc.CancelReservation(context.Background(), tc.inputID, tc.quantity)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
