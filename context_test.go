package gonebot

import (
	"testing"
	"time"
)

func Test_WaitForNextEvent(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context, a *Action) {
			ev := c.WaitForNextEvent(90)
			if ev == nil {
				t.Error("ev should not be nil")
			}
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)

	go handler.handleEvent(ctx, &Action{})

	time.Sleep(time.Second * 2)
	msgEvent.Message = MsgPrint("999")
	handler.handleEvent(ctx, &Action{})
	time.Sleep(time.Second)
}

func Test_WaitForNextEventTimeout(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}
	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context, a *Action) {
			ev := c.WaitForNextEvent(1)
			if ev != nil {
				t.Error("ev shoule be nil")
			}
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)

	go handler.handleEvent(ctx, &Action{})

	time.Sleep(time.Second * 3)
	msgEvent.Message = MsgPrint("999")
	handler.handleEvent(ctx, &Action{})
}
