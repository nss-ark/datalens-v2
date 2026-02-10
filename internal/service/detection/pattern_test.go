package detection

import (
	"context"
	"testing"

	"github.com/complyark/datalens/pkg/types"
)

func TestPatternStrategy_Email(t *testing.T) {
	s := NewPatternStrategy()
	input := Input{
		ColumnName: "email",
		DataType:   "varchar",
		Samples:    []string{"user@example.com", "admin@test.org"},
	}

	results, err := s.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == types.PIITypeEmail {
			found = true
			if r.Confidence < 0.5 {
				t.Errorf("email confidence too low: %f", r.Confidence)
			}
			break
		}
	}
	if !found {
		t.Error("expected EMAIL detection, got none")
	}
}

func TestPatternStrategy_Aadhaar(t *testing.T) {
	s := NewPatternStrategy()
	input := Input{
		ColumnName: "aadhaar_number",
		DataType:   "varchar",
		Samples:    []string{"2345 6789 0123", "9876 5432 1098"},
	}

	results, err := s.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == types.PIITypeAadhaar {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected AADHAAR detection, got none")
	}
}

func TestPatternStrategy_NoMatch(t *testing.T) {
	s := NewPatternStrategy()
	input := Input{
		ColumnName: "status",
		DataType:   "varchar",
		Samples:    []string{"active", "inactive", "pending"},
	}

	results, err := s.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for non-PII, got %d", len(results))
	}
}

func TestPatternStrategy_CreditCard(t *testing.T) {
	s := NewPatternStrategy()
	input := Input{
		ColumnName: "card_number",
		DataType:   "varchar",
		Samples:    []string{"4111 1111 1111 1111", "5500 0000 0000 0004"},
	}

	results, err := s.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == types.PIITypeCreditCard {
			found = true
			if r.Sensitivity != types.SensitivityCritical {
				t.Errorf("credit card should be CRITICAL, got %s", r.Sensitivity)
			}
			break
		}
	}
	if !found {
		t.Error("expected CREDIT_CARD detection, got none")
	}
}

func TestPatternStrategy_Metadata(t *testing.T) {
	s := NewPatternStrategy()
	if s.Name() != "pattern" {
		t.Errorf("name: got %q, want %q", s.Name(), "pattern")
	}
	if s.Method() != types.DetectionMethodRegex {
		t.Errorf("method: got %q, want %q", s.Method(), types.DetectionMethodRegex)
	}
	if s.Weight() < 0 || s.Weight() > 1.0 {
		t.Errorf("weight out of range: %f", s.Weight())
	}
}
