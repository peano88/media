package s3

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

	return &MediaSaver{
		client:         client,
		presignClient:  s3.NewPresignClient(client),
		bucketName:     cfg.BucketName,
		uploadExpiry:   time.Duration(cfg.UploadExpiry) * time.Second,
		endpoint:       cfg.Endpoint,
		publicEndpoint: cfg.PublicEndpoint,
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

// GenerateUploadURL generates a presigned URL for uploading a media file
func (m *MediaSaver) GenerateUploadURL(ctx context.Context, media domain.Media) (string, error) {
	key := fmt.Sprintf("%s/%s", media.SHA256, media.Filename)

	request, err := m.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:         aws.String(m.bucketName),
		Key:            aws.String(key),
		ChecksumSHA256: aws.String(media.SHA256),
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

	url := request.URL

	// Replace internal endpoint with public endpoint if they differ
	if m.endpoint != "" && m.publicEndpoint != "" && m.endpoint != m.publicEndpoint {
		url = strings.Replace(url, m.endpoint, m.publicEndpoint, 1)
	}

	return url, nil
}
