package gonebot

import (
	"fmt"
	"testing"
)

type TestPlugin struct {
	Config struct {
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
}

func (p TestPlugin) Info() PluginInfo {
	return PluginInfo{
		Name:        "test",
		Description: "test",
		Version:     "test",
		Author:      "test",
	}
}

func (p TestPlugin) Init(engine *Engine) {}

func Test_fillPluginConfigIntoStruct(t *testing.T) {
	p := TestPlugin{}
	fillPluginConfigIntoStruct(&p, PluginConfig{
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
	fmt.Println(p.Config)
}
