package jsonx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestToJSONBytes(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}

	jsonBytes, err := ToJSONBytes(p)
	assert.NoError(t, err)
	assert.NotNil(t, jsonBytes)

	var out Person
	err = json.Unmarshal(jsonBytes, &out)
	assert.NoError(t, err)
	assert.Equal(t, p, out)
}

func TestFromJSONBytes(t *testing.T) {
	jsonData := []byte(`{"name":"Bob","age":25}`)

	p, err := FromJSONBytes[Person](jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "Bob", p.Name)
	assert.Equal(t, 25, p.Age)
}

func TestFromJSON(t *testing.T) {
	jsonStr := `{"name":"Carol","age":40}`

	p, err := FromJSON[Person](jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, "Carol", p.Name)
	assert.Equal(t, 40, p.Age)
}

func TestToJSON(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}

	jsonStr := ToJSON(p)
	assert.NotEmpty(t, jsonStr)

	var out Person
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	assert.Equal(t, p, out)
}

func TestObjectToMap(t *testing.T) {
	p := Person{Name: "Dan", Age: 22}

	m := ObjectToMap(&p)
	assert.NotNil(t, m)

	assert.Equal(t, "Dan", m["name"])
	assert.Equal(t, float64(22), m["age"])
}

func TestJSONToMap(t *testing.T) {
	jsonStr := `{
		"age": 20,
		"name": "Alice",
		"email": "alice@example.com"
	}`

	m, err := FromJSON[map[string]interface{}](jsonStr)
	assert.NoError(t, err)

	assert.Equal(t, "Alice", m["name"])
	assert.Equal(t, float64(20), m["age"])
}

func TestFromJSON_ObjectToStruct(t *testing.T) {
	jsonStr := `{
		"age": 20,
		"name": "Alice",
		"email": "alice@example.com"
	}`

	user, err := FromJSON[Person](jsonStr)
	assert.NoError(t, err)

	assert.Equal(t, 20, user.Age)
	assert.Equal(t, "Alice", user.Name)
}

func TestFromJSON_ArrayToSlice(t *testing.T) {
	jsonStr := `[
		{ "age": 10, "name": "Alice" },
		{ "age": 22, "name": "Bob" }
	]`

	users, err := FromJSON[[]Person](jsonStr)
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	assert.Equal(t, 10, users[0].Age)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, 22, users[1].Age)
	assert.Equal(t, "Bob", users[1].Name)
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
	require.NoError(t, err)
	assert.Equal(t, original, cloned)

	cloned.Name = "Alice"
	assert.NotEqual(t, original.Name, cloned.Name, "clone should be independent")
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	_, err := FromJSON[Person](`{invalid json`)
	assert.Error(t, err)
}

func TestClone_UnmarshalError(t *testing.T) {
	ch := make(chan int)
	_, err := Clone(ch)
	assert.Error(t, err)
}
