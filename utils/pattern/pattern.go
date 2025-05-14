package pattern

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	Email         = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	TenDigitPhone = `^\d{10}$`
	UUID          = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`
	AlphaNumeric  = `^[a-zA-Z0-9]+$`
	DateYYYYMMDD  = `^\d{4}-\d{2}-\d{2}$`
	TimeHHMMSS    = `^\d{2}:\d{2}:\d{2}$`
	IPv4          = `^(\d{1,3}\.){3}\d{1,3}$`
	HexColor      = `^#?([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`
	VNIDNumber    = `^\d{9}(\d{3})?$`
	SafeFile      = `^[\w,\s-]+\.[A-Za-z]{1,4}$`
)

func Matches(s, pattern string) bool {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return matched
}

func IsEmail(s string) bool {
	return Matches(s, Email)
}

func IsPhoneNumber(s string) bool {
	return Matches(s, TenDigitPhone)
}

func IsUUID(s string) bool {
	return Matches(s, UUID)
}

func IsDate(s string) bool {
	return Matches(s, DateYYYYMMDD)
}

func IsTime(s string) bool {
	return Matches(s, TimeHHMMSS)
}

func IsIPv4(s string) bool {
	matched := Matches(s, IPv4)
	if !matched {
		return false
	}

	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}

func IsAlphaNumeric(s string) bool {
	return Matches(s, AlphaNumeric)
}

func IsHexColor(s string) bool {
	return Matches(s, HexColor)
}

func IsVietnamID(s string) bool {
	return Matches(s, VNIDNumber)
}

func IsStrongPassword(s string) bool {
	if len(s) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r), unicode.IsSymbol(r):
			hasSpecial = true
		}

		if hasUpper && hasLower && hasDigit && hasSpecial {
			return true
		}
	}

	return false
}

func IsSafeFileName(s string) bool {
	return Matches(s, SafeFile)
}

func ExtractAllMatches(s, pattern string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindAllString(s, -1)
}

func CompileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}
