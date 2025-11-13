package getmedia

import (
	"context"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/domain"
)

// MediaRepository defines the repository contract for getting media
type MediaRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (domain.Media, error)
}

// URLGenerator defines the contract for generating download URLs
type URLGenerator interface {
	GenerateDownloadURL(ctx context.Context, media domain.Media) (string, error)
}

// UseCase handles retrieving media records by ID
type UseCase struct {
	mediaRepo    MediaRepository
	urlGenerator URLGenerator
}

// New creates a new GetMedia use case
func New(mediaRepo MediaRepository, urlGenerator URLGenerator) *UseCase {
	return &UseCase{
		mediaRepo:    mediaRepo,
		urlGenerator: urlGenerator,
	}
}

// Execute retrieves a media record by ID and generates a download URL
func (uc *UseCase) Execute(ctx context.Context, id uuid.UUID) (domain.Media, error) {
	media, err := uc.mediaRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails("error finding media"),
		)
	}

	// Generate presigned download URL
	downloadURL, err := uc.urlGenerator.GenerateDownloadURL(ctx, media)
	if err != nil {
		return domain.Media{}, domain.NewErrorFrom(err,
			domain.WithDetails("error generating download URL"),
		)
	}

	media.URL = downloadURL
	return media, nil
}
