package httpclient

import (
	"context"
	"fmt"
	"net/url"
	"testing"
)

func TestClient_Get(t *testing.T) {

	c, err := NewClient("a", WithRetryCount(1), WithTimeout("2s"))
	if err != nil {
		t.Error(err)
		return
	}

	type Data struct {
		Tag   string `json:"tag"`
		IsHit int    `json:"is_hit"`
	}

	type Reply struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		Data      []Data `json:"data"`
	}

	urlValues := url.Values{}

	reply := &Reply{}
	ret := c.Get(context.Background(), "http://127.0.0.1:8011/v1/open/activity/testing/hit/list?"+urlValues.Encode(), nil, reply)
	if ret.Error != nil {
		t.Error(ret.Error)
		fmt.Println(ret.LogEntry.Text())
		return
	}

	fmt.Println(ret.LogEntry.Text(), reply)

	//time.Sleep(10*time.Second)
}

func TestClient_Post(t *testing.T) {
	httpClient, err := NewClient("test", WithTimeout("1s"), WithRetryCount(1))
	if err != nil {
		t.Error(err)
		return
	}

	ret := httpClient.Post(context.Background(), "https://github.com/go-light/", nil, nil, nil)
	if ret.Error != nil {
		t.Error(ret.Error)
		return
	}

	fmt.Println(ret.LogEntry.Text(), string(ret.Body))
}
