package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMediaRepository_FindByFilenameAndSHA256(t *testing.T) {
	resetDB(t)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		filename string
		sha256   string
		validate func(*testing.T, domain.Media, error)
	}{
		{
			name:     "success - finds finalized media with tags",
			ctx:      ctx,
			filename: "world-cup-final.jpg",
			sha256:   "w0rldcup2023",
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("111e1111-e11b-11d1-a111-111111111111"), result.ID)
				assert.Equal(t, "world-cup-final.jpg", result.Filename)
				assert.Equal(t, "Amazing goal from the final match", *result.Description)
				assert.Equal(t, domain.MediaStatusFinalized, result.Status)
				assert.Equal(t, domain.MediaTypeImage, result.Type)
				assert.Equal(t, "image/jpeg", result.MimeType)
				assert.Equal(t, int64(2048000), result.Size)
				assert.Equal(t, "w0rldcup2023", result.SHA256)
				assert.Len(t, result.Tags, 2) // soccer and football tags
				// Verify tag names
				tagNames := []string{result.Tags[0].Name, result.Tags[1].Name}
				assert.Contains(t, tagNames, "soccer")
				assert.Contains(t, tagNames, "football")
			},
		},
		{
			name:     "success - finds reserved media without tags",
			ctx:      ctx,
			filename: "tennis-serve.mp4",
			sha256:   "t3nn1ss3rv3",
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("222e2222-e22b-22d2-a222-222222222222"), result.ID)
				assert.Equal(t, "tennis-serve.mp4", result.Filename)
				assert.Nil(t, result.Description)
				assert.Equal(t, domain.MediaStatusReserved, result.Status)
				assert.Equal(t, domain.MediaTypeVideo, result.Type)
				assert.Len(t, result.Tags, 0)
			},
		},
		{
			name:     "success - finds failed media",
			ctx:      ctx,
			filename: "hockey-goal.jpg",
			sha256:   "h0ck3yg04l",
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, domain.MediaStatusFailed, result.Status)
			},
		},
		{
			name:     "not found - media doesn't exist",
			ctx:      ctx,
			filename: "nonexistent.jpg",
			sha256:   "n0t3x1st",
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.NotFoundCode, domainErr.Code)
				}
			},
		},
		{
			name:     "cancelled context",
			ctx:      cancelCtx,
			filename: "world-cup-final.jpg",
			sha256:   "w0rldcup2023",
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
				}
			},
		},
	}

	repo := NewMediaRepository(testPool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindByFilenameAndSHA256(tt.ctx, tt.filename, tt.sha256)
			tt.validate(t, result, err)
		})
	}
}

func TestMediaRepository_CreateMedia(t *testing.T) {
	resetDB(t)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		media    domain.Media
		tagNames []string
		validate func(*testing.T, domain.Media, error)
	}{
		{
			name: "success - create media with existing tags",
			ctx:  ctx,
			media: domain.Media{
				Filename:    "marathon-finish.jpg",
				Description: stringPtr("Crossing the finish line"),
				Status:      domain.MediaStatusReserved,
				Type:        domain.MediaTypeImage,
				MimeType:    "image/jpeg",
				Size:        3000000,
				SHA256:      "m4r4th0n",
			},
			tagNames: []string{"soccer", "basketball"},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, "marathon-finish.jpg", result.Filename)
				assert.Equal(t, "Crossing the finish line", *result.Description)
				assert.Equal(t, domain.MediaStatusReserved, result.Status)
				assert.Equal(t, domain.MediaTypeImage, result.Type)
				assert.Len(t, result.Tags, 2)
				// Verify tag names
				tagNames := []string{result.Tags[0].Name, result.Tags[1].Name}
				assert.Contains(t, tagNames, "soccer")
				assert.Contains(t, tagNames, "basketball")
			},
		},
		{
			name: "success - create media without tags",
			ctx:  ctx,
			media: domain.Media{
				Filename: "volleyball-spike.mp4",
				Status:   domain.MediaStatusReserved,
				Type:     domain.MediaTypeVideo,
				MimeType: "video/mp4",
				Size:     12000000,
				SHA256:   "v0ll3yb4ll",
			},
			tagNames: []string{},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, "volleyball-spike.mp4", result.Filename)
				assert.Nil(t, result.Description)
				assert.Len(t, result.Tags, 0)
			},
		},
		{
			name: "error - tag not found",
			ctx:  ctx,
			media: domain.Media{
				Filename: "badminton-smash.jpg",
				Status:   domain.MediaStatusReserved,
				Type:     domain.MediaTypeImage,
				MimeType: "image/jpeg",
				Size:     2000000,
				SHA256:   "b4dm1nt0n",
			},
			tagNames: []string{"nonexistent-tag"},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InvalidEntityCode, domainErr.Code)
					assert.Contains(t, domainErr.Message, "some tags not found")
				}
			},
		},
		{
			name: "cancelled context",
			ctx:  cancelCtx,
			media: domain.Media{
				Filename: "rugby-tackle.mp4",
				Status:   domain.MediaStatusReserved,
				Type:     domain.MediaTypeVideo,
				MimeType: "video/mp4",
				Size:     10000000,
				SHA256:   "rugbyt4ckl3",
			},
			tagNames: []string{},
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
				}
			},
		},
	}

	repo := NewMediaRepository(testPool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.CreateMedia(tt.ctx, tt.media, tt.tagNames)
			tt.validate(t, result, err)
		})
	}
}

func TestMediaRepository_FindByID(t *testing.T) {
	resetDB(t)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		id       uuid.UUID
		validate func(*testing.T, domain.Media, error)
	}{
		{
			name: "success - finds finalized media with tags",
			ctx:  ctx,
			id:   uuid.MustParse("111e1111-e11b-11d1-a111-111111111111"),
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("111e1111-e11b-11d1-a111-111111111111"), result.ID)
				assert.Equal(t, "world-cup-final.jpg", result.Filename)
				assert.Equal(t, "Amazing goal from the final match", *result.Description)
				assert.Equal(t, domain.MediaStatusFinalized, result.Status)
				assert.Equal(t, domain.MediaTypeImage, result.Type)
				assert.Equal(t, "image/jpeg", result.MimeType)
				assert.Equal(t, int64(2048000), result.Size)
				assert.Equal(t, "w0rldcup2023", result.SHA256)
				assert.Len(t, result.Tags, 2) // soccer and football tags
			},
		},
		{
			name: "success - finds reserved media without tags",
			ctx:  ctx,
			id:   uuid.MustParse("222e2222-e22b-22d2-a222-222222222222"),
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("222e2222-e22b-22d2-a222-222222222222"), result.ID)
				assert.Equal(t, "tennis-serve.mp4", result.Filename)
				assert.Nil(t, result.Description)
				assert.Equal(t, domain.MediaStatusReserved, result.Status)
				assert.Equal(t, domain.MediaTypeVideo, result.Type)
				assert.Len(t, result.Tags, 0)
			},
		},
		{
			name: "not found - media doesn't exist",
			ctx:  ctx,
			id:   uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.NotFoundCode, domainErr.Code)
				}
			},
		},
		{
			name: "cancelled context",
			ctx:  cancelCtx,
			id:   uuid.MustParse("111e1111-e11b-11d1-a111-111111111111"),
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
				}
			},
		},
	}

	repo := NewMediaRepository(testPool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindByID(tt.ctx, tt.id)
			tt.validate(t, result, err)
		})
	}
}

func TestMediaRepository_UpdateStatus(t *testing.T) {
	resetDB(t)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		media    domain.Media
		status   domain.MediaStatus
		validate func(*testing.T, domain.Media, error)
	}{
		{
			name: "success - update reserved to finalized",
			ctx:  ctx,
			media: domain.Media{
				ID:          uuid.MustParse("222e2222-e22b-22d2-a222-222222222222"),
				Filename:    "tennis-serve.mp4",
				Description: nil,
				Status:      domain.MediaStatusReserved,
				Type:        domain.MediaTypeVideo,
				MimeType:    "video/mp4",
				Size:        8000000,
				SHA256:      "t3nn1ss3rv3",
				Tags:        []domain.Tag{},
			},
			status: domain.MediaStatusFinalized,
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("222e2222-e22b-22d2-a222-222222222222"), result.ID)
				assert.Equal(t, "tennis-serve.mp4", result.Filename)
				assert.Equal(t, domain.MediaStatusFinalized, result.Status)
				assert.Equal(t, domain.MediaTypeVideo, result.Type)
				assert.NotEmpty(t, result.UpdatedAt)
				// Verify blueprint fields are preserved
				assert.Equal(t, "video/mp4", result.MimeType)
				assert.Equal(t, int64(8000000), result.Size)
				assert.Equal(t, "t3nn1ss3rv3", result.SHA256)
				assert.Len(t, result.Tags, 0)
			},
		},
		{
			name: "success - update reserved to failed",
			ctx:  ctx,
			media: domain.Media{
				ID:          uuid.MustParse("333e3333-e33b-33d3-a333-333333333333"),
				Filename:    "golf-putt.jpg",
				Description: nil,
				Status:      domain.MediaStatusReserved,
				Type:        domain.MediaTypeImage,
				MimeType:    "image/jpeg",
				Size:        2000000,
				SHA256:      "g0lfputt",
				Tags:        []domain.Tag{},
			},
			status: domain.MediaStatusFailed,
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.NoError(t, err)
				assert.Equal(t, uuid.MustParse("333e3333-e33b-33d3-a333-333333333333"), result.ID)
				assert.Equal(t, "golf-putt.jpg", result.Filename)
				assert.Equal(t, domain.MediaStatusFailed, result.Status)
				assert.NotEmpty(t, result.UpdatedAt)
			},
		},
		{
			name: "not found - media doesn't exist",
			ctx:  ctx,
			media: domain.Media{
				ID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			},
			status: domain.MediaStatusFinalized,
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.NotFoundCode, domainErr.Code)
				}
			},
		},
		{
			name: "cancelled context",
			ctx:  cancelCtx,
			media: domain.Media{
				ID: uuid.MustParse("222e2222-e22b-22d2-a222-222222222222"),
			},
			status: domain.MediaStatusFinalized,
			validate: func(t *testing.T, result domain.Media, err error) {
				assert.Error(t, err)
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, domain.InternalCode, domainErr.Code)
				}
			},
		},
	}

	repo := NewMediaRepository(testPool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.UpdateStatus(tt.ctx, tt.media, tt.status)
			tt.validate(t, result, err)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
