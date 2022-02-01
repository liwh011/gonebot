package event

// 加好友请求事件
type FriendRequestEvent struct {
	Event
	RequestType string `json:"request_type"` // 请求类型，friend
	UserId      int64  `json:"user_id"`      // 发送请求的QQ号
	Comment     string `json:"comment"`      // 验证消息
	Flag        string `json:"flag"`         // 请求 flag，在调用处理请求的 API 时需要传入
}

// 加群请求事件
type GroupRequestEvent struct {
	Event
	RequestType string `json:"request_type"` // 请求类型，group
	SubType     string `json:"sub_type"`     // 请求子类型，add、invite，分别表示加群请求、邀请登录号入群
	GroupId     int64  `json:"group_id"`     // 群号
	UserId      int64  `json:"user_id"`      // 发送请求的QQ号
	Comment     string `json:"comment"`      // 验证消息
	Flag        string `json:"flag"`         // 请求请求 flag，在调用处理请求的 API 时需要传入标识
}
