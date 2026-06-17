# DP热复位插件开发

Ascend Device Plugin在离线热复位流程中开放了插件机制，支持开发者在复位的关键节点插入自定义逻辑，例如复位前采集NPU信息、复位后记录事件等。

## 插件机制概述

热复位流程在三个关键节点开放钩子，按以下顺序执行：

```
PreReset → 驱动带内复位 → CustomReset → AfterReset
```

**插件与钩子的关系：** 插件是实现`HotResetPlugin`接口的独立模块，钩子是插件在复位流程中可挂载的执行点。一个插件可按需实现任意组合的钩子，框架根据插件实际覆盖的方法决定其挂载到哪些执行点。

| 钩子          | 执行时机    | 超时时间 | 典型用途                  |
| ----------- | ------- | ---- | --------------------- |
| PreReset    | 驱动复位前   | 10秒  | 采集NPU信息、记录复位前状态       |
| CustomReset | 驱动带内复位后 | 5分钟  | 自定义复位方式（如带外复位）、修复复位失败 |
| AfterReset  | 复位流程结束后 | 10秒  | 记录复位结果、发送事件通知         |

**何时开发新插件：** 当功能具有独立的职责边界时，应开发新插件而非在已有插件中添加逻辑。例如，信息采集与复位记录属于不同职责，应分别实现为独立插件，便于独立配置开关和维护。若多个钩子的逻辑紧密耦合（如带外复位需在CustomReset中执行复位、在AfterReset中清理状态），则应放在同一插件中。

**执行规则：**

- 不同插件的同一钩子函数按配置顺序依次执行
- PreReset/AfterReset无返回值，不阻断后续流程
- CustomReset采用链式传递：前一个插件返回的error作为下一个插件的入参
- 每个插件有独立的超时控制，超时后框架自动跳过该插件继续执行

## 接口定义

插件需实现`HotResetPlugin`接口：

```go
type HotResetPlugin interface {
    Name() string
    PreReset(ctx context.Context, deviceList []ResetDevice)
    CustomReset(ctx context.Context, deviceList []ResetDevice, resetErr error) error
    AfterReset(ctx context.Context, deviceList []ResetDevice, resetErr error)
}
```

**ResetDevice结构体：**

```go
type ResetDevice struct {
    LogicID    int32   // 逻辑ID
    CardID     int32   // 卡ID
    DeviceID   int32   // 设备ID
    PhyID      int32   // 物理ID
    CardType   string  // 芯片类型
    IsFaultDev bool    // 是否为故障设备
    TokensLeft int32   // 故障设备剩余令牌数
}
```

## 开发步骤

### 1. 实现插件

以下示例实现一个NPU日志采集插件，在驱动复位前通过`msnpureport`命令采集设备侧日志，仅覆盖`PreReset`钩子：

- **插件功能**：在热复位前采集NPU device侧日志，便于复位后的问题定位
- **钩子实现**：仅实现`PreReset`，在驱动复位前执行`msnpureport`命令；`CustomReset`和`AfterReset`使用`HotResetPluginAdapter`的默认实现（不执行自定义复位、不处理复位结果）
- **超时控制**：通过`exec.CommandContext`将context传递给子进程，框架超时后自动终止命令执行

嵌入`HotResetPluginAdapter`可只覆盖需要的钩子，无需实现全部方法：

```go
package myplugin

import (
    "context"
    "os/exec"

    "Ascend-device-plugin/pkg/plugin"
    "ascend-common/common-utils/hwlog"
)

const (
    npuLogCollectName = "npuLogCollect"
    msnpureportCmd    = "msnpureport"
)

type NpuLogCollectPlugin struct {
    plugin.HotResetPluginAdapter
}

func (p *NpuLogCollectPlugin) Name() string {
    return npuLogCollectName
}

func (p *NpuLogCollectPlugin) PreReset(ctx context.Context, deviceList []plugin.ResetDevice) {
    if err := exec.CommandContext(ctx, msnpureportCmd).Run(); err != nil {
        hwlog.RunLog.Errorf("collect NPU log failed, err: %v", err)
        return
    }
    hwlog.RunLog.Infof("collect NPU log success, device count: %d", len(deviceList))
}
```

> **注意**：`HotResetPluginAdapter`中`CustomReset`的默认实现直接返回入参`resetErr`，确保链式传递中不修改错误状态。如需自定义复位逻辑，覆盖此方法即可。`exec.CommandContext`会在context取消时自动终止子进程，无需手动检查`ctx.Done()`。

### 2. 注册插件

在`InitPluginManager`中注册自定义插件：

```go
func InitPluginManager(dmgr devmanager.DeviceInterface,
    kubeClient *kubeclient.ClientK8s) (*plugin.PluginManager, error) {
    pm := plugin.NewPluginManager()
    // 注册内置插件
    pm.RegisterPlugin(NewOutBandResetPlugin(dmgr))
    pm.RegisterPlugin(NewResetRecordPlugin(kubeClient))
    // 注册自定义插件
    pm.RegisterPlugin(&NpuLogCollectPlugin{})
    pm.Init()
    return pm, nil
}
```

### 3. 配置插件开关

通过配置文件`/usr/local/hotResetPluginConfiguration.json`控制插件启用状态，配置文件位于device-plugin容器内：

- `pluginName`：需与插件`Name()`方法返回值一致
- `state`：`ON`启用，`OFF`禁用
- 若配置文件不存在或格式错误，使用默认配置

**配置示例：**

启用`npuLogCollect`插件：

```json
[
  {"pluginName": "outbandReset", "state": "ON"},
  {"pluginName": "resetRecord", "state": "OFF"},
  {"pluginName": "npuLogCollect", "state": "ON"}
]
```

禁用`npuLogCollect`插件：

```json
[
  {"pluginName": "outbandReset", "state": "ON"},
  {"pluginName": "resetRecord", "state": "OFF"},
  {"pluginName": "npuLogCollect", "state": "OFF"}
]
```
- 其他插件可参考该示例打开关闭

**默认配置：**

当配置文件不存在或解析失败时，使用以下默认配置：

```json
[
  {"pluginName": "outbandReset", "state": "ON"},
  {"pluginName": "resetRecord", "state": "OFF"}
]
```

| 插件 | 默认状态 | 说明 |
|------|---------|------|
| outbandReset | ON | 带外复位插件，带内复位失败时执行带外复位 |
| resetRecord | OFF | 复位记录插件，在复位前后创建K8s Event记录复位状态 |

## 开发注意事项

- **正确处理context取消**：钩子函数中涉及耗时操作时，务必通过`select`检查`ctx.Done()`，确保超时后可优雅退出，避免goroutine泄漏
- **PreReset不阻断复位**：PreReset无返回值，不会阻止后续复位流程
- **CustomReset链式语义**：多个CustomReset插件顺序执行，前一个返回的error传递给下一个；若需修复复位失败，返回`nil`表示修复成功
- **插件名称唯一**：重复注册同名插件会返回错误
- **并发安全**：复位流程可能在协程中并发执行，插件实现需保证线程安全

## 插件代码参考

| 插件 | 说明 | 代码路径 |
|------|------|---------|
| 带外复位插件 | 带内复位失败时对芯片执行带外复位 | `component/ascend-device-plugin/pkg/plugin/builtin/outband_reset_plugin.go` |
| 插件接口与适配器 | HotResetPlugin接口定义及HotResetPluginAdapter默认实现 | `component/ascend-device-plugin/pkg/plugin/hot_reset_plugin.go` |
| 插件管理器 | 插件注册、配置加载、钩子执行与超时控制 | `component/ascend-device-plugin/pkg/plugin/plugin_manager.go` |
| 插件初始化 | InitPluginManager函数，注册内置插件并初始化插件管理器 | `component/ascend-device-plugin/pkg/plugin/builtin/init.go` |
| HotResetPluginAdapter | 插件适配器，提供钩子方法的默认实现，嵌入后只需覆盖需要的钩子 | `component/ascend-device-plugin/pkg/plugin/hot_reset_plugin.go` |
