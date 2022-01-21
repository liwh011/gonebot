package event

type LifeCycleMetaEvent struct {
	Event
	MetaEventType string `json:"meta_event_type"` // 元事件类型，lifecycle
	SubType       string `json:"sub_type"`        // 元事件子类型，enable、disable、connect
}

type HeartbeatMetaEvent struct {
	Event
	MetaEventType string `json:"meta_event_type"` // 元事件类型，heartbeat
	Status        struct {
		Online bool `json:"online"` // 在线状态
		Good   bool `json:"good"`   // 同online
	} `json:"status"`
	Interval int64 `json:"interval"` // 元事件心跳间隔，单位ms
}
