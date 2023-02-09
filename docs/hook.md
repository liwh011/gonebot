# 钩子
占坑

## 全局钩子
通过`gonebot.GlobalHooks`来访问。不隶属于任何一个Engine实例

目前有：
- `EngineCreated` Engine创建完成后调用
- `EngineWillTerminate` 收到CTRL+C，在结束前调用，一般用于做一些清理工作

## Engine钩子
通过`engine.Hooks`来访问。

- 插件生命周期
  - `PluginWillLoad` 每个插件将要加载时调用
  - `PluginLoaded` 每个插件加载完毕时调用

- 事件生命周期
  - `EventRecieved` 接收到事件，但仍未开始处理时触发
  - `EventHandled` 处理完毕该事件后触发