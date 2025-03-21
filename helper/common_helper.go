package helper

import (
	"context"
	"errors"
	"math"
	"regexp"
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

func CopyCtx(c context.Context) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "state", GetState(c))
	return ctx
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

func Chunks[T any](source []T, length int) ([][]T, error) {
	if length < 0 {
		return nil, errors.New("length cannot be less than 0")
	}

	var result [][]T
	size := len(source)
	if size <= 0 {
		return result, nil
	}

	fullChunks := int(math.Ceil(float64(size) / float64(length)))

	for n := 0; n < fullChunks; n++ {
		start := n * length
		end := start + length
		if end > size {
			end = size
		}
		result = append(result, source[start:end])
	}

	return result, nil
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

func SplitNumberBatches(n, length int) [][]int {
	var batches [][]int
	start := 1

	for start <= n {
		end := start + length - 1
		if end > n {
			end = n
		}

		batch := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			batch = append(batch, i)
		}
		batches = append(batches, batch)

		start = end + 1
	}

	return batches
}
