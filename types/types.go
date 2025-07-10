package types

import (
	"github.com/shopspring/decimal"
)

// Money type
type (
	Money = decimal.Decimal
	VND   = int64
)

// DB type
type (
	KindDB         int
	DBJSONTemplate string
)

// Integer type
type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Float type
type SignedFloat interface {
	~float32 | ~float64
}
