// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

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
