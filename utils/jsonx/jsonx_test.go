package jsonx

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
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

func TestPretty_Struct(t *testing.T) {
	type User struct {
		ID    int
		Name  string
		Email *string
	}

	email := "alice@example.com"
	u := User{ID: 1, Name: "Alice", Email: &email}

	output := Pretty(u)

	expected := `{
  "ID": 1,
  "Name": "Alice",
  "Email": "alice@example.com"
}`

	assert.JSONEq(t, expected, output)
}

func TestPretty_Map(t *testing.T) {
	m := map[string]int{
		"apple":  3,
		"banana": 5,
	}
	result := Pretty(m)

	// Validate it's valid JSON
	var check map[string]int
	err := json.Unmarshal([]byte(result), &check)
	assert.NoError(t, err)
	assert.Equal(t, 3, check["apple"])
	assert.Equal(t, 5, check["banana"])
}

func TestPretty_Slice(t *testing.T) {
	list := []string{"a", "b", "c"}
	result := Pretty(list)

	expected := `[
  "a",
  "b",
  "c"
]`

	assert.JSONEq(t, expected, result)
}

func TestPretty_Nil(t *testing.T) {
	var m map[string]int = nil
	result := Pretty(m)

	assert.Equal(t, "null", result)
}

func TestClonePerson(t *testing.T) {
	original := Person{
		Name: "Bob",
		Age:  40,
	}

	cloned, err := Clone(original)
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	if cloned != original {
		t.Errorf("Cloned object does not match original. Got %+v", cloned)
	}

	// Modify cloned to ensure it's not the same reference (deep copy)
	cloned.Name = "Alice"
	if original.Name == cloned.Name {
		t.Errorf("Original changed when cloned modified!")
	}
}
