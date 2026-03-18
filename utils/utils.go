package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/types"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/random"
	"github.com/BevisDev/godev/utils/str"
	"golang.org/x/exp/constraints"
)

func NewCtxWithRequest(r context.Context) context.Context {
	var rid = GetRID(r)
	ctx := NewCtx()
	return SetValueCtx(ctx, consts.RID, rid)
}

func NewCtx() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, consts.RID, random.NewUUID())
}

func SetValueCtx(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetRID(ctx context.Context) string {
	if ctx == nil {
		return random.NewUUID()
	}

	rid, ok := ctx.Value(consts.RID).(string)
	if !ok {
		rid = random.NewUUID()
	}

	return rid
}

func NewCtxTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

func NewCtxCancel(ctx context.Context) (context.Context, context.CancelFunc) {
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
		strings.HasPrefix(contentType, "application/vnd."),
		strings.HasPrefix(contentType, "application/x-protobuf"),
		strings.HasPrefix(contentType, consts.ApplicationOctetStream),
		strings.HasPrefix(contentType, consts.MultipartFormData),
		strings.HasPrefix(contentType, consts.ApplicationPDF),
		strings.HasPrefix(contentType, consts.ApplicationMSWord),
		strings.HasPrefix(contentType, consts.ApplicationZip),
		strings.HasPrefix(contentType, consts.ApplicationX7z),
		strings.HasPrefix(contentType, consts.ApplicationXZip):
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

func ParseValueMap[T any](key string, objMap types.Object) (T, error) {
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

// A Million multiplies an integer by one million.
//
// It is useful when you want to convert a base unit into millions.
// For example, Million(5) returns 5,000,000.
//
// Example:
//
//	n := 5
//	result := Million(n)
//	// result == 5_000_000
func Million(n int64) int64 {
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

// GetPointer returns a pointer to the given value.
//
// Useful in tests and code where you want to pass a pointer literal.
//
// Example:
//
//	s := ptrTo("hello")  // *string → "hello"
//	n := ptrTo(123)      // *int → 123
func GetPointer[T any](v T) *T {
	return &v
}

// ValueFromPointer returns the value pointed to by v.
// If v is nil, it returns the zero value of type T.
// Safe to use with any pointer type without causing panic.
func ValueFromPointer[T any](v *T) T {
	var zero T
	if v == nil {
		return zero
	}
	return *v
}

func ToBytes(value any) ([]byte, error) {
	if value == nil {
		return nil, errors.New("value is nil")
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if rv.IsNil() {
			return nil, errors.New("value is nil")
		}
	}

	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return []byte(str.ToString(v)), nil
	default:
		body, err := jsonx.ToJSONBytes(v)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}

// ValueFromBytes decodes bytes into type T.
// It tries JSON unmarshal first; if that fails and T is string, the raw bytes are returned as the string.
// Empty data: returns zero string for T string, empty slice for T []byte, error for other types.
func ValueFromBytes[T any](data []byte) (T, error) {
	var zero T
	if len(data) == 0 {
		if _, ok := any(zero).(string); ok {
			return any("").(T), nil
		}
		if _, ok := any(zero).([]byte); ok {
			return any([]byte(nil)).(T), nil
		}
		return zero, errors.New("empty data")
	}

	rt := reflect.TypeOf(zero)
	if rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Uint8 {
		out := make([]byte, len(data))
		copy(out, data)
		return any(out).(T), nil
	}

	// Try JSON first (so "world" decodes to world for T string)
	t, err := jsonx.FromJSONBytes[T](data)
	if err == nil {
		return t, nil
	}
	// If JSON fails and T is string, use raw bytes
	if rt.Kind() == reflect.String {
		return any(string(data)).(T), nil
	}
	return zero, err
}

// ValueFromString converts a raw string to T.
// For built-in string and []byte returns directly without JSON; other types use ValueFromBytes.
func ValueFromString[T any](raw string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case string:
		return any(raw).(T), nil
	case []byte:
		if raw == "" {
			return any([]byte(nil)).(T), nil
		}
		return any([]byte(raw)).(T), nil
	default:
		return ValueFromBytes[T]([]byte(raw))
	}
}

// ValueFromAny converts an interface{} to T.
// Supported input:
//   - nil           → return zero T
//   - string        → direct for T=string, []byte; JSON decode otherwise
//   - []byte        → direct for T=[]byte, string; JSON decode otherwise
//   - other types   → JSON marshal → JSON unmarshal into T
func ValueFromAny[T any](v any) (T, error) {
	var zero T

	if v == nil {
		return zero, nil
	}

	switch val := v.(type) {

	case T:
		return val, nil

	case string:
		switch any(zero).(type) {
		case string:
			return any(val).(T), nil
		case []byte:
			return any([]byte(val)).(T), nil
		default:
			return ValueFromBytes[T]([]byte(val))
		}

	case []byte:
		switch any(zero).(type) {
		case []byte:
			return any(val).(T), nil
		case string:
			return any(string(val)).(T), nil
		default:
			return ValueFromBytes[T](val)
		}

	default:
		b, err := ToBytes(val)
		if err != nil {
			return zero, err
		}
		return ValueFromBytes[T](b)
	}
}

// ToFloat converts common numeric types (and numeric strings) into float64.
// Returns an error for unsupported types or invalid strings.
func ToFloat(v any) (float64, error) {
	switch t := v.(type) {
	case float32:
		return float64(t), nil
	case float64:
		return t, nil
	case int:
		return float64(t), nil
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
		return float64(t), nil
	case uint8:
		return float64(t), nil
	case uint16:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", v)
	}
}

// ToInt64 converts common numeric types (and numeric strings) into int64.
// Floats are truncated toward zero; uint64 values > MaxInt64 return an error.
func ToInt64(v any) (int64, error) {
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		if t > math.MaxInt64 {
			return 0, fmt.Errorf("overflow uint64 -> int64")
		}
		return int64(t), nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err == nil {
			return i, nil
		}

		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, err
		}

		return int64(f), nil

	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

func ToInt(v any) (int, error) {
	i64, err := ToInt64(v)
	if err != nil {
		return 0, err
	}

	if i64 > int64(math.MaxInt) || i64 < int64(math.MinInt) {
		return 0, fmt.Errorf("overflow int")
	}

	return int(i64), nil
}

func ToBool(v any) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case int:
		return t != 0, nil
	case int8:
		return t != 0, nil
	case int16:
		return t != 0, nil
	case int32:
		return t != 0, nil
	case int64:
		return t != 0, nil
	case uint:
		return t != 0, nil
	case uint64:
		return t != 0, nil
	case float32:
		return t != 0, nil
	case float64:
		return t != 0, nil
	case string:
		b, err := strconv.ParseBool(t)
		if err != nil {
			return false, err
		}
		return b, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// ToSlice converts a few common slice types into []any.
// Non-slice inputs are wrapped as a single-element slice.
func ToSlice(v any) []any {
	switch t := v.(type) {
	case []any:
		return t

	case []string:
		out := make([]any, len(t))
		for i, v := range t {
			out[i] = v
		}
		return out

	case []int:
		out := make([]any, len(t))
		for i, v := range t {
			out[i] = v
		}
		return out

	case []int64:
		out := make([]any, len(t))
		for i, v := range t {
			out[i] = v
		}
		return out

	case []float64:
		out := make([]any, len(t))
		for i, v := range t {
			out[i] = v
		}
		return out

	default:
		return []any{v}
	}
}
