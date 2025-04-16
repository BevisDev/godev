package list

import (
	"slices"
)

func IsContains[T comparable](slice []T, value T) bool {
	return slices.Contains(slice, value)
}

func IndexOf[T comparable](slice []T, value T) int {
	return slices.Index(slice, value)
}
