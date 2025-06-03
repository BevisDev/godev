package logger

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
	"time"
)

type User struct {
	ID   int
	Name string
}

func ptrTo[T any](v T) *T {
	return &v
}

func TestFormatMessage(t *testing.T) {
	logger := &AppLogger{}
	tests := []struct {
		msg      string
		args     []interface{}
		expected string
	}{
		{"hello", nil, "hello"},
		{"value is {}", []interface{}{123}, "value is 123"},
		{"value is {}", []interface{}{ptrTo("abc")}, "value is abc"},
		{"value is {}", []interface{}{(*string)(nil)}, "value is <nil>"},
		{"value is {}", []interface{}{User{ID: 1, Name: "Alice"}}, "value is {ID:1 Name:Alice}"},
		{"value is {}", []interface{}{&User{ID: 2, Name: "Bob"}}, "value is {ID:2 Name:Bob}"},
		{"multiple placeholders: {}, {}", []interface{}{123, "abc"}, "multiple placeholders: 123, abc"},
	}

	for _, tt := range tests {
		result := logger.formatMessage(tt.msg, tt.args...)
		assert.Equal(t, tt.expected, result, "msg: %q", tt.msg)
	}
}

func derefAll(v reflect.Value) (reflect.Value, bool) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
	return v, true
}

func TestDerefAll(t *testing.T) {
	t.Run("non-pointer value", func(t *testing.T) {
		v := reflect.ValueOf(42)
		res, ok := derefAll(v)
		assert.True(t, ok)
		assert.Equal(t, 42, res.Interface())
	})

	t.Run("single pointer", func(t *testing.T) {
		n := 99
		v := reflect.ValueOf(&n)
		res, ok := derefAll(v)
		assert.True(t, ok)
		assert.Equal(t, 99, res.Interface())
	})

	t.Run("double pointer", func(t *testing.T) {
		s := "hello"
		p1 := &s
		p2 := &p1
		v := reflect.ValueOf(p2)
		res, ok := derefAll(v)
		assert.True(t, ok)
		assert.Equal(t, "hello", res.Interface())
	})

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *string = nil
		v := reflect.ValueOf(ptr)
		res, ok := derefAll(v)
		assert.False(t, ok)
		assert.False(t, res.IsValid())
	})

	t.Run("nil in deep pointer", func(t *testing.T) {
		var s *string = nil
		var p2 **string = &s
		v := reflect.ValueOf(p2)
		res, ok := derefAll(v)
		assert.False(t, ok)
		assert.False(t, res.IsValid())
	})
}

func TestInfoLog(t *testing.T) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.InfoLevel,
	)
	zapLogger := zap.New(core)

	logger := &AppLogger{Logger: zapLogger}
	logger.Info("TEST_STATE", "Hello {}", "World")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Hello World")
	assert.Contains(t, logOutput, "TEST_STATE")
}

func TestErrorLog(t *testing.T) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.ErrorLevel,
	)
	zapLogger := zap.New(core)

	logger := &AppLogger{Logger: zapLogger}
	logger.Error("ERR_STATE", "Something went wrong: {}", "disk full")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Something went wrong: disk full")
	assert.Contains(t, logOutput, "ERR_STATE")
}

func TestLogRequest(t *testing.T) {
	tLogger := zaptest.NewLogger(t)
	appLogger := &AppLogger{Logger: tLogger}

	req := &RequestLogger{
		State:  "REQ_TEST",
		URL:    "/api/test",
		Time:   time.Now(),
		Method: "GET",
		Query:  "id=1",
		Header: map[string]string{"Authorization": "Bearer token"},
		Body:   map[string]string{"data": "value"},
	}

	appLogger.LogRequest(req)
}

func TestLogResponse(t *testing.T) {
	tLogger := zaptest.NewLogger(t)
	appLogger := &AppLogger{Logger: tLogger}

	resp := &ResponseLogger{
		State:       "RESP_TEST",
		DurationSec: 2 * time.Second,
		Status:      200,
		Header:      map[string]string{"Content-Type": "application/json"},
		Body:        map[string]string{"result": "ok"},
	}

	appLogger.LogResponse(resp)
}
