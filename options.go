package httpclient

import "time"

// Option represents the client options
type Option interface {
	Apply(*Client)
}

// OptionFunc is a function that configures a client.
type OptionFunc func(*Client)

// Apply calls f(client)
func (f OptionFunc) Apply(client *Client) {
	f(client)
}

func WithTimeout(timeout time.Duration) Option {
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

// WithMaxIdleConns sets controls the maximum number of idle (keep-alive)
func WithMaxIdleConns(maxIdleConns int) Option {
	return OptionFunc(func(c *Client) {
		c.maxIdleConns = maxIdleConns
	})
}

// WithMaxIdleConnsPerHost sets controls the maximum idle
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return OptionFunc(func(c *Client) {
		c.maxIdleConnsPerHost = maxIdleConnsPerHost
	})
}
