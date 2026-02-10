package ai

import (
	"testing"
)

func TestSanitizer_Email(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("user@example.com")
	if got != "[EMAIL: format valid]" {
		t.Errorf("email: got %q, want %q", got, "[EMAIL: format valid]")
	}
}

func TestSanitizer_Aadhaar(t *testing.T) {
	s := NewSanitizer()
	cases := []string{"2345 6789 0123", "234567890123", "2345-6789-0123"}
	for _, c := range cases {
		got := s.Sanitize(c)
		if got != "[AADHAAR: 12 digits]" {
			t.Errorf("aadhaar %q: got %q, want %q", c, got, "[AADHAAR: 12 digits]")
		}
	}
}

func TestSanitizer_PAN(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("ABCDE1234F")
	if got != "[PAN: format valid]" {
		t.Errorf("pan: got %q, want %q", got, "[PAN: format valid]")
	}
}

func TestSanitizer_CreditCard(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("4111 1111 1111 1111")
	if got != "[CARD: 16 digits]" {
		t.Errorf("credit card: got %q, want %q", got, "[CARD: 16 digits]")
	}
}

func TestSanitizer_IP(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("192.168.1.1")
	if got != "[IP: v4]" {
		t.Errorf("ip: got %q, want %q", got, "[IP: v4]")
	}
}

func TestSanitizer_Phone(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("9876543210")
	if got != "[PHONE: 10 digits]" {
		t.Errorf("phone: got %q, want %q", got, "[PHONE: 10 digits]")
	}
}

func TestSanitizer_Date(t *testing.T) {
	s := NewSanitizer()
	cases := []string{"15/01/1990", "1990-01-15", "15-01-1990"}
	for _, c := range cases {
		got := s.Sanitize(c)
		if got != "[DATE: format detected]" {
			t.Errorf("date %q: got %q, want %q", c, got, "[DATE: format detected]")
		}
	}
}

func TestSanitizer_Name(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("Rahul Kumar")
	if got != "[TEXT: 2 words]" {
		t.Errorf("name: got %q, want %q", got, "[TEXT: 2 words]")
	}
}

func TestSanitizer_Empty(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("")
	if got != "[EMPTY]" {
		t.Errorf("empty: got %q, want %q", got, "[EMPTY]")
	}
}

func TestSanitizer_Numeric(t *testing.T) {
	s := NewSanitizer()
	got := s.Sanitize("12345")
	if got != "[NUMERIC: 5 digits]" {
		t.Errorf("numeric: got %q, want %q", got, "[NUMERIC: 5 digits]")
	}
}

func TestSanitizer_SanitizeSamples(t *testing.T) {
	s := NewSanitizer()
	samples := []string{"user@example.com", "ABCDE1234F", "Rahul Kumar"}
	got := s.SanitizeSamples(samples)
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}
	if got[0] != "[EMAIL: format valid]" {
		t.Errorf("sample 0: got %q", got[0])
	}
	if got[1] != "[PAN: format valid]" {
		t.Errorf("sample 1: got %q", got[1])
	}
	if got[2] != "[TEXT: 2 words]" {
		t.Errorf("sample 2: got %q", got[2])
	}
}
