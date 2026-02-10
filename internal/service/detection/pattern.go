package detection

import (
	"context"
	"regexp"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// PatternStrategy detects PII by matching sample values against compiled
// regex patterns. Ported from DataLens v1 with added India-specific patterns.
//
// Base confidence: 0.90 (regex matches are high-confidence).
type PatternStrategy struct{}

// NewPatternStrategy creates a new regex-based detection strategy.
func NewPatternStrategy() *PatternStrategy {
	return &PatternStrategy{}
}

func (s *PatternStrategy) Name() string                  { return "pattern" }
func (s *PatternStrategy) Method() types.DetectionMethod { return types.DetectionMethodRegex }
func (s *PatternStrategy) Weight() float64               { return 0.90 }

// Detect runs all regex patterns against the input samples.
func (s *PatternStrategy) Detect(ctx context.Context, input Input) ([]Result, error) {
	start := time.Now()
	_ = start // used implicitly by caller for timing

	if len(input.Samples) == 0 {
		return nil, nil
	}

	// Run each pattern against each sample, tally matches per category
	type matchInfo struct {
		category    types.PIICategory
		piiType     types.PIIType
		sensitivity types.SensitivityLevel
		matches     int
	}

	counts := make(map[types.PIIType]*matchInfo)

	for _, sample := range input.Samples {
		for _, p := range piiPatterns {
			if p.regex.MatchString(sample) {
				info, ok := counts[p.piiType]
				if !ok {
					info = &matchInfo{
						category:    p.category,
						piiType:     p.piiType,
						sensitivity: p.sensitivity,
					}
					counts[p.piiType] = info
				}
				info.matches++
			}
		}
	}

	if len(counts) == 0 {
		return nil, nil
	}

	// Convert to results, adjusting confidence by detection rate
	total := len(input.Samples)
	var results []Result
	for _, info := range counts {
		detectionRate := float64(info.matches) / float64(total)
		// adjustedScore = baseScore × (0.7 + 0.3 × detectionRate)
		confidence := 0.90 * (0.7 + 0.3*detectionRate)

		results = append(results, Result{
			Category:    info.category,
			Type:        info.piiType,
			Sensitivity: info.sensitivity,
			Confidence:  confidence,
			Method:      types.DetectionMethodRegex,
			Reasoning:   "Regex pattern matched",
		})
	}

	return results, nil
}

// =============================================================================
// Compiled Regex Patterns (ported from v1 + new India patterns)
// =============================================================================

type piiPattern struct {
	name        string
	category    types.PIICategory
	piiType     types.PIIType
	sensitivity types.SensitivityLevel
	regex       *regexp.Regexp
}

var piiPatterns = []piiPattern{
	// --- Email ---
	{
		name:        "email",
		category:    types.PIICategoryContact,
		piiType:     types.PIITypeEmail,
		sensitivity: types.SensitivityMedium,
		regex:       regexp.MustCompile(`(?i)^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
	},

	// --- Phone (India + US) ---
	{
		name:        "phone_india",
		category:    types.PIICategoryContact,
		piiType:     types.PIITypePhone,
		sensitivity: types.SensitivityMedium,
		regex:       regexp.MustCompile(`^(?:\+?91[\s\-]?)?[6-9]\d{9}$`),
	},
	{
		name:        "phone_us",
		category:    types.PIICategoryContact,
		piiType:     types.PIITypePhone,
		sensitivity: types.SensitivityMedium,
		regex:       regexp.MustCompile(`^\(?[2-9]\d{2}\)?[\s.\-]?\d{3}[\s.\-]?\d{4}$`),
	},

	// --- Aadhaar (India) ---
	{
		name:        "aadhaar",
		category:    types.PIICategoryGovernmentID,
		piiType:     types.PIITypeAadhaar,
		sensitivity: types.SensitivityCritical,
		regex:       regexp.MustCompile(`^[2-9]\d{3}[\s\-]?\d{4}[\s\-]?\d{4}$`),
	},

	// --- PAN (India) ---
	{
		name:        "pan",
		category:    types.PIICategoryGovernmentID,
		piiType:     types.PIITypePAN,
		sensitivity: types.SensitivityHigh,
		regex:       regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`),
	},

	// --- Credit Card (Visa, MasterCard, Amex, Discover) ---
	{
		name:        "credit_card",
		category:    types.PIICategoryFinancial,
		piiType:     types.PIITypeCreditCard,
		sensitivity: types.SensitivityCritical,
		regex:       regexp.MustCompile(`^(?:4\d{3}|5[1-5]\d{2}|3[47]\d{2}|6(?:011|5\d{2}))[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}$`),
	},

	// --- SSN (US) ---
	{
		name:        "ssn",
		category:    types.PIICategoryGovernmentID,
		piiType:     types.PIITypeSSN,
		sensitivity: types.SensitivityCritical,
		regex:       regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`),
	},

	// --- IP Address (IPv4) ---
	{
		name:        "ip_v4",
		category:    types.PIICategoryBehavioral,
		piiType:     types.PIITypeIPAddress,
		sensitivity: types.SensitivityLow,
		regex:       regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)$`),
	},

	// --- Date of Birth ---
	{
		name:        "dob",
		category:    types.PIICategoryIdentity,
		piiType:     types.PIITypeDOB,
		sensitivity: types.SensitivityMedium,
		regex:       regexp.MustCompile(`^(?:\d{1,2}[/\-]\d{1,2}[/\-]\d{2,4}|\d{4}[/\-]\d{1,2}[/\-]\d{1,2})$`),
	},

	// --- Passport ---
	{
		name:        "passport_india",
		category:    types.PIICategoryGovernmentID,
		piiType:     types.PIITypePassport,
		sensitivity: types.SensitivityHigh,
		regex:       regexp.MustCompile(`^[A-Z][1-9]\d{6}[1-9]$`),
	},
}
