package rest

import (
	"net/http"
	"time"
)

type Response[T any] struct {
	StatusCode int
	Header     http.Header
	Data       T
	Duration   time.Duration
	RawBody    []byte
	Body       string
	HasBody    bool
}
