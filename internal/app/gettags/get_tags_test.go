package gettags

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/peano88/medias/internal/app/gettags TagRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/app/gettags/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		params    domain.PaginationParams
		setupMock func(*mocks.MockTagRepository)
		validate  func(*testing.T, *domain.PaginatedResult[domain.Tag], error)
	}{
		{
			name: "success with default pagination",
			params: domain.PaginationParams{
				Limit:  0, // Should default to 50
				Offset: 0,
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				tags := []domain.Tag{
					{
						ID:          uuid.New(),
						Name:        "soccer",
						Description: stringPtr("Football matches"),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
					{
						ID:          uuid.New(),
						Name:        "basketball",
						Description: nil,
						CreatedAt:   time.Now().Add(-1 * time.Hour),
						UpdatedAt:   time.Now().Add(-1 * time.Hour),
					},
				}
				expectedParams := domain.PaginationParams{Limit: domain.DefaultLimit, Offset: 0}
				repo.EXPECT().
					FindAllTags(ctx, expectedParams).
					Return(tags, 100, nil)
			},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.NoError(t, err)
				assert.Len(t, result.Items, 2)
				assert.Equal(t, domain.DefaultLimit, result.Limit)
				assert.Equal(t, 0, result.Offset)
				assert.Equal(t, 100, result.Total)
			},
		},
		{
			name: "success with custom pagination",
			params: domain.PaginationParams{
				Limit:  10,
				Offset: 20,
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				tags := []domain.Tag{{ID: uuid.New(), Name: "soccer"}}
				expectedParams := domain.PaginationParams{Limit: 10, Offset: 20}
				repo.EXPECT().
					FindAllTags(ctx, expectedParams).
					Return(tags, 100, nil)
			},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.NoError(t, err)
				assert.Equal(t, 10, result.Limit)
				assert.Equal(t, 20, result.Offset)
				assert.Equal(t, 100, result.Total)
			},
		},
		{
			name: "success with empty result",
			params: domain.PaginationParams{
				Limit:  50,
				Offset: 0,
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				expectedParams := domain.PaginationParams{Limit: 50, Offset: 0}
				repo.EXPECT().
					FindAllTags(ctx, expectedParams).
					Return([]domain.Tag{}, 0, nil)
			},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.NoError(t, err)
				assert.Empty(t, result.Items)
				assert.Equal(t, 0, result.Total)
			},
		},
		{
			name: "validation error - negative limit",
			params: domain.PaginationParams{
				Limit:  -1,
				Offset: 0,
			},
			setupMock: func(repo *mocks.MockTagRepository) {},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "limit cannot be negative")
				}
			},
		},
		{
			name: "validation error - negative offset",
			params: domain.PaginationParams{
				Limit:  10,
				Offset: -5,
			},
			setupMock: func(repo *mocks.MockTagRepository) {},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "offset cannot be negative")
				}
			},
		},
		{
			name: "validation error - limit exceeds max",
			params: domain.PaginationParams{
				Limit:  200,
				Offset: 0,
			},
			setupMock: func(repo *mocks.MockTagRepository) {},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "cannot exceed")
				}
			},
		},
		{
			name: "repository error",
			params: domain.PaginationParams{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				expectedParams := domain.PaginationParams{Limit: 10, Offset: 0}
				repo.EXPECT().
					FindAllTags(ctx, expectedParams).
					Return(nil, 0, errors.New("database connection failed"))
			},
			validate: func(t *testing.T, result *domain.PaginatedResult[domain.Tag], err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error retrieving tag")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockTagRepository(ctrl)
			tt.setupMock(repo)

			uc := New(repo)
			result, err := uc.Execute(ctx, tt.params)

			tt.validate(t, result, err)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
