package gonebot

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type WebsocketConfig struct {
	Host        string `yaml:"host"`         // WebSocket服务器地址
	Port        int    `yaml:"port"`         // WebSocket服务器端口
	AccessToken string `yaml:"access_token"` // 访问令牌，应与WS服务器设定的一致
}

type Config struct {
	Websocket WebsocketConfig `yaml:"websocket"`

	ApiCallTimeout int `yaml:"apicall_timeout"` // API调用超时时间，单位：秒
}

// 载入配置文件
func LoadConfig(path string) *Config {
	yamlFile, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("打开配置文件失败：%s", err)
		panic(err)
	}
	defer yamlFile.Close()

	var cfg Config
	err = yaml.NewDecoder(yamlFile).Decode(&cfg)
	if err != nil {
		err = fmt.Errorf("解析配置文件失败：%s", err)
		panic(err)
	}

	return &cfg
}

// 载入配置文件
func LoadCustomConfig(path string, cfgPtr interface{}) {
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
