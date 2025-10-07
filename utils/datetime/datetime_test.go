package datetime

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimeToString(t *testing.T) {
	tm := time.Date(2024, 4, 10, 12, 30, 45, 0, time.UTC)
	expected := "2024-04-10 12:30:45"

	result := ToString(tm, DateTime)
	if result != expected {
		t.Errorf("ToString failed, expected %s, got %s", expected, result)
	}
}

func TestStringToTime(t *testing.T) {
	input := "2024-04-10 12:30:45"
	expected := time.Date(2024, 4, 10, 12, 30, 45, 0, time.UTC)

	tm, err := ToTime(input, DateTime)
	if err != nil {
		t.Fatalf("ToTime returned unexpected error: %v", err)
	}

	if !tm.Equal(expected) {
		t.Errorf("ToTime failed, expected %v, got %v", expected, tm)
	}
}

func TestStringToTime_InvalidFormat(t *testing.T) {
	input := "invalid time"
	_, err := ToTime(input, DateTime)

	if err == nil {
		t.Error("expected error for invalid time string, got nil")
	}
}

func TestBeginDay(t *testing.T) {
	date := time.Date(2024, 4, 10, 14, 30, 45, 999999999, time.UTC)
	begin := BeginDay(date)
	expected := time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC)

	if !begin.Equal(expected) {
		t.Errorf("BeginDay = %v; want %v", begin, expected)
	}
}

func TestEndDay(t *testing.T) {
	date := time.Date(2024, 4, 10, 14, 30, 45, 0, time.UTC)
	end := EndDay(date)
	expected := time.Date(2024, 4, 10, 23, 59, 59, 999999000, time.UTC)

	if !end.Equal(expected) {
		t.Errorf("EndDay = %v; want %v", end, expected)
	}
}

func TestAddTime(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		unit     Unit
		value    int
		expected time.Time
	}{
		{"Nanosecond", Nanosecond, 10, base.Add(10 * time.Nanosecond)},
		{"Millisecond", Millisecond, 10, base.Add(10 * time.Millisecond)},
		{"Second", Second, 10, base.Add(10 * time.Second)},
		{"Minute", Minute, 5, base.Add(5 * time.Minute)},
		{"Hour", Hour, 2, base.Add(2 * time.Hour)},
		{"Day", Day, 3, base.AddDate(0, 0, 3)},
		{"Month", Month, 2, base.AddDate(0, 2, 0)},
		{"Year", Year, 1, base.AddDate(1, 0, 0)},
		{"invalid", 0, 999, base},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddTime(base, tt.value, tt.unit)
			if !result.Equal(tt.expected) {
				t.Errorf("AddTime kind=%q = %v; want %v", tt.unit, result, tt.expected)
			}
		})
	}
}

func TestIsSameDate(t *testing.T) {
	t1 := time.Date(2025, 5, 12, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 5, 12, 22, 59, 59, 0, time.UTC)
	t3 := time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC)

	if !IsSameDate(t1, t2) {
		t.Errorf("Expected t1 and t2 to be same date")
	}
	if IsSameDate(t1, t3) {
		t.Errorf("Expected t1 and t3 to be different date")
	}
}

func TestIsWithin(t *testing.T) {
	now := time.Now()
	threeDaysAgo := now.AddDate(0, 0, -3)
	tenDaysAgo := now.AddDate(0, 0, -10)

	if !IsWithinDays(threeDaysAgo, 5) {
		t.Errorf("Expected threeDaysAgo to be within 5 days")
	}
	if IsWithinDays(tenDaysAgo, 5) {
		t.Errorf("Expected tenDaysAgo to be outside 5 days")
	}
}

func TestDaysBetween(t *testing.T) {
	t1 := time.Date(2025, 5, 10, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 5, 12, 0, 0, 0, 0, time.UTC)

	if got := DaysBetween(t1, t2); got != 2 {
		t.Errorf("Expected 2 days between, got %d", got)
	}

	if got := DaysBetween(t2, t1); got != 2 {
		t.Errorf("Expected 2 days between (reversed), got %d", got)
	}
}

func TestStartOfMonth(t *testing.T) {
	t1 := time.Date(2025, 5, 12, 15, 30, 0, 0, time.UTC)
	start := StartOfMonth(t1)
	expected := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)

	if !start.Equal(expected) {
		t.Errorf("Expected start of month %v, got %v", expected, start)
	}
}

func TestEndOfMonth(t *testing.T) {
	t1 := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
	end := EndOfMonth(t1)
	expected := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)

	if !end.Equal(expected) {
		t.Errorf("Expected end of month %v, got %v", expected, end)
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected bool
	}{
		{time.Date(2025, 5, 17, 0, 0, 0, 0, time.UTC), true},  // Saturday
		{time.Date(2025, 5, 18, 0, 0, 0, 0, time.UTC), true},  // Sunday
		{time.Date(2025, 5, 19, 0, 0, 0, 0, time.UTC), false}, // Monday
	}

	for _, tt := range tests {
		if result := IsWeekend(tt.input); result != tt.expected {
			t.Errorf("IsWeekend(%v) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}

func TestGetTimestamp(t *testing.T) {
	ts := GetTimestamp()
	assert.Greater(t, ts, int64(0))
}

func TestCalcAgeAt(t *testing.T) {
	tests := []struct {
		dob      string
		now      string
		expected int
		name     string
	}{
		{"2000-04-20", "2025-04-21", 25, "Birthday passed this year"},
		{"2000-05-10", "2025-04-21", 24, "Birthday not yet this year"},
		{"2000-04-21", "2025-04-21", 25, "Birthday is today"},
		{"2025-04-21", "2025-04-21", 0, "Born today"},
		{"2026-01-01", "2025-04-21", -1, "Future date"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dob := mustParse(tc.dob)
			now := mustParse(tc.now)
			age := CalculateAge(dob, now)
			if age != tc.expected {
				t.Errorf("Expected age %d, got %d", tc.expected, age)
			}
		})
	}
}

func mustParse(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return t
}

func TestIsToday(t *testing.T) {
	now := time.Now()

	if !IsToday(now) {
		t.Error("Expected IsToday to return true for current time")
	}

	// Tomorrow
	tomorrow := now.AddDate(0, 0, 1)
	if IsToday(tomorrow) {
		t.Error("Expected IsToday to return false for tomorrow")
	}

	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if IsToday(yesterday) {
		t.Error("Expected IsToday to return false for yesterday")
	}
}

func TestIsYesterday(t *testing.T) {
	now := time.Now()

	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if !IsYesterday(yesterday) {
		t.Error("Expected IsYesterday to return true for yesterday")
	}

	// Today
	if IsYesterday(now) {
		t.Error("Expected IsYesterday to return false for today")
	}

	// Two days ago
	twoDaysAgo := now.AddDate(0, 0, -2)
	if IsYesterday(twoDaysAgo) {
		t.Error("Expected IsYesterday to return false for two days ago")
	}
}

func TestStartOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Wednesday",
			input:    time.Date(2025, 7, 9, 15, 30, 0, 0, time.UTC), // Wednesday
			expected: time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC),   // Monday
		},
		{
			name:     "Saturday",
			input:    time.Date(2025, 7, 12, 12, 0, 0, 0, time.UTC), // Sunday
			expected: time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC),   // Monday
		},
		{
			name:     "Sunday",
			input:    time.Date(2025, 7, 13, 12, 0, 0, 0, time.UTC), // Sunday
			expected: time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC),   // Monday
		},
		{
			name:     "Monday",
			input:    time.Date(2025, 7, 7, 10, 0, 0, 0, time.UTC), // Monday
			expected: time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StartOfWeek(tt.input)
			if !got.Equal(tt.expected) {
				t.Errorf("StartOfWeek() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEndOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Wednesday",
			input:    time.Date(2025, 7, 9, 15, 30, 0, 0, time.UTC),           // Wednesday
			expected: time.Date(2025, 7, 13, 23, 59, 59, 999999000, time.UTC), // Sunday
		},
		{
			name:     "Saturday",
			input:    time.Date(2025, 7, 12, 12, 0, 0, 0, time.UTC), // Sunday
			expected: time.Date(2025, 7, 13, 23, 59, 59, 999999000, time.UTC),
		},
		{
			name:     "Sunday",
			input:    time.Date(2025, 7, 13, 12, 0, 0, 0, time.UTC), // Sunday
			expected: time.Date(2025, 7, 13, 23, 59, 59, 999999000, time.UTC),
		},
		{
			name:     "Monday",
			input:    time.Date(2025, 7, 7, 10, 0, 0, 0, time.UTC), // Monday
			expected: time.Date(2025, 7, 13, 23, 59, 59, 999999000, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EndOfWeek(tt.input)
			if !got.Equal(tt.expected) {
				t.Errorf("EndOfWeek() = %v, want %v", got, tt.expected)
			}
		})
	}
}
