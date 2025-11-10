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
		url       string
		setupMock func(*mocks.MockTagRetriever)
		validate  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success with default pagination",
			url:  "/api/v1/tags",
			setupMock: func(tr *mocks.MockTagRetriever) {
				expectedParams := domain.PaginationParams{Limit: 0, Offset: 0}
				result := &domain.PaginatedResult[domain.Tag]{
					Items: []domain.Tag{
						{
							ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
							Name:        "soccer",
							Description: stringPtr("Football matches"),
							CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
							UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Total:  100,
					Limit:  domain.DefaultLimit,
					Offset: 0,
				}
				tr.EXPECT().
					Execute(gomock.Any(), expectedParams).
					Return(result, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var response getTagsResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response.Data, 1)
				assert.Equal(t, domain.DefaultLimit, response.Pagination.Limit)
				assert.Equal(t, 0, response.Pagination.Offset)
				assert.Equal(t, 100, response.Pagination.Total)
			},
		},
		{
			name: "success with custom pagination",
			url:  "/api/v1/tags?limit=10&offset=20",
			setupMock: func(tr *mocks.MockTagRetriever) {
				expectedParams := domain.PaginationParams{Limit: 10, Offset: 20}
				result := &domain.PaginatedResult[domain.Tag]{
					Items: []domain.Tag{
						{
							ID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
							Name: "basketball",
						},
					},
					Total:  100,
					Limit:  10,
					Offset: 20,
				}
				tr.EXPECT().
					Execute(gomock.Any(), expectedParams).
					Return(result, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)

				var response getTagsResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, 10, response.Pagination.Limit)
				assert.Equal(t, 20, response.Pagination.Offset)
				assert.Equal(t, 100, response.Pagination.Total)
			},
		},
		{
			name: "success with empty result",
			url:  "/api/v1/tags",
			setupMock: func(tr *mocks.MockTagRetriever) {
				expectedParams := domain.PaginationParams{Limit: 0, Offset: 0}
				result := &domain.PaginatedResult[domain.Tag]{
					Items:  []domain.Tag{},
					Total:  0,
					Limit:  domain.DefaultLimit,
					Offset: 0,
				}
				tr.EXPECT().
					Execute(gomock.Any(), expectedParams).
					Return(result, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rec.Code)

				var response getTagsResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Empty(t, response.Data)
				assert.NotNil(t, response.Data) // Should be empty array, not null
				assert.Equal(t, 0, response.Pagination.Total)
			},
		},
		{
			name: "invalid pagination parameters",
			url:  "/api/v1/tags?limit=-1",
			setupMock: func(tr *mocks.MockTagRetriever) {
				expectedParams := domain.PaginationParams{Limit: -1, Offset: 0}
				validationErr := domain.NewError(domain.InvalidEntityCode,
					domain.WithMessage("invalid pagination parameters"),
					domain.WithDetails("limit cannot be negative"),
				)
				tr.EXPECT().
					Execute(gomock.Any(), expectedParams).
					Return(nil, validationErr)
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
			name: "internal error",
			url:  "/api/v1/tags",
			setupMock: func(tr *mocks.MockTagRetriever) {
				expectedParams := domain.PaginationParams{Limit: 0, Offset: 0}
				internalErr := domain.NewError(domain.InternalCode,
					domain.WithMessage("database error"),
					domain.WithTS(time.Now()),
				)
				tr.EXPECT().
					Execute(gomock.Any(), expectedParams).
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

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			tt.validate(t, rec)
		})
	}
}
