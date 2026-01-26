package rest

import (
	"time"

	"github.com/BevisDev/godev/logger"
)

type Option func(*options)

type options struct {
	// timeout for rest client operations.
	timeout time.Duration

	// logger instance for logging
	logger *logger.Logger

	// useLog is the flag use logger
	useLog bool

	// skipHeader Skip logging HTTP headers if true
	skipHeader bool

	// skipBodyByPaths defines API paths for which request/response bodies should not be logged.
	skipBodyByPaths map[string]struct{}

	// skipBodyContentTypes defines content types for which bodies should not be logged.
	skipBodyByContentTypes map[string]struct{}

	// skipDefaultContentTypeCheck disables the default content-type based body logging checks.
	skipDefaultContentTypeCheck bool
}

func withDefaults() *options {
	return &options{
		timeout:                5 * time.Second,
		skipBodyByPaths:        make(map[string]struct{}),
		skipBodyByContentTypes: make(map[string]struct{}),
	}
}

func WithLogger(logger *logger.Logger) Option {
	return func(o *options) {
		if o.logger == nil && o.logger != nil {
			o.logger = logger
			o.useLog = true
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.timeout = timeout
		}
	}
}

func WithSkipHeader() Option {
	return func(o *options) {
		o.skipHeader = true
	}
}

func WithSkipBodyByPaths(paths ...string) Option {
	return func(o *options) {
		for _, p := range paths {
			o.skipBodyByPaths[p] = struct{}{}
		}
	}
}

func WithSkipBodyByContentTypes(contentTypes ...string) Option {
	return func(o *options) {
		for _, c := range contentTypes {
			o.skipBodyByContentTypes[c] = struct{}{}
		}
	}
}

func WithSkipDefaultContentTypeCheck() Option {
	return func(o *options) {
		o.skipDefaultContentTypeCheck = true
	}
}
