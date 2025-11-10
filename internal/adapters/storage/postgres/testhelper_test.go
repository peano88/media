package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/peano88/medias/migrations"
	"github.com/pressly/goose/v3"
)

var (
	testDB       *sql.DB       // For migrations and fixtures
	testPool     *pgxpool.Pool // For actual tests
	fixtures     *testfixtures.Loader
	dockerPool   *dockertest.Pool
	testResource *dockertest.Resource
)

// TestMain sets up the test database once for all tests
func TestMain(m *testing.M) {
	var err error

	// Setup dockertest pool
	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = dockerPool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// Start PostgreSQL container
	testResource, err = dockerPool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=test",
			"POSTGRES_DB=testdb",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := testResource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://test:secret@%s/testdb?sslmode=disable", hostAndPort)

	_ = testResource.Expire(300) // Container will be killed after 5 minutes

	// Wait for database to be ready and create sql.DB for migrations/fixtures
	dockerPool.MaxWait = 30 * time.Second
	if err = dockerPool.Retry(func() error {
		testDB, err = sql.Open("pgx", databaseUrl)
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Create pgxpool for actual tests
	poolConfig, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalf("Could not parse pool config: %s", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	testPool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Could not create connection pool: %s", err)
	}

	// Verify pool connection
	if err = testPool.Ping(context.Background()); err != nil {
		log.Fatalf("Could not ping pool: %s", err)
	}

	// Run migrations
	goose.SetBaseFS(migrations.FS)
	if err := goose.Up(testDB, "."); err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}

	// Setup testfixtures
	fixtures, err = testfixtures.New(
		testfixtures.Database(testDB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures"),
	)
	if err != nil {
		log.Fatalf("Could not setup testfixtures: %s", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testPool.Close()
	if err := testDB.Close(); err != nil {
		log.Printf("Failed to close database: %s", err)
	}
	if err := dockerPool.Purge(testResource); err != nil {
		log.Printf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// resetDB loads fixtures to reset database state between tests
func resetDB(t *testing.T) {
	t.Helper()
	if err := fixtures.Load(); err != nil {
		t.Fatalf("Could not load fixtures: %s", err)
	}
}
