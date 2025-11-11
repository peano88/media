package createmedia

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/peano88/medias/internal/app/createmedia MediaRepository
//go:generate mockgen -destination=mocks/mock_saver.go -package=mocks github.com/peano88/medias/internal/app/createmedia MediaSaver

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/app/createmedia/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		input      domain.Media
		tagNames   []string
		setupMocks func(*mocks.MockMediaRepository, *mocks.MockMediaSaver)
		validate   func(*testing.T, domain.Media, error)
	}{
		{
			name: "success - create new media",
			input: domain.Media{
				Filename:    "world-cup-goal.jpg",
				Description: stringPtr("Amazing goal from World Cup final"),
				MimeType:    "image/jpeg",
				Size:        2048000,
				SHA256:      "a1b2c3d4e5f6",
			},
			tagNames: []string{"soccer", "world-cup"},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media doesn't exist
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "world-cup-goal.jpg", "a1b2c3d4e5f6").
					Return(domain.Media{}, domain.NewError(domain.NotFoundCode))

				// Generate URL - expect exact media with derived type
				expectedMedia := domain.Media{
					Filename:    "world-cup-goal.jpg",
					Description: stringPtr("Amazing goal from World Cup final"),
					MimeType:    "image/jpeg",
					Type:        domain.MediaTypeImage,
					Size:        2048000,
					SHA256:      "a1b2c3d4e5f6",
					Status:      domain.MediaStatusReserved,
				}
				saver.EXPECT().
					GenerateUploadURL(ctx, expectedMedia).
					Return("http://localhost:8080/upload/a1b2c3d4e5f6/world-cup-goal.jpg", nil)

				// Create media
				createdMedia := domain.Media{
					ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Filename:    "world-cup-goal.jpg",
					Description: stringPtr("Amazing goal from World Cup final"),
					MimeType:    "image/jpeg",
					Type:        domain.MediaTypeImage,
					Size:        2048000,
					SHA256:      "a1b2c3d4e5f6",
					Status:      domain.MediaStatusReserved,
					Tags: []domain.Tag{
						{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "soccer"},
						{ID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), Name: "world-cup"},
					},
					CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				}
				repo.EXPECT().
					CreateMedia(ctx, expectedMedia, []string{"soccer", "world-cup"}).
					Return(createdMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, domain.MediaOperationCreate, result.Operation)
				assert.Equal(t, "world-cup-goal.jpg", result.Filename)
				assert.Equal(t, domain.MediaTypeImage, result.Type)
				assert.Equal(t, domain.MediaStatusReserved, result.Status)
				assert.Equal(t, "http://localhost:8080/upload/a1b2c3d4e5f6/world-cup-goal.jpg", result.URL)
				assert.Len(t, result.Tags, 2)
			},
		},
		{
			name: "success - update existing reserved media",
			input: domain.Media{
				Filename: "basketball-dunk.mp4",
				MimeType: "video/mp4",
				Size:     15000000,
				SHA256:   "x9y8z7w6v5",
			},
			tagNames: []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media exists with reserved status
				existingMedia := domain.Media{
					ID:        uuid.MustParse("44444444-4444-4444-4444-444444444444"),
					Filename:  "basketball-dunk.mp4",
					MimeType:  "video/mp4",
					Type:      domain.MediaTypeVideo,
					Size:      15000000,
					SHA256:    "x9y8z7w6v5",
					Status:    domain.MediaStatusReserved,
					Tags:      []domain.Tag{},
					CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				}
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "basketball-dunk.mp4", "x9y8z7w6v5").
					Return(existingMedia, nil)

				// Regenerate URL
				saver.EXPECT().
					GenerateUploadURL(ctx, existingMedia).
					Return("http://localhost:8080/upload/x9y8z7w6v5/basketball-dunk.mp4", nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, domain.MediaOperationUpdate, result.Operation)
				assert.Equal(t, "basketball-dunk.mp4", result.Filename)
				assert.Equal(t, domain.MediaTypeVideo, result.Type)
				assert.Equal(t, domain.MediaStatusReserved, result.Status)
				assert.Equal(t, "http://localhost:8080/upload/x9y8z7w6v5/basketball-dunk.mp4", result.URL)
			},
		},
		{
			name: "validation error - empty filename",
			input: domain.Media{
				Filename: "",
				MimeType: "image/jpeg",
				Size:     1024000,
				SHA256:   "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "filename cannot be empty")
				}
			},
		},
		{
			name: "validation error - filename too long",
			input: domain.Media{
				Filename: string(make([]byte, 256)),
				MimeType: "image/jpeg",
				Size:     1024000,
				SHA256:   "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "filename cannot exceed 255 characters")
				}
			},
		},
		{
			name: "validation error - empty mimeType",
			input: domain.Media{
				Filename: "tennis-match.jpg",
				MimeType: "",
				Size:     1024000,
				SHA256:   "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "mimeType cannot be empty")
				}
			},
		},
		{
			name: "validation error - unsupported mimeType",
			input: domain.Media{
				Filename: "rugby-tackle.pdf",
				MimeType: "application/pdf",
				Size:     1024000,
				SHA256:   "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "mimeType must be image/* or video/*")
				}
			},
		},
		{
			name: "validation error - negative size",
			input: domain.Media{
				Filename: "hockey-game.mp4",
				MimeType: "video/mp4",
				Size:     -100,
				SHA256:   "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "size must be positive")
				}
			},
		},
		{
			name: "validation error - zero size",
			input: domain.Media{
				Filename: "cricket-match.jpg",
				MimeType: "image/jpeg",
				Size:     0,
				SHA256:   "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "size must be positive")
				}
			},
		},
		{
			name: "validation error - empty sha256",
			input: domain.Media{
				Filename: "volleyball-spike.jpg",
				MimeType: "image/jpeg",
				Size:     1024000,
				SHA256:   "",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "sha256 cannot be empty")
				}
			},
		},
		{
			name: "validation error - description too long",
			input: domain.Media{
				Filename:    "golf-swing.jpg",
				Description: stringPtr(string(make([]byte, 1001))),
				MimeType:    "image/jpeg",
				Size:        1024000,
				SHA256:      "abc123",
			},
			tagNames:   []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "description cannot exceed 1000 characters")
				}
			},
		},
		{
			name: "conflict error - media already finalized",
			input: domain.Media{
				Filename: "formula1-race.mp4",
				MimeType: "video/mp4",
				Size:     20000000,
				SHA256:   "f1n4l1z3d",
			},
			tagNames: []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media exists with finalized status
				existingMedia := domain.Media{
					ID:        uuid.MustParse("55555555-5555-5555-5555-555555555555"),
					Filename:  "formula1-race.mp4",
					MimeType:  "video/mp4",
					Type:      domain.MediaTypeVideo,
					Size:      20000000,
					SHA256:    "f1n4l1z3d",
					Status:    domain.MediaStatusFinalized,
					CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 2, 15, 30, 0, 0, time.UTC),
				}
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "formula1-race.mp4", "f1n4l1z3d").
					Return(existingMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.ConflictCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "media already exists")
					assert.Contains(t, domainErr.Details, "finalized")
				}
			},
		},
		{
			name: "conflict error - media upload previously failed",
			input: domain.Media{
				Filename: "skiing-downhill.jpg",
				MimeType: "image/jpeg",
				Size:     3000000,
				SHA256:   "f41l3d",
			},
			tagNames: []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media exists with failed status
				existingMedia := domain.Media{
					ID:        uuid.MustParse("66666666-6666-6666-6666-666666666666"),
					Filename:  "skiing-downhill.jpg",
					MimeType:  "image/jpeg",
					Type:      domain.MediaTypeImage,
					Size:      3000000,
					SHA256:    "f41l3d",
					Status:    domain.MediaStatusFailed,
					CreatedAt: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				}
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "skiing-downhill.jpg", "f41l3d").
					Return(existingMedia, nil)
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.ConflictCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "upload previously failed")
					assert.Contains(t, domainErr.Details, "cannot retry")
				}
			},
		},
		{
			name: "repository error - find by filename and sha256 fails",
			input: domain.Media{
				Filename: "boxing-match.mp4",
				MimeType: "video/mp4",
				Size:     10000000,
				SHA256:   "b0x1ng",
			},
			tagNames: []string{"boxing"},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Repository returns internal error
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "boxing-match.mp4", "b0x1ng").
					Return(domain.Media{}, domain.NewError(domain.InternalCode,
						domain.WithMessage("database connection failed"),
						domain.WithDetails("connection timeout"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error checking for existing media")
				}
			},
		},
		{
			name: "repository error - create media fails",
			input: domain.Media{
				Filename:    "surfing-wave.jpg",
				Description: stringPtr("Epic wave ride"),
				MimeType:    "image/jpeg",
				Size:        4000000,
				SHA256:      "surf123",
			},
			tagNames: []string{"surfing", "ocean"},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media doesn't exist
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "surfing-wave.jpg", "surf123").
					Return(domain.Media{}, domain.NewError(domain.NotFoundCode))

				// Generate URL succeeds
				expectedMedia := domain.Media{
					Filename:    "surfing-wave.jpg",
					Description: stringPtr("Epic wave ride"),
					MimeType:    "image/jpeg",
					Type:        domain.MediaTypeImage,
					Size:        4000000,
					SHA256:      "surf123",
					Status:      domain.MediaStatusReserved,
				}
				saver.EXPECT().
					GenerateUploadURL(ctx, expectedMedia).
					Return("http://localhost:8080/upload/surf123/surfing-wave.jpg", nil)

				// Create media fails
				repo.EXPECT().
					CreateMedia(ctx, expectedMedia, []string{"surfing", "ocean"}).
					Return(domain.Media{}, domain.NewError(domain.InternalCode,
						domain.WithMessage("failed to insert into database"),
						domain.WithDetails("constraint violation"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "error creating media")
				}
			},
		},
		{
			name: "media saver error - generate upload URL fails for new media",
			input: domain.Media{
				Filename: "cycling-race.jpg",
				MimeType: "image/jpeg",
				Size:     2500000,
				SHA256:   "cycl3",
			},
			tagNames: []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media doesn't exist
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "cycling-race.jpg", "cycl3").
					Return(domain.Media{}, domain.NewError(domain.NotFoundCode))

				// Generate URL fails
				expectedMedia := domain.Media{
					Filename: "cycling-race.jpg",
					MimeType: "image/jpeg",
					Type:     domain.MediaTypeImage,
					Size:     2500000,
					SHA256:   "cycl3",
					Status:   domain.MediaStatusReserved,
				}
				saver.EXPECT().
					GenerateUploadURL(ctx, expectedMedia).
					Return("", domain.NewError(domain.InternalCode,
						domain.WithMessage("storage service unavailable"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "failed to generate upload URL")
				}
			},
		},
		{
			name: "media saver error - generate upload URL fails for existing reserved media",
			input: domain.Media{
				Filename: "marathon-run.mp4",
				MimeType: "video/mp4",
				Size:     12000000,
				SHA256:   "m4r4th0n",
			},
			tagNames: []string{},
			setupMocks: func(repo *mocks.MockMediaRepository, saver *mocks.MockMediaSaver) {
				// Media exists with reserved status
				existingMedia := domain.Media{
					ID:        uuid.MustParse("77777777-7777-7777-7777-777777777777"),
					Filename:  "marathon-run.mp4",
					MimeType:  "video/mp4",
					Type:      domain.MediaTypeVideo,
					Size:      12000000,
					SHA256:    "m4r4th0n",
					Status:    domain.MediaStatusReserved,
					CreatedAt: time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC),
				}
				repo.EXPECT().
					FindByFilenameAndSHA256(ctx, "marathon-run.mp4", "m4r4th0n").
					Return(existingMedia, nil)

				// Regenerate URL fails
				saver.EXPECT().
					GenerateUploadURL(ctx, existingMedia).
					Return("", domain.NewError(domain.InternalCode,
						domain.WithMessage("storage service timeout"),
					))
			},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
					assert.Contains(t, domainErr.Details, "failed to generate upload URL")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMediaRepository(ctrl)
			saver := mocks.NewMockMediaSaver(ctrl)
			tt.setupMocks(repo, saver)

			uc := New(repo, saver)
			result, err := uc.Execute(ctx, tt.input, tt.tagNames)

			tt.validate(t, result, err)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
