package detection

import (
	"context"
	"testing"

	"github.com/complyark/datalens/pkg/types"
)

func TestComposableDetector_MergesStrategies(t *testing.T) {
	// Use both pattern + heuristic — they should agree on "email" column
	// with email samples, boosting confidence via multi-method agreement
	detector := NewOfflineDetector()

	input := Input{
		ColumnName: "email_address",
		DataType:   "varchar",
		Samples:    []string{"user@example.com", "admin@test.org", "info@company.co"},
	}

	report, err := detector.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.IsPII {
		t.Fatal("expected IsPII=true for email column with email samples")
	}

	if len(report.Detections) == 0 {
		t.Fatal("expected at least one detection")
	}

	// The top match should be EMAIL
	if report.TopMatch == nil {
		t.Fatal("expected TopMatch to be set")
	}
	if report.TopMatch.Type != types.PIITypeEmail {
		t.Errorf("top match type: got %q, want %q", report.TopMatch.Type, types.PIITypeEmail)
	}

	// Multi-method agreement should boost confidence
	if len(report.TopMatch.Methods) < 2 {
		t.Errorf("expected multi-method detection, got %d methods", len(report.TopMatch.Methods))
	}
}

func TestComposableDetector_NonPII(t *testing.T) {
	detector := NewOfflineDetector()

	input := Input{
		ColumnName: "status",
		DataType:   "varchar",
		Samples:    []string{"active", "inactive", "pending"},
	}

	report, err := detector.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.IsPII {
		t.Error("expected IsPII=false for non-PII column")
	}
	if len(report.Detections) != 0 {
		t.Errorf("expected 0 detections, got %d", len(report.Detections))
	}
}

func TestComposableDetector_NameColumn(t *testing.T) {
	detector := NewOfflineDetector()

	input := Input{
		ColumnName: "first_name",
		DataType:   "varchar",
		Samples:    []string{"Rahul", "Priya", "Amit"},
	}

	report, err := detector.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.IsPII {
		t.Fatal("expected IsPII=true for first_name column")
	}

	// Should be detected by heuristic at minimum
	found := false
	for _, d := range report.Detections {
		if d.Type == types.PIITypeName {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected NAME detection for first_name column")
	}
}

func TestComposableDetector_AadhaarMultiMethod(t *testing.T) {
	detector := NewOfflineDetector()

	input := Input{
		ColumnName: "aadhaar_number",
		DataType:   "varchar",
		Samples:    []string{"2345 6789 0123", "9876 5432 1098"},
	}

	report, err := detector.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.IsPII {
		t.Fatal("expected IsPII=true for aadhaar")
	}

	// Both pattern (regex) and heuristic (column name) should detect it
	for _, d := range report.Detections {
		if d.Type == types.PIITypeAadhaar {
			if len(d.Methods) < 2 {
				t.Errorf("expected multi-method for aadhaar, got %d", len(d.Methods))
			}
			if d.FinalConfidence <= 0.5 {
				t.Errorf("aadhaar confidence too low: %f", d.FinalConfidence)
			}
			return
		}
	}
	t.Error("expected AADHAAR detection")
}

func TestComposableDetector_StrategyOutcomes(t *testing.T) {
	detector := NewOfflineDetector()

	input := Input{
		ColumnName: "email",
		DataType:   "varchar",
		Samples:    []string{"user@example.com"},
	}

	report, err := detector.Detect(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have outcomes for all strategies (pattern + heuristic + industry)
	if len(report.Strategies) != 3 {
		t.Errorf("expected 3 strategy outcomes, got %d", len(report.Strategies))
	}

	// Verify strategy outcomes are recorded
	for _, s := range report.Strategies {
		if s.Name == "" {
			t.Error("strategy outcome missing name")
		}
		// Duration may be zero for very fast strategies (sub-microsecond)
	}
}

func TestRouteByConfidence(t *testing.T) {
	tests := []struct {
		confidence float64
		want       ConfidenceLevel
	}{
		{0.99, ConfidenceAutoVerify},
		{0.95, ConfidenceAutoVerify},
		{0.90, ConfidenceQuickVerify},
		{0.80, ConfidenceQuickVerify},
		{0.75, ConfidenceManualReview},
		{0.50, ConfidenceManualReview},
		{0.40, ConfidenceLow},
		{0.10, ConfidenceLow},
		{0.0, ConfidenceLow},
	}

	for _, tt := range tests {
		got := RouteByConfidence(tt.confidence)
		if got != tt.want {
			t.Errorf("RouteByConfidence(%f): got %q, want %q", tt.confidence, got, tt.want)
		}
	}
}
