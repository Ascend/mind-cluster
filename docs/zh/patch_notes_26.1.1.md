# 补丁说明

## 补丁描述

### 补丁基本信息

**补丁基本信息**

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>
<table><tbody><tr><th class="firstcol" valign="top" width="25%"><p>补丁号</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p>MindCluster 26.1.1</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>产品基础版本</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p>MindCluster 26.1.0</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>发布时间</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p>2026-09-15</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>与同一版本内其他补丁关系</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p>-</p>
</td>
</tr>
</tbody>
</table>

**软件包信息**

| 软件包名                                                                                 | 软件包说明                         |
|--------------------------------------------------------------------------------------|-------------------------------|
| Ascend-docker-runtime_<i>\<version></i>_linux-<i>\<arch></i>.run                     | Ascend Docker Runtime软件包      |
| Ascend-helm-deploy-tool_<i>\<version></i>_linux.zip                                  | Helm部署软件包                     |
| Ascend-mindxdl-ascend-operator_<i>\<version></i>_linux-<i>\<arch></i>.zip            | Ascend Operator软件包            |
| Ascend-mindxdl-clusterd_<i>\<version></i>_linux-<i>\<arch></i>.zip                   | ClusterD软件包                   |
| Ascend-mindxdl-container-manager_<i>\<version></i>_linux-<i>\<arch></i>.zip          | Container Manager软件包          |
| Ascend-mindxdl-device-plugin_<i>\<version></i>_linux-<i>\<arch></i>.zip              | Ascend Device Plugin软件包       |
| Ascend-mindxdl-faultdiag_<i>\<version></i>_linux-<i>\<arch></i>.zip                  | 故障诊断软件包                       |
| Ascend-mindxdl-infer-operator_<i>\<version></i>_linux-<i>\<arch></i>.zip             | Infer Operator软件包             |
| Ascend-mindxdl-k8s-rdma-shared-dev-plugin_<i>\<version></i>_linux-<i>\<arch></i>.zip | K8s RDMA Shared Dev Plugin软件包 |
| Ascend-mindxdl-mindio_<i>\<version></i>_linux-<i>\<arch></i>.zip                     | MindIO软件包                     |
| Ascend-mindxdl-noded_<i>\<version></i>_linux-<i>\<arch></i>.zip                      | NodeD软件包                      |
| Ascend-mindxdl-npu-exporter_<i>\<version></i>_linux-<i>\<arch></i>.zip               | NPU Exporter软件包               |
| Ascend-mindxdl-taskd_<i>\<version></i>_linux-<i>\<arch></i>.zip                      | TaskD软件包                      |
| Ascend-mindxdl-volcano_<i>\<version></i>_linux-<i>\<arch></i>.zip                    | Volcano软件包                    |

>[!NOTE]
><i>\<version></i>为软件包的版本号；<i>\<arch></i>为CPU架构。

**兼容性说明**

无

### 安装补丁的影响

#### 安装过程中对现行系统的影响

**对业务的影响**

无

**对网络通信的影响**

无

#### 安装后对现行系统的影响

无

### 配套关系说明

**表 1**  MindCluster软件版本配套表

| MindCluster |CANN| HDK                                                                   |MindSpeed-LLM|TorchNPU|MindSpore|
|-------------|--|-----------------------------------------------------------------------|--|--|--|
| 26.1.1      |9.1.0| <ul><li>Atlas 350 加速卡：25.7.RC1</li><li>Atlas 950 SuperPoD 超节点：25.1.RC1</li><li>Atlas 850E 超节点/Atlas 650E 服务器：25.6.RC1</li><li>其他产品：26.1.0</li></ul> |26.1.0|26.1.0|2.10.0|

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
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.1</td>
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
    <th colspan="1"><term>Ascend 950 系列产品</term>HDK版本</th>
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
    <td>26.1.1</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
<col style="width: 88px">
<col style="width: 91px">
<col style="width: 98px">
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">其他产品HDK版本</th>
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
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.1</td>
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
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.1</td>
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
</colgroup>
<thead>
  <tr>
    <th rowspan="2">MindCluster</th>
    <th colspan="3">TorchNPU版本</th>
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
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.1</td>
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
    <td> </td>
    <td> </td>
  </tr>
  <tr>
    <td>26.0.0</td>
    <td>Y</td>
    <td>Y</td>
    <td> </td>
  </tr>
  <tr>
    <td>26.1.0</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
  <tr>
    <td>26.1.1</td>
    <td>Y</td>
    <td>Y</td>
    <td>Y</td>
  </tr>
</tbody>
</table>

### 病毒扫描结果

病毒扫描通过。

## 解决的问题

**MindCluster集群调度**

- 修复Helm安装时镜像拉取失败的问题。
- 修复在软切分场景下的ConfigMap激增问题。
- 修复ClusterD的任务故障频次计数问题。
- 修复Volcano超节点调度场景下的冗余资源配置失效问题。
- 修复MindCluster计算节点组件部署在Atlas 推理系列产品上时的DCMI初始化失败问题。
- 修复在实际物理超节点数量大于配置的super-pod-size时，Volcano的崩溃问题。
- 修复Volcano的保留节点数设置得比逻辑超节点包含的节点数大时，调度结果不满足亲和性的问题。
- 修复Volcano在Evict驱逐Pod时，Volcano框架侧并发导致Volcano崩溃的问题。
- 修复Infer Operator镜像非root用户启动时启动失败的问题。
- 修复dpu-dp对1825的bond口检测失效问题。
- 修复NodeD在IPMI失效后恢复的场景下无法自愈的问题。
- 修复x86的dockerfile支持无UMDK包安装的问题。

## 遗留问题

无

## 漏洞修补列表

无

## 版本配套文档

|文档名称|内容简介|
|--|--|
|《[MindCluster 集群调度用户指南](./scheduling/01_introduction/00_overview.md)》|提供集群调度组件说明、特性原理和使用参考，包括各组件的安装部署、集成适配示例和API参考，以及部分调度方案的原理介绍参考。|
|《[MindCluster 故障诊断用户指南](./faultdiag/ascend-faultdiag/01_introduction/01_overview.md)》|提供日志采集、日志清洗与转储、故障诊断等功能的使用指导。|
