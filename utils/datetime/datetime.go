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

// ToString formats a time.Time to string using the specified layout.
//
// Example:
//
//	ToString(time.Now(), "2006-01-02 15:04:05")
func ToString(time time.Time, format string) string {
	return time.Format(format)
}

// ToTime parses a string into a time.Time pointer using the specified layout.
//
// Example:
//
//	t, err := ToTime("2024-01-02", "2006-01-02")
func ToTime(timeStr string, format string) (*time.Time, error) {
	parsedTime, err := time.Parse(format, timeStr)
	if err != nil {
		return nil, err
	}
	return &parsedTime, nil
}

// BeginDay returns a time.Time representing the start of the day (00:00:00) in the same location.
func BeginDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

// EndDay returns a time.Time representing the end of the day (23:59:59.999999000) in the same location.
func EndDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999000, date.Location())
}

// AddTime adds a specified amount of time to the given date based on the unit kind.
// Supported kinds: Nanosecond, Millisecond, Second, Minute, Hour, Day, Month, Year.
//
// Example:
//
//	AddTime(time.Now(), 3, Day)
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

// IsSameDate returns true if t1 and t2 have the same year, month, and day.
func IsSameDate(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}

// IsWithinDays returns true if the given time is within the past N days from now.
func IsWithinDays(t time.Time, day int) bool {
	return time.Since(t) <= time.Duration(day)*24*time.Hour
}

// DaysBetween returns the absolute number of full days between two dates.
func DaysBetween(t1, t2 time.Time) int {
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return int(diff.Hours()) / 24
}

// StartOfMonth returns a time.Time representing the start of the month (the first day at midnight).
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns a time.Time representing the last day of the month at midnight.
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1)
}

// StartOfWeek returns the start of the week (Monday 00:00:00) for the given date.
//
// Logic:
// - Go's time.Weekday() returns 0 for Sunday, 1 for Monday, ..., 6 for Saturday.
// - If weekday == 0 (Sunday), we map it to 7 to consider Sunday as the last day of the week.
// - Subtract (weekday - 1) days to get back to Monday.
// - Finally, set time to 00:00:00.
func StartOfWeek(t time.Time) time.Time {
	// Get numeric weekday (0=Sunday, 1=Monday, ...)
	weekday := int(t.Weekday())

	// Treat Sunday as 7 so Monday=1, Sunday=7
	if weekday == 0 {
		weekday = 7
	}

	// Subtract days to go back to Monday
	start := AddTime(t, -(weekday - 1), Day)

	// Return the date set to 00:00:00
	return BeginDay(start)
}

// EndOfWeek returns the end of the week (Sunday 23:59:59.999999) for the given time.
//
// The week is considered to start on Monday and end on Sunday.
//
// Example:
//
//	t := time.Date(2024, 7, 10, 15, 30, 0, 0, time.UTC) // Wednesday
//	end := EndOfWeek(t) // Sunday 2024-07-14 23:59:59.999999 UTC
func EndOfWeek(t time.Time) time.Time {
	// Get the start of the week (Monday 00:00:00)
	startOfWeek := StartOfWeek(t)

	// Add 6 days to reach Sunday
	end := AddTime(startOfWeek, 6, Day)

	// Set time to end of day (23:59:59.999999)
	return EndDay(end)
}

// IsWeekend returns true if the given time falls on Saturday or Sunday.
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsToday returns true if the given time t falls on today's date
// (same year, month, and day as the current system date).
//
// Example:
//
//	if IsToday(time.Now()) {
//	    fmt.Println("This is today!")
//	}
func IsToday(t time.Time) bool {
	now := time.Now()
	return IsSameDate(t, now)
}

// IsYesterday returns true if the given time t falls on yesterday's date
// (one day before the current system date).
//
// Example:
//
//	y := time.Now().AddDate(0, 0, -1)
//	if IsYesterday(y) {
//	    fmt.Println("This was yesterday.")
//	}
func IsYesterday(t time.Time) bool {
	now := time.Now()
	yesterday := AddTime(now, -1, Day)
	return IsSameDate(t, yesterday)
}

// GetTimestamp returns the current Unix timestamp in milliseconds.
func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

// CalculateAge computes the age in years based on a date of birth and a reference date.
//
// If the reference date (calcTime) is before the birthday in the current year,
// the age is decreased by 1.
//
// Example:
//
//	dob := time.Date(1990, 7, 10, 0, 0, 0, 0, time.UTC)
//	ref := time.Date(2024, 7, 9, 0, 0, 0, 0, time.UTC)
//	age := CalculateAge(dob, ref) // returns 33
//
// Parameters:
//   - dob: date of birth
//   - calcTime: date to calculate the age at
//
// Returns:
//   - The integer age in years
func CalculateAge(dob, calcTime time.Time) int {
	age := calcTime.Year() - dob.Year()

	if calcTime.Month() < dob.Month() ||
		(calcTime.Month() == dob.Month() && calcTime.Day() < dob.Day()) {
		age--
	}
	return age
}
