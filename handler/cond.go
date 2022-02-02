package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/liwh011/gonebot/event"
)

type Condition func(*Context) bool

// 满足某一事件类型
func EventType(t event.EventName) Condition {
	return func(ctx *Context) bool {
		return strings.HasPrefix(ctx.Event.GetEventName(), string(t))
	}
}

// 满足消息内容以某个字符串开头
func StartsWith(prefix ...string) Condition {
	return func(ctx *Context) bool {
		if e, ok := ctx.Event.(event.I_MessageEvent); ok {
			msgText := e.ExtractPlainText()
			reg := regexp.MustCompile(fmt.Sprintf("^(%s)", strings.Join(prefix, "|")))
			find := reg.FindString(msgText)
			if find != "" {
				writePrefixToState(ctx.State, find, msgText)
				return true
			}
		}
		return false
	}

}

// 满足消息内容以某个字符串结尾
func EndsWith(suffix ...string) Condition {
	return func(ctx *Context) bool {
		if e, ok := ctx.Event.(event.I_MessageEvent); ok {
			msgText := e.ExtractPlainText()
			reg := regexp.MustCompile(fmt.Sprintf("(%s)$", strings.Join(suffix, "|")))
			find := reg.FindString(msgText)
			if find != "" {
				writeSuffixToState(ctx.State, find, msgText)
				return true
			}
		}
		return false
	}
}

// 满足消息是个命令
func Command(cmdPref string, cmd ...string) Condition {
	return func(ctx *Context) bool {
		if e, ok := ctx.Event.(event.I_MessageEvent); ok {
			msgText := e.ExtractPlainText()
			reg := regexp.MustCompile(fmt.Sprintf("^%s(%s)", cmdPref, strings.Join(cmd, "|")))
			find := reg.FindString(msgText)
			if find != "" {
				writeCommandToState(ctx.State, find, find[len(cmdPref):], msgText)
				return true
			}
		}
		return false
	}
}

// 满足消息内容含有某个字符串
func Keyword(kw ...string) Condition {
	return func(ctx *Context) bool {
		if e, ok := ctx.Event.(event.I_MessageEvent); ok {
			msgText := e.ExtractPlainText()
			reg := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(kw, "|")))
			find := reg.FindString(msgText)
			if find != "" {
				writeKeywordToState(ctx.State, find)
				return true
			}
		}
		return false
	}
}

// 满足消息内容存在字串符合正则表达式
func Regex(regex regexp.Regexp) Condition {
	return func(ctx *Context) bool {
		if e, ok := ctx.Event.(event.I_MessageEvent); ok {
			msgText := e.ExtractPlainText()
			if regex.MatchString(msgText) {
				find := regex.FindStringSubmatch(msgText)
				writeRegexToState(ctx.State, find)
				return true
			}
		}
		return false
	}
}

// 满足实现与Bot有关
func ToMe() Condition {
	return func(ctx *Context) bool {
		if e, ok := ctx.Event.(event.I_MessageEvent); ok {
			if e.IsToMe() {
				return true
			}
		}
		return false
	}
}
