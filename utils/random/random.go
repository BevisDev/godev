package random

import (
	"github.com/google/uuid"
	"math/rand"
)

const (
	upperAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerAlphabet = "abcdefghijklmnopqrstuvwxyz"
	numeric       = "0123456789"
	upperCharset  = upperAlphabet + numeric
	lowerCharset  = lowerAlphabet + numeric
	charset       = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz" + numeric
)

// RandUUID generates a new random UUID and returns it as a string.
//
// Example:
//
//	id := RandUUID()
//	fmt.Println(id) // "550e8400-e29b-41d4-a716-446655440000"
func RandUUID() string {
	return uuid.NewString()
}

// RandInt returns a random integer in the half-open interval [min, max).
// The result is always >= min and < max.
//
// Special cases:
//   - If min == max, the function returns min.
//   - If min > max, min and max are swapped.
//
// Example:
//
//	n := RandInt(0, 10)
//	// n is between 0 and 9 (inclusive)
func RandInt(min, max int) int {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Intn(max-min)
}

// RandFloat returns a random float64 in the half-open interval [min, max).
// The result is always >= min and < max.
//
// Special cases:
//   - If min == max, the function returns min.
//   - If min > max, min and max are swapped.
//
// Example:
//
//	f := RandFloat(1.5, 5.5)
//	// f is >= 1.5 and < 5.5
//
//	f = RandFloat(3.0, 3.0)
//	// f == 3.0
//
//	f = RandFloat(10.0, 2.0)
//	// f >= 2.0 and < 10.0
func RandFloat(min, max float64) float64 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Float64()*(max-min)
}

// RandPick returns a random element from the given slice.
// If the slice is empty, it returns the zero value of type T.
//
// Example:
//
//	names := []string{"Alice", "Bob", "Charlie"}
//	name := RandPick(names)
//	// name is randomly one of "Alice", "Bob", "Charlie"
//
//	empty := []int{}
//	n := RandPick(empty)
//	// n == 0 (zero value for int)
func RandPick[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}

	return slice[RandInt(0, len(slice))]
}

// randStr generates a random string of the specified length,
// using the provided layout (a string containing allowed characters).
//
// For each character in the result, a random character is picked from layout.
//
// If layout is empty or length <= 0, it returns an empty string.
//
// Example:
//
//	s := randStr(8, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
//	// s might be "aZ3bT1xQ"
func randStr(length int, layout string) string {
	if length <= 0 || len(layout) == 0 {
		return ""
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = layout[RandInt(0, len(layout))]
	}
	
	return string(result)
}

func RandString(length int) string {
	return randStr(length, charset)
}

func RandStringNumeric(length int) string {
	return randStr(length, numeric)
}

func RandStringUpper(length int) string {
	return randStr(length, upperAlphabet)
}

func RandStringUpperNumeric(length int) string {
	return randStr(length, upperCharset)
}

func RandStringLower(length int) string {
	return randStr(length, lowerAlphabet)
}

func RandStringLowerNumeric(length int) string {
	return randStr(length, lowerCharset)
}
