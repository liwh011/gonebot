# 配置
由于go是需要编译的语言，硬编码配置通常不是一个好的选择。本框架使用yml作为配置文件的格式。

## 基本配置文件
你的配置文件应至少包含以下内容：
```yml
websocket:
  # Ws服务器主机地址
  host: 127.0.0.1
  # Ws服务器端口
  port: 6700
  # 与Ws服务器配置的Acess token一致
  access_token: dabsfckadakjdbkafbafa
  # 调用API超时时间（秒）
  apicall_timeout: 30
```
随后调用`gonebot.LoadConfig(路径)`来载入配置。
```go
cfg := gonebot.LoadConfig("config.yml")
```

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
websocket:
  # Ws服务器主机地址
  host: 127.0.0.1
  # Ws服务器端口
  port: 6700
  # 与Ws服务器配置的Acess token一致
  access_token: dabsfckadakjdbkafbafa
  # 调用API超时时间（秒）
  apicall_timeout: 30

item1: 114514

item2: abc

obj1: 
  field1: true
```

# EOF
你也可以不使用yml作为配置，总之，你的配置结构体只需实现`gonebot.Config`接口，能够返回框架需要的`BaseConfig`即可。