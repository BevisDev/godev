package httplogger

import "github.com/BevisDev/godev/logger"

type Option func(*options)

type options struct {
	logger *logger.Logger

	// useStructuredLogger routes logs through logger.Logger instead of the std log package.
	useStructuredLogger bool

	// skipHeader omits HTTP headers from log output when true.
	skipHeader bool

	// skipDefaultContentTypeCheck disables content-type based body logging rules from utils.SkipContentType.
	skipDefaultContentTypeCheck bool
}

func defaultOptions() *options {
	return &options{
		useStructuredLogger:         false,
		skipHeader:                  false,
		skipDefaultContentTypeCheck: false,
	}
}

func WithLogger(l *logger.Logger) Option {
	return func(o *options) {
		if l != nil {
			o.logger = l
			o.useStructuredLogger = true
		}
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
