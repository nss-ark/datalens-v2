package ai

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode"
)

// Sanitizer transforms raw data samples into safe pattern descriptions
// before sending them to external AI providers. This ensures no real PII
// is ever transmitted to cloud LLMs.
//
// Examples:
//
//	"Rahul Kumar"       → "[NAME: 2 words]"
//	"rahul@example.com" → "[EMAIL: format valid]"
//	"2345 6789 0123"    → "[AADHAAR: 12 digits]"
//	"ABCDE1234F"        → "[PAN: format valid]"
//	"192.168.1.1"       → "[IP: v4]"
type Sanitizer struct{}

// NewSanitizer creates a new PII sanitizer.
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

// SanitizeSamples transforms a slice of raw values into safe pattern descriptions.
func (s *Sanitizer) SanitizeSamples(samples []string) []string {
	sanitized := make([]string, len(samples))
	for i, sample := range samples {
		sanitized[i] = s.Sanitize(sample)
	}
	return sanitized
}

// Sanitize transforms a single raw value into a safe pattern description.
func (s *Sanitizer) Sanitize(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "[EMPTY]"
	}

	// Check specific patterns (most specific first)
	if isEmail(trimmed) {
		return "[EMAIL: format valid]"
	}
	if isAadhaar(trimmed) {
		return "[AADHAAR: 12 digits]"
	}
	if isPAN(trimmed) {
		return "[PAN: format valid]"
	}
	if isCreditCard(trimmed) {
		return fmt.Sprintf("[CARD: %d digits]", countDigits(trimmed))
	}
	if isIP(trimmed) {
		return "[IP: v4]"
	}
	if isPhone(trimmed) {
		return fmt.Sprintf("[PHONE: %d digits]", countDigits(trimmed))
	}
	if isDate(trimmed) {
		return "[DATE: format detected]"
	}

	// Generic pattern analysis
	if isAllDigits(trimmed) {
		return fmt.Sprintf("[NUMERIC: %d digits]", len(trimmed))
	}
	if isAllAlpha(trimmed) {
		words := strings.Fields(trimmed)
		if len(words) > 1 {
			return fmt.Sprintf("[TEXT: %d words]", len(words))
		}
		return fmt.Sprintf("[TEXT: %d chars]", len(trimmed))
	}

	// Mixed content
	words := strings.Fields(trimmed)
	return fmt.Sprintf("[MIXED: %d chars, %d words]", len(trimmed), len(words))
}

// --- Pattern matchers ---

var (
	emailRe      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	aadhaarRe    = regexp.MustCompile(`^\d{4}[\s-]?\d{4}[\s-]?\d{4}$`)
	panRe        = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)
	creditCardRe = regexp.MustCompile(`^\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}$`)
	phoneRe      = regexp.MustCompile(`^(?:\+?91[\s-]?)?[6-9]\d{9}$`)
	dateRe       = regexp.MustCompile(`^(?:\d{1,2}[/\-]\d{1,2}[/\-]\d{2,4}|\d{4}[/\-]\d{1,2}[/\-]\d{1,2})$`)
)

func isEmail(s string) bool      { return emailRe.MatchString(s) }
func isAadhaar(s string) bool    { return aadhaarRe.MatchString(s) }
func isPAN(s string) bool        { return panRe.MatchString(s) }
func isCreditCard(s string) bool { return creditCardRe.MatchString(s) }
func isPhone(s string) bool      { return phoneRe.MatchString(s) }
func isDate(s string) bool       { return dateRe.MatchString(s) }

func isIP(s string) bool {
	return net.ParseIP(s) != nil
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func isAllAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return len(s) > 0
}

func countDigits(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			count++
		}
	}
	return count
}
