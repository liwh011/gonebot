package gonebot

// 任意函数类型的指针
type pHookFunc interface{}

// Engine的生命周期钩子
type hookType int

// hook管理器基类
type hookManager struct {
	hookMap map[hookType][]pHookFunc
}

func (eh *hookManager) runHook(hookType hookType, exec func(pHookFunc)) {
	for _, hook := range eh.hookMap[hookType] {
		exec(hook)
	}
}

func (eh *hookManager) addHook(hookType hookType, hook pHookFunc) (cancel func()) {
	eh.hookMap[hookType] = append(eh.hookMap[hookType], hook)
	return func() {
		eh.removeHook(hookType, hook)
	}
}

func (eh *hookManager) removeHook(hookType hookType, hook pHookFunc) {
	for i, f := range eh.hookMap[hookType] {
		if f == hook {
			eh.hookMap[hookType] = append(eh.hookMap[hookType][:i], eh.hookMap[hookType][i+1:]...)
			break
		}
	}
}

/*
 * 全局钩子
 */

// 全局钩子，不隶属于特定的Engine实例。Engine的生命周期钩子也会在这里触发
type globalHookManager struct {
	hookManager
}

// 全局钩子，不隶属于特定的Engine实例。Engine的生命周期钩子也会在这里触发
var GlobalHooks globalHookManager = globalHookManager{
	hookManager: hookManager{
		hookMap: make(map[hookType][]pHookFunc),
	},
}

const (
	// Engine创建后触发
	engineLifecycleHook_EngineCreated hookType = iota + 1000
	engineLifecycleHook_EngineWillTerminate
)

type EngineHookCallback func(*Engine)

// 注册EngineCreated钩子
func (eh *globalHookManager) EngineCreated(f EngineHookCallback) (cancel func()) {
	return eh.addHook(engineLifecycleHook_EngineCreated, &f)
}

// 注册EngineWillTerminate钩子
func (eh *globalHookManager) EngineWillTerminate(f EngineHookCallback) (cancel func()) {
	return eh.addHook(engineLifecycleHook_EngineWillTerminate, &f)
}

// 将插件钩子放在Global的考量是，让插件在模块的init函数中就可以注册钩子，以监听到所有插件。
// 而如果放在engine实例中，按照设计，插件必须等待engine实例调用Init函数才能获取到engine实例。
// 又因为插件是按先后顺序加载的，如果一个插件在它被加载的时候才注册钩子，那么它就会错过前面插件。

// 插件生命周期钩子
const (
	pluginLifecycleHook_PluginWillLoad hookType = iota + 2000
	pluginLifecycleHook_PluginLoaded
)

type PluginHookCallback func(*PluginHub)

func (eh *globalHookManager) firePluginHook(hookType hookType, hub *PluginHub) {
	eh.runHook(hookType, func(hook pHookFunc) {
		(*hook.(*PluginHookCallback))(hub)
	})
}

// 每个插件即将加载时触发
func (eh *globalHookManager) PluginWillLoad(f PluginHookCallback) (cancel func()) {
	return eh.addHook(pluginLifecycleHook_PluginWillLoad, &f)
}

// 每个插件加载完毕时触发
func (eh *globalHookManager) PluginLoaded(f PluginHookCallback) (cancel func()) {
	return eh.addHook(pluginLifecycleHook_PluginLoaded, &f)
}

/*
 * 以下为挂在Engine上的钩子，需要通过engine.Hooks访问
 */

type engineHookManager struct {
	hookManager
}

type EventHookCallback func(I_Event)

// 事件生命周期
const (
	eventLifecycleHook_EventRecieved hookType = iota + 3000
	eventLifecycleHook_EventHandled
)

// 发出事件
func (eh *engineHookManager) fireEventHook(hookType hookType, event I_Event) {
	eh.runHook(hookType, func(hook pHookFunc) {
		(*hook.(*EventHookCallback))(event)
	})
}

// 每个事件被接收时触发
func (eh *engineHookManager) EventRecieved(f EventHookCallback) (cancel func()) {
	return eh.addHook(eventLifecycleHook_EventRecieved, &f)
}

// 每个事件被处理完毕时触发
func (eh *engineHookManager) EventHandled(f EventHookCallback) (cancel func()) {
	return eh.addHook(eventLifecycleHook_EventHandled, &f)
}
