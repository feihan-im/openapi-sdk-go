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
}

type SendMessageResp struct {
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
