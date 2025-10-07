package rest

import "github.com/BevisDev/godev/logger"

type HttpConfig struct {
	TimeoutSec      int      // timeout in seconds
	logger.Exec              // Logger instance for logging
	SkipLogHeader   bool     // Skip logging HTTP headers if true
	SkipLogAPIs     []string // List of API paths to skip logging
	SkipContentType []string // List of Content DBType to skip logging
}
