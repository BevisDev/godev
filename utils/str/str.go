package str

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

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

func ToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func ToFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return f
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
