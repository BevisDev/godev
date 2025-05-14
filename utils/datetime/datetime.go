package datetime

import (
	"time"
)

const (
	// format common
	DD_MM_YYYYY       = "02-01-2006"
	DD_MM_YYYYY_FLASH = "02/01/2006"
	DD_MMM_YYYY       = "02-Jan-2006"
	YYYYMMDDHHMMSS    = "20060102150405"
	YYYYMMDD          = "20060102"

	// format ISO 8601 / RFC3339
	DateOnly       = "2006-01-02"
	DateTime       = "2006-01-02 15:04:05"
	DateTimeOffset = "2006-01-02T15:04:05Z07:00"
	DatetimeUTC    = "2006-01-02T15:04:05Z"
	DateTimeSQL    = "2006-01-02 15:04:05.000"
	DateTimeNoTZ   = "2006-01-02T15:04:05"

	// format time
	TimeOnly    = "15:04:05"
	TimeCompact = "150405"
	TimeNoSec   = "1504"

	// using for AddTime
	Nanosecond  = "Nanosecond"
	Millisecond = "Millisecond"
	Second      = "Second"
	Minute      = "Minute"
	Hour        = "Hour"
	Day         = "Day"
	Month       = "Month"
	Year        = "Year"
)

func ToString(time time.Time, format string) string {
	return time.Format(format)
}

func ToTime(timeStr string, format string) (*time.Time, error) {
	parsedTime, err := time.Parse(format, timeStr)
	if err != nil {
		return nil, err
	}
	return &parsedTime, nil
}

func BeginDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func EndDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999000, date.Location())
}

func AddTime(date time.Time, v int, kind string) time.Time {
	switch kind {
	case Nanosecond:
		return date.Add(time.Duration(v) * time.Nanosecond)
	case Millisecond:
		return date.Add(time.Duration(v) * time.Millisecond)
	case Second:
		return date.Add(time.Duration(v) * time.Second)
	case Minute:
		return date.Add(time.Duration(v) * time.Minute)
	case Hour:
		return date.Add(time.Duration(v) * time.Hour)
	case Day:
		return date.AddDate(0, 0, v)
	case Month:
		return date.AddDate(0, v, 0)
	case Year:
		return date.AddDate(v, 0, 0)
	default:
		return date
	}
}

func IsSameDate(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}

func IsWithinDays(t time.Time, day int) bool {
	return time.Since(t) <= time.Duration(day)*24*time.Hour
}

func DaysBetween(t1, t2 time.Time) int {
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return int(diff.Hours()) / 24
}

func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1)
}

func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}
