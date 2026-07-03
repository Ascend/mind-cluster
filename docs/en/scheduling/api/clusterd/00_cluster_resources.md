# Cluster Resources<a name="ZH-CN_TOPIC_0000002511346785"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:45:45.765Z pushedAt=2026-06-09T02:05:50.682Z -->

## ConfigMap Description<a name="section17868183824213"></a>

After ClusterD starts, the following ConfigMaps are created:

- `cluster-info-node-cm`. For details, see [Table 1](#table25031946405).
- `cluster-info-device-${m}`. For details, see [Table 2](#table915714719368). `m` is an integer that increments from 0. For every additional 1,000 nodes in a cluster, a new ConfigMap file of this type is created.
- `cluster-info-switch-${x}`. For details, see [Table 3](#table9246232250). `x` is an integer that increments from 0. For every additional 2,000 nodes in a cluster, a new ConfigMap file of this type is added.

**Table 1** cluster-info-node-cm

<a name="table25031946405"></a>

|Parameter|Description|
|--|--|
|mindx-dl-nodeinfo-kwok-node-0|The prefix is fixed as `mindx-dl-nodeinfo`, and `kwok-node-0` is the node name, which facilitates locating the specific node where a fault occurs.|
|NodeInfo|Node fault information.|
|FaultDevList|List of faulty devices on a node.|
|- DeviceType|Type of the faulty device.|
|- DeviceId|ID of the faulty device.|
|- FaultCode|Fault code, a hexadecimal string composed of English letters and numbers.|
|- FaultLevel|Fault handling level.<ul><li>`NotHandleFault`: No handling required.</li><li>`PreSeparateFault`: If there are jobs on this node, no handling is performed; subsequent scheduling will not assign jobs to this node.</li><li>`SeparateFault`: Job rescheduling.</li></ul>|
|NodeStatus|Node health status, determined by the device with the most severe fault handling level on this node.<ul><li>`Healthy`: The fault handling level of this node exists and is not higher than `NotHandleFault`, the node is considered healthy and can participate in normal training. If the fault handling level of this node is `PreSeparateFault` and the node has NPUs in use, the node is also considered healthy. However, after the job completes, the node will become faulty.</li><li>`UnHealthy`: If the fault handling level of this node includes `SeparateFault`,the node is considered faulty and will affect training jobs. Jobs will be immediately evicted from this node. If the node's fault handling level is `PreSeparateFault` and no NPUs are currently in use on the node, the node is also considered faulty and jobs must not be scheduled to it.</li></ul>|

**Table 2** cluster-info-device-${m}

<a name="table915714719368"></a>

|Parameter|Description|
|--|--|
|mindx-dl-deviceinfo-kwok-node-0|The prefix is fixed as `mindx-dl-deviceinfo`, and `kwok-node-0` is the node name, used to locate the specific node where a fault occurs.|
|huawei.com/Ascend910|<ul><li>Name information of the chips available on the current node. When there are multiple chips, they are concatenated with commas.</li><li>Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use `huawei.com/npu` as the parameter.</li></ul>|
|huawei.com/Ascend910-NetworkUnhealthy|<ul><li>Name information of the chips with unhealthy network on the current node. When there are multiple chips, they are concatenated with commas.</li><li>Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use `huawei.com/npu-NetworkUnhealthy` as the parameter.</li></ul>|
|huawei.com/Ascend910-Unhealthy|<ul><li>Name information of the unhealthy chips on the current node. When there are multiple chips, they are concatenated with commas.</li><li>Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use `huawei.com/npu-Unhealthy` as the parameter.</li></ul>|
|huawei.com/Ascend910-Fault|<ul><li>Array object. The object contains the `fault_type`, `npu_name`, `large_model_fault_level`, `fault_level`, `fault_handling`, `fault_code`, and `fault_time_and_level_map` fields.</li><li>Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use `huawei.com/npu-Fault` as the parameter.</li></ul>|
|- fault_type|Fault type.<ul><li>`CardUnhealthy`: chip fault</li><li>`CardNetworkUnhealthy`: parameter plane network fault (chip network-related fault)</li><li>`NodeUnhealthy`: node fault</li><li>`PublicFault`: common fault</li></ul>|
|- npu_name|Name of the faulty chip; null when a node fault occurs.|
|<p>- large_model_fault_level</p><p>- fault_level</p><p>- fault_handling</p>|Fault handling type. The value is empty when a node fault occurs.<ul><li>`NotHandleFault`: No handling is performed.</li><li>`RestartRequest`: In inference scenarios, the inference request needs to be re-executed. In training scenarios, the training service needs to be re-executed.</li><li>`RestartBusiness`: The service needs to be re-executed.</li><li>`FreeRestartNPU`: Service execution is affected. The chip needs to be reset when it is idle.</li><li>`RestartNPU`: Directly reset the chip and re-execute the service.</li><li>`SeparateNPU`: Isolate the chip.</li><li>`PreSeparateNPU`: Pre-isolate the chip. Whether to reschedule will be determined based on the actual running status of the training job.</li><li>`ManuallySeparateNPU`: Manually isolate the chip. When the respective fault frequencies of Ascend Device Plugin and ClusterD are reached, Ascend Device Plugin and ClusterD will manually isolate the faulty chip.</li></ul><div class="note"><span class="notetitle">**NOTE**</span><div class="notebody"><ul><li>The `large_model_fault_level`, `fault_handling`, and `fault_level` parameters have the same function. It is recommended to use `fault_handling`.</li><li>If the inference job subscribes to fault information, and a `RestartRequest` fault occurs on the inference card used by the job and the fault duration does not exceed 60 seconds, job rescheduling will not be performed. If the fault duration exceeds 60 seconds and is not recovered, the chip will be isolated and job rescheduling will be performed.</li></ul></div></div>|
|- fault_code|Fault code, a string concatenated with commas.|
|- fault_time_and_level_map|Fault code, fault occurrence time, and fault handling level.|
|UpdateTime|Update time of the current node information, in timestamp format, used to identify the latest reporting time of the fault information or device status.|
|CmName|ConfigMap name of the configuration corresponding to this node in the cluster.|
|SuperPodID|SuperPoD ID.|
|RackID|Rack ID.|
|ServerIndex|The relative position of the current node in the SuperPoD.<ul><li>When the value of `SuperPodID` or `ServerIndex` reported by the driver is `0xffffffff`, the value of `SuperPodID` or `ServerIndex` is `-1`.</li><li>The value of `SuperPodID` or `ServerIndex` is `-2` in the following cases.</li><ul><li>The current device does not support querying SuperPoD information.</li><li>Failed to obtain SuperPoD information due to a driver issue.</li></ul></ul>|

**Table 3**  cluster-info-switch-${x}

<a name="table9246232250"></a>

|Parameter|Description|
|--|--|
|FaultCode|List of UnifiedBus device fault codes for the current node. The array object contains fields such as `EventType`, `AssembledFaultCode`, `PeerPortDevice`, `PeerPortId`, `SwitchChipId`, `SwitchPortId`, `Severity`, `Assertion`, and `AlarmRaisedTime`.|
|-EventType|Alarm ID.|
|-AssembledFaultCode|Fault code.|
|-PeerPortDevice|Peer device type.<ul><li>0: CPU</li><li>1: NPU</li><li>2: SW</li><li>0xFFFF: NA</li></ul>|
|-PeerPortId|Peer device ID.|
|-SwitchChipId|UnifiedBus fault chip ID, starting from 0.|
|-SwitchPortId|UnifiedBus fault port ID, starting from 0.|
|-Severity|Fault level.<ul><li>0: Info</li><li>1: Minor</li><li>2: Major</li><li>3: Critical</li></ul>|
|-Assertion|Event type.<ul><li>0: Fault recovery</li><li>1: Fault occurrence</li><li>2: Notification event</li></ul>|
|-AlarmRaisedTime|Time when the fault/event occurred.|
|FaultLevel|Fault handling level of the current node.<p>Takes the highest fault level among all faults in `FaultCode`. Values include: `NotHandle`, `SubHealthFault`, `Separate`, and `RestartRequest`.</p>|
|UpdateTime|Time when the fault report was refreshed.|
|NodeStatus|Health status of the current node.<p>Corresponds to the `FaultLevel` value: `NotHandle:Healthy`, `SubHealthFault:SubHealthy`, `Separate:UnHealthy`, and `RestartRequest:UnHealthy`.</p>|
|FaultTimeAndLevelMap|List of fault occurrence times and fault handling levels. The array object contains fields for fault code, UnifiedBus fault chip ID, UnifiedBus fault port ID, fault_time, and fault_level. The key is composed of the fault code, UnifiedBus fault chip ID, and UnifiedBus fault port ID, connected by underscores.|
|-fault_time|Time when the fault occurred.|
|-fault_level|Fault handling level.|

## statistic-fault-info<a name="section1153232554520"></a>

This ConfigMap is located in the user-created `cluster-system` namespace, with the label `mc-statistic-fault=true`. It is used to display fault information in a cluster (currently only common fault information is displayed).

**Table 4** Data information

|Parameter|Description|
|--|--|
|PublicFaults|Details of common faults. When the number of faults is too large, this field will no longer be updated. For details about the following fields, see [Fault Information Description Table](./03_public_fault_apis.md#configmap).|
|-node_name|Name of the faulty node|
|-resource|Fault sender<p>The default configuration includes CCAE, fd-online, pingmesh, and Netmind.</p>|
|-devIds|Physical ID of the faulty chip|
|-faultId|Fault instance ID|
|-type|Fault type<ul><li>NPU: chip fault.</li><li>Node: node fault.</li><li>Network: network fault.</li><li>Storage: storage fault.</li></ul>|
|-faultCode|Fault code|
|-level|Fault level<ul><li>`NotHandleFault`: Not handled for now.</li><li>`SubHealthFault`: Sub-health.</li><li>`SeparateNPU`: Unrecoverable; the chip needs to be isolated.</li><li>`PreSeparateNPU`: Does not affect services for now, and no more jobs will be scheduled to this chip.</li></ul>|
|-faultTime|Fault occurrence time|
|FaultNum|Number of faults|
|-publicFaultNum|Sum of common faults across all nodes.|
|Description|Prompt information when the number of common faults is too large.|

>[!NOTE]
>Common faults display 1 MB of data externally, approximately 4,500 entries. When the number exceeds 4,500, some data will no longer be displayed externally, and a `Description` will be added to the ConfigMap as a prompt, while the internal cache continues to operate normally.

## super-pod-<super-pod-id\><a name="section53741611135414"></a>

This ConfigMap is located in the user-created `cluster-system` namespace, with the label `app=pingmesh`.

**Table 5** super-pod-<super-pod-id\>

|Parameter|Description|
|--|--|
|app|Label key required by NodeD to identify the ConfigMap. The value is pingmesh.|
|superPodDevice|Key for SuperPoD information.|
|SuperPodID|SuperPoD ID|
|NodeDeviceMap|Information about all nodes contained in the SuperPoD.|
|NodeName|Node name|
|DeviceMap|Information about all NPUs in the node, in the format of `physicID:superDeviceID`.|

## fault-job-info<a name="section1548342116513"></a>

This ConfigMap is located in the `cluster-system` namespace created by the user. It is used to display fault job information that requires forced release of communication resources in the cluster. It takes effect only when process-level rescheduling is performed on the Atlas 900 A3 SuperPoD.

**Table 6** fault-job-info

|Parameter|Description|Value|
|--|--|--|
|SdIds|SDID of the faulty card.|String |
|NodeNames|Name of the node whose resources need to be forcibly released.|String|
|FaultTimes|Time when the fault occurred.|64-bit integer|
|JobId|UID of the job.|String|

## clusterd-manual-info-cm<a name="section15483421165190"></a>

This ConfigMap is located in the `cluster-system` namespace created by the user. It is used to display the chips and fault information of manual isolation in the cluster.

The following is an example:

```json
Name:         clusterd-manual-info-cm
Namespace:    cluster-system
Labels:       <none>
Annotations:  <none>

Data
====
localhost.localdomain:
----
{"Total":["Ascend910-0","Ascend910-2","Ascend910-3"],"Detail":{"Ascend910-0":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-2":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-3":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}]}}

Events:  <none>
```

**Table 7**  clusterd-manual-info-cm

|Parameter|Description|
|--|--|
|<i>localhost.localdomain</i>|Node name, for example, `localhost.localdomain`.|
|Total|Name of the faulty chip.|
|Detail|Chip fault information.|
|-<i>Ascend910-0</i>|Chip name, for example, `Ascend910-0`.|
|-FaultCode|Fault code.|
|-FaultLevel|Fault level.|
|-LastSeparateTime|Time of the last fault when the manual isolation frequency is reached. If a fault that has triggered manual chip isolation reaches the manual isolation frequency again, this time will be refreshed.|
