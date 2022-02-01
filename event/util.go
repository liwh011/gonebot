package event

import (
	"reflect"
)

// // 获取事件的名称，形如：notice.group.set
// func GetEventName(event I_Event) string {
// 	// ev := reflect.ValueOf(event)
// 	// type1 := ev.FieldByName("PostType").String()
// 	// type2 := ""
// 	// switch type1 {
// 	// case "message":
// 	// 	type2 = ev.FieldByName("MessageType").String()
// 	// case "notice":
// 	// 	type2 = ev.FieldByName("NoticeType").String()
// 	// case "request":
// 	// 	type2 = ev.FieldByName("RequestType").String()
// 	// case "meta_event":
// 	// 	type2 = ev.FieldByName("MetaEventType").String()
// 	// default: // 理论上不会出现这个情况
// 	// 	panic(fmt.Sprintf("unknown post type: %s", type1))
// 	// }
// 	// type3 := ""
// 	// if ev.FieldByName("SubType").IsValid() {
// 	// 	type3 = ev.FieldByName("SubType").String()
// 	// 	return fmt.Sprintf("%s.%s.%s", type1, type2, type3)
// 	// }

// 	// return fmt.Sprintf("%s.%s", type1, type2)

// 	return GetEventField(event, "EventName").(string)
// }

// // 获取事件的描述，一般用于日志输出
// func GetEventDescription(event *T_Event) string {
// 	switch event := (*event).(type) {
// 	case PrivateMessageEvent:
// 		return fmt.Sprintf("[私聊消息](#%d 来自%d %v): ", event.MessageId, event.UserId, event.Message)
// 	case GroupMessageEvent:
// 		return fmt.Sprintf("[群聊消息](#%d 来自%d@群%d %v): ", event.MessageId, event.UserId, event.GroupId, event.Message)
// 	}
// 	return fmt.Sprintf("[%s]: %+v", GetEventName(event), *event)
// }

func GetEventField(e I_Event, field string) (interface{}, bool) {
	f := reflect.ValueOf(e).Elem().FieldByName(field)
	if f.IsValid() {
		return f.Interface(), true
	}
	return nil, false

}

func SetEventField(e I_Event, field string, value interface{}) {
	reflect.ValueOf(e).Elem().FieldByName(field).Set(reflect.ValueOf(value))
}

// func IsMessageEvent(e I_Event) bool {
// 	return GetEventField(e, "PostType") == "message"
// }
