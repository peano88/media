package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "uploader",
	Short: "Media upload tool for managing media files",
	Long:  `A CLI tool to upload media files to the media management service.`,
}

func main() {
	rootCmd.AddCommand(mediaCmd())
	rootCmd.AddCommand(s3Cmd())
	rootCmd.AddCommand(fullCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
