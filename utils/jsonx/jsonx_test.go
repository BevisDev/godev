package jsonx

import (
	"testing"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestToJSONBytes(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}
	jsonBytes := ToJSONBytes(p)

	if jsonBytes == nil {
		t.Fatal("ToJSONBytes returned nil")
	}
}

func TestJSONBytesToStruct(t *testing.T) {
	jsonData := []byte(`{"name":"Bob","age":25}`)

	var p Person
	err := JSONBytesToStruct(jsonData, &p)

	if err != nil {
		t.Fatalf("JSONBytesToStruct failed: %v", err)
	}
	if p.Name != "Bob" || p.Age != 25 {
		t.Errorf("Got %+v; want {Name:Bob Age:25}", p)
	}
}

func TestJSONBytesToStruct_NotPointer(t *testing.T) {
	jsonData := []byte(`{"name":"Bob","age":25}`)

	var p Person
	err := JSONBytesToStruct(jsonData, p) // not pointer

	if err == nil || err.Error() != "must be a pointer" {
		t.Errorf("Expected 'must be a pointer' error, got %v", err)
	}
}

func TestToJSON(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}
	jsonStr := ToJSON(p)

	if jsonStr == "" {
		t.Fatal("ToJSON returned empty string")
	}
	if jsonStr != `{"name":"Alice","age":30}` && jsonStr != `{"age":30,"name":"Alice"}` {
		t.Errorf("ToJSON returned unexpected string: %s", jsonStr)
	}
}

func TestJSONToStruct(t *testing.T) {
	jsonStr := `{"name":"Carol","age":40}`

	var p Person
	err := ToStruct(jsonStr, &p)

	if err != nil {
		t.Fatalf("ToStruct failed: %v", err)
	}
	if p.Name != "Carol" || p.Age != 40 {
		t.Errorf("Got %+v; want {Name:Carol Age:40}", p)
	}
}

func TestJSONToStruct_NotPointer(t *testing.T) {
	jsonStr := `{"name":"Carol","age":40}`

	var p Person
	err := ToStruct(jsonStr, p)

	if err == nil || err.Error() != "must be a pointer" {
		t.Errorf("Expected 'must be a pointer' error, got %v", err)
	}
}

func TestStructToMap(t *testing.T) {
	p := Person{Name: "Dan", Age: 22}
	result := StructToMap(p)

	if result == nil {
		t.Fatal("StructToMap returned nil")
	}

	if result["name"] != "Dan" || int(result["age"].(float64)) != 22 {
		t.Errorf("Got %+v; want {name: Dan, age: 22}", result)
	}
}
