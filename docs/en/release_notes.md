# Release Notes

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-04T12:22:44.436Z pushedAt=2026-06-04T12:26:58.409Z -->

## Version Mapping Description<a name="ZH-CN_TOPIC_0000002492283212"></a>

### Product Version Information<a name="ZH-CN_TOPIC_0000002524562895"></a>

<a name="zh-cn_topic_0000001935094108__Ref249955742"></a>
<table><tbody><tr id="zh-cn_topic_0000001935094108_row244mcpsimp"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000001935094108_p246mcpsimp"><a name="zh-cn_topic_0000001935094108_p246mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p246mcpsimp"></a>Name</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.1.1 "><p id="p92555221126"><a name="p92555221126"></a><a name="p92555221126"></a><span id="ph19255162231216"><a name="ph19255162231216"></a><a name="ph19255162231216"></a>MindCluster</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row255mcpsimp"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.2.1"><p id="zh-cn_topic_0000001935094108_p257mcpsimp"><a name="zh-cn_topic_0000001935094108_p257mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p257mcpsimp"></a>Version Number</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.2.1 "><p id="zh-cn_topic_0000001935094108_p233mcpsimp"><a name="zh-cn_topic_0000001935094108_p233mcpsimp"></a><a name="zh-cn_topic_0000001935094108_p233mcpsimp"></a>26.0.0</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001935094108_row7259721105019"><th class="firstcol" valign="top" width="25%" id="mcps1.1.3.3.1"><p id="zh-cn_topic_0000001935094108_p7260182135013"><a name="zh-cn_topic_0000001935094108_p7260182135013"></a><a name="zh-cn_topic_0000001935094108_p7260182135013"></a>Version Type</p>
</th>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.1.3.3.1 "><p id="zh-cn_topic_0000001935094108_p72606219501"><a name="zh-cn_topic_0000001935094108_p72606219501"></a><a name="zh-cn_topic_0000001935094108_p72606219501"></a>Release Version</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>
> MindCluster 26.0 version planning: MindCluster 26.0.0, MindCluster 26.1.0, MindCluster 26.2.0, and MindCluster 26.3.0.

### Product Version Mapping<a name="ZH-CN_TOPIC_0000002524562893"></a>

|Product Name|Version|
|--|--|
|Ascend HDK| <ul><li>Atlas 350 PCIe card: 25.7.RC1</li><li>Other products: 26.0.RC1</li></ul> |
|CANN|9.0.0|

### Virus Scan Results<a name="ZH-CN_TOPIC_0000002492443186"></a>

Virus scan passed.

## Version Compatibility Description<a name="ZH-CN_TOPIC_0000002524442915"></a>

MindCluster components must be used together. Do not mix components from different versions.

**Table 1**  Software version compatibility description

|MindCluster Software Version|MindCluster Version to Upgrade|CANN Version|Ascend HDK Version|FrameworkPTAdapter Version|MindSpore Version|
|--|--|--|--|--|--|
|MindCluster 26.0.0|<ul><li>MindCluster 7.0.RC1 and patch version</li><li>MindCluster 7.1.RC1 and patch version</li><li>MindCluster 7.2.RC1 and patch version</li><li>MindCluster 7.3.0 and patch version</li></ul>|<ul><li>CANN 8.5.0 and patch version</li><li>CANN 9.0.0 and patch version</li></ul>|<ul><li>Ascend HDK 25.5.0 and patch version</li><li>Ascend HDK 26.0.RC1 and patch version</li><li>Ascend HDK 25.7.RC1 and patch version</li></ul>|<ul><li>FrameworkPTAdapter 7.3.0 and patch version</li><li>FrameworkPTAdapter 26.0.0 and patch version</li></ul>|<ul><li>MindSpore 2.7.2 and patch version</li><li>MindSpore 2.9.0 and patch version</li></ul>|

## Version Usage Notes<a name="ZH-CN_TOPIC_0000002492283210"></a>

None

## 26.0.0 Release Notes<a name="ZH-CN_TOPIC_0000002492443184"></a>

### New Features<a name="ZH-CN_TOPIC_0000002524442919"></a>

|Feature|Description|
|--|--|
|MindIO ACP|Supports Async CheckPoint Persistence (ACP) and Training Fault Tolerance (TFT).|
|MindIO TFT|<ul><li>Supports ACP and TFT.</li><li>Supports online recovery at specified checkpoint steps after precision anomalies.</li><li>Supports optimizer-differentiated replica scenarios for the iFLYTEK HULK framework.</li></ul>|
|MindCluster Ascend FaultDiag|<ul><li>Supports troubleshooting for the Atlas 350 PCIe card.</li><li>Adds Ascend-faultdiag-toolkit tool to support NPU disconnection and infrastructure link diagnosis.</li><li>No longer supports building the fault mode library into a binary; the fault mode library is now directly open-sourced.</li></ul>|
|MindCluster basic components|<ul><li>The service plane network newly supports IPv6.</li><li>Supports configuring self-healing fault levels or specific fault codes based on task dimensions.</li><li>Support soft partitioning-based scheduling for Atlas A2/A3 series.</li><li>Support hard partitioning-based scheduling for Atlas A2/A3 series.</li><li>MoE EP tasks based on MindIE support switch affinity scheduling.</li><li>The gRPC heartbeat detection interval of ClusterD is adjusted from the default 5 minutes to 5 seconds.</li><li>Support post-processing of self-healing faults and rescheduling of non-self-healing faults.</li><li>Identify whether a fault is a hardware fault based on the cluster dimension; recurring hardware faults are automatically and forcibly isolated to prevent repeated task interruptions.</li><li>Both automatic forced isolation at the cluster dimension and at the node dimension support configuring an automatic release time.</li><li>Add network affinity scheduling algorithm of any level, adapted to the Atlas 9000 A3 SuperPoD cluster computing system.</li><li>New task information subscription interface `SubscribeJobSummarySignalList` is added, supporting returning historical task information on the first subscription.</li><li>A `ConfigMap` is added to display the reason for task scheduling failures, facilitating fault location.</li><li>NPU Exporter supports listening and reporting custom metrics through configuration files.</li><li>NPU Exporter has been refactored to obtain all NPU utilization metrics through a single interface instead of multiple interfaces, thereby preventing data inconsistency.</li><li>The ClusterD fault notification service supports registration via domain name.</li><li>The ClusterD job information subscription interface adds a unique job identifier field.</li><li>NPU Exporter supports metric reporting for the Atlas 350 PCIe card.</li><li>Ascend Docker Runtime supports the Atlas 350 PCIe card.</li><li>The Atlas 350 PCIe card supports affinity scheduling, device discovery, RankTable generation, and fault rescheduling.</li><li>Infer Operator can manage inference tasks through custom CRD.</li></ul>|

### Key Feature Changes<a name="ZH-CN_TOPIC_0000002524562891"></a>

MindCluster basic components:

- The gRPC heartbeat detection interval of ClusterD is adjusted from the default 5 minutes to 5 seconds.
- The automatic forced isolation and release of repeatedly faulty chips at the cluster dimension is supported.
- For the Atlas 350 PCIe card:
  - The task resource request `huawei.com/Ascend910` is changed to `huawei.com/npu`.
  - The underlying DCMI is changed to the DCMI V2.

### Service Interface Changes<a name="ZH-CN_TOPIC_0000002492443182"></a>

|Feature|Interface Change|
|--|--|
|MindIO ACP|None|
|MindIO TFT|New `tft_register_exception_handler` interface that registers an exception handler.|
|MindCluster Ascend FaultDiag|New interfaces related to Ascend-faultdiag-toolkit. For details, see [API Description](./faultdiag/ascend-faultdiag-toolkit/01_api_description.md).|
|MindCluster basic components|<ul><li>Configuration fields for self-healing fault level, fault code, and self-healing duration are added to the task creation interface.</li><li>Configuration fields for soft partitioning mode, AICore percentage, and high-bandwidth memory size are added to the task creation interface.</li><li>ClusterD supports configuring the startup switch, triggering frequency, and isolation duration for automatic forced fault isolation.</li><li>Ascend Device Plugin adds an isolation duration configuration field for automatic forced isolation.</li><li>Support multi-level network topology configuration and multi-level network affinity configuration for tasks.</li><li>New task information subscription interface `SubscribeJobSummarySignalList`.</li><li>New interface for querying the cause of task scheduling exceptions.</li><li>New file-based custom metric interface.</li><li>The calculation method of the NPU utilization interface is optimized for NPU Exporter.</li><li>Basic device information, fault code, and chip name for the Atlas 350 PCIe card are added.</li></ul>|

### Resolved Issues<a name="ZH-CN_TOPIC_0000002492283206"></a>

- When resources such as MindIO processor are released and the program crashes, TaskD Agent cannot exit. A fallback exit mechanism is added.
- After training ends, when TaskD Worker calls the `mspti_activity_flush_all` method, a `double free` error is reported.
- Concurrent map read/write by TaskD Manager causes process crash.
- Delayed update of pg cache in ClusterD.
- Forced requirement for healthy RoCE network between MindIE instances causes MindIE task scheduling failure.
- Pod does not exit after training completion in the user-defined torch log file scenario.
- When NodeD is installed and the cluster has more than 1024 cards, the gRPC connection limit of ClusterD is exceeded, preventing other connections from being established

### Known Issues<a name="ZH-CN_TOPIC_0000002492443180"></a>

None

## Upgrade Impact<a name="ZH-CN_TOPIC_0000002492283208"></a>

### Impact of the Upgrade Process on the Current System<a name="ZH-CN_TOPIC_0000002524442911"></a>

None

### Impact on the Current System After Upgrade<a name="ZH-CN_TOPIC_0000002492443178"></a>

None

## 26.0.0 Documents<a name="ZH-CN_TOPIC_0000002524562889"></a>

| Document | Content Description | Release Notes |
|--|--|--|
| [MindCluster Cluster Scheduling User Guide](./scheduling/introduction/00_overview.md) | Introduces cluster scheduling components, feature principles, and usage references, including installation and deployment, integration and adaptation examples, and API references for each component, as well as working principles for some scheduling solutions. | Adds soft partitioning-based scheduling and multi-level scheduling. |
| [MindCluster Fault Diagnosis User Guide](./faultdiag/introduction.md) | Provides usage guidance for features such as log collection, cleaning and dumping, and troubleshooting. | Adds the Atlas 350 PCIe card fault mode and ascend-faultdiag-toolkit. |

## List of Fixed Vulnerabilities<a name="ZH-CN_TOPIC_0000002524442913"></a>

None
