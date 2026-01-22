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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

var client = New()

func TestRestClient_Get(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := MockResponse{Message: "hello"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	type resultStruct struct {
		Message string `json:"message"`
	}

	result, err := NewRequest[*resultStruct](client).
		URL(server.URL).
		GET(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Data)
	assert.Equal(t, "hello", result.Data.Message)
}

func TestRestClient_Get_WithQueryParam(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/hello/GoLang", r.URL.Path)
		assert.Equal(t, "lang=en", strings.TrimPrefix(r.URL.RawQuery, "?"))
		resp := MockResponse{Message: "Hello GoLang"}
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	result, err := NewRequest[MockResponse](client).
		URL(server.URL + "/hello/:name").
		PathParams(map[string]string{"name": "GoLang"}).
		QueryParams(map[string]string{"lang": "en"}).
		GET(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Data)
	assert.Equal(t, "Hello GoLang", result.Data.Message)
}

func TestRestClient_Timeout(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(slowHandler)
	defer server.Close()

	clientTimeout := New(WithTimeout(1 * time.Second))

	start := time.Now()
	_, err := NewRequest[any](clientTimeout).
		URL(server.URL).
		GET(context.Background())
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.True(t, elapsed < 2*time.Second, "expected timeout before 2s, took %s", elapsed)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expected DeadlineExceeded, got %v", err)
}

func TestRestClient_PostForm_WithBodyFormAndHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "Test123", r.Header.Get("X-Custom-Header"))
		bodyBytes, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(bodyBytes))
		assert.Equal(t, "testuser", values.Get("username"))
		assert.Equal(t, "vi", values.Get("lang"))
		resp := MockResponse{Status: "ok"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	result, err := NewRequest[MockResponse](client).
		URL(server.URL).
		BodyForm(map[string]string{"username": "testuser", "lang": "vi"}).
		Headers(map[string]string{"X-Custom-Header": "Test123"}).
		PostForm(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Data)
	assert.Equal(t, "ok", result.Data.Status)
}

func TestRestClient_Server500Error(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "internal server error"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	_, err := NewRequest[any](client).
		URL(server.URL).
		GET(context.Background())
	require.Error(t, err)
	httpErr, ok := AsHttpError(err)
	require.True(t, ok)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func TestRestClient_Post(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
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
	require.NoError(t, err)
	require.NotNil(t, res.Data)
	assert.Equal(t, "Post called", res.Data.Message)
}

func TestRestClient_Put(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
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
	require.NoError(t, err)
	require.NotNil(t, res.Data)
	assert.Equal(t, "Put called", res.Data.Message)
}

func TestRestClient_Patch(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
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
	require.NoError(t, err)
	require.NotNil(t, res.Data)
	assert.Equal(t, "Patch called", res.Data.Message)
}

func TestRestClient_Delete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
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
	require.NoError(t, err)
	require.NotNil(t, res.Data)
	assert.Equal(t, "Delete called", res.Data.Message)
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
	require.NoError(t, err)
}

func TestExecute_GatewayTimeout504(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(504)
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, `<html><head><title>504 Gateway Time-out</title></head><body><h1>504 Gateway Time-out</h1></body></html>`)
	}))
	defer srv.Close()

	_, err := NewRequest[string](nil).
		URL(srv.URL).
		GET(context.Background())
	require.Error(t, err)

	httpErr, ok := AsHttpError(err)
	require.True(t, ok, "expected HttpError, got %T", err)
	assert.Equal(t, 504, httpErr.StatusCode)
	assert.Contains(t, httpErr.Body, "504 Gateway Time-out")
}

func TestExecute_ConnectionResetByPeer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	_, err = NewRequest[string](nil).
		URL("http://" + ln.Addr().String()).
		GET(context.Background())
	require.Error(t, err)
	errStr := err.Error()
	assert.True(t,
		strings.Contains(errStr, "EOF") ||
			strings.Contains(errStr, "connection reset") ||
			strings.Contains(errStr, "reset by peer"),
		"expected connection reset or EOF, got: %v", err)
}

func TestAsHttpError_NonHttpError(t *testing.T) {
	plainErr := errors.New("plain error")
	httpErr, ok := AsHttpError(plainErr)
	assert.False(t, ok)
	assert.Nil(t, httpErr)
}
