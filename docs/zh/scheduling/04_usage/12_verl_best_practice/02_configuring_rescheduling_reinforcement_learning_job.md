<!-- markdownlint-disable MD033 -->
# 配置强化学习任务Pod重调度<a name="ZH-CN_TOPIC_0000002480738958"></a>

当verl强化学习任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。该功能需要verl框架侧配合，在故障发生时完成资源清理，并在新资源重调度后，继续执行强化学习任务。

本示例通过verl-ascend-recipe中的fault_recover实现上述功能，该recipe面向RL训推同步共卡部署场景，即训练和推理角色部署在相同NPU资源上，RL算法仅支持GRPO。fault_recover在感知到软件或硬件故障后能够自动进行故障恢复并基于持久化数据续训，recipe详情请参见[verl-ascend-recipe/fault_recover](https://github.com/verl-project/verl-ascend-recipe/tree/main/fault_recover)。如需了解故障的检测原理，请参见[故障检测](../11_fault_detection_and_diagnosis/01_working_principle.md)章节。

## 前提条件<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

已完成[部署基于verl的强化学习任务](./01_deploying_verl_reinforcement_learning_job.md)。

## Pod重调度原理<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

### 故障实例Pod的删除

当任务运行的某个节点发生如NPU掉卡等硬件故障时，Ascend Device Plugin或者NodeD上报硬件故障到ClusterD，Volcano获取到故障节点，删除节点上的Pod，并隔离故障节点。

### 故障实例Pod的重新创建和调度

故障所属的Pod被Volcano删除之后，重新创建Pod并尝试重新调度，加入强化学习任务资源池。verl会尝试基于当前资源重建worker group，并恢复训练，其中推理侧会根据历史保存的CKPT进行token级续推，减少故障恢复代价，token级续推原理可参考[RFC](https://github.com/verl-project/verl/discussions/4355)。

## 配置Pod重调度<a name="section96795436354"></a>

下面以[部署强化学习任务](./01_deploying_verl_reinforcement_learning_job.md)中的示例YAML为例配置重调度能力，上一节示例YAML中已开启了重调度能力，下面针对具体参数作说明。

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindspeed-rl
  labels:
    framework: pytorch
    <strong>fault-scheduling: "force"          # 开启重调度功能</strong>
    <strong>pod-rescheduling: "on"             # 开启Pod级重调度</strong>
    <strong>fault-retry-times: "3"             # 开启业务面故障无条件重试功能，值为重调度次数</strong>
spec:
...</pre>

## 验证重调度功能<a name="section10786547021"></a>

1. 查看Pod运行状态。

    ```shell
    kubectl get pods -A
    ```

    回显示例如下，STATUS字段为Running，表示Pod运行正常：

    ```ColdFusion
    NAMESPACE   NAME                          READY   STATUS    RESTARTS   AGE
    default     verl-test-master-0            1/1     Running   0          <time>
    default     verl-test-worker-0            1/1     Running   0          <time>
    default     verl-test-worker-1            1/1     Running   0          <time>
    ```

    其中`<time>`为Pod运行时长。

2. 手动构造故障。

    ```shell
    kubectl exec -it verl-test-worker-0 -- kill -9 <pid>
    ```

    其中`<pid>`为容器内verl进程PID。

3. 立即查看相关实例的信息。

    ```shell
    kubectl get pods -A
    ```

    会发现verl-test-worker-0的状态变为Error：

    ```ColdFusion
    NAMESPACE   NAME                          READY   STATUS    RESTARTS   AGE
    default     verl-test-master-0            1/1     Running   0          <time>
    default     verl-test-worker-0            1/1     Error     0          <time>
    default     verl-test-worker-1            1/1     Running   0          <time>
    ```

    若重调度配置正确，verl-test-worker-0会自动被重调度回Running状态，对应的\<time>值会更新。随后verl框架会完成worker group重新拉起，正常拉起后可继续进行训练，其中推理任务可从上一次CKPT处进行续推。
