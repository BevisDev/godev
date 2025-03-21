package rest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/BevisDev/godev/helper"
	"github.com/BevisDev/godev/logger"
)

type RestRequest struct {
	State    string
	URL      string
	Params   map[string]any
	Header   map[string]string
	Body     any
	BodyForm url.Values
	Result   any
}

type RestClient struct {
	client     *http.Client
	timeoutSec int
	logger     *logger.AppLogger
}

func NewRestClient(timeoutSec int, logger *logger.AppLogger) *RestClient {
	return &RestClient{
		client:     &http.Client{},
		timeoutSec: timeoutSec,
		logger:     logger,
	}
}

func addHeaders(rq *http.Request, headers map[string]string) {
	if helper.IsNilOrEmpty(headers) || headers[helper.ContentType] == "" {
		rq.Header.Set(helper.ContentType, helper.ApplicationJSON)
	}
	for key, value := range headers {
		rq.Header.Add(key, value)
	}
}

func (r *RestClient) execute(state string, request *http.Request, restReq *RestRequest, startTime time.Time) error {
	var (
		respBodyBytes []byte
	)
	response, err := r.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// response logger
	responseLogger := &logger.ResponseLogger{
		State:       state,
		Status:      response.StatusCode,
		DurationSec: time.Since(startTime),
	}
	defer r.logger.LogExtResponse(responseLogger)

	// read body
	respBodyBytes, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// check body
	hasBody := !helper.IsNilOrEmpty(respBodyBytes)
	if hasBody {
		responseLogger.Body = helper.JSONToStr(respBodyBytes)

		if err = helper.ToStruct(respBodyBytes, restReq.Result); err != nil {
			return err
		}
	}

	return nil
}

func (r *RestClient) Post(c context.Context, restReq *RestRequest) error {
	var (
		state        = helper.GetState(c)
		reqBodyBytes []byte
	)

	// serialize body
	if !helper.IsNilOrEmpty(restReq.Body) {
		reqBodyBytes = helper.ToJSON(restReq.Body)
	}

	startTime := time.Now()
	// log request
	r.logger.LogExtRequest(&logger.RequestLogger{
		URL:    restReq.URL,
		Method: http.MethodPost,
		Body:   helper.ToJSONStr(restReq.Body),
		Time:   startTime,
	})

	ctx, cancel := helper.CreateCtxTimeout(c, r.timeoutSec)
	defer cancel()

	// created request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		restReq.URL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	// build header
	addHeaders(request, restReq.Header)

	// execute request
	return r.execute(state, request, restReq, startTime)
}

func (r *RestClient) PostForm(c context.Context, restReq *RestRequest) error {
	var (
		state   = helper.GetState(c)
		reqBody = restReq.BodyForm.Encode()
	)
	startTime := time.Now()
	// log request
	r.logger.LogExtRequest(&logger.RequestLogger{
		State:  state,
		URL:    restReq.URL,
		Method: http.MethodPost,
		Body:   reqBody,
		Time:   startTime,
	})
	ctx, cancel := helper.CreateCtxTimeout(c, r.timeoutSec)
	defer cancel()

	// created request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, restReq.URL,
		bytes.NewBufferString(reqBody))
	if err != nil {
		return err
	}

	// build header
	if helper.IsNilOrEmpty(restReq.Header) {
		restReq.Header = make(map[string]string)
		restReq.Header[helper.ContentType] = helper.ApplicationFormData
	} else if restReq.Header[helper.ContentType] == "" {
		restReq.Header[helper.ContentType] = helper.ApplicationFormData
	}
	addHeaders(request, restReq.Header)

	// execute request
	return r.execute(state, request, restReq, startTime)
}
