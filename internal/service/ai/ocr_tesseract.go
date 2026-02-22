package ai

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

type TesseractAdapter struct {
	available bool
	logger    *slog.Logger
}

func NewTesseractAdapter(logger *slog.Logger) *TesseractAdapter {
	available := false
	if _, err := exec.LookPath("tesseract"); err == nil {
		available = true
	}
	return &TesseractAdapter{available: available, logger: logger}
}

func (t *TesseractAdapter) Name() string { return "tesseract" }

func (t *TesseractAdapter) IsAvailable() bool { return t.available }

func (t *TesseractAdapter) SupportedFormats() []string {
	return []string{".png", ".jpg", ".jpeg", ".tiff", ".bmp", ".gif"}
}

func (t *TesseractAdapter) ExtractText(ctx context.Context, filePath string, language string) (string, error) {
	if language == "" {
		language = "eng"
	}
	cmd := exec.CommandContext(ctx, "tesseract", filePath, "stdout", "-l", language)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract execution failed: %w", err)
	}
	return string(output), nil
}
