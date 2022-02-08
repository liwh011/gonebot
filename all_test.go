package gonebot

// import (
// 	"fmt"
// 	"testing"
// )

// func Test_Run(t *testing.T) {
// 	cfg := Config{
// 		WsHost:         "127.0.0.1",
// 		WsPort:         6700,
// 		ApiCallTimeout: 10,
// 		AccessToken:    "",
// 	}
// 	engine := NewEngine(&cfg)

// 	sv := engine.NewService("test")
// 	sv.
// 		NewHandler().
// 		Use(OnlyToMe(), StartsWith("哼哼")).
// 		Handle(func(c *Context, a *Action) {
// 			c.Reply("好丑啊")
// 			fmt.Printf("%+v", c)
// 		}, EVENTNAME_MESSAGE)

// 	engine.Run()
// }
