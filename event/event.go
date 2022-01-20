package event

type Event struct {
	Time     int64  // 事件发生的时间戳
	SelfId   int64  // 收到事件的机器人的QQ号
	PostType string // 事件的类型，message, notice, request, meta_event
}

