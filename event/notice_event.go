package event

type NoticeEvent struct {
	Event
	NoticeType string // 通知类型，group, private
}

// 群文件上传通知
type GroupUploadNoticeEvent struct {
	NoticeEvent
	GroupId int64 // 群号
	UserId  int64 // 上传者的QQ号
	File    struct {
		Id    string // 文件 ID
		Name  string // 文件名
		Size  int64  // 文件大小
		BusId string // 文件公众号 ID
	}
}

// 群管理员变动通知
type GroupAdminNoticeEvent struct {
	NoticeEvent
	SubType string // 通知子类型，set unset
	GroupId int64  // 群号
	UserId  int64  // 管理员 QQ 号
}

// 群成员增加通知
type GroupIncreaseNoticeEvent struct {
	NoticeEvent
	SubType    string // 通知子类型，approve, invite
	GroupId    int64  // 群号
	UserId     int64  // 新成员 QQ 号
	OperatorId int64  // 操作者 QQ 号
}

// 群成员减少通知
type GroupDecreaseNoticeEvent struct {
	NoticeEvent
	SubType    string // 通知子类型，leave, kick, kick_me
	GroupId    int64  // 群号
	UserId     int64  // 离开者 QQ 号
	OperatorId int64  // 操作者 QQ 号
}

// 群禁言通知
type GroupBanNoticeEvent struct {
	NoticeEvent
	SubType    string // 通知子类型，ban, lift_ban
	GroupId    int64  // 群号
	UserId     int64  // 被禁言 QQ 号
	OperatorId int64  // 操作者 QQ 号
	Duration   int64  // 禁言时长，单位秒
}

// 好友添加通知
type FriendAddNoticeEvent struct {
	NoticeEvent
	UserId int64 // 好友 QQ 号
}

// 群消息撤回通知
type GroupRecallNoticeEvent struct {
	NoticeEvent
	GroupId    int64 // 群号
	UserId     int64 // 撤回者 QQ 号
	OperatorId int64 // 操作者 QQ 号
	MessageId  int64 // 消息 ID
}

// 好友消息撤回通知
type FriendRecallNoticeEvent struct {
	NoticeEvent
	UserId     int64 // 撤回者 QQ 号
	MessageId  int64 // 消息 ID
}

// 戳一戳通知
type PokeNoticeEvent struct {
	NoticeEvent
	SubType string // 通知子类型，poke
	GroupId int64  // 群号
	UserId  int64  // 发送戳一戳的 QQ 号
	TargetId int64  // 被戳一戳的 QQ 号
}

// 运气王通知
type LuckyKingNoticeEvent struct {
	NoticeEvent
	SubType string // 通知子类型，lucky_king
	GroupId int64  // 群号
	UserId  int64  // 发红包者的 QQ 号
	TargetId int64  // 运气王的 QQ 号
}

// 群成员荣誉变更
type HonorNoticeEvent struct {
	NoticeEvent
	SubType string // 通知子类型，honor
	GroupId int64  // 群号
	UserId  int64  // QQ 号
	HonorType string // 荣誉类型，talkative、performer、emotion，分别表示龙王、群聊之火、快乐源泉
}