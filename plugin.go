package gonebot

import (
	"fmt"
	"reflect"
)

type PluginInfo struct {
	Name        string
	Description string
	Version     string
	Author      string
}

type Plugin struct {
	Handler
	id     string
	engine *Engine // 插件所属的engine
	Info   PluginInfo
	Config interface{} // 结构体指针
	onInit func(*Plugin)
}

func (p *Plugin) GetEngine() *Engine {
	return p.engine
}

// 添加在某个Engine上
func (p *Plugin) AttachTo(engine *Engine) {
	p.engine = engine
	engine.addSubHandler(&p.Handler, EventNameAllEvent)
}

// 卸载插件
func (p *Plugin) Unload() {
	p.engine.removeSubHandler(&p.Handler, EventNameAllEvent)
}

var plugins = make(map[string]*Plugin)

// 注册一个新插件。
//
// info: 插件信息，包括名称、描述、版本、作者
//
// cfg: 插件配置，传入结构体指针，如果无需配置可以为 nil
//
// onInit: 插件初始化时的回调函数，如果无需初始化可以为 nil
func NewPlugin(info PluginInfo, cfg interface{}, onInit func(*Plugin)) *Plugin {
	np := &Plugin{
		Handler: Handler{
			subHandlers: make(map[EventName][]*Handler),
		},
		id:     fmt.Sprintf("%s@%s", info.Name, info.Author),
		Info:   info,
		Config: cfg,
		onInit: onInit,
	}

	if p, ok := plugins[np.id]; ok {
		panic(fmt.Errorf("插件 %s 已经注册。该插件为 %v", np.id, p.Info))
	}

	plugins[np.id] = np
	return np
}

func GetPlugin(id string) *Plugin {
	return plugins[id]
}

func LoadPlugins(engine *Engine) {
	cfg := engine.Config

	// 默认加载每一个插件。如果配置中指定了某插件的启用状态，则按配置的来。
	for id, plugin := range plugins {
		if enable, ok := cfg.Plugin.Enable[id]; ok && enable || !ok {
			plugin.AttachTo(engine)
			plugin.convertMapToConfig(cfg.Plugin.Config[id])
			plugin.onInit(plugin)
		}
	}
}

func (target *Plugin) convertMapToConfig(src PluginConfig) {
	if target == nil {
		return
	}

	if src == nil {
		return
	}

	cfg := &target.Config
	value := reflect.ValueOf(cfg).Elem().Elem()
	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("插件配置必须是结构体指针类型，而不是 %T(%v)", target.Config, value.Kind()))
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		panic(fmt.Errorf("插件配置必须是结构体指针类型，而不是 %v", value.Kind()))
	}

	mapToStruct(src, value.Addr().Interface())
}
