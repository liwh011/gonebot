package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	WsHost         string `json:"ws_host"`          // WebSocket服务器地址
	WsPort         int    `json:"ws_port"`          // WebSocket服务器端口
	ApiCallTimeout int    `json:"api_call_timeout"` // API调用超时时间
	AccessToken    string `json:"access_token"`     // 访问令牌，应与WS服务器设定的一致
}

// 载入配置文件
func LoadConfig(path string) *Config {
	jsonFile, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("打开配置文件失败：%s", err)
		panic(err)
	}
	defer jsonFile.Close()

	var config Config
	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		err = fmt.Errorf("解析配置文件失败：%s", err)
		panic(err)
	}
	return &config
}
