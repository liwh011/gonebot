package onebotmock

import (
	"encoding/json"

	"github.com/liwh011/gonebot"
	"github.com/liwh011/gonebot/mock"
	"github.com/tidwall/gjson"
)

type OnebotMock struct {
	recievers  []chan<- []byte
	mockServer *mock.MockServer
}

func NewOnebotMock(mockServer *mock.MockServer) *OnebotMock {
	return &OnebotMock{
		mockServer: mockServer,
	}
}

func (o *OnebotMock) Init(cfg gonebot.Config) {

}

func (o *OnebotMock) Start() {
	ch := make(chan gonebot.I_Event)
	go func() {
		for {
			event := <-ch
			data, _ := json.Marshal(event)
			for _, reciever := range o.recievers {
				reciever <- data
			}
		}
	}()
	o.mockServer.RegisterEventReciever(ch)
}

func (o *OnebotMock) Stop() {

}

func (o *OnebotMock) Send(data []byte) (interface{}, error) {
	req := gjson.ParseBytes(data)
	action := req.Get("action").String()
	params := req.Get("params")
	result, err := o.mockServer.HandleRequest(action, params)

	var resp map[string]interface{}
	if err == nil {
		resp = map[string]interface{}{
			"status":  "ok",
			"retcode": 0,
			"data":    result,
			"echo":    req.Get("echo").Value(),
		}
	} else {
		resp = map[string]interface{}{
			"status":  "failed",
			"retcode": 114514,
			"msg":     err.Error(),
			"wording": err.Error(),
			"echo":    req.Get("echo").Value(),
		}
	}
	respJson, _ := json.Marshal(resp)

	for _, reciever := range o.recievers {
		reciever <- respJson
	}

	return nil, nil
}

func (o *OnebotMock) Recieve(ch chan<- []byte) {
	o.recievers = append(o.recievers, ch)
}
