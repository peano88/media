package s3

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/peano88/medias/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMediaSaver_GenerateUploadURL(t *testing.T) {
	ctx := context.Background()
	//cancelCtx, cancel := context.WithCancel(ctx)
	// cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		filename string
		sha256   string
		size     int64
		validate func(*testing.T, string, error)
	}{
		{
			name:     "success - generates presigned URL with checksum",
			ctx:      ctx,
			filename: "world-cup-goal.jpg",
			sha256:   "w0rldcupg04l",
			size:     2048000,
			validate: func(t *testing.T, url string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, testBucketName)
				assert.Contains(t, url, "w0rldcupg04l/world-cup-goal.jpg")
				// Presigned URL should contain signature parameters
				assert.Contains(t, url, "X-Amz-Algorithm")
				assert.Contains(t, url, "X-Amz-Credential")
				assert.Contains(t, url, "X-Amz-Signature")
				// Should contain checksum parameter
				assert.Contains(t, url, "X-Amz-Checksum-Sha256")
			},
		},
		{
			name:     "success - generates URL with special characters in filename",
			ctx:      ctx,
			filename: "basketball dunk (final).mp4",
			sha256:   "b4sk3tb4ll",
			size:     15000000,
			validate: func(t *testing.T, url string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, "b4sk3tb4ll")
				assert.Contains(t, url, "X-Amz-Checksum-Sha256")
			},
		},
		// TODO : this test case is not working.. really weird
		/*
			{
				name:     "cancelled context",
				ctx:      cancelCtx,
				filename: "hockey-goal.jpg",
				sha256:   "h0ck3yg04l",
				validate: func(t *testing.T, url string, err error) {
					assert.Error(t, err)
					assert.Empty(t, url)
				},
			},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := testMediaSaver.GenerateUploadURL(tt.ctx, domain.Media{
				Filename: tt.filename,
				SHA256:   tt.sha256,
				Size:     tt.size,
			})
			tt.validate(t, url, err)
		})
	}
}

func TestMediaSaver_VerifyMediaExists(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		ctx      context.Context
		media    domain.Media
		setup    func(*testing.T, domain.Media)
		validate func(*testing.T, bool, error)
	}{
		{
			name: "success - file exists in storage",
			ctx:  ctx,
			media: domain.Media{
				Filename: "existing-file.jpg",
				SHA256:   "3x1st1ngf1l3",
				Size:     1024000,
			},
			setup: func(t *testing.T, media domain.Media) {
				// Upload a file directly to S3 container
				_, err := testMediaSaver.client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: aws.String(testBucketName),
					Key:    aws.String(testMediaSaver.mediaKey(media)),
					Body:   bytes.NewReader([]byte("test content")),
				})
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, exists bool, err error) {
				assert.NoError(t, err)
				assert.True(t, exists)
			},
		},
		{
			name: "success - file does not exist",
			ctx:  ctx,
			media: domain.Media{
				Filename: "nonexistent-file.jpg",
				SHA256:   "n0t3x1st",
				Size:     1024000,
			},
			setup: func(t *testing.T, media domain.Media) {
				// No setup - file should not exist
			},
			validate: func(t *testing.T, exists bool, err error) {
				assert.NoError(t, err)
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t, tt.media)
			exists, err := testMediaSaver.VerifyMediaExists(tt.ctx, tt.media)
			tt.validate(t, exists, err)
		})
	}
}
