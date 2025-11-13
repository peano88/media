package http

//go:generate mockgen -destination=mocks/mock_media_retriever.go -package=mocks github.com/peano88/medias/internal/adapters/http MediaRetriever

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

func TestHandleGetMedia(t *testing.T) {
	tests := []struct {
		name      string
		mediaID   string
		setupMock func(*mocks.MockMediaRetriever)
		validate  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - returns media with download URL",
			mediaID: "11111111-1111-1111-1111-111111111111",
			setupMock: func(mr *mocks.MockMediaRetriever) {
				media := domain.Media{
					ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Filename:    "world-cup-final.jpg",
					Description: stringPtr("Amazing goal from the final match"),
					Status:      domain.MediaStatusFinalized,
					URL:         "https://s3.example.com/bucket/w0rldcup2023/world-cup-final.jpg?X-Amz-Signature=...",
					Type:        domain.MediaTypeImage,
					MimeType:    "image/jpeg",
					Size:        2048000,
					Tags: []domain.Tag{
						{
							ID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
							Name:      "soccer",
							CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
						},
					},
					CreatedAt: time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC),
				}
				mr.EXPECT().
					Execute(gomock.Any(), uuid.MustParse("11111111-1111-1111-1111-111111111111")).
					Return(media, nil)
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
				assert.Contains(t, response.Data.URL, "X-Amz-Signature")
				assert.Len(t, response.Data.Tags, 1)
			},
		},
		{
			name:    "error - invalid media ID format",
			mediaID: "invalid-uuid",
			setupMock: func(mr *mocks.MockMediaRetriever) {
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
			setupMock: func(mr *mocks.MockMediaRetriever) {
				mr.EXPECT().
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRetriever := mocks.NewMockMediaRetriever(ctrl)
			tt.setupMock(mockRetriever)

			handler := HandleGetMedia(mockRetriever)

			req := httptest.NewRequest(http.MethodGet, "/media/"+tt.mediaID, nil)
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
