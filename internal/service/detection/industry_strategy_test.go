package detection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/pkg/types"
)

func TestIndustryStrategy_Detect(t *testing.T) {
	strategy := NewIndustryStrategy()
	ctx := context.Background()

	tests := []struct {
		name          string
		industry      string
		samples       []string
		wantCategory  types.PIICategory
		wantType      types.PIIType
		wantMatch     bool
		minConfidence float64
	}{
		{
			name:      "No Industry Context",
			industry:  "",
			samples:   []string{"1234567890"},
			wantMatch: false,
		},
		{
			name:      "Wrong Industry Context",
			industry:  "Automotive",
			samples:   []string{"1234567890"},
			wantMatch: false,
		},
		{
			name:          "Healthcare - NPI",
			industry:      "Healthcare",
			samples:       []string{"1234567893", "9876543210"}, // Valid 10 digits
			wantMatch:     true,
			wantCategory:  types.PIICategoryProfessional,
			wantType:      types.PIITypeNationalID,
			minConfidence: 0.8,
		},
		{
			name:          "Healthcare - DEA",
			industry:      "Medical Services",
			samples:       []string{"AB1234567", "XY9876543"},
			wantMatch:     true,
			wantCategory:  types.PIICategoryProfessional,
			wantType:      types.PIITypeNationalID,
			minConfidence: 0.9,
		},
		{
			name:          "BFSI - SWIFT",
			industry:      "Banking",
			samples:       []string{"BANKUS33", "CITIUSNYXXX"},
			wantMatch:     true,
			wantCategory:  types.PIICategoryFinancial,
			wantType:      types.PIITypeBankAccount, // Strategy maps SWIFT/BIC to BankAccount type usually
			minConfidence: 0.8,
		},
		{
			name:          "BFSI - IBAN (Simplified)",
			industry:      "Finance",
			samples:       []string{"DE89370400440532013000"},
			wantMatch:     true,
			wantCategory:  types.PIICategoryFinancial,
			wantType:      types.PIITypeBankAccount,
			minConfidence: 0.9,
		},
		{
			name:          "HR - Employee ID",
			industry:      "Human Resources",
			samples:       []string{"EMP-123", "emp-4567"},
			wantMatch:     true,
			wantCategory:  types.PIICategoryProfessional,
			wantType:      types.PIITypeNationalID, // Strategy maps EMP ID to NationalID/Generic
			minConfidence: 0.8,
		},
		{
			name:      "No Match",
			industry:  "Healthcare",
			samples:   []string{"Not an NPI", "Hello World"},
			wantMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := Input{
				Industry: tc.industry,
				Samples:  tc.samples,
			}

			results, err := strategy.Detect(ctx, input)
			require.NoError(t, err)

			if !tc.wantMatch {
				assert.Empty(t, results)
				return
			}

			require.NotEmpty(t, results)
			match := results[0]
			assert.Equal(t, tc.wantCategory, match.Category, "Category mismatch")
			assert.Equal(t, tc.wantType, match.Type, "Type mismatch")
			assert.GreaterOrEqual(t, match.Confidence, tc.minConfidence, "Confidence too low")
			assert.Equal(t, types.DetectionMethodIndustry, match.Method)
		})
	}
}
