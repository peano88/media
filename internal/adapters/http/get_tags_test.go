package http

//go:generate mockgen -destination=mocks/mock_tag_retriever.go -package=mocks github.com/peano88/medias/internal/adapters/http TagRetriever

import (
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

func TestHandleGetTags(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockTagRetriever)
		validate  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success with multiple tags",
			setupMock: func(tr *mocks.MockTagRetriever) {
				tags := []domain.Tag{
					{
						ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
						Name:        "soccer",
						Description: stringPtr("Football matches"),
						CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					},
					{
						ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
						Name:        "basketball",
						Description: nil,
						CreatedAt:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
					},
				}
				tr.EXPECT().
					Execute(gomock.Any()).
					Return(tags, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var response getTagsResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response.Data, 2)
				assert.Equal(t, "soccer", response.Data[0].Name)
				assert.Equal(t, "basketball", response.Data[1].Name)
				assert.NotNil(t, response.Data[0].Description)
				assert.Nil(t, response.Data[1].Description)
			},
		},
		{
			name: "success with empty result",
			setupMock: func(tr *mocks.MockTagRetriever) {
				tr.EXPECT().
					Execute(gomock.Any()).
					Return([]domain.Tag{}, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)

				var response getTagsResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Empty(t, response.Data)
				assert.NotNil(t, response.Data) // Should be empty array, not null
			},
		},
		{
			name: "internal error",
			setupMock: func(tr *mocks.MockTagRetriever) {
				internalErr := domain.NewError(domain.InternalCode,
					domain.WithMessage("database error"),
					domain.WithTS(time.Now()),
				)
				tr.EXPECT().
					Execute(gomock.Any()).
					Return(nil, internalErr)
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

			mockTR := mocks.NewMockTagRetriever(ctrl)
			tt.setupMock(mockTR)

			handler := HandleGetTags(mockTR)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			tt.validate(t, rec)
		})
	}
}
