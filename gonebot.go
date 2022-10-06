package gonebot

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func init() {
	EngineHookManager = &engineHookManager{
		hooks: make(map[lifecycleHookType][]*func(*Engine)),
	}
}

type Engine struct {
	Handler
	Config *BaseConfig
	bot    *Bot
	ws     *WebsocketClient
}

func NewEngine(cfg Config) *Engine {
	engine := &Engine{}
	engine.Config = cfg.GetBaseConfig()

	engine.ws = NewWebsocketClient(engine.Config)
	engine.bot = NewBot(engine.ws, engine.Config)

	// 初始化handler
	engine.Handler = Handler{
		subHandlers: make(map[EventName][]*Handler),
		parent:      nil,
	}

	// 通知钩子
	EngineHookManager.runHook(LifecycleHookTypeOnCreated, engine)

	return engine
}

func (engine *Engine) Run() {
	// 启动连接到WebSocket服务器
	log.Info("开始连接到WebSocket服务器，地址：", engine.ws.url)
	go engine.ws.Start()

	// 初始化Bot
	engine.bot.Init()

	osc := make(chan os.Signal, 1)
	signal.Notify(osc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)

	// 处理消息
	ch := engine.ws.Subscribe()
MSG_LOOP:
	for {
		select {
		case by := <-ch:
			// 生成事件
			data := gjson.ParseBytes(by)
			ev := convertJsonObjectToEvent(data)

			if ev.GetPostType() == PostTypeMetaEvent {
				// log.Debug(ev.GetEventDescription())
			} else {
				log.Info(ev.GetEventDescription())
			}

			ctx := newContext(ev, engine.bot)
			go engine.handleEvent(ctx)

		case s := <-osc:
			log.Infof("收到信号%s，停止处理消息", s)
			break MSG_LOOP
		}
	}

	log.Info("正在执行清理工作")
	EngineHookManager.runHook(LifecycleHookTypeWillTerminate, engine)
}

// Engine的生命周期钩子
type engineHookManager struct {
	hooks map[lifecycleHookType][]*func(*Engine)
}

// Engine的生命周期钩子
var EngineHookManager *engineHookManager

type lifecycleHookType int

const (
	// Engine创建后触发
	LifecycleHookTypeOnCreated lifecycleHookType = iota
	LifecycleHookTypeWillTerminate
)

func (eh *engineHookManager) runHook(hookType lifecycleHookType, engine *Engine) {
	for _, hook := range eh.hooks[hookType] {
		(*hook)(engine)
	}
}

func (eh *engineHookManager) removeHook(hookType lifecycleHookType, hook *func(*Engine)) {
	for i, f := range eh.hooks[hookType] {
		if f == hook {
			eh.hooks[hookType] = append(eh.hooks[hookType][:i], eh.hooks[hookType][i+1:]...)
			break
		}
	}
}

func (eh *engineHookManager) addHook(hookType lifecycleHookType, hook *func(*Engine)) (cancel func()) {
	eh.hooks[hookType] = append(eh.hooks[hookType], hook)
	return func() {
		eh.removeHook(hookType, hook)
	}
}

// 注册OnCreated钩子
func (eh *engineHookManager) OnCreated(f func(*Engine)) (cancel func()) {
	return eh.addHook(LifecycleHookTypeOnCreated, &f)
}

// 注册WillTerminate钩子
func (eh *engineHookManager) WillTerminate(f func(*Engine)) (cancel func()) {
	return eh.addHook(LifecycleHookTypeWillTerminate, &f)
}
