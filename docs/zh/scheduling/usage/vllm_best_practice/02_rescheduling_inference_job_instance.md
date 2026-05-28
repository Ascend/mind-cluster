# 配置推理任务实例重调度<a name="ZH-CN_TOPIC_0000002484224852"></a>

## 前提条件<a name="section12668751507"></a>

已完成AIBrix服务部署，详细请参见[AIBrix文档](https://aibrix.readthedocs.io/latest/designs/aibrix-stormservice.html)。

## 支持的故障类型<a name="section17669195209"></a>

芯片、软件故障

## 重调度原理<a name="section06701958011"></a>

AIBrix根据任务YAML中的role实例生成PodGroup，对应实例发生故障时，重调度实例PodGroup下的所有Pod，若实例配置的podGroupSize均配置为1，只会生成一个PodGroup，发生故障时重调度对应实例的故障Pod。

## 配置实例级重调度<a name="section86725515019"></a>

以StormService YAML为例配置实例级重调度，添加以下加粗部分配置。

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
                <strong>#pod-rescheduling: "on"   # 若podGroupSize均为1则需配置该标签，podGroupSize大于1时，无需配置</strong>
                <strong>fault-retry-times: "10"</strong>
            spec:
              <strong>schedulerName: volcano  # 指定调度器</strong>
              <strong>restartPolicy: Never</strong>
              nodeSelector:
                accelerator-type: module-910b-8
              containers:
                - name: prefill
...
                  resources:
                    limits:
                      huawei.com/Ascend910: 8  # 配置所需NPU数
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
                <strong>#pod-rescheduling: "on"   # 若podGroupSize均为1则需配置该标签，podGroupSize大于1时，无需配置</strong>
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

## 重调度功能验证<a name="section96726516020"></a>

以示例yaml中的配置为例：

- prefill实例：`replicas=1`，`podGroupSize=2`
- decode实例：`replicas=1`，`podGroupSize=2`

成功下发任务后，通过如下命令查看对应pod的运行状态：

```shell
kubectl get pods -A
```

可以看到类似下面的prefill和decode实例的信息：

```
NAMESPACE   NAME                                                  READY   STATUS    RESTARTS   AGE
default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-0      1/1     Running   0          <time>
default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1      1/1     Running   0          <time>
default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-0        1/1     Running   0          <time>
default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-1        1/1     Running   0          <time>
```

- 其中`<id>`为roleset标识，`<prefill-ins>`为prefill实例索引，`<decode-ins>`为decode实例索引，`<time>`为Pod运行时长。

此时若手动构造故障：

```shell
kubectl exec -it vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1 -- kill -9 <pid>
```

- 其中`<pid>`为容器内vllm进程ID

立即查看prefill和decode实例的信息，会发现vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1的状态变为Error：

```
NAMESPACE   NAME                                                  READY   STATUS    RESTARTS   AGE
default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-0      1/1     Running   0          <time>
default     vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1      1/1     Error     0          <time>
default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-0        1/1     Running   0          <time>
default     vllm-1p1d-roleset-<id>-decode-<decode-ins>-0-1        1/1     Running   0          <time>
```

若实例级重调度配置正确，vllm-1p1d-roleset-<id>-prefill-<prefill-ins>-0-1 与 vllm-1p1d-roleset-<id>-prefill-<prefill-ins>
-0-0 会自动被重调度为Running状态，对应的<time>值会更新。
待vllm推理服务拉起后，可正常处理推理请求。
