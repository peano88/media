package http

import (
	"context"
	"net/http"

	"github.com/peano88/medias/internal/domain"
)

type TagRetriever interface {
	Execute(context.Context) ([]domain.Tag, error)
}

func HandleGetTags(tr TagRetriever) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Execute business logic
		tags, err := tr.Execute(r.Context())
		if err != nil {
			handleExecutorError(r.Context(), rw, err)
			return
		}

		// Convert domain tags to response format
		tagDataList := make([]tagData, len(tags))
		for i, tag := range tags {
			tagDataList[i] = tagData{
				ID:          tag.ID.String(),
				Name:        tag.Name,
				Description: tag.Description,
				CreatedAt:   tag.CreatedAt,
				UpdatedAt:   tag.UpdatedAt,
			}
		}

		resp := getTagsResponse{
			Data: tagDataList,
		}

		JSONOut(rw, http.StatusOK, resp)
	}
}
