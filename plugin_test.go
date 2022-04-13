package gonebot

import (
	"fmt"
	"testing"
)

func Test_convertMapToConfig(t *testing.T) {
	type Config struct {
		Id     int64
		Name   string
		Enable bool
		Obj    *struct {
			F *bool
		}

		AbcAbc []string
		AaaAaa map[string]int64
		BbbBbb string `json:"bbb"`

		Ccc []struct {
			Aaa map[string]int64
		}
	}
	cfg := Config{}

	// p := TestPlugin{}
	convertConfigMapToStruct(&cfg, PluginConfigMap{
		"id":     1,
		"name":   "test",
		"enable": true,
		"obj": map[string]interface{}{
			"f": true,
		},
		"abc_abc": []interface{}{"a", "b", "c"},
		"aaaAaa":  map[string]interface{}{"a": 1, "b": 2},
		"bbb":     "bbb",

		"ccc": []interface{}{
			map[string]interface{}{
				"aaa": map[string]interface{}{"a": 1, "b": 2},
			},
		},
	})
	fmt.Println(cfg)
}

type TestPlugin struct{}

func (p *TestPlugin) Init(proxy *EngineProxy) {
	fmt.Println("init")
	proxy.NewHandler(EventNamePrivateMessage).
		Handle(func(c *Context, a *Action) {
			c.Reply("好丑啊")
		})
}

func (p *TestPlugin) GetPluginInfo() PluginInfo {
	return PluginInfo{
		Name:        "test",
		Author:      "test",
		Description: "test",
		Version:     "test",
	}
}

func Test_pluginInit(t *testing.T) {
	RegisterPlugin(&TestPlugin{}, nil)
	cfg := BaseConfig{
		Websocket: WebsocketConfig{
			Host:           "127.0.0.1",
			Port:           6700,
			AccessToken:    "asdsss",
			ApiCallTimeout: 10,
		},
	}
	cfg.Plugin.Enable = make(map[string]bool)
	cfg.Plugin.Enable["test@test"] = false
	engine := NewEngine(&cfg)
	engine.Run()
}
