package event

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tidwall/gjson"
)

type Event struct {
	Time     int64  `json:"time"`      // 事件发生的时间戳
	SelfId   int64  `json:"self_id"`   // 收到事件的机器人的QQ号
	PostType string `json:"post_type"` // 事件的类型，message, notice, request, meta_event

	EventName string `json:"-"` // 事件的名称，形如：notice.group.set
}

type T_Event interface{}

// 获取事件的名称，形如：notice.group.set
func GetEventName(event *T_Event) string {
	// ev := reflect.ValueOf(event)
	// type1 := ev.FieldByName("PostType").String()
	// type2 := ""
	// switch type1 {
	// case "message":
	// 	type2 = ev.FieldByName("MessageType").String()
	// case "notice":
	// 	type2 = ev.FieldByName("NoticeType").String()
	// case "request":
	// 	type2 = ev.FieldByName("RequestType").String()
	// case "meta_event":
	// 	type2 = ev.FieldByName("MetaEventType").String()
	// default: // 理论上不会出现这个情况
	// 	panic(fmt.Sprintf("unknown post type: %s", type1))
	// }
	// type3 := ""
	// if ev.FieldByName("SubType").IsValid() {
	// 	type3 = ev.FieldByName("SubType").String()
	// 	return fmt.Sprintf("%s.%s.%s", type1, type2, type3)
	// }

	// return fmt.Sprintf("%s.%s", type1, type2)

	return GetEventField(event, "EventName").(string)
}

// 获取事件的描述，一般用于日志输出
func GetEventDescription(event *T_Event) string {
	switch event := (*event).(type) {
	case PrivateMessageEvent:
		return fmt.Sprintf("[私聊消息](#%d 来自%d %v): ", event.MessageId, event.UserId, event.Message)
	case GroupMessageEvent:
		return fmt.Sprintf("[群聊消息](#%d 来自%d@群%d %v): ", event.MessageId, event.UserId, event.GroupId, event.Message)
	}
	return fmt.Sprintf("[%s]: %+v", GetEventName(event), *event)
}

func FromJsonObject(obj gjson.Result) *T_Event {
	postType := obj.Get("post_type").String()
	nextType := obj.Get(postType + "_type").String()
	typeName := fmt.Sprintf("%s.%s", postType, nextType)

	subType := ""
	fullTypeName := typeName
	if obj.Get("sub_type").Exists() {
		subType = obj.Get("sub_type").String()
		fullTypeName = fmt.Sprintf("%s.%s", typeName, subType)
	}

	var ev interface{}
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

	err := json.Unmarshal([]byte(obj.Raw), ev)
	if err != nil {
		panic(err)
	}

	reflect.ValueOf(ev).Elem().FieldByName("EventName").SetString(fullTypeName)
	var v T_Event = reflect.ValueOf(ev).Elem().Interface()

	return &v
}

// type EventDispatcher struct {
// 	handlers map[string][]func(interface{})
// }

// func NewEventDispatcher() *EventDispatcher {
// 	return &EventDispatcher{
// 		handlers: make(map[string][]func(interface{})),
// 	}
// }

// func (e *EventDispatcher) Register(eventName string, handler func(interface{})) {
// 	if _, ok := e.handlers[eventName]; !ok {
// 		e.handlers[eventName] = make([]func(interface{}), 0)
// 	}
// 	e.handlers[eventName] = append(e.handlers[eventName], handler)
// }

// func (e *EventDispatcher) Dispatch(event interface{}) {
// 	name := GetEventName(&event)
// 	if _, ok := e.handlers[name]; ok {
// 		for _, handler := range e.handlers[name] {
// 			handler(event)
// 		}
// 	}
// }
