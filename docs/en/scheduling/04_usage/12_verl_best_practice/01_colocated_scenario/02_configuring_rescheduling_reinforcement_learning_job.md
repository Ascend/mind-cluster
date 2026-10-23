<!-- markdownlint-disable MD033 -->
# Configuring Pod Rescheduling for Reinforcement Learning Jobs<a name="ZH-CN_TOPIC_0000002480738958"></a>

When a node, chip, or other fault occurs in a verl reinforcement learning job, the MindCluster cluster scheduling components can isolate the faulty resource and automatically perform rescheduling. This feature requires coordination on the verl framework side to complete resource cleanup when a fault occurs and to continue executing the reinforcement learning job after the resources are rescheduled.

This example implements the preceding function through `fault_recover` in `verl-ascend-recipe`. This recipe is intended for the RL training-inference colocated deployment scenario, that is, the training and inference roles are deployed on the same NPU resources, and the RL algorithm supports only GRPO. After perceiving a software or hardware fault, `fault_recover` can automatically perform fault recovery and resume training based on persisted data. For details about the recipe and the fault detection principle, see [verl-ascend-recipe/fault_recover](https://github.com/verl-project/verl-ascend-recipe/tree/main/fault_recover) and [Fault Detection](../../04_fault_detection_and_diagnosis/01_working_principle.md).

## Prerequisites<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

The procedure in [Deploying a verl Reinforcement Learning Job](./01_deploying_verl_reinforcement_learning_job.md) has been completed.

## Pod Rescheduling Principle<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

### Deleting Pods of a Faulty Instance

When a hardware fault such as NPU loss occurs on a node where a job is running, Ascend Device Plugin or NodeD reports the hardware fault to ClusterD. Volcano obtains the faulty node, deletes the Pods on the node, and isolates the faulty node.

### Recreating and Scheduling Pods of a Faulty Instance

After the Pod to which the fault belongs is deleted by Volcano, it is recreated and rescheduled to join the resource pool of the reinforcement learning job. verl attempts to rebuild the worker group based on the current resources and resume training. The inference side performs token-level continuation based on the historically saved CKPT, reducing the fault recovery cost. For details about the token-level continuation principle, see [RFC](https://github.com/verl-project/verl/discussions/4355).

## Configuring Pod Rescheduling<a name="section96795436354"></a>

The following uses the example YAML in [Deploying a verl Reinforcement Learning Job](./01_deploying_verl_reinforcement_learning_job.md) to configure the rescheduling capability, which has been enabled in the previous section. The specific parameters are described as follows.

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindspeed-rl
  labels:
    framework: pytorch
    <strong>fault-scheduling: "force"          # Enable rescheduling</strong>
    <strong>pod-rescheduling: "on"             # Enable Pod-level rescheduling</strong>
    <strong>fault-retry-times: "3"             # Enable unconditional retry upon service plane faults; the value is the rescheduling count</strong>
spec:
...</pre>

## Verifying the Rescheduling Function<a name="section10786547021"></a>

1. Check the Pod running status.

    ```shell
    kubectl get pods -A
    ```

    The following is the output example. If the `STATUS` field is `Running`, the Pod is running properly:

    ```ColdFusion
    NAMESPACE   NAME                          READY   STATUS    RESTARTS   AGE
    default     verl-test-master-0            1/1     Running   0          <time>
    default     verl-test-worker-0            1/1     Running   0          <time>
    default     verl-test-worker-1            1/1     Running   0          <time>
    ```

    `<time>` is the Pod running duration.

2. Manually construct a fault.

    ```shell
    kubectl exec -it verl-test-worker-0 -- kill -9 <pid>
    ```

    `<pid>` is the PID of the verl process in the container.

3. Immediately check the information about the related instance.

    ```shell
    kubectl get pods -A
    ```

    You will find that the status of `verl-test-worker-0` changes to `Error`:

    ```ColdFusion
    NAMESPACE   NAME                          READY   STATUS    RESTARTS   AGE
    default     verl-test-master-0            1/1     Running   0          <time>
    default     verl-test-worker-0            1/1     Error     0          <time>
    default     verl-test-worker-1            1/1     Running   0          <time>
    ```

    If the rescheduling configuration is correct, `verl-test-worker-0` is automatically rescheduled and returns to the `Running` state, and the corresponding `<time>` value is updated. The verl framework then rebuilds the worker group. After it is started properly, training can continue, and the inference job can be continued from the last CKPT.
