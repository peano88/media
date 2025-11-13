package http

//go:generate mockgen -destination=mocks/mock_media_finalizer.go -package=mocks github.com/peano88/medias/internal/adapters/http MediaFinalizer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/peano88/medias/internal/adapters/http/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlePostFinalizeMedia(t *testing.T) {
	tests := []struct {
		name      string
		mediaID   string
		setupMock func(*mocks.MockMediaFinalizer)
		validate  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - finalize media",
			mediaID: "11111111-1111-1111-1111-111111111111",
			setupMock: func(mf *mocks.MockMediaFinalizer) {
				finalizedMedia := domain.Media{
					ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Filename:    "world-cup-final.jpg",
					Description: nil,
					Status:      domain.MediaStatusFinalized,
					URL:         "http://s3.example.com/w0rldcup2023/world-cup-final.jpg",
					Type:        domain.MediaTypeImage,
					MimeType:    "image/jpeg",
					Size:        2048000,
					Tags:        []domain.Tag{},
					CreatedAt:   time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC),
				}
				mf.EXPECT().
					Execute(gomock.Any(), uuid.MustParse("11111111-1111-1111-1111-111111111111")).
					Return(finalizedMedia, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				var response mediaResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "11111111-1111-1111-1111-111111111111", response.Data.ID)
				assert.Equal(t, "world-cup-final.jpg", response.Data.Filename)
				assert.Equal(t, "finalized", response.Data.Status)
			},
		},
		{
			name:    "error - invalid media ID format",
			mediaID: "invalid-uuid",
			setupMock: func(mf *mocks.MockMediaFinalizer) {
				// No mock setup - should fail before calling use case
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
				assert.Contains(t, response.Error.Message, "Invalid media ID")
			},
		},
		{
			name:    "error - media not found",
			mediaID: "22222222-2222-2222-2222-222222222222",
			setupMock: func(mf *mocks.MockMediaFinalizer) {
				mf.EXPECT().
					Execute(gomock.Any(), uuid.MustParse("22222222-2222-2222-2222-222222222222")).
					Return(domain.Media{}, domain.NewError(domain.NotFoundCode,
						domain.WithMessage("media not found"),
					))
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rec.Code)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.NotFoundCode, response.Error.Code)
				assert.Contains(t, response.Error.Message, "media not found")
			},
		},
		{
			name:    "error - media already finalized",
			mediaID: "33333333-3333-3333-3333-333333333333",
			setupMock: func(mf *mocks.MockMediaFinalizer) {
				mf.EXPECT().
					Execute(gomock.Any(), uuid.MustParse("33333333-3333-3333-3333-333333333333")).
					Return(domain.Media{}, domain.NewError(domain.ConflictCode,
						domain.WithMessage("media already finalized"),
					))
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusConflict, rec.Code)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.ConflictCode, response.Error.Code)
				assert.Contains(t, response.Error.Message, "already finalized")
			},
		},
		{
			name:    "special case - file not found in storage (returns failed media with 422)",
			mediaID: "44444444-4444-4444-4444-444444444444",
			setupMock: func(mf *mocks.MockMediaFinalizer) {
				failedMedia := domain.Media{
					ID:          uuid.MustParse("44444444-4444-4444-4444-444444444444"),
					Filename:    "golf-putt.jpg",
					Description: nil,
					Status:      domain.MediaStatusFailed,
					URL:         "http://s3.example.com/g0lfputt/golf-putt.jpg",
					Type:        domain.MediaTypeImage,
					MimeType:    "image/jpeg",
					Size:        2000000,
					Tags:        []domain.Tag{},
					CreatedAt:   time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC),
				}
				mf.EXPECT().
					Execute(gomock.Any(), uuid.MustParse("44444444-4444-4444-4444-444444444444")).
					Return(failedMedia, domain.NewError(domain.InvalidEntityCode,
						domain.WithMessage("media file not found in file storage"),
					))
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				// Should return media object despite error
				var response mediaResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "44444444-4444-4444-4444-444444444444", response.Data.ID)
				assert.Equal(t, "golf-putt.jpg", response.Data.Filename)
				assert.Equal(t, "failed", response.Data.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFinalizer := mocks.NewMockMediaFinalizer(ctrl)
			tt.setupMock(mockFinalizer)

			handler := HandlePostFinalizeMedia(mockFinalizer)

			req := httptest.NewRequest(http.MethodPost, "/media/"+tt.mediaID+"/finalize", nil)
			rec := httptest.NewRecorder()

			// Setup chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.mediaID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(rec, req)

			tt.validate(t, rec)
		})
	}
}
