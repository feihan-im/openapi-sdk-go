package fhim

import (
	"context"

	fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

type v1Message struct {
	config *fhcore.Config
	Event  *v1MessageEvent
}

type SendMessageReq struct {
	MessageType    *string         `json:"message_type,omitempty"`
	MessageContent *MessageContent `json:"message_content,omitempty"`
	ChatId         *string         `json:"chat_id,omitempty"`
	ReplyMessageId *string         `json:"reply_message_id,omitempty"`
}

type SendMessageResp struct {
	MessageId *string `json:"message_id,omitempty"`
}

func (v1 *v1Message) SendMessage(ctx context.Context, req *SendMessageReq) (*SendMessageResp, error) {
	apiResp, err := v1.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
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

type RecallMessageReq struct {
	MessageId *string `json:"message_id,omitempty"`
}

type RecallMessageResp struct {
}

func (v1 *v1Message) RecallMessage(ctx context.Context, req *RecallMessageReq) (*RecallMessageResp, error) {
	apiResp, err := v1.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
		Method: "POST",
		Path:   "/oapi/im/v1/messages/:message_id/recall",
		PathParams: map[string]string{
			"message_id": stringOrEmpty(req.MessageId),
		},
		Body:               req,
		WithAppAccessToken: true,
		WithWebSocket:      true,
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

type ReadMessageReq struct {
	MessageId *string `json:"message_id,omitempty"`
}

type ReadMessageResp struct {
}

func (v1 *v1Message) ReadMessage(ctx context.Context, req *ReadMessageReq) (*ReadMessageResp, error) {
	apiResp, err := v1.config.ApiClient.Request(ctx, &fhcore.ApiRequest{
		Method: "POST",
		Path:   "/oapi/im/v1/messages/:message_id/read",
		PathParams: map[string]string{
			"message_id": stringOrEmpty(req.MessageId),
		},
		Body:               req,
		WithAppAccessToken: true,
		WithWebSocket:      true,
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
