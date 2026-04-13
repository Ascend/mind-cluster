# 应用场景及方案<a name="ZH-CN_TOPIC_0000002511426823"></a>

**应用场景<a name="section198715461917"></a>**

基于HDK的虚拟化实例功能适用于多用户多任务并行，且每个任务算力需求较小的场景。对算力需求较大的大模型任务，不支持使用昇腾虚拟化实例。

**虚拟化场景<a name="section1618382307"></a>**

昇腾虚拟化实例功能在物理机或虚拟机使用时，支持以下虚拟化场景，如[表1](#table197838103018)所示。本文主要介绍在昇腾设备划分vNPU支持的场景和方法，如果涉及虚拟机相关的配置，需要结合另一本文档《Atlas 系列硬件产品 25.5.0 虚拟机配置指南》的“安装虚拟机\>配置NPU直通虚拟机\>[NPU直通虚拟机](https://support.huawei.com/enterprise/zh/doc/EDOC1100540506/2689d3e6)”章节一起使用。

划分vNPU有以下两种方式。

- 静态虚拟化：通过npu-smi工具**手动**创建多个vNPU。物理机和虚拟机场景均支持静态虚拟化。
- 动态虚拟化：通过软件配置，在收到虚拟化任务请求后，动态地**自动**创建vNPU、挂载任务、回收vNPU。

**表 1**  使用场景

<a name="table197838103018"></a>
<table><thead align="left"><tr id="row16723873015"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p871338103019"><a name="p871338103019"></a><a name="p871338103019"></a>昇腾虚拟化实例功能支持场景</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p14014521402"><a name="p14014521402"></a><a name="p14014521402"></a>操作流程</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p18893873015"><a name="p18893873015"></a><a name="p18893873015"></a>支持的虚拟化方式</p>
</th>
</tr>
</thead>
<tbody><tr id="row158123818304"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1819384303"><a name="p1819384303"></a><a name="p1819384303"></a>在物理机划分vNPU，挂载vNPU到虚拟机</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1290518155817"><a name="p1290518155817"></a><a name="p1290518155817"></a>在物理机划分vNPU和挂载vNPU到虚拟机的步骤请参见<span id="ph15232948195013"><a name="ph15232948195013"></a><a name="ph15232948195013"></a>《Atlas 系列硬件产品 25.5.0 虚拟机配置指南》的“安装虚拟机&gt;配置NPU直通虚拟机&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100540506/bf80825c" target="_blank" rel="noopener noreferrer">vNPU直通虚拟机</a>”章节</span>。</p>
<p id="p134351910131711"><a name="p134351910131711"></a><a name="p134351910131711"></a></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10921030123711"><a name="p10921030123711"></a><a name="p10921030123711"></a>静态虚拟化</p>
<p id="p333261621717"><a name="p333261621717"></a><a name="p333261621717"></a></p>
</td>
</tr>
<tr id="row89138123014"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p391138203014"><a name="p391138203014"></a><a name="p391138203014"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol4232523123116"></a><a name="ol4232523123116"></a><ol id="ol4232523123116"><li>在物理机划分vNPU的步骤请参见<a href="./04_creating_vnpu.md">创建vNPU</a>。</li><li>挂载vNPU到容器的步骤请参见<a href="./06_mounting_vnpu.md">挂载vNPU</a>。</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p671845534711"><a name="p671845534711"></a><a name="p671845534711"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row174318393462"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><div class="p" id="p879861715488"><a name="p879861715488"></a><a name="p879861715488"></a>动态虚拟化：<a name="ul1028016496477"></a><a name="ul1028016496477"></a><ul id="ul1028016496477"><li>使用<span id="ph112801498478"><a name="ph112801498478"></a><a name="ph112801498478"></a>Ascend Docker Runtime</span>挂载</li><li>使用<span id="ph828016490479"><a name="ph828016490479"></a><a name="ph828016490479"></a><span id="ph728054934716"><a name="ph728054934716"></a><a name="ph728054934716"></a>Kubernetes</span>挂载</span></li></ul>
</div>
</td>
</tr>
<tr id="row131012387307"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1010133833013"><a name="p1010133833013"></a><a name="p1010133833013"></a>在物理机划分vNPU，挂载vNPU到虚拟机，在虚拟机内将vNPU挂载到容器</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol14307634103119"></a><a name="ol14307634103119"></a><ol id="ol14307634103119"><li>在物理机划分vNPU和挂载vNPU到虚拟机的步骤请参见<span id="ph452785715619"><a name="ph452785715619"></a><a name="ph452785715619"></a>《Atlas 系列硬件产品 25.5.0 虚拟机配置指南》的“安装虚拟机&gt;配置NPU直通虚拟机&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100540506/bf80825c" target="_blank" rel="noopener noreferrer">vNPU直通虚拟机</a>”章节</span>。</li><li>在虚拟机内挂载vNPU到容器的步骤请参见<a href="./06_mounting_vnpu.md">挂载vNPU</a>。</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p13911193234713"><a name="p13911193234713"></a><a name="p13911193234713"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row3124381309"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20127385307"><a name="p20127385307"></a><a name="p20127385307"></a>在物理机直通NPU到虚拟机，在虚拟机内划分vNPU，再将vNPU挂载到虚拟机内的容器</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol441318447318"></a><a name="ol441318447318"></a><ol id="ol441318447318"><li>在物理机直通NPU到虚拟机的步骤请参见<span id="ph970622925815"><a name="ph970622925815"></a><a name="ph970622925815"></a>《Atlas 系列硬件产品 25.5.0 虚拟机配置指南》的“安装虚拟机&gt;配置NPU直通虚拟机&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100540506/2689d3e6" target="_blank" rel="noopener noreferrer">NPU直通虚拟机</a>”章节</span>。</li><li>在虚拟机内划分vNPU步骤请参见<a href="./04_creating_vnpu.md">创建vNPU</a>。</li><li>将vNPU挂载到虚拟机内的容器的步骤请参见<a href="./06_mounting_vnpu.md">挂载vNPU</a>。</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1486613195014"><a name="p1486613195014"></a><a name="p1486613195014"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row8918450194820"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><div class="p" id="p14998206105010"><a name="p14998206105010"></a><a name="p14998206105010"></a>动态虚拟化：<a name="ul55138515017"></a><a name="ul55138515017"></a><ul id="ul55138515017"><li>使用<span id="ph1051325105019"><a name="ph1051325105019"></a><a name="ph1051325105019"></a>Ascend Docker Runtime</span>挂载</li><li>使用<span id="ph1951314565011"><a name="ph1951314565011"></a><a name="ph1951314565011"></a><span id="ph1251318575010"><a name="ph1251318575010"></a><a name="ph1251318575010"></a>Kubernetes</span>挂载</span></li></ul>
</div>
</td>
</tr>
</tbody>
</table>

**vNPU挂载到容器方案<a name="section84114107544"></a>**

将vNPU挂载到容器有以下方案：

- 原生Docker：结合原生Docker使用。仅支持静态虚拟化（通过npu-smi工具创建多个vNPU），通过Docker拉起容器时将vNPU挂载到容器。

    >[!NOTE] 
    >不支持通过原生Containerd拉起容器时将vNPU挂载到容器。

- 结合MindCluster组件：
    - Ascend Docker Runtime：单独基于Ascend Docker Runtime（容器引擎插件）使用。支持静态虚拟化和动态虚拟化，通过Ascend Docker Runtime拉起容器时将vNPU挂载到容器。
    - Kubernetes：结合MindCluster组件Ascend Device Plugin、Volcano，通过Kubernetes拉起容器时将vNPU挂载到容器。支持静态虚拟化和动态虚拟化。
        - 静态虚拟化：通过npu-smi工具提前创建多个vNPU，当用户需要使用vNPU资源时，基于Ascend Device Plugin组件的设备发现、设备分配、设备健康状态上报功能，分配vNPU资源提供给上层用户使用，此方案下，集群调度组件的Volcano组件为可选。
        - 动态虚拟化：Ascend Device Plugin组件上报其所在机器的可用AICore数目。虚拟化任务上报后，Volcano经过计算将该任务调度到满足其要求的节点。该节点的Ascend Device Plugin在收到请求后自动切分出vNPU设备并挂载该任务，从而完成整个动态虚拟化过程。该过程不需要用户提前切分vNPU，在任务使用完成后又能自动回收，很好地支持用户算力需求不断变化的场景。
