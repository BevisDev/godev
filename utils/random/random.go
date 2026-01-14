package random

import (
	"math/rand"

	"github.com/google/uuid"
)

const (
	upperAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerAlphabet = "abcdefghijklmnopqrstuvwxyz"
	numeric       = "0123456789"
	upperCharset  = upperAlphabet + numeric
	lowerCharset  = lowerAlphabet + numeric
	charset       = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz" + numeric
)

// NewUUID generates a new random UUID and returns it as a string.
//
// Example:
//
//	id := NewUUID()
//	fmt.Println(id) // "550e8400-e29b-71d4-a716-446655440000"
func NewUUID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// NewInt returns a random integer in the half-open interval [min, max).
// The result is always >= min and < max.
//
// Special cases:
//   - If min == max, the function returns min.
//   - If min > max, min and max are swapped.
//
// Example:
//
//	n := NewInt(0, 10)
//	// n is between 0 and 9 (inclusive)
func NewInt(min, max int) int {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Intn(max-min)
}

// NewFloat returns a random float64 in the half-open interval [min, max).
// The result is always >= min and < max.
//
// Special cases:
//   - If min == max, the function returns min.
//   - If min > max, min and max are swapped.
//
// Example:
//
//	f := NewFloat(1.5, 5.5)
//	// f is >= 1.5 and < 5.5
//
//	f = NewFloat(3.0, 3.0)
//	// f == 3.0
//
//	f = NewFloat(10.0, 2.0)
//	// f >= 2.0 and < 10.0
func NewFloat(min, max float64) float64 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Float64()*(max-min)
}

// Item returns a random element from the given slice.
// If the slice is empty, it returns the zero value of type T.
//
// Example:
//
//	names := []string{"Alice", "Bob", "Charlie"}
//	name := Item(names)
//	// name is randomly one of "Alice", "Bob", "Charlie"
//
//	empty := []int{}
//	n := Item(empty)
//	// n == 0 (zero value for int)
func Item[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}

	return slice[NewInt(0, len(slice))]
}

func NewString(length int) string {
	return randStr(length, charset)
}

func NewNumeric(length int) string {
	return randStr(length, numeric)
}

func NewNumericString(length int) string {
	return randStr(length, numeric)
}

func NewUpperString(length int) string {
	return randStr(length, upperAlphabet)
}

func NewUpperStringNumeric(length int) string {
	return randStr(length, upperCharset)
}

func NewLowerString(length int) string {
	return randStr(length, lowerAlphabet)
}

func NewLowerStringNumeric(length int) string {
	return randStr(length, lowerCharset)
}

// randStr generates a random string of the specified length,
// using the provided charset  (a string containing allowed characters).
//
// For each character in the result, a random character is picked from layout.
//
// If charset is empty or length <= 0, it returns an empty string.
//
// Example:
//
//	s := randStr(8, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
//	// s might be "aZ3bT1xQ"
func randStr(length int, charset string) string {
	if length <= 0 || len(charset) == 0 {
		return ""
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[NewInt(0, len(charset))]
	}
	return string(result)
}
