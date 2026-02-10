package detection

import (
	"context"
	"regexp"
	"strings"

	"github.com/complyark/datalens/pkg/types"
)

// IndustryStrategy detects PII using industry-specific regex patterns.
// It is useful for high-precision detection of domain-specific identifiers
// like Healthcare IDs (NPI, DEA), Financial codes (SWIFT, IBAN), etc.
type IndustryStrategy struct {
	patterns map[string][]industryPattern
}

type industryPattern struct {
	Name        string
	Regex       *regexp.Regexp
	Type        types.PIIType
	Category    types.PIICategory
	Sensitivity types.SensitivityLevel
}

// NewIndustryStrategy creates a new IndustryStrategy with built-in pattern packs.
func NewIndustryStrategy() *IndustryStrategy {
	s := &IndustryStrategy{
		patterns: make(map[string][]industryPattern),
	}
	s.loadHealthcarePatterns()
	s.loadFinancialPatterns() // BFSI
	s.loadHRPatterns()
	return s
}

func (s *IndustryStrategy) Name() string                  { return "industry_pattern" }
func (s *IndustryStrategy) Method() types.DetectionMethod { return types.DetectionMethodIndustry }
func (s *IndustryStrategy) Weight() float64               { return 0.9 } // High confidence if matched

func (s *IndustryStrategy) Detect(ctx context.Context, input Input) ([]Result, error) {
	// 1. Check if industry is specified
	if input.Industry == "" {
		return nil, nil
	}

	// 2. Get patterns for this industry
	// We support exact match or general category (e.g., "Healthcare")
	industryKey := normalizeIndustry(input.Industry)
	patterns, ok := s.patterns[industryKey]
	if !ok {
		return nil, nil
	}

	var results []Result

	// 3. Check samples against patterns
	// Naive approach: if any sample matches, we count it.
	// Better approach: threshold (e.g., >50% of samples match)

	for _, p := range patterns {
		matchCount := 0
		sampleCount := 0

		targetSamples := input.SanitizedSamples
		if len(targetSamples) == 0 {
			targetSamples = input.Samples
		}

		if len(targetSamples) == 0 {
			continue
		}

		for _, sample := range targetSamples {
			if sample == "" {
				continue
			}
			sampleCount++
			if p.Regex.MatchString(sample) {
				matchCount++
			}
		}

		// Threshold: at least 1 match? or 20%?
		// For specific IDs like DEA/NPI, even 1 match is significant if format is complex.
		// Let's go with > 0 for now, but confidence scales with match rate.
		if matchCount > 0 {
			confidence := float64(matchCount) / float64(sampleCount)
			if confidence > 0.95 {
				confidence = 0.95 // Cap at 0.95 for regex
			}

			results = append(results, Result{
				Category:    p.Category,
				Type:        p.Type,
				Sensitivity: p.Sensitivity,
				Confidence:  confidence,
				Method:      types.DetectionMethodIndustry,
				Reasoning:   "Matched industry pattern: " + p.Name,
			})
		}
	}

	return results, nil
}

func normalizeIndustry(ind string) string {
	ind = strings.ToLower(ind)
	if strings.Contains(ind, "health") || strings.Contains(ind, "medical") {
		return "healthcare"
	}
	if strings.Contains(ind, "finance") || strings.Contains(ind, "bank") || strings.Contains(ind, "bfsi") {
		return "bfsi"
	}
	if strings.Contains(ind, "hr") || strings.Contains(ind, "human") || strings.Contains(ind, "recru") {
		return "hr"
	}
	return ind
}

// =============================================================================
// Pattern Packs
// =============================================================================

func (s *IndustryStrategy) loadHealthcarePatterns() {
	s.patterns["healthcare"] = []industryPattern{
		{
			Name:        "NPI (National Provider Identifier)",
			Regex:       regexp.MustCompile(`^\d{10}$`), // Simplified NPI (Luhn check needed for strict)
			Type:        types.PIITypeNationalID,
			Category:    types.PIICategoryProfessional,
			Sensitivity: types.SensitivityMedium,
		},
		{
			Name:        "DEA Number",
			Regex:       regexp.MustCompile(`^[A-Z]{2}\d{7}$`),
			Type:        types.PIITypeNationalID,
			Category:    types.PIICategoryProfessional,
			Sensitivity: types.SensitivityHigh, // Prescribing authority
		},
		{
			Name:        "ICD-10 Code",
			Regex:       regexp.MustCompile(`^[A-Z][0-9][0-9A-Z](\.[0-9A-Z]{1,4})?$`),
			Type:        types.PIITypeMedicalRecord,
			Category:    types.PIICategoryHealth,
			Sensitivity: types.SensitivityMedium, // Diagnosis code itself is medium, attached to person is critical
		},
	}
}

func (s *IndustryStrategy) loadFinancialPatterns() {
	s.patterns["bfsi"] = []industryPattern{
		{
			Name:        "SWIFT/BIC Code",
			Regex:       regexp.MustCompile(`^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$`),
			Type:        types.PIITypeBankAccount,
			Category:    types.PIICategoryFinancial,
			Sensitivity: types.SensitivityMedium,
		},
		{
			Name:        "IBAN",
			Regex:       regexp.MustCompile(`^[A-Z]{2}\d{2}[A-Z0-9]{11,30}$`), // Simplified
			Type:        types.PIITypeBankAccount,
			Category:    types.PIICategoryFinancial,
			Sensitivity: types.SensitivityHigh,
		},
	}
}

func (s *IndustryStrategy) loadHRPatterns() {
	s.patterns["hr"] = []industryPattern{
		{
			Name:        "Employee ID (Generic)",
			Regex:       regexp.MustCompile(`^(?i)EMP-?\d{3,6}$`), // Heuristic: EMP-12345
			Type:        types.PIITypeNationalID,                  // Using NationalID as proxy for generic ID
			Category:    types.PIICategoryProfessional,
			Sensitivity: types.SensitivityLow,
		},
	}
}
