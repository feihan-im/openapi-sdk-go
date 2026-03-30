// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhcore

import "time"

type Config struct {
	AppId      string
	AppSecret  string
	BackendUrl string

	HttpClient       HttpClient
	ApiClient        ApiClient
	EnableEncryption bool
	RequestTimeout   time.Duration
	TimeManager      TimeManager

	Logger Logger

	JsonMarshal   Marshaller
	JsonUnmarshal Unmarshaller
}
