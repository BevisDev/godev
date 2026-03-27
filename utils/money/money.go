package money

import (
	"fmt"
	"math"
	"reflect"

	"github.com/shopspring/decimal"
)

type Money = decimal.Decimal

// FromFloat creates Money from a float64 value.
func FromFloat(f float64) Money {
	return decimal.NewFromFloat(f)
}

// ToFloat converts Money to float64.
func ToFloat(m Money) float64 {
	f, _ := m.Float64()
	return f
}

// FromString creates Money from a decimal string.
func FromString(s string) (Money, error) {
	if s == "" {
		return decimal.Zero, nil
	}

	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid decimal string: %s", s)
	}
	return d, nil
}

// ToDecimal converts common numeric/string inputs to Money.
// It returns decimal.Zero for nil and an error for unsupported values.
func ToDecimal(val interface{}) (Money, error) {
	if val == nil {
		return decimal.Zero, nil
	}

	switch v := val.(type) {
	case decimal.Decimal:
		return v, nil
	case *decimal.Decimal:
		if v == nil {
			return decimal.Zero, nil
		}
		return *v, nil
	case float64:
		return FromFloat(v), nil
	case float32:
		return FromFloat(float64(v)), nil
	case int:
		return FromInt(v), nil
	case int64:
		return FromInt64(v), nil
	case int32:
		return FromInt64(int64(v)), nil
	case uint:
		if uint64(v) > math.MaxInt64 {
			return decimal.Zero, fmt.Errorf("overflow uint -> int64")
		}
		return FromInt64(int64(v)), nil
	case uint64:
		if v > math.MaxInt64 {
			return decimal.Zero, fmt.Errorf("overflow uint64 -> int64")
		}
		return FromInt64(int64(v)), nil
	case string:
		return FromString(v)
	default:
		return decimal.Zero, fmt.Errorf("unsupported type: %T", v)
	}
}

// IsDecimal reports whether val is a decimal.Decimal or *decimal.Decimal.
func IsDecimal(val interface{}) bool {
	switch val.(type) {
	case decimal.Decimal, *decimal.Decimal:
		return true
	default:
		return false
	}
}

// GreaterThanFloat reports whether m is greater than f.
func GreaterThanFloat(m Money, f float64) bool {
	return m.GreaterThan(FromFloat(f))
}

// GreaterThanOrEqualFloat reports whether m is greater than or equal to f.
func GreaterThanOrEqualFloat(m Money, f float64) bool {
	return m.GreaterThanOrEqual(FromFloat(f))
}

// LessThanFloat reports whether m is less than f.
func LessThanFloat(m Money, f float64) bool {
	return m.LessThan(FromFloat(f))
}

// LessThanOrEqualFloat reports whether m is less than or equal to f.
func LessThanOrEqualFloat(m Money, f float64) bool {
	return m.LessThanOrEqual(FromFloat(f))
}

// FromInt creates Money from an int value.
func FromInt(i int) Money {
	return decimal.NewFromInt(int64(i))
}

// ToInt converts Money to int by taking the integer part.
func ToInt(m Money) int {
	return int(m.IntPart())
}

// FromInt64 creates Money from an int64 value.
func FromInt64(i int64) Money {
	return decimal.NewFromInt(i)
}

// ToInt64 converts Money to int64 by taking the integer part.
func ToInt64(m Money) int64 {
	return m.IntPart()
}

// ToString returns the canonical string form of Money.
func ToString(m Money) string {
	return m.String()
}

// InSlice reports whether target exists in list.
func InSlice(target Money, list []Money) bool {
	for _, item := range list {
		if item.Equal(target) {
			return true
		}
	}
	return false
}

// ToMoneySlice converts an interface value to []Money.
// It accepts slices/arrays (including pointers) whose elements are supported by ToDecimal.
func ToMoneySlice(val interface{}) ([]Money, error) {
	if val == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("unsupported type: %T", val)
	}

	n := rv.Len()
	if n == 0 {
		return []Money{}, nil
	}

	out := make([]Money, n)
	for i := range out {
		item := rv.Index(i).Interface()
		m, err := ToDecimal(item)
		if err != nil {
			return nil, fmt.Errorf("invalid element at index %d: %w", i, err)
		}
		out[i] = m
	}

	return out, nil
}

// GreaterThanInt reports whether m is greater than i.
func GreaterThanInt(m Money, i int) bool {
	return m.GreaterThan(FromInt(i))
}

// GreaterThanOrEqualInt reports whether m is greater than or equal to i.
func GreaterThanOrEqualInt(m Money, i int) bool {
	return m.GreaterThanOrEqual(FromInt(i))
}

// LessThanInt reports whether m is less than i.
func LessThanInt(m Money, i int) bool {
	return m.LessThan(FromInt(i))
}

// LessThanOrEqualInt reports whether m is less than or equal to i.
func LessThanOrEqualInt(m Money, i int) bool {
	return m.LessThanOrEqual(FromInt(i))
}

// IsZero reports whether m equals zero.
func IsZero(m Money) bool {
	return m.IsZero()
}

// IsPositive reports whether m is strictly greater than zero.
func IsPositive(m Money) bool {
	return m.IsPositive()
}

// IsNegative reports whether m is strictly less than zero.
func IsNegative(m Money) bool {
	return m.IsNegative()
}

// Round rounds m to the given number of decimal places.
func Round(m Money, places int32) Money {
	return m.Round(places)
}

// Abs returns the absolute value of m.
func Abs(m Money) Money {
	return m.Abs()
}

// Format returns m formatted with fixed decimal places.
func Format(m Money, decimalPlaces int32) string {
	return m.StringFixed(decimalPlaces)
}

// Min returns the smaller value between a and b.
func Min(a, b Money) Money {
	if a.LessThan(b) {
		return a
	}
	return b
}

// Max returns the greater value between a and b.
func Max(a, b Money) Money {
	if a.GreaterThan(b) {
		return a
	}
	return b
}
