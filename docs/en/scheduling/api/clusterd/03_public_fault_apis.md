# Public Fault APIs<a name="ZH-CN_TOPIC_0000002479226838"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:46:30.743Z pushedAt=2026-06-09T02:05:50.690Z -->

## ConfigMap<a name="ZH-CN_TOPIC_0000002479386788"></a>

**Description<a name="section359310211618"></a>**

Receives ConfigMap information for public faults and accesses the resumable training process.

>[!NOTE]
>
>- If the parameters in the actual ConfigMap do not match the defined value ranges, ClusterD will discard the fault information and ignore it.
>- For public faults injected through the ConfigMap or gRPC interface, the total number of faults across all nodes is capped at 50,000. When the number of faults exceeds 50,000, any further fault injection will cause ClusterD to discard the fault information and ignore it.
>- The label of the ConfigMap must be `mc-consumer-publicfault=true`, and the key of `Data` must be `PublicFault`.
>- When sending public faults through the ConfigMap method, ensure that the data volume of a single request cannot exceed 1 MB; otherwise, the ConfigMap update will fail.

**Parameters<a name="section4809204015614"></a>**

For details about the parameters, see the following table.

**Table 1** Fault information description

|Parameter Name|Description|Value|Type|Mandatory|
|--|--|--|--|--|
|id|Unique message identifier|A string of 8 to 128 characters, supporting uppercase and lowercase letters, digits, hyphens (-), underscores (_), and dots (.), and must be unique.|string|Yes|
|timestamp|Timestamp when the message is sent|Timestamp (unit: ms), 13 digits, must be after 2025-01-01T00:00:00Z.|int64|Yes|
|version|Message version number|Value: 1.0.|string|Yes|
|resource|Fault sender|Default configuration: CCAE, fd-online, pingmesh, Netmind, dpcStorage.<ul><li>The fault sender for public faults must exist in the publicFaultResource of the fault configuration file.</li><li>For a new fault sender, you need to manually configure it in the fault configuration file. For details, see [(Optional) Configuring Public Fault Levels and Senders](../../usage/resumable_training/03_configuring_fault_detection_levels.md).</li></ul>|string|Yes|
|faults|Fault content|Slice, length >0 and ≤100.|[]object, [fault](#fault0023698)|Yes|

**Table 2** fault field description

<a name="fault0023698"></a>

|Parameter Name|Description|Value|Type|Mandatory|
|--|--|--|--|--|
|faultId|Fault instance ID|A string of 8 to 128 characters, supporting uppercase and lowercase letters, digits, hyphens (-), underscores (_), and dots (.). Must be unique.<p>For the same fault instance, the faultId must be unique.</p>|string|Yes|
|faultType|Fault type|Value:<ul><li>`NPU`: NPU fault.</li><li>`Node`: Node fault.</li><li>`Network`: Network fault.</li><li>`Storage`: Storage fault.</li></ul>This field is displayed as `"PublicFault"` in `cluster-info-cm`.|string|Yes|
|faultCode|Fault code|User-definable, must be a unique 9-digit code.<ul><li>Fault codes for accessing resumable training must exist in the `publicFaultCode` of the fault configuration file.</li><li>For new fault codes, configure their fault levels in the fault configuration file. For details, see [(Optional) Configuring Public Fault Levels and Senders](../../usage/resumable_training/03_configuring_fault_detection_levels.md).</li><li>It is recommended that fault codes follow the rules defined in the fault code description table for easier maintenance.</li><li>If the same fault code occurs twice on an NPU, the `fault_code` field in `cluster-info-cm` will record two identical fault codes.</li></ul>|string|Yes|
|faultTime|Fault generation time|Timestamp (unit: ms), 13 digits, must be after 2025-01-01T00:00:00Z.<ul><li>Whether it is fault generation or fault recovery, this field is always the fault generation time.</li><li>This field is displayed in seconds in `cluster-info-cm`.</li></ul>|int64|Yes|
|assertion|Fault status|Value:<ul><li>`occur`: Fault generation.</li><li>`recover`: Fault recovery.</li><li>`once`: One-time event.</li></ul><div class="note"><span class="notetitle">NOTE</span><div class="notebody"><ul><li>Public fault recovery requires writing the recover event of the corresponding fault into the ConfigMap, and cannot be achieved by deleting the ConfigMap.</li><li>For one-time events, the fault is automatically cleared after a few seconds.</li></ul></div></div>|string|Yes|
|faultLocation|Fault location information|Fault source information, length ≤10, map key length ≤16, value length ≤128. e.g. `key: npuIp, value: ip`|map[string]string|No|
|influence|Scope of fault impact|Slice, length >0 and ≤1000.|[]object, [faultInfo](#faultinfo0023698)|Yes|
|description|Fault description|0 to 512 characters. Contains non-whitespace characters and spaces.|string|No|

**Table 3**  faultInfo description

<a name="faultinfo0023698"></a>

|Parameter Name|Description|Value|Type|Mandatory|
|--|--|--|--|--|
|nodeName|Node Name. Can be queried using the **kubectl get nodes -owide** command.|A string of 1 to 253 characters, supporting lowercase letters, digits, hyphens (-), and dots (.), and must start and end with an alphanumeric character. When this field is present, `nodeSN` is not used.<p>If the node name does not exist in the K8s cluster, ClusterD will not report a node name error, nor write this fault information to `cluster-info-device-cm`.</p>|string|Either `nodeName` or `nodeSN` is required|
|nodeSN|Node SN|The SN of the node. The value is the node annotation written by NodeD, with the key `product-serial-number`.<p>If this field is used instead of nodeName, NodeD must be installed in advance.</p>|string|Either `nodeName` or `nodeSN` is required|
|deviceIds|NPU Physical ID|Length (0, 32], each element value [0, 32), and duplicates are not allowed.<ul><li>If the faulty NPU cannot be accurately identified, all NPU physical IDs on the node must be filled in.</li><li>If a non-existent NPU physical ID on the node is passed in, ClusterD will still display it in `cluster-info-device-cm`.</li></ul>|[]int32|Yes|

## gRPC Interface<a name="ZH-CN_TOPIC_0000002479226854"></a>

**Function Description<a name="section125411749115817"></a>**

Receives and processes public fault sending requests from gRPC clients, and accesses the resumable training process.

>[!NOTE]
>
>- If the actual gRPC request parameters do not conform to the defined value ranges, ClusterD will discard the fault information and ignore it.
>- For public faults injected via ConfigMap or gRPC interface, the total number of faults across all nodes is capped at 50,000. When the number of faults exceeds 50,000, any further fault injection will cause ClusterD to discard the fault information and ignore it.
>- To clear a public fault, the recover event of the corresponding fault must be sent to ClusterD via the gRPC interface.

**Prototype<a name="section1698941035919"></a>**

```proto
rpc SendPublicFault(PublicFaultRequest) returns (RespStatus){}
```

**Input Parameters<a name="section52771657118"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|PublicFaultRequest|<p>message PublicFaultRequest{<p>string id = 1;</p><p>int64 timestamp = 2;</p><p>string version = 3;</p><p>string resource = 4;</p><p>repeated Fault faults = 5;</p>}</p><p>message Fault{<p>string faultId = 1;</p><p>string faultType = 2;</p><p>string faultCode = 3;</p><p>int64 faultTime = 4;</p><p>string assertion = 5;</p><p>map<string, string> faultLocation = 6;</p><p>repeated PubFaultInfo influence = 7;</p><p>string description = 8;</p>}</p><p>message PubFaultInfo{<p>string nodeName = 1;</p><p>string nodeSN = 2;</p><p>repeated int32 deviceIds = 3;</p>}</p>|<p>**PublicFaultRequest.id**: Unique message identifier</p><p>**PublicFaultRequest.timestamp**: Timestamp when the message is sent</p><p>**PublicFaultRequest.version**: Message version number</p><p>**PublicFaultRequest.resource**: Fault sender</p><p>**PublicFaultRequest.faults**: Fault content</p><p>**Fault.faultId**: Fault instance ID</p><p>**Fault.faultType**: Fault type</p><p>**Fault.faultCode**: Fault code</p><p>**Fault.faultTime**: Fault generation time</p><p>**Fault.assertion**: Fault status</p><p>**Fault.faultLocation**: Fault location information</p><p>**Fault.influence**: Scope of fault impact</p><p>**Fault.description**: Fault description</p><p>**PubFaultInfo.nodeName**: Node name</p><p>**PubFaultInfo.nodeSN**: Node SN</p><p>**PubFaultInfo.deviceIds**: NPU physical ID</p><p>For details about the above parameters and their values, see [ConfigMap](#configmap).</p>|

**Return Value<a name="section521319321415"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|RespStatus|message RespStatus{<p>int32 code = 1;</p><p>string info = 2;</p>}|**RespStatus.code**: Return code.<ul><li>`0`: Fault sent successfully.</li><li>Other values: Fault sending failed. `409` indicates invalid request parameters, and `410` indicates the message sending frequency exceeds the limit.</li></ul>**RespStatus.info**: Return information.|
