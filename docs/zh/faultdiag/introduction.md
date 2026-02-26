# 前言<a name="ZH-CN_TOPIC_0000002017481558"></a>

**免责声明<a name="zh-cn_topic_0000001461453746_section582264385620"></a>**

-   本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。
-   本特性会读取并处理输入目录下用户采集的相关原日志及监测指标文件，请用户确保相关文件内无敏感信息和个人数据。华为不控制且不对输入数据内容承担任何责任。

-   用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。

**读者对象<a name="section55821246876"></a>**

本文档主要适用于以下人员：

-   华为技术支持工程师
-   渠道伙伴技术支持工程师
-   企业管理员
-   企业终端用户

# 简介<a name="ZH-CN_TOPIC_0000002005252541"></a>

MindCluster Ascend FaultDiag（故障诊断工具）主要功能如下：提供日志清洗和故障诊断功能，提取训练及推理过程相关日志的关键信息，并根据集群所有节点清洗后的关键信息，分析故障根因节点以及故障事件。

**关键功能特性<a name="section14223166131518"></a>**

故障诊断工具主要提供以下2大功能：

**日志清洗**

当训练及推理任务失败后，由MindCluster Ascend FaultDiag对原日志和监测指标信息进行一系列清洗工作，清洗结果随原始信息一并转储到同一路径，为诊断任务提供数据支持。

目前支持清洗的内容主要包括：用户训练及推理日志、CANN应用类日志、主机侧资源信息、NPU网口资源信息等原始训练、推理日志及监测指标清洗。

**故障诊断**

MindCluster Ascend FaultDiag针对下表中的2类问题提供诊断功能：支持根因节点分析、故障事件分析、设备资源分析、网络拥塞分析。

|故障分类|诊断内容|
|--|--|
|训练及推理任务异常退出问题|<ul><li>根因节点分析用于分析故障根因节点：根据集群通信HCCL报错信息，定位引发错误的根因节点。</li><li>故障事件分析用于分析故障根因节点所在设备的根因错误：根据故障知识图谱包含的故障模式，分析节点所在设备的问题。</li></ul>|
|训练及推理过程中性能劣化问题|<ul><li>设备资源分析用于分析设备的资源状态：通过分析用户采集的设备相关指标文件，定位计算降频与CPU资源抢占等问题。</li><li>网络拥塞分析用于分析节点间的网络状态：通常用于定位Spine+Leaf双层组网场景下的网络问题，通过分析用户采集的NPU网口统计监测指标文件，分析是否发生节点链路网络拥塞异常问题。</li></ul>|
|说明：仅当训练及推理任务未异常退出的情况下，才会对性能劣化问题进行诊断。|


**使用流程<a name="section7779135035518"></a>**

MindCluster Ascend FaultDiag的使用流程如下表所示。

|步骤|描述|操作参考|
|--|--|--|
|日志采集|<p>当发现训练或推理任务失败或异常后，用户需收集每台训练或推理设备的日志，并按预设结构进行存储。</p><p>采集的日志请参见<a href="#应用场景及方案">应用场景及方案</a>中“训练及推理任务日志与指标信息参考表”。</p>|请参见<a href="./user_guide.md#日志采集">日志采集</a>章节。|
|日志清洗|日志收集完成后，用户需在每台训练或推理设备上使用MindCluster Ascend FaultDiag工具的清洗功能对收集的原始日志及指标数据进行清洗，过滤并提取有效信息。|请参见<a href="./user_guide.md#日志清洗与转储">日志清洗与转储</a>章节。|
|清洗结果转储|日志清洗完成后，用户需将每台训练或推理设备的清洗结果转储汇总到同一台训练设备或通用设备，并按预设结构进行存储。|请参见<a href="./user_guide.md#日志清洗与转储">日志清洗与转储</a>章节。|
|故障诊断|基于转储汇总的清洗结果，用户使用MindCluster Ascend FaultDiag工具的诊断功能分析训练或推理任务失败或异常的故障根因。|请参见<a href="./user_guide.md#故障诊断">故障诊断</a>章节。|
|说明：在以上使用流程中，日志采集和清洗结果转储非MindCluster Ascend FaultDiag提供的功能。本文档仅提供其操作指导。|

# 应用场景及方案<a name="ZH-CN_TOPIC_0000001592414013"></a>

智能故障诊断特性可用于解决集群训练及推理任务中故障定位定界困难的问题。由于集群日志量大、AI全栈日志分析复杂，同时也可能涉及计算、网络、存储的跨域问题分析，用户在集群训练及推理业务中遇到的问题通常定位难度大、耗时长、需要拉通的领域多。

智能故障诊断特性可以有效提高训练及推理作业问题定位能力，提升用户使用意愿和促进产品生态拓展。

具体而言，本特性针对训练及推理集群的各个设备提供日志清洗和故障诊断功能，用户需要完成日志采集、清洗并将清洗后的信息文件转储至特定路径进行诊断，通过分析诊断结果实现问题的快速定界。同时，支持用户自定义故障实体或对CANN应用类日志中的ERROR报错信息进行屏蔽。

根据实际业务的不同，当前主要分为以下2种场景。

|场景名称|用户|任务类型|特点|
|--|--|--|--|
|<a href="#section1511514596338">全量应用场景</a>|企业、政府、事业单位等（具有AI集群运维平台能力）|训练任务、推理任务|依赖于训练及推理、CANN和主机侧资源以及硬件相关数据，采集内容较复杂，适用于AI集群运维平台用户进行复杂的任务诊断。|
|<a href="#section587911381388">基础应用场景</a>|个人|训练任务、推理任务|仅依赖训练或推理日志与CANN日志，采集内容和方法简单，适用于个人用户进行基础任务诊断。|


**全量应用场景<a name="section1511514596338"></a>**

全量应用场景在**训练业务场景下**需依赖于训练日志、主机资源日志、NPU日志以及硬件日志等多类日志、指标数据信息。

**推理场景下**需依赖于“推理任务日志”、“CANN应用类日志”、“Device侧日志”、“MindIE组件日志”。

其中部分指标数据需通过额外操作采集获取，因此，全量应用场景推荐具有AI集群运维平台能力的用户对接使用。

全量应用场景方案以下图为例，用户需要在所有训练或推理设备中安装MindCluster Ascend FaultDiag组件，当训练或推理任务结束后，每台设备需要收集上述所有日志及指标数据信息，然后使用MindCluster Ascend FaultDiag的清洗功能进行过滤、提取有效信息，最后将所有设备的原始日志及指标信息和清洗结果转储到AI集群运维平台，平台使用MindCluster Ascend FaultDiag的诊断功能分析故障根因。同时，支持用户自定义故障实体或对CANN应用类日志中的ERROR报错信息进行屏蔽。

**图 1**  全量应用场景方案<a name="fig21091944182810"></a>  
![](../figures/faultdiag/全量应用场景方案.png "全量应用场景方案")

在全量场景中，待采集日志及指标数据信息对应的数据来源以及数据用途如[表1](#table7211162233417)所示。

**表 1**  训练及推理任务日志与指标信息参考表

<a name="table7211162233417"></a>
|**数据类别**|**日志说明**|数据来源|**数据用途**|
|--|--|--|--|
|训练任务日志|模型训练进程产生的日志|训练任务|用于故障事件分析|
|训练前后NPU网口检查|在执行训练任务前后，使用hccn_tool工具检查每个NPU的网口信息|训练任务|用于故障事件分析|
|主机侧资源信息|NPU状态监测等指标，包括各NPU卡训练主进程使用的CPU占比（%CPU）、使用的物理内存（RES）等信息|训练任务|用于设备资源分析|
|NPU网口资源信息|NPU网口收发报文统计信息等指标|训练任务|用于网络拥塞分析|
|OS日志|Linux系统日志|训练任务|用于故障事件分析|
|MindCluster组件日志|MindCluster的组件Ascend Device Plugin、NodeD、Ascend Docker Runtime、NPU Exporter、Volcano采集的超节点设备日志、AI服务器日志、组件日志。|训练任务|用于故障事件分析|
|推理任务日志|推理任务进程产生的日志|推理任务|用于故障事件分析|
|NPU Device运行日志|Device侧的相关日志和文件，包括slog日志、hisi_logs日志等|训练&推理任务|用于故障事件分析|
|CANN应用类日志|CANN软件产生的运行日志|训练&推理任务|用于根因节点分析与故障事件分析|
|MindIE组件日志|MindIE的组件MindIE Server、MindIE LLM、MindIE SD、MindIE RT、MindIE Torch、MindIE MS、MindIE Benchmark、MindIE Client产生的日志|推理任务|用于故障事件分析|
|AMCT组件日志|AMCT进行模型压缩流程产生的日志|模型压缩|用于AMCT工具故障事件分析|
|MindIE Pod控制台日志|MindIE Pod控制台日志|推理任务|用于根因节点分析|
|MindIO组件日志|MindIO组件产生的日志|训练&推理任务|用于故障事件分析|


>[!NOTE] 说明 
>所有日志及指标数据对应的采集方法参见[日志采集](./user_guide.md#日志采集)章节。

**基础应用场景<a name="section587911381388"></a>**

考虑到不同用户的应用需求，支持仅依赖于训练或推理日志及CANN应用类日志进行诊断的基础应用场景。此类日志为训练或推理任务产生，无需额外采集。

基础应用场景方案以下图为例，用户需要在所有训练或推理设备中安装MindCluster Ascend FaultDiag组件。当训练或推理任务结束后，每台设备需要至少收集训练或推理日志和CANN应用类日志，然后使用MindCluster Ascend FaultDiag的清洗功能过滤、提取有效信息，最后将所有设备的原始日志和清洗结果转储到同一通用设备上，用户使用MindCluster Ascend FaultDiag的诊断功能分析故障根因。同时，支持用户自定义故障实体或对CANN应用类日志中的ERROR报错信息进行屏蔽。

**图 2**  基础应用场景方案<a name="fig3750917713"></a>  
![](../figures/faultdiag/基础应用场景方案.png "基础应用场景方案")

待采集日志及指标数据信息对应的数据来源以及数据用途如下表所示。

所有日志及指标数据对应的采集方法参见[日志采集](./user_guide.md#日志采集)章节。

**表 2**  训练&推理任务日志与指标信息参考表

|数据类别|日志说明|数据来源|数据用途|
|--|--|--|--|
|训练任务日志|训练任务进程产生的日志|训练任务|用于故障事件分析|
|CANN应用类日志|CANN软件产生的运行日志|训练&推理任务|用于根因节点分析与故障事件分析|
|推理任务日志|推理任务进程产生的日志|推理任务|用于故障事件分析|

# 支持的产品形态<a name="ZH-CN_TOPIC_0000001888518362"></a>

MindCluster Ascend FaultDiag组件旨在昇腾训练及推理任务下，为用户提供高效定位任务失败与训练及推理性能劣化问题的能力。支持的产品如[表1](#table557565881416)所示。

**表 1**  产品说明

<a name="table557565881416"></a>
<table><thead align="left"><tr id="row20575358111415"><th class="cellrowborder" valign="top" width="31.89%" id="mcps1.2.3.1.1"><p id="p17575115810143"><a name="p17575115810143"></a><a name="p17575115810143"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="68.11%" id="mcps1.2.3.1.2"><p id="p873817148518"><a name="p873817148518"></a><a name="p873817148518"></a>产品名称</p>
</th>
</tr>
</thead>
<tbody><tr id="row1357695881416"><td class="cellrowborder" rowspan="6" valign="top" width="31.89%" headers="mcps1.2.3.1.1 "><p id="p19576185814144"><a name="p19576185814144"></a><a name="p19576185814144"></a><span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas 训练系列产品</span></p>
<p id="p187773141516"><a name="p187773141516"></a><a name="p187773141516"></a></p>
</td>
<td class="cellrowborder" valign="top" width="68.11%" headers="mcps1.2.3.1.2 "><p id="p0253711155"><a name="p0253711155"></a><a name="p0253711155"></a><span id="ph82143771517"><a name="ph82143771517"></a><a name="ph82143771517"></a>Atlas 300T 训练卡（型号 9000）</span></p>
</td>
</tr>
<tr id="row4186152131114"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p14186122151111"><a name="p14186122151111"></a><a name="p14186122151111"></a><span id="ph194596387512"><a name="ph194596387512"></a><a name="ph194596387512"></a>Atlas 300T Pro 训练卡（型号 9000）</span></p>
</td>
</tr>
<tr id="row8497172661914"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1289216213517"><a name="p1289216213517"></a><a name="p1289216213517"></a><span id="ph941317566562"><a name="ph941317566562"></a><a name="ph941317566562"></a>Atlas 800 训练服务器（型号 9000）</span></p>
</td>
</tr>
<tr id="row18273132861913"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p8829623165112"><a name="p8829623165112"></a><a name="p8829623165112"></a><span id="ph1627888115712"><a name="ph1627888115712"></a><a name="ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span></p>
</td>
</tr>
<tr id="row257635861416"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p423372151"><a name="p423372151"></a><a name="p423372151"></a><span id="ph2210374151"><a name="ph2210374151"></a><a name="ph2210374151"></a>Atlas 900 PoD（型号 9000）</span></p>
</td>
</tr>
<tr id="row4876103116158"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p62113791512"><a name="p62113791512"></a><a name="p62113791512"></a><span id="ph1021637121517"><a name="ph1021637121517"></a><a name="ph1021637121517"></a>Atlas 900T PoD Lite</span></p>
</td>
</tr>
<tr id="row1576105819146"><td class="cellrowborder" rowspan="3" valign="top" width="31.89%" headers="mcps1.2.3.1.1 "><p id="p1581214200339"><a name="p1581214200339"></a><a name="p1581214200339"></a><span id="ph44851267332"><a name="ph44851267332"></a><a name="ph44851267332"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="68.11%" headers="mcps1.2.3.1.2 "><p id="p48441383162"><a name="p48441383162"></a><a name="p48441383162"></a><span id="ph158441387167"><a name="ph158441387167"></a><a name="ph158441387167"></a>Atlas 200T A2 Box16 异构子框</span></p>
</td>
</tr>
<tr id="row63611221615"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1084448141616"><a name="p1084448141616"></a><a name="p1084448141616"></a><span id="ph108441781168"><a name="ph108441781168"></a><a name="ph108441781168"></a>Atlas 800T A2 训练服务器</span></p>
</td>
</tr>
<tr id="row147611836168"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p184414841616"><a name="p184414841616"></a><a name="p184414841616"></a><span id="ph1584468101616"><a name="ph1584468101616"></a><a name="ph1584468101616"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
</tr>
<tr id="row8419734175712"><td class="cellrowborder" valign="top" width="31.89%" headers="mcps1.2.3.1.1 "><p id="p041912343573"><a name="p041912343573"></a><a name="p041912343573"></a><span id="ph19590185162111"><a name="ph19590185162111"></a><a name="ph19590185162111"></a>Atlas 推理系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="68.11%" headers="mcps1.2.3.1.2 "><p id="p46292041163911"><a name="p46292041163911"></a><a name="p46292041163911"></a><span id="ph1653016814441"><a name="ph1653016814441"></a><a name="ph1653016814441"></a>Atlas 300I Duo 推理卡</span></p>
</td>
</tr>
<tr id="row94201343574"><td class="cellrowborder" rowspan="3" valign="top" width="31.89%" headers="mcps1.2.3.1.1 "><p id="p18176151918"><a name="p18176151918"></a><a name="p18176151918"></a><span id="ph382710211428"><a name="ph382710211428"></a><a name="ph382710211428"></a>Atlas 800I A2推理产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="68.11%" headers="mcps1.2.3.1.2 "><p id="p983193816410"><a name="p983193816410"></a><a name="p983193816410"></a><span id="ph1772425892511"><a name="ph1772425892511"></a><a name="ph1772425892511"></a>Atlas 800I A2 推理服务器（32GB HCCS款）</span></p>
</td>
</tr>
<tr id="row189661824234"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1756516448410"><a name="p1756516448410"></a><a name="p1756516448410"></a><span id="ph158932819263"><a name="ph158932819263"></a><a name="ph158932819263"></a>Atlas 800I A2 推理服务器（32GB PCIe款）</span></p>
</td>
</tr>
<tr id="row8341240171220"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p135540131214"><a name="p135540131214"></a><a name="p135540131214"></a><span id="ph144486301267"><a name="ph144486301267"></a><a name="ph144486301267"></a>Atlas 800I A2 推理服务器（64GB HCCS款）</span></p>
</td>
</tr>
<tr id="row1394247192215"><td class="cellrowborder" rowspan="5" valign="top" width="31.89%" headers="mcps1.2.3.1.1 "><p id="p1325452716187"><a name="p1325452716187"></a><a name="p1325452716187"></a><span id="ph18411121792018"><a name="ph18411121792018"></a><a name="ph18411121792018"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="68.11%" headers="mcps1.2.3.1.2 "><p id="p49581614112219"><a name="p49581614112219"></a><a name="p49581614112219"></a><span id="ph2822112219"><a name="ph2822112219"></a><a name="ph2822112219"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
</tr>
<tr id="row19768174911227"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1801101613220"><a name="p1801101613220"></a><a name="p1801101613220"></a><span id="ph1653373317227"><a name="ph1653373317227"></a><a name="ph1653373317227"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></p>
</td>
</tr>
<tr id="row15338530181016"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p12338163012109"><a name="p12338163012109"></a><a name="p12338163012109"></a><span id="ph1182812426107"><a name="ph1182812426107"></a><a name="ph1182812426107"></a>Atlas 800I A3 超节点服务器</span></p>
</td>
</tr>
<tr id="row36947612111"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p369406181115"><a name="p369406181115"></a><a name="p369406181115"></a><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></p>
</td>
</tr>
<tr id="row1045214921112"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p154531897117"><a name="p154531897117"></a><a name="p154531897117"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
</tr>
</tbody>
</table>

