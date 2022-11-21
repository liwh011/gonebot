注：以下内容均在`mock`子包下。

# Mock Server
脱离真实真实环境的第一步是把服务端（如go-cqhttp）换掉，服务端通常负责处理API调用以及下发事件。
这里的`Mock Server`作为一个假服务端，理所当然地具备这些能力。它提供了API调用桩，同时我们也可以通过它来手动下发事件。

## 创建
通过`mock.NewMockServer`即可创建。创建时需要传入选项`NewMockServerOptions`，在处理API请求以及模拟事件时会用到这些数据。
| 选项字段           | 类型            | 默认值      | 说明                                                           |
| ------------------ | --------------- | ----------- | -------------------------------------------------------------- |
| `BotId`            | `int64`         | `10000`     | botQQ号                                                        |
| `BotName`          | `string`        | `"MockBot"` | bot昵称                                                        |
| `Friends`          | `[]mock.User`   | `nil`       | 好友列表，会影响生成的消息事件的类型                           |
| `Groups`           | `[]mock.Group`  | `nil`       | 群组列表，会影响`get_group_list`接口的结果                     |
| `SendEventTimeout` | `time.Duration` | 1s          | 推送事件的超时时间，一般不用管。超时则意味着什么地方逻辑写错了 |


## 模拟事件
`MockServer`以虚拟会话的方式来模拟真实的会话场景，可以基于会话方便地模拟事件。

另外，所有模拟事件的方法除了返回生成的事件以外，默认会自动将事件发出。如果你仅仅只需要生成事件，请将`mockServer.AutoSendEvent`设为`false`。

### PrivateSession
模拟一个私聊会话。使用`mockServer.NewPrivateSession`创建，函数接受一个`userId`的参数。当该id是bot的好友（即存在于mockServer.Friends中），则模拟出来的事件就是好友私聊；否则是临时会话。

- `GetMessageHistory` 获取该会话的聊天记录。
- `MessageEvent` 模拟私聊消息事件。
- `MessageEventByText` 上面那个的纯文字快捷版
- `RecallEvent` 模拟撤回事件，参数指定待撤回消息的Id
- `PokeEvent` 模拟戳一戳事件

### GroupSession
模拟一个群聊会话。使用`mockServer.NewGroupSession`创建，函数接受一个`groupId`的参数，将以该id所对应的`Group`中的成员信息来模拟事件，若该id不存在，则默认为没有人的空群。

- `GetMessageHistory` 获取该会话的聊天记录。
- `MessageEvent` 模拟某成员发消息的事件。id不存在则默认为普通成员。
- `AnonymousMessageEvent` 模拟匿名消息事件

### 其他事件
虽然大部分事件都可以归为私聊、群聊事件，但终归有例外，例如元事件、添加好友事件等。这些事件则直接挂在`MockServer`上。

- `ConnectedEvent` 生命周期的连接事件，初次连接时必须发送一个。`MockServer`会自动发送。

### 如何自行补充
目前就只有这么些事件，本懒鬼表示将来不够再加。你当然也可以选择自己补充，session中的字段全都是Public的。当然，顺手提个PR会更好~

本质上以上模拟事件的方法都在其内部调用了`mockServer.SendEvent`来将事件发出。
**请注意，`mockServer.SendEvent`方法仅会在`mockServer.AutoSendEvent`为`true`时发送事件**。

你可以编写一个函数，以session和其他必要的数据作为入参，函数中构造好事件后调用`mockServer.SendEvent`来下发事件。
如果session中保存的数据无法满足你的需求，你也可以“继承”或自行编写，按照本框架类似的方法来下发事件即可。

## API调用处理
实际上也没处理什么，只是简单的为无需返回值的API返回一个成功的响应而已。因此绝大多数`get_xxx`的API是无法使用的（除了`get_group_list`、`get_login_info`可以正常工作）。


## 一些结构体及方法
- `User` 用户。通常用于Bot的好友列表
- `Group` 群聊
  - `GetMember` 根据id获取`GroupMember`
  - `AddMember` 添加一个成员
  - `RemoveMember` 删除一个成员
- `GroupMember` 群成员，比`User`多一些群聊相关的字段
  - `SetMember` 将该成员设为普通成员
  - `SetAdmin` 将该成员设为管理员
  - `SetOwner` 将该成员设为群主
- `MessageRecord` 单条消息
- `MessageHistory` 消息历史记录。实现了排序接口，可以按时间排序
