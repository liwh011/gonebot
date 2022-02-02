package event

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

const (
	POST_TYPE_META    = "meta_event"
	POST_TYPE_MESSAGE = "message"
	POST_TYPE_NOTICE  = "notice"
	POST_TYPE_REQUEST = "request"
)

type EventName string

const (
	EVENT_NAME_MESSAGE           EventName = "message"
	EVENT_NAME_PRIVATE_MESSAGE   EventName = "message.private"
	EVENT_NAME_GROUP_MESSAGE     EventName = "message.group"
	EVENT_NAME_NOTICE            EventName = "notice"
	EVENT_NAME_GROUP_UPLOAD      EventName = "notice.group_upload"
	EVENT_NAME_GROUP_ADMIN       EventName = "notice.group_admin"
	EVENT_NAME_GROUP_DECREASE    EventName = "notice.group_decrease"
	EVENT_NAME_GROUP_INCREASE    EventName = "notice.group_increase"
	EVENT_NAME_GROUP_BAN         EventName = "notice.group_ban"
	EVENT_NAME_FRIEND_ADD        EventName = "notice.friend_add"
	EVENT_NAME_GROUP_RECALL      EventName = "notice.group_recall"
	EVENT_NAME_FRIEND_RECALL     EventName = "notice.friend_recall"
	EVENT_NAME_NOTIFY            EventName = "notice.notify"
	EVENT_NAME_NOTIFY_POKE       EventName = "notice.notify.poke"
	EVENT_NAME_NOTIFY_LUCKY_KING EventName = "notice.notify.lucky_king"
	EVENT_NAME_NOTIFY_HONOR      EventName = "notice.notify.honor"
	EVENT_NAME_REQUEST           EventName = "request"
	EVENT_NAME_REQUEST_FRIEND    EventName = "request.friend"
	EVENT_NAME_REQUEST_GROUP     EventName = "request.group"
	EVENT_NAME_META              EventName = "meta_event"
	EVENT_NAME_META_LIFECYCLE    EventName = "meta_event.lifecycle"
	EVENT_NAME_META_HEARTBEAT    EventName = "meta_event.heartbeat"
)

type I_Event interface {
	GetPostType() string
	GetEventName() string
	GetEventDescription() string

	IsMessageEvent() bool
	IsToMe() bool
}

type Event struct {
	Time     int64  `json:"time"`      // 事件发生的时间戳
	SelfId   int64  `json:"self_id"`   // 收到事件的机器人的QQ号
	PostType string `json:"post_type"` // 事件的类型，message, notice, request, meta_event

	EventName string `json:"-"` // 事件的名称，形如：notice.group.set
	ToMe      bool   `json:"-"` // 是否与我（bot）有关（即私聊我、或群聊At我、我被踢了、等等）
}

// 获取事件的上报类型，有message, notice, request, meta_event
func (e *Event) GetPostType() string {
	return e.PostType
}

// 获取事件的名称，形如：notice.group.set
func (e *Event) GetEventName() string {
	return e.EventName
}

// 获取事件的描述，一般用于日志输出
func (e *Event) GetEventDescription() string {
	return fmt.Sprintf("[%s]: %+v", e.EventName, *e)
}

func (e *Event) IsMessageEvent() bool {
	return e.PostType == POST_TYPE_MESSAGE
}

func (e *Event) IsToMe() bool {
	return e.ToMe
}

// 从JSON对象中生成Event对象（指针）
func FromJsonObject(obj gjson.Result) I_Event {
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
	switch typeName {
	case "message.private":
		ev = &PrivateMessageEvent{}
	case "message.group":
		ev = &GroupMessageEvent{}
	case "notice.group_upload":
		ev = &GroupUploadNoticeEvent{}
	case "notice.group_admin":
		ev = &GroupAdminNoticeEvent{}
	case "notice.group_decrease":
		ev = &GroupDecreaseNoticeEvent{}
	case "notice.group_increase":
		ev = &GroupIncreaseNoticeEvent{}
	case "notice.group_ban":
		ev = &GroupBanNoticeEvent{}
	case "notice.friend_add":
		ev = &FriendAddNoticeEvent{}
	case "notice.group_recall":
		ev = &GroupRecallNoticeEvent{}
	case "notice.friend_recall":
		ev = &FriendRecallNoticeEvent{}
	case "notice.notify":
		switch subType {
		case "poke":
			ev = &PokeNoticeEvent{}
		case "lucky_king":
			ev = &LuckyKingNoticeEvent{}
		case "honor":
			ev = &HonorNoticeEvent{}
		}
	case "request.friend":
		ev = &FriendRequestEvent{}
	case "request.group":
		ev = &GroupRequestEvent{}
	case "meta_event.lifecycle":
		ev = &LifeCycleMetaEvent{}
	case "meta_event.heartbeat":
		ev = &HeartbeatMetaEvent{}
	default:
		panic(fmt.Sprintf("unknown event type: %s", typeName))
	}

	// 借助json库将JSON对象中的字段赋值给Event对象，懒得自个写反射了
	err := json.Unmarshal([]byte(obj.Raw), ev)
	if err != nil {
		panic(err)
	}

	// 设置事件的名称
	SetEventField(ev, "EventName", fullTypeName)
	if isToMe(ev) {
		SetEventField(ev, "ToMe", true)
	}

	return ev
}
