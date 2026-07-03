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
                <strong>#pod-rescheduling: "on"   # If podGroupSize is 1, this label must be configured. If podGroupSize is greater than 1, no configuration is needed.</strong>
                <strong>fault-retry-times: "10"</strong>
            spec:
              <strong>schedulerName: volcano  # Specify the scheduler</strong>
              <strong>restartPolicy: Never</strong>
              nodeSelector:
                accelerator-type: module-910b-8
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
                <strong>#pod-rescheduling: "on"   # If podGroupSize is 1, configure this label. If podGroupSize is greater than 1, no configuration is needed.</strong>
                <strong>fault-retry-times: "10"</strong>
            spec:
              nodeSelector:
                accelerator-type: module-910b-8
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
