package rest

import (
	"bytes"
	"context"
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

	// Result is a pointer to the variable where the response should be unmarshaled.
	// Example: &MyResponseStruct
	Result any
}

const defaultTimeoutSec = 30

type RestConfig struct {
	TimeoutSec      int
	Logger          *logger.AppLogger
	IgnoreLogHeader bool
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
//	    IgnoreLogHeader: true,
//	}
//
//	restClient := NewRestClient(cf)
func NewRestClient(cf *RestConfig) *RestClient {
	if cf == nil {
		return nil
	}

	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	return &RestClient{
		Client:          new(http.Client),
		TimeoutSec:      cf.TimeoutSec,
		logger:          cf.Logger,
		IgnoreLogHeader: cf.IgnoreLogHeader,
	}
}

// RestClient wraps an HTTP client with a configurable timeout and optional logger.
//
// It is intended for making REST API calls with consistent timeout settings
// and optional logging support via AppLogger.
type RestClient struct {
	Client          *http.Client
	TimeoutSec      int
	logger          *logger.AppLogger
	IgnoreLogHeader bool
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

func (r *RestClient) restTemplate(c context.Context, method string, req *Request) error {
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
	if isFormData {
		formValues := url.Values{}
		for k, v := range req.BodyForm {
			formValues.Add(k, v)
		}
		bodyStr = formValues.Encode()
	} else if !validate.IsNilOrEmpty(req.Body) {
		reqBody = jsonx.ToJSONBytes(req.Body)
		bodyStr = str.ToString(reqBody)
	}

	// log request
	startTime := time.Now()
	if isLog {
		reqLogger := &logger.RequestLogger{
			State:  state,
			URL:    req.URL,
			Method: method,
			Body:   bodyStr,
			Time:   startTime,
		}
		if !r.IgnoreLogHeader {
			reqLogger.Header = req.Header
		}
		r.logger.LogExtRequest(reqLogger)
	} else {
		var sb strings.Builder
		sb.WriteString("========== REST REQUEST INFO ==========\n")
		sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
		sb.WriteString(fmt.Sprintf(consts.Url+": %s\n", req.URL))
		sb.WriteString(fmt.Sprintf(consts.Method+": %s\n", method))
		sb.WriteString(fmt.Sprintf(consts.Time+": %s\n", datetime.ToString(startTime, datetime.DateTimeOffset)))
		if !r.IgnoreLogHeader {
			sb.WriteString(fmt.Sprintf(consts.Header+": %s\n", req.Header))
		}
		if bodyStr != "" {
			sb.WriteString(fmt.Sprintf(consts.Body+": %s\n", bodyStr))
		}
		sb.WriteString("=================================\n")
		log.Println(sb.String())
	}

	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
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

	// build header
	if isFormData {
		r.buildHeaders(request, req.Header, consts.ApplicationFormData)
	} else {
		r.buildHeaders(request, req.Header, consts.ApplicationJSON)
	}

	// execute request
	return r.execute(request, req, startTime, state)
}

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
		if !r.IgnoreLogHeader {
			respLogger.Header = response.Header
		}
	} else {
		sb.WriteString("========== REST RESPONSE INFO ==========\n")
		sb.WriteString(fmt.Sprintf(consts.State+": %s\n", state))
		sb.WriteString(fmt.Sprintf(consts.Status+": %d\n", response.StatusCode))
		sb.WriteString(fmt.Sprintf(consts.Duration+": %s\n", time.Since(startTime)))
		if !r.IgnoreLogHeader {
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

	// check body
	hasBody := !validate.IsNilOrEmpty(respBodyBytes)
	if hasBody {
		respBodyStr = str.ToString(respBodyBytes)
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

	if req.Result == nil || !hasBody {
		return nil
	}

	return jsonx.JSONBytesToStruct(respBodyBytes, req.Result)
}

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

func (r *RestClient) buildHeaders(rq *http.Request, headers map[string]string, contentType string) {
	if validate.IsNilOrEmpty(headers) || headers[consts.ContentType] == "" {
		rq.Header.Set(consts.ContentType, contentType)
	}
	for key, value := range headers {
		rq.Header.Add(key, value)
	}
}
