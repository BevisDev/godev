package money

import (
	"github.com/BevisDev/godev/types"
	"github.com/shopspring/decimal"
	"testing"
)

func TestFromFloat_ToFloat(t *testing.T) {
	val := 123.45
	m := FromFloat(val)
	f := ToFloat(m)

	if f != val {
		t.Errorf("Expected %f, got %f", val, f)
	}
}

func TestFromInt_ToInt(t *testing.T) {
	val := 100
	m := FromInt(val)
	i := ToInt(m)

	if i != val {
		t.Errorf("Expected %d, got %d", val, i)
	}
}

func TestIsZero(t *testing.T) {
	zero := FromFloat(0.0)
	if !IsZero(zero) {
		t.Errorf("Expected IsZero to be true")
	}
}

func TestIsPositive(t *testing.T) {
	m := FromFloat(10.5)
	if !IsPositive(m) {
		t.Errorf("Expected IsPositive to be true")
	}
}

func TestIsNegative(t *testing.T) {
	m := FromFloat(-5.5)
	if !IsNegative(m) {
		t.Errorf("Expected IsNegative to be true")
	}
}

func TestFloatComparisons(t *testing.T) {
	m := FromFloat(100.5)

	if !GreaterThanFloat(m, 100.0) {
		t.Errorf("Expected > float to be true")
	}
	if !GreaterThanOrEqualFloat(m, 100.5) {
		t.Errorf("Expected >= float to be true")
	}
	if !LessThanFloat(m, 101.0) {
		t.Errorf("Expected < float to be true")
	}
	if !LessThanOrEqualFloat(m, 100.5) {
		t.Errorf("Expected <= float to be true")
	}
}

func TestIntComparisons(t *testing.T) {
	m := FromInt(200)

	if !GreaterThanInt(m, 199) {
		t.Errorf("Expected > int to be true")
	}
	if !GreaterThanOrEqualInt(m, 200) {
		t.Errorf("Expected >= int to be true")
	}
	if !LessThanInt(m, 201) {
		t.Errorf("Expected < int to be true")
	}
	if !LessThanOrEqualInt(m, 200) {
		t.Errorf("Expected <= int to be true")
	}
}

func newMoney(val string) types.Money {
	m, err := decimal.NewFromString(val)
	if err != nil {
		panic(err)
	}
	return m
}

func TestRound(t *testing.T) {
	tests := []struct {
		input    string
		places   int32
		expected string
	}{
		{"123.4567", 2, "123.46"},
		{"123.444", 2, "123.44"},
		{"-123.555", 1, "-123.6"},
	}

	for _, tc := range tests {
		m := newMoney(tc.input)
		result := Round(m, tc.places)

		if result.StringFixed(tc.places) != tc.expected {
			t.Errorf("Round(%s, %d): expected %s, got %s",
				tc.input, tc.places, tc.expected, result.StringFixed(tc.places))
		}
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-100.25", "100.25"},
		{"0.00", "0.00"},
		{"99.99", "99.99"},
	}

	for _, tc := range tests {
		m := newMoney(tc.input)
		result := Abs(m)

		if !result.Equal(newMoney(tc.expected)) {
			t.Errorf("Abs(%s): expected %s, got %s",
				tc.input, tc.expected, result.String())
		}
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input    string
		places   int32
		expected string
	}{
		{"100", 2, "100.00"},
		{"123.4", 2, "123.40"},
		{"123.456", 1, "123.5"},
		{"-1.2345", 3, "-1.235"},
	}

	for _, tc := range tests {
		m := newMoney(tc.input)
		result := Format(m, tc.places)

		if result != tc.expected {
			t.Errorf("Format(%s, %d): expected %s, got %s",
				tc.input, tc.places, tc.expected, result)
		}
	}
}

func TestIsDecimal(t *testing.T) {
	d := decimal.NewFromFloat(1.23)
	dPtr := &d

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"Decimal value", d, true},
		{"Pointer to Decimal", dPtr, true},
		{"Float64", 1.23, false},
		{"String", "1.23", false},
		{"Nil", nil, false},
		{"Int", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDecimal(tt.input)
			if result != tt.expected {
				t.Errorf("IsDecimal(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
