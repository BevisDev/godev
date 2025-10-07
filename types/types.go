package types

import (
	"github.com/shopspring/decimal"
)

// Money type
type (
	Money = decimal.Decimal
	VND   = int64
)

// Integer type
type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Uint type
type SignedUint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Float type
type SignedFloat interface {
	~float32 | ~float64
}
