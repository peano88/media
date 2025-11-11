package domain

import (
	"time"

	"github.com/google/uuid"
)

// MediaType represents the type of media file
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

// MediaStatus represents the status of a media file
type MediaStatus string

const (
	MediaStatusReserved  MediaStatus = "reserved"
	MediaStatusFinalized MediaStatus = "finalized"
	MediaStatusFailed    MediaStatus = "failed"
)

type MediaOperation string

const (
	MediaOperationCreate MediaOperation = "create"
	MediaOperationUpdate MediaOperation = "update"
)

// Media represents a media file in the system
type Media struct {
	ID          uuid.UUID
	Operation   MediaOperation
	Filename    string
	Description *string
	Status      MediaStatus
	URL         string
	Type        MediaType
	MimeType    string
	Size        int64
	SHA256      string
	Tags        []Tag
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
