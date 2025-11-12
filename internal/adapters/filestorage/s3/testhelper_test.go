package s3

import (
	"context"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	testMediaSaver *MediaSaver
	dockerPool     *dockertest.Pool
	testResource   *dockertest.Resource
	testBucketName = "test-bucket"
)

// TestMain sets up the S3 test environment once for all tests
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

	// Start S3 Ninja container
	testResource, err = dockerPool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "scireum/s3-ninja",
		Tag:          "latest",
		ExposedPorts: []string{"9000/tcp"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := testResource.GetHostPort("9000/tcp")
	endpoint := "http://" + hostAndPort

	_ = testResource.Expire(300) // Container will be killed after 5 minutes

	// Wait for S3 Ninja to be ready
	dockerPool.MaxWait = 30 * time.Second

	// Set credentials as environment variables
	if err := os.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE"); err != nil {
		log.Fatalf("can't set aws access key: %s", err)
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"); err != nil {
		log.Fatalf("can't set aws secret access key: %s", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors during tests
	}))

	cfg := Config{
		Region:       "eu-central-1",
		Endpoint:     endpoint,
		BucketName:   testBucketName,
		UploadExpiry: 3600,
	}

	// Retry creating MediaSaver until S3 Ninja is ready
	if err = dockerPool.Retry(func() error {
		ctx := context.Background()
		testMediaSaver, err = NewMediaSaver(ctx, cfg, logger)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to S3: %s", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := dockerPool.Purge(testResource); err != nil {
		log.Printf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
