package config

import (
	"regexp"
	"testing"
)

func TestSanitizePassword(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123", "ABC123"},
		{"Hello@World#99", "HELLOWORLD99"},
		{"  xyz  ", "XYZ"},
		{"123456", "123456"},
		{"!@#$%^", ""},
	}

	for _, tt := range tests {
		got := SanitizePassword(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizePassword(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9A-Z]{8}$`)
	for i := 0; i < 50; i++ {
		pwd := generateRandomPassword(8)
		if !pattern.MatchString(pwd) {
			t.Fatalf("Generated password %q does not match [0-9A-Z]{8}", pwd)
		}
	}
}
