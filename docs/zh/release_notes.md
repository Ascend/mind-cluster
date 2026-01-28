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
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p id="zh-cn_topic_0000001935094108_p233mcpsimp"><a name="zh-cn_topic_0000001935094108_p233mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p233mcpsimp"></a>7.3.0</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row7259721105019"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.3.1"><p id="zh-cn_topic_0000001935094108_p7260182135013"><a name="zh-cn_topic_0000001935094108_p7260182135013"></a><a name="zh-cn_topic_0000001935094108_p7260182135013"></a>版本类型</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p id="zh-cn_topic_0000001935094108_p72606219501"><a name="zh-cn_topic_0000001935094108_p72606219501"></a><a name="zh-cn_topic_0000001935094108_p72606219501"></a>正式版本</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row880773455018"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.4.1"><p id="zh-cn_topic_0000001935094108_p198071234135017"><a name="zh-cn_topic_0000001935094108_p198071234135017"></a><a name="zh-cn_topic_0000001935094108_p198071234135017"></a>维护周期</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.4.1 "><p id="zh-cn_topic_0000001935094108_p15807123412509"><a name="zh-cn_topic_0000001935094108_p15807123412509"></a><a name="zh-cn_topic_0000001935094108_p15807123412509"></a>1年</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>MindCluster  7.0版本规划：MindCluster  7.0.RC1（候选版本）、MindCluster  7.1.RC1（候选版本）、MindCluster  7.2.RC1（候选版本）和MindCluster  7.3.0（正式版本）。

## 相关产品版本配套说明<a name="ZH-CN_TOPIC_0000002524562893"></a>

|产品名称|版本|
|--|--|
|Ascend HDK|25.5.0|
|CANN|8.5.0|


## 病毒扫描结果<a name="ZH-CN_TOPIC_0000002492443186"></a>

病毒扫描通过，详细请参见[MindCluster 7.3.0 virus scan report.docx](./resource/MindCluster%207.3.0%20virus%20scan%20report.docx)。

# 版本兼容性说明<a name="ZH-CN_TOPIC_0000002524442915"></a>

MindCluster各组件需要配套使用，请勿跨版本混用各组件。

**表 1**  软件版本兼容性说明

|MindCluster软件版本|MindCluster待升级版本|CANN版本兼容性|Ascend HDK版本兼容性|
|--|--|--|--|
|MindCluster 7.3.0|<ul><li>MindCluster 6.0.0及6.0.0.x</li><li>MindCluster 7.0.RC1及7.0.RC1.x</li><li>MindCluster 7.1.RC1及7.1.RC1.x</li><li>MindCluster 7.2.RC1及7.2.RC1.x</li><li>MindCluster 7.3.0及7.3.0.x</li></ul>|<ul><li>CANN 8.1.RC1及8.1.RC1.x</li><li>CANN 8.2.RC1及8.2.RC1.x</li><li>CANN 8.3.RC1及8.3.RC1.x</li><li>CANN 8.5.0及8.5.0.x</li></ul>|<ul><li>Ascend HDK 25.0.RC1及25.0.RC1.x</li><li>Ascend HDK 25.2.0及25.2.0.x</li><li>Ascend HDK 25.3.RC1及25.3.RC1.x</li><li>Ascend HDK 25.5.0及25.5.0.x</li></ul>|


# 版本使用注意事项<a name="ZH-CN_TOPIC_0000002492283210"></a>

无。

# 7.3.0更新说明<a name="ZH-CN_TOPIC_0000002492443184"></a>






## 新增特性<a name="ZH-CN_TOPIC_0000002524442919"></a>

|特性名称|特性描述|
|--|--|
|MindIO ACP|MindIO ACP蓝区开源。|
|MindIO TFT|<ul><li>MindIO TFT支持MindSpore场景亚健康热切。</li><li>MindIO TFT蓝区开源。|
|MindCluster ToolBox|<ul><li>新增A3 A+X PCIE全眼图测试、HBM压测耗时和算法优化。</li><li>新增310P功耗压力。</li><li>新增DSA随机数算子压测。|
|MindCluster Ascend FaultDiag|新增A3 AI服务器故障事件。|
|MindCluster Ascend Deployer|无|
|MindCluster基础组件|<ul><li>关闭算子重执行下支持灵衢L1-L2链路故障的进程级在线恢复。</li><li>支持基于AIBrix vLLM部署NPU的故障实例流量隔离。</li><li>NPU Exporter支持输出SN序列号。</li><li>支持基于AIBrix vLLM服务化实例级重调度。</li><li>基于AIBrix社区CRD定义，支持一键式脚本生成对应YAML，支持一键式配置和下发。</li><li>基于社区原生CRD定义，支持一键式脚本生成对应YAML，支持一键式配置和下发。</li><li>支持SGLang OME部署与实例级重调度。</li><li>支持灵衢故障上报可靠性增强。</li><li>Volcano新增适配层，隔离不同任务控制器的差异，支持所有满足格式要求的podGroup下的亲和性调度。</li><li>调度资源占用优化，未完成调度时，任务通过一定时间重新入队。</li><li>公共故障支持预隔离处理级别。</li><li>NPU Exporter支持自定义指标。</li><li>支持A3推理多实例任务调度。</li><li>支持A3兼容A2 accelerator-type资源类型。</li><li>生态组件兼容验证。</li><li>新增推理任务守护进程参考设计。</li><li>支持一体机NPU故障检测与恢复。</li><li>Volcano调度支持StatefulSet。</li><li>支持MindSpore框架下的亚健康热切。</li><li>训练快恢易用性增强。</li></ul>|


## 关键特性变更<a name="ZH-CN_TOPIC_0000002524562891"></a>

-   MindCluster ToolBox：新增A3 A+X PCIE全眼图测试；新增DSA随机数算子压测。
-   MindCluster基础组件：支持SGLang和vLLM的推理框架部署与重调度，一键式的任务下发；支持一体机NPU故障检测与恢复。

## 业务接口变更<a name="ZH-CN_TOPIC_0000002492443182"></a>

|特性名称|接口变更|
|--|--|
|MindIO ACP|无|
|MindIO TFT|<ul><li>删除<p>tft_register_set_stream_handler接口。</p></li><li>新增<p><ul><li>tft_get_reboot_type接口：提供给MindSpore调用，在故障重新拉起节点后，训练框架从mindio_ttp获取节点重启场景类型，进程启动后仅支持调用一次。</li><li>tft_report_load_ckpt_step接口：使用周期Checkpoint修复时，上报从Checkpoint加载的步数。</li><li>tft_register_decrypt_handler接口：用户如果开启TLS加密，使用该接口注册私钥口令解密函数。</li></ul></p></li></ul>|
|MindCluster ToolBox|新增DSA压测接口ascend-dmi -dg -i random -q。|
|MindCluster Ascend FaultDiag|无|
|MindCluster Ascend Deployer|无|
|MindCluster基础组件|新增一体机NPU故障检测与恢复操作相关参数，ranktable相关的键值定位遵循HCCL的定义。|


## 已解决的问题<a name="ZH-CN_TOPIC_0000002492283206"></a>

无。

## 遗留问题<a name="ZH-CN_TOPIC_0000002492443180"></a>

无。

# 升级影响<a name="ZH-CN_TOPIC_0000002492283208"></a>



## 升级过程对现行系统的影响<a name="ZH-CN_TOPIC_0000002524442911"></a>

无。

## 升级后对现行系统的影响<a name="ZH-CN_TOPIC_0000002492443178"></a>

无。

# 7.3.0版本配套文档<a name="ZH-CN_TOPIC_0000002524562889"></a>

|文档名称|内容简介|更新说明|
|--|--|--|
|<a href="./scheduling/installation_guide.md">《MindCluster 集群调度用户指南》</a>|提供集群调度组件说明、特性原理和使用参考，包括各组件的安装部署、集成适配示例和API参考，以及部分调度方案的原理介绍参考。|<ul><li>新增SGLang推理任务最佳实践、vLLM推理任务最佳实践、一体机特性指南等章节。</li><li>删除“使用 > 断点续训特性指南 > 通过平台使用”章节。</li></ul>其他变更详见<a href="./scheduling/installation_guide.md">《MindCluster 集群调度用户指南》</a>。|
|<a href="https://www.hiascend.com/document/detail/zh/mindcluster/730/toolbox/toolboxug/toolboxug_0002.html">《MindCluster Toolbox用户指南》</a>|提供带宽测试、算力测试、功耗测试、日志收集和软件包签名校验等功能的使用指导。|新增DSA压测命令，其他变更详见<a href="https://www.hiascend.com/document/detail/zh/mindcluster/730/toolbox/toolboxug/toolboxug_0002.html">《MindCluster Toolbox用户指南》</a>。|
|<a href="./faultdiag/installation_guide.md">《MindCluster 故障诊断用户指南》</a>|提供日志采集、日志清洗与转储、故障诊断等功能的使用指导。|新增LCN、BMC日志清洗分析、推理模型/实例级分析、清洗/诊断SDK接口等，其他变更详见<a href="./faultdiag/installation_guide.md">《MindCluster 故障诊断用户指南》</a>。|


# 漏洞修补列表<a name="ZH-CN_TOPIC_0000002524442913"></a>

请参见[MindCluster 7.3.0 漏洞修补列表.xlsx](./resource/MindCluster%207.3.0%20漏洞修补列表.xlsx)。

