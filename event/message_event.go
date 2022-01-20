package event

import (
	"fmt"

	"github.com/liwh011/gonebot/message"
)

type MessageEvent struct {
	Event
	MessageType string          // 消息类型，group, private
	SubType     string          // 消息子类型，friend, group, other
	MessageId   int32           // 消息ID
	UserId      int64           // 消息发送者的QQ号
	Message     message.Message // 消息内容
	RawMessage  string          // 原始消息内容
	Font        int32           // 字体

	toMe bool // 是否与我（bot）有关（即私聊我、或群聊At我）
}

func (m MessageEvent) IsToMe() bool {
	return m.toMe
}

type MessageEventSender struct {
	UserId   int64  // 消息发送者的QQ号
	Nickname string // 消息发送者的昵称
	Sex      string // 性别，male 或 female 或 unknown
	Age      int32
}

type GroupMessageEventSender struct {
	MessageEventSender
	Card  string // 群名片/备注
	Area  string // 地区
	Level string // 成员等级
	Role  string // 角色，owner 或 admin 或 member
	Title string // 专属头衔
}

type PrivateMessageEvent struct {
	MessageEvent
	Sender *MessageEventSender // 发送人信息
}

func (e PrivateMessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

type GroupMessageEvent struct {
	MessageEvent
	GroupId   int64                    // 群号
	Sender    *GroupMessageEventSender // 发送人信息
	Anonymous *struct {
		Id   int64  // 匿名用户的ID
		Name string // 匿名用户的名词
		Flag string // 匿名用户 flag，在调用禁言 API 时需要传入
	}
}

func (e GroupMessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}
