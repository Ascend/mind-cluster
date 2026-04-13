# 配置推理任务实例重调度<a name="ZH-CN_TOPIC_0000002480738948"></a>

当推理任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。如需了解故障的检测原理，请参见[故障检测](../resumable_training.md#故障检测)章节。

**前提条件<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>**

已完成部署基于OME的SGLang推理服务。

**实例重调度原理<a name="zh-cn_topic_0000002356060805_section4253197539"></a>**

**故障实例Pod的删除**

OME子工作负载为Deployment时（一个P/D实例由一个Pod组成）：

- 业务面故障：Pod所属的容器发生非零退出的情况下自动重拉。
- 硬件故障：Ascend Device Plugin或者NodeD上报硬件故障到ClusterD之后，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。

OME子工作负载为LeaderWorkerSet时（一个P/D实例由多个Pod组成）：

- 业务面故障：对于任意实例所属Pod的容器发生非零退出之后，LWS Controller自动删除实例所属整个PodGroup。
- 硬件故障：Ascend Device Plugin或者NodeD上报硬件故障到ClusterD之后，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。LWS Controller自动删除实例所属整个PodGroup。

**故障实例Pod的重新创建和调度**

Deployment或者LeaderWorkerSet所属的Pod被Volcano删除之后，由各自对应的Controller重新创建被删除的Pod，并由Volcano执行对恢复Pod的重新调度。

>[!NOTE] 
>OME任务进行故障恢复时只会重调度故障的P/D实例。

**配置实例级别重调度<a name="section96795436354"></a>**

下面以ClusterServingRuntime为例配置实例级别重调度。

<pre codetype="yaml">
apiVersion: ome.io/v1beta1
kind: ClusterServingRuntime
metadata:
  name: lws-runtime
  annotations:
    sp-block: "16"
  labels:
    <strong>fault-scheduling: "force"          # 开启重调度功能
    pod-rescheduling: "on"             # 开启Pod级重调度
    fault-retry-times: "3"             # 开启业务面故障无条件重试能力</strong>
spec:
...</pre>
