package logger

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestFormatMessage(t *testing.T) {
	logger := &AppLogger{}

	tests := []struct {
		msg      string
		args     []interface{}
		expected string
	}{
		{"hello", nil, "hello"},
		{"value is {}", []interface{}{123}, "value is 123"},
		{"value is", []interface{}{123}, "value is:123"},
		{"value is %v", []interface{}{123}, "value is 123"},
	}

	for _, tt := range tests {
		result := logger.formatMessage(tt.msg, tt.args...)
		assert.Equal(t, tt.expected, result)
	}
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
