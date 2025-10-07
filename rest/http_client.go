package rest

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/logger"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/datetime"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpClient[T any] struct {
	*RestClient

	// url is the target API endpoint (e.g., "/users/:id").
	url string

	// query contains query parameters to be appended to the URL (?key=value).
	query map[string]string

	// params contains path parameters to replace placeholders in the URL (e.g., ":id").
	params map[string]string

	// bodyForm contains form data (application/x-www-form-urlencoded).
	// Used only if Body is nil and the request requires form encoding.
	bodyForm map[string]string

	// headers allows you to set custom headers (e.g., Authorization, Content-DBType).
	headers map[string]string

	// body is the raw request body (typically a struct to be JSON-encoded).
	// This is ignored if BodyForm is set.
	body any
}

func NewRequest[T any](restClient *RestClient) HttpExec[T] {
	if restClient == nil {
		restClient = New(new(HttpConfig))
	}

	return &HttpClient[T]{
		RestClient: restClient,
	}
}

func (h *HttpClient[T]) URL(url string) HttpExec[T] {
	h.url = url
	return h
}

func (h *HttpClient[T]) Query(query map[string]string) HttpExec[T] {
	h.query = query
	return h
}

func (h *HttpClient[T]) Params(params map[string]string) HttpExec[T] {
	h.params = params
	return h
}

func (h *HttpClient[T]) Headers(headers map[string]string) HttpExec[T] {
	h.headers = headers
	return h
}

func (h *HttpClient[T]) Body(body any) HttpExec[T] {
	h.body = body
	return h
}

func (h *HttpClient[T]) BodyForm(bodyForm map[string]string) HttpExec[T] {
	h.bodyForm = bodyForm
	return h
}

func (h *HttpClient[T]) GET(c context.Context) (*T, error) {
	return h.restTemplate(c, http.MethodGet)
}

func (h *HttpClient[T]) POST(c context.Context) (*T, error) {
	return h.restTemplate(c, http.MethodPost)
}

func (h *HttpClient[T]) PostForm(c context.Context) (*T, error) {
	return h.restTemplate(c, http.MethodPost)
}

func (h *HttpClient[T]) PUT(c context.Context) (*T, error) {
	return h.restTemplate(c, http.MethodPut)
}

func (h *HttpClient[T]) PATCH(c context.Context) (*T, error) {
	return h.restTemplate(c, http.MethodPatch)
}

func (h *HttpClient[T]) DELETE(c context.Context) (*T, error) {
	return h.restTemplate(c, http.MethodDelete)
}

func (h *HttpClient[T]) restTemplate(c context.Context, method string) (*T, error) {
	var state = utils.GetState(c)

	// build URL
	h.buildURL()

	// serialize body
	// If form-data, encodes BodyForm as URL-encoded string.
	// If Body is []byte, use it directly and log as "[binary body]".
	// Otherwise, marshal Body to JSON and convert to string for logging.
	var (
		isFormData = !validate.IsNilOrEmpty(h.bodyForm)
		reqBody    []byte // send request
		bodyStr    string // send post and log
	)
	if isFormData {
		formValues := url.Values{}
		for k, v := range h.bodyForm {
			formValues.Add(k, v)
		}
		bodyStr = formValues.Encode()
	} else if !validate.IsNilOrEmpty(h.body) {
		switch b := h.body.(type) {
		case []byte:
			reqBody = b
			bodyStr = "[binary body]"
		default:
			reqBody = jsonx.ToJSONBytes(h.Body)
			bodyStr = string(reqBody)
		}
	}

	// build headers
	if validate.IsNilOrEmpty(h.headers) {
		h.headers = make(map[string]string)
	}
	if isFormData {
		h.setContentType(consts.ApplicationFormData)
	} else {
		h.setContentType(consts.ApplicationJSON)
	}

	// log request
	var (
		isLog     = h.Exec != nil
		startTime = time.Now()
	)
	if isLog {
		h.requestInfoLogger(state, method, bodyStr, startTime)
	} else {
		h.requestInfoConsole(state, method, bodyStr, startTime)
	}

	ctx, cancel := utils.NewCtxTimeout(c, h.TimeoutSec)
	defer cancel()

	// create request
	var (
		request *http.Request
		err     error
	)
	if isFormData {
		request, err = http.NewRequestWithContext(ctx, method, h.url, bytes.NewBufferString(bodyStr))
	} else if validate.IsNilOrEmpty(reqBody) {
		request, err = http.NewRequestWithContext(ctx, method, h.url, nil)
	} else {
		request, err = http.NewRequestWithContext(ctx, method, h.url, bytes.NewBuffer(reqBody))
	}
	if err != nil {
		return nil, err
	}

	// set headers
	h.setHeaders(request)

	// Execute the HTTP request
	return h.execute(request, startTime, state)
}

func (h *HttpClient[T]) requestInfoLogger(state, method, bodyStr string, startTime time.Time) {
	reqLogger := &logger.RequestLogger{
		State:       state,
		URL:         h.url,
		Method:      method,
		RequestTime: startTime,
	}
	if !validate.IsNilOrEmpty(h.query) {
		reqLogger.Query = str.ToString(h.query)
	}

	// skip log header
	if !h.SkipLogHeader {
		reqLogger.Header = h.headers
	}

	// log body
	if bodyStr != "" {
		// get content-type
		var contentType = h.headers[consts.ContentType]

		// write log body
		if !h.skipLogBody(contentType) {
			reqLogger.Body = bodyStr
		}
	} else {
		reqLogger.Body = "no request body"
	}

	h.LogExtRequest(reqLogger)
}

func (h *HttpClient[T]) requestInfoConsole(state, method, bodyStr string, startTime time.Time) {
	var sb strings.Builder

	sb.WriteString("========== REST REQUEST INFO ==========\n")
	sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
	sb.WriteString(fmt.Sprintf(consts.Url+": %s\n", h.url))
	sb.WriteString(fmt.Sprintf(consts.Query+": %v\n", h.query))
	sb.WriteString(fmt.Sprintf(consts.Method+": %s\n", method))
	sb.WriteString(fmt.Sprintf(consts.RequestTime+": %s\n", datetime.ToString(startTime, datetime.DateTimeOffset)))
	if !h.SkipLogHeader {
		sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", h.headers))
	}

	if bodyStr != "" {
		// get content-type
		var contentType = h.headers[consts.ContentType]

		// write log body
		if !h.skipLogBody(contentType) {
			sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", bodyStr))
		}
	} else {
		sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", "no request body"))
	}

	sb.WriteString("=================================\n")
	log.Println(sb.String())
}

func (h *HttpClient[T]) execute(request *http.Request, startTime time.Time, state string) (*T, error) {
	client := h.GetClient()
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// read body
	respBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var respBodyStr string
	hasBody := len(respBodyBytes) > 0
	if hasBody {
		respBodyStr = string(respBodyBytes)
	}

	// log response
	var (
		isLog      = h.Exec != nil
		respLogger logger.ResponseLogger
		sb         strings.Builder
	)

	// write log
	if isLog {
		respLogger = logger.ResponseLogger{
			State:       state,
			Status:      response.StatusCode,
			DurationSec: time.Since(startTime),
		}
		if !h.SkipLogHeader {
			respLogger.Header = response.Header
		}
	} else {
		sb.WriteString("========== REST RESPONSE INFO ==========\n")
		sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
		sb.WriteString(fmt.Sprintf(consts.Status+": %d\n", response.StatusCode))
		sb.WriteString(fmt.Sprintf(consts.Duration+": %s\n", time.Since(startTime)))
		if !h.SkipLogHeader {
			sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", response.Header))
		}
	}

	// defer log
	// check write log body in response
	defer func() {
		bodyContent := "hidden or empty body"
		if hasBody && !h.skipLogBody(response.Header.Get(consts.ContentType)) {
			bodyContent = respBodyStr
		}

		if isLog {
			respLogger.Body = bodyContent
			h.LogExtResponse(&respLogger)
		} else {
			sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", bodyContent))
			sb.WriteString("==================================\n")
			log.Println(sb.String())
		}
	}()

	// check error
	if response.StatusCode >= 400 {
		return nil, &HttpError{
			StatusCode: response.StatusCode,
			Body:       respBodyStr,
		}
	}

	if !hasBody {
		return nil, nil
	}

	var result T
	switch any(result).(type) {
	case []byte:
		result = any(respBodyBytes).(T)
	default:
		if err = jsonx.JSONBytesToStruct(respBodyBytes, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response to %T failed: %w", result, err)
		}
	}

	return &result, nil
}

func (h *HttpClient[T]) skipLogBody(contentType string) bool {
	// check skip by api
	if h.skipLogBodyAPIs(h.url, h.SkipLogAPIs) {
		return true
	}

	// check skip content-type
	if !validate.IsNilOrEmpty(h.SkipLogHeader) && contentType != "" {
		for _, c := range h.SkipContentType {
			if strings.HasPrefix(contentType, c) {
				return true
			}
		}
	} else if contentType != "" {
		return utils.SkipContentType(contentType)
	}

	return false
}

func (h *HttpClient[T]) skipLogBodyAPIs(u string, apis []string) bool {
	if len(apis) == 0 {
		return false
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	path := parsed.Path

	for _, pattern := range apis {
		if strings.HasSuffix(path, pattern) || strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (h *HttpClient[T]) buildURL() {
	for key, val := range h.query {
		h.url = strings.ReplaceAll(h.url, ":"+key, val)
	}

	if !validate.IsNilOrEmpty(h.params) {
		q := url.Values{}
		for k, v := range h.params {
			q.Add(k, v)
		}
		if strings.Contains(h.url, "?") {
			h.url += "&" + q.Encode()
		} else {
			h.url += "?" + q.Encode()
		}
	}
}

func (h *HttpClient[T]) setContentType(contentTypeDefault string) {
	if h.headers[consts.ContentType] == "" {
		h.headers[consts.ContentType] = contentTypeDefault
	}
}

func (h *HttpClient[T]) setHeaders(rq *http.Request) {
	for key, value := range h.headers {
		rq.Header.Set(key, value)
	}
}
