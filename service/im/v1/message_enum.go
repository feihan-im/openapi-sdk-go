// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhim

// 消息类型
const (
	MessageType_TEXT               string = "text"               // 文本消息
	MessageType_IMAGE              string = "image"              // 图片消息
	MessageType_STICKER            string = "sticker"            // 表情包消息
	MessageType_VIDEO              string = "video"              // 视频消息
	MessageType_AUDIO              string = "audio"              // 语音消息
	MessageType_FILE               string = "file"               // 文件消息
	MessageType_USER_CARD          string = "user_card"          // 个人名片消息
	MessageType_GROUP_CARD         string = "group_card"         // 群名片消息
	MessageType_GROUP_ANNOUNCEMENT string = "group_announcement" // 群公告
	MessageType_CARD               string = "card"               // 卡片消息
)

// 附件类型
const (
	MessageTextAttachmentType_IMAGE string = "image" // 图片
)

// 消息状态
const (
	MessageStatus_VISIBLE  string = "visible"  // 消息可见
	MessageStatus_RECALLED string = "recalled" // 消息已撤回
)
