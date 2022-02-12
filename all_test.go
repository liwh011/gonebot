package gonebot

// import (
// 	"fmt"
// 	"testing"
// )

// func Test_Run(t *testing.T) {
// 	cfg := BaseConfig{
// 		Websocket: WebsocketConfig{
// 			Host:           "127.0.0.1",
// 			Port:           6700,
// 			AccessToken:    "",
// 			ApiCallTimeout: 10,
// 		},
// 	}
// 	engine := NewEngine(&cfg)

// 	sv := engine.NewService("test")
// 	sv.
// 		NewHandler(EventNameGroupMessage).
// 		Use(OnlyToMe(), StartsWith("哼哼")).
// 		Handle(func(c *Context, a *Action) {
// 			c.Reply("好丑啊")
// 			fmt.Printf("%+v", c)
// 		})

// 	engine.Run()
// }
