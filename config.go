package httpclient

import (
	xtime "time"
)

// Duration be used toml unmarshal string time, like 1s, 500ms.
type Duration xtime.Duration

// UnmarshalText unmarshal text to duration.
func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := xtime.ParseDuration(string(text))
	if err == nil {
		*d = Duration(tmp)
	}
	return err
}

// ClientConfig is http client conf.
type ClientConfig struct {
	Name       string
	Timeout    Duration
	RetryCount int

	MaxIdleConns        int
	MaxIdleConnsPerHost int
}
