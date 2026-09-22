# 性能劣化、慢节点与慢网络

## 性能劣化故障<a name="ZH-CN_TOPIC_0000002479386488"></a>

### 使用7.1.RC1及以上版本TaskD<a name="ZH-CN_TOPIC_0000002511346475"></a>

MindCluster集群调度组件结合MindStudio提供的profiling能力，对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态打点和打点数据持久化功能、可动态启停训练任务打点功能，无需重启任务进行诊断，对训练无损耗。

当前支持的打点数据如[表1](#zh-cn_topic_0000002194466236_table5530103025919)所示。

**表 1**  打点数据说明

<a name="zh-cn_topic_0000002194466236_table5530103025919"></a>

|打点数据的类型|支持的AI框架|提供支持的组件|
|--|--|--|
|<p>FP</p><p>（标识前向传播数据）</p>|<p>PyTorch</p><p>仅支持单算子场景。</p>|mstx_torch_plugin|
|<p>Step</p><p>（标识Step时延）</p>|PyTorch、MindSpore|<ul><li>PyTorch<ul><li>原生优化器场景：若TorchNPU为7.1.RC1版本，需使用mstx_torch_plugin；若TorchNPU为7.1.RC1以上版本，无需使用mstx_torch_plugin，TorchNPU自带Step打点。</li><li>自定义优化器场景：手动增加打点数据。</li></ul></li><li>MindSpore<ul><li>MindFormers场景：Step打点数据由MindFormers提供。</li><li>MindSpeed场景：不提供Step打点数据。</li></ul></li></ul>|
|<p>Communication</p><p>（标识通信算子）</p>|PyTorch、MindSpore|<ul><li>PyTorch：TorchNPU</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>SaveCheckpoint</p><p>（标识SaveCheckpoint耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：TorchNPU</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>DataLoader</p><p>（标识DataLoader耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：TorchNPU</li><li>MindSpore：MindSpore框架</li></ul>|

#### 使用约束<a name="zh-cn_topic_0000002194466236_section487603614"></a>

- 当前Step、SaveCheckpoint、FP、DataLoader仅支持同步开启。如需关闭以上四类打点数据，需同时关闭Communication。
- Communication通信算子数据支持单独开启、关闭。
- 动态轻量打点功能与MindStudio的全量打点功能不可同时开启，开启全量打点功能会导致性能劣化故障不能正常采集数据。

### 使用其他版本TaskD<a name="ZH-CN_TOPIC_0000002511346483"></a>

MindCluster集群调度组件结合MindStudio提供的profiling能力，对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态打点和打点数据持久化功能、可动态启停训练任务打点功能，无需重启任务进行诊断，对训练无损耗。

当前支持的打点数据如[表2](#zh-cn_topic_0000002194466236_table553010302591923)所示。

**表 2**  打点数据说明

<a name="zh-cn_topic_0000002194466236_table553010302591923"></a>

|打点数据的类型|支持的AI框架|提供支持的组件|
|--|--|--|
|<p>FP</p><p>（标识前向传播数据）</p>|<p>PyTorch</p><p>仅支持单算子场景。</p>|mstx_torch_plugin|
|<p>Step</p><p>（标识Step时延）</p>|PyTorch、MindSpore|<ul><li>PyTorch<ul><li>原生优化器场景：若TorchNPU为7.1.RC1及以下版本，需使用mstx_torch_plugin；若TorchNPU为7.1.RC1以上版本，无需使用mstx_torch_plugin，TorchNPU自带Step打点。</li><li>自定义优化器场景：手动增加打点数据。</li></ul></li><li>MindSpore<ul><li>MindFormers场景：Step打点数据由MindFormers提供。</li><li>MindSpeed场景：不提供Step打点数据。</li></ul></li></ul>|
|<p>Communication</p><p>（标识通信算子）</p>|PyTorch、MindSpore|<ul><li>PyTorch：TorchNPU</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>SaveCheckpoint</p><p>（标识SaveCheckpoint耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：TorchNPU</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>DataLoader</p><p>（标识DataLoader耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：TorchNPU</li><li>MindSpore：MindSpore框架</li></ul>|

#### 使用约束<a name="zh-cn_topic_0000002194466236_section487603614-duplicate-2"></a>

- 当前Step、SaveCheckpoint、FP、DataLoader仅支持同步开启。如需关闭以上四类打点数据，需同时关闭Communication。
- Communication通信算子数据支持单独开启、关闭。
- 动态轻量打点功能与MindStudio的全量打点功能不可同时开启，开启全量打点功能会导致性能劣化故障不能正常采集数据。

## 慢节点&慢网络故障<a name="ZH-CN_TOPIC_0000002511426421"></a>

<a name="ZH-CN_TOPIC_0000002532640773"></a>

MindCluster集群调度组件结合MindCluster Ascend FaultDiag（故障诊断工具）提供的在线诊断能力，为集群中的慢节点&慢网络故障提供诊断功能。

### 慢节点诊断<a name="ZH-CN_TOPIC_0000002500880704"></a>

#### 功能说明<a name="zh-cn_topic_0000002278667326_section27999216294"></a>

对于AI集群中出现的节点训练性能劣化现象，提供支持实时检测计算域问题或网络导致的慢节点，以便用户通过切换或其他方式隔离慢节点。

当前仅支持与ClusterD和NodeD集成进行在线部署，请参见[安装部署](../../../03_installation_guide/02_installation/00_helm_installation.md)章节完成ClusterD和NodeD部署。

- 慢节点算法：基于训练场景关键性能指标，感知实时劣化状态；针对通信算子、计算算子同步关系，实现慢计算卡、慢通信域问题定界。
- 慢节点清洗：对节点内部增量数据转化并清洗，生成清洗结果csv文件。
- 慢节点调度：调度慢节点整体流程，控制数据清洗和慢节点算法。

### 慢网络诊断<a name="ZH-CN_TOPIC_0000002500720860"></a>

#### 功能说明<a name="zh-cn_topic_0000002313236861_section27999216294"></a>

支持提供参数面网络连通性检测，实时进行网络监测和异常上报，辅助故障分析和定界定位，提前预警网络故障和亚健康风险信息，保障集群网络的长期稳定运行。

当前仅支持与ClusterD和NodeD集成进行在线部署，请参见[安装部署](../../../03_installation_guide/02_installation/00_helm_installation.md)章节完成ClusterD和NodeD部署。

- 慢网络算法：对节点之间的网络拨测数据进行分析、检测，并输出网络诊断结果。
- 慢网络调度：把控探测任务启停，上报故障结果，调度慢网络整体流程。

## 相关操作

- [配置性能劣化诊断](../03_configuration/06_performance_diagnosis.md)
- [查询和验证故障](../04_querying_and_verifying_faults.md)
