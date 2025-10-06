package rest

import (
	"net/http"
)

// defaultTimeoutSec defines the default timeout (in seconds) for rest client operations.
const defaultTimeoutSec = 60

// RestClient wraps an HTTP client with a configurable timeout and optional logger.
//
// It is intended for making REST API calls with consistent timeout settings
// and optional logging support via AppLogger.
type RestClient struct {
	*HttpConfig
	client *http.Client
}

// New creates a new RestClient instance using the provided HttpConfig.
// It initializes the internal HTTP client and applies the specified timeout in seconds.
func New(cf *HttpConfig) *RestClient {
	if cf == nil {
		cf = new(HttpConfig)
	}

	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	restClient := &RestClient{
		client:     new(http.Client),
		HttpConfig: cf,
	}
	return restClient
}

func (r *RestClient) GetClient() *http.Client {
	return r.client
}
