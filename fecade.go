package gonebot

import (
	"fmt"

	"github.com/liwh011/gonebot/bot"
	"github.com/liwh011/gonebot/config"
	"github.com/liwh011/gonebot/driver"
	"github.com/liwh011/gonebot/event"
	"github.com/liwh011/gonebot/handler"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	bot_     *bot.Bot
	ws       *driver.WebsocketClient
	handlers []*handler.EventHandler
)

// func Init(cfg *config.Config) {
// 	wsAddr := fmt.Sprintf("ws://%s:%d/", cfg.WsHost, cfg.WsPort)
// 	ws = driver.NewWsClient(wsAddr, cfg.ApiCallTimeout)
// 	bot_ = bot.NewBot(ws, cfg)
// }

func Run(cfg *config.Config) {
	// 启动连接到WebSocket服务器
	wsAddr := fmt.Sprintf("ws://%s:%d/", cfg.WsHost, cfg.WsPort)
	ws = driver.NewWsClient(wsAddr, cfg.ApiCallTimeout)
	go ws.Start()

	// 初始化Bot
	bot_ = bot.NewBot(ws, cfg)
	bot_.Init()

	// 处理消息
	ch := ws.Subscribe()
	for by := range ch {
		// 生成事件
		data := gjson.ParseBytes(by)
		ev := event.FromJsonObject(data)

		if ev.GetPostType() == event.POST_TYPE_META {
			// log.Debug(ev.GetEventDescription())
		} else {
			log.Info(ev.GetEventDescription())
		}

		// 打包成Context
		ctx := handler.Context{
			Event: ev,
			Bot:   bot_,
			State: make(map[string]interface{}),
		}

		// 分发给handlers
		// sort.Slice(handlers, func(i, j int) bool {
		// 	return handlers[i].Priority > handlers[j].Priority
		// })
		for _, h := range handlers {
			h.Handle(&ctx)
		}
	}

}
