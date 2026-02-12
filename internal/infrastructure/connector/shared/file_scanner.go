package shared

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

// FileScanner is a helper for scanning file content for PII.
type FileScanner struct {
	detector *detection.ComposableDetector
	logger   *slog.Logger
}

// NewFileScanner creates a new FileScanner.
func NewFileScanner(detector *detection.ComposableDetector, logger *slog.Logger) *FileScanner {
	return &FileScanner{
		detector: detector,
		logger:   logger.With("component", "file_scanner"),
	}
}

// ScanStream scans a file stream for PII. It samples the content and runs detection.
// It supports CSV, JSON, JSONL, and plain text.
// limit: max bytes to read for sampling (e.g. 10MB).
func (s *FileScanner) ScanStream(ctx context.Context, r io.Reader, filename string, limit int64) ([]discovery.PIIClassification, error) {
	// 1. Read sample (up to limit)
	// We read enough to detect PII. For structured files, we might need to parse.
	// For now, we use a simple approach: read a chunk and treat it as samples.

	content, err := io.ReadAll(io.LimitReader(r, limit))
	if err != nil {
		return nil, err
	}

	var samples []string
	// 2. Parse based on extension
	ext := strings.ToLower(getExtension(filename))
	switch ext {
	case ".csv":
		// TODO: Parse CSV to get columns and samples for each column
		// For now, we'll just treat lines as samples for a "content" column
		samples = splitLines(string(content))

	case ".json", ".jsonl":
		// TODO: Flatten JSON
		samples = splitLines(string(content)) // Very naive for JSON

	default:
		// Plain text or unknown
		samples = []string{string(content)}
	}

	// 3. Run Detection
	var findings []discovery.PIIClassification

	// We run detection on the "content"
	input := detection.Input{
		TableName:  filename,
		ColumnName: "content", // Generic column name
		Samples:    samples,
	}

	report, err := s.detector.Detect(ctx, input)
	if err != nil {
		return nil, err
	}

	if report.IsPII && report.TopMatch != nil {
		// We found PII in the file content
		c := discovery.PIIClassification{
			// FieldID is unknown here, caller must handle or we use dummy/generated
			EntityName:      filename,
			FieldName:       "content", // or detected type?
			Category:        report.TopMatch.Category,
			Type:            report.TopMatch.Type,
			Sensitivity:     report.TopMatch.Sensitivity,
			Confidence:      report.TopMatch.FinalConfidence,
			DetectionMethod: report.TopMatch.Methods[0],
			Status:          types.VerificationPending,
			Reasoning:       report.TopMatch.Reasoning,
		}
		findings = append(findings, c)
	}

	return findings, nil
}

func getExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[idx:]
	}
	return ""
}

func splitLines(s string) []string {
	// Simple split, trim empty
	raw := strings.Split(s, "\n")
	var out []string
	for _, l := range raw {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
