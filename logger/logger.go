package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BevisDev/godev/helper"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type AppLogger struct {
	logger *zap.Logger
}

type ConfigLogger struct {
	Profile    string
	DirName    string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	IsSplit    bool
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
	return &AppLogger{logger: zapLogger}
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

	now := time.Now().Format(helper.YYYY_MM_DD)
	lumberLogger := lumberjack.Logger{
		Filename:   filepath.Join(cf.DirName, now, "app.log"),
		MaxSize:    cf.MaxSize,
		MaxBackups: cf.MaxBackups,
		MaxAge:     cf.MaxAge,
		Compress:   cf.Compress,
	}

	// job runner to split log every day
	if cf.IsSplit {
		c := cron.New()
		c.AddFunc("0 0 * * *", func() {
			lumberLogger.Filename = filepath.Join(cf.DirName, now, "app.log")
			lumberLogger.Close()
		})
		c.Start()
	}

	return zapcore.AddSync(&lumberLogger)
}

func (l *AppLogger) logApp(level zapcore.Level, state string, msg string, args ...interface{}) {
	if l.logger == nil {
		return
	}
	// formater message
	var message string
	if len(args) != 0 {
		message = l.formatMessage(msg, args...)
	} else {
		message = msg
	}
	// skip caller before
	logging := l.logger.WithOptions(zap.AddCallerSkip(2))
	switch level {
	case zapcore.InfoLevel:
		logging.Info(message, zap.String("state", state))
		break
	case zapcore.WarnLevel:
		logging.Warn(message, zap.String("state", state))
		break
	case zapcore.ErrorLevel:
		logging.Error(message, zap.String("state", state))
		break
	case zapcore.FatalLevel:
		logging.Fatal(message, zap.String("state", state))
		break
	default:
		logging.Info(message, zap.String("state", state))
	}
}

func (l *AppLogger) formatMessage(msg string, args ...interface{}) string {
	var message string
	if !strings.Contains(msg, "%") && strings.Contains(msg, "{}") {
		message = strings.ReplaceAll(msg, "{}", "%+v")
	} else {
		message = msg
	}
	return fmt.Sprintf(message, args...)
}

func (l *AppLogger) SyncAll() {
	if l.logger != nil {
		l.logger.Sync()
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
	l.logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== REQUEST INFO =====]",
		zap.String("state", req.State),
		zap.String("url", req.URL),
		zap.Time("time", req.Time),
		zap.String("method", req.Method),
		zap.String("query", req.Query),
		zap.Any("header", req.Header),
		zap.Any("body", req.Body),
	)
}

func (l *AppLogger) LogResponse(resp *ResponseLogger) {
	l.logger.WithOptions(
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
	l.logger.WithOptions(
		zap.AddCallerSkip(2)).Info(
		"[===== REQUEST EXTERNAL INFO =====]",
		zap.String("state", req.State),
		zap.String("url", req.URL),
		zap.Time("time", req.Time),
		zap.String("method", req.Method),
		zap.String("query", req.Query),
		zap.Any("body", req.Body),
	)
}

func (l *AppLogger) LogExtResponse(resp *ResponseLogger) {
	l.logger.WithOptions(
		zap.AddCallerSkip(1)).Info(
		"[===== RESPONSE EXTERNAL INFO =====]",
		zap.String("state", resp.State),
		zap.Int("status", resp.Status),
		zap.Float64("durationSec", resp.DurationSec.Seconds()),
		zap.Any("body", resp.Body),
	)
}
