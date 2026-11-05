# 补丁说明

## 补丁描述

### 补丁基本信息

**补丁基本信息**

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>
<table><tbody><tr><th class="firstcol" valign="top" width="25%"><p>补丁号</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p>MindCluster 26.2.1</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>产品基础版本</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p>MindCluster 26.2.0</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>发布时间</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p>2026-11-05</p>
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

| MindCluster |CANN| HDK  |MindSpeed-LLM|TorchNPU|
|-------------|--|--------|--|--|
| 26.2.1      |9.2.0| 26.2.0 |26.2.0|26.2.0|

### 版本兼容性说明

MindCluster各组件需要配套使用，请勿跨版本混用各组件。

>[!NOTE]
>本节表格中“/”表示不可配套，“Y”表示可配套。

**表 2**  MindCluster与CANN版本兼容

<table style="table-layout: fixed; width: 433px"><colgroup>
<col style="width: 156px">
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
  <tr>
    <td>26.2.1</td>
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
  <tr>
    <td>26.2.1</td>
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
  <tr>
    <td>26.2.1</td>
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
  <tr>
    <td>26.2.1</td>
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
  <tr>
    <td>26.2.1</td>
    <td>Y</td>
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

- 自定义挂载路径支持配置主机路径、主机路径:容器内路径两种方式。

## 遗留问题

无

## 漏洞修补列表

无

## 版本配套文档

|文档名称|内容简介|
|--|--|
|《[MindCluster 集群调度用户指南](./scheduling/01_introduction/00_overview.md)》|提供集群调度组件说明、特性原理和使用参考，包括各组件的安装部署、集成适配示例和API参考，以及部分调度方案的原理介绍参考。|
|《[MindCluster 故障诊断用户指南](./faultdiag/ascend-faultdiag/01_introduction/01_overview.md)》|提供日志采集、日志清洗与转储、故障诊断等功能的使用指导。|
