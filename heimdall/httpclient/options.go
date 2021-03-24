package httpclient

import (
	heimdall "github.com/go-light/httpclient/heimdall"
	"time"
)

// Option represents the client options
type Option func(*Client)

// WithHTTPTimeout sets hystrix timeout
func WithHTTPTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetryCount sets the retry count for the hystrixHTTPClient
func WithRetryCount(retryCount int) Option {
	return func(c *Client) {
		c.retryCount = retryCount
	}
}

// WithRetrier sets the strategy for retrying
func WithRetrier(retrier heimdall.Retriable) Option {
	return func(c *Client) {
		c.retrier = retrier
	}
}

// WithHTTPClient sets a custom http client
func WithHTTPClient(client heimdall.Doer) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithMaxIdleConns sets controls the maximum number of idle (keep-alive)
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Client) {
		c.maxIdleConns = maxIdleConns
	}
}

// WithMaxIdleConnsPerHost sets controls the maximum idle
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *Client) {
		c.maxIdleConnsPerHost = maxIdleConnsPerHost
	}
}
