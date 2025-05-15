package rest

import (
	"errors"
	"fmt"
)

// HttpError represents an HTTP error response with a status code and body.
//
// It implements the `error` interface and can be used to identify
// client-side (4xx) or server-side (5xx) HTTP errors.
type HttpError struct {
	StatusCode int
	Body       string
}

// Error returns the formatted error string including status code and body
func (e *HttpError) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Body)
}

// IsClientError returns true if the status code is in the 4xx range.
func (e *HttpError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if the status code is 500 or higher.
func (e *HttpError) IsServerError() bool {
	return e.StatusCode >= 500
}

// AsHttpError attempts to cast a generic error to *HttpError using errors.As.
//
// Returns the typed error and true if the cast succeeded.
func AsHttpError(err error) (*HttpError, bool) {
	var httpErr *HttpError
	ok := errors.As(err, &httpErr)
	return httpErr, ok
}
