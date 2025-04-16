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
		{"nil", nil, true},
		{"empty string", "", true},
		{"space string", "   ", true},
		{"non-empty string", "hello", false},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},
		{"empty array", [0]int{}, true},
		{"non-empty array", [2]int{1, 2}, false},
		{"empty map", map[string]string{}, true},
		{"non-empty map", map[string]string{"a": "b"}, false},
		{"nil chan", (chan int)(nil), true},
		{"empty chan", make(chan int), true},
		{"non-empty chan", ch, false},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", func() interface{} { i := 42; return &i }(), false},
		{"int", 123, false},
		{"struct", struct{ Name string }{"Hi"}, false},
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

func TestIsErrorOrEmpty(t *testing.T) {
	errSample := errors.New("some error")
	tests := []struct {
		name     string
		err      error
		value    interface{}
		expected bool
	}{
		{"no error, not empty", nil, "abc", false},
		{"with error", errSample, "abc", true},
		{"nil and empty", nil, "", true},
		{"error and empty", errSample, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorOrEmpty(tt.err, tt.value)
			if result != tt.expected {
				t.Errorf("IsErrorOrEmpty(%v, %v) = %v; want %v", tt.err, tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsPtr(t *testing.T) {
	a := 5
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"pointer int", &a, true},
		{"non-pointer", a, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPtr(tt.input)
			if result != tt.expected {
				t.Errorf("IsPtr(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsStruct(t *testing.T) {
	type MyStruct struct{ Name string }
	var s *MyStruct
	var i interface{}

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"nil", nil, false},
		{"int", 123, false},
		{"string", "abc", false},
		{"struct", MyStruct{"hello"}, true},
		{"pointer to struct", &MyStruct{"hi"}, true},
		{"untyped nil interface", i, false},
		{"typed nil pointer", s, true}, // still a typed nil pointer to struct
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
