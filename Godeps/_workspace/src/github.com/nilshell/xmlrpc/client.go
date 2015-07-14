package xmlrpc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"reflect"
)

type Client struct {
	*rpc.Client
}

// clientCodec is rpc.ClientCodec interface implementation.
type clientCodec struct {
	// url presents url of xmlrpc service
	url string

	// httpClient works with HTTP protocol
	httpClient *http.Client

	// cookies stores cookies received on last request
	cookies []*http.Cookie

	// responses presents map of active requests. It is required to return request id, that
	// rpc.Client can mark them as done.
	responses map[uint64]*http.Response

	// responseBody holds response body of last request.
	responseBody []byte

	// ready presents channel, that is used to link request and it`s response.
	ready chan uint64
}

func (codec *clientCodec) WriteRequest(request *rpc.Request, params interface{}) (err error) {
	httpRequest, err := newRequest(codec.url, request.ServiceMethod, params)

	if codec.cookies != nil {
		for _, cookie := range codec.cookies {
			httpRequest.AddCookie(cookie)
		}
	}

	if err != nil {
		return err
	}

	var httpResponse *http.Response
	httpResponse, err = codec.httpClient.Do(httpRequest)

	if err != nil {
		return err
	}

	if codec.cookies == nil {
		codec.cookies = httpResponse.Cookies()
	}

	codec.responses[request.Seq] = httpResponse
	codec.ready <- request.Seq

	return nil
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) (err error) {
	seq := <-codec.ready
	httpResponse := codec.responses[seq]

	codec.responseBody, err = ioutil.ReadAll(httpResponse.Body)

	if err != nil {
		return err
	}

	httpResponse.Body.Close()

	if fault, _ := responseFailed(codec.responseBody); fault {
		response.Error = fmt.Sprintf("%v", parseFailedResponse(codec.responseBody))
	}

	response.Seq = seq
	delete(codec.responses, seq)

	return nil
}

func (codec *clientCodec) ReadResponseBody(x interface{}) (err error) {
	if x == nil {
		return nil
	}

	var result interface{}
	result, err = parseSuccessfulResponse(codec.responseBody)

	if err != nil {
		return err
	}

	v := reflect.ValueOf(x)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	v.Set(reflect.ValueOf(result))

	return nil
}

func (codec *clientCodec) Close() error {
	transport := codec.httpClient.Transport.(*http.Transport)
	transport.CloseIdleConnections()
	return nil
}

// NewClient returns instance of rpc.Client object, that is used to send request to xmlrpc service.
func NewClient(url string, transport *http.Transport) (*Client, error) {
	if transport == nil {
		transport = &http.Transport{}
	}

	httpClient := &http.Client{Transport: transport}

	codec := clientCodec{
		url:        url,
		httpClient: httpClient,
		ready:      make(chan uint64),
		responses:  make(map[uint64]*http.Response),
	}

	return &Client{rpc.NewClientWithCodec(&codec)}, nil
}
