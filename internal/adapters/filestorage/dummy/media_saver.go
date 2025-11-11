package dummy

import (
	"context"
	"fmt"

	"github.com/peano88/medias/internal/domain"
)

// MediaSaver is a dummy implementation of MediaSaver interface
type MediaSaver struct {
	baseURL string
}

// NewMediaSaver creates a new dummy MediaSaver
func NewMediaSaver(baseURL string) *MediaSaver {
	return &MediaSaver{
		baseURL: baseURL,
	}
}

// GenerateUploadURL generates a dummy upload URL
func (ms *MediaSaver) GenerateUploadURL(ctx context.Context, media domain.Media) (string, error) {
	// Generate a simple URL based on filename and sha256
	// In a real implementation, this would generate a presigned S3 URL or similar
	url := fmt.Sprintf("%s/upload/%s/%s", ms.baseURL, media.SHA256, media.Filename)
	return url, nil
}
