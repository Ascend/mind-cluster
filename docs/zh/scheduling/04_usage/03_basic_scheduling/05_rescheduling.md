# 重调度<a name="ZH-CN_TOPIC_0000002479387124"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002479387116"></a>

重调度特性指的是当任务运行过程中发生硬件故障或软件故障时，集群调度组件能够将任务重新调度到新的健康节点或芯片上，以确保任务能够重新正常运行。
重调度模式默认为Job级别重调度，即每次故障停止所有Pod，重新创建并重调度所有Pod后，重启任务。

**前提条件<a name="section166381652174516"></a>**

- 使用重调度特性，需要确保已经安装如下组件。若没有安装，可以参考[安装部署](../../03_installation_guide/02_installation/00_helm_installation.md)章节进行操作。
    - Volcano（本特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - Ascend Operator（使用AscendJob必须安装）
    - ClusterD
    - NodeD
    - Infer Operator（使用InferServiceSet任务必须安装）

**使用约束<a name="section1178044918127"></a>**

- 本功能仅支持在6.0.RC2及以上版本中使用。
- 大规模K8s集群场景下，ConfigMap映射时延不可控，建议RankTable使用共享存储方式。

**支持的产品形态<a name="section169961844182917"></a>**

支持以下产品使用**重调度**。

- <term>Atlas 推理系列产品</term>
- <term>Atlas 训练系列产品</term>
- <term>Atlas A2 推理系列产品</term>
- <term>Atlas A2 训练系列产品</term>
- <term>Atlas A3 推理系列产品</term>
- <term>Atlas A3 训练系列产品</term>
- <term>Ascend 950 系列产品</term>

**重调度原理<a name="section57901137171110"></a>**

任务运行过程中如果出现了软硬件故障，将导致任务状态异常。Job级别重调度首先销毁所有的任务容器，然后隔离故障设备，再重新将任务容器调度启动。任务容器重新启动后重新拉起任务，该行为类似任务首次拉起过程。

**图 1**  原理图<a name="fig18343114924113"></a>

![Job重调度原理图](../../../figures/scheduling/原理图.png "原理图")

在以上原理图中，各个步骤的说明如下：

1. 检测到故障后，首先删除当前任务所有的Pod和容器。
2. 隔离故障所在的设备，防止再次使用该设备。
3. 重新创建和调度任务Pod和容器。
4. 容器启动后，拉起任务进程恢复任务。

## 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002511427039"></a>

本章节的重调度特性配置在整卡调度基础上增加相应的配置项。

### 制作镜像<a name="ZH-CN_TOPIC_0000002511427053"></a>

详细请参见整卡调度中的[制作镜像](./03_full_npu_scheduling.md#制作镜像)。

### 脚本适配<a name="ZH-CN_TOPIC_0000002479227172"></a>

详细请参见整卡调度中的[脚本适配](./03_full_npu_scheduling.md#脚本适配)。

### 配置Job级别重调度<a name="配置job级别重调度"></a>

Job级别重调度默认开启。在[整卡调度](./03_full_npu_scheduling.md#准备任务yaml)的任务YAML基础上，在`metadata.labels`中新增以下字段，开启Job级别重调度。

```yaml
...
metadata:
  labels:
    ...
    fault-scheduling: "grace"      # 可以根据实际情况选择force或者grace，配置为force时Pod不建议使用主机网络，配置force时可能存在重调度失败并触发多次重调度直到成功的现象。
    fault-retry-times: 100         # 配置无条件尝试次数，软件故障场景需要配置。
```

fault-scheduling配置项取值含义如下。

**表 2**  fault-scheduling配置项值列表

<a name="table0396162644916"></a>

|参数|取值|含义|
|--|--|--|
|fault-scheduling|grace|任务使用重调度开关，并在过程中先优雅删除原Pod。|
|fault-scheduling|force|配置任务采用强制删除模式，在过程中强制删除原Pod。|

### 下发任务<a name="ZH-CN_TOPIC_0000002511427027"></a>

详细请参见整卡调度中的[下发任务](./03_full_npu_scheduling.md#下发任务)。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002511427025"></a>

详细请参见整卡调度中的[查看任务进程](./03_full_npu_scheduling.md#查看任务进程)。

### 验证Job级别重调度<a name="验证job级别重调度"></a>

1. 构造故障。

   1. 在任务运行节点上查询任务进程。

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      回显示例如下：

      ```ColdFusion
      2398104
      2398105
      2398107
      ```

   2. 终止进程模拟故障。

      ```bash
      kill -9 2398104
      ```

2. 观察重调度过程。

   监控该Job的Pod状态变化。

   ```bash
   kubectl get pod -n <namespce> -o wide -w | grep <pod-name>
   ```

   该Job的Pod历史状态示例如下，观察加粗字段的变化可以发现该Job的Pod会经历Terminating→Pending→ContainerCreating→Running阶段，然后正常运行，表示Job重调度成功：

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running             0          2s      xx.xx.xx.xx       node173                 &lt;none&gt;           &lt;none&gt;
   // ===================== 注入故障 ======================
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Running</strong>             0          45s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   </pre>

## 集成后使用

参考整卡调度中的[集成后使用](./03_full_npu_scheduling.md#集成后使用)，在其基础上加上重调度的配置项即可。
