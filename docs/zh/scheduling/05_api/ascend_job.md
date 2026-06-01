# Ascend Job<a name="ZH-CN_TOPIC_0000002479226878"></a>

Ascend Job：简称acjob，是MindCluster自定义的一种任务类型，当前支持通过环境变量配置资源信息及文件配置资源信息两种方式拉起训练或推理任务。

## 支持的AI框架<a name="zh-cn_topic_0000002377698613_section1580601414413"></a>

- MindSpore
- PyTorch

## 样例<a name="zh-cn_topic_0000002377698613_section7389161784012"></a>

pytorch\_multinodes\_acjob\_910b.yaml示例如下。

```Yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: default-test-pytorch
  labels:
    framework: pytorch
    ring-controller.atlas: ascend-910b
    tor-affinity: "null" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不适用。large-model-schema表示大模型任务，normal-schema 普通任务
  annotations:
      huawei.com/schedule_policy: "chip8-node8"
spec:
  schedulerName: volcano   # work when enableGangScheduling is true
  runPolicy:
    schedulingPolicy:      # work when enableGangScheduling is true
      minAvailable: 2
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:
      replicas: 1
      restartPolicy: Never
      template:
        metadata:
          labels:
            ring-controller.atlas: ascend-910b
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                      - key: job-name
                        operator: In
                        values:
                          - default-test-pytorch
                  topologyKey: kubernetes.io/hostname
          nodeSelector:
            example-key: example-value    # 示例值，用户可根据调度意图自行配置nodeSelector
          containers:
          - name: ascend # do not modify
            image: pytorch-test:latest         # training framework image， which can be modified
            imagePullPolicy: IfNotPresent
            env:
              - name: XDL_IP                                       # IP address of the physical node, which is used to identify the node where the pod is running
                valueFrom:
                  fieldRef:
                    fieldPath: status.hostIP
              # ASCEND_VISIBLE_DEVICES env variable is used by ascend-docker-runtime when in the whole card scheduling scene with volcano scheduler.
              # Please delete it when in the static vNPU scheduling, dynamic vNPU scheduling, volcano without Ascend-volcano-plugin, without volcano scenes.
              - name: ASCEND_VISIBLE_DEVICES
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # The value must be the same as resources.requests
            command:                           # training command, which can be modified
              - /bin/bash
              - -c
            args: [ "cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code /job/output main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=4096" ]
            ports:                          # default value containerPort: 2222 name: ascendjob-port if not set
              - containerPort: 2222         # determined by user
                name: ascendjob-port        # do not modify
            resources:
              limits:
                huawei.com/Ascend910: 2
              requests:
                huawei.com/Ascend910: 2
            volumeMounts:
            - name: code
              mountPath: /job/code
            - name: data
              mountPath: /job/data
            - name: output
              mountPath: /job/output
            - name: ascend-driver
              mountPath: /usr/local/Ascend/driver
            - name: ascend-add-ons
              mountPath: /usr/local/Ascend/add-ons
            - name: dshm
              mountPath: /dev/shm
            - name: localtime
              mountPath: /etc/localtime
          volumes:
          - name: code
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/"
          - name: data
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/dataset/"
          - name: output
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/output/"
          - name: ascend-driver
            hostPath:
              path: /usr/local/Ascend/driver
          - name: ascend-add-ons
            hostPath:
              path: /usr/local/Ascend/add-ons
          - name: dshm
            emptyDir:
              medium: Memory
          - name: localtime
            hostPath:
              path: /etc/localtime
    Worker:
      replicas: 1
      restartPolicy: Never
      template:
        metadata:
          labels:
            ring-controller.atlas: ascend-910b
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                      - key: job-name
                        operator: In
                        values:
                          - default-test-pytorch
                  topologyKey: kubernetes.io/hostname
          nodeSelector:
            example-key: example-value    # 示例值，用户可根据调度意图自行配置nodeSelector
          containers:
          - name: ascend # do not modify
            image: pytorch-test:latest                # training framework image， which can be modified
            imagePullPolicy: IfNotPresent
            env:
              - name: XDL_IP                                       # IP address of the physical node, which is used to identify the node where the pod is running
                valueFrom:
                  fieldRef:
                    fieldPath: status.hostIP
              # ASCEND_VISIBLE_DEVICES env variable is used by ascend-docker-runtime when in the whole card scheduling scene with volcano scheduler.
          # Please delete it when in the static vNPU scheduling, dynamic vNPU scheduling, volcano without Ascend-volcano-plugin, without volcano scenes.
              - name: ASCEND_VISIBLE_DEVICES
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # The value must be the same as resources.requests
            command:                                  # training command, which can be modified
              - /bin/bash
              - -c
            args: ["cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code /job/output main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=4096"]
            ports:                          # default value containerPort: 2222 name: ascendjob-port if not set
              - containerPort: 2222         # determined by user
                name: ascendjob-port        # do not modify
            resources:
              limits:
                huawei.com/Ascend910: 2
              requests:
                huawei.com/Ascend910: 2
            volumeMounts:
            - name: code
              mountPath: /job/code
            - name: data
              mountPath: /job/data
            - name: output
              mountPath: /job/output
            - name: ascend-driver
              mountPath: /usr/local/Ascend/driver
            - name: ascend-add-ons
              mountPath: /usr/local/Ascend/add-ons
            - name: dshm
              mountPath: /dev/shm
            - name: localtime
              mountPath: /etc/localtime
          volumes:
          - name: code
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/"
          - name: data
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/dataset/"
          - name: output
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/output/"
          - name: ascend-driver
            hostPath:
              path: /usr/local/Ascend/driver
          - name: ascend-add-ons
            hostPath:
              path: /usr/local/Ascend/add-ons
          - name: dshm
            emptyDir:
              medium: Memory
          - name: localtime
            hostPath:
              path: /etc/localtime

```

## 任务状态说明<a name="zh-cn_topic_0000002377698613_section177175313294"></a>

拉起训练任务后，用户可以通过**kubectl get acjob**命令查看acjob任务的运行状态，当前运行状态有以下几种。

**表 2**  acjob任务运行状态说明

|状态名称|说明|
|--|--|
|Created|Job已经创建，但其中一个或多个子资源(Pod/Service)尚未就绪。|
|Running|Job的所有子资源(Pod/Service)已经调度并启动。|
|Restarting|Job的一个或多个子资源(Pod/Service)运行失败，但是根据重启策略正在重新启动。|
|Succeeded|Job的所有子资源(Pod/Service)处于成功终止阶段。|
|Failed|Job的一个或多个子资源(Pod/Service)运行失败。|

## 任务异常条件说明<a name="zh-cn_topic_0000002377698613_section177175313295"></a>

当任务出现异常时，AscendJob 的 status.conditions 字段会记录详细的异常信息。每个 condition 包含以下字段：

|字段|类型|说明|
|--|--|--|
|type|字符串|条件类型，如 Failed、Restarting、Running、Succeeded、Created|
|status|字符串|条件状态：True、False、Unknown|
|lastTransitionTime|字符串|条件状态转换的时间（RFC3339格式）|
|lastUpdateTime|字符串|条件更新后的最终时间（RFC3339格式）|
|message|字符串|条件的详细描述信息|
|reason|字符串|条件转换的原因代码|

## 常见异常原因（reason）说明

|原因代码|说明|
|--|--|
|JobFailed|任务失败，通常是因为有 Pod 失败|
|jobRestarting|任务正在重启，根据重启策略重新启动失败的 Pod|
|SyncPodGroupFailed|同步 PodGroup 失败|
|PodGroupNotInitialized|PodGroup 未初始化，通常是因为 volcano-scheduler 未运行|
|PodGroupPending|PodGroup 处于等待状态，通常是因为集群资源不足|
|SyncServiceFailed|同步 Service 失败|
|PodCreateFailed|创建 Pod 失败|
|JobValidFailed|任务验证失败|

## 异常条件示例

```yaml
status:
  conditions:
  - type: Failed
    status: "True"
    lastTransitionTime: "2024-01-01T10:00:00Z"
    lastUpdateTime: "2024-01-01T10:00:00Z"
    message: "Job default/test-job has failed because has pod failed."
    reason: "JobFailed"
  - type: Restarting
    status: "True"
    lastTransitionTime: "2024-01-01T10:00:00Z"
    lastUpdateTime: "2024-01-01T10:00:00Z"
    message: "Job default/test-job is unconditional retry job and remain retry times is <3>."
    reason: "jobRestarting"
  - type: Failed
    status: "True"
    lastTransitionTime: "2024-01-01T10:00:00Z"
    lastUpdateTime: "2024-01-01T10:00:00Z"
    message: "Job test-job has failed because it has reached the specified backoff limit"
    reason: "JobFailed"
```

## 查看任务异常信息

使用以下命令查看任务的详细状态和异常信息：

```bash
# 查看 AscendJob 的状态
kubectl get acjob -n <namespace> <job-name> -o yaml

# 查看 AscendJob 的状态摘要
kubectl get acjob -n <namespace> <job-name> -o jsonpath={.status.conditions}

# 查看 AscendJob 的最新状态
kubectl get acjob -n <namespace> <job-name> -o jsonpath={.status.conditions[-1]}
```
