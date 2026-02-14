package ai

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsingService_Parse_TextFile(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewParsingService(logger)

	// Create temp file
	tmpFile, err := ioutil.TempFile("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "Hello, DataLens!"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Test
	text, err := svc.Parse(context.Background(), tmpFile.Name(), "text/plain")
	require.NoError(t, err)
	assert.Equal(t, content, text)
}

func TestParsingService_Parse_Unsupported(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewParsingService(logger)

	// Create temp file with unknown extension but text content
	tmpFile, err := ioutil.TempFile("", "test-*.xyz")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "This is text content"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Test fallback to text
	text, err := svc.Parse(context.Background(), tmpFile.Name(), "")
	require.NoError(t, err)
	assert.Equal(t, content, text)
}
