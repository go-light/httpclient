package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"unsafe"

	"github.com/go-light/logentry"
	xhttpclient "github.com/gojektech/heimdall/v6/httpclient"
	"github.com/pkg/errors"
)

var httpClientMap sync.Map
var mu sync.Mutex

type HttpClient interface {
	Get(ctx context.Context, url string, headers http.Header, res interface{}) (ret *Resp)
	Post(ctx context.Context, url string, body io.Reader, headers http.Header, res interface{}) (ret *Resp)
}

type Client struct {
	xhttpclient *xhttpclient.Client
	timeout     string
	retryCount  int
}

type Resp struct {
	StatusCode int
	Body       []byte
	Error      error
	LogEntry   logentry.HttpClientLogEntry
}

func NewClient(name string, options ...Option) (HttpClient, error) {
	if val, ok := httpClientMap.Load(name); ok {
		return val.(*Client), nil
	}

	mu.Lock()
	defer mu.Unlock()

	if val, ok := httpClientMap.Load(name); ok {
		return val.(*Client), nil
	}

	client := &Client{
		timeout:    "1s",
		retryCount: 1,
	}
	for _, o := range options {
		o.Apply(client)
	}

	// Create a new HTTP client with a default timeout
	timeout, err := time.ParseDuration(client.timeout)
	if err != nil {
		return nil, err
	}

	client.xhttpclient = xhttpclient.NewClient(xhttpclient.WithHTTPTimeout(timeout), xhttpclient.WithRetryCount(client.retryCount))
	httpClientMap.Store(name, client)

	return client, nil
}

func (c *Client) Get(ctx context.Context, url string, headers http.Header, res interface{}) (ret *Resp) {
	return c.do(ctx, url, http.MethodGet, headers, nil, res)
}

func (c *Client) Post(ctx context.Context, url string, body io.Reader, headers http.Header, res interface{}) (ret *Resp) {
	return c.do(ctx, url, http.MethodPost, headers, body, res)
}

func (c *Client) do(ctx context.Context, url string, method string, headers http.Header, body io.Reader, res interface{}) (ret *Resp) {
	var (
		resp          *http.Response
		err           error
		statusCode    int
		respSizeBytes string
	)

	logEntry := logentry.NewHttpClientLogEntry(ctx)
	logEntry.Start()
	logEntry.SetReqUrl(url)
	logEntry.SetMethod(method)

	defer func() {
		logEntry.SetStatusCode(statusCode)
		logEntry.SetRespSizeBytes(respSizeBytes)
		logEntry.End()
		ret.LogEntry = logEntry
	}()

	ret = &Resp{
		StatusCode: 0,
		Body:       nil,
		Error:      nil,
	}

	httpClient := c.xhttpclient

	switch method {
	case http.MethodGet:
		// Use the clients GET method to create and execute the request
		resp, err = httpClient.Get(url, headers)
	case http.MethodPost:
		// Use the clients GET method to create and execute the request
		resp, err = httpClient.Post(url, body, headers)
	default:
		err = fmt.Errorf("undefined method")
		return
	}

	if err != nil {
		ret.Error = err
		return
	}

	statusCode = resp.StatusCode
	ret.StatusCode = statusCode

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ret.Error = err
		return
	}

	ret.Body = respBody
	if statusCode >= http.StatusBadRequest {
		ret.Error = errors.New(resp.Status)
		return
	}

	respSizeBytes = fmt.Sprintf("%d", unsafe.Sizeof(respBody))

	if res != nil {
		err := json.Unmarshal(respBody, &res)
		if err != nil {
			ret.Error = err
			return
		}
	}

	return
}

// An Option configures a mutex.
type Option interface {
	Apply(*Client)
}

// OptionFunc is a function that configures a mutex.
type OptionFunc func(*Client)

// Apply calls f(mutex)
func (f OptionFunc) Apply(mutex *Client) {
	f(mutex)
}

func WithTimeout(timeout string) Option {
	return OptionFunc(func(m *Client) {
		m.timeout = timeout
	})
}

// WithRetryCount can be used to
func WithRetryCount(retryCount int) Option {
	return OptionFunc(func(m *Client) {
		m.retryCount = retryCount
	})
}
