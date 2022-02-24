package gonebot

import (
	"testing"
)

func Test_Run(t *testing.T) {
	cfg := BaseConfig{
		Websocket: WebsocketConfig{
			Host:           "127.0.0.1",
			Port:           6700,
			AccessToken:    "asdsss",
			ApiCallTimeout: 10,
		},
	}
	engine := NewEngine(&cfg)

	NewPlugin(PluginInfo{}, nil, func(p *Plugin) {
		p.NewHandler(EventNameGroupMessage).Use(StartsWith("啊啊啊")).Handle(func(c *Context, a *Action) {
			c.ReplyText("啊")
			p.Unload()
		})
	})

	LoadPlugins(engine)

	engine.Run()
}
