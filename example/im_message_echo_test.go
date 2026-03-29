package example

import (
	"context"
	"encoding/json"
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
	client.Im.Message.Event.OnMessageReceive(func(ctx context.Context, event *fhim.EventMessageReceive) {
		_, _ = client.Im.Chat.CreateTyping(ctx, &fhim.CreateTypingReq{
			ChatId: event.Body.Message.ChatId,
		})
		defer func() {
			_, _ = client.Im.Chat.DeleteTyping(ctx, &fhim.DeleteTypingReq{
				ChatId: event.Body.Message.ChatId,
			})
		}()
		_, _ = json.MarshalIndent(event, "", "  ")
		_, _ = client.Im.Message.ReadMessage(ctx, &fhim.ReadMessageReq{
			MessageId: event.Body.Message.MessageId,
		})
		_, err := client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{
			MessageType: fhsdk.String(fhim.MessageType_CARD),
			MessageContent: &fhim.MessageContent{
				Card: &fhim.MessageCard{
					Schema: fhsdk.String("1.0"),
					V1: &fhim.MessageCardV1{
						Header: &fhim.MessageCardV1Header{
							Title: fhsdk.String("Feihan new version released!"),
							TitleI18n: map[string]string{
								"en": "Feihan new version released!",
							},
							Template: fhsdk.String("green"),
						},
						Body: &fhim.MessageCardV1Body{
							MessageText: &fhim.MessageText{
								Content: fhsdk.String("New version features:\n- Added a Night Mode theme\n- Added multilingual support\n- Fixed the iOS video playback crash issue"),
							},
							MessageTextI18n: map[string]*fhim.MessageText{
								"en": {
									Content: fhsdk.String("New version features:\n- Added a Night Mode theme\n- Added multilingual support\n- Fixed the iOS video playback crash issue"),
								},
							},
						},
						Footer: &fhim.MessageCardV1Footer{
							ButtonList: []*fhim.MessageCardV1Button{{
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("default"),
							}, {
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("primary_filled"),
							}, {
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("primary"),
							}, {
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("danger"),
							}, {
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("danger_filled"),
							}, {
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("danger_text"),
							}, {
								ButtonText: fhsdk.String("Open website"),
								ButtonTextI18n: map[string]string{
									"en": "Jump to official website",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("primary_text"),
							}},
							ButtonAlign: fhsdk.String("start"),
						},
					},
				},
			},
			ChatId: event.Body.Message.ChatId,
			// ReplyMessageId: event.Body.Message.MessageId,
		})
		if err != nil {
			return
		}
		time.Sleep(2 * time.Second)
		// _, _ = client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{
		// 	MessageType:    event.Body.Message.MessageType,
		// 	MessageContent: event.Body.Message.MessageContent,
		// 	ChatId:         event.Body.Message.ChatId,
		// 	ReplyMessageId: event.Body.Message.MessageId,
		// })
		// _, _ = client.Im.Message.RecallMessage(ctx, &fhim.RecallMessageReq{
		// 	MessageId: resp.MessageId,
		// })
	})
	time.Sleep(10 * time.Second)
}
