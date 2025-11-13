package getmedia

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/peano88/medias/internal/app/getmedia MediaRepository
//go:generate mockgen -destination=mocks/mock_url_generator.go -package=mocks github.com/peano88/medias/internal/app/getmedia URLGenerator

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/app/getmedia/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		id         uuid.UUID
		setupMocks func(*mocks.MockMediaRepository, *mocks.MockURLGenerator)
		validate   func(*testing.T, domain.Media, error)
	}{
		{
			name: "success - returns media with download URL",
			id:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			setupMocks: func(repo *mocks.MockMediaRepository, urlGen *mocks.MockURLGenerator) {
				media := domain.Media{
					ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Filename:    "world-cup-final.jpg",
					Description: stringPtr("Amazing goal"),
					Status:      domain.MediaStatusFinalized,
					Type:        domain.MediaTypeImage,
					MimeType:    "image/jpeg",
					Size:        2048000,
					SHA256:      "w0rldcup2023",
					Tags:        []domain.Tag{},
					CreatedAt:   time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC),
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111")).
					Return(media, nil)

				urlGen.EXPECT().
					GenerateDownloadURL(ctx, media).
					Return("https://s3.example.com/bucket/w0rldcup2023/world-cup-final.jpg?X-Amz-Signature=...", nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("11111111-1111-1111-1111-111111111111"), result.ID)
				assert.Equal(t, "world-cup-final.jpg", result.Filename)
				assert.Equal(t, domain.MediaStatusFinalized, result.Status)
				assert.Contains(t, result.URL, "X-Amz-Signature")
				assert.Contains(t, result.URL, "w0rldcup2023/world-cup-final.jpg")
			},
		},
		{
			name: "error - media not found",
			id:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupMocks: func(repo *mocks.MockMediaRepository, urlGen *mocks.MockURLGenerator) {
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("22222222-2222-2222-2222-222222222222")).
					Return(domain.Media{}, domain.NewError(domain.NotFoundCode,
						domain.WithMessage("media not found"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.NotFoundCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error finding media")
				}
			},
		},
		{
			name: "error - URL generation fails",
			id:   uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			setupMocks: func(repo *mocks.MockMediaRepository, urlGen *mocks.MockURLGenerator) {
				media := domain.Media{
					ID:       uuid.MustParse("33333333-3333-3333-3333-333333333333"),
					Filename: "tennis-serve.mp4",
					Status:   domain.MediaStatusFinalized,
					Type:     domain.MediaTypeVideo,
					MimeType: "video/mp4",
					Size:     8000000,
					SHA256:   "t3nn1ss3rv3",
					Tags:     []domain.Tag{},
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("33333333-3333-3333-3333-333333333333")).
					Return(media, nil)

				urlGen.EXPECT().
					GenerateDownloadURL(ctx, media).
					Return("", domain.NewError(domain.InternalCode,
						domain.WithMessage("S3 service unavailable"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error generating download URL")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMediaRepository(ctrl)
			urlGen := mocks.NewMockURLGenerator(ctrl)
			tt.setupMocks(repo, urlGen)

			uc := New(repo, urlGen)
			result, err := uc.Execute(ctx, tt.id)

			tt.validate(t, result, err)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
