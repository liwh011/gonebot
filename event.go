package gonebot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const (
	PostType_MetaEvent    = "meta_event"
	PostType_MessageEvent = "message"
	PostType_NoticeEvent  = "notice"
	PostType_RequestEvent = "request"
)

type EventName string

const (
	EventName_AllEvent        EventName = "all"
	EventName_Message         EventName = "message"
	EventName_PrivateMessage  EventName = "message.private"
	EventName_GroupMessage    EventName = "message.group"
	EventName_Notice          EventName = "notice"
	EventName_GroupUpload     EventName = "notice.group_upload"
	EventName_GroupAdmin      EventName = "notice.group_admin"
	EventName_GroupDecrease   EventName = "notice.group_decrease"
	EventName_GroupIncrease   EventName = "notice.group_increase"
	EventName_GroupBan        EventName = "notice.group_ban"
	EventName_FriendAdd       EventName = "notice.friend_add"
	EventName_GroupRecall     EventName = "notice.group_recall"
	EventName_FriendRecall    EventName = "notice.friend_recall"
	EventName_Notify          EventName = "notice.notify"
	EventName_NotifyPoke      EventName = "notice.notify.poke"
	EventName_NotifyLuckyKing EventName = "notice.notify.lucky_king"
	EventName_NotifyHonor     EventName = "notice.notify.honor"
	EventName_Request         EventName = "request"
	EventName_RequestFriend   EventName = "request.friend"
	EventName_RequestGroup    EventName = "request.group"
	EventName_Meta            EventName = "meta_event"
	EventName_MetaLifecycle   EventName = "meta_event.lifecycle"
	EventName_MetaHeartbeat   EventName = "meta_event.heartbeat"
)

var eventTypeMap map[string]I_Event

func init() {
	eventTypeMap = map[string]I_Event{
		"message.private":          &PrivateMessageEvent{},
		"message.group":            &GroupMessageEvent{},
		"notice.group_upload":      &GroupUploadNoticeEvent{},
		"notice.group_admin":       &GroupAdminNoticeEvent{},
		"notice.group_decrease":    &GroupDecreaseNoticeEvent{},
		"notice.group_increase":    &GroupIncreaseNoticeEvent{},
		"notice.group_ban":         &GroupBanNoticeEvent{},
		"notice.friend_add":        &FriendAddNoticeEvent{},
		"notice.group_recall":      &GroupRecallNoticeEvent{},
		"notice.friend_recall":     &FriendRecallNoticeEvent{},
		"notice.notify.poke":       &PokeNoticeEvent{},
		"notice.notify.lucky_king": &LuckyKingNoticeEvent{},
		"notice.notify.honor":      &HonorNoticeEvent{},
		"notice.group_card":        &GroupCardNoticeEvent{},
		"notice.offline_file":      &OfflineFileNoticeEvent{},
		"notice.client_status":     &ClientStatusNoticeEvent{},
		"notice.essence":           &EssenceNoticeEvent{},
		"request.friend":           &FriendRequestEvent{},
		"request.group":            &GroupRequestEvent{},
		"meta_event.lifecycle":     &LifeCycleMetaEvent{},
		"meta_event.heartbeat":     &HeartbeatMetaEvent{},
	}
}

type I_Event interface {
	GetPostType() string         // 获取事件的上报类型，有message, notice, request, meta_event
	GetSecondType() string       // 获取事件的第二级类型。
	GetSubType() string          // 获取第三级类型。部分事件没有这个字段，则返回空字符串
	GetEventName() EventName     // 获取事件的名称，即完整类型。形如：notice.group.set
	GetEventDescription() string // 获取事件的描述，一般用于日志输出

	IsMessageEvent() bool
	IsToMe() bool

	GetSessionId() string     // 获取事件的会话ID，用于区分不同的会话。私聊为"QQ号"，群聊中对应"QQ号@群号"。
	GetMessage() *Message     // 提取消息。非消息事件返回nil
	ExtractPlainText() string // 提取消息的纯文本。非消息事件返回空字符串
}

type Event struct {
	Time     int64  `json:"time"`      // 事件发生的时间戳
	SelfId   int64  `json:"self_id"`   // 收到事件的机器人的QQ号
	PostType string `json:"post_type"` // 事件的类型，message, notice, request, meta_event

	EventName EventName `json:"-"` // 事件的名称，形如：notice.group.set
	ToMe      bool      `json:"-"` // 是否与我（bot）有关（即私聊我、或群聊At我、我被踢了、等等）
}

// 获取事件的上报类型，有message, notice, request, meta_event
func (e *Event) GetPostType() string {
	return e.PostType
}

func (e *Event) GetSecondType() string {
	return ""
}

func (e *Event) GetSubType() string {
	return ""
}

// 获取事件的名称，形如：notice.group.set
func (e *Event) GetEventName() EventName {
	return e.EventName
}

// 获取事件的描述，一般用于日志输出
func (e *Event) GetEventDescription() string {
	return fmt.Sprintf("[%s]: %+v", e.EventName, *e)
}

func (e *Event) IsMessageEvent() bool {
	return e.PostType == PostType_MessageEvent
}

func (e *Event) IsToMe() bool {
	return e.ToMe
}

func (e *Event) GetSessionId() string {
	return ""
}

func (e *Event) GetMessage() *Message {
	return nil
}

func (e *Event) ExtractPlainText() string {
	return ""
}

// 从JSON对象中生成Event对象（指针）
func ConvertJsonObjectToEvent(obj gjson.Result) I_Event {
	// 大多数事件的事件类型有3级，而第三级的subtype通常不影响事件的结构
	// 所以下面只用了第一级和第二级类型来构造事件对象
	postType := obj.Get("post_type").String()
	nextType := obj.Get(postType + "_type").String()
	typeName := fmt.Sprintf("%s.%s", postType, nextType) // 前两段类型

	subType := ""
	fullTypeName := typeName
	if obj.Get("sub_type").Exists() {
		subType = obj.Get("sub_type").String()
		fullTypeName = fmt.Sprintf("%s.%s", typeName, subType)
	}

	var ev I_Event
	if event, ok := eventTypeMap[fullTypeName]; ok {
		ev = event
	} else if event, ok := eventTypeMap[typeName]; ok {
		ev = event
	} else {
		logrus.Warnf("暂未支持的事件类型 %s ，将转换为基本事件。", fullTypeName)
		ev = &Event{}
	}

	// 由于使用的是指针，需要拷贝一份Event对象再来填充数据
	ev = createUnderlyingStruct(ev).(I_Event)

	// 借助json库将JSON对象中的字段赋值给Event对象，懒得自个写反射了
	err := json.Unmarshal([]byte(obj.Raw), ev)
	if err != nil {
		panic(err)
	}

	// 设置事件的名称
	setEventField(ev, "EventName", EventName(fullTypeName))
	if isEventRelativeToBot(ev) {
		setEventField(ev, "ToMe", true)
	}

	return ev
}

type MessageEvent struct {
	Event
	MessageType string  `json:"message_type"` // 消息类型，group, private
	SubType     string  `json:"sub_type"`     // 消息子类型，friend, group, other | normal, anonymous, notice
	MessageId   int32   `json:"message_id"`   // 消息ID
	UserId      int64   `json:"user_id"`      // 消息发送者的QQ号
	Message     Message `json:"message"`      // 消息内容
	RawMessage  string  `json:"raw_message"`  // 原始消息内容
	Font        int32   `json:"font"`         // 字体

}

func (e *MessageEvent) GetSecondType() string {
	return e.MessageType
}

func (e *MessageEvent) GetSubType() string {
	return e.SubType
}

func (e *MessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

func (e *MessageEvent) GetMessage() *Message {
	return &e.Message
}

func (e *MessageEvent) ExtractPlainText() string {
	return e.Message.ExtractPlainText()
}

func (e *MessageEvent) IsToMe() bool {
	return e.ToMe
}

type MessageEventSender struct {
	UserId   int64  `json:"user_id"`  // 消息发送者的QQ号
	Nickname string `json:"nickname"` // 消息发送者的昵称
	Sex      string `json:"sex"`      // 性别，male 或 female 或 unknown
	Age      int32  `json:"age"`
}

type GroupMessageEventSender struct {
	MessageEventSender
	Card  string `json:"card"`  // 群名片/备注
	Area  string `json:"area"`  // 地区
	Level string `json:"level"` // 成员等级
	Role  string `json:"role"`  // 角色，owner 或 admin 或 member
	Title string `json:"title"` // 专属头衔
}

type PrivateMessageEvent struct {
	MessageEvent
	Sender *MessageEventSender `json:"sender"` // 发送人信息
}

func (e *PrivateMessageEvent) GetEventDescription() string {
	msg := e.Message.String()
	msgRune := []rune(msg)
	if len(msgRune) > 100 {
		msg = fmt.Sprintf("%s...(省略%d个字符)...%s", string(msgRune[:50]), len(msgRune)-100, string(msgRune[len(msgRune)-50:]))
	}
	msg = strings.Replace(msg, "\n", "\\n", -1)
	return fmt.Sprintf("[私聊消息](#%d 来自%d): %v", e.MessageId, e.UserId, msg)
}

type Anonymous struct {
	Id   int64  `json:"id"`   // 匿名用户的ID
	Name string `json:"name"` // 匿名用户的名词
	Flag string `json:"flag"` // 匿名用户 flag，在调用禁言 API 时需要传入
}

type GroupMessageEvent struct {
	MessageEvent
	GroupId   int64                    `json:"group_id"` // 群号
	Sender    *GroupMessageEventSender `json:"sender"`   // 发送人信息
	Anonymous *Anonymous               `json:"anonymous"`
}

func (e *GroupMessageEvent) GetEventDescription() string {
	msg := e.Message.String()
	msgRune := []rune(msg)
	if len(msgRune) > 100 {
		msg = fmt.Sprintf("%s...(省略%d个字符)...%s", string(msgRune[:50]), len(msgRune)-100, string(msgRune[len(msgRune)-50:]))
	}
	msg = strings.Replace(msg, "\n", "\\n", -1)
	return fmt.Sprintf("[群聊消息](#%d 来自%d@群%d): %v", e.MessageId, e.UserId, e.GroupId, msg)
}

func (e *GroupMessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

type LifeCycleMetaEvent struct {
	Event
	MetaEventType string `json:"meta_event_type"` // 元事件类型，lifecycle
	SubType       string `json:"sub_type"`        // 元事件子类型，enable、disable、connect
}

func (e *LifeCycleMetaEvent) GetSecondType() string {
	return e.MetaEventType
}

func (e *LifeCycleMetaEvent) GetSubType() string {
	return e.SubType
}

type HeartbeatMetaEvent struct {
	Event
	MetaEventType string `json:"meta_event_type"` // 元事件类型，heartbeat
	Status        struct {
		Online bool `json:"online"` // 在线状态
		Good   bool `json:"good"`   // 同online
	} `json:"status"`
	Interval int64 `json:"interval"` // 元事件心跳间隔，单位ms
}

func (e *HeartbeatMetaEvent) GetSecondType() string {
	return e.MetaEventType
}

type NoticeEvent struct {
	Event
	NoticeType string `json:"notice_type"` // 通知类型，group, private
}

func (e *NoticeEvent) GetSecondType() string {
	return e.NoticeType
}

// 群文件上传通知
type GroupUploadNoticeEvent struct {
	NoticeEvent
	GroupId int64 `json:"group_id"` // 群号
	UserId  int64 `json:"user_id"`  // 上传者的QQ号
	File    struct {
		Id    string `json:"id"`     // 文件 ID
		Name  string `json:"name"`   // 文件名
		Size  int64  `json:"size"`   // 文件大小
		BusId string `json:"bus_id"` // 文件公众号 ID
	} `json:"file"`
}

func (e *GroupUploadNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

// 群管理员变动通知
type GroupAdminNoticeEvent struct {
	NoticeEvent
	SubType string `json:"sub_type"` // 通知子类型，set unset
	GroupId int64  `json:"group_id"` // 群号
	UserId  int64  `json:"user_id"`  // 管理员 QQ 号
}

func (e *GroupAdminNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *GroupAdminNoticeEvent) GetSubType() string {
	return e.SubType
}

// 群成员增加通知
type GroupIncreaseNoticeEvent struct {
	NoticeEvent
	SubType    string `json:"sub_type"`    // 通知子类型，approve, invite
	GroupId    int64  `json:"group_id"`    // 群号
	UserId     int64  `json:"user_id"`     // 新成员 QQ 号
	OperatorId int64  `json:"operator_id"` // 操作者 QQ 号
}

func (e *GroupIncreaseNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *GroupIncreaseNoticeEvent) GetSubType() string {
	return e.SubType
}

// 群成员减少通知
type GroupDecreaseNoticeEvent struct {
	NoticeEvent
	SubType    string `json:"sub_type"`    // 通知子类型，leave, kick, kick_me
	GroupId    int64  `json:"group_id"`    // 群号
	UserId     int64  `json:"user_id"`     // 离开者 QQ 号
	OperatorId int64  `json:"operator_id"` // 操作者 QQ 号
}

func (e *GroupDecreaseNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *GroupDecreaseNoticeEvent) GetSubType() string {
	return e.SubType
}

// 群禁言通知
type GroupBanNoticeEvent struct {
	NoticeEvent
	SubType    string `json:"sub_type"`    // 通知子类型，ban, lift_ban
	GroupId    int64  `json:"group_id"`    // 群号
	UserId     int64  `json:"user_id"`     // 被禁言 QQ 号
	OperatorId int64  `json:"operator_id"` // 操作者 QQ 号
	Duration   int64  `json:"duration"`    // 禁言时长，单位秒
}

func (e *GroupBanNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *GroupBanNoticeEvent) GetSubType() string {
	return e.SubType
}

// 好友添加通知
type FriendAddNoticeEvent struct {
	NoticeEvent
	UserId int64 `json:"user_id"` // 好友 QQ 号
}

func (e *FriendAddNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

// 群消息撤回通知
type GroupRecallNoticeEvent struct {
	NoticeEvent
	GroupId    int64 `json:"group_id"`    // 群号
	UserId     int64 `json:"user_id"`     // 撤回者 QQ 号
	OperatorId int64 `json:"operator_id"` // 操作者 QQ 号
	MessageId  int64 `json:"message_id"`  // 消息 ID
}

func (e *GroupRecallNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

// 好友消息撤回通知
type FriendRecallNoticeEvent struct {
	NoticeEvent
	UserId    int64 `json:"user_id"`    // 撤回者 QQ 号
	MessageId int64 `json:"message_id"` // 消息 ID
}

func (e *FriendRecallNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

// 戳一戳通知
type PokeNoticeEvent struct {
	NoticeEvent
	SubType  string `json:"sub_type"`  // 通知子类型，poke
	GroupId  int64  `json:"group_id"`  // 群号
	UserId   int64  `json:"user_id"`   // 发送戳一戳的 QQ 号
	TargetId int64  `json:"target_id"` // 被戳一戳的 QQ 号
}

func (e *PokeNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *PokeNoticeEvent) GetSubType() string {
	return e.SubType
}

// 运气王通知
type LuckyKingNoticeEvent struct {
	NoticeEvent
	SubType  string `json:"sub_type"`  // 通知子类型，lucky_king
	GroupId  int64  `json:"group_id"`  // 群号
	UserId   int64  `json:"user_id"`   // 发红包者的 QQ 号
	TargetId int64  `json:"target_id"` // 运气王的 QQ 号
}

func (e *LuckyKingNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *LuckyKingNoticeEvent) GetSubType() string {
	return e.SubType
}

// 群成员荣誉变更
type HonorNoticeEvent struct {
	NoticeEvent
	SubType   string `json:"sub_type"`   // 通知子类型，honor
	GroupId   int64  `json:"group_id"`   // 群号
	UserId    int64  `json:"user_id"`    // QQ 号
	HonorType string `json:"honor_type"` // 荣誉类型，talkative、performer、emotion，分别表示龙王、群聊之火、快乐源泉
}

func (e *HonorNoticeEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *HonorNoticeEvent) GetSubType() string {
	return e.SubType
}

// 群成员名片更新
type GroupCardNoticeEvent struct {
	NoticeEvent
	GroupId int64  `json:"group_id"` // 群号
	UserId  int64  `json:"user_id"`  // 群成员 QQ 号
	CardNew string `json:"card_new"` // 新名片
	CardOld string `json:"card_old"` // 旧名片
}

// 接收到离线文件
type OfflineFileNoticeEvent struct {
	NoticeEvent
	UserId int64 `json:"user_id"` // 用户 ID
	File   struct {
		Name string `json:"name"` // 文件名
		Size int64  `json:"size"` // 文件大小
		Url  string `json:"url"`  // 下载链接
	} `json:"file"`
}

// 其他客户端在线状态变更
type ClientStatusNoticeEvent struct {
	NoticeEvent
	Client struct {
		AppId      int64  `json:"app_id"`      // 客户端 ID
		DeviceName string `json:"device_name"` // 设备名称
		DeviceKind string `json:"device_kind"` // 设备类型
	} `json:"client"`
	Online bool `json:"online"` // 在线状态
}

// 精华消息
type EssenceNoticeEvent struct {
	NoticeEvent
	SubType    string `json:"sub_type"` // 通知子类型，essence
	SenderId   int64  `json:"sender_id"`
	OperatorId int64  `json:"operator_id"`
	MessageId  int32  `json:"message_id"`
}

// ==========================
// 请求事件
// ==========================

// 加好友请求事件
type FriendRequestEvent struct {
	Event
	RequestType string `json:"request_type"` // 请求类型，friend
	UserId      int64  `json:"user_id"`      // 发送请求的QQ号
	Comment     string `json:"comment"`      // 验证消息
	Flag        string `json:"flag"`         // 请求 flag，在调用处理请求的 API 时需要传入
}

func (e *FriendRequestEvent) GetSecondType() string {
	return e.RequestType
}

func (e *FriendRequestEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

// 加群请求事件
type GroupRequestEvent struct {
	Event
	RequestType string `json:"request_type"` // 请求类型，group
	SubType     string `json:"sub_type"`     // 请求子类型，add、invite，分别表示加群请求、邀请登录号入群
	GroupId     int64  `json:"group_id"`     // 群号
	UserId      int64  `json:"user_id"`      // 发送请求的QQ号
	Comment     string `json:"comment"`      // 验证消息
	Flag        string `json:"flag"`         // 请求请求 flag，在调用处理请求的 API 时需要传入标识
}

func (e *GroupRequestEvent) GetSecondType() string {
	return e.RequestType
}

func (e *GroupRequestEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}

func (e *GroupRequestEvent) GetSubType() string {
	return e.SubType
}
