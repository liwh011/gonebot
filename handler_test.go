package gonebot

import (
	"testing"
)

// 正常流程，只会执行第一个handler
func Test_handleEvent_default(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}
	ch := make(chan string, 2)
	handler.
		NewHandler().
		// Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			ch <- "A"
		})
	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			ch <- "B"
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if len(ch) != 1 || <-ch != "A" {
		t.Error("handleEvent error")
	}
}

// 两个兄弟handler，其中一个调用了next。
func Test_handleEvent_next(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}
	ch := make(chan string, 2)
	handler.
		NewHandler().
		// Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			ch <- "A"
			c.next()
		})
	handler.
		NewHandler().
		Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			ch <- "B"
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if len(ch) != 2 || <-ch != "A" || <-ch != "B" {
		t.Error("handleEvent error")
	}
}

// 左子树的结点调用了next
func Test_handleEvent_next2(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}
	ch := make(chan string, 5)

	h2 := handler.NewHandler()
	h2.NewHandler().
		Handle(func(c *Context) {
			ch <- "A"
			c.Next()
		})
	h2.NewHandler().
		Handle(func(c *Context) {
			ch <- "B"
			c.Next()
		})

	handler.
		NewHandler().
		// Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			ch <- "C"
		})
	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if len(ch) != 3 || <-ch != "A" || <-ch != "B" || <-ch != "C" {
		t.Error("handleEvent error")
	}
}

// 中间件中调用了break，同时两个handler为兄弟关系
func Test_handleEvent_break(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	ch := make(chan string, 5)

	handler.
		NewHandler().
		Use(func(ctx *Context) { ctx.Break() }).
		// Use(Command("哈哈哈")).
		Handle(func(c *Context) {
			ch <- "A"
		})

	handler.
		NewHandler().
		Handle(func(c *Context) {
			ch <- "B"
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if len(ch) != 1 || <-ch != "B" {
		t.Error("handleEvent error")
	}
}

// 左子树的handler中调用了break
func Test_handleEvent_break2(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	ch := make(chan string, 5)

	h2 := handler.NewHandler()
	h2.NewHandler().
		Use(func(ctx *Context) { ctx.Break() }).
		Handle(func(c *Context) {
			ch <- "A"
		})

	handler.
		NewHandler().
		Handle(func(c *Context) {
			ch <- "B"
		})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.GroupId = 114514
	msgEvent.UserId = 1919810
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if len(ch) != 1 || <-ch != "B" {
		t.Error("handleEvent error")
	}
}

func Test_handleEvent_callnext(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	ret := ""

	handler.NewHandler().Handle(func(c *Context) {
		ret += "A"
		c.callNext()
	})
	handler.NewHandler().Handle(func(c *Context) {
		ret += "B"
		c.callNext()
	})

	h2 := handler.NewHandler()
	h2.NewHandler().Handle(func(ctx *Context) {
		ret += "C"
	})
	h2.NewHandler().Handle(func(ctx *Context) {
		ret += "D"
	})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if ret != "ABC" {
		t.Error("handleEvent error")
	}
}

func Test_handleEvent_callnext2(t *testing.T) {
	handler := &Handler{
		parent:      nil,
		subHandlers: make(map[EventName][]*Handler),
	}

	ret := ""

	handler.NewHandler().Handle(func(c *Context) {
		c.callNext()
		ret += "A"
		c.callNext()
		c.callNext()
		c.callNext()
	})
	handler.NewHandler().Handle(func(c *Context) {
		ret += "B"
		c.callNext()
		c.callNext()
		c.callNext()
		c.callNext()
	})

	h2 := handler.NewHandler()
	h2.NewHandler().Handle(func(ctx *Context) {
		ret += "C"
	})
	h2.NewHandler().Handle(func(ctx *Context) {
		ret += "D"
	})

	msgEvent := &GroupMessageEvent{}
	msgEvent.EventName = EventNameGroupMessage
	msgEvent.PostType = PostTypeMessageEvent
	msgEvent.Message = MsgPrint("哈哈哈")

	ctx := newContext(msgEvent, nil)
	handler.handleEvent(ctx)

	if ret != "BCA" {
		t.Error("handleEvent error")
	}
}
