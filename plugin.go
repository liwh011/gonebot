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
	id     string
	Info   PluginInfo
	config interface{}
	onInit func(*Engine)
}

// 获取插件配置，返回指针
func (p *Plugin) GetConfig() interface{} {
	return p.config
}

var plugins = make(map[string]*Plugin)

// 注册一个新插件。
//
// info: 插件信息，包括名称、描述、版本、作者
//
// cfg: 插件配置，传入结构体指针，如果无需配置可以为 nil
//
// onInit: 插件初始化时的回调函数，如果无需初始化可以为 nil
func NewPlugin(info PluginInfo, cfg interface{}, onInit func(*Engine)) *Plugin {
	np := &Plugin{
		id:     fmt.Sprintf("%s@%s", info.Name, info.Author),
		Info:   info,
		config: cfg,
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

func InitPlugins(engine *Engine) {
	cfg := engine.Config

	// 默认加载每一个插件。如果配置中指定了某插件的启用状态，则按配置的来。
	for id, plugin := range plugins {
		if enable, ok := cfg.Plugin.Enable[id]; ok && enable || !ok {
			plugin.convertMapToConfig(cfg.Plugin.Config[id])
			plugin.onInit(engine)
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

	cfg := &target.config
	value := reflect.ValueOf(cfg).Elem().Elem()
	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("插件配置必须是结构体指针类型，而不是 %T(%v)", target.config, value.Kind()))
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		panic(fmt.Errorf("插件配置必须是结构体指针类型，而不是 %v", value.Kind()))
	}

	mapToStruct(src, value.Addr().Interface())
}
