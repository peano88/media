package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/peano88/medias/internal/domain"
)

type MediaFinalizer interface {
	Execute(ctx context.Context, id uuid.UUID) (domain.Media, error)
}

func HandlePostFinalizeMedia(mf MediaFinalizer) func(http.ResponseWriter, *http.Request) {
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
		finalizedMedia, err := mf.Execute(r.Context(), mediaID)
		if err != nil {
			// Special case: if media object is populated despite error,
			// it means the media was updated (e.g., marked as failed)
			// Return media object with appropriate error status
			if finalizedMedia.ID != uuid.Nil {
				var domainErr *domain.Error
				if errors.As(err, &domainErr) {
					statusCode := errorCodeToHTTPCode(domainErr.Code)
					resp := buildMediaResponse(finalizedMedia)
					JSONOut(rw, statusCode, resp)
					return
				}
			}

			// Regular error handling when no media object or not a domain error
			handleExecutorError(r.Context(), rw, err)
			return
		}

		// Success case
		resp := buildMediaResponse(finalizedMedia)
		JSONOut(rw, http.StatusOK, resp)
	}
}
