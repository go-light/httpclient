package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	for i := 0; i < 100000; i++ {
		c, err := NewClient("demo", WithRetryCount(1), WithTimeout("2s"),
			WithMaxIdleConns(0), WithMaxIdleConnsPerHost(0))
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println(c)

		go func() {
			c, err := NewClient("demo", WithRetryCount(1), WithTimeout("2s"))
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Println(c)
		}()
	}
}

func TestClient_Get(t *testing.T) {
	c, err := NewClient("test.get",
		WithRetryCount(1),
		WithTimeout("2s"),
		WithMaxIdleConns(20000),
		WithMaxIdleConnsPerHost(100),
	)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", c)

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "en", r.Header.Get("Accept-Language"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "error_code": 0, "error_msg": "ok", "data": [{"method":"GET"}] }`))
	}

	server := httptest.NewServer(http.HandlerFunc(dummyHandler))
	defer server.Close()

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept-Language", "en")

	type Data struct {
		Method string `json:"method"`
	}

	type Reply struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		Data      []Data `json:"data"`
	}

	reply := Reply{}
	ret := c.Get(context.Background(), server.URL, headers, &reply)
	require.NoError(t, err, "should not have failed to make a GET request")

	assert.Equal(t, http.StatusOK, ret.StatusCode)
	assert.Equal(t, "{ \"error_code\": 0, \"error_msg\": \"ok\", \"data\": [{\"method\":\"GET\"}] }", string(ret.Body))

	fmt.Println(ret.LogEntry.Text())
	fmt.Printf("%+v\n", reply)

	//time.Sleep(10*time.Second)
}

func TestClient_Post(t *testing.T) {
	c, err := NewClient("test.post",
		WithRetryCount(1),
		WithTimeout("2s"),
		WithMaxIdleConns(20000),
		WithMaxIdleConnsPerHost(100),
	)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", c)

	requestBodyString := `{ "name": "heimdall" }`

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "en", r.Header.Get("Accept-Language"))

		rBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err, "should not have failed to extract request body")

		assert.Equal(t, requestBodyString, string(rBody))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "error_code": 0, "error_msg": "ok", "data": [{"method":"POST"}] }`))
	}

	server := httptest.NewServer(http.HandlerFunc(dummyHandler))
	defer server.Close()

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept-Language", "en")

	type Data struct {
		Method string `json:"method"`
	}

	type Reply struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		Data      []Data `json:"data"`
	}

	reply := Reply{}
	ret := c.Post(context.Background(), server.URL, bytes.NewBuffer([]byte(requestBodyString)), headers, &reply)
	require.NoError(t, err, "should not have failed to make a GET request")

	assert.Equal(t, http.StatusOK, ret.StatusCode)
	assert.Equal(t, "{ \"error_code\": 0, \"error_msg\": \"ok\", \"data\": [{\"method\":\"POST\"}] }", string(ret.Body))

	fmt.Println(ret.LogEntry.Text())
	fmt.Printf("%+v\n", reply)
}
