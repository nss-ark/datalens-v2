package ai

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
	"github.com/xuri/excelize/v2"
)

// ParsingService defines the interface for extracting text from files.
type ParsingService interface {
	// Parse extracts text from the file at the given path.
	// It automatically detects the file type based on extension or provided mimeType.
	Parse(ctx context.Context, filePath string, mimeType string) (string, error)
	// IsOCRAvailable returns true if the OCR engine is available.
	IsOCRAvailable() bool
}

// parsingServiceImpl implements ParsingService.
type parsingServiceImpl struct {
	logger       *slog.Logger
	ocrAvailable bool
}

// NewParsingService creates a new ParsingService.
func NewParsingService(logger *slog.Logger) ParsingService {
	s := &parsingServiceImpl{
		logger: logger.With("service", "parsing_service"),
	}

	// Check if tesseract binary is available in PATH
	if _, err := exec.LookPath("tesseract"); err == nil {
		s.ocrAvailable = true
		s.logger.Info("OCR available: tesseract binary found")
	} else {
		s.logger.Warn("OCR unavailable: tesseract binary not found in PATH")
		s.ocrAvailable = false
	}

	return s
}

// IsOCRAvailable check.
func (s *parsingServiceImpl) IsOCRAvailable() bool {
	return s.ocrAvailable
}

// Parse implements the parsing logic.
func (s *parsingServiceImpl) Parse(ctx context.Context, filePath string, mimeType string) (string, error) {
	// 1. Determine File Type
	ext := strings.ToLower(filepath.Ext(filePath))
	if mimeType == "" {
		switch ext {
		case ".pdf":
			mimeType = "application/pdf"
		case ".docx":
			mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case ".xlsx":
			mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case ".png", ".jpg", ".jpeg", ".tiff", ".bmp", ".gif":
			mimeType = "image/general"
		}
	}

	s.logger.InfoContext(ctx, "parsing file", "path", filePath, "mime", mimeType)

	var text string
	var err error

	// 2. Dispatch to specific parsers
	switch {
	case strings.Contains(mimeType, "pdf") || ext == ".pdf":
		text, err = s.parsePDF(ctx, filePath)
	case strings.Contains(mimeType, "word") || ext == ".docx":
		text, err = s.parseDocx(ctx, filePath)
	case strings.Contains(mimeType, "spreadsheet") || strings.Contains(mimeType, "excel") || ext == ".xlsx":
		text, err = s.parseXlsx(ctx, filePath)
	case strings.HasPrefix(mimeType, "image/"):
		text, err = s.parseImage(ctx, filePath)
	default:
		// Fallback: Try to read as plain text
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			return "", fmt.Errorf("unsupported file type and failed to read as text: %w", readErr)
		}
		text = string(content)
	}

	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return strings.TrimSpace(text), nil
}

// parsePDF attempts native extraction first, then falls back to OCR if text is empty.
func (s *parsingServiceImpl) parsePDF(ctx context.Context, filePath string) (string, error) {
	// Native extraction using ledongthuc/pdf
	f, r, err := pdf.Open(filePath)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to open pdf natively", "error", err)
		// Don't return error yet, maybe OCR can handle it if it's openable as image?
		// Actually pdf.Open failing usually means file access error or corrupt PDF.
		// We'll try OCR only if we can't extract text but file is valid.
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	// ledongthuc/pdf reader extraction
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		// Extract text
		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			for _, word := range row.Content {
				buf.WriteString(word.S)
				buf.WriteString(" ")
			}
			buf.WriteString("\n")
		}
	}

	text := buf.String()

	// Heuristic: If text is very short relative to pages, it might be a scanned PDF.
	// Also if text is truly empty.
	if (len(strings.TrimSpace(text)) < 50) && s.ocrAvailable {
		s.logger.InfoContext(ctx, "pdf native text empty or sparse, attempting OCR", "file", filePath)
		// Tesseract can handle PDF input if built with PDF support, but usually it expects images.
		// Standard tesseract CLI might not handle PDF directly without Ghostscript/ImageMagick piping.
		// Safest way without extra deps is: "tesseract pdf_file stdout" IF tesseract supports it.
		// If not, we might fail.
		// Using Tesseract on PDF directly often requires 'pdftotext' or converting PDF to TIFF.
		// Given we want to avoid complex deps, we might skip OCR for PDF if native fails, unless we know tesseract handles it.
		// Let's TRY it. If it fails, we fall back to what we have.

		ocrText, err := s.parseImage(ctx, filePath)
		if err == nil && len(ocrText) > len(text) {
			return ocrText, nil
		}
		// If OCR failed or returned less text, return native text (even if sparse)
	}

	return text, nil
}

func (s *parsingServiceImpl) parseDocx(ctx context.Context, filePath string) (string, error) {
	// Open the file
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	content := r.Editable()
	return content.GetContent(), nil
}

func (s *parsingServiceImpl) parseXlsx(ctx context.Context, filePath string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	// Iterate over all sheets
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to get rows from sheet", "sheet", sheet, "error", err)
			continue
		}
		for _, row := range rows {
			for _, colCell := range row {
				buf.WriteString(colCell)
				buf.WriteString(" ")
			}
			buf.WriteString("\n")
		}
	}
	return buf.String(), nil
}

func (s *parsingServiceImpl) parseImage(ctx context.Context, filePath string) (string, error) {
	if !s.ocrAvailable {
		// Just warn and return empty or placeholder
		s.logger.WarnContext(ctx, "OCR requested but unavailable", "file", filePath)
		return "[OCR Unavailable - Text could not be extracted]", nil
	}

	// Execute tesseract CLI
	// Format: tesseract <image> stdout
	cmd := exec.CommandContext(ctx, "tesseract", filePath, "stdout")

	// Sanity check filePath to prevent injection (though typically internal)
	// exec.Command avoids shell injection, but let's be safe.
	// filePath is from local FS.

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract execution failed: %w", err)
	}

	return string(output), nil
}

// Close cleans up resources
func (s *parsingServiceImpl) Close() error {
	return nil
}
