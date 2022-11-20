package internal

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	conn        *websocket.Conn
	url         string // websocket服务器地址
	accessToken string

	reconnectTimeout int // 重连超时时间
	closeSignal      chan bool

	recieveChan chan<- []byte
}

func NewWebsocketClient(url string, accessToken string, recieveChan chan<- []byte) *WebsocketClient {
	return &WebsocketClient{
		url:              url,
		accessToken:      accessToken,
		reconnectTimeout: 3,
		closeSignal:      make(chan bool),
		recieveChan:      recieveChan,
	}
}

func (wsc *WebsocketClient) connect() error {
	var err error

	header := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", wsc.accessToken)},
	}
	wsc.conn, _, err = websocket.DefaultDialer.Dial(wsc.url, header)
	if err != nil {
		return err
	}
	return nil
}

func (wsc *WebsocketClient) readMessage(msgChan chan<- []byte, errChan chan<- error) {
	_, msgBytes, err := wsc.conn.ReadMessage()
	if err != nil {
		errChan <- err
	} else {
		msgChan <- msgBytes
	}
}

// 读取消息，手动关闭时不会返回错误，其他情况下会返回错误
func (wsc *WebsocketClient) readMsgLoop() error {
	for {
		msgChan := make(chan []byte)
		errChan := make(chan error)
		go wsc.readMessage(msgChan, errChan)

		select {
		case msg := <-msgChan:
			wsc.recieveChan <- msg
		case err := <-errChan:
			wsc.closeSignal <- true
			return err

		case <-wsc.closeSignal:
			return nil
		}
	}
}

func (wsc *WebsocketClient) Send(data []byte) error {
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return err
}

// 开启服务并重连
func (wsc *WebsocketClient) Start() {
	go func() {
		// 启动连接到WebSocket服务器
		for {
			log.Infof("正在连接到Websocket服务器：%s", wsc.url)
			err := wsc.connect()
			if err != nil {
				log.Errorf("连接到WebSocket服务器失败：%v", err)
				time.Sleep(time.Duration(wsc.reconnectTimeout) * time.Second)
				continue
			}
			log.Info("连接到Websocket服务器成功")

			err = wsc.readMsgLoop()
			if err != nil {
				log.Errorf("读取消息失败：%v", err)
				time.Sleep(time.Duration(wsc.reconnectTimeout) * time.Second)
			} else {
				// 正常关闭
				close(wsc.recieveChan)
				wsc.conn.Close()
				log.Info("与WebSocket服务器的连接已断开")
				return
			}
		}
	}()
}

func (wsc *WebsocketClient) Stop() {
	log.Info("正在断开与WebSocket服务器的连接")
	wsc.closeSignal <- true
	// for _, sub := range wsc.subscribers {
	// 	close(sub)
	// }
}
