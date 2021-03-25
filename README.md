# httpclient

http client

## Install:

	go get github.com/go-light/httpclient/v2

# Usage

## Making a simple GET request
    
    headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	type Data struct {
		Method string `json:"method"`
	}

	type Reply struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		Data      []Data `json:"data"`
	}

	reply := Reply{}
	ret := NewClient("test.get",
		WithRetryCount(1),
		WithTimeout(2*time.Second),
		WithMaxIdleConns(20000),
		WithMaxIdleConnsPerHost(100),
	).Get(context.Background(), url, headers, &reply)

	fmt.Println(ret.LogEntry.Text())
	fmt.Printf("%+v\n", reply)  