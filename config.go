package gonebot

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type Config interface {
	GetBaseConfig() *BaseConfig
}

type PluginConfigMap map[string]interface{}   // 插件配置
type ProviderConfigMap map[string]interface{} // 服务提供者配置
type BaseConfig struct {
	// Websocket WebsocketConfig `yaml:"websocket"`
	CmdPrefix      []string `yaml:"cmd_prefix"`      // 命令前缀
	Superuser      []int64  `yaml:"superuser"`       // 超级用户
	ApiCallTimeout int      `yaml:"apicall_timeout"` // API调用超时时间，单位：秒
	Plugin         struct {
		Enable map[string]bool            `yaml:"enable"`
		Config map[string]PluginConfigMap `yaml:"config"`
	} `yaml:"plugin"`
	Provider       string                       `yaml:"provider"`
	ProviderConfig map[string]ProviderConfigMap `yaml:"provider_config"`
}

func (mp ProviderConfigMap) DecodeTo(v interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		MatchName: func(mapKey, fieldName string) bool {
			switch {
			case mapKey == camelCaseToSnakeCase(fieldName):
				return true
			case mapKey == fieldName:
				return true
			case mapKey == strings.ToLower(fieldName[:1])+fieldName[1:]:
				return true
			}
			return false
		},
		Result: v,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(mp)
}

func (cfg *BaseConfig) GetBaseConfig() *BaseConfig {
	return cfg
}

// 载入配置文件
func LoadConfig(path string) *BaseConfig {
	yamlFile, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("打开配置文件失败：%s", err)
		panic(err)
	}
	defer yamlFile.Close()

	var cfg BaseConfig
	err = yaml.NewDecoder(yamlFile).Decode(&cfg)
	if err != nil {
		err = fmt.Errorf("解析配置文件失败：%s", err)
		panic(err)
	}

	return &cfg
}

// 载入自定义配置文件，文件为YAML。自定义配置的结构体必须继承BaseConfig
func LoadCustomConfig(path string, cfgPtr Config) {
	yamlFile, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("打开配置文件失败：%s", err)
		panic(err)
	}
	defer yamlFile.Close()

	err = yaml.NewDecoder(yamlFile).Decode(cfgPtr)
	if err != nil {
		err = fmt.Errorf("解析配置文件失败：%s", err)
		panic(err)
	}
}
