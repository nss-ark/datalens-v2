package shared

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

// MockStrategy for testing
type MockStrategy struct{}

func (s *MockStrategy) Name() string                  { return "mock" }
func (s *MockStrategy) Method() types.DetectionMethod { return types.DetectionMethodRegex }
func (s *MockStrategy) Weight() float64               { return 1.0 }
func (s *MockStrategy) Detect(ctx context.Context, input detection.Input) ([]detection.Result, error) {
	for _, sample := range input.Samples {
		if strings.Contains(sample, "test@example.com") {
			return []detection.Result{{
				Category:    types.PIICategoryContact,
				Type:        types.PIITypeEmail,
				Sensitivity: types.SensitivityMedium,
				Confidence:  0.9,
				Method:      types.DetectionMethodRegex,
			}}, nil
		}
	}
	return nil, nil
}

func TestFileScanner_ScanStream(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Use Mock Strategy
	detector := detection.NewComposableDetector(&MockStrategy{})

	scanner := NewFileScanner(detector, logger)

	tests := []struct {
		name     string
		content  string
		filename string
		wantPII  bool
	}{
		{
			name:     "Email in text",
			content:  "Contact me at test@example.com for more info.",
			filename: "test.txt",
			wantPII:  true,
		},
		{
			name:     "Clean text",
			content:  "Hello world, nothing to see here.",
			filename: "notes.txt",
			wantPII:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.content)
			findings, err := scanner.ScanStream(context.Background(), r, tt.filename, 1024)
			if err != nil {
				t.Fatalf("ScanStream() error = %v", err)
			}

			if tt.wantPII {
				if len(findings) == 0 {
					t.Errorf("ScanStream() expected PII findings, got none")
				}
			} else {
				if len(findings) > 0 {
					t.Errorf("ScanStream() expected no findings, got %d", len(findings))
				}
			}
		})
	}
}
