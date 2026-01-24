package rest

import (
	"log"
	"net/http"
	"time"
)

// Client wraps an HTTP client with a configurable timeout and optional logger.
//
// It is intended for making REST API calls with consistent timeout settings
// and optional logging support via AppLogger.
type Client struct {
	opt    *options
	client *http.Client
}

// New creates a new Client instance using the provided Options.
// It initializes the internal HTTP client and applies the specified timeout in seconds.
func New(opts ...Option) *Client {
	opt := withDefaults()
	for _, op := range opts {
		op(opt)
	}

	c := &Client{
		client: new(http.Client),
		opt:    opt,
	}

	log.Printf("[rest] client started successfully")
	return c
}

func (r *Client) GetClient() *http.Client {
	return r.client
}

func (r *Client) SetTimeout(timeout time.Duration) {
	r.opt.timeout = timeout
}
