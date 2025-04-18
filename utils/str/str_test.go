package str

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int", 123, "123"},
		{"int64", int64(999), "999"},
		{"float64", 3.14159, "3.14159"},
		{"float32", float32(1.618), "1.618"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"string", "hello", "hello"},
		{"[]byte", []byte("world"), "world"},
		{"nil", nil, ""},

		// pointer cases
		{"*int", func() any { v := 42; return &v }(), "42"},
		{"*int64", func() any { v := int64(888); return &v }(), "888"},
		{"*string", func() any { s := "pointer"; return &s }(), "pointer"},
		{"*bool", func() any { b := true; return &b }(), "true"},
		{"*float64", func() any { f := 3.3; return &f }(), "3.3"},
		{"nil *int", func() any { var p *int = nil; return p }(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.input)
			if result != tt.expected {
				t.Errorf("ToString(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{"0", 0},
		{"-123", -123},
		{"abc", 0},
		{"", 0},
		{"9999999999", 9999999999},
	}

	for _, tt := range tests {
		result := ToInt(tt.input)
		assert.Equal(t, tt.expected, result, "ToInt(%s) should be %d", tt.input, tt.expected)
	}
}

func TestToFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"0", 0},
		{"-2.718", -2.718},
		{"abc", 0.0},
		{"", 0.0},
		{"1e10", 1e10},
	}

	for _, tt := range tests {
		result := ToFloat(tt.input)
		assert.InDelta(t, tt.expected, result, 0.0001, "ToFloat(%s) should be approx %.4f", tt.input, tt.expected)
	}
}

func TestNormalizeToASCII(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cà phê sữa đá", "Ca phe sua da"},
		{"Thành phố Hồ Chí Minh", "Thanh pho Ho Chi Minh"},
		{"Điện Biên Phủ", "Dien Bien Phu"},
		{"Tôi yêu tiếng Việt", "Toi yeu tieng Viet"},
		{"Không dấu", "Khong dau"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeToASCII(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeToASCII(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cà-phê! sữa. đá?", "Caphe sua da"},
		{"Hello @#$%^&*()", "Hello "},
		{"Tên tôi là: Trần Văn *A*", "Ten toi la Tran Van A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CleanText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanText(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveWhiteSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "HelloWorld"},
		{"  A B  C ", "ABC"},
		{"Không có khoảng trắng", "Khôngcókhoảngtrắng"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RemoveWhiteSpace(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveWhiteSpace(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "Short string",
			input:    "Hello",
			max:      10,
			expected: "Hello",
		},
		{
			name:     "Exact length",
			input:    "HelloWorld",
			max:      10,
			expected: "HelloWorld",
		},
		{
			name:     "Long string",
			input:    "Hello, this is a long message",
			max:      5,
			expected: "Hello",
		},
		{
			name:     "Empty string",
			input:    "",
			max:      10,
			expected: "",
		},
		{
			name:     "Zero max",
			input:    "Hello",
			max:      0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := TruncateText(tt.input, tt.max)
			if actual != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, actual)
			}
		})
	}
}
