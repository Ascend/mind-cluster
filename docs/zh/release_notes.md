# 版本说明

## 版本配套说明

### 产品版本信息

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>
<table><tbody><tr><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.1.1"><p>产品名称</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p>MindCluster</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.2.1"><p>产品版本</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p>26.1.0</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.3.1"><p>版本类型</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p>Release版本</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>MindCluster 26.0版本规划：MindCluster 26.0.0、MindCluster 26.1.0、MindCluster 26.2.0和MindCluster 26.3.0。

### 相关产品版本配套说明

**表 1**  MindCluster软件版本配套表

|MindCluster|CANN| HDK                                                                   |MindSpeed-LLM|TorchNPU|MindSpore|
|--|--|-----------------------------------------------------------------------|--|--|--|
|26.1.0|9.1.0| <ul><li>Atlas A2/A3 系列产品：26.1.0</li><li>Ascend 950 系列产品：25.1.RC1</li></ul> |26.1.0|26.1.0|2.10.0|

### 版本兼容性说明

MindCluster各组件需要配套使用，请勿跨版本混用各组件。

>[!NOTE]
>本节表格中“/”表示不可配套，“Y”表示可配套。

**表 2**  MindCluster与CANN版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">CANN版本</th>
  </tr>
  <tr>
    <th>8.5.X</th>
    <th>9.0.X</th>
    <th>9.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td>/</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**表 3**  MindCluster与HDK版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">HDK版本</th>
  </tr>
  <tr>
    <th>25.5.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td>/</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.1.0</td>
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
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">MindSpeed-LLM版本</th>
  </tr>
  <tr>
    <th>2.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td>/</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

**表 5**  MindCluster与FrameworkPTAdapter版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">FrameworkPTAdapter版本</th>
  </tr>
  <tr>
    <th>7.3.X</th>
    <th>26.0.X</th>
    <th>26.1.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td>/</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.1.0</td>
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
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">MindSpore版本</th>
  </tr>
  <tr>
    <th>2.7.2</th>
    <th>2.9.X</th>
    <th>2.10.X</th>
  </tr></thead>
<tbody>
  <tr>
    <td>7.3.0</td>
    <td>Y</td>
    <td>/</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>/</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

### 病毒扫描结果

病毒扫描通过。

## 版本使用注意事项

无

## 26.1.0更新说明

### 新增特性

|特性名称| 特性描述                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|--|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|MindIO| MindIO支持IPv6场景。                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
|MindCluster Ascend FaultDiag| <ul><li>新增Ascend 950 系列产品的故障模式库。</li><li>新增基于pyMotor+vLLM的故障模式库。</li><li>优化链路诊断工具输出报告的内容。</li><li>故障诊断支持IPv6场景。</li></ul>                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
|MindCluster基础组件| <ul><li>支持存储DTFS故障。</li><li>Volcano支持Pod优先调度回原运行节点。</li><li>Ascend Device Plugin故障处理插件化。</li><li>支持带内检测1825故障上报。</li><li>提供1825网卡的RDMA设备插件。</li><li>发布镜像新增openEuler。</li><li>社区资料、组件的安装与部署易用性提升。</li><li>NPU Exporter支持分组配置采集周期。</li><li>Atlas 9000 A3 SuperPoD 集群算力系统能力补齐。</li><li>支持配置存活探针。</li><li>新增卡死检测与恢复功能。</li><li>支持Atlas 850 系列硬件产品和Atlas 950 SuperPoD的设备管理、亲和性调度、指标监控、故障检测，RankTable生成等基础能力。</li><li>支持Atlas 850 系列硬件产品的基础断点续训能力；支持Atlas 950 SuperPoD的全量断点续训能力。</li><li>支持Atlas 850 系列硬件产品和Atlas 950 SuperPoD的容器化能力。</li><li>支持推理故障恢复能力，包括优先级调度、缩P保D和实例级重调度。</li><li>支持推理场景基于负载的弹性扩缩容能力和容器快照能力。</li></ul> |

### 关键特性变更

MindCluster基础组件：

- Ascend Docker Runtime支持默认配置LD_LIBRARY_PATH环境变量，以便npu-smi工具能够正常使用。
- 启动Ascend Device Plugin、NPU Exporter、NodeD等组件时，若芯片数量不足，则等待驱动上报完整芯片的最大时长参数“-deviceResetTimeout”的默认值由60s修改为600s。

### 业务接口变更

|特性名称|接口变更|
|--|--|
|MindCluster Ascend FaultDiag|<ul><li>链路诊断工具新增配置命令set_config_dir，当前仅支持设置组网配置文件LLD.xlsx所在路径。</li><li>性能劣化功能（资源抢占和网络拥塞）采集指标数据时对训练、推理业务性能有一定影响，该特性将在后续版本日落。</li></ul>|
|MindCluster基础组件|所有K8s组件新增存活探针。|

### 已解决的问题

- 修复Atlas 350 标卡驱动部署后的软链接校验报错问题。
- 修复Ascend 950 系列产品huawei.com/AscendReal注解赋值phyID和logicID存在不同含义时，导致Ascend Device Plugin判断卡是否被占用出现异常的问题。
- 修复Ascend 950 系列产品device-cm的ManuallySeparateNPU中芯片名称未适配为NPU的问题。
- 修复NPU Exporter采集光模块指标失效时，optical_index存在内容但未上报的问题。

### 遗留问题

进程级别重调度特性在多次重调度恢复后，可能存在PyTorch原生组件gloo的段错误问题，概率约0.00125，详细请参见[issue 188266](https://github.com/pytorch/pytorch/issues/188266)。可以通过配置Job/Pod重调度作为兜底措施。

## 升级影响

### 升级过程对现行系统的影响

无

### 升级后对现行系统的影响

Infer Operator组件从26.1.0之前版本升级到26.1.0及之后版本时，需删除日志目录重新创建为root权限或修改日志目录及日志文件的权限为root。

## 26.1.0版本配套文档

|文档名称|内容简介|更新说明|
|--|--|--|
|《[MindCluster 集群调度用户指南](./scheduling/01_introduction/00_overview.md)》|提供集群调度组件说明、特性原理和使用参考，包括各组件的安装部署、集成适配示例和API参考，以及部分调度方案的原理介绍参考。|新增使用helm安装组件、开发者指南、容器快照部署及使用等，其他变更详见《[MindCluster 集群调度用户指南](./scheduling/01_introduction/00_overview.md)》。|
|《[MindCluster 故障诊断用户指南](./faultdiag/README.md)》|提供日志采集、日志清洗与转储、故障诊断等功能的使用指导。|新增Ascend 950 系列产品、基于pyMotor+vLLM的故障模式等，其他变更详见《[MindCluster 故障诊断用户指南](./faultdiag/README.md)》。|

## 漏洞修补列表

无
