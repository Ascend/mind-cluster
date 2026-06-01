# 配置推理任务重调度<a name="ZH-CN_TOPIC_0000002479386400"></a>

当推理任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。如需了解故障的检测原理，请参见[故障检测](../resumable_training/01_solutions_principles.md#故障检测)章节。

## 前提条件<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

已完成[部署MindIE Motor](./01_deploying_mindie_motor.md)。

## 支持的故障类型<a name="section121201333144919"></a>

- MindIE Server：节点、芯片或其他故障
- MindIE MS：节点故障

## 重调度原理<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

- Job级别重调度：MindIE Server和MindIE MS均支持。当MindIE Server或MindIE MS发生故障时，对应的MindIE Server实例或MindIE MS停止所有Pod，重新创建并重调度所有Pod后，最新的global-ranktable.json重新推送给MS Controller，推理任务被重启。

    MindIE Server在PD分离的场景下，例如MindIE Server包含一个Prefill实例和一个Decode实例，Prefill实例发生故障，仅停止Prefill实例的所有Pod，不会影响其他正常运行的实例。

- Pod级别重调度：仅MindIE MS支持。在开启主备倒换功能场景下，MS Controller或MS Coordinator对应的Pod数量均大于1，当某节点发生故障时，仅停止该节点对应的Pod。例如，MS Coordinator包含主MS Coordinator和备MS Coordinator，主MS Coordinator发生故障时，仅停止主MS Coordinator对应的Pod，不会影响备MS Coordinator。

    >[!NOTE]
    >若Pod级别重调度恢复失败，则会回退到Job级别重调度处理方式。

## 配置Job级别重调度<a name="zh-cn_topic_0000002356060805_section20633874524"></a>

Job级别重调度默认开启，用户只需完成准备任务YAML的步骤即可。下面以MindIE Server为例说明Job级别重调度的配置。

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-server        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test        # 当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置
    <strong>fault-scheduling: force      # 开启重调度功能</strong>
    fault-retry-times: "10000"     # 开启业务面故障重调度，值为业务面故障时的重调度次数
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
  runPolicy:
    schedulingPolicy:      # Ascend Operator启用“gang”调度生效，且调度器为Volcano时，本字段才生效
      minAvailable: 2      # 任务运行总副本数
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:</pre>

### 下发任务<a name="ZH-CN_TOPIC_0000002511427027"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```shell
kubectl apply -f XXX.yaml
```

例如：

```shell
kubectl apply -f infer-job.yaml
```

回显示例如下：

```ColdFusion
ascendjob.mindxdl.gitee.com/mindie-server-0 created
```

>[!NOTE]
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f _XXX_.yaml命令删除原任务，再重新下发任务。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002511427025"></a>

执行以下命令，查看Pod运行状况。

```shell
kubectl get pod --all-namespaces
```

回显示例如下：

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-server-master-0                     1/1     Running   0          20m
...
```

### 查看推理卡故障重调度结果<a name="ZH-CN_TOPIC_0000002511347069"></a>

当推理任务运行中出现故障时(可以通过业务代码主动触发报错)，Volcano会将该任务调度到其他NPU上。

执行以下命令，查看任务运行状况。

```shell
kubectl get pod --all-namespaces
```

查看任务Pod，可以看到任务被删除之后重新创建，原Pod状态Error后被删除，新的任务Pod被创建，Pod状态从Pending到ContainerCreating再到Running， AGE从0秒开始，表示故障重调度特性运行成功。

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-server-0-master-0                   1/1     Running   0          1s
...
```

### 删除任务<a name="ZH-CN_TOPIC_0000002479387108"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```shell
kubectl delete -f XXX.yaml
```

例如：

```shell
kubectl delete -f infer-job.yaml
```

回显示例如下：

```ColdFusion
ascendjob.mindxdl.gitee.com "mindie-server-0" deleted
```

## 配置Pod级别重调度<a name="section5620411141"></a>

Pod级别重调度目前只支持MS Controller和MS Coordinator，建议在开启主备倒换功能场景下使用。下面以MS Coordinator开启主备倒换功能为例说明Pod级别重调度的配置。

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-coordinator
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-coordinator        # 表示MindIE Motor在Ascend Job任务中的角色,不可修改
    jobID: mindie-ms-test             # 当前MindIE Motor推理任务在集群中的唯一识别ID，用户可根据实际情况进行配置
    <strong>fault-scheduling: force          # 开启重调度功能</strong>
    <strong>pod-rescheduling: "on"           # 开启Pod级别重调度</strong>
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # Ascend Operator启用“gang”调度时所选择的调度器
  runPolicy:
    schedulingPolicy:      # Ascend Operator启用“gang”调度生效，且调度器为Volcano时，本字段才生效
      <strong>minAvailable: 2      # 任务运行总副本数</strong>
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:</pre>

### 下发任务<a name="ZH-CN_TOPIC_0000002511427027"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发多Pod的推理任务（只有在多Pod任务的场景下才能在k8s中看到与Job级重调度的区别）。

```shell
kubectl apply -f XXX.yaml
```

例如：

```shell
kubectl apply -f infer-job.yaml
```

回显示例如下：

```ColdFusion
ascendjob.mindxdl.gitee.com/mindie-coordinator created
```

>[!NOTE]
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f _XXX_.yaml命令删除原任务，再重新下发任务。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002511427025"></a>

执行以下命令，查看Pod运行状况。

```shell
kubectl get pod --all-namespaces
```

回显示例如下：

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-coordinator-master-0                1/1     Running   0          20m
default          mindie-coordinator-worker-1                1/1     Running   0          20m
...
```

### 查看推理卡故障重调度结果<a name="ZH-CN_TOPIC_0000002511347069"></a>

当推理任务运行中某个Pod出现故障时(可以通过业务代码主动触发报错)，Volcano会将该Pod调度到其他NPU上。

执行以下命令，查看任务运行状况。

```shell
watch -n 1 kubectl get pod --all-namespaces
```

查看任务pod，可以看到报错Pod被删除之后重新创建，报错Pod状态Error后被删除，新的Pod被创建，Pod状态从Pending到ContainerCreating再到Running， AGE从0秒开始，表示故障重调度特性运行成功。

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-coordinator-master-0                1/1     Running   0          20m
default          mindie-coordinator-worker-1                1/1     Running   0          1s
...
```

### 删除任务<a name="ZH-CN_TOPIC_0000002479387108"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```shell
kubectl delete -f XXX.yaml
```

例如：

```shell
kubectl delete -f infer-job.yaml
```

回显示例如下：

```ColdFusion
ascendjob.mindxdl.gitee.com "mindie-coordinator" deleted
```
