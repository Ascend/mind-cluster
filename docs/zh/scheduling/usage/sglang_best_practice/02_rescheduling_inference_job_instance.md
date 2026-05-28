# 配置推理任务实例重调度<a name="ZH-CN_TOPIC_0000002480738948"></a>

当推理任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。如需了解故障的检测原理，请参见[故障检测](../resumable_training/01_solutions_principles.md#故障检测)
章节。

## 前提条件<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

已完成部署基于OME的SGLang推理服务。

## 实例重调度原理<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

**故障实例Pod的删除**

OME子工作负载为Deployment时（一个P/D实例由一个Pod组成）：

- 业务面故障：Pod所属的容器发生非零退出的情况下自动重拉。
- 硬件故障：Ascend Device Plugin或者NodeD上报硬件故障到ClusterD之后，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。

OME子工作负载为LeaderWorkerSet时（一个P/D实例由多个Pod组成）：

- 业务面故障：对于任意实例所属Pod的容器发生非零退出之后，LWS Controller自动删除实例所属整个PodGroup。
- 硬件故障：Ascend Device Plugin或者NodeD上报硬件故障到ClusterD之后，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。LWS
  Controller自动删除实例所属整个PodGroup。

**故障实例Pod的重新创建和调度**

Deployment或者LeaderWorkerSet所属的Pod被Volcano删除之后，由各自对应的Controller重新创建被删除的Pod，并由Volcano执行对恢复Pod的重新调度。

> [!NOTE]
> OME任务进行故障恢复时只会重调度故障的P/D实例。

## 配置实例级别重调度<a name="section96795436354"></a>

下面以ClusterServingRuntime为例配置实例级别重调度。

<pre codetype="yaml">
apiVersion: ome.io/v1beta1
kind: ClusterServingRuntime
metadata:
  name: lws-runtime
  annotations:
    sp-block: "16"
  labels:
    <strong>fault-scheduling: "force"          # 开启重调度功能</strong>
    <strong>pod-rescheduling: "on"             # 开启Pod级重调度</strong>
    <strong>fault-retry-times: "3"             # 开启业务面故障无条件重试能力</strong>
spec:
...</pre>

## 重调度功能验证<a name="section10786547021"></a>

以[通过脚本一键式部署使用](./01_deploying_ome_sglang_inference_job.md#通过脚本一键式部署使用)
方式下发PD分离推理任务（1P1D，PD实例不跨机）为例，成功下发任务后，通过如下命令查看对应pod的运行状态：

```shell
kubectl get pods -A
```

可以看到类似下面的Pod信息：

```
NAMESPACE   NAME                                READY   STATUS    RESTARTS   AGE
default     my-test-decoder-xxx-xxx             1/1     Running   0          <time>
default     my-test-engine-xxx-xxx              1/1     Running   0          <time>
default     my-test-mf-store-xxx-xxx            1/1     Running   0          <time>
default     my-test-router-xxx-xxx              1/1     Running   0          <time>
```

- 其中`<time>`为Pod运行时长。

此时若手动构造故障：

```shell
kubectl exec -it my-test-engine-xxx-xxx -- kill -9 <pid>
```

- 其中`<pid>`为容器内sglang进程ID

立即查看相关实例的信息，会发现my-test-engine-xxx-xxx的状态变为Error：

```
NAMESPACE   NAME                                READY   STATUS    RESTARTS   AGE
default     my-test-decoder-xxx-xxx             1/1     Running   0          <time>
default     my-test-engine-xxx-xxx              1/1     Error     0          <time>
default     my-test-mf-store-xxx-xxx            1/1     Running   0          <time>
default     my-test-router-xxx-xxx              1/1     Running   0          <time>
```

若实例级重调度配置正确，my-test-engine-xxx-xxx 会自动被重调度回Running状态，对应的<time>值会更新。
待sglang推理服务拉起后，可正常处理推理请求。
