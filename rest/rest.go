package rest

import (
	"net/http"

	"github.com/BevisDev/godev/logx"
)

type HttpConfig struct {
	TimeoutSec      int         // TimeoutSec in seconds
	Logger          logx.Logger // Logger instance for logging
	SkipLogHeader   bool        // Skip logging HTTP headers if true
	SkipLogAPIs     []string    // List of API paths to skip logging
	SkipContentType []string    // List of Content Type to skip logging
}

// defaultTimeoutSec defines the default timeout (in seconds) for rest client operations.
const defaultTimeoutSec = 60

// Client wraps an HTTP client with a configurable timeout and optional logger.
//
// It is intended for making REST API calls with consistent timeout settings
// and optional logging support via AppLogger.
type Client struct {
	*HttpConfig
	client *http.Client
	hasLog bool
}

// NewClient creates a new Client instance using the provided HttpConfig.
// It initializes the internal HTTP client and applies the specified timeout in seconds.
func NewClient(cf *HttpConfig) *Client {
	if cf == nil {
		cf = new(HttpConfig)
	}

	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	c := &Client{
		client:     new(http.Client),
		HttpConfig: cf,
		hasLog:     cf.Logger != nil,
	}
	return c
}

func (r *Client) GetClient() *http.Client {
	return r.client
}
