# 版本配套说明<a name="ZH-CN_TOPIC_0000002492283212"></a>

## 产品版本信息<a name="ZH-CN_TOPIC_0000002524562895"></a>

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>
<table><tbody><tr id="zh-cn_topic_0000001935094108_row244mcpsimp"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000001935094108_p246mcpsimp"><a name="zh-cn_topic_0000001935094108_p246mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p246mcpsimp"></a>产品名称</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p id="p92555221126"><a name="p92555221126"></a><a name="p92555221126"></a><span id="ph19255162231216"><a name="ph19255162231216"></a><a name="ph19255162231216"></a>MindCluster</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row255mcpsimp"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.2.1"><p id="zh-cn_topic_0000001935094108_p257mcpsimp"><a name="zh-cn_topic_0000001935094108_p257mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p257mcpsimp"></a>产品版本</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p id="zh-cn_topic_0000001935094108_p233mcpsimp"><a name="zh-cn_topic_0000001935094108_p233mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p233mcpsimp"></a>26.0.0</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row7259721105019"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.3.1"><p id="zh-cn_topic_0000001935094108_p7260182135013"><a name="zh-cn_topic_0000001935094108_p7260182135013"></a><a name="zh-cn_topic_0000001935094108_p7260182135013"></a>版本类型</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p id="zh-cn_topic_0000001935094108_p72606219501"><a name="zh-cn_topic_0000001935094108_p72606219501"></a><a name="zh-cn_topic_0000001935094108_p72606219501"></a>候选版本</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row880773455018"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.4.1"><p id="zh-cn_topic_0000001935094108_p198071234135017"><a name="zh-cn_topic_0000001935094108_p198071234135017"></a><a name="zh-cn_topic_0000001935094108_p198071234135017"></a>维护周期</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.4.1 "><p id="zh-cn_topic_0000001935094108_p15807123412509"><a name="zh-cn_topic_0000001935094108_p15807123412509"></a><a name="zh-cn_topic_0000001935094108_p15807123412509"></a>1年</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>MindCluster 26.0版本规划：MindCluster 26.0.0（候选版本）、MindCluster 26.1.0（候选版本）、MindCluster 26.2.0（候选版本）和MindCluster 26.3.0（正式版本）。

## 相关产品版本配套说明<a name="ZH-CN_TOPIC_0000002524562893"></a>

|产品名称|版本|
|--|--|
|Ascend HDK| <ul><li>Atlas 350 加速卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD：25.1.RC1</li><li>其他产品：26.0.RC1</li></ul> |
|CANN|9.0.0|

## 病毒扫描结果<a name="ZH-CN_TOPIC_0000002492443186"></a>

病毒扫描通过。

# 版本兼容性说明<a name="ZH-CN_TOPIC_0000002524442915"></a>

MindCluster各组件需要配套使用，请勿跨版本混用各组件。

**表 1**  软件版本兼容性说明

|MindCluster软件版本|MindCluster待升级版本|CANN版本兼容性|Ascend HDK版本兼容性|
|--|--|--|--|
|MindCluster 26.0.0|<ul><li>MindCluster 7.0.RC1及7.0.RC1.x</li><li>MindCluster 7.1.RC1及7.1.RC1.x</li><li>MindCluster 7.2.RC1及7.2.RC1.x</li><li>MindCluster 7.3.0及7.3.0.x</li></ul>|<ul><li>CANN 8.5.0及8.5.0.x</li><li>CANN 9.0.0及9.0.0.x</li></ul>|<ul><li>Ascend HDK 25.1.RC1及25.1.RC1.x</li><li>Ascend HDK 25.5.0及25.5.0.x</li><li>Ascend HDK 26.0.RC1及26.0.RC1.x</li></ul>|

# 版本使用注意事项<a name="ZH-CN_TOPIC_0000002492283210"></a>

无

# 26.0.0更新说明<a name="ZH-CN_TOPIC_0000002492443184"></a>

## 新增特性<a name="ZH-CN_TOPIC_0000002524442919"></a>

|特性名称|特性描述|
|--|--|
|MindIO ACP|支持ACP\&TFT能力兼容。|
|MindIO TFT|<ul><li>支持ACP\&TFT能力兼容。</li><li>支持精度异常后按照指定Checkpoint步数在线恢复。</li><li>支持讯飞Hulk框架的优化器差异化副本场景。</li></ul>|
|MindCluster Ascend FaultDiag|<ul><li>故障诊断支持A5故障模式。</li><li>新增Ascend-faultdiag-toolkit工具，支持掉卡故障诊断和基础设施链路诊断。</li><li>不再支持故障模式库构建成二进制，直接开源故障模式库。</li></ul>|
|MindCluster基础组件|<ul><li>业务面网络支持IPv6。</li><li>支持基于任务维度配置可容忍的故障级别或具体故障码。</li><li>支持Atlas A2 系列产品/Atlas A3 系列产品的软切分调度。</li><li>支持Atlas A2 系列产品/Atlas A3 系列产品的硬切分调度。</li><li>大EP任务支持交换机亲和性调度。</li><li>ClusterD的gRPC心跳检测周期从默认的5分钟调整为5秒。</li><li>Verl支持CKPT异步保存。</li><li>基于集群维度识别故障是否为硬件故障，反复发生的硬件故障自动强制隔离，避免反复造成任务中断。</li><li>集群维度的自动强制隔离，以及节点维度的自动强制隔离，都支持配置自动释放时间。</li><li>新增任意层级网络亲和性调度算法，适配天工超节点。</li><li>任务信息订阅接口新增V2版本，V2版本首次订阅返回历史任务信息。</li><li>新增ConfigMap展示任务调度失败原因，方便快速定位。</li><li>NPU Exporter支持通过配置文件监听上报自定义指标。</li><li>NPU Exporter的多个NPU利用率获取方式从多个接口改为1个接口，避免数据不对应的情况。</li><li>进程级重调度流程优化。</li><li>ClusterD故障通知服务支持通过域名注册。</li><li>ClusterD作业信息订阅接口新增作业唯一标识符字段。</li><li>NPU Exporter支持Atlas 350 加速卡上报UB场景的指标。</li><li>Ascend Docker Runtime支持Atlas 350 加速卡。</li><li>Atlas 350 加速卡支持亲和性调度、设备发现、Ranktable生成、故障重调度。</li><li>Atlas 350 加速卡新增故障码。</li><li>适配Atlas 350 加速卡NPU ID系统变更。</li><li>Atlas 350 加速卡去掉SDID字段及常量整改。</li><li>支持Infer Operator通过自定义CRD管理推理任务。</li></ul>|

## 关键特性变更<a name="ZH-CN_TOPIC_0000002524562891"></a>

MindCluster基础组件：

- 支持任意网络层级的通用亲和性调度算法。
- 支持Atlas A2 系列产品/Atlas A3 系列产品的软切分和硬切分。
- 支持通过ConfigMap查看命令，展示任务调度失败原因。
- 支持集群维度下的反复故障芯片的自动强制隔离和自动释放。
- 支持基于任务维度配置可容忍的故障级别或故障码。
- Atlas 350 加速卡场景下：
  - 任务申请资源“huawei.com/Ascend910”变更为“huawei.com/npu”。
  - 底层dcmi接口调用变更为dcmiV2接口调用。

MindCluster Ascend FaultDiag：

- 故障诊断新增Ascend-faultdiag-toolkit工具。

## 业务接口变更<a name="ZH-CN_TOPIC_0000002492443182"></a>

|特性名称|接口变更|
|--|--|
|MindIO ACP|无|
|MindIO TFT|新增tft_register_exception_handler：注册异常处理程序。|
|MindCluster Ascend FaultDiag|新增Ascend-faultdiag-toolkit工具相关接口。|
|MindCluster基础组件|<ul><li>任务创建接口新增可容忍故障级别、故障码、容忍时长配置字段。</li><li>任务创建接口新增软切分模式、AI Core百分比、高带宽内存量配置字段。</li><li>ClusterD支持配置故障自动强制隔离的使能开关、触发频率、隔离时长。</li><li>Ascend Device Plugin新增自动强制隔离的隔离时长配置字段。</li><li>支持多级网络拓扑配置，以及任务的多级网络亲和配置。</li><li>新增任务订阅接口V2。</li><li>新增任务调度异常原因查询接口。</li><li>新增文件形式的自定义指标接口。</li><li>NPU Exporter的NPU利用率接口优化。</li><li>新增Atlas 350 加速卡的设备基础信息、故障码和芯片名称。</li></ul>|

## 已解决的问题<a name="ZH-CN_TOPIC_0000002492283206"></a>

无

## 遗留问题<a name="ZH-CN_TOPIC_0000002492443180"></a>

无

# 升级影响<a name="ZH-CN_TOPIC_0000002492283208"></a>

## 升级过程对现行系统的影响<a name="ZH-CN_TOPIC_0000002524442911"></a>

无

## 升级后对现行系统的影响<a name="ZH-CN_TOPIC_0000002492443178"></a>

无

# 26.0.0版本配套文档<a name="ZH-CN_TOPIC_0000002524562889"></a>

|文档名称|内容简介|更新说明|
|--|--|--|
|[《MindCluster 集群调度用户指南》](./scheduling/introduction.md)|提供集群调度组件说明、特性原理和使用参考，包括各组件的安装部署、集成适配示例和API参考，以及部分调度方案的原理介绍参考。|新增软切分调度、多级调度等，其他变更详见[《MindCluster 集群调度用户指南》](./scheduling/introduction.md)。|
|[《MindCluster 故障诊断用户指南》](./faultdiag/introduction.md)|提供日志采集、日志清洗与转储、故障诊断等功能的使用指导。|新增A5故障模式、Ascend-faultdiag-toolkit工具等，其他变更详见[《MindCluster 故障诊断用户指南》](./faultdiag/introduction.md)。|

# 漏洞修补列表<a name="ZH-CN_TOPIC_0000002524442913"></a>

无
