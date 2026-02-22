package ai

import "context"

// OCRAdapter defines a pluggable interface for OCR providers.
type OCRAdapter interface {
	Name() string
	IsAvailable() bool
	ExtractText(ctx context.Context, filePath string, language string) (string, error)
	SupportedFormats() []string
}
