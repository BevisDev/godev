package helper

import (
	"context"
	"fmt"
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
	state, ok := ctx.Value("state").(string)
	if !ok {
		state = GenUUID()
	}
	return state
}

func CreateCtx(state string) context.Context {
	ctx := context.Background()
	if state == "" {
		state = GenUUID()
	}
	ctx = context.WithValue(ctx, "state", state)
	return ctx
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
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', 2, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func RemoveAccent(str string) string {
	result := norm.NFD.String(str)
	var output []rune
	for _, r := range result {
		if unicode.Is(unicode.M, r) {
			continue
		}
		output = append(output, r)
	}
	return string(output)
}

func RemoveSpecialChars(str string) string {
	o := RemoveAccent(str)
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	return re.ReplaceAllString(o, "")
}

func RemoveWhiteSpace(str string) string {
	return strings.ReplaceAll(str, " ", "")
}
