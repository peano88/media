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
