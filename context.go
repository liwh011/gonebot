package gonebot

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type action struct {
	next  func() bool
	abort func()
}

// 继续后续执行（后续执行完毕后才返回），返回值代表是否事件被Handler处理过
func (a *action) Next() bool {
	return a.next()
}

func (a *action) Abort() {
	a.abort()
}

type Context struct {
	Event             I_Event                // 事件（实际上是个指针）
	Keys              map[string]interface{} // 存放一些提取出来的数据
	Bot               *Bot                   // Bot实例
	Engine            *Engine                // Engine实例
	Handler           *Handler
	atSenderWhenReply bool
	mu                sync.RWMutex
	action
}

func newContext(event I_Event, engine *Engine) *Context {
	return &Context{
		Event:  event,
		Keys:   make(map[string]interface{}),
		Bot:    engine.bot,
		Engine: engine,

		atSenderWhenReply: true,

		action: action{
			next:  func() bool { return false },
			abort: func() {},
		},
	}
}

// ===================
//
// 快速操作
//
// ===================

func (ctx *Context) AtSender(b bool) *Context {
	ctx.atSenderWhenReply = b
	return ctx
}

// 回复
func (ctx *Context) Reply(args ...interface{}) (err error) {
	msg := MsgPrint(args...)
	return ctx.ReplyMsg(msg)
}

// 回复
func (ctx *Context) Replyf(tmpl string, args ...interface{}) (err error) {
	msg, _ := MsgPrintf(tmpl, args...)
	return ctx.ReplyMsg(msg)
}

// 使用Message对象回复
func (ctx *Context) ReplyMsg(msg Message) (err error) {
	return ctx.replyBasic(msg, nil)
}

// 文字回复
func (ctx *Context) ReplyText(text string) (err error) {
	return ctx.ReplyRaw(MsgPrint(text))
}

// 回复，并对消息存在的CQ码进行转义
func (ctx *Context) ReplyRaw(msg Message) (err error) {
	return ctx.replyBasic(msg, quickOperationParams{
		"auto_escape": true,
	})
}

func (ctx *Context) replyBasic(msg Message, params quickOperationParams) (err error) {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法回复消息。(类型%s)", ctx.Event.GetEventName())
		return
	}

	data := quickOperationParams{
		"reply":       msg,
		"auto_escape": false,
		"at_sender":   ctx.atSenderWhenReply,
	}
	for k, v := range params {
		data[k] = v
	}

	err = ctx.Bot.handleQuickOperation(ctx.Event, data)
	if err != nil {
		log.Errorf("回复消息失败: %s", err.Error())
	}
	return
}

// 撤回事件对应的消息
func (ctx *Context) Delete() (err error) {
	if !ctx.Event.IsMessageEvent() {
		log.Warnf("该事件不是消息事件，无法撤回消息。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"delete": true,
	})
	if err != nil {
		log.Errorf("撤回消息失败: %s", err.Error())
	}
	return
}

// 踢出群聊
func (ctx *Context) Kick() (err error) {
	if _, ok := ctx.Event.(*GroupMessageEvent); !ok {
		log.Warnf("该事件不是群聊事件，无法踢出群员。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"kick": true,
	})
	if err != nil {
		log.Errorf("踢出群员失败: %s", err.Error())
	}
	return
}

// 禁言
func (ctx *Context) Ban(duration int) (err error) {
	if _, ok := ctx.Event.(*GroupMessageEvent); !ok {
		log.Warnf("该事件不是群聊事件，无法禁言群员。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"ban":          true,
		"ban_duration": duration,
	})
	if err != nil {
		log.Errorf("禁言失败: %s", err.Error())
	}
	return
}

// 同意加好友请求
func (ctx *Context) ApproveFriendRequest() (err error) {
	if _, ok := ctx.Event.(*FriendRequestEvent); !ok {
		log.Warnf("该事件不是好友添加事件，无法同意好友请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": true,
	})
	if err != nil {
		log.Errorf("同意好友请求失败: %s", err.Error())
	}
	return
}

// 拒绝加好友请求
func (ctx *Context) RejectFriendRequest() (err error) {
	if _, ok := ctx.Event.(*FriendRequestEvent); !ok {
		log.Warnf("该事件不是好友添加事件，无法拒绝好友请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": false,
	})
	if err != nil {
		log.Errorf("拒绝好友请求失败: %s", err.Error())
	}
	return
}

// 同意加群请求、或被邀请入群请求
func (ctx *Context) ApproveGroupRequest() (err error) {
	if _, ok := ctx.Event.(*GroupRequestEvent); !ok {
		log.Warnf("该事件不是群添加事件，无法同意群添加请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": true,
	})
	if err != nil {
		log.Errorf("同意群添加请求失败: %s", err.Error())
	}
	return
}

// 拒绝加群请求、或被邀请入群请求
func (ctx *Context) RejectGroupRequest(reason string) (err error) {
	if _, ok := ctx.Event.(*GroupRequestEvent); !ok {
		log.Warnf("该事件不是群添加事件，无法拒绝群添加请求。(类型%s)", ctx.Event.GetEventName())
		return
	}
	err = ctx.Bot.handleQuickOperation(ctx.Event, quickOperationParams{
		"approve": false,
		"reason":  reason,
	})
	if err != nil {
		log.Errorf("拒绝群添加请求失败: %s", err.Error())
	}
	return
}

// ===================
//
// 数据操作
//
// ===================

func (ctx *Context) Set(key string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.Keys[key] = value
}

func (ctx *Context) Get(key string) (v interface{}, exist bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	v, exist = ctx.Keys[key]
	return
}

func (ctx *Context) MustGet(key string) interface{} {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	v, exist := ctx.Keys[key]
	if !exist {
		panic(fmt.Sprintf("键 %s 不存在", key))
	}
	return v
}

func (ctx *Context) GetString(key string) (s string) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

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
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

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
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

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
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

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
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

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
func (ctx *Context) WaitForNextEvent(timeout int, middlewares ...Middleware) I_Event {
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
func (ctx *Context) WaitForNextEventInSameSession(timeout int, middlewares ...Middleware) I_Event {
	middlewares = append(middlewares, FromSession(ctx.Event.GetSessionId()))
	return ctx.WaitForNextEvent(timeout, middlewares...)
}

// 发送提示消息，并获取它的回复（同一Session）。
func (ctx *Context) Prompt(message Message, timeout int) *Message {
	ctx.ReplyMsg(message)
	return ctx.WaitForNextEventInSameSession(timeout).GetMessage()
}
