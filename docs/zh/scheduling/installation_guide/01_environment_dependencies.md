# 环境依赖<a name="ZH-CN_TOPIC_0000002511346315"></a>

## 软件依赖<a name="ZH-CN_TOPIC_0000002479226378"></a>

**Ascend Docker Runtime<a name="section14779174114012"></a>**

- 当前环境的Docker版本需要为18.09及以上版本。
- 宿主机已安装驱动和固件，详情请参见《CANN 软件安装指南》中的“[安装NPU驱动和固件](https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian)”章节（商用版）或“[安装NPU驱动和固件](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=openEuler)”章节（社区版）。
- Atlas 500 A2 智能小站安装Ascend Docker Runtime需要修改Docker配置。执行**vi /etc/sysconfig/docker**命令，将--config-file=""参数删除；并执行**systemctl restart docker**使配置生效。
- Atlas 500 A2 智能小站预置的MEF服务会对Docker进行安全加固配置，Ascend Docker Runtime不支持在安全加固后的Docker环境下使用。若需要使用Ascend Docker Runtime，请手动卸载MEF服务，参考《MindEdge Framework  用户指南》中的“[卸载MEF Edge](https://www.hiascend.com/document/detail/zh/mindedge/730/mef/mefug/mefug_0034.html)”章节进行操作。

    >[!NOTE]
    >
    >执行**systemctl status docker**命令，如果返回信息里包含“/docker\_entrypoint.sh”字段，则为MEF服务安全加固后的Docker。

**其他集群调度组件<a name="section172351929104018"></a>**

ARM架构和x86\_64架构对应的依赖不一样，请根据系统架构选择。集群调度组件支持IPv4和IPv6，默认使用IPv4。

**表 1**  软件环境

<a name="table20235172944010"></a>

|软件名称|支持的版本|安装位置|说明|
|--|--|--|--|
|Kubernetes|1.17.x~1.34.x（推荐使用1.19.x及以上版本）<ul><li>建议选择最新的bugfix版本。</li><li>如需安装Volcano组件，请安装1.19.x及以上版本的Kubernetes，具体Kubernetes版本请参见[Volcano官网中对应的Kubernetes版本](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility)。</li></ul>|所有节点|了解K8s的使用请参见[Kubernetes文档](https://kubernetes.io/zh-cn/docs/)。|
|（可选）Docker|18.09.x~28.5.1|所有节点|可从[Docker社区或官网](https://docs.docker.com/engine/install/)获取。使用的Docker版本需要与Kubernetes配套，配套关系可参考Kubernetes的[说明](https://github.com/kubernetes/kubernetes/tree/master/CHANGELOG)，或者从Kubernetes社区获取。建议选择最新的bugfix版本。|
|（可选）Containerd|1.4.x~2.1.4（推荐使用1.6.x版本）|所有节点|可从Containerd的[官网](https://containerd.io/downloads/)或者[社区](https://github.com/containerd/containerd/blob/main/docs/getting-started.md#installing-containerd)获取，建议选择最新的bugfix版本。请关注配套Kubernetes使用的[CRI接口版本](https://kubernetes.io/zh-cn/docs/setup/production-environment/container-runtimes/#cri-versions)。|
|昇腾AI处理器驱动和固件|请参见[版本配套表](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-training-solution-pid-258915853/software)（训练）或[版本配套表](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-inference-solution-pid-258915651/software)（推理），根据实际硬件设备型号选择与MindCluster配套的驱动、固件。|计算节点|请参见各硬件产品中[驱动和固件安装升级指南](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-hdk-pid-252764743)获取对应的指导。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>为保证NPU Exporter以二进制部署时可使用非root用户安装（如hwMindX），请在安装驱动时使用--install-for-all参数。示例如下。</p><pre class="screen">./Ascend-hdk-&lt;chip_type&gt;-npu-driver_&lt;version&gt;_linux-&lt;arch&gt;.run --full --install-for-all</pre></div></div>|
|（可选）CANN|只安装集群调度组件的情况下可不安装CANN，用户可根据实际需要选择安装所需的CANN软件包，可参见版本配套表安装对应的软件包。|计算节点或者训练推理容器内|在宿主机上安装CANN软件包，请参见《[CANN 软件安装指南](https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0000.html?Mode=PmIns&InstallType=netconda&OS=openEuler)》（商用版）或《[CANN 软件安装指南](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0000.html?Mode=PmIns&InstallType=netconda&OS=openEuler)》（社区版）。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody">若使用软切分虚拟化功能，必须安装CANN，且CANN的配套版本为8.5.0，详情请参见[vCANN-RT](https://gitcode.com/openeuler/ubs-virt/blob/master/ubs-virt-enpu/vcann-rt/README.md)。</div></div>|
|Python|3.8~3.12|训练或推理容器内|使用时Python版本请以具体AI框架为准。|

>[!NOTE]
>
>- 请根据业务的实际使用场景，选择安装Docker或者Containerd。
>- Atlas 服务器产品安装操作系统可以参见[安装指导书](https://support.huawei.com/enterprise/zh/ascend-computing/a800-9000-pid-250702818?category=installation-upgrade&subcategory=software-deployment-guide)（ARM）和[安装指导书](https://support.huawei.com/enterprise/zh/ascend-computing/a800-9010-pid-250702809?category=installation-upgrade&subcategory=software-deployment-guide)（x86\_64），安装指导书并不包含上述所有操作系统，仅供参考。
>- <term>Atlas A2 训练系列产品</term>在虚拟机场景下对操作系统的要求不同，具体的操作系统约束请参见《Atlas A2 中心推理和训练硬件 25.0.RC1 NPU驱动和固件安装指南》中的“[虚拟机安装与卸载](https://support.huawei.com/enterprise/zh/doc/EDOC1100468900/cb91d9dc)”章节。

## 组网要求<a name="ZH-CN_TOPIC_0000002479386452"></a>

由于集群调度的核心调度组件Volcano目前是部署在K8s（即Kubernetes）的管理节点，为保证业务健康稳定，部署管理节点根据K8s的部署要求作出如下建议，客户可根据自身业务特点作出调整。

- 管理节点与计算节点、存储节点分离，建议使用单独服务器部署。
- 若集群规模较大或者对业务可靠性要求较高，管理节点需使用多节点方式。

**部署逻辑示意图<a name="section10677192773320"></a>**

**图 1**  部署逻辑示意图<a name="zh-cn_topic_0000001382921066_fig1081627298"></a>  
![](../../figures/scheduling/部署逻辑示意图-3.png "部署逻辑示意图")

数据中心集群中的节点类型一般分为以下三种：

- 管理节点（即Master节点）：管理集群，负责分发训练、推理任务到各个计算节点执行，可安装与Master节点相关联的集群调度组件。
- 计算节点（即Worker节点）：实际执行训练和推理任务，可安装与Worker节点相关联的集群调度组件。
- 存储节点：存储数据集、训练输出的模型等数据。

用户需要将网络平面划分为：

- 业务面：用于K8s集群业务管理。
- 存储面：用于从存储节点读取训练用的数据集。因为对带宽有要求，所以建议使用单独的网络平面和网络端口，将训练节点（管理节点或计算节点）和存储节点连通。
- 参数面：用于分布式训练时训练节点之间的参数交换，可参考以下组网说明。
    - 《[Ascend Training Solution 23.0.RC1 组网指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100302398/3a822881)》：提供华为训练计算设备（包括Atlas 800 训练服务器、Atlas 900 PoD（型号 9000）等）搭建组网的相关说明。
    - [《Ascend Training Solution 25.0.RC1 组网指南（Atlas A2训练产品）》](https://support.huawei.com/enterprise/zh/doc/EDOC1100471246?idPath=23710424|251366513|22892968|252309113|258915853)：提供华为训练计算设备（包括Atlas 800T A2 训练服务器、Atlas 900 A2 PoD 集群基础单元、集成Atlas 200T A2 Box16 异构子框的训练服务器）搭建组网的相关说明。

## 软硬件规格要求<a name="ZH-CN_TOPIC_0000002479386424"></a>

**操作系统磁盘分区<a name="section13457101811533"></a>**

操作系统磁盘分区推荐如[表1](#table147711423499)所示。

**表 1**  磁盘空间规划

<a name="table147711423499"></a>

|分区|说明|大小|启动标志|
|--|--|--|--|
|/boot|启动分区。|500 MB|on|
|/var|软件运行所产生的数据存放分区，如日志、缓存等。|> 300 GB|off|
|/var/lib/docker|Docker镜像与容器存放分区。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody">Docker镜像和容器默认放在/var/lib/docker分区下，如果/var/lib/docker分区使用率大于85%，K8s会启动资源驱逐机制，使用时请确保/var/lib/docker分区使用率在85%以下。</div></div>|> 300 GB|off|
|/etc/mindx-dl|该分区会存放导入的证书、KubeConfig等文件。建议配置100MB，可根据实际情况调整。|100 MB|off|
|/|主分区。|> 300 GB|off|

**硬件规格要求<a name="section8991121132815"></a>**

硬件产品需要满足如下要求：

**表 2**  资源要求

<a name="table292311420386"></a>

|名称|要求|
|--|--|
|CPU|管理节点CPU＞32核|
|内存|管理节点内存＞64GB|
|磁盘空间|＞1TB，磁盘空间规划请参见[表1](#table147711423499)|
|网络|<ul><li>带外管理（BMC）：≥1Gbit/s</li><li>带内管理（SSH）：≥1Gbit/s</li><li>业务面：≥10Gbit/s</li><li>存储面：≥25Gbit/s</li><li>参数面：100Gbit/s或200Gbit/s</li></ul>|

**集群调度组件资源配置要求<a name="section168471642185618"></a>**

集群调度组件资源配置需要满足如下要求：

**表 3**  管理节点组件资源配置要求

<a name="table1259491717587"></a>
<table><thead align="left"><tr id="row13594171755810"><th class="cellrowborder" rowspan="2" valign="top" id="mcps1.2.8.1.1"><p id="p16594201795815"><a name="p16594201795815"></a><a name="p16594201795815"></a>组件名称</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.2"><p id="p115944179581"><a name="p115944179581"></a><a name="p115944179581"></a>100节点以内</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.3"><p id="p12594217115811"><a name="p12594217115811"></a><a name="p12594217115811"></a>500节点</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.4"><p id="p1293144765912"><a name="p1293144765912"></a><a name="p1293144765912"></a>1000节点</p>
</th>
</tr>
<tr id="row124371832162218"><th class="cellrowborder" valign="top" id="mcps1.2.8.2.1"><p id="p243733210220"><a name="p243733210220"></a><a name="p243733210220"></a>CPU（单位：核）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.2"><p id="p1437193212222"><a name="p1437193212222"></a><a name="p1437193212222"></a>内存（单位：GB）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.3"><p id="p543773216224"><a name="p543773216224"></a><a name="p543773216224"></a>CPU（单位：核）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.4"><p id="p1260710142233"><a name="p1260710142233"></a><a name="p1260710142233"></a>内存（单位：GB）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.5"><p id="p54375321221"><a name="p54375321221"></a><a name="p54375321221"></a>CPU（单位：核）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.6"><p id="p1965417425238"><a name="p1965417425238"></a><a name="p1965417425238"></a>内存（单位：GB）</p>
</th>
</tr>
</thead>
<tbody><tr id="row10594191713589"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p75941817155816"><a name="p75941817155816"></a><a name="p75941817155816"></a><span id="ph525518226126"><a name="ph525518226126"></a><a name="ph525518226126"></a>Volcano</span> Scheduler</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p16594161785820"><a name="p16594161785820"></a><a name="p16594161785820"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p75661624202111"><a name="p75661624202111"></a><a name="p75661624202111"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p359415171584"><a name="p359415171584"></a><a name="p359415171584"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p960771462311"><a name="p960771462311"></a><a name="p960771462311"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p5931164705914"><a name="p5931164705914"></a><a name="p5931164705914"></a>5.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p14654342142315"><a name="p14654342142315"></a><a name="p14654342142315"></a>8</p>
</td>
</tr>
<tr id="row859401719586"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p185941517145811"><a name="p185941517145811"></a><a name="p185941517145811"></a><span id="ph154111939182412"><a name="ph154111939182412"></a><a name="ph154111939182412"></a>Volcano</span> Controller</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p19594101705811"><a name="p19594101705811"></a><a name="p19594101705811"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p18566172414211"><a name="p18566172414211"></a><a name="p18566172414211"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p15941817195810"><a name="p15941817195810"></a><a name="p15941817195810"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p19607101417238"><a name="p19607101417238"></a><a name="p19607101417238"></a>3</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p9931144725915"><a name="p9931144725915"></a><a name="p9931144725915"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p10654542122312"><a name="p10654542122312"></a><a name="p10654542122312"></a>4</p>
</td>
</tr>
<tr id="row19828191113591"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p182871175916"><a name="p182871175916"></a><a name="p182871175916"></a>Infer Operator</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p2082881195911"><a name="p2082881195911"></a><a name="p2082881195911"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p12567122472117"><a name="p12567122472117"></a><a name="p12567122472117"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p15828111145917"><a name="p15828111145917"></a><a name="p15828111145917"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p460741414230"><a name="p460741414230"></a><a name="p460741414230"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p1993154735916"><a name="p1993154735916"></a><a name="p1993154735916"></a>8</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p265434215235"><a name="p265434215235"></a><a name="p265434215235"></a>8</p>
</td>
</tr>
<tr id="row19828191113591"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p182871175916"><a name="p182871175916"></a><a name="p182871175916"></a>Ascend Operator</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p2082881195911"><a name="p2082881195911"></a><a name="p2082881195911"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p12567122472117"><a name="p12567122472117"></a><a name="p12567122472117"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p15828111145917"><a name="p15828111145917"></a><a name="p15828111145917"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p460741414230"><a name="p460741414230"></a><a name="p460741414230"></a>3</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p1993154735916"><a name="p1993154735916"></a><a name="p1993154735916"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p265434215235"><a name="p265434215235"></a><a name="p265434215235"></a>4</p>
</td>
</tr>
<tr id="row138951522135910"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p9895152214593"><a name="p9895152214593"></a><a name="p9895152214593"></a><span id="ph189871659101117"><a name="ph189871659101117"></a><a name="ph189871659101117"></a>ClusterD</span></p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p10895152216594"><a name="p10895152216594"></a><a name="p10895152216594"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p175671244212"><a name="p175671244212"></a><a name="p175671244212"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p158951922165910"><a name="p158951922165910"></a><a name="p158951922165910"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p1460717147234"><a name="p1460717147234"></a><a name="p1460717147234"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p129318474598"><a name="p129318474598"></a><a name="p129318474598"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p1265416427235"><a name="p1265416427235"></a><a name="p1265416427235"></a>8</p>
</td>
</tr>
</tbody>
</table>

**表 4**  计算节点组件资源配置要求

<a name="table8522160193317"></a>

|组件名称|CPU（单位：核）|内存（单位：GB）|
|--|--|--|
|Ascend Device Plugin|0.5|0.5|
|NodeD|0.5|0.3|
|NPU Exporter|1|1|
|Ascend Docker Runtime|Docker的业务插件，无需单独的CPU和内存空间|
