package gettags

import (
	"context"
	"fmt"

	"github.com/peano88/medias/internal/domain"
)

// TagRepository defines the repository contract for retrieving tags
type TagRepository interface {
	FindAllTags(ctx context.Context, params domain.PaginationParams) ([]domain.Tag, int, error)
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

// Execute retrieves paginated tags from the repository
func (uc *UseCase) Execute(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult[domain.Tag], error) {
	// Validate and apply defaults
	if err := validatePaginationParams(&params); err != nil {
		return nil, err
	}

	tags, total, err := uc.repo.FindAllTags(ctx, params)
	if err != nil {
		return nil, domain.NewErrorFrom(err,
			domain.WithDetails(fmt.Sprintf("error retrieving tag: %s", err)))
	}

	return &domain.PaginatedResult[domain.Tag]{
		Items:  tags,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func validatePaginationParams(params *domain.PaginationParams) error {
	// Validate offset
	if params.Offset < 0 {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid pagination parameters"),
			domain.WithDetails("offset cannot be negative"),
		)
	}

	// Validate limit
	if params.Limit < 0 {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid pagination parameters"),
			domain.WithDetails("limit cannot be negative"),
		)
	}

	// Apply default limit if not provided
	if params.Limit == 0 {
		params.Limit = domain.DefaultLimit
	}

	// Validate max limit
	if params.Limit > domain.MaxLimit {
		return domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("invalid pagination parameters"),
			domain.WithDetails(fmt.Sprintf("limit cannot exceed %d", domain.MaxLimit)),
		)
	}

	return nil
}
