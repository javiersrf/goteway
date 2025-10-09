package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const TIMEOUT = 120

var client http.Client

func InitClient() {
	var timeoutInt int
	timeoutInt = TIMEOUT
	requestTimeout := os.Getenv("REQUEST_TIMEOUT")

	timeoutInt, err := strconv.Atoi(requestTimeout)
	if err != nil {
		fmt.Printf("Error starting http client: %v\n", err)
		timeoutInt = TIMEOUT
	}

	client = http.Client{
		Timeout: time.Duration(timeoutInt) * time.Second,
	}
}

func MakeRequest(method string, path string, headers http.Header, query map[string]string, body io.Reader) (*http.Response, error) {
	baseURL := os.Getenv("PROXYURL")
	if baseURL == "" {
		panic("PROXYURL environment variable is not set")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid PROXYURL: %w", err)
	}

	u.Path = fmt.Sprintf("%s%s", u.Path, path)

	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		req.Header = headers
	}

	return client.Do(req)
}
