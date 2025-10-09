package utils

import (
	"fmt"
	"io"
	"net/http"
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

func MakeRequest(method string, path string, headers http.Header, body io.Reader) (resp *http.Response, err error) {

	baseUrl := os.Getenv("PROXYURL")
	if baseUrl == "" {
		panic("PROXYURL environment variable is not set")
	}

	requestURL := fmt.Sprintf("%s%s", baseUrl, path)
	req, err := http.NewRequest(method, requestURL, body)
	req.Header = headers
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
