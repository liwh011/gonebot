package gonebot

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
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

func mapToStruct(m map[string]interface{}, s interface{}) {
	sValue := reflect.ValueOf(s).Elem()
	sType := sValue.Type()

	for i := 0; i < sValue.NumField(); i++ {
		f := sValue.Field(i)
		if !f.CanSet() {
			continue
		}

		fName := sType.Field(i).Name
		names := []string{
			strings.Split(sType.Field(i).Tag.Get("json"), ",")[0], // json tag
			strings.Split(sType.Field(i).Tag.Get("yaml"), ",")[0], // yaml tag
			camelCaseToSnakeCase(fName),                           // snake_case
			strings.ToLower(fName[:1]) + fName[1:],                // camelCase
			fName,                                                 // original (CamelCase)
		}
		for _, name := range names {
			if name == "" {
				continue
			}

			if v, ok := m[name]; ok {
				typeConvert(v, f)
				break
			}
		}
	}

}

func camelCaseToSnakeCase(s string) string {
	var res []rune
	for i, r := range s {
		if i == 0 {
			res = append(res, unicode.ToLower(r))
		} else if unicode.IsUpper(r) {
			res = append(res, '_')
			res = append(res, unicode.ToLower(r))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

func typeConvert(v interface{}, f reflect.Value) {
	switch f.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Float32, reflect.Float64:
		f.Set(reflect.ValueOf(v).Convert(f.Type()))

	case reflect.Struct:
		mapToStruct(v.(map[string]interface{}), f.Addr().Interface())

	case reflect.Ptr:
		elemType := f.Type().Elem() // 获取指针指向的元素类型
		nv := reflect.New(elemType) // 创建类型为elemType的零值
		f.Set(nv)                   // 设置指针指向nv
		typeConvert(v, f.Elem())

	case reflect.Slice:
		newSlice := reflect.MakeSlice(f.Type(), 0, 0)
		vv := reflect.ValueOf(v)
		for i := 0; i < vv.Len(); i++ {
			nv := reflect.New(f.Type().Elem())
			typeConvert(vv.Index(i).Interface(), nv.Elem())
			newSlice = reflect.Append(newSlice, nv.Elem())
		}
		f.Set(newSlice)

	case reflect.Map:
		newMap := reflect.MakeMap(f.Type())
		vv := reflect.ValueOf(v)
		for _, k := range vv.MapKeys() {
			nv := reflect.New(f.Type().Elem())
			typeConvert(vv.MapIndex(k).Interface(), nv.Elem())
			newMap.SetMapIndex(k, nv.Elem())
		}
		f.Set(newMap)

	default:
		f.Set(reflect.ValueOf(v))
	}

}

// 创建相同结构的新对象，返回指针
func createUnderlyingStruct(args interface{}) interface{} {
	v := reflect.ValueOf(args)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("args must be a struct")
	}

	// 先转成interface{}才能用Addr()
	newValue := reflect.New(v.Type()).Interface()
	return reflect.ValueOf(newValue).Elem().Addr().Interface()
}
