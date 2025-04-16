package list

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsContains(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		value    int
		expected bool
	}{
		{"value exists", []int{1, 2, 3}, 2, true},
		{"value not exists", []int{1, 2, 3}, 5, false},
		{"empty array", []int{}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsContains(tt.arr, tt.value)
			if result != tt.expected {
				t.Errorf("IsContains(%v, %v) = %v; want %v", tt.arr, tt.value, result, tt.expected)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	slice := []int{10, 20, 30, 40}

	assert.Equal(t, 2, IndexOf(slice, 30))
	assert.Equal(t, -1, IndexOf(slice, 100))
}
