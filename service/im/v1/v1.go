// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhim

import (
	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

type V1 struct {
	Chat    *v1Chat
	Message *v1Message
}

func New(config *fhcore.Config) *V1 {
	return &V1{
		Chat:    &v1Chat{config: config},
		Message: &v1Message{config: config, Event: &v1MessageEvent{config: config}},
	}
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
