package rest

import (
	"errors"
	"fmt"
)

// HTTPError represents an HTTP error response with a status code and body.
//
// It implements the `error` interface and can be used to identify
// client-side (4xx) or server-side (5xx) HTTP errors.
type HTTPError struct {
	Status int
	Body   string
}

// Error returns the formatted error string including status code and body
func (e *HTTPError) Error() string {
	return fmt.Sprintf("status %d: %s", e.Status, e.Body)
}

// IsClientError returns true if the status code is in the 4xx range.
func (e *HTTPError) IsClientError() bool {
	return e.Status >= 400 && e.Status < 500
}

// IsServerError returns true if the status code is 500 or higher.
func (e *HTTPError) IsServerError() bool {
	return e.Status >= 500
}

// AsHTTPError attempts to cast a generic error to *HTTPError using errors.As.
//
// Returns the typed error and true if the cast succeeded.
func AsHTTPError(err error) (*HTTPError, bool) {
	var httpErr *HTTPError
	ok := errors.As(err, &httpErr)
	return httpErr, ok
}
