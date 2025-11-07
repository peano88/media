package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a tag entity in the domain model.
// Tags are used to categorize and organize media files.
type Tag struct {
	// ID is the unique identifier for the tag (UUID format)
	ID uuid.UUID

	// Name is the tag name (max 100 characters)
	Name string

	// Description is an optional description of the tag (max 255 characters)
	Description *string

	// CreatedAt is the timestamp when the tag was created
	CreatedAt time.Time

	// UpdatedAt is the timestamp when the tag was last updated
	UpdatedAt time.Time
}
