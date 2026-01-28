package logger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

type User struct {
	ID   int
	Name string
}

func ptrTo[T any](v T) *T {
	return &v
}

func TestFormatMessage(t *testing.T) {
	logger := &Logger{}
	now := time.Date(2025, 6, 9, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		msg      string
		args     []interface{}
		expected string
		errCount int
	}{
		{"hello", nil, "hello", 0},
		{"value is {}", []interface{}{123}, "value is 123", 0},
		{"value is {}", []interface{}{ptrTo("abc")}, "value is abc", 0},
		{"value is {}", []interface{}{(*string)(nil)}, "value is <nil>", 0},
		{"value is {}", []interface{}{User{ID: 1, Name: "Alice"}}, `value is {"ID":1,"Name":"Alice"}`, 0},
		{"value is {}", []interface{}{&User{ID: 2, Name: "Bob"}}, `value is {"ID":2,"Name":"Bob"}`, 0},
		{"multiple placeholders: {}, {}", []interface{}{123, "abc"}, "multiple placeholders: 123, abc", 0},
		{"decimal: {}", []interface{}{decimal.NewFromFloat(12.34)}, "decimal: 12.34", 0},
		{"nullstring: {}", []interface{}{sql.NullString{String: "ok", Valid: true}}, "nullstring: ok", 0},
		{"nullstring: {}", []interface{}{sql.NullString{Valid: false}}, "nullstring: <null>", 0},
		{"time: {}", []interface{}{now}, "time: 2025-06-09T10:00:00Z", 0},
		{"bytes: {}", []interface{}{[]byte("hello")}, `bytes: "hello"`, 0},
		{"bytes: {}", []interface{}{[]byte{0xff, 0xfe}}, "bytes: []byte(len=2)", 0},
		{"raw json: {}", []interface{}{json.RawMessage(`{"foo":"bar"}`)}, `raw json: {"foo":"bar"}`, 0},
		{"err: {}", []interface{}{fmt.Errorf("something went wrong")}, "err: ", 1},
		{"ctx: {}", []interface{}{context.Background()}, "ctx: <context>", 0},
	}

	for _, tt := range tests {
		result, errs := logger.formatMessage(tt.msg, tt.args...)

		assert.Equal(t, tt.expected, result, "msg: %q", tt.msg)
		assert.Len(t, errs, tt.errCount, "msg: %q", tt.msg)
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

	logger := &Logger{zap: zapLogger}
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

	logger := &Logger{zap: zapLogger}
	logger.Error("ERR_STATE", "Something went wrong: {}", "disk full")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Something went wrong: disk full")
	assert.Contains(t, logOutput, "ERR_STATE")
}

func TestErrorLog_WithErrorAndValue(t *testing.T) {
	buf := &bytes.Buffer{}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.ErrorLevel,
	)

	zapLogger := zap.New(core)
	logger := &Logger{zap: zapLogger}

	err := errors.New("disk full")

	logger.Error(
		"ERR_STATE",
		"Something went wrong: {} (retry={})",
		err,
		3,
	)

	logOutput := buf.String()

	// message
	assert.Contains(t, logOutput, "Something went wrong:  (retry=3)")

	// error field (zap.Error)
	assert.Contains(t, logOutput, "disk full")

	// error code / tag
	assert.Contains(t, logOutput, "ERR_STATE")
}

func TestLogger_StackTrace(t *testing.T) {
	// Arrange
	core, recorded := observer.New(zapcore.ErrorLevel)
	zapLogger := zap.New(core, zap.AddCaller())

	logx := &Logger{
		zap: zapLogger,
	}

	rid := "rid-123"
	msg := "unexpected error occurred"
	stack := []byte("fake stacktrace")

	// Act
	logx.StackTrace(rid, msg, stack)

	// Assert
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}

	entry := logs[0]

	if entry.Level != zapcore.ErrorLevel {
		t.Errorf("expected level ERROR, got %v", entry.Level)
	}

	if entry.Message != msg {
		t.Errorf("expected message %q, got %q", msg, entry.Message)
	}

	fields := entry.ContextMap()

	if fields["rid"] != rid {
		t.Errorf("expected rid %q, got %v", rid, fields["rid"])
	}

	if fields["stack"] != string(stack) {
		t.Errorf("stack mismatch")
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot get runtime caller")
	}

	expectedFile := filepath.Base(file)
	actualFile := filepath.Base(entry.Caller.File)

	if expectedFile != actualFile {
		t.Errorf(
			"expected caller file %s, got %s",
			expectedFile,
			actualFile,
		)
	}
}

func TestWarnLog(t *testing.T) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.WarnLevel,
	)
	zapLogger := zap.New(core)

	logger := &Logger{zap: zapLogger}
	logger.Warn("WARN_STATE", "Careful: {}", "low disk")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Careful: low disk")
	assert.Contains(t, logOutput, "WARN_STATE")
}

func TestLogRequest(t *testing.T) {
	tLogger := zaptest.NewLogger(t)
	appLogger := &Logger{
		zap: tLogger,
		cf: &Config{
			CallerConfig: CallerConfig{},
		},
	}

	req := &RequestLogger{
		RID:    "REQ_TEST",
		URL:    "/api/test",
		Time:   time.Now(),
		Method: "GET",
		Query:  "id=1",
		Header: map[string]string{"Authorization": "Bearer token"},
		Body:   jsonx.ToJSON(map[string]string{"data": "value"}),
	}

	appLogger.LogRequest(req)
}

func TestLogResponse(t *testing.T) {
	tLogger := zaptest.NewLogger(t)
	appLogger := &Logger{
		zap: tLogger,
		cf: &Config{
			CallerConfig: CallerConfig{},
		},
	}

	now := time.Now()
	time.Sleep(2 * time.Second)
	resp := &ResponseLogger{
		RID:      "RESP_TEST",
		Duration: time.Since(now),
		Status:   200,
		Header:   map[string]string{"Content-Type": "application/json"},
		Body:     jsonx.ToJSON(map[string]string{"result": "ok"}),
	}

	appLogger.LogResponse(resp)
}
