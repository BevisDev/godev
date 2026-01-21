package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
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

// Setup client
var client = New()

func TestRestClient_Get(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := MockResponse{Message: "hello"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Setup request
	type resultStruct struct {
		Message string `json:"message"`
	}

	result, err := NewRequest[*resultStruct](client).
		URL(server.URL).
		GET(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// check response
	data := result.Data
	if data.Message != "hello" {
		t.Errorf("expected result to be 'hello', got: %s", data.Message)
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

	// Call GET
	result, err := NewRequest[MockResponse](client).
		URL(server.URL + "/hello/:name").
		PathParams(map[string]string{"name": "GoLang"}).
		QueryParams(map[string]string{"lang": "en"}).
		GET(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// check response
	if result.Data.Message != "Hello GoLang" {
		t.Errorf("expected message 'Hello GoLang', got %s", result.Data.Message)
	}
}

func TestRestClient_Timeout(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(slowHandler)
	defer server.Close()

	clientTimeout := New(
		WithTimeout(1 * time.Second),
	)

	// Do request
	start := time.Now()
	_, err := NewRequest[any](clientTimeout).
		URL(server.URL).
		GET(context.Background())

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

	// Call PostForm
	result, err := NewRequest[MockResponse](client).
		URL(server.URL).
		BodyForm(map[string]string{
			"username": "testuser",
			"lang":     "vi",
		}).
		Headers(map[string]string{
			"X-Custom-Header": "Test123",
		}).
		PostForm(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// check response
	if result.Data.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", result.Data.Status)
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

	// Call GET
	_, err := NewRequest[any](client).
		URL(server.URL).
		GET(context.Background())
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

	type Response struct {
		Message string `json:"message"`
	}

	res, err := NewRequest[Response](client).
		URL(server.URL).
		Body(map[string]string{"key": "value"}).
		POST(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Data.Message != "Post called" {
		t.Errorf("Expected message 'Post called', got %s", res.Data.Message)
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

	type Response struct {
		Message string `json:"message"`
	}

	res, err := NewRequest[Response](client).
		URL(server.URL).
		Body(map[string]string{"key": "value"}).
		PUT(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Data.Message != "Put called" {
		t.Errorf("Expected message 'Put called', got %s", res.Data.Message)
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

	type Response struct {
		Message string `json:"message"`
	}

	res, err := NewRequest[Response](client).
		URL(server.URL).
		Body(map[string]string{"key": "value"}).
		PATCH(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Data.Message != "Patch called" {
		t.Errorf("Expected message 'Patch called', got %s", res.Data.Message)
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

	type Response struct {
		Message string `json:"message"`
	}

	res, err := NewRequest[Response](client).
		URL(server.URL).
		Body(map[string]string{"key": "value"}).
		DELETE(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Data.Message != "Delete called" {
		t.Errorf("Expected message 'Delete called', got %s", res.Data.Message)
	}
}

func TestRestClient_Response_NoBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	_, err := NewRequest[any](client).
		URL(server.URL).
		Body(map[string]string{"key": "value"}).
		POST(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_GatewayTimeout504(t *testing.T) {
	// Mock server trả 504 Gateway Time-out HTML
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(504)
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, `
<html>
<head><title>504 Gateway Time-out</title></head>
<body>
<center><h1>504 Gateway Time-out</h1></center>
<hr><center>nginx</center>
</body>
</html>
`)
	}))
	defer srv.Close()

	_, err := NewRequest[string](nil).
		URL(srv.URL).
		GET(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	httpErr, ok := AsHttpError(err)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != 504 {
		t.Fatalf("expected status 504, got %d", httpErr.StatusCode)
	}

	if !strings.Contains(httpErr.Body, "504 Gateway Time-out") {
		t.Fatalf("expected HTML body, got %s", httpErr.Body)
	}

	t.Logf("Got expected error: %v", httpErr)
}

func TestExecute_ConnectionResetByPeer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close() // đóng ngay → simulate ECONNRESET
		}
	}()

	_, err = NewRequest[string](nil).
		URL("http://" + ln.Addr().String()).
		GET(context.Background())
	if err == nil {
		t.Fatal("expected connection reset error but got nil")
	}

	if !strings.Contains(err.Error(), "EOF") &&
		!strings.Contains(err.Error(), "connection reset") &&
		!strings.Contains(err.Error(), "reset by peer") {
		t.Fatalf("expected connection reset or EOF, got: %v", err)
	}

	t.Logf("Got expected network error: %v", err)
}
