package rest

import "github.com/BevisDev/godev/logx"

type HttpConfig struct {
	TimeoutSec      int         // TimeoutSec in seconds
	Logger          logx.Logger // Logger instance for logging
	SkipLogHeader   bool        // Skip logging HTTP headers if true
	SkipLogAPIs     []string    // List of API paths to skip logging
	SkipContentType []string    // List of Content Type to skip logging
}
