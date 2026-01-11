package types

// VND Money type
type (
	VND = int64
)

// SignedInteger Integer type
type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// SignedUint Uint type
type SignedUint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// SignedFloat Float type
type SignedFloat interface {
	~float32 | ~float64
}
