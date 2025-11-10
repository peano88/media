package createtag

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/peano88/medias/internal/app/createtag TagRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/app/createtag/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     domain.Tag
		setupMock func(*mocks.MockTagRepository)
		validate  func(*testing.T, domain.Tag, error)
	}{
		{
			name: "success",
			input: domain.Tag{
				Name:        "Soccer",
				Description: stringPtr("Football matches"),
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				normalizedInput := domain.Tag{
					Name:        "soccer",
					Description: stringPtr("Football matches"),
				}
				returnedTag := domain.Tag{
					ID:          uuid.New(),
					Name:        "soccer",
					Description: stringPtr("Football matches"),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				repo.EXPECT().
					CreateTag(ctx, normalizedInput).
					Return(returnedTag, nil)
			},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "soccer", result.Name)
				assert.NotEmpty(t, result.ID)
			},
		},
		{
			name: "invalid name - empty",
			input: domain.Tag{
				Name: "",
			},
			setupMock: func(repo *mocks.MockTagRepository) {},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "invalid name")
				}
			},
		},
		{
			name: "invalid name - too long",
			input: domain.Tag{
				Name: string(make([]byte, 101)),
			},
			setupMock: func(repo *mocks.MockTagRepository) {},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
				}
			},
		},
		{
			name: "invalid description - too long",
			input: domain.Tag{
				Name:        "basketball",
				Description: stringPtr(string(make([]byte, 256))),
			},
			setupMock: func(repo *mocks.MockTagRepository) {},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "invalid description")
				}
			},
		},
		{
			name: "name normalization - uppercase",
			input: domain.Tag{
				Name: "BASKETBALL",
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				normalizedInput := domain.Tag{Name: "basketball"}
				returnedTag := domain.Tag{
					ID:   uuid.New(),
					Name: "basketball",
				}
				repo.EXPECT().
					CreateTag(ctx, normalizedInput).
					Return(returnedTag, nil)
			},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "basketball", result.Name)
			},
		},
		{
			name: "name normalization - with spaces",
			input: domain.Tag{
				Name: "  tennis  ",
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				normalizedInput := domain.Tag{Name: "tennis"}
				returnedTag := domain.Tag{
					ID:   uuid.New(),
					Name: "tennis",
				}
				repo.EXPECT().
					CreateTag(ctx, normalizedInput).
					Return(returnedTag, nil)
			},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "tennis", result.Name)
			},
		},
		{
			name: "repository error",
			input: domain.Tag{
				Name: "rugby",
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				normalizedInput := domain.Tag{Name: "rugby"}
				repo.EXPECT().
					CreateTag(ctx, normalizedInput).
					Return(domain.Tag{}, errors.New("database connection failed"))
			},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error creating tag")
				}
			},
		},
		{
			name: "conflict error",
			input: domain.Tag{
				Name: "football",
			},
			setupMock: func(repo *mocks.MockTagRepository) {
				normalizedInput := domain.Tag{Name: "football"}
				conflictErr := domain.NewError(domain.ConflictCode,
					domain.WithMessage("tag already exists"),
					domain.WithTS(time.Now()),
				)
				repo.EXPECT().
					CreateTag(ctx, normalizedInput).
					Return(domain.Tag{}, conflictErr)
			},
			validate: func(t *testing.T, result domain.Tag, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.ConflictCode, domainErr.Code)
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
			result, err := uc.Execute(ctx, tt.input)

			tt.validate(t, result, err)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
