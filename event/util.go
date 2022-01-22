package event

import "reflect"

func GetEventField(e *T_Event, field string) interface{} {
	return reflect.ValueOf(e).Elem().Elem().FieldByName(field).Interface()

}

func SetEventField(e *T_Event, field string, value interface{}) {
	reflect.ValueOf(e).Elem().Elem().FieldByName(field).Set(reflect.ValueOf(value))
}
