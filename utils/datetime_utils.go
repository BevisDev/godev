package utils

import (
	"time"
)

// Layout datetime
type Layout = string

const (
	YYYY_MM_DD       Layout = "2006-01-02"
	DD_MM_YYYY       Layout = "02-01-2006"
	DD_MM_YYYY_SLASH Layout = "02-01-2006"
	DD_MMM_YYYY      Layout = "02-Jan-2006"
	YYYYMMDDHHMMSS   Layout = "20060102150405"

	// Date time ISO 8601
	DATETIME_FULL      Layout = "2006-01-02T15:04:05Z07:00"
	DATETIME_NO_OFFSET Layout = "2006-01-02T15:04:05Z"
	DATETIME_NO_TZ     Layout = "2006-01-02 15:04:05.000"

	// format time
	TIME_FULL   Layout = "150405"
	TIME_NO_SEC Layout = "1504"

	// string time
	Second Layout = "Second"
	Minute Layout = "Minute"
	Hour   Layout = "Hour"
	Day    Layout = "Day"
	Month  Layout = "Month"
	Year   Layout = "Year"
)

func TimeToString(time time.Time, format Layout) string {
	return time.Format(format)
}

func StringToTime(timeStr string, format Layout) (*time.Time, error) {
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
