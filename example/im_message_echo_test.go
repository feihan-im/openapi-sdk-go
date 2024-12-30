package example

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	fhsdk "github.com/feihan-im/openapi-sdk-go"
	fhim "github.com/feihan-im/openapi-sdk-go/service/im/v1"
)

func TestImMessageSend(t *testing.T) {
	client := fhsdk.NewClient(
		"http://localhost:11000",
		"c-TestAppId2",
		"TestAppSecret2",
	)
	onMessageReceive := func(ctx context.Context, event *fhim.EventImV1MessageReceive) {
		b, _ := json.MarshalIndent(event, "", "  ")
		_, _ = client.Im.Message.ReadMessage(ctx, &fhim.ReadMessageReq{
			MessageId: event.Body.Message.MessageId,
		})
		resp, _ := client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{
			MessageType: fhsdk.String(fhim.MessageType_Text),
			MessageContent: &fhim.MessageContent{
				Text: &fhim.MessageText{
					Content: fhsdk.String(fmt.Sprintf("Receive an event:\n%s", string(b))),
				},
			},
			ChatId:         event.Body.Message.ChatId,
			ReplyMessageId: event.Body.Message.MessageId,
		})
		_, _ = client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{
			MessageType:    event.Body.Message.MessageType,
			MessageContent: event.Body.Message.MessageContent,
			ChatId:         event.Body.Message.ChatId,
			ReplyMessageId: event.Body.Message.MessageId,
		})
		time.Sleep(2 * time.Second)
		_, _ = client.Im.Message.RecallMessage(ctx, &fhim.RecallMessageReq{
			MessageId: resp.MessageId,
		})
	}
	client.Im.Message.Event.OnMessageReceive(onMessageReceive)
	time.Sleep(10 * time.Second)
}
