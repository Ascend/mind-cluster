# 特性说明<a name="ZH-CN_TOPIC_0000002511426281"></a>

基于HDK的虚拟化实例功能是指通过资源虚拟化的方式将物理机或虚拟机配置的NPU切分成若干份vNPU（虚拟NPU）挂载到容器中使用，虚拟化管理能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放资源的操作请求。

昇腾基于HDK的虚拟化实例功能的优点是可实现多个用户按需申请共同使用一台服务器，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的平稳和安全，且资源分配，资源回收过程统一，方便多租户管理。

**原理介绍<a name="section154002962818"></a>**

昇腾NPU硬件资源主要包括AICore（用于AI模型的计算）、AICPU、内存等，基于HDK的虚拟化实例功能主要原理是将上述硬件资源根据用户指定的资源需求划分出vNPU，每个vNPU对应若干AICore、AICPU、内存资源。例如用户只需要使用4个AICore的算力，系统就会创建一个vNPU，通过vNPU向NPU芯片获取4个AICore提供给容器使用，基于HDK的虚拟化实例方案如[图1 基于HDK的虚拟化实例方案](#fig987114711574)所示。

**图 1**  基于HDK的虚拟化实例方案<a name="fig987114711574"></a>  
![](../../../../figures/scheduling/虚拟化实例方案.png "虚拟化实例方案")

**产品支持说明<a name="section17326115542216"></a>**

**表 1**  产品支持情况说明

<a name="table32786155236"></a>
<table><thead align="left"><tr id="row4278815202313"><th class="cellrowborder" valign="top" width="31.78%" id="mcps1.2.5.1.1"><p id="p22785157230"><a name="p22785157230"></a><a name="p22785157230"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.339999999999996%" id="mcps1.2.5.1.2"><p id="p7669919322"><a name="p7669919322"></a><a name="p7669919322"></a>支持的场景</p>
</th>
<th class="cellrowborder" valign="top" width="21.87%" id="mcps1.2.5.1.3"><p id="p127814159230"><a name="p127814159230"></a><a name="p127814159230"></a>虚拟化方式</p>
</th>
<th class="cellrowborder" valign="top" width="13.01%" id="mcps1.2.5.1.4"><p id="p20791155318232"><a name="p20791155318232"></a><a name="p20791155318232"></a>是否支持</p>
</th>
</tr>
</thead>
<tbody><tr id="row147414361945"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p1842320153510"><a name="p1842320153510"></a><a name="p1842320153510"></a><span id="ph118421720103512"><a name="ph118421720103512"></a><a name="ph118421720103512"></a>Atlas 推理系列产品</span></p>
<a name="ul3750195712510"></a><a name="ul3750195712510"></a><ul id="ul3750195712510"><li><span id="ph9750185716519"><a name="ph9750185716519"></a><a name="ph9750185716519"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph17500571858"><a name="ph17500571858"></a><a name="ph17500571858"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph1475016578518"><a name="ph1475016578518"></a><a name="ph1475016578518"></a>Atlas 300V Pro 视频解析卡</span></li><li><span id="ph167502575514"><a name="ph167502575514"></a><a name="ph167502575514"></a>Atlas 300I Duo 推理卡</span></li><li><span id="ph271718714435"><a name="ph271718714435"></a><a name="ph271718714435"></a>Atlas 200I SoC A1 核心板</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p11251183411474"><a name="p11251183411474"></a><a name="p11251183411474"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p753561834914"><a name="p753561834914"></a><a name="p753561834914"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p125113347470"><a name="p125113347470"></a><a name="p125113347470"></a>是</p>
</td>
</tr>
<tr id="row798113134910"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p52561887496"><a name="p52561887496"></a><a name="p52561887496"></a><span id="ph32565816491"><a name="ph32565816491"></a><a name="ph32565816491"></a>Atlas 推理系列产品</span></p>
<a name="ul12655521159"></a><a name="ul12655521159"></a><ul id="ul12655521159"><li><span id="ph12659521752"><a name="ph12659521752"></a><a name="ph12659521752"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph1651052155"><a name="ph1651052155"></a><a name="ph1651052155"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph46595214515"><a name="ph46595214515"></a><a name="ph46595214515"></a>Atlas 300V Pro 视频解析卡</span></li><li><span id="ph454745517216"><a name="ph454745517216"></a><a name="ph454745517216"></a>Atlas 200I SoC A1 核心板</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p152563874915"><a name="p152563874915"></a><a name="p152563874915"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p22562816494"><a name="p22562816494"></a><a name="p22562816494"></a>动态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p125614814491"><a name="p125614814491"></a><a name="p125614814491"></a>是</p>
</td>
</tr>
<tr id="row1327811510231"><td class="cellrowborder" rowspan="3" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph20484134417286"><a name="ph20484134417286"></a><a name="ph20484134417286"></a>Atlas 推理系列产品</span></p>
<a name="ul937113279519"></a><a name="ul937113279519"></a><ul id="ul937113279519"><li><span id="ph1837112720513"><a name="ph1837112720513"></a><a name="ph1837112720513"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph13371927759"><a name="ph13371927759"></a><a name="ph13371927759"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph73711027752"><a name="ph73711027752"></a><a name="ph73711027752"></a>Atlas 300V Pro 视频解析卡</span></li><li><span id="ph1037114272517"><a name="ph1037114272517"></a><a name="ph1037114272517"></a>Atlas 300I Duo 推理卡</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p85154811485"><a name="p85154811485"></a><a name="p85154811485"></a>在物理机划分vNPU，挂载vNPU到虚拟机</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p22781615142312"><a name="p22781615142312"></a><a name="p22781615142312"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p16791753182316"><a name="p16791753182316"></a><a name="p16791753182316"></a>是</p>
</td>
</tr>
<tr id="row11765455154717"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1470994219485"><a name="p1470994219485"></a><a name="p1470994219485"></a>在物理机划分vNPU，挂载vNPU到虚拟机，在虚拟机内将vNPU挂载到容器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p107651055174716"><a name="p107651055174716"></a><a name="p107651055174716"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p7765955134713"><a name="p7765955134713"></a><a name="p7765955134713"></a>是</p>
</td>
</tr>
<tr id="row250075974919"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p450045915490"><a name="p450045915490"></a><a name="p450045915490"></a>在物理机直通NPU到虚拟机，在虚拟机内划分vNPU，再将vNPU挂载到虚拟机内的容器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1150005964915"><a name="p1150005964915"></a><a name="p1150005964915"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p185001059104912"><a name="p185001059104912"></a><a name="p185001059104912"></a>是</p>
</td>
</tr>
<tr id="row258393195019"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p8957191110518"><a name="p8957191110518"></a><a name="p8957191110518"></a><span id="ph3957151113515"><a name="ph3957151113515"></a><a name="ph3957151113515"></a>Atlas 推理系列产品</span></p>
<a name="ul12701420650"></a><a name="ul12701420650"></a><ul id="ul12701420650"><li><span id="ph3701162014511"><a name="ph3701162014511"></a><a name="ph3701162014511"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph197019201513"><a name="ph197019201513"></a><a name="ph197019201513"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph187019209515"><a name="ph187019209515"></a><a name="ph187019209515"></a>Atlas 300V Pro 视频解析卡</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p1945835955014"><a name="p1945835955014"></a><a name="p1945835955014"></a>在物理机直通NPU到虚拟机，在虚拟机内划分vNPU，再将vNPU挂载到虚拟机内的容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p204261621515"><a name="p204261621515"></a><a name="p204261621515"></a>动态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p1458343205019"><a name="p1458343205019"></a><a name="p1458343205019"></a>是</p>
</td>
</tr>
<tr id="row0278415202314"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p6398459171311"><a name="p6398459171311"></a><a name="p6398459171311"></a><span id="ph158146714142"><a name="ph158146714142"></a><a name="ph158146714142"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p10669161183218"><a name="p10669161183218"></a><a name="p10669161183218"></a>在物理机划分vNPU，挂载vNPU到虚拟机</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p1252932516357"><a name="p1252932516357"></a><a name="p1252932516357"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p1679165352319"><a name="p1679165352319"></a><a name="p1679165352319"></a>是</p>
</td>
</tr>
<tr id="row2010035054514"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p510014508453"><a name="p510014508453"></a><a name="p510014508453"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a>Atlas 训练系列产品</span></p>
<a name="ul20127114712811"></a><a name="ul20127114712811"></a><ul id="ul20127114712811"><li><span id="ph1412724722816"><a name="ph1412724722816"></a><a name="ph1412724722816"></a>Atlas 300T 训练卡（型号 9000）</span></li><li><span id="ph1012754772811"><a name="ph1012754772811"></a><a name="ph1012754772811"></a>Atlas 300T Pro 训练卡（型号 9000）</span></li><li><span id="ph0127347172818"><a name="ph0127347172818"></a><a name="ph0127347172818"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="ph912713473289"><a name="ph912713473289"></a><a name="ph912713473289"></a>Atlas 800 训练服务器（型号 9010）</span></li><li><span id="ph012784742819"><a name="ph012784742819"></a><a name="ph012784742819"></a>Atlas 900 PoD（型号 9000）</span></li><li><span id="ph1012713477284"><a name="ph1012713477284"></a><a name="ph1012713477284"></a>Atlas 900T PoD Lite</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p710095010451"><a name="p710095010451"></a><a name="p710095010451"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p4222125217395"><a name="p4222125217395"></a><a name="p4222125217395"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p19101175084517"><a name="p19101175084517"></a><a name="p19101175084517"></a>是</p>
</td>
</tr>
<tr id="row32781215162311"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p162786153239"><a name="p162786153239"></a><a name="p162786153239"></a><span id="ph151431757142112"><a name="ph151431757142112"></a><a name="ph151431757142112"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<ul><li><span>Atlas 800T A2 训练服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p366920193216"><a name="p366920193216"></a><a name="p366920193216"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p42788151236"><a name="p42788151236"></a><a name="p42788151236"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p1154214466369"><a name="p1154214466369"></a><a name="p1154214466369"></a>是</p>
</td>
</tr>
<tr id="row11243152011236"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p18243192015230"><a name="p18243192015230"></a><a name="p18243192015230"></a><span id="ph18411121792018"><a name="ph18411121792018"></a><a name="ph18411121792018"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
<ul><li><span>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p82441020122317"><a name="p82441020122317"></a><a name="p82441020122317"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p1724417204233"><a name="p1724417204233"></a><a name="p1724417204233"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p2244122042319"><a name="p2244122042319"></a><a name="p2244122042319"></a>是</p>
</td>
</tr>
<tr id="row18359185713363"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p18176151918"><a name="p18176151918"></a><a name="p18176151918"></a><span id="ph996833614580"><a name="ph996833614580"></a><a name="ph996833614580"></a><term id="zh-cn_topic_0000001094307702_term99602034117"><a name="zh-cn_topic_0000001094307702_term99602034117"></a><a name="zh-cn_topic_0000001094307702_term99602034117"></a>Atlas A2 推理系列产品</term></span></p>
<ul><li><span>Atlas 800I A2 推理服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p1035910576364"><a name="p1035910576364"></a><a name="p1035910576364"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p1535975773615"><a name="p1535975773615"></a><a name="p1535975773615"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p143597578361"><a name="p143597578361"></a><a name="p143597578361"></a>是</p>
</td>
</tr>
<tr><td><p><span><term>Atlas A3 推理系列产品</term></span></p>
<ul><li><span>Atlas 800I A3 超节点服务器</span></li></ul>
</td>
<td><p>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td><p>静态虚拟化</p>
</td>
<td><p>是</p>
</td>
</tr>
<tr id="row188952007382"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p1746332773811"><a name="p1746332773811"></a><a name="p1746332773811"></a><span id="ph97104582114"><a name="ph97104582114"></a><a name="ph97104582114"></a><term id="zh-cn_topic_0000001519959665_term169221139190"><a name="zh-cn_topic_0000001519959665_term169221139190"></a><a name="zh-cn_topic_0000001519959665_term169221139190"></a>Atlas 200/300/500 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p1089520010381"><a name="p1089520010381"></a><a name="p1089520010381"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p188951909380"><a name="p188951909380"></a><a name="p188951909380"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p148955013384"><a name="p148955013384"></a><a name="p148955013384"></a>否</p>
</td>
</tr>
<tr id="row946362719389"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p17582910104710"><a name="p17582910104710"></a><a name="p17582910104710"></a><span id="ph5263854152111"><a name="ph5263854152111"></a><a name="ph5263854152111"></a><term id="zh-cn_topic_0000001519959665_term7466858493"><a name="zh-cn_topic_0000001519959665_term7466858493"></a><a name="zh-cn_topic_0000001519959665_term7466858493"></a>Atlas 200I/500 A2 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p194639272387"><a name="p194639272387"></a><a name="p194639272387"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p2463827143819"><a name="p2463827143819"></a><a name="p2463827143819"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p94636273386"><a name="p94636273386"></a><a name="p94636273386"></a>否</p>
</td>
</tr>
</tbody>
</table>

**使用说明<a name="section1296713336303"></a>**

- 静态虚拟化、动态虚拟化基于HDK实现，通过HDK接口将芯片切分成vNPU后，挂载到容器中使用。
- 如果使用动态虚拟化功能，请直接参见[动态虚拟化](./06_mounting_vnpu.md#动态虚拟化)章节，不需要提前使用npu-smi命令创建vNPU。
- 如果使用静态虚拟化功能，需要先参见[创建vNPU](./04_creating_vnpu.md)，再进行挂载到容器操作。

**使用约束<a name="section911013420264"></a>**

- 物理NPU虚拟化出vNPU后，不支持再将该物理NPU挂载到容器使用，也不支持再将该物理NPU直通到虚拟机使用。
- 一个vNPU只能被一个任务容器使用，不支持多个任务容器使用同一个vNPU。
- Atlas 300I Duo 推理卡上两个芯片的工作模式必须一致。即均使用虚拟化实例功能，或均整卡使用。请根据业务自行规划。
- 虚拟化实例模板是用于对整台服务器上所有NPU进行资源切分，不支持不同规格的标卡混插。如Atlas 300V Pro 视频解析卡支持24G和48G内存规格，不支持这两种内存规格的卡混插进行虚拟化；不支持30个AICore的Atlas 训练系列产品和32个AICore的Atlas 训练系列产品混插。
- 当服务器为Atlas 训练系列产品时，仅NPU芯片工作在AMP模式时支持虚拟化功能，不支持SMP模式。查询和设置NPU芯片工作模式操作步骤如下（确保服务器操作系统处于下电状态）。

    1. 登录iBMC命令行。
    2. 执行**ipmcget -d npuworkmode**命令查询NPU芯片的工作模式，若为AMP模式，则无需切换。
    3. 执行**ipmcset -d npuworkmode -v 0**命令设置NPU芯片的工作模式为AMP模式。

    查询和设置NPU芯片工作模式的详细介绍请参见《[Atlas 800 训练服务器 iBMC用户指南（型号 9000）](https://support.huawei.com/enterprise/zh/doc/EDOC1100136583)》中的“命令行介绍 \> 服务器命令 \>  [查询和设置NPU芯片工作模式（npuworkmode）](https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a)”章节。
