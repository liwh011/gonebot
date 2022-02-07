package gonebot

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Engine struct {
	bot *Bot
	ws  *WebsocketClient
}

func (engine *Engine) Run(cfg *Config) {
	// 启动连接到WebSocket服务器
	wsAddr := fmt.Sprintf("ws://%s:%d/", cfg.WsHost, cfg.WsPort)
	engine.ws = NewWebsocketClient(wsAddr, cfg.ApiCallTimeout)
	go engine.ws.Start()

	// 初始化Bot
	engine.bot = NewBot(engine.ws, cfg)
	engine.bot.Init()


	// 处理消息
	ch := engine.ws.Subscribe()
	for by := range ch {
		// 生成事件
		data := gjson.ParseBytes(by)
		ev := convertJsonObjectToEvent(data)

		if ev.GetPostType() == POST_TYPE_META {
			// log.Debug(ev.GetEventDescription())
		} else {
			log.Info(ev.GetEventDescription())
		}

	}

}
