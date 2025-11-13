package s3

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/peano88/medias/internal/domain"
)

type MediaSaver struct {
	client         *s3.Client
	presignClient  *s3.PresignClient
	bucketName     string
	uploadExpiry   time.Duration
	endpoint       string
	publicEndpoint string
}

// NewMediaSaver creates a new S3 media saver and ensures the bucket exists
func NewMediaSaver(ctx context.Context, cfg Config, logger *slog.Logger) (*MediaSaver, error) {
	accessKeyID, secretAccessKey := cfg.Credentials()

	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	logger.Info("initializing S3 media saver", "bucket", cfg.BucketName, "region", cfg.Region)

	if err := ensureBucketExists(ctx, client, cfg.BucketName, cfg.Region, logger); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	// Create a separate client for presigning with the public endpoint
	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}

	presignClient := client
	if publicEndpoint != cfg.Endpoint {
		// Create a new client configured with public endpoint for presigning
		presignClient = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			if publicEndpoint != "" {
				o.BaseEndpoint = aws.String(publicEndpoint)
				o.UsePathStyle = true
			}
		})
	}

	return &MediaSaver{
		client:         client,
		presignClient:  s3.NewPresignClient(presignClient),
		bucketName:     cfg.BucketName,
		uploadExpiry:   time.Duration(cfg.UploadExpiry) * time.Second,
		endpoint:       cfg.Endpoint,
		publicEndpoint: publicEndpoint,
	}, nil
}

// ensureBucketExists checks if the bucket exists and creates it if it doesn't
func ensureBucketExists(ctx context.Context, client *s3.Client, bucketName, region string, logger *slog.Logger) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err == nil {
		logger.Info("bucket already exists", "bucket", bucketName)
		return nil
	}

	logger.Info("bucket does not exist, creating", "bucket", bucketName, "region", region)

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	logger.Info("bucket created successfully", "bucket", bucketName)
	return nil
}

// mediaKey generates the S3 key for a media file
func (m *MediaSaver) mediaKey(media domain.Media) string {
	return fmt.Sprintf("%s/%s", media.SHA256, media.Filename)
}

// GenerateUploadURL generates a presigned URL for uploading a media file
func (m *MediaSaver) GenerateUploadURL(ctx context.Context, media domain.Media) (string, error) {
	key := m.mediaKey(media)

	request, err := m.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:         aws.String(m.bucketName),
		Key:            aws.String(key),
		ChecksumSHA256: aws.String(media.SHA256),
		//ContentLength:  aws.Int64(media.Size),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = m.uploadExpiry
	})

	if err != nil {
		return "", domain.NewError(
			domain.InternalCode,
			domain.WithMessage("failed to generate upload URL"),
			domain.WithDetails(err.Error()),
		)
	}

	return request.URL, nil
}

// GenerateDownloadURL generates a presigned URL for downloading a media file
func (m *MediaSaver) GenerateDownloadURL(ctx context.Context, media domain.Media) (string, error) {
	key := m.mediaKey(media)

	request, err := m.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = m.uploadExpiry // reuse same expiry duration
	})

	if err != nil {
		return "", domain.NewError(
			domain.InternalCode,
			domain.WithMessage("failed to generate download URL"),
			domain.WithDetails(err.Error()),
		)
	}

	return request.URL, nil
}

// VerifyMediaExists checks if a media file exists in S3
func (m *MediaSaver) VerifyMediaExists(ctx context.Context, media domain.Media) (bool, error) {
	key := m.mediaKey(media)

	_, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if the error is because the object doesn't exist
		var notFound *types.NotFound
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &notFound) || errors.As(err, &noSuchKey) {
			return false, nil
		}
		// Other errors (permissions, network, etc.) should be returned
		return false, domain.NewError(
			domain.InternalCode,
			domain.WithMessage("failed to verify media existence"),
			domain.WithDetails(err.Error()),
		)
	}

	return true, nil
}
