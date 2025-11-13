package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func fullCmd() *cobra.Command {
	var (
		apiURL      string
		filePath    string
		description string
		mimeType    string
		tags        string
	)

	cmd := &cobra.Command{
		Use:   "full",
		Short: "Create media record and upload to S3 (complete workflow)",
		Long:  `Creates a media record via API and uploads the file to S3 using the returned presigned URL.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, fileSize, sha256Hash, mimeTypeStr, err := prepareFile(filePath, mimeType)
			if err != nil {
				return err
			}
			// Note: file will be closed by http.Client.Do() in uploadToS3
			// since *os.File implements io.ReadCloser

			fmt.Printf("File prepared: size=%d bytes, sha256=%s, mimeType=%s\n", fileSize, sha256Hash, mimeTypeStr)

			tagList := parseTags(tags)
			var desc *string
			if description != "" {
				desc = &description
			}

			reqBody := createMediaRequest{
				Title:       filepath.Base(filePath),
				Description: desc,
				MimeType:    mimeTypeStr,
				Size:        fileSize,
				SHA256:      sha256Hash,
				Tags:        tagList,
			}

			fmt.Printf("Creating media record for %s...\n", filepath.Base(filePath))
			createdMedia, err := createMediaRecord(apiURL, reqBody)
			presignedURL := createdMedia.Data.URL
			if err != nil {
				return fmt.Errorf("error creating media record: %w", err)
			}

			fmt.Printf("Media record created successfully\n")
			fmt.Printf("Presigned URL: %s\n", presignedURL)
			fmt.Printf("Id of created media: %s\n", createdMedia.Data.ID)

			// Reset file pointer for upload
			if _, err := file.Seek(0, 0); err != nil {
				return fmt.Errorf("error resetting file pointer: %w", err)
			}

			fmt.Println("Uploading file to S3...")
			if err := uploadToS3(presignedURL, file); err != nil {
				return fmt.Errorf("error uploading to S3: %w", err)
			}

			fmt.Println("Upload completed successfully!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&apiURL, "api", "a", "http://localhost:8080/api/v1/media", "API URL")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to file (required)")
	cmd.Flags().StringVarP(&description, "desc", "d", "", "Media description")
	cmd.Flags().StringVarP(&mimeType, "mime", "m", "", "MIME type (auto-detected if not provided)")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Comma-separated list of tags")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
