# 交互
## 创建事件处理器
一个事件处理器（Handler）用于响应一种或多种类型的事件（Event）并作出一定动作。当机器人收到相应类型的事件后，如果上游Handler允许了事件向后传播（后面会解释），那么该Handler将会被调用。Handler由中间件、处理函数、子Handler构成。

### 基本操作
下面来看上面Hello World中的一段代码：
```go
engine.NewHandler(gonebot.EventNamePrivateMessage).
    Use(gonebot.FullMatch("你几岁")).
    Handle(func(ctx *gonebot.Context) {
        ctx.Reply("24岁，是学生")
    })
```
`NewHandler(...EventName)`可以创建一个处理指定类型事件的Handler。

在OneBot中，Event通常具有一级类型、二级类型、三级类型，如好友私聊事件的三个类型依次为`message`、`private`、`friend`。这里定义Event Name为将三者以半角句号`.`串联得到的字符串，因此它描述了一个事件的具体类型。

反过来，可以得到：
- `"message.private.friend"`仅匹配好友私聊事件
- `"message.private"`匹配所有私聊事件，不局限于好友 (friend)，临时会话也可。（即不限第三级类型）
- `"message"`匹配所有聊天事件，可以是私聊 (private)，也可以是群聊 (group)。
- 特别地，我们定义`"all"`匹配所有类型的事件。

Event Name的所有枚举均已定义在`EventName`开头的常量中了。

回到`NewHandler`函数，你可以传入多个EventName来让这个Handler能处理多种事件；你也可以不传入任何参数，这个时候将默认为`"all"`。

### 可移除的Handler
`NewRemovableHandler`会返回Handler的指针和一个专属的删除函数，通过调用该函数，可以删除该Handler。

合理使用该函数，可以动态地增删Handler。下面给出了一个实现一次性Handler的例子：
```go
h, remove := engine.NewRemovableHandler(gonebot.EventNamePrivateMessage)
h.Handle(func(ctx *gonebot.Context) {
    defer remove() // 干完活就删掉
    // do something
})

```


## 使用中间件
观察到上述例子中使用了`Use`函数，这个函数接收一系列中间件 (middleware) 并将它们放入handler当中。在正式处理事件之前，这些中间件将会被按顺序调用。

中间件本质上是函数，但它可以有很多用途。你可以用它作为先决条件、预处理器、甚至后处理器。

### 内置中间件
上述例子中`gonebot.FullMatch("你几岁")`函数返回了一个中间件，它被用作先决条件，作用是保证消息内容为“你几岁”。如果不满足，则中止这个Handler的处理过程，后面跟着的中间件和处理函数均不会被执行。我们提供了多个中间件供使用：

- `OnlyToMe` : 事件与Bot相关
- `FromGroup` : 事件来源于群聊或特定群聊（可多个）
- `FromPrivate` : 事件来源于私聊或特定用户的私聊（可多个）
- `FromUser` : 事件来源于特定用户（可多个）
- `FromSession` : 事件来源于特定会话（一个私聊为一个会话；群聊中一个用户为一个会话）
- `StartsWith` : 事件为消息事件，且消息以某个前缀开头
- `EndsWith` : 事件为消息事件，且消息以某个后缀结尾
- `Command` : 事件为消息事件，且消息以指令前缀+指令名称开头
- `FullMatch` : 事件为消息事件，且消息完全匹配一段文字
- `Keyword` : 事件为消息事件，且消息中含有某个关键此
- `Regex` : 事件为消息事件，且消息存在子串符合该正则表达式

### 编写中间件
中间件本质上是`func(*gonebot.Context) bool`类型的函数，返回true代表继续下一个中间件，返回false表示令当前Handler停止处理。

这里简单介绍一下参数，因为这里是起步部分，所以不深入展开。Context中含有当前Event、Bot实例、当前Handler以及存放自由数据的Map。Context提供了一些快速操作（如回复），也提供了一些向其中写入数据的函数，你的中间件可以写入数据以供处理函数使用。


让我们一起来实操一下，开始写个卖瓜功能：
```go
// 为卖瓜编写一个中间件，看看顾客是否在找茬
func CheckZhaoCha(ctx *gonebot.Context) bool {
    text := ctx.Event.ExtractPlainText()  // 访问Event，获取消息中的纯文字
    if text == "我问你这瓜保熟吗？" {
        ctx.Set("找茬", true)  // 向CTX写入数据
        if ctx.Event.(*PrivateMessageEvent).Sender.Nickname == "刘华强" {
            return false // 中断Handler，这生意我不做了
        }
    }
    return true // 继续
}
```

### 指定处理函数
处理函数是`func(*gonebot.Context)`类型的函数，通过调用`Handler.Handle(func)`函数来指定。

在处理函数中，你可以根据你的需要来调用Context提供的快速操作，或者通过Context中的Bot实例来调用其他的API，以实现机器人对事件的反应。你也可以调用Context中的函数来控制事件处理流程。下面，我们继续完成卖瓜：
```go
engine.NewHandler(gonebot.EventNamePrivateMessage).
    Use(CheckZhaoCha).  // 使用刚刚写的找茬中间件
    Handle(func (ctx *gonebot.Context) {
        if ctx.GetBool("找茬") == true {  // 从CTX获取先前写入的数据
            ctx.Reply("你是故意找茬是不是？你要不要吧！")  // 调用CTX的快速操作
        } else {
            ctx.Reply("我开水果摊的，能卖给你生瓜蛋子？")
        }
    })
```

事实上，处理函数不是必须要指定的。由于在结构上所有Handler构成了一棵以Engine为根的树，如果你的Handler并非叶子结点，你完全可以不指定处理函数，而仅仅把它用作一个容器，为这个容器下的所有子Handler提供共同的中间件。关于Handler具体结构的内容将在后续文档说明。

# EOF
好了，到这里你已经掌握了基本的使用方法，这些使用方法足以让你写出一个简单的Bot。后面我们将会继续深入，介绍更多的用法。