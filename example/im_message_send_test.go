package example

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	fhsdk "github.com/feihan-im/openapi-sdk-go"
	fhim "github.com/feihan-im/openapi-sdk-go/service/im/v1"
)

func TestImMessageSend(t *testing.T) {
	ctx := context.Background()
	client := fhsdk.NewClient(
		"http://localhost:11000",
		"c-TestAppId1",
		"TestAppSecret1",
	)
	onMessageReceive := func(ctx context.Context, event *fhim.EventImV1MessageReceive) {
		log.Println("onMessageReceive: " + event.Header.EventId)
	}
	client.Im.Message.Event.OnMessageReceive(onMessageReceive)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{})
			if err != nil {
				panic(err)
			}
			log.Println(fhsdk.Pretty(resp))
		}()
	}
	wg.Wait()
	time.Sleep(500 * time.Millisecond)
	client.Im.Message.Event.OffMessageReceive(onMessageReceive)

	wg = sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{})
			if err != nil {
				panic(err)
			}
			log.Println(fhsdk.Pretty(resp))
		}()
	}
	wg.Wait()

	time.Sleep(500 * time.Millisecond)

	err := client.Close()
	if err != nil {
		panic(err)
	}
}
