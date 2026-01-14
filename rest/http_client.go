package rest

import (
	"context"
)

type HttpClient[T any] interface {
	// URL sets the request URL.
	URL(url string) HttpClient[T]

	// QueryParams contains query parameters to be appended to the URL (?key=value).
	QueryParams(query map[string]string) HttpClient[T]

	// PathParams contains path parameters to replace placeholders in the URL (e.g., ":id").
	PathParams(params map[string]string) HttpClient[T]

	// Headers sets HTTP headers for the request.
	Headers(headers map[string]string) HttpClient[T]

	// Body sets the request body (JSON or raw).
	Body(body any) HttpClient[T]

	// BodyForm sets form-encoded body parameters.
	BodyForm(bodyForm map[string]string) HttpClient[T]

	// GET sends an HTTP GET request using the provided context.
	// It returns an error if the request fails or the response cannot be processed.
	GET(c context.Context) (Response[T], error)

	// POST sends an HTTP POST request with a JSON or byte body using the provided context.
	// Returns an error if the request fails or response processing fails
	POST(c context.Context) (Response[T], error)

	// PostForm sends an HTTP POST request with form-data (application/x-www-form-urlencoded) body.
	// The body is taken from BodyForm. Returns an error on failure.
	PostForm(c context.Context) (Response[T], error)

	// PUT sends an HTTP PUT request using the provided context.
	// Returns an error if the request or response handling fails.
	PUT(c context.Context) (Response[T], error)

	// PATCH sends an HTTP PATCH request using the provided context.
	// Returns an error on failure.
	PATCH(c context.Context) (Response[T], error)

	// DELETE sends an HTTP DELETE request using the provided context.
	// Returns an error if the request fails or response cannot be handled.
	DELETE(c context.Context) (Response[T], error)
}
