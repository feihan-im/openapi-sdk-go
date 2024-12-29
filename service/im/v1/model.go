package fhim

const (
	MessageType_Text              string = "text"
	MessageType_Image             string = "image"
	MessageType_Sticker           string = "sticker"
	MessageType_Video             string = "video"
	MessageType_Audio             string = "audio"
	MessageType_File              string = "file"
	MessageType_UserCard          string = "user_card"
	MessageType_GroupCard         string = "group_card"
	MessageType_GroupAnnouncement string = "group_announcement"
)

type MessageContent struct {
	Text              *MessageText              `json:"text,omitempty"`
	Image             *MessageImage             `json:"image,omitempty"`
	Sticker           *MessageSticker           `json:"sticker,omitempty"`
	Video             *MessageVideo             `json:"video,omitempty"`
	Audio             *MessageAudio             `json:"audio,omitempty"`
	File              *MessageFile              `json:"file,omitempty"`
	UserCard          *MessageUserCard          `json:"user_card,omitempty"`
	GroupCard         *MessageGroupCard         `json:"group_card,omitempty"`
	GroupAnnouncement *MessageGroupAnnouncement `json:"group_announcement,omitempty"`
}

type MessageText struct {
	Content         *string                   `json:"content,omitempty"`
	AttachmentList  []*MessageTextAttachment  `json:"attachment_list,omitempty"`
	MentionUserList []*MessageTextMentionUser `json:"mention_user_list,omitempty"`
	EmojiList       []*MessageTextEmoji       `json:"emoji_list,omitempty"`
}

type MessageTextAttachment struct {
	AttachmentId      *string                       `json:"attachment_id,omitempty"`
	AttachmentType    *string                       `json:"attachment_type,omitempty"`
	AttachmentContent *MessageTextAttachmentContent `json:"attachment_content,omitempty"`
}

const (
	MessageTextAttachmentType_Image string = "image"
)

type MessageTextAttachmentContent struct {
	Image *FileImage `json:"image,omitempty"`
}

type MessageTextMentionUser struct {
	UserId   *UserId `json:"user_id,omitempty"`
	UserName *string `json:"user_name,omitempty"`
	IsInChat *bool   `json:"is_in_chat,omitempty"`
}

type MessageTextEmoji struct {
	EmojiId   *string `json:"emoji_id,omitempty"`
	EmojiName *string `json:"emoji_name,omitempty"`
}

type MessageImage struct {
	Image *FileImage `json:"image,omitempty"`
}

type MessageSticker struct {
	Sticker *Sticker `json:"sticker,omitempty"`
}

type MessageVideo struct {
	Video *FileVideo `json:"video,omitempty"`
}

type MessageAudio struct {
	Audio *FileAudio `json:"audio,omitempty"`
}

type MessageFile struct {
	File     *File   `json:"file,omitempty"`
	Filename *string `json:"filename,omitempty"`
}

type MessageUserCard struct {
	UserId *UserId `json:"user_id,omitempty"`
}

type MessageGroupCard struct {
	ChatId *string `json:"chat_id,omitempty"`
}

type MessageGroupAnnouncement struct {
	MessageText *MessageText `json:"message_text,omitempty"`
}

type UserId struct {
	UserId      *string `json:"user_id,omitempty"`
	UnionUserId *string `json:"union_user_id,omitempty"`
	OpenUserId  *string `json:"open_user_id,omitempty"`
}

type Encryption struct {
	EncryptionAlgorithm *string `json:"encryption_algorithm,omitempty"`
	EncryptionKey       []byte  `json:"encryption_key,omitempty"`
	EncryptionSize      *uint64 `json:"encryption_size,omitempty"`
}

type File struct {
	FileId         *string     `json:"file_id,omitempty"`
	FileMime       *string     `json:"file_mime,omitempty"`
	FileEncryption *Encryption `json:"file_encryption,omitempty"`
}

type FileImage struct {
	Image              *File   `json:"image,omitempty"`
	ImageWidth         *uint64 `json:"image_width,omitempty"`
	ImageHeight        *uint64 `json:"image_height,omitempty"`
	ImageOrigin        *File   `json:"image_origin,omitempty"`
	ImageOriginWidth   *uint64 `json:"image_origin_width,omitempty"`
	ImageOriginHeight  *uint64 `json:"image_origin_height,omitempty"`
	ImageThumbBytes    []byte  `json:"image_thumb_bytes,omitempty"`
	ImageThumbMime     *string `json:"image_thumb_mime,omitempty"`
	ImageDominantColor *string `json:"image_dominant_color,omitempty"`
}

type FileVideo struct {
	Video         *File      `json:"video,omitempty"`
	VideoWidth    *uint64    `json:"video_width,omitempty"`
	VideoHeight   *uint64    `json:"video_height,omitempty"`
	VideoDuration *float64   `json:"video_duration,omitempty"`
	VideoPreview  *FileImage `json:"video_preview,omitempty"`
}

type FileAudio struct {
	Audio         *File   `json:"audio,omitempty"`
	AudioDuration *uint64 `json:"audio_duration,omitempty"`
}

type Emoji struct {
	EmojiId       *string           `json:"emoji_id,omitempty"`
	EmojiName     *string           `json:"emoji_name,omitempty"`
	EmojiNameI18n map[string]string `json:"emoji_name_i18n,omitempty"`
	EmojiImage    *FileImage        `json:"emoji_image,omitempty"`
}

type Sticker struct {
	StickerId       *string           `json:"sticker_id,omitempty"`
	StickerName     *string           `json:"sticker_name,omitempty"`
	StickerNameI18n map[string]string `json:"sticker_name_i18n,omitempty"`
	StickerImage    *FileImage        `json:"sticker_image,omitempty"`
}
