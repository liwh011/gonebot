package gonebot

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Engine struct {
	Handler
	bot *Bot
	ws  *WebsocketClient
}

func NewEngine(cfg *Config) *Engine {
	engine := &Engine{}

	wsAddr := fmt.Sprintf("ws://%s:%d/", cfg.WsHost, cfg.WsPort)
	engine.ws = NewWebsocketClient(wsAddr, cfg.ApiCallTimeout)

	engine.bot = NewBot(engine.ws, cfg)

	engine.Handler = Handler{
		subHandlers: make(map[EventName][]*Handler),
		parent:      nil,
	}

	return engine
}

func (engine *Engine) Run() {
	// 启动连接到WebSocket服务器
	go engine.ws.Start()

	// 初始化Bot
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

		ctx := newContext(ev, engine.bot)
		engine.handleEvent(ctx, &Action{func() {}, func() {}, func() {}})
	}

}

func (engine *Engine) NewService(name string) *Service {
	h := engine.NewHandler()
	h.Handle(doNothingHandlerFunc)
	sv := newService(name)
	sv.Handler = *h
	return sv
}
