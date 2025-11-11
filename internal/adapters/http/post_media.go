package http

import (
	"context"
	"net/http"

	"github.com/peano88/medias/internal/domain"
)

type MediaCreator interface {
	Execute(context.Context, domain.Media, []string) (domain.Media, error)
}

func HandlePostMedia(mc MediaCreator) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {

		req, err := JSONIn[createMediaRequest](rw, r)
		if err != nil {
			return
		}

		// Map request to domain
		media := domain.Media{
			Filename:    req.Title, // title is filename per OpenAPI spec
			Description: req.Description,
			MimeType:    req.MimeType,
			Size:        req.Size,
			SHA256:      req.SHA256,
		}

		// Execute business logic
		createdMedia, err := mc.Execute(r.Context(), media, req.Tags)
		if err != nil {
			handleExecutorError(r.Context(), rw, err)
			return
		}

		// Convert domain tags to response format
		tagDataList := make([]tagData, len(createdMedia.Tags))
		for i, tag := range createdMedia.Tags {
			tagDataList[i] = tagData{
				ID:          tag.ID.String(),
				Name:        tag.Name,
				Description: tag.Description,
				CreatedAt:   tag.CreatedAt,
				UpdatedAt:   tag.UpdatedAt,
			}
		}

		// Determine HTTP status code based on media real operation
		statusCode := http.StatusCreated
		if createdMedia.Operation == domain.MediaOperationUpdate {
			statusCode = http.StatusOK
		}

		resp := createMediaResponse{
			Data: mediaData{
				ID:          createdMedia.ID.String(),
				Filename:    createdMedia.Filename,
				Description: createdMedia.Description,
				Status:      string(createdMedia.Status),
				URL:         createdMedia.URL,
				Type:        string(createdMedia.Type),
				MimeType:    createdMedia.MimeType,
				Size:        createdMedia.Size,
				Tags:        tagDataList,
				CreatedAt:   createdMedia.CreatedAt,
				UpdatedAt:   createdMedia.UpdatedAt,
			},
		}

		JSONOut(rw, statusCode, resp)
	}
}
