package gettags

import (
	"context"
	"fmt"

	"github.com/peano88/medias/internal/domain"
)

// TagRepository defines the repository contract for retrieving tags
type TagRepository interface {
	FindAllTags(ctx context.Context) ([]domain.Tag, error)
}

// UseCase handles retrieving all tags
type UseCase struct {
	repo TagRepository
}

// New creates a new GetTags use case
func New(repo TagRepository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// Execute retrieves all tags from the repository
func (uc *UseCase) Execute(ctx context.Context) ([]domain.Tag, error) {
	tags, err := uc.repo.FindAllTags(ctx)
	if err != nil {
		return nil, domain.NewErrorFrom(err,
			domain.WithDetails(fmt.Sprintf("error retrieving tag: %s", err)))
	}

	return tags, nil
}
