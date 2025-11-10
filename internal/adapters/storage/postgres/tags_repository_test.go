package postgres

import (
	"context"
	"testing"

	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Example test showing how to use the test helper with TestMain
func TestTagRepository_Example(t *testing.T) {
	// Reset database to known state using fixtures
	resetDB(t)

	// Verify the tags table exists by running a simple query
	var count int
	err := testPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM tags").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tags table: %s", err)
	}

	t.Logf("Database ready, tags table has %d rows", count)
}

func TestTagRepository_Create(t *testing.T) {
	resetDB(t)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	testData := []struct {
		name              string
		tag               domain.Tag
		ctx               context.Context
		expectedErrorCode string
	}{
		{
			name: "valid tag",
			tag: domain.Tag{
				Name: "rugby",
			},
			ctx:               ctx,
			expectedErrorCode: "",
		},
		{
			name: "cancelled context",
			tag: domain.Tag{
				Name: "volleyball",
			},
			ctx:               cancelCtx,
			expectedErrorCode: domain.InternalCode,
		},
		{
			name: "name conflict",
			tag: domain.Tag{
				Name: "football",
			},
			ctx:               ctx,
			expectedErrorCode: domain.ConflictCode,
		},
	}

	repo := NewTagRepository(testPool)
	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			created, err := repo.CreateTag(td.ctx, td.tag)
			if td.expectedErrorCode == "" {
				assert.NoError(t, err)
				assert.Equal(t, td.tag.Name, created.Name)
				assert.NotEmpty(t, created.ID)
				assert.NotEmpty(t, created.CreatedAt)
			} else {
				// all errors should be of type domain.Error
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, td.expectedErrorCode, domainErr.Code)
				}
			}
		})
	}
}

func TestTagRepository_FindAllTags(t *testing.T) {
	resetDB(t)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name              string
		ctx               context.Context
		params            domain.PaginationParams
		expectedNames     []string
		expectedTotal     int
		expectedErrorCode string
	}{
		{
			name: "returns all tags ordered by created_at asc",
			ctx:  ctx,
			params: domain.PaginationParams{
				Limit:  50,
				Offset: 0,
			},
			expectedNames:     []string{"soccer", "basketball", "football"},
			expectedTotal:     3,
			expectedErrorCode: "",
		},
		{
			name: "pagination with limit",
			ctx:  ctx,
			params: domain.PaginationParams{
				Limit:  2,
				Offset: 0,
			},
			expectedNames:     []string{"soccer", "basketball"},
			expectedTotal:     3,
			expectedErrorCode: "",
		},
		{
			name: "pagination with offset",
			ctx:  ctx,
			params: domain.PaginationParams{
				Limit:  2,
				Offset: 1,
			},
			expectedNames:     []string{"basketball", "football"},
			expectedTotal:     3,
			expectedErrorCode: "",
		},
		{
			name: "pagination beyond available data",
			ctx:  ctx,
			params: domain.PaginationParams{
				Limit:  10,
				Offset: 100,
			},
			expectedNames:     []string{},
			expectedTotal:     3,
			expectedErrorCode: "",
		},
		{
			name: "cancelled context",
			ctx:  cancelCtx,
			params: domain.PaginationParams{
				Limit:  10,
				Offset: 0,
			},
			expectedNames:     nil,
			expectedTotal:     0,
			expectedErrorCode: domain.InternalCode,
		},
	}

	repo := NewTagRepository(testPool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, total, err := repo.FindAllTags(tt.ctx, tt.params)
			if tt.expectedErrorCode == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTotal, total)
				assert.Len(t, tags, len(tt.expectedNames))
				assert.NotNil(t, tags) // Should return empty slice, not nil

				// Extract tag names for easier comparison
				var tagNames []string
				for _, tag := range tags {
					tagNames = append(tagNames, tag.Name)

					// Verify all fields are populated
					assert.NotEmpty(t, tag.ID)
					assert.NotEmpty(t, tag.Name)
					assert.NotEmpty(t, tag.CreatedAt)
					assert.NotEmpty(t, tag.UpdatedAt)
				}

				// Verify expected tags are present in order
				for i, expectedName := range tt.expectedNames {
					assert.Equal(t, expectedName, tagNames[i])
				}
			} else {
				var domainErr *domain.Error
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, domainErr.Code)
				}
			}
		})
	}
}

func TestTagRepository_FindAllTags_EmptyTable(t *testing.T) {
	resetDB(t)

	// Delete all tags to test empty result
	_, err := testPool.Exec(context.Background(), "DELETE FROM tags")
	assert.NoError(t, err)

	repo := NewTagRepository(testPool)
	params := domain.PaginationParams{Limit: 50, Offset: 0}
	tags, total, err := repo.FindAllTags(context.Background(), params)

	assert.NoError(t, err)
	assert.Empty(t, tags)
	assert.Equal(t, 0, total)
	assert.NotNil(t, tags) // Should be empty slice, not nil
}
