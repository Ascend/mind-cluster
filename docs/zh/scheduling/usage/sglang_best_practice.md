# SGLang推理任务最佳实践<a name="ZH-CN_TOPIC_0000002480719278"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002512753445"></a>

MindCluster集群调度组件支持用户通过OME（Open Model Engine）部署SGLang推理任务进行调度和故障实例重调度。

本章节说明相关特性原理及对应配置示例，用户可以参考配置示例部署基于OME的SGLang推理任务。

**前提条件<a name="zh-cn_topic_0000002322062116_section52051339787"></a>**

在部署SGLang推理服务前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。

-   Volcano
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   ClusterD
-   NodeD（可选）

**支持的产品形态<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>**

-   Atlas 800I A2 推理服务器
-   Atlas 800I A3 超节点服务器

**使用方式<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>**

MindCluster集群调度组件支持用户通过以下方式进行SGLang推理服务的容器化部署、故障重调度。本章节仅介绍通过命令行使用和通过脚本一键式部署使用方式。

-   通过命令行使用：通过配置的YAML文件部署任务。
-   通过脚本一键式部署使用：通过自动化脚本参考设计部署任务。
-   集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。


## 部署基于OME的SGLang推理任务<a name="ZH-CN_TOPIC_0000002480571816"></a>

### 实现原理<a name="ZH-CN_TOPIC_0000002512818803"></a>

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到节点对象（node）中。
    -   Ascend Device Plugin上报芯片内存和拓扑信息。

        对于包含片上内存的芯片，Ascend Device Plugin启动时上报芯片内存情况，见node-label说明；上报整卡信息，将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）、已使用的芯片数量（allocated）和芯片的基础信息（device ip和super\_device\_ip）上报到node中，用于整卡调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中的信息后，将信息整合到cluster-info-cm中。
3.  用户通过kubectl或者其他深度学习平台下发OME框架的SGLang推理任务，OME根据推理任务的配置生成Deployment或者LeaderWorkerSet（LWS）的子工作负载，再由对应的子工作负载生成多个推理服务的任务Pod。关于Deployment或者LeaderWorkerSet的详细说明，可以参见[OME文档](https://docs.sglang.ai/ome/docs/concepts/inference_service/)。
4.  volcano-controller或者LeaderWorkerSet为任务创建相应的PodGroup。关于PodGroup的详细说明，可以参见[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  对于SGLang推理任务Pod，volcano-scheduler根据节点内存、CPU及标签、亲和性选择合适的节点，volcano-scheduler还会参考芯片拓扑信息为其选择合适的节点，并在Pod的annotation上写入选择的芯片信息以及节点硬件信息。
6.  kubelet创建容器时，对于基于OME部署的SGLang推理任务，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin或volcano-scheduler在Pod的annotation上写入芯片和节点硬件信息。Ascend Docker Runtime协助挂载相应资源。


### 通过命令行使用<a name="ZH-CN_TOPIC_0000002480898900"></a>

#### 流程说明<a name="ZH-CN_TOPIC_0000002480995454"></a>

基于OME的SGLang推理任务包含Router  Pod和推理实例Pod，推理实例Pod可以分为Prefill实例Pod和Decode实例Pod，其中Router  Pod不需要使用NPU资源，OME根据不同的推理服务配置方式生成不同的工作负载，用于创建不同的推理实例，并由Router统一对外提供推理服务。MindCluster集群调度组件支持对Deployment和LeaderWorkerSet两种OME推理任务的工作负载进行调度。LeaderWorkerSet任务场景下需要开启LWS的组调度功能。

关于OME任务部署的详细说明可参见[OME文档](https://docs.sglang.ai/ome/docs/)。LWS的组调度功能开启可以参考[LWS文档](https://github.com/kubernetes-sigs/lws/tree/main/docs/examples/sample/gang-scheduling)。

**使用流程<a name="section19644656124210"></a>**

通过命令行使用MindCluster集群调度组件部署基于OME的SGLang推理任务时，使用流程如[图1](#fig38991911205815)所示。

**图 1**  使用流程<a name="fig38991911205815"></a>  
![](../../figures/scheduling/使用流程-15.png "使用流程-15")


#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002480835892"></a>

用户可根据实际情况完成制作镜像的准备工作，然后选择相应的YAML示例，对示例进行修改。

**前提条件<a name="section3759720141513"></a>**

已完成镜像的准备工作。SGLang推理镜像可通过[SGLang文档](https://docs.sglang.ai/get_started/install.html)获取，镜像中依赖的MemFabric Hybrid可通过[MemFabric Hybrid](https://gitcode.com/Ascend/memfabric_hybrid)获取。

**选择YAML示例<a name="section1419519264165"></a>**

基于OME框架的SGLang推理任务可以由Base Model、Serving Runtime和Inference Service三类CRD拉起，Base Model和Inference Service的资源使用和部署请参见[OME文档](https://docs.sglang.ai/ome/docs/)。

集群调度为用户提供OME任务的ClusterServingRuntime资源的YAML示例，用户需要根据使用的组件、芯片类型和任务类型等，选择相应的YAML示例并根据需求进行相应修改后才可使用。

<a name="zh-cn_topic_0000002362848597_table74058394335"></a>
|类型|硬件型号|YAML名称|获取链接|
|--|--|--|--|
|实例不跨机（Deployment场景）|<p>Atlas 800I A2 推理服务器</p><p>Atlas 800I A3 超节点服务器</p>|llama-3-2-1b-instruct-rt-pd-standalone.yaml|<a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/master/k8s-deploy-tool/example/ome-runtimes/llama-3-2-1b-instruct-rt-pd-standalone.yaml">获取YAML</a>|
|实例跨机（LeaderWorkerSet场景）|<p>Atlas 800I A2 推理服务器</p><p>Atlas 800I A3 超节点服务器</p>|llama-3-2-1b-instruct-rt-pd-distributed.yaml|<a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/master/k8s-deploy-tool/example/ome-runtimes/llama-3-2-1b-instruct-rt-pd-distributed.yaml">获取YAML</a>|
|注：当前示例仅供测试使用，用户可根据模型实际情况进行修改。|


用户根据OME框架的部署方式依此完成Base Model、Serving Runtime和Inference Service三个YAML修改之后，由OME及其依赖组件负责拉起子工作负载（Deployment或LeaderWorkerSet）和对应的Pod，并由OME及其依赖组件管理推理服务Pod的生命周期，在推理服务对应的Pod创建完成之后，MindCluster负责对Pod进行调度。

**任务YAML说明<a name="section238217472163"></a>**

```
apiVersion: ome.io/v1beta1
kind: ClusterServingRuntime
metadata:
  name: srt-llama-3-2-1b-instruct-distributed     
spec:
  decoderConfig:
    annotations:
      sp-block: "16"  #仅Atlas 900 A3 SuperPoD 超节点场景配置，大小为一个P/D实例对应的Pod请求的NPU总数       
      huawei.com/schedule_minAvailable: "2" #仅在实例不跨机，即Deployment场景下配置，大小为D实例(在engineConfig字段中为P实例)的副本数量
    leader:
      nodeSelector:
        accelerator-type: module-a3-16-super-pod   #根据实际节点类型配置
        schedulerName: volcano  #设置调度器为Volcano
      runner:
        name: sglang-decoder
        image: "sglang:xxx"
        command:
        ...
        env:
        ...
        - name: ASCEND_VISIBLE_DEVICES
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['huawei.com/Ascend910']
        resources:
          limits:
           huawei.com/Ascend910: 16  #根据实际每个Pod所需NPU数量进行配置
          requests:
           huawei.com/Ascend910: 16  #根据实际每个Pod所需NPU数量进行配置
       volumeMounts:
       ...
       - name: driver
         mountPath: /usr/local/Ascend/driver
       ...
     volumes:
      ...
      - name: driver
        hostPath:
        path: /usr/local/Ascend/driver
    ...
```


#### YAML参数说明<a name="ZH-CN_TOPIC_0000002513115345"></a>

下表仅说明OME的Serving Runtime YAML中与MindCluster有关的字段。

**表 1**  YAML参数说明

<a name="zh-cn_topic_0000002329010086_table7602101418317"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1460212146313"><th class="cellrowborder" valign="top" width="27.16%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196029147318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196029147318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196029147318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.28%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1560213143314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1560213143314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1560213143314"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106023141317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106023141317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106023141317"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1960421417318"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p460415143318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p460415143318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p460415143318"></a>schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56045145317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56045145317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56045145317"></a>取值为<span class="parmvalue" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue10604111417319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue10604111417319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue10604111417319"></a>“volcano”</span>。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p7109162113916"><a name="p7109162113916"></a><a name="p7109162113916"></a>配置调度器为<span id="zh-cn_topic_0000002322062116_ph175881448132716"><a name="zh-cn_topic_0000002322062116_ph175881448132716"></a><a name="zh-cn_topic_0000002322062116_ph175881448132716"></a>Volcano</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row260820141037"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860814141536"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860814141536"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860814141536"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><a name="ul451248112016"></a><a name="ul451248112016"></a><ul id="ul451248112016"><li><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph5608814330"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph5608814330"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph5608814330"></a>Arm</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></li><li><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph186088141531"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph186088141531"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph186088141531"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060801414315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060801414315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060801414315"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76084142313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76084142313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76084142313"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row56081214237"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960818141031"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960818141031"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960818141031"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p14755536454"><a name="zh-cn_topic_0000002039339953_p14755536454"></a><a name="zh-cn_topic_0000002039339953_p14755536454"></a>指定逻辑超节点芯片数量。</p>
<p id="p161001559326"><a name="p161001559326"></a><a name="p161001559326"></a>需要是节点芯片数量的整数倍，且P/D实例的总芯片数量是其整数倍。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1670155202912"><a name="p1670155202912"></a><a name="p1670155202912"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19701652112917"><a name="p19701652112917"></a><a name="p19701652112917"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<p id="p12461828061"><a name="p12461828061"></a><a name="p12461828061"></a>仅支持在<span id="ph914694014812"><a name="ph914694014812"></a><a name="ph914694014812"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</p>
</td>
</tr>
<tr id="row656523055610"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="p7566193012561"><a name="p7566193012561"></a><a name="p7566193012561"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><p id="p115661830165610"><a name="p115661830165610"></a><a name="p115661830165610"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p12566330135620"><a name="p12566330135620"></a><a name="p12566330135620"></a>任务能够调度的最小副本数。在实例不跨机，即Deployment场景下必须指定该字段，根据该字段所属的P实例或者D实例，配置为engine或者decoder的生效副本数量。其他场景下不需要指定该字段。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row46101144312"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1861010140316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1861010140316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1861010140316"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul186101614131"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul186101614131"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul186101614131"><li>on：开启<span id="ph164712399403"><a name="ph164712399403"></a><a name="ph164712399403"></a>Pod</span>级别重调度</li><li>其他值或不使用该字段：关闭<span id="ph126431540134014"><a name="ph126431540134014"></a><a name="ph126431540134014"></a>Pod</span>级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p661016141437"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p661016141437"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p661016141437"></a><span id="ph4287194517407"><a name="ph4287194517407"></a><a name="ph4287194517407"></a>Pod</span>级重调度，表示任务发生故障后，不会删除PodGroup内的所有任务<span id="ph1595595534015"><a name="ph1595595534015"></a><a name="ph1595595534015"></a>Pod</span>，而是将发生故障的<span id="ph750524344015"><a name="ph750524344015"></a><a name="ph750524344015"></a>Pod</span>进行删除，由控制器重新创建新<span id="ph1521154416407"><a name="ph1521154416407"></a><a name="ph1521154416407"></a>Pod</span>后进行重调度。</p>
<p id="p1745415523710"><a name="p1745415523710"></a><a name="p1745415523710"></a>OME推理任务需要将此字段配置为<span class="uicontrol" id="uicontrol172211234283"><a name="uicontrol172211234283"></a><a name="uicontrol172211234283"></a>“on”</span>，<span id="ph19262113954410"><a name="ph19262113954410"></a><a name="ph19262113954410"></a>MindCluster</span>对发生故障的P/D实例进行重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row36114148312"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1461116146318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1461116146318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1461116146318"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461118141037"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461118141037"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461118141037"><li><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>：module-910b-8</li><li><span id="ph2385246171619"><a name="ph2385246171619"></a><a name="ph2385246171619"></a>Atlas 800I A3 超节点服务器</span>：module-a3-16</li><li><span id="ph261924414289"><a name="ph261924414289"></a><a name="ph261924414289"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li>Atlas 350 标卡（可选）：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>Atlas 850 服务器（可选）：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>Atlas 950 PoD 超节点（可选）：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p6612914039"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p6612914039"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p6612914039"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row18613714439"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196131140315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196131140315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196131140315"></a>huawei.com/Ascend910</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><a name="ul5849154316123"></a><a name="ul5849154316123"></a><ul id="ul5849154316123"><li><span id="ph20407103618121"><a name="ph20407103618121"></a><a name="ph20407103618121"></a>Atlas 800I A2 推理服务器</span>：8</li><li><span id="zh-cn_topic_0000002329010086_ph747840144217"><a name="zh-cn_topic_0000002329010086_ph747840144217"></a><a name="zh-cn_topic_0000002329010086_ph747840144217"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph2061955101216"><a name="ph2061955101216"></a><a name="ph2061955101216"></a>Atlas 800I A3 超节点服务器</span>: 16</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p561713141331"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p561713141331"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p561713141331"></a>请求的NPU数量。当前仅支持整机调度，请根据实际硬件卡数进行修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row11621414533"><td class="cellrowborder" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p894317013244"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p894317013244"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p894317013244"></a>env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p7622914235"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p7622914235"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p7622914235"></a>取值为metadata.annotations['huawei.com/Ascend910']，和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136226142031"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136226142031"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136226142031"></a><span id="zh-cn_topic_0000002362968521_ph230432885618"><a name="zh-cn_topic_0000002362968521_ph230432885618"></a><a name="zh-cn_topic_0000002362968521_ph230432885618"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186225141637"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186225141637"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186225141637"></a>该参数只支持使用<span id="ph123731542141613"><a name="ph123731542141613"></a><a name="ph123731542141613"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row662216141939"><td class="cellrowborder" rowspan="5" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106221514533"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106221514533"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106221514533"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p206221814637"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p206221814637"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p206221814637"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462216142314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462216142314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462216142314"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="ph12553844101710"><a name="ph12553844101710"></a><a name="ph12553844101710"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="ph1086910463172"><a name="ph1086910463172"></a><a name="ph1086910463172"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1262313144314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762371419310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762371419310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762371419310"></a>force</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p562301420319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p562301420319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p562301420319"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="ph7477195381711"><a name="ph7477195381711"></a><a name="ph7477195381711"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row46230146312"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126231144312"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126231144312"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126231144312"></a>off</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p26233141317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p26233141317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p26233141317"></a>该推理任务不使用故障重调度特性。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row7623191419310"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11624181414311"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11624181414311"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11624181414311"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row106241614036"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624191420310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624191420310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624191420310"></a>其他值</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row76241014637"><td class="cellrowborder" rowspan="2" valign="top" width="27.16%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p262451412319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p262451412319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p262451412319"></a>fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.28%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76249141830"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76249141830"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76249141830"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624101418314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624101418314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624101418314"></a>处理业务面故障，必须配置业务面无条件重试的次数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row06256141536"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462511141837"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462511141837"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462511141837"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p12625914632"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p12625914632"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p12625914632"></a>该任务不使用无条件重试功能，发生业务面故障之后<span id="ph174393280178"><a name="ph174393280178"></a><a name="ph174393280178"></a>Volcano</span>不会主动删除故障的<span id="ph210853715184"><a name="ph210853715184"></a><a name="ph210853715184"></a>Pod</span>。</p>
</td>
</tr>
</tbody>
</table>


#### 推理任务的下发、查看与删除<a name="ZH-CN_TOPIC_0000002513375093"></a>

用户完成任务YAML的准备工作之后，就可以进行以下操作：

1.  下发推理任务
2.  查看调度结果
3.  查看推理任务运行情况
4.  （可选）删除任务

了解以上步骤的详细说明，请参见[OME文档](https://docs.sglang.ai/ome/docs/tasks/run-workloads/deploy-inference-service/)。



### 通过脚本一键式部署使用<a name="ZH-CN_TOPIC_0000002480866426"></a>

用户在K8s集群中部署多个相关联的推理任务，手动编写和维护大量的K8s YAML文件效率低下且容易出错。为此，MindCluster提供一个自动化脚本参考设计，替代繁琐的手动操作。用户只需提供基本的应用信息（如应用名、镜像版本、副本数等），脚本就能自动生成所有必要的、符合规范的K8s YAML文件，并直接部署到指定集群。同时，MindCluster提供一种简单的方式（如指定同一个应用名）一键删除所有相关资源。

当前脚本仅支持P/D分离部署，可以为用户同时拉起多个P/D实例、Router以及Memfabric\_Store服务端。

**前提条件<a name="section178303526285"></a>**

-   环境已安装Python，并可联网下载依赖包。
-   存在KubeConfig文件，可以与K8s集群正常通信。
-   已部署MindCluster和OME。
-   已部署任务所需的Base Model和Serving Runtime。

**操作步骤<a name="section116575516299"></a>**

1.  从mindcluster-deploy仓库获取源码，进入“k8s-deploy-tool”目录。

    ```
    git clone https://gitcode.com/Ascend/mindcluster-deploy.git && cd mindcluster-deploy/k8s-deploy-tool
    ```

2.  （可选）创建并激活Python虚拟环境。该操作可以使得不同Python项目使用不同版本的库而互不干扰。

    ```
    python -m venv venv && source venv/bin/activate
    ```

    根据环境实际情况使用Python或Python3。

3.  安装依赖。

    ```
    pip install -r requirements.txt
    ```

4.  （可选）部署示例Serving Runtime。该示例用作测试使用，用户可以根据任务实际情况部署对应的Serving Runtime。

    ```
    kubectl apply -f example/ome-runtimes/
    ```

5.  编辑用户配置文件“config/isvc-config.yaml”。
    1.  打开“config/isvc-config.yaml”文件。

        ```
        vi config/isvc-config.yaml
        ```

    2.  按“i”进入编辑模式，按实际情况修改文件中的字段。
    3.  按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。

6.  （可选）创建任务名称空间。"xxx"为“config/isvc-config.yaml”设置的“app\_namespace”。如果“app\_namespace”为“default”或未设置，可以不创建名称空间。

    ```
    kubectl create ns xxx
    ```

7.  （可选）设置服务框架类型。当前支持ome和aibrix，若不设置，默认使用ome。

    ```
    export SERVING_FRAMEWORK=ome
    ```

8.  部署推理任务。

    ```
    python main.py deploy -c config/isvc-config.yaml
    ```

    根据环境实际情况使用Python或Python3。参数说明如下：

    -   -c, --config：配置文件路径，必填。
    -   -k, --kubeconfig：KubeConfig文件路径，选填。默认值为\~/.kube/config。
    -   --dry-run：试运行（不实际部署，展示生成的YAML），选填。

9.  查看任务运行状态。

    ```
    python main.py status -n my-test -ns default
    ```

    参数说明如下：

    -   -n, --app-name：应用名称，必填。my-test为“config/isvc-config.yaml”中设置的“app\_name”。
    -   -ns, --namespace：应用命名空间，选填。默认值为"default" 。
    -   -k, --kubeconfig：KubeConfig文件路径，选填。默认值为\~/.kube/config。

    >[!NOTE] 说明 
    >用户也可以使用kubectl命令行工具查看任务运行状态。

10. 新建终端窗口，在当前K8s集群的节点中执行以下命令，访问推理服务。若请求成功返回，表示推理服务部署成功。

    ```
    curl --location 'http://<router-podip>:<router-port>/generate' --header 'Content-Type: application/json' --data '{
    "text": "Who are you",
    "sampling_params": {
    "temperature": 0,
    "max_new_tokens": 20
    },
    "stream": true
    }'
    ```

    -   <router-podip\>为Router Pod的IP地址，可以通过以下命令查看。

        ```
        kubectl get pod -A -o wide
        ```

    -   <router-port\>为Serving Runtime中Router设置的服务端口。

11. （可选）删除推理任务。若用户需要删除任务，可以执行该步骤。

    ```
    python main.py delete -n my-test 
    ```

    根据环境实际情况使用Python或Python3。参数说明如下：

    -   -n, --app-name：应用名称，必填。
    -   -ns, --namespace：应用命名空间，选填。默认值为"default" 。
    -   -k, --kubeconfig：KubeConfig文件路径，选填。默认值为\~/.kube/config。



## 配置推理任务实例重调度<a name="ZH-CN_TOPIC_0000002480738948"></a>

当推理任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。如需了解故障的检测原理，请参见[故障检测](../usage/resumable_training.md#故障检测)章节。

**前提条件<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>**

已完成部署基于OME的SGLang推理服务。

**实例重调度原理<a name="zh-cn_topic_0000002356060805_section4253197539"></a>**

**故障实例Pod的删除**

OME子工作负载为Deployment时（一个P/D实例由一个Pod组成）：

-   业务面故障：Pod所属的容器发生非零退出的情况下自动重拉。
-   硬件故障：Ascend Device Plugin或者NodeD上报硬件故障到ClusterD之后，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。

OME子工作负载为LeaderWorkerSet时（一个P/D实例由多个Pod组成）：

-   业务面故障：对于任意实例所属Pod的容器发生非零退出之后，LWS Controller自动删除实例所属整个PodGroup。
-   硬件故障：Ascend Device Plugin或者NodeD上报硬件故障到ClusterD之后，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。LWS Controller自动删除实例所属整个PodGroup。

**故障实例Pod的重新创建和调度**

Deployment或者LeaderWorkerSet所属的Pod被Volcano删除之后，由各自对应的Controller重新创建被删除的Pod，并由Volcano执行对恢复Pod的重新调度。

>[!NOTE] 说明 
>OME任务进行故障恢复时只会重调度故障的P/D实例。

**配置实例级别重调度<a name="section96795436354"></a>**

下面以ClusterServingRuntime为例配置实例级别重调度。

```
apiVersion: ome.io/v1beta1
kind: ClusterServingRuntime
metadata:
  name: lws-runtime
  annotations:
    sp-block: "16"
  labels:
    fault-scheduling: "force"          # 开启重调度功能
    pod-rescheduling: "on"             # 开启Pod级重调度
    fault-retry-times: "3"             # 开启业务面故障无条件重试能力
spec:
...
```


