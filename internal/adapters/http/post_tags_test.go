package http

//go:generate mockgen -destination=mocks/mock_tag_creator.go -package=mocks github.com/peano88/medias/internal/adapters/http TagCreator

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

func TestHandlePostTags(t *testing.T) {
	tests := []struct {
		name        string
		requestBody any
		setupMock   func(*mocks.MockTagCreator)
		validate    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: createTagRequest{
				Name:        "soccer",
				Description: stringPtr("Football matches"),
			},
			setupMock: func(tc *mocks.MockTagCreator) {
				input := domain.Tag{
					Name:        "soccer",
					Description: stringPtr("Football matches"),
				}
				createdTag := domain.Tag{
					ID:          uuid.New(),
					Name:        "soccer",
					Description: stringPtr("Football matches"),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				tc.EXPECT().
					Execute(gomock.Any(), input).
					Return(createdTag, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Header().Get("Location"), BasePath+"/")
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var response createTagResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "soccer", response.Data.Name)
				assert.NotEmpty(t, response.Data.ID)
			},
		},
		{
			name:        "invalid json",
			requestBody: `{"name": invalid json}`,
			setupMock:   func(tc *mocks.MockTagCreator) {},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Error.Code)
			},
		},
		{
			name: "validation error from use case",
			requestBody: createTagRequest{
				Name: "",
			},
			setupMock: func(tc *mocks.MockTagCreator) {
				input := domain.Tag{Name: ""}
				validationErr := domain.NewError(domain.InvalidEntityCode,
					domain.WithMessage("invalid name"),
					domain.WithDetails("name is mandatory"),
					domain.WithTS(time.Now()),
				)
				tc.EXPECT().
					Execute(gomock.Any(), input).
					Return(domain.Tag{}, validationErr)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.InvalidEntityCode, response.Error.Code)
				assert.Contains(t, response.Error.Message, "invalid name")
			},
		},
		{
			name: "conflict error",
			requestBody: createTagRequest{
				Name: "football",
			},
			setupMock: func(tc *mocks.MockTagCreator) {
				input := domain.Tag{Name: "football"}
				conflictErr := domain.NewError(domain.ConflictCode,
					domain.WithMessage("tag already exists"),
					domain.WithTS(time.Now()),
				)
				tc.EXPECT().
					Execute(gomock.Any(), input).
					Return(domain.Tag{}, conflictErr)
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
			requestBody: createTagRequest{
				Name: "tennis",
			},
			setupMock: func(tc *mocks.MockTagCreator) {
				input := domain.Tag{Name: "tennis"}
				internalErr := domain.NewError(domain.InternalCode,
					domain.WithMessage("database error"),
					domain.WithTS(time.Now()),
				)
				tc.EXPECT().
					Execute(gomock.Any(), input).
					Return(domain.Tag{}, internalErr)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rec.Code)

				var response errorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, domain.InternalCode, response.Error.Code)
			},
		},
		{
			name: "success with nil description",
			requestBody: createTagRequest{
				Name:        "basketball",
				Description: nil,
			},
			setupMock: func(tc *mocks.MockTagCreator) {
				input := domain.Tag{
					Name:        "basketball",
					Description: nil,
				}
				createdTag := domain.Tag{
					ID:          uuid.New(),
					Name:        "basketball",
					Description: nil,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				tc.EXPECT().
					Execute(gomock.Any(), input).
					Return(createdTag, nil)
			},
			validate: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, rec.Code)

				var response createTagResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "basketball", response.Data.Name)
				assert.Nil(t, response.Data.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTC := mocks.NewMockTagCreator(ctrl)
			tt.setupMock(mockTC)

			handler := HandlePostTags(mockTC)

			var body []byte
			var err error

			// Handle both struct and string request bodies
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler(rec, req)

			tt.validate(t, rec)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
