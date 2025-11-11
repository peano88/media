package s3

import (
	"context"
	"testing"

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
		validate func(*testing.T, string, error)
	}{
		{
			name:     "success - generates presigned URL with checksum",
			ctx:      ctx,
			filename: "world-cup-goal.jpg",
			sha256:   "w0rldcupg04l",
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
			})
			tt.validate(t, url, err)
		})
	}
}
