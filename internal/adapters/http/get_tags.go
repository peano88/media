package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/peano88/medias/internal/domain"
)

type TagRetriever interface {
	Execute(context.Context, domain.PaginationParams) (*domain.PaginatedResult[domain.Tag], error)
}

func HandleGetTags(tr TagRetriever) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters from query string
		params := domain.PaginationParams{
			Limit:  parseIntQueryParam(r, "limit", 0),
			Offset: parseIntQueryParam(r, "offset", 0),
		}

		// Execute business logic
		result, err := tr.Execute(r.Context(), params)
		if err != nil {
			handleExecutorError(r.Context(), rw, err)
			return
		}

		// Convert domain tags to response format
		tagDataList := make([]tagData, len(result.Items))
		for i, tag := range result.Items {
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
			Pagination: paginationMetadata{
				Limit:  result.Limit,
				Offset: result.Offset,
				Total:  result.Total,
			},
		}

		JSONOut(rw, http.StatusOK, resp)
	}
}

// parseIntQueryParam parses an integer query parameter, returning defaultValue if not present or invalid
func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
