package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/liwh011/gonebot"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

func init() {
	gonebot.RegisterProvider("websocket", &WebsocketClient{})
}

type WebsocketClient struct {
	conn        *websocket.Conn
	url         string // websocket服务器地址
	accessToken string

	isAlive          bool // 是否连接
	reconnectTimeout int  // 重连超时时间
	shouldReconnect  bool // 是否需要重连

	subscribers []chan<- []byte // 订阅者
}

type WebsocketConfig struct {
	Host        string `yaml:"host"`         // WebSocket服务器地址
	Port        int    `yaml:"port"`         // WebSocket服务器端口
	AccessToken string `yaml:"access_token"` // 访问令牌，应与WS服务器设定的一致
}

func (wsc *WebsocketClient) Init(cfg gonebot.Config) {
	wsCfg := &WebsocketConfig{}
	mp, ok := cfg.GetBaseConfig().ProviderConfig["websocket"]
	if !ok {
		log.Panic("未找到websocket配置")
	}
	err := mp.DecodeTo(wsCfg)
	if err != nil {
		log.Panic(err)
	}

	url := fmt.Sprintf("ws://%s:%d/", wsCfg.Host, wsCfg.Port)

	wsc.conn = nil
	wsc.url = url
	wsc.accessToken = wsCfg.AccessToken

	wsc.isAlive = false
	wsc.reconnectTimeout = 3
	wsc.shouldReconnect = true
	wsc.subscribers = []chan<- []byte{}

}

func (wsc *WebsocketClient) connect() {
	var err error

	log.Infof("正在连接到Websocket服务器：%s", wsc.url)
	header := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", wsc.accessToken)},
	}
	wsc.conn, _, err = websocket.DefaultDialer.Dial(wsc.url, header)
	if err != nil {
		log.Errorf("连接到Websocket服务器失败：%v", err)
		return
	}

	wsc.isAlive = true
	log.Info("连接到Websocket服务器成功")
}

// 读取消息
func (wsc *WebsocketClient) readMsgThread() {
	for wsc.isAlive {
		_, msgBytes, err := wsc.conn.ReadMessage()
		if err != nil {
			log.Errorf("从ws服务器读取消息时发生错误：%v", err)
			wsc.isAlive = false
			break
		}

		for _, sub := range wsc.subscribers {
			sub <- msgBytes
		}
	}
}

func (wsc *WebsocketClient) Send(data []byte) (interface{}, error) {
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return nil, err
}

// 订阅消息，返回一个通道，可以从中读取消息
func (wsc *WebsocketClient) Recieve(sub chan<- []byte) {
	wsc.subscribers = append(wsc.subscribers, sub)
}

// 开启服务并重连
func (wsc *WebsocketClient) Start() {
	go func() {
		// 启动连接到WebSocket服务器
		log.Info("开始连接到WebSocket服务器，地址：", wsc.url)
		for {
			if !wsc.isAlive {
				if wsc.shouldReconnect {
					wsc.connect()
					go wsc.readMsgThread()
				} else {
					break
				}
			}
			time.Sleep(time.Second * time.Duration(wsc.reconnectTimeout))
		}
	}()
}

func (wsc *WebsocketClient) Stop() {
	log.Info("正在断开与WebSocket服务器的连接")
	wsc.shouldReconnect = false
	wsc.conn.Close()
	for _, sub := range wsc.subscribers {
		close(sub)
	}
}
