# Component Status Confirmation<a name="ZH-CN_TOPIC_0000002479386390"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:20:55.643Z pushedAt=2026-06-30T12:23:24.365Z -->

## Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002511426307"></a>

If Ascend Docker Runtime has been installed, perform the following steps on all nodes where this component is installed to confirm the status of Ascend Docker Runtime.

**Procedure<a name="section44081649104318"></a>**

1. Run the following command to check whether the base image exists.

    ```shell
    docker images | grep ubuntu
    ```

    The following is an example of the command output, indicating that the base image `ubuntu:22.04` exists. If the base image does not exist, run the `docker pull ubuntu:22.04` command to pull the base image.

    ```ColdFusion
    ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
    ```

2. Run the following command to mount the chip with physical chip ID 0 using Ascend Docker Runtime.

    - Docker (or K8s integration with Docker scenario).

        ```shell
        docker run -it -e ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 /bin/bash
        ```

    - Containerd (or K8s integration with Containerd scenario).

        Run the following command to mount the physical chip.

        ```shell
        ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 containerID
        ```

    >[!NOTE]
    >- The `ASCEND_VISIBLE_DEVICES` parameter indicates the chip ID to be mounted.
    >- `containerID` is a user-defined container ID.

3. Run the following command to check whether the chip is mounted successfully.

    ```shell
    ls /dev
    ```

    If the `davinci0` field appears in the output, the chip is mounted successfully, indicating that Ascend Docker Runtime is installed successfully and the component functions properly.

## NPU Exporter<a name="ZH-CN_TOPIC_0000002511346363"></a>

This section uses integration with Prometheus and reporting Prometheus data as an example to verify whether NPU Exporter is running properly.

**NPU Exporter Deployed Using an Image<a name="section1595201114126"></a>**

Run the following steps on any node to verify the installation status of NPU Exporter.

1. Use the following command to view the NPU Exporter Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If NPU Exporter is installed on multiple nodes in the cluster, verify each one individually.

    ```shell
    kubectl get pods -n npu-exporter -o wide | grep npu-exporter
    ```

    Example output:

    ```ColdFusion
    npu-exporter-4ln8w   1/1     Running   0          36m   192.168.102.109   ubuntu       <none>           <none>
    ```

2. Use the following command to view the logs of NPU Exporter in the K8s cluster.

    ```shell
    kubectl logs -n npu-exporter {NPU Exporter Pod name}
    ```

    Example output:

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

**NPU Exporter Deployed Using a Binary File<a name="zh-cn_topic_0000001497205429_section2976165515363"></a>**

Perform the following steps on the node where NPU Exporter is installed to verify the component installation status.

1. Log in to the node where NPU Exporter is deployed and run the following command to check the component service status. The component status must be `active (running)`.

    ```shell
    systemctl status npu-exporter
    ```

    Example output:

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

2. Check the component logs.

    ```shell
    cat /var/log/mindx-dl/npu-exporter/npu-exporter.log
    ```

Example of the displayed information:

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

Run the following steps on any node to verify the installation status of Ascend Device Plugin.

**Procedure<a name="zh-cn_topic_0000001497205413_section197491249115016"></a>**

1. Run the following command to view the pods of Ascend Device Plugin in the K8s cluster. Ensure that the `STATUS` of the Pod is `Running` and `READY` is `1/1`. If Ascend Device Plugin is installed on multiple nodes in the cluster, you need to verify this on each node.

    ```shell
    kubectl get pods -n kube-system -o wide | grep device-plugin
    ```

    Example output:

    ```ColdFusion
    ascend-device-plugin-daemonset-910-85p9v   1/1     Running   0          19h     192.168.185.251   ubuntu       <none>           <none>
    ```

2. Run the following command to view the logs of Ascend Device Plugin in the K8s cluster.

    ```shell
    kubectl logs -n kube-system {Ascend Device Plugin Pod name}
    ```

    The following example output indicates that the component is normal.

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
    [INFO]     2022/11/21 11:20:06.007755 35      server/plugin.go:87    ListAndWatch resp devices: Ascend910-4 Healthy# Chips reported to K8s. Use the actual information as the standard.
    [INFO]     2022/11/21 11:20:11.063218 21      kubeclient/client_server.go:123    reset annotation success
    ...
    ```

3. Run the following command to view the details of the node in K8s. If the "Capacity" and "Allocatable" fields in the node details contain information about the Ascend AI processor, it indicates that Ascend Device Plugin has reported the chips to K8s normally and the component is running properly.

    ```shell
    kubectl describe node {Node name in the K8s}
    ```

    >[!NOTE]
    >
    >Run the following command on the K8s management node to query the node name in K8s.
    >
    >```shell
    >kubectl get node
    >```
    >
    >The following is an example of the command output:
    >
    >```ColdFusion
    >NAME       STATUS   ROLES           AGE   VERSION
    >ubuntu     Ready    worker          23h   v1.17.3
    >```

    - Taking the Atlas 800 training server as an example, the command output is as follows:

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
          huawei.com/Ascend910:     8# K8s has detected that the node has a total of 8 NPUs
        ...
        Allocatable:
          cpu:                      72
          ephemeral-storage:        441969440446
          huawei.com/Ascend910:     8  # K8s has detected that the total number of allocatable NPUs on the node is 8
        ...
        ```

    - Taking a server (with an Atlas 300I inference card installed) as an example, the command output is as follows. The number of chips on the node is subject to the actual situation.

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

    - Take a server (with an Atlas 300I Pro inference card) as an example. In non-mixed insertion mode, the node contains Atlas inference series products. The following is an example of the command output. The number of chips on the node is subject to the actual situation.

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

    - Take a server (with an Atlas 300I Pro inference card) as an example. In mixed insertion mode, the node contains Atlas inference series products. The following is an example of the command output. The number of chips on the node is subject to the actual situation.

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

1. Run the following command to view the two Volcano pods in the K8s cluster. The `STATUS` of the Pods must be `Running`, and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n volcano-system -o wide | grep volcano
    ```

    Example output:

    ```ColdFusion
    volcano-controllers-758b6d8bdd-b7g89   1/1     Running   2          166m   192.168.102.69   ubuntu       <none>           <none>
    volcano-scheduler-86775f88f-w649w      1/1     Running   2          166m   192.168.102.91   ubuntu       <none>           <none>
    ```

2. Log in to the node where the Volcano Pod is running, and use the following command to view the Volcano component logs.
    - View the logs of volcano-controllers.

        ```shell
        cat /var/log/mindx-dl/volcano-controller/volcano-controller.log
        ```

        The following example output indicates that the component is running normally.

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

    - View the logs of volcano-scheduler.

        ```shell
        cat /var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log
        ```

        The following example output indicates that the component is running normally.

        ```ColdFusion
        Log file created at: 2022/10/14 11:22:32
        Running on machine: volcano-scheduler-86775f88f-6dtqf
        Binary: Built with gc go1.17.8-htrunk4 for linux/arm64
        Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
        ...
        ```

## ClusterD<a name="ZH-CN_TOPIC_0000002479386380"></a>

Perform the following steps on any node to verify the installation status of ClusterD.

1. Use the following command to view the ClusterD Pod in the K8s cluster. The number of Pods must be 1, the `STATUS` must be `Running`, and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep clusterd
    ```

    Example output:

    ```ColdFusion
    clusterd-7844cb867d-fwcj7   1/1     Running   0          2m14s   <none>   node133   <none>           <none>
    ```

2. Run the following command to query the Pod logs of ClusterD.

    ```shell
    kubectl logs -f -n mindx-dl {ClusterD Pod name}
    ```

The following is an example of the command output, indicating that the component is running properly.

    ```ColdFusion
    [INFO]     2024/07/24 13:58:30.602051 CST 1       hwlog@v0.10.12/api.go:105    cluster-info.log's logger init success
    W0724 13:58:30.602197       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.603416 CST 1       grpc/grpc_init.go:57    cluster info server start listen
    ...
    W0724 13:58:30.621433       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.621911 CST 258     job/factory.go:172    delete job summary cm goroutine started
    ```

## Ascend Operator<a name="ZH-CN_TOPIC_0000002479386462"></a>

Perform the following steps on any node to verify the installation status of Ascend Operator.

1. Run the following command to view the Ascend Operator Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep ascend-operator
    ```

    Example output:

    ```ColdFusion
    ascend-operator-manager-b59774f7-8l5gn         1/1     Running   0          6m52s   192.168.102.67   ubuntu       <none>           <none>
    ```

2. View the logs of Ascend Operator in the K8s cluster using the following command.

    ```shell
    kubectl logs -n mindx-dl {Ascend Operator Pod name}
    ```

    The following is an example of the command output, indicating that the component is running properly.

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl ascend-operator-manager-b59774f7-8l5gn
    [INFO]     2023/03/20 17:48:34.308373 1       hwlog/api.go:108    ascend-operator.log's logger init success
    [INFO]     2023/03/20 17:48:34.308469 1       ascend-operator/main.go:86    ascend-operator starting and the version is xxx
    [INFO]     2023/03/20 17:48:34.964296 1       ascend-operator/main.go:101    starting manager
    ...
    ```

## Infer Operator<a name="ZH-CN_TOPIC_0000002479386462"></a>

Perform the following steps on any node to verify the installation status of Infer Operator.

1. View the Pod of Infer Operator in the K8s cluster using the following command. Ensure that the Pod `STATUS` is `Running` and `READY` is `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep infer-operator
    ```

    Example output:

    ```ColdFusion
    infer-operator-manager-6bf95f6956-sdkbd         1/1     Running   0          6m52s   192.168.2.166   ubuntu       <none>           <none>
    ```

2. Run the following command to view the logs of Infer Operator in the K8s cluster.

    ```shell
    kubectl logs -n mindx-dl {Infer Operator Pod name}
    ```

    The following example output indicates that the component is running normally.

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl infer-operator-manager-6bf95f6956-sdkbd
    [INFO]     2026/03/20 16:22:12.668888 1       hwlog/api.go:164    infer-operator.log's logger init success
    ...
    ```

## NodeD<a name="ZH-CN_TOPIC_0000002479386440"></a>

Run the following steps on any node to verify the installation status of NodeD.

1. Use the following command to view the NodeD Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If NodeD is installed on multiple nodes in the cluster, verify each node.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep noded
    ```

    Example output:

    ```ColdFusion
    noded-bnmwt                        1/1     Running   10         40d    192.168.41.28     ubuntu       <none>           <none>
    ```

2. Use the following command to view the NodeD component logs.

    ```shell
    kubectl logs -n mindx-dl {NodeD Pod name}
    ```

    The following example output indicates that the component is running normally.

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

Run the following steps on any node to verify the installation status of Resilience Controller.

1. Use the following command to check the Resilience Controller Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep resilience-controller
    ```

    Example output:

    ```ColdFusion
    resilience-controller-76f4476bb5-fs986         1/1     Running   0          6m52s   192.168.102.67   ubuntu       <none>           <none>
    ```

2. Use the following command to check the Resilience Controller logs in the K8s cluster.

    ```shell
    kubectl logs -n mindx-dl {Resilience Controller Pod name}
    ```

    The following example output indicates that the component is running properly.

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

Perform the following steps on the node where Container Manager is deployed to verify its installation status.

1. Check the component service status. The component status must be `active (running)`.

    ```shell
    systemctl status container-manager.service
    ```

    Example output:

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
    >If information similar to the following appears in the output, you can ignore it. It does not affect actual functionality and may be caused by the RoCE NIC IP address and subnet mask not being configured.
    >
    >```ColdFusion
    >[dsmi_common_interface.c:1017][ascend][curpid:244135,244135][drv][dmp][dsmi_get_device_ip_address]devid 0 dsmi_cmd_get_device_ip_address return 1 error!
    >```

2. View the component logs.

    ```shell
    cat /var/log/mindx-dl/container-manager/container-manager.log
    ```

    The following uses the Atlas 800I A3 SuperPoD server as an example:

    ```ColdFusion
    [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
    [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
    [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
    [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
    [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
    [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
    ...
    ```

    If the following information is printed, the component is running properly.

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
