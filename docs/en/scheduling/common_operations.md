# Common Operations<a name="ZH-CN_TOPIC_0000002511346991"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-26T11:48:28.750Z pushedAt=2026-06-27T00:32:25.589Z -->

## Scheduling Configuration<a name="ZH-CN_TOPIC_0000002511427007"></a>

Volcano supports native K8s scheduling and can use `nodeAffinity` for scheduling. The following example uses mandatory node affinity for scheduling. For more information about the `nodeAffinity` field, see the [official Kubernetes official](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/).

- In the Volcano Job YAML, add the following fields in bold.

    <pre codetype="yaml">
    apiVersion: batch.volcano.sh/v1alpha1
    kind: Job
    metadata:
      name: mindx-test
      labels:
    ...
    spec:
    ...
      maxRetry: 3
      queue: default
      tasks:
      - name: "default-test"
        replicas: 1
        template:
          metadata:
            labels:
    ...
          spec:
            <strong>affinity:      # Add the following fields in bold</strong>
              <strong>nodeAffinity:                             # Node affinity configuration</strong>
                <strong>requiredDuringSchedulingIgnoredDuringExecution:</strong>
                  <strong>nodeSelectorTerms:                    # Node selector list</strong>
                    <strong>- matchExpressions:</strong>
                        <strong>- key: aaa               # Match nodes with label key "aaa" and value "yyy"</strong>
                          <strong>operator: In</strong>
                          <strong>values:</strong>
                            <strong>- yyy</strong>
                 podAntiAffinity:
                   requiredDuringSchedulingIgnoredDuringExecution:
    ...
              nodeSelector:
                host-arch: huawei-arm
    ...</pre>

- In the Ascend Job YAML, add the following fields in bold.

    <pre codetype="yaml">
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: test-2
    ...
    spec:
      schedulerName: volcano
      runPolicy:
        schedulingPolicy:
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
    ...
            <strong>spec:</strong>
              <strong>affinity:   #  New field</strong>
                <strong>nodeAffinity:                           # Node affinity configuration</strong>
                  <strong>requiredDuringSchedulingIgnoredDuringExecution:</strong>
                    <strong>nodeSelectorTerms:                  # Node selector list</strong>
                      <strong>- matchExpressions:</strong>
                        <strong>- key: aaa            # Match nodes with label key "aaa" and value "yyy"</strong>
                          <strong>operator: In</strong>
                          <strong>values:</strong>
                            <strong>- yyy</strong>
              nodeSelector:
                host-arch: huawei-arm
    ...</pre>

    >[!NOTE]
    >You can query node labels by running the `kubectl get node --show-labels` command. In the `LABELS` field, the value before the equal sign is the label key, and the value after the equal sign is the label value, for example, `aaa=yyy`.

## Installing NFS<a name="ZH-CN_TOPIC_0000002479227106"></a>

### Ubuntu OS<a name="ZH-CN_TOPIC_0000002479227110"></a>

Network File System (NFS) allows computers in a network to share resources. In cluster scheduling scenarios, an NFS environment is required to ensure the normal operation of training or inference jobs. NFS can be installed on the server side or the client side, and you can choose as needed.

**Installing on the Server<a name="zh-cn_topic_0000001497364925_section119917347402"></a>**

1. Log in to the storage node using an administrator account and run the following command to install the NFS server.

    ```shell
    apt install -y nfs-kernel-server
    ```

2. Fix the NFS-related ports based on the actual situation and configure the firewall for the relevant ports.
3. Create a shared directory (such as `/data/atlas_dls`) and modify the directory permissions.

    ```shell
    mkdir -p /data/atlas_dls
    chmod 750 /data/atlas_dls/
    ```

4. Append the following content to the end of the `/etc/exports` file, configure the allowed IP addresses as needed, and strengthen the relevant permission settings.

    ```shell
    /data/atlas_dls Service IP address (configure the necessary permission)
    ```

5. Start rpcbind.

    ```shell
    systemctl restart rpcbind.service
    systemctl enable rpcbind.service
    ```

6. Run the following command to check whether rpcbind has been started.

    ```shell
    systemctl status rpcbind.service
    ```

    If information similar to the following appears, the service is normal.

    ```ColdFusion
    ● rpcbind.service - RPC bind portmap service
       Loaded: loaded (/lib/systemd/system/rpcbind.service; enabled; vendor preset: enabled)
       Active: active (running) since Fri 2024-01-08 16:39:03 CST; 6 days ago
         Docs: man:rpcbind(8)
     Main PID: 2952 (rpcbind)
        Tasks: 1 (limit: 29491)
       CGroup: /system.slice/rpcbind.service
               └─2952 /sbin/rpcbind -f -w


    Jan 08 16:39:03 ubuntu-211 systemd[1]: Starting RPC bind portmap service...
    Jan 08 16:39:03 ubuntu-211 systemd[1]: Started RPC bind portmap service.
    ```

7. After rpcbind starts, start the NFS service.

    ```shell
    systemctl restart nfs-server.service
    systemctl enable nfs-server.service
    ```

8. Check whether the NFS service has started.

    ```shell
    systemctl status nfs-server.service
    ```

    If information similar to the following appears, the service is normal. If the NFS service fails to start, see the [df -h execution failure, causing NFS startup failure](https://gitcode.com/Ascend/mind-cluster/issues/353) section for troubleshooting.

    ```ColdFusion
    ● nfs-server.service - NFS server and services
       Loaded: loaded (/lib/systemd/system/nfs-server.service; enabled; vendor preset: enabled)
       Active: active (exited) since Fri 2024-01-08 16:39:03 CST; 6 days ago
     Main PID: 3220 (code=exited, status=0/SUCCESS)
        Tasks: 0 (limit: 29491)
       CGroup: /system.slice/nfs-server.service


    Jan 08 16:39:03 ubuntu-211 systemd[1]: Starting NFS server and services...
    Jan 08 16:39:03 ubuntu-211 exportfs[3181]: exportfs: /etc/exports [1]: Neither 'subtree_check' or 'no_subtree_check' specified for export "*:/data/atlas_dls".
    Jan 08 16:39:03 ubuntu-211 exportfs[3181]:   Assuming default behaviour ('no_subtree_check').
    Jan 08 16:39:03 ubuntu-211 exportfs[3181]:   NOTE: this default has changed since nfs-utils version 1.0.x
    Jan 08 16:39:03 ubuntu-211 systemd[1]: Started NFS server and services.
    ```

9. View the mount permissions of the shared directory (such as `/data/atlas_dls`).

    ```shell
    cat /var/lib/nfs/etab
    ```

    If information similar to the following appears, the service is normal.

    ```ColdFusion
    /data/atlas_dls *(rw,...Configured permission)
    ```

**Installing on the Client<a name="zh-cn_topic_0000001497364925_section10189114704512"></a>**

Log in to other servers using an administrator account and run the following command to install the NFS client.

```shell
apt install -y nfs-common
```

### CentOS OS<a name="ZH-CN_TOPIC_0000002511427005"></a>

NFS is a network file system that allows computers on a network to share resources. In cluster scheduling scenarios, an NFS environment is required to ensure the normal running of training or inference tasks. NFS can be installed on the server or the client, and you can choose based on your needs.

**Install on the server<a name="zh-cn_topic_0000001446805000_section1398218463486"></a>**

1. Log in to the storage node using an administrator account and run the following command to install the NFS server.

    ```shell
    yum install nfs-utils -y
    ```

2. Fix the NFS-related ports and configure the firewall for these ports based on your actual situation.
3. Create a shared directory (such as `/data/atlas_dls`) and modify the directory permissions.

    ```shell
    mkdir -p /data/atlas_dls
    chmod 750 /data/atlas_dls/
    ```

4. Run the `vi /etc/exports` command, append the following content to the end of the file, configure the allowed IP addresses as needed, and strengthen the relevant permission settings.

    ```shell
    /data/atlas_dls Service IP address (configure necessary permission)
    ```

5. Start rpcbind.

    ```shell
    systemctl restart rpcbind.service
    systemctl enable rpcbind.service
    ```

6. Check whether rpcbind has started.

    ```shell
    systemctl status rpcbind.service
    ```

    If information similar to the following appears, the service is normal.

    ```ColdFusion
    ● rpcbind.service - RPC bind service
       Loaded: loaded (/usr/lib/systemd/system/rpcbind.service; enabled; vendor preset: enabled)
       Active: active (running) since Fri 2024-01-15 15:54:44 CST; 28s ago
     Main PID: 63008 (rpcbind)
       CGroup: /system.slice/rpcbind.service
               └─63008 /sbin/rpcbind -w


    Jan 15 15:54:44 centos39 systemd[1]: Starting RPC bind service...
    Jan 15 15:54:44 centos39 systemd[1]: Started RPC bind service.
    ```

7. After rpcbind starts, start the NFS service.

    ```shell
    systemctl restart nfs-server.service
    systemctl enable nfs-server.service
    ```

8. Check whether the NFS service has started.

    ```shell
    systemctl status nfs-server.service
    ```

    If information similar to the following appears, the service is normal. If the NFS service fails to start, see the [df -h execution failure, causing NFS startup failure](https://gitcode.com/Ascend/mind-cluster/issues/353) section for troubleshooting.

    ```ColdFusion
    ● nfs-server.service - NFS server and services
       Loaded: loaded (/usr/lib/systemd/system/nfs-server.service; enabled; vendor preset: disabled)
      Drop-In: /run/systemd/generator/nfs-server.service.d
               └─order-with-mounts.conf
       Active: active (exited) since Fri 2024-01-15 15:56:15 CST; 8s ago
     Main PID: 67145 (code=exited, status=0/SUCCESS)
       CGroup: /system.slice/nfs-server.service


    Jan 15 15:56:15 centos39 systemd[1]: Starting NFS server and services...
    Jan 15 15:56:15 centos39 systemd[1]: Started NFS server and services.
    ```

9. View the mount permissions of the shared directory (e.g., `/data/atlas_dls`).

    ```shell
    cat /var/lib/nfs/etab
    ```

    If the following response appears, the service is normal.

    ```ColdFusion
    /data/atlas_dls *(rw,...Configured permission)
    ```

**Installing on the Client<a name="zh-cn_topic_0000001446805000_section1862665118118"></a>**

1. Log in to other servers using an administrator account and run the following command to install the NFS client.

    ```shell
    yum install nfs-utils -y
    ```

2. Start rpcbind.

    ```shell
    systemctl restart rpcbind.service
    systemctl enable rpcbind.service
    ```

3. Check whether rpcbind is started.

    ```shell
    systemctl status rpcbind.service
    ```

    If information similar to the following appears, the service is normal.

    ```ColdFusion
    ● rpcbind.service - RPC Bind
       Loaded: loaded (/usr/lib/systemd/system/rpcbind.service; enabled; vendor preset: enabled)
       Active: active (running) since Thu 2024-03-14 04:59:22 EDT; 8s ago
         Docs: man:rpcbind(8)
     Main PID: 1681425 (rpcbind)
        Tasks: 1 (limit: 3355442)
       Memory: 956.0K
       CGroup: /system.slice/rpcbind.service
               └─1681425 /usr/bin/rpcbind -w -f
    Mar 14 04:59:22 localhost.localdomain systemd[1]: Starting RPC Bind...
    Mar 14 04:59:22 localhost.localdomain systemd[1]: Started RPC Bind.
    ```

4. After rpcbind is started, start the NFS service.

    ```shell
    systemctl restart nfs-server.service
    systemctl enable nfs-server.service
    ```

5. Check whether the NFS service is started.

    ```shell
    systemctl status nfs-server.service
    ```

If information similar to the following appears, the service is normal.

    ```ColdFusion
    ● nfs-server.service - NFS server and services
       Loaded: loaded (/usr/lib/systemd/system/nfs-server.service; enabled; vendor preset: disabled)
      Drop-In: /run/systemd/generator/nfs-server.service.d
               └─order-with-mounts.conf
       Active: active (exited) since Thu 2024-03-14 04:59:40 EDT; 8s ago
     Main PID: 1681567 (code=exited, status=0/SUCCESS)
        Tasks: 0 (limit: 3355442)
       Memory: 0B
       CGroup: /system.slice/nfs-server.service
    Mar 14 04:59:39 localhost.localdomain systemd[1]: Starting NFS server and services...
    Mar 14 04:59:39 localhost.localdomain exportfs[1681536]: exportfs: Failed to stat /data/atlas_dls: No such file or directory
    Mar 14 04:59:40 localhost.localdomain systemd[1]: Started NFS server and services.
    ```

1. (Optional) NFS requires the `mount` and `umount` commands. Generally, the system has the `mount` command built-in. If the current client does not have this command, run the following command.

    ```shell
    yum install -y  util-linux
    ```

## Querying Reported Fault Information<a name="ZH-CN_TOPIC_0000002479387090"></a>

### Volcano<a name="ZH-CN_TOPIC_0000002479387088"></a>

Volcano collects internal chip faults, parameter plane network faults, and node fault information, and places them as external information in the K8s ConfigMap for external query and use.

The query command is `kubectl describe cm -n volcano-system vcjob-fault-npu-cm`. The command response example is as follows. For key parameterdescriptions, see [Table 2 vcjob-fault-npu-cm field description](./api/volcano.md#job-information).

```ColdFusion
Name:         vcjob-fault-npu-cm
Namespace:    volcano-system
Labels:       <none>
Annotations:  <none>

Data
====
fault-node:
----
[{"FaultDeviceList":[{"fault_type":"CardNetworkUnhealthy","npu_name":"Ascend910-0","fault_level":"PreSeparateNPU","fault_handling":"PreSeparateNPU","large_model_fault_level":"PreSeparateNPU","fault_code":"81078603"},{"fault_type":"CardUnhealthy","npu_name":"Ascend910-4","fault_level":"SeparateNPU","fault_handling":"SeparateNPU","large_model_fault_level":"SeparateNPU","fault_code":"A8028801,A4028801,80E18402,80E18401"}],"NodeName":"node133","UnhealthyNPU":["Ascend910-4"],"NetworkUnhealthyNPU":["Ascend910-0"],"NodeDEnable":true,"NodeHealthState":"CardUnhealthy","UpdateTime":1744182212}]
remain-retry-times:
----


BinaryData
====

Events:  <none>
```

### Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511347041"></a>

#### Fault Information<a name="ZH-CN_TOPIC_0000002479387086"></a>

Ascend Device Plugin collects internal chip faults, parameter plane network faults, and node faults, and places them as external information in K8s ConfigMaps. One ConfigMap stores the information of one node for external query and use.

Query command: `kubectl describe cm -n kube-system mindx-dl-deviceinfo-$*_\{node\_name\}_`

Taking <term>Atlas A3 Training Series Products</term> as an example, the response example is as follows. The response parameters may vary for different devices, and the actual output shall prevail. For key parameter descriptions, see [Table 1 DeviceInfoCfg](./api/ascend_device_plugin.md#chip-resources).

```ColdFusion
{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[{\"fault_type\":\"CardNetworkUnhealthy\",\"npu_name\":\"Ascend910-0\",\"large_model_fault_level\":\"PreSeparateNPU\",\"fault_level\":\"PreSeparateNPU\",\"fault_handling\":\"PreSeparateNPU\",\"fault_code\":\"81078603\",\"fault_time_and_level_map\":{\"81078603\":{\"fault_time\":1744168468259,\"fault_level\":\"PreSeparateNPU\"}}},{\"fault_type\":\"CardUnhealthy\",\"npu_name\":\"Ascend910-4\",\"large_model_fault_level\":\"SeparateNPU\",\"fault_level\":\"SeparateNPU\",\"fault_handling\":\"SeparateNPU\",\"fault_code\":\"A8028801,A4028801,80E18402,80E18401\",\"fault_time_and_level_map\":{\"80E18401\":{\"fault_time\":1744167455784,\"fault_level\":\"NotHandleFault\"},\"80E18402\":{\"fault_time\":1744167455784,\"fault_level\":\"SeparateNPU\"},\"A4028801\":{\"fault_time\":1744167455784,\"fault_level\":\"NotHandleFault\"},\"A8028801\":{\"fault_time\":1744167455784,\"fault_level\":\"SeparateNPU\"}}}]","huawei.com/Ascend910-NetworkUnhealthy":"Ascend910-0","huawei.com/Ascend910-Recovering":"","huawei.com/Ascend910-Unhealthy":"Ascend910-4"},"UpdateTime":1744182144},"SuperPodID":-2,"ServerIndex":-2,"CheckCode":"a550811fdfafb5717555526816af2ca4ac6c3e102f5907574048578e0c8fcc73"}
```

#### Fault Event Information<a name="ZH-CN_TOPIC_0000002511347039"></a>

Fault events collected by Ascend Device Plugin can be reported through K8s event events. The query command is `kubectl get events -n kube-system`. Taking Atlas Training Series Products as an example, the response example is as follows. For parameter descriptions, see [Table 1](#table66076214393).

```ColdFusion
NAMESPACE     LAST SEEN   TYPE      REASON     OBJECT                                         MESSAGE
kube-system   8s          Warning   Occur      pod/ascend-device-plugin-daemonset-910-dlpmv   device fault, nodeName:k8smaster, assertion:Occur, cardID:2, deviceID:0, faultCodes:8C084E00, faultLevelName:RestartBusiness, alarmRaisedTime:2023-11-21 05:36:53
```

**Table 1**  Parameter description

<a name="table66076214393"></a>

|Name|Description|
|--|--|
|NAMESPACE|Namespace name, with the value kube-system.|
|LAST SEEN|Time when the event occurred.|
|TYPE|<p>Event type, with values of <span>"Normal"</span> and <span>"Warning"</span>.</p>|
|REASON|<p>Reason for the event. The values are described as follows:</p><ul><li>Occur: fault occurrence</li><li>Recovery: Fault recovery</li><li>Notice: Notification</li></ul>|
|OBJECT|<p>Event object, with the value specification of pod/<span><em>Ascend Device Plugin</em></span><em> Pod name</em>, such as pod/ascend-device-plugin-daemonset-910-dlpmv.</p>|
|MESSAGE|<p>Description of the event information content. The fields of the event content are described as follows:</p><ul><li>nodeName: node name</li><li>assertion: information type<ul><li>Occur: fault occurrence</li><li>Recovery: Fault recovery</li><li>Notice: Notification</li></ul></li><li>cardID: NPU management unit ID (NPU device ID)</li><li>deviceID: Device number</li><li>faultCodes: Fault code, such as 8C084E00</li><li>faultLevelName: Fault level name<ul><li>NotHandleFault: no handling required</li><li>RestartRequest: <span>affect services execution; need to re-execute service requests</span></li><li>RestartBusiness: <span>affect services execution;</span> need to restart services</li><li>FreeRestartNPU: affect services execution; need to reset the chip when it is idle</li><li>RestartNPU: directly reset the chip and re-execute services</li><li>SeparateNPU: isolate chip</li><li>PreSeparateNPU: does not affect services temporarily, and no more jobs will be scheduled to this chip subsequently.</li><li>SubHealthFault: dandled according to the subHealthyStrategy parameter configured in the job YAML</li></ul></li><li>alarmRaisedTime: fault occurrence time</li></ul>|

### ClusterD<a name="ZH-CN_TOPIC_0000002511347035"></a>

ClusterD collects internal node faults, chip faults, and UnifiedBus device faults, and places them as external information in the K8s ConfigMap for external query and use.

**Node Fault<a name="section208771421687"></a>**

Query command: `kubectl describe cm -n mindx-dl cluster-info-node-cm`

Taking the <term>Atlas A3 Training Series products</term> as an example, the response example is as follows. The response parameters may vary for different devices, and the actual output shall prevail. For key parameter descriptions, see [Table 1 cluster-info-node-cm](./api/clusterd/00_cluster_resources.md#configmap-description).

```ColdFusion
{"mindx-dl-nodeinfo-kwok-node-0":{"FaultDevList":[],"NodeStatus":"Healthy","CmName":"mindx-dl-nodeinfo-kwok-node-0"},"mindx-dl-deviceinfo-kwok-node-1001":{"FaultDevList":[],"NodeStatus":"Healthy","CmName":"mindx-dl-nodeinfo-kwok-node-1001"}}
```

**Chip Fault<a name="section834865016504"></a>**

Query command: **kubectl describe cm -n mindx-dl cluster-info-device-$**_\{m\}_

*m* is an integer incrementing from 0. For every additional 1000 nodes in ae cluster, a new ConfigMap file `cluster-info-device-$\{m\}` is added.

Taking the <term>Atlas A3 Training Series Products</term> as an example, the response example is as follows. The response parameters may vary for different devices, and the actual output shall prevail. For key parameter descriptions, see [Table 2 cluster-info-device-$\{m\}](./api/clusterd/00_cluster_resources.md#configmap-description).

```ColdFusion
{"mindx-dl-deviceinfo-kwok-node-0":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1693899390,"CmName":"mindx-dl-deviceinfo-kwok-node-0","SuperPodID":0,"ServerIndex":0},"mindx-dl-deviceinfo-kwok-node-1001":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1693899390,"CmName":"mindx-dl-deviceinfo-kwok-node-1001","SuperPodID":0,"ServerIndex":0}}
```

**UnifiedBus Device Fault<a name="section1728713587242"></a>**

Query command: **kubectl describe cm -n mindx-dl cluster-info-switch-$**_\{m\}_

*m* is an integer incrementing from 0. For every additional 2000 nodes in a cluster, a new ConfigMap file `cluster-info-switch-$\{m\}` is created.

Taking <term>Atlas A3 Training Series Products</term> as an example, the response example is as follows. The response parameters may vary for different devices, and the actual output shall prevail. For key parameter descriptions, see [Table 1](#table9246232250).

```ColdFusion
{"FaultCode":[000001c1],"FaultLevel":"NotHandle","UpdateTime":1722845555,"NodeStatus":"Healthy"}
```

**Table 1**  Lingqu bus device fault parameter description

<a name="table9246232250"></a>

|Name|Description|
|--|--|
|FaultCode|Fault code, a string composed of English letters and numbers, where the string represents the fault code in hexadecimal.|
|FaultLevel|<p>The handling policy corresponding to the highest-level fault among the current faults.</p><ul><li>NotHandle: No action is taken.</li><li>SubHealth: Handled according to the configured policy.</li><li>Reset: Isolate the node.</li><li>Separate: Isolate the node.</li><li>RestartRequest: Isolate the node.</li></ul>|
|UpdateTime|ConfigMap update time.|
|NodeStatus|<p>Current node status.</p><ul><li>Healthy: The node is healthy.</li><li>SubHealthy: The node is pre-isolated. The current job is not processed, and subsequent jobs will no longer be scheduled to this node.</li><li>UnHealthy: The node is unhealthy. Isolate the node and reschedule jobs.</li></ul>|

### NodeD<a name="ZH-CN_TOPIC_0000002511427003"></a>

NodeD collects node fault information and node health status information, and places it as external information in a K8s ConfigMap for external query and use.

The query command is `kubectl describe cm mindx-dl-nodeinfo-<nodename> -n mindx-dl`. A command response example is shown below. For key parameter descriptions, see [Table 1 mindx-dl-nodeinfo-_<nodename\>_](./api/noded.md#node-resources).

```ColdFusion
Name:         mindx-dl-nodeinfo-<nodename>
Namespace:    mindx-dl
Labels:       <none>
Annotations:  <none>

Data
====
NodeInfo:
----
{"NodeInfo":{"FaultDevList":[{"DeviceType":"CPU","DeviceId":1,"FaultCode":["00000011"],"FaultLevel":"SeparateFault"}],"NodeStatus":"UnHealthy"},"CheckCode":"3a2934c3cb875f2256c770c75a6fdf24594fcf64481ac6cd0d0f74b8fea88855"}
Events:  <none>
```

## Creating an Image<a name="ZH-CN_TOPIC_0000002479227114"></a>

### Building a Container Image Using Dockerfile (PyTorch)<a name="ZH-CN_TOPIC_0000002511426595"></a>

**Prerequisites<a name="zh-cn_topic_0000001497364957_section193545302315"></a>**

As shown in [Table 1](#zh-cn_topic_0000001497364957_table13971125465512), obtain the packages for the corresponding operating system, as well as the Dockerfile and script files required for image packaging.

In the package name, *{version}* indicates the version, *{arch}* indicates the architecture, and *{chip_type}* indicates the chip type. The corresponding CANN packages for versions 6.3.RC3, 6.2.RC3, and later include an installation prompt "Do you accept the EULA to install CANN (Y/N)"; in the Dockerfile writing examples, the installation command includes the `--quiet` parameter to accept the EULA by default, which you can modify as needed.

**Table 1** Required software

<a name="zh-cn_topic_0000001497364957_table13971125465512"></a>

|Software Package|Description|Obtaining Method|
|--|--|--|
|Ascend-cann-toolkit_<i>{version}</i>_linux-<i>{arch}</i>.run|CANN Toolkit.|<p>[Download Link](https://www.hiascend.com/developer/download/community/result?module=cann)</p>|
|<p>Ascend-cann-<em>{chip_type}</em>-ops_<em>{version}</em>_linux-<em>{arch}</em>.run</p>|<p>CANN operator package.</p><p>Before CANN 8.5.0, this package was named Ascend-cann-kernels-<em>{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run</p>|[Download Link](https://www.hiascend.com/developer/download/community/result?module=cann)|
|apex-0.1+ascend-cp3x-cp3x-linux_<em>{arch}</em>.whl|<p>Mixed precision module.</p><p>cp3x in the package name indicates the Python version. For example, x being 10 indicates Python 3.10.</p>|Compile the APEX software package based on the actual situation.|
|<ul><li><span>x86_64</span>: torch-<em>v{version}</em>+cpu-cp3x-cp3x-linux_x86_64.whl</li><li><span>ARM</span>: torch-<em>v{version}</em>-cp3x-cp3x-manylinux_2_17_aarch64.manylinux2014_aarch64.whl</li></ul>|<p>Official <span>PyTorch</span> package.</p><p>cp3x in the package name indicates the Python version. For example, x being 10 indicates Python 3.10.</p><p><em>{version}</em> indicates the <span>PyTorch</span> version. Currently, <span>PyTorch</span> 2.1.0 to 2.7.1 is supported.</p>|<p>[Download Link](https://download.pytorch.org/whl/torch/)</p><p>Select the <span>PyTorch</span> version to install based on the actual situation.</p>|
|<p>torch_npu-<em>v{version}</em><em>.</em>post<em>{version}</em>-cp3x-cp3x-manylinux_2_17_<em>{arch}</em>.manylinux2014_<em>{arch}</em>.whl</p>|<p><span>Ascend Extension for PyTorch</span> plugin.</p><p>cp3x in the package name indicates the Python version. For example, x being 10 indicates Python 3.10.</p>|<p>[Download Link](https://www.hiascend.com/document/detail/zh/Pytorch/600/configandinstg/instg/insg_0001.html)</p><ul><li>Select a torch_npu version that is compatible with <span>PyTorch</span>.</li><li>If using <span>PyTorch</span> models from the MindSpeed-LLM repository, <span>Ascend Extension for PyTorch</span> 2.1.0 or later is required.</li></ul>|
|Dockerfile|Required for creating images.|Refer to [Dockerfile Writing Example](#zh-cn_topic_0000001497364957_li104026527188)|
|dllogger-master|PyTorch logging tool.|[Download Link](https://github.com/NVIDIA/dllogger)|
|ascend_install.info|Driver installation information file.|Copy the "/etc/ascend_install.info" file from the host.|
|version.info|Driver version information file.|Copy the "/usr/local/Ascend/driver/version.info" file from the host.|
|prebuild.sh|Performs preparation for training runtime environment installation, such as configuring proxies.|Refer to [Step 3](#zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li206652021677)|
|install_ascend_pkgs.sh|Ascend software package installation script.|Refer to [Step 4](#zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li538351517716)|
|postbuild.sh|Clears installation packages, scripts, proxy configurations, etc., that do not need to be kept in the container.|Refer to [Step 5](#zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li154641047879)|

To prevent software packages from being maliciously tampered with during transmission or storage, you need to download the corresponding digital signature file for integrity verification when downloading the software packages.

After downloading the software packages, refer to the *[OpenPGP Signature Verification Guide](https://support.huawei.com/enterprise/en/doc/EDOC1100209376)* to perform PGP digital signature verification on the software packages downloaded from the Support website. If the verification fails, do not use the software package and contact Huawei technical support engineers first.

Before installing or upgrading using a software package, you must also verify the digital signature of the software package following the above process to ensure that the software package has not been tampered with.

For carrier customers, please visit [https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction).

For enterprise customers, please visit [https://support.huawei.com/enterprise/en/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/en/tool/pgp-verify-TL1000000054).

>[!NOTE]
>This chapter uses the Ubuntu OS with Python 3.10 and CANN 8.5.0 as an example to introduce the detailed process of building a container image using a Dockerfile. You need to modify the relevant steps according to the actual situation during use.

**Procedure<a name="zh-cn_topic_0000001497364957_section38151530134817"></a>**

1. Upload the prepared software packages, deep learning framework-related packages, host-side driver installation information file, and driver version information file to the same directory on the server (such as `/home/test`).
    - Ascend-cann-toolkit\__\{version\}_\_linux-_\{arch\}_.run
    - Ascend-cann-_\{chip\_type\}_-ops\__\{version\}_\_linux-_\{arch\}_.run
    - apex-0.1+ascend-cp310-cp310-linux\__\{arch\}_.whl
    - torch-_v\{version\}_+cpu.cxx11.abi-cp310-cp310-linux\__\{arch\}_.whl or torch-_v\{version\}_-cp3x-cp3x-manylinux\_2\_17\_aarch64.manylinux2014\_aarch64.whl
    - torch\_npu-_v\{version\}_.post<i>\{version\}</i>-cp310-cp310-manylinux\_2\_17\__\{arch\}_.manylinux2014\__\{arch\}_.whl
    - dllogger-master
    - ascend\_install.info
    - version.info

2. Log in to the server as the `root` user.
3. <a name="zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li206652021677"></a>Perform the following steps to prepare the `prebuild.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `prebuild.sh` file.

        ```shell
        vi prebuild.sh
        ```

    2. For the content to be written, refer to the [prebuild.sh](#zh-cn_topic_0000001497364957_li270512519175) writing example. After writing, run the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

4. <a name="zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li538351517716"></a>Perform the following steps to prepare the `install_ascend_pkgs.sh` file.
    1. Go to the directory where software packages are located and run the following command to create the `install_ascend_pkgs.sh` file.

        ```shell
        vi install_ascend_pkgs.sh
        ```

    2. For the content to write, refer to the [install_ascend_pkgs.sh](#zh-cn_topic_0000001497364957_li58501140151720) writing example. After writing, execute the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

5. <a name="zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li154641047879"></a>Perform the following steps to prepare the `postbuild.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `postbuild.sh` file.

        ```shell
        vi postbuild.sh
        ```

    2. For the content to write, refer to the [postbuild.sh](#zh-cn_topic_0000001497364957_li14267051141712) writing example. After writing, execute the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

6. Perform the following steps to prepare a `Dockerfile` file.
    1. Go to the directory where the software packages are located and run the following command to create a `Dockerfile` file (file name example: "Dockerfile").

        ```shell
        vi Dockerfile
        ```

    2. For the content to be written, refer to the [Dockerfile](#zh-cn_topic_0000001497364957_li104026527188) writing example, then run the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

7. Go to the directory where the software packages are located and run the following command to build the container image. **Note: Do not omit the the dot (".") at the end of the command.**

    ```shell
    docker build -t Image name_OS architecture: Image tag .
    ```

    In the above command, the description of each parameter is shown in the following table.

    **Table 2** Command parameters

    <a name="table18728186182510"></a>

    |Name|Description|
    |--|--|
    |-t|Specifies the image name.|
    |Image name_OS architecture: Image tag|Image name and tag. Enter the actual values.|

    Example:

    ```shell
    docker build -t test_train_arm64:v1.0 .
    ```

    When "Successfully built xxx" appears, it indicates that the image has been built successfully.

8. After the build is complete, run the following command to view the image information.

    ```shell
    docker images
    ```

    The response example is as follows.

    ```ColdFusion
    REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
    test_train_arm64    v1.0                d82746acd7f0        27 minutes ago      749MB
    ```

**Writing Examples<a name="zh-cn_topic_0000001497364957_section3523631151714"></a>**

1. <a name="zh-cn_topic_0000001497364957_li270512519175"></a>`prebuild.sh`

    Ubuntu Arm:

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # Write script code here using bash syntax for installation preparation, such as configuring proxies
    # This script will be executed before the formal build process starts
    #
    # Note: This script will not be automatically removed after execution. If it does not need to be retained in the image, please remove it in the postbuild.sh script
    #--------------------------------------------------------------------------------
    # DNS configuration
    tee /etc/resolv.conf <<- EOF
    nameserver xxx.xxx.xxx.xxx  #DNS server IP, multiple can be specified, configure according to actual setup
    nameserver xxx.xxx.xxx.xxx
    nameserver xxx.xxx.xxx.xxx
    EOF
    # apt proxy configuration
    tee /etc/apt/apt.conf.d/80proxy <<- EOF
    Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";  # HTTP proxy server IP address and port
    Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";  # HTTPS proxy server IP address and port
    EOF
    chmod 777 -R /tmp
    rm /var/lib/apt/lists/*
    # apt source configuration (using Ubuntu 18.04 Arm source as an example; configure according to actual setup)
    tee /etc/apt/sources.list <<- EOF
    deb http://mirrors.aliyun.com/ubuntu-ports/ bionic main restricted universe multiverse
    deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic main restricted universe multiverse
    deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-security main restricted universe multiverse
    deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-security main restricted universe multiverse
    deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-updates main restricted universe multiverse
    deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-updates main restricted universe multiverse
    deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-proposed main restricted universe multiverse
    deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-proposed main restricted universe multiverse
    deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-backports main restricted universe multiverse
    deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-backports main restricted universe multiverse
    EOF
    ```

    Ubuntu x86_64:

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------

    # Use bash syntax to write script code here to for installation preparation, such as configuring proxies
    # This script will be executed before the formal build process starts
    #
    # Note: This script will not be automatically removed after execution. If it does not need to be retained in the image, please clean it up in the postbuild.sh script
    #--------------------------------------------------------------------------------
    # apt proxy settings
    tee /etc/apt/apt.conf.d/80proxy <<- EOF
    Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";    #HTTP proxy server IP address and port
    Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";   # HTTPS proxy server IP address and port
    EOF

    #apt source configuration (using Ubuntu 18.04 x86_64 source as an example; configure according to actual setup)
    tee /etc/apt/sources.list <<- EOF
    deb http://mirrors.ustc.edu.cn/ubuntu/ bionic main multiverse restricted universe
    deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-backports main multiverse restricted universe
    deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-proposed main multiverse restricted universe
    deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-security main multiverse restricted universe
    deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-updates main multiverse restricted universe
    deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic main multiverse restricted universe
    deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-backports main multiverse restricted universe
    deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-proposed main multiverse restricted universe
    deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-security main multiverse restricted universe
    deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-updates main multiverse restricted universe
    EOF
    ```

2. <a name="zh-cn_topic_0000001497364957_li58501140151720"></a>`install_ascend_pkgs.sh`

    ```shell
    #--------------------------------------------------------------------------------
    # Use bash syntax to write script code here to install Ascend software packages
    #
    # Note: This script will not be automatically removed after execution. If it does not need to be retained in the image, please clean it up in the postbuild.sh script
    #--------------------------------------------------------------------------------
    umask 0022
    cp ascend_install.info /etc/
    # Before building, copy /usr/local/Ascend/driver/version.info from the host to the current directory
    mkdir -p /usr/local/Ascend/driver/
    cp version.info /usr/local/Ascend/driver/
    # Ascend-cann-toolkit_{version}_linux-{arch}.run
    chmod +x Ascend-cann-toolkit_{version}_linux-{arch}.run
    chmod +x Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run
    ./Ascend-cann-toolkit_{version}_linux-{arch}.run --install-path=/usr/local/Ascend/ --install --quiet
    echo y | ./Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run --install
    # Only installed for the toolkit package, so cleanup is needed; it will be mounted via ascend docker when the container starts
    rm -f version.info
    rm -rf /usr/local/Ascend/driver/
    ```

3. <a name="zh-cn_topic_0000001497364957_li14267051141712"></a>postbuild.sh

    ```shell
    #--------------------------------------------------------------------------------
    # Please use bash syntax to write script code here to clean up installation packages, scripts, proxy configurations, etc. that do not need to be retained in the container
    # This script will be executed after the formal build process is completed
    #
    # Note: This script will be automatically cleared after execution and will not remain in the image; the script location and Working Dir location are /tmp
    #--------------------------------------------------------------------------------
    rm -f ascend_install.info
    rm -f prebuild.sh
    rm -f install_ascend_pkgs.sh
    rm -f Dockerfile
    rm -f Ascend-cann-toolkit_{version}_linux-{arch}.run
    rm -f Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run
    rm -f apex-0.1+ascend-cp310-cp310-linux_{arch}.whl
    rm -f torch-v{version}+cpu.cxx11.abi-cp310-cp310-linux_{arch}.whl
    rm -f torch_npu-v{version}.post7-cp310-cp310-manylinux_2_17_{arch}.manylinux2014_{arch}.whl
    rm -f /etc/apt/apt.conf.d/80proxy

    ```

4. <a name="zh-cn_topic_0000001497364957_li104026527188"></a>Dockerfile
    - Ubuntu Arm with Python 3.10:

        ```Dockerfile
        FROM ubuntu:18.04
        ARG PYTORCH_PKG=torch-v{version}+cpu.cxx11.abi-cp310-cp310-linux_aarch64.whl
        ARG PYTORCH_NPU_PKG=torch_npu-v{version}.post{version}-cp310-cp310-manylinux_2_17_aarch64.manylinux2014_aarch64.whl
        ARG APEX_PKG=apex-0.1_ascend-cp310-cp310-linux_aarch64.whl
        ARG HOST_ASCEND_BASE=/usr/local/Ascend
        ARG TOOLKIT_PATH=/usr/local/Ascend/cann
        ARG INSTALL_ASCEND_PKGS_SH=install_ascend_pkgs.sh
        ARG PREBUILD_SH=prebuild.sh
        ARG POSTBUILD_SH=postbuild.sh
        WORKDIR /tmp
        COPY . ./
        # Trigger prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH || true"
        ENV http_proxy http://xxx.xxx.xxx.xxx:xxx
        ENV https_proxy http://xxx.xxx.xxx.xxx:xxx
        # System packages
        RUN apt update && \
            apt install -y --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 \
                patch libbz2-dev llvm libncursesw5-dev xz-utils liblzma-dev m4 dos2unix libopenblas-dev libsqlite3-dev
        RUN wget https://www.python.org/ftp/python/3.10.5/Python-3.10.5.tgz
        RUN tar -zxvf Python-3.10.5.tgz && cd Python-3.10.5 && ./configure --prefix=/usr/local/python3.10.5 --enable-shared && make && make install
        RUN ln -s /usr/local/python3.10.5/bin/python3.10 /usr/local/python3.10.5/bin/python && \
            ln -s /usr/local/python3.10.5/bin/pip3.10 /usr/local/python3.10.5/bin/pip
        # Configure Python pip source
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        ENV LD_LIBRARY_PATH=/usr/local/python3.10.5/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.10.5/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.10.5/lib/python3.10/site-packages:$PYTHONPATH
        # Python packages
        RUN pip3 install decorator && \
            pip3 install sympy && \
            pip3 install cffi && \
            pip3 install pyyaml && \
            pip3 install pathlib2 && \
            pip3 install grpcio && \
            pip3 install grpcio-tools && \
            pip3 install protobuf && \
            pip3 install scipy && \
            pip3 install requests && \
            pip3 install attrs && \
            pip3 install Pillow==9.1.0 && \
            pip3 install torchvision==0.16.0 && \
            pip3 install numpy==1.23.5 && \
            pip3 install psutil && \
            pip3 install absl-py

        # Create the HwHiAiUser user and owner. Ensure the UID and GID are consistent with the physical machine to avoid ownerless files. In this example, the user and corresponding group are automatically created, with both UID and GID set to 1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
        # Ascend packages
        RUN umask 0022 && bash $INSTALL_ASCEND_PKGS_SH
        RUN umask 0022 && pip3 install $APEX_PKG
        RUN umask 0022 && pip3 install $PYTORCH_PKG
        RUN umask 0022 && pip3 install $PYTORCH_NPU_PKG
        RUN cd /tmp/dllogger-master/ && \
            python3 setup.py build && \
            python3 setup.py install
        # Environment variables
        ENV HCCL_WHITELIST_DISABLE=1
        ENV PYTHONPATH=/tmp/dllogger-master
        # Create /lib64/ld-linux-aarch64.so.1
        RUN umask 0022 && \
            if [ ! -d "/lib64" ]; \
            then \
                mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \
            fi
        ENV http_proxy ""
        ENV https_proxy ""
        # Trigger postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH || true" && \
            rm $POSTBUILD_SH
        ```

    - Ubuntu x86_64 with Python 3.10:

        ```Dockerfile
        FROM ubuntu:18.04
        ARG PYTORCH_PKG=torch-v{version}+cpu.cxx11.abi-cp310-cp310-linux_x86_64.whl
        ARG PYTORCH_NPU_PKG=torch_npu-v{version}.post{version}-cp310-cp310-manylinux_2_17_x86_64.manylinux2014_x86_64.whl
        ARG APEX_PKG=apex-0.1_ascend-cp310-cp310-linux_x86_64.whl
        ARG HOST_ASCEND_BASE=/usr/local/Ascend
        ARG TOOLKIT_PATH=/usr/local/Ascend/cann
        ARG INSTALL_ASCEND_PKGS_SH=install_ascend_pkgs.sh
        ARG PREBUILD_SH=prebuild.sh
        ARG POSTBUILD_SH=postbuild.sh
        WORKDIR /tmp
        COPY . ./
        # Trigger prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH || true"
        ENV http_proxy http://xxx.xxx.xxx.xxx:xxx
        ENV https_proxy http://xxx.xxx.xxx.xxx:xxx
        # System packages
        RUN apt update && \
            apt install -y --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 \
                patch libbz2-dev llvm libncursesw5-dev xz-utils liblzma-dev m4 dos2unix libopenblas-dev libsqlite3-dev
        RUN wget https://www.python.org/ftp/python/3.10.5/Python-3.10.5.tgz
        RUN tar -zxvf Python-3.10.5.tgz && cd Python-3.10.5 && ./configure --prefix=/usr/local/python3.10.5 --enable-shared && make && make install
        RUN ln -s /usr/local/python3.10.5/bin/python3.10 /usr/local/python3.10.5/bin/python && \
            ln -s /usr/local/python3.10.5/bin/pip3.10 /usr/local/python3.10.5/bin/pip
        # Configure Python pip source
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        ENV LD_LIBRARY_PATH=/usr/local/python3.10.5/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.10.5/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.10.5/lib/python3.10/site-packages:$PYTHONPATH
        # Python packages
        RUN pip3 install decorator && \
            pip3 install sympy && \
            pip3 install cffi && \
            pip3 install pyyaml && \
            pip3 install pathlib2 && \
            pip3 install grpcio && \
            pip3 install grpcio-tools && \
            pip3 install protobuf && \
            pip3 install scipy && \
            pip3 install requests && \
            pip3 install attrs && \
            pip3 install Pillow==9.1.0 && \
            pip3 install torchvision==0.16.0 && \
            pip3 install numpy==1.23.5 && \
            pip3 install psutil && \
            pip3 install absl-py

        # Create the HwHiAiUser user and owner. Ensure the UID and GID are consistent with the physical machine to avoid files without an owner. In this example, the user and corresponding group are automatically created, with both UID and GID set to 1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
        # Ascend packages
        RUN bash $INSTALL_ASCEND_PKGS_SH
        RUN pip3 install $APEX_PKG
        RUN pip3 install $PYTORCH_PKG
        RUN pip3 install $PYTORCH_NPU_PKG
        RUN cd /tmp/dllogger-master/ && \
            python3 setup.py build && \
            python3 setup.py install
        # Environment variables
        ENV HCCL_WHITELIST_DISABLE=1
        ENV PYTHONPATH=/tmp/dllogger-master
        ENV http_proxy ""
        ENV https_proxy ""
        # Trigger postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH || true" && \
            rm $POSTBUILD_SH
        ```

### Building a Container Image Using a Dockerfile (MindSpore)<a name="ZH-CN_TOPIC_0000002511346627"></a>

**Prerequisites<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_section193545302315"></a>**

As shown in [Table 1](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_table13971125465512), obtain the software packages for the corresponding OS, as well as the Dockerfile and script files required for packaging the image.

In the package name, *{version}* indicates the version number, *{arch}* indicates the architecture, and *{chip_type}* indicates the chip type. The corresponding CANN packages in version 6.3.RC3, 6.2.RC3, and later include an installation prompt "Do you accept the EULA to install CANN (Y/N)"; in the Dockerfile writing examples, the installation command includes the "--quiet" parameter to accept the EULA by default, which users can modify as needed.

>[!NOTE]
>The MindSpore package and software used in Atlas training series products must meet the corresponding compatibility requirements. See [MindSpore Installation Guide](https://www.mindspore.cn/install/en) for details.

**Table 1**  Required software

<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_table13971125465512"></a>

|Software Package|Description|Obtaining Method|
|--|--|--|
|Ascend-cann-toolkit_<i>{version}</i>_linux-<i>{arch}</i>.run|CANN Toolkit.|<p>[Download Link](https://www.hiascend.com/developer/download/community/result?module=cann)</p><p>A version earlier than CANN 8.5.0 is required.</p>|
|Ascend-cann-kernels<em>-{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run|CANN operator package.|<p>[Download Link](https://www.hiascend.com/developer/download/community/result?module=cann)</p><p>A version earlier than CANN 8.5.0 is required.</p>|
|mindspore-<em>{version}</em>-cp3<em>x</em>-cp3<em>x</em>-linux_<em>{arch}</em>.whl|<p>MindSpore framework whl package.</p><p>Currently supports Python 3.9 to 3.11. The x in the package name indicates 9, 10, or 11. Select the corresponding software package based on the actual situation.</p><p>For versions earlier than MindSpore 2.0.0, the package name is changed from mindspore to mindspore-ascend.</p>|[Download Link](https://www.mindspore.cn/install)|
|Dockerfile|Required for creating an image.|Refer to [Dockerfile writing example](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li104026527188).|
|ascend_install.info|Driver installation information file.|Copy the "/etc/ascend_install.info" file from the host.|
|version.info|Driver version information file.|Copy the "/usr/local/Ascend/driver/version.info" file from the host.|
|prebuild.sh|Performs preparation for training runtime environment installation, such as configuring proxies.|Refer to [Step 3](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li206652021677).|
|install_ascend_pkgs.sh|Ascend software package installation script.|Refer to [Step 4](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li538351517716).|
|postbuild.sh|Clears installation packages, scripts, proxy configurations, and other items that do not need to be retained in the container.|Refer to [Step 5](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li154641047879).|

To prevent software packages from being maliciously tampered with during transmission or storage, you need to download the corresponding digital signature file for integrity verification when downloading the software package.

After downloading the software package, refer to *[OpenPGP Signature Verification Guide](https://support.huawei.com/enterprise/en/doc/EDOC1100209376)* to perform PGP digital signature verification on the software package downloaded from the Support website. If the verification fails, do not use the software package and contact Huawei technical support engineers for resolution first.

Before installing or upgrading using a software package, you must also verify the digital signature of the software package as described above to ensure that it has not been tampered with.

For carrier customers, please visit [https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction).

For enterprise customers, please visit [https://support.huawei.com/enterprise/en/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/en/tool/pgp-verify-TL1000000054)

>[!NOTE]
>
>- This section uses Ubuntu 18.04 with Python 3.9 as an example to describe the detailed process of building a container image using a Dockerfile. You need to modify the relevant steps based on the actual situation.
>- If you are using MindSpore 2.0.3 or later, you need to use ubuntu:20.04.

**Procedure<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_section38151530134817"></a>**

1. Upload the prepared software packages, deep learning framework, host-side driver installation information file, and driver version information file to the same directory on the server (such as `/home/test`).
    - Ascend-cann-toolkit\__\{version\}_\_linux-_\{arch\}_.run
    - Ascend-cann-kernels-_\{chip\_type\}_\__\{version\}_\_linux-_\{arch\}_.run
    - mindspore-_\{version\}_-cp3x-cp3x-linux\__\{arch\}_.whl
    - ascend\_install.info
    - version.info

2. Log in to the server as the `root` user.
3. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li206652021677"></a>Perform the following steps to prepare the `prebuild.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `prebuild.sh` file.

        ```shell
        vi prebuild.sh
        ```

    2. For the content to write, see the [prebuild.sh](#zh-cn_topic_0000001497124729_li146241711142818) writing example. After writing, run the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

4. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li538351517716"></a>Perform the following steps to prepare the `install_ascend_pkgs.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `install_ascend_pkgs.sh` file.

        ```shell
        vi install_ascend_pkgs.sh
        ```

    2. For the content to write, see the [install_ascend_pkgs.sh](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li58501140151720) writing example. After writing, run the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

5. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li154641047879"></a>Perform the following steps to prepare the `postbuild.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `postbuild.sh` file.

        ```shell
        vi postbuild.sh
        ```

    2. For the content to write, refer to the [postbuild.sh](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li14267051141712) writing example. After writing, execute the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

6. Perform the following steps to prepare a Dockerfile.
    1. Go to the directory where the software package is located and run the following command to create a Dockerfile (file name example: "Dockerfile").

        ```shell
        vi Dockerfile
        ```

    2. For the content to write, refer to the [Dockerfile](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li104026527188) writing example. After writing, execute the `:wq` command to save the content. The content uses the Ubuntu OS as an example.

7. Go to the directory where the software packages are located and run the following command to build the container image. **Note: Do not omit the dot (".") at the end of the command**.

    ```shell
    docker build -t image_name_system_architecture:image_tag .
    ```

In the above command, the parameters are described in the following table.

**Table 2** Command parameter description

<a name="table1021203815279"></a>

    |Name|Description|
    |--|--|
    |-t|Image name|
    |image_name_system_architecture:image_tag|Image name and tag. Use the actual values.|
For example:

    ```shell
    docker build -t test_train_arm64:v1.0 .
    ```

    When "Successfully built xxx" appears, it indicates that the image has been built successfully.

1. After the build is complete, run the following command to view the image information.

    ```shell
    docker images
    ```

    The response example is as follows:

    ```ColdFusion
    REPOSITORY                TAG                 IMAGE ID            CREATED             SIZE
    test_train_arm64          v1.0                d82746acd7f0        27 minutes ago      749MB
    ```

2. (Optional) Verify that the base image is available.
    1. Use Ascend Docker Runtime to mount the driver in the base image, using the base image `test_train_arm64:v1.0` as an example.

        ```shell
        docker run -it --privileged -e ASCEND_VISIBLE_DEVICES=0 test_train_arm64:v1.0 /bin/bash
        ```

    2. Check whether the MindSpore software in the base image is installed successfully.

        ```shell
        python -c "import mindspore;mindspore.set_context(device_target='Ascend');mindspore.run_check()"
        ```

        The response example is as follows, indicating that MindSpore is installed successfully.

        ```ColdFusion
        MindSpore version: Version number
        The result of multiplication calculation is correct, MindSpore has been installed on platform [Ascend] successfully!
        ```

**Writing Examples<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_section3523631151714"></a>**

1. <a name="zh-cn_topic_0000001497124729_li146241711142818"></a>`prebuild.sh`
    - Ubuntu Arm:

        ```shell
        #!/bin/bash
        #--------------------------------------------------------------------------------
        # # Write script code here using bash syntax to for installation preparations, such as configuring proxies
        # # This script will be executed before the formal build process starts
        #
        # # Note: This script will not be automatically cleaned up after execution. If it does not need to be retained in the image, please clean it up in the postbuild.sh script
        #--------------------------------------------------------------------------------
        # # DNS settings
        tee /etc/resolv.conf <<- EOF
        nameserver xxx.xxx.xxx.xxx  ## DNS server IP, multiple can be filled in, configure according to actual conditions
        nameserver xxx.xxx.xxx.xxx
        nameserver xxx.xxx.xxx.xxx
        EOF
        # apt proxy settings
        tee /etc/apt/apt.conf.d/80proxy <<- EOF
        Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";  #HTTP proxy server IP address and port.
        Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";  #HTTPS proxy server IP address and port.
        EOF
        chmod 777 -R /tmp
        rm /var/lib/apt/lists/*
        #apt source settings (using Ubuntu 18.04 ARM source as an example, configure according to actual setup)
        tee /etc/apt/sources.list <<- EOF
        deb http://mirrors.aliyun.com/ubuntu-ports/ bionic main restricted universe multiverse
        deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic main restricted universe multiverse
        deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-security main restricted universe multiverse
        deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-security main restricted universe multiverse
        deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-updates main restricted universe multiverse
        deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-updates main restricted universe multiverse
        deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-proposed main restricted universe multiverse
        deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-proposed main restricted universe multiverse
        deb http://mirrors.aliyun.com/ubuntu-ports/ bionic-backports main restricted universe multiverse
        deb-src http://mirrors.aliyun.com/ubuntu-ports/ bionic-backports main restricted universe multiverse
        EOF
        ```

    - Ubuntu x86_64:

        ```shell
        #!/bin/bash
        #--------------------------------------------------------------------------------

        # # Write script code here using bash syntax for installation preparations, such as configuring proxies
        # # This script will be executed before the formal build process starts
        #
        # # Note: This script will not be automatically removed after execution. If it does not need to be retained in the image, please remove it in the postbuild.sh script
        #--------------------------------------------------------------------------------
        # # apt proxy settings
        tee /etc/apt/apt.conf.d/80proxy <<- EOF
        Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";    ## HTTP proxy server IP address and port
        Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";   #HTTPS proxy server IP address and port
        EOF

        #apt source configuration (using Ubuntu 18.04 x86_64 source as an example, configure according to actual setup)
        tee /etc/apt/sources.list <<- EOF
        deb http://mirrors.ustc.edu.cn/ubuntu/ bionic main multiverse restricted universe
        deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-backports main multiverse restricted universe
        deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-proposed main multiverse restricted universe
        deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-security main multiverse restricted universe
        deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-updates main multiverse restricted universe
        deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic main multiverse restricted universe
        deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-backports main multiverse restricted universe
        deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-proposed main multiverse restricted universe
        deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-security main multiverse restricted universe
        deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-updates main multiverse restricted universe
        EOF
        ```

2. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li58501140151720"></a>`install_ascend_pkgs.sh`

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # Use bash syntax to write script code here to install Ascend software packages
    #
    # Note: This script will not be automatically removed after execution. If it does not need to be retained in the image, please clean it up in the postbuild.sh script
    #--------------------------------------------------------------------------------
    # Before building, copy /etc/ascend_install.info from the host to the current directory
    cp ascend_install.info /etc/
    mkdir -p /usr/local/Ascend/driver/
    cp version.info /usr/local/Ascend/driver/

    # Ascend-cann-toolkit_{version}_linux-{arch}.run
    chmod +x Ascend-cann-toolkit_{version}_linux-{arch}.run
    ./Ascend-cann-toolkit_{version}_linux-{arch}.run --install-path=/usr/local/Ascend/ --install --quiet
    chmod +x Ascend-cann-kernels-{chip_type}_{version}_linux-{arch}.run
    ./Ascend-cann-kernels-{chip_type}_{version}_linux-{arch}.run --install --quiet

    # Only install the toolkit package, which needs to be cleaned up. It will be mounted via ascend docker when the container starts
    rm -f version.info
    rm -rf /usr/local/Ascend/driver/
    ```

3. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li14267051141712"></a>`postbuild.sh`

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # Please write script code here using bash syntax to clean up installation packages, scripts, proxy configurations, etc. that do not need to be retained in the container
    # This script will be executed after the formal build process is completed
    #
    # Note: This script will be automatically cleared after execution and will not remain in the image. The script location and Working Dir location are /root.
    #--------------------------------------------------------------------------------

    rm -f ascend_install.info
    rm -f prebuild.sh
    rm -f install_ascend_pkgs.sh
    rm -f Dockerfile
    rm -f version.info
    rm -f Ascend-cann-toolkit_{version}_linux-{arch}.run
    rm -f Ascend-cann-kernels-{chip_type}_{version}_linux-{arch}.run
    # Select the packages to delete based on the actual installed version.
    rm -f mindspore-{version}-cp3x-cp3x-linux_{arch}.whl
    rm -f /etc/apt/apt.conf.d/80proxy

    tee /etc/resolv.conf <<- EOF
    # This file is managed by man:systemd-resolved(8). Do not edit.
    #
    # This is a dynamic resolv.conf file for connecting local clients to the
    # internal DNS stub resolver of systemd-resolved. This file lists all
    # configured search domains.
    #
    # Run "systemd-resolve --status" to see details about the uplink DNS servers
    # currently in use.
    #
    # Third party programs must not access this file directly, but only through the
    # symlink at /etc/resolv.conf. To manage man:resolv.conf(5) in a different way,
    # replace this symlink by a static file or a different symlink.
    #
    # See man:systemd-resolved.service(8) for details about the supported modes of
    # operation for /etc/resolv.conf.

    options edns0

    nameserver xxx.xxx.xxx.xxx
    nameserver xxx.xxx.xxx.xxx
    EOF
    ```

4. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li104026527188"></a>Dockerfile
    - Ubuntu Arm system with Python 3.9:

        ```Dockerfile
        FROM ubuntu:18.04

        ARG HOST_ASCEND_BASE=/usr/local/Ascend
        ARG INSTALL_ASCEND_PKGS_SH=install_ascend_pkgs.sh
        ARG TOOLKIT_PATH=/usr/local/Ascend/ascend-toolkit/latest
        ARG MINDSPORE_PKG=mindspore-{version}-cp39-cp39-linux_aarch64.whl
        ARG PREBUILD_SH=prebuild.sh
        ARG POSTBUILD_SH=postbuild.sh
        WORKDIR /tmp
        COPY . ./

        # Trigger prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH"

        ENV http_proxy http://xxx
        ENV https_proxy http://xxx


        # # system packages
        RUN apt update && \
            apt install --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 -y

        RUN wget https://www.python.org/ftp/python/3.9.2/Python-3.9.2.tgz
        RUN tar -zxvf Python-3.9.2.tgz && cd Python-3.9.2 && ./configure --prefix=/usr/local/python3.9.2 --enable-shared && make && make install

        RUN ln -s /usr/local/python3.9.2/bin/python3.9 /usr/local/python3.9.2/bin/python && \
            ln -s /usr/local/python3.9.2/bin/pip3.9 /usr/local/python3.9.2/bin/pip


        # # configure Python pip source
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        # # users need to modify the PYTHONPATH path according to the actual situation
        ENV LD_LIBRARY_PATH=/usr/local/python3.9.2/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.9.2/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.9.2/lib/python3.9/site-packages:$PYTHONPATH
        # # create HwHiAiUser user and owner. UID and GID must be consistent with the physical machine to avoid ownerless files. The example will automatically create the user and corresponding group, with both UID and GID set to 1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser

        # # install Python 3.9. If installing other versions, modify the following commands according to the actual situation
        RUN pip install numpy && \
            pip install decorator && \
            pip install sympy==1.4 && \
            pip install cffi==1.12.3 && \
            pip install pyyaml && \
            pip install pathlib2 && \
            pip install grpcio && \
            pip install grpcio-tools && \
            pip install protobuf && \
            pip install scipy && \
            pip install requests && \
            pip install kubernetes && \
            pip install attrs && \
            pip install psutil && \
            pip install absl-py

        # Ascend packages
        RUN umask 0022 && bash $INSTALL_ASCEND_PKGS_SH

        # MindSpore installation
        RUN pip install $MINDSPORE_PKG

        ENV http_proxy ""
        ENV https_proxy ""

        # Trigger postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH" && \
            rm $POSTBUILD_SH
        ```

    - Ubuntu x86_64 with Python 3.9:

        ```Dockerfile
        FROM ubuntu:18.04

        ARG HOST_ASCEND_BASE=/usr/local/Ascend
        ARG INSTALL_ASCEND_PKGS_SH=install_ascend_pkgs.sh
        ARG TOOLKIT_PATH=/usr/local/Ascend/ascend-toolkit/latest
        ARG MINDSPORE_PKG=mindspore-{version}-cp39-cp39-linux_x86_64.whl
        ARG PREBUILD_SH=prebuild.sh
        ARG POSTBUILD_SH=postbuild.sh
        WORKDIR /tmp
        COPY . ./

        # Trigger prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH"

        ENV http_proxy http://xxx
        ENV https_proxy http://xxx


        # # System packages
        RUN apt update && \
            apt install --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 -y

        RUN wget https://www.python.org/ftp/python/3.9.2/Python-3.9.2.tgz
        RUN tar -zxvf Python-3.9.2.tgz && cd Python-3.9.2 && ./configure --prefix=/usr/local/python3.9.2 --enable-shared && make && make install

        RUN ln -s /usr/local/python3.9.2/bin/python3.9 /usr/local/python3.9.2/bin/python && \
            ln -s /usr/local/python3.9.2/bin/pip3.9 /usr/local/python3.9.2/bin/pip

        # # Configure Python pip source
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        # # Modify the PYTHONPATH path according to the actual situation
        ENV LD_LIBRARY_PATH=/usr/local/python3.9.2/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.9.2/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.9.2/lib/python3.9/site-packages:$PYTHONPATH
        # # Create HwHiAiUser user and owner. UID and GID should be consistent with the physical machine to avoid ownerless files. In this example, the user and corresponding group are automatically created, with both UID and GID set to 1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser

        # # Install Python 3.9. If installing other versions, modify the following commands according to the actual situation
        RUN pip install numpy && \
            pip install decorator && \
            pip install sympy==1.4 && \
            pip install cffi==1.12.3 && \
            pip install pyyaml && \
            pip install pathlib2 && \
            pip install grpcio && \
            pip install grpcio-tools && \
            pip install protobuf && \
            pip install scipy && \
            pip install requests && \
            pip install kubernetes && \
            pip install attrs && \
            pip install psutil && \
            pip install absl-py

        # Ascend packages
        RUN umask 0022 && bash $INSTALL_ASCEND_PKGS_SH

        # MindSpore installation
        RUN pip install $MINDSPORE_PKG

        ENV http_proxy ""
        ENV https_proxy ""

        # Trigger postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH" && \
            rm $POSTBUILD_SH
        ```

### Building an inference image using a Dockerfile<a name="ZH-CN_TOPIC_0000002479386680"></a>

**Prerequisites<a name="zh-cn_topic_0000001497364777_section193545302315"></a>**

Obtain the software packages for the corresponding operating system, as well as the Dockerfile and script files required for building the image, as shown in [Table 1](#zh-cn_topic_0000001497364777_table13971125465512).

In the package name, `{version}` indicates the version number, `{arch}` indicates the architecture, and `{chip_type}` indicates the chip type. For the matching CANN software packages at version 6.3.RC3, 6.2.RC3, and later, an installation prompt "Do you accept the EULA to install CANN (Y/N)" has been added. In the Dockerfile writing examples, the installation command includes the `--quiet` parameter to accept the EULA by default, which users can modify as needed.

**Table 1** Required software

<a name="zh-cn_topic_0000001497364777_table13971125465512"></a>

|Software Package|Description|Obtaining Method|
|--|--|--|
|Ascend-cann-toolkit_<i>{version}</i>_linux-<i>{arch}</i>.run|CANN Toolkit|[Download Link](https://www.hiascend.com/developer/download/community/result?module=cann)|
|Ascend-cann-<em>{chip_type}</em>-ops_<em>{version}</em>_linux-<em>{arch}</em>.run|<p>CANN operator package.</p><p>Before CANN 8.5.0, this package was named Ascend-cann-kernels-<em>{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run</p>|[Download Link](https://www.hiascend.com/developer/download/community/result?module=cann)|
|Dockerfile|Required for creating images.|Refer to [Dockerfile Writing Example](#zh-cn_topic_0000001497364777_li166241028113511).|
|install.sh|Script for installing the inference service.|For creating inference models, refer to [ResNet50 Inference Guide](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/ACL_PyTorch/built-in/cv/Resnet50_Pytorch_Infer).|
|<em>XXX</em>.tar|Name of the inference service code package, prepared by the user based on the inference service. This chapter uses dvpp_resnet.tar as an example.|For creating inference models, refer to [ResNet50 Inference Guide](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/ACL_PyTorch/built-in/cv/Resnet50_Pytorch_Infer).|
|run.sh|Script for starting the inference service.|For creating inference models, refer to [ResNet50 Inference Guide](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/ACL_PyTorch/built-in/cv/Resnet50_Pytorch_Infer).|

> [!NOTE]
> You need to prepare other software packages and code required for inference on your own.

To prevent software packages from being maliciously tampered with during transmission or storage, you need to download the corresponding digital signature file for integrity verification when downloading software packages.

After downloading the software package, refer to the *[OpenPGP Signature Verification Guide](https://support.huawei.com/enterprise/en/doc/EDOC1100209376)* to perform PGP digital signature verification on the software package downloaded from the Support website. If the verification fails, do not use the software package and contact Huawei technical support engineers for resolution first.

Before using a software package for installation or upgrade, you also need to verify the digital signature of the software package following the above process to ensure that the software package has not been tampered with.

For carrier customers, please visit [https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction).

For enterprise customers, please visit [https://support.huawei.com/enterprise/en/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/en/tool/pgp-verify-TL1000000054).

This section uses Ubuntu x86_64 as an example. The code in the following steps is sample code. You can customize it based on the examples, and it is recommended that you perform security hardening on the sample code and images. Refer to [Container Security Hardening](./security_hardening.md#container-security-hardening).

**Procedure<a name="zh-cn_topic_0000001497364777_section9307172524312"></a>**

1. Upload the prepared software packages and files to the same directory on the server (for example, `/home/infer`).
    - Ascend-cann-toolkit\__\{version\}_\_linux-_\{arch\}_.run
    - Ascend-cann-_{chip_type}_-ops__{version}__linux-_{arch}_.run
    - Dockerfile
    - install.sh
    - run.sh
    - _XXX_.tar (self-prepared inference code or script)

2. Log in to the server as the `root` user.
3. Perform the following steps to prepare the `install.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `install.sh` file.

        ```shell
        vi install.sh
        ```

    2. Refer to the [install.sh](#zh-cn_topic_0000001497364777_li18749540133416) writing example and write the file based on your actual service requirements. After writing, run the `:wq` command to save the content.

4. Perform the following steps to prepare the `run.sh` file.
    1. Go to the directory where the software packages are located and run the following command to create the `run.sh` file.

        ```shell
        vi run.sh
        ```

    2. Refer to the [run.sh](#zh-cn_topic_0000001497364777_li18234181353511) example and write the script based on your actual service requirements. After writing, run the `:wq` command to save the content.

5. Perform the following steps to prepare a Dockerfile.
    1. Go to the directory where the software package is located and run the following command to create a Dockerfile (example filename: "Dockerfile").

        ```shell
        vi Dockerfile
        ```

    2. Refer to the [Dockerfile](#zh-cn_topic_0000001497364777_li166241028113511) example and write the file based on your actual service requirements. After writing, run the `:wq` command to save the content.

6. Go to the directory where the software packages are located and run the following command to build the container image. **Be careful not to omit the dot (".") at the end of the command**.

    ```shell
    docker build --build-arg TOOLKIT_VERSION={version} --build-arg TOOLKIT_ARCH={arch} --build-arg DIST_PKG=XXX.tar -t image_name_system_architecture:image_tag .
    ```

    In the above commands, the description of each parameter is shown in the following table.

    **Table 2** Command parameters

    <a name="zh-cn_topic_0000001497364777_zh-cn_topic_0256378845_table47051919193111"></a>

    |Name|Description|
    |--|--|
    |--build-arg|Parameters within the Dockerfile.|
    |<em>{version}</em>|Toolkit package version number. Enter the actual value.|
    |<em>{arch}</em>|Toolkit package architecture. Enter the actual value based on your situation.|
    |<em>XXX</em>.tar|Name of the inference service code package. Enter the actual value based on your situation.|
    |-t|Image name.|
    |<em>Image name</em><em>_system architecture:</em><em>Image tag</em>|Image name and tag. Enter the actual values based on your situation.|

    An example is as follows:

    ```shell
    docker build --build-arg TOOLKIT_VERSION=20.1.rc3 --build-arg TOOLKIT_ARCH=x86_64 --build-arg DIST_PKG=dvpp_resnet.tar -t ubuntu-infer:v1 .
    ```

    "Successfully built xxx" indicates that the image has been built successfully.

7. After the build is complete, run the following command to view the image information.

    ```shell
    docker images
    ```

    The response example is as follows:

    ```ColdFusion
    REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
    ubuntu-infer        v1                  fffbd83be42a        2 minutes ago       293MB
    ```

**Writing Examples <a name="zh-cn_topic_0000001497364777_section158942057133318"></a>**

1. <a name="zh-cn_topic_0000001497364777_li18749540133416"></a>`install.sh`

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # Install the inference service script. This example uses the inference service package dvpp_resnet.tar for illustration. You can modify the service package name as needed.
    #-------------------------------------
    tar -xvf dvpp_resnet.tar
    # It is also recommended to modify the permissions and owner of the extracted files.
    ```

2. <a name="zh-cn_topic_0000001497364777_li18234181353511"></a>`run.sh writing`

    ```shell
    #!/bin/bash
    # Run service code
    cd /home/out
    numbers=`ls /dev/| grep davinci | grep -v davinci_manager | wc -l`
    # Update logs every 5 minutes
    #./main $numbers|grep -nE '.*\[.*[[:digit:]]{2}:[[:digit:]]{1}[05]:00\]' >./log.txt
    # Load offline inference environment variables
    export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64:${LD_LIBRARY_PATH}
    source /usr/local/Ascend/cann/set_env.sh
    ./main $numbers
    ```

    >[!NOTE]
    >The driver-related paths are configured in `LD_LIBRARY_PATH`, and the files within them will be used when executing inference jobs. It is recommended that the running user for inference jobs be consistent with the running user specified during driver installation to avoid privilege escalation risks caused by user mismatch.

3. <a name="zh-cn_topic_0000001497364777_li166241028113511"></a>Dockerfile ( Please customize and modify it according to the actual situation.)

    ```Dockerfile
    #The base image ubuntu:18.04 does not include the Toolkit package. You can refer to some steps in the Dockerfile example for installation, and you need to prepare the Toolkit package in advance.
    #It is recommended to pull the inference base image from the Ascend image repository, which already has the Toolkit package installed. Also, confirm whether the Toolkit package matches the driver version on the physical machine.
    FROM ubuntu:18.04

    # Set Toolkit and OPS package parameters
    ARG TOOLKIT_VERSION
    ARG TOOLKIT_ARCH
    ARG TOOLKIT_PKG=Ascend-cann-toolkit_{version}_linux-{arch}.run
    ARG OPS_PKG=Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run


    # Set environment variables
    ARG ASCEND_BASE=/usr/local/Ascend

    # Set the working directory for the container after startup
    WORKDIR /home

    ## Copy the Toolkit package and OPS package
    COPY $TOOLKIT_PKG .
    COPY $OPS_PKG .

    # # Install the Toolkit package and OPS package
    RUN umask 0022 && \
        groupadd xxx (User-defined; be consistent with that specified during driver installation) && \
        useradd -g xxx(User-defined; be consistent with that specified during driver installation)） -s /usr/sbin/nologin (user login disabling; Ubuntu as an example) -m -d /home/xxx xxx(User-defined; be consistent with that specified during driver installation) && \
        chmod +x ${TOOLKIT_PKG} &&\
        ./${TOOLKIT_PKG} --quiet --install --install-for-all --whitelist=nnrt --force &&\
        rm ${TOOLKIT_PKG}
        chmod +x ${OPS_PKG} &&\
        ./${OPS_PKG} --install --install-for-all --quiet --force &&\
        rm ${OPS_PKG}

    # # Copy the compressed package of the service inference program, installation script, and running script
    ARG DIST_PKG
    COPY $DIST_PKG .
    COPY install.sh .
    COPY run.sh .

    # # Run the installation script
    RUN mkdir -p /usr/slog && \
        mkdir -p /var/log/npu/slog/slogd && \
        chmod u+x run.sh install.sh && \
        sh install.sh && \
        rm $DIST_PKG && \
        rm install.sh

    CMD bash run.sh
    ```

    >[!NOTE]
    >For CANN package versions 6.2.RC1, 6.3.RC1, and later, the `--force` parameter is added when installing the package. This parameter has already been included in the Dockerfile example above. If you are using a package version earlier than 6.2.RC1 or 6.3.RC1, you need to remove this parameter from the Dockerfile example.

## Quering Information About Currently Available Devices in the Cluster<a name="ZH-CN_TOPIC_0000002516255287"></a>

1. Query the ConfigMap.

    ```shell
    kubectl get cm -A | grep cluster-info
    ```

    The response example is as follows:

    ```ColdFusion
    kube-public            cluster-info                                           1      19d
    mindx-dl               cluster-info-device-0                                  1      19h
    mindx-dl               cluster-info-node-cm                                   1      19h
    mindx-dl               cluster-info-switch-0                                  1      19h
    ```

2. Query the detailed information of the ConfigMap to obtain available device information. The following uses the node name `localhost.localdomain` as an example.

    1. Query the detailed information of the device-related ConfigMap to obtain the available chip information of the node.

        ```shell
        kubectl describe cm -n mindx-dl cluster-info-device-0
        ```

        The response example is as follows:

        ```ColdFusion
        Name:         cluster-info-device-0
        Namespace:    mindx-dl
        Labels:       mx-consumer-volcano=true
        Annotations:  <none>
        Data
        ====
        cluster-info-device-0:
        ----
        {"mindx-dl-deviceinfo-localhost.localdomain":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[{\"fault_type\":\"PublicFault\",\"npu_name\":\"Ascend910-0\",\"large_model_fault_level\":\"SeparateNPU\",\"fault_level\":\"SeparateNPU\",\"fault_handling\":\"SeparateNPU\",\"fault_code\":\"220001001\",\"fault_time_and_level_map\":{\"220001001\":{\"fault_time\":1736926605,\"fault_level\":\"SeparateNPU\"}}},{\"fault_type\":\"PublicFault\",\"npu_name\":\"Ascend910-1\",\"large_model_fault_level\":\"SeparateNPU\",\"fault_level\":\"SeparateNPU\",\"fault_handling\":\"SeparateNPU\",\"fault_code\":\"220001001\",\"fault_time_and_level_map\":{\"220001001\":{\"fault_time\":1736926605,\"fault_level\":\"SeparateNPU\"}}},{\"fault_type\":\"PublicFault\",\"npu_name\":\"Ascend910-2\",\"large_model_fault_level\":\"SeparateNPU\",\"fault_level\":\"SeparateNPU\",\"fault_handling\":\"SeparateNPU\",\"fault_code\":\"220001001\",\"fault_time_and_level_map\":{\"220001001\":{\"fault_time\":1736926605,\"fault_level\":\"SeparateNPU\"}}}]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Recovering":"","huawei.com/Ascend910-Unhealthy":"Ascend910-0,Ascend910-1,Ascend910-2"},"UpdateTime":1759214666,"CmName":"mindx-dl-deviceinfo-localhost.localdomain","SuperPodID":-2,"ServerIndex":-2},"mindx-dl-deviceinfo-node173":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Recovering":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1759202968,"CmName":"mindx-dl-deviceinfo-node173","SuperPodID":-2,"ServerIndex":-2}}
        Events:  <none>
        ```

        From the above response information, you can see that the available chips for this node are `Ascend910-3`, `Ascend910-4`, `Ascend910-5`, `Ascend910-6`, and `Ascend910-7`.

    2. Query the detailed information of the node-related ConfigMap to obtain node status information.

        ```shell
        kubectl describe cm -n mindx-dl cluster-info-node-cm
        ```

        The response example is as follows:

        ```ColdFusion
        Name:         cluster-info-node-cm
        Namespace:    mindx-dl
        Labels:       mx-consumer-volcano=true
        Annotations:  <none>

        Data
        ====
        cluster-info-node-cm:
        ----
        {"mindx-dl-nodeinfo- localhost.localdomain":{"FaultDevList":[{"DeviceType":"PSU","DeviceId":4,"FaultCode":["0300000D"],"FaultLevel":"NotHandleFault"}],"NodeStatus":"Healthy","CmName":"mindx-dl-nodeinfo-localhost.localdomain "}}

        BinaryData
        ====

        Events:  <none>
        ```

        From the above response information, you can see that `NodeStatus` of this node is `Healthy`, indicating that the current node health is normal.

    3. Query the detailed information of the switch-related ConfigMap to obtain node status information.

        ```shell
        kubectl describe cm -n mindx-dl cluster-info-switch-0
        ```

        The response example is as follows:

        ```ColdFusion
        Name:         cluster-info-switch-0
        Namespace:    mindx-dl
        Labels:       mx-consumer-volcano=true
        Annotations:  <none>

        Data
        ====
        cluster-info-switch-0:
        ----
        {"mindx-dl-switchinfo-localhost.localdomain ":{"FaultCode":[],"FaultLevel":"","UpdateTime":1763544679,"NodeStatus":"Healthy","FaultTimeAndLevelMap":{},"CmName":"mindx-dl-switchinfo-localhost.localdomain "}}

        BinaryData
        ====

        Events:  <none>
        ```

        From the above response information, you can see that `NodeStatus` of this node is `Healthy`, indicating that the current node is healthy.

    Based on the above query results, the available chips for this node are `Ascend910-3`, `Ascend910-4`, `Ascend910-5`, `Ascend910-6`, and `Ascend910-7`.

    If `NodeStatus` in the response information of step 2 or step 3 is `UnHealthy`, it indicates that all devices on the current node are unavailable. Combined with the query results from step 1, the available chips for this node are empty.

    >[!NOTE]
    >When the cluster scale exceeds 1,000 nodes, the ConfigMaps corresponding to `cluster-info-device-` and `mindx-dl-switchinfo-` will be sharded. Each `cluster-info-device-` or `mindx-dl-switchinfo-` contains device information for a maximum of 1,000 nodes. In this scenario, you need to perform the query operations in step 1 and step 3 on all `cluster-info-device-` ConfigMaps to find the detailed information of the target node, so as to confirm the available chip information of that node.
