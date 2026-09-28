# 关键特性

**MindCluster集群调度组件关键特性**

- **断点续训与故障恢复能力增强**：支持DPU故障与DPU亚健康触发断点续训，支持RL场景实例级恢复。Atlas 950 SuperPoD超节点支持UB网络故障进程级在线恢复。
- **推理高可用增强**：支持弹性伸缩复合指标定义；支持弹性扩缩容时使能容器快照；支持推理任务缩卡、缩Pod恢复。
- **调度与设备管理**：支持任务调度忽略ROCE网络健康状态；支持节点内的通用亲和性调度与跨迭代混合调度。支持Atlas 950 SuperPoD Flex超节点亲和性调度。
- **生态版本适配**：支持Kubernetes 1.36版本和Volcano 1.15版本。
- **运行时能力扩展**：Ascend Docker Runtime新增支持CRI-O，支持挂载UMDK和UB驱动用户态文件。
- **可观测性增强**：新增DPU Exporter组件，支持1825指标采集；NPU Exporter支持上报可用芯片数量，Ascend Device Plugin支持上报设备NUMA信息。
- **机型适配**：适配Atlas 950 SuperPoD Flex机型基础能力。

**MindCluster Ascend FaultDiag关键特性**

- **新增K8s集群运维Agent组件**：支持根据用户指令收集K8s集群中失败的昇腾训练/推理任务日志并执行故障诊断；支持接入模型API，优化诊断报告。
- **链路诊断工具增强**：新增支持Ascend 950PR&950DT系列产品光链路故障诊断；故障阈值支持通过配置文件修改，并可按光模块类型、指标设置不同阈值。

# 版本配套说明

## 产品版本信息

| 产品名称 | MindCluster |
| --- |-------------|
| 产品版本 | 26.2.0      |
| 版本类型 | Release版本   |
| 维护周期 | 1年          |

> [!NOTE] 说明
>
> MindCluster 26.0版本规划：MindCluster 26.0.0、MindCluster 26.1.0、MindCluster 26.2.0和MindCluster 26.3.0。

## 相关产品版本配套说明

**表 1**  MindCluster软件版本配套表

|MindCluster|CANN| HDK    |MindSpeed-LLM|TorchNPU|
|--|--|--------|--|--|
|26.2.0|9.2.0| 26.2.0 |26.2.0|26.2.0|

# 版本兼容性说明

MindCluster各组件需要配套使用，请勿跨版本混用各组件。
>[!NOTE]
>本节表格中“/”表示不可配套，“Y”表示可配套。

**表 2**  MindCluster与CANN版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 98px">
<col style="width: 98px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">CANN版本</th>
  </tr>
  <tr>
    <th>8.5.X</th>
    <th>9.0.X</th>
    <th>9.1.X</th>
    <th>9.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**表 3**  MindCluster与HDK版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 232px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="1">Ascend 950PR&950DT系列产品26年630以及之前的HDK版本</th>
  </tr>
  <tr>
    <th>25.1.RC1/25.6.RC1/25.7.RC1</th>
  </tr></thead>
<tbody>

  <tr>
    <td>26.0.0</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">HDK版本</th>
  </tr>
  <tr>
    <th>25.5.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
    <th>26.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**表 4**  MindCluster与MindSpeed-LLM版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">MindSpeed-LLM版本</th>
  </tr>
  <tr>
    <th>2.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
    <th>26.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**表 5**  MindCluster与TorchNPU版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">TorchNPU版本</th>
  </tr>
  <tr>
    <th>7.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
    <th>26.2.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**表 6**  MindCluster与MindSpore版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="4">MindSpore版本</th>
  </tr>
  <tr>
    <th>2.7.2</th>
    <th>2.9.X</th>
    <th>2.10.X</th>
    <th>2.11.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.2.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

# 版本使用注意事项

- HDK版本针对不同硬件、不同代际的版本号各不相同，需要先明确硬件形态，再根据硬件形态找到对应HDK版本，最后根据对照表找到MindCluster配套版本。
- MindSpore在2.11.0版本不再支持配套MindCluster使能进程级重调度、进程级在线恢复、在线压测、借轨通信等高阶断点续训能力，仅保留基础Job级和Pod级重调度能力。

# 更新说明

## 新增特性说明

**MindCluster集群调度组件**

**1、断点续训与故障恢复**

- 支持DPU故障触发断点续训。
- 支持DPU亚健康可配置策略主动触发断点续训。
- RL场景支持实例级恢复。
- Atlas 950 SuperPoD超节点支持UB网络故障进程级在线恢复。

**2、推理高可靠**

- 支持弹性伸缩复合指标定义。
- 支持弹性扩缩容时使能容器快照。
- 支持推理任务缩卡、缩Pod恢复。

**3、可靠性与健壮性**

- 完成快恢可靠性能力验证。
- 完成降lane等亚健康故障主动隔离节点可靠性能力验证。
- 支持一体机自愈型故障100%自愈。

**4、调度**

- K8s资源调度支持PD实例异构。
- 未申请NPU的Pod无需配置skip标识即可正常调度。
- 支持Kubernetes 1.36版本kube-scheduler与worker节点组件。
- 支持Volcano 1.15版本。
- job-summary中的ranktable支持Ascend 950PR&950DT系列产品。
- 支持节点内的通用亲和性调度与跨迭代混合调度。
- 支持Atlas 950 SuperPoD Flex超节点亲和性调度。
- 亲和性调度支持Pod类型。
- 提供硬件基础能力标签。
- Infer Operator支持自动配置实例亲和性。
- Infer Operator支持MetaService拉起。
- 支持任务调度忽略ROCE网络健康状态。

**5、设备管理**

- hccn_tool改为dcmi接口，优化调度耗时。
- 支持组件对外版本信息输出。

**6、Ascend Docker Runtime**

- 支持CRI-O。
- 支持挂载UMDK和UB驱动用户态文件。

**7、可观测性**

- DPU Exporter支持1825指标采集。
- NPU Exporter支持上报可用芯片数量。
- Ascend Device Plugin支持上报设备NUMA信息。
- vnpu硬切分指标监控增强。
- 支持Ascend 950PR&950DT的NPU利用率新指标接口。

**8、机型适配**

- 适配Atlas 950 SuperPoD Flex超节点机型的基础能力。

**9、架构优化**

- 完成MindCluster内部组件解耦。
- 消减三方依赖。

**10、资料**

- 优化断点续训和故障监测相关资料。

**MindCluster Ascend FaultDiag故障诊断**

**1、日志诊断工具**

- 基于Cqe error status错误类型并结合日志细化、增强UB 链路故障诊断能力。
- 新增支持Atlas 950 SuperPoD Flex服务器故障诊断。

**2、链路诊断工具**

- 新增支持Ascend 950PR&950DT系列产品光链路故障诊断。
- 故障阈值支持通过配置文件修改，并可按光模块类型、指标设置不同阈值。

**3、K8s集群运维Agent组件**

- 支持根据用户指令收集K8s集群中失败的昇腾训练/推理任务日志并执行故障诊断；支持接入模型API，优化诊断报告。

## 关键特性变更

本版本继承MindCluster 26.1.0及其之前发布版本的部分特性。主要变更如下：

- hccn_tool工具接口调整为基于DCMI接口实现，优化调度耗时。

## 业务接口变更

- Ascend Docker Runtime新增支持CRI-O。
- Ascend Device Plugin新增上报设备NUMA信息。
- NPU Exporter新增上报可用芯片数量。
- hccn_tool工具接口调整为基于DCMI接口实现，优化调度耗时。
- 未申请NPU的Pod无需配置skip标识即可正常调度（调度行为调整）。
- Volcano支持任务调度忽略ROCE网络健康状态。
- 其他兼容性问题请参见[MindCluster组件兼容性问题公告](https://gitcode.com/Ascend/mind-cluster/issues/588)。

## 已解决的问题

- 旧版本中针对81078603故障码的恢复检测周期为5分钟，本版本优化为带有退避机制的短周期快速检测机制，能够更快速检测恢复事件并消除故障。
- MindCluster 26.1.0版本以及之前的版本中，Atlas A3系列产品开启算子重执行，在linkdown故障场景下，进程级重调度有概率失败。本版本中已修复，参考PR：https://gitcode.com/Ascend/mind-cluster/pull/4253。

## 遗留问题

进程级别重调度特性在多次重调度恢复后，存在PyTorch原生组件gloo的段错误问题，概率约0.00125，详细请参见[issue 188266](https://github.com/pytorch/pytorch/issues/188266)。可以通过配置Job/Pod重调度作为兜底措施。

# 升级影响

## 升级过程对现行系统的影响

无。

## 升级后对现行系统的影响

Infer Operator组件从26.1.0之前版本升级到26.1.0及之后版本时，需删除日志目录重新创建为root权限或修改日志目录及日志文件的权限为root。

# 版本配套文档

| 文档名称 | 内容简介 | 更新说明                                                                          |
| --- | --- |-------------------------------------------------------------------------------|
| 《MindCluster 26.2.0 集群调度用户指南》 | 提供集群调度组件说明、特性原理和使用参考，包括各组件的安装部署、集成适配示例和API参考，以及部分调度方案的原理介绍参考。 | 新增断点续训（DPU故障/亚健康触发）、PD实例异构等特性指导、重构断点续训章节，其余变更详见《MindCluster 26.2.0 集群调度用户指南》。 |
| 《MindCluster 26.2.0 故障诊断用户指南》 | 提供日志采集、日志清洗与转储、故障诊断等功能的使用指导。 | 优化断点续训和故障监测相关指导，其余变更详见《MindCluster 26.2.0 故障诊断用户指南》。                          |

# 病毒扫描结果

病毒扫描通过。

# 漏洞修补列表

无。

# 修订记录

| 文档版本 | 发布日期       | 修改说明 |
| --- |------------| --- |
| 01 | 2026-09-30 | 第一次正式发布。 |
