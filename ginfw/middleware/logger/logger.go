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

type responseWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger(fs ...OptionFunc) gin.HandlerFunc {
	o := withDefaults()
	for _, f := range fs {
		f(o)
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		// generate state per request and attach state to context
		var state = random.NewUUID()
		ctx := utils.SetValueCtx(c.Request.Context(), consts.State, state)
		c.Request = c.Request.WithContext(ctx)

		// ===== REQUEST LOG =====
		var contentType = c.Request.Header.Get(consts.ContentType)
		var reqBody string
		if o.skipDefaultContentTypeCheck || !utils.SkipContentType(contentType) {
			raw, _ := io.ReadAll(c.Request.Body)
			reqBody = string(raw)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		}

		if o.useLog {
			reqLog := &logx.RequestLogger{
				State:       state,
				URL:         c.Request.URL.String(),
				RequestTime: startTime,
				Query:       c.Request.URL.RawQuery,
				Method:      c.Request.Method,
				Body:        reqBody,
			}
			if !o.skipHeader {
				reqLog.Header = c.Request.Header
			}
			o.logger.LogRequest(reqLog)
		} else {
			var sb strings.Builder
			sb.WriteString("\n========== REQUEST INFO ==========\n")
			sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
			sb.WriteString(fmt.Sprintf(consts.Url+": %s\n", c.Request.URL.String()))
			sb.WriteString(fmt.Sprintf(consts.Method+": %s\n", c.Request.Method))
			sb.WriteString(fmt.Sprintf(consts.RequestTime+": %s\n",
				datetime.ToString(startTime, datetime.DateTimeLayoutMilli)))
			sb.WriteString(fmt.Sprintf(consts.Query+": %v\n", c.Request.URL.RawQuery))
			if !o.skipHeader {
				sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", c.Request.Header))
			}
			if reqBody != "" {
				sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", reqBody))
			}
			sb.WriteString("==================================\n")
			log.Println(sb.String())
		}

		// ===== RESPONSE WRAP =====
		// wrap the responseWriter to capture the response body
		buf := &bytes.Buffer{}
		writer := &responseWrapper{
			ResponseWriter: c.Writer,
			body:           buf,
		}
		c.Writer = writer

		// process next
		c.Next()

		// ===== RESPONSE LOG =====
		duration := time.Since(startTime)
		var resBody string
		if o.skipDefaultContentTypeCheck || !utils.SkipContentType(c.Writer.Header().Get(consts.ContentType)) {
			resBody = buf.String()
		}

		if o.useLog {
			resLog := &logx.ResponseLogger{
				State:    state,
				Status:   c.Writer.Status(),
				Duration: duration,
				Body:     resBody,
			}
			if !o.skipHeader {
				resLog.Header = c.Writer.Header()
			}
			o.logger.LogResponse(resLog)
		} else {
			var sb strings.Builder
			sb.WriteString("\n========== RESPONSE INFO ==========\n")
			sb.WriteString(fmt.Sprintf("state: %s\n", state))
			sb.WriteString(fmt.Sprintf("status: %d\n", c.Writer.Status()))
			sb.WriteString(fmt.Sprintf("duration: %s\n", duration))
			if !o.skipHeader {
				sb.WriteString(fmt.Sprintf("header: %v\n", c.Writer.Header()))
			}
			if resBody != "" {
				sb.WriteString(fmt.Sprintf("body: %s\n", resBody))
			}
			sb.WriteString("==================================\n")
			log.Println(sb.String())
		}
	}
}
