package datetime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateSQL_UnmarshalJSON_Valid(t *testing.T) {
	input := `"2024-04-21 15:30:00.000"`
	var d DBTime
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	expected, _ := time.Parse(DateTimeMillis, "2024-04-21 15:30:00.000")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateSQL_UnmarshalJSON_Null(t *testing.T) {
	input := `null`
	var d DBTime
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed on null: %v", err)
	}

	if !d.IsZero() {
		t.Errorf("Expected zero value, got %v", d.Time)
	}
}

func TestDateSQL_MarshalJSON(t *testing.T) {
	dt, _ := time.Parse(DateTimeMillis, "2025-01-01 08:00:00.000")
	d := DBTime{Time: dt}

	data, err := json.Marshal(&d)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"2025-01-01 08:00:00.000"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, data)
	}
}

func TestDateSQL_Scan_String(t *testing.T) {
	var d DBTime
	err := d.Scan("2023-12-01 10:00:00.000")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected, _ := time.Parse(DateTimeMillis, "2023-12-01 10:00:00.000")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateSQL_Scan_Bytes(t *testing.T) {
	var d DBTime
	err := d.Scan([]byte("2023-12-01 10:00:00.000"))
	if err != nil {
		t.Fatalf("Scan []byte failed: %v", err)
	}

	expected, _ := time.Parse(DateTimeMillis, "2023-12-01 10:00:00.000")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateSQL_Scan_Time(t *testing.T) {
	tm := time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC)
	var d DBTime
	err := d.Scan(tm)
	if err != nil {
		t.Fatalf("Scan time.Time failed: %v", err)
	}

	if !d.Equal(tm) {
		t.Errorf("Expected %v, got %v", tm, d.Time)
	}
}

func TestDateSQL_Scan_InvalidType(t *testing.T) {
	var d DBTime
	err := d.Scan(12345)
	if err == nil {
		t.Errorf("Expected error for invalid Scan type, got nil")
	}
}

func TestDateSQL_ToString(t *testing.T) {
	d := DBTime{}
	_ = d.Scan("2024-04-21 00:00:00.000")

	str := d.String()
	if str != "2024-04-21 00:00:00.000" {
		t.Errorf("Expected 2024-04-21 00:00:00.000, got %s", str)
	}
}
