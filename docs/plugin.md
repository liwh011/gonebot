# 模块化
你可能已经意识到了一个问题，按照先前的方式来编写Bot的话，代码结构会比较乱。下面将介绍模块化的方式——插件。

## 编写插件
先来看一个插件的例子：
```go
package helloworld

import (
    "github.com/liwh011/gonebot"
)

func init() {
    // 注册插件，传入插件的指针。
    gonebot.NewPlugin(info, nil, onInit)
}

var info = gonebot.PluginInfo {
    Name:        "HelloWorld",
    Description: "一个插件样例",
    Version:     "0.0.1",
    Author:      "liwh011",
}

// 初始化插件
func onInit(engine *gonebot.Engine) {
    engine.
        NewHandler(gonebot.EventNamePrivateMessage).
        Use(gonebot.Keyword("你好")).
        Handle(onPrivateHello)

    engine.
        NewHandler(gonebot.EventNameGroupMessage).
        Use(gonebot.OnlyToMe(), gonebot.Keyword("老婆")).
        Handle(onLaopo)
}

func onPrivateHello(ctx *gonebot.Context, act *gonebot.Action) {
    ctx.Reply("你好！")
}

func onLaopo(ctx *gonebot.Context, act *gonebot.Action) {
    ctx.Reply("爬")
}
```

插件是一个结构体。创建这个结构体需要几个参数：
- `info` 插件信息，类型`PluginInfo`。插件信息没有什么功能性作用，只是作为一个介绍。一个插件以“名字@作者”作为唯一标识，本例中为`"HelloWorld@liwh011"`，不同插件不应出现冲突。
- `cfg` 插件配置。可以传入结构体指针。如无需配置则传入nil。配置详见下一节。
- `onInit` 用于初始化插件。这个函数接受一个engine参数，表示插件要挂载在的Engine实例。在这个函数中，你可以尽情添加你的事件处理器。

根据Go Module的特性，包被导入时会执行`init`函数。因此你需要在`init`函数中调用`gonebot.NewPlugin`来创建你的插件，这样才能被识别并使用。


## 使用插件
只需稍微改变main的写法，再把你的插件import进来即可。
```go
package main

import (
    _ "你的插件1"
    _ "你的插件2"
)

func main() {
    cfg := gonebot.LoadConfig("config.yml")
    engine := gonebot.NewEngine(cfg)
    // 使用这个Engine实例来加载插件
    gonebot.InitPlugins(engine)
    engine.Run()
}
```
假如你在GitHub相中了某个插件（假设真的有人会使用这个框架来写插件），你直接import它就可以使用了。

# EOF


