package response

import (
	"context"
	"net/http"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/datetime"
	"github.com/gin-gonic/gin"
)

var Code = map[string]string{
	"400": "Invalid Request",
	"401": "Unauthorized",
	"403": "Forbidden",
	"404": "Not Found",
	"405": "Method Not Allowed",
	"409": "Conflict",
	"429": "Too Many Requests",
	"500": "Internal Server Error",
	"503": "Service Unavailable",
	"504": "Gateway Timeout",
}

// Response represents a standardized API response structure.
type Response[T any] struct {
	RID        string `json:"rid,omitempty"`
	Success    bool   `json:"success"`
	Data       T      `json:"data,omitempty"`
	ResponseAt string `json:"response_at,omitempty"`
	Error      *Error `json:"error,omitempty"`
}

// NewSuccess creates a successful response with the provided data.
func NewSuccess[T any](ctx context.Context, data T) *Response[T] {
	return &Response[T]{
		RID:        utils.GetRID(ctx),
		Success:    true,
		Data:       data,
		ResponseAt: datetime.ToString(time.Now(), datetime.DateTimeLayout),
	}
}

// NewFailure creates a failure response with error code and message.
func NewFailure(ctx context.Context, code, message string) *Response[any] {
	return &Response[any]{
		RID:        utils.GetRID(ctx),
		Success:    false,
		ResponseAt: datetime.ToString(time.Now(), datetime.DateTimeLayout),
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}

// Error represents an error in the API response.
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func GetCode(code, message string, defCode string) (string, string) {
	if code == "" {
		code = defCode
	}
	if message == "" {
		message = Code[code]
	}
	return code, message
}

// Success sends a 200 OK response with the provided data.
func Success[T any](c *gin.Context, data T) {
	res := NewSuccess[T](c.Request.Context(), data)
	c.JSON(http.StatusOK, res)
}

// Created sends a 201 Created response with the provided data.
func Created[T any](c *gin.Context, data T) {
	res := NewSuccess[T](c.Request.Context(), data)
	c.JSON(http.StatusCreated, res)
}

// Accepted sends a 202 Accepted response.
func Accepted(c *gin.Context) {
	res := NewSuccess[any](c.Request.Context(), nil)
	c.JSON(http.StatusAccepted, res)
}

// NotModified sends a 304 Not Modified response.
func NotModified(c *gin.Context) {
	res := NewSuccess[any](c.Request.Context(), nil)
	c.JSON(http.StatusNotModified, res)
}

// BadRequest sends a 400 Bad Request response with error code and message.
func BadRequest(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "400")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusBadRequest, res)
}

// Unauthorized sends a 401 Unauthorized response with error code and message.
func Unauthorized(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "401")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusUnauthorized, res)
}

// Forbidden sends a 403 Forbidden response with error code and message.
func Forbidden(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "403")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusForbidden, res)
}

// NotFound sends a 404 Not Found response with error code and message.
func NotFound(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "404")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusNotFound, res)
}

// MethodNotAllow sends a 405 Method Not Allowed response with error code and message.
func MethodNotAllow(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "405")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusMethodNotAllowed, res)
}

// Conflict sends a 409 Conflict response with error code and message.
func Conflict(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "409")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusConflict, res)
}

// TooManyRequests sends a 429 Too Many Requests response with error code and message.
func TooManyRequests(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "429")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusTooManyRequests, res)
}

// ServerError sends a 500 Internal Server Error response with error code and message.
func ServerError(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "500")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusInternalServerError, res)
}

// ServiceUnavailable sends a 503 Service Unavailable response with error code and message.
func ServiceUnavailable(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "503")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusServiceUnavailable, res)
}

// ServerTimeout sends a 504 Gateway Timeout response with error code and message.
func ServerTimeout(c *gin.Context, code, message string) {
	code, message = GetCode(code, message, "504")
	res := NewFailure(c.Request.Context(), code, message)
	c.JSON(http.StatusGatewayTimeout, res)
}
