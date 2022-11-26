package gonebot

type Provider interface {
	Init(cfg Config)
	Start()
	Stop()
	// 发送API请求。
	// 参数：route为API路径；data为请求数据。
	// 返回值：API返回值，可为map、struct、数组，总之是合法的json格式；error为错误信息。
	Request(route string, data interface{}) (interface{}, error)
	RecieveEvent(chan<- I_Event)
	OnEventHandled(I_Event)
}
