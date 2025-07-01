package money

import (
	"github.com/BevisDev/godev/types"
	"github.com/shopspring/decimal"
)

func FromFloat(f float64) types.Money {
	return decimal.NewFromFloat(f)
}

func ToFloat(m types.Money) float64 {
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

func GreaterThanFloat(m types.Money, f float64) bool {
	return m.GreaterThan(FromFloat(f))
}

func GreaterThanOrEqualFloat(m types.Money, f float64) bool {
	return m.GreaterThanOrEqual(FromFloat(f))
}

func LessThanFloat(m types.Money, f float64) bool {
	return m.LessThan(FromFloat(f))
}

func LessThanOrEqualFloat(m types.Money, f float64) bool {
	return m.LessThanOrEqual(FromFloat(f))
}

func FromInt(i int) types.Money {
	return decimal.NewFromInt(int64(i))
}

func ToInt(m types.Money) int {
	return int(m.IntPart())
}

func FromInt64(i int64) types.Money {
	return decimal.NewFromInt(i)
}

func ToInt64(m types.Money) int64 {
	return m.IntPart()
}

func GreaterThanInt(m types.Money, i int) bool {
	return m.GreaterThan(FromInt(i))
}

func GreaterThanOrEqualInt(m types.Money, i int) bool {
	return m.GreaterThanOrEqual(FromInt(i))
}

func LessThanInt(m types.Money, i int) bool {
	return m.LessThan(FromInt(i))
}

func LessThanOrEqualInt(m types.Money, i int) bool {
	return m.LessThanOrEqual(FromInt(i))
}

func IsZero(m types.Money) bool {
	return m.IsZero()
}

// IsPositive
// 0.0 false
// -0.0 false
// 1.00 true
// 1234.5678 true
func IsPositive(m types.Money) bool {
	return m.IsPositive()
}

// IsNegative
// 0.0 false
// -0.0 false
// -1.00 true
// -1234.5678 true
func IsNegative(m types.Money) bool {
	return m.IsNegative()
}

func Round(m types.Money, places int32) types.Money {
	return m.Round(places)
}

func Abs(m types.Money) types.Money {
	return m.Abs()
}

func Format(m types.Money, decimalPlaces int32) string {
	return m.StringFixed(decimalPlaces)
}

func Min(a, b types.Money) types.Money {
	if a.LessThan(b) {
		return a
	}
	return b
}

func Max(a, b types.Money) types.Money {
	if a.GreaterThan(b) {
		return a
	}
	return b
}
