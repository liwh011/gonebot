package gonebot

import (
	"reflect"
	"strconv"
	"strings"
)

func boolToInt01(b bool) int {
	if b {
		return 1
	}
	return 0
}

// 转义CQ码
func Escape(s string, escapeComma bool) (res string) {
	res = strings.Replace(s, "&", "&amp;", -1)
	res = strings.Replace(res, "[", "&#91;", -1)
	res = strings.Replace(res, "]", "&#93;", -1)
	if escapeComma {
		res = strings.Replace(res, ",", "&#44;", -1)
	}
	return res
}

// 反转义CQ码
func Unescape(s string) (res string) {
	res = strings.Replace(s, "&#44;", ",", -1)
	res = strings.Replace(res, "&#91;", "[", -1)
	res = strings.Replace(res, "&#93;", "]", -1)
	res = strings.Replace(res, "&amp;", "&", -1)
	return res
}

// 获取事件的某个字段的值，如果不存在，则返回false
func getEventField(e I_Event, field string) (interface{}, bool) {
	f := reflect.ValueOf(e).Elem().FieldByName(field)
	if f.IsValid() {
		return f.Interface(), true
	}
	return nil, false

}

func setEventField(e I_Event, field string, value interface{}) {
	reflect.ValueOf(e).Elem().FieldByName(field).Set(reflect.ValueOf(value))
}

// 判断事件是否与bot相关
func isEventRelativeToBot(event I_Event) bool {
	selfId, _ := getEventField(event, "SelfId")
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
		if targetId, exist := getEventField(event, "TargetId"); exist {
			return targetId.(int64) == myId
		}
		if userId, exist := getEventField(event, "UserId"); exist {
			return userId.(int64) == myId
		}
	}
	return false
}
