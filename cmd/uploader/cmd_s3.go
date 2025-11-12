package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func s3Cmd() *cobra.Command {
	var (
		filePath string
		url      string
	)

	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Upload file directly to S3 using presigned URL",
		Long:  `Uploads a file to S3 using a presigned URL (obtained from the API).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _, _, _, err := prepareFile(filePath, "")
			if err != nil {
				return err
			}
			// Note: file will be closed by http.Client.Do() in uploadToS3
			// since *os.File implements io.ReadCloser

			fmt.Println("Uploading file to S3...")
			if err := uploadToS3(url, file); err != nil {
				return fmt.Errorf("error uploading to S3: %w", err)
			}

			fmt.Println("Upload completed successfully!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to file (required)")
	cmd.Flags().StringVarP(&url, "url", "u", "", "Presigned S3 URL (required)")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}
