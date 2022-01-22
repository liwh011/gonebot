package bot

import (
	"github.com/liwh011/gonebot/driver"
	"github.com/liwh011/gonebot/event"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Bot struct {
	driver *driver.WebsocketClient

	selfId      int64
	accessToken string
}

func NewBot() *Bot {
	bot := &Bot{}
	bot.driver = driver.NewWsClient("ws://127.0.0.1:6700/", 30)

	return bot
}

func (bot *Bot) Run() {
	go bot.driver.Start()

	// dis := event.NewEventDispatcher()
	ret, err := bot.GetLoginInfo()
	if err != nil {
		log.Error(err)
		return
	}
	bot.selfId = ret.UserId
	bot.processEvent()
}

func (bot *Bot) processEvent() {
	ch := bot.driver.Subscribe()
	for {
		msg := <-ch
		data := gjson.ParseBytes(msg)
		ev := event.FromJsonObject(data)
		if event.GetEventField(ev, "PostType") == "meta_event" {
			log.Debugf("%s", event.GetEventDescription(ev))
		} else {
			log.Infof("%s", event.GetEventDescription(ev))
		}
		// if event, ok := (*ev).(event.PrivateMessageEvent); ok {
		// 	bot.SendPrivateMsg(304180208, event.Message, false)
		// }
	}
}
