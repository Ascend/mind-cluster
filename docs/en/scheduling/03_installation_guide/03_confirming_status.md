# Component Status Confirmation<a name="ZH-CN_TOPIC_0000002479386390"></a>

## Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002511426307"></a>

If Ascend Docker Runtime has been installed, perform the following procedure on all nodes where this component is installed to confirm the status of Ascend Docker Runtime.

**Procedure<a name="section44081649104318"></a>**

1. Run the following command to check whether the configuration file has been modified successfully.

   - Docker (or K8s integrated with Docker)

       ```shell
       cat /etc/docker/daemon.json | grep ascend
       ```

     The output example is as follows, indicating that the modification is successful.

       ```ColdFusion
       "default-runtime": "ascend",
           "ascend": {
               "path": "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime",
       ```

   - Containerd (or K8s integrated with Containerd)

       ```shell
       cat /etc/containerd/config.toml | grep ascend
       ```

     The output example is as follows, indicating that the modification is successful.

       ```ColdFusion
       default_runtime_name = "ascend"
           [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.ascend]
               [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.ascend.options]
                   BinaryName = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"
       ```

   - CRI-O (or K8s integrated with CRI-O).

       ```shell
       cat /etc/crio/crio.conf.d/99-ascend-runtime.conf | grep ascend
       ```

     The output example is as follows, indicating that the modification is successful.

        ```ColdFusion
        default_runtime = "ascend"
            [crio.runtime.runtimes.ascend]
                runtime_path = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"
        ```

   >[!NOTE]
   >If the configuration file path of Docker, Containerd, or CRI-O differs from the actual path on your system, modify it accordingly.

2. Run the following command to check whether the base image exists.

   - Docker (or K8s integrated with Docker)

       ```shell
       docker images | grep ubuntu
       ```

     The output example is as follows, indicating that the base image ubuntu:22.04 exists. If the base image does not exist, run the **docker pull ubuntu:22.04** command to pull the base image.

       ```ColdFusion
       ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
       ```

   - Containerd (or K8s integrated Containerd)

       ```shell
       ctr i ls | grep ubuntu
       ```

     The output example is as follows, indicating that the base image ubuntu:22.04 exists. If the base image does not exist, run the **ctr images pull docker.io/library/ubuntu:22.04** command to pull the base image.

       ```ColdFusion
       docker.io/library/ubuntu:22.04    application/vnd.docker.distribution.manifest.v2+json sha256:555f8bd7441bb97303961a52ec7dec94d755f9f39077801de3e11c706c9ee7dc 68.5 MiB  linux/arm64 io.cri-containerd.image=managed
       ```

   - CRI-O (or K8s integrated with CRI-O)

       ```shell
       crictl images | grep ubuntu
       ```

     The output example is as follows, indicating that the base image ubuntu:22.04 exists. If the base image does not exist, run the **crictl pull docker.io/library/ubuntu:22.04** command to pull the base image.

       ```ColdFusion
       docker.io/library/ubuntu:22.04     22.04              ec1c05cc9451e       246MB


3. Run the following command to use Ascend Docker Runtime to mount the chip whose physical chip ID is 0.

   - Docker (or K8s integrated with Docker)

       ```shell
       docker run -it -e ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 /bin/bash
       ```

   - Containerd (or K8s integrated with Containerd)

     Run the following command to mount the physical chip.

       ```shell
       ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 docker.io/library/ubuntu:22.04 containerID
       ```

   - CRI-O (or K8s integrated with CRI-O).

     For details about how to use CRI-O, see: [Usage on the CRI-O Client Using the crictl CLI](../04_usage/00_containerization/04_usage_on_the_crio_client.md)

   >[!NOTE]
   >- `ASCEND_VISIBLE_DEVICES` indicates the ID of the mounted chip.
   >- `containerID` is a user-defined container ID.

4. Run the following command to query whether the chip is mounted successfully.

    ```shell
    ls /dev
    ```

   If the output contains the **davinci0** field, the chip is mounted successfully, indicating that Ascend Docker Runtime is installed successfully and the component functions properly.

## NPU Exporter<a name="ZH-CN_TOPIC_0000002511346363"></a>

This section uses the integration with Prometheus and reporting of Prometheus data as an example to confirm whether NPU Exporter is running properly.

**NPU Exporter Deployed Using an Image<a name="section1595201114126"></a>**

Perform the following procedure on any node to verify the installation status of NPU Exporter.

1. Run the following command to view the NPU Exporter Pods in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If NPU Exporter is installed on multiple nodes in the cluster, confirm each node one by one.

    ```shell
    kubectl get pods -n npu-exporter -o wide | grep npu-exporter
    ```

   Output example:

    ```ColdFusion
    npu-exporter-4ln8w   1/1     Running   0          36m   192.0.2.x   ubuntu       <none>           <none>
    ```

2. Run the following command to view the NPU Exporter logs in the K8s cluster.

    ```shell
    kubectl logs -n npu-exporter {npu-exporter_pod_name}
    ```

   Output example:

    ```ColdFusion
    [INFO]     2023/12/08 07:38:56.551173 1       hwlog/api.go:108    npu-exporter.log's logger init success
    [INFO]     2023/12/08 07:38:56.551275 1       npu-exporter/main.go:205    listen on: 0.0.0.0
    [INFO]     2023/12/08 07:38:56.551369 1       npu-exporter/main.go:325    npu exporter starting and the version is v26.2.0_linux-x86_64
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

**NPU Exporter Deployed Using Binaries<a name="zh-cn_topic_0000001497205429_section2976165515363"></a>**

Perform the following procedure on the node where NPU Exporter is installed to verify the installation status.

1. Log in to the node where NPU Exporter is deployed and run the following command to check the component service status. The component status must be `active (running)`.

    ```shell
    systemctl status npu-exporter
    ```

   Output example:

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

2. Check the component log.

    ```shell
    cat /var/log/mindx-dl/npu-exporter/npu-exporter.log
    ```

   Output example:

    ```ColdFusion
    [INFO]     2023/12/08 07:38:56.551173 1       hwlog/api.go:108    npu-exporter.log's logger init success
    [INFO]     2023/12/08 07:38:56.551275 1       npu-exporter/main.go:205    listen on: 0.0.0.0
    [INFO]     2023/12/08 07:38:56.551369 1       npu-exporter/main.go:325    npu exporter starting and the version is v26.2.0_linux-x86_64
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

Perform the following procedure on any node to verify the installation status of Ascend Device Plugin.

**Procedure<a name="zh-cn_topic_0000001497205413_section197491249115016"></a>**

1. Run the following command to view the Pods of Ascend Device Plugin in the K8s cluster. The `STATUS` of the Pod must be `Running` and `READY` must be `1/1`. If Ascend Device Plugin is installed on multiple nodes in the cluster, confirm the status on each node.

    ```shell
    kubectl get pods -n kube-system -o wide | grep device-plugin
    ```

   Output example:

    ```ColdFusion
    ascend-device-plugin-daemonset-910-85p9v   1/1     Running   0          19h     192.168.x.x   ubuntu       <none>           <none>
    ```

2. Run the following command to view the logs of Ascend Device Plugin in the K8s cluster.

    ```shell
    kubectl logs -n kube-system {Ascend Device Plugin_pod_name}
    ```

   The following output indicates that the component is running normally:

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
    [INFO]     2022/11/21 11:20:06.007755 35      server/plugin.go:87    ListAndWatch resp devices: Ascend910-4 Healthy# Chips reported to K8s. The actual value prevails.
    [INFO]     2022/11/21 11:20:11.063218 21      kubeclient/client_server.go:123    reset annotation success
    ...
    ```

3. Run the following command to view the details of the node in K8s. If the `Capacity` field and the `Allocatable` field in the node details contain information about the Ascend AI Processor, it indicates that Ascend Device Plugin reports chips to K8s normally and the component is running normally.

    ```shell
    kubectl describe node {node_name_in_K8s}
    ```

   >[!NOTE]
   >
   >Run the following command on the K8s management node to query the node name in K8s.
   >
   >```shell
    >kubectl get node
    >```
   >
   >Output example:
   >
   >```ColdFusion
    >NAME       STATUS   ROLES           AGE   VERSION
    >ubuntu     Ready    worker          23h   v1.17.3
    >```

   - Taking the Atlas 800 training server as an example, the following output is displayed:

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
         huawei.com/Ascend910:     8  # K8s has detected that the node has a total of 8 NPUs.
       ...
       Allocatable:
         cpu:                      72
         ephemeral-storage:        441969440446
         huawei.com/Ascend910:     8  # K8s has detected that the total number of allocatable NPUs on the node is 8.
       ...
       ```

   - Taking a server (with an Atlas 300I inference card installed) as an example, the following output is displayed. The number of chips on the node is subject to the actual situation.

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

   - Taking a server (with an Atlas 300I Pro inference card installed) as an example. In non-mixed insertion mode, the node contains <term>Atlas inference products</term>. The following output is displayed. The number of chips on the node is subject to the actual situation.

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

   - Take a server (with an Atlas 300I Pro inference card installed) as an example. In mixed insertion mode, the node contains <term>Atlas inference products</term>. The following output is an example, and the number of chips on the node is subject to the actual situation.

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

## K8s RDMA Shared Dev Plugin<a name="ZH-CN_TOPIC_0000002524312662"></a>

Perform the following procedure on any node to verify the installation status of K8s RDMA Shared Dev Plugin.

**Procedure<a name="section-k8s-rdma-check"></a>**

1. Run the following command to view the Pods of K8s RDMA Shared Dev Plugin in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If K8s RDMA Shared Dev Plugin is installed on multiple compute nodes in the cluster, confirm the status on each node.

    ```shell
    kubectl get pods -n kube-system -o wide | grep rdma-shared-dp
    ```

   Output example:

    ```ColdFusion
    rdma-shared-dp-ds-fd6t8                           1/1     Running   0          74s   192.0.2.x   ubuntu       <none>           <none>
    ```

   > [!NOTE]
   > If the Pod status is not `Running`, refer to [Component Pod Status Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342) for troubleshooting. If the Pod status is `ContainerCreating`, refer to [Cluster Scheduling Component Pod in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343) for troubleshooting.

2. Run the following command to view the logs of K8s RDMA Shared Dev Plugin in the K8s cluster.

    ```shell
    kubectl logs -n kube-system {K8s RDMA Shared Dev Plugin_pod_name}
    ```

   The following output indicates that the component is running normally:

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n kube-system rdma-shared-dp-ds-fd6t8
    [INFO]     2026/07/15 10:30:25.123456 1       hwlog/api.go:108    k8s-rdma-shared-dp.log's logger init success
    [INFO]     2026/07/15 10:30:25.456789 7       ub_device/server.go:157    huawei.com/ub_rdma device plugin endpoint started serving
    [INFO]     2026/07/15 10:30:25.567890 7       ub_device/server.go:280    ListAndWatch called by kubelet for: huawei.com/ub_rdma
    [INFO]     2026/07/15 10:30:25.567890 7       ub_device/server.go:284    Updating "huawei.com/ub_rdma" devices
    [INFO]     2026/07/15 10:30:25.567890 7       ub_device/server.go:289    exposing "8" devices
    ```

3. Run the following command to view the RDMA resource information of the node in K8s. If the `Capacity` field and the `Allocatable` field in the node details contain information about RDMA shared devices, it indicates that K8s RDMA Shared Dev Plugin reports resources to K8s normally and the component is running normally.

    ```shell
    kubectl describe node {node_name_in_K8s}
    ```

  > [!NOTE]
  > Run the following command on the K8s management node to query the node name in K8s.
  >
  > ```shell
   > kubectl get node
   > ```
  >
  > Output example:
  >
  > ```ColdFusion
   > NAME             STATUS   ROLES           AGE   VERSION
   > compute-node-1   Ready    worker          23h   v1.28.0
   > ```

   The following output shows the number of RDMA devices on the node. The actual number prevails.

    ```ColdFusion
    root@ubuntu:~# kubectl describe node compute-node-1
    Name:               compute-node-1
    Roles:              worker
    Labels:             kubernetes.io/arch=amd64
                        ...
    CreationTimestamp:  Wed, 15 Jul 2026 10:30:00 +0800
    Taints:             <none>
    Unschedulable:      false
    ...
    Capacity:
      cpu:                      64
      ephemeral-storage:        479567536Ki
      memory:                   256Gi
      huawei.com/ub_rdma:           8  # K8s has detected a total of 8 RDMA shared devices on this node.
    ...
    Allocatable:
      cpu:                      64
      ephemeral-storage:        441969440446
      memory:                   256Gi
      huawei.com/ub_rdma:           8  # K8s has detected that the total number of RDMA shared devices available for allocation on this node is 8.
    ...
    ```

## Volcano<a name="ZH-CN_TOPIC_0000002511346325"></a>

1. Run the following command to view the two Volcano Pods in the K8s cluster. The `STATUS` of both Pods must be `Running` and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n volcano-system -o wide | grep volcano
    ```

   Output example:

    ```ColdFusion
    volcano-controllers-758b6d8bdd-b7g89   1/1     Running   2          166m   192.168.x.x   ubuntu       <none>           <none>
    volcano-scheduler-86775f88f-w649w      1/1     Running   2          166m   192.168.x.x     ubuntu       <none>           <none>
    ```

2. Log in to the node where the Volcano Pods run, and run the following command to view the Volcano component logs.
   - View the logs of volcano-controllers.

       ```shell
       cat /var/log/mindx-dl/volcano-controller/volcano-controller.log
       ```

     The following output indicates that the component is running normally:

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

     The following output indicates that the component is running normally:

       ```ColdFusion
       Log file created at: 2022/10/14 11:22:32
       Running on machine: volcano-scheduler-86775f88f-6dtqf
       Binary: Built with gc go1.17.8-htrunk4 for linux/arm64
       Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
       ...
       ```

## ClusterD<a name="ZH-CN_TOPIC_0000002479386380"></a>

Perform the following procedure on any node to verify the installation status of ClusterD.

1. Run the following command to view the ClusterD Pods in the K8s cluster. The number of Pods must be 1, `STATUS` must be `Running`, and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep clusterd
    ```

   Output example:

    ```ColdFusion
    clusterd-7844cb867d-fwcj7   1/1     Running   0          2m14s   <none>   node133   <none>           <none>
    ```

2. Run the following command to query the logs of the ClusterD Pod.

    ```shell
    kubectl logs -f -n mindx-dl {ClusterD_pod_name}
    ```

   The following output indicates that the component is running normally:

    ```ColdFusion
    [INFO]     2024/07/24 13:58:30.602051 CST 1       hwlog@v0.10.12/api.go:105    cluster-info.log's logger init success
    W0724 13:58:30.602197       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.603416 CST 1       grpc/grpc_init.go:57    cluster info server start listen
    ...
    W0724 13:58:30.621433       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.621911 CST 258     job/factory.go:172    delete job summary cm goroutine started
    ```

## Ascend Operator<a name="ZH-CN_TOPIC_0000002479386462"></a>

Perform the following procedure on any node to verify the installation status of Ascend Operator.

1. Run the following command to view the Ascend Operator Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep ascend-operator
    ```

   Output example:

    ```ColdFusion
    ascend-operator-manager-b59774f7-8l5gn         1/1     Running   0          6m52s   192.168.x.x   ubuntu       <none>           <none>
    ```

2. Run the following command to view the Ascend Operator logs in the K8s cluster.

    ```shell
    kubectl logs -n mindx-dl {Ascend Operator_pod_name}
    ```

   The following output indicates that the component is running normally:

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl ascend-operator-manager-b59774f7-8l5gn
    [INFO]     2023/03/20 17:48:34.308373 1       hwlog/api.go:108    ascend-operator.log's logger init success
    [INFO]     2023/03/20 17:48:34.308469 1       ascend-operator/main.go:86    ascend-operator starting and the version is xxx
    [INFO]     2023/03/20 17:48:34.964296 1       ascend-operator/main.go:101    starting manager
    ...
    ```

## Infer Operator<a name="ZH-CN_TOPIC_0000002511346370"></a>

Perform the following procedure on any node to verify the installation status of Infer Operator.

1. Run the following command to view the Infer Operator Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep infer-operator
    ```

   Output example:

    ```ColdFusion
    infer-operator-manager-6bf95f6956-sdkbd         1/1     Running   0          6m52s   192.168.x.x   ubuntu       <none>           <none>
    ```

2. Run the following command to view the Infer Operator logs in the K8s cluster.

    ```shell
    kubectl logs -n mindx-dl {Infer Operator_pod_name}
    ```

   The following output indicates that the component is running normally:

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n mindx-dl infer-operator-manager-6bf95f6956-sdkbd
    [INFO]     2026/03/20 16:22:12.668888 1       hwlog/api.go:164    infer-operator.log's logger init success
    ...
    ```

## NodeD<a name="ZH-CN_TOPIC_0000002479386440"></a>

Perform the following procedure on any node to verify the installation status of NodeD.

1. Run the following command to view the NodeD Pod in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If NodeD is installed on multiple nodes in the cluster, each node must be confirmed.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep noded
    ```

   Output example:

    ```ColdFusion
    noded-bnmwt                        1/1     Running   10         40d    192.168.x.x     ubuntu       <none>           <none>
    ```

2. Run the following command to view the NodeD component logs.

    ```shell
    kubectl logs -n mindx-dl {NodeD_pod_name}
    ```

   The following output indicates that the component is running normally:

    ```ColdFusion
    [INFO] 2025/05/25 15:24:19.897280 1 hwlog/api.go:108 noded.log's logger init success
    [INFO] 2025/05/25 15:24:19.897392 1 noded/main.go:93 noded starting and the version is v26.2.0_linux-x86_64
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

## Container Manager<a name="ZH-CN_TOPIC_0000002492269056"></a>

Perform the following procedure on the node where the Container Manager component is deployed to verify the installation status of the Container Manager component.

1. Check the component service status. The component status must be `active(running)`.

    ```shell
    systemctl status container-manager.service
    ```

   Output example:

    ```ColdFusion
    ● container-manager.service - Ascend container manager
         Loaded: loaded (/etc/systemd/system/container-manager.service; disabled; vendor preset: enabled)
         Active: active (running) since Wed 2025-11-26 20:56:50 UTC; 16s ago
        Process: 41459 ExecStart=/bin/bash -c container-manager run  -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
       Main PID: 41464 (container-manage)
          Tasks: 10 (limit: 629145)
         Memory: 13.3M
         CGroup: /system.slice/container-manager.service
                 └─41464 /home/container-manager/container-manager run -ctrStrategy ringRecover
    ...
    ```

   >[!NOTE]
   >If information similar to the following appears in the output, it can be ignored, as it does not affect the actual functionality. The possible cause is that the IP address and subnet mask of the RoCE NIC are not configured.
   >
   >```ColdFusion
    >[dsmi_common_interface.c:1017][ascend][curpid:244135,244135][drv][dmp][dsmi_get_device_ip_address]devid 0 dsmi_cmd_get_device_ip_address return 1 error!
    >```

2. View the component log.

    ```shell
    cat /var/log/mindx-dl/container-manager/container-manager.log
    ```

   The following output uses the Atlas 800I A3 as an example:

    ```ColdFusion
    [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
    [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
    [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
    [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
    [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
    [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
    ...
    ```

   If the following output appears, the component is running normally:

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

## DPU Exporter<a name="ZH-CN_TOPIC_0000002511346364"></a>

This section uses the integration with Prometheus and reporting of Prometheus data as an example to confirm whether DPU Exporter is running properly.

**DPU Exporter Deployed Using an Image<a name="section1595201114127"></a>**

Perform the following procedure on any node to verify the installation status of DPU Exporter.

1. Run the following command to view the DPU Exporter Pods in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If DPU Exporter is installed on multiple nodes in the cluster, confirm each node one by one.

    ```shell
    kubectl get pods -n dpu-exporter -o wide | grep dpu-exporter
    ```

   Output example:

    ```ColdFusion
    dpu-exporter-4ln8w   1/1     Running   0          36m   192.0.2.x   ubuntu       <none>           <none>
    ```

2. Run the following command to view the DPU Exporter logs in the K8s cluster.

    ```shell
    kubectl logs -n dpu-exporter {dpu-exporter_pod_name}
    ```

   Output example:

    ```ColdFusion
    [2026-08-26 11:46:23.543462][INFO]     1       hwlog/api.go:164    dpu-exporter.log's logger init success
    [2026-08-26 11:46:23.543612][INFO]     1       logger/logger.go:59    starting dpu-exporter
    [2026-08-26 11:46:23.571542][INFO]     1       logger/logger.go:79    config loaded from /etc/dpu-exporter/config.json
    [2026-08-26 11:46:23.603552][INFO]     1       logger/logger.go:79    discovered 4 DPU cards
    [2026-08-26 11:46:23.603624][INFO]     1       logger/logger.go:79    huawei dpu manager initialized, dpu count: 4
    [2026-08-26 11:46:23.603658][INFO]     1       logger/logger.go:59    collector chains initialized
    [2026-08-26 11:46:23.603746][INFO]     1       logger/logger.go:79    config hot-reload watching /etc/dpu-exporter
    [2026-08-26 11:46:23.603801][INFO]     1       logger/logger.go:79    refresh dpu list, count: 4, time cost: 230ns
    [2026-08-26 11:46:23.603843][INFO]     1       logger/logger.go:79    dpu-exporter started, serving metrics on port 8083
    [2026-08-26 11:46:23.603963][INFO]     454     logger/logger.go:79    refresh dpu list, count: 4, time cost: 860ns
    [2026-08-26 11:46:23.604187][INFO]     460     logger/logger.go:79    starting prometheus metrics server on :8083
    ```

**DPU Exporter Deployed Using Binaries<a name="zh-cn_topic_0000002511346365"></a>**

Perform the following procedure on the node where DPU Exporter is installed to verify the installation status.

1. Log in to the node where DPU Exporter is deployed and run the following command to check the component service status. The component status must be `active (running)`.

    ```shell
    systemctl status dpu-exporter
    ```

   Output example:

    ```ColdFusion
    root@ubuntu:~# systemctl status dpu-exporter
    ● dpu-exporter.service - Ascend dpu exporter
         Loaded: loaded (/etc/systemd/system/dpu-exporter.service; disabled; preset: disabled)
         Active: active (running) since Tue 2026-09-15 15:36:07 CST; 12min ago
    TriggeredBy: ● dpu-exporter.timer
         Process: 102249 ExecStart=/bin/bash -c /usr/local/bin/dpu-exporter -config=/etc/dpu-exporter/config.json -port=8083 >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
       Main PID: 102252 (dpu-exporter)
          Tasks: 8 (limit: 3355442)
         Memory: 9.7M
         CGroup: /system.slice/dpu-exporter.service
                 └─102252 /usr/local/bin/dpu-exporter -config=/etc/dpu-exporter/config.json -port=8083
    ...
    ```

2. Check the component log.

    ```shell
    cat /var/log/mindx-dl/dpu-exporter/dpu-exporter.log
    ```

   Output example:

    ```ColdFusion
    [2026-08-26 11:46:23.543462][INFO]     1       hwlog/api.go:164    dpu-exporter.log's logger init success
    [2026-08-26 11:46:23.543612][INFO]     1       logger/logger.go:59    starting dpu-exporter
    [2026-08-26 11:46:23.571542][INFO]     1       logger/logger.go:79    config loaded from /etc/dpu-exporter/config.json
    [2026-08-26 11:46:23.603552][INFO]     1       logger/logger.go:79    discovered 4 DPU cards
    [2026-08-26 11:46:23.603624][INFO]     1       logger/logger.go:79    huawei dpu manager initialized, dpu count: 4
    [2026-08-26 11:46:23.603658][INFO]     1       logger/logger.go:59    collector chains initialized
    [2026-08-26 11:46:23.603746][INFO]     1       logger/logger.go:79    config hot-reload watching /etc/dpu-exporter
    [2026-08-26 11:46:23.603801][INFO]     1       logger/logger.go:79    refresh dpu list, count: 4, time cost: 230ns
    [2026-08-26 11:46:23.603843][INFO]     1       logger/logger.go:79    dpu-exporter started, serving metrics on port 8083
    [2026-08-26 11:46:23.603963][INFO]     454     logger/logger.go:79    refresh dpu list, count: 4, time cost: 860ns
    [2026-08-26 11:46:23.604187][INFO]     460     logger/logger.go:79    starting prometheus metrics server on :8083
    ...
    ```

## Ascend Dynamic Resource Allocation<a name="ZH-CN_TOPIC_0000002511426342"></a>

Ascend Dynamic Resource Allocation (hereinafter referred to as the DRA driver) is required only for users who apply for NPU resources using the K8s Dynamic Resource Allocation (DRA) mechanism. Perform the following procedure on any node to verify the installation status of the DRA driver.

**Procedure<a name="section-dra-confirm-status"></a>**

1. Run the following command to view the DRA driver Pods in the K8s cluster. The Pod `STATUS` must be `Running` and `READY` must be `1/1`. If the DRA driver is installed on multiple compute nodes in the cluster, confirm each node one by one.

    ```shell
    kubectl get pods -n kube-system -o wide | grep ascend-dra-driver
    ```

   Output example:

    ```ColdFusion
    NAME                                        READY   STATUS    RESTARTS   AGE     IP            NODE       NOMINATED NODE   READINESS GATES
    ascend-dra-driver-kubeletplugin-5m2xv       1/1     Running   0          74s     192.168.x.x  node1      <none>           <none>
    ```

   >[!NOTE]
   >If the Pod status is not `Running`, refer to [Component Pod Status Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342) for troubleshooting. If the Pod status is `ContainerCreating`, refer to [Cluster Scheduling Component Pod in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343) for troubleshooting.

2. Run the following command to view the DRA driver logs.

    ```shell
    kubectl logs -n kube-system {DRA_driver_pod_name}
    ```

   The following output indicates that the component is running normally:

    ```ColdFusion
    root@ubuntu:~# kubectl logs -n kube-system ascend-dra-driver-kubeletplugin-5m2xv
    [INFO]     2026/08/30 10:20:15.123456 1       hwlog/api.go:158    ascend-dra.log's logger init success
    [INFO]     2026/08/30 10:20:15.123567 1       main.go:44    init log module successfully.
    [INFO]     2026/08/30 10:20:15.124001 1       main.go:50    validate dra option successfully.
    [INFO]     2026/08/30 10:20:15.130221 1       main.go:57    auto-init device manager successfully.
    [INFO]     2026/08/30 10:20:15.135008 1       device/generation_910.go:69    Ascend910 enumerated 8 devices
    [INFO]     2026/08/30 10:20:15.140021 1       driver/driver.go:105    publishing ResourceSlice
    [INFO]     2026/08/30 10:20:15.141112 1       plugin/plugin.go:80    kubelet plugin registered, node=node1, driver=npu.huawei.com
    [INFO]     2026/08/30 10:20:15.142530 1       plugin/health.go:53    healthz server started, addr=11258
    [INFO]     2026/08/30 10:20:15.143008 1       main.go:71    ascend dra manager started successfully.
    ...
    ```

3. Run the following commands to verify whether NPU resources are reported successfully. If the DeviceClass named "npu.huawei.com" can be queried and a corresponding ResourceSlice exists on each compute node, the resources are reported successfully and the component is running properly.

    ```shell
    kubectl get deviceclass npu.huawei.com
    kubectl get resourceslices
    ```

   The following output is displayed. The number of chips on the node is subject to the actual situation.

    ```ColdFusion
    NAME            AGE
    npu.huawei.com  2m

    NAME                                  POOLNAME     DEVICECOUNT   AGE
    npu.huawei.com-node1-1a2b3c           node1        8             2m
    ```

## Agent Core<a name="ZH-CN_TOPIC_0000002524428761"></a>

Perform the following procedure on any node to verify the installation status of Agent Core.

**Procedure<a name="section-agent-core-confirm-status"></a>**

1. Run the following command to view the Agent Core Pod in the K8s cluster. The number of Pods must be 1, the Pod `STATUS` must be `Running`, and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep agent-core
    ```

   Output example:

    ```ColdFusion
    agent-core-7844cb867d-fwcj7   1/1     Running   0          2m14s   <none>   node133   <none>           <none>
    ```

2. Run the following command to view the Agent Core logs in the K8s cluster. The "agent-core started" log indicates that the component is running properly.

    ```shell
    kubectl logs -n mindx-dl {agent-core_pod_name}
    ```

   Output example:

    ```ColdFusion
    [2026-09-09 10:20:15,123] INFO   [agent_core.main][Process MainProcess:1][main.py:53] agent-core started: relcache + pathmap-writer + llm-secret-watch + upload gRPC :9710
    [2026-09-09 10:20:15,130] INFO   [agent_core.relcache][Process MainProcess:1][relcache.py:64] task CRs supported by agent-core: ["mindxdl.gitee.com/v1/AscendJob", "mindcluster.huawei.com/v1/InferServiceSet"]
    ...
    ```

## Node Collector<a name="ZH-CN_TOPIC_0000002524428762"></a>

Perform the following procedure on any node to verify the installation status of Node Collector.

**Procedure<a name="section-node-collector-confirm-status"></a>**

1. Run the following command to view the Node Collector Pods in the K8s cluster. Each compute node must have one Pod, the Pod `STATUS` must be `Running`, and `READY` must be `1/1`.

    ```shell
    kubectl get pods -n mindx-dl -o wide | grep node-collector
    ```

   Output example:

    ```ColdFusion
    node-collector-8j4kq   1/1     Running   0          45s   <none>   node1   <none>           <none>
    node-collector-m4j4r   1/1     Running   0          45s   <none>   node2   <none>           <none>
    ```

2. Run the following command to view the Node Collector logs in the K8s cluster. The "diag node-collector gRPC listening" log indicates that the component is running properly.

    ```shell
    kubectl logs -n mindx-dl {node-collector_pod_name}
    ```

   Output example:

    ```ColdFusion
    [2026-09-09 10:20:15,123] INFO   [node_collector.collector][Process MainProcess:1][collector.py:639] diag node-collector gRPC listening on :9720
    [2026-09-09 10:20:15,130] INFO   [node_collector.pathmap][Process MainProcess:1][pathmap.py:176] collect-manifest watch started: CM cluster-system/collect-manifest (fallback file /home/hwMindX/collect_manifest.yaml)
    ...
    ```

## Kubectl Plugin<a name="ZH-CN_TOPIC_0000002524428760"></a>

Kubectl Plugin is a client command-line tool installed on the user machine and does not include any Pod. Perform the following steps on the user machine to verify the installation status of Kubectl Plugin.

**Procedure<a name="section-kubectl-plugin-confirm-status"></a>**

1. Run the following commands to check whether the plugin is available. If the help information is displayed, the plugin is installed successfully.

    ```shell
    kubectl ascend_diag --help
    kubectl clusterops --help
    ```

   Output example:

    ```ColdFusion
    usage: kubectl-ascend_diag [-h] [--job JOB] [-n NAMESPACE] [--agent-core-service AGENT_CORE_SERVICE] [--refresh] [--json] [--collect-manifest PATH]

    Ascend fault diagnosis kubectl plugin

    options:
      -h, --help            show this help message and exit
      --job JOB             job name (job CR name)
      -n NAMESPACE, --namespace NAMESPACE
                            namespace
      --agent-core-service AGENT_CORE_SERVICE
                            agent-core Service, format svc.ns:9700
      --refresh             ignore the cache, force rerun and refresh the cache
      --json                print the full JSON response
      --collect-manifest PATH
                            write a local collect_manifest.yaml into the cluster ConfigMap and exit
    ```

2. (Optional) Run the following command to trigger the ClusterOps Agent once to confirm that the diagnosis link is working properly.

    ```shell
    kubectl ascend_diag --job {job_name}
    ```
