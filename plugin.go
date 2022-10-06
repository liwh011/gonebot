package gonebot

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

func init() {
	defaultPluginManager = newPluginManager()
	// 添加钩子，Engine创建完毕后，初始化插件
	EngineHookManager.OnCreated(defaultPluginManager.InitPlugins)
}

// TODO 运行时装卸

// 插件管理器，理应是个单例
type pluginManager struct {
	plugins             map[string]Plugin
	pluginConfigStructs map[string]interface{}
}

var defaultPluginManager *pluginManager

func newPluginManager() *pluginManager {
	pm := &pluginManager{
		plugins:             make(map[string]Plugin),
		pluginConfigStructs: make(map[string]interface{}),
	}

	return pm
}

func (pm *pluginManager) RegisterPlugin(plugin Plugin, pluginConfig interface{}) {
	if plugin == nil {
		panic("插件不能为空")
	}
	id := getPluginId(plugin)
	pm.plugins[id] = plugin

	if pluginConfig != nil {
		pm.pluginConfigStructs[id] = pluginConfig
	}
}

// 初始化插件
func (pm *pluginManager) InitPlugins(engine *Engine) {
	pluginControlConfig := engine.Config.Plugin.Enable

	for id, plugin := range pm.plugins {
		// 仅当配置中指定为禁用的插件才不加载。配置中未指定的插件默认启用
		if enabled, ok := pluginControlConfig[id]; ok && !enabled {
			continue
		}

		// 填充插件配置结构体字段
		if plgCfgStruct, ok := pm.pluginConfigStructs[id]; ok {
			convertConfigMapToStruct(plgCfgStruct, engine.Config.Plugin.Config[id])
		}

		proxy := newEngineProxy(engine)
		plugin.Init(&proxy)
		log.Infof("插件%s加载完毕", id)
	}
}

func (pm *pluginManager) GetPluginById(id string) Plugin {
	if ret, ok := pm.plugins[id]; ok {
		return ret
	} else {
		return nil
	}
}

func (pm *pluginManager) GetPlugin(name, author string) Plugin {
	id := fmt.Sprintf("%s@%s", name, author)
	return pm.GetPluginById(id)
}

func (pm *pluginManager) GetPluginConfig(plugin Plugin) interface{} {
	if ret, ok := pm.pluginConfigStructs[getPluginId(plugin)]; ok {
		return ret
	} else {
		return nil
	}
}

// 注册插件。插件配置结构体pluginConfig可选，使用反射映射到字段，无则传nil
func RegisterPlugin(plugin Plugin, pluginConfig interface{}) {
	defaultPluginManager.RegisterPlugin(plugin, pluginConfig)
}

// 获取插件配置
func GetPluginConfig(plugin Plugin) interface{} {
	return defaultPluginManager.GetPluginConfig(plugin)
}

// 插件信息，其中Name和Author共同唯一标识一个插件
type PluginInfo struct {
	Name        string
	Author      string
	Version     string
	Description string
}

type Plugin interface {
	Init(engine *EngineProxy) // 初始化插件
	GetPluginInfo() PluginInfo
}

// 获取插件的唯一标识，格式为：“插件名@作者”
func getPluginId(plugin Plugin) string {
	info := plugin.GetPluginInfo()
	return fmt.Sprintf("%s@%s", info.Name, info.Author)
}

// Engine对象的代理，仅提供有限的访问。
type EngineProxy struct {
	engine        *Engine
	pluginHandler *Handler
}

func newEngineProxy(engine *Engine) EngineProxy {
	ret := EngineProxy{engine: engine}
	ret.pluginHandler, _ = engine.NewRemovableHandler()
	return ret
}

// 新建一个Handler，用于处理指定类型的事件，不写则处理所有类型的事件
func (p *EngineProxy) NewHandler(eventTypes ...EventName) *Handler {
	return p.pluginHandler.NewHandler(eventTypes...)
}

func (p *EngineProxy) GetEngineConfig() *BaseConfig {
	return p.engine.Config
}

func (p *EngineProxy) GetEngine() *Engine {
	return p.engine
}

func (p *EngineProxy) GetBot() *Bot {
	return p.engine.bot
}

func convertConfigMapToStruct(cfgStruct interface{}, srcMap PluginConfigMap) {
	if srcMap == nil {
		return
	}

	if cfgStruct == nil {
		panic("cfgStruct不能为空")
	}

	// cfg := &target.Config
	value := reflect.ValueOf(cfgStruct)

	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("必须传入结构体指针，而不是 %T (%v)", cfgStruct, value.Kind()))
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		panic(fmt.Errorf("必须传入结构体指针，而不是 %v", value.Kind()))
	}

	mapToStruct(srcMap, value.Addr().Interface())
}
