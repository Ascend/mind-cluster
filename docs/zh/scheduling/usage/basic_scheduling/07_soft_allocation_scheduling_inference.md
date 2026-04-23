# 软切分调度（推理）<a name="ZH-CN_TOPIC_0000002511428569"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002511347125"></a>

**前提条件**

在命令行场景下使用软切分调度特性，需要确保已经安装如下组件；若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。软切分调度特性只支持使用Volcano作为调度器，不支持使用其他调度器。

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**使用方式**

软切分调度特性的使用方式如下：

- 通过命令行使用：安装集群调度组件，通过命令行使用软切分调度特性。
- 集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明**

- 软切分调度特性需要搭配算力虚拟化特性一起使用，关于软切分虚拟化的相关说明和操作请参见[软切分虚拟化](../virtual_instance/virtual_instance_with_vcann_rt/01_soft_allocation_virtualization.md)章节。
- 软切分调度仅支持下发单副本数或者多副本数的单机任务，每个副本独立工作，不支持分布式任务。

**支持的产品形态**

- Atlas A2 推理系列产品
- Atlas A3 推理系列产品

**使用流程**

通过命令行使用软切分调度特性流程可以参见[图1](#fig24252498666vcann)。

**图 1**  使用流程<a name="fig24252498666vcann"></a>  
![](../../../figures/scheduling/basic_scheduling_001.png "basic_scheduling_001")

软切分虚拟化实例涉及到相关集群调度组件的参数配置，请参见[软切分虚拟化](../virtual_instance/virtual_instance_with_vcann_rt/01_soft_allocation_virtualization.md)章节完成修改。

## 实现原理

目前仅支持acjob任务类型，其原理图如[图1](#fig23698010123)所示。

**图 1**  acjob任务调度原理图<a name="fig23698010123"></a>  
![](../../../figures/scheduling/basic_scheduling_002.PNG "basic_scheduling_002")

各步骤说明如下：

1. 集群调度组件定期上报节点和芯片信息。
    - kubelet上报节点芯片数量到节点对象（node）中。
    - Ascend Device Plugin定期上报芯片拓扑信息。

        上报软切分NPU信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片百分比总量（allocatable）、已使用的芯片百分比数量（allocated）和芯片的基础信息（device ip和super\_device\_ip）上报到Node中，用于软切分调度。

    - 当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2. ClusterD读取device-info-cm和node-info-cm中的信息后，将信息写入cluster-info-cm。
3. 用户通过kubectl或者其他深度学习平台下发acjob任务。
4. Ascend Operator为任务创建相应的PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5. Ascend Operator为任务创建相应的Pod，并在容器中注入集合通信所需环境变量。
6. volcano-scheduler根据节点的芯片AI Core百分比总量和芯片高带宽内存总量以及该节点上已部署Pod的annotation已使用信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
7. kubelet创建容器时，调用Ascend Device Plugin挂载芯片及芯片共享所需文件，Ascend Device Plugin或volcano-scheduler在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源。

## 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_00000024792271456"></a>

### 制作镜像<a name="ZH-CN_TOPIC_0000002511427026"></a>

**获取推理镜像**

可选择以下方式中的一种来获取推理镜像。

- 推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据系统架构（ARM或者x86\_64）下载**推理基础镜像（**如：[ascend-infer](https://www.hiascend.com/developer/ascendhub/detail/e02f286eef0847c2be426f370e0c2596)、[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)**）**。

    请注意，21.0.4版本之后推理基础镜像默认用户为非root用户，需要在下载基础镜像后对其进行修改，将默认用户修改为root。

    >[!NOTE]  
    >基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。

- （可选）可基于推理基础镜像定制用户自己的推理镜像，制作过程请参见[使用Dockerfile构建推理镜像](../../common_operations.md#使用dockerfile构建推理镜像)。

    完成定制化修改后，用户可以给推理镜像重命名，以便管理和使用。

**加固镜像**

下载或者制作的推理基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../../security_hardening.md#容器安全加固)章节进行操作。

### 脚本适配<a name="ZH-CN_TOPIC_000000251134706701"></a>

本章节以昇腾镜像仓库中推理镜像为例为用户介绍操作流程，该镜像已经包含了推理示例脚本，实际推理场景需要用户自行准备推理脚本。在拉取镜像前，需要确保当前环境的网络代理已经配置完成，确保该环境可以正常访问昇腾镜像仓库。

**从昇腾镜像仓库获取示例脚本<a name="section8181015175911"></a>**

1. 确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2. 在左侧导航栏选择推理镜像，然后选择[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)镜像，获取推理示例脚本。

    >[!NOTE]  
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。

### 准备任务YAML<a name="ZH-CN_TOPIC_00000024793871220102"></a>

>[!NOTE]  
>如果用户不使用Ascend Docker Runtime组件，Ascend Device Plugin只会帮助用户挂载“/dev”目录下的设备。其他目录（如“/usr”）用户需要自行修改YAML文件，挂载对应的驱动目录和文件。容器内挂载路径和宿主机路径保持一致。
>因为Atlas 200I SoC A1 核心板场景不支持Ascend Docker Runtime，用户也无需修改YAML文件。

**操作步骤<a name="zh-cn_topic_0000001558853680_zh-cn_topic_0000001609074213_section14665181617334"></a>**

1. 获取相应的YAML文件。

    **表 1**  YAML说明

    |任务类型|硬件型号|YAML名称|获取链接|
    |--|--|--|--|
    |Ascend Job|<ul><li>Atlas A2 推理系列产品</li><li>Atlas A3 推理系列产品</li></ul>|pytorch_acjob_infer_<i>\{xxx\}</i>b_softsharedev.yaml|[获取YAML](https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/inference/volcano/pytorch_acjob_infer_910b_softsharedev.yaml)|

2. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    在Atlas 800I A2 推理服务器上，以pytorch_acjob_infer_910b_softsharedev.yaml为例，申请芯片AI Core百分比为50%，芯片高带宽内存量为2048MB，软切分策略为fixed-share的参数配置示例如下。

    <pre>
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: default-infer-test-pytorch-910b
      labels:
        framework: pytorch
        ring-controller.atlas: ascend-910b
        fault-scheduling: "force"
        <strong>huawei.com/scheduler.softShareDev.aicoreQuota: "50" # 软切分任务请求的芯片AI Core百分比，单位为%
        huawei.com/scheduler.softShareDev.hbmQuota: "2048" # 软切分任务请求的芯片高带宽内存量，单位为MB
        huawei.com/scheduler.softShareDev.policy: "fixed-share" # 软切分策略，取值为fixed-share、elastic和best-effort
      annotations:
        huawei.com/schedule_policy: "chip1-softShareDev" # 软切分场景Volcano调度策略</strong>
    spec:
      schedulerName: volcano   # work when enableGangScheduling is true
      runPolicy:
        schedulingPolicy:      # work when enableGangScheduling is true
          minAvailable: 1
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
              automountServiceAccountToken: false
              nodeSelector:
                host-arch: huawei-arm
                accelerator-type: module-910b-8 # depend on your device model, 910bx8 is module-910b-8 ,910bx16 is module-910b-16
              containers:
                - name: ascend # do not modify
                  image: pytorch-test:latest         # trainning framework image， which can be modified
                  imagePullPolicy: IfNotPresent
                  env:
                    - name: XDL_IP                                       # IP address of the physical node, which is used to identify the node where the pod is running
                      valueFrom:
                        fieldRef:
                          fieldPath: status.hostIP
                  command:                           # training command,  which can be modified
                    - /bin/bash
                    - -c
                  args: [ "./infer.sh" ]
                  ports:                          # default value       containerPort: 2222 name: ascendjob-port if not set
                    - containerPort: 2222         # determined by user
                      name: ascendjob-port        # do not modify
                  resources:
                    requests:
                      <strong>huawei.com/Ascend910: 50 # 此处需要与huawei.com/scheduler.softShareDev.aicoreQuota的值保持一致，表示软切分任务请求的AI Core百分比</strong>
                    limits:
                      <strong>huawei.com/Ascend910: 50 # 数值与requests保持一致</strong>
                  volumeMounts:
                    - name: ascend-driver
                      mountPath: /usr/local/Ascend/driver
                    - name: ascend-add-ons
                      mountPath: /usr/local/Ascend/add-ons
                    - name: localtime
                      mountPath: /etc/localtime
                    <strong>- name: libpreload # 软切分动态库地址
                      mountPath: /opt/enpu/vcann-rt/lib/libvruntime.so
                    - name: preload # preload配置文件地址
                      mountPath: ${preload_path}/ld.so.preload</strong>
              volumes:
                - name: ascend-driver
                  hostPath:
                    path: /usr/local/Ascend/driver
                - name: ascend-add-ons
                  hostPath:
                    path: /usr/local/Ascend/add-ons
                - name: localtime
                  hostPath:
                    path: /etc/localtime
                <strong>- name: libpreload # 软切分动态库地址
                  hostPath:
                    path: /opt/enpu/vcann-rt/lib/libvruntime.so
                - name: preload # preload配置文件地址
                  hostPath:
                    path: ${preload_path}/ld.so.preload</strong>
    </pre>

    **表 2**  pytorch_acjob_infer_910b_softsharedev.yaml参数说明

    |参数|取值|说明|
    |--|--|--|
    |huawei.com/scheduler.softShareDev.aicoreQuota|[1, 100]|请求的AI Core百分比。|
    |huawei.com/scheduler.softShareDev.hbmQuota|<p>[1, maxHBM]</p><p>maxHBM为通过<b>npu-smi info</b>命令查询出的HBM-Usage(MB)中HBM的值。</p>|请求的高带宽内存量，单位为MB。|
    |huawei.com/scheduler.softShareDev.policy|<ul><li>fixed-share</li><li>elastic</li><li>best-effort</li></ul>|软切分策略。|
    |huawei.com/schedule_policy|chip1-softShareDev|软切分场景调度策略。|

### 下发任务<a name="ZH-CN_TOPIC_000000247922713402"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```shell
kubectl apply -f XXX.yaml
```

例如：

```shell
kubectl apply -f pytorch_acjob_infer_910b_softsharedev.yaml
```

回显示例如下：

```ColdFusion
ascendjob.mindxdl.gitee.com/default-infer-test-pytorch-910b created
```

>[!NOTE]  
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f <i>XXX</i>.yaml命令删除原任务，再重新下发任务。

### 查看任务进程<a name="ZH-CN_TOPIC_00000025113470710203"></a>

**操作步骤**

1. <a name="ZH-CN_TOPIC_00000025113470710203step01"></a>执行以下命令，查看Pod运行状况。

    ```shell
    kubectl get pod --all-namespaces
    ```

    回显示例如下：

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
    ...
    default         default-infer-test-pytorch-910b-master-0    1/1     Running   0          8s
    ...
    ```

2. 查看运行推理任务的节点详情。
    1. 执行以下命令查看节点的名称。

        ```shell
        kubectl get node -A
        ```

    2. 根据上一步骤中查询到的节点名称，执行以下命令查看节点详情。

        ```shell
        kubectl describe node <nodename>
        ```

        回显示例如下：

        ```ColdFusion
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests     Limits
          --------              --------     ------
          cpu                   4 (2%)       3500m (1%)
          memory                2140Mi (0%)  4040Mi (0%)
          ephemeral-storage     0 (0%)       0 (0%)
          huawei.com/Ascend910  50           50
        Events:
          Type    Reason    Age   From                Message
          ----    ------    ----  ----                -------
          Normal  Starting  36m   kube-proxy, ubuntu  Starting kube-proxy.
        ...
        ```

        在显示的信息中，找到“Allocated resources”下的**huawei.com/Ascend910**，该参数取值在执行推理任务之后会增加，增加数量为推理任务使用的NPU芯片的AI Core百分比总量。

### 查看软切分调度结果<a name="ZH-CN_TOPIC_000000247938712002"></a>

**操作步骤**

在管理节点执行以下命令查看推理结果。

```shell
kubectl logs -f default-infer-test-pytorch-910b-master-0
```

回显示例如下，以实际回显为准。

```ColdFusion
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:config.c:145] Success to load config: physical-npu-id, value: 2
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:config.c:145] Success to load config: virtual-npu-id, value: 0
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:config.c:145] Success to load config: aicore-quota, value: 100
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:config.c:145] Success to load config: memory-quota, value: 60000
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:config.c:145] Success to load config: shm-id, value: C281A66C-80A047F2-0A645632-CC500485-100301E3
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:config.c:145] Success to load config: scheduling-policy, value: 2
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:npu-manager.c:127] Successfully to initialize vnpu device.
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:mem-limiter.c:69] create /run/enpu/vcann-rt/ success
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281460942893344:core-limiter.c:290] The scheduling process has been detected to exit, and the scheduling is being taken over.
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:npu-manager.c:168] Successfully to initialize all module.
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:memory.c:91] Hook mem rtMemGetInfoEx.
[20260304150146] [INFO] [eNPU] [vCANN_RT] [1799:281472853921824:memory.c:91] Hook mem rtMemGetInfoEx.
```

>[!NOTE] 
><i>default-infer-test-pytorch-910b-master-0</i>：查看任务进程章节[步骤1](#ZH-CN_TOPIC_00000025113470710203step01)中运行的任务名称。

### 删除任务<a name="ZH-CN_TOPIC_00000025113470650102"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```shell
kubectl delete -f XXX.yaml
```

例如：

```shell
kubectl delete -f pytorch_acjob_infer_910b_softsharedev.yaml
```

回显示例如下：

```ColdFusion
root@ubuntu:/home/test/yaml# kubectl delete -f pytorch_acjob_infer_910b_softsharedev.yaml 
ascendjob.mindxdl.gitee.com "default-infer-test-pytorch-910b" deleted
```

## 集成后使用<a name="ZH-CN_TOPIC_00000025113470730102"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1. 根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2. 根据K8s的官方API库，对任务进行创建、查询、删除等操作。
3. 创建、查询或删除任务时，用户需要将[示例YAML](#准备任务yaml)的内容转换成K8s官方API中定义的对象，通过官方API发送给K8s的API Server或者将YAML内容转换成JSON格式直接发送给K8s的API Server。
