# 常用操作<a name="ZH-CN_TOPIC_0000002511346991"></a>

## 调度配置<a name="ZH-CN_TOPIC_0000002511427007"></a>

Volcano组件支持K8s原生调度，可以使用nodeAffinity进行调度。以下示例使用强制的节点亲和性进行调度，更多关于nodeAffinity字段的说明请参见[Kubernetes官方文档](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/)。

- Volcano Job的任务YAML中，需要添加如下加粗字段。

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
            <strong>affinity:      # 新增以下加粗字段</strong>
              <strong>nodeAffinity:                             # 节点亲和性配置</strong>
                <strong>requiredDuringSchedulingIgnoredDuringExecution:</strong>
                  <strong>nodeSelectorTerms:                    # 节点选择列表</strong>
                    <strong>- matchExpressions:</strong>
                        <strong>- key: aaa               # 匹配标签的key为aaa的节点，并且value是yyy的节点</strong>
                          <strong>operator: In</strong>
                          <strong>values:</strong>
                            <strong>- yyy</strong>
                 podAntiAffinity:
                   requiredDuringSchedulingIgnoredDuringExecution:
    ...
              nodeSelector:
                example-key: example-value    # 示例值，用户可根据调度意图自行配置nodeSelector
    ...</pre>

- Ascend Job的任务YAML中，需要添加如下加粗字段。

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
              <strong>affinity:   #  新增字段</strong>
                <strong>nodeAffinity:                           # 节点亲和性配置</strong>
                  <strong>requiredDuringSchedulingIgnoredDuringExecution:</strong>
                    <strong>nodeSelectorTerms:                  # 节点选择列表</strong>
                      <strong>- matchExpressions:</strong>
                        <strong>- key: aaa            # 匹配标签的key为aaa的节点，并且value是yyy的节点</strong>
                          <strong>operator: In</strong>
                          <strong>values:</strong>
                            <strong>- yyy</strong>
              nodeSelector:
                example-key: example-value    # 可选值，用户可根据实际需求配置nodeSelector
    ...</pre>

    >[!NOTE]
    >可通过执行**kubectl get  node --show-labels**命令，查询节点的标签。在LABELS字段下，等号前的值为标签的key值，等号后的值为标签的value值，如aaa=yyy。

## 安装NFS<a name="ZH-CN_TOPIC_0000002479227106"></a>

### Ubuntu操作系统<a name="ZH-CN_TOPIC_0000002479227110"></a>

NFS（Network File System）网络文件系统，它允许网络中的计算机之间共享资源。在集群调度场景下，需要依赖NFS环境实现训练任务或推理任务的正常运行。NFS可以安装在服务器端或者客户端，用户可以根据需要进行选择。

**在服务器端安装<a name="zh-cn_topic_0000001497364925_section119917347402"></a>**

1. 使用管理员账号登录存储节点，执行以下命令安装NFS服务端。

    ```shell
    apt install -y nfs-kernel-server
    ```

2. 根据实际情况固定NFS相关端口并配置相关端口的防火墙。
3. 执行以下命令，创建一个共享目录（如“/data/atlas\_dls”）并修改目录权限。

    ```shell
    mkdir -p /data/atlas_dls
    chmod 750 /data/atlas_dls/
    ```

4. 在“/etc/exports”文件末尾追加以下内容，根据需要配置允许的IP地址并加固相关权限设置。

    ```shell
    /data/atlas_dls 业务IP地址（配置必要的权限）
    ```

5. 执行以下命令，启动rpcbind。

    ```shell
    systemctl restart rpcbind.service
    systemctl enable rpcbind.service
    ```

6. 执行以下命令，查看rpcbind是否已启动。

    ```shell
    systemctl status rpcbind.service
    ```

    出现以下回显，说明服务正常。

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

7. rpcbind启动后，执行以下命令，启动NFS服务。

    ```shell
    systemctl restart nfs-server.service
    systemctl enable nfs-server.service
    ```

8. 执行以下命令，查看NFS服务是否已启动。

    ```shell
    systemctl status nfs-server.service
    ```

    出现以下回显，说明服务正常。若NFS服务启动失败，可以参见[df -h执行失败，NFS启动失败](https://gitcode.com/Ascend/mind-cluster/issues/353)章节进行处理。

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

9. 执行以下命令，查看共享目录（如“/data/atlas\_dls”）挂载权限。

    ```shell
    cat /var/lib/nfs/etab
    ```

    出现以下回显，说明服务正常。

    ```ColdFusion
    /data/atlas_dls *(rw,...会显示配置的对应权限)
    ```

**在客户端安装<a name="zh-cn_topic_0000001497364925_section10189114704512"></a>**

使用管理员账号登录其他服务器，执行以下命令安装NFS客户端。

```shell
apt install -y nfs-common
```

### CentOS操作系统<a name="ZH-CN_TOPIC_0000002511427005"></a>

NFS网络文件系统，它允许网络中的计算机之间共享资源。在集群调度场景下，需要依赖NFS环境实现训练任务或推理任务的正常运行。NFS可以安装在服务器端或者客户端，用户可以根据需要进行选择。

**在服务器端安装<a name="zh-cn_topic_0000001446805000_section1398218463486"></a>**

1. 使用管理员账号登录存储节点，执行以下命令安装NFS服务端。

    ```shell
    yum install nfs-utils -y
    ```

2. 根据实际情况固定NFS相关端口并配置相关端口的防火墙。
3. 执行以下命令，创建一个共享目录（如“/data/atlas\_dls”）并修改目录权限。

    ```shell
    mkdir -p /data/atlas_dls
    chmod 750 /data/atlas_dls/
    ```

4. 执行**vi /etc/exports**命令，在文件末尾追加以下内容，根据需要配置允许的IP地址并加固相关权限设置。

    ```shell
    /data/atlas_dls 业务IP地址（配置必要的权限）
    ```

5. 执行以下命令，启动rpcbind。

    ```shell
    systemctl restart rpcbind.service
    systemctl enable rpcbind.service
    ```

6. 执行以下命令，查看rpcbind是否已启动。

    ```shell
    systemctl status rpcbind.service
    ```

    出现以下回显，说明服务正常。

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

7. rpcbind启动后，执行以下命令，启动NFS服务。

    ```shell
    systemctl restart nfs-server.service
    systemctl enable nfs-server.service
    ```

8. 执行以下命令，查看NFS服务是否已启动。

    ```shell
    systemctl status nfs-server.service
    ```

    出现以下回显，说明服务正常。若NFS服务启动失败，可以参见[df -h执行失败，NFS启动失败](https://gitcode.com/Ascend/mind-cluster/issues/353)章节进行处理。

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

9. 执行以下命令，查看共享目录（如“/data/atlas\_dls”）挂载权限。

    ```shell
    cat /var/lib/nfs/etab
    ```

    出现以下回显，说明服务正常。

    ```ColdFusion
    /data/atlas_dls *(rw,...会显示配置的对应权限)
    ```

**在客户端安装<a name="zh-cn_topic_0000001446805000_section1862665118118"></a>**

1. 使用管理员账号登录其他服务器，执行以下命令安装NFS客户端。

    ```shell
    yum install nfs-utils -y
    ```

2. 执行以下命令，启动rpcbind。

    ```shell
    systemctl restart rpcbind.service
    systemctl enable rpcbind.service
    ```

3. 执行以下命令，查看rpcbind是否启动。

    ```shell
    systemctl status rpcbind.service
    ```

    出现以下回显，说明服务正常。

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

4. rpcbind启动后，执行以下命令，启动NFS服务。

    ```shell
    systemctl restart nfs-server.service
    systemctl enable nfs-server.service
    ```

5. 执行以下命令，查看NFS服务是否启动。

    ```shell
    systemctl status nfs-server.service
    ```

    出现以下回显，说明服务正常。

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

6. （可选）NFS需要使用mount和umount命令，一般情况下，系统自带mount命令。若当前客户端没有此命令，可执行以下步骤进行安装。

    ```shell
    yum install -y  util-linux
    ```

## 查询上报的故障信息<a name="ZH-CN_TOPIC_0000002479387090"></a>

### Volcano<a name="ZH-CN_TOPIC_0000002479387088"></a>

Volcano收集了内部的芯片故障、参数面网络故障和节点故障信息，将其作为对外的信息放在K8s的ConfigMap中，以供外部查询和使用。

查询命令为**kubectl describe cm -n volcano-system  vcjob-fault-npu-cm**，命令回显示例如下，**关键参数**说明请参见[表2 vcjob-fault-npu-cm字段说明](../06_api/01_volcano.md#任务信息)。

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

#### 故障信息<a name="ZH-CN_TOPIC_0000002479387086"></a>

Ascend Device Plugin收集了内部的芯片故障、参数面网络故障和节点故障，将其作为对外的信息放在了K8s的ConfigMap中，一个ConfigMap放置一个节点的信息，以供外部查询和使用。

查询命令：**kubectl describe cm -n kube-system  mindx-dl-deviceinfo-$**_\{node\_name\}_

以<term>Atlas A3 训练系列产品</term>为例，回显示例如下；不同设备的回显参数可能不同，以实际为准。关键参数说明请参见[表1 DeviceInfoCfg](../06_api/02_ascend_device_plugin.md#芯片资源)。

```ColdFusion
{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[{\"fault_type\":\"CardNetworkUnhealthy\",\"npu_name\":\"Ascend910-0\",\"large_model_fault_level\":\"PreSeparateNPU\",\"fault_level\":\"PreSeparateNPU\",\"fault_handling\":\"PreSeparateNPU\",\"fault_code\":\"81078603\",\"fault_time_and_level_map\":{\"81078603\":{\"fault_time\":1744168468259,\"fault_level\":\"PreSeparateNPU\"}}},{\"fault_type\":\"CardUnhealthy\",\"npu_name\":\"Ascend910-4\",\"large_model_fault_level\":\"SeparateNPU\",\"fault_level\":\"SeparateNPU\",\"fault_handling\":\"SeparateNPU\",\"fault_code\":\"A8028801,A4028801,80E18402,80E18401\",\"fault_time_and_level_map\":{\"80E18401\":{\"fault_time\":1744167455784,\"fault_level\":\"NotHandleFault\"},\"80E18402\":{\"fault_time\":1744167455784,\"fault_level\":\"SeparateNPU\"},\"A4028801\":{\"fault_time\":1744167455784,\"fault_level\":\"NotHandleFault\"},\"A8028801\":{\"fault_time\":1744167455784,\"fault_level\":\"SeparateNPU\"}}}]","huawei.com/Ascend910-NetworkUnhealthy":"Ascend910-0","huawei.com/Ascend910-Recovering":"","huawei.com/Ascend910-Unhealthy":"Ascend910-4"},"UpdateTime":1744182144},"SuperPodID":-2,"ServerIndex":-2,"CheckCode":"a550811fdfafb5717555526816af2ca4ac6c3e102f5907574048578e0c8fcc73"}
```

#### 故障事件信息<a name="ZH-CN_TOPIC_0000002511347039"></a>

Ascend Device Plugin收集到的故障事件可以通过K8s的event事件进行上报，查询命令为**kubectl get events -n kube-system**。以Atlas 训练系列产品为例，回显示例如下，参数说明请参见[表1](#table66076214393)。

```ColdFusion
NAMESPACE     LAST SEEN   TYPE      REASON     OBJECT                                         MESSAGE
kube-system   8s          Warning   Occur      pod/ascend-device-plugin-daemonset-910-dlpmv   device fault, nodeName:k8smaster, assertion:Occur, cardID:2, deviceID:0, faultCodes:8C084E00, faultLevelName:RestartBusiness, alarmRaisedTime:2023-11-21 05:36:53
```

**表 1**  参数说明

<a name="table66076214393"></a>

|参数名|描述|
|--|--|
|NAMESPACE|命名空间名称，取值为kube-system。|
|LAST SEEN|事件产生时间。|
|TYPE|<p>事件的类型，取值为<span>“Normal”</span>和<span>“Warning”</span>。</p>|
|REASON|<p>事件产生原因。取值说明如下：</p><ul><li>Occur：故障发生</li><li>Recovery：故障恢复</li><li>Notice：通知</li></ul>|
|OBJECT|<p>事件对象，取值规范为pod/<span><em>Ascend Device Plugin</em></span><em>的Pod名称</em>，如pod/ascend-device-plugin-daemonset-910-dlpmv。</p>|
|MESSAGE|<p>事件信息内容描述。事件内容的字段说明如下：</p><ul><li>nodeName：节点名称</li><li>assertion：信息类型<ul><li>Occur：故障发生</li><li>Recovery：故障恢复</li><li>Notice：通知</li></ul></li><li>cardID：NPU管理单元ID（NPU设备ID）</li><li>deviceID：设备编号</li><li>faultCodes：故障码，取值如8C084E00</li><li>faultLevelName：故障级别名称<ul><li>NotHandleFault：不做处理</li><li>RestartRequest：<span>影响业务执行，需要重新执行业务请求</span></li><li>RestartBusiness：<span>影响业务执行，</span>需要重新执行业务</li><li>FreeRestartNPU：影响业务执行，待芯片空闲时需复位芯片</li><li>RestartNPU：直接复位芯片并重新执行业务</li><li>SeparateNPU：隔离芯片</li><li>PreSeparateNPU：暂不影响业务，后续不再调度任务到该芯片</li><li>SubHealthFault：根据任务YAML中配置的subHealthyStrategy参数取值进行处理</li></ul></li><li>alarmRaisedTime：故障发生时间</li></ul>|

### ClusterD<a name="ZH-CN_TOPIC_0000002511347035"></a>

ClusterD收集了内部的节点故障、芯片故障和灵衢总线设备故障，将其作为对外的信息放在了K8s的ConfigMap中，以供外部查询和使用。

**节点故障<a name="section208771421687"></a>**

查询命令：**kubectl describe cm -n mindx-dl cluster-info-node-cm**

以<term>Atlas A3 训练系列产品</term>为例，回显示例如下；不同设备的回显参数可能不同，以实际为准。关键参数说明请参见[表1 cluster-info-node-cm](../06_api/04_clusterd/00_cluster_resources.md#configmap说明)。

```ColdFusion
{"mindx-dl-nodeinfo-kwok-node-0":{"FaultDevList":[],"NodeStatus":"Healthy","CmName":"mindx-dl-nodeinfo-kwok-node-0"},"mindx-dl-deviceinfo-kwok-node-1001":{"FaultDevList":[],"NodeStatus":"Healthy","CmName":"mindx-dl-nodeinfo-kwok-node-1001"}}
```

**芯片故障<a name="section834865016504"></a>**

查询命令：**kubectl describe cm -n mindx-dl cluster-info-device-$**_\{m\}_

m为从0开始递增的整数。集群规模每增加1000个节点，则会新增一个ConfigMap文件cluster-info-device-$\{m\}。

以<term>Atlas A3 训练系列产品</term>为例，回显示例如下；不同设备的回显参数可能不同，以实际为准，关键参数说明请参见[表2 cluster-info-device-$\{m\}](../06_api/04_clusterd/00_cluster_resources.md#configmap说明)。

```ColdFusion
{"mindx-dl-deviceinfo-kwok-node-0":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1693899390,"CmName":"mindx-dl-deviceinfo-kwok-node-0","SuperPodID":0,"ServerIndex":0},"mindx-dl-deviceinfo-kwok-node-1001":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1693899390,"CmName":"mindx-dl-deviceinfo-kwok-node-1001","SuperPodID":0,"ServerIndex":0}}
```

**灵衢总线设备故障<a name="section1728713587242"></a>**

查询命令：**kubectl describe cm -n mindx-dl cluster-info-switch-$**_\{m\}_

m为从0开始递增的整数。集群规模每增加2000个节点，则会新增一个ConfigMap文件cluster-info-switch-$\{m\}。

以<term>Atlas A3 训练系列产品</term>为例，回显示例如下；不同设备的回显参数可能不同，以实际为准。关键参数说明请参见[表1](#table9246232250)。

```ColdFusion
{"FaultCode":[000001c1],"FaultLevel":"NotHandle","UpdateTime":1722845555,"NodeStatus":"Healthy"}
```

**表 1**  灵衢总线设备故障参数说明

<a name="table9246232250"></a>

|参数|说明|
|--|--|
|FaultCode|故障码，由英文和数字拼接而成的字符串，字符串表示故障码的十六进制。|
|FaultLevel|<p>当前故障中等级最高的故障所对应的处理策略。</p><ul><li>NotHandle：不做处理。</li><li>SubHealth：根据配置策略决定如何处理。</li><li>Reset：隔离节点。</li><li>Separate：隔离节点。</li><li>RestartRequest：隔离节点。</li></ul>|
|UpdateTime|ConfigMap更新时间。|
|NodeStatus|<p>当前节点状态。</p><ul><li>Healthy：节点健康。</li><li>SubHealthy：节点预隔离，当前任务不做处理，后续任务不再调度该节点。</li><li>UnHealthy：节点不健康，隔离节点，进行任务重调度。</li></ul>|

### NodeD<a name="ZH-CN_TOPIC_0000002511427003"></a>

NodeD收集了节点故障信息和节点健康状态信息，将其作为对外的信息放在K8s的ConfigMap中，以供外部查询和使用。

查询命令为**kubectl describe cm mindx-dl-nodeinfo-**<i>\<nodename\></i> **-n mindx-dl**，命令回显示例如下，关键参数说明请参见[表1 mindx-dl-nodeinfo-_<nodename\>_](../06_api/03_noded.md#节点资源)。

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

## 制作镜像<a name="ZH-CN_TOPIC_0000002479227114"></a>

### 使用Dockerfile构建容器镜像（PyTorch）<a name="ZH-CN_TOPIC_0000002511426595"></a>

**前提条件<a name="zh-cn_topic_0000001497364957_section193545302315"></a>**

请按照[表1](#zh-cn_topic_0000001497364957_table13971125465512)所示，获取对应操作系统的软件包与打包镜像所需Dockerfile文件与脚本文件。

软件包名称中\{version\}表示版本号、\{arch\}表示架构、\{chip\_type\}表示芯片类型。配套的CANN软件包在6.3.RC3、6.2.RC3及以上版本增加了“您是否接受EULA来安装CANN（Y/N）”的安装提示；在Dockerfile编写示例中的安装命令包含“--quiet”参数的默认同意EULA，用户可自行修改。

**表 1**  所需软件

<a name="zh-cn_topic_0000001497364957_table13971125465512"></a>

|软件包|说明|获取方法|
|--|--|--|
|Ascend-cann-toolkit_<i>{version}</i>_linux-<i>{arch}</i>.run|CANN Toolkit开发套件包。|<p>[获取链接](https://www.hiascend.com/developer/download/community/result?module=cann)</p>|
|<p>Ascend-cann-<em>{chip_type}</em>-ops_<em>{version}</em>_linux-<em>{arch}</em>.run</p>|<p>CANN算子包。</p><p>CANN 8.5.0之前版本该包名为Ascend-cann-kernels-<em>{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run</p>|[获取链接](https://www.hiascend.com/developer/download/community/result?module=cann)|
|apex-0.1+ascend-cp3x-cp3x-linux_<em>{arch}</em>.whl|<p>混合精度模块。</p><p>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10。</p>|请参见<span>《Ascend Extension for PyTorch 软件安装指南》中的“[安装APEX模块](https://gitcode.com/Ascend/apex/blob/master/docs/zh/installing_apex.md)”章节</span>，根据实际情况编译APEX软件包。|
|<ul><li><span>x86_64</span>：torch-<em>v{version}</em>+cpu-cp3x-cp3x-linux_x86_64.whl</li><li><span>ARM</span>：torch-<em>v{version}</em>-cp3x-cp3x-manylinux_2_17_aarch64.manylinux2014_aarch64.whl</li></ul>|<p>官方<span>PyTorch</span>包。</p><p>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10。</p><p><em>{version}</em>表示<span>PyTorch</span>版本号，当前可支持<span>PyTorch</span> 2.1.0~2.7.1。</p>|<p>[获取链接](https://download.pytorch.org/whl/torch/)</p><p>请根据实际情况选择要安装的<span>PyTorch</span>版本。</p>|
|<p>torch_npu-<em>v{version}</em><em>.</em>post<em>{version}</em>-cp3x-cp3x-manylinux_2_17_<em>{arch}</em>.manylinux2014_<em>{arch}</em>.whl</p>|<p><span>Ascend Extension for PyTorch</span>插件。</p><p>软件包中的cp3x表示Python版本号，例如x为10表示Python 3.10。</p>|<p>[获取链接](https://www.hiascend.com/document/detail/zh/Pytorch/600/configandinstg/instg/insg_0001.html)</p><ul><li>请选择与<span>PyTorch</span>配套的TorchNPU版本。</li><li>如果使用MindSpeed-LLM代码仓上的<span>PyTorch</span>模型，需要使用<span>Ascend Extension for PyTorch</span> 2.1.0及以上版本。</li></ul>|
|Dockerfile|制作镜像需要。|参考[Dockerfile编写示例](#zh-cn_topic_0000001497364957_li104026527188)|
|dllogger-master|PyTorch日志工具。|[获取链接](https://github.com/NVIDIA/dllogger)|
|ascend_install.info|驱动安装信息文件。|从host拷贝“/etc/ascend_install.info”文件。|
|version.info|驱动版本信息文件。|从host拷贝“/usr/local/Ascend/driver/version.info”文件。|
|prebuild.sh|执行训练运行环境安装准备工作，例如配置代理等。|参考[步骤3](#zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li206652021677)|
|install_ascend_pkgs.sh|昇腾软件包安装脚本。|参考[步骤4](#zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li538351517716)|
|postbuild.sh|清除不需要保留在容器中的安装包、脚本、代理配置等。|参考[步骤5](#zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li154641047879)|

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE]
>本章节以Ubuntu操作系统，配套Python  3.10，CANN 8.5.0为例来介绍使用Dockerfile构建容器镜像的详细过程，使用过程中需根据实际情况修改相关步骤。

**操作步骤<a name="zh-cn_topic_0000001497364957_section38151530134817"></a>**

1. 将准备的软件包、深度学习框架相关包、host侧驱动安装信息文件及驱动版本信息文件上传到服务器同一目录（如“/home/test”）。
    - Ascend-cann-toolkit\__\{version\}_\_linux-_\{arch\}_.run
    - Ascend-cann-_\{chip\_type\}_-ops\__\{version\}_\_linux-_\{arch\}_.run
    - apex-0.1+ascend-cp310-cp310-linux\__\{arch\}_.whl
    - torch-_v\{version\}_+cpu.cxx11.abi-cp310-cp310-linux\__\{arch\}_.whl或torch-_v\{version\}_-cp3x-cp3x-manylinux\_2\_17\_aarch64.manylinux2014\_aarch64.whl
    - torch\_npu-_v\{version\}_.post<i>\{version\}</i>-cp310-cp310-manylinux\_2\_17\__\{arch\}_.manylinux2014\__\{arch\}_.whl
    - dllogger-master
    - ascend\_install.info
    - version.info

2. 以**root**用户登录服务器。
3. <a name="zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li206652021677"></a>执行以下步骤准备prebuild.sh文件。
    1. 进入软件包所在目录，执行以下命令创建prebuild.sh文件。

        ```shell
        vi prebuild.sh
        ```

    2. 写入内容参见[prebuild.sh](#zh-cn_topic_0000001497364957_li270512519175)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

4. <a name="zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li538351517716"></a>执行以下步骤准备install\_ascend\_pkgs.sh文件。
    1. 进入软件包所在目录，执行以下命令创建install\_ascend\_pkgs.sh文件。

        ```shell
        vi install_ascend_pkgs.sh
        ```

    2. 写入内容参见[install\_ascend\_pkgs.sh](#zh-cn_topic_0000001497364957_li58501140151720)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

5. <a name="zh-cn_topic_0000001497364957_zh-cn_topic_0256378845_li154641047879"></a>执行以下步骤准备postbuild.sh文件。
    1. 进入软件包所在目录，执行以下命令创建postbuild.sh文件。

        ```shell
        vi postbuild.sh
        ```

    2. 写入内容参见[postbuild.sh](#zh-cn_topic_0000001497364957_li14267051141712)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

6. 执行以下步骤准备Dockerfile文件。
    1. 进入软件包所在目录，执行以下命令创建Dockerfile文件（文件名示例“Dockerfile”）。

        ```shell
        vi Dockerfile
        ```

    2. 写入内容参见[Dockerfile](#zh-cn_topic_0000001497364957_li104026527188)编写示例，然后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

7. 进入软件包所在目录，执行以下命令，构建容器镜像，**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build -t 镜像名_系统架构:镜像tag .
    ```

    在以上命令中，各参数说明如下表所示。

    **表 2**  命令参数说明

    <a name="table18728186182510"></a>

    |参数|说明|
    |--|--|
    |-t|指定镜像名称。|
    |<em>镜像名</em><em>_系统架构:</em><em>镜像tag</em>|镜像名称与标签，请用户根据实际情况写入。|

    例如：

    ```shell
    docker build -t test_train_arm64:v1.0 .
    ```

    当出现“Successfully built xxx”表示镜像构建成功。

8. 构建完成后，执行以下命令查看镜像信息。

    ```shell
    docker images
    ```

    回显示例如下。

    ```ColdFusion
    REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
    test_train_arm64    v1.0                d82746acd7f0        27 minutes ago      749MB
    ```

**编写示例<a name="zh-cn_topic_0000001497364957_section3523631151714"></a>**

1. <a name="zh-cn_topic_0000001497364957_li270512519175"></a>prebuild.sh编写示例。

    Ubuntu  ARM系统prebuild.sh编写示例。

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # 请在此处使用bash语法编写脚本代码，执行安装准备工作，例如配置代理等
    # 本脚本将会在正式构建过程启动前被执行
    #
    # 注：本脚本运行结束后不会被自动清除，若无需保留在镜像中请在postbuild.sh脚本中清除
    #--------------------------------------------------------------------------------
    # DNS设置
    tee /etc/resolv.conf <<- EOF
    nameserver xxx.xxx.xxx.xxx  #DNS服务器IP，可填写多个，根据实际配置
    nameserver xxx.xxx.xxx.xxx
    nameserver xxx.xxx.xxx.xxx
    EOF
    # apt代理设置
    tee /etc/apt/apt.conf.d/80proxy <<- EOF
    Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";  #HTTP代理服务器IP地址及端口
    Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";  #HTTPS代理服务器IP地址及端口
    EOF
    chmod 777 -R /tmp
    rm /var/lib/apt/lists/*
    #apt源设置（以Ubuntu 18.04 ARM源为示例，请根据实际配置）
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

    Ubuntu  x86\_64系统prebuild.sh编写示例。

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------

    # 请在此处使用bash语法编写脚本代码，执行安装准备工作，例如配置代理等
    # 本脚本将会在正式构建过程启动前被执行
    #
    # 注：本脚本运行结束后不会被自动清除，若无需保留在镜像中请在postbuild.sh脚本中清除
    #--------------------------------------------------------------------------------
    # apt代理设置
    tee /etc/apt/apt.conf.d/80proxy <<- EOF
    Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";    #HTTP代理服务器IP地址及端口
    Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";   #HTTPS代理服务器IP地址及端口
    EOF

    #apt源设置（以Ubuntu 18.04 x86_64源为示例，请根据实际配置）
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

2. <a name="zh-cn_topic_0000001497364957_li58501140151720"></a>install\_ascend\_pkgs.sh编写示例。

    ```shell
    #--------------------------------------------------------------------------------
    # 请在此处使用bash语法编写脚本代码，安装昇腾软件包
    #
    # 注：本脚本运行结束后不会被自动清除，若无需保留在镜像中请在postbuild.sh脚本中清除
    #--------------------------------------------------------------------------------
    umask 0022
    cp ascend_install.info /etc/
    # 构建之前把host的/usr/local/Ascend/driver/version.info拷贝一份到当前目录
    mkdir -p /usr/local/Ascend/driver/
    cp version.info /usr/local/Ascend/driver/
    # Ascend-cann-toolkit_{version}_linux-{arch}.run
    chmod +x Ascend-cann-toolkit_{version}_linux-{arch}.run
    chmod +x Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run
    ./Ascend-cann-toolkit_{version}_linux-{arch}.run --install-path=/usr/local/Ascend/ --install --quiet
    echo y | ./Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run --install
    # 只为了安装toolkit包，所以需要清理，容器启动时通过ascend docker挂载进来
    rm -f version.info
    rm -rf /usr/local/Ascend/driver/
    ```

3. <a name="zh-cn_topic_0000001497364957_li14267051141712"></a>postbuild.sh编写示例。

    ```shell
    #--------------------------------------------------------------------------------
    # 请在此处使用bash语法编写脚本代码，清除不需要保留在容器中的安装包、脚本、代理配置等
    # 本脚本将会在正式构建过程结束后被执行
    #
    # 注：本脚本运行结束后会被自动清除，不会残留在镜像中；脚本所在位置和Working Dir位置为/tmp
    #--------------------------------------------------------------------------------
    rm -f ascend_install.info
    rm -f prebuild.sh
    rm -f install_ascend_pkgs.sh
    rm -f Dockerfile
    rm -f Ascend-cann-toolkit_{version}_linux-{arch}.run
    rm -f Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run
    rm -f apex-0.1+ascend-cp310-cp310-linux_{arch}.whl
    rm -f torch-v{version}+cpu.cxx11.abi-cp310-cp310-linux_{arch}.whl
    rm -f torch_npu-v{version}.post{version}-cp310-cp310-manylinux_2_17_{arch}.manylinux2014_{arch}.whl
    rm -f /etc/apt/apt.conf.d/80proxy

    ```

4. <a name="zh-cn_topic_0000001497364957_li104026527188"></a>Dockerfile编写示例。
    - Ubuntu  ARM系统，配套Python  3.10的Dockerfile示例。

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
        # 触发prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH || true"
        ENV http_proxy http://xxx.xxx.xxx.xxx:xxx
        ENV https_proxy http://xxx.xxx.xxx.xxx:xxx
        # 系统包
        RUN apt update && \
            apt install -y --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 \
                patch libbz2-dev llvm libncursesw5-dev xz-utils liblzma-dev m4 dos2unix libopenblas-dev libsqlite3-dev
        RUN wget https://www.python.org/ftp/python/3.10.5/Python-3.10.5.tgz
        RUN tar -zxvf Python-3.10.5.tgz && cd Python-3.10.5 && ./configure --prefix=/usr/local/python3.10.5 --enable-shared && make && make install
        RUN ln -s /usr/local/python3.10.5/bin/python3.10 /usr/local/python3.10.5/bin/python && \
            ln -s /usr/local/python3.10.5/bin/pip3.10 /usr/local/python3.10.5/bin/pip
        # 配置Python pip源
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        ENV LD_LIBRARY_PATH=/usr/local/python3.10.5/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.10.5/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.10.5/lib/python3.10/site-packages:$PYTHONPATH
        # Python包
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

        # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
        # Ascend包
        RUN umask 0022 && bash $INSTALL_ASCEND_PKGS_SH
        RUN umask 0022 && pip3 install $APEX_PKG
        RUN umask 0022 && pip3 install $PYTORCH_PKG
        RUN umask 0022 && pip3 install $PYTORCH_NPU_PKG
        RUN cd /tmp/dllogger-master/ && \
            python3 setup.py build && \
            python3 setup.py install
        # 环境变量
        ENV HCCL_WHITELIST_DISABLE=1
        ENV PYTHONPATH=/tmp/dllogger-master
        # 创建/lib64/ld-linux-aarch64.so.1
        RUN umask 0022 && \
            if [ ! -d "/lib64" ]; \
            then \
                mkdir /lib64 && ln -sf /lib/ld-linux-aarch64.so.1 /lib64/ld-linux-aarch64.so.1; \
            fi
        ENV http_proxy ""
        ENV https_proxy ""
        # 触发postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH || true" && \
            rm $POSTBUILD_SH
        ```

    - Ubuntu  x86\_64系统，配套Python  3.10的Dockerfile示例

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
        # 触发prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH || true"
        ENV http_proxy http://xxx.xxx.xxx.xxx:xxx
        ENV https_proxy http://xxx.xxx.xxx.xxx:xxx
        # 系统包
        RUN apt update && \
            apt install -y --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 \
                patch libbz2-dev llvm libncursesw5-dev xz-utils liblzma-dev m4 dos2unix libopenblas-dev libsqlite3-dev
        RUN wget https://www.python.org/ftp/python/3.10.5/Python-3.10.5.tgz
        RUN tar -zxvf Python-3.10.5.tgz && cd Python-3.10.5 && ./configure --prefix=/usr/local/python3.10.5 --enable-shared && make && make install
        RUN ln -s /usr/local/python3.10.5/bin/python3.10 /usr/local/python3.10.5/bin/python && \
            ln -s /usr/local/python3.10.5/bin/pip3.10 /usr/local/python3.10.5/bin/pip
        # 配置Python pip源
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        ENV LD_LIBRARY_PATH=/usr/local/python3.10.5/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.10.5/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.10.5/lib/python3.10/site-packages:$PYTHONPATH
        # Python包
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

        # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser
        # Ascend包
        RUN bash $INSTALL_ASCEND_PKGS_SH
        RUN pip3 install $APEX_PKG
        RUN pip3 install $PYTORCH_PKG
        RUN pip3 install $PYTORCH_NPU_PKG
        RUN cd /tmp/dllogger-master/ && \
            python3 setup.py build && \
            python3 setup.py install
        # 环境变量
        ENV HCCL_WHITELIST_DISABLE=1
        ENV PYTHONPATH=/tmp/dllogger-master
        ENV http_proxy ""
        ENV https_proxy ""
        # 触发postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH || true" && \
            rm $POSTBUILD_SH
        ```

### 使用Dockerfile构建容器镜像（MindSpore）<a name="ZH-CN_TOPIC_0000002511346627"></a>

**前提条件<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_section193545302315"></a>**

请按照[表1](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_table13971125465512)所示，获取对应操作系统的软件包与打包镜像所需Dockerfile文件与脚本文件。

软件包名称中\{version\}表示版本号、\{arch\}表示架构、\{chip\_type\}表示芯片类型。配套的CANN软件包在6.3.RC3、6.2.RC3及以上版本增加了“您是否接受EULA来安装CANN（Y/N）”的安装提示；在Dockerfile编写示例中的安装命令包含“--quiet”参数的默认同意EULA，用户可自行修改。

>[!NOTE]
>MindSpore软件包与Atlas 训练系列产品软件配套需满足对应关系，请参见MindSpore[安装指南](https://www.mindspore.cn/install)查看对应关系。

**表 1**  所需软件

<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_table13971125465512"></a>

|软件包|说明|获取方法|
|--|--|--|
|Ascend-cann-toolkit_<i>{version}</i>_linux-<i>{arch}</i>.run|CANN Toolkit开发套件包。|<p>[获取链接](https://www.hiascend.com/developer/download/community/result?module=cann)</p><p>需使用CANN 8.5.0之前版本。</p>|
|Ascend-cann-kernels<em>-{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run|CANN算子包。|<p>[获取链接](https://www.hiascend.com/developer/download/community/result?module=cann)</p><p>需使用CANN 8.5.0之前版本。</p>|
|mindspore-<em>{version}</em>-cp3<em>x</em>-cp3<em>x</em>-linux_<em>{arch}</em>.whl|<p>MindSpore框架whl包。</p><p>当前可支持Python 3.9~3.11，软件包名中x表示9、10或11，请根据实际情况选择对应软件包。</p><p>MindSpore 2.0.0版本前的软件包名由mindspore修改为mindspore-ascend。</p>|[获取链接](https://www.mindspore.cn/install)|
|Dockerfile|制作镜像需要。|参考[Dockerfile编写示例](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li104026527188)。|
|ascend_install.info|驱动安装信息文件。|从host拷贝“/etc/ascend_install.info”文件。|
|version.info|驱动版本信息文件。|从host拷贝“/usr/local/Ascend/driver/version.info”文件。|
|prebuild.sh|执行训练运行环境安装准备工作，例如配置代理等。|参考[步骤3](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li206652021677)。|
|install_ascend_pkgs.sh|昇腾软件包安装脚本。|参考[步骤4](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li538351517716)。|
|postbuild.sh|清除不需要保留在容器中的安装包、脚本、代理配置等。|参考[步骤5](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li154641047879)。|

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

>[!NOTE]
>
>- 本章节以Ubuntu 18.04、配套Python  3.9为例来介绍使用Dockerfile构建容器镜像的详细过程，使用过程中需根据实际情况修改相关步骤。
>- 如使用MindSpore  2.0.3及以上版本，需要配套使用ubuntu:20.04。

**操作步骤<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_section38151530134817"></a>**

1. 将准备的软件包、深度学习框架、host侧驱动安装信息文件及驱动版本信息文件上传到服务器同一目录（如“/home/test“）。
    - Ascend-cann-toolkit\__\{version\}_\_linux-_\{arch\}_.run
    - Ascend-cann-kernels-_\{chip\_type\}_\__\{version\}_\_linux-_\{arch\}_.run
    - mindspore-_\{version\}_-cp3x-cp3x-linux\__\{arch\}_.whl
    - ascend\_install.info
    - version.info

2. 以**root**用户登录服务器。
3. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li206652021677"></a>执行以下步骤准备prebuild.sh文件。
    1. 进入软件包所在目录，执行以下命令创建prebuild.sh文件。

        ```shell
        vi prebuild.sh
        ```

    2. 写入内容参见[prebuild.sh](#zh-cn_topic_0000001497124729_li146241711142818)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

4. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li538351517716"></a>执行以下步骤准备install\_ascend\_pkgs.sh文件。
    1. 进入软件包所在目录，执行以下命令创建install\_ascend\_pkgs.sh文件。

        ```shell
        vi install_ascend_pkgs.sh
        ```

    2. 写入内容参见[install\_ascend\_pkgs.sh](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li58501140151720)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

5. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_zh-cn_topic_0256378845_li154641047879"></a>执行以下步骤准备postbuild.sh文件。
    1. 进入软件包所在目录，执行以下命令创建postbuild.sh文件。

        ```shell
        vi postbuild.sh
        ```

    2. 写入内容参见[postbuild.sh](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li14267051141712)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

6. 执行以下步骤准备Dockerfile文件。
    1. 进入软件包所在目录，执行以下命令创建Dockerfile文件（文件名示例“Dockerfile”）。

        ```shell
        vi Dockerfile
        ```

    2. 写入内容参见[Dockerfile](#zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li104026527188)编写示例，写入后执行:wq命令保存内容，内容以Ubuntu操作系统为例。

7. 进入软件包所在目录，执行以下命令，构建容器镜像，**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build -t 镜像名_系统架构:镜像tag .
    ```

    在以上命令中，各参数说明如下表所示。

    **表 2**  命令参数说明

    <a name="table1021203815279"></a>

    |参数|说明|
    |--|--|
    |**-t**|指定镜像名称。|
    |*镜像名_系统架构:镜像tag*|镜像名称与标签，请用户根据实际情况写入。|

    例如：

    ```shell
    docker build -t test_train_arm64:v1.0 .
    ```

    当出现“Successfully built xxx”表示镜像构建成功。

8. 构建完成后，执行以下命令查看镜像信息。

    ```shell
    docker images
    ```

    回显示例如下。

    ```ColdFusion
    REPOSITORY                TAG                 IMAGE ID            CREATED             SIZE
    test_train_arm64          v1.0                d82746acd7f0        27 minutes ago      749MB
    ```

9. （可选）验证基础镜像是否可用。
    1. 执行以下命令，使用Ascend Docker Runtime在基础镜像中挂载驱动，以基础镜像test\_train\_arm64:v1.0为例。

        ```shell
        docker run -it --privileged -e ASCEND_VISIBLE_DEVICES=0 test_train_arm64:v1.0 /bin/bash
        ```

    2. 执行以下命令，查看基础镜像中MindSpore软件是否安装成功。

        ```shell
        python -c "import mindspore;mindspore.set_context(device_target='Ascend');mindspore.run_check()"
        ```

        回显示例如下，表示MindSpore软件安装成功。

        ```ColdFusion
        MindSpore version: 版本号
        The result of multiplication calculation is correct, MindSpore has been installed on platform [Ascend] successfully!
        ```

**编写示例<a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_section3523631151714"></a>**

1. <a name="zh-cn_topic_0000001497124729_li146241711142818"></a>prebuild.sh编写示例。
    - Ubuntu  ARM系统prebuild.sh编写示例。

        ```shell
        #!/bin/bash
        #--------------------------------------------------------------------------------
        # 请在此处使用bash语法编写脚本代码，执行安装准备工作，例如配置代理等
        # 本脚本将会在正式构建过程启动前被执行
        #
        # 注：本脚本运行结束后不会被自动清除，若无需保留在镜像中请在postbuild.sh脚本中清除
        #--------------------------------------------------------------------------------
        # DNS设置
        tee /etc/resolv.conf <<- EOF
        nameserver xxx.xxx.xxx.xxx  #DNS服务器IP，可填写多个，根据实际配置。
        nameserver xxx.xxx.xxx.xxx
        nameserver xxx.xxx.xxx.xxx
        EOF
        # apt代理设置
        tee /etc/apt/apt.conf.d/80proxy <<- EOF
        Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";  #HTTP代理服务器IP地址及端口。
        Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";  #HTTPS代理服务器IP地址及端口。
        EOF
        chmod 777 -R /tmp
        rm /var/lib/apt/lists/*
        #apt源设置（以Ubuntu 18.04 ARM源为示例，请根据实际配置）
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

    - Ubuntu  x86\_64系统prebuild.sh编写示例。

        ```shell
        #!/bin/bash
        #--------------------------------------------------------------------------------

        # 请在此处使用bash语法编写脚本代码，执行安装准备工作，例如配置代理等
        # 本脚本将会在正式构建过程启动前被执行
        #
        # 注：本脚本运行结束后不会被自动清除，若无需保留在镜像中请在postbuild.sh脚本中清除
        #--------------------------------------------------------------------------------
        # apt代理设置
        tee /etc/apt/apt.conf.d/80proxy <<- EOF
        Acquire::http::Proxy "http://xxx.xxx.xxx.xxx:xxx";    #HTTP代理服务器IP地址及端口
        Acquire::https::Proxy "http://xxx.xxx.xxx.xxx:xxx";   #HTTPS代理服务器IP地址及端口
        EOF

        #apt源设置（以Ubuntu 18.04 x86_64源为示例，请根据实际配置）
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

2. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li58501140151720"></a>install\_ascend\_pkgs.sh编写示例。

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # 请在此处使用bash语法编写脚本代码，安装昇腾软件包
    #
    # 注：本脚本运行结束后不会被自动清除，若无需保留在镜像中请在postbuild.sh脚本中清除
    #--------------------------------------------------------------------------------
    # 构建之前把host上的/etc/ascend_install.info拷贝一份到当前目录
    cp ascend_install.info /etc/
    mkdir -p /usr/local/Ascend/driver/
    cp version.info /usr/local/Ascend/driver/

    # Ascend-cann-toolkit_{version}_linux-{arch}.run
    chmod +x Ascend-cann-toolkit_{version}_linux-{arch}.run
    ./Ascend-cann-toolkit_{version}_linux-{arch}.run --install-path=/usr/local/Ascend/ --install --quiet
    chmod +x Ascend-cann-kernels-{chip_type}_{version}_linux-{arch}.run
    ./Ascend-cann-kernels-{chip_type}_{version}_linux-{arch}.run --install --quiet

    # 只安装toolkit包，需要清理，容器启动时通过ascend docker挂载进来
    rm -f version.info
    rm -rf /usr/local/Ascend/driver/
    ```

3. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li14267051141712"></a>postbuild.sh编写示例。

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # 请在此处使用bash语法编写脚本代码，清除不需要保留在容器中的安装包、脚本、代理配置等
    # 本脚本将会在正式构建过程结束后被执行
    #
    # 注：本脚本运行结束后会被自动清除，不会残留在镜像中；脚本所在位置和Working Dir位置为/root
    #--------------------------------------------------------------------------------

    rm -f ascend_install.info
    rm -f prebuild.sh
    rm -f install_ascend_pkgs.sh
    rm -f Dockerfile
    rm -f version.info
    rm -f Ascend-cann-toolkit_{version}_linux-{arch}.run
    rm -f Ascend-cann-kernels-{chip_type}_{version}_linux-{arch}.run
    # 请根据实际安装的版本选择需要删除的包
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

4. <a name="zh-cn_topic_0000001497124729_zh-cn_topic_0272789326_li104026527188"></a>Dockerfile编写示例。
    - Ubuntu  ARM系统，配套Python  3.9的Dockerfile示例。

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

        # 触发prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH"

        ENV http_proxy http://xxx
        ENV https_proxy http://xxx


        # 系统包
        RUN apt update && \
            apt install --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 -y

        RUN wget https://www.python.org/ftp/python/3.9.2/Python-3.9.2.tgz
        RUN tar -zxvf Python-3.9.2.tgz && cd Python-3.9.2 && ./configure --prefix=/usr/local/python3.9.2 --enable-shared && make && make install

        RUN ln -s /usr/local/python3.9.2/bin/python3.9 /usr/local/python3.9.2/bin/python && \
            ln -s /usr/local/python3.9.2/bin/pip3.9 /usr/local/python3.9.2/bin/pip


        # 配置Python pip源
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        # 用户需根据实际情况修改PYTHONPATH的路径
        ENV LD_LIBRARY_PATH=/usr/local/python3.9.2/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.9.2/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.9.2/lib/python3.9/site-packages:$PYTHONPATH
        # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser

        # 安装Python3.9。若安装其他版本，请根据实际情况修改以下命令
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

        # Ascend包
        RUN umask 0022 && bash $INSTALL_ASCEND_PKGS_SH

        # MindSpore安装
        RUN pip install $MINDSPORE_PKG

        ENV http_proxy ""
        ENV https_proxy ""

        # 触发postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH" && \
            rm $POSTBUILD_SH
        ```

    - Ubuntu  x86\_64系统，配套Python  3.9的Dockerfile示例。

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

        # 触发prebuild.sh
        RUN bash -c "test -f $PREBUILD_SH && bash $PREBUILD_SH"

        ENV http_proxy http://xxx
        ENV https_proxy http://xxx


        # 系统包
        RUN apt update && \
            apt install --no-install-recommends curl g++ pkg-config unzip wget build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libreadline-dev libffi-dev \
                libblas3 liblapack3 liblapack-dev openssl libssl-dev libblas-dev gfortran libhdf5-dev libffi-dev libicu60 libxml2 -y

        RUN wget https://www.python.org/ftp/python/3.9.2/Python-3.9.2.tgz
        RUN tar -zxvf Python-3.9.2.tgz && cd Python-3.9.2 && ./configure --prefix=/usr/local/python3.9.2 --enable-shared && make && make install

        RUN ln -s /usr/local/python3.9.2/bin/python3.9 /usr/local/python3.9.2/bin/python && \
            ln -s /usr/local/python3.9.2/bin/pip3.9 /usr/local/python3.9.2/bin/pip

        # 配置Python pip源
        RUN mkdir -p ~/.pip \
        && echo '[global] \n\
        index-url=https://pypi.doubanio.com/simple/\n\
        trusted-host=pypi.doubanio.com' >> ~/.pip/pip.conf

        # 用户需根据实际情况修改PYTHONPATH的路径
        ENV LD_LIBRARY_PATH=/usr/local/python3.9.2/lib:$LD_LIBRARY_PATH
        ENV PATH=/usr/local/python3.9.2/bin:$PATH
        ENV PYTHONPATH=/usr/local/python3.9.2/lib/python3.9/site-packages:$PYTHONPATH
        # 创建HwHiAiUser用户和属主，UID和GID请与物理机保持一致避免出现无属主文件。示例中会自动创建user和对应的group，UID和GID都为1000
        RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser

        # 安装Python3.9。若安装其他版本，请根据实际情况修改以下命令
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

        # Ascend包
        RUN umask 0022 && bash $INSTALL_ASCEND_PKGS_SH

        # MindSpore安装
        RUN pip install $MINDSPORE_PKG

        ENV http_proxy ""
        ENV https_proxy ""

        # 触发postbuild.sh
        RUN bash -c "test -f $POSTBUILD_SH && bash $POSTBUILD_SH" && \
            rm $POSTBUILD_SH
        ```

### 使用Dockerfile构建推理镜像<a name="ZH-CN_TOPIC_0000002479386680"></a>

**前提条件<a name="zh-cn_topic_0000001497364777_section193545302315"></a>**

请按照[表1](#zh-cn_topic_0000001497364777_table13971125465512)所示，获取对应操作系统的软件包与打包镜像所需Dockerfile文件与脚本文件。

软件包名称中\{version\}表示版本号、\{arch\}表示架构、\{chip\_type\}表示芯片类型。配套的CANN软件包在6.3.RC3、6.2.RC3及以上版本增加了“您是否接受EULA来安装CANN（Y/N）”的安装提示；在Dockerfile编写示例中的安装命令包含“--quiet”参数的默认同意EULA，用户可自行修改。

**表 1**  所需软件

<a name="zh-cn_topic_0000001497364777_table13971125465512"></a>

|软件包|说明|获取方法|
|--|--|--|
|Ascend-cann-toolkit_<i>{version}</i>_linux-<i>{arch}</i>.run|CANN Toolkit开发套件包。|[获取链接](https://www.hiascend.com/developer/download/community/result?module=cann)|
|Ascend-cann-<em>{chip_type}</em>-ops_<em>{version}</em>_linux-<em>{arch}</em>.run|<p>CANN算子包。</p><p>CANN 8.5.0之前版本该包名为Ascend-cann-kernels-<em>{chip_type}</em>_<em>{version}</em>_linux-<em>{arch}</em>.run</p>|[获取链接](https://www.hiascend.com/developer/download/community/result?module=cann)|
|Dockerfile|制作镜像需要。|参考[Dockerfile编写示例](#zh-cn_topic_0000001497364777_li166241028113511)。|
|install.sh|安装推理业务的脚本。|推理模型的制作可以参考[ResNet50推理指导](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/ACL_PyTorch/built-in/cv/Resnet50_Pytorch_Infer)。|
|<em>XXX</em>.tar|推理业务代码包名称，用户根据推理业务准备。本章以dvpp_resnet.tar为例说明。|推理模型的制作可以参考[ResNet50推理指导](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/ACL_PyTorch/built-in/cv/Resnet50_Pytorch_Infer)。|
|run.sh|启动推理服务的脚本。|推理模型的制作可以参考[ResNet50推理指导](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/ACL_PyTorch/built-in/cv/Resnet50_Pytorch_Infer)。|

>[!NOTE]
>推理需要的其他软件包和代码请用户自行准备。

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

本章节以Ubuntu  x86\_64操作系统为例，以下操作步骤中的代码为示例代码，用户可参考示例进行定制化修改，并且建议用户对示例代码和镜像做安全加固，可参考[容器安全加固](./04_security_hardening.md#容器安全加固)。

**操作步骤<a name="zh-cn_topic_0000001497364777_section9307172524312"></a>**

1. 将准备的软件包及文件上传到服务器同一目录（如“/home/infer”）。
    - Ascend-cann-toolkit\__\{version\}_\_linux-_\{arch\}_.run
    - Ascend-cann-_\{chip\_type\}_-ops\__\{version\}_\_linux-_\{arch\}_.run
    - Dockerfile
    - install.sh
    - run.sh
    - _XXX_.tar（自行准备的推理代码或脚本）

2. 以**root**用户登录服务器。
3. 执行以下步骤准备install.sh文件。
    1. 进入软件包所在目录，执行以下命令创建install.sh文件。

        ```shell
        vi install.sh
        ```

    2. 参见[install.sh](#zh-cn_topic_0000001497364777_li18749540133416)编写示例，请根据业务实际编写，写入后执行:wq命令保存内容。

4. 执行以下步骤准备run.sh文件。
    1. 进入软件包所在目录，执行以下命令创建run.sh文件。

        ```shell
        vi run.sh
        ```

    2. 参见[run.sh](#zh-cn_topic_0000001497364777_li18234181353511)编写示例，请根据业务实际编写，写入后执行:wq命令保存内容。

5. 执行以下步骤准备Dockerfile文件。
    1. 进入软件包所在目录，执行以下命令创建Dockerfile文件（文件名示例“Dockerfile”）。

        ```shell
        vi Dockerfile
        ```

    2. 参见[Dockerfile](#zh-cn_topic_0000001497364777_li166241028113511)编写示例，请根据业务实际编写，写入后执行:wq命令保存内容。

6. 进入软件包所在目录，执行以下命令，构建容器镜像，**注意不要遗漏命令结尾的**“.”。

    ```shell
    docker build --build-arg TOOLKIT_VERSION={version} --build-arg TOOLKIT_ARCH={arch} --build-arg DIST_PKG=XXX.tar -t 镜像名_系统架构:镜像tag .
    ```

    在以上命令中，各参数说明如下表所示。

    **表 2**  命令参数说明

    <a name="zh-cn_topic_0000001497364777_zh-cn_topic_0256378845_table47051919193111"></a>

    |参数|说明|
    |--|--|
    |--build-arg|指定Dockerfile文件内的参数。|
    |<em>{version}</em>|Toolkit包版本号，请用户根据实际情况写入。|
    |<em>{arch}</em>|Toolkit包架构，请用户根据实际情况写入。|
    |<em>XXX</em>.tar|推理业务代码包名称，用户根据实际情况写入。|
    |-t|指定镜像名称。|
    |<em>镜像名</em><em>_系统架构:</em><em>镜像tag</em>|镜像名称与标签，请用户根据实际情况写入。|

    示例如下。

    ```shell
    docker build --build-arg TOOLKIT_VERSION=20.1.rc3 --build-arg TOOLKIT_ARCH=x86_64 --build-arg DIST_PKG=dvpp_resnet.tar -t ubuntu-infer:v1 .
    ```

    当出现“Successfully built xxx”表示镜像构建成功。

7. 构建完成后，执行以下命令查看镜像信息。

    ```shell
    docker images
    ```

    回显示例如下：

    ```ColdFusion
    REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
    ubuntu-infer        v1                  fffbd83be42a        2 minutes ago       293MB
    ```

**编写示例<a name="zh-cn_topic_0000001497364777_section158942057133318"></a>**

1. <a name="zh-cn_topic_0000001497364777_li18749540133416"></a>install.sh编写示例。

    ```shell
    #!/bin/bash
    #--------------------------------------------------------------------------------
    # 安装推理业务脚本，此处以推理业务包dvpp_resnet.tar为例说明，用户可自行修改业务包名
    #-------------------------------------
    tar -xvf dvpp_resnet.tar
    # 同时建议修改解压后文件的权限和属主
    ```

2. <a name="zh-cn_topic_0000001497364777_li18234181353511"></a>run.sh编写示例。

    ```shell
    #!/bin/bash
    # 运行业务代码
    cd /home/out
    numbers=`ls /dev/| grep davinci | grep -v davinci_manager | wc -l`
    # 每5分钟更新日志
    #./main $numbers|grep -nE '.*\[.*[[:digit:]]{2}:[[:digit:]]{1}[05]:00\]' >./log.txt
    # 加载离线推理环境变量
    export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64:${LD_LIBRARY_PATH}
    source /usr/local/Ascend/cann/set_env.sh
    ./main $numbers
    ```

    >[!NOTE]
    >LD\_LIBRARY\_PATH中配置了驱动相关路径，在执行推理作业时，会使用到其中的文件。建议推理作业的运行用户和驱动安装时指定的运行用户保持一致，避免用户不一致带来的提权风险。

3. <a name="zh-cn_topic_0000001497364777_li166241028113511"></a>Dockerfile编写示例，请根据实际情况进行定制化修改。

    ```Dockerfile
    #基础镜像ubuntu:18.04不包含Toolkit包，可参考Dockerfile示例中的部分步骤进行安装，需提前准备好Toolkit包。
    #推荐从昇腾镜像仓库拉取推理基础镜像，此时的镜像中已安装Toolkit包。同时请确认Toolkit包是否与物理机上的驱动版本匹配
    FROM ubuntu:18.04

    # 设置Toolkit包、OPS包参数
    ARG TOOLKIT_VERSION
    ARG TOOLKIT_ARCH
    ARG TOOLKIT_PKG=Ascend-cann-toolkit_{version}_linux-{arch}.run
    ARG OPS_PKG=Ascend-cann-{chip_type}-ops_{version}_linux-{arch}.run


    # 设置环境变量
    ARG ASCEND_BASE=/usr/local/Ascend

    # 设置进入启动后的容器的目录
    WORKDIR /home

    #拷贝Toolkit包和OPS包
    COPY $TOOLKIT_PKG .
    COPY $OPS_PKG .

    # 安装Toolkit包和OPS包
    RUN umask 0022 && \
        groupadd xxx（自定义用户,需要与驱动安装指定的一致） && \
        useradd -g xxx（自定义用户,需要与驱动安装指定的一致） -s /usr/sbin/nologin（禁止用户登录，示例为Ubuntu系统） -m -d /home/xxx xxx（自定义用户,需要与驱动安装指定的一致） && \
        chmod +x ${TOOLKIT_PKG} &&\
        ./${TOOLKIT_PKG} --quiet --install --install-for-all --whitelist=nnrt --force &&\
        rm ${TOOLKIT_PKG} &&\
        chmod +x ${OPS_PKG} &&\
        ./${OPS_PKG} --install --install-for-all --quiet --force &&\
        rm ${OPS_PKG}

    # 拷贝业务推理程序压缩包、安装脚本与运行脚本
    ARG DIST_PKG
    COPY $DIST_PKG .
    COPY install.sh .
    COPY run.sh .

    # 运行安装脚本
    RUN mkdir -p /usr/slog && \
        mkdir -p /var/log/npu/slog/slogd && \
        chmod u+x run.sh install.sh && \
        sh install.sh && \
        rm $DIST_PKG && \
        rm install.sh

    CMD bash run.sh
    ```

    >[!NOTE]
    >CANN软件包版本为6.2.RC1、6.3.RC1及其之后的版本，在安装软件包时新增--force参数。上述Dockerfile编写示例已经加上该参数。若用户使用6.2.RC1、6.3.RC1之前版本的软件包，需要去除Dockerfile编写示例中该参数。

## 获取集群内当前可用设备信息<a name="ZH-CN_TOPIC_0000002516255287"></a>

1. 查询ConfigMap。

    ```shell
    kubectl get cm -A | grep cluster-info
    ```

    回显示例如下：

    ```ColdFusion
    kube-public            cluster-info                                           1      19d
    mindx-dl               cluster-info-device-0                                  1      19h
    mindx-dl               cluster-info-node-cm                                   1      19h
    mindx-dl               cluster-info-switch-0                                  1      19h
    ```

2. 查询ConfigMap的详细信息，获取可用设备信息。下面以节点名为localhost.localdomain为例。

    1. 查询与device相关的ConfigMap的详细信息，获取节点可用芯片信息。

        ```shell
        kubectl describe cm -n mindx-dl cluster-info-device-0
        ```

        回显示例如下：

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

        从以上回显信息可以看到，该节点的可用芯片为Ascend910-3、Ascend910-4、Ascend910-5、Ascend910-6、Ascend910-7。

    2. 查询与node相关的ConfigMap的详细信息，获取节点状态信息。

        ```shell
        kubectl describe cm -n mindx-dl cluster-info-node-cm
        ```

        回显示例如下：

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

        从以上回显信息可以看到，该节点的NodeStatus为Healthy，表示当前节点健康。

    3. 查询与Switch相关的ConfigMap的详细信息，获取节点状态信息。

        ```shell
        kubectl describe cm -n mindx-dl cluster-info-switch-0
        ```

        回显示例如下：

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

        从以上回显信息可以看到，该节点的NodeStatus为Healthy，表示当前节点健康。

    综合以上查询结果可知，该节点的可用芯片为Ascend910-3、Ascend910-4、Ascend910-5、Ascend910-6、Ascend910-7。

    若步骤2或步骤3的回显信息中NodeStatus为UnHealthy，则说明当前节点上的设备均不可用。结合步骤1的查询结果可知，该节点的可用芯片为空。

    >[!NOTE]
    >当集群规模超过1000节点时，cluster-info-device-和mindx-dl-switchinfo-对应的ConfigMap会进行分片。每个cluster-info-device-或mindx-dl-switchinfo-最多包含1000个节点的设备信息。针对此种场景，需要对所有cluster-info-device-的ConfigMap都执行步骤1和步骤3的查询操作，找到目标节点的详细信息，才能确认该节点的可用芯片信息。

## 容器基础镜像集成UMDK安装指导<a name="ZH-CN_TOPIC_0000002516255288"></a>

1. 从[华为云镜像仓地址](https://mirrors.huaweicloud.com/ascend/)的archive目录，根据需要下载对应格式的UMDK软件包，以umdk-urma-*.aarch64.rpm.tar.gz为例。
2. 以root用户登录服务器。
3. 创建/home/umdk目录，将UMDK软件包上传到该目录下。

   ```shell
   mkdir -p /home/umdk
   ```

4. 进入/home/umdk目录，执行以下步骤准备Dockerfile文件。
    1. 执行以下命令创建Dockerfile文件（文件名示例“Dockerfile”）。

        ```shell
        vi Dockerfile
        ```

    2. 在文件中新增UMDK安装的相关指令。
       - openEuler容器基础镜像，以openEuler 24.03-lts版本为例。

            ```Dockerfile
            FROM openeuler/openeuler:24.03-lts

            ARG UMDK_PKG=""
            COPY ./${UMDK_PKG} /tmp

            RUN yum update -y && \
                yum install -y wget unzip shadow libnl3-devel && \
                if [ -n "${UMDK_PKG}" ] && [ -f "/tmp/${UMDK_PKG}" ]; then \
                    echo "installing umdk from /tmp/${UMDK_PKG}"; \
                    mkdir /tmp/umdk_pkgs; \
                    tar -mzxf "/tmp/${UMDK_PKG}" -C /tmp/umdk_pkgs; \
                    rpm -ivh /tmp/umdk_pkgs/*.rpm; \
                else \
                    echo "warning: umdk package not provided, install from yum"; \
                    yum install -y umdk-urma-bin umdk-urma-devel umdk-urma-lib umdk-urma-tools; \
                fi && \
                yum clean all && \
                rm -rf /var/cache/yum && \
                rm -rf /tmp/*
            ```

       - Ubuntu容器基础镜像，以Ubuntu 24.04版本为例。

           ```Dockerfile
           FROM ubuntu:24.04

           ARG UMDK_PKG=""

           COPY ./${UMDK_PKG} /tmp/

           RUN apt update && \
               apt install -y libnl-3-dev libnl-genl-3-dev && \
               if [ -n "${UMDK_PKG}" ] && [ -f "/tmp/${UMDK_PKG}" ]; then \
                   echo "installing umdk from /tmp/${UMDK_PKG}"; \
                   mkdir /tmp/umdk_pkgs; \
                   tar -mzxf "/tmp/${UMDK_PKG}" -C /tmp/umdk_pkgs ; \
                   dpkg -i /tmp/umdk_pkgs/*.deb; \
               else \
                   echo "error: umdk package not provided"; \
                   exit 1; \
              fi && \
              apt-get clean && \
              rm -rf /var/lib/apt/lists/* && \
              rm -rf /tmp/*
           ```

    3. 将新增内容写入Dockerfile文件后执行:wq命令保存内容。

5. 进入Dockerfile文件所在目录，执行镜像构建指令，构建镜像。注意不要遗漏命令结尾的“.”。

    ```shell
    docker build [OPTIONS] -t 镜像名_系统架构:镜像tag --build-arg UMDK_PKG=UMDK软件包名 .
    ```

   命令解释如下表所示。

   **表 1**  命令参数说明

   <a name="zh_table47051919193112"></a>

   |参数| 说明                        |
   |--|---------------------------|
   |<em>OPTIONS</em>| “--no-cache”选项：不使用缓存重建镜像。 |
   |-t| 指定镜像名称。                   |
   |<em>镜像名</em><em>_系统架构:</em><em>镜像tag</em>| 镜像名称与标签，请用户根据实际情况写入。      |
   |--build-arg UMDK_PKG| 指定UMDK软件包名。               |

   例如：

   ```shell
   docker build --no-cache -t urma_aarch64:oe_2403lts --build-arg UMDK_PKG=umdk-urma-25.12.0-B090.sp4.aarch64.rpm.tar.gz .
   ```

   当出现“Successfully built xxx”表示镜像构建成功。
6. 构建完成后，执行以下命令查看镜像信息。

    ```shell
    docker images
    ```

   回显示例如下。

    ```ColdFusion
    REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
    urma_aarch64        oe_2403lts          d82746acd7f1        26 minutes ago      200MB
    ```
