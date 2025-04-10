package utils

import (
	"context"
	"fmt"
	"github.com/BevisDev/godev/constants"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

func GenUUID() string {
	return uuid.NewString()
}

func GetState(ctx context.Context) string {
	if ctx == nil {
		return GenUUID()
	}
	state, ok := ctx.Value(constants.State).(string)
	if !ok {
		state = GenUUID()
	}
	return state
}

func CreateCtx() context.Context {
	return context.WithValue(context.Background(), constants.State, GenUUID())
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

func ToString(value any) string {
	if value == nil {
		return ""
	}
	val := reflect.ValueOf(value)

	// handle ptr
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), 'g', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return string(val.Bytes()) // handle []byte
		}
	default:
		return fmt.Sprintf("%+v", val.Interface())
	}
	return fmt.Sprintf("%+v", val.Interface())
}

func NormalizeToASCII(str string) string {
	result := norm.NFD.String(str)
	var output []rune
	for _, r := range result {
		if unicode.Is(unicode.M, r) {
			continue
		}
		output = append(output, r)
	}
	normalized := string(output)
	normalized = strings.ReplaceAll(normalized, "Đ", "D")
	normalized = strings.ReplaceAll(normalized, "đ", "d")
	return normalized
}

func CleanText(str string) string {
	o := NormalizeToASCII(str)
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	return re.ReplaceAllString(o, "")
}

func RemoveWhiteSpace(str string) string {
	return strings.ReplaceAll(str, " ", "")
}

func TruncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}
