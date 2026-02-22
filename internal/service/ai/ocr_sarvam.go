package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type SarvamAdapter struct {
	apiKey    string
	baseURL   string
	available bool
	client    *http.Client
	logger    *slog.Logger
}

func NewSarvamAdapter(logger *slog.Logger) *SarvamAdapter {
	apiKey := os.Getenv("SARVAM_API_KEY")
	baseURL := os.Getenv("SARVAM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.sarvam.ai"
	}
	return &SarvamAdapter{
		apiKey:    apiKey,
		baseURL:   baseURL,
		available: apiKey != "",
		client:    &http.Client{Timeout: 30 * time.Second},
		logger:    logger,
	}
}

func (s *SarvamAdapter) Name() string { return "sarvam" }

func (s *SarvamAdapter) IsAvailable() bool { return s.available }

func (s *SarvamAdapter) SupportedFormats() []string {
	return []string{".png", ".jpg", ".jpeg", ".pdf", ".tiff"}
}

func (s *SarvamAdapter) ExtractText(ctx context.Context, filePath string, language string) (string, error) {
	if !s.available {
		return "", fmt.Errorf("sarvam: API key not configured")
	}

	// Read file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("sarvam: read file: %w", err)
	}

	// Build multipart request
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", filepath.Base(filePath))
	part.Write(fileBytes)
	if language != "" {
		writer.WriteField("language", language)
	}
	writer.Close()

	// POST to Sarvam Vision API
	// TODO: Verify exact endpoint from docs.sarvam.ai
	url := s.baseURL + "/api/document/ocr"
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("sarvam: create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sarvam: API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sarvam: API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response — TODO: adjust based on actual API response format
	var result struct {
		Text   string `json:"text"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("sarvam: decode response: %w", err)
	}

	return result.Text, nil
}
