# Configuring Inference Job Instance Rescheduling<a name="ZH-CN_TOPIC_0000002484224852"></a>

## Prerequisites<a name="section12668751507"></a>

AIBrix service deployment has been completed. For details, see [AIBrix Documentation](https://aibrix.readthedocs.io/latest/designs/aibrix-stormservice.html).

## Supported Fault Types<a name="section17669195209"></a>

Chip and software faults

## Rescheduling Principle<a name="section06701958011"></a>

AIBrix generates a PodGroup based on the role instances in the job YAML. When a corresponding instance fails, all Pods under the instance's PodGroup are rescheduled. If the `podGroupSize` of the instance is configured as `1`, only one PodGroup is generated, and only the failed Pod of the corresponding instance is rescheduled when a fault occurs.

## Configuring Instance-level Rescheduling<a name="section86725515019"></a>

Take the `StormService` YAML as an example to configure instance-level rescheduling. Add the following configuration in bold.

<pre codetype="yaml">
apiVersion: orchestration.aibrix.ai/v1alpha1
kind: StormService
metadata:
  name: vllm-1p1d
spec:
  replicas: 1
  updateStrategy:
    type: InPlaceUpdate
  stateful: true
  selector:
    matchLabels:
      app: vllm-1p1d
  template:
    metadata:
      labels:
        app: vllm-1p1d
    spec:
      roles:
        - name: prefill
          replicas: 1
          stateful: true
          podGroupSize: 2
          template:
            metadata:
              labels:
                model.aibrix.ai/name: qwen3-8B
                model.aibrix.ai/port: "8000"
                model.aibrix.ai/engine: vllm
                <strong>fault-scheduling: "force"</strong>
                <strong>#pod-rescheduling: "on"   # If podGroupSize is 1 for all roles, this label needs to be configured; when podGroupSize is greater than 1, no configuration is required</strong>
                <strong>fault-retry-times: "10"</strong>
              annotations:
                <strong>huawei.com/schedule_policy: "chip2-node8"</strong>
            spec:
              <strong>schedulerName: volcano  # Specify the scheduler</strong>
              <strong>restartPolicy: Never</strong>
              nodeSelector:
                example-key: example-value    # Example value; users can configure nodeSelector according to scheduling intent
              containers:
                - name: prefill
...
                  resources:
                    limits:
                      huawei.com/Ascend910: 8  # Configure the required number of NPUs
                    requests:
                      huawei.com/Ascend910: 8
                  securityContext:
...
        - name: decode
          replicas: 1
          podGroupSize: 2
          stateful: true
          template:
            metadata:
              labels:
                model.aibrix.ai/name: qwen3-8B
                model.aibrix.ai/port: "8000"
                model.aibrix.ai/engine: vllm
                <strong>fault-scheduling: "force"</strong>
                <strong>#pod-rescheduling: "on"   # If podGroupSize is 1 for all roles, this label needs to be configured; when podGroupSize is greater than 1, no configuration is required</strong>
                <strong>fault-retry-times: "10"</strong>
              annotations:
                <strong>huawei.com/schedule_policy: "chip2-node8"</strong>
            spec:
              nodeSelector:
                example-key: example-value    # Example value; users can configure nodeSelector according to scheduling intent
              <strong>schedulerName: volcano</strong>
              <strong>restartPolicy: Never</strong>
              containers:
                - name: decode
...
                  resources:
                    limits:
                      huawei.com/Ascend910: 8
                    requests:
                      huawei.com/Ascend910: 8
                  securityContext:
...</pre>

## Verifying the Rescheduling Feature<a name="section96726516020"></a>

Taking the configuration in the example YAML as an example:

- Prefill instances: `replicas=1`, `podGroupSize=2`
- Decode instances: `replicas=1`, `podGroupSize=2`

After the task is successfully deployed, perform the following operations:

1. Check the running status of the corresponding Pods.

    ```shell
    kubectl get pods -A
    ```

    You can see Prefill and Decode instance information similar to the following:

    ```ColdFusion
    NAMESPACE   NAME                                                  READY   STATUS    RESTARTS   AGE
    default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-0      1/1     Running   0          <time>
    default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1      1/1     Running   0          <time>
    default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-0        1/1     Running   0          <time>
    default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-1        1/1     Running   0          <time>
    ```

    `<id>` is the Roleset identifier; `<prefill-ins>` is the Prefill instance index; `<decode-ins>` is the Decode instance index; `<time>` is the Pod running duration.
2. Manually simulate a fault.

    ```shell
    kubectl exec -it vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1 -- kill -9 <pid>
    ```

    `<pid>` is the vLLM process ID inside the container.
3. Immediately check the information of the Prefill and Decode instances.

    ```shell
    kubectl get pods -A
    ```

    You can see that `vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1` has changed to the `Error` status:

    ```ColdFusion
    NAMESPACE   NAME                                                  READY   STATUS    RESTARTS   AGE
    default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-0      1/1     Running   0          <time>
    default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1      1/1     Error     0          <time>
    default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-0        1/1     Running   0          <time>
    default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-1        1/1     Running   0          <time>
    ```

If the instance-level rescheduling is configured correctly, `vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1` and `vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-0` will be automatically rescheduled to the Running status, and the corresponding `<time>` values will be updated. After the vLLM inference service is successfully started, inference requests can be processed normally.
