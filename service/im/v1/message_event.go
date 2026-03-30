// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhim

import (
	"context"
	"reflect"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

// 接收消息
type EventMessageReceive struct {
	Header *fhcore.EventHeader      `json:"header,omitempty"`
	Body   *EventMessageReceiveBody `json:"body,omitempty"`
}

// 接收消息
type EventMessageReceiveBody struct {
	Message *Message `json:"message,omitempty"` // 消息
}

// 接收消息
func (impl *v1MessageEvent) OnMessageReceive(handler func(ctx context.Context, event *EventMessageReceive)) {
	var eventHandler fhcore.EventHandler = func(ctx context.Context, header *fhcore.EventHeader, body []byte) error {
		event := &EventMessageReceive{Header: header, Body: &EventMessageReceiveBody{}}
		if err := impl.config.JsonUnmarshal(body, event.Body); err != nil {
			return err
		}
		handler(ctx, event)
		return nil
	}
	impl.handlerMap.Store(reflect.ValueOf(handler).Pointer(), eventHandler)
	impl.config.ApiClient.OnEvent("im.v1.message.receive", eventHandler)
}

func (impl *v1MessageEvent) OffMessageReceive(handler func(ctx context.Context, event *EventMessageReceive)) {
	key := reflect.ValueOf(handler).Pointer()
	eventHandler, ok := impl.handlerMap.Load(key)
	if ok {
		impl.config.ApiClient.OffEvent("im.v1.message.receive", eventHandler.(fhcore.EventHandler))
		impl.handlerMap.Delete(key)
	}
}
