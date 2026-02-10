package detection

import (
	"context"
	"testing"

	"github.com/complyark/datalens/pkg/types"
)

func TestHeuristicStrategy_ExactMatch(t *testing.T) {
	s := NewHeuristicStrategy()

	tests := []struct {
		columnName string
		wantType   types.PIIType
	}{
		{"email", types.PIITypeEmail},
		{"first_name", types.PIITypeName},
		{"last_name", types.PIITypeName},
		{"phone_number", types.PIITypePhone},
		{"date_of_birth", types.PIITypeDOB},
		{"aadhaar_number", types.PIITypeAadhaar},
		{"pan_number", types.PIITypePAN},
		{"credit_card", types.PIITypeCreditCard},
		{"ip_address", types.PIITypeIPAddress},
		{"passport_number", types.PIITypePassport},
	}

	for _, tt := range tests {
		t.Run(tt.columnName, func(t *testing.T) {
			input := Input{ColumnName: tt.columnName, DataType: "varchar"}
			results, err := s.Detect(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) == 0 {
				t.Fatalf("expected detection for %q, got none", tt.columnName)
			}
			if results[0].Type != tt.wantType {
				t.Errorf("type: got %q, want %q", results[0].Type, tt.wantType)
			}
		})
	}
}

func TestHeuristicStrategy_CaseInsensitive(t *testing.T) {
	s := NewHeuristicStrategy()
	input := Input{ColumnName: "Email_Address", DataType: "varchar"}
	results, err := s.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected detection for 'Email_Address', got none")
	}
}

func TestHeuristicStrategy_NoMatch(t *testing.T) {
	s := NewHeuristicStrategy()
	nonPII := []string{"status", "created_at", "updated_at", "count", "total", "is_active"}

	for _, col := range nonPII {
		t.Run(col, func(t *testing.T) {
			input := Input{ColumnName: col, DataType: "varchar"}
			results, err := s.Detect(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != 0 {
				t.Errorf("expected no detection for %q, got %d", col, len(results))
			}
		})
	}
}

func TestHeuristicStrategy_Metadata(t *testing.T) {
	s := NewHeuristicStrategy()
	if s.Name() != "heuristic" {
		t.Errorf("name: got %q, want %q", s.Name(), "heuristic")
	}
	if s.Method() != types.DetectionMethodHeuristic {
		t.Errorf("method: got %q, want %q", s.Method(), types.DetectionMethodHeuristic)
	}
}
