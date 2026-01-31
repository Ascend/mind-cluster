# 基础调度特性指南<a name="ZH-CN_TOPIC_0000002511346993"></a>

## 特性说明<a name="ZH-CN_TOPIC_0000002511347091"></a>

基础调度包含如下特性：

-   训练任务：[整卡调度](../introduction.md#整卡调度)、[静态vNPU调度](../introduction.md#静态vnpu调度)和[弹性训练](../introduction.md#弹性训练)。若使用断点续训请参见[断点续训](../usage/resumable_training.md)。
-   推理任务：[整卡调度](../introduction.md#整卡调度)、[静态vNPU调度](../introduction.md#静态vnpu调度)、[动态vNPU调度](../introduction.md#动态vnpu调度)、[推理卡故障恢复](../introduction.md#推理卡故障恢复)和[推理卡故障重调度](../introduction.md#推理卡故障重调度)。

    不同的特性依赖不同的组件，详细介绍请参见[基础调度](../introduction.md#基础调度)章节。

本文档演示如何基于某模型部署并执行使用NPU的训练或推理任务。生产环境与示例存在差异，本章节内示例仅做参考，用户需要根据实际生产环境做修改。

**任务类型<a name="section14151030191813"></a>**

Ascend Operator提供以下2种方式配置资源信息：

-   通过环境变量配置资源信息：为不同AI框架的分布式训练任务提供相应的环境变量，请参见[环境变量说明](../appendix.md#环境变量说明)中"Ascend Operator环境变量说明"。使用此方式的用户仅支持创建Ascend Job（以下简称acjob）对象。
-   通过文件配置资源信息：训练任务集合通信配置文件（RankTable File，也叫[hccl.json](../appendix.md#hccljson文件说明)）。使用此方式的用户支持创建以下3种类型的对象：Volcano Job（以下简称vcjob）、Ascend Job（以下简称acjob）和Deployment（以下简称deploy）。
    -   （推荐）Ascend Job：简称acjob，是MindCluster自定义的一种任务类型，当前支持通过环境变量配置资源信息及文件配置资源信息这2种方式拉起训练或推理任务。

        每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../appendix.md#acjob关键字段说明)。

    -   Volcano Job：简称vcjob，适用于批处理任务，任务有完成状态。
    -   Deployment：简称deploy，适用于后台常驻任务，任务没有完成状态。在需要持续训练任务、持续占用资源，调试训练任务，或者提供推理服务接口的时候选用。

        >[!NOTE] 说明 
        >不支持Deployment的更新操作，如果需要更新，请先删除再创建。

**调度时间说明<a name="section12177114564719"></a>**

Volcano在多任务或者单任务场景下，在Atlas 800T A2 训练服务器设备上acjob任务的调度参考时间说明如下。若要达到以下参考时间，需要确保CPU的频率至少为2.60GHz，API Server时延不超过80毫秒。其中调度时间是指任务下发到Pod状态为Running的时间。

-   多任务调度时间说明。
    -   并发创建多个单机单卡任务数量的峰值为100个，即用100个任务YAML同时创建100个单机单卡任务，这100个单机单卡任务的调度时间为107秒。
    -   每秒稳定创建单机单卡任务数为5个，连续稳定创建1分钟后，可以创建300个单机单卡任务，这300个单机单卡任务的调度时间为293秒。

-   单任务调度时间说明如[表1](#table18378013481)所示。

    **表 1**  单任务多Pod调度说明

    <a name="table18378013481"></a>
    <table><thead align="left"><tr id="row2083715012487"><th class="cellrowborder" valign="top" width="29.81%" id="mcps1.2.4.1.1"><p id="p883712024819"><a name="p883712024819"></a><a name="p883712024819"></a>集群节点数</p>
    </th>
    <th class="cellrowborder" valign="top" width="23.93%" id="mcps1.2.4.1.2"><p id="p767633115568"><a name="p767633115568"></a><a name="p767633115568"></a>Pod数量</p>
    </th>
    <th class="cellrowborder" valign="top" width="46.26%" id="mcps1.2.4.1.3"><p id="p18389064813"><a name="p18389064813"></a><a name="p18389064813"></a>调度时间</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row2435111731111"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p5435417191115"><a name="p5435417191115"></a><a name="p5435417191115"></a>100</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p1951444875619"><a name="p1951444875619"></a><a name="p1951444875619"></a>100</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p54368172115"><a name="p54368172115"></a><a name="p54368172115"></a>14秒</p>
    </td>
    </tr>
    <tr id="row1783811034814"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p9838180184817"><a name="p9838180184817"></a><a name="p9838180184817"></a>500</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p8514104855614"><a name="p8514104855614"></a><a name="p8514104855614"></a>500</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p12838509482"><a name="p12838509482"></a><a name="p12838509482"></a>57秒</p>
    </td>
    </tr>
    <tr id="row13838801481"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p283840194818"><a name="p283840194818"></a><a name="p283840194818"></a>1000</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p0514164875617"><a name="p0514164875617"></a><a name="p0514164875617"></a>1000</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p683830204819"><a name="p683830204819"></a><a name="p683830204819"></a>114秒</p>
    </td>
    </tr>
    <tr id="row1583813013482"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p983811012482"><a name="p983811012482"></a><a name="p983811012482"></a>2000</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p051418481560"><a name="p051418481560"></a><a name="p051418481560"></a>2000</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p17838110174818"><a name="p17838110174818"></a><a name="p17838110174818"></a>228秒</p>
    </td>
    </tr>
    <tr id="row1883860174816"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p198381504486"><a name="p198381504486"></a><a name="p198381504486"></a>3000</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p16514134835618"><a name="p16514134835618"></a><a name="p16514134835618"></a>3000</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p138388012486"><a name="p138388012486"></a><a name="p138388012486"></a>269秒</p>
    </td>
    </tr>
    <tr id="row108384024817"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p10838160124814"><a name="p10838160124814"></a><a name="p10838160124814"></a>4000</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p7514248135611"><a name="p7514248135611"></a><a name="p7514248135611"></a>4000</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p583817016489"><a name="p583817016489"></a><a name="p583817016489"></a>300秒</p>
    </td>
    </tr>
    <tr id="row09919300814"><td class="cellrowborder" valign="top" width="29.81%" headers="mcps1.2.4.1.1 "><p id="p191005301585"><a name="p191005301585"></a><a name="p191005301585"></a>5000</p>
    </td>
    <td class="cellrowborder" valign="top" width="23.93%" headers="mcps1.2.4.1.2 "><p id="p3515248195612"><a name="p3515248195612"></a><a name="p3515248195612"></a>5000</p>
    </td>
    <td class="cellrowborder" valign="top" width="46.26%" headers="mcps1.2.4.1.3 "><p id="p2010020302814"><a name="p2010020302814"></a><a name="p2010020302814"></a>400秒</p>
    </td>
    </tr>
    <tr id="row1725016321888"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p118004408812"><a name="p118004408812"></a><a name="p118004408812"></a>注：</p>
    <a name="ul4483154173211"></a><a name="ul4483154173211"></a><ul id="ul4483154173211"><li>单任务多Pod场景即用1个任务YAML创建多个Pod，比如1个任务YAML创建100个Pod，这100个Pod分别调度到100个节点上的调度时间为14秒。</li><li>若想要达到4000或5000节点的优化调度参考时间，需要参见<a href="../installation_guide.md#安装volcano">安装Volcano</a>中调度时间性能调优步骤进行相应修改。</li><li>当前vcjob任务的调度规格最大支持1000节点。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>


## 昇腾AI处理器的调度流程<a name="ZH-CN_TOPIC_0000002511427051"></a>

整体调度逻辑如下所示，其中Ascend Device Plugin的功能是发现昇腾AI处理器资源并上报；Volcano组件为华为在开源Volcano框架上进行适配修改的调度器。

**调度流程<a name="section7296392473"></a>**

-   **调度流程1**

    **图 1**  调度流程1<a name="fig63404531065"></a>  
    ![](../../figures/scheduling/调度流程1.png "调度流程1")

    默认情况下，Volcano启动YAML中self-maintain-available-card参数的值为true。昇腾AI处理器的调度流程如下所示：

    1.  Ascend Device Plugin组件上报昇腾AI处理器健康状态。
    2.  用户调用kube-apiserver创建使用NPU的业务容器，如vcjob。
    3.  Volcano组件通过节点信息和ConfigMap信息计算当前可用的昇腾AI处理器。
    4.  Volcano组件根据亲和性调度原则，将昇腾AI处理器分配的情况写入Pod的Annotations字段中，同时写入分配时的时间戳。Volcano组件写入资源信息后向Kubernetes提交绑定Pod申请。
    5.  在每个信息上报周期，Ascend Device Plugin从Pod的Annotations中读取芯片的挂载信息。如需修正，则通过kube-apiserver更新回Pod的Annotation中。修正的Annotation包括：huawei.com/资源名、huawei.com/AscendReal、ascend.kubectl.kubernetes.io/ascend-910-configuration。
    6.  kubelet监测到有Pod调度到自己所在节点，调用Ascend Device Plugin的Allocate函数挂载NPU设备。同时也支持使用Ascend Docker Runtime挂载NPU设备。
    7.  Ascend Device Plugin查询当前所在的Node中处于Pending状态的Pod列表，得到亲和性调度后时间戳最小的Pod，获取挂载的device ID，反馈给kubelet进行设备挂载。

-   **调度流程2**

    **图 2**  调度流程2<a name="fig39301952134114"></a>  
    ![](../../figures/scheduling/调度流程2.png "调度流程2")

    如果Volcano启动YAML中的self-maintain-available-card参数的值配置为false，昇腾AI处理器的调度流程如下所示：

    1.  Ascend Device Plugin组件上报昇腾AI处理器健康状态。
    2.  Ascend Device Plugin通过kube-apiserver将当前空闲的昇腾AI处理器（健康昇腾AI处理器  - 已使用的昇腾AI处理器）信息写到ConfigMap“mindx-dl-deviceinfo-\{_nodeName_\}”的“DeviceInfo“字段中。
    3.  用户调用kube-apiserver创建使用NPU的业务容器，如vcjob。
    4.  Volcano组件通过“DeviceInfo“获取当前可用的昇腾AI处理器。
    5.  Volcano组件根据亲和性调度原则，将昇腾AI处理器分配的情况写入Pod的“Annotations“字段中，同时写入分配时的时间戳。Volcano组件写入资源信息后向Kubernetes提交绑定Pod申请。
    6.  kubelet监测到有Pod调度到自己所在节点，调用Ascend Device Plugin的Allocate函数挂载NPU设备。同时也支持使用Ascend Docker Runtime挂载NPU设备。
    7.  Ascend Device Plugin查询当前所在的Node中处于Pending状态的Pod列表，得到亲和性调度后、时间戳最小的Pod，获取挂载的device ID，反馈给kubelet进行设备挂载。
    8.  Ascend Device Plugin更新“DeviceInfo“字段中的可分配昇腾AI处理器。

**具体交互字段说明<a name="section154080418522"></a>**

1.  Ascend Device Plugin（开源代码版本）以ConfigMap形式上报节点资源，上报资源的形式为“huawei.com/资源名：资源名+物理ID”。格式如[图3](#fig83207421331)所示。图中标出部分表示可用昇腾AI处理器列表，是全部的健康昇腾AI处理器减去被Volcano分配的昇腾AI处理器。全部的健康昇腾AI处理器信息通过调用NPU驱动接口获取，而被Volcano分配的芯片是通过遍历当前Node上所有满足条件的Pod，即Pod的状态为非Failed或者Succeeded，且Pod的“Annotations“字段上有Volcano分配的昇腾AI处理器信息。

    >[!NOTE] 说明 
    >-   用户可通过登录后台环境，执行**kubectl describe cm mindx-dl-deviceinfo-_\{__nodeName__\}_  -n kube-system**命令获取上报的资源信息。
    >-   该字段“huawei.com/资源名”正在日落，后续版本该字段不再呈现。默认情况下，节点的可用芯片由Volcano维护，该字段不生效。如果需要生效，可以修改Volcano的配置参数“self-maintain-available-card”值为false。

    **图 3**  节点NPU资源信息<a name="fig83207421331"></a>  
    ![](../../figures/scheduling/节点NPU资源信息.png "节点NPU资源信息")

2.  Volcano组件通过节点信息和ConfigMap信息计算当前可用的昇腾AI处理器。（如果Volcano配置开关self-maintain-available-card关闭，Volcano会以“huawei.com/资源名”为key，读取“DeviceInfo“字段信息作为可用昇腾AI处理器的依据。）根据亲和性调度策略，判断出任务需要的符合亲和性规则的昇腾AI处理器后（即分配给任务的昇腾AI处理器）。Volcano会将分配芯片信息写入任务Pod的“Annotations“，如[图4](#fig29119551778)标出的第一个部分所示；第二个需要写入的字段为“predicate-time“，表示为任务分配资源的当前时间，不需要向可读时间格式做转换，可比较大小即可。kubelet监测到有Pod调度到自己所在节点，调用Device-plugin的Allocate函数挂载NPU设备。

    **图 4**  分配给Pod的NPU信息<a name="fig29119551778"></a>  
    ![](../../figures/scheduling/分配给Pod的NPU信息.png "分配给Pod的NPU信息")

3.  Ascend Device Plugin在收到Allocate请求时（以2卡任务为例），因为Allocate输入的参数是kubelet随机分配的，如[图4](#fig29119551778)中的“huawei.com/kltDev“字段所示，可能是不符合亲和性规则的昇腾AI处理器ID，例如Ascend910-7和Ascend910-0。

    此时Ascend Device Plugin会找到当前Node上所有的满足条件的Pod（Pod的状态为非Failed或者Succeeded），且Pod的“Annotations“字段中存在Volcano写入的分配的昇腾AI处理器ID，昇腾AI处理器数量和kubelet分配昇腾AI处理器数量要一致。

    再从满足条件的Pod中，选择“predicate-time“最小的Pod，并把这个Pod的“predicate-time“改为最大的Uint值（避免下次再选到）。解析Pod的“Annotations“字段，得到Volcano分配的昇腾AI处理器信息，例如Ascend910-0和Ascend910-1，把它们对应的挂载路径等信息返回，并且将真正分配的昇腾AI处理器信息写入到Pod的“Annotations“中的“huawei.com/AscendReal“字段中。


## 整卡调度或静态vNPU调度（训练）<a name="ZH-CN_TOPIC_0000002479387138"></a>

### 使用前必读<a name="ZH-CN_TOPIC_0000002511347093"></a>

**前提条件<a name="section52051339787"></a>**

-   确保环境中有配置相应的存储方案，比如使用NFS（Network File System），用户可以参见[安装NFS](../common_operations.md#安装nfs)进行操作。
-   在使用整卡调度或静态vNPU调度特性前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。
    -   调度器（Volcano或其他调度器）
    -   Ascend Device Plugin
    -   Ascend Docker Runtime
    -   Ascend Operator
    -   ClusterD
    -   NodeD

-   对于训练任务类型为acjob，调度器为Volcano的整卡调度，支持批量创建Pod和批量调度功能。
    -   若要使用批量创建Pod功能，安装Ascend Operator组件时需使用openFuyao定制Kubernetes组件。
    -   若要使用批量调度功能，安装Volcano组件时需使用openFuyao定制Kubernetes和volcano-ext组件，并开启批量调度功能。
    -   批量调度功能适用于超大规模集群场景，在此场景下请根据实际需要扩展MindCluster组件分配的CPU和内存资源，防止MindCluster组件出现性能不足或者超出分配内存使用，导致组件被Kubernetes驱逐。

**使用方式<a name="section179431435174811"></a>**

-   通过命令行使用：整卡调度或静态vNPU调度特性需要使用调度器，用户可以选择使用Volcano调度器和其他调度器。无论选择哪种调度器，都需要使用Ascend Operator组件设置资源信息。
-   集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section577625973520"></a>**

-   资源监测可以和训练场景下的所有特性一起使用。
-   集群中同时跑多个训练任务，每个任务使用的特性可以不同。
-   静态vNPU调度特性需要搭配算力虚拟化特性一起使用，关于静态虚拟化的相关说明和操作请参见[静态虚拟化](./virtual_instance.md#静态虚拟化)章节。

**支持的产品形态<a name="section169961844182917"></a>**

-   支持以下产品使用**整卡调度**。
    -   Atlas 训练系列产品
    -   Atlas A2 训练系列产品
    -   Atlas A3 训练系列产品

-   支持以下产品使用**静态vNPU调度**。

    Atlas 训练系列产品

**使用流程<a name="section5640184231810"></a>**

整卡调度、静态vNPU调度有3种使用场景，分别是通过命令行使用（Volcano）、通过命令行使用（其他调度器）和集成后使用。

通过命令行使用Volcano和其他调度器的使用流程一致。使用其他调度器准备任务YAML需要参考[通过命令行使用（其他调度器）](#通过命令行使用其他调度器)章节创建任务YAML。使用其他调度器的其余操作和使用Volcano一致，可以参考[通过命令行使用（Volcano）](#通过命令行使用volcano)进行操作。

**图 1**  整卡调度和静态vNPU调度使用流程<a name="fig107864120214"></a>  
![](../../figures/scheduling/整卡调度和静态vNPU调度使用流程.png "整卡调度和静态vNPU调度使用流程")

1.  脚本适配时，用户可根据实际情况选择通过环境变量或文件配置资源信息。
2.  在准备任务YAML时，下发的任务YAML需要根据具体的NPU型号，选择不同的YAML进行修改适配。选择YAML时可以参考[准备任务YAML](#准备任务yaml)，根据实际情况选择合适的YAML。


### 实现原理<a name="ZH-CN_TOPIC_0000002479387150"></a>

根据训练任务类型的不同，特性的原理图略有差异。静态vNPU调度需要使用npu-smi工具提前创建好需要的vNPU。

**acjob任务<a name="section9971431567"></a>**

acjob任务原理图如[图1](#fig5188536014)所示。

**图 1**  acjob任务调度原理图<a name="fig5188536014"></a>  
![](../../figures/scheduling/acjob任务调度原理图.png "acjob任务调度原理图")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到节点对象（Node）中。
    -   Ascend Device Plugin定期上报芯片拓扑信息。
        -   上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）、已使用的芯片数量（allocated）和芯片的基础信息（device ip和super\_device\_ip）上报到Node中，用于整卡调度。
        -   上报vNPU相关信息到Node中，用于静态vNPU调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3.  用户通过kubectl或者其他深度学习平台下发acjob任务。
4.  Ascend Operator为任务创建相应的PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  Ascend Operator为任务创建相应的Pod，并在容器中注入集合通信所需环境变量。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
    -   整卡调度写入整卡信息。
    -   静态vNPU调度写入vNPU相关信息。

7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin或volcano-scheduler在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源。
8.  Ascend Operator读取Pod的annotation信息，将相关信息写入hccl.json。
9.  容器读取环境变量或者hccl.json信息，建立通信通道，开始执行训练任务。

**vcjob任务<a name="section13884164615313"></a>**

vcjob任务的原理图如[图2](#fig8717151315416)所示。

**图 2**  vcjob任务调度原理图<a name="fig8717151315416"></a>  
![](../../figures/scheduling/vcjob任务调度原理图.png "vcjob任务调度原理图")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到Node（节点对象）中。
    -   Ascend Device Plugin定期上报芯片拓扑信息。
        -   上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）和已使用的芯片数量（allocated）上报到Node中，用于整卡调度。
        -   上报vNPU相关信息到Node中，用于静态vNPU调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3.  用户通过kubectl或者其他深度学习平台下发vcjob任务。
4.  volcano-controller为任务创建相应PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  当集群资源满足任务要求时，volcano-controller创建任务Pod。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
    -   整卡调度写入整卡信息。
    -   静态vNPU调度写入vNPU相关信息。

7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源，将hccl.json挂载进入容器。
8.  Ascend Operator获取每个Pod的annotation信息，写入hccl.json。
9.  容器读取hccl.json信息，建立通信渠道，开始执行训练任务。

**deploy任务<a name="section32752223579"></a>**

deploy任务原理图如[图3](#fig06571541566)所示。

**图 3**  deploy任务调度原理图<a name="fig06571541566"></a>  
![](../../figures/scheduling/deploy任务调度原理图.png "deploy任务调度原理图")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到Node（节点对象）中。
    -   Ascend Device Plugin定期上报芯片拓扑信息。
        -   上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）和已使用的芯片数量（allocated）上报到Node中，用于整卡调度。
        -   上报vNPU相关信息到Node中，用于静态vNPU调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3.  用户通过kubectl或者其他深度学习平台下发deploy任务。
4.  kube-controller为任务创建相应Pod。
5.  volcano-controller创建任务PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
    -   整卡调度写入整卡信息。
    -   静态vNPU调度写入vNPU相关信息。

7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源，将hccl.json挂载进入容器。
8.  Ascend Operator获取每个Pod的annotation信息，写入hccl.json。
9.  容器读取hccl.json信息，建立通信渠道，开始执行训练任务。


### 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002479227158"></a>

#### 制作镜像<a name="ZH-CN_TOPIC_0000002479227164"></a>

**获取训练镜像<a name="zh-cn_topic_0000001609314597_section971616541059"></a>**

可选择以下方式中的一种来获取训练镜像：

-   （推荐）从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据系统架构（ARM/x86\_64）、模型框架（TensorFlow、PyTorch、MindSpore）下载配套驱动版本的**训练基础镜像**。基于训练基础镜像进行修改，将容器中默认用户修改为root（21.0.4版本之后训练基础镜像默认用户为非root）。基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。
-   从头开始定制用户自己的训练镜像，制作过程请参考[制作镜像](../common_operations.md#制作镜像)中制作容器相关章节。

可将下载/制作的训练基础镜像重命名，如：training:v7.3.0。

**加固镜像<a name="zh-cn_topic_0000001609314597_section8425732111611"></a>**

下载或者制作的训练基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../references.md#容器安全加固)章节进行操作。


#### 脚本适配<a name="ZH-CN_TOPIC_0000002511347097"></a>

##### 通过环境变量配置资源信息<a name="ZH-CN_TOPIC_0000002479387142"></a>

根据模型框架选择对应的指导示例。

-   [TensorFlow](#zh-cn_topic_0000001558834814_section146363219252)
-   [PyTorch](#zh-cn_topic_0000001558834814_section17760205783316)
-   [MindSpore](#zh-cn_topic_0000001558834814_section868111733711)

    >[!NOTE] 说明 
    >-   本节中使用的数据集为[ImageNet2012](https://image-net.org/challenges/LSVRC/2012/2012-downloads.php)数据集（**注：如使用该数据集需遵循数据集提供者的使用规范**）。TensorFlow框架请参考**数据集准备**部分内容进行数据集预处理，详情请参见《TensorFlow 1.15模型迁移指南》的“样例参考\>训练前准备”章节。
    >-   下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
    >-   以下TensorFlow和MindSpore示例需使用CANN 8.5.0之前版本。   

**TensorFlow<a name="zh-cn_topic_0000001558834814_section146363219252"></a>**

1.  <a name="zh-cn_topic_0000001558834814_li1040412108620"></a>下载[TensorFlow代码仓](https://gitee.com/ascend/ModelZoo-TensorFlow/tree/master/TensorFlow2/built-in/cv/image_classification/ResNet50_ID0360_for_TensorFlow2.X)中master分支中的“ResNet50\_ID0360\_for\_TensorFlow2.X”作为训练代码，请根据该模型代码TensorFlow版本选择训练镜像中的TensorFlow版本包。
2.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet\_TF“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/resnet50/imagenet_TF
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# du -sh
        ```

        回显示例如下：

        ```
        42G
        ```

3.  <a name="zh-cn_topic_0000001558834814_li1630573712375"></a>在本地解压[1](#zh-cn_topic_0000001558834814_li1040412108620)中下载的训练代码，将“ModelZoo-TensorFlow-master/TensorFlow2/built-in/cv/image\_classification/“下的“ResNet50\_ID0360\_for\_TensorFlow2.X“目录重命名为“ResNet50\_for\_TensorFlow\_2.6\_code/“目录。
4.  将ResNet50\_for\_TensorFlow\_2.6\_code文件上传至环境的“/data/atlas\_dls/public/code/“路径下。
5.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/without-ranktable/tensorflow“目录中的“train\_start.sh“文件，结合[3](#zh-cn_topic_0000001558834814_li1630573712375)中的“ResNet50\_for\_TensorFlow\_2.6\_code“目录，在host的“/data/atlas\_dls/public/code“路径下，构造如下的目录结构。

    ```
    /data/atlas_dls/public/code/ResNet50_for_TensorFlow_2.6_code/
    ├──  scripts
    │   ├──  train_start.sh
    │    ...
    │        ...
    ├──  tensorflow
    │   ├──  resnet_ctl_imagenet_main.py
    │   ├──  resnet_model.py
    │   ├──  resnet_runnable.py
    │    ...
    │        ...
    ├──  benchmark.sh
    ├──  modelzoo_level.txt
     ...
    └──  requirements.txt
    ```

**PyTorch<a name="zh-cn_topic_0000001558834814_section17760205783316"></a>**

1.  <a name="zh-cn_topic_0000001558834814_li1298552813512"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50\_ID4149\_for\_PyTorch”作为训练代码。
2.  自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# du -sh
        ```

        回显示例如下：

        ```
        11G
        ```

4.  将[1](#zh-cn_topic_0000001558834814_li1298552813512)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50\_ID4149\_for\_PyTorch“目录上传至环境，如“/data/atlas\_dls/public/code/“路径下。
5.  在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch“路径下，注释或删除main.py文件中的加粗字段。

    ```
    def main():
        args = parser.parse_args()
        os.environ['MASTER_ADDR'] = args.addr
        #os.environ['MASTER_PORT'] = '29501'  # 注释或删除该行代码
        if os.getenv('ALLOW_FP32', False) and os.getenv('ALLOW_HF32', False):
            raise RuntimeError('ALLOW_FP32 and ALLOW_HF32 cannot be set at the same time!')
        elif os.getenv('ALLOW_HF32', False):
            torch.npu.conv.allow_hf32 = True
        elif os.getenv('ALLOW_FP32', False):
            torch.npu.conv.allow_hf32 = False
            torch.npu.matmul.allow_hf32 = False
    ```

6.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/without-ranktable/pytorch“目录中的train\_start.sh，在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch/scripts“路径下，构造如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts#
    scripts/
         ├── train_start.sh
    ```

**MindSpore<a name="zh-cn_topic_0000001558834814_section868111733711"></a>**

1.  <a name="zh-cn_topic_0000001558834814_li1141932513379"></a>下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/cv/ResNet)中master分支的“ResNet”代码作为训练代码。
2.  自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/imagenet“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/imagenet
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# du -sh
        ```

        回显示例如下：

        ```
        11G
        ```

4.  在本地解压[1](#zh-cn_topic_0000001558834814_li1141932513379)中下载的训练代码，将“models/official/cv/“下的“ResNet”目录重命名为“ResNet50\_for\_MindSpore\_2.0\_code“。后续步骤以“ResNet50\_for\_MindSpore\_2.0\_code“目录为例。
5.  将ResNet50\_for\_MindSpore\_2.0\_code文件上传至环境“/data/atlas\_dls/public/code/“路径下。
6.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/without-ranktable/mindspore“目录中的“train\_start.sh“文件，结合训练代码中“scripts“目录，在host上构造如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_for_MindSpore_2.0_code/scripts/#
    scripts/
    ├── docker_start.sh
    ├── run_standalone_train_gpu.sh
    ├── run_standalone_train.sh
     ...
    └── train_start.sh
    ```

7.  进入“/data/atlas\_dls/public/code/ResNet50\_for\_MindSpore\_2.0\_code/train.py“目录下，修改train.py对应部分，如下所示。

    ```
     ...
         if config.run_distribute:
             if target == "Ascend":
               #device_id = int(os.getenv('DEVICE_ID', '0'))   #注释该行代码
               #ms.set_context(device_id=device_id)     #注释该行代码
                 ms.set_auto_parallel_context(device_num=config.device_num, parallel_mode=ms.ParallelMode.DATA_PARALLEL,
                                              gradients_mean=True)
                 set_algo_parameters(elementwise_op_strategy_follow=True)
                 if config.net_name == "resnet50" or config.net_name == "se-resnet50":
                     if config.boost_mode not in ["O1", "O2"]:
                         ms.set_auto_parallel_context(all_reduce_fusion_config=config.all_reduce_fusion_config)
                 elif config.net_name in ["resnet101", "resnet152"]:
                     ms.set_auto_parallel_context(all_reduce_fusion_config=config.all_reduce_fusion_config)
                 init()
             # GPU target
     ...
    ```


##### 通过文件配置资源信息<a name="ZH-CN_TOPIC_0000002479387136"></a>

通过文件变量配置资源信息支持创建以下3种类型的对象：acjob、vcjob及deploy。下面将以vcjob和deploy为例，介绍脚本适配的操作示例。

-   [TensorFlow](#zh-cn_topic_0000001558834798_section146363219252)
-   [PyTorch](#zh-cn_topic_0000001558834798_section17760205783316)
-   [MindSpore](#zh-cn_topic_0000001558834798_section868111733711)

>[!NOTE] 说明 
>-   本节中使用的数据集为[ImageNet2012](https://image-net.org/challenges/LSVRC/2012/2012-downloads.php)数据集（**注：如使用该数据集需遵循数据集提供者的使用规范**）。TensorFlow框架请参考**数据集准备**部分内容进行数据集预处理，详情请参见《TensorFlow 1.15模型迁移指南》的“样例参考\>训练前准备”章节。
>-   下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
>-   以下TensorFlow和MindSpore示例需使用CANN 8.5.0之前版本。

**TensorFlow<a name="zh-cn_topic_0000001558834798_section146363219252"></a>**

1.  <a name="zh-cn_topic_0000001558834798_li360413424258"></a>下载[TensorFlow代码仓](https://gitee.com/ascend/ModelZoo-TensorFlow/tree/master/TensorFlow2/built-in/cv/image_classification/ResNet50_ID0360_for_TensorFlow2.X)中master分支中的“ResNet50\_ID0360\_for\_TensorFlow2.X”作为训练代码，请根据该模型代码TensorFlow版本选择训练镜像中的TensorFlow版本包。
2.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet\_TF“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/resnet50/imagenet_TF
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# du -sh
        ```

        回显示例如下：

        ```
        42G
        ```

3.  <a name="zh-cn_topic_0000001558834798_li1630573712375"></a>在本地解压[1](#zh-cn_topic_0000001558834798_li360413424258)中下载的训练代码，将“ModelZoo-TensorFlow-master/TensorFlow2/built-in/cv/image\_classification/“下的“ResNet50\_ID0360\_for\_TensorFlow2.X“目录重命名为“ResNet50\_for\_TensorFlow\_2.6\_code/“目录。
4.  将“ResNet50\_for\_TensorFlow\_2.6\_code“上传至环境的“/data/atlas\_dls/public/code/“路径下。
5.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/ranktable“目录中的“train\_start.sh“、“rank\_table.sh“和“utils.sh“文件，结合[3](#zh-cn_topic_0000001558834798_li1630573712375)中的“ResNet50\_for\_TensorFlow\_2.6\_code“目录，在host的“/data/atlas\_dls/public/code“路径下，构造如下的目录结构。

    ```
    /data/atlas_dls/public/code/ResNet50_for_TensorFlow_2.6_code/
    ├──  scripts
    │   ├──  train_start.sh
    │   ├──  utils.sh
    │   ├──  rank_table.sh
    │    ...
    │        ...
    ├──  tensorflow
    │   ├──  resnet_ctl_imagenet_main.py
    │   ├──  resnet_model.py
    │   ├──  resnet_runnable.py
    │    ...
    │        ...
    ├──  benchmark.sh
    ├──  modelzoo_level.txt
     ...
    └──  requirements.txt
    ```

**PyTorch<a name="zh-cn_topic_0000001558834798_section17760205783316"></a>**

1.  <a name="zh-cn_topic_0000001558834798_li1298552813512"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50\_ID4149\_for\_PyTorch”作为训练代码。
2.  自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# du -sh
        ```

        回显示例如下：

        ```
        11G
        ```

4.  将[1](#zh-cn_topic_0000001558834798_li1298552813512)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50\_ID4149\_for\_PyTorch“目录上传至环境，如“/data/atlas\_dls/public/code/”路径下。
5.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/ranktable“目录中的train\_start.sh、rank\_table.sh和utils.sh文件，在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch/scripts“路径下，构造如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts#
    scripts/
         ├── train_start.sh
         ├── utils.sh
         └── rank_table.sh
    ```

**MindSpore<a name="zh-cn_topic_0000001558834798_section868111733711"></a>**

1.  <a name="zh-cn_topic_0000001558834798_li1141932513379"></a>下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/cv/ResNet)中master分支的“ResNet”代码作为训练代码。
2.  自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/imagenet“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/imagenet
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# du -sh
        ```

        回显示例如下：

        ```
        11G
        ```

4.  在本地解压[1](#zh-cn_topic_0000001558834798_li1141932513379)中下载的训练代码，将“models/official/cv/“下的“ResNet“目录重命名为“ResNet50\_for\_MindSpore\_2.0\_code“。后续步骤以“ResNet50\_for\_MindSpore\_2.0\_code“目录为例。
5.  将ResNet50\_for\_MindSpore\_2.0\_code文件上传至环境“/data/atlas\_dls/public/code/“路径下。
6.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/ranktable“目录中的“train\_start.sh“、“utils.sh“和“rank\_table.sh“文件，结合训练代码中“scripts“目录，在host上构造如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_for_MindSpore_2.0_code/scripts/#
    scripts/
    ├── cache_util.sh
    ├── docker_start.sh
    ├── run_standalone_train_gpu.sh
    ├── run_standalone_train.sh
     ...
    ├── rank_table.sh
    ├── utils.sh
    └── train_start.sh
    ```



#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479227170"></a>

##### 选择YAML示例<a name="ZH-CN_TOPIC_0000002479227150"></a>

集群调度为用户提供YAML示例，用户需要根据使用的组件、芯片类型和任务类型等，选择相应的YAML示例并根据需求进行相应修改后才可使用。

**通过环境变量配置资源信息场景<a name="section1969664932615"></a>**

-   若当前环境使用的是Atlas A2 训练系列产品，选择[表1](#table529015783811)获取相应的YAML示例。

    根据[表1](#table529015783811)获取示例YAML后，Atlas 800T A2 训练服务器、Atlas 200T A2 Box16 异构子框和A200T A3 Box8 超节点服务器可基于[2.3.3.2-表 使用Ascend Job的YAML参数](#yaml参数说明)给出的参数说明进行修改适配。

-   若当前环境使用的是Atlas 训练系列产品，选择[表2](#table18698184918261)获取相应的YAML示例。

    根据[表2](#table18698184918261)获取示例YAML后，服务器（插Atlas 300T 训练卡）可基于Atlas 800 训练服务器的YAML，以及参考[2.3.3.2-表 使用Ascend Job的YAML参数](#yaml参数说明)给出的参数说明进行修改适配。

-   若当前环境使用的是Atlas A3 训练系列产品，选择[表3](#table57051049102614)获取相应的YAML示例。

**表 1** Atlas A2 训练系列产品支持的YAML

<a name="table529015783811"></a>
<table><thead align="left"><tr id="row52903576386"><th class="cellrowborder" valign="top" width="8.8%" id="mcps1.2.7.1.1"><p id="p129019578385"><a name="p129019578385"></a><a name="p129019578385"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.000000000000002%" id="mcps1.2.7.1.2"><p id="p14290115712387"><a name="p14290115712387"></a><a name="p14290115712387"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.650000000000002%" id="mcps1.2.7.1.3"><p id="p1329015723817"><a name="p1329015723817"></a><a name="p1329015723817"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="30.740000000000006%" id="mcps1.2.7.1.4"><p id="p14291125717389"><a name="p14291125717389"></a><a name="p14291125717389"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="18.810000000000002%" id="mcps1.2.7.1.5"><p id="p1129114571381"><a name="p1129114571381"></a><a name="p1129114571381"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.000000000000002%" id="mcps1.2.7.1.6"><p id="p1229110574387"><a name="p1229110574387"></a><a name="p1229110574387"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row13291757163813"><td class="cellrowborder" rowspan="7" valign="top" width="8.8%" headers="mcps1.2.7.1.1 "><p id="p11291115783810"><a name="p11291115783810"></a><a name="p11291115783810"></a>Ascend Job</p>
<p id="p1629145703816"><a name="p1629145703816"></a><a name="p1629145703816"></a></p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="15.000000000000002%" headers="mcps1.2.7.1.2 "><p id="p14227163913366"><a name="p14227163913366"></a><a name="p14227163913366"></a><span id="ph13291155773812"><a name="ph13291155773812"></a><a name="ph13291155773812"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.650000000000002%" headers="mcps1.2.7.1.3 "><p id="p729135713817"><a name="p729135713817"></a><a name="p729135713817"></a><span id="ph1029120577382"><a name="ph1029120577382"></a><a name="ph1029120577382"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.740000000000006%" headers="mcps1.2.7.1.4 "><p id="p52921057113812"><a name="p52921057113812"></a><a name="p52921057113812"></a>tensorflow_multinodes_acjob_<span id="ph0292957193810"><a name="ph0292957193810"></a><a name="ph0292957193810"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.7.1.5 "><p id="p1729235703814"><a name="p1729235703814"></a><a name="p1729235703814"></a>示例默认为双机2卡任务。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="15.000000000000002%" headers="mcps1.2.7.1.6 "><p id="p17292357133814"><a name="p17292357133814"></a><a name="p17292357133814"></a>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
<div class="note" id="note14933145219586"><a name="note14933145219586"></a><a name="note14933145219586"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p1027616512420"><a name="p1027616512420"></a><a name="p1027616512420"></a><span id="ph9014016509"><a name="ph9014016509"></a><a name="ph9014016509"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001519959665_b168254314713"><a name="zh-cn_topic_0000001519959665_b168254314713"></a><a name="zh-cn_topic_0000001519959665_b168254314713"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row829235719380"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p52921579382"><a name="p52921579382"></a><a name="p52921579382"></a><span id="ph1829255713389"><a name="ph1829255713389"></a><a name="ph1829255713389"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1929295713389"><a name="p1929295713389"></a><a name="p1929295713389"></a>pytorch_multinodes_acjob_<span id="ph7292757133817"><a name="ph7292757133817"></a><a name="ph7292757133817"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1129285753810"><a name="p1129285753810"></a><a name="p1129285753810"></a>示例默认为双机2卡任务。</p>
</td>
</tr>
<tr id="row0292357163814"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1529295723813"><a name="p1529295723813"></a><a name="p1529295723813"></a><span id="ph15292125733819"><a name="ph15292125733819"></a><a name="ph15292125733819"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p182921757103818"><a name="p182921757103818"></a><a name="p182921757103818"></a>mindspore_multinodes_acjob_<span id="ph15292205723815"><a name="ph15292205723815"></a><a name="ph15292205723815"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p52921657103816"><a name="p52921657103816"></a><a name="p52921657103816"></a>示例默认为双机16卡任务。</p>
</td>
</tr>
<tr id="row829335719385"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1293195793816"><a name="p1293195793816"></a><a name="p1293195793816"></a><span id="ph14293145733816"><a name="ph14293145733816"></a><a name="ph14293145733816"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p929305793820"><a name="p929305793820"></a><a name="p929305793820"></a>tensorflow_standalone_acjob_<span id="ph529314572388"><a name="ph529314572388"></a><a name="ph529314572388"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.3 "><p id="p2293205715380"><a name="p2293205715380"></a><a name="p2293205715380"></a>示例默认为单机单卡任务。</p>
<p id="p1924194410282"><a name="p1924194410282"></a><a name="p1924194410282"></a></p>
<p id="p13241444192817"><a name="p13241444192817"></a><a name="p13241444192817"></a></p>
</td>
</tr>
<tr id="row751217295286"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2012563417286"><a name="p2012563417286"></a><a name="p2012563417286"></a><span id="ph1512583402811"><a name="ph1512583402811"></a><a name="ph1512583402811"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p61257345281"><a name="p61257345281"></a><a name="p61257345281"></a>mindspore_standalone_acjob_<span id="ph51251834122810"><a name="ph51251834122810"></a><a name="ph51251834122810"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_4"><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a>{xxx}</em></span>b.yaml</p>
</td>
</tr>
<tr id="row429313576389"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.7.1.1 "><p id="p7293457173815"><a name="p7293457173815"></a><a name="p7293457173815"></a><span id="ph2293125773811"><a name="ph2293125773811"></a><a name="ph2293125773811"></a>PyTorch</span></p>
<p id="p1469421012715"><a name="p1469421012715"></a><a name="p1469421012715"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p8293457173812"><a name="p8293457173812"></a><a name="p8293457173812"></a>pytorch_standalone_acjob_<span id="ph10293195714383"><a name="ph10293195714383"></a><a name="ph10293195714383"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b.yaml</p>
</td>
</tr>
<tr id="row7693111092718"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p10694181017275"><a name="p10694181017275"></a><a name="p10694181017275"></a>pytorch_multinodes_acjob_<span id="ph11166172612911"><a name="ph11166172612911"></a><a name="ph11166172612911"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_6"><a name="zh-cn_topic_0000001519959665_i1489729141619_6"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_6"></a>{xxx}</em></span>b_with_ranktable.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1524254417282"><a name="p1524254417282"></a><a name="p1524254417282"></a>示例默认为单机2卡任务。使用<span id="ph747115394291"><a name="ph747115394291"></a><a name="ph747115394291"></a>Ascend Operator</span>组件生成RankTable文件。</p>
</td>
</tr>
</tbody>
</table>

**表 2** Atlas 训练系列产品支持的YAML

<a name="table18698184918261"></a>
<table><thead align="left"><tr id="row6698849162611"><th class="cellrowborder" valign="top" width="10.000000000000002%" id="mcps1.2.7.1.1"><p id="p15698549192614"><a name="p15698549192614"></a><a name="p15698549192614"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.000000000000002%" id="mcps1.2.7.1.2"><p id="p11698849132612"><a name="p11698849132612"></a><a name="p11698849132612"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="12.000000000000002%" id="mcps1.2.7.1.3"><p id="p2698124919262"><a name="p2698124919262"></a><a name="p2698124919262"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="30.000000000000004%" id="mcps1.2.7.1.4"><p id="p069914491269"><a name="p069914491269"></a><a name="p069914491269"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.000000000000004%" id="mcps1.2.7.1.5"><p id="p66993497268"><a name="p66993497268"></a><a name="p66993497268"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="13.000000000000004%" id="mcps1.2.7.1.6"><p id="p3699649192614"><a name="p3699649192614"></a><a name="p3699649192614"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row206991499261"><td class="cellrowborder" rowspan="6" valign="top" width="10.000000000000002%" headers="mcps1.2.7.1.1 "><p id="p10699249132614"><a name="p10699249132614"></a><a name="p10699249132614"></a>Ascend Job</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="15.000000000000002%" headers="mcps1.2.7.1.2 "><p id="p196992049112613"><a name="p196992049112613"></a><a name="p196992049112613"></a><span id="ph76991749122617"><a name="ph76991749122617"></a><a name="ph76991749122617"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="12.000000000000002%" headers="mcps1.2.7.1.3 "><p id="p146991494269"><a name="p146991494269"></a><a name="p146991494269"></a><span id="ph1669934972611"><a name="ph1669934972611"></a><a name="ph1669934972611"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.000000000000004%" headers="mcps1.2.7.1.4 "><p id="p6699104910262"><a name="p6699104910262"></a><a name="p6699104910262"></a>tensorflow_multinodes_acjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="20.000000000000004%" headers="mcps1.2.7.1.5 "><p id="p369954932614"><a name="p369954932614"></a><a name="p369954932614"></a>示例默认为双机8卡任务。</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="13.000000000000004%" headers="mcps1.2.7.1.6 "><p id="p369974917262"><a name="p369974917262"></a><a name="p369974917262"></a>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row669964932617"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p146994498265"><a name="p146994498265"></a><a name="p146994498265"></a><span id="ph3699154922611"><a name="ph3699154922611"></a><a name="ph3699154922611"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p469944911268"><a name="p469944911268"></a><a name="p469944911268"></a>pytorch_multinodes_acjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p269913494266"><a name="p269913494266"></a><a name="p269913494266"></a>示例默认为双机16卡任务。</p>
</td>
</tr>
<tr id="row670044914266"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2070054918265"><a name="p2070054918265"></a><a name="p2070054918265"></a><span id="ph177001649182618"><a name="ph177001649182618"></a><a name="ph177001649182618"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p370024915266"><a name="p370024915266"></a><a name="p370024915266"></a>mindspore_multinodes_acjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p6700104952618"><a name="p6700104952618"></a><a name="p6700104952618"></a>示例默认为双机8卡任务。</p>
<div class="note" id="note170014493266"><a name="note170014493266"></a><a name="note170014493266"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p1370015494264"><a name="p1370015494264"></a><a name="p1370015494264"></a>若下发单机8卡的<span id="ph6700049182610"><a name="ph6700049182610"></a><a name="ph6700049182610"></a>MindSpore</span>任务，需要将mindspore_multinodes_acjob.yaml中minAvailable修改为2，Worker的replicas修改为1。</p>
</div></div>
</td>
</tr>
<tr id="row11700124942615"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1470004911264"><a name="p1470004911264"></a><a name="p1470004911264"></a><span id="ph16700154942618"><a name="ph16700154942618"></a><a name="ph16700154942618"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p15700174913266"><a name="p15700174913266"></a><a name="p15700174913266"></a>tensorflow_standalone_acjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.3 "><p id="p107007497265"><a name="p107007497265"></a><a name="p107007497265"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
<tr id="row15700184992617"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p8700114917265"><a name="p8700114917265"></a><a name="p8700114917265"></a><span id="ph1970044942611"><a name="ph1970044942611"></a><a name="ph1970044942611"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p117007498269"><a name="p117007498269"></a><a name="p117007498269"></a>pytorch_standalone_acjob.yaml</p>
</td>
</tr>
<tr id="row1170074952614"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1970114911261"><a name="p1970114911261"></a><a name="p1970114911261"></a><span id="ph770124962613"><a name="ph770124962613"></a><a name="ph770124962613"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p770164982610"><a name="p770164982610"></a><a name="p770164982610"></a>mindspore_standalone_acjob.yaml</p>
</td>
</tr>
</tbody>
</table>

**表 3** Atlas A3 训练系列产品支持的YAML

<a name="table57051049102614"></a>
<table><thead align="left"><tr id="row107051249172610"><th class="cellrowborder" valign="top" width="8.799999999999999%" id="mcps1.2.7.1.1"><p id="p8705114972617"><a name="p8705114972617"></a><a name="p8705114972617"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.2"><p id="p1870594972615"><a name="p1870594972615"></a><a name="p1870594972615"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.65%" id="mcps1.2.7.1.3"><p id="p97063498264"><a name="p97063498264"></a><a name="p97063498264"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="39.269999999999996%" id="mcps1.2.7.1.4"><p id="p1706204911262"><a name="p1706204911262"></a><a name="p1706204911262"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="10.280000000000001%" id="mcps1.2.7.1.5"><p id="p970615497266"><a name="p970615497266"></a><a name="p970615497266"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.6"><p id="p170694910264"><a name="p170694910264"></a><a name="p170694910264"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row570610499268"><td class="cellrowborder" rowspan="3" valign="top" width="8.799999999999999%" headers="mcps1.2.7.1.1 "><p id="p1770624902616"><a name="p1770624902616"></a><a name="p1770624902616"></a>Ascend Job</p>
<p id="p167068495269"><a name="p167068495269"></a><a name="p167068495269"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15%" headers="mcps1.2.7.1.2 "><p id="p19706849182618"><a name="p19706849182618"></a><a name="p19706849182618"></a><span id="ph167064499269"><a name="ph167064499269"></a><a name="ph167064499269"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.65%" headers="mcps1.2.7.1.3 "><p id="p6706349192610"><a name="p6706349192610"></a><a name="p6706349192610"></a><span id="ph13706154962618"><a name="ph13706154962618"></a><a name="ph13706154962618"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="39.269999999999996%" headers="mcps1.2.7.1.4 "><p id="p1370624952612"><a name="p1370624952612"></a><a name="p1370624952612"></a>tensorflow_standalone_acjob_super_pod.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="10.280000000000001%" headers="mcps1.2.7.1.5 "><p id="p3707749162616"><a name="p3707749162616"></a><a name="p3707749162616"></a>示例默认为单机单卡任务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15%" headers="mcps1.2.7.1.6 "><p id="p1670794911264"><a name="p1670794911264"></a><a name="p1670794911264"></a>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
<p id="p1770716492268"><a name="p1770716492268"></a><a name="p1770716492268"></a></p>
</td>
</tr>
<tr id="row770724972619"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p0707749172618"><a name="p0707749172618"></a><a name="p0707749172618"></a><span id="ph12707184972613"><a name="ph12707184972613"></a><a name="ph12707184972613"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1370794914264"><a name="p1370794914264"></a><a name="p1370794914264"></a>pytorch_standalone_acjob_super_pod.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p167072049132613"><a name="p167072049132613"></a><a name="p167072049132613"></a>示例默认为单机16卡任务。</p>
</td>
</tr>
<tr id="row7707164912262"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p270754962617"><a name="p270754962617"></a><a name="p270754962617"></a><span id="ph1570754952617"><a name="ph1570754952617"></a><a name="ph1570754952617"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p127081949202617"><a name="p127081949202617"></a><a name="p127081949202617"></a>mindspore_standalone_acjob_super_pod.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1070817499261"><a name="p1070817499261"></a><a name="p1070817499261"></a>示例默认为双机16卡任务。</p>
</td>
</tr>
</tbody>
</table>

**通过文件配置资源信息场景<a name="section158807920347"></a>**

-   若当前环境使用的是Atlas A2 训练系列产品，选择[表4](#table62591594016)获取相应的YAML示例。

    根据[表1](#table529015783811)获取示例YAML后，Atlas 800T A2 训练服务器、Atlas 200T A2 Box16 异构子框和A200T A3 Box8 超节点服务器可基于[表2](#yaml参数说明)给出的参数说明进行修改适配。

-   若当前环境使用的是Atlas 训练系列产品，选择[表5](#table21811158146)获取相应的YAML示例。

**表 4** Atlas A2 训练系列产品支持的YAML

<a name="table62591594016"></a>
<table><thead align="left"><tr id="row72551515403"><th class="cellrowborder" valign="top" width="9.35%" id="mcps1.2.7.1.1"><p id="p72510154400"><a name="p72510154400"></a><a name="p72510154400"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="14.99%" id="mcps1.2.7.1.2"><p id="p122531515408"><a name="p122531515408"></a><a name="p122531515408"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.87%" id="mcps1.2.7.1.3"><p id="p1325131584014"><a name="p1325131584014"></a><a name="p1325131584014"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="36.51%" id="mcps1.2.7.1.4"><p id="p225815114016"><a name="p225815114016"></a><a name="p225815114016"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="12.26%" id="mcps1.2.7.1.5"><p id="p2261615184014"><a name="p2261615184014"></a><a name="p2261615184014"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.7.1.6"><p id="p32613153408"><a name="p32613153408"></a><a name="p32613153408"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row1726201544016"><td class="cellrowborder" rowspan="3" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p id="p326111516407"><a name="p326111516407"></a><a name="p326111516407"></a>Volcano Job</p>
<p id="p12475353114815"><a name="p12475353114815"></a><a name="p12475353114815"></a></p>
<p id="p18475175312481"><a name="p18475175312481"></a><a name="p18475175312481"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.99%" headers="mcps1.2.7.1.2 "><p id="p455716252506"><a name="p455716252506"></a><a name="p455716252506"></a><span id="ph1262151402"><a name="ph1262151402"></a><a name="ph1262151402"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.87%" headers="mcps1.2.7.1.3 "><p id="p1126215134018"><a name="p1126215134018"></a><a name="p1126215134018"></a><span id="ph22631519407"><a name="ph22631519407"></a><a name="ph22631519407"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.51%" headers="mcps1.2.7.1.4 "><p id="p1726151594018"><a name="p1726151594018"></a><a name="p1726151594018"></a>a800_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="12.26%" headers="mcps1.2.7.1.5 "><p id="p15261215104017"><a name="p15261215104017"></a><a name="p15261215104017"></a>示例默认为单机16卡任务。</p>
<p id="p8271715184013"><a name="p8271715184013"></a><a name="p8271715184013"></a></p>
<p id="p7271115194015"><a name="p7271115194015"></a><a name="p7271115194015"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15.02%" headers="mcps1.2.7.1.6 "><p id="p142781511408"><a name="p142781511408"></a><a name="p142781511408"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/ranktable/yaml/910b" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
<p id="p71901195499"><a name="p71901195499"></a><a name="p71901195499"></a></p>
<p id="p151911091496"><a name="p151911091496"></a><a name="p151911091496"></a></p>
</td>
</tr>
<tr id="row11271915204017"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p102791534015"><a name="p102791534015"></a><a name="p102791534015"></a><span id="ph15271015144017"><a name="ph15271015144017"></a><a name="ph15271015144017"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p14271815154017"><a name="p14271815154017"></a><a name="p14271815154017"></a>a800_pytorch_vcjob.yaml</p>
</td>
</tr>
<tr id="row14272155406"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p20273150409"><a name="p20273150409"></a><a name="p20273150409"></a><span id="ph62717152401"><a name="ph62717152401"></a><a name="ph62717152401"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p32771519408"><a name="p32771519408"></a><a name="p32771519408"></a>a800_mindspore_vcjob.yaml</p>
</td>
</tr>
<tr id="row728141517408"><td class="cellrowborder" rowspan="3" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p id="p1289158408"><a name="p1289158408"></a><a name="p1289158408"></a>Deployment</p>
<p id="p93517386498"><a name="p93517386498"></a><a name="p93517386498"></a></p>
<p id="p12352113874920"><a name="p12352113874920"></a><a name="p12352113874920"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.99%" headers="mcps1.2.7.1.2 "><p id="p1538185310530"><a name="p1538185310530"></a><a name="p1538185310530"></a><span id="ph2029215114013"><a name="ph2029215114013"></a><a name="ph2029215114013"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.87%" headers="mcps1.2.7.1.3 "><p id="p14296151409"><a name="p14296151409"></a><a name="p14296151409"></a><span id="ph1729111584013"><a name="ph1729111584013"></a><a name="ph1729111584013"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.51%" headers="mcps1.2.7.1.4 "><p id="p15291415194010"><a name="p15291415194010"></a><a name="p15291415194010"></a>a800_tensorflow_deployment.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="12.26%" headers="mcps1.2.7.1.5 "><p id="p142910157401"><a name="p142910157401"></a><a name="p142910157401"></a>示例默认为单机16卡任务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15.02%" headers="mcps1.2.7.1.6 "><p id="p7243709503"><a name="p7243709503"></a><a name="p7243709503"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/ranktable/yaml/910b" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row12914156400"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p172910152406"><a name="p172910152406"></a><a name="p172910152406"></a><span id="ph1029181516406"><a name="ph1029181516406"></a><a name="ph1029181516406"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p2029191584010"><a name="p2029191584010"></a><a name="p2029191584010"></a>a800_pytorch_deployment.yaml</p>
</td>
</tr>
<tr id="row32915158403"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p7291515164014"><a name="p7291515164014"></a><a name="p7291515164014"></a><span id="ph102941514401"><a name="ph102941514401"></a><a name="ph102941514401"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p202915155405"><a name="p202915155405"></a><a name="p202915155405"></a>a800_mindspore_deployment.yaml</p>
</td>
</tr>
</tbody>
</table>

**表 5** Atlas 训练系列产品支持的YAML

<a name="table21811158146"></a>
<table><thead align="left"><tr id="row10181111518146"><th class="cellrowborder" valign="top" width="9.35%" id="mcps1.2.7.1.1"><p id="p51941552181410"><a name="p51941552181410"></a><a name="p51941552181410"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.2"><p id="p20181111517147"><a name="p20181111517147"></a><a name="p20181111517147"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.86%" id="mcps1.2.7.1.3"><p id="p5821153911586"><a name="p5821153911586"></a><a name="p5821153911586"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="36.51%" id="mcps1.2.7.1.4"><p id="p181811156149"><a name="p181811156149"></a><a name="p181811156149"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="12.280000000000001%" id="mcps1.2.7.1.5"><p id="p86271732132719"><a name="p86271732132719"></a><a name="p86271732132719"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.6"><p id="p11672113624010"><a name="p11672113624010"></a><a name="p11672113624010"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row71811415111417"><td class="cellrowborder" rowspan="6" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p id="p191941452171418"><a name="p191941452171418"></a><a name="p191941452171418"></a>Volcano Job</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15%" headers="mcps1.2.7.1.2 "><p id="p218101516149"><a name="p218101516149"></a><a name="p218101516149"></a><span id="ph158146714142"><a name="ph158146714142"></a><a name="ph158146714142"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.86%" headers="mcps1.2.7.1.3 "><p id="zh-cn_topic_0000001609074269_p15865151810597"><a name="zh-cn_topic_0000001609074269_p15865151810597"></a><a name="zh-cn_topic_0000001609074269_p15865151810597"></a><span id="ph12195638125217"><a name="ph12195638125217"></a><a name="ph12195638125217"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.51%" headers="mcps1.2.7.1.4 "><p id="p19182161511148"><a name="p19182161511148"></a><a name="p19182161511148"></a>a800_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="12.280000000000001%" headers="mcps1.2.7.1.5 "><p id="p16627332172713"><a name="p16627332172713"></a><a name="p16627332172713"></a>示例默认为单机8卡任务。</p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="15%" headers="mcps1.2.7.1.6 "><p id="p6510121394114"><a name="p6510121394114"></a><a name="p6510121394114"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/ranktable/yaml/910" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row1598044745910"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p208651518105919"><a name="zh-cn_topic_0000001609074269_p208651518105919"></a><a name="zh-cn_topic_0000001609074269_p208651518105919"></a><span id="ph19355165113512"><a name="ph19355165113512"></a><a name="ph19355165113512"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p779714251577"><a name="p779714251577"></a><a name="p779714251577"></a>a800_pytorch_vcjob.yaml</p>
</td>
</tr>
<tr id="row66819525592"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p85061291815"><a name="zh-cn_topic_0000001609074269_p85061291815"></a><a name="zh-cn_topic_0000001609074269_p85061291815"></a><span id="ph13573184092614"><a name="ph13573184092614"></a><a name="ph13573184092614"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p2682524593"><a name="p2682524593"></a><a name="p2682524593"></a>a800_mindspore_vcjob.yaml</p>
</td>
</tr>
<tr id="row181824157147"><td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.1 "><p id="p11182141513140"><a name="p11182141513140"></a><a name="p11182141513140"></a>服务器（插<span id="ph97657495514"><a name="ph97657495514"></a><a name="ph97657495514"></a>Atlas 300T 训练卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="zh-cn_topic_0000001609074269_p764812181714"><a name="zh-cn_topic_0000001609074269_p764812181714"></a><a name="zh-cn_topic_0000001609074269_p764812181714"></a><span id="ph204675452274"><a name="ph204675452274"></a><a name="ph204675452274"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p171821215171419"><a name="p171821215171419"></a><a name="p171821215171419"></a>a300t_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.4 "><p id="p5627143215276"><a name="p5627143215276"></a><a name="p5627143215276"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
<tr id="row11788135445919"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p1864871820119"><a name="zh-cn_topic_0000001609074269_p1864871820119"></a><a name="zh-cn_topic_0000001609074269_p1864871820119"></a><span id="ph134441022151619"><a name="ph134441022151619"></a><a name="ph134441022151619"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1798025195718"><a name="p1798025195718"></a><a name="p1798025195718"></a>a300t_pytorch_vcjob.yaml</p>
</td>
</tr>
<tr id="row161351656205911"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p2648618616"><a name="zh-cn_topic_0000001609074269_p2648618616"></a><a name="zh-cn_topic_0000001609074269_p2648618616"></a><span id="ph114081559152716"><a name="ph114081559152716"></a><a name="ph114081559152716"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p9135145619598"><a name="p9135145619598"></a><a name="p9135145619598"></a>a300t_mindspore_vcjob.yaml</p>
</td>
</tr>
<tr id="row1182815141410"><td class="cellrowborder" rowspan="6" valign="top" headers="mcps1.2.7.1.1 "><p id="p1519415221416"><a name="p1519415221416"></a><a name="p1519415221416"></a>Deployment</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.2 "><p id="p151831029101812"><a name="p151831029101812"></a><a name="p151831029101812"></a><span id="ph17662124432"><a name="ph17662124432"></a><a name="ph17662124432"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="zh-cn_topic_0000001609074269_p12218446435"><a name="zh-cn_topic_0000001609074269_p12218446435"></a><a name="zh-cn_topic_0000001609074269_p12218446435"></a><span id="ph3202748132715"><a name="ph3202748132715"></a><a name="ph3202748132715"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1118218150140"><a name="p1118218150140"></a><a name="p1118218150140"></a>a800_tensorflow_deployment.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.5 "><p id="p15627143202718"><a name="p15627143202718"></a><a name="p15627143202718"></a>示例默认为单机8卡任务。</p>
</td>
</tr>
<tr id="row365114361833"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p122181468314"><a name="zh-cn_topic_0000001609074269_p122181468314"></a><a name="zh-cn_topic_0000001609074269_p122181468314"></a><span id="ph724411337162"><a name="ph724411337162"></a><a name="ph724411337162"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p96517361736"><a name="p96517361736"></a><a name="p96517361736"></a>a800_pytorch_deployment.yaml</p>
</td>
</tr>
<tr id="row23490341239"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p142187461335"><a name="zh-cn_topic_0000001609074269_p142187461335"></a><a name="zh-cn_topic_0000001609074269_p142187461335"></a><span id="ph541921262813"><a name="ph541921262813"></a><a name="ph541921262813"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1734918342033"><a name="p1734918342033"></a><a name="p1734918342033"></a>a800_mindspore_deployment.yaml</p>
</td>
</tr>
<tr id="row11821815111419"><td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.1 "><p id="p166661021172117"><a name="p166661021172117"></a><a name="p166661021172117"></a>服务器（插<span id="ph39359582495"><a name="ph39359582495"></a><a name="ph39359582495"></a>Atlas 300T 训练卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="zh-cn_topic_0000001609074269_p1160218474318"><a name="zh-cn_topic_0000001609074269_p1160218474318"></a><a name="zh-cn_topic_0000001609074269_p1160218474318"></a><span id="ph168741250162713"><a name="ph168741250162713"></a><a name="ph168741250162713"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p16182131581416"><a name="p16182131581416"></a><a name="p16182131581416"></a>a300t_tensorflow_deployment.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p462719324273"><a name="p462719324273"></a><a name="p462719324273"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
<tr id="row768144119312"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p160284718315"><a name="zh-cn_topic_0000001609074269_p160284718315"></a><a name="zh-cn_topic_0000001609074269_p160284718315"></a><span id="ph162683501617"><a name="ph162683501617"></a><a name="ph162683501617"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p768164120310"><a name="p768164120310"></a><a name="p768164120310"></a>a300t_pytorch_deployment.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p656851019573"><a name="p656851019573"></a><a name="p656851019573"></a>示例默认为单机8卡任务。</p>
</td>
</tr>
<tr id="row3166392032"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p18602047631"><a name="zh-cn_topic_0000001609074269_p18602047631"></a><a name="zh-cn_topic_0000001609074269_p18602047631"></a><span id="ph5731131512820"><a name="ph5731131512820"></a><a name="ph5731131512820"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p958514331647"><a name="p958514331647"></a><a name="p958514331647"></a>a300t_mindspore_deployment.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p185698104572"><a name="p185698104572"></a><a name="p185698104572"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
</tbody>
</table>


##### YAML参数说明<a name="ZH-CN_TOPIC_0000002511347099"></a>

本章节提供使用整卡调度或静态vNPU调度配置YAML的操作示例。在操作前，用户需要了解YAML示例的参数说明，再进行操作。

-   使用Ascend Job的用户，请参考[表1](#table159746356276)。
-   使用Volcano Job的用户，请参考[表2](#zh-cn_topic_0000001609074269_table1565872494511)。

**YAML参数说明（acjob任务）<a name="section3507205517910"></a>**

在acjob训练任务中，可使用的YAML参数说明如下表所示。

每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../appendix.md#acjob关键字段说明)。

**表 1**  YAML参数说明

<a name="table159746356276"></a>
<table><thead align="left"><tr id="row69742355274"><th class="cellrowborder" valign="top" width="27.21%" id="mcps1.2.4.1.1"><p id="p199750359270"><a name="p199750359270"></a><a name="p199750359270"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.18%" id="mcps1.2.4.1.2"><p id="p99751135152710"><a name="p99751135152710"></a><a name="p99751135152710"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.61%" id="mcps1.2.4.1.3"><p id="p39754354276"><a name="p39754354276"></a><a name="p39754354276"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row99751335102713"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p189751735172716"><a name="p189751735172716"></a><a name="p189751735172716"></a>framework</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul4975113512712"></a><a name="ul4975113512712"></a><ul id="ul4975113512712"><li>mindspore</li><li>pytorch</li><li>tensorflow</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p79751235102712"><a name="p79751235102712"></a><a name="p79751235102712"></a>框架类型，目前只支持三种。</p>
</td>
</tr>
<tr id="row2097633513272"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p6976153532710"><a name="p6976153532710"></a><a name="p6976153532710"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul16230203215710"></a><a name="ul16230203215710"></a><ul id="ul16230203215710"><li><span id="ph20976435102713"><a name="ph20976435102713"></a><a name="ph20976435102713"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="ph163483412215"><a name="ph163483412215"></a><a name="ph163483412215"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph11976935122715"><a name="ph11976935122715"></a><a name="ph11976935122715"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>Atlas 800 训练服务器，服务器（插<span id="ph2099203201811"><a name="ph2099203201811"></a><a name="ph2099203201811"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p139761835102718"><a name="p139761835102718"></a><a name="p139761835102718"></a>用于区分任务使用的芯片的类型。</p>
</td>
</tr>
<tr id="row1067294111274"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p9672541162718"><a name="p9672541162718"></a><a name="p9672541162718"></a>podgroup-sched-enable</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p6895115683718"><a name="p6895115683718"></a><a name="p6895115683718"></a>"true"</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p73261850175013"><a name="p73261850175013"></a><a name="p73261850175013"></a>仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。</p>
<a name="ul12357145235010"></a><a name="ul12357145235010"></a><ul id="ul12357145235010"><li>取值配置为字符串"true"时，表示开启批量调度功能。</li><li>取值配置为其他字符串时，表示批量调度功能不生效，使用普通调度。</li></ul>
<p id="p16526183318537"><a name="p16526183318537"></a><a name="p16526183318537"></a>若不配置该参数，表示批量调度功能不生效，使用普通调度。</p>
<div class="note" id="note1385644144618"><a name="note1385644144618"></a><a name="note1385644144618"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul1431912124720"></a><a name="ul1431912124720"></a><ul id="ul1431912124720"><li>该参数只支持使用<span id="ph1361434468"><a name="ph1361434468"></a><a name="ph1361434468"></a>Volcano</span>调度器的整卡调度特性。</li><li>仅支持在<span id="ph1731512317424"><a name="ph1731512317424"></a><a name="ph1731512317424"></a>Atlas 900 A3 SuperPoD 超节点</span>和<span id="ph1031518232426"><a name="ph1031518232426"></a><a name="ph1031518232426"></a>Atlas 800T A3 超节点服务器</span>中使用本参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="row6977735132710"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1397753512272"><a name="p1397753512272"></a><a name="p1397753512272"></a>schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p5977163532714"><a name="p5977163532714"></a><a name="p5977163532714"></a>默认值为<span class="parmvalue" id="parmvalue99772035172710"><a name="parmvalue99772035172710"></a><a name="parmvalue99772035172710"></a>“volcano”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1397873502710"><a name="p1397873502710"></a><a name="p1397873502710"></a><span id="ph8978183512271"><a name="ph8978183512271"></a><a name="ph8978183512271"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
</td>
</tr>
<tr id="row169789353270"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p0978123511271"><a name="p0978123511271"></a><a name="p0978123511271"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p2978113514276"><a name="p2978113514276"></a><a name="p2978113514276"></a>默认值为任务总副本数</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p897873582712"><a name="p897873582712"></a><a name="p897873582712"></a><span id="ph129781235182715"><a name="ph129781235182715"></a><a name="ph129781235182715"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="ph197814351275"><a name="ph197814351275"></a><a name="ph197814351275"></a>Volcano</span>时，任务运行总副本数。</p>
</td>
</tr>
<tr id="row997815356274"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p16978193511276"><a name="p16978193511276"></a><a name="p16978193511276"></a>queue</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p10978123512277"><a name="p10978123512277"></a><a name="p10978123512277"></a>默认值为<span class="parmvalue" id="parmvalue199781835112716"><a name="parmvalue199781835112716"></a><a name="parmvalue199781835112716"></a>“default”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1497812357278"><a name="p1497812357278"></a><a name="p1497812357278"></a><span id="ph0978173562717"><a name="ph0978173562717"></a><a name="ph0978173562717"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="ph797816353270"><a name="ph797816353270"></a><a name="ph797816353270"></a>Volcano</span>时，任务所属队列。</p>
</td>
</tr>
<tr id="row1397863562711"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p8979163582710"><a name="p8979163582710"></a><a name="p8979163582710"></a>（可选）successPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul897943582712"></a><a name="ul897943582712"></a><ul id="ul897943582712"><li>默认值为空，若用户不填写该参数，则默认取空值</li><li>AllWorkers</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p11979535122719"><a name="p11979535122719"></a><a name="p11979535122719"></a>表明任务成功的前提。空值代表只需要一个<span id="ph149798358273"><a name="ph149798358273"></a><a name="ph149798358273"></a>Pod</span>成功，整个任务判定为成功。取值为<span class="parmvalue" id="parmvalue1797983518271"><a name="parmvalue1797983518271"></a><a name="parmvalue1797983518271"></a>“AllWorkers”</span>表示所有<span id="ph8979163522716"><a name="ph8979163522716"></a><a name="ph8979163522716"></a>Pod</span>都成功，任务才判定为成功。</p>
</td>
</tr>
<tr id="row16979103514275"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p7979123522712"><a name="p7979123522712"></a><a name="p7979123522712"></a>container.name</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p39791535192717"><a name="p39791535192717"></a><a name="p39791535192717"></a>ascend</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p3979235182716"><a name="p3979235182716"></a><a name="p3979235182716"></a>容器的名称必须是<span class="parmvalue" id="parmvalue12979735192712"><a name="parmvalue12979735192712"></a><a name="parmvalue12979735192712"></a>“ascend”</span>。</p>
</td>
</tr>
<tr id="row20979735172712"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p4979143514276"><a name="p4979143514276"></a><a name="p4979143514276"></a>（可选）ports</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p1297916359270"><a name="p1297916359270"></a><a name="p1297916359270"></a>若用户未进行设置，系统默认填写以下参数：</p>
<a name="ul69804359278"></a><a name="ul69804359278"></a><ul id="ul69804359278"><li>name: ascendjob-port</li><li>containerPort: 2222</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p2980183592716"><a name="p2980183592716"></a><a name="p2980183592716"></a>分布式训练集合通信端口。<span class="parmname" id="parmname1198063542717"><a name="parmname1198063542717"></a><a name="parmname1198063542717"></a>“name”</span>取值只能为<span class="parmvalue" id="parmvalue17980153515270"><a name="parmvalue17980153515270"></a><a name="parmvalue17980153515270"></a>“ascendjob-port”</span>，<span class="parmname" id="parmname8980135102711"><a name="parmname8980135102711"></a><a name="parmname8980135102711"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
</td>
</tr>
<tr id="row12980535152720"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1398023582714"><a name="p1398023582714"></a><a name="p1398023582714"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul498063562713"></a><a name="ul498063562713"></a><ul id="ul498063562713"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1798016353272"><a name="p1798016353272"></a><a name="p1798016353272"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="row1198010356278"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1398113517275"><a name="p1398113517275"></a><a name="p1398113517275"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p209811535192718"><a name="p209811535192718"></a><a name="p209811535192718"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1898143513271"><a name="p1898143513271"></a><a name="p1898143513271"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr id="row1298116359278"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p898193522717"><a name="p898193522717"></a><a name="p898193522717"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p149811335192718"><a name="p149811335192718"></a><a name="p149811335192718"></a><span id="ph8981193532719"><a name="ph8981193532719"></a><a name="ph8981193532719"></a>ARM</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
<p id="p2098193513276"><a name="p2098193513276"></a><a name="p2098193513276"></a><span id="ph2098163502710"><a name="ph2098163502710"></a><a name="ph2098163502710"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p6981203592717"><a name="p6981203592717"></a><a name="p6981203592717"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="p99813357275"><a name="p99813357275"></a><a name="p99813357275"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="row711915310345"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p19119935347"><a name="p19119935347"></a><a name="p19119935347"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p131191634340"><a name="p131191634340"></a><a name="p131191634340"></a>pod：只支持Pod级重调度，不升级为Job级别。</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p21194343415"><a name="p21194343415"></a><a name="p21194343415"></a>任务重调度策略。</p>
</td>
</tr>
<tr id="row14366105123419"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p123665543412"><a name="p123665543412"></a><a name="p123665543412"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p636605193410"><a name="p636605193410"></a><a name="p636605193410"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p436675193411"><a name="p436675193411"></a><a name="p436675193411"></a>任务能够调度的最小副本数。</p>
</td>
</tr>
<tr id="row7926131215173"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1377484145110"><a name="p1377484145110"></a><a name="p1377484145110"></a><span id="ph19817116185120"><a name="ph19817116185120"></a><a name="ph19817116185120"></a>huawei.com/schedule_policy</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p1877410425111"><a name="p1877410425111"></a><a name="p1877410425111"></a><span id="ph135426519519"><a name="ph135426519519"></a><a name="ph135426519519"></a>目前支持</span><a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1671165052112"><a name="p1671165052112"></a><a name="p1671165052112"></a>配置任务需要调度的AI芯片布局形态。<span id="ph204811934163414"><a name="ph204811934163414"></a><a name="ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
<div class="note" id="note114071739202519"><a name="note114071739202519"></a><a name="note114071739202519"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1767434372512"><a name="p1767434372512"></a><a name="p1767434372512"></a>仅支持在<span id="ph1331492318423"><a name="ph1331492318423"></a><a name="ph1331492318423"></a>Atlas 训练系列产品</span>、<span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>和<span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>中使用该字段。</p>
</div></div>
</td>
</tr>
<tr id="row2098119351272"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p12982163592711"><a name="p12982163592711"></a><a name="p12982163592711"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p159821035122717"><a name="p159821035122717"></a><a name="p159821035122717"></a>指定逻辑超节点芯片数量。</p>
<a name="ul10451144414619"></a><a name="ul10451144414619"></a><ul id="ul10451144414619"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p398216353271"><a name="p398216353271"></a><a name="p398216353271"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="ph521204025916"><a name="ph521204025916"></a><a name="ph521204025916"></a>若用户未指定该字段，</span><span id="ph172121408590"><a name="ph172121408590"></a><a name="ph172121408590"></a>Volcano</span><span id="ph192121140135911"><a name="ph192121140135911"></a><a name="ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19975815131410"><a name="p19975815131410"></a><a name="p19975815131410"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note1998233513279"><a name="note1998233513279"></a><a name="note1998233513279"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul546892712569"></a><a name="ul546892712569"></a><ul id="ul546892712569"><li>仅支持在<span id="ph34244153594"><a name="ph34244153594"></a><a name="ph34244153594"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row14982435112715"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p18982635202720"><a name="p18982635202720"></a><a name="p18982635202720"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul109821735182715"></a><a name="ul109821735182715"></a><ul id="ul109821735182715"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="note7983133512715"><a name="note7983133512715"></a><a name="note7983133512715"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p16983103515274"><a name="p16983103515274"></a><a name="p16983103515274"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p2983173572716"><a name="p2983173572716"></a><a name="p2983173572716"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="note179831235202712"><a name="note179831235202712"></a><a name="note179831235202712"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul59831535122714"></a><a name="ul59831535122714"></a><ul id="ul59831535122714"><li>交换机亲和性调度1.0版本支持<span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="ph168598363399"><a name="ph168598363399"></a><a name="ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917_2"><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph4181625925"><a name="ph4181625925"></a><a name="ph4181625925"></a>PyTorch</span>和<span id="ph61882510210"><a name="ph61882510210"></a><a name="ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph311717506401"><a name="ph311717506401"></a><a name="ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_3"><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph619244413568"><a name="ph619244413568"></a><a name="ph619244413568"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>
</div></div>
</td>
</tr>
<tr id="row9984235132714"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p18984535132720"><a name="p18984535132720"></a><a name="p18984535132720"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p16729326173616"><a name="p16729326173616"></a><a name="p16729326173616"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul139845353279"></a><a name="ul139845353279"></a><ul id="ul139845353279"><li><span id="ph169841935102711"><a name="ph169841935102711"></a><a name="ph169841935102711"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="ph10984193517273"><a name="ph10984193517273"></a><a name="ph10984193517273"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="ph169855357273"><a name="ph169855357273"></a><a name="ph169855357273"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span>和<span id="ph14985135162710"><a name="ph14985135162710"></a><a name="ph14985135162710"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="ph1898523510277"><a name="ph1898523510277"></a><a name="ph1898523510277"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li><li><span id="ph19985163514279"><a name="ph19985163514279"></a><a name="ph19985163514279"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="ph4985183516277"><a name="ph4985183516277"></a><a name="ph4985183516277"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-16</li><li><span id="ph1514953013253"><a name="ph1514953013253"></a><a name="ph1514953013253"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph8619174411286"><a name="ph8619174411286"></a><a name="ph8619174411286"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签。</li><li><span id="ph261924414289"><a name="ph261924414289"></a><a name="ph261924414289"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p5986235142714"><a name="p5986235142714"></a><a name="p5986235142714"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
<div class="note" id="note898773512719"><a name="note898773512719"></a><a name="note898773512719"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1027616512420"><a name="p1027616512420"></a><a name="p1027616512420"></a><span id="ph9014016509"><a name="ph9014016509"></a><a name="ph9014016509"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001519959665_b168254314713"><a name="zh-cn_topic_0000001519959665_b168254314713"></a><a name="zh-cn_topic_0000001519959665_b168254314713"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row1598720359275"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p2987143519279"><a name="p2987143519279"></a><a name="p2987143519279"></a>requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p1198733572715"><a name="p1198733572715"></a><a name="p1198733572715"></a><strong id="b2098733513275"><a name="b2098733513275"></a><a name="b2098733513275"></a>整卡调度：</strong></p>
<p id="p998783518273"><a name="p998783518273"></a><a name="p998783518273"></a>huawei.com/Ascend910: <em id="i6988935192715"><a name="i6988935192715"></a><a name="i6988935192715"></a>x</em></p>
<p id="p39887359272"><a name="p39887359272"></a><a name="p39887359272"></a>根据所使用芯片类型不同，x取值如下：</p>
<a name="ul1798823522713"></a><a name="ul1798823522713"></a><ul id="ul1798823522713"><li><span id="ph1598853514277"><a name="ph1598853514277"></a><a name="ph1598853514277"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="ul1998810354278"></a><a name="ul1998810354278"></a><ul id="ul1998810354278"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul>
</li><li><span id="ph2988153582716"><a name="ph2988153582716"></a><a name="ph2988153582716"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="ul159887358276"></a><a name="ul159887358276"></a><ul id="ul159887358276"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul>
</li><li>服务器（插<span id="ph15989835132717"><a name="ph15989835132717"></a><a name="ph15989835132717"></a>Atlas 300T 训练卡</span>）：<a name="ul18989235102719"></a><a name="ul18989235102719"></a><ul id="ul18989235102719"><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul>
</li><li><span id="ph52261657163417"><a name="ph52261657163417"></a><a name="ph52261657163417"></a>Atlas 800T A2 训练服务器</span>和<span id="ph999063517276"><a name="ph999063517276"></a><a name="ph999063517276"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="ul199020351273"></a><a name="ul199020351273"></a><ul id="ul199020351273"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="ph18991173510273"><a name="ph18991173510273"></a><a name="ph18991173510273"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="ul69913354274"></a><a name="ul69913354274"></a><ul id="ul69913354274"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li><li><span id="ph18991153532717"><a name="ph18991153532717"></a><a name="ph18991153532717"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph20855174373110"><a name="ph20855174373110"></a><a name="ph20855174373110"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph27041232195"><a name="ph27041232195"></a><a name="ph27041232195"></a>Atlas 800T A3 超节点服务器</span>：<a name="ul499153515277"></a><a name="ul499153515277"></a><ul id="ul499153515277"><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对<span id="ph798511112819"><a name="ph798511112819"></a><a name="ph798511112819"></a>Atlas 900 A3 SuperPoD 超节点</span>的逻辑超节点亲和任务：16</li></ul>
</li></ul>
<p id="p4993163513278"><a name="p4993163513278"></a><a name="p4993163513278"></a><strong id="b099353510277"><a name="b099353510277"></a><a name="b099353510277"></a>静态vNPU调度：</strong></p>
<p id="p39931635142716"><a name="p39931635142716"></a><a name="p39931635142716"></a>huawei.com/Ascend910-<strong id="b12993153514272"><a name="b12993153514272"></a><a name="b12993153514272"></a><em id="i16993153522711"><a name="i16993153522711"></a><a name="i16993153522711"></a>Y</em></strong>: 1</p>
<p id="p1399323512713"><a name="p1399323512713"></a><a name="p1399323512713"></a>取值为1。只能使用一个NPU下的vNPU。</p>
<p id="p119931335142715"><a name="p119931335142715"></a><a name="p119931335142715"></a>如huawei.com/Ascend910-<em id="i20993935152714"><a name="i20993935152714"></a><a name="i20993935152714"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p4994193520276"><a name="p4994193520276"></a><a name="p4994193520276"></a>至少请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<div class="note" id="note6994535162719"><a name="note6994535162719"></a><a name="note6994535162719"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p169941635192717"><a name="p169941635192717"></a><a name="p169941635192717"></a><strong id="b179941535122710"><a name="b179941535122710"></a><a name="b179941535122710"></a><em id="i1999493582720"><a name="i1999493582720"></a><a name="i1999493582720"></a>Y</em></strong>取值可参考<a href="./virtual_instance.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与虚拟设备类型关系表的“vNPU类型”列。</p>
<p id="p17994143511279"><a name="p17994143511279"></a><a name="p17994143511279"></a>以vNPU类型Ascend910-6c.1cpu.16g为例，<strong id="b29941735102716"><a name="b29941735102716"></a><a name="b29941735102716"></a><em id="i11995143582719"><a name="i11995143582719"></a><a name="i11995143582719"></a>Y</em></strong>取值为6c.1cpu.16g，不包括前面的Ascend910。</p>
<p id="p699519351275"><a name="p699519351275"></a><a name="p699519351275"></a>虚拟化模板的更多信息可以参考<a href="./virtual_instance.md#虚拟化规则">虚拟化规则</a>章节中的“虚拟化模板”。</p>
</div></div>
</td>
</tr>
<tr id="row1899517352272"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p999523582714"><a name="p999523582714"></a><a name="p999523582714"></a>limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p203039915181"><a name="p203039915181"></a><a name="p203039915181"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<p id="p4739213121713"><a name="p4739213121713"></a><a name="p4739213121713"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
</td>
</tr>
<tr id="row119951035132715"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p6995103513275"><a name="p6995103513275"></a><a name="p6995103513275"></a>metadata.annotations['huawei.com/Ascend<em id="i1599519355274"><a name="i1599519355274"></a><a name="i1599519355274"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p12995163592715"><a name="p12995163592715"></a><a name="p12995163592715"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p17995235152710"><a name="p17995235152710"></a><a name="p17995235152710"></a><span id="ph13995133582718"><a name="ph13995133582718"></a><a name="ph13995133582718"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<div class="note" id="note6995123592720"><a name="note6995123592720"></a><a name="note6995123592720"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1299613592712"><a name="p1299613592712"></a><a name="p1299613592712"></a>该参数只支持使用<span id="ph699616354271"><a name="ph699616354271"></a><a name="ph699616354271"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
</div></div>
</td>
</tr>
<tr id="row158871116111016"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1062816462407"><a name="p1062816462407"></a><a name="p1062816462407"></a>hostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul960434424111"></a><a name="ul960434424111"></a><ul id="ul960434424111"><li>true：使用HostIP创建Pod。<p id="p3565137193114"><a name="p3565137193114"></a><a name="p3565137193114"></a>此种情况下，需要在YAML中同步配置环境变量HCCL_IF_IP为status.hostIP。</p>
</li><li>false：不使用HostIP创建Pod。<p id="p15499134563115"><a name="p15499134563115"></a><a name="p15499134563115"></a>未传入此参数或此参数的值为false时，不需要配置上述环境变量。</p>
<p id="p8537842173118"><a name="p8537842173118"></a><a name="p8537842173118"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><a name="ul5695536155919"></a><a name="ul5695536155919"></a><ul id="ul5695536155919"><li>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</li><li>不传入此参数时，默认不使用HostIP创建Pod。<div class="note" id="note1962023375819"><a name="note1962023375819"></a><a name="note1962023375819"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p21989143551"><a name="p21989143551"></a><a name="p21989143551"></a>当采用HostIp方式创建Pod，依然存在创建Pod速度慢且Pod之间通信速度慢的问题。此时推荐采用挂载RankTable文件的方式，通过解析RankTable文件获得Pod的hostIP，并将其注入到对应框架任务的环境变量中（如ms框架注入到环境变量MS_SCHED_HOST中），实现建链。</p>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row1175517234373"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p2755623183718"><a name="p2755623183718"></a><a name="p2755623183718"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p20956101093820"><a name="p20956101093820"></a><a name="p20956101093820"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul959814611371"></a><a name="ul959814611371"></a><ul id="ul959814611371"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><div class="note" id="note14495132616360"><a name="note14495132616360"></a><a name="note14495132616360"></a><div class="notebody"><p id="p20494202617363"><a name="p20494202617363"></a><a name="p20494202617363"></a>仅支持在<span id="ph7230184917387"><a name="ph7230184917387"></a><a name="ph7230184917387"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

**YAML参数说明（deploy任务或vcjob任务）<a name="section645410711512"></a>**

在deploy任务或vcjob训练任务中，可使用的YAML参数说明如下表所示。

**表 2**  YAML参数说明

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="22.58%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="40.86%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001609074269_p19658152414451"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001609074269_ul1531417539259"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001609074269_p11302326164814"></a>N为节点个数，Deployment类型的任务不需要该参数，该参数建议与replicas保持一致。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a>ARM</span>环境：<span id="ph4569134274515"><a name="ph4569134274515"></a><a name="ph4569134274515"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：<span id="ph7394135434515"><a name="ph7394135434515"></a><a name="ph7394135434515"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="row319913141385"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p17879179384"><a name="p17879179384"></a><a name="p17879179384"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p11787717143811"><a name="p11787717143811"></a><a name="p11787717143811"></a>pod：只支持Pod级重调度，不升级为Job级别。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1278741713381"><a name="p1278741713381"></a><a name="p1278741713381"></a>任务重调度策略。</p>
</td>
</tr>
<tr id="row675991618389"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p778791715380"><a name="p778791715380"></a><a name="p778791715380"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1378781718388"><a name="p1378781718388"></a><a name="p1378781718388"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1378741712380"><a name="p1378741712380"></a><a name="p1378741712380"></a>任务能够调度的最小副本数。</p>
</td>
</tr>
<tr id="row492051125013"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
<div class="note" id="note1230363125010"><a name="note1230363125010"></a><a name="note1230363125010"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002511347099_p1767434372512"><a name="zh-cn_topic_0000002511347099_p1767434372512"></a><a name="zh-cn_topic_0000002511347099_p1767434372512"></a>仅支持在<span id="zh-cn_topic_0000002511347099_ph1331492318423"><a name="zh-cn_topic_0000002511347099_ph1331492318423"></a><a name="zh-cn_topic_0000002511347099_ph1331492318423"></a>Atlas 训练系列产品</span>、<span id="zh-cn_topic_0000002511347099_ph2314323124211"><a name="zh-cn_topic_0000002511347099_ph2314323124211"></a><a name="zh-cn_topic_0000002511347099_ph2314323124211"></a><term id="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>和<span id="zh-cn_topic_0000002511347099_ph531432344210"><a name="zh-cn_topic_0000002511347099_ph531432344210"></a><a name="zh-cn_topic_0000002511347099_ph531432344210"></a><term id="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000002511347099_zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>中使用该字段。</p>
</div></div>
</td>
</tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>指定逻辑超节点芯片数量。</p>
<a name="ul1150756144219"></a><a name="ul1150756144219"></a><ul id="ul1150756144219"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p1250719624216"><a name="p1250719624216"></a><a name="p1250719624216"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在<span id="zh-cn_topic_0000002511347099_ph34244153594"><a name="zh-cn_topic_0000002511347099_ph34244153594"></a><a name="zh-cn_topic_0000002511347099_ph34244153594"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row862818313577"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p132726845716"><a name="p132726845716"></a><a name="p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="ul1427218195710"></a><a name="ul1427218195710"></a><ul id="ul1427218195710"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="note32586245294"><a name="note32586245294"></a><a name="note32586245294"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p5258102462916"><a name="p5258102462916"></a><a name="p5258102462916"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p32732087577"><a name="p32732087577"></a><a name="p32732087577"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="note13620817512"><a name="note13620817512"></a><a name="note13620817512"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul961424647"></a><a name="ul961424647"></a><ul id="ul961424647"><li>交换机亲和性调度1.0版本支持<span id="ph63831524184110"><a name="ph63831524184110"></a><a name="ph63831524184110"></a>Atlas 训练系列产品</span>和<span id="ph138318245414"><a name="ph138318245414"></a><a name="ph138318245414"></a><term id="zh-cn_topic_0000001519959665_term57208119917_4"><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph17383182419412"><a name="ph17383182419412"></a><a name="ph17383182419412"></a>PyTorch</span>和<span id="ph1383224134120"><a name="ph1383224134120"></a><a name="ph1383224134120"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph438320243412"><a name="ph438320243412"></a><a name="ph438320243412"></a><term id="zh-cn_topic_0000001519959665_term57208119917_5"><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph134821711841"><a name="ph134821711841"></a><a name="ph134821711841"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001609074269_p1449413229314"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p7665323173618"><a name="p7665323173618"></a><a name="p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul14200073713"></a><a name="ul14200073713"></a><ul id="ul14200073713"><li><span id="zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000001609074269_ph1284164912438"><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph4528511506"><a name="zh-cn_topic_0000001609074269_ph4528511506"></a><a name="zh-cn_topic_0000001609074269_ph4528511506"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph486033685311"><a name="ph486033685311"></a><a name="ph486033685311"></a>Atlas 800T A2 训练服务器</span>和<span id="ph1296712308221"><a name="ph1296712308221"></a><a name="ph1296712308221"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="ph4487202241512"><a name="ph4487202241512"></a><a name="ph4487202241512"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-8</li><li><span id="ph1114211211203"><a name="ph1114211211203"></a><a name="ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="ph5811017182112"><a name="ph5811017182112"></a><a name="ph5811017182112"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_4"><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a>{xxx}</em></span>b-16</li><li><span id="ph115277505269"><a name="ph115277505269"></a><a name="ph115277505269"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph7730165573912"><a name="ph7730165573912"></a><a name="ph7730165573912"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签。</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1954213851616"><a name="p1954213851616"></a><a name="p1954213851616"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
<div class="note" id="note19666163011214"><a name="note19666163011214"></a><a name="note19666163011214"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1105153313533"><a name="p1105153313533"></a><a name="p1105153313533"></a><span id="ph710573305319"><a name="ph710573305319"></a><a name="ph710573305319"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001519959665_b168254314713_1"><a name="zh-cn_topic_0000001519959665_b168254314713_1"></a><a name="zh-cn_topic_0000001519959665_b168254314713_1"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209_1"><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1996615912482"><a name="p1996615912482"></a><a name="p1996615912482"></a><strong id="b118963916494"><a name="b118963916494"></a><a name="b118963916494"></a>整卡调度：</strong></p>
<p id="p44675644812"><a name="p44675644812"></a><a name="p44675644812"></a><span id="ph1567213131613"><a name="ph1567213131613"></a><a name="ph1567213131613"></a>huawei.com/Ascend910</span>: <em id="i2478131910511"><a name="i2478131910511"></a><a name="i2478131910511"></a>x</em></p>
<p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>根据所使用芯片类型不同，x取值如下：</p>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a><ul id="ul4403181216571"><li><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</li><li><span id="zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001609074269_ul1713712328597"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph1223449506"><a name="zh-cn_topic_0000001609074269_ph1223449506"></a><a name="zh-cn_topic_0000001609074269_ph1223449506"></a>Atlas 300T 训练卡</span>）：<a name="ul3519194217372"></a><a name="ul3519194217372"></a><ul id="ul3519194217372"><li>单机单芯片：1</li><li>单机多芯片：2</li><li>分布式：2</li></ul>
</li><li><span id="ph1176216314557"><a name="ph1176216314557"></a><a name="ph1176216314557"></a>Atlas 800T A2 训练服务器</span>和<span id="ph107421743105017"><a name="ph107421743105017"></a><a name="ph107421743105017"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="ul169264817234"></a><a name="ul169264817234"></a><ul id="ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="ph129391532155719"><a name="ph129391532155719"></a><a name="ph129391532155719"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="ul555885820439"></a><a name="ul555885820439"></a><ul id="ul555885820439"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li><li><span id="ph133001904447"><a name="ph133001904447"></a><a name="ph133001904447"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph830011074420"><a name="ph830011074420"></a><a name="ph830011074420"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 超节点服务器</span>：<a name="ul130020074415"></a><a name="ul130020074415"></a><ul id="ul130020074415"><li>单机多芯片：2、4、6、8、10、12、14、16</li><li>分布式：16</li></ul>
</li></ul>
<p id="p1498123034911"><a name="p1498123034911"></a><a name="p1498123034911"></a><strong id="b7488133134911"><a name="b7488133134911"></a><a name="b7488133134911"></a>静态vNPU调度：</strong></p>
<p id="p19104113195111"><a name="p19104113195111"></a><a name="p19104113195111"></a>huawei.com/Ascend910-<strong id="b14105734512"><a name="b14105734512"></a><a name="b14105734512"></a><em id="i17105533512"><a name="i17105533512"></a><a name="i17105533512"></a>Y</em></strong>: 1</p>
<p id="p1851116142917"><a name="p1851116142917"></a><a name="p1851116142917"></a>取值为1。只能使用一个NPU下的vNPU。</p>
<p id="p11413153312435"><a name="p11413153312435"></a><a name="p11413153312435"></a>如huawei.com/Ascend910-<em id="i94134332434"><a name="i94134332434"></a><a name="i94134332434"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5498134535310"><a name="p5498134535310"></a><a name="p5498134535310"></a>至少请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<div class="note" id="note1648201912419"><a name="note1648201912419"></a><a name="note1648201912419"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p124851954116"><a name="p124851954116"></a><a name="p124851954116"></a><strong id="b1327175513443"><a name="b1327175513443"></a><a name="b1327175513443"></a><em id="i827655184412"><a name="i827655184412"></a><a name="i827655184412"></a>Y</em></strong>取值可参考<a href="./virtual_instance.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与虚拟设备类型关系表的“vNPU类型”列。</p>
<p id="p128388914430"><a name="p128388914430"></a><a name="p128388914430"></a>以vNPU类型Ascend910-6c.1cpu.16g为例，<strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong>取值为6c.1cpu.16g，不包括前面的Ascend910。</p>
<p id="p2491818192318"><a name="p2491818192318"></a><a name="p2491818192318"></a>虚拟化模板的更多信息可以参考<a href="./virtual_instance.md#虚拟化规则">虚拟化规则</a>章节中的“虚拟化模板”。</p>
</div></div>
</td>
</tr>
<tr id="row25918533287"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p05117110298"><a name="p05117110298"></a><a name="p05117110298"></a>limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p13683185074711"><a name="p13683185074711"></a><a name="p13683185074711"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<p id="p16683135019479"><a name="p16683135019479"></a><a name="p16683135019479"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
</td>
</tr>
<tr id="row14747131720228"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/Ascend<em id="i103895254475"><a name="i103895254475"></a><a name="i103895254475"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>该参数只支持使用<span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span>调度器的整卡调度特性。使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
</div></div>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1941725316543"><a name="p1941725316543"></a><a name="p1941725316543"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul2750122165318"></a><a name="ul2750122165318"></a><ul id="ul2750122165318"><li>Atlas 800 训练服务器，服务器（插<span id="ph6581133055411"><a name="ph6581133055411"></a><a name="ph6581133055411"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li><span id="ph10656173717129"><a name="ph10656173717129"></a><a name="ph10656173717129"></a><term id="zh-cn_topic_0000001519959665_term57208119917_6"><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a>Atlas A2 训练系列产品</term></span>、<span id="ph1665620377128"><a name="ph1665620377128"></a><a name="ph1665620377128"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph14656337131215"><a name="ph14656337131215"></a><a name="ph14656337131215"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph12656113717123"><a name="ph12656113717123"></a><a name="ph12656113717123"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph1265633714121"><a name="ph1265633714121"></a><a name="ph1265633714121"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>用于区分任务使用的芯片的类型。需要在<span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
<div class="note" id="note14282027593"><a name="note14282027593"></a><a name="note14282027593"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1328162720912"><a name="p1328162720912"></a><a name="p1328162720912"></a><span id="ph19729197"><a name="ph19729197"></a><a name="ph19729197"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001519959665_b168254314713_2"><a name="zh-cn_topic_0000001519959665_b168254314713_2"></a><a name="zh-cn_topic_0000001519959665_b168254314713_2"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，此处的{<em id="zh-cn_topic_0000001519959665_i1914312018209_2"><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row141124616406"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><div class="note" id="note1031313718402"><a name="note1031313718402"></a><a name="note1031313718402"></a><div class="notebody"><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>仅支持在<span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

**表 3**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>
<table><thead align="left"><tr id="row192066612155"><th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.3.1.1"><p id="p132062614153"><a name="p132062614153"></a><a name="p132062614153"></a>配置</p>
</th>
<th class="cellrowborder" valign="top" width="77.7%" id="mcps1.2.3.1.2"><p id="p5206126181520"><a name="p5206126181520"></a><a name="p5206126181520"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row201261346162"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p457945418181"><a name="p457945418181"></a><a name="p457945418181"></a>chip4-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p7579105411817"><a name="p7579105411817"></a><a name="p7579105411817"></a>1个节点8张芯片，每4个芯片形成1个互联环。例如，<span id="ph18314192319429"><a name="ph18314192319429"></a><a name="ph18314192319429"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="ph631452384213"><a name="ph631452384213"></a><a name="ph631452384213"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的整模块场景 /Atlas 350 推理卡内部共8张卡，每4张卡通过UB扣板连接。</p>
</td>
</tr>
<tr id="row102574171610"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p205801254151810"><a name="p205801254151810"></a><a name="p205801254151810"></a>chip1-node2</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p65801354101816"><a name="p65801354101816"></a><a name="p65801354101816"></a>1个节点2张芯片。例如，<span id="ph97657495514"><a name="ph97657495514"></a><a name="ph97657495514"></a>Atlas 300T 训练卡</span>的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。</p>
</td>
</tr>
<tr id="row825811151619"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p17580854201815"><a name="p17580854201815"></a><a name="p17580854201815"></a>chip4-node4</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p858019546184"><a name="p858019546184"></a><a name="p858019546184"></a>1个节点4张芯片，形成1个互联环。例如，<span id="ph1165491719811"><a name="ph1165491719811"></a><a name="ph1165491719811"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="ph15654111712815"><a name="ph15654111712815"></a><a name="ph15654111712815"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的半配场景。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip8-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点8张卡，8张卡都在1个互联环上。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span>。</p>
</td>
</tr>
<tr id="row1820613612158"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p1358111544185"><a name="p1358111544185"></a><a name="p1358111544185"></a>chip8-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p9581135461815"><a name="p9581135461815"></a><a name="p9581135461815"></a>1个节点16张卡，每8张卡在1个互联环上。例如，<span id="ph1831422311424"><a name="ph1831422311424"></a><a name="ph1831422311424"></a>Atlas 200T A2 Box16 异构子框</span>。</p>
</td>
</tr>
<tr id="row2020613616154"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2581854121811"><a name="p2581854121811"></a><a name="p2581854121811"></a>chip2-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p758125481813"><a name="p758125481813"></a><a name="p758125481813"></a>1个节点16张卡，每2张卡在1个互联环上。例如，<span id="ph855133261011"><a name="ph855133261011"></a><a name="ph855133261011"></a>Atlas 800T A3 超节点服务器</span>。</p>
</td>
</tr>
<tr id="row22064621511"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p558111549188"><a name="p558111549188"></a><a name="p558111549188"></a>chip2-node16-sp</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p258115548187"><a name="p258115548187"></a><a name="p258115548187"></a>1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，<span id="ph1990844161011"><a name="ph1990844161011"></a><a name="ph1990844161011"></a>Atlas 900 A3 SuperPoD 超节点</span>。</p>
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


##### 配置YAML<a name="ZH-CN_TOPIC_0000002511347101"></a>

本章节指导用户配置整卡调度或静态vNPU调度特性的任务YAML，通过环境变量配置资源信息的用户请参考[通过环境变量配置资源信息场景](#section598118132817)；通过文件配置资源信息的用户请参考[通过文件配置资源信息场景](#section6131855154814)。

**通过环境变量配置资源信息场景<a name="section598118132817"></a>**

>[!NOTE] 说明 
>此场景下，用户需已创建[hccl.json](../appendix.md#hccljson文件说明)文件的具体挂载路径才能执行以下操作，详细操作步骤请参见[步骤4](../installation_guide.md#ascend-operator)。

1.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    **表 1**  操作参考

    <a name="table9830101615287"></a>
    <table><thead align="left"><tr id="row1183115167289"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p1883131617285"><a name="p1883131617285"></a><a name="p1883131617285"></a>特性名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p118311416122815"><a name="p118311416122815"></a><a name="p118311416122815"></a>操作示例</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1383111642810"><td class="cellrowborder" rowspan="2" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p183111662814"><a name="p183111662814"></a><a name="p183111662814"></a>整卡调度</p>
    <p id="p6831131682816"><a name="p6831131682816"></a><a name="p6831131682816"></a></p>
    <p id="p4831141617286"><a name="p4831141617286"></a><a name="p4831141617286"></a></p>
    <p id="p1783141615287"><a name="p1783141615287"></a><a name="p1783141615287"></a></p>
    <p id="p53986186431"><a name="p53986186431"></a><a name="p53986186431"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1083151622813"><a name="p1083151622813"></a><a name="p1083151622813"></a><a href="#li28347161281">在Atlas 800 训练服务器上创建单机任务</a></p>
    <div class="note" id="note19832191615281"><a name="note19832191615281"></a><a name="note19832191615281"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p128321216132812"><a name="p128321216132812"></a><a name="p128321216132812"></a>若需要使用<span id="ph1183219160285"><a name="ph1183219160285"></a><a name="ph1183219160285"></a>PyTorch</span>或<span id="ph8832121614281"><a name="ph8832121614281"></a><a name="ph8832121614281"></a>MindSpore</span>框架支持的交换机亲和性调度，配置示例请参见<a href="#li583911163280">配置交换机亲和性调度参考示例</a>。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row0832171652816"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p178320163285"><a name="p178320163285"></a><a name="p178320163285"></a><a href="#li1731218243100">在Atlas 800T A2 训练服务器上创建分布式任务</a></p>
    </td>
    </tr>
    <tr id="row108334168282"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p2833141612811"><a name="p2833141612811"></a><a name="p2833141612811"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p18339161282"><a name="p18339161282"></a><a name="p18339161282"></a><a href="#li1086213163289">Atlas 900 A3 SuperPoD 超节点上创建单机训练任务</a></p>
    </td>
    </tr>
    <tr id="row524620253494"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1324682574913"><a name="p1324682574913"></a><a name="p1324682574913"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1924622524915"><a name="p1924622524915"></a><a name="p1924622524915"></a><a href="#li164321720423">在Atlas&nbsp;800T&nbsp;A2&nbsp;训练服务器上创建训练任务（Scheduler挂载芯片的方式）</a></p>
    </td>
    </tr>
    <tr id="row10833191620281"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1983421672814"><a name="p1983421672814"></a><a name="p1983421672814"></a>静态vNPU调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p16834616182815"><a name="p16834616182815"></a><a name="p16834616182815"></a><a href="#li1987314168284">在Atlas 800 训练服务器上创建单机任务</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    -   <a name="li28347161281"></a>使用**整卡调度**特性，参考本配置。以tensorflow\_standalone\_acjob.yaml为例，在Atlas 800 训练服务器节点创建**单机训练**任务，执行1\*8芯片训练任务，修改示例如下。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-tensorflow
          labels:
            framework: tensorflow  # 训练框架
        spec:
          schedulerName: volcano        #当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:           # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 1        # 任务总副本数
              queue: default         # 任务所属队列
          successPolicy: AllWorkers    # 任务成功的前提
          replicaSpecs:
            Chief:
              replicas: 1      # 任务副本数
              restartPolicy: Never
              template:
                spec:
                  nodeSelector:
                    host-arch: huawei-arm               # 可选值，根据实际情况填写
                    accelerator-type: module           #节点类型
                  containers:
                  - name: ascend                        # 必须为ascend，不能修改
                    image: tensorflow-test:latest        # 镜像名称
        ...
                  env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                   ports:                           # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 8 # 申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8 #与limits取值一致
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

    -   <a name="li583911163280"></a>使用**整卡调度**特性，参考本配置。PyTorch和MindSpore框架新增了使用交换机亲和性调度的功能，该功能支持大模型任务和普通任务。以pytorch\_standalone\_acjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用1个芯片，修改示例如下。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-pytorch
          labels:
            framework: pytorch   # 镜像名称
            tor-affinity: "normal-schema" # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
        spec:
          schedulerName: volcano  # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 1    # 任务总副本数
              queue: default    # 任务所属队列
          successPolicy: AllWorkers   # 任务成功的前提
          replicaSpecs:
            Master:
              replicas: 
              restartPolicy: Never
              template:
                spec:
                  nodeSelector:
                    host-arch: huawei-arm               # 可选值，根据实际情况填写
                    accelerator-type: module         # 节点类型
                  containers:
                  - name: ascend                       # 必须为ascend，不能修改
                  image: PyTorch-test:latest       # 镜像名称
        ...
                  env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                    ports:                          #分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 1    # 任务申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 1   # 与limits取值一致
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

        >[!NOTE] 说明 
        >TensorFlow、PyTorch、MindSpore框架中对应的Chief、Master、Scheduler的“replicas”字段不能超过1。单机任务时，TensorFlow、PyTorch框架不需要Worker。单卡任务时，MindSpore框架不需要Scheduler。

    -   使用**整卡调度**特性，参考本配置。tensorflow\_multinodes\_acjob\__\{xxx\}_b.yaml为例，在两台Atlas 800T A2 训练服务器节点创建**分布式训练**任务，执行2\*8芯片训练任务，修改示例如下，分布式任务的每个Pod只能调度到不同节点。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-tensorflow        # 任务名
          labels:
            framework: tensorflow     # 训练框架名称
            ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
        spec:
          schedulerName: volcano    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:       # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 2   #任务总副本数
              queue: default     # 任务所属队列
          successPolicy: AllWorkers  #任务成功的前提
          replicaSpecs:
            Chief:
              replicas: 1   # 任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  affinity:                                         # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: job-name
                                operator: In
                                values:
                                  - default-test-tensorflow         # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                  nodeSelector:
                    host-arch: huawei-arm               # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8   # 节点类型
                  containers:
                  - name: ascend                                     # 必须为ascend，不能修改
                  image: tensorflow-test:latest  #镜像名称
        ...
                    resources:
                      limits:
                        huawei.com/Ascend910: 8     #申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8     # 与limits取值一致
                    volumeMounts:
        ...
                  volumes:
        ...
            Worker:
              replicas: 1   #任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b   # 标识产品类型
                spec:
                  affinity:            # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: job-name
                                operator: In
                                values:
                                  - default-test-tensorflow        # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                  nodeSelector:
                    host-arch: huawei-arm               # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8  # 节点类型
                  containers:
                  - name: ascend                                   # 必须为ascend，不能修改
        ...
                  env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                    ports:                          # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 8   # 任务申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8   # 与limits取值一致
                    volumeMounts:
        ...
                  volumes:
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

    -   <a name="li1086213163289"></a>使用**整卡调度**特性，参考本配置。以pytorch\_standalone\_acjob\_super\_pod.yaml为例，在一台Atlas 900 A3 SuperPoD 超节点上创建**单机训练**任务，修改示例如下。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-pytorch
          labels:
            framework: pytorch    # 框架类型
            ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
            podgroup-sched-enable: "true"  # 仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。取值为字符串"true"时，表示开启批量调度功能；取值为其他字符串时，表示批量调度功能不生效，使用普通调度。若不配置该参数，表示批量调度功能不生效，使用普通调度
          annotations:
            sp-block: "16"  # 需要和申请的芯片数量一致
        spec:
          schedulerName: volcano  # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 1     # 任务总副本数
              queue: default  # 任务所属队列
          successPolicy: AllWorkers     # 任务成功的前提
          replicaSpecs:
            Master:
              replicas: 1   # 任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b
                spec:
                  nodeSelector:
                    host-arch: huawei-arm      # 可选值，根据实际情况填写
                    accelerator-type: module-a3-16-super-pod    # 节点类型
                  containers:
                  - name: ascend  # 必须为ascend，不能修改
                    image: pytorch-test:latest      # 训练基础镜像
                    imagePullPolicy: IfNotPresent
                    env:
        ...
                      - name: ASCEND_VISIBLE_DEVICES     # Ascend Docker Runtime会使用该字段
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.annotations['huawei.com/Ascend910']               
        ...
                    ports:                     # 分布式训练集合通信端口   
                      - containerPort: 2222         # determined by user
                        name: ascendjob-port        # do not modify
                    resources:
                      limits:
                        huawei.com/Ascend910: 16   # 任务申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 16   # 与limits取值一致
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

    -   <a name="li1987314168284"></a>使用**静态vNPU调度**特性，参考本配置。以tensorflow\_standalone\_acjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，申请2个AI Core的任务为例，修改示例如下。静态vNPU调度只支持单机训练任务。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-tensorflow
          labels:
            framework: tensorflow  # 训练框架
            ring-controller.atlas: ascend-910   # 标识产品类型
        spec:
          schedulerName: volcano        # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:           # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 1   # 任务总副本数
              queue: default   # 任务所属队列
          successPolicy: AllWorkers  # 任务成功的前提
          replicaSpecs:
            Chief:
              replicas: 1 # 任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-910   # 标识产品类型
                spec:
                  nodeSelector:
                    host-arch: huawei-arm               # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8  # 节点类型
                  containers:
                  - name: ascend                          # 必须为ascend，不能修改
                  image: tensorflow-test:latest       # 镜像名称
        ...
                  env:
        ...
                 # 静态vNPU调度暂不支持ASCEND_VISIBLE_DEVICES相关字段，需要删除以下加粗字段
                  - name: ASCEND_VISIBLE_DEVICES                       
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               
        ...
                    ports:                 # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910-2c: 1# vNPU调度此处数量只能为1
                      requests:
                        huawei.com/Ascend910-2c: 1# vNPU调度此处数量只能为1
                    volumeMounts:
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

    -   <a name="li164321720423"></a>使用整卡调度特性，参考本配置。以mindspore\_multinodes\_acjob\_\{xxx\}b.yaml为例，在一台Atlas 800T A2 训练服务器上以Scheduler挂载芯片的方式执行2\*8卡训练任务，修改示例如下。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-mindspore
          labels:
            framework: mindspore     # 训练框架名称
            ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
        spec:
          schedulerName: volcano    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:      # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 2  #任务总副本数
              queue: default     # 任务所属队列
          successPolicy: AllWorkers  #任务成功的前提
          replicaSpecs:
            Scheduler:
              replicas: 1   # 任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  hostNetwork: true    # 可选值，根据实际情况填写，true支持hostIP创建Pod，false不支持hostIP创建Pod
                  affinity:                                         # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: job-name
                                operator: In
                                values:
                                  - default-test-mindspore         # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                  nodeSelector:
                    host-arch: huawei-arm              # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8   # 节点类型
                  containers:
                  - name: ascend                                     # 必须为ascend，不能修改
                    image: mindspore-test:latest  #镜像名称
                    imagePullPolicy: IfNotPresent
        ...
                    env:                                    
                      - name: HCCL_IF_IP                    # 可选值，根据实际情况填写
                        valueFrom:                          # 若hostNetwork配置为true，需要同步配置HCCL_IF_IP环境变量
                          fieldRef:                         # 若hostNetwork未配置或配置为false，不可配置HCCL_IF_IP环境变量
                            fieldPath: status.hostIP        # 
        ...            
                    ports:                          # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 8 # 申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8 #与limits取值一致
                    volumeMounts:
        ...            
                  volumes:
        ...            
            Worker:
              replicas: 1   #任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b   # 标识产品类型
                spec:
                  hostNetwork: true    # 可选值，根据实际情况填写，true支持hostIP创建Pod，false不支持hostIP创建Pod
                  affinity:            # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: job-name
                                operator: In
                                values:
                                  - default-test-mindspore        # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                  nodeSelector:
                    host-arch: huawei-arm              # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8  # 节点类型
                  containers:
                  - name: ascend                            # 必须为ascend，不能修改
        ...
                    env:                                    
                      - name: HCCL_IF_IP                    # 可选值，根据实际情况填写
                        valueFrom:                          # 若hostNetwork配置为true，需要同步配置HCCL_IF_IP环境变量
                          fieldRef:                         # 若hostNetwork未配置或配置为false，不可配置HCCL_IF_IP环境变量
                            fieldPath: status.hostIP        # 
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                    ports:                          # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 8  # 申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8  #与limits取值一致
                    volumeMounts:
        ...
                  volumes:
        ...
        ```

    >[!NOTE] 说明 
    >整卡调度或静态vNPU调度特性配置YAML的操作只在步骤1中有区别，整卡调度和静态vNPU调度特性在步骤1之后的操作相同。

2.  <a name="li118885168281"></a>若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu“和“memory“参数和对应的参数值，具体数值请根据实际情况配置。

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

3.  修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```
              volumeMounts:
              - name: ascend-server-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ```

4.  如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
    -   **TensorFlow命令参数**

        ```
           command:
          - /bin/bash
          - -c
        args: [ "cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code/ /job/output/ tensorflow/resnet_ctl_imagenet_main.py --data_dir=/job/data/resnet50/imagenet_TF/ --distribution_strategy=one_device --use_tf_while_loop=true --epochs_between_evals=1 --skip_eval --enable_checkpoint_and_export" ]
        ...
        ```

    -   **PyTorch命令参数**

        ```
        command:
          - /bin/bash
          - -c
        args: ["cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code /job/output main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --epochs=90 --batch-size=512"]
        ...
        ```

    -   **MindSpore命令参数**

        ```
        command:
          - /bin/bash
          - -c
        args: ["cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code/ /job/code/output train.py  --data_path=/job/data/resnet50/imagenet/train --config=/job/code/config/resnet50_imagenet2012_config.yaml"]
        ...
        ```

        >[!NOTE] 说明 
        >以TensorFlow命令参数为例。
        >-   /job/code/：为步骤[3](#li112747151117)中用户自定义的容器中训练脚本路径。
        >-   /job/output/：步骤[3](#li112747151117)中用户自定义的容器中训练数据集路径。
        >-   tensorflow/resnet\_ctl\_imagenet\_main.py：启动训练脚本路径。

5.  YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

    ```
    ...
              volumeMounts:
              - name: ascend-server-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ...
               # 可选，使用组件为训练任务生成RankTable文件，需要新增以下加粗字段，设置容器中hccl.json文件保存路径。该路径不可修改。
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
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
             # 可选，使用组件为PyTorch和MindSpore框架生成RankTable文件，需要新增以下加粗字段，设置hccl.json文件保存路径
            - name: ranktable           # 请勿修改此参数的默认值，Ascend Operator会用于检查是否开启文件挂载hccl.json。
              hostPath:                 #请使用hostpath挂载或NFS挂载
                path: /user/mindx-dl/ranktable/default.default-test-pytorch   # 共享存储或者本地存储路径，/user/mindx-dl/ranktable/为前缀路径，必须和Ascend Operator挂载的Ranktable根目录保持一致。default.default-test-pytorch为后缀路径，建议改为:namespace.job-name。
    ```

**通过文件配置资源信息场景<a name="section6131855154814"></a>**

1.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    **表 2**  操作参考

    <a name="table353271710226"></a>
    <table><thead align="left"><tr id="row8532181713223"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p8532151719228"><a name="p8532151719228"></a><a name="p8532151719228"></a>特性名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p853216179225"><a name="p853216179225"></a><a name="p853216179225"></a>操作示例</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row4532141711226"><td class="cellrowborder" rowspan="2" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p16533181762219"><a name="p16533181762219"></a><a name="p16533181762219"></a>整卡调度</p>
    <p id="p144064285710"><a name="p144064285710"></a><a name="p144064285710"></a></p>
    <p id="p1966518379556"><a name="p1966518379556"></a><a name="p1966518379556"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p2533121714227"><a name="p2533121714227"></a><a name="p2533121714227"></a><a href="#li103534014484">在Atlas 800 训练服务器上创建单机任务</a></p>
    </td>
    </tr>
    <tr id="row1753371710227"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p053317172221"><a name="p053317172221"></a><a name="p053317172221"></a><a href="#li21411371493">在Atlas 800 训练服务器上创建分布式任务</a></p>
    </td>
    </tr>
    <tr id="row173101655202217"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1366503795516"><a name="p1366503795516"></a><a name="p1366503795516"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p133113557220"><a name="p133113557220"></a><a name="p133113557220"></a><a href="#li1487005712813">在Atlas 800T A2 训练服务器上创建分布式任务</a></p>
    <div class="note" id="note1836416491472"><a name="note1836416491472"></a><a name="note1836416491472"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p918205224718"><a name="p918205224718"></a><a name="p918205224718"></a>若需要使用<span id="ph181814524472"><a name="ph181814524472"></a><a name="ph181814524472"></a>PyTorch</span>或<span id="ph318155294712"><a name="ph318155294712"></a><a name="ph318155294712"></a>MindSpore</span>框架支持的交换机亲和性调度，配置示例请参见<a href="#li1460553372">配置交换机亲和性调度参考示例</a>。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row1140175742214"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p24055715225"><a name="p24055715225"></a><a name="p24055715225"></a>静态vNPU调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1140155720221"><a name="p1140155720221"></a><a name="p1140155720221"></a><a href="#li1328115394814">在Atlas 800 训练服务器上创建单机任务</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    -   <a name="li103534014484"></a>使用**整卡调度**特性，参考本配置。以a800\_tensorflow\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用8个芯片，修改示例如下。

        ```
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-910   # 标识任务使用的芯片的产品类型
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API。
        kind: Job                               # 目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名，可自定义
        ...
        spec:
          minAvailable: 1                  # 单机为1
        ...
          - name: "default-test"
              replicas: 1                  # 单机为1
              template:
                metadata:
        ...
                spec:
        ...
                   containers:
                   - image: tensorflow-test:latest   #镜像名称
        ...
                     env:
        ...
                     - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime使用该字段
                       valueFrom:
                         fieldRef:
                           fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                    resources:  
                      requests:
                        huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8。
                      limits:
                        huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
        ...
                    nodeSelector:
                      host-arch: huawei-arm              # 可选值，根据实际情况填写
                      accelerator-type: module        # 调度到Atlas 800 训练服务器
        ...
        ```

        修改完成后执行[步骤2](#li832632419711)，配置YAML的其他字段。

    -   <a name="li21411371493"></a>使用**整卡调度**特性，参考本配置。以a800\_tensorflow\_vcjob.yaml为例，在两台Atlas 800 训练服务器节点创建**分布式训练**任务，任务使用2\*8个芯片，修改示例如下，分布式任务的每个Pod只能调度到不同节点。

        ```
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-910  # 标识任务使用的芯片的产品类型
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API
        kind: Job                               # 目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名，可自定义
        ...
        spec:
          minAvailable: 2                  # 2节点分布式任务则为2，N节点则为N，Deployment类型的任务不需要该参数
        ...
          - name: "default-test"
              replicas: 2                  # N节点分布式场景为N
              template:
                metadata:
        ...
                spec:
                  affinity:                            # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: volcano.sh/job-name      # vcjob固定字段，当任务类型为deployment时，key为deploy-name
                                operator: In                   # 固定字段
                                values:
                                  - mindx-dls-test             # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                containers:
                - image: tensorflow-test:latest  # 镜像名称
        ...
                  env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        
                    resources:  
                      requests:
                        huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8。可在下方添加行，配置memory、cpu等资源
                      limits:
                        huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
        ...
                    nodeSelector:
                      host-arch: huawei-arm               # 可选值，根据实际情况填写
                      accelerator-type: module     # 调度到Atlas 800 训练服务器
        ...
        ```

        修改完成后执行[步骤2](#li832632419711)，配置YAML的其他字段。

    -   <a name="li1487005712813"></a>使用**整卡调度**特性，参考本配置。以a800\_tensorflow\_vcjob.yaml为例，在两台Atlas 800T A2 训练服务器节点创建**分布式训练**任务，任务使用2\*8个芯片，修改示例如下，分布式任务的每个Pod只能调度到不同节点。

        ```
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-{xxx}b   # 产品类型
        ..
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改，必须使用Volcano的API
        kind: Job                               # 目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名字
        ...
          labels:
            ring-controller.atlas: ascend-{xxx}b   # 必须与ConfigMap中的标签保持一致，不可修改
        ...
        spec:
          minAvailable: 2                      # 此处建议与下面的为节点个数保持一致
          schedulerName: volcano                # 使用Volcano进行调度
        ...
          tasks:
          - name: "default-test"
            replicas: 2                         # 此处为节点个数
            template:
              metadata:
                labels:
                  app: tf
                  ring-controller.atlas: ascend-{xxx}b  # 必须与ConfigMap中的标签一致，不可修改
              spec:
                affinity:                                   # 本段配置表示分布式任务的Pod调度到不同节点
                  podAntiAffinity:
                    requiredDuringSchedulingIgnoredDuringExecution:
                      - labelSelector:
                          matchExpressions:
                            - key: volcano.sh/job-name      # vcjob固定字段，当任务类型为deployment时，key为deploy-name
                              operator: In                   # 固定字段
                              values:
                                - mindx-dls-test             # 需要和上面的任务名一致                  
                        topologyKey: kubernetes.io/hostname
                containers:
                - image: tensorflow-test:latest               # 训练框架镜像，根据实际情况修改
        ...
                  env:
        ...
                  - name: XDL_IP                 # 本段固定不变
                    valueFrom:
                      fieldRef:
                        fieldPath: status.hostIP
                  - name: framework
                    value: "Tensorflow"          # 根据实际框架变化进行修改
                  - name: ASCEND_VISIBLE_DEVICES                       # 会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                  resources:
                    requests:
                      huawei.com/Ascend910: 8    # 每台Atlas 800T A2 训练服务器芯片数量最多为8
                    limits:
                      huawei.com/Ascend910: 8    # 每台Atlas 800T A2 训练服务器芯片数量最多为8
        ...
                nodeSelector:
                  host-arch: huawei-arm              # 可选值，根据实际情况填写
                  accelerator-type: module-{xxx}b-8          # 调度到Atlas 800T A2 训练服务器节点
        ...
        ```

        修改完成后执行[步骤2](#li832632419711)，配置YAML的其他字段。

    -   <a name="li1460553372"></a>使用**整卡调度**特性，参考本配置。PyTorch和MindSpore框架新增了使用交换机亲和性调度的功能，该功能支持大模型任务和普通任务。以a800\_pytorch\_vcjob.yaml为例，在一台Atlas 800T A2 训练服务器节点创建**分布式训练**任务，任务使用1\*8个芯片，修改示例如下。

        ```
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
          namespace: vcjob                      
          labels:
            ring-controller.atlas: ascend-{xxx}b   # 产品类型
        data:
          hccl.json: |
            {
                "status":"initializing"
            }
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改，必须使用Volcano的API
        kind: Job                               # 目前只支持Job类型
        metadata:
        ...                 
          labels:
            ring-controller.atlas: ascend-{xxx}b   # 必须与ConfigMap中的标签保持一致，不可修改
            fault-scheduling: "force"
            tor-affinity: "normal-schema"      # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
        spec:
          minAvailable: 1                       # 此处建议与下面的为节点个数保持一致
          schedulerName: volcano                # 使用Volcano进行调度
        ...
          tasks:
          - name: "default-test"
            replicas: 1                              # 此处为节点个数
            template:
              metadata:
                labels:
                  app: pytorch
                  ring-controller.atlas: ascend-{xxx}b  # 必须与ConfigMap中的标签一致，不可修改
              spec:
                affinity:                           # 本段配置表示分布式任务的Pod调度到不同节点
                  podAntiAffinity:
                    requiredDuringSchedulingIgnoredDuringExecution:
                      - labelSelector:
                          matchExpressions:
                            - key: volcano.sh/job-name
                              operator: In
                              values:
                                - mindx-dls-test
                        topologyKey: kubernetes.io/hostname
                hostNetwork: true
                containers:
                - image: torch:b030               # 训练框架镜像，根据实际情况修改
                  - name: XDL_IP                   # 本段固定不变
                    valueFrom:
                      fieldRef:
                        fieldPath: status.hostIP
                  - name: POD_UID
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.uid
                  - name: framework
                    value: "PyTorch"
        ...
                  - name: ASCEND_VISIBLE_DEVICES
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                  resources:
                    requests:
                      huawei.com/Ascend910: 8                # 每台Atlas 800T A2 训练服务器芯片数量最多为8
                    limits:
                      huawei.com/Ascend910: 8                 # 每台Atlas 800T A2 训练服务器芯片数量最多为8
         ...
                nodeSelector:
                  host-arch: huawei-x86       # 可选值，根据实际情况填写
                  accelerator-type: module-{xxx}b-8    #调度到Atlas 800T A2 训练服务器节点
        ...
        ```

        >[!NOTE] 说明 
        >其余示例可参考[表5](#选择yaml示例)和[表4](#选择yaml示例)，以及YAML对应的参数说明[表2](#yaml参数说明)进行适配修改。修改完成后执行[步骤2](#li832632419711)，继续配置yaml的其他字段。

    -   <a name="li1328115394814"></a>使用**静态vNPU调度**特性，参考本配置。以a800\_tensorflow\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，申请2个AI Core的任务为例，修改示例如下。静态vNPU调度特性只支持**单机训练**任务。

        ```
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-910   # 产品类型
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改，必须使用Volcano的API
        kind: Job                               #目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名，可自定义
        ...
        spec:
          minAvailable: 1                  # 如果使用静态vNPU调度，此处取值需为1
        ...
          - name: "default-test"
              replicas: 1                  # 如果使用静态vNPU调度，此处取值需为1
              template:
                metadata:
        ...
                spec:
        ...
                containers:
                - image: tensorflow-test:latest  # 训练镜像
        ...
                  env:
        ...
                 # 静态vNPU调度暂不支持ASCEND_VISIBLE_DEVICES相关字段，需要删除以下加粗字段
                  - name: ASCEND_VISIBLE_DEVICES                                   
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']              
        ...
                    resources:  
                      requests:
                        huawei.com/Ascend910-2c: 1          # 如果使用静态vNPU调度，此处数量只能为1
                      limits:
                        huawei.com/Ascend910-2c: 1          # 如果使用静态vNPU调度此处数量只能为1
        ...
                    nodeSelector:
                      host-arch: huawei-arm               # 可选值，根据实际情况填写
                      accelerator-type: module    # 调度到Atlas 800 训练服务器上
        ...
        ```

        修改完成后执行[步骤2](#li832632419711)，配置YAML的其他字段。

    >[!NOTE] 说明 
    >整卡调度或静态vNPU调度特性配置YAML的操作只在步骤1中有区别，整卡调度和静态vNPU调度特性在步骤1之后的操作相同。

2.  <a name="li832632419711"></a>若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu“和“memory“参数和对应的参数值，具体数值请根据实际情况配置。

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

3.  <a name="li112747151117"></a>修改训练脚本、代码的挂载路径。

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

4.  如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
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
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --world-size=1 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024;"
        ...
        ```

    -   **MindSpore命令参数**

        ```
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ train.py  --config_path=/job/code/config/resnet50_imagenet2012_config.yaml --output_dir=/job/output --run_distribute=True --device_num=8 --data_path=/job/data/imagenet/train"
        ...
        ```

        >[!NOTE] 说明 
        >以TensorFlow命令参数为例。
        >-   /job/code/：步骤[3](#li112747151117)中用户自定义的容器中训练脚本路径。
        >-   /job/output/：步骤[3](#li112747151117)中用户自定义的容器中训练数据集路径。
        >-   tensorflow/resnet\_ctl\_imagenet\_main.py：启动训练脚本路径。

5.  YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

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
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
    ```



#### 下发任务<a name="ZH-CN_TOPIC_0000002511427065"></a>

**操作步骤<a name="zh-cn_topic_0000001608595413_section15243115731317"></a>**

1.  示例a800\_tensorflow\_vcjob.yaml中，任务部署在vcjob命名空间下，因此需要在管理节点执行以下命令，为训练任务创建命名空间。如果任务创建到非默认的命名空间，则需要根据实际情况创建命名空间。

    ```
    kubectl create namespace vcjob
    ```

2.  在管理节点示例YAML所在路径，执行以下命令，使用YAML下发训练任务。

    ```
    kubectl apply -f XXX.yaml
    ```

    >[!NOTE] 说明 
    >如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f** **_XXX__.yaml_命令删除原任务，再重新下发任务。

    -   通过环境变量配置资源信息场景的示例如下：

        ```
        kubectl apply -f tensorflow_standalone_acjob.yaml
        ```

        回显示例如下：

        ```
        ascendjob.mindxdl.gitee.com/default-tensorflow-test created
        ```

    -   通过文件配置资源信息场景的示例如下：

        ```
        kubectl apply -f a800_tensorflow_vcjob.yaml
        ```

        回显示例如下：

        ```
        configmap/rings-config-mindx-dls-test created
        job.batch.volcano.sh/mindx-dls-test created
        ```

>[!NOTE] 说明 
>-   若下发训练任务后，任务一直处于Pending状态，可以参见[训练任务处于Pending状态，原因：nodes are unavailable](../faq.md#训练任务处于pending状态原因nodes-are-unavailable)或者[资源不足时，任务处于Pending状态](../faq.md#资源不足时任务处于pending状态)章节进行处理。
>-   若成功启动训练任务后，发现训练任务容器内部hccl.json文件处于initializing状态，可以参见[hccl.json文件没有生成](../faq.md#hccljson文件没有生成)章节进行处理。


#### 查看任务进程<a name="ZH-CN_TOPIC_0000002479387130"></a>

**操作步骤<a name="zh-cn_topic_0000001558675462_section15243115731317"></a>**

1.  在管理节点查看任务Pod的状态，需要保证Pod状态为Running。

    执行以下命令，查看Pod运行情况。

    ```
    kubectl get pod --all-namespaces -o wide
    ```

    -   单机单芯片训练任务回显示例。

        ```
        NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE           NOMINATED NODE   READINESS GATES
        ...
        vcjob            mindx-dls-test-default-test-0             1/1     Running            0          4m      192.168.243.198   ubuntu         <none>           <none>
        ...
        ```

    -   两个训练节点，执行2\*8芯片分布式训练任务回显示例。

        ```
        NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE           NOMINATED NODE   READINESS GATES
        ...
        vcjob            mindx-dls-test-default-test-0             1/1     Running            0          3m      192.168.243.198   ubuntu         <none>           <none>
        vcjob            mindx-dls-test-default-test-1             1/1     Running            0          3m      192.168.243.199   ubuntu         <none>           <none>
        ...
        ```

2.  查看计算节点的NPU分配情况，在管理节点执行以下命令查看。

    ```
    kubectl describe nodes {任务运行节点的节点名}
    ```

    -   使用**整卡调度**特性，单机单芯片训练任务回显示例。

        ```
        Name:               ubuntu
        Roles:              master,worker
        Labels:             accelerator=huawei-Ascend910
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests        Limits
          --------              --------        ------
          cpu                   37250m (19%)    37500m (19%)
          memory                117536Mi (15%)  119236Mi (15%)
          ephemeral-storage     0 (0%)          0 (0%)
          huawei.com/Ascend910  1               1
        Events:                 <none>
        ```

        >[!NOTE] 说明 
        >**Allocated resources**的字段huawei.com/Ascend910的值为1，表明训练使用了一个NPU。

    -   使用**静态vNPU调度**特性，单机单芯片训练任务回显示例。

        ```
        Name:               ubuntu
        Roles:              master,worker
        Labels:             accelerator=huawei-Ascend910
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests        Limits
          --------              --------        ------
          cpu                   37250m (19%)    37500m (19%)
          memory                117536Mi (15%)  119236Mi (15%)
          ephemeral-storage     0 (0%)          0 (0%)
          huawei.com/Ascend910-2c  1               1
        Events:                 <none>
        ```

        >[!NOTE] 说明 
        >**Allocated resources**的字段**huawei.com/Ascend910-2c**的值为1，表明训练使用了一个包含了2个AI Core的vNPU。

    -   两个训练节点，执行2\*8芯片分布式训练任务，查看其中一个节点示例。**静态vNPU调度**不支持分布式训练任务。

        ```
        Name:               ubuntu
        Roles:              master,worker
        Labels:             accelerator=huawei-Ascend910
                            beta.kubernetes.io/arch=arm64
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests        Limits
          --------              --------        ------
          cpu                   37250m (19%)    37500m (19%)
          memory                117536Mi (15%)  119236Mi (15%)
          ephemeral-storage     0 (0%)          0 (0%)
          huawei.com/Ascend910  8               8
        Events:                 <none>
        ```

        >[!NOTE] 说明 
        >**Allocated resources**的字段huawei.com/Ascend910的值为8，表明分布式训练使用了节点上所有的NPU。

3.  查看Pod的NPU使用情况。

    本例中使用**kubectl describe pod mindx-dls-test-default-test-0 -n vcjob**命令查看运行Pod的情况。

    -   单机单芯片训练任务示例，有如下加粗的内容表示正常。

        ```
        root@ubuntu:/home/test/yaml# kubectl describe pod mindx-dls-test-default-test-0 -n vcjob
        Name:         mindx-dls-test-default-test-0
        Namespace:    vcjob
        Priority:     0
        Node:         ubuntu/XXX.XXX.XXX.XXX
        Start Time:   Wed, 30 Sep 2020 15:38:22 +0800
        Labels:       app=tf
                      ring-controller.atlas=ascend-910
                      volcano.sh/job-name=mindx-dls-test
                      volcano.sh/job-namespace=vcjob
        Annotations:  ascend.kubectl.kubernetes.io/ascend-910-configuration:
                        {"pod_name":"0","server_id":"xx-xx-xx-xx","devices":[{"device_id":"3","device_ip":"192.168.20.102"}...
                      cni.projectcalico.org/podIP: 192.168.243.195/32
                      cni.projectcalico.org/podIPs: 192.168.243.195/32
                      huawei.com/Ascend910: Ascend910-3
                      huawei.com/AscendReal: Ascend910-3
                      huawei.com/kltDev: Ascend910-3
                      predicate-time: 18446744073709551615
                      scheduling.k8s.io/group-name: mindx-dls-test
                      volcano.sh/job-name: mindx-dls-test
                      volcano.sh/job-version: 0
                      volcano.sh/task-spec: default-test
        Status:       Running
        ```

    -   两个训练节点，执行2\*8芯片分布式训练任务示例，有如下加粗的内容表示正常。

        ```
        root@ubuntu:/home/test/yaml# kubectl describe pod mindx-dls-test-default-test-0 -n vcjob
        Name:         mindx-dls-test-default-test-0
        Namespace:    vcjob
        Priority:     0
        Node:         ubuntu/XXX.XXX.XXX.XXX
        Start Time:   Wed, 30 Sep 2020 15:38:22 +0800
        Labels:       app=tf
                      ring-controller.atlas=ascend-910
                      volcano.sh/job-name=mindx-dls-test
                      volcano.sh/job-namespace=vcjob
        Annotations:  ascend.kubectl.kubernetes.io/ascend-910-configuration:
                        {"pod_name":"0","server_id":"xx-xx-xx-xx","devices":[{"device_id":"0","device_ip":"192.168.20.100"}...
                      cni.projectcalico.org/podIP: 192.168.243.195/32
                      cni.projectcalico.org/podIPs: 192.168.243.195/32
                      huawei.com/Ascend910: Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7
                      huawei.com/AscendReal: Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7,Ascend910-0
                      huawei.com/kltDev: Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7,Ascend910-0,Ascend910-1,Ascend910-2
                      predicate-time: 18446744073709551615
                      scheduling.k8s.io/group-name: mindx-dls-test
                      volcano.sh/job-name: mindx-dls-test
                      volcano.sh/job-version: 0
                      volcano.sh/task-spec: default-test
        Status:       Running
        ```


#### 查看整卡调度或静态vNPU调度结果<a name="ZH-CN_TOPIC_0000002479387140"></a>

**TensorFlow<a name="zh-cn_topic_0000001609474257_section1188814431750"></a>**

1.  在执行如下命令，查看训练结果。

    ```
    kubectl logs -n  <namespace> <pod-name>
    ```

    如：

    ```
    kubectl logs -n vcjob mindx-dls-test-default-test-0
    ```

2.  查看训练日志，如果出现如下内容表示训练成功。

    ```
    ...
    I1123 16:20:11.016411 139889740781376 controller.py:458] train | step:    112 | steps/sec:    4.0 | output: {'train_accuracy': 0.0, 'train_loss': 12.339745}
    train | step:    112 | steps/sec:    4.0 | output: {'train_accuracy': 0.0, 'train_loss': 12.339745}
    2022-11-23 16:20:11.541361: I core/op_executors/npu_concrete_graph.cpp:84] Start consume iterator resource AnonymousIterator0 4 times
    2022-11-23 16:20:11.541499: I core/op_executors/npu_concrete_graph.cpp:118] Start run ge graph 445 pin to cpu, loop size 4
    2022-11-23 16:20:11.565552: I core/op_executors/npu_concrete_graph.cpp:92] Iterator resource AnonymousIterator0 consume 4 times done with status OK
    I1123 16:20:12.046172 139889740781376 controller.py:458] train | step:    116 | steps/sec:    4.0 | output: {'train_accuracy': 0.0, 'train_loss': 12.389724}
    train | step:    116 | steps/sec:    4.0 | output: {'train_accuracy': 0.0, 'train_loss': 12.389724}
    2022-11-23 16:20:12.542817: I core/op_executors/npu_concrete_graph.cpp:84] Start consume iterator resource AnonymousIterator0 4 times
    2022-11-23 16:20:12.542937: I core/op_executors/npu_concrete_graph.cpp:118] Start run ge graph 445 pin to cpu, loop size 4
    2022-11-23 16:20:12.571535: I core/op_executors/npu_concrete_graph.cpp:92] Iterator resource AnonymousIterator0 consume 4 times done with status OK
    I1123 16:20:13.038832 139889740781376 controller.py:458] train | step:    120 | steps/sec:    4.2 | output: {'train_accuracy': 0.0, 'train_loss': 12.421794}
    train | step:    120 | steps/sec:    4.2 | output: {'train_accuracy': 0.0, 'train_loss': 12.421794}
    2022-11-23 16:20:13.559254: I core/op_executors/npu_concrete_graph.cpp:84] Start consume iterator resource AnonymousIterator0 4 times
    2022-11-23 16:20:13.559394: I core/op_executors/npu_concrete_graph.cpp:118] Start run ge graph 445 pin to cpu, loop size 4
    2022-11-23 16:20:13.604791: I core/op_executors/npu_concrete_graph.cpp:92] Iterator resource AnonymousIterator0 consume 4 times done with status OK
    I1123 16:20:14.052418 139889740781376 controller.py:458] train | step:    124 | steps/sec:    4.1 | output: {'train_accuracy': 0.0, 'train_loss': 12.335646}
    train | step:    124 | steps/sec:    4.1 | output: {'train_accuracy': 0.0, 'train_loss': 12.335646}
    2022-11-23 16:20:14.555126: I core/op_executors/npu_concrete_graph.cpp:84] Start consume iterator resource AnonymousIterator0 4 times
    2022-11-23 16:20:14.555217: I core/op_executors/npu_concrete_graph.cpp:118] Start run ge graph 445 pin to cpu, loop size 4
    2022-11-23 16:20:14.601171: I core/op_executors/npu_concrete_graph.cpp:92] Iterator resource AnonymousIterator0 consume 4 times done with status OK
    I1123 16:20:15.058790 139889740781376 controller.py:458] train | step:    128 | steps/sec:    4.1 | output: {'train_accuracy': 0.0, 'train_loss': 12.415506}
    train | step:    128 | steps/sec:    4.1 | output: {'train_accuracy': 0.0, 'train_loss': 12.415506}
    I1123 16:20:15.228246 139889740781376 resnet_ctl_imagenet_main.py:191] Run stats:
    {'step_timestamp_log': ['BatchTimestamp<batch_index: 0, timestamp: 1669191532.9730577>', 'BatchTimestamp<batch_index: 100, timestamp: 1669191607.7925153>'], 'train_finish_time': 1669191615.2273297, 'avg_exp_per_second': 24.973437296848516}
    2022-11-23 16:20:15.232802: I core/npu_logger.cpp:58] Stopping npu stdout receiver of device 0
    2022-11-23 16:20:15.232901: I core/npu_device.cpp:122] Stopping iterator resource provider for AnonymousMultiDeviceIterator0
    2022-11-23 16:20:15.233013: I core/npu_device.cpp:122] Stopping iterator resource provider for AnonymousIterator0
    2022-11-23 16:20:15.235151: I core/npu_wrapper.cpp:230] Stop tensorflow model parser succeed
    2022-11-23 16:20:18.289648: I core/npu_wrapper.cpp:240] Stop graph engine succeed
    ...
    ```

3.  进入模型输出目录，查看生成的模型文件。

    ```
    drwxr-xr-x 1 root root       4096 Dec  2 11:36 ./
    drwxrwxrwx 1 root root       4096 Dec  2 11:36 ../
    -rw-r--r--. 1 root root       999 Dec  2 11:36 checkpoint
    -rw-r--r--. 1 root root 306986892 Dec  2 11:35 ckpt-111.data-00000-of-00001
    -rw-r--r--. 1 root root     44311 Dec  2 11:35 ckpt-111.index
    -rw-r--r--. 1 root root 306986892 Dec  2 11:36 ckpt-128.data-00000-of-00001
    -rw-r--r--. 1 root root     44311 Dec  2 11:36 ckpt-128.index
    ```

**PyTorch<a name="zh-cn_topic_0000001609474257_section15657195014514"></a>**

1.  在执行如下命令，查看训练结果。

    ```
    kubectl logs -n  <namespace> <pod-name>
    ```

    如：

    ```
    kubectl logs -n vcjob mindx-dls-test-default-test-0
    ```

2.  查看训练日志，如果出现如下内容表示训练成功。

    ```
    [gpu id: 0 ] Test: [77/85]      Time  0.117 ( 0.281)    Loss 1.073741e+01 (1.078090e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.00 (  0.12)
    [gpu id: 0 ] Test: [78/85]      Time  0.114 ( 0.279)    Loss 1.072909e+01 (1.078015e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.00 (  0.12)
    [gpu id: 0 ] Test: [79/85]      Time  0.115 ( 0.277)    Loss 1.073733e+01 (1.077953e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.20 (  0.12)
    [gpu id: 0 ] Test: [80/85]      Time  2.385 ( 0.306)    Loss 1.087646e+01 (1.078090e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.00 (  0.12)
    [gpu id: 0 ] Test: [81/85]      Time  1.139 ( 0.318)    Loss 1.075754e+01 (1.078058e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.39 (  0.12)
    [gpu id: 0 ] Test: [82/85]      Time  0.115 ( 0.315)    Loss 1.068419e+01 (1.077925e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.20 (  0.13)
    [gpu id: 0 ] Test: [83/85]      Time  0.129 ( 0.313)    Loss 1.075079e+01 (1.077887e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.20 (  0.13)
    [gpu id: 0 ] Test: [84/85]      Time  0.134 ( 0.310)    Loss 1.093459e+01 (1.078095e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.39 (  0.13)
    [gpu id: 0 ] [AVG-ACC] * Acc@1 0.016 Acc@5 0.130
    validate acc1 tensor(0.0156, device='npu:0')
    Complete 90 epoch training, take time:1.05h
    ...
    ```

3.  进入模型输出目录，查看生成的模型文件。

    ```
    drwxrwx--- 2 root root      4096 Mar  4 19:28 ./
    drwxrwx--- 4 root root      4096 Mar  4 19:28 ../
    -rw-rw---- 1 root root 102489869 Mar  4 19:28 checkpoint_npu0model_best.pth.tar
    -rw-rw---- 1 root root 102489869 Mar  4 19:28 checkpoint_npu0.pth.tar
    ...
    ```

    可以参考ModelZoo上，PyTorch框架的[ResNet-50](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_for_PyTorch#%E6%A8%A1%E5%9E%8B%E6%8E%A8%E7%90%86)模型中的“模型推理”章节，对生成的模型文件进行模型转换处理。

**MindSpore<a name="zh-cn_topic_0000001609474257_section1533335719514"></a>**

1.  在执行如下命令，查看训练结果。

    ```
    kubectl logs -n  <namespace> <pod-name>
    ```

    如：

    ```
    kubectl logs -n vcjob mindx-dls-test-default-test-0
    ```

2.  查看训练日志，如果出现如下内容表示训练成功。

    ```
    ...
    2023-06-09 17:55:04,837:INFO:epoch: [70/90] loss: 1.541062, epoch time: 7.563 s, per step time: 157.554 ms
    2023-06-09 17:55:10,540:INFO:epoch: [71/90] loss: 1.544771, epoch time: 5.702 s, per step time: 118.796 ms
    2023-06-09 17:55:16,347:INFO:epoch: [72/90] loss: 1.506525, epoch time: 5.807 s, per step time: 120.979 ms
    2023-06-09 17:55:24,904:INFO:epoch: [73/90] loss: 1.519342, epoch time: 8.556 s, per step time: 178.260 ms
    2023-06-09 17:55:29,887:INFO:epoch: [74/90] loss: 1.387423, epoch time: 4.982 s, per step time: 103.783 ms
    2023-06-09 17:55:39,785:INFO:epoch: [75/90] loss: 1.440862, epoch time: 9.897 s, per step time: 206.194 ms
    2023-06-09 17:55:48,780:INFO:epoch: [76/90] loss: 1.431275, epoch time: 8.995 s, per step time: 187.399 ms
    2023-06-09 17:55:55,764:INFO:epoch: [77/90] loss: 1.411003, epoch time: 6.984 s, per step time: 145.492 ms
    2023-06-09 17:56:03,962:INFO:epoch: [78/90] loss: 1.457689, epoch time: 8.198 s, per step time: 170.783 ms
    2023-06-09 17:56:11,517:INFO:epoch: [79/90] loss: 1.410896, epoch time: 7.554 s, per step time: 157.372 ms
    2023-06-09 17:56:16,643:INFO:epoch: [80/90] loss: 1.517990, epoch time: 5.126 s, per step time: 106.789 ms
    2023-06-09 17:56:23,364:INFO:epoch: [81/90] loss: 1.342399, epoch time: 6.720 s, per step time: 140.005 ms
    2023-06-09 17:56:31,835:INFO:epoch: [82/90] loss: 1.352396, epoch time: 8.471 s, per step time: 176.470 ms
    2023-06-09 17:56:36,971:INFO:epoch: [83/90] loss: 1.358075, epoch time: 5.135 s, per step time: 106.984 ms
    2023-06-09 17:56:44,259:INFO:epoch: [84/90] loss: 1.400720, epoch time: 7.288 s, per step time: 151.838 ms
    2023-06-09 17:56:52,868:INFO:epoch: [85/90] loss: 1.371813, epoch time: 8.608 s, per step time: 179.339 ms
    2023-06-09 17:56:57,613:INFO:epoch: [86/90] loss: 1.303416, epoch time: 4.745 s, per step time: 98.858 ms
    2023-06-09 17:57:04,177:INFO:epoch: [87/90] loss: 1.290425, epoch time: 6.564 s, per step time: 136.744 ms
    2023-06-09 17:57:11,797:INFO:epoch: [88/90] loss: 1.298486, epoch time: 7.619 s, per step time: 158.738 ms
    2023-06-09 17:57:16,807:INFO:epoch: [89/90] loss: 1.297104, epoch time: 5.009 s, per step time: 104.363 ms
    2023-06-09 17:57:25,568:INFO:epoch: [90/90] loss: 1.401816, epoch time: 8.759 s, per step time: 182.486 ms
    ```

3.  进入模型输出目录，查看生成的模型文件。

    ```
    drwx------  2 root root      4096 Dec 21 15:35 ./
    drwxrwxrwx 10 root root      4096 Dec 21 15:26 ../
    -r--------  1 root root 188546464 Dec 21 15:31 resnet50-45_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:31 resnet50-50_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:32 resnet50-55_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:32 resnet50-60_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:33 resnet50-65_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:33 resnet50-70_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:33 resnet50-75_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:34 resnet50-80_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:34 resnet50-85_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:35 resnet50-90_234.ckpt
    -rw-------  1 root root    769071 Dec 21 15:28 resnet50-graph.meta
    ```


#### 删除任务<a name="ZH-CN_TOPIC_0000002479227168"></a>

**操作步骤<a name="zh-cn_topic_0000001609474265_section1595872772813"></a>**

在示例YAML所在路径下，执行以下命令，删除对应的训练任务。

```
kubectl delete -f XXX.yaml
```

-   通过环境变量配置资源信息场景的示例如下：

    ```
    kubectl delete -f tensorflow_standalone_acjob.yaml
    ```

    回显示例如下：

    ```
    ascendjob.mindxdl.gitee.com "default-test-tensorflow" deleted
    ```

-   通过文件配置资源信息场景的示例如下：

    ```
    kubectl delete -f a800_tensorflow_vcjob.yaml
    ```

    回显示例如下：

    ```
    configmap "rings-config-mindx-dls-test" deleted
    job.batch.volcano.sh "mindx-dls-test" deleted
    ```

>[!NOTE] 说明 
>若删除训练任务后，Pod一直处于Terminating状态，可以参见[手动删除vcjob后Pod一直处于Terminating状态](../faq.md#手动删除vcjob后pod一直处于terminating状态)章节进行处理。



### 通过命令行使用（其他调度器）<a name="ZH-CN_TOPIC_0000002511427069"></a>

通过命令行使用（其他调度器）和通过命令行使用（Volcano）使用流程一致，只有任务YAML有所不同，用户可以准备好相应YAML后参考[通过命令行使用（Volcano）](#通过命令行使用volcano)章节使用。

**操作步骤<a name="section1780564613381"></a>**

1.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    集群调度并未专门提供使用其他调度器的YAML示例，用户可以获取使用Volcano的YAML示例并做如下修改即可使用。

    以tensorflow\_standalone\_acjob.yaml为例，在一台Atlas 800T A2 训练服务器节点创建**单机训练**任务，执行1\*8芯片训练任务，修改示例如下。

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: default-test-tensorflow
      labels:
        framework: tensorflow
        ring-controller.atlas: ascend-{xxx}b   
    spec:
      schedulerName: volcano        # 使用其他调度器时，删除该字段
      runPolicy:                    # 使用其他调度器时，删除该字段
        schedulingPolicy:           
          minAvailable: 1
          queue: default
      successPolicy: AllWorkers
      replicaSpecs:
        Chief:
          replicas: 1
          restartPolicy: Never
          template:
            metadata:
              labels:
                ring-controller.atlas: ascend-{xxx}b   
            spec:
              nodeSelector:
                host-arch: huawei-arm
                accelerator-type: module-{xxx}b-8
              containers:
              - name: ascend                    
    ...
                env:
    ...
             # 使用其他调度器暂不支持ASCEND_VISIBLE_DEVICES相关字段，需要删除以下加粗字段
              - name: ASCEND_VISIBLE_DEVICES                       
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               
    ...
                resources:
                  limits:
                    huawei.com/Ascend910: 8
                  requests:
                    huawei.com/Ascend910: 8
                volumeMounts:
    ...
    ```

2.  若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu“和“memory“参数和对应的参数值，具体数值请根据实际情况配置。

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

3.  <a name="li112747151117"></a>修改训练脚本、代码的挂载路径。

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

4.  如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。

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
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --world-size=1 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024;"
        ...
        ```

    -   **MindSpore命令参数**

        ```
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ train.py  --config_path=/job/code/config/resnet50_imagenet2012_config.yaml --output_dir=/job/output --run_distribute=True --device_num=8 --data_path=/job/data/imagenet/train"
        ...
        ```

    >[!NOTE] 说明 
    >以TensorFlow命令参数为例。
    >-   /job/code/：为步骤[3](#li112747151117)中用户自定义的容器中训练脚本路径。
    >-   /job/output/：步骤[3](#li112747151117)中用户自定义的容器中训练数据集路径。
    >-   tensorflow/resnet\_ctl\_imagenet\_main.py：启动训练脚本路径。

5.  YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

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
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
    ```


### 集成后使用<a name="ZH-CN_TOPIC_0000002511347081"></a>

本章节以**整卡调度**特性为例，介绍如何将整卡调度特性集成在AI平台上的关键操作步骤。在下发训练任务时，平台需要实现获取认证文件，创建客户端，创建Job对象，创建命名空间和调用接口下发训练任务等，将整卡调度特性提供的示例YAML转换成K8s提供的Go编程语言的API对象。

**集成前说明<a name="section16646104012516"></a>**

集成操作中会涉及到很多接口，请用户根据实际情况去相关官网了解接口的详细信息，本文档不再进行二次说明。

-   K8s相关接口请根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)了解相关内容。
-   Ascend Job可参考[表1](#yaml参数说明)中的参数说明，了解相关内容。
-   Volcano Job相关接口可参见《云容器实例 API参考》中“[创建Volcano Job](https://support.huaweicloud.com/api-cci/createBatchVolcanoShV1alpha1NamespacedJob.html)”章节了解相关内容。
-   芯片型号的数值可通过**npu-smi info**命令查询，返回的“Name”字段对应信息为芯片型号，下文的\{_xxx_\}即取“910”字符作为芯片型号数值。

**集成操作<a name="section9868584469"></a>**

1.  获取K8s认证文件。

    用户根据实际情况选择合适的[集群认证方式](https://kubernetes.io/zh-cn/docs/concepts/security/controlling-access/)，创建相应的集群配置。使用ServiceAccount创建集群配置（InCluster模式）示例代码如下。

    ```
           // 使用ServiceAccount创建集群配置（InCluster模式）
           if config, err = rest.InClusterConfig(); err != nil {
                  // 使用KubeConfig文件创建集群配置
                  if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
                         panic(err.Error())
                  }
           }
    ```

2.  创建客户端Clientset。

    ```
    Client, err := NewForConfig(cfg)
    ```

    >[!NOTE] 说明 
    >NewForConfig\(cfg\)的函数原型为**NewForConfig\(c \*rest.Config\)\(\*Clientset, error\)。**
    >参数说明如下：
    >-   **\*rest.Config**：客户端配置文件，由K8s提供的接口生成；包括cluster host、证书等信息。
    >-   **\*Clientset：**Client集合，包括AscendJob client（或VolcanoJob client）和discovery client。
    >-   **error**：错误信息。

3.  创建Job对象。通过环境变量配置资源信息的用户需要创建Ascend Job对象；通过文件配置资源信息的用户需要创建Volcano Job对象。

    >[!NOTE] 说明 
    >在进行本步骤操作之前，建议用户详细阅读[准备任务YAML](#准备任务yaml)章节，了解示例YAML实现逻辑和关键字段说明，可以更好地帮助用户进行接下来的操作。

    -   创建Ascend Job对象。

        创建Ascend Job对象，初始化Ascend Job相关字段，示例如下。

        ```
        import (
           v1 "ascend-operator-apis/pkg/apis/batch/v1"
           "ascend-operator-apis/pkg/client/clientset/versioned"
           commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
           corev1 "k8s.io/api/core/v1"
           "k8s.io/apimachinery/pkg/api/resource"
           metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
           "k8s.io/client-go/tools/clientcmd"
        )
        
        func initAcJob() v1.AscendJob {
           job := newAcJob().
              initName("default-test-pytorch"). // 初始化任务名
              initNameSpace("default").         // 初始化命名空间
              initLabels(map[string]string{     // 初始化任务标签
                 "ring-controller.atlas": "ascend-{xxx}b",   // 标识任务使用的芯片的产品类型
                 "framework":             "pytorch",       // 使用的训练框架名
                 "tor-affinity":          "normal-schema", // 是否使用交换亲和性调度。value为large-model-schema，表示使用大模型调度模式；value为normal-schema，使用普通任务调度模式；value为null，表示关闭交换机亲和性调度
                 "podgroup-sched-enable": "true",  // 仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。取值为字符串"true"时，表示开启批量调度功能；取值为其他字符串时，表示批量调度功能不生效，使用普通调度。若不配置该参数，表示批量调度功能不生效，使用普通调度
              }).
              initSchedulerName("volcano"). // 初始化调度器名
              initRunPolicy(&maNum).        // 初始化RunPolicy
              initSuccessPolicy().
              addReplicaSpecs("Master", newReplica(). // 初始化Master副本
                             initRcReplicas(&rcNum).                           // 初始化pod副本数
                             initRcRestartPolicy(commonv1.RestartPolicyNever). // 初始化容器重启策略
                             initRcLabels(map[string]string{                   // 初始化Master的标签
                    "ring-controller.atlas": "ascend-{xxx}b", // 标识任务使用的芯片的产品类型
                 }).                                   //
                 initRcNodeSelector(map[string]string{ // 初始化Master的NodeSelector
                    "host-arch":        "huawei-x86",     // 可选字段，用户根据实际需求进行配置
                    "accelerator-type": "module-{xxx}b-16", // 用户根据服务器类型进行配置，取值可以参考YAML参数
                             }).
                             initRcVolumes(). // 初始化挂载项
                             addRcContainers(newContainer().
                                initContainerName("ascend").                                             // 初始化容器名
                                initContainerImage("pt-arm:b120").                                       // 初始化镜像名
                                initContainerImagePullPolicy(corev1.PullIfNotPresent).                   // 初始化镜像拉取策略
                                initContainerEnv().                                                      // 初始化容器环境变量
                                initContainerCommand([]string{"/bin/bash", "-c", "bash train_start.sh ..."}).  // 初始化容器启动命令，具体参数参考示例YAML
                                initContainerArgs([]string{"/bin/bash", "-c", "bash train_start.sh ..."}).  // 初始化容器启动命令，具体参数参考示例YAML
                                initContainerPorts(2222).                                                // 初始化容器端口
                                initContainerLimits("huawei.com/Ascend910", "8").                        // 初始化任务资源
                                initContainerRequests("huawei.com/Ascend910", "8").                      // 初始化任务资源
                                initContainerVolumeMounts()).                                            // 初始化容器挂载项
                             initReplica()).
              addReplicaSpecs("Worker", newReplica(). // 初始化Worker副本
                             initRcReplicas(&rcNum).                           // 初始化pod副本数
                             initRcRestartPolicy(commonv1.RestartPolicyNever). // 初始化容器重启策略
                             initRcLabels(map[string]string{                   // 初始化Worker的标签
                    "ring-controller.atlas": "ascend-{xxx}b", // 标识任务使用的芯片的产品类型
                 }).
                 initRcAffinity("default-test-pytorch"). // 初始化Worker的反亲和性字段
                 initRcNodeSelector(map[string]string{   // 初始化Worker的NodeSelector
                    "host-arch":        "huawei-x86",     // 用户根据实际架构配置arm架构value为huawei-arm
                    "accelerator-type": "module-{xxx}b-8", // 用户根据服务器类型进行配置，value值可以参考YAML参数
                 }).
                 initRcVolumes().
                 addRcContainers(newContainer().
                    initContainerName("ascend").                                                  // 初始化容器名
                    initContainerImage("pt-arm:b120").                                            // 初始化镜像名
                    initContainerImagePullPolicy(corev1.PullIfNotPresent).                        // 初始化镜像拉取策略
                    initContainerEnv().                                                           // 初始化容器环境变量
                    initContainerCommand([]string{"/bin/bash", "-c", "bash train_start.sh ..."}). // 初始化容器启动命令，具体参数参考示例YAML
                    initContainerArgs([]string{"/bin/bash", "-c", "bash train_start.sh ..."}).    // 初始化容器启动命令，具体参数参考示例YAML
                    initContainerPorts(2222).                                                     // 初始化容器端口
                    initContainerLimits("huawei.com/Ascend910", "8").                             // 初始化任务资源
                    initContainerRequests("huawei.com/Ascend910", "8").                           // 初始化任务资源
                    initContainerVolumeMounts()).
                 initReplica())
           return v1.AscendJob(job)
        }
        
        type acJob v1.AscendJob
        type Replica commonv1.ReplicaSpec
        type container corev1.Container
        
        func (job acJob) initRunPolicy(n *int32) acJob {
           job.Spec.RunPolicy = commonv1.RunPolicy{SchedulingPolicy: &commonv1.SchedulingPolicy{MinAvailable: n, Queue: "default"}}
           return job
        }
        ...
        func (rc Replica) initRcReplicas(rs *int32) Replica {
           rc.Replicas = rs
           return rc
        }
        ...
        func (ct container) initContainerEnv() container {
           ct.Env = []corev1.EnvVar{
              {
                 Name: "XDL_IP",
                 ValueFrom: &corev1.EnvVarSource{
                    FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.hostIP"},
                 },
              },
           }
           return ct
        }
        ```

    -   创建Volcano Job对象。
        1.  初始化Volcano Job挂载的ConfigMap。初始化ConfigMap相关字段，示例如下。

            ```
            import "k8s.io/api/core/v1"                                                 
            func newConfigMap(name string) *v1.ConfigMap {
                   cm := &v1.ConfigMap{}                           
                   cm.Name = name
                   cm.Labels = map[string]string{
                          "ring-controller.atlas": "ascend-{xxx}b",  # 标识任务使用的芯片的产品类型
                   }
                   cm.Data = map[string]string{
                          "hccl.json": `{"status": "initializing"}`,  
                   }
                   return cm
            }
            ```

        2.  初始化Volcano Job。创建Volcano Job对象，初始化Volcano Job相关字段，示例如下。

            ```
            import (
               "k8s.io/api/core/v1"
               metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
               "k8s.io/client-go/kubernetes"
               "k8s.io/client-go/tools/clientcmd"
               "volcano.sh/apis/pkg/apis/batch/v1alpha1"
               "volcano.sh/apis/pkg/client/clientset/versioned"
            )
            
            func initJob() v1alpha1.Job {
               job := newJobBuilder().
                  initNameSpace("vcjob").       // 初始化命名空间
                  initName("mindx-dls-test").   // 初始化任务名
                  initLabels(map[string]string{ // 初始化任务标签
                     "ring-controller.atlas": "ascend-{xxx}b", // 标识任务使用的芯片的产品类型
                  }).
                  initMinAvailable(2).          // minAvailable表示运行该job所要运行的最少Pod数量。只有当job中处于running状态的Pod数量不小于minAvailable时，才认为该job运行正常
                  initSchedulerName("volcano"). // 使用Volcano作为调度，用户可根据实际情况修改
                  initPolicies([]v1alpha1.LifecyclePolicy{{Event: "PodEvicted", Action: "RestartJob"}}).
                  initPlugins(map[string][]string{"ssh": {}, "env": {}, "svc": {}}). // 初始化调度插件
                  initMaxRetry(3).                                                   // maxRetry表示该job可以进行的最大重启次数
                  initQueue("default" ).                                              // queue表示该job所属的队列
                  addTask(v1alpha1.TaskSpec(newTaskSpec().
                     initTaskName("default-test").       // 初始化task名
                     initTaskReplicas(2).                // 初始化task副本数
                     initTaskAffinity("mindx-dls-test"). // 初始化任务反亲和性字段，输入值与任务名相同
                     initTaskLabels(map[string]string{
                        "app":                   "mindspore",   // 固定字段
                        "ring-controller.atlas": "ascend-{xxx}b", //  标识任务使用的芯片的产品类型
                     }).
                     initTaskNodeSelector(map[string]string{
                        "host-arch":        "huawei-x86",     // 可选字段，用户根据实际需求配置
                        "accelerator-type": "module-{xxx}b-8", // 用户根据服务器类型进行配置
                     }).
                     initTaskVolumes(). // 初始化挂载项
                     addTaskContainers(v1.Container(newContainer().
                        initContainerName("mindspore").                                             // 初始化容器名
                        initContainerImage("ms-arm:b120").                                          // 初始化镜像名
                        initContainerImagePullPolicy("IfNotPresent").                               // 初始化镜像拉取策略
                        initContainerLimits("huawei.com/Ascend910", "8").                           // 初始化任务资源
                        initContainerRequests("huawei.com/Ascend910", "8").                         // 初始化任务资源
                        initContainerVolumeMounts().                                                // 初始化容器挂载项
                        initContainerEnv("MindSpore").                                              // 初始化容器环境变量
                        initContainerCommand([]string{"/bin/bash", "-c", "bash train_start.sh ..."}))))) // 初始化容器启动命令，具体参数参考示例YAML
               return v1alpha1.Job(job)
            }
            
            type vcJob v1alpha1.Job
            type vcTask v1alpha1.TaskSpec
            type container v1.Container
            
            // 初始化任务名
            func (job *vcJob) initName(n string) *vcJob {
               job.Name = n
               return job
            }
            ...
            // 初始化task名
            func (task vcTask) initTaskName(tn string) vcTask {
               task.Name = tn
               return task
            }
            ...
            // 初始化容器环境变量
            func (ct container) initContainerEnv(framework string) container {
               ct.Env = []v1.EnvVar{
                  {
                     Name: "mindx-dls-test", // 任务名称
                     ValueFrom: &v1.EnvVarSource{
                        FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
                     },
                  },
                  {
                     Name: "XDL_IP", // 固定字段
                     ValueFrom: &v1.EnvVarSource{
                        FieldRef: &v1.ObjectFieldSelector{FieldPath: "status.hostIP"},
                     },
                  },
                  {
                     Name:  "framework",
                     Value: framework, // 使用的训练框架名称。支持MindSpore、PyTorch和Tensorflow
                  },
               }
               return ct
            }
            ```

4.  创建命名空间，以vcjob为例，示例如下。

    ```
    clientset.CoreV1().Namespaces().Create(context.TODO(), newNameSpace("vcjob"), metav1.CreateOptions{})
    ```

5.  （可选）如果通过文件配置资源信息，还需要创建RankTable的ConfigMap，示例如下。

    ```
    clientset.CoreV1().ConfigMaps(job.Namespace).Create(context.TODO(), newConfigMap("rings-config-"+job.Name), metav1.CreateOptions{})
    ```

6.  调用Create接口，下发训练任务。
    -   Ascend Job

        ```
        acjobClient.BatchV1().Jobs("default").Create(context.TODO(), &job, metav1.CreateOptions{})
        ```

    -   Volcano Job

        ```
        vcjobClient.BatchV1alpha1().Jobs("vcjob").Create(context.TODO(), &job, metav1.CreateOptions{})
        ```

7.  查看任务进程。调用Get接口查看job是否创建成功。
    -   Ascend Job

        ```
        acjobClient.BatchV1().Jobs("default").Get(context.TODO(), job.Name, metav1.GetOptions{})
        ```

    -   Volcano Job

        ```
        vcjobClient.BatchV1alpha1().Jobs("vcjob").Get(context.TODO(), job.Name, metav1.GetOptions{})
        ```

8.  删除任务。调用Delete接口删除任务。
    -   Ascend Job

        ```
        acjobClient.BatchV1().Jobs("default").Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
        ```

    -   Volcano Job

        ```
        vcjobClient.BatchV1alpha1().Jobs("vcjob").Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
        ```

**集成后使用<a name="section1027912153611"></a>**

1.  制作相应的镜像，可参考[制作镜像](../common_operations.md#制作镜像)章节进行操作。
2.  完成相应的脚本适配，可参考[脚本适配](#脚本适配)章节进行操作。
3.  创建任务。
4.  运行训练任务。可通过平台配置并创建训练任务，下发任务后查看结果。



## 整卡调度或静态vNPU调度（推理）<a name="ZH-CN_TOPIC_0000002511347095"></a>

### 使用前必读<a name="ZH-CN_TOPIC_0000002511427055"></a>

**前提条件<a name="section116017220425"></a>**

在命令行场景下使用整卡调度和静态vNPU调度特性，需要确保已经安装如下组件；若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。

-   调度器（Volcano或其他调度器）
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator
-   ClusterD
-   NodeD

**使用方式<a name="section91871616135119"></a>**

整卡调度或静态vNPU调度特性的使用方式如下：

-   通过命令行使用：安装集群调度组件，通过命令行使用整卡调度特性。
-   集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section577625973520"></a>**

-   资源监测可以和推理场景下的所有特性一起使用。
-   集群中同时跑多个推理任务，每个任务使用的特性可以不同，但不能同时存在使用静态vNPU的任务和使用动态vNPU的任务。
-   推理卡故障恢复特性可以搭配整卡调度特性一起使用，开启整卡故障恢复特性只需要将Ascend Device Plugin的启动参数“-hotReset”取值设置为“0”或“2”（默认为“-1”，不支持故障恢复功能）。
-   整卡调度支持下发单副本数或者多副本数的单机任务，每个副本独立工作，只支持推理服务器（插Atlas 300I Duo 推理卡）和Atlas 800I A2 推理服务器、A200I A2 Box 异构组件部署acjob类型的分布式任务。
-   静态vNPU调度只支持下发单副本数的单机任务，不支持分布式任务。
-   静态vNPU调度特性需要搭配算力虚拟化特性一起使用，关于静态虚拟化的相关说明和操作请参见[静态虚拟化](./virtual_instance.md#静态虚拟化)章节。

**支持的产品形态<a name="section169961844182917"></a>**

-   支持以下产品使用**整卡调度**。
    -   推理服务器（插Atlas 300I 推理卡）
    -   Atlas 推理系列产品
    -   Atlas 800I A2 推理服务器
    -   A200I A2 Box 异构组件
    -   Atlas 800I A3 超节点服务器

-   支持以下产品使用**静态vNPU调度**。

    Atlas 推理系列产品

**使用流程<a name="section246711128536"></a>**

通过命令行使用整卡调度或静态vNPU调度特性的流程可以参见[图1](#fig242524985412)。

通过命令行使用Volcano和其他调度器的使用流程一致，主要区别在使用其他调度器准备任务YAML需要参考[通过命令行使用（其他调度器）](#通过命令行使用其他调度器-1)章节创建任务YAML。使用其他调度器的其余操作和使用Volcano一致，可以参考[通过命令行使用（Volcano）](#通过命令行使用volcano-1)进行操作。

**图 1**  使用流程<a name="fig242524985412"></a>  
![](../../figures/scheduling/使用流程.png "使用流程")


### 实现原理<a name="ZH-CN_TOPIC_0000002479227174"></a>

根据推理任务类型的不同，特性的原理图略有差异。静态vNPU调度需要使用npu-smi工具提前创建好需要的vNPU。

**acjob任务<a name="section9971431567"></a>**

acjob任务原理图如[图1](实现原理-3.md#fig5188536014)所示。

**图 1**  acjob任务调度原理图<a name="fig36890512379"></a>  
![](../../figures/scheduling/acjob任务调度原理图-0.png "acjob任务调度原理图-0")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到节点对象（Node）中。
    -   Ascend Device Plugin定期上报芯片拓扑信息。

        上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）、已使用的芯片数量（allocated）和芯片的基础信息（device ip和super\_device\_ip）上报到Node中，用于整卡调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3.  用户通过kubectl或者其他深度学习平台下发acjob任务。
4.  Ascend Operator为任务创建相应的PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  Ascend Operator为任务创建相应的Pod，并在容器中注入集合通信所需环境变量。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。整卡调度写入整卡信息。
7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin或volcano-scheduler在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源。
8.  Ascend Operator读取Pod的annotation信息，将相关信息写入hccl.json。
9.  容器读取环境变量或者hccl.json信息，建立通信渠道，开始执行推理任务。

    >[!NOTE] 说明 
    >Ascend Operator当前仅支持为PyTorch任务生成hccl.json。

**vcjob任务<a name="section428321965913"></a>**

vcjob任务的原理图如[图2](#fig8231124765)所示。

**图 2**  vcjob任务调度原理图<a name="fig8231124765"></a>  
![](../../figures/scheduling/vcjob任务调度原理图-1.png "vcjob任务调度原理图-1")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到Node（节点对象）中。
    -   Ascend Device Plugin定期上报芯片拓扑信息。
        -   上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）和已使用的芯片数量（allocated）上报到Node中，用于整卡调度。
        -   上报vNPU相关信息到Node中，用于静态vNPU调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3.  用户通过kubectl或者其他深度学习平台下发vcjob任务。
4.  volcano-controller为任务创建相应PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  当集群资源满足任务要求时，volcano-controller创建任务Pod。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源。

**deploy任务<a name="section148711820709"></a>**

deploy任务原理图如[图3](#fig178781320593)所示。

**图 3**  deploy任务调度原理图<a name="fig178781320593"></a>  
![](../../figures/scheduling/deploy任务调度原理图-2.png "deploy任务调度原理图-2")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到Node（节点对象）中。
    -   Ascend Device Plugin定期上报芯片拓扑信息。
        -   上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）和已使用的芯片数量（allocated）上报到Node中，用于整卡调度。
        -   上报vNPU相关信息到Node中，用于静态vNPU调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3.  用户通过kubectl或者其他深度学习平台下发deploy任务。
4.  kube-controller为任务创建相应Pod。
5.  volcano-controller创建任务PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源。


### 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002511427059"></a>

#### 制作镜像<a name="ZH-CN_TOPIC_0000002479227156"></a>

**获取推理镜像<a name="zh-cn_topic_0000001558675566_section971616541059"></a>**

可选择以下方式中的一种来获取推理镜像。

-   推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据用户的系统架构（ARM或者x86\_64）下载推理基础镜像（如：[ascend-infer](https://www.hiascend.com/developer/ascendhub/detail/e02f286eef0847c2be426f370e0c2596)、[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)）。

    请注意，21.0.4版本之后推理基础镜像默认用户为非root用户，需要在下载基础镜像后对其进行修改，将默认用户修改为root。

    >[!NOTE] 说明 
    >基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。

-   （可选）如果用户需要更个性化的推理环境，可基于已下载的推理基础镜像，再[使用Dockerfile对其进行修改](../common_operations.md#使用dockerfile构建容器镜像tensorflow)。

    完成定制化修改后，用户可以给推理镜像重命名，以便管理和使用。

**加固镜像<a name="zh-cn_topic_0000001558675566_section1294572963118"></a>**

下载或者制作的推理基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../references.md#容器安全加固)章节进行操作。


#### 脚本适配<a name="ZH-CN_TOPIC_0000002479227176"></a>

本章节以昇腾镜像仓库中推理镜像为例，为用户介绍下发推理任务的操作流程。该镜像已经包含了推理示例脚本，实际推理场景需要用户自行准备推理脚本。在拉取镜像前，需要确保当前环境的网络代理已经配置完成，且能成功访问昇腾镜像仓库。

**从昇腾镜像仓库获取示例脚本<a name="section8181015175911"></a>**

1.  确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2.  在左侧导航栏选择推理镜像，然后选择[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)镜像，获取推理示例脚本。

    >[!NOTE] 说明 
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。


#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479387148"></a>

>[!NOTE] 说明 
>如果用户不使用Ascend Docker Runtime组件，Ascend Device Plugin只会帮助用户挂载“/dev“目录下的设备。其他目录（如“/usr“）用户需要自行修改YAML文件，挂载对应的驱动目录和文件。容器内挂载路径和宿主机路径保持一致。
>因为Atlas 200I SoC A1 核心板场景不支持Ascend Docker Runtime，用户也无需修改YAML文件。

**操作步骤<a name="zh-cn_topic_0000001609074213_section14665181617334"></a>**

1.  下载YAML文件。

    **表 1**  任务类型与硬件型号对应YAML文件

    <a name="zh-cn_topic_0000001609074213_table15169151021912"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001609074213_row16169201019192"><th class="cellrowborder" valign="top" width="18.48%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001609074213_p4169191017192"><a name="zh-cn_topic_0000001609074213_p4169191017192"></a><a name="zh-cn_topic_0000001609074213_p4169191017192"></a>任务类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="26.479999999999997%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001609074213_p20181111517147"><a name="zh-cn_topic_0000001609074213_p20181111517147"></a><a name="zh-cn_topic_0000001609074213_p20181111517147"></a>硬件型号</p>
    </th>
    <th class="cellrowborder" valign="top" width="42.59%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001609074213_p181811156149"><a name="zh-cn_topic_0000001609074213_p181811156149"></a><a name="zh-cn_topic_0000001609074213_p181811156149"></a>YAML名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="12.45%" id="mcps1.2.5.1.4"><p id="p1693015221828"><a name="p1693015221828"></a><a name="p1693015221828"></a>获取链接</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001609074213_row2169191091919"><td class="cellrowborder" rowspan="2" valign="top" width="18.48%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p6169510191913"><a name="zh-cn_topic_0000001609074213_p6169510191913"></a><a name="zh-cn_topic_0000001609074213_p6169510191913"></a><span id="zh-cn_topic_0000001609074213_ph183921109162"><a name="zh-cn_topic_0000001609074213_ph183921109162"></a><a name="zh-cn_topic_0000001609074213_ph183921109162"></a>Volcano</span>调度的Deployment任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="26.479999999999997%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p8853185832112"><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><span id="zh-cn_topic_0000001609074213_ph238151934915"><a name="zh-cn_topic_0000001609074213_ph238151934915"></a><a name="zh-cn_topic_0000001609074213_ph238151934915"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.59%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001609074213_p1116971091915"><a name="zh-cn_topic_0000001609074213_p1116971091915"></a><a name="zh-cn_topic_0000001609074213_p1116971091915"></a>infer-deploy-310p-1usoc.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="12.45%" headers="mcps1.2.5.1.4 "><p id="p784716567219"><a name="p784716567219"></a><a name="p784716567219"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-deploy-310p-1usoc.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row17169201091917"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p14853125832110"><a name="zh-cn_topic_0000001609074213_p14853125832110"></a><a name="zh-cn_topic_0000001609074213_p14853125832110"></a>其他类型推理节点</p>
    <p id="p1144215219166"><a name="p1144215219166"></a><a name="p1144215219166"></a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p51692100191"><a name="zh-cn_topic_0000001609074213_p51692100191"></a><a name="zh-cn_topic_0000001609074213_p51692100191"></a>infer-deploy.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p74352718168"><a name="p74352718168"></a><a name="p74352718168"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-deploy.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row114428221610"><td class="cellrowborder" valign="top" width="18.48%" headers="mcps1.2.5.1.1 "><p id="p9442102131620"><a name="p9442102131620"></a><a name="p9442102131620"></a>Volcano Job任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="26.479999999999997%" headers="mcps1.2.5.1.2 "><p id="p367438101714"><a name="p367438101714"></a><a name="p367438101714"></a><span id="ph313817549316"><a name="ph313817549316"></a><a name="ph313817549316"></a>Atlas 800I A2 推理服务器</span></p>
    <p id="p20458181019389"><a name="p20458181019389"></a><a name="p20458181019389"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
    <p id="p1792637151014"><a name="p1792637151014"></a><a name="p1792637151014"></a><span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.59%" headers="mcps1.2.5.1.3 "><p id="p8442112171619"><a name="p8442112171619"></a><a name="p8442112171619"></a>infer-vcjob-910.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="12.45%" headers="mcps1.2.5.1.4 "><p id="p15442424164"><a name="p15442424164"></a><a name="p15442424164"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-vcjob-910.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row16861151313547"><td class="cellrowborder" rowspan="2" valign="top" width="18.48%" headers="mcps1.2.5.1.1 "><p id="p6861171325411"><a name="p6861171325411"></a><a name="p6861171325411"></a>Ascend Job任务</p>
    <p id="p12446175211817"><a name="p12446175211817"></a><a name="p12446175211817"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="26.479999999999997%" headers="mcps1.2.5.1.2 "><p id="p1328416110919"><a name="p1328416110919"></a><a name="p1328416110919"></a>推理服务器（插<span id="ph93658382564"><a name="ph93658382564"></a><a name="ph93658382564"></a>Atlas 300I Duo 推理卡</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.59%" headers="mcps1.2.5.1.3 "><p id="p10861813135419"><a name="p10861813135419"></a><a name="p10861813135419"></a>pytorch_acjob_infer_310p_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="12.45%" headers="mcps1.2.5.1.4 "><p id="p1986116136544"><a name="p1986116136544"></a><a name="p1986116136544"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/pytorch_acjob_infer_310p_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row18446115212811"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1611216221297"><a name="p1611216221297"></a><a name="p1611216221297"></a><span id="ph10342125017508"><a name="ph10342125017508"></a><a name="ph10342125017508"></a>Atlas 800I A2 推理服务器</span></p>
    <p id="p1877419343388"><a name="p1877419343388"></a><a name="p1877419343388"></a><span id="ph1311636133812"><a name="ph1311636133812"></a><a name="ph1311636133812"></a>A200I A2 Box 异构组件</span></p>
    <p id="p1368016125100"><a name="p1368016125100"></a><a name="p1368016125100"></a><span id="ph17176513111020"><a name="ph17176513111020"></a><a name="ph17176513111020"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p4446185212815"><a name="p4446185212815"></a><a name="p4446185212815"></a>pytorch_multinodes_acjob_infer_<em id="i232224205019"><a name="i232224205019"></a><a name="i232224205019"></a>{</em><em id="i133214249507"><a name="i133214249507"></a><a name="i133214249507"></a>xxx}</em>b_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p962512301913"><a name="p962512301913"></a><a name="p962512301913"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/pytorch_multinodes_acjob_infer_910b_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    </tbody>
    </table>

2.  将YAML文件上传至管理节点任意目录，并参考[表2](#zh-cn_topic_0000001609074213_table5589101114528)修改示例YAML。

    **表 2**  YAML文件参数说明

    <a name="zh-cn_topic_0000001609074213_table5589101114528"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001609074213_row125891211155216"><th class="cellrowborder" valign="top" width="26.12261226122612%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074213_p13658124194513"><a name="zh-cn_topic_0000001609074213_p13658124194513"></a><a name="zh-cn_topic_0000001609074213_p13658124194513"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="36.16361636163616%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074213_p4658152420459"><a name="zh-cn_topic_0000001609074213_p4658152420459"></a><a name="zh-cn_topic_0000001609074213_p4658152420459"></a>取值</p>
    </th>
    <th class="cellrowborder" valign="top" width="37.71377137713771%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074213_p8302202619484"><a name="zh-cn_topic_0000001609074213_p8302202619484"></a><a name="zh-cn_topic_0000001609074213_p8302202619484"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001609074213_row145900112522"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p65901011115213"><a name="zh-cn_topic_0000001609074213_p65901011115213"></a><a name="zh-cn_topic_0000001609074213_p65901011115213"></a>image</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p105901311195216"><a name="zh-cn_topic_0000001609074213_p105901311195216"></a><a name="zh-cn_topic_0000001609074213_p105901311195216"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p5590191185217"><a name="zh-cn_topic_0000001609074213_p5590191185217"></a><a name="zh-cn_topic_0000001609074213_p5590191185217"></a>推理镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row14141145104416"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p4393104684410"><a name="zh-cn_topic_0000001609074213_p4393104684410"></a><a name="zh-cn_topic_0000001609074213_p4393104684410"></a>replicas</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p83938460446"><a name="zh-cn_topic_0000001609074213_p83938460446"></a><a name="zh-cn_topic_0000001609074213_p83938460446"></a>整数</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p11393154664410"><a name="zh-cn_topic_0000001609074213_p11393154664410"></a><a name="zh-cn_topic_0000001609074213_p11393154664410"></a>运行的任务副本数量。通常情况一般为1。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row2059051145219"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p8590101118522"><a name="zh-cn_topic_0000001609074213_p8590101118522"></a><a name="zh-cn_topic_0000001609074213_p8590101118522"></a>requests</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p4120153112010"><a name="p4120153112010"></a><a name="p4120153112010"></a><strong id="b91271758152414"><a name="b91271758152414"></a><a name="b91271758152414"></a>整卡调度：</strong></p>
    <a name="zh-cn_topic_0000001609074213_ul1180139155411"></a><a name="zh-cn_topic_0000001609074213_ul1180139155411"></a><ul id="zh-cn_topic_0000001609074213_ul1180139155411"><li>推理服务器（插<span id="ph163696166292"><a name="ph163696166292"></a><a name="ph163696166292"></a>Atlas 300I 推理卡</span>）：<p id="zh-cn_topic_0000001609074213_p364765019017"><a name="zh-cn_topic_0000001609074213_p364765019017"></a><a name="zh-cn_topic_0000001609074213_p364765019017"></a>huawei.com/Ascend310: <em id="zh-cn_topic_0000001609074213_i126472503016"><a name="zh-cn_topic_0000001609074213_i126472503016"></a><a name="zh-cn_topic_0000001609074213_i126472503016"></a>芯片数量</em></p>
    </li></ul>
    <a name="zh-cn_topic_0000001609074213_ul8938201113543"></a><a name="zh-cn_topic_0000001609074213_ul8938201113543"></a><ul id="zh-cn_topic_0000001609074213_ul8938201113543"><li><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>非混插模式：<p id="zh-cn_topic_0000001609074213_p464718509014"><a name="zh-cn_topic_0000001609074213_p464718509014"></a><a name="zh-cn_topic_0000001609074213_p464718509014"></a>huawei.com/Ascend310P: <em id="zh-cn_topic_0000001609074213_i06475509019"><a name="zh-cn_topic_0000001609074213_i06475509019"></a><a name="zh-cn_topic_0000001609074213_i06475509019"></a>芯片数量</em></p>
    </li></ul>
    <a name="zh-cn_topic_0000001609074213_ul13727161475413"></a><a name="zh-cn_topic_0000001609074213_ul13727161475413"></a><ul id="zh-cn_topic_0000001609074213_ul13727161475413"><li><span id="ph1883472303917"><a name="ph1883472303917"></a><a name="ph1883472303917"></a>Atlas 推理系列产品</span>混插模式：<a name="zh-cn_topic_0000001609074213_ul8401842105312"></a><a name="zh-cn_topic_0000001609074213_ul8401842105312"></a><ul id="zh-cn_topic_0000001609074213_ul8401842105312"><li>huawei.com/Ascend310P-V: <em id="zh-cn_topic_0000001609074213_i16471550409"><a name="zh-cn_topic_0000001609074213_i16471550409"></a><a name="zh-cn_topic_0000001609074213_i16471550409"></a>芯片数量</em></li><li>huawei.com/Ascend310P-VPro: <em id="zh-cn_topic_0000001609074213_i146476501013"><a name="zh-cn_topic_0000001609074213_i146476501013"></a><a name="zh-cn_topic_0000001609074213_i146476501013"></a>芯片数量</em></li><li>huawei.com/Ascend310P-IPro: <em id="zh-cn_topic_0000001609074213_i06476501014"><a name="zh-cn_topic_0000001609074213_i06476501014"></a><a name="zh-cn_topic_0000001609074213_i06476501014"></a>芯片数量</em></li></ul>
    </li><li><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>、<span id="ph1222716463422"><a name="ph1222716463422"></a><a name="ph1222716463422"></a>A200I A2 Box 异构组件</span>、<span id="ph8551122472512"><a name="ph8551122472512"></a><a name="ph8551122472512"></a>Atlas 800I A3 超节点服务器</span>：huawei.com/Ascend910：<em id="i68164310406"><a name="i68164310406"></a><a name="i68164310406"></a>芯片数量</em></li></ul>
    <p id="p11520941142018"><a name="p11520941142018"></a><a name="p11520941142018"></a><strong id="b1023211223510"><a name="b1023211223510"></a><a name="b1023211223510"></a>静态vNPU调度：</strong>取值为1。只能使用一个NPU下的vNPU。</p>
    <p id="p99844293552"><a name="p99844293552"></a><a name="p99844293552"></a><span id="ph876864313911"><a name="ph876864313911"></a><a name="ph876864313911"></a>Atlas 推理系列产品</span>非混插模式：huawei.com/Ascend310P-<em id="i1745936185214"><a name="i1745936185214"></a><a name="i1745936185214"></a>Y</em>: 1</p>
    <a name="ul1257013663114"></a><a name="ul1257013663114"></a>
    <p id="p444515595295"><a name="p444515595295"></a><a name="p444515595295"></a>如<em id="i154056256459"><a name="i154056256459"></a><a name="i154056256459"></a>huawei.com/Ascend310P-4c.3cpu</em>: 1</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p11590211155219"><a name="zh-cn_topic_0000001609074213_p11590211155219"></a><a name="zh-cn_topic_0000001609074213_p11590211155219"></a>请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。requests和limits下，芯片的名字和数量需保持一致。</p>
    <div class="note" id="note1648201912419"><a name="note1648201912419"></a><a name="note1648201912419"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul10782193418818"></a><a name="ul10782193418818"></a><ul id="ul10782193418818"><li>仅<span id="ph1038285416813"><a name="ph1038285416813"></a><a name="ph1038285416813"></a>Atlas 推理系列产品</span>非混插模式支持静态vNPU调度。</li><li>推理服务器（插<span id="ph1990710374611"><a name="ph1990710374611"></a><a name="ph1990710374611"></a>Atlas 300I 推理卡</span>）和<span id="ph629210161695"><a name="ph629210161695"></a><a name="ph629210161695"></a>Atlas 推理系列产品</span>混插模式不支持静态vNPU调度。</li><li><strong id="b179331118122318"><a name="b179331118122318"></a><a name="b179331118122318"></a><em id="i14933131862318"><a name="i14933131862318"></a><a name="i14933131862318"></a>Y</em></strong>取值可参考<a href="./virtual_instance.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与虚拟设备类型关系表的对应产品的“vNPU类型”列。<p id="p208621211164518"><a name="p208621211164518"></a><a name="p208621211164518"></a>以vNPU类型<em id="i412654718449"><a name="i412654718449"></a><a name="i412654718449"></a>Ascend310P-4c.3cpu</em>为例，<strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong>取值为4c.3cpu，不包括前面的Ascend310P。</p>
    </li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row114301545157"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p16864941511"><a name="zh-cn_topic_0000001609074213_p16864941511"></a><a name="zh-cn_topic_0000001609074213_p16864941511"></a>limits</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p203039915181"><a name="p203039915181"></a><a name="p203039915181"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
    <p id="p4739213121713"><a name="p4739213121713"></a><a name="p4739213121713"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row105901411135220"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p1590711195215"><a name="zh-cn_topic_0000001609074213_p1590711195215"></a><a name="zh-cn_topic_0000001609074213_p1590711195215"></a>（可选）host-arch</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p1650105613241"><a name="zh-cn_topic_0000001609074213_p1650105613241"></a><a name="zh-cn_topic_0000001609074213_p1650105613241"></a><span id="zh-cn_topic_0000001609074213_ph16676195493717"><a name="zh-cn_topic_0000001609074213_ph16676195493717"></a><a name="zh-cn_topic_0000001609074213_ph16676195493717"></a>ARM</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
    <p id="zh-cn_topic_0000001609074213_p0658124184512"><a name="zh-cn_topic_0000001609074213_p0658124184512"></a><a name="zh-cn_topic_0000001609074213_p0658124184512"></a><span id="zh-cn_topic_0000001609074213_ph1274682034217"><a name="zh-cn_topic_0000001609074213_ph1274682034217"></a><a name="zh-cn_topic_0000001609074213_ph1274682034217"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p8590711115212"><a name="zh-cn_topic_0000001609074213_p8590711115212"></a><a name="zh-cn_topic_0000001609074213_p8590711115212"></a>需要运行推理任务的节点架构，请根据实际修改。<span id="zh-cn_topic_0000001609074213_ph183338272492"><a name="zh-cn_topic_0000001609074213_ph183338272492"></a><a name="zh-cn_topic_0000001609074213_ph183338272492"></a>Atlas 200I SoC A1 核心板</span>节点仅支持huawei-arm。</p>
    </td>
    </tr>
    <tr id="row4336163153417"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p19119935347"><a name="p19119935347"></a><a name="p19119935347"></a>huawei.com/recover_policy_path</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p131191634340"><a name="p131191634340"></a><a name="p131191634340"></a>pod：只支持Pod级重调度，不升级为Job级别。</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p21194343415"><a name="p21194343415"></a><a name="p21194343415"></a>任务重调度策略。</p>
    </td>
    </tr>
    <tr id="row13531183303414"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p123665543412"><a name="p123665543412"></a><a name="p123665543412"></a>huawei.com/schedule_minAvailable</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p636605193410"><a name="p636605193410"></a><a name="p636605193410"></a>整数</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p436675193411"><a name="p436675193411"></a><a name="p436675193411"></a>任务能够调度的最小副本数。</p>
    </td>
    </tr>
    <tr id="row12652195216521"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
    <div class="note" id="note1230363125010"><a name="note1230363125010"></a><a name="note1230363125010"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000002511347099_p1767434372512"><a name="zh-cn_topic_0000002511347099_p1767434372512"></a><a name="zh-cn_topic_0000002511347099_p1767434372512"></a>仅支持在<span id="ph996833614580"><a name="ph996833614580"></a><a name="ph996833614580"></a><term id="zh-cn_topic_0000001094307702_term99602034117"><a name="zh-cn_topic_0000001094307702_term99602034117"></a><a name="zh-cn_topic_0000001094307702_term99602034117"></a>Atlas A2 推理系列产品</term></span>和<span id="ph791742714211"><a name="ph791742714211"></a><a name="ph791742714211"></a><term id="zh-cn_topic_0000001519959665_term176419491615"><a name="zh-cn_topic_0000001519959665_term176419491615"></a><a name="zh-cn_topic_0000001519959665_term176419491615"></a>Atlas A3 推理系列产品</term></span>中使用该字段。</p>
    </div></div>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row144522615563"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p944513266569"><a name="zh-cn_topic_0000001609074213_p944513266569"></a><a name="zh-cn_topic_0000001609074213_p944513266569"></a>servertype</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p544515265567"><a name="zh-cn_topic_0000001609074213_p544515265567"></a><a name="zh-cn_topic_0000001609074213_p544515265567"></a>soc</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p202093166576"><a name="zh-cn_topic_0000001609074213_p202093166576"></a><a name="zh-cn_topic_0000001609074213_p202093166576"></a>服务器类型。</p>
    <a name="zh-cn_topic_0000001609074213_ul87677178911"></a><a name="zh-cn_topic_0000001609074213_ul87677178911"></a><ul id="zh-cn_topic_0000001609074213_ul87677178911"><li>调度到<span id="zh-cn_topic_0000001609074213_ph126801133164916"><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a>Atlas 200I SoC A1 核心板</span>节点上，必须要加上此配置，并参考<span class="filepath" id="zh-cn_topic_0000001609074213_filepath127811055718"><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a>“infer-310p-1usoc.yaml”</span>文件进行目录挂载。</li><li>其他类型节点不需要此参数。</li></ul>
    </td>
    </tr>
    <tr id="row18924102118319"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/AscendXXX']</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
    <div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>该参数只支持使用<span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span>调度器的整卡调度特性。使用静态vNPU调度、动态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row157451231699"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p495717111797"><a name="p495717111797"></a><a name="p495717111797"></a>以下参数仅支持推理服务器（插<span id="ph18312482615"><a name="ph18312482615"></a><a name="ph18312482615"></a>Atlas 300I 推理卡</span>）使用：</p>
    </td>
    </tr>
    <tr id="row33457612918"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p137845934610"><a name="p137845934610"></a><a name="p137845934610"></a>npu-310-strategy</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul1967514291118"></a><a name="ul1967514291118"></a><ul id="ul1967514291118"><li>card：按推理卡调度，request请求的<span id="ph1978781173013"><a name="ph1978781173013"></a><a name="ph1978781173013"></a>昇腾AI处理器</span>个数不超过4，使用同一张<span id="ph933971152917"><a name="ph933971152917"></a><a name="ph933971152917"></a>Atlas 300I 推理卡</span>上的<span id="ph77331623132919"><a name="ph77331623132919"></a><a name="ph77331623132919"></a>昇腾AI处理器</span>。</li><li>chip：按<span id="ph14705121219305"><a name="ph14705121219305"></a><a name="ph14705121219305"></a>昇腾AI处理器</span>调度，请求的芯片个数不超过单个节点的最大值。</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p2799246194816"><a name="p2799246194816"></a><a name="p2799246194816"></a>-</p>
    </td>
    </tr>
    <tr id="row319911311103"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p146931353154819"><a name="p146931353154819"></a><a name="p146931353154819"></a>schedulerName</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p18693105384810"><a name="p18693105384810"></a><a name="p18693105384810"></a>volcano</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p16693165318488"><a name="p16693165318488"></a><a name="p16693165318488"></a>如果切换调度器，需要将之前调度的任务都释放。</p>
    </td>
    </tr>
    <tr id="row1885310221269"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p12531944132618"><a name="p12531944132618"></a><a name="p12531944132618"></a>以下参数仅支持推理服务器（插<span id="ph1653016814441"><a name="ph1653016814441"></a><a name="ph1653016814441"></a>Atlas 300I Duo 推理卡</span>）使用：</p>
    </td>
    </tr>
    <tr id="row4640132422614"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p4504336171112"><a name="p4504336171112"></a><a name="p4504336171112"></a>duo</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul145791915920"></a><a name="ul145791915920"></a><ul id="ul145791915920"><li>true：使用<span id="ph19427048143715"><a name="ph19427048143715"></a><a name="ph19427048143715"></a>Atlas 300I Duo 推理卡</span>。</li><li>false：不使用<span id="ph1069395411377"><a name="ph1069395411377"></a><a name="ph1069395411377"></a>Atlas 300I Duo 推理卡</span>。</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p76419248265"><a name="p76419248265"></a><a name="p76419248265"></a>推理卡类型。</p>
    </td>
    </tr>
    <tr id="row6543122611264"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p1050423615113"><a name="p1050423615113"></a><a name="p1050423615113"></a>npu-310-strategy</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul11389536112712"></a><a name="ul11389536112712"></a><ul id="ul11389536112712"><li>card：按推理卡调度，request请求的<span id="ph78655288388"><a name="ph78655288388"></a><a name="ph78655288388"></a>昇腾AI处理器</span>个数不超过2，使用同一张<span id="ph12172145383816"><a name="ph12172145383816"></a><a name="ph12172145383816"></a>Atlas 300I Duo 推理卡</span>上的<span id="ph314316597299"><a name="ph314316597299"></a><a name="ph314316597299"></a>昇腾AI处理器</span>。</li><li>chip：按<span id="ph3342123919383"><a name="ph3342123919383"></a><a name="ph3342123919383"></a>昇腾AI处理器</span>调度，请求的<span id="ph1810712517398"><a name="ph1810712517398"></a><a name="ph1810712517398"></a>昇腾AI处理器</span>个数不超过单个节点的最大值。</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p354472662613"><a name="p354472662613"></a><a name="p354472662613"></a>-</p>
    </td>
    </tr>
    <tr id="row183293410266"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p750513618118"><a name="p750513618118"></a><a name="p750513618118"></a>distributed</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul869614532313"></a><a name="ul869614532313"></a><ul id="ul869614532313"><li>true：使用分布式推理。使用chip模式时，必须将任务调度到整张<span id="ph11938141213910"><a name="ph11938141213910"></a><a name="ph11938141213910"></a>Atlas 300I Duo 推理卡</span>。若任务需要的<span id="ph941811717011"><a name="ph941811717011"></a><a name="ph941811717011"></a>昇腾AI处理器</span>数量为单数时，使用单个<span id="ph1827593613120"><a name="ph1827593613120"></a><a name="ph1827593613120"></a>昇腾AI处理器</span>的部分，将优先调度到剩余<span id="ph6991738103714"><a name="ph6991738103714"></a><a name="ph6991738103714"></a>昇腾AI处理器</span>数量为1的<span id="ph19399122710392"><a name="ph19399122710392"></a><a name="ph19399122710392"></a>Atlas 300I Duo 推理卡</span>上。</li><li>false：使用非分布式推理。使用chip模式时，请求的<span id="ph281018043220"><a name="ph281018043220"></a><a name="ph281018043220"></a>昇腾AI处理器</span>个数不超过单个节点的最大值。<div class="note" id="note595619820324"><a name="note595619820324"></a><a name="note595619820324"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul1857014516418"></a><a name="ul1857014516418"></a><ul id="ul1857014516418"><li>无论是否为分布式推理，card模式的调度策略不变。</li><li>当distributed为true时，只支持单机多卡；当distributed为false时，只支持多机多卡。</li><li>当distributed为true时，不支持Deployment任务。</li></ul>
    </div></div>
    </li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p10832534102618"><a name="p10832534102618"></a><a name="p10832534102618"></a>是否使用分布式推理。</p>
    </td>
    </tr>
    <tr id="row672524614316"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p13217282100"><a name="p13217282100"></a><a name="p13217282100"></a>以下参数仅支持<span id="ph182182818109"><a name="ph182182818109"></a><a name="ph182182818109"></a>Atlas 800I A2 推理服务器</span>、<span id="ph5359173454115"><a name="ph5359173454115"></a><a name="ph5359173454115"></a>A200I A2 Box 异构组件</span>、<span id="ph115512198285"><a name="ph115512198285"></a><a name="ph115512198285"></a>Atlas 800I A3 超节点服务器</span>使用：</p>
    </td>
    </tr>
    <tr id="row784311116441"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p284311118441"><a name="p284311118441"></a><a name="p284311118441"></a>nodeSelector</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p11843121164417"><a name="p11843121164417"></a><a name="p11843121164417"></a>module-<span id="ph81841512631"><a name="ph81841512631"></a><a name="ph81841512631"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b-8</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p1184371164415"><a name="p1184371164415"></a><a name="p1184371164415"></a>运行推理任务的节点类型。</p>
    </td>
    </tr>
    <tr id="row3259104216122"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p146661518161318"><a name="p146661518161318"></a><a name="p146661518161318"></a>以下参数仅acjob任务使用：</p>
    </td>
    </tr>
    <tr id="row13852155214127"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p6976153532710"><a name="p6976153532710"></a><a name="p6976153532710"></a>ring-controller.atlas</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul102731315166"></a><a name="ul102731315166"></a><ul id="ul102731315166"><li><span id="ph5149330124"><a name="ph5149330124"></a><a name="ph5149330124"></a>Atlas 800I A2 推理服务器</span>、<span id="ph01041750104112"><a name="ph01041750104112"></a><a name="ph01041750104112"></a>A200I A2 Box 异构组件</span>：ascend-<span id="ph11976935122715"><a name="ph11976935122715"></a><a name="ph11976935122715"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b</li><li>推理服务器（插<span id="ph2081684415121"><a name="ph2081684415121"></a><a name="ph2081684415121"></a>Atlas 300I Duo 推理卡</span>）：ascend-310P</li></ul>
    <p id="p1618162514154"><a name="p1618162514154"></a><a name="p1618162514154"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p139761835102718"><a name="p139761835102718"></a><a name="p139761835102718"></a>芯片类型。</p>
    </td>
    </tr>
    <tr id="row242655413123"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p1397753512272"><a name="p1397753512272"></a><a name="p1397753512272"></a>schedulerName</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p5977163532714"><a name="p5977163532714"></a><a name="p5977163532714"></a>默认值为<span class="parmvalue" id="parmvalue99772035172710"><a name="parmvalue99772035172710"></a><a name="parmvalue99772035172710"></a>“volcano”</span>，用户需根据自身情况填写</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p1397873502710"><a name="p1397873502710"></a><a name="p1397873502710"></a><span id="ph8978183512271"><a name="ph8978183512271"></a><a name="ph8978183512271"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
    </td>
    </tr>
    <tr id="row1318125611122"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p0978123511271"><a name="p0978123511271"></a><a name="p0978123511271"></a>minAvailable</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p2978113514276"><a name="p2978113514276"></a><a name="p2978113514276"></a>默认值为任务总副本数</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p897873582712"><a name="p897873582712"></a><a name="p897873582712"></a><span id="ph129781235182715"><a name="ph129781235182715"></a><a name="ph129781235182715"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="ph197814351275"><a name="ph197814351275"></a><a name="ph197814351275"></a>Volcano</span>时，任务运行总副本数。</p>
    </td>
    </tr>
    <tr id="row1352395716121"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p16978193511276"><a name="p16978193511276"></a><a name="p16978193511276"></a>queue</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p10978123512277"><a name="p10978123512277"></a><a name="p10978123512277"></a>默认值为<span class="parmvalue" id="parmvalue199781835112716"><a name="parmvalue199781835112716"></a><a name="parmvalue199781835112716"></a>“default”</span>，用户需根据自身情况填写</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p1497812357278"><a name="p1497812357278"></a><a name="p1497812357278"></a><span id="ph0978173562717"><a name="ph0978173562717"></a><a name="ph0978173562717"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="ph797816353270"><a name="ph797816353270"></a><a name="ph797816353270"></a>Volcano</span>时，任务所属队列。</p>
    </td>
    </tr>
    <tr id="row4212175991217"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p8979163582710"><a name="p8979163582710"></a><a name="p8979163582710"></a>（可选）successPolicy</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul897943582712"></a><a name="ul897943582712"></a><ul id="ul897943582712"><li>默认值为空，若用户不填写该参数，则默认取空值。</li><li>AllWorkers</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p11979535122719"><a name="p11979535122719"></a><a name="p11979535122719"></a>表明任务成功的前提。空值代表只需要一个<span id="ph149798358273"><a name="ph149798358273"></a><a name="ph149798358273"></a>Pod</span>成功，整个任务判定为成功。取值为<span class="parmvalue" id="parmvalue1797983518271"><a name="parmvalue1797983518271"></a><a name="parmvalue1797983518271"></a>“AllWorkers”</span>表示所有<span id="ph8979163522716"><a name="ph8979163522716"></a><a name="ph8979163522716"></a>Pod</span>都成功，任务才判定为成功。</p>
    </td>
    </tr>
    <tr id="row892973121315"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p7979123522712"><a name="p7979123522712"></a><a name="p7979123522712"></a>container.name</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p39791535192717"><a name="p39791535192717"></a><a name="p39791535192717"></a>ascend</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p3979235182716"><a name="p3979235182716"></a><a name="p3979235182716"></a>容器的名称必须是<span class="parmvalue" id="parmvalue12979735192712"><a name="parmvalue12979735192712"></a><a name="parmvalue12979735192712"></a>“ascend”</span>。</p>
    </td>
    </tr>
    <tr id="row172121265137"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p4979143514276"><a name="p4979143514276"></a><a name="p4979143514276"></a>（可选）ports</p>
    </td>
    <td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><p id="p1297916359270"><a name="p1297916359270"></a><a name="p1297916359270"></a>若用户未进行设置，系统默认填写以下参数：</p>
    <a name="ul69804359278"></a><a name="ul69804359278"></a><ul id="ul69804359278"><li>name: ascendjob-port</li><li>containerPort: 2222</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p2980183592716"><a name="p2980183592716"></a><a name="p2980183592716"></a>分布式训练集合通信端口。<span class="parmname" id="parmname1198063542717"><a name="parmname1198063542717"></a><a name="parmname1198063542717"></a>“name”</span>取值只能为<span class="parmvalue" id="parmvalue17980153515270"><a name="parmvalue17980153515270"></a><a name="parmvalue17980153515270"></a>“ascendjob-port”</span>，<span class="parmname" id="parmname8980135102711"><a name="parmname8980135102711"></a><a name="parmname8980135102711"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
    </td>
    </tr>
    </tbody>
    </table>

    **表 3**  huawei.com/schedule\_policy配置说明

    <a name="table1120511613153"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002511347099_row192066612155"><th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002511347099_p132062614153"><a name="zh-cn_topic_0000002511347099_p132062614153"></a><a name="zh-cn_topic_0000002511347099_p132062614153"></a>配置</p>
    </th>
    <th class="cellrowborder" valign="top" width="77.7%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002511347099_p5206126181520"><a name="zh-cn_topic_0000002511347099_p5206126181520"></a><a name="zh-cn_topic_0000002511347099_p5206126181520"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002511347099_row201261346162"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p457945418181"><a name="zh-cn_topic_0000002511347099_p457945418181"></a><a name="zh-cn_topic_0000002511347099_p457945418181"></a>chip4-node8</p>
    </td>
    <td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p7579105411817"><a name="zh-cn_topic_0000002511347099_p7579105411817"></a><a name="zh-cn_topic_0000002511347099_p7579105411817"></a>1个节点8张芯片，每4个芯片形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph18314192319429"><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph631452384213"><a name="zh-cn_topic_0000002511347099_ph631452384213"></a><a name="zh-cn_topic_0000002511347099_ph631452384213"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的整模块场景</span>/<span id="ph631452384213"><a name="ph631452384213"></a><a name="ph631452384213"></a>Atlas 350 推理卡内部共8张卡，每4张卡通过UB扣板连接。</p>
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

3.  根据实际需求，选择YAML示例并进行如下修改。

    **表 4**  操作示例

    <a name="table1990975873315"></a>
    <table><thead align="left"><tr id="row890916589334"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p169091858183312"><a name="p169091858183312"></a><a name="p169091858183312"></a>特性名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p1690905820337"><a name="p1690905820337"></a><a name="p1690905820337"></a>操作参考</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row690965893312"><td class="cellrowborder" rowspan="5" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p690915813336"><a name="p690915813336"></a><a name="p690915813336"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p79091587334"><a name="p79091587334"></a><a name="p79091587334"></a><a href="#li1888133815128">在推理服务器（插Atlas 300I 推理卡）上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row42351537182719"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p11235173717273"><a name="p11235173717273"></a><a name="p11235173717273"></a><a href="#li108651415102917">在推理服务器（插Atlas 300I Duo 推理卡）上创建分布式任务</a></p>
    </td>
    </tr>
    <tr id="row59097587338"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p99091858183320"><a name="p99091858183320"></a><a name="p99091858183320"></a><a href="#li727503931310">在Atlas 推理系列产品（非Atlas 200I SoC A1 核心板和Atlas 300I Duo 推理卡）上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row1890917580338"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p149091958123319"><a name="p149091958123319"></a><a name="p149091958123319"></a><a href="#li132621943121411">在Atlas 200I SoC A1 核心板上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row1843115298483"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p4432729124818"><a name="p4432729124818"></a><a name="p4432729124818"></a><a href="#li1134113548015">在Atlas 800I A2 推理服务器上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row11909558173312"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p109091558193312"><a name="p109091558193312"></a><a name="p109091558193312"></a>静态vNPU调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p590919584334"><a name="p590919584334"></a><a name="p590919584334"></a><a href="#li21860112612">在Atlas 推理系列产品（非Atlas 200I SoC A1 核心板）上创建单卡任务</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    -   <a name="li1888133815128"></a>使用**整卡调度**特性，参考本配置。以infer-deploy.yaml为例，在推理服务器（插Atlas 300I 推理卡）节点创建一个单卡推理任务，并且启用了调度策略，示例如下。修改完成后直接执行[步骤4](#li59320351213)。

        ```
        apiVersion: apps/v1
        kind: Deployment
        ...
        spec:
          template:
            metadata: 
              labels:
                 app: infers
                 host-arch: huawei-arm
                 npu-310-strategy: card     # 按推理卡调度
        ...
            spec:
              schedulerName: volcano        # 此时调度器必须为Volcano
              nodeSelector:
                host-arch: huawei-arm    # 可选值，根据实际情况填写
        ...
              containers:
              - image: ubuntu-infer:v1
        ...
              env:
              - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend310']               # 需要和下面resources.requests保持一致
                resources:
                  requests:
                    huawei.com/Ascend310: 1                   # 申请的芯片数量
                  limits:
                    huawei.com/Ascend310: 1
        ...
        ```

    -   使用**整卡调度**特性，参考本配置。以pytorch\_acjob\_infer\_310p\_with\_ranktable.yaml为例，在推理服务器（插Atlas 300I Duo 推理卡）节点创建一个分布式推理任务，并且启用了调度策略，示例如下。修改完成后直接执行[步骤4](#li59320351213)。

        ```
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-infer-test
          labels:
        ...
            app: infers
            npu-310-strategy: chip      # 按昇腾AI处理器调度
            distributed: "true"         # 分布式推理
            duo: "true"             # 使用Atlas 300I Duo 推理卡
            ring-controller.atlas: ascend-310P  # 标识任务使用的芯片的产品类型
            framework: pytorch       # 框架类型
        
        spec:
          schedulerName: volcano     #当Ascend Operator组件的启动参数enableGangScheduling为true时生效  
          runPolicy:
            schedulingPolicy:    
              minAvailable: 2  # 任务总副本数
              queue: default      # 任务所属队列
          successPolicy: AllWorkers # 任务成功的前提
          replicaSpecs:
            Master:
              replicas: 1     # 任务副本数
        ...
                spec:
                  nodeSelector:
                    servertype: Ascend310P
                  containers:
                    - name: ascend         # 必须为ascend，不能修改
                      image: ubuntu:22.04          # 根据实际情况修改镜像名称
        ...
                        - name: ASCEND_VISIBLE_DEVICES
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.annotations['huawei.com/Ascend310P']       # 给容器挂载相应类型的芯片
        ...
                      ports:                  # 分布式训练集合通信端口
                        - containerPort: 2222     
                          name: ascendjob-port    
                      resources:
                        limits:
                          huawei.com/Ascend310P: 1   # 申请的芯片数量
                        requests:
                          huawei.com/Ascend310P: 1  #与limits取值一致
                      volumeMounts:
        ...
                        - name: ranktable                  
                          mountPath: /user/serverid/devindex/config
        ...
                  volumes:
        ...
                    - name: ranktable
                      hostPath:
                        path: /user/mindx-dl/ranktable/default.default-infer-test  
        ...
            Worker:
        ...
                spec:
                  containers:
                    - name: ascend     #必须为ascend，不能修改
                      image: ubuntu:22.04      # 根据实际情况修改镜像名称
                      env:
        ...
                        - name: ASCEND_VISIBLE_DEVICES
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.annotations['huawei.com/Ascend310P']      # 给容器挂载相应类型的芯片
        ...
                      ports:     # 分布式训练集合通信端口
                        - containerPort: 2222      
                          name: ascendjob-port      
                      resources:
                        limits:
                          huawei.com/Ascend310P: 1   # 申请的芯片数
                        requests:
                          huawei.com/Ascend310P: 1   #与limits取值一致
                      volumeMounts:
        ...
                          # 可选，使用Ascend Operator组件为PyTorch和MindSpore框架生成RankTable文件，需要新增以下加粗字段，设置容器中hccl.json文件保存路径
                        - name: ranktable                  
                          mountPath: /user/serverid/devindex/config
        ...
                  volumes:
        ...
                    # 可选，使用Ascend Operator组件为PyTorch框架生成RankTable文件，需要新增以下加粗字段，设置hccl.json文件保存路径
                    - name: ranktable
                      hostPath:
                        path: /user/mindx-dl/ranktable/default.default-infer-test  # 共享存储或者本地存储路径，请根据实际情况修改
        ...
        ```

    -   使用**整卡调度**特性，参考本配置。以infer-deploy.yaml为例，在Atlas 推理系列产品节点（非Atlas 200I SoC A1 核心板和Atlas 300I Duo 推理卡节点）创建一个不使用混插模式的单卡推理任务，示例如下。修改完成后直接执行[步骤4](#li59320351213)。

        ```
        apiVersion: apps/v1
        kind: Deployment
        ...
        spec:
          template:
            metadata: 
              labels:
                 app: infers
        ...
            spec:
              affinity:        # 本段代码表示不调度到Atlas 200I SoC A1 核心板节点
                nodeAffinity:
                  requiredDuringSchedulingIgnoredDuringExecution:
                    nodeSelectorTerms:
                      - matchExpressions:
                          - key: servertype
                            operator: NotIn
                            values:
                              - soc
              schedulerName: volcano 
              nodeSelector:
                host-arch: huawei-arm 
        ...
              containers:
              - image: ubuntu-infer:v1
        ...
              env:
              - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend310P']               # 给容器挂载相应类型的芯片
        ...
                resources:
                  requests:
                    huawei.com/Ascend310P: 1     # 申请的芯片数量
                  limits:
                    huawei.com/Ascend310P: 1
        ...
        ```

        >[!NOTE] 说明 
        >因为Atlas 200I SoC A1 核心板节点需要挂载的目录和文件与其他类型节点不一致，为了避免推理失败，如果需要使用Atlas 推理系列产品芯片，且集群中有Atlas 200I SoC A1 核心板节点但是不希望调度到这类节点上，请在示例的YAML中增加“affinity“字段，表示不调度到有“servertype=soc“标签的节点上。

    -   <a name="li132621943121411"></a>使用**整卡调度**特性，参考本配置。以infer-deploy-310p-1usoc.yaml为例，在Atlas 200I SoC A1 核心板节点（不支持混插模式）创建一个单卡推理任务，示例如下。修改完成后直接执行[步骤4](#li59320351213)。

        ```
        apiVersion: apps/v1
        kind: Deployment
        ...
        spec:
          template:
            metadata: 
              labels:
                 app: infers
        ...
            spec:
              schedulerName: volcano 
              nodeSelector:
                host-arch: huawei-arm
                servertype: soc      # 该标签表示仅能调度到Atlas 200I SoC A1 核心板节点
        ...
              containers:
              - image: ubuntu-infer:v1
        ...
              env:
              - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend310P']               # 给容器挂载相应类型的芯片
        ...
                resources:
                  requests:
                    huawei.com/Ascend310P: 1     # 申请的芯片数量
                  limits:
                    huawei.com/Ascend310P: 1
        ...
        ```

    -   <a name="li1134113548015"></a>使用**整卡调度**特性，参考本配置。以infer-vcjob-910.yaml为例，在Atlas 800I A2 推理服务器上创建一个单卡推理任务，示例如下。修改完成后直接执行[步骤4](#li59320351213)。

        ```
        apiVersion: batch.volcano.sh/v1alpha1
        kind: Job
        metadata:
          name: mindx-infer-test
          namespace: vcjob                      # 根据实际情况选择合适的命名空间
          labels:
            ring-controller.atlas: ascend-{xxx}b
            fault-scheduling: "force"
        spec:
        ...
            template:
              metadata:
                labels:
                  app: infer
                  ring-controller.atlas: ascend-{xxx}b
              spec:
                containers:
                  - image: infer_image:latest             # 推理镜像名称，以实际情况为准
        ...
              env:
              - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
                      requests:
                        huawei.com/Ascend910: 1          # 所需的芯片数量
                      limits:
                        huawei.com/Ascend910: 1          # 必须与requests的值一致.
                    volumeMounts:
                      - name: localtime                  # 容器时间必须与主机时间一致
                        mountPath: /etc/localtime
                nodeSelector:
                  host-arch: huawei-arm                  # 根据实际情况进行配置
                  accelerator-type: module-{xxx}b-8      # Atlas 800I A2 推理服务器
                volumes:
                - name: localtime
                  hostPath:
                    path: /etc/localtime
                restartPolicy: OnFailure
        ```

    -   使用**静态vNPU调度**特性，参考本配置。以infer-deploy.yaml为例，在Atlas 推理系列产品节点（非Atlas 200I SoC A1 核心板节点）创建一个使用vNPU的推理任务，示例如下。修改完成后直接执行[步骤4](#li59320351213)。

        ```
        apiVersion: apps/v1
        kind: Deployment
        ...
        spec:
          template:
            metadata: 
              labels:
                 app: infers
        ...
            spec:
              schedulerName: volcano 
              nodeSelector:
                host-arch: huawei-arm 
        ...
              containers:
              - image: ubuntu-infer:v1
        ...
        # 静态vNPU调度暂不支持ASCEND_VISIBLE_DEVICES相关字段，需要删除以下加粗字段
                env:
                - name: ASCEND_VISIBLE_DEVICES
                  valueFrom:
                    fieldRef:
                      fieldPath: metadata.annotations['huawei.com/Ascend310P']    # 删除到此行
                resources:
                  requests:
                    huawei.com/Ascend310P-2c: 1      # vNPU调度此处数量只能为1
                  limits:
                    huawei.com/Ascend310P-2c: 1       # 必须与requests的值一致
        ...
        ```

4.  挂载权重文件。

    ```
    ...
                  ports:     # 分布式训练集合通信端口
                    - containerPort: 2222      
                      name: ascendjob-port      
                  resources:
                    limits:
                      huawei.com/Ascend310P: 1   # 申请的芯片数
                    requests:
                      huawei.com/Ascend310P: 1   # 与limits取值一致
                  volumeMounts:
    ...
                      # 权重文件挂载路径
                    - name: weights                  
                      mountPath: /path-to-weights
    ...
              volumes:
    ...
                # 权重文件挂载路径
                - name: weights
                  hostPath:
                    path: /path-to-weights  # 共享存储或者本地存储路径，请根据实际情况修改
    ...
    ```

    >[!NOTE] 说明 
    >-   /path-to-weights为模型权重，需要用户自行准备。mindie镜像可以参考镜像中$ATB\_SPEED\_HOME\_PATH/examples/models/llama3/README.md文件中的说明进行下载。
    >-   ATB_SPEED_HOME_PATH默认路径为“/usr/local/Ascend/atb-models”，在source模型仓中set_env.sh脚本时已配置，用户无需自行配置。

5.  <a name="li59320351213"></a>修改示例YAML中容器启动命令，如下加粗部分所示，如果没有则添加“command”字段。

    ```
    ...
          containers:
          - image: ubuntu-infer:v1
    ...
            command: ["/bin/bash", "-c", "cd $ATB_SPEED_HOME_PATH; python examples/run_pa.py --model_path /path-to-weights"]
            resources:
              requests:
    ...
    ```


#### 下发任务<a name="ZH-CN_TOPIC_0000002479387146"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```
kubectl apply -f XXX.yaml
```

例如：

```
kubectl apply -f infer-310p-1usoc.yaml
```

回显示例如下：

```
job.batch/resnetinfer1-2 created
```

>[!NOTE] 说明 
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f** **_XXX__.yaml_命令删除原任务，再重新下发任务。


#### 查看任务进程<a name="ZH-CN_TOPIC_0000002511347103"></a>

**操作步骤<a name="zh-cn_topic_0000001609474293_section96791230183711"></a>**

1.  执行以下命令，查看Pod运行状况。

    ```
    kubectl get pod --all-namespaces
    ```

    回显示例如下：

    ```
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
    ...
    default          resnetinfer1-2-scpr5                      1/1     Running   0          8s
    ...
    ```

2.  执行以下命令，查看运行推理任务的节点详情。

    ```
    kubectl describe node <hostname>
    ```

    例如：

    ```
    kubectl describe node ubuntu
    ```

    -   **整卡调度**回显示例如下：

        ```
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests     Limits
          --------              --------     ------
          cpu                   4 (2%)       3500m (1%)
          memory                2140Mi (0%)  4040Mi (0%)
          ephemeral-storage     0 (0%)       0 (0%)
          huawei.com/Ascend310P  1            1
        Events:
          Type    Reason    Age   From                Message
          ----    ------    ----  ----                -------
          Normal  Starting  36m   kube-proxy, ubuntu  Starting kube-proxy.
        ...
        ```

        在显示的信息中，找到“Allocated resources“下的**huawei.com/Ascend310P**，该参数取值在执行推理任务之后会增加，增加数量为推理任务使用的NPU芯片个数。

    -   **静态vNPU调度**回显示例如下：

        ```
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests     Limits
          --------              --------     ------
          cpu                   4 (2%)       3500m (1%)
          memory                2140Mi (0%)  4040Mi (0%)
          ephemeral-storage     0 (0%)       0 (0%)
          Ascend310P-2c  1            1
        Events:
          Type    Reason    Age   From                Message
          ----    ------    ----  ----                -------
          Normal  Starting  36m   kube-proxy, ubuntu  Starting kube-proxy.
        ...
        ```

        在显示的信息中，找到“Allocated resources“下的**Ascend310P-2c**，该参数取值在执行推理任务之后会增加，增加数量为推理任务使用的vNPU芯片个数。

    >[!NOTE] 说明 
    >-   如果使用的是Atlas 推理系列产品非混插模式，则上述字段显示为**Ascend310P，Ascend310P-2c**。
    >-   如果使用的是Atlas 推理系列产品混插模式，则上述字段显示为**Ascend310P-V、Ascend310P-VPro、Ascend310P-IPro之一**。


#### 查看整卡调度或静态vNPU调度结果<a name="ZH-CN_TOPIC_0000002511347083"></a>

**操作步骤<a name="zh-cn_topic_0000001558675486_section96791230183711"></a>**

在管理节点执行以下命令，查看推理结果。

```
kubectl logs -f resnetinfer1-2-scpr5
```

回显示例如下，以实际回显为准。

```
[2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Answer[0]:  Deep learning is a subset of machine learning that uses neural networks with multiple layers to model complex relationships between
[2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Generate[0] token num: (0, 20)
```

>[!NOTE] 说明 
>_resnetinfer1-2-scpr5_为[1](查看任务进程-11.md#zh-cn_topic_0000001609474293_li251162355411)中创建任务对应的Pod名称。


#### （可选）查看推理卡故障恢复结果<a name="ZH-CN_TOPIC_0000002511427061"></a>

当NPU故障时，Volcano组件会自动将该NPU上运行的推理任务调度到其他节点上（其他调度器不支持该功能，需要用户自行实现）；再由Ascend Device Plugin组件实现NPU的复位操作，使NPU恢复健康。用户可以通过**npu-smi info**命令查看NPU信息，若故障的NPU当前“health“字段显示的信息为“OK“，表示NPU已经恢复健康。

>[!NOTE] 说明 
>Ascend Device Plugin组件实现NPU的复位功能，需要确保当前故障NPU上没有推理任务或者推理任务已经被调走。若用户使用其他调度器且该调度器没有实现重调度功能，可以手动删除该NPU上的推理任务。


#### 删除任务<a name="ZH-CN_TOPIC_0000002511427043"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```
kubectl delete -f XXX.yaml
```

例如：

```
kubectl delete -f infer-310p-1usoc.yaml
```

回显示例如下：

```
root@ubuntu:/home/test/yaml# kubectl delete -f infer-310p-1usoc.yaml 
job "resnetinfer1-2" deleted
```



### 通过命令行使用（其他调度器）<a name="ZH-CN_TOPIC_0000002479227152"></a>

通过命令行使用（其他调度器）和通过命令行使用（Volcano）使用流程一致，只有任务YAML有所不同，用户可以准备好相应YAML后参考[通过命令行使用（Volcano）](#通过命令行使用volcano-1)章节使用。

**操作步骤<a name="section1290513712233"></a>**

1.  请从集群调度代码仓中下载YAML文件。

    **表 1**  任务类型与硬件型号对应YAML文件

    <a name="zh-cn_topic_0000001609074213_table15169151021912"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001609074213_row16169201019192"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001609074213_p4169191017192"><a name="zh-cn_topic_0000001609074213_p4169191017192"></a><a name="zh-cn_topic_0000001609074213_p4169191017192"></a>任务类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001609074213_p20181111517147"><a name="zh-cn_topic_0000001609074213_p20181111517147"></a><a name="zh-cn_topic_0000001609074213_p20181111517147"></a>硬件型号</p>
    </th>
    <th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001609074213_p181811156149"><a name="zh-cn_topic_0000001609074213_p181811156149"></a><a name="zh-cn_topic_0000001609074213_p181811156149"></a>YAML文件名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.4"><p id="p1510912587514"><a name="p1510912587514"></a><a name="p1510912587514"></a>获取链接</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001609074213_row81696106197"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p18169161011913"><a name="zh-cn_topic_0000001609074213_p18169161011913"></a><a name="zh-cn_topic_0000001609074213_p18169161011913"></a><span id="zh-cn_topic_0000001609074213_ph1319220540374"><a name="zh-cn_topic_0000001609074213_ph1319220540374"></a><a name="zh-cn_topic_0000001609074213_ph1319220540374"></a>K8s</span>或其他调度器场景下的Job任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p4169310141916"><a name="zh-cn_topic_0000001609074213_p4169310141916"></a><a name="zh-cn_topic_0000001609074213_p4169310141916"></a><span id="zh-cn_topic_0000001609074213_ph1355971413491"><a name="zh-cn_topic_0000001609074213_ph1355971413491"></a><a name="zh-cn_topic_0000001609074213_ph1355971413491"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001609074213_p17169210171914"><a name="zh-cn_topic_0000001609074213_p17169210171914"></a><a name="zh-cn_topic_0000001609074213_p17169210171914"></a>infer-310p-1usoc.yaml</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p63731221566"><a name="p63731221566"></a><a name="p63731221566"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/inference/without-volcano" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row63291517182014"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p13330817142010"><a name="zh-cn_topic_0000001609074213_p13330817142010"></a><a name="zh-cn_topic_0000001609074213_p13330817142010"></a>其他类型推理节点</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p433071711206"><a name="zh-cn_topic_0000001609074213_p433071711206"></a><a name="zh-cn_topic_0000001609074213_p433071711206"></a>infer.yaml</p>
    </td>
    </tr>
    </tbody>
    </table>

2.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    **表 2**  YAML文件参数说明

    <a name="zh-cn_topic_0000001609074213_table5589101114528"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001609074213_row125891211155216"><th class="cellrowborder" valign="top" width="21.122112211221122%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074213_p13658124194513"><a name="zh-cn_topic_0000001609074213_p13658124194513"></a><a name="zh-cn_topic_0000001609074213_p13658124194513"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="41.16411641164117%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074213_p4658152420459"><a name="zh-cn_topic_0000001609074213_p4658152420459"></a><a name="zh-cn_topic_0000001609074213_p4658152420459"></a>取值</p>
    </th>
    <th class="cellrowborder" valign="top" width="37.71377137713771%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074213_p8302202619484"><a name="zh-cn_topic_0000001609074213_p8302202619484"></a><a name="zh-cn_topic_0000001609074213_p8302202619484"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001609074213_row145900112522"><td class="cellrowborder" valign="top" width="21.122112211221122%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p65901011115213"><a name="zh-cn_topic_0000001609074213_p65901011115213"></a><a name="zh-cn_topic_0000001609074213_p65901011115213"></a>image</p>
    </td>
    <td class="cellrowborder" valign="top" width="41.16411641164117%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p105901311195216"><a name="zh-cn_topic_0000001609074213_p105901311195216"></a><a name="zh-cn_topic_0000001609074213_p105901311195216"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p5590191185217"><a name="zh-cn_topic_0000001609074213_p5590191185217"></a><a name="zh-cn_topic_0000001609074213_p5590191185217"></a>推理镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row14141145104416"><td class="cellrowborder" valign="top" width="21.122112211221122%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p4393104684410"><a name="zh-cn_topic_0000001609074213_p4393104684410"></a><a name="zh-cn_topic_0000001609074213_p4393104684410"></a>replicas</p>
    </td>
    <td class="cellrowborder" valign="top" width="41.16411641164117%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p83938460446"><a name="zh-cn_topic_0000001609074213_p83938460446"></a><a name="zh-cn_topic_0000001609074213_p83938460446"></a>整数</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p11393154664410"><a name="zh-cn_topic_0000001609074213_p11393154664410"></a><a name="zh-cn_topic_0000001609074213_p11393154664410"></a>运行的任务副本数量。通常情况一般为1。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row2059051145219"><td class="cellrowborder" valign="top" width="21.122112211221122%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p8590101118522"><a name="zh-cn_topic_0000001609074213_p8590101118522"></a><a name="zh-cn_topic_0000001609074213_p8590101118522"></a>requests</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="41.16411641164117%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074213_ul1180139155411"></a><a name="zh-cn_topic_0000001609074213_ul1180139155411"></a><ul id="zh-cn_topic_0000001609074213_ul1180139155411"><li>推理服务器（插<span id="ph163696166292"><a name="ph163696166292"></a><a name="ph163696166292"></a>Atlas 300I 推理卡</span>）：<p id="zh-cn_topic_0000001609074213_p364765019017"><a name="zh-cn_topic_0000001609074213_p364765019017"></a><a name="zh-cn_topic_0000001609074213_p364765019017"></a>huawei.com/Ascend310: <em id="zh-cn_topic_0000001609074213_i126472503016"><a name="zh-cn_topic_0000001609074213_i126472503016"></a><a name="zh-cn_topic_0000001609074213_i126472503016"></a>芯片数量</em></p>
    </li></ul>
    <a name="zh-cn_topic_0000001609074213_ul8938201113543"></a><a name="zh-cn_topic_0000001609074213_ul8938201113543"></a><ul id="zh-cn_topic_0000001609074213_ul8938201113543"><li><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>非混插模式：<p id="zh-cn_topic_0000001609074213_p464718509014"><a name="zh-cn_topic_0000001609074213_p464718509014"></a><a name="zh-cn_topic_0000001609074213_p464718509014"></a>huawei.com/Ascend310P: <em id="zh-cn_topic_0000001609074213_i06475509019"><a name="zh-cn_topic_0000001609074213_i06475509019"></a><a name="zh-cn_topic_0000001609074213_i06475509019"></a>芯片数量。</em></p>
    </li></ul>
    <a name="zh-cn_topic_0000001609074213_ul13727161475413"></a><a name="zh-cn_topic_0000001609074213_ul13727161475413"></a><ul id="zh-cn_topic_0000001609074213_ul13727161475413"><li><span id="ph181541455134013"><a name="ph181541455134013"></a><a name="ph181541455134013"></a>Atlas 推理系列产品</span>混插模式环境：<a name="zh-cn_topic_0000001609074213_ul8401842105312"></a><a name="zh-cn_topic_0000001609074213_ul8401842105312"></a><ul id="zh-cn_topic_0000001609074213_ul8401842105312"><li>huawei.com/Ascend310P-V: <em id="zh-cn_topic_0000001609074213_i16471550409"><a name="zh-cn_topic_0000001609074213_i16471550409"></a><a name="zh-cn_topic_0000001609074213_i16471550409"></a>芯片数量。</em></li><li>huawei.com/Ascend310P-VPro: <em id="zh-cn_topic_0000001609074213_i146476501013"><a name="zh-cn_topic_0000001609074213_i146476501013"></a><a name="zh-cn_topic_0000001609074213_i146476501013"></a>芯片数量。</em></li><li>huawei.com/Ascend310P-IPro: <em id="zh-cn_topic_0000001609074213_i06476501014"><a name="zh-cn_topic_0000001609074213_i06476501014"></a><a name="zh-cn_topic_0000001609074213_i06476501014"></a>芯片数量。</em></li></ul>
    </li><li><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>、<span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span>、<span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span>：huawei.com/Ascend910：<em id="i68164310406"><a name="i68164310406"></a><a name="i68164310406"></a>芯片数量</em></li></ul>
    <p id="zh-cn_topic_0000001609074213_p764710501704"><a name="zh-cn_topic_0000001609074213_p764710501704"></a><a name="zh-cn_topic_0000001609074213_p764710501704"></a>如：<em id="zh-cn_topic_0000001609074213_i564765010019"><a name="zh-cn_topic_0000001609074213_i564765010019"></a><a name="zh-cn_topic_0000001609074213_i564765010019"></a>huawei.com/Ascend310: 1</em></p>
    <a name="ul314865731117"></a><a name="ul314865731117"></a>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p11590211155219"><a name="zh-cn_topic_0000001609074213_p11590211155219"></a><a name="zh-cn_topic_0000001609074213_p11590211155219"></a>请求的NPU类型、数量，请根据实际修改。requests和limits下，芯片的名字和数量需保持一致。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row114301545157"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p16864941511"><a name="zh-cn_topic_0000001609074213_p16864941511"></a><a name="zh-cn_topic_0000001609074213_p16864941511"></a>limits</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row105901411135220"><td class="cellrowborder" valign="top" width="21.122112211221122%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p1590711195215"><a name="zh-cn_topic_0000001609074213_p1590711195215"></a><a name="zh-cn_topic_0000001609074213_p1590711195215"></a>（可选）host-arch</p>
    </td>
    <td class="cellrowborder" valign="top" width="41.16411641164117%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p1650105613241"><a name="zh-cn_topic_0000001609074213_p1650105613241"></a><a name="zh-cn_topic_0000001609074213_p1650105613241"></a><span id="zh-cn_topic_0000001609074213_ph16676195493717"><a name="zh-cn_topic_0000001609074213_ph16676195493717"></a><a name="zh-cn_topic_0000001609074213_ph16676195493717"></a>ARM</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
    <p id="zh-cn_topic_0000001609074213_p0658124184512"><a name="zh-cn_topic_0000001609074213_p0658124184512"></a><a name="zh-cn_topic_0000001609074213_p0658124184512"></a><span id="zh-cn_topic_0000001609074213_ph1274682034217"><a name="zh-cn_topic_0000001609074213_ph1274682034217"></a><a name="zh-cn_topic_0000001609074213_ph1274682034217"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p8590711115212"><a name="zh-cn_topic_0000001609074213_p8590711115212"></a><a name="zh-cn_topic_0000001609074213_p8590711115212"></a>需要运行推理任务的节点架构，请根据实际修改。<span id="zh-cn_topic_0000001609074213_ph183338272492"><a name="zh-cn_topic_0000001609074213_ph183338272492"></a><a name="zh-cn_topic_0000001609074213_ph183338272492"></a>Atlas 200I SoC A1 核心板</span>节点仅支持huawei-arm。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row144522615563"><td class="cellrowborder" valign="top" width="21.122112211221122%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074213_p944513266569"><a name="zh-cn_topic_0000001609074213_p944513266569"></a><a name="zh-cn_topic_0000001609074213_p944513266569"></a>servertype</p>
    </td>
    <td class="cellrowborder" valign="top" width="41.16411641164117%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074213_p544515265567"><a name="zh-cn_topic_0000001609074213_p544515265567"></a><a name="zh-cn_topic_0000001609074213_p544515265567"></a>soc</p>
    </td>
    <td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p202093166576"><a name="zh-cn_topic_0000001609074213_p202093166576"></a><a name="zh-cn_topic_0000001609074213_p202093166576"></a>服务器类型。</p>
    <a name="zh-cn_topic_0000001609074213_ul87677178911"></a><a name="zh-cn_topic_0000001609074213_ul87677178911"></a><ul id="zh-cn_topic_0000001609074213_ul87677178911"><li>调度到<span id="zh-cn_topic_0000001609074213_ph126801133164916"><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a>Atlas 200I SoC A1 核心板</span>节点上，必须要加上此配置，并参考<span class="filepath" id="zh-cn_topic_0000001609074213_filepath127811055718"><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a>“infer-310p-1usoc.yaml”</span>文件进行目录挂载。</li><li>其他类型节点不需要此参数。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

3.  根据实际需求，选择YAML示例并进行如下修改。

    **表 3**  操作示例

    <a name="table1819282912379"></a>
    <table><thead align="left"><tr id="row1719292923716"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p219217290379"><a name="p219217290379"></a><a name="p219217290379"></a>特性名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p81921229193710"><a name="p81921229193710"></a><a name="p81921229193710"></a>操作参考</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row19193152915372"><td class="cellrowborder" rowspan="3" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1619313291370"><a name="p1619313291370"></a><a name="p1619313291370"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1719312293376"><a name="p1719312293376"></a><a name="p1719312293376"></a><a href="#li1888133815128">在Atlas推理系列产品节点（非Atlas 200I SoC A1 核心板）上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row18193142910374"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p161931629153719"><a name="p161931629153719"></a><a name="p161931629153719"></a><a href="#li727503931310">在Atlas 200I SoC A1 核心板上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row119193361316"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p4432729124818"><a name="p4432729124818"></a><a name="p4432729124818"></a><a href="#li1134113548015">在Atlas 800I A2 推理服务器上创建单卡任务</a></p>
    </td>
    </tr>
    <tr id="row1319312910372"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1019310298378"><a name="p1019310298378"></a><a name="p1019310298378"></a>静态vNPU</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p171935298379"><a name="p171935298379"></a><a name="p171935298379"></a><a href="#li11239121841616">在Atlas 推理系列产品（非Atlas 200I SoC A1 核心板）上创建单卡任务</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    -   <a name="li1888133815128"></a>以infer.yaml为例，在Atlas 推理系列产品节点（非Atlas 200I SoC A1 核心板节点）创建一个不使用混插模式的单卡推理任务，示例如下。

        ```
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: resnetinfer1-1
        spec:
          template:
            spec:
              nodeSelector:
                host-arch: huawei-arm    # 可选值，根据实际情况填写
              affinity:        # 本段表示不调度到Atlas 200I SoC A1 核心板节点
                nodeAffinity:
                  requiredDuringSchedulingIgnoredDuringExecution:
                    nodeSelectorTerms:
                      - matchExpressions:
                          - key: servertype
                            operator: NotIn
                            values:
                              - soc
              containers:
              - image: ubuntu-infer:v1
        ...
                resources:
                  requests:
                    huawei.com/Ascend310P: 1
                  limits:
                    huawei.com/Ascend310P: 1
        ...
        ```

    -   <a name="li727503931310"></a>以infer-310p-1usoc.yaml为例，在Atlas 200I SoC A1 核心板节点（不支持混插模式）创建一个单卡推理任务，示例如下。

        ```
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: resnetinfer1-1-1usoc
        spec:
          template:
            spec:
              nodeSelector:
                host-arch: huawei-arm     # 可选值，根据实际情况填写
                servertype: soc               # 该标签表示仅能调度到Atlas 200I SoC A1 核心板节点
              containers:
              - image: ubuntu-infer:v1
        ...
                resources:
                  requests:
                    huawei.com/Ascend310P: 1
                  limits:
                    huawei.com/Ascend310P: 1
        ...
        ```

        >[!NOTE] 说明 
        >因为Atlas 200I SoC A1 核心板节点需要挂载的目录和文件与其他类型节点不一致，为了避免推理失败，如果需要使用Atlas 推理系列产品，且集群中有Atlas 200I SoC A1 核心板节点但是不希望调度到这类节点上，请在示例的YAML中增加“affinity“字段，表示不调度到有“servertype=soc“标签的节点上。

    -   <a name="li1134113548015"></a>使用**整卡调度**特性，参考本配置。以infer.yaml为例，在Atlas 800I A2 推理服务器上创建一个单卡推理任务，示例如下。

        ```
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: resnetinfer1-1
        spec:
          template:
            spec:
              nodeSelector:
                host-arch: huawei-arm   # 可选值，根据实际情况填写
        ...
              containers:
              - image: ubuntu-infer:v1
        ...
                resources:
                  requests:
                    huawei.com/Ascend910: 1
                  limits:
                    huawei.com/Ascend910: 1
        ...
        ```

    -   以infer.yaml为例，在Atlas 推理系列产品节点（非Atlas 200I SoC A1 核心板节点）创建一个使用vNPU的推理任务，示例如下。

        ```
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: resnetinfer1-1
        spec:
          template:
            spec:
              nodeSelector:
                host-arch: huawei-arm    # 可选值，根据实际情况填写
              containers:
              - image: ubuntu-infer:v1
        ...
                resources:
                  requests:
                    huawei.com/Ascend310P-2c: 1
                  limits:
                    huawei.com/Ascend310P-2c: 1
        ...
        ```


### 集成后使用<a name="ZH-CN_TOPIC_0000002479387128"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1.  根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2.  根据K8s官方提供的API库，来对任务进行创建、查询、删除等操作。
3.  创建、查询或删除操作任务时，用户需要将[示例YAML](#准备任务yaml-1)的内容转换成K8s官方API中定义的对象，通过官方库里面提供的API发送给K8s的API Server或者将YAML内容转换为JSON格式直接发送给K8s的API Server。



## 动态vNPU调度（推理）<a name="ZH-CN_TOPIC_0000002511427045"></a>

### 使用前必读<a name="ZH-CN_TOPIC_0000002511347087"></a>

**前提条件<a name="section121807404519"></a>**

在命令行场景下使用动态vNPU调度特性，需要确保已经安装如下组件；若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。动态vNPU调度特性只支持使用Volcano作为调度器，不支持使用其他调度器。

-   Volcano
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   ClusterD
-   NodeD

**使用方式<a name="zh-cn_topic_0000001559979444_section91871616135119"></a>**

动态vNPU调度特性的使用方式如下：

-   通过命令行使用：安装集群调度组件，通过命令行使用动态vNPU调度特性。
-   集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section10769161412815"></a>**

-   资源监测可以和推理场景下的所有特性一起使用。
-   集群中同时运行多个推理任务，每个任务使用的特性可以不同，但不能同时存在使用静态vNPU的任务和使用动态vNPU的任务。
-   动态vNPU调度特性需要搭配算力虚拟化特性一起使用，关于动态虚拟化的相关说明和操作请参见[动态虚拟化](./virtual_instance.md#动态虚拟化)章节。
-   动态vNPU调度仅支持下发单副本数或者多副本数的单机任务，每个副本独立工作，不支持分布式任务。

**支持的产品形态<a name="section169961844182917"></a>**

Atlas 推理系列产品

**使用流程<a name="zh-cn_topic_0000001559979444_section246711128536"></a>**

通过命令行使用动态vNPU调度特性流程可以参见[图1](#zh-cn_topic_0000001559979444_fig242524985412)。

**图 1**  使用流程<a name="zh-cn_topic_0000001559979444_fig242524985412"></a>  
![](../../figures/scheduling/使用流程-3.png "使用流程-3")

算力动态虚拟化实例涉及到相关集群调度组件的参数配置，请参见[动态虚拟化](./virtual_instance.md#动态虚拟化)章节完成修改。


### 实现原理<a name="ZH-CN_TOPIC_0000002511427057"></a>

根据推理任务类型的不同，特性的原理图略有差异。

**vcjob任务<a name="section11346231114"></a>**

vcjob任务原理图如[图1](#fig1918122131712)所示。

**图 1**  vcjob任务调度原理图<a name="fig1918122131712"></a>  
![](../../figures/scheduling/vcjob任务调度原理图-4.png "vcjob任务调度原理图-4")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到Node（节点对象）中。
    -   Ascend Device Plugin定期上报AI Core数量到Node中。
    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息分别写入cluster-info-device-cm和cluster-info-node-cm中。
3.  用户通过kubectl或者其他深度学习平台下发vcjob任务。
4.  volcano-controller为任务创建相应PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  当集群资源满足任务要求时，volcano-controller创建任务Pod。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入动态虚拟化的模板信息。
7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin根据模板信息动态虚拟化NPU。Ascend Docker Runtime协助挂载相应资源。

**deploy任务<a name="section41019364253"></a>**

deploy任务原理图如[图2](#fig349112913199)所示。

**图 2**  deploy任务调度原理图<a name="fig349112913199"></a>  
![](../../figures/scheduling/deploy任务调度原理图-5.png "deploy任务调度原理图-5")

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息。
    -   Ascend Device Plugin定期上报AI Core数量到Node中。
    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息分别写入cluster-info-device-cm和cluster-info-node-cm中。
3.  用户通过kubectl或者其他深度学习平台下发deploy任务。
4.  kube-controller为任务创建相应Pod。
5.  volcano-controller创建任务PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
6.  volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入动态虚拟化的模板信息。
7.  kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin根据Pod的annotation模板信息动态虚拟化NPU。Ascend Docker Runtime协助挂载相应资源。


### 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002479227144"></a>

#### 制作镜像<a name="ZH-CN_TOPIC_0000002511427049"></a>

**获取推理镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section971616541059"></a>**

可选择以下方式中的一种来获取推理镜像。

-   推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据系统架构（ARM或者x86\_64）下载**推理基础镜像（**如：[ascend-infer](https://www.hiascend.com/developer/ascendhub/detail/e02f286eef0847c2be426f370e0c2596)、[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)**）**。

    请注意，21.0.4版本之后推理基础镜像默认用户为非root用户，需要在下载基础镜像后对其进行修改，将默认用户修改为root。

    >[!NOTE] 说明 
    >基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。

-   （可选）可基于推理基础镜像定制用户自己的推理镜像，制作过程请参见[使用Dockerfile构建推理镜像](../common_operations.md#使用dockerfile构建推理镜像)。

    完成定制化修改后，用户可以给推理镜像重命名，以便管理和使用。

**加固镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section1294572963118"></a>**

下载或者制作的推理基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../references.md#容器安全加固)章节进行操作。


#### 脚本适配<a name="ZH-CN_TOPIC_0000002511347067"></a>

本章节以昇腾镜像仓库中推理镜像为例为用户介绍操作流程，该镜像已经包含了推理示例脚本，实际推理场景需要用户自行准备推理脚本。在拉取镜像前，需要确保当前环境的网络代理已经配置完成，确保该环境可以正常访问昇腾镜像仓库。

**从昇腾镜像仓库获取示例脚本<a name="section8181015175911"></a>**

1.  确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2.  在左侧导航栏选择推理镜像，然后选择[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)镜像，获取推理示例脚本。

    >[!NOTE] 说明 
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。


#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479387122"></a>

>[!NOTE] 说明 
>如果用户不使用Ascend Docker Runtime组件，Ascend Device Plugin只会帮助用户挂载“/dev“目录下的设备。其他目录（如“/usr“）用户需要自行修改YAML文件，挂载对应的驱动目录和文件。容器内挂载路径和宿主机路径保持一致。
>因为Atlas 200I SoC A1 核心板场景不支持Ascend Docker Runtime，用户也无需修改YAML文件。

**操作步骤<a name="zh-cn_topic_0000001558853680_zh-cn_topic_0000001609074213_section14665181617334"></a>**

1.  获取相应的YAML文件。

    **表 1**  YAML说明

    <a name="table0265132716351"></a>
    <table><thead align="left"><tr id="row132651727163516"><th class="cellrowborder" valign="top" width="15.36%" id="mcps1.2.5.1.1"><p id="p1447515933616"><a name="p1447515933616"></a><a name="p1447515933616"></a>任务类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="18.2%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001609074213_p20181111517147"><a name="zh-cn_topic_0000001609074213_p20181111517147"></a><a name="zh-cn_topic_0000001609074213_p20181111517147"></a>硬件型号</p>
    </th>
    <th class="cellrowborder" valign="top" width="37.769999999999996%" id="mcps1.2.5.1.3"><p id="p626512711358"><a name="p626512711358"></a><a name="p626512711358"></a>YAML名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="28.67%" id="mcps1.2.5.1.4"><p id="p3265172773514"><a name="p3265172773514"></a><a name="p3265172773514"></a>获取链接</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row826513275355"><td class="cellrowborder" valign="top" width="15.36%" headers="mcps1.2.5.1.1 "><p id="p278965223717"><a name="p278965223717"></a><a name="p278965223717"></a>Deployment</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="18.2%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p8853185832112"><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><span id="ph165178910439"><a name="ph165178910439"></a><a name="ph165178910439"></a>Atlas 推理系列产品</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="37.769999999999996%" headers="mcps1.2.5.1.3 "><p id="p142651427103519"><a name="p142651427103519"></a><a name="p142651427103519"></a>infer-deploy-dynamic.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="28.67%" headers="mcps1.2.5.1.4 "><p id="p1826522718352"><a name="p1826522718352"></a><a name="p1826522718352"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-deploy-dynamic.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row9265727173515"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p191941452171418"><a name="p191941452171418"></a><a name="p191941452171418"></a>Volcano Job</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15629131423715"><a name="p15629131423715"></a><a name="p15629131423715"></a>infer-vcjob-dynamic.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1626592713355"><a name="p1626592713355"></a><a name="p1626592713355"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-vcjob-dynamic.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    </tbody>
    </table>

2.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    在Atlas 推理系列产品上，以infer-deploy-dynamic.yaml为例，申请1个AI Core的参数配置示例如下。

    ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: resnetinfer1-1-deploy
      labels:
        app: infers
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: infers
      template:
        metadata:
          labels:
            app: infers
            fault-scheduling: "grace"           # 重调度所使用的label
             # 以下参数说明请参见表infer-deploy-dynamic.yaml参数说明
            ring-controller.atlas: ascend-310P 
            vnpu-dvpp: "null"         
            vnpu-level: "low"           
        spec:
          schedulerName: volcano              # 需要使用MindCluster的调度器Volcano
          nodeSelector:
            host-arch: huawei-arm
          containers:
            - image: ubuntu-infer:v1   # 示例镜像
    ...
    
              resources:
                requests:
                  huawei.com/npu-core: 1        # 使用静态虚拟化的vir01模板动态虚拟化NPU
                limits:
                  huawei.com/npu-core: 1        # 数值与requests保持一致
    ```

    **表 2**  infer-deploy-dynamic.yaml参数说明

    <a name="table116201128162111"></a>
    <table><thead align="left"><tr id="row362062812113"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p11620628192119"><a name="p11620628192119"></a><a name="p11620628192119"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p13620192817213"><a name="p13620192817213"></a><a name="p13620192817213"></a>取值</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p862022892120"><a name="p862022892120"></a><a name="p862022892120"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row136201528182116"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p56210289215"><a name="p56210289215"></a><a name="p56210289215"></a>vnpu-level</p>
    <p id="p262172815213"><a name="p262172815213"></a><a name="p262172815213"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p562182842111"><a name="p562182842111"></a><a name="p562182842111"></a>low</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662112892120"><a name="p662112892120"></a><a name="p662112892120"></a>低配，默认值，选择最低配置的虚拟化实例模板。</p>
    </td>
    </tr>
    <tr id="row196219286214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p146219285218"><a name="p146219285218"></a><a name="p146219285218"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p19621528112118"><a name="p19621528112118"></a><a name="p19621528112118"></a>性能优先。</p>
    <p id="p6621152812214"><a name="p6621152812214"></a><a name="p6621152812214"></a>在集群资源充足的情况下，将选择尽量高配的虚拟化实例模板；在整个集群资源已使用过多的情况下，如大部分物理NPU都已使用，每个物理NPU只剩下小部分AI Core，不足以满足高配虚拟化实例模板时，将使用相同AI Core数量下较低配置的其他模板。具体选择请参考<a href="./virtual_instance.md#虚拟化规则">虚拟化规则</a>章节中的“虚拟化模板”。</p>
    </td>
    </tr>
    <tr id="row1762192862114"><td class="cellrowborder" rowspan="3" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p462112842110"><a name="p462112842110"></a><a name="p462112842110"></a>vnpu-dvpp</p>
    <p id="p362120286216"><a name="p362120286216"></a><a name="p362120286216"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621122816219"><a name="p8621122816219"></a><a name="p8621122816219"></a>yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662162819213"><a name="p662162819213"></a><a name="p662162819213"></a>该<span id="ph1762113285210"><a name="ph1762113285210"></a><a name="ph1762113285210"></a>Pod</span>使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1762172862117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p46214285213"><a name="p46214285213"></a><a name="p46214285213"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5621162812213"><a name="p5621162812213"></a><a name="p5621162812213"></a>该<span id="ph1362102815215"><a name="ph1362102815215"></a><a name="ph1362102815215"></a>Pod</span>不使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1262122852117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p462192852111"><a name="p462192852111"></a><a name="p462192852111"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11621102818211"><a name="p11621102818211"></a><a name="p11621102818211"></a>默认值，不关注是否使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1762110285219"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2062182822111"><a name="p2062182822111"></a><a name="p2062182822111"></a>ring-controller.atlas</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621102882111"><a name="p8621102882111"></a><a name="p8621102882111"></a>ascend-310P</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1762182892113"><a name="p1762182892113"></a><a name="p1762182892113"></a>任务使用<span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>的标识。</p>
    </td>
    </tr>
    </tbody>
    </table>

    vnpu-level和vnpu-dvpp作用后，选择的vNPU模板可参考[表3](#zh-cn_topic_0000001557486210_table83781115185619)。

    **表 3**  dvpp和level作用结果

    <a name="zh-cn_topic_0000001557486210_table83781115185619"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001557486210_row1837817157565"><th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001557486210_p1024717408463"><a name="zh-cn_topic_0000001557486210_p1024717408463"></a><a name="zh-cn_topic_0000001557486210_p1024717408463"></a>AI Core请求数量</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001557486210_p192479402463"><a name="zh-cn_topic_0000001557486210_p192479402463"></a><a name="zh-cn_topic_0000001557486210_p192479402463"></a>vnpu-dvpp</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001557486210_p1024716402460"><a name="zh-cn_topic_0000001557486210_p1024716402460"></a><a name="zh-cn_topic_0000001557486210_p1024716402460"></a>vnpu-level</p>
    </th>
    <th class="cellrowborder" valign="top" width="9.221844368873775%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001557486210_p8247440174613"><a name="zh-cn_topic_0000001557486210_p8247440174613"></a><a name="zh-cn_topic_0000001557486210_p8247440174613"></a>是否降级</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001557486210_p0247164034611"><a name="zh-cn_topic_0000001557486210_p0247164034611"></a><a name="zh-cn_topic_0000001557486210_p0247164034611"></a>选择模板</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001557486210_row1937814158561"><td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p685714392538"><a name="zh-cn_topic_0000001557486210_p685714392538"></a><a name="zh-cn_topic_0000001557486210_p685714392538"></a>1</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p14248174010469"><a name="zh-cn_topic_0000001557486210_p14248174010469"></a><a name="zh-cn_topic_0000001557486210_p14248174010469"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p1385717396538"><a name="zh-cn_topic_0000001557486210_p1385717396538"></a><a name="zh-cn_topic_0000001557486210_p1385717396538"></a>任意值</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p38575391531"><a name="zh-cn_topic_0000001557486210_p38575391531"></a><a name="zh-cn_topic_0000001557486210_p38575391531"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p385603935319"><a name="zh-cn_topic_0000001557486210_p385603935319"></a><a name="zh-cn_topic_0000001557486210_p385603935319"></a>vir01</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row14379191520562"><td class="cellrowborder" rowspan="3" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p146191732155316"><a name="zh-cn_topic_0000001557486210_p146191732155316"></a><a name="zh-cn_topic_0000001557486210_p146191732155316"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p1248174014614"><a name="zh-cn_topic_0000001557486210_p1248174014614"></a><a name="zh-cn_topic_0000001557486210_p1248174014614"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p14619832145315"><a name="zh-cn_topic_0000001557486210_p14619832145315"></a><a name="zh-cn_topic_0000001557486210_p14619832145315"></a>low/其他值</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p126198326538"><a name="zh-cn_topic_0000001557486210_p126198326538"></a><a name="zh-cn_topic_0000001557486210_p126198326538"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p3248164094613"><a name="zh-cn_topic_0000001557486210_p3248164094613"></a><a name="zh-cn_topic_0000001557486210_p3248164094613"></a>vir02_1c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row0379131512562"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p122481840184611"><a name="zh-cn_topic_0000001557486210_p122481840184611"></a><a name="zh-cn_topic_0000001557486210_p122481840184611"></a>null</p>
    <p id="zh-cn_topic_0000001557486210_p13302164084616"><a name="zh-cn_topic_0000001557486210_p13302164084616"></a><a name="zh-cn_topic_0000001557486210_p13302164084616"></a></p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p162489402463"><a name="zh-cn_topic_0000001557486210_p162489402463"></a><a name="zh-cn_topic_0000001557486210_p162489402463"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p22482040124615"><a name="zh-cn_topic_0000001557486210_p22482040124615"></a><a name="zh-cn_topic_0000001557486210_p22482040124615"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p182481740174611"><a name="zh-cn_topic_0000001557486210_p182481740174611"></a><a name="zh-cn_topic_0000001557486210_p182481740174611"></a>vir02</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row53795153568"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p1324834017468"><a name="zh-cn_topic_0000001557486210_p1324834017468"></a><a name="zh-cn_topic_0000001557486210_p1324834017468"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p16248840154619"><a name="zh-cn_topic_0000001557486210_p16248840154619"></a><a name="zh-cn_topic_0000001557486210_p16248840154619"></a>vir02_1c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row1038061520567"><td class="cellrowborder" rowspan="7" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p0248240134616"><a name="zh-cn_topic_0000001557486210_p0248240134616"></a><a name="zh-cn_topic_0000001557486210_p0248240134616"></a>4</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p10248164012460"><a name="zh-cn_topic_0000001557486210_p10248164012460"></a><a name="zh-cn_topic_0000001557486210_p10248164012460"></a>yes</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p3248184024610"><a name="zh-cn_topic_0000001557486210_p3248184024610"></a><a name="zh-cn_topic_0000001557486210_p3248184024610"></a>low/其他值</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p4249114074618"><a name="zh-cn_topic_0000001557486210_p4249114074618"></a><a name="zh-cn_topic_0000001557486210_p4249114074618"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p8249540164619"><a name="zh-cn_topic_0000001557486210_p8249540164619"></a><a name="zh-cn_topic_0000001557486210_p8249540164619"></a>vir04_4c_dvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row338051565613"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p192491540164619"><a name="zh-cn_topic_0000001557486210_p192491540164619"></a><a name="zh-cn_topic_0000001557486210_p192491540164619"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p5249124011467"><a name="zh-cn_topic_0000001557486210_p5249124011467"></a><a name="zh-cn_topic_0000001557486210_p5249124011467"></a>vir04_3c_ndvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row738071535612"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p424914004612"><a name="zh-cn_topic_0000001557486210_p424914004612"></a><a name="zh-cn_topic_0000001557486210_p424914004612"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p192493409466"><a name="zh-cn_topic_0000001557486210_p192493409466"></a><a name="zh-cn_topic_0000001557486210_p192493409466"></a>vir04_3c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row1538131575615"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p924924018462"><a name="zh-cn_topic_0000001557486210_p924924018462"></a><a name="zh-cn_topic_0000001557486210_p924924018462"></a>yes</p>
    </td>
    <td class="cellrowborder" rowspan="4" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p2249440184619"><a name="zh-cn_topic_0000001557486210_p2249440184619"></a><a name="zh-cn_topic_0000001557486210_p2249440184619"></a>high</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p14272035114811"><a name="zh-cn_topic_0000001557486210_p14272035114811"></a><a name="zh-cn_topic_0000001557486210_p14272035114811"></a>-</p>
    <p id="zh-cn_topic_0000001557486210_p14249184019467"><a name="zh-cn_topic_0000001557486210_p14249184019467"></a><a name="zh-cn_topic_0000001557486210_p14249184019467"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p1324984017461"><a name="zh-cn_topic_0000001557486210_p1324984017461"></a><a name="zh-cn_topic_0000001557486210_p1324984017461"></a>vir04_4c_dvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row1438113158568"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p824916403462"><a name="zh-cn_topic_0000001557486210_p824916403462"></a><a name="zh-cn_topic_0000001557486210_p824916403462"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p15249440164616"><a name="zh-cn_topic_0000001557486210_p15249440164616"></a><a name="zh-cn_topic_0000001557486210_p15249440164616"></a>vir04_3c_ndvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row238110156568"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p1824974014620"><a name="zh-cn_topic_0000001557486210_p1824974014620"></a><a name="zh-cn_topic_0000001557486210_p1824974014620"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p10249124011467"><a name="zh-cn_topic_0000001557486210_p10249124011467"></a><a name="zh-cn_topic_0000001557486210_p10249124011467"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p324964074618"><a name="zh-cn_topic_0000001557486210_p324964074618"></a><a name="zh-cn_topic_0000001557486210_p324964074618"></a>vir04</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row17381415115611"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p2249340144615"><a name="zh-cn_topic_0000001557486210_p2249340144615"></a><a name="zh-cn_topic_0000001557486210_p2249340144615"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p924924064613"><a name="zh-cn_topic_0000001557486210_p924924064613"></a><a name="zh-cn_topic_0000001557486210_p924924064613"></a>vir04_3c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row8381181518563"><td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p102497405465"><a name="zh-cn_topic_0000001557486210_p102497405465"></a><a name="zh-cn_topic_0000001557486210_p102497405465"></a>8或8的倍数</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p42491440174615"><a name="zh-cn_topic_0000001557486210_p42491440174615"></a><a name="zh-cn_topic_0000001557486210_p42491440174615"></a>任意值</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p5249114074614"><a name="zh-cn_topic_0000001557486210_p5249114074614"></a><a name="zh-cn_topic_0000001557486210_p5249114074614"></a>任意值</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p1224920403467"><a name="zh-cn_topic_0000001557486210_p1224920403467"></a><a name="zh-cn_topic_0000001557486210_p1224920403467"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p102491940124613"><a name="zh-cn_topic_0000001557486210_p102491940124613"></a><a name="zh-cn_topic_0000001557486210_p102491940124613"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 说明 
    >AI Core的申请数量为8或8的倍数，表示使用整张NPU卡。

3.  挂载权重文件。

    ```
    ...
                  ports:     # 分布式训练集合通信端口
                    - containerPort: 2222      
                      name: ascendjob-port      
                  resources:
                    limits:
                      huawei.com/Ascend310P: 1   # 申请的芯片数
                    requests:
                      huawei.com/Ascend310P: 1   #与limits取值一致
                  volumeMounts:
    ...
                      # 权重文件挂载路径
                    - name: weights                  
                      mountPath: /path-to-weights
    ...
              volumes:
    ...
                # 权重文件挂载路径
                - name: weights
                  hostPath:
                    path: /path-to-weights  # 共享存储或者本地存储路径，请根据实际情况修改
    ...
    ```

    >[!NOTE] 说明 
    >-   /path-to-weights为模型权重，需要用户自行准备。mindie镜像可以参考镜像中$ATB\_SPEED\_HOME\_PATH/examples/models/llama3/README.md文件中的说明进行下载。
    >-   ATB_SPEED_HOME_PATH默认路径为“/usr/local/Ascend/atb-models”，在source模型仓中set_env.sh脚本时已配置，用户无需自行配置。

4.  修改所选YAML中的容器启动命令，如下加粗部分，如果没有则添加“command”字段。

    ```
    ...
          containers:
          - image: ubuntu-infer:v1
    ...
            command: ["/bin/bash", "-c", "cd $ATB_SPEED_HOME_PATH; python examples/run_pa.py --model_path /path-to-weights"]
            resources:
              requests:
    ...
    ```


#### 下发任务<a name="ZH-CN_TOPIC_0000002479227134"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```
kubectl apply -f XXX.yaml
```

例如：

```
kubectl apply -f infer-deploy-dynamic.yaml
```

回显示例如下：

```
job.batch/resnetinfer1-2 created
```

>[!NOTE] 说明 
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f** **_XXX__.yaml_命令删除原任务，再重新下发任务。


#### 查看任务进程<a name="ZH-CN_TOPIC_0000002511347071"></a>

**操作步骤<a name="zh-cn_topic_0000001609093161_zh-cn_topic_0000001609474293_section96791230183711"></a>**

1.  执行以下命令，查看Pod运行状况。

    ```
    kubectl get pod --all-namespaces
    ```

    回显示例如下：

    ```
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
    ...
    default          resnetinfer1-2-scpr5                      1/1     Running   0          8s
    ...
    ```

2.  查看运行推理任务的节点详情。
    1.  执行以下命令查看节点的名称。

        ```
        kubectl get node -A
        ```

    2.  根据上一步骤中查询到的节点名称，执行以下命令查看节点详情。

        ```
        kubectl describe node <nodename>
        ```

        回显示例如下：

        ```
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests     Limits
          --------              --------     ------
          cpu                   4 (2%)       3500m (1%)
          memory                2140Mi (0%)  4040Mi (0%)
          ephemeral-storage     0 (0%)       0 (0%)
          huawei.com/npu-core  4            4
        Events:
          Type    Reason    Age   From                Message
          ----    ------    ----  ----                -------
          Normal  Starting  36m   kube-proxy, ubuntu  Starting kube-proxy.
        ...
        ```

        在显示的信息中，找到“Allocated resources“下的**huawei.com/npu-core**，该参数取值在执行推理任务之后会增加，增加数量为推理任务使用的NPU芯片个数。


#### 查看动态vNPU调度结果<a name="ZH-CN_TOPIC_0000002479387120"></a>

**操作步骤<a name="zh-cn_topic_0000001559013282_zh-cn_topic_0000001558675486_section96791230183711"></a>**

在管理节点执行以下命令查看推理结果。

```
kubectl logs -f resnetinfer1-2-scpr5
```

回显示例如下，以实际回显为准。

```
[2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Answer[0]:  Deep learning is a subset of machine learning that uses neural networks with multiple layers to model complex relationships between
[2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Generate[0] token num: (0, 20)
```

>[!NOTE] 说明 
>_resnetinfer1-2-scpr5_：查看任务进程章节[步骤1](#查看任务进程-2)中运行的任务名称。


#### 删除任务<a name="ZH-CN_TOPIC_0000002511347065"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```
kubectl delete -f XXX.yaml
```

例如：

```
kubectl delete -f infer-deploy-dynamic.yaml
```

回显示例如下：

```
root@ubuntu:/home/test/yaml# kubectl delete -f infer-310p-1usoc.yaml 
job "resnetinfer1-1" deleted
```



### 集成后使用<a name="ZH-CN_TOPIC_0000002511347073"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1.  根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2.  根据K8s的官方API库，对任务进行创建、查询、删除等操作。
3.  创建、查询或删除任务时，用户需要将[示例YAML](#准备任务yaml-2)的内容转换成K8s官方API中定义的对象，通过官方API发送给K8s的API Server或者将YAML内容转换成JSON格式直接发送给K8s的API Server。



## 弹性训练<a name="ZH-CN_TOPIC_0000002479227142"></a>

### 使用前必读<a name="ZH-CN_TOPIC_0000002479227148"></a>

>[!NOTE] 说明 
>本章节描述的是基于Resilience Controller组件的弹性训练，该组件已经日落，相关资料将于8.2.RC1版本删除。最新的弹性训练能力请参见[弹性训练](./resumable_training.md#弹性训练)。

当出现硬件故障，且无备用设备时，集群调度组件将对故障节点进行隔离，并根据任务预设的规模和当前集群中可用的节点数，重新设置任务副本数，然后进行重调度和重训练（需进行脚本适配）。

**前提条件<a name="section722033433815"></a>**

-   确保环境中有配置相应的存储方案，比如使用NFS（Network File System），用户可以参见[安装NFS](../common_operations.md#安装nfs)进行操作。

    NFS需要用户根据使用情况进行目录隔离，NFS的随机读写性能必须能够在15分钟内保存完整的CKPT文件，建议用户使用专业的存储服务器，NFS具体性能要求给出如下参考。

    ![](../../figures/scheduling/6-2-2-1-折线图.png)

-   在命令行场景下使用弹性训练特性，需要确保已经安装如下组件。
    -   Ascend Device Plugin
    -   Ascend Docker Runtime
    -   Volcano（弹性训练特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    -   Ascend Operator
    -   NodeD
    -   Resilience Controller
    -   ClusterD

-   若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。

**使用方式<a name="section1215781619816"></a>**

弹性训练特性的使用方式如下：

-   [通过命令行使用](#通过命令行使用volcano-3)：安装集群调度组件，通过命令行使用弹性训练特性。
-   [集成后使用](#集成后使用-3)：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section252320491398"></a>**

-   资源监测可以和训练场景下的所有特性一起使用。
-   集群中同时跑多个训练任务，每个任务使用的特性可以不同。
-   集群调度组件管理的训练节点出现故障（安装昇腾AI处理器并启用NodeD的节点网络故障或者芯片故障）后，集群调度组件将对故障节点进行隔离，并根据任务预设的规模和当前集群中可用的节点数重新设置任务副本数，然后进行重调度和重训练（需进行脚本适配）。
-   重调度功能由Kubernetes（简称K8s）配合Volcano或者其他调度器实现。
-   更多说明详见[表1](#table1337017499206)。

    **表 1**  使用说明

    <a name="table1337017499206"></a>
    <table><thead align="left"><tr id="row1537112499205"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="p1737115497204"><a name="p1737115497204"></a><a name="p1737115497204"></a>场景</p>
    </th>
    <th class="cellrowborder" valign="top" width="80%" id="mcps1.2.3.1.2"><p id="p73711249152017"><a name="p73711249152017"></a><a name="p73711249152017"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row12371949152010"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="p5371204918208"><a name="p5371204918208"></a><a name="p5371204918208"></a>环境要求</p>
    <p id="p53711949192013"><a name="p53711949192013"></a><a name="p53711949192013"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p317010582217"><a name="p317010582217"></a><a name="p317010582217"></a>需要保证<span id="ph1093220553219"><a name="ph1093220553219"></a><a name="ph1093220553219"></a>K8s</span>集群中各节点时间一致，避免程序误判。</p>
    </td>
    </tr>
    <tr id="row1937117496204"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p132798152215"><a name="p132798152215"></a><a name="p132798152215"></a>用于检测NPU芯片间连通性的IP地址推荐配置为路由器的IP地址。</p>
    </td>
    </tr>
    <tr id="row12371154922014"><td class="cellrowborder" rowspan="3" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="p83713498203"><a name="p83713498203"></a><a name="p83713498203"></a>故障处理</p>
    <p id="p16371114912201"><a name="p16371114912201"></a><a name="p16371114912201"></a></p>
    <p id="p20371194932016"><a name="p20371194932016"></a><a name="p20371194932016"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p471123192217"><a name="p471123192217"></a><a name="p471123192217"></a>使用单机多卡进行训练，当出现故障时，优先按照原任务规格进行恢复，且任务规格遵循8、4、2、1卡的恢复策略。</p>
    </td>
    </tr>
    <tr id="row43711949102018"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1611164119224"><a name="p1611164119224"></a><a name="p1611164119224"></a>若<span id="ph23311639112215"><a name="ph23311639112215"></a><a name="ph23311639112215"></a>Resilience Controller</span>在重新调度任务的过程中，该任务出现新的故障，将不再进行处理。</p>
    </td>
    </tr>
    <tr id="row53711449152015"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p45358632316"><a name="p45358632316"></a><a name="p45358632316"></a>若在集群资源有限的场景中，当多个任务同时故障触发重调度，可能会出现由于资源不足而导致任务处于Pending状态。</p>
    </td>
    </tr>
    <tr id="row046671715231"><td class="cellrowborder" rowspan="3" valign="top" width="20%" headers="mcps1.2.3.1.1 "><p id="p1471143519233"><a name="p1471143519233"></a><a name="p1471143519233"></a>特性说明</p>
    </td>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.2 "><p id="p87431448172312"><a name="p87431448172312"></a><a name="p87431448172312"></a>本特性不适用于虚拟化实例场景。</p>
    </td>
    </tr>
    <tr id="row1526201918238"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p57511456152312"><a name="p57511456152312"></a><a name="p57511456152312"></a>本特性目前支持服务器和芯片间数据并行和混合并行的分布式vcjob类型的训练任务。</p>
    </td>
    </tr>
    <tr id="row16951221172311"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1993291632412"><a name="p1993291632412"></a><a name="p1993291632412"></a>本特性仅支持设备故障和服务器网络故障检测，说明如下：</p>
    <a name="ul175182082413"></a><a name="ul175182082413"></a><ul id="ul175182082413"><li>设备故障支持<span id="ph1914015620494"><a name="ph1914015620494"></a><a name="ph1914015620494"></a>《<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100469490" target="_blank" rel="noopener noreferrer">Atlas 中心训练服务器 25.0.RC1 健康管理故障定义</a>》</span>中DCMI接口上报的<span class="parmvalue" id="parmvalue11232151011244"><a name="parmvalue11232151011244"></a><a name="parmvalue11232151011244"></a>“重执行业务”</span>、<span class="parmvalue" id="parmvalue112321310182410"><a name="parmvalue112321310182410"></a><a name="parmvalue112321310182410"></a>“热复位芯片”</span>和<span class="parmvalue" id="parmvalue1423217104248"><a name="parmvalue1423217104248"></a><a name="parmvalue1423217104248"></a>“隔离芯片”</span>类型的错误。</li><li>设备网络探测工具hccn_tool检测到的设备网络故障；服务器网络故障依赖于<span id="ph1523271015245"><a name="ph1523271015245"></a><a name="ph1523271015245"></a>NodeD</span>组件的节点状态上报机制，<span id="ph1123281019246"><a name="ph1123281019246"></a><a name="ph1123281019246"></a>NodeD</span>未正确安装或者节点间网络不通都会影响该故障检测功能。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

**支持的产品形态<a name="section10503153618487"></a>**

支持Atlas 800 训练服务器产品使用弹性训练。

**使用流程<a name="section9435132545416"></a>**

通过命令行使用弹性训练特性流程可以参见[图1](#fig1445992135513)。

**图 1**  使用流程<a name="fig1445992135513"></a>  
![](../../figures/scheduling/使用流程-6.png "使用流程-6")


### 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002511427031"></a>

#### （可选）配置组件<a name="ZH-CN_TOPIC_0000002479227154"></a>

如果用户在安装Ascend Device Plugin和NodeD时，已经配置了弹性训练相关功能，则可以跳过本章节；若没有配置，则需要对组件[MindCluster Ascend Device Plugin](#zh-cn_topic_0000001609393673_section22911654123018)和[MindCluster NodeD](#section4599195414500)进行相关配置才能正常使用本特性。

**配置Ascend Device Plugin<a name="zh-cn_topic_0000001609393673_section22911654123018"></a>**

在重调度策略开启的情况下，Ascend Device Plugin的异常也会触发故障重调度。

1.  修改Ascend Device Plugin组件的启动YAML，修改如下所示加粗部分。

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
                     -volcanoType=true                    # 重调度场景下必须使用Volcano。
                     -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                     -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒。
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

    如在Atlas 训练系列产品启动该组件，示例如下。

    ```
    kubectl apply -f device-plugin-volcano-v7.3.0.yaml
    ```

**配置NodeD<a name="section4599195414500"></a>**

用户可以通过手动修改NodeD的启动YAML来配置节点状态上报间隔。

1.  执行以下命令，编辑NodeD组件的启动YAML文件。

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
              args: [ "/home/hwMindX/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 -reportInterval=5" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: false
                capabilities:
                  drop: [ "ALL" ]
                runAsUser: 9000
                runAsGroup: 9000
              volumeMounts:
                - name: log-noded
    ...
    ```

    >[!NOTE] 说明 
    >-   K8s[默认40秒未收到节点响应时](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)将该节点置为NotReady。
    >-   当K8s APIServer请求压力变大时，可根据实际情况增大间隔时间，以减轻APIServer压力。


#### 制作镜像<a name="ZH-CN_TOPIC_0000002511427037"></a>

弹性训练需要训练基础镜像，用户需要根据所使用的训练框架参见[制作镜像](../common_operations.md#制作镜像)章节进行制作。

>[!NOTE] 说明
>MindSpore框架的[盘古模型](#选择yaml示例-1)，还需要参考本章继续制作适配盘古模型的镜像。

**前提条件<a name="zh-cn_topic_0272789326_section193545302315"></a>**

请按照[表1](#zh-cn_topic_0272789326_table13971125465512)所示，获取对应操作系统的软件包与打包镜像所需Dockerfile文件与脚本文件，断点续训软件包名称中\{version\}表示版本号。

**表 1**  所需软件

<a name="zh-cn_topic_0272789326_table13971125465512"></a>
<table><thead align="left"><tr id="zh-cn_topic_0272789326_row19971185414551"><th class="cellrowborder" valign="top" width="28.88%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0272789326_p0971105411555"><a name="zh-cn_topic_0272789326_p0971105411555"></a><a name="zh-cn_topic_0272789326_p0971105411555"></a>软件包</p>
</th>
<th class="cellrowborder" valign="top" width="13.96%" id="mcps1.2.5.1.2"><p id="p326921620610"><a name="p326921620610"></a><a name="p326921620610"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30.42%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0272789326_p1097165410558"><a name="zh-cn_topic_0272789326_p1097165410558"></a><a name="zh-cn_topic_0272789326_p1097165410558"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="26.740000000000002%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0272789326_p39711454155520"><a name="zh-cn_topic_0272789326_p39711454155520"></a><a name="zh-cn_topic_0272789326_p39711454155520"></a>获取方法</p>
</th>
</tr>
</thead>
<tbody><tr id="row4288175712492"><td class="cellrowborder" valign="top" width="28.88%" headers="mcps1.2.5.1.1 "><p id="p5857163214118"><a name="p5857163214118"></a><a name="p5857163214118"></a>mindformers-<em id="i979747912"><a name="i979747912"></a><a name="i979747912"></a>{version}</em>-py3-none-any.whl</p>
</td>
<td class="cellrowborder" valign="top" width="13.96%" headers="mcps1.2.5.1.2 "><p id="p72691016660"><a name="p72691016660"></a><a name="p72691016660"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="30.42%" headers="mcps1.2.5.1.3 "><p id="p8288195784912"><a name="p8288195784912"></a><a name="p8288195784912"></a><span id="ph560917053119"><a name="ph560917053119"></a><a name="ph560917053119"></a>MindSpore</span> Transformers套件，构建大模型训练、微调、评估、推理、部署的全流程开发套件。<span id="ph13894758121219"><a name="ph13894758121219"></a><a name="ph13894758121219"></a>MindSpore</span>的master版本请使用r0.3分支代码版本。</p>
</td>
<td class="cellrowborder" valign="top" width="26.740000000000002%" headers="mcps1.2.5.1.4 "><p id="p528965710494"><a name="p528965710494"></a><a name="p528965710494"></a><a href="https://gitee.com/mindspore/mindformers/tree/r0.3/" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="zh-cn_topic_0272789326_row1997115417555"><td class="cellrowborder" valign="top" width="28.88%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0272789326_p897155412550"><a name="zh-cn_topic_0272789326_p897155412550"></a><a name="zh-cn_topic_0272789326_p897155412550"></a><span id="ph5948195051914"><a name="ph5948195051914"></a><a name="ph5948195051914"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.96%" headers="mcps1.2.5.1.2 "><p id="p13269191612612"><a name="p13269191612612"></a><a name="p13269191612612"></a>是</p>
</td>
<td class="cellrowborder" valign="top" width="30.42%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0272789326_p19971115435517"><a name="zh-cn_topic_0272789326_p19971115435517"></a><a name="zh-cn_topic_0272789326_p19971115435517"></a>制作镜像需要。</p>
</td>
<td class="cellrowborder" valign="top" width="26.740000000000002%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0272789326_p179726546557"><a name="zh-cn_topic_0272789326_p179726546557"></a><a name="zh-cn_topic_0272789326_p179726546557"></a>用户根据业务自行准备。</p>
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
>本章节以Ubuntu操作系统为例。

**操作步骤<a name="section173381914413"></a>**

1.  以**root**用户登录服务器。
2.  将准备的软件包MindFormers源码上传到服务器任意目录（如“/home/test“）。
3.  执行以下步骤准备Dockerfile文件。
    1.  进入软件包所在目录，执行以下命令创建Dockerfile文件（文件名示例“Dockerfile“）。

        ```
        vi Dockerfile
        ```

    2.  请参见[Dockerfile](#zh-cn_topic_0272789326_li104026527188)编写示例，将内容写入Dockerfile文件后执行**:wq**命令保存内容。

4.  进入软件包所在目录，执行以下命令，构建容器镜像，**注意不要遗漏命令结尾的**“.“。

    ```
    docker build -t  [OPTIONS] 镜像名_系统架构:镜像tag .
    ```

    例如：

    ```
    docker build -t test_train_arm64:v1.0 .
    ```

    命令解释如[表2](#zh-cn_topic_0272789326_zh-cn_topic_0256378845_table47051919193111)所示。

    **表 2**  命令参数说明

    <a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_table47051919193111"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_row77069193317"><th class="cellrowborder" valign="top" width="40%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p17061819143111"><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p17061819143111"></a><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p17061819143111"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="60%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p127066198319"><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p127066198319"></a><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p127066198319"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_row370601913312"><td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p5706161915311"><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p5706161915311"></a><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p5706161915311"></a>-t</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p8706119153115"><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p8706119153115"></a><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p8706119153115"></a>指定镜像名称。</p>
    </td>
    </tr>
    <tr id="row829312195610"><td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.3.1.1 "><p id="p9595273564"><a name="p9595273564"></a><a name="p9595273564"></a><em id="i17523236244"><a name="i17523236244"></a><a name="i17523236244"></a>OPTIONS</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.3.1.2 "><p id="p1265113317560"><a name="p1265113317560"></a><a name="p1265113317560"></a><span class="parmvalue" id="parmvalue07316106516"><a name="parmvalue07316106516"></a><a name="parmvalue07316106516"></a>“--disable-content-trust”</span>选项：忽略校验，默认开启。出于安全考虑，这里推荐设置关闭。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_row15532335367"><td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p18119431094"><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p18119431094"></a><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p18119431094"></a><em id="i611144218574"><a name="i611144218574"></a><a name="i611144218574"></a>镜像名</em><em id="i1311164225713"><a name="i1311164225713"></a><a name="i1311164225713"></a>_系统架构:</em><em id="i1711113429571"><a name="i1711113429571"></a><a name="i1711113429571"></a>镜像tag</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p115321037368"><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p115321037368"></a><a name="zh-cn_topic_0272789326_zh-cn_topic_0256378845_p115321037368"></a>镜像名称与标签，请用户根据实际情况写入。</p>
    </td>
    </tr>
    </tbody>
    </table>

    当出现“Successfully built xxx“表示镜像构建成功。

5.  构建完成后，执行以下命令查看镜像信息。

    ```
    docker images
    ```

    回显示例如下。

    ```
    REPOSITORY                TAG                 IMAGE ID            CREATED             SIZE
    test_train_arm64          v1.0                d82746acd7f0        27 minutes ago      749MB
    ```

**编写示例<a name="zh-cn_topic_0272789326_section3523631151714"></a>**

使用过程中请根据实际情况修改软件包版本及架构。

1.  <a name="zh-cn_topic_0272789326_li104026527188"></a>Dockerfile编写示例。

    -   Ubuntu  ARM系统Dockerfile示例。

        ```
        FROM xxx # 基础训练镜像 
        ARG MINDFORMERS_PKG=mindformers-{version}-py3-none-any.whl
        
        WORKDIR /tmp 
        COPY . ./ 
         
        ENV http_proxy xxx
        ENV https_proxy xxx
        
        # 配置Python pip源 
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf
         
        # 安装MindFormers
        RUN pip install $MINDFORMERS_PKG
        
         
        ENV http_proxy "" 
        ENV https_proxy "" 
        
        ```

    -   Ubuntu  x86\_64系统Dockerfile示例。

        ```
        FROM xxx # 基础训练镜像 
        ARG MINDFORMERS_PKG=mindformers-{version}-py3-none-any.whl
        
        WORKDIR /tmp 
        COPY . ./ 
         
        ENV http_proxy xxx
        ENV https_proxy xxx
        
        # 配置Python pip源 
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf
        
        # 安装MindFormers
        RUN pip install $MINDFORMERS_PKG
        
         
        ENV http_proxy "" 
        ENV https_proxy "" 
        
        ```

    为了使Dockerfile更加安全，用户可以根据业务在其中定义HEALTHCHECK检查。通过在容器内部运行**HEALTHCHECK** _\[OPTIONS\]_ **CMD**命令来检查容器的运行状况。


#### 脚本适配<a name="ZH-CN_TOPIC_0000002479387132"></a>

本章节提供了故障恢复脚本适配示例。用户请根据实际情况选择对应的脚本适配示例。

-   ResNet50模型适配
    -   [基于PyTorch的故障恢复](#section72859254718)
    -   [基于MindSpore的故障恢复](#section127532091511)
    -   [基于TensorFlow的故障恢复](#section2352206112211)

-   Pangu\_alpha模型适配（MindSpore框架）

    [基于Pangu\_alpha的故障恢复示例](#section1844516123710)

>[!NOTE] 说明 
>下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。

**PyTorch的故障恢复示例<a name="section72859254718"></a>**

1.  <a name="li14102111234717"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50\_ID4149\_for\_PyTorch”作为训练代码。
2.  自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

        回显示例如下：

        ```
        /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# du -sh
        ```

        回显示例如下：

        ```
        11G
        ```

4.  将[1](#li14102111234717)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50\_ID4149\_for\_PyTorch“目录上传至环境，如“/data/atlas\_dls/public/code/”目录。
5.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-rescheduling/withRanktable/pytorch/resnet50“目录中的train\_start.sh、utils.sh和rank\_table.sh文件，在训练代码中创建“scripts“目录，在管理节点构造如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts/#
    scripts/
    ├── rank_table.sh
    ├── utils.sh
    └── train_start.sh
    ```

6.  在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch“路径下修改main.py代码，修改以下加粗内容，改动内容涉及模型保存和加载的逻辑调整。

    ```
    import argparse
    import glob
    import os
    ...
        if args.resume:
            candidate_ckpt_path = ""
            for p in glob.glob(f"./rank*"):
                best_ckpt_path = os.path.join(p, "model_best.pth.tar")
                if os.path.exists(best_ckpt_path):
                    candidate_ckpt_path = best_ckpt_path
                    break
            if candidate_ckpt_path:
                print("[gpu id:", args.gpu, "]", "=> loading checkpoint '{}'".format(candidate_ckpt_path))
                # Map model to be loaded to specified single npu.
                loc = 'npu:{}'.format(args.gpu)
                checkpoint = torch.load(candidate_ckpt_path, map_location=loc)
                print(f"load checkpoint to : {loc}")
                args.start_epoch = checkpoint['epoch']
                best_acc1 = checkpoint['best_acc1']
                model.load_state_dict(checkpoint['state_dict'])
                optimizer.load_state_dict(checkpoint['optimizer'])
                print("[gpu id:", args.gpu, "]", "=> loaded checkpoint '{}' (epoch {})".format(candidate_ckpt_path, checkpoint['epoch']))
            else:
                print("no valid ckpt found to resume.")
    ...
            if not args.multiprocessing_distributed or (args.multiprocessing_distributed and args.rank % ngpus_per_node == 0):
                save_path = f"./rank_{args.rank}"
                if not os.path.exists(save_path):
                    os.makedirs(save_path, exist_ok=True)
                save_checkpoint({
                    'epoch': epoch + 1,
                    'arch': args.arch,
                    'state_dict': model.state_dict(),
                    'best_acc1': best_acc1,
                    'optimizer': optimizer.state_dict(),
                }, is_best, save_path=save_path)
    ...
    ...
    # 修改原有save_checkpoint函数
    def save_checkpoint(state, is_best, filename='checkpoint.pth.tar', save_path="./"):
        if is_best:
            target_path = os.path.join(save_path, 'model_best.pth.tar')
            torch.save(state, target_path)
            print(f"save ckpt to {target_path} done. Best epoch for now is :{state['epoch']}")
    ```

**MindSpore的故障恢复示例<a name="section127532091511"></a>**

1.  下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/cv/ResNet)中master分支代码，将“models/official/cv/ResNet“目录重命名为“resnet”并作为训练代码。
2.  执行以下命令，在管理节点创建代码目录，并上传训练代码到该目录。

    ```
    mkdir /data/atlas_dls/code
    ```

3.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/resnet50“目录中的“train\_start.sh“和“main.sh“文件，结合训练代码中“resnet/scripts“目录，在管理节点构造如下的目录结构。

    ```
    root@ubuntu:/data/atlas_dls/public/code/resnet/scripts/#
    scripts/
    ├── main.sh
     ...
    ├── run_distribute_train.sh
    ├── run_distribute_train_gpu.sh
    └── train_start.sh
    ```

4.  修改“/data/atlas\_dls/public/code/resnet/scripts“目录下的“train\_start.sh“文件。

    1.  将“dataset\_path“修改为容器内实际的数据集目录。
    2.  “config\_yaml\_path“修改为容器内实际的配置文件路径。

    ```
    根据实际情况进行修改，全局配置参数：数据集路径，配置参数文件路径；其他模型适配，请根据实际情况增删参数。
    dataset_path=/job/data/imagenet/train
    config_yaml_path=/job/code/resnet/resnet50_imagenet2012_config.yaml
    ```

    train\_start.sh脚本通过调用main.sh脚本启动训练任务。在适配其他模型时，请根据其训练启动脚本（本示例为train.py）的使用指导，调整main.sh脚本中的环境变量配置、启动脚本路径、启动脚本参数。

    ```
    # main.sh: 针对本示例（ResNet50模型），用户不需要再修改此脚本；其他模型适配，请根据实际情况，增、删或修改环境变量配置，然后修改训练启动脚本路径和对应的参数，即main.sh脚本中Python命令调用的部分。
    # 本例中，单机单卡的Python命令如下：
    python ${ROOT_PATH}/../train.py --data_path=${DATA_PATH} --config_path=${CONFIG_PATH} 
    # 本例中，单机多卡和分布式的命令如下：
    python ${ROOT_PATH}/../train.py --run_distribute=True --device_num=${RANK_SIZE} --data_path=${DATA_PATH} --config_path=${CONFIG_PATH} 
    ```

5.  修改“/data/atlas\_dls/public/code/resnet/config/“目录的配置文件“resnet50\_imagenet2012\_config.yaml“。模型保存和加载设置，图编译保存和加载设置。

    ```
    ...
    run_distribute: False
    enable_profiling: False
    data_path: "/cache/data"
    output_dir: "/job/code/output" # 修改checkpoint保存路径，请用户根据实际情况进行修改
    load_path: "/cache/checkpoint_path/"
    device_target: "Ascend"
    checkpoint_path: "./checkpoint/"
    checkpoint_file_path: ""
    ...
    net_name: "resnet50"
    dataset: "imagenet2012"
    device_num: 1
    pre_trained: "/job/code/output/resnet50/imagenet2012/ckpt" # 容器内预训练模型加载路径（支持目录和文件），支持在指定路径下对.ckpt文件进行模糊查找，将搜寻最新的.ckpt文件进行加载，请用户参考训练YAML根据实际情况进行修改
    run_eval: False
    eval_dataset_path: ""
    parameter_server: False
    filter_weight: False
    save_best_ckpt: True
    eval_start_epoch: 40
    ...
    network_dataset: "resnet50_imagenet2012"
    
    
    # 再训练选项 
    save_graphs: False  # 是否开启图编译结果保存
    save_graphs_path: "./graphs" # 图编译结果保存路径
    has_trained_epoch: 0 # 模型预训练的epoch，默认是0
    has_trained_step: 0 # 模型预训练的step，默认是0
    ---
    # 每项配置的帮助说明
    enable_modelarts: "Whether training on modelarts, default: False"
    ...
    batch_size: "Batch size for training and evaluation"
    epoch_size: "Total training epochs."
    checkpoint_path: "The location of the checkpoint file."
    checkpoint_file_path: "The location of the checkpoint file."
    save_graphs: "Whether save graphs during training, default: False."
    save_graphs_path: "Path to save graphs."
    ```

6.  resnet代码的启动脚本为“train.py“，检查“train.py“中是否存在保存CheckPoint的代码，示例代码如下。

    -   如果存在，则跳过本步骤。
    -   如果不存在，则补充以下保存CheckPoint的代码样例，其中所用参数需要用户在配置文件中定义和设置。其他模型适配，请参考如下片段，根据启动脚本具体内容，添加保存CheckPoint的代码。如有需要，请参考[MindSpore官网](https://www.mindspore.cn/)教程进行修改。

    ```
    ...
        # 模型保存代码
        if config.save_checkpoint:
            ckpt_append_info = [{"epoch_num": 0, "step_num": 0}]
            config_ck = CheckpointConfig(save_checkpoint_steps=config.save_checkpoint_epochs * step_size,
                                         keep_checkpoint_max=config.keep_checkpoint_max,
                                         append_info=ckpt_append_info)
            ckpt_cb = ModelCheckpoint(prefix=config.net_name, directory=config.save_ckpt_dir+"_"+str(config.rank_id), config=config_ck)
            cb += [ckpt_cb]
    ...
    ```

7.  resnet代码的启动脚本为train.py，检查train.py中是否存在加载checkpoint的代码，如果存在，则执行配置完成，进行下一章节操作；否则执行[8](#li1621315181018)。
8.  <a name="li1621315181018"></a>在train.py中补充加载checkpoint的代码。以下为checkpoint加载样例，其中所用参数需要用户在配置文件中定义和设置。其他模型适配，请参考如下片段，根据启动脚本具体内容，添加加载checkpoint的代码。如有需要，请参考[MindSpore官网](https://www.mindspore.cn/)教程进行修改。
    1.  修改“src/utils.py“，添加读取epoch代码，加载CKPT后，训练日志中将从CKPT保存时刻所处的epoch开始打印。

        ```
        ...
        def init_weight(net, cfg):
            """init_weight"""
            if cfg.pre_trained:
                if not os.path.isfile(cfg.pre_trained):
                    cfg.logger.warning("There is not ckpt file: %s", cfg.pre_trained)
                else:
                    param_dict = ms.load_checkpoint(cfg.pre_trained)
                    if cfg.filter_weight:
                        filter_list = [x.name for x in net.end_point.get_parameters()]
                        filter_checkpoint_parameter_by_list(param_dict, filter_list)
                    ms.load_param_into_net(net, param_dict)
                    cfg.start_epoch = int(param_dict.get('epoch_num', ms.Tensor(0, ms.int32)).asnumpy().item())
                    cfg.logger.info("Pre trained ckpt mode: %s loading", cfg.pre_trained)
        ...
        ```

    2.  修改train.py，替换原有的init\_weight函数，使用\_try\_to\_init\_weight尝试加载CKPT文件，避免出现加载到不完整的CKPT，导致训练报错的问题。

        ```
        import glob
        ...
        # 找寻pre_trained目录下最新的*.ckpt文件
        def _find_latest_ckpt():
            ckpt_files = glob.glob(config.pre_trained+"*/*.ckpt")
            if ckpt_files:
                ckpt_files.sort(key=os.path.getmtime, reverse=True)
            return ckpt_files
        
        # 尝试加载CKPT文件，尝试次数为INIT_WEIGHT_MAX_ATTEMPTS次
        def _try_to_init_weight(net, config):
            if os.path.isfile(config.pre_trained):
                latest_ckpt = [config.pre_trained]
            else:
                latest_ckpt = _find_latest_ckpt()
        
            if not latest_ckpt:
                config.logger.warning("There is not ckpt file: %s", config.pre_trained)
                return
        
            init_weight_attempts = 0
            INIT_WEIGHT_MAX_ATTEMPTS = 5
            while(latest_ckpt and init_weight_attempts < INIT_WEIGHT_MAX_ATTEMPTS): 
                try:
                    config.pre_trained = latest_ckpt[0]
                    init_weight(net, config)
                    break
                except Exception:
                    config.logger.warning("Pre trained ckpt %s format is incorrect, try to load the last most recent ckpt", config.pre_trained)
                    if latest_ckpt[1:]:
                        latest_ckpt = latest_ckpt[1:]
                        init_weight_attempts+=1
                        continue
                    else:
                        config.logger.error("no more ckpt to load", config.pre_trained)
                        raise ValueError("ckpt format is incorrect, no more ckpt to load, load ckpt failed.")
        
        ...
        @moxing_wrapper()
        def train_net():
            """train net"""
            target = config.device_target
            set_parameter()
            set_output_dir(config)
            config.logger = get_logger(config.log_dir, config.rank_id, config.parameter_server)
            dataset = create_dataset(dataset_path=config.data_path, do_train=True,
                                     batch_size=config.batch_size, train_image_size=config.train_image_size,
                                     eval_image_size=config.eval_image_size, target=target,
                                     distribute=config.run_distribute)
            step_size = dataset.get_dataset_size()
            net = resnet(class_num=config.class_num)
            if config.parameter_server:
                net.set_param_ps()
            # 替换原有的init_weight函数，使用_try_to_init_weight尝试加载CKPT文件，避免加载到不完整的CKPT，导致训练报错
            _try_to_init_weight(net, config)
        
            if config.resume_ckpt:
                resume_param = ms.load_checkpoint(config.resume_ckpt,
                                                  choice_func=lambda x: not x.startswith(('learning_rate', 'global_step')))
                config.start_epoch = int(resume_param.get('epoch_num', ms.Tensor(0, ms.int32)).asnumpy().item())
            lr = ms.Tensor(init_lr(step_size=step_size))
        ...
        ```

**TensorFlow的故障恢复示例<a name="section2352206112211"></a>**

1.  <a name="li360413424258"></a>下载[TensorFlow代码仓](https://gitee.com/ascend/ModelZoo-TensorFlow/tree/master/TensorFlow2/built-in/cv/image_classification/ResNet50_ID0360_for_TensorFlow2.X)中master分支中的“ResNet50\_ID0360\_for\_TensorFlow2.X”作为训练代码，请根据该模型代码TensorFlow版本选择训练镜像中的TensorFlow版本包。
2.  管理员用户上传数据集到存储节点。
    1.  进入“/data/atlas\_dls/public“目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet\_TF“。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# pwd
        /data/atlas_dls/public/dataset/resnet50/imagenet_TF
        ```

    2.  执行**du -sh**命令，查看数据集大小。

        ```
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet_TF# du -sh
        42G
        ```

3.  在本地解压[1](#li360413424258)中下载的训练代码，将“ModelZoo-TensorFlow-master/TensorFlow2/built-in/cv/image\_classification/“下的“ResNet50\_ID0360\_for\_TensorFlow2.X“目录重命名为“ResNet50\_for\_TensorFlow\_2.6\_code/“目录。
4.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/basic-training/ranktable“目录中的train\_start.sh、utils.sh和rank\_table.sh文件，在训练代码中创建“scripts“目录，在管理节点构造如下的目录结构。

    ```
    /data/atlas_dls/public/code/ResNet50_for_TensorFlow_2.6_code/
    ├──  scripts
    │   ├──  train_start.sh
    │   ├──  utils.sh
    │   ├──  rank_table.sh
    │    ...
    ```

5.  修改训练代码。补充加载CKPT文件时的日志打印。修改"tensorflow/tf2\_common/training/controller.py"。

    ```
    class Controller(object):
      """Class that facilitates training and evaluation of models."""
      def __init__(
        ...
        # Restore Model if needed.
        if self.checkpoint_manager is not None:
          model_restored = self._restore_model()
          logging.info("loading checkpoint %s", model_restored)
          if not model_restored and self.checkpoint_manager.checkpoint_interval:
            # If the model is not restored from a checkpoint, save an initial
            # checkpoint.
            ckpt_path = self.checkpoint_manager.save(
                checkpoint_number=self.global_step)
            logging.info("Saved checkpoints in %s", ckpt_path)
        # Create and initialize the interval triggers.
        self.eval_trigger = utils.IntervalTrigger(self.eval_interval,
                                                  self.eval_offset)
    ```

**Pangu\_alpha模型适配示例<a name="section1844516123710"></a>**

1.  下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/nlp/Pangu_alpha)中master分支代码，将“models/official/nlp/Pangu\_alpha“目录重命名为“pangu\_alpha”并作为训练代码，使用该版本模型脚本需保证在镜像中安装的MindSpore版本不低于2.0.0，并且安装mindformers组件。
2.  执行以下命令，在管理节点创建代码目录。

    ```
    mkdir /data/atlas_dls/code
    ```

3.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/pangu\_alpha“目录中的“train\_start.sh“和“main.sh“文件，结合训练代码中“pangu\_alpha/scripts“目录，在管理节点构造如下的目录结构。对于盘古百亿模型，使用“samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/pangu\_alpha\_13B“目录中的对应文件。

    ```
    root@ubuntu:/data/atlas_dls/code/pangu_alpha/scripts/# 
    scripts/
    ├── main.sh
    ├── run_cluster_export.sh
    ├── run_distribute_eval_gpu.sh
    ├── run_distribute_eval.sh
     ...
    ├── run_distribute_train.sh
    ├── run_standalone_eval.sh
    ├── run_standalone_export.sh
    ├── run_standalone_predict.sh
    └── train_start.sh
    ```

4.  修改“/data/atlas\_dls/code/pangu\_alpha/scripts“目录下的“train\_start.sh“文件，将“dataset“修改为容器内实际的数据集目录。

    ```
    ...
    # 训练数据集路径，根据实际情况修改
    # 安全提示，涉及对路径和输入参数的校验
    dataset="/job/data/train_data"
    
    # 设置训练环境变量
    set_env
    
    # 单节点训练场景
    if [[ "$server_count" == "1" ]]; then
        server_id=0
        if [ ${device_count} -lt 8 ]; then
            echo "Less than 8 card training is not supported for pangu alpha model." | tee log
        fi
        if [ ${device_count} -eq 8 ]; then
            bash main.sh ${device_count} ${server_count} ${RANK_TABLE_FILE} ${server_id} ${dataset}
        fi
    
    # 分布式训练场景
    else
        server_id=$(get_server_id)
        if [ $? -eq 1 ];then
            echo "get server id failed."
            exit 1
        fi
        echo "server id is: "${server_id}
        bash main.sh ${device_count} ${server_count} ${RANK_TABLE_FILE} ${server_id} ${dataset}
    
    ```

5.  百亿及以下模型可跳过该步骤。训练千亿模型时，期望恢复时间小于5min，需要进行额外脚本适配。下文以[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/nlp/Pangu_alpha)中pangu\_alpha的master分支为例（**已完成弹性训练任务配置和脚本适配**）。
    1.  修改“src/pangu\_alpha\_config.py“文件，主要涉及三个参数的更改：args\_opt.num\_layers、args\_opt.stage\_num、args\_opt.micro\_size。

        ```
        def set_parse_200B(args_opt):
            """
                Set config for 200B mode
            """
            args_opt.embedding_size = 16384
            args_opt.num_layers = 32                 # 模型层次
            args_opt.num_heads = 128
            if args_opt.per_batch_size == 0:
                args_opt.per_batch_size = 1
            args_opt.word_emb_dp = 0
            if args_opt.run_type == "train":
                args_opt.start_lr = 6e-5
                args_opt.end_lr = 6e-6
               args_opt.stage_num = 8               # 流水线阶段的数量
               args_opt.micro_size = 16             # 流水线并行模式下的微批次大小，其取值应大于args_opt.stage_num
                args_opt.op_level_model_parallel_num = 16
                if args_opt.optimizer_shard = 1:
                    args_opt.op_level_model_parallel_num = 8
            elif args_opt.run_type == "predict":
                args_opt.stage_num = 4
                args_opt.micro_size = 1
                args_opt.op_level_model_parallel_num = 16
                if args_opt.optimizer_shard == 1:
                    args_opt.op_level_model_parallel_num = 8
        ```

    2.  此外，需要指定或者直接修改“src/utils.py“中的“micro\_batch\_interleaved“参数为“1“（请参考“train.py“脚本的“run\_train\_pipeline”函数中“stage\_device\_num”、“data\_parallel\_num”、“batch\_size”、“micro\_batch\_interleaved”之间的计算关系。最终结果需要满足“PanguAlphaConfig”的“batch\_size”值是“TransformerOpParallelConfig”的“data\_parallel”的倍数）。

6.  pangu代码的启动脚本为train.py，检查train.py中是否存在保存CheckPoint的代码，代码示例如下。

    -   如果存在，则跳过本步骤。
    -   如果不存在，则补充以下保存CheckPoint的代码样例，其中所用参数可参照[9](#li13178638874)在配置文件“src/utils.py“中定义和设置。

    ```
    ...
    
        # 保存CheckPoint的代码调用
        add_checkpoint_callback_policy(args_opt, callback, rank)
    ...
    # 保存checkpoint代码定义
    def add_checkpoint_callback_policy(args_param, callback, rank_id):
        r"""
        Add checkpoint policy to callback.
        """
        # 安全提示，涉及对路径和输入参数的校验
        if args_param.save_checkpoint:
            # checkpoint保存epoch_num和step_num info信息
            ckpt_append_info = [{"epoch_num": args_param.has_trained_epoches, "step_num": args_param.has_trained_steps}]
            ckpt_config = CheckpointConfig(save_checkpoint_steps=args_param.save_checkpoint_steps,
                                           keep_checkpoint_max=args_param.keep_checkpoint_max,
                                           integrated_save=False,
                                           append_info=ckpt_append_info
                                           )
    
    
            ckpoint_cb = ModelCheckpoint(prefix=args_param.ckpt_name_prefix + str(rank_id),
                                         directory=os.path.join(args_param.save_checkpoint_path, f"rank_{rank_id}"),
                                         config=ckpt_config)
    
    
            callback.append(ckpoint_cb)
    ...
    ```

7.  pangu代码的启动脚本为train.py，检查train.py中是否存在加载checkpoint的代码，如果存在，则执行[10](#li6181138370)；否则执行[8](#li12175938673)。
8.  <a name="li12175938673"></a>在train.py中补充加载checkpoint的代码。以下为checkpoint加载样例，存在部分加载checkpoint的代码，需要添加弹性训练特性相关checkpoint加载代码，其中所用参数可参照[9](#li13178638874)在配置文件“src/utils.py“中定义和设置。

    ```
    ...
    # 如果运行的模型没有开启pipeline并行，则修改在以下函数
    def set_parallel_context(args_opt):
    # 如果运行的模型开启pipeline并行，则修改在以下函数
    # 安全提示，涉及对路径和输入参数的校验
    def set_pipeline_parallel_context(args_opt):
    # 在mindspore.set_auto_parallel_context前添加以下代码，请参考[MindSpore文档分布式并行接口说明](https://www.mindspore.cn/tutorials/experts/zh-CN/r2.0/index.html)对set_auto_parallel_context参数的使用说明
            
             
            # 弹性训练中增加内容
            if not os.path.exists(args_opt.strategy_load_ckpt_path):
                args_opt.strategy_load_ckpt_path = ""
    
            # 弹性训练中增加内容，strategy_ckpt_save_file_path参数可以根据容器内路径指定
            strategy_ckpt_save_file_path = '/job/data/code/fault_torlence/pangu_alpha/strategy.ckpt' 
            if args_opt.strategy_load_ckpt_path == strategy_ckpt_save_file_path:
                 strategy_ckpt_save_file_path = '/job/data/code/fault_torlence/pangu_alpha/strategy_new.ckpt'
     
            # 将strategy_ckpt_save_file='strategy.ckpt'修改成strategy_ckpt_save_file=strategy_ckpt_save_file_path，如果set_auto_parallel_context里没有指定strategy_ckpt_save_file参数，则需要手动添加strategy_ckpt_save_file=strategy_ckpt_save_file_path，如下粗体所示
            mindspore.set_auto_parallel_context(
                parallel_mode=args_opt.parallel_mode, gradients_mean=False, search_mode=args_opt.search_mode,
                full_batch=bool(args_opt.full_batch), loss_repeated_mean=True,
                device_num=device_num, enable_parallel_optimizer=bool(args_opt.optimizer_shard),
                pipeline_stages=args_opt.stage_num, enable_alltoall=bool(args_opt.enable_alltoall),
                strategy_ckpt_save_file=strategy_ckpt_save_file_path)
           
    ...
    ...
    # checkpoint加载代码定义
    # 安全提示，涉及对路径和输入参数的校验
    def restore_checkpoint(args_param, sink_size, dataset, model, network, epoch):
        r"""
        Load checkpoint process.
        """
        print("======start single checkpoint", flush=True)
        ckpt_name = args_param.ckpt_name_prefix
        # 为了文档简洁易读, 此处省略了命令行参数save_checkpoint_path和ckpt_name的校验, 请用户自行添加相关校验
        ckpt_pattern = os.path.join(args_param.save_checkpoint_path, "rank_{}".format(D.get_rank()),
                                    f"{ckpt_name}*.ckpt")
        ckpt_all_files = glob.glob(ckpt_pattern)
        if not ckpt_all_files:
            print(f"There is no ckpt file in {args_param.save_checkpoint_path}, "
                  f"current ckpt_files found is {ckpt_files} "
                  f"with pattern {ckpt_pattern}, so skip the loading.")
            return
        ckpt_exp_pattern = os.path.join(
            args_param.save_checkpoint_path,
            "rank_{}".format(D.get_rank()),
            f"{ckpt_name}*_breakpoint.ckpt",
        )
        ckpt_exp_files = glob.glob(ckpt_exp_pattern)
        ckpt_files = []
        for file in ckpt_all_files:
            if file not in ckpt_exp_files:
                ckpt_files.append(file)
    
        if not ckpt_files:
            print(
                f"There is no ckpt file in {args_param.save_checkpoint_path}, "
                f"current ckpt_files found is {ckpt_files} "
                f"with pattern {ckpt_pattern}, so skip the loading."
            )
            return
        ckpt_files.sort(key=os.path.getmtime, reverse=True)
        time_stamp = datetime.datetime.now()
        print(f"time stamp {time_stamp.strftime('%Y.%m.%d-%H:%M:%S')} pre trained ckpt model {ckpt_files} loading",
              flush=True)
        # 加载checkpoint最新文件
        print(f'Start to load from {ckpt_files[0]}')
        param_dict = load_checkpoint(ckpt_files[0])
        if param_dict.get("epoch_num") and param_dict.get("step_num"):
            args_param.has_trained_epoches = int(param_dict["epoch_num"].data.asnumpy())
            args_param.has_trained_steps = int(param_dict["step_num"].data.asnumpy())
        model.build(train_dataset=dataset, sink_size=sink_size, epoch=epoch)
        load_param_into_net(network, param_dict)
    ...
    ```

9.  <a name="li13178638874"></a>修改“src/utils.py“文件中的参数。

    ```
    ...
        opt.add_argument("--vocab_size",
                          type=int,
                          default=50304, # 根据训练数据集进行修改，此处已修改为样例数据集的取值
                          help="vocabulary size, default is 40000.")
    ...
        opt.add_argument("--data_column_name",
                         type=str,
                         default="text", # 根据数据集定义的字段进行修改，此处已修改为样例数据集的取值
                         help="Column name of datasets")
    ...
        parser.add_argument("--strategy_load_ckpt_path",
                            type=str,
                            default="/job/data/code/fault_torlence/pangu_alpha/strategy/strategy.ckpt", # 弹性训练中，根据用户习惯指定容器内路径，且路径不会被训练覆盖。
                            help="The training prallel strategy for the model.")
        parser.add_argument("--tokenizer_path",
                            type=str,
                            default="./tokenizer_path",
                            help="The path where stores vocab and vocab model file")
    ...
    def add_retrain_params(opt):
        """
        Add parameters about retrain.
        """
        opt.add_argument("--pre_trained",
                         type=str,
                         default="/job/data/code/fault_torlence/pangu_alpha/8p", # 指定预训练模型路径，
                         help="Pretrained checkpoint path.")
        opt.add_argument("--save_checkpoint_path",  
                         type=str,
                         default="/job/data/code/fault_torlence/pangu_alpha/8p",   # 指定模型保存路径
                         help="Save checkpoint path.")
        opt.add_argument("--keep_checkpoint_max", # 指定模型保存策略：最大数量
                         type=int,
                         default=1,
                         help="Max checkpoint save number.")
        opt.add_argument("--save_checkpoint_steps", # 指定模型保存策略：保存间隔
                         type=int,
                         default=20,
                         help="Save checkpoint step number.")
        opt.add_argument("--save_checkpoint", # 指定当次训练是否保存模型
                         type=ast.literal_eval,
                         default=True,
                         help="Whether save checkpoint in local disk.")
        opt.add_argument("--ckpt_name_prefix", # 指定模型保存策略：文件名前缀
                         type=str,
                         default="pangu",
                         help="Saving checkpoint name prefix.")
    ...
    ```

10. <a name="li6181138370"></a>在“/data/atlas\_dls/code/pangu\_alpha“目录下构建空文件“group\_info\_env“。

    ```
    root@ubuntu:/data/atlas_dls/code/pangu_alpha/# 
    pangu_alpha/
    ├── README.md
    ├── README_CN.md
    ├── group_info_env
     ...
    ├── scripts
    ├── serving_increment
    ├── src
    ├── tasks.py
    └── train.py
    ```

11. 修改train.py文件中的“group\_info\_env“路径。

    ```
    ...
        # env variable prepare
        group_info_file = os.getenv("GROUP_INFO_FILE")
        if group_info_file:
            with open(os.path.expanduser("/job/code/group_info_env"), "a") as outfile:
                outfile.write(f"export GROUP_INFO_FILE_REFLECT={group_info_file}\n")
    ...
    ```


#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479227132"></a>

##### 选择YAML示例<a name="ZH-CN_TOPIC_0000002479387110"></a>

集群调度并未专门为用户提供弹性训练的YAML示例，用户可以获取断点续训的YAML并进行修改即可使用。

**表 1**  获取YAML

<a name="table194871213135113"></a>
<table><thead align="left"><tr id="row15488613165116"><th class="cellrowborder" valign="top" width="12.98259651930386%" id="mcps1.2.6.1.1"><p id="p1694212055114"><a name="p1694212055114"></a><a name="p1694212055114"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.2"><p id="p894252065116"><a name="p894252065116"></a><a name="p894252065116"></a>模型</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.3"><p id="p1694216209514"><a name="p1694216209514"></a><a name="p1694216209514"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.4"><p id="p8942122025113"><a name="p8942122025113"></a><a name="p8942122025113"></a>获取链接</p>
</th>
<th class="cellrowborder" valign="top" width="21.754350870174033%" id="mcps1.2.6.1.5"><p id="p7942172025115"><a name="p7942172025115"></a><a name="p7942172025115"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row10488181375112"><td class="cellrowborder" rowspan="4" valign="top" width="12.98259651930386%" headers="mcps1.2.6.1.1 "><p id="p9942132045117"><a name="p9942132045117"></a><a name="p9942132045117"></a>Volcano Job</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.2 "><p id="p1394262015120"><a name="p1394262015120"></a><a name="p1394262015120"></a>ResNet50</p>
</td>
<td class="cellrowborder" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.3 "><p id="p10942152013510"><a name="p10942152013510"></a><a name="p10942152013510"></a>a800_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.4 "><p id="p9942102010518"><a name="p9942102010518"></a><a name="p9942102010518"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/train/basic-training/ranktable/yaml/910" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="21.754350870174033%" headers="mcps1.2.6.1.5 "><p id="p1694222015119"><a name="p1694222015119"></a><a name="p1694222015119"></a>示例默认为单机8卡任务</p>
<p id="p1161014614466"><a name="p1161014614466"></a><a name="p1161014614466"></a></p>
</td>
</tr>
<tr id="row20488131310512"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p92173245117"><a name="p92173245117"></a><a name="p92173245117"></a>a800_pytorch_vcjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p191773377514"><a name="p191773377514"></a><a name="p191773377514"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/train/resumable-training/fault-rescheduling/withRanktable/pytorch/resnet50/yamls/910/a800_pytorch_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row348851319516"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p32203210515"><a name="p32203210515"></a><a name="p32203210515"></a>a800_vcjob.yaml（MindSpore架构）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5177173765117"><a name="p5177173765117"></a><a name="p5177173765117"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/resnet50/yamls/a800_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1561116674613"><a name="p1561116674613"></a><a name="p1561116674613"></a>示例默认为单机单卡任务</p>
</td>
</tr>
<tr id="row16489613125118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p548911315117"><a name="p548911315117"></a><a name="p548911315117"></a>盘古</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19293215114"><a name="p19293215114"></a><a name="p19293215114"></a>a800_vcjob.yaml（MindSpore架构）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p317719373516"><a name="p317719373516"></a><a name="p317719373516"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/train/resumable-training/fault-rescheduling/withRanktable/mindspore/pangu_alpha/yamls/a800_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p156115614615"><a name="p156115614615"></a><a name="p156115614615"></a>示例默认为2*8卡任务</p>
</td>
</tr>
</tbody>
</table>


##### YAML参数说明<a name="ZH-CN_TOPIC_0000002479387134"></a>

本章节提供使用弹性训练配置YAML的操作示例。在具体操作前，用户需要了解相关YAML示例的参数说明，再进行操作。

**表 1**  YAML参数说明

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="27.21%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.230000000000004%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001609074269_p19658152414451"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001609074269_ul1531417539259"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001609074269_p11302326164814"></a>N为节点个数，Deployment类型的任务不需要该参数，该参数建议与replicas保持一致。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="row12812154533917"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p14813194563914"><a name="p14813194563914"></a><a name="p14813194563914"></a>maxRetry</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p6813134514395"><a name="p6813134514395"></a><a name="p6813134514395"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p10813124553917"><a name="p10813124553917"></a><a name="p10813124553917"></a><span id="ph41465155410"><a name="ph41465155410"></a><a name="ph41465155410"></a>Pod</span>删除重启次数，弹性训练需关闭<span id="ph1357741225415"><a name="ph1357741225415"></a><a name="ph1357741225415"></a>Pod</span>重启，需要设置为0。</p>
</td>
</tr>
<tr id="row917012162413"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p871672217415"><a name="p871672217415"></a><a name="p871672217415"></a>minReplicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p14170516144111"><a name="p14170516144111"></a><a name="p14170516144111"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p317081614417"><a name="p317081614417"></a><a name="p317081614417"></a>最小副本数，需要设置为任务需要的最小节点的数量。</p>
</td>
</tr>
<tr id="row20143651155812"><td class="cellrowborder" rowspan="5" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p19309122810151"><a name="p19309122810151"></a><a name="p19309122810151"></a>fault-scheduling</p>
<p id="p1430895155915"><a name="p1430895155915"></a><a name="p1430895155915"></a></p>
<p id="p7308145135911"><a name="p7308145135911"></a><a name="p7308145135911"></a></p>
<p id="p111494487596"><a name="p111494487596"></a><a name="p111494487596"></a></p>
<p id="p1190084975913"><a name="p1190084975913"></a><a name="p1190084975913"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p1058012817249"><a name="p1058012817249"></a><a name="p1058012817249"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1432806105912"><a name="p1432806105912"></a><a name="p1432806105912"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="ph0511135011612"><a name="ph0511135011612"></a><a name="ph0511135011612"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="ph2149141202811"><a name="ph2149141202811"></a><a name="ph2149141202811"></a>Pod</span>。</p>
</td>
</tr>
<tr id="row6949124320594"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p2032819617590"><a name="zh-cn_topic_0000001570873348_p2032819617590"></a><a name="zh-cn_topic_0000001570873348_p2032819617590"></a>force</p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" headers="mcps1.2.4.1.2 "><p id="p4784925400"><a name="p4784925400"></a><a name="p4784925400"></a>暂不支持。</p>
<div class="note" id="note65068564019"><a name="note65068564019"></a><a name="note65068564019"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p85061256401"><a name="p85061256401"></a><a name="p85061256401"></a>当前仅支持grace模式。</p>
</div></div>
</td>
</tr>
<tr id="row69001145205918"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p153287615911"><a name="zh-cn_topic_0000001570873348_p153287615911"></a><a name="zh-cn_topic_0000001570873348_p153287615911"></a>off</p>
</td>
</tr>
<tr id="row16149114820595"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p8667174914916"><a name="zh-cn_topic_0000001570873348_p8667174914916"></a><a name="zh-cn_topic_0000001570873348_p8667174914916"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="row2900144905917"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p4602135219491"><a name="zh-cn_topic_0000001570873348_p4602135219491"></a><a name="zh-cn_topic_0000001570873348_p4602135219491"></a>其他值</p>
</td>
</tr>
<tr id="row128861384219"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p11288121310421"><a name="p11288121310421"></a><a name="p11288121310421"></a>elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p7288191354217"><a name="p7288191354217"></a><a name="p7288191354217"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1628816134422"><a name="p1628816134422"></a><a name="p1628816134422"></a>开启弹性训练。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在准备训练镜像章节制作或者获取的镜像名称）。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a>ARM</span>环境：huawei-arm</p>
<p id="zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：huawei-x86</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001609074269_p1449413229314"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p7665323173618"><a name="p7665323173618"></a><a name="p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<p id="p533665619531"><a name="p533665619531"></a><a name="p533665619531"></a><span id="zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</p>
<a name="ul14200073713"></a><a name="ul14200073713"></a>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p155013815543"><a name="p155013815543"></a><a name="p155013815543"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>huawei.com/Ascend910</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>根据所使用芯片类型不同，取值如下：</p>
<div class="p" id="p106711919185315"><a name="p106711919185315"></a><a name="p106711919185315"></a><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</div>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p530216266485"><a name="zh-cn_topic_0000001609074269_p530216266485"></a><a name="zh-cn_topic_0000001609074269_p530216266485"></a>请求的NPU数量，请根据实际修改，请求整卡时不能再请求vNPU。</p>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p1294216211553"><a name="p1294216211553"></a><a name="p1294216211553"></a>Atlas 800 训练服务器（NPU满配）取值为：ascend-910</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>用于区分任务使用的芯片的类型。需要在<span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="row15462632114"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/AscendXXX']</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
</td>
</tr>
<tr id="row1149173454010"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><div class="note" id="note1031313718402"><a name="note1031313718402"></a><div class="notebody"><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>仅支持在<span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>新任务副本数范围为\[minReplicas, replicas\]，具体数值由当前集群中的可用节点数确定，多节点分布式训练时有效。


##### 配置YAML<a name="ZH-CN_TOPIC_0000002479227138"></a>

**操作步骤<a name="section6131855154814"></a>**

1.  将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    使用**弹性训练**特性，参考本配置。以a800\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用8个芯片，修改示例如下。

    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
    ...
      labels:
        ring-controller.atlas: ascend-910   # 标识任务使用的芯片的产品类型
    ...
    ---
    apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API
    kind: Job                               # 目前只支持Job类型
    metadata:
      name: mindx-dls-test                  # 任务名，可自定义
      labels:
        ring-controller.atlas: ascend-910    # 标识任务使用的芯片的产品类型
        fault-scheduling: "grace"        # 开启故障重调度
        elastic-scheduling: "on"          # 开启弹性训练，需添加""号
      annotations:
        minReplicas: "1"                 # 最小副本数
    ...
    spec:
      minAvailable: 1                  # 设置为1
    ...
      maxRetry: 0              #设置为0
    ...
      - name: "default-test"
          template:
            metadata:
    ...
            spec:
    ...
              env:
    ...
              - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
    ...
                resources:  
                  requests:
                    huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8
                  limits:
                    huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
    ...
                nodeSelector:
                  host-arch: huawei-arm       # 可选值，根据实际情况填写
    ...
    ```

2.  使用弹性训练功能，需要扩展内存，请按注释添加参数。此外还要使用“maxRetry”机制，示例如下。

    ```
    ...
              volumeMounts:                             #弹性训练扩容
              - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

3.  若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu“和“memory“参数和对应的参数值，具体数值请根据实际情况配置。

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

4.  修改训练脚本、代码的挂载路径。

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

5.  如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
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

    -   使用**MindSpore架构**的模型，包括ResNet50模型和Pangu\_alpha模型需要跳过此步骤。

6.  YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

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
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
    ```



#### 下发任务<a name="ZH-CN_TOPIC_0000002511427035"></a>

**操作步骤<a name="section12502215114011"></a>**

本章节以MindSpore框架的ResNet50模型为例，下发训练任务。

1.  登录管理节点，进入YAML文件所在路径。
2.  在管理节点执行以下命令，使用YAML下发训练任务。

    ```
    kubectl apply -f XXX.yaml
    ```

    例如：

    ```
    kubectl apply -f a800_vcjob.yaml
    ```

    回显如下：

    ```
    configmap/rings-config-mindx-dls-test created
    job.batch.volcano.sh/mindx-dls-test created
    ```

    >[!NOTE] 说明 
    >如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f** **_XXX__.yaml_命令删除原任务，再重新下发任务。


#### 查看任务进程<a name="ZH-CN_TOPIC_0000002479227140"></a>

训练任务下发成功后，训练任务就可正常运行。可通过如下内容查看训练任务运行情况。

**查看所有训练任务<a name="section181299581348"></a>**

查看当前节点上运行的所有训练任务，操作步骤如下。

1.  登录管理节点。
2.  执行以下命令，查看训练任务运行情况。

    ```
    kubectl get pods -A -o wide
    ```

    回显示例如下

    ```
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              1/1     Running   0          92s     192.168.70.118    ubuntu-155   <none>           <none>
    vcjob            mindx-dls-test-default-test-1              1/1     Running   0          92s     192.168.185.213   ubuntu-177   <none>           <none>
    ```


#### 查看结果<a name="ZH-CN_TOPIC_0000002479387114"></a>

##### 构造故障<a name="ZH-CN_TOPIC_0000002511347079"></a>

用户可以参考本章节构造故障。

**（可选）构造NPU芯片故障<a name="section182989331585"></a>**

通过断开NPU网络链路模拟的参数面网络故障。NPU网络故障不影响单机训练任务。用户在断开链路后需手动恢复，否则该故障会一直存在。

1.  登录计算节点。
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

4.  执行以下命令，恢复NPU网络链路故障。

    ```
    hccn_tool -i {device_id} -cfg recovery
    ```

5.  执行以下命令，查看NPU链路状态。

    ```
    hccn_tool -i {device_id} -net_health -g
    ```

    回显示例如下，表示NPU网络链路故障已经恢复。

    ```
    net health status: Success
    ```


##### 查看运行结果<a name="ZH-CN_TOPIC_0000002511427041"></a>

当节点发生故障时，Volcano会将该训练任务删除，Resilience Controller根据可用资源修改任务资源需求，Volcano调度到剩余可用资源上继续运行。

**弹性训练情况<a name="section55191324318"></a>**

1.  登录管理节点，执行以下命令查看训练任务运行情况。

    ```
    ~# kubectl get pods -A -o wide
    ```

    以全部资源为2节点16卡，下发2节点16卡任务为例，回显示例如下。该回显表示训练任务正常执行时的任务运行情况。

    ```
    NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              1/1     Running   0          47s     192.168.70.82   Node-1   <none>           <none>
    vcjob            mindx-dls-test-default-test-1              1/1     Running   0          47s     192.168.39.9    Node-2     <none>           <none>
    ……
    ```

2.  当Node-1发生NPU网络故障时，Volcano删除任务。执行以下命令查看训练任务终止情况。

    ```
     kubectl get pods -A -o wide
    ```

    回显示例如下，表示训练任务被删除。

    ```
    NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              0/1     Terminating   0          6m59s     192.168.70.82   Node-1   <none>           <none>
    vcjob            mindx-dls-test-default-test-1              1/1     Terminating   0          6m59s     192.168.39.9    Node-2     <none>           <none>
    ……
    ```

3.  等待一段时间，执行以下命令查看训练任务弹性伸缩情况。

    ```
     kubectl get pods -A -o wide
    ```

    回显示例如下，表示训练任务根据当前可用节点数将2节点16卡任务伸缩为1节点8卡任务。

    ```
    NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE         NOMINATED NODE   READINESS GATES
    ……
    vcjob            mindx-dls-test-default-test-0              1/1     Running   0          107s    192.168.70.86   Node-2   <none>           <none>
    ……
    ```

**查看单个Pod运行情况<a name="section89223312467"></a>**

执行以下命令，查看单个Pod的训练任务运行情况。

```
kubectl logs mindx-dls-test-default-test-0 -n vcjob -f
```

-   回显示例如下表示发生故障时，使用最近保存的第39步的checkpoint文件恢复，实现训练任务第40个epoch开始继续训练。

    ```
    ...
    2023-06-09 22:17:33,441:INFO:--> pre_trained: /job/code/mindspore/output/resnet50/imagenet2012/ckpt_0/resnet50-39_48.ckpt
    2023-06-09 22:17:33,441:INFO:--> run_eval: False
    2023-06-09 22:17:33,441:INFO:--> eval_dataset_path: 
    2023-06-09 22:17:33,441:INFO:--> parameter_server: False
    2023-06-09 22:17:33,441:INFO:--> filter_weight: False
    2023-06-09 22:17:33,441:INFO:--> save_best_ckpt: True
    2023-06-09 22:17:33,441:INFO:--> eval_start_epoch: 40
    2023-06-09 22:17:33,441:INFO:--> eval_interval: 1
    2023-06-09 22:17:33,441:INFO:--> enable_cache: False
    2023-06-09 22:17:33,441:INFO:--> cache_session_id: 
    2023-06-09 22:17:33,441:INFO:--> mode_name: GRAPH
    2023-06-09 22:17:33,441:INFO:--> boost_mode: O0
    2023-06-09 22:17:33,441:INFO:--> conv_init: XavierUniform
    2023-06-09 22:17:33,441:INFO:--> dense_init: TruncatedNormal
    2023-06-09 22:17:33,442:INFO:--> all_reduce_fusion_config: [85, 160]
    2023-06-09 22:17:33,442:INFO:--> train_image_size: 224
    2023-06-09 22:17:33,442:INFO:--> eval_image_size: 224
    2023-06-09 22:17:33,442:INFO:--> device_id: 0
    2023-06-09 22:17:33,442:INFO:--> width: 224
    2023-06-09 22:17:33,442:INFO:--> height: 224
    2023-06-09 22:17:33,442:INFO:--> file_name: resnet50
    2023-06-09 22:17:33,442:INFO:--> file_format: MINDIR
    2023-06-09 22:17:33,442:INFO:--> ckpt_file: 
    2023-06-09 22:17:33,442:INFO:--> network_dataset: resnet50_imagenet2012
    2023-06-09 22:17:33,442:INFO:--> save_graphs: False
    2023-06-09 22:17:33,442:INFO:--> save_graphs_path: ./graphs
    2023-06-09 22:17:33,442:INFO:--> has_trained_epoch: 0
    2023-06-09 22:17:33,442:INFO:--> has_trained_step: 0
    2023-06-09 22:17:33,442:INFO:--> result_path: 
    2023-06-09 22:17:33,442:INFO:--> label_path: 
    2023-06-09 22:17:33,442:INFO:--> config_path: /job/code/mindspore/config/resnet50_imagenet2012_config.yaml
    2023-06-09 22:17:33,442:INFO:--> rank_id: 0
    2023-06-09 22:17:33,442:INFO:--> save_ckpt_dir: /job/code/mindspore/output/resnet50/imagenet2012/ckpt
    2023-06-09 22:17:33,442:INFO:--> log_dir: /job/code/mindspore/output/resnet50/imagenet2012/log
    2023-06-09 22:17:33,442:INFO:--> logger: <LOGGER resnet (NOTSET)>
    2023-06-09 22:17:33,442:INFO:
    [WARNING] DEVICE(312,fffd6e363470,python):2023-06-09-22:17:33.999.925 [mindspore/ccsrc/plugin/device/ascend/hal/hardware/ge_graph_executor.cc:128] RunGEInitGraph] Can not find init_subgraph.kernel_graph_0 subgraph, don't need data init subgraph in INFER mode.
    [WARNING] DEVICE(312,fffd6e363470,python):2023-06-09-22:17:43.733.157 [mindspore/ccsrc/plugin/device/ascend/hal/hardware/ge_graph_executor.cc:128] RunGEInitGraph] Can not find init_subgraph.kernel_graph_1 sub graph, don't need data init subgraph in INFER mode.
    ....2023-06-09 22:18:45,025:INFO:epoch: [40/90] loss: 3.465011, epoch time: 71.582 s, per step time: 1491.285 ms
    2023-06-09 22:18:49,453:INFO:epoch: [41/90] loss: 3.396700, epoch time: 4.428 s, per step time: 92.245 ms
    .2023-06-09 22:19:02,685:INFO:epoch: [42/90] loss: 3.297215, epoch time: 13.232 s, per step time: 275.659 ms
    2023-06-09 22:19:07,323:INFO:epoch: [43/90] loss: 3.289656, epoch time: 4.638 s, per step time: 96.622 ms
    2023-06-09 22:19:11,746:INFO:epoch: [44/90] loss: 3.266534, epoch time: 4.423 s, per step time: 92.139 ms
    2023-06-09 22:19:16,913:INFO:epoch: [45/90] loss: 3.180886, epoch time: 5.167 s, per step time: 107.650 ms
    2023-06-09 22:19:21,377:INFO:epoch: [46/90] loss: 2.895963, epoch time: 4.464 s, per step time: 92.997 ms
    2023-06-09 22:19:25,798:INFO:epoch: [47/90] loss: 2.815258, epoch time: 4.420 s, per step time: 92.090 ms
    2023-06-09 22:19:31,122:INFO:epoch: [48/90] loss: 2.826911, epoch time: 5.324 s, per step time: 110.918 ms
    2023-06-09 22:19:35,591:INFO:epoch: [49/90] loss: 2.712467, epoch time: 4.469 s, per step time: 93.098 ms
    ...
    ```



#### 删除任务<a name="ZH-CN_TOPIC_0000002511347063"></a>

1.  登录管理节点，进入YAML文件所在路径。
2.  在管理节点执行以下命令，使用YAML删除训练任务。

    ```
    kubectl delete -f XXX.yaml
    ```

    例如：

    ```
    kubectl delete -f a800_vcjob.yaml
    ```

    回显如下：

    ```
    configmap/rings-config-mindx-dls-test deleted
    job.batch.volcano.sh/mindx-dls-test deleted
    ```



### 集成后使用<a name="ZH-CN_TOPIC_0000002511347077"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1.  根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2.  根据K8s官方提供的API库，来对任务进行创建、查询、删除等操作。
3.  创建、查询或删除操作任务时，用户需要将[示例YAML](#准备任务yaml-3)的内容转换成K8s官方API中定义的对象，通过官方库中提供的API发送给K8s的API Server或者将YAML内容转换为JSON格式直接发送给K8s的API Server。



## 推理卡故障重调度<a name="ZH-CN_TOPIC_0000002479387124"></a>

### 使用前必读<a name="ZH-CN_TOPIC_0000002479387116"></a>

集群调度组件管理的推理芯片资源出现故障后，集群调度组件可以对故障资源（对应芯片）进行隔离并自动进行重调度。

**前提条件<a name="section166381652174516"></a>**

-   使用推理卡故障重调度特性，需要确保已经安装如下组件。
    -   Volcano（本特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    -   Ascend Device Plugin
    -   Ascend Docker Runtime
    -   Ascend Operator
    -   ClusterD
    -   NodeD

-   若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。

**使用方式<a name="zh-cn_topic_0000001559979444_section91871616135119"></a>**

推理卡故障重调度的使用方式如下：

-   [通过命令行使用](#通过命令行使用volcano-4)：安装集群调度组件，通过命令行使用推理卡故障重调度特性。
-   [集成后使用](#集成后使用-4)：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section10769161412815"></a>**

-   资源监测可以和推理场景下的所有特性一起使用。
-   集群中同时跑多个推理任务，每个任务使用的特性可以不同，但不能同时存在使用静态vNPU的任务和使用动态vNPU的任务。
-   推理卡故障重调度特性默认使用整卡调度；不支持静态vNPU调度；支持Atlas 推理系列产品使用动态vNPU调度。
-   推理卡故障重调度支持下发单副本数或者多副本数的单机任务，每个副本独立工作；只支持推理服务器（插Atlas 300I Duo 推理卡）和Atlas 800I A2 推理服务器、A200I A2 Box 异构组件部署acjob类型的分布式任务。

-   推理卡故障重调度支持vcjob或Deployment类型任务，且需在该类任务中增加故障重调度的开关的标签“fault-scheduling”，并将其设置为“grace”或者“force”。

**支持的产品形态<a name="section169961844182917"></a>**

支持以下产品使用推理卡故障重调度。

-   推理服务器（插Atlas 300I 推理卡）
-   Atlas 推理系列产品
-   Atlas 800I A2 推理服务器
-   A200I A2 Box 异构组件
-   Atlas 800I A3 超节点服务器
-   推理服务器（插Atlas 350 标卡）

**使用流程<a name="zh-cn_topic_0000001559979444_section246711128536"></a>**

通过命令行使用推理卡故障重调度特性流程可以参见[图1](#zh-cn_topic_0000001559979444_fig242524985412)。

**图 1**  使用流程<a name="zh-cn_topic_0000001559979444_fig242524985412"></a>  
![](../../figures/scheduling/使用流程-7.png "使用流程-7")


### 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002511427039"></a>

#### 制作镜像<a name="ZH-CN_TOPIC_0000002511427053"></a>

**获取推理镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section971616541059"></a>**

可选择以下方式中的一种来获取推理镜像。

-   推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据用户的系统架构（ARM或者x86\_64）下载推理基础镜像（如：[ascend-infer](https://www.hiascend.com/developer/ascendhub/detail/e02f286eef0847c2be426f370e0c2596)、[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)）。

    请注意，21.0.4版本之后推理基础镜像默认用户为非root用户，需要在下载基础镜像后对其进行修改，将默认用户修改为root。

    >[!NOTE] 说明 
    >基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。

-   （可选）如果用户需要更个性化的推理环境，可基于已下载的推理基础镜像，再[使用Dockerfile对其进行修改](../common_operations.md#使用dockerfile构建容器镜像tensorflow)。

    完成定制化修改后，用户可以给推理镜像重命名，以便管理和使用。

**加固镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section1294572963118"></a>**

下载或者制作的推理基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../references.md#容器安全加固)章节进行操作。


#### 脚本适配<a name="ZH-CN_TOPIC_0000002479227172"></a>

本章节以昇腾镜像仓库中推理镜像为例为用户介绍操作流程，该镜像已经包含了推理示例脚本，实际推理场景需要用户自行准备推理脚本。在拉取镜像前，需要确保当前环境的网络代理已经配置完成，且能成功访问昇腾镜像仓库。

**从昇腾镜像仓库获取示例脚本<a name="section8181015175911"></a>**

1.  确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2.  在左侧导航栏选择推理镜像，然后选择[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)镜像，获取推理示例脚本。

    >[!NOTE] 说明 
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。


#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002511427029"></a>

>[!NOTE] 说明 
>-   如果用户不使用Ascend Docker Runtime组件，Ascend Device Plugin只会帮助用户挂载“/dev“目录下的设备。其他目录（如“/usr“）用户需要自行修改YAML文件，挂载对应的驱动目录和文件。容器内挂载路径和宿主机路径保持一致。
>-   因为Atlas 200I SoC A1 核心板场景不支持Ascend Docker Runtime，用户也无需修改YAML文件。

**操作步骤<a name="zh-cn_topic_0000001558853680_zh-cn_topic_0000001609074213_section14665181617334"></a>**

1.  下载YAML文件。

    **表 1**  任务类型与硬件型号对应YAML文件

    <a name="zh-cn_topic_0000001609074213_table15169151021912"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001609074213_row16169201019192"><th class="cellrowborder" valign="top" width="19.97%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001609074213_p4169191017192"><a name="zh-cn_topic_0000001609074213_p4169191017192"></a><a name="zh-cn_topic_0000001609074213_p4169191017192"></a>任务类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="20.03%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001609074213_p20181111517147"><a name="zh-cn_topic_0000001609074213_p20181111517147"></a><a name="zh-cn_topic_0000001609074213_p20181111517147"></a>硬件型号</p>
    </th>
    <th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001609074213_p181811156149"><a name="zh-cn_topic_0000001609074213_p181811156149"></a><a name="zh-cn_topic_0000001609074213_p181811156149"></a>YAML文件路径</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.4"><p id="p1693015221828"><a name="p1693015221828"></a><a name="p1693015221828"></a>获取链接</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001609074213_row2169191091919"><td class="cellrowborder" rowspan="2" valign="top" width="19.97%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p6169510191913"><a name="zh-cn_topic_0000001609074213_p6169510191913"></a><a name="zh-cn_topic_0000001609074213_p6169510191913"></a><span id="zh-cn_topic_0000001609074213_ph183921109162"><a name="zh-cn_topic_0000001609074213_ph183921109162"></a><a name="zh-cn_topic_0000001609074213_ph183921109162"></a>Volcano</span>调度的Deployment任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p8853185832112"><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><span id="zh-cn_topic_0000001609074213_ph238151934915"><a name="zh-cn_topic_0000001609074213_ph238151934915"></a><a name="zh-cn_topic_0000001609074213_ph238151934915"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001609074213_p1116971091915"><a name="zh-cn_topic_0000001609074213_p1116971091915"></a><a name="zh-cn_topic_0000001609074213_p1116971091915"></a>infer-deploy-310p-1usoc.yaml</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p784716567219"><a name="p784716567219"></a><a name="p784716567219"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v7.2.RC1/samples/inference/volcano" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row17169201091917"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p14853125832110"><a name="zh-cn_topic_0000001609074213_p14853125832110"></a><a name="zh-cn_topic_0000001609074213_p14853125832110"></a>其他类型推理节点</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p51692100191"><a name="zh-cn_topic_0000001609074213_p51692100191"></a><a name="zh-cn_topic_0000001609074213_p51692100191"></a>infer-deploy.yaml</p>
    </td>
    </tr>
    <tr id="row1137784216212"><td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.5.1.1 "><p id="p9442102131620"><a name="p9442102131620"></a><a name="p9442102131620"></a>Volcano Job任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.5.1.2 "><p id="p367438101714"><a name="p367438101714"></a><a name="p367438101714"></a><span id="ph56332010913"><a name="ph56332010913"></a><a name="ph56332010913"></a>Atlas 800I A2 推理服务器</span></p>
    <p id="p168721535300"><a name="p168721535300"></a><a name="p168721535300"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
    <p id="p17604333153213"><a name="p17604333153213"></a><a name="p17604333153213"></a><span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="p8442112171619"><a name="p8442112171619"></a><a name="p8442112171619"></a>infer-vcjob-910.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p15442424164"><a name="p15442424164"></a><a name="p15442424164"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-vcjob-910.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row3552077269"><td class="cellrowborder" rowspan="2" valign="top" width="19.97%" headers="mcps1.2.5.1.1 "><p id="p6861171325411"><a name="p6861171325411"></a><a name="p6861171325411"></a>Ascend Job任务</p>
    <p id="p12446175211817"><a name="p12446175211817"></a><a name="p12446175211817"></a></p>
    <p id="p5735201117263"><a name="p5735201117263"></a><a name="p5735201117263"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.5.1.2 "><p id="p1328416110919"><a name="p1328416110919"></a><a name="p1328416110919"></a>推理服务器（插<span id="ph93658382564"><a name="ph93658382564"></a><a name="ph93658382564"></a>Atlas 300I Duo 推理卡</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="p10861813135419"><a name="p10861813135419"></a><a name="p10861813135419"></a>pytorch_acjob_infer_310p_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p1986116136544"><a name="p1986116136544"></a><a name="p1986116136544"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/pytorch_acjob_infer_310p_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row512231072611"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1611216221297"><a name="p1611216221297"></a><a name="p1611216221297"></a><span id="ph10342125017508"><a name="ph10342125017508"></a><a name="ph10342125017508"></a>Atlas 800I A2 推理服务器</span></p>
    <p id="p981315183317"><a name="p981315183317"></a><a name="p981315183317"></a><span id="ph176921116163312"><a name="ph176921116163312"></a><a name="ph176921116163312"></a>A200I A2 Box 异构组件</span></p>
    <p id="p4470103717329"><a name="p4470103717329"></a><a name="p4470103717329"></a><span id="ph1695943783214"><a name="ph1695943783214"></a><a name="ph1695943783214"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p4446185212815"><a name="p4446185212815"></a><a name="p4446185212815"></a>pytorch_multinodes_acjob_infer_{xxx}b_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p962512301913"><a name="p962512301913"></a><a name="p962512301913"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/pytorch_multinodes_acjob_infer_910b_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 说明 
    >Volcano支持Job类型任务，但是Job类型任务的YAML需要用户自行根据示例YAML修改适配。

2.  在[整卡调度](#准备任务yaml-1)或者[动态vNPU调度](#准备任务yaml-2)的YAML配置基础上，增加如下加粗字段启用重调度功能，以整卡调度的infer-deploy.yaml为例。

    ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: resnetinfer1-1-deploy
      labels:
          app: infers
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: infers
      template:
        metadata:
          labels:
    ...
             fault-scheduling: grace               # 添加该字段
             ring-controller.atlas: ascend-310   # 添加该字段
        spec:
          schedulerName: volcano
          nodeSelector:
            host-arch: huawei-arm           # Select the os arch. If the os arch is x86, change it to huawei-x86.
    ...
    ```

    **表 2**  fault-scheduling配置项值列表

    <a name="table0396162644916"></a>
    <table><thead align="left"><tr id="row7397112634917"><th class="cellrowborder" valign="top" width="16.48%" id="mcps1.2.4.1.1"><p id="p1339762674911"><a name="p1339762674911"></a><a name="p1339762674911"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="29.42%" id="mcps1.2.4.1.2"><p id="p1139718264499"><a name="p1139718264499"></a><a name="p1139718264499"></a>取值</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.1%" id="mcps1.2.4.1.3"><p id="p123971426144911"><a name="p123971426144911"></a><a name="p123971426144911"></a>含义</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row9397182614491"><td class="cellrowborder" rowspan="2" valign="top" width="16.48%" headers="mcps1.2.4.1.1 "><p id="p113974261490"><a name="p113974261490"></a><a name="p113974261490"></a>fault-scheduling</p>
    <p id="p878610718561"><a name="p878610718561"></a><a name="p878610718561"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="29.42%" headers="mcps1.2.4.1.2 "><p id="p18397192617495"><a name="p18397192617495"></a><a name="p18397192617495"></a>grace</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.1%" headers="mcps1.2.4.1.3 "><p id="p4397112614920"><a name="p4397112614920"></a><a name="p4397112614920"></a>任务使用重调度开关，并在过程中先优雅删除原Pod。</p>
    </td>
    </tr>
    <tr id="row1378627165613"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p2032819617590"><a name="zh-cn_topic_0000001570873348_p2032819617590"></a><a name="zh-cn_topic_0000001570873348_p2032819617590"></a>force</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001570873348_p113286645910"><a name="zh-cn_topic_0000001570873348_p113286645910"></a><a name="zh-cn_topic_0000001570873348_p113286645910"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="zh-cn_topic_0000001570873348_ph38454178285"><a name="zh-cn_topic_0000001570873348_ph38454178285"></a><a name="zh-cn_topic_0000001570873348_ph38454178285"></a>Pod</span>。</p>
    <p id="zh-cn_topic_0000001570873348_p3206181674916"><a name="zh-cn_topic_0000001570873348_p3206181674916"></a><a name="zh-cn_topic_0000001570873348_p3206181674916"></a></p>
    </td>
    </tr>
    <tr id="row11397142634918"><td class="cellrowborder" valign="top" width="16.48%" headers="mcps1.2.4.1.1 "><p id="p7397026134913"><a name="p7397026134913"></a><a name="p7397026134913"></a>ring-controller.atlas</p>
    </td>
    <td class="cellrowborder" valign="top" width="29.42%" headers="mcps1.2.4.1.2 "><a name="ul16397426184918"></a><a name="ul16397426184918"></a><ul id="ul16397426184918"><li>推理服务器（插<span id="ph3690191194813"><a name="ph3690191194813"></a><a name="ph3690191194813"></a>Atlas 300I 推理卡</span>）：ascend-310</li><li><span id="ph56912120486"><a name="ph56912120486"></a><a name="ph56912120486"></a>Atlas 推理系列产品</span>：ascend-310P</li><li><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>、<span id="ph344045773615"><a name="ph344045773615"></a><a name="ph344045773615"></a>A200I A2 Box 异构组件</span>、<span id="ph1175141233710"><a name="ph1175141233710"></a><a name="ph1175141233710"></a>Atlas 800I A3 超节点服务器</span>：ascend-<span id="ph4487202241512"><a name="ph4487202241512"></a><a name="ph4487202241512"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="54.1%" headers="mcps1.2.4.1.3 "><p id="p1397826104915"><a name="p1397826104915"></a><a name="p1397826104915"></a>用于校验任务使用的芯片类型。</p>
    </td>
    </tr>
    </tbody>
    </table>

3.  挂载权重文件。

    ```
    ...
                  ports:     # 分布式训练集合通信端口
                    - containerPort: 2222      
                      name: ascendjob-port      
                  resources:
                    limits:
                      huawei.com/Ascend310P: 1   # 申请的芯片数
                    requests:
                      huawei.com/Ascend310P: 1   #与limits取值一致
                  volumeMounts:
    ...
                      # 权重文件挂载路径
                    - name: weights                  
                      mountPath: /path-to-weights
    ...
              volumes:
    ...
                # 权重文件挂载路径
                - name: weights
                  hostPath:
                    path: /path-to-weights  # 共享存储或者本地存储路径，请根据实际情况修改
    ...
    ```

    >[!NOTE] 说明 
    >-   /path-to-weights为模型权重，需要用户自行准备。mindie镜像可以参考镜像中$ATB\_SPEED\_HOME\_PATH/examples/models/llama3/README.md文件中的说明进行下载。
    >-   ATB_SPEED_HOME_PATH默认路径为“/usr/local/Ascend/atb-models”，在source模型仓中set_env.sh脚本时已配置，用户无需自行配置。

4.  修改所选YAML中的容器启动命令，即“command”字段内容，如果没有则添加“command”字段。

    ```
    ...
          containers:
          - image: ubuntu-infer:v1
    ...
            command: ["/bin/bash", "-c", "cd $ATB_SPEED_HOME_PATH; python examples/run_pa.py --model_path /path-to-weights"]
            resources:
              requests:
    ...
    ```


#### 下发任务<a name="ZH-CN_TOPIC_0000002511427027"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```
kubectl apply -f XXX.yaml
```

例如：

```
kubectl apply -f infer-310p-1usoc.yaml
```

回显示例如下：

```
job.batch/resnetinfer1-2 created
```

>> [!NOTE] 说明 
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f _XXX_.yaml命令删除原任务，再重新下发任务。


#### 查看任务进程<a name="ZH-CN_TOPIC_0000002511427025"></a>

**操作步骤<a name="zh-cn_topic_0000001609093161_zh-cn_topic_0000001609474293_section96791230183711"></a>**

执行以下命令，查看Pod运行状况。

```
kubectl get pod --all-namespaces
```

回显示例如下：

```
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          resnetinfer1-2-scpr5                      1/1     Running   0          20m
...
```


#### 查看推理卡故障重调度结果<a name="ZH-CN_TOPIC_0000002511347069"></a>

当推理任务运行中出现故障时，Volcano会将该任务调度到其他NPU上。

**操作步骤<a name="section18664151111415"></a>**

1.  执行以下命令，查看任务运行状况。

    ```
    kubectl get pod --all-namespaces
    ```

    回显示例如下，任务名称由**resnetinfer1-2-scpr5**变为**resnetinfer1-2-xsdsf**，表示故障重调度特性运行成功。该任务名称由随机字符串生成，以实际名称为准。

    ```
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
    ...
    default      resnetinfer1-2-xsdsf                    1/1    Running   0       10s
    ...
    ```

2.  执行如下命令，查看该任务的日志。

    ```
    kubectl logs -f resnetinfer1-2-xsdsf
    ```

    回显示例如下。

    ```
    [2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Answer[0]:  Deep learning is a subset of machine learning that uses neural networks with multiple layers to model complex relationships between
    [2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Generate[0] token num: (0, 20)
    ```


#### 删除任务<a name="ZH-CN_TOPIC_0000002479387108"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```
kubectl delete -f XXX.yaml
```

例如：

```
kubectl delete -f infer-310p-1usoc.yaml
```

回显示例如下：

```
root@ubuntu:/home/test/yaml# kubectl delete -f infer-310p-1usoc.yaml 
job "resnetinfer1-2" deleted
```



### 集成后使用<a name="ZH-CN_TOPIC_0000002479387118"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1.  根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2.  根据K8s官方提供的API库，来对任务进行创建、查询、删除等操作。
3.  创建、查询或删除操作任务时，用户需要将[示例YAML](#准备任务yaml-4)的内容转换成K8s官方API中定义的对象，通过官方库里面提供的API发送给K8s的API Server或者将YAML内容转换为JSON格式直接发送给K8s的API Server。



## 推理卡故障恢复<a name="ZH-CN_TOPIC_0000002479227136"></a>

**推理卡故障恢复特性**需要搭配**整卡调度特性**一起使用，开启推理卡故障恢复特性只需要将Ascend Device Plugin的启动参数“-hotReset”取值设置为“0”或“2”（默认为“-1”，不支持故障恢复功能）。具体使用方式请参考[整卡调度或静态vNPU调度（推理）](#整卡调度或静态vnpu调度推理)。

Atlas 800I A2 推理服务器、A200I A2 Box 异构组件使用**推理卡故障恢复特性**，仅支持下发单机单卡任务，不支持分布式任务，且需要单独使用[infer-vcjob-910-hotreset.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v7.2.RC1/samples/inference/volcano/infer-vcjob-910-hotreset.yaml)示例下发任务。

>[!NOTE] 说明 
>Atlas 800I A2 推理服务器存在以下两种故障恢复方式，一台Atlas 800I A2 推理服务器只能使用一种故障恢复方式，由集群调度组件自动识别使用哪种故障恢复方式。
>-   方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。
>-   方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。


