package bot

import (
	"github.com/liwh011/gonebot/config"
	"github.com/liwh011/gonebot/driver"
)

type Bot struct {
	driver *driver.WebsocketClient

	selfId      int64
	accessToken string
}

func NewBot(d *driver.WebsocketClient, cfg *config.Config) *Bot {
	bot := &Bot{}
	bot.driver = d
	bot.accessToken = cfg.AccessToken
	return bot
}

func (bot *Bot) Init() {
	info, err := bot.GetLoginInfo()
	if err != nil {
		panic(err)
	}
	bot.selfId = info.UserId
}

func (bot *Bot) GetSelfId() int64 {
	return bot.selfId
}
