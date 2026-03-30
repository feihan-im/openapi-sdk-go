// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhim

import (
	"context"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

// 发送消息（请求）
type SendMessageReq struct {
	MessageType    *string         `json:"message_type,omitempty"`     // 消息类型
	MessageContent *MessageContent `json:"message_content,omitempty"`  // 消息内容
	ChatId         *string         `json:"chat_id,omitempty"`          // 聊天 id
	ReplyMessageId *string         `json:"reply_message_id,omitempty"` // 本条消息回复的消息 id
}

// 发送消息（响应）
type SendMessageResp struct {
	MessageId *string `json:"message_id,omitempty"` // 消息 id
}

// 发送消息
func (impl *v1Message) SendMessage(ctx context.Context, req *SendMessageReq) (*SendMessageResp, error) {
	apiResp, err := impl.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
		Method:             "POST",
		Path:               "/oapi/im/v1/messages",
		Body:               req,
		WithAppAccessToken: true,
		WithWebSocket:      true,
	})
	if err != nil {
		return nil, err
	}
	var resp SendMessageResp
	if err = apiResp.JSON(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取消息（请求）
type GetMessageReq struct {
	MessageId *string `json:"message_id,omitempty"` // 消息 id
}

// 获取消息（响应）
type GetMessageResp struct {
	Message *Message `json:"message,omitempty"` // 消息
}

// 获取消息
func (impl *v1Message) GetMessage(ctx context.Context, req *GetMessageReq) (*GetMessageResp, error) {
	apiResp, err := impl.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
		Method: "GET",
		Path:   "/oapi/im/v1/messages/:message_id",
		Body:   req,
		PathParams: map[string]string{
			"message_id": stringOrEmpty(req.MessageId),
		},
		WithAppAccessToken: true,
	})
	if err != nil {
		return nil, err
	}
	var resp GetMessageResp
	if err = apiResp.JSON(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 撤回消息（请求）
type RecallMessageReq struct {
	MessageId *string `json:"message_id,omitempty"` // 消息 id
}

// 撤回消息（响应）
type RecallMessageResp struct {
}

// 撤回消息
func (impl *v1Message) RecallMessage(ctx context.Context, req *RecallMessageReq) (*RecallMessageResp, error) {
	apiResp, err := impl.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
		Method: "POST",
		Path:   "/oapi/im/v1/messages/:message_id/recall",
		Body:   req,
		PathParams: map[string]string{
			"message_id": stringOrEmpty(req.MessageId),
		},
		WithAppAccessToken: true,
	})
	if err != nil {
		return nil, err
	}
	var resp RecallMessageResp
	if err = apiResp.JSON(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 阅读消息（请求）
type ReadMessageReq struct {
	MessageId *string `json:"message_id,omitempty"` // 消息 id
}

// 阅读消息（响应）
type ReadMessageResp struct {
}

// 阅读消息
func (impl *v1Message) ReadMessage(ctx context.Context, req *ReadMessageReq) (*ReadMessageResp, error) {
	apiResp, err := impl.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
		Method: "POST",
		Path:   "/oapi/im/v1/messages/:message_id/read",
		Body:   req,
		PathParams: map[string]string{
			"message_id": stringOrEmpty(req.MessageId),
		},
		WithAppAccessToken: true,
	})
	if err != nil {
		return nil, err
	}
	var resp ReadMessageResp
	if err = apiResp.JSON(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
