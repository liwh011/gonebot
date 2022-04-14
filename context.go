package gonebot

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Action struct {
	next                 func() // 继续后续中间件的执行
	abortHandler         func() // 中止后续中间件执行
	stopEventPropagation func() // 停止事件传播给后续Handler（当前Handler仍会继续执行完毕）
}

func (a *Action) Next() {
	a.next()
}

func (a *Action) AbortHandler() {
	a.abortHandler()
}

func (a *Action) StopEventPropagation() {
	a.stopEventPropagation()
}

func (a *Action) AbortAndStop() {
	a.AbortHandler()
	a.StopEventPropagation()
}

type Context struct {
	Event   I_Event                // 事件（实际上是个指针）
	Keys    map[string]interface{} // 存放一些提取出来的数据
	Bot     *Bot                   // Bot实例
	Handler *Handler
	Action
}

func newContext(event I_Event, bot *Bot) *Context {
	return &Context{
		Event: event,
		Keys:  make(map[string]interface{}),
		Bot:   bot,

		Action: Action{
			next:                 func() {},
			abortHandler:         func() {},
			stopEventPropagation: func() {},
		},
	}
}

// ===================
//
// 快速操作
//
// ===================

// 回复
func (ctx *Context) Reply(args ...interface{}) {
	msg := MsgPrint(args...)
	ctx.ReplyMsg(msg)
}

// 回复
func (ctx *Context) Replyf(tmpl string, args ...interface{}) {
	msg, _ := MsgPrintf(tmpl, args...)
	ctx.ReplyMsg(msg)
}

// 使用Message对象回复
func (ctx *Context) ReplyMsg(msg Message) {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法回复消息。(类型%s)", ctx.Event.GetEventName())
		return
	}
	data := quickOperationParams{
		"reply":       msg,
		"auto_escape": false,
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, data)
	if err != nil {
		log.Errorf("回复消息失败: %s", err.Error())
	}
}

// 文字回复
func (ctx *Context) ReplyText(text string) {
	ctx.ReplyRaw(MsgPrint(text))
}

// 回复，并对消息存在的CQ码进行转义
func (ctx *Context) ReplyRaw(msg Message) {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法回复消息。(类型%s)", ctx.Event.GetEventName())
		return
	}

	data := quickOperationParams{
		"reply":       msg,
		"auto_escape": true,
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, data)
	if err != nil {
		log.Errorf("回复消息失败: %s", err.Error())
	}
}

// 撤回事件对应的消息
func (ctx *Context) Delete() {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法撤回消息。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"delete": true,
	})
	if err != nil {
		log.Errorf("撤回消息失败: %s", err.Error())
	}
}

// 踢出群聊
func (ctx *Context) Kick() {
	if _, ok := ctx.Event.(*GroupMessageEvent); !ok {
		log.Warnf("该事件不是群聊事件，无法踢出群员。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"kick": true,
	})
	if err != nil {
		log.Errorf("踢出群员失败: %s", err.Error())
	}
}

// 禁言
func (ctx *Context) Ban(duration int) {
	if _, ok := ctx.Event.(*GroupMessageEvent); !ok {
		log.Warnf("该事件不是群聊事件，无法禁言群员。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"ban":          true,
		"ban_duration": duration,
	})
	if err != nil {
		log.Errorf("禁言失败: %s", err.Error())
	}
}

// 同意加好友请求
func (ctx *Context) ApproveFriendRequest() {
	if _, ok := ctx.Event.(*FriendRequestEvent); !ok {
		log.Warnf("该事件不是好友添加事件，无法同意好友请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": true,
	})
	if err != nil {
		log.Errorf("同意好友请求失败: %s", err.Error())
	}
}

// 拒绝加好友请求
func (ctx *Context) RejectFriendRequest() {
	if _, ok := ctx.Event.(*FriendRequestEvent); !ok {
		log.Warnf("该事件不是好友添加事件，无法拒绝好友请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": false,
	})
	if err != nil {
		log.Errorf("拒绝好友请求失败: %s", err.Error())
	}
}

// 同意加群请求、或被邀请入群请求
func (ctx *Context) ApproveGroupRequest() {
	if _, ok := ctx.Event.(*GroupRequestEvent); !ok {
		log.Warnf("该事件不是群添加事件，无法同意群添加请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": true,
	})
	if err != nil {
		log.Errorf("同意群添加请求失败: %s", err.Error())
	}
}

// 拒绝加群请求、或被邀请入群请求
func (ctx *Context) RejectGroupRequest(reason string) {
	if _, ok := ctx.Event.(*GroupRequestEvent); !ok {
		log.Warnf("该事件不是群添加事件，无法拒绝群添加请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err := ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": false,
		"reason":  reason,
	})
	if err != nil {
		log.Errorf("拒绝群添加请求失败: %s", err.Error())
	}
}

// ===================
//
// 数据操作
//
// ===================

func (ctx *Context) Set(key string, value interface{}) {
	ctx.Keys[key] = value
}

func (ctx *Context) Get(key string) (v interface{}, exist bool) {
	v, exist = ctx.Keys[key]
	return
}

func (ctx *Context) MustGet(key string) interface{} {
	v, exist := ctx.Keys[key]
	if !exist {
		panic(fmt.Sprintf("键 %s 不存在", key))
	}
	return v
}

func (ctx *Context) GetString(key string) (s string) {
	v, exist := ctx.Keys[key]
	if !exist {
		return
	}
	s, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("键 %s 的值不是string", key))
	}
	return
}

func (ctx *Context) GetInt(key string) (i int) {
	v, exist := ctx.Keys[key]
	if !exist {
		return
	}
	i, ok := v.(int)
	if !ok {
		panic(fmt.Sprintf("键 %s 的值不是int", key))
	}
	return
}

func (ctx *Context) GetInt64(key string) (i64 int64) {
	v, exist := ctx.Keys[key]
	if !exist {
		return
	}
	i64, ok := v.(int64)
	if !ok {
		panic(fmt.Sprintf("键 %s 的值不是int64", key))
	}
	return
}

func (ctx *Context) GetMap(key string) (m map[string]interface{}) {
	v, exist := ctx.Keys[key]
	if !exist {
		return
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("键 %s 的值不是map[string]interface{}", key))
	}
	return
}

func (ctx *Context) GetSlice(key string) (s []interface{}) {
	v, exist := ctx.Keys[key]
	if !exist {
		return
	}
	s, ok := v.([]interface{})
	if !ok {
		panic(fmt.Sprintf("键 %s 的值不是[]interface{}", key))
	}
	return
}

// ===================
//
// 交互
//
// ===================

// 获取下一个符合条件的事件，如果没有则阻塞本事件的处理流程。
//
// timeout为超时时间（单位为秒），超时返回nil。
// middlewares为事件处理中间件，可以添加筛选条件。
func (ctx *Context) WaitForNextEvent(timeout int, middlewares ...HandlerFunc) I_Event {
	ch := make(chan I_Event, 1)

	tempHandler, remove := ctx.Handler.parent.NewRemovableHandler()
	tempHandler.Use(middlewares...).Handle(func(c *Context) {
		ch <- c.Event
		close(ch)
	})
	defer remove()

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	case event := <-ch:
		return event
	}
}

// 获取同一个Session的下一个符合条件的事件。
func (ctx *Context) WaitForNextEventInSameSession(timeout int, middlewares ...HandlerFunc) I_Event {
	middlewares = append(middlewares, FromSession(ctx.Event.GetSessionId()))
	return ctx.WaitForNextEvent(timeout, middlewares...)
}

// 发送提示消息，并获取它的回复（同一Session）。
func (ctx *Context) Prompt(message Message, timeout int) *Message {
	ctx.ReplyMsg(message)
	return ctx.WaitForNextEventInSameSession(timeout).GetMessage()
}
