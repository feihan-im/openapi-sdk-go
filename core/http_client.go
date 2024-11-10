package fhcore

import (
	"net/http"
	"time"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewDefaultHttpClient(requestTimeout time.Duration) HttpClient {
	if requestTimeout == 0 {
		return http.DefaultClient
	} else {
		return &http.Client{Timeout: requestTimeout}
	}
}
