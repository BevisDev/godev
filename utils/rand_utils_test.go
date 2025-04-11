package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unicode"
)

func TestRandIntInRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := RandInt(5, 10)
		assert.GreaterOrEqual(t, val, 5)
		assert.Less(t, val, 10)
	}
}

func TestRandomFloatInRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := RandomFloatInRange(1.5, 3.0)
		assert.GreaterOrEqual(t, val, 1.5)
		assert.Less(t, val, 3.0)
	}
}

func TestRandStringLength(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(int) string
		length int
	}{
		{"RandString", RandString, 16},
		{"RandStringNumeric", RandStringNumeric, 10},
		{"RandStringUpper", RandStringUpper, 12},
		{"RandStringLower", RandStringLower, 8},
		{"RandStringUpperNumeric", RandStringUpperNumeric, 20},
		{"RandStringLowerNumeric", RandStringLowerNumeric, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fn(tt.length)
			assert.Equal(t, tt.length, len(s))
		})
	}
}

func TestRandStringContent(t *testing.T) {
	s := RandStringNumeric(50)
	for _, ch := range s {
		assert.True(t, unicode.IsDigit(ch), "expected numeric char, got %c", ch)
	}

	s = RandStringUpper(50)
	for _, ch := range s {
		assert.True(t, unicode.IsUpper(ch), "expected upper case char, got %c", ch)
	}

	s = RandStringLower(50)
	for _, ch := range s {
		assert.True(t, unicode.IsLower(ch), "expected lower case char, got %c", ch)
	}

	s = RandStringUpperNumeric(50)
	for _, ch := range s {
		assert.True(t, unicode.IsUpper(ch) || unicode.IsDigit(ch), "expected upper or digit, got %c", ch)
	}

	s = RandStringLowerNumeric(50)
	for _, ch := range s {
		assert.True(t, unicode.IsLower(ch) || unicode.IsDigit(ch), "expected lower or digit, got %c", ch)
	}
}

func TestRandPickFromStrings(t *testing.T) {
	input := []string{"apple", "banana", "cherry"}
	found := make(map[string]bool)

	for i := 0; i < 100; i++ {
		picked := RandPick(input)
		found[picked] = true
	}

	for _, val := range input {
		assert.True(t, found[val], "expected value %s to be picked", val)
	}
}

func TestRandPickFromInts(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := RandPick(input)
	assert.Contains(t, input, result)
}

func TestRandPickFromEmpty(t *testing.T) {
	var input []int
	result := RandPick(input)
	assert.Equal(t, 0, result)
}

func TestRandPickFromStruct(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	picked := RandPick(users)
	assert.Contains(t, []int{1, 2, 3}, picked.ID)
}
