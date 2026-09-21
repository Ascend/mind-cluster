# Public Fault Reporting<a name="ZH-CN_TOPIC_0000002512192054"></a>

<!-- md-trans-meta sourceCommit=0f2bb53c19b633e38025bcfa28764005938006f7 translatedAt=2026-09-01T09:28:41.471Z pushedAt=2026-09-01T10:13:08.267Z -->

ClusterD provides the public fault reporting capability, supporting access to public faults through both ConfigMap and gRPC interfaces, and linking with the resumable training process.

>[!NOTE]
>
>- If the actual ConfigMap or gRPC request parameters do not conform to the defined value ranges, ClusterD discards the fault information without processing it.
>- For public faults injected through the ConfigMap or gRPC interface, the total number of faults across all nodes is capped at 50,000. When the number of faults exceeds 50,000, ClusterD discards the fault information without processing it upon further fault injection.

## Reporting Faults Through ConfigMap<a name="ZH-CN_TOPIC_0000002479386788"></a>

**Function Description**

Receives the ConfigMap information of public faults and accesses the resumable training process.

>[!NOTE]
>
>- The Label of the ConfigMap must be `mc-consumer-publicfault=true`, and the key of Data must be `PublicFault`.
>- When sending public faults through ConfigMap, the data volume of a single request must not exceed 1 MB; otherwise, the ConfigMap update will fail.

**ConfigMap Example**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: public-fault-example
  namespace: mindx-dl
  labels:
    mc-consumer-publicfault: "true"
data:
  PublicFault: |
    {
      "id": "public-fault-msg-001",
      "timestamp": 1735689601000,
      "version": "1.0",
      "resource": "CCAE",
      "faults": [
        {
          "faultId": "fault-instance-001",
          "faultType": "NPU",
          "faultCode": "100000001",
          "faultTime": 1735689601000,
          "assertion": "occur",
          "faultLocation": {
            "npuIp": "192.168.xx.xx"
          },
          "influence": [
            {
              "nodeName": "node1",
              "deviceIds": [0, 1]
            }
          ],
          "description": "NPU chip fault occurred"
        }
      ]
    }
```

**Parameter Description**

See the following table for detailed parameter descriptions.

**Table 1** Fault information

|Parameter Name|Meaning|Value|Type|Mandatory|
|--|--|--|--|--|
|id|Message unique identifier|A string of 8 to 128 characters, supporting uppercase and lowercase letters, digits, hyphens (-), underscores (_), and dots (.), and must be unique.|string|Yes|
|timestamp|Timestamp when the message is sent|Timestamp (unit: ms), a 13-digit number, and must be later than 2025-01-01T00:00:00Z.|int64|Yes|
|version|Message version number|The value is 1.0.|string|Yes|
|resource|Fault sender|The default configuration includes CCAE, fd-online, pingmesh, Netmind, dpcStorage, and dtfsStorage.<ul><li>The fault sender of a public fault must exist in publicFaultResource of the fault configuration file.</li><li>For a newly added fault sender, you need to manually configure it in the fault configuration file. For details, see [(Optional) Configure the Level and Sender of Public Faults](../04_usage/04_resumable_training/03_configuration/01_configuring_fault_detection_levels.md#optional-configuring-the-level-and-sender-of-public-faults).</li></ul>|string|Yes|
|faults|Fault content|Slice, with a length greater than 0 and less than or equal to 100.|[]object, [fault](#fault0023698)|Yes|

**Table 2** fault Field Description

<a name="fault0023698"></a>

|Parameter Name|Meaning|Value|Type|Mandatory|
|--|--|--|--|--|
|faultId|Fault instance ID|A string of 8 to 128 characters, supporting uppercase and lowercase letters, digits, hyphens (-), underscores (_), and dots (.), and must be unique.<p>For the same fault instance, faultId must be unique.</p>|string|Yes|
|faultType|Fault type|The value is NPU, Node, Network, or Storage.<ul><li>NPU: chip fault.</li><li>Node: node fault.</li><li>Network: network fault.</li><li>Storage: storage fault.</li></ul>This field is displayed as "PublicFault" in cluster-info-cm.|string|Yes|
|faultCode|Fault code|User-defined, and 9 digits that are unique are sufficient.<ul><li>A fault code that accesses breakpoint resumable training must exist in publicFaultCode of the fault configuration file.</li><li>For a newly added fault code, its fault level must be configured in the fault configuration file. For details, see [(Optional) Configure the Level and Sender of Public Faults](../04_usage/04_resumable_training/03_configuration/01_configuring_fault_detection_levels.md#optional-configuring-the-level-and-sender-of-public-faults).</li><li>It is recommended that fault codes follow the rules defined in the fault code description table for easier subsequent maintenance.</li><li>If the same fault code appears twice in succession on one NPU, the fault_code field in cluster-info-cm records the two identical fault codes at the same time.</li></ul>|string|Yes|
|faultTime|Fault generation time|Timestamp (unit: ms), 13 digits, and must be after 2025-01-01T00:00:00Z.<ul><li>Whether it is fault generation or fault recovery, this field is always the fault generation time.</li><li>This field is displayed in seconds in cluster-info-cm.</li></ul>|int64|Yes|
|assertion|Fault status|The value is occur, recover, or once.<ul><li>occur: fault generation.</li><li>recover: fault recovery.</li><li>once: one-time event.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>To clear a public fault, the recover event of the corresponding fault must be written into the ConfigMap. It cannot be implemented by deleting the ConfigMap.</li><li>For a one-time event, the fault is automatically cleared after a few seconds.</li></ul></div></div>|string|Yes|
|faultLocation|Fault location information|Fault source information. Length ≤10, the key length of the map ≤16, and the value length ≤128. eg. key: npuIp, value: ip|map[string]string|No|
|influence|Scope affected by the fault|Slice, length >0 and ≤1000.|[]object, [faultInfo](#faultinfo0023698)|Yes|
|description|Fault description|0 to 512 characters. Contains non-whitespace characters and spaces.|string|No|

**Table 3** `faultInfo` description

<a name="faultinfo0023698"></a>

|Parameter Name|Meaning|Value|Type|Mandatory|
|--|--|--|--|--|
|nodeName|Node name. It can be queried using the **kubectl get nodes -owide** command.|A string of 1 to 253 characters, supporting lowercase letters, digits, hyphens (-), and dots (.), and must start and end with an alphanumeric character. When this field exists, nodeSN is not used.<p>If the node name does not exist in the K8s cluster, ClusterD does not report a node name error, but does not write the fault information into cluster-info-device-cm.</p>|string|Either nodeName or nodeSN|
|nodeSN|Node SN|The SN of the node. The value is the node annotation written by NodeD, and the key is product-serial-number.<p>If this field is used instead of nodeName, the NodeD component must be installed in advance.</p>|string|Either nodeName or nodeSN|
|deviceIds|Physical chip ID|Length (0, 32], each element value [0, 32), and duplicates are not allowed.<ul><li>If the faulty chip cannot be accurately located, all physical chip IDs on the node must be filled in.</li><li>If a physical chip ID that does not exist on a node is passed in, ClusterD also displays it in cluster-info-device-cm.</li></ul>|[]int32|Yes|

## Reporting Faults Through the gRPC Interface<a name="ZH-CN_TOPIC_0000002479226854"></a>

**Function**

Receives and processes public fault sending requests from gRPC clients and accesses the resumable training process.

>[!NOTE]
>
>To clear a public fault, the recover event of the corresponding fault must be sent to ClusterD through the gRPC interface.

**Prototype**

```proto
rpc SendPublicFault(PublicFaultRequest) returns (RespStatus){}
```

**Input Parameter Description**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|PublicFaultRequest|<p>message PublicFaultRequest{<p>string id = 1;</p><p>int64 timestamp = 2;</p><p>string version = 3;</p><p>string resource = 4;</p><p>repeated Fault faults = 5;</p>}</p><p>message Fault{<p>string faultId = 1;</p><p>string faultType = 2;</p><p>string faultCode = 3;</p><p>int64 faultTime = 4;</p><p>string assertion = 5;</p><p>map<string, string> faultLocation = 6;</p><p>repeated PubFaultInfo influence = 7;</p><p>string description = 8;</p>}</p><p>message PubFaultInfo{<p>string nodeName = 1;</p><p>string nodeSN = 2;</p><p>repeated int32 deviceIds = 3;</p>}</p>|<p>**PublicFaultRequest.id**: Unique identifier of the message.</p><p>**PublicFaultRequest.timestamp**: Timestamp when the message is sent.</p><p>**PublicFaultRequest.version**: Message version number.</p><p>**PublicFaultRequest.resource**: Fault sender.</p><p>**PublicFaultRequest.faults**: Fault content.</p><p>**Fault.faultId**: Fault instance ID.</p><p>**Fault.faultType**: Fault type.</p><p>**Fault.faultCode**: Fault code.</p><p>**Fault.faultTime**: Fault generation time.</p><p>**Fault.assertion**: Fault status.</p><p>**Fault.faultLocation**: Fault location information.</p><p>**Fault.influence**: Scope affected by the fault.</p><p>**Fault.description**: Fault description.</p><p>**PubFaultInfo.nodeName**: Node name.</p><p>**PubFaultInfo.nodeSN**: Node SN.</p><p>**PubFaultInfo.deviceIds**: Physical chip ID.</p><p>For detailed descriptions and values of the preceding parameters, see [Reporting Faults Through ConfigMap](#reporting-faults-through-configmap).</p>|

**Return Value Description**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|RespStatus|message RespStatus{<p>int32 code = 1;</p><p>string info = 2;</p>}|**RespStatus.code**: Return code.<ul><li>Value 0: indicates that the fault is sent successfully.</li><li>Other values: indicate that the fault fails to be sent. 409 indicates that the request parameters are incorrect, and 410 indicates that the message sending frequency exceeds the limit.</li></ul>**RespStatus.info**: Return information description.|

## Related References

- For the complete API definition of the public fault interface, see [Public Fault Interface](../06_api/04_clusterd/03_public_fault_apis.md).
- For configuring the level and sender of public faults, see [(Optional) Configure the Level and Sender of Public Faults](../04_usage/04_resumable_training/03_configuration/01_configuring_fault_detection_levels.md#optional-configuring-the-level-and-sender-of-public-faults).
