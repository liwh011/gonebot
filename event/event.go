package event

import (
	"fmt"
	"reflect"
)

type Event struct {
	Time     int64  // 事件发生的时间戳
	SelfId   int64  // 收到事件的机器人的QQ号
	PostType string // 事件的类型，message, notice, request, meta_event
}

// 获取事件的名称，形如：notice.group.set
func GetEventName(event interface{}) string {
	ev := reflect.ValueOf(event)
	type1 := ev.FieldByName("PostType").String()
	type2 := ""
	switch type1 {
	case "message":
		type2 = ev.FieldByName("MessageType").String()
	case "notice":
		type2 = ev.FieldByName("NoticeType").String()
	case "request":
		type2 = ev.FieldByName("RequestType").String()
	case "meta_event":
		type2 = ev.FieldByName("MetaEventType").String()
	default: // 理论上不会出现这个情况
		panic(fmt.Sprintf("unknown post type: %s", type1))
	}
	type3 := ""
	if ev.FieldByName("SubType").IsValid() {
		type3 = ev.FieldByName("SubType").String()
		return fmt.Sprintf("%s.%s.%s", type1, type2, type3)
	}

	return fmt.Sprintf("%s.%s", type1, type2)
}

// 获取事件的描述，一般用于日志输出
func GetEventDescription(event interface{}) string {
	switch event := event.(type) {
	case PrivateMessageEvent:
		return fmt.Sprintf("消息#%d 来自%d %v", event.MessageId, event.UserId, event.Message)
	case GroupMessageEvent:
		return fmt.Sprintf("消息#%d 来自%d@群%d %v", event.MessageId, event.UserId, event.GroupId, event.Message)
	}
	return fmt.Sprintf("%v", event)
}
