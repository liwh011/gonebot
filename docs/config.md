# 配置
由于go是需要编译的语言，硬编码配置通常不是一个好的选择。本框架使用yml作为配置文件的格式。

## 基本配置文件
你的配置文件应至少包含以下内容：
```yml
apicall_timeout: 30
provider: websocket
provider_config:
  websocket:
    host: 127.0.0.1
    port: 6700
    access_token: asdsss
```
- `apicall_timeout` 调用协议端API超时时间（秒）
- `provider` 以什么方式与协议端连接
- `provider_config` 该种连接方式的配置
  - `websocket` 本框架默认采用正向WebSocket的方式连接到协议端，这里配置Ws服务端（协议提供端）的信息
    - `host` Ws服务器主机地址
    - `port` Ws服务器端口
    - `access_token` 与Ws服务器配置的Access token一致


随后调用`gonebot.LoadConfig(路径)`来载入配置。
```go
cfg := gonebot.LoadConfig("config.yml")
```

## 可选的配置项
```yml
cmd_prefix: 
  - "."
  - "/"

superuser:
  - 114514
  - 1919810

plugin:
# 略
```
- `cmd_prefix` 当使用`gonebot.Command("cmd")`或`gonebot.ShellLikeCommand("cmd", ..., ...)`时，若你指定了`cmd_prefix`，则在发送消息时，需要在命令前加上其中任意一个前缀才可以触发，如`/cmd xxxxx`。
- `superuser` 至高无上的超级管理员的QQ号，通常指定为Bot的拥有者。可以结合`gonebot.FromSuperuser`来实现特权功能。
- `plugin` 见[插件配置](./plug_config.md)

## 自定义配置文件
有时候随着功能的增长，你需要新增配置项，那么你需要用新的方式来载入配置。

你的自定义配置结构体应继承`gonebot.BaseConfig`，随后改用`gonebot.LoadCustomConfig`来载入配置。
```go
type MyConfig struct {
    gonebot.BaseConfig
    Item1 int       `yaml:"item1"`
    Item2 string    `yaml:"item2"`
    Obj1  struct {
        Field1 bool     `yaml:"field1"`
    } `yaml:"obj1"`
}


cfg := MyConfig{}
gonebot.LoadCustomConfig("config.yml", &cfg)
```

对应的配置文件如下：
```yaml
apicall_timeout: 30
provider: websocket
provider_config:
  websocket:
    host: 127.0.0.1
    port: 6700
    access_token: asdsss

item1: 114514
item2: abc
obj1: 
  field1: true
```

# EOF
你也可以不使用yml作为配置，总之，你的配置结构体只需实现`gonebot.Config`接口，能够返回框架需要的`BaseConfig`即可。