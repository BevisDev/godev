package console

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Info(t *testing.T) {
	buf := new(bytes.Buffer)

	lg := New("rabbitmq")
	lg.l = log.New(buf, "", 0)

	lg.Info("consumer started")

	output := buf.String()

	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "[rabbitmq]")
	assert.Contains(t, output, "consumer started")
}

func TestLogger_WithFormat(t *testing.T) {
	buf := new(bytes.Buffer)

	lg := New("scheduler")
	lg.l = log.New(buf, "", 0)

	lg.Warn("retry in %d seconds", 5)

	output := buf.String()

	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "[scheduler]")
	assert.Contains(t, output, "retry in 5 seconds")
}

func TestLogger_Error(t *testing.T) {
	buf := new(bytes.Buffer)

	lg := New("db")
	lg.l = log.New(buf, "", 0)

	lg.Error("connection failed")

	output := buf.String()

	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "[db]")
	assert.Contains(t, output, "connection failed")
}

func TestLogger_Debug(t *testing.T) {
	buf := new(bytes.Buffer)

	lg := New("cache")
	lg.l = log.New(buf, "", 0)

	lg.Debug("set key success")

	output := buf.String()

	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "[cache]")
	assert.Contains(t, output, "set key success")
}
