# 故障处理原理

## 故障输入与触发流程

在故障检测完成后，针对每一种故障模式，断点续训通过故障处理或故障容错来恢复训练业务。断点续训特性根据恢复粒度由粗到细提供Job级别重调度、Pod级别重调度、进程级别重调度、弹性训练、进程级在线恢复、算子级在线恢复多层故障处理系统。用户可根据实际情况选择使用对应的子特性。

故障检测的详细原理和公共故障级别定义请参见[故障检测特性指南](../../11_fault_detection_and_diagnosis/01_working_principle.md)。

## 故障处理策略选择

**图 1**  故障决策说明<a name="fig2639326192019"></a>

![](../../../../figures/scheduling/故障决策说明.png "故障决策说明")

上图中，容错速度代表故障发生到故障恢复的速度，成功率代表故障发生后故障完成恢复的成功率，易用性代表用户使用或集成的成本。

Job级别重调度、Pod级别重调度、进程级别重调度可支持当前断点续训支持的全部故障模式，但依赖存在备份冗余计算服务器资源。如果存在不可修复的硬件故障且无备份冗余计算服务器时，可以通过配置弹性训练功能进行缩容训练。进程级在线恢复当前支持片上内存故障和网络故障。算子级在线恢复当前支持芯片网络故障和灵衢网络故障。

断点续训多层故障处理系统不同层级根据恢复粒度由细到粗可以逐级回退，如[图2](#fig477415371217)所示，如果上一层恢复失败则可以回退到下一层处理方式。

**图 2**  恢复失败说明<a name="fig477415371217"></a>

![](../../../../figures/scheduling/恢复失败说明.png "恢复失败说明")

### 重调度模式<a name="zh-cn_topic_0000002198051753_section1536115719358"></a>

#### 重调度级别

重调度模式：将任务调度到健康的芯片上，并隔离故障芯片。

重调度模式默认为**Job级别重调度**，每次故障会停止所有的Pod。但在大规模任务中，停止所有Pod后再重调度的成本较高，存在故障恢复时间过长的问题。除此之外断点续训还提供**Pod级别重调度**功能，用户可根据任务规模配置，在故障时刻只停止故障相关的Pod后重调度少量Pod，从而达成故障的快速恢复。为了进一步缩短故障恢复时间、降低故障影响范围，断点续训还提供进程级别重调度及进程级在线恢复功能。

**表 1**  各种重调度级别的差异

<a name="zh-cn_topic_0000002198051753_table18771108163419"></a>

|重调度的级别|恢复训练耗时|配置步骤|说明|
|--|--|--|--|
|Job级别重调度|Job级重调度的恢复时间较长，随着任务规模增加恢复时间超线性劣化。|<p>Job级重调度操作步骤简单，使用MindCluster的用户仅打开配置开关即可使用。</p><p>关键配置步骤请参见[配置Job级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置job级别重调度)。</p>|为了进一步降低恢复中资源调度时间，用户可以选择在Job级重调度上开启Pod级重调度能力。|
|Pod级别重调度|Pod级重调度可以将资源调度时间缩短，且与任务规模无关。但是，Pod级重调度并不能优化训练初始化过程中的时间开销，整体恢复时间仍然会随着任务规模增加而超线性劣化。|<p>Pod级重调度用户需要额外在训练容器中集成训练进程管理能力，使用MindCluster的用户具备对应进程管理能力后即可使用。</p><p>关键配置步骤请参见[配置Pod级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置pod级别重调度)。</p>|为了进一步降低训练初始化中的恢复时间，用户可以选择在Pod级重调度上开启进程级重调度能力。|
|进程级别重调度（进程级恢复）|进程级重调度可以减少训练初始化时间，将整体恢复时间缩短，且与任务规模无关或者弱相关。|<p>相比Pod级重调度，进程级重调度用户需要额外在训练框架中集成高可用训练能力，使用MindCluster的用户需要修改训练脚本，并开启对应配置开关后使用。</p><p>关键配置步骤请参见[配置进程级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级别重调度)。</p>|为了解决大规模场景下MTBF时间较短的问题，进一步降低整体恢复时间，用户可以选择在进程级重调度上开启进程级在线恢复能力。|
|进程级在线恢复|进程级在线恢复比起进程级重调度，恢复训练耗时更低。|<p>相比进程级重调度，进程级在线恢复用户需要配置对应的配置开关后使用。</p><p>关键配置步骤请参见[配置进程级在线恢复](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级在线恢复)。</p>|当前进程级在线恢复支持片上内存故障和网络故障，其余故障场景将回退其他处理方式。|
|算子级在线恢复|-|关键配置步骤请参见[配置算子级在线恢复](../03_configuration/01_configuring_fault_handling_policies.md#配置算子级在线恢复)。|-|

#### 重调度策略

重调度模式存在以下两种重调度策略。

- **直接重调度**：训练过程中发生集群调度组件可以探测到的硬件故障，系统将故障节点或芯片进行隔离，直接对任务进行重调度。
- **无条件重试**：训练过程中发生集群调度组件不能探测到的故障，导致任务容器异常退出，系统无条件对任务进行重调度。

**表 2**  重调度策略说明

<a name="zh-cn_topic_0000002198051753_table37727194382"></a>

|重调度策略|说明|支持的故障类型|
|--|--|--|
|直接重调度|系统将故障的节点或芯片进行隔离，然后直接对任务进行重调度。|已知的节点故障或重调度处理级别芯片故障。|
|无条件重试|<p>系统对配置了无条件重试次数的任务，进行指定次数内的重调度。</p><p>成功重调度后，任务可重试次数将减1，当可重试次数为0时无法再次触发重调度。</p><p>如需使用无条件重试功能，需在YAML中配置fault-retry-times参数，详细参数说明请参见[YAML配置说明](../../../06_api/15_yaml_configuration.md#yaml_configuration)。</p>|由于参数面网络故障或者训练相关软件故障等，导致任务异常退出，Pod的Status变为Failed状态的相关故障。|

### 故障级别与训练行为

故障级别及其对应的重调度处理、优雅容错处理说明，请参见[故障级别及处理说明](../../11_fault_detection_and_diagnosis/03_configuration/01_fault_classification.md#table103716651410)。

### 亚健康故障处理策略

各亚健康故障处理策略的当前任务行为、CKPT要求、后续动作和配置入口，请参见[配置亚健康故障处理策略](../03_configuration/01_configuring_fault_handling_policies.md#配置亚健康故障处理策略)。

## 重调度恢复

### Job级别重调度<a name="ZH-CN_TOPIC_0000002479226586"></a>

**Job级别重调度**即每次故障停止所有Pod，重新创建并重调度所有Pod后，重启训练任务。重调度模式默认为**Job级别重调度**。

了解Job级别重调度的关键配置步骤，请参见[配置Job级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置job级别重调度)。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

- 本功能仅支持在6.0.RC2及以上版本中使用。
- 大规模K8s集群场景下，ConfigMap映射时延不可控，建议RankTable使用共享存储方式。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 3**  Job级别重调度支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>

|产品类型|硬件形态|训练框架|
|--|--|--|
|Atlas 训练系列产品|<ul><li>Atlas 800 训练服务器（型号 9000）</li><li>Atlas 800 训练服务器（型号 9010）</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody">若Atlas 800 训练服务器的芯片工作模式为SMP模式，且每个Pod申请的NPU数量为1、2时，不支持使用重调度模式。查询和设置NPU芯片工作模式的详细介绍请参见《Atlas 800 训练服务器 iBMC用户指南（型号 9000）》中的“[查询和设置NPU芯片工作模式（npuworkmode）](https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a)”章节。</div></div>|<ul><li>MindSpore</li><li>PyTorch</li></ul>|
|Atlas A2 训练系列产品|<ul><li>Atlas 800T A2 训练服务器</li><li>Atlas 200T A2 Box16 异构子框</li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>|<ul><li>MindSpore</li><li>PyTorch</li></ul>|
|Atlas A3 训练系列产品|<ul><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 800T A3 超节点服务器</li><li>Atlas 9000 A3 SuperPoD 集群算力系统</li></ul>|<ul><li>MindSpore</li><li>PyTorch</li></ul>|
|A200T A3 Box8 超节点服务器|A200T A3 Box8 超节点服务器|<ul><li>MindSpore</li><li>PyTorch</li></ul>|
|Ascend 950 系列产品|<ul><li>Atlas 950 SuperPoD 超节点</li><li>Atlas 850E 超节点</li><li>Atlas 650E 服务器</li></ul>|<ul><li>PyTorch</li></ul>|

**重调度原理<a name="zh-cn_topic_0000002039194017_section57901137171110"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。Job级别重调度首先销毁所有的训练容器，然后隔离故障设备，再重新将训练容器调度启动。训练容器重新启动后重新拉起训练，该行为类似训练首次拉起过程。

**图 3**  原理图<a name="fig18343114924113"></a>

![](../../../../figures/scheduling/原理图.png "原理图")

在以上原理图中，各个步骤的说明如下。

1. 检测到故障后，首先删除当前任务所有的Pod和容器。
2. 隔离故障所在的设备，防止再次使用该设备。
3. 重新创建和调度训练Pod和容器。
4. 容器启动后，拉起训练进程恢复训练任务。

### Pod级别重调度<a name="ZH-CN_TOPIC_0000002511346429"></a>

**Pod级别重调度**即每次故障只停止故障相关的Pod，重新创建并重调度故障相关的Pod后，重启训练任务。如果当前故障不能恢复，则回退至Job级重调度模式。相比于Job级别重调度，Pod级别重调度会减少部分资源调度、Pod创建的时间。

了解Pod级别重调度的关键配置步骤，请参见[配置Pod级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置pod级别重调度)。

**使用约束<a name="zh-cn_topic_0000002003034876_section11983145119441"></a>**

- 在大集群训练任务中使用**Pod级别重调度**时，建议设置open files参数（可以打开的最大文件数目）足够大，设置过小可能导致Pod重调度出现异常。例如执行**ulimit -n 100000**命令，将open files参数设置为100000。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不触发Pod级别重调度和进程级别重调度，直接触发Job级别重调度。
- 请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002003034876_section48174410591"></a>**

**表 4**  重调度支持的产品和框架

<a name="zh-cn_topic_0000002003034876_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003034876_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002003034876_p199171819164417"><a name="zh-cn_topic_0000002003034876_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="63.10631063106311%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002003034876_p2917819114420"><a name="zh-cn_topic_0000002003034876_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="16.43164316431643%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002003034876_p27578257424"><a name="zh-cn_topic_0000002003034876_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003034876_row12917151994410"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p339114714459"><a name="zh-cn_topic_0000002003034876_p339114714459"></a><span id="zh-cn_topic_0000002003034876_ph327965117217"><a name="zh-cn_topic_0000002003034876_ph327965117217"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul17412295261"></a><ul id="zh-cn_topic_0000002003034876_ul17412295261"><li><span id="ph1179307345"><a name="ph1179307345"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span><div class="note" id="zh-cn_topic_0000002003034876_note186291241356"><a name="zh-cn_topic_0000002003034876_note186291241356"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003034876_p86294411854"><a name="zh-cn_topic_0000002003034876_p86294411854"></a>若<span id="zh-cn_topic_0000002003034876_ph1162924110518"><a name="zh-cn_topic_0000002003034876_ph1162924110518"></a>Atlas 800 训练服务器</span>的芯片工作模式为SMP模式，且每个Pod申请的NPU数量为1、2时，不支持使用重调度模式。查询和设置NPU芯片工作模式的详细介绍请参见<span id="zh-cn_topic_0000002003034876_ph66296417518"><a name="zh-cn_topic_0000002003034876_ph66296417518"></a>《Atlas 800 训练服务器 iBMC用户指南（型号 9000）》中的“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a" target="_blank" rel="noopener noreferrer">查询和设置NPU芯片工作模式（npuworkmode）</a>”</span>章节。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002003034876_ul353572894311"></a><ul id="zh-cn_topic_0000002003034876_ul353572894311"><li><span id="zh-cn_topic_0000002003034876_ph2075216585425"><a name="zh-cn_topic_0000002003034876_ph2075216585425"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002003034876_ph19355165113512"><a name="zh-cn_topic_0000002003034876_ph19355165113512"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002003034876_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p153913472453"><a name="zh-cn_topic_0000002003034876_p153913472453"></a><span id="zh-cn_topic_0000002003034876_ph151431757142112"><a name="zh-cn_topic_0000002003034876_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p15647160165615"><a name="p15647160165615"></a></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul1843217118563"></a><ul id="zh-cn_topic_0000002003034876_ul1843217118563"><li><span id="ph2153181425619"><a name="ph2153181425619"></a>Atlas 800T A2 训练服务器</span></li><li><span id="zh-cn_topic_0000002003034876_ph1114211211203"><a name="zh-cn_topic_0000002003034876_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002003034876_ph495114991519"><a name="zh-cn_topic_0000002003034876_ph495114991519"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002003034876_ul693112434815"></a><ul id="zh-cn_topic_0000002003034876_ul693112434815"><li><span id="zh-cn_topic_0000002003034876_ph1393112494820"><a name="zh-cn_topic_0000002003034876_ph1393112494820"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002003034876_ph2093210246488"><a name="zh-cn_topic_0000002003034876_ph2093210246488"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002003034876_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p18222246142212"><a name="zh-cn_topic_0000002003034876_p18222246142212"></a><span id="zh-cn_topic_0000002003034876_ph18411121792018"><a name="zh-cn_topic_0000002003034876_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul1367372444211"></a><ul id="zh-cn_topic_0000002003034876_ul1367372444211"><li><p id="p14426829306"><a name="p14426829306"></a><span id="ph077885871817"><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</li><li><span id="ph10355115144111"><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li><li><span id="ph9000a3superpod"><a name="ph9000a3superpod"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><ul id="zh-cn_topic_0000002039194017_ul7201511105411"><li><span id="zh-cn_topic_0000002039194017_ph52034113546"><a name="zh-cn_topic_0000002039194017_ph52034113546"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039194017_ph620418118547"><a name="zh-cn_topic_0000002039194017_ph620418118547"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row999211122017"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="p09912115201"><a name="p09912115201"></a><span id="ph126247155413"><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><p id="p49961172020"><a name="p49961172020"></a><span id="ph6124114710214"><a name="ph6124114710214"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="ul5581185452113"></a><ul id="ul5581185452113"><li><span id="ph19581195472117"><a name="ph19581195472117"></a>MindSpore</span></li><li><span id="ph8581154132114"><a name="ph8581154132114"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row_ascend950_pod"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="p_ascend950_pod"><a name="p_ascend950_pod"></a><span id="ph_ascend950_pod"><a name="ph_ascend950_pod"></a>Ascend 950 系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><ul id="ul_ascend950_pod"><li><span id="ph_ascend950_superpod_pod"><a name="ph_ascend950_superpod_pod"></a>Atlas 950 SuperPoD 超节点</span></li><li><span id="ph_ascend850_hardware_pod"><a name="ph_ascend850_hardware_pod"></a>Atlas 850E 超节点</span></li><li><span id="ph_ascend850e_hardware_pod"><a name="ph_ascend850e_hardware_pod"></a>Atlas 650E 服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><ul id="ul_ascend950_fw_pod"><li><span id="ph_ascend950_pt_pod"><a name="ph_ascend950_pt_pod"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002003034876_section19557184814234"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。Pod级别重调度首先销毁当前任务中故障的Pod和容器，并通知其他训练容器中的管理进程销毁所有训练进程，然后隔离故障设备，再重新将训练容器调度启动。待训练容器重新启动后，通知所有容器中的管理进程重新拉起训练进程恢复训练。

1. 检测到故障后，仅删除当前任务中故障的Pod和容器，销毁所有训练进程。
2. 隔离故障所在的设备，防止再次使用该设备。
3. 重新创建和调度训练Pod和容器。
4. 容器启动后，拉起训练进程恢复训练。

### 进程级别重调度<a name="ZH-CN_TOPIC_0000002511346457"></a>

进程级别重调度即每次故障只停止故障相关节点的进程，根据配置策略判断是否退出故障节点。

- recover策略：将故障节点的容器迁移到健康节点；
- recover-in-place策略：对于发生以下两类故障的节点，仅重启故障进程，不迁移故障节点的容器。若多个节点同时发生故障，则只发生以下两类故障的节点仅重启故障进程，不迁移容器，发生其他类型故障的节点会迁移容器。若多个节点发生故障的类型只包含业务进程异常故障，则所有故障节点均会迁移容器。
    - 业务进程异常故障。
    - RestartRequest和RestartBusiness级别的芯片故障。

不能恢复则回退至Job级或Pod级重调度模式。相比于Pod级别重调度，本功能仅重调度故障进程，减少了大量进程间不同步的等待耗时。同时利用了新的HCCL建链方案大大降低了建链耗时，且通过NPU卡间的参数面高速网络P2P传递CKPT信息，避免了CKPT保存和加载的耗时。

了解进程级别重调度的关键配置步骤，请参见[配置进程级别重调度](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级别重调度)。

>[!NOTE]
>
>- 参数面传递CKPT信息依赖故障卡中的全量优化器副本，如果不存在全量优化器副本则回退为加载存储中的CKPT文件恢复参数。
>- 优化器副本依赖额外的显存占用，如果用户的显存较为紧张，可选择本地加载模式，无论是否存在优化器副本都直接加载存储中的CKPT文件恢复参数。

**使用约束<a name="zh-cn_topic_0000002039353153_section514611624316"></a>**

- 对于PyTorch训练框架，需配套MindSpeed版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配套MindFormers版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障，且需迁移容器时，不触发Pod级别重调度和进程级别重调度，直接触发Job级别重调度。
- 不能和优雅容错功能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
- MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
- MindSpore场景下，受框架机制限制，进程级重调度存在极小概率失败风险。
- 请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。
- PyTorch只支持单算子模式、基于Megatron框架的模型。
- 只支持acjob类型训练任务。
- 只支持单容器迁移，不支持按照亲和性迁移。
- 不支持多模态模型。
- 不支持开启watchdog功能。
- 不支持在保存Checkpoint期间触发进程级别重调度。
- Atlas A3 训练系列产品场景下，若发生NPU掉卡类、OS断连类的故障，可导致进程级别重调度失败。
- 当故障发生在HCCL建链阶段时，会导致进程级别重调度失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考[配置HCCL主动触发建链](../03_configuration/02_configuring_training_recovery.md#配置hccl主动触发建链)章节进行提前建链，防止故障出现在HCCL建链阶段。
- 本功能依赖MindIO组件，使用前请先了解MindIO的[约束限制](../../../07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md#约束限制)。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039353153_section136131584164"></a>**

**表 5**  重调度支持的产品和框架

<a name="zh-cn_topic_0000002039353153_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039353153_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039353153_p199171819164417"><a name="zh-cn_topic_0000002039353153_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="66.2966296629663%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039353153_p2917819114420"><a name="zh-cn_topic_0000002039353153_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="13.24132413241324%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039353153_p27578257424"><a name="zh-cn_topic_0000002039353153_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039353153_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p153913472453"><a name="zh-cn_topic_0000002039353153_p153913472453"></a><span id="zh-cn_topic_0000002039353153_ph151431757142112"><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p737515258512"><a name="p737515258512"></a></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039353153_ul1843217118563"></a><ul id="zh-cn_topic_0000002039353153_ul1843217118563"><li><p id="p1546725019404"><a name="p1546725019404"></a><span id="ph157633217501"><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></p>
</li><li><span id="zh-cn_topic_0000002039353153_ph1114211211203"><a name="zh-cn_topic_0000002039353153_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002039353153_ph495114991519"><a name="zh-cn_topic_0000002039353153_ph495114991519"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039353153_ul693112434815"></a><ul id="zh-cn_topic_0000002039353153_ul693112434815"><li><span id="zh-cn_topic_0000002039353153_ph1393112494820"><a name="zh-cn_topic_0000002039353153_ph1393112494820"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039353153_ph2093210246488"><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039353153_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p18222246142212"><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><span id="zh-cn_topic_0000002039353153_ph18411121792018"><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><a name="ul61561253231"></a><ul id="ul61561253231"><li><span id="ph077885871817"><a name="ph077885871817-duplicate-2"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111-duplicate-2"></a>Atlas 800T A3 超节点服务器</span></li><li><span id="ph9000a3superpod_process_resched"><a name="ph9000a3superpod_process_resched"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><a name="ul18946810161311"></a><ul id="ul18946810161311"><li><span id="ph99461100137"><a name="ph99461100137"></a>MindSpore</span><p id="p664545214"><a name="p664545214"></a><span id="ph294661010130"><a name="ph294661010130"></a></span></p>
</li><li><span id="ph99469109139"><a name="ph99469109139"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row_ascend950_process"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="p_ascend950_process"><a name="p_ascend950_process"></a><span id="ph_ascend950_process"><a name="ph_ascend950_process"></a>Ascend 950 系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><ul id="ul_ascend950_process"><li><span id="ph_ascend950_superpod_process"><a name="ph_ascend950_superpod_process"></a>Atlas 950 SuperPoD 超节点</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><ul id="ul_ascend950_process_fw"><li><span id="ph_ascend950_pt_process"><a name="ph_ascend950_pt_process"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002039353153_section12206164333619"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。进程级重调度根据配置策略首先销毁故障的训练进程或容器，并通知其他训练容器中的训练进程暂停当前训练任务，然后隔离故障设备，再重新将训练容器调度启动。故障训练容器重新启动后，通知所有容器中的训练进程进行集合通信重建链。建链完成后，将CKPT通过参数面发送给新拉起的训练进程恢复参数，恢复后所有进程重新执行当前step恢复训练。

**图 4**  进程级别重调度原理示意图<a name="fig1373016583373"></a>

![](../../../../figures/scheduling/进程级别重调度原理示意图.png "进程级别重调度原理示意图")

在以上原理图中，各个步骤的说明如下。

1. 设备出现硬件故障后，MindCluster在服务器上的检测组件上报故障信息到ClusterD中，软件故障由容器内MindIO Controller感知并上报到ClusterD。
2. ClusterD将故障服务器上的任务容器退出故障训练进程，重新调度到备用的服务器上。
3. ClusterD通知Master节点上的MindIO Controller进行容错，容错流程包括通知停止训练、通知全局故障、通知恢复策略。
4. MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用TorchNPU强制停止训练进程。MindIO Processor清理正常节点的资源，销毁通信域，清理后等待新进程加入。
5. 备用服务器上的管理进程拉起训练进程后，创建新的MindIO Processor，MindIO Controller通知每个训练进程中的MindIO Processor恢复训练。
6. 各个进程进行集合通信建链。
7. 正常服务器上的NPU通过参数面将CKPT传递到备用服务器上，完成参数状态恢复后继续训练。

**功能适配点<a name="section1446615300284"></a>**

在进程级别重调度中，集群大脑会根据全局故障信息决策恢复策略并将策略下发到MindIO，调度器需要支持故障Pod调度，而非整个任务重调度，支持恢复策略依次回退。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，在恢复时执行算子资源清理，节点重启后触发通信重建。通过参数面在线修复和状态回滚，完成进程级重调度恢复。

对于非MindSpeed-LLM和MindCluster平台用户，需在框架侧完成[表6](#table1995514113610)的功能适配。

**表 6**  进程级别重调度框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="16.77167716771677%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="43.23432343234324%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.13181318131813%" id="mcps1.2.5.1.3"><p id="p104301715185316"><a name="p104301715185316"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="21.862186218621858%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="16.77167716771677%" headers="mcps1.2.5.1.1 "><p id="p18221046175418"><a name="p18221046175418"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="43.23432343234324%" headers="mcps1.2.5.1.2 "><p id="p14221746205412"><a name="p14221746205412"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="18.13181318131813%" headers="mcps1.2.5.1.3 "><p id="p5119132211596"><a name="p5119132211596"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="21.862186218621858%" headers="mcps1.2.5.1.4 "><p id="p252632095917"><a name="p252632095917"></a><a href="../../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1122104645414"><a name="p1122104645414"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p322446155419"><a name="p322446155419"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p152294645418"><a name="p152294645418"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p6221046125412"><a name="p6221046125412"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row144014113397"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p52294618541"><a name="p52294618541"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p72284615410"><a name="p72284615410"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row74014111391"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p522194614547"><a name="p522194614547"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2221346125415"><a name="p2221346125415"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row74025111392"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9229467542"><a name="p9229467542"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p102218466541"><a name="p102218466541"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row19531411367"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p422846105412"><a name="p422846105412"></a>节点重启及通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p922194612545"><a name="p922194612545"></a>通过注册重建回调实现健康节点与故障节点重建通信域。</p>
</td>
</tr>
<tr id="row1708112845416"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p722164611549"><a name="p722164611549"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1122246145417"><a name="p1122246145417"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row1911610240547"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p92214463547"><a name="p92214463547"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p222154613543"><a name="p222154613543"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row1311652445414"><td class="cellrowborder" valign="top" width="16.77167716771677%" headers="mcps1.2.5.1.1 "><p id="p202220467541"><a name="p202220467541"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="43.23432343234324%" headers="mcps1.2.5.1.2 "><p id="p1022184612549"><a name="p1022184612549"></a>根据全局故障信息决策恢复策略，并下发到MindIO，支持恢复策略回退，进程级重调度失败回退到Pod级别、Job级别重调度。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="18.13181318131813%" headers="mcps1.2.5.1.3 "><p id="p488619172591"><a name="p488619172591"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="21.862186218621858%" headers="mcps1.2.5.1.4 "><p id="p1211652412545"><a name="p1211652412545"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row18952145365"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p72274605415"><a name="p72274605415"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p522104615410"><a name="p522104615410"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p11417425315"><a name="p11417425315"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

### 亚健康退出与重调度

- graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。
- forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。

使用graceExit策略时，需保证任务开启了临终CKPT保存功能。

### 亚健康热切<a name="ZH-CN_TOPIC_0000002479386544"></a>

训练任务配置为亚健康热切策略（hotSwitch）后，当发生亚健康故障时，拉起备份节点后暂停训练进程，再使用备份节点重新拉起训练任务。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127-duplicate-2"></a>**

- 对于PyTorch训练框架，需配合MindSpeed-LLM 2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配合MindFormers master版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 只支持PyTorch单算子模式、基于Megatron框架的模型以及acjob类型训练任务。
- MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
- 不支持多模态模型。
- 不支持开启watchdog功能。
- 训练任务未出迭代时触发热切，可能会造成MindIO阻塞，最后触发Job级别重调度。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生亚健康故障时，不支持触发亚健康热切。
- 以下异常情况会回退至Job级别重调度，且任务亚健康处理策略降级为ignore，不再处理亚健康故障：
    - 备份Pod拉起后，训练暂停失败。
    - 备份Pod拉起后，MindCluster等待上报训练暂停状态超时（15分钟）。
    - 备份Pod运行失败。
    - 原Pod删除后，训练恢复失败。
    - 原Pod删除后，MindCluster等待上报训练恢复状态超时（15分钟）。

- 配置亚健康热切策略后，会自动增加进程级恢复开关，若发生非亚健康故障，将触发进程级恢复流程。
- 无备节点场景下，无法完成热切流程，任务亚健康处理策略降级为ignore，不再处理亚健康故障。
- 本功能依赖MindIO组件，使用前请先了解MindIO的[约束限制](../../../07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md#约束限制)。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318-duplicate-2"></a>**

**表 14**  亚健康热切支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="25.172517251725175%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="59.82598259825983%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="15.001500150015001%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="25.172517251725175%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p773278122616"><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="59.82598259825983%" headers="mcps1.2.4.1.2 "><p id="p3799611168"><a name="p3799611168"></a><span id="ph14314162316427"><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.001500150015001%" headers="mcps1.2.4.1.3 "><a name="ul15879359132214"></a><ul id="ul15879359132214"><li><span id="ph135835207394"><a name="ph135835207394-duplicate-2"></a>MindSpore</span></li><li><span id="ph19425111582712"><a name="ph19425111582712"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="25.172517251725175%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="ph531432344210"><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="59.82598259825983%" headers="mcps1.2.4.1.2 "><a name="ul9000a3superpod_hotswitch"></a><ul id="ul9000a3superpod_hotswitch"><li><span id="ph10355115144111"><a name="ph10355115144111-duplicate-5"></a>Atlas 800T A3 超节点服务器</span></li><li><span id="ph9000a3superpod_hotswitch"><a name="ph9000a3superpod_hotswitch"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="15.001500150015001%" headers="mcps1.2.4.1.3 "><a name="ul6531100274"></a><ul id="ul6531100274"><li><span id="ph1053216019718"><a name="ph1053216019718"></a>MindSpore</span></li><li><span id="ph35321001570"><a name="ph35321001570"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**亚健康热切原理<a name="zh-cn_topic_0000002039194017_section57901137171110-duplicate-2"></a>**

**图 9**  原理图<a name="fig1770171514241"></a>

![](../../../../figures/scheduling/原理图-11.png "原理图-11")

在以上原理图中，各个步骤的说明如下。

1. ClusterD通过Ascend Device Plugin感知到亚健康故障。
2. ClusterD根据配置策略决策是否进行亚健康热切恢复。
3. ClusterD通知Ascend Operator拉起备份Pod。
4. Volcano调度备份Pod。
5. 备份Pod中创建新的MindIO Processor，MindIO Processor向MindIO Controller发起注册。
6. MindIO Controller下发训练暂停通知。
7. MindIO Controller通知ClusterD训练暂停。
8. ClusterD通知Volcano删除故障Pod。
9. ClusterD通知MindIO恢复训练。

**适配功能点<a name="section1446615300284-duplicate-4"></a>**

在亚健康热切中，集群大脑根据亚健康故障信息，为故障Pod设置注解，拉起并调度备份Pod，通知热切策略到MindIO，训练切换到备份Pod后恢复训练。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。在异常发生时，通过异常捕获装饰器捕获故障模式。在新节点启动后，正常节点暂停训练，之后重建通信域，完成新节点参数面恢复，训练状态完成后完成节点热切换。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表15](#table19955141136104)的功能适配。

**表 15**  亚健康热切框架适配功能点

<a name="table19955141136104"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.200000000000003%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387-duplicate-4"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="39.330000000000005%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816-duplicate-4"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="19.670000000000005%" id="mcps1.2.5.1.3"><p id="p237216122367"><a name="p237216122367"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="22.800000000000004%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812-duplicate-4"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.200000000000003%" headers="mcps1.2.5.1.1 "><p id="p1698525618364"><a name="p1698525618364"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="39.330000000000005%" headers="mcps1.2.5.1.2 "><p id="p117503011375"><a name="p117503011375"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="19.670000000000005%" headers="mcps1.2.5.1.3 "><p id="p444112643720"><a name="p444112643720"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="22.800000000000004%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212-duplicate-3"></a><a href="../../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1598625612366"><a name="p1598625612366"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2986125603612"><a name="p2986125603612"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1798635633611"><a name="p1798635633611"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p898615619362"><a name="p898615619362"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6986115693618"><a name="p6986115693618"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p39861056113613"><a name="p39861056113613"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p99861756103611"><a name="p99861756103611"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p19986125623612"><a name="p19986125623612"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p698615618367"><a name="p698615618367"></a>节点重启及通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11986856193616"><a name="p11986856193616"></a>通过注册重建回调实现健康节点与故障节点重建通信域。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10986256103613"><a name="p10986256103613"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p198635693613"><a name="p198635693613"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row876023415918"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p49861556183618"><a name="p49861556183618"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1398655643610"><a name="p1398655643610"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row17605341596"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11986656153611"><a name="p11986656153611"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p199862056113617"><a name="p199862056113617"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row144412445361"><td class="cellrowborder" valign="top" width="18.200000000000003%" headers="mcps1.2.5.1.1 "><p id="p129861056183611"><a name="p129861056183611"></a>热切流程控制</p>
</td>
<td class="cellrowborder" valign="top" width="39.330000000000005%" headers="mcps1.2.5.1.2 "><p id="p14986125653610"><a name="p14986125653610"></a>管理热切恢复流程，通过设置注解方式管理备份Pod和故障Pod。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="19.670000000000005%" headers="mcps1.2.5.1.3 "><p id="p1045122693710"><a name="p1045122693710"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="22.800000000000004%" headers="mcps1.2.5.1.4 "><p id="p64451744113612"><a name="p64451744113612"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.1.0/component/clusterd/pkg/application/recover/hot_switch_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row14716101112393"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1371681114396"><a name="p1371681114396"></a>Pod创建删除</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p071681117390"><a name="p071681117390"></a>通过识别特定注解删除和创建Pod。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p071621117393"><a name="p071621117393"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.1.0/component/ascend-operator/pkg/controllers/v1/ascendjob_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

## 在线恢复

### 算子级在线恢复<a name="ZH-CN_TOPIC_0000002479386484"></a>

Atlas A3 训练系列产品支持在发生参数面网络故障时，HCCL会执行通信算子重传。在故障进程不退出的情况下，算子级在线恢复可容忍更长时间的网络异常，训练任务不中断。

若网络故障的算子级在线恢复（HCCL通信算子重执行）执行失败，则回退至进程级在线恢复

了解算子级在线恢复的关键配置步骤，请参见[配置算子级在线恢复](../03_configuration/01_configuring_fault_handling_policies.md#配置算子级在线恢复)。

>[!NOTE]
>HCCL（Huawei Collective Communication Library，华为集合通信库）是华为专为昇腾（Ascend）AI处理器设计的分布式通信库，旨在优化多设备（如NPU/GPU）间的高效协作，以加速深度学习模型的分布式训练，适用于需要大规模算力的AI场景。在分布式训练中，HCCL负责协调多个昇腾处理器之间的数据同步（如梯度聚合、参数更新），减少通信开销，提升训练效率。

**使用场景<a name="section4314241154917"></a>**

当前支持在以下2种故障场景下使用算子级在线恢复功能。

- 对于芯片网络相关故障，当算子重传成功时，Volcano会将任务作为亚健康任务处理。当算子重传失败时，Volcano触发重调度处理。
- 对于灵衢总线设备相关故障，HCCL执行算子级在线恢复后，Volcano会将任务作为亚健康任务处理。

**使用约束<a name="section1915719315116"></a>**

- 本特性不支持MC2开启场景。
- 不支持开启watchdog功能。

**算子级在线恢复支持的产品和框架<a name="section996215473410"></a>**

**表 7**  支持的产品和框架

<a name="table11647101624213"></a>
<table><thead align="left"><tr id="row17647111614214"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="p1664831610428"><a name="p1664831610428"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="p1664816167422"><a name="p1664816167422"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="p17648141664214"><a name="p17648141664214"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="row14649101615422"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p8649101644216"><a name="p8649101644216"></a><span id="ph96491216144210"><a name="ph96491216144210"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><ul id="ul_op_online_product_list"><li><span id="ph264911612426"><a name="ph264911612426"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph9000a3superpod_op_online_name"><a name="ph9000a3superpod_op_online_name"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><p id="p1664981614218"><a name="p1664981614218"></a>-</p>
</td>
</tr>
</tbody>
</table>

**算子级在线恢复原理<a name="section41453583611"></a>**

**图 5**  原理图<a name="fig151851746103612"></a>

![](../../../../figures/scheduling/原理图-8.png "原理图-8")

在以上原理图中，各个步骤的说明如下。

1. 训练过程中，发生HCCS网络平面LinkDown故障或RoCE网络平面LinkDown故障。
2. CANN检测到网络故障，当前算子终止后，进行网络链路恢复（HCCS网络平面进行BGP切路，RoCE网络平面进行借轨通信），通信链路恢复后进行网络算子重执行。
3. 算子重执行成功后，恢复训练迭代。

### 进程级在线恢复<a name="ZH-CN_TOPIC_0000002479386460"></a>

进程级在线恢复（Step级别重计算恢复）主要针对以下2种故障类型进行故障处理：

- 网络故障：当前仅支持以下三种场景。
    - HCCS L1-L2端口或链路故障时，BGP切路后，若开启算子级在线恢复且执行失败后进行Step级重试，实现进程不退出的故障快速恢复；若关闭算子级在线恢复，则对训练进程进行Step级重试，实现进程不退出的故障快速恢复。
    - RoCE到上级端口或链路故障，且开启算子级在线恢复并执行失败时，对训练进程进行Step级重试，实现进程不退出的故障快速恢复。
    - Atlas 950 SuperPoD超节点场景下，unions到上级unions/5808的UB端口或链路发生故障时，对训练进程进行Step级重试，实现进程不退出的故障快速恢复。

- 片上内存故障：片上内存上出现的不可纠正错误（如故障码0x80E01801），先隔离故障片上内存空间，然后对训练进程进行Step级重试，实现进程不退出的故障快速恢复。

在以上2种场景下，如果故障不能恢复，则回退至**重调度模式**。

相比于进程级别重调度，进程级在线恢复不会重调度故障进程，减少了大量进程间不同步的等待耗时。同时通过NPU卡间的参数面高速网络P2P传递CKPT信息，避免了CKPT保存和加载的耗时。

该故障处理模式默认关闭，若要开启请参见[（可选）配置组件](../04_examples_and_verification/menu_examples_and_verification.md#ZH-CN_TOPIC_0000002511346449)。

了解进程级在线恢复的关键配置步骤，请参见[配置进程级在线恢复](../03_configuration/01_configuring_fault_handling_policies.md#配置进程级在线恢复)。

>[!NOTE]
>
>- 参数面传递CKPT信息依赖未故障卡中的全量优化器副本，如果不存在全量优化器副本，则回退为加载存储中的CKPT文件恢复参数。
>- 优化器副本依赖额外的显存占用，如果用户的显存较为紧张，可选择本地加载模式，无论是否存在优化器副本都直接加载存储中的CKPT文件恢复参数。

**使用约束<a name="zh-cn_topic_0000002003193196_section17145122992213"></a>**

- 对于PyTorch训练框架，需配套MindSpeed版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配套MindFormers版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 依赖于PyTorch的内存管理机制，仅在PYTORCH\_NO\_NPU\_MEMORY\_CACHING未配置时才能使用此功能。
- 针对部分片上内存故障场景无法生效，例如HCCL集合通信使用的内存地址故障，仍需通过进程级重调度或更上层的容错方案恢复。
- 针对MindSpeed-LLM、MindSpeed等模型或训练脚本中定义的全局变量发生故障的场景，详细处理策略请参见[FAQ](https://gitcode.com/Ascend/mind-cluster/issues/368)。
- 与优雅容错不能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
- MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
- MindSpore场景下，需要在启动TaskD Manager前设置export TASKD\_PROCESS\_ENABLE="on"。
- 请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。
- 不支持多模态模型。
- 不支持MC2开启场景。
- 不支持开启watchdog功能。
- 当故障发生在HCCL建链阶段时，会导致进程级在线恢复失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考[配置HCCL主动触发建链](../03_configuration/02_configuring_training_recovery.md#配置hccl主动触发建链)章节进行提前建链，防止故障出现在HCCL建链阶段。
- 本功能依赖MindIO组件，使用前请先了解MindIO的[约束限制](../../../07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md#约束限制)。

**支持的产品型号及AI框架<a name="zh-cn_topic_0000002003193196_section108582044132214"></a>**

**表 8**  网络故障进程级在线恢复支持的产品和框架

<a name="zh-cn_topic_0000002003193196_table18104314924"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003193196_row81042144212"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002003193196_p51041814022"><a name="zh-cn_topic_0000002003193196_p51041814022"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002003193196_p91041414627"><a name="zh-cn_topic_0000002003193196_p91041414627"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002003193196_p11040145218"><a name="zh-cn_topic_0000002003193196_p11040145218"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003193196_row1910518141229"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003193196_p191051114524"><a name="zh-cn_topic_0000002003193196_p191051114524"></a><span id="zh-cn_topic_0000002003193196_ph19105814420"><a name="zh-cn_topic_0000002003193196_ph19105814420"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul18927338231"></a><ul id="ul18927338231"><li><span id="ph077885871817"><a name="ph077885871817-duplicate-3"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111-duplicate-3"></a>Atlas 800T A3 超节点服务器</span></li><li><span id="ph9000a3superpod_net_online"><a name="ph9000a3superpod_net_online"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul17506112910131"></a><ul id="ul17506112910131"><li><span id="ph135064298139"><a name="ph135064298139"></a>MindSpore</span></li></ul>
<a name="ul7506132918139"></a><ul id="ul7506132918139"><li><span id="ph550610294136"><a name="ph550610294136"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row_ascend950_network"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p_ascend950_network"><a name="p_ascend950_network"></a><span id="ph_ascend950_network"><a name="ph_ascend950_network"></a>Ascend 950 系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><ul id="ul_ascend950_network"><li><span id="ph_ascend950_superpod_network"><a name="ph_ascend950_superpod_network"></a>Atlas 950 SuperPoD 超节点</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><ul id="ul_ascend950_network_framework"><li><span id="ph_ascend950_pytorch_network"><a name="ph_ascend950_pytorch_network"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**表 9** 片上内存故障进程级在线恢复支持的产品和框架

<a name="table0630917154413"></a>
<table><thead align="left"><tr id="row13630161784418"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="p963031734417"><a name="p963031734417"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="p663151714415"><a name="p663151714415"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="p13631111710444"><a name="p13631111710444"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="row5631517114410"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p166312178442"><a name="p166312178442"></a><span id="ph1463121734416"><a name="ph1463121734416"></a>Atlas A2 训练系列产品</span></p>
<p id="p12631191713449"><a name="p12631191713449"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul0631181774417"></a><ul id="ul0631181774417"><li><span id="ph46319177449"><a name="ph46319177449"></a>Atlas 800T A2 训练服务器</span></li><li><span id="ph1463131724413"><a name="ph1463131724413"></a>Atlas 900 A2 PoD 集群基础单元</span></li><li><span id="ph46311417154417"><a name="ph46311417154417"></a>Atlas 900 A2 PoDc 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul3631151714415"></a><ul id="ul3631151714415"><li><span id="ph36311817154419"><a name="ph36311817154419"></a>MindSpore</span></li></ul>
<a name="ul1263181794418"></a><ul id="ul1263181794418"><li><span id="ph1263191704413"><a name="ph1263191704413"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row16631181714416"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p563111714440"><a name="p563111714440"></a><span id="ph363111714444"><a name="ph363111714444"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul1763161764415"></a><ul id="ul1763161764415"><li><span id="ph1963121720449"><a name="ph1963121720449"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph1363115172443"><a name="ph1363115172443"></a>Atlas 800T A3 超节点服务器</span></li><li><span id="ph9000a3superpod_uce_online"><a name="ph9000a3superpod_uce_online"></a>Atlas 9000 A3 SuperPoD 集群算力系统</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul96311517144415"></a><ul id="ul96311517144415"><li><span id="ph96310177449"><a name="ph96310177449"></a>MindSpore</span></li></ul>
<a name="ul7631141712447"></a><ul id="ul7631141712447"><li><span id="ph1563101734413"><a name="ph1563101734413"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row_ascend950_hbm"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p_ascend950_hbm"><a name="p_ascend950_hbm"></a><span id="ph_ascend950_hbm"><a name="ph_ascend950_hbm"></a>Ascend 950 系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><ul id="ul_ascend950_hbm"><li><span id="ph_ascend950_superpod_hbm"><a name="ph_ascend950_superpod_hbm"></a>Atlas 950 SuperPoD 超节点</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><ul id="ul_ascend950_hbm_fw2"><li><span id="ph_ascend950_pt_hbm"><a name="ph_ascend950_pt_hbm"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**进程级在线恢复原理<a name="zh-cn_topic_0000002003193196_section961210366427"></a>**

训练过程中如果出现了片上内存故障或网络故障，将导致训练状态异常。进程级在线恢复首先通知所有训练进程停止当前训练，然后保留当前训练信息并修复故障。修复完成后，所有训练进程回退训练状态到当前上一个Step结束时，正常服务器通过参数面将CKPT传递到故障服务器上，完成参数恢复后重新执行当前Step，然后恢复训练任务。

**图 6**  进程级在线恢复原理<a name="fig37536398327"></a>

![](../../../../figures/scheduling/进程级在线恢复原理.png "进程级在线恢复原理")

在以上原理图中，各个步骤的说明如下。

1. 设备出现片上内存故障或网络故障后，MindCluster在服务器上的检测组件上报故障信息到集群大脑ClusterD中。
2. 片上内存故障或网络故障被CANN软件感知，经训练框架上报给MindIO Processor和MindIO Controller。
3. MindIO Controller向集群大脑请求决策是否进行Step级别重计算恢复，集群大脑综合集群其他节点的健康状态给出决策。
4. MindIO Controller通知每个训练进程中的MindIO Processor，调用训练框架停止任务、修复故障，保留通信域信息。
5. 正常服务器上的NPU通过参数面将CKPT传递到故障（已修复）服务器上，完成参数状态恢复后继续训练，重新启动当前Step计算。

**适配功能点<a name="section1446615300284-duplicate-2"></a>**

在进程级在线恢复中，集群大脑根据故障信息识别网络故障和片上内存故障，下发对应恢复策略，支持恢复策略回退。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，在恢复时针对不同故障执行算子资源清理、UCE模型优化器重建、参数面在线修复、状态回滚，完成进程级在线恢复。

对于非MindSpeed-LLM、MindCluster平台用户，针对不同故障需在框架侧完成以下功能适配。

**表 10**  进程级在线恢复针对网络故障框架适配功能点

<a name="table19955141136101"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.61186118611861%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387-duplicate-2"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="36.72367236723672%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816-duplicate-2"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="17.981798179817982%" id="mcps1.2.5.1.3"><p id="p1912785111610"><a name="p1912785111610"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="26.68266826682668%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812-duplicate-2"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row199614191876"><td class="cellrowborder" valign="top" width="18.61186118611861%" headers="mcps1.2.5.1.1 "><p id="p174797321974"><a name="p174797321974"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="36.72367236723672%" headers="mcps1.2.5.1.2 "><p id="p1847910326710"><a name="p1847910326710"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="17.981798179817982%" headers="mcps1.2.5.1.3 "><p id="p12303135518715"><a name="p12303135518715"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="26.68266826682668%" headers="mcps1.2.5.1.4 "><p id="p1878873515913"><a name="p1878873515913"></a><a href="../../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row149661916713"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1947943212711"><a name="p1947943212711"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p04796323710"><a name="p04796323710"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row1239411299541"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p94791332771"><a name="p94791332771"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p5479103211720"><a name="p5479103211720"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row13395629115418"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17479113217716"><a name="p17479113217716"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p10479532172"><a name="p10479532172"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row7395142913549"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1447912327718"><a name="p1447912327718"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1847916321477"><a name="p1847916321477"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row539519296541"><td class="cellrowborder" valign="top" width="18.61186118611861%" headers="mcps1.2.5.1.1 "><p id="p114808324711"><a name="p114808324711"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="36.72367236723672%" headers="mcps1.2.5.1.2 "><p id="p248011324715"><a name="p248011324715"></a>根据故障信息识别网络故障或片上内存故障，下发对应恢复策略，支持恢复策略回退。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="17.981798179817982%" headers="mcps1.2.5.1.3 "><p id="p16303135517718"><a name="p16303135517718"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="26.68266826682668%" headers="mcps1.2.5.1.4 "><p id="p19472244965"><a name="p19472244965"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row7396029145419"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9480632573"><a name="p9480632573"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p84806321578"><a name="p84806321578"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p12472134412615"><a name="p12472134412615"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

**表 11**  进程级在线恢复针对片上内存故障框架适配功能点

<a name="table14662336155516"></a>
<table><thead align="left"><tr id="row866213619553"><th class="cellrowborder" valign="top" width="17.119999999999997%" id="mcps1.2.5.1.1"><p id="p36629367550"><a name="p36629367550"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="38.769999999999996%" id="mcps1.2.5.1.2"><p id="p6662103635520"><a name="p6662103635520"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="17.43%" id="mcps1.2.5.1.3"><p id="p1857674501116"><a name="p1857674501116"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="26.68%" id="mcps1.2.5.1.4"><p id="p966243617552"><a name="p966243617552"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row19662436145518"><td class="cellrowborder" valign="top" width="17.119999999999997%" headers="mcps1.2.5.1.1 "><p id="p339173741211"><a name="p339173741211"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="38.769999999999996%" headers="mcps1.2.5.1.2 "><p id="p1739537151219"><a name="p1739537151219"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="17.43%" headers="mcps1.2.5.1.3 "><p id="p9527145711216"><a name="p9527145711216"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="26.68%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a href="../../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row566215364551"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p23923716123"><a name="p23923716123"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11397374123"><a name="p11397374123"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row06621936185512"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p639183714129"><a name="p639183714129"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15391437141215"><a name="p15391437141215"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row2662133617558"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p73913781212"><a name="p73913781212"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1039113741210"><a name="p1039113741210"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row066213685511"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p143953791220"><a name="p143953791220"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p203910376122"><a name="p203910376122"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row666243613555"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1396379125"><a name="p1396379125"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p143993761219"><a name="p143993761219"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row14662143645516"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p639113711122"><a name="p639113711122"></a>UCE模型优化器重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p83953718128"><a name="p83953718128"></a>通过回调函数完成故障卡模型优化器对象操作清理、重建操作。</p>
</td>
</tr>
<tr id="row43068171121"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p139737201218"><a name="p139737201218"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p539103711214"><a name="p539103711214"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row17307161715127"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p183923781214"><a name="p183923781214"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p114017372126"><a name="p114017372126"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row966233613558"><td class="cellrowborder" valign="top" width="17.119999999999997%" headers="mcps1.2.5.1.1 "><p id="p114023721211"><a name="p114023721211"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="38.769999999999996%" headers="mcps1.2.5.1.2 "><p id="p124083761213"><a name="p124083761213"></a>根据故障信息识别网络故障或片上内存故障，下发对应恢复策略，支持恢复策略回退。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="17.43%" headers="mcps1.2.5.1.3 "><p id="p65272572124"><a name="p65272572124"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="26.68%" headers="mcps1.2.5.1.4 "><p id="p14571125414116"><a name="p14571125414116"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row16621936105516"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17407378128"><a name="p17407378128"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p140173701218"><a name="p140173701218"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p957195451114"><a name="p957195451114"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

## 切换恢复

### 借轨通信任务暂停与回切<a name="ZH-CN_TOPIC_0000002479226530"></a>

Atlas A3 训练系列产品场景下，MindCluster集群调度组件提供训练任务借轨通信的暂停与回切功能。即在训练过程中，使用主动借轨回切接口，可自由切换NPU芯片使用的RoCE网口。

使用借轨回切功能时，NPU芯片的组网关系可参考《Ascend Training Solution 组网指南（Atlas A3训练产品）》中的“网络平面介绍 \> 参数面网络 \> [端口对接策略](https://support.huawei.com/enterprise/zh/doc/EDOC1100570090/3e6a1479)”章节。

了解借轨通信任务暂停与回切功能的详细配置方法，请参见[配置借轨通信任务暂停与回切](../03_configuration/01_configuring_fault_handling_policies.md#配置借轨通信任务暂停与回切)。

- 调用[借轨回切接口](../../../06_api/04_clusterd/08_link_failover_and_switchback_apis.md)执行借轨回切动作前，请先了解NPU芯片组网关系，保证目标NPU的网络链路正常，如果目标NPU为linkdown状态会导致操作失败。
- 以上述组网指南中的接口对接关系为例，对于以下几种情况，调用SwitchNicTrack接口时，指定的dev与op如下：
    1. 若将device0，device8从QDD8借轨切到QDD7，传参dev为\[device0，device8\]，op为\[true，true\]
    2. 若将device0，device8从QDD7回切到QDD8，传参dev为\[device0，device8\]，op为\[false，false\]
    3. 如果单独将device0从QDD8的PortA借轨切到QDD7的PortA，传参dev为\[device0\]，op为\[true\]
    4. 如果单独将device0从QDD7的PortA回切到QDD8的PortA，传参dev为\[device0\]，op为\[false\]
    5. 如果将Leaf1下的全部device借轨切到Leaf2下，传参dev为\[device0，device8，device2，device10，device4，device12，device6，device14 \]，op为\[true，true，true，true，true，true，true，true\]
    6. 如果将Leaf2下的全部device回切到Leaf1下，传参dev为\[device0，device8，device2，device10，device4，device12，device6，device14 \]，op为\[false，false，false，false，false，false，false，false\]

    **图 7**  接口对接关系<a name="fig111354543222"></a>

    ![](../../../../figures/scheduling/接口对接关系.png "接口对接关系")

**使用场景<a name="section14336140104818"></a>**

当前支持在以下2种场景下使用借轨通信任务暂停与回切功能。

- 交换机升级场景：人工触发借轨后升级交换机，再回切。
- 故障处理场景：发生借轨的故障端口在修复完成后，再做人工回切。

**使用约束<a name="section620412554441"></a>**

- 请在训练正常迭代后，再进行借轨或回切指令的下发。
- 确保已开启进程级恢复相关功能特性。
- 仅支持Pod间为Roce通信的场景。
- 本功能依赖MindIO组件，使用前请先了解MindIO的[约束限制](../../../07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md#约束限制)。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002098609234_section4771115416256"></a>**

**表 12**  支持的产品和框架

<a name="zh-cn_topic_0000002098609234_table1526819106465"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002098609234_row22681310134611"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002098609234_p137295354447"><a name="zh-cn_topic_0000002098609234_p137295354447"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002098609234_p1172993554412"><a name="zh-cn_topic_0000002098609234_p1172993554412"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002098609234_p97299357449"><a name="zh-cn_topic_0000002098609234_p97299357449"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002098609234_row71691214122315"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p112681620231"><a name="zh-cn_topic_0000002098609234_p112681620231"></a><span id="zh-cn_topic_0000002098609234_ph9126121617231"><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul13725194132419"></a><ul id="ul13725194132419"><li><span id="ph077885871817"><a name="ph077885871817-duplicate-4"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111-duplicate-4"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul7583132019396"></a><ul id="ul7583132019396"><li><span id="ph135835207394"><a name="ph135835207394"></a>MindSpore</span></li></ul>
<a name="ul75831320173911"></a><ul id="ul75831320173911"><li><span id="ph13583142013394"><a name="ph13583142013394"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**借轨通信任务暂停与回切原理<a name="section56986212179"></a>**

**图 8**  原理图<a name="fig9336113210132"></a>

![](../../../../figures/scheduling/原理图-9.png "原理图-9")

在以上原理图中，各个步骤的说明如下。

1. AI平台集成ClusterD，调用ClusterD的gRPC接口下发切换操作，指定需要切换的NPU卡。
2. ClusterD通知MindIO暂停训练。
3. TaskD Manager通知所有TaskD Worker调用训练框架接口执行切换操作。
4. 训练框架按照通信域逐一调用CANN接口执行切换操作。
5. ClusterD判断所有NPU卡的切换操作完成后，再由TaskD通知MindIO在切换完成后继续执行下一个Step训练。

**适配功能点<a name="section1446615300284-duplicate-3"></a>**

在借轨通信任务暂停与回切中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。通过主动调用优雅暂停机制，完成当前卡上任务暂停和任务切换。集群大脑需提供对外接口，接收切换指令并管理借轨通信流程。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表13](#table19955141136102)的功能适配。

**表 13**  借轨通信任务暂停与回切框架适配功能点

<a name="table19955141136102"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.87%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387-duplicate-3"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="43.419999999999995%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816-duplicate-3"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="14.719999999999999%" id="mcps1.2.5.1.3"><p id="p10978953142414"><a name="p10978953142414"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="22.99%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812-duplicate-3"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.87%" headers="mcps1.2.5.1.1 "><p id="p1987424102519"><a name="p1987424102519"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="43.419999999999995%" headers="mcps1.2.5.1.2 "><p id="p14351731182511"><a name="p14351731182511"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.719999999999999%" headers="mcps1.2.5.1.3 "><p id="p922524114255"><a name="p922524114255"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="22.99%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212-duplicate-2"></a><a href="../../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9871924102517"><a name="p9871924102517"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p16810326255"><a name="p16810326255"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p12878242257"><a name="p12878242257"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p687122472518"><a name="p687122472518"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row1297881015253"><td class="cellrowborder" valign="top" width="18.87%" headers="mcps1.2.5.1.1 "><p id="p168711249252"><a name="p168711249252"></a>借轨切换过程管理</p>
</td>
<td class="cellrowborder" valign="top" width="43.419999999999995%" headers="mcps1.2.5.1.2 "><p id="p168762416258"><a name="p168762416258"></a>提供借轨切换请求下发能力，控制训练进程暂停与重启。</p>
</td>
<td class="cellrowborder" valign="top" width="14.719999999999999%" headers="mcps1.2.5.1.3 "><p id="p10461144315257"><a name="p10461144315257"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="22.99%" headers="mcps1.2.5.1.4 "><p id="p10979110172511"><a name="p10979110172511"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.1.0/component/clusterd/pkg/application/recover/om_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

## 弹性恢复

### 弹性训练<a name="ZH-CN_TOPIC_0000002479226542"></a>

当出现硬件故障，且K8s集群中无可用备份资源时，MindCluster会先按照数据并行域缩容部分节点继续训练，当集群中有可用空闲资源时，再触发扩容恢复原有规模训练。相比于进程级别重调度，解决了集群中无可用备份资源被重调度的问题。

**使用约束<a name="zh-cn_topic_0000002039353153_section514611624316-duplicate-2"></a>**

- 仅支持PyTorch配合MindSpeed-LLM 2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 仅支持acjob类型训练任务。
- 依赖于MindIO的优化器副本，需要存在全量优化器副本，故需要安装MindIO和TaskD配合使用。
- 不能和优雅容错功能同时开启。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不支持触发弹性训练。
- 不支持多模态模型。
- 不支持开启watchdog功能。
- 由于弹性训练会额外创建新的通信组，因此可能会导致片上内存占用增加。

    增加内存大小计算公式：增加内存最大值（MB）= HCCL\_BUFFSIZE \* 2 \* 9，其中，HCCL\_BUFFSIZE默认为200MB，HCCL\_BUFFSIZE的说明详细请参见《CANN HCCL集合通信库》中的“[HCCL_BUFFSIZE](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/910/commlib/hcclug/docs/zh/user_guide/hccl_env/HCCL_BUFFSIZE.md)”章节。

- 本功能依赖MindIO组件，使用前请先了解MindIO的[约束限制](../../../07_references/00_fault_recovery_acceleration/02_installation_and_deployment.md#约束限制)。

更多使用约束可参考[MindSpeed-LLM弹性训练功能使用约束](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/docs/pytorch/features/high_availability.md)。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039353153_section136131584164-duplicate-2"></a>**

**表 16**  弹性训练支持的产品和框架

<a name="zh-cn_topic_0000002039353153_table1991711954417-duplicate-2"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039353153_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039353153_p199171819164417"><a name="zh-cn_topic_0000002039353153_p199171819164417-duplicate-2"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="66.2966296629663%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039353153_p2917819114420"><a name="zh-cn_topic_0000002039353153_p2917819114420-duplicate-2"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="13.24132413241324%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039353153_p27578257424"><a name="zh-cn_topic_0000002039353153_p27578257424-duplicate-2"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039353153_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p153913472453"><a name="zh-cn_topic_0000002039353153_p153913472453-duplicate-2"></a><span id="zh-cn_topic_0000002039353153_ph151431757142112"><a name="zh-cn_topic_0000002039353153_ph151431757142112-duplicate-2"></a>Atlas A2 训练系列产品</span></p>
<p id="p737515258512"><a name="p737515258512-duplicate-2"></a></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><p id="p697681955215"><a name="p697681955215"></a><span id="ph157633217501"><a name="ph157633217501-duplicate-2"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><p id="p139316519435"><a name="p139316519435"></a><span id="zh-cn_topic_0000002039353153_ph2093210246488"><a name="zh-cn_topic_0000002039353153_ph2093210246488-duplicate-2"></a>PyTorch</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039353153_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p18222246142212"><a name="zh-cn_topic_0000002039353153_p18222246142212-duplicate-2"></a><span id="zh-cn_topic_0000002039353153_ph18411121792018"><a name="zh-cn_topic_0000002039353153_ph18411121792018-duplicate-2"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><p id="p1711620216528"><a name="p1711620216528"></a><span id="ph077885871817"><a name="ph077885871817-duplicate-5"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><p id="p16887149174313"><a name="p16887149174313"></a><span id="ph99469109139"><a name="ph99469109139-duplicate-2"></a>PyTorch</span></p>
</td>
</tr>
</tbody>
</table>

**弹性训练原理<a name="section3841210162013"></a>**

**图 10**  原理图<a name="fig130013397201"></a>

![](../../../../figures/scheduling/原理图-12.png "原理图-12")

以上示意图仅以缩容1个DP域为例，实际弹性训练过程中可能会一次缩容多个DP域。图中每个方格代表一个rank。

1. 按照TP（Tensor Parallelism，张量并行）、PP（Pipeline Parallelism，流水线并行）、DP（Data Parallelism，数据并行）正常进行分布式训练。
2. 训练到某一时刻，若某张卡发生故障，且集群中无更多空闲资源可被调度进行断点续训，则按照DP域缩容，即缩容1个DP域对应的Pod（可能包含多个Pod）后继续训练。
3. 缩容训练到某一时刻，集群中有空闲资源时，缩容的Pod会被重新调度，扩容恢复到原有规模继续训练。

**图 11**  流程图<a name="fig7783192415293"></a>

![](../../../../figures/scheduling/流程图.png "流程图")

在以上流程图中，各个步骤的说明如下。

1. 设备出现硬件故障后，MindCluster在服务器上的检测组件上报故障信息到ClusterD中，软件故障由容器内MindIO Controller感知并上报到ClusterD。
2. ClusterD将故障服务器上的任务容器销毁。
3. 若没有备份节点调度新容器，ClusterD通知Master节点上的MindIO Controller进行缩容训练。
4. MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用TorchNPU停止训练进程，清理正常节点的资源。
5. MindIO Controller通知正常的训练进程中的MindIO Processor执行通信组重建等缩容流程，进行缩容训练。
6. 检测到缩容时删除的Pod重调度成功。
7. ClusterD通过TaskD Manager通知MindIO Controller执行扩容。
8. MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用TorchNPU停止训练进程，清理正常节点的资源。
9. 各个进程进行集合通信建链。
10. 正常服务器上的NPU通过参数面将CKPT传递到备用服务器上，完成参数状态恢复后继续训练。

**适配功能点<a name="section1446615300284-duplicate-5"></a>**

在弹性训练中，集群大脑会根据全局故障信息决策恢复策略，并将策略下发到MindIO。调度器需要支持故障Pod调度，而非整个任务重调度，支持恢复策略依次回退。在训练容器中，框架首先初始化MindIO服务。启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保证模型参数的冗余备份。当异常发生时，通过异常捕获装饰器捕获故障模式，并由MindIO上报给集群大脑决策。

- 当集群大脑检测到故障，且无冗余备份资源时，下发缩容策略到MindIO，执行算子资源清理、缩容重建，以缩容状态继续训练。
- 当集群大脑检测到有可用资源且新节点成功拉起时，下发扩容策略到MindIO，执行算子资源清理、扩容通信重建、扩容参数面恢复和扩容状态回滚，完成弹性扩容恢复原有规模继续训练。

对于非MindSpeed-LLM和MindCluster平台用户，需在框架侧完成[表17](#table19955141136107)的功能适配。

**表 17**  弹性训练框架适配功能点

<a name="table19955141136107"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="7.520000000000001%" id="mcps1.2.6.1.1"><p id="p4637165993110"><a name="p4637165993110"></a>序号</p>
</th>
<th class="cellrowborder" valign="top" width="18.810000000000002%" id="mcps1.2.6.1.2"><p id="p46603387387"><a name="p46603387387-duplicate-5"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="34.39%" id="mcps1.2.6.1.3"><p id="p176601638153816"><a name="p176601638153816-duplicate-5"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.190000000000005%" id="mcps1.2.6.1.4"><p id="p237216122367"><a name="p237216122367-duplicate-2"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="21.090000000000003%" id="mcps1.2.6.1.5"><p id="p4660113823812"><a name="p4660113823812-duplicate-5"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="7.520000000000001%" headers="mcps1.2.6.1.1 "><p id="p26376591313"><a name="p26376591313"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.6.1.2 "><p id="p1142119117913"><a name="p1142119117913"></a>初始化启动</p>
</td>
<td class="cellrowborder" valign="top" width="34.39%" headers="mcps1.2.6.1.3 "><p id="p112827185916"><a name="p112827185916"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="16" valign="top" width="18.190000000000005%" headers="mcps1.2.6.1.4 "><p id="p444112643720"><a name="p444112643720-duplicate-2"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="21.090000000000003%" headers="mcps1.2.6.1.5 "><p id="p7146223174212"><a name="p7146223174212-duplicate-4"></a><a href="../../../07_references/00_fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">表2</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p106371759163113"><a name="p106371759163113"></a>2</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1942113117919"><a name="p1942113117919"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p92821518193"><a name="p92821518193"></a>优化器更新前上报优化器更新的开始和结束状态。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p363765912314"><a name="p363765912314"></a>3</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p164211711596"><a name="p164211711596"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p22829180917"><a name="p22829180917"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7637175903115"><a name="p7637175903115"></a>4</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p134219118919"><a name="p134219118919"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p192829181594"><a name="p192829181594"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1963725993118"><a name="p1963725993118"></a>5</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1542111118913"><a name="p1542111118913"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p112826181914"><a name="p112826181914"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p363711591310"><a name="p363711591310"></a>6</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p6421121796"><a name="p6421121796"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11282181811916"><a name="p11282181811916"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p06375599316"><a name="p06375599316"></a>7</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p34212017916"><a name="p34212017916"></a>弹性训练回调注册</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1528212181493"><a name="p1528212181493"></a>将弹性训练各个回调函数注册到MindIO。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1163581571719"><a name="p1163581571719"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_register.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row876023415918"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4637259163118"><a name="p4637259163118"></a>8</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1342113114912"><a name="p1342113114912"></a>缩容重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p528210181396"><a name="p528210181396"></a>重建缩容后的通信组、数据迭代器、记录并更新部分框架变量等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p106351815121720"><a name="p106351815121720"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_scale_in_rebuild.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row17605341596"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1763711599312"><a name="p1763711599312"></a>9</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p74211911599"><a name="p74211911599"></a>扩容通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p52821818299"><a name="p52821818299"></a>新节点与缩容节点重建通信组。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p126358155177"><a name="p126358155177"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_scale_out_rebuild.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row144412445361"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p56378590319"><a name="p56378590319"></a>10</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p64221014919"><a name="p64221014919"></a>扩容参数面恢复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p728213181097"><a name="p728213181097"></a>通过副本rank与新拉rank参数传输恢复新节点优化器等参数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p935519615208"><a name="p935519615208"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_repair.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row14716101112393"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1963713597315"><a name="p1963713597315"></a>11</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p842217115916"><a name="p842217115916"></a>扩容状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p72821618294"><a name="p72821618294"></a>恢复缩容时更改框架变量、重建数据集等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p135516622010"><a name="p135516622010"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_rollback.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row164994019817"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p563705923117"><a name="p563705923117"></a>12</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17422911918"><a name="p17422911918"></a>新拉起节点torch通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1228218181798"><a name="p1228218181798"></a>新拉起节点恢复前跳过通信。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.4 "><p id="p627084962414"><a name="p627084962414"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py#:~text=def pre_register_patches(self, patch_manager, args):" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row5499401185"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5637135973120"><a name="p5637135973120"></a>13</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11422518917"><a name="p11422518917"></a>缩容训练全局组通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p32829185918"><a name="p32829185918"></a>缩容训练时使用缩容后全局组替换原全局组通信。</p>
</td>
</tr>
<tr id="row1550640684"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3637125953118"><a name="p3637125953118"></a>14</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14221517919"><a name="p14221517919"></a>缩容训练副本组通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p6282018699"><a name="p6282018699"></a>缩容训练时副本rank替代故障rank与故障rank所在副本组通信。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.4 "><p id="p1320441192517"><a name="p1320441192517"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row1650540489"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p06374591313"><a name="p06374591313"></a>15</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5422415915"><a name="p5422415915"></a>缩容训练参数适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p112827184913"><a name="p112827184913"></a>缩容训练时修改num_microbatches、world_size、global_batch_size等参数。</p>
</td>
</tr>
<tr id="row7501940584"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6637165923110"><a name="p6637165923110"></a>16</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5422211095"><a name="p5422211095"></a>梯度精度计算适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p5282161815916"><a name="p5282161815916"></a>适配因缩容num_microbatches等变化导致的精度梯度变化。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17653121214259"><a name="p17653121214259"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接1</a></p>
<p id="p116531412172515"><a name="p116531412172515"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/pretrain_gpt.py#:~text=if args.enable_elastic_training:" target="_blank" rel="noopener noreferrer">LLM仓参考链接2</a></p>
</td>
</tr>
<tr id="row145017409813"><td class="cellrowborder" valign="top" width="7.520000000000001%" headers="mcps1.2.6.1.1 "><p id="p17637165943120"><a name="p17637165943120"></a>17</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.6.1.2 "><p id="p15422131799"><a name="p15422131799"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="34.39%" headers="mcps1.2.6.1.3 "><p id="p1628291814912"><a name="p1628291814912"></a>根据全局故障信息决策恢复策略，并将策略下发到MindIO。支持恢复策略回退，弹性训练失败回退到临终遗言等策略。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="18.190000000000005%" headers="mcps1.2.6.1.4 "><p id="p1504404816"><a name="p1504404816"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="21.090000000000003%" headers="mcps1.2.6.1.5 "><p id="p20447192572312"><a name="p20447192572312"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row155014017818"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p063755903111"><a name="p063755903111"></a>18</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1042215113910"><a name="p1042215113910"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18282181818913"><a name="p18282181818913"></a>调度故障Pod。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p10446112592310"><a name="p10446112592310"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.1.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

[表17](#table19955141136107)中序号为1-6的适配项为MindIO TFT（MindCluster MindIO Training Fault Tolerance）公共逻辑，序号为17-18的适配项为断点续训公共逻辑，本章节不再详细描述。以下针对弹性训练特有功能点，基于Megatron 0.12.1版本进行简要介绍。

- 弹性训练回调注册

    在训练拉起初始化时调用，将弹性训练缩容和扩容恢复过程中需要执行的回调函数注册到MindIO中，进而在恢复过程中被调用。

- 缩容重建
    1. 基于缩容后成员创建新的全局通信组并记录，后续将替代原全局通信组进行通信。
    2. 记录框架原始DP size、num\_microbatches等参数作为后续扩容恢复使用，并更新为缩容后数据。
    3. 基于故障Rank信息重建缩容后其他局部通信组，并更新模型、优化器等实例对象中的通信组。
    4. 重建数据集、重新初始化部分框架实例、参数等。

- 扩容通信重建
    1. 重建扩容后全局和局部通信组，并更新模型、优化器等实例对象中的通信组。
    2. 恢复框架DP size等参数、重新初始化部分框架实例等。

- 扩容参数面恢复
    1. 为新拉起的rank训练进程和备份rank训练进程创建通信组，用于发送和接收优化器参数等。
    2. 备份rank训练进程向新拉起的rank训练进程发送恢复所需的优化器参数。
    3. 新拉起的rank训练进程接收优化器参数后，按需更新optimizer、opt\_param\_scheduler、全局args等参数。

- 扩容状态回滚
    1. 恢复框架num\_microbatches等参数。
    2. 恢复训练前将优化器参数拷贝到模型参数中，并在对应DP域内进行一次all\_gather通信操作，确保模型参数为最新状态。
    3. 修复打印训练迭代日志。
    4. 重建数据集，重新初始化部分框架实例、参数等。
    5. 销毁恢复过程中发送和接收参数的通信组。

- 新拉起节点torch通信适配
    1. 对于重启节点，从pretrain启动流程到进入train之间，会下发通信算子，但正常训练rank在该阶段并未与重启节点配套重建通信域，集合通信无法成功，因此直接跳过。
    2. 对于重启节点，从pretrain启动流程到进入train之间，会创建并行通信域，但正常训练rank在该阶段并未与重启节点配套重建通信域，对于gloo组会报错，因此直接跳过新建gloo通信组。

- 缩容训练全局组通信适配

    在缩容训练过程中，由于故障节点已经被删除，因此使用原全局通信组通信会失败，需替换为缩容后的全局通信组。

- 缩容训练副本组通信适配

    在[LLM仓参考链接](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，start\_param\_sync\_wrapper、get\_grad\_norm\_fp32\_wrapper、get\_parameter\_state\_dp\_zero\_wrapper等是为了适配缩容训练时副本组通信而patch，下面以get\_parameter\_state\_dp\_zero\_wrapper为例介绍副本组适配原理：

    假设当前tp=8、pp=1、dp=4。DP组分别为rank \[0,8,16,24\]、\[1,9,17,25\]、\[2,10,18,26\]、…、\[7,15,23,31\]，按照副本优化器原理，副本组分别为rank \[0,8\]、\[16,24\]、\[1,9\]、\[17,25\]、\[2,10\]、\[18,26\]、…、\[7,15\]、\[23,31\]，rank 0-15与rank 16-31互为副本。rank 31故障后，将rank 24-31对应DP域删除继续缩容训练。

    原生Megatron会使用优化器实例的data\_parallel\_group\_gloo成员变量对应的group（即DP组，在使用MindIO的优化器副本时为副本组）进行通信。缩容后不包含删除的rank 24-31的副本组，继续按照原有通信组进行通信，包含缩容rank的副本组使用组内正常rank与缩容rank对应的副本rank组成的缩容组进行通信，例如副本组rank \[23,31\]缩容后，通信使用的通信组为rank \[23,15\]。

- 缩容训练参数适配

    在[LLM仓参考链接](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，patch\_world\_size\_func\_wrapper、log\_wrapper、is\_last\_rank\_wrapper、optimizer\_param\_scheduler\_step\_wrapper、track\_app\_tag\_wrapper、print\_rank\_last\_wrapper、num\_floating\_point\_operations\_wrapper等是为了适配global\_batch\_size、world\_size等训练中使用的参数而patch。例如：原生使用dp\_size\*micro\_batch\_size\*num\_microbatches，缩容后各个DP内num\_microbatches可能不一样，因此直接使用args.global\_batch\_size。缩容后判断是否最后一个rank使用缩容后的全局组；全局组大小修改为缩容后的大小等。

- 梯度精度计算适配

    在[LLM仓参考链接1](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，start\_grad\_sync\_wrapper、forward\_step\_wrapper、elastic\_training\_get\_forward\_backward\_func\_wrapper以及[LLM仓参考链接2](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/pretrain_gpt.py#:~text=if%20args.enable_elastic_training%3A)所指向的loss\_func的代码是为了适配因缩容导致的精度梯度变化而patch或修改。

    - loss\_func由每个micro\_batch都要进行DP组内all\_reduce通信修改为缩容训练时不进行通信，原因是缩容后每个DP域内num\_micro\_batches数量可能不一样，导致前几个DP会多执行一次all\_reduce而卡住。
    - start\_grad\_sync\_wrapper中将梯度缩放因子gradient\_scaling\_factor修改为1.0 / \(arguments.global\_batch\_size / arguments.micro\_batch\_size\)，即在原1/dp\_size基础上再除以num\_micro\_batches。
    - forward\_step\_wrapper将入参num\_microbatches修改为1，目的是loss计算时不再除以num\_microbatches，因为在start\_grad\_sync\_wrapper中已经除以了num\_microbatches。
    - elastic\_training\_get\_forward\_backward\_func\_wrapper因为loss\_func没有执行DP组内all\_reduce，原生forward\_backward\_func执行完成后，在最后一个PP时将losses\_reduced每个key的和（即所有micro\_batch的lm loss相加）在DP组内执行all\_reduce操作求和。

## 优雅容错机制

### （可选）优雅容错<a name="ZH-CN_TOPIC_0000002479226564"></a>

>[!NOTE]
>该功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。

当用户进行没有备用资源的训练任务，或者期望设备自动恢复时，可以选择使用**优雅容错**功能。即当训练时的芯片设备出现故障后，系统将尝试对故障芯片进行自动恢复，如果可以恢复，则在保持Pod运行状态下，将任务原地拉起继续训练，不能恢复则回退至**重调度模式**。

优雅容错功能无需进行资源调度，即可自动将故障设备恢复。但是它无法降低训练初始化中的恢复时间，通常情况下，优雅容错所需恢复时间大于进程级重调度和进程级在线恢复功能。

了解优雅容错的关键配置步骤，请参见[配置优雅容错](../03_configuration/01_configuring_fault_handling_policies.md#配置优雅容错)。

**使用约束<a name="zh-cn_topic_0000002098609234_section1137610139461"></a>**

- 当前只支持芯片故障使用优雅容错功能。
- 优雅容错功能与进程级别重调度、进程级在线恢复功能不能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002098609234_section4771115416256-duplicate-2"></a>**

**表 18**  优雅容错支持的产品和框架

<a name="zh-cn_topic_0000002098609234_table1526819106465-duplicate-2"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002098609234_row22681310134611"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002098609234_p137295354447"><a name="zh-cn_topic_0000002098609234_p137295354447-duplicate-2"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002098609234_p1172993554412"><a name="zh-cn_topic_0000002098609234_p1172993554412-duplicate-2"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002098609234_p97299357449"><a name="zh-cn_topic_0000002098609234_p97299357449-duplicate-2"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002098609234_row17268131014613"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p889791444417"><a name="zh-cn_topic_0000002098609234_p889791444417"></a><span id="zh-cn_topic_0000002098609234_ph289810142442"><a name="zh-cn_topic_0000002098609234_ph289810142442"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039353153_ul17412295261"></a><ul id="zh-cn_topic_0000002039353153_ul17412295261"><li><span id="ph1638757114220"><a name="ph1638757114220"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712-duplicate-2"></a>Atlas 800 训练服务器（型号 9010）</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002098609234_ul1381333331316"></a><ul id="zh-cn_topic_0000002098609234_ul1381333331316"><li><span id="zh-cn_topic_0000002098609234_ph1246144904420"><a name="zh-cn_topic_0000002098609234_ph1246144904420"></a>MindSpore</span></li></ul>
<a name="zh-cn_topic_0000002098609234_ul10570112811135"></a><ul id="zh-cn_topic_0000002098609234_ul10570112811135"><li><span id="zh-cn_topic_0000002098609234_ph473115306133"><a name="zh-cn_topic_0000002098609234_ph473115306133"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002098609234_row181221631185611"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p128991832165620"><a name="zh-cn_topic_0000002098609234_p128991832165620"></a><span id="zh-cn_topic_0000002098609234_ph13899123211565"><a name="zh-cn_topic_0000002098609234_ph13899123211565"></a>Atlas A2 训练系列产品</span></p>
<p id="p96481557151918"><a name="p96481557151918"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002098609234_ul13899193245613"></a><ul id="zh-cn_topic_0000002098609234_ul13899193245613"><li><span id="ph157633217501"><a name="ph157633217501-duplicate-3"></a>Atlas 800T A2 训练服务器</span></li><li><span id="zh-cn_topic_0000002098609234_ph189001332105615"><a name="zh-cn_topic_0000002098609234_ph189001332105615"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002098609234_ul664419915495"></a><ul id="zh-cn_topic_0000002098609234_ul664419915495"><li><span id="zh-cn_topic_0000002098609234_ph146444924919"><a name="zh-cn_topic_0000002098609234_ph146444924919"></a>MindSpore</span></li></ul>
<a name="zh-cn_topic_0000002098609234_ul36445934915"></a><ul id="zh-cn_topic_0000002098609234_ul36445934915"><li><span id="zh-cn_topic_0000002098609234_ph364489174917"><a name="zh-cn_topic_0000002098609234_ph364489174917"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002098609234_row71691214122315"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p112681620231"><a name="zh-cn_topic_0000002098609234_p112681620231-duplicate-2"></a><span id="zh-cn_topic_0000002098609234_ph9126121617231"><a name="zh-cn_topic_0000002098609234_ph9126121617231-duplicate-2"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul13725194132419-duplicate-2"></a><ul id="ul13725194132419"><li><span id="ph077885871817"><a name="ph077885871817-duplicate-6"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111-duplicate-6"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul7583132019396-duplicate-2"></a><ul id="ul7583132019396"><li><span id="ph135835207394"><a name="ph135835207394-duplicate-3"></a>MindSpore</span></li></ul>
<a name="ul75831320173911-duplicate-2"></a><ul id="ul75831320173911"><li><span id="ph13583142013394"><a name="ph13583142013394-duplicate-2"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**优雅容错原理<a name="zh-cn_topic_0000002098609234_section882584011262"></a>**

在节点或芯片故障处理过程中，若使用重调度模式，需要运维人员手动恢复设备。若任务恢复不及时可能导致训练集群中出现大量散点故障，降低集群算力利用率。因此，断点续训在**重调度模式**上增加了**优雅容错**功能，用于优化NPU芯片的部分故障容错能力。

NPU芯片故障中的部分故障可以通过退出芯片上的训练进程以及热复位芯片来恢复，优雅容错功能即针对这部分故障进行恢复处理，不需要重调度任务。

Ascend Device Plugin负责故障的上报以及设备的恢复，管理进程（PyTorch场景下为Elastic Agent组件，MindSpore场景下为TaskD组件）根据Ascend Device Plugin上报的信息进行训练进程的停止与重新拉起，完成故障恢复（不能恢复则回退至**重调度模式**）。集成优雅容错模式需要在业务容器中添加管理进程，管理进程需要具备故障感知、停止训练任务和重启训练任务等能力。

优雅容错模式直接将故障上报到业务容器内的管理进程中（通常通过挂载文件的方式），容器内的管理进程读取故障文件信息获取到故障信息，获取故障信息的流程如[图12](#zh-cn_topic_0000002098609234_fig135111361314)所示。

**图 12**  获取故障信息<a name="zh-cn_topic_0000002098609234_fig135111361314"></a>

![](../../../../figures/scheduling/获取故障信息.png "获取故障信息")

优雅容错模式将故障区分为以下四类，**无需处理**、**重新执行业务**、**需要复位芯片**和**需要重调度**，对于每类故障的处理如[图13](#zh-cn_topic_0000002098609234_fig12620181591012)所示。

**图 13**  优雅容错故障处理流程<a name="zh-cn_topic_0000002098609234_fig12620181591012"></a>

![](../../../../figures/scheduling/优雅容错故障处理流程.png "优雅容错故障处理流程")
