package bot

import (
	"github.com/liwh011/gonebot/driver"
)

type Bot struct {
	driver *driver.WebsocketClient

	selfId      int64
	accessToken string
}

func NewBot(d *driver.WebsocketClient) *Bot {
	bot := &Bot{}
	bot.driver = d

	return bot
}

// func (bot *Bot) Run() {
// 	go bot.driver.Start()

// 	// dis := event.NewEventDispatcher()
// 	ret, err := bot.GetLoginInfo()
// 	if err != nil {
// 		log.Error(err)
// 		return
// 	}
// 	bot.selfId = ret.UserId
// 	bot.processEvent()
// }

// func (bot *Bot) processEvent() {
// 	ch := bot.driver.Subscribe()
// 	for {
// 		msg := <-ch
// 		data := gjson.ParseBytes(msg)
// 		ev := event.FromJsonObject(data)
// 		if ev.GetPostType() == event.POST_TYPE_META {
// 			log.Debug(ev.GetEventDescription())
// 		} else {
// 			log.Info(ev.GetEventDescription())
// 		}
// 		// if event, ok := (*ev).(event.PrivateMessageEvent); ok {
// 		// 	bot.SendPrivateMsg(304180208, event.Message, false)
// 		// }
// 	}
// }
