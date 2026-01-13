package rest

import (
	"log"
	"net/http"
)

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
	cf.withDefaults()
	
	c := &Client{
		client:     new(http.Client),
		HttpConfig: cf,
		hasLog:     cf.Logger != nil,
	}

	log.Printf("[rest] client started successfully")
	return c
}

func (r *Client) GetClient() *http.Client {
	return r.client
}
