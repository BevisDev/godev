package console

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

type Logger struct {
	pkg string
	mu  sync.Mutex
	l   *log.Logger
}

// New creates a console logger with package name
func New(pkg string) *Logger {
	return &Logger{
		pkg: pkg,
		l:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (lg *Logger) log(level Level, format string, args ...interface{}) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	msg := fmt.Sprintf(format, args...)

	lg.l.Printf(
		"[%s] [%s] %s",
		level,
		lg.pkg,
		msg,
	)
}

func (lg *Logger) Debug(format string, args ...interface{}) {
	lg.log(LevelDebug, format, args...)
}

func (lg *Logger) Info(format string, args ...interface{}) {
	lg.log(LevelInfo, format, args...)
}

func (lg *Logger) Warn(format string, args ...interface{}) {
	lg.log(LevelWarn, format, args...)
}

func (lg *Logger) Error(format string, args ...interface{}) {
	lg.log(LevelError, format, args...)
}
