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
		Handle(func(c *Context) {
			ev := c.WaitForNextEvent(90)
			if ev == nil {
				t.Error("ev should not be nil")
			}
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventName_GroupMessage
	msgEvent.PostType = PostType_MessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	ctx.Handler = handler

	go handler.handleEvent(ctx)

	time.Sleep(time.Second * 2)
	msgEvent.Message = MsgPrint("999")
	handler.handleEvent(ctx)
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
		Handle(func(c *Context) {
			ev := c.WaitForNextEvent(1)
			if ev != nil {
				t.Error("ev shoule be nil")
			}
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventName_GroupMessage
	msgEvent.PostType = PostType_MessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	ctx.Handler = handler

	go handler.handleEvent(ctx)

	time.Sleep(time.Second * 3)
	msgEvent.Message = MsgPrint("999")
	handler.handleEvent(ctx)
}

func Test_ReturnValueOfNext(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	handler.
		NewHandler().
		Handle(func(c *Context) {
			processed := c.Next()
			if processed {
				t.Log("Next的返回值为", processed)
			} else {
				t.Error("Next的返回值为", processed, "，应该为true")
			}
		})

	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			t.Log("事件被处理")
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventName_GroupMessage
	msgEvent.PostType = PostType_MessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")
	ctx := newContext(msgEvent, nil)

	handler.handleEvent(ctx)
}

func Test_ReturnValueOfNext2(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	handler.
		NewHandler().
		Handle(func(c *Context) {
			processed := c.Next()
			if !processed {
				t.Log("Next的返回值为", processed)
			} else {
				t.Error("Next的返回值为", processed, "，应该为false")
			}
		})

	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			t.Log("事件被处理")
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventName_GroupMessage
	msgEvent.PostType = PostType_MessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("114514")
	ctx := newContext(msgEvent, nil)

	handler.handleEvent(ctx)
}

func Test_ReturnValueOfNext3(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	handler.
		NewHandler().
		Handle(func(c *Context) {
			processed := c.Next()
			expect := true
			if processed != expect {
				t.Error("Next的返回值为", processed, "，应该为", expect)
			}
		})

	handler.
		NewHandler().
		Use(func(c *Context) bool {
			processed := c.Next()
			expect := true
			if processed != expect {
				t.Error("Next的返回值为", processed, "，应该为", expect)
			}
			return false
		})

	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			t.Log("事件被处理")
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventName_GroupMessage
	msgEvent.PostType = PostType_MessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")
	ctx := newContext(msgEvent, nil)

	handler.handleEvent(ctx)
}
