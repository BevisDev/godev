package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type MockResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func TestRestClient_Get(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := MockResponse{Message: "hello"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Setup client
	client := NewRestClient(nil)

	// Setup request
	type resultStruct struct {
		Message string `json:"message"`
	}
	result := &resultStruct{}
	req := &Request{
		URL:    server.URL,
		Result: result,
	}

	err := client.Get(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// check response
	if result.Message != "hello" {
		t.Errorf("expected result to be 'hello', got: %s", result.Message)
	}
}

func TestRestClient_Get_WithQueryParam(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/hello/GoLang"
		expectedQuery := "lang=en"

		// check path
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// check query
		if strings.TrimPrefix(r.URL.RawQuery, "?") != expectedQuery {
			t.Errorf("expected query %s, got %s", expectedQuery, r.URL.RawQuery)
		}

		resp := MockResponse{Message: "Hello GoLang"}
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create client
	client := NewRestClient(nil)

	// Setup RestRequest
	result := &MockResponse{}
	req := &Request{
		URL:    server.URL + "/hello/:name",
		Query:  map[string]string{"name": "GoLang"},
		Params: map[string]string{"lang": "en"},
		Result: result,
	}

	// Call GET
	err := client.Get(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// check response
	if result.Message != "Hello GoLang" {
		t.Errorf("expected message 'Hello GoLang', got %s", result.Message)
	}
}

func TestRestClient_Timeout(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(slowHandler)
	defer server.Close()

	// Tạo RestClient với timeout thấp hơn thời gian phản hồi server
	client := NewRestClient(&RestConfig{
		TimeoutSec: 1,
	})

	// Request mock
	req := &Request{
		URL: server.URL,
	}

	// Do Request
	start := time.Now()
	err := client.Get(context.Background(), req)
	elapsed := time.Since(start)

	// Check err timeout
	if err == nil {
		t.Fatal("expected timeout error but got nil")
	}

	// make sure time smaller than time sleep in server
	if elapsed >= 2*time.Second {
		t.Errorf("expected timeout before 2s, but took %s", elapsed)
	}

	// check error time out
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context deadline exceeded, got: %v", err)
	}
}

func TestRestClient_PostForm_WithBodyFormAndHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check content-type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected content-type form, got %s", r.Header.Get("Content-Type"))
		}

		// check header
		if r.Header.Get("X-Custom-Header") != "Test123" {
			t.Errorf("missing or wrong X-Custom-Header: %s", r.Header.Get("X-Custom-Header"))
		}

		// check body form
		bodyBytes, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(bodyBytes))
		if values.Get("username") != "testuser" || values.Get("lang") != "vi" {
			t.Errorf("unexpected form body: %s", string(bodyBytes))
		}

		resp := MockResponse{Status: "ok"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Setup client
	client := NewRestClient(nil)

	// Setup request
	result := &MockResponse{}
	req := &Request{
		URL: server.URL,
		BodyForm: map[string]string{
			"username": "testuser",
			"lang":     "vi",
		},
		Header: map[string]string{
			"X-Custom-Header": "Test123",
		},
		Result: result,
	}

	// Call PostForm
	err := client.PostForm(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// check response
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", result.Status)
	}
}

func TestRestClient_Server500Error(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// response error 500
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "internal server error"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// new rest client
	client := NewRestClient(nil)

	// Setup RestRequest
	req := &Request{
		URL:    server.URL,
		Result: &struct{}{},
	}

	// Call GET
	err := client.Get(context.Background(), req)

	// Check error
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	t.Logf("Received error as expected: %v", err)
}

func TestRestClient_Post(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method Post, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Post called"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewRestClient(nil)

	type Response struct {
		Message string `json:"message"`
	}
	res := &Response{}

	err := client.Post(context.Background(), &Request{
		URL:    server.URL,
		Body:   map[string]string{"key": "value"},
		Result: res,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Message != "Post called" {
		t.Errorf("Expected message 'Post called', got %s", res.Message)
	}
}

func TestRestClient_Put(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected method Put, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Put called"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewRestClient(nil)

	type Response struct {
		Message string `json:"message"`
	}
	res := &Response{}

	err := client.Put(context.Background(), &Request{
		URL:    server.URL,
		Body:   map[string]string{"key": "value"},
		Result: res,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Message != "Put called" {
		t.Errorf("Expected message 'Put called', got %s", res.Message)
	}
}

func TestRestClient_Patch(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected method Patch, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Patch called"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewRestClient(nil)

	type Response struct {
		Message string `json:"message"`
	}
	res := &Response{}

	err := client.Patch(context.Background(), &Request{
		URL:    server.URL,
		Body:   map[string]string{"key": "value"},
		Result: res,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Message != "Patch called" {
		t.Errorf("Expected message 'Patch called', got %s", res.Message)
	}
}

func TestRestClient_Delete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected method Delete, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Delete called"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewRestClient(nil)

	type Response struct {
		Message string `json:"message"`
	}
	res := &Response{}

	err := client.Delete(context.Background(), &Request{
		URL:    server.URL,
		Body:   map[string]string{"key": "value"},
		Result: res,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Message != "Delete called" {
		t.Errorf("Expected message 'Delete called', got %s", res.Message)
	}
}

func TestRestClient_Response_NoBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewRestClient(nil)

	type Response struct {
		Message string `json:"message"`
	}
	res := &Response{}

	err := client.Post(context.Background(), &Request{
		URL:    server.URL,
		Body:   map[string]string{"key": "value"},
		Result: res,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Message != "" {
		t.Errorf("expected empty message, got %s", res.Message)
	}
}
