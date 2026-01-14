package random

import (
	"regexp"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestGenUUID(t *testing.T) {
	uuid := NewUUID()
	if uuid == "" {
		t.Errorf("NewUUID() = empty string")
	}
	r := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-7[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	if !r.MatchString(uuid) {
		t.Errorf("NewUUID() = %q, not a valid UUID", uuid)
	}
}

func TestRandInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := NewInt(5, 10)
		assert.GreaterOrEqual(t, val, 5)
		assert.Less(t, val, 10)
	}
}

func TestRandomFloatInRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := NewFloat(1.5, 3.0)
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
		{"NewString", NewString, 16},
		{"NewNumericString", NewNumericString, 10},
		{"NewUpperString", NewUpperString, 12},
		{"NewLowerString", NewLowerString, 8},
		{"NewUpperStringNumeric", NewUpperStringNumeric, 20},
		{"NewLowerStringNumeric", NewLowerStringNumeric, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fn(tt.length)
			assert.Equal(t, tt.length, len(s))
		})
	}
}

func TestRandStringContent(t *testing.T) {
	s := NewNumericString(50)
	for _, ch := range s {
		assert.True(t, unicode.IsDigit(ch), "expected numeric char, got %c", ch)
	}

	s = NewUpperString(50)
	for _, ch := range s {
		assert.True(t, unicode.IsUpper(ch), "expected upper case char, got %c", ch)
	}

	s = NewLowerString(50)
	for _, ch := range s {
		assert.True(t, unicode.IsLower(ch), "expected lower case char, got %c", ch)
	}

	s = NewUpperStringNumeric(50)
	for _, ch := range s {
		assert.True(t, unicode.IsUpper(ch) || unicode.IsDigit(ch), "expected upper or digit, got %c", ch)
	}

	s = NewLowerStringNumeric(50)
	for _, ch := range s {
		assert.True(t, unicode.IsLower(ch) || unicode.IsDigit(ch), "expected lower or digit, got %c", ch)
	}
}

func TestRandPickFromStrings(t *testing.T) {
	input := []string{"apple", "banana", "cherry"}
	found := make(map[string]bool)

	for i := 0; i < 100; i++ {
		picked := Item(input)
		found[picked] = true
	}

	for _, val := range input {
		assert.True(t, found[val], "expected value %s to be picked", val)
	}
}

func TestRandPickFromInts(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := Item(input)
	assert.Contains(t, input, result)
}

func TestRandPickFromEmpty(t *testing.T) {
	var input []int
	result := Item(input)
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
	picked := Item(users)
	assert.Contains(t, []int{1, 2, 3}, picked.ID)
}
