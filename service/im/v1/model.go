package fhim

const (
	MessageTypeText              string = "text"
	MessageTypeImage             string = "image"
	MessageTypeSticker           string = "sticker"
	MessageTypeVideo             string = "video"
	MessageTypeAudio             string = "audio"
	MessageTypeFile              string = "file"
	MessageTypeUserCard          string = "user_card"
	MessageTypeGroupCard         string = "group_card"
	MessageTypeArticleLink       string = "article_link"
	MessageTypeLocation          string = "location"
	MessageTypeVote              string = "vote"
	MessageTypeMergeForward      string = "merge_forward"
	MessageTypeGroupAnnouncement string = "group_announcement"
	MessageTypeDriveCard         string = "drive_card"
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
	Content *string `json:"content,omitempty"`
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
	UserId      *string `json:"user_id,omitempty"`
	UnionUserId *string `json:"union_user_id,omitempty"`
	OpenUserId  *string `json:"open_user_id,omitempty"`
}
