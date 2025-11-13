package finalizemedia

import (
	"context"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/domain"
)

// MediaRepository defines the repository contract for finalizing media
type MediaRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (domain.Media, error)
	UpdateStatus(ctx context.Context, media domain.Media, status domain.MediaStatus) (domain.Media, error)
}

// MediaVerifier defines the contract for verifying media existence in file storage
type MediaVerifier interface {
	VerifyMediaExists(ctx context.Context, media domain.Media) (bool, error)
}

// UseCase handles finalizing media records after successful upload
type UseCase struct {
	mediaRepo MediaRepository
	verifier  MediaVerifier
}

// New creates a new FinalizeMedia use case
func New(mediaRepo MediaRepository, verifier MediaVerifier) *UseCase {
	return &UseCase{
		mediaRepo: mediaRepo,
		verifier:  verifier,
	}
}

// Execute finalizes a media record after successful upload to file storage
func (uc *UseCase) Execute(ctx context.Context, id uuid.UUID) (domain.Media, error) {
	// Find media by ID
	media, err := uc.mediaRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails("error finding media"),
		)
	}

	// Check current status
	switch media.Status {
	case domain.MediaStatusReserved:
		// OK - can finalize
	case domain.MediaStatusFinalized:
		return domain.Media{}, domain.NewError(domain.ConflictCode,
			domain.WithMessage("media already finalized"),
			domain.WithDetails("cannot finalize a media that is already finalized"),
		)
	case domain.MediaStatusFailed:
		return domain.Media{}, domain.NewError(domain.ConflictCode,
			domain.WithMessage("media upload failed"),
			domain.WithDetails("cannot finalize a media that previously failed"),
		)
	default:
		return domain.Media{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("unknown media status"),
			domain.WithDetails("unexpected media status"),
		)
	}

	// Verify file exists in file storage
	exists, err := uc.verifier.VerifyMediaExists(ctx, media)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails("error verifying media in file storage"),
		)
	}

	if !exists {
		// Mark as failed if file doesn't exist
		updatedMedia, err := uc.mediaRepo.UpdateStatus(ctx, media, domain.MediaStatusFailed)
		if err != nil {
			return domain.Media{}, domain.NewErrorFrom(err,
				domain.WithDetails("error marking media as failed"),
			)
		}
		return updatedMedia, domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("media file not found in file storage"),
			domain.WithDetails("upload was not completed or file was deleted"),
		)
	}

	// Update status to finalized
	updatedMedia, err := uc.mediaRepo.UpdateStatus(ctx, media, domain.MediaStatusFinalized)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails("error finalizing media"),
		)
	}

	return updatedMedia, nil
}
