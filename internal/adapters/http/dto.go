package http

import (
	"time"

	"github.com/peano88/medias/internal/domain"
)

type createTagRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type createTagResponse struct {
	Data tagData `json:"data"`
}

type getTagsResponse struct {
	Data       []tagData          `json:"data"`
	Pagination paginationMetadata `json:"pagination"`
}

type paginationMetadata struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type tagData struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type errorResponse struct {
	Error errorDetails `json:"error"`
}

type errorDetails struct {
	Code      string     `json:"code"`
	Message   string     `json:"message"`
	Details   *string    `json:"details,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
	RequestID *string    `json:"request_id,omitempty"`
}

type createMediaRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	MimeType    string   `json:"mime_type"`
	Size        int64    `json:"size"`
	SHA256      string   `json:"sha256"`
	Tags        []string `json:"tags,omitempty"`
}

type createMediaResponse struct {
	Data mediaData `json:"data"`
}

type mediaResponse struct {
	Data mediaData `json:"data"`
}

type mediaData struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	URL         string    `json:"url"`
	Type        string    `json:"type"`
	MimeType    string    `json:"mime_type"`
	Size        int64     `json:"size"`
	Tags        []tagData `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func buildMediaResponse(media domain.Media) mediaResponse {
	tagDataList := make([]tagData, len(media.Tags))
	for i, tag := range media.Tags {
		tagDataList[i] = tagData{
			ID:          tag.ID.String(),
			Name:        tag.Name,
			Description: tag.Description,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		}
	}

	return mediaResponse{
		Data: mediaData{
			ID:          media.ID.String(),
			Filename:    media.Filename,
			Description: media.Description,
			Status:      string(media.Status),
			URL:         media.URL,
			Type:        string(media.Type),
			MimeType:    media.MimeType,
			Size:        media.Size,
			Tags:        tagDataList,
			CreatedAt:   media.CreatedAt,
			UpdatedAt:   media.UpdatedAt,
		},
	}
}
