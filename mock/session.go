package mock

import (
	"fmt"
	"time"

	"github.com/liwh011/gonebot"
)

// 私聊会话
type PrivateSession struct {
	server   *MockServer
	isFriend bool

	UserId   int64
	Nickname string
	Sex      string
	Age      int32
	BotId    int64
}

func (s *PrivateSession) getMsgId() int32 {
	return s.server.getMsgId()
}

// 获取聊天记录
func (s *PrivateSession) GetMessageHistory() MessageHistory {
	return s.server.messageHistory.getPrivateHistory(s.UserId)
}

// 模拟一个私聊消息事件
func (s *PrivateSession) MessageEvent(msg gonebot.Message) gonebot.PrivateMessageEvent {
	ev := gonebot.PrivateMessageEvent{
		MessageEvent: gonebot.MessageEvent{
			Event: gonebot.Event{
				Time:      time.Now().Unix(),
				SelfId:    s.BotId,
				PostType:  gonebot.PostType_MessageEvent,
				EventName: "message.private.friend",
				ToMe:      true,
			},
			MessageType: "private",
			SubType:     "friend",
			MessageId:   s.getMsgId(),
			UserId:      s.UserId,
			Message:     msg,
			RawMessage:  msg.String(),
			Font:        0,
		},
		Sender: &gonebot.MessageEventSender{
			UserId:   s.UserId,
			Nickname: s.Nickname,
			Sex:      s.Sex,
			Age:      s.Age,
		},
	}
	if !s.isFriend {
		ev.SubType = "other"
	}
	s.server.sendEvent(&ev)
	return ev
}

// 模拟一个私聊消息事件
func (s *PrivateSession) MessageEventByText(txt string) gonebot.PrivateMessageEvent {
	return s.MessageEvent(gonebot.MsgPrint(txt))
}

// 模拟撤回消息事件
func (s *PrivateSession) RecallEvent(msgId int32) gonebot.FriendRecallNoticeEvent {
	ev := gonebot.FriendRecallNoticeEvent{
		NoticeEvent: gonebot.NoticeEvent{
			Event: gonebot.Event{
				Time:      time.Now().Unix(),
				SelfId:    s.BotId,
				PostType:  gonebot.PostType_NoticeEvent,
				EventName: "notice.friend_recall",
				ToMe:      false,
			},
			NoticeType: "friend_recall",
		},
		UserId:    s.UserId,
		MessageId: int64(msgId),
	}
	s.server.sendEvent(&ev)
	return ev
}

// 模拟戳一戳事件
func (s *PrivateSession) PokeEvent() gonebot.PokeNoticeEvent {
	ev := gonebot.PokeNoticeEvent{
		NoticeEvent: gonebot.NoticeEvent{
			Event: gonebot.Event{
				Time:      time.Now().Unix(),
				SelfId:    s.BotId,
				PostType:  gonebot.PostType_NoticeEvent,
				EventName: "notice.notify.poke",
				ToMe:      true,
			},
			NoticeType: "notify",
		},
		SubType:  "poke",
		GroupId:  0,
		UserId:   s.UserId,
		TargetId: s.BotId,
	}
	s.server.sendEvent(&ev)
	return ev
}

// 群聊会话
type GroupSession struct {
	server *MockServer
	BotId  int64 // 机器人QQ号

	GroupId   int64  // 群号
	GroupName string // 群名
	group     Group
}

func (s *GroupSession) getMsgId() int32 {
	return s.server.getMsgId()
}

// 获取聊天记录
func (s *GroupSession) GetMessageHistory() MessageHistory {
	return s.server.messageHistory.getGroupHistory(s.GroupId)
}

// 模拟一个群聊消息事件。当userId不存在时，会当作普通成员发送消息。
func (s *GroupSession) MessageEvent(userId int64, msg gonebot.Message) gonebot.GroupMessageEvent {
	member := s.group.GetMember(userId)
	if member == nil {
		member = &GroupMember{
			UserId:   userId,
			Nickname: fmt.Sprintf("未知群员%d", userId),
			Role:     "member",
		}
	}

	ev := gonebot.GroupMessageEvent{
		MessageEvent: gonebot.MessageEvent{
			Event: gonebot.Event{
				Time:      time.Now().Unix(),
				SelfId:    s.BotId,
				PostType:  gonebot.PostType_MessageEvent,
				EventName: "message.group.normal",
				ToMe:      false,
			},
			MessageType: "group",
			SubType:     "normal",
			MessageId:   s.getMsgId(),
			UserId:      member.UserId,
			Message:     msg,
			RawMessage:  msg.String(),
			Font:        0,
		},
		GroupId: s.GroupId,
		Sender: &gonebot.GroupMessageEventSender{
			MessageEventSender: gonebot.MessageEventSender{
				UserId:   member.UserId,
				Nickname: member.Nickname,
				Sex:      member.Sex,
				Age:      member.Age,
			},
			Card:  member.Card,
			Area:  member.Area,
			Level: member.Level,
			Role:  member.Role,
			Title: member.Title,
		},
		Anonymous: nil,
	}
	s.server.sendEvent(&ev)
	return ev
}

// 模拟一个群聊匿名消息事件
func (s *GroupSession) AnonymousMessageEvent(anonymous gonebot.Anonymous, msg gonebot.Message) gonebot.GroupMessageEvent {
	ev := gonebot.GroupMessageEvent{
		MessageEvent: gonebot.MessageEvent{
			Event: gonebot.Event{
				Time:      time.Now().Unix(),
				SelfId:    s.BotId,
				PostType:  gonebot.PostType_MessageEvent,
				EventName: "message.group.anonymous",
				ToMe:      false,
			},
			MessageType: "group",
			SubType:     "anonymous",
			MessageId:   s.getMsgId(),
			UserId:      0,
			Message:     msg,
			RawMessage:  msg.String(),
			Font:        0,
		},
		GroupId:   s.GroupId,
		Sender:    &gonebot.GroupMessageEventSender{},
		Anonymous: &anonymous,
	}
	s.server.sendEvent(&ev)
	return ev
}
