package gonebot

// import (
// 	"testing"

// 	"github.com/liwh011/gonebot/config"
// 	"github.com/liwh011/gonebot/event"
// 	"github.com/liwh011/gonebot/handler"
// )

// func Test_Run(t *testing.T) {
// 	cfg := config.Config{
// 		WsHost:         "127.0.0.1",
// 		WsPort:         6700,
// 		ApiCallTimeout: 10,
// 		AccessToken:    "",
// 	}
// 	sv := NewService("test")
// 	sv.On(event.EVENT_NAME_PRIVATE_MESSAGE).SetHandler(func(c *handler.Context) {
// 		c.Reply(message.MustJoin(c.Event.GetEventDescription()))
// 	})
// 	sv.OnStartsWith("皓哥哼哼哼").SetHandler(func(c *handler.Context) {
// 		// c.Reply(message.MustJoin("哈哈"))
// 		c.Bot.SendGroupMsg(c.Event.(*event.GroupMessageEvent).GroupId, message.MustJoin("哈哈"), false)
// 	})

// 	Run(&cfg)
// }
