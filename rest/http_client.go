package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/datetime"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
)

type HttpClient[T any] struct {
	*Client

	// url is the target API endpoint (e.g., "/users/:id").
	url string

	// queryParams contains query parameters to be appended to the URL (?key=value).
	queryParams map[string]string

	// pathParams contains path parameters to replace placeholders in the URL (e.g., ":id").
	pathParams map[string]string

	// bodyForm contains form data (application/x-www-form-urlencoded).
	// Used only if Body is nil and the request requires form encoding.
	bodyForm map[string]string

	// headers allows you to set custom headers (e.g., Authorization, Content-DBType).
	headers map[string]string

	// body is the raw request body (typically a struct to be JSON-encoded).
	// This is ignored if BodyForm is set.
	body any

	// method execute request
	method string

	// state: ID request
	state string

	// startTime time begin request
	startTime time.Time
}

type Response[T any] struct {
	StatusCode int
	Header     http.Header
	Data       T
	Duration   time.Duration
	RawBody    []byte
	Body       string
	HasBody    bool
}

func NewRequest[T any](client *Client) HttpHandler[T] {
	if client == nil {
		client = NewClient(nil)
	}

	return &HttpClient[T]{
		Client: client,
	}
}

func (h *HttpClient[T]) URL(url string) HttpHandler[T] {
	h.url = url
	return h
}

func (h *HttpClient[T]) QueryParams(query map[string]string) HttpHandler[T] {
	h.queryParams = query
	return h
}

func (h *HttpClient[T]) PathParams(params map[string]string) HttpHandler[T] {
	h.pathParams = params
	return h
}

func (h *HttpClient[T]) Headers(headers map[string]string) HttpHandler[T] {
	h.headers = headers
	return h
}

func (h *HttpClient[T]) Body(body any) HttpHandler[T] {
	h.body = body
	return h
}

func (h *HttpClient[T]) BodyForm(bodyForm map[string]string) HttpHandler[T] {
	h.bodyForm = bodyForm
	return h
}

func (h *HttpClient[T]) GET(c context.Context) (Response[T], error) {
	h.method = http.MethodGet
	return h.restTemplate(c)
}

func (h *HttpClient[T]) POST(c context.Context) (Response[T], error) {
	h.method = http.MethodPost
	return h.restTemplate(c)
}

func (h *HttpClient[T]) PostForm(c context.Context) (Response[T], error) {
	h.method = http.MethodPost
	return h.restTemplate(c)
}

func (h *HttpClient[T]) PUT(c context.Context) (Response[T], error) {
	h.method = http.MethodPut
	return h.restTemplate(c)
}

func (h *HttpClient[T]) PATCH(c context.Context) (Response[T], error) {
	h.method = http.MethodPatch
	return h.restTemplate(c)
}

func (h *HttpClient[T]) DELETE(c context.Context) (Response[T], error) {
	h.method = http.MethodDelete
	return h.restTemplate(c)
}

func (h *HttpClient[T]) restTemplate(c context.Context) (Response[T], error) {
	// set metadata
	h.state = utils.GetState(c)
	h.startTime = time.Now()
	if validate.IsNilOrEmpty(h.headers) {
		h.headers = make(map[string]string)
	}

	// build URL
	h.buildURL()

	// flag check request form-data
	isFormData := !validate.IsNilOrEmpty(h.bodyForm)

	// raw to send request
	// body send form-data and log
	raw, body := h.serializeBody(isFormData)

	// set content-type
	h.setContentType(isFormData)

	// log request
	h.logRequest(body)

	ctx, cancel := utils.NewCtxTimeout(c, h.TimeoutSec)
	defer cancel()

	// create request
	var (
		request *http.Request
		err     error
	)
	if isFormData {
		request, err = http.NewRequestWithContext(ctx, h.method, h.url, bytes.NewBufferString(body))
	} else if validate.IsNilOrEmpty(raw) {
		request, err = http.NewRequestWithContext(ctx, h.method, h.url, nil)
	} else {
		request, err = http.NewRequestWithContext(ctx, h.method, h.url, bytes.NewBuffer(raw))
	}
	if err != nil {
		return Response[T]{}, err
	}

	// set headers
	h.setHeaders(request)

	// Execute the HTTP request
	return h.execute(request)
}

// serializeBody
// If form-data, encodes BodyForm as URL-encoded string.
// If Body is []byte, use it directly and log as "[binary body]".
// Otherwise, marshal Body to JSON and convert to string for logging.
func (h *HttpClient[T]) serializeBody(isFormData bool) ([]byte, string) {
	// CASE: form-data
	if isFormData {
		formValues := url.Values{}
		for k, v := range h.bodyForm {
			formValues.Add(k, v)
		}
		return nil, formValues.Encode()
	}

	// CASE: []byte and JSON
	if !validate.IsNilOrEmpty(h.body) {
		switch b := h.body.(type) {
		case []byte:
			return b, ""
		default:
			raw := jsonx.ToJSONBytes(h.body)
			return raw, string(raw)
		}
	}

	return nil, ""
}

func (h *HttpClient[T]) logRequest(body string) {
	if h.hasLog {
		log := &logx.RequestLogger{
			State:       h.state,
			URL:         h.url,
			Method:      h.method,
			RequestTime: h.startTime,
		}
		if !validate.IsNilOrEmpty(h.queryParams) {
			log.Query = str.ToString(h.queryParams)
		}

		// skip log header
		if !h.SkipLogHeader {
			log.Header = h.headers
		}

		// log body
		if body != "" && h.logBody(h.headers[consts.ContentType]) {
			log.Body = body
		}

		h.Logger.LogExtRequest(log)
		return
	}

	var sb strings.Builder
	sb.WriteString("========== REQUEST INFO ==========\n")
	sb.WriteString(fmt.Sprintf(consts.State+": %s\n", h.state))
	sb.WriteString(fmt.Sprintf(consts.Url+": %s\n", h.url))
	sb.WriteString(fmt.Sprintf(consts.Method+": %s\n", h.method))
	sb.WriteString(fmt.Sprintf(consts.RequestTime+": %s\n", datetime.ToString(h.startTime, datetime.DateTimeOffset)))
	if !validate.IsNilOrEmpty(h.queryParams) {
		sb.WriteString(fmt.Sprintf(consts.Query+": %v\n", h.queryParams))
	}
	if !h.SkipLogHeader {
		sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", h.headers))
	}

	if body != "" && h.logBody(h.headers[consts.ContentType]) {
		sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", body))
	}

	sb.WriteString("==================================\n")
	log.Println(sb.String())
}

func (h *HttpClient[T]) execute(request *http.Request) (Response[T], error) {
	client := h.GetClient()
	response, err := client.Do(request)
	if err != nil {
		return Response[T]{}, err
	}
	defer response.Body.Close()

	// READ BODY
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return Response[T]{}, err
	}

	// BUILD RESPONSE
	var resp = Response[T]{
		StatusCode: response.StatusCode,
		Body:       string(raw),
		RawBody:    raw,
		Header:     response.Header,
		HasBody:    len(raw) > 0,
	}

	// log response
	h.logResponse(response, resp.HasBody, resp.Body)

	// check error
	if resp.StatusCode >= 400 {
		return resp, &HttpError{
			StatusCode: resp.StatusCode,
			Body:       resp.Body,
		}
	}

	if !resp.HasBody {
		return resp, nil
	}

	// GET DATA
	var result T
	switch any(result).(type) {
	case []byte, json.RawMessage:
		resp.Data = any(raw).(T)
	case string:
		resp.Data = any(resp.Body).(T)
	default:
		if err = jsonx.JSONBytesToStruct(raw, &result); err != nil {
			return resp, fmt.Errorf("unmarshal response to %T failed: %w", result, err)
		}
		resp.Data = result
	}

	return resp, nil
}

func (h *HttpClient[T]) logResponse(response *http.Response,
	hasBody bool, body string) {
	if h.hasLog {
		logger := &logx.ResponseLogger{
			State:    h.state,
			Status:   response.StatusCode,
			Duration: time.Since(h.startTime),
		}
		if !h.SkipLogHeader {
			logger.Header = response.Header
		}
		if hasBody && h.logBody(response.Header.Get(consts.ContentType)) {
			logger.Body = body
		}
		h.Logger.LogExtResponse(logger)
	} else {
		var sb strings.Builder
		sb.WriteString("========== RESPONSE INFO ==========\n")
		sb.WriteString(fmt.Sprintf(consts.State+": %s\n", h.state))
		sb.WriteString(fmt.Sprintf(consts.Status+": %d\n", response.StatusCode))
		sb.WriteString(fmt.Sprintf(consts.Duration+": %s\n", time.Since(h.startTime)))
		if !h.SkipLogHeader {
			sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", response.Header))
		}
		if hasBody && h.logBody(response.Header.Get(consts.ContentType)) {
			sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", body))
		}
		sb.WriteString("==================================\n")
		log.Println(sb.String())
	}
}

func (h *HttpClient[T]) logBody(contentType string) bool {
	// skip by apis
	if h.skipBodyByAPIs(h.url, h.SkipLogAPIs) {
		return false
	}

	// skip by content-type
	if !validate.IsNilOrEmpty(h.SkipContentType) {
		for _, c := range h.SkipContentType {
			if strings.HasPrefix(contentType, c) {
				return false
			}
		}
	}

	return !utils.SkipContentType(contentType)
}

func (h *HttpClient[T]) skipBodyByAPIs(u string, apis []string) bool {
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
	for key, val := range h.pathParams {
		if strings.HasPrefix(key, ":") {
			h.url = strings.ReplaceAll(h.url, key, val)
		} else {
			h.url = strings.ReplaceAll(h.url, ":"+key, val)
		}
	}

	if !validate.IsNilOrEmpty(h.queryParams) {
		q := url.Values{}
		for k, v := range h.queryParams {
			q.Add(k, v)
		}

		if strings.Contains(h.url, "?") {
			h.url += "&" + q.Encode()
		} else {
			h.url += "?" + q.Encode()
		}
	}
}

func (h *HttpClient[T]) setContentType(isFormData bool) {
	if h.headers[consts.ContentType] == "" {
		if isFormData {
			h.headers[consts.ContentType] = consts.ApplicationFormData
		} else {
			h.headers[consts.ContentType] = consts.ApplicationJSON
		}
	}
}

func (h *HttpClient[T]) setHeaders(rq *http.Request) {
	for key, value := range h.headers {
		rq.Header.Set(key, value)
	}
}
