# 断点续训特性指南<a name="ZH-CN_TOPIC_0000002511426955"></a>

## 特性说明<a name="ZH-CN_TOPIC_0000002511346415"></a>

### 应用场景<a name="ZH-CN_TOPIC_0000002511346439"></a>

随着神经网络和数据集的规模越来越大，单台服务器已经难以完成大规模的训练任务。为了应对这一挑战，通常需要使用多台服务器（配备更多的AI芯片）组成高密度训练集群，进行长时间的分布式训练。但随着硬件数量的增加，设备出现故障的概率也会上升，训练中断也更加频繁。因此，如何提升集群的可用性，成为当前亟需解决的重要问题。

提升集群可用性需要降低每次训练后的故障恢复成本。当前故障恢复通常需要人工排查硬件故障或者软件异常，需要大量人工成本；并且隔离故障设备后再重新拉起训练任务，需要耗费较长时间，影响整体效率。

为了解决这些问题，断点续训提供以下关键功能特性，能够在训练过程中有效应对故障，减少恢复时间，从而显著提升集群的可用性和稳定性。

**关键功能特性<a name="section15584171017252"></a>**

<a name="table1866285218270"></a>
<table><thead align="left"><tr id="row7663135222713"><th class="cellrowborder" valign="top" width="12.09120912091209%" id="mcps1.1.4.1.1"><p id="p266355252712"><a name="p266355252712"></a><a name="p266355252712"></a>功能名称</p>
</th>
<th class="cellrowborder" valign="top" width="67.54675467546754%" id="mcps1.1.4.1.2"><p id="p066313523276"><a name="p066313523276"></a><a name="p066313523276"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="20.362036203620363%" id="mcps1.1.4.1.3"><p id="p866385212720"><a name="p866385212720"></a><a name="p866385212720"></a>配置步骤</p>
</th>
</tr>
</thead>
<tbody><tr id="row1466395215279"><td class="cellrowborder" valign="top" width="12.09120912091209%" headers="mcps1.1.4.1.1 "><p id="p16631352172718"><a name="p16631352172718"></a><a name="p16631352172718"></a><strong id="b10475175393015"><a name="b10475175393015"></a><a name="b10475175393015"></a>故障检测</strong></p>
</td>
<td class="cellrowborder" valign="top" width="67.54675467546754%" headers="mcps1.1.4.1.2 "><p id="p1376115912478"><a name="p1376115912478"></a><a name="p1376115912478"></a>断点续训具有故障检测功能，支持实时监测训练场景下的20+软件类故障及90+硬件类故障的故障检测。</p>
<p id="p6269526143113"><a name="p6269526143113"></a><a name="p6269526143113"></a>详细功能及原理介绍请参见<a href="#故障检测">故障检测</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="20.362036203620363%" headers="mcps1.1.4.1.3 "><p id="p3664115213274"><a name="p3664115213274"></a><a name="p3664115213274"></a><a href="#可选配置故障检测级别">（可选）配置故障检测级别</a></p>
</td>
</tr>
<tr id="row8664195222715"><td class="cellrowborder" valign="top" width="12.09120912091209%" headers="mcps1.1.4.1.1 "><p id="p9664252172717"><a name="p9664252172717"></a><a name="p9664252172717"></a><strong id="b1847710538305"><a name="b1847710538305"></a><a name="b1847710538305"></a>故障处理</strong></p>
</td>
<td class="cellrowborder" valign="top" width="67.54675467546754%" headers="mcps1.1.4.1.2 "><p id="p1411817177434"><a name="p1411817177434"></a><a name="p1411817177434"></a>断点续训具有故障处理功能，出现故障后不需要人工介入就可自动隔离故障设备。</p>
<p id="p2333154715317"><a name="p2333154715317"></a><a name="p2333154715317"></a>详细功能及原理介绍请参见<a href="#故障处理">故障处理</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="20.362036203620363%" headers="mcps1.1.4.1.3 "><p id="p166425282717"><a name="p166425282717"></a><a name="p166425282717"></a><a href="#配置故障处理">配置故障处理</a></p>
</td>
</tr>
<tr id="row1781918488293"><td class="cellrowborder" valign="top" width="12.09120912091209%" headers="mcps1.1.4.1.1 "><p id="p381964862916"><a name="p381964862916"></a><a name="p381964862916"></a><strong id="b18478195318307"><a name="b18478195318307"></a><a name="b18478195318307"></a>训练恢复</strong></p>
</td>
<td class="cellrowborder" valign="top" width="67.54675467546754%" headers="mcps1.1.4.1.2 "><p id="p10678103713430"><a name="p10678103713430"></a><a name="p10678103713430"></a>断点续训具有训练恢复功能，用户可自定义训练恢复的策略，以最小粒度恢复训练状态，降低训练拉起时间。</p>
<p id="p1311063703219"><a name="p1311063703219"></a><a name="p1311063703219"></a>详细功能及原理介绍请参见<a href="#训练恢复">训练恢复</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="20.362036203620363%" headers="mcps1.1.4.1.3 "><p id="p17819948202912"><a name="p17819948202912"></a><a name="p17819948202912"></a><a href="#配置训练恢复">配置训练恢复</a></p>
</td>
</tr>
</tbody>
</table>

**应用场景<a name="section1498618364358"></a>**

<a name="table1716618439356"></a>
<table><thead align="left"><tr id="row51662043153514"><th class="cellrowborder" valign="top" width="12.437810945273633%" id="mcps1.1.4.1.1"><p id="p19166143203512"><a name="p19166143203512"></a><a name="p19166143203512"></a>场景分类</p>
</th>
<th class="cellrowborder" valign="top" width="48.23319300931241%" id="mcps1.1.4.1.2"><p id="p1216754353516"><a name="p1216754353516"></a><a name="p1216754353516"></a>主要业务</p>
</th>
<th class="cellrowborder" valign="top" width="39.32899604541396%" id="mcps1.1.4.1.3"><p id="p16167134363516"><a name="p16167134363516"></a><a name="p16167134363516"></a>业务价值</p>
</th>
</tr>
</thead>
<tbody><tr id="row121677433356"><td class="cellrowborder" valign="top" width="12.437810945273633%" headers="mcps1.1.4.1.1 "><p id="p316744312356"><a name="p316744312356"></a><a name="p316744312356"></a>AI训练场景</p>
<p id="p1995614753618"><a name="p1995614753618"></a><a name="p1995614753618"></a></p>
</td>
<td class="cellrowborder" valign="top" width="48.23319300931241%" headers="mcps1.1.4.1.2 "><p id="p2016754317352"><a name="p2016754317352"></a><a name="p2016754317352"></a>支持对计算、网络和存储设备资源的监测，AI环境的健康检查和AI作业故障诊断。</p>
</td>
<td class="cellrowborder" valign="top" width="39.32899604541396%" headers="mcps1.1.4.1.3 "><a name="ul71678434353"></a><a name="ul71678434353"></a><ul id="ul71678434353"><li>整体监测集群环境资源。</li><li>提升AI训练业务的作业成功率。</li><li>减少AI作业训练故障的处理及恢复时间。</li></ul>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>-   较小规模的模型任务训练用时较短（时长 < 1h），硬件出现故障的频率较低，不推荐用户使用断点续训特性。
>-   本特性不适用于算力虚拟化场景。


### 整体架构<a name="ZH-CN_TOPIC_0000002479226568"></a>

在K8s（Kubernetes）集群中训练任务出现故障时，断点续训特性使系统能够感知故障，将故障资源进行处理或隔离，并根据训练任务需要重新分配资源，通过周期性保存或临终保存的CKPT（CheckPoint）重新拉起训练任务，缩短损失时间。

**断点续训整体架构<a name="section1483121685613"></a>**

断点续训架构原理如[图1](#fig1285977919)所示。

**图 1**  整体架构<a name="fig1285977919"></a>  
![](../../figures/scheduling/整体架构.png "整体架构")

其中各个部分的能力如下：

1.  Ascend Device Plugin：故障发现组件，提供NPU资源管理、NPU芯片故障和NPU网络故障上报、执行芯片热复位等能力。
2.  NodeD：故障发现组件，提供节点健康状态、节点硬件（包括CPU、内存、芯片等部件）故障、灵衢网络故障和DPC共享存储故障上报能力。
3.  Volcano：故障处理组件，提供故障任务重调度的能力。
4.  Ascend Operator：为不同AI框架的分布式训练任务生成相应的环境变量；提供静态组网集合通信所需的RankTable信息。
5.  ClusterD：获取集群中所有Ascend Device Plugin和NodeD上报的数据，整理后发送给Volcano。
6.  TaskD：提供与K8s集群的训练集群控制中心的通信功能，完成恢复训练；提供昇腾设备上训练及推理任务的训练状态监测和训练状态控制能力。
7.  MindIO TTP：在大模型训练过程中发生故障后，校验中间状态数据的完整性和一致性，生成一次临终CKPT数据，恢复训练时能够通过该CKPT数据恢复，减少故障造成的训练迭代损失。
8.  训练模型代码：需要进行断点续训相关能力的适配操作。

**端到端流程<a name="section11999131581210"></a>**

断点续训特性基于故障触发，触发成功后经过故障检测、故障处理和训练恢复三个阶段后可恢复训练。

**图 2**  端到端流程<a name="fig1968122593112"></a>  
![](../../figures/scheduling/端到端流程.png "端到端流程")

各步骤说明如下：

1.  通过轮询的方式查询设备状态，Ascend Device Plugin从DCMI接口获取NPU状态以及NodeD上报的节点健康状态、节点硬件故障信息和灵衢网络故障信息，ClusterD整理所有的故障信息，确定最终故障状态后，上报给Volcano。
2.  查询到节点或芯片故障后，对故障节点或芯片进行隔离，防止再次调度到该设备上。
3.  停止训练进程，退出训练容器。
4.  节点或芯片故障后，系统会将训练任务重调度到健康的设备上，重启训练容器；该训练任务被重调度选择资源时，优先选用未导致本次训练任务重调度的节点。
5.  训练脚本重新拉起训练进程。
6.  运维人员可以根据节点或芯片的故障类型判断是否可进行热复位。
7.  进行故障热复位，使设备恢复健康状态。
8.  恢复后的设备自动重新加入集群中。
9.  不可恢复的设备通过运维监测系统上报告警。
10. 对不可恢复的设备进行线下人工维修和换件。

>[!NOTE] 说明 
>业务面故障触发的断点续训功能，将只执行上述步骤3\~步骤5。

**组件调用流程<a name="section1726018439396"></a>**

断点续训组件调用流程如[图3](#fig1710473818543)所示。

**图 3**  架构原理<a name="fig1710473818543"></a>  
![](../../figures/scheduling/架构原理.png "架构原理")

各步骤说明如下：

1.  Ascend Device Plugin发现和上报故障及健康状态。
2.  NodeD更新节点硬件故障信息，以便Volcano可以准确判断节点故障类型。
3.  ClusterD根据Ascend Device Plugin提供的芯片信息，判断芯片是否健康。
4.  ClusterD获取NodeD上报的故障信息。
5.  ClusterD将收集来的芯片及节点信息汇总后，放入ConfigMap。
6.  Volcano获取整个集群的设备信息，若任务使用的设备上存在故障信息，Volcano会将任务调度到其他健康设备上。
7.  Volcano按照亲和性规则选择节点和芯片，并由Ascend Operator创建新的Pod后，再调度训练任务到符合要求的节点上。
8.  Ascend Device Plugin根据Pod上Volcano指定的芯片ID来分配芯片，并将芯片IP信息写入容器。
9.  容器启动之前，Ascend Docker Runtime为训练容器自动挂载NPU相关设备，驱动so等文件和目录。
10. Ascend Operator将训练任务需要的相关环境变量（如集合通信信息和训练配置信息等）写入容器中。并且获取训练任务容器上的芯片信息，自动生成分布式训练任务需要的集合通信信息。

**使用条件<a name="section179221342141518"></a>**

-   使用断点续训功能需要安装的组件详见[所需组件](../introduction.md#断点续训)章节。
-   断点续训特性是基于MindCluster集群调度组件的高阶特性，使用断点续训特性前需要完成的准备工作详见[准备K8s和共享存储](#准备k8s和共享存储)章节。


### 性能说明<a name="ZH-CN_TOPIC_0000002479386472"></a>

断点续训特性可以在训练发生故障后恢复训练，降低故障导致的训练损失。断点续训的故障整体恢复时间可以分为训练回滚时间和训练拉起时间，如[图1](#zh-cn_topic_0000002003001306_fig13371418134510)所示。

**图 1**  故障恢复阶段<a name="zh-cn_topic_0000002003001306_fig13371418134510"></a>  
![](../../figures/scheduling/故障恢复阶段.png "故障恢复阶段")

**训练回滚时间**

训练出现故障后会丢失原有的训练数据，需要从保存的CKPT文件中恢复训练。在大模型训练中，由于每次保存CKPT会降低训练效率，因此通常1小时以上才会保存一次CKPT文件，每次故障后将会丢失上次保存CKPT时间点到当前故障时间点的训练数据。训练回滚时间即使用上次保存的CKPT文件训练到出现故障点的时间。设平均训练回滚时间为T<sub>0</sub>，CKPT保存周期为G<sub>f</sub>，则故障平均训练回滚时间T<sub>0</sub>=G<sub>f</sub>/2。

**训练拉起时间**

训练出现故障后，需要重新拉起训练任务，恢复训练容器及训练进程，完成资源重调度、集合通信初始化、CKPT加载和编译等流程后继续往后训练。训练故障后需要完整走完一段训练拉起时间后才能继续训练，训练拉起时间过长会导致资源浪费。设资源重调度时间为T<sub>1</sub>，集合通信时间为T<sub>2</sub>，CKPT加载时间为T<sub>3</sub>，编译时间为T<sub>4</sub>，因此训练拉起时间为T<sub>1</sub>+T<sub>2</sub>+T<sub>3</sub>+T<sub>4</sub>。

单次故障总训练损失时间T=T<sub>0</sub>+T<sub>1</sub>+T<sub>2</sub>+T<sub>3</sub>+T<sub>4</sub>。具体的时间参考请参见[训练恢复耗时参考](#zh-cn_topic_0000002003001306_section1672017599123)。

>[!NOTE] 说明 
>其中每部分时间与参数规模和集群规模相关，网络与存储性能也会影响总训练损失时间。

**训练恢复耗时参考<a name="zh-cn_topic_0000002003001306_section1672017599123"></a>**

以PyTorch框架下的GPT-3模型，其在NFS存储下写入速度为2.7GB/s，读取速度为4.8GB/s的情况下，参数量大小为3B或15B的单机8卡任务为例。**（故障处理模式为重调度，若使用优雅容错模式，可不参考该指标。）**

-   参数量大小为3B，如[图2](#zh-cn_topic_0000002003001306_fig175521679432)所示，该模型的CKPT落盘时间约为**30**秒，断点续训在设备发现阶段用时小于**5**秒，设备处理阶段用时小于**30**秒，训练重启阶段用时大约在**70**秒左右，训练重启阶段的CKPT加载功能用时约**3**秒。
-   参数量大小为15B，如[图3](#zh-cn_topic_0000002003001306_fig10995142020518)所示，该模型的CKPT落盘时间约为**120**秒，断点续训在设备发现阶段用时小于**5**秒，设备处理阶段用时小于**30**秒，训练重启阶段用时大约在**210**秒左右，训练重启阶段的CKPT加载功能用时约**90**秒。

**图 2**  3B模型时间指标<a name="zh-cn_topic_0000002003001306_fig175521679432"></a>  
![](../../figures/scheduling/3B模型时间指标.png "3B模型时间指标")

**图 3**  15B模型时间指标<a name="zh-cn_topic_0000002003001306_fig10995142020518"></a>  
![](../../figures/scheduling/15B模型时间指标.png "15B模型时间指标")



## 方案和原理<a name="ZH-CN_TOPIC_0000002511346509"></a>

### 故障检测<a name="ZH-CN_TOPIC_0000002479226514"></a>

#### 故障说明<a name="ZH-CN_TOPIC_0000002511426413"></a>

断点续训基于故障检测能力获取集群和训练业务的故障状态，根据检测结果进行故障处理。当前，断点续训特性主要提供以下几个方面的故障检测能力：昇腾硬件故障、训练业务故障、其他故障发送方的故障。

MindCluster集群调度组件Ascend Device Plugin提供NPU芯片故障检测能力及NPU参数面网络故障检测能力，NodeD提供服务器节点故障、DPC共享存储故障和灵衢网络故障检测能力，ClusterD提供公共故障检测能力，Volcano提供业务面容器异常检测能力，故障检测整体架构如下图所示。

![](../../figures/scheduling/250411110432760.png)

1.  计算服务器上的Ascend Device Plugin通过驱动获取NPU芯片故障以及参数面网络故障后，将故障信息上报到管理服务器。
2.  计算服务器上的NodeD通过驱动获取服务器节点故障、DPC共享存储故障和灵衢网络故障信息后，将故障信息上报到管理服务器。
3.  计算服务器上的K8s监测训练容器状态，训练容器异常后上报到K8s中，管理服务器上的Volcano通过K8s获取训练容器的故障信息。
4.  管理服务器上的ClusterD通过公共故障接口获取公共故障后，将接收到的信息进行汇总写入cluster-info-device-cm。
5.  （可选）管理服务器上的ClusterD汇总集群内所有Ascend Device Plugin和NodeD上报的故障信息。

**支持的故障模式<a name="zh-cn_topic_0000002039699773_section8301627182117"></a>**

当前已支持200+故障的检测。支持的故障类型请参见[表1](#zh-cn_topic_0000002039699773_table9980135316395)，详细的故障说明请参见[典型故障.xlsx](../resource/典型故障.xlsx)。

**表 1**  故障类型说明

<a name="zh-cn_topic_0000002039699773_table9980135316395"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039699773_row1980185311394"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002039699773_p7980853183910"><a name="zh-cn_topic_0000002039699773_p7980853183910"></a><a name="zh-cn_topic_0000002039699773_p7980853183910"></a>故障类型</p>
</th>
<th class="cellrowborder" valign="top" width="80%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002039699773_p99801953193918"><a name="zh-cn_topic_0000002039699773_p99801953193918"></a><a name="zh-cn_topic_0000002039699773_p99801953193918"></a>故障说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039699773_row139804539392"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p1198014539399"><a name="zh-cn_topic_0000002039699773_p1198014539399"></a><a name="zh-cn_topic_0000002039699773_p1198014539399"></a>节点故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002039699773_p2960204381116"><a name="zh-cn_topic_0000002039699773_p2960204381116"></a><a name="zh-cn_topic_0000002039699773_p2960204381116"></a>包括节点健康状态、节点硬件故障和DPC共享存储故障。</p>
<p id="p3883113092117"><a name="p3883113092117"></a><a name="p3883113092117"></a>故障码说明请参见<a href="../appendix.md#节点故障码参考文档">节点故障码参考文档</a>。</p>
<div class="note" id="zh-cn_topic_0000002039699773_note113491120161019"><a name="zh-cn_topic_0000002039699773_note113491120161019"></a><a name="zh-cn_topic_0000002039699773_note113491120161019"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039699773_p1634916206103"><a name="zh-cn_topic_0000002039699773_p1634916206103"></a><a name="zh-cn_topic_0000002039699773_p1634916206103"></a>若节点的硬件故障导致节点宕机或重启，则<span id="zh-cn_topic_0000002039699773_ph469218596150"><a name="zh-cn_topic_0000002039699773_ph469218596150"></a><a name="zh-cn_topic_0000002039699773_ph469218596150"></a>NodeD</span>无法检测到具体的故障类型并上报。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039699773_row16980753123914"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p2980553133917"><a name="zh-cn_topic_0000002039699773_p2980553133917"></a><a name="zh-cn_topic_0000002039699773_p2980553133917"></a>芯片故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p68031627162116"><a name="p68031627162116"></a><a name="p68031627162116"></a>DCMI接口上报的芯片故障和设备网络探测工具hccn_tool检测到的芯片网络故障。</p>
<p id="zh-cn_topic_0000002039699773_p13762144301315"><a name="zh-cn_topic_0000002039699773_p13762144301315"></a><a name="zh-cn_topic_0000002039699773_p13762144301315"></a>故障码说明请参见<a href="../appendix.md#芯片故障码参考文档">芯片故障码参考文档</a>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039699773_row9980165319394"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p11981453103911"><a name="zh-cn_topic_0000002039699773_p11981453103911"></a><a name="zh-cn_topic_0000002039699773_p11981453103911"></a>参数面网络故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><div class="p" id="zh-cn_topic_0000002039699773_p1357910482138"><a name="zh-cn_topic_0000002039699773_p1357910482138"></a><a name="zh-cn_topic_0000002039699773_p1357910482138"></a>包括芯片网络相关故障和灵衢总线设备故障。<a name="zh-cn_topic_0000002039699773_ul823018522254"></a><a name="zh-cn_topic_0000002039699773_ul823018522254"></a><ul id="zh-cn_topic_0000002039699773_ul823018522254"><li>芯片网络相关故障：芯片之间进行参数交换的专用网络出现故障，如NPU网口故障。</li><li>灵衢总线设备故障：<span id="zh-cn_topic_0000002039699773_ph1893095102412"><a name="zh-cn_topic_0000002039699773_ph1893095102412"></a><a name="zh-cn_topic_0000002039699773_ph1893095102412"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>的灵衢总线设备发生故障。</li></ul>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039699773_row11506155782216"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p6383230124815"><a name="zh-cn_topic_0000002039699773_p6383230124815"></a><a name="zh-cn_topic_0000002039699773_p6383230124815"></a>业务面故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002039699773_p538343064810"><a name="zh-cn_topic_0000002039699773_p538343064810"></a><a name="zh-cn_topic_0000002039699773_p538343064810"></a>训练任务异常退出，导致Pod的Status变为Failed状态。</p>
<div class="note" id="zh-cn_topic_0000002039699773_note113831617138"><a name="zh-cn_topic_0000002039699773_note113831617138"></a><a name="zh-cn_topic_0000002039699773_note113831617138"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><div class="p" id="zh-cn_topic_0000002039699773_p1938313171939"><a name="zh-cn_topic_0000002039699773_p1938313171939"></a><a name="zh-cn_topic_0000002039699773_p1938313171939"></a>可执行<strong id="zh-cn_topic_0000002039699773_b101591617517"><a name="zh-cn_topic_0000002039699773_b101591617517"></a><a name="zh-cn_topic_0000002039699773_b101591617517"></a>kubectl describe pod <em id="zh-cn_topic_0000002039699773_i135141149345"><a name="zh-cn_topic_0000002039699773_i135141149345"></a><a name="zh-cn_topic_0000002039699773_i135141149345"></a>{pod名称} </em>-n<em id="zh-cn_topic_0000002039699773_i137271056545"><a name="zh-cn_topic_0000002039699773_i137271056545"></a><a name="zh-cn_topic_0000002039699773_i137271056545"></a> {NAMESPACE}</em> |grep Status:</strong>命令，查看当前Pod的Status是否为Failed状态。回显示例如下：<pre class="screen" id="zh-cn_topic_0000002039699773_screen926115341054"><a name="zh-cn_topic_0000002039699773_screen926115341054"></a><a name="zh-cn_topic_0000002039699773_screen926115341054"></a><strong id="zh-cn_topic_0000002039699773_b11562201310616"><a name="zh-cn_topic_0000002039699773_b11562201310616"></a><a name="zh-cn_topic_0000002039699773_b11562201310616"></a>Status:       Failed</strong></pre>
</div>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039699773_row4734144233211"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p1173419427324"><a name="zh-cn_topic_0000002039699773_p1173419427324"></a><a name="zh-cn_topic_0000002039699773_p1173419427324"></a>公共故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002039699773_p75681253134114"><a name="zh-cn_topic_0000002039699773_p75681253134114"></a><a name="zh-cn_topic_0000002039699773_p75681253134114"></a>公共故障指的是其他故障发现者（非<span id="zh-cn_topic_0000002039699773_ph14169931153018"><a name="zh-cn_topic_0000002039699773_ph14169931153018"></a><a name="zh-cn_topic_0000002039699773_ph14169931153018"></a>MindCluster</span>组件）提供的故障，公共故障包括以下几种类型：NPU故障、节点故障、网络故障和存储故障。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039699773_row165491043113014"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p2277188103115"><a name="zh-cn_topic_0000002039699773_p2277188103115"></a><a name="zh-cn_topic_0000002039699773_p2277188103115"></a>pingmesh灵衢网络故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002039699773_p8830914203219"><a name="zh-cn_topic_0000002039699773_p8830914203219"></a><a name="zh-cn_topic_0000002039699773_p8830914203219"></a>灵衢网络故障是针对超节点内部（包括节点内和节点间）的<span id="ph17233131243911"><a name="ph17233131243911"></a><a name="ph17233131243911"></a>HCCS</span>网络提供的NPU网络故障检测。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039699773_row1267815329323"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002039699773_p166788326323"><a name="zh-cn_topic_0000002039699773_p166788326323"></a><a name="zh-cn_topic_0000002039699773_p166788326323"></a>性能劣化故障</p>
</td>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002039699773_p46391813195611"><a name="zh-cn_topic_0000002039699773_p46391813195611"></a><a name="zh-cn_topic_0000002039699773_p46391813195611"></a><span id="zh-cn_topic_0000002039699773_ph1669181417563"><a name="zh-cn_topic_0000002039699773_ph1669181417563"></a><a name="zh-cn_topic_0000002039699773_ph1669181417563"></a>MindCluster</span>结合MindStudio提供的profiling能力对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态使能打点和打点数据持久化功能、可动态启停，无需重启任务进行诊断，对训练无损耗。</p>
</td>
</tr>
</tbody>
</table>

**ConfigMap说明<a name="zh-cn_topic_0000002039699773_section49901206282"></a>**

-   每个计算节点的Ascend Device Plugin均会创建记录本节点NPU和灵衢总线设备信息的ConfigMap文件。该ConfigMap文件名为mindx-dl-deviceinfo-_<nodename\>_（以下简称device-info-cm），故障信息会通过该ConfigMap进行上报。该ConfigMap文件中各字段的说明，请参见[DeviceInfoCfg](../api/ascend_device_plugin.md#芯片资源)表。
-   当节点上存在节点故障时，每个计算节点的NodeD会创建记录本节点设备信息的ConfigMap文件。该ConfigMap文件名为mindx-dl-nodeinfo-_<nodename\>_（以下简称node-info-cm），节点故障信息会通过该ConfigMap进行上报。该ConfigMap文件中各字段的说明，请参见[mindx-dl-nodeinfo-<nodename\>](../api/noded.md#节点资源)表。
-   ClusterD会创建记录本集群设备信息的ConfigMap文件，该ConfigMap文件名为cluster-info-<device/switch\>-<\[0-5\]\>、cluster-info-node-cm（以下简称cluster-info-cm）。节点及芯片故障信息会通过[cluster-info-cm](../api/clusterd.md#集群资源)进行上报。
-   创建每个任务时，需要在YAML中配置ConfigMap文件，该ConfigMap文件名称为reset-config-_<job-name\>_（以下简称reset-info-cm）。该ConfigMap挂载到容器的“/user/restore/reset/config“路径下。Ascend Device Plugin会自动将ConfigMap挂载到本节点的“/user/restore/reset/<job-namespace\>.<job-name\>”路径下。

    也可以将节点上/user/restore/reset/<job-namespace\>.<job-name\>替代ConfigMap，挂载到容器的“/user/restore/reset/config”路径下。该ConfigMap文件字段说明，请参见[reset-config-<job-name\>](../api/ascend_device_plugin.md#任务信息)表。


#### 节点故障<a name="ZH-CN_TOPIC_0000002479386528"></a>

节点故障的发现主要通过NodeD组件实现。节点故障包括节点健康状态和节点硬件故障、节点DPC共享存储故障，详细说明如下：

-   节点健康状态

    NodeD完成当前节点的节点状态诊断后，收集本节点内的故障信息。当节点发生故障时，通过节点状态上报机制不断向Volcano发送节点状态（当前仅收集本节点内的硬件故障信息）。

-   节点硬件故障

    针对节点硬件故障，NodeD通过IPMI驱动向iBMC发送故障查询请求，iBMC将当前硬件告警信息响应给NodeD。NodeD收集硬件告警信息后，将节点硬件状态上报给Volcano。

-   节点DPC共享存储故障

    针对使用Scale-Out Storage DPC产品的节点，可以使用NodeD安装包下的noded-dpc-\{version\}.yaml启动NodeD服务。开启对DPC的进程异常及内存不足异常的检测和上报。

    >[!NOTE] 说明 
    >当节点发生故障时，NodeD会上报节点健康状态和节点硬件故障。无故障时，默认节点健康。

**图 1**  节点故障上报<a name="fig1329112151382"></a>  
![](../../figures/scheduling/节点故障上报.png "节点故障上报")

-   当节点发生故障时，NodeD最短5秒（默认）更新本节点的node-info-cm内容，其中字段说明见[mindx-dl-nodeinfo-<nodename\>](../api/noded.md#节点资源)表。
-   NodeD每隔60秒（默认），当从iBMC查询到故障信息或与上次上报的时间间隔30分钟以上时，会在1秒内上报到node-info-cm中。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证节点故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator、NodeD、ClusterD

**使用约束<a name="section16867482102"></a>**

-   NodeD的节点硬件故障上报能力仅支持以下产品型号：Atlas 800T A2 训练服务器、Atlas 900 A2 PoD 集群基础单元、Atlas 900 A3 SuperPoD 超节点。
-   仅V2 3.15.0.1及以上版本或者V2 3.10.02.55版本的iBMC，且安装了IPMC驱动的产品，支持NodeD的节点硬件故障上报能力。低版本的iBMC或IPMI获取节点故障信息失败时，将只上报节点健康状态。
-   如需使用超节点故障检测功能，需使用V3 5.8.3.35及以上版本的iBMC。
-   如需使用DPC故障检测功能，需使用Scale-Out Storage DPC 24.2.0及以上版本。

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度

**（可选）配置故障检测的级别<a name="section1343172016386"></a>**

断点续训针对节点故障中**节点硬件故障**的不同故障码，提供了默认的故障级别和对应级别的故障处理策略。若用户需要修改故障处理策略，可参见[节点硬件故障](#节点硬件故障)。若无特殊需求，请勿随意修改。


#### 芯片故障<a name="ZH-CN_TOPIC_0000002511346395"></a>

芯片故障指的是NPU出现的基础软件类故障和芯片硬件类故障。断点续训特性中芯片故障的检测和上报由设备管理组件Ascend Device Plugin负责。

**NPU上报机制<a name="section15950121613265"></a>**

NPU发生故障时，故障管理框架获取到故障信息后，将该信息上传给NPU驱动的故障管理框架。故障管理框架收到故障信息后，通过DCMI接口上报给Ascend Device Plugin，如[图1](#fig3951191610267)所示。

Ascend Device Plugin通过DCMI接口获取芯片健康状态。当前提供如下两种获取模式：

-   故障订阅模式。Ascend Device Plugin启动时会先调用DCMI故障订阅接口注册监测，故障发生时驱动通过该接口将故障事件上报给Ascend Device Plugin。故障恢复时通过该接口将恢复事件上报给Ascend Device Plugin。
-   故障轮询模式。每隔固定时间，通过故障查询接口查询芯片故障状态，当设备驱动不支持订阅能力时将切换该模式。

**图 1**  芯片故障上报<a name="fig3951191610267"></a>  
![](../../figures/scheduling/芯片故障上报.png "芯片故障上报")

**Ascend Device Plugin上报机制<a name="section0951116132615"></a>**

Ascend Device Plugin获取到芯片故障信息后，通过ConfigMap的形式上报给K8s。Ascend Device Plugin的故障上报机制如下：

**图 2**  上报故障到K8s<a name="fig10951101692610"></a>  
![](../../figures/scheduling/上报故障到K8s.png "上报故障到K8s")

对于不同故障处理模式，上报的路径会有一定差别。

-   重调度模式：Ascend Device Plugin获取到芯片故障后，将芯片故障信息写入该节点所属的device-info-cm中，其中字段说明见[DeviceInfoCfg](../api/ascend_device_plugin.md#芯片资源)表。ClusterD读取每个节点的device-info-cm感知芯片故障并上报给调度器。
-   优雅容错模式：Ascend Device Plugin获取到可恢复的芯片故障后，将芯片故障信息写入该任务所属的reset-info-cm中，业务容器通过将reset-info-cm挂载为文件的形式，读取文件感知芯片故障。

    >[!NOTE] 说明 
    >若优雅容错模式处理故障失败，回退至重调度模式后，故障上报的路径则按照重调度模式进行上报。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证芯片故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator、Ascend Device Plugin、ClusterD

**（可选）配置故障检测的级别<a name="section1343172016386"></a>**

断点续训针对**芯片故障**提供了默认的故障频率、时长、故障级别以及对应级别的故障处理策略。若用户需要修改故障处理策略，可参见[芯片故障](#芯片故障-1)。若无特殊需求，请勿随意修改。

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度、进程级在线恢复、优雅容错

>[!NOTE] 说明 
>仅片上内存出现的不可纠正错误支持进程级在线恢复，其他类型的芯片故障不支持进程级在线恢复。


#### 参数面网络故障<a name="ZH-CN_TOPIC_0000002511426381"></a>

NPU的参数面网络故障包括芯片网络相关故障和灵衢总线设备故障。

参数面网络出现故障时，将导致训练任务中断或者训练任务性能较差。灵衢总线设备发生故障后，MindCluster集群调度组件将根据故障级别进行相应的重调度处理。

>[!NOTE] 说明 
>-   参数面网络故障不会直接触发任务重调度，当参数面故障导致训练任务异常中断时才触发任务重调度。
>-   如果需要对参数面网络故障进行故障处理，需要同时开启业务面故障无条件重试能力。开启业务面故障无条件重试需要在任务YAML中同时配置以下3个参数：fault-retry-times，restartPolicy及policies。关于参数的详细说明请参见[YAML参数说明](#yaml参数说明)。

参数面网络故障检测由设备管理组件Ascend Device Plugin负责，详细原理如[图1](#fig68743107307)所示。

**图 1**  故障检测<a name="fig68743107307"></a>  
![](../../figures/scheduling/故障检测.png "故障检测")

**关键步骤说明<a name="section1787471017308"></a>**

**芯片网络故障**：

1.  NPU定时检测和网关地址的通信是否正常，探测周期为2.5秒，通过故障管理框架上报结果。
2.  RoCE驱动实时监测NPU网口Link状态，通过故障管理框架上报Linkdown或Linkup事件。
3.  Ascend Device Plugin通过DCMI接口从故障管理框架获取信息，通过轮询的方式查询网关探测结果，并实时订阅网口Linkdown或Linkup事件并进行上报。Ascend Device Plugin统计网关检测异常持续时间、Linkdown持续时间。如果小于或等于RoCE网络超时时间（默认为20秒）则标记为NPU网络故障（默认不处理，可能会引起参数面网络故障）；如果大于20秒，则升级成配置的故障等级。

**灵衢总线设备故障**：

1.  灵衢总线设备将设备发生的故障写入本地队列中。
2.  灵衢查询接口通过查询上述队列，将故障缓存至查询接口，并进行汇总处理。
3.  Ascend Device Plugin通过订阅或轮询的方式调用接口获取灵衢总线设备相关故障，并写入device-info-cm进行上报。

**故障上报机制<a name="section1874141093019"></a>**

-   **芯片发生网络故障时**，NPU故障管理框架获取故障信息后，将该信息上报给NPU驱动。NPU驱动收到故障信息后，通过DCMI接口上报给Ascend Device Plugin。Ascend Device Plugin通过DCMI接口获取芯片健康状态。当前提供如下两种获取模式：
    -   故障订阅模式。Ascend Device Plugin启动时会先调用DCMI故障订阅接口注册监测，故障发生或恢复时，驱动通过该接口将故障发生或恢复事件上报给Ascend Device Plugin。
    -   故障轮询模式。每隔固定时间，通过故障查询接口查询芯片故障状态。当设备驱动不支持订阅能力时将切换该模式。

-   **灵衢总线设备发生故障时**，Ascend Device Plugin通过灵衢查询接口获取故障信息，当前故障查询提供两种模式：
    -   故障订阅模式：在Ascend Device Plugin启动过程中向灵衢查询接口注册故障处理回调。故障发生后，该回调被调用后将故障上报给Ascend Device Plugin，故障恢复时通过该接口上报恢复事件。
    -   故障轮询模式：Ascend Device Plugin每隔5分钟调用一次全量故障查询接口。

**Ascend Device Plugin上报机制<a name="section1875111093017"></a>**

Ascend Device Plugin获取到参数面网络故障后，将故障信息写入到device-info-cm中，并通过ConfigMap的形式上报给K8s。device-info-cm中各字段的说明，请参见[DeviceInfoCfg](../api/ascend_device_plugin.md#芯片资源)表。

Ascend Device Plugin的故障上报机制如[图2](#fig1587571063011)所示。

**图 2**  故障上报<a name="fig1587571063011"></a>  
![](../../figures/scheduling/故障上报.png "故障上报")

**watchdog故障检测<a name="section4599926103917"></a>**

参数面网络链路异常（参数面网络故障）可能导致任务中正常NPU无法与故障NPU通信，使所有NPU集合通信陷入超时等待状态；并使任务集合通信出现等待超时异常后才退出（默认为30分钟）。

开启watchdog功能（且开启了业务面故障无条件重试能力）可以在参数面网络链路异常发生后，隔离故障NPU，将任务重调度到健康的NPU上，从而实现6分钟内使任务快速退出。

>[!NOTE] 说明 
>仅支持在PyTorch及MindSpore框架下使用watchdog功能。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证参数面网络故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator、Ascend Device Plugin、ClusterD

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度

**（可选）配置故障检测的级别<a name="section1343172016386"></a>**

断点续训针对**参数面故障**提供了默认的故障级别以及对应级别的故障处理策略，若用户需要修改故障处理策略可参见[参数面网络故障](#参数面网络故障-1)。若无特殊需求，请勿随意修改。


#### 业务面故障<a name="ZH-CN_TOPIC_0000002479386512"></a>

断点续训特性支持通过Volcano调度器感知并处理因业务面故障导致的任务失败。业务面故障是因容器内的训练进程均异常退出后引起容器异常退出，导致Pod的Status变为Failed状态。在使用Ascend Operator的场景下，业务面故障仅支持任务的部分Pod发生故障的场景，若任务所有Pod在几秒内Status都转变为Failed，任务不会发生重调度，认定任务为失败状态。

业务面故障发现原理如[图1](#fig1761563615337)所示。

**图 1**  发现原理<a name="fig1761563615337"></a>  
![](../../figures/scheduling/发现原理.png "发现原理")

调度器不断轮询地查询每个任务的Pod状态，从而感知到业务面故障并上报该故障。用户可根据具体业务需求对业务面故障做处理。断点续训获取到业务面故障后，Volcano会检测是否开启无条件重试功能，开启后会将任务重新调度到未导致本次训练任务重调度的新节点，并重新执行训练任务，重试次数减1；当重试次数为0或者没有开启无条件重试功能时，不会对业务容器故障进行处理。

>[!NOTE] 说明 
>-   如需使用无条件重试功能，需在任务YAML中配置以下3个参数：fault-retry-times，restartPolicy及policies，详细参数说明请参见[YAML参数说明](#yaml参数说明)。
>-   在使用Ascend Operator的场景下，若希望任务所有Pod的Status在转变为Failed后仍发生重调度，可参考[使用Volcano和Ascend Operator组件场景下，业务面故障的任务所有Pod的Status全部变为Failed，任务无法触发无条件重试重调度](../faq.md#使用volcano和ascend-operator组件场景下业务面故障的任务所有pod的status全部变为failed任务无法触发无条件重试重调度)。

**watchdog故障检测<a name="section59641929143117"></a>**

NPU上Task执行异常（业务面故障）可能导致任务中正常NPU无法与故障NPU通信，使正常NPU集合通信陷入超时等待状态，任务集合通信出现等待超时异常后才退出（默认为30分钟）。开启watchdog功能（需同时开启业务面故障无条件重试能力），可以在该异常发生后，隔离故障NPU，将任务重调度到健康的NPU上，从而实现6分钟内使任务快速退出。

>[!NOTE] 说明 
>NPU上Task执行异常仅支持Atlas A2 训练系列产品的PyTorch框架使用watchdog功能。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证业务面故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度、优雅容错


#### 公共故障<a name="ZH-CN_TOPIC_0000002511426387"></a>

公共故障指的是其他故障发送方（非MindCluster组件）上报的故障，公共故障包括以下几种类型：NPU故障、节点故障、网络故障和存储故障。

>[!NOTE] 说明 
>ClusterD支持接收公共故障的前提是需要在节点上安装Ascend Device Plugin，并且生成了相应的device-info-cm。

**上报机制<a name="zh-cn_topic_0000002216292813_section64469192378"></a>**

公共故障发送方发现故障后，将通过ConfigMap或gRPC方式，将获取到的故障信息发送给ClusterD。ClusterD会将接收到的信息进行汇总写入cluster-info-device-cm，再上报给Ascend-volcano-plugin。

-   通过ConfigMap获取。故障发现者将故障信息写入ConfigMap中，然后由ClusterD获取故障信息。用户可通过调用ConfigMap接口的方式来注入公共故障，详细说明请参见[ConfigMap](../api/clusterd.md#configmap)。
-   通过gRPC获取。故障发现者将故障信息通过gRPC通道发送给ClusterD，然后由ClusterD获取故障信息。用户可通过调用gRPC接口的方式来注入公共故障，说明请参见[公共故障接口](../api/clusterd.md#公共故障接口)中"gRPC接口"章节。

**图 1**  公共故障上报<a name="fig72618571585"></a>  
![](../../figures/scheduling/公共故障上报.png "公共故障上报")

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证公共故障检测功能的正常使用，需要安装以下组件。

-   必选组件：Volcano、Ascend Operator、Ascend Device Plugin、ClusterD
-   可选组件：NodeD

**支持的故障处理类型<a name="zh-cn_topic_0000002216292813_section177211923175116"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度

**（可选）配置故障检测的级别和发送方<a name="zh-cn_topic_0000002216292813_section1343172016386"></a>**

断点续训针对**公共故障**提供了默认的故障级别以及支持的故障发送方。若用户需要修改公共故障的级别及故障发送方，可参见[公共故障](#可选配置公共故障的级别和发送方)。若无特殊需求，请勿随意修改。


#### pingmesh灵衢网络故障<a name="ZH-CN_TOPIC_0000002511426437"></a>

灵衢网络故障是针对超节点内部（包括节点内和节点间）的HCCS网络提供的NPU网络故障检测。

**上报机制<a name="zh-cn_topic_0000002193288232_section68367256347"></a>**

NodeD调用DCMI接口启动pingmesh任务，并周期性查询pingmesh结果，将该结果写入文件<nodename\>.log。该文件所在目录在容器中为固定路径：/user/mind-cluster/pingmesh，物理机默认目录/user/mind-cluster/pingmesh。物理机路径可以修改，修改方式如以下说明所示。

>[!NOTE] 说明 
>-   <nodename\>非固定值，为K8s中查询到的节点名称。
>-   <nodename\>.log文件物理机路径可由用户根据实际情况自行配置：在NodeD的启动YAML中修改挂载卷名称为pingmesh-result的物理机挂载路径。

获取pingmesh结果后，ClusterD会对结果进行初步分析，将故障信息写入到名为[pingmesh-fault-<nodename\>](#zh-cn_topic_0000002193288232_table2371535113510)的ConfigMap文件中。ClusterD会侦听该ConfigMap信息，并将故障汇总后上报给Volcano，由Volcano进行调度。

**前提条件<a name="zh-cn_topic_0000002193288232_section8281518121516"></a>**

-   （必选）已[创建命名空间](../installation_guide.md#创建命名空间)
-   在相应节点上完成以下组件的安装：[NodeD](../installation_guide.md#noded)（必选）、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)（可选）、[ClusterD](../installation_guide.md#clusterd)（可选）
-   （必选）已[配置NodeD启动参数resultMaxAge](../installation_guide.md#noded)

**使用约束<a name="zh-cn_topic_0000002193288232_section156679598384"></a>**

本功能仅支持在以下产品型号中使用：Atlas 900 A3 SuperPoD 超节点。

**配置灵衢网络检测<a name="zh-cn_topic_0000002193288232_section18190175418362"></a>**

配置灵衢网络检测，需执行以下步骤。

1.  配置共享存储。

    ClusterD和NodeD通过共享存储进行交互，两者的共享存储根路径需要保持一致。共享目录的根路径属主为9000用户，与ClusterD运行用户一致。

    1.  配置server。

        ![](../../figures/scheduling/zh-cn_image_0000002479386634.png)

    2.  修改NodeD配置。

        ![](../../figures/scheduling/zh-cn_image_0000002479386638.png)

    3.  如果存在ClusterD，则需修改ClusterD配置。

        ![](../../figures/scheduling/zh-cn_image_0000002511346583.png)

    4.  执行**kubectl get pods -o -wide -A**命令出现如下示例，则表示已完成共享存储配置。

        ![](../../figures/scheduling/zh-cn_image_0000002479226664.png)

2.  启用或关闭灵衢网络检测。
    -   （推荐）已安装Ascend Device Plugin和ClusterD
        1.  登录环境，进入NodeD解压目录。
        2.  执行以下命令创建名为pingmesh-config的ConfigMap文件。

            pingmesh-config.yaml为pingmesh配置文件，可从NodeD安装包中获取。

            ```
            kubectl apply -f pingmesh-config.yaml  
            ```

            回显示例如下。

            ```
            configmap/pingmesh-config created
            ```

        3.  执行以下命令编辑pingmesh-config文件。该文件中各参数的填写说明如[表1](#zh-cn_topic_0000002193288232_table985012534578)所示。

            ```
            kubectl edit cm -n cluster-system   pingmesh-config
            ```

            **表 1**  pingmesh-config cm

            <a name="zh-cn_topic_0000002193288232_table985012534578"></a>
            <table><thead align="left"><tr id="zh-cn_topic_0000002193288232_row9850195355712"><th class="cellrowborder" valign="top" width="18.86188618861886%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002193288232_p38501532579"><a name="zh-cn_topic_0000002193288232_p38501532579"></a><a name="zh-cn_topic_0000002193288232_p38501532579"></a>参数</p>
            </th>
            <th class="cellrowborder" valign="top" width="67.65676567656766%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002193288232_p2850165315579"><a name="zh-cn_topic_0000002193288232_p2850165315579"></a><a name="zh-cn_topic_0000002193288232_p2850165315579"></a>说明</p>
            </th>
            <th class="cellrowborder" valign="top" width="13.481348134813482%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002193288232_p985065320578"><a name="zh-cn_topic_0000002193288232_p985065320578"></a><a name="zh-cn_topic_0000002193288232_p985065320578"></a>取值</p>
            </th>
            </tr>
            </thead>
            <tbody><tr id="zh-cn_topic_0000002193288232_row1885095315578"><td class="cellrowborder" valign="top" width="18.86188618861886%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p1785015355715"><a name="zh-cn_topic_0000002193288232_p1785015355715"></a><a name="zh-cn_topic_0000002193288232_p1785015355715"></a>app</p>
            </td>
            <td class="cellrowborder" valign="top" width="67.65676567656766%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p138501353195716"><a name="zh-cn_topic_0000002193288232_p138501353195716"></a><a name="zh-cn_topic_0000002193288232_p138501353195716"></a>ConfigMap其中一个label的key。</p>
            </td>
            <td class="cellrowborder" valign="top" width="13.481348134813482%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p138504535570"><a name="zh-cn_topic_0000002193288232_p138504535570"></a><a name="zh-cn_topic_0000002193288232_p138504535570"></a>pingmesh</p>
            </td>
            </tr>
            <tr id="zh-cn_topic_0000002193288232_row68509536570"><td class="cellrowborder" valign="top" width="18.86188618861886%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p1185045317579"><a name="zh-cn_topic_0000002193288232_p1185045317579"></a><a name="zh-cn_topic_0000002193288232_p1185045317579"></a>global</p>
            </td>
            <td class="cellrowborder" valign="top" width="67.65676567656766%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p9850135314575"><a name="zh-cn_topic_0000002193288232_p9850135314575"></a><a name="zh-cn_topic_0000002193288232_p9850135314575"></a>集群配置信息</p>
            </td>
            <td class="cellrowborder" valign="top" width="13.481348134813482%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p3850165314571"><a name="zh-cn_topic_0000002193288232_p3850165314571"></a><a name="zh-cn_topic_0000002193288232_p3850165314571"></a>-</p>
            </td>
            </tr>
            <tr id="zh-cn_topic_0000002193288232_row9850185313579"><td class="cellrowborder" valign="top" width="18.86188618861886%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p1585095395717"><a name="zh-cn_topic_0000002193288232_p1585095395717"></a><a name="zh-cn_topic_0000002193288232_p1585095395717"></a>"1"</p>
            </td>
            <td class="cellrowborder" valign="top" width="67.65676567656766%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p385045385710"><a name="zh-cn_topic_0000002193288232_p385045385710"></a><a name="zh-cn_topic_0000002193288232_p385045385710"></a>超节点ID为1的配置示例，用户可根据实际情况进行修改或新增。当配置了某个超节点后，NodeD会采用超节点的配置信息而忽略global配置信息。</p>
            </td>
            <td class="cellrowborder" valign="top" width="13.481348134813482%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p1085025315711"><a name="zh-cn_topic_0000002193288232_p1085025315711"></a><a name="zh-cn_topic_0000002193288232_p1085025315711"></a>超节点ID</p>
            </td>
            </tr>
            <tr id="zh-cn_topic_0000002193288232_row1585014537573"><td class="cellrowborder" valign="top" width="18.86188618861886%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p12850165395719"><a name="zh-cn_topic_0000002193288232_p12850165395719"></a><a name="zh-cn_topic_0000002193288232_p12850165395719"></a>activate</p>
            </td>
            <td class="cellrowborder" valign="top" width="67.65676567656766%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p138500538571"><a name="zh-cn_topic_0000002193288232_p138500538571"></a><a name="zh-cn_topic_0000002193288232_p138500538571"></a>是否启用pingmesh功能。</p>
            </td>
            <td class="cellrowborder" valign="top" width="13.481348134813482%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p1850153195711"><a name="zh-cn_topic_0000002193288232_p1850153195711"></a><a name="zh-cn_topic_0000002193288232_p1850153195711"></a>on或off</p>
            </td>
            </tr>
            <tr id="zh-cn_topic_0000002193288232_row28501353175710"><td class="cellrowborder" valign="top" width="18.86188618861886%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p5850953125712"><a name="zh-cn_topic_0000002193288232_p5850953125712"></a><a name="zh-cn_topic_0000002193288232_p5850953125712"></a>task_interval</p>
            </td>
            <td class="cellrowborder" valign="top" width="67.65676567656766%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1585015310572"><a name="zh-cn_topic_0000002193288232_p1585015310572"></a><a name="zh-cn_topic_0000002193288232_p1585015310572"></a>pingmesh任务间隔。单位为秒。</p>
            </td>
            <td class="cellrowborder" valign="top" width="13.481348134813482%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p5850165335719"><a name="zh-cn_topic_0000002193288232_p5850165335719"></a><a name="zh-cn_topic_0000002193288232_p5850165335719"></a>[1~60]</p>
            </td>
            </tr>
            </tbody>
            </table>

    -   未安装Ascend Device Plugin和ClusterD

        自行生成名为cluster-system的命名空间， name为super-pod-<superPodID\>、label为app=pingmesh的ConfigMap。且该ConfigMap中各字段需按照[cluster-system super-pod-<super-pod-id\>](../api/clusterd.md#集群资源)表填写。示例如下。

        ```
        apiVersion: v1
        data:
          superPodDevice: '{"SuperPodID":"0","NodeDeviceMap":{"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"62914560","1":"62980097","10":"64225290","11":"64290827","12":"64487436","13":"64552973","14":"64749582","15":"64815119","2":"63176706","3":"63242243","4":"63438852","5":"63504389","6":"63700998","7":"63766535","8":"63963144","9":"64028681"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"67108864","1":"67174401","10":"68419594","11":"68485131","12":"68681740","13":"68747277","14":"68943886","15":"69009423","2":"67371010","3":"67436547","4":"67633156","5":"67698693","6":"67895302","7":"67960839","8":"68157448","9":"68222985"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"104857600","1":"104923137","10":"106168330","11":"106233867","12":"106430476","13":"106496013","14":"106692622","15":"106758159","2":"105119746","3":"105185283","4":"105381892","5":"105447429","6":"105644038","7":"105709575","8":"105906184","9":"105971721"}},"node-**-*":{"NodeName":"node-**-*","DeviceMap":{"0":"4194304","1":"4259841","10":"5505034","11":"5570571","12":"5767180","13":"5832717","14":"6029326","15":"6094863","2":"4456450","3":"4521987","4":"4718596","5":"4784133","6":"4980742","7":"5046279","8":"5242888","9":"5308425"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"142606336","1":"142671873","10":"143917066","11":"143982603","12":"144179212","13":"144244749","14":"144441358","15":"144506895","2":"142868482","3":"142934019","4":"143130628","5":"143196165","6":"143392774","7":"143458311","8":"143654920","9":"143720457"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"146800640","1":"146866177","10":"148111370","11":"148176907","12":"148373516","13":"148439053","14":"148635662","15":"148701199","2":"147062786","3":"147128323","4":"147324932","5":"147390469","6":"147587078","7":"147652615","8":"147849224","9":"147914761"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"83886080","1":"83951617","10":"85196810","11":"85262347","12":"85458956","13":"85524493","14":"85721102","15":"85786639","2":"84148226","3":"84213763","4":"84410372","5":"84475909","6":"84672518","7":"84738055","8":"84934664","9":"85000201"}}}}'
        kind: ConfigMap
        metadata:
          labels:
            app: pingmesh
          name: super-pod-0       # 0为超节点ID
          namespace: cluster-system
        ```

**查看检测结果信息<a name="zh-cn_topic_0000002193288232_section772614207398"></a>**

>[!NOTE] 说明 
>检测结果查询周期为配置参数“task\_interval“的10倍。

灵衢网络检测的pingmesh结果写入文件<nodename\>.log中。该文件中各字段的详细说明如下表所示。

**表 2**  <nodename\>.log

<a name="zh-cn_topic_0000002193288232_table313985322113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002193288232_row9139145315219"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002193288232_p814016532215"><a name="zh-cn_topic_0000002193288232_p814016532215"></a><a name="zh-cn_topic_0000002193288232_p814016532215"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002193288232_p5140165319210"><a name="zh-cn_topic_0000002193288232_p5140165319210"></a><a name="zh-cn_topic_0000002193288232_p5140165319210"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002193288232_p1140353102113"><a name="zh-cn_topic_0000002193288232_p1140353102113"></a><a name="zh-cn_topic_0000002193288232_p1140353102113"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002193288232_row111401453172110"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p1214065317213"><a name="zh-cn_topic_0000002193288232_p1214065317213"></a><a name="zh-cn_topic_0000002193288232_p1214065317213"></a>uid</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p9140175310211"><a name="zh-cn_topic_0000002193288232_p9140175310211"></a><a name="zh-cn_topic_0000002193288232_p9140175310211"></a>该次pingmesh任务的ID。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p1140353192111"><a name="zh-cn_topic_0000002193288232_p1140353192111"></a><a name="zh-cn_topic_0000002193288232_p1140353192111"></a>长度为64的字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row714035319219"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p181401253102117"><a name="zh-cn_topic_0000002193288232_p181401253102117"></a><a name="zh-cn_topic_0000002193288232_p181401253102117"></a>config</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1914013539210"><a name="zh-cn_topic_0000002193288232_p1914013539210"></a><a name="zh-cn_topic_0000002193288232_p1914013539210"></a>该次pingmesh任务的用户配置。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p11405537216"><a name="zh-cn_topic_0000002193288232_p11405537216"></a><a name="zh-cn_topic_0000002193288232_p11405537216"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row814010533215"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p7140185332115"><a name="zh-cn_topic_0000002193288232_p7140185332115"></a><a name="zh-cn_topic_0000002193288232_p7140185332115"></a>physicID</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1014015332114"><a name="zh-cn_topic_0000002193288232_p1014015332114"></a><a name="zh-cn_topic_0000002193288232_p1014015332114"></a>NPU卡物理ID。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p91401653122115"><a name="zh-cn_topic_0000002193288232_p91401653122115"></a><a name="zh-cn_topic_0000002193288232_p91401653122115"></a>[0~15]</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row1092154019225"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p1092114406228"><a name="zh-cn_topic_0000002193288232_p1092114406228"></a><a name="zh-cn_topic_0000002193288232_p1092114406228"></a>taskID</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p179211740172215"><a name="zh-cn_topic_0000002193288232_p179211740172215"></a><a name="zh-cn_topic_0000002193288232_p179211740172215"></a>任务ID，0代表节点内部、1代表节点间。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p149211840132218"><a name="zh-cn_topic_0000002193288232_p149211840132218"></a><a name="zh-cn_topic_0000002193288232_p149211840132218"></a>0或1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row9947144615229"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p15947546152212"><a name="zh-cn_topic_0000002193288232_p15947546152212"></a><a name="zh-cn_topic_0000002193288232_p15947546152212"></a>DestNum</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1694711467225"><a name="zh-cn_topic_0000002193288232_p1694711467225"></a><a name="zh-cn_topic_0000002193288232_p1694711467225"></a>本次pingmesh目标地址数量。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p16947646172215"><a name="zh-cn_topic_0000002193288232_p16947646172215"></a><a name="zh-cn_topic_0000002193288232_p16947646172215"></a>[0~47]</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row22024432213"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p2202445229"><a name="zh-cn_topic_0000002193288232_p2202445229"></a><a name="zh-cn_topic_0000002193288232_p2202445229"></a>source_addr</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p102013447221"><a name="zh-cn_topic_0000002193288232_p102013447221"></a><a name="zh-cn_topic_0000002193288232_p102013447221"></a>源地址</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p4203446223"><a name="zh-cn_topic_0000002193288232_p4203446223"></a><a name="zh-cn_topic_0000002193288232_p4203446223"></a>IPv4网络地址</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row66281311112311"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p36281111192311"><a name="zh-cn_topic_0000002193288232_p36281111192311"></a><a name="zh-cn_topic_0000002193288232_p36281111192311"></a>target_addr</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p15628311132314"><a name="zh-cn_topic_0000002193288232_p15628311132314"></a><a name="zh-cn_topic_0000002193288232_p15628311132314"></a>目标地址</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p176281611112319"><a name="zh-cn_topic_0000002193288232_p176281611112319"></a><a name="zh-cn_topic_0000002193288232_p176281611112319"></a>IPv4网络地址</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row1246282342317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p246218233238"><a name="zh-cn_topic_0000002193288232_p246218233238"></a><a name="zh-cn_topic_0000002193288232_p246218233238"></a>suc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p194621623162317"><a name="zh-cn_topic_0000002193288232_p194621623162317"></a><a name="zh-cn_topic_0000002193288232_p194621623162317"></a>发送成功的包数量。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p13462112317238"><a name="zh-cn_topic_0000002193288232_p13462112317238"></a><a name="zh-cn_topic_0000002193288232_p13462112317238"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row19827635192318"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p582717356232"><a name="zh-cn_topic_0000002193288232_p582717356232"></a><a name="zh-cn_topic_0000002193288232_p582717356232"></a>fail_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p18827163511233"><a name="zh-cn_topic_0000002193288232_p18827163511233"></a><a name="zh-cn_topic_0000002193288232_p18827163511233"></a>发送失败的包数量。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p96135810315"><a name="zh-cn_topic_0000002193288232_p96135810315"></a><a name="zh-cn_topic_0000002193288232_p96135810315"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row1062174552313"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p19621154592319"><a name="zh-cn_topic_0000002193288232_p19621154592319"></a><a name="zh-cn_topic_0000002193288232_p19621154592319"></a>max_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p116211445162317"><a name="zh-cn_topic_0000002193288232_p116211445162317"></a><a name="zh-cn_topic_0000002193288232_p116211445162317"></a>最长响应时间</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul123111616574"></a><a name="ul123111616574"></a><ul id="ul123111616574"><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row13637135652312"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p116371556172312"><a name="zh-cn_topic_0000002193288232_p116371556172312"></a><a name="zh-cn_topic_0000002193288232_p116371556172312"></a>min_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1183124763011"><a name="zh-cn_topic_0000002193288232_p1183124763011"></a><a name="zh-cn_topic_0000002193288232_p1183124763011"></a>最短响应时间</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul8487153175715"></a><a name="ul8487153175715"></a><ul id="ul8487153175715"><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row161362219246"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p101360210240"><a name="zh-cn_topic_0000002193288232_p101360210240"></a><a name="zh-cn_topic_0000002193288232_p101360210240"></a>avg_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p6522651103018"><a name="zh-cn_topic_0000002193288232_p6522651103018"></a><a name="zh-cn_topic_0000002193288232_p6522651103018"></a>平均响应时间</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul1542252195912"></a><a name="ul1542252195912"></a><ul id="ul1542252195912"><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row17545059142313"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p11545659192314"><a name="zh-cn_topic_0000002193288232_p11545659192314"></a><a name="zh-cn_topic_0000002193288232_p11545659192314"></a>tp95_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1354555914234"><a name="zh-cn_topic_0000002193288232_p1354555914234"></a><a name="zh-cn_topic_0000002193288232_p1354555914234"></a>处于95%位置的响应时间。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul1746711165919"></a><a name="ul1746711165919"></a><ul id="ul1746711165919"><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row1951626122416"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p29542618248"><a name="zh-cn_topic_0000002193288232_p29542618248"></a><a name="zh-cn_topic_0000002193288232_p29542618248"></a>reply_stat_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1495326202416"><a name="zh-cn_topic_0000002193288232_p1495326202416"></a><a name="zh-cn_topic_0000002193288232_p1495326202416"></a>本次查询到的响应数量。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p895192652413"><a name="zh-cn_topic_0000002193288232_p895192652413"></a><a name="zh-cn_topic_0000002193288232_p895192652413"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row62832386245"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p1628343810249"><a name="zh-cn_topic_0000002193288232_p1628343810249"></a><a name="zh-cn_topic_0000002193288232_p1628343810249"></a>ping_total_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p12831538122417"><a name="zh-cn_topic_0000002193288232_p12831538122417"></a><a name="zh-cn_topic_0000002193288232_p12831538122417"></a>本次任务累计的响应数量。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p928313842416"><a name="zh-cn_topic_0000002193288232_p928313842416"></a><a name="zh-cn_topic_0000002193288232_p928313842416"></a>-</p>
</td>
</tr>
</tbody>
</table>

**查看故障信息<a name="zh-cn_topic_0000002193288232_section7712929183110"></a>**

在管理节点上执行以下命令，查看灵衢网络检测的故障信息。

```
kubectl describe cm -n cluster-system  pingmesh-fault-<nodename>
```

故障信息中各字段的详细说明如下所示。

**表 3**  pingmesh-fault-<nodename\>

<a name="zh-cn_topic_0000002193288232_table2371535113510"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002193288232_row10378359354"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002193288232_p23763516354"><a name="zh-cn_topic_0000002193288232_p23763516354"></a><a name="zh-cn_topic_0000002193288232_p23763516354"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002193288232_p0371335143511"><a name="zh-cn_topic_0000002193288232_p0371335143511"></a><a name="zh-cn_topic_0000002193288232_p0371335143511"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002193288232_p9371635163516"><a name="zh-cn_topic_0000002193288232_p9371635163516"></a><a name="zh-cn_topic_0000002193288232_p9371635163516"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002193288232_row237123519354"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p11375353357"><a name="zh-cn_topic_0000002193288232_p11375353357"></a><a name="zh-cn_topic_0000002193288232_p11375353357"></a>mc-consumer-publicfault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p837335103515"><a name="zh-cn_topic_0000002193288232_p837335103515"></a><a name="zh-cn_topic_0000002193288232_p837335103515"></a><span id="zh-cn_topic_0000002193288232_ph9255132215122"><a name="zh-cn_topic_0000002193288232_ph9255132215122"></a><a name="zh-cn_topic_0000002193288232_ph9255132215122"></a>ClusterD</span>侦听所需的label key</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p73743563520"><a name="zh-cn_topic_0000002193288232_p73743563520"></a><a name="zh-cn_topic_0000002193288232_p73743563520"></a>true</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002193288232_row203793520353"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002193288232_p16371535143510"><a name="zh-cn_topic_0000002193288232_p16371535143510"></a><a name="zh-cn_topic_0000002193288232_p16371535143510"></a>PublicFault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1937113573511"><a name="zh-cn_topic_0000002193288232_p1937113573511"></a><a name="zh-cn_topic_0000002193288232_p1937113573511"></a>公共故障信息key</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002193288232_p153833563511"><a name="zh-cn_topic_0000002193288232_p153833563511"></a><a name="zh-cn_topic_0000002193288232_p153833563511"></a>详细说明请参见<a href="../api/clusterd.md#configmap">fault字段说明</a>表。</p>
</td>
</tr>
</tbody>
</table>

**已支持的灵衢网络故障<a name="zh-cn_topic_0000002193288232_section4960201383813"></a>**

<a name="zh-cn_topic_0000002193288232_table31451934163811"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002193288232_row514523493819"><th class="cellrowborder" valign="top" width="12.26122612261226%" id="mcps1.1.4.1.1"><p id="zh-cn_topic_0000002193288232_p1114523420389"><a name="zh-cn_topic_0000002193288232_p1114523420389"></a><a name="zh-cn_topic_0000002193288232_p1114523420389"></a>故障码</p>
</th>
<th class="cellrowborder" valign="top" width="18.651865186518652%" id="mcps1.1.4.1.2"><p id="zh-cn_topic_0000002193288232_p9145143412387"><a name="zh-cn_topic_0000002193288232_p9145143412387"></a><a name="zh-cn_topic_0000002193288232_p9145143412387"></a>故障说明</p>
</th>
<th class="cellrowborder" valign="top" width="69.08690869086908%" id="mcps1.1.4.1.3"><p id="zh-cn_topic_0000002193288232_p15145193413388"><a name="zh-cn_topic_0000002193288232_p15145193413388"></a><a name="zh-cn_topic_0000002193288232_p15145193413388"></a>故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002193288232_row131782214434"><td class="cellrowborder" valign="top" width="12.26122612261226%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0000002193288232_p41752216438"><a name="zh-cn_topic_0000002193288232_p41752216438"></a><a name="zh-cn_topic_0000002193288232_p41752216438"></a>220001001</p>
</td>
<td class="cellrowborder" valign="top" width="18.651865186518652%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0000002193288232_p1171822134316"><a name="zh-cn_topic_0000002193288232_p1171822134316"></a><a name="zh-cn_topic_0000002193288232_p1171822134316"></a>NPU卡<span id="ph612871931615"><a name="ph612871931615"></a><a name="ph612871931615"></a>HCCS</span>网络故障</p>
</td>
<td class="cellrowborder" valign="top" width="69.08690869086908%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0000002193288232_p1566710511444"><a name="zh-cn_topic_0000002193288232_p1566710511444"></a><a name="zh-cn_topic_0000002193288232_p1566710511444"></a>SeparateNPU</p>
<div class="note" id="zh-cn_topic_0000002193288232_note13181172251512"><a name="zh-cn_topic_0000002193288232_note13181172251512"></a><a name="zh-cn_topic_0000002193288232_note13181172251512"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002193288232_p17181722121510"><a name="zh-cn_topic_0000002193288232_p17181722121510"></a><a name="zh-cn_topic_0000002193288232_p17181722121510"></a>该故障级别不支持自行配置。</p>
</div></div>
</td>
</tr>
</tbody>
</table>


#### 性能劣化故障<a name="ZH-CN_TOPIC_0000002479386488"></a>

##### 使用7.1.RC1及以上版本TaskD<a name="ZH-CN_TOPIC_0000002511346475"></a>

MindCluster集群调度组件结合MindStudio提供的profiling能力，对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态打点和打点数据持久化功能、可动态启停训练任务打点功能，无需重启任务进行诊断，对训练无损耗。

当前支持的打点数据如[表1](#zh-cn_topic_0000002194466236_table5530103025919)所示。

**表 1**  打点数据说明

<a name="zh-cn_topic_0000002194466236_table5530103025919"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002194466236_row105301830105911"><th class="cellrowborder" valign="top" width="21.12%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002194466236_p7530133018591"><a name="zh-cn_topic_0000002194466236_p7530133018591"></a><a name="zh-cn_topic_0000002194466236_p7530133018591"></a>打点数据的类型</p>
</th>
<th class="cellrowborder" valign="top" width="26.540000000000003%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002194466236_p165301308591"><a name="zh-cn_topic_0000002194466236_p165301308591"></a><a name="zh-cn_topic_0000002194466236_p165301308591"></a>支持的AI框架</p>
</th>
<th class="cellrowborder" valign="top" width="52.339999999999996%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002194466236_p153043055911"><a name="zh-cn_topic_0000002194466236_p153043055911"></a><a name="zh-cn_topic_0000002194466236_p153043055911"></a>提供支持的组件</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002194466236_row1753023011598"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p1692418239310"><a name="zh-cn_topic_0000002194466236_p1692418239310"></a><a name="zh-cn_topic_0000002194466236_p1692418239310"></a>FP</p>
<p id="zh-cn_topic_0000002194466236_p195301030165916"><a name="zh-cn_topic_0000002194466236_p195301030165916"></a><a name="zh-cn_topic_0000002194466236_p195301030165916"></a>（标识前向传播数据）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p1353016307593"><a name="zh-cn_topic_0000002194466236_p1353016307593"></a><a name="zh-cn_topic_0000002194466236_p1353016307593"></a>PyTorch</p>
<div class="note" id="zh-cn_topic_0000002194466236_note8765349888"><a name="zh-cn_topic_0000002194466236_note8765349888"></a><a name="zh-cn_topic_0000002194466236_note8765349888"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000002194466236_p1276511491787"><a name="zh-cn_topic_0000002194466236_p1276511491787"></a><a name="zh-cn_topic_0000002194466236_p1276511491787"></a>仅支持单算子场景。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002194466236_p5530630205914"><a name="zh-cn_topic_0000002194466236_p5530630205914"></a><a name="zh-cn_topic_0000002194466236_p5530630205914"></a>mstx_torch_plugin</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row13415176111"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p3788646727"><a name="zh-cn_topic_0000002194466236_p3788646727"></a><a name="zh-cn_topic_0000002194466236_p3788646727"></a>Step</p>
<p id="zh-cn_topic_0000002194466236_p2341175120"><a name="zh-cn_topic_0000002194466236_p2341175120"></a><a name="zh-cn_topic_0000002194466236_p2341175120"></a>（标识Step时延）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p834517411"><a name="zh-cn_topic_0000002194466236_p834517411"></a><a name="zh-cn_topic_0000002194466236_p834517411"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul17713185812916"></a><a name="zh-cn_topic_0000002194466236_ul17713185812916"></a><ul id="zh-cn_topic_0000002194466236_ul17713185812916"><li>PyTorch<a name="ul144271853105714"></a><a name="ul144271853105714"></a><ul id="ul144271853105714"><li>原生优化器场景：若torch_npu为7.1.RC1版本，需使用mstx_torch_plugin；若torch_npu为7.1.RC1以上版本，无需使用mstx_torch_plugin，torch_npu自带Step打点。</li><li>自定义优化器场景：手动增加打点数据。</li></ul>
</li><li>MindSpore<a name="zh-cn_topic_0000002194466236_ul4814121617106"></a><a name="zh-cn_topic_0000002194466236_ul4814121617106"></a><ul id="zh-cn_topic_0000002194466236_ul4814121617106"><li>MindFormers场景：Step打点数据由MindFormers提供。</li><li>MindSpeed场景：不提供Step打点数据。</li></ul>
</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row9530630195919"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p169247495217"><a name="zh-cn_topic_0000002194466236_p169247495217"></a><a name="zh-cn_topic_0000002194466236_p169247495217"></a>Communication</p>
<p id="zh-cn_topic_0000002194466236_p1753012305591"><a name="zh-cn_topic_0000002194466236_p1753012305591"></a><a name="zh-cn_topic_0000002194466236_p1753012305591"></a>（标识通信算子）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p2401691783"><a name="zh-cn_topic_0000002194466236_p2401691783"></a><a name="zh-cn_topic_0000002194466236_p2401691783"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul18432034131111"></a><a name="zh-cn_topic_0000002194466236_ul18432034131111"></a><ul id="zh-cn_topic_0000002194466236_ul18432034131111"><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row953063010598"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p112455417213"><a name="zh-cn_topic_0000002194466236_p112455417213"></a><a name="zh-cn_topic_0000002194466236_p112455417213"></a>SaveCheckpoint</p>
<p id="zh-cn_topic_0000002194466236_p353043014592"><a name="zh-cn_topic_0000002194466236_p353043014592"></a><a name="zh-cn_topic_0000002194466236_p353043014592"></a>（标识SaveCheckpoint耗时）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p20176311281"><a name="zh-cn_topic_0000002194466236_p20176311281"></a><a name="zh-cn_topic_0000002194466236_p20176311281"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul1450711814124"></a><a name="zh-cn_topic_0000002194466236_ul1450711814124"></a><ul id="zh-cn_topic_0000002194466236_ul1450711814124"><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row1234614241805"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p1271312516318"><a name="zh-cn_topic_0000002194466236_p1271312516318"></a><a name="zh-cn_topic_0000002194466236_p1271312516318"></a>DataLoader</p>
<p id="zh-cn_topic_0000002194466236_p834615241902"><a name="zh-cn_topic_0000002194466236_p834615241902"></a><a name="zh-cn_topic_0000002194466236_p834615241902"></a>（标识DataLoader耗时）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p459471310811"><a name="zh-cn_topic_0000002194466236_p459471310811"></a><a name="zh-cn_topic_0000002194466236_p459471310811"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul827633418123"></a><a name="zh-cn_topic_0000002194466236_ul827633418123"></a><ul id="zh-cn_topic_0000002194466236_ul827633418123"><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架。</li></ul>
</td>
</tr>
</tbody>
</table>

**使用约束<a name="zh-cn_topic_0000002194466236_section487603614"></a>**

-   当前Step、SaveCheckpoint、FP、DataLoader仅支持同步开启。如需关闭以上四类打点数据，需同时关闭Communication。
-   Communication通信算子数据支持单独开启、关闭。
-   动态轻量打点功能与MindStudio的全量打点功能不可同时开启，开启全量打点功能会导致性能劣化故障不能正常采集数据。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   （可选）已安装[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装[torch\_npu](#制作mindspeed-llm训练镜像pytorch框架)（**可选**，PyTorch场景需安装、版本号≥7.1.RC1）、MindSpore（**可选**，MindSpore场景需安装、版本号≥2.7.0）、[CANN](#制作mindformers训练镜像mindspore框架)（**必选**，版本号≥8.2.RC1）、[TaskD](#制作mindformers训练镜像mindspore框架)（**必选**）

**准备软件包<a name="zh-cn_topic_0000002194466236_section8281518121516"></a>**

**表 2**  准备软件包

<a name="zh-cn_topic_0000002194466236_table232305471415"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002194466236_row33231354121420"><th class="cellrowborder" valign="top" width="14.08%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000002194466236_p1441653254"><a name="zh-cn_topic_0000002194466236_p1441653254"></a><a name="zh-cn_topic_0000002194466236_p1441653254"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="6.74%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000002194466236_p2052053751"><a name="zh-cn_topic_0000002194466236_p2052053751"></a><a name="zh-cn_topic_0000002194466236_p2052053751"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="60.75000000000001%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000002194466236_p657531455"><a name="zh-cn_topic_0000002194466236_p657531455"></a><a name="zh-cn_topic_0000002194466236_p657531455"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="8.58%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000002194466236_p1859531759"><a name="zh-cn_topic_0000002194466236_p1859531759"></a><a name="zh-cn_topic_0000002194466236_p1859531759"></a>获取方法</p>
</th>
<th class="cellrowborder" valign="top" width="9.85%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000002194466236_p1444313132617"><a name="zh-cn_topic_0000002194466236_p1444313132617"></a><a name="zh-cn_topic_0000002194466236_p1444313132617"></a>使用场景</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002194466236_row1715722244511"><td class="cellrowborder" valign="top" width="14.08%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000002194466236_p19157182244515"><a name="zh-cn_topic_0000002194466236_p19157182244515"></a><a name="zh-cn_topic_0000002194466236_p19157182244515"></a>mstx_torch_plugin</p>
</td>
<td class="cellrowborder" valign="top" width="6.74%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000002194466236_p31576228454"><a name="zh-cn_topic_0000002194466236_p31576228454"></a><a name="zh-cn_topic_0000002194466236_p31576228454"></a>否</p>
</td>
<td class="cellrowborder" valign="top" width="60.75000000000001%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000002194466236_p15157192234518"><a name="zh-cn_topic_0000002194466236_p15157192234518"></a><a name="zh-cn_topic_0000002194466236_p15157192234518"></a><span>Ascend PyTorch Profiler中的</span><a href="https://www.hiascend.com/document/detail/zh/canncommercial/800/devaids/devtools/profiling/atlasprofiling_16_0033.html" target="_blank" rel="noopener noreferrer">采集并解析msproftx数据</a><span>功能已经内置了通信算子的打点。为了方便用户在不修改业务代码的基础上获取更多关键阶段的耗时数据，mstx_torch_plugin在Ascend PyTorch Profiler内置了</span><span>dataloader</span><span>、</span><span>forward</span><span>、</span><span>step</span><span>、</span><span>save_checkpoint</span><span>这四个关键阶段函数的打点。</span></p>
<div class="note" id="zh-cn_topic_0000002194466236_note179451154301"><a name="zh-cn_topic_0000002194466236_note179451154301"></a><a name="zh-cn_topic_0000002194466236_note179451154301"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="zh-cn_topic_0000002194466236_ul17791934115613"></a><a name="zh-cn_topic_0000002194466236_ul17791934115613"></a><ul id="zh-cn_topic_0000002194466236_ul17791934115613"><li>如需使用FP打点数据，需安装mstx_torch_plugin。其他场景下无需安装。</li><li>需使用1.0及以上版本的mstx_torch_plugin。</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="8.58%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000002194466236_p5157162210453"><a name="zh-cn_topic_0000002194466236_p5157162210453"></a><a name="zh-cn_topic_0000002194466236_p5157162210453"></a><a href="https://gitee.com/link?target=https://ptdbg.obs.myhuaweicloud.com/profiler/example/1.0/mstx_torch_plugin-1.0-py3-none-any.whl" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
<td class="cellrowborder" valign="top" width="9.85%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000002194466236_p1244314111268"><a name="zh-cn_topic_0000002194466236_p1244314111268"></a><a name="zh-cn_topic_0000002194466236_p1244314111268"></a>PyTorch</p>
</td>
</tr>
</tbody>
</table>

**配置性能劣化故障检测<a name="section1831691464111"></a>**

本方案仅针对7.1.RC1及以上版本的TaskD组件。如使用7.1.RC1以下版本的组件请参考[使用其他版本TaskD](#使用其他版本taskd)章节进行操作。

1.  以下两种方式请根据实际需要进行二选一。
    -   在容器内安装mstx\_torch\_plugin。
        1.  下载mstx\_torch\_plugin的whl包。whl包链接：[mstx\_torch\_plugin](https://gitee.com/link?target=https%3A%2F%2Fptdbg.obs.myhuaweicloud.com%2Fprofiler%2Fexample%2F1.0%2Fmstx_torch_plugin-1.0-py3-none-any.whl)。
        2.  安装软件包。

            ```
            pip install mstx_torch_plugin-1.0-py3-none-any.whl
            ```

        3.  在AI任务执行脚本中import导入该whl包。

            需保证import的顺序在import torch和import torch\_npu后面，示例如下。

            ```
            import torch 
            import torch_npu  
            import mstx_torch_plugin
            ```

    -   在PyTorch场景下，非原生优化器或不使用mstx\_torch\_plugin的情况下，为获取训练的Step耗时数据需修改训练脚本中的训练迭代循环，需增加Step打点代码。

        以下示例为PyTorch-MindSpeed场景，需修改./mindspeed\_llm/training/training.py文件增加如下加粗字段。MindSpore场景请根据实际情况修改。

        ```
        def train(forward_step_func, model, optimizer, opt_param_scheduler,
                  train_data_iterator, valid_data_iterator,
                  process_non_loss_data_func, config):
                    # Cache into one-logger for callback
            ……
            ……
            if is_profile_enabled():
                prof = get_profiler()
                prof.start()
            step_id = iteration
            while iteration < args.train_iters:
                stream = torch.npu.current_stream()      # 获取当前环境的执行流，用于获取NPU侧时间
                range_id = torch.npu.mstx.range_start(f"step {step_id}", stream) # 标识当前训练step的开始
                ……
                ……
                if args.manual_gc:
                    if args.manual_gc_interval != 0 and iteration % args.manual_gc_interval == 0:
                        gc.collect()
        
                if is_profile_enabled():
                    prof.step()
                step_id +=1  # 训练step加一，用于标识下一step
                torch.npu.mstx.range_end(range_id) # 标识当前训练step的结束
        ```

2.  在容器内，以CANN软件包的运行用户登录环境，执行**source $\{install\_path\}/set\_env.sh**命令设置环境变量。其中$\{install\_path\}为CANN软件的安装目录。示例如下。

    ```
    source /usr/local/Ascend/cann/set_env.sh
    ```

3.  训练启动前，在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。

    ```
    export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    ```

    -   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。

    -   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：  TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。

        TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。

        ```
        pip show taskd
        ```

4.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
    1.  <a name="li399811541"></a>（可选）拉起TaskD  Manager和TaskD  Proxy。若通过gRPC接口方式开启轻量profiling获取落盘数据，则需执行如下步骤；若通过ConfigMap方式开启轻量profiling获取落盘数据，则跳过该步骤。
        -   **PyTorch场景**
            1.  创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

                ```
                from taskd.api import init_taskd_manager, start_taskd_manager
                import os
                
                job_id=os.getenv("MINDX_TASK_ID")
                node_nums=XX         # 用户填入任务节点总数
                proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
                
                init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
                start_taskd_manager()
                ```

                >[!NOTE] 说明 
                >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

            2.  在训练脚本中增加以下代码，拉起TaskD  Manager和TaskD  Proxy。

                ```
                sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
                
                if [[ "${RANK}" -eq 0 ]]; then
                    export MASTER_ADDR=${POD_IP}
                    python manager.py &
                fi
                    
                torchrun ...
                ```

        -   **MindSpore场景**
            1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

                ```
                from taskd.api import init_taskd_manager, start_taskd_manager
                import os
                
                job_id=os.getenv("MINDX_TASK_ID")
                node_nums=XX         # 用户填入任务节点总数
                proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
                
                init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
                start_taskd_manager()
                ```

                >[!NOTE] 说明 
                >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

            2.  在训练脚本中增加以下代码拉起TaskD  Manager。

                ```
                if [[ "${MS_SCHED_HOST}" -eq "${POD_IP}" ]]; then
                    python manager.py &
                fi
                    
                msrun ...
                ```

            3.  修改mindspore/python/mindspore/parallel/cluster/process\_entity/\_api.py文件，拉起TaskD  Proxy。示例如下。

                ```
                ...
                  if ("TTP:1" in tft_env) or ("UCE:1" in tft_env) or ("ARF:1" in tft_env):
                            try:
                                from taskd.python.framework.agent.ms_mgr.msrun_plugin import MSRunPlugin
                                from taskd .api.taskd_proxy_api import init_taskd_proxy
                                from taskd.python.framework.common.type import CONFIG_UPSTREAMIP_KEY, LOCAL_HOST
                              import threading
                              proxy = threading.Thread(target=init_taskd_proxy, args=({CONFIG_UPSTREAMIP_KEY : os.getenv("MS_SCHED_HOST", LOCAL_HOST)},))
                              proxy.daemon = True
                              proxy.start()
                                self.msmgr = MSRunPlugin()
                                self.msmgr.register_callbacks("KILL_WORKER", self.kill_workers)
                                self.msmgr.register_callbacks("START_ALL_WORKER", self.start_all_workers)
                                self.msmgr.register_callbacks("START_WORKER_LIST", self.start_worker_list)
                                self.msmgr.register_callbacks("MONITOR", self.monitor_rank_status)
                                self.enable_mindx = True
                                os.environ["MS_ENABLE_RECOVERY"] = str(1)
                ...
                ```

    2.  拉起TaskD  Worker。
        -   **PyTorch-MindSpeed场景：**修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，在代码中增加如下加粗字段。

            ```
            def pretrain(train_valid_test_dataset_provider,
                         model_provider,
                         model_type,
                         forward_step_func,
                         process_non_loss_data_func=None,
                         extra_args_provider=None,
                         args_defaults={}):
                print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                    time.time() - _TRAIN_START_TIME))
                print_datetime('after megatron is initialized')
                import torch.distributed as dist
                if dist.is_initialized():
                   rank = dist.get_rank()
                   from taskd.api.taskd_worker_api import init_taskd_worker
                   from taskd.api.taskd_worker_api import start_taskd_worker
                   init_taskd_worker(rank,5000)
                   start_taskd_worker()
                app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
                one_logger_utils.on_pretrain_start()
            ```

        -   **MindSpore-MindFormers场景**：修改./mindformers/trainer/base\_trainer.py文件，在代码中增加如下加粗字段。

            ```
                def training_process(
                        self,
                        config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                        network: Optional[Union[Cell, PreTrainedModel]] = None,
                        dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                        optimizer: Optional[Optimizer] = None,
                        callbacks: Optional[Union[Callback, List[Callback]]] = None,
                        compute_metrics: Optional[Union[dict, set]] = None,
                        **kwargs):
                    ……
                    ……
            
                    logger.info(".........Starting Training Model..........")
                    if get_real_rank() % 8 == 0:
                        pprint(config)
                    logger.info(".........Model Compiling, Please Wait a Moment...........")
                    try:
                        rank = get_rank()
                        from taskd.api.taskd_worker_api import init_taskd_worker
                        from taskd.api.taskd_worker_api import start_taskd_worker
                        init_taskd_worker(rank,5000)
                        start_taskd_worker()
                    except Exception as e:
                        print("failed to call mindcluster taskd")
                    model.train(config.runner_config.epochs, dataset,
                                callbacks=callbacks,
                                dataset_sink_mode=config.runner_config.sink_mode,
                                sink_size=config.runner_config.sink_size,
                                initial_epoch=config.runner_config.initial_epoch)
            ```

5.  修改任务YAML。
    1.  修改容器暴露端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ports:                         
           - containerPort: 9601             
             name: taskd-port 
        ```

    2.  挂载文件。
        1.  挂载轻量profiling配置文件：需将宿主机上任务对应的data-trace ConfigMap落盘到/user/cluster-info/datatrace-config/命名空间.data-trace-任务名称/文件夹下。将名为profilingSwitch的文件挂载到容器指定路径：/user/cluster-info/datatrace-config/。
        2.  挂载轻量profiling落盘文件：轻量profiling数据写在容器内的/user/cluster-info/profiling路径下。如需在宿主机获取，请修改任务YAML，将该路径挂出。
            -   容器内YAML挂载示例如下。

                ```
                volumeMounts:
                - name: profilingdata
                  mountPath: /user/cluster-info/
                - name: profileswitch
                  mountPath: /user/cluster-info/datatrace-config
                ```

            -   宿主机内YAML挂载示例如下。

                ```
                volumes:
                - name: profileswitch
                  hostPath:
                    path: /user/cluster-info/datatrace-config/default.data-trace-default-test-pytorch-fault-mixtral
                - name: profilingdata
                  hostPath:
                     path: /home/profilingdatapath
                ```

6.  开启轻量profiling获取落盘数据。支持如下两种方式：
    -   修改ClusterD提供的gRPC接口：若配置了[4.a](#li399811541)，需要使用该方式开启。详细接口信息请参见[ModifyTrainingDataTraceSwitch](../api/clusterd.md#modifytrainingdatatraceswitch)。

        >[!NOTE] 说明 
        >通过ClusterD提供的gRPC接口开启或修改轻量profiling获取落盘数据，创建的data-trace-<任务名称\> ConfigMap的生命周期会随着任务的删除而删除。当任务不存在时，该接口会调用失败。

    -   修改任务对应的data-trace ConfigMap：若未配置[4.a](#li399811541)，需要使用该方式开启。具体操作步骤如下：

        以default命名空间下的名为_default-test-pytorch-fault-mixtral_的任务为例，以编辑ConfigMap的方式开启轻量profiling获取落盘数据，示例如下。

        1.  在master节点执行以下命令查询该任务对应的配置ConfigMap。

            ```
            kubectl get cm
            ```

            -   如果data-trace-default-test-pytorch-fault-mixtral cm已经存在，执行步骤[3](#zh-cn_topic_0000002194466236_li4751182133418)编辑该文件。

                回显示例如下。

                ```
                NAME                                              DATA   AGE
                data-trace-default-test-pytorch-fault-mixtral     1      18h
                ```

            -   如果data-trace-default-test-pytorch-fault-mixtral cm不存在，执行步骤[2](#zh-cn_topic_0000002194466236_li1633768104412)创建该文件。

        2.  <a name="zh-cn_topic_0000002194466236_li1633768104412"></a>执行以下命令，创建配置轻量profiling获取落盘数据所需ConfigMap文件。
            1.  将以下内容写入datacm.yaml。

                ```
                apiVersion: v1
                kind: ConfigMap
                metadata:
                  name: data-trace-default-test-pytorch-fault-mixtral  # cm的名字需以data-trace为前缀+任务名     
                  labels:
                    reset: "true"
                data:
                  profilingSwitch: '{"CommunicationOperator":"off","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
                ```

            2.  在master节点执行以下命令，创建ConfigMap。

                ```
                kubectl apply -f datacm.yaml
                ```

                回显如下所示，表示ConfigMap创建成功。

                ```
                [root@master~]# kubectl apply -f datacm.yaml 
                configmap/data-trace-default-test-pytorch-fault-mixtral created
                ```

        3.  <a name="zh-cn_topic_0000002194466236_li4751182133418"></a>执行以下命令编辑ConfigMap文件。

            ```
            kubectl edit cm data-trace-default-test-pytorch-fault-mixtral
            ```

        4.  如需开启通信算子，请将CommunicationOperator字段的取值改为“on”。

            ```
            apiVersion: v1
            data:
              profilingSwitch: '{"CommunicationOperator":"on","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
            ```

            >[!NOTE] 说明 
            >开启通信算子后可能造成训练性能下降，不建议常态开启通信算子。

        5.  按“Esc”键，输入:wq!保存并退出。

**获取性能劣化故障检测数据<a name="zh-cn_topic_0000002194466236_section435518556912"></a>**

-   落盘数据按rank进行分类，轻量profiling数据写在容器内的/user/cluster-info/profiling路径。
-   对于存在环境变量[MINDX\_TASK\_ID](../appendix.md#环境变量说明)的Pod，rank 0数据在容器内的路径为/user/cluster-info/profiling/$MINDX\_TASK\_ID/0。

    >[!NOTE] 说明 
    >-   如无该环境变量，默认会落盘到名为default\_task\_id\__时间戳_的文件夹内。
    >-   /user/cluster-info/profiling达到配置的上限大小（[def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool](../api/taskd.md#def-init_taskd_workerrank_id-int-upper_limit_of_disk_in_mb-int--5000-framework-str--pt---bool)中“upper\_limit\_of\_disk\_in\_mb”参数配置值）后，将进行文件老化，默认每次删除修改时间最早的20%个文件。老化过程中仅删除profiling目录下rank文件夹中的以数字命名的文件，建议不手动添加其他文件到profiling文件夹下。如果用户手动添加其他文件，TaskD不会将该文件删除，但该文件会占用空间。
    >-   轻量profiling文件以时间戳命名，各条记录以换行分割，每次追加写入rank下最新文件。最新文件大小超过10MB时，TaskD会新建profiling文件。如果使用NFS等网络存储方式，当数据同步较慢时，可能存在文件大小未达到10MB即创建新文件的情况。


##### 使用其他版本TaskD<a name="ZH-CN_TOPIC_0000002511346483"></a>

MindCluster集群调度组件结合MindStudio提供的profiling能力，对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态打点和打点数据持久化功能、可动态启停训练任务打点功能，无需重启任务进行诊断，对训练无损耗。

当前支持的打点数据如[表1](#zh-cn_topic_0000002194466236_table5530103025919)所示。

**表 1**  打点数据说明

<a name="zh-cn_topic_0000002194466236_table5530103025919"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002194466236_row105301830105911"><th class="cellrowborder" valign="top" width="21.12%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002194466236_p7530133018591"><a name="zh-cn_topic_0000002194466236_p7530133018591"></a><a name="zh-cn_topic_0000002194466236_p7530133018591"></a>打点数据的类型</p>
</th>
<th class="cellrowborder" valign="top" width="26.540000000000003%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002194466236_p165301308591"><a name="zh-cn_topic_0000002194466236_p165301308591"></a><a name="zh-cn_topic_0000002194466236_p165301308591"></a>支持的AI框架</p>
</th>
<th class="cellrowborder" valign="top" width="52.339999999999996%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002194466236_p153043055911"><a name="zh-cn_topic_0000002194466236_p153043055911"></a><a name="zh-cn_topic_0000002194466236_p153043055911"></a>提供支持的组件</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002194466236_row1753023011598"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p1692418239310"><a name="zh-cn_topic_0000002194466236_p1692418239310"></a><a name="zh-cn_topic_0000002194466236_p1692418239310"></a>FP</p>
<p id="zh-cn_topic_0000002194466236_p195301030165916"><a name="zh-cn_topic_0000002194466236_p195301030165916"></a><a name="zh-cn_topic_0000002194466236_p195301030165916"></a>（标识前向传播数据）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p1353016307593"><a name="zh-cn_topic_0000002194466236_p1353016307593"></a><a name="zh-cn_topic_0000002194466236_p1353016307593"></a>PyTorch</p>
<div class="note" id="zh-cn_topic_0000002194466236_note8765349888"><a name="zh-cn_topic_0000002194466236_note8765349888"></a><a name="zh-cn_topic_0000002194466236_note8765349888"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000002194466236_p1276511491787"><a name="zh-cn_topic_0000002194466236_p1276511491787"></a><a name="zh-cn_topic_0000002194466236_p1276511491787"></a>仅支持单算子场景。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002194466236_p5530630205914"><a name="zh-cn_topic_0000002194466236_p5530630205914"></a><a name="zh-cn_topic_0000002194466236_p5530630205914"></a>mstx_torch_plugin</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row13415176111"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p3788646727"><a name="zh-cn_topic_0000002194466236_p3788646727"></a><a name="zh-cn_topic_0000002194466236_p3788646727"></a>Step</p>
<p id="zh-cn_topic_0000002194466236_p2341175120"><a name="zh-cn_topic_0000002194466236_p2341175120"></a><a name="zh-cn_topic_0000002194466236_p2341175120"></a>（标识Step时延）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p834517411"><a name="zh-cn_topic_0000002194466236_p834517411"></a><a name="zh-cn_topic_0000002194466236_p834517411"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul17713185812916"></a><a name="zh-cn_topic_0000002194466236_ul17713185812916"></a><ul id="zh-cn_topic_0000002194466236_ul17713185812916"><li>PyTorch<a name="ul144271853105714"></a><a name="ul144271853105714"></a><ul id="ul144271853105714"><li>原生优化器场景：若torch_npu为7.1.RC1及以下版本，需使用mstx_torch_plugin；若torch_npu为7.1.RC1以上版本，无需使用mstx_torch_plugin，torch_npu自带Step打点。</li><li>自定义优化器场景：手动增加打点数据。</li></ul>
</li><li>MindSpore<a name="zh-cn_topic_0000002194466236_ul4814121617106"></a><a name="zh-cn_topic_0000002194466236_ul4814121617106"></a><ul id="zh-cn_topic_0000002194466236_ul4814121617106"><li>MindFormers场景：Step打点数据由MindFormers提供。</li><li>MindSpeed场景：不提供Step打点数据。</li></ul>
</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row9530630195919"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p169247495217"><a name="zh-cn_topic_0000002194466236_p169247495217"></a><a name="zh-cn_topic_0000002194466236_p169247495217"></a>Communication</p>
<p id="zh-cn_topic_0000002194466236_p1753012305591"><a name="zh-cn_topic_0000002194466236_p1753012305591"></a><a name="zh-cn_topic_0000002194466236_p1753012305591"></a>（标识通信算子）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p2401691783"><a name="zh-cn_topic_0000002194466236_p2401691783"></a><a name="zh-cn_topic_0000002194466236_p2401691783"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul18432034131111"></a><a name="zh-cn_topic_0000002194466236_ul18432034131111"></a><ul id="zh-cn_topic_0000002194466236_ul18432034131111"><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row953063010598"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p112455417213"><a name="zh-cn_topic_0000002194466236_p112455417213"></a><a name="zh-cn_topic_0000002194466236_p112455417213"></a>SaveCheckpoint</p>
<p id="zh-cn_topic_0000002194466236_p353043014592"><a name="zh-cn_topic_0000002194466236_p353043014592"></a><a name="zh-cn_topic_0000002194466236_p353043014592"></a>（标识SaveCheckpoint耗时）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p20176311281"><a name="zh-cn_topic_0000002194466236_p20176311281"></a><a name="zh-cn_topic_0000002194466236_p20176311281"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul1450711814124"></a><a name="zh-cn_topic_0000002194466236_ul1450711814124"></a><ul id="zh-cn_topic_0000002194466236_ul1450711814124"><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002194466236_row1234614241805"><td class="cellrowborder" valign="top" width="21.12%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002194466236_p1271312516318"><a name="zh-cn_topic_0000002194466236_p1271312516318"></a><a name="zh-cn_topic_0000002194466236_p1271312516318"></a>DataLoader</p>
<p id="zh-cn_topic_0000002194466236_p834615241902"><a name="zh-cn_topic_0000002194466236_p834615241902"></a><a name="zh-cn_topic_0000002194466236_p834615241902"></a>（标识DataLoader耗时）</p>
</td>
<td class="cellrowborder" valign="top" width="26.540000000000003%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002194466236_p459471310811"><a name="zh-cn_topic_0000002194466236_p459471310811"></a><a name="zh-cn_topic_0000002194466236_p459471310811"></a>PyTorch、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="52.339999999999996%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002194466236_ul827633418123"></a><a name="zh-cn_topic_0000002194466236_ul827633418123"></a><ul id="zh-cn_topic_0000002194466236_ul827633418123"><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架。</li></ul>
</td>
</tr>
</tbody>
</table>

**使用约束<a name="zh-cn_topic_0000002194466236_section487603614"></a>**

-   当前Step、SaveCheckpoint、FP、DataLoader仅支持同步开启。如需关闭以上四类打点数据，需同时关闭Communication。
-   Communication通信算子数据支持单独开启、关闭。
-   动态轻量打点功能与MindStudio的全量打点功能不可同时开启，开启全量打点功能会导致性能劣化故障不能正常采集数据。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   （可选）已安装[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装[torch\_npu](#制作mindspeed-llm训练镜像pytorch框架)（**可选**，PyTorch场景需安装、版本号≥7.0.0）、MindSpore（**可选**，MindSpore场景需安装、版本号≥2.6.RC1）、[CANN](#制作mindformers训练镜像mindspore框架)（**必选**，版本号≥8.1.RC1）、[TaskD](#制作mindformers训练镜像mindspore框架)（**必选**，版本号≥7.0.RC1）

**准备软件包<a name="zh-cn_topic_0000002194466236_section8281518121516"></a>**

**表 2**  准备软件包

<a name="zh-cn_topic_0000002194466236_table232305471415"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002194466236_row33231354121420"><th class="cellrowborder" valign="top" width="14.08%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000002194466236_p1441653254"><a name="zh-cn_topic_0000002194466236_p1441653254"></a><a name="zh-cn_topic_0000002194466236_p1441653254"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="6.74%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000002194466236_p2052053751"><a name="zh-cn_topic_0000002194466236_p2052053751"></a><a name="zh-cn_topic_0000002194466236_p2052053751"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="60.75000000000001%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000002194466236_p657531455"><a name="zh-cn_topic_0000002194466236_p657531455"></a><a name="zh-cn_topic_0000002194466236_p657531455"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="8.58%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000002194466236_p1859531759"><a name="zh-cn_topic_0000002194466236_p1859531759"></a><a name="zh-cn_topic_0000002194466236_p1859531759"></a>获取方法</p>
</th>
<th class="cellrowborder" valign="top" width="9.85%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000002194466236_p1444313132617"><a name="zh-cn_topic_0000002194466236_p1444313132617"></a><a name="zh-cn_topic_0000002194466236_p1444313132617"></a>使用场景</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002194466236_row1715722244511"><td class="cellrowborder" valign="top" width="14.08%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000002194466236_p19157182244515"><a name="zh-cn_topic_0000002194466236_p19157182244515"></a><a name="zh-cn_topic_0000002194466236_p19157182244515"></a>mstx_torch_plugin</p>
</td>
<td class="cellrowborder" valign="top" width="6.74%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000002194466236_p31576228454"><a name="zh-cn_topic_0000002194466236_p31576228454"></a><a name="zh-cn_topic_0000002194466236_p31576228454"></a>否</p>
</td>
<td class="cellrowborder" valign="top" width="60.75000000000001%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000002194466236_p15157192234518"><a name="zh-cn_topic_0000002194466236_p15157192234518"></a><a name="zh-cn_topic_0000002194466236_p15157192234518"></a><span>Ascend PyTorch Profiler中的</span><a href="https://www.hiascend.com/document/detail/zh/canncommercial/800/devaids/devtools/profiling/atlasprofiling_16_0033.html" target="_blank" rel="noopener noreferrer">采集并解析msproftx数据</a><span>功能已经内置了通信算子的打点。为了方便用户在不修改业务代码的基础上获取更多关键阶段的耗时数据，mstx_torch_plugin在Ascend PyTorch Profiler内置了</span><span>dataloader</span><span>、</span><span>forward</span><span>、</span><span>step</span><span>、</span><span>save_checkpoint</span><span>这四个关键阶段函数的打点。</span></p>
<div class="note" id="zh-cn_topic_0000002194466236_note179451154301"><a name="zh-cn_topic_0000002194466236_note179451154301"></a><a name="zh-cn_topic_0000002194466236_note179451154301"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="zh-cn_topic_0000002194466236_ul17791934115613"></a><a name="zh-cn_topic_0000002194466236_ul17791934115613"></a><ul id="zh-cn_topic_0000002194466236_ul17791934115613"><li>如需使用FP打点数据，需安装mstx_torch_plugin。其他场景下无需安装。</li><li>需使用1.0及以上版本的mstx_torch_plugin。</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="8.58%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000002194466236_p5157162210453"><a name="zh-cn_topic_0000002194466236_p5157162210453"></a><a name="zh-cn_topic_0000002194466236_p5157162210453"></a><a href="https://gitee.com/link?target=https://ptdbg.obs.myhuaweicloud.com/profiler/example/1.0/mstx_torch_plugin-1.0-py3-none-any.whl" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
<td class="cellrowborder" valign="top" width="9.85%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000002194466236_p1244314111268"><a name="zh-cn_topic_0000002194466236_p1244314111268"></a><a name="zh-cn_topic_0000002194466236_p1244314111268"></a>PyTorch</p>
</td>
</tr>
</tbody>
</table>

**配置性能劣化故障检测<a name="section167141313174510"></a>**

本方案仅针对7.1.RC1以下版本的TaskD组件。如使用7.1.RC1及以上版本的组件请参见[使用7.1.RC1及以上版本TaskD](#使用71rc1及以上版本taskd)章节。

1.  （可选）在容器内安装mstx\_torch\_plugin。
    1.  下载mstx\_torch\_plugin的whl包。whl包链接：[mstx\_torch\_plugin](https://gitee.com/link?target=https%3A%2F%2Fptdbg.obs.myhuaweicloud.com%2Fprofiler%2Fexample%2F1.0%2Fmstx_torch_plugin-1.0-py3-none-any.whl)。
    2.  安装软件包。

        ```
        pip install mstx_torch_plugin-1.0-py3-none-any.whl
        ```

    3.  在AI任务执行脚本中import导入该whl包。

        需保证import的顺序在import torch和import torch\_npu后面，示例如下。

        ```
        import torch 
        import torch_npu  
        import mstx_torch_plugin
        ```

2.  （可选）在PyTorch场景下，非原生优化器或不使用mstx\_torch\_plugin的情况下，为获取训练的Step耗时数据需修改训练脚本中的训练迭代循环，需增加Step打点代码。

    以下示例为PyTorch-MindSpeed场景，需修改./mindspeed\_llm/training/training.py文件增加如下加粗字段。MindSpore场景请根据实际情况修改。

    ```
    def train(forward_step_func, model, optimizer, opt_param_scheduler,
              train_data_iterator, valid_data_iterator,
              process_non_loss_data_func, config):
                # Cache into one-logger for callback
        ……
        ……
        if is_profile_enabled():
            prof = get_profiler()
            prof.start()
        step_id = iteration
        while iteration < args.train_iters:
            stream = torch.npu.current_stream()      # 获取当前环境的执行流，用于获取NPU侧时间
            range_id = torch.npu.mstx.range_start(f"step {step_id}", stream) # 标识当前训练step的开始
            ……
            ……
            if args.manual_gc:
                if args.manual_gc_interval != 0 and iteration % args.manual_gc_interval == 0:
                    gc.collect()
    
            if is_profile_enabled():
                prof.step()
            step_id +=1  # 训练step加一，用于标识下一step
            torch.npu.mstx.range_end(range_id) # 标识当前训练step的结束
    ```

3.  在容器内，以CANN软件包的运行用户登录环境，执行**source $\{install\_path\}/set\_env.sh**命令设置环境变量。其中$\{install\_path\}为CANN软件的安装目录。示例如下。

    ```
    source /usr/local/Ascend/cann/set_env.sh
    ```

4.  训练启动前，在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。

    ```
    export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    ```

    -   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。

    -   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：  TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。

        TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。

        ```
        pip show taskd
        ```

5.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练进程内部拉起TaskD  Worker。
    -   **PyTorch-MindSpeed场景**

        修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，在代码中增加如下加粗字段。

        ```
        def pretrain(train_valid_test_dataset_provider,
                     model_provider,
                     model_type,
                     forward_step_func,
                     process_non_loss_data_func=None,
                     extra_args_provider=None,
                     args_defaults={}):
            print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                time.time() - _TRAIN_START_TIME))
            print_datetime('after megatron is initialized')
            import torch.distributed as dist
            if dist.is_initialized():
               rank = dist.get_rank()
               from taskd.api.taskd_worker_api import init_taskd_worker
               from taskd.api.taskd_worker_api import start_taskd_worker
               init_taskd_worker(rank,5000)
               start_taskd_worker()
            app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
            one_logger_utils.on_pretrain_start()
        ```

    -   **MindSpore-MindFormers场景**

        修改./mindformers/trainer/base\_trainer.py文件，在代码中增加如下加粗字段。

        ```
            def training_process(
                    self,
                    config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                    network: Optional[Union[Cell, PreTrainedModel]] = None,
                    dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                    optimizer: Optional[Optimizer] = None,
                    callbacks: Optional[Union[Callback, List[Callback]]] = None,
                    compute_metrics: Optional[Union[dict, set]] = None,
                    **kwargs):
                ……
                ……
        
                logger.info(".........Starting Training Model..........")
                if get_real_rank() % 8 == 0:
                    pprint(config)
                logger.info(".........Model Compiling, Please Wait a Moment...........")
                try:
                    rank = get_rank()
                    from taskd.api.taskd_worker_api import init_taskd_worker
                    from taskd.api.taskd_worker_api import start_taskd_worker
                    init_taskd_worker(rank,5000)
                    start_taskd_worker()
                except Exception as e:
                    print("failed to call mindcluster taskd")
                model.train(config.runner_config.epochs, dataset,
                            callbacks=callbacks,
                            dataset_sink_mode=config.runner_config.sink_mode,
                            sink_size=config.runner_config.sink_size,
                            initial_epoch=config.runner_config.initial_epoch)
        ```

6.  修改任务YAML。
    1.  挂载轻量profiling配置文件：需将宿主机上任务对应的data-trace ConfigMap落盘到/user/cluster-info/datatrace-config/命名空间.data-trace-任务名称/文件夹下。将名为profilingSwitch的文件挂载到容器指定路径：/user/cluster-info/datatrace-config/。
    2.  挂载轻量profiling落盘文件：轻量profiling数据写在容器内的/user/cluster-info/profiling路径下。如需在宿主机获取，请修改任务YAML，将该路径挂出。
        -   容器内YAML挂载示例如下。

            ```
            volumeMounts:
            - name: profilingdata
              mountPath: /user/cluster-info/
            - name: profileswitch
              mountPath: /user/cluster-info/datatrace-config
            ```

        -   宿主机内YAML挂载示例如下。

            ```
            volumes:
            - name: profileswitch
              hostPath:
                path: /user/cluster-info/datatrace-config/default.data-trace-default-test-pytorch-fault-mixtral
            - name: profilingdata
              hostPath:
                 path: /home/profilingdatapath
            ```

7.  开启轻量profiling获取落盘数据。修改任务对应的data-trace ConfigMap或ClusterD提供的gRPC接口，接口信息见[ModifyTrainingDataTraceSwitch](../api/clusterd.md#modifytrainingdatatraceswitch)，动态开启或关闭轻量profiling能力。

    以default命名空间下的名为_default-test-pytorch-fault-mixtral_的任务为例，以编辑ConfigMap的方式开启轻量profiling获取落盘数据，示例如下。

    1.  在master节点执行以下命令查询该任务对应的配置ConfigMap。

        ```
        kubectl get cm
        ```

        -   如果data-trace-default-test-pytorch-fault-mixtral cm已经存在，执行步骤[3](#zh-cn_topic_0000002194466236_li4751182133418)编辑该文件。

            回显示例如下。

            ```
            NAME                                              DATA   AGE
            data-trace-default-test-pytorch-fault-mixtral     1      18h
            ```

        -   如果data-trace-default-test-pytorch-fault-mixtral cm不存在，执行步骤[2](#zh-cn_topic_0000002194466236_li1633768104412)创建该文件。

    2.  <a name="zh-cn_topic_0000002194466236_li1633768104412"></a>执行以下命令，创建配置轻量profiling获取落盘数据所需ConfigMap文件。
        1.  将以下内容写入datacm.yaml。

            ```
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: data-trace-default-test-pytorch-fault-mixtral  # cm的名字需以data-trace为前缀+任务名     
              labels:
                reset: "true"
            data:
              profilingSwitch: '{"CommunicationOperator":"off","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
            ```

        2.  在master节点执行以下命令，创建ConfigMap。

            ```
            kubectl apply -f datacm.yaml
            ```

            回显如下所示，表示ConfigMap创建成功。

            ```
            [root@master~]# kubectl apply -f datacm.yaml 
            configmap/data-trace-default-test-pytorch-fault-mixtral created
            ```

    3.  <a name="zh-cn_topic_0000002194466236_li4751182133418"></a>执行以下命令编辑ConfigMap文件。

        ```
        kubectl edit cm data-trace-default-test-pytorch-fault-mixtral
        ```

    4.  如需开启通信算子，请将CommunicationOperator字段的取值改为“on”。

        ```
        apiVersion: v1
        data:
          profilingSwitch: '{"CommunicationOperator":"on","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
        ```

        >[!NOTE] 说明 
        >开启通信算子后可能造成训练性能下降，不建议常态开启通信算子。

    5.  按“Esc”键，输入:wq!保存并退出。

**获取性能劣化故障检测数据<a name="zh-cn_topic_0000002194466236_section435518556912"></a>**

-   落盘数据按rank进行分类，轻量profiling数据写在容器内的/user/cluster-info/profiling路径。
-   对于存在环境变量[MINDX\_TASK\_ID](../appendix.md#环境变量说明)的Pod，rank 0数据在容器内的路径为/user/cluster-info/profiling/$MINDX\_TASK\_ID/0。

    >[!NOTE] 说明 
    >-   如无该环境变量，默认会落盘到名为default\_task\_id\__时间戳_的文件夹内。
    >-   /user/cluster-info/profiling达到配置的上限大小（[def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool](../api/taskd.md#def-init_taskd_workerrank_id-int-upper_limit_of_disk_in_mb-int--5000-framework-str--pt---bool)中“upper\_limit\_of\_disk\_in\_mb”参数配置值）后，将进行文件老化，默认每次删除修改时间最早的20%个文件。老化过程中仅删除profiling目录下rank文件夹中的以数字命名的文件，建议不手动添加其他文件到profiling文件夹下。如果用户手动添加其他文件，TaskD不会将该文件删除，但该文件会占用空间。
    >-   轻量profiling文件以时间戳命名，各条记录以换行分割，每次追加写入rank下最新文件。最新文件大小超过10MB时，TaskD会新建profiling文件。如果使用NFS等网络存储方式，当数据同步较慢时，可能存在文件大小未达到10MB即创建新文件的情况。



#### 慢节点&慢网络故障<a name="ZH-CN_TOPIC_0000002511426421"></a>

##### 简介<a name="ZH-CN_TOPIC_0000002532640773"></a>

MindCluster集群调度组件结合MindCluster Ascend FaultDiag（故障诊断工具）提供的在线诊断能力，为集群中的慢节点&慢网络故障提供诊断功能。

**使用前准备<a name="zh-cn_topic_0000002333550505_section420815439315"></a>**

使用慢节点&慢网络故障诊断功能前，需增加NodeD中CPU和内存的资源大小，在NodeD启动YAML文件中更改资源信息。

当前YAML文件内容如下：

```
resources:
            requests:
              memory: 300Mi
              cpu: 500m
            limits:
              memory: 300Mi
              cpu: 500m
```

修改后YAML文件内容如下：

```
resources:
            requests:
              memory: 10Gi
              cpu: 5000m
            limits:
              memory: 10Gi
              cpu: 5000m
```

**部署形态<a name="zh-cn_topic_0000002333550505_section1048011118418"></a>**

ClusterD与FD-OL（Fault Diagnose Online）框架在同一进程中，都部署在管理节点。ClusterD启动时将自动拉起FD-OL框架。


##### 慢节点诊断<a name="ZH-CN_TOPIC_0000002500880704"></a>

**功能说明<a name="zh-cn_topic_0000002278667326_section27999216294"></a>**

对于AI集群中出现的节点训练性能劣化现象，提供支持实时检测计算域问题或网络导致的慢节点，以便用户通过切换或其他方式隔离慢节点。

当前仅支持与ClusterD和NodeD集成进行在线部署，请参见[安装部署](../installation_guide.md#安装部署)章节完成ClusterD和NodeD部署。

-   慢节点算法：基于训练场景关键性能指标，感知实时劣化状态；针对通信算子、计算算子同步关系，实现慢计算卡、慢通信域问题定界。
-   慢节点清洗：对节点内部增量数据转化并清洗，生成清洗结果csv文件。
-   慢节点调度：调度慢节点整体流程，控制数据清洗和慢节点算法。

**使用示例<a name="zh-cn_topic_0000002278667326_section19867823600"></a>**

启动慢节点诊断任务。

1.  为获取并行域信息，需在训练脚本的训练迭代循环中增加获取并行域信息的函数调用。以下示例为PyTorch-MindSpeed场景，需在./mindspeed\_llm/training/training.py文件增加如下加粗字段。

    ```
    def train(forward_step_func, model, optimizer, opt_param_scheduler,
              train_data_iterator, valid_data_iterator,
              process_non_loss_data_func, config):
        ……
        if is_profile_enabled():
            prof = get_profiler()
            prof.start()
        m_iter = 0
        while iteration < args.train_iters:
            ……
            args.curr_iteration = iteration
            loss_dict, skipped_iter, grad_norm, num_zeros_in_grad = \
                train_step(forward_step_func,
                           train_data_iterator,
                           model,
                           optimizer,
                           opt_param_scheduler,
                           config)
            iteration += 1
            m_iter += 1
            if m_iter == 5:
                from taskd.python.adaptor.pytorch.group_info import dump_group_info
                dump_group_info()
            batch_size = mpu.get_data_parallel_world_size() * \
                         args.micro_batch_size * \
                         get_num_microbatches()
    ```

2.  完成[使用前准备](#zh-cn_topic_0000002333550505_section420815439315)和[部署形态](#zh-cn_topic_0000002333550505_section1048011118418)。
3.  使用**kubectl apply -f ajob-2pod-16npu.yaml**命令，创建慢节点诊断任务写入configMap。

    ![](../../figures/scheduling/zh-cn_image_0000002333860285.png)

4.  ajob-2pod-16npu.yaml内容如下所示，各回显数据说明请见[表1](#zh-cn_topic_0000002278667326_table1834456175114)。

    ![](../../figures/scheduling/zh-cn_image_0000002509443757.png)

    以下为YAML示例，不可以直接拷贝编译运行，仅供参考。

    ```
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ras-feature-slownode-default-test-pytorch-2pod-16npu    # The value of JobName must be the same as the name attribute of the following job. The prefix ras-feature-slownode- cannot be modified.
      namespace: mindx-dl
      labels:
        fd-ol-slow-node: "true"
    data:
      FeatConf: |
        {"jobName":"default-test-pytorch-2pod-16npu","jobNamespace":"default","normalNumber":20,"nSigma":3,"degradationPercentage":0.3,"nConsecAnomaliesSignifySlow":3,"nSecondsDoOneDetection":30,"clusterMeanDistance":1.3,"cardOneNode":16,"SlowNode":1}
    ---
    ```

    **表 1**  YAML文件回显说明

    <a name="zh-cn_topic_0000002278667326_table1834456175114"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002278667326_row53355612518"><th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002278667326_p173385615518"><a name="zh-cn_topic_0000002278667326_p173385615518"></a><a name="zh-cn_topic_0000002278667326_p173385615518"></a>字段名</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002278667326_p3331256105116"><a name="zh-cn_topic_0000002278667326_p3331256105116"></a><a name="zh-cn_topic_0000002278667326_p3331256105116"></a>默认值</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002278667326_p533256115114"><a name="zh-cn_topic_0000002278667326_p533256115114"></a><a name="zh-cn_topic_0000002278667326_p533256115114"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002278667326_row103315695115"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p9335560512"><a name="zh-cn_topic_0000002278667326_p9335560512"></a><a name="zh-cn_topic_0000002278667326_p9335560512"></a>jobNamespace</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p933205615511"><a name="zh-cn_topic_0000002278667326_p933205615511"></a><a name="zh-cn_topic_0000002278667326_p933205615511"></a>default</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p9331156155112"><a name="zh-cn_topic_0000002278667326_p9331156155112"></a><a name="zh-cn_topic_0000002278667326_p9331156155112"></a>任务所在的namespace。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row183318567511"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p83345616518"><a name="zh-cn_topic_0000002278667326_p83345616518"></a><a name="zh-cn_topic_0000002278667326_p83345616518"></a>jobName</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p7331456165110"><a name="zh-cn_topic_0000002278667326_p7331456165110"></a><a name="zh-cn_topic_0000002278667326_p7331456165110"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p13318567516"><a name="zh-cn_topic_0000002278667326_p13318567516"></a><a name="zh-cn_topic_0000002278667326_p13318567516"></a>任务名。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row123375613510"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p1933185614515"><a name="zh-cn_topic_0000002278667326_p1933185614515"></a><a name="zh-cn_topic_0000002278667326_p1933185614515"></a>normalNumber</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p733175616519"><a name="zh-cn_topic_0000002278667326_p733175616519"></a><a name="zh-cn_topic_0000002278667326_p733175616519"></a>20</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p113365613511"><a name="zh-cn_topic_0000002278667326_p113365613511"></a><a name="zh-cn_topic_0000002278667326_p113365613511"></a>计算初始阈值（正常数量）。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row1633135619514"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p43325618514"><a name="zh-cn_topic_0000002278667326_p43325618514"></a><a name="zh-cn_topic_0000002278667326_p43325618514"></a>nSigma</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p933155635114"><a name="zh-cn_topic_0000002278667326_p933155635114"></a><a name="zh-cn_topic_0000002278667326_p933155635114"></a>3个</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p1433175635115"><a name="zh-cn_topic_0000002278667326_p1433175635115"></a><a name="zh-cn_topic_0000002278667326_p1433175635115"></a>设置σ的个数以计算其上下界。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row1033155665119"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p23325625112"><a name="zh-cn_topic_0000002278667326_p23325625112"></a><a name="zh-cn_topic_0000002278667326_p23325625112"></a>degradationPercentage</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p23319569517"><a name="zh-cn_topic_0000002278667326_p23319569517"></a><a name="zh-cn_topic_0000002278667326_p23319569517"></a>0.3</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p133156155116"><a name="zh-cn_topic_0000002278667326_p133156155116"></a><a name="zh-cn_topic_0000002278667326_p133156155116"></a>阈值，劣化的百分比，0.3表示劣化30%。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row11331956135114"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p933155695120"><a name="zh-cn_topic_0000002278667326_p933155695120"></a><a name="zh-cn_topic_0000002278667326_p933155695120"></a>nConsecAnomaliesSignifySlow</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p93305614511"><a name="zh-cn_topic_0000002278667326_p93305614511"></a><a name="zh-cn_topic_0000002278667326_p93305614511"></a>3次</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p733165645112"><a name="zh-cn_topic_0000002278667326_p733165645112"></a><a name="zh-cn_topic_0000002278667326_p733165645112"></a>设置异常次数，连续出现多次异常后才进行检测。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row933135610517"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p113325655119"><a name="zh-cn_topic_0000002278667326_p113325655119"></a><a name="zh-cn_topic_0000002278667326_p113325655119"></a>nSecondsDoOneDetection</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p83325685114"><a name="zh-cn_topic_0000002278667326_p83325685114"></a><a name="zh-cn_topic_0000002278667326_p83325685114"></a>30秒</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p18330569519"><a name="zh-cn_topic_0000002278667326_p18330569519"></a><a name="zh-cn_topic_0000002278667326_p18330569519"></a>设置间隔时长，进行检测，单位为秒。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row733155614517"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p143335655111"><a name="zh-cn_topic_0000002278667326_p143335655111"></a><a name="zh-cn_topic_0000002278667326_p143335655111"></a>clusterMeanDistance</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p1633125613516"><a name="zh-cn_topic_0000002278667326_p1633125613516"></a><a name="zh-cn_topic_0000002278667326_p1633125613516"></a>1.3</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p1433155612517"><a name="zh-cn_topic_0000002278667326_p1433155612517"></a><a name="zh-cn_topic_0000002278667326_p1433155612517"></a>聚类后，两个类别之间的阈值距离（mean1、mean2）。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row7341056115120"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p433185612512"><a name="zh-cn_topic_0000002278667326_p433185612512"></a><a name="zh-cn_topic_0000002278667326_p433185612512"></a>cardOneNode</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p1433125655117"><a name="zh-cn_topic_0000002278667326_p1433125655117"></a><a name="zh-cn_topic_0000002278667326_p1433125655117"></a>16张卡</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p1433145625116"><a name="zh-cn_topic_0000002278667326_p1433145625116"></a><a name="zh-cn_topic_0000002278667326_p1433145625116"></a>一个节点的卡片数量。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002278667326_row534105613515"><td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002278667326_p1734125613512"><a name="zh-cn_topic_0000002278667326_p1734125613512"></a><a name="zh-cn_topic_0000002278667326_p1734125613512"></a>slowNode</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002278667326_p034256155110"><a name="zh-cn_topic_0000002278667326_p034256155110"></a><a name="zh-cn_topic_0000002278667326_p034256155110"></a>默认为1，开启任务。</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002278667326_p1534175616518"><a name="zh-cn_topic_0000002278667326_p1534175616518"></a><a name="zh-cn_topic_0000002278667326_p1534175616518"></a>是否开启任务。</p>
    <a name="zh-cn_topic_0000002278667326_ul11341756165115"></a><a name="zh-cn_topic_0000002278667326_ul11341756165115"></a><ul id="zh-cn_topic_0000002278667326_ul11341756165115"><li>1：开启任务。</li><li>0：关闭任务。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

**查询慢节点诊断结果<a name="zh-cn_topic_0000002278667326_section208199121010"></a>**

在创建慢节点任务后，可通过查询ClusterD和NodeD的日志查看其诊断任务详情。

**方式一：通过K8s日志查询集群侧慢节点诊断日志**

1.  通过**kubectl get pods -n mindx-dl**命令，查询启动的ClusterD和NodeD节点数据。

    ![](../../figures/scheduling/zh-cn_image_0000002477523808.png)

2.  再使用**kubectl logs -n mindx-dl clusterd-7d5db546d8-kdslz | grep "got degradation, slow rank"**查询日志数据。
3.  若日志中出现如下图所示，则表明出现节点劣化。

    ![](../../figures/scheduling/zh-cn_image_0000002457147010.png)

**方式二：通过落盘日志查询集群侧慢节点诊断日志**

1.  使用**cat /var/log/mindx-dl.clusterd.clusterd.log | grep "got degradation, slow rank"**命令查询日志数据。
2.  若日志中出现如下图所示，则表明出现节点劣化。

    ![](../../figures/scheduling/zh-cn_image_0000002490267057.png)

**方式三：查询节点侧的慢节点诊断日志。**

使用**kubectl logs -n mindx-dl node-9ld8k | grep "is degradation"**命令进行查询，若日志中出现如下图所示数据，则表明出现节点劣化。

![](../../figures/scheduling/zh-cn_image_0000002457149146.png)

**已支持的慢节点网络故障<a name="zh-cn_topic_0000002278667326_section10496211245"></a>**

<a name="zh-cn_topic_0000002278667326_table4804164084414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002278667326_row1680414018449"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="zh-cn_topic_0000002278667326_p1680411405446"><a name="zh-cn_topic_0000002278667326_p1680411405446"></a><a name="zh-cn_topic_0000002278667326_p1680411405446"></a>故障码</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="zh-cn_topic_0000002278667326_p280464074412"><a name="zh-cn_topic_0000002278667326_p280464074412"></a><a name="zh-cn_topic_0000002278667326_p280464074412"></a>故障说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="zh-cn_topic_0000002278667326_p3804114018440"><a name="zh-cn_topic_0000002278667326_p3804114018440"></a><a name="zh-cn_topic_0000002278667326_p3804114018440"></a>故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002278667326_row1080414409444"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0000002278667326_p17804114034413"><a name="zh-cn_topic_0000002278667326_p17804114034413"></a><a name="zh-cn_topic_0000002278667326_p17804114034413"></a>110001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0000002278667326_p1780424011445"><a name="zh-cn_topic_0000002278667326_p1780424011445"></a><a name="zh-cn_topic_0000002278667326_p1780424011445"></a>慢节点故障，一次性消息上报。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0000002278667326_p15804114074414"><a name="zh-cn_topic_0000002278667326_p15804114074414"></a><a name="zh-cn_topic_0000002278667326_p15804114074414"></a>SubHealthFault：亚健康故障。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002278667326_row4804140134413"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0000002278667326_p11804340204417"><a name="zh-cn_topic_0000002278667326_p11804340204417"></a><a name="zh-cn_topic_0000002278667326_p11804340204417"></a>100001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0000002278667326_p5804154074418"><a name="zh-cn_topic_0000002278667326_p5804154074418"></a><a name="zh-cn_topic_0000002278667326_p5804154074418"></a>故障劣化已恢复。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0000002278667326_p171970401993"><a name="zh-cn_topic_0000002278667326_p171970401993"></a><a name="zh-cn_topic_0000002278667326_p171970401993"></a>NotHandleFault：暂不处理故障。</p>
</td>
</tr>
</tbody>
</table>


##### 慢网络诊断<a name="ZH-CN_TOPIC_0000002500720860"></a>

**功能说明<a name="zh-cn_topic_0000002313236861_section27999216294"></a>**

支持提供参数面网络连通性检测，实时进行网络监测和异常上报，辅助故障分析和定界定位，提前预警网络故障和亚健康风险信息，保障集群网络的长期稳定运行。

当前仅支持与ClusterD和NodeD集成进行在线部署，请参见[安装部署](../installation_guide.md#安装部署)章节完成ClusterD和NodeD部署。

-   慢网络算法：对节点之间的网络拨测数据进行分析、检测，并输出网络诊断结果。
-   慢网络调度：把控探测任务启停，上报故障结果，调度慢网络整体流程。

**使用示例<a name="zh-cn_topic_0000002313236861_section1969604665710"></a>**

1.  配置共享存储。

    ClusterD和NodeD通过共享存储进行交互，两者的共享存储根路径需要保持一致。共享目录的根路径属主为9000用户，与ClusterD运行用户一致。

    1.  配置server。

        ![](../../figures/scheduling/zh-cn_image_0000002300566136.png)

    2.  修改NodeD配置。

        ![](../../figures/scheduling/zh-cn_image_0000002384880596.png)

    3.  修改ClusterD配置。

        ![](../../figures/scheduling/zh-cn_image_0000002385041140.png)

    4.  执行**kubectl get pods -o wide -A**命令出现如下示例，则表示已完成共享存储配置。

        ![](../../figures/scheduling/zh-cn_image_0000002300409300.png)

2.  开启故障检测开关。
    1.  登录环境，进入NodeD解压目录。
    2.  执行以下命令创建名为pingmesh-config的ConfigMap文件。pingmesh-config.yaml为pingmesh配置文件，可从NodeD安装包中获取。

        ```
        kubectl apply -f pingmesh-config.yaml
        ```

        回显示例如下：

        ```
        configmap/pingmesh-config created
        ```

    3.  执行以下命令编辑pingmesh-config文件，该文件中各参数的填写说明如下表所示。

        ```
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

        **表 1**  pingmesh-config文件参数说明

        <a name="zh-cn_topic_0000002313236861_table15591134151811"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002313236861_row8591133431815"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002313236861_p10591534121814"><a name="zh-cn_topic_0000002313236861_p10591534121814"></a><a name="zh-cn_topic_0000002313236861_p10591534121814"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="33.3033303330333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002313236861_p185915343188"><a name="zh-cn_topic_0000002313236861_p185915343188"></a><a name="zh-cn_topic_0000002313236861_p185915343188"></a>取值</p>
        </th>
        <th class="cellrowborder" valign="top" width="33.36333633363336%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002313236861_p16591934171818"><a name="zh-cn_topic_0000002313236861_p16591934171818"></a><a name="zh-cn_topic_0000002313236861_p16591934171818"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002313236861_row759112347182"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p125919340187"><a name="zh-cn_topic_0000002313236861_p125919340187"></a><a name="zh-cn_topic_0000002313236861_p125919340187"></a>app</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.3033303330333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p459193419186"><a name="zh-cn_topic_0000002313236861_p459193419186"></a><a name="zh-cn_topic_0000002313236861_p459193419186"></a>pingmesh</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.36333633363336%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p3591153418181"><a name="zh-cn_topic_0000002313236861_p3591153418181"></a><a name="zh-cn_topic_0000002313236861_p3591153418181"></a>ConfigMap其中一个label的key。</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002313236861_row8591183414183"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p185912034201811"><a name="zh-cn_topic_0000002313236861_p185912034201811"></a><a name="zh-cn_topic_0000002313236861_p185912034201811"></a>global</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.3033303330333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p9591434121818"><a name="zh-cn_topic_0000002313236861_p9591434121818"></a><a name="zh-cn_topic_0000002313236861_p9591434121818"></a>-</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.36333633363336%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p1159143411186"><a name="zh-cn_topic_0000002313236861_p1159143411186"></a><a name="zh-cn_topic_0000002313236861_p1159143411186"></a>集群配置信息。</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002313236861_row2059133481816"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p13591193461817"><a name="zh-cn_topic_0000002313236861_p13591193461817"></a><a name="zh-cn_topic_0000002313236861_p13591193461817"></a>"1"</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.3033303330333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p459113412187"><a name="zh-cn_topic_0000002313236861_p459113412187"></a><a name="zh-cn_topic_0000002313236861_p459113412187"></a>超节点ID</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.36333633363336%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p1159120349189"><a name="zh-cn_topic_0000002313236861_p1159120349189"></a><a name="zh-cn_topic_0000002313236861_p1159120349189"></a>超节点ID为1的配置示例，用户可根据实际情况进行修改或新增。当配置了某个超节点后，NodeD会采用超节点的配置信息而忽略global配置信息。</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002313236861_row11955391011"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p149125351019"><a name="zh-cn_topic_0000002313236861_p149125351019"></a><a name="zh-cn_topic_0000002313236861_p149125351019"></a>activate</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.3033303330333%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002313236861_ul16388568116"></a><a name="zh-cn_topic_0000002313236861_ul16388568116"></a><ul id="zh-cn_topic_0000002313236861_ul16388568116"><li>on：开启</li><li>off：关闭</li></ul>
        </td>
        <td class="cellrowborder" valign="top" width="33.36333633363336%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p61005311016"><a name="zh-cn_topic_0000002313236861_p61005311016"></a><a name="zh-cn_topic_0000002313236861_p61005311016"></a>是否启用pingmesh功能。</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002313236861_row175911434111820"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p6591193431816"><a name="zh-cn_topic_0000002313236861_p6591193431816"></a><a name="zh-cn_topic_0000002313236861_p6591193431816"></a>task_interval</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.3033303330333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p1259111345187"><a name="zh-cn_topic_0000002313236861_p1259111345187"></a><a name="zh-cn_topic_0000002313236861_p1259111345187"></a>[1~60]</p>
        </td>
        <td class="cellrowborder" valign="top" width="33.36333633363336%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p5591634201811"><a name="zh-cn_topic_0000002313236861_p5591634201811"></a><a name="zh-cn_topic_0000002313236861_p5591634201811"></a>pingmesh任务间隔时间，单位为秒。</p>
        </td>
        </tr>
        </tbody>
        </table>

**查看检测结果<a name="zh-cn_topic_0000002313236861_section74321914202214"></a>**

网络检测的pingmesh结果将写入文件<nodename\>.log中，该文件中各字段的详细说明如下表所示。

**表 2**  <nodename\>.log文件参数说明

<a name="zh-cn_topic_0000002313236861_table1485915561131"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002313236861_row1786015564131"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002313236861_p1686085661319"><a name="zh-cn_topic_0000002313236861_p1686085661319"></a><a name="zh-cn_topic_0000002313236861_p1686085661319"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002313236861_p48601656141315"><a name="zh-cn_topic_0000002313236861_p48601656141315"></a><a name="zh-cn_topic_0000002313236861_p48601656141315"></a>取值范围</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002313236861_p6860145611312"><a name="zh-cn_topic_0000002313236861_p6860145611312"></a><a name="zh-cn_topic_0000002313236861_p6860145611312"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002313236861_row16860175615131"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p1286075615138"><a name="zh-cn_topic_0000002313236861_p1286075615138"></a><a name="zh-cn_topic_0000002313236861_p1286075615138"></a>uid</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p48601956101317"><a name="zh-cn_topic_0000002313236861_p48601956101317"></a><a name="zh-cn_topic_0000002313236861_p48601956101317"></a>长度为64的字符串。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p2860135601311"><a name="zh-cn_topic_0000002313236861_p2860135601311"></a><a name="zh-cn_topic_0000002313236861_p2860135601311"></a>本次pingmesh任务的ID。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row686012562133"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p1086045610132"><a name="zh-cn_topic_0000002313236861_p1086045610132"></a><a name="zh-cn_topic_0000002313236861_p1086045610132"></a>config</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p15860956121314"><a name="zh-cn_topic_0000002313236861_p15860956121314"></a><a name="zh-cn_topic_0000002313236861_p15860956121314"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p16860195691318"><a name="zh-cn_topic_0000002313236861_p16860195691318"></a><a name="zh-cn_topic_0000002313236861_p16860195691318"></a>本次pingmesh任务的用户配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row78601956191317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p1486065681312"><a name="zh-cn_topic_0000002313236861_p1486065681312"></a><a name="zh-cn_topic_0000002313236861_p1486065681312"></a>physicID</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p6860125616138"><a name="zh-cn_topic_0000002313236861_p6860125616138"></a><a name="zh-cn_topic_0000002313236861_p6860125616138"></a>[0~15]</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p986085620136"><a name="zh-cn_topic_0000002313236861_p986085620136"></a><a name="zh-cn_topic_0000002313236861_p986085620136"></a>NPU卡物理ID。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row8938512181616"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p14938612121611"><a name="zh-cn_topic_0000002313236861_p14938612121611"></a><a name="zh-cn_topic_0000002313236861_p14938612121611"></a>taskID</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002313236861_ul1816692919182"></a><a name="zh-cn_topic_0000002313236861_ul1816692919182"></a><ul id="zh-cn_topic_0000002313236861_ul1816692919182"><li>节点内部的任务：0</li><li>节点间的任务：1</li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p1893918126167"><a name="zh-cn_topic_0000002313236861_p1893918126167"></a><a name="zh-cn_topic_0000002313236861_p1893918126167"></a>任务ID。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row111821021619"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p20185100166"><a name="zh-cn_topic_0000002313236861_p20185100166"></a><a name="zh-cn_topic_0000002313236861_p20185100166"></a>DestNum</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p71816107161"><a name="zh-cn_topic_0000002313236861_p71816107161"></a><a name="zh-cn_topic_0000002313236861_p71816107161"></a>[0~47]</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p518121014165"><a name="zh-cn_topic_0000002313236861_p518121014165"></a><a name="zh-cn_topic_0000002313236861_p518121014165"></a>本次pingmesh目标地址数量。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row1744217771618"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p114431271166"><a name="zh-cn_topic_0000002313236861_p114431271166"></a><a name="zh-cn_topic_0000002313236861_p114431271166"></a>source_addr</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p1244397151614"><a name="zh-cn_topic_0000002313236861_p1244397151614"></a><a name="zh-cn_topic_0000002313236861_p1244397151614"></a>ipv4网络地址。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p194431775168"><a name="zh-cn_topic_0000002313236861_p194431775168"></a><a name="zh-cn_topic_0000002313236861_p194431775168"></a>源地址。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row9675175214190"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p46751527193"><a name="zh-cn_topic_0000002313236861_p46751527193"></a><a name="zh-cn_topic_0000002313236861_p46751527193"></a>target_addr</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p196751452151912"><a name="zh-cn_topic_0000002313236861_p196751452151912"></a><a name="zh-cn_topic_0000002313236861_p196751452151912"></a>ipv4网络地址。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p18675105212199"><a name="zh-cn_topic_0000002313236861_p18675105212199"></a><a name="zh-cn_topic_0000002313236861_p18675105212199"></a>目标地址。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row27685011196"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p10761750121920"><a name="zh-cn_topic_0000002313236861_p10761750121920"></a><a name="zh-cn_topic_0000002313236861_p10761750121920"></a>suc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p1176105011199"><a name="zh-cn_topic_0000002313236861_p1176105011199"></a><a name="zh-cn_topic_0000002313236861_p1176105011199"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p12761750131912"><a name="zh-cn_topic_0000002313236861_p12761750131912"></a><a name="zh-cn_topic_0000002313236861_p12761750131912"></a>发送成功的包数量。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row109393284209"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p593982816204"><a name="zh-cn_topic_0000002313236861_p593982816204"></a><a name="zh-cn_topic_0000002313236861_p593982816204"></a>fail_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p59391128182012"><a name="zh-cn_topic_0000002313236861_p59391128182012"></a><a name="zh-cn_topic_0000002313236861_p59391128182012"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p39392028102015"><a name="zh-cn_topic_0000002313236861_p39392028102015"></a><a name="zh-cn_topic_0000002313236861_p39392028102015"></a>发送失败的包数量。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row1531531162018"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p5531153117202"><a name="zh-cn_topic_0000002313236861_p5531153117202"></a><a name="zh-cn_topic_0000002313236861_p5531153117202"></a>max_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002313236861_ul12976155462213"></a><a name="zh-cn_topic_0000002313236861_ul12976155462213"></a><ul id="zh-cn_topic_0000002313236861_ul12976155462213"><li>正常情况：非负值</li><li>ping失败：-1</li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p175311431142013"><a name="zh-cn_topic_0000002313236861_p175311431142013"></a><a name="zh-cn_topic_0000002313236861_p175311431142013"></a>最长响应时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row776203442013"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p117683417207"><a name="zh-cn_topic_0000002313236861_p117683417207"></a><a name="zh-cn_topic_0000002313236861_p117683417207"></a>min_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002313236861_ul157185718227"></a><a name="zh-cn_topic_0000002313236861_ul157185718227"></a><ul id="zh-cn_topic_0000002313236861_ul157185718227"><li>正常情况：非负值</li><li>ping失败：-1</li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p107643412205"><a name="zh-cn_topic_0000002313236861_p107643412205"></a><a name="zh-cn_topic_0000002313236861_p107643412205"></a>最短响应时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row1586065661317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p38609563137"><a name="zh-cn_topic_0000002313236861_p38609563137"></a><a name="zh-cn_topic_0000002313236861_p38609563137"></a>avg_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002313236861_ul1162483112311"></a><a name="zh-cn_topic_0000002313236861_ul1162483112311"></a><ul id="zh-cn_topic_0000002313236861_ul1162483112311"><li>正常情况：非负值</li><li>ping失败：-1</li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p158601156141317"><a name="zh-cn_topic_0000002313236861_p158601156141317"></a><a name="zh-cn_topic_0000002313236861_p158601156141317"></a>平均响应时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row0324102320222"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p1632422316225"><a name="zh-cn_topic_0000002313236861_p1632422316225"></a><a name="zh-cn_topic_0000002313236861_p1632422316225"></a>tp95_time</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002313236861_ul157411422316"></a><a name="zh-cn_topic_0000002313236861_ul157411422316"></a><ul id="zh-cn_topic_0000002313236861_ul157411422316"><li>正常情况：非负值</li><li>ping失败：-1</li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p13324162382219"><a name="zh-cn_topic_0000002313236861_p13324162382219"></a><a name="zh-cn_topic_0000002313236861_p13324162382219"></a>处于95%位置时的响应时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row11875172510222"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p13875192515228"><a name="zh-cn_topic_0000002313236861_p13875192515228"></a><a name="zh-cn_topic_0000002313236861_p13875192515228"></a>reply_stat_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p487511257221"><a name="zh-cn_topic_0000002313236861_p487511257221"></a><a name="zh-cn_topic_0000002313236861_p487511257221"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p9875192512213"><a name="zh-cn_topic_0000002313236861_p9875192512213"></a><a name="zh-cn_topic_0000002313236861_p9875192512213"></a>本次查询到的响应数量。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row19747191113231"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002313236861_p1874715117236"><a name="zh-cn_topic_0000002313236861_p1874715117236"></a><a name="zh-cn_topic_0000002313236861_p1874715117236"></a>ping_total_num</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002313236861_p1074721111230"><a name="zh-cn_topic_0000002313236861_p1074721111230"></a><a name="zh-cn_topic_0000002313236861_p1074721111230"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002313236861_p2747511152312"><a name="zh-cn_topic_0000002313236861_p2747511152312"></a><a name="zh-cn_topic_0000002313236861_p2747511152312"></a>本次任务累计的响应数量。</p>
</td>
</tr>
</tbody>
</table>

**查看gRPC上报结果<a name="zh-cn_topic_0000002313236861_section28851054410"></a>**

慢网络诊断到故障，会通过gRPC上报至ClusterD的公共故障管理中心。

ConfigMap文件会显示相关信息，5秒钟之后自动清除。

![](../../figures/scheduling/zh-cn_image_0000002300581874.png)

**已支持的慢网络故障<a name="zh-cn_topic_0000002313236861_section19919834124518"></a>**

<a name="zh-cn_topic_0000002313236861_table4804164084414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002313236861_row1680414018449"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="zh-cn_topic_0000002313236861_p1680411405446"><a name="zh-cn_topic_0000002313236861_p1680411405446"></a><a name="zh-cn_topic_0000002313236861_p1680411405446"></a>故障码</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="zh-cn_topic_0000002313236861_p280464074412"><a name="zh-cn_topic_0000002313236861_p280464074412"></a><a name="zh-cn_topic_0000002313236861_p280464074412"></a>故障说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="zh-cn_topic_0000002313236861_p3804114018440"><a name="zh-cn_topic_0000002313236861_p3804114018440"></a><a name="zh-cn_topic_0000002313236861_p3804114018440"></a>故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002313236861_row1080414409444"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0000002313236861_p17804114034413"><a name="zh-cn_topic_0000002313236861_p17804114034413"></a><a name="zh-cn_topic_0000002313236861_p17804114034413"></a>200001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0000002313236861_p1780424011445"><a name="zh-cn_topic_0000002313236861_p1780424011445"></a><a name="zh-cn_topic_0000002313236861_p1780424011445"></a>某节点中产生/恢复慢网络。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0000002313236861_p15804114074414"><a name="zh-cn_topic_0000002313236861_p15804114074414"></a><a name="zh-cn_topic_0000002313236861_p15804114074414"></a>NotHandleFault：暂不处理故障。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row4804140134413"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0000002313236861_p11804340204417"><a name="zh-cn_topic_0000002313236861_p11804340204417"></a><a name="zh-cn_topic_0000002313236861_p11804340204417"></a>200001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0000002313236861_p5804154074418"><a name="zh-cn_topic_0000002313236861_p5804154074418"></a><a name="zh-cn_topic_0000002313236861_p5804154074418"></a>超节点内的节点间产生/恢复慢网络。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0000002313236861_p1680412409449"><a name="zh-cn_topic_0000002313236861_p1680412409449"></a><a name="zh-cn_topic_0000002313236861_p1680412409449"></a>NotHandleFault：暂不处理故障。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002313236861_row1640781094918"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0000002313236861_p8407101016493"><a name="zh-cn_topic_0000002313236861_p8407101016493"></a><a name="zh-cn_topic_0000002313236861_p8407101016493"></a>200001012</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0000002313236861_p194071110134917"><a name="zh-cn_topic_0000002313236861_p194071110134917"></a><a name="zh-cn_topic_0000002313236861_p194071110134917"></a>未收敛到卡。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0000002313236861_p1176073411918"><a name="zh-cn_topic_0000002313236861_p1176073411918"></a><a name="zh-cn_topic_0000002313236861_p1176073411918"></a>NotHandleFault：暂不处理故障。</p>
</td>
</tr>
</tbody>
</table>



### 故障处理<a name="ZH-CN_TOPIC_0000002511346405"></a>

#### 故障决策说明<a name="ZH-CN_TOPIC_0000002511346435"></a>

在故障检测完成后，针对每一种故障模式，断点续训通过故障处理或故障容错来恢复训练业务。断点续训特性根据恢复粒度由粗到细提供Job级别重调度、Pod级别重调度、进程级别重调度、弹性训练、进程级在线恢复、算子级在线恢复多层故障处理系统。用户可根据实际情况选择使用对应的子特性。

**图 1**  故障决策说明<a name="fig2639326192019"></a>  
![](../../figures/scheduling/故障决策说明.png "故障决策说明")

上图中，容错速度代表故障发生到故障恢复的速度，成功率代表故障发生后故障完成恢复的成功率，易用性代表用户使用或集成的成本。

Job级别重调度、Pod级别重调度、进程级别重调度可支持当前断点续训支持的全部故障模式，但依赖存在备份冗余计算服务器资源。如果存在不可修复的硬件故障且无备份冗余计算服务器时，可以通过配置弹性训练功能进行缩容训练。进程级在线恢复当前支持片上内存故障和网络故障。算子级在线恢复当前支持芯片网络故障和灵衢网络故障。

断点续训多层故障处理系统不同层级根据恢复粒度由细到粗可以逐级回退，如[图2](#fig477415371217)所示，如果上一层恢复失败则可以回退到下一层处理方式。

**图 2**  恢复失败说明<a name="fig477415371217"></a>  
![](../../figures/scheduling/恢复失败说明.png "恢复失败说明")

**重调度模式<a name="zh-cn_topic_0000002198051753_section1536115719358"></a>**

1.  重调度模式：将任务调度到健康的芯片上，并隔离故障芯片。

    重调度模式默认为**Job级别重调度**，每次故障会停止所有的Pod。但在大规模任务中，停止所有Pod后再重调度的成本较高，存在故障恢复时间过长的问题。除此之外断点续训还提供**Pod级别重调度**功能，用户可根据任务规模配置，在故障时刻只停止故障相关的Pod后重调度少量Pod，从而达成故障的快速恢复。为了进一步缩短故障恢复时间、降低故障影响范围，断点续训还提供进程级别重调度及进程级在线恢复功能。

    **表 1**  各种重调度级别的差异

    <a name="zh-cn_topic_0000002198051753_table18771108163419"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002198051753_row1677218810341"><th class="cellrowborder" valign="top" width="18.3%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002198051753_p121383187176"><a name="zh-cn_topic_0000002198051753_p121383187176"></a><a name="zh-cn_topic_0000002198051753_p121383187176"></a>重调度的级别</p>
    </th>
    <th class="cellrowborder" valign="top" width="31.7%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002198051753_p17773983341"><a name="zh-cn_topic_0000002198051753_p17773983341"></a><a name="zh-cn_topic_0000002198051753_p17773983341"></a>恢复训练耗时</p>
    </th>
    <th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002198051753_p1077316873410"><a name="zh-cn_topic_0000002198051753_p1077316873410"></a><a name="zh-cn_topic_0000002198051753_p1077316873410"></a>配置步骤</p>
    </th>
    <th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002198051753_p7773082347"><a name="zh-cn_topic_0000002198051753_p7773082347"></a><a name="zh-cn_topic_0000002198051753_p7773082347"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002198051753_row6773118193410"><td class="cellrowborder" valign="top" width="18.3%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002198051753_p2138131821719"><a name="zh-cn_topic_0000002198051753_p2138131821719"></a><a name="zh-cn_topic_0000002198051753_p2138131821719"></a>Job级别重调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="31.7%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002198051753_p12773128133419"><a name="zh-cn_topic_0000002198051753_p12773128133419"></a><a name="zh-cn_topic_0000002198051753_p12773128133419"></a>Job级重调度的恢复时间较长，随着任务规模增加恢复时间超线性劣化。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002198051753_p17773684341"><a name="zh-cn_topic_0000002198051753_p17773684341"></a><a name="zh-cn_topic_0000002198051753_p17773684341"></a>Job级重调度操作步骤简单，使用MindCluster的用户仅打开配置开关即可使用。</p>
    <p id="zh-cn_topic_0000002198051753_p6932163394014"><a name="zh-cn_topic_0000002198051753_p6932163394014"></a><a name="zh-cn_topic_0000002198051753_p6932163394014"></a>关键配置步骤请参见<a href="#配置job级别重调度">配置Job级别重调度</a>。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002198051753_p17353194312415"><a name="zh-cn_topic_0000002198051753_p17353194312415"></a><a name="zh-cn_topic_0000002198051753_p17353194312415"></a>为了进一步降低恢复中资源调度时间，用户可以选择在Job级重调度上开启Pod级重调度能力。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002198051753_row6773198163411"><td class="cellrowborder" valign="top" width="18.3%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002198051753_p151382018151714"><a name="zh-cn_topic_0000002198051753_p151382018151714"></a><a name="zh-cn_topic_0000002198051753_p151382018151714"></a>Pod级别重调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="31.7%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002198051753_p15773581345"><a name="zh-cn_topic_0000002198051753_p15773581345"></a><a name="zh-cn_topic_0000002198051753_p15773581345"></a>Pod级重调度可以将资源调度时间缩短，且与任务规模无关。但是，Pod级重调度并不能优化训练初始化过程中的时间开销，整体恢复时间仍然会随着任务规模增加而超线性劣化。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002198051753_p87735823415"><a name="zh-cn_topic_0000002198051753_p87735823415"></a><a name="zh-cn_topic_0000002198051753_p87735823415"></a>Pod级重调度用户需要额外在训练容器中集成训练进程管理能力，使用MindCluster的用户具备对应进程管理能力后即可使用。</p>
    <p id="zh-cn_topic_0000002198051753_p128503416618"><a name="zh-cn_topic_0000002198051753_p128503416618"></a><a name="zh-cn_topic_0000002198051753_p128503416618"></a>关键配置步骤请参见<a href="#配置pod级别重调度">配置Pod级别重调度</a>。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002198051753_p12146019184418"><a name="zh-cn_topic_0000002198051753_p12146019184418"></a><a name="zh-cn_topic_0000002198051753_p12146019184418"></a>为了进一步降低训练初始化中的恢复时间，用户可以选择在Pod级重调度上开启进程级重调度能力。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002198051753_row127749818346"><td class="cellrowborder" valign="top" width="18.3%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002198051753_p161385189177"><a name="zh-cn_topic_0000002198051753_p161385189177"></a><a name="zh-cn_topic_0000002198051753_p161385189177"></a>进程级别重调度（进程级恢复）</p>
    </td>
    <td class="cellrowborder" valign="top" width="31.7%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002198051753_p7774168103418"><a name="zh-cn_topic_0000002198051753_p7774168103418"></a><a name="zh-cn_topic_0000002198051753_p7774168103418"></a>进程级重调度可以减少训练初始化时间，将整体恢复时间缩短，且与任务规模无关或者弱相关。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002198051753_p977468113412"><a name="zh-cn_topic_0000002198051753_p977468113412"></a><a name="zh-cn_topic_0000002198051753_p977468113412"></a>相比Pod级重调度，进程级重调度用户需要额外在训练框架中集成高可用训练能力，使用MindCluster的用户需要修改训练脚本，并开启对应配置开关后使用。</p>
    <p id="zh-cn_topic_0000002198051753_p13754195013114"><a name="zh-cn_topic_0000002198051753_p13754195013114"></a><a name="zh-cn_topic_0000002198051753_p13754195013114"></a>关键配置步骤请参见<a href="#配置进程级别重调度">配置进程级别重调度</a>。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002198051753_p19902018144519"><a name="zh-cn_topic_0000002198051753_p19902018144519"></a><a name="zh-cn_topic_0000002198051753_p19902018144519"></a>为了解决大规模场景下MTBF时间较短的问题，进一步降低整体恢复时间，用户可以选择在进程级重调度上开启进程级在线恢复能力。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002198051753_row54527214384"><td class="cellrowborder" valign="top" width="18.3%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002198051753_p103941343197"><a name="zh-cn_topic_0000002198051753_p103941343197"></a><a name="zh-cn_topic_0000002198051753_p103941343197"></a>进程级在线恢复</p>
    </td>
    <td class="cellrowborder" valign="top" width="31.7%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002198051753_p14536218388"><a name="zh-cn_topic_0000002198051753_p14536218388"></a><a name="zh-cn_topic_0000002198051753_p14536218388"></a>进程级在线恢复比起进程级重调度，恢复训练耗时更低。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002198051753_p194531243818"><a name="zh-cn_topic_0000002198051753_p194531243818"></a><a name="zh-cn_topic_0000002198051753_p194531243818"></a>相比进程级重调度，进程级在线恢复用户需要配置对应的配置开关后使用。</p>
    <p id="zh-cn_topic_0000002198051753_p1471913721316"><a name="zh-cn_topic_0000002198051753_p1471913721316"></a><a name="zh-cn_topic_0000002198051753_p1471913721316"></a>关键配置步骤请参见<a href="#配置进程级在线恢复">配置进程级在线恢复</a>。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002198051753_p24533216388"><a name="zh-cn_topic_0000002198051753_p24533216388"></a><a name="zh-cn_topic_0000002198051753_p24533216388"></a>当前进程级在线恢复支持<span id="ph1024411844215"><a name="ph1024411844215"></a><a name="ph1024411844215"></a>片上内存</span>故障和网络故障，其余故障场景将回退其他处理方式。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002198051753_row04530223810"><td class="cellrowborder" valign="top" width="18.3%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002198051753_p51079323718"><a name="zh-cn_topic_0000002198051753_p51079323718"></a><a name="zh-cn_topic_0000002198051753_p51079323718"></a>算子级在线恢复</p>
    </td>
    <td class="cellrowborder" valign="top" width="31.7%" headers="mcps1.2.5.1.2 "><p id="p18195195524"><a name="p18195195524"></a><a name="p18195195524"></a>--</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p14207044175119"><a name="p14207044175119"></a><a name="p14207044175119"></a>关键配置步骤请参见<a href="#配置算子级在线恢复">配置算子级在线恢复</a>。</p>
    </td>
    <td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p364851219509"><a name="p364851219509"></a><a name="p364851219509"></a>--</p>
    </td>
    </tr>
    </tbody>
    </table>

2.  重调度模式存在以下两种重调度策略。

    -   **直接重调度**：训练过程中发生集群调度组件可以探测到的硬件故障，系统将故障节点或芯片进行隔离，直接对任务进行重调度。
    -   **无条件重试**：训练过程中发生集群调度组件不能探测到的故障，导致任务容器异常退出，系统无条件对任务进行重调度。

    **表 2**  重调度策略说明

    <a name="zh-cn_topic_0000002198051753_table37727194382"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002198051753_row17721919153816"><th class="cellrowborder" valign="top" width="12.22%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002198051753_p47721719203817"><a name="zh-cn_topic_0000002198051753_p47721719203817"></a><a name="zh-cn_topic_0000002198051753_p47721719203817"></a>重调度策略</p>
    </th>
    <th class="cellrowborder" valign="top" width="63.68000000000001%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002198051753_p177371918382"><a name="zh-cn_topic_0000002198051753_p177371918382"></a><a name="zh-cn_topic_0000002198051753_p177371918382"></a>说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="24.099999999999998%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002198051753_p13773111983816"><a name="zh-cn_topic_0000002198051753_p13773111983816"></a><a name="zh-cn_topic_0000002198051753_p13773111983816"></a>支持的故障类型</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002198051753_row477315199387"><td class="cellrowborder" valign="top" width="12.22%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002198051753_p107731019113818"><a name="zh-cn_topic_0000002198051753_p107731019113818"></a><a name="zh-cn_topic_0000002198051753_p107731019113818"></a>直接重调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="63.68000000000001%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002198051753_p18773131915387"><a name="zh-cn_topic_0000002198051753_p18773131915387"></a><a name="zh-cn_topic_0000002198051753_p18773131915387"></a>系统将故障的节点或芯片进行隔离，然后直接对任务进行重调度。</p>
    </td>
    <td class="cellrowborder" valign="top" width="24.099999999999998%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002198051753_p877331915384"><a name="zh-cn_topic_0000002198051753_p877331915384"></a><a name="zh-cn_topic_0000002198051753_p877331915384"></a>已知的节点故障或重调度处理级别芯片故障。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002198051753_row277331919383"><td class="cellrowborder" valign="top" width="12.22%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002198051753_p1773519123817"><a name="zh-cn_topic_0000002198051753_p1773519123817"></a><a name="zh-cn_topic_0000002198051753_p1773519123817"></a>无条件重试</p>
    </td>
    <td class="cellrowborder" valign="top" width="63.68000000000001%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002198051753_p7773131963816"><a name="zh-cn_topic_0000002198051753_p7773131963816"></a><a name="zh-cn_topic_0000002198051753_p7773131963816"></a>系统对配置了无条件重试次数的任务，进行指定次数内的重调度。</p>
    <p id="zh-cn_topic_0000002198051753_p127739195385"><a name="zh-cn_topic_0000002198051753_p127739195385"></a><a name="zh-cn_topic_0000002198051753_p127739195385"></a>成功重调度后，任务可重试次数将减1，当可重试次数为0时无法再次触发重调度。</p>
    <div class="note" id="zh-cn_topic_0000002198051753_note1878524412312"><a name="zh-cn_topic_0000002198051753_note1878524412312"></a><div class="notebody"><p id="zh-cn_topic_0000002198051753_p1178574452313"><a name="zh-cn_topic_0000002198051753_p1178574452313"></a><a name="zh-cn_topic_0000002198051753_p1178574452313"></a>如需使用无条件重试功能，需在YAML中配置fault-retry-times参数，详细参数说明请参见<a href="#yaml参数说明">YAML参数说明</a>。</p>
    </div></div>
    </td>
    <td class="cellrowborder" valign="top" width="24.099999999999998%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002198051753_p977341963818"><a name="zh-cn_topic_0000002198051753_p977341963818"></a><a name="zh-cn_topic_0000002198051753_p977341963818"></a>由于参数面网络故障或者训练相关软件故障等，导致任务异常退出，Pod的Status变为Failed状态的相关故障。</p>
    </td>
    </tr>
    </tbody>
    </table>


#### Job级别重调度<a name="ZH-CN_TOPIC_0000002479226586"></a>

**Job级别重调度**即每次故障停止所有Pod，重新创建并重调度所有Pod后，重启训练任务。重调度模式默认为**Job级别重调度**。

了解Job级别重调度的关键配置步骤，请参见[配置Job级别重调度](#配置job级别重调度)。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

-   本功能仅支持在6.0.RC2及以上版本中使用。
-   大规模K8s集群场景下，ConfigMap映射时延不可控，建议RankTable使用共享存储方式。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 1**  job级别重调度支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="17.691769176917692%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="67.3067306730673%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="15.001500150014998%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row15199711175415"><td class="cellrowborder" valign="top" width="17.691769176917692%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p1119914117545"><a name="zh-cn_topic_0000002039194017_p1119914117545"></a><a name="zh-cn_topic_0000002039194017_p1119914117545"></a><span id="zh-cn_topic_0000002039194017_ph7199711155420"><a name="zh-cn_topic_0000002039194017_ph7199711155420"></a><a name="zh-cn_topic_0000002039194017_ph7199711155420"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="67.3067306730673%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039194017_ul11991811125414"></a><a name="zh-cn_topic_0000002039194017_ul11991811125414"></a><ul id="zh-cn_topic_0000002039194017_ul11991811125414"><li><span id="zh-cn_topic_0000002039194017_ph13085521289"><a name="zh-cn_topic_0000002039194017_ph13085521289"></a><a name="zh-cn_topic_0000002039194017_ph13085521289"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span><div class="note" id="zh-cn_topic_0000002039194017_note11304039162817"><a name="zh-cn_topic_0000002039194017_note11304039162817"></a><a name="zh-cn_topic_0000002039194017_note11304039162817"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039194017_p1030415395282"><a name="zh-cn_topic_0000002039194017_p1030415395282"></a><a name="zh-cn_topic_0000002039194017_p1030415395282"></a>若<span id="zh-cn_topic_0000002039194017_ph143042039102813"><a name="zh-cn_topic_0000002039194017_ph143042039102813"></a><a name="zh-cn_topic_0000002039194017_ph143042039102813"></a>Atlas 800 训练服务器</span>的芯片工作模式为SMP模式，且每个Pod申请的NPU数量为1、2时，不支持使用重调度模式。查询和设置NPU芯片工作模式的详细介绍请参见<span id="zh-cn_topic_0000002039194017_ph4304193972810"><a name="zh-cn_topic_0000002039194017_ph4304193972810"></a><a name="zh-cn_topic_0000002039194017_ph4304193972810"></a>《Atlas 800 训练服务器 iBMC用户指南（型号 9000）》中的“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a" target="_blank" rel="noopener noreferrer">查询和设置NPU芯片工作模式（npuworkmode）</a>”</span>章节。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="15.001500150014998%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039194017_ul102005117544"></a><a name="zh-cn_topic_0000002039194017_ul102005117544"></a><ul id="zh-cn_topic_0000002039194017_ul102005117544"><li><span id="zh-cn_topic_0000002039194017_ph102009114549"><a name="zh-cn_topic_0000002039194017_ph102009114549"></a><a name="zh-cn_topic_0000002039194017_ph102009114549"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039194017_ph1200131111547"><a name="zh-cn_topic_0000002039194017_ph1200131111547"></a><a name="zh-cn_topic_0000002039194017_ph1200131111547"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002039194017_ph9200511185413"><a name="zh-cn_topic_0000002039194017_ph9200511185413"></a><a name="zh-cn_topic_0000002039194017_ph9200511185413"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="17.691769176917692%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="zh-cn_topic_0000002039194017_ph72011311155419"><a name="zh-cn_topic_0000002039194017_ph72011311155419"></a><a name="zh-cn_topic_0000002039194017_ph72011311155419"></a>Atlas A2 训练系列产品</span></p>
<p id="p773278122616"><a name="p773278122616"></a><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="67.3067306730673%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039194017_ul1720110113545"></a><a name="zh-cn_topic_0000002039194017_ul1720110113545"></a><ul id="zh-cn_topic_0000002039194017_ul1720110113545"><li>Atlas 800T A2 训练服务器</li><li><span id="zh-cn_topic_0000002039194017_ph32011711115415"><a name="zh-cn_topic_0000002039194017_ph32011711115415"></a><a name="zh-cn_topic_0000002039194017_ph32011711115415"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002039194017_ph19201511175411"><a name="zh-cn_topic_0000002039194017_ph19201511175411"></a><a name="zh-cn_topic_0000002039194017_ph19201511175411"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="15.001500150014998%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><ul id="zh-cn_topic_0000002039194017_ul7201511105411"><li><span id="zh-cn_topic_0000002039194017_ph52034113546"><a name="zh-cn_topic_0000002039194017_ph52034113546"></a><a name="zh-cn_topic_0000002039194017_ph52034113546"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039194017_ph13203811195412"><a name="zh-cn_topic_0000002039194017_ph13203811195412"></a><a name="zh-cn_topic_0000002039194017_ph13203811195412"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002039194017_ph620418118547"><a name="zh-cn_topic_0000002039194017_ph620418118547"></a><a name="zh-cn_topic_0000002039194017_ph620418118547"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="17.691769176917692%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="zh-cn_topic_0000002039194017_ph1020491175416"><a name="zh-cn_topic_0000002039194017_ph1020491175416"></a><a name="zh-cn_topic_0000002039194017_ph1020491175416"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="67.3067306730673%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039194017_ul2592123411423"></a><a name="zh-cn_topic_0000002039194017_ul2592123411423"></a><ul id="zh-cn_topic_0000002039194017_ul2592123411423"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="15.001500150014998%" headers="mcps1.2.4.1.3 "><a name="ul989793310113"></a><a name="ul989793310113"></a><ul id="ul989793310113"><li><span id="ph15897173313113"><a name="ph15897173313113"></a><a name="ph15897173313113"></a>MindSpore</span></li><li><span id="ph9897153311119"><a name="ph9897153311119"></a><a name="ph9897153311119"></a>TensorFlow</span></li><li><span id="ph3897123341117"><a name="ph3897123341117"></a><a name="ph3897123341117"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row32411310161716"><td class="cellrowborder" valign="top" width="17.691769176917692%" headers="mcps1.2.4.1.1 "><p id="p27061515111717"><a name="p27061515111717"></a><a name="p27061515111717"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="67.3067306730673%" headers="mcps1.2.4.1.2 "><p id="p1653852019211"><a name="p1653852019211"></a><a name="p1653852019211"></a><span id="ph15531211214"><a name="ph15531211214"></a><a name="ph15531211214"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.001500150014998%" headers="mcps1.2.4.1.3 "><a name="ul866972191811"></a><a name="ul866972191811"></a><ul id="ul866972191811"><li><span id="ph66691921201815"><a name="ph66691921201815"></a><a name="ph66691921201815"></a>MindSpore</span></li><li><span id="ph126691021121818"><a name="ph126691021121818"></a><a name="ph126691021121818"></a>TensorFlow</span></li><li><span id="ph16692216187"><a name="ph16692216187"></a><a name="ph16692216187"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002039194017_section57901137171110"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。Job级别重调度首先销毁所有的训练容器，然后隔离故障设备，再重新将训练容器调度启动。训练容器重新启动后重新拉起训练，该行为类似训练首次拉起过程。

**图 1**  原理图<a name="fig18343114924113"></a>  
![](../../figures/scheduling/原理图.png "原理图")

在以上原理图中，各个步骤的说明如下。

1.  检测到故障后，首先删除当前任务所有的Pod和容器。
2.  隔离故障所在的设备，防止再次使用该设备。
3.  重新创建和调度训练Pod和容器。
4.  容器启动后，拉起训练进程恢复训练任务。


#### Pod级别重调度<a name="ZH-CN_TOPIC_0000002511346429"></a>

**Pod级别重调度**即每次故障只停止故障相关的Pod，重新创建并重调度故障相关的Pod后，重启训练任务。如果当前故障不能恢复，则回退至Job级重调度模式。相比于Job级别重调度，Pod级别重调度会减少部分资源调度、Pod创建的时间。

了解Pod级别重调度的关键配置步骤，请参见[配置Pod级别重调度](#配置pod级别重调度)。

**使用约束<a name="zh-cn_topic_0000002003034876_section11983145119441"></a>**

-   在大集群训练任务中使用**Pod级别重调度**时，建议设置open files参数（可以打开的最大文件数目）足够大，设置过小可能导致Pod重调度出现异常。例如执行**ulimit -n 100000**命令，将open files参数设置为100000。
-   当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不触发Pod级别重调度和进程级别重调度，直接触发Job级别重调度。
-   请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002003034876_section48174410591"></a>**

**表 1**  重调度支持的产品和框架

<a name="zh-cn_topic_0000002003034876_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003034876_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002003034876_p199171819164417"><a name="zh-cn_topic_0000002003034876_p199171819164417"></a><a name="zh-cn_topic_0000002003034876_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="63.10631063106311%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002003034876_p2917819114420"><a name="zh-cn_topic_0000002003034876_p2917819114420"></a><a name="zh-cn_topic_0000002003034876_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="16.43164316431643%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002003034876_p27578257424"><a name="zh-cn_topic_0000002003034876_p27578257424"></a><a name="zh-cn_topic_0000002003034876_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003034876_row12917151994410"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p339114714459"><a name="zh-cn_topic_0000002003034876_p339114714459"></a><a name="zh-cn_topic_0000002003034876_p339114714459"></a><span id="zh-cn_topic_0000002003034876_ph327965117217"><a name="zh-cn_topic_0000002003034876_ph327965117217"></a><a name="zh-cn_topic_0000002003034876_ph327965117217"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul17412295261"></a><a name="zh-cn_topic_0000002003034876_ul17412295261"></a><ul id="zh-cn_topic_0000002003034876_ul17412295261"><li><span id="ph1179307345"><a name="ph1179307345"></a><a name="ph1179307345"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span><div class="note" id="zh-cn_topic_0000002003034876_note186291241356"><a name="zh-cn_topic_0000002003034876_note186291241356"></a><a name="zh-cn_topic_0000002003034876_note186291241356"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003034876_p86294411854"><a name="zh-cn_topic_0000002003034876_p86294411854"></a><a name="zh-cn_topic_0000002003034876_p86294411854"></a>若<span id="zh-cn_topic_0000002003034876_ph1162924110518"><a name="zh-cn_topic_0000002003034876_ph1162924110518"></a><a name="zh-cn_topic_0000002003034876_ph1162924110518"></a>Atlas 800 训练服务器</span>的芯片工作模式为SMP模式，且每个Pod申请的NPU数量为1、2时，不支持使用重调度模式。查询和设置NPU芯片工作模式的详细介绍请参见<span id="zh-cn_topic_0000002003034876_ph66296417518"><a name="zh-cn_topic_0000002003034876_ph66296417518"></a><a name="zh-cn_topic_0000002003034876_ph66296417518"></a>《Atlas 800 训练服务器 iBMC用户指南（型号 9000）》中的“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a" target="_blank" rel="noopener noreferrer">查询和设置NPU芯片工作模式（npuworkmode）</a>”</span>章节。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002003034876_ul353572894311"></a><a name="zh-cn_topic_0000002003034876_ul353572894311"></a><ul id="zh-cn_topic_0000002003034876_ul353572894311"><li><span id="zh-cn_topic_0000002003034876_ph2075216585425"><a name="zh-cn_topic_0000002003034876_ph2075216585425"></a><a name="zh-cn_topic_0000002003034876_ph2075216585425"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002003034876_ph12195638125217"><a name="zh-cn_topic_0000002003034876_ph12195638125217"></a><a name="zh-cn_topic_0000002003034876_ph12195638125217"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002003034876_ph19355165113512"><a name="zh-cn_topic_0000002003034876_ph19355165113512"></a><a name="zh-cn_topic_0000002003034876_ph19355165113512"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002003034876_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p153913472453"><a name="zh-cn_topic_0000002003034876_p153913472453"></a><a name="zh-cn_topic_0000002003034876_p153913472453"></a><span id="zh-cn_topic_0000002003034876_ph151431757142112"><a name="zh-cn_topic_0000002003034876_ph151431757142112"></a><a name="zh-cn_topic_0000002003034876_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p15647160165615"><a name="p15647160165615"></a><a name="p15647160165615"></a></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul1843217118563"></a><a name="zh-cn_topic_0000002003034876_ul1843217118563"></a><ul id="zh-cn_topic_0000002003034876_ul1843217118563"><li><span id="ph2153181425619"><a name="ph2153181425619"></a><a name="ph2153181425619"></a>Atlas 800T A2 训练服务器</span></li><li><span id="zh-cn_topic_0000002003034876_ph1114211211203"><a name="zh-cn_topic_0000002003034876_ph1114211211203"></a><a name="zh-cn_topic_0000002003034876_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002003034876_ph495114991519"><a name="zh-cn_topic_0000002003034876_ph495114991519"></a><a name="zh-cn_topic_0000002003034876_ph495114991519"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002003034876_ul693112434815"></a><a name="zh-cn_topic_0000002003034876_ul693112434815"></a><ul id="zh-cn_topic_0000002003034876_ul693112434815"><li><span id="zh-cn_topic_0000002003034876_ph1393112494820"><a name="zh-cn_topic_0000002003034876_ph1393112494820"></a><a name="zh-cn_topic_0000002003034876_ph1393112494820"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002003034876_ph2932182416487"><a name="zh-cn_topic_0000002003034876_ph2932182416487"></a><a name="zh-cn_topic_0000002003034876_ph2932182416487"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002003034876_ph2093210246488"><a name="zh-cn_topic_0000002003034876_ph2093210246488"></a><a name="zh-cn_topic_0000002003034876_ph2093210246488"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002003034876_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p18222246142212"><a name="zh-cn_topic_0000002003034876_p18222246142212"></a><a name="zh-cn_topic_0000002003034876_p18222246142212"></a><span id="zh-cn_topic_0000002003034876_ph18411121792018"><a name="zh-cn_topic_0000002003034876_ph18411121792018"></a><a name="zh-cn_topic_0000002003034876_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul1367372444211"></a><a name="zh-cn_topic_0000002003034876_ul1367372444211"></a><ul id="zh-cn_topic_0000002003034876_ul1367372444211"><li><p id="p14426829306"><a name="p14426829306"></a><a name="p14426829306"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><ul id="zh-cn_topic_0000002039194017_ul7201511105411"><li><span id="zh-cn_topic_0000002039194017_ph52034113546"><a name="zh-cn_topic_0000002039194017_ph52034113546"></a><a name="zh-cn_topic_0000002039194017_ph52034113546"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039194017_ph13203811195412"><a name="zh-cn_topic_0000002039194017_ph13203811195412"></a><a name="zh-cn_topic_0000002039194017_ph13203811195412"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002039194017_ph620418118547"><a name="zh-cn_topic_0000002039194017_ph620418118547"></a><a name="zh-cn_topic_0000002039194017_ph620418118547"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row999211122017"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="p09912115201"><a name="p09912115201"></a><a name="p09912115201"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><p id="p49961172020"><a name="p49961172020"></a><a name="p49961172020"></a><span id="ph6124114710214"><a name="ph6124114710214"></a><a name="ph6124114710214"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="ul5581185452113"></a><a name="ul5581185452113"></a><ul id="ul5581185452113"><li><span id="ph19581195472117"><a name="ph19581195472117"></a><a name="ph19581195472117"></a>MindSpore</span></li><li><span id="ph1758125422113"><a name="ph1758125422113"></a><a name="ph1758125422113"></a>TensorFlow</span></li><li><span id="ph8581154132114"><a name="ph8581154132114"></a><a name="ph8581154132114"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002003034876_section19557184814234"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。Pod级别重调度首先销毁当前任务中故障的Pod和容器，并通知其他训练容器中的管理进程销毁所有训练进程，然后隔离故障设备，再重新将训练容器调度启动。待训练容器重新启动后，通知所有容器中的管理进程重新拉起训练进程恢复训练。

1.  检测到故障后，仅删除当前任务中故障的Pod和容器，销毁所有训练进程。
2.  隔离故障所在的设备，防止再次使用该设备。
3.  重新创建和调度训练Pod和容器。
4.  容器启动后，拉起训练进程恢复训练。


#### 进程级别重调度<a name="ZH-CN_TOPIC_0000002511346457"></a>

进程级别重调度即每次故障只停止故障相关节点的进程，根据配置策略判断是否退出故障节点。

-   recover策略：将故障节点的容器迁移到健康节点；
-   recover-in-place策略：对于发生以下两类故障的节点，仅重启故障进程，不迁移故障节点的容器。若多个节点同时发生故障，则只发生以下两类故障的节点仅重启故障进程，不迁移容器，发生其他类型故障的节点会迁移容器。若多个节点发生故障的类型只包含业务进程异常故障，则所有故障节点均会迁移容器。
    -   业务进程异常故障。
    -   RestartRequest和RestartBusiness级别的芯片故障。

不能恢复则回退至Job级或Pod级重调度模式。相比于Pod级别重调度，本功能仅重调度故障进程，减少了大量进程间不同步的等待耗时。同时利用了新的HCCL建链方案大大降低了建链耗时，且通过NPU卡间的参数面高速网络P2P传递CKPT信息，避免了CKPT保存和加载的耗时。

了解进程级别重调度的关键配置步骤，请参见[配置进程级别重调度](#配置进程级别重调度)。

>[!NOTE] 说明 
>-   参数面传递CKPT信息依赖故障卡中的全量优化器副本，如果不存在全量优化器副本则回退为加载存储上的CKPT文件恢复参数。
>-   优化器副本依赖额外的显存占用，如果用户的显存较为紧张，可选择本地加载模式，无论是否存在优化器副本都直接加载存储上的CKPT文件恢复参数。

**使用约束<a name="zh-cn_topic_0000002039353153_section514611624316"></a>**

-   进程级重调度支持的版本配套关系如下。
    -   PyTorch版本为2.7.1。
    -   MindSpeed-LLM版本为2.3.0版本。

-   当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不触发Pod级别重调度和进程级别重调度，直接触发Job级别重调度。
-   不能和优雅容错功能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
-   MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
-   MindSpore场景下，需要在启动TaskD  Manager前设置export TASKD\_PROCESS\_ENABLE="on"。
-   请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

-   只支持PyTorch单算子模式，只支持基于Megatron框架的模型，只支持acjob类型训练任务。

-   只支持单容器迁移，不支持按照亲和性迁移。
-   不支持多模态模型。
-   不支持开启watchdog功能。
-   Atlas A3 训练系列产品场景下，若发生NPU掉卡类、OS断连类的故障，可导致进程级别重调度失败。
-   当故障发生在HCCL建链阶段时，会导致进程级别重调度失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考[配置HCCL主动触发建链](#配置hccl主动触发建链)章节进行提前建链，防止故障出现在HCCL建链阶段。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039353153_section136131584164"></a>**

**表 1**  重调度支持的产品和框架

<a name="zh-cn_topic_0000002039353153_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039353153_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039353153_p199171819164417"><a name="zh-cn_topic_0000002039353153_p199171819164417"></a><a name="zh-cn_topic_0000002039353153_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="66.2966296629663%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039353153_p2917819114420"><a name="zh-cn_topic_0000002039353153_p2917819114420"></a><a name="zh-cn_topic_0000002039353153_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="13.24132413241324%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039353153_p27578257424"><a name="zh-cn_topic_0000002039353153_p27578257424"></a><a name="zh-cn_topic_0000002039353153_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039353153_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p153913472453"><a name="zh-cn_topic_0000002039353153_p153913472453"></a><a name="zh-cn_topic_0000002039353153_p153913472453"></a><span id="zh-cn_topic_0000002039353153_ph151431757142112"><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p737515258512"><a name="p737515258512"></a><a name="p737515258512"></a></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039353153_ul1843217118563"></a><a name="zh-cn_topic_0000002039353153_ul1843217118563"></a><ul id="zh-cn_topic_0000002039353153_ul1843217118563"><li><p id="p1546725019404"><a name="p1546725019404"></a><a name="p1546725019404"></a><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></p>
</li><li><span id="zh-cn_topic_0000002039353153_ph1114211211203"><a name="zh-cn_topic_0000002039353153_ph1114211211203"></a><a name="zh-cn_topic_0000002039353153_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002039353153_ph495114991519"><a name="zh-cn_topic_0000002039353153_ph495114991519"></a><a name="zh-cn_topic_0000002039353153_ph495114991519"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039353153_ul693112434815"></a><a name="zh-cn_topic_0000002039353153_ul693112434815"></a><ul id="zh-cn_topic_0000002039353153_ul693112434815"><li><span id="zh-cn_topic_0000002039353153_ph1393112494820"><a name="zh-cn_topic_0000002039353153_ph1393112494820"></a><a name="zh-cn_topic_0000002039353153_ph1393112494820"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039353153_ph2093210246488"><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039353153_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p18222246142212"><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><span id="zh-cn_topic_0000002039353153_ph18411121792018"><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><a name="ul61561253231"></a><a name="ul61561253231"></a><ul id="ul61561253231"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><a name="ul18946810161311"></a><a name="ul18946810161311"></a><ul id="ul18946810161311"><li><span id="ph99461100137"><a name="ph99461100137"></a><a name="ph99461100137"></a>MindSpore</span><p id="p664545214"><a name="p664545214"></a><a name="p664545214"></a><span id="ph294661010130"><a name="ph294661010130"></a><a name="ph294661010130"></a></span></p>
</li><li><span id="ph99469109139"><a name="ph99469109139"></a><a name="ph99469109139"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002039353153_section12206164333619"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。进程级重调度根据配置策略首先销毁故障的训练进程或容器，并通知其他训练容器中的训练进程暂停当前训练任务，然后隔离故障设备，再重新将训练容器调度启动。故障训练容器重新启动后，通知所有容器中的训练进程进行集合通信重建链。建链完成后，将CKPT通过参数面发送给新拉起的训练进程恢复参数，恢复后所有进程重新执行当前step恢复训练。

**图 1**  进程级别重调度原理示意图<a name="fig1373016583373"></a>  
![](../../figures/scheduling/进程级别重调度原理示意图.png "进程级别重调度原理示意图")

在以上原理图中，各个步骤的说明如下。

1.  设备出现硬件故障后，MindCluster在服务器上的检测组件上报故障信息到ClusterD中，软件故障由容器内MindIO Controller感知并上报到ClusterD。
2.  ClusterD将故障服务器上的任务容器退出故障训练进程，重新调度到备用的服务器上。
3.  ClusterD通知Master节点上的MindIO Controller进行容错，容错流程包括通知停止训练、通知全局故障、通知恢复策略。
4.  MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用PTA强制停止训练进程。MindIO Processor清理正常节点的资源，销毁通信域，清理后等待新进程加入。
5.  备用服务器上的管理进程拉起训练进程后，创建新的MindIO Processor，MindIO Controller通知每个训练进程中的MindIO Processor恢复训练。
6.  各个进程进行集合通信建链。
7.  正常服务器上的NPU通过参数面将CKPT传递到备用服务器上，完成参数状态恢复后继续训练。

**功能适配点<a name="section1446615300284"></a>**

在进程级别重调度中，集群大脑会根据全局故障信息决策恢复策略并将策略下发到MindIO，调度器需要支持故障Pod调度，而非整个任务重调度，支持恢复策略依次回退。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，在恢复时执行算子资源清理，节点重启后触发通信重建。通过参数面在线修复和状态回滚，完成进程级重调度恢复。

对于非MindSpeed-LLM和MindCluster平台用户，需在框架侧完成[表2](#table1995514113610)的功能适配。

**表 2**  进程级别重调度框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="16.77167716771677%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="43.23432343234324%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.13181318131813%" id="mcps1.2.5.1.3"><p id="p104301715185316"><a name="p104301715185316"></a><a name="p104301715185316"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="21.862186218621858%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="16.77167716771677%" headers="mcps1.2.5.1.1 "><p id="p18221046175418"><a name="p18221046175418"></a><a name="p18221046175418"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="43.23432343234324%" headers="mcps1.2.5.1.2 "><p id="p14221746205412"><a name="p14221746205412"></a><a name="p14221746205412"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="18.13181318131813%" headers="mcps1.2.5.1.3 "><p id="p5119132211596"><a name="p5119132211596"></a><a name="p5119132211596"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="21.862186218621858%" headers="mcps1.2.5.1.4 "><p id="p252632095917"><a name="p252632095917"></a><a name="p252632095917"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1122104645414"><a name="p1122104645414"></a><a name="p1122104645414"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p322446155419"><a name="p322446155419"></a><a name="p322446155419"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p152294645418"><a name="p152294645418"></a><a name="p152294645418"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p6221046125412"><a name="p6221046125412"></a><a name="p6221046125412"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row144014113397"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p52294618541"><a name="p52294618541"></a><a name="p52294618541"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p72284615410"><a name="p72284615410"></a><a name="p72284615410"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row74014111391"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p522194614547"><a name="p522194614547"></a><a name="p522194614547"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2221346125415"><a name="p2221346125415"></a><a name="p2221346125415"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row74025111392"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9229467542"><a name="p9229467542"></a><a name="p9229467542"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p102218466541"><a name="p102218466541"></a><a name="p102218466541"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row19531411367"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p422846105412"><a name="p422846105412"></a><a name="p422846105412"></a>节点重启及通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p922194612545"><a name="p922194612545"></a><a name="p922194612545"></a>通过注册重建回调实现健康节点与故障节点重建通信域。</p>
</td>
</tr>
<tr id="row1708112845416"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p722164611549"><a name="p722164611549"></a><a name="p722164611549"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1122246145417"><a name="p1122246145417"></a><a name="p1122246145417"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row1911610240547"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p92214463547"><a name="p92214463547"></a><a name="p92214463547"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p222154613543"><a name="p222154613543"></a><a name="p222154613543"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row1311652445414"><td class="cellrowborder" valign="top" width="16.77167716771677%" headers="mcps1.2.5.1.1 "><p id="p202220467541"><a name="p202220467541"></a><a name="p202220467541"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="43.23432343234324%" headers="mcps1.2.5.1.2 "><p id="p1022184612549"><a name="p1022184612549"></a><a name="p1022184612549"></a>根据全局故障信息决策恢复策略，并下发到MindIO，支持恢复策略回退，进程级重调度失败回退到Pod级别、Job级别重调度。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="18.13181318131813%" headers="mcps1.2.5.1.3 "><p id="p488619172591"><a name="p488619172591"></a><a name="p488619172591"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="21.862186218621858%" headers="mcps1.2.5.1.4 "><p id="p1211652412545"><a name="p1211652412545"></a><a name="p1211652412545"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row18952145365"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p72274605415"><a name="p72274605415"></a><a name="p72274605415"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p522104615410"><a name="p522104615410"></a><a name="p522104615410"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p11417425315"><a name="p11417425315"></a><a name="p11417425315"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>


#### 进程级在线恢复<a name="ZH-CN_TOPIC_0000002479386460"></a>

进程级在线恢复（Step级别重计算恢复）主要针对以下2种故障类型进行故障处理：

-   网络故障：当前仅支持以下两种场景。
    -   HCCS L1-L2端口或链路故障时，BGP切路后，若开启算子级在线恢复且执行失败后进行Step级重试，实现进程不退出的故障快速恢复；若关闭算子级在线恢复，则对训练进程进行Step级重试，实现进程不退出的故障快速恢复。
    -   RoCE到上级端口或链路故障，且开启算子级在线恢复并执行失败时，对训练进程进行Step级重试，实现进程不退出的故障快速恢复。

-   片上内存故障：片上内存上出现的不可纠正错误（如故障码0x80E01801），先隔离故障片上内存空间，然后对训练进程进行Step级重试，实现进程不退出的故障快速恢复。

在以上2种场景下，如果故障不能恢复，则回退至**重调度模式**。

相比于进程级别重调度，进程级在线恢复不会重调度故障进程，减少了大量进程间不同步的等待耗时。同时通过NPU卡间的参数面高速网络P2P传递CKPT信息，避免了CKPT保存和加载的耗时。

该故障处理模式默认关闭，若要开启请参考[（可选）配置组件](#可选配置组件)。

了解进程级在线恢复的关键配置步骤，请参见[配置进程级在线恢复](#配置进程级在线恢复)。

>[!NOTE] 说明 
>-   参数面传递CKPT信息依赖未故障卡中的全量优化器副本，如果不存在全量优化器副本，则回退为加载存储上的CKPT文件恢复参数。
>-   优化器副本依赖额外的显存占用，如果用户的显存较为紧张，可选择本地加载模式，无论是否存在优化器副本都直接加载存储上的CKPT文件恢复参数。

**使用约束<a name="zh-cn_topic_0000002003193196_section17145122992213"></a>**

-   使用进程级别在线恢复需要满足的版本配套关系如下。
    -   PyTorch版本为2.7.1。
    -   MindSpeed-LLM版本为2.3.0版本。

-   依赖于PyTorch的内存管理机制，仅在PYTORCH\_NO\_NPU\_MEMORY\_CACHING未配置时才能使用此功能。
-   针对部分片上内存故障场景无法生效，例如HCCL集合通信使用的内存地址故障，仍需通过进程级重调度或更上层的容错方案恢复。
-   针对MindSpeed-LLM、MindSpeed等模型或训练脚本中定义的全局变量发生故障的场景，详细处理策略请参见[FAQ](../faq.md#启用进程级在线恢复后报错there-is-unsafe-data-in-the-input-tensor恢复失败)。
-   与优雅容错不能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
-   MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
-   MindSpore场景下，需要在启动TaskD  Manager前设置export TASKD\_PROCESS\_ENABLE="on"。
-   请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。
-   不支持多模态模型。
-   不支持MC2开启场景。
-   不支持开启watchdog功能。
-   当故障发生在HCCL建链阶段时，会导致进程级在线恢复失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考[配置HCCL主动触发建链](#配置hccl主动触发建链)章节进行提前建链，防止故障出现在HCCL建链阶段。

**支持的产品型号及AI框架<a name="zh-cn_topic_0000002003193196_section108582044132214"></a>**

**表 1**  网络故障进程级在线恢复支持的产品和框架

<a name="zh-cn_topic_0000002003193196_table18104314924"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003193196_row81042144212"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002003193196_p51041814022"><a name="zh-cn_topic_0000002003193196_p51041814022"></a><a name="zh-cn_topic_0000002003193196_p51041814022"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002003193196_p91041414627"><a name="zh-cn_topic_0000002003193196_p91041414627"></a><a name="zh-cn_topic_0000002003193196_p91041414627"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002003193196_p11040145218"><a name="zh-cn_topic_0000002003193196_p11040145218"></a><a name="zh-cn_topic_0000002003193196_p11040145218"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003193196_row1910518141229"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003193196_p191051114524"><a name="zh-cn_topic_0000002003193196_p191051114524"></a><a name="zh-cn_topic_0000002003193196_p191051114524"></a><span id="zh-cn_topic_0000002003193196_ph19105814420"><a name="zh-cn_topic_0000002003193196_ph19105814420"></a><a name="zh-cn_topic_0000002003193196_ph19105814420"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul18927338231"></a><a name="ul18927338231"></a><ul id="ul18927338231"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul17506112910131"></a><a name="ul17506112910131"></a><ul id="ul17506112910131"><li><span id="ph135064298139"><a name="ph135064298139"></a><a name="ph135064298139"></a>MindSpore</span></li></ul>
<a name="ul7506132918139"></a><a name="ul7506132918139"></a><ul id="ul7506132918139"><li><span id="ph550610294136"><a name="ph550610294136"></a><a name="ph550610294136"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**表 2** 片上内存故障进程级在线恢复支持的产品和框架

<a name="table0630917154413"></a>
<table><thead align="left"><tr id="row13630161784418"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="p963031734417"><a name="p963031734417"></a><a name="p963031734417"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="p663151714415"><a name="p663151714415"></a><a name="p663151714415"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="p13631111710444"><a name="p13631111710444"></a><a name="p13631111710444"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="row5631517114410"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p166312178442"><a name="p166312178442"></a><a name="p166312178442"></a><span id="ph1463121734416"><a name="ph1463121734416"></a><a name="ph1463121734416"></a>Atlas A2 训练系列产品</span></p>
<p id="p12631191713449"><a name="p12631191713449"></a><a name="p12631191713449"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul0631181774417"></a><a name="ul0631181774417"></a><ul id="ul0631181774417"><li><span id="ph46319177449"><a name="ph46319177449"></a><a name="ph46319177449"></a>Atlas 800T A2 训练服务器</span></li><li><span id="ph1463131724413"><a name="ph1463131724413"></a><a name="ph1463131724413"></a>Atlas 900 A2 PoD 集群基础单元</span></li><li><span id="ph46311417154417"><a name="ph46311417154417"></a><a name="ph46311417154417"></a>Atlas 900 A2 PoDc 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul3631151714415"></a><a name="ul3631151714415"></a><ul id="ul3631151714415"><li><span id="ph36311817154419"><a name="ph36311817154419"></a><a name="ph36311817154419"></a>MindSpore</span></li></ul>
<a name="ul1263181794418"></a><a name="ul1263181794418"></a><ul id="ul1263181794418"><li><span id="ph1263191704413"><a name="ph1263191704413"></a><a name="ph1263191704413"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row16631181714416"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p563111714440"><a name="p563111714440"></a><a name="p563111714440"></a><span id="ph363111714444"><a name="ph363111714444"></a><a name="ph363111714444"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul1763161764415"></a><a name="ul1763161764415"></a><ul id="ul1763161764415"><li><span id="ph1963121720449"><a name="ph1963121720449"></a><a name="ph1963121720449"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph1363115172443"><a name="ph1363115172443"></a><a name="ph1363115172443"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul96311517144415"></a><a name="ul96311517144415"></a><ul id="ul96311517144415"><li><span id="ph96310177449"><a name="ph96310177449"></a><a name="ph96310177449"></a>MindSpore</span></li></ul>
<a name="ul7631141712447"></a><a name="ul7631141712447"></a><ul id="ul7631141712447"><li><span id="ph1563101734413"><a name="ph1563101734413"></a><a name="ph1563101734413"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**进程级在线恢复原理<a name="zh-cn_topic_0000002003193196_section961210366427"></a>**

训练过程中如果出现了片上内存故障或网络故障，将导致训练状态异常。进程级在线恢复首先通知所有训练进程停止当前训练，然后保留当前训练信息并修复故障。修复完成后，所有训练进程回退训练状态到当前上一个Step结束时，正常服务器通过参数面将CKPT传递到故障服务器上，完成参数恢复后重新执行当前Step，然后恢复训练任务。

**图 1**  进程级在线恢复原理<a name="fig37536398327"></a>  
![](../../figures/scheduling/进程级在线恢复原理.png "进程级在线恢复原理")

在以上原理图中，各个步骤的说明如下。

1.  设备出现片上内存故障或网络故障后，MindCluster在服务器上的检测组件上报故障信息到集群大脑ClusterD中。
2.  片上内存故障或网络故障被CANN软件感知，经训练框架上报给MindIO  Processor和MindIO  Controller。
3.  MindIO  Controller向集群大脑请求决策是否进行Step级别重计算恢复，集群大脑综合集群其他节点的健康状态给出决策。
4.  MindIO  Controller通知每个训练进程中的MindIO  Processor，调用训练框架停止任务、修复故障，保留通信域信息。
5.  正常服务器上的NPU通过参数面将CKPT传递到故障（已修复）服务器上，完成参数状态恢复后继续训练，重新启动当前Step计算。

**适配功能点<a name="section1446615300284"></a>**

在进程级在线恢复中，集群大脑根据故障信息识别网络故障和片上内存故障，下发对应恢复策略，支持恢复策略回退。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，在恢复时针对不同故障执行算子资源清理、UCE模型优化器重建、参数面在线修复、状态回滚，完成进程级在线恢复。

对于非MindSpeed-LLM、MindCluster平台用户，针对不同故障需在框架侧完成以下功能适配。

**表 3**  进程级在线恢复针对网络故障框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.61186118611861%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="36.72367236723672%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="17.981798179817982%" id="mcps1.2.5.1.3"><p id="p1912785111610"><a name="p1912785111610"></a><a name="p1912785111610"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="26.68266826682668%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row199614191876"><td class="cellrowborder" valign="top" width="18.61186118611861%" headers="mcps1.2.5.1.1 "><p id="p174797321974"><a name="p174797321974"></a><a name="p174797321974"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="36.72367236723672%" headers="mcps1.2.5.1.2 "><p id="p1847910326710"><a name="p1847910326710"></a><a name="p1847910326710"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="17.981798179817982%" headers="mcps1.2.5.1.3 "><p id="p12303135518715"><a name="p12303135518715"></a><a name="p12303135518715"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="26.68266826682668%" headers="mcps1.2.5.1.4 "><p id="p1878873515913"><a name="p1878873515913"></a><a name="p1878873515913"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row149661916713"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1947943212711"><a name="p1947943212711"></a><a name="p1947943212711"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p04796323710"><a name="p04796323710"></a><a name="p04796323710"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row1239411299541"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p94791332771"><a name="p94791332771"></a><a name="p94791332771"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p5479103211720"><a name="p5479103211720"></a><a name="p5479103211720"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row13395629115418"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17479113217716"><a name="p17479113217716"></a><a name="p17479113217716"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p10479532172"><a name="p10479532172"></a><a name="p10479532172"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row7395142913549"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1447912327718"><a name="p1447912327718"></a><a name="p1447912327718"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1847916321477"><a name="p1847916321477"></a><a name="p1847916321477"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row539519296541"><td class="cellrowborder" valign="top" width="18.61186118611861%" headers="mcps1.2.5.1.1 "><p id="p114808324711"><a name="p114808324711"></a><a name="p114808324711"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="36.72367236723672%" headers="mcps1.2.5.1.2 "><p id="p248011324715"><a name="p248011324715"></a><a name="p248011324715"></a>根据故障信息识别网路故障或片上内存故障，下发对应恢复策略，支持恢复策略回退。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="17.981798179817982%" headers="mcps1.2.5.1.3 "><p id="p16303135517718"><a name="p16303135517718"></a><a name="p16303135517718"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="26.68266826682668%" headers="mcps1.2.5.1.4 "><p id="p19472244965"><a name="p19472244965"></a><a name="p19472244965"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row7396029145419"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9480632573"><a name="p9480632573"></a><a name="p9480632573"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p84806321578"><a name="p84806321578"></a><a name="p84806321578"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p12472134412615"><a name="p12472134412615"></a><a name="p12472134412615"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

**表 4**  进程级在线恢复针对片上内存故障框架适配功能点

<a name="table14662336155516"></a>
<table><thead align="left"><tr id="row866213619553"><th class="cellrowborder" valign="top" width="17.119999999999997%" id="mcps1.2.5.1.1"><p id="p36629367550"><a name="p36629367550"></a><a name="p36629367550"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="38.769999999999996%" id="mcps1.2.5.1.2"><p id="p6662103635520"><a name="p6662103635520"></a><a name="p6662103635520"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="17.43%" id="mcps1.2.5.1.3"><p id="p1857674501116"><a name="p1857674501116"></a><a name="p1857674501116"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="26.68%" id="mcps1.2.5.1.4"><p id="p966243617552"><a name="p966243617552"></a><a name="p966243617552"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row19662436145518"><td class="cellrowborder" valign="top" width="17.119999999999997%" headers="mcps1.2.5.1.1 "><p id="p339173741211"><a name="p339173741211"></a><a name="p339173741211"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="38.769999999999996%" headers="mcps1.2.5.1.2 "><p id="p1739537151219"><a name="p1739537151219"></a><a name="p1739537151219"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="17.43%" headers="mcps1.2.5.1.3 "><p id="p9527145711216"><a name="p9527145711216"></a><a name="p9527145711216"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="26.68%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row566215364551"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p23923716123"><a name="p23923716123"></a><a name="p23923716123"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11397374123"><a name="p11397374123"></a><a name="p11397374123"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row06621936185512"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p639183714129"><a name="p639183714129"></a><a name="p639183714129"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15391437141215"><a name="p15391437141215"></a><a name="p15391437141215"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row2662133617558"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p73913781212"><a name="p73913781212"></a><a name="p73913781212"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1039113741210"><a name="p1039113741210"></a><a name="p1039113741210"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row066213685511"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p143953791220"><a name="p143953791220"></a><a name="p143953791220"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p203910376122"><a name="p203910376122"></a><a name="p203910376122"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row666243613555"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1396379125"><a name="p1396379125"></a><a name="p1396379125"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p143993761219"><a name="p143993761219"></a><a name="p143993761219"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row14662143645516"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p639113711122"><a name="p639113711122"></a><a name="p639113711122"></a>UCE模型优化器重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p83953718128"><a name="p83953718128"></a><a name="p83953718128"></a>通过回调函数完成故障卡模型优化器对象操作清理、重建操作。</p>
</td>
</tr>
<tr id="row43068171121"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p139737201218"><a name="p139737201218"></a><a name="p139737201218"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p539103711214"><a name="p539103711214"></a><a name="p539103711214"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row17307161715127"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p183923781214"><a name="p183923781214"></a><a name="p183923781214"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p114017372126"><a name="p114017372126"></a><a name="p114017372126"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row966233613558"><td class="cellrowborder" valign="top" width="17.119999999999997%" headers="mcps1.2.5.1.1 "><p id="p114023721211"><a name="p114023721211"></a><a name="p114023721211"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="38.769999999999996%" headers="mcps1.2.5.1.2 "><p id="p124083761213"><a name="p124083761213"></a><a name="p124083761213"></a>根据故障信息识别网路故障或片上内存故障，下发对应恢复策略，支持恢复策略回退。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="17.43%" headers="mcps1.2.5.1.3 "><p id="p65272572124"><a name="p65272572124"></a><a name="p65272572124"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="26.68%" headers="mcps1.2.5.1.4 "><p id="p14571125414116"><a name="p14571125414116"></a><a name="p14571125414116"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row16621936105516"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17407378128"><a name="p17407378128"></a><a name="p17407378128"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p140173701218"><a name="p140173701218"></a><a name="p140173701218"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p957195451114"><a name="p957195451114"></a><a name="p957195451114"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>


#### 算子级在线恢复<a name="ZH-CN_TOPIC_0000002479386484"></a>

Atlas A3 训练系列产品支持在发生参数面网络故障时，HCCL会执行通信算子重传。在故障进程不退出的情况下，算子级在线恢复可容忍更长时间的网络异常，训练任务不中断。

若网络故障的算子级在线恢复（HCCL通信算子重执行）执行失败，则回退至进程级在线恢复

了解算子级在线恢复的关键配置步骤，请参见[配置算子级在线恢复](#配置算子级在线恢复)。

>[!NOTE] 说明 
>HCCL（Huawei Collective Communication Library，华为集合通信库）是华为专为昇腾（Ascend）AI处理器设计的分布式通信库，旨在优化多设备（如NPU/GPU）间的高效协作，以加速深度学习模型的分布式训练，适用于需要大规模算力的AI场景。在分布式训练中，HCCL负责协调多个昇腾处理器之间的数据同步（如梯度聚合、参数更新），减少通信开销，提升训练效率。

**使用场景<a name="section4314241154917"></a>**

当前支持在以下2种故障场景下使用算子级在线恢复功能。

-   对于芯片网络相关故障，当算子重传成功时，Volcano会将任务作为亚健康任务处理。当算子重传失败时，Volcano触发重调度处理。
-   对于灵衢总线设备相关故障，HCCL执行算子级在线恢复后，Volcano会将任务作为亚健康任务处理。

**使用约束<a name="section1915719315116"></a>**

-   本特性不支持MC2开启场景。
-   不支持开启watchdog功能。

**算子级在线恢复支持的产品和框架<a name="section996215473410"></a>**

**表 1**  支持的产品和框架

<a name="table11647101624213"></a>
<table><thead align="left"><tr id="row17647111614214"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="p1664831610428"><a name="p1664831610428"></a><a name="p1664831610428"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="p1664816167422"><a name="p1664816167422"></a><a name="p1664816167422"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="p17648141664214"><a name="p17648141664214"></a><a name="p17648141664214"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="row14649101615422"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p8649101644216"><a name="p8649101644216"></a><a name="p8649101644216"></a><span id="ph96491216144210"><a name="ph96491216144210"></a><a name="ph96491216144210"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><p id="p10649816134219"><a name="p10649816134219"></a><a name="p10649816134219"></a><span id="ph264911612426"><a name="ph264911612426"></a><a name="ph264911612426"></a>Atlas 900 A3 SuperPoD 集群算力系统</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><p id="p1664981614218"><a name="p1664981614218"></a><a name="p1664981614218"></a>-</p>
</td>
</tr>
</tbody>
</table>

**算子级在线恢复原理<a name="section41453583611"></a>**

**图 1**  原理图<a name="fig151851746103612"></a>  
![](../../figures/scheduling/原理图-8.png "原理图-8")

在以上原理图中，各个步骤的说明如下。

1.  训练过程中，发生HCCS网络平面LinkDown故障或RoCE网络平面LinkDown故障。
2.  CANN检测到网络故障，当前算子终止后，进行网络链路恢复（HCCS网络平面进行BGP切路，RoCE网络平面进行借轨通信），通信链路恢复后进行网络算子重执行。
3.  算子重执行成功后，恢复训练迭代。


#### 借轨通信任务暂停与回切<a name="ZH-CN_TOPIC_0000002479226530"></a>

Atlas A3 训练系列产品场景下，MindCluster集群调度组件提供训练任务借轨通信的暂停与回切功能。即在训练过程中，使用主动借轨回切接口，可自由切换NPU芯片使用的RoCE网口。

使用借轨回切功能时，NPU芯片的组网关系可参考《Ascend Training Solution 25.1.RC1 组网指南（Atlas A3训练产品）》中的“网络平面介绍 \> 参数面网络 \> 端口对接策略”章节。

了解借轨通信任务暂停与回切功能的详细配置方法，请参见[配置借轨通信任务暂停与回切](#配置借轨通信任务暂停与回切)。

>[!NOTE] 说明 
>-   调用[借轨回切接口](../api/clusterd.md#借轨回切接口)执行借轨回切动作前，请先了解NPU芯片组网关系，保证目标NPU的网络链路正常，如果目标NPU为linkdown状态会导致操作失败。
>-   以上述组网指南中的接口对接关系为例，对于以下几种情况，调用SwitchNicTrack接口时，指定的dev与op如下：
>    1.  若将device0，device8从QDD8借轨切到QDD7，传参dev为\[device0 ，device8\]，op为\[true，true\]
>    2.  若将device0，device8从QDD7回切到QDD8，传参dev为\[device0 ，device8\]，op为\[false，false\]
>    3.  如果单独将device0从QDD8的PortA借轨切到QDD7的PortA，传参dev为\[device0\]，op为\[true\]
>    4.  如果单独将device0从QDD7的PortA回切到QDD8的PortA，传参dev为\[device0\]，op为\[false\]
>    5.  如果将Leaf1下的全部device借轨切到Leaf2下，传参dev为\[device0，device8，device2，device10，device4，device12，device6，device14 \]，op为\[true，true，true，true，true，true，true，true\]
>    6.  如果将Leaf2下的全部device回切到Leaf1下，传参dev为\[device0，device8，device2，device10，device4，device12，device6，device14 \]，op为\[false，false，false，false，false，false，false，false\]
>    **图 1**  接口对接关系<a name="fig111354543222"></a>  
>    ![](../../figures/scheduling/接口对接关系.png "接口对接关系")

**使用场景<a name="section14336140104818"></a>**

当前支持在以下2种场景下使用借轨通信任务暂停与回切功能。

-   交换机升级场景：人工触发借轨后升级交换机，再回切。
-   故障处理场景：发生借轨的故障端口在修复完成后，再做人工回切。

**使用约束<a name="section620412554441"></a>**

-   请在训练正常迭代后，再进行借轨或回切指令的下发。
-   确保已开启进程级恢复相关功能特性。
-   对于MindSpore训练框架，需要在启动TaskD  Manager前设置export TASKD\_PROCESS\_ENABLE="on"。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002098609234_section4771115416256"></a>**

**表 1**  支持的产品和框架

<a name="zh-cn_topic_0000002098609234_table1526819106465"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002098609234_row22681310134611"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002098609234_p137295354447"><a name="zh-cn_topic_0000002098609234_p137295354447"></a><a name="zh-cn_topic_0000002098609234_p137295354447"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002098609234_p1172993554412"><a name="zh-cn_topic_0000002098609234_p1172993554412"></a><a name="zh-cn_topic_0000002098609234_p1172993554412"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002098609234_p97299357449"><a name="zh-cn_topic_0000002098609234_p97299357449"></a><a name="zh-cn_topic_0000002098609234_p97299357449"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002098609234_row71691214122315"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p112681620231"><a name="zh-cn_topic_0000002098609234_p112681620231"></a><a name="zh-cn_topic_0000002098609234_p112681620231"></a><span id="zh-cn_topic_0000002098609234_ph9126121617231"><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul13725194132419"></a><a name="ul13725194132419"></a><ul id="ul13725194132419"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul7583132019396"></a><a name="ul7583132019396"></a><ul id="ul7583132019396"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li></ul>
<a name="ul75831320173911"></a><a name="ul75831320173911"></a><ul id="ul75831320173911"><li><span id="ph13583142013394"><a name="ph13583142013394"></a><a name="ph13583142013394"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**借轨通信任务暂停与回切原理<a name="section56986212179"></a>**

**图 2**  原理图<a name="fig9336113210132"></a>  
![](../../figures/scheduling/原理图-9.png "原理图-9")

在以上原理图中，各个步骤的说明如下。

1.  AI平台集成ClusterD，调用ClusterD的gRPC接口下发切换操作，指定需要切换的NPU卡。
2.  ClusterD通知MindIO暂停训练。
3.  TaskD Manager通知所有TaskD Worker调用训练框架接口执行切换操作。
4.  训练框架按照通信域逐一调用CANN接口执行切换操作。
5.  ClusterD判断所有NPU卡的切换操作完成后，再由TaskD通知MindIO在切换完成后继续执行下一个Step训练。

**适配功能点<a name="section1446615300284"></a>**

在借轨通信任务暂停与回切中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。通过主动调用优雅暂停机制，完成当前卡上任务暂停和任务切换。集群大脑需提供对外接口，接受切换指令并管理借轨通信流程。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table1995514113610)的功能适配。

**表 2**  借轨通信任务暂停与回切框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.87%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="43.419999999999995%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="14.719999999999999%" id="mcps1.2.5.1.3"><p id="p10978953142414"><a name="p10978953142414"></a><a name="p10978953142414"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="22.99%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.87%" headers="mcps1.2.5.1.1 "><p id="p1987424102519"><a name="p1987424102519"></a><a name="p1987424102519"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="43.419999999999995%" headers="mcps1.2.5.1.2 "><p id="p14351731182511"><a name="p14351731182511"></a><a name="p14351731182511"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.719999999999999%" headers="mcps1.2.5.1.3 "><p id="p922524114255"><a name="p922524114255"></a><a name="p922524114255"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="22.99%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9871924102517"><a name="p9871924102517"></a><a name="p9871924102517"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p16810326255"><a name="p16810326255"></a><a name="p16810326255"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p12878242257"><a name="p12878242257"></a><a name="p12878242257"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p687122472518"><a name="p687122472518"></a><a name="p687122472518"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row1297881015253"><td class="cellrowborder" valign="top" width="18.87%" headers="mcps1.2.5.1.1 "><p id="p168711249252"><a name="p168711249252"></a><a name="p168711249252"></a>借轨切换过程管理</p>
</td>
<td class="cellrowborder" valign="top" width="43.419999999999995%" headers="mcps1.2.5.1.2 "><p id="p168762416258"><a name="p168762416258"></a><a name="p168762416258"></a>提供借轨切换请求下发能力，控制训练进程暂停与重启。</p>
</td>
<td class="cellrowborder" valign="top" width="14.719999999999999%" headers="mcps1.2.5.1.3 "><p id="p10461144315257"><a name="p10461144315257"></a><a name="p10461144315257"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="22.99%" headers="mcps1.2.5.1.4 "><p id="p10979110172511"><a name="p10979110172511"></a><a name="p10979110172511"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v7.3.0/component/clusterd/pkg/application/recover/om_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>


#### （可选）优雅容错<a name="ZH-CN_TOPIC_0000002479226564"></a>

>[!NOTE] 说明 
>该功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。

当用户进行没有备用资源的训练任务，或者期望设备自动恢复时，可以选择使用**优雅容错**功能。即当训练时的芯片设备出现故障后，系统将尝试对故障芯片进行自动恢复，如果可以恢复，则在保持Pod运行状态下，将任务原地拉起继续训练，不能恢复则回退至**重调度模式**。

优雅容错功能无需进行资源调度，即可自动将故障设备恢复。但是它无法降低训练初始化中的恢复时间，通常情况下，优雅容错所需恢复时间大于进程级重调度和进程级在线恢复功能。

了解优雅容错的关键配置步骤，请参见[配置优雅容错](#配置优雅容错)。

**使用约束<a name="zh-cn_topic_0000002098609234_section1137610139461"></a>**

-   当前只支持芯片故障使用优雅容错功能。
-   优雅容错功能与进程级别重调度、进程级在线恢复功能不能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002098609234_section4771115416256"></a>**

**表 1**  优雅容错支持的产品和框架

<a name="zh-cn_topic_0000002098609234_table1526819106465"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002098609234_row22681310134611"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002098609234_p137295354447"><a name="zh-cn_topic_0000002098609234_p137295354447"></a><a name="zh-cn_topic_0000002098609234_p137295354447"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002098609234_p1172993554412"><a name="zh-cn_topic_0000002098609234_p1172993554412"></a><a name="zh-cn_topic_0000002098609234_p1172993554412"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002098609234_p97299357449"><a name="zh-cn_topic_0000002098609234_p97299357449"></a><a name="zh-cn_topic_0000002098609234_p97299357449"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002098609234_row17268131014613"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p889791444417"><a name="zh-cn_topic_0000002098609234_p889791444417"></a><a name="zh-cn_topic_0000002098609234_p889791444417"></a><span id="zh-cn_topic_0000002098609234_ph289810142442"><a name="zh-cn_topic_0000002098609234_ph289810142442"></a><a name="zh-cn_topic_0000002098609234_ph289810142442"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039353153_ul17412295261"></a><a name="zh-cn_topic_0000002039353153_ul17412295261"></a><ul id="zh-cn_topic_0000002039353153_ul17412295261"><li><span id="ph1638757114220"><a name="ph1638757114220"></a><a name="ph1638757114220"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002098609234_ul1381333331316"></a><a name="zh-cn_topic_0000002098609234_ul1381333331316"></a><ul id="zh-cn_topic_0000002098609234_ul1381333331316"><li><span id="zh-cn_topic_0000002098609234_ph1246144904420"><a name="zh-cn_topic_0000002098609234_ph1246144904420"></a><a name="zh-cn_topic_0000002098609234_ph1246144904420"></a>MindSpore</span></li></ul>
<a name="zh-cn_topic_0000002098609234_ul10570112811135"></a><a name="zh-cn_topic_0000002098609234_ul10570112811135"></a><ul id="zh-cn_topic_0000002098609234_ul10570112811135"><li><span id="zh-cn_topic_0000002098609234_ph473115306133"><a name="zh-cn_topic_0000002098609234_ph473115306133"></a><a name="zh-cn_topic_0000002098609234_ph473115306133"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002098609234_row181221631185611"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p128991832165620"><a name="zh-cn_topic_0000002098609234_p128991832165620"></a><a name="zh-cn_topic_0000002098609234_p128991832165620"></a><span id="zh-cn_topic_0000002098609234_ph13899123211565"><a name="zh-cn_topic_0000002098609234_ph13899123211565"></a><a name="zh-cn_topic_0000002098609234_ph13899123211565"></a>Atlas A2 训练系列产品</span></p>
<p id="p96481557151918"><a name="p96481557151918"></a><a name="p96481557151918"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002098609234_ul13899193245613"></a><a name="zh-cn_topic_0000002098609234_ul13899193245613"></a><ul id="zh-cn_topic_0000002098609234_ul13899193245613"><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></li><li><span id="zh-cn_topic_0000002098609234_ph189001332105615"><a name="zh-cn_topic_0000002098609234_ph189001332105615"></a><a name="zh-cn_topic_0000002098609234_ph189001332105615"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002098609234_ul664419915495"></a><a name="zh-cn_topic_0000002098609234_ul664419915495"></a><ul id="zh-cn_topic_0000002098609234_ul664419915495"><li><span id="zh-cn_topic_0000002098609234_ph146444924919"><a name="zh-cn_topic_0000002098609234_ph146444924919"></a><a name="zh-cn_topic_0000002098609234_ph146444924919"></a>MindSpore</span></li></ul>
<a name="zh-cn_topic_0000002098609234_ul36445934915"></a><a name="zh-cn_topic_0000002098609234_ul36445934915"></a><ul id="zh-cn_topic_0000002098609234_ul36445934915"><li><span id="zh-cn_topic_0000002098609234_ph364489174917"><a name="zh-cn_topic_0000002098609234_ph364489174917"></a><a name="zh-cn_topic_0000002098609234_ph364489174917"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002098609234_row71691214122315"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p112681620231"><a name="zh-cn_topic_0000002098609234_p112681620231"></a><a name="zh-cn_topic_0000002098609234_p112681620231"></a><span id="zh-cn_topic_0000002098609234_ph9126121617231"><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul13725194132419"></a><a name="ul13725194132419"></a><ul id="ul13725194132419"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul7583132019396"></a><a name="ul7583132019396"></a><ul id="ul7583132019396"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li></ul>
<a name="ul75831320173911"></a><a name="ul75831320173911"></a><ul id="ul75831320173911"><li><span id="ph13583142013394"><a name="ph13583142013394"></a><a name="ph13583142013394"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**优雅容错原理<a name="zh-cn_topic_0000002098609234_section882584011262"></a>**

在节点或芯片故障处理过程中，若使用重调度模式，需要运维人员手动恢复设备。若任务恢复不及时可能导致训练集群中出现大量散点故障，降低集群算力利用率。因此，断点续训在**重调度模式**上增加了**优雅容错**功能，用于优化NPU芯片的部分故障容错能力。

NPU芯片故障中的部分故障可以通过退出芯片上的训练进程以及热复位芯片来恢复，优雅容错功能即针对这部分故障进行恢复处理，不需要重调度任务。

Ascend Device Plugin负责故障的上报以及设备的恢复，管理进程（PyTorch场景下为Elastic Agent组件，MindSpore场景下为TaskD组件）根据Ascend Device Plugin上报的信息进行训练进程的停止与重新拉起，完成故障恢复（不能恢复则回退至**重调度模式**）。集成优雅容错模式需要在业务容器中添加管理进程，管理进程需要具备故障感知、停止训练任务和重启训练任务等能力。

优雅容错模式直接将故障上报到业务容器内的管理进程中（通常通过挂载文件的方式），容器内的管理进程读取故障文件信息获取到故障信息，获取故障信息的流程如[图1](#zh-cn_topic_0000002098609234_fig135111361314)所示。

**图 1**  获取故障信息<a name="zh-cn_topic_0000002098609234_fig135111361314"></a>  
![](../../figures/scheduling/获取故障信息.png "获取故障信息")

优雅容错模式将故障区分为以下四类，**无需处理**、**重新执行业务**、**需要复位芯片**和**需要重调度**，对于每类故障的处理如[图2](#zh-cn_topic_0000002098609234_fig12620181591012)所示。

**图 2**  优雅容错故障处理流程<a name="zh-cn_topic_0000002098609234_fig12620181591012"></a>  
![](../../figures/scheduling/优雅容错故障处理流程.png "优雅容错故障处理流程")


#### 在线压测<a name="ZH-CN_TOPIC_0000002479226572"></a>

MindCluster支持训练在线压测特性，即在训练过程中可以调用在线压测接口，暂停指定训练任务，对任务使用的节点进行硬件P2P或AIC压力测试。若不存在故障则恢复训练；若存在故障则隔离故障节点，触发断点续训。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

-   对于PyTorch训练框架，需配合MindSpeed-LLM  2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
-   对于MindSpore训练框架，需配合MindFormers master版本使用，版本配套请参见[MindSpore MindFormers](https://gitee.com/mindspore/mindformers/tree/r0.3/)。
-   请在训练正常迭代后，再进行在线压测指令的下发。
-   确保已开启进程级恢复相关功能特性。
-   压测过程中不支持重启ClusterD，如果ClusterD异常重启，需要重启训练下发压测任务。
-   压测过程中，需要关闭热复位功能。
-   P2P压测需确保device侧有10G以上的空闲内存。
-   需要在节点增加nodeDEnable=on标签，保证出现压测的节点可以隔离。
-   对于MindSpore训练框架，需要在启动TaskD  Manager前设置export TASKD\_PROCESS\_ENABLE="on"。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 1**  在线压测支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="25.172517251725168%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="43.834383438343835%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="30.993099309930994%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="25.172517251725168%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p773278122616"><a name="p773278122616"></a><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="43.834383438343835%" headers="mcps1.2.4.1.2 "><p id="p17354133423610"><a name="p17354133423610"></a><a name="p17354133423610"></a><span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.993099309930994%" headers="mcps1.2.4.1.3 "><a name="ul15879359132214"></a><a name="ul15879359132214"></a><ul id="ul15879359132214"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li><li><span id="ph19425111582712"><a name="ph19425111582712"></a><a name="ph19425111582712"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="25.172517251725168%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="43.834383438343835%" headers="mcps1.2.4.1.2 "><p id="p4897194703620"><a name="p4897194703620"></a><a name="p4897194703620"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.993099309930994%" headers="mcps1.2.4.1.3 "><a name="ul13821123132320"></a><a name="ul13821123132320"></a><ul id="ul13821123132320"><li><span id="ph19127156230"><a name="ph19127156230"></a><a name="ph19127156230"></a>MindSpore</span></li><li><span id="ph310231710274"><a name="ph310231710274"></a><a name="ph310231710274"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**在线压测原理<a name="section56986212179"></a>**

**图 1**  原理图<a name="fig9336113210132"></a>  
![](../../figures/scheduling/原理图-10.png "原理图-10")

在以上原理图中，各个步骤的说明如下。

1.  AI平台集成ClusterD，调用ClusterD的gRPC接口下发压测操作，指定需要压测的节点。
2.  ClusterD通知MindIO暂停训练。
3.  TaskD Manager通知指定TaskD Worker调用训练框架接口执行压测操作。
4.  训练框架调用指定NPU卡上的CANN接口执行压测操作。
5.  ClusterD判断指定NPU卡的压测操作完成后，再由TaskD通知MindIO在压测完成后继续执行下一个Step训练。

**适配功能点<a name="section1446615300284"></a>**

在在线压测中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。通过主动调用优雅暂停机制，完成当前卡上任务暂停，暂停后进行硬件压力测试，测试完成后继续训练。集群大脑需提供对外接口，接受压测指令并管理压测流程。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table1995514113610)的功能适配。

**表 2**  在线压测框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.98%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="39.26%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.01%" id="mcps1.2.5.1.3"><p id="p106021527183014"><a name="p106021527183014"></a><a name="p106021527183014"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="23.75%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.98%" headers="mcps1.2.5.1.1 "><p id="p0609650313"><a name="p0609650313"></a><a name="p0609650313"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="39.26%" headers="mcps1.2.5.1.2 "><p id="p195191085319"><a name="p195191085319"></a><a name="p195191085319"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="18.01%" headers="mcps1.2.5.1.3 "><p id="p1855311819317"><a name="p1855311819317"></a><a name="p1855311819317"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.75%" headers="mcps1.2.5.1.4 "><p id="p10701822403"><a name="p10701822403"></a><a name="p10701822403"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1960918515317"><a name="p1960918515317"></a><a name="p1960918515317"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p156091533120"><a name="p156091533120"></a><a name="p156091533120"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p76093513110"><a name="p76093513110"></a><a name="p76093513110"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p136091519311"><a name="p136091519311"></a><a name="p136091519311"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row46026594305"><td class="cellrowborder" valign="top" width="18.98%" headers="mcps1.2.5.1.1 "><p id="p26091514318"><a name="p26091514318"></a><a name="p26091514318"></a>在线压测过程管理</p>
</td>
<td class="cellrowborder" valign="top" width="39.26%" headers="mcps1.2.5.1.2 "><p id="p14609155183114"><a name="p14609155183114"></a><a name="p14609155183114"></a>提供在线压测请求下发能力，控制训练进程暂停与恢复。</p>
</td>
<td class="cellrowborder" valign="top" width="18.01%" headers="mcps1.2.5.1.3 "><p id="p6553121803118"><a name="p6553121803118"></a><a name="p6553121803118"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="23.75%" headers="mcps1.2.5.1.4 "><p id="p1660265933015"><a name="p1660265933015"></a><a name="p1660265933015"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v7.3.0/component/clusterd/pkg/application/recover/om_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>


#### 亚健康热切<a name="ZH-CN_TOPIC_0000002479386544"></a>

训练任务配置为亚健康热切策略（hotSwitch）后，当发生亚健康故障时，拉起备份节点后暂停训练进程，再使用备份节点重新拉起训练任务。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

-   对于PyTorch训练框架，需配合MindSpeed-LLM 2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
-   对于MindSpore训练框架，需配合MindFormers master版本使用，版本配套请参见[MindSpore MindFormers](https://gitee.com/mindspore/mindformers/tree/r0.3/)。
-   只支持PyTorch单算子模式、基于Megatron框架的模型以及acjob类型训练任务。
-   MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
-   不支持多模态模型。
-   不支持开启watchdog功能。
-   训练任务未出迭代时触发热切，可能会造成MindIO阻塞，最后触发Job级别重调度。
-   当训练任务的annotation中hccl/rankIndex字段为0的Pod发生亚健康故障时，不支持触发亚健康热切。
-   以下异常情况会回退至Job级别重调度，且任务亚健康处理策略降级为ignore，不再处理亚健康故障：
    -   备份Pod拉起后，训练暂停失败。
    -   备份Pod拉起后，MindCluster等待上报训练暂停状态超时（15分钟）。
    -   备份Pod运行失败。
    -   原Pod删除后，训练恢复失败。
    -   原Pod删除后，MindCluster等待上报训练恢复状态超时（15分钟）。

-   配置亚健康热切策略后，会自动增加进程级恢复开关，若发生非亚健康故障，将触发进程级恢复流程。
-   无备节点场景下，无法完成热切流程，任务亚健康处理策略降级为ignore，不再处理亚健康故障。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 1**  亚健康热切支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="25.172517251725175%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="59.82598259825983%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="15.001500150015001%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="25.172517251725175%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p773278122616"><a name="p773278122616"></a><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="59.82598259825983%" headers="mcps1.2.4.1.2 "><p id="p3799611168"><a name="p3799611168"></a><a name="p3799611168"></a><span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.001500150015001%" headers="mcps1.2.4.1.3 "><a name="ul15879359132214"></a><a name="ul15879359132214"></a><ul id="ul15879359132214"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li><li><span id="ph19425111582712"><a name="ph19425111582712"></a><a name="ph19425111582712"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="25.172517251725175%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="59.82598259825983%" headers="mcps1.2.4.1.2 "><p id="p13693112166"><a name="p13693112166"></a><a name="p13693112166"></a><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.001500150015001%" headers="mcps1.2.4.1.3 "><a name="ul6531100274"></a><a name="ul6531100274"></a><ul id="ul6531100274"><li><span id="ph1053216019718"><a name="ph1053216019718"></a><a name="ph1053216019718"></a>MindSpore</span></li><li><span id="ph35321001570"><a name="ph35321001570"></a><a name="ph35321001570"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**亚健康热切原理<a name="zh-cn_topic_0000002039194017_section57901137171110"></a>**

**图 1**  原理图<a name="fig1770171514241"></a>  
![](../../figures/scheduling/原理图-11.png "原理图-11")

在以上原理图中，各个步骤的说明如下。

1.  ClusterD通过Ascend Device Plugin感知到亚健康故障。
2.  ClusterD根据配置策略决策是否进行亚健康热切恢复。
3.  ClusterD通知Ascend Operator拉起备份Pod。
4.  Volcano调度备份Pod。
5.  备份Pod中创建新的MindIO Processor，MindIO Processor向MindIO Controller发起注册。
6.  MindIO Controller下发训练暂停通知。
7.  MindIO Controller通知ClusterD训练暂停。
8.  ClusterD通知Volcano删除故障Pod。
9.  ClusterD通知MindIO恢复训练。

**适配功能点<a name="section1446615300284"></a>**

在亚健康热切中，集群大脑根据亚健康故障信息，为故障Pod设置注解，拉起并调度备份Pod，通知热切策略到MindIO，训练切换到备份Pod后恢复训练。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。在异常发生时，通过异常捕获装饰器捕获故障模式。在新节点启动后，正常节点暂停训练，之后重建通信域，完成新节点参数面恢复，训练状态完成后完成节点热切换。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table1995514113610)的功能适配。

**表 2**  亚健康热切框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.200000000000003%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="39.330000000000005%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="19.670000000000005%" id="mcps1.2.5.1.3"><p id="p237216122367"><a name="p237216122367"></a><a name="p237216122367"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="22.800000000000004%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.200000000000003%" headers="mcps1.2.5.1.1 "><p id="p1698525618364"><a name="p1698525618364"></a><a name="p1698525618364"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="39.330000000000005%" headers="mcps1.2.5.1.2 "><p id="p117503011375"><a name="p117503011375"></a><a name="p117503011375"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="19.670000000000005%" headers="mcps1.2.5.1.3 "><p id="p444112643720"><a name="p444112643720"></a><a name="p444112643720"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="22.800000000000004%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1598625612366"><a name="p1598625612366"></a><a name="p1598625612366"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2986125603612"><a name="p2986125603612"></a><a name="p2986125603612"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1798635633611"><a name="p1798635633611"></a><a name="p1798635633611"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p898615619362"><a name="p898615619362"></a><a name="p898615619362"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6986115693618"><a name="p6986115693618"></a><a name="p6986115693618"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p39861056113613"><a name="p39861056113613"></a><a name="p39861056113613"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p99861756103611"><a name="p99861756103611"></a><a name="p99861756103611"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p19986125623612"><a name="p19986125623612"></a><a name="p19986125623612"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p698615618367"><a name="p698615618367"></a><a name="p698615618367"></a>节点重启及通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11986856193616"><a name="p11986856193616"></a><a name="p11986856193616"></a>通过注册重建回调实现健康节点与故障节点重建通信域。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10986256103613"><a name="p10986256103613"></a><a name="p10986256103613"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p198635693613"><a name="p198635693613"></a><a name="p198635693613"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row876023415918"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p49861556183618"><a name="p49861556183618"></a><a name="p49861556183618"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1398655643610"><a name="p1398655643610"></a><a name="p1398655643610"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row17605341596"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11986656153611"><a name="p11986656153611"></a><a name="p11986656153611"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p199862056113617"><a name="p199862056113617"></a><a name="p199862056113617"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row144412445361"><td class="cellrowborder" valign="top" width="18.200000000000003%" headers="mcps1.2.5.1.1 "><p id="p129861056183611"><a name="p129861056183611"></a><a name="p129861056183611"></a>热切流程控制</p>
</td>
<td class="cellrowborder" valign="top" width="39.330000000000005%" headers="mcps1.2.5.1.2 "><p id="p14986125653610"><a name="p14986125653610"></a><a name="p14986125653610"></a>管理热切恢复流程，通过设置注解方式管理备份Pod和故障Pod。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="19.670000000000005%" headers="mcps1.2.5.1.3 "><p id="p1045122693710"><a name="p1045122693710"></a><a name="p1045122693710"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="22.800000000000004%" headers="mcps1.2.5.1.4 "><p id="p64451744113612"><a name="p64451744113612"></a><a name="p64451744113612"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v7.3.0/component/clusterd/pkg/application/recover/hot_switch_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row14716101112393"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1371681114396"><a name="p1371681114396"></a><a name="p1371681114396"></a>Pod创建删除</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p071681117390"><a name="p071681117390"></a><a name="p071681117390"></a>通过识别特定注解删除和创建Pod。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p071621117393"><a name="p071621117393"></a><a name="p071621117393"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v7.3.0/component/ascend-operator/pkg/controllers/v1/ascendjob_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>


#### 弹性训练<a name="ZH-CN_TOPIC_0000002479226542"></a>

当出现硬件故障，且K8s集群中无可用备份资源时，MindCluster会先按照数据并行域缩掉部分节点继续训练，当集群中有可用空闲资源时，再触发扩容恢复原有规模训练。相比于进程级别重调度，解决了集群中无可用备份资源被重调度的问题。

**使用约束<a name="zh-cn_topic_0000002039353153_section514611624316"></a>**

-   仅支持PyTorch配合MindSpeed-LLM 2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
-   仅支持acjob类型训练任务。
-   依赖于MindIO的优化器副本，需要存在全量优化器副本，故需要安装MindIO和TaskD配合使用。
-   不能和优雅容错功能同时开启。
-   当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不支持触发弹性训练。
-   不支持多模态模型。
-   不支持开启watchdog功能。
-   由于弹性训练会额外创建新的通信组，因此可能会导致片上内存占用增加。

    增加内存大小计算公式：增加内存最大值（MB） = HCCL\_BUFFSIZE \* 2 \* 9，其中，HCCL\_BUFFSIZE默认为200MB，HCCL\_BUFFSIZE的说明详细请参见[CANN环境变量参考](https://www.hiascend.com/document/detail/zh/canncommercial/82RC1/maintenref/envvar/envref_07_0080.html)。

更多使用约束可参考[MindSpeed-LLM弹性训练功能使用约束](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/docs/pytorch/features/high_availability.md)。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039353153_section136131584164"></a>**

**表 1**  弹性训练支持的产品和框架

<a name="zh-cn_topic_0000002039353153_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039353153_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039353153_p199171819164417"><a name="zh-cn_topic_0000002039353153_p199171819164417"></a><a name="zh-cn_topic_0000002039353153_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="66.2966296629663%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039353153_p2917819114420"><a name="zh-cn_topic_0000002039353153_p2917819114420"></a><a name="zh-cn_topic_0000002039353153_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="13.24132413241324%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039353153_p27578257424"><a name="zh-cn_topic_0000002039353153_p27578257424"></a><a name="zh-cn_topic_0000002039353153_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039353153_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p153913472453"><a name="zh-cn_topic_0000002039353153_p153913472453"></a><a name="zh-cn_topic_0000002039353153_p153913472453"></a><span id="zh-cn_topic_0000002039353153_ph151431757142112"><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p737515258512"><a name="p737515258512"></a><a name="p737515258512"></a></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><p id="p697681955215"><a name="p697681955215"></a><a name="p697681955215"></a><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><p id="p139316519435"><a name="p139316519435"></a><a name="p139316519435"></a><span id="zh-cn_topic_0000002039353153_ph2093210246488"><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a>PyTorch</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039353153_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p18222246142212"><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><span id="zh-cn_topic_0000002039353153_ph18411121792018"><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><p id="p1711620216528"><a name="p1711620216528"></a><a name="p1711620216528"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><p id="p16887149174313"><a name="p16887149174313"></a><a name="p16887149174313"></a><span id="ph99469109139"><a name="ph99469109139"></a><a name="ph99469109139"></a>PyTorch</span></p>
</td>
</tr>
</tbody>
</table>

**弹性训练原理<a name="section3841210162013"></a>**

**图 1**  原理图<a name="fig130013397201"></a>  
![](../../figures/scheduling/原理图-12.png "原理图-12")

以上示意图仅以缩容1个DP域为例，实际弹性训练过程中可能会一次缩容多个DP域。图中每个方格代表一个rank。

1.  按照TP（Tensor Parallelism，张量并行）、PP（Pipeline Parallelism，流水线并行）、DP（Data Parallelism，数据并行）正常进行分布式训练。
2.  训练到某一时刻，若某张卡发生故障，且集群中无更多空闲资源可被调度进行断点续训，则按照DP域缩容，即缩容1个DP域对应的Pod（可能包含多个Pod）后继续训练。
3.  缩容训练到某一时刻，集群中有空闲资源时，缩容的Pod会被重新调度，扩容恢复到原有规模继续训练。

**图 2**  流程图<a name="fig7783192415293"></a>  
![](../../figures/scheduling/流程图.png "流程图")

在以上流程图中，各个步骤的说明如下。

1.  设备出现硬件故障后，MindCluster在服务器上的检测组件上报故障信息到ClusterD中，软件故障由容器内MindIO Controller感知并上报到ClusterD。
2.  ClusterD将故障服务器上的任务容器销毁。
3.  若没有备份节点调度新容器，ClusterD通知Master节点上的MindIO Controller进行缩容训练。
4.  MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用PTA停止训练进程，清理正常节点的资源。
5.  MindIO Controller通知正常的训练进程中的MindIO Processor执行通信组重建等缩容流程，进行缩容训练。
6.  检测到缩容时删除的Pod重调度成功。
7.  ClusterD通过TaskD  Manager通知MindIO Controller执行扩容。
8.  MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用PTA停止训练进程，清理正常节点的资源。
9.  各个进程进行集合通信建链。
10. 正常服务器上的NPU通过参数面将CKPT传递到备用服务器上，完成参数状态恢复后继续训练。

**适配功能点<a name="section1446615300284"></a>**

在弹性训练中，集群大脑会根据全局故障信息决策恢复策略，并将策略下发到MindIO。调度器需要支持故障Pod调度，而非整个任务重调度，支持恢复策略依次回退。在训练容器中，框架首先初始化MindIO服务。启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保证模型参数的冗余备份。当异常发生时，通过异常捕获装饰器捕获故障模式，并由MindIO上报给集群大脑决策。

-   当集群大脑检测到故障，且无冗余备份资源时，下发缩容策略到MindIO，执行算子资源清理、缩容重建，以缩容状态继续训练。
-   当集群大脑检测到有可用资源且新节点成功拉起时，下发扩容策略到MindIO，执行算子资源清理、扩容通信重建、扩容参数面恢复和扩容状态回滚，完成弹性扩容恢复原有规模继续训练。

对于非MindSpeed-LLM和MindCluster平台用户，需在框架侧完成[表2](#table1995514113610)的功能适配。

**表 2**  弹性训练框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="7.520000000000001%" id="mcps1.2.6.1.1"><p id="p4637165993110"><a name="p4637165993110"></a><a name="p4637165993110"></a>序号</p>
</th>
<th class="cellrowborder" valign="top" width="18.810000000000002%" id="mcps1.2.6.1.2"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="34.39%" id="mcps1.2.6.1.3"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.190000000000005%" id="mcps1.2.6.1.4"><p id="p237216122367"><a name="p237216122367"></a><a name="p237216122367"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="21.090000000000003%" id="mcps1.2.6.1.5"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="7.520000000000001%" headers="mcps1.2.6.1.1 "><p id="p26376591313"><a name="p26376591313"></a><a name="p26376591313"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.6.1.2 "><p id="p1142119117913"><a name="p1142119117913"></a><a name="p1142119117913"></a>初始化启动</p>
</td>
<td class="cellrowborder" valign="top" width="34.39%" headers="mcps1.2.6.1.3 "><p id="p112827185916"><a name="p112827185916"></a><a name="p112827185916"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="16" valign="top" width="18.190000000000005%" headers="mcps1.2.6.1.4 "><p id="p444112643720"><a name="p444112643720"></a><a name="p444112643720"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="21.090000000000003%" headers="mcps1.2.6.1.5 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../references.md#对接非mindspeed-llm框架">表2</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p106371759163113"><a name="p106371759163113"></a><a name="p106371759163113"></a>2</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1942113117919"><a name="p1942113117919"></a><a name="p1942113117919"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p92821518193"><a name="p92821518193"></a><a name="p92821518193"></a>优化器更新前上报优化器更新的开始和结束状态。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p363765912314"><a name="p363765912314"></a><a name="p363765912314"></a>3</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p164211711596"><a name="p164211711596"></a><a name="p164211711596"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p22829180917"><a name="p22829180917"></a><a name="p22829180917"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7637175903115"><a name="p7637175903115"></a><a name="p7637175903115"></a>4</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p134219118919"><a name="p134219118919"></a><a name="p134219118919"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p192829181594"><a name="p192829181594"></a><a name="p192829181594"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1963725993118"><a name="p1963725993118"></a><a name="p1963725993118"></a>5</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1542111118913"><a name="p1542111118913"></a><a name="p1542111118913"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p112826181914"><a name="p112826181914"></a><a name="p112826181914"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p363711591310"><a name="p363711591310"></a><a name="p363711591310"></a>6</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p6421121796"><a name="p6421121796"></a><a name="p6421121796"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11282181811916"><a name="p11282181811916"></a><a name="p11282181811916"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p06375599316"><a name="p06375599316"></a><a name="p06375599316"></a>7</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p34212017916"><a name="p34212017916"></a><a name="p34212017916"></a>弹性训练回调注册</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1528212181493"><a name="p1528212181493"></a><a name="p1528212181493"></a>将弹性训练各个回调函数注册到MindIO。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1163581571719"><a name="p1163581571719"></a><a name="p1163581571719"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_register.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row876023415918"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4637259163118"><a name="p4637259163118"></a><a name="p4637259163118"></a>8</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1342113114912"><a name="p1342113114912"></a><a name="p1342113114912"></a>缩容重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p528210181396"><a name="p528210181396"></a><a name="p528210181396"></a>重建缩容后的通信组、数据迭代器、记录并更新部分框架变量等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p106351815121720"><a name="p106351815121720"></a><a name="p106351815121720"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_scale_in_rebuild.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row17605341596"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1763711599312"><a name="p1763711599312"></a><a name="p1763711599312"></a>9</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p74211911599"><a name="p74211911599"></a><a name="p74211911599"></a>扩容通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p52821818299"><a name="p52821818299"></a><a name="p52821818299"></a>新节点与缩容节点重建通信组。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p126358155177"><a name="p126358155177"></a><a name="p126358155177"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_scale_out_rebuild.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row144412445361"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p56378590319"><a name="p56378590319"></a><a name="p56378590319"></a>10</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p64221014919"><a name="p64221014919"></a><a name="p64221014919"></a>扩容参数面恢复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p728213181097"><a name="p728213181097"></a><a name="p728213181097"></a>通过副本rank与新拉rank参数传输恢复新节点优化器等参数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p935519615208"><a name="p935519615208"></a><a name="p935519615208"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_repair.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row14716101112393"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1963713597315"><a name="p1963713597315"></a><a name="p1963713597315"></a>11</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p842217115916"><a name="p842217115916"></a><a name="p842217115916"></a>扩容状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p72821618294"><a name="p72821618294"></a><a name="p72821618294"></a>恢复缩容时更改框架变量、重建数据集等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p135516622010"><a name="p135516622010"></a><a name="p135516622010"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_rollback.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row164994019817"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p563705923117"><a name="p563705923117"></a><a name="p563705923117"></a>12</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17422911918"><a name="p17422911918"></a><a name="p17422911918"></a>新拉起节点torch通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1228218181798"><a name="p1228218181798"></a><a name="p1228218181798"></a>新拉起节点恢复前跳过通信。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.4 "><p id="p627084962414"><a name="p627084962414"></a><a name="p627084962414"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py#:~text=def pre_register_patches(self, patch_manager, args):" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row5499401185"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5637135973120"><a name="p5637135973120"></a><a name="p5637135973120"></a>13</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11422518917"><a name="p11422518917"></a><a name="p11422518917"></a>缩容训练全局组通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p32829185918"><a name="p32829185918"></a><a name="p32829185918"></a>缩容训练时使用缩容后全局组替换原全局组通信。</p>
</td>
</tr>
<tr id="row1550640684"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3637125953118"><a name="p3637125953118"></a><a name="p3637125953118"></a>14</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14221517919"><a name="p14221517919"></a><a name="p14221517919"></a>缩容训练副本组通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p6282018699"><a name="p6282018699"></a><a name="p6282018699"></a>缩容训练时副本rank替代故障rank与故障rank所在副本组通信。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.4 "><p id="p1320441192517"><a name="p1320441192517"></a><a name="p1320441192517"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row1650540489"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p06374591313"><a name="p06374591313"></a><a name="p06374591313"></a>15</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5422415915"><a name="p5422415915"></a><a name="p5422415915"></a>缩容训练参数适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p112827184913"><a name="p112827184913"></a><a name="p112827184913"></a>缩容训练时修改num_microbatches、world_size、global_batch_size等参数。</p>
</td>
</tr>
<tr id="row7501940584"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6637165923110"><a name="p6637165923110"></a><a name="p6637165923110"></a>16</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5422211095"><a name="p5422211095"></a><a name="p5422211095"></a>梯度精度计算适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p5282161815916"><a name="p5282161815916"></a><a name="p5282161815916"></a>适配因缩容num_micro_batches等变化导致的精度梯度变化。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17653121214259"><a name="p17653121214259"></a><a name="p17653121214259"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接1</a></p>
<p id="p116531412172515"><a name="p116531412172515"></a><a name="p116531412172515"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/pretrain_gpt.py#:~text=if args.enable_elastic_training:" target="_blank" rel="noopener noreferrer">LLM仓参考链接2</a></p>
</td>
</tr>
<tr id="row145017409813"><td class="cellrowborder" valign="top" width="7.520000000000001%" headers="mcps1.2.6.1.1 "><p id="p17637165943120"><a name="p17637165943120"></a><a name="p17637165943120"></a>17</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.6.1.2 "><p id="p15422131799"><a name="p15422131799"></a><a name="p15422131799"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="34.39%" headers="mcps1.2.6.1.3 "><p id="p1628291814912"><a name="p1628291814912"></a><a name="p1628291814912"></a>根据全局故障信息决策恢复策略，并将策略下发到MindIO。支持恢复策略回退，弹性训练失败回退到临终遗言等策略。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="18.190000000000005%" headers="mcps1.2.6.1.4 "><p id="p1504404816"><a name="p1504404816"></a><a name="p1504404816"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="21.090000000000003%" headers="mcps1.2.6.1.5 "><p id="p20447192572312"><a name="p20447192572312"></a><a name="p20447192572312"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row155014017818"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p063755903111"><a name="p063755903111"></a><a name="p063755903111"></a>18</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1042215113910"><a name="p1042215113910"></a><a name="p1042215113910"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18282181818913"><a name="p18282181818913"></a><a name="p18282181818913"></a>调度故障Pod。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p10446112592310"><a name="p10446112592310"></a><a name="p10446112592310"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v7.3.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

[表2](#table1995514113610)中序号为1-6的适配项为MindIO TFT（MindCluster MindIO Training Fault Tolerance）公共逻辑，序号为17-18的适配项为断点续训公共逻辑，本章节不再详细描述。以下针对弹性训练特有功能点，基于Megatron 0.12.1版本进行简要介绍。

-   弹性训练回调注册

    在训练拉起初始化时调用，将弹性训练缩容和扩容恢复过程中需要执行的回调函数注册到MindIO中，进而在恢复过程中被调用。

-   缩容重建
    1.  基于缩容后成员创建新的全局通信组并记录，后续将替代原全局通信组进行通信。
    2.  记录框架原始DP size、num\_microbatches等参数作为后续扩容恢复使用，并更新为缩容后数据。
    3.  基于故障Rank信息重建缩容后其他局部通信组，并更新模型、优化器等实例对象中的通信组。
    4.  重建数据集、重新初始化部分框架实例、参数等。

-   扩容通信重建
    1.  重建扩容后全局和局部通信组，并更新模型、优化器等实例对象中的通信组。
    2.  恢复框架DP size等参数、重新初始化部分框架实例等。

-   扩容参数面恢复
    1.  为新拉起的rank训练进程和备份rank训练进程创建通信组，用于发送和接收优化器参数等。
    2.  备份rank训练进程向新拉起的rank训练进程发送恢复所需的优化器参数。
    3.  新拉起的rank训练进程接收优化器参数后，按需更新optimizer、opt\_param\_scheduler、全局args等参数。

-   扩容状态回滚
    1.  恢复框架num\_microbatches等参数。
    2.  恢复训练前将优化器参数拷贝到模型参数中，并在对应DP域内进行一次all\_gather通信操作，确保模型参数为最新状态。
    3.  修复打印训练迭代日志。
    4.  重建数据集，重新初始化部分框架实例、参数等。
    5.  销毁恢复过程中发送和接收参数的通信组。

-   新拉起节点torch通信适配
    1.  对于重启节点，从pretrain启动流程到进入train之间，会下发通信算子，但正常训练rank在该阶段并未与重启节点配套重建通信域，集合通信无法成功，因此直接跳过。
    2.  对于重启节点，从pretrain启动流程到进入train之间，会创建并行通信域，但正常训练rank在该阶段并未与重启节点配套重建通信域，对于gloo组会报错，因此直接跳过新建gloo通信组。

-   缩容训练全局组通信适配

    在缩容训练过程中，由于故障节点已经被删除，因此使用原全局通信组通信会失败，需替换为缩容后的全局通信组。

-   缩容训练副本组通信适配

    在[LLM仓参考链接](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，start\_param\_sync\_wrapper、get\_grad\_norm\_fp32\_wrapper、get\_parameter\_state\_dp\_zero\_wrapper等是为了适配缩容训练时副本组通信而patch，下面以get\_parameter\_state\_dp\_zero\_wrapper为例介绍副本组适配原理：

    假设当前tp=8、pp=1、dp=4。DP组分别为rank \[0,8,16,24\]、\[1,9,17,25\]、\[2,10,18,26\]、…、\[7,15,23,31\]，按照副本优化器原理，副本组分别为rank \[0,8\]、\[16,24\]、\[1,9\]、\[17,25\]、\[2,10\]、\[18,26\]、…、\[7,15\]、\[23,31\]，rank 0-15与rank 16-31互为副本。rank 31故障后，将rank 24-31对应DP域删除继续缩容训练。

    原生Megatron会使用优化器实例的data\_parallel\_group\_gloo成员变量对应的group（即DP组，在使用MindIO的优化器副本时为副本组）进行通信。缩容后不包含删除的rank 24-31的副本组，继续按照原有通信组进行通信，包含缩容rank的副本组使用组内正常rank与缩容rank对应的副本rank组成的缩容组进行通信，例如副本组rank \[23,31\]缩容后，通信使用的通信组为rank \[23,15\]。

-   缩容训练参数适配

    在[LLM仓参考链接](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，patch\_world\_size\_func\_wrapper、log\_wrapper、is\_last\_rank\_wrapper、optimizer\_param\_scheduler\_step\_wrapper、track\_app\_tag\_wrapper、print\_rank\_last\_wrapper、num\_floating\_point\_operations\_wrapper等是为了适配global\_batch\_size、world\_size等训练中使用的参数而patch。例如：原生使用dp\_size\*micro\_batch\_size\*num\_microbatches，缩容后各个DP内num\_microbatches可能不一样，因此直接使用args.globatch\_size。缩容后判断是否最后一个rank使用缩容后的全局组；全局组大小修改为缩容后的大小等。

-   梯度精度计算适配

    在[LLM仓参考链接1](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，start\_grad\_sync\_wrapper、forward\_step\_wrapper、elastic\_training\_get\_forward\_backward\_func\_wrapper以及[LLM仓参考链接2](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/pretrain_gpt.py#:~text=if%20args.enable_elastic_training%3A)所指向的loss\_func的代码是为了适配因缩容导致的精度梯度变化而patch或修改。

    -   loss\_func由每个micro\_batch都要进行DP组内all\_reduce通信修改为缩容训练时不进行通信，原因是缩容后每个DP域内num\_micro\_batches数量可能不一样，导致前几个DP会多执行一次all\_reduce而卡住。
    -   start\_grad\_sync\_wrapper中将梯度缩放因子gradient\_scaling\_factor修改为1.0 / \(arguments.global\_batch\_size / arguments.micro\_batch\_size\)，即在原1/dp\_size基础上再除以num\_micro\_batches。
    -   forward\_step\_wrapper将入参num\_microbatches修改为1，目的是loss计算时不再除以num\_microbatches，因为在start\_grad\_sync\_wrapper中已经除以了num\_microbatches。
    -   elastic\_training\_get\_forward\_backward\_func\_wrapper因为loss\_func没有执行DP组内all\_reduce，原生forward\_backward\_func执行完成后，在最后一个PP时将losses\_reduced每个key的和（即所有micro\_batch的lm loss相加）在DP组内执行all\_reduce操作求和。


### 训练恢复<a name="ZH-CN_TOPIC_0000002511426359"></a>

#### 训练恢复原理说明<a name="ZH-CN_TOPIC_0000002479226500"></a>

在完成故障处理后，训练进程会被重新拉起，拉起的训练进程需要完成模型权重的保存和加载，才能回到任务中断时的训练状态。在正常训练中，每隔一段时间保存训练模型权重的CKPT（CheckPoint）文件，在任务中断后，新拉起的进程可以加载之前保存的CKPT文件，从而恢复到之前保存点的模型权重状态，减少训练时间。对于不同框架，保存和加载CKPT的方法不一样，以下给出了TensorFlow、PyTorch、MindSpore保存和加载CKPT的示例，用户需按照示例修改自己的**训练模型脚本**。

**PyTorch<a name="section77915151121"></a>**

1.  保存CKPT。

    ```
    def save_checkpoint(state, is_best, args, filename='checkpoint.pth.tar'):
        filename2 = os.path.join(args.save_ckpt_path, filename)
        torch.save(state, filename2)
        if is_best:
            shutil.copyfile(filename2, os.path.join(args.save_ckpt_path, 'model_best.pth.tar'))
    ```

2.  加载CKPT。

    ```
    checkpoint = torch.load(args.checkpoint_path, map_location=loc)
                args.start_epoch = checkpoint['epoch']
                best_acc1 = checkpoint['best_acc1']
                model.load_state_dict(checkpoint['state_dict'])
                optimizer.load_state_dict(checkpoint['optimizer'])
    ```

**MindSpore<a name="section104642081315"></a>**

1.  保存CKPT。

    ```
    ms.save_checkpoint(net, "./lenet.ckpt",
                       choice_func=lambda x: x.startswith("conv") and not x.startswith("conv1"))
    ```

2.  加载CKPT。

    ```
    param_dict = ms.load_checkpoint("./lenet.ckpt")
    ```

**TensorFlow<a name="section20353419915"></a>**

1.  使用tf.compat.v1.train.CheckpointManager接口进行CKPT管理。

    ```
      checkpoint_manager = tf.train.CheckpointManager(
          runnable.checkpoint,
          directory=flags_obj.model_dir,
          max_to_keep=10,
          step_counter=runnable.global_step,
          checkpoint_interval=checkpoint_interval)
    ```

2.  保存CKPT（创建一个新的CKPT）。

    ```
    Save(
        Checkpoint_number=None, check_internal=True, options=None
    )
    ```

3.  加载保存的CKPT（尝试加载从目录中的最新的CKPT）。

    ```
    Restore_or_initialize()
    ```


#### 周期性CKPT保存<a name="ZH-CN_TOPIC_0000002479386434"></a>

现有大规模集群训练主要通过CKPT（CheckPoint）机制，即在训练过程中周期性保存训练过程数据（模型参数等）作为CKPT。当业务平台检测到故障发生后，可退出当前训练任务，通过重新加载CKPT数据，从CKPT保存时刻开始恢复训练，避免从头开始重新进行训练。

周期性CKPT保存分为2个部分：异步CKPT保存以及内存CKPT加载。

-   **异步CKPT保存**

    MindIO ACP提供异步保存周期性CKPT的能力。未使用MindIO ACP时，需要将需要保存的参数从设备拷贝到主机侧，再从主机侧落盘到存储上，这一时间通常在分钟级。MindIO ACP提供异步落盘的能力，当需要保存的参数从设备拷贝到主机侧后，通过异步进程进行落盘到存储，不会阻塞训练进程，落盘的过程中训练可以继续进行。

-   **内存CKPT加载**

    MindIO ACP提供基于内存的周期性CKPT加载的能力。在训练恢复时，通常需要从存储加载之前保存的周期性CKPT，加载完成后恢复训练状态再继续训练。但是，由于数据量较大和存储性能限制，大模型任务通常加载时间在分钟级。为了降低CKPT加载时间，从而降低训练恢复的时间，MindIO ACP提供基于内存的周期性CKPT加载机制，故障后直接基于内存加载，将降低大量加载的时间。

**推荐配置<a name="section883116216236"></a>**

在使用故障重调度的CKPT保存能力时，需根据实际情况选择周期性保存CKPT频率，用户可参考如[图1](#fig41241253101)所示的推荐频率。

**图 1**  周期性CKPT保存频率推荐<a name="fig41241253101"></a>  
![](../../figures/scheduling/周期性CKPT保存频率推荐.png "周期性CKPT保存频率推荐")

使用周期CKPT恢复能力，训练恢复后将丢失上一次周期保存点到故障点这一时间段的训练状态。因此，如果想要降低每次故障导致的训练状态损失，需要降低周期性保存的间隔。但是，每次保存需要中断训练后将CKPT从设备侧落盘到存储侧，这浪费了大量的训练时间。如果降低周期性保存的间隔，将导致训练时间的浪费，从而也会带来训练时间的损失。综上所述，如果单次保存时间恒定，通常需要作出保存损失和故障损失的综合权衡。

为了降低上述损失，需要降低单次保存时间。单次保存时间受到保存数据量及存储性能的影响，通常难以改变这两者。本产品提供MindIO ACP产品解决周期性CKPT恢复损失高的问题。


#### 临终CKPT保存<a name="ZH-CN_TOPIC_0000002511426397"></a>

尽管通过异步保存周期性CKPT能够降低周期性保存间隔，从而降低每次故障的损失，但是由于仍然具有保存开销，难以做到秒级的故障损失。因此，MindCluster集群调度组件提供临终保存CKPT能力，在故障时刻保存当前step初始的参数状态，从而将训练恢复的状态损失降低到一个“step”以内。

MindCluster MindIO Try To Persist（下文简称MindIO TTP）提供临终CKPT能力，帮助用户在故障时刻保存临终时刻CKPT。

了解临终CKPT保存的详细介绍，请参见[故障恢复与加速](../references.md#故障恢复加速)。

了解临终CKPT保存的配置步骤，请参见[配置临终CKPT保存](#配置临终ckpt保存)。

**适配功能点<a name="section1446615300284"></a>**

在临终CKPT中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，之后执行算子资源清理，基于副本完成临终CKPT保存。

对于非MindSpeed-LLM用户，需在框架侧完成[表1](#table1995514113610)的功能适配。

**表 1**  临终CKPT保存框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="20.632063206320634%" id="mcps1.2.4.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="50.51505150515051%" id="mcps1.2.4.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="28.852885288528853%" id="mcps1.2.4.1.3"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="20.632063206320634%" headers="mcps1.2.4.1.1 "><p id="p19821165420516"><a name="p19821165420516"></a><a name="p19821165420516"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="50.51505150515051%" headers="mcps1.2.4.1.2 "><p id="p5821185419518"><a name="p5821185419518"></a><a name="p5821185419518"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="28.852885288528853%" headers="mcps1.2.4.1.3 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../references.md#非mindspeed-llm用户对接指导">对接非MindSpeed-LLM用户</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p6821754125118"><a name="p6821754125118"></a><a name="p6821754125118"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p182111545511"><a name="p182111545511"></a><a name="p182111545511"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p082105435116"><a name="p082105435116"></a><a name="p082105435116"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11821254135115"><a name="p11821254135115"></a><a name="p11821254135115"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1282145475118"><a name="p1282145475118"></a><a name="p1282145475118"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p128211854175117"><a name="p128211854175117"></a><a name="p128211854175117"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p7821115445113"><a name="p7821115445113"></a><a name="p7821115445113"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p138216541513"><a name="p138216541513"></a><a name="p138216541513"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p6821754135112"><a name="p6821754135112"></a><a name="p6821754135112"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p14822105475115"><a name="p14822105475115"></a><a name="p14822105475115"></a>通过回调函数完成算子清理、恢复算子下发能力。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p138221254105112"><a name="p138221254105112"></a><a name="p138221254105112"></a>临终CKPT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1582210547513"><a name="p1582210547513"></a><a name="p1582210547513"></a>通过新增回调函数、优化器副本dump方法完成临终CKPT保存。</p>
</td>
</tr>
</tbody>
</table>


#### 参数面CKPT传输恢复<a name="ZH-CN_TOPIC_0000002511426371"></a>

通过临终CKPT能力可以将每次训练由于CKPT回滚机制导致的训练回滚损失降到一个“step”内，但是在故障时刻时需要进行落盘保存，并在容错完成训练恢复后需要加载存储上的CKPT进行恢复，将导致整体故障恢复时间延长。因此，为了降低故障恢复时间，MindCluster集群调度组件提供参数面CKPT传输恢复能力。

在故障时刻将参数状态保持在设备侧，在容错完成训练恢复时将正常卡内的参数状态通过参数面网络传输到容错处理的卡上，从而快速恢复容错处理卡的参数状态。当前该能力需要结合进程级别重调度和进程级在线恢复使用，不支持用户独立使用。

了解参数面CKPT的配置步骤，请参见[配置参数面传参恢复](#配置参数面传参恢复)。




## 准备K8s和共享存储<a name="ZH-CN_TOPIC_0000002479386542"></a>

断点续训特性是基于MindCluster集群调度组件的高阶特性，结合昇腾软硬件全栈实现训练故障恢复，使用断点续训特性前需要满足以下前置条件。

-   完成K8s集群基础性能调优，详情请参见[K8s集群基础性能调优](../appendix.md#k8s集群基础性能调优)。

-   具备共享存储系统

    断点续训特性的部分流程依赖读取存储数据，如加载CKPT、拉起训练和编译缓存加载等，存储性能会影响断点续训整体恢复时间。为避免训练恢复时间劣化，建议进行存储性能配置优化，以下提供的推荐配置以万卡规模集群为例。

    -   8K IO读IOPS：\>1024W
    -   8K IO写IOPS：\>128W
    -   大文件顺序读带宽：\>288GB/s
    -   大文件创建写带宽：\>173GB/s


## （可选）配置故障检测级别<a name="ZH-CN_TOPIC_0000002479386556"></a>

### 配置说明<a name="ZH-CN_TOPIC_0000002479386448"></a>

断点续训针对节点故障中**节点硬件故障**、**芯片故障、灵衢总线设备故障**和**公共故障**的不同故障码，提供了默认的故障级别和对应级别的故障处理策略；**芯片故障**还提供了默认的故障频率和时长，以及对应的故障处理策略。

若用户需要修改故障处理策略可参见本章节。若无特殊需求，请勿随意修改。

**支持配置的故障级别说明<a name="section257513292065"></a>**

不同类型的故障支持配置的故障级别如下表所示。

**表 1**  支持配置的故障级别

<a name="table4710459145316"></a>
<table><thead align="left"><tr id="row37104590534"><th class="cellrowborder" valign="top" id="mcps1.2.5.1.1"><p id="p7710135925316"><a name="p7710135925316"></a><a name="p7710135925316"></a>故障名称</p>
</th>
<th class="cellrowborder" colspan="3" valign="top" id="mcps1.2.5.1.2"><p id="p11175192213564"><a name="p11175192213564"></a><a name="p11175192213564"></a>支持配置的故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="row271045905320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p271015916536"><a name="p271015916536"></a><a name="p271015916536"></a>节点故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p66711187562"><a name="p66711187562"></a><a name="p66711187562"></a>NotHandleFault、PreSeparateFault、SeparateFault</p>
</td>
</tr>
<tr id="row3710165935311"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17710125955315"><a name="p17710125955315"></a><a name="p17710125955315"></a>芯片故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p21371428713"><a name="p21371428713"></a><a name="p21371428713"></a>NotHandleFault、RestartRequest、RestartBusiness、FreeRestartNPU、RestartNPU、SeparateNPU、PreSeparateNPU、SubHealthFault</p>
</td>
</tr>
<tr id="row5710125913537"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10710959185319"><a name="p10710959185319"></a><a name="p10710959185319"></a>灵衢总线设备故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p6631112135616"><a name="p6631112135616"></a><a name="p6631112135616"></a>NotHandleFault、SubHealthFault、ResetFault、SeparateFault<span id="ph51441721217"><a name="ph51441721217"></a><a name="ph51441721217"></a>、</span><span id="ph375517710129"><a name="ph375517710129"></a><a name="ph375517710129"></a>RestartRequestFault</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913517"><a name="p116115913517"></a><a name="p116115913517"></a>公共故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536717"><a name="p147536536717"></a><a name="p147536536717"></a>NotHandleFault、SeparateNPU、SubHealthFault<span id="ph632635517598"><a name="ph632635517598"></a><a name="ph632635517598"></a>、PreSeparateNPU</span></p>
</td>
</tr>
</tbody>
</table>

在以上表格中，每种故障级别的处理策略说明如下。

**表 2**  故障级别及处理说明

<a name="table103716651410"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="35.74%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="23.39%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a>重调度处理</p>
</th>
<th class="cellrowborder" valign="top" width="21.81%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a>优雅容错处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a>对业务无影响的故障，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a>暂不处理</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a>影响业务执行，需要重新执行业务请求。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a>隔离芯片，进行任务重调度。</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><a name="note11901123612819"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1069261722310"><a name="p1069261722310"></a><a name="p1069261722310"></a>若推理任务订阅<span id="ph4356222144812"><a name="ph4356222144812"></a><a name="ph4356222144812"></a>了</span>故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a>影响业务执行，需要重新执行业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a>重新执行业务</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a>影响业务执行，待芯片空闲时需复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a>等待芯片空闲后复位芯片。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a>影响业务执行，需立即复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a>立即停止训练业务，复位芯片后重新执行业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="row870814247412"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5708202454117"><a name="p5708202454117"></a><a name="p5708202454117"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p12708162474117"><a name="p12708162474117"></a><a name="p12708162474117"></a>任务一定会受到影响。</p>
<div class="note" id="note1521013164613"><a name="note1521013164613"></a><a name="note1521013164613"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p92101114465"><a name="p92101114465"></a><a name="p92101114465"></a>灵衢总线设备故障级别为SeparateFault时，表示业务运行失败，需更换器件或板卡。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p0708624204112"><a name="p0708624204112"></a><a name="p0708624204112"></a>任务重调度</p>
<div class="note" id="note44451347164716"><a name="note44451347164716"></a><a name="note44451347164716"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p64453471479"><a name="p64453471479"></a><a name="p64453471479"></a>灵衢总线设备故障下，本故障级别代表的故障处理策略为停止当前训练任务，隔离节点，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p137081824174117"><a name="p137081824174117"></a><a name="p137081824174117"></a>-</p>
</td>
</tr>
<tr id="row5706333131216"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p177061833201220"><a name="p177061833201220"></a><a name="p177061833201220"></a><span id="ph141513510124"><a name="ph141513510124"></a><a name="ph141513510124"></a>RestartRequestFault</span></p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p070623351220"><a name="p070623351220"></a><a name="p070623351220"></a><span id="ph18501459184"><a name="ph18501459184"></a><a name="ph18501459184"></a>业务运行失败，需要重新执行业务请求。</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p770653313124"><a name="p770653313124"></a><a name="p770653313124"></a><span id="ph38912127169"><a name="ph38912127169"></a><a name="ph38912127169"></a>停止当前训练任务，隔离节点，进行任务重调度。</span></p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p6706113331213"><a name="p6706113331213"></a><a name="p6706113331213"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="row3938182254418"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p39381822174417"><a name="p39381822174417"></a><a name="p39381822174417"></a>ResetFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p1193862274418"><a name="p1193862274418"></a><a name="p1193862274418"></a>业务运行失败</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p184323519501"><a name="p184323519501"></a><a name="p184323519501"></a>停止当前训练任务，隔离节点，进行任务重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p18938822204411"><a name="p18938822204411"></a><a name="p18938822204411"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row102215292529"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a>暂不影响业务，后续不再调度任务到该芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a>预隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a>预隔离芯片</p>
</td>
</tr>
<tr id="row84541721401"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p174559214016"><a name="p174559214016"></a><a name="p174559214016"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p145562114011"><a name="p145562114011"></a><a name="p145562114011"></a>可能导致任务受到影响。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p54556214409"><a name="p54556214409"></a><a name="p54556214409"></a>该节点上有任务则不处理，后续调度时不调度任务到该节点。</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p1245572144015"><a name="p1245572144015"></a><a name="p1245572144015"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row0352224175218"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="#yaml参数说明">YAML参数说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="#任务yaml配置示例">配置YAML</a>策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault</p>
<p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p109369476532"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p109369476532"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p109369476532"></a>处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a>根据策略进行处理。</p>
</td>
</tr>
</tbody>
</table>


### 节点硬件故障<a name="ZH-CN_TOPIC_0000002479226584"></a>

#### 配置文件说明<a name="ZH-CN_TOPIC_0000002479226562"></a>

断点续训针对节点故障中**节点硬件故障**的不同级别进行分级处理。NodeD组件会获取到当前故障的故障码，根据NodeDConfiguration.json中故障码配置的故障级别，对故障进行相应处理。节点硬件故障支持的故障级别和处理方式说明如下。

NodeD组件的配置文件NodeDConfiguration.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码的故障级别，可以通过由NodeDConfiguration.json创建的mindx-dl-node-fault-config文件实现，操作指导请参见[（可选）配置节点硬件故障级别](#可选配置节点硬件故障级别)。

**表 1**  故障说明

<a name="table1124934413485"></a>
<table><thead align="left"><tr id="row92491944194810"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p13249944104811"><a name="p13249944104811"></a><a name="p13249944104811"></a>故障级别</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p1425034414484"><a name="p1425034414484"></a><a name="p1425034414484"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p325034411485"><a name="p325034411485"></a><a name="p325034411485"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row62506448481"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p52501244194817"><a name="p52501244194817"></a><a name="p52501244194817"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p12501441483"><a name="p12501441483"></a><a name="p12501441483"></a>无需处理</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p19250744134815"><a name="p19250744134815"></a><a name="p19250744134815"></a>对任务无影响</p>
</td>
</tr>
<tr id="row142501844164817"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1825015449485"><a name="p1825015449485"></a><a name="p1825015449485"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p52501443481"><a name="p52501443481"></a><a name="p52501443481"></a>该节点上有任务则不处理，后续调度时不调度任务到该节点</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p17250204410487"><a name="p17250204410487"></a><a name="p17250204410487"></a>可能导致任务受到影响</p>
</td>
</tr>
<tr id="row202502044184816"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p125074424815"><a name="p125074424815"></a><a name="p125074424815"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p12502044184819"><a name="p12502044184819"></a><a name="p12502044184819"></a>任务重调度</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p14250134434812"><a name="p14250134434812"></a><a name="p14250134434812"></a>任务一定会受到影响</p>
</td>
</tr>
<tr id="row182503448486"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p3250644124818"><a name="p3250644124818"></a><a name="p3250644124818"></a>注：</p>
<p id="p1125014484811"><a name="p1125014484811"></a><a name="p1125014484811"></a>故障级别的高低为NotHandleFault &lt; PreSeparateFault &lt; SeparateFault。</p>
</td>
</tr>
</tbody>
</table>

**表 2**  节点状态说明

<a name="table20250114420483"></a>
<table><thead align="left"><tr id="row6250144154810"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p16250124444814"><a name="p16250124444814"></a><a name="p16250124444814"></a>节点状态</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p13250144413488"><a name="p13250144413488"></a><a name="p13250144413488"></a>最高故障级别</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p2250104484818"><a name="p2250104484818"></a><a name="p2250104484818"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p1625011445483"><a name="p1625011445483"></a><a name="p1625011445483"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row625094424817"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p142501044164817"><a name="p142501044164817"></a><a name="p142501044164817"></a>Healthy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p142501844194814"><a name="p142501844194814"></a><a name="p142501844194814"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p9250164474815"><a name="p9250164474815"></a><a name="p9250164474815"></a>无需处理</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1625019440486"><a name="p1625019440486"></a><a name="p1625019440486"></a>该节点为健康节点，可以正常训练。</p>
</td>
</tr>
<tr id="row18250134484811"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p182507446489"><a name="p182507446489"></a><a name="p182507446489"></a>PreSeparate</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p102501344124811"><a name="p102501344124811"></a><a name="p102501344124811"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1825014412488"><a name="p1825014412488"></a><a name="p1825014412488"></a>该节点上有任务则不处理，后续调度时不调度任务到该节点</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1325004411485"><a name="p1325004411485"></a><a name="p1325004411485"></a>该节点为亚健康节点，暂时可能对任务无影响，待任务受到影响退出后，后续不会再调度任务到该节点。</p>
</td>
</tr>
<tr id="row3250124404811"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p9250134434813"><a name="p9250134434813"></a><a name="p9250134434813"></a>UnHealthy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p102505443488"><a name="p102505443488"></a><a name="p102505443488"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p17250644104819"><a name="p17250644104819"></a><a name="p17250644104819"></a>任务重调度</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1625016444482"><a name="p1625016444482"></a><a name="p1625016444482"></a>该节点为故障节点，将影响训练任务，立即将任务调离该节点。</p>
</td>
</tr>
<tr id="row22501244104814"><td class="cellrowborder" colspan="4" valign="top" headers="mcps1.2.5.1.1 mcps1.2.5.1.2 mcps1.2.5.1.3 mcps1.2.5.1.4 "><p id="p2250134419486"><a name="p2250134419486"></a><a name="p2250134419486"></a>注：</p>
<a name="ul52503441488"></a><a name="ul52503441488"></a><ul id="ul52503441488"><li>当前节点的健康状态，主要通过本节点硬件故障的最高故障级别判断。</li></ul>
<a name="ul325084417483"></a><a name="ul325084417483"></a><ul id="ul325084417483"><li>Healthy、PreSeparate和UnHealthy是<span id="ph425019442481"><a name="ph425019442481"></a><a name="ph425019442481"></a>MindCluster</span>自定义的节点状态，主要是用于后续任务的调度和处理。</li><li>查看节点状态和节点硬件故障信息，可参见<a href="../common_operations.md#查询上报的故障信息">查询上报的故障信息</a>章节进行操作。</li></ul>
</td>
</tr>
</tbody>
</table>


#### （可选）配置节点硬件故障级别<a name="ZH-CN_TOPIC_0000002511346507"></a>

在制作NodeD镜像时，会将故障级别配置文件NodeDConfiguration.json内置在镜像中，启动NodeD时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别，可以在集群中创建ConfigMap文件（mindx-dl-node-fault-config）。

-   如果NodeD启动时，集群中已经存在该mindx-dl-node-fault-config，NodeD会优先按照已存在的mindx-dl-node-fault-config中配置的内容，作为当前故障处理依据。
-   如果重新安装NodeD后，集群中已经存在mindx-dl-node-fault-config，NodeD的默认NodeDConfiguration.json将不会生效，使用集群中已经存在mindx-dl-node-fault-config。若想要使用NodeDConfiguration.json的默认配置，可以删除mindx-dl-node-fault-config，使NodeD读取默认的NodeDConfiguration.json文件。
-   如果mindx-dl-node-fault-config内容存在格式错误等问题，NodeD会默认读取镜像中内置的NodeDConfiguration.json文件的内容，作为当前故障处理依据。

**操作步骤<a name="section25164134219"></a>**

以故障码0100001D为例，将当前故障的处理策略NotHandleFault（无需处理）修改为PreSeparateFault（该节点上有任务则不处理，后续不调度任务到该节点）的操作示例如下。

1.  登录环境，进入NodeD解压目录。
2.  执行以下命令，创建动态配置故障级别所需ConfigMap文件（mindx-dl-node-fault-config）。

    ```
    kubectl create cm mindx-dl-node-fault-config -n mindx-dl  --from-file=./NodeDConfiguration.json
    ```

    回显示例如下：

    ```
    configmap/mindx-dl-node-fault-config created
    ```

    **表 1**  参数说明

    <a name="table1925220306444"></a>
    <table><thead align="left"><tr id="row172531430134411"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p16253163094420"><a name="p16253163094420"></a><a name="p16253163094420"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p152534301443"><a name="p152534301443"></a><a name="p152534301443"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1325318306446"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p15214952162210"><a name="p15214952162210"></a><a name="p15214952162210"></a>mindx-dl-node-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p621417523229"><a name="p621417523229"></a><a name="p621417523229"></a>创建的<span id="ph188631730142314"><a name="ph188631730142314"></a><a name="ph188631730142314"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="row925343011442"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p82141952122212"><a name="p82141952122212"></a><a name="p82141952122212"></a>mindx-dl</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p0214952142217"><a name="p0214952142217"></a><a name="p0214952142217"></a>命名空间名称，不能修改该命名空间。</p>
    </td>
    </tr>
    <tr id="row1253183012444"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p182141521222"><a name="p182141521222"></a><a name="p182141521222"></a>NodeDConfiguration.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p22148525226"><a name="p22148525226"></a><a name="p22148525226"></a>用于配置故障码以及对应的故障级别，必须与NodeDConfiguration.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

3.  执行以下命令，编辑mindx-dl-node-fault-config文件。

    ```
    kubectl edit cm -n mindx-dl mindx-dl-node-fault-config
    ```

4.  在mindx-dl-node-fault-config文件中，找到故障码0100001D。

    ```
     "FaultTypeCode": {
            "NotHandleFaultCodes":[
              "0100001D","03000009","03000013","0300000D","03000011"
            ],
    ...
      ],
    ...
    ```

    >[!NOTE] 说明 
    >自定义故障级别时，若不小心导致出现以下问题，则本次修改无效，NodeD将会使用上一次保存的配置进行处理。
    >-   文件格式异常或故障码取值错误，故障码只能为8位的包含数字和字母的字符串。
    >-   同一故障码同时配置在多个故障级别中。

5.  将故障码0100001D在**NotHandleFaultCodes**中删除，并添加到**PreSeparateFaultCodes**中。

    ```
     "FaultTypeCode": {
            "NotHandleFaultCodes":[
             "03000009","03000013","0300000D","03000011"
            ],
            "PreSeparateFaultCodes":[
              "28000037","00000011", "0100001D"
    ...
            ],
    ...
    ```

6.  修改完成后，按“Esc”键，输入:wq!保存并退出。
7.  等mindx-dl-node-fault-config文件更新后，查看操作是否成功。
    1.  执行以下命令，查询NodeD组件日志名称。

        ```
        kubectl get pods -A | grep noded
        ```

        回显示例如下：

        ```
        mindx-dl      noded-c5f52   1/1     Running   0               2m16s
        ```

    2.  通过查询到的组件日志名称，查询NodeD的组件日志信息。

        ```
        kubectl logs noded-c5f52 -n mindx-dl -f
        ```

        若日志出现“update fault config success”，表示动态配置故障码操作成功。



### 芯片故障<a name="ZH-CN_TOPIC_0000002479226466"></a>

#### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346521"></a>

断点续训针对**芯片故障**，支持按故障级别、故障频率和故障时长的配置进行处理。

-   针对芯片故障的**不同级别**进行分级处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**faultCode.json**中故障码配置的故障级别，对故障进行相应处理。
-   针对芯片故障的**故障频率及时长**进行处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**faultCustomization.json**中故障配置的故障频率和时长，对故障进行相应处理。

faultCode.json、faultCustomization.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码对应的故障级别，可以通过由faultCode.json和faultCustomization.json创建的**mindx-dl-fault-config**文件实现。

>[!NOTE] 说明 
>-   每个故障对应的故障码请参见[芯片故障码参考文档](../appendix.md#芯片故障码参考文档)章节。
>-   芯片故障支持配置的故障级别参见[故障级别](#zh-cn_topic_0000002171521445_section5245155017242)。
>-   芯片故障支持配置的故障频率和时长参见[故障频率及时长](#zh-cn_topic_0000002171521445_section115842029104220)。

**faultCode.json中的故障级别<a name="zh-cn_topic_0000002171521445_section5245155017242"></a>**

断点续训针对芯片故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置芯片故障级别](#可选配置芯片故障级别)。

Ascend Device Plugin从驱动获取到芯片故障码后，将根据故障码对设备及业务的影响将故障划分为以下八种级别，详细说明请参见[表1](#zh-cn_topic_0000002171521445_table7618951152212)。

**表 1**  故障级别及处理说明

<a name="zh-cn_topic_0000002171521445_table7618951152212"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002171521445_p12618851162220"></a><a name="zh-cn_topic_0000002171521445_p12618851162220"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="35.78%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002171521445_p16618125162219"></a><a name="zh-cn_topic_0000002171521445_p16618125162219"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="20.01%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002171521445_p1163819316544"><a name="zh-cn_topic_0000002171521445_p1163819316544"></a><a name="zh-cn_topic_0000002171521445_p1163819316544"></a>重调度处理</p>
</th>
<th class="cellrowborder" valign="top" width="25.15%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002171521445_p171971327125410"></a><a name="zh-cn_topic_0000002171521445_p171971327125410"></a>优雅容错处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002171521445_p7618125114229"></a><a name="zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002171521445_p1261835110227"></a><a name="zh-cn_topic_0000002171521445_p1261835110227"></a>对业务无影响的故障，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="20.01%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p10638123115414"><a name="zh-cn_topic_0000002171521445_p10638123115414"></a><a name="zh-cn_topic_0000002171521445_p10638123115414"></a>暂不处理</p>
</td>
<td class="cellrowborder" valign="top" width="25.15%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002171521445_p719714273546"></a><a name="zh-cn_topic_0000002171521445_p719714273546"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002171521445_p5618751102216"></a><a name="zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002171521445_p05771854113911"></a><a name="zh-cn_topic_0000002171521445_p05771854113911"></a>影响业务执行，需要重新执行业务请求。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="20.01%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p13855131912555"><a name="zh-cn_topic_0000002171521445_p13855131912555"></a><a name="zh-cn_topic_0000002171521445_p13855131912555"></a>隔离芯片，进行任务重调度。</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><a name="note11901123612819"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002479386448_p1069261722310"><a name="zh-cn_topic_0000002479386448_p1069261722310"></a><a name="zh-cn_topic_0000002479386448_p1069261722310"></a>若推理任务订阅<span id="zh-cn_topic_0000002479386448_ph4356222144812"><a name="zh-cn_topic_0000002479386448_ph4356222144812"></a><a name="zh-cn_topic_0000002479386448_ph4356222144812"></a>了</span>故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="25.15%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p9145165785517"><a name="zh-cn_topic_0000002171521445_p9145165785517"></a><a name="zh-cn_topic_0000002171521445_p9145165785517"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002171521445_p15618851132212"></a><a name="zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p3618851182216"><a name="zh-cn_topic_0000002171521445_p3618851182216"></a><a name="zh-cn_topic_0000002171521445_p3618851182216"></a>影响业务执行，需要重新执行业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p1419712272549"><a name="zh-cn_topic_0000002171521445_p1419712272549"></a><a name="zh-cn_topic_0000002171521445_p1419712272549"></a>重新执行业务</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002171521445_p66188511222"></a><a name="zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p661865162211"><a name="zh-cn_topic_0000002171521445_p661865162211"></a><a name="zh-cn_topic_0000002171521445_p661865162211"></a>影响业务执行，待芯片空闲时需复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p178789204535"><a name="zh-cn_topic_0000002171521445_p178789204535"></a><a name="zh-cn_topic_0000002171521445_p178789204535"></a>等待芯片空闲后复位芯片。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002171521445_p17618155116227"></a><a name="zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p108302057102114"><a name="zh-cn_topic_0000002171521445_p108302057102114"></a><a name="zh-cn_topic_0000002171521445_p108302057102114"></a>影响业务执行，需立即复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p969972925312"><a name="zh-cn_topic_0000002171521445_p969972925312"></a><a name="zh-cn_topic_0000002171521445_p969972925312"></a>立即停止训练业务，复位芯片后重新执行业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002171521445_p019742745411"></a><a name="zh-cn_topic_0000002171521445_p019742745411"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row102215292529"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002171521445_p16227299522"></a><a name="zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002171521445_p546081915499"></a><a name="zh-cn_topic_0000002171521445_p546081915499"></a>暂不影响业务，后续不再调度任务到该芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="20.01%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002171521445_p222102912521"></a><a name="zh-cn_topic_0000002171521445_p222102912521"></a>预隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="25.15%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002171521445_p12221329155217"></a><a name="zh-cn_topic_0000002171521445_p12221329155217"></a>预隔离芯片</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row0352224175218"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002171521445_p835213245522"></a><a name="zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="#yaml参数说明">YAML参数说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="20.01%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="#任务yaml配置示例">配置YAML</a>策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002171521445_note7936204710536"></a><a name="zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault</p>
<p id="zh-cn_topic_0000002171521445_p109369476532"><a name="zh-cn_topic_0000002171521445_p109369476532"></a><a name="zh-cn_topic_0000002171521445_p109369476532"></a>处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="25.15%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002171521445_p8352172425218"></a><a name="zh-cn_topic_0000002171521445_p8352172425218"></a>根据策略进行处理。</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>-   复位芯片前需要停止训练进程，否则复位将失败。
>-   若Ascend Device Plugin通过订阅的方式收到了无法识别的故障码（未保存在faultCode.json中），默认按照订阅接口给的处理意见进行故障处理。若订阅接口收到的故障等级为“提示”或“次要”，则按照NotHandleFault级别处理；若故障等级为其他等级，则按照SeparateNPU级别处理。

**故障频率及时长<a name="zh-cn_topic_0000002171521445_section115842029104220"></a>**

断点续训针对芯片故障的故障频率及时长进行处理。某些硬件类故障可能在一次训练任务中反复出现，导致训练任务中断反复进行重调度。集群调度组件针对这些故障对应的故障码，提供了提升故障级别的初始化配置文件faultCustomization.json。

-   faultCustomization.json文件提供的初始化配置和故障类型关系如下[初始化配置和故障类型](#zh-cn_topic_0000002171521445_section13684172919539)。
-   faultCustomization.json文件的默认配置（默认值）如下[表2](#zh-cn_topic_0000002171521445_table1519814413572)。
-   若用户需要修改故障频率及时长配置，操作指导请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)。

**初始化配置和故障类型<a name="zh-cn_topic_0000002171521445_section13684172919539"></a>**

当前faultCustomization.json文件中仅提供对可识别的硬件类故障进行提升故障级别的初始化配置。

24小时内发生3次以下故障，则将芯片故障级别提升至需要人工干预的故障级别ManuallySeparateNPU，详细说明请参见[faultCustomization.json参数说明](#zh-cn_topic_0000002171521445_section33036167576)。

下面将以故障名称HBMC Ca Parity错误，对应故障码80E18005为例，将当前的故障级别提升至ManuallySeparateNPU（需要人工干预的故障级别），示例如下。

```
  "FaultFrequency": [
    {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    },
    {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
    }
  ],
```

>[!NOTE] 说明 
>-   故障的处理策略为ManuallySeparateNPU时，即使故障恢复也仍然隔离芯片，需要手动恢复强制隔离的芯片，可以参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)中"手动恢复强制隔离的芯片"步骤进行处理。
>-   除可以识别的硬件故障外，faultCustomization.json文件中还包含以下几类故障。
>    -   无需处理的故障：该类故障出现不影响训练任务及设备，不提供提升故障级别的初始化配置。
>    -   无法识别出是硬件还是软件类故障：该类故障无法准确识别是硬件还是软件故障，且会影响训练任务。该类故障不提供提升故障级别的初始化配置，建议用户根据实际情况手动配置任务支持的断点续训最大次数和达到最大次数后故障的处理策略，可以参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)进行配置。
>    -   软件配置类故障：该类故障为软件配置类问题，正常情况下不会出现。该类故障不提供提升故障级别的初始化配置，建议用户检查软件版本是否配套。

**faultCustomization.json参数说明<a name="zh-cn_topic_0000002171521445_section33036167576"></a>**

用户不手动修改faultCustomization.json文件时，Ascend Device Plugin按照faultCustomization.json的默认配置（默认值）进行故障处理。

**表 2**  faultCustomization.json文件参数说明

<a name="zh-cn_topic_0000002171521445_table1519814413572"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002171521445_row51981644195714"><th class="cellrowborder" valign="top" width="17.27172717271727%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002171521445_p319813443571"><a name="zh-cn_topic_0000002171521445_p319813443571"></a><a name="zh-cn_topic_0000002171521445_p319813443571"></a>一级参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="22.082208220822082%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002171521445_p6198194414574"><a name="zh-cn_topic_0000002171521445_p6198194414574"></a><a name="zh-cn_topic_0000002171521445_p6198194414574"></a>二级参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="60.64606460646065%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002171521445_p19198204485718"><a name="zh-cn_topic_0000002171521445_p19198204485718"></a><a name="zh-cn_topic_0000002171521445_p19198204485718"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002171521445_row31983444574"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p2019934445711"><a name="zh-cn_topic_0000002171521445_p2019934445711"></a><a name="zh-cn_topic_0000002171521445_p2019934445711"></a>GraceTolerance</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p17199114414575"><a name="zh-cn_topic_0000002171521445_p17199114414575"></a><a name="zh-cn_topic_0000002171521445_p17199114414575"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p11991944185713"><a name="zh-cn_topic_0000002171521445_p11991944185713"></a><a name="zh-cn_topic_0000002171521445_p11991944185713"></a>优雅容错相关配置。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note1946012577292"><a name="zh-cn_topic_0000002171521445_note1946012577292"></a><a name="zh-cn_topic_0000002171521445_note1946012577292"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002171521445_p1746035792910"><a name="zh-cn_topic_0000002171521445_p1746035792910"></a><a name="zh-cn_topic_0000002171521445_p1746035792910"></a>GraceTolerance及其子参数不存在或者超出取值范围，则使用默认值。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row141991044175720"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p919911442577"><a name="zh-cn_topic_0000002171521445_p919911442577"></a><a name="zh-cn_topic_0000002171521445_p919911442577"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p0199194414578"><a name="zh-cn_topic_0000002171521445_p0199194414578"></a><a name="zh-cn_topic_0000002171521445_p0199194414578"></a>WaitProcessReadCMTime</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p111991844115710"><a name="zh-cn_topic_0000002171521445_p111991844115710"></a><a name="zh-cn_topic_0000002171521445_p111991844115710"></a>使用优雅容错模式时，等待管理进程读取<span id="zh-cn_topic_0000002171521445_ph1919924435715"><a name="zh-cn_topic_0000002171521445_ph1919924435715"></a><a name="zh-cn_topic_0000002171521445_ph1919924435715"></a>ConfigMap</span>文件的时间，单位为秒，取值范围为5~90，默认值为30。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row15199644205714"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p191995444575"><a name="zh-cn_topic_0000002171521445_p191995444575"></a><a name="zh-cn_topic_0000002171521445_p191995444575"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p419914442579"><a name="zh-cn_topic_0000002171521445_p419914442579"></a><a name="zh-cn_topic_0000002171521445_p419914442579"></a>WaitDeviceResetTime</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p9199144415575"><a name="zh-cn_topic_0000002171521445_p9199144415575"></a><a name="zh-cn_topic_0000002171521445_p9199144415575"></a>使用优雅容错模式时，等待芯片重启的最大时长，单位为秒，取值范围为60~180，默认值为150。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row7199444155712"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p12199164435718"><a name="zh-cn_topic_0000002171521445_p12199164435718"></a><a name="zh-cn_topic_0000002171521445_p12199164435718"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p419974419571"><a name="zh-cn_topic_0000002171521445_p419974419571"></a><a name="zh-cn_topic_0000002171521445_p419974419571"></a>WaitFaultSelfHealingTime</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p4199124416579"><a name="zh-cn_topic_0000002171521445_p4199124416579"></a><a name="zh-cn_topic_0000002171521445_p4199124416579"></a>使用优雅容错模式时，等待RestartBusiness级别故障恢复时间，单位为秒，取值范围为1~30，默认值为15。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row7199184485717"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p21991544155716"><a name="zh-cn_topic_0000002171521445_p21991544155716"></a><a name="zh-cn_topic_0000002171521445_p21991544155716"></a>FaultFrequency</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p220024425711"><a name="zh-cn_topic_0000002171521445_p220024425711"></a><a name="zh-cn_topic_0000002171521445_p220024425711"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p15200044195710"><a name="zh-cn_topic_0000002171521445_p15200044195710"></a><a name="zh-cn_topic_0000002171521445_p15200044195710"></a>自定义故障频率，即某一故障在时间窗口内出现次数达到次数上限时，根据配置的故障处理策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note7518141620301"><a name="zh-cn_topic_0000002171521445_note7518141620301"></a><div class="notebody"><a name="zh-cn_topic_0000002171521445_ul7689137141019"></a><a name="zh-cn_topic_0000002171521445_ul7689137141019"></a><ul id="zh-cn_topic_0000002171521445_ul7689137141019"><li>FaultFrequency及其子参数取值范围不正确，则忽略该条配置。</li><li>FaultFrequency及其子参数数据格式不正确，则会使用默认配置。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row12200204495711"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p1520012443576"><a name="zh-cn_topic_0000002171521445_p1520012443576"></a><a name="zh-cn_topic_0000002171521445_p1520012443576"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p1820084455716"><a name="zh-cn_topic_0000002171521445_p1820084455716"></a><a name="zh-cn_topic_0000002171521445_p1820084455716"></a>EventId</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p02008448576"><a name="zh-cn_topic_0000002171521445_p02008448576"></a><a name="zh-cn_topic_0000002171521445_p02008448576"></a>故障码ID。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note4302258102812"><a name="zh-cn_topic_0000002171521445_note4302258102812"></a><a name="zh-cn_topic_0000002171521445_note4302258102812"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002171521445_p16462113290"><a name="zh-cn_topic_0000002171521445_p16462113290"></a><a name="zh-cn_topic_0000002171521445_p16462113290"></a>每个故障码（EventId）只允许配置一个FaultFrequency参数，如果配置了多个，则只有第一条正确的会生效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row11200114414572"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p13200644195716"><a name="zh-cn_topic_0000002171521445_p13200644195716"></a><a name="zh-cn_topic_0000002171521445_p13200644195716"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p132001044175717"><a name="zh-cn_topic_0000002171521445_p132001044175717"></a><a name="zh-cn_topic_0000002171521445_p132001044175717"></a>TimeWindow</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p1620084410575"><a name="zh-cn_topic_0000002171521445_p1620084410575"></a><a name="zh-cn_topic_0000002171521445_p1620084410575"></a>时间窗口，即统计当前时间减去TimeWindow的时间至当前时间，这段时间范围内的故障次数，单位为秒，取值范围为60~864000。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row1620016445577"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p9200744115710"><a name="zh-cn_topic_0000002171521445_p9200744115710"></a><a name="zh-cn_topic_0000002171521445_p9200744115710"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p8200194411570"><a name="zh-cn_topic_0000002171521445_p8200194411570"></a><a name="zh-cn_topic_0000002171521445_p8200194411570"></a>Times</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p320074415716"><a name="zh-cn_topic_0000002171521445_p320074415716"></a><a name="zh-cn_topic_0000002171521445_p320074415716"></a>任务支持的断点续训最大次数，即同一个故障出现的次数上限，取值范围为1~100。如果在时间窗口内该故障出现次数大于或等于该值，则按照FaultHandling中定义的策略处理和上报。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row7200154435714"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p7200944135719"><a name="zh-cn_topic_0000002171521445_p7200944135719"></a><a name="zh-cn_topic_0000002171521445_p7200944135719"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p122001344155715"><a name="zh-cn_topic_0000002171521445_p122001344155715"></a><a name="zh-cn_topic_0000002171521445_p122001344155715"></a>FaultHandling</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p620084413572"><a name="zh-cn_topic_0000002171521445_p620084413572"></a><a name="zh-cn_topic_0000002171521445_p620084413572"></a>达到断点续训最大次数后故障的处理策略，支持配置不同级别的故障处理策略。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note1120011443576"><a name="zh-cn_topic_0000002171521445_note1120011443576"></a><div class="notebody"><a name="zh-cn_topic_0000002171521445_ul5201124425715"></a><a name="zh-cn_topic_0000002171521445_ul5201124425715"></a><ul id="zh-cn_topic_0000002171521445_ul5201124425715"><li>PreSeparateNPU：大模型的故障处理策略。该故障处理模式为预隔离芯片，根据训练任务实际运行情况判断是否重调度。</li><li>ManuallySeparateNPU：需人工干预的故障处理策略。<a name="zh-cn_topic_0000002171521445_ul1020184411575"></a><a name="zh-cn_topic_0000002171521445_ul1020184411575"></a><ul id="zh-cn_topic_0000002171521445_ul1020184411575"><li>出现该策略时，将直接上报<span id="zh-cn_topic_0000002171521445_ph1920110440571"><a name="zh-cn_topic_0000002171521445_ph1920110440571"></a><a name="zh-cn_topic_0000002171521445_ph1920110440571"></a>K8s</span>该芯片不健康并将芯片名字写入<span id="zh-cn_topic_0000002171521445_ph10507145912293"><a name="zh-cn_topic_0000002171521445_ph10507145912293"></a><a name="zh-cn_topic_0000002171521445_ph10507145912293"></a>device-info-cm</span>。</li><li>芯片名称只要保存于该字段中，即使故障恢复也仍然隔离芯片，直到运维人员手动在该字段中删除芯片名称。可以参见<a href="#可选配置芯片故障频率及时长">（可选）配置芯片故障频率及时长</a>中"手动恢复强制隔离的芯片"步骤进行处理。</li><li>该字段只允许<span id="zh-cn_topic_0000002171521445_ph32011444155712"><a name="zh-cn_topic_0000002171521445_ph32011444155712"></a><a name="zh-cn_topic_0000002171521445_ph32011444155712"></a>Ascend Device Plugin</span>新增或修改，维护人员只能删除该字段中的芯片名称。</li><li>faultCode.json暂不支持该策略。</li></ul>
</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row320118444575"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p12201104413579"><a name="zh-cn_topic_0000002171521445_p12201104413579"></a><a name="zh-cn_topic_0000002171521445_p12201104413579"></a>FaultDuration</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p202015445572"><a name="zh-cn_topic_0000002171521445_p202015445572"></a><a name="zh-cn_topic_0000002171521445_p202015445572"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p920174425716"><a name="zh-cn_topic_0000002171521445_p920174425716"></a><a name="zh-cn_topic_0000002171521445_p920174425716"></a>自定义故障超时策略，当某一故障持续时间达到配置上限时，该故障会按照指定的故障处理策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note471793673013"><a name="zh-cn_topic_0000002171521445_note471793673013"></a><div class="notebody"><a name="zh-cn_topic_0000002171521445_ul13183103116309"></a><a name="zh-cn_topic_0000002171521445_ul13183103116309"></a><ul id="zh-cn_topic_0000002171521445_ul13183103116309"><li>FaultDuration及其子参数取值范围不正确，则忽略该条配置。</li><li>FaultDuration及其子参数数据格式不正确，则会使用默认配置。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row172021244205714"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p12202164415575"><a name="zh-cn_topic_0000002171521445_p12202164415575"></a><a name="zh-cn_topic_0000002171521445_p12202164415575"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p19202544195715"><a name="zh-cn_topic_0000002171521445_p19202544195715"></a><a name="zh-cn_topic_0000002171521445_p19202544195715"></a>EventId</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p122021244155715"><a name="zh-cn_topic_0000002171521445_p122021244155715"></a><a name="zh-cn_topic_0000002171521445_p122021244155715"></a>故障ID。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note199919401295"><a name="zh-cn_topic_0000002171521445_note199919401295"></a><a name="zh-cn_topic_0000002171521445_note199919401295"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002171521445_p0875204382911"><a name="zh-cn_topic_0000002171521445_p0875204382911"></a><a name="zh-cn_topic_0000002171521445_p0875204382911"></a>每个故障码（EventId）只允许配置一个FaultDuration参数，如果配置了多个，则只有第一条正确的会生效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row15202154415571"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p4202144495711"><a name="zh-cn_topic_0000002171521445_p4202144495711"></a><a name="zh-cn_topic_0000002171521445_p4202144495711"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p120234475717"><a name="zh-cn_topic_0000002171521445_p120234475717"></a><a name="zh-cn_topic_0000002171521445_p120234475717"></a>FaultTimeout</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><div class="p" id="zh-cn_topic_0000002171521445_p2202444155717"><a name="zh-cn_topic_0000002171521445_p2202444155717"></a><a name="zh-cn_topic_0000002171521445_p2202444155717"></a>故障持续时间超过该值，则按照FaultHandling中定义的故障处理策略进行处理，单位为秒，取值范围为0~600，默认值说明如下。<a name="zh-cn_topic_0000002171521445_ul156251327007"></a><a name="zh-cn_topic_0000002171521445_ul156251327007"></a><ul id="zh-cn_topic_0000002171521445_ul156251327007"><li>故障ID为81078603的参数面网络故障默认值为20。<p id="zh-cn_topic_0000002171521445_p653165701211"><a name="zh-cn_topic_0000002171521445_p653165701211"></a><a name="zh-cn_topic_0000002171521445_p653165701211"></a></p>
</li><li>其余故障默认值为0。</li></ul>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row4202134413572"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p142023446570"><a name="zh-cn_topic_0000002171521445_p142023446570"></a><a name="zh-cn_topic_0000002171521445_p142023446570"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p320244415574"><a name="zh-cn_topic_0000002171521445_p320244415574"></a><a name="zh-cn_topic_0000002171521445_p320244415574"></a>RecoverTimeout</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><div class="p" id="zh-cn_topic_0000002171521445_p11202124411571"><a name="zh-cn_topic_0000002171521445_p11202124411571"></a><a name="zh-cn_topic_0000002171521445_p11202124411571"></a>故障恢复时间超过该值，则上报故障恢复，单位为秒，取值范围为0~86400，默认值说明如下。<a name="zh-cn_topic_0000002171521445_ul55713519410"></a><a name="zh-cn_topic_0000002171521445_ul55713519410"></a><ul id="zh-cn_topic_0000002171521445_ul55713519410"><li>故障ID为81078603的参数面网络故障默认值为60。</li><li>其余故障默认值为0。</li></ul>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row32027446576"><td class="cellrowborder" valign="top" width="17.27172717271727%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p1620214417579"><a name="zh-cn_topic_0000002171521445_p1620214417579"></a><a name="zh-cn_topic_0000002171521445_p1620214417579"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.082208220822082%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p520254435718"><a name="zh-cn_topic_0000002171521445_p520254435718"></a><a name="zh-cn_topic_0000002171521445_p520254435718"></a>FaultHandling</p>
</td>
<td class="cellrowborder" valign="top" width="60.64606460646065%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p11180846751"><a name="zh-cn_topic_0000002171521445_p11180846751"></a><a name="zh-cn_topic_0000002171521445_p11180846751"></a>超过故障持续时间后的故障处理策略，支持配置不同级别的故障处理策略。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note19791226361"><a name="zh-cn_topic_0000002171521445_note19791226361"></a><div class="notebody"><p id="zh-cn_topic_0000002171521445_p179116261168"><a name="zh-cn_topic_0000002171521445_p179116261168"></a><a name="zh-cn_topic_0000002171521445_p179116261168"></a>超过故障持续时间后的故障处理策略，建议高于故障本身的故障处理策略，否则配置不生效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row1297682783618"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p1496743143619"><a name="zh-cn_topic_0000002171521445_p1496743143619"></a><a name="zh-cn_topic_0000002171521445_p1496743143619"></a>注</p>
<a name="zh-cn_topic_0000002171521445_ul181621184612"></a><a name="zh-cn_topic_0000002171521445_ul181621184612"></a><ul id="zh-cn_topic_0000002171521445_ul181621184612"><li>如果一个故障码同时配置了故障频率（FaultFrequency）和故障超时策略（FaultDuration），该故障码在TimeWindow时间窗口中超时次数达到任务支持的最大次数时，则采用以下三者中最严重的等级进行处理。这三者分别为：故障本身的故障处理策略、FaultFrequency和FaultDuration中配置的故障处理策略。</li><li>如果一个故障码同时配置了故障频率和故障超时策略，只有当故障超时后，故障频次才会增加一次。</li><li>故障ID为81078603的网络故障只支持配置为NotHandleFault、PreSeparateNPU或SeparateNPU三种故障处理策略，若配置为其他策略则使用默认配置NotHandleFault。</li></ul>
</td>
</tr>
</tbody>
</table>


#### （可选）配置芯片故障级别<a name="ZH-CN_TOPIC_0000002479226532"></a>

在制作Ascend Device Plugin镜像时，会将faultCode.json和faultCustomization.json配置文件内置在镜像中，启动Ascend Device Plugin时会读取这两个文件的默认配置，作为当前故障处理依据。faultCode.json和faultCustomization.json的说明请参见[配置文件说明](#配置文件说明-1)。

如果用户想要自定义故障级别或者优雅容错相关配置，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

-   如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
-   如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认faultCode.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。
-   若想要使用faultCode.json或faultCustomization.json的默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认faultCode.json、SwitchFaultCode.json或faultCustomization.json文件。
-   如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

**使用faultCode.json配置故障级别<a name="zh-cn_topic_0000001951258609_section112139052513"></a>**

以故障名称dmp\_daemon节点状态检测异常，对应故障码80E21007为例。将当前故障的处理策略NotHandleFaultCodes（无需处理）修改为RestartNPUCodes（隔离芯片，进行任务重调度）的操作示例如下。

1.  登录环境，进入Ascend Device Plugin解压目录。
2.  执行以下命令，创建动态配置故障码所需ConfigMap文件（mindx-dl-fault-config）。

    ```
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json
    ```

    回显示例如下。

    ```
    configmap/mindx-dl-fault-config created
    ```

    **表 1**  参数说明

    <a name="zh-cn_topic_0000001951258609_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001951258609_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951258609_p16631548171910"><a name="zh-cn_topic_0000001951258609_p16631548171910"></a><a name="zh-cn_topic_0000001951258609_p16631548171910"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951258609_p1663144816197"><a name="zh-cn_topic_0000001951258609_p1663144816197"></a><a name="zh-cn_topic_0000001951258609_p1663144816197"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951258609_p775918210209"><a name="zh-cn_topic_0000001951258609_p775918210209"></a><a name="zh-cn_topic_0000001951258609_p775918210209"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001951258609_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1863164816197"><a name="zh-cn_topic_0000001951258609_p1863164816197"></a><a name="zh-cn_topic_0000001951258609_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1063194861910"><a name="zh-cn_topic_0000001951258609_p1063194861910"></a><a name="zh-cn_topic_0000001951258609_p1063194861910"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p157595292015"><a name="zh-cn_topic_0000001951258609_p157595292015"></a><a name="zh-cn_topic_0000001951258609_p157595292015"></a>动态配置故障码所需的<span id="zh-cn_topic_0000001951258609_ph126311642183015"><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1963194819195"><a name="zh-cn_topic_0000001951258609_p1963194819195"></a><a name="zh-cn_topic_0000001951258609_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p76316488192"><a name="zh-cn_topic_0000001951258609_p76316488192"></a><a name="zh-cn_topic_0000001951258609_p76316488192"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p276092142019"><a name="zh-cn_topic_0000001951258609_p276092142019"></a><a name="zh-cn_topic_0000001951258609_p276092142019"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p964144891914"><a name="zh-cn_topic_0000001951258609_p964144891914"></a><a name="zh-cn_topic_0000001951258609_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1164748191916"><a name="zh-cn_topic_0000001951258609_p1164748191916"></a><a name="zh-cn_topic_0000001951258609_p1164748191916"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p876012211206"><a name="zh-cn_topic_0000001951258609_p876012211206"></a><a name="zh-cn_topic_0000001951258609_p876012211206"></a>不指定该参数则默认取值为300s。用于指定查询mindx-dl-fault-config文件是否更新的周期时间，单位为秒，取值范围为30~3600。PollInterval的修改将在下一个周期生效。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1964748141915"><a name="zh-cn_topic_0000001951258609_p1964748141915"></a><a name="zh-cn_topic_0000001951258609_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p10641648191915"><a name="zh-cn_topic_0000001951258609_p10641648191915"></a><a name="zh-cn_topic_0000001951258609_p10641648191915"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p147602211206"><a name="zh-cn_topic_0000001951258609_p147602211206"></a><a name="zh-cn_topic_0000001951258609_p147602211206"></a>用于保存故障码，必须与faultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

3.  执行以下命令，编辑mindx-dl-fault-config文件。

    ```
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

4.  在mindx-dl-fault-config文件中，找到故障码80E21007。

    ```
    "NotHandleFaultCodes":[
        
    "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    ```

    >[!NOTE] 说明 
    >同一故障码配置在多个故障级别中，会显示设置成功，但默认按照高等级故障处理。

5.  将故障码80E21007在（NotHandleFaultCodes）中删除，并添加到（RestartNPUCodes）中。

    ```
    "NotHandleFaultCodes":[ 
         "80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    "RestartNPUCodes":[
       "8C204E00","A8028802","A4302003","A4302004","A4302005","A4302006","A4302009","A430200A","80CF8009","80CF8008","80E21007",... 
    ...
       ],
    ```

6.  修改完成后，按“Esc”键，输入:wq!保存并退出。
7.  等mindx-dl-fault-config文件更新生效（PollInterval取值，不指定则为300s）后，查看操作是否成功。
    1.  执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2.  通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        若日志出现“load fault code from configmap success”，表示手动配置故障码操作成功。


#### （可选）配置芯片故障频率及时长<a name="ZH-CN_TOPIC_0000002511426473"></a>

在制作Ascend Device Plugin镜像时，会将faultCode.json和faultCustomization.json配置文件内置在镜像中，启动Ascend Device Plugin时会读取这两个文件的默认配置，作为当前故障处理依据。faultCode.json和faultCustomization.json的说明请参见[配置文件说明](#配置文件说明-1)。

如果用户想要自定义芯片故障频率及时长，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

-   如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
-   如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认faultCustomization.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。若想要使用faultCustomization.json的默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认faultCustomization.json文件。
-   如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

**操作步骤<a name="section141902103110"></a>**

以故障码80CB8002为例，如果某张芯片反复发生80CB8002故障，导致训练业务反复重调度，可以手动配置24小时内任务支持的断点续训最大次数为2，达到最大次数后故障的处理策略为ManuallySeparateNPU。

1.  登录环境，进入Ascend Device Plugin解压目录。
2.  执行以下命令，查询是否已经基于faultCode.json文件创建了mindx-dl-fault-config。

    ```
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    -   如果mindx-dl-fault-config已经存在，且存在faultCustomization.json的相关字段，执行[4](#zh-cn_topic_0000002136360238_li38432520129)编辑该文件。
    -   如果mindx-dl-fault-config已经存在，但是不存在faultCustomization.json的相关字段，需要先保存mindx-dl-fault-config内容，再删除mindx-dl-fault-config文件后，执行[3](#zh-cn_topic_0000002136360238_li1946014413123)创建该文件。
    -   如果不存在mindx-dl-fault-config，执行[3](#zh-cn_topic_0000002136360238_li1946014413123)创建该文件。

3.  <a name="zh-cn_topic_0000002136360238_li1946014413123"></a>执行以下命令，创建配置芯片故障频率及时长所需ConfigMap文件（mindx-dl-fault-config）。

    ```
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json --from-file=./faultCustomization.json
    ```

    回显示例如下。

    ```
    configmap/mindx-dl-fault-config created
    ```

    **表 1**  参数说明

    <a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a>动态配置故障码所需的<span id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a>不指定该参数则默认取值为300s。用于指定查询mindx-dl-fault-config文件是否更新的周期时间，单位为秒，取值范围为30~3600。PollInterval的修改将在下一个周期生效。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a>用于保存故障码，必须与faultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_row9289716194614"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a>faultCustomization.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a>用于自定义优雅容错时间、故障频率、故障持续时间（仅支持参数面网络故障）等配置，不指定该参数则没有故障频率配置，其余配置使用默认值进行处理。必须与faultCustomization.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

4.  <a name="zh-cn_topic_0000002136360238_li38432520129"></a>执行以下命令，编辑mindx-dl-fault-config文件。

    ```
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

    根据实际情况，修改芯片的故障频率和时长。

    ```
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
    PollInterval: "300"
    # 修改芯片故障的故障级别
    faultCode.json: |
    {
    "NotHandleFaultCodes":[
    ...
    }
    # 修改芯片故障的故障频率和时长
    faultCustomization.json: |
    {
      "GraceTolerance": {
        "WaitProcessReadCMTime": 30,
        "WaitDeviceResetTime": 150,
        "WaitFaultSelfHealingTime": 15
    },
      "FaultFrequency": [
    {
        "EventId": [
          "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
          "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
          "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
          "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
          "81578008","815F8008","81938008","81958008","81978008"
    ],
        "TimeWindow": 86400,
        "Times": 2,
        "FaultHandling": "ManuallySeparateNPU"
    },
    {
    "EventId": ["80E18005"],
    "TimeWindow": 86400,
    "Times": 3,
    "FaultHandling": "ManuallySeparateNPU"
    }
    ],
      "FaultDuration": [
    {
        "EventId": ["81078603"],
        "FaultTimeout": 20,
        "RecoverTimeout": 60,
        "FaultHandling": "PreSeparateNPU"
    }
    ]
    }
    kind: ConfigMap
    metadata:
    creationTimestamp: "2024-06-20T10:12:07Z"
    name: mindx-dl-fault-config
    namespace: kube-system
    resourceVersion: "52893696"
    selfLink: /api/v1/namespaces/kube-system/configmaps/mindx-dl-fault-config
    uid: bba9e17f-41dd-43b3-848e-3d29cb8c595a
    ```

5.  在mindx-dl-fault-config文件中，在FaultFrequency字段下新增以下代码，设置80CB8002故障在24小时内任务支持的断点续训最大次数为2，达到最大次数后故障的处理策略为ManuallySeparateNPU。

    ```
    {
      "EventId": ["80CB8002"],
      "TimeWindow": 86400,
      "Times": 2,      
      "FaultHandling": "ManuallySeparateNPU"
    }
    ```

6.  修改完成后，按“Esc”键，输入:wq!保存并退出。
7.  等mindx-dl-fault-config文件更新生效（PollInterval取值，不指定则为300s）后，查看操作是否成功。
    1.  执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2.  通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        >[!NOTE] 说明 
        >-   若日志出现“load fault customization from configmap complete”，表示手动配置故障频率操作成功。
        >-   若日志出现“modify  _xxx_  success”，表示ConfigMap中faultCustomization.json里的_xxx_参数设置成功。
        >-   若日志出现“insert fault frequency success”，表示记录了一次频率故障发生时间，在频率窗口内，该卡的该故障记录次数达到频率故障触发次数以后，就会上报频率故障对应的故障级别。

8.  （可选）手动恢复强制隔离的芯片。故障的处理策略为ManuallySeparateNPU时，故障恢复后该芯片也处于隔离状态，需要手动恢复强制隔离的芯片。
    1.  执行以下命令，查找该节点的Ascend Device Plugin上报的device-info-cm。

        ```
        kubectl get cm -n kube-system | grep deviceinfo | grep {nodeName}
        ```

    2.  执行以下命令，编辑该device-info-cm。

        ```
        kubectl edit cm -n kube-system {configMapName}
        ```

    3.  将data下面的ManuallySeparateNPU后面已恢复健康的芯片名称删除。

        ```
        apiVersion: v1
        kind: ConfigMap
        data:
          DeviceInfoCfg: '{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1718702470},"CheckCode":"4f00cf1d220da26a8fdbeb5ba163a751d4b264c48b81d22149257e272ae3b413"}'
          ManuallySeparateNPU: Ascend910-0  
        ```

        >[!NOTE] 说明 
        >删除ManuallySeparateNPU字段后所有芯片名称，并将取值设置为空“”。

    4.  修改完成后，按“Esc”键，输入:wq!保存并退出。
    5.  等待1个上报周期（若设备信息有变化，那么在健康状态检查周期内就会上报，如果设备信息没有变化，那么上报周期固定为5分钟）后，执行以下命令，查看device-info-cm中ManuallySeparateNPU是否存在刚才删除的芯片名称。若不存在，则芯片恢复健康成功，可继续正常使用该芯片。

        ```
        kubectl describe cm -n kube-system {configMapName}
        ```



### 参数面网络故障<a name="ZH-CN_TOPIC_0000002479226486"></a>

#### 总线设备故障<a name="ZH-CN_TOPIC_0000002511346423"></a>

##### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346513"></a>

针对**总线设备**故障的**不同级别**进行分级处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**SwitchFaultCode.json**中故障码配置的故障级别，对故障进行相应处理。SwitchFaultCode.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码对应的故障级别，可以通过由faultCode.json和SwitchFaultCode.json创建的**mindx-dl-fault-config**文件实现。

>[!NOTE] 说明 
>只有Atlas A3 训练系列产品存在**总线设备**，该设备的故障码可以查看SwitchFaultCode.json文件**。**

**SwitchFaultCode.json中的故障级别<a name="section681495612012"></a>**

断点续训针对**总线设备**故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置总线设备故障级别](#可选配置总线设备故障级别)。

Ascend Device Plugin从驱动获取到故障码后，将根据故障码对设备及业务的影响将故障划分为以下五种级别并进行相应的重调度处理，详细说明请参见[表1](#table212253274720)。

**表 1**  故障级别及处理说明

<a name="table212253274720"></a>
<table><thead align="left"><tr id="row0123203211474"><th class="cellrowborder" valign="top" width="19.148085191480853%" id="mcps1.2.4.1.1"><p id="p17123193212474"><a name="p17123193212474"></a><a name="p17123193212474"></a>故障类型</p>
</th>
<th class="cellrowborder" valign="top" width="44.81551844815518%" id="mcps1.2.4.1.2"><p id="p3123532194719"><a name="p3123532194719"></a><a name="p3123532194719"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="36.036396360363966%" id="mcps1.2.4.1.3"><p id="p6123123216475"><a name="p6123123216475"></a><a name="p6123123216475"></a>重调度处理</p>
</th>
</tr>
</thead>
<tbody><tr id="row41231732164712"><td class="cellrowborder" valign="top" width="19.148085191480853%" headers="mcps1.2.4.1.1 "><p id="p13123193219471"><a name="p13123193219471"></a><a name="p13123193219471"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="44.81551844815518%" headers="mcps1.2.4.1.2 "><p id="p5123183224712"><a name="p5123183224712"></a><a name="p5123183224712"></a>暂不影响业务，可以自行恢复，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="36.036396360363966%" headers="mcps1.2.4.1.3 "><p id="p19980163485813"><a name="p19980163485813"></a><a name="p19980163485813"></a>暂不处理。</p>
</td>
</tr>
<tr id="row184593196494"><td class="cellrowborder" valign="top" width="19.148085191480853%" headers="mcps1.2.4.1.1 "><p id="p15459191984920"><a name="p15459191984920"></a><a name="p15459191984920"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="44.81551844815518%" headers="mcps1.2.4.1.2 "><p id="p169108334372"><a name="p169108334372"></a><a name="p169108334372"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="#yaml参数说明">YAML参数说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="36.036396360363966%" headers="mcps1.2.4.1.3 "><p id="p6792114863712"><a name="p6792114863712"></a><a name="p6792114863712"></a>当芯片出现亚健康故障时，需根据<a href="#任务yaml配置示例">任务YAML配置示例</a>策略进行处理。</p>
<div class="note" id="note379214817373"><a name="note379214817373"></a><div class="notebody"><p id="p117921248133718"><a name="p117921248133718"></a><a name="p117921248133718"></a>如果后续芯片出现其他级别故障，此时</p>
<p id="p879214843715"><a name="p879214843715"></a><a name="p879214843715"></a>SubHealthFault处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
</tr>
<tr id="row13451544203113"><td class="cellrowborder" valign="top" width="19.148085191480853%" headers="mcps1.2.4.1.1 "><p id="p7345144143117"><a name="p7345144143117"></a><a name="p7345144143117"></a>RestartRequestFault</p>
</td>
<td class="cellrowborder" valign="top" width="44.81551844815518%" headers="mcps1.2.4.1.2 "><p id="p5345144193111"><a name="p5345144193111"></a><a name="p5345144193111"></a>业务运行失败，需要重新执行业务请求。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="36.036396360363966%" headers="mcps1.2.4.1.3 "><p id="p123451744133111"><a name="p123451744133111"></a><a name="p123451744133111"></a>停止当前训练任务，隔离节点，进行任务重调度。</p>
</td>
</tr>
<tr id="row1137117255497"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p143725254498"><a name="p143725254498"></a><a name="p143725254498"></a>ResetFault</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p14372132514919"><a name="p14372132514919"></a><a name="p14372132514919"></a>业务运行失败。</p>
</td>
</tr>
<tr id="row13514203017499"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p5514183044914"><a name="p5514183044914"></a><a name="p5514183044914"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p14514130124912"><a name="p14514130124912"></a><a name="p14514130124912"></a>业务运行失败，需更换器件或板卡。</p>
</td>
</tr>
</tbody>
</table>


##### （可选）配置总线设备故障级别<a name="ZH-CN_TOPIC_0000002511426433"></a>

在制作Ascend Device Plugin镜像时，会将故障级别配置文件**SwitchFaultCode.json**内置在镜像中，启动Ascend Device Plugin时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别或者优雅容错相关配置，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

-   如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
-   如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认**SwitchFaultCode.json**将不会生效，使用集群中已经存在的mindx-dl-fault-config。
-   如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config且该ConfigMap中存在SwitchFaultCode.json字段，Ascend Device Plugin的默认SwitchFaultCode.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。
-   若想要使用SwitchFaultCode.json默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认SwitchFaultCode.json文件。
-   如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

**使用SwitchFaultCode.json配置故障级别<a name="section067783615137"></a>**

以总线设备故障码\[0x00f1ff09,155913,cpu,na\]为例。该故障码由四部分组成：告警ID、故障ID、对端设备类型、端口号，如[表1 故障码说明](#zh-cn_topic_0000002007978080_table167355241939)所示。

**表 1**  故障码说明

<a name="zh-cn_topic_0000002007978080_table167355241939"></a>
|参数|说明|取值|
|--|--|--|
|告警ID|在以上示例中，告警ID为0x00f1ff09。|带内带外一致。|
|故障ID|在以上示例中，故障ID为155913。|带内带外一致。|
|对端设备类型|该故障所对应的对端设备类型。在以上示例中，对端设备类型为cpu。|<ul><li>取值为na：该故障为芯片故障，不涉及对端设备。</li><li>取值为cpu：该故障所对应的对端设备为CPU。</li><li>取值为npu：该故障所对应的对端设备为NPU。</li><li>取值为L2：该故障所对应的对端设备为L2。</li></ul>|
|端口号|在以上示例中，端口号为na。|取值只能为na。|

将当前故障的处理策略NotHandleFaultCodes（无需处理）修改为SeparateFaultCodes（隔离芯片，进行任务重调度）的操作示例如下。

1.  登录环境，进入Ascend Device Plugin解压目录。
2.  执行以下命令，查询是否已经基于SwitchFaultCode.json文件创建了mindx-dl-fault-config。

    ```
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    -   如果mindx-dl-fault-config已经存在，且存在SwitchFaultCode.json的相关字段，执行[4](#zh-cn_topic_0000002007978080_li1014819812423)编辑该文件。
    -   如果mindx-dl-fault-config已经存在，但是不存在SwitchFaultCode.json的相关字段，需要先保存mindx-dl-fault-config内容，再删除mindx-dl-fault-config文件后，执行[3](#zh-cn_topic_0000002007978080_li14147485427)创建该文件。
    -   如果不存在mindx-dl-fault-config，执行[3](#zh-cn_topic_0000002007978080_li14147485427)创建该文件。

3.  <a name="zh-cn_topic_0000002007978080_li14147485427"></a>执行以下命令，创建动态配置故障码所需ConfigMap文件（mindx-dl-fault-config）。

    ```
    kubectl create cm mindx-dl-fault-config -n kube-system  --from-file=./faultCode.json --from-file=./SwitchFaultCode.json --from-literal="PollInterval=300"
    ```

    回显示例如下。

    ```
    configmap/mindx-dl-fault-config created
    ```

    **表 2**  参数说明

    <a name="zh-cn_topic_0000002007978080_table14147138184211"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002007978080_row1814716812426"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002007978080_p141471483423"><a name="zh-cn_topic_0000002007978080_p141471483423"></a><a name="zh-cn_topic_0000002007978080_p141471483423"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002007978080_p101477811428"><a name="zh-cn_topic_0000002007978080_p101477811428"></a><a name="zh-cn_topic_0000002007978080_p101477811428"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002007978080_p1014718154210"><a name="zh-cn_topic_0000002007978080_p1014718154210"></a><a name="zh-cn_topic_0000002007978080_p1014718154210"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002007978080_row1514810811424"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p714817819421"><a name="zh-cn_topic_0000002007978080_p714817819421"></a><a name="zh-cn_topic_0000002007978080_p714817819421"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p201488804220"><a name="zh-cn_topic_0000002007978080_p201488804220"></a><a name="zh-cn_topic_0000002007978080_p201488804220"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p161481689426"><a name="zh-cn_topic_0000002007978080_p161481689426"></a><a name="zh-cn_topic_0000002007978080_p161481689426"></a>动态配置故障码所需的<span id="zh-cn_topic_0000002007978080_ph214819813425"><a name="zh-cn_topic_0000002007978080_ph214819813425"></a><a name="zh-cn_topic_0000002007978080_ph214819813425"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002007978080_row814819819422"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p101481589424"><a name="zh-cn_topic_0000002007978080_p101481589424"></a><a name="zh-cn_topic_0000002007978080_p101481589424"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p214814819427"><a name="zh-cn_topic_0000002007978080_p214814819427"></a><a name="zh-cn_topic_0000002007978080_p214814819427"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p1614814815424"><a name="zh-cn_topic_0000002007978080_p1614814815424"></a><a name="zh-cn_topic_0000002007978080_p1614814815424"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002007978080_row1714868114215"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p182611591222"><a name="zh-cn_topic_0000002007978080_p182611591222"></a><a name="zh-cn_topic_0000002007978080_p182611591222"></a>SwitchFaultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p1314868184217"><a name="zh-cn_topic_0000002007978080_p1314868184217"></a><a name="zh-cn_topic_0000002007978080_p1314868184217"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p17148118174218"><a name="zh-cn_topic_0000002007978080_p17148118174218"></a><a name="zh-cn_topic_0000002007978080_p17148118174218"></a>用于保存故障码，必须与SwitchFaultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

4.  <a name="zh-cn_topic_0000002007978080_li1014819812423"></a>执行以下命令，编辑mindx-dl-fault-config文件。

    ```
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

5.  在mindx-dl-fault-config文件中，找到故障码\[0x00f1ff09,155913,cpu,na\]。

    ```
    Data
    ====
    SwitchFaultCode.json:
    ----
    {"NotHandleFaultCodes":[0x00f1ff09,155913,cpu,na],
    ...
    ```

6.  将故障码在（NotHandleFaultCodes）中删除，并添加到（SeparateFaultCodes）中。

    ```
    Data
    ====
    SwitchFaultCode.json:
    ----
    {"NotHandleFaultCodes":[],
    ```

    ```
    ...
    "SeparateFaultCodes":["0x00f1ff09,155913,cpu,na","[0x00f103b0,155907,na,na]"…]
    }
    ```

7.  修改完成后，按“Esc”键，输入:wq!保存并退出。
8.  等mindx-dl-fault-config文件更新生效后（PollInterval取值，不指定则为300s），查看操作是否成功。
    1.  执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2.  通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        若日志出现“load switch fault code from configmap success”，表示手动配置故障码操作成功。



#### 关联故障<a name="ZH-CN_TOPIC_0000002511426403"></a>

##### 配置文件说明<a name="ZH-CN_TOPIC_0000002479386560"></a>

断点续训针对关联故障（特殊故障会伴生其他相关联的故障场景），需要忽略特殊故障诱发的伴生故障。ClusterD组件会获取到特殊故障，根据**relationFaultCustomization.json**和**faultDuration.json**文件中配置的关联故障策略对故障任务进行特殊处理。

relationFaultCustomization.json、faultDuration.json为系统配置文件，若用户无特殊需求，请勿随意修改。

**表 1**  relationFaultCustomization文件说明

<a name="zh-cn_topic_0000002157130117_table5148194813113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002157130117_row1614914482114"><th class="cellrowborder" valign="top" width="13.701370137013702%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002157130117_p278365710116"><a name="zh-cn_topic_0000002157130117_p278365710116"></a><a name="zh-cn_topic_0000002157130117_p278365710116"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="69.05690569056905%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002157130117_p127832571915"><a name="zh-cn_topic_0000002157130117_p127832571915"></a><a name="zh-cn_topic_0000002157130117_p127832571915"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="17.241724172417243%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002157130117_p47831857912"><a name="zh-cn_topic_0000002157130117_p47831857912"></a><a name="zh-cn_topic_0000002157130117_p47831857912"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002157130117_row1514912481715"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p14783115717117"><a name="zh-cn_topic_0000002157130117_p14783115717117"></a><a name="zh-cn_topic_0000002157130117_p14783115717117"></a>TriggerFault</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p1878313577120"><a name="zh-cn_topic_0000002157130117_p1878313577120"></a><a name="zh-cn_topic_0000002157130117_p1878313577120"></a>伴生故障码，当前支持faultCode.json和SwitchFaultCode.json配置的故障码。</p>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p117831557615"><a name="zh-cn_topic_0000002157130117_p117831557615"></a><a name="zh-cn_topic_0000002157130117_p117831557615"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row1714944814110"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p6783657411"><a name="zh-cn_topic_0000002157130117_p6783657411"></a><a name="zh-cn_topic_0000002157130117_p6783657411"></a>RelationFaults</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p278411575113"><a name="zh-cn_topic_0000002157130117_p278411575113"></a><a name="zh-cn_topic_0000002157130117_p278411575113"></a>需要被关联的故障列表，可以是一个或多个故障码。当前支持faultCode.json和SwitchFaultCode.json配置的故障码。</p>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p178414571018"><a name="zh-cn_topic_0000002157130117_p178414571018"></a><a name="zh-cn_topic_0000002157130117_p178414571018"></a>字符串列表</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row111493481216"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p578414571818"><a name="zh-cn_topic_0000002157130117_p578414571818"></a><a name="zh-cn_topic_0000002157130117_p578414571818"></a>FaultStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p178405710112"><a name="zh-cn_topic_0000002157130117_p178405710112"></a><a name="zh-cn_topic_0000002157130117_p178405710112"></a>关联故障匹配成功时对应任务的处理策略。</p>
<a name="zh-cn_topic_0000002157130117_ul17849570118"></a><a name="zh-cn_topic_0000002157130117_ul17849570118"></a><ul id="zh-cn_topic_0000002157130117_ul17849570118"><li>Separate：任务隔离</li><li>SubHealth：任务亚健康</li></ul>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p1378413577119"><a name="zh-cn_topic_0000002157130117_p1378413577119"></a><a name="zh-cn_topic_0000002157130117_p1378413577119"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row84116191226"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p114317681616"><a name="zh-cn_topic_0000002157130117_p114317681616"></a><a name="zh-cn_topic_0000002157130117_p114317681616"></a>注：</p>
<p id="zh-cn_topic_0000002157130117_p47413216213"><a name="zh-cn_topic_0000002157130117_p47413216213"></a><a name="zh-cn_topic_0000002157130117_p47413216213"></a>当设备发生配置的RelationFaults时，<span id="zh-cn_topic_0000002157130117_ph12291515161616"><a name="zh-cn_topic_0000002157130117_ph12291515161616"></a><a name="zh-cn_topic_0000002157130117_ph12291515161616"></a>ClusterD</span>会将对应的故障加入待处理的故障码队列。在配置的TimeOutInterval时间内，如果发生了TriggerFault对应的故障，会按照用户配置的FaultStrategy策略对任务进行处理。如果超过配置的TimeOutInterval时间，总线设备故障类型，按照任务亚健康进行处理，芯片故障或者参数面网络故障，会忽略该故障。</p>
</td>
</tr>
</tbody>
</table>

**表 2**  faultDuration.json文件说明

<a name="zh-cn_topic_0000002157130117_table1484617498414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002157130117_row1284615492415"><th class="cellrowborder" valign="top" width="13.36133613361336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002157130117_p116699222514"><a name="zh-cn_topic_0000002157130117_p116699222514"></a><a name="zh-cn_topic_0000002157130117_p116699222514"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="70.36703670367037%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002157130117_p56691922055"><a name="zh-cn_topic_0000002157130117_p56691922055"></a><a name="zh-cn_topic_0000002157130117_p56691922055"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="16.271627162716275%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002157130117_p466911221257"><a name="zh-cn_topic_0000002157130117_p466911221257"></a><a name="zh-cn_topic_0000002157130117_p466911221257"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002157130117_row084615491413"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p066920221954"><a name="zh-cn_topic_0000002157130117_p066920221954"></a><a name="zh-cn_topic_0000002157130117_p066920221954"></a>FaultCode</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p96702227514"><a name="zh-cn_topic_0000002157130117_p96702227514"></a><a name="zh-cn_topic_0000002157130117_p96702227514"></a>故障码，当前支持faultCode.json和SwitchFaultCode.json配置的故障码。</p>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p56701922954"><a name="zh-cn_topic_0000002157130117_p56701922954"></a><a name="zh-cn_topic_0000002157130117_p56701922954"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row18467491043"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p167022212517"><a name="zh-cn_topic_0000002157130117_p167022212517"></a><a name="zh-cn_topic_0000002157130117_p167022212517"></a>FaultType</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p667020225515"><a name="zh-cn_topic_0000002157130117_p667020225515"></a><a name="zh-cn_topic_0000002157130117_p667020225515"></a>故障类型：</p>
<a name="zh-cn_topic_0000002157130117_ul1367017221559"></a><a name="zh-cn_topic_0000002157130117_ul1367017221559"></a><ul id="zh-cn_topic_0000002157130117_ul1367017221559"><li>faultDevice：芯片故障或者参数面网络故障</li><li>faultSwitch：总线设备故障</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p967010221359"><a name="zh-cn_topic_0000002157130117_p967010221359"></a><a name="zh-cn_topic_0000002157130117_p967010221359"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row208478499416"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p16713225511"><a name="zh-cn_topic_0000002157130117_p16713225511"></a><a name="zh-cn_topic_0000002157130117_p16713225511"></a>TimeOutInterval</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p36711221159"><a name="zh-cn_topic_0000002157130117_p36711221159"></a><a name="zh-cn_topic_0000002157130117_p36711221159"></a>故障码最长被关联时间。单位为秒。</p>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p186718221450"><a name="zh-cn_topic_0000002157130117_p186718221450"></a><a name="zh-cn_topic_0000002157130117_p186718221450"></a>整数</p>
</td>
</tr>
</tbody>
</table>


##### （可选）配置关联故障的处理策略<a name="ZH-CN_TOPIC_0000002479226478"></a>

在制作ClusterD镜像时，会将关联故障的两个配置文件内置在镜像中，启动ClusterD会读取这两个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义关联的故障码以及对应的处理策略。可以在制作ClusterD镜像时，修改对应的relationFaultCustomization.json和faultDuration.json配置文件。

**操作步骤<a name="zh-cn_topic_0000002157048501_section2086912531189"></a>**

_以RelationFaults为故障码81078603，TriggerFault为故障码8C1F8609为例。如果发生了芯片81078603的故障码，需要在后面60s内出现8C1F8609故障时忽略8C1F8609故障，并且隔离发生的81078603故障的任务。可以手动配置关联故障的处理策略为Separate。_

1.  登录环境，进入ClusterD解压后的目录。
2.  执行**vi relationFaultCustomization.json**命令编辑**配置文件**。

    ```
    vi relationFaultCustomization.json
    ```

    将2个故障进行关联。修改完成后，按“Esc”键，输入**:wq!**保存并退出。

    ```
    …
      {
        "TriggerFault": "8C1F8609",
        "RelationFaults": [
          "81078603"
        ],
        "FaultStrategy": "Separate"
      }
    …
    ```

3.  执行**vi faultDuration.json命令编辑**配置文件。

    ```
    vi faultDuration.json
    ```

    配置故障类型、故障关联时间等。修改完成后，按“Esc”键，输入**:wq!**保存并退出。

    ```
    …
      {
        "FaultCode": "81078603",
        "FaultType": "faultDevice",
        "TimeOutInterval": 60
      }
    …
    ```




### 公共故障<a name="ZH-CN_TOPIC_0000002479386564"></a>

#### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346487"></a>

断点续训针对公共故障的不同级别进行分级处理。ClusterD组件会获取到当前故障的故障码，根据publicFaultConfiguration.json文件中故障码配置的故障级别，对故障进行相应处理。特殊情况下，若ClusterD收到了无法识别的故障码（未保存在配置文件中），会将此故障丢弃。

[publicFaultConfiguration.json](#zh-cn_topic_0000002181110120_table8202741102717)为公共故障的系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改公共故障的级别和发送方，可以通过在/user1/mindx-dl/clusterd写入自定义配置文件publicCustomization.json实现。该文件路径支持配置，配置方式如下所示。

>[!NOTE] 说明 
>-   文件publicCustomization.json在容器内路径为/user1/mindx-dl/clusterd，不支持修改，不支持软链接；主机路径默认为/user1/mindx-dl/clusterd。
>-   主机路径可由用户根据实际情况自行配置：在ClusterD的启动YAML中修改挂载卷名称为config-clusterd的主机挂载路径。
>-   多master场景下，建议每个master节点上都同步一份最新的publicCustomization.json文件。避免重启ClusterD后，ClusterD被调度到其他master节点，从而导致自定义故障配置文件丢失的问题。

**表 1**  故障级别及处理说明

<a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_table169151711124319"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row19916131120434"><th class="cellrowborder" valign="top" width="15.09499941718149%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a>故障级别</p>
</th>
<th class="cellrowborder" valign="top" width="42.54575125305979%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="42.35924932975871%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p2218314171716"><a name="zh-cn_topic_0000002181110120_p2218314171716"></a><a name="zh-cn_topic_0000002181110120_p2218314171716"></a>重调度处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row6916711144312"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a>无需处理</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row1991661104316"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p12165431710"><a name="zh-cn_topic_0000002181110120_p12165431710"></a><a name="zh-cn_topic_0000002181110120_p12165431710"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row191716112431"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="#yaml参数说明">YAML参数说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="#任务yaml配置示例">任务YAML配置示例</a>策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"></a><div class="notebody"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
</tr>
<tr id="row16800523414"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="p88011823817"><a name="p88011823817"></a><a name="p88011823817"></a><span id="ph1339214581915"><a name="ph1339214581915"></a><a name="ph1339214581915"></a>PreSeparateNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="p980117231413"><a name="p980117231413"></a><a name="p980117231413"></a><span id="ph739245817113"><a name="ph739245817113"></a><a name="ph739245817113"></a>暂不影响业务，后续不再调度任务到该芯片。</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="p1280114235116"><a name="p1280114235116"></a><a name="p1280114235116"></a><span id="ph3392758212"><a name="ph3392758212"></a><a name="ph3392758212"></a>预隔离芯片。</span></p>
</td>
</tr>
</tbody>
</table>

**表 2**  publicFaultConfiguration.json字段说明

<a name="zh-cn_topic_0000002181110120_table8202741102717"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row18202164117272"><th class="cellrowborder" valign="top" width="28.93%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p1120213413271"><a name="zh-cn_topic_0000002181110120_p1120213413271"></a><a name="zh-cn_topic_0000002181110120_p1120213413271"></a>参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="71.07%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p22024417279"><a name="zh-cn_topic_0000002181110120_p22024417279"></a><a name="zh-cn_topic_0000002181110120_p22024417279"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row172028412278"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1220219412279"><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a href="#zh-cn_topic_0000002181110120_table1689274753416">publicFaultCode</a></p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p220284110271"><a name="zh-cn_topic_0000002181110120_p220284110271"></a><a name="zh-cn_topic_0000002181110120_p220284110271"></a>公共故障码相关配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row14606121802219"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1760617182224"><a name="zh-cn_topic_0000002181110120_p1760617182224"></a><a name="zh-cn_topic_0000002181110120_p1760617182224"></a>publicFaultResource</p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p1606118102218"><a name="zh-cn_topic_0000002181110120_p1606118102218"></a><a name="zh-cn_topic_0000002181110120_p1606118102218"></a>公共故障发送方配置。</p>
</td>
</tr>
</tbody>
</table>

**表 3**  publicFaultCode字段说明

<a name="zh-cn_topic_0000002181110120_table1689274753416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row16892144733413"><th class="cellrowborder" valign="top" width="28.849999999999998%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p689264723412"><a name="zh-cn_topic_0000002181110120_p689264723412"></a><a name="zh-cn_topic_0000002181110120_p689264723412"></a>参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="71.15%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p889274783418"><a name="zh-cn_topic_0000002181110120_p889274783418"></a><a name="zh-cn_topic_0000002181110120_p889274783418"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row28921647103410"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p48921847143412"><a name="zh-cn_topic_0000002181110120_p48921847143412"></a><a name="zh-cn_topic_0000002181110120_p48921847143412"></a>NotHandleFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p58921747183416"><a name="zh-cn_topic_0000002181110120_p58921747183416"></a><a name="zh-cn_topic_0000002181110120_p58921747183416"></a>故障级别为NotHandleFault（无需处理）的故障码。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row989224719346"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p118928476343"><a name="zh-cn_topic_0000002181110120_p118928476343"></a><a name="zh-cn_topic_0000002181110120_p118928476343"></a>SubHealthFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p17892947113410"><a name="zh-cn_topic_0000002181110120_p17892947113410"></a><a name="zh-cn_topic_0000002181110120_p17892947113410"></a>故障级别为SubHealthFault（亚健康）的故障码。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row289264713349"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p38921547193418"><a name="zh-cn_topic_0000002181110120_p38921547193418"></a><a name="zh-cn_topic_0000002181110120_p38921547193418"></a>SeparateNPUCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p689274714341"><a name="zh-cn_topic_0000002181110120_p689274714341"></a><a name="zh-cn_topic_0000002181110120_p689274714341"></a>故障级别为SeparateNPU（无法恢复，需要隔离芯片）的故障码。</p>
</td>
</tr>
<tr id="row107385344217"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="p187397341724"><a name="p187397341724"></a><a name="p187397341724"></a><span id="ph791817016319"><a name="ph791817016319"></a><a name="ph791817016319"></a>PreSeparateNPUCodes</span></p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="p15739113415210"><a name="p15739113415210"></a><a name="p15739113415210"></a><span id="ph8918120234"><a name="ph8918120234"></a><a name="ph8918120234"></a>故障级别为</span><span id="ph491890639"><a name="ph491890639"></a><a name="ph491890639"></a>PreSeparateNPU</span><span id="ph6918601336"><a name="ph6918601336"></a><a name="ph6918601336"></a>（暂不影响业务，后续不再调度任务到该芯片）的故障码。</span></p>
</td>
</tr>
</tbody>
</table>

**故障码说明<a name="zh-cn_topic_0000002181110120_section1440314273418"></a>**

公共故障的故障码为9位，说明如下。

**表 4**  故障码说明

<a name="table1237891465117"></a>
<table><thead align="left"><tr id="row1137891413516"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p1937816143519"><a name="p1937816143519"></a><a name="p1937816143519"></a>位数</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p837812144514"><a name="p837812144514"></a><a name="p837812144514"></a>描述</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p14378201455110"><a name="p14378201455110"></a><a name="p14378201455110"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="row1137861419517"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p123782149514"><a name="p123782149514"></a><a name="p123782149514"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15378914185120"><a name="p15378914185120"></a><a name="p15378914185120"></a>故障类型</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p10378161419517"><a name="p10378161419517"></a><a name="p10378161419517"></a>0：芯片故障</p>
<p id="p037871414515"><a name="p037871414515"></a><a name="p037871414515"></a>1：节点故障</p>
<p id="p33781414125113"><a name="p33781414125113"></a><a name="p33781414125113"></a>2：网络故障</p>
<p id="p10379101414516"><a name="p10379101414516"></a><a name="p10379101414516"></a>3：存储故障</p>
</td>
</tr>
<tr id="row337901415519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2379181475111"><a name="p2379181475111"></a><a name="p2379181475111"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p133796146513"><a name="p133796146513"></a><a name="p133796146513"></a>故障默认的级别</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p103791114185115"><a name="p103791114185115"></a><a name="p103791114185115"></a>0: NotHandleFault</p>
<p id="p193791214175112"><a name="p193791214175112"></a><a name="p193791214175112"></a>1: SubHealthFault</p>
<p id="p737991475119"><a name="p737991475119"></a><a name="p737991475119"></a>2: SeparateNPU</p>
</td>
</tr>
<tr id="row1737917147519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p133793145514"><a name="p133793145514"></a><a name="p133793145514"></a>3、4</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1137901435119"><a name="p1137901435119"></a><a name="p1137901435119"></a>预留扩展位</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p43795142516"><a name="p43795142516"></a><a name="p43795142516"></a>暂为00</p>
</td>
</tr>
<tr id="row1337961495114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17379121416515"><a name="p17379121416515"></a><a name="p17379121416515"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p17379141465112"><a name="p17379141465112"></a><a name="p17379141465112"></a>第6-9位的故障码是否为用户自定义，避免冲突</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1237917146517"><a name="p1237917146517"></a><a name="p1237917146517"></a>0：发布包中定义</p>
<p id="p12379191418513"><a name="p12379191418513"></a><a name="p12379191418513"></a>1：用户自定义</p>
</td>
</tr>
<tr id="row1937911425114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p12379161465115"><a name="p12379161465115"></a><a name="p12379161465115"></a>6-9</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1437931425115"><a name="p1437931425115"></a><a name="p1437931425115"></a>具体的十进制故障码</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p8379121413512"><a name="p8379121413512"></a><a name="p8379121413512"></a>示例：1001</p>
</td>
</tr>
<tr id="row6379214165114"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p1137911410515"><a name="p1137911410515"></a><a name="p1137911410515"></a>示例如下：</p>
<p id="p1837941416513"><a name="p1837941416513"></a><a name="p1837941416513"></a>0100 01001：芯片故障，SubHealthFault，发布包中定义，故障1001。</p>
<p id="p1037911455117"><a name="p1037911455117"></a><a name="p1037911455117"></a>1000 11002：节点故障，NotHandleFault，用户自定义，故障1002。</p>
<p id="p8379181455115"><a name="p8379181455115"></a><a name="p8379181455115"></a>2200 01003：网络故障，SeparateNPU，发布包中定义，故障1003。</p>
</td>
</tr>
</tbody>
</table>

**已支持的公共故障<a name="zh-cn_topic_0000002181110120_section4960201383813"></a>**

**表 5**  已支持的公共故障

<a name="zh-cn_topic_0000002181110120_table31451934163811"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row514523493819"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_p1114523420389"><a name="zh-cn_topic_0000002181110120_p1114523420389"></a><a name="zh-cn_topic_0000002181110120_p1114523420389"></a>故障码</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_p9145143412387"><a name="zh-cn_topic_0000002181110120_p9145143412387"></a><a name="zh-cn_topic_0000002181110120_p9145143412387"></a>故障说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p15145193413388"><a name="zh-cn_topic_0000002181110120_p15145193413388"></a><a name="zh-cn_topic_0000002181110120_p15145193413388"></a>默认故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row1514593415388"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p181451134193811"><a name="zh-cn_topic_0000002181110120_p181451134193811"></a><a name="zh-cn_topic_0000002181110120_p181451134193811"></a>010001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p814593412386"><a name="zh-cn_topic_0000002181110120_p814593412386"></a><a name="zh-cn_topic_0000002181110120_p814593412386"></a>光链路脏污（芯片故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p414533483811"><a name="zh-cn_topic_0000002181110120_p414533483811"></a><a name="zh-cn_topic_0000002181110120_p414533483811"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row175241157181818"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p580896101918"><a name="p580896101918"></a><a name="p580896101918"></a>210001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15808166121915"><a name="p15808166121915"></a><a name="p15808166121915"></a>光链路脏污（网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1180917617197"><a name="p1180917617197"></a><a name="p1180917617197"></a>SubHealthFault</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row131782214434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p41752216438"><a name="zh-cn_topic_0000002181110120_p41752216438"></a><a name="zh-cn_topic_0000002181110120_p41752216438"></a>220001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p1171822134316"><a name="zh-cn_topic_0000002181110120_p1171822134316"></a><a name="zh-cn_topic_0000002181110120_p1171822134316"></a>NPU卡<span id="ph17233131243911"><a name="ph17233131243911"></a><a name="ph17233131243911"></a>HCCS</span>网络故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p1566710511444"><a name="zh-cn_topic_0000002181110120_p1566710511444"></a><a name="zh-cn_topic_0000002181110120_p1566710511444"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row192881812184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3289131210473"><a name="p3289131210473"></a><a name="p3289131210473"></a>010001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1628951244719"><a name="p1628951244719"></a><a name="p1628951244719"></a>光链路松动（芯片故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p828971254715"><a name="p828971254715"></a><a name="p828971254715"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row38601828161910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p6168163671911"><a name="p6168163671911"></a><a name="p6168163671911"></a>210001008</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p316816364194"><a name="p316816364194"></a><a name="p316816364194"></a>光链路松动（网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2168436121911"><a name="p2168436121911"></a><a name="p2168436121911"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row172051674711"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p127201168472"><a name="p127201168472"></a><a name="p127201168472"></a>310001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1572071644717"><a name="p1572071644717"></a><a name="p1572071644717"></a>DPC客户端失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p141495491488"><a name="p141495491488"></a><a name="p141495491488"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row4720816104713"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17720131674712"><a name="p17720131674712"></a><a name="p17720131674712"></a>200001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1972020169475"><a name="p1972020169475"></a><a name="p1972020169475"></a>疑似光链路亚健康</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1572061684719"><a name="p1572061684719"></a><a name="p1572061684719"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row191121122184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p7112152234711"><a name="p7112152234711"></a><a name="p7112152234711"></a>210001009</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10112152210476"><a name="p10112152210476"></a><a name="p10112152210476"></a>光模块器件亚健康</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2011213229474"><a name="p2011213229474"></a><a name="p2011213229474"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row19731102610435"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1324180124413"><a name="p1324180124413"></a><a name="p1324180124413"></a>220001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p13241019443"><a name="p13241019443"></a><a name="p13241019443"></a>备份超节点场景下，调度使用不存在的备份框资源。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p12161558145818"><a name="p12161558145818"></a><a name="p12161558145818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row13731626174317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3241309446"><a name="p3241309446"></a><a name="p3241309446"></a>220001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1424190104412"><a name="p1424190104412"></a><a name="p1424190104412"></a>备份框资源端口故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p17362234428"><a name="p17362234428"></a><a name="p17362234428"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row127318268434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p52416011444"><a name="p52416011444"></a><a name="p52416011444"></a>220001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p124800442"><a name="p124800442"></a><a name="p124800442"></a>备份框任务ID占用冲突</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p191615588586"><a name="p191615588586"></a><a name="p191615588586"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row20731826154318"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5241103443"><a name="p5241103443"></a><a name="p5241103443"></a>220001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p824705444"><a name="p824705444"></a><a name="p824705444"></a>NetMind失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1016110586589"><a name="p1016110586589"></a><a name="p1016110586589"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673142624317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p162412019440"><a name="p162412019440"></a><a name="p162412019440"></a>220001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10241109444"><a name="p10241109444"></a><a name="p10241109444"></a>疑似备份框链路端口部分失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p916119583580"><a name="p916119583580"></a><a name="p916119583580"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673116264438"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p9247064419"><a name="p9247064419"></a><a name="p9247064419"></a>220001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6249018447"><a name="p6249018447"></a><a name="p6249018447"></a>光链路调整失败</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p132215501211"><a name="p132215501211"></a><a name="p132215501211"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row8926105693315"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1792695643311"><a name="p1792695643311"></a><a name="p1792695643311"></a>200001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18926165614336"><a name="p18926165614336"></a><a name="p18926165614336"></a>某节点内产生/恢复慢网络（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p24091634153411"><a name="p24091634153411"></a><a name="p24091634153411"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row10526205273417"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3526052153416"><a name="p3526052153416"></a><a name="p3526052153416"></a>200001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5804154074418"><a name="p5804154074418"></a><a name="p5804154074418"></a>超节点内的节点间产生/恢复慢网络。（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p352695212349"><a name="p352695212349"></a><a name="p352695212349"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row663164316353"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17634437355"><a name="p17634437355"></a><a name="p17634437355"></a>200001012</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18631743123513"><a name="p18631743123513"></a><a name="p18631743123513"></a>未收敛到卡（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p101021310163612"><a name="p101021310163612"></a><a name="p101021310163612"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row178327182364"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p383221833611"><a name="p383221833611"></a><a name="p383221833611"></a>110001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1683231816361"><a name="p1683231816361"></a><a name="p1683231816361"></a>慢节点故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p6832131833614"><a name="p6832131833614"></a><a name="p6832131833614"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1179514189380"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p979511810389"><a name="p979511810389"></a><a name="p979511810389"></a>100001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1579521818381"><a name="p1579521818381"></a><a name="p1579521818381"></a>劣化已恢复（慢节点故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p27883220394"><a name="p27883220394"></a><a name="p27883220394"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row121732048142813"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p4359165915289"><a name="p4359165915289"></a><a name="p4359165915289"></a>110001020</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1817494822816"><a name="p1817494822816"></a><a name="p1817494822816"></a>共享存储DPC进程异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p121741348132817"><a name="p121741348132817"></a><a name="p121741348132817"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277135492816"><a name="p5277135492816"></a><a name="p5277135492816"></a>110001021</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277854132820"><a name="p7277854132820"></a><a name="p7277854132820"></a>共享存储DPC内存不足异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182781654132815"><a name="p182781654132815"></a><a name="p182781654132815"></a>SubHealthFault</p>
</td>
</tr>
</tbody>
</table>


#### （可选）配置公共故障的级别和发送方<a name="ZH-CN_TOPIC_0000002479226494"></a>

在制作ClusterD镜像时，会将故障级别配置文件publicFaultConfiguration.json内置在镜像中，启动ClusterD时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别，可以在主机上创建/user1/mindx-dl/clusterd/publicCustomization.json文件。

-   如果ClusterD启动时，已经存在该文件，ClusterD会优先按照已存在的文件中配置的内容，作为当前故障处理依据。
-   如果重新安装ClusterD后，已经存在该文件，ClusterD的默认publicFaultConfiguration.json将不会生效，使用已经存在的publicCustomization.json文件。若想要使用publicFaultConfiguration.json的默认配置，可以删除已存在的publicCustomization.json文件，使ClusterD读取默认的publicFaultConfiguration.json文件。
-   如果publicCustomization.json文件内容存在格式错误等问题，ClusterD会默认读取镜像中内置的publicFaultConfiguration.json文件的内容，作为当前故障处理依据。

**配置公共故障码的故障级别<a name="zh-cn_topic_0000002180950420_section1384121854711"></a>**

配置公共故障码的故障级别分为以下2种场景。

-   对已有故障码的故障级别进行调整。
-   新增故障码及其故障级别。

    下面将以故障码010001008为例，介绍公共故障码故障级别的配置步骤。

1.  登录环境，进入/user1/mindx-dl/clusterd目录。
2.  执行**vi publicCustomization.json**命令，编辑文件。publicCustomization.json的详细说明请参见[表2](#配置文件说明-4)。

    >[!NOTE] 说明 
    >-   创建文件publicCustomization.json之后，用户需要保证该文件有ClusterD用户hwMindX的可读权限。例如，如果用户权限为root，该文件权限建议设置为644。
    >-   文件权限安全需要用户保证，如果权限过大，可能存在安全风险。

    ```
    {
      "publicFaultCode": {
        "NotHandleFaultCodes":[],
        "SubHealthFaultCodes":[],
        "SeparateNPUCodes":["010001008"],
        "PreSeparateNPUCodes":[]
      },
      "publicFaultResource": [
        "CCAE", "fd-online", "pingmesh", "Netmind", "dpcStorage"
      ]
    }
    ```

3.  修改完成后，按“Esc”键，输入:wq!保存并退出。
4.  几秒钟后，文件生效。查看操作是否成功。

    若日志出现“load fault config from <publicCustomization.json\> success”，表示手动配置故障码操作成功。

**配置公共故障的发送方<a name="zh-cn_topic_0000002180950420_section5532327614"></a>**

下面将以新增故障发送方XXX为例，介绍公共故障码发送方的配置步骤。

1.  登录环境，进入/user1/mindx-dl/clusterd目录。
2.  执行**vi publicCustomization.json**命令，编辑文件。publicCustomization.json的详细说明请参见[表2](#配置文件说明-4)。

    ```
    {
      "publicFaultCode": {
        "NotHandleFaultCodes":[],
        "SubHealthFaultCodes":[],
        "SeparateNPUCodes":[],
        "PreSeparateNPUCodes":[]
      },
      "publicFaultResource": [
        "CCAE", "fd-online", "pingmesh", "Netmind", "dpcStorage", "XXX"
      ]
    }
    ```

3.  修改完成后，按“Esc”键，输入:wq!保存并退出。
4.  几秒钟后，文件生效。查看操作是否成功。

    若日志出现“load fault config from <publicCustomization.json\> success”，表示手动配置故障码操作成功。




## 配置故障处理<a name="ZH-CN_TOPIC_0000002479386478"></a>

### 配置Job级别重调度<a name="ZH-CN_TOPIC_0000002479226580"></a>

Job级别重调度默认开启，用户只需完成制作镜像的步骤及准备任务YAML的步骤即可。Job级别重调度的特性介绍、使用约束、支持的产品型号及原理请参见[Job级别重调度](#job级别重调度)。

**准备任务YAML<a name="zh-cn_topic_0000002098814658_section463203519254"></a>**

在任务YAML中，新增以下字段，开启Job级别重调度。

```
... 
metadata:  
   labels:  
     ...  
     fault-scheduling: "force"
```


### 配置Pod级别重调度<a name="ZH-CN_TOPIC_0000002479226508"></a>

本章节将指导用户了解配置Pod级别重调度的关键步骤。Pod级别重调度的特性介绍、使用约束、支持的产品型号及原理请参见[Pod级别重调度](#pod级别重调度)。

**构建镜像<a name="zh-cn_topic_0000002098654822_section11751140165911"></a>**

使用Dockerfile构建容器镜像，新增启动命令。示例如下。

```
# MindCluster断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，请根据实际情况填写   
# 可选，PyTorch框架下，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN pip install $TASKD_WHL 
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用Pod级别重调度需配置以下命令
RUN pip install $TASKD_WHL
```

**准备任务YAML<a name="zh-cn_topic_0000002098654822_section027517423166"></a>**

在任务YAML中，新增以下字段，开启Pod级别重调度，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

```
... 
metadata:  
   labels:  
     ...  
     pod-rescheduling: "on"
     fault-scheduling: "force"
...
ports:                         
   - containerPort: 9601             
     name: taskd-port 
...
```

**适配训练脚本<a name="zh-cn_topic_0000002098654822_section17330181621710"></a>**

1.  在启动脚本（例如train\_start.sh）中，新增以下加粗字段，示例如下。

    ```
    ...
    export MS_ENABLE_TFT='{RSC:1}'      # MindSpore场景下配置此字段开启Pod级别重调度
    ...
    # 可选，PyTorch场景下，设置容器内重启次数和训练进程监控间隔
       logger "server id is: ""${server_id}" 
       if [ "${framework}" == "PyTorch" ]; then 
         get_env_for_pytorch_multi_node_job 
         DISTRIBUTED_ARGS="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT --max_restarts 32767" 
    ```
     其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

2.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 说明 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2.  在训练脚本（例如train\_start.sh）中增加以下代码，拉起TaskD Manager。在以下代码中：

        -   TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。
        -   TASKD\_PROCESS\_ENABLE环境变量配置说明：若任务YAML中“recover-strategy“未配置恢复策略且未使能亚健康热切，需要配置**export TASKD\_PROCESS\_ENABLE="off"**；若“recover-strategy“配置了恢复策略或使能了亚健康热切，则无需配置**export TASKD\_PROCESS\_ENABLE="off"**。

        ```
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="off" 
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python manager.py &           # 具体执行路径由当前路径决定
        fi
        # MindSpore框架下
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
            python manager.py &   # 具体执行路径由当前路径决定
        fi
        ```


### 配置进程级别重调度<a name="ZH-CN_TOPIC_0000002511426407"></a>

本章节将指导用户了解配置进程级别重调度的关键步骤。进程级别重调度的特性介绍、使用约束、支持的产品型号及原理请参见[进程级别重调度](#进程级别重调度)。

**构建镜像<a name="zh-cn_topic_0000002134293721_section18253151810133"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```
# MindCluster无损失断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，MINDIO_TTP_PKG为MindIO的whl安装包的路径，请根据实际情况填写  
# 可选，PyTorch框架下，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN pip3 install $TASKD_WHL   
RUN pip3 install $MINDIO_TTP_PKG 
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用进程级别重调度需配置以下命令
RUN pip3 install $MINDIO_TTP_PKG
RUN pip3 install $TASKD_WHL
```

**准备任务YAML<a name="zh-cn_topic_0000002134293721_section2492121411271"></a>**

在任务YAML中，修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

```
...
ports:                         
   - containerPort: 9601             
     name: taskd-port 
...
```

在任务YAML中，新增以下字段，开启进程级别重调度。recover-strategy是训练进程恢复使用的策略，其中的recover代表开启进程级别恢复。

目前进程级别重调度支持以下2种方式，用户可根据实际使用场景，选择其中一种方式进行使用。

-   方式一：故障后迁移故障Pod到健康节点

    ```
    ...  
    metadata: 
       labels:  
         ...  
         fault-scheduling: "grace"
     ... 
    ...  
       annotations:  
         ...  
         recover-strategy: "recover"   # 任务可用恢复策略（retry：进程级在线恢复；recover：进程级别重调度；recover-in-place: 进程级原地恢复；elastic-training：弹性训练；dump：保存临终遗言；exit：退出训练），6种策略可随意组合，策略之间由逗号分割
     ... 
    ...
    spec:
      replicaSpecs:
        Master:
          template:
            spec:
              containers:
              - name: ascend       # do not modify
                ...
                args:
                  - | 
                    ... 
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
        Worker:
          template:
            spec:
              containers:
              - name: ascend # do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
    ...
    ```

-   方式二：故障后不迁移故障Pod，仅重启故障进程

    ```
    ...  
    metadata: 
       labels:  
         ...  
         fault-scheduling: "grace"
     ... 
    ...  
       annotations:  
         ...  
         recover-strategy: "recover-in-place"   # 任务可用恢复策略（retry：进程级在线恢复；recover：进程级别重调度；recover-in-place: 进程级原地恢复；elastic-training：弹性训练；dump：保存临终遗言；exit：退出训练），6种策略可随意组合，策略之间由逗号分割
     ... 
    ...
    spec:
      replicaSpecs:
        Master:
          template:
            spec:
              containers:
              - name: ascend # do not modify
                ...
                args:
                  - | 
                    ... 
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
        Worker:
          template:
            spec:
              containers:
              - name: ascend # do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
    ...
    ```

**适配训练脚本<a name="zh-cn_topic_0000002134293721_section1829103214273"></a>**

1.  （可选）在启动脚本（例如train\_start.sh）中，配置--max_restarts参数，示例如下。

    ```
    # PyTorch场景下，设置训练进程监控间隔
    ...
       logger "server id is: ""${server_id}"
       if [ "${framework}" == "PyTorch" ]; then
         get_env_for_pytorch_multi_node_job
         DISTRIBUTED_ARGS="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT  --max_restarts 32767"
    ...
    ```
     其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

2.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 说明 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2.  在训练脚本（例如train\_start.sh）中增加以下代码，拉起TaskD Manager。在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

        ```
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="on"
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
           export MASTER_ADDR=${POD_IP}
           python manager.py &           # 具体执行路径由当前路径决定
        fi
        # MindSpore框架下
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
           python manager.py &   # 具体执行路径由当前路径决定
        fi
        ```


### 配置进程级在线恢复<a name="ZH-CN_TOPIC_0000002479386492"></a>

本章节将指导用户了解配置进程级在线恢复的关键步骤。进程级在线恢复的特性介绍、使用约束、支持的产品型号及原理请参见[进程级在线恢复](#进程级在线恢复)。

**构建镜像<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```
# MindCluster断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，MINDIO_TTP_PKG为MindIO的whl安装包的路径，请根据实际情况填写
# 可选，PyTorch框架下，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN pip3 install $TASKD_WHL 
RUN pip3 install $MINDIO_TTP_PKG 
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用进程级在线恢复需配置以下命令
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
```

**准备任务YAML<a name="zh-cn_topic_0000002134174097_section98717593512"></a>**

在任务YAML中，修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

```
...
ports:                         
   - containerPort: 9601             
     name: taskd-port 
...
```

在任务YAML中，新增以下字段，开启进程级别恢复。recover-strategy是训练进程恢复使用的策略，其中的retry代表开启进程级在线恢复。

```
...  
   labels:  
     ...  
     fault-scheduling: "grace"
 ... 
...  
   annotations:  
     ...  
     recover-strategy: "retry"    # 任务可用恢复策略（retry：进程级在线恢复；recover：进程级别重调度；recover-in-place: 进程级原地恢复；elastic-training：弹性训练；dump：保存临终遗言；exit：退出训练），6种策略可随意组合，策略之间由逗号分割
 ... 
...
spec:
  replicaSpecs:
    Master:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            ...
            args:
              - | 
                ... 
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
    Worker:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            ...
            args:
              - |
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
...
```

MindSpore场景下，用户需修改模型参数配置YAML。打开QWEN3\_for\_MS\_code/configs/qwen3/pretrain\_qwen3\_32b\_4k.yaml文件，在代码中增加以下加粗字段。

```
# mindspore context init config
context:
  mode: 0  #0--Graph Mode; 1-Pynative Mode
  device_target: "Ascend"
  graph_kernel_flags: "--disable_pass=cluster.floatstatus_fusion,preprocess.depend_elimination"
  max_call_depth: 10000
  max_device_memory: "59GB"
  mempool_block_size: "59GB"
  save_graphs: True
  save_graphs_path: "./graph"
  device_id: 0
  jit_config:
    jit_level: "O1"
  memory_optimize_level: "00"
  ascend_config:
    hccl_watchdog: False
```

**适配训练脚本<a name="zh-cn_topic_0000002134174097_section189248183358"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 说明 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2.  在训练脚本（例如train\_start.sh）中增加以下代码，拉起TaskD Manager。在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

        ```
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="on"
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
           export MASTER_ADDR=${POD_IP}
           python manager.py &           # 具体执行路径由当前路径决定
        fi
        # MindSpore框架下
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
           python manager.py &   # 具体执行路径由当前路径决定
        fi
        ```

2.  （可选）在启动脚本（例如train\_start.sh）中，新增--max\_restarts参数，示例如下。

    ```
    ... 
       logger "server id is: ""${server_id}" 
       if [ "${framework}" == "PyTorch" ]; then 
         get_env_for_pytorch_multi_node_job 
         DISTRIBUTED_ARGS="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT --max_restarts 32767" 
    ```
        其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

    -   MindSpeed场景下，用户需修改训练启动脚本train\_start.sh，在代码中增加如下加粗字段，示例如下。

        ```
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"   # 开启HCCL算子的重执行特性（算子级在线恢复）。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
        export HCCL_ASYNC_ERROR_HANDLING=0
        ```

    -   MindFormers场景下，用户需修改训练启动脚本msrun\_launcher.sh文件，在代码中增加如下加粗字段，示例如下。

        ```
        export MS_ENABLE_TFT='{UCE:1, HCCE:1}'     # 分别开启片上内存故障进程级在线恢复和网络故障进程级在线恢复
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  # 此环境变量用于配置是否开启HCCL算子的重执行特性。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
        ```

>[!NOTE] 说明 
>用户若要测试进程级在线恢复功能，可参考[进程级在线恢复验证](../appendix.md#进程级在线恢复验证)进行配置。


### 配置算子级在线恢复<a name="ZH-CN_TOPIC_0000002511426477"></a>

本章节将指导用户了解配置算子级在线恢复的关键步骤。算子级在线恢复的特性介绍、使用约束、支持的产品型号及原理请参见[算子级在线恢复](#算子级在线恢复)。

**配置环境变量<a name="section12610013287"></a>**

使用算子级在线恢复前，用户需在启动训练的脚本中配置环境变量HCCL\_OP\_RETRY\_ENABLE和HCCL\_OP\_RETRY\_PARAMS。关于该环境变量的详细说明请参见《CANN 环境变量参考》。配置示例如下。

```
export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"     # 是否开启HCCL算子的重执行特性
export HCCL_OP_RETRY_PARAMS="MaxCnt:3, HoldTime:5000, IntervalTime:1000"    # 配置HCCL算子重执行的具体参数，包括最大重执行次数、第一次重执行的等待时间以及两次重执行的间隔时间
```


### 配置借轨通信任务暂停与回切<a name="ZH-CN_TOPIC_0000002511346495"></a>

#### PyTorch场景（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002511426445"></a>

本章节将指导用户了解配置借轨通信任务暂停与回切的关键步骤。借轨通信任务暂停与回切的特性介绍、使用约束、支持的产品型号及原理请参见[借轨通信任务暂停与回切](#借轨通信任务暂停与回切)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../installation_guide.md#ascend-docker-runtime)、[Ascend Operator](../installation_guide.md#ascend-operator)、[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装[torch\_npu](#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）、[CANN](#制作mindspeed-llm训练镜像pytorch框架)（8.2.RC1及以上版本）、[TaskD](#制作mindspeed-llm训练镜像pytorch框架)和[MindIO](#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）

**操作步骤<a name="section188080175496"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在训练进程内部拉起TaskD  Worker。

    1.  拉起TaskD  Manager。
        1.  创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

            ```
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
            
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
            
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 说明 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

        2.  在训练脚本中增加以下代码，拉起TaskD  Manager。

            ```
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python manager.py &           # 具体执行路径由当前路径决定
            fi
                
            torchrun ...
            ```

    2.  拉起TaskD  Worker。

        修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，在代码中增加如下加粗字段。

        ```
        def pretrain(train_valid_test_dataset_provider,
                     model_provider,
                     model_type,
                     forward_step_func,
                     process_non_loss_data_func=None,
                     extra_args_provider=None,
                     args_defaults={}):
            print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                time.time() - _TRAIN_START_TIME))
            print_datetime('after megatron is initialized')
            import torch.distributed as dist
            if dist.is_initialized():
               rank = dist.get_rank()
               from taskd.api.taskd_worker_api import init_taskd_worker
               from taskd.api.taskd_worker_api import start_taskd_worker
               init_taskd_worker(rank,5000,"pt")
               start_taskd_worker()
            app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
            one_logger_utils.on_pretrain_start()
        ```

    >[!NOTE] 说明 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >```
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >-   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >-   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：  TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。
    >    TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >    ```
    >    pip show taskd
    >    ```

2.  修改训练框架代码。
    1.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3“目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

        ```
        root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
        scripts/
        └── train_start.sh
        ```

    2.  配置训练启动脚本train\_start.sh，在代码中增加如下加粗字段。

        ```
        # 开启HCCL算子的重执行特性。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  
        ```

3.  修改任务YAML。
    1.  在任务YAML中修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ports:                         
           - containerPort: 9601             
             name: taskd-port 
        ```

    2.  在任务YAML中新增以下加粗字段。

        ```
        ...  
           labels:  
             ...  
             fault-scheduling: "force"
         ... 
        ...  
           annotations:  
             ...  
             recover-strategy: "retry"    # 任务可用恢复策略，取值为retry，表示开启进程级在线恢复
         ... 
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - | 
                        ... 
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - |
                        ...
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
        ...
        ```


#### MindSpore场景（基于MindFormers）<a name="ZH-CN_TOPIC_0000002511346443"></a>

本章节将指导用户了解配置借轨通信任务暂停与回切的关键步骤。借轨通信任务暂停与回切的特性介绍、使用约束、支持的产品型号及原理请参见[借轨通信任务暂停与回切](#借轨通信任务暂停与回切)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../installation_guide.md#ascend-docker-runtime)、[Ascend Operator](../installation_guide.md#ascend-operator)、[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装MindSpore（2.7.0及以上版本）、[CANN](#制作mindformers训练镜像mindspore框架)（8.2.RC1及以上版本）、[TaskD](#制作mindformers训练镜像mindspore框架)和[MindIO](#制作mindformers训练镜像mindspore框架)（7.1.RC1及以上版本）

**操作步骤<a name="section9479182019317"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
    1.  拉起TaskD  Manager。
        1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

            ```
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
            
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
            
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 说明 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

        2.  在训练脚本中增加以下代码拉起TaskD  Manager。

            ```
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python manager.py &   # 具体执行路径由当前路径决定
            fi
                
            msrun ...
            ```

    2.  拉起TaskD  Worker。修改./mindformers/trainer/base\_trainer.py文件，在代码中增加如下加粗字段。

        ```
            def training_process(
                    self,
                    config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                    network: Optional[Union[Cell, PreTrainedModel]] = None,
                    dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                    optimizer: Optional[Optimizer] = None,
                    callbacks: Optional[Union[Callback, List[Callback]]] = None,
                    compute_metrics: Optional[Union[dict, set]] = None,
                    **kwargs):
                ……
                ……
        
                logger.info(".........Starting Training Model..........")
                if get_real_rank() % 8 == 0:
                    pprint(config)
                logger.info(".........Model Compiling, Please Wait a Moment...........")
                try:
                    rank = get_rank()
                    from taskd.api.taskd_worker_api import init_taskd_worker
                    from taskd.api.taskd_worker_api import start_taskd_worker
                    init_taskd_worker(rank,5000,"ms")
                    start_taskd_worker()
                except Exception as e:
                    print("failed to call mindcluster taskd")
                model.train(config.runner_config.epochs, dataset,
                            callbacks=callbacks,
                            dataset_sink_mode=config.runner_config.sink_mode,
                            sink_size=config.runner_config.sink_size,
                            initial_epoch=config.runner_config.initial_epoch)
        ```

2.  修改训练框架代码，打开借轨开关。

    编辑启动脚本QWEN3\_for\_MS\_code/scripts/msrun\_launcher.sh文件，在代码中增加如下加粗字段。

    ```
    export MS_ENABLE_TFT='{TTP:1,TSP:1}'           # 开启临终遗言和借轨回切
    export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  # 此环境变量用于配置是否开启HCCL算子的重执行特性。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
    ```

    >[!NOTE] 说明 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >```
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >-   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >-   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：  TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。
    >    TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >    ```
    >    pip show taskd
    >    ```

3.  修改任务YAML。
    1.  在任务YAML中修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ports:                         
           - containerPort: 9601             
             name: taskd-port 
        ```

    2.  在任务YAML中新增以下加粗字段。

        ```
        ...  
           labels:  
             ...  
             fault-scheduling: "force"
         ... 
        ...  
           annotations:  
             ...  
             recover-strategy: "retry"    # 任务可用恢复策略，取值为retry，表示开启进程级在线恢复
         ... 
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - | 
                        ... 
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - |
                        ...
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
        ...
        ```



### 配置优雅容错<a name="ZH-CN_TOPIC_0000002511346501"></a>

>[!NOTE] 说明 
>该功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。

本章节将指导用户了解配置优雅容错的关键步骤。优雅容错的特性介绍、使用约束、支持的产品型号及原理请参见[（可选）优雅容错](#可选优雅容错)。

**构建镜像<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```
# MindCluster断点续训适配脚本，MINDIO_TTP_PKG为MindIO的whl安装包的路径，请根据实际情况填写
RUN pip3 install $MINDIO_TTP_PKG 
```

**适配训练脚本<a name="section731511818483"></a>**

在启动脚本（例如train\_start.sh）中，新增以下加粗字段，示例如下。

```
...
export MS_ENABLE_TFT='{RSC:1}'      # MindSpore场景下配置此字段开启优雅容错
...
```

**配置启动YAML<a name="zh-cn_topic_0000002138594553_section18371651403"></a>**

修改Ascend Device Plugin组件的启动YAML，设置 -hotReset=1开启热复位，使用优雅容错模式。**注意：优雅容错和进程级别重调度、进程级在线恢复不可同时开启。**

```
...
      containers:
      - image: ascend-k8sdeviceplugin:v{version}
        name: device-plugin-01
        resources:
          requests:
            memory: 500Mi
            cpu: 500m
          limits:
            memory: 500Mi
            cpu: 500m
        command: [ "/bin/bash", "-c", "--"]
        args: [ "device-plugin  
                 -useAscendDocker=true 
                 -volcanoType=true                    # 重调度场景下必须使用Volcano
                 -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                 -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒
                 -hotReset=1      # 使用断点续训时，可以在Job级或Pod级重调度的基础上，开启热复位功能，使用优雅容错模式
                 -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                 -logLevel=0" ]
        securityContext:
          privileged: true
          readOnlyRootFilesystem: true
...
```


### 配置在线压测<a name="ZH-CN_TOPIC_0000002511426487"></a>

#### PyTorch场景（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002479386572"></a>

本章节将指导用户了解配置在线压测的关键步骤。在线压测的特性介绍、使用约束、支持的产品型号等请参见[在线压测](#在线压测)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../installation_guide.md#ascend-docker-runtime)、[Ascend Operator](../installation_guide.md#ascend-operator)、[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装[torch\_npu](#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）、[CANN](#制作mindspeed-llm训练镜像pytorch框架)（8.2.RC1及以上版本）、[TaskD](#制作mindspeed-llm训练镜像pytorch框架)和[MindIO](#制作mindspeed-llm训练镜像pytorch框架)（7.2.RC1及以上版本）

**操作步骤<a name="section188080175496"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在训练进程内部拉起TaskD  Worker。

    1.  拉起TaskD  Manager。
        1.  创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

            ```
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
            
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
            
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 说明 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

        2.  在训练脚本中增加以下代码，拉起TaskD  Manager。

            ```
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python manager.py &           # 具体执行路径由当前路径决定
            fi
                
            torchrun ...
            ```

    2.  拉起TaskD  Worker。

        修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，在代码中增加如下加粗字段。

        ```
        def pretrain(train_valid_test_dataset_provider,
                     model_provider,
                     model_type,
                     forward_step_func,
                     process_non_loss_data_func=None,
                     extra_args_provider=None,
                     args_defaults={}):
            print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                time.time() - _TRAIN_START_TIME))
            print_datetime('after megatron is initialized')
            import torch.distributed as dist
            if dist.is_initialized():
               rank = dist.get_rank()
               from taskd.api.taskd_worker_api import init_taskd_worker
               from taskd.api.taskd_worker_api import start_taskd_worker
               init_taskd_worker(rank,5000,"pt")
               start_taskd_worker()
            app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
            one_logger_utils.on_pretrain_start()
        ```

    >[!NOTE] 说明 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >```
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >-   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >-   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：  TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。
    >    TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >    ```
    >    pip show taskd
    >    ```

2.  修改任务YAML。
    1.  在任务YAML中修改容器暴露端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ports:                         
           - containerPort: 9601             
             name: taskd-port 
        ```

    2.  在任务YAML中新增以下加粗字段。

        ```
        ...  
           labels:  
             ...  
             fault-scheduling: "force"
         ... 
        ...  
           annotations:  
             ...  
             recover-strategy: "recover"    # 任务可用恢复策略，取值为recover，表示开启进程级别重调度
         ... 
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - | 
                        cd /job/code; 
                        chmod +x scripts/train_start.sh; 
                        bash scripts/train_start.sh
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - |
                        cd /job/code; 
                        chmod +x scripts/train_start.sh; 
                        bash scripts/train_start.sh
        ...
        ```


#### MindSpore场景（基于MindFormers）<a name="ZH-CN_TOPIC_0000002479226554"></a>

本章节将指导用户了解配置在线压测的关键步骤。在线压测的特性介绍、使用约束、支持的产品型号等请参见[在线压测](#在线压测)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../installation_guide.md#ascend-docker-runtime)、[Ascend Operator](../installation_guide.md#ascend-operator)、[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装MindSpore（2.7.0及以上版本）、[CANN](#制作mindformers训练镜像mindspore框架)（8.2.RC1及以上版本）、[TaskD](#制作mindformers训练镜像mindspore框架)和[MindIO](#制作mindformers训练镜像mindspore框架)（7.2.RC1及以上版本）

**操作步骤<a name="section9479182019317"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
    1.  拉起TaskD  Manager。
        1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

            ```
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
            
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
            
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 说明 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

        2.  在训练脚本中增加以下代码拉起TaskD  Manager。

            ```
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python manager.py &   # 具体执行路径由当前路径决定
            fi
                
            msrun ...
            ```

    2.  拉起TaskD  Worker。修改./mindformers/trainer/base\_trainer.py文件，在代码中增加如下加粗字段。

        ```
            def training_process(
                    self,
                    config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                    network: Optional[Union[Cell, PreTrainedModel]] = None,
                    dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                    optimizer: Optional[Optimizer] = None,
                    callbacks: Optional[Union[Callback, List[Callback]]] = None,
                    compute_metrics: Optional[Union[dict, set]] = None,
                    **kwargs):
                ……
                ……
        
                logger.info(".........Starting Training Model..........")
                if get_real_rank() % 8 == 0:
                    pprint(config)
                logger.info(".........Model Compiling, Please Wait a Moment...........")
                try:
                    rank = get_rank()
                    from taskd.api.taskd_worker_api import init_taskd_worker
                    from taskd.api.taskd_worker_api import start_taskd_worker
                    init_taskd_worker(rank,5000,"ms")
                    start_taskd_worker()
                except Exception as e:
                    print("failed to call mindcluster taskd")
                model.train(config.runner_config.epochs, dataset,
                            callbacks=callbacks,
                            dataset_sink_mode=config.runner_config.sink_mode,
                            sink_size=config.runner_config.sink_size,
                            initial_epoch=config.runner_config.initial_epoch)
        ```

2.  修改训练框架代码，打开在线压测开关。

    编辑启动脚本QWEN3\_for\_MS\_code/scripts/msrun\_launcher.sh文件，在代码中增加如下加粗字段。

    ```
    export MS_ENABLE_TFT='{TTP:1,TSP:1}'           # 开启临终遗言和在线压测
    ```

    >[!NOTE] 说明 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >```
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >-   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >-   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：  TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。
    >    TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >    ```
    >    pip show taskd
    >    ```

3.  修改任务YAML。
    1.  在任务YAML中修改容器暴露端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ports:                         
           - containerPort: 9601             
             name: taskd-port 
        ```

    2.  在任务YAML中新增以下加粗字段。

        ```
        ...  
           labels:  
             ...  
             fault-scheduling: "force"
         ... 
        ...  
           annotations:  
             ...  
             recover-strategy: "recover"    # 任务可用恢复策略，取值为recover，表示开启进程级别重调度
         ... 
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    command:                           # training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    command:                           # training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
        ...
        ```



### 配置亚健康热切<a name="ZH-CN_TOPIC_0000002511426471"></a>

本章节将指导用户了解配置亚健康热切的关键步骤。亚健康热切的特性介绍、使用约束、支持的产品型号及原理请参见[亚健康热切](#亚健康热切)。

**构建镜像<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

使用Dockerfile构建容器镜像，新增启动命令。示例如下。

```
# MindCluster断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，MINDIO_TTP_PKG为MindIO的whl安装包的路径，MINDSPORE_WHL为MindSpore的whl安装包的路径，请根据实际情况填写
# 可选，PyTorch框架下，使用亚健康热切时必须配置以下命令
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用亚健康热切需配置以下命令
RUN pip3 install $MINDIO_TTP_PKG
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDSPORE_WHL
```

**准备任务YAML<a name="zh-cn_topic_0000002134174097_section98717593512"></a>**

在任务YAML中，新增以下字段，配置亚健康热切，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

```
... 
metadata:  
   labels:  
     ... 
     subHealthyStrategy: "hotSwitch"
...
ports:                          
   - containerPort: 9601              
     name: taskd-port
...
```

**适配训练脚本<a name="zh-cn_topic_0000002098654822_section17330181621710"></a>**

在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。

1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

    ```
    from taskd.api import init_taskd_manager, start_taskd_manager
    import os
     
    job_id=os.getenv("MINDX_TASK_ID")
    node_nums=XX          # 用户填入任务节点总数
    proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
     
    init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
    start_taskd_manager()
    ```

    >[!NOTE] 说明 
    >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

2.  在训练脚本中增加以下代码拉起TaskD  Manager。

    ```
    export TASKD_PROCESS_ENABLE="on" 
     
    # PyTorch框架下
    if [[ "${RANK}" == 0 ]]; then
        export MASTER_ADDR=${POD_IP} 
        python manager.py &           # 具体执行路径由当前路径决定
    fi 
          
    torchrun ...
     
    # MindSpore框架下
    if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then 
        python manager.py &   # 具体执行路径由当前路径决定
    fi 
          
    msrun ...
    ```

    >[!NOTE] 说明 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >```
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >-   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >-   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。
    >    TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >    ```
    >    pip show taskd
    >    ```


### 配置弹性训练<a name="ZH-CN_TOPIC_0000002511346471"></a>

本章节将指导用户了解配置弹性训练的关键步骤。弹性训练的特性介绍、使用约束、支持的产品型号及原理请参见[弹性训练](#弹性训练)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

-   在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../installation_guide.md#ascend-docker-runtime)、[Ascend Operator](../installation_guide.md#ascend-operator)、[ClusterD](../installation_guide.md#clusterd)、[Ascend Device Plugin](../installation_guide.md#ascend-device-plugin)和[Volcano](../installation_guide.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
-   在容器内安装[torch\_npu](#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）、[CANN](#制作mindspeed-llm训练镜像pytorch框架)（8.2.RC1及以上版本）、[TaskD](#制作mindspeed-llm训练镜像pytorch框架)和[MindIO](#制作mindspeed-llm训练镜像pytorch框架)（7.2.RC1及以上版本）

**操作步骤<a name="section188080175496"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1.  创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

        ```
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
        
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX         # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
        
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 说明 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2.  在训练脚本中增加以下代码，拉起TaskD  Manager。

        ```
        sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        export TASKD_PROCESS_ENABLE="on"
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python manager.py &           # 具体执行路径由当前路径决定
        fi
            
        torchrun ...
        ```

2.  修改任务YAML。
    1.  在任务YAML中修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ports:                         
           - containerPort: 9601             
             name: taskd-port 
        ```

    2.  在任务YAML中新增以下加粗字段。

        ```
        ...  
           labels:  
             ...  
             fault-scheduling: "force"
         ... 
        ...  
           annotations:  
             ...
             wait-reschedule-timeout: "270" # 进程级恢复等待故障节点重调度的超时时间，默认为270秒，取值范围为30~270。进程级恢复和弹性训练均开启时，等待此时间后若故障节点调度成功，则进行进程级恢复，否则触发弹性训练
             recover-strategy: "elastic-training"    # 任务可用恢复策略，取值为elastic-training，表示开启弹性训练
         ... 
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    env:
                      - name: MINDIO_WAIT_MINDX_TIME         # 未开启进程级恢复，开启弹性训练场景下建议配置60以上
                        value: "60"
                    args:
                      - | 
                        ...
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    env:
                      - name: MINDIO_WAIT_MINDX_TIME         # 未开启进程级恢复，开启弹性训练场景下建议配置60以上
                        value: "60"
                    args:
                      - |
                        ...
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
        ...
        ```

3.  修改训练框架代码。

    进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3“目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```


### 参数说明<a name="ZH-CN_TOPIC_0000002511346491"></a>

不同的故障处理模式需要配置的参数各不相同，如[表1](#table1247342123814)所示，每个参数所表示的含义及填写说明详见[表2](#zh-cn_topic_0000002163392281_table1474820818115)。Ascend Operator在进程级别重调度、进程级在线恢复、进程级原地恢复和弹性训练场景下，会根据用户配置的recover-strategy和pod-rescheduling注入不同的环境变量，自动给任务打上process-recover-enable=on标签开启进程级恢复开关，无需用户手动指定。具体注入的环境变量如[表3](#table10283161512105)所示。

**表 1**  故障处理所需参数

<a name="table1247342123814"></a>
<table><tbody><tr id="row624717420389"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1724711429388"><a name="p1724711429388"></a><a name="p1724711429388"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p882622715436"><a name="p882622715436"></a><a name="p882622715436"></a>Job级别重调度</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p117701032194319"><a name="p117701032194319"></a><a name="p117701032194319"></a>Pod级别重调度</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p182471842183813"><a name="p182471842183813"></a><a name="p182471842183813"></a>进程级别重调度（recover策略）</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p35155320438"><a name="p35155320438"></a><a name="p35155320438"></a>进程级别原地恢复</p>
<p id="p1259532434"><a name="p1259532434"></a><a name="p1259532434"></a>（recover-in-place策略）</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p37521612448"><a name="p37521612448"></a><a name="p37521612448"></a>进程级在线恢复</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p14247442143812"><a name="p14247442143812"></a><a name="p14247442143812"></a>优雅容错</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p84114121442"><a name="p84114121442"></a><a name="p84114121442"></a>弹性训练</p>
</td>
</tr>
<tr id="row7247154215383"><td class="cellrowborder" valign="top" width="15.24%"><p id="p22316366391"><a name="p22316366391"></a><a name="p22316366391"></a>hotReset</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p32233618390"><a name="p32233618390"></a><a name="p32233618390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1422133683919"><a name="p1422133683919"></a><a name="p1422133683919"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p12221636143918"><a name="p12221636143918"></a><a name="p12221636143918"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p52116364390"><a name="p52116364390"></a><a name="p52116364390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p221173643911"><a name="p221173643911"></a><a name="p221173643911"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p0211836143914"><a name="p0211836143914"></a><a name="p0211836143914"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1321203633918"><a name="p1321203633918"></a><a name="p1321203633918"></a>-</p>
</td>
</tr>
<tr id="row1024894243810"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1218113612390"><a name="p1218113612390"></a><a name="p1218113612390"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p518736123916"><a name="p518736123916"></a><a name="p518736123916"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1917936133912"><a name="p1917936133912"></a><a name="p1917936133912"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p21753673912"><a name="p21753673912"></a><a name="p21753673912"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1171736143916"><a name="p1171736143916"></a><a name="p1171736143916"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p9171236113916"><a name="p9171236113916"></a><a name="p9171236113916"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p12171236163917"><a name="p12171236163917"></a><a name="p12171236163917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p10161836143914"><a name="p10161836143914"></a><a name="p10161836143914"></a>√</p>
</td>
</tr>
<tr id="row1824884293812"><td class="cellrowborder" valign="top" width="15.24%"><p id="p91533663919"><a name="p91533663919"></a><a name="p91533663919"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p5147367393"><a name="p5147367393"></a><a name="p5147367393"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p151433693911"><a name="p151433693911"></a><a name="p151433693911"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1714183618396"><a name="p1714183618396"></a><a name="p1714183618396"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p111413620399"><a name="p111413620399"></a><a name="p111413620399"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p713136113913"><a name="p713136113913"></a><a name="p713136113913"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p6134364397"><a name="p6134364397"></a><a name="p6134364397"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1813203617390"><a name="p1813203617390"></a><a name="p1813203617390"></a>-</p>
</td>
</tr>
<tr id="row2248144273815"><td class="cellrowborder" valign="top" width="15.24%"><p id="p15112368396"><a name="p15112368396"></a><a name="p15112368396"></a>process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p511163613915"><a name="p511163613915"></a><a name="p511163613915"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p2011203616395"><a name="p2011203616395"></a><a name="p2011203616395"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1610133603920"><a name="p1610133603920"></a><a name="p1610133603920"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p310103643919"><a name="p310103643919"></a><a name="p310103643919"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p91015365399"><a name="p91015365399"></a><a name="p91015365399"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p6105364395"><a name="p6105364395"></a><a name="p6105364395"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p191836133916"><a name="p191836133916"></a><a name="p191836133916"></a>√</p>
</td>
</tr>
<tr id="row2248154243818"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1814364394"><a name="p1814364394"></a><a name="p1814364394"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p12773693917"><a name="p12773693917"></a><a name="p12773693917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p147836193919"><a name="p147836193919"></a><a name="p147836193919"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p13712364397"><a name="p13712364397"></a><a name="p13712364397"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p127193633919"><a name="p127193633919"></a><a name="p127193633919"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p86173683915"><a name="p86173683915"></a><a name="p86173683915"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p9673673919"><a name="p9673673919"></a><a name="p9673673919"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p66123611391"><a name="p66123611391"></a><a name="p66123611391"></a>√</p>
</td>
</tr>
<tr id="row424864214389"><td class="cellrowborder" valign="top" width="15.24%"><p id="p134936133911"><a name="p134936133911"></a><a name="p134936133911"></a>PROCESS_RECOVER</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p184173618395"><a name="p184173618395"></a><a name="p184173618395"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p164123616391"><a name="p164123616391"></a><a name="p164123616391"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p63193643915"><a name="p63193643915"></a><a name="p63193643915"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p83113663920"><a name="p83113663920"></a><a name="p83113663920"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p4323643914"><a name="p4323643914"></a><a name="p4323643914"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p736362397"><a name="p736362397"></a><a name="p736362397"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p42163683917"><a name="p42163683917"></a><a name="p42163683917"></a>√</p>
</td>
</tr>
<tr id="row1924904210386"><td class="cellrowborder" valign="top" width="15.24%"><p id="p212036183913"><a name="p212036183913"></a><a name="p212036183913"></a>ENABLE_RESTART_FAULT_PROCESS</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p170133620393"><a name="p170133620393"></a><a name="p170133620393"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p140173613390"><a name="p140173613390"></a><a name="p140173613390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p799983573910"><a name="p799983573910"></a><a name="p799983573910"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1899973516399"><a name="p1899973516399"></a><a name="p1899973516399"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p1199873573913"><a name="p1199873573913"></a><a name="p1199873573913"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p49985358399"><a name="p49985358399"></a><a name="p49985358399"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1799893511398"><a name="p1799893511398"></a><a name="p1799893511398"></a>-</p>
</td>
</tr>
<tr id="row19391344114010"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1794014446408"><a name="p1794014446408"></a><a name="p1794014446408"></a>ELASTIC_PROCESS_RECOVER_ENABLE</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p1494064419401"><a name="p1494064419401"></a><a name="p1494064419401"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p19940944134019"><a name="p19940944134019"></a><a name="p19940944134019"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p2094054414013"><a name="p2094054414013"></a><a name="p2094054414013"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p494084464017"><a name="p494084464017"></a><a name="p494084464017"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p69402446400"><a name="p69402446400"></a><a name="p69402446400"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p894014464011"><a name="p894014464011"></a><a name="p894014464011"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p199403441401"><a name="p199403441401"></a><a name="p199403441401"></a>-</p>
</td>
</tr>
<tr id="row448045664010"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1848011565403"><a name="p1848011565403"></a><a name="p1848011565403"></a>--enable-high-availability（MindSpeed-LLM侧参数）</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p948075615409"><a name="p948075615409"></a><a name="p948075615409"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1748045664018"><a name="p1748045664018"></a><a name="p1748045664018"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p548025618403"><a name="p548025618403"></a><a name="p548025618403"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p448016564402"><a name="p448016564402"></a><a name="p448016564402"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p9480145674013"><a name="p9480145674013"></a><a name="p9480145674013"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p1848055694012"><a name="p1848055694012"></a><a name="p1848055694012"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p124809565404"><a name="p124809565404"></a><a name="p124809565404"></a>√</p>
</td>
</tr>
<tr id="row112463954119"><td class="cellrowborder" valign="top" width="15.24%"><p id="p76389163416"><a name="p76389163416"></a><a name="p76389163416"></a>--enable-hbmfault-repair（MindSpeed-LLM侧参数）</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p1124699204118"><a name="p1124699204118"></a><a name="p1124699204118"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p192469924110"><a name="p192469924110"></a><a name="p192469924110"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1824619916411"><a name="p1824619916411"></a><a name="p1824619916411"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p724719904114"><a name="p724719904114"></a><a name="p724719904114"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p1424720924114"><a name="p1424720924114"></a><a name="p1424720924114"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p152478914111"><a name="p152478914111"></a><a name="p152478914111"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p32471934114"><a name="p32471934114"></a><a name="p32471934114"></a>-</p>
</td>
</tr>
<tr id="row3150821154117"><td class="cellrowborder" valign="top" width="15.24%"><p id="p7151102124114"><a name="p7151102124114"></a><a name="p7151102124114"></a>--enable-worker-reboot（MindSpeed-LLM侧参数）</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p815192144117"><a name="p815192144117"></a><a name="p815192144117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p315110217417"><a name="p315110217417"></a><a name="p315110217417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p31511721114110"><a name="p31511721114110"></a><a name="p31511721114110"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p115118211413"><a name="p115118211413"></a><a name="p115118211413"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p10151142164119"><a name="p10151142164119"></a><a name="p10151142164119"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p161517211417"><a name="p161517211417"></a><a name="p161517211417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p10151221164119"><a name="p10151221164119"></a><a name="p10151221164119"></a>-</p>
</td>
</tr>
<tr id="row4799183364111"><td class="cellrowborder" valign="top" width="15.24%"><p id="p7799233154115"><a name="p7799233154115"></a><a name="p7799233154115"></a>--enable-elastic-training（MindSpeed-LLM侧参数）</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p2799133324116"><a name="p2799133324116"></a><a name="p2799133324116"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1079915338414"><a name="p1079915338414"></a><a name="p1079915338414"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p127999338410"><a name="p127999338410"></a><a name="p127999338410"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1280083317415"><a name="p1280083317415"></a><a name="p1280083317415"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p11800143320417"><a name="p11800143320417"></a><a name="p11800143320417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p9800153364112"><a name="p9800153364112"></a><a name="p9800153364112"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1780063324111"><a name="p1780063324111"></a><a name="p1780063324111"></a>√</p>
</td>
</tr>
<tr id="row1551285114419"><td class="cellrowborder" valign="top" width="15.24%"><p id="p12512175154117"><a name="p12512175154117"></a><a name="p12512175154117"></a>max_restarts</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p7512125115416"><a name="p7512125115416"></a><a name="p7512125115416"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p13512145114412"><a name="p13512145114412"></a><a name="p13512145114412"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p151211510417"><a name="p151211510417"></a><a name="p151211510417"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1751215515414"><a name="p1751215515414"></a><a name="p1751215515414"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p10512125104113"><a name="p10512125104113"></a><a name="p10512125104113"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p751285111417"><a name="p751285111417"></a><a name="p751285111417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p051245113418"><a name="p051245113418"></a><a name="p051245113418"></a>-</p>
</td>
</tr>
<tr id="row1810414334211"><td class="cellrowborder" valign="top" width="15.24%"><p id="p171048313421"><a name="p171048313421"></a><a name="p171048313421"></a>monitor_interval</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p1910414312422"><a name="p1910414312422"></a><a name="p1910414312422"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1710483114210"><a name="p1710483114210"></a><a name="p1710483114210"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1810415310420"><a name="p1810415310420"></a><a name="p1810415310420"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p201044364218"><a name="p201044364218"></a><a name="p201044364218"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p9104335421"><a name="p9104335421"></a><a name="p9104335421"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p1104163194212"><a name="p1104163194212"></a><a name="p1104163194212"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p4104183114212"><a name="p4104183114212"></a><a name="p4104183114212"></a>-</p>
</td>
</tr>
<tr id="row1260817211339"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1960910211831"><a name="p1960910211831"></a><a name="p1960910211831"></a><span id="ph48451032338"><a name="ph48451032338"></a><a name="ph48451032338"></a>fault-retry-times</span></p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p96093216315"><a name="p96093216315"></a><a name="p96093216315"></a><span id="ph116361991246"><a name="ph116361991246"></a><a name="ph116361991246"></a>√</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1660911211319"><a name="p1660911211319"></a><a name="p1660911211319"></a><span id="ph151830137416"><a name="ph151830137416"></a><a name="ph151830137416"></a>√</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p18609192118312"><a name="p18609192118312"></a><a name="p18609192118312"></a><span id="ph5658101813418"><a name="ph5658101813418"></a><a name="ph5658101813418"></a>√</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p160912214316"><a name="p160912214316"></a><a name="p160912214316"></a><span id="ph115481430045"><a name="ph115481430045"></a><a name="ph115481430045"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p960913211138"><a name="p960913211138"></a><a name="p960913211138"></a><span id="ph8478143111413"><a name="ph8478143111413"></a><a name="ph8478143111413"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p12609102110312"><a name="p12609102110312"></a><a name="p12609102110312"></a><span id="ph1129711326411"><a name="ph1129711326411"></a><a name="ph1129711326411"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1609421731"><a name="p1609421731"></a><a name="p1609421731"></a><span id="ph1301142311415"><a name="ph1301142311415"></a><a name="ph1301142311415"></a>√</span></p>
</td>
</tr>
</tbody>
</table>

**表 2**  参数填写说明

<a name="zh-cn_topic_0000002163392281_table1474820818115"></a>
|参数名称|参数位置|参数说明|
|--|--|--|
|hotReset|Ascend Device Plugin组件的启动YAML|优雅容错功能开关。<ul><li>取值为1：使用断点续训时，可以在Job级或Pod级重调度的基础上，开启热复位功能，使用优雅容错模式；</li><li>取值为2：使用进程级恢复时，请将hotReset参数值设置为2，开启离线恢复模式。</li></ul><span> 说明： </span><p>取值为1对应的功能已经日落，请配置其他取值。</p>|
|pod-rescheduling|训练任务YAML的metadata.labels|<ul><li>on：开启Pod级别重调度。</li><li>其他值或不使用该字段：关闭Pod级别重调度。</li></ul>|
|fault-scheduling|训练任务YAML的metadata.labels|重调度开关。|
|process-recover-enable|训练任务YAML的metadata.labels|<ul><li>on：开启进程级别重调度及进程级在线恢复。进程级别重调度和优雅容错不能同时开启，若同时开启，断点续训将通过job级重调度恢复训练。</li><li>pause：暂时关闭进程级别重调度及进程级在线恢复。</li><li>off或不使用该字段：关闭进程级别重调度及进程级在线恢复。</li></ul>|
|recover-strategy|训练任务YAML的metadata.annotations|任务可用恢复策略。<ul><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>elastic-training：弹性训练。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>|
|PROCESS_RECOVER|训练任务YAML的spec.replicaSpecs.{ Master \|Scheduler\| Worker}.template.spec.containers.env|进程级别重调度及进程级在线恢复Elastic Agent/TaskD侧总开关。<ul><li>on：开启。</li><li>off：关闭。</li></ul>|
|ELASTIC_PROCESS_RECOVER_ENABLE|启动训练YAML的spec.replicaSpecs.{ Master\|Scheduler\| Worker}. template.spec.containers.args|Elastic Agent侧进程级别重调度、进程级在线恢复、临终CKPT恢复功能开关。<ul><li>取值为1：开启本功能。</li><li>其他值：关闭本功能。<p>关闭本功能时，MindIO侧相关功能需同时关闭。</p></li></ul><span> 说明： </span><p>Elastic Agent组件已经日落，相关资料将于2026年的8.3.0版本删除。该环境变量会随之删除。</p>|
|ENABLE_RESTART_FAULT_PROCESS|启动训练YAML的spec.replicaSpecs.{ Master\|Scheduler\| Worker}. template.spec.containers.args|Elastic Agent/TaskD组件开启故障进程原地恢复功能的开关。<ul><li>on：开启本功能；</li><li>其他值：关闭本功能</li></ul>|
|--enable-high-availability|训练脚本pretrain_gpt.py的启动参数|故障快速恢复特性开关，默认关闭，配置后即开启临终遗言功能。|
|--enable-hbmfault-repair|训练脚本pretrain_gpt.py的启动参数|进程级在线恢复功能开关，默认关闭，配置后对片上内存进行故障检测，并完成在线修复。需同时开启enable-high-availability。|
|--enable-worker-reboot|训练脚本pretrain_gpt.py的启动参数|进程级别重调度功能开关，默认关闭。配置后在发生一般性故障时，进行进程级别调度。需同时开启enable-high-availability。|
|--enable-elastic-training|训练脚本pretrain_gpt.py的启动参数|弹性训练功能开关，默认关闭。|
|max_restarts|启动训练的shell脚本（例如train_start.sh）中|配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。|
|monitor_interval|启动训练的shell脚本（例如train_start.sh）中|配置监测训练进程状态的时间间隔，单位为秒，取值为整数。不配置该参数时默认为5秒。|
|HIGH_AVAILABILITY|Ascend Operator注入容器的环境变量中|Ascend Operator根据任务类型自动注入该环境变量，使用2.3.0版本MindSpeed-LLM会自动读取该环境变量，无需在train_start.sh中手动添加--enable-high-availability、--enable-hbmfault-repair、--enable-worker-reboot和--enable-elastic-training参数开启对应功能。|

**表 3** Ascend Operator注入的环境变量

<a name="table10283161512105"></a>
<table><tbody><tr id="row928321541018"><td class="cellrowborder" valign="top" width="7.580000000000002%"><p id="p7676133111213"><a name="p7676133111213"></a><a name="p7676133111213"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="15.500000000000004%"><p id="p98460430123"><a name="p98460430123"></a><a name="p98460430123"></a>recover</p>
</td>
<td class="cellrowborder" valign="top" width="14.830000000000002%"><p id="p58461043151210"><a name="p58461043151210"></a><a name="p58461043151210"></a>retry</p>
</td>
<td class="cellrowborder" valign="top" width="19.39%"><p id="p20846164319122"><a name="p20846164319122"></a><a name="p20846164319122"></a>recover-in-place</p>
</td>
<td class="cellrowborder" valign="top" width="9.920000000000002%"><p id="p14247569165"><a name="p14247569165"></a><a name="p14247569165"></a>elastic-training</p>
</td>
<td class="cellrowborder" valign="top" width="16.200000000000003%"><p id="p158461443201210"><a name="p158461443201210"></a><a name="p158461443201210"></a>dump</p>
</td>
<td class="cellrowborder" valign="top" width="7.000000000000002%"><p id="p148465435121"><a name="p148465435121"></a><a name="p148465435121"></a>exit</p>
</td>
<td class="cellrowborder" valign="top" width="9.580000000000002%"><p id="p584614351218"><a name="p584614351218"></a><a name="p584614351218"></a>pod-rescheduling</p>
</td>
</tr>
<tr id="row10283115171018"><td class="cellrowborder" valign="top" width="7.580000000000002%"><p id="p1621320131319"><a name="p1621320131319"></a><a name="p1621320131319"></a><span id="ph1551815244211"><a name="ph1551815244211"></a><a name="ph1551815244211"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.500000000000004%"><a name="ul22620111313"></a><a name="ul22620111313"></a><ul id="ul22620111313"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>HIGH_AVAILABILITY=recover</li></ul>
</td>
<td class="cellrowborder" valign="top" width="14.830000000000002%"><a name="ul62102015138"></a><a name="ul62102015138"></a><ul id="ul62102015138"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>HIGH_AVAILABILITY=retry<p id="p102420141318"><a name="p102420141318"></a><a name="p102420141318"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="19.39%"><a name="ul721720131318"></a><a name="ul721720131318"></a><ul id="ul721720131318"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>ENABLE_RESTART_FAULT_PROCESS=on</li><li>HIGH_AVAILABILITY=recover</li></ul>
</td>
<td class="cellrowborder" valign="top" width="9.920000000000002%"><a name="ul4462124162110"></a><a name="ul4462124162110"></a><ul id="ul4462124162110"><li>PROCESS_RECOVER=on</li><li>HIGH_AVAILABILITY=elastic-training</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.200000000000003%"><a name="ul182142017135"></a><a name="ul182142017135"></a><ul id="ul182142017135"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>HIGH_AVAILABILITY=dump<p id="p5216201131"><a name="p5216201131"></a><a name="p5216201131"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="7.000000000000002%"><p id="p102102012138"><a name="p102102012138"></a><a name="p102102012138"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.580000000000002%"><p id="p163162071320"><a name="p163162071320"></a><a name="p163162071320"></a>-</p>
</td>
</tr>
<tr id="row1628391516107"><td class="cellrowborder" valign="top" width="7.580000000000002%"><p id="p139556322136"><a name="p139556322136"></a><a name="p139556322136"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="15.500000000000004%"><a name="ul295563211139"></a><a name="ul295563211139"></a><ul id="ul295563211139"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT='{ ARF:1}'<p id="p16955153219134"><a name="p16955153219134"></a><a name="p16955153219134"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="14.830000000000002%"><a name="ul109551632111320"></a><a name="ul109551632111320"></a><ul id="ul109551632111320"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT='{ UCE:1, HCCE:1}'<p id="p1595593241313"><a name="p1595593241313"></a><a name="p1595593241313"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="19.39%"><a name="ul195583215139"></a><a name="ul195583215139"></a><ul id="ul195583215139"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>ENABLE_RESTART_FAULT_PROCESS=on</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT='{ ARF:1}'<p id="p1095514325133"><a name="p1095514325133"></a><a name="p1095514325133"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="9.920000000000002%"><p id="p324776161612"><a name="p324776161612"></a><a name="p324776161612"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="16.200000000000003%"><a name="ul1595513218133"></a><a name="ul1595513218133"></a><ul id="ul1595513218133"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT='{ TTP:1}'<p id="p495553210131"><a name="p495553210131"></a><a name="p495553210131"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="7.000000000000002%"><p id="p7955193220131"><a name="p7955193220131"></a><a name="p7955193220131"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.580000000000002%"><p id="p275141842218"><a name="p275141842218"></a><a name="p275141842218"></a>MS_ENABLE_TFT='{ RSC:1}'</p>
</td>
</tr>
</tbody>
</table>



## 配置训练恢复<a name="ZH-CN_TOPIC_0000002479386506"></a>

### 配置周期性CKPT保存<a name="ZH-CN_TOPIC_0000002479226552"></a>

本章节将指导用户了解周期性CKPT保存的关键步骤。周期性CKPT保存的特性介绍请参见[周期性CKPT保存](#周期性ckpt保存)。

**配置存储CKPT加载<a name="zh-cn_topic_0000002111866386_section1296017551704"></a>**

从存储加载CKPT可基于AI框架提供的加载接口进行加载，用户需要传入需要加载的文件路径到AI框架中。以MindSpeed-LLM框架为例，用户如果需要配置从存储加载CKPT功能，可参考以下示例。

在任务YAML中，新增以下加粗字段，开启存储CKPT加载。其中“--load“是训练进程恢复的统一开关，打开后训练进程恢复才生效。

```
...
spec:
  replicaSpecs:
    Master:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            args:
              - | 
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                  --load /data/ckpt/XXX \  # 填写CKPT所在存储的路径   
                  ...
    Worker:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            ...
            args:
              - |
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                --load /data/ckpt/XXX \    # 填写CKPT所在存储的路径   
                  ...
...
```


### 配置临终CKPT保存<a name="ZH-CN_TOPIC_0000002479226544"></a>

本章节将指导用户了解临终CKPT保存的关键步骤。临终CKPT保存的特性介绍请参见[临终CKPT保存](#临终ckpt保存)。

**构建镜像<a name="zh-cn_topic_0000002112026142_section26738428458"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```
... 
# MindCluster无损失断点续训适配脚本
RUN pip3 install $TASKD_WHL 
RUN pip3 install $MINDIO_TTP_PKG 

# 可选，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
```
**准备任务YAML<a name="zh-cn_topic_0000002112026142_section2671124124612"></a>**

在训练任务YAML中，新增以下字段，开启进程级别恢复。recover-strategy是训练进程恢复使用的策略，其中的dump代表开启临终CKPT。在ports中增加ttp-port为8000，增加TaskD通信使用的端口9601。

临终CKPT保存可以作为进程级别恢复流程中的一个策略，名为“dump”策略，设置到recover-strategy中。示例如下。

```
... 
metadata:  
   labels:  
     ...  
 ... 
...  
   annotations:  
     ...  
     recover-strategy: "dump"       # 任务可用恢复策略为保存临终遗言
 ... 
  
... 
spec:  
   replicaSpecs:  
      Master: 
         template: 
            spec: 
              containers: 
                 env: 
                   - name: TTP_PORT 
                     value: "8000" 
                 args: […] 
                 ports: 
                   - containerPort: 8000 
                     name: ttp-port 
                   - containerPort: 9601  
                     name: taskd-port
     ...  
     Worker: 
        template: 
          spec: 
            containers: 
               env: 
                 - name: TTP_PORT 
                   value: "8000" 
               args: […] 
               ports: 
                 - containerPort: 8000 
                   name: ttp-port 
                 - containerPort: 9601  
                   name: taskd-port
 ...
```

**适配训练脚本<a name="zh-cn_topic_0000002112026142_section058501610462"></a>**

1.  在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 说明 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2.  在训练脚本中增加以下代码拉起TaskD  Manager。

        ```
        export TASKD_PROCESS_ENABLE="on" 
         
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP} 
            python manager.py &           # 具体执行路径由当前路径决定
        fi 
              
        torchrun ...
        ```

2.  在启动脚本（例如train\_start.sh）中，新增--max\_restarts参数，示例如下。

    ```
    ... 
       logger "server id is: ""${server_id}" 
       if [ "${framework}" == "PyTorch" ]; then 
         get_env_for_pytorch_multi_node_job 
         DISTRIBUTED_ARGS="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT  --max_restarts 32767" 
     ...
    ```

         其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

>[!NOTE] 说明 
>如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
>```
>export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
>```
>-   libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
>-   libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。
>    TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
>    ```
>    pip show taskd
>    ```


### 配置参数面传参恢复<a name="ZH-CN_TOPIC_0000002479386502"></a>

当前仅支持在进程级别重调度和进程级在线恢复特性中使用该能力，按照[配置进程级别重调度](#配置进程级别重调度)和[配置进程级在线恢复](#配置进程级在线恢复)特性适配后默认开启该能力。

**（可选）关闭参数面传参恢复<a name="zh-cn_topic_0000002181310402_section199132050405"></a>**

在进程级别重调度和进程级在线恢复特性中，如果用户想要关闭该功能，修改为从存储CKPT加载参数恢复，需修改任务YAML。以使用进程级重调度且关闭参数面传参恢复为例，示例如下。

```
...  
metadata: 
   labels:  
     ...  
     fault-scheduling: "grace"
 ... 
...  
   annotations:  
     ...  
     recover-strategy: "recover"   # 任务可用恢复策略为进程级别重调度 
... 
...
spec:
  replicaSpecs:
    Master:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            ...
            args:
              - | 
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                  --distributed-optimizer-no-replica \
                  ...
    Worker:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            ...
            args:
              - |
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
	          --distributed-optimizer-no-replica \
                  ...
...
```

**distributed-optimizer-no-replica**：数据修复支持周期CKPT功能开关，默认关闭，配置后副本优化器无副本，减小内存占用，在进程级别重调度和进程级在线恢复场景下，使用周期CKPT进行修复。本开关需开启进程级别重调度或进程级在线恢复。


### 配置集成时间优化方案<a name="ZH-CN_TOPIC_0000002479386526"></a>

#### 恢复时间优化（PyTorch）<a name="ZH-CN_TOPIC_0000002479386516"></a>

本章节介绍在PyTorch框架上使用断点续训特性时，用户可以选择使用的缩短断点续训时间的相关功能，包括[故障检测时间优化](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section195202179141)、[集合通信初始化时间优化](#zh-cn_topic_0000002163883997_section725312412292)、[训练回滚及加载CheckPoint时间优化](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section71731855173720)和[算子编译时间优化](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section14599171031417)。

**故障检测时间优化<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section195202179141"></a>**

由于集群中出现的参数面网络故障不一定会影响训练任务，因此集群调度组件不会强制中断任务；当参数面网络故障影响训练任务时，会触发集合通信的网络超时等待机制，在等待时间（通常默认为30分钟）后，集群调度组件才能感知到该故障，从而触发断点续训。针对该问题，PyTorch  Adapter插件（torch\_npu）提供**watchdog故障检测**功能，可用于检测训练任务是否受到影响，缩短故障检测时间，该功能的详细说明请参考[表1](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table4822175901415)。

**表 1** watchdog故障检测功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table4822175901415"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row9823145931412"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"></a>watchdog</span>故障检测。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row58231859181412"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"></a>训练启动时，同时启动一个监测线程不断获取通信异常以及task执行异常。监测到故障发生后，快速抛出异常并终止训练任务进程，触发重调度流程。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row138235598144"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"></a>仅支持<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"></a>PyTorch</span> 1.11.0、2.1.0及以上版本；<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC1。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row11823195941410"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"></a>PyTorch</span> 2.1.0及以上版本默认开启<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"></a>watchdog</span>故障检测，<strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"></a>无需手动配置环境变量</strong>。</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"></a>（可选）如需关闭<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"></a>watchdog</span>故障检测，需在训练的shell启动脚本（例如train_start.sh）中，修改以下环境变量。</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"></a>...
# env for breakpoint ckpt
export RESUME_MODE_ENABLE=1

export HCCL_ASYNC_ERROR_HANDLING=0  <strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"></a>          </strong># 该环境变量的详细说明请参见<a href="../appendix.md#环境变量说明">TaskD环境变量说明</a></pre>
</td>
</tr>
</tbody>
</table>

**集合通信初始化时间优化<a name="zh-cn_topic_0000002163883997_section725312412292"></a>**

Parallel Store多线程建链优化：PyTorch框架创建通信组时，使用TCP Store进行信息交换。随着任务规模变大会影响原生TCP Store的信息处理性能，导致创建通信组时间过长。针对该问题，PyTorch Adapter插件支持使用原生TCP Store的优化版本Parallel Store，详细说明请参考[表2](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table14133757143220)。

**表 2**  Parallel Store功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table14133757143220"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row15133115723218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"></a>Parallel Store</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row2013316574328"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"></a>多线程处理建链请求，减少建链请求队列等待时间，降低总体建链时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1913318574324"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"></a>PyTorch</span> 1.11.0版本：<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC1。</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"></a>PyTorch</span> 2.1.0及以上版本：<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC3。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16133957183217"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"></a>将启动训练的shell脚本（例如train_start.sh）中，torchrun启动命令修改为torch_npu_run。</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"></a>比如将</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"></a><strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"></a>torchrun train.py --train_parameter=xxx ....</strong></pre>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"></a>修改为</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"></a><strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"></a>torch_npu_run train.py --train_parameter=xxx ....</strong></pre>
</td>
</tr>
</tbody>
</table>

-   原生HCCL建链性能优化：PyTorch框架在NPU侧交换集合通信信息后进行NPU间连接建链。随任务规模变大，导致建链时间大幅度增加。针对该问题，CANN对原生HCCL建链进行了性能优化，详细说明请参考[表3](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table10637950133911)。

    **表 3**  原生HCCL建链性能优化功能说明

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table10637950133911"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row763710506398"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"></a>原生HCCL建链性能优化。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row963765019395"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"></a>多线程异步完成集合通信信息协商，减少通信信息协商时间，降低总体建链时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row263845043913"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"></a>仅支持CANN 8.0.RC2及以上版本。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1563845013912"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"></a>无。</p>
    </td>
    </tr>
    </tbody>
    </table>

-   RankTable模式建链优化：集群调度Ascend Operator组件为PyTorch框架提供生成集合通信配置文件（RankTable File，也叫hccl.json文件）功能，可以通过RankTable模式建链，缩短集群通信建链时间，详细说明请参考[表4](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1749892464019)。

    **表 4**  集合通信使用RankTable模式建链

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1749892464019"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row84981324184016"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"></a>RankTable模式建链。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16498162484017"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"></a>使用<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"></a>Ascend Operator</span>为PyTorch任务生成集合通信配置文件，缩短集群通信建链时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row2499424194015"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC3。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row4499524124018"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"></a><ol id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"><li>启动YAML中已经默认挂载了hccl.json文件的父目录，用户可以根据实际情况进行修改。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"></a>volumes:
           - name: ranktable-dir
             hostPath:
               path: /user/mindx-dl/ranktable  # 该宿主机目录需要在共享目录下
               type: DirectoryOrCreate</pre>
    <div class="p" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"></a>执行以下命令，在宿主机目录下创建hccl.json文件的具体挂载路径，并修改所属用户。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"></a>mkdir -m 777 /user/mindx-dl/ranktable/任务运行的命名空间.任务名称
    chown 9000:9000 /user/mindx-dl/ranktable/default.pytorch-test</pre>
    </div>
    <div class="p" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"></a>例如：<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"></a>mkdir -m 777 /user/mindx-dl/ranktable/default.pytorch-test
    chown 9000:9000 /user/mindx-dl/ranktable/default.pytorch-test</pre>
    </div>
    </li><li>修改训练脚本，添加如下环境变量。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"></a>export RANK_TABLE_FILE=/user/mindx-dl/ranktable/hccl.json</pre>
    </li><li>修改训练YAML，添加如下设置。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"></a>yaml
          volumeMounts:
          - name: ranktable
            mountPath: /user/mindx-dl/ranktable
                     
           volumes:
           - name: ranktable
             hostPath:
               path: /user/mindx-dl/ranktable/任务运行的命名空间.任务名称  # 宿主机目录下hccl.json文件的实际路径</pre>
    </li></ol>
    </td>
    </tr>
    </tbody>
    </table>

**训练回滚及加载CheckPoint时间优化<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section71731855173720"></a>**

-   异步保存CheckPoint：训练任务会定期保存CheckPoint文件，用于保存参数信息，故障恢复需要从上一次保存的CheckPoint回滚恢复训练。由于每次保存CheckPoint文件均会浪费一定的训练时间，为了保证训练效率，保存CheckPoint的时间间隔通常较大，而保存间隔越大，每次故障时训练回滚浪费的时间就会越长。针对该问题，集群调度组件支持通过MindIO ACP异步保存CheckPoint，详细说明请参考[表5](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table5173115519372)。

    **表 5**  异步保存CheckPoint功能说明

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table5173115519372"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row717435514373"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"></a>MindIO ACP</span>异步保存CheckPoint。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row10174115583714"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"></a>从NPU中获取CheckPoint后，异步写入存储中，降低每次保存CheckPoint的训练损失和保存周期，从而降低训练回滚时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row6174655153715"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row171741155133719"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"></a>安装和使用<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"></a>MindIO</span>组件，请参考<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"></a><a href="../references.md#checkpoint保存与加载优化">CheckPoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

-   高效恢复CheckPoint：回滚恢复训练时，通常需要从存储中加载保存的CheckPoint，由于CheckPoint数据量较大，直接从存储读取加载CheckPoint的耗时较长。针对该问题，集群调度组件支持通过MindIO ACP进行CheckPoint高效恢复，详细说明请参考[表6](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1163115196618)。

    **表 6**  CheckPoint高效恢复功能说明

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1163115196618"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row763114191366"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"></a>MindIO</span> CheckPoint高效恢复。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row14631191914615"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"></a>MindIO</span>将最新的CheckPoint存储到内存中，故障恢复时可直接从内存中读取CheckPoint，降低CheckPoint读取时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row26321219766"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row9632219868"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"></a>安装和使用<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"></a>MindIO</span>组件，请参考<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"></a><a href="../references.md#checkpoint保存与加载优化">CheckPoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

**算子编译时间优化<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section14599171031417"></a>**

断点续训过程中拉起训练需要重新执行算子时，算子编译需要消耗大量时间。针对该问题，可选择算子二进制或算子编译缓存降低编译时间，详细说明请参考[表7](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table8599191019143)和[表8](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table2193759172110)。

>[!NOTE] 说明 
>算子二进制和算子编译缓存二者不兼容，请选择其中之一进行使用。

**表 7**  算子二进制功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table8599191019143"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1599111016145"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"></a>使用算子二进制。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1059931012143"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"></a>算子编译时提前加载预置的算子二进制，直接免编译执行算子。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16599161015147"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"></a>仅支持CANN 8.0.RC2及以上版本。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row7600610181419"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"></a>在<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"></a>Python</span>启动脚本中，添加算子二进制配置命令，开启算子二进制。</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"></a>torch.npu.set_compile_mode(jit_compile=False)</pre>
</td>
</tr>
</tbody>
</table>

**表 8**  算子编译缓存功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table2193759172110"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1819335920218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"></a>算子编译缓存。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row11193185913215"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"></a>算子编译时加载存储上保存的算子编译缓存文件，加载后可降低编译时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1719310593218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"></a>仅支持CANN 8.0.RC2及以上版本。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1193195962112"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"></a><ol id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"><li>在<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"></a>Python</span>启动脚本中，添加算子编译缓存配置命令，开启算子编译缓存。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"></a>torch.npu.set_compile_mode(jit_compile=True)</pre>
</li><li>在训练的shell启动脚本中（例如train_start.sh），添加如下环境变量。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"></a>export ASCEND_CACHE_PATH<strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"></a>=xxx</strong>   # 添加共享存储路径
export ASCEND_MAX_OP_CACHE_SIZE=-1    # 使用共享存储时建议开启，可解决多节点读取共享存储缓存资源争抢严重问题</pre>
</li></ol>
</td>
</tr>
</tbody>
</table>


#### 恢复时间优化（MindSpore）<a name="ZH-CN_TOPIC_0000002511346499"></a>

断点续训特性在使用MindSpore框架场景时，可以使用以下功能，缩短断点续训整体恢复时间，包括[故障检测时间优化](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section517194154019)、[训练回滚及加载CheckPoint时间优化](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section2743164217401)和[编译缓存时间优化](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section1139444324019)。

**故障检测时间优化<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section517194154019"></a>**

由于集群中出现的参数面网络故障不一定会影响训练任务，因此集群调度组件不会强制中断任务；当参数面网络故障影响训练任务时，会触发集合通信的网络超时等待机制，在等待时间（通常默认为30分钟）后，集群调度组件才能感知到该故障，从而触发断点续训。针对该问题，MindSpore提供**watchdog故障检测**功能，可用于检测训练任务是否受到影响，缩短故障检测时间，该功能的详细说明请参考[表1](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table17897155873217)。

**表 1** watchdog故障检测功能说明

<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table17897155873217"></a>
<table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row289715810326"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"></a>watchdog</span>故障检测。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1589716585326"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"></a>训练启动时，同时启动一个监测线程不断获取通信异常以及task执行异常。监测到故障发生后，快速抛出异常并终止训练任务进程，触发重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row189775853220"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"></a>仅支持MindSpore 2.4版本以上</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row5898058143215"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"></a>MindSpore<strong id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"></a>默认开启</strong><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"></a>watchdog</span>故障检测，<strong id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"></a>无需手动配置</strong>。如果需要关闭<span id="ph1052517411176"><a name="ph1052517411176"></a><a name="ph1052517411176"></a>watchdog</span>故障检测，请在模型配置文件中新增如下加粗字段。</p>
<pre class="screen" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"></a>...
context:
  <strong id="b393317297113"><a name="b393317297113"></a><a name="b393317297113"></a>ascend_config:</strong>
    <strong id="b12660461696"><a name="b12660461696"></a><a name="b12660461696"></a>hccl_watchdog: False</strong>    
...</pre>
</td>
</tr>
</tbody>
</table>

**训练回滚及加载CheckPoint时间优化<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section2743164217401"></a>**

-   异步保存CheckPoint：训练任务会定期保存CheckPoint文件，用于保存参数信息，故障恢复需要从上一次保存的CheckPoint回滚恢复训练。由于每次保存CheckPoint文件均会浪费一定的训练时间，为了保证训练效率，保存CheckPoint的时间间隔通常较大，而保存间隔越大，每次故障时训练回滚浪费的时间就会越长。针对该问题，集群调度组件支持通过MindIO ACP异步保存CheckPoint，详细说明请参考[表2](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table56063271212)。

    **表 2**  异步保存CheckPoint功能说明

    <a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table56063271212"></a>
    <table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row4606162713214"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"></a>MindIO ACP</span>异步保存CheckPoint。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row260619272216"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"></a>从NPU中获取CheckPoint后，异步写入存储中，降低每次保存CheckPoint的训练损失和保存周期，从而降低训练回滚时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row126061827152113"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row136069278219"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"></a>安装和使用<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"></a>MindIO</span>组件，请参考<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"></a><a href="../references.md#checkpoint保存与加载优化">CheckPoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

-   高效恢复CheckPoint：回滚恢复训练时，通常需要从存储中加载保存的CheckPoint，由于CheckPoint数据量较大，直接从存储读取加载CheckPoint的耗时较长。针对该问题，集群调度组件支持通过MindIO ACP进行CheckPoint高效恢复，详细说明请参考[表3](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table66066274216)。

    **表 3**  CheckPoint高效恢复功能说明

    <a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table66066274216"></a>
    <table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row106071271216"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"></a>MindIO</span> CheckPoint高效恢复。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row360715276216"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"></a>将最新的CheckPoint存储到内存中，故障恢复时可直接从内存中读取CheckPoint，降低CheckPoint读取时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1860772716217"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row196071127102110"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"></a>安装和使用<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"></a>MindIO</span>组件，请参考<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"></a><a href="../references.md#checkpoint保存与加载优化">CheckPoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

**编译缓存时间优化<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section1139444324019"></a>**

断点续训过程中拉起训练时需要构建计算图，在大模型场景下，构建计算图并编译需要消耗大量时间。针对该问题，MindSpore支持在首次编译时将编译缓存文件进行存储，进行故障恢复时可以直接读取存储中的图编译缓存，降低图编译时间，详细说明请参考[表4](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table175224282139)。

**表 4**  图编译缓存功能说明

<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table175224282139"></a>
<table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row135238284132"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"></a>图编译缓存。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1052322818133"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"></a>图编译时加载存储中保存的图编译缓存文件，加载后可降低图编译时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row5523628191316"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"></a>仅支持MindSpore2.3.0及以上版本。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row352313282136"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><div class="p" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"></a>在训练的shell启动脚本中（例如train_start.sh），添加如下环境变量。<pre class="screen" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"></a>export MS_COMPILER_CACHE_ENABLE=1
export MS_COMPILER_CACHE_ENABLE=1  # 开启图编译缓存
export MS_COMPILER_CACHE_PATH=xxx  # 设置图编译缓存路径</pre>
</div>
</td>
</tr>
</tbody>
</table>



### 配置HCCL主动触发建链<a name="ZH-CN_TOPIC_0000002511346489"></a>

当故障发生在HCCL建链阶段时，会导致进程级别重调度或进程级别在线恢复失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考以下步骤进行提前建链，防止故障出现在HCCL建链阶段。

**PyTorch单算子场景<a name="section145466566911"></a>**

PyTorch单算子场景HCCL建链为懒加载模式，当建立Torch通信组后，该通信组下发的第一个算子将触发HCCL通信域的创建，创建后完成卡间建链。因此，如果需要在训练初始化阶段完成所有通信域的建链，只需要在初始化阶段给每个通信组下发一个通信算子。

以下为创建通信组主动创建的示例：

```
rank = 0 # 设置本进程rank
sub_ranks = [0, 1, 2]  # 假设为一个包含0、1、2的通信组
groupX = torch.distributed.new_group(ranks=sub_ranks,...) # 创建通信组X
test_tensor = torch.ones(1).to(f'npu:{rank}') * (rank + 1)  # 构建一个测试数据tensor
torch.distributed.all_reduce(test_tensor, op=dist.ReduceOp.SUM, group=groupX)  # 在通信组X执行all reduce算子
```



## 配置任务YAML<a name="ZH-CN_TOPIC_0000002479226518"></a>

### YAML参数说明<a name="ZH-CN_TOPIC_0000002479386550"></a>

如果是acjob任务，在配置YAML前，请先了解相关YAML参数说明，详细说明如[表1](#zh-cn_topic_0000002039339953_table11351193062117)所示。

每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../appendix.md#acjob关键字段说明)。

**表 1**  YAML参数说明

<a name="zh-cn_topic_0000002039339953_table11351193062117"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339953_row635183013217"><th class="cellrowborder" valign="top" width="25.042504250425047%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039339953_p94514441221"><a name="zh-cn_topic_0000002039339953_p94514441221"></a><a name="zh-cn_topic_0000002039339953_p94514441221"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="24.76247624762476%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039339953_p1045113449226"><a name="zh-cn_topic_0000002039339953_p1045113449226"></a><a name="zh-cn_topic_0000002039339953_p1045113449226"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="50.1950195019502%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039339953_p18451154472214"><a name="zh-cn_topic_0000002039339953_p18451154472214"></a><a name="zh-cn_topic_0000002039339953_p18451154472214"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339953_row43521630112117"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p64516449226"><a name="zh-cn_topic_0000002039339953_p64516449226"></a><a name="zh-cn_topic_0000002039339953_p64516449226"></a>(.kind=="AscendJob").metadata.labels.framework</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul745174412210"></a><a name="zh-cn_topic_0000002039339953_ul745174412210"></a><ul id="zh-cn_topic_0000002039339953_ul745174412210"><li>mindspore</li><li>pytorch</li><li>tensorflow</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1745112442227"><a name="zh-cn_topic_0000002039339953_p1745112442227"></a><a name="zh-cn_topic_0000002039339953_p1745112442227"></a>框架类型，目前只支持三种。</p>
</td>
</tr>
<tr id="row133645515273"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2436102814254"><a name="zh-cn_topic_0000001951418201_p2436102814254"></a><a name="zh-cn_topic_0000001951418201_p2436102814254"></a>(.kind=="AscendJob").metadata.labels.jobID</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p1619843317517"><a name="p1619843317517"></a><a name="p1619843317517"></a>当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p9111039132818"><a name="p9111039132818"></a><a name="p9111039132818"></a>该参数仅支持在<span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>上使用。</p>
</td>
</tr>
<tr id="row1199016622817"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p13524833182513"><a name="zh-cn_topic_0000001951418201_p13524833182513"></a><a name="zh-cn_topic_0000001951418201_p13524833182513"></a>(.kind=="AscendJob").metadata.labels.app</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p5524103317257"><a name="zh-cn_topic_0000001951418201_p5524103317257"></a><a name="zh-cn_topic_0000001951418201_p5524103317257"></a>表明MindIE Motor任务在Ascend Job中的角色，取值包括mindie-ms-controller、mindie-ms-coordinator、mindie-ms-server。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><div class="note" id="zh-cn_topic_0000001951418201_note4367125713295"><a name="zh-cn_topic_0000001951418201_note4367125713295"></a><div class="notebody"><a name="ul139591420161415"></a><a name="ul139591420161415"></a><ul id="ul139591420161415"><li>acjob的任务YAML同时包含jobID和app这2个字段时，<span id="zh-cn_topic_0000001951418201_ph1566531814589"><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a>Ascend Operator</span>组件会自动传入环境变量MINDX_TASK_ID、APP_TYPE及MINDX_SERVICE_IP，并将其标识为MindIE推理任务。</li><li>关于以上环境变量的详细说明请参见<a href="../appendix.md#环境变量说明">表2</a>。</li><li>该参数仅支持在<span id="ph1493312176292"><a name="ph1493312176292"></a><a name="ph1493312176292"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph1893331752914"><a name="ph1893331752914"></a><a name="ph1893331752914"></a>Atlas 800I A2 推理服务器</span>上使用。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1553412124289"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p143891219113814"><a name="p143891219113814"></a><a name="p143891219113814"></a><span>(.kind=="AscendJob").metadata.labels.mind-cluster/scaling-rule: scaling-rule</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p3389151918383"><a name="p3389151918383"></a><a name="p3389151918383"></a>标记扩缩容规则对应的ConfigMap名称。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p10101730182112"><a name="p10101730182112"></a><a name="p10101730182112"></a>仅支持MindIE Motor推理任务在<span id="ph13640202812297"><a name="ph13640202812297"></a><a name="ph13640202812297"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph136401628192912"><a name="ph136401628192912"></a><a name="ph136401628192912"></a>Atlas 800I A2 推理服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="row9133171112813"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p16388101913817"><a name="p16388101913817"></a><a name="p16388101913817"></a><span>(.kind=="AscendJob").metadata.labels.mind-cluster/group-name: group0</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p13387171983812"><a name="p13387171983812"></a><a name="p13387171983812"></a>标记扩缩容规则中对应的group名称。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p14289143313188"><a name="p14289143313188"></a><a name="p14289143313188"></a>仅支持MindIE Motor推理任务在<span id="ph156731391296"><a name="ph156731391296"></a><a name="ph156731391296"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph26731039182916"><a name="ph26731039182916"></a><a name="ph26731039182916"></a>Atlas 800I A2 推理服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row7208139102014"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1447163824215"><a name="zh-cn_topic_0000002039339953_p1447163824215"></a><a name="zh-cn_topic_0000002039339953_p1447163824215"></a>(.kind=="AscendJob").metadata.labels."ring-controller.atlas"</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul16300163019516"></a><a name="zh-cn_topic_0000002039339953_ul16300163019516"></a><ul id="zh-cn_topic_0000002039339953_ul16300163019516"><li>Atlas 800 训练服务器：ascend-910</li><li><span id="zh-cn_topic_0000002039339953_ph760316141835"><a name="zh-cn_topic_0000002039339953_ph760316141835"></a><a name="zh-cn_topic_0000002039339953_ph760316141835"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="zh-cn_topic_0000002039339953_ph163483412215"><a name="zh-cn_topic_0000002039339953_ph163483412215"></a><a name="zh-cn_topic_0000002039339953_ph163483412215"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span>、<span id="ph1086225213421"><a name="ph1086225213421"></a><a name="ph1086225213421"></a>Atlas 800I A3 超节点服务器</span>和<span id="zh-cn_topic_0000002039339953_ph960313141731"><a name="zh-cn_topic_0000002039339953_ph960313141731"></a><a name="zh-cn_topic_0000002039339953_ph960313141731"></a>Atlas 900 A3 SuperPoD 超节点</span>：ascend-<span id="zh-cn_topic_0000002039339953_ph1360341413317"><a name="zh-cn_topic_0000002039339953_ph1360341413317"></a><a name="zh-cn_topic_0000002039339953_ph1360341413317"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p43811639112614"><a name="zh-cn_topic_0000002039339953_p43811639112614"></a><a name="zh-cn_topic_0000002039339953_p43811639112614"></a>标识任务使用的芯片的产品类型。</p>
<p id="zh-cn_topic_0000002039339953_zh-cn_topic_0000001570873348_p19220131902512"><a name="zh-cn_topic_0000002039339953_zh-cn_topic_0000001570873348_p19220131902512"></a><a name="zh-cn_topic_0000002039339953_zh-cn_topic_0000001570873348_p19220131902512"></a>需要在<span id="zh-cn_topic_0000002039339953_ph12290749162911"><a name="zh-cn_topic_0000002039339953_ph12290749162911"></a><a name="zh-cn_topic_0000002039339953_ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1735283013214"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p8451174482214"><a name="zh-cn_topic_0000002039339953_p8451174482214"></a><a name="zh-cn_topic_0000002039339953_p8451174482214"></a>(.kind=="AscendJob").metadata.labels.tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul645114462218"></a><a name="zh-cn_topic_0000002039339953_ul645114462218"></a><ul id="zh-cn_topic_0000002039339953_ul645114462218"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度</li></ul>
<div class="note" id="zh-cn_topic_0000002039339953_note62680445222"><a name="zh-cn_topic_0000002039339953_note62680445222"></a><a name="zh-cn_topic_0000002039339953_note62680445222"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p14582042192513"><a name="zh-cn_topic_0000002039339953_p14582042192513"></a><a name="zh-cn_topic_0000002039339953_p14582042192513"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
<p id="zh-cn_topic_0000002039339953_p112971504251"><a name="zh-cn_topic_0000002039339953_p112971504251"></a><a name="zh-cn_topic_0000002039339953_p112971504251"></a></p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p6452124472216"><a name="zh-cn_topic_0000002039339953_p6452124472216"></a><a name="zh-cn_topic_0000002039339953_p6452124472216"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1528312443229"><a name="zh-cn_topic_0000002039339953_note1528312443229"></a><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul176092014030"></a><a name="zh-cn_topic_0000002039339953_ul176092014030"></a><ul id="zh-cn_topic_0000002039339953_ul176092014030"><li>交换机亲和性调度1.0版本支持<span id="zh-cn_topic_0000002039339953_ph1157665817140"><a name="zh-cn_topic_0000002039339953_ph1157665817140"></a><a name="zh-cn_topic_0000002039339953_ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="zh-cn_topic_0000002039339953_ph168598363399"><a name="zh-cn_topic_0000002039339953_ph168598363399"></a><a name="zh-cn_topic_0000002039339953_ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>；支持<span id="zh-cn_topic_0000002039339953_ph4181625925"><a name="zh-cn_topic_0000002039339953_ph4181625925"></a><a name="zh-cn_topic_0000002039339953_ph4181625925"></a>PyTorch</span>和<span id="zh-cn_topic_0000002039339953_ph61882510210"><a name="zh-cn_topic_0000002039339953_ph61882510210"></a><a name="zh-cn_topic_0000002039339953_ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="zh-cn_topic_0000002039339953_ph311717506401"><a name="zh-cn_topic_0000002039339953_ph311717506401"></a><a name="zh-cn_topic_0000002039339953_ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_2"><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a>Atlas A2 训练系列产品</term></span>；支持<span id="zh-cn_topic_0000002039339953_ph17383182419412"><a name="zh-cn_topic_0000002039339953_ph17383182419412"></a><a name="zh-cn_topic_0000002039339953_ph17383182419412"></a>PyTorch</span>框架。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row83521230102117"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1745254442212"><a name="zh-cn_topic_0000002039339953_p1745254442212"></a><a name="zh-cn_topic_0000002039339953_p1745254442212"></a>(.kind=="AscendJob").metadata.labels.pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul14521244102210"></a><a name="zh-cn_topic_0000002039339953_ul14521244102210"></a><ul id="zh-cn_topic_0000002039339953_ul14521244102210"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p12453044102217"><a name="zh-cn_topic_0000002039339953_p12453044102217"></a><a name="zh-cn_topic_0000002039339953_p12453044102217"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1430334413223"><a name="zh-cn_topic_0000002039339953_note1430334413223"></a><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul461013147314"></a><a name="zh-cn_topic_0000002039339953_ul461013147314"></a><ul id="zh-cn_topic_0000002039339953_ul461013147314"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li><li><span id="zh-cn_topic_0000002039339953_ph1061091414318"><a name="zh-cn_topic_0000002039339953_ph1061091414318"></a><a name="zh-cn_topic_0000002039339953_ph1061091414318"></a>TensorFlow</span>暂不支持Pod级别重调度。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row435215305211"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p18454184442214"><a name="zh-cn_topic_0000002039339953_p18454184442214"></a><a name="zh-cn_topic_0000002039339953_p18454184442214"></a>(.kind=="AscendJob").metadata.labels.process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul71592205015"></a><a name="zh-cn_topic_0000002039339953_ul71592205015"></a><ul id="zh-cn_topic_0000002039339953_ul71592205015"><li>on：开启进程级别重调度及进程级在线恢复。<p>进程级别重调度和优雅容错不能同时开启，若同时开启，断点续训将通过Job级别重调度恢复训练。</p>
</li><li>pause：暂时关闭进程级别重调度及进程级在线恢复。</li><li>off或不使用该字段：关闭进程级别重调度及进程级在线恢复。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p10523132473119"><a name="zh-cn_topic_0000002039339953_p10523132473119"></a><a name="zh-cn_topic_0000002039339953_p10523132473119"></a>Ascend Operator会根据用户配置的recover-strategy自动给任务打上process-recover-enable=on标签，无需用户手动指定。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1635217304212"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10454144492212"><a name="zh-cn_topic_0000002039339953_p10454144492212"></a><a name="zh-cn_topic_0000002039339953_p10454144492212"></a>(.kind=="AscendJob").metadata.annotations.recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p156641331161715"><a name="zh-cn_topic_0000002039339953_p156641331161715"></a><a name="zh-cn_topic_0000002039339953_p156641331161715"></a>任务可用恢复策略。</p>
<a name="zh-cn_topic_0000002039339953_ul6665173119177"></a><a name="zh-cn_topic_0000002039339953_ul6665173119177"></a><ul id="zh-cn_topic_0000002039339953_ul6665173119177"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li><span>recover-in-place：进程级原地恢复</span>。</li><li>elastic-training：弹性训练。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039339953_ul18941121318614"></a><a name="zh-cn_topic_0000002039339953_ul18941121318614"></a>recover-strategy配置在任务YAML的annotations下，取值为6种策略的随意组合，策略之间由逗号分割。
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row5353133052111"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1745411440228"><a name="zh-cn_topic_0000002039339953_p1745411440228"></a><a name="zh-cn_topic_0000002039339953_p1745411440228"></a>(.kind=="AscendJob").metadata.labels.subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul18716519102210"></a><a name="zh-cn_topic_0000002039339953_ul18716519102210"></a><ul id="zh-cn_topic_0000002039339953_ul18716519102210"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>hotSwitch：执行亚健康热切，拉起备份Pod后，暂停训练任务，并使用新节点重新拉起训练。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p10455114482217"><a name="zh-cn_topic_0000002039339953_p10455114482217"></a><a name="zh-cn_topic_0000002039339953_p10455114482217"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1751515456204"><a name="zh-cn_topic_0000002039339953_note1751515456204"></a><div class="notebody"><a name="ul24832019017"></a><a name="ul24832019017"></a><ul id="ul24832019017"><li>使用graceExit策略时，需保证训练框架能够接收SIGTERM信号并保存CKPT文件。</li><li>hotSwitch策略的使用约束请参见<a href="#亚健康热切">使用约束</a>。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row163537304214"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10887172213238"><a name="zh-cn_topic_0000002039339953_p10887172213238"></a><a name="zh-cn_topic_0000002039339953_p10887172213238"></a>(.kind=="AscendJob").specs.schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1888718226237"><a name="zh-cn_topic_0000002039339953_p1888718226237"></a><a name="zh-cn_topic_0000002039339953_p1888718226237"></a>默认值为“volcano”，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p17887322122310"><a name="zh-cn_topic_0000002039339953_p17887322122310"></a><a name="zh-cn_topic_0000002039339953_p17887322122310"></a><span id="zh-cn_topic_0000002039339953_ph6604131419312"><a name="zh-cn_topic_0000002039339953_ph6604131419312"></a><a name="zh-cn_topic_0000002039339953_ph6604131419312"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row11548142102118"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p15887122212235"><a name="zh-cn_topic_0000002039339953_p15887122212235"></a><a name="zh-cn_topic_0000002039339953_p15887122212235"></a>(.kind=="AscendJob").spec.runPolicy.schedulingPolicy.minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1188782292317"><a name="zh-cn_topic_0000002039339953_p1188782292317"></a><a name="zh-cn_topic_0000002039339953_p1188782292317"></a>默认值为任务总副本数</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p660420141935"><a name="zh-cn_topic_0000002039339953_p660420141935"></a><a name="zh-cn_topic_0000002039339953_p660420141935"></a><span id="zh-cn_topic_0000002039339953_ph16604181418316"><a name="zh-cn_topic_0000002039339953_ph16604181418316"></a><a name="zh-cn_topic_0000002039339953_ph16604181418316"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000002039339953_ph156050141033"><a name="zh-cn_topic_0000002039339953_ph156050141033"></a><a name="zh-cn_topic_0000002039339953_ph156050141033"></a>Volcano</span>时，任务运行总副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row6549642112110"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p108872226233"><a name="zh-cn_topic_0000002039339953_p108872226233"></a><a name="zh-cn_topic_0000002039339953_p108872226233"></a>(.kind=="AscendJob").spec.runPolicy.schedulingPolicy.queue</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p19887112222310"><a name="zh-cn_topic_0000002039339953_p19887112222310"></a><a name="zh-cn_topic_0000002039339953_p19887112222310"></a>默认值为“default”，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p3887422182315"><a name="zh-cn_topic_0000002039339953_p3887422182315"></a><a name="zh-cn_topic_0000002039339953_p3887422182315"></a><span id="zh-cn_topic_0000002039339953_ph10605114231"><a name="zh-cn_topic_0000002039339953_ph10605114231"></a><a name="zh-cn_topic_0000002039339953_ph10605114231"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000002039339953_ph1660520141632"><a name="zh-cn_topic_0000002039339953_ph1660520141632"></a><a name="zh-cn_topic_0000002039339953_ph1660520141632"></a>Volcano</span>时，任务所属队列。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1054916421215"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1388862217239"><a name="zh-cn_topic_0000002039339953_p1388862217239"></a><a name="zh-cn_topic_0000002039339953_p1388862217239"></a>（可选）(.kind=="AscendJob").spec.successPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul688815228238"></a><a name="zh-cn_topic_0000002039339953_ul688815228238"></a><ul id="zh-cn_topic_0000002039339953_ul688815228238"><li>默认值为空，若用户不填写该参数，则默认取空值。</li><li>AllWorkers</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1688812219234"><a name="zh-cn_topic_0000002039339953_p1688812219234"></a><a name="zh-cn_topic_0000002039339953_p1688812219234"></a>表明任务成功的前提。空值代表只需要一个Pod成功，整个任务判定为成功。取值为“AllWorkers”表示所有Pod都成功，任务才判定为成功。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row15549114252110"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p178882222231"><a name="zh-cn_topic_0000002039339953_p178882222231"></a><a name="zh-cn_topic_0000002039339953_p178882222231"></a>(.kind=="AscendJob").spec.replicaSpecs.[Master|Scheduler|Worker].template.spec.containers[0].name</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p28887225231"><a name="zh-cn_topic_0000002039339953_p28887225231"></a><a name="zh-cn_topic_0000002039339953_p28887225231"></a>ascend</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p0888172216231"><a name="zh-cn_topic_0000002039339953_p0888172216231"></a><a name="zh-cn_topic_0000002039339953_p0888172216231"></a>容器的名称必须是“ascend”。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row8549242152117"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p688811220233"><a name="zh-cn_topic_0000002039339953_p688811220233"></a><a name="zh-cn_topic_0000002039339953_p688811220233"></a>（可选）(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].ports</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p188882022132320"><a name="zh-cn_topic_0000002039339953_p188882022132320"></a><a name="zh-cn_topic_0000002039339953_p188882022132320"></a>若用户未进行设置，系统默认填写以下参数：</p>
<a name="zh-cn_topic_0000002039339953_ul1488862272310"></a><a name="zh-cn_topic_0000002039339953_ul1488862272310"></a><ul id="zh-cn_topic_0000002039339953_ul1488862272310"><li>name: ascendjob-port</li><li>containerPort: 2222</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p2980183592716"><a name="p2980183592716"></a><a name="p2980183592716"></a>分布式训练集合通信端口。<span class="parmname" id="parmname1198063542717"><a name="parmname1198063542717"></a><a name="parmname1198063542717"></a>“name”</span>取值只能为<span class="parmvalue" id="parmvalue17980153515270"><a name="parmvalue17980153515270"></a><a name="parmvalue17980153515270"></a>“ascendjob-port”</span>，<span class="parmname" id="parmname8980135102711"><a name="parmname8980135102711"></a><a name="parmname8980135102711"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1054994210210"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p14889142214238"><a name="zh-cn_topic_0000002039339953_p14889142214238"></a><a name="zh-cn_topic_0000002039339953_p14889142214238"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.replicas</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul28892022132316"></a><a name="zh-cn_topic_0000002039339953_ul28892022132316"></a><ul id="zh-cn_topic_0000002039339953_ul28892022132316"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p788952218239"><a name="zh-cn_topic_0000002039339953_p788952218239"></a><a name="zh-cn_topic_0000002039339953_p788952218239"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row13550142102114"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p188915220236"><a name="zh-cn_topic_0000002039339953_p188915220236"></a><a name="zh-cn_topic_0000002039339953_p188915220236"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].image</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p11889192242311"><a name="zh-cn_topic_0000002039339953_p11889192242311"></a><a name="zh-cn_topic_0000002039339953_p11889192242311"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p188942210238"><a name="zh-cn_topic_0000002039339953_p188942210238"></a><a name="zh-cn_topic_0000002039339953_p188942210238"></a>训练镜像名称，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row2256185652210"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10889022132310"><a name="zh-cn_topic_0000002039339953_p10889022132310"></a><a name="zh-cn_topic_0000002039339953_p10889022132310"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec. nodeSelector.host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1889192292316"><a name="zh-cn_topic_0000002039339953_p1889192292316"></a><a name="zh-cn_topic_0000002039339953_p1889192292316"></a>Arm环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000002039339953_p18889822162315"><a name="zh-cn_topic_0000002039339953_p18889822162315"></a><a name="zh-cn_topic_0000002039339953_p18889822162315"></a>x86_64环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p7889152202315"><a name="zh-cn_topic_0000002039339953_p7889152202315"></a><a name="zh-cn_topic_0000002039339953_p7889152202315"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000002039339953_p168891122112316"><a name="zh-cn_topic_0000002039339953_p168891122112316"></a><a name="zh-cn_topic_0000002039339953_p168891122112316"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row13257165619223"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p11889152272316"><a name="zh-cn_topic_0000002039339953_p11889152272316"></a><a name="zh-cn_topic_0000002039339953_p11889152272316"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec. nodeSelector.accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul461118141037"></a><a name="zh-cn_topic_0000002039339953_ul461118141037"></a><ul id="zh-cn_topic_0000002039339953_ul461118141037"><li><span id="zh-cn_topic_0000002039339953_ph136117141331"><a name="zh-cn_topic_0000002039339953_ph136117141331"></a><a name="zh-cn_topic_0000002039339953_ph136117141331"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000002039339953_ph26111143315"><a name="zh-cn_topic_0000002039339953_ph26111143315"></a><a name="zh-cn_topic_0000002039339953_ph26111143315"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000002039339953_ph26117141237"><a name="zh-cn_topic_0000002039339953_ph26117141237"></a><a name="zh-cn_topic_0000002039339953_ph26117141237"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002039339953_ph061115145320"><a name="zh-cn_topic_0000002039339953_ph061115145320"></a><a name="zh-cn_topic_0000002039339953_ph061115145320"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="zh-cn_topic_0000002039339953_ph1661112141731"><a name="zh-cn_topic_0000002039339953_ph1661112141731"></a><a name="zh-cn_topic_0000002039339953_ph1661112141731"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li><li><span id="zh-cn_topic_0000002039339953_ph1161214145319"><a name="zh-cn_topic_0000002039339953_ph1161214145319"></a><a name="zh-cn_topic_0000002039339953_ph1161214145319"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="zh-cn_topic_0000002039339953_ph116121514934"><a name="zh-cn_topic_0000002039339953_ph116121514934"></a><a name="zh-cn_topic_0000002039339953_ph116121514934"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-16</li><li><span id="zh-cn_topic_0000002039339953_ph1514953013253"><a name="zh-cn_topic_0000002039339953_ph1514953013253"></a><a name="zh-cn_topic_0000002039339953_ph1514953013253"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph7730165573912"><a name="ph7730165573912"></a><a name="ph7730165573912"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 350 标卡</span>：（可选）与node的accelerator-type标签保持一致即可。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><div class="p" id="zh-cn_topic_0000002039339953_p1989014223239"><a name="zh-cn_topic_0000002039339953_p1989014223239"></a><a name="zh-cn_topic_0000002039339953_p1989014223239"></a>根据需要运行训练任务的节点类型，选取不同的值。<div class="note" id="zh-cn_topic_0000002039339953_note1861316141738"><a name="zh-cn_topic_0000002039339953_note1861316141738"></a><a name="zh-cn_topic_0000002039339953_note1861316141738"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p1027616512420"><a name="zh-cn_topic_0000002039339953_p1027616512420"></a><a name="zh-cn_topic_0000002039339953_p1027616512420"></a><span id="zh-cn_topic_0000002039339953_ph9014016509"><a name="zh-cn_topic_0000002039339953_ph9014016509"></a><a name="zh-cn_topic_0000002039339953_ph9014016509"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001519959665_b168254314713"><a name="zh-cn_topic_0000001519959665_b168254314713"></a><a name="zh-cn_topic_0000001519959665_b168254314713"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</div>
</td>
</tr>
<tr id="row17952121918262"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p149528198263"><a name="p149528198263"></a><a name="p149528198263"></a><span id="ph12817122692620"><a name="ph12817122692620"></a><a name="ph12817122692620"></a>(.kind=="AscendJob").metadata.annotations.</span><span id="ph5346181126"><a name="ph5346181126"></a><a name="ph5346181126"></a>"</span><span id="ph19817116185120"><a name="ph19817116185120"></a><a name="ph19817116185120"></a>huawei.com/schedule_policy</span><span id="ph9572712181216"><a name="ph9572712181216"></a><a name="ph9572712181216"></a>"</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p1877410425111"><a name="p1877410425111"></a><a name="p1877410425111"></a><span id="ph135426519519"><a name="ph135426519519"></a><a name="ph135426519519"></a>目前支持</span><a href="#table1120511613153">表2</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
<div class="note" id="note1230363125010"><a name="note1230363125010"></a><a name="note1230363125010"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002511347099_p1767434372512"><a name="zh-cn_topic_0000002511347099_p1767434372512"></a><a name="zh-cn_topic_0000002511347099_p1767434372512"></a>仅支持在<span id="zh-cn_topic_0000002511347099_ph1331492318423"><a name="zh-cn_topic_0000002511347099_ph1331492318423"></a><a name="zh-cn_topic_0000002511347099_ph1331492318423"></a>Atlas 训练系列产品</span>、<span id="zh-cn_topic_0000002511347099_ph2314323124211"><a name="zh-cn_topic_0000002511347099_ph2314323124211"></a><a name="zh-cn_topic_0000002511347099_ph2314323124211"></a><term id="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>和<span id="zh-cn_topic_0000002511347099_ph531432344210"><a name="zh-cn_topic_0000002511347099_ph531432344210"></a><a name="zh-cn_topic_0000002511347099_ph531432344210"></a><term id="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>中使用该字段。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row142575564228"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p8290194641017"><a name="zh-cn_topic_0000002039339953_p8290194641017"></a><a name="zh-cn_topic_0000002039339953_p8290194641017"></a>(.kind=="AscendJob").metadata.annotations.sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p14755536454"><a name="zh-cn_topic_0000002039339953_p14755536454"></a><a name="zh-cn_topic_0000002039339953_p14755536454"></a>指定逻辑超节点芯片数量。</p>
<a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><ul id="zh-cn_topic_0000002039339953_ul10451144414619"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p1670155202912"><a name="p1670155202912"></a><a name="p1670155202912"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19701652112917"><a name="p19701652112917"></a><a name="p19701652112917"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note47015215291"><a name="note47015215291"></a><a name="note47015215291"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在<span id="zh-cn_topic_0000002511347099_ph34244153594"><a name="zh-cn_topic_0000002511347099_ph34244153594"></a><a name="zh-cn_topic_0000002511347099_ph34244153594"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row17257145682219"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1993916082411"><a name="zh-cn_topic_0000002039339953_p1993916082411"></a><a name="zh-cn_topic_0000002039339953_p1993916082411"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].resources.requests.<span id="ph8846537141214"><a name="ph8846537141214"></a><a name="ph8846537141214"></a>"</span>huawei.com/Ascend910<span id="ph1636632921214"><a name="ph1636632921214"></a><a name="ph1636632921214"></a>"</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><div class="p" id="zh-cn_topic_0000002039339953_p361318141733"><a name="zh-cn_topic_0000002039339953_p361318141733"></a><a name="zh-cn_topic_0000002039339953_p361318141733"></a><span id="zh-cn_topic_0000002039339953_ph106131514134"><a name="zh-cn_topic_0000002039339953_ph106131514134"></a><a name="zh-cn_topic_0000002039339953_ph106131514134"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000002039339953_ul126147145311"></a><a name="zh-cn_topic_0000002039339953_ul126147145311"></a><ul id="zh-cn_topic_0000002039339953_ul126147145311"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p3614151416317"><a name="zh-cn_topic_0000002039339953_p3614151416317"></a><a name="zh-cn_topic_0000002039339953_p3614151416317"></a><span id="zh-cn_topic_0000002039339953_ph261416147313"><a name="zh-cn_topic_0000002039339953_ph261416147313"></a><a name="zh-cn_topic_0000002039339953_ph261416147313"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000002039339953_ul1961418141132"></a><a name="zh-cn_topic_0000002039339953_ul1961418141132"></a><ul id="zh-cn_topic_0000002039339953_ul1961418141132"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p10614214538"><a name="zh-cn_topic_0000002039339953_p10614214538"></a><a name="zh-cn_topic_0000002039339953_p10614214538"></a>服务器（插<span id="zh-cn_topic_0000002039339953_ph13615131417315"><a name="zh-cn_topic_0000002039339953_ph13615131417315"></a><a name="zh-cn_topic_0000002039339953_ph13615131417315"></a>Atlas 300T 训练卡</span>）：<a name="zh-cn_topic_0000002039339953_ul1261519142311"></a><a name="zh-cn_topic_0000002039339953_ul1261519142311"></a><ul id="zh-cn_topic_0000002039339953_ul1261519142311"><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p11615161416311"><a name="zh-cn_topic_0000002039339953_p11615161416311"></a><a name="zh-cn_topic_0000002039339953_p11615161416311"></a><span id="ph9683124520355"><a name="ph9683124520355"></a><a name="ph9683124520355"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002039339953_ph46156141634"><a name="zh-cn_topic_0000002039339953_ph46156141634"></a><a name="zh-cn_topic_0000002039339953_ph46156141634"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="zh-cn_topic_0000002039339953_ul1961514143314"></a><a name="zh-cn_topic_0000002039339953_ul1961514143314"></a><ul id="zh-cn_topic_0000002039339953_ul1961514143314"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p15616514739"><a name="zh-cn_topic_0000002039339953_p15616514739"></a><a name="zh-cn_topic_0000002039339953_p15616514739"></a><span id="zh-cn_topic_0000002039339953_ph161611419319"><a name="zh-cn_topic_0000002039339953_ph161611419319"></a><a name="zh-cn_topic_0000002039339953_ph161611419319"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="zh-cn_topic_0000002039339953_ul661611418316"></a><a name="zh-cn_topic_0000002039339953_ul661611418316"></a><ul id="zh-cn_topic_0000002039339953_ul661611418316"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p136171314938"><a name="zh-cn_topic_0000002039339953_p136171314938"></a><a name="zh-cn_topic_0000002039339953_p136171314938"></a><span id="zh-cn_topic_0000002039339953_ph161712141313"><a name="zh-cn_topic_0000002039339953_ph161712141313"></a><a name="zh-cn_topic_0000002039339953_ph161712141313"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="zh-cn_topic_0000002039339953_ph188291824164611"><a name="zh-cn_topic_0000002039339953_ph188291824164611"></a><a name="zh-cn_topic_0000002039339953_ph188291824164611"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 超节点服务器</span>：<a name="zh-cn_topic_0000002039339953_ul261751412316"></a><a name="zh-cn_topic_0000002039339953_ul261751412316"></a><ul id="zh-cn_topic_0000002039339953_ul261751412316"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对<span id="ph6583110111012"><a name="ph6583110111012"></a><a name="ph6583110111012"></a>Atlas 900 A3 SuperPoD 超节点</span>的逻辑超节点亲和任务：16</li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p3943130162412"><a name="zh-cn_topic_0000002039339953_p3943130162412"></a><a name="zh-cn_topic_0000002039339953_p3943130162412"></a>请求的NPU数量，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row11257156102214"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p894317013244"><a name="zh-cn_topic_0000002039339953_p894317013244"></a><a name="zh-cn_topic_0000002039339953_p894317013244"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p159431102243"><a name="zh-cn_topic_0000002039339953_p159431102243"></a><a name="zh-cn_topic_0000002039339953_p159431102243"></a>取值为metadata.annotations['huawei.com/AscendXXX']，其中XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p136226142031"><a name="zh-cn_topic_0000002039339953_p136226142031"></a><a name="zh-cn_topic_0000002039339953_p136226142031"></a><span id="zh-cn_topic_0000002039339953_ph1062212140315"><a name="zh-cn_topic_0000002039339953_ph1062212140315"></a><a name="zh-cn_topic_0000002039339953_ph1062212140315"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note462214141730"><a name="zh-cn_topic_0000002039339953_note462214141730"></a><a name="zh-cn_topic_0000002039339953_note462214141730"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p186225141637"><a name="zh-cn_topic_0000002039339953_p186225141637"></a><a name="zh-cn_topic_0000002039339953_p186225141637"></a>该参数只支持使用<span id="zh-cn_topic_0000002039339953_ph962251412315"><a name="zh-cn_topic_0000002039339953_ph962251412315"></a><a name="zh-cn_topic_0000002039339953_ph962251412315"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row925815602217"><td class="cellrowborder" rowspan="5" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p39441803247"><a name="zh-cn_topic_0000002039339953_p39441803247"></a><a name="zh-cn_topic_0000002039339953_p39441803247"></a>(.kind=="AscendJob").metadata.labels.fault-scheduling</p>
<p id="zh-cn_topic_0000002039339953_p54121202412"><a name="zh-cn_topic_0000002039339953_p54121202412"></a><a name="zh-cn_topic_0000002039339953_p54121202412"></a></p>
<p id="zh-cn_topic_0000002039339953_p13419162418"><a name="zh-cn_topic_0000002039339953_p13419162418"></a><a name="zh-cn_topic_0000002039339953_p13419162418"></a></p>
<p id="zh-cn_topic_0000002039339953_p23118240"><a name="zh-cn_topic_0000002039339953_p23118240"></a><a name="zh-cn_topic_0000002039339953_p23118240"></a></p>
<p id="zh-cn_topic_0000002039339953_p8211122417"><a name="zh-cn_topic_0000002039339953_p8211122417"></a><a name="zh-cn_topic_0000002039339953_p8211122417"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p494450122411"><a name="zh-cn_topic_0000002039339953_p494450122411"></a><a name="zh-cn_topic_0000002039339953_p494450122411"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p1028610425357"><a name="p1028610425357"></a><a name="p1028610425357"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="ph19623131417313"><a name="ph19623131417313"></a><a name="ph19623131417313"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="ph96231114734"><a name="ph96231114734"></a><a name="ph96231114734"></a>Pod</span>。</p>
<p id="p1462216142314"><a name="p1462216142314"></a><a name="p1462216142314"></a>进程级别重调度和进程级在线恢复场景，需将本参数配置为grace。</p>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1258135615221"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p8944120192414"><a name="zh-cn_topic_0000002039339953_p8944120192414"></a><a name="zh-cn_topic_0000002039339953_p8944120192414"></a>force</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p562301420319"><a name="p562301420319"></a><a name="p562301420319"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="ph19623151420318"><a name="ph19623151420318"></a><a name="ph19623151420318"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1125885682215"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p4944603241"><a name="zh-cn_topic_0000002039339953_p4944603241"></a><a name="zh-cn_topic_0000002039339953_p4944603241"></a>off</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.2 "><p id="p26233141317"><a name="p26233141317"></a><a name="p26233141317"></a>该任务不使用断点续训特性，<span id="ph8623191418313"><a name="ph8623191418313"></a><a name="ph8623191418313"></a>K8s</span>的maxRetry仍然生效。</p>
<p id="p186239141631"><a name="p186239141631"></a><a name="p186239141631"></a></p>
<p id="p1623191419310"><a name="p1623191419310"></a><a name="p1623191419310"></a></p>
<p id="p10269114741310"><a name="p10269114741310"></a><a name="p10269114741310"></a></p>
<p id="p1326917477139"><a name="p1326917477139"></a><a name="p1326917477139"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1225812563228"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1794470102412"><a name="zh-cn_topic_0000002039339953_p1794470102412"></a><a name="zh-cn_topic_0000002039339953_p1794470102412"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row870173811239"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1494460142412"><a name="zh-cn_topic_0000002039339953_p1494460142412"></a><a name="zh-cn_topic_0000002039339953_p1494460142412"></a>其他值</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row97116382233"><td class="cellrowborder" rowspan="2" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p149441205245"><a name="zh-cn_topic_0000002039339953_p149441205245"></a><a name="zh-cn_topic_0000002039339953_p149441205245"></a>(.kind=="AscendJob").metadata.labels.fault-retry-times</p>
<p id="zh-cn_topic_0000002039339953_p18018142420"><a name="zh-cn_topic_0000002039339953_p18018142420"></a><a name="zh-cn_topic_0000002039339953_p18018142420"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1394470122411"><a name="zh-cn_topic_0000002039339953_p1394470122411"></a><a name="zh-cn_topic_0000002039339953_p1394470122411"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p13944180132415"><a name="zh-cn_topic_0000002039339953_p13944180132415"></a><a name="zh-cn_topic_0000002039339953_p13944180132415"></a>处理业务面故障，必须配置业务面可无条件重试的次数。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note146647015242"><a name="zh-cn_topic_0000002039339953_note146647015242"></a><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul13624161415314"></a><a name="zh-cn_topic_0000002039339953_ul13624161415314"></a><ul id="zh-cn_topic_0000002039339953_ul13624161415314"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph41689437364"><a name="ph41689437364"></a><a name="ph41689437364"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002039339953_ph166251314730"><a name="zh-cn_topic_0000002039339953_ph166251314730"></a><a name="zh-cn_topic_0000002039339953_ph166251314730"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row17113822319"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p159451022413"><a name="zh-cn_topic_0000002039339953_p159451022413"></a><a name="zh-cn_topic_0000002039339953_p159451022413"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p894515062415"><a name="zh-cn_topic_0000002039339953_p894515062415"></a><a name="zh-cn_topic_0000002039339953_p894515062415"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row12722038182310"><td class="cellrowborder" rowspan="2" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p39451803247"><a name="zh-cn_topic_0000002039339953_p39451803247"></a><a name="zh-cn_topic_0000002039339953_p39451803247"></a>(.kind=="AscendJob").spec.runPolicy.backoffLimit</p>
<p id="zh-cn_topic_0000002039339953_p1999716016245"><a name="zh-cn_topic_0000002039339953_p1999716016245"></a><a name="zh-cn_topic_0000002039339953_p1999716016245"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1894519012417"><a name="zh-cn_topic_0000002039339953_p1894519012417"></a><a name="zh-cn_topic_0000002039339953_p1894519012417"></a>0 &lt; backoffLimit</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1194560112412"><a name="zh-cn_topic_0000002039339953_p1194560112412"></a><a name="zh-cn_topic_0000002039339953_p1194560112412"></a>任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与backoffLimit取值相同时，任务将不再进行重调度。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1680005244"><a name="zh-cn_topic_0000002039339953_note1680005244"></a><div class="notebody"><p id="zh-cn_topic_0000002039339953_p59459062418"><a name="zh-cn_topic_0000002039339953_p59459062418"></a><a name="zh-cn_topic_0000002039339953_p59459062418"></a>同时配置了backoffLimit和fault-retry-times参数时，当已经重调度次数与backoffLimit或fault-retry-times取值有一个相同时，将不再进行重调度。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row167283811232"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p6945904245"><a name="zh-cn_topic_0000002039339953_p6945904245"></a><a name="zh-cn_topic_0000002039339953_p6945904245"></a>无（无backoffLimit）或backoffLimit ≤ 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1594511016244"><a name="zh-cn_topic_0000002039339953_p1594511016244"></a><a name="zh-cn_topic_0000002039339953_p1594511016244"></a>不限制总重调度次数。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note176907017241"><a name="zh-cn_topic_0000002039339953_note176907017241"></a><div class="notebody"><p id="zh-cn_topic_0000002039339953_p159468016247"><a name="zh-cn_topic_0000002039339953_p159468016247"></a><a name="zh-cn_topic_0000002039339953_p159468016247"></a>若不配置backoffLimit，但是配置了fault-retry-times参数，则使用fault-retry-times的重调度次数。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1372163816239"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p4946190192415"><a name="zh-cn_topic_0000002039339953_p4946190192415"></a><a name="zh-cn_topic_0000002039339953_p4946190192415"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul59469052410"></a><a name="zh-cn_topic_0000002039339953_ul59469052410"></a><ul id="zh-cn_topic_0000002039339953_ul59469052410"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。</li></ul>
<div class="note" id="zh-cn_topic_0000002039339953_note8696110162419"><a name="zh-cn_topic_0000002039339953_note8696110162419"></a><a name="zh-cn_topic_0000002039339953_note8696110162419"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p1694680112418"><a name="zh-cn_topic_0000002039339953_p1694680112418"></a><a name="zh-cn_topic_0000002039339953_p1694680112418"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1594690102418"><a name="zh-cn_topic_0000002039339953_p1594690102418"></a><a name="zh-cn_topic_0000002039339953_p1594690102418"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为“Never”。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row18731938162313"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10946140152416"><a name="zh-cn_topic_0000002039339953_p10946140152416"></a><a name="zh-cn_topic_0000002039339953_p10946140152416"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1694690152418"><a name="zh-cn_topic_0000002039339953_p1694690152418"></a><a name="zh-cn_topic_0000002039339953_p1694690152418"></a>0 &lt; terminationGracePeriodSeconds &lt; <strong id="zh-cn_topic_0000002039339953_b09468052417"><a name="zh-cn_topic_0000002039339953_b09468052417"></a><a name="zh-cn_topic_0000002039339953_b09468052417"></a>grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1494610072416"><a name="zh-cn_topic_0000002039339953_p1494610072416"></a><a name="zh-cn_topic_0000002039339953_p1494610072416"></a>容器收到SIGTERM到被K8s强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000002039339953_i1394616092412"><a name="zh-cn_topic_0000002039339953_i1394616092412"></a><a name="zh-cn_topic_0000002039339953_i1394616092412"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000002039339953_b794617013242"><a name="zh-cn_topic_0000002039339953_b794617013242"></a><a name="zh-cn_topic_0000002039339953_b794617013242"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参考K8s官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note5717204249"><a name="zh-cn_topic_0000002039339953_note5717204249"></a><div class="notebody"><p id="zh-cn_topic_0000002039339953_p1894750112418"><a name="zh-cn_topic_0000002039339953_p1894750112418"></a><a name="zh-cn_topic_0000002039339953_p1894750112418"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row15963544152013"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p7418049182319"><a name="zh-cn_topic_0000002039339953_p7418049182319"></a><a name="zh-cn_topic_0000002039339953_p7418049182319"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.hostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul960434424111"></a><a name="zh-cn_topic_0000002039339953_ul960434424111"></a><ul id="zh-cn_topic_0000002039339953_ul960434424111"><li>true：使用HostIP创建Pod。</li><li>false：不使用HostIP创建Pod。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039339953_ul14611159182815"></a><a name="zh-cn_topic_0000002039339953_ul14611159182815"></a><ul id="zh-cn_topic_0000002039339953_ul14611159182815"><li>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</li><li>不传入此参数时，默认不使用HostIP创建Pod。<div class="note" id="zh-cn_topic_0000002039339953_note1423653119592"><a name="zh-cn_topic_0000002039339953_note1423653119592"></a><a name="zh-cn_topic_0000002039339953_note1423653119592"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p461933317584"><a name="zh-cn_topic_0000002039339953_p461933317584"></a><a name="zh-cn_topic_0000002039339953_p461933317584"></a>当HostNetwork取值为true时，若当前任务YAML挂载了<span id="ph01944310814"><a name="ph01944310814"></a><a name="ph01944310814"></a>RankTable</span>文件路径，则可以通过在训练脚本中解析<span id="ph158525327162"><a name="ph158525327162"></a><a name="ph158525327162"></a>RankTable</span>文件获取Pod的hostIP来实现建链。若任务YAML未挂载<span id="ph094714211613"><a name="ph094714211613"></a><a name="ph094714211613"></a>RankTable</span>文件路径，则与原始保持一致，使用serviceIP来实现建链。</p>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row13351447164012"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p6351124764011"><a name="p6351124764011"></a><a name="p6351124764011"></a>(.kind=="AscendJob").metadata.annotations.wait-reschedule-timeout</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p1235154718404"><a name="p1235154718404"></a><a name="p1235154718404"></a>30~270</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p8351747134014"><a name="p8351747134014"></a><a name="p8351747134014"></a>进程级别重调度处理时等待故障节点重调度的超时时间，单位为秒，默认值为270。</p>
</td>
</tr>
</tbody>
</table>

**表 2**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002511347099_row192066612155"><th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002511347099_p132062614153"><a name="zh-cn_topic_0000002511347099_p132062614153"></a><a name="zh-cn_topic_0000002511347099_p132062614153"></a>配置</p>
</th>
<th class="cellrowborder" valign="top" width="77.7%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002511347099_p5206126181520"><a name="zh-cn_topic_0000002511347099_p5206126181520"></a><a name="zh-cn_topic_0000002511347099_p5206126181520"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002511347099_row201261346162"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p457945418181"><a name="zh-cn_topic_0000002511347099_p457945418181"></a><a name="zh-cn_topic_0000002511347099_p457945418181"></a>chip4-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p7579105411817"><a name="zh-cn_topic_0000002511347099_p7579105411817"></a><a name="zh-cn_topic_0000002511347099_p7579105411817"></a>1个节点8张芯片，每4个芯片形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph18314192319429"><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph631452384213"><a name="zh-cn_topic_0000002511347099_ph631452384213"></a><a name="zh-cn_topic_0000002511347099_ph631452384213"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的整模块场景 /Atlas 350 推理卡内部共8张卡，每4张卡通过UB扣板连接。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row102574171610"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p205801254151810"><a name="zh-cn_topic_0000002511347099_p205801254151810"></a><a name="zh-cn_topic_0000002511347099_p205801254151810"></a>chip1-node2</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p65801354101816"><a name="zh-cn_topic_0000002511347099_p65801354101816"></a><a name="zh-cn_topic_0000002511347099_p65801354101816"></a>1个节点2张芯片。例如，<span id="zh-cn_topic_0000002511347099_ph97657495514"><a name="zh-cn_topic_0000002511347099_ph97657495514"></a><a name="zh-cn_topic_0000002511347099_ph97657495514"></a>Atlas 300T 训练卡</span>的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row825811151619"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p17580854201815"><a name="zh-cn_topic_0000002511347099_p17580854201815"></a><a name="zh-cn_topic_0000002511347099_p17580854201815"></a>chip4-node4</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p858019546184"><a name="zh-cn_topic_0000002511347099_p858019546184"></a><a name="zh-cn_topic_0000002511347099_p858019546184"></a>1个节点4张芯片，形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph1165491719811"><a name="zh-cn_topic_0000002511347099_ph1165491719811"></a><a name="zh-cn_topic_0000002511347099_ph1165491719811"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph15654111712815"><a name="zh-cn_topic_0000002511347099_ph15654111712815"></a><a name="zh-cn_topic_0000002511347099_ph15654111712815"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的半配场景。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p2580654131819"><a name="zh-cn_topic_0000002511347099_p2580654131819"></a><a name="zh-cn_topic_0000002511347099_p2580654131819"></a>chip8-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p85801654181818"><a name="zh-cn_topic_0000002511347099_p85801654181818"></a><a name="zh-cn_topic_0000002511347099_p85801654181818"></a>1个节点8张卡，8张卡都在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph14314162316427"><a name="zh-cn_topic_0000002511347099_ph14314162316427"></a><a name="zh-cn_topic_0000002511347099_ph14314162316427"></a>Atlas 800T A2 训练服务器</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row1820613612158"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p1358111544185"><a name="zh-cn_topic_0000002511347099_p1358111544185"></a><a name="zh-cn_topic_0000002511347099_p1358111544185"></a>chip8-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p9581135461815"><a name="zh-cn_topic_0000002511347099_p9581135461815"></a><a name="zh-cn_topic_0000002511347099_p9581135461815"></a>1个节点16张卡，每8张卡在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph1831422311424"><a name="zh-cn_topic_0000002511347099_ph1831422311424"></a><a name="zh-cn_topic_0000002511347099_ph1831422311424"></a>Atlas 200T A2 Box16 异构子框</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row2020613616154"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p2581854121811"><a name="zh-cn_topic_0000002511347099_p2581854121811"></a><a name="zh-cn_topic_0000002511347099_p2581854121811"></a>chip2-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p758125481813"><a name="zh-cn_topic_0000002511347099_p758125481813"></a><a name="zh-cn_topic_0000002511347099_p758125481813"></a>1个节点16张卡，每2张卡在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph855133261011"><a name="zh-cn_topic_0000002511347099_ph855133261011"></a><a name="zh-cn_topic_0000002511347099_ph855133261011"></a>Atlas 800T A3 超节点服务器</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row22064621511"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p558111549188"><a name="zh-cn_topic_0000002511347099_p558111549188"></a><a name="zh-cn_topic_0000002511347099_p558111549188"></a>chip2-node16-sp</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p258115548187"><a name="zh-cn_topic_0000002511347099_p258115548187"></a><a name="zh-cn_topic_0000002511347099_p258115548187"></a>1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，<span id="zh-cn_topic_0000002511347099_ph1990844161011"><a name="zh-cn_topic_0000002511347099_ph1990844161011"></a><a name="zh-cn_topic_0000002511347099_ph1990844161011"></a>Atlas 900 A3 SuperPoD 超节点</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip4-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点16张卡，每4张卡都在1个互联环上。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共16张卡，每4张卡通过UB扣板连接</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip1-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点8张卡，每张卡之间无互联。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共8张卡，每张卡之间无互联</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip1-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点16张卡，每张卡之间无互联。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共16张卡，每张卡之间无互联</span>。</p>
</td>
</tr>
</tbody>
</table>


### 任务YAML配置示例<a name="ZH-CN_TOPIC_0000002511346461"></a>

重调度模式和优雅容错模式可参考如下[操作步骤](#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219)配置示例。当**subHealthyStrategy**取值为graceExit时，需要参考[（可选）修改训练脚本](#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section1048332432310)适配启动脚本确保训练框架能够配合重调度。

**前提条件<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section7585519135117"></a>**

用户已创建[hccl.json](../appendix.md#hccljson文件说明)文件的具体挂载路径，详细操作步骤请参见[步骤4](../installation_guide.md#ascend-operator)。

**操作步骤<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219"></a>**

1.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。
    -   以a800\_AscendJob\__\{xxx\}_b.yaml为例，在一台Atlas 200T A2 Box16 异构子框节点创建**分布式训练**任务，任务使用2\*4个芯片，修改示例如下。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-mindspore
          labels:
            framework: mindspore  # 训练框架名称
            fault-scheduling: "grace"     # 开启优雅删除模式
            ring-controller.atlas: ascend-{xxx}b
            fault-retry-times: "3"            # 开启业务面故障无条件重试能力，同时需要将restartPolicy取值设置为Never
            tor-affinity: "normal-schema" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
            pod-rescheduling: "on"     # 开启Pod级别重调度
            subHealthyStrategy: "ignore"  # 忽略健康状态为亚健康的节点，后续任务在亲和性调度上不优先调度该节点
        spec:
          schedulerName: volcano    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            backoffLimit: 3      # 任务重调度次数
            schedulingPolicy:
              minAvailable: 3       # 任务总副本数
              queue: default     # 任务所属队列
          successPolicy: AllWorkers  # 任务成功的前提
          replicaSpecs:
            Scheduler:
              replicas: 1            #只能为1
              restartPolicy:  Never   #容器重启策略
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  terminationGracePeriodSeconds: 360  #容器收到SIGTERM到被K8s强制停止经历的时间
                  nodeSelector:                       
                    host-arch: huawei-x86          # Atlas 200T A2 Box16 异构子框只有x86_64架构
                    accelerator-type: module-{xxx}b-16   # 节点类型
                  containers:
                  - name: ascend     # 不能修改
        ...
                    ports:                     # 可选，分布式训练集合通信端口
                      - containerPort: 2222    
                        name: ascendjob-port 
                    volumeMounts:
        ...
          
            Worker:
              replicas: 2
              restartPolicy: Never  # 容器重启策略
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  terminationGracePeriodSeconds: 360   #容器收到SIGTERM到被K8s强制停止经历的时间
                  affinity:
        ...
                  nodeSelector:           
                    host-arch: huawei-x86      # Atlas 200T A2 Box16 异构子框只有x86_64架构
                    accelerator-type: module-{xxx}b-16   # 节点类型
                  containers:
                  - name: ascend      # 不能修改
        ...
                    env:
                    - name: ASCEND_VISIBLE_DEVICES
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.annotations['huawei.com/Ascend910']         # 需要和下面resources和requests保持一致     
        ...
        
                    ports:        # 可选，分布式训练集合通信端口
                      - containerPort: 2222    
                        name: ascendjob-port  
                    resources:
                      limits:
                        huawei.com/Ascend910: 4      # 需要的NPU芯片个数为4
                      requests:
                        huawei.com/Ascend910: 4       # 与limits取值一致
        ```

    -   以a800\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用8个芯片，修改示例如下。

        ```
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-910  # 标识产品类型
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API。
        kind: Job                               # 目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名，可自定义
          labels:
            ring-controller.atlas: ascend-910   
            fault-scheduling: "force"        # 开启强制删除模式
            fault-retry-times: "3"            # 开启业务面故障无条件重试能力，同时需要将restartPolicy取值设置为Never；并将policies的event设置为PodFailed，action设置为Ignore
            tor-affinity: "normal-schema" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
            pod-rescheduling: "on"     # 开启Pod级别重调度
            subHealthyStrategy: "ignore"     # 忽略健康状态为亚健康的节点，后续任务在亲和性调度上不优先调度该节点
        ...
        spec:
          policies:  #使用Pod级别重调度，需要删除policies及其子参数event和action
            - event: PodEvicted   # 使用业务面故障无条件重试（或同时使用Pod级别重调度和业务面故障无条件重试），需要将event设置为PodFailed
              action: RestartJob  # 使用业务面故障无条件重试（或同时使用Pod级别重调度和业务面故障无条件重试），需要将action设置为Ignore
        ...
          minAvailable: 1                  # 单机为1
        ...
          maxRetry: 3              # 重调度次数
        ...
          - name: "default-test"
              replicas: 1                  # 单机为1
              template:
                metadata:
        ...
                spec:
                  terminationGracePeriodSeconds: 360  #容器收到SIGTERM到被K8s强制停止经历的时间
        ...
                    env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources和requests保持一致
        ...
                    resources:  
                      requests:
                        huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8。可在下方添加行，配置memory、cpu等资源
                      limits:
                        huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
        ...
                    nodeSelector:
                      host-arch: huawei-arm               # 可选值，根据实际情况填写
                      accelerator-type: module      #调度到Atlas 800 训练服务器节点上
        ...
                restartPolicy: Never   # 容器重启策略
        ```

2.  配置MindIO的通信地址。在代码中新增以下加粗内容。

    ```
    ...
       Master:
    ...
                env:        
                  - name: POD_IP
                    valueFrom:
                      fieldRef:
                        fieldPath: status.podIP             # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。
    ```

3.  （可选）如果开启了临终遗言，需要在训练YAML中增加临终遗言通信的端口信息，以pytorch\_multinodes\_acjob\__\{xxx\}_b.yaml为例，新增以下加粗内容。

    ```
    ...
       Master:
    ...
              env:
                  - name: TTP_PORT                  
                    value: "8000"     # 用于临终遗言通信，请注意上下保持一致
    ...
                ports:                         
                    - containerPort: 2222        
                      name: ascendjob-port       
                  - containerPort: 8000     # 用于临终遗言通信，请注意上下保持一致
                    name: ttp-port
                  - containerPort: 9601     # TaskD Pod间通信端口
                    name: taskd-port
    ...
       Worker:
    ...
              env:
                  - name: TTP_PORT                  
                    value: "8000"            # 用于临终遗言通信，请注意上下保持一致
    ...
                ports:                          
                    - containerPort: 2222         
                      name: ascendjob-port       
                  - containerPort: 8000     # 用于临终遗言通信，请注意上下保持一致
                    name: ttp-port
                  - containerPort: 9601     # TaskD Pod间通信端口
                    name: taskd-port
    
    ...
    ```

4.  （可选）如果使用临终遗言和进程级恢复，需要在训练YAML中增加临终遗言通信的端口信息和进程级恢复开关等信息，以pytorch\_multinodes\_acjob\__\{xxx\}_b.yaml为例，新增以下加粗内容。

    ```
    ...
      labels:    
           framework: pytorch   
           ring-controller.atlas: ascend-{xxx}b    
           fault-scheduling: "grace"    
           fault-retry-times: "10"   // 开启无条件重试
           pod-rescheduling: "on"   // 开启Pod级重调度
           tor-affinity: "null" # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不适用。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
    ...
      annotations:  
         ...  
         recover-strategy: "recover,dump"
      replicaSpecs:    
          Master:     
            replicas: 1      
            restartPolicy: Never      
            template:        
                metadata:
    ...
               - name: TTP_PORT
                 value: "8000"  # 用于MindIO通信，请注意上下保持一致
            command:                           # training command, which can be modified             
              - /bin/bash              
              - -c            
            args:
              - | 
                cd /job/code; 
                chmod +x scripts/train_start.sh; 
                bash scripts/train_start.sh
             ports:                          # default value 
               - containerPort: 2222 
                 name: ascendjob-port if not set              
              - containerPort: 8000    # 用于MindIO通信，请注意上下保持一致
               name: ttp-port
             - containerPort: 9601    # TaskD Pod间通信端口
               name: taskd-port
    ...
    
    ...
      replicaSpecs:    
          Worker:     
            replicas: 1      
            restartPolicy: Never      
            template:        
                metadata:
    ...
                - name: TTP_PORT
                value: "8000"  # 用于MindIO通信，请注意上下保持一致
            command:                           # training command, which can be modified             
              - /bin/bash              
              - -c            
            args:
              - | 
                cd /job/code; 
                chmod +x scripts/train_start.sh; 
                bash scripts/train_start.sh
             ports:                          # default value 
               - containerPort: 2222 
                 name: ascendjob-port if not set              
              - containerPort: 8000    # 用于MindIO通信，请注意上下保持一致
               name: ttp-port
             - containerPort: 9601    # TaskD Pod间通信端口
               name: taskd-port
    ...
    ```

5.  使用断点续训功能，建议扩展内存，请按注释添加参数，示例如下。

    ```
    ...
              volumeMounts:                             #断点续训扩容
             - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

6.  若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu“和“memory“参数和对应的参数值，具体数值请根据实际情况配置。

    ```
    ...
              resources:  
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m               
                  memory: 100Gi           
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

7.  修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径        
    ```

8.  **（可选）**如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径（PyTorch命令参数不涉及启动脚本）。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。

    >[!NOTE] 说明 
    >使用**优雅容错模式**可跳过该步骤。

    -   **TensorFlow命令参数**

        ```
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ tensorflow/resnet_ctl_imagenet_main.py --data_dir=/job/data/imagenet_TF --distribution_strategy=one_device --use_tf_while_loop=true --epochs_between_evals=1 --skip_eval --enable_checkpoint_and_export;"
        ...
        ```

    -   **PyTorch命令参数**

        ```
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024 --resume=true;"
        ...
        ```

    -   使用**MindSpore架构**的模型**，**包括ResNet50模型和Pangu\_alpha模型需要跳过此步骤。

9.  选择存储方式。
    -   （可选）NFS场景需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

        >[!NOTE] 说明 
        >请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

        ```
        ...
                  volumeMounts:
                  - name: ascend-910-config
                    mountPath: /user/serverid/devindex/config
                  - name: code
                    mountPath: /job/code                     # 容器中训练脚本路径
                  - name: data
                    mountPath: /job/data                      # 容器中训练数据集路径
                  - name: output
                    mountPath: /job/output                    # 容器中训练输出路径
        ...
                   # 可选，使用Ascend Operator组件为训练任务生成RankTable文件，需要新增以下加粗字段，设置容器中hccl.json文件保存路径，该路径不可修改。
                  - name: ranktable        
                    mountPath: /user/serverid/devindex/config
        ...
                volumes:
        ...
                - name: code
                  nfs:
                    server: 127.0.0.1        # NFS服务器IP地址
                    path: "xxxxxx"           # 配置训练脚本路径
                - name: data
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # 配置训练集路径
                - name: output
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # 设置脚本相关模型的保存路径
        ...
                   # 可选，使用组件为PyTorch框架生成RankTable文件，需要新增以下加粗字段，设置hccl.json文件保存路径
                - name: ranktable         #请勿修改此参数的默认值，Ascend Operator会用于检查是否开启文件挂载hccl.json。
                  hostPath:                    #请使用hostpath挂载或NFS挂载
                    path: /user/mindx-dl/ranktable/default.default-test-pytorch   # 共享存储或者本地存储路径，/user/mindx-dl/ranktable/为前缀路径，必须和[Ascend Operator挂载的Ranktable根目录](zh-cn_topic_0000002479386414.md#li488612012223)保持一致。default.default-test-pytorch为后缀路径，建议改为:namespace.job-name。
        ...
        ```

    -   （可选）如果使用本地存储的挂载方式，需要将YAML中的NFS方式改为hostPath。

        ```
                  volumes:
                  - name: code
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/code/resnet/"
                  - name: data
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/public/dataset/"
                  - name: output
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/output/"
                  - name: ascend-driver
                    hostPath:
                      path: /usr/local/Ascend/driver
                  - name: dshm
                    emptyDir:
                      medium: Memory
                  - name: localtime
                    hostPath:
                      path: /etc/localtime
        ```

**（可选）修改训练脚本<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section1048332432310"></a>**

如果开启了graceExit策略，需修改任务YAML，配置故障恢复策略为“dump“，确保TaskD和ClusterD可以正常使用。

```
...  
  labels:  
     ... 
     subHealthyStrategy: "graceExit"
...
   annotations:  
     ...  
     recover-strategy: "dump"
...
```



## 通过命令行使用<a name="ZH-CN_TOPIC_0000002479386546"></a>

### （可选）配置组件<a name="ZH-CN_TOPIC_0000002511346449"></a>

如果用户在安装Ascend Device Plugin和NodeD时，已经配置了断点续训相关功能，则可以跳过本章节；若没有配置，则需要对[Ascend Device Plugin](#section14208511958)和[NodeD](#section162092113510)进行相关配置。

**配置Ascend Device Plugin<a name="section14208511958"></a>**

只支持以容器化方式启动Ascend Device Plugin。

1.  根据所使用的故障处理模式，修改Ascend Device Plugin组件的启动YAML，修改如下所示加粗部分。
    1.  重调度模式

        >[!NOTE] 说明 
        >在重调度模式下，Ascend Device Plugin的异常也会触发故障重调度。

        ```
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin  
                         -useAscendDocker=true 
                         -volcanoType=true                    # 重调度场景下必须使用Volcano
                         -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                         -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...
        ```

    2.  （可选）优雅容错模式：在重调度配置的基础上，新增“-hotReset“字段。

        >[!NOTE] 说明 
        >-   优雅容错功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。
        >-   “-hotReset“字段取值为1对应的功能已经日落。

        ```
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin  
                         -useAscendDocker=true 
                         -volcanoType=true                    # 重调度场景下必须使用Volcano
                         -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                         -hotReset=1 # 开启优雅容错模式，系统会尝试自动复位故障芯片
                         -listWatchPeriod=5                   # 健康状态检查周期，范围[3,1800]；单位为秒
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...
        ```

2.  在K8s管理节点执行以下命令，启动Ascend Device Plugin。

    ```
    kubectl apply -f device-plugin-xxx-v{version}.yaml
    ```

    如在Atlas 训练系列产品环境下启动该组件，示例如下。

    ```
    kubectl apply -f device-plugin-volcano-v{version}.yaml
    ```

**配置NodeD<a name="section162092113510"></a>**

配置节点状态发送间隔时间。用户可以通过手动修改NodeD的启动YAML，配置上报节点状态的间隔时间。

1.  进入组件解压目录，执行以下命令，打开NodeD组件的启动YAML文件。

    ```
    vi noded-v{version}.yaml
    ```

2.  在YAML文件的“args”行修改“-**reportInterval**”参数，如下所示：

    ```
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 -reportInterval=5" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: true
              volumeMounts:
                - name: log-noded
    ...
    ```

    >[!NOTE] 说明 
    >-   K8s[默认40秒未收到节点响应时](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)将该节点置为NotReady。
    >-   当K8s API Server请求压力变大时，可根据实际情况增大间隔时间，以减轻API Server压力。


### 制作镜像<a name="ZH-CN_TOPIC_0000002511426469"></a>

#### 制作MindSpeed-LLM训练镜像（PyTorch框架）<a name="ZH-CN_TOPIC_0000002479386504"></a>

[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)作为昇腾大模型训练框架，旨在为昇腾芯片提供端到端的大语言模型训练方案，包含分布式预训练、分布式指令微调、分布式偏好对齐以及对应的开发工具链。[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md)包括了仓库拉取、环境搭建与大模型训练等章节，制作MindSpeed-LLM训练框架镜像可以结合本章节和[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/1.0.0/docs/USER_GUIDE.md)。

断点续训可以基于**基础训练镜像**制作，**基础训练镜像**的制作可参考[使用Dockerfile构建容器镜像（PyTorch）](../common_operations.md#使用dockerfile构建容器镜像pytorch)章节进行操作。

本章节结合基础训练镜像的制作步骤，展示基于Ubuntu 20.04来构建训练镜像。

>[!NOTE] 说明 
>以下示例使用MindSpeed-LLM  2.3.0版本。

**准备软件包<a name="zh-cn_topic_0000002039339945_section18254161612586"></a>**

请按照[表1](#zh-cn_topic_0000002039339945_table1172542119019)所示，获取对应操作系统的软件包，并准备镜像所需的Dockerfile文件与脚本文件。软件包名称中{version}表示版本号、{arch}表示架构、{chip_type}表示芯片类型。

**表 1**  准备软件包

<a name="zh-cn_topic_0000002039339945_table1172542119019"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339945_row157251121508"><th class="cellrowborder" valign="top" width="24.55%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002039339945_p1441653254"><a name="zh-cn_topic_0000002039339945_p1441653254"></a><a name="zh-cn_topic_0000002039339945_p1441653254"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="25.45%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002039339945_p2052053751"><a name="zh-cn_topic_0000002039339945_p2052053751"></a><a name="zh-cn_topic_0000002039339945_p2052053751"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002039339945_p657531455"><a name="zh-cn_topic_0000002039339945_p657531455"></a><a name="zh-cn_topic_0000002039339945_p657531455"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002039339945_p1859531759"><a name="zh-cn_topic_0000002039339945_p1859531759"></a><a name="zh-cn_topic_0000002039339945_p1859531759"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339945_row16726192116014"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1754534515"><a name="zh-cn_topic_0000002039339945_p1754534515"></a><a name="zh-cn_topic_0000002039339945_p1754534515"></a>taskd-<em id="i2511021165615"><a name="i2511021165615"></a><a name="i2511021165615"></a>{version}</em>-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p4321152352612"><a name="p4321152352612"></a><a name="p4321152352612"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p205155310512"><a name="zh-cn_topic_0000002039339945_p205155310512"></a><a name="zh-cn_topic_0000002039339945_p205155310512"></a>集群调度组件断点续训whl包。</p>
<div class="note" id="zh-cn_topic_0000002039339945_note494818501423"><a name="zh-cn_topic_0000002039339945_note494818501423"></a><a name="zh-cn_topic_0000002039339945_note494818501423"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p159489506423"><a name="zh-cn_topic_0000002039339945_p159489506423"></a><a name="zh-cn_topic_0000002039339945_p159489506423"></a>安装<span id="ph1670711477256"><a name="ph1670711477256"></a><a name="ph1670711477256"></a>TaskD</span>组件前需确保<span id="zh-cn_topic_0000002039339945_ph998914174412"><a name="zh-cn_topic_0000002039339945_ph998914174412"></a><a name="zh-cn_topic_0000002039339945_ph998914174412"></a>PyTorch</span>框架已正确安装，当前支持的<span id="zh-cn_topic_0000002039339945_ph2908133144419"><a name="zh-cn_topic_0000002039339945_ph2908133144419"></a><a name="zh-cn_topic_0000002039339945_ph2908133144419"></a>PyTorch</span>版本为：2.1.0、2.3.0、2.4.0、2.5.0、2.6.0、2.7.1。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p19595310517"><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1386820525510"><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p12512532515"><a name="zh-cn_topic_0000002039339945_p12512532515"></a><a name="zh-cn_topic_0000002039339945_p12512532515"></a>用户通过获取链接得到的是<span id="ph480901420289"><a name="ph480901420289"></a><a name="ph480901420289"></a>TaskD</span>压缩包Ascend-mindxdl-taskd_<em id="i112838253389"><a name="i112838253389"></a><a name="i112838253389"></a>{version}</em>_linux-<em id="i1328312515383"><a name="i1328312515383"></a><a name="i1328312515383"></a>{arch}</em>.zip，需要通过解压后，获得相应的whl软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row572619211108"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1863531756"><a name="zh-cn_topic_0000002039339945_p1863531756"></a><a name="zh-cn_topic_0000002039339945_p1863531756"></a>mindio_ttp-<em id="zh-cn_topic_0000002039339945_i15340181201416"><a name="zh-cn_topic_0000002039339945_i15340181201416"></a><a name="zh-cn_topic_0000002039339945_i15340181201416"></a>{version}</em>-py3-none-linux_<em id="zh-cn_topic_0000002039339945_i19614531957"><a name="zh-cn_topic_0000002039339945_i19614531957"></a><a name="zh-cn_topic_0000002039339945_i19614531957"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p96553958"><a name="zh-cn_topic_0000002039339945_p96553958"></a><a name="zh-cn_topic_0000002039339945_p96553958"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p66253359"><a name="zh-cn_topic_0000002039339945_p66253359"></a><a name="zh-cn_topic_0000002039339945_p66253359"></a><span id="zh-cn_topic_0000002039339945_ph845710020145"><a name="zh-cn_topic_0000002039339945_ph845710020145"></a><a name="zh-cn_topic_0000002039339945_ph845710020145"></a>MindIO TFT</span>安装包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1862053354"><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row672652117018"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1468532510"><a name="zh-cn_topic_0000002039339945_p1468532510"></a><a name="zh-cn_topic_0000002039339945_p1468532510"></a>apex-0.1+ascend-cp3x-cp3x-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1869531256"><a name="zh-cn_topic_0000002039339945_p1869531256"></a><a name="zh-cn_topic_0000002039339945_p1869531256"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p156353454"><a name="zh-cn_topic_0000002039339945_p156353454"></a><a name="zh-cn_topic_0000002039339945_p156353454"></a>MindSpeed-LLM依赖</p>
<p id="zh-cn_topic_0000002039339945_p1861353651"><a name="zh-cn_topic_0000002039339945_p1861353651"></a><a name="zh-cn_topic_0000002039339945_p1861353651"></a></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p166105316515"><a name="zh-cn_topic_0000002039339945_p166105316515"></a><a name="zh-cn_topic_0000002039339945_p166105316515"></a>混合精度训练是在训练时混合使用单精度（float32）与半精度(float16)数据类型，将两者结合在一起，并使用相同的超参数实现了与float32几乎相同的精度。</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a>x表示10或11，当前可支持<span id="zh-cn_topic_0000002039339945_ph17691173812277"><a name="zh-cn_topic_0000002039339945_ph17691173812277"></a><a name="zh-cn_topic_0000002039339945_ph17691173812277"></a>Python</span> 3.10和<span id="zh-cn_topic_0000002039339945_ph1369143872716"><a name="zh-cn_topic_0000002039339945_ph1369143872716"></a><a name="zh-cn_topic_0000002039339945_ph1369143872716"></a>Python</span>3.11。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a>请参见<span id="zh-cn_topic_0000002039339945_ph156792413596"><a name="zh-cn_topic_0000002039339945_ph156792413596"></a><a name="zh-cn_topic_0000002039339945_ph156792413596"></a>《Ascend Extension for PyTorch 软件安装指南》中的“安装APEX模块”章节</span>，根据实际情况编译APEX软件包。</p>
<p id="zh-cn_topic_0000002039339945_p1761531257"><a name="zh-cn_topic_0000002039339945_p1761531257"></a><a name="zh-cn_topic_0000002039339945_p1761531257"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row197268213011"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1361953254"><a name="zh-cn_topic_0000002039339945_p1361953254"></a><a name="zh-cn_topic_0000002039339945_p1361953254"></a>torch_npu-2.7.1.<em id="zh-cn_topic_0000002039339945_i16204112111321"><a name="zh-cn_topic_0000002039339945_i16204112111321"></a><a name="zh-cn_topic_0000002039339945_i16204112111321"></a>{version}</em>-cp3x-cp3x-manylinux_2_28_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p56185314510"><a name="zh-cn_topic_0000002039339945_p56185314510"></a><a name="zh-cn_topic_0000002039339945_p56185314510"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p186653757"><a name="zh-cn_topic_0000002039339945_p186653757"></a><a name="zh-cn_topic_0000002039339945_p186653757"></a>MindSpeed-LLM依赖</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p15720535517"><a name="zh-cn_topic_0000002039339945_p15720535517"></a><a name="zh-cn_topic_0000002039339945_p15720535517"></a>Ascend Extension for PyTorch插件是基于昇腾的深度学习适配框架，使昇腾NPU可以支持PyTorch框架，为PyTorch框架的使用者提供昇腾AI处理器的超强算力。</p>
<p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p849562217019"></a>x表示10或11，当前可支持<span id="zh-cn_topic_0000002039339945_ph1572718405287"><a name="zh-cn_topic_0000002039339945_ph1572718405287"></a><a name="zh-cn_topic_0000002039339945_ph1572718405287"></a>Python</span> 3.10和<span id="zh-cn_topic_0000002039339945_ph1727184012819"><a name="zh-cn_topic_0000002039339945_ph1727184012819"></a><a name="zh-cn_topic_0000002039339945_ph1727184012819"></a>Python</span>3.11。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p10718533510"><a name="zh-cn_topic_0000002039339945_p10718533510"></a><a name="zh-cn_topic_0000002039339945_p10718533510"></a><a href="https://www.hiascend.com/document/detail/zh/Pytorch/720/configandinstg/instg/insg_0004.html" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1165115165020"><a name="zh-cn_topic_0000002039339945_note1165115165020"></a><a name="zh-cn_topic_0000002039339945_note1165115165020"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p167047813263"><a name="zh-cn_topic_0000002039339945_p167047813263"></a><a name="zh-cn_topic_0000002039339945_p167047813263"></a>如果使用MindSpeed-LLM代码仓上的<span id="zh-cn_topic_0000002039339945_ph1987542822613"><a name="zh-cn_topic_0000002039339945_ph1987542822613"></a><a name="zh-cn_topic_0000002039339945_ph1987542822613"></a>PyTorch</span>模型，需要使用<span id="zh-cn_topic_0000002039339945_ph1412723132619"><a name="zh-cn_topic_0000002039339945_ph1412723132619"></a><a name="zh-cn_topic_0000002039339945_ph1412723132619"></a>Ascend Extension for PyTorch</span> 2.6.0及以上版本。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row1412215399516"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><a name="zh-cn_topic_0000002039339945_ul104867135415"></a><a name="zh-cn_topic_0000002039339945_ul104867135415"></a><ul id="zh-cn_topic_0000002039339945_ul104867135415"><li><span id="zh-cn_topic_0000002039339945_ph0853174512272"><a name="zh-cn_topic_0000002039339945_ph0853174512272"></a><a name="zh-cn_topic_0000002039339945_ph0853174512272"></a>x86_64</span>架构：torch-2.7.1+cpu.cxx11.abi-cp3x-cp3x-linux_x86_64.whl</li><li><span id="zh-cn_topic_0000002039339945_ph7852164518272"><a name="zh-cn_topic_0000002039339945_ph7852164518272"></a><a name="zh-cn_topic_0000002039339945_ph7852164518272"></a>ARM</span>架构：torch-2.7.1+cpu-cp3x-cp3x-manylinux_2_28_aarch64.whl</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p871953357"><a name="zh-cn_topic_0000002039339945_p871953357"></a><a name="zh-cn_topic_0000002039339945_p871953357"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p17715531453"><a name="zh-cn_topic_0000002039339945_p17715531453"></a><a name="zh-cn_topic_0000002039339945_p17715531453"></a>MindSpeed-LLM依赖</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p11461347141013"></a>官方<span id="zh-cn_topic_0000002039339945_ph19355165113512"><a name="zh-cn_topic_0000002039339945_ph19355165113512"></a><a name="zh-cn_topic_0000002039339945_ph19355165113512"></a>PyTorch</span>包。x表示10或11，当前可支持<span id="zh-cn_topic_0000002039339945_ph184492055182714"><a name="zh-cn_topic_0000002039339945_ph184492055182714"></a><a name="zh-cn_topic_0000002039339945_ph184492055182714"></a>Python</span>3.10和<span id="zh-cn_topic_0000002039339945_ph94491255112719"><a name="zh-cn_topic_0000002039339945_ph94491255112719"></a><a name="zh-cn_topic_0000002039339945_ph94491255112719"></a>Python</span>3.11。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p99745421447"><a name="zh-cn_topic_0000002039339945_p99745421447"></a><a name="zh-cn_topic_0000002039339945_p99745421447"></a><a href="https://download.pytorch.org/whl/torch/" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<p id="zh-cn_topic_0000002039339945_p483943610920"><a name="zh-cn_topic_0000002039339945_p483943610920"></a><a name="zh-cn_topic_0000002039339945_p483943610920"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row151232039750"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p4714531658"><a name="zh-cn_topic_0000002039339945_p4714531658"></a><a name="zh-cn_topic_0000002039339945_p4714531658"></a>Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1774531250"><a name="zh-cn_topic_0000002039339945_p1774531250"></a><a name="zh-cn_topic_0000002039339945_p1774531250"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5720531514"><a name="zh-cn_topic_0000002039339945_p5720531514"></a><a name="zh-cn_topic_0000002039339945_p5720531514"></a>CANN算子包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p87153755"><a name="zh-cn_topic_0000002039339945_p87153755"></a><a name="zh-cn_topic_0000002039339945_p87153755"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note13775154104217"><a name="zh-cn_topic_0000002039339945_note13775154104217"></a><a name="zh-cn_topic_0000002039339945_note13775154104217"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p2075812144313"><a name="zh-cn_topic_0000002039339945_p2075812144313"></a><a name="zh-cn_topic_0000002039339945_p2075812144313"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="row1173819266428"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="p7738142674217"><a name="p7738142674217"></a><a name="p7738142674217"></a>Ascend-cann-toolkit_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="p12738626184214"><a name="p12738626184214"></a><a name="p12738626184214"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p13738226114218"><a name="p13738226114218"></a><a name="p13738226114218"></a>CANN Toolkit开发套件包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p19271154916428"><a name="p19271154916428"></a><a name="p19271154916428"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="note3272104913427"><a name="note3272104913427"></a><a name="note3272104913427"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p2272174920421"><a name="p2272174920421"></a><a name="p2272174920421"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row121231639952"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1048218391768"><a name="zh-cn_topic_0000002039339945_p1048218391768"></a><a name="zh-cn_topic_0000002039339945_p1048218391768"></a>MindSpeed</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p3482939367"><a name="zh-cn_topic_0000002039339945_p3482939367"></a><a name="zh-cn_topic_0000002039339945_p3482939367"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p048215393610"><a name="zh-cn_topic_0000002039339945_p048215393610"></a><a name="zh-cn_topic_0000002039339945_p048215393610"></a>MindSpeed是针对昇腾设备的大模型加速库。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p7482193913619"><a name="zh-cn_topic_0000002039339945_p7482193913619"></a><a name="zh-cn_topic_0000002039339945_p7482193913619"></a>git clone https://gitcode.com/Ascend/MindSpeed.git</p>
<p id="zh-cn_topic_0000002039339945_p9482139663"><a name="zh-cn_topic_0000002039339945_p9482139663"></a><a name="zh-cn_topic_0000002039339945_p9482139663"></a>cd MindSpeed</p>
<p id="zh-cn_topic_0000002039339945_p1948213912618"><a name="zh-cn_topic_0000002039339945_p1948213912618"></a><a name="zh-cn_topic_0000002039339945_p1948213912618"></a>git checkout 2.3.0_core_r0.12.1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row144125121466"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p848215391768"><a name="zh-cn_topic_0000002039339945_p848215391768"></a><a name="zh-cn_topic_0000002039339945_p848215391768"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1548319391465"><a name="zh-cn_topic_0000002039339945_p1548319391465"></a><a name="zh-cn_topic_0000002039339945_p1548319391465"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p16483239968"><a name="zh-cn_topic_0000002039339945_p16483239968"></a><a name="zh-cn_topic_0000002039339945_p16483239968"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p14483939561"><a name="zh-cn_topic_0000002039339945_p14483939561"></a><a name="zh-cn_topic_0000002039339945_p14483939561"></a>驱动版本信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p748314394617"><a name="zh-cn_topic_0000002039339945_p748314394617"></a><a name="zh-cn_topic_0000002039339945_p748314394617"></a>从host拷贝“/usr/local/Ascend/driver/version.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row17301171913614"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p148310396616"><a name="zh-cn_topic_0000002039339945_p148310396616"></a><a name="zh-cn_topic_0000002039339945_p148310396616"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p348313399610"><a name="zh-cn_topic_0000002039339945_p348313399610"></a><a name="zh-cn_topic_0000002039339945_p348313399610"></a>是</p>
<p id="zh-cn_topic_0000002039339945_p1348313391961"><a name="zh-cn_topic_0000002039339945_p1348313391961"></a><a name="zh-cn_topic_0000002039339945_p1348313391961"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p5483239564"><a name="zh-cn_topic_0000002039339945_p5483239564"></a><a name="zh-cn_topic_0000002039339945_p5483239564"></a>驱动安装信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p2483339861"><a name="zh-cn_topic_0000002039339945_p2483339861"></a><a name="zh-cn_topic_0000002039339945_p2483339861"></a>从host拷贝“/etc/ascend_install.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row93022191368"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15485173911615"><a name="zh-cn_topic_0000002039339945_p15485173911615"></a><a name="zh-cn_topic_0000002039339945_p15485173911615"></a>Dllogger代码仓</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p248514391563"><a name="zh-cn_topic_0000002039339945_p248514391563"></a><a name="zh-cn_topic_0000002039339945_p248514391563"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164850391565"><a name="zh-cn_topic_0000002039339945_p164850391565"></a><a name="zh-cn_topic_0000002039339945_p164850391565"></a>PyTorch日志工具。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1548619391164"><a name="zh-cn_topic_0000002039339945_p1548619391164"></a><a name="zh-cn_topic_0000002039339945_p1548619391164"></a>git clone https://github.com/NVIDIA/dllogger.git</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row33025197610"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p0486133919618"><a name="zh-cn_topic_0000002039339945_p0486133919618"></a><a name="zh-cn_topic_0000002039339945_p0486133919618"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1148619393610"><a name="zh-cn_topic_0000002039339945_p1148619393610"></a><a name="zh-cn_topic_0000002039339945_p1148619393610"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p164861439461"><a name="zh-cn_topic_0000002039339945_p164861439461"></a><a name="zh-cn_topic_0000002039339945_p164861439461"></a>用于安装pip模块。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p34865396611"><a name="zh-cn_topic_0000002039339945_p34865396611"></a><a name="zh-cn_topic_0000002039339945_p34865396611"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row03021191165"><td class="cellrowborder" valign="top" width="24.55%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p15486163919616"><a name="zh-cn_topic_0000002039339945_p15486163919616"></a><a name="zh-cn_topic_0000002039339945_p15486163919616"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="25.45%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1448693915619"><a name="zh-cn_topic_0000002039339945_p1448693915619"></a><a name="zh-cn_topic_0000002039339945_p1448693915619"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p74862391165"><a name="zh-cn_topic_0000002039339945_p74862391165"></a><a name="zh-cn_topic_0000002039339945_p74862391165"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p9571628104613"><a name="zh-cn_topic_0000002039339945_p9571628104613"></a><a name="zh-cn_topic_0000002039339945_p9571628104613"></a>-</p>
</td>
</tr>
</tbody>
</table>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE] 说明 
>本章节以单台Atlas 800T A2 训练服务器、Ubuntu 20.04 Arm、配套Python  3.10为例来介绍训练镜像的制作，使用过程中需根据实际情况修改相关步骤。

**操作步骤<a name="zh-cn_topic_0000002039339945_section20489630477"></a>**

1.  参照[表1](#zh-cn_topic_0000002039339945_table1172542119019)，在宿主机上完成软件包的准备工作。
2.  编写如下Dockerfile。

    ```
    FROM ubuntu:20.04 
    WORKDIR /root 
    COPY . . 
      
    ARG PYTORCH_WHL=torch-2.7.1+cpu-cp310-cp310-manylinux_2_28_aarch64.whl 
    ARG PYTORCH_NPU_WHL=torch_npu-2.7.1.{version}-cp310-cp310-manylinux_2_28_aarch64.whl 
    ARG APEX_WHL=apex-0.1+ascend-cp310-cp310-linux_aarch64.whl 
    ARG HOST_ASCEND_BASE=/usr/local/Ascend 
    ARG TOOLKIT_PATH=/usr/local/Ascend/cann 
    # 示例使用的CANN版本为8.5.0,使用过程中请根据实际情况修改
    ARG TOOLKIT=Ascend-cann-toolkit_8.5.0_linux-aarch64.run    
    ARG OPS=Ascend-cann-910b-ops_8.5.0_linux-aarch64.run 
    ARG TASKD_WHL=taskd-7.3.0-py3-none-linux_aarch64.whl   
    ARG MINDIO_TTP_WHL=mindio_ttp-1.0.0-py3-none-linux_aarch64.whl 
    ARG MINDSPEED=MindSpeed 
    ARG DLLOGGER=dllogger 
      
    RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf 
      
    RUN echo "deb http://repo.huaweicloud.com/ubuntu-ports/ focal main restricted universe multiverse\n\ 
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-updates main restricted universe multiverse\n\ 
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-backports main restricted universe multiverse\n\ 
    deb http://ports.ubuntu.com/ubuntu-ports/ focal-security main restricted universe multiverse" > /etc/apt/sources.list 
      
    ARG DEBIAN_FRONTEND=noninteractive 
      
    # 系统包 
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
        software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && \
        apt update && \
        apt autoremove -y python python3 && \
        apt install -y python3.10 python3.10-dev
    # 建立Python软链
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
    # 系统
    RUN umask 0022 && apt update && \
            apt-get install -y --no-install-recommends \
            gcc g++ make cmake vim \
            zlib1g zlib1g-dev \
            openssl libsqlite3-dev libssl-dev \
            libffi-dev unzip pciutils \
            net-tools libblas-dev \
            gfortran libblas3 libopenblas-dev \
            curl unzip liblapack3 liblapack-dev \
            libhdf5-dev libxml2 patch
    # 时区
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime
    # 配置pip源
    RUN mkdir -p ~/.pip \
    && echo '[global] \n\
    index-url=https://mirrors.huaweicloud.com/repository/pypi/simple\n\
    trusted-host=mirrors.huaweicloud.com' >> ~/.pip/pip.conf
    # pip3.10
    RUN cd /tmp && \
        apt-get download python3-distutils && \
        dpkg-deb -x python3-distutils_*.deb / && \
        rm python3-distutils_*.deb && \
        cd - && \
        python get-pip.py && \
        rm get-pip.py
    RUN umask 0022 && \
        pip install sympy==1.4 && \
        pip install cffi && \
        pip install pathlib2 && \
        pip install grpcio && \
        pip install grpcio-tools && \
        pip install torchvision==0.22.1 && \
        pip install transformers==4.51.0 && \
        pip install absl-py && \
        pip install datasets && \
        pip install tokenizers==0.20.1 && \
        pip install pyOpenSSL
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
    # 安装torch、torch_npu、apex包
    RUN umask 0022 && pip install $PYTORCH_WHL && \
        pip install $PYTORCH_NPU_WHL && \
        pip install $APEX_WHL
      
    # Ascend包 
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录 
    RUN umask 0022 &&  \ 
        cp ascend_install.info /etc/ && \ 
        mkdir -p /usr/local/Ascend/driver/ && \ 
        cp version.info /usr/local/Ascend/driver/ && \ 
        chmod +x $TOOLKIT && \ 
        chmod +x $OPS 
      
    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet 
    RUN echo "source /usr/local/Ascend/cann/set_env.sh" >> ~/.bashrc 
    RUN umask 0022 && ./$OPS --install --quiet 
      
    # 只为了安装toolkit包，所以需要清理，容器启动时通过Ascend Docker Runtime挂载进来 
    RUN rm -f version.info && rm -f ascend_install.info \ 
        rm -rf /usr/local/Ascend/driver/ 
      
    RUN umask 0022 && cd $MINDSPEED && \ 
        pip install -r requirements.txt && \ 
        pip install -e . && \ 
        echo "export PYTHONPATH=/root/MindSpeed:\$PYTHONPATH" >> ~/.bashrc 
      
    RUN umask 0022 && cd $DLLOGGER && \ 
        python setup.py build && \ 
        python setup.py install 
      
    # 导入环境变量 
    ENV HCCL_WHITELIST_DISABLE=1 
      
    # 创建/lib64/ld-linux-aarch64.so.1 
    RUN umask 0022 && \ 
        if [ ! -d "/lib64" ]; \ 
        then \ 
            mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \ 
        fi 
      
    # MindCluster断点续训适配脚本。
    RUN umask 0022 && \ 
        pip install $TASKD_WHL && \ 
        pip install $MINDIO_TTP_WHL 
      
      
    # 可选，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令。
    RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py  
    
    # 增加安装任务调度依赖库 
    RUN pip install apscheduler 
      
    RUN rm -rf tmp && \ 
        rm -f $PYTORCH_WHL && \ 
        rm -f $PYTORCH_NPU_WHL && \ 
        rm -f $APEX_WHL && \ 
        rm -f $TOOLKIT && \ 
        rm -f $OPS && \ 
        rm -f $TASKD_WHL && \ 
        rm -f $MINDIO_TTP_WHL && \ 
        rm -rf $DLLOGGER && \ 
        rm -rf Dockerfile 
    ## 最后打包成镜像mindspeed-dl:v1
    ```

    >[!NOTE] 说明 
    >Python 3.10若无法通过PPA直接安装成功，或者deadsnakes PPA不提供Python 3.10版本的镜像源，则可下载源码手动编译安装。

3.  构建镜像。执行以下命令生成镜像。为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。**注意不要遗漏命令结尾的**“.“。

    ```
    docker build -t mindspeed-dl:v1 .
    ```


#### 制作MindFormers训练镜像（MindSpore框架）<a name="ZH-CN_TOPIC_0000002511426451"></a>

[MindSpore Transformers套件](https://gitee.com/mindspore/mindformers)（以下简称MindFormers）的目标是构建一个大模型训练、微调、评估、推理、部署的全流程开发套件，提供业内主流的Transformer类预训练模型和SOTA下游任务应用，涵盖丰富的并行特性。期望帮助用户轻松地实现大模型训练和创新研发。

[MindSpore Transformers文档](https://www.mindspore.cn/mindformers/docs/zh-CN/r1.3.0/start/overview.html)的快速入门包括了安装与快速启动章节，可以在镜像制作时参考。

训练镜像可以基于**基础训练镜像，**结合MindFormers**文档自行制作，基础训练镜像**的制作可参考[使用Dockerfile构建容器镜像（MindSpore）](../common_operations.md#使用dockerfile构建容器镜像mindspore)章节进行操作。

本章节结合基础训练镜像的制作步骤，展示基于Ubuntu 20.04来构建训练镜像。

**准备软件包<a name="zh-cn_topic_0000002003180012_section181941327124212"></a>**

请按照[表1](#zh-cn_topic_0000002003180012_table223643812168)所示，获取对应操作系统的软件包，并准备镜像所需的Dockerfile文件与脚本文件。软件包名称中{version}表示版本号、{arch}表示架构、{chip_type}表示芯片类型。

**表 1**  准备软件包

<a name="zh-cn_topic_0000002003180012_table223643812168"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003180012_row6236938171619"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002003180012_p3390131317171"><a name="zh-cn_topic_0000002003180012_p3390131317171"></a><a name="zh-cn_topic_0000002003180012_p3390131317171"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002003180012_p173901213151712"><a name="zh-cn_topic_0000002003180012_p173901213151712"></a><a name="zh-cn_topic_0000002003180012_p173901213151712"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002003180012_p239018134178"><a name="zh-cn_topic_0000002003180012_p239018134178"></a><a name="zh-cn_topic_0000002003180012_p239018134178"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002003180012_p1539051321714"><a name="zh-cn_topic_0000002003180012_p1539051321714"></a><a name="zh-cn_topic_0000002003180012_p1539051321714"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003180012_row13237173817161"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p6390191319171"><a name="zh-cn_topic_0000002003180012_p6390191319171"></a><a name="zh-cn_topic_0000002003180012_p6390191319171"></a>MindFormers代码仓</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p13390113131712"><a name="zh-cn_topic_0000002003180012_p13390113131712"></a><a name="zh-cn_topic_0000002003180012_p13390113131712"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p153901913131711"><a name="zh-cn_topic_0000002003180012_p153901913131711"></a><a name="zh-cn_topic_0000002003180012_p153901913131711"></a>构建一个大模型训练、微调、评估、推理、部署的全流程开发套件，提供业内主流的Transformer类预训练模型和SOTA下游任务应用，涵盖丰富的并行特性<span id="ph19351335211"><a name="ph19351335211"></a><a name="ph19351335211"></a>。</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p3390131316172"><a name="zh-cn_topic_0000002003180012_p3390131316172"></a><a name="zh-cn_topic_0000002003180012_p3390131316172"></a>git clone https://gitee.com/mindspore/mindformers.git</p>
<p id="zh-cn_topic_0000002003180012_p5390101317175"><a name="zh-cn_topic_0000002003180012_p5390101317175"></a><a name="zh-cn_topic_0000002003180012_p5390101317175"></a>cd mindformers</p>
<p id="zh-cn_topic_0000002003180012_p9390151318171"><a name="zh-cn_topic_0000002003180012_p9390151318171"></a><a name="zh-cn_topic_0000002003180012_p9390151318171"></a>git checkout f06a946af29c8c7e002a6c49458f513d47b642e5</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row14237113817167"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p133901013201717"><a name="zh-cn_topic_0000002003180012_p133901013201717"></a><a name="zh-cn_topic_0000002003180012_p133901013201717"></a>requirements.txt文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p10390813171719"><a name="zh-cn_topic_0000002003180012_p10390813171719"></a><a name="zh-cn_topic_0000002003180012_p10390813171719"></a>否</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p439011371714"><a name="zh-cn_topic_0000002003180012_p439011371714"></a><a name="zh-cn_topic_0000002003180012_p439011371714"></a>由于通过pip安装MindSpore时，可能出现依赖的组件安装报错，故可以先安装依赖。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p6390121315177"><a name="zh-cn_topic_0000002003180012_p6390121315177"></a><a name="zh-cn_topic_0000002003180012_p6390121315177"></a>wget https://gitee.com/mindspore/mindspore/raw/r2.4.1/requirements.txt</p>
<div class="note" id="zh-cn_topic_0000002003180012_note14449193224617"><a name="zh-cn_topic_0000002003180012_note14449193224617"></a><a name="zh-cn_topic_0000002003180012_note14449193224617"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003180012_p15449133274617"><a name="zh-cn_topic_0000002003180012_p15449133274617"></a><a name="zh-cn_topic_0000002003180012_p15449133274617"></a>MindSpore软件包与<span id="zh-cn_topic_0000002003180012_ph327965117217"><a name="zh-cn_topic_0000002003180012_ph327965117217"></a><a name="zh-cn_topic_0000002003180012_ph327965117217"></a>Atlas 训练系列产品</span>需配套使用，请参见MindSpore<a href="https://www.mindspore.cn/install" target="_blank" rel="noopener noreferrer">安装指南</a>查看对应关系。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row2023743821619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p123901513131714"><a name="zh-cn_topic_0000002003180012_p123901513131714"></a><a name="zh-cn_topic_0000002003180012_p123901513131714"></a>mindspore-<em id="zh-cn_topic_0000002003180012_i42701940155017"><a name="zh-cn_topic_0000002003180012_i42701940155017"></a><a name="zh-cn_topic_0000002003180012_i42701940155017"></a>{version}</em>-cp3x-cp3x-linux_aarch64.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p73901313101718"><a name="zh-cn_topic_0000002003180012_p73901313101718"></a><a name="zh-cn_topic_0000002003180012_p73901313101718"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p1839181315178"><a name="zh-cn_topic_0000002003180012_p1839181315178"></a><a name="zh-cn_topic_0000002003180012_p1839181315178"></a>MindSpore whl包<span id="ph441575419329"><a name="ph441575419329"></a><a name="ph441575419329"></a>。</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p6391181310177"><a name="zh-cn_topic_0000002003180012_p6391181310177"></a><a name="zh-cn_topic_0000002003180012_p6391181310177"></a><a href="https://www.mindspore.cn/install/" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row32371838111619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p13917139175"><a name="zh-cn_topic_0000002003180012_p13917139175"></a><a name="zh-cn_topic_0000002003180012_p13917139175"></a>mindio_ttp-<em id="zh-cn_topic_0000002003180012_i14277191551111"><a name="zh-cn_topic_0000002003180012_i14277191551111"></a><a name="zh-cn_topic_0000002003180012_i14277191551111"></a>{version}</em>-py3-none-linux_<em id="zh-cn_topic_0000002003180012_i16391113201710"><a name="zh-cn_topic_0000002003180012_i16391113201710"></a><a name="zh-cn_topic_0000002003180012_i16391113201710"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p183915136176"><a name="zh-cn_topic_0000002003180012_p183915136176"></a><a name="zh-cn_topic_0000002003180012_p183915136176"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p13906311171017"><a name="zh-cn_topic_0000002003180012_p13906311171017"></a><a name="zh-cn_topic_0000002003180012_p13906311171017"></a><span id="zh-cn_topic_0000002003180012_ph845710020145"><a name="zh-cn_topic_0000002003180012_ph845710020145"></a><a name="zh-cn_topic_0000002003180012_ph845710020145"></a>MindIO TFT</span>安装包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p11392111316172"><a name="zh-cn_topic_0000002003180012_p11392111316172"></a><a name="zh-cn_topic_0000002003180012_p11392111316172"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row1423815380168"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p0915027134813"><a name="zh-cn_topic_0000002003180012_p0915027134813"></a><a name="zh-cn_topic_0000002003180012_p0915027134813"></a>Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p1139314132170"><a name="zh-cn_topic_0000002003180012_p1139314132170"></a><a name="zh-cn_topic_0000002003180012_p1139314132170"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p193931713181710"><a name="zh-cn_topic_0000002003180012_p193931713181710"></a><a name="zh-cn_topic_0000002003180012_p193931713181710"></a>CANN算子包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p139312131171"><a name="zh-cn_topic_0000002003180012_p139312131171"></a><a name="zh-cn_topic_0000002003180012_p139312131171"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002003180012_note13501612171513"><a name="zh-cn_topic_0000002003180012_note13501612171513"></a><a name="zh-cn_topic_0000002003180012_note13501612171513"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003180012_p1519161921516"><a name="zh-cn_topic_0000002003180012_p1519161921516"></a><a name="zh-cn_topic_0000002003180012_p1519161921516"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row8238173810165"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p1439381351714"><a name="zh-cn_topic_0000002003180012_p1439381351714"></a><a name="zh-cn_topic_0000002003180012_p1439381351714"></a>Ascend-cann-toolkit_{version}_linux-{arch}.run</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p1239319131176"><a name="zh-cn_topic_0000002003180012_p1239319131176"></a><a name="zh-cn_topic_0000002003180012_p1239319131176"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p5393121371719"><a name="zh-cn_topic_0000002003180012_p5393121371719"></a><a name="zh-cn_topic_0000002003180012_p5393121371719"></a>CANN Toolkit开发套件包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p239319132175"><a name="zh-cn_topic_0000002003180012_p239319132175"></a><a name="zh-cn_topic_0000002003180012_p239319132175"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002003180012_note1733918441613"><a name="zh-cn_topic_0000002003180012_note1733918441613"></a><a name="zh-cn_topic_0000002003180012_note1733918441613"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003180012_p533924121616"><a name="zh-cn_topic_0000002003180012_p533924121616"></a><a name="zh-cn_topic_0000002003180012_p533924121616"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row4825411181413"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p1882511116142"><a name="zh-cn_topic_0000002003180012_p1882511116142"></a><a name="zh-cn_topic_0000002003180012_p1882511116142"></a>taskd-{version}-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p10825811101413"><a name="zh-cn_topic_0000002003180012_p10825811101413"></a><a name="zh-cn_topic_0000002003180012_p10825811101413"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p4825711171415"><a name="zh-cn_topic_0000002003180012_p4825711171415"></a><a name="zh-cn_topic_0000002003180012_p4825711171415"></a>集群调度组件断点续训whl包。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p18169645192413"><a name="zh-cn_topic_0000002003180012_p18169645192413"></a><a name="zh-cn_topic_0000002003180012_p18169645192413"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002003180012_note079418496154"><a name="zh-cn_topic_0000002003180012_note079418496154"></a><a name="zh-cn_topic_0000002003180012_note079418496154"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002003180012_ul79293962319"></a><a name="zh-cn_topic_0000002003180012_ul79293962319"></a><ul id="zh-cn_topic_0000002003180012_ul79293962319"><li>MindSpore场景下使用优雅容错、Pod级别重调度、进程级别重调度、进程级在线恢复，必须安装该whl包。</li><li>用户通过获取链接得到的是<span id="zh-cn_topic_0000002003180012_ph11742444163719"><a name="zh-cn_topic_0000002003180012_ph11742444163719"></a><a name="zh-cn_topic_0000002003180012_ph11742444163719"></a>TaskD</span>压缩包Ascend-mindxdl-taskd_<em id="i112838253389"><a name="i112838253389"></a><a name="i112838253389"></a>{version}</em>_linux-<em id="i1328312515383"><a name="i1328312515383"></a><a name="i1328312515383"></a>{arch}</em>.zip，需要通过解压后，获得相应的whl软件包。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row15183115071614"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p83941613131719"><a name="zh-cn_topic_0000002003180012_p83941613131719"></a><a name="zh-cn_topic_0000002003180012_p83941613131719"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p4394913121712"><a name="zh-cn_topic_0000002003180012_p4394913121712"></a><a name="zh-cn_topic_0000002003180012_p4394913121712"></a>是</p>
<p id="zh-cn_topic_0000002003180012_p0394413131713"><a name="zh-cn_topic_0000002003180012_p0394413131713"></a><a name="zh-cn_topic_0000002003180012_p0394413131713"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p73942134170"><a name="zh-cn_topic_0000002003180012_p73942134170"></a><a name="zh-cn_topic_0000002003180012_p73942134170"></a>驱动版本信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p239415135179"><a name="zh-cn_topic_0000002003180012_p239415135179"></a><a name="zh-cn_topic_0000002003180012_p239415135179"></a>从host拷贝“/usr/local/Ascend/driver/version.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row218375021618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p5394141315176"><a name="zh-cn_topic_0000002003180012_p5394141315176"></a><a name="zh-cn_topic_0000002003180012_p5394141315176"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p1039401316175"><a name="zh-cn_topic_0000002003180012_p1039401316175"></a><a name="zh-cn_topic_0000002003180012_p1039401316175"></a>是</p>
<p id="zh-cn_topic_0000002003180012_p14394171331717"><a name="zh-cn_topic_0000002003180012_p14394171331717"></a><a name="zh-cn_topic_0000002003180012_p14394171331717"></a>安装CANN的依赖文件</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p03941613151715"><a name="zh-cn_topic_0000002003180012_p03941613151715"></a><a name="zh-cn_topic_0000002003180012_p03941613151715"></a>驱动安装信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p11394141321712"><a name="zh-cn_topic_0000002003180012_p11394141321712"></a><a name="zh-cn_topic_0000002003180012_p11394141321712"></a>从host拷贝“/etc/ascend_install.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row61841150171618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p6394713201715"><a name="zh-cn_topic_0000002003180012_p6394713201715"></a><a name="zh-cn_topic_0000002003180012_p6394713201715"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p14394121310177"><a name="zh-cn_topic_0000002003180012_p14394121310177"></a><a name="zh-cn_topic_0000002003180012_p14394121310177"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p63959135172"><a name="zh-cn_topic_0000002003180012_p63959135172"></a><a name="zh-cn_topic_0000002003180012_p63959135172"></a>用于安装pip模块</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p6395913121711"><a name="zh-cn_topic_0000002003180012_p6395913121711"></a><a name="zh-cn_topic_0000002003180012_p6395913121711"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002003180012_row618410501169"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002003180012_p123971013121710"><a name="zh-cn_topic_0000002003180012_p123971013121710"></a><a name="zh-cn_topic_0000002003180012_p123971013121710"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002003180012_p18397713121711"><a name="zh-cn_topic_0000002003180012_p18397713121711"></a><a name="zh-cn_topic_0000002003180012_p18397713121711"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002003180012_p639719136172"><a name="zh-cn_topic_0000002003180012_p639719136172"></a><a name="zh-cn_topic_0000002003180012_p639719136172"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002003180012_p639714133170"><a name="zh-cn_topic_0000002003180012_p639714133170"></a><a name="zh-cn_topic_0000002003180012_p639714133170"></a>-</p>
</td>
</tr>
</tbody>
</table>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE] 说明 
>本章节以单台Atlas 800T A2 训练服务器、Ubuntu 20.04、配套Python  3.10为例来介绍制作镜像的详细过程，使用过程中需根据实际情况修改相关步骤。

**操作步骤<a name="zh-cn_topic_0000002003180012_section614453171018"></a>**

1.  在宿主机上完成软件包的准备工作。
2.  构建如下的Dockerfile。

    ```
    FROM ubuntu:20.04
     
    WORKDIR /root
     
    COPY . .
     
    ARG HOST_ASCEND_BASE=/usr/local/Ascend
    ARG TOOLKIT_PATH=/usr/local/Ascend/cann
    # 示例使用的CANN版本为8.5.0,使用过程中请根据实际情况修改
    ARG TOOLKIT=Ascend-cann-toolkit_8.5.0_linux-aarch64.run    
    ARG OPS=Ascend-cann-910b-ops_8.5.0_linux-aarch64.run
    ARG MINDIO_TTP_WHL=mindio_ttp-1.0.0-py3-none-linux_aarch64.whl
    ARG MINDFORMERS=mindformers
    ARG MINDSPORE_REQUIREMENTS=requirements.txt
    ARG MINDSPORE_WHL=mindspore-2.5.0-cp310-cp310-linux_aarch64.whl
    ARG TASKD_WHL=taskd-7.0.RC1-py3-none-linux_aarch64.whl    
     
    RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf
     
    RUN echo "deb http://repo.huaweicloud.com/ubuntu-ports/ focal main restricted universe multiverse\n\
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-updates main restricted universe multiverse\n\
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-backports main restricted universe multiverse\n\
    deb http://ports.ubuntu.com/ubuntu-ports/ focal-security main restricted universe multiverse" > /etc/apt/sources.list
     
     
    ARG DEBIAN_FRONTEND=noninteractive
     
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
        software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && \
        apt update && \
        apt autoremove -y python python3 && \
        apt install -y python3.10 python3.10-dev
     
    # 建立Python软链接
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
     
    # 系统包
    RUN umask 0022 && apt update && \
        apt-get install -y --no-install-recommends \
            gcc g++ make cmake vim \
            zlib1g zlib1g-dev \
            openssl libsqlite3-dev libssl-dev \
            libffi-dev unzip pciutils \
            net-tools libblas-dev \
            gfortran libblas3 libopenblas-dev \
            curl unzip liblapack3 liblapack-dev \
            libhdf5-dev libxml2 patch
     
    # 时区
    # RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime
     
    # 配置pip源
    RUN mkdir -p ~/.pip \
    && echo '[global] \n\
    index-url=https://mirrors.huaweicloud.com/repository/pypi/simple\n\
    trusted-host=mirrors.huaweicloud.com' >> ~/.pip/pip.conf
     
    # pip3.10
    RUN cd /tmp && \
        apt-get download python3-distutils && \
        dpkg-deb -x python3-distutils_*.deb / && \
        rm python3-distutils_*.deb && \
        cd - && \
        python get-pip.py && \
    rm get-pip.py
     
    RUN umask 0022 && \
        pip install sympy==1.4 && \
        pip install cffi && \
        pip install pathlib2 && \
        pip install grpcio && \
        pip install grpcio-tools && \
        pip install absl-py && \
        pip install datasets && \
        pip install tokenizers==0.20.1 && \
        pip install pyOpenSSL
     
    # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
     
    # Ascend包
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录
    RUN umask 0022 &&  \
        cp ascend_install.info /etc/ && \
        mkdir -p /usr/local/Ascend/driver/ && \
        cp version.info /usr/local/Ascend/driver/ && \
        chmod +x $TOOLKIT && \
        chmod +x $OPS
     
    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet
    RUN echo "source /usr/local/Ascend/cann/set_env.sh" >> ~/.bashrc
    RUN umask 0022 && ./$OPS --install --quiet
     
    # 只为了安装toolkit包，所以需要清理，容器启动时通过ascend docker挂载进来
    RUN rm -f version.info && \
        rm -rf /usr/local/Ascend/driver/
     
    # 安装mindspore
    RUN umask 0022 && pip uninstall te topi hccl -y && \
             pip install sympy && \
             pip install /usr/local/Ascend/cann/lib64/hccl-*-py3-none-any.whl
    RUN umask 0022 && \
        pip install -r $MINDSPORE_REQUIREMENTS && \
        pip install $MINDSPORE_WHL
     
    # 安装mindformers
    RUN umask 0022 && cd $MINDFORMERS && \
        pip install -r requirements.txt
     
    # MindCluster无损失断点续训适配脚本
    RUN umask 0022 && \
        pip install $MINDIO_TTP_WHL --target=$(pip show mindspore | awk '/Location:/ {print $2}') && \
        pip install $TASKD_WHL
     
     
    # 环境变量
    ENV HCCL_WHITELIST_DISABLE=1
     
    # 创建/lib64/ld-linux-aarch64.so.1
    RUN umask 0022 && \
        if [ ! -d "/lib64" ]; \
        then \
            mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \
        fi
     
    # 增加安装任务调度依赖库
    RUN pip install apscheduler
     
     
    RUN rm -rf tmp && \
        rm -f $TOOLKIT && \
        rm -f $OPS && \
        rm -f $MINDIO_TTP_WHL && \
        rm -f $MINDSPORE_REQUIREMENTS && \
        rm -f $MINDSPORE_WHL
    ## 最后打包成镜像mindformers-dl:v1
    ```

3.  构建镜像。执行以下命令生成镜像**。**为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。**注意不要遗漏命令结尾的**“.“。

    ```
    docker build -t mindformers-dl:v1 .
    ```


#### 制作强化学习后训练镜像（Verl框架）<a name="ZH-CN_TOPIC_0000002511426439"></a>

[Verl](https://verl.readthedocs.io/en/latest/index.html)是一款专为大语言模型（LLM）后训练阶段设计的灵活、高效且具备生产就绪能力的强化学习训练框架。本章节基于Ubuntu 20.04来构建Verl的后训练镜像。

**准备软件包<a name="zh-cn_topic_0000002039339945_section18254161612586"></a>**

请按照[表1](#zh-cn_topic_0000002039339945_table1172542119019)所示，获取对应操作系统的软件包，并准备镜像所需的Dockerfile文件与脚本文件。

**表 1**  准备软件包

<a name="zh-cn_topic_0000002039339945_table1172542119019"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339945_row157251121508"><th class="cellrowborder" valign="top" width="21.150000000000002%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002039339945_p1441653254"><a name="zh-cn_topic_0000002039339945_p1441653254"></a><a name="zh-cn_topic_0000002039339945_p1441653254"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="13.750000000000004%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002039339945_p2052053751"><a name="zh-cn_topic_0000002039339945_p2052053751"></a><a name="zh-cn_topic_0000002039339945_p2052053751"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="33.730000000000004%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002039339945_p657531455"><a name="zh-cn_topic_0000002039339945_p657531455"></a><a name="zh-cn_topic_0000002039339945_p657531455"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="31.370000000000005%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002039339945_p1859531759"><a name="zh-cn_topic_0000002039339945_p1859531759"></a><a name="zh-cn_topic_0000002039339945_p1859531759"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339945_row16726192116014"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1754534515"><a name="zh-cn_topic_0000002039339945_p1754534515"></a><a name="zh-cn_topic_0000002039339945_p1754534515"></a>Kernels</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p155195316518"><a name="zh-cn_topic_0000002039339945_p155195316518"></a><a name="zh-cn_topic_0000002039339945_p155195316518"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_p205155310512"><a name="zh-cn_topic_0000002039339945_p205155310512"></a><a name="zh-cn_topic_0000002039339945_p205155310512"></a>CANN二进制算子包，arch可选aarch64或x86_64。示例使用8.2.RC1版本。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p19595310517"><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a name="zh-cn_topic_0000002039339945_p19595310517"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="zh-cn_topic_0000002039339945_note1386820525510"><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><a name="zh-cn_topic_0000002039339945_note1386820525510"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339945_p12512532515"><a name="zh-cn_topic_0000002039339945_p12512532515"></a><a name="zh-cn_topic_0000002039339945_p12512532515"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row572619211108"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1863531756"><a name="zh-cn_topic_0000002039339945_p1863531756"></a><a name="zh-cn_topic_0000002039339945_p1863531756"></a>CANN</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p96553958"><a name="zh-cn_topic_0000002039339945_p96553958"></a><a name="zh-cn_topic_0000002039339945_p96553958"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p1245164917410"><a name="p1245164917410"></a><a name="p1245164917410"></a>CANN开发套件包，安装Toolkit和NNAL组件。示例使用8.2.RC1版本。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_p1862053354"><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a name="zh-cn_topic_0000002039339945_p1862053354"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=cann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="note53215352"><a name="note53215352"></a><a name="note53215352"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p43218520515"><a name="p43218520515"></a><a name="p43218520515"></a>请获取和服务器型号匹配的软件包。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row672652117018"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002039339945_p1468532510"><a name="zh-cn_topic_0000002039339945_p1468532510"></a><a name="zh-cn_topic_0000002039339945_p1468532510"></a>get-pip.py</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002039339945_p1869531256"><a name="zh-cn_topic_0000002039339945_p1869531256"></a><a name="zh-cn_topic_0000002039339945_p1869531256"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p626262173118"></a>用于安装pip模块。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a><a name="zh-cn_topic_0000002039339945_zh-cn_topic_0000001497364957_p39761346403"></a>curl -k https://bootstrap.pypa.io/get-pip.py -o get-pip.py</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row197268213011"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p1741916261997"><a name="p1741916261997"></a><a name="p1741916261997"></a>version.info</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p64191226897"><a name="p64191226897"></a><a name="p64191226897"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p1541911261191"><a name="p1541911261191"></a><a name="p1541911261191"></a>驱动版本信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p641914261992"><a name="p641914261992"></a><a name="p641914261992"></a>从host拷贝“/usr/local/Ascend/driver/version.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row1412215399516"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p441932611914"><a name="p441932611914"></a><a name="p441932611914"></a>ascend_install.info</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p1441912261994"><a name="p1441912261994"></a><a name="p1441912261994"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p241915261398"><a name="p241915261398"></a><a name="p241915261398"></a>驱动安装信息文件。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p1419226997"><a name="p1419226997"></a><a name="p1419226997"></a>从host拷贝“/etc/ascend_install.info”文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row151232039750"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p1341915261691"><a name="p1341915261691"></a><a name="p1341915261691"></a>vLLM</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p15419626298"><a name="p15419626298"></a><a name="p15419626298"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p1241919261392"><a name="p1241919261392"></a><a name="p1241919261392"></a>示例使用的推理引擎，使用v0.9.1分支。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p194190261916"><a name="p194190261916"></a><a name="p194190261916"></a>git clone -b v0.9.1 https://github.com/vllm-project/vllm.git</p>
<p id="p94197261593"><a name="p94197261593"></a><a name="p94197261593"></a>下载后，将<span class="filepath" id="filepath254119104122"><a name="filepath254119104122"></a><a name="filepath254119104122"></a>“vllm/requirements/build.txt”</span>中的torch版本修改为2.5.1。</p>
</td>
</tr>
<tr id="row1173819266428"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p541912261292"><a name="p541912261292"></a><a name="p541912261292"></a>vllm-ascend</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p9419326291"><a name="p9419326291"></a><a name="p9419326291"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p54194266916"><a name="p54194266916"></a><a name="p54194266916"></a>vLLM推理引擎在NPU上的适配插件，使用commitid：4014ad2a46e01c79fd8d98d6283404d0bc414dce。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p1241912261913"><a name="p1241912261913"></a><a name="p1241912261913"></a>git clone -b v0.9.1-dev https://github.com/vllm-project/vllm-ascend.git</p>
<p id="p174191826995"><a name="p174191826995"></a><a name="p174191826995"></a>cd vllm-ascend</p>
<p id="p144191526194"><a name="p144191526194"></a><a name="p144191526194"></a>git checkout 4014ad2a46e01c79fd8d98d6283404d0bc414dce</p>
<p id="p1498453811140"><a name="p1498453811140"></a><a name="p1498453811140"></a>然后修改requirements.txt中的torch-npu版本为2.5.1.post1。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row121231639952"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p164191026291"><a name="p164191026291"></a><a name="p164191026291"></a>Megatron-LM</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p84199262919"><a name="p84199262919"></a><a name="p84199262919"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p1741911261398"><a name="p1741911261398"></a><a name="p1741911261398"></a>训练后端使用Megatron的v0.12.1版本。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p5419326496"><a name="p5419326496"></a><a name="p5419326496"></a>git clone https://github.com/NVIDIA/Megatron-LM.git</p>
<p id="p041910261494"><a name="p041910261494"></a><a name="p041910261494"></a>cd Megatron-LM</p>
<p id="p641915267916"><a name="p641915267916"></a><a name="p641915267916"></a>git checkout core_v0.12.1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row144125121466"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p64199261893"><a name="p64199261893"></a><a name="p64199261893"></a>MindSpeed</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p1041919265911"><a name="p1041919265911"></a><a name="p1041919265911"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p14195261495"><a name="p14195261495"></a><a name="p14195261495"></a>训练后端使用MindSpeed，使用commitid：1f13e6fdbfd701ea7e045c8d6bb2469fab9775a7。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p10419726392"><a name="p10419726392"></a><a name="p10419726392"></a>git clone https://gitcode.com/Ascend/MindSpeed.git</p>
<p id="p204196261891"><a name="p204196261891"></a><a name="p204196261891"></a>cd MindSpeed</p>
<p id="p184203267912"><a name="p184203267912"></a><a name="p184203267912"></a>git checkout 1f13e6fdbfd701ea7e045c8d6bb2469fab9775a7</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row17301171913614"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p104201226695"><a name="p104201226695"></a><a name="p104201226695"></a>Verl</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p11420326694"><a name="p11420326694"></a><a name="p11420326694"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p5420122616912"><a name="p5420122616912"></a><a name="p5420122616912"></a>后训练框架，使用commitid：02f4386ae89c9a25863dca0bb8b6e119b2f01385。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p1942015268915"><a name="p1942015268915"></a><a name="p1942015268915"></a>git clone https://github.com/volcengine/verl.git</p>
<p id="p842011261918"><a name="p842011261918"></a><a name="p842011261918"></a>cd verl</p>
<p id="p17420152614911"><a name="p17420152614911"></a><a name="p17420152614911"></a>git checkout 02f4386ae89c9a25863dca0bb8b6e119b2f01385</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row93022191368"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p194201026591"><a name="p194201026591"></a><a name="p194201026591"></a>rl-plugin</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p742013268916"><a name="p742013268916"></a><a name="p742013268916"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p342010261294"><a name="p342010261294"></a><a name="p342010261294"></a>Verl在NPU上的适配插件，使用commitid：9a679fc3be95d162b78d42e9e3df569c30a89a5e。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p12420162614917"><a name="p12420162614917"></a><a name="p12420162614917"></a>git clone https://gitcode.com/Ascend/MindSpeed-RL.git</p>
<p id="p442017261493"><a name="p442017261493"></a><a name="p442017261493"></a>cd MindSpeed-RL/rl-plugin</p>
<p id="p18420826699"><a name="p18420826699"></a><a name="p18420826699"></a>git checkout 9a679fc3be95d162b78d42e9e3df569c30a89a5e</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339945_row33025197610"><td class="cellrowborder" valign="top" width="21.150000000000002%" headers="mcps1.2.5.1.1 "><p id="p1742032612912"><a name="p1742032612912"></a><a name="p1742032612912"></a>Dockerfile</p>
</td>
<td class="cellrowborder" valign="top" width="13.750000000000004%" headers="mcps1.2.5.1.2 "><p id="p10420426095"><a name="p10420426095"></a><a name="p10420426095"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="33.730000000000004%" headers="mcps1.2.5.1.3 "><p id="p442032611911"><a name="p442032611911"></a><a name="p442032611911"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="31.370000000000005%" headers="mcps1.2.5.1.4 "><p id="p64203261914"><a name="p64203261914"></a><a name="p64203261914"></a>-</p>
</td>
</tr>
</tbody>
</table>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE] 说明 
>-   本章节以两台Atlas 900 A3 SuperPoD 超节点，配套Ubuntu 20.04、Python  3.10、CANN 8.2.RC1版本为例介绍后训练镜像的制作，使用过程中需根据实际情况修改相关步骤。
>-   详细操作步骤及版本配套关系请参见[相关文档](https://verl.readthedocs.io/en/latest/ascend_tutorial/ascend_quick_start.html)。

**操作步骤<a name="section975144917188"></a>**

1.  参照[表1](#zh-cn_topic_0000002039339945_table1172542119019)，在宿主机上完成软件包的准备工作。
2.  编写如下Dockerfile。

    ```
    FROM ubuntu:20.04 
    WORKDIR /root 
    COPY . . 
      
    ARG HOST_ASCEND_BASE=/usr/local/Ascend 
    
    ARG TOOLKIT_PATH=/usr/local/Ascend/toolkit/latest 
    ARG TOOLKIT=Ascend-cann-toolkit_8.2.RC1_linux-aarch64.run
    ARG NNAL=Ascend-cann-nnal_8.2.RC1_linux-aarch64.run
    ARG KERNEL=Atlas-A3-cann-kernels_8.2.RC1_linux-aarch64.run 
     
    RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf 
      
    RUN echo "deb http://repo.huaweicloud.com/ubuntu-ports/ focal main restricted universe multiverse\n\ 
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-updates main restricted universe multiverse\n\ 
    deb http://repo.huaweicloud.com/ubuntu-ports/ focal-backports main restricted universe multiverse\n\ 
    deb http://ports.ubuntu.com/ubuntu-ports/ focal-security main restricted universe multiverse" > /etc/apt/sources.list 
     
    RUN umask 0022 && apt update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends software-properties-common
    RUN umask 0022 && add-apt-repository ppa:deadsnakes/ppa && apt update && apt autoremove -y python python3 && apt install -y python3.10 python3.10-dev vim patch gcc g++ make cmake build-essential libbz2-dev libreadline-dev wget curl llvm libncurses5-dev libncursesw5-dev xz-utils tk-dev liblzma-dev m4 dos2unix libopenblas-dev git libjemalloc2 libomp-dev net-tools
     
     
    # 建立Python软链接
    RUN ln -s /usr/bin/python3.10 /usr/bin/python
    RUN unlink /usr/bin/python3
    RUN ln -s /usr/bin/python3.10 /usr/bin/python3
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python-config
    RUN ln -s /usr/bin/python3.10-config /usr/bin/python3-config
      
    RUN umask 0022 && python get-pip.py
    
    # 配置pip源 
    RUN mkdir -p ~/.pip \ 
    && echo '[global] \n\ 
    index-url=https://mirrors.huaweicloud.com/repository/pypi/simple\n\ 
    trusted-host=mirrors.huaweicloud.com' >> ~/.pip/pip.conf 
     
    # 时区 
    RUN ln -sf /usr/share/zoneinfo/UTC /etc/localtime 
      
     
    # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser 
      
    # Ascend包 
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录 
    RUN umask 0022 &&  \ 
        cp ascend_install.info /etc/ && \ 
        mkdir -p /usr/local/Ascend/driver/ && \ 
        cp version.info /usr/local/Ascend/driver/ && \ 
        chmod +x $TOOLKIT && \ 
        chmod +x $KERNEL && \
        chmod +x $NNAL
      
    RUN umask 0022 && ./$TOOLKIT --install-path=/usr/local/Ascend/ --install --quiet 
    RUN umask 0022 && . /usr/local/Ascend/ascend-toolkit/set_env.sh && ./$KERNEL --install --quiet 
    RUN umask 0022 && . /usr/local/Ascend/ascend-toolkit/set_env.sh && ./$NNAL --install --quiet 
     
    ```

3.  构建镜像。执行以下命令生成镜像**。注意不要遗漏命令结尾的**“.“。

    ```
    docker build -t verl-train:v1 .
    ```

4.  安装推理服务包。执行以下命令启动容器。

    ```
    docker run -it \
    -v /usr/local/Ascend/driver:/usr/local/Ascend/driver \
    verl-train:v1 /bin/bash
    ```

    然后在容器内执行以下命令：

    ```
    source /usr/local/Ascend/driver/bin/setenv.bash;
    source /usr/local/Ascend/ascend-toolkit/set_env.sh;
    source /usr/local/Ascend/nnal/atb/set_env.sh;
    source /usr/local/Ascend/nnal/asdsip/set_env.sh;
    # vLLM安装 
    cd vllm && pip install -r requirements/build.txt -i https://mirrors.aliyun.com/pypi/simple/ && pip install -r requirements/common.txt -i https://mirrors.aliyun.com/pypi/simple/ && VLLM_TARGET_DEVICE=empty python setup.py develop && cd ..
    # vllm-ascend安装 
    cd vllm-ascend && pip install -v -e . && cd ..
    # Megatron安装 
    cd Megatron-LM && git checkout core_v0.12.1 && pip install -e .  && cd ..
      
    # MindSpeed安装 
    cd MindSpeed && pip install -e . && cd ..
      
    # Verl安装 
    cd verl && pip install -e . && cd ..
      
    # Verl插件安装 
    cd MindSpeed-RL/rl-plugin && pip install -v -e . && cd ..
    ```

    -   如果在安装vllm-ascend过程中，出现找不到torch的cmake路径的报错，可以参考以下命令指定“CMAKE\_PREFIX\_PATH“进行安装：

        ```
        CMAKE_PREFIX_PATH=/usr/local/lib/python3.10/dist-packages/torch/share/cmake/Torch/ pip install -v -e .
        ```

    -   如果在安装Verl过程中，出现找不到README.md的报错，可以在“MindSpeed-RL/rl-plugin“创建README.md文件，内容不限。
    -   安装完成后，如果发现torch版本不是2.5.1，torchvision版本不是0.20.1，则重新安装torch 2.5.1和torchvision 0.20.1。

5.  在另一个窗口执行以下命令保存镜像。为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。

    ```
    # 找到容器ID
    docker ps | grep verl-train
    # 保存容器为镜像，<container_id>替换为实际的容器ID
    docker commit <container_id> verl-train:v1
    ```



### 脚本适配<a name="ZH-CN_TOPIC_0000002511426481"></a>

#### 流程说明<a name="ZH-CN_TOPIC_0000002511346469"></a>

模型脚本需要适配CKPT之后才可以使用断点续训功能，脚本适配大致流程和逻辑如[图1](#fig88341718121515)所示。整体示例可参考[图1](#fig88341718121515)。

**图 1**  脚本适配流程<a name="fig88341718121515"></a>  
![](../../figures/scheduling/脚本适配流程.png "脚本适配流程")


#### 适配示例<a name="ZH-CN_TOPIC_0000002511346445"></a>

本章节将指导用户step by step地完成断点续训的适配步骤。

-   [PyTorch场景适配示例（基于MindSpeed-LLM）](#zh-cn_topic_0000002003180016_section412442472511)
-   [MindSpore场景适配示例（基于MindFormers）](#zh-cn_topic_0000002003180016_section718243883518)
-   [强化学习后训练场景适配示例（基于Verl）](#section1335017512276)

>[!NOTE] 说明 
>-   为保证优雅容错与进程级在线恢复功能的正常使用，请将K8s集群master节点与worker节点的时钟保持一致。
>-   断点续训展示的组件代码为开源代码，其中涉及到相关安全说明请参见[安全说明](../appendix.md#安全说明)。
>-   下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
>-   模型的参数配置，根据模型仓的模型配置以实际情况来写。若修改不当，可能会引发不可预知的问题。
>-   若训练过程中出现“Failed to bind the IP port. Reason: The IP address and port have been bound already”报错，可以按照如下进行配置，详情请参见《CANN 环境变量参考》中的“HCCL_HOST_SOCKET_PORT_RANGE”章节。

       export HCCL_HOST_SOCKET_PORT_RANGE="60000-60050"
       export HCCL_NPU_SOCKET_PORT_RANGE="61000-61050"

**PyTorch场景适配示例（基于MindSpeed-LLM）<a name="zh-cn_topic_0000002003180016_section412442472511"></a>**

训练代码与数据集准备，可以参考[MindSpeed-LLM使用指南](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/docs/pytorch/solutions/pretrain/pretrain.md)。下面以两台Atlas 800T A2 训练服务器为例，说明具体操作步骤。

1.  拉取训练代码。

    ```
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/MindSpeed-LLM.git
    git clone https://github.com/NVIDIA/Megatron-LM.git
    cd MindSpeed-LLM
    git checkout 2.3.0
    cd ..
    cd Megatron-LM 
    git checkout core_v0.12.1
    cp -r megatron ../MindSpeed-LLM #此处目的是将Megatron-LM项目下的Megatron目录复制到MindSpeed-LLM项目下
    ## 重命名MindSpeed-LLM为QWEN3_for_PyTorch_2.7_code
    cd ..
    mv MindSpeed-LLM QWEN3_for_PyTorch_2.7_code
    ```

2.  获取模型权重。

    请用户自行从[Qwen3](https://huggingface.co/Qwen/Qwen3-8B/tree/main)下载模型权重放到服务器某目录下，如“/data/atlas\_dls/public/dataset/qwen3-8b-hf“。

3.  获取数据集。

    请用户自行从[Alpaca](https://huggingface.co/datasets/tatsu-lab/alpaca/blob/main/data/train-00000-of-00001-a09b74b3ef9c3b56.parquet)下载数据集（以Alpaca数据集为例）放到服务器某目录下，如“/data/atlas\_dls/public/dataset/qwen3-alpaca“。

4.  处理数据集。
    1.  启动容器。

        ```
        docker run -it -v /data/atlas_dls/public/:/data/atlas_dls/public/ -e ASCEND_VISIBLE_DEVICES=0-7 mindspeed-dl:v1 bash
        ```

    2.  在容器中执行如下操作。

        ```
        export TORCH_DEVICE_BACKEND_AUTOLOAD=0
        source /usr/local/Ascend/cann/set_env.sh
        cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code
        # 可选，如下为安装MindSpeed加速库操作，可在任意目录下执行。若制作镜像时已安装，则跳过该操作
        git clone https://gitcode.com/ascend/MindSpeed.git 
        cd MindSpeed 
        git checkout 2.3.0_core_r0.12.1
        pip install -r requirements.txt 
        pip install -e . 
        export PYTHONPATH=/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/MindSpeed:$PYTHONPATH
        cd ..
        ```

    3.  处理数据集。

        Qwen3要求使用Transformers\>=4.51.0，因此Python需使用3.9及以上版本且需要安装4.51.0及以上的Transformers。

        ```
        python preprocess_data.py \
            --input /data/atlas_dls/public/dataset/qwen3-alpaca/train-00000-of-00001-a09b74b3ef9c3b56.parquet \ # 数据集文件路径
            --tokenizer-name-or-path /data/atlas_dls/public/dataset/qwen3-8b-hf \ # 开源模型权重文件目录
            --tokenizer-type PretrainedFromHF \
            --handler-name GeneralPretrainHandler \
            --output-prefix /data/atlas_dls/public/dataset/qwen3-alpaca/alpaca \ # 会生成alpaca_text_document.bin和.idx文件
            --json-keys text \
            --workers 4 \
            --log-interval 1000
        ```

        >[!NOTE] 说明 
        >若出现报错：/usr/local/lib/python3.10/dist-packages/sklearn/utils/../../scikit\_learn.libs/libgomp-947d5fa1.so.1.0.0: cannot allocate memory in static TLS block，可执行以下命令预加载libgomp库。
        >```
        >export LD_PRELOAD="/usr/local/lib/python3.10/dist-packages/scikit_learn.libs/libgomp-947d5fa1.so.1.0.0"
        >```

5.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3“目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```

6.  获取[训练任务YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml)。该YAML中已经配置了Pod级别重调度、进程级别重调度、进程级在线恢复、弹性训练等。根据实际情况配置挂载卷的服务器IP地址、各种重调度级别等。

    进程级别重调度、进程级在线恢复、弹性训练等训练进程级别的恢复与优雅容错不可同时存在。优雅容错的配置步骤请参见[优雅容错模式](#可选配置组件)。

7.  配置训练启动脚本train\_start.sh和训练任务YAML，请根据实际情况进行修改。
    1.  修改启动脚本基础参数。

        ```
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/ttplogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/trainlogs
        mkdir -p /job/code/alllogs/$MINDX_TASK_ID/demo/
        # 日志保存路径，可根据实际情况修改
        export ASCEND_PROCESS_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/plogs/$XDL_IP       # 设置plog保存路径，其中$MINDX_TASK_ID为Ascend Operator注入的任务UID环境变量，$XDL_IP为任务YAML中写入的环境变量status.hostIP
        export TTP_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/ttplogs/ttplog$XDL_IP-$RANK    # 设置TTP日志保存路径，其中$RANK为Ascend Operator为PyTorch框架注入的环境变量
        export TRAIN_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/trainlogs/$XDL_IP-$RANK      # 设置训练日志保存路径
        export GLOO_SOCKET_IFNAME=enp189s0f0               # 物理机上可以通信的网口，根据主节点高速网卡实际情况进行配置，如任务YAML中配置hostNetwork为false，则设置为eth0
        export HCCL_SOCKET_IFNAME=enp189s0f0               # 如任务YAML中配置hostNetwork为false，则设置为eth0
         
        CKPT_SAVE_DIR="/job/code/output/ckpt" # 训练完成后的权重保存路径
        DATA_PATH="/job/data/alpaca_text_document" # 数据集路径，填入数据预处理时保存的数据路径
        TOKENIZER_PATH="/job/data/qwen3-8b-hf" # 词表路径，填入下载的开源权重词表路径
        CKPT_LOAD_DIR="/job/code/output/ckpt" # 权重加载路径
        ```

    2.  使用TaskD完成进程级别重调度、进程级在线恢复、进程级别原地恢复或弹性训练，还需拉起TaskD  Manager。
        1.  创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

            ```
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
             
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
             
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 说明 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

        2.  在训练脚本中增加以下代码，拉起TaskD  Manager。

            在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

            ```
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
            export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python manager.py &           # 具体执行路径由当前路径决定
            fi
                

            torchrun $DISTRIBUTED_ARGS ...
            ```

        3.  修改训练任务YAML，新增容器端口，在所有的Pod下增加TaskD通信使用的端口9601（如已有则跳过）。

            ```
            ...
            ports:                         
               - containerPort: 9601             
                 name: taskd-port 
            ...
            ```

**MindSpore场景适配示例（基于MindFormers）<a name="zh-cn_topic_0000002003180016_section718243883518"></a>**

训练代码与数据集准备，可以参考[MindFormers文档](https://gitee.com/mindspore/mindformers/tree/master/configs/qwen3)。下面以两台Atlas 900 A3 SuperPoD 超节点为例，说明具体操作步骤。

1.  准备代码。

    ```
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitee.com/mindspore/mindformers.git
    cd mindformers
    git checkout f06a946af29c8c7e002a6c49458f513d47b642e5
    # 将mindformers重命名为QWEN3_for_MS_code
    cd ..
    mv mindformers QWEN3_for_MS_code
    ```

2.  准备数据集。

    请用户自行从[DagsHub](https://dagshub.com/DagsHub/WIkiText-103/src/main/dataset/tokens/wiki.train.tokens)下载数据集并放到服务器某目录下，如“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset“。

3.  转换数据集。
    1.  下载数据集转换脚本。

        从[数据集转换](https://gitee.com/mindspore/mindformers/issues/ICOKGY)下载数据集转换脚本并放到服务器某目录下，如“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset/gen\_wiki\_json.py“。

    2.  下载tokenizer文件。

        从[Qwen3-32B](https://huggingface.co/Qwen/Qwen3-32B/tree/main)下载tokenizer文件并放到服务器某目录下，如“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset/Qwen3-32B-tokenizer“。

    3.  转换数据集。
        1.  启动容器并挂载所需文件。

            ```
            docker run -it -v /data/atlas_dls/public/code/:/data/atlas_dls/public/code/ mindformers-dl:v1 bash
            ```

        2.  执行转换脚本，将wiki.train.tokens转换为jsonl格式。

            ```
            # 执行该脚本需要的Python环境，请提前准备Python环境
            cd /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset
            python gen_wiki_json.py --input wiki.train.tokens  --output wiki.jsonl 
            ```

        3.  将jsonl格式数据转为bin格式数据。

            ```
            # 执行时若报错ModuleNotFoundError: No module named 'xxx'，请自行安装依赖
            cd /data/atlas_dls/public/code/QWEN3_for_MS_code
            python toolkit/data_preprocess/megatron/preprocess_indexed_dataset.py \
              --input /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset /wiki.jsonl \
              --output-prefix /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset /wiki103-megatron \
              --tokenizer-type HuggingFaceTokenizer \
              --tokenizer-dir /data/atlas_dls/public/code/QWEN3_for_MS_code/dataset/Qwen3-32B-tokenizer # 其他规格的模型可以调整为对应的tokenizer路径
            ```

            运行完成后，“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset“目录下会生成“wiki103-megatron\_text\_document.bin“和“wiki103-megatron\_text\_document.idx“文件。 填写数据集路径时，需要使用“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/dataset/wiki103-megatron\_text\_document“，不需要带后缀名。

4.  获取[训练任务YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/ranktable/mindspore/Qwen3/yamls/ms_multinodes_acjob_superpod.yaml)和[训练启动脚本](https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/ranktable/mindspore/Qwen3/msrun_launcher.sh)，并进行修改。
    1.  若训练任务YAML中“hostNetwork“参数值为“false“，则需要将启动脚本中“GLOO\_SOCKET\_IFNAME“的值设置为“eth0“。示例如下：

        ```
        export GLOO_SOCKET_IFNAME=eth0  #eth0是容器内可以通信的网口
        export HCCL_SOCKET_IFNAME=eth0
        ```

        然后根据实际情况修改启动脚本中的其他参数。

    2.  根据实际情况修改任务YAML中挂载卷的服务器IP地址等配置。
    3.  使用TaskD完成进程级别重调度、进程级在线恢复、进程级别原地恢复、借轨通信任务暂停与回切或在线压测，还需拉起TaskD  Manager。
        1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

            ```
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os
            
            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
            
            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE] 说明 
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

        2.  在训练脚本中增加以下代码拉起TaskD  Manager。在以下代码中，前两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

            ```
            TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
            export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python manager.py &   # 具体执行路径由当前路径决定
            fi
            msrun ...
            ```

        3.  修改训练任务YAML，新增容器端口，在所有的Pod下增加TaskD通信使用的端口9601（如已有则跳过）。

            ```
            ...
            ports:                         
               - containerPort: 9601             
                 name: taskd-port 
            ...
            ```

5.  修改参数模型配置文件。
    1.  打开代码目录下“configs/qwen3/pretrain\_qwen3\_32b\_4k.yaml“文件。

        ```
        vi configs/qwen3/pretrain_qwen3_32b_4k.yaml
        ```

    2.  按“i”进入编辑模式，修改参数模型配置文件。
        1.  修改如下加粗配置，包括数据集路径、分布式并行参数、模型参数等。以下模型参数仅供参考，如有需要请自行修改。

            ```
            train_dataset: &train_dataset
              data_loader:
                type: BlendedMegatronDatasetDataLoader
                datasets_type: "GPTDataset"
                sizes:
                  - 8000  # Number of samples in the training set
                  - 0     # Number of samples in the test set (currently unsupported)
                  - 0     # Number of samples in the evaluation set (currently unsupported)
                config:
                  seed: 1234  # Random seed for data sampling
                  split: "1, 0, 0"  # Proportions for training, test, and evaluation sets (test/eval currently unsupported)
                  seq_length: 4096  # Sequence length of the dataset
                  eod_mask_loss: False  # Whether to calculate loss at the end-of-document (EOD)
                  reset_position_ids: False  # Whether to reset position_ids at EOD
                  create_attention_mask: True  # Whether to include attention_mask in the dataset
                  reset_attention_mask: False  # Whether to reset attention_mask at EOD, creating a stepped attention_mask
                  create_compressed_eod_mask: False  # Whether to include a compressed attention_mask
                  eod_pad_length: 128  # Length of the compressed attention_mask
                  eod: 1  # Token ID for EOD in the dataset
                  pad: -1  # Token ID for padding in the dataset
                  data_path:  # Sampling proportion and path for the Megatron dataset
                    - '1'
                    - "/job/data/wiki103-megatron_text_document" # 数据集路径
            ……
            # Parallel configuration
            parallel_config:
              data_parallel: &dp 4  # Number of data parallel. If using the high availability feature, it must be an even number.
              model_parallel: 8  # Number of model parallel
              pipeline_stage: 1  # Number of pipeline parallel
              micro_batch_num: 1  # Pipeline parallel microbatch size
              use_seq_parallel: False  # Whether to enable sequence parallelism
              gradient_aggregation_group: 1  # Size of the gradient communication operator fusion group
            # When model_parallel > 1, setting micro_batch_interleave_num to 2 may accelerate the training process.
            micro_batch_interleave_num: 1
            ……
            model:
              model_config:
                # Configurations from Hugging Face
                vocab_size: 75968            # 此处改小了模型参数仅供测试，如有需要请自行调整
                hidden_size: 2560           # 此处改小了模型参数仅供测试，如有需要请自行调整
                intermediate_size: 12800   # 此处改小了模型参数仅供测试，如有需要请自行调整
                num_hidden_layers: 32      # 此处改小了模型参数仅供测试，如有需要请自行调整
                num_attention_heads: 32    # 此处改小了模型参数仅供测试，如有需要请自行调整
                num_key_value_heads: 8
                head_dim: 128
                hidden_act: 'swiglu'
                max_position_embeddings: 4096
                seq_length: 4096
                initializer_range: 0.02
                rms_norm_eps: 1.e-6
                use_cache: True
                tie_word_embeddings: False
                rope_theta: 1000000.
                attention_bias: False
                use_flash_attention: True
                add_bias_linear: False
                eos_token_id: 151645
                pad_token_id: 151643
                bos_token_id: 151643
                attention_dropout: 0.0
                # Configurations from MindFormers
                hidden_dropout: 0.0
                input_sliced_sig: True
                untie_embeddings_and_output_weights: True
                position_embedding_type: "rope"
                qk_layernorm: True
                use_contiguous_weight_layout_attention: False
                qkv_concat: True
                offset: [0]
                params_dtype: "float32"
                compute_dtype: "bfloat16"
                layernorm_compute_dtype: "float32"
                softmax_compute_dtype: "float32"
                rotary_dtype: "float32"
                residual_dtype: "float32"
                model_type: "qwen3"
                architectures: ["Qwen3ForCausalLM"]
            ```

        2.  （可选）使用临终CKPT的场景，在保存CKPT后通过Pod级别重调度加载CKPT，需修改如下配置字段。

            首次拉起必须保证“load\_checkpoint“参数值的目录下存在正常可用的CKPT或该目录为空，否则可能导致训练无法正常拉起。

            ```
            resume_training: True 
            src_strategy_path_or_dir: './output/strategy'
            load_checkpoint: './output/checkpoint'
            ```

    3.  按“Esc”键，输入**:wq!**，按“Enter”保存并退出编辑。

**强化学习后训练场景适配示例（基于Verl）<a name="section1335017512276"></a>**

MindCluster仅支持Job级别重调度。Verl的训练任务被Ray集群所管理，为适配MindCluster的Ascend Job任务部署，每个Worker节点上部署一个Pod，Pod内承载该Ray集群上的所有进程。Ray集群的head节点根据Ascend Operator注入的环境变量RANK=0所在的节点决定。RANK=0节点的Pod启动Ray集群，提交Verl后训练任务，其他Worker节点的Pod加入Ray集群。最后所有节点都检测提交的训练任务是否存在异常。

-   若存在异常，则以非0退出，Volcano感知到业务异常触发Job级别重调度。
-   若没有异常且任务已结束，则以0退出。

>[!NOTE] 说明 
>-   以下所有步骤确保在每台Worker节点均执行。
>-   本示例使用[Qwen3 30B MoE](https://modelscope.cn/models/Qwen/Qwen3-30B-A3B-Instruct-2507)模型与[DAPO-Math-17k](https://modelscope.cn/datasets/AI-ModelScope/DAPO-Math-17k)数据集。

下面以两台Atlas 900 A3 SuperPoD 超节点为例，说明具体操作步骤。

1.  模型权重转换，将HuggingFace模型转换为Megatron模型，可参考[Verl模型转化脚本](https://verl.readthedocs.io/en/latest/advance/checkpoint.html#huggingface-to-megatron-distcheckpoint-details)。

    ```
    # 启动容器，具体模型路径根据实际情况修改
    docker run -it \
    -v /qwen30b/Qwen3-30B-A3B-Instruct-2507:/qwen30b/Qwen3-30B-A3B-Instruct-2507 \
    -v /usr/local/Ascend/driver:/usr/local/Ascend/driver \
    -e ASCEND_VISIBLE_DEVICES=0-15 \
    verl:v1 /bin/bash
     
    # 执行权重转化
    cd ~/verl
    python scripts/converter_hf_to_mcore.py \
    --hf_model_path /qwen30b/Qwen3-30B-A3B-Instruct-2507 \
    --output_path /qwen30b/Qwen3-30B-A3B-Instruct-Mcore \
    ```

    若出现如下错误，则先执行如下命令：

    ```
    export LD_PRELOAD="/usr/local/lib/python3.10/dist-packages/sklearn/utils/../../scikit_learn.libs/libgomp-947d5fa1.so.1.0.0"
    ```

    ![](figures/报错.png)

2.  构建Verl的Qwen3 30B MoE的训练脚本。其中推理后端为vLLM，训练后端为Megatron。

    获取脚本示例[run\_dapo\_qwen3\_30b\_a3b\_megatron.sh](https://gitcode.com/Ascend/mindxdl-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/verl/run_dapo_qwen3_30b_a3b_megatron.sh)，并将其放置到verl路径中examples\_npu下。同时在“examples\_npu/config“路径下创建两个文件dapo\_trainer-megatron.yaml和runtime\_env.yaml，其内容如下：

    -   dapo\_trainer-megatron.yaml

        ```
        # examples_npu/config/dapo_trainer-megatron.yaml
        hydra:
          searchpath:
            - file://verl/trainer/config
        defaults:
          - ppo_megatron_trainer
          - _self_
        data:
          gen_batch_size: ${data.train_batch_size}
        reward_model:
          reward_manager: dapo
          overlong_buffer: 
            enable: False # We try to avoid forgetting to set enable
            len: 0
            penalty_factor: 0.0
            log: False
        algorithm:
          filter_groups:
            _target_: verl.trainer.config.FilterGroupsConfig
            enable: False # We try to avoid forgetting to set enable
            metric: null # acc / score / seq_reward / seq_final_reward / ...
            max_num_gen_batches: 0 # Non-positive values mean no upper limit
        trainer:
          project_name: verl-dapo
        ```

    -   runtime\_env.yaml

        ```
        # examples_npu/config/runtime_env.yaml
        working_dir: ./
        excludes: ["/.git/"]
        env_vars:
          HCCL_EXEC_TIMEOUT: "7200"
          HCCL_CONNECT_TIMEOUT: "7200"
          VLLM_USE_V1: "1"
          VLLM_VERSION: "0.9.1"
          HCCL_IF_BASE_PORT: "23999"
          HCCL_ASYNC_ERROR_HANDLING: "0"
          P2P_HCCL_BUFFSIZE: "20"
        ```

3.  构建适配MindCluster的Ray启动脚本。在每台Worker节点上准备好Ray启动脚本，放到两台Atlas 900 A3 SuperPoD 超节点上。其中的网卡信息需根据实际情况配置，其余脚本可以保持不变。

    获取脚本示例[start.sh](https://gitcode.com/Ascend/mindxdl-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/verl/start.sh)，并将脚本放置到verl目录下。

4.  获取[准备任务YAML](#准备任务yaml)中的verl-resche.yaml，根据实际情况修改其中的参数，然后执行如下命令启动任务。

    ```
    kubectl apply -f verl-resche.yaml
    ```

    启动任务后，会显示如下迭代信息：

    ![](figures/迭代信息.png)



### 准备任务YAML<a name="ZH-CN_TOPIC_0000002511426415"></a>

集群调度组件为用户提供YAML示例，用户需要根据使用的功能、模型类型和任务类型等，并根据使用的故障处理模式，选择相应的YAML示例并根据需求进行相应修改后才可使用。

**表 1**  训练任务YAML示例

<a name="table350244433714"></a>
<table><thead align="left"><tr id="row135031644183710"><th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.1"><p id="p8503244173715"><a name="p8503244173715"></a><a name="p8503244173715"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="16.173234646929384%" id="mcps1.2.8.1.2"><p id="p145038448375"><a name="p145038448375"></a><a name="p145038448375"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="8.521704340868173%" id="mcps1.2.8.1.3"><p id="p919210345266"><a name="p919210345266"></a><a name="p919210345266"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="13.672734546909378%" id="mcps1.2.8.1.4"><p id="p5503544193713"><a name="p5503544193713"></a><a name="p5503544193713"></a>模型</p>
</th>
<th class="cellrowborder" valign="top" width="15.393078615723146%" id="mcps1.2.8.1.5"><p id="p19672186404"><a name="p19672186404"></a><a name="p19672186404"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="15.433086617323463%" id="mcps1.2.8.1.6"><p id="p1096741894013"><a name="p1096741894013"></a><a name="p1096741894013"></a>获取链接</p>
</th>
<th class="cellrowborder" valign="top" width="15.413082616523303%" id="mcps1.2.8.1.7"><p id="p2967518174012"><a name="p2967518174012"></a><a name="p2967518174012"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row4503174412371"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p09365292408"><a name="p09365292408"></a><a name="p09365292408"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><a name="ul129364297402"></a><a name="ul129364297402"></a><ul id="ul129364297402"><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p319343422611"><a name="p319343422611"></a><a name="p319343422611"></a><span id="ph310231710274"><a name="ph310231710274"></a><a name="ph310231710274"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p493616294406"><a name="p493616294406"></a><a name="p493616294406"></a><span id="ph22631282914"><a name="ph22631282914"></a><a name="ph22631282914"></a>Qwen3</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p893610293406"><a name="p893610293406"></a><a name="p893610293406"></a>pytorch_multinodes_acjob_910b.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1987716427402"><a name="p1987716427402"></a><a name="p1987716427402"></a><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3/yamls/pytorch_multinodes_acjob_910b.yaml" target="_blank" rel="noopener noreferrer">pytorch_multinodes_acjob_910b.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p8936152964011"><a name="p8936152964011"></a><a name="p8936152964011"></a>示例默认使用2*8卡任务</p>
</td>
</tr>
<tr id="row91607510384"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p89371529174019"><a name="p89371529174019"></a><a name="p89371529174019"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><a name="ul393742934014"></a><a name="ul393742934014"></a><ul id="ul393742934014"><li><span id="ph139426426441"><a name="ph139426426441"></a><a name="ph139426426441"></a>Atlas 800T A2 训练服务器</span></li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p1319333422617"><a name="p1319333422617"></a><a name="p1319333422617"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p1893752924017"><a name="p1893752924017"></a><a name="p1893752924017"></a><span id="ph234505228"><a name="ph234505228"></a><a name="ph234505228"></a>Qwen3</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p1493742904013"><a name="p1493742904013"></a><a name="p1493742904013"></a><span id="ph153229411739"><a name="ph153229411739"></a><a name="ph153229411739"></a>ms_multinodes_acjob_superpod.yaml</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1637217494110"><a name="p1637217494110"></a><a name="p1637217494110"></a><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/master/samples/train/resumable-training/fault-tolerance/ranktable/mindspore/Qwen3/yamls/ms_multinodes_acjob_superpod.yaml" target="_blank" rel="noopener noreferrer">ms_multinodes_acjob_superpod.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p79373296408"><a name="p79373296408"></a><a name="p79373296408"></a>示例默认使用2*16卡任务</p>
</td>
</tr>
<tr id="row4955626202119"><td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.1 "><p id="p95874374213"><a name="p95874374213"></a><a name="p95874374213"></a>Ascend Job</p>
</td>
<td class="cellrowborder" valign="top" width="16.173234646929384%" headers="mcps1.2.8.1.2 "><p id="p553054612118"><a name="p553054612118"></a><a name="p553054612118"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="8.521704340868173%" headers="mcps1.2.8.1.3 "><p id="p1419311343262"><a name="p1419311343262"></a><a name="p1419311343262"></a>Verl</p>
</td>
<td class="cellrowborder" valign="top" width="13.672734546909378%" headers="mcps1.2.8.1.4 "><p id="p558710375211"><a name="p558710375211"></a><a name="p558710375211"></a>Qwen3-30B</p>
</td>
<td class="cellrowborder" valign="top" width="15.393078615723146%" headers="mcps1.2.8.1.5 "><p id="p85871337112114"><a name="p85871337112114"></a><a name="p85871337112114"></a>verl-resche.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="15.433086617323463%" headers="mcps1.2.8.1.6 "><p id="p1833456152315"><a name="p1833456152315"></a><a name="p1833456152315"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/master/samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/verl/verl-resche.yaml" target="_blank" rel="noopener noreferrer">verl-resche.yaml</a></p>
</td>
<td class="cellrowborder" valign="top" width="15.413082616523303%" headers="mcps1.2.8.1.7 "><p id="p4587737102119"><a name="p4587737102119"></a><a name="p4587737102119"></a>示例默认使用2*16卡任务</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>当前断点续训并未提供Atlas 900 A3 SuperPoD 超节点产品的示例YAML，用户可以在示例YAML中的labels下新增annotations字段即可。示例如下：
>```
>...
>  labels: 
>...
>  annotations:
>    sp-block: "32"   # 逻辑超节点芯片数量，sp-block字段的详细说明，可以参见[YAML参数说明](YAML参数说明-58.md)。
>...
>```


### 下发任务<a name="ZH-CN_TOPIC_0000002479226548"></a>

示例YAML中，任务部署在default命名空间下。本章节以Pytorch框架为例，下发训练任务。

1.  登录管理节点，进入YAML文件所在路径。
2.  在管理节点执行以下命令，使用YAML下发训练任务。

    ```
    kubectl apply -f XXX.yaml
    ```

    例如：

    ```
    kubectl apply -f pytorch_multinodes_acjob_910b.yaml
    ```

    回显如下：

    ```
    configmap/reset-config-default-test-pytorch created
    ascendjob.mindxdl.gitee.com/default-test-pytorch created
    ```


### 查看任务进程<a name="ZH-CN_TOPIC_0000002511426461"></a>

训练任务下发成功后，训练任务就可正常运行。可通过如下内容查看训练任务运行情况。

**查看所有训练任务<a name="section16792164211375"></a>**

查看当前节点上运行的所有训练任务，操作步骤如下。

1.  登录管理节点，进入YAML文件所在路径。
2.  执行以下命令，查看训练任务运行情况。

    ```
    kubectl get pods -A -o wide
    ```

    回显示例如下。

    ```
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node1          <none>           <none>
    default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
    ……
    ```

**查看单个Pod的训练任务<a name="zh-cn_topic_0000001621551937_section1141119143319"></a>**

查看其中一个Pod上运行的训练任务，操作步骤如下。

执行以下命令，查看训练任务运行情况。

```
kubectl logs default-test-pytorch-worker-0 -n default -f
```

回显示例如下，出现loss即表示任务正常运行。

![](../../figures/scheduling/unnaming-(7).png)

**查看是否存在CKPT文件<a name="section979416428371"></a>**

故障恢复功能是通过参考CKPT文件实现的，用户需要查看**存储节点**上是否存在CKPT文件。

用户可以等待训练任务运行时间超过用户设置的保存CKPT文件的时间后，查看设置的保存CKPT文件的路径下是否存在周期性CKPT文件，操作步骤如下。

1.  登录存储节点，执行以下步骤，进入CKPT文件路径。

    ```
    cd /data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/output/ckpt
    ```

2.  执行以下命令，查看当前目录是否存在周期性CKPT文件。

    ```
    ll ./
    ```

    回显示例如下，说明存在周期性CKPT文件。

    ```
    total 8
    drwx-xr-x-  18 root root   8192 Jun 22 18:39 iter_0000100
    -rw-r--r--  1 root root    2    Jun 22 18:39 latest_checkpointed_iteration.txt
    ```

3.  （可选）如果使用临终遗言，可以在保存CKPT的路径下，执行以下命令，查看当前目录是否存在临终CKPT文件。

    ```
    ll ./
    ```

    回显示例如下，说明存在临终CKPT文件。

    ```
    total 8
    drwx-xr-x-  18 root root   8192 Jun 22 15:39 iter_0000009
    -rw-r--r--  1 root root    2    Jun 22 15:39 latest_checkpointed_iteration.txt
    ```


### 查看训练结果<a name="ZH-CN_TOPIC_0000002479386554"></a>

#### （可选）构造故障<a name="ZH-CN_TOPIC_0000002511426449"></a>

本章节将指导用户构造简单的故障，包括节点故障、参数面网络故障和业务面故障。

>[!NOTE] 说明 
>构造芯片故障存在安全风险，如需构造请联系华为技术支持工程师处理。

**构造节点故障<a name="section173881558133914"></a>**

通过重启训练节点，模拟节点下电导致节点状态丢失。该故障在节点重启完成后可自动恢复。

1.  在训练任务正常训练出iteration后，登录正在训练的节点。
2.  执行以下命令，重启该训练节点，模拟节点状态丢失故障。

    ```
    reboot
    ```

3.  在Master节点多次执行以下命令，查看Pod状态。

    ```
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

4.  在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的CheckPoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

**构造参数面网络故障<a name="section22113033919"></a>**

通过断开NPU网络链路模拟参数面网络故障。NPU网络故障不影响单机训练任务。用户在断开链路后需手动恢复，否则该故障会一直存在。

1.  在训练任务正常训练出iteration后，登录正在训练的节点。
2.  执行以下命令，构造NPU网络链路故障。

    ```
    hccn_tool -i {device_id} -link -s down
    ```

    >[!NOTE] 说明 
    >_device\_id_为NPU的ID，可以通过npu-smi info命令查看NPU的ID。

3.  执行以下命令，查看NPU链路状态。

    ```
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障构造成功。

    ```
    net health status: Fault
    ```

4.  在Master节点多次执行以下命令，查看Pod状态。

    ```
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

5.  在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的CheckPoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-06-22 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```

6.  执行以下命令，恢复NPU网络链路故障。

    ```
    hccn_tool -i {device_id} -cfg recovery
    ```

7.  执行以下命令，查看NPU链路状态。

    ```
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障已经恢复。

    ```
    net health status: Success
    ```

**构造业务面故障<a name="section9891038124213"></a>**

通过删除训练进程，模拟业务面故障。

1.  在训练任务正常训练出iteration后，登录正在训练的节点。
2.  执行以下命令，使用训练启动脚本，查询训练进程信息。

    ```
    ps -ef | grep python| grep 训练启动脚本.py
    ```

3.  执行以下命令，手动删除PID最小的训练进程。

    ```
    kill -9 pid
    ```

4.  在Master节点多次执行以下命令，查看Pod状态。

    ```
    kubectl get pod -A
    ```

    可以看到Pod状态从Terminating到Pending，最后为Running状态，表示训练任务已经重新拉起。

5.  在Master节点执行以下命令，查看训练日志，记录续训成功时间。

    ```
    kubectl logs -n 命名空间名称 Pod名称
    ```

    回显示例如下，表示发生故障时，使用最近保存的第9步的CheckPoint文件恢复，实现训练任务第10个iteration开始继续训练。

    ```
    [2025-06-22 14:47:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 1932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    [2025-10-16 14:47:02] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
    ......
    ```


#### 重调度模式<a name="ZH-CN_TOPIC_0000002479386534"></a>

**重调度情况<a name="section87441013105513"></a>**

>[!NOTE] 说明 
>当节点发生故障时，Volcano会将该训练任务调度到其他满足条件的节点上继续运行。

登录管理节点，执行以下命令查看训练任务运行情况。

```
kubectl get pods -A -o wide
```

故障前，若训练任务调度到了node1和node2上面，当node1节点上发生故障，此时Volcano组件会将node1和node2上训练任务重调度到node2和node3节点上，重调度后回显示例如下。

```
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE   IP                NODE           NOMINATED NODE   READINESS GATES
default          default-test-pytorch-master-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node2          <none>           <none>
default          default-test-pytorch-worker-0              1/1     Running   0          5s    xxx.xxx.xxx.xxx   node3          <none>           <none>
……
```

**查看其中一个Pod运行情况<a name="section28985295314"></a>**

执行以下命令，查看单个Pod的训练任务运行情况。

```
kubectl logs default-test-pytorch-worker-0 -n default -f
```

回显如下表示发生故障时，使用最近保存的第9步的CheckPoint文件恢复，实现训练任务第10个iteration开始继续训练。

```
2025-09-08 11:34:00.400331 warn 1900637 [77840][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.401841 warn 1900631 [28432][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.402489 warn 1900639 [10928][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.426989 warn 1900627 [98608][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
2025-09-08 11:34:00.429141 warn 1900634 [24592][PYH tft_replica_optimizer.py:659] Replica optimizer increase Memory On Chip Usage by:0.6572 GB!
(min, max) time across ranks (ms):
    load-checkpoint ................................: (32107.12, 32108.53)
(min, max) time across ranks (ms):
    model-and-optimizer-setup ......................: (32528.79, 32544.35)
    train/valid/test-data-iterators-setup ..........: (72.68, 656.79)
[rank16]:[W908 11:34:01.252908110 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank24]:[W908 11:34:01.254614170 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank17]:[W908 11:34:01.421349990 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank20]:[W908 11:34:01.431165020 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank19]:[W908 11:34:01.431240250 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
[rank30]:[W908 11:34:01.431707980 compiler_depend.ts:335] Warning: Cannot create tensor with interal format while allow_internel_format=False, tensor will be created with base format. (function operator())
...
/root/MindSpeed/mindspeed/core/fp8_utils.py:11: UserWarning: Currently, it is not supported to Cast shard fp32 main params to fp8 model params
  warnings.warn("Currently, it is not supported to Cast shard fp32 main params to fp8 model params")
/root/MindSpeed/mindspeed/core/fp8_utils.py:11: UserWarning: Currently, it is not supported to Cast shard fp32 main params to fp8 model params
  warnings.warn("Currently, it is not supported to Cast shard fp32 main params to fp8 model params")
 [2025-09-08 11:37:00] iteration       10/    5000 | consumed samples:          640 | elapsed time per iteration (ms): 6932.5 | learning rate: 2.500000E-07 | global batch size:    64 | lm loss: 1.053084E+01 | loss scale: 1.0 | g      rad norm: 56.739 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 |
 [2025-09-08 11:37:03] iteration       11/    5000 | consumed samples:          704 | elapsed time per iteration (ms): 1981.0 | learning rate: 2.750000E-07 | global batch size:    64 | lm loss: 1.044677E+01 | loss scale: 1.0 | g      rad norm: 57.590 | num zeros: 0 | number of skipped iterations:   0 | number of nan iterations:   0 | 
...
```

**查看任务重调度记录<a name="section97707231547"></a>**

执行如下命令查看任务重调度记录。

```
kubectl describe cm -n mindx-dl job-reschedule-reason
```

回显示例如下。

```
Name:         job-reschedule-reason
Namespace:    mindx-dl
Labels:       <none>
Annotations:  <none>
Data
====
recent-reschedule-records:
----
{"default/default-test-pytorch-141274b7-ce93-4d31-adde-6c24456a8a3b":{"JobID":"default/default-test-pytorch-141274b7-ce93-4d31-adde-6c24456a8a3b","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0908 11:36:10","RescheduleTimeStamp":1759683370,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"default-test-pytorch-worker-0","NodeName":"node2","NodeRankIndex":"1"}]}]}}
Events:  <none>
```


#### 优雅容错模式<a name="ZH-CN_TOPIC_0000002511346479"></a>

本章节指导用户查看使用故障处理的优雅容错模式的训练信息。当芯片发生故障时，进程退出后进行优雅容错处理，恢复后重新拉起进程。

**日志说明<a name="section83075820188"></a>**

重新拉起的训练进程的训练日志在“_训练脚本路径_/newlog“中，具体说明如下。

-   QWEN3（PyTorch）训练日志：“/data/atlas\_dls/public/code/QWEN3\_for\_PyTorch\_2.7\_code/alllogs“。
-   QWEN3（MindSpore）训练日志：“/data/atlas\_dls/public/code/QWEN3\_for\_MS\_code/alllogs“。

**操作步骤<a name="section25042117188"></a>**

1.  登录管理节点，执行以下命令查看芯片情况。

    ```
    npu-smi info
    ```

    回显示例如下，此时表示训练进程占用片上内存，正常训练中。

    ![](../../figures/scheduling/1-13.png)

2.  故障发生后，执行以下命令查看芯片信息。

    ```
    npu-smi info
    ```

    回显示例如下，此时表示训练进程已退出，释放片上内存。

    ![](../../figures/scheduling/2.png)

3.  故障恢复后，执行以下命令查看芯片信息。

    ```
    npu-smi info
    ```

    回显示例如下，此时表示训练进程已重新拉起占用片上内存，正常训练中。

    ![](../../figures/scheduling/3.png)



### 删除任务<a name="ZH-CN_TOPIC_0000002479386566"></a>

**操作步骤<a name="section324819211118"></a>**

在下发任务的YAML目录执行以下命令，删除对应的训练任务。

```
kubectl delete -f XXX.yaml
```

示例如下：

```
kubectl delete -f pytorch_multinodes_acjob_910b.yaml
```

回显示例如下：

```
configmap "reset-config-default-test-pytorch" deleted
ascendjob.mindxdl.gitee.com "default-test-pytorch" deleted
```


### 运行维护<a name="ZH-CN_TOPIC_0000002479386520"></a>

**前提条件<a name="section18751194535314"></a>**

此功能只适用于特定场景下，用户需要使用重调度功能，且Ascend Device Plugin的启动YAML中已设置autoStowing参数为false。

**操作方法<a name="section8557331115714"></a>**

-   用户可以使用以下命令，将健康状态由unhealthy恢复为healthy的芯片重新放入资源池。

    ```
    kubectl label nodes node_name huawei.com/Ascend910-Recover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-Recover**”标签，该标签中的芯片会重新放入资源池中供程序调度。

    >[!NOTE] 说明 
    >该命令仅做清除Recover标签信息使用，请不要用于添加标签。

-   用户可以使用以下命令，将参数面网络健康状态由unhealthy恢复为healthy的芯片重新放入资源池。

    ```
    kubectl label nodes node_name huawei.com/Ascend910-NetworkRecover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-NetworkRecover**”标签，同时也会清除“**huawei.com/Ascend910-NetworkUnhealthy**”中对应的芯片。

    >[!NOTE] 说明 
    >该命令仅做清除NetworkRecover标签信息使用，请不要用于添加标签。



