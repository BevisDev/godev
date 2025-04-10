package logger

import (
	"fmt"
	"github.com/BevisDev/godev/constants"
	"github.com/BevisDev/godev/utils"
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

	lumberLogger := lumberjack.Logger{
		Filename:   getFilename(cf.DirName, cf.Filename),
		MaxSize:    cf.MaxSize,
		MaxBackups: cf.MaxBackups,
		MaxAge:     cf.MaxAge,
		Compress:   cf.Compress,
	}

	// job runner to split log every day
	if cf.IsSplit {
		c := cron.New()
		c.AddFunc("0 0 * * *", func() {
			lumberLogger.Filename = getFilename(cf.DirName, cf.Filename)
			lumberLogger.Rotate()
		})
		c.Start()
	}

	return zapcore.AddSync(&lumberLogger)
}

func getFilename(dir, fileName string) string {
	now := time.Now().Format(utils.DateOnly)
	return filepath.Join(dir, now, fileName)
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
		logging.Info(message, zap.String(constants.State, state))
		break
	case zapcore.WarnLevel:
		logging.Warn(message, zap.String(constants.State, state))
		break
	case zapcore.ErrorLevel:
		logging.Error(message, zap.String(constants.State, state))
		break
	case zapcore.FatalLevel:
		logging.Fatal(message, zap.String(constants.State, state))
		break
	default:
		logging.Info(message, zap.String(constants.State, state))
	}
}

func (l *AppLogger) formatMessage(msg string, args ...interface{}) string {
	var message string
	if len(args) == 0 {
		return msg
	}

	if strings.Contains(msg, "{}") {
		message = strings.ReplaceAll(msg, "{}", "%+v")
	}
	if !strings.Contains(msg, "%") {
		msg += strings.Repeat(" :%+v", len(args))
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

func (l *AppLogger) Fatal(state, msg string, args ...interface{}) {
	l.logApp(zapcore.FatalLevel, state, msg, args...)
}

func (l *AppLogger) LogRequest(req *RequestLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== REQUEST INFO =====]",
		zap.String(constants.State, req.State),
		zap.String("url", req.URL),
		zap.Time("time", req.Time),
		zap.String("method", req.Method),
		zap.String("query", req.Query),
		zap.Any("header", req.Header),
		zap.Any("body", req.Body),
	)
}

func (l *AppLogger) LogResponse(resp *ResponseLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== RESPONSE INFO =====]",
		zap.String("state", resp.State),
		zap.Int("status", resp.Status),
		zap.Float64("durationSec", resp.DurationSec.Seconds()),
		zap.Any("header", resp.Header),
		zap.Any("body", resp.Body),
	)
}

func (l *AppLogger) LogExtRequest(req *RequestLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(2)).Info(
		"[===== REQUEST EXTERNAL INFO =====]",
		zap.String(constants.State, req.State),
		zap.String("url", req.URL),
		zap.Time("time", req.Time),
		zap.String("method", req.Method),
		zap.String("query", req.Query),
		zap.Any("body", req.Body),
	)
}

func (l *AppLogger) LogExtResponse(resp *ResponseLogger) {
	l.Logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== RESPONSE EXTERNAL INFO =====]",
		zap.String(constants.State, resp.State),
		zap.Int("status", resp.Status),
		zap.Float64("durationSec", resp.DurationSec.Seconds()),
		zap.Any("body", resp.Body),
	)
}
