package rest

import (
	"context"
)

type HTTPClient[T any] interface {
	// URL sets the httpRequest URL.
	URL(url string) HTTPClient[T]

	// QueryParams contains query parameters to be appended to the URL (?key=value).
	QueryParams(query map[string]string) HTTPClient[T]

	// PathParams contains path parameters to replace placeholders in the URL (e.g., ":id").
	PathParams(params map[string]string) HTTPClient[T]

	// Headers sets HTTP headers for the httpRequest.
	Headers(headers map[string]string) HTTPClient[T]

	// Body sets the httpRequest body (JSON or raw).
	Body(body any) HTTPClient[T]

	// BodyForm sets form-encoded body parameters.
	BodyForm(bodyForm map[string]string) HTTPClient[T]

	// GET sends an HTTP GET httpRequest using the provided context.
	// It returns an error if the httpRequest fails or the response cannot be processed.
	GET(c context.Context) (HTTPResponse[T], error)

	// POST sends an HTTP POST httpRequest with a JSON or byte body using the provided context.
	// Returns an error if the httpRequest fails or response processing fails
	POST(c context.Context) (HTTPResponse[T], error)

	// PostForm sends an HTTP POST httpRequest with form-data (application/x-www-form-urlencoded) body.
	// The body is taken from BodyForm. Returns an error on failure.
	PostForm(c context.Context) (HTTPResponse[T], error)

	// PUT sends an HTTP PUT httpRequest using the provided context.
	// Returns an error if the httpRequest or response handling fails.
	PUT(c context.Context) (HTTPResponse[T], error)

	// PATCH sends an HTTP PATCH httpRequest using the provided context.
	// Returns an error on failure.
	PATCH(c context.Context) (HTTPResponse[T], error)

	// DELETE sends an HTTP DELETE httpRequest using the provided context.
	// Returns an error if the httpRequest fails or response cannot be handled.
	DELETE(c context.Context) (HTTPResponse[T], error)
}
