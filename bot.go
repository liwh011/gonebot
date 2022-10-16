package gonebot

type Bot struct {
	adapter *OneBotAdapter

	selfId int64
}

func (bot *Bot) Init(adapter *OneBotAdapter) {
	bot.adapter = adapter
}

func (bot *Bot) GetSelfId() int64 {
	return bot.selfId
}
