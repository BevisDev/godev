package datetime

import (
	"time"
)

// Layout datetime
type Layout = string

const (
	// format common
	DD_MM_YYYYY       Layout = "02-01-2006"
	DD_MM_YYYYY_FLASH Layout = "02/01/2006"
	DD_MMM_YYYY       Layout = "02-Jan-2006"
	YYYYMMDDHHMMSS    Layout = "20060102150405"
	YYYYMMDD          Layout = "20060102"

	// format ISO 8601 / RFC3339
	DateOnly       Layout = "2006-01-02"
	DateTime       Layout = "2006-01-02 15:04:05"
	DateTimeOffset Layout = "2006-01-02T15:04:05Z07:00"
	DatetimeUTC    Layout = "2006-01-02T15:04:05Z"
	DateTimeSQL    Layout = "2006-01-02 15:04:05.000"

	// format time
	TimeOnly    Layout = "15:04:05"
	TimeCompact Layout = "150405"
	TimeNoSec   Layout = "1504"

	// AddTime layout
	Second Layout = "Second"
	Minute Layout = "Minute"
	Hour   Layout = "Hour"
	Day    Layout = "Day"
	Month  Layout = "Month"
	Year   Layout = "Year"
)

func ToString(time time.Time, format Layout) string {
	return time.Format(format)
}

func ToTime(timeStr string, format Layout) (*time.Time, error) {
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

func AddTime(date time.Time, v int, kind Layout) time.Time {
	switch kind {
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
