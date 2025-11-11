package http

//go:generate mockgen -destination=mocks/mock_media_creator.go -package=mocks github.com/peano88/medias/internal/adapters/http MediaCreator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/adapters/http/mocks"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlePostMedia(t *testing.T) {
	tests := []struct {
		name        string
		requestBody any
		setupMock   func(*mocks.MockMediaCreator)
		validate    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - create new media",
			requestBody: createMediaRequest{
				Title:       "penalty-kick.jpg",
				Description: stringPtr("Winning penalty kick"),
				MimeType:    "image/jpeg",
				Size:        3500000,
				SHA256:      "p3n4lty",
				Tags:        []string{"soccer", "penalty"},
			},
			setupMock: func(mc *mocks.MockMediaCreator) {
				inputMedia := domain.Media{
					Filename:    "penalty-kick.jpg",
					Description: stringPtr("Winning penalty kick"),
					MimeType:    "image/jpeg",
					Size:        3500000,
					SHA256:      "p3n4lty",
				}
				createdMedia := domain.Media{
					ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
					Operation:   domain.MediaOperationCreate,
					Filename:    "penalty-kick.jpg",
					Description: stringPtr("Winning penalty kick"),
					Status:      domain.MediaStatusReserved,
					URL:         "http://localhost:8080/upload/p3n4lty/penalty-kick.jpg",
					Type:        domain.MediaTypeImage,
					MimeType:    "image/jpeg",
					Size:        3500000,
					Tags: []domain.Tag{
						{
							ID:          uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
							Name:        "soccer",
							Description: nil,
							CreatedAt:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							UpdatedAt:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
						},
						{
							ID:          uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
							Name:        "penalty",
							Description: nil,
							CreatedAt:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							UpdatedAt:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
						},
					},
					CreatedAt: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
				}
				mc.EXPECT().
					Execute(gomock.Any(), inputMedia, []string{"soccer", "penalty"}).
					Return(createdMedia, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				var response createMediaResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "penalty-kick.jpg", response.Data.Filename)
				assert.Equal(t, "reserved", response.Data.Status)
				assert.Equal(t, "image", response.Data.Type)
				assert.Equal(t, "http://localhost:8080/upload/p3n4lty/penalty-kick.jpg", response.Data.URL)
				assert.Len(t, response.Data.Tags, 2)
			},
		},
		{
			name: "success - update existing reserved media",
			requestBody: createMediaRequest{
				Title:    "slam-dunk.mp4",
				MimeType: "video/mp4",
				Size:     18000000,
				SHA256:   "sl4md",
			},
			setupMock: func(mc *mocks.MockMediaCreator) {
				inputMedia := domain.Media{
					Filename: "slam-dunk.mp4",
					MimeType: "video/mp4",
					Size:     18000000,
					SHA256:   "sl4md",
				}
				existingMedia := domain.Media{
					ID:        uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
					Operation: domain.MediaOperationUpdate,
					Filename:  "slam-dunk.mp4",
					Status:    domain.MediaStatusReserved,
					URL:       "http://localhost:8080/upload/sl4md/slam-dunk.mp4",
					Type:      domain.MediaTypeVideo,
					MimeType:  "video/mp4",
					Size:      18000000,
					Tags:      []domain.Tag{},
					CreatedAt: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC),
				}
				mc.EXPECT().
					Execute(gomock.Any(), inputMedia, nil).
					Return(existingMedia, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)

				var response createMediaResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "slam-dunk.mp4", response.Data.Filename)
				assert.Equal(t, "reserved", response.Data.Status)
				assert.Equal(t, "video", response.Data.Type)
			},
		},
		{
			name:        "invalid json",
			requestBody: `{"title": invalid json}`,
			setupMock:   func(mc *mocks.MockMediaCreator) {},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Error.Code)
			},
		},
		{
			name: "validation error - empty filename",
			requestBody: createMediaRequest{
				Title:    "",
				MimeType: "image/jpeg",
				Size:     1000000,
				SHA256:   "abc",
			},
			setupMock: func(mc *mocks.MockMediaCreator) {
				inputMedia := domain.Media{
					Filename: "",
					MimeType: "image/jpeg",
					Size:     1000000,
					SHA256:   "abc",
				}
				validationErr := domain.NewError(domain.InvalidEntityCode,
					domain.WithMessage("invalid filename"),
					domain.WithDetails("filename cannot be empty"),
				)
				mc.EXPECT().
					Execute(gomock.Any(), inputMedia, nil).
					Return(domain.Media{}, validationErr)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.InvalidEntityCode, response.Error.Code)
			},
		},
		{
			name: "conflict error - media already finalized",
			requestBody: createMediaRequest{
				Title:    "final-score.jpg",
				MimeType: "image/jpeg",
				Size:     2000000,
				SHA256:   "f1n4l",
			},
			setupMock: func(mc *mocks.MockMediaCreator) {
				inputMedia := domain.Media{
					Filename: "final-score.jpg",
					MimeType: "image/jpeg",
					Size:     2000000,
					SHA256:   "f1n4l",
				}
				conflictErr := domain.NewError(domain.ConflictCode,
					domain.WithMessage("media already exists"),
					domain.WithDetails("a finalized media file with this filename and sha256 already exists"),
				)
				mc.EXPECT().
					Execute(gomock.Any(), inputMedia, nil).
					Return(domain.Media{}, conflictErr)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusConflict, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.ConflictCode, response.Error.Code)
			},
		},
		{
			name: "internal error",
			requestBody: createMediaRequest{
				Title:    "training-session.mp4",
				MimeType: "video/mp4",
				Size:     25000000,
				SHA256:   "tr41n",
			},
			setupMock: func(mc *mocks.MockMediaCreator) {
				inputMedia := domain.Media{
					Filename: "training-session.mp4",
					MimeType: "video/mp4",
					Size:     25000000,
					SHA256:   "tr41n",
				}
				internalErr := domain.NewError(domain.InternalCode,
					domain.WithMessage("database error"),
					domain.WithDetails("connection timeout"),
				)
				mc.EXPECT().
					Execute(gomock.Any(), inputMedia, nil).
					Return(domain.Media{}, internalErr)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.InternalCode, response.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMC := mocks.NewMockMediaCreator(ctrl)
			tt.setupMock(mockMC)

			handler := HandlePostMedia(mockMC)

			var body []byte
			var err error

			// Handle both struct and string request bodies
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/media", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler(rec, req)

			tt.validate(t, rec)
		})
	}
}
