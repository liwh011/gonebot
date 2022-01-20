package event

type LifeCycleMetaEvent struct {
	Event
	MetaEventType string // 元事件类型，lifecycle
	SubType       string // 元事件子类型，enable、disable、connect
}

type HeartbeatMetaEvent struct {
	Event
	MetaEventType string // 元事件类型，heartbeat
	Status        struct {
		Online bool // 在线状态
		Good   bool // 同online
	}
	Interval int64 // 元事件心跳间隔，单位ms
}
