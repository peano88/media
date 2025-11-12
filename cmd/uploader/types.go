package main

type createMediaRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	MimeType    string   `json:"mime_type"`
	Size        int64    `json:"size"`
	SHA256      string   `json:"sha256"`
	Tags        []string `json:"tags,omitempty"`
}

type fullTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createMediaResponse struct {
	Data struct {
		ID          string    `json:"id"`
		Filename    string    `json:"filename"`
		Description *string   `json:"description"`
		Status      string    `json:"status"`
		URL         string    `json:"url"`
		Type        string    `json:"type"`
		MimeType    string    `json:"mime_type"`
		Size        int64     `json:"size"`
		Tags        []fullTag `json:"tags"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
	} `json:"data"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
	} `json:"error"`
}
