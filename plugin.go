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

type Plugin interface {
	Init(*Engine)
	Info() PluginInfo
}

var plugins = make(map[string]Plugin)

func RegisterPlugin(plugin Plugin) {
	if plugin == nil {
		panic("插件不能为nil")
	}

	info := plugin.Info()
	id := fmt.Sprintf("%s@%s", info.Name, info.Author)
	if p, ok := plugins[id]; ok {
		panic(fmt.Errorf("插件 %s 已经注册。该插件为 %v", id, p.Info()))
	}

	plugins[id] = plugin
}

func GetPlugin(id string) Plugin {
	return plugins[id]
}

func InitPlugins(engine *Engine) {
	cfg := engine.Config

	// 默认加载每一个插件。如果配置中指定了某插件的启用状态，则按配置的来。
	for id, plugin := range plugins {
		if enable, ok := cfg.Plugin.Enable[id]; ok && enable || !ok {
			fillPluginConfigIntoStruct(plugin, cfg.Plugin.Config[id])
			plugin.Init(engine)
		}
	}
}

func fillPluginConfigIntoStruct(plugin Plugin, cfg PluginConfig) {
	if plugin == nil {
		return
	}

	if cfg == nil {
		return
	}

	value := reflect.ValueOf(plugin)
	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("传入了不是指针的参数，类型：%T（需要传递指针）", plugin))
	}

	value = value.Elem()
	cfgField := value.FieldByName("Config")
	if !cfgField.IsValid() {
		return
	}

	if cfgField.Kind() != reflect.Struct {
		return
	}

	mapToStruct(cfg, cfgField.Addr().Interface())

}
