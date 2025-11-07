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
func Execute(ctx context.Context, repo TagRepository, input domain.Tag) (domain.Tag, error) {
	if err := validateInput(input); err != nil {
		return domain.Tag{}, err
	}

	input.Name = normalizeName(input.Name)

	created, err := repo.CreateTag(ctx, input)
	if err != nil {
		return domain.Tag{}, domain.NewErrorFrom(err,
			domain.WithDetails(fmt.Sprintf("error creating tag: %s", err)))
	}

	return created, nil
}
