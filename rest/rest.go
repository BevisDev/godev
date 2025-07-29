package rest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/consts"
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

	"github.com/BevisDev/godev/logger"
	"github.com/BevisDev/godev/utils"
)

// defaultTimeoutSec defines the default timeout (in seconds) for rest client operations.
const defaultTimeoutSec = 60

// Request defines a standardized structure for making HTTP requests using RestClient.
//
// It supports query parameters, path parameters, custom headers, body payloads (as JSON),
// form-encoded bodies, and an optional Result destination for unmarshaling the response.
type Request struct {
	// URL is the target API endpoint (e.g., "/users/:id").
	URL string

	// Query contains query parameters to be appended to the URL (?key=value).
	Query map[string]string

	// Params contains path parameters to replace placeholders in the URL (e.g., ":id").
	Params map[string]string

	// BodyForm contains form data (application/x-www-form-urlencoded).
	// Used only if Body is nil and the request requires form encoding.
	BodyForm map[string]string

	// Header allows you to set custom headers (e.g., Authorization, Content-Type).
	Header map[string]string

	// Body is the raw request body (typically a struct to be JSON-encoded).
	// This is ignored if BodyForm is set.
	Body any

	// Result is a pointer to the variable where the response should be unmarshalled.
	// Example: &MyResponseStruct
	Result any
}

type RestConfig struct {
	TimeoutSec    int
	Logger        *logger.AppLogger
	SkipLogHeader bool
	SkipLogAPIs   []string
}

// RestClient wraps an HTTP client with a configurable timeout and optional logger.
//
// It is intended for making REST API calls with consistent timeout settings
// and optional logging support via AppLogger.
type RestClient struct {
	Client        *http.Client
	TimeoutSec    int
	logger        *logger.AppLogger
	SkipLogHeader bool
	SkipLogAPIs   []string
}

// NewRestClient creates a new RestClient instance using the provided RestConfig.
//
// It initializes the internal HTTP client and applies the specified timeout in seconds.
// If `TimeoutSec` is less than or equal to zero, a default timeout is used.
// Logging is disabled in this version unless added manually.
// The `HeaderLog` flag can be used to exclude headers from being logged (if logging is enabled later).
//
// Returns nil if the provided config is nil.
//
// Example:
//
//	cf := &RestConfig{
//	    TimeoutSec:      10,
//	    SkipLogHeader: true,
//	}
//
//	restClient := NewRestClient(cf)
func NewRestClient(cf *RestConfig) *RestClient {
	if cf == nil {
		cf = new(RestConfig)
	}

	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	return &RestClient{
		Client:        new(http.Client),
		TimeoutSec:    cf.TimeoutSec,
		logger:        cf.Logger,
		SkipLogHeader: cf.SkipLogHeader,
		SkipLogAPIs:   cf.SkipLogAPIs,
	}
}

func (r *RestClient) Get(c context.Context, req *Request) error {
	return r.restTemplate(c, http.MethodGet, req)
}

func (r *RestClient) Post(c context.Context, req *Request) error {
	return r.restTemplate(c, http.MethodPost, req)
}

func (r *RestClient) PostForm(c context.Context, req *Request) error {
	return r.restTemplate(c, http.MethodPost, req)
}

func (r *RestClient) Put(c context.Context, req *Request) error {
	return r.restTemplate(c, http.MethodPut, req)
}

func (r *RestClient) Patch(c context.Context, req *Request) error {
	return r.restTemplate(c, http.MethodPatch, req)
}

func (r *RestClient) Delete(c context.Context, req *Request) error {
	return r.restTemplate(c, http.MethodDelete, req)
}

// restTemplate builds and executes an HTTP request with optional logging and timeout.
//
// It constructs the URL, serializes the body (JSON or form data),
// logs the request if logging is enabled, sets headers, and sends the HTTP request.
//
// Params:
//   - c: Context for cancellation and timeout.
//   - method: HTTP method (e.g., "GET", "POST").
//   - req: Request configuration (URL, headers, query, body).
//
// Returns:
//   - error if request creation or execution fails.
func (r *RestClient) restTemplate(c context.Context, method string, req *Request) error {
	if req.Result == nil {
		return errors.New("result is nil")
	}

	var (
		state      = utils.GetState(c)
		reqBody    []byte
		bodyStr    string
		isLog      = !validate.IsNilOrEmpty(r.logger)
		isFormData = !validate.IsNilOrEmpty(req.BodyForm)
	)
	// build URL
	urlStr := r.buildURL(req.URL, req.Query, req.Params)

	// serialize body
	// If form-data, encodes BodyForm as URL-encoded string.
	// If Body is []byte, use it directly and log as "[binary body]".
	// Otherwise, marshal Body to JSON and convert to string for logging.
	if isFormData {
		formValues := url.Values{}
		for k, v := range req.BodyForm {
			formValues.Add(k, v)
		}
		bodyStr = formValues.Encode()
	} else if !validate.IsNilOrEmpty(req.Body) {
		switch b := req.Body.(type) {
		case []byte:
			reqBody = b
			bodyStr = "[binary body]"
		default:
			reqBody = jsonx.ToJSONBytes(req.Body)
			bodyStr = string(reqBody)
		}
	}

	// log request
	startTime := time.Now()
	if isLog {
		// Use structured logging
		reqLogger := &logger.RequestLogger{
			State:       state,
			URL:         req.URL,
			Method:      method,
			RequestTime: startTime,
		}
		if !validate.IsNilOrEmpty(req.Query) {
			reqLogger.Query = str.ToString(req.Query)
		}
		if bodyStr != "" && !SkipAPI(urlStr, r.SkipLogAPIs) {
			reqLogger.Body = bodyStr
		}
		if !r.SkipLogHeader {
			reqLogger.Header = req.Header
		}
		r.logger.LogExtRequest(reqLogger)
	} else {
		// Use simple stdout log
		var sb strings.Builder
		sb.WriteString("========== REST REQUEST INFO ==========\n")
		sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
		sb.WriteString(fmt.Sprintf(consts.Url+": %s\n", urlStr))
		sb.WriteString(fmt.Sprintf(consts.Query+": %v\n", req.Query))
		sb.WriteString(fmt.Sprintf(consts.Method+": %s\n", method))
		sb.WriteString(fmt.Sprintf(consts.RequestTime+": %s\n", datetime.ToString(startTime, datetime.DateTimeOffset)))
		if !r.SkipLogHeader {
			sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", req.Header))
		}
		if !SkipAPI(urlStr, r.SkipLogAPIs) && bodyStr != "" {
			sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", bodyStr))
		}
		sb.WriteString("=================================\n")
		log.Println(sb.String())
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	// create request
	var (
		request *http.Request
		err     error
	)
	if isFormData {
		request, err = http.NewRequestWithContext(ctx, method, urlStr, bytes.NewBufferString(bodyStr))
	} else if validate.IsNilOrEmpty(reqBody) {
		request, err = http.NewRequestWithContext(ctx, method, urlStr, nil)
	} else {
		request, err = http.NewRequestWithContext(ctx, method, urlStr, bytes.NewBuffer(reqBody))
	}
	if err != nil {
		return err
	}

	// Set headers based on content type
	if isFormData {
		r.buildHeaders(request, req.Header, consts.ApplicationFormData)
	} else {
		r.buildHeaders(request, req.Header, consts.ApplicationJSON)
	}

	// Execute the HTTP request
	return r.execute(request, req, startTime, state)
}

// execute sends the prepared HTTP request, reads the response,
// logs metadata and body (if enabled), and deserializes the response body.
//
// Behavior:
//   - If req.Result is *[]byte, assigns raw body bytes.
//   - Otherwise, deserializes JSON into req.Result.
//   - Logs headers and body based on configuration.
//   - Returns HttpError if the status code >= 400.
//
// Parameters:
//   - request: The prepared *http.Request.
//   - req: The original Request configuration.
//   - startTime: The request start time (for logging).
//   - state: A correlation ID or request state.
//
// Returns:
//   - error if the request fails, the response is an error, or decoding fails.
func (r *RestClient) execute(request *http.Request, req *Request, startTime time.Time, state string) error {
	response, err := r.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// read body
	var (
		isLog         = r.logger != nil
		respBodyBytes []byte
		respBodyStr   string
	)
	respBodyBytes, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	hasBody := !validate.IsNilOrEmpty(respBodyBytes)
	if hasBody {
		respBodyStr = string(respBodyBytes)
	}

	// log response
	var (
		respLogger logger.ResponseLogger
		sb         strings.Builder
	)
	if isLog {
		respLogger = logger.ResponseLogger{
			State:       state,
			Status:      response.StatusCode,
			DurationSec: time.Since(startTime),
		}
		if !r.SkipLogHeader {
			respLogger.Header = response.Header
		}
	} else {
		sb.WriteString("========== REST RESPONSE INFO ==========\n")
		sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
		sb.WriteString(fmt.Sprintf(consts.Status+": %d\n", response.StatusCode))
		sb.WriteString(fmt.Sprintf(consts.Duration+": %s\n", time.Since(startTime)))
		if !r.SkipLogHeader {
			sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", response.Header))
		}
	}
	defer func() {
		if isLog {
			r.logger.LogExtResponse(&respLogger)
		} else {
			sb.WriteString("==================================\n")
			log.Println(sb.String())
		}
	}()

	// don't body in response
	if !hasBody {
		if isLog {
			respLogger.Body = "no body response"
		} else {
			sb.WriteString(fmt.Sprintf(consts.Body + ": no body response \n"))
		}
		return nil
	}

	// Return raw bytes if expected
	if b, ok := req.Result.(*[]byte); ok {
		if response.StatusCode >= 400 {
			return &HttpError{
				StatusCode: response.StatusCode,
				Body:       respBodyStr,
			}
		}
		*b = respBodyBytes
		return nil
	}

	// log body
	if !utils.SkipContentType(response.Header.Get(consts.ContentType)) &&
		!SkipAPI(request.URL.Path, r.SkipLogAPIs) {
		if isLog {
			respLogger.Body = respBodyStr
		} else {
			sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", respBodyStr))
		}
	}

	if response.StatusCode >= 400 {
		return &HttpError{
			StatusCode: response.StatusCode,
			Body:       respBodyStr,
		}
	}

	return jsonx.JSONBytesToStruct(respBodyBytes, req.Result)
}

// SkipAPI checks whether the URL path matches any of the ignored API patterns.
//
// Returns true if the URL path ends with or contains any string in apis.
//
// Parameters:
//   - u: The URL to check.
//   - apis: A list of substrings or suffixes to skip.
//
// Returns:
//   - true if the URL matches any pattern, false otherwise.
func SkipAPI(u string, apis []string) bool {
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

// buildURL replaces path parameters and appends query parameters to a base URL.
//
// Parameters:
//   - urlStr: Base URL string (e.g., "/api/:id").
//   - query: Map of path placeholders to replace.
//   - params: Map of query parameters to append.
//
// Returns:
//   - The final URL string.
func (r *RestClient) buildURL(urlStr string, query map[string]string, params map[string]string) string {
	for key, val := range query {
		urlStr = strings.ReplaceAll(urlStr, ":"+key, val)
	}

	if !validate.IsNilOrEmpty(params) {
		q := url.Values{}
		for k, v := range params {
			q.Add(k, v)
		}
		if strings.Contains(urlStr, "?") {
			urlStr += "&" + q.Encode()
		} else {
			urlStr += "?" + q.Encode()
		}
	}
	return urlStr
}

// buildHeaders sets HTTP headers on the request,
// defaulting to the given Content-Type if none is provided.
//
// Parameters:
//   - rq: The HTTP request to set headers on.
//   - headers: Map of headers to add.
//   - contentType: Default Content-Type to set if none is provided.
func (r *RestClient) buildHeaders(rq *http.Request, headers map[string]string, contentType string) {
	if validate.IsNilOrEmpty(headers) || headers[consts.ContentType] == "" {
		rq.Header.Set(consts.ContentType, contentType)
	}
	for key, value := range headers {
		rq.Header.Set(key, value)
	}
}
