# DP Hot Reset Plugin Development

Ascend Device Plugin provides a plugin mechanism in the offline hot reset flow, allowing developers to insert custom logic at key points of the reset, such as collecting NPU information before the reset and recording events after the reset.

## Plugin Mechanism Overview

The hot reset flow exposes hooks at three key nodes, which are executed in the following order:

```text
PreReset → Driver in-band reset → CustomReset → AfterReset
```

**Relationship Between Plugins and Hooks**

A plugin is an independent module that implements the `HotResetPlugin` interface, and a hook is an execution point in the reset flow where a plugin can be mounted. A plugin may implement any combination of hooks as needed, and the framework determines which execution points it is mounted to based on the methods the plugin actually overrides.

| Hook        | Execution Time        | Timeout Duration | Typical Use                              |
| ----------- | --------------------- | ---------------- | ---------------------------------------- |
| PreReset    | Before driver reset   | 10 seconds       | Collect NPU information, record pre-reset state |
| CustomReset | After driver in-band reset | 5 minutes    | Custom reset method (such as out-of-band reset), repair reset failures |
| AfterReset  | After the reset flow ends | 10 seconds   | Record reset results, send event notifications |

**When to Develop a New Plugin**

When a feature has an independent responsibility boundary, a new plugin should be developed rather than adding logic to an existing plugin. For example, information collection and reset logging belong to different responsibilities and should be implemented as separate plugins to facilitate independent configuration of switches and maintenance. If the logic of multiple hooks is tightly coupled (for example, out-of-band reset requires performing the reset in `CustomReset` and cleaning up state in `AfterReset`), they should be placed in the same plugin.

**Execution Rules**

- The same hook of different plugins is executed sequentially in the configured order.
- `PreReset` and `AfterReset` have no return value and do not block subsequent flows.
- `CustomReset` uses chained passing: the error returned by the previous plugin is passed as the input to the next plugin.
- Each plugin has independent timeout control. After a timeout, the framework automatically skips the plugin and continues execution.

## Interface Definition

The plugin must implement the `HotResetPlugin` interface:

```go
type HotResetPlugin interface {
    Name() string
    PreReset(ctx context.Context, deviceList []ResetDevice)
    CustomReset(ctx context.Context, deviceList []ResetDevice, resetErr error) error
    AfterReset(ctx context.Context, deviceList []ResetDevice, resetErr error)
}
```

**ResetDevice struct**

```go
type ResetDevice struct {
    LogicID    int32   // Logic ID
    CardID     int32   // Card ID
    DeviceID   int32   // Device ID
    PhyID      int32   // Physical ID
    CardType   string  // Chip type
    IsFaultDev bool    // Whether it is a faulty device
    TokensLeft int32   // Number of remaining tokens for the faulty device
}
```

## How to Develop

1. Implement the plugin.

   The following example implements an NPU log collection plugin that collects device-side logs through the `msnpureport` command before driver reset, covering only the `PreReset` hook:

   - **Plugin function**: Collect NPU device-side logs before hot reset to facilitate problem locating after reset.
   - **Hook implementation**: Only `PreReset` is implemented, which executes the `msnpureport` command before driver reset. `CustomReset` and `AfterReset` use the default implementation of `HotResetPluginAdapter` (no custom reset is performed and reset results are not processed).
   - **Timeout control**: Pass the context to the child process through `exec.CommandContext`, so that command execution is automatically terminated after the framework timeout.

    Embedding `HotResetPluginAdapter` allows you to override only the required hooks without implementing all methods:

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

    >[!NOTE]
    >The default implementation of `CustomReset` in `HotResetPluginAdapter` directly returns the input parameter `resetErr`, ensuring that the error state is not modified during chained passing. If custom reset logic is required, override this method. `exec.CommandContext` automatically terminates the child process when the context is canceled, so there is no need to manually check `ctx.Done()`.

2. Register the plugin.

    Register the custom plugin in `InitPluginManager`:

    ```go
    func InitPluginManager(dmgr devmanager.DeviceInterface,
        kubeClient *kubeclient.ClientK8s) (*plugin.PluginManager, error) {
        pm := plugin.NewPluginManager()
        // Register the built-in plugin.
        pm.RegisterPlugin(NewOutBandResetPlugin(dmgr))
        pm.RegisterPlugin(NewResetRecordPlugin(kubeClient))
        // Register the custom plugin.
        pm.RegisterPlugin(&NpuLogCollectPlugin{})
        pm.Init()
        return pm, nil
    }
    ```

3. Configure the plugin switch.

    Control the plugin enablement status through the configuration file `/usr/local/hotResetPluginConfiguration.json`, which is located inside the device-plugin container:

    - `pluginName`: must be consistent with the value returned by the plugin's `Name()` method.
    - `state`: `ON` indicates that the plugin is enabled, and `OFF` indicates that the plugin is disabled.
    - If the configuration file does not exist or is malformed, the default configuration is used.

    A configuration example is provided below, using the `npuLogCollect` plugin as an example. Other plugins can be configured by referring to this example.

    - Enable the `npuLogCollect` plugin:

        ```json
        [
        {"pluginName": "outbandReset", "state": "ON"},
        {"pluginName": "resetRecord", "state": "OFF"},
        {"pluginName": "npuLogCollect", "state": "ON"}
        ]
        ```

    - Disable the `npuLogCollect` plugin:

        ```json
        [
        {"pluginName": "outbandReset", "state": "ON"},
        {"pluginName": "resetRecord", "state": "OFF"},
        {"pluginName": "npuLogCollect", "state": "OFF"}
        ]
        ```

**Default Configuration**

When the configuration file does not exist or fails to parse, the following default configuration is used:

```json
[
  {"pluginName": "outbandReset", "state": "ON"},
  {"pluginName": "resetRecord", "state": "OFF"}
]
```

| Plugin | Default Value | Description |
|------|---------|------|
| outbandReset | ON | Out-of-band reset plugin. Performs out-of-band reset when in-band reset fails. |
| resetRecord | OFF | Reset recording plugin. Creates K8s Events before and after reset to record the reset status. |

## Development Notes

- **Handle context cancellation correctly**: When a hook function involves time-consuming operations, check `ctx.Done()` via `select` to ensure graceful exit after timeout and avoid goroutine leaks.
- **PreReset does not block the reset**: `PreReset` has no return value and does not prevent the subsequent reset flow.
- **CustomReset chain semantics**: Multiple `CustomReset` plugins execute sequentially, and the error returned by the previous one is passed to the next; to fix a reset failure, return `nil` to indicate a successful fix.
- **Unique plugin names**: Registering multiple plugins with the same name returns an error.
- **Concurrency safety**: The reset flow may be executed concurrently in goroutines, so the plugin implementation must ensure thread safety.

## Plugin Code Reference

| Plugin | Description | Code Path |
|------|------|---------|
| Out-of-band reset plugin | Performs an out-of-band reset on chips when in-band reset fails. | `component/ascend-device-plugin/pkg/plugin/builtin/outband_reset_plugin.go` |
| Plugin interface and adapter | The HotResetPlugin interface definition and the default implementation of HotResetPluginAdapter. | `component/ascend-device-plugin/pkg/plugin/hot_reset_plugin.go` |
| Plugin manager | Plugin registration, configuration loading, hook execution, and timeout control. | `component/ascend-device-plugin/pkg/plugin/plugin_manager.go` |
| Plugin initialization | The InitPluginManager function, which registers built-in plugins and initializes the plugin manager. | `component/ascend-device-plugin/pkg/plugin/builtin/init.go` |
| HotResetPluginAdapter | A plugin adapter that provides default implementations of hook methods. After embedding it, you only need to override the required hooks. | `component/ascend-device-plugin/pkg/plugin/hot_reset_plugin.go` |
