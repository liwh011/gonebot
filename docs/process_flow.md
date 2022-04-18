# 流程控制
上一节我们把Context给玩明白了，这一节我们来认识与流程控制相关的操作。
但在此之前，要先介绍一下Handler的结构、以及事件处理流程，它是本节的基础。

## Handler的结构
Handler由中间件、处理函数、子Handler组成。这意味着每个Handler都能拥有自己的子Handler，可以通过调用某一Handler的`NewHandler`方法来为它创建子结点。因此，所有Handler构成了一棵树。那么这棵树的跟结点是什么？是`Engine`。实际上Engine也是一个Handler。

在前面章节的例子中，我们一直在使用`engine.NewHandler`来创建Handler，这样使得所有的Handler都是Engine的子结点。
```go
sub := engine.NewHandler(gonebot.EventNamePrivateMessage)
```

## Handler如何处理事件
1. 顺序调用中间件。
   1. 若中间件返回true，则继续。
   2. 否则，返回。
2. 如果是叶子节点（没有子Handler），则：
   1. 调用处理函数
   2. 若处理函数中未调用`ctx.Next()`，则停止处理流程；否则继续。
3. 否则，遍历每个子Handler，递归地让它处理事件。

这个过程实际上是先序遍历，事件在这棵树上沿着这个顺序流动，直到流到一个叶子节点为止。若该叶子节点放弃处理（中间件`return false`）或调用了`ctx.Next()`，则交由下一个叶子节点处理。


## 流程控制
目前市面上有些Bot框架仅仅只是将事件派发给各个处理器，这会造成一个问题：可能有多个处理器响应这个事件，表现为Bot对一条消息响应了多次。虽然说无伤大雅，但实在影响Bot的形象，一点都不高性能！为了解决这个问题，我们引入了流程处理函数来控制事件的处理流程，这些函数由Context提供。

### 继续后续执行
`Next`函数可以手动继续后续执行，当执行完毕返回，接着函数调用处执行。

注意，仅当在处理函数中调用`Next`才可以将事件继续传播！在中间件中调用`Next`仅仅只是继续当前处理器的后续。


- **例1**：在处理函数中使用Next以将事件交给下一个处理器
```go
// 功能1：接入某闲聊API，很智能，但有时候会响应失败。
engine.NewHandler(gonebot.EventNameGroupMessage).
    Use().
    Handle(func(ctx *gonebot.Context) {
        reply, ok := 某人工智能闲聊API(ctx.Event.ExtractPlainText())
        if ok {
            ctx.Reply(reply)
        } else {
            // 调用失败了，交给其他处理器来处理
            ctx.Next()
        }
    })

// 功能2：无趣的关键词匹配自动回复，谁看了都不感兴趣！也就只配处理别人不要的事件了。
engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.FullMatch("老婆")).
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("爬爬爬！")
    })

```

- **例2**：在中间件使用Next以达到置后执行的效果。“置后执行”是指，所有可能的Handler执行完毕之后。
- **例2-1**：仅有一个处理器处理事件。
```go
// 添加口癖的中间件
func KouPi(ctx *gonebot.Context) bool {
    ctx.Next() // 继续后续执行
    ctx.Reply("なのじゃ～")
    return true
}

engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.FullMatch("你几岁"), KouPi).
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("24岁，是学生")
    })

engine.NewHandler(gonebot.EventNameGroupMessage).
    Use().
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("压力马斯内") // 将不会被执行！！
    })

// 预期：
// >> 你几岁
// << 24岁，是学生
// << なのじゃ～
```

- **例2-2**：在其中一个处理器中调用了Next，共有两个处理器处理事件。
```go
// 添加口癖
func KouPi(ctx *gonebot.Context) bool {
    ctx.Next() // 继续后续执行
    ctx.Reply("なのじゃ～")
    return true
}

engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.FullMatch("你几岁"), KouPi).
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("24岁，是学生")
        ctx.Next() // 注意此处：处理函数中调用，继续事件传播
    })

engine.NewHandler(gonebot.EventNameGroupMessage).
    Use().
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("压力马斯内") // 会被执行！
    })

// 预期：
// >> 你几岁
// << 24岁，是学生
// << 压力马斯内
// << なのじゃ～
```


- **例3**：在中间件调用Next来处理Panic
```go
func Recover(ctx *gonebot.Context) bool {
    defer func (){
        p := recover()
        if p != nil {
            // 做一些错误恢复操作
            log.Error(p)
            ctx.Replyf("发生错误：%v", p)
        }
    }()
    ctx.Next()
    return true
}

engine.Use(Recover)
engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.FullMatch("给你一脚")).
    Handle(func(ctx *gonebot.Context) {
        panic("屎山倒了！")
    })
```


### 停止事件处理
`Abort`可以中止处理，中断整个事件处理流程。

- 例
```go
func SensitiveWordsFilter(ctx *gonebot.Context) bool {
    if 里面有敏感词(ctx.Event.ExtractPlainText()) {
        ctx.Abort() // 中止整个流程，后面的全部木大
    }
    return false
}

engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(SensitiveWordsFilter)
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("这条发不出去")
    })

engine.NewHandler(gonebot.EventNameGroupMessage).
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("这条也发不出去")
    })

// Expect:
// >> 敏感词敏感词敏感词敏感词
// ...bot并没有任何反应

```



# EOF
综上，事件默认只会传播到先序的第一个叶子节点，如需继续传播，则要调用`ctx.Next`。