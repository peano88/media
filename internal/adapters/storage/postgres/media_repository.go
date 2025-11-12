package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peano88/medias/internal/domain"
)

// MediaRepository allows interaction with the media storage in postgres
type MediaRepository struct {
	pool *pgxpool.Pool
}

// NewMediaRepository creates a new MediaRepository with the given connection pool
func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{pool: pool}
}

// FindByFilenameAndSHA256 finds a media record by filename and sha256
func (mr *MediaRepository) FindByFilenameAndSHA256(ctx context.Context, filename, sha256 string) (domain.Media, error) {
	query := `
		SELECT id, filename, description, status, type, mime_type, size, sha256, created_at, updated_at
		FROM media
		WHERE filename = $1 AND sha256 = $2
	`

	var media domain.Media
	err := mr.pool.QueryRow(ctx, query, filename, sha256).Scan(
		&media.ID,
		&media.Filename,
		&media.Description,
		&media.Status,
		&media.Type,
		&media.MimeType,
		&media.Size,
		&media.SHA256,
		&media.CreatedAt,
		&media.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Media{}, domain.NewError(domain.NotFoundCode,
				domain.WithMessage("media not found"),
				domain.WithTS(time.Now()),
			)
		}
		return domain.Media{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to find media"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}

	// Load associated tags
	tags, err := mr.loadMediaTags(ctx, media.ID)
	if err != nil {
		return domain.Media{}, err
	}
	media.Tags = tags

	return media, nil
}

// CreateMedia creates a new media record with tag associations in a transaction
func (mr *MediaRepository) CreateMedia(ctx context.Context, media domain.Media, tagNames []string) (domain.Media, error) {
	tx, err := mr.pool.Begin(ctx)
	if err != nil {
		return domain.Media{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to begin transaction"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Insert media record
	query := `
		INSERT INTO media (filename, description, status, type, mime_type, size, sha256)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, filename, description, status, type, mime_type, size, sha256, created_at, updated_at
	`

	var created domain.Media
	err = tx.QueryRow(ctx, query,
		media.Filename,
		media.Description,
		media.Status,
		media.Type,
		media.MimeType,
		media.Size,
		media.SHA256,
	).Scan(
		&created.ID,
		&created.Filename,
		&created.Description,
		&created.Status,
		&created.Type,
		&created.MimeType,
		&created.Size,
		&created.SHA256,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return domain.Media{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to create media"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}

	// Associate tags if provided
	if len(tagNames) > 0 {
		tags, err := mr.associateTags(ctx, tx, created.ID, tagNames)
		if err != nil {
			return domain.Media{}, err
		}
		created.Tags = tags
	} else {
		created.Tags = []domain.Tag{}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Media{}, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to commit transaction"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}

	return created, nil
}

// loadMediaTags loads all tags associated with a media record
func (mr *MediaRepository) loadMediaTags(ctx context.Context, mediaID uuid.UUID) ([]domain.Tag, error) {
	query := `
		SELECT t.id, t.name, t.description, t.created_at, t.updated_at
		FROM tags t
		INNER JOIN media_tags mt ON t.id = mt.tag_id
		WHERE mt.media_id = $1
		ORDER BY t.name ASC
	`

	rows, err := mr.pool.Query(ctx, query, mediaID)
	if err != nil {
		return nil, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to load media tags"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}
	defer rows.Close()

	tags, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Tag])
	if err != nil {
		return nil, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to collect tags"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}

	return tags, nil
}

// associateTags associates tags with a media record by tag names
func (mr *MediaRepository) associateTags(ctx context.Context, tx pgx.Tx, mediaID uuid.UUID, tagNames []string) ([]domain.Tag, error) {
	// Find tags by names
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM tags
		WHERE name = ANY($1)
	`

	rows, err := tx.Query(ctx, query, tagNames)
	if err != nil {
		return nil, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to find tags"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}
	defer rows.Close()

	tags, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Tag])
	if err != nil {
		return nil, domain.NewError(domain.InternalCode,
			domain.WithMessage("failed to collect tags"),
			domain.WithDetails(err.Error()),
			domain.WithTS(time.Now()),
		)
	}

	// Check if all tags were found
	if len(tags) != len(tagNames) {
		return nil, domain.NewError(domain.InvalidEntityCode,
			domain.WithMessage("some tags not found"),
			domain.WithDetails("one or more tag names do not exist"),
			domain.WithTS(time.Now()),
		)
	}

	// Create media_tags associations
	for _, tag := range tags {
		insertQuery := `
			INSERT INTO media_tags (media_id, tag_id)
			VALUES ($1, $2)
		`
		_, err := tx.Exec(ctx, insertQuery, mediaID, tag.ID)
		if err != nil {
			return nil, domain.NewError(domain.InternalCode,
				domain.WithMessage("failed to associate tag"),
				domain.WithDetails(err.Error()),
				domain.WithTS(time.Now()),
			)
		}
	}

	return tags, nil
}
