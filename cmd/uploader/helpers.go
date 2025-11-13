package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// prepareFile opens a file, detects MIME type, calculates SHA256, and returns file info
func prepareFile(filePath, mimeType string) (*os.File, int64, string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, "", "", fmt.Errorf("error opening file: %w", err)
	}

	// Detect MIME type if not provided
	mimeTypeStr := mimeType
	if mimeTypeStr == "" {
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			_ = file.Close()
			return nil, 0, "", "", fmt.Errorf("error reading file for mime detection: %w", err)
		}
		mimeTypeStr = http.DetectContentType(buffer[:n])

		if _, err := file.Seek(0, 0); err != nil {
			_ = file.Close()
			return nil, 0, "", "", fmt.Errorf("error resetting file pointer: %w", err)
		}
	}

	// Calculate SHA256
	fmt.Println("Calculating SHA256...")
	hash := sha256.New()
	fileSize, err := io.Copy(hash, file)
	if err != nil {
		_ = file.Close()
		return nil, 0, "", "", fmt.Errorf("error calculating hash: %w", err)
	}
	sha256Hash := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Reset file pointer for upload
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("Error resetting file pointer:", err)
		_ = file.Close()
		return nil, 0, "", "", fmt.Errorf("error resetting file pointer: %w", err)
	}

	return file, fileSize, sha256Hash, mimeTypeStr, nil
}

func parseTags(tags string) []string {
	if tags == "" {
		return nil
	}
	tagList := strings.Split(tags, ",")
	for i := range tagList {
		tagList[i] = strings.TrimSpace(tagList[i])
	}
	return tagList
}

func createMediaRecord(apiURL string, req createMediaRequest) (createMediaResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return createMediaResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return createMediaResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return createMediaResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return createMediaResponse{}, fmt.Errorf("API error (%s): %s - %s", errResp.Error.Code, errResp.Error.Message, errResp.Error.Details)
		}
		return createMediaResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var mediaResp createMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		return createMediaResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return mediaResp, nil
}

func uploadToS3(presignedURL string, file *os.File) error {
	// Get file size for Content-Length header
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	req, err := http.NewRequest("PUT", presignedURL, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// S3/MinIO requires Content-Length header
	req.ContentLength = fileInfo.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
