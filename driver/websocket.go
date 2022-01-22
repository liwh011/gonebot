package driver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/gorilla/websocket"
)

type Params map[string]interface{}

type request struct {
	Action string `json:"action"`
	Params Params `json:"params"`
	Echo   uint64 `json:"echo"`
}

type response struct {
	Status  string       `json:"status"` // 'ok' or 'failed'
	Data    gjson.Result `json:"data"`
	Msg     string       `json:"msg"`     // error message
	Wording string       `json:"wording"` // error message in Chinese
	RetCode int64        `json:"retcode"` // error code, 0 for success
	Echo    uint64       `json:"echo"`    // 回复消息的序列号，用于匹配
}

type WebsocketClient struct {
	conn *websocket.Conn
	url  string // websocket服务器地址

	requestChan  chan request  // 发送消息通道
	responseChan chan response // 接收消息通道
	eventChan    chan []byte

	isAlive          bool // 是否连接
	reconnectTimeout int  // 重连超时时间
	apiCallTimeout   int  // 调用API超时时间

	seqNum     uint64 // 消息序号
	seqNumLock sync.Mutex
	seq2Chan   sync.Map // 消息序号到存取通道的映射

	subscribers []subscriber // 订阅者
}

type subscriber struct {
	recvChan chan []byte
}

func NewWsClient(url string, timeout int) *WebsocketClient {
	return &WebsocketClient{
		url:              url,
		conn:             nil,
		requestChan:      make(chan request, 10),
		responseChan:     make(chan response, 10),
		eventChan:        make(chan []byte, 10),
		isAlive:          false,
		apiCallTimeout:   timeout,
		reconnectTimeout: timeout,
		seqNum:           1,
		seqNumLock:       sync.Mutex{},
	}
}

func (wsc *WebsocketClient) connect() {
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
func (wsc *WebsocketClient) sendMsgThread() {
	for wsc.isAlive {
		req := <-wsc.requestChan

		jsonData, err := json.Marshal(req)
		if err != nil {
			log.Errorf("序列化消息失败：%v。消息：%v", err, req)
			wsc.responseChan <- makeErrorResponse(req, err) // 直接返回错误
			continue
		}

		err = wsc.conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Errorf("发送消息到ws服务器时发生错误：%v", err)
			wsc.responseChan <- makeErrorResponse(req, err)
			continue
		}
	}
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

		// log.Debugf("从ws服务器接收到消息：%s", msgBytes)

		jsonData := gjson.ParseBytes(msgBytes)
		if isResponse(jsonData) {
			msg := response{
				Status:  jsonData.Get("status").String(),
				Data:    jsonData.Get("data"),
				Msg:     jsonData.Get("msg").String(),
				Wording: jsonData.Get("wording").String(),
				RetCode: jsonData.Get("retcode").Int(),
				Echo:    jsonData.Get("echo").Uint(),
			}
			wsc.responseChan <- msg
		} else {
			wsc.eventChan <- msgBytes
		}

	}
}

// 分发收到的消息
func (wsc *WebsocketClient) dispatchMessages() {
	for {
		select {
		case msg := <-wsc.responseChan:
			seqChan, ok := wsc.seq2Chan.Load(msg.Echo)
			if !ok {
				log.Errorf("收到未知的序列号：%d", msg.Echo)
				continue
			}
			seqChan.(chan response) <- msg
			close(seqChan.(chan response)) // 该通道只会被使用一次，必须主动关闭通道，否则会出现死锁

		case msg := <-wsc.eventChan:
			for _, sub := range wsc.subscribers {
				select {
				case sub.recvChan <- msg:
				default: // 通道已满，抛弃消息，避免阻塞
					log.Warnf("订阅者通道已满，消息被抛弃：%v", msg)
				}
			}
		}

	}
}

// 订阅消息，返回一个通道，可以从中读取消息
func (wsc *WebsocketClient) Subscribe() chan []byte {
	ch := make(chan []byte, 10)
	wsc.subscribers = append(wsc.subscribers, subscriber{
		recvChan: ch,
	})
	return ch
}

// 开启服务并重连
func (wsc *WebsocketClient) Start() {
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

// 获取序号
func (wsc *WebsocketClient) getSeqNum() uint64 {
	wsc.seqNumLock.Lock()
	defer wsc.seqNumLock.Unlock()
	wsc.seqNum++
	return wsc.seqNum
}

// 调用API。超时时间内没有收到回复，则返回错误
func (wsc *WebsocketClient) CallApi(apiName string, params Params) (rsp response, err error) {
	req := request{
		Action: apiName,
		Params: params,
		Echo:   wsc.getSeqNum(),
	}

	rspChan := make(chan response, 1)
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
			err = fmt.Errorf("调用API失败：[%d %s]%s", rsp.RetCode, rsp.Msg, rsp.Wording)
		}
		return

	case <-time.After(time.Second * time.Duration(wsc.apiCallTimeout)):
		err = fmt.Errorf("调用API超时（%d秒）", wsc.apiCallTimeout)
		return
	}
}
