package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/liwh011/gonebot/event"
)

type Condition func(*Context) bool

func EventType(t string) Condition {
	return func(ctx *Context) bool {
		return strings.HasPrefix(ctx.Event.GetEventName(), t)
	}
}

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
