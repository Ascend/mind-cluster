# 组件状态确认<a name="ZH-CN_TOPIC_0000002479386390"></a>

## Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002511426307"></a>

如已安装Ascend Docker Runtime，请在所有安装了该组件的节点上执行如下步骤确认Ascend Docker Runtime的状态。

**操作步骤<a name="section44081649104318"></a>**

1. 执行以下命令，查看是否存在基础镜像。

    ```shell
    docker images | grep ubuntu
    ```

    回显示例如下，表示存在基础镜像ubuntu:22.04。若不存在基础镜像，可以执行**docker pull ubuntu:22.04**命令，拉取基础镜像。

    ```ColdFusion
    ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
    ```

2. 执行以下命令，使用Ascend Docker Runtime挂载物理芯片ID为0的芯片。

    - Docker（或K8s集成Docker场景）。

        ```shell
        docker run -it -e ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 /bin/bash
        ```

    - Containerd（或K8s集成Containerd场景）。

        执行以下命令，挂载物理芯片。

        ```shell
        ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 containerID
        ```

    >[!NOTE]
    >- ASCEND\_VISIBLE\_DEVICES参数表示挂载的芯片ID。
    >- containerID为用户自定义的容器ID。

3. 执行以下命令，查询芯片是否挂载成功。

    ```shell
    ls /dev
    ```

    若回显中存在**davinci0**字段，表示芯片挂载成功，安装Ascend Docker Runtime成功且组件功能正常。

## NPU Exporter<a name="ZH-CN_TOPIC_0000002511346363"></a>

本章节以对接Prometheus，上报Prometheus数据为例，确认NPU Exporter组件是否正常运行。

**NPU Exporter使用容器部署<a name="section1595201114126"></a>**

请在任意节点执行以下步骤验证NPU Exporter的安装状态。

1. 通过如下命令查看K8s集群中NPU Exporter的Pod，需要满足Pod的STATUS为Running，READY为1/1。如果集群中有多个节点安装了NPU Exporter，需要逐个确认。

    ```shell
    kubectl get pods -n npu-exporter -o wide | grep npu-exporter
    ```

    回显示例：

    ```ColdFusion
    npu-exporter-4ln8w   1/1     Running   0          36m   192.168.102.109   ubuntu       <none>           <none>
    ```

2. 通过如下命令查看K8s集群中NPU Exporter的日志。

    ```shell
    kubectl logs -n npu-exporter {npu-exporter组件的Pod名字}
    ```

    回显示例：

    ```ColdFusion
    [INFO]     2023/12/08 07:38:56.551173 1       hwlog/api.go:108    npu-exporter.log's logger init success
    [INFO]     2023/12/08 07:38:56.551275 1       npu-exporter/main.go:205    listen on: 0.0.0.0
    [INFO]     2023/12/08 07:38:56.551369 1       npu-exporter/main.go:325    npu exporter starting and the version is v26.0.0_linux-x86_64
    [WARN]     2023/12/08 07:38:56.684424 1       npu-exporter/main.go:339    enable unsafe http server
    [WARN]     2023/12/08 07:39:01.686205 98      container/runtime_ops.go:150    failed to get OCI connection: context deadline exceeded
    [WARN]     2023/12/08 07:39:01.686311 98      container/runtime_ops.go:152    use backup address to try again
    [INFO]     2023/12/08 07:39:01.687444 98      collector/npu_collector.go:418    Starting update cache every 5 seconds
    [WARN]     2023/12/08 07:39:01.688039 157     collector/npu_collector.go:463    get info of npu-exporter-network-info failed: no value found, so use initial net info
    [INFO]     2023/12/08 07:39:01.744739 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:01.852413 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:05.055247 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:06.688352 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:06.750876 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:09.843914 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:11.688505 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:11.701081 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:14.859243 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    ...
    ```

**NPU Exporter使用二进制部署<a name="zh-cn_topic_0000001497205429_section2976165515363"></a>**

请在安装NPU Exporter的节点执行以下步骤验证组件的安装状态。

1. 登录部署NPU Exporter的节点，使用如下命令，查看组件服务的状态，需要满足组件状态为active \(running\)。

    ```shell
    systemctl status npu-exporter
    ```

    回显示例：

    ```ColdFusion
    root@ubuntu:~# systemctl status npu-exporter
    ● npu-exporter.service - Ascend npu exporter
       Loaded: loaded (/etc/systemd/system/npu-exporter.service; enabled; vendor preset: enabled)
       Active: active (running) since Thu 2022-11-17 16:24:41 CST; 3 days ago
     Main PID: 25121 (npu-exporter)
        Tasks: 8 (limit: 7372)
       CGroup: /system.slice/npu-exporter.service
               └─25121 /usr/local/bin/npu-exporter -ip=127.0.0.1 -port=8082 -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log
    ...
    ```

2. 查看组件日志。

    ```shell
    cat /var/log/mindx-dl/npu-exporter/npu-exporter.log
    ```

    回显示例：

    ```ColdFusion
    [INFO]     2023/12/08 07:38:56.551173 1       hwlog/api.go:108    npu-exporter.log's logger init success
    [INFO]     2023/12/08 07:38:56.551275 1       npu-exporter/main.go:205    listen on: 0.0.0.0
    [INFO]     2023/12/08 07:38:56.551369 1       npu-exporter/main.go:325    npu exporter starting and the version is v26.0.0_linux-x86_64
    [WARN]     2023/12/08 07:38:56.684424 1       npu-exporter/main.go:339    enable unsafe http server
    [WARN]     2023/12/08 07:39:01.686205 98      container/runtime_ops.go:150    failed to get OCI connection: context deadline exceeded
    [WARN]     2023/12/08 07:39:01.686311 98      container/runtime_ops.go:152    use backup address to try again
    [INFO]     2023/12/08 07:39:01.687444 98      collector/npu_collector.go:418    Starting update cache every 5 seconds
    [WARN]     2023/12/08 07:39:01.688039 157     collector/npu_collector.go:463    get info of npu-exporter-network-info failed: no value found, so use initial net info
    [INFO]     2023/12/08 07:39:01.744739 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:01.852413 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:05.055247 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:06.688352 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:06.750876 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:09.843914 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:11.688505 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:11.701081 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:14.859243 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    ...
    ```

## Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426319"></a>

请在任意节点执行以下步骤验证Ascend Device Plugin的安装状态。

**操作步骤<a name="zh-cn_topic_0000001497205413_section197491249115016"></a>**

1. 通过如下命令查看K8s集群中Ascend Device Plugin的Pod，需要满足Pod的“STATUS”为Running，READY为1/1。如果集群中有多个节点安装了Ascend Device Plugin，每一个节点都需要确认。

    ```shell
    kubectl get pods -n kube-system -o wide | grep device-plugin
    ```

    回显示例：

    ```ColdFusion
    ascend-device-plugin-daemonset-910-85p9v   1/1     Running   0          19h     192.168.185.251   ubuntu       <none>           <none>
    ```

2. 通过如下命令查看K8s集群中Ascend Device Plugin的日志。

    ```shell
    kubectl logs -n kube-system Ascend Device Plugin组件的Pod名字
    ```

    回显示例如下，表示组件正常。

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n kube-system ascend-device-plugin-daemonset-910-85p9v 
    [INFO]     2022/11/21 11:20:04.534992 1       hwlog@v0.0.0/api.go:96    devicePlugin.log's logger init success
    [INFO]     2022/11/21 11:20:04.535750 1       main.go:127    ascend device plugin starting and the version is xxx_linux-x86_64
    [INFO]     2022/11/21 11:20:05.992823 1       K8stool@v0.0.0/self_K8s_client.go:116    start to decrypt cfg
    [INFO]     2022/11/21 11:20:06.002773 1       K8stool@v0.0.0/self_K8s_client.go:125    Config loaded from file: ****tc/mindx-dl/device-plugin/.config/config6
    [INFO]     2022/11/21 11:20:06.003751 1       main.go:153    init kube client success 
    [INFO]     2022/11/21 11:20:06.003923 1       device/ascendcommon.go:104    Found Huawei Ascend, deviceType: Ascend910, deviceName: Ascend910-4
    [INFO]     2022/11/21 11:20:06.003970 1       main.go:160    init device manager success
    [INFO]     2022/11/21 11:20:06.004157 21      device/manager.go:125    starting the listen device
    [INFO]     2022/11/21 11:20:06.004285 7       device/manager.go:206    Serve start
    [INFO]     2022/11/21 11:20:06.004970 7       server/server.go:88    device plugin (Ascend910) start serving.
    [INFO]     2022/11/21 11:20:06.007285 7       server/server.go:36    register Ascend910 to kubelet success.
    [INFO]     2022/11/21 11:20:06.007521 7       server/pod_resource.go:44    pod resource client init success.
    [INFO]     2022/11/21 11:20:06.007755 35      server/plugin.go:87    ListAndWatch resp devices: Ascend910-4 Healthy# 上报K8s的芯片，请以实际为准
    [INFO]     2022/11/21 11:20:11.063218 21      kubeclient/client_server.go:123    reset annotation success
    ...
    ```

3. 通过如下命令查看K8s中节点的详细情况。如果节点详情中的“Capacity”字段和“Allocatable”字段出现了昇腾AI处理器的相关信息，表示Ascend Device Plugin给K8s上报芯片正常，组件运行正常。

    ```shell
    kubectl describe node K8s中的节点名
    ```

    - 以Atlas 800 训练服务器为例，回显示例如下：

        ```ColdFusion
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend910
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                      72
          ephemeral-storage:        479567536Ki
          huawei.com/Ascend910:     8# K8s已感知到该节点总共有8个NPU
        ...
        Allocatable:
          cpu:                      72
          ephemeral-storage:        441969440446
          huawei.com/Ascend910:     8  # K8s已感知到该节点可供分配的NPU总个数为8
        ...
        ```

    - 以服务器（插Atlas 300I 推理卡）为例，回显示例如下，节点上芯片个数请以实际为准。

        ```ColdFusion
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend310
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                       72
          ephemeral-storage:         163760Mi
          huawei.com/Ascend310:      4
        ...
        Allocatable:
          cpu:                       72
          ephemeral-storage:         154543324929
          huawei.com/Ascend310:      4
        ...
        ```

    - 以服务器（插Atlas 300I Pro 推理卡）为例。非混插模式，节点包含Atlas 推理系列产品，回显示例如下，节点上芯片个数请以实际为准。

        ```ColdFusion
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend310
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                      96
          ephemeral-storage:        95596964Ki
          huawei.com/Ascend310P:    3
        ...
        Allocatable:
          cpu:                      96
          ephemeral-storage:        88102161877
          huawei.com/Ascend310P:    3
        ...
        ```

    - 以服务器（插Atlas 300I Pro 推理卡）为例。混插模式，节点包含Atlas 推理系列产品，回显示例如下，节点上芯片个数请以实际为准。

        ```ColdFusion
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend310
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                      96
          ephemeral-storage:        95596964Ki
          huawei.com/Ascend310P-IPro:    1
          huawei.com/Ascend310P-V:       1
          huawei.com/Ascend310P-VPro:    1
        ...
        Allocatable:
          cpu:                      96
          ephemeral-storage:        88102161877
          huawei.com/Ascend310P-IPro:    1
          huawei.com/Ascend310P-V:       1
          huawei.com/Ascend310P-VPro:    1
        ...
        ```

## Volcano<a name="ZH-CN_TOPIC_0000002511346325"></a>

1. 通过如下命令查看K8s集群中Volcano的两个Pod，需要满足Pod的STATUS都为Running，READY都为1/1。

    ```shell
    kubectl get pods -n volcano-system -o wide | grep volcano
    ```

    回显示例：

    ```ColdFusion
    volcano-controllers-758b6d8bdd-b7g89   1/1     Running   2          166m   192.168.102.69   ubuntu       <none>           <none>
    volcano-scheduler-86775f88f-w649w      1/1     Running   2          166m   192.168.102.91   ubuntu       <none>           <none>
    ```

2. 登录Volcano Pod运行的节点，使用如下命令查看Volcano组件日志。
    - 查看volcano-controllers的日志。

        ```shell
        cat /var/log/mindx-dl/volcano-controller/volcano-controller.log
        ```

        回显示例如下，表示组件正常运行。

        ```ColdFusion
        Log file created at: 2022/10/14 11:22:32
        Running on machine: volcano-controllers-758b6d8bdd-wc49r
        Binary: Built with gc go1.17.8-htrunk4 for linux/arm64
        Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
        I1014 11:22:32.070656       1 garbagecollector.go:91] Starting garbage collector
        I1014 11:22:32.072772       1 queue_controller.go:171] Starting queue controller.
        I1014 11:22:32.652887       1 queue_controller.go:238] Begin execute SyncQueue action for queue default, current status
        I1014 11:22:32.653026       1 queue_controller_action.go:36] Begin to sync queue default.
        I1014 11:22:32.756216       1 queue_controller_action.go:82] End sync queue default.
        I1014 11:22:32.756254       1 queue_controller.go:220] Finished syncing queue default (103.399375ms).
        I1014 11:22:32.972001       1 pg_controller.go:109] PodgroupController is running ......
        I1014 11:22:32.972396       1 job_controller.go:252] JobController is running ......
        I1014 11:22:32.972423       1 job_controller.go:256] worker 1 start ......
        I1014 11:22:32.972426       1 job_controller.go:256] worker 0 start ......
        I1014 11:22:32.972426       1 job_controller.go:256] worker 2 start ......
        ...
        ```

    - 查看volcano-scheduler的日志。

        ```shell
        cat /var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log
        ```

        回显示例如下，表示组件运行正常。

        ```ColdFusion
        Log file created at: 2022/10/14 11:22:32
        Running on machine: volcano-scheduler-86775f88f-6dtqf
        Binary: Built with gc go1.17.8-htrunk4 for linux/arm64
        Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
        ...
        ```

## ClusterD<a name="ZH-CN_TOPIC_0000002479386380"></a>

请在任意节点执行以下步骤验证ClusterD的安装状态。

1. 通过如下命令查看K8s集群中ClusterD的Pod，需要满足Pod的数量为1，STATUS为Running，READY为1/1。

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep clusterd
    ```

    回显示例：

    ```ColdFusion
    clusterd-7844cb867d-fwcj7   1/1     Running   0          2m14s   <none>   node133   <none>           <none>
    ```

2. 执行以下命令，查询ClusterD的Pod日志。

    ```shell
    kubectl logs -f -n mindx-dl {ClusterD组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```ColdFusion
    [INFO]     2024/07/24 13:58:30.602051 CST 1       hwlog@v0.10.12/api.go:105    cluster-info.log's logger init success
    W0724 13:58:30.602197       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.603416 CST 1       grpc/grpc_init.go:57    cluster info server start listen
    ...
    W0724 13:58:30.621433       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.621911 CST 258     job/factory.go:172    delete job summary cm goroutine started
    ```

## Ascend Operator<a name="ZH-CN_TOPIC_0000002479386462"></a>

请在任意节点执行以下步骤验证Ascend Operator的安装状态。

1. 通过如下命令查看K8s集群中Ascend Operator的Pod，需要满足Pod的STATUS为Running，READY为1/1。

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep ascend-operator
    ```

    回显示例：

    ```ColdFusion
    ascend-operator-manager-b59774f7-8l5gn         1/1     Running   0          6m52s   192.168.102.67   ubuntu       <none>           <none>
    ```

2. 通过如下命令查看K8s集群中Ascend Operator的日志。

    ```shell
    kubectl logs -n mindx-dl {Ascend Operator组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl ascend-operator-manager-b59774f7-8l5gn 
    [INFO]     2023/03/20 17:48:34.308373 1       hwlog/api.go:108    ascend-operator.log's logger init success
    [INFO]     2023/03/20 17:48:34.308469 1       ascend-operator/main.go:86    ascend-operator starting and the version is xxx
    [INFO]     2023/03/20 17:48:34.964296 1       ascend-operator/main.go:101    starting manager
    ...
    ```

## Infer Operator<a name="ZH-CN_TOPIC_0000002479386462"></a>

请在任意节点执行以下步骤验证Infer Operator的安装状态。

1. 通过如下命令查看K8s集群中Infer Operator的Pod，需要满足Pod的STATUS为Running，READY为1/1。

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep infer-operator
    ```

    回显示例：

    ```ColdFusion
    infer-operator-manager-6bf95f6956-sdkbd         1/1     Running   0          6m52s   192.168.2.166   ubuntu       <none>           <none>
    ```

2. 通过如下命令查看K8s集群中Infer Operator的日志。

    ```shell
    kubectl logs -n mindx-dl {Infer Operator组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl infer-operator-manager-6bf95f6956-sdkbd 
    [INFO]     2026/03/20 16:22:12.668888 1       hwlog/api.go:164    infer-operator.log's logger init success
    ...
    ```

## NodeD<a name="ZH-CN_TOPIC_0000002479386440"></a>

请在任意节点执行以下步骤验证NodeD的安装状态。

1. 通过如下命令查看K8s集群中NodeD的Pod，需要满足Pod的STATUS为Running，READY为1/1。如果集群中有多个节点安装了NodeD，每个节点都需要确认。

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep noded
    ```

    回显示例：

    ```ColdFusion
    noded-bnmwt                        1/1     Running   10         40d    192.168.41.28     ubuntu       <none>           <none>
    ```

2. 通过如下命令查看NodeD组件日志。

    ```shell
    kubectl logs -n mindx-dl {NodeD组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```ColdFusion
    [INFO] 2025/05/25 15:24:19.897280 1 hwlog/api.go:108 noded.log's logger init success
    [INFO] 2025/05/25 15:24:19.897392 1 noded/main.go:93 noded starting and the version is v26.0.0_linux-x86_64
    W0525 15:24:19.897410 1 client_config.go:617] Neither --kubeconfig nor --master was specified. Using the inClusterConfig. This might not work.
    [INFO] 2025/05/25 15:24:19.994306 1 devmanager/devmanager.go:123 the dcmi version is 24.1.rc3.b060
    [INFO] 2025/05/25 15:24:19.994360 1 devmanager/devmanager.go:1071 get chip base info, cardID: 0, deviceID: 0, logicID: 0, physicID: 0
    [INFO] 2025/05/25 15:24:19.994386 1 devmanager/devmanager.go:1071 get chip base info, cardID: 1, deviceID: 0, logicID: 1, physicID: 1
    [INFO] 2025/05/25 15:24:19.994408 1 devmanager/devmanager.go:1071 get chip base info, cardID: 2, deviceID: 0, logicID: 2, physicID: 2
    [INFO] 2025/05/25 15:24:19.994430 1 devmanager/devmanager.go:1071 get chip base info, cardID: 3, deviceID: 0, logicID: 3, physicID: 3
    [INFO] 2025/05/25 15:24:19.994449 1 devmanager/devmanager.go:1071 get chip base info, cardID: 4, deviceID: 0, logicID: 4, physicID: 4
    [INFO] 2025/05/25 15:24:19.994476 1 devmanager/devmanager.go:1071 get chip base info, cardID: 5, deviceID: 0, logicID: 5, physicID: 5
    [INFO] 2025/05/25 15:24:19.994505 1 devmanager/devmanager.go:1071 get chip base info, cardID: 6, deviceID: 0, logicID: 6, physicID: 6
    [INFO] 2025/05/25 15:24:19.994528 1 devmanager/devmanager.go:1071 get chip base info, cardID: 7, deviceID: 0, logicID: 7, physicID: 7
    [WARN] 2025/05/25 15:24:19.994564 1 executor/dev_manager.go:71 deviceManager get hccsPingMeshState failed, err: dcmi get hccs ping mesh state failed cardID(0) deviceID(0) error code: -99998
    [ERROR] 2025/05/25 15:24:19.994588 1 pingmesh/controller.go:68 new device manager failed, err: dcmi get hccs ping mesh state failed cardID(0) deviceID(0) error code: -99998
    [INFO] 2025/05/25 15:24:19.999314 1 config/configurator.go:98 update fault config success
    [INFO] 2025/05/25 15:24:19.999350 1 config/configurator.go:231 init fault config from config map success
    [INFO] 2025/05/25 15:24:39.037815 1 control/controller.go:220 get node SN success, add SN(HS20200764) to node annotation
    ...
    ```

## Resilience Controller<a name="ZH-CN_TOPIC_0000002511426295"></a>

请在任意节点执行以下步骤验证Resilience Controller的安装状态。

1. 通过如下命令查看K8s集群中Resilience Controller的Pod，需要满足Pod的STATUS为Running，READY为1/1。

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep resilience-controller
    ```

    回显示例：

    ```ColdFusion
    resilience-controller-76f4476bb5-fs986         1/1     Running   0          6m52s   192.168.102.67   ubuntu       <none>           <none>
    ```

2. 通过如下命令查看K8s集群中Resilience Controller的日志。

    ```shell
    kubectl logs -n mindx-dl {Resilience组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl resilience-controller-76f4476bb5-fs986 
    [INFO]     2022/11/17 17:18:46.697010 1       hwlog@v0.0.0/api.go:96    run.log's logger init success
    [INFO]     2022/11/17 17:18:46.697139 1       cmd/main.go:57    resilience-controller starting and the version is xxx_linux-x86_64
    [INFO]     2022/11/17 17:18:47.227913 1       K8stool@v0.0.0/self_K8s_client.go:116    start to decrypt cfg
    [INFO]     2022/11/17 17:18:47.297559 1       K8stool@v0.0.0/self_K8s_client.go:125    Config loaded from file: ****tc/mindx-dl/resilience-controller/.config/config6
    [INFO]     2022/11/17 17:18:47.300066 1       elastic/controller.go:45    Setting up elastic event handlers
    [INFO]     2022/11/17 17:18:47.300179 1       elastic/controller.go:63    Starting elastic controller, waiting for informer caches to sync
    [INFO]     2022/11/17 17:18:47.401246 1       cmd/main.go:80    elastic controller started
    ...
    ```

## Container Manager<a name="ZH-CN_TOPIC_0000002492269056"></a>

请在Container Manager组件部署的节点上执行以下步骤验证Container Manager组件的安装状态。

1. 查看组件服务的状态，需要满足组件状态为active \(running\)。

    ```shell
    systemctl status container-manager.service
    ```

    回显示例：

    ```ColdFusion
    ● container-manager.service - Ascend container manager
         Loaded: loaded (/etc/systemd/system/container-manager.service; disabled; vendor preset: enabled)
         Active: active (running) since Wed 2025-11-26 20:56:50 UTC; 16s ago
        Process: 41459 ExecStart=/bin/bash -c container-manager run  -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
       Main PID: 41464 (container-manag)
          Tasks: 10 (limit: 629145)
         Memory: 13.3M
         CGroup: /system.slice/container-manager.service
                 └─41464 /home/container-manager/container-manager run -ctrStrategy ringRecover
    ...
    ```

    >[!NOTE]
    >若回显中出现类似如下信息，可忽略，不影响实际功能，可能原因是未配置RoCE网卡IP地址和子网掩码。若不想打印该信息，可参见《Atlas A2 中心推理和训练硬件 25.5.0 HCCN Tool 接口参考》的“[配置功能\>配置RoCE网卡IP地址和子网掩码](https://support.huawei.com/enterprise/zh/doc/EDOC1100540101/44299f2a)”章节配置。
    >
    >```ColdFusion
    >[dsmi_common_interface.c:1017][ascend][curpid:244135,244135][drv][dmp][dsmi_get_device_ip_address]devid 0 dsmi_cmd_get_device_ip_address return 1 error!
    >```

2. 查看组件日志。

    ```shell
    cat /var/log/mindx-dl/container-manager/container-manager.log
    ```

    回显以Atlas 800I A3 超节点服务器为例：

    ```ColdFusion
    [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
    [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
    [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
    [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
    [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
    [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
    ...
    ```

    如果出现如下打印信息，表示组件运行正常。

    ```ColdFusion
    ...
    [INFO]     2025/11/25 22:46:59.289352 1       devmgr/workflow.go:57    init module <hwDev manager> success
    [INFO]     2025/11/25 22:46:59.293773 1       app/config.go:40    load fault config from faultCode.json success
    [INFO]     2025/11/25 22:46:59.293866 1       app/workflow.go:50    init module <fault manager> success
    [INFO]     2025/11/25 22:46:59.293901 1       app/workflow.go:76    init module <container controller> success
    [INFO]     2025/11/25 22:46:59.293930 1       app/workflow.go:64    init module <reset-manager> success
    [INFO]     2025/11/25 22:46:59.315101 378     devmgr/hwdevmgr.go:365    subscribe device fault event success
    ...
    ```
