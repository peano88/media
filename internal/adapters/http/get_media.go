package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/peano88/medias/internal/domain"
)

type MediaRetriever interface {
	Execute(ctx context.Context, id uuid.UUID) (domain.Media, error)
}

func HandleGetMedia(mr MediaRetriever) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Extract media ID from URL path
		mediaIDStr := chi.URLParam(r, "id")
		if mediaIDStr == "" {
			respondWithError(rw, http.StatusBadRequest, "INVALID_REQUEST",
				"Media ID is required", nil, nil)
			return
		}

		// Parse UUID
		mediaID, err := uuid.Parse(mediaIDStr)
		if err != nil {
			errDetails := "Invalid UUID format"
			respondWithError(rw, http.StatusBadRequest, "INVALID_REQUEST",
				"Invalid media ID", &errDetails, nil)
			return
		}

		// Execute business logic
		media, err := mr.Execute(r.Context(), mediaID)
		if err != nil {
			handleExecutorError(r.Context(), rw, err)
			return
		}

		// Build response
		resp := buildMediaResponse(media)
		JSONOut(rw, http.StatusOK, resp)
	}
}
