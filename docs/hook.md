# 钩子
占坑

## 生命周期钩子
通过`gonebot.EngineHookManager`来访问。目前有：
- `EngineCreated` Engine创建完成后调用
- `PluginWillLoad` 每个插件将要加载时调用
- `PluginLoaded` 每个插件加载完毕时调用
- `EngineWillTerminate` 收到CTRL+C，在结束前调用，一般用于做一些清理工作