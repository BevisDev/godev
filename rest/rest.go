package rest

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BevisDev/godev/constants"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BevisDev/godev/logger"
	"github.com/BevisDev/godev/utils"
)

type RestRequest struct {
	URL      string
	Query    map[string]string
	Params   map[string]string
	BodyForm map[string]string
	Header   map[string]string
	Body     any
	Result   any
}

var defaultTimeoutSec = 30

type RestClient struct {
	Client     *http.Client
	TimeoutSec int
	logger     *logger.AppLogger
}

func NewRestClient(timeoutSec int) *RestClient {
	if timeoutSec <= 0 {
		timeoutSec = defaultTimeoutSec
	}
	return &RestClient{
		Client:     &http.Client{},
		TimeoutSec: timeoutSec,
		logger:     nil,
	}
}

func NewRestWithLogger(timeoutSec int, logger *logger.AppLogger) *RestClient {
	if timeoutSec <= 0 {
		timeoutSec = defaultTimeoutSec
	}
	return &RestClient{
		Client:     &http.Client{},
		TimeoutSec: timeoutSec,
		logger:     logger,
	}
}

func (r RestClient) Get(c context.Context, restReq *RestRequest) error {
	urlStr := r.buildQuery(restReq.URL, restReq.Query)

	// build params
	params := r.buildParams(restReq.Params)
	if params != "" {
		if strings.Contains(urlStr, "?") {
			urlStr += "&" + params
		} else {
			urlStr += "?" + params
		}
	}

	restReq.URL = urlStr
	return r.restTemplate(c, http.MethodGet, restReq)
}

func (r RestClient) Post(c context.Context, restReq *RestRequest) error {
	return r.restTemplate(c, http.MethodPost, restReq)
}

func (r RestClient) PostForm(c context.Context, restReq *RestRequest) error {
	return r.restTemplate(c, http.MethodPost, restReq)
}

func (r RestClient) Put(c context.Context, restReq *RestRequest) error {
	return r.restTemplate(c, http.MethodPut, restReq)
}

func (r RestClient) Patch(c context.Context, restReq *RestRequest) error {
	return r.restTemplate(c, http.MethodPatch, restReq)
}

func (r RestClient) Delete(c context.Context, restReq *RestRequest) error {
	return r.restTemplate(c, http.MethodDelete, restReq)
}

func (r RestClient) restTemplate(c context.Context, method string, restReq *RestRequest) error {
	var (
		state      = utils.GetState(c)
		reqBody    []byte
		bodyStr    string
		isLog      = !utils.IsNilOrEmpty(r.logger)
		isFormData = !utils.IsNilOrEmpty(restReq.BodyForm)
	)

	// serialize body
	if isFormData {
		bodyStr = r.buildParams(restReq.BodyForm)
	} else if !utils.IsNilOrEmpty(restReq.Body) {
		reqBody = utils.ToJSONBytes(restReq.Body)
		bodyStr = utils.ToString(reqBody)
	}

	// log request
	startTime := time.Now()
	if isLog {
		r.logger.LogExtRequest(&logger.RequestLogger{
			State:  state,
			URL:    restReq.URL,
			Method: method,
			Body:   bodyStr,
			Time:   startTime,
		})
	} else {
		var sb strings.Builder
		sb.WriteString("===== REQUEST EXTERNAL INFO =====\n")
		sb.WriteString(fmt.Sprintf("State: %s\n", state))
		sb.WriteString(fmt.Sprintf("URL: %s\n", restReq.URL))
		sb.WriteString(fmt.Sprintf("Method: %s\n", method))
		sb.WriteString(fmt.Sprintf("Time: %s\n", utils.TimeToString(startTime, utils.DATETIME_FULL)))
		if bodyStr != "" {
			sb.WriteString(fmt.Sprintf("Body: %s\n", bodyStr))
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
		request, err = http.NewRequestWithContext(ctx, method, restReq.URL, bytes.NewBufferString(bodyStr))
	} else if utils.IsNilOrEmpty(reqBody) {
		request, err = http.NewRequestWithContext(ctx, method, restReq.URL, nil)
	} else {
		request, err = http.NewRequestWithContext(ctx, method, restReq.URL, bytes.NewBuffer(reqBody))
	}
	if err != nil {
		return err
	}

	// build header
	if isFormData {
		r.buildHeaders(request, restReq.Header, constants.ApplicationFormData)
	} else {
		r.buildHeaders(request, restReq.Header, constants.ApplicationJSON)
	}

	// execute request
	return r.execute(request, restReq, startTime, state)
}

func (r RestClient) buildQuery(url string, queryParams map[string]string) string {
	for key, value := range queryParams {
		placeholder := ":" + key
		url = strings.ReplaceAll(url, placeholder, value)
	}
	return url
}

func (r RestClient) execute(request *http.Request, restReq *RestRequest,
	startTime time.Time, state string) error {
	var (
		respBodyBytes []byte
		isLog         = r.logger != nil
	)
	response, err := r.Client.Do(request)
	if err != nil {
		return err
	}

	// log response
	var (
		responseLogger *logger.ResponseLogger
		sb             strings.Builder
	)
	defer func() {
		response.Body.Close()
		if isLog {
			r.logger.LogExtResponse(responseLogger)
		} else {
			sb.WriteString("==================================\n")
			log.Println(sb.String())
		}
	}()

	if isLog {
		responseLogger = &logger.ResponseLogger{
			State:       state,
			Status:      response.StatusCode,
			DurationSec: time.Since(startTime),
		}
	} else {
		sb.WriteString("===== RESPONSE EXTERNAL INFO =====\n")
		sb.WriteString(fmt.Sprintf("State: %s\n", state))
		sb.WriteString(fmt.Sprintf("Status: %d\n", response.StatusCode))
		sb.WriteString(fmt.Sprintf("Duration: %s\n", time.Since(startTime)))
	}

	// read body
	respBodyBytes, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// check body
	hasBody := !utils.IsNilOrEmpty(respBodyBytes)
	if hasBody {
		respBodyStr := utils.ToString(respBodyBytes)
		if isLog {
			responseLogger.Body = respBodyStr
		} else {
			sb.WriteString(fmt.Sprintf("Body: %s\n", respBodyStr))
		}

		err = utils.JSONBytesToStruct(respBodyBytes, restReq.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r RestClient) buildParams(params map[string]string) string {
	if utils.IsNilOrEmpty(params) {
		return ""
	}
	urlParams := url.Values{}
	for k, v := range params {
		urlParams.Add(k, v)
	}
	return urlParams.Encode()
}

func (r RestClient) buildHeaders(rq *http.Request, headers map[string]string, contentType string) {
	if utils.IsNilOrEmpty(headers) || headers[constants.ContentType] == "" {
		rq.Header.Set(constants.ContentType, contentType)
	}
	for key, value := range headers {
		rq.Header.Add(key, value)
	}
}
