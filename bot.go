package gonebot

type Bot struct {
	provider Provider

	selfId int64
}

func (bot *Bot) Init(provider Provider) {
	bot.provider = provider
}

func (bot *Bot) GetSelfId() int64 {
	return bot.selfId
}
