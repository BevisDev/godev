package pattern

import (
	"testing"
)

func TestPatterns(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) bool
		input    string
		expected bool
	}{
		{"Email valid", IsEmail, "test@example.com", true},
		{"Email invalid", IsEmail, "invalid@", false},

		{"Phone valid", IsPhoneNumber, "0123456789", true},
		{"Phone invalid", IsPhoneNumber, "123456789", false},

		{"UUID valid", IsUUID, "550e8400-e29b-41d4-a716-446655440000", true},
		{"UUID invalid", IsUUID, "550e8400", false},

		{"Date valid", IsDate, "2024-12-31", true},
		{"Date invalid", IsDate, "31-12-2024", false},

		{"Time valid", IsTime, "23:59:59", true},
		{"Time invalid", IsTime, "12:60:00", true},

		{"IPv4 valid", IsIPv4, "192.168.1.1", true},
		{"IPv4 invalid", IsIPv4, "999.999.999.999", false},

		{"AlphaNumeric valid", IsAlphaNumeric, "abc123", true},
		{"AlphaNumeric invalid", IsAlphaNumeric, "abc 123", false},

		{"HexColor valid", IsHexColor, "#fff", true},
		{"HexColor invalid", IsHexColor, "#zzzzzz", false},

		{"VietnamID CMND", IsVietnamID, "123456789", true},
		{"VietnamID CCCD", IsVietnamID, "123456789012", true},
		{"VietnamID invalid", IsVietnamID, "12345678", false},

		{"SafeFilename valid", IsSafeFileName, "file.txt", true},
		{"SafeFilename invalid", IsSafeFileName, "file@.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.expected {
				t.Errorf("Got %v, want %v for input %q", result, tt.expected, tt.input)
			}
		})
	}
}

func TestIsStrongPassword(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"P@ssw0rd", true},
		{"Strong1!", true},
		{"My$ecret9", true},

		{"password", false},
		{"Passw0rd", false},
		{"12345678!", false},
		{"PASSWORD1!", false},
		{"password1!", false},
		{"Pa1!", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsStrongPassword(tt.input)
			if result != tt.expected {
				t.Errorf("IsStrongPassword(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractAllMatches(t *testing.T) {
	input := "Emails: test1@mail.com, test2@abc.org"
	pattern := `[\w\.-]+@[\w\.-]+\.\w+`
	expected := []string{"test1@mail.com", "test2@abc.org"}

	result := ExtractAllMatches(input, pattern)
	if len(result) != len(expected) {
		t.Fatalf("Expected %d matches, got %d", len(expected), len(result))
	}
	for i, match := range result {
		if match != expected[i] {
			t.Errorf("Expected %q, got %q", expected[i], match)
		}
	}
}
