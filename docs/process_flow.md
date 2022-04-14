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
（本节将中间件和处理函数合并，统一叫做中间件）
1. 顺序调用中间件。
2. 遍历每个子Handler，递归地让它处理事件。（如有）

这个过程实际上是一个先序遍历，事件沿着这个顺序流动。


## 流程控制
目前市面上有些Bot框架仅仅只是将事件派发给各个处理器，这会造成一个问题：可能有多个处理器响应这个事件，表现为Bot对一条消息响应了多次。虽然说无伤大雅，但实在影响Bot的形象，一点都不高性能！为了解决这个问题，我们引入了流程处理函数来控制事件的处理流程，这些函数由Context提供。

### 中止当前Handler
`AbortHandler`可以中断当前Handler的执行。在中间件内部调用该函数，后面的所有中间件都不会被调用，子Handler也不会被调用。
```go
// 该handler匹配以“开头”开头、以“结尾”结尾、中间含有“关键词”的群消息
h := engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.StartsWith("开头"), gonebot.EndsWith("结尾"), gonebot.Keyword("关键词"))

s1 := h.NewHandler()  // h的儿子1
s2 := h.NewHandler()  // h的儿子2
```
例子中，当收到“开头xxx”的消息时，由于不符合`EndsWith("结尾")`的要求，该Handler将中断，后续的`Keyword`及s1、s2都不会被调用。

### 停止事件传播
`StopEventPropagation`可以停止事件传播，中断整个事件处理流程，但会继续运行完当前的Handler。

- **情形1**：
```go
/*
        engine
       /     \
     <p1>    p2❌
    /    \
   s1    s2
*/
p1 := engine.NewHandler(gonebot.EventNameGroupMessage).
p1.Use(gonebot.Command("打我"))).
    Handle(func (ctx *gonebot.Context) {
        ctx.StopEventPropagation()
        // do something
    })

p2 := engine.NewHandler(gonebot.EventNameGroupMessage) // 略

s1 := p1.NewHandler()  // p1的儿子1
s2 := p1.NewHandler()  // p2的儿子2
```
当收到“打我”这条消息时，`StopEventPropagation`被调用，事件将无法被传播到p2，而p1将会继续执行完毕（子Handler s1和s2会被调用）。

- **情形2**：
```go
/*
        engine
       /     \
      p1     p2❌
    /    \
  <s1>   s2❌
    |
   ss1
*/
p1 := engine.NewHandler(gonebot.EventNameGroupMessage) // 略
p2 := engine.NewHandler(gonebot.EventNameGroupMessage) // 略

s1 := p1.NewHandler()  // p1的儿子
s1.Use(gonebot.Command("打我"))).
    Handle(func (ctx *gonebot.Context) {
        ctx.StopEventPropagation()
        // do something
    })
ss1 := s1.NewHandler() // s1的儿子1

s2 := p1.NewHandler()  // p2的儿子2
```
当收到“打我”这条消息时，s1会继续执行完毕（ss1会被调用），而s2、p2不会被调用。

### 继续后续执行
`Next`函数可以手动继续后续中间件的执行，当执行完毕返回，接着函数调用处执行。
```go
// 添加口癖
func KouPi(ctx *gonebot.Context, act *gonebot.Action) {
    act.Next() // 继续后续执行
    ctx.Reply("なのじゃ～")
}

engine.NewHandler(gonebot.EventNamePrivateMessage).
    Use(KouPi, gonebot.FullMatch("你几岁")).
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("24岁，是学生")
        ctx.StopEventPropagation()
    })

// 预期：
// > 24岁，是学生
// > なのじゃ～
```

# EOF
当你不调用任何控制函数时，是默认继续把事件传播下去。我阅读过一些别的库，它们默认停止事件的传播。我也不知道怎么样比较好，如果你有这方面的想法，欢迎讨论改进。