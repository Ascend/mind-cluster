# 补丁说明

## 补丁描述

### 补丁基本信息

**补丁基本信息**

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>
<table><tbody><tr><th class="firstcol" valign="top" width="25%"><p>补丁号</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p>MindCluster 26.0.1</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>产品基础版本</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p>MindCluster 26.0.0</p>
</td>
</tr>
<tr><th class="firstcol" valign="top" width="25%"><p>发布时间</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p>2026-06-05</p>
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

|软件包名|软件包说明|
|--|--|
|Ascend-docker-runtime_<i>\<version></i>_linux-<i>\<arch></i>.run|Ascend Docker Runtime软件包|
|Ascend-mindxdl-ascend-operator_<i>\<version></i>_linux-<i>\<arch></i>.zip|Ascend Operator软件包|
|Ascend-mindxdl-clusterd_<i>\<version></i>_linux-<i>\<arch></i>.zip|ClusterD软件包|
|Ascend-mindxdl-container-manager_<i>\<version></i>_linux-<i>\<arch></i>.zip|Container Manager软件包|
|Ascend-mindxdl-device-plugin_<i>\<version></i>_linux-<i>\<arch></i>.zip|Ascend Device Plugin软件包|
|Ascend-mindxdl-infer-operator_<i>\<version></i>_linux-<i>\<arch></i>.zip|Infer Operator软件包|
|Ascend-mindxdl-mindio_<i>\<version></i>_linux-<i>\<arch></i>.zip|MindIO软件包|
|Ascend-mindxdl-noded_<i>\<version></i>_linux-<i>\<arch></i>.zip|NodeD软件包|
|Ascend-mindxdl-npu-exporter_<i>\<version></i>_linux-<i>\<arch></i>.zip|NPU Exporter软件包|
|Ascend-mindxdl-taskd_<i>\<version></i>_linux-<i>\<arch></i>.zip|TaskD软件包|
|Ascend-mindxdl-volcano_<i>\<version></i>_linux-<i>\<arch></i>.zip|Volcano软件包|

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

**表 1**  建议配套其他版本补丁

|软件名称|补丁版本|说明|
|--|--|--|
|Ascend HDK| <ul><li>Atlas 350 标卡：1.0.RC1</li><li>其他产品：26.0.RC1</li></ul> |[华为support](https://support.huawei.com/carrierindex/zh/hwe/index.html)，或在[昇腾社区](https://www.hiascend.com/zh/developer/download)下载|
|CANN|9.0.0|[华为support](https://support.huawei.com/carrierindex/zh/hwe/index.html)，或在[昇腾社区](https://www.hiascend.com/zh/developer/download)下载|

### 版本兼容性说明

MindCluster各组件需要配套使用，请勿跨版本混用各组件。

**表 2**  软件版本兼容性说明

|MindCluster软件版本|MindCluster待升级版本|CANN版本兼容性|Ascend HDK版本兼容性|FrameworkPTAdapter版本兼容性|MindSpore版本兼容性|
|--|--|--|--|--|--|
|MindCluster 26.0.1|<ul><li>MindCluster 7.0.RC1及补丁版本</li><li>MindCluster 7.1.RC1及补丁版本</li><li>MindCluster 7.2.RC1及补丁版本</li><li>MindCluster 7.3.0及补丁版本</li></ul>|<ul><li>CANN 8.5.0及补丁版本</li><li>CANN 9.0.0及补丁版本</li></ul>|<ul><li>Ascend HDK 25.5.0及补丁版本</li><li>Ascend HDK 26.0.RC1及补丁版本</li><li>Ascend HDK 1.0.RC1及补丁版本</li></ul>|<ul><li>FrameworkPTAdapter 7.3.0及补丁版本</li><li>FrameworkPTAdapter 26.0.0及补丁版本</li></ul>|<ul><li>MindSpore 2.7.2及补丁版本</li><li>MindSpore 2.9.0及补丁版本</li></ul>|

### 病毒扫描结果 

病毒扫描通过。

## 解决的问题

**MindCluster基础组件**

- 解决了在某些场景下，cluster-info-switch和cluster-info-node ConfigMap更新错误的问题。
- 解决了Atlas 350 标卡device-info ConfigMap中上报芯片名称错误的问题。
- 解决了Atlas 350 标卡Ascend Docker Runtime挂载设备时校验驱动中存在软链接导致的挂载失败问题。

## 遗留问题

无

## 漏洞修补列表

无

## 基础版本配套产品文档获取方法

您可以通过以下路径浏览和获取相关的文档：[昇腾社区文档中心](https://www.hiascend.com/document)。
