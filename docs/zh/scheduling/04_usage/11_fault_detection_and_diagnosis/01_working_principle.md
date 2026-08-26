# 实现原理

## 故障检测整体架构

MindCluster集群调度组件Ascend Device Plugin提供NPU芯片故障检测能力及NPU参数面网络故障检测能力，NodeD提供服务器节点故障、共享存储故障和灵衢慢网络故障检测能力，ClusterD提供公共故障检测能力，Volcano提供业务面容器异常检测能力，故障检测整体架构如下图所示。

![](../../../figures/scheduling/250411110432760.png)

1. 计算服务器上的Ascend Device Plugin通过驱动获取NPU芯片故障以及参数面网络故障后，将故障信息上报到管理服务器。
2. 计算服务器上的NodeD通过驱动获取服务器节点故障、共享存储故障和灵衢慢网络故障信息后，将故障信息上报到管理服务器。
3. 计算服务器上的K8s监测训练容器状态，训练容器异常后上报到K8s中，管理服务器上的Volcano通过K8s获取训练容器的故障信息。
4. 管理服务器上的ClusterD通过公共故障接口获取公共故障后，将接收到的信息进行汇总写入cluster-info-device-cm。
5. （可选）管理服务器上的ClusterD汇总集群内所有Ascend Device Plugin和NodeD上报的故障信息。

## ConfigMap说明

- 每个计算节点的Ascend Device Plugin均会创建记录本节点NPU和灵衢总线设备信息的ConfigMap文件。该ConfigMap文件名为mindx-dl-deviceinfo-&lt;nodename&gt;（以下简称device-info-cm），故障信息会通过该ConfigMap进行上报。该ConfigMap文件中各字段的说明，请参见[DeviceInfoCfg](../../06_api/02_ascend_device_plugin.md#芯片资源)表。
- 当节点上存在节点故障时，每个计算节点的NodeD会创建记录本节点设备信息的ConfigMap文件。该ConfigMap文件名为mindx-dl-nodeinfo-&lt;nodename&gt;（以下简称node-info-cm），节点故障信息会通过该ConfigMap进行上报。该ConfigMap文件中各字段的说明，请参见[mindx-dl-nodeinfo-&lt;nodename&gt;](../../06_api/03_noded.md#节点资源)表。
- ClusterD会创建记录本集群设备信息的ConfigMap文件，该ConfigMap文件名为cluster-info-<device/switch>-<[0-5]>、cluster-info-node-cm（以下简称cluster-info-cm）。节点及芯片故障信息会通过[cluster-info-cm](../../06_api/04_clusterd/00_cluster_resources.md)进行上报。
- 创建每个任务时，需要在YAML中配置ConfigMap文件，该ConfigMap文件名称为reset-config-&lt;job-name&gt;（以下简称reset-info-cm）。该ConfigMap挂载到容器的“/user/restore/reset/config”路径下。Ascend Device Plugin会自动将ConfigMap挂载到本节点的“/user/restore/reset/&lt;job-namespace&gt;.&lt;job-name&gt;”路径下。

    也可以将节点上/user/restore/reset/&lt;job-namespace&gt;.&lt;job-name&gt;替代ConfigMap，挂载到容器的“/user/restore/reset/config”路径下。该ConfigMap文件字段说明，请参见[reset-config-&lt;job-name&gt;](../../06_api/02_ascend_device_plugin.md#任务信息)表。

## 在线压测<a name="ZH-CN_TOPIC_0000002479226572"></a>

MindCluster支持训练在线压测特性，即在训练过程中可以调用在线压测接口，暂停指定训练任务，对任务使用的节点进行硬件P2P或AIC压力测试，主动判断节点是否存在异常。若不存在故障则恢复训练；若存在故障则隔离故障节点，触发断点续训。

### 使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>

- 对于PyTorch训练框架，需配合MindSpeed-LLM  2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配合MindFormers master版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 请在训练正常迭代后，再进行在线压测指令的下发。
- 确保已开启进程级恢复相关功能特性。
- 压测过程中不支持重启ClusterD，如果ClusterD异常重启，需要重启训练并下发压测任务。
- 压测过程中，需要关闭热复位功能。
- P2P压测需确保device侧有10GB以上的空闲内存。
- 对于MindSpore训练框架，需要在启动TaskD Manager前设置export TASKD\_PROCESS\_ENABLE="on"。
- 本功能依赖MindIO组件，使用前请先了解MindIO的[约束限制](../../07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md#约束限制)。

### 支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>

**表 1**  在线压测支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416_benchmark"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="25.172517251725168%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="43.834383438343835%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="30.993099309930994%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="25.172517251725168%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p773278122616"><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="43.834383438343835%" headers="mcps1.2.4.1.2 "><p id="p17354133423610"><a name="p17354133423610"></a><span id="ph14314162316427"><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.993099309930994%" headers="mcps1.2.4.1.3 "><a name="ul15879359132214"></a><ul id="ul15879359132214"><li><span id="ph135835207394"><a name="ph135835207394"></a>MindSpore</span></li><li><span id="ph19425111582712"><a name="ph19425111582712"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="25.172517251725168%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="ph531432344210"><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="43.834383438343835%" headers="mcps1.2.4.1.2 "><p id="p4897194703620"><a name="p4897194703620"></a><span id="ph077885871817"><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.993099309930994%" headers="mcps1.2.4.1.3 "><a name="ul13821123132320"></a><ul id="ul13821123132320"><li><span id="ph19127156230"><a name="ph19127156230"></a>MindSpore</span></li><li><span id="ph310231710274"><a name="ph310231710274"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

### 在线压测原理<a name="section56986212179"></a>

**图 1**  原理图<a name="fig9336113210132"></a>

![](../../../figures/scheduling/原理图-10.png "原理图-10")

在以上原理图中，各个步骤的说明如下。

1. AI平台集成ClusterD，调用ClusterD的gRPC接口下发压测操作，指定需要压测的节点。
2. ClusterD通知MindIO暂停训练。
3. TaskD Manager通知指定TaskD Worker调用训练框架接口执行压测操作。
4. 训练框架调用指定NPU卡上的CANN接口执行压测操作。
5. ClusterD判断指定NPU卡的压测操作完成后，再由TaskD通知MindIO在压测完成后继续执行下一个Step训练。

### 适配功能点<a name="section1446615300284"></a>

在在线压测中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。通过主动调用优雅暂停机制，完成当前卡上任务暂停，暂停后进行硬件压力测试，测试完成后继续训练。集群大脑需提供对外接口，接收压测指令并管理压测流程。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table19955141136103)的功能适配。

**表 2**  在线压测框架适配功能点

<a name="table19955141136103"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.98%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="39.26%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.01%" id="mcps1.2.5.1.3"><p id="p106021527183014"><a name="p106021527183014"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="23.75%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.98%" headers="mcps1.2.5.1.1 "><p id="p0609650313"><a name="p0609650313"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="39.26%" headers="mcps1.2.5.1.2 "><p id="p195191085319"><a name="p195191085319"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="18.01%" headers="mcps1.2.5.1.3 "><p id="p1855311819317"><a name="p1855311819317"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.75%" headers="mcps1.2.5.1.4 "><p id="p10701822403"><a name="p10701822403"></a><a href="../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1960918515317"><a name="p1960918515317"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p156091533120"><a name="p156091533120"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p76093513110"><a name="p76093513110"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p136091519311"><a name="p136091519311"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row46026594305"><td class="cellrowborder" valign="top" width="18.98%" headers="mcps1.2.5.1.1 "><p id="p26091514318"><a name="p26091514318"></a>在线压测过程管理</p>
</td>
<td class="cellrowborder" valign="top" width="39.26%" headers="mcps1.2.5.1.2 "><p id="p14609155183114"><a name="p14609155183114"></a>提供在线压测请求下发能力，控制训练进程暂停与恢复。</p>
</td>
<td class="cellrowborder" valign="top" width="18.01%" headers="mcps1.2.5.1.3 "><p id="p6553121803118"><a name="p6553121803118"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="23.75%" headers="mcps1.2.5.1.4 "><p id="p1660265933015"><a name="p1660265933015"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.1.0/component/clusterd/pkg/application/recover/om_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>
