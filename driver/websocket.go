package driver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type Params map[string]interface{}

type request struct {
	Action string               `json:"action"`
	Params Params               `json:"params"`
	Echo   struct{ seq uint64 } `json:"echo"`
}

type message struct {
	Status  string                 `json:"status"` // 'ok' or 'failed'
	Data    map[string]interface{} `json:"data"`
	Msg     string                 `json:"msg"`     // error message
	Wording string                 `json:"wording"` // error message in Chinese
	RetCode int64                  `json:"retcode"` // error code, 0 for success
	Echo    struct{ seq uint64 }   `json:"echo"`    // 回复消息的序列号，用于匹配。如果消息不是回复，则为0
}

type websocketClient struct {
	conn *websocket.Conn
	url  string // websocket服务器地址

	requestChan chan request // 发送消息通道
	receiveChan chan message // 接收消息通道

	isAlive          bool // 是否连接
	reconnectTimeout int  // 重连超时时间
	apiCallTimeout   int  // 调用API超时时间

	seqNum   uint64   // 消息序号
	seq2Chan sync.Map // 消息序号到存取通道的映射

	subscribers []subscriber // 订阅者
}

type subscriber struct {
	filterFunc func(msg message) bool
	recvChan   chan message
}

func NewWsClient(url string, timeout int) *websocketClient {
	return &websocketClient{
		url:              url,
		conn:             nil,
		requestChan:      make(chan request, 10),
		receiveChan:      make(chan message, 10),
		isAlive:          false,
		apiCallTimeout:   timeout,
		reconnectTimeout: timeout,
		seqNum:           1,
	}
}

func (wsc *websocketClient) connect() {
	var err error

	log.Infof("正在连接到Websocket服务器：%s", wsc.url)
	wsc.conn, _, err = websocket.DefaultDialer.Dial(wsc.url, nil)
	if err != nil {
		log.Errorf("连接到Websocket服务器失败：%v", err)
		return
	}

	wsc.isAlive = true
	log.Info("连接到Websocket服务器成功")
}

// 发送消息
func (wsc *websocketClient) sendMsgThread() {
	for wsc.isAlive {
		req := <-wsc.requestChan
		req.Echo.seq = wsc.getSeqNum()
		jsonData, err := json.Marshal(req)
		if err != nil {
			log.Errorf("序列化消息失败：%v。消息：%v", err, req)
			wsc.receiveChan <- makeErrorResponse(req, err) // 直接返回错误
			continue
		}

		err = wsc.conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Errorf("发送消息到ws服务器时发生错误：%v", err)
			wsc.receiveChan <- makeErrorResponse(req, err)
			continue
		}
	}
}

// 读取消息
func (wsc *websocketClient) readMsgThread() {
	for wsc.isAlive {
		_, msgBytes, err := wsc.conn.ReadMessage()
		if err != nil {
			log.Errorf("从ws服务器读取消息时发生错误：%v", err)
			wsc.isAlive = false
			break
		}

		log.Debugf("从ws服务器接收到消息：%s", msgBytes)
		var msg message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			log.Errorf("消息无法被解码：", err)
		}
		wsc.receiveChan <- msg
	}
}

// 分发收到的消息
func (wsc *websocketClient) dispatchMessages() {
	for {
		msg := <-wsc.receiveChan
		if isReplyMessage(msg) { // 是回复消息，则分发给对应的请求消息
			seqChan, ok := wsc.seq2Chan.Load(msg.Echo.seq)
			if !ok {
				log.Errorf("收到未知的序列号：%d", msg.Echo.seq)
				continue
			}
			seqChan.(chan message) <- msg
			close(seqChan.(chan message)) // 该通道只会被使用一次，必须主动关闭通道，否则会出现死锁

		} else { // 消息不是回复消息，则分发给订阅者
			for _, sub := range wsc.subscribers {
				if !sub.filterFunc(msg) {
					continue
				}
				select {
				case sub.recvChan <- msg:
				default: // 通道已满，抛弃消息，避免阻塞
					log.Errorf("订阅者通道已满，消息被抛弃：%v", msg)
				}
			}
		}

	}
}

// 订阅消息，返回一个通道，可以从中读取消息
func (wsc *websocketClient) Subscribe(filterFunc func(msg message) bool) chan message {
	ch := make(chan message, 10)
	wsc.subscribers = append(wsc.subscribers, subscriber{
		filterFunc: filterFunc,
		recvChan:   ch,
	})
	return ch
}

// 开启服务并重连
func (wsc *websocketClient) Start() {
	go wsc.dispatchMessages()
	for {
		if !wsc.isAlive {
			wsc.connect()
			go wsc.sendMsgThread()
			go wsc.readMsgThread()
		}
		time.Sleep(time.Second * time.Duration(wsc.reconnectTimeout))
	}
}

// 获取序号。该函数只有一个调用者，不需要加锁
func (wsc *websocketClient) getSeqNum() uint64 {
	wsc.seqNum++
	return wsc.seqNum
}

// 调用API。超时时间内没有收到回复，则返回错误
func (wsc *websocketClient) CallApi(apiName string, params Params) (rsp message, err error) {
	req := request{
		Action: apiName,
		Params: params,
	}

	rspChan := make(chan message, 1)
	wsc.seq2Chan.Store(req.Echo, rspChan)
	wsc.requestChan <- req

	select {
	case rspData, ok := <-rspChan:
		if !ok {
			err = fmt.Errorf("rspChan被意外关闭")
			return
		}
		rsp = rspData
		if rsp.RetCode != 0 {
			err = fmt.Errorf("调用API失败：%s", rsp.Msg)
		}
		return

	case <-time.After(time.Second * time.Duration(wsc.apiCallTimeout)):
		err = fmt.Errorf("调用API超时（%d秒）", wsc.apiCallTimeout)
		return
	}
}
