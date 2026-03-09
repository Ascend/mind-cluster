# 简介

## 概述<a name="ZH-CN_TOPIC_0000002511426835"></a>

集群调度组件基于业界流行的集群调度系统Kubernetes，增加了昇腾AI处理器（NPU）的支持，提供NPU资源管理、优化调度和分布式训练集合通信配置等基础功能。深度学习平台开发厂商可以有效减少底层资源调度相关软件开发工作量，使能用户基于MindCluster快速开发深度学习平台。

本文档是用户使用集群调度组件的指导文档，在安装和使用集群调度组件前，用户需要提前了解[集群调度组件的特性](#特性介绍)，并根据具体特性的特点和功能，选择需要使用的特性并[安装相应的组件](./installation_guide.md#安装部署)。

**使用流程<a name="section10118105218514"></a>**

集群调度组件的安装和使用流程如下图所示。

![](../figures/scheduling/zh-cn_image_0000002511426865.png)

**表 1**  使用流程

<a name="table475516228316"></a>
|步骤|描述|
|--|--|
|选择特性|集群调度组件支持训练任务和推理任务的多种特性。每种特性所需要的组件不同，组件的配置也各不相同。用户可以根据需要，选择相应的特性进行使用，支持多个特性同时使用。|
|安装相应组件|在选择特性后，需要安装相应的组件。组件的安装支持手动安装和使用工具安装。|
|使用示例参考|集群调度组件为用户提供全流程的特性使用示例，包括训练任务示例和推理任务示例。示例中包含集群调度组件支持的框架、模型和相应的脚本适配操作，帮助用户更好地了解和使用集群调度组件。|

**免责声明<a name="section7267115610496"></a>**

-   本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。
-   用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。

## 组件介绍<a name="ZH-CN_TOPIC_0000002479386906"></a>

### Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002511426843"></a>

**应用场景<a name="section15761025111720"></a>**

创建容器时，为了容器内部能够正常使用昇腾AI处理器，需要引入昇腾驱动相关的脚本和命令。这些脚本和命令分布在不同的文件中，且存在变更的可能性。为了避免容器创建时冗长的文件挂载，MindCluster提供了部署在计算节点上的Ascend Docker Runtime组件。通过输入需要挂载的昇腾AI处理器编号，即可完成昇腾AI处理器及相关驱动的文件挂载。

**组件功能<a name="section586382712395"></a>**

-   提供Docker或Containerd的昇腾容器化支持，自动挂载所需文件和设备依赖。
-   部分硬件形态支持输入vNPU信息，完成vNPU的创建和销毁。

**组件上下游依赖<a name="section10767161681"></a>**

Ascend Docker Runtime逻辑接口如[图1](#fig98811251715)所示。

**图 1**  组件上下游依赖<a name="fig98811251715"></a>  
![](../figures/scheduling/组件上下游依赖.png "组件上下游依赖")

### NPU Exporter<a name="ZH-CN_TOPIC_0000002479226948"></a>

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

### Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002479226928"></a>

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

### Volcano<a name="ZH-CN_TOPIC_0000002479386902"></a>

**应用场景<a name="section15761025111720"></a>**

K8s基础调度仅能通过感知昇腾芯片的数量进行资源调度。为实现亲和性调度，最大化资源利用，需要感知昇腾芯片之间的网络连接方式，选择网络最优的资源。MindCluster提供了部署在管理节点的Volcano服务，针对不同的昇腾设备和组网方式提供网络亲和性调度。

**组件功能<a name="section1112014512117"></a>**

-   根据集群调度底层组件上报的故障信息及节点信息计算集群的可用设备信息。（self-maintain-available-card默认开启。self-maintain-available-card关闭的情况下，从集群调度底层组件获取集群的可用设备信息。）
-   从K8s的任务对象中获取用户期望的资源数量，结合集群的设备数量、设备类型和设备组网方式，选择最优资源分配给任务。
-   任务资源故障时，重新调度任务。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1**  组件上下游依赖<a name="fig1383773934815"></a>  
![](../figures/scheduling/组件上下游依赖-2.png "组件上下游依赖-2")

1.  根据ClusterD上报的信息计算集群资源信息。此为默认使用ClusterD的场景。
2.  接收第三方下发的任务拉起配置，根据集群资源信息，选择最优节点资源。
3.  向计算节点的Ascend Device Plugin传递具体的资源选中信息，完成设备挂载。

### ClusterD<a name="ZH-CN_TOPIC_0000002511346859"></a>

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

### Ascend Operator<a name="ZH-CN_TOPIC_0000002511426817"></a>

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

### NodeD<a name="ZH-CN_TOPIC_0000002479386924"></a>

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

### Resilience Controller<a name="ZH-CN_TOPIC_0000002511426827"></a>

>[!NOTE]
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

### Elastic Agent<a name="ZH-CN_TOPIC_0000002479386918"></a>

>[!NOTE]
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

### TaskD<a name="ZH-CN_TOPIC_0000002479386914"></a>

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

### MindIO ACP<a name="ZH-CN_TOPIC_0000002479226942"></a>

**组件应用场景<a name="section15761025111720"></a>**

Checkpoint是模型中断训练后恢复的关键点，Checkpoint的密集程度、保存和恢复的性能较为关键，它可以提高训练系统的有效吞吐率。MindIO ACP针对Checkpoint的加速方案，支持昇腾产品在LLM模型领域扩展市场空间。

**组件功能<a name="section1112014512117"></a>**

在大模型训练中，使用训练服务器内存作为缓存，对Checkpoint的保存及加载进行加速。

**组件上下游依赖<a name="section4941922192110"></a>**

**图 1** MindIO ACP<a name="fig24667426549"></a>  
![](../figures/scheduling/MindIO-ACP.png "MindIO-ACP")

### MindIO TFT<a name="ZH-CN_TOPIC_0000002511426847"></a>

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

### Container Manager<a name="ZH-CN_TOPIC_0000002524312655"></a>

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

## Infer Operator

**应用场景**

MindCluster提供Infer Operator组件，根据推理服务的实例配置，拉起推理服务，并支持推理实例的手动扩缩容。

**组件功能**

- 创建推理实例Workload与Service。
- 推理实例的手动扩缩容。

**组件上下游依赖**

**图 1**  组件上下游依赖<a name="fig107831859288"></a>  
![](../figures/scheduling/组件上下游依赖-8.png "组件上下游依赖-8")

1. 基于用户配置的任务yaml创建推理实例Workload。
2. Workload Controller创建Pod后，Volcano进行资源的最终选定。
3. 若Workload申请占用NPU卡，Ascend Device Plugin获取NPU信息，完成设备的挂载。


# 特性介绍<a name="ZH-CN_TOPIC_0000002511426839"></a>

### 容器化支持<a name="ZH-CN_TOPIC_0000002479386930"></a>

**功能特点<a name="section1788818281655"></a>**

为所有的训练或推理作业提供NPU容器化支持，自动挂载所需文件和设备依赖，使用户AI作业能够以Docker容器的方式平滑运行在昇腾设备之上。

**所需组件<a name="section15655185785119"></a>**

Ascend Docker Runtime

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[容器化支持](./usage/containerization.md)章节进行操作。

### 资源监测<a name="ZH-CN_TOPIC_0000002479386910"></a>

**功能特点<a name="section1788818281655"></a>**

支持在执行训练或者推理任务时，对昇腾AI处理器资源各种数据信息的实时监测，可实时获取昇腾AI处理器利用率、温度、电压、内存，以及昇腾AI处理器在容器中的分配状况等信息，实现资源的实时监测。支持对虚拟NPU（vNPU）的AI Core利用率、vNPU总内存和vNPU使用中内存进行监测。目前NPU Exporter仅支持对Atlas 推理系列产品的vNPU资源监测。

**所需组件<a name="section15655185785119"></a>**

NPU Exporter

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[资源监测](./usage/resource_monitoring.md)章节进行操作。

### 虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855"></a>

#### 基于HDK的虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855hdk"></a>

**功能介绍<a name="section1337420477275"></a>**

基于HDK的昇腾虚拟化实例功能是指通过资源虚拟化的方式将物理机或虚拟机配置的NPU（昇腾AI处理器）切分成若干份vNPU（虚拟NPU）挂载到容器中使用，虚拟化管理方式能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放的资源操作请求。

昇腾基于HDK的虚拟化实例功能的优点是可实现多个用户共同使用一台服务器，用户可以按需申请vNPU，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的平稳和安全，且资源分配，资源回收过程统一，方便多租户管理。

**所需组件<a name="ZH-CN_TOPIC_0000002479226932"></a>**

根据创建或挂载vNPU的方式不同，所需组件不同，可以参考如下内容。

创建vNPU所需组件：

创建vNPU有以下两种方式。

- 静态虚拟化：通过npu-smi工具**手动**创建多个vNPU。
- 动态虚拟化：通过MindCluster中的以下组件创建vNPU。
    - 方式一：通过Ascend Docker Runtime**手动**创建vNPU，容器进程结束时，自动销毁vNPU。
    - 方式二：通过Volcano和Ascend Device Plugin动态地**自动**创建vNPU，容器进程结束时，自动销毁vNPU。

挂载vNPU所需组件：

根据创建vNPU的方式的不同，将vNPU挂载到容器的方式也不同，说明如下：

-   基于原生Docker挂载vNPU（只支持静态虚拟化）
-   基于MindCluster组件挂载vNPU（支持静态虚拟化和动态虚拟化）
    -   方式一：通过Ascend Docker Runtime+Docker方式挂载vNPU（此方式相比只使用原生Docker易用性更高）。
    -   方式二：通过Kubernetes挂载vNPU。

**使用说明<a name="section1350915844811"></a>**

- 驱动安装后会默认安装npu-smi工具，安装操作请参考《CANN 软件安装指南》中的“<a href="https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian">安装NPU驱动和固件</a>”章节（商用版）或“<a href="https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=openEuler">安装NPU驱动和固件</a>”章节（社区版）；安装成功后，npu-smi放置在“/usr/local/sbin/”和“/usr/local/bin/”路径下。
- 安装MindCluster中的Ascend Docker Runtime、Ascend Device Plugin和Volcano组件，请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
- 安装Docker，请参考[安装Docker](https://docs.docker.com/engine/install/)。
- 安装Kubernetes，请参见[安装Kubernetes](https://kubernetes.io/zh/docs/setup/production-environment/tools/)。

#### 基于VCANN的虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855vcann"></a>

**功能介绍<a name="section1337420477275vcann"></a>**

基于VCANN的虚拟化实例功能是指通过向VCANN提供软切分配置文件的方式将物理机配置的NPU（昇腾AI处理器）挂载到容器中使用，虚拟化管理方式能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放资源的操作请求。

昇腾基于VCANN的虚拟化实例功能的优点是可实现多个用户共同使用一台服务器，用户可以按需申请NPU的资源，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的平稳和安全，且资源分配与回收过程统一，从而方便多租户管理。

**所需组件<a name="ZH-CN_TOPIC_0000002479226932vcann"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**使用说明<a name="section1350915844811vcann"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[软切分调度（推理）](./usage/basic_scheduling.md#软切分调度推理)章节进行操作。

### 基础调度<a name="ZH-CN_TOPIC_0000002511346871"></a>

#### 整卡调度<a name="ZH-CN_TOPIC_0000002479386926"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的整张NPU卡上，独占整张卡执行训练或者推理任务。整卡调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，根据NPU设备物理拓扑，选择合适的NPU设备，最大化发挥NPU性能，实现训练或者推理任务的NPU卡的调度和其他资源的最佳分配。

使用集群调度组件提供的Volcano组件，可以实现交换机亲和性调度和昇腾AI处理器亲和性调度。Volcano是基于昇腾AI处理器的互联拓扑结构和处理逻辑，实现了昇腾AI处理器最佳利用的调度器组件，可以最大化发挥昇腾AI处理器计算性能。关于交换机亲和性调度和昇腾AI处理器亲和性调度的详细说明，可以参见[亲和性调度](./references.md#方案介绍)。

**所需组件<a name="section15655185785119"></a>**

-   调度器（Volcano或其他调度器）
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   Infer Operator
-   ClusterD
-   NodeD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[整卡调度或静态vNPU调度（训练）](./usage/basic_scheduling.md#整卡调度或静态vnpu调度训练)章节进行操作。

#### 静态vNPU调度<a name="ZH-CN_TOPIC_0000002511426831"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的vNPU卡上，使用vNPU执行训练或者推理任务。静态vNPU调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，实现训练或者推理任务的vNPU卡的调度和其他资源的最佳分配。

**所需组件<a name="section15655185785119"></a>**

训练任务及推理任务下需要安装以下组件

-   调度器（Volcano或其他调度器）
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   Infer Operator
-   ClusterD
-   NodeD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[整卡调度或静态vNPU调度（训练）](./usage/basic_scheduling.md#整卡调度或静态vnpu调度训练)章节进行操作。

#### 动态vNPU调度<a name="ZH-CN_TOPIC_0000002479226956"></a>

**功能特点<a name="section1788818281655"></a>**

动态vNPU调度需要Ascend Device Plugin组件上报其所在节点的可用AI Core数目。虚拟化任务上报后，Volcano经过计算将该任务调度到满足其要求的节点。该节点的Ascend Device Plugin在收到请求后自动切分出vNPU设备并挂载该任务，从而完成整个动态虚拟化过程。该过程不需要用户提前切分vNPU，在任务使用完成后又能自动回收，支持用户算力需求不断变化的场景。

**所需组件<a name="section15655185785119"></a>**

-   Volcano
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   ClusterD
-   NodeD

**使用说明<a name="section1245612501584"></a>**

1.  安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
2.  特性使用指导请参考[动态vNPU调度（推理）](./usage/basic_scheduling.md#动态vnpu调度推理)章节进行操作。

#### 弹性训练<a name="ZH-CN_TOPIC_0000002479226936"></a>

>[!NOTE]
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

#### 推理卡故障恢复<a name="ZH-CN_TOPIC_0000002479226952"></a>

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

#### 推理卡故障重调度<a name="ZH-CN_TOPIC_0000002511346875"></a>

**功能特点<a name="section119259203315"></a>**

集群调度组件管理的推理NPU资源出现故障后，集群调度组件将对故障资源（对应NPU）进行隔离并自动进行重调度。

**所需组件<a name="section15655185785119"></a>**

-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   Infer Operator
-   Volcano
-   ClusterD
-   NodeD

**使用说明<a name="section18894171918127"></a>**

-   安装组件请参考[安装部署](./installation_guide.md#安装部署)章节进行操作。
-   特性使用指导请参考[推理卡故障重调度](./usage/basic_scheduling.md#推理卡故障重调度)章节进行操作。

### 断点续训<a name="ZH-CN_TOPIC_0000002511346867"></a>

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
5.  MindIO TFT的详细介绍及安装步骤请参见[故障恢复加速](./references.md#故障恢复加速)。

### 容器恢复<a name="ZH-CN_TOPIC_0000002492192948"></a>

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
