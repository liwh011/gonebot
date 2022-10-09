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
		hooks: make(map[lifecycleHookType][]pHookFunc),
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
	EngineHookManager.runHook(LifecycleHookType_EngineCreated, func(phf pHookFunc) {
		f := *phf.(*EngineHookCallback)
		f(engine)
	})

	return engine
}

func (engine *Engine) Run() {
	// 启动连接到WebSocket服务器
	log.Info("开始连接到WebSocket服务器，地址：", engine.ws.url)
	go engine.ws.Start()

	// 初始化Bot
	engine.bot.Init()

	// 注册操作系统信号接收
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

			ctx := newContext(ev, engine)
			go engine.handleEvent(ctx)

		case s := <-osc:
			log.Infof("收到信号%s，停止处理消息", s)
			break MSG_LOOP
		}
	}

	log.Info("正在执行清理工作")
	EngineHookManager.runHook(LifecycleHookType_EngineWillTerminate, func(phf pHookFunc) {
		f := *phf.(*EngineHookCallback)
		f(engine)
	})
}

type pHookFunc interface{}

// Engine的生命周期钩子
type engineHookManager struct {
	hooks map[lifecycleHookType][]pHookFunc
}

// Engine的生命周期钩子
var EngineHookManager *engineHookManager

type lifecycleHookType int

const (
	// Engine创建后触发
	LifecycleHookType_EngineCreated lifecycleHookType = iota

	LifecycleHookType_PluginWillLoad
	LifecycleHookType_PluginLoaded

	LifecycleHookType_EngineWillTerminate
)

func (eh *engineHookManager) runHook(hookType lifecycleHookType, exec func(pHookFunc)) {
	for _, hook := range eh.hooks[hookType] {
		exec(hook)
	}
}

func (eh *engineHookManager) removeHook(hookType lifecycleHookType, hook pHookFunc) {
	for i, f := range eh.hooks[hookType] {
		if f == hook {
			eh.hooks[hookType] = append(eh.hooks[hookType][:i], eh.hooks[hookType][i+1:]...)
			break
		}
	}
}

func (eh *engineHookManager) addHook(hookType lifecycleHookType, hook pHookFunc) (cancel func()) {
	eh.hooks[hookType] = append(eh.hooks[hookType], hook)
	return func() {
		eh.removeHook(hookType, hook)
	}
}

type EngineHookCallback func(*Engine)

// 注册EngineCreated钩子
func (eh *engineHookManager) EngineCreated(f EngineHookCallback) (cancel func()) {
	return eh.addHook(LifecycleHookType_EngineCreated, &f)
}

// 注册EngineWillTerminate钩子
func (eh *engineHookManager) EngineWillTerminate(f EngineHookCallback) (cancel func()) {
	return eh.addHook(LifecycleHookType_EngineWillTerminate, &f)
}

type PluginHookCallback func(*PluginHub)

// 每个插件即将加载时触发
func (eh *engineHookManager) PluginWillLoad(f PluginHookCallback) (cancel func()) {
	return eh.addHook(LifecycleHookType_PluginWillLoad, &f)
}

// 每个插件加载完毕时触发
func (eh *engineHookManager) PluginLoaded(f PluginHookCallback) (cancel func()) {
	return eh.addHook(LifecycleHookType_PluginLoaded, &f)
}
