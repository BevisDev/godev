package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/BevisDev/godev/utils/datetime"
)

func TestDateTime_UnmarshalJSON_Valid(t *testing.T) {
	input := `"2024-04-21T15:30:00"`
	var d DateTime
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	expected, _ := time.Parse(datetime.DateTimeNoTZ, "2024-04-21T15:30:00")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateTime_UnmarshalJSON_Null(t *testing.T) {
	input := `null`
	var d DateTime
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed on null: %v", err)
	}

	if !d.IsZero() {
		t.Errorf("Expected zero value, got %v", d.Time)
	}
}

func TestDateTime_MarshalJSON(t *testing.T) {
	dt, _ := time.Parse(datetime.DateTimeNoTZ, "2025-01-01T08:00:00")
	d := DateTime{Time: dt}

	data, err := json.Marshal(&d)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"2025-01-01T08:00:00"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, data)
	}
}

func TestDateTime_Scan_String(t *testing.T) {
	var d DateTime
	err := d.Scan("2023-12-01T10:00:00")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected, _ := time.Parse(datetime.DateTimeNoTZ, "2023-12-01T10:00:00")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateTime_Scan_Bytes(t *testing.T) {
	var d DateTime
	err := d.Scan([]byte("2023-12-01T10:00:00"))
	if err != nil {
		t.Fatalf("Scan []byte failed: %v", err)
	}

	expected, _ := time.Parse(datetime.DateTimeNoTZ, "2023-12-01T10:00:00")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDateTime_Scan_Time(t *testing.T) {
	tm := time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC)
	var d DateTime
	err := d.Scan(tm)
	if err != nil {
		t.Fatalf("Scan time.Time failed: %v", err)
	}

	if !d.Equal(tm) {
		t.Errorf("Expected %v, got %v", tm, d.Time)
	}
}

func TestDateTime_Scan_InvalidType(t *testing.T) {
	var d DateTime
	err := d.Scan(12345)
	if err == nil {
		t.Errorf("Expected error for invalid Scan type, got nil")
	}
}

func TestDateTime_ToString(t *testing.T) {
	d := DateTime{}
	_ = d.Scan("2024-04-21T00:00:00")

	str := d.ToString()
	if str != "2024-04-21T00:00:00" {
		t.Errorf("Expected 2024-04-21T00:00:00, got %s", str)
	}
}
