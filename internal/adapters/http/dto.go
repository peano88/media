package http

import "time"

type createTagRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type createTagResponse struct {
	Data tagData `json:"data"`
}

type getTagsResponse struct {
	Data []tagData `json:"data"`
}

type tagData struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
