package shared

import (
	"encoding/json"
)

// ParseCredentials parses the credentials string from a DataSource.
// It handles both JSON-encoded credentials (returning a map) and a raw string
// (returning a map with a single "connection_string" key).
func ParseCredentials(creds string) (map[string]any, error) {
	if creds == "" {
		return map[string]any{}, nil
	}

	var result map[string]any
	// Try parsing as JSON first
	if err := json.Unmarshal([]byte(creds), &result); err == nil {
		return result, nil
	}

	// If not JSON, treat as raw connection string
	return map[string]any{
		"connection_string": creds,
	}, nil
}
