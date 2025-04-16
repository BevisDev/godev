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

func RandUUID() string {
	return uuid.NewString()
}

func RandInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func RandFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
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

func randStr(length int, layout string) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = layout[RandInt(0, len(layout))]
	}
	return string(result)
}

func RandPick[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[RandInt(0, len(slice))]
}
