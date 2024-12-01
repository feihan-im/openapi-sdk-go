package fhim

import (
	"context"
	"reflect"
	"sync"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

type v1MessageEvent struct {
	config     *fhcore.Config
	handlerMap sync.Map
}

type EventImV1MessageReceive struct {
	Header *fhcore.EventHeader          `json:"header,omitempty"`
	Body   *EventImV1MessageReceiveBody `json:"body,omitempty"`
}

type EventImV1MessageReceiveBody struct {
	Message *EventImV1MessageReceiveBodyMessage `json:"message,omitempty"`
}

type EventImV1MessageReceiveBodyMessage struct {
	MessageId        *string         `json:"message_id,omitempty"`
	MessageType      *string         `json:"message_type,omitempty"`
	MessageContent   *MessageContent `json:"message_content,omitempty"`
	MessageCreatedAt *uint64         `json:"message_created_at,omitempty"`
	ChatId           *string         `json:"chat_id,omitempty"`
	ChatSeqId        *uint64         `json:"chat_seq_id,omitempty"`
	SenderId         *UserId         `json:"sender_id,omitempty"`
}

func (v1 *v1MessageEvent) OnMessageReceive(handler func(ctx context.Context, event *EventImV1MessageReceive)) {
	var eventHandler fhcore.EventHandler = func(ctx context.Context, header *fhcore.EventHeader, body []byte) error {
		event := &EventImV1MessageReceive{Header: header, Body: &EventImV1MessageReceiveBody{}}
		if err := v1.config.JsonUnmarshal(body, event.Body); err != nil {
			return err
		}
		handler(ctx, event)
		return nil
	}
	v1.handlerMap.Store(reflect.ValueOf(handler).Pointer(), eventHandler)
	v1.config.ApiClient.OnEvent("im.v1.message.receive", eventHandler)
}

func (v1 *v1MessageEvent) OffMessageReceive(handler func(ctx context.Context, event *EventImV1MessageReceive)) {
	key := reflect.ValueOf(handler).Pointer()
	eventHandler, ok := v1.handlerMap.Load(key)
	if ok {
		v1.config.ApiClient.OffEvent("im.v1.message.receive", eventHandler.(fhcore.EventHandler))
		v1.handlerMap.Delete(key)
	}
}
