package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	WsHost         string `json:"ws_host"`
	WsPort         int    `json:"ws_port"`
	ApiCallTimeout int    `json:"api_call_timeout"`
	AccessToken    string `json:"access_token"`
}

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
