package logger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils/datetime"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/shopspring/decimal"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type RequestLogger struct {
	RID    string
	URL    string
	Time   time.Time
	Query  string
	Method string
	Header any
	Body   string
}

type ResponseLogger struct {
	RID      string
	Duration time.Duration
	Status   int
	Header   any
	Body     string
}

type Logger struct {
	cf   *Config
	zap  *zap.Logger
	cron *cron.Cron
}

// New creates and returns a new logger instance using Zap.
// Configures encoder format (JSON/console), output destination (file/stdout),
// and log rotation via Lumberjack. Includes caller information and uses InfoLevel by default.
func New(cfg *Config) (*Logger, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	cf := cfg.clone()
	l := &Logger{
		cf: cf,
	}

	// job runner to rotate log every day
	if cf.IsRotate {
		l.cron = cron.New()
	}

	encoder := l.getEncoderLog()
	writer := l.writeSync()

	l.zap = zap.New(
		zapcore.NewCore(
			encoder,
			writer,
			zapcore.InfoLevel,
		),
		zap.AddCaller(),
	)

	l.zap.Info("[logger] started successfully")
	return l, nil
}

// GetZap returns instance *zap.logger
func (l *Logger) GetZap() *zap.Logger {
	return l.zap
}

func (l *Logger) getEncoderLog() zapcore.Encoder {
	var encodeConfig zapcore.EncoderConfig

	if l.cf.IsProduction {
		encodeConfig = zap.NewProductionEncoderConfig()
		// 1716714967.877995 -> 2024-12-19T20:04:31.255+0700
		encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// ts -> time
		encodeConfig.TimeKey = "time"
		// msg -> message
		encodeConfig.MessageKey = "message"
		// info -> INFO
		encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		//"caller": logger/logger.go:91
		encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
		return zapcore.NewJSONEncoder(encodeConfig)
	}

	// for development
	encodeConfig = zap.NewDevelopmentEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.LevelKey = "level"
	encodeConfig.CallerKey = "caller"
	encodeConfig.MessageKey = "message"

	if l.cf.IsLocal {
		return zapcore.NewConsoleEncoder(encodeConfig)
	}
	return zapcore.NewJSONEncoder(encodeConfig)
}

func (l *Logger) writeSync() zapcore.WriteSyncer {
	if l.cf.IsLocal {
		return zapcore.AddSync(os.Stdout)
	}

	fileName := getFilename(l.cf.DirName, l.cf.Filename, l.cf.IsRotate)
	lumber := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    l.cf.MaxSize,
		MaxBackups: l.cf.MaxBackups,
		MaxAge:     l.cf.MaxAge,
		Compress:   l.cf.Compress,
	}

	// job runner to rotate log every day
	if l.cf.IsRotate {
		l.cron.AddFunc(l.cf.Cron, func() {
			lumber.Filename = getFilename(l.cf.DirName, l.cf.Filename, l.cf.IsRotate)
			if err := lumber.Rotate(); err != nil {
				log.Printf("[logger] failed to rotate log file: %v", err)
			}
		})
		l.cron.Start()
	}

	return zapcore.AddSync(lumber)
}

func getFilename(dir, fileName string, isRotate bool) string {
	if isRotate {
		now := datetime.ToString(time.Now(), datetime.DateLayoutISO)
		return filepath.Join(dir, now, fileName)
	}
	return filepath.Join(dir, fileName)
}

func (l *Logger) log(
	level zapcore.Level,
	callerSkips int,
	rid, msg string,
	fields []zap.Field,
	args ...interface{},
) {
	// format message
	message, errs := l.formatMessage(msg, args...)

	// skip caller before
	logging := l.zap.WithOptions(
		zap.AddCallerSkip(callerSkips),
	)

	// declare field
	fs := []zap.Field{
		zap.String(consts.RID, rid),
	}

	if len(errs) > 0 {
		if err := l.formatErrors(errs); err != nil {
			fs = append(fs,
				zap.Error(err),
			)
		}
	}

	fs = append(fs, fields...)

	switch level {
	case zapcore.InfoLevel:
		logging.Info(message, fs...)
	case zapcore.WarnLevel:
		logging.Warn(message, fs...)
	case zapcore.ErrorLevel:
		logging.Error(message, fs...)
	default:
		logging.Info(message, fs...)
	}
}

func (l *Logger) formatMessage(msg string, args ...interface{}) (string, []error) {
	if len(args) == 0 {
		return msg, nil
	}

	formatted, errors := l.formatArgs(args...)
	numArgs := len(formatted)

	// Handle {} placeholder pattern
	if strings.Contains(msg, "{}") {
		count := strings.Count(msg, "{}")
		if count < numArgs {
			msg += strings.Repeat(" :{}", numArgs-count)
		}

		// replace {} with actual values
		result := msg
		for _, arg := range formatted {
			result = strings.Replace(result, "{}", arg, 1)
		}
		return result, errors
	}

	// Default: append all args
	for _, arg := range formatted {
		msg += " :" + arg
	}

	return msg, errors
}

func (l *Logger) formatArgs(args ...interface{}) ([]string, []error) {
	out := make([]string, len(args))
	var errors []error

	for i, arg := range args {
		v, err := l.formatAny(arg)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		out[i] = v
	}
	return out, errors
}

func (l *Logger) formatAny(v interface{}) (string, error) {
	if v == nil {
		return "<nil>", nil
	}

	// Handle error type
	if err, ok := v.(error); ok {
		return "", err
	}

	rv := reflect.ValueOf(v)

	// Dereference pointers
	rv, v = l.dereferencePointer(rv)
	if !rv.IsValid() {
		return "<nil>", nil
	}

	// Handle special types
	if formatted := l.formatSpecialType(v); formatted != "" {
		return formatted, nil
	}

	// Handle complex types (struct, map, slice, array) via JSON
	if rv.Kind() == reflect.Struct || rv.Kind() == reflect.Map ||
		rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return jsonx.ToJSON(v), nil
	}

	// Default formatting
	if rv.CanInterface() {
		return fmt.Sprintf("%+v", rv.Interface()), nil
	}
	return fmt.Sprintf("<unreadable: %T>", v), nil
}

func (l *Logger) dereferencePointer(rv reflect.Value) (reflect.Value, interface{}) {
	v := rv.Interface()
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return reflect.Value{}, nil
		}
		rv = rv.Elem()
		if !rv.IsValid() {
			return reflect.Value{}, nil
		}
		v = rv.Interface()
	}
	return rv, v
}

func (l *Logger) formatErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	msgs := make([]string, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			msgs = append(msgs, err.Error())
		}
	}

	if len(msgs) == 0 {
		return nil
	}

	return errors.New(strings.Join(msgs, " | "))
}

func (l *Logger) formatSpecialType(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return fmt.Sprintf("%v", val)
	case decimal.Decimal:
		return val.String()
	case time.Time:
		return val.Format(time.RFC3339)
	case json.RawMessage:
		return string(val)
	case []byte:
		if utf8.Valid(val) {
			return fmt.Sprintf("%q", val)
		}
		return fmt.Sprintf("[]byte(len=%d)", len(val))
	case context.Context:
		return "<context>"
	case sql.NullString:
		if val.Valid {
			return val.String
		}
		return "<null>"
	case sql.NullInt64:
		if val.Valid {
			return fmt.Sprintf("%d", val.Int64)
		}
		return "<null>"
	case sql.NullFloat64:
		if val.Valid {
			return fmt.Sprintf("%f", val.Float64)
		}
		return "<null>"
	case sql.NullBool:
		if val.Valid {
			return fmt.Sprintf("%t", val.Bool)
		}
		return "<null>"
	case sql.NullTime:
		if val.Valid {
			return val.Time.Format(time.RFC3339)
		}
		return "<null>"
	default:
		return ""
	}
}

// Sync Forces any buffered log entries to be written out to the destination.
// Crucial for ensuring all logs are saved before application exit.
func (l *Logger) Sync() {
	if l.zap != nil {
		_ = l.zap.Sync()
	}
	// Stop cron scheduler if it exists
	if l.cron != nil {
		ctx := l.cron.Stop()
		<-ctx.Done()
	}
}

// Info Logs an informational message
func (l *Logger) Info(rid, msg string, args ...interface{}) {
	l.log(zapcore.InfoLevel,
		2,
		rid, msg,
		nil,
		args...,
	)
}

// Error Logs a recoverable error that occurred during execution.
func (l *Logger) Error(rid, msg string, args ...interface{}) {
	l.log(zapcore.ErrorLevel,
		2,
		rid, msg,
		nil,
		args...,
	)
}

// StackTrace logs a recoverable error with stacktrace attached.
// This should be used for unexpected errors where stack information is needed
// but the application can continue running.
func (l *Logger) StackTrace(rid, msg string, stack []byte, args ...interface{}) {
	l.log(zapcore.ErrorLevel,
		2,
		rid, msg,
		[]zap.Field{
			zap.ByteString("stack", stack),
		},
		args...,
	)
}

// Warn Logs a potentially harmful situation or an unexpected event that isn't an error.
func (l *Logger) Warn(rid, msg string, args ...interface{}) {
	l.log(zapcore.WarnLevel,
		2,
		rid, msg,
		nil,
		args...,
	)
}

// LogRequest Logs an incoming request to the application (e.g., an HTTP server receiving a client request).
func (l *Logger) LogRequest(req *RequestLogger) {
	l.logRequest(req, "[===== REQUEST INFO =====]", l.cf.CallerConfig.Request.Internal)
}

// LogResponse Logs the outgoing response generated by the application for an incoming request.
func (l *Logger) LogResponse(resp *ResponseLogger) {
	l.logResponse(resp, "[===== RESPONSE INFO =====]", l.cf.CallerConfig.Response.Internal)
}

// LogExtRequest Logs an outgoing request made by the application to an external service (External Request).
func (l *Logger) LogExtRequest(req *RequestLogger) {
	l.logRequest(req, "[===== REQUEST EXTERNAL INFO =====]", l.cf.CallerConfig.Request.External)
}

// LogExtResponse Logs the response received from an external service after an outgoing request (External Response).
func (l *Logger) LogExtResponse(resp *ResponseLogger) {
	l.logResponse(resp, "[===== RESPONSE EXTERNAL INFO =====]", l.cf.CallerConfig.Response.External)
}

func (l *Logger) logRequest(req *RequestLogger, message string, callerSkip int) {
	fields := []zap.Field{
		zap.String(consts.RID, req.RID),
		zap.String(consts.Url, req.URL),
		zap.Time(consts.Time, req.Time),
		zap.String(consts.Method, req.Method),
	}
	if req.Header != nil {
		fields = append(fields, zap.Any(consts.Header, req.Header))
	}
	fields = append(fields, zap.String(consts.Query, req.Query))
	fields = append(fields, zap.String(consts.Body, req.Body))

	l.zap.WithOptions(zap.AddCallerSkip(callerSkip)).
		Info(message, fields...)
}

func (l *Logger) logResponse(resp *ResponseLogger, message string, callerSkip int) {
	fields := []zap.Field{
		zap.String(consts.RID, resp.RID),
		zap.Int(consts.Status, resp.Status),
		zap.String(consts.Duration, resp.Duration.String()),
	}
	if resp.Header != nil {
		fields = append(fields, zap.Any(consts.Header, resp.Header))
	}
	fields = append(fields, zap.String(consts.Body, resp.Body))

	l.zap.WithOptions(zap.AddCallerSkip(callerSkip)).
		Info(message, fields...)
}
