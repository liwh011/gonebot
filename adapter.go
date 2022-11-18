package gonebot

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type request struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   uint64      `json:"echo"`
}

type response struct {
	Status  string       `json:"status"` // 'ok' or 'failed'
	Data    gjson.Result `json:"data"`
	Msg     string       `json:"msg"`     // error message
	Wording string       `json:"wording"` // error message in Chinese
	RetCode int64        `json:"retcode"` // error code, 0 for success
	Echo    uint64       `json:"echo"`    // 回复消息的序列号，用于匹配
}

type OneBotAdapter struct {
	provider Provider

	seqNum     uint64 // 消息序号
	seqNumLock sync.Mutex
	seq2Chan   sync.Map // 消息序号到存取通道的映射

	eventRecievers []chan<- I_Event

	apiCallTimeout int
}

func (adapter *OneBotAdapter) Init(cfg Config, provider Provider) {
	adapter.provider = provider
	adapter.seqNum = 0
	adapter.seqNumLock = sync.Mutex{}
	adapter.seq2Chan = sync.Map{}
	adapter.eventRecievers = make([]chan<- I_Event, 0)
	adapter.apiCallTimeout = cfg.GetBaseConfig().ApiCallTimeout

	ch := make(chan []byte)
	adapter.provider.Recieve(ch)
	go func() {
		for data := range ch {
			adapter.handleMessage(data)
		}
	}()
}

func (adapter *OneBotAdapter) handleMessage(data []byte) {
	jsonData := gjson.ParseBytes(data)
	if isApiResponse(jsonData) {
		msg := response{
			Status:  jsonData.Get("status").String(),
			Data:    jsonData.Get("data"),
			Msg:     jsonData.Get("msg").String(),
			Wording: jsonData.Get("wording").String(),
			RetCode: jsonData.Get("retcode").Int(),
			Echo:    jsonData.Get("echo").Uint(),
		}
		if ch, ok := adapter.seq2Chan.Load(msg.Echo); ok {
			ch.(chan response) <- msg
			close(ch.(chan response))
		} else {
			log.Warnf("没有找到对应的回复通道，消息序号为%d", msg.Echo)
		}
	} else {
		ev := ConvertJsonObjectToEvent(jsonData)
		for _, ch := range adapter.eventRecievers {
			ch <- ev
		}
	}
}

func isApiResponse(msg gjson.Result) bool {
	return msg.Get("status").Exists() && msg.Get("retcode").Exists()
}

func (adapter *OneBotAdapter) RecieveEvent(ch chan<- I_Event) {
	adapter.eventRecievers = append(adapter.eventRecievers, ch)
}

func (adapter *OneBotAdapter) getSeqNum() uint64 {
	adapter.seqNumLock.Lock()
	defer adapter.seqNumLock.Unlock()
	adapter.seqNum++
	return adapter.seqNum
}

func (adapter *OneBotAdapter) Request(route string, data interface{}) (rsp interface{}, err error) {
	req := request{
		Action: route,
		Params: data,
		Echo:   adapter.getSeqNum(),
	}

	rspChan := make(chan response, 1)
	adapter.seq2Chan.Store(req.Echo, rspChan)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, err = adapter.provider.Send(reqBytes)
	if err != nil {
		return nil, err
	}

	select {
	case rspData := <-rspChan:
		rsp = rspData
		if rspData.RetCode != 0 {
			err = fmt.Errorf("调用API失败：[%d %s]%s", rspData.RetCode, rspData.Msg, rspData.Wording)
		}
		return

	case <-time.After(time.Second * time.Duration(adapter.apiCallTimeout)):
		err = fmt.Errorf("调用API超时（%d秒）", adapter.apiCallTimeout)
		return
	}
}
