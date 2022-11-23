package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/liwh011/gonebot"
	"github.com/liwh011/gonebot/providers/websocket/internal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func init() {
	gonebot.RegisterProvider("websocket", &WebsocketClientProvider{})
}

type WebsocketClientProvider struct {
	wsc       *internal.WebsocketClient
	msgChan   chan []byte
	respStore responseStore
	gbConfig  gonebot.Config

	eventRecievers []chan<- gonebot.I_Event
}

type WebsocketConfig struct {
	Host        string `yaml:"host"`         // WebSocket服务器地址
	Port        int    `yaml:"port"`         // WebSocket服务器端口
	AccessToken string `yaml:"access_token"` // 访问令牌，应与WS服务器设定的一致
}

func (p *WebsocketClientProvider) Init(cfg gonebot.Config) {
	p.gbConfig = cfg

	mp, ok := cfg.GetBaseConfig().ProviderConfig["websocket"]
	if !ok {
		log.Panic("未找到websocket配置")
	}
	wsCfg := &WebsocketConfig{}
	err := mp.DecodeTo(wsCfg)
	if err != nil {
		log.Panicf("解析websocket配置失败：%s", err)
	}

	url := fmt.Sprintf("ws://%s:%d/", wsCfg.Host, wsCfg.Port)
	p.msgChan = make(chan []byte)
	p.wsc = internal.NewWebsocketClient(url, wsCfg.AccessToken, p.msgChan)
}

func (p *WebsocketClientProvider) handleMessageLoop() {
	for msg := range p.msgChan {
		jsonData := gjson.ParseBytes(msg)
		if isApiResponse(jsonData) {
			msg := response{
				Status:  jsonData.Get("status").String(),
				Data:    jsonData.Get("data"),
				Msg:     jsonData.Get("msg").String(),
				Wording: jsonData.Get("wording").String(),
				RetCode: jsonData.Get("retcode").Int(),
				Echo:    jsonData.Get("echo").Uint(),
			}
			p.respStore.store(msg)
		} else {
			ev := gonebot.ConvertJsonObjectToEvent(jsonData)
			for _, ch := range p.eventRecievers {
				ch <- ev
			}
		}
	}
}

func (p *WebsocketClientProvider) Start() {
	p.wsc.Start()
	go p.handleMessageLoop()
}

func (p *WebsocketClientProvider) Stop() {
	p.wsc.Stop()
}

func (p *WebsocketClientProvider) Request(route string, data interface{}) (interface{}, error) {
	req := request{
		Action: route,
		Params: data,
		Echo:   p.respStore.getSeqNum(),
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	rspChan := p.respStore.get(req.Echo)
	err = p.wsc.Send(reqBytes)
	if err != nil {
		return nil, err
	}

	apiCallTimeout := p.gbConfig.GetBaseConfig().ApiCallTimeout
	select {
	case rspData := <-rspChan:
		if rspData.RetCode != 0 {
			err = fmt.Errorf("调用API失败：[%d %s]%s", rspData.RetCode, rspData.Msg, rspData.Wording)
		}
		return rspData.Data, err

	case <-time.After(time.Second * time.Duration(apiCallTimeout)):
		err = fmt.Errorf("调用API超时（%d秒）", apiCallTimeout)
		return nil, err
	}
}

func (p *WebsocketClientProvider) RecieveEvent(ch chan<- gonebot.I_Event) {
	p.eventRecievers = append(p.eventRecievers, ch)
}
