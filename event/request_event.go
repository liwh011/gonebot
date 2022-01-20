package event

// 加好友请求事件
type FriendRequestEvent struct {
	Event
	RequestType string // 请求类型，friend
	UserId      int64  // 发送请求的QQ号
	Comment     string // 验证消息
	Flag        int64  // 请求 flag，在调用处理请求的 API 时需要传入
}

// 加群请求事件
type GroupRequestEvent struct {
	Event
	RequestType string // 请求类型，group
	SubType     string // 请求子类型，add、invite，分别表示加群请求、邀请登录号入群
	GroupId     int64  // 群号
	UserId      int64  // 发送请求的QQ号
	Comment     string // 验证消息
	Flag        int64  // 请求请求 flag，在调用处理请求的 API 时需要传入标识
}
