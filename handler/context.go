package handler

import (
	"github.com/liwh011/gonebot/bot"
	"github.com/liwh011/gonebot/event"
	"github.com/liwh011/gonebot/message"
	log "github.com/sirupsen/logrus"
)

type Context struct {
	Event event.I_Event
	State State
	Bot   *bot.Bot
}

// 回复
func (ctx *Context) Reply(msg message.Message) {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法回复消息。(类型%s)", ctx.Event.GetEventName())
		return
	}
	data := bot.QuickOperationParams{
		"reply":       msg,
		"auto_escape": false,
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, data)
	if err != nil {
		log.Errorf("回复消息失败：%s", err.Error())
	}
}

// 回复，并对消息存在的CQ码进行转义
func (ctx *Context) ReplyRaw(msg message.Message) {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法回复消息。(类型%s)", ctx.Event.GetEventName())
		return
	}

	data := bot.QuickOperationParams{
		"reply":       msg,
		"auto_escape": true,
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, data)
	if err != nil {
		log.Errorf("回复消息失败：%s", err.Error())
	}
}

// 撤回事件对应的消息
func (ctx *Context) Delete() {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法撤回消息。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"delete": true,
	})
	if err != nil {
		log.Errorf("撤回消息失败：%s", err.Error())
	}
}

// 踢出群聊
func (ctx *Context) Kick() {
	if _, ok := ctx.Event.(*event.GroupMessageEvent); !ok {
		log.Warnf("该事件不是群聊事件，无法踢出群员。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"kick": true,
	})
	if err != nil {
		log.Errorf("踢出群员失败：%s", err.Error())
	}
}

// 禁言
func (ctx *Context) Ban(duration int) {
	if _, ok := ctx.Event.(*event.GroupMessageEvent); !ok {
		log.Warnf("该事件不是群聊事件，无法禁言群员。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"ban":          true,
		"ban_duration": duration,
	})
	if err != nil {
		log.Errorf("禁言失败：%s", err.Error())
	}
}

// 同意加好友请求
func (ctx *Context) ApproveFriendRequest() {
	if _, ok := ctx.Event.(*event.FriendRequestEvent); !ok {
		log.Warnf("该事件不是好友添加事件，无法同意好友请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"approve": true,
	})
	if err != nil {
		log.Errorf("同意好友请求失败：%s", err.Error())
	}
}

// 拒绝加好友请求
func (ctx *Context) RejectFriendRequest() {
	if _, ok := ctx.Event.(*event.FriendRequestEvent); !ok {
		log.Warnf("该事件不是好友添加事件，无法拒绝好友请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"approve": false,
	})
	if err != nil {
		log.Errorf("拒绝好友请求失败：%s", err.Error())
	}
}

// 同意加群请求、或被邀请入群请求
func (ctx *Context) ApproveGroupRequest() {
	if _, ok := ctx.Event.(*event.GroupRequestEvent); !ok {
		log.Warnf("该事件不是群添加事件，无法同意群添加请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"approve": true,
	})
	if err != nil {
		log.Errorf("同意群添加请求失败：%s", err.Error())
	}
}

// 拒绝加群请求、或被邀请入群请求
func (ctx *Context) RejectGroupRequest(reason string) {
	if _, ok := ctx.Event.(*event.GroupRequestEvent); !ok {
		log.Warnf("该事件不是群添加事件，无法拒绝群添加请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.HandleQuickOperation(ctx.Event, bot.QuickOperationParams{
		"approve": false,
		"reason":  reason,
	})
	if err != nil {
		log.Errorf("拒绝群添加请求失败：%s", err.Error())
	}
}
