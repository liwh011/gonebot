package gonebot

import (
	"fmt"
	"testing"
)

func Test_convertMapToConfig(t *testing.T) {
	info := PluginInfo{
		Name:        "test",
		Description: "test",
		Version:     "test",
		Author:      "test",
	}

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
	p := NewPlugin(info, &cfg, nil)
	p.convertMapToConfig(PluginConfig{
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
	fmt.Println(p.GetConfig())
}
