package logger

import (
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils/datetime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type AppLogger struct {
	Logger *zap.Logger
}

type ConfigLogger struct {
	Profile    string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	IsSplit    bool
	DirName    string
	Filename   string
}

type RequestLogger struct {
	State  string
	URL    string
	Time   time.Time
	Query  string
	Method string
	Header any
	Body   any
}

type ResponseLogger struct {
	State       string
	DurationSec time.Duration
	Status      int
	Header      any
	Body        any
}

func NewLogger(cf *ConfigLogger) *AppLogger {
	var zapLogger *zap.Logger
	encoder := getEncoderLog(cf)
	appWrite := writeSync(cf)
	appCore := zapcore.NewCore(encoder, appWrite, zapcore.InfoLevel)
	zapLogger = zap.New(appCore, zap.AddCaller())
	return &AppLogger{Logger: zapLogger}
}

func getEncoderLog(cf *ConfigLogger) zapcore.Encoder {
	var encodeConfig zapcore.EncoderConfig
	// for prod
	if cf.Profile == "prod" {
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

	// for dev
	if cf.Profile == "dev" {
		return zapcore.NewConsoleEncoder(encodeConfig)
	}
	return zapcore.NewJSONEncoder(encodeConfig)
}

func writeSync(cf *ConfigLogger) zapcore.WriteSyncer {
	// handle profile dev
	if cf.Profile == "dev" {
		return zapcore.AddSync(os.Stdout)
	}

	var fileName = getFilename(cf.DirName, cf.Filename, cf.IsSplit)
	lumberLogger := lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    cf.MaxSize,
		MaxBackups: cf.MaxBackups,
		MaxAge:     cf.MaxAge,
		Compress:   cf.Compress,
	}

	// job runner to split log every day
	if cf.IsSplit {
		c := cron.New()
		c.AddFunc("0 0 * * *", func() {
			lumberLogger.Filename = getFilename(cf.DirName, cf.Filename, cf.IsSplit)
			lumberLogger.Rotate()
		})
		c.Start()
	}

	return zapcore.AddSync(&lumberLogger)
}

func getFilename(dir, fileName string, isSplit bool) string {
	if isSplit {
		now := datetime.ToString(time.Now(), datetime.DateOnly)
		return filepath.Join(dir, now, fileName)
	}
	return filepath.Join(dir, fileName)
}

func (l *AppLogger) logApp(level zapcore.Level, state string, msg string, args ...interface{}) {
	if l.Logger == nil {
		return
	}

	// format message
	var message = l.formatMessage(msg, args...)

	// skip caller before
	logging := l.Logger.WithOptions(zap.AddCallerSkip(2))
	switch level {
	case zapcore.InfoLevel:
		logging.Info(message, zap.String(consts.State, state))
	case zapcore.WarnLevel:
		logging.Warn(message, zap.String(consts.State, state))
	case zapcore.ErrorLevel:
		logging.Error(message, zap.String(consts.State, state))
	case zapcore.PanicLevel:
		logging.Panic(message, zap.String(consts.State, state))
	case zapcore.FatalLevel:
		logging.Fatal(message, zap.String(consts.State, state))
	default:
		logging.Info(message, zap.String(consts.State, state))
	}
}

func (l *AppLogger) formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}

	var message string
	if strings.Contains(msg, "{}") {
		message = strings.ReplaceAll(msg, "{}", "%+v")
	} else if strings.Contains(msg, "%") {
		message = msg
	} else {
		msg += strings.Repeat(":%+v", len(args))
		message = msg
	}

	return fmt.Sprintf(message, args...)
}

func (l *AppLogger) Sync() {
	if l.Logger != nil {
		l.Logger.Sync()
	}
}

func (l *AppLogger) Info(state, msg string, args ...interface{}) {
	l.logApp(zapcore.InfoLevel, state, msg, args...)
}

func (l *AppLogger) Error(state, msg string, args ...interface{}) {
	l.logApp(zapcore.ErrorLevel, state, msg, args...)
}

func (l *AppLogger) Warn(state, msg string, args ...interface{}) {
	l.logApp(zapcore.WarnLevel, state, msg, args...)
}

func (l *AppLogger) Panic(state, msg string, args ...interface{}) {
	l.logApp(zapcore.PanicLevel, state, msg, args...)
}

func (l *AppLogger) Fatal(state, msg string, args ...interface{}) {
	l.logApp(zapcore.FatalLevel, state, msg, args...)
}

func (l *AppLogger) LogRequest(req *RequestLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== REQUEST INFO =====]",
		zap.String("State", req.State),
		zap.String("URL", req.URL),
		zap.Time("Time", req.Time),
		zap.String("Method", req.Method),
		zap.String("Query", req.Query),
		zap.Any("Header", req.Header),
		zap.Any("Body", req.Body),
	)
}

func (l *AppLogger) LogResponse(resp *ResponseLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== RESPONSE INFO =====]",
		zap.String("State", resp.State),
		zap.Int("Status", resp.Status),
		zap.Float64("DurationSec", resp.DurationSec.Seconds()),
		zap.Any("Header", resp.Header),
		zap.Any("Body", resp.Body),
	)
}

func (l *AppLogger) LogExtRequest(req *RequestLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(2)).Info(
		"[===== REQUEST EXTERNAL INFO =====]",
		zap.String("State", req.State),
		zap.String("URL", req.URL),
		zap.Time("Time", req.Time),
		zap.String("Method", req.Method),
		zap.String("Query", req.Query),
		zap.Any("Header", req.Header),
		zap.Any("Body", req.Body),
	)
}

func (l *AppLogger) LogExtResponse(resp *ResponseLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== RESPONSE EXTERNAL INFO =====]",
		zap.String("State", resp.State),
		zap.Int("Status", resp.Status),
		zap.Float64("DurationSec", resp.DurationSec.Seconds()),
		zap.Any("Header", resp.Header),
		zap.Any("Body", resp.Body),
	)
}
