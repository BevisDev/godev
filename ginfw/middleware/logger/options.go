package logger

import "github.com/BevisDev/godev/logx"

type Option func(*options)

type options struct {
	// logger instance for logging
	logger logx.Logger

	// useLog is the flag use logx
	useLog bool

	// skipHeader Skip logging HTTP headers if true
	skipHeader bool

	// skipDefaultContentTypeCheck disables the default content-type based body logging checks.
	skipDefaultContentTypeCheck bool
}

func withDefaults() *options {
	return &options{
		useLog:                      false,
		skipHeader:                  false,
		skipDefaultContentTypeCheck: false,
	}
}

func WithLogger(l logx.Logger) Option {
	return func(o *options) {
		o.logger = l
		o.useLog = true
	}
}

func WithSkipHeader() Option {
	return func(o *options) {
		o.skipHeader = true
	}
}

func WithSkipDefaultContentTypeCheck() Option {
	return func(o *options) {
		o.skipDefaultContentTypeCheck = true
	}
}
