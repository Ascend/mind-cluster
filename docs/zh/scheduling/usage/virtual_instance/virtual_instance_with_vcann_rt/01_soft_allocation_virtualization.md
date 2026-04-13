# 软切分虚拟化<a name="ZH-CN_TOPIC_000000968vcann"></a>

**使用软切分NPU说明<a name="ZH-CN_TOPIC_00000025113463450356vcann"></a>**

在Kubernetes场景下，当用户需要使用NPU资源时，需要结合集群调度组件Ascend Device Plugin和Volcano的使用，使Kubernetes可以管理并调度昇腾处理器资源。昇腾软切分虚拟化实例特性需要的集群调度组件包括Ascend Device Plugin、Volcano、Ascend Operator和ClusterD。支持的产品型号请参见[表1 产品支持情况说明](./00_description.md)。

**场景说明<a name="section1576110260450vcann"></a>**

使用软切分虚拟化前，需要提前了解[表1](#table62551184461989657)中的场景说明。

**表 1**  场景说明

<a name="table62551184461989657"></a>
<table><thead align="left"><tr><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p>场景</p>
</th>
<th class="cellrowborder" valign="top" width="80.02%" id="mcps1.2.3.1.2"><p>说明</p>
</th>
</tr>
</thead>
<tbody><tr><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p>通用说明</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p>分配的芯片信息会在PodGroup的label中体现出来，关于PodGroup label的详细说明请参见<a href="../../../api/volcano.md#podgroup">PodGroup label</a>中的如下参数：<ul><li>huawei.com/scheduler.softShareDev.aicoreQuota</li><li>huawei.com/scheduler.softShareDev.hbmQuota</li><li>huawei.com/scheduler.softShareDev.policy</li></ul></p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>软切分功能必须配合vCANN-RT使用。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>分配软切分NPU时，经MindCluster调度，将优先占满剩余算力最少的物理NPU。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>目前任务的每个Pod请求的NPU数量为1个。物理上使用的NPU数量为1，但任务YAML中请求的NPU数量需要与huawei.com/scheduler.softShareDev.aicoreQuota配置保持一致。</p>
</td>
</tr>
<tr><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p>特性支持的场景</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p>支持多副本，但多副本中的每个Pod所使用的NPU软切分策略必须一致。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>支持K8s的机制，如亲和性等。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>支持芯片故障和节点故障的重调度。具体参考<a href="../../basic_scheduling.md#推理卡故障恢复">推理卡故障恢复</a>和<a href="../../basic_scheduling.md#推理卡故障重调度">推理卡故障重调度</a>章节。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>支持集群中软切分虚拟化功能和非软切分虚拟化功能混合部署的场景。</p>
</td>
</tr>
<tr><td class="cellrowborder" rowspan="3" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p>特性不支持的场景</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p>不支持不同芯片在一个任务内混用。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>任务运行过程中，不支持卸载Volcano。</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>不支持与Docker场景的操作混用。</p>
</td>
</tr>
</tbody>
</table>

**前提条件**

1. 需要在节点上增加标签huawei.com/scheduler.chip1softsharedev.enable=true，表示该节点支持软切分功能。

    ```shell
    kubectl label nodes 节点名称 huawei.com/scheduler.chip1softsharedev.enable=true            
    ```

    在软切分虚拟化功能和非软切分虚拟化功能混合部署场景下，若节点不支持软切分虚拟化功能，则需要为节点增加标签huawei.com/scheduler.chip1softsharedev.enable=false。

2. 需要先获取“Ascend-docker-runtime\_\{version\}\_linux-\{arch\}.run”，安装容器引擎插件。
3. 参见[安装部署](../../../installation_guide/03_installation.md)章节，完成各组件的安装。

    虚拟化实例涉及修改相关参数的集群调度组件为Ascend Device Plugin，请按如下要求修改并使用对应的YAML安装部署：

    1. 在device-plugin-volcano-v\{version\}.yaml中添加-shareDevCount=100 -softShareDevConfigDir=/share_device/，其中/share_device/由用户手动创建。当Atlas A3 推理系列产品使用软切分虚拟化功能时，需额外增加启动参数-useSingleDieMode=true。

       ```Yaml
       ...

               args: [ "device-plugin  -useAscendDocker=true -volcanoType=true -presetVirtualDevice=true
                 -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=0 -shareDevCount=100 -softShareDevConfigDir=/share_device/ -useSingleDieMode=true" ]   # 只有Atlas A3 推理系列产品使用软切分虚拟化功能时，才需增加-useSingleDieMode=true
             ...
               volumeMounts:
             ...
                 - name:  enpu-config-dir                                       
                   mountPath: /etc/enpu/
                 - name: share-device-config-dir                                     
                   mountPath: /share_device/
           ...   
       volumes:
             ...
         - name: enpu-config-dir                             
           hostPath:
             path: /etc/enpu/
         - name: share-device-config-dir
           hostPath:
             path: /share_device/  
             type: DirectoryOrCreate             
       ```

        软切分虚拟化实例启动参数说明如下：

       **表 3** Ascend Device Plugin启动参数

       <a name="table1064314568229"></a>

       |参数|类型|默认值|说明|
       |--|--|--|--|
       |-shareDevCount|uint|1|使用软切分虚拟化功能时，值只能为100。|
       |-softShareDevConfigDir|string|""|软切分虚拟化场景配置目录。|
       |-useSingleDieMode|bool|false|Atlas A3 推理系列产品是否开启单die直通模式。<ul><li>true：开启单die直通模式。</li><li>false：关闭单die直通模式。</li></ul>使用软切分虚拟化功能时，该参数必须配置为true。|

    2. （可选）针对软切分虚拟化功能和非软切分虚拟化功能混合部署场景，需要对Ascend Device Plugin的YAML进行如下修改。

       - 在支持软切分虚拟化功能的节点上安装支持软切分功能的Ascend Device Plugin，将device-plugin-volcano-v\{version\}.yaml拷贝为softsharedev-device-plugin-volcano-v\{version\}.yaml。softsharedev-device-plugin-volcano-v\{version\}.yaml修改如下：

         ```Yaml
         apiVersion: apps/v1
         kind: DaemonSet
         metadata:
           name: ascend-device-plugin-daemonset-910-softShareDev #标识Ascend Device Plugin在软切分虚拟化功能和非软切分虚拟化功能混合部署场景下支持软切分虚拟化功能
           namespace: kube-system
         spec:
           ...
           template:
           ...
             spec:
             ...
               nodeSelector:
                 huawei.com/scheduler.chip1softsharedev.enable: "true"  #选择支持软切分虚拟化功能的节点部署Ascend Device Plugin
                 accelerator: huawei-Ascend910
               serviceAccountName: ascend-device-plugin-sa-910
               containers:
               ...
                 command: [ "/bin/bash", "-c", "--"]
                 args: [ "device-plugin  -useAscendDocker=true -volcanoType=true -presetVirtualDevice=true
                 -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=0 -shareDevCount=100 -softShareDevConfigDir=/share_device/" ]
               ...
                 volumeMounts:
               ...
                   - name: enpu-config-dir                                       
                     mountPath: /etc/enpu/
                   - name: share-device-config-dir                                   
                     mountPath: /share_device/
             ...   
         volumes:
               ...
           - name: enpu-config-dir                             
             hostPath:
               path: /etc/enpu/
           - name: share-device-config-dir
             hostPath:
               path: /share_device/  
               type: DirectoryOrCreate     
         ```

       - 在不支持软切分虚拟化功能的节点上安装原始的Ascend Device Plugin，device-plugin-volcano-v\{version\}.yaml修改如下：

         ```Yaml
         apiVersion: apps/v1
         kind: DaemonSet
         metadata:
           name: ascend-device-plugin-daemonset-910 #标识Ascend Device Plugin在软切分虚拟化功能和非软切分虚拟化功能混合部署场景下不支持软切分虚拟化功能
           namespace: kube-system
         spec:
           ...
           template:
           ...
             spec:
             ...
               nodeSelector:
                 huawei.com/scheduler.chip1softsharedev.enable: "false"  #选择不支持软切分虚拟化功能的节点部署Ascend Device Plugin
                 accelerator: huawei-Ascend910
               serviceAccountName: ascend-device-plugin-sa-910
           ...     
         ```

**使用方法**

创建推理任务时，需要在创建YAML文件时，修改如下配置。以Atlas 800I A2推理服务器为例。

申请芯片AI Core百分比为50%，芯片高带宽内存量为2048MB，软切分策略为fixed-share的参数配置示例如下。

<pre codetype="yaml">
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
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                      - key: job-name
                        operator: In
                        values:
                          - default-infer-test-pytorch-910b
                  topologyKey: kubernetes.io/hostname
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

>[!NOTE] 
>Atlas A3 推理系列产品下发软切分虚拟化任务时，在任务容器中，/dev/实际挂载1个die，但是执行<b>npu-smi info</b>命令查询显示挂载了2个die。回显示例如下：
>
> ```ColdFusion
> +-----------------------------------------------------------------------------------------------+
> | npu-smi xxx.xxx.xxx                Version: xxx.xxx.xxx                                       |
> +---------------------------+---------------+---------------------------------------------------+
> | NPU   Name         | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)      |
> | Chip  Phy-ID       | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)              |
> +===========================+===============+===================================================+
> | 0     xxx          | OK            | 157.3       32                0    / 0                   |
> | 0     0            | 0000:9D:00.0  | 0           0        / 0      3130 / 65536               |
> +---------------------------+---------------+---------------------------------------------------+
> | 0     xxx          | OK            | -           32                0    / 0                   |
> | 1     0            | 0000:9D:00.0  | 0           0        / 0      3130 / 65536               |
> +===========================+===============+===================================================+
> +---------------------------+---------------+---------------------------------------------------+
> | NPU     Chip       | Process id    | Process name| Process memory(MB) |Process id in container|
> +===========================+===============+===================================================+
> | No running processes found in NPU 0                                                           |
> +===========================+===============+===================================================+
> ```
