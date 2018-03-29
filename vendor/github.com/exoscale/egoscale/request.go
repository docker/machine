package egoscale

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Command represents a CloudStack request
type Command interface {
	// CloudStack API command name
	APIName() string
}

// SyncCommand represents a CloudStack synchronous request
type syncCommand interface {
	Command
	// Response interface to Unmarshal the JSON into
	response() interface{}
}

// asyncCommand represents a async CloudStack request
type asyncCommand interface {
	Command
	// Response interface to Unmarshal the JSON into
	asyncResponse() interface{}
}

// onBeforeHook represents an action to be done on the params before sending them
//
// This little took helps with issue of relying on JSON serialization logic only.
// `omitempty` may make sense in some cases but not all the time.
type onBeforeHook interface {
	onBeforeSend(params *url.Values) error
}

const (
	// Pending represents a job in progress
	Pending JobStatusType = iota
	// Success represents a successfully completed job
	Success
	// Failure represents a job that has failed to complete
	Failure
)

// JobStatusType represents the status of a Job
type JobStatusType int

const (
	// Unauthorized represents ... (TODO)
	Unauthorized ErrorCode = 401
	// MethodNotAllowed represents ... (TODO)
	MethodNotAllowed = 405
	// UnsupportedActionError represents ... (TODO)
	UnsupportedActionError = 422
	// APILimitExceeded represents ... (TODO)
	APILimitExceeded = 429
	// MalformedParameterError represents ... (TODO)
	MalformedParameterError = 430
	// ParamError represents ... (TODO)
	ParamError = 431

	// InternalError represents a server error
	InternalError = 530
	// AccountError represents ... (TODO)
	AccountError = 531
	// AccountResourceLimitError represents ... (TODO)
	AccountResourceLimitError = 532
	// InsufficientCapacityError represents ... (TODO)
	InsufficientCapacityError = 533
	// ResourceUnavailableError represents ... (TODO)
	ResourceUnavailableError = 534
	// ResourceAllocationError represents ... (TODO)
	ResourceAllocationError = 535
	// ResourceInUseError represents ... (TODO)
	ResourceInUseError = 536
	// NetworkRuleConflictError represents ... (TODO)
	NetworkRuleConflictError = 537
)

// ErrorCode represents the CloudStack ApiErrorCode enum
//
// See: https://github.com/apache/cloudstack/blob/master/api/src/org/apache/cloudstack/api/ApiErrorCode.java
type ErrorCode int

// JobResultResponse represents a generic response to a job task
type JobResultResponse struct {
	AccountID     string           `json:"accountid,omitempty"`
	Cmd           string           `json:"cmd"`
	Created       string           `json:"created"`
	JobID         string           `json:"jobid"`
	JobProcStatus int              `json:"jobprocstatus"`
	JobResult     *json.RawMessage `json:"jobresult"`
	JobStatus     JobStatusType    `json:"jobstatus"`
	JobResultType string           `json:"jobresulttype"`
	UserID        string           `json:"userid,omitempty"`
}

// ErrorResponse represents the standard error response from CloudStack
type ErrorResponse struct {
	ErrorCode   ErrorCode  `json:"errorcode"`
	CsErrorCode int        `json:"cserrorcode"`
	ErrorText   string     `json:"errortext"`
	UUIDList    []UUIDItem `json:"uuidList,omitempty"` // uuid*L*ist is not a typo
}

// UUIDItem represents an item of the UUIDList part of an ErrorResponse
type UUIDItem struct {
	Description      string `json:"description,omitempty"`
	SerialVersionUID int64  `json:"serialVersionUID,omitempty"`
	UUID             string `json:"uuid"`
}

// Error formats a CloudStack error into a standard error
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API error %d (internal code: %d): %s", e.ErrorCode, e.CsErrorCode, e.ErrorText)
}

// booleanAsyncResponse represents a boolean response (usually after a deletion)
type booleanAsyncResponse struct {
	Success     bool   `json:"success"`
	DisplayText string `json:"diplaytext,omitempty"`
}

// Error formats a CloudStack job response into a standard error
func (e *booleanAsyncResponse) Error() error {
	if e.Success {
		return nil
	}
	return fmt.Errorf("API error: %s", e.DisplayText)
}

// booleanAsyncResponse represents a boolean response for sync calls
type booleanSyncResponse struct {
	Success     string `json:"success"`
	DisplayText string `json:"displaytext,omitempty"`
}

func (e *booleanSyncResponse) Error() error {
	if e.Success == "true" {
		return nil
	}

	return fmt.Errorf("API error: %s", e.DisplayText)
}

func (exo *Client) parseResponse(resp *http.Response) (json.RawMessage, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	a, err := rawValues(b)

	if a == nil {
		b, err = rawValue(b)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 400 {
		errorResponse := new(ErrorResponse)
		if json.Unmarshal(b, errorResponse) == nil {
			return nil, errorResponse
		}
		return nil, fmt.Errorf("%d %s", resp.StatusCode, b)
	}

	return b, nil
}

// asyncRequest perform an asynchronous job with a context
func (exo *Client) asyncRequest(ctx context.Context, request asyncCommand) (interface{}, error) {
	body, err := exo.request(ctx, request.APIName(), request)
	if err != nil {
		return nil, err
	}

	jobResult := new(AsyncJobResult)
	if err := json.Unmarshal(body, jobResult); err != nil {
		r := new(ErrorResponse)
		if e := json.Unmarshal(body, r); e != nil {
			return nil, r
		}
		return nil, err
	}

	// Successful response
	if jobResult.JobID == "" || jobResult.JobStatus != Pending {
		response := request.asyncResponse()
		if err := json.Unmarshal(*(jobResult.JobResult), response); err != nil {
			return nil, err
		}
		return response, nil
	}

	for iteration := 0; ; iteration++ {
		time.Sleep(exo.RetryStrategy(int64(iteration)))

		req := &QueryAsyncJobResult{JobID: jobResult.JobID}
		resp, err := exo.Request(req)
		if err != nil {
			return nil, err
		}

		result := resp.(*QueryAsyncJobResultResponse)
		if result.JobStatus == Success {
			response := request.asyncResponse()
			if err := json.Unmarshal(*(result.JobResult), response); err != nil {
				return nil, err
			}
			return response, nil

		} else if result.JobStatus == Failure {
			r := new(ErrorResponse)
			if e := json.Unmarshal(*result.JobResult, r); e != nil {
				return nil, e
			}
			return nil, r
		}
	}
}

// syncRequest performs a sync request with a context
func (exo *Client) syncRequest(ctx context.Context, request syncCommand) (interface{}, error) {
	body, err := exo.request(ctx, request.APIName(), request)
	if err != nil {
		return nil, err
	}

	response := request.response()
	if err := json.Unmarshal(body, response); err != nil {
		errResponse := new(ErrorResponse)
		if json.Unmarshal(body, errResponse) == nil {
			return errResponse, nil
		}
		return nil, err
	}

	return response, nil
}

// BooleanRequest performs the given boolean command
func (exo *Client) BooleanRequest(req Command) error {
	resp, err := exo.Request(req)
	if err != nil {
		return err
	}

	// CloudStack returns a different type between sync and async success responses
	if b, ok := resp.(*booleanSyncResponse); ok {
		return b.Error()
	}

	if b, ok := resp.(*booleanAsyncResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("The command %s is not a proper boolean response. %#v", req.APIName(), resp))
}

// BooleanRequestWithContext performs the given boolean command
func (exo *Client) BooleanRequestWithContext(ctx context.Context, req Command) error {
	resp, err := exo.RequestWithContext(ctx, req)
	if err != nil {
		return err
	}

	// CloudStack returns a different type between sync and async success responses
	if b, ok := resp.(*booleanSyncResponse); ok {
		return b.Error()
	}
	if b, ok := resp.(*booleanAsyncResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("The command %s is not a proper boolean response. %#v", req.APIName(), resp))
}

// Request performs the given command
func (exo *Client) Request(request Command) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exo.Timeout)
	defer cancel()

	switch request.(type) {
	case syncCommand:
		return exo.syncRequest(ctx, request.(syncCommand))
	case asyncCommand:
		return exo.asyncRequest(ctx, request.(asyncCommand))
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", request.APIName()))
	}
}

// RequestWithContext preforms a request with a context
func (exo *Client) RequestWithContext(ctx context.Context, request Command) (interface{}, error) {
	switch request.(type) {
	case syncCommand:
		return exo.syncRequest(ctx, request.(syncCommand))
	case asyncCommand:
		return exo.asyncRequest(ctx, request.(asyncCommand))
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", request.APIName()))
	}
}

// request makes a Request while being close to the metal
func (exo *Client) request(ctx context.Context, command string, req interface{}) (json.RawMessage, error) {
	params := url.Values{}
	err := prepareValues("", &params, req)
	if err != nil {
		return nil, err
	}
	if hookReq, ok := req.(onBeforeHook); ok {
		hookReq.onBeforeSend(&params)
	}
	params.Set("apikey", exo.apiKey)
	params.Set("command", command)
	params.Set("response", "json")

	// This code is borrowed from net/url/url.go
	// The way it's encoded by net/url doesn't match
	// how CloudStack works.
	var buf bytes.Buffer
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		prefix := csEncode(k) + "="
		for _, v := range params[k] {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(csEncode(v))
		}
	}

	query := buf.String()

	mac := hmac.New(sha1.New, []byte(exo.apiSecret))
	mac.Write([]byte(strings.ToLower(query)))
	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	payload := fmt.Sprintf("%s&signature=%s", csQuotePlus(query), signature)

	request, err := http.NewRequest("POST", exo.endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(payload)))
	request = request.WithContext(ctx)

	resp, err := exo.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := exo.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return body, nil
}
