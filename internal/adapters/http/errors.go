package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/httplog/v3"
	"github.com/peano88/medias/internal/domain"
)

func errorCodeToHTTPCode(code string) int {
	switch code {
	case domain.InvalidEntityCode:
		return http.StatusUnprocessableEntity
	case domain.InternalCode:
		return http.StatusInternalServerError
	case domain.ConflictCode:
		return http.StatusConflict
	case domain.NotFoundCode:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func handleExecutorError(ctx context.Context, rw http.ResponseWriter, err error) {
	// log the error in any case. the function returns the provided error
	_ = httplog.SetError(ctx, err)
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		respondWithError(rw, errorCodeToHTTPCode(domainErr.Code), domainErr.Code,
			domainErr.Message, &domainErr.Details, &domainErr.Timestamp)
		return
	}

	respondWithError(rw, http.StatusInternalServerError, domain.InternalCode,
		"An unexpected error occurred", nil, nil)
}

func respondWithError(rw http.ResponseWriter, statusCode int, code, message string, details *string, ts *time.Time) {

	resp := errorResponse{
		Error: errorDetails{
			Code:      code,
			Message:   message,
			Details:   details,
			Timestamp: ts,
		},
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		// write the plain error in the response
		_, _ = rw.Write([]byte(message))
	}
}
