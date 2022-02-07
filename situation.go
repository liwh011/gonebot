package gonebot

import (
	"fmt"
	"regexp"
	"strings"
)

type Action struct {
	Next  func() // 手动调用下一个
	Abort func() // 中止后续执行
}

type HandlerFunc func(*Context, *Action)

type Situation struct {
	middlewares []HandlerFunc
	// handlers    []HandlerFunc
}

func (s *Situation) handleEvent(ctx *Context) {
	idx := 0

	action := &Action{}
	action.Next = func() {
		for idx < len(s.middlewares) {
			s.middlewares[idx](ctx, action)
			idx++
		}
	}
	action.Abort = func() {
		idx = len(s.middlewares)
	}

	for idx < len(s.middlewares) {
		s.middlewares[idx](ctx, action)
		idx++
	}
}

func (s *Situation) OnEvent(eventName EventName) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		if ctx.Event.GetEventName() != string(eventName) {
			action.Abort()
		}
	})
	return s
}

// 与Bot相关
func (s *Situation) OnlyToMe() *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		if !ctx.Event.IsToMe() {
			action.Abort()
		}
	})
	return s
}

// 限制来自某些群聊，当参数为空时，表示全部群聊都可
func (s *Situation) FromGroup(groupIds ...int64) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		gid, exist := getEventField(ctx.Event, "GroupId")
		if !exist {
			action.Abort()
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
		action.Abort()
	})
	return s
}

// 消息来源于某些人，必须传入至少一个参数
func (s *Situation) FromUser(userIds ...int64) *Situation {
	// 没有传入任何参数时，这个函数无意义
	if len(userIds) == 0 {
		return s
	}
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			action.Abort()
			return
		}
		for _, id := range userIds {
			if id == uid {
				return
			}
		}
		action.Abort()
	})
	return s
}

// 事件为MessageEvent，且消息以某个前缀开头
func (s *Situation) StartsWith(prefix ...string) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.Abort()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)", strings.Join(prefix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.Abort()
			return
		}

		ctx.Set("prefix", map[string]interface{}{
			"matched": find,
			"text":    msgText[len(prefix):],
			"raw":     msgText,
		})
	})
	return s
}

// 事件为MessageEvent，且消息以某个后缀结尾
func (s *Situation) EndsWith(suffix ...string) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.Abort()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)$", strings.Join(suffix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.Abort()
			return
		}

		ctx.Set("suffix", map[string]interface{}{
			"matched": find,
			"text":    msgText[:len(msgText)-len(find)],
			"raw":     msgText,
		})
	})
	return s
}

// 事件为MessageEvent，且消息开头为指令
func (s *Situation) Command(cmdPrefix string, cmd ...string) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.Abort()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^%s(%s)", cmdPrefix, strings.Join(cmd, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.Abort()
			return
		}

		ctx.Set("command", map[string]interface{}{
			"raw_cmd": find,
			"matched": find[len(cmdPrefix):],
			"text":    msgText[len(find):],
			"raw":     msgText,
		})
	})
	return s
}

// 事件为MessageEvent，且消息中包含其中某个关键词
func (s *Situation) Keyword(keywords ...string) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.Abort()
			return
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(keywords, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			action.Abort()
			return
		}

		ctx.Set("keyword", map[string]interface{}{
			"matched": find,
		})
	})
	return s
}

// 事件为MessageEvent，且消息中存在子串满足正则表达式
func (s *Situation) Regex(regex regexp.Regexp) *Situation {
	s.middlewares = append(s.middlewares, func(ctx *Context, action *Action) {
		e, ok := ctx.Event.(I_MessageEvent)
		if !ok {
			action.Abort()
			return
		}

		msgText := e.ExtractPlainText()
		find := regex.FindStringSubmatch(msgText)
		if find == nil {
			action.Abort()
			return
		}

		ctx.Set("regex", map[string]interface{}{
			"matched": find,
		})
	})
	return s
}
