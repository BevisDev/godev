package datetime

import (
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
		kind     string
		value    int
		expected time.Time
	}{
		{Nanosecond, 10, base.Add(10 * time.Nanosecond)},
		{Millisecond, 10, base.Add(10 * time.Millisecond)},
		{Second, 10, base.Add(10 * time.Second)},
		{Minute, 5, base.Add(5 * time.Minute)},
		{Hour, 2, base.Add(2 * time.Hour)},
		{Day, 3, base.AddDate(0, 0, 3)},
		{Month, 2, base.AddDate(0, 2, 0)},
		{Year, 1, base.AddDate(1, 0, 0)},
		{"invalid", 999, base},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			result := AddTime(base, tt.value, tt.kind)
			if !result.Equal(tt.expected) {
				t.Errorf("AddTime kind=%q = %v; want %v", tt.kind, result, tt.expected)
			}
		})
	}
}
