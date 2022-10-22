package mock

import (
	"errors"
	"fmt"
	"time"

	"github.com/liwh011/gonebot"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type resp map[string]interface{}

// 模拟的协议提供端
type MockServer struct {
	msgId   int32
	BotId   int64  // 机器人QQ号
	BotName string // 机器人昵称

	Friends []User  // 好友列表
	Groups  []Group // 群聊列表

	AutoSendEvent    bool          // 是否自动发送事件，默认true
	sendEventTimeout time.Duration // 发送事件超时时间，默认1秒
	eventRecievers   []chan<- gonebot.I_Event

	messageHistory MessageHistory // 所有消息记录
}

type NewMockServerOptions struct {
	BotId   int64
	BotName string
	Friends []User
	Groups  []Group

	SendEventTimeout time.Duration
}

func NewMockServer(options NewMockServerOptions) *MockServer {
	if options.BotId == 0 {
		options.BotId = 10000
	}
	if options.BotName == "" {
		options.BotName = "MockBot"
	}
	if options.SendEventTimeout == 0 {
		options.SendEventTimeout = time.Second
	}
	return &MockServer{
		BotId:            options.BotId,
		BotName:          options.BotName,
		Friends:          options.Friends,
		Groups:           options.Groups,
		AutoSendEvent:    true,
		sendEventTimeout: options.SendEventTimeout,
	}
}

// 注册事件接收器，注册后会自动发送一个Connected的生命周期事件
func (server *MockServer) RegisterEventReciever(reciever chan<- gonebot.I_Event) {
	server.eventRecievers = append(server.eventRecievers, reciever)
	server.ConnectedEvent()
}

func (server *MockServer) getMsgId() int32 {
	server.msgId++
	return server.msgId
}

// 下发事件
func (server *MockServer) SendEvent(event gonebot.I_Event) {
	if server.AutoSendEvent {
		for i, reciever := range server.eventRecievers {
			select {
			case reciever <- event:
				// pass
			case <-time.After(server.sendEventTimeout):
				logrus.Errorf("下发事件%s超时，阻塞在第%d个接收者", event.GetEventName(), i)
			}
		}

		server.addEventToMessageHistory(event)
	}
}

// 模拟一个生命周期的Connected事件
func (server *MockServer) ConnectedEvent() gonebot.LifeCycleMetaEvent {
	ev := gonebot.LifeCycleMetaEvent{
		Event: gonebot.Event{
			Time:      time.Now().Unix(),
			SelfId:    server.BotId,
			PostType:  gonebot.PostType_MetaEvent,
			EventName: "meta_event.lifecycle.connected",
			ToMe:      false,
		},
		MetaEventType: "lifecycle",
		SubType:       "connect",
	}
	server.SendEvent(&ev)
	return ev
}

// 处理API请求
func (server *MockServer) HandleRequest(action string, params gjson.Result) (interface{}, error) {
	switch action {
	case "send_private_msg":
		userId := params.Get("user_id").Int()
		groupId := params.Get("group_id").Int()
		message := gonebot.ConvertJsonArrayToMessage(params.Get("message").Array())
		msgId := server.getMsgId()
		server.addMyMessageToMessageHistory(msgId, message, userId, 0)
		if groupId == 0 {
			logrus.Infof("发送私聊消息给%d：%s", userId, message.String())
		} else {
			logrus.Infof("发送私聊消息（临时会话）给群聊%d的成员%d：%s", groupId, userId, message.String())
		}
		return resp{"message_id": msgId}, nil

	case "send_group_msg":
		groupId := params.Get("group_id").Int()
		message := gonebot.ConvertJsonArrayToMessage(params.Get("message").Array())
		msgId := server.getMsgId()
		server.addMyMessageToMessageHistory(msgId, message, 0, groupId)
		logrus.Infof("发送群聊消息到%d：%s", groupId, message.String())
		return resp{"message_id": msgId}, nil

	case "send_msg":
		userId := params.Get("user_id").Int()
		groupId := params.Get("group_id").Int()
		message := gonebot.ConvertJsonArrayToMessage(params.Get("message").Array())
		msgId := server.getMsgId()
		server.addMyMessageToMessageHistory(msgId, message, userId, groupId)
		return resp{"message_id": msgId}, nil

	case "send_group_forward_msg":
		return nil, nil
	case "delete_msg":
		msgId := params.Get("message_id").Int()
		logrus.Infof("撤回消息%d", msgId)
		return nil, nil
	case "mark_msg_as_read":
		return nil, nil
	case "set_group_kick":
		return nil, nil
	case "set_group_ban":
		return nil, nil
	case "set_group_anonymous_ban":
		return nil, nil
	case "set_group_whole_ban":
		return nil, nil
	case "set_group_admin":
		return nil, nil
	case "set_group_anonymous":
		return nil, nil
	case "set_group_card":
		return nil, nil
	case "set_group_name":
		return nil, nil
	case "set_group_leave":
		return nil, nil
	case "set_group_special_title":
		return nil, nil
	case "send_group_sign":
		return nil, nil
	case "set_friend_add_request":
		return nil, nil
	case "set_group_add_request":
		return nil, nil
	case "get_login_info":
		return resp{"user_id": server.BotId, "nickname": server.BotName}, nil
	case "set_qq_profile":
		return nil, nil
	case "delete_friend":
		return nil, nil
	case "get_group_list":
		return server.Groups, nil

	case ".handle_quick_operation":
		opParams := params.Get("operation")
		switch {
		case opParams.Get("reply").Exists():
			reply := opParams.Get("reply").Array()
			msg := gonebot.ConvertJsonArrayToMessage(reply)
			msgId := server.getMsgId()
			ev := gonebot.ConvertJsonObjectToEvent(params.Get("context"))
			switch ev := ev.(type) {
			case *gonebot.GroupMessageEvent:
				server.addMyMessageToMessageHistory(msgId, msg, 0, ev.GroupId)
			case *gonebot.PrivateMessageEvent:
				server.addMyMessageToMessageHistory(msgId, msg, ev.UserId, 0)
			}
			logrus.Infof("回复消息：[起始]%s[结束]", msg.String())
		case opParams.Get("ban").Exists():
			logrus.Infof("禁言消息发送者%d分钟", opParams.Get("ban_duration").Int())
		case opParams.Get("kick").Exists():
			logrus.Infof("将消息发送者踢出群聊")
		case opParams.Get("delete").Exists():
			logrus.Infof("撤回该消息")
		default:
			logrus.Infof("快速操作：%v", opParams)
		}
		return nil, nil
	}

	return nil, errors.New("not implemented action: " + action)
}

// 创建私聊会话。userId为对方的QQ号，如果这个QQ号是好友，后续模拟的消息事件的SubType将为"friend"，否则为"other"
func (server *MockServer) NewPrivateSession(userId int64) *PrivateSession {
	var user *User
	for _, u := range server.Friends {
		if u.UserId == userId {
			user = &u
			break
		}
	}

	if user == nil {
		user = &User{
			UserId:   userId,
			Nickname: fmt.Sprintf("陌生人%d", userId),
		}
	}

	return &PrivateSession{
		Server:   server,
		UserId:   userId,
		Nickname: user.Nickname,
		Sex:      user.Sex,
		Age:      user.Age,
		BotId:    server.BotId,
		IsFriend: user != nil,
	}
}

// 创建群聊会话。groupId为群号。若群号不存在，则创建一个空群聊。
func (server *MockServer) NewGroupSession(groupId int64) *GroupSession {
	var group *Group
	for _, g := range server.Groups {
		if g.GroupId == groupId {
			group = &g
			break
		}
	}

	if group == nil {
		group = &Group{
			GroupId:   groupId,
			GroupName: fmt.Sprintf("未知群聊%d", groupId),
		}
	}

	return &GroupSession{
		Server:    server,
		GroupId:   groupId,
		GroupName: group.GroupName,
		BotId:     server.BotId,
		Group:     *group,
	}
}

// 将Event转为消息记录
func (server *MockServer) addEventToMessageHistory(event gonebot.I_Event) {
	var rcd MessageRecord
	switch event := event.(type) {
	case *gonebot.PrivateMessageEvent:
		rcd = MessageRecord{
			MsgId:     event.MessageId,
			Msg:       event.Message,
			UserId:    event.UserId,
			Nickname:  event.Sender.Nickname,
			Time:      event.Time,
			SessionId: event.UserId,
		}
	case *gonebot.GroupMessageEvent:
		rcd = MessageRecord{
			MsgId:     event.MessageId,
			Msg:       event.Message,
			UserId:    event.UserId,
			Nickname:  event.Sender.Nickname,
			Time:      event.Time,
			GroupId:   event.GroupId,
			SessionId: event.GroupId,
		}
	default:
		return
	}
	server.messageHistory = append(server.messageHistory, rcd)
}

func (server *MockServer) addMyMessageToMessageHistory(msgId int32, msg gonebot.Message, toUser int64, toGroup int64) {
	rcd := MessageRecord{
		MsgId:     msgId,
		Msg:       msg,
		UserId:    server.BotId,
		Nickname:  server.BotName,
		Time:      time.Now().Unix(),
		GroupId:   toGroup,
		SessionId: toGroup,
	}
	if toGroup != 0 {
		rcd.GroupId = toGroup
		rcd.SessionId = toGroup
	} else {
		rcd.GroupId = 0
		rcd.SessionId = toUser
	}
	server.messageHistory = append(server.messageHistory, rcd)
}
