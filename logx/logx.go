package logx

import (
	"context"
	"database/sql"
	"encoding/json"
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

type AppLogger struct {
	*Config
	zap  *zap.Logger
	cron *cron.Cron
}

// New creates and returns a new logger instance using Zap.
// Configures encoder format (JSON/console), output destination (file/stdout),
// and log rotation via Lumberjack. Includes caller information and uses InfoLevel by default.
func New(cf *Config) Logger {
	l := &AppLogger{
		Config: cf,
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
	return l
}

func (l *AppLogger) GetZap() *zap.Logger {
	return l.zap
}

func (l *AppLogger) getEncoderLog() zapcore.Encoder {
	var encodeConfig zapcore.EncoderConfig

	if l.IsProduction {
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

	if l.IsLocal {
		return zapcore.NewConsoleEncoder(encodeConfig)
	}
	return zapcore.NewJSONEncoder(encodeConfig)
}

func (l *AppLogger) writeSync() zapcore.WriteSyncer {
	if l.IsLocal {
		return zapcore.AddSync(os.Stdout)
	}

	fileName := getFilename(l.DirName, l.Filename, l.IsRotate)
	lumber := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    l.MaxSize,
		MaxBackups: l.MaxBackups,
		MaxAge:     l.MaxAge,
		Compress:   l.Compress,
	}

	// job runner to rotate log every day
	if l.IsRotate {
		l.cron.AddFunc("0 0 * * *", func() {
			lumber.Filename = getFilename(l.DirName, l.Filename, l.IsRotate)
			if err := lumber.Rotate(); err != nil {
				log.Printf("[logger] failed to rotate log file: %v", err)
			}
		})
		l.cron.Start()
	}

	return zapcore.AddSync(lumber)
}

func getFilename(dir, fileName string, isSplit bool) string {
	if isSplit {
		now := datetime.ToString(time.Now(), datetime.DateLayoutISO)
		return filepath.Join(dir, now, fileName)
	}
	return filepath.Join(dir, fileName)
}

func (l *AppLogger) log(level zapcore.Level,
	rid, msg string,
	args ...interface{},
) {
	// format message
	var message = l.formatMessage(msg, args...)

	// skip caller before
	logging := l.zap.WithOptions(
		zap.AddCallerSkip(2),
	)

	// declare field
	fields := []zap.Field{zap.String(consts.RID, rid)}

	switch level {
	case zapcore.InfoLevel:
		logging.Info(message, fields...)
	case zapcore.WarnLevel:
		logging.Warn(message, fields...)
	case zapcore.ErrorLevel:
		logging.Error(message, fields...)
	default:
		logging.Info(message, fields...)
	}
}

func (l *AppLogger) formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}

	numArgs := len(args)

	// Handle {} placeholder pattern
	if strings.Contains(msg, "{}") {
		count := strings.Count(msg, "{}")
		if count < numArgs {
			msg += strings.Repeat(" :{}", numArgs-count)
		}
		message := strings.ReplaceAll(msg, "{}", "%+v")
		return fmt.Sprintf(message, l.deferArgs(args...)...)
	}

	// Handle printf-style formatting (append formatted values)
	if strings.Contains(msg, "%") {
		for _, arg := range args {
			msg += " :" + l.formatAny(arg)
		}
		return msg
	}

	// Default: append all args
	msg += strings.Repeat(":%+v", numArgs)
	return fmt.Sprintf(msg, l.deferArgs(args...)...)
}

func (l *AppLogger) deferArgs(args ...interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, arg := range args {
		out[i] = l.formatAny(arg)
	}
	return out
}

func (l *AppLogger) formatAny(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	// Handle error type
	if err, ok := v.(error); ok {
		return err.Error()
	}

	rv := reflect.ValueOf(v)

	// Dereference pointers
	rv, v = l.dereferencePointer(rv)
	if !rv.IsValid() {
		return "<nil>"
	}

	// Handle special types
	if formatted := l.formatSpecialType(v); formatted != "" {
		return formatted
	}

	// Handle complex types (struct, map, slice, array) via JSON
	if rv.Kind() == reflect.Struct || rv.Kind() == reflect.Map ||
		rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return jsonx.ToJSON(v)
	}

	// Default formatting
	if rv.CanInterface() {
		return fmt.Sprintf("%+v", rv.Interface())
	}
	return fmt.Sprintf("<unreadable: %T>", v)
}

func (l *AppLogger) dereferencePointer(rv reflect.Value) (reflect.Value, interface{}) {
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

func (l *AppLogger) formatSpecialType(v interface{}) string {
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

func (l *AppLogger) Sync() {
	if l.zap != nil {
		_ = l.zap.Sync()
	}
	// Stop cron scheduler if it exists
	if l.cron != nil {
		ctx := l.cron.Stop()
		<-ctx.Done()
	}
}

func (l *AppLogger) Info(rid, msg string, args ...interface{}) {
	l.log(zapcore.InfoLevel, rid, msg, args...)
}

func (l *AppLogger) Error(rid, msg string, args ...interface{}) {
	l.log(zapcore.ErrorLevel, rid, msg, args...)
}

func (l *AppLogger) Warn(rid, msg string, args ...interface{}) {
	l.log(zapcore.WarnLevel, rid, msg, args...)
}

func (l *AppLogger) LogRequest(req *RequestLogger) {
	l.logRequest(req, "[===== REQUEST INFO =====]", l.CallerConfig.Request.Internal)
}

func (l *AppLogger) LogResponse(resp *ResponseLogger) {
	l.logResponse(resp, "[===== RESPONSE INFO =====]", l.CallerConfig.Response.Internal)
}

func (l *AppLogger) LogExtRequest(req *RequestLogger) {
	l.logRequest(req, "[===== REQUEST EXTERNAL INFO =====]", l.CallerConfig.Request.External)
}

func (l *AppLogger) LogExtResponse(resp *ResponseLogger) {
	l.logResponse(resp, "[===== RESPONSE EXTERNAL INFO =====]", l.CallerConfig.Response.External)
}

func (l *AppLogger) logRequest(req *RequestLogger, message string, callerSkip int) {
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

func (l *AppLogger) logResponse(resp *ResponseLogger, message string, callerSkip int) {
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
