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

type request[T any] struct {
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

	// headers allows you to set custom headers (e.g., Authorization, Content-Type).
	headers map[string]string

	// body is the raw request body (typically a struct to be JSON-encoded).
	// This is ignored if BodyForm is set.
	body any

	// method execute request
	method string

	// rid: ID request
	rid string

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

func NewRequest[T any](client *Client) HttpClient[T] {
	if client == nil {
		client = New()
	}

	return &request[T]{
		Client: client,
	}
}

func (r *request[T]) URL(url string) HttpClient[T] {
	r.url = url
	return r
}

func (r *request[T]) QueryParams(query map[string]string) HttpClient[T] {
	r.queryParams = query
	return r
}

func (r *request[T]) PathParams(params map[string]string) HttpClient[T] {
	r.pathParams = params
	return r
}

func (r *request[T]) Headers(headers map[string]string) HttpClient[T] {
	r.headers = headers
	return r
}

func (r *request[T]) Body(body any) HttpClient[T] {
	r.body = body
	return r
}

func (r *request[T]) BodyForm(bodyForm map[string]string) HttpClient[T] {
	r.bodyForm = bodyForm
	return r
}

func (r *request[T]) GET(c context.Context) (Response[T], error) {
	r.method = http.MethodGet
	return r.restTemplate(c)
}

func (r *request[T]) POST(c context.Context) (Response[T], error) {
	r.method = http.MethodPost
	return r.restTemplate(c)
}

func (r *request[T]) PostForm(c context.Context) (Response[T], error) {
	r.method = http.MethodPost
	return r.restTemplate(c)
}

func (r *request[T]) PUT(c context.Context) (Response[T], error) {
	r.method = http.MethodPut
	return r.restTemplate(c)
}

func (r *request[T]) PATCH(c context.Context) (Response[T], error) {
	r.method = http.MethodPatch
	return r.restTemplate(c)
}

func (r *request[T]) DELETE(c context.Context) (Response[T], error) {
	r.method = http.MethodDelete
	return r.restTemplate(c)
}

func (r *request[T]) restTemplate(c context.Context) (Response[T], error) {
	// set metadata
	r.rid = utils.GetRID(c)
	r.startTime = time.Now()

	// determine request shape and prepare URL/body/headers
	isFormData := !validate.IsNilOrEmpty(r.bodyForm)
	r.setContentType(isFormData)
	r.buildURL()

	// serialise body for transport and logging
	raw, body, err := r.serializeBody(isFormData)
	if err != nil {
		return Response[T]{}, err
	}

	// log request
	r.logRequest(body)

	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	// create request
	request, err := r.httpRequest(ctx, isFormData, raw, body)
	if err != nil {
		return Response[T]{}, err
	}

	// set headers
	r.setHeaders(request)

	// Execute the HTTP request
	return r.execute(request)
}

// serializeBody
// If form-data, encodes BodyForm as URL-encoded string.
// If Body is []byte, use it directly and log as "[binary body]".
// Otherwise, marshal Body to JSON and convert to string for logging.
func (r *request[T]) serializeBody(isFormData bool) ([]byte, string, error) {
	// CASE: form-data
	if isFormData {
		formValues := url.Values{}
		for k, v := range r.bodyForm {
			formValues.Add(k, v)
		}
		return nil, formValues.Encode(), nil
	}

	// CASE: []byte and JSON
	if !validate.IsNilOrEmpty(r.body) {
		switch b := r.body.(type) {
		case []byte:
			return b, "", nil
		default:
			raw, err := jsonx.ToJSONBytes(r.body)
			if err != nil {
				return nil, "", err
			}
			return raw, string(raw), nil
		}
	}

	return nil, "", nil
}

// httpRequest constructs the underlying *http.Request based on
// the previously prepared URL, headers and body serialisation.
func (r *request[T]) httpRequest(
	ctx context.Context,
	isFormData bool,
	raw []byte,
	body string,
) (*http.Request, error) {
	switch {
	case isFormData:
		return http.NewRequestWithContext(ctx, r.method, r.url, bytes.NewBufferString(body))
	case validate.IsNilOrEmpty(raw):
		return http.NewRequestWithContext(ctx, r.method, r.url, nil)
	default:
		return http.NewRequestWithContext(ctx, r.method, r.url, bytes.NewBuffer(raw))
	}
}

func (r *request[T]) logRequest(body string) {
	if r.useLog {
		log := &logx.RequestLogger{
			RID:    r.rid,
			URL:    r.url,
			Method: r.method,
			Time:   r.startTime,
		}
		if !validate.IsNilOrEmpty(r.queryParams) {
			log.Query = str.ToString(r.queryParams)
		}
		if !r.skipHeader {
			log.Header = r.headers
		}
		if r.logBody(r.headers[consts.ContentType]) {
			log.Body = body
		}
		r.logger.LogExtRequest(log)
		return
	}

	var sb strings.Builder
	sb.WriteString("\n========== REQUEST INFO ==========\n")
	fmt.Fprintf(&sb, "%s: %s\n", consts.RID, r.rid)
	fmt.Fprintf(&sb, "%s: %s\n", consts.Url, r.url)
	fmt.Fprintf(&sb, "%s: %s\n", consts.Method, r.method)
	fmt.Fprintf(&sb, "%s: %s\n", consts.Time,
		datetime.ToString(r.startTime, datetime.DateTimeLayoutMilli))
	if !validate.IsNilOrEmpty(r.queryParams) {
		fmt.Fprintf(&sb, "%s: %v\n", consts.Query, r.queryParams)
	}
	if !r.skipHeader {
		fmt.Fprintf(&sb, "%s: %s\n", consts.Header, r.headers)
	}
	if r.logBody(r.headers[consts.ContentType]) {
		fmt.Fprintf(&sb, "%s: %s\n", consts.Body, body)
	}
	sb.WriteString("==================================\n")
	log.Println(sb.String())
}

func (r *request[T]) execute(request *http.Request) (Response[T], error) {
	client := r.GetClient()
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
		Duration:   time.Since(r.startTime),
	}

	// log response
	r.logResponse(response, resp.Body)

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

	result, err := r.getData(raw)
	if err != nil {
		return resp, err
	}
	resp.Data = result

	return resp, nil
}

// getData converts the raw HTTP response bytes into the generic
// type T. It supports []byte, json.RawMessage, string and arbitrary structs.
func (r *request[T]) getData(raw []byte) (T, error) {
	var result T

	switch any(result).(type) {
	case []byte, json.RawMessage:
		return any(raw).(T), nil
	case string:
		return any(string(raw)).(T), nil
	default:
		result, err := jsonx.FromJSONBytes[T](raw)
		if err != nil {
			return result, fmt.Errorf("unmarshal response to %T failed: %w", result, err)
		}
		return result, nil
	}
}

func (r *request[T]) logResponse(response *http.Response, body string) {
	if r.useLog {
		logger := &logx.ResponseLogger{
			RID:      r.rid,
			Status:   response.StatusCode,
			Duration: time.Since(r.startTime),
		}
		if !r.skipHeader {
			logger.Header = response.Header
		}
		if r.logBody(response.Header.Get(consts.ContentType)) {
			logger.Body = body
		}
		r.logger.LogExtResponse(logger)
	} else {
		var sb strings.Builder
		sb.WriteString("\n========== RESPONSE INFO ==========\n")
		fmt.Fprintf(&sb, "%s: %s\n", consts.RID, r.rid)
		fmt.Fprintf(&sb, "%s: %d\n", consts.Status, response.StatusCode)
		fmt.Fprintf(&sb, "%s: %s\n", consts.Duration, time.Since(r.startTime))
		if !r.skipHeader {
			fmt.Fprintf(&sb, "%s: %s\n", consts.Header, response.Header)
		}
		if r.logBody(response.Header.Get(consts.ContentType)) {
			fmt.Fprintf(&sb, "%s: %s\n", consts.Body, body)
		}
		sb.WriteString("==================================\n")
		log.Println(sb.String())
	}
}

func (r *request[T]) logBody(contentType string) bool {
	// ---- skip by content-type ----
	for c, _ := range r.skipBodyByContentTypes {
		if strings.HasPrefix(contentType, c) {
			return false
		}
	}

	// ---- check default content-type ----
	if !r.skipDefaultContentTypeCheck && utils.SkipContentType(contentType) {
		return false
	}

	// ---- skip by path ----
	if len(r.skipBodyByPaths) > 0 {
		parsed, err := url.Parse(r.url)
		if err != nil {
			return true
		}
		path := parsed.Path

		for p, _ := range r.skipBodyByPaths {
			if p == path {
				return false
			}

			// prefix wildcard match: /internal/*
			if strings.HasSuffix(p, "*") &&
				strings.HasPrefix(path, strings.TrimSuffix(p, "*")) {
				return false
			}

			// suffix wildcard match: */token
			if strings.HasPrefix(p, "*") &&
				strings.HasSuffix(path, strings.TrimPrefix(p, "*")) {
				return false
			}
		}
	}

	return true
}

func (r *request[T]) buildURL() {
	for key, val := range r.pathParams {
		if strings.HasPrefix(key, ":") {
			r.url = strings.ReplaceAll(r.url, key, val)
		} else {
			r.url = strings.ReplaceAll(r.url, ":"+key, val)
		}
	}

	if !validate.IsNilOrEmpty(r.queryParams) {
		q := url.Values{}
		for k, v := range r.queryParams {
			q.Add(k, v)
		}

		if strings.Contains(r.url, "?") {
			r.url += "&" + q.Encode()
		} else {
			r.url += "?" + q.Encode()
		}
	}
}

func (r *request[T]) setContentType(isFormData bool) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}

	if r.headers[consts.ContentType] == "" {
		if isFormData {
			r.headers[consts.ContentType] = consts.ApplicationFormData
		} else {
			r.headers[consts.ContentType] = consts.ApplicationJSON
		}
	}
}

func (r *request[T]) setHeaders(rq *http.Request) {
	for key, value := range r.headers {
		rq.Header.Set(key, value)
	}
}
