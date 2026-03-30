// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhsdk

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
	fhserviceim "github.com/feihan-im/openapi-sdk-go/service/im"
)

type Client struct {
	config    *fhcore.Config
	ApiClient fhcore.ApiClient
	Im        *fhserviceim.Service
}

// Creates a client
func NewClient(backendUrl string, appId string, appSecret string, options ...clientOptionFunc) *Client {
	// init option
	option := &clientOption{
		logLevel:         fhcore.LoggerLevelInfo,
		requestTimeout:   1 * time.Minute,
		enableEncryption: true,
		jsonMarshaller:   json.Marshal,
		jsonUnmarshaller: json.Unmarshal,
	}
	for _, fn := range options {
		fn(option)
	}

	// init config
	config := &fhcore.Config{
		AppId:            appId,
		AppSecret:        appSecret,
		BackendUrl:       strings.TrimSpace(strings.TrimSuffix(backendUrl, "/")),
		HttpClient:       option.httpClient,
		EnableEncryption: option.enableEncryption,
		RequestTimeout:   option.requestTimeout,
		TimeManager:      option.timeManager,
		Logger:           option.logger,
		JsonMarshal:      option.jsonMarshaller,
		JsonUnmarshal:    option.jsonUnmarshaller,
	}
	if config.TimeManager == nil {
		config.TimeManager = fhcore.NewDefaultTimeManager()
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

// Preheating to prevent delay in the first request
func (c *Client) Preheat(ctx context.Context) error {
	return c.ApiClient.Preheat(ctx)
}

// Close client
func (c *Client) Close() error {
	return c.ApiClient.Close()
}
