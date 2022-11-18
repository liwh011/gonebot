# 模块化
你可能已经意识到了一个问题，按照先前的方式来编写Bot的话，代码结构会比较乱。下面将介绍模块化的方式——插件。

## 编写插件

### Example
先来看一个插件的例子，看完我们再细细道来。
```go
package helloworld

import (
    "github.com/liwh011/gonebot"
)

func init() {
    // 注册插件，传入插件的指针。
    gonebot.RegisterPlugin(&TestPlugin{}, nil)
}

type TestPlugin struct{}

func (p *TestPlugin) GetPluginInfo() PluginInfo {
    return gonebot.PluginInfo {
        Name:        "HelloWorld",
        Description: "一个插件样例",
        Version:     "0.0.1",
        Author:      "liwh011",
    }
}

// 初始化插件
func (p *TestPlugin) Init(hub *PluginHub) {
    hub.
        NewHandler(gonebot.EventNamePrivateMessage).
        Use(gonebot.Keyword("你好")).
        Handle(onPrivateHello)

    hub.
        NewHandler(gonebot.EventNameGroupMessage).
        Use(gonebot.OnlyToMe(), gonebot.Keyword("老婆")).
        Handle(onLaopo)
}


func onPrivateHello(ctx *gonebot.Context) {
    ctx.Reply("你好！")
}

func onLaopo(ctx *gonebot.Context) {
    ctx.Reply("爬")
}
```

### 插件介绍
```go
// 插件信息，其中Name和Author共同唯一标识一个插件
type PluginInfo struct {
	Name        string
	Author      string
	Version     string
	Description string
}

// 插件接口
type Plugin interface {
	Init(hub *PluginHub)   // 初始化插件时调用
	GetPluginInfo() PluginInfo
}
```
插件是一个接口，你需要实现以下两个函数：
- `GetPluginInfo` 获取插件信息，类型`PluginInfo`。一个插件以“名字@作者”作为唯一标识，不同插件不应出现冲突，例子中为`"HelloWorld@liwh011"`。其余字段没有什么功能性作用，只是作为一个介绍。
- `Init` 用于初始化插件。这个函数接受一个`PluginHub`对象，表示插件的“插口”，在这个函数中，你可以尽情使用`hub.NewHandler`添加你的事件处理器。

### 注册插件
写完一个插件摆在那并没有什么用，你需要注册这个插件来让框架知道插件的存在，方式为`gonebot.RegisterPlugin(pPlugin, pCfgStruct)`。这个函数接收两个参数：
- `pPlugin` 插件指针。
- `pCfgStruct` 插件配置。可以传入结构体指针。如无需配置则传入nil。配置详见[下一节](./plug_config.md)。

根据Go Module的特性，包被导入时会执行`init`函数。因此你需要在`init`函数中调用`gonebot.RegisterPlugin`来注册你的插件，这样才能被识别并使用。


## 使用插件
只需把你的插件import进来即可。
```go
package main

import (
    _ "你的插件1"
    _ "你的插件2"
)

func main() {
    cfg := gonebot.LoadConfig("config.yml")
    engine := gonebot.NewEngine(cfg)
    engine.Run()
}
```
假如你在GitHub相中了某个插件（假设真的有人会使用这个框架来写插件），你直接import它就可以使用了。

# EOF


