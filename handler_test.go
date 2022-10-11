package gonebot

/*
	TODO
	等到写完Mock后再来补这个测试吧
*/

// import (
// 	"testing"
// )

// // 正常流程，只会执行第一个handler
// func Test_handleEvent_default(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}
// 	ret := ""
// 	handler.
// 		NewHandler().
// 		// Use(Command("哈哈哈")).
// 		Handle(func(c *Context) {
// 			ret += "A"
// 		})
// 	handler.
// 		NewHandler().
// 		Use(Command("哈哈哈")).
// 		Handle(func(c *Context) {
// 			ret += "B"
// 		})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.GroupId = 114514
// 	msgEvent.UserId = 1919810
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "A" {
// 		t.Error("handleEvent error ", ret)
// 	}
// }

// // 两个兄弟handler，其中一个调用了next。
// func Test_handleEvent_next(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}
// 	ret := ""
// 	handler.
// 		NewHandler().
// 		// Use(Command("哈哈哈")).
// 		Handle(func(c *Context) {
// 			ret += "A"
// 			c.Next()
// 		})
// 	handler.
// 		NewHandler().
// 		Use(Command("哈哈哈")).
// 		Handle(func(c *Context) {
// 			ret += "B"
// 		})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.GroupId = 114514
// 	msgEvent.UserId = 1919810
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "AB" {
// 		t.Error("handleEvent error")
// 	}
// }

// // 左子树的结点调用了next
// func Test_handleEvent_next2(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}
// 	ret := ""

// 	h2 := handler.NewHandler()
// 	h2.NewHandler().
// 		Handle(func(c *Context) {
// 			ret += "A"
// 			c.Next()
// 		})
// 	h2.NewHandler().
// 		Handle(func(c *Context) {
// 			ret += "B"
// 			c.Next()
// 		})

// 	handler.
// 		NewHandler().
// 		// Use(Command("哈哈哈")).
// 		Handle(func(c *Context) {
// 			ret += "C"
// 		})
// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.GroupId = 114514
// 	msgEvent.UserId = 1919810
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "ABC" {
// 		t.Error("handleEvent error")
// 	}
// }

// // 中间件返回false，同时两个handler为兄弟关系
// func Test_handleEvent_break(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}

// 	ret := ""

// 	handler.
// 		NewHandler().
// 		Use(func(ctx *Context) bool { return false }).
// 		// Use(Command("哈哈哈")).
// 		Handle(func(c *Context) {
// 			ret += "A"
// 		})

// 	handler.
// 		NewHandler().
// 		Handle(func(c *Context) {
// 			ret += "B"
// 		})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.GroupId = 114514
// 	msgEvent.UserId = 1919810
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "B" {
// 		t.Error("handleEvent error")
// 	}
// }

// func Test_handleEvent_break2(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}

// 	ret := ""

// 	h2 := handler.NewHandler()
// 	h2.NewHandler().
// 		Use(func(ctx *Context) bool { return false }).
// 		Handle(func(c *Context) {
// 			ret += "A"
// 		})

// 	handler.
// 		NewHandler().
// 		Handle(func(c *Context) {
// 			ret += "B"
// 		})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.GroupId = 114514
// 	msgEvent.UserId = 1919810
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "B" {
// 		t.Error("handleEvent error")
// 	}
// }

// func Test_handleEvent_callnext(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}

// 	ret := ""

// 	handler.NewHandler().Handle(func(c *Context) {
// 		ret += "A"
// 		c.Next()
// 	})
// 	handler.NewHandler().Handle(func(c *Context) {
// 		ret += "B"
// 		c.Next()
// 	})

// 	h2 := handler.NewHandler()
// 	h2.NewHandler().Handle(func(ctx *Context) {
// 		ret += "C"
// 	})
// 	h2.NewHandler().Handle(func(ctx *Context) {
// 		ret += "D"
// 	})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "ABC" {
// 		t.Error("handleEvent error")
// 	}
// }

// func Test_handleEvent_callnext2(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}

// 	ret := ""

// 	handler.NewHandler().Handle(func(c *Context) {
// 		c.Next()
// 		ret += "A"
// 		c.Next()
// 		c.Next()
// 		c.Next()
// 	})
// 	handler.NewHandler().Handle(func(c *Context) {
// 		ret += "B"
// 		c.Next()
// 		c.Next()
// 		c.Next()
// 		c.Next()
// 	})

// 	h2 := handler.NewHandler()
// 	h2.NewHandler().Handle(func(ctx *Context) {
// 		ret += "C"
// 	})
// 	h2.NewHandler().Handle(func(ctx *Context) {
// 		ret += "D"
// 	})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "BCA" {
// 		t.Error("handleEvent error")
// 	}
// }

// func Test_handleEvent_callnext3(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}

// 	ret := ""

// 	handler.NewHandler().
// 		Use(func(ctx *Context) bool { ctx.Next(); return true }).
// 		Handle(func(c *Context) {
// 			ret += "A"
// 		})
// 	handler.NewHandler().Handle(func(c *Context) {
// 		ret += "B"
// 	})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "A" {
// 		t.Error("handleEvent error")
// 	}
// }

// func Test_handleEvent_callnext4(t *testing.T) {
// 	handler := &Handler{
// 		parent:      nil,
// 		subHandlers: make(map[EventName][]*Handler),
// 	}

// 	ret := ""

// 	handler.NewHandler().
// 		Use(func(ctx *Context) bool {
// 			ctx.Next()
// 			ret += "A"
// 			return true
// 		}).
// 		Handle(func(ctx *Context) {
// 			ret += "B"
// 			ctx.Next()
// 		})
// 	handler.NewHandler().Handle(func(c *Context) {
// 		ret += "C"
// 	})

// 	msgEvent := &GroupMessageEvent{}
// 	msgEvent.EventName = EventNameGroupMessage
// 	msgEvent.PostType = PostTypeMessageEvent
// 	msgEvent.Message = MsgPrint("哈哈哈")

// 	ctx := newContext(msgEvent, nil)
// 	handler.handleEvent(ctx)

// 	if ret != "BCA" {
// 		t.Error("handleEvent error", ret)
// 	}
// }
