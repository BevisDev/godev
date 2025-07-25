package utils

import (
	"context"
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils/random"
	"golang.org/x/exp/constraints"
	"slices"
	"strings"
	"time"
)

type M map[string]interface{}

func GetState(ctx context.Context) string {
	if ctx == nil {
		return random.RandUUID()
	}
	state, ok := ctx.Value(consts.State).(string)
	if !ok {
		state = random.RandUUID()
	}
	return state
}

func SetState(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, consts.State, state)
}

func CreateCtx() context.Context {
	return context.WithValue(context.Background(), consts.State, random.RandUUID())
}

func CreateCtxTimeout(ctx context.Context, timeoutSec int) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
}

func CreateCtxCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithCancel(ctx)
}

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

func MaskLeft(s string, size int) string {
	if size <= 0 || size > len(s) {
		size = len(s)
	}
	mask := strings.Repeat("*", size)
	return mask + s[size:]
}

func MaskRight(s string, size int) string {
	if size <= 0 || size > len(s) {
		size = len(s)
	}
	mask := strings.Repeat("*", size)
	return s[:len(s)-size] + mask
}

func MaskCenter(s string, size int) string {
	n := len(s)
	if size <= 0 || size >= n {
		return strings.Repeat("*", n)
	}

	left := (n - size) / 2
	right := n - size - left

	return s[:left] + strings.Repeat("*", size) + s[n-right:]
}

// MaskEmail masks a portion of the local and domain parts of an email address.
//
// The `sizeLocal` and `sizeDomain` parameters specify how many characters to mask
// from the end of the local part and the domain part, respectively.
//
// If `sizeLocal` or `sizeDomain` is greater than the length of their respective parts,
// the entire part will be masked. If either is zero or negative, masking is skipped.
//
// Returns the masked email address. If the input is not a valid email (no "@"), it returns the input unchanged.
//
// Examples:
//
//	MaskEmail("john.doe@example.com", 3, 4)   // "john.***@exam****"
//	MaskEmail("abc@domain.com", 10, 10)       // "***@**********"
//	MaskEmail("a@x.com", 1, 1)                // "*@x.co*"
//	MaskEmail("invalid-email", 3, 3)          // "invalid-email"
func MaskEmail(email string, sizeLocal, sizeDomain int) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	var maskedLocal, maskedDomain string
	local, domain := parts[0], parts[1]

	// --- Mask local ---
	if sizeLocal < 0 || sizeLocal > len(local) {
		sizeLocal = len(local)
	}
	if sizeLocal == 0 {
		maskedLocal = local
	} else {
		maskedLocal = local[:len(local)-sizeLocal] + strings.Repeat("*", sizeLocal)
	}

	// --- Mask domain ---
	if sizeDomain < 0 || sizeDomain > len(domain) {
		sizeDomain = len(domain)
	}
	if sizeDomain == 0 {
		maskedDomain = domain
	} else {
		maskedDomain = domain[:len(domain)-sizeDomain] + strings.Repeat("*", sizeDomain)
	}

	return maskedLocal + "@" + maskedDomain
}

func SkipContentType(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "image"),
		strings.HasPrefix(contentType, "video"),
		strings.HasPrefix(contentType, "audio"),
		strings.HasPrefix(contentType, "application/octet-stream"),
		strings.HasPrefix(contentType, "multipart/form-data"),
		strings.HasPrefix(contentType, "application/pdf"),
		strings.HasPrefix(contentType, "application/zip"),
		strings.HasPrefix(contentType, "application/x-zip-compressed"),
		strings.HasPrefix(contentType, "application/vnd."),
		strings.HasPrefix(contentType, "application/msword"),
		strings.HasPrefix(contentType, "application/x-protobuf"):
		return true
	default:
		return false
	}
}

func Parse[T any](obj interface{}) (T, error) {
	val, ok := obj.(T)
	if !ok {
		return val, fmt.Errorf("cannot cast %T to target type", obj)
	}
	return val, nil
}

func ParseMap[T any](key string, objMap M) (T, error) {
	var zero T

	raw, ok := objMap[key]
	if !ok {
		return zero, fmt.Errorf("key %q not found in map", key)
	}

	val, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("cannot cast value of key %q (type %T) to target type", key, raw)
	}

	return val, nil
}

func IsContains[T comparable](slice []T, value T) bool {
	return slices.Contains(slice, value)
}

func IndexOf[T comparable](slice []T, value T) int {
	return slices.Index(slice, value)
}

// Min returns the smaller of two ordered values a and b.
//
// Works with any comparable type (int, float64, string, etc.).
//
// Example:
//
//	min := Min(3, 7)        // min = 3
//	min := Min("a", "b")    // min = "a"
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two ordered values a and b.
//
// Works with any comparable type (int, float64, string, etc.).
//
// Example:
//
//	max := Max(3, 7)        // max = 7
//	max := Max("a", "b")    // max = "b"
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Percent converts an integer value to its percentage form as a float64.
//
// For example, Percent(65) returns 0.65.
//
// Example:
//
//	rate := Percent(30)     // rate = 0.3
//	tax := Percent(8) * 250 // tax = 20.0
func Percent[T constraints.Integer](n T) float64 {
	return float64(n) / 100
}

// Milli multiplies an integer by one million.
//
// It is useful when you want to convert a base unit into millions.
// For example, Milli(5) returns 5,000,000.
//
// Example:
//
//	n := 5
//	result := Milli(n)
//	// result == 5_000_000
func Milli(n int64) int64 {
	return n * 1_000_000
}

// RoundDownToMul rounds down n to the nearest multiple of "multiple".
// Example:
//
//	RoundDownToMul(47, 5) = 45
//	RoundDownToMul(42_000_000, 5_000_000) = 40_000_000
//	RoundDownToMul(-1, 5)        // = -5
//	RoundDownToMul(-7, 5)        // = -10
//	RoundDownToMul(-13, 5)       // = -15
func RoundDownToMul[T constraints.Integer](n T, mul T) T {
	if n == 0 || mul <= 0 || n%mul == 0 {
		return n
	}
	if n < 0 {
		return ((n / mul) - 1) * mul
	}
	return (n / mul) * mul
}

// RoundUpToMul rounds up n to the nearest multiple of mul.
// Example:
//
//	RoundUpToMul(47, 5)        // = 50
//	RoundUpToMul(41, 10)       // = 50
//	RoundUpToMul(40, 10)       // = 40
//	RoundUpToMul(42_000_000, 5_000_000) // = 45_000_000
//	RoundUpToMul(-7, 5)        // = -5
//	RoundUpToMul(-11, 5)       // = -10
//	RoundUpToMul(-4, 5)        // = 0
//	RoundUpToMul(-5, 5)        // = -5
func RoundUpToMul[T constraints.Integer](n T, mul T) T {
	if n == 0 || mul <= 0 || n%mul == 0 {
		return n
	}
	if n < 0 {
		return (n / mul) * mul
	}
	return ((n / mul) + 1) * mul
}

// PtrTo returns a pointer to the given value.
//
// Useful in tests and code where you want to pass a pointer literal.
//
// Example:
//
//	s := ptrTo("hello")  // *string → "hello"
//	n := ptrTo(123)      // *int → 123
func PtrTo[T any](v T) *T {
	return &v
}
