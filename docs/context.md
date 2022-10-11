# 上下文
先前我们了解了处理事件的基本流程和方法，在本节我们将了解Context的用处。

## 了解上下文
上下文 (Context) 包括了事件本身以及参与事件处理的相关方。每收到一个事件，会创建一个上下文对象，并将它传给Handler来进入事件处理流程。Context具有以下几个字段：
- `Event` : 事件
- `Bot` : 收到事件的Bot
- `Handler` : 正在处理该事件的Handler
- `Keys` : 可以自由读写的Map

## 快速操作
Context基于Bot提供的API再次封装，提供了一些快速操作，让你能方便地对事件做出响应。

- `Reply` : 回复，仅限消息事件
- `Delete` : 撤回，仅限消息事件，机器人需要有权限
- `Kick` : 踢出，仅限消息事件，机器人需要有权限
- `Ban` : 禁言，仅限消息事件，机器人需要有权限
- `ApproveFriendRequest` : 同意加好友请求
- `RejectFriendRequest` : 拒绝加好友请求
- `ApproveGroupRequest` : 同意他人的加群请求、或同意自己被邀请入群请求
- `RejectGroupRequest` : 拒绝他人的加群请求、或拒绝自己被邀请入群请求

```go
engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.Keyword("涩图")).
    Handle(func(ctx *gonebot.Context) {
        ctx.Ban(10)  // 禁言
        ctx.Reply("不可以涩涩")
    })
```

## 调用Bot
我们来看一个例子，这个例子可以在群聊收集机器人的问题并告知Bot管理员：
```go
// 将在群聊中报告的问题转发给Bot管理员
engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.StartsWith("报告问题")).
    Handle(func(ctx *gonebot.Context) {
        // 构造消息对象
        msg := gonebot.MsgMustPrintf("来自%s报告的问题：{}", ctx.Event.GetSessionId(), ctx.Event.GetMessage())
        // 调用SendPrivateMsg，私发给超管
        superUserId := 1919810
        ctx.Bot.SendPrivateMsg(superUserId, msg, false)
    })
```
Context提供的快速操作显然无法实现定向发送给某个用户的功能，因此需要靠调用Bot提供的原始API来实现。详细API暂未有文档，可以参考[OneBotAPI](https://github.com/botuniverse/onebot-11/blob/master/api/public.md)

## 处理流程控制
Context提供了若干事件处理流程的控制函数，详见[下一节](./process_flow.md)

## 读写数据
Context中存在`Keys`字段，可以供你存取你自己的数据（一般是中间处理结果）。


### 写入
回顾上一节中的卖瓜例子：
```go
func CheckZhaoCha(ctx *gonebot.Context) bool {
    text := ctx.Event.ExtractPlainText() 
    if text == "我问你这瓜保熟吗？" {
        ctx.Set("找茬", true)  // 向CTX写入数据
        if ctx.Event.(*PrivateMessageEvent).Sender.Nickname == "刘华强" {
            return false
        }
    }
    return true
}
```
该例子通过使用Context的`Set`函数，向Keys写入数据，以供后续使用。

类似地，内置的中间件也会向其中写入一些数据。例如`Command`将会使用`"command"`这个键并写入处理后的数据，并提供了`ctx.GetCommandMatchResult`方法来便捷地获取这些数据。你可以合理使用这些字段来避免手动处理文本。

其他内置中间件暂且还请[阅读源代码](https://github.com/liwh011/gonebot/blob/master/handler.go)~

### 读取
Context提供了一系列Get方法来获取数据。
- `Get` 将会返回interface{}和是否存在。
- `MustGet` 则会仅返回interface{}，忽略是否存在。
- `GetXXX` （XXX为类型）将会返回转换后的值。若对应键不存在，则返回零值。若类型不正确，则会panic。

```go
// ctx中只有一个age字段。
ctx.Set("age", 24)

v, exist := ctx.Get("age")   // 24 (类型为interface{}), true
v, exist := ctx.Get("name")  // nil (类型为interface{}), false

v := ctx.MustGet("age")   // 24 (类型为interface{})
v := ctx.MustGet("name")  // nil (类型为interface{})

v := ctx.GetInt("age")       // 24
v := ctx.GetString("age")    // panic
v := ctx.GetInt("name")      // 0
v := ctx.GetString("name")   // ""
```

## 交互增强
有时候我们往往希望能够保持状态来响应下一次内容。脑补一下这个场景：

现有一个指令为`雷普<@用户> <次数>次`。由于这条指令比较复杂，用户经常遗漏某个参数。而你的Handler仅仅只是发送一条提示，告诉用户输入错了，就结束了事件处理过程。用户还得再尝试一次甚至多次才能输入正确。
> User: 雷普@野兽先辈   
> Bot: 指令错误！缺少次数   
> User: 雷普10次@野兽先辈   
> Bot: 指令错误！   
> User: 雷普@野兽先辈 114次   
> Bot: 雷普成功！   

想必用户都想把Bot撅烂了罢！有没有一种可能，机器人只问缺失的东西，用户只用回复需要的东西？
> User: 雷普@野兽先辈   
> Bot: 你想要雷普多少次？   
> User: 114514次   
> Bot: 雷普成功！   

有！Context提供了几个Wait函数，用于在不结束当前Handler的情况下，获取下一个符合条件的事件。它将会阻塞当前Handler（甚至当前事件处理流程），直到接收到符合条件的事件或超过超时时间。

### 获取未来事件
`WaitForNextEvent`接受timeout参数和middleware参数，你可以传入你的middleware来筛选事件。超时则返回nil。
```go
// 打断复读
engine.NewHandler(gonebot.EventNameGroupMessage).
    Handle(func(ctx *gonebot.Context) {
        text := ctx.Event.ExtractPlainText()
        // 获取下一个文字内容与当前事件完全一致的事件
        next := ctx.WaitForNextEvent(10, gonebot.FullMatch(text))
        // 超时返回nil
        if next != nil {
            ctx.Reply("禁止复读！")
        }
    })
```

### 等待用户输入
`Prompt`则是对`WaitForNextEvent`的进一步封装。它接受message和timeout，将提示消息发出去后并等待该用户的回复。若用户在timeout时间内回复，则返回该消息对象，否则返回nil。
```go
// 简单起见，我们将指令简化为`雷普<@某人>`
engine.NewHandler(gonebot.EventNameGroupMessage).
    Use(gonebot.Command("雷普")).
    Handle(func(ctx *gonebot.Context) {
        msg := ctx.Event.GetMessage()
        targetIdStr := ""
        if msg.Len() <= 1 || (*msg)[1].Type != "at" {
            // 发送提示并等待10s
            m1 := ctx.Prompt("你要雷普谁？请@他", 10)
            // 超时了
            if m1 == nil {
                return
            }
            targetIdStr = (*m1)[0].Data["qq"]
        } else {
            targetIdStr = (*msg)[1].Data["qq"]
        }
        // do something...
    })
```
实际上，我推荐将参数解析过程写成中间件的形式，将解析结果写入Context。而不是统统塞在处理函数里面。

# EOF
你已经掌握Context啦，你能写更复杂的处理逻辑了，快去试试吧。