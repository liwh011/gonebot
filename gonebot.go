package gonebot

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Engine struct {
	Handler
	Config   Config
	bot      *Bot
	provider Provider
	Hooks    engineHookManager
}

func NewEngine(cfg Config) *Engine {
	providerName := cfg.GetBaseConfig().Provider
	providerList := providers.List()
	providerListPrompt := strings.Join(providerList, ", ")
	if providerName == "" {
		if len(providerList) > 0 {
			log.Fatalf("配置文件中未设置`provider`字段，可用的Provider有：%s", providerListPrompt)
		} else {
			log.Fatal("配置文件中未设置`provider`字段，且未找到任何已注册的Provider，请导入并填写配置文件")
		}
	}
	provider, ok := providers[providerName]
	if !ok {
		if len(providerList) > 0 {
			log.Fatalf("不存在名为%s的Provider，可用的Provider有：%s", providerName, providerListPrompt)
		} else {
			log.Fatalf("不存在名为%s的Provider，且未找到任何已注册的Provider，请先导入", providerName)
		}
	}

	return NewEngineWithProvider(cfg, provider)
}

// 直接指定Provider，忽略配置文件
func NewEngineWithProvider(cfg Config, provider Provider) *Engine {
	engine := &Engine{}
	engine.Config = cfg
	engine.Hooks = engineHookManager{
		hookManager: hookManager{
			hookMap: make(map[hookType][]pHookFunc),
		},
	}

	engine.provider = provider
	engine.provider.Init(cfg)

	engine.bot = &Bot{}
	engine.bot.Init(engine.provider)

	// 初始化handler
	engine.Handler = Handler{
		subHandlers: make(map[EventName][]*Handler),
		parent:      nil,
	}

	// 通知钩子
	GlobalHooks.runHook(engineLifecycleHook_EngineCreated, func(phf pHookFunc) {
		f := *phf.(*EngineHookCallback)
		f(engine)
	})

	return engine
}

func (engine *Engine) Run() {
	if engine.provider == nil {
		log.Fatal("尚未设置Provider，请import任意一个Provider")
	}
	engine.Hooks.EventHandled(engine.provider.OnEventHandled)
	go engine.provider.Start()

	// 注册操作系统信号接收
	osc := make(chan os.Signal, 1)
	signal.Notify(osc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)

	// 处理消息
	eventCh := make(chan I_Event)
	engine.provider.RecieveEvent(eventCh)

	wg := sync.WaitGroup{}
	eventCnt := int64(0)
MSG_LOOP:
	for {
		select {
		case ev := <-eventCh:
			if ev.GetPostType() == PostType_MetaEvent {
				if ev, ok := ev.(*LifeCycleMetaEvent); ok {
					engine.bot.selfId = ev.SelfId
				}
			} else {
				log.Info(ev.GetEventDescription())
			}
			engine.Hooks.fireEventHook(eventLifecycleHook_EventRecieved, ev)

			ctx := newContext(ev, engine)
			wg.Add(1)
			atomic.AddInt64(&eventCnt, 1)
			go func() {
				defer func() {
					wg.Done()
					atomic.AddInt64(&eventCnt, -1)
				}()
				defer engine.Hooks.fireEventHook(eventLifecycleHook_EventHandled, ev)
				engine.handleEvent(ctx)
			}()

		case s := <-osc:
			log.Infof("收到信号%s，停止处理新的消息", s)
			break MSG_LOOP
		}
	}

	engine.provider.Stop()

	if eventCnt > 0 {
		log.Infof("等待剩余%d个消息处理完成，Ctrl+C以强制跳过", eventCnt)
		allEventHandled := make(chan struct{})
		go func() {
			wg.Wait()
			close(allEventHandled)
		}()
		select {
		case <-allEventHandled:
			log.Info("所有事件处理完毕")
		case <-osc:
			log.Info("已跳过等待")
		}
	}

	log.Info("正在执行清理工作")
	GlobalHooks.runHook(engineLifecycleHook_EngineWillTerminate, func(phf pHookFunc) {
		f := *phf.(*EngineHookCallback)
		f(engine)
	})
}

type providerRegistry map[string]Provider

var providers = make(providerRegistry)

// 注册一个Provider
func RegisterProvider(name string, provider Provider) {
	providers[name] = provider
}

// 列出所有已注册的provider名字
func (reg providerRegistry) List() []string {
	var list []string
	for k := range reg {
		list = append(list, k)
	}
	return list
}
