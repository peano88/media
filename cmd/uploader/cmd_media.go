package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func mediaCmd() *cobra.Command {
	var (
		apiURL      string
		filePath    string
		description string
		mimeType    string
		tags        string
	)

	cmd := &cobra.Command{
		Use:   "media",
		Short: "Create media record via API (returns presigned URL)",
		Long:  `Creates a media record in the database and returns a presigned S3 upload URL.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, fileSize, sha256Hash, mimeTypeStr, err := prepareFile(filePath, mimeType)
			if err != nil {
				return err
			}
			// File is only used for metadata in this command, so we must close it
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to close file: %v\n", closeErr)
				}
			}()

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
			if err != nil {
				return fmt.Errorf("error creating media record: %w", err)
			}

			fmt.Printf("Media record created successfully\n")
			fmt.Printf("Presigned URL: %s\n", createdMedia.Data.URL)
			fmt.Printf("id of the created media record: %s\n", createdMedia.Data.ID)
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
