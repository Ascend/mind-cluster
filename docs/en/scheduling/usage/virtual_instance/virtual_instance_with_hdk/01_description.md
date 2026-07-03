# Feature Description<a name="ZH-CN_TOPIC_0000002511426281"></a>

The HDK-based virtual instance feature partitions the NPUs configured on a physical machine or virtual machine into multiple vNPUs (virtual NPUs) through resource virtualization, and mounts them into containers for use. The virtualization management enables the allocation and reclamation of resources in various specifications, accommodating repeated resource request and release operations from multiple users.

The Ascend HDK-based virtual instance function allows multiple users to share a single server on demand, lowering the entry barrier and cost of accessing NPU computing power. By enabling resource isolation through containers, this approach ensures a stable and secure runtime environment. The unified resource allocation and reclamation process also facilitates multi-tenant management.

## Principles<a name="section154002962818"></a>

Ascend NPU hardware resources mainly include AICore (used for AI model computation), AICPU, memory, etc. The main principle of the HDK-based virtual instance function is to partition the hardware resources into vNPUs according to user‑specified resource requirements, with each vNPU corresponding to a set of AICores, AICPUs, and memory resources. For example, when a user needs only 4 AICores of computing power, the system creates a vNPU, through which it obtains 4 AICores from the NPU and provides them to the container. The HDK-based virtual instance solution is shown in [Figure 1](#fig987114711574).

**Figure 1**  Virtual instance solution based on HDK<a name="fig987114711574"></a>
![](../../../../figures/scheduling/hdk-based-virtual-instance-solution.PNG)

## Products Support Notes<a name="section17326115542216"></a>

**Table 1**  Product support notes

<a name="table32786155236"></a>
<table><thead align="left"><tr id="row4278815202313"><th class="cellrowborder" valign="top" width="31.78%" id="mcps1.2.5.1.1"><p id="p22785157230"><a name="p22785157230"></a><a name="p22785157230"></a>Product Series</p>
</th>
<th class="cellrowborder" valign="top" width="33.339999999999996%" id="mcps1.2.5.1.2"><p id="p7669919322"><a name="p7669919322"></a><a name="p7669919322"></a>Supported Scenarios</p>
</th>
<th class="cellrowborder" valign="top" width="21.87%" id="mcps1.2.5.1.3"><p id="p127814159230"><a name="p127814159230"></a><a name="p127814159230"></a>Virtualization Mode</p>
</th>
<th class="cellrowborder" valign="top" width="13.01%" id="mcps1.2.5.1.4"><p id="p20791155318232"><a name="p20791155318232"></a><a name="p20791155318232"></a>Supported</p>
</th>
</tr>
</thead>
<tbody><tr id="row147414361945"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p1842320153510"><a name="p1842320153510"></a><a name="p1842320153510"></a><span id="ph118421720103512"><a name="ph118421720103512"></a><a name="ph118421720103512"></a>Atlas Inference Series Products</span></p>
<a name="ul3750195712510"></a><a name="ul3750195712510"></a><ul id="ul3750195712510"><li><span id="ph9750185716519"><a name="ph9750185716519"></a><a name="ph9750185716519"></a>Atlas 300I Pro Inference Card</span></li><li><span id="ph17500571858"><a name="ph17500571858"></a><a name="ph17500571858"></a>Atlas 300V Video Analysis Card</span></li><li><span id="ph1475016578518"><a name="ph1475016578518"></a><a name="ph1475016578518"></a>Atlas 300V Pro Video Analysis Card</span></li><li><span id="ph167502575514"><a name="ph167502575514"></a><a name="ph167502575514"></a>Atlas 300I Duo Inference Card</span></li><li><span id="ph271718714435"><a name="ph271718714435"></a><a name="ph271718714435"></a>Atlas 200I SoC A1 Core Board</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p11251183411474"><a name="p11251183411474"></a><a name="p11251183411474"></a>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p753561834914"><a name="p753561834914"></a><a name="p753561834914"></a>Static Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p125113347470"><a name="p125113347470"></a><a name="p125113347470"></a>Yes</p>
</td>
</tr>
<tr id="row798113134910"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p52561887496"><a name="p52561887496"></a><a name="p52561887496"></a><span id="ph32565816491"><a name="ph32565816491"></a><a name="ph32565816491"></a>Atlas Inference Series Products</span></p>
<a name="ul12655521159"></a><a name="ul12655521159"></a><ul id="ul12655521159"><li><span id="ph12659521752"><a name="ph12659521752"></a><a name="ph12659521752"></a>Atlas 300I Pro Inference Card</span></li><li><span id="ph1651052155"><a name="ph1651052155"></a><a name="ph1651052155"></a>Atlas 300V Video Analysis Card</span></li><li><span id="ph46595214515"><a name="ph46595214515"></a><a name="ph46595214515"></a>Atlas 300V Pro Video Analysis Card</span></li><li><span id="ph454745517216"><a name="ph454745517216"></a><a name="ph454745517216"></a>Atlas 200I SoC A1 Core Board</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p152563874915"><a name="p152563874915"></a><a name="p152563874915"></a>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p22562816494"><a name="p22562816494"></a><a name="p22562816494"></a>Dynamic Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p125614814491"><a name="p125614814491"></a><a name="p125614814491"></a>Yes</p>
</td>
</tr>
<tr id="row1327811510231"><td class="cellrowborder" rowspan="3" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph20484134417286"><a name="ph20484134417286"></a><a name="ph20484134417286"></a>Atlas Inference Series Products</span></p>
<a name="ul937113279519"></a><a name="ul937113279519"></a><ul id="ul937113279519"><li><span id="ph1837112720513"><a name="ph1837112720513"></a><a name="ph1837112720513"></a>Atlas 300I Pro Inference Card</span></li><li><span id="ph13371927759"><a name="ph13371927759"></a><a name="ph13371927759"></a>Atlas 300V Video Analysis Card</span></li><li><span id="ph73711027752"><a name="ph73711027752"></a><a name="ph73711027752"></a>Atlas 300V Pro Video Analysis Card</span></li><li><span id="ph1037114272517"><a name="ph1037114272517"></a><a name="ph1037114272517"></a>Atlas 300I Duo Inference Card</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p85154811485"><a name="p85154811485"></a><a name="p85154811485"></a>Partition vNPUs on the physical machine, and mount vNPUs to the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p22781615142312"><a name="p22781615142312"></a><a name="p22781615142312"></a>Static Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p16791753182316"><a name="p16791753182316"></a><a name="p16791753182316"></a>Yes</p>
</td>
</tr>
<tr id="row11765455154717"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1470994219485"><a name="p1470994219485"></a><a name="p1470994219485"></a>Partition vNPUs on the physical machine, mount vNPUs to the virtual machine, and then mount vNPU to  containers within the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p107651055174716"><a name="p107651055174716"></a><a name="p107651055174716"></a>Static Virtualization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p7765955134713"><a name="p7765955134713"></a><a name="p7765955134713"></a>Yes</p>
</td>
</tr>
<tr id="row250075974919"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p450045915490"><a name="p450045915490"></a><a name="p450045915490"></a>Pass through the NPU from the physical machine to the virtual machine, partition vNPUs within the virtual machine, and then mount the vNPUs to containers inside the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1150005964915"><a name="p1150005964915"></a><a name="p1150005964915"></a>Static Virtualization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p185001059104912"><a name="p185001059104912"></a><a name="p185001059104912"></a>Yes</p>
</td>
</tr>
<tr id="row258393195019"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p8957191110518"><a name="p8957191110518"></a><a name="p8957191110518"></a><span id="ph3957151113515"><a name="ph3957151113515"></a><a name="ph3957151113515"></a>Atlas Inference Series Products</span></p>
<a name="ul12701420650"></a><a name="ul12701420650"></a><ul id="ul12701420650"><li><span id="ph3701162014511"><a name="ph3701162014511"></a><a name="ph3701162014511"></a>Atlas 300I Pro Inference Card</span></li><li><span id="ph197019201513"><a name="ph197019201513"></a><a name="ph197019201513"></a>Atlas 300V Video Analysis Card</span></li><li><span id="ph187019209515"><a name="ph187019209515"></a><a name="ph187019209515"></a>Atlas 300V Pro Video Analysis Card</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p1945835955014"><a name="p1945835955014"></a><a name="p1945835955014"></a>Pass through  the NPU from the physical machine to the virtual machine, partition vNPUs within the virtual machine, and then mount the vNPUs to containers inside the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p204261621515"><a name="p204261621515"></a><a name="p204261621515"></a>Dynamic Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p1458343205019"><a name="p1458343205019"></a><a name="p1458343205019"></a>Yes</p>
</td>
</tr>
<tr id="row0278415202314"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p6398459171311"><a name="p6398459171311"></a><a name="p6398459171311"></a><span id="ph158146714142"><a name="ph158146714142"></a><a name="ph158146714142"></a>Atlas 800 Training Server</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p10669161183218"><a name="p10669161183218"></a><a name="p10669161183218"></a>Partition vNPUs on the physical machine, and mount vNPUs to the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p1252932516357"><a name="p1252932516357"></a><a name="p1252932516357"></a>Static Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p1679165352319"><a name="p1679165352319"></a><a name="p1679165352319"></a>Yes</p>
</td>
</tr>
<tr id="row2010035054514"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p510014508453"><a name="p510014508453"></a><a name="p510014508453"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a>Atlas Training Series Products</span></p>
<a name="ul20127114712811"></a><a name="ul20127114712811"></a><ul id="ul20127114712811"><li><span id="ph1412724722816"><a name="ph1412724722816"></a><a name="ph1412724722816"></a>Atlas 300T Training Card (Model 9000)</span></li><li><span id="ph1012754772811"><a name="ph1012754772811"></a><a name="ph1012754772811"></a>Atlas 300T Pro Training Card (Model 9000)</span></li><li><span id="ph0127347172818"><a name="ph0127347172818"></a><a name="ph0127347172818"></a>Atlas 800 Training Server (Model 9000)</span></li><li><span id="ph912713473289"><a name="ph912713473289"></a><a name="ph912713473289"></a>Atlas 800 Training Server (Model 9010)</span></li><li><span id="ph012784742819"><a name="ph012784742819"></a><a name="ph012784742819"></a>Atlas 900 PoD (Model 9000)</span></li><li><span id="ph1012713477284"><a name="ph1012713477284"></a><a name="ph1012713477284"></a>Atlas 900T PoD Lite</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p710095010451"><a name="p710095010451"></a><a name="p710095010451"></a>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p4222125217395"><a name="p4222125217395"></a><a name="p4222125217395"></a>Static Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p19101175084517"><a name="p19101175084517"></a><a name="p19101175084517"></a>Yes</p>
</td>
</tr>
<tr id="row32781215162311"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p162786153239"><a name="p162786153239"></a><a name="p162786153239"></a><span id="ph151431757142112"><a name="ph151431757142112"></a><a name="ph151431757142112"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 Training Series Products</term></span></p>
<ul><li><span>Atlas 800T A2 Training Server (24 AICores)</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p366920193216"><a name="p366920193216"></a><a name="p366920193216"></a>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><ul><li>Static Virtualization</li><li>Dynamic Virtualization</li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p1154214466369"><a name="p1154214466369"></a><a name="p1154214466369"></a>Yes</p>
</td>
</tr>
<tr id="row11243152011236"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p18243192015230"><a name="p18243192015230"></a><a name="p18243192015230"></a><span id="ph18411121792018"><a name="ph18411121792018"></a><a name="ph18411121792018"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 Training Series Products</term></span></p>
<ul><li><span>Atlas 800T A3 SuperPoD Server</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p82441020122317"><a name="p82441020122317"></a><a name="p82441020122317"></a>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><ul><li>Static Virtualization</li><li>Dynamic Virtualization</li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p2244122042319"><a name="p2244122042319"></a><a name="p2244122042319"></a>Yes</p>
</td>
</tr>
<tr id="row18359185713363"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p18176151918"><a name="p18176151918"></a><a name="p18176151918"></a><span id="ph996833614580"><a name="ph996833614580"></a><a name="ph996833614580"></a><term id="zh-cn_topic_0000001094307702_term99602034117"><a name="zh-cn_topic_0000001094307702_term99602034117"></a><a name="zh-cn_topic_0000001094307702_term99602034117"></a>Atlas A2 Inference Series Products</term></span></p>
<ul><li><span>Atlas 800I A2 Inference Server</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p1035910576364"><a name="p1035910576364"></a><a name="p1035910576364"></a>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><ul><li>Static Virtualization</li><li>Dynamic Virtualization</li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p143597578361"><a name="p143597578361"></a><a name="p143597578361"></a>Yes</p>
</td>
</tr>
<tr><td><p><span><term>Atlas A3 Inference Series Products</term></span></p>
<ul><li><span>Atlas 800I A3 SuperPoD Server</span></li></ul>
</td>
<td><p>Partition vNPUs on the physical machine, and mount vNPUs to the container.</p>
</td>
<td><ul><li>Static Virtualization</li><li>Dynamic Virtualization</li></ul>
</td>
<td><p>Yes</p>
</td>
</tr>
<tr id="row188952007382"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p1746332773811"><a name="p1746332773811"></a><a name="p1746332773811"></a><span id="ph97104582114"><a name="ph97104582114"></a><a name="ph97104582114"></a><term id="zh-cn_topic_0000001519959665_term169221139190"><a name="zh-cn_topic_0000001519959665_term169221139190"></a><a name="zh-cn_topic_0000001519959665_term169221139190"></a>Atlas 200/300/500 Inference Products</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p1089520010381"><a name="p1089520010381"></a><a name="p1089520010381"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p188951909380"><a name="p188951909380"></a><a name="p188951909380"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p148955013384"><a name="p148955013384"></a><a name="p148955013384"></a>No</p>
</td>
</tr>
<tr id="row946362719389"><td class="cellrowborder" valign="top" width="31.78%" headers="mcps1.2.5.1.1 "><p id="p17582910104710"><a name="p17582910104710"></a><a name="p17582910104710"></a><span id="ph5263854152111"><a name="ph5263854152111"></a><a name="ph5263854152111"></a><term id="zh-cn_topic_0000001519959665_term7466858493"><a name="zh-cn_topic_0000001519959665_term7466858493"></a><a name="zh-cn_topic_0000001519959665_term7466858493"></a>Atlas 200I/500 A2 Inference Products</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="33.339999999999996%" headers="mcps1.2.5.1.2 "><p id="p194639272387"><a name="p194639272387"></a><a name="p194639272387"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="21.87%" headers="mcps1.2.5.1.3 "><p id="p2463827143819"><a name="p2463827143819"></a><a name="p2463827143819"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.01%" headers="mcps1.2.5.1.4 "><p id="p94636273386"><a name="p94636273386"></a><a name="p94636273386"></a>No</p>
</td>
</tr>
</tbody>
</table>

## Usage Instructions<a name="section1296713336303"></a>

- Static virtualization and dynamic virtualization are implemented based on HDK. The NPU is partitioned into vNPUs through HDK interfaces and then mounted to containers for use.
- If you are using dynamic virtualization, directly refer to the "Dynamic vNPU Scheduling (Inference)" section. You do not need to use the `npu-smi` command to create vNPUs in advance.
- If you are using static virtualization, you need to first refer to "Creating vNPU", and then perform the operation of mounting them to a container.
- For detailed descriptions of `npu-smi` commands, see the "[Ascend Virtual Instance (AVI) Commands](https://support.huawei.com/enterprise/en/doc/EDOC1100568420/690dda6e)" section in the *Atlas A3 Center Inference and Training Hardware 26.0.RC1 npu-smi Command Reference*.

## Usage Constraints<a name="section911013420264"></a>

- After a physical NPU is virtualized into vNPUs, it is no longer supported to mount the physical NPU to a container, nor is it supported to pass through the physical NPU to a virtual machine.
- A vNPU can only be used by one task container. It is not supported for multiple task containers to use the same vNPU.
- The operating modes of the two chips on the Atlas 300I Duo inference card must be consistent. That is, both chips must use the virtualization instance function, or the entire card must use this function. Plan according to your service requirements.
- The virtual instance template is used to partition resources across all NPUs on an entire server. Mixing cards of different specifications is not supported. For example, the Atlas 300V Pro video analysis card supports 24 GB and 48 GB memory specifications, virtualization does not support mixing these two memory specifications on the same server. Similarly, mixing Atlas training series products with 30 AICores and those with 32 AICores is not supported.
- For Atlas training series servers, the virtual instance function is supported only when the NPU operates in AMP mode, not in SMP mode. The steps for querying and setting the NPU working mode are as follows (ensure the server operating system is powered off).

    1. Log in to the iBMC command line.
    2. Run the `ipmcget -d npuworkmode` command to query the NPU working mode. If it is AMP mode, no switching is required.
    3. Run the `ipmcset -d npuworkmode -v 0` command to set the NPU working mode to AMP mode.

For detailed instructions on querying and setting the NPU working mode, see the "Command Line Introduction > Server Commands > [Querying and Setting the NPU Working Mode (npuworkmode)](https://support.huawei.com/enterprise/en/doc/EDOC1100136580/51fd0a0/querying-and-setting-the-npu-chip-working-mode-npuworkmode?idPath=23710424|251366513|22892968|252309113|250702818)" in the *[Atlas 800 Training Server iBMC User Guide (model 9000)](https://support.huawei.com/enterprise/en/doc/EDOC1100136580/426cffd9/about-this-document?idPath=23710424|251366513|22892968|252309113|250702818)*.
