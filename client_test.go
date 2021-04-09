package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	for i := 0; i < 100000; i++ {
		c := NewClientV3(WithRetryCount(1), WithTimeout(Duration(2*time.Second)),
			WithMaxIdleConns(0), WithMaxIdleConnsPerHost(0))

		fmt.Println(c)

		go func() {
			c := NewClientV3(WithRetryCount(1), WithTimeout(Duration(2*time.Second)))
			fmt.Println(c)
		}()
	}
}

func TestClient_Get(t *testing.T) {
	httpClient := NewClientV3(
		WithRetryCount(1),
		WithTimeout(Duration(1*time.Second)),
		WithMaxIdleConns(20000),
		WithMaxIdleConnsPerHost(100),
	)

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
	ret := httpClient.Get(context.Background(), server.URL, headers, &reply)
	require.NoError(t, ret.Error, "should not have failed to make a GET request")

	assert.Equal(t, http.StatusOK, ret.StatusCode)
	assert.Equal(t, "{ \"error_code\": 0, \"error_msg\": \"ok\", \"data\": [{\"method\":\"GET\"}] }", string(ret.Body))

	fmt.Println(ret.LogEntry.Text())
	fmt.Printf("%+v\n", reply)

	//time.Sleep(10*time.Second)
}

func TestClient_Post(t *testing.T) {
	httpClient := NewClientV3(
		WithRetryCount(1),
		WithTimeout(Duration(2*time.Second)),
		WithMaxIdleConns(20000),
		WithMaxIdleConnsPerHost(100),
	)

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
	ret := httpClient.Post(context.Background(), server.URL, bytes.NewBuffer([]byte(requestBodyString)), headers, &reply)
	require.NoError(t, ret.Error, "should not have failed to make a GET request")

	assert.Equal(t, http.StatusOK, ret.StatusCode)
	assert.Equal(t, "{ \"error_code\": 0, \"error_msg\": \"ok\", \"data\": [{\"method\":\"POST\"}] }", string(ret.Body))

	fmt.Println(ret.LogEntry.Text())
	fmt.Printf("%+v\n", reply)
}
