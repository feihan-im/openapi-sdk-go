package fhim

const (
	MessageTypeText              = "text"
	MessageTypeImage             = "image"
	MessageTypeSticker           = "sticker"
	MessageTypeVideo             = "video"
	MessageTypeAudio             = "audio"
	MessageTypeFile              = "file"
	MessageTypeUserCard          = "user_card"
	MessageTypeGroupCard         = "group_card"
	MessageTypeArticleLink       = "article_link"
	MessageTypeLocation          = "location"
	MessageTypeVote              = "vote"
	MessageTypeMergeForward      = "merge_forward"
	MessageTypeGroupAnnouncement = "group_announcement"
	MessageTypeDriveCard         = "drive_card"
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
	ArticleLink       *MessageArticleLink       `json:"article_link,omitempty"`
	Location          *MessageLocation          `json:"location,omitempty"`
	Vote              *MessageVote              `json:"vote,omitempty"`
	MergeForward      *MessageMergeForward      `json:"merge_forward,omitempty"`
	GroupAnnouncement *MessageGroupAnnouncement `json:"group_announcement,omitempty"`
	DriveCard         *MessageDriveCard         `json:"drive_card,omitempty"`
}

type MessageText struct {
}

type MessageImage struct {
}

type MessageSticker struct {
}

type MessageVideo struct {
}

type MessageAudio struct {
}

type MessageFile struct {
}

type MessageUserCard struct {
}

type MessageGroupCard struct {
}

type MessageArticleLink struct {
}

type MessageLocation struct {
}

type MessageVote struct {
}

type MessageMergeForward struct {
}

type MessageGroupAnnouncement struct {
}

type MessageDriveCard struct {
}

type UserId struct {
	UserId  *string `json:"user_id,omitempty"`
	UnionId *string `json:"union_id,omitempty"`
	OpenId  *string `json:"open_id,omitempty"`
}
