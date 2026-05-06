# feature: NPU Exporter新增node_base_info指标

## 1. 概述

本特性在NPU Exporter中新增`node_base_info`指标，用于暴露当前节点的NPU Exporter版本信息和驱动版本信息，使得用户可以通过Prometheus或Telegraf直接查询节点的驱动版本，便于集群运维和版本管理。

## 2. 需求背景

### 2.1 问题描述

当前NPU Exporter未提供节点级别的驱动版本信息指标，用户无法通过Prometheus或Telegraf直接获取各节点的驱动版本。在实际运维场景中，用户经常需要了解集群中各节点的驱动版本信息，以便：

1. **版本一致性检查**：确认集群中各节点的驱动版本是否一致，避免因驱动版本不一致导致的兼容性问题。
2. **升级验证**：在驱动升级后，快速验证各节点的驱动版本是否已更新到目标版本。
3. **故障排查**：当出现设备异常时，快速获取节点的驱动版本信息，辅助问题定位。

目前获取驱动版本需要登录到各个节点手动执行命令，效率低下且不便于自动化运维。

### 2.2 解决方案

新增`node_base_info`指标，包含`exporterVersion`和`driverVersion`两个标签，通过Prometheus指标的方式暴露节点的驱动版本信息，实现：

- 通过Prometheus查询即可获取各节点的驱动版本
- 通过Telegraf采集即可将驱动版本信息纳入监控体系
- 与现有指标体系保持一致，无需额外的查询接口

## 3. 术语定义

| 术语                | 解释                                                       |
|-------------------|----------------------------------------------------------|
| NPU Exporter      | 华为自研的专门收集华为NPU各种监测信息和指标，并封装成Prometheus专用数据格式的服务组件        |
| node_base_info    | 新增的节点信息指标，包含exporterVersion和driverVersion标签              |
| exporterVersion   | NPU Exporter组件的构建版本号                                     |
| driverVersion     | 昇腾AI处理器的驱动版本号，通过HDK接口GetDcmiVersion获取                    |
| Prometheus        | 开源的系统监测和警报工具包                                            |
| Telegraf          | InfluxData开发的指标采集代理                                      |
| NodeBaseCollector | 负责采集node_base_info指标的采集器结构体                              |
| nodeBaseInfoCache | 节点基础信息的缓存结构体，包含timestamp、exporterVersion和driverVersion字段 |

## 4. 设计方案

### 4.1 指标定义

新增指标`node_base_info`，具体定义如下：

| 属性     | 值                                   |
|--------|-------------------------------------|
| 指标分组   | nodeBase                            |
| 默认开关配置 | 默认开                                 |
| 指标名称   | node_base_info                      |
| 指标类型   | gauge                               |
| 指标值    | 固定为1（占位字符，无实际含义）                    |
| 指标说明   | the common information of this node |
| 标签     | exporterVersion、driverVersion       |

**标签说明：**

| 标签名             | 类型     | 说明                 |
|-----------------|--------|--------------------|
| exporterVersion | string | 当前NPU Exporter版本信息 |
| driverVersion   | string | 驱动版本信息             |

### 4.2 支持的产品形态

- Atlas 训练系列产品
- Atlas A2 训练系列产品
- Atlas A3 训练系列产品
- 推理服务器（插Atlas 300I 推理卡）
- Atlas 推理系列产品
- Atlas 800I A2 推理服务器
- A200I A2 Box 异构组件
- Atlas 350 标卡
- Atlas 850 系列硬件产品
- Atlas 950 SuperPoD

## 5. 实现细节

### 5.1 核心代码变更

变更文件：`component/npu-exporter/collector/metrics/collector_for_node_base.go`

#### 5.1.1 指标描述定义

```go
nodeInfoDesc = common.BuildDescWithLabel("node_base_info", "the common information of this node",
    []string{exporterVersionLabel, driverVersionLabel})
```

使用`common.BuildDescWithLabel`函数构建指标描述，标签列表为`["exporterVersion", "driverVersion"]`。

#### 5.1.2 采集器与缓存结构体定义

```go
type NodeBaseCollector struct {
    common.MetricsCollectorAdapter
}

type nodeBaseInfoCache struct {
    timestamp       time.Time
    exporterVersion string
    driverVersion   string
}
```

- `NodeBaseCollector`：嵌入`common.MetricsCollectorAdapter`，继承`LocalCache`等基础能力
- `nodeBaseInfoCache`：缓存结构体，包含采集时间戳、exporter版本和驱动版本信息

#### 5.1.3 Describe方法

```go
func (c *NodeBaseCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- nodeInfoDesc
}
```

在`Describe`方法中注册`nodeInfoDesc`，使Prometheus能够发现该指标。

#### 5.1.4 CollectToCache方法

```go
func (c *NodeBaseCollector) CollectToCache(n *common.NpuCollector, chipList []common.HuaWeiAIChip) {
    c.LocalCache.Store(common.GetCacheKey(c), nodeBaseInfoCache{
        timestamp:       time.Now(),
        exporterVersion: versions.BuildVersion,
        driverVersion:   n.Dmgr.GetDcmiVersion(),
    })
}
```

- 在采集阶段将数据写入`LocalCache`，缓存key通过`common.GetCacheKey(c)`获取（即结构体名称`NodeBaseCollector`）
- `exporterVersion`取自`versions.BuildVersion`（编译时注入的版本号）
- `driverVersion`通过`n.Dmgr.GetDcmiVersion()`获取，调用底层HDK接口获取驱动版本
- `timestamp`记录采集时间，用于后续Prometheus和Telegraf上报时携带时间戳

#### 5.1.5 UpdatePrometheus方法

```go
func (c *NodeBaseCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *common.NpuCollector,
    containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) {
    nodeBaseInfo, ok := c.LocalCache.Load(common.GetCacheKey(c))
    if !ok {
        logger.Debugf("cacheKey(%v) not found", common.GetCacheKey(c))
        return
    }
    cache, ok := nodeBaseInfo.(nodeBaseInfoCache)
    if !ok {
        logger.Error("cache type mismatch")
		return
    }
    doUpdateMetric(ch, cache.timestamp, 1, []string{cache.exporterVersion, cache.driverVersion}, nodeInfoDesc)
}
```

- 从`LocalCache`中读取缓存数据，若缓存不存在则记录日志并返回
- 对缓存数据进行类型断言，若类型不匹配则记录错误日志
- 调用`doUpdateMetric`辅助函数上报指标，携带缓存中的时间戳
- 指标值固定为1（gauge类型），作为占位字符
- 标签值按`exporterVersion`、`driverVersion`顺序传入

#### 5.1.6 UpdateTelegraf方法

```go
func (c *NodeBaseCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *common.NpuCollector,
    containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) map[string]map[string]interface{} {
    nodeBaseInfo, ok := c.LocalCache.Load(common.GetCacheKey(c))
    if !ok {
        logger.Debugf("cacheKey(%v) not found", common.GetCacheKey(c))
        return fieldsMap
    }
    cache, ok := nodeBaseInfo.(nodeBaseInfoCache)
    if !ok {
        logger.Error("cache type mismatch")
    }

    if fieldsMap[common.KeyForTextMetrics] == nil {
        fieldsMap[common.KeyForTextMetrics] = make(map[string]interface{})
    }

    labelsMap := make(map[string]string)
    labelsMap["exporterVersion"] = cache.exporterVersion
    labelsMap["driverVersion"] = cache.driverVersion

    tetegrafData := common.TelegrafData{
        Labels:    labelsMap,
        Metrics:   map[string]interface{}{utils.GetDescName(nodeInfoDesc): 1},
        Timestamp: cache.timestamp,
    }
    fieldsMap[common.KeyForTextMetrics]["ascend"] = tetegrafData
    return fieldsMap
}
```

Telegraf场景下：
- 从`LocalCache`中读取缓存数据，若缓存不存在则记录日志并返回原始fieldsMap
- 初始化`KeyForTextMetrics`字段（若不存在），用于存放文本类指标数据
- 构建`TelegrafData`结构体，包含：
  - `Labels`：标签映射，包含`exporterVersion`和`driverVersion`
  - `Metrics`：指标值映射，指标名通过`utils.GetDescName(nodeInfoDesc)`获取，值为1
  - `Timestamp`：使用缓存中的采集时间戳
- 将`TelegrafData`写入`fieldsMap[common.KeyForTextMetrics]["ascend"]`

### 5.2 数据获取方式

驱动版本信息通过`n.Dmgr.GetDcmiVersion()`获取，该方法底层调用HDK的DCMI接口`dcmi_get_dcmi_version`获取驱动版本号。`Dmgr`是`devmanager.DeviceInterface`接口的实现，在NPU Exporter启动时初始化。

数据采集采用缓存机制：在`CollectToCache`阶段获取数据并缓存到`LocalCache`中，在`UpdatePrometheus`和`UpdateTelegraf`阶段从缓存读取数据进行上报。缓存key为结构体名称`NodeBaseCollector`，通过`common.GetCacheKey(c)`获取。缓存中同时记录采集时间戳，确保Prometheus和Telegraf上报时使用一致的采集时间。

### 5.3 指标输出示例

#### Prometheus格式

```
# HELP node_base_info the common information of this node
# TYPE node_base_info gauge
node_base_info{exporterVersion="v26.1.0",driverVersion="26.0.3"} 1 1694772754612
```

#### Telegraf格式

Telegraf场景下，数据通过`TelegrafData`结构体上报，最终由`handleTextMetrics`函数处理，调用`acc.AddFields`输出：
- measurement名称：`ascend`
- tags：`exporterVersion="v26.1.0"`, `driverVersion="26.0.1"`
- fields：`node_base_info=1i`
- timestamp：缓存中的采集时间戳

## 6. 文档变更

### 6.1 Prometheus Metrics接口文档

在`docs/zh/scheduling/api/npu_exporter/01_prometheus_metrics_api.md`的"表1 版本数据信息"中新增一行：

| 类别       | 数据信息名称         | 数据信息说明                              | 数据信息标签字段                                                  | 字段类型            | 单位               | 支持的产品形态                                                                                                                                                                                   |
|----------|----------------|-------------------------------------|-----------------------------------------------------------|-----------------|------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| nodeBase | node_base_info | the common information of this node | exporterVersion：当前NPU Exporter版本信息 / driverVersion：驱动版本信息 | string / string | 1：占位字符，无实际含义 / - | Atlas 训练系列产品 / Atlas A2 训练系列产品 / Atlas A3 训练系列产品 / 推理服务器（插Atlas 300I 推理卡） / Atlas 推理系列产品 / Atlas 800I A2 推理服务器 / A200I A2 Box 异构组件 / Atlas 350 标卡 / Atlas 850 系列硬件产品 / Atlas 950 SuperPoD |

### 6.2 Telegraf数据信息说明文档

在`docs/zh/scheduling/api/npu_exporter/02_telegraf_data_description.md`的"表1 版本数据信息"中新增一行：

| 类别       | 数据信息名称         | 数据信息说明                                                                                           | 单位 | 支持的产品形态                                                                                                                                                                                   |
|----------|----------------|--------------------------------------------------------------------------------------------------|----|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| nodeBase | node_base_info | the common information of this node，包括：exporterVersion：当前NPU Exporter版本信息 / driverVersion：驱动版本信息 | -  | Atlas 训练系列产品 / Atlas A2 训练系列产品 / Atlas A3 训练系列产品 / 推理服务器（插Atlas 300I 推理卡） / Atlas 推理系列产品 / Atlas 800I A2 推理服务器 / A200I A2 Box 异构组件 / Atlas 350 标卡 / Atlas 850 系列硬件产品 / Atlas 950 SuperPoD |

## 7. 兼容性和风险

### 7.1 兼容性

- **向后兼容**：新增指标不影响现有指标，已有的Prometheus查询和告警规则不受影响。
- **数据格式兼容**：遵循Prometheus标准的标签指标格式，与现有指标体系完全兼容。
- **Telegraf兼容**：Telegraf数据格式中新增字段，不影响已有字段的采集和解析。
- **采集周期**：正常情况下安装驱动后，需要重启服务器才能正常使用驱动，npu-exporter随着重启，所以驱动版本信息仅在npu-exporter组件重启时采集一次，数据准确性不影响；同时，后续支持按分组配置采集频率后，可将采集周期配置为一天一次。

### 7.2 风险评估

| 风险        | 影响                                              | 缓解措施                                  |
|-----------|-------------------------------------------------|---------------------------------------|
| HDK接口调用失败 | driverVersion标签值为空，指标仍会上报                       | 指标值固定为1，不影响其他指标采集；接口失败时返回空字符串，用户可据此判断 |
| 驱动版本信息变更  | 版本信息在采集周期内缓存，驱动升级后需等待下一个采集周期或重启NPU Exporter才能更新 | 在文档中说明此行为，建议驱动升级后重启NPU Exporter       |

## 8. 测试验证

### 8.1 功能测试

1. **Prometheus场景**：
    - 启动NPU Exporter后，访问`/metrics`接口，验证`node_base_info`指标存在
    - 验证`exporterVersion`标签值与NPU Exporter构建版本一致
    - 验证`driverVersion`标签值与实际驱动版本一致
    - 验证指标值固定为1

2. **Telegraf场景**：
    - 配置Telegraf采集NPU Exporter数据
    - 验证采集数据中包含`node_base_info`字段
    - 验证采集数据中包含正确的`exporterVersion`和`driverVersion`标签信息


## 9. 总结

本特性通过新增`node_base_info`指标，暴露节点的exporter版本和驱动版本信息，使用户能够通过Prometheus或Telegraf直接查询节点的驱动版本。该特性采用缓存机制，在`CollectToCache`阶段采集并缓存数据，在`UpdatePrometheus`和`UpdateTelegraf`阶段从缓存读取数据进行上报，与其他采集器（如NetworkCollector、HccsCollector等）保持一致的架构风格。该特性完全向后兼容，对系统性能影响极小，有效提升了集群运维的便捷性和自动化水平。
