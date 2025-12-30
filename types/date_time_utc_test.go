package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/BevisDev/godev/utils/datetime"
)

func TestDateTimeUTC_UnmarshalJSON_Valid(t *testing.T) {
	input := `"2024-04-21T15:30:00Z"`
	var d DateTimeUTC
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	expected, _ := time.Parse(datetime.DatetimeUTC, "2024-04-21T15:30:00Z")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateTimeUTC_UnmarshalJSON_Null(t *testing.T) {
	input := `null`
	var d DateTimeUTC
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed on null: %v", err)
	}

	if !d.IsZero() {
		t.Errorf("Expected zero value, got %v", d.Time)
	}
}

func TestDateTimeUTC_MarshalJSON(t *testing.T) {
	dt, _ := time.Parse(datetime.DatetimeUTC, "2025-01-01T08:00:00Z")
	d := DateTimeUTC{Time: dt}

	data, err := json.Marshal(&d)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"2025-01-01T08:00:00Z"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, data)
	}
}

func TestDateTimeUTC_Scan_String(t *testing.T) {
	var d DateTimeUTC
	err := d.Scan("2023-12-01T10:00:00Z")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected, _ := time.Parse(datetime.DatetimeUTC, "2023-12-01T10:00:00Z")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateTimeUTC_Scan_Bytes(t *testing.T) {
	var d DateTimeUTC
	err := d.Scan([]byte("2023-12-01T10:00:00Z"))
	if err != nil {
		t.Fatalf("Scan []byte failed: %v", err)
	}

	expected, _ := time.Parse(datetime.DatetimeUTC, "2023-12-01T10:00:00Z")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateTimeUTC_Scan_Time(t *testing.T) {
	tm := time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC)
	var d DateTimeUTC
	err := d.Scan(tm)
	if err != nil {
		t.Fatalf("Scan time.Time failed: %v", err)
	}

	if !d.Equal(tm) {
		t.Errorf("Expected %v, got %v", tm, d.Time)
	}
}

func TestDateTimeUTC_Scan_InvalidType(t *testing.T) {
	var d DateTimeUTC
	err := d.Scan(12345)
	if err == nil {
		t.Errorf("Expected error for invalid Scan type, got nil")
	}
}

func TestDateTimeUTC_ToString(t *testing.T) {
	d := DateTimeUTC{}
	_ = d.Scan("2024-04-21T00:00:00Z")

	str := d.ToString()
	if str != "2024-04-21T00:00:00Z" {
		t.Errorf("Expected 2024-04-21T00:00:00Z, got %s", str)
	}
}
