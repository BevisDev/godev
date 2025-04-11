package utils

import (
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

func RandInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func RandomFloatInRange(min, max float64) float64 {
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
		result[i] = layout[RandInt(0, len(layout)-1)]
	}
	return string(result)
}

func RandPick[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[rand.Intn(len(slice))]
}
