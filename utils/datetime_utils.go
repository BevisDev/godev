package utils

import (
	"time"
)

// datetime
const (
	YYYY_MM_DD       = "2006-01-02"
	DD_MM_YYYY       = "02-01-2006"
	DD_MM_YYYY_SLASH = "02-01-2006"
	DD_MMM_YYYY      = "02-Jan-2006"
	YYYYMMDDHHMMSS   = "20060102150405"

	// Date time ISO 8601
	DATETIME_FULL      = "2006-01-02T15:04:05Z07:00"
	DATETIME_NO_OFFSET = "2006-01-02T15:04:05Z"
	DATETIME_NO_TZ     = "2006-01-02 15:04:05.000"

	// format time
	TIME_FULL   = "150405"
	TIME_NO_SEC = "1504"

	// string time
	Second = "Second"
	Minute = "Minute"
	Hour   = "Hour"
	Day    = "Day"
	Month  = "Month"
	Year   = "Year"
)

func TimeToString(time time.Time, format string) string {
	return time.Format(format)
}

func StringToTime(timeStr string, format string) (*time.Time, error) {
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
