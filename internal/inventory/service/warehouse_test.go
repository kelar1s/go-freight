package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kelar1s/go-freight/internal/inventory/model"
	"github.com/kelar1s/go-freight/internal/inventory/service"
	"github.com/kelar1s/go-freight/internal/inventory/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errRepoExplosion = errors.New("repo explosion")

func TestWarehouseService_Create(t *testing.T) {
	type TestCase struct {
		name           string
		inputName      string
		inputLocation  string
		mockSetup      func(r *mocks.WarehouseRepo, c *mocks.Cache)
		expectedResult *model.Warehouse
		expectedError  error
	}

	tests := []TestCase{
		{
			name:          "Success with trim",
			inputName:     "  Central Moscow  ",
			inputLocation: "  Russia Moscow  ",
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				r.EXPECT().Create(mock.Anything, mock.MatchedBy(func(w *model.Warehouse) bool {
					return w.Name == "Central Moscow" && w.Location == "Russia Moscow"
				})).Run(func(ctx context.Context, w *model.Warehouse) {
					w.ID = 1
					w.CreatedAt = mockTime
				}).Return(nil).Once()
			},
			expectedResult: &model.Warehouse{ID: 1, Name: "Central Moscow", Location: "Russia Moscow", CreatedAt: mockTime},
			expectedError:  nil,
		},
		{
			name:           "Error - Empty Name",
			inputName:      "   ",
			inputLocation:  "Moscow",
			mockSetup:      func(r *mocks.WarehouseRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrEmptyWarehouseName,
		},
		{
			name:           "Error - Empty Location",
			inputName:      "Main",
			inputLocation:  "   ",
			mockSetup:      func(r *mocks.WarehouseRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrEmptyWarehouseLocation,
		},
		{
			name:          "Error - Repo Failure",
			inputName:     "Main",
			inputLocation: "Moscow",
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				r.EXPECT().Create(mock.Anything, mock.Anything).Return(errRepoExplosion).Once()
			},
			expectedResult: nil,
			expectedError:  errRepoExplosion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewWarehouseRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewWarehouseService(mockRepo, mockCache, mockTTL)

			res, err := svc.Create(context.Background(), tc.inputName, tc.inputLocation)

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

func TestWarehouseService_Get(t *testing.T) {
	type TestCase struct {
		name           string
		inputID        int64
		mockSetup      func(r *mocks.WarehouseRepo, c *mocks.Cache)
		expectedResult *model.Warehouse
		expectedError  error
	}

	tests := []TestCase{
		{
			name:    "Success - Cache Miss (goes to DB)",
			inputID: 1,
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				c.EXPECT().Get(mock.Anything, "warehouse:1", mock.Anything).Return(errors.New("cache miss")).Once()

				w := &model.Warehouse{ID: 1, Name: "Main"}
				r.EXPECT().Get(mock.Anything, int64(1)).Return(w, nil).Once()

				c.EXPECT().Set(mock.Anything, "warehouse:1", w, mockTTL).Return(nil).Once()
			},
			expectedResult: &model.Warehouse{ID: 1, Name: "Main"},
			expectedError:  nil,
		},
		{
			name:    "Success - Cache Hit (no DB call)",
			inputID: 2,
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				c.EXPECT().Get(mock.Anything, "warehouse:2", mock.Anything).RunAndReturn(func(ctx context.Context, key string, dest interface{}) error {
					w := dest.(**model.Warehouse)
					*w = &model.Warehouse{ID: 2, Name: "Cached WH"}
					return nil
				}).Once()
			},
			expectedResult: &model.Warehouse{ID: 2, Name: "Cached WH"},
			expectedError:  nil,
		},
		{
			name:           "Error - Invalid ID",
			inputID:        -5,
			mockSetup:      func(r *mocks.WarehouseRepo, c *mocks.Cache) {},
			expectedResult: nil,
			expectedError:  model.ErrInvalidWarehouseID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewWarehouseRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewWarehouseService(mockRepo, mockCache, mockTTL)

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

func TestWarehouseService_List(t *testing.T) {
	type TestCase struct {
		name           string
		mockSetup      func(r *mocks.WarehouseRepo, c *mocks.Cache)
		expectedResult []model.Warehouse
		expectedError  error
	}

	tests := []TestCase{
		{
			name: "Success",
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				r.EXPECT().List(mock.Anything).Return([]model.Warehouse{{ID: 1, Name: "Main"}}, nil).Once()
			},
			expectedResult: []model.Warehouse{{ID: 1, Name: "Main"}},
			expectedError:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewWarehouseRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewWarehouseService(mockRepo, mockCache, mockTTL)

			res, err := svc.List(context.Background())

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res)
			}
		})
	}
}

func TestWarehouseService_Update(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		inputName     string
		inputLocation string
		mockSetup     func(r *mocks.WarehouseRepo, c *mocks.Cache)
		expectedError error
	}

	tests := []TestCase{
		{
			name:          "Success with trim",
			inputID:       1,
			inputName:     " New ",
			inputLocation: " Loc ",
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				r.EXPECT().Update(mock.Anything, mock.MatchedBy(func(w *model.Warehouse) bool {
					return w.ID == 1 && w.Name == "New" && w.Location == "Loc"
				})).Return(nil).Once()

				c.EXPECT().Delete(mock.Anything, "warehouse:1").Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			inputName:     "New",
			inputLocation: "Loc",
			mockSetup:     func(r *mocks.WarehouseRepo, c *mocks.Cache) {},
			expectedError: model.ErrInvalidWarehouseID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewWarehouseRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewWarehouseService(mockRepo, mockCache, mockTTL)

			err := svc.Update(context.Background(), tc.inputID, tc.inputName, tc.inputLocation)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestWarehouseService_Delete(t *testing.T) {
	type TestCase struct {
		name          string
		inputID       int64
		mockSetup     func(r *mocks.WarehouseRepo, c *mocks.Cache)
		expectedError error
	}

	tests := []TestCase{
		{
			name:    "Success",
			inputID: 1,
			mockSetup: func(r *mocks.WarehouseRepo, c *mocks.Cache) {
				r.EXPECT().Delete(mock.Anything, int64(1)).Return(nil).Once()
				c.EXPECT().Delete(mock.Anything, "warehouse:1").Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:          "Error - Invalid ID",
			inputID:       0,
			mockSetup:     func(r *mocks.WarehouseRepo, c *mocks.Cache) {},
			expectedError: model.ErrInvalidWarehouseID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewWarehouseRepo(t)
			mockCache := mocks.NewCache(t)
			tc.mockSetup(mockRepo, mockCache)

			svc := service.NewWarehouseService(mockRepo, mockCache, mockTTL)

			err := svc.Delete(context.Background(), tc.inputID)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
