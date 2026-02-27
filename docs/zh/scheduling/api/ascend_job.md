# Ascend Job<a name="ZH-CN_TOPIC_0000002479226878"></a>

Ascend Job：简称acjob，是MindCluster自定义的一种任务类型，当前支持通过环境变量配置资源信息及文件配置资源信息这2种方式拉起训练或推理任务。

**支持的AI框架<a name="zh-cn_topic_0000002377698613_section1580601414413"></a>**

-   MindSpore
-   TensorFlow
-   PyTorch

**样例<a name="zh-cn_topic_0000002377698613_section7389161784012"></a>**

pytorch\_multinodes\_acjob\_910b.yaml示例如下。

```
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: default-test-pytorch
  labels:
    framework: pytorch
    ring-controller.atlas: ascend-910b
    tor-affinity: "null" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不适用。large-model-schema表示大模型任务，normal-schema 普通任务
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
            host-arch: huawei-arm
            accelerator-type: card-910b-2 # depend on your device model, 910bx8 is module-910b-8 ,910bx16 is module-910b-16
          containers:
          - name: ascend # do not modify
            image: pytorch-test:latest         # trainning framework image， which can be modified
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
            host-arch: huawei-arm
            accelerator-type: card-910b-2 # depend on your device model, 910bx8 is module-910b-8 ,910bx16 is module-910b-16
          containers:
          - name: ascend # do not modify
            image: pytorch-test:latest                # trainning framework image， which can be modified
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

**关键字段<a name="zh-cn_topic_0000002377698613_section92451045124014"></a>**

**表 1**  acjob字段说明

|字段路径|类型|格式|描述|
|--|--|--|--|
|apiVersion|字符串 (string)|-|定义对象表示的版本化资源模式。服务器会转换为最新内部值，拒绝不识别的版本。 更多信息请参见<a href="https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds">Types</a>。|
|kind|字符串 (string)|-|表示此对象对应的REST资源类型。值通过端点推断，不可更新，采用驼峰命名。 更多信息请参见<a href="https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources">Resources</a>。|
|metadata|对象 (object)|-|Kubernetes元数据（如命名空间、标签等）。更多信息请参见<a href="https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata">Metadata</a>。|
|spec|对象 (object)|-|AscendJob期望状态的规格描述。必填字段：replicaSpecs。|
|spec.replicaSpecs|对象 (object)|-|ReplicaType到ReplicaSpec的映射，指定MS集群配置。示例：{ "Scheduler": ReplicaSpec, "Worker": ReplicaSpec }。|
|spec.replicaSpecs.[ReplicaType]|对象 (object)|-|副本的描述。|
|spec.replicaSpecs.[ReplicaType].replicas|整数 (integer)|int32|副本数量，表示给定模板所需的副本数。默认为1。|
|spec.replicaSpecs.[ReplicaType].restartPolicy|字符串 (string)|-|重启策略：Always、OnFailure、Never、ExitCode。默认为Never。|
|spec.replicaSpecs.[ReplicaType].template|对象 (object)|-|Kubernetes Pod模板，更多信息请参见<a href="https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-template-v1/">Kubernetes Pod模板</a>。|
|spec.runPolicy|对象 (object)|-|封装分布式训练作业的运行时策略（如资源清理、活动时间）。|
|spec.runPolicy.backoffLimit|整数 (integer)|int32|作业失败前允许的重试次数（可选）。|
|spec.runPolicy.activeDeadlineSeconds|整数 (integer)|int64|作业保持活动的最长时间（秒），值必须为正整数。当前无意义，后续版本将会删除。|
|spec.runPolicy.cleanPodPolicy|字符串 (string)|-|作业完成后清理Pod的策略。默认值为Running。当前无意义，后续版本将会删除。|
|spec.runPolicy.ttlSecondsAfterFinished|整数 (integer)|int32|作业完成后的TTL（生存时间）。默认为无限，实际删除可能延迟。当前无意义，后续版本将会删除。|
|spec.runPolicy.schedulingPolicy|对象 (object)|-|调度策略（如gang-scheduling）。|
|spec.runPolicy.schedulingPolicy.minAvailable|整数 (integer)|int32|最小可用资源数。|
|spec.runPolicy.schedulingPolicy.minResources|对象 (object)|-|按资源名称分配的最小资源集合（支持整数或字符串格式）。|
|spec.runPolicy.schedulingPolicy.priorityClass|字符串 (string)|-|优先级类名称。|
|spec.runPolicy.schedulingPolicy.queue|字符串 (string)|-|调度队列名称。|
|spec.schedulerName|字符串 (string)|-|指定在开启gang-scheduling情况下的调度器，当前仅支持Volcano。|
|spec.successPolicy|字符串 (string)|-|标记AscendJob成功的标准，当前无意义，仅当所有Pod成功时，才会判定任务成功。后续版本将会删除。|
|status|对象 (object)|-|AscendJob的最新观察状态（只读）。必填字段：conditions、replicaStatuses。|
|status.completionTime|字符串 (string)|date-time|作业完成时间（RFC3339格式，UTC）。|
|status.conditions|数组 (array)|-|当前作业条件数组。|
|status.conditions[type]|字符串 (string)|-|作业条件的类型（如 "Complete"）。|
|status.conditions[status]|字符串 (string)|-|条件状态：True、False、Unknown。|
|status.conditions[lastTransitionTime]|字符串 (string)|date-time|条件状态转换的时间。|
|status.conditions[lastUpdateTime]|字符串 (string)|date-time|条件更新后的最终时间。|
|status.conditions[message]|字符串 (string)|-|条件的详细描述。|
|status.conditions[reason]|字符串 (string)|-|条件转换的原因。|
|status.lastReconcileTime|字符串 (string)|date-time|作业最后一次调和的时间（RFC3339格式，UTC）。|
|status.replicaStatuses|对象 (object)|-|副本类型到副本状态的映射。|
|status.replicaStatuses.[ReplicaType].active|整数 (integer)|int32|正在运行的Pod数量。|
|status.replicaStatuses.[ReplicaType].failed|整数 (integer)|int32|已失败的Pod数量。|
|status.replicaStatuses.[ReplicaType].succeeded|整数 (integer)|int32|已成功的Pod数量。|
|status.replicaStatuses.[ReplicaType].labelSelector|对象 (object)|-|Pod标签选择器（定义如何筛选Pod）。|
|status.replicaStatuses.[ReplicaType].labelSelector.matchExpressions|数组 (array)|-|标签匹配规则（支持In、NotIn、Exists、DoesNotExist等操作符）。|
|status.replicaStatuses.[ReplicaType].labelSelector.matchLabels|对象 (object)|-|标签匹配的键值对（等价于matchExpressions条件）。|
|status.startTime|字符串 (string)|date-time|作业开始时间（RFC3339格式，UTC）。|


**任务状态说明<a name="zh-cn_topic_0000002377698613_section177175313294"></a>**

拉起训练任务后，用户可以通过**kubectl get acjob**命令查看acjob任务的运行状态，当前运行状态有以下几种。

**表 2**  acjob任务运行状态说明

|状态名称|说明|
|--|--|
|Created|Job已经创建，但其中一个或多个子资源(Pod/Service)尚未就绪。|
|Running|Job的所有子资源(Pod/Service)已经调度并启动。|
|Restarting|Job的一个或多个子资源(Pod/Service)运行失败，但是根据重启策略正在重新启动。|
|Succeeded|Job的所有子资源(Pod/Service)处于成功终止阶段。|
|Failed|Job的一个或多个子资源(Pod/Service)运行失败。|

**任务异常条件说明<a name="zh-cn_topic_0000002377698613_section177175313295"></a>**

当任务出现异常时，AscendJob 的 status.conditions 字段会记录详细的异常信息。每个 condition 包含以下字段：

|字段|类型|说明|
|--|--|--|
|type|字符串|条件类型，如 Failed、Restarting、Running、Succeeded、Created|
|status|字符串|条件状态：True、False、Unknown|
|lastTransitionTime|字符串|条件状态转换的时间（RFC3339格式）|
|lastUpdateTime|字符串|条件更新后的最终时间（RFC3339格式）|
|message|字符串|条件的详细描述信息|
|reason|字符串|条件转换的原因代码|

**常见异常原因（reason）说明**

|原因代码|说明|
|--|--|
|JobFailed|任务失败，通常是因为有 Pod 失败|
|jobRestarting|任务正在重启，根据重启策略重新启动失败的 Pod|
|ExitedWithCode|任务以非零退出码退出|
|FailedDeleteJob|删除任务失败|
|SuccessfulDeleteJob|删除任务成功|
|SyncPodGroupFailed|同步 PodGroup 失败|
|PodGroupNotInitialized|PodGroup 未初始化，通常是因为 volcano-scheduler 未运行|
|PodGroupPending|PodGroup 处于等待状态，通常是因为集群资源不足|
|SyncServiceFailed|同步 Service 失败|
|PodCreateFailed|创建 Pod 失败|
|ArgumentError|参数错误|
|InvalidScalingConfig|无效的扩缩容配置|
|InvalidScaleOutConfig|无效的扩容配置|
|InvalidSpecs|无效的规格配置|
|InvalidReplicaType|无效的副本类型|
|InvalidSuccessPolicy|无效的成功策略|
|InvalidQueue|无效的队列配置|
|InvalidFramework|无效的框架配置|
|InvalidReplicaSpec|无效的副本规格配置|
|InvalidContainer|无效的容器配置|
|JobValidFailed|任务验证失败|

**异常条件示例**

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

**查看任务异常信息**

使用以下命令查看任务的详细状态和异常信息：

```bash
# 查看 AscendJob 的状态
kubectl get acjob <job-name> -o yaml

# 查看 AscendJob 的状态摘要
kubectl get acjob <job-name> -o jsonpath='{.status.conditions}'

# 查看 AscendJob 的最新状态
kubectl get acjob <job-name> -o jsonpath='{.status.conditions[-1]}'
```


