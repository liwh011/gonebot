# 快速起步
## 在此之前
你需要有一个支持正向Websocket的[OneBot实现](https://onebot.dev/ecosystem.html#onebot-%E5%AE%9E%E7%8E%B0)（如[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)），并配置好Websocket的相关参数。

**注意：如果你使用的是go-cqhttp，请将config.yml中的`message.post-format`的值更改为`array`。因为本框架暂不支持string类型消息的解析。**

## 安装
```sh
go get github.com/liwh011/gonebot
```

## Hello World!
建立一个Go项目，使用`go init`来启用Go Module支持。

创建一个go文件并输入以下内容，将WebsocketConfig配置中的参数改为你的先前为OneBot配置的参数。
```go
package main

import "github.com/liwh011/gonebot"

func main() {
	cfg := gonebot.BaseConfig{
		// 正向Websocket，ws服务器的地址
		Websocket: gonebot.WebsocketConfig{
			Host:           "127.0.0.1",
			Port:           6700,
			AccessToken:    "",
			ApiCallTimeout: 10,
		},
	}
	engine := gonebot.NewEngine(&cfg)

	// 创建一个Handler，用来处理私聊消息事件。
	engine.NewHandler(gonebot.EventNamePrivateMessage).
		Use(gonebot.FullMatch("你几岁")).
		Handle(func(ctx *gonebot.Context, act *gonebot.Action) {
			ctx.Reply("24岁，是学生")
			act.StopEventPropagation()
		})

	engine.NewHandler(gonebot.EventNamePrivateMessage).
		Handle(func(ctx *gonebot.Context, act *gonebot.Action) {
			ctx.Reply("哼哼啊啊啊啊啊啊啊啊啊啊啊啊")
		})

	// 启动
	engine.Run()
}
```
编译并运行。

这样，你就有一个野兽先辈啦！   
私聊Bot，发送“你几岁”，机器人便会回复你“24岁，是学生”；发送别的文本，则会回复你“哼哼啊啊啊啊啊啊啊啊啊啊啊啊”。

