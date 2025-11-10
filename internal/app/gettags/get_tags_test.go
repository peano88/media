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
		setupMock func(*mocks.MockTagRepository)
		validate  func(*testing.T, []domain.Tag, error)
	}{
		{
			name: "success with multiple tags",
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
				repo.EXPECT().
					FindAllTags(ctx).
					Return(tags, nil)
			},
			validate: func(t *testing.T, result []domain.Tag, err error) {
				assert.NoError(t, err)
				assert.Len(t, result, 2)
				assert.Equal(t, "soccer", result[0].Name)
				assert.Equal(t, "basketball", result[1].Name)
			},
		},
		{
			name: "success with empty result",
			setupMock: func(repo *mocks.MockTagRepository) {
				repo.EXPECT().
					FindAllTags(ctx).
					Return([]domain.Tag{}, nil)
			},
			validate: func(t *testing.T, result []domain.Tag, err error) {
				assert.NoError(t, err)
				assert.Empty(t, result)
				assert.NotNil(t, result) // Should be empty slice, not nil
			},
		},
		{
			name: "repository error",
			setupMock: func(repo *mocks.MockTagRepository) {
				repo.EXPECT().
					FindAllTags(ctx).
					Return(nil, errors.New("database connection failed"))
			},
			validate: func(t *testing.T, result []domain.Tag, err error) {
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
			result, err := uc.Execute(ctx)

			tt.validate(t, result, err)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
