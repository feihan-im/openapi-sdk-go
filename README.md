# 飞函 IM OpenAPI SDK - Golang

[![Go Reference](https://pkg.go.dev/badge/github.com/feihan-im/openapi-sdk-go.svg)](https://pkg.go.dev/github.com/feihan-im/openapi-sdk-go)
[![Go](https://github.com/feihan-im/openapi-sdk-go/actions/workflows/go.yaml/badge.svg)](https://github.com/feihan-im/openapi-sdk-go/actions/workflows/go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/feihan-im/openapi-sdk-go)](https://goreportcard.com/report/github.com/feihan-im/openapi-sdk-go)
[![License](https://img.shields.io/github/license/feihan-im/openapi-sdk-go)](LICENSE)

[English](README_en.md) | 中文

飞函，是安全稳定的私有化一站式办公平台，功能包括即时通讯、组织架构、音视频会议、网盘等。

本项目是飞函服务端的 Go SDK，用于通过 OpenAPI 与飞函服务端进行交互。使用前需要先自行部署飞函服务端，部署教程请参考[快速部署文档](https://feihanim.cn/docs/admin/install/quick-install)。

## 安装

```bash
go get github.com/feihan-im/openapi-sdk-go
```

## 快速开始

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

    // 可选：预热可提前获取访问凭证和同步服务端时间，减少首次调用的延迟
    _ = client.Preheat(context.Background())

    // 调用 API
    resp, err := client.Im.Message.SendMessage(context.Background(), &fhim.SendMessageReq{
        ChatId:      fhsdk.String("chat-id"),
        MessageType: fhsdk.String("text"),
        MessageContent: &fhim.MessageContent{
            Text: &fhim.MessageText{
                Content: fhsdk.String("飞函新版本发布！"),
            },
        },
    })
    fmt.Println(resp, err)
}
```

## 客户端配置

`NewClient()` 支持通过可选参数配置客户端行为：

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
    fhsdk.WithLogLevel(fhcore.LoggerLevelDebug), // 日志级别（默认: Info）
    fhsdk.WithRequestTimeout(30 * time.Second),             // 请求超时（默认: 60s）
    fhsdk.WithEnableEncryption(false),                      // 启用请求加密（默认: true）
)
```

## 事件订阅

通过 WebSocket 接收实时事件推送：

```go
// 注册事件处理函数
client.Im.Message.Event.OnMessageReceive(func(ctx context.Context, event *fhim.EventMessageReceive) {
    fmt.Println("收到消息:", event.Body.Message.MessageId)
})
```

## 错误处理

API 调用失败时返回的 `error` 可以断言为 `*fhcore.ApiError`，包含 `Code`、`Msg` 和 `LogId` 字段：

```go
import fhcore "github.com/feihan-im/openapi-sdk-go/core"

resp, err := client.Im.Message.SendMessage(ctx, req)
if err != nil {
    if apiErr, ok := err.(*fhcore.ApiError); ok {
        fmt.Printf("API 错误: code=%d, msg=%s, logId=%s\n", apiErr.Code, apiErr.Msg, apiErr.LogId)
    } else {
        fmt.Println("请求错误:", err)
    }
}
```

## 环境要求

- Go 1.12 及以上版本

## 相关链接

- [官网](https://feihanim.cn/)

## 许可证

[Apache-2.0 License](LICENSE)
