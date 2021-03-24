package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-light/httpclient/heimdall"

	xhttpclient "github.com/go-light/httpclient/heimdall/httpclient"
	"github.com/go-light/logentry"
	"github.com/pkg/errors"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 1 * time.Second

	defaultMaxIdleConns        = 20000
	defaultMaxIdleConnsPerHost = 1000
)

var (
	httpClientMap sync.Map
	mu            sync.Mutex
)

type myHTTPClient struct {
	client http.Client
}

func (c *myHTTPClient) Do(request *http.Request) (*http.Response, error) {
	return c.client.Do(request)
}

type HttpClient interface {
	Get(ctx context.Context, url string, headers http.Header, res interface{}) (ret *Resp)
	Post(ctx context.Context, url string, body io.Reader, headers http.Header, res interface{}) (ret *Resp)
}

type Client struct {
	xhttpclient *xhttpclient.Client
	timeout     time.Duration
	retryCount  int

	maxIdleConns        int
	maxIdleConnsPerHost int
}

type Resp struct {
	StatusCode int
	Body       []byte
	Error      error
	LogEntry   logentry.HttpClientLogEntry
}

func NewClient(name string, options ...Option) HttpClient {
	if val, ok := httpClientMap.Load(name); ok {
		return val.(*Client)
	}

	mu.Lock()
	defer mu.Unlock()

	if val, ok := httpClientMap.Load(name); ok {
		return val.(*Client)
	}

	client := &Client{
		timeout:    defaultHTTPTimeout,
		retryCount: defaultRetryCount,
	}
	for _, o := range options {
		o.Apply(client)
	}

	if client.maxIdleConns == 0 {
		client.maxIdleConns = defaultMaxIdleConns
	}

	if client.maxIdleConnsPerHost == 0 {
		client.maxIdleConnsPerHost = defaultMaxIdleConnsPerHost
	}

	var rt http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        client.maxIdleConns,
		MaxIdleConnsPerHost: client.maxIdleConnsPerHost, // see https://github.com/golang/go/issues/13801
		// 5 minutes is typically above the maximum sane scrape interval. So we can
		// use keepalive for all configurations.
		IdleConnTimeout:       5 * time.Minute,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client.xhttpclient = xhttpclient.NewClient(
		xhttpclient.WithHTTPTimeout(client.timeout),
		xhttpclient.WithHTTPClient(&myHTTPClient{
			// replace with custom HTTP client
			client: http.Client{
				Transport: rt,
				Timeout:   client.timeout,
			},
		}),
		xhttpclient.WithRetryCount(client.retryCount),
		xhttpclient.WithRetrier(heimdall.NewRetrier(heimdall.NewConstantBackoff(1*time.Millisecond, 5*time.Millisecond))),
	)
	httpClientMap.Store(name, client)

	return client
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

	respSizeBytes = fmt.Sprintf("%d", len(respBody))

	if res != nil {
		err := json.Unmarshal(respBody, &res)
		if err != nil {
			ret.Error = err
			return
		}
	}

	return
}
