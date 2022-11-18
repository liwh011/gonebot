package gonebot

type Provider interface {
	Init(cfg Config)
	Start()
	Stop()
	Send(data []byte) (interface{}, error)
	Recieve(chan<- []byte)
}

type Adapter interface {
	Init(cfg Config, provider Provider)
	RecieveEvent(chan<- I_Event)
	Request(route string, data interface{}) (interface{}, error)
}
