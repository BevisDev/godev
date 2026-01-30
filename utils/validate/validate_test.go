package validate

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestIsNilOrEmpty(t *testing.T) {
	type testCase struct {
		name     string
		input    interface{}
		expected bool
	}

	ch := make(chan int, 1)
	ch <- 1

	tests := []testCase{
		// --- basic ---
		{"nil", nil, true},
		{"empty string", "", true},
		{"space string", "   ", true},
		{"non-empty string", "hello", false},

		// -- slice ---
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},

		// -- array
		{"empty array", [0]int{}, true},
		{"non-empty array", [2]int{1, 2}, false},

		// -- map
		{"empty map", map[string]string{}, true},
		{"non-empty map", map[string]string{"a": "b"}, false},

		// channel
		{"nil chan", (chan int)(nil), true},
		{"empty chan", make(chan int), true},
		{"non-empty chan", ch, false},

		// pointer
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", func() interface{} { i := 42; return &i }(), false},

		{"int", 123, false},
		{"struct", struct{ Name string }{"Hi"}, false},
		{"nil *string", (*string)(nil), true},
		{"empty *string", func() interface{} { s := ""; return &s }(), true},
		{"non-empty *string", func() interface{} { s := "hello"; return &s }(), false},
		{"nil *[]int", (*[]int)(nil), true},
		{"empty *[]int", func() interface{} { var s []int; return &s }(), true},
		{"non-empty *[]int", func() interface{} { var s = []int{1}; return &s }(), false},
		{"nil *map", (*map[string]string)(nil), true},
		{"empty *map", func() interface{} { m := map[string]string{}; return &m }(), true},
		{"non-empty *map", func() interface{} { m := map[string]string{"x": "y"}; return &m }(), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNilOrEmpty(tc.input)
			if result != tc.expected {
				t.Errorf("IsNilOrEmpty(%v) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsNilOrNumericZero(t *testing.T) {
	var (
		nilIntPointer   *int
		nilFloatPointer *float64
	)

	var (
		zeroInt = 0
		i       = 42
	)

	var (
		zeroFloat = float64(0)
		f         = 3.14
	)

	tests := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"Nil interface", nil, true},
		{"Nil int pointer", nilIntPointer, true},
		{"Nil float pointer", nilFloatPointer, true},
		{"Zero int value", 0, true},
		{"Non-zero int value", 100, false},
		{"Zero int pointer", &zeroInt, true},
		{"Non-zero int pointer", &i, false},
		{"Zero float value", 0.0, true},
		{"Non-zero float value", 1.23, false},
		{"Zero float pointer", &zeroFloat, true},
		{"Non-zero float pointer", &f, false},
		{"Zero uint value", uint(0), true},
		{"Non-zero uint value", uint(55), false},
		{"Zero int64 value", int64(0), true},
		{"Non-zero int64 value", int64(-1), false},
		{"Unsupported type (string)", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNilOrNumericZero(tt.v)
			if got != tt.want {
				t.Errorf("IsNilOrNumericZero(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestMustSucceed(t *testing.T) {
	errSample := errors.New("some error")
	tests := []struct {
		name     string
		err      error
		value    interface{}
		hasError bool
	}{
		{"no error, not empty", nil, "abc", false},
		{"with error", errSample, "abc", true},
		{"nil and empty", nil, "", true},
		{"error and empty", errSample, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNonEmpty(tt.err, tt.value)
			if (err != nil) != tt.hasError {
				t.Errorf(
					"RequireNonEmpty(%v, %v) error = %v; want error = %v",
					tt.err, tt.value, err, tt.hasError,
				)
			}
		})
	}
}

func TestIsNonNilPointer(t *testing.T) {
	a := 5
	var pNil *int

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"pointer int", &a, true},    // IsPointer returns nil (no error) for pointer
		{"non-pointer", a, false},    // IsPointer returns error for non-pointer
		{"nil", nil, false},          // IsPointer returns error for nil
		{"nil pointer", pNil, false}, // IsPointer returns error for nil pointer (nil value)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNonNilPointer(tt.input)
			if result != tt.wantErr {
				t.Errorf(
					"IsNonNilPointer(%v) result = %v; wantErr = %v",
					tt.input, result, tt.wantErr,
				)
			}
		})
	}
}

func TestIsStruct(t *testing.T) {
	type MyStruct struct{ Name string }

	var s *MyStruct
	var i interface{}
	var ip *int

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"struct", MyStruct{"hello"}, true},
		{"pointer to struct", &MyStruct{"hi"}, true},
		{"nil", nil, false},
		{"int", 123, false},
		{"string", "abc", false},
		{"untyped nil interface", i, false},
		{"typed nil pointer", s, false},
		{"pointer to int", ip, false},
		{"slice", []int{1, 2, 3}, false},
		{"map", map[string]int{"a": 1}, false},
		{"array", [2]int{1, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStruct(tt.input)
			if result != tt.expected {
				t.Errorf("IsStruct(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsTimedOut(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"timeout error", context.DeadlineExceeded, true},
		{"wrapped timeout", fmt.Errorf("wrapped: %w", context.DeadlineExceeded), true},
		{"non-timeout error", errors.New("some other error"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimedOut(tt.err)
			if result != tt.expected {
				t.Errorf("IsTimedOut(%v) = %v; want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsDigits(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// check valid phone number
		{"0123456789", true},
		{"01234", true},
		{"012345678901", true},

		// check string
		{"", false},
		{"    ", false},
		{"01234567890", true},
		{"abcdefghij", false},
		{"01234-7890", false},

		// Contains non-numeric characters
		{"12345678a901", false},
		{"abcdefghi", false},
		{"1234 56789", false},

		// Empty
		{"", false},
	}

	for _, tt := range tests {
		got := IsDigits(tt.input)
		if got != tt.want {
			t.Errorf("IsDigits(%q) = %v; want %v", tt.input, got, tt.want)
		}
	}
}

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

		{"IPv4 valid", IsIPv4, "192.168.1.1", true},
		{"IPv4 invalid", IsIPv4, "999.999.999.999", false},

		{"AlphaNumeric valid", IsAlphaNumeric, "abc123", true},
		{"AlphaNumeric invalid", IsAlphaNumeric, "abc 123", false},

		{"VietnamID CMND", IsVietnamID, "123456789", true},
		{"VietnamID CCCD", IsVietnamID, "123456789012", true},
		{"VietnamID invalid", IsVietnamID, "12345678", false},
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
			result := IsStrongPassword(tt.input, 8)
			if result != tt.expected {
				t.Errorf("IsStrongPassword(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidFileName(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		allowedExt []string
		want       bool
	}{
		{
			name:       "Valid txt file",
			filename:   "report.txt",
			allowedExt: []string{"txt", "csv"},
			want:       true,
		},
		{
			name:       "Valid CSV file",
			filename:   "data-file_2023.csv",
			allowedExt: []string{"csv"},
			want:       true,
		},
		{
			name:       "Invalid extension",
			filename:   "archive.7z",
			allowedExt: []string{"zip"},
			want:       false,
		},
		{
			name:       "No extension",
			filename:   "readme",
			allowedExt: []string{"txt"},
			want:       false,
		},
		{
			name:       "Ends with dot",
			filename:   "file.",
			allowedExt: []string{"txt"},
			want:       false,
		},
		{
			name:       "Uppercase extension",
			filename:   "presentation.PDF",
			allowedExt: []string{"pdf"},
			want:       true,
		},
		{
			name:       "Ext with number",
			filename:   "archive.v2",
			allowedExt: []string{"v2"},
			want:       true,
		},
		{
			name:       "Pattern mismatch",
			filename:   "bad*name.txt",
			allowedExt: []string{"txt"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidFileName(tt.filename, tt.allowedExt)
			if got != tt.want {
				t.Errorf("IsValidFileName(%q, %v) = %v, want %v",
					tt.filename, tt.allowedExt, got, tt.want)
			}
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "valid json string",
			input: `{"a":1,"b":true}`,
			want:  true,
		},
		{
			name:  "invalid json string redundant ,",
			input: `{"a":1,,"b":true}`,
			want:  false,
		},
		{
			name:  "invalid json string miss }",
			input: `{"a":1,`,
			want:  false,
		},
		{
			name:  "valid json bytes",
			input: []byte(`["a", 1, true]`),
			want:  true,
		},
		{
			name:  "invalid json bytes",
			input: []byte(`{abc}`),
			want:  false,
		},
		{
			name: "map is valid json",
			input: map[string]interface{}{
				"a": 1,
				"b": true,
			},
			want: true,
		},
		{
			name: "struct is valid json",
			input: struct {
				A int
				B string
			}{A: 1, B: "x"},
			want: true,
		},
		{
			name:  "channel cannot be marshaled",
			input: make(chan int),
			want:  false,
		},
		{
			name:  "function cannot be marshaled",
			input: func() {},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidJSON(tt.input); got != tt.want {
				t.Fatalf("IsValidJSON(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"int", 1, true},
		{"int64", int64(10), true},
		{"uint", uint(5), true},
		{"float32", float32(1.5), true},
		{"float64", 3.14, true},

		{"string number", "123", false},
		{"bool", true, false},
		{"nil", nil, false},
		{"struct", struct{}{}, false},
		{"json number as string", `1`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumber(tt.input); got != tt.want {
				t.Fatalf("IsNumber(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsBoolean(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"true", true, true},
		{"false", false, true},

		{"string true", "true", false},
		{"number", 1, false},
		{"nil", nil, false},
		{"struct", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBoolean(tt.input); got != tt.want {
				t.Fatalf("IsBoolean(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSlice(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "nil",
			input: nil,
			want:  false,
		},
		{
			name:  "int slice",
			input: []int{1, 2, 3},
			want:  true,
		},
		{
			name:  "string slice",
			input: []string{"a", "b"},
			want:  true,
		},
		{
			name:  "empty slice",
			input: []int{},
			want:  true,
		},
		{
			name:  "array is not slice",
			input: [3]int{1, 2, 3},
			want:  false,
		},
		{
			name:  "map is not slice",
			input: map[string]int{"a": 1},
			want:  false,
		},
		{
			name:  "string is not slice",
			input: "abc",
			want:  false,
		},
		{
			name:  "struct is not slice",
			input: struct{}{},
			want:  false,
		},
		{
			name:  "typed nil slice",
			input: []int(nil),
			want:  true,
		},
		{
			name:  "pointer to slice is not slice",
			input: &[]int{1, 2},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSlice(tt.input)
			if got != tt.want {
				t.Fatalf("IsSlice(%#v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
