package gonebot

import (
	"fmt"
	"regexp"
	"strings"
)

type HandlerFunc func(*Context)

type Handler struct {
	middlewares []HandlerFunc
	handleFunc  HandlerFunc
	parent      *Handler
	subHandlers map[EventName][]*Handler
}

// 使用中间件
func (h *Handler) Use(middlewares ...HandlerFunc) *Handler {
	h.middlewares = append(h.middlewares, middlewares...)
	return h
}

// 指定事件处理函数
func (h *Handler) Handle(f HandlerFunc) {
	h.handleFunc = f
}

// 添加子Handler
func (h *Handler) addSubHandler(subHandler *Handler, eventType ...EventName) {
	subHandler.parent = h
	for _, event := range eventType {
		h.subHandlers[event] = append(h.subHandlers[event], subHandler)
	}
}

// 移除指定的子Handler
func (h *Handler) removeSubHandler(handler *Handler, eventType ...EventName) {
	if h.subHandlers == nil {
		return
	}
	for _, event := range eventType {
		for i, subHandler := range h.subHandlers[event] {
			if subHandler == handler {
				h.subHandlers[event] = append(h.subHandlers[event][:i], h.subHandlers[event][i+1:]...)
				break
			}
		}
	}
}

// 新建一个可以被删除的Handler，用于处理指定类型的事件。
//
// 调用remove方法可以删除当前Handler。
func (h *Handler) NewRemovableHandler(eventTypes ...EventName) (handler *Handler, remove func()) {
	handler = &Handler{
		parent:      h,
		subHandlers: make(map[EventName][]*Handler),
	}
	if len(eventTypes) == 0 {
		eventTypes = append(eventTypes, EventNameAllEvent)
	}
	h.addSubHandler(handler, eventTypes...)
	return handler, func() {
		h.removeSubHandler(handler, eventTypes...)
	}
}

// 新建一个Handler，用于处理指定类型的事件
func (h *Handler) NewHandler(eventTypes ...EventName) (handler *Handler) {
	nh, _ := h.NewRemovableHandler(eventTypes...)
	return nh
}

func (h *Handler) handleEvent(ctx *Context) {
	// 替换Context中的Handler为当前正在处理的Handler
	prevHandler := ctx.Handler
	prevAction := ctx.Action
	ctx.Handler = h
	// 复原
	defer func() {
		ctx.Handler = prevHandler
		ctx.Action = prevAction
	}()

	middlewares := h.middlewares
	mwIdx := 0
	aborted := false

	subHandlers := make([]*Handler, 0)
	shIdx := 0
	next := true

	// 以下构造Handler链，以message.private.friend事件为例，
	// 按message.private.friend、message.private、message、all的顺序将这些Handler放入链中
	eventName := ctx.Event.GetEventName()
	parts := strings.Split(string(eventName), ".")
	for i := len(parts); i >= 0; i-- {
		if i == 0 {
			subHandlers = append(subHandlers, h.subHandlers[EventNameAllEvent]...)
			break
		}
		shs := h.subHandlers[EventName(strings.Join(parts[:i], "."))]
		subHandlers = append(subHandlers, shs...)
	}

	newAction := Action{}
	// middleware特供版next函数，在middleware中执行next是没用的。
	newAction.next = func() {
		// 在中间件调用，不起任何作用
		// 在处理函数调用（最后一个中间件），认为是继续下一个Handler

		// 中间件，默认向下执行，无需别的操作。
		if !aborted && mwIdx < len(middlewares) {
			return
		}
		// 处理函数，让父handler继续调用下一个子handler
		prevAction.next()
	}
	newAction.callNext = func() {
		for !aborted && mwIdx < len(middlewares) {
			mw := middlewares[mwIdx]
			mwIdx++ // 放在这里是因为中间件调用callnext时，idx必须是当前的下一个，否则会造成无限递归
			mw(ctx)
		}
		if !aborted && h.handleFunc != nil && mwIdx == len(middlewares) {
			mwIdx++ // 防止无限递归
			h.handleFunc(ctx)
		}
		if aborted {
			prevAction.next()
			return
		}
		for next && shIdx < len(subHandlers) {
			next = false
			sh := subHandlers[shIdx]
			shIdx++ // 防止无限递归
			sh.handleEvent(ctx)
		}
		if !next {
			return
		}
		prevAction.next()
		prevAction.callNext()
	}
	newAction.break_ = func() {
		aborted = true
		mwIdx = len(middlewares)
	}

	ctx.Action = newAction

	// 顺序执行中间件
	for !aborted && mwIdx < len(middlewares) {
		mw := middlewares[mwIdx]
		mwIdx++ // 放在这里是因为中间件调用callnext时，idx必须是当前的下一个，否则会造成无限递归
		mw(ctx)
	}
	// 执行处理函数
	if !aborted && h.handleFunc != nil && mwIdx == len(middlewares) {
		mwIdx++ // 防止无限递归
		h.handleFunc(ctx)
	}
	if aborted {
		// 如果中断本handler，意味着本handler没有处理事件，则继续把事件交由下一个handler
		prevAction.next()
		return
	}

	// subhandler特供版next函数，用于给子handler调用来继续处理事件
	newAction.next = func() {
		next = true
		// 若是最后一个subhandler，则继续调用父handler的next函数（先序遍历）
		if shIdx == len(subHandlers)-1 {
			prevAction.next()
		}
	}
	ctx.Action = newAction

	// 顺序执行subhandler
	for next && shIdx < len(subHandlers) {
		next = false
		sh := subHandlers[shIdx]
		shIdx++ // 防止无限递归
		sh.handleEvent(ctx)
	}
}

func OnEvent(eventName EventName) HandlerFunc {
	return func(ctx *Context) {
		if ctx.Event.GetEventName() != eventName {
			ctx.Break()
		}
	}
}

// 与Bot相关
func OnlyToMe() HandlerFunc {
	return func(ctx *Context) {
		if !ctx.Event.IsToMe() {
			ctx.Break()
		}
	}
}

// 限制来自某些群聊，当参数为空时，表示全部群聊都可
func FromGroup(groupIds ...int64) HandlerFunc {
	return func(ctx *Context) {
		gid, exist := getEventField(ctx.Event, "GroupId")
		if !exist {
			ctx.Break()
			return
		}
		if len(groupIds) == 0 {
			return
		}

		for _, id := range groupIds {
			if id == gid {
				return
			}
		}
		ctx.Break()
	}
}

// 限制来自某些人的私聊，当参数为空时，表示只要是私聊都可
func FromPrivate(userIds ...int64) HandlerFunc {
	return func(ctx *Context) {
		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			ctx.Break()
			return
		}
		if len(userIds) == 0 {
			return
		}

		for _, id := range userIds {
			if id == uid {
				return
			}
		}
		ctx.Break()
	}
}

// 消息来源于某些人，必须传入至少一个参数
func FromUser(userIds ...int64) HandlerFunc {
	return func(ctx *Context) {
		if len(userIds) == 0 {
			return
		}

		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			ctx.Break()
			return
		}
		for _, id := range userIds {
			if id == uid {
				return
			}
		}
		ctx.Break()
	}
}

func FromSession(sessionId string) HandlerFunc {
	return func(ctx *Context) {
		if ctx.Event.GetSessionId() != sessionId {
			ctx.Break()
		}
	}
}

// 事件为MessageEvent，且消息以某个前缀开头
func StartsWith(prefix ...string) HandlerFunc {
	return func(ctx *Context) {
		e := ctx.Event
		if !e.IsMessageEvent() {
			ctx.Break()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)", strings.Join(prefix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			ctx.Break()
			return
		}

		ctx.Set("prefix", map[string]interface{}{
			"matched": find,
			"text":    msgText[len(prefix):],
			"raw":     msgText,
		})
	}
}

// 事件为MessageEvent，且消息以某个后缀结尾
func EndsWith(suffix ...string) HandlerFunc {
	return func(ctx *Context) {
		e := ctx.Event
		if !e.IsMessageEvent() {
			ctx.Break()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)$", strings.Join(suffix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			ctx.Break()
			return
		}

		ctx.Set("suffix", map[string]interface{}{
			"matched": find,
			"text":    msgText[:len(msgText)-len(find)],
			"raw":     msgText,
		})
	}
}

// 事件为MessageEvent，且消息开头为指令
func Command(cmdPrefix string, cmd ...string) HandlerFunc {
	return func(ctx *Context) {
		e := ctx.Event
		if !e.IsMessageEvent() {
			ctx.Break()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^%s(%s)", cmdPrefix, strings.Join(cmd, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			ctx.Break()
			return
		}

		ctx.Set("command", map[string]interface{}{
			"raw_cmd": find,
			"matched": find[len(cmdPrefix):],
			"text":    msgText[len(find):],
			"raw":     msgText,
		})
	}
}

func FullMatch(text ...string) HandlerFunc {
	return func(ctx *Context) {
		e := ctx.Event
		if !e.IsMessageEvent() {
			ctx.Break()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(text, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			ctx.Break()
			return
		}

		ctx.Set("fullMatch", text)
	}
}

// 事件为MessageEvent，且消息中包含其中某个关键词
func Keyword(keywords ...string) HandlerFunc {
	return func(ctx *Context) {
		e := ctx.Event
		if !e.IsMessageEvent() {
			ctx.Break()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(keywords, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			ctx.Break()
			return
		}

		ctx.Set("keyword", map[string]interface{}{
			"matched": find,
		})
	}
}

// 事件为MessageEvent，且消息中存在子串满足正则表达式
func Regex(regex regexp.Regexp) HandlerFunc {
	return func(ctx *Context) {
		e := ctx.Event
		if !e.IsMessageEvent() {
			ctx.Break()
			return
		}

		msgText := e.ExtractPlainText()
		find := regex.FindStringSubmatch(msgText)
		if find == nil {
			ctx.Break()
			return
		}

		ctx.Set("regex", map[string]interface{}{
			"matched": find,
		})
	}
}
