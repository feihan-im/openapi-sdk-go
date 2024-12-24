package fhsdk

import (
	"time"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

type clientOption struct {
	httpClient       fhcore.HttpClient
	requestTimeout   time.Duration
	enableEncryption bool

	logLevel    fhcore.LoggerLevel
	logger      fhcore.Logger
	timeManager fhcore.TimeManager

	jsonMarshaller   fhcore.Marshaller
	jsonUnmarshaller fhcore.Unmarshaller
}

type clientOptionFunc func(option *clientOption)

func WithHttpClient(httpClient fhcore.HttpClient) clientOptionFunc {
	return func(option *clientOption) {
		option.httpClient = httpClient
	}
}

func WithRequestTimeout(requestTimeout time.Duration) clientOptionFunc {
	return func(option *clientOption) {
		if requestTimeout > 0 {
			option.requestTimeout = requestTimeout
		}
	}
}

func WithEnableEncryption(enableEncryption bool) clientOptionFunc {
	return func(option *clientOption) {
		option.enableEncryption = enableEncryption
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

func WithTimeManager(timeManager fhcore.TimeManager) clientOptionFunc {
	return func(option *clientOption) {
		option.timeManager = timeManager
	}
}

func WithJsonMarshaller(jsonMarshaller fhcore.Marshaller) clientOptionFunc {
	return func(option *clientOption) {
		option.jsonMarshaller = jsonMarshaller
	}
}

func WithJsonUnmarshaller(jsonUnmarshaller fhcore.Unmarshaller) clientOptionFunc {
	return func(option *clientOption) {
		option.jsonUnmarshaller = jsonUnmarshaller
	}
}
