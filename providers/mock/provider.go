package onebotmock

import (
	"encoding/json"

	"github.com/liwh011/gonebot"
	"github.com/liwh011/gonebot/mock"
	"github.com/tidwall/gjson"
)

type MockProvider struct {
	recievers  []chan<- gonebot.I_Event
	mockServer *mock.MockServer
}

func NewMockProvider(mockServer *mock.MockServer) *MockProvider {
	return &MockProvider{
		mockServer: mockServer,
	}
}

func (o *MockProvider) Init(cfg gonebot.Config) {

}

func (o *MockProvider) Start() {
	ch := make(chan gonebot.I_Event)
	go func() {
		for event := range ch {
			for _, reciever := range o.recievers {
				reciever <- event
			}
		}
	}()
	o.mockServer.RegisterEventReciever(ch)
}

func (o *MockProvider) Stop() {

}

func (o *MockProvider) Request(route string, data interface{}) (interface{}, error) {
	dataStr, _ := json.Marshal(data)
	result, err := o.mockServer.HandleRequest(route, gjson.ParseBytes(dataStr))
	if err != nil {
		return nil, err
	}

	resultStr, _ := json.Marshal(result)
	ret := gjson.ParseBytes(resultStr)
	return &ret, nil
}

func (o *MockProvider) Recieve(ch chan<- gonebot.I_Event) {
	o.recievers = append(o.recievers, ch)
}
