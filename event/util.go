package event

import (
	"reflect"
	"strconv"
)

// 获取事件的某个字段的值，如果不存在，则返回false
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

// 判断事件是否与bot相关
func isToMe(event I_Event) bool {
	selfId, _ := GetEventField(event, "SelfId")
	myId := selfId.(int64)

	switch event := event.(type) {
	// 私聊消息一定是发给bot的
	case *PrivateMessageEvent:
		return true

	// 群聊消息，如果At了bot，则是
	case *GroupMessageEvent:
		myIdStr := strconv.FormatInt(myId, 10)
		atSegs := event.Message.FilterByType("at")
		for _, seg := range atSegs {
			if seg.Data["qq"].(string) == myIdStr {
				return true
			}
		}
		return false

	// 事件中有targetid字段的，则用targetid判断，否则用userid判断
	default:
		if targetId, exist := GetEventField(event, "TargetId"); exist {
			return targetId.(int64) == myId
		}
		if userId, exist := GetEventField(event, "UserId"); exist {
			return userId.(int64) == myId
		}
	}
	return false
}
