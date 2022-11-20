package gonebot

type Provider interface {
	Init(cfg Config)
	Start()
	Stop()
	Request(route string, data interface{}) (interface{}, error)
	RecieveEvent(chan<- I_Event)
}
