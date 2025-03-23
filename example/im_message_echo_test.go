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
							Title: fhsdk.String("飞函新版本发布！"),
							TitleI18n: map[string]string{
								"en":      "Feihan new version released!",
								"zh-Hant": "飛函新版發布！",
							},
							Template: fhsdk.String("green"),
						},
						Body: &fhim.MessageCardV1Body{
							MessageText: &fhim.MessageText{
								Content: fhsdk.String("新版本特性:\n新增「夜间模式」主题，支持根据时间自动切换，添加多语言支持\n修复 iOS 端视频播放时闪退的兼容性问题"),
							},
							MessageTextI18n: map[string]*fhim.MessageText{
								"en": {
									Content: fhsdk.String("New version features: \nNew 'Night Mode' theme, support for automatic switching according to time, added multi-language support\nFixed the compatibility issue of flash back when playing videos on iOS"),
								},
								"zh-Hant": {
									Content: fhsdk.String("新版本特性:\n新增「夜間模式」主題，支援根據時間自動切換，新增多語言支援\n修復iOS端影片播放時閃退的相容性問題"),
								},
							},
						},
						Footer: &fhim.MessageCardV1Footer{
							ButtonList: []*fhim.MessageCardV1Button{{
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("default"),
							}, {
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("primary_filled"),
							}, {
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("primary"),
							}, {
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("danger"),
							}, {
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("danger_filled"),
							}, {
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
								},
								Link: &fhim.MessageCardV1ButtonLink{
									Url: fhsdk.String("https://feihanim.cn/"),
								},
								Template: fhsdk.String("danger_text"),
							}, {
								ButtonText: fhsdk.String("跳转到官网"),
								ButtonTextI18n: map[string]string{
									"en":      "Jump to official website",
									"zh-Hant": "跳到官網",
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
