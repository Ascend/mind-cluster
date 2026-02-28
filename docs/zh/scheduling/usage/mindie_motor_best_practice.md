# MindIE Motor推理任务最佳实践<a name="ZH-CN_TOPIC_0000002479227060"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002511346371"></a>

MindCluster集群调度组件支持用户通过生成acjob推理任务的方式进行MindIE Motor的容器化部署、故障重调度和弹性扩缩容。

本章节仅说明相关特性原理及对应配置示例，所提供的YAML示例不足以完成MindIE任务的部署。了解MindIE Motor的详细部署流程请参见《<a href="https://www.hiascend.com/document/detail/zh/mindie/230/mindiemotor/motordev/mindie_service0001.html">MindIE Motor开发指南</a>》。

**前提条件<a name="zh-cn_topic_0000002322062116_section52051339787"></a>**

在部署MindIE Service前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。

-   Volcano
-   Ascend Device Plugin
-   Ascend Docker Runtime
-   Ascend Operator，且需将启动参数[enableGangScheduling](../installation_guide.md#ascend-operator)的取值设置为true
-   ClusterD
-   NodeD

**支持的产品形态<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>**

-   Atlas 800I A2 推理服务器
-   Atlas 800I A3 超节点服务器

**使用方式<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>**

MindCluster集群调度组件支持用户通过以下2种方式进行MindIE Service的容器化部署、故障重调度和弹性扩缩容。本章节仅介绍通过命令行使用这种方式。

-   通过命令行使用：通过配置的YAML文件部署任务。
-   集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。


## 部署MindIE Motor<a name="ZH-CN_TOPIC_0000002511346333"></a>

### 实现原理<a name="ZH-CN_TOPIC_0000002511426301"></a>

![](../../figures/scheduling/zh-cn_image_0000002511426353.png)

各步骤说明如下：

1.  集群调度组件定期上报节点和芯片信息；kubelet上报节点芯片数量到节点对象（node）中。
    -   Ascend Device Plugin上报芯片内存和拓扑信息。

        对于包含片上内存的芯片，Ascend Device Plugin启动时上报芯片内存情况，见node-label说明；上报整卡信息，将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）、已使用的芯片数量（allocated）和芯片的基础信息（device ip和super\_device\_ip）上报到node中，用于整卡调度。

    -   当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2.  ClusterD读取device-info-cm和node-info-cm中信息后，将信息整合到cluster-info-cm中。
3.  用户通过kubectl或者其他深度学习平台下发不使用NPU卡的MS Controller、MS Coordinator以及数个使用NPU卡的MindIE Server任务。
4.  Ascend Operator为任务创建相应的podGroup。关于podGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5.  Ascend Operator为任务创建相应的Pod，并注入MindIE Server服务启动所需的环境变量。关于环境变量的详细说明请参见[环境变量说明](../appendix.md#环境变量说明)中"Ascend Operator注入的训练环境变量"表。
6.  对于MS Controller、MS Coordinator任务，volcano-scheduler根据节点内存、CPU及标签、亲和性选择合适节点。对于MindIE Server任务volcano-scheduler还会参考芯片拓扑信息为其选择合适节点，并在Pod的annotation上写入选择的芯片信息以及节点硬件信息。
7.  kubelet创建容器时，对于MindIE Server任务，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin或volcano-scheduler在Pod的annotation上写入芯片和节点硬件信息。Ascend Docker Runtime协助挂载相应资源。
8.  Ascend Operator读取每个MindIE Server任务Pod的annotation信息，生成各自的集合通信文件hccl.json，以ConfigMap形式存储在etcd中。
9.  ClusterD侦听MS Controller、MS Coordinator任务Pod信息以及各个hccl.json对应ConfigMap的变化，实时生成global-ranktable。关于global-ranktable的详细说明请参见[SubscribeRankTable](../api/clusterd.md#subscriberanktable)中"global-ranktable文件说明"部分。
10. MS Controller启动后，与ClusterD建立通信，通过gRPC接口订阅global-ranktable的变化。


### 通过命令行使用<a name="ZH-CN_TOPIC_0000002511426327"></a>

#### 流程说明<a name="ZH-CN_TOPIC_0000002511426315"></a>

MindIE Motor包含两个部分，MindIE MS（MindIE Management Service）和MindIE Server。其中MindIE MS包含MS Controller和MS Coordinator，MindIE Server可以分为Prefill实例和Decode实例。其中MS Controller、MS Coordinator不需要使用NPU资源，MindIE Server需要NPU资源。

MindCluster集群调度组件支持MS Controller、MS Coordinator和MindIE Server组件分别运行在独立的Pod内。使用MindCluster集群调度组件进行MindIE Motor任务部署时，MS Controller、MS Coordinator以及MindIE Server中的每个实例分别以一个AscendJob进行部署，例如一个推理任务包含2个Prefill实例和1个Decode实例，则需要部署5个AscendJob。

了解PD分离服务部署的详细说明可参考《MindIE Motor开发指南》中的“集群服务部署 \> <a href="https://www.hiascend.com/document/detail/zh/mindie/230/mindiemotor/motordev/mindie_service0049.html">PD分离服务部署</a>”章节。

**使用流程<a name="zh-cn_topic_0000002328850238_section5640184231810"></a>**

通过命令行使用MindCluster集群调度组件部署MindIE Motor推理任务时，使用流程如下图所示。

**图 1**  使用流程<a name="fig38991911205815"></a>  
![](../../figures/scheduling/使用流程-14.png "使用流程-14")


#### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479386386"></a>

用户可根据实际情况完成制作镜像的准备工作，然后选择相应的YAML示例，对示例进行修改。

**前提条件<a name="zh-cn_topic_0000002362848597_section629963815311"></a>**

已完成镜像的准备工作。

**选择YAML示例<a name="zh-cn_topic_0000002362848597_section132746121119"></a>**

集群调度为用户提供YAML示例，用户需要根据使用的组件、芯片类型和任务类型等，选择相应的YAML示例并根据需求进行相应修改后才可使用。

<a name="zh-cn_topic_0000002362848597_table74058394335"></a>
|类型|硬件型号|YAML名称|获取链接|
|--|--|--|--|
|ms_controller|-|controller.yaml|<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/c20d2ea32f5ccca8b06b735d31cf36240ed1407f/samples/inference/volcano/mindie-ms">获取YAML</a>|
|ms_coordinator|-|coordinator.yaml|<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/c20d2ea32f5ccca8b06b735d31cf36240ed1407f/samples/inference/volcano/mindie-ms">获取YAML</a>|
|mindie_server|<p>Atlas 800I A2 推理服务器</p><p>Atlas 800I A3 超节点服务器</p>|server.yaml|<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/c20d2ea32f5ccca8b06b735d31cf36240ed1407f/samples/inference/volcano/mindie-ms">获取YAML</a>|
|<p>注：</p><p>如使用的设备为Atlas 800I A3 超节点服务器，请在获取YAML后，参考<a href="#li7390175311918">以下的示例</a>对部分参数进行修改。</p>|

**任务YAML说明<a name="zh-cn_topic_0000002362848597_section1870105118125"></a>**

与普通Ascend Job任务相比，MindIE Motor推理任务需要额外增加以下两个label：app和jobID。MindIE Server使用NPU卡，用户需根据Prefill实例和Decode实例数，下发等量的AscendJob。

>[!NOTE] 说明 
>关于等量acjob的说明如下：例如一个MindIE Motor推理任务包含1个controller、1个coordinator，x个P实例，y个D实例，则需要部署以下数量的acjob：1+1+x+y。

-   **MS Controller、MS Coordinator**不使用NPU卡，分别以一个AscendJob进行部署，支持多副本。MS Controller、MS Coordinator的YAML示例如下。

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: mindie-ms-test-controller
      namespace: mindie
      labels:
        framework: pytorch          
        app: mindie-ms-controller   # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
        jobID: mindie-ms-test       # 当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置
        ring-controller.atlas: ascend-910b
    spec:
      schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
      runPolicy:
        schedulingPolicy:      # Ascend Operator启用“gang”调度生效且调度器为Volcano时，本字段才生效
          minAvailable: 1      # 任务运行总副本数
          queue: default
      successPolicy: AllWorkers
      replicaSpecs:
        Master:
          replicas: 1
          restartPolicy: Always
          template:
            metadata:
              ...
    ```

在以上示例中，关于app和jobID的参数说明如下。如果想了解其他参数的详细说明请参见[YAML参数说明](#yaml参数说明)。

**app**：当前MindIE Motor在Ascend Job任务中的角色，取值包括mindie-ms-controller、mindie-ms-coordinator、mindie-ms-server。

**jobID**：当前MindIE Motor任务在集群中的唯一识别ID，用户可根据需要进行配置。

-   **MindIE Server**的YAML示例如下。

    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: rings-config-mindie-server-0  # 名称必须与以下AscendJob的名称属性相同。前缀“rings-config-”不能修改。
      namespace: mindie
      labels:
        jobID: mindie-ms-test
        ring-controller.atlas: ascend-910b
        mx-consumer-cim: "true"
    data:
      hccl.json: |
        {
            "status":"initializing"
        }
    ---
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: mindie-server-0
      namespace: mindie
      labels:
        framework: pytorch        
        app: mindie-ms-server        # 表示当前MindIE Motor在Ascend Job任务中的角色,不可修改
        jobID: mindie-ms-test        # 当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置
        ring-controller.atlas: ascend-910b
      annotations:
        huawei.com/schedule.filter.faultCode: "8C1F8608,4C1F8608,80E01801"       # 增加该annotation，配置方法请参见YAML参数说明
        huawei.com/schedule.filter.faultLevel: "RestartRequest"       # 增加该annotation，配置方法请参见YAML参数说明
    spec:
      schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
      runPolicy:
        schedulingPolicy:      # Ascend Operator启用“gang”调度生效且调度器为Volcano时，本字段才生效
          minAvailable: 2      # 任务运行总副本数
          queue: default
      successPolicy: AllWorkers
      replicaSpecs:
        Master:
    ```

-   <a name="li7390175311918"></a>如果硬件型号为Atlas 800I A3 超节点服务器，**MindIE Server**的任务YAML需要做以下修改：

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: mindie-server-0
      namespace: mindie
      labels:
        framework: pytorch
        app: mindie-ms-server        # 不可修改
        jobID: mindie-ms-test        # MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置
        ring-controller.atlas: ascend-910b
        fault-scheduling: force
      annotations:
        sp-block: "16"         # 增加该annotation，配置方法请参见YAML参数说明
        huawei.com/schedule.filter.faultCode: "8C1F8608,4C1F8608,80E01801"       # 增加该annotation，配置方法请参见YAML参数说明
        huawei.com/schedule.filter.faultLevel: "RestartRequest"       # 增加该annotation，配置方法请参见YAML参数说明
    spec:
      schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
      runPolicy:
        schedulingPolicy:      # Ascend Operator启用“gang”调度生效且调度器为Volcano时，本字段才生效
          minAvailable: 2      # 任务运行总副本数
          queue: default
      successPolicy: AllWorkers
      replicaSpecs:
        Master:
          replicas: 1
          restartPolicy: Always
          template:
            metadata:
              labels:
                ring-controller.atlas: ascend-910b
                app: mindie-ms-server
                jobID: mindie-ms-test
            spec:
              nodeSelector:
                accelerator: huawei-Ascend910
                # accelerator-type: module-910b-8  # 删除或注释掉该nodeSelector
    ```


#### （可选）配置实例级亲和调度<a name="ZH-CN_TOPIC_0000002511346349"></a>

Atlas 800I A3 超节点服务器场景下，MindCluster集群调度组件支持MindIE Motor推理任务配置任务级别亲和性调度策略，可实现将MindIE Server实例尽量调度到同一个物理超节点中，充分利用HCCS网络，加速实例间的网络通信。

关于逻辑超节点的亲和性调度规则的详细说明，请参见[灵衢总线设备节点网络说明](../references.md#atlas-900-a3-superpod-超节点)章节。

**图 1**  灵衢总线设备节点网络<a name="zh-cn_topic_0000002362872425_fig1054553210321"></a>  
![](../../figures/scheduling/灵衢总线设备节点网络.png "灵衢总线设备节点网络")

**配置实例级亲和性调度<a name="zh-cn_topic_0000002362872425_section18872194156"></a>**

在已完成镜像的准备工作后，用户在进行[准备任务YAML](#准备任务yaml)时，如需为MindIE Motor推理任务配置实例级亲和性调度策略，可同时进行如下配置。

-   任务YAML中指定sp-block字段，sp-block的值必须和job芯片数量一致，保证整个Job调度到一个物理超节点中。

-   MindIE Server实例调度优先保证物理超节点内有预留节点。

-   设置sp-fit为idlest时，MindIE Server实例往更空闲的物理超节点调度。
-   设置podAffinity时，MindIE Server实例往具有更多亲和性Pod的物理超节点调度。

YAML示例如下。

```
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch        
    app: mindie-ms-server        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test        # 当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置
    ring-controller.atlas: ascend-910b
    fault-scheduling: force
  annotations:
    sp-block: "16"              # 指定sp-block字段，集群调度组件会在物理超节点的基础上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度
    sp-fit: "idlest"            # 超节点调度策略，详细说明请参见YAML参数说明
spec:
  schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
  runPolicy:
    schedulingPolicy:      # Ascend Operator启用“gang”调度生效时且调度器为Volcano时，本字段才生效
      minAvailable: 2      # 任务运行总副本数
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:
      restartPolicy: Never
      template:
        metadata:
          labels:
            ring-controller.atlas: ascend-910b
        spec:
          affinity:
            podAffinity:            # 表示逻辑超节点会往具有更多亲和性Pod的物理超节点调度
              preferredDuringSchedulingIgnoredDuringExecution:
              - weight: 100         # 不可修改
                podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      jobID: mindie-ms-test  # 亲和Pod所需要的标签 
                  topologyKey: kubernetes.io/hostname
```


#### YAML参数说明<a name="ZH-CN_TOPIC_0000002511346361"></a>

acjob任务下，任务YAML中各参数的说明如下表所示。

**表 1**  YAML参数说明

<a name="zh-cn_topic_0000002329010086_table7602101418317"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1460212146313"><th class="cellrowborder" valign="top" width="27.18%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196029147318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196029147318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196029147318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.26%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1560213143314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1560213143314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1560213143314"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106023141317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106023141317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106023141317"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row260211141136"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1660311140313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1660311140313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1660311140313"></a>framework</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="ul4975113512712"></a><a name="ul4975113512712"></a><ul id="ul4975113512712"><li>mindspore</li><li>pytorch</li><li>tensorflow</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p4389131101318"><a name="zh-cn_topic_0000002329010086_p4389131101318"></a><a name="zh-cn_topic_0000002329010086_p4389131101318"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row10436102842510"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p2436102814254"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p2436102814254"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p2436102814254"></a>jobID</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_p1619843317517"><a name="zh-cn_topic_0000002329010086_p1619843317517"></a><a name="zh-cn_topic_0000002329010086_p1619843317517"></a>当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p9111039132818"><a name="zh-cn_topic_0000002329010086_p9111039132818"></a><a name="zh-cn_topic_0000002329010086_p9111039132818"></a>该参数仅支持在<span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>、<span id="zh-cn_topic_0000002329010086_ph2472821203013"><a name="zh-cn_topic_0000002329010086_ph2472821203013"></a><a name="zh-cn_topic_0000002329010086_ph2472821203013"></a>Atlas 800I A3 超节点服务器</span>上使用。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row16523123316254"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p13524833182513"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p13524833182513"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p13524833182513"></a>app</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p5524103317257"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p5524103317257"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p5524103317257"></a>表示当前MindIE Motor推理任务在Ascend Job任务中的角色，取值包括mindie-ms-controller、mindie-ms-coordinator、mindie-ms-server。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note4367125713295"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note4367125713295"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note4367125713295"></a><div class="notebody"><a name="zh-cn_topic_0000002329010086_ul139591420161415"></a><a name="zh-cn_topic_0000002329010086_ul139591420161415"></a><ul id="zh-cn_topic_0000002329010086_ul139591420161415"><li>acjob的任务YAML同时包含jobID和app这2个字段时，<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1566531814589"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1566531814589"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1566531814589"></a>Ascend Operator</span>组件会自动传入环境变量MINDX_TASK_ID、APP_TYPE及MINDX_SERVICE_IP，并将其标识为MindIE推理任务。</li><li>关于以上环境变量的详细说明请参见<a href="../appendix.md#环境变量说明">环境变量说明</a>中"Ascend Operator注入的训练环境变量"表。</li><li>该参数仅支持在<span id="ph0338135542520"><a name="ph0338135542520"></a><a name="ph0338135542520"></a>Atlas 800I A2 推理服务器</span>、<span id="zh-cn_topic_0000002329010086_ph2790182618303"><a name="zh-cn_topic_0000002329010086_ph2790182618303"></a><a name="zh-cn_topic_0000002329010086_ph2790182618303"></a>Atlas 800I A3 超节点服务器</span>上使用。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_row17549141912"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_p7754191171915"><a name="zh-cn_topic_0000002329010086_p7754191171915"></a><a name="zh-cn_topic_0000002329010086_p7754191171915"></a>mx-consumer-cim</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_p975441171911"><a name="zh-cn_topic_0000002329010086_p975441171911"></a><a name="zh-cn_topic_0000002329010086_p975441171911"></a>标记该<span id="zh-cn_topic_0000002039699773_ph16931540174313"><a name="zh-cn_topic_0000002039699773_ph16931540174313"></a><a name="zh-cn_topic_0000002039699773_ph16931540174313"></a>ConfigMap</span>是否会被ClusterD侦听。</p>
<p id="p5557415773"><a name="p5557415773"></a><a name="p5557415773"></a>true：是</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p17754171181911"><a name="zh-cn_topic_0000002329010086_p17754171181911"></a><a name="zh-cn_topic_0000002329010086_p17754171181911"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_row1541745918171"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_p143891219113814"><a name="zh-cn_topic_0000002329010086_p143891219113814"></a><a name="zh-cn_topic_0000002329010086_p143891219113814"></a><span>mind-cluster/scaling-rule</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_p3389151918383"><a name="zh-cn_topic_0000002329010086_p3389151918383"></a><a name="zh-cn_topic_0000002329010086_p3389151918383"></a>标记扩缩容规则对应的ConfigMap名称。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p10101730182112"><a name="zh-cn_topic_0000002329010086_p10101730182112"></a><a name="zh-cn_topic_0000002329010086_p10101730182112"></a>仅支持MindIE Motor推理任务在<span id="ph03861516202613"><a name="ph03861516202613"></a><a name="ph03861516202613"></a>Atlas 800I A2 推理服务器</span>、<span id="zh-cn_topic_0000002329010086_ph74121431163016"><a name="zh-cn_topic_0000002329010086_ph74121431163016"></a><a name="zh-cn_topic_0000002329010086_ph74121431163016"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_row172898336182"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_p16388101913817"><a name="zh-cn_topic_0000002329010086_p16388101913817"></a><a name="zh-cn_topic_0000002329010086_p16388101913817"></a><span>mind-cluster/group-name</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_p13387171983812"><a name="zh-cn_topic_0000002329010086_p13387171983812"></a><a name="zh-cn_topic_0000002329010086_p13387171983812"></a>标记扩缩容规则中对应的group名称。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p14289143313188"><a name="zh-cn_topic_0000002329010086_p14289143313188"></a><a name="zh-cn_topic_0000002329010086_p14289143313188"></a>仅支持MindIE Motor推理任务在<span id="ph47631139102619"><a name="ph47631139102619"></a><a name="ph47631139102619"></a>Atlas 800I A2 推理服务器</span>、<span id="zh-cn_topic_0000002329010086_ph81373673012"><a name="zh-cn_topic_0000002329010086_ph81373673012"></a><a name="zh-cn_topic_0000002329010086_ph81373673012"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_row1996920561501"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_p20969356175010"><a name="zh-cn_topic_0000002329010086_p20969356175010"></a><a name="zh-cn_topic_0000002329010086_p20969356175010"></a>podAffinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_p196965635015"><a name="zh-cn_topic_0000002329010086_p196965635015"></a><a name="zh-cn_topic_0000002329010086_p196965635015"></a>表示逻辑超节点会往具有更多亲和性Pod的物理超节点调度。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p3969155611509"><a name="zh-cn_topic_0000002329010086_p3969155611509"></a><a name="zh-cn_topic_0000002329010086_p3969155611509"></a>仅支持MindIE Motor推理任务<span id="zh-cn_topic_0000002329010086_ph249517547298"><a name="zh-cn_topic_0000002329010086_ph249517547298"></a><a name="zh-cn_topic_0000002329010086_ph249517547298"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_row22810219519"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_p128102185119"><a name="zh-cn_topic_0000002329010086_p128102185119"></a><a name="zh-cn_topic_0000002329010086_p128102185119"></a>sp-fit</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_p72820245118"><a name="zh-cn_topic_0000002329010086_p72820245118"></a><a name="zh-cn_topic_0000002329010086_p72820245118"></a>超节点调度策略。</p>
<p id="p9265820378"><a name="p9265820378"></a><a name="p9265820378"></a>idlest：逻辑超节点会往更空闲的物理超节点调度。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_p755841425314"><a name="zh-cn_topic_0000002329010086_p755841425314"></a><a name="zh-cn_topic_0000002329010086_p755841425314"></a>仅支持MindIE Motor推理任务<span id="zh-cn_topic_0000002329010086_ph1858015143594"><a name="zh-cn_topic_0000002329010086_ph1858015143594"></a><a name="zh-cn_topic_0000002329010086_ph1858015143594"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1060320149314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860317149314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860317149314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860317149314"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_ul16230203215710"></a><a name="zh-cn_topic_0000002329010086_ul16230203215710"></a><ul id="zh-cn_topic_0000002329010086_ul16230203215710"><li><span id="zh-cn_topic_0000002329010086_ph20976435102713"><a name="zh-cn_topic_0000002329010086_ph20976435102713"></a><a name="zh-cn_topic_0000002329010086_ph20976435102713"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="zh-cn_topic_0000002329010086_ph163483412215"><a name="zh-cn_topic_0000002329010086_ph163483412215"></a><a name="zh-cn_topic_0000002329010086_ph163483412215"></a>A200T A3 Box8 超节点服务器</span>、<span id="zh-cn_topic_0000002329010086_ph136651315478"><a name="zh-cn_topic_0000002329010086_ph136651315478"></a><a name="zh-cn_topic_0000002329010086_ph136651315478"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="zh-cn_topic_0000002329010086_ph10355115144111"><a name="zh-cn_topic_0000002329010086_ph10355115144111"></a><a name="zh-cn_topic_0000002329010086_ph10355115144111"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="zh-cn_topic_0000002329010086_ph11976935122715"><a name="zh-cn_topic_0000002329010086_ph11976935122715"></a><a name="zh-cn_topic_0000002329010086_ph11976935122715"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>Atlas 800 训练服务器，服务器（插<span id="zh-cn_topic_0000002329010086_ph2099203201811"><a name="zh-cn_topic_0000002329010086_ph2099203201811"></a><a name="zh-cn_topic_0000002329010086_ph2099203201811"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p43811639112614"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p43811639112614"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p43811639112614"></a>标识任务使用的芯片的产品类型。</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p4409148135"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p4409148135"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p4409148135"></a>需要在<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph7409748837"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph7409748837"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph7409748837"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1960421417318"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p460415143318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p460415143318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p460415143318"></a>schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56045145317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56045145317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56045145317"></a>默认值为<span class="parmvalue" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue10604111417319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue10604111417319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue10604111417319"></a>“volcano”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46041014430"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46041014430"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46041014430"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph6604131419312"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph6604131419312"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph6604131419312"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row19604714936"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46047147315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46047147315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46047147315"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106041114132"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106041114132"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106041114132"></a>默认值为任务总副本数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660420141935"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660420141935"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660420141935"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph16604181418316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph16604181418316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph16604181418316"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph156050141033"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph156050141033"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph156050141033"></a>Volcano</span>时，任务运行总副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row86054141139"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p16051014839"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p16051014839"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p16051014839"></a>queue</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1760581413313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1760581413313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1760581413313"></a>默认值为<span class="parmvalue" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue1260516142036"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue1260516142036"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue1260516142036"></a>“default”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p13605414637"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p13605414637"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p13605414637"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph10605114231"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph10605114231"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph10605114231"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1660520141632"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1660520141632"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1660520141632"></a>Volcano</span>时，任务所属队列。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row18605114739"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136051014737"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136051014737"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136051014737"></a>（可选）successPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul6605121420317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul6605121420317"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul6605121420317"><li>默认值为空，若用户不填写该参数，则默认取空值。</li><li>AllWorkers</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126052141730"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126052141730"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126052141730"></a>表明任务成功的前提。空值代表只需要一个<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46053142034"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46053142034"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46053142034"></a>Pod</span>成功，整个任务判定为成功。取值为<span class="parmvalue" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue460514143318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue460514143318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue460514143318"></a>“AllWorkers”</span>表示所有<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46065141434"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46065141434"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46065141434"></a>Pod</span>都成功，任务才判定为成功。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row560612147315"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560613141833"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560613141833"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560613141833"></a>container.name</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1360671419316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1360671419316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1360671419316"></a>ascend</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p860691418319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p860691418319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p860691418319"></a>训练容器的名称必须是<span class="parmvalue" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue1260612147320"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue1260612147320"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue1260612147320"></a>“ascend”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row116068141134"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p5606151413318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p5606151413318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p5606151413318"></a>（可选）ports</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p176065144312"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p176065144312"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p176065144312"></a>若用户未进行设置，系统默认填写以下参数：</p>
<a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul106061214438"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul106061214438"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul106061214438"><li>name：ascendjob-port</li><li>containerPort：2222</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11607414537"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11607414537"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11607414537"></a>分布式训练集合通讯端口。<span class="parmname" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmname160711141237"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmname160711141237"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmname160711141237"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row3607151417314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560721419320"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560721419320"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560721419320"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul156070141730"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul156070141730"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul156070141730"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196071814834"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196071814834"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196071814834"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row86071144316"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660791413315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660791413315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660791413315"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060731418315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060731418315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060731418315"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560811141335"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560811141335"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p560811141335"></a>训练镜像名称，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row260820141037"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860814141536"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860814141536"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1860814141536"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56089142319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56089142319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p56089142319"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph5608814330"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph5608814330"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph5608814330"></a>Arm</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960819141639"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960819141639"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960819141639"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph186088141531"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph186088141531"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph186088141531"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060801414315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060801414315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1060801414315"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76084142313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76084142313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76084142313"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row56081214237"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960818141031"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960818141031"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1960818141031"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p14755536454"><a name="zh-cn_topic_0000002039339953_p14755536454"></a><a name="zh-cn_topic_0000002039339953_p14755536454"></a>指定逻辑超节点芯片数量。</p>
<a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><ul id="zh-cn_topic_0000002039339953_ul10451144414619"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1670155202912"><a name="p1670155202912"></a><a name="p1670155202912"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19701652112917"><a name="p19701652112917"></a><a name="p19701652112917"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note47015215291"><a name="note47015215291"></a><a name="note47015215291"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在<span id="zh-cn_topic_0000002511347099_ph34244153594"><a name="zh-cn_topic_0000002511347099_ph34244153594"></a><a name="zh-cn_topic_0000002511347099_ph34244153594"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row760915145311"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p06096144316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p06096144316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p06096144316"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul206095141139"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul206095141139"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul206095141139"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note66098141039"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note66098141039"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note66098141039"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p8609121419312"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p8609121419312"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p8609121419312"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660912140310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660912140310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p660912140310"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note16091141837"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note16091141837"></a><div class="notebody"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul176092014030"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul176092014030"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul176092014030"><li>交换机亲和性调度1.0版本支持<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1157665817140"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1157665817140"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph168598363399"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph168598363399"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph168598363399"></a>Atlas A2 训练系列产品</span>；支持<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph4181625925"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph4181625925"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph4181625925"></a>PyTorch</span>和<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph61882510210"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph61882510210"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph311717506401"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph311717506401"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph311717506401"></a>Atlas A2 训练系列产品</span>；支持<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph17383182419412"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph17383182419412"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph17383182419412"></a>PyTorch</span>框架。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row46101144312"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1861010140316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1861010140316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1861010140316"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul186101614131"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul186101614131"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul186101614131"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p661016141437"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p661016141437"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p661016141437"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note1561010145316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note1561010145316"></a><div class="notebody"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461013147314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461013147314"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461013147314"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li><li><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1061091414318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1061091414318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1061091414318"></a>TensorFlow</span>暂不支持Pod级别重调度。</li><li>Pod级别重调度目前只支持MS Controller和MS Coordinator。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row205285218207"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p95281922209"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p95281922209"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p95281922209"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul18716519102210"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul18716519102210"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul18716519102210"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p3641163291319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p3641163291319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p3641163291319"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note635023119333"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note635023119333"></a><div class="notebody"><p id="p76416015243"><a name="p76416015243"></a><a name="p76416015243"></a>使用graceExit策略时，需保证训练框架能够接收SIGTERM信号并保存CKPT文件。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row36114148312"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1461116146318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1461116146318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1461116146318"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461118141037"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461118141037"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul461118141037"><li><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph136117141331"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph136117141331"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph136117141331"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph26111143315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph26111143315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph26111143315"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph26117141237"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph26117141237"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph26117141237"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph061115145320"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph061115145320"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph061115145320"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1661112141731"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1661112141731"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1661112141731"></a><em id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b-8</li><li><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1161214145319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1161214145319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1161214145319"></a>Atlas 200T A2 Box16 异构子框</span>和<span id="ph10949202261219"><a name="ph10949202261219"></a><a name="ph10949202261219"></a>Atlas 200I A2 Box16 异构子框</span>：module-<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph116121514934"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph116121514934"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph116121514934"></a><em id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-16</li><li><span id="ph261924414289"><a name="ph261924414289"></a><a name="ph261924414289"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 350 标卡（可选）</span>：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>Atlas 850 服务器（可选）：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>Atlas 950 PoD 超节点（可选）：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p6612914039"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p6612914039"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p6612914039"></a>根据需要运行训练任务的节点类型，选取不同的值。如果节点是<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1961291412314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1961291412314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1961291412314"></a>Atlas 800 训练服务器（NPU满配）</span>，可以省略该标签。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note1861316141738"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note1861316141738"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note1861316141738"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1027616512420"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1027616512420"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1027616512420"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph9014016509"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph9014016509"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph9014016509"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row18613714439"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196131140315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196131140315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p196131140315"></a>huawei.com/Ascend910</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><div class="p" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p361318141733"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p361318141733"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p361318141733"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph106131514134"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph106131514134"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph106131514134"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul126147145311"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul126147145311"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul126147145311"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p3614151416317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p3614151416317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p3614151416317"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph261416147313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph261416147313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph261416147313"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1961418141132"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1961418141132"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1961418141132"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p10614214538"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p10614214538"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p10614214538"></a>服务器（插<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph13615131417315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph13615131417315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph13615131417315"></a>Atlas 300T 训练卡</span>）：<a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1261519142311"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1261519142311"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1261519142311"><li>单机单芯片：1</li><li>单机多芯片：2</li><li>分布式：2</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11615161416311"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11615161416311"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11615161416311"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46156141634"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46156141634"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph46156141634"></a><span id="ph46105378332"><a name="ph46105378332"></a><a name="ph46105378332"></a>Atlas 800T A2 训练服务器</span>和Atlas 900 A2 PoD 集群基础单元</span>：<a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1961514143314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1961514143314"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul1961514143314"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p15616514739"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p15616514739"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p15616514739"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph161611419319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph161611419319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph161611419319"></a>Atlas 200T A2 Box16 异构子框</span>和<span id="ph115117127373"><a name="ph115117127373"></a><a name="ph115117127373"></a>Atlas 200I A2 Box16 异构子框</span>：<a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul661611418316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul661611418316"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul661611418316"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136171314938"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136171314938"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136171314938"></a><span id="zh-cn_topic_0000002329010086_ph747840144217"><a name="zh-cn_topic_0000002329010086_ph747840144217"></a><a name="zh-cn_topic_0000002329010086_ph747840144217"></a>Atlas 900 A3 SuperPoD 超节点</span><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul261751412316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul261751412316"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul261751412316"><li>单机单芯片：1</li><li>单机多芯片：2、4、6、8、10、12、14、16</li><li>分布式：16</li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p561713141331"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p561713141331"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p561713141331"></a>请求的NPU数量，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row11621414533"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p894317013244"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p894317013244"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p894317013244"></a>(.kind=="AscendJob").spec.replicaSpecs.[Master|Scheduler|Worker].template.spec.containers[0].env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p7622914235"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p7622914235"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p7622914235"></a>取值为metadata.annotations['huawei.com/AscendXXX']，其中XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136226142031"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136226142031"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p136226142031"></a><span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1062212140315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1062212140315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph1062212140315"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note462214141730"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note462214141730"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note462214141730"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186225141637"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186225141637"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186225141637"></a>该参数只支持使用<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph962251412315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph962251412315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph962251412315"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row662216141939"><td class="cellrowborder" rowspan="5" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106221514533"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106221514533"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106221514533"></a>fault-scheduling</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p176223141233"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p176223141233"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p176223141233"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106222141235"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106222141235"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106222141235"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17622171414319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17622171414319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17622171414319"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p86225141938"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p86225141938"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p86225141938"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p206221814637"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p206221814637"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p206221814637"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462216142314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462216142314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462216142314"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph19623131417313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph19623131417313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph19623131417313"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph96231114734"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph96231114734"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph96231114734"></a>Pod</span>。</p>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1262313144314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762371419310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762371419310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762371419310"></a>force</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p562301420319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p562301420319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p562301420319"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph19623151420318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph19623151420318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph19623151420318"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row46230146312"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126231144312"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126231144312"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p126231144312"></a>off</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p26233141317"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p26233141317"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p26233141317"></a>该任务不使用断点续训特性，<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph8623191418313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph8623191418313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph8623191418313"></a>K8s</span>的maxRetry仍然生效。</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186239141631"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186239141631"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p186239141631"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1623191419310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1623191419310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1623191419310"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row7623191419310"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11624181414311"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11624181414311"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p11624181414311"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row106241614036"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624191420310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624191420310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624191420310"></a>其他值</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row76241014637"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p262451412319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p262451412319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p262451412319"></a>fault-retry-times</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p96241714134"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p96241714134"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p96241714134"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p662413146319"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p662413146319"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p662413146319"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76249141830"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76249141830"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p76249141830"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624101418314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624101418314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p9624101418314"></a>处理业务面故障，必须配置业务面无条件重试的次数。</p>
<ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul13624161415314"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph121662039143413"><a name="ph121662039143413"></a><a name="ph121662039143413"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph166251314730"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph166251314730"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph166251314730"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row06256141536"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462511141837"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462511141837"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1462511141837"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p12625914632"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p12625914632"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p12625914632"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row146252141832"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17625214933"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17625214933"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17625214933"></a>backoffLimit</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106261014332"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106261014332"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p106261014332"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p146262147310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p146262147310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p146262147310"></a></p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1962631414310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1962631414310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1962631414310"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p66263141230"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p66263141230"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p66263141230"></a>0 &lt; backoffLimit</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p18626614532"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p18626614532"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p18626614532"></a>任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与backoffLimit取值相同时，任务将不再进行重调度。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note15626214934"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note15626214934"></a><div class="notebody"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p362641413318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p362641413318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p362641413318"></a>同时配置了backoffLimit和fault-retry-times参数时，当已经重调度次数与backoffLimit或fault-retry-times取值有一个相同时，将不再进行重调度。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row662614145317"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1062612142315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1062612142315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1062612142315"></a>无（无backoffLimit）或backoffLimit ≤ 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46267141139"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46267141139"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p46267141139"></a>不限制总重调度次数。</p>
<p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p962712140314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p962712140314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p962712140314"></a>若不配置backoffLimit，但是配置了fault-retry-times参数，则使用fault-retry-times的重调度次数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row1662711144310"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p462713149315"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p462713149315"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p462713149315"></a>restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul36271614531"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul36271614531"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul36271614531"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note156280141037"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note156280141037"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note156280141037"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17628131415316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17628131415316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p17628131415316"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762813148318"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762813148318"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p762813148318"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为<span class="parmvalue" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue18628191413311"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue18628191413311"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_parmvalue18628191413311"></a>“Never”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row14628131414314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1362811415312"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1362811415312"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1362811415312"></a>terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1862811412316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1862811412316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1862811412316"></a>0 &lt; terminationGracePeriodSeconds &lt; <strong id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_b1962881410316"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_b1962881410316"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_b1962881410316"></a>grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p4628914935"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p4628914935"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p4628914935"></a>容器收到SIGTERM到被<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph176283140314"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph176283140314"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph176283140314"></a>K8s</span>强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_i1562819141131"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_i1562819141131"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_i1562819141131"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_b10628161415313"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_b10628161415313"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_b10628161415313"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参考<span id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph136288141334"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph136288141334"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ph136288141334"></a>K8s</span>官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note146291714533"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_note146291714533"></a><div class="notebody"><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1962991416310"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1962991416310"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1962991416310"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_row962814644010"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1062816462407"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1062816462407"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p1062816462407"></a>hostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul960434424111"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul960434424111"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul960434424111"><li>true：使用HostIP创建Pod。</li><li>false：不使用HostIP创建Pod。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul14611159182815"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul14611159182815"></a><ul id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_ul14611159182815"><li>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</li><li>不传入此参数时，默认不使用HostIP创建Pod。</li></ul><p id="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p461933317584"><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p461933317584"></a><a name="zh-cn_topic_0000002329010086_zh-cn_topic_0000001951418201_p461933317584"></a>当HostNetwork取值为true时，若当前任务YAML挂载了RankTable文件路径，则可以通过在训练脚本中解析RankTable文件获取Pod的hostIP来实现建链。若任务YAML未挂载RankTable文件路径，则与原始保持一致，使用serviceIP来实现建链。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p>huawei.com/schedule.filter.faultCode</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p>取值示例："8C1F8608:30, 80E01801"，表示在30秒时间窗内，静默8C1F8608故障；在60秒时间窗内，静默80E01801故障。</p><p>若未配置时间窗，则默认为60，取值范围为0~86400，单位为秒。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>配置任务需要静默的故障码和时间窗。</p><ul><li>故障码仅支持配置芯片故障和灵衢总线设备故障的故障码。支持的故障码详细请参见faultCode.json和SwitchFaultCode.json文件。</li><li>支持配置多个故障码和时间窗，多个配置使用英文逗号分隔。</li><li>对于MindIE Service，若YAML文件中无此配置项，则默认静默以下故障码：<ul><li>8C1F8608静默60秒</li><li>4C1F8608静默60秒</li><li>80E01801静默60秒</li></ul></li></ul>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p>huawei.com/schedule.filter.faultLevel</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p>取值示例："RestartRequest:30, RestartBusiness"，表示在30秒时间窗内，静默所有RestartRequest级别的故障；在60秒时间窗内，静默所有RestartBusiness级别的故障。</p><p>若未配置时间窗，则默认为60，取值范围为0~86400，单位为秒。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>配置任务需要静默的故障级别和时间窗。</p><ul><li>故障级别仅支持配置芯片故障和灵衢总线设备故障的级别。支持的故障级别详细请参见<a href="./resumable_training.md#配置说明">配置说明</a>。</li><li>支持配置多个故障级别和时间窗，多个配置使用英文逗号分隔。</li><li>对于MindIE Service，若YAML文件中无此配置项，则默认所有RestartRequest级别的故障静默60秒。</li><li>huawei.com/schedule.filter.faultCode的优先级高于huawei.com/schedule.filter.faultLevel。</li></ul>
</td>
</tr>
</tbody>
</table>


#### 推理任务的下发、查看与删除<a name="ZH-CN_TOPIC_0000002479386412"></a>

用户完成任务YAML的准备工作之后，就可以进行以下操作：

1.  下发推理任务
2.  查看调度结果
3.  查看推理任务运行情况
4.  （可选）删除任务

了解以上步骤的详细说明，请参见《MindIE Motor开发指南》中的“集群服务部署 \> PD分离服务部署 \> 安装部署 \> <a href="https://www.hiascend.com/document/detail/zh/mindie/230/mindiemotor/motordev/mindie_service0318.html">使用kubectl部署单机PD分离服务示例</a>”章节。


#### global-ranktable说明<a name="ZH-CN_TOPIC_0000002479226414"></a>

ClusterD侦听MS Controller、MS Coordinator任务Pod信息以及各个hccl.json对应ConfigMap的变化，实时生成global-ranktable。global-ranktable中部分字段来自于hccl.json文件，关于hccl.json文件的详细说明请参见[hccl.json文件说明](../appendix.md#hccljson文件说明)。

-   Atlas A2 训练系列产品global-ranktable示例如下。

    ```
    {
        "version": "1.0",
        "status": "completed",
        "server_group_list": [
            {
                "group_id": "2",
                "deploy_server": "0",
                "server_count": "1",
                "server_list": [
                    {
                        "device": [
                            {
                                "device_id": "x",
                                "device_ip": "xx.xx.xx.xx",
                                "device_logical_id": "x",
                                "rank_id": "x"
                            }
                        ],                   
                        "server_id": "xx.xx.xx.xx",
                        "server_ip": "xx.xx.xx.xx"
                    }
                ]
            }
        ]
    }
    ```

-   Atlas A3 训练系列产品global-ranktable示例如下。

    ```
    {
        "version": "1.2",
        "status": "completed",
        "server_group_list": [
            {
                "group_id": "2",
                "deploy_server": "1",
                "server_count": "1",
                "server_list": [
                    {
                        "device": [
                            {
                                "device_id": "0",
                                "device_ip": "xx.xx.xx.xx",
                                "super_device_id": "xxxxx",
                                "device_logical_id": "0",
                                "rank_id": "0"
                            }
                        ],
                        "server_id": "xx.xx.xx.xx",
                        "server_ip": "xx.xx.xx.xx"
                    }
                ],
                "super_pod_list": [
                    {
                        "super_pod_id": "0",
                        "server_list": [
                            {
                                "server_id": "xx.xx.xx.xx"
                            }
                        ]
                    }
                ]
            }
        ]
    }
    ```

**表 1**  global-ranktable字段说明

<a name="zh-cn_topic_0000002324328268_table5843145110294"></a>
|字段|说明|
|--|--|
|version|版本|
|status|状态|
|server_group_list|服务组列表|
|group_id|任务组编号|
|server_count|服务器数量|
|server_list|服务器列表|
|server_id|AI Server标识，全局唯一|
|server_ip|Pod IP|
|device_id|NPU的设备ID|
|device_ip|NPU的设备IP|
|super_device_id|<span><term>Atlas A3 训练系列产品</term></span>超节点内NPU的唯一标识|
|rank_id|NPU对应的训练Rank ID|
|device_logical_id|NPU的逻辑ID|
|super_pod_list|超节点列表|
|super_pod_id|逻辑超节点ID|



## 配置推理任务重调度<a name="ZH-CN_TOPIC_0000002479386400"></a>

当推理任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。如需了解故障的检测原理，请参见[故障检测](../usage/resumable_training.md#故障检测)章节。

**前提条件<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>**

已完成[部署MindIE Motor](#部署mindie-motor)。

**支持的故障类型<a name="section121201333144919"></a>**

-   MindIE Server：节点、芯片或其他故障
-   MindIE MS：节点故障

**重调度原理<a name="zh-cn_topic_0000002356060805_section4253197539"></a>**

-   Job级别重调度：MindIE Server和MindIE MS均支持。当MindIE Server或MindIE MS发生故障时，对应的MindIE Server实例或MindIE MS停止所有Pod，重新创建并重调度所有Pod后，最新的global-ranktable.json重新推送给MS Controller，推理任务被重启。

    MindIE Server在PD分离的场景下，例如MindIE Server包含一个Prefill实例和一个Decode实例，Prefill实例发生故障，仅停止Prefill实例的所有Pod，不会影响其他正常运行的实例。

-   Pod级别重调度：仅MindIE MS支持。在开启主备倒换功能场景下，MS Controller或MS Coordinator对应的Pod数量均大于1，当某节点发生故障时，仅停止该节点对应的Pod。例如，MS Coordinator包含主MS Coordinator和备MS Coordinator，主MS Coordinator发生故障时，仅停止主MS Coordinator对应的Pod，不会影响备MS Coordinator。

    >[!NOTE] 说明 
    >若Pod级别重调度恢复失败，则会回退到Job级别重调度处理方式。

**配置Job级别重调度<a name="zh-cn_topic_0000002356060805_section20633874524"></a>**

Job级别重调度默认开启，用户只需完成准备任务YAML的步骤即可。下面以MindIE Server为例说明Job级别重调度的配置。

```
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch        
    app: mindie-ms-server        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test        # 当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置
    fault-scheduling: force    # 开启重调度功能
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
  runPolicy:
    schedulingPolicy:      # Ascend Operator启用“gang”调度生效，且调度器为Volcano时，本字段才生效
      minAvailable: 2      # 任务运行总副本数
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:
```

**配置Pod级别重调度<a name="section5620411141"></a>**

Pod级别重调度目前只支持MS Controller和MS Coordinator，建议在开启主备倒换功能场景下使用。下面以MS Coordinator开启主备倒换功能为例说明Pod级别重调度的配置。

```
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-coordinator
  namespace: mindie
  labels:
    framework: pytorch        
    app: mindie-ms-coordinator        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test             # 当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置
   fault-scheduling: force          # 开启重调度功能
   pod-rescheduling: "on"           # 开启Pod级别重调度
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
  runPolicy:
    schedulingPolicy:      # Ascend Operator启用“gang”调度生效，且调度器为Volcano时，本字段才生效
      minAvailable: 2      # 任务运行总副本数
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:
```


## 配置推理任务场景下的离线复位<a name="ZH-CN_TOPIC_0000002479226442"></a>

当前仅支持Atlas 800I A2 推理服务器、Atlas 800I A3 超节点服务器的离线复位，开启此功能，芯片发生故障后，会进行热复位操作，让芯片恢复健康。

开启MindIE Motor推理任务的离线复位功能只需要将Ascend Device Plugin的启动参数“-hotReset”取值设置为“0”或“2”。

**表 1**  参数说明

<a name="table173461839165111"></a>
|参数|类型|默认值|说明|
|--|--|--|--|
|-hotReset|int|-1|设备热复位功能参数。开启此功能，芯片发生故障后，Ascend Device Plugin会进行热复位操作，使芯片恢复健康。<ul><li>-1：关闭芯片复位功能</li><li>0：开启推理设备复位功能</li><li>1：开启训练设备在线复位功能</li><li>2：开启训练/推理设备离线复位功能</li></ul><span> 说明： </span><p>取值为1对应的功能已经日落，请配置其他取值。</p>该参数支持的训练设备：<ul><li>Atlas 800 训练服务器（型号 9000）（NPU满配）</li><li>Atlas 800 训练服务器（型号 9010）（NPU满配）</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD（型号 9000）</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A2 PoD 集群基础单元</li><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 800T A3 超节点服务器</li></ul>该参数支持的推理设备：<ul><li>Atlas 300I Pro 推理卡</li><li>Atlas 300V 视频解析卡</li><li>Atlas 300V Pro 视频解析卡</li><li>Atlas 300I Duo 推理卡</li><li>Atlas 300I 推理卡（型号 3000）（整卡）</li><li>Atlas 300I 推理卡（型号 3010）</li><li>Atlas 800I A2 推理服务器</li><li>A200I A2 Box 异构组件</li><li>Atlas 800I A3 超节点服务器</li></ul>|

>[!NOTE] 说明 
>Atlas 800I A2 推理服务器存在以下两种故障恢复方式，一台Atlas 800I A2 推理服务器只能使用一种故障恢复方式，由集群调度组件自动识别使用哪种故障恢复方式。
>-   方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。
>-   方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。


## 配置推理任务的弹性扩缩容<a name="ZH-CN_TOPIC_0000002479226430"></a>

MindIE Motor推理任务中，用户可通过配置Job级别弹性扩缩容功能，在发生硬件或软件故障且当前资源不满足所有实例拉起时，降低运行的实例数量，尽量保证推理任务继续运行。在故障恢复或新的硬件加入时，等待拉起的Job实例会重新被调度。

**使用约束<a name="zh-cn_topic_0000002356673977_section270417201799"></a>**

当前仅支持MindIE Motor推理任务使用本功能。

**支持的产品型号<a name="zh-cn_topic_0000002356673977_section618313391397"></a>**

- Atlas 800I A2 推理服务器
- Atlas 800I A3 超节点服务器

**原理说明<a name="zh-cn_topic_0000002356673977_section1445672111019"></a>**

**图 1**  弹性扩缩容原理<a name="zh-cn_topic_0000002356673977_fig685814101278"></a>  
![](../../figures/scheduling/弹性扩缩容原理.png "弹性扩缩容原理")

1.  用户配置多个Job属于同一个推理任务，并将Job分成多个组别，并配置一个扩缩容规则（scaling-rule）。
2.  弹性扩缩容规则以ConfigMap形式部署在集群中，不同类别的实例对应scaling-rule中的不同group。例如可以将所有的Prefill实例分类为group0，所有的Decode实例分类为group1。
3.  配置重调度的场景下，当发生硬件或软件故障时，Ascend Device Plugin和NodeD对故障进行上报，Volcano删除该实例下的所有Pod。
4.  ClusterD将global-ranktable发送给MindIE Controller，关于global-ranktable的说明请参见[SubscribeRankTable](../api/clusterd.md#subscriberanktable)中"global-ranktable文件说明"表。
5.  MindIE Controller根据global-ranktable确定需要退出的实例，通知容器中的进程非0退出。
6.  Volcano-Scheduler感知到Pod异常后，将实例的所有Pod删除。
7.  Ascend Operator感知到Pod被删除后，会收集当前MindIE Motor对应scaling-rule下的所有实例运行情况。
8.  Ascend Operator根据scaling-rule确认当前实例是否需要创建Pod。
9.  如果可以创建Pod，待Pod创建完成后，由调度器完成调度或处于Pending状态等待调度。
10. 处于Pending状态的Pod待资源充足时，自动完成调度。
11. 如果当前不可以创建Pod，则等待其他实例成功运行后再进行创建。

**创建扩缩容规则ConfigMap<a name="zh-cn_topic_0000002356673977_section476902931213"></a>**

用户需要设置特定的扩缩容规则，将其以ConfigMap的形式部署到k8s集群中，示例如下

```
apiVersion: v1
data:
  elastic_scaling.json: |          # 固定字段，请勿修改
    {
      "version": "1.0",            # 固定字段，请勿修改
      "elastic_scaling_list": [    # 以下为模板，用户根据自身需求进行设置
        {
          "group_list": [                 # 一个可以正常运行的任务配比状态            {
              "group_name": "group0",      # 用户自行设置
              "group_num": "2",            # 用户自行设置，要求从上往下不能增加
              "server_num_per_group": "2"  # 用户自行设置，要求相同的group_name，该值保持不变
            },
            {
              "group_name": "group1",      
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        },
        {
          "group_list": [                 # 另一个可以正常运行的任务配比状态
            {
              "group_name": "group0",
              "group_num": "1",
              "server_num_per_group": "2"
            },
            {
              "group_name": "group1",
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        }
      ]
    }
kind: ConfigMap
metadata:
  name: scaling-rule              # 用户自行设置
  namespace: mindie-service       # 用户自行设置，与推理任务保持一致
```

>[!NOTE] 说明 
>-   例如当前运行的group\_name为group0和group1的Job都为0个，则会选择索引为1的group\_list，即group0和group1都需要运行1个，那么此时group0或group1对应的Job就会创建对应的Pod，然后等待调度。
>-   如果当前group\_name为group0的Job运行了1个，group\_name为group1的Job运行了0个，此时只会为group\_name为group1的Job创建Pod，group\_name为group0的Job会等待group\_name为group1的Job成功运行后才创建Pod。

在以上ConfigMap中，可以修改的字段说明如下表所示。

**表 2**  参数说明

<a name="zh-cn_topic_0000002356673977_zh-cn_topic_0000002193288232_table985012534578"></a>
|参数|说明|取值|是否必填|
|--|--|--|--|
|metadata.name|承载scaling-rule的ConfigMap的名称。<p>用户可以自行设置，Job的label“mind-cluster/scaling-rule”的值需要与之对应，表明该Job受该scaling-rule控制。</p>|string|是|
|metadata.namespace|承载scaling-rule的ConfigMap的命名空间。<p>用户可以自行设置，但需要与推理任务保持一致。如果不设置，那么命名空间默认是"default"。</p>|string|否|
|group_name|group组名称。<p>Job的label "mind-cluster/group-name"，需要与之对应，表明该Job属于该group组。</p>|string|是|
|group_num|group组目标Job数量。<p>若当前运行中的该group下的Job数量未达该目标，会尝试拉起该group下的一个Job。</p>|string|是|
|server_num_per_group|group组目标Job的副本数。<p>不同group_list中相同group_name下，该值需保持一致。</p>|string|是|

**修改扩缩容规则<a name="zh-cn_topic_0000002356673977_section1769411616405"></a>**

如果此时已经运行了2个group0和1个group1的Job，用户需要增加运行一个group0的Job，那么用户需提前修改扩缩容模板，再下发新的任务，修改示例如下：

```
apiVersion: v1
data:
  elastic_scaling.json: |          # 固定字段，请勿修改
    {
      "version": "1.0",            # 固定字段，请勿修改
      "elastic_scaling_list": [    # 以下为模板，用户根据自身需求进行设置
        {
          "group_list": [                    # 新增一个条目到elastic_scaling_list中             
            {
              "group_name": "group0",     
              "group_num": "3",              # 修改group0的group_num
              "server_num_per_group": "2"  
            },
            {
              "group_name": "group1",      
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        },
        {
          "group_list": [                             
            {
              "group_name": "group0",      
              "group_num": "2",            
              "server_num_per_group": "2"  
            },
            {
              "group_name": "group1",      
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        },
        {
          "group_list": [                 # 另一个可以正常运行的任务配比状态
            {
              "group_name": "group0",
              "group_num": "1",
              "server_num_per_group": "2"
            },
            {
              "group_name": "group1",
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        }
      ]
    }
kind: ConfigMap
metadata:
  name: scaling-rule              # 用户自行设置
  namespace: mindie-service       # 用户自行设置，与推理任务保持一致
```

如果在任务正常运行的情况下，需要减少其中一个group0的Job，用户需要修改模板，再删除任务，修改示例如下。

```
apiVersion: v1
data:
  elastic_scaling.json: |          # 固定字段，请勿修改
    {
      "version": "1.0",            # 固定字段，请勿修改
      "elastic_scaling_list": [    # 以下为模板，用户根据自身需求进行设置
        {
          "group_list": [                 # 删除了一个group_list
            {
              "group_name": "group0",
              "group_num": "1",           # group0目标group_num为"1"
              "server_num_per_group": "2"
            },
            {
              "group_name": "group1",
              "group_num": "1",
              "server_num_per_group": "2"
            }
          ]
        }
      ]
    }
kind: ConfigMap
metadata:
  name: scaling-rule              # 用户自行设置
  namespace: mindie-service       # 用户自行设置，与推理任务保持一致
```

**准备任务YAML<a name="zh-cn_topic_0000002356673977_zh-cn_topic_0000002098814658_section463203519254"></a>**

在任务YAML中，修改或新增以下字段，开启Job级别弹性扩缩容。

```
... 
metadata:  
   labels:  
     ...  
     fault-scheduling: "force"
     fault-retry-times: "100000000"    # 处理业务面故障，必须配置业务面无条件重试的次数
     jobID: mindie-xxx      # 由用户自行定义
     app: mindeie-ms-server
     mind-cluster/scaling-rule: scaling-rule   # 需与扩缩容规则ConfigMap的名称保持一致
     mind-cluster/group-name: group0           # 需与扩缩容规则ConfigMap中的group_name取值保持一致
spec:
  schedulerName: volcano      # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
  runPolicy:
    backoffLimit: 3         # 任务重调度次数
...
```

## 配置推理任务交换机亲和性

当前仅支持Atlas 800I A2 推理服务器配置交换机亲和性功能。开启此功能，可规避Spine交换机下行流量冲突问题。如需了解该功能的原理，请参见[交换机亲和性调度1.0](../references.md#交换机亲和性调度10)章节。

**前提条件**

已完成[（可选）使用Volcano交换机亲和性调度](../installation_guide.md#可选使用volcano交换机亲和性调度)。


**配置推理任务交换机亲和性**

将交换机亲和性tor-affinity配置为normal-schema，YAML示例如下：

```
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch        
    app: mindie-ms-server        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test        # 当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置
    tor-affinity: normal-schema    # 开启交换机亲和性
    ring-controller.atlas: ascend-910b
```

