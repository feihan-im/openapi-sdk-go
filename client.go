package fhsdk

import (
	fhcore "github.com/feihan-im/openapi-sdk-go/core"
	fhserviceim "github.com/feihan-im/openapi-sdk-go/service/im"
)

type Client struct {
	config    *fhcore.Config
	ApiClient fhcore.ApiClient
	Im        *fhserviceim.Service
}

// NewClient creates a client
func NewClient(backendUrl string, appId string, appSecret string, options ...clientOptionFunc) *Client {
	// init option
	option := &clientOption{
		logLevel: fhcore.LoggerLevelInfo,
	}
	for _, fn := range options {
		fn(option)
	}

	// init config
	config := &fhcore.Config{
		AppId:      appId,
		AppSecret:  appSecret,
		BackendUrl: backendUrl,
		HttpClient: option.httpClient,
		Logger:     option.logger,
	}
	if config.Logger == nil {
		config.Logger = fhcore.NewDefaultLogger(option.logLevel)
	}
	if config.HttpClient == nil {
		config.HttpClient = fhcore.NewDefaultHttpClient(option.requestTimeout)
	}
	config.ApiClient = fhcore.NewDefaultApiClient(config)

	// init client
	client := &Client{
		config:    config,
		ApiClient: config.ApiClient,
		Im:        fhserviceim.New(config),
	}

	return client
}
