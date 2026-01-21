package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/datetime"
	"github.com/BevisDev/godev/utils/random"
	"github.com/gin-gonic/gin"
)

type HttpLogger struct {
	*options
}

type responseWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func New(fs ...OptionFunc) *HttpLogger {
	o := withDefaults()
	for _, f := range fs {
		f(o)
	}

	return &HttpLogger{
		options: o,
	}
}

func (h *HttpLogger) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Generate RID per request and attach to context
		rid := random.NewUUID()
		ctx := utils.SetValueCtx(c.Request.Context(), consts.RID, rid)
		c.Request = c.Request.WithContext(ctx)

		// Read and log request
		reqBody := h.readRequestBody(c)
		h.logRequest(c, rid, startTime, reqBody)

		// Wrap response writer to capture response body
		buf := h.wrapResponseWriter(c)

		// Process request
		c.Next()

		// Log response
		duration := time.Since(startTime)
		resBody := h.readResponseBody(buf, c.Writer.Header().Get(consts.ContentType))
		h.logResponse(c, rid, duration, resBody)
	}
}

func (h *HttpLogger) readRequestBody(c *gin.Context) string {
	contentType := c.Request.Header.Get(consts.ContentType)
	if h.skipDefaultContentTypeCheck || !utils.SkipContentType(contentType) {
		raw, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		return string(raw)
	}
	return ""
}

func (h *HttpLogger) wrapResponseWriter(c *gin.Context) *bytes.Buffer {
	buf := &bytes.Buffer{}
	c.Writer = &responseWrapper{
		ResponseWriter: c.Writer,
		body:           buf,
	}
	return buf
}

func (h *HttpLogger) readResponseBody(buf *bytes.Buffer, contentType string) string {
	if h.skipDefaultContentTypeCheck || !utils.SkipContentType(contentType) {
		return buf.String()
	}
	return ""
}

func (h *HttpLogger) logRequest(c *gin.Context, rid string, startTime time.Time, reqBody string) {
	if h.useLog {
		h.logRequestWithLogx(c, rid, startTime, reqBody)
	} else {
		h.logRequestConsole(c, rid, startTime, reqBody)
	}
}

func (h *HttpLogger) logRequestWithLogx(c *gin.Context, rid string, startTime time.Time, reqBody string) {
	reqLog := &logx.RequestLogger{
		RID:    rid,
		URL:    c.Request.URL.String(),
		Time:   startTime,
		Query:  c.Request.URL.RawQuery,
		Method: c.Request.Method,
		Body:   reqBody,
	}
	if !h.skipHeader {
		reqLog.Header = c.Request.Header
	}
	h.logger.LogRequest(reqLog)
}

func (h *HttpLogger) logRequestConsole(c *gin.Context, rid string, startTime time.Time, reqBody string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n========== REQUEST INFO ==========\n")
	fmt.Fprintf(&sb, "%s: %s\n", consts.RID, rid)
	fmt.Fprintf(&sb, "%s: %s\n", consts.Url, c.Request.URL.String())
	fmt.Fprintf(&sb, "%s: %s\n", consts.Method, c.Request.Method)
	fmt.Fprintf(&sb, "%s: %s\n", consts.Time,
		datetime.ToString(startTime, datetime.DateTimeLayoutMilli))
	fmt.Fprintf(&sb, "%s: %v\n", consts.Query, c.Request.URL.RawQuery)
	if !h.skipHeader {
		fmt.Fprintf(&sb, "%s: %s\n", consts.Header, c.Request.Header)
	}
	fmt.Fprintf(&sb, "%s: %s\n", consts.Body, reqBody)
	fmt.Fprintf(&sb, "==================================\n")
	log.Println(sb.String())
}

func (h *HttpLogger) logResponse(c *gin.Context, rid string, duration time.Duration, resBody string) {
	if h.useLog {
		h.logResponseWithLogx(c, rid, duration, resBody)
	} else {
		h.logResponseConsole(c, rid, duration, resBody)
	}
}

func (h *HttpLogger) logResponseWithLogx(c *gin.Context, rid string, duration time.Duration, resBody string) {
	resLog := &logx.ResponseLogger{
		RID:      rid,
		Status:   c.Writer.Status(),
		Duration: duration,
		Body:     resBody,
	}
	if !h.skipHeader {
		resLog.Header = c.Writer.Header()
	}
	h.logger.LogResponse(resLog)
}

func (h *HttpLogger) logResponseConsole(c *gin.Context, rid string, duration time.Duration, resBody string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n========== RESPONSE INFO ==========\n")
	fmt.Fprintf(&sb, "%s: %s\n", consts.RID, rid)
	fmt.Fprintf(&sb, "%s: %d\n", consts.Status, c.Writer.Status())
	fmt.Fprintf(&sb, "%s: %s\n", consts.Duration, duration)
	if !h.skipHeader {
		fmt.Fprintf(&sb, "%s: %v\n", consts.Header, c.Writer.Header())
	}
	fmt.Fprintf(&sb, "%s: %s\n", consts.Body, resBody)
	fmt.Fprintf(&sb, "==================================\n")
	log.Println(sb.String())
}
