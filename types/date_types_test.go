package types

import (
	"encoding/json"
	"testing"
	"time"
)

type Person struct {
	Name      string `json:"name"`
	BirthDate Date   `json:"birth_date"`
}

func TestDate_UnmarshalJSON(t *testing.T) {
	var d Date
	input := `"2023-12-25"`

	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	if !d.Time.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	d := Date{Time: time.Date(2024, 4, 21, 15, 30, 0, 0, time.UTC)}

	data, err := json.Marshal(&d)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"2024-04-21"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestDate_UnmarshalInvalidFormat(t *testing.T) {
	var d Date
	input := `"21-04-2024"`

	err := json.Unmarshal([]byte(input), &d)
	if err == nil {
		t.Errorf("Expected error for invalid date format, got nil")
	}
}

func TestBindJSONWithCustomDate(t *testing.T) {
	jsonInput := `{
		"name": "Alice",
		"birth_date": "1995-07-20"
	}`

	var p Person
	err := json.Unmarshal([]byte(jsonInput), &p)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedDate := time.Date(1995, 7, 20, 0, 0, 0, 0, time.UTC)
	if p.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", p.Name)
	}
	if !p.BirthDate.Equal(expectedDate) {
		t.Errorf("Expected birth_date %v, got %v", expectedDate, p.BirthDate)
	}
}

func TestDecodeJSONWithInvalidDate(t *testing.T) {
	jsonInput := `{
		"name": "Bob",
		"birth_date": "20-07-1995"
	}`

	var p Person
	err := json.Unmarshal([]byte(jsonInput), &p)
	if err == nil {
		t.Error("Expected error for invalid date format, got nil")
	}
}

func TestDate_Scan_String(t *testing.T) {
	var d DateSQL
	err := d.Scan("2023-12-01 10:00:00.000")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected, _ := time.Parse("2006-01-02 15:04:05.000", "2023-12-01 10:00:00.000")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDate_Scan_Bytes(t *testing.T) {
	var d Date
	err := d.Scan([]byte("2023-12-01"))
	if err != nil {
		t.Fatalf("Scan []byte failed: %v", err)
	}

	expected, _ := time.Parse("2006-01-02", "2023-12-01")
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d.Time)
	}
}

func TestDate_Scan_Time(t *testing.T) {
	tm := time.Date(2022, 10, 10, 10, 10, 10, 0, time.UTC)
	var d Date
	err := d.Scan(tm)
	if err != nil {
		t.Fatalf("Scan time.Time failed: %v", err)
	}

	if !d.Equal(tm) {
		t.Errorf("Expected %v, got %v", tm, d.Time)
	}
}

func TestDate_Scan_InvalidType(t *testing.T) {
	var d Date
	err := d.Scan(12345)
	if err == nil {
		t.Errorf("Expected error for invalid Scan type, got nil")
	}
}

func TestDate_ToString(t *testing.T) {
	d := Date{}
	_ = d.Scan("2024-04-21")

	str := d.ToString()
	if str != "2024-04-21" {
		t.Errorf("Expected 2024-04-21, got %s", str)
	}
}
