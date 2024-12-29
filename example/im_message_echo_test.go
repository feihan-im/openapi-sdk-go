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
		_, _ = client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{
			MessageType: fhsdk.String(fhim.MessageTypeText),
			MessageContent: &fhim.MessageContent{
				Text: &fhim.MessageText{
					Content: fhsdk.String(fmt.Sprintf("Receive an event:\n%s", string(b))),
				},
			},
			ChatId:         event.Body.Message.ChatId,
			ReplyMessageId: event.Body.Message.MessageId,
		})
	}
	client.Im.Message.Event.OnMessageReceive(onMessageReceive)
	time.Sleep(10 * time.Second)
}
