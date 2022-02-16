package gonebot

type Bot struct {
	driver *WebsocketClient

	selfId int64
}

func NewBot(d *WebsocketClient, cfg *BaseConfig) *Bot {
	bot := &Bot{}
	bot.driver = d
	return bot
}

func (bot *Bot) Init() {
	// 获取bot自己的qq号
	info, err := bot.GetLoginInfo()
	if err != nil {
		panic(err)
	}
	bot.selfId = info.UserId
}

func (bot *Bot) GetSelfId() int64 {
	return bot.selfId
}
