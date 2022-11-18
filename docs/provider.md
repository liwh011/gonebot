# Provider
**本部分计划变动，先占坑**

顾名思义，`provider`是供应商，从原产地收购货物再出售的角色（没错，就是个皮包公司）。
在本框架中，`provider`将会连接至协议方，获取下发的事件分派给框架处理，或接收框架的请求并转发至协议端。



## 自行编写
框架规定`Provider`应实现该接口
```go
type Provider interface {
	Init(cfg Config)    // 使用配置来初始化内部状态
	Start()     // 开始运行
	Stop()      // 停止运行
	Send(data []byte) (interface{}, error)  // API调用
	Recieve(chan<- []byte) // 事件接受
}
```