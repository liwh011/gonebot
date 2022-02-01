package event

import (
	// "fmt"

	"fmt"
	"strings"

	"github.com/liwh011/gonebot/message"
)

type I_MessageEvent interface {
	GetPostType() string
	GetEventName() string
	GetEventDescription() string

	GetMessageType() string
	GetSessionId() string
	GetMessage() *message.Message
	ExtractPlainText() string
	IsToMe() bool
}

type MessageEvent struct {
	Event
	MessageType string          `json:"message_type"` // 消息类型，group, private
	SubType     string          `json:"sub_type"`     // 消息子类型，friend, group, other
	MessageId   int32           `json:"message_id"`   // 消息ID
	UserId      int64           `json:"user_id"`      // 消息发送者的QQ号
	Message     message.Message `json:"message"`      // 消息内容
	RawMessage  string          `json:"raw_message"`  // 原始消息内容
	Font        int32           `json:"font"`         // 字体

	ToMe bool `json:"-"` // 是否与我（bot）有关（即私聊我、或群聊At我）
}

func (e *MessageEvent) GetMessageType() string {
	return e.MessageType
}

func (e *MessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d", e.UserId)
}

func (e *MessageEvent) GetMessage() *message.Message {
	return &e.Message
}

func (e *MessageEvent) ExtractPlainText() string {
	return e.Message.ExtractPlainText()
}

func (e *MessageEvent) IsToMe() bool {
	return e.ToMe
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

func (e *PrivateMessageEvent) GetEventDescription() string {
	msg := e.Message.String()
	msgRune := []rune(msg)
	if len(msgRune) > 100 {
		msg = fmt.Sprintf("%s...(省略%d个字符)...%s", string(msgRune[:50]), len(msgRune)-100, string(msgRune[len(msgRune)-50:]))
	}
	msg = strings.Replace(msg, "\n", "\\n", -1)
	return fmt.Sprintf("[私聊消息](#%d 来自%d): %v", e.MessageId, e.UserId, msg)
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

func (e *GroupMessageEvent) GetEventDescription() string {
	msg := e.Message.String()
	msgRune := []rune(msg)
	if len(msgRune) > 100 {
		msg = fmt.Sprintf("%s...(省略%d个字符)...%s", string(msgRune[:50]), len(msgRune)-100, string(msgRune[len(msgRune)-50:]))
	}
	msg = strings.Replace(msg, "\n", "\\n", -1)
	return fmt.Sprintf("[群聊消息](#%d 来自%d@群%d): %v", e.MessageId, e.UserId, e.GroupId, msg)
}

func (e *GroupMessageEvent) GetSessionId() string {
	return fmt.Sprintf("%d@%d", e.UserId, e.GroupId)
}
