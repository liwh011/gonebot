# 前言
在编写新功能时，我们往往需要不断测试来保证代码按照预期运行。我想不少开发者应该有过这样的经历：
- 直接SSH连接到服务器上修改代码，不断重新运行来查看修改效果
- 本地部署机器人，不断重新运行来查看修改效果

不管怎么样都麻烦的要死，每次需要编译运行、手动按流程触发机器人，我才不想测试呢！这种重复的工作为什么不能交给程序自己完成？于是，本节内容应运而生。

# 例子
先看个例子，目前大体上的流程是模拟一个事件，然后输出消息记录，由人工判断输出结果是否正确。
其实感觉这样还是有点蠢，没有达到完全的自动化。
（我所设想的是，每模拟一个事件，可以获取bot的做出的行为。。总之先鸽着，日后有想法了再回来写）
```go
// 配置好友、加入的群
var mockServerOptions mock.NewMockServerOptions = mock.NewMockServerOptions{
	BotId: 114514,
	Friends: []mock.User{
		{
			UserId:   1919810,
			Nickname: "至高无上的SU",
		},
	},
	Groups: []mock.Group{
		{
			GroupId:   90000001,
			GroupName: "测试群",
			Members:   []mock.GroupMember{},
		},
	},
}

var config gonebot.BaseConfig = gonebot.BaseConfig{
	ApiCallTimeout: 10,
	Superuser:      []int64{1919810},
}

func TestHello(t *testing.T) {
	server := mock.NewMockServer(mockServerOptions)
	provider := mockProvider.NewMockProvider(server)
	engine := gonebot.NewEngineWithProvider(&config, provider)
	go engine.Run()
	time.Sleep(time.Millisecond * 200) // todo

	session := server.NewPrivateSession(1919810)
	session.MessageEventByText("我是你主人！")
	time.Sleep(time.Millisecond * 200) // todo
	t.Log(session.GetMessageHistory()) // 输出消息记录，人工检查
}
```