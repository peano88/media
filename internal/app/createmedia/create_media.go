package createmedia

import (
	"context"
	"fmt"
	"strings"

	"github.com/peano88/medias/internal/domain"
)

// MediaRepository defines the repository contract for creating media
type MediaRepository interface {
	FindByFilenameAndSHA256(ctx context.Context, filename, sha256 string) (domain.Media, error)
	CreateMedia(ctx context.Context, media domain.Media, tagNames []string) (domain.Media, error)
}

// MediaSaver defines the contract for generating media URLs
type MediaSaver interface {
	GenerateUploadURL(ctx context.Context, media domain.Media) (string, error)
}

// UseCase handles creating new media records
type UseCase struct {
	mediaRepo MediaRepository
	saver     MediaSaver
}

// New creates a new CreateMedia use case
func New(mediaRepo MediaRepository, saver MediaSaver) *UseCase {
	return &UseCase{
		mediaRepo: mediaRepo,
		saver:     saver,
	}
}

// Execute creates a new media record with reserved status or returns existing one
func (uc *UseCase) Execute(ctx context.Context, input domain.Media, tagNames []string) (domain.Media, error) {
	if err := validateMedia(&input); err != nil {
		return domain.Media{}, err
	}

	existing, err := uc.mediaRepo.FindByFilenameAndSHA256(ctx, input.Filename, input.SHA256)
	if err == nil {
		// Media exists - handle based on status
		return uc.handleExistingMedia(ctx, existing, &input, tagNames)
	}

	if !domain.HasCode(err, domain.NotFoundCode) {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails("error checking for existing media"),
		)
	}

	input.Status = domain.MediaStatusReserved

	url, err := uc.saver.GenerateUploadURL(ctx, input)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails(fmt.Sprintf("failed to generate upload URL: %s", err)),
		)
	}

	createdMedia, err := uc.mediaRepo.CreateMedia(ctx, input, tagNames)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails(fmt.Sprintf("error creating media: %s", err)),
		)
	}
	createdMedia.URL = url
	createdMedia.Operation = domain.MediaOperationCreate

	return createdMedia, nil
}

func (uc *UseCase) handleExistingMedia(ctx context.Context, existing domain.Media, input *domain.Media, tagNames []string) (domain.Media, error) {
	switch existing.Status {
	case domain.MediaStatusReserved:
		// Check if tags match
		if !tagsMatch(existing.Tags, tagNames) {
			return domain.Media{}, domain.NewError(domain.ConflictCode,
				domain.WithMessage("media already exists with different tags"),
				domain.WithDetails("a reserved media file with this filename and sha256 already exists but has different tags"),
			)
		}

		// Regenerate URL for reserved media (in case previous one expired)
		url, err := uc.saver.GenerateUploadURL(ctx, existing)
		if err != nil {
			return domain.Media{}, domain.NewErrorFrom(err,
				domain.WithDetails(fmt.Sprintf("failed to generate upload URL: %s", err)),
			)
		}
		existing.URL = url
		existing.Operation = domain.MediaOperationUpdate
		return existing, nil

	case domain.MediaStatusFinalized:
		return domain.Media{}, domain.NewError(domain.ConflictCode,
			domain.WithMessage("media already exists"),
			domain.WithDetails("a finalized media file with this filename and sha256 already exists"),
		)

	case domain.MediaStatusFailed:
		return domain.Media{}, domain.NewError(domain.ConflictCode,
			domain.WithMessage("media upload previously failed"),
			domain.WithDetails("cannot retry upload with same filename and sha256"),
		)

	default:
		return domain.Media{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("unknown media status"),
			domain.WithDetails(fmt.Sprintf("unexpected status: %s", existing.Status)),
		)
	}
}

func validateMedia(media *domain.Media) error {
	// Validate filename
	if strings.TrimSpace(media.Filename) == "" {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid filename"),
			domain.WithDetails("filename cannot be empty"),
		)
	}
	if len(media.Filename) > 255 {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid filename"),
			domain.WithDetails("filename cannot exceed 255 characters"),
		)
	}

	// Validate description
	if media.Description != nil && len(*media.Description) > 1000 {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid description"),
			domain.WithDetails("description cannot exceed 1000 characters"),
		)
	}

	// Validate MIME type and derive media type
	if strings.TrimSpace(media.MimeType) == "" {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid mimeType"),
			domain.WithDetails("mimeType cannot be empty"),
		)
	}

	mediaType, err := deriveMediaType(media.MimeType)
	if err != nil {
		return err
	}
	media.Type = mediaType

	// Validate size
	if media.Size <= 0 {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid size"),
			domain.WithDetails("size must be positive"),
		)
	}

	// Validate SHA256
	if strings.TrimSpace(media.SHA256) == "" {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid sha256"),
			domain.WithDetails("sha256 cannot be empty"),
		)
	}

	return nil
}

func deriveMediaType(mimeType string) (domain.MediaType, error) {
	lower := strings.ToLower(mimeType)

	if strings.HasPrefix(lower, "image/") {
		return domain.MediaTypeImage, nil
	}
	if strings.HasPrefix(lower, "video/") {
		return domain.MediaTypeVideo, nil
	}

	return "", domain.NewError(domain.InvalidEntityCode,
		domain.WithMessage("unsupported media type"),
		domain.WithDetails(fmt.Sprintf("mimeType must be image/* or video/*, got: %s", mimeType)),
	)
}

// tagsMatch checks if the existing tags match the provided tag names
func tagsMatch(existingTags []domain.Tag, newTagNames []string) bool {
	if len(existingTags) != len(newTagNames) {
		return false
	}

	// Create a map of existing tag names for O(1) lookup
	existingTagNames := make(map[string]bool, len(existingTags))
	for _, tag := range existingTags {
		existingTagNames[tag.Name] = true
	}

	// Check if all new tag names exist in existing tags
	for _, tagName := range newTagNames {
		if !existingTagNames[tagName] {
			return false
		}
	}

	return true
}
