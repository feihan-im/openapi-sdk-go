# Feihan IM OpenAPI SDK - Golang

[![Go Reference](https://pkg.go.dev/badge/github.com/feihan-im/openapi-sdk-go.svg)](https://pkg.go.dev/github.com/feihan-im/openapi-sdk-go)
[![Go](https://github.com/feihan-im/openapi-sdk-go/actions/workflows/go.yaml/badge.svg)](https://github.com/feihan-im/openapi-sdk-go/actions/workflows/go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/feihan-im/openapi-sdk-go)](https://goreportcard.com/report/github.com/feihan-im/openapi-sdk-go)
[![License](https://img.shields.io/github/license/feihan-im/openapi-sdk-go)](LICENSE)

English | [中文](README.md)

Feihan is a secure, self-hosted productivity platform, integrating instant messaging, organizational structures, video conferencing, and file storage.

This is the official Go SDK for Feihan server, used to interact with the Feihan server via OpenAPI. You need to deploy the Feihan server before using this SDK. See the [Quick Deploy Guide](https://feihanim.cn/docs/admin/install/quick-install) for setup instructions.

## Installation

```bash
go get github.com/feihan-im/openapi-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    fhsdk "github.com/feihan-im/openapi-sdk-go"
    fhim "github.com/feihan-im/openapi-sdk-go/service/im/v1"
)

func main() {
    client := fhsdk.NewClient("https://your-backend-url.com", "your-app-id", "your-app-secret")
    defer client.Close()

    // Optional: preheat fetches access token and syncs server time upfront,
    // reducing latency on the first API call
    _ = client.Preheat(context.Background())

    // Call API
    resp, err := client.Im.Message.SendMessage(context.Background(), &fhim.SendMessageReq{
        ChatId:      fhsdk.String("chat-id"),
        MessageType: fhsdk.String("text"),
        MessageContent: &fhim.MessageContent{
            Text: &fhim.MessageText{
                Content: fhsdk.String("Feihan new version released!"),
            },
        },
    })
    fmt.Println(resp, err)
}
```

## Configuration

`NewClient()` accepts optional functional options to configure client behavior:

```go
import (
    "time"

    fhsdk "github.com/feihan-im/openapi-sdk-go"
    fhcore "github.com/feihan-im/openapi-sdk-go/core"
)

client := fhsdk.NewClient(
    "https://your-backend-url.com",
    "your-app-id",
    "your-app-secret",
    fhsdk.WithLogLevel(fhcore.LoggerLevelDebug), // Log level (default: Info)
    fhsdk.WithRequestTimeout(30 * time.Second),             // Request timeout (default: 60s)
    fhsdk.WithEnableEncryption(false),                      // Enable request encryption (default: true)
)
```

## Event Subscription

Receive real-time events via WebSocket:

```go
// Register event handler
client.Im.Message.Event.OnMessageReceive(func(ctx context.Context, event *fhim.EventMessageReceive) {
    fmt.Println("Message received:", event.Body.Message.MessageId)
})
```

## Error Handling

When an API call fails, the returned `error` can be type-asserted to `*fhcore.ApiError`, which contains `Code`, `Msg`, and `LogId` fields:

```go
import fhcore "github.com/feihan-im/openapi-sdk-go/core"

resp, err := client.Im.Message.SendMessage(ctx, req)
if err != nil {
    if apiErr, ok := err.(*fhcore.ApiError); ok {
        fmt.Printf("API error: code=%d, msg=%s, logId=%s\n", apiErr.Code, apiErr.Msg, apiErr.LogId)
    } else {
        fmt.Println("Request error:", err)
    }
}
```

## Requirements

- Go 1.12 or later

## Links

- [Website](https://feihanim.cn/)

## License

[Apache-2.0 License](LICENSE)
