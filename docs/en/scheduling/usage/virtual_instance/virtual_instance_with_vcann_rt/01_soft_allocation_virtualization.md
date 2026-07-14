# Soft Partitioning-based Virtualization

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:22:24.734Z pushedAt=2026-06-30T12:23:24.398Z -->

## Notes on Using Soft NPU Partitioning <a name="ZH-CN_TOPIC_00000025113463450356vcann"></a>

In Kubernetes scenarios, when you need to use NPU resources, you must combine the cluster scheduling components Ascend Device Plugin and Volcano to enable Kubernetes to manage and schedule Ascend processor resources. The cluster scheduling components required for the Ascend soft partitioning-based virtual instance include Ascend Device Plugin, Volcano, Ascend Docker Runtime, Ascend Operator, and ClusterD. For supported product models, see "Table 1 Supported products" in [Feature Description](./00_description.md).

## Scenario Description<a name="section1576110260450vcann"></a>

Before using soft partitioning-based virtualization, you need to understand the scenario descriptions in [Table 1 Scenario description](#table62551184461989657).

**Table 1**  Scenario description

<a name="table62551184461989657"></a>
<table><thead align="left"><tr><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p>Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="80.02%" id="mcps1.2.3.1.2"><p>Description</p>
</th>
</tr>
</thead>
<tbody><tr><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p>General description</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p>The allocated chip information is reflected in the labels of the PodGroup. For detailed descriptions of PodGroup labels, see the following parameters in <a href="../../../api/volcano.md#podgroup">PodGroup label</a>: <ul><li>huawei.com/scheduler.softShareDev.aicoreQuota</li><li>huawei.com/scheduler.softShareDev.hbmQuota</li><li>huawei.com/scheduler.softShareDev.policy</li></ul></p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>Soft partitioning function must be used in conjunction with vCANN-RT.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>When allocating soft-partitioned NPUs, MindCluster scheduling prioritizes fully occupying the physical NPU with the least remaining computing power.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>Currently, each Pod of a job requests 1 NPU. The number of NPUs physically used is 1, but the number of NPUs requested in the job YAML must be consistent with the huawei.com/scheduler.softShareDev.aicoreQuota configuration.</p>
</td>
</tr>
<tr><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p>Supported scenarios</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p>Multiple replicas are supported, but the NPU soft partitioning policy used by each Pod in the multiple replicas must be consistent.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>K8s mechanisms, such as affinity, are supported.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>Rescheduling upon chip faults and node faults is supported. For details, see the <a href="../../basic_scheduling/10_recovery_of_inference_card_faults.md">Recovery of Inference Card Faults</a> and <a href="../../basic_scheduling/09_rescheduling_upon_inference_card_faults.md">Rescheduling upon Inference Card Faults</a>.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>Supports scenarios where soft partitioning-based virtualization and non-soft partitioning-based virtualization functions are deployed in a mixed manner within a cluster.</p>
</td>
</tr>
<tr><td class="cellrowborder" rowspan="3" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p>Unsupported scenarios</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p>Mixing different chips within a single job is not supported.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>Uninstalling Volcano during task execution is not supported.</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p>Mixing with operations in Docker scenarios is not supported.</p>
</td>
</tr>
</tbody>
</table>

## Prerequisites

1. You need to add the label `huawei.com/scheduler.chip1softsharedev.enable=true` to the node, indicating that the node supports the soft partitioning function.

    ```shell
    kubectl label nodes <Node_name> huawei.com/scheduler.chip1softsharedev.enable=true
    ```

    In a mixed deployment scenario of soft partitioning-based virtualization and non-soft partitioning-based virtualization, if a node does not support soft partitioning-based virtualization, you need to add the label `huawei.com/scheduler.chip1softsharedev.enable=false` to the node.

2. You need to obtain `Ascend-docker-runtime_{version}_linux-{arch}.run` to install the container engine plugin.
3. See the [Installation and Deployment](../../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) chapter to complete the installation of each component.

    The involved component in modifying related parameters for virtual instances is Ascend Device Plugin. Please modify and use the corresponding YAML for installation and deployment as required below:

    1. Add `-shareDevCount=100 -softShareDevConfigDir=/share_device/` in `device-plugin-volcano-v{version}.yaml`, where `/share_device/` is manually created by the user. When Atlas A3 inference series products use soft partitioning-based virtualization, you need to additionally add the startup parameter `-useSingleDieMode=true`.

       ```Yaml
       ...

               args: [ "device-plugin  -useAscendDocker=true -volcanoType=true -presetVirtualDevice=true
                 -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=0 -shareDevCount=100 -softShareDevConfigDir=/share_device/ -useSingleDieMode=true" ]   # Only when Atlas A3 inference series products use the soft partitioning-based virtualization function, you need to add -useSingleDieMode=true
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

        The startup parameters for soft partitioning-based virtualization are described as follows:

        **Table 2** Ascend Device Plugin startup parameters

       <a name="table1064314568229"></a>

       |Name|Type|Mandatory|Description|
       |--|--|--|--|
       |-shareDevCount|uint|1|To use soft partitioning-based virtualization function, the value can only be 100.|
       |-softShareDevConfigDir|string|""|Configuration directory for the soft partitioning-based virtualization scenario.|
       |-useSingleDieMode|bool|false|Whether to enable single-die passthrough mode for Atlas A3 inference series products.<ul><li>true: Enable single-die passthrough mode.</li><li>false: Disable single-die passthrough mode.</li></ul>To use the soft partitioning-based virtualization function, this parameter must be set to true.|

    2. (Optional) For hybrid deployment scenarios involving soft partitioning-based virtualization and non-soft partitioning-based virtualization, the YAML of Ascend Device Plugin needs to be modified as follows.

       - Install Ascend Device Plugin that supports soft partitioning on nodes that support soft partitioning-based virtualization, and copy `device-plugin-volcano-v{version}.yaml to softsharedev-device-plugin-volcano-v{version}.yaml`. Modify `softsharedev-device-plugin-volcano-v{version}`.yaml as follows:

         ```Yaml
         apiVersion: apps/v1
         kind: DaemonSet
         metadata:
           name: ascend-device-plugin-daemonset-910-softsharedev # Identifies that Ascend Device Plugin supports the soft partitioning-based virtualization function in a mixed deployment scenario with both soft partitioning-based virtualization and non-soft partitioning-based virtualization functions
           namespace: kube-system
         spec:
           ...
           template:
           ...
             spec:
             ...
               nodeSelector:
                 huawei.com/scheduler.chip1softsharedev.enable: "true"  # Select nodes that support the soft partitioning-based virtualization function to deploy Ascend Device Plugin
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

       - Install the original Ascend Device Plugin on nodes that do not support the soft partitioning-based virtualization function. Modify `device-plugin-volcano-v{version}.yaml` as follows:

         ```Yaml
         apiVersion: apps/v1
         kind: DaemonSet
         metadata:
           name: ascend-device-plugin-daemonset-910 # Identifies that Ascend Device Plugin does not support the soft partitioning-based virtualization function in a mixed deployment scenario with both soft partitioning-based virtualization and non-soft partitioning-based virtualization functions
           namespace: kube-system
         spec:
           ...
           template:
           ...
             spec:
             ...
               nodeSelector:
                 huawei.com/scheduler.chip1softsharedev.enable: "false"  # Select nodes that do not support the soft partitioning-based virtualization function to deploy Ascend Device Plugin
                 accelerator: huawei-Ascend910
               serviceAccountName: ascend-device-plugin-sa-910
           ...
         ```

## Usage

Modify the following configuration when creating a YAML file upon inference job creation (Atlas 800I A2 inference server as an example).

The parameter configuration example for applying for a chip AICore percentage of 50%, chip high-bandwidth memory of 2048 MB, and a soft partitioning policy of fixed-share is as follows. 

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: default-infer-test-pytorch-910b
  labels:
    framework: pytorch
    ring-controller.atlas: ascend-910b
    fault-scheduling: "force"
    <strong>huawei.com/scheduler.softShareDev.aicoreQuota: "50" # Percentage of chip AICore requested by the soft partitioning task, in %</strong>
    <strong>huawei.com/scheduler.softShareDev.hbmQuota: "2048" # Amount of chip high-bandwidth memory requested by the soft partitioning task, in MB</strong>
    <strong>huawei.com/scheduler.softShareDev.policy: "fixed-share" # Soft partitioning policy, with values of fixed-share, elastic, and best-effort</strong>
  annotations:
    <strong>huawei.com/schedule_policy: "chip1-softShareDev" # Volcano scheduling policy in soft partitioning scenarios</strong>
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
                  <strong>huawei.com/Ascend910: 50 # This value must be consistent with the value of huawei.com/scheduler.softShareDev.aicoreQuota, indicating the AICore percentage requested by the soft partitioning task</strong>
                limits:
                  <strong>huawei.com/Ascend910: 50 # The value must be consistent with requests</strong>
              volumeMounts:
                - name: ascend-driver
                  mountPath: /usr/local/Ascend/driver
                - name: ascend-add-ons
                  mountPath: /usr/local/Ascend/add-ons
                - name: localtime
                  mountPath: /etc/localtime
                <strong>- name: libpreload # soft partitioning dynamic library path</strong>
                  <strong>mountPath: /opt/enpu/vcann-rt/lib/libvruntime.so</strong>
                <strong>- name: preload # preload configuration file path</strong>
                  <strong>mountPath: ${preload_path}/ld.so.preload</strong>
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
            <strong>- name: libpreload # soft partitioning dynamic library path</strong>
              <strong>hostPath:</strong>
                <strong>path: /opt/enpu/vcann-rt/lib/libvruntime.so</strong>
            <strong>- name: preload # preload configuration file path</strong>
              <strong>hostPath:</strong>
                <strong>path: ${preload_path}/ld.so.preload</strong>
</pre>

>[!NOTE]
>When submitting a soft partitioning-based virtualization task for <term>Atlas A3 inference series products</term>, in the container, `/dev/` actually mounts 1 die, but running the <b>npu-smi info</b> command shows that 2 dies are mounted. The echo example is as follows:
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
> +===========================+---------------+===================================================+
> +---------------------------+---------------+---------------------------------------------------+
> | NPU     Chip       | Process id    | Process name| Process memory(MB) |Process id in container|
> +===========================+===============+===================================================+
> | No running processes found in NPU 0                                                           |
> +===========================+===============+===================================================+
> ```
