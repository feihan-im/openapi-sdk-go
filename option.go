package fhsdk

import (
	"time"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

type clientOption struct {
	httpClient     fhcore.HttpClient
	requestTimeout time.Duration

	logLevel fhcore.LoggerLevel
	logger   fhcore.Logger
}

type clientOptionFunc func(option *clientOption)

func WithHttpClient(httpClient fhcore.HttpClient) clientOptionFunc {
	return func(option *clientOption) {
		option.httpClient = httpClient
	}
}

func WithRequestTimeout(requestTimeout time.Duration) clientOptionFunc {
	return func(option *clientOption) {
		option.requestTimeout = requestTimeout
	}
}

func WithLogLevel(logLevel fhcore.LoggerLevel) clientOptionFunc {
	return func(option *clientOption) {
		option.logLevel = logLevel
	}
}

func WithLogger(logger fhcore.Logger) clientOptionFunc {
	return func(option *clientOption) {
		option.logger = logger
	}
}
