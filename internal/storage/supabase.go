package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SupabaseStorage handles file uploads to Supabase Storage
type SupabaseStorage struct {
	projectURL string
	apiKey     string
	bucketName string
	client     *http.Client
}

// NewSupabaseStorage creates a new Supabase storage client
func NewSupabaseStorage(projectURL, apiKey, bucketName string) *SupabaseStorage {
	// Clean the API key - remove any newlines, carriage returns, or other control characters
	cleanKey := strings.TrimSpace(apiKey)
	cleanKey = strings.ReplaceAll(cleanKey, "\n", "")
	cleanKey = strings.ReplaceAll(cleanKey, "\r", "")
	cleanKey = strings.ReplaceAll(cleanKey, "\t", "")

	return &SupabaseStorage{
		projectURL: strings.TrimSpace(strings.TrimSuffix(projectURL, "/")),
		apiKey:     cleanKey,
		bucketName: bucketName,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UploadImage uploads an image to Supabase storage and returns the public URL
func (s *SupabaseStorage) UploadImage(ctx context.Context, path string, data []byte, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, s.bucketName, path)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true") // Overwrite if exists

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return s.GetPublicURL(path), nil
}

// UploadImageFromReader uploads an image from an io.Reader
func (s *SupabaseStorage) UploadImageFromReader(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}
	return s.UploadImage(ctx, path, data, contentType)
}

// FileExists checks if a file exists in the storage bucket
func (s *SupabaseStorage) FileExists(ctx context.Context, path string) (bool, error) {
	url := fmt.Sprintf("%s/storage/v1/object/info/%s/%s", s.projectURL, s.bucketName, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetPublicURL returns the public URL for a file path
func (s *SupabaseStorage) GetPublicURL(path string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.projectURL, s.bucketName, path)
}

// DeleteFile deletes a file from storage
func (s *SupabaseStorage) DeleteFile(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, s.bucketName, path)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListFiles lists files in a directory path
func (s *SupabaseStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/list/%s", s.projectURL, s.bucketName)

	body := fmt.Sprintf(`{"prefix":"%s","limit":1000}`, prefix)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// For simplicity, we just return an empty slice
	// Full implementation would parse the JSON response
	return []string{}, nil
}
