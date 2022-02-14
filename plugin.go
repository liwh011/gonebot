package gonebot

import "fmt"

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
	if p, ok := plugins[plugin.Info().Name]; ok {
		panic(fmt.Errorf("插件 %s 已经注册。该插件为 %v", plugin.Info().Name, p))
	}
	plugins[plugin.Info().Name] = plugin
}

func GetPlugin(name string) Plugin {
	return plugins[name]
}

func InitPlugins(engine *Engine) {
	for _, plugin := range plugins {
		plugin.Init(engine)
	}
}
