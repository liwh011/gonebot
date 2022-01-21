package event

import (
	"fmt"

	"github.com/liwh011/gonebot/message"
)

type MessageEvent struct {
	Event
	MessageType string          `json:"message_type"` // 消息类型，group, private
	SubType     string          `json:"sub_type"`     // 消息子类型，friend, group, other
	MessageId   int32           `json:"message_id"`   // 消息ID
	UserId      int64           `json:"user_id"`      // 消息发送者的QQ号
	Message     message.Message `json:"message"`      // 消息内容
	RawMessage  string          `json:"raw_message"`  // 原始消息内容
	Font        int32           `json:"font"`         // 字体

	toMe bool `json:"-"` // 是否与我（bot）有关（即私聊我、或群聊At我）
}

func (m MessageEvent) IsToMe() bool {
	return m.toMe
}

type MessageEventSender struct {
	UserId   int64  `json:"user_id"`  // 消息发送者的QQ号
	Nickname string `json:"nickname"` // 消息发送者的昵称
	Sex      string `json:"sex"`      // 性别，male 或 female 或 unknown
	Age      int32  `json:"age"`
}

type GroupMessageEventSender struct {
	MessageEventSender
	Card  string `json:"card"`  // 群名片/备注
	Area  string `json:"area"`  // 地区
	Level string `json:"level"` // 成员等级
	Role  string `json:"role"`  // 角色，owner 或 admin 或 member
	Title string `json:"title"` // 专属头衔
}

type PrivateMessageEvent struct {
	MessageEvent
	Sender *MessageEventSender // 发送人信息
}

func (e PrivateMessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

type Anonymous struct {
	Id   int64  `json:"id"`   // 匿名用户的ID
	Name string `json:"name"` // 匿名用户的名词
	Flag string `json:"flag"` // 匿名用户 flag，在调用禁言 API 时需要传入
}

type GroupMessageEvent struct {
	MessageEvent
	GroupId   int64                    `json:"group_id"` // 群号
	Sender    *GroupMessageEventSender `json:"sender"`   // 发送人信息
	Anonymous *Anonymous               `json:"anonymous"`
}

func (e GroupMessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}
