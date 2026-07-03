# Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426737"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:42:28.463Z pushedAt=2026-06-09T02:05:50.632Z -->

## Chip Resources<a name="ZH-CN_TOPIC_0000002511346781"></a>

**mindx-dl-deviceinfo-<nodename\><a name="section11555858123711"></a>**

The NPU information reported by Ascend Device Plugin is shown in [Table 1](#table13817185391117).

**Table 1**  DeviceInfoCfg

<a name="table13817185391117"></a>

|Name|Meaning|Description|
|--|--|--|
|huawei.com/Ascend910|Marks the chip name information available on the current node. When multiple chips exist, they are concatenated with commas. |<ul><li>This field is currently being phased out and will no longer be presented in subsequent versions. By default, the available chips on a node are maintained by Volcano, and this field does not take effect. To make it take effect, you can modify the Volcano configuration parameter "self-maintain-available-card" to false.</li><li>Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use huawei.com/npu as the parameter name.</li></ul>|
|huawei.com/Ascend910-NetworkUnhealthy|Marks the chip name information for chips with unhealthy networks on the current node. When multiple chips exist, they are concatenated with commas. |Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use huawei.com/npu-NetworkUnhealthy as the parameter name.|
|huawei.com/Ascend910-Unhealthy|Marks the chip name information for unhealthy chips on the current node. When multiple chips exist, they are concatenated with commas. |Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use huawei.com/npu-Unhealthy as the parameter name.|
|huawei.com/Ascend910-Recovering|Marks the chips that are currently being recovered on the current node. When multiple chips exist, they are concatenated with commas. |Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use huawei.com/npu-Recovering as the parameter name.|
|huawei.com/Ascend910-Fault|Records specific fault information of the chip. |<ul><li>Array object. The object contains seven fields: fault_type, npu_name, large_model_fault_level, fault_level, fault_handling, fault_code, and fault_time_and_level_map.</li><li>Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD use huawei.com/npu-Fault as the parameter name.</li></ul>|
|-fault_type|Fault type. |<ul><li>CardUnhealthy: Chip fault</li><li>CardNetworkUnhealthy: Chip network fault</li><li>NodeUnhealthy: Node fault</li></ul>|
|-npu_name|Name of the faulty chip. It is empty for a node fault. |String|
|<p>-large_model_fault_level</p><p>-fault_level</p><p>-fault_handling</p>|Fault handling type. The value is empty for a node fault. |<ul><li>NotHandleFault: No handling</li><li>RestartRequest: In inference scenarios, the inference request needs to be re-executed. In training scenarios, the training workload needs to be re-executed.</li><li>RestartBusiness: The workload needs to be re-executed.</li><li>FreeRestartNPU: Affects workload execution. The chip needs to be reset when it is idle.</li><li>RestartNPU: Directly reset the chip and re-execute the workload.</li><li>SeparateNPU: Isolate chip</li><li>PreSeparateNPU: Pre-isolate chip. Whether to reschedule is determined based on the actual running status of the training task.</li></ul><div class="note"><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><ul><li>The large_model_fault_level, fault_level, and fault_handling parameters have the same function. It is recommended to use fault_handling.</li><li>If an inference task subscribes to fault information, and a RestartRequest fault occurs on the inference card used by the task, and the fault duration does not exceed 60 seconds, task rescheduling is not performed. If the fault duration exceeds 60 seconds and the fault is not recovered, the chip is isolated and task rescheduling is performed.</li></ul></div></div>|
|-fault_code|Fault code, a string concatenated with commas. |For a detailed description of chip fault codes, see [Chip Fault Code References](../references/appendix.md#chip-fault-code-references).|
|-fault_time_and_level_map|Fault code, fault generation time, and fault handling level. |-|
|SuperPodID|SuperPod ID. |String|
|ServerIndex|The relative position of the current node within the SuperPod. |<ul><li>When the SuperPodID or ServerIndex value reported by the driver is 0xffffffff, the value of SuperPodID or ServerIndex is -1.</li><li>The value of SuperPodID or ServerIndex is -2 in the following cases.<ul><li>The current device does not support querying SuperPod information.</li><li>Failed to obtain SuperPod information due to a driver issue.</li></ul></li></ul>|
|CheckCode|Check code. |-|

The UnifiedBus device fault information reported by Ascend Device Plugin is shown in [Table 2](#table13455135662318).

**Table 2**  SwitchInfoCfg

<a name="table13455135662318"></a>

|Name|Meaning|Description|
|--|--|--|
|FaultCode|List of UnifiedBus device fault codes for the current node.|Array object, containing fields such as EventType, AssembledFaultCode, PeerPortDevice, PeerPortId, SwitchChipId, SwitchPortId, Severity, Assertion, and AlarmRaisedTime.|
|-EventType|Alarm ID.|-|
|-AssembledFaultCode|Fault code.|-|
|-PeerPortDevice|Peer device type.|<ul><li>0: CPU</li><li>1: NPU</li><li>2: SW</li><li>0xFFFF: NA</li></ul>|
|-PeerPortId|Peer device ID.|-|
|-SwitchChipId|UnifiedBus device fault chip ID.|Numbering starts from 0.|
|-SwitchPortId|UnifiedBus device fault port ID.|Numbering starts from 0.|
|-Severity|Fault level.|<ul><li>0: Info</li><li>1: Minor</li><li>2: Major</li><li>3: Critical</li></ul>|
|-Assertion|Event type.|<ul><li>0: Fault recovery</li><li>1: Fault generation</li><li>2: Notification event</li></ul>|
|FaultLevel|Fault handling level of the current node.|Takes the highest fault level among all faults in FaultCode. Values include: NotHandle, SubHealthFault, Separate, and RestartRequest.|
|UpdateTime|Fault report refresh time.|-|
|NodeStatus|Health status of the current node.|Corresponds to FaultLevel values: NotHandle:Healthy, SubHealthFault:SubHealthy, Separate:UnHealthy, and RestartRequest:UnHealthy.|
|FaultTimeAndLevelMap|List of fault occurrence times and fault handling levels.|Array object, containing fields for fault code, UnifiedBus device fault chip ID, UnifiedBus device fault port ID, fault_time, and fault_level. The key is composed of the fault code, UnifiedBus device fault chip ID, and UnifiedBus device fault port ID, connected by underscores.|
|-fault_time|Fault occurrence time.|-|
|-fault_level|Fault handling level.|-|

The manually intervened fault-level chip information reported by the ConfigMap of Ascend Device Plugin is shown in [Table 3](#table9710232).

**Table 3** ManuallySeparateNPU

<a name="table9710232"></a>

|Name|Meaning|Description|
|--|--|--|
|ManuallySeparateNPU|The chip is recorded in this key by the ConfigMap because multiple chip faults have triggered the frequency-based fault escalation policy.|Multiple chip names are separated by commas (,).|

The fault policy escalation reasons reported by the ConfigMap of Ascend Device Plugin are shown in [Table 4](#table9710233).

**Table 4** UpgradeFaultReason

<a name="table9710233"></a>

|Name|Meaning|Description|
|--|--|--|
|UpgradeFaultReason|After a fault code is configured with frequency-based and duration-based policies, when fault escalation is triggered, this records the reason for the fault escalation and the escalation time.|JSON Map format, where the key is the chip name and the value is the reason that caused the chip fault escalation.|
|-fault_code|Fault code for chip fault escalation.|-|
|-fault_level|Fault level after escalation.|-|
|-upgrade_type|Fault escalation type.|<ul><li>Frequency-based escalation: FaultFrequency</li><li>Duration-based escalation: FaultDuration</li><li>Autofill escalation: FaultAutofill</li></ul>|
|-upgrade_time|Time point of fault escalation.|-|

>[!NOTE]
>
>- When Ascend Device Plugin is upgraded from a version earlier than 26.0.0 to 26.0.0 or later, if an existing chip fault in the ConfigMap of Ascend Device Plugin is escalated to ManuallySeparateNPU, a reason for isolating that chip will be automatically filled. The -fault_code value will be AutofillFaultCode, and the -upgrade_type will be FaultAutofill.
>- The fault escalation reason is deleted along with fault downgrade, and the deletion event is recorded in the event events under the kube-system namespace of K8s.

For the description information in the ConfigMap of Ascend Device Plugin, see [Table 5](#table97108314503).

**Table 5**  Description

<a name="table97108314503"></a>

|Name|Meaning|Description|
|--|--|--|
|Description|Description information.|The available chip information of the node in this ConfigMap is being phased out. By default, the available chips of the node are maintained by Volcano, and the information maintained in this ConfigMap does not take effect. If it needs to take effect, you can modify the Volcano configuration parameter "self-maintain-available-card" to false.|

The NPU device fault information reported by Ascend Device Plugin is shown in [Table 6](#table68216761214). The object name is `<device-plugin-pod-name>.<reporting-time><fault- chip-ID>`, and the object type is `Event`.

>[!NOTE]
>The following table only shows the field descriptions related to MindCluster services. For details about more fields, see [Event core](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#event-v1-core).

**Table 6**  NPU device fault information

<a name="table68216761214"></a>

|Name|Meaning|Note|
|--|--|--|
|type|Event level.|Unique value: Warning|
|message|Event content, including node name, chip ID, fault generation or recovery type, fault code, and fault level information.|String|
|reason|Reason for event reporting.|<ul><li>Recovery: Fault recovery</li><li>Occur: Fault generation</li><li>Notice: One-time fault notification</li></ul>|
|action|Fault level.|String. For details, see [Table 1](#custom-chip-faults).|
|source|Source of the fault.|Struct. Indicates the node where the fault occurred.|
|eventTime|Time when the fault occurred.|Timestamp|
|involvedObject|Object to which the fault is bound for display.|Struct. Points to the Pod name of the current Ascend Device Plugin through Kind, Namespace, and Name. After specification, the event can be viewed not only through the Event object query but also when viewing the details of the current Pod.|
|reportingComponent|Event controller.|Unique value: device-plugin|
|reportingInstance|Event reporting instance.|String. Takes the Pod name of the current Ascend Device Plugin.|

**deviceNameCustomization.json<a name="section579455712489"></a>**

`deviceNameCustomization.json` supports customizing device names. When building the Ascend Device Plugin image, place this file in the same directory as the binary package to change the resource type and resource name displayed by Ascend Device Plugin to custom names. Atlas 350 PCIe card, Atlas 850 series hardware products, and Atlas 950 SuperPoD do not currently support this feature.

**Table 7**  Custom device names supported by deviceNameCustomization.json

<a name="table76189511522121"></a>

|Name|Meaning|Note|
|--|--|--|
|ResourceType|Initial name of the device. Mandatory.|Only Ascend910, Ascend310, or Ascend310P is supported.|
|DevicePublicType|Type displayed externally for the device, for example, huawei.com/Ascend910. Mandatory.|Only the xxx.xxx/xxx format is supported. xxx can be uppercase and lowercase letters and digits, with a length range of 10 to 32 characters.|
|DevicePublicNamePre|Prefix of the device name displayed externally, for example, Ascend910-. For the actual displayed name, Ascend Device Plugin appends the physical ID of the chip after the prefix. Mandatory.|Can contain uppercase and lowercase letters, hyphens (-), and digits. Must start with an uppercase or lowercase letter. Length range: 2 to 16 characters.|
|PodConfigurationName|Details of the mounted chip information displayed in the Pod annotation. Mandatory when ResourceType is Ascend910.|Can contain uppercase and lowercase letters, hyphens (-), slashes (/), dots (.), and digits. Must start with an uppercase or lowercase letter and end with an uppercase or lowercase letter or digit. Length range: 10 to 63 characters.|

## Job Information<a name="ZH-CN_TOPIC_0000002479226860"></a>

**fault-config-job-name <a name="section1786481083812"></a>**

**Table 1**  fault-config-job-name

<a name="table68216761214"></a>

|Field |Description|Value|Remarks|
|--|--|--|--|
|fault-npus|Rank information of the faulty chip in a failed job.|String|-|
|checkCode|Check code.|String|-|

**reset-config-job-name <a name="section3394547123916"></a>**

**Table 2**  reset-config-job-name

<a name="table1213115712136"></a>
<table><thead align="left"><tr id="row3132772132"><th class="cellrowborder" valign="top" width="15.950000000000001%" id="mcps1.2.6.1.1"><p id="p1022487193411"><a name="p1022487193411"></a><a name="p1022487193411"></a>Field</p>
</th>
<th class="cellrowborder" valign="top" width="14.69%" id="mcps1.2.6.1.2"><p id="p1313212741314"><a name="p1313212741314"></a><a name="p1313212741314"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="24.23%" id="mcps1.2.6.1.3"><p id="p513317151314"><a name="p513317151314"></a><a name="p513317151314"></a>Meaning</p>
</th>
<th class="cellrowborder" valign="top" width="28.82%" id="mcps1.2.6.1.4"><p id="p313315721314"><a name="p313315721314"></a><a name="p313315721314"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="16.31%" id="mcps1.2.6.1.5"><p id="p1313327191318"><a name="p1313327191318"></a><a name="p1313327191318"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row41336711317"><td class="cellrowborder" rowspan="13" valign="top" width="15.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p20565164533410"><a name="p20565164533410"></a><a name="p20565164533410"></a>reset.json</p>
<p id="p111446396589"><a name="p111446396589"></a><a name="p111446396589"></a></p>
<p id="p1811413311215"><a name="p1811413311215"></a><a name="p1811413311215"></a></p>
<p id="p0452951162310"><a name="p0452951162310"></a><a name="p0452951162310"></a></p>
</td>
<td class="cellrowborder" valign="top" width="14.69%" headers="mcps1.2.6.1.2 "><p id="p813420781315"><a name="p813420781315"></a><a name="p813420781315"></a>RankList</p>
</td>
<td class="cellrowborder" valign="top" width="24.23%" headers="mcps1.2.6.1.3 "><p id="p121346712134"><a name="p121346712134"></a><a name="p121346712134"></a>Chip list</p>
</td>
<td class="cellrowborder" valign="top" width="28.82%" headers="mcps1.2.6.1.4 "><p id="p5134137121315"><a name="p5134137121315"></a><a name="p5134137121315"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="16.31%" headers="mcps1.2.6.1.5 "><p id="p1513427131320"><a name="p1513427131320"></a><a name="p1513427131320"></a>-</p>
</td>
</tr>
<tr id="row21341174135"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171346791316"><a name="p171346791316"></a><a name="p171346791316"></a>-RankId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p3134177131313"><a name="p3134177131313"></a><a name="p3134177131313"></a>Rank information used by the failed job</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1413587161310"><a name="p1413587161310"></a><a name="p1413587161310"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1413511721318"><a name="p1413511721318"></a><a name="p1413511721318"></a>-</p>
</td>
</tr>
<tr id="row1713512717138"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p161352712139"><a name="p161352712139"></a><a name="p161352712139"></a>-LogicId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1135127181319"><a name="p1135127181319"></a><a name="p1135127181319"></a>Chip logic ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15135157131311"><a name="p15135157131311"></a><a name="p15135157131311"></a>32-bit integer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p181366715137"><a name="p181366715137"></a><a name="p181366715137"></a>-</p>
</td>
</tr>
<tr id="row013914719136"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1313927191317"><a name="p1313927191317"></a><a name="p1313927191317"></a>-Status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8139177171315"><a name="p8139177171315"></a><a name="p8139177171315"></a>Chip status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul8436530113"></a><a name="ul8436530113"></a><ul id="ul8436530113"><li>unrecovered: Not recovered</li><li>recovered: Recovery succeeded</li><li>failed: Recovery failed</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11394791316"><a name="p11394791316"></a><a name="p11394791316"></a>-</p>
</td>
</tr>
<tr id="row814016761315"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1814015719134"><a name="p1814015719134"></a><a name="p1814015719134"></a>-Policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11140676132"><a name="p11140676132"></a><a name="p11140676132"></a>Hot reset policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1156918243817"></a><a name="ul1156918243817"></a><ul id="ul1156918243817"><li>empty: No fault</li><li>ignore: Ignore the fault</li><li>restart_request: Re-execute the current request</li><li>restart: Re-execute the training task</li><li>free_reset: When no job is running on the NPU, the device needs to be reset</li><li>reset: The device needs to be reset</li><li>isolate: The device needs to be isolated</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11140672134"><a name="p11140672134"></a><a name="p11140672134"></a>-</p>
</td>
</tr>
<tr id="row151401717139"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p101413711136"><a name="p101413711136"></a><a name="p101413711136"></a>-InitialPolicy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12141176132"><a name="p12141176132"></a><a name="p12141176132"></a>Initial hot reset policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul16378161213281"></a><a name="ul16378161213281"></a><ul id="ul16378161213281"><li>empty: No fault</li><li>ignore: Ignore the fault</li><li>restart_request: Re-execute the current request</li><li>restart: Re-execute the training job</li><li>free_reset: When no job is running on the NPU, the device needs to be reset</li><li>reset: The device needs to be reset</li><li>isolate: The device needs to be isolated</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p171419712133"><a name="p171419712133"></a><a name="p171419712133"></a>-</p>
</td>
</tr>
<tr id="row2141187121312"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3141197161312"><a name="p3141197161312"></a><a name="p3141197161312"></a>-ErrorCode</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19141576138"><a name="p19141576138"></a><a name="p19141576138"></a>Decimal fault code</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p151429710139"><a name="p151429710139"></a><a name="p151429710139"></a>64-bit integer array</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1314257131311"><a name="p1314257131311"></a><a name="p1314257131311"></a>-</p>
</td>
</tr>
<tr id="row14142137191314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171421973132"><a name="p171421973132"></a><a name="p171421973132"></a>-ErrorCodeHex</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p0142577133"><a name="p0142577133"></a><a name="p0142577133"></a>Hexadecimal fault code</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p8142177131320"><a name="p8142177131320"></a><a name="p8142177131320"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p31421070133"><a name="p31421070133"></a><a name="p31421070133"></a>-</p>
</td>
</tr>
<tr id="row41431139195820"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3233191110537"><a name="p3233191110537"></a><a name="p3233191110537"></a>GracefulExit</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p321511543920"><a name="p321511543920"></a><a name="p321511543920"></a>Manage training processes</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p33363655012"><a name="p33363655012"></a><a name="p33363655012"></a>0 or 1</p>
<a name="ul7532185975011"></a><a name="ul7532185975011"></a><ul id="ul7532185975011"><li>Value 1: Terminate all training processes</li><li>Value 0: No action</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p921615511390"><a name="p921615511390"></a><a name="p921615511390"></a>-</p>
</td>
</tr>
<tr id="row167775084714"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p108353401829"><a name="p108353401829"></a><a name="p108353401829"></a>UpdateTime</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p118356401224"><a name="p118356401224"></a><a name="p118356401224"></a>ConfigMap update time</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p58359402214"><a name="p58359402214"></a><a name="p58359402214"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p7835114017213"><a name="p7835114017213"></a><a name="p7835114017213"></a>-</p>
</td>
</tr>
<tr id="row189371153471"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p2066862744"><a name="p2066862744"></a><a name="p2066862744"></a>RetryTime</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p126681521149"><a name="p126681521149"></a><a name="p126681521149"></a>Number of Pod rescheduling attempts</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p66683214418"><a name="p66683214418"></a><a name="p66683214418"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p18668142844"><a name="p18668142844"></a><a name="p18668142844"></a>-</p>
</td>
</tr>
<tr id="row13113203322"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1254115251666"><a name="p1254115251666"></a><a name="p1254115251666"></a>FaultFlushing</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7541192512618"><a name="p7541192512618"></a><a name="p7541192512618"></a>Informs <span id="ph14256162281217"><a name="ph14256162281217"></a><a name="ph14256162281217"></a>Elastic Agent</span> whether a fault is currently being flushed</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13813147101216"><a name="p13813147101216"></a><a name="p13813147101216"></a>Value is true or false</p>
<a name="ul1563191521213"></a><a name="ul1563191521213"></a><ul id="ul1563191521213"><li>true: Indicates a fault is being flushed</li><li>false: Indicates no fault is being flushed currently</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p19951631131314"><a name="p19951631131314"></a><a name="p19951631131314"></a><span id="ph952618296564"><a name="ph952618296564"></a><a name="ph952618296564"></a>Elastic Agent</span> needs to wait until this field is false and the faulty RankList contains no faults for this node before starting the training process</p>
</td>
</tr>
<tr id="row18452151202319"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p64521951162319"><a name="p64521951162319"></a><a name="p64521951162319"></a><span>RestartFaultProcess</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17453851172311"><a name="p17453851172311"></a><a name="p17453851172311"></a><span>Informs </span><span id="ph262783362516"><a name="ph262783362516"></a><a name="ph262783362516"></a>Elastic Agent</span><span> whether to restart only the faulty process on this node</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2012431813258"><a name="p2012431813258"></a><a name="p2012431813258"></a><span>Value is true or false</span></p>
<a name="ul5650162018256"></a><a name="ul5650162018256"></a><ul id="ul5650162018256"><li><span>true: Indicates not to exit </span><span id="ph8849103812259"><a name="ph8849103812259"></a><a name="ph8849103812259"></a>Elastic Agent</span><span>, only restart the faulty process on this node</span></li><li><span>false: When this node has a faulty process, exit </span><span id="ph1888614312613"><a name="ph1888614312613"></a><a name="ph1888614312613"></a>Elastic Agent</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p94534513233"><a name="p94534513233"></a><a name="p94534513233"></a>-</p>
</td>
</tr>
<tr id="row859053413417"><td class="cellrowborder" valign="top" width="15.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p844941513297"><a name="p844941513297"></a><a name="p844941513297"></a>restartType</p>
</td>
<td class="cellrowborder" valign="top" width="14.69%" headers="mcps1.2.6.1.2 "><p id="p220916992912"><a name="p220916992912"></a><a name="p220916992912"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24.23%" headers="mcps1.2.6.1.3 "><p id="p1820909182911"><a name="p1820909182911"></a><a name="p1820909182911"></a>Type of reset.json update</p>
</td>
<td class="cellrowborder" valign="top" width="28.82%" headers="mcps1.2.6.1.4 "><p id="p15209596295"><a name="p15209596295"></a><a name="p15209596295"></a>podReschedule or hotReset</p>
</td>
<td class="cellrowborder" valign="top" width="16.31%" headers="mcps1.2.6.1.5 "><p id="p95471047133013"><a name="p95471047133013"></a><a name="p95471047133013"></a>Value is podReschedule for single Pod rescheduling, and hotReset for hot recovery scenarios</p>
</td>
</tr>
<tr id="row165081157153910"><td class="cellrowborder" valign="top" width="15.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p750805713392"><a name="p750805713392"></a><a name="p750805713392"></a>checkCode</p>
</td>
<td class="cellrowborder" valign="top" width="14.69%" headers="mcps1.2.6.1.2 "><p id="p0508145711393"><a name="p0508145711393"></a><a name="p0508145711393"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24.23%" headers="mcps1.2.6.1.3 "><p id="p250845783917"><a name="p250845783917"></a><a name="p250845783917"></a>Check code</p>
</td>
<td class="cellrowborder" valign="top" width="28.82%" headers="mcps1.2.6.1.4 "><p id="p750835713919"><a name="p750835713919"></a><a name="p750835713919"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="16.31%" headers="mcps1.2.6.1.5 "><p id="p175081157113917"><a name="p175081157113917"></a><a name="p175081157113917"></a>-</p>
</td>
</tr>
</tbody>
</table>

**data-trace-job_name<a name="section19954856135618"></a>**

Stores the `on`/`off` status of various tracing types for the current job. It is mounted to the compute node by Ascend Device Plugin for storage. After the training container mounts this file, TaskD reads it to trace various data.

**Table 3**  data-trace-job_name ConfigMap

<a name="table97521457610"></a>

|Field|Meaning|Value|Type|
|--|--|--|--|
|Communication|Identifies the communication operator.|on/off|string|
|Step|Identifies the Step latency.|on/off|string|
|SaveCheckpoint|Identifies the SaveCheckpoint duration.|on/off|string|
|FP|Identifies the forward propagation data.|on/off|string|
|DataLoader|Identifies the DataLoader duration.|on/off|string|

>[!NOTE]
>
>- This ConfigMap must be in the same namespace as the training job (`data-trace-<job_name>`), and include the label `reset=true`.
>- This ConfigMap is mounted by the Ascend Device Plugin to the `/user/cluster-info/datatrace-config/namespace.data-trace-job_name/*` folder on the training node, with the file name `profilingSwitch`.
>- If the user does not create this ConfigMap, ClusterD will attempt to automatically create it upon the first call to the gRPC interface `ModifyTrainingDataTraceSwitch`.
>- To use this feature, the user should mount the `profilingSwitch` file on the node into the `/user/cluster-info/datatrace-config/` directory within the container using the `hostPath` method.
>- Currently, `Step`, `SaveCheckpoint`, `FP`, and `DataLoader` are enabled by default, and these four can only be enabled or disabled synchronously. When all five fields are `off`, all instrumentation is disabled; otherwise, the above four are enabled by default, and the communication operator is enabled or disabled based on its switch status.

**steptime-dtpgroup<a name="section1146122513469"></a>**

Stores the save path and start/stop switch for the job's iteration latency and grouping information. When starting a job, users can configure ConfigMap parameters through the CCAE management platform to determine whether the job has degraded.

**Table 4**  steptime-dtpgroup ConfigMap

<a name="table3610611144615"></a>

|First-level Parameter|Second-level Parameter|Meaning|Value|Remarks|
|--|--|--|--|--|
|data|PerfDumpPath|Path for saving iteration latency and grouping information.|String|-|
|-|PerfDumpConfig|Start/stop switch for iteration latency and grouping information.|String|-|

## Custom Chip Faults<a name="ZH-CN_TOPIC_0000002511346805"></a>

**Fault Levels in faultCode.json<a name="section579455712489"></a>**

The hierarchical handling based on different chip fault levels is introduced in resumable training. If you need to modify the fault level of a fault code, see [(Optional) Configuring Chip Fault Levels](../usage/resumable_training/03_configuring_fault_detection_levels.md).

After Ascend Device Plugin obtains chip fault codes from the driver, it classifies the faults into the following levels based on their impact on devices and workloads. For a detailed description, see [Table 1](#table7618951152212).

**Table 1**  Fault level and handling polices

<a name="table7618951152212"></a>
<table><thead align="left"><tr id="row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="p12618851162220"><a name="p12618851162220"></a><a name="p12618851162220"></a>Fault Handling Policy</p>
</th>
<th class="cellrowborder" valign="top" width="35.78%" id="mcps1.2.5.1.2"><p id="p16618125162219"><a name="p16618125162219"></a><a name="p16618125162219"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="20.349999999999998%" id="mcps1.2.5.1.3"><p id="p1163819316544"><a name="p1163819316544"></a><a name="p1163819316544"></a>Rescheduling Policy</p>
</th>
<th class="cellrowborder" valign="top" width="24.81%" id="mcps1.2.5.1.4"><p id="p171971327125410"><a name="p171971327125410"></a><a name="p171971327125410"></a>Graceful Fault Tolerance</p>
</th>
</tr>
</thead>
<tbody><tr id="row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p7618125114229"><a name="p7618125114229"></a><a name="p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="p1261835110227"><a name="p1261835110227"></a><a name="p1261835110227"></a>A fault that does not affect services and does not need to be handled.</p>
</td>
<td class="cellrowborder" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="p10638123115414"><a name="p10638123115414"></a><a name="p10638123115414"></a>Not handled for now</p>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="p719714273546"><a name="p719714273546"></a><a name="p719714273546"></a>Not handled for now</p>
</td>
</tr>
<tr id="row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5618751102216"><a name="p5618751102216"></a><a name="p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="p05771854113911"><a name="p05771854113911"></a><a name="p05771854113911"></a>Affects service execution and requires re-execution of the service request.</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="p13855131912555"><a name="p13855131912555"></a><a name="p13855131912555"></a>Isolate the chip and reschedule the job</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><a name="note11901123612819"></a><span class="notetitle">Note:</span><div class="notebody"><p id="zh-cn_topic_0000002479386448_p1069261722310"><a name="zh-cn_topic_0000002479386448_p1069261722310"></a><a name="zh-cn_topic_0000002479386448_p1069261722310"></a>If an inference job subscribes<span id="zh-cn_topic_0000002479386448_ph4356222144812"><a name="zh-cn_topic_0000002479386448_ph4356222144812"></a><a name="zh-cn_topic_0000002479386448_ph4356222144812"></a> to</span> fault information, and a RestartRequest fault occurs on the inference card used by the job and the fault duration does not exceed 60 seconds, job rescheduling will not be performed. If the fault duration exceeds 60 seconds and the fault is not recovered, the chip will be isolated and job rescheduling will be performed.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="p9145165785517"><a name="p9145165785517"></a><a name="p9145165785517"></a>Re-execute the inference request in inference scenarios, and re-execute the training workload in training scenarios</p>
</td>
</tr>
<tr id="row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15618851132212"><a name="p15618851132212"></a><a name="p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p3618851182216"><a name="p3618851182216"></a><a name="p3618851182216"></a>Affects service execution and requires re-execution of the service.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1419712272549"><a name="p1419712272549"></a><a name="p1419712272549"></a>Re-execute the service</p>
</td>
</tr>
<tr id="row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p66188511222"><a name="p66188511222"></a><a name="p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p661865162211"><a name="p661865162211"></a><a name="p661865162211"></a>Affects service execution. The chip needs to be reset when it is idle.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p178789204535"><a name="p178789204535"></a><a name="p178789204535"></a>Wait for the chip to become idle and then reset it</p>
</td>
</tr>
<tr id="row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17618155116227"><a name="p17618155116227"></a><a name="p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p108302057102114"><a name="p108302057102114"></a><a name="p108302057102114"></a>Affects service execution. The chip needs to be reset immediately.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p969972925312"><a name="p969972925312"></a><a name="p969972925312"></a>Immediately stop the training workload, reset the chip, and then re-execute the service</p>
</td>
</tr>
<tr id="row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p961885142215"><a name="p961885142215"></a><a name="p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p18618151202216"><a name="p18618151202216"></a><a name="p18618151202216"></a>Unrecoverable. The chip needs to be isolated.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p019742745411"><a name="p019742745411"></a><a name="p019742745411"></a>Isolate the chip and reschedule the job</p>
</td>
</tr>
<tr id="row1930365771212"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002171521445_p16227299522"></a><a name="zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002171521445_p546081915499"></a><a name="zh-cn_topic_0000002171521445_p546081915499"></a>Does not affect services for now, but jobs will no longer be scheduled to this chip subsequently.</p>
</td>
<td class="cellrowborder" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002171521445_p222102912521"></a><a name="zh-cn_topic_0000002171521445_p222102912521"></a>Pre-isolate chip</p>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002171521445_p12221329155217"></a><a name="zh-cn_topic_0000002171521445_p12221329155217"></a>Pre-isolate chip</p>
</td>
</tr>
<tr id="row89346317136"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002171521445_p835213245522"></a><a name="zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002171521445_p1354813311915"></a>Handled based on the value of the subHealthyStrategy parameter configured in the job YAML. For details, see <a href="../api/ascend_operator.md">Table 1 YAML parameter description</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002171521445_p3352524125220"></a>When a sub-health fault occurs on a chip, it needs to be handled according to the policy in <a href="../usage/resumable_training/06_configuring_the_job_yaml_file.md">Configuring YAML</a>.</p>
<div class="note" id="zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002171521445_note7936204710536"></a><a name="zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">[!NOTE] Note:</span><div class="notebody"><p id="zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002171521445_p15222114115810"></a>If a fault of another level occurs on the chip subsequently, the SubHealthFault</p>
<p id="zh-cn_topic_0000002171521445_p109369476532"><a name="zh-cn_topic_0000002171521445_p109369476532"></a><a name="zh-cn_topic_0000002171521445_p109369476532"></a>handling policy does not affect the handling of faults at other levels.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002171521445_p8352172425218"></a><a name="zh-cn_topic_0000002171521445_p8352172425218"></a>Handle according to the policy</p>
</td>
</tr>
</tbody>
</table>

**faultCustomization.json<a name="section33036167576"></a>**

If the user does not manually modify the `faultCustomization.json` file, Ascend Device Plugin performs fault handling according to the default configuration (default values) of `faultCustomization.json`.

**Table 2**  faultCustomization.json

<a name="table1519814413572"></a>

| First-level Parameter | Second-level Parameter | Description |
|--|--|--|
| GraceTolerance | - | Graceful tolerance related configuration.<p>If GraceTolerance and its sub-parameters do not exist or exceed the value range, the default values are used.</p> |
| - | WaitProcessReadCMTime | When using graceful tolerance mode, the time to wait for the management process to read the ConfigMap file, in seconds. Value range: 5 to 90. Default value is 30. |
| - | WaitDeviceResetTime | When using graceful tolerance mode, the maximum time to wait for the chip to restart, in seconds. Value range: 60 to 180. Default value is 150. |
| - | WaitFaultSelfHealingTime | When using graceful tolerance mode, the time to wait for RestartBusiness level fault recovery, in seconds. Value range: 1 to 30. Default value is 15. |
| FaultFrequency | - | Custom fault frequency. When the number of occurrences of a fault within the time window reaches the upper limit, the fault is handled according to the configured fault handling policy.<ul><li>If the value range of FaultFrequency or its sub-parameters is incorrect, this configuration is ignored.</li><li>If the data format of FaultFrequency or its sub-parameters is incorrect, the default configuration is used.</li></ul> |
| - | EventId | Fault code ID.<p>Only one FaultFrequency parameter is allowed for each fault code (EventId). If multiple are configured, only the first valid one takes effect.</p> |
| - | TimeWindow | Time window, which counts the number of fault occurrences from the current time minus TimeWindow to the current time, in seconds. Value range: 60 to 864,000. |
| - | Times | The upper limit of occurrences for the same fault. Value range: 1 to 100. If the number of occurrences of this fault within the time window is greater than or equal to this value, it is handled and reported according to the policy defined in FaultHandling. |
| - | FaultHandling | <p>The handling policy for the fault after the number of occurrences reaches the upper limit. Supports configuring fault handling policies at different levels. If ReleaseTimeWindow is configured, the policy can be automatically released when conditions are met. To support manual fault removal, configure the handling policy as ManuallySeparateNPU.</p><ul><li>PreSeparateNPU: This fault handling mode pre-isolates the chip and determines whether to reschedule based on the actual running status of the training task.</li><li>ManuallySeparateNPU:<ul><li>When this policy is triggered, it directly reports to K8s that the chip is unhealthy and writes the chip name to device-info-cm.</li><li>As long as the chip name is saved in this field, the chip remains isolated even if the fault recovers, until the maintenance personnel manually delete the chip name from this field, or the recovery duration exceeds ReleaseTimeWindow.</li><li>This field only allows Ascend Device Plugin to add or modify it. Maintenance personnel can only delete chip names from this field.</li><li>This policy is currently not supported in faultCode.json.</li></ul></li></ul> |
| - | ReleaseTimeWindow | If the fault has recovered and no recurrence of the fault occurs for a duration exceeding ReleaseTimeWindow, the escalated fault handling policy is released. The value range of this parameter is 60 to the maximum value of uint32, in seconds. If this parameter is not configured, it means the policy will not be downgraded after escalation. Since ManuallySeparateNPU supports manual release, this parameter must be configured for all other handling policies except ManuallySeparateNPU to prevent the device from being permanently unusable after fault recovery. |
| FaultDuration | - | Custom fault timeout policy. When the duration of a fault reaches the configured upper limit, the fault is handled according to the specified fault handling policy.<ul><li>If the value range of FaultDuration or its sub-parameters is incorrect, this configuration is ignored.</li><li>If the data format of FaultDuration or its sub-parameters is incorrect, the default configuration is used.</li></ul> |
| - | EventId | Fault ID.<p>Only one FaultDuration parameter is allowed for each fault code (EventId). If multiple are configured, only the first valid one takes effect.</p> |
| - | FaultTimeout | If the fault duration exceeds this value, the fault is handled according to the fault handling policy defined in FaultHandling, in seconds. Value range: 0 to 600. The default values are described as follows.<ul><li>For the parameter plane network fault with fault ID 81078603, the default value is 20.</li><li>For the on-chip memory multi-bit fault with fault ID 80E01801, the default value is 30.</li><li>For other faults, the default value is 0.</li></ul> |
| - | RecoverTimeout | If the fault recovery time exceeds this value, fault recovery is reported, in seconds. Value range: 0 to 86,400. The default values are described as follows.<ul><li>For the parameter plane network fault with fault ID 81078603, the default value is 60. Setting it to 0 is not recommended; it is recommended to be greater than the listWatchPeriod health status check cycle. For details about listWatchPeriod, see the "Ascend Device Plugin Startup Parameters" table in [Ascend Device Plugin](../developer_guide/installation_deployment/manual_installation/04_ascend_device_plugin.md).</li><li>For other faults, the default value is 0.</li></ul> |
| - | FaultHandling | <p>The fault handling policy after the fault duration is exceeded. Supports configuring fault handling policies at different levels, and also supports configuring the PreSeparateNPU fault handling policy.</p><p>The fault handling policy after the fault duration is exceeded is recommended to be higher than the fault handling policy of the fault itself; otherwise, the configuration does not take effect.</p><p>Configuring the ManuallySeparateNPU policy is not supported, and the configuration does not take effect.</p> |

>[!NOTE]
>
>- If a fault code is configured with both a fault frequency (FaultFrequency) and a fault timeout policy (FaultDuration), and the number of timeouts for this fault code within the TimeWindow reaches the maximum number supported by the job, the most severe level among the following three will be used for handling. The three are: the fault handling policy of the fault itself, and the fault handling policies configured in FaultFrequency and FaultDuration.
>- If a fault code is configured with both a fault frequency and a fault timeout policy, the fault is considered to have occurred only after the fault timeout, and the frequency count increases by one. The fault is considered recovered only after the recovery exceeds RecoverTimeout. After recovery, the next count can only be accumulated upon another fault timeout.
>- The network fault with fault ID 81078603 only supports configuration of three fault handling policies: NotHandleFault, PreSeparateNPU, or SeparateNPU. If configured with other policies, the default configuration NotHandleFault will be used.
>- When Ascend Device Plugin is upgraded from a version earlier than 26.0.0 to version 26.0.0 or later, if the ConfigMap of Ascend Device Plugin already contains the ManuallySeparateNPU key-value, its degradation time window is the maximum ReleaseTimeWindow value in faultCustomization.json. If no fault code is configured with ReleaseTimeWindow, the existing ManuallySeparateNPU in the ConfigMap will not be degraded.
>- After modifying/deleting policies in mindx-dl-fault-config, the upgraded fault handling policies will also be updated with the configuration changes. However, after the ManuallySeparateNPU policy is upgraded, deleting the fault handling policy configuration item will not remove it; the removal method for ManuallySeparateNPU must be followed. Additionally, when other fault level configurations are updated to ManuallySeparateNPU, the configuration will not take effect or will not meet expectations.

## Custom UnifiedBus Device Faults<a name="ZH-CN_TOPIC_0000002511426735"></a>

Resumable training performs hierarchical handling based on different levels of UnifiedBus bus device faults. If users need to modify the fault level of a fault code, see [(Optional) Configuring UnifiedBus Device Fault Levels](../usage/resumable_training/03_configuring_fault_detection_levels.md).

After Ascend Device Plugin obtains fault codes from the driver, it classifies faults into the following five levels based on their impact on devices and workloads, and performs corresponding rescheduling handling. For detailed description, see [Table 1](#table212253274720).

**Table 1**  Fault levels and handling policies

<a name="table212253274720"></a>

|Fault Type|Description|Rescheduling Policy|
|--|--|--|
|NotHandleFault|Does not affect workloads temporarily and can recover automatically. No handling required.|Not handled for now.|
|SubHealthFault|Affects workload running performance. The cause of sub-health fault needs to be investigated.|When a sub-health fault occurs, it needs to be handled according to the sub-health policy specified by the subHealthyStrategy parameter in [Table 1 YAML parameter description](../api/ascend_operator.md).|
|RestartRequestFault|Workload running fails. The workload request needs to be re-executed.|Stop the current training job, isolate the node, and reschedule the job.|
|ResetFault|Workload running fails.|Stop the current training job, isolate the node, and reschedule the job.|
|SeparateFault|Workload running fails. The device or board needs to be replaced.|Stop the current training job, isolate the node, and reschedule the job.|
