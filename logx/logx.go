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
	zap *zap.Logger
}

// NewLogger initializes and returns a new application logger (`*AppLogger`) using the Zap logging library.
//
// It configures the log encoder format (e.g., JSON or console), the log output (e.g., file path),
// and log rotation settings based on the provided `ConfigLogger`.
//
// The logger includes caller information (`zap.AddCaller`) and uses `zapcore.InfoLevel` by default.
// Log rotation is handled via Lumberjack based on `MaxSize`, `MaxBackups`, `MaxAge`, and `Compress`.
//
// Example:
//
//	logger := NewLogger(&Config{
//	    isProduction: true,
//	    MaxSize		: 100,             // 100 MB per file
//	    MaxBackups	: 7,               // keep 7 rotated logs
//	    MaxAge		: 30,              // keep logs for 30 days
//	    Compress	: true,            // compress old logs
//	    IsRotate	: false,           // no daily split
//	    DirName		: "./logs",
//	    Filename	: "app.log",
//	})
//
//	logger.Info("Application started")
func NewLogger(cf *Config) Logger {
	var l = &AppLogger{Config: cf}
	encoder := l.getEncoderLog()
	writer := l.writeSync()

	var z = new(zap.Logger)
	z = zap.New(
		zapcore.NewCore(
			encoder,
			writer,
			zapcore.InfoLevel,
		),
		zap.AddCaller(),
	)
	l.zap = z

	l.zap.Info("[logger] started successfully")
	return l
}

func (l *AppLogger) GetZap() *zap.Logger {
	return l.zap
}

func (l *AppLogger) getEncoderLog() zapcore.Encoder {
	var encodeConfig zapcore.EncoderConfig
	// for production
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
	// for other
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

	var fileName = getFilename(l.DirName, l.Filename, l.IsRotate)
	lumber := lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    l.MaxSize,
		MaxBackups: l.MaxBackups,
		MaxAge:     l.MaxAge,
		Compress:   l.Compress,
	}

	// job runner to rotate log every day
	if l.IsRotate {
		c := cron.New()
		c.AddFunc("0 0 * * *", func() {
			lumber.Filename = getFilename(l.DirName, l.Filename, l.IsRotate)
			err := lumber.Rotate()
			if err != nil {
				log.Println(err)
				return
			}
		})
		c.Start()
	}

	return zapcore.AddSync(&lumber)
}

func getFilename(dir, fileName string, isSplit bool) string {
	if isSplit {
		now := datetime.ToString(time.Now(), datetime.DateLayoutISO)
		return filepath.Join(dir, now, fileName)
	}
	return filepath.Join(dir, fileName)
}

func (l *AppLogger) mustHaveZap() {
	if l.zap == nil {
		panic("logger is not initialized")
	}
}

func (l *AppLogger) log(level zapcore.Level,
	rid, msg string,
	args ...interface{},
) {
	l.mustHaveZap()

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
	case zapcore.PanicLevel:
		logging.Panic(message, fields...)
	case zapcore.FatalLevel:
		logging.Fatal(message, fields...)
	default:
		logging.Info(message, fields...)
	}
}

func (l *AppLogger) formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	numArgs := len(args)

	if strings.Contains(msg, "{}") {
		count := strings.Count(msg, "{}")
		if count < numArgs {
			msg += strings.Repeat(" :{}", numArgs-count)
		}
		message := strings.ReplaceAll(msg, "{}", "%+v")
		return fmt.Sprintf(message, l.deferArgs(args...)...)
	}

	if strings.Contains(msg, "%") {
		for _, arg := range args {
			msg += " :" + l.formatAny(arg)
		}
		return msg
	}

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

	// -- error
	if err, ok := v.(error); ok {
		return err.Error()
	}

	rv := reflect.ValueOf(v)

	// Deref multiple levels if it's pointer
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return "<nil>"
		}
		rv = rv.Elem()
		if !rv.IsValid() {
			return "<invalid>"
		}
		v = rv.Interface()
	}

	// special type
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
	}

	// struct, map, slice: serialize via JSON
	switch rv.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return jsonx.ToJSON(v)
	default:
		if rv.CanInterface() {
			return fmt.Sprintf("%+v", rv.Interface())
		}
		return fmt.Sprintf("<unreadable: %T>", v)
	}
}

func (l *AppLogger) Sync() {
	if l.zap != nil {
		_ = l.zap.Sync()
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

func (l *AppLogger) Panic(rid, msg string, args ...interface{}) {
	l.log(zapcore.PanicLevel, rid, msg, args...)
}

func (l *AppLogger) Fatal(rid, msg string, args ...interface{}) {
	l.log(zapcore.FatalLevel, rid, msg, args...)
}

func (l *AppLogger) LogRequest(req *RequestLogger) {
	l.mustHaveZap()

	fields := []zap.Field{
		zap.String(consts.RID, req.RID),
		zap.String(consts.Url, req.URL),
		zap.Time(consts.Time, req.Time),
		zap.String(consts.Method, req.Method),
	}
	if req.Query != "" {
		fields = append(fields, zap.String(consts.Query, req.Query))
	}
	if req.Header != nil {
		fields = append(fields, zap.Any(consts.Header, req.Header))
	}
	if req.Body != "" {
		fields = append(fields, zap.String(consts.Body, req.Body))
	}

	l.zap.WithOptions(
		zap.AddCallerSkip(l.CallerConfig.Request.Internal)).
		Info(
			"[===== REQUEST INFO =====]",
			fields...,
		)
}

func (l *AppLogger) LogResponse(resp *ResponseLogger) {
	l.mustHaveZap()

	fields := []zap.Field{
		zap.String(consts.RID, resp.RID),
		zap.Int(consts.Status, resp.Status),
		zap.String(consts.Duration, resp.Duration.String()),
	}
	if resp.Header != nil {
		fields = append(fields, zap.Any(consts.Header, resp.Header))
	}
	if resp.Body != "" {
		fields = append(fields, zap.String(consts.Body, resp.Body))
	}

	l.zap.WithOptions(
		zap.AddCallerSkip(l.CallerConfig.Response.Internal)).
		Info(
			"[===== RESPONSE INFO =====]",
			fields...,
		)
}

func (l *AppLogger) LogExtRequest(req *RequestLogger) {
	l.mustHaveZap()

	fields := []zap.Field{
		zap.String(consts.RID, req.RID),
		zap.String(consts.Url, req.URL),
		zap.Time(consts.Time, req.Time),
		zap.String(consts.Method, req.Method),
	}
	if req.Query != "" {
		fields = append(fields, zap.String(consts.Query, req.Query))
	}
	if req.Header != nil {
		fields = append(fields, zap.Any(consts.Header, req.Header))
	}
	if req.Body != "" {
		fields = append(fields, zap.String(consts.Body, req.Body))
	}

	l.zap.WithOptions(
		zap.AddCallerSkip(l.CallerConfig.Request.External)).
		Info(
			"[===== REQUEST EXTERNAL INFO =====]",
			fields...,
		)
}

func (l *AppLogger) LogExtResponse(resp *ResponseLogger) {
	l.mustHaveZap()

	fields := []zap.Field{
		zap.String(consts.RID, resp.RID),
		zap.Int(consts.Status, resp.Status),
		zap.String(consts.Duration, resp.Duration.String()),
	}
	if resp.Header != nil {
		fields = append(fields, zap.Any(consts.Header, resp.Header))
	}
	if resp.Body != "" {
		fields = append(fields, zap.String(consts.Body, resp.Body))
	}

	l.zap.WithOptions(
		zap.AddCallerSkip(l.CallerConfig.Response.External)).
		Info(
			"[===== RESPONSE EXTERNAL INFO =====]",
			fields...,
		)
}
