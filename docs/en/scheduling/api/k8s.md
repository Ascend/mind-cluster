# K8s Native Object Description<a name="ZH-CN_TOPIC_0000002511346725"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:42:29.577Z pushedAt=2026-06-09T02:05:50.635Z -->

## Service Labels<a name="section17127184555719"></a>

**Table 1**  Service labels used by cluster scheduling components

|Name|Function|Value|Component|
|--|--|--|--|
|group-name|Marks the group name of the acjob corresponding to the Pod|mindxdl.gitee.com|Volcano, Ascend Operator|
|job-name|Marks the acjob name corresponding to the Pod|String|Ascend Operator|
|replica-index|Marks the Pod sequence number (to be deleted later)|[0-{Pod Quantity-1}]|Ascend Operator|
|replica-type|Marks the Pod type (to be deleted later)|<ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul>|Ascend Operator|
|training.kubeflow.org/job-name|Marks the acjob name corresponding to the Pod|String|Ascend Operator|
|training.kubeflow.org/operator-name|Marks the operator name that created the Pod|ascendjob-controller|Ascend Operator|
|training.kubeflow.org/replica-index|Marks the Pod sequence number|[0-{Pod Quantity-1}]|Ascend Operator|
|training.kubeflow.org/replica-type|Marks the Pod type|<ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul>|Ascend Operator|

## Job Labels<a name="section3960559173617"></a>

**Table 2** Job labels used by cluster scheduling components

|Job Label|Function|Value|Component|
|--|--|--|--|
|mind-cluster/scaling-rule: scaling-rule|Marks the ConfigMap name corresponding to the scaling rule.|String|Ascend Operator|
|mind-cluster/group-name: group0|Marks the group name corresponding to the scaling rule.|String|Ascend Operator|

## Job Annotations

**Table 3** Job annotations used by cluster scheduling components

|Job Annotation|Function|Value|Component|
|--|--|--|--|
|huawei.com/schedule.filter.faultCode|<p>Configures the fault codes and time windows that need to be silenced for the configuration task.</p><ul><li>Fault codes only support configuring fault codes for chip faults and UnifiedBus device faults. For details on supported fault codes, see the faultCode.json and SwitchFaultCode.json files.</li><li>Supports configuring multiple fault codes and time windows, separated by commas.</li></ul>|<p>Value example: "8C1F8608:30, 80E01801", indicating that within a 30-second time window, the 8C1F8608 fault is silenced; within a 60-second time window, the 80E01801 fault is silenced.</p><p>If no time window is configured, the default is 60. The value range is 0 to 86400, in seconds.</p>|ClusterD|
|huawei.com/schedule.filter.faultLevel|<p>Configures the fault levels and time windows that need to be silenced for the configuration task.</p><ul><li>Fault levels only support configuring levels for chip faults and UnifiedBus device faults. For details on supported fault levels, see [Configuration Description](../usage/resumable_training/03_configuring_fault_detection_levels.md#configuration-description).</li><li>Supports configuring multiple fault levels and time windows, separated by commas.</li><li>For MindIE Service, if this configuration item is absent in the YAML file, all RestartRequest-level faults are silenced for 60 seconds by default.</li><li>The priority of huawei.com/schedule.filter.faultCode is higher than that of huawei.com/schedule.filter.faultLevel.</li><li>For notification-type faults, after ClusterD silences such faults, Volcano may not actively reschedule the faulty Pod. The task can subscribe to ClusterD's fault subscription interface to handle received faults accordingly. If processing fails, the Pod must actively exit with an error.</li></ul>|<p>Value example: "RestartRequest:30, RestartBusiness", indicating that within a 30-second time window, all RestartRequest-level faults are silenced; within a 60-second time window, all RestartBusiness-level faults are silenced.</p><p>If no time window is configured, the default is 60. The value range is 0 to 86400, in seconds.</p>|ClusterD|

## Node Labels<a name="section121401114162912"></a>

**Table 4**  Node labels used by cluster scheduling components

|Node Label|Function|Value|Component|
|--|--|--|--|
|accelerator|Identifies the processing chip of the node|<ul><li>huawei-npu</li><li>huawei-Ascend910</li><li>huawei-Ascend310</li><li>huawei-Ascend310P</li></ul>|Ascend Device Plugin|
|host-arch|Identifies the CPU architecture of the node|<ul><li>huawei-x86</li><li>huawei-arm</li></ul>|Volcano|
|masterselector|Identifies the management node of MindCluster|dls-master-node|Volcano, Ascend Operator, Resilience Controller, ClusterD|
|node.kubernetes.io/npu.chip.name|Reports the specific type of the current chip|<ul><li>310</li><li>310P1</li><li>310P2</li><li>310P3</li><li>310P4</li><li>{xxx}A</li><li>910PremiumA</li><li>910ProA</li><li>910ProB</li><li>{xxx}Bx (x can be 1, 2, 3, or 4)</li><li>Ascend950PR</li><li>Ascend950DT</li></ul>|<p>Ascend Device Plugin</p><div class="note"><span class="notetitle">[!NOTE] Description</span><div class="notebody">In the following text, {*xxx*} represents the chip model number using the characters "910".</div></div>|
|nodeDEnable|Switch for starting the NodeD node|on|Volcano, Resilience Controller<div class="note"><span class="notetitle">[!NOTE] Description</span><div class="notebody"><ul><li>The nodeDEnable=on label enables the NodeD node status monitoring function, which is used to obtain node status Information and determine whether a node is Faulty.</li><li>A value of off or the absence of this parameter indicates that only node Information is reported, without determining whether the node is Faulty.</li><li>When using **Containerized Support** or **Resource Monitoring**, this label does not need to be configured; for other features, this label must be configured.</li></ul></div></div>|
|workerselector|Identifies the compute node of MindCluster|dls-worker-node|Ascend Device Plugin, NodeD, NPU Exporter|
|accelerator-type|Identifies the Atlas server type|<ul><li>card</li><li>module</li><li>half</li><li>module-{xxx}b-8</li><li>module-{xxx}b-16</li><li>card-{xxx}b-2</li><li>card-{xxx}b-infer</li><li>module-a3-16</li><li>module-a3-16-super-pod</li><li>module-a3-8-super-pod</li><li>350-Atlas-8</li><li>350-Atlas-16</li><li>350-Atlas-4p-8</li><li>350-Atlas-4p-16</li><li>850-Atlas-8p-8</li><li>850-SuperPod-Atlas-8</li><li>950-SuperPod-Atlas-8</li></ul>|Ascend Device Plugin, Volcano|
|servertype|Device type|<ul><li>npu-{Number of Cores}</li><li>soc</li><li>Ascend910-{Number of Cores}</li><li>Ascend310P-{Number of Cores}</li></ul>|Volcano, Ascend Device Plugin|
|<p>huawei.com/Ascend910-Recover</p><p>huawei.com/npu-Recover</p>|Fault recovery Identifier for Atlas training series products|Faulty Chip ID|Ascend Device Plugin|
|<p>huawei.com/Ascend910-NetworkRecover</p><p>huawei.com/npu-NetworkRecover</p>|Network fault recovery identifier for Atlas training series products|Faulty Chip ID|Ascend Device Plugin|
|infer-card-type|Written by Ascend Device Plugin, indicates the node's inference card Type.|card-300i-duo|Volcano|
|mind-cluster/npu-chip-memory|On-chip memory|mind-cluster/npu-chip-memory=64G|Volcano, Ascend Device Plugin|
|huawei.com/scheduler.chip1softsharedev.enable|Indicates whether the node supports the soft partitioning virtualization function|<ul><li>true</li><li>false</li></ul>|Volcano, Ascend Device Plugin<div class="note"><span class="notetitle">[!NOTE] Description</span><div class="notebody"><ul><li>The huawei.com/scheduler.chip1softsharedev.enable=true label indicates that the node supports the soft partitioning virtualization function.</li><li>The huawei.com/scheduler.chip1softsharedev.enable=false label indicates that the node does not support the soft partitioning virtualization function.</li></ul></div></div>|
|huawei.com/topotree.rackid|Identifies the rack ID of the node|Rack ID to which the node belongs|Volcano|
|huawei.com/topotree.superpodid|Identifies the super node ID of the node|Super node ID to which the node belongs|Volcano|
|huawei.com/topotree.groupid|Identifies the Pod group ID of the node|Pod group ID to which the node belongs|Volcano|
|huawei.com/topotree|Identifies the network topology tree ID of the node|Network topology tree ID to which the node belongs|Volcano|

## Pod Labels<a name="section1019341142914"></a>

**Table 5** Pod Labels used by the cluster scheduling components

| Name | Function | Value | Component |
|--|--|--|--|
| ring-controller.atlas | Identify Atlas Pod | <li>ascend-910</li><li>ascend-{xxx}b</li><li>ascend-npu</li> | Ascend Device Plugin |
| vnpu-dvpp | Mark the DVPP set for the Pod | <li>yes: This Pod uses DVPP.</li><li>no: This Pod does not use DVPP.</li><li>null: Default value. Does not care whether DVPP is used.</li> | Volcano |
| vnpu-level | Mark the level of the selected virtualization instance template | <li>low: Low configuration, default value.</li><li>high: Performance priority.</li> | Volcano |
| version | Mark the version of the Pod | String | Ascend Operator |
| volcano.sh/job-name | Mark the vcjob name corresponding to the Pod | String | Volcano |
| volcano.sh/job-namespace | Mark the vcjob namespace corresponding to the Pod | String | Volcano |
| volcano.sh/queue-name | Mark the queue name corresponding to the Pod | String | Volcano |
| volcano.sh/task-spec | Mark the job name corresponding to the Pod | String | Volcano |
| fault-type | Mark the Pod fault handling policy | <ul><li>SubHealth</li><li>Separate</li></ul> | Volcano |
| deploy-name | Mark the deployment name corresponding to the Pod | String | Ascend Operator |
| group-name | Mark the group name of the acjob corresponding to the Pod | mindxdl.gitee.com | Volcano, Ascend Operator |
| job-name | Mark the acjob name corresponding to the Pod | String | Ascend Operator |
| replica-index | Mark the Pod index (to be deleted later) | [0-{Pod Quantity-1}] | Ascend Operator |
| replica-type | Mark the Pod Type (to be deleted later) | <ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul> | Ascend Operator |
| training.kubeflow.org/job-name | Mark the acjob name corresponding to the Pod | String | Ascend Operator |
| training.kubeflow.org/job-role | Mark the Pod Type | master | Ascend Operator |
| training.kubeflow.org/operator-name | Mark the operator name that created the Pod | ascendjob-controller | Ascend Operator |
| training.kubeflow.org/replica-index | Mark the Pod index | [0-{Pod Quantity-1}] | Ascend Operator |
| training.kubeflow.org/replica-type | Mark the Pod type | <ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul> | Ascend Operator |
| super-pod-affinity | Affinity scheduling policy used by SuperPoD jobs | <ul><li>soft</li><li>hard</li></ul> | Ascend Operator, Volcano |

## Pod Annotations<a name="section16927154663513"></a>

**Table 6** Pod Annotations used by the cluster scheduling components

|Name|Function|Value|Component|
|--|--|--|--|
|<p>ascend.kubectl.kubernetes.io/ascend-910-configuration</p><p>ascend.kubectl.kubernetes.io/ascend-npu-configuration</p>|Data source for Ascend Operator to generate hccl.json|String map|Ascend Device Plugin, Ascend Operator|
|super_pod_id|Provides SuperPoD ID information for Ascend Operator|Number|Ascend Operator|
|hccl/rankIndex|Basis for retaining the original rank ID during resumable training|[0,1000]|Volcano, Ascend Operator|
|distributed-job|Marks the training job type|<ul><li>true: The current job is a distributed job</li><li>false: The current job is a single-server job</li></ul>|Volcano|
|<p>huawei.com/Ascend910</p><p>huawei.com/npu</p>|Basis for Ascend Device Plugin to allocate chips to Pods.|String|Volcano, Ascend Device Plugin|
|huawei.com/AscendReal|Record of the actual chips allocated by Ascend Device Plugin to the Pod|String|Volcano, Ascend Device Plugin|
|huawei.com/npu-core|Marks the physical ID and slicing template of the NPU card used by the Pod|String|Volcano, Ascend Device Plugin|
|huawei.com/kltDev|Record of chips allocated by kubelet to the Pod|String|Ascend Device Plugin|
|huawei.com/recover_policy_path|Job rescheduling policy|pod: Only supports Pod-level rescheduling, will not escalate to Job level (when using vcjob, this policy needs to be configured: policies: -event:PodFailed -action:RestartTask)|Volcano|
|huawei.com/schedule_minAvailable|Minimum number of replicas required for the job to be scheduled|Integer|Volcano|
|predicate-time|Basis for the order in which Ascend Device Plugin allocates chips to Pods|String|Volcano, Ascend Device Plugin|
|isSharedTor|Marks the switch attributes corresponding to the Pod|Integer|Volcano|
|isHealthy|Marks the switch status corresponding to the Pod|Integer|Volcano|
|scheduling.k8s.io/group-name|Marks the podGroup name corresponding to the Pod|String|Volcano|
|volcano.sh/job-name|Marks the vcjob name corresponding to the Pod|String|Volcano|
|volcano.sh/job-version|Marks the vcjob version corresponding to the Pod|String|Volcano|
|volcano.sh/queue-name|Marks the queue version corresponding to the Pod|String|Volcano|
|volcano.sh/task-spec|Marks the job  name corresponding to the Pod|String|Volcano|
|volcano.sh/template-uid|Marks the pod-template name corresponding to the Pod|String|Volcano|
|sharedTorIp|Marks the shared switch information used by the job |String|Volcano, ClusterD|
|fault-job-delete|Marks the rank information of the job|String|Volcano|
|mind-cluster/hardware-type=800I-A2-xx|xx indicates the on-chip memory of the current node, for example, mind-cluster/hardware-type=800I-A2-64G|String|Volcano|
|super-pod-rank|Logical SuperPoDe rank of the job |Number|Ascend Operator, Volcano|
|inHotSwitchFlow|Marks that the current Pod (faulty Pod and backup Pod) is in a hot switching process|true|ClusterD, Ascend Operator|
|backupNewPodName|Marks the name of the backup Pod created for the current faulty Pod|Corresponding backup Pod name|ClusterD, Ascend Operator|
|backupSourcePodName|Marks the original Pod name corresponding to the current backup Pod|Corresponding original Pod name|Ascend Operator|
|needOperatorOpe|Marks that the current Pod needs to be processed by Ascend Operator|<ul><li>create: Ascend Operator needs to create a backup Pod based on the current Pod</li><li>delete: Ascend Operator needs to delete the current Pod</li></ul>|ClusterD, Ascend Operator|
|needVolcanoOpe|Marks that the current Pod needs to be processed by Volcano|delete: Volcano needs to delete the current Pod|ClusterD, Volcano|
|podType|Marks that the current Pod is a backup Pod|backup|ClusterD, Ascend Operator|
|huawei.com/scheduler.softShareDev.aicoreQuota|Marks the percentage of AICore required by the current Pod.|[1, 100]|Volcano, Ascend Device Plugin|
|huawei.com/scheduler.softShareDev.hbmQuota|Marks the amount of high-bandwidth memory required by the current Pod.|<p>[1, maxHBM]</p><p>maxHBM is the HBM value in HBM-Usage(MB) queried using the <b>npu-smi info</b> command.</p>|Volcano, Ascend Device Plugin|
|huawei.com/scheduler.softShareDev.policy|Marks the policy of the soft partitioning job executed by the current Pod.|<ul><li>fixed-share</li><li>elastic</li><li>best-effort</li></ul>|Volcano, Ascend Device Plugin|
|huawei.com/affinity-config|Configures the affinity level for multi-level scheduling of the job.|<p>level1=x,level2=y,...</p><p>Where x, y... are the sub-job sizes for the corresponding network levels.</p><p>This field is used to configure the affinity level for multi-level scheduling of the job.</p><p>It must be a concatenation of strings in the format leveli=ni, separated by commas. Here, i is the network level sequence number, and ni is the number of replicas for the sub-job at that network level. For example, for a job with a total of 8 replicas, "level1=2,level2=4" means that every 2 Pods in the job are assigned to nodes with the same level1 label, and every 4 Pods are assigned to nodes with the same level2 label.</p><p>The network level configuration must meet the following requirements:<ul><li>When the job has more than one level, the value of level n must be an integer multiple of n-1.</li><li>The total number of job replicas must be an integer multiple of all levels.</li><li>The job level configuration must start from level1 and be consecutive in ascending order.</li></ul></p>|Volcano|
|huawei.com/schedule_policy|Specifies the scheduling policy.|Currently supports the configurations in [Table 3 huawei.com/schedule_policy Configuration Description](./volcano.md#podgroup).|Volcano|

## Node Annotations<a name="section9144358124519"></a>

**Table 7** Node annotations used by the cluster scheduling components

|Name|Function|Value|Component|
|--|--|--|--|
|baseDeviceInfos|Displays basic chip information, such as IP, for use during Volcano scheduling.|String|Volcano|
|product-serial-number|NodeD obtains the node SN through the IPMI and writes it into the annotation for use when ClusterD receives a common fault.|String|ClusterD|
|superPodID|Indicates the ID of the SuperPoD to which this node belongs.|String|ClusterD|
|ResetInfo|Displays information about chips that failed automatic reset by the Ascend Device Plugin, such as the chip's physical ID, Card ID, etc.|String|Ascend Device Plugin|

The content format of ResetInfo is as follows.

```json
{
    "ThirdPartyResetDevs": [
        {
            "CardId": 0,
            "DeviceId": 0,
            "AssociatedCardId": 4,
            "PhyID": 0,
            "LogicID": 0
        }
    ],
    "ManualResetDevs": [
        {
            "CardId": 1,
            "DeviceId": 0,
            "AssociatedCardId": 5,
            "PhyID": 2,
            "LogicID": 2
        }
    ]
}
```

## K8s ServiceAccount<a name="section168254015405"></a>

**Table 8** List of ServiceAccounts created by components in K8s

|Account Name|Description|
|--|--|
|volcano-controllers|User created in K8s by the controller component of open-source Volcano.|
|volcano-scheduler|User created in K8s by the scheduler component of open-source Volcano.|
|<p>ascend-device-plugin-sa-npu</p><p>ascend-device-plugin-sa-910</p><p>ascend-device-plugin-sa-310p</p><p>ascend-device-plugin-sa-310</p>|When starting the service using YAML, this user will be created in K8s. The Account Name used varies for different device models.|
|ascend-operator-manager|When starting the service using YAML, this user will be created in K8s, for example: ascend-operator-v{version}.yaml.|
|resilience-controller|It is recommended to start with security hardening. Use the YAML with `without-token` to start the service, create and use the resilience-controller account in K8s, and grant appropriate permissions to this account.|
|noded|When starting the service using YAML, this user will be created in K8s, for example: noded-v{version}.yaml.|
|clusterd|When starting the service using YAML, this user will be created in K8s, for example: clusterd-v{version}.yaml.|
|default|User automatically created in K8s when deploying MindCluster components or open-source Volcano.|
