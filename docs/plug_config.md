# 插件配置
本框架将插件的配置放在了主配置文件当中，并会为你自动加载配置。

## 开关插件
默认情况下会加载所有已注册的插件。如果你想禁用某些插件，但不想修改代码再重新编译一次，那么你可以在配置中指定某插件的开关状态。被禁用的插件在**重启程序后**将不会被加载。

假设现有一个插件，它注册的信息为`Name: HelloWorld, Author: liwh011`，那么你可以这样禁用它：
```yaml
plugin:
  enable:
    # true为开启，false为禁用。
    HelloWorld@liwh011: false
```
其中，`HelloWorld@liwh011`为插件的唯一标识，在上节中有提到。


## 配置
如果你的插件需要外部配置，请为结构体添加`Config`字段，名字必须为`Config`。框架将会使用反射，在`Init`被调用前准备好Config的内容（不是模块的`init`），因此你可以尽情使用这个Config，不用担心加载问题。

如果配置需要默认值，请在注册插件之前手动初始化默认值，配置文件中未出现的字段将不会覆盖默认值。

配置文件中的字段风格可以使用大驼峰、小驼峰、蛇形，框架将会为你自动转换。
```go
// 配置
type HelloWorldConfig struct {
    Hello     int
    World     string
    // 配置文件中写`camel_case`、`camelCase`都可。
    CamelCase string
}

// 初始化默认配置
var defaultConfig = HelloWorldConfig {
    "Hello": 12345,
    "World": "helloworld",
}

type HelloWorld struct {
    // 如果你的插件有外部配置，请添加`Config`字段（一定要叫Config）。
    // 将在Init函数被调用前准备好这个Config
    // 如果没有可以不用。
    Config HelloWorldConfig
}

func init() {
    p := HelloWorld {
        Config: defaultConfig
    }
    // 注册插件，传入插件的指针。
    gonebot.RegisterPlugin(&p)
}

// 初始化插件
func (p *HelloWorld) Init(engine *gonebot.Engine) {
    // 在这里，Config已经加载好了，可以使用了。
}
```

对应配置：
```yaml
plugin:
  config:
    HelloWorld@liwh011:
      hello: 999
      world: aaaaaaaabbbbb
      camelCase: asddsas      # ok
      # camel_case: asddsas   # ok
      # CamelCase: asddsas    # ok
```

## 完整配置文件参考
```yaml
websocket:
  host: 127.0.0.1
  port: 6700
  access_token: dabsfckadakjdbkafbafa
  apicall_timeout: 30

plugin:
  # 开关控制
  enable:
    HelloWorld@liwh011: false # 禁用

  # 配置
  config:
    HelloWorld@liwh011:
      hello: 999
      world: aaaaaaaabbbbb
      camelCase: asddsas
```