package gonebot

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type Middleware func(*Context) bool
type HandlerFunc func(*Context)

type Handler struct {
	middlewares []Middleware
	handleFunc  HandlerFunc
	parent      *Handler
	subHandlers map[EventName][]*Handler
	mu          sync.RWMutex
}

// 使用中间件
func (h *Handler) Use(middlewares ...Middleware) *Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.middlewares = append(h.middlewares, middlewares...)
	return h
}

// 指定事件处理函数
func (h *Handler) Handle(f HandlerFunc) {
	h.handleFunc = f
}

// 添加子Handler
func (h *Handler) addSubHandler(subHandler *Handler, eventType ...EventName) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subHandler.parent = h
	for _, event := range eventType {
		h.subHandlers[event] = append(h.subHandlers[event], subHandler)
	}
}

// 移除指定的子Handler
func (h *Handler) removeSubHandler(handler *Handler, eventType ...EventName) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subHandlers == nil {
		return
	}
	for _, event := range eventType {
		for i, subHandler := range h.subHandlers[event] {
			if subHandler == handler {
				h.subHandlers[event] = append(h.subHandlers[event][:i], h.subHandlers[event][i+1:]...)
				return
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

func (h *Handler) getMatchedHandler(eventName EventName) (handlers []*Handler) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 以下构造Handler链，以message.private.friend事件为例，
	// 按message.private.friend、message.private、message、all的顺序将这些Handler放入链中
	parts := strings.Split(string(eventName), ".")
	for i := len(parts); i >= 0; i-- {
		if i == 0 {
			handlers = append(handlers, h.subHandlers[EventNameAllEvent]...)
			break
		}
		shs := h.subHandlers[EventName(strings.Join(parts[:i], "."))]
		handlers = append(handlers, shs...)
	}
	return
}

type execution struct {
	handlerQueue []*Handler
	curHandler   *Handler
	isLeaf       bool
	middlewares  []Middleware
	mwIdx        int
	aborted      bool
	next         bool
	ctx          *Context
	eventName    EventName
}

func (exe *execution) nextHandler() bool {
	if exe.aborted {
		return false
	}
	if !exe.next {
		return false
	}

	// 按照先序的顺序，子Handler应塞在队头
	if !exe.isLeaf {
		exe.handlerQueue = append(exe.curHandler.getMatchedHandler(exe.eventName), exe.handlerQueue...)
	}

	// 队列空，没有了
	if len(exe.handlerQueue) == 0 {
		return false
	}

	exe.curHandler = exe.handlerQueue[0]
	exe.handlerQueue = exe.handlerQueue[1:]
	exe.isLeaf = len(exe.curHandler.subHandlers) == 0
	exe.middlewares = exe.curHandler.middlewares
	exe.mwIdx = 0
	return true
}

func (exe *execution) abort() {
	exe.aborted = true
}

func (exe *execution) run() {
handlerLoop:
	for exe.mwIdx < len(exe.middlewares) || exe.nextHandler() {
		for !exe.aborted && exe.mwIdx < len(exe.middlewares) {
			mw := exe.middlewares[exe.mwIdx]
			exe.mwIdx++
			if !mw(exe.ctx) {
				exe.mwIdx = len(exe.middlewares)
				continue handlerLoop
			}
		}
		// 如果不是叶子节点，则继续执行子Handler
		if !exe.isLeaf {
			continue
		}

		if exe.aborted {
			return
		}

		exe.next = false
		if exe.curHandler.handleFunc != nil {
			exe.curHandler.handleFunc(exe.ctx)
		}
	}
}

func (exe *execution) forkAndNext() {
	if exe.aborted {
		return
	}

	newExe := *exe
	if newExe.mwIdx <= len(newExe.middlewares) {
		//
	} else {
		newExe.next = true
	}

	newExe.ctx.abort = newExe.abort
	newExe.ctx.next = newExe.forkAndNext
	defer func() {
		exe.ctx.abort = exe.abort
		exe.ctx.next = exe.forkAndNext
	}()

	// 为了防止并发调用next，导致两个goroutine同时向后执行
	// 故规定，调用next之后，原先的execution将停止处理，转由新的execution处理
	exe.aborted = true

	newExe.run()
}

func (h *Handler) handleEvent(ctx *Context) {
	exe := execution{
		handlerQueue: []*Handler{},
		curHandler:   h,
		isLeaf:       len(h.subHandlers) == 0,
		middlewares:  h.middlewares,
		mwIdx:        0,
		aborted:      false,
		next:         true,
		ctx:          ctx,
		eventName:    ctx.Event.GetEventName(),
	}

	ctx.action.abort = exe.abort
	ctx.action.next = exe.forkAndNext

	exe.run()
}

func OnEvent(eventName EventName) Middleware {
	return func(ctx *Context) bool {
		return ctx.Event.GetEventName() == eventName
	}
}

// 与Bot相关
func OnlyToMe() Middleware {
	return func(ctx *Context) bool {
		return ctx.Event.IsToMe()
	}
}

// 限制来自某些群聊，当参数为空时，表示全部群聊都可
func FromGroup(groupIds ...int64) Middleware {
	return func(ctx *Context) bool {
		gid, exist := getEventField(ctx.Event, "GroupId")
		if !exist {
			return false
		}
		if len(groupIds) == 0 {
			return true
		}

		for _, id := range groupIds {
			if id == gid {
				return true
			}
		}
		return false
	}
}

// 限制来自某些人的私聊，当参数为空时，表示只要是私聊都可
func FromPrivate(userIds ...int64) Middleware {
	return func(ctx *Context) bool {
		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			return false
		}
		if len(userIds) == 0 {
			return true
		}

		for _, id := range userIds {
			if id == uid {
				return true
			}
		}
		return false
	}
}

// 消息来源于某些人，必须传入至少一个参数
func FromUser(userIds ...int64) Middleware {
	return func(ctx *Context) bool {
		if len(userIds) == 0 {
			return true
		}

		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			return false
		}
		for _, id := range userIds {
			if id == uid {
				return true
			}
		}
		return false
	}
}

func FromSession(sessionId string) Middleware {
	return func(ctx *Context) bool {
		return ctx.Event.GetSessionId() == sessionId
	}
}

// 事件为MessageEvent，且消息以某个前缀开头
func StartsWith(prefix ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)", strings.Join(prefix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("prefix", map[string]interface{}{
			"matched": find,
			"text":    msgText[len(prefix):],
			"raw":     msgText,
		})

		return true
	}
}

// 事件为MessageEvent，且消息以某个后缀结尾
func EndsWith(suffix ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)$", strings.Join(suffix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("suffix", map[string]interface{}{
			"matched": find,
			"text":    msgText[:len(msgText)-len(find)],
			"raw":     msgText,
		})

		return true
	}
}

// 事件为MessageEvent，且消息开头为指令
func Command(cmdPrefix string, cmd ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^%s(%s)", cmdPrefix, strings.Join(cmd, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("command", map[string]interface{}{
			"raw_cmd": find,
			"matched": find[len(cmdPrefix):],
			"text":    msgText[len(find):],
			"raw":     msgText,
		})

		return true
	}
}

func FullMatch(text ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(text, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("fullMatch", text)

		return true
	}
}

// 事件为MessageEvent，且消息中包含其中某个关键词
func Keyword(keywords ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(keywords, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("keyword", map[string]interface{}{
			"matched": find,
		})

		return true
	}
}

// 事件为MessageEvent，且消息中存在子串满足正则表达式
func Regex(regex regexp.Regexp) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		find := regex.FindStringSubmatch(msgText)
		if find == nil {
			return false
		}

		ctx.Set("regex", map[string]interface{}{
			"matched": find,
		})

		return true
	}
}
