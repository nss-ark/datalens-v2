package templates

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
)

//go:embed *.json
var templateFS embed.FS

// Template defines a sector-specific mapping configuration.
type Template struct {
	Sector   string    `json:"sector"`
	Mappings []Mapping `json:"mappings"`
}

// Mapping defines a rule for matching data to a purpose.
type Mapping struct {
	TablePattern  string  `json:"table_pattern"`
	ColumnPattern string  `json:"column_pattern"`
	PurposeCode   string  `json:"purpose_code"`
	Confidence    float64 `json:"confidence"`
	Reason        string  `json:"reason"`
}

// Loader handles loading and parsing sector templates.
type Loader struct {
	templates []Template
}

// NewLoader initializes the template loader with embedded templates.
func NewLoader() (*Loader, error) {
	loader := &Loader{}
	entries, err := templateFS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := templateFS.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read template file %s: %w", entry.Name(), err)
		}

		var tmpl Template
		if err := json.Unmarshal(data, &tmpl); err != nil {
			return nil, fmt.Errorf("failed to parse template file %s: %w", entry.Name(), err)
		}

		loader.templates = append(loader.templates, tmpl)
	}

	return loader, nil
}

// FindMatch checks if a table/column pair matches any template rule.
// Returns the purpose code, confidence, reason, and true if a match is found.
func (l *Loader) FindMatch(tableName, columnName string) (string, float64, string, bool) {
	for _, tmpl := range l.templates {
		for _, mapping := range tmpl.Mappings {
			// Check table pattern
			if mapping.TablePattern != "" {
				matched, _ := regexp.MatchString(mapping.TablePattern, tableName)
				if !matched {
					continue
				}
			}

			// Check column pattern (if specified)
			if mapping.ColumnPattern != "" {
				matched, _ := regexp.MatchString(mapping.ColumnPattern, columnName)
				if !matched {
					continue
				}
			}

			// If we got here, it's a match
			return mapping.PurposeCode, mapping.Confidence, mapping.Reason, true
		}
	}

	return "", 0, "", false
}
