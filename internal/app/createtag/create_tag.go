package createtag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/peano88/medias/internal/domain"
)

// TagRepository defines the interface for tag persistence operations
type TagRepository interface {
	// CreateTag stores a new tag in the repository
	CreateTag(context.Context, domain.Tag) (domain.Tag, error)
}

// UseCase encapsulates the create tag business logic
type UseCase struct {
	repo TagRepository
}

// New creates a new create tag use case
func New(repo TagRepository) *UseCase {
	return &UseCase{repo: repo}
}

// validateInput validates the input data for creating a tag
func validateInput(input domain.Tag) error {
	// Validate name
	name := strings.TrimSpace(input.Name)
	if len(name) == 0 || len(name) > 100 {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid name"),
			domain.WithDetails("name is mandatory and should be less than 100 characters"),
			domain.WithTS(time.Now()),
		)
	}

	// Validate description if provided
	if input.Description != nil {
		if len(*input.Description) > 255 {
			return domain.NewError(domain.InvalidEntityCode,
				domain.WithMessage("invalid description"),
				domain.WithDetails("description should be less than 255 characters"),
				domain.WithTS(time.Now()),
			)
		}
	}

	return nil
}

func normalizeName(original string) string {
	return strings.ToLower(strings.TrimSpace(original))
}

// Execute creates a new tag with the given input
func (uc *UseCase) Execute(ctx context.Context, input domain.Tag) (domain.Tag, error) {
	if err := validateInput(input); err != nil {
		return domain.Tag{}, err
	}

	input.Name = normalizeName(input.Name)

	created, err := uc.repo.CreateTag(ctx, input)
	if err != nil {
		return domain.Tag{}, domain.NewErrorFrom(err,
			domain.WithDetails(fmt.Sprintf("error creating tag: %s", err)))
	}

	return created, nil
}
