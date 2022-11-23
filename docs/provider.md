# Provider
顾名思义，`provider`是供应商，从原产地收购货物再出售的角色（没错，就是个皮包公司）。
在本框架中，`provider`将会连接至协议方，获取下发的事件分派给框架处理，或接收框架的请求并转发至协议端。
因此，必须指定一个`provider`来保证功能正常运行。

## 内置
框架内置的Provider有：
- `websocket` 正向ws

## 使用
1. 导入Provider所在的包
   ```go
   import (
   	_ "github.com/liwh011/gonebot/providers/websocket"
   )
   ```
2. 在配置文件中修改（或添加）`provider`字段为对应的名字。如不知道名字可以先不指定该字段或乱填一个，然后编译并启动，查看控制台输出即可。
   ```yml
   provider: websocket
   ```
3. 在`provider_config`字段下提供所设定provider的配置。示例展示的是`websocket`的完整配置。
   ```yml
   provider_config:
     websocket:
       host: 127.0.0.1
       port: 6700
       access_token: asdsss
	 xxxx:
	   xxxx: xxxxx
	   ...
   ```

## 自行编写
框架规定`Provider`应实现该接口
```go
type Provider interface {
	Init(cfg gonebot.Config)
	Start()
	Stop()
	Request(route string, data interface{}) (interface{}, error)
	RecieveEvent(chan<- gonebot.I_Event)
}
```
各方法的说明如下：
- `Init` 接受配置文件，初始化内部状态
- `Start` 开始运行。可以在其中做一些操作，例如启动http服务器等。
- `Stop` 停止运行。可以在其中做一些操作，如关闭服务器等。
- `Request` 向协议端发出API请求。
- `RecieveEvent` 调用方使用该通道来接收事件，Provider将事件投入到这个通道中。

编写完成后，应在包的`init`函数中调用`gonebot.RegisterProvider`向框架注册，例如：
```go
func init() {
	gonebot.RegisterProvider("websocket", &WebsocketClientProvider{})
}
```
第一个参数为provider的名字，名字应是全局唯一的，框架依靠它来从配置文件中确定应使用哪个provider。