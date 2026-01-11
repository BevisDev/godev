package money

import (
	"github.com/shopspring/decimal"
)

type Money = decimal.Decimal

func FromFloat(f float64) Money {
	return decimal.NewFromFloat(f)
}

func ToFloat(m Money) float64 {
	f, _ := m.Float64()
	return f
}

func IsDecimal(val interface{}) bool {
	switch val.(type) {
	case decimal.Decimal, *decimal.Decimal:
		return true
	default:
		return false
	}
}

func GreaterThanFloat(m Money, f float64) bool {
	return m.GreaterThan(FromFloat(f))
}

func GreaterThanOrEqualFloat(m Money, f float64) bool {
	return m.GreaterThanOrEqual(FromFloat(f))
}

func LessThanFloat(m Money, f float64) bool {
	return m.LessThan(FromFloat(f))
}

func LessThanOrEqualFloat(m Money, f float64) bool {
	return m.LessThanOrEqual(FromFloat(f))
}

func FromInt(i int) Money {
	return decimal.NewFromInt(int64(i))
}

func ToInt(m Money) int {
	return int(m.IntPart())
}

func FromInt64(i int64) Money {
	return decimal.NewFromInt(i)
}

func ToInt64(m Money) int64 {
	return m.IntPart()
}

func GreaterThanInt(m Money, i int) bool {
	return m.GreaterThan(FromInt(i))
}

func GreaterThanOrEqualInt(m Money, i int) bool {
	return m.GreaterThanOrEqual(FromInt(i))
}

func LessThanInt(m Money, i int) bool {
	return m.LessThan(FromInt(i))
}

func LessThanOrEqualInt(m Money, i int) bool {
	return m.LessThanOrEqual(FromInt(i))
}

func IsZero(m Money) bool {
	return m.IsZero()
}

// IsPositive
// 0.0 false
// -0.0 false
// 1.00 true
// 1234.5678 true
func IsPositive(m Money) bool {
	return m.IsPositive()
}

// IsNegative
// 0.0 false
// -0.0 false
// -1.00 true
// -1234.5678 true
func IsNegative(m Money) bool {
	return m.IsNegative()
}

func Round(m Money, places int32) Money {
	return m.Round(places)
}

func Abs(m Money) Money {
	return m.Abs()
}

func Format(m Money, decimalPlaces int32) string {
	return m.StringFixed(decimalPlaces)
}

func Min(a, b Money) Money {
	if a.LessThan(b) {
		return a
	}
	return b
}

func Max(a, b Money) Money {
	if a.GreaterThan(b) {
		return a
	}
	return b
}
