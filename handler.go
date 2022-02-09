package gonebot

import (
	"fmt"
	"regexp"
	"strings"
)

type Action struct {
	Next                 func() // 继续后续中间件的执行
	AbortHandler         func() // 中止后续中间件执行
	StopEventPropagation func() // 停止事件传播给后续Handler（当前Handler仍会继续执行完毕）
}

type HandlerFunc func(*Context, *Action)

func doNothingHandlerFunc(*Context, *Action) {}

type Handler struct {
	middlewares []HandlerFunc
	parent      *Handler
	subHandlers map[EventName][]*Handler
}

// 使用中间件
func (h *Handler) Use(middlewares ...HandlerFunc) *Handler {
	h.middlewares = append(h.middlewares, middlewares...)
	return h
}

// 指定事件及处理函数，并将其加入到Handler树中，使其生效
func (h *Handler) Handle(f HandlerFunc, eventType ...EventName) (handler *Handler, remove func()) {
	h.middlewares = append(h.middlewares, f)
	if h.parent == nil {
		return h, func() {}
	}

	// 0个参数时，表示响应全部事件
	if len(eventType) == 0 {
		eventType = append(eventType, EventNameAllEvent)
	}

	h.parent.addSubHandler(h, eventType...)
	return h, func() {
		h.parent.removeSubHandler(h, eventType...)
	}
}

// 添加子Handler
func (h *Handler) addSubHandler(handler *Handler, eventType ...EventName) {
	for _, event := range eventType {
		h.subHandlers[event] = append(h.subHandlers[event], handler)
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

// 新建一个Handler，注意此时并不会将其加入到Handler树中，不会接收到任何事件。
// 如果要将其加入到Handler树中，请对其调用Handler.Handle()
func (h *Handler) NewHandler(middlewares ...HandlerFunc) *Handler {
	return &Handler{
		parent:      h,
		subHandlers: make(map[EventName][]*Handler),
		middlewares: middlewares,
	}
}

// 新建一个单纯的Handler容器，该容器只是用来为一组Handler提供共同的middlewares
func (h *Handler) NewHandlerGroup(eventTypes ...EventName) *Handler {
	nh := &Handler{
		parent:      h,
		subHandlers: make(map[EventName][]*Handler),
		middlewares: []HandlerFunc{},
	}
	if len(eventTypes) == 0 {
		eventTypes = append(eventTypes, EventNameAllEvent)
	}
	for _, eventType := range eventTypes {
		h.addSubHandler(nh, eventType)
	}
	return nh
}

func (h *Handler) handleEvent(ctx *Context, action *Action) {
	/*
		对于middleware的action
		Next()：
			继续后续中间件的执行
		AbortHandler()：
			中止当前Handler，即停止后续中间件执行，并且subHandler也不会执行
		StopEventPropagation()：
			停止事件传播给后续Handler，但当前Handler仍会继续执行完毕
			（当前Handler的subHandler也会执行）
	*/
	idx := 0
	abort := false

	middlewareAction := &Action{}
	middlewareAction.Next = func() {
		for !abort && idx < len(h.middlewares) {
			h.middlewares[idx](ctx, middlewareAction)
			idx++
		}
	}
	middlewareAction.AbortHandler = func() {
		idx = len(h.middlewares)
		abort = true
	}
	middlewareAction.StopEventPropagation = action.StopEventPropagation

	for !abort && idx < len(h.middlewares) {
		h.middlewares[idx](ctx, middlewareAction)
		idx++
	}
	if abort {
		return
	}

	/*
		对于Handler的action
		Next()：
			继续后续Handler的执行
		AbortHandler()：
			无意义
		StopEventPropagation()：
			停止事件传播给后续Handler，但当前Handler仍会继续执行完毕
	*/

	// 以下构造Handler链，以message.private.friend事件为例，
	// 按message.private.friend、message.private、message、all的顺序将这些Handler放入链中
	eventName := ctx.Event.GetEventName()
	parts := strings.Split(string(eventName), ".")
	subHandlers := make([]*Handler, 0)
	for i := len(parts); i >= 0; i-- {
		if i == 0 {
			subHandlers = append(subHandlers, h.subHandlers[EventNameAllEvent]...)
			break
		}
		shs := h.subHandlers[EventName(strings.Join(parts[:i], "."))]
		subHandlers = append(subHandlers, shs...)
	}

	idx = 0
	stop := false

	subHandlerAction := &Action{}
	subHandlerAction.Next = func() {
		for idx < len(subHandlers) {
			subHandlers[idx].handleEvent(ctx, subHandlerAction)
			idx++
		}
	}
	subHandlerAction.AbortHandler = func() {
		panic("不应在Handler中调用AbortHandler")
	}
	subHandlerAction.StopEventPropagation = func() {
		idx = len(subHandlers)
		stop = true
	}

	for !stop && idx < len(subHandlers) {
		subHandlers[idx].handleEvent(ctx, subHandlerAction)
		idx++
	}
}

func OnEvent(eventName EventName) HandlerFunc {
	return func(ctx *Context, action *Action) {
		if ctx.Event.GetEventName() != eventName {
			action.AbortHandler()
		}
	}
}

// 与Bot相关
func OnlyToMe() HandlerFunc {
	return func(ctx *Context, action *Action) {
		if !ctx.Event.IsToMe() {
			action.AbortHandler()
		}
	}
}

// 限制来自某些群聊，当参数为空时，表示全部群聊都可
func FromGroup(groupIds ...int64) HandlerFunc {
	return func(ctx *Context, action *Action) {
		gid, exist := getEventField(ctx.Event, "GroupId")
		if !exist {
			action.AbortHandler()
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
		action.AbortHandler()
	}
}

// 限制来自某些人的私聊，当参数为空时，表示只要是私聊都可
func FromPrivate(userIds ...int64) HandlerFunc {
	return func(ctx *Context, action *Action) {
		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			action.AbortHandler()
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
		action.AbortHandler()
	}
}

// 消息来源于某些人，必须传入至少一个参数
func FromUser(userIds ...int64) HandlerFunc {
	return func(ctx *Context, action *Action) {
		if len(userIds) == 0 {
			return
		}

		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			action.AbortHandler()
			return
		}
		for _, id := range userIds {
			if id == uid {
				return
			}
		}
		action.AbortHandler()
	}
}

// 事件为MessageEvent，且消息以某个前缀开头
func StartsWith(prefix ...string) HandlerFunc {
	return func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.AbortHandler()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)", strings.Join(prefix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.AbortHandler()
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
	return func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.AbortHandler()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)$", strings.Join(suffix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.AbortHandler()
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
	return func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.AbortHandler()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^%s(%s)", cmdPrefix, strings.Join(cmd, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.AbortHandler()
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

// 事件为MessageEvent，且消息中包含其中某个关键词
func Keyword(keywords ...string) HandlerFunc {
	return func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.AbortHandler()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(keywords, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.AbortHandler()
			return
		}

		ctx.Set("keyword", map[string]interface{}{
			"matched": find,
		})
	}
}

// 事件为MessageEvent，且消息中存在子串满足正则表达式
func Regex(regex regexp.Regexp) HandlerFunc {
	return func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.AbortHandler()
			return
		}

		msgText := e.ExtractPlainText()
		find := regex.FindStringSubmatch(msgText)
		if find == nil {
			action.AbortHandler()
			return
		}

		ctx.Set("regex", map[string]interface{}{
			"matched": find,
		})
	}
}
