package example

import (
	"context"
	"log"
	"sync"
	"testing"

	fhsdk "github.com/feihan-im/openapi-sdk-go"
	fhcore "github.com/feihan-im/openapi-sdk-go/core"
	fhim "github.com/feihan-im/openapi-sdk-go/service/im/v1"
)

func TestImMessageSend(t *testing.T) {
	ctx := context.Background()
	client := fhsdk.NewClient(
		"http://localhost:11000",
		"c-TestAppId1",
		"TestAppSecret1",
	)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Im.Message.SendMessage(ctx, &fhim.SendMessageReq{})
			if err != nil {
				panic(err)
			}
			log.Println(fhcore.Pretty(resp))
		}()
	}
	wg.Wait()
}
