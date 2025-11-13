package finalizemedia

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/peano88/medias/internal/app/finalizemedia MediaRepository
//go:generate mockgen -destination=mocks/mock_verifier.go -package=mocks github.com/peano88/medias/internal/app/finalizemedia MediaVerifier

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/app/finalizemedia/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		id         uuid.UUID
		setupMocks func(*mocks.MockMediaRepository, *mocks.MockMediaVerifier)
		validate   func(*testing.T, domain.Media, error)
	}{
		{
			name: "success - finalize reserved media",
			id:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				existingMedia := domain.Media{
					ID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Filename: "world-cup-final.jpg",
					Status:   domain.MediaStatusReserved,
					Type:     domain.MediaTypeImage,
					MimeType: "image/jpeg",
					Size:     2048000,
					SHA256:   "w0rldcup2023",
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111")).
					Return(existingMedia, nil)

				verifier.EXPECT().
					VerifyMediaExists(ctx, existingMedia).
					Return(true, nil)

				finalizedMedia := existingMedia
				finalizedMedia.Status = domain.MediaStatusFinalized
				finalizedMedia.UpdatedAt = time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)
				repo.EXPECT().
					UpdateStatus(ctx, existingMedia, domain.MediaStatusFinalized).
					Return(finalizedMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, domain.MediaStatusFinalized, result.Status)
				assert.Equal(t, "world-cup-final.jpg", result.Filename)
			},
		},
		{
			name: "conflict error - media already finalized",
			id:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				existingMedia := domain.Media{
					ID:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					Filename: "basketball-dunk.mp4",
					Status:   domain.MediaStatusFinalized,
					Type:     domain.MediaTypeVideo,
					MimeType: "video/mp4",
					Size:     15000000,
					SHA256:   "b4sk3tb4ll",
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("22222222-2222-2222-2222-222222222222")).
					Return(existingMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.ConflictCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "already finalized")
				}
			},
		},
		{
			name: "conflict error - media upload failed",
			id:   uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				existingMedia := domain.Media{
					ID:       uuid.MustParse("33333333-3333-3333-3333-333333333333"),
					Filename: "tennis-serve.mp4",
					Status:   domain.MediaStatusFailed,
					Type:     domain.MediaTypeVideo,
					MimeType: "video/mp4",
					Size:     8000000,
					SHA256:   "t3nn1ss3rv3",
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("33333333-3333-3333-3333-333333333333")).
					Return(existingMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.ConflictCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "upload failed")
				}
			},
		},
		{
			name: "validation error - file not found in file storage (marks as failed)",
			id:   uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				existingMedia := domain.Media{
					ID:       uuid.MustParse("44444444-4444-4444-4444-444444444444"),
					Filename: "golf-putt.jpg",
					Status:   domain.MediaStatusReserved,
					Type:     domain.MediaTypeImage,
					MimeType: "image/jpeg",
					Size:     2000000,
					SHA256:   "g0lfputt",
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("44444444-4444-4444-4444-444444444444")).
					Return(existingMedia, nil)

				verifier.EXPECT().
					VerifyMediaExists(ctx, existingMedia).
					Return(false, nil)

				failedMedia := existingMedia
				failedMedia.Status = domain.MediaStatusFailed
				failedMedia.UpdatedAt = time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)
				repo.EXPECT().
					UpdateStatus(ctx, existingMedia, domain.MediaStatusFailed).
					Return(failedMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "not found in file storage")
				}
				// Result should still contain the failed media
				assert.Equal(t, domain.MediaStatusFailed, result.Status)
				assert.Equal(t, "golf-putt.jpg", result.Filename)
			},
		},
		{
			name: "repository error - media not found",
			id:   uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("55555555-5555-5555-5555-555555555555")).
					Return(domain.Media{}, domain.NewError(domain.NotFoundCode,
						domain.WithMessage("media not found"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.NotFoundCode, domainErr.Code)
				}
			},
		},
		{
			name: "repository error - update status fails",
			id:   uuid.MustParse("66666666-6666-6666-6666-666666666666"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				existingMedia := domain.Media{
					ID:       uuid.MustParse("66666666-6666-6666-6666-666666666666"),
					Filename: "badminton-smash.jpg",
					Status:   domain.MediaStatusReserved,
					Type:     domain.MediaTypeImage,
					MimeType: "image/jpeg",
					Size:     1500000,
					SHA256:   "b4dm1nt0n",
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("66666666-6666-6666-6666-666666666666")).
					Return(existingMedia, nil)

				verifier.EXPECT().
					VerifyMediaExists(ctx, existingMedia).
					Return(true, nil)

				repo.EXPECT().
					UpdateStatus(ctx, existingMedia, domain.MediaStatusFinalized).
					Return(domain.Media{}, domain.NewError(domain.InternalCode,
						domain.WithMessage("database error"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error finalizing media")
				}
			},
		},
		{
			name: "verifier error - file storage verification fails",
			id:   uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			setupMocks: func(repo *mocks.MockMediaRepository, verifier *mocks.MockMediaVerifier) {
				existingMedia := domain.Media{
					ID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
					Filename: "hockey-goal.jpg",
					Status:   domain.MediaStatusReserved,
					Type:     domain.MediaTypeImage,
					MimeType: "image/jpeg",
					Size:     1800000,
					SHA256:   "h0ck3yg04l",
				}
				repo.EXPECT().
					FindByID(ctx, uuid.MustParse("77777777-7777-7777-7777-777777777777")).
					Return(existingMedia, nil)

				verifier.EXPECT().
					VerifyMediaExists(ctx, existingMedia).
					Return(false, domain.NewError(domain.InternalCode,
						domain.WithMessage("S3 connection error"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error verifying media in file storage")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMediaRepository(ctrl)
			verifier := mocks.NewMockMediaVerifier(ctrl)
			tt.setupMocks(repo, verifier)

			uc := New(repo, verifier)
			result, err := uc.Execute(ctx, tt.id)

			tt.validate(t, result, err)
		})
	}
}
