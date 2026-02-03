# 概述<a name="ZH-CN_TOPIC_0000002511426835"></a>

集群调度组件基于业界流行的集群调度系统Kubernetes，增加了昇腾AI处理器（NPU）的支持，提供NPU资源管理、优化调度和分布式训练集合通信配置等基础功能。深度学习平台开发厂商可以有效减少底层资源调度相关软件开发工作量，使能用户基于MindCluster快速开发深度学习平台。

本文档是用户使用集群调度组件的指导文档，在安装和使用集群调度组件前，用户需要提前了解[集群调度组件的特性](#特性介绍)，并根据具体特性的特点和功能，选择需要使用的特性并[安装相应的组件](./installation_guide.md#安装部署)。

**使用流程<a name="section10118105218514"></a>**

集群调度组件的安装和使用流程如下图所示。

![](../figures/scheduling/zh-cn_image_0000002511426865.png)

**表 1**  使用流程

<a name="table475516228316"></a>
<table><thead align="left"><tr id="row875522218318"><th class="cellrowborder" valign="top" width="30.620000000000005%" id="mcps1.2.3.1.1"><p id="p1675542213119"><a name="p1675542213119"></a><a name="p1675542213119"></a>步骤</p>
</th>
<th class="cellrowborder" valign="top" width="69.38%" id="mcps1.2.3.1.2"><p id="p775562214311"><a name="p775562214311"></a><a name="p775562214311"></a>描述</p>
</th>
</tr>
</thead>
<tbody><tr id="row37551222183114"><td class="cellrowborder" valign="top" width="30.620000000000005%" headers="mcps1.2.3.1.1 "><p id="p1075552220317"><a name="p1075552220317"></a><a name="p1075552220317"></a>选择特性</p>
</td>
<td class="cellrowborder" valign="top" width="69.38%" headers="mcps1.2.3.1.2 "><p id="p3756102263112"><a name="p3756102263112"></a><a name="p3756102263112"></a>集群调度组件支持训练任务和推理任务的多种特性。每种特性所需要的组件不同，组件的配置也各不相同。用户可以根据需要，选择相应的特性进行使用，支持多个特性同时使用。</p>
</td>
</tr>
<tr id="row1075612273118"><td class="cellrowborder" valign="top" width="30.620000000000005%" headers="mcps1.2.3.1.1 "><p id="p275614224312"><a name="p275614224312"></a><a name="p275614224312"></a>安装相应组件</p>
</td>
<td class="cellrowborder" valign="top" width="69.38%" headers="mcps1.2.3.1.2 "><p id="p575612229310"><a name="p575612229310"></a><a name="p575612229310"></a>在选择特性后，需要安装相应的组件。组件的安装支持手动安装和使用工具安装。</p>
</td>
</tr>
<tr id="row2075611226311"><td class="cellrowborder" valign="top" width="30.620000000000005%" headers="mcps1.2.3.1.1 "><p id="p1275612212319"><a name="p1275612212319"></a><a name="p1275612212319"></a>使用示例参考</p>
</td>
<td class="cellrowborder" valign="top" width="69.38%" headers="mcps1.2.3.1.2 "><p id="p575632273111"><a name="p575632273111"></a><a name="p575632273111"></a>集群调度组件为用户提供全流程的特性使用示例，包括训练任务示例和推理任务示例。示例中包含集群调度组件支持的框架、模型和相应的脚本适配操作，帮助用户更好地了解和使用集群调度组件。</p>
</td>
</tr>
</tbody>
</table>

**免责声明<a name="section7267115610496"></a>**

-   本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。
-   用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。

# 组件介绍<a name="ZH-CN_TOPIC_0000002479386906"></a>














## Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002511426843"></a>

**应用场景<a name="section15761025111720"></a>**

创建容器时，为了容器内部能够正常使用昇腾AI处理器，需要引入昇腾驱动相关的脚本和命令。这些脚本和命令分布在不同的文件中，且存在变更的可能性。为了避免容器创建时冗长的文件挂载，MindCluster提供了部署在计算节点上的Ascend Docker Runtime组件。通过输入需要挂载的昇腾AI处理器编号，即可完成昇腾AI处理器及相关驱动的文件挂载。

**组件功能<a name="section586382712395"></a>**

-   提供Docker或Containerd的昇腾容器化支持，自动挂载所需文件和设备依赖。
-   部分硬件形态支持输入vNPU信息，完成vNPU的创建和销毁。

**组件上下游依赖<a name="section10767161681"></a>**

Ascend Docker Runtime逻辑接口如[图1](#fig98811251715)所示。

**图 1**  组件上下游依赖<a name="fig98811251715"></a>  
![](../figures/scheduling/组件上下游依赖.png "组件上下游依赖")

## NPU Exporter<a name="ZH-CN_TOPIC_0000002479226948"></a>

**应用场景<a name="section15761025111720"></a>**

在任务运行过程中，除芯片故障外，往往需要关注芯片的网络和算力使用情况，以便确认任务运行过程中的性能瓶颈，找到提升任务性能的方向。MindCluster提供了部署在计算节点的NPU Exporter组件，用于上报芯片的各项数据信息。

**组件功能<a name="section388944161719"></a>**

-   从驱动中获取芯片、网络的各项数据信息。
-   适配Prometheus钩子函数，提供标准的接口供Prometheus服务调用。
-   适配Telegraf钩子函数，提供标准的接口供Telegraf服务调用。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig129782047111818"></a>  
![](../figures/scheduling/组件上下游依赖-0.png "组件上下游依赖-0")

1.  从驱动中获取芯片以及网络信息，并放入本地缓存。
2.  从K8s标准化接口CRI中获取容器信息，并放入本地缓存。
3.  实现Prometheus或者Telegraf的接口，供二者周期性获取缓存中的数据信息。

## Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002479226928"></a>

**应用场景<a name="section15761025111720"></a>**

K8s需要感知资源信息来实现对资源信息的调度。除基础的CPU和内存信息以外，需通过K8s提供的设备插件机制，供用户自定义新的资源类型，从而定制个性化的资源发现和上报策略。MindCluster提供了部署在计算节点的Ascend Device Plugin服务，用于提供适合昇腾设备的资源发现和上报策略。

**组件功能<a name="section1112014512117"></a>**

-   从驱动中获取芯片的类型及型号，并上报给kubelet和资源调度的上层服务ClusterD。
-   从驱动中订阅芯片故障信息，并将芯片状态上报给kubelet，同时将芯片状态和具体故障信息上报给资源调度的上层服务。
-   从灵衢驱动中订阅灵衢网络故障信息，并将网络状态上报给kubelet，同时将灵衢网络状态和具体故障信息上报给资源调度的上层服务。
-   可配置故障的处理级别，且可在故障反复发生，或者长时间连续存在的情况下提升故障处理级别。
-   在资源挂载阶段，负责获取集群调度选中的芯片信息，并通过环境变量传递给Ascend Docker Runtime挂载。
-   若故障芯片处于空闲状态，且重启后可恢复，对芯片执行热复位。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig18917163118163"></a>  
![](../figures/scheduling/组件上下游依赖-1.png "组件上下游依赖-1")

1.  从DCMI中获取芯片的类型、数量、健康状态信息，或者下发芯片复位命令。
2.  上报芯片的类型、数量和状态给kubelet。
3.  上报芯片的类型、数量和具体故障信息给ClusterD。
4.  将调度器选中的芯片信息，以环境变量的方式告知给Ascend Docker Runtime。
5.  向容器内部下发训练任务拉起、停止的命令。

## Volcano<a name="ZH-CN_TOPIC_0000002479386902"></a>

**应用场景<a name="section15761025111720"></a>**

K8s基础调度仅能通过感知昇腾芯片的数量进行资源调度。为实现亲和性调度，最大化资源利用，需要感知昇腾芯片之间的网络连接方式，选择网络最优的资源。MindCluster提供了部署在管理节点的Volcano服务，针对不同的昇腾设备和组网方式提供网络亲和性调度。

**组件功能<a name="section1112014512117"></a>**

-   根据集群调度底层组件上报的故障信息及节点信息计算集群的可用设备信息。（self-maintain-available-card默认开启。self-maintain-available-card关闭的情况下，从集群调度底层组件获取集群的可用设备信息。）
-   从K8s的任务对象中获取用户期望的资源数量，结合集群的设备数量、设备类型和设备组网方式，选择最优资源分配给任务。
-   任务资源故障时，重新调度任务。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig1383773934815"></a>  
![](../figures/scheduling/组件上下游依赖-2.png "组件上下游依赖-2")

1.  根据ClusterD上报的信息计算集群资源信息。（此为默认使用ClusterD的场景）
2.  接收第三方下发的任务拉起配置，根据集群资源信息，选择最优节点资源。
3.  向计算节点的Ascend Device Plugin传递具体的资源选中信息，完成设备挂载。

## ClusterD<a name="ZH-CN_TOPIC_0000002511346859"></a>

**应用场景<a name="section15761025111720"></a>**

一个节点可能发生多个故障，如果由各个节点自发进行故障处理，会造成任务同时处于多种恢复策略的场景。为了协调任务的处理级别，MindCluster提供了部署在管理节点的ClusterD服务。ClusterD收集并汇总集群任务、资源和故障信息及影响范围，从任务、芯片和故障维度统计分析，统一判定故障处理级别和策略。

**组件功能<a name="section1112014512117"></a>**

-   从Ascend Device Plugin和NodeD组件获取芯片、节点和网络信息，从ConfigMap或gRPC获取公共故障信息。
-   汇总以上故障信息，供集群调度上层服务调用。
-   与训练容器内部建立连接，控制训练进程进行重计算动作。
-   与带外服务交互，传输任务信息。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig17906165344115"></a>  
![](../figures/scheduling/组件上下游依赖-3.png "组件上下游依赖-3")

1.  从各个计算节点的Ascend Device Plugin中获取芯片的信息。
2.  从各个计算节点的NodeD中获取计算节点的CPU、内存和硬盘的健康状态信息、节点DPC共享存储故障信息和灵衢网络故障信息。
3.  从ConfigMap或gRPC获取公共故障信息。
4.  汇总整个集群的资源信息，上报给Ascend-volcano-plugin。
5.  侦听集群的任务信息，将任务状态、资源使用情况等信息上报给CCAE。
6.  与容器内进程交互，控制训练进程进行重计算。

## Ascend Operator<a name="ZH-CN_TOPIC_0000002511426817"></a>

**应用场景<a name="section15761025111720"></a>**

MindCluster提供Ascend Operator组件，输入集合通信所需的主进程IP、静态组网集合通信所需的RankTable信息、当前Pod的rankId等信息。

**组件功能<a name="section1112014512117"></a>**

-   创建Pod，并将集合通信参数按照环境变量的方式注入。
-   创建RankTable文件，并按照共享存储或ConfigMap的方式挂载到容器，优化集合通信建链性能。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig1853091182713"></a>  
![](../figures/scheduling/组件上下游依赖-4.png "组件上下游依赖-4")

1.  通过Volcano感知当前任务所需资源是否满足。
2.  资源满足后，针对任务创建对应的Pod并注入集合通信参数的环境变量。
3.  Pod创建完成后，Volcano进行资源的最终选定。
4.  从Ascend Device Plugin获取任务的芯片编号、IP、rankId信息，汇总后生成集合通信文件。
5.  通过共享存储或ConfigMap，将集合通信文件挂载到容器内。

## NodeD<a name="ZH-CN_TOPIC_0000002479386924"></a>

**应用场景<a name="section15761025111720"></a>**

节点的CPU、内存或硬盘发生某些故障后，训练任务会失败。为了让训练任务在节点故障情况下快速退出，并且后续的新任务不再调度到故障节点上，MindCluster提供了NodeD组件，用于检测节点的异常。

**组件功能<a name="section1112014512117"></a>**

-   从IPMI中获取节点异常，并上报给资源调度的上层服务。
-   定时发送节点故障信息给资源调度的上层服务。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig10531114511617"></a>  
![](../figures/scheduling/组件上下游依赖-5.png "组件上下游依赖-5")

1.  从IPMI中获取计算节点的CPU、内存、硬盘的故障信息。
2.  将计算节点的CPU、内存、硬盘的故障信息上报给ClusterD。

## Resilience Controller<a name="ZH-CN_TOPIC_0000002511426827"></a>

>[!NOTE] 说明 
>Resilience Controller组件已经日落，相关内容将于2026年的8.2.RC1版本删除。最新的弹性训练能力请参见[弹性训练](./usage/resumable_training.md#弹性训练)。

**组件应用场景<a name="section15761025111720"></a>**

训练任务遇到故障，且无充足的健康资源替换故障资源时，可使用动态缩容的方式保证训练任务继续进行，待资源充足后，再通过动态扩容的方式恢复训练任务。集群调度提供了Resilience Controller组件，用于训练任务过程中的动态扩缩容。

**组件功能<a name="section1112014512117"></a>**

提供弹性缩容训练服务。在训练任务使用的硬件发生故障时，剔除该硬件并继续训练。

**组件上下游依赖<a name="section4941922192110"></a>**

Resilience Controller组件属于Kubernetes插件，需要安装到K8s集群中。Resilience Controller仅支持VolcanoJob类型的任务，需要集群中同时安装Volcano。Resilience Controller运行过程中仅与K8s交互，相关交互如下图所示。

**图 1** Resilience Controller组件上下游依赖<a name="fig11643146182015"></a>  
![](../figures/scheduling/Resilience-Controller组件上下游依赖.png "Resilience-Controller组件上下游依赖")

-   MindCluster集群调度组件通过K8s将NPU设备、节点状态以及调度配置等信息写入ConfigMap中。
-   Resilience Controller读取mindx-dl命名空间下，name前缀为"mindx-dl-nodeinfo-"ConfigMap中的“**NodeInfo**”字段，获取节点心跳情况。
-   Resilience Controller读取kube-system命名空间下，name前缀为"mindx-dl-deviceinfo-"的ConfigMap，读取其中“**DeviceInfoCfg**”字段，获取NPU设备健康状态。
-   Resilience Controller读取volcano-system命名空间下，名为volcano-scheduler的ConfigMap，读取其中“**grace-over-time**”字段，获取重调度pod优雅删除超时配置。
-   Resilience Controller获取集群中所有包含label为“**nodeDEnable=on**”的节点，作为调度资源池。
-   Resilience Controller获取集群中所有vcjob对应的pod，读取“**huawei.com/AscendReal**”获取pod实际使用的NPU列表。
-   Resilience Controller读取Volcano Job，获取**“fault-scheduling”、**“**elastic-scheduling**”、“**minReplicas**”、“**phase**”等字段，确定该Volcano Job是否可以进行弹性训练。
-   当设备和节点发生故障时，Resilience Controller根据原有Volcano Job的副本数和集群资源情况，创建NPU需求减半的Volcano Job。

## Elastic Agent<a name="ZH-CN_TOPIC_0000002479386918"></a>

>[!NOTE] 说明 
>Elastic Agent组件已经日落，相关内容将于2026年的8.3.0版本删除。后续进程级恢复能力将使用TaskD组件承载。

**组件应用场景<a name="zh-cn_topic_0000002062230220_zh-cn_topic_0000002046307045_section15761025111720"></a>**

因大模型训练任务过程中容易出现各种软硬件故障，导致训练任务受到影响，MindCluster集群调度组件提供了部署在计算节点的Elastic Agent的二进制包，用于提供昇腾设备上训练任务的管理功能。

**组件功能<a name="zh-cn_topic_0000002062230220_zh-cn_topic_0000002046307045_section1112014512117"></a>**

-   针对PyTorch框架提供适配昇腾设备的进程管理功能，在出现软硬件故障时，完成训练进程的停止或重启。
-   负责对接K8s集群中的集群控制中心，根据集群控制中心完成训练管理。

**组件上下游依赖<a name="zh-cn_topic_0000002062230220_zh-cn_topic_0000002046307045_section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig19841330125219"></a>  
![](../figures/scheduling/组件上下游依赖-6.png "组件上下游依赖-6")

-   MindCluster集群调度组件通过K8s将设备和训练任务状态等信息写入ConfigMap中，并映射到容器内，ConfigMap名称为[reset-config-任务名称](./api/volcano.md#任务信息)。
-   Elastic Agent通过ConfigMap获取当前训练容器所使用的设备状况和训练任务状态等信息。
-   Elastic Agent对接K8s集群控制中心，根据集群控制中心完成训练管理。

## TaskD<a name="ZH-CN_TOPIC_0000002479386914"></a>

**组件应用场景<a name="zh-cn_topic_0000002062230220_zh-cn_topic_0000002046307045_section15761025111720"></a>**

大模型训练及推理任务在业务执行中会出现故障、性能劣化等问题，导致任务受影响。MindCluster集群调度的TaskD组件提供昇腾设备上训练及推理任务的状态监测和状态控制能力。

当前版本TaskD存在两套业务流，业务流一为PyTorch、MindSpore场景下故障快速恢复业务；业务流二为训练业务运维管理业务（当前版本两套业务流存在安装部署使用和上下游依赖为两套机制的情况，后续版本将在安装部署使用和上下游依赖归一为一套机制）。

**组件架构<a name="section64107568348"></a>**

**图 1**  软件架构图<a name="fig1131414418422"></a>  
![](../figures/scheduling/软件架构图.png "软件架构图")

其中：

-   TaskD Manager：任务管理中心控制模块，通过管理其他TaskD模块完成业务状态控制
-   TaskD Proxy：消息转发模块，作为每个容器内的消息代理将消息发送到TaskD Manager中
-   TaskD Agent：进程管理模块，作为业务进程的管理进程完成业务进程生命周期管理
-   TaskD Worker：业务管理模块，作为业务进程的线程完成业务进程状态管理

**组件功能<a name="zh-cn_topic_0000002062230220_zh-cn_topic_0000002046307045_section1112014512117"></a>**

-   **业务流一场景下各组件的功能说明如下。**
    -   PyTorch、MindSpore框架提供适配昇腾设备的进程管理功能，在出现软硬件故障时，完成训练进程的停止与重启。

    -   负责对接K8s的集群控制中心，根据集群控制中心完成训练管理，管理训练任务的状态。

-   **业务流二场景下各组件的功能说明如下。**
    -   提供训练数据的轻量级profiling能力，根据集群控制中心控制完成profiling数据采集。
    -   提供借轨回切、在线压测能力。

**组件上下游依赖<a name="section1880392415224"></a>**

-   **业务流一场景下组件的上下游依赖说明如下。**

    -   MindCluster集群调度组件通过K8s将设备和训练状态等信息写入ConfigMap中，并映射到容器内，ConfigMap名称为[reset-config-<任务名称\>](./api/ascend_device_plugin.md#任务信息)。
    -   MindCluster集群调度组件通过K8s将训练状态检测指令写入ConfigMap中，并映射到容器内。
    -   TaskD  Manager通过ConfigMap获取当前训练容器所使用的设备状况和训练任务状态等信息。
    -   TaskD  Manager对接K8s集群控制中心，根据集群控制中心完成训练管理。

    **图 2**  组件上下游依赖\_业务流**一**<a name="fig113811033154417"></a>  
    ![](../figures/scheduling/组件上下游依赖_业务流一.png "组件上下游依赖_业务流一")

-   **业务流二场景下组件的上下游依赖说明如下。**

    -   TaskD  Worker通过ConfigMap获取当前任务的训练检测功能开启指令。
    -   TaskD  Manager通过gRPC获取当前任务的训练检测功能开启指令。

    **图 3**  组件上下游依赖\_业务流二<a name="fig1894945324911"></a>  
    ![](../figures/scheduling/组件上下游依赖_业务流二.png "组件上下游依赖_业务流二")

## MindIO ACP<a name="ZH-CN_TOPIC_0000002479226942"></a>

**组件应用场景<a name="section15761025111720"></a>**

Checkpoint是模型中断训练后恢复的关键点，Checkpoint的密集程度、保存和恢复的性能较为关键，它可以提高训练系统的有效吞吐率。MindIO ACP针对Checkpoint的加速方案，支持昇腾产品在LLM模型领域扩展市场空间。

**组件功能<a name="section1112014512117"></a>**

在大模型训练中，使用训练服务器内存作为缓存，对Checkpoint的保存及加载进行加速。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1** MindIO ACP<a name="fig24667426549"></a>  
![](../figures/scheduling/MindIO-ACP.png "MindIO-ACP")

## MindIO TFT<a name="ZH-CN_TOPIC_0000002511426847"></a>

**组件应用场景<a name="section15761025111720"></a>**

LLM训练中，每次保存Checkpoint数据，加载数据重新迭代训练，保存和加载周期Checkpoint，都需要比较长的时间。在故障发生后，MindIO TFT特性，立即生成一次Checkpoint数据，恢复时也能立即恢复到故障前一刻的状态，减少迭代损失。MindIO UCE和MindIO ARF针对不同的故障类型，完成在线修复或仅故障节点重启级别的在线修复，节约集群停止重启时间。

**组件功能<a name="section1112014512117"></a>**

MindIO TFT包括临终Checkpoint保存、进程级在线恢复和优雅容错等功能，分别对应：

-   MindIO TTP主要是在大模型训练过程中发生故障后，校验中间状态数据的完整性和一致性，生成一次临终Checkpoint数据，恢复训练时能够通过该Checkpoint数据恢复，减少故障造成的训练迭代损失。
-   MindIO UCE主要针对大模型训练过程中片上内存的UCE故障检测，并完成在线修复，达到Step级重计算。
-   MindIO ARF主要针对训练发生异常后，不用重新拉起整个集群，只需以节点为单位进行重启或替换，完成修复并继续训练。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1** MindIO TFT<a name="fig117818118588"></a>  
![](../figures/scheduling/MindIO-TFT.png "MindIO-TFT")

## Container Manager<a name="ZH-CN_TOPIC_0000002524312655"></a>

**应用场景<a name="section11132193111423"></a>**

在无K8s的场景下，推理或者训练进程异常后，无法通过Volcano和Ascend Device Plugin停止并重新调度业务容器、隔离故障节点、复位NPU芯片。MindCluster提供了Container Manager组件，用于无K8s场景下的容器管理和芯片复位功能。

**组件功能<a name="section1112014512117"></a>**

-   从驱动中订阅芯片故障信息，同时将芯片状态和具体故障信息存入缓存，用于后续的容器管理和芯片复位功能。
-   可配置故障的处理级别。
-   若故障芯片处于空闲状态，且重启后可恢复，对芯片执行热复位。
-   若故障芯片当前正在被容器使用，根据用户的启动配置，对占用故障芯片的容器执行停止操作，在故障芯片复位成功后，重新将容器拉起。

**组件上下游依赖<a name="section16318132318112"></a>**

**图 1**  组件上下游依赖<a name="fig107831859288"></a>  
![](../figures/scheduling/组件上下游依赖-7.png "组件上下游依赖-7")

1.  从DCMI中获取芯片的类型、数量、健康状态信息。
2.  向DCMI下发芯片复位命令。
3.  从容器运行时Docker或者Containerd中获取当前运行中的容器和芯片挂载信息。
4.  向容器运行时下发容器停止、启动命令。

# 特性介绍<a name="ZH-CN_TOPIC_0000002511426839"></a>








## 使用说明<a name="ZH-CN_TOPIC_0000002511346863"></a>

本章节描述集群调度组件特性的使用说明，包括场景说明、特性介绍、组件和特性之间的支持关系，以及使用Volcano调度器和其他调度器时特性支持的产品列表。
>[!NOTE] 说明 
>不支持Volcano调度器和其他调度器管理相同的节点资源。

**场景说明<a name="section186363476238"></a>**

训练场景：支持的特性包括资源监测、整卡调度、静态vNPU调度、断点续训和弹性训练。

推理场景：支持的特性包括资源监测、整卡调度、静态vNPU调度、动态vNPU调度、推理卡故障恢复和推理卡故障重调度。

同一集群中可能同时存在训练和推理任务，同一任务中不能同时使用仅支持训练（断点续训和弹性训练）和仅支持推理（动态vNPU调度、推理卡故障恢复和推理卡故障重调度）的特性。

**使用Volcano调度器<a name="section13135123319187"></a>**

集群调度组件支持的特性与产品的对应关系如[表1](#table192581235104013)所示，√表示支持在训练或推理任务场景下使用该特性；×表示不支持在该场景下使用该特性。

**表 1**  特性支持的产品型号

<a name="table192581235104013"></a>
<table><thead align="left"><tr id="row986220128426"><th class="cellrowborder" valign="top" id="mcps1.2.12.1.1"><p id="p1186217120422"><a name="p1186217120422"></a><a name="p1186217120422"></a>特性名称</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.12.1.2"><p id="p925233184210"><a name="p925233184210"></a><a name="p925233184210"></a>训练任务</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.12.1.3"><p id="p5663749194514"><a name="p5663749194514"></a><a name="p5663749194514"></a>训练任务</p>
</th>
<th class="cellrowborder" colspan="7" valign="top" id="mcps1.2.12.1.4"><p id="p117201852174219"><a name="p117201852174219"></a><a name="p117201852174219"></a>推理任务</p>
</th>
</tr>
</thead>
<tbody><tr id="row10785518435"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p178153438"><a name="p178153438"></a><a name="p178153438"></a><strong id="b874610505539"><a name="b874610505539"></a><a name="b874610505539"></a>产品系列</strong></p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p551710910430"><a name="p551710910430"></a><a name="p551710910430"></a><span id="ph25173916436"><a name="ph25173916436"></a><a name="ph25173916436"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p145171799431"><a name="p145171799431"></a><a name="p145171799431"></a><span id="ph155178916436"><a name="ph155178916436"></a><a name="ph155178916436"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p75623439161"><a name="p75623439161"></a><a name="p75623439161"></a><span id="ph18411121792018"><a name="ph18411121792018"></a><a name="ph18411121792018"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p125171395432"><a name="p125171395432"></a><a name="p125171395432"></a>推理服务器（插<span id="ph163696166292"><a name="ph163696166292"></a><a name="ph163696166292"></a>Atlas 300I 推理卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p2082716471591"><a name="p2082716471591"></a><a name="p2082716471591"></a><span id="ph97104582114"><a name="ph97104582114"></a><a name="ph97104582114"></a><term id="zh-cn_topic_0000001519959665_term169221139190"><a name="zh-cn_topic_0000001519959665_term169221139190"></a><a name="zh-cn_topic_0000001519959665_term169221139190"></a>Atlas 200/300/500 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p117333495593"><a name="p117333495593"></a><a name="p117333495593"></a><span id="ph5263854152111"><a name="ph5263854152111"></a><a name="ph5263854152111"></a><term id="zh-cn_topic_0000001519959665_term7466858493"><a name="zh-cn_topic_0000001519959665_term7466858493"></a><a name="zh-cn_topic_0000001519959665_term7466858493"></a>Atlas 200I/500 A2 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p551749144310"><a name="p551749144310"></a><a name="p551749144310"></a><span id="ph165178910439"><a name="ph165178910439"></a><a name="ph165178910439"></a>Atlas 推理系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p1085134151611"><a name="p1085134151611"></a><a name="p1085134151611"></a><span id="ph313817549316"><a name="ph313817549316"></a><a name="ph313817549316"></a>Atlas 800I A2 推理服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p275033515910"><a name="p275033515910"></a><a name="p275033515910"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p63371231184713"><a name="p63371231184713"></a><a name="p63371231184713"></a><span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span></p>
</td>
</tr>
<tr id="row1734610113573"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p10347181175712"><a name="p10347181175712"></a><a name="p10347181175712"></a>容器化支持</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p879771716570"><a name="p879771716570"></a><a name="p879771716570"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p0797111713577"><a name="p0797111713577"></a><a name="p0797111713577"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p3797131719576"><a name="p3797131719576"></a><a name="p3797131719576"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p479721735711"><a name="p479721735711"></a><a name="p479721735711"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p1524508010"><a name="p1524508010"></a><a name="p1524508010"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p185215501604"><a name="p185215501604"></a><a name="p185215501604"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p1079751719575"><a name="p1079751719575"></a><a name="p1079751719575"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p13798517105713"><a name="p13798517105713"></a><a name="p13798517105713"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p156796461697"><a name="p156796461697"></a><a name="p156796461697"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p11275132213535"><a name="p11275132213535"></a><a name="p11275132213535"></a>√</p>
</td>
</tr>
<tr id="row72594358400"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p566760194316"><a name="p566760194316"></a><a name="p566760194316"></a>资源监测</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p62591735144020"><a name="p62591735144020"></a><a name="p62591735144020"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p132591535194010"><a name="p132591535194010"></a><a name="p132591535194010"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p4562943181611"><a name="p4562943181611"></a><a name="p4562943181611"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p925915354402"><a name="p925915354402"></a><a name="p925915354402"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p68051754704"><a name="p68051754704"></a><a name="p68051754704"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p149041901412"><a name="p149041901412"></a><a name="p149041901412"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p7259113594016"><a name="p7259113594016"></a><a name="p7259113594016"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p1526233711165"><a name="p1526233711165"></a><a name="p1526233711165"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p7679114614920"><a name="p7679114614920"></a><a name="p7679114614920"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p8275222175315"><a name="p8275222175315"></a><a name="p8275222175315"></a>√</p>
</td>
</tr>
<tr id="row2162237102110"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p66197386216"><a name="p66197386216"></a><a name="p66197386216"></a>整卡调度</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p19619163862110"><a name="p19619163862110"></a><a name="p19619163862110"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p96199383216"><a name="p96199383216"></a><a name="p96199383216"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p1556224314160"><a name="p1556224314160"></a><a name="p1556224314160"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p5619133822115"><a name="p5619133822115"></a><a name="p5619133822115"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p1805154703"><a name="p1805154703"></a><a name="p1805154703"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p179041901218"><a name="p179041901218"></a><a name="p179041901218"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p461923816213"><a name="p461923816213"></a><a name="p461923816213"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p156191838162117"><a name="p156191838162117"></a><a name="p156191838162117"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p1167912468918"><a name="p1167912468918"></a><a name="p1167912468918"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p1627592218537"><a name="p1627592218537"></a><a name="p1627592218537"></a>√</p>
</td>
</tr>
<tr id="row18259143516408"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p1666750154313"><a name="p1666750154313"></a><a name="p1666750154313"></a>静态vNPU调度</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p4259735114012"><a name="p4259735114012"></a><a name="p4259735114012"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p5259143510406"><a name="p5259143510406"></a><a name="p5259143510406"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p185621543141613"><a name="p185621543141613"></a><a name="p185621543141613"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p025983510405"><a name="p025983510405"></a><a name="p025983510405"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p2805175413010"><a name="p2805175413010"></a><a name="p2805175413010"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p39041201319"><a name="p39041201319"></a><a name="p39041201319"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p22591435164017"><a name="p22591435164017"></a><a name="p22591435164017"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p1026223713161"><a name="p1026223713161"></a><a name="p1026223713161"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p19679846193"><a name="p19679846193"></a><a name="p19679846193"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p15337113115475"><a name="p15337113115475"></a><a name="p15337113115475"></a>×</p>
</td>
</tr>
<tr id="row0259103584014"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p1566716019437"><a name="p1566716019437"></a><a name="p1566716019437"></a>动态vNPU调度</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p14259183564011"><a name="p14259183564011"></a><a name="p14259183564011"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p2259735174016"><a name="p2259735174016"></a><a name="p2259735174016"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p7562943111620"><a name="p7562943111620"></a><a name="p7562943111620"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p325923514404"><a name="p325923514404"></a><a name="p325923514404"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p78052546016"><a name="p78052546016"></a><a name="p78052546016"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p190440213"><a name="p190440213"></a><a name="p190440213"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p10259183517402"><a name="p10259183517402"></a><a name="p10259183517402"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p226223710160"><a name="p226223710160"></a><a name="p226223710160"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p155004471565"><a name="p155004471565"></a><a name="p155004471565"></a>×</p>
<p id="p196791467917"><a name="p196791467917"></a><a name="p196791467917"></a></p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p2033703174717"><a name="p2033703174717"></a><a name="p2033703174717"></a>×</p>
</td>
</tr>
<tr id="row4260143519406"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p166688024316"><a name="p166688024316"></a><a name="p166688024316"></a>断点续训</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p162604351404"><a name="p162604351404"></a><a name="p162604351404"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p82601735174017"><a name="p82601735174017"></a><a name="p82601735174017"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p8562164310162"><a name="p8562164310162"></a><a name="p8562164310162"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p42603357401"><a name="p42603357401"></a><a name="p42603357401"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p2635185610014"><a name="p2635185610014"></a><a name="p2635185610014"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p17684021110"><a name="p17684021110"></a><a name="p17684021110"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p026033510405"><a name="p026033510405"></a><a name="p026033510405"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p1226213731618"><a name="p1226213731618"></a><a name="p1226213731618"></a><a href="#li1757340152314">1</a></p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p56796466914"><a name="p56796466914"></a><a name="p56796466914"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p23371631114714"><a name="p23371631114714"></a><a name="p23371631114714"></a>×</p>
</td>
</tr>
<tr id="row1260123513405"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p1668180104319"><a name="p1668180104319"></a><a name="p1668180104319"></a>弹性训练</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p226019358403"><a name="p226019358403"></a><a name="p226019358403"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p192601535184010"><a name="p192601535184010"></a><a name="p192601535184010"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p75621243191619"><a name="p75621243191619"></a><a name="p75621243191619"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p82601535114017"><a name="p82601535114017"></a><a name="p82601535114017"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p36354561704"><a name="p36354561704"></a><a name="p36354561704"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p206841921717"><a name="p206841921717"></a><a name="p206841921717"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p14260435144018"><a name="p14260435144018"></a><a name="p14260435144018"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p32621337161615"><a name="p32621337161615"></a><a name="p32621337161615"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p2679204619912"><a name="p2679204619912"></a><a name="p2679204619912"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p123371131114713"><a name="p123371131114713"></a><a name="p123371131114713"></a>×</p>
</td>
</tr>
<tr id="row52601135184019"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p1366810194319"><a name="p1366810194319"></a><a name="p1366810194319"></a>推理卡故障恢复</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p7260113534013"><a name="p7260113534013"></a><a name="p7260113534013"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p3260183518403"><a name="p3260183518403"></a><a name="p3260183518403"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p856224316163"><a name="p856224316163"></a><a name="p856224316163"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p9260193514019"><a name="p9260193514019"></a><a name="p9260193514019"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p1563617561303"><a name="p1563617561303"></a><a name="p1563617561303"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p19684122216"><a name="p19684122216"></a><a name="p19684122216"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p192608358409"><a name="p192608358409"></a><a name="p192608358409"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p19262123761610"><a name="p19262123761610"></a><a name="p19262123761610"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p146797461390"><a name="p146797461390"></a><a name="p146797461390"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p033743174718"><a name="p033743174718"></a><a name="p033743174718"></a>√</p>
</td>
</tr>
<tr id="row7159342432"><td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.1 "><p id="p31591444433"><a name="p31591444433"></a><a name="p31591444433"></a>推理卡故障重调度</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p171605411433"><a name="p171605411433"></a><a name="p171605411433"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.2 "><p id="p4160149434"><a name="p4160149434"></a><a name="p4160149434"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.3 "><p id="p65621434169"><a name="p65621434169"></a><a name="p65621434169"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p19160134164314"><a name="p19160134164314"></a><a name="p19160134164314"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p1663612561011"><a name="p1663612561011"></a><a name="p1663612561011"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.502850855256579%" headers="mcps1.2.12.1.4 "><p id="p106847211115"><a name="p106847211115"></a><a name="p106847211115"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.492847854356308%" headers="mcps1.2.12.1.4 "><p id="p1160114114311"><a name="p1160114114311"></a><a name="p1160114114311"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.512853856156848%" headers="mcps1.2.12.1.4 "><p id="p16262937161619"><a name="p16262937161619"></a><a name="p16262937161619"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.952085625687708%" headers="mcps1.2.12.1.4 "><p id="p9679194615915"><a name="p9679194615915"></a><a name="p9679194615915"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.522256677003101%" headers="mcps1.2.12.1.4 "><p id="p633743110470"><a name="p633743110470"></a><a name="p633743110470"></a>√</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>-   <a name="li1757340152314"></a>1：当前仅支持MindIE Motor推理任务使用本功能，其他场景下不支持使用本功能。
>-   Atlas 200I SoC A1 核心板不支持使用动态vNPU调度。
>-   当前Atlas A3 训练系列产品中仅Atlas 900 A3 SuperPoD 超节点和Atlas 800T A3 超节点服务器支持使用整卡调度和断点续训。

**表 2**  特性及对应组件

<a name="table195276470219"></a>
<table><thead align="left"><tr id="row65281647102117"><th class="cellrowborder" rowspan="2" valign="top" id="mcps1.2.13.1.1"><p id="p1552864716217"><a name="p1552864716217"></a><a name="p1552864716217"></a>组件安装位置</p>
</th>
<th class="cellrowborder" rowspan="2" valign="top" id="mcps1.2.13.1.2"><p id="p552864782120"><a name="p552864782120"></a><a name="p552864782120"></a>组件名称</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.13.1.3"><p id="p15281476216"><a name="p15281476216"></a><a name="p15281476216"></a>整卡调度或静态vNPU调度</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.4"><p id="p2257135417413"><a name="p2257135417413"></a><a name="p2257135417413"></a>容器化支持</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.5"><p id="p18212122110597"><a name="p18212122110597"></a><a name="p18212122110597"></a>资源监测</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.6"><p id="p4827184719264"><a name="p4827184719264"></a><a name="p4827184719264"></a>断点续训</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.7"><p id="p1172614818264"><a name="p1172614818264"></a><a name="p1172614818264"></a>弹性训练</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.8"><p id="p2044444933110"><a name="p2044444933110"></a><a name="p2044444933110"></a>断点续训</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.9"><p id="p12647558174615"><a name="p12647558174615"></a><a name="p12647558174615"></a>动态vNPU调度</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.10"><p id="p10854855172610"><a name="p10854855172610"></a><a name="p10854855172610"></a>推理卡故障恢复</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.1.11"><p id="p85766562261"><a name="p85766562261"></a><a name="p85766562261"></a>推理卡故障重调度</p>
</th>
</tr>
<tr id="row3779120184419"><th class="cellrowborder" valign="top" id="mcps1.2.13.2.1"><p id="p137437223441"><a name="p137437223441"></a><a name="p137437223441"></a>训练</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.2"><p id="p18744122244413"><a name="p18744122244413"></a><a name="p18744122244413"></a>推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.3"><p id="p525718541144"><a name="p525718541144"></a><a name="p525718541144"></a>训练和推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.4"><p id="p1174462264412"><a name="p1174462264412"></a><a name="p1174462264412"></a>训练和推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.5"><p id="p1874472224410"><a name="p1874472224410"></a><a name="p1874472224410"></a>训练</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.6"><p id="p87446221441"><a name="p87446221441"></a><a name="p87446221441"></a>训练</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.7"><p id="p1144494913116"><a name="p1144494913116"></a><a name="p1144494913116"></a>推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.8"><p id="p137441622204411"><a name="p137441622204411"></a><a name="p137441622204411"></a>推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.9"><p id="p18744102213441"><a name="p18744102213441"></a><a name="p18744102213441"></a>推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.13.2.10"><p id="p4744112294419"><a name="p4744112294419"></a><a name="p4744112294419"></a>推理</p>
</th>
</tr>
</thead>
<tbody><tr id="row92941695259"><td class="cellrowborder" rowspan="4" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p112948972513"><a name="p112948972513"></a><a name="p112948972513"></a>管理节点</p>
<p id="p17569141614209"><a name="p17569141614209"></a><a name="p17569141614209"></a></p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p11691924181610"><a name="p11691924181610"></a><a name="p11691924181610"></a><span id="ph13691142411165"><a name="ph13691142411165"></a><a name="ph13691142411165"></a>Volcano</span></p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p69555415347"><a name="p69555415347"></a><a name="p69555415347"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p18478223194518"><a name="p18478223194518"></a><a name="p18478223194518"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p49999851"><a name="p49999851"></a><a name="p49999851"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p59753201317"><a name="p59753201317"></a><a name="p59753201317"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p1482794782619"><a name="p1482794782619"></a><a name="p1482794782619"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p07261648202615"><a name="p07261648202615"></a><a name="p07261648202615"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.144698364561897%" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p1344454916314"><a name="p1344454916314"></a><a name="p1344454916314"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p1884685413269"><a name="p1884685413269"></a><a name="p1884685413269"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.10 "><p id="p1285485511266"><a name="p1285485511266"></a><a name="p1285485511266"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.11 "><p id="p85765560264"><a name="p85765560264"></a><a name="p85765560264"></a>√</p>
</td>
</tr>
<tr id="row175282478217"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p16916240167"><a name="p16916240167"></a><a name="p16916240167"></a><span id="ph6691524151614"><a name="ph6691524151614"></a><a name="ph6691524151614"></a>Resilience Controller</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p10528124782111"><a name="p10528124782111"></a><a name="p10528124782111"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p10479823174517"><a name="p10479823174517"></a><a name="p10479823174517"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p1199139552"><a name="p1199139552"></a><a name="p1199139552"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p99751204113"><a name="p99751204113"></a><a name="p99751204113"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p982744718267"><a name="p982744718267"></a><a name="p982744718267"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p1072634819266"><a name="p1072634819266"></a><a name="p1072634819266"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p244418496316"><a name="p244418496316"></a><a name="p244418496316"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p984675411262"><a name="p984675411262"></a><a name="p984675411262"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p985415554263"><a name="p985415554263"></a><a name="p985415554263"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p3576165610263"><a name="p3576165610263"></a><a name="p3576165610263"></a>×</p>
</td>
</tr>
<tr id="row2528164782119"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p6691112410162"><a name="p6691112410162"></a><a name="p6691112410162"></a><span id="ph1691112411613"><a name="ph1691112411613"></a><a name="ph1691112411613"></a>Ascend Operator</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p1068914250448"><a name="p1068914250448"></a><a name="p1068914250448"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p547982320450"><a name="p547982320450"></a><a name="p547982320450"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p19100691454"><a name="p19100691454"></a><a name="p19100691454"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p1921382114599"><a name="p1921382114599"></a><a name="p1921382114599"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p1259034324417"><a name="p1259034324417"></a><a name="p1259034324417"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p13726194812264"><a name="p13726194812264"></a><a name="p13726194812264"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p114441949143114"><a name="p114441949143114"></a><a name="p114441949143114"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p584635418261"><a name="p584635418261"></a><a name="p584635418261"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p18545557266"><a name="p18545557266"></a><a name="p18545557266"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p1857625610267"><a name="p1857625610267"></a><a name="p1857625610267"></a>√</p>
</td>
</tr>
<tr id="row17568416152014"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p18517453102"><a name="p18517453102"></a><a name="p18517453102"></a><span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p10148181861120"><a name="p10148181861120"></a><a name="p10148181861120"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p12148191819114"><a name="p12148191819114"></a><a name="p12148191819114"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p81003919515"><a name="p81003919515"></a><a name="p81003919515"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p14148141891118"><a name="p14148141891118"></a><a name="p14148141891118"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p414941819114"><a name="p414941819114"></a><a name="p414941819114"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p3149191810117"><a name="p3149191810117"></a><a name="p3149191810117"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p844454910310"><a name="p844454910310"></a><a name="p844454910310"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p10149151816114"><a name="p10149151816114"></a><a name="p10149151816114"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p1614961815117"><a name="p1614961815117"></a><a name="p1614961815117"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p155724168208"><a name="p155724168208"></a><a name="p155724168208"></a>√</p>
</td>
</tr>
<tr id="row13119145112593"><td class="cellrowborder" rowspan="4" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p31193514592"><a name="p31193514592"></a><a name="p31193514592"></a>计算节点</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p1634683810169"><a name="p1634683810169"></a><a name="p1634683810169"></a><span id="ph5346193861612"><a name="ph5346193861612"></a><a name="ph5346193861612"></a>Ascend Device Plugin</span></p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p153014713216"><a name="p153014713216"></a><a name="p153014713216"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p147922319454"><a name="p147922319454"></a><a name="p147922319454"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p31001491514"><a name="p31001491514"></a><a name="p31001491514"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p112131021115912"><a name="p112131021115912"></a><a name="p112131021115912"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p13828114782612"><a name="p13828114782612"></a><a name="p13828114782612"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p17261248112611"><a name="p17261248112611"></a><a name="p17261248112611"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.144698364561897%" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p10444164914319"><a name="p10444164914319"></a><a name="p10444164914319"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p148461654162612"><a name="p148461654162612"></a><a name="p148461654162612"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.10 "><p id="p8854145542611"><a name="p8854145542611"></a><a name="p8854145542611"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.11 "><p id="p16576125622614"><a name="p16576125622614"></a><a name="p16576125622614"></a>√</p>
</td>
</tr>
<tr id="row11529174712115"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p934653817165"><a name="p934653817165"></a><a name="p934653817165"></a><span id="ph19346193821610"><a name="ph19346193821610"></a><a name="ph19346193821610"></a>Ascend Docker Runtime</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p1452915479216"><a name="p1452915479216"></a><a name="p1452915479216"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p124791223104519"><a name="p124791223104519"></a><a name="p124791223104519"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p201006911519"><a name="p201006911519"></a><a name="p201006911519"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p1621352117593"><a name="p1621352117593"></a><a name="p1621352117593"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p1382704732612"><a name="p1382704732612"></a><a name="p1382704732612"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p3726848112617"><a name="p3726848112617"></a><a name="p3726848112617"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p0444174918314"><a name="p0444174918314"></a><a name="p0444174918314"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p68462543261"><a name="p68462543261"></a><a name="p68462543261"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p1785445562617"><a name="p1785445562617"></a><a name="p1785445562617"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p205761556162619"><a name="p205761556162619"></a><a name="p205761556162619"></a>√</p>
</td>
</tr>
<tr id="row1452954772114"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p6346338191611"><a name="p6346338191611"></a><a name="p6346338191611"></a><span id="ph183461738181610"><a name="ph183461738181610"></a><a name="ph183461738181610"></a>NodeD</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p1937810018342"><a name="p1937810018342"></a><a name="p1937810018342"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p194791923184512"><a name="p194791923184512"></a><a name="p194791923184512"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p191005911510"><a name="p191005911510"></a><a name="p191005911510"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p321342145918"><a name="p321342145918"></a><a name="p321342145918"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p1682744717262"><a name="p1682744717262"></a><a name="p1682744717262"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p1872614819268"><a name="p1872614819268"></a><a name="p1872614819268"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p54441049133110"><a name="p54441049133110"></a><a name="p54441049133110"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p16987173371917"><a name="p16987173371917"></a><a name="p16987173371917"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p715414772110"><a name="p715414772110"></a><a name="p715414772110"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p113561050172111"><a name="p113561050172111"></a><a name="p113561050172111"></a>√</p>
</td>
</tr>
<tr id="row102451381705"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p5346103815163"><a name="p5346103815163"></a><a name="p5346103815163"></a><span id="ph113461838171610"><a name="ph113461838171610"></a><a name="ph113461838171610"></a>NPU Exporter</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p342991215115"><a name="p342991215115"></a><a name="p342991215115"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p44791923124519"><a name="p44791923124519"></a><a name="p44791923124519"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p9100591251"><a name="p9100591251"></a><a name="p9100591251"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p132467385013"><a name="p132467385013"></a><a name="p132467385013"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p1746171417114"><a name="p1746171417114"></a><a name="p1746171417114"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p94614141219"><a name="p94614141219"></a><a name="p94614141219"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p164451249103116"><a name="p164451249103116"></a><a name="p164451249103116"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p99861338198"><a name="p99861338198"></a><a name="p99861338198"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p12154114782116"><a name="p12154114782116"></a><a name="p12154114782116"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p163563503216"><a name="p163563503216"></a><a name="p163563503216"></a>×</p>
</td>
</tr>
<tr id="row17782155710297"><td class="cellrowborder" rowspan="2" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p47821557172917"><a name="p47821557172917"></a><a name="p47821557172917"></a>训练容器内</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p0346238191618"><a name="p0346238191618"></a><a name="p0346238191618"></a><span id="ph334653851617"><a name="ph334653851617"></a><a name="ph334653851617"></a>Elastic Agent</span></p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p97821257162915"><a name="p97821257162915"></a><a name="p97821257162915"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p6479123144519"><a name="p6479123144519"></a><a name="p6479123144519"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p31001191152"><a name="p31001191152"></a><a name="p31001191152"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p112138214597"><a name="p112138214597"></a><a name="p112138214597"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p1878819574299"><a name="p1878819574299"></a><a name="p1878819574299"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p14788165710295"><a name="p14788165710295"></a><a name="p14788165710295"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.144698364561897%" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p044594953119"><a name="p044594953119"></a><a name="p044594953119"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p398416339190"><a name="p398416339190"></a><a name="p398416339190"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.10 "><p id="p14154184732118"><a name="p14154184732118"></a><a name="p14154184732118"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="8.350481966858009%" headers="mcps1.2.13.1.11 "><p id="p23561050132117"><a name="p23561050132117"></a><a name="p23561050132117"></a>×</p>
</td>
</tr>
<tr id="row8334153935919"><td class="cellrowborder" valign="top" headers="mcps1.2.13.1.1 mcps1.2.13.2.1 "><p id="p9334183925920"><a name="p9334183925920"></a><a name="p9334183925920"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.2 mcps1.2.13.2.2 "><p id="p4335113995916"><a name="p4335113995916"></a><a name="p4335113995916"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.3 "><p id="p233523925914"><a name="p233523925914"></a><a name="p233523925914"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.3 mcps1.2.13.2.4 "><p id="p433518393597"><a name="p433518393597"></a><a name="p433518393597"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.4 mcps1.2.13.2.5 "><p id="p1433514398599"><a name="p1433514398599"></a><a name="p1433514398599"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.5 mcps1.2.13.2.6 "><p id="p1533510391599"><a name="p1533510391599"></a><a name="p1533510391599"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.6 mcps1.2.13.2.7 "><p id="p14335103915591"><a name="p14335103915591"></a><a name="p14335103915591"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.7 mcps1.2.13.2.8 "><p id="p114458495311"><a name="p114458495311"></a><a name="p114458495311"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.8 mcps1.2.13.2.9 "><p id="p19335163965915"><a name="p19335163965915"></a><a name="p19335163965915"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.9 mcps1.2.13.2.10 "><p id="p0335139135914"><a name="p0335139135914"></a><a name="p0335139135914"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.13.1.10 "><p id="p3335113916594"><a name="p3335113916594"></a><a name="p3335113916594"></a>×</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>在上表中，推理场景下仅支持MindIE Motor推理任务使用断点续训功能，其他场景不支持。

**使用其他调度器<a name="section38687534188"></a>**

不使用Volcano作为调度器时，仅支持容器化支持、资源监测、整卡调度、静态vNPU调度和推理卡故障恢复特性，如[表3](#table94882020161913)所示。√表示支持在训练或推理任务场景下使用该特性；×表示不支持在该场景下使用该特性。

**表 3**  特性支持的产品型号

<a name="table94882020161913"></a>
<table><thead align="left"><tr id="row2048817205192"><th class="cellrowborder" valign="top" id="mcps1.2.12.1.1"><p id="p648832015196"><a name="p648832015196"></a><a name="p648832015196"></a>特性名称</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.12.1.2"><p id="p144887207192"><a name="p144887207192"></a><a name="p144887207192"></a>训练任务</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.12.1.3"><p id="p471711396495"><a name="p471711396495"></a><a name="p471711396495"></a>训练任务</p>
</th>
<th class="cellrowborder" colspan="7" valign="top" id="mcps1.2.12.1.4"><p id="p174889202195"><a name="p174889202195"></a><a name="p174889202195"></a>推理任务</p>
<p id="p4923144165416"><a name="p4923144165416"></a><a name="p4923144165416"></a></p>
</th>
</tr>
</thead>
<tbody><tr id="row64881420181918"><td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.1 "><p id="p1548822041912"><a name="p1548822041912"></a><a name="p1548822041912"></a><strong id="b2488122013192"><a name="b2488122013192"></a><a name="b2488122013192"></a>产品系列</strong></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p124881420171910"><a name="p124881420171910"></a><a name="p124881420171910"></a><span id="ph948852071914"><a name="ph948852071914"></a><a name="ph948852071914"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p1748932021917"><a name="p1748932021917"></a><a name="p1748932021917"></a><span id="ph5489920101915"><a name="ph5489920101915"></a><a name="ph5489920101915"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.3 "><p id="p229112316206"><a name="p229112316206"></a><a name="p229112316206"></a><span id="ph1329122310202"><a name="ph1329122310202"></a><a name="ph1329122310202"></a><term id="zh-cn_topic_0000001519959665_term26764913715_2"><a name="zh-cn_topic_0000001519959665_term26764913715_2"></a><a name="zh-cn_topic_0000001519959665_term26764913715_2"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p84891120151920"><a name="p84891120151920"></a><a name="p84891120151920"></a>推理服务器（插<span id="ph6489182001918"><a name="ph6489182001918"></a><a name="ph6489182001918"></a>Atlas 300I 推理卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p12133926921"><a name="p12133926921"></a><a name="p12133926921"></a><span id="ph513312263214"><a name="ph513312263214"></a><a name="ph513312263214"></a><term id="zh-cn_topic_0000001519959665_term169221139190_1"><a name="zh-cn_topic_0000001519959665_term169221139190_1"></a><a name="zh-cn_topic_0000001519959665_term169221139190_1"></a>Atlas 200/300/500 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p51331126622"><a name="p51331126622"></a><a name="p51331126622"></a><span id="ph613419261921"><a name="ph613419261921"></a><a name="ph613419261921"></a><term id="zh-cn_topic_0000001519959665_term7466858493_1"><a name="zh-cn_topic_0000001519959665_term7466858493_1"></a><a name="zh-cn_topic_0000001519959665_term7466858493_1"></a>Atlas 200I/500 A2 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1948913200195"><a name="p1948913200195"></a><a name="p1948913200195"></a><span id="ph448962018197"><a name="ph448962018197"></a><a name="ph448962018197"></a>Atlas 推理系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p2489162010196"><a name="p2489162010196"></a><a name="p2489162010196"></a><span id="ph19489120131910"><a name="ph19489120131910"></a><a name="ph19489120131910"></a>Atlas 800I A2 推理服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="6.951390278055612%" headers="mcps1.2.12.1.4 "><p id="p11615541121117"><a name="p11615541121117"></a><a name="p11615541121117"></a><span id="ph181778442114"><a name="ph181778442114"></a><a name="ph181778442114"></a>A200I A2 Box 异构组件</span></p>
</td>
<td class="cellrowborder" valign="top" width="7.5315063012602534%" headers="mcps1.2.12.1.4 "><p id="p69239419541"><a name="p69239419541"></a><a name="p69239419541"></a><span id="ph19454208554"><a name="ph19454208554"></a><a name="ph19454208554"></a>Atlas 800I A3 超节点服务器</span></p>
</td>
</tr>
<tr id="row118661729217"><td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.1 "><p id="p286752925"><a name="p286752925"></a><a name="p286752925"></a>容器化支持</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p1469413102035"><a name="p1469413102035"></a><a name="p1469413102035"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p869419101033"><a name="p869419101033"></a><a name="p869419101033"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.3 "><p id="p469415101736"><a name="p469415101736"></a><a name="p469415101736"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p11694181020310"><a name="p11694181020310"></a><a name="p11694181020310"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1013402614217"><a name="p1013402614217"></a><a name="p1013402614217"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1513410261024"><a name="p1513410261024"></a><a name="p1513410261024"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1958320151733"><a name="p1958320151733"></a><a name="p1958320151733"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p8583115434"><a name="p8583115434"></a><a name="p8583115434"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.951390278055612%" headers="mcps1.2.12.1.4 "><p id="p1673210127"><a name="p1673210127"></a><a name="p1673210127"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.5315063012602534%" headers="mcps1.2.12.1.4 "><p id="p1020715145511"><a name="p1020715145511"></a><a name="p1020715145511"></a>√</p>
</td>
</tr>
<tr id="row1489020121915"><td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.1 "><p id="p3489520201910"><a name="p3489520201910"></a><a name="p3489520201910"></a>资源监测</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p1948972071919"><a name="p1948972071919"></a><a name="p1948972071919"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p8490920131912"><a name="p8490920131912"></a><a name="p8490920131912"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.3 "><p id="p72919235201"><a name="p72919235201"></a><a name="p72919235201"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1949062031915"><a name="p1949062031915"></a><a name="p1949062031915"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1134122614212"><a name="p1134122614212"></a><a name="p1134122614212"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p111340266214"><a name="p111340266214"></a><a name="p111340266214"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p184901620151913"><a name="p184901620151913"></a><a name="p184901620151913"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p14490182011190"><a name="p14490182011190"></a><a name="p14490182011190"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.951390278055612%" headers="mcps1.2.12.1.4 "><p id="p5674241214"><a name="p5674241214"></a><a name="p5674241214"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.5315063012602534%" headers="mcps1.2.12.1.4 "><p id="p32014158558"><a name="p32014158558"></a><a name="p32014158558"></a>√</p>
</td>
</tr>
<tr id="row8112247122111"><td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.1 "><p id="p17667200114313"><a name="p17667200114313"></a><a name="p17667200114313"></a>整卡调度</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p4259113515405"><a name="p4259113515405"></a><a name="p4259113515405"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p15259173516409"><a name="p15259173516409"></a><a name="p15259173516409"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.3 "><p id="p4291423112016"><a name="p4291423112016"></a><a name="p4291423112016"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p3259163574019"><a name="p3259163574019"></a><a name="p3259163574019"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p15134926127"><a name="p15134926127"></a><a name="p15134926127"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p813414261927"><a name="p813414261927"></a><a name="p813414261927"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p1125913514012"><a name="p1125913514012"></a><a name="p1125913514012"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p4261173712166"><a name="p4261173712166"></a><a name="p4261173712166"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.951390278055612%" headers="mcps1.2.12.1.4 "><p id="p16679201213"><a name="p16679201213"></a><a name="p16679201213"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.5315063012602534%" headers="mcps1.2.12.1.4 "><p id="p1020151514558"><a name="p1020151514558"></a><a name="p1020151514558"></a>√</p>
</td>
</tr>
<tr id="row1449012018195"><td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.1 "><p id="p1849018201191"><a name="p1849018201191"></a><a name="p1849018201191"></a>静态vNPU调度</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p5490162017192"><a name="p5490162017192"></a><a name="p5490162017192"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p1249002018196"><a name="p1249002018196"></a><a name="p1249002018196"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.3 "><p id="p329212239208"><a name="p329212239208"></a><a name="p329212239208"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p24901020141915"><a name="p24901020141915"></a><a name="p24901020141915"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p111341526920"><a name="p111341526920"></a><a name="p111341526920"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p191345261427"><a name="p191345261427"></a><a name="p191345261427"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p34901420181914"><a name="p34901420181914"></a><a name="p34901420181914"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p194901020131918"><a name="p194901020131918"></a><a name="p194901020131918"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.951390278055612%" headers="mcps1.2.12.1.4 "><p id="p16713217122"><a name="p16713217122"></a><a name="p16713217122"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.5315063012602534%" headers="mcps1.2.12.1.4 "><p id="p68774377550"><a name="p68774377550"></a><a name="p68774377550"></a>×</p>
</td>
</tr>
<tr id="row1749113204194"><td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.1 "><p id="p3491020141919"><a name="p3491020141919"></a><a name="p3491020141919"></a>推理卡故障恢复</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p13491182014191"><a name="p13491182014191"></a><a name="p13491182014191"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.2 "><p id="p1649182012199"><a name="p1649182012199"></a><a name="p1649182012199"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.3 "><p id="p229272392014"><a name="p229272392014"></a><a name="p229272392014"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p134913206197"><a name="p134913206197"></a><a name="p134913206197"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p17134626726"><a name="p17134626726"></a><a name="p17134626726"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p14134142616217"><a name="p14134142616217"></a><a name="p14134142616217"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p134911820121912"><a name="p134911820121912"></a><a name="p134911820121912"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.501900380076016%" headers="mcps1.2.12.1.4 "><p id="p15492122013190"><a name="p15492122013190"></a><a name="p15492122013190"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="6.951390278055612%" headers="mcps1.2.12.1.4 "><p id="p967192171210"><a name="p967192171210"></a><a name="p967192171210"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="7.5315063012602534%" headers="mcps1.2.12.1.4 "><p id="p17877163717557"><a name="p17877163717557"></a><a name="p17877163717557"></a>√</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>-   Atlas 200I SoC A1 核心板不支持使用动态vNPU调度。
>-   当前Atlas A3 训练系列产品中仅Atlas 900 A3 SuperPoD 超节点和Atlas 800T A3 超节点服务器支持整卡调度和断点续训。

**表 4**  特性及对应组件

<a name="table148781511582"></a>
<table><thead align="left"><tr id="row148785111183"><th class="cellrowborder" rowspan="2" valign="top" id="mcps1.2.8.1.1"><p id="p1859584691520"><a name="p1859584691520"></a><a name="p1859584691520"></a>组件安装位置</p>
<p id="p9682136597"><a name="p9682136597"></a><a name="p9682136597"></a></p>
</th>
<th class="cellrowborder" rowspan="2" valign="top" id="mcps1.2.8.1.2"><p id="p148785114815"><a name="p148785114815"></a><a name="p148785114815"></a>组件名称</p>
<p id="p1668113135914"><a name="p1668113135914"></a><a name="p1668113135914"></a></p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.3"><p id="p6878611689"><a name="p6878611689"></a><a name="p6878611689"></a>整卡调度或静态vNPU调度</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.1.4"><p id="p1530482712317"><a name="p1530482712317"></a><a name="p1530482712317"></a>容器化支持</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.1.5"><p id="p7646114614112"><a name="p7646114614112"></a><a name="p7646114614112"></a>资源监测</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.1.6"><p id="p17612184911318"><a name="p17612184911318"></a><a name="p17612184911318"></a>推理卡故障恢复</p>
</th>
</tr>
<tr id="row116771316599"><th class="cellrowborder" valign="top" id="mcps1.2.8.2.1"><p id="p06851312591"><a name="p06851312591"></a><a name="p06851312591"></a>训练</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.2"><p id="p19683132591"><a name="p19683132591"></a><a name="p19683132591"></a>推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.3"><p id="p1430410278311"><a name="p1430410278311"></a><a name="p1430410278311"></a>训练和推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.4"><p id="p136812138591"><a name="p136812138591"></a><a name="p136812138591"></a>训练和推理</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.5"><p id="p1968101315914"><a name="p1968101315914"></a><a name="p1968101315914"></a>推理</p>
</th>
</tr>
</thead>
<tbody><tr id="row287961119820"><td class="cellrowborder" rowspan="3" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p1811501812819"><a name="p1811501812819"></a><a name="p1811501812819"></a>管理节点</p>
<p id="p1299114692010"><a name="p1299114692010"></a><a name="p1299114692010"></a></p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p189961448141619"><a name="p189961448141619"></a><a name="p189961448141619"></a><span id="ph1996184817165"><a name="ph1996184817165"></a><a name="ph1996184817165"></a>Resilience Controller</span></p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p2879161118816"><a name="p2879161118816"></a><a name="p2879161118816"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p38797111886"><a name="p38797111886"></a><a name="p38797111886"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p181816556311"><a name="p181816556311"></a><a name="p181816556311"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.5 "><p id="p5646446817"><a name="p5646446817"></a><a name="p5646446817"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.6 "><p id="p1961274903113"><a name="p1961274903113"></a><a name="p1961274903113"></a>×</p>
</td>
</tr>
<tr id="row158791211681"><td class="cellrowborder" valign="top" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p15997448171614"><a name="p15997448171614"></a><a name="p15997448171614"></a><span id="ph15997184816168"><a name="ph15997184816168"></a><a name="ph15997184816168"></a>Ascend Operator</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p1644336104117"><a name="p1644336104117"></a><a name="p1644336104117"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p198798112814"><a name="p198798112814"></a><a name="p198798112814"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p419125516310"><a name="p419125516310"></a><a name="p419125516310"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p96461646512"><a name="p96461646512"></a><a name="p96461646512"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.5 "><p id="p761214963119"><a name="p761214963119"></a><a name="p761214963119"></a>×</p>
</td>
</tr>
<tr id="row99911546182012"><td class="cellrowborder" valign="top" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p121088587200"><a name="p121088587200"></a><a name="p121088587200"></a><span id="ph01082058102011"><a name="ph01082058102011"></a><a name="ph01082058102011"></a>ClusterD</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p14108195852014"><a name="p14108195852014"></a><a name="p14108195852014"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p12108205820202"><a name="p12108205820202"></a><a name="p12108205820202"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p11199551236"><a name="p11199551236"></a><a name="p11199551236"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p1810845816209"><a name="p1810845816209"></a><a name="p1810845816209"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.5 "><p id="p6108165810202"><a name="p6108165810202"></a><a name="p6108165810202"></a>√</p>
</td>
</tr>
<tr id="row1976128105918"><td class="cellrowborder" rowspan="4" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p13977168105916"><a name="p13977168105916"></a><a name="p13977168105916"></a>计算节点</p>
<p id="p66141571117"><a name="p66141571117"></a><a name="p66141571117"></a></p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p967213181716"><a name="p967213181716"></a><a name="p967213181716"></a><span id="ph116721618170"><a name="ph116721618170"></a><a name="ph116721618170"></a>Ascend Device Plugin</span></p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p108811111181"><a name="p108811111181"></a><a name="p108811111181"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p1188141117810"><a name="p1188141117810"></a><a name="p1188141117810"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p181975516317"><a name="p181975516317"></a><a name="p181975516317"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.5 "><p id="p46478461412"><a name="p46478461412"></a><a name="p46478461412"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.6 "><p id="p261217493310"><a name="p261217493310"></a><a name="p261217493310"></a>√</p>
</td>
</tr>
<tr id="row4880911181"><td class="cellrowborder" valign="top" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p1167218181716"><a name="p1167218181716"></a><a name="p1167218181716"></a><span id="ph36721114179"><a name="ph36721114179"></a><a name="ph36721114179"></a>Ascend Docker Runtime</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p158803111482"><a name="p158803111482"></a><a name="p158803111482"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p20880131111810"><a name="p20880131111810"></a><a name="p20880131111810"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p151021217414"><a name="p151021217414"></a><a name="p151021217414"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p1664784612120"><a name="p1664784612120"></a><a name="p1664784612120"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.5 "><p id="p11135294134"><a name="p11135294134"></a><a name="p11135294134"></a>√</p>
</td>
</tr>
<tr id="row68800116815"><td class="cellrowborder" valign="top" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p1967219161720"><a name="p1967219161720"></a><a name="p1967219161720"></a><span id="ph1567261101717"><a name="ph1567261101717"></a><a name="ph1567261101717"></a>NodeD</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p388018111810"><a name="p388018111810"></a><a name="p388018111810"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p48800111587"><a name="p48800111587"></a><a name="p48800111587"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p1031315101743"><a name="p1031315101743"></a><a name="p1031315101743"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p116474461614"><a name="p116474461614"></a><a name="p116474461614"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.5 "><p id="p561284911312"><a name="p561284911312"></a><a name="p561284911312"></a>√</p>
</td>
</tr>
<tr id="row361465711116"><td class="cellrowborder" valign="top" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p2672181101713"><a name="p2672181101713"></a><a name="p2672181101713"></a><span id="ph186723111714"><a name="ph186723111714"></a><a name="ph186723111714"></a>NPU Exporter</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p1614757411"><a name="p1614757411"></a><a name="p1614757411"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p26144578115"><a name="p26144578115"></a><a name="p26144578115"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p1531311108415"><a name="p1531311108415"></a><a name="p1531311108415"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p106141571717"><a name="p106141571717"></a><a name="p106141571717"></a>√</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.5 "><p id="p473119291618"><a name="p473119291618"></a><a name="p473119291618"></a>×</p>
</td>
</tr>
<tr id="row10881211083"><td class="cellrowborder" rowspan="2" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p288112111089"><a name="p288112111089"></a><a name="p288112111089"></a>训练容器内</p>
<p id="p12231961724"><a name="p12231961724"></a><a name="p12231961724"></a></p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p1067251141718"><a name="p1067251141718"></a><a name="p1067251141718"></a><span id="ph667210151711"><a name="ph667210151711"></a><a name="ph667210151711"></a>Elastic Agent</span></p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p14881151111819"><a name="p14881151111819"></a><a name="p14881151111819"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p48812113813"><a name="p48812113813"></a><a name="p48812113813"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p73141410740"><a name="p73141410740"></a><a name="p73141410740"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.5 "><p id="p36474461814"><a name="p36474461814"></a><a name="p36474461814"></a>×</p>
</td>
<td class="cellrowborder" valign="top" width="14.285714285714285%" headers="mcps1.2.8.1.6 "><p id="p177271312122"><a name="p177271312122"></a><a name="p177271312122"></a>×</p>
</td>
</tr>
<tr id="row1231567217"><td class="cellrowborder" valign="top" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p12311867211"><a name="p12311867211"></a><a name="p12311867211"></a><span id="ph6603952724"><a name="ph6603952724"></a><a name="ph6603952724"></a>TaskD</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p42312611217"><a name="p42312611217"></a><a name="p42312611217"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.3 "><p id="p132312061226"><a name="p132312061226"></a><a name="p132312061226"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p6231186925"><a name="p6231186925"></a><a name="p6231186925"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.4 mcps1.2.8.2.5 "><p id="p13231166227"><a name="p13231166227"></a><a name="p13231166227"></a>×</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.8.1.5 "><p id="p32316614212"><a name="p32316614212"></a><a name="p32316614212"></a>×</p>
</td>
</tr>
</tbody>
</table>

## 容器化支持<a name="ZH-CN_TOPIC_0000002479386930"></a>

**功能特点<a name="section1788818281655"></a>**

为所有的训练或推理作业提供NPU容器化支持，自动挂载所需文件和设备依赖，使用户AI作业能够以Docker容器的方式平滑运行在昇腾设备之上。

**所需组件<a name="section15655185785119"></a>**

Ascend Docker Runtime

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[容器化支持](./usage/containerization.md)章节进行操作。

## 资源监测<a name="ZH-CN_TOPIC_0000002479386910"></a>

**功能特点<a name="section1788818281655"></a>**

支持在执行训练或者推理任务时，对昇腾AI处理器资源各种数据信息的实时监测，可实时获取昇腾AI处理器利用率、温度、电压、内存，以及昇腾AI处理器在容器中的分配状况等信息，实现资源的实时监测。支持对虚拟NPU（vNPU）的AI Core利用率、vNPU总内存和vNPU使用中内存进行监测。目前NPU Exporter仅支持对Atlas 推理系列产品的vNPU资源监测。

**所需组件<a name="section15655185785119"></a>**

NPU Exporter

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[资源监测](./usage/resource_monitoring.md)章节进行操作。

## 虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855"></a>




### 功能特点<a name="ZH-CN_TOPIC_0000002511346849"></a>

**功能介绍<a name="section1337420477275"></a>**

昇腾虚拟化实例功能是指通过资源虚拟化的方式将物理机或虚拟机配置的NPU（昇腾AI处理器）切分成若干份vNPU（虚拟NPU）挂载到容器中使用，虚拟化管理方式能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放的资源操作请求。

昇腾虚拟化实例功能的优点是可实现多个用户共同使用一台服务器，用户可以按需申请vNPU，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的平稳和安全，且资源分配，资源回收过程统一，方便多租户管理。

**原理介绍<a name="section154002962818"></a>**

昇腾NPU硬件资源主要包括AICore（用于AI模型的计算）、AICPU、内存等，昇腾虚拟化实例功能主要原理是将上述硬件资源根据用户指定的资源需求划分出vNPU，每个vNPU对应若干AICore、AICPU、内存资源。比如用户只需要使用4个AICore的算力，那么系统就会创建一个vNPU，通过vNPU向NPU芯片获取4个AICore提供给容器使用，整体昇腾虚拟化实例方案如[图1 虚拟化实例方案](#fig987114711574)所示。

**图 1**  虚拟化实例方案<a name="fig987114711574"></a>  
![](../figures/scheduling/虚拟化实例方案.png "虚拟化实例方案")

### 应用场景及方案<a name="ZH-CN_TOPIC_0000002511426823"></a>

**应用场景<a name="section198715461917"></a>**

昇腾虚拟化实例功能适用于多用户多任务并行，且每个任务算力需求较小的场景。对算力需求较大的大模型任务，不支持使用昇腾虚拟化实例。

**虚拟化场景<a name="section1618382307"></a>**

昇腾虚拟化实例功能在物理机或虚拟机使用时，支持以下虚拟化场景，如[表1](#table197838103018)所示。本文主要介绍在昇腾设备划分vNPU支持的场景和方法，如果涉及虚拟机相关的配置，需要结合另一本文档《Atlas 系列硬件产品 24.1.0 虚拟机配置指南》的“安装虚拟机\>配置NPU直通虚拟机\>[NPU直通虚拟机](https://support.huawei.com/enterprise/zh/doc/EDOC1100438515/2689d3e6?idPath=23710424|251366513|254884019|261408772|252764743)”章节一起使用。

划分vNPU有以下两种方式。

-   静态虚拟化：通过npu-smi工具**手动**创建多个vNPU。物理机和虚拟机场景均支持静态虚拟化。
-   动态虚拟化：通过软件配置，在收到虚拟化任务请求后，动态地**自动**创建vNPU、挂载任务、回收vNPU。

**表 1**  使用场景

<a name="table197838103018"></a>
<table><thead align="left"><tr id="row16723873015"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p871338103019"><a name="p871338103019"></a><a name="p871338103019"></a>昇腾虚拟化实例功能支持场景</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p14014521402"><a name="p14014521402"></a><a name="p14014521402"></a>操作流程</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p6810383303"><a name="p6810383303"></a><a name="p6810383303"></a>支持昇腾硬件</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p18893873015"><a name="p18893873015"></a><a name="p18893873015"></a>支持的虚拟化方式</p>
</th>
</tr>
</thead>
<tbody><tr id="row158123818304"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1819384303"><a name="p1819384303"></a><a name="p1819384303"></a>在物理机划分vNPU，挂载vNPU到虚拟机</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1290518155817"><a name="p1290518155817"></a><a name="p1290518155817"></a>在物理机划分vNPU和挂载vNPU到虚拟机的步骤请参见<span id="ph15232948195013"><a name="ph15232948195013"></a><a name="ph15232948195013"></a>《Atlas 系列硬件产品 24.1.0 虚拟机配置指南》的“安装虚拟机&gt;配置NPU直通虚拟机&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100438515/bf80825c" target="_blank" rel="noopener noreferrer">vNPU直通虚拟机</a>”章节</span>。</p>
<p id="p134351910131711"><a name="p134351910131711"></a><a name="p134351910131711"></a></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul13936440103519"></a><a name="ul13936440103519"></a><ul id="ul13936440103519"><li><span id="ph1093694010356"><a name="ph1093694010356"></a><a name="ph1093694010356"></a>Atlas 推理系列产品</span>：<a name="ul1857157125012"></a><a name="ul1857157125012"></a><ul id="ul1857157125012"><li><span id="ph124782031105012"><a name="ph124782031105012"></a><a name="ph124782031105012"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph414110557507"><a name="ph414110557507"></a><a name="ph414110557507"></a>Atlas 300I Duo 推理卡</span></li><li><span id="ph1251751217512"><a name="ph1251751217512"></a><a name="ph1251751217512"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph129362408352"><a name="ph129362408352"></a><a name="ph129362408352"></a>Atlas 300V Pro 视频解析卡</span></li></ul>
</li></ul>
<a name="ul84351038183618"></a><a name="ul84351038183618"></a><ul id="ul84351038183618"><li><span id="ph1743573863613"><a name="ph1743573863613"></a><a name="ph1743573863613"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="ph5435103819364"><a name="ph5435103819364"></a><a name="ph5435103819364"></a>Atlas 800 训练服务器（型号 9010）</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10921030123711"><a name="p10921030123711"></a><a name="p10921030123711"></a>静态虚拟化</p>
<p id="p333261621717"><a name="p333261621717"></a><a name="p333261621717"></a></p>
</td>
</tr>
<tr id="row89138123014"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p391138203014"><a name="p391138203014"></a><a name="p391138203014"></a>在物理机划分vNPU，挂载vNPU到容器</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol4232523123116"></a><a name="ol4232523123116"></a><ol id="ol4232523123116"><li>在物理机划分vNPU的步骤请参见<a href="./usage/virtual_instance.md#创建vnpu">创建vNPU</a>。</li><li>挂载vNPU到容器的步骤请参见<a href="./usage/virtual_instance.md#挂载vnpu">挂载vNPU</a>。</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p04706138383"><a name="p04706138383"></a><a name="p04706138383"></a><span id="ph9232422123715"><a name="ph9232422123715"></a><a name="ph9232422123715"></a>Atlas 推理系列产品</span></p>
<a name="ul20606165512016"></a><a name="ul20606165512016"></a><ul id="ul20606165512016"><li><span id="ph66061055192018"><a name="ph66061055192018"></a><a name="ph66061055192018"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph1060616557200"><a name="ph1060616557200"></a><a name="ph1060616557200"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph8606655142015"><a name="ph8606655142015"></a><a name="ph8606655142015"></a>Atlas 300V Pro 视频解析卡</span></li><li><span id="ph960625515203"><a name="ph960625515203"></a><a name="ph960625515203"></a>Atlas 300I Duo 推理卡</span></li><li><span id="ph271718714435"><a name="ph271718714435"></a><a name="ph271718714435"></a>Atlas 200I SoC A1 核心板</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p671845534711"><a name="p671845534711"></a><a name="p671845534711"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row174318393462"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1528315810485"><a name="p1528315810485"></a><a name="p1528315810485"></a><span id="ph1128310819486"><a name="ph1128310819486"></a><a name="ph1128310819486"></a>Atlas 推理系列产品</span></p>
<a name="ul1528312814482"></a><a name="ul1528312814482"></a><ul id="ul1528312814482"><li><span id="ph528318174814"><a name="ph528318174814"></a><a name="ph528318174814"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph3283582485"><a name="ph3283582485"></a><a name="ph3283582485"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph192839854815"><a name="ph192839854815"></a><a name="ph192839854815"></a>Atlas 300V Pro 视频解析卡</span></li><li><span id="ph17718434852"><a name="ph17718434852"></a><a name="ph17718434852"></a>Atlas 200I SoC A1 核心板</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><div class="p" id="p879861715488"><a name="p879861715488"></a><a name="p879861715488"></a>动态虚拟化：<a name="ul1028016496477"></a><a name="ul1028016496477"></a><ul id="ul1028016496477"><li>使用<span id="ph112801498478"><a name="ph112801498478"></a><a name="ph112801498478"></a>Ascend Docker Runtime</span>挂载</li><li>使用<span id="ph828016490479"><a name="ph828016490479"></a><a name="ph828016490479"></a><span id="ph728054934716"><a name="ph728054934716"></a><a name="ph728054934716"></a>Kubernetes</span>挂载</span></li></ul>
</div>
</td>
</tr>
<tr id="row18969819380"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p955711111389"><a name="p955711111389"></a><a name="p955711111389"></a><span id="ph1160255613617"><a name="ph1160255613617"></a><a name="ph1160255613617"></a>Atlas 训练系列产品</span></p>
<a name="ul20127114712811"></a><a name="ul20127114712811"></a><ul id="ul20127114712811"><li><span id="ph1412724722816"><a name="ph1412724722816"></a><a name="ph1412724722816"></a>Atlas 300T 训练卡（型号 9000）</span></li><li><span id="ph1012754772811"><a name="ph1012754772811"></a><a name="ph1012754772811"></a>Atlas 300T Pro 训练卡（型号 9000）</span></li><li><span id="ph0127347172818"><a name="ph0127347172818"></a><a name="ph0127347172818"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="ph912713473289"><a name="ph912713473289"></a><a name="ph912713473289"></a>Atlas 800 训练服务器（型号 9010）</span></li><li><span id="ph012784742819"><a name="ph012784742819"></a><a name="ph012784742819"></a>Atlas 900 PoD（型号 9000）</span></li><li><span id="ph1012713477284"><a name="ph1012713477284"></a><a name="ph1012713477284"></a>Atlas 900T PoD Lite</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12119142253"><a name="p12119142253"></a><a name="p12119142253"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row131012387307"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1010133833013"><a name="p1010133833013"></a><a name="p1010133833013"></a>在物理机划分vNPU，挂载vNPU到虚拟机，在虚拟机内将vNPU挂载到容器</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol14307634103119"></a><a name="ol14307634103119"></a><ol id="ol14307634103119"><li>在物理机划分vNPU和挂载vNPU到虚拟机的步骤请参见<span id="ph452785715619"><a name="ph452785715619"></a><a name="ph452785715619"></a>《Atlas 系列硬件产品 24.1.0 虚拟机配置指南》的“安装虚拟机&gt;配置NPU直通虚拟机&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100438515/bf80825c" target="_blank" rel="noopener noreferrer">vNPU直通虚拟机</a>”章节</span>。</li><li>在虚拟机内挂载vNPU到容器的步骤请参见<a href="./usage/virtual_instance.md#挂载vnpu">挂载vNPU</a>。</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><div class="p" id="p1711938143015"><a name="p1711938143015"></a><a name="p1711938143015"></a><span id="ph11726191810373"><a name="ph11726191810373"></a><a name="ph11726191810373"></a>Atlas 推理系列产品</span>：<a name="ul211163816305"></a><a name="ul211163816305"></a><ul id="ul211163816305"><li><span id="ph5967551195016"><a name="ph5967551195016"></a><a name="ph5967551195016"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph126311222155714"><a name="ph126311222155714"></a><a name="ph126311222155714"></a>Atlas 300I Duo 推理卡</span></li><li><span id="ph1470163516116"><a name="ph1470163516116"></a><a name="ph1470163516116"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph11483102195114"><a name="ph11483102195114"></a><a name="ph11483102195114"></a>Atlas 300V Pro 视频解析卡</span></li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p13911193234713"><a name="p13911193234713"></a><a name="p13911193234713"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row3124381309"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20127385307"><a name="p20127385307"></a><a name="p20127385307"></a>在物理机直通NPU到虚拟机，在虚拟机内划分vNPU，再将vNPU挂载到虚拟机内的容器</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol441318447318"></a><a name="ol441318447318"></a><ol id="ol441318447318"><li>在物理机直通NPU到虚拟机的步骤请参见<span id="ph970622925815"><a name="ph970622925815"></a><a name="ph970622925815"></a>《Atlas 系列硬件产品 24.1.0 虚拟机配置指南》的“安装虚拟机&gt;配置NPU直通虚拟机&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100438515/2689d3e6?idPath=23710424|251366513|254884019|261408772|252764743" target="_blank" rel="noopener noreferrer">NPU直通虚拟机</a>”章节</span>。</li><li>在虚拟机内划分vNPU步骤请参见<a href="./usage/virtual_instance.md#创建vnpu">创建vNPU</a>。</li><li>将vNPU挂载到虚拟机内的容器的步骤请参见<a href="./usage/virtual_instance.md#挂载vnpu">挂载vNPU</a>。</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p22935614382"><a name="p22935614382"></a><a name="p22935614382"></a><span id="ph32931769386"><a name="ph32931769386"></a><a name="ph32931769386"></a>Atlas 推理系列产品</span>：</p>
<a name="ul229314623811"></a><a name="ul229314623811"></a><ul id="ul229314623811"><li><span id="ph2029314619382"><a name="ph2029314619382"></a><a name="ph2029314619382"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph0349814593"><a name="ph0349814593"></a><a name="ph0349814593"></a>Atlas 300I Duo 推理卡</span></li><li><span id="ph2247511161217"><a name="ph2247511161217"></a><a name="ph2247511161217"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph2293668388"><a name="ph2293668388"></a><a name="ph2293668388"></a>Atlas 300V Pro 视频解析卡</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1486613195014"><a name="p1486613195014"></a><a name="p1486613195014"></a>静态虚拟化</p>
</td>
</tr>
<tr id="row8918450194820"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15300165019495"><a name="p15300165019495"></a><a name="p15300165019495"></a><span id="ph0300150104913"><a name="ph0300150104913"></a><a name="ph0300150104913"></a>Atlas 推理系列产品</span>：</p>
<a name="ul11300350184920"></a><a name="ul11300350184920"></a><ul id="ul11300350184920"><li><span id="ph4300145015499"><a name="ph4300145015499"></a><a name="ph4300145015499"></a>Atlas 300I Pro 推理卡</span></li><li><span id="ph1300050194911"><a name="ph1300050194911"></a><a name="ph1300050194911"></a>Atlas 300V 视频解析卡</span></li><li><span id="ph18300185014917"><a name="ph18300185014917"></a><a name="ph18300185014917"></a>Atlas 300V Pro 视频解析卡</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><div class="p" id="p14998206105010"><a name="p14998206105010"></a><a name="p14998206105010"></a>动态虚拟化：<a name="ul55138515017"></a><a name="ul55138515017"></a><ul id="ul55138515017"><li>使用<span id="ph1051325105019"><a name="ph1051325105019"></a><a name="ph1051325105019"></a>Ascend Docker Runtime</span>挂载</li><li>使用<span id="ph1951314565011"><a name="ph1951314565011"></a><a name="ph1951314565011"></a><span id="ph1251318575010"><a name="ph1251318575010"></a><a name="ph1251318575010"></a>Kubernetes</span>挂载</span></li></ul>
</div>
</td>
</tr>
</tbody>
</table>

**vNPU挂载到容器方案<a name="section84114107544"></a>**

将vNPU挂载到容器有以下方案：

-   原生Docker：结合原生Docker使用。仅支持静态虚拟化（通过npu-smi工具创建多个vNPU），通过Docker拉起容器时将vNPU挂载到容器。

    >[!NOTE] 说明 
    >不支持通过原生Containerd拉起容器时将vNPU挂载到容器。

-   结合MindCluster组件：
    -   Ascend Docker Runtime：单独基于Ascend Docker Runtime（容器引擎插件）使用。支持静态虚拟化和动态虚拟化，通过Ascend Docker Runtime拉起容器时将vNPU挂载到容器。
    -   Kubernetes：结合MindCluster组件Ascend Device Plugin、Volcano，通过Kubernetes拉起容器时将vNPU挂载到容器。支持静态虚拟化和动态虚拟化。
        -   静态虚拟化：通过npu-smi工具提前创建多个vNPU，当用户需要使用vNPU资源时，基于Ascend Device Plugin组件的设备发现、设备分配、设备健康状态上报功能，分配vNPU资源提供给上层用户使用，此方案下，集群调度组件的Volcano组件为可选。
        -   动态虚拟化：Ascend Device Plugin组件上报其所在机器的可用AICore数目。虚拟化任务上报后，Volcano经过计算将该任务调度到满足其要求的节点。该节点的Ascend Device Plugin在收到请求后自动切分出vNPU设备并挂载该任务，从而完成整个动态虚拟化过程。该过程不需要用户提前切分vNPU，在任务使用完成后又能自动回收，很好地支持用户算力需求不断变化的场景。

### 所需组件<a name="ZH-CN_TOPIC_0000002479226932"></a>

根据创建或挂载vNPU的方式不同，所需组件不同，可以参考如下内容。

**创建vNPU所需组件<a name="section17158108347"></a>**

创建vNPU有以下两种方式。

-   静态虚拟化：通过npu-smi工具**手动**创建多个vNPU。
-   动态虚拟化：通过MindCluster中的以下组件创建vNPU。
    -   方式一：通过Ascend Docker Runtime**手动**创建vNPU，容器进程结束时，自动销毁vNPU。
    -   方式二：通过Volcano和Ascend Device Plugin动态地**自动**创建vNPU，容器进程结束时，自动销毁vNPU。

**挂载vNPU所需组件<a name="section18777164353415"></a>**

根据创建vNPU的方式的不同，将vNPU挂载到容器的方式也不同，说明如下：

-   基于原生Docker挂载vNPU（只支持静态虚拟化）
-   基于MindCluster组件挂载vNPU（支持静态虚拟化和动态虚拟化）
    -   方式一：通过Ascend Docker Runtime+Docker方式挂载vNPU（此方式相比只使用原生Docker易用性更高）。
    -   方式二：通过Kubernetes挂载vNPU。

**安装说明<a name="section1350915844811"></a>**

-   驱动安装后会默认安装npu-smi工具，安装操作请参考《CANN 软件安装指南》中的“安装NPU驱动和固件”章节（商用版）或“安装NPU驱动和固件”章节（社区版）；安装成功后，npu-smi放置在“/usr/local/sbin/”和“/usr/local/bin/”路径下。
-   安装MindCluster中的Ascend Docker Runtime、Ascend Device Plugin和Volcano组件，请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
-   安装Docker，请参考[安装Docker](https://docs.docker.com/engine/install/)。
-   安装Kubernetes，请参见[安装Kubernetes](https://kubernetes.io/zh/docs/setup/production-environment/tools/)。

## 基础调度<a name="ZH-CN_TOPIC_0000002511346871"></a>







### 整卡调度<a name="ZH-CN_TOPIC_0000002479386926"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的整张NPU卡上，独占整张卡执行训练或者推理任务。整卡调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，根据NPU设备物理拓扑，选择合适的NPU设备，最大化发挥NPU性能，实现训练或者推理任务的NPU卡的调度和其他资源的最佳分配。

使用集群调度组件提供的Volcano组件，可以实现交换机亲和性调度和昇腾AI处理器亲和性调度。Volcano是基于昇腾AI处理器的互联拓扑结构和处理逻辑，实现了昇腾AI处理器最佳利用的调度器组件，可以最大化发挥昇腾AI处理器计算性能。关于交换机亲和性调度和昇腾AI处理器亲和性调度的详细说明，可以参见[亲和性调度](./references.md#方案介绍)。

**所需组件<a name="section15655185785119"></a>**

-   调度器（Volcano或其他调度器）
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   ClusterD
-   NodeD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[整卡调度或静态vNPU调度（训练）](./usage/basic_scheduling.md#整卡调度或静态vnpu调度训练)章节进行操作。

### 静态vNPU调度<a name="ZH-CN_TOPIC_0000002511426831"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的vNPU卡上，使用vNPU执行训练或者推理任务。静态vNPU调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，实现训练或者推理任务的vNPU卡的调度和其他资源的最佳分配。

**使用须知<a name="section4448516137"></a>**

使用静态vNPU调度前，用户需要通过npu-smi工具提前创建多个vNPU（虚拟NPU），当用户需要使用vNPU资源时，需要将vNPU挂载到容器中使用。使用算力虚拟化需要了解昇腾AI处理器支持的芯片类型、切分规则和切分模板等，详细信息请参见[虚拟化实例](./usage/virtual_instance.md)。

**所需组件<a name="section15655185785119"></a>**

训练任务及推理任务下需要安装以下组件

-   调度器（Volcano或其他调度器）
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   ClusterD
-   NodeD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[整卡调度或静态vNPU调度（训练）](./usage/basic_scheduling.md#整卡调度或静态vnpu调度训练)章节进行操作。

### 动态vNPU调度<a name="ZH-CN_TOPIC_0000002479226956"></a>

**功能特点<a name="section1788818281655"></a>**

动态vNPU调度需要Ascend Device Plugin组件上报其所在节点的可用AI Core数目。虚拟化任务上报后，Volcano经过计算将该任务调度到满足其要求的节点。该节点的Ascend Device Plugin在收到请求后自动切分出vNPU设备并挂载该任务，从而完成整个动态虚拟化过程。该过程不需要用户提前切分vNPU，在任务使用完成后又能自动回收，支持用户算力需求不断变化的场景。

**使用须知<a name="section4448516137"></a>**

使用动态vNPU调度前，用户需要提前了解昇腾AI处理器支持的芯片类型、切分规则和切分模板等，详细信息请参见[虚拟化实例](./usage/virtual_instance.md)。

**所需组件<a name="section15655185785119"></a>**

-   Volcano
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   ClusterD
-   NodeD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[动态vNPU调度（推理）](./usage/basic_scheduling.md#动态vnpu调度推理)章节进行操作。

### 弹性训练<a name="ZH-CN_TOPIC_0000002479226936"></a>

>[!NOTE] 说明 
>本章节描述的是基于Resilience Controller组件的弹性训练，该组件已经日落，相关资料将于2026年的8.2.RC1版本删除。最新的弹性训练能力请参见[弹性训练](./usage/resumable_training.md#弹性训练)。

**功能特点<a name="section1788818281655"></a>**

训练节点出现故障后，集群调度组件将对故障节点进行隔离，并根据任务预设的规模和当前集群中可用的节点数，重新设置任务副本数，然后进行重调度和重训练（需进行脚本适配）。

**所需组件<a name="section15655185785119"></a>**

-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   Volcano
-   NodeD
-   Resilience Controller
-   ClusterD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[弹性训练](./usage/basic_scheduling.md#弹性训练)章节进行操作。

### 推理卡故障恢复<a name="ZH-CN_TOPIC_0000002479226952"></a>

**功能特点<a name="section113779818313"></a>**

集群调度组件管理的推理NPU资源出现故障后，将对故障资源（对应NPU）进行热复位操作，使NPU恢复健康。

**所需组件<a name="section143231032154719"></a>**

-   调度器（Volcano或其他调度器）
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   ClusterD
-   NodeD

**使用说明<a name="section74221327111220"></a>**

-   安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
-   特性使用指导请参考[推理卡故障恢复](./usage/basic_scheduling.md#推理卡故障恢复)章节进行操作。

### 推理卡故障重调度<a name="ZH-CN_TOPIC_0000002511346875"></a>

**功能特点<a name="section119259203315"></a>**

集群调度组件管理的推理NPU资源出现故障后，集群调度组件将对故障资源（对应NPU）进行隔离并自动进行重调度。

**所需组件<a name="section15655185785119"></a>**

-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   Volcano
-   ClusterD
-   NodeD

**使用说明<a name="section18894171918127"></a>**

-   安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
-   特性使用指导请参考[推理卡故障重调度](./usage/basic_scheduling.md#推理卡故障重调度)章节进行操作。

## 断点续训<a name="ZH-CN_TOPIC_0000002511346867"></a>

**功能特点<a name="section1788818281655"></a>**

当训练任务出现故障时，将任务重调度到健康设备上继续训练或者对故障芯片进行自动恢复。

-   **故障检测**：通过Ascend Device Plugin、Volcano、ClusterD和NodeD四个组件，发现任务故障。
-   **故障处理**：故障发生后，根据上报的故障信息进行故障处理。分为以下两种模式。
    -   **重调度模式**：故障发生后将任务重调度到其他健康设备上继续运行。
    -   **优雅容错模式**：当训练时芯片出现故障后，系统将尝试对故障芯片进行自动恢复。

-   **训练恢复**：在任务重新调度之后，训练任务会使用故障前自动保存的CKPT，重新拉起训练任务继续训练。

**所需组件<a name="section15655185785119"></a>**

-   Volcano
-   Ascend Operator
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   NodeD
-   ClusterD
-   TaskD
-   MindIO ACP（可选）
-   MindIO TFT（可选）

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[断点续训](./usage/resumable_training.md)章节进行操作。
3.  TaskD需安装在容器内，详见[制作镜像](./usage/resumable_training.md#制作镜像)章节。
4.  MindIO ACP的详细介绍及安装步骤请参见[Checkpoint保存与加载优化](./references.md#checkpoint保存与加载优化)章节。
5.  MindIO TFT的详细介绍及安装步骤请参见[故障恢复与加速](./references.md#故障恢复加速)。

## 容器恢复<a name="ZH-CN_TOPIC_0000002492192948"></a>

**功能特点<a name="section1788818281655"></a>**

在无K8s的场景下，训练或推理进程异常后，通过配置容器恢复功能，可以进行容器故障恢复。

-   **故障检测**：通过Container Manager组件，发现任务故障。
-   **故障处理**：故障发生后，不需要人工介入就可自动恢复故障设备。
-   **容器恢复**：故障发生时，将容器停止，故障恢复后重新将容器拉起。

**所需组件<a name="section15655185785119"></a>**

Container Manager

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[一体机特性指南](./usage/appliance.md)章节进行操作。

