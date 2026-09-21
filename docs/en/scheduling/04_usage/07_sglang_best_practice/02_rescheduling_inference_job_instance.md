# Configuring Rescheduling of Inference Job Instances<a name="ZH-CN_TOPIC_0000002480738948"></a>

When node, chip, or other faults occur in an inference job, the MindCluster scheduling components can isolate the faulty resources and automatically reschedule them. For details about the fault detection principles, see [Fault Detection](../04_resumable_training/01_solutions_principles.md#fault-detection).

## Prerequisites<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

The OME-based SGLang inference service has been deployed.

## Instance Rescheduling Principles<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

**Deletion of Faulty Instance Pods**

When the OME sub-workload is `Deployment` (one Prefill/Decoded instance consists of one Pod):

- Service plane fault: When a container to which the Pod belongs exits with a non-zero code, it is automatically restarted.
- Hardware fault: After Ascend Device Plugin or NodeD reports a hardware fault to ClusterD, Volcano obtains the faulty node, deletes the Pods on the node, and isolates the faulty node.

When the OME sub-workload is `LeaderWorkerSet` (one Prefill/Decode instance consists of multiple Pods):

- Service plane fault: When a container of any Pod belonging to an instance exits with a non-zero code, LWS Controller automatically deletes the entire PodGroup of the instance.
- Hardware fault: After Ascend Device Plugin or NodeD reports a hardware fault to ClusterD, Volcano obtains the faulty node, deletes the Pods on the node, and isolates the faulty node.
LWS Controller automatically deletes the entire PodGroup of the instance.

**Recreating and Rescheduling Faulty Instance Pods**

After the Pods belonging to a `Deployment` or `LeaderWorkerSet` are deleted by Volcano, the corresponding Controller recreates the deleted Pods, and Volcano reschedules the recovered Pods.

> [!NOTE]
> During fault recovery of an OME job, only the faulty Prefill/Decode instance is rescheduled.

## Configuring Instance-level Rescheduling<a name="section96795436354"></a>

The following uses `ClusterServingRuntime` as an example to configure instance-level rescheduling.

<pre codetype="yaml">
apiVersion: ome.io/v1beta1
kind: ClusterServingRuntime
metadata:
  name: lws-runtime
  annotations:
    sp-block: "16"
    huawei.com/schedule_policy: "chip2-node16-sp"    #  Set scheduling policy based on hardware forms
  labels:
    <strong>fault-scheduling: "force"          # Enable rescheduling feature</strong>
    <strong>pod-rescheduling: "on"             # Enable Pod-level rescheduling</strong>
    <strong>fault-retry-times: "3"             # Enable unconditional retry capability for service-layer faults</strong>
spec:
...</pre>

## Verifying the Rescheduling Feature<a name="section10786547021"></a>

Taking the [deployment using a one-click script](./01_deploying_ome_sglang_inference_job.md#deploying-inference-jobs-using-a-script-in-one-click-mode) to deploy a PD-disaggregated inference job (1P1D, PD instances not cross-node) as an example, after the job is successfully deployed, perform the following operations:

1. Check the running status of the corresponding Pods.

    ```shell
    kubectl get pods -A
    ```

    You can see Pod information similar to the following:

    ```ColdFusion
    NAMESPACE   NAME                                READY   STATUS    RESTARTS   AGE
    default     my-test-decoder-xxx-xxx             1/1     Running   0          <time>
    default     my-test-engine-xxx-xxx              1/1     Running   0          <time>
    default     my-test-mf-store-xxx-xxx            1/1     Running   0          <time>
    default     my-test-router-xxx-xxx              1/1     Running   0          <time>
    ```

    `<time>` is the Pod running duration.
2. Manually simulate a fault.

    ```shell
    kubectl exec -it my-test-engine-xxx-xxx -- kill -9 <pid>
    ```

    `<pid>` is the SGLang process ID inside the container.

3. Immediately check the information of the relevant instances.

    ```shell
    kubectl get pods -A
    ```

   You can see that `my-test-engine-xxx-xxx` has changed to the `Error` status:

    ```ColdFusion
    NAMESPACE   NAME                                READY   STATUS    RESTARTS   AGE
    default     my-test-decoder-xxx-xxx             1/1     Running   0          <time>
    default     my-test-engine-xxx-xxx              1/1     Error     0          <time>
    default     my-test-mf-store-xxx-xxx            1/1     Running   0          <time>
    default     my-test-router-xxx-xxx              1/1     Running   0          <time>
    ```

If the instance-level rescheduling is configured correctly, `my-test-engine-xxx-xxx` will be automatically rescheduled back to the `Running` status, and the corresponding `<time>` value will be updated. After the SGLang inference service is successfully started, inference requests can be processed normally.
