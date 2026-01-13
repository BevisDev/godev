package rest

import "github.com/BevisDev/godev/logx"

var (
	// defaultTimeoutSec defines the default timeout (in seconds) for rest client operations.
	defaultTimeoutSec = 60
)

type HttpConfig struct {
	TimeoutSec      int         // TimeoutSec in seconds
	Logger          logx.Logger // Logger instance for logging
	SkipLogHeader   bool        // Skip logging HTTP headers if true
	SkipLogAPIs     []string    // List of API paths to skip logging
	SkipContentType []string    // List of Content Type to skip logging
}

func (h *HttpConfig) withDefaults() {
	if h.TimeoutSec <= 0 {
		h.TimeoutSec = defaultTimeoutSec
	}
}
