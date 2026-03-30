// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhim

import (
	"sync"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

type v1Message struct {
	config *fhcore.Config
	Event  *v1MessageEvent
}

type v1MessageEvent struct {
	config     *fhcore.Config
	handlerMap sync.Map
}
