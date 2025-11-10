package http

import (
	"encoding/json"
	"net/http"
)

func JSONIn[T any](rw http.ResponseWriter, r *http.Request) (T, error) {
	var req T

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errDetails := err.Error()
		respondWithError(rw, http.StatusBadRequest, "INVALID_REQUEST",
			"Failed to parse request body", &errDetails, nil)
		return req, err
	}
	defer func() {
		_ = r.Body.Close()
	}()
	return req, nil
}

func JSONOut[T any](rw http.ResponseWriter, statusCode int, resp T) {
	// Marshal and send response
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		// If we can't encode the response, we can't send error to client
		// as headers are already written
		//TODO
		return
	}
}
