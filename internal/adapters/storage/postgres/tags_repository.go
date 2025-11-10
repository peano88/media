package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peano88/medias/internal/domain"
)

// TagRepository allows interaction with the tags storage in postgres
type TagRepository struct {
	pool *pgxpool.Pool
}

// NewTagRepository creates a new TagRepository with the given connection pool
func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

// CreateTag stores a new tag in the database and returns it with DB-generated fields
func (tr *TagRepository) CreateTag(ctx context.Context, tag domain.Tag) (domain.Tag, error) {
	// Insert into database and return all fields (including DB-generated ones)
	query := `
		INSERT INTO tags (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`

	var created domain.Tag
	err := tr.pool.QueryRow(ctx, query, tag.Name, tag.Description).Scan(
		&created.ID,
		&created.Name,
		&created.Description,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Tag{}, domain.NewError(domain.ConflictCode,
				domain.WithMessage("tag name already exists"),
				domain.WithDetails("a tag with this name already exists in the database"),
				domain.WithTS(time.Now()),
			)
		}

		// Generic database error (includes context cancellation, timeouts, etc.)
		return domain.Tag{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to create tag"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}

	return created, nil
}
