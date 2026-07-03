# Link Failover and Switchback<a name="ZH-CN_TOPIC_0000002511426725"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:47:25.643Z pushedAt=2026-06-09T02:05:50.708Z -->

## SwitchNicTrack<a name="ZH-CN_TOPIC_0000002511346727"></a>

**Description<a name="section143314311911"></a>**

Receives link failover requests from the O&M platform, and issues link failover/switchback operations to the device of the specified node of the training job. This interface needs to wait until the training job has run successfully and iterations have been produced before being called, to ensure that the job has been registered with ClusterD. The link failover/switchback interface is a manual O&M operation. For repeated switching scenarios, if each switching fails, it will cause frequent CKPT saving, posing a risk of disk space exhaustion.

>[!NOTE]
>Please issue the switchback command or link failover command only after the training iterations have become normal.

**Function Prototype<a name="section3958124212115"></a>**

```proto
rpc SwitchNicTrack(SwitchNics) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|SwitchNics|<p>message SwitchNics{</p><p>string jobID;</p><p>map<string, DeviceList> nicOps;</p>}<p>message DeviceList {<p>repeated string dev;</p><p>repeated bool op;</p>}</p>|<p>**SwitchNics.jobID**: Job ID.</p><p>**SwitchNics.nicOps**: Devices and operations for which the user issues link failover/switchback Commands. The key is the node name, and the value is the Device to be operated on that node.</p><p>**DeviceList.dev**: List of device IDs on this node. The number must match that of DeviceList.op.</p><p>**DeviceList.op**: List of link failover operations to be performed for the devices corresponding to the DeviceIDs on this node. `true` indicates switching to the standby link, and `false` indicates using the primary link.</p>|

**Return Value Description<a name="section146221236193515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p>}|**Status.code**: Return code.<ul><li>Value `0`: Indicates that the command was issued successfully.</li><li>Other values: Indicates that the issuance failed.</li></ul>**Status.info**: Return information.|

## SubscribeSwitchNicSignal<a name="ZH-CN_TOPIC_0000002479226844"></a>

**Description<a name="section143314311911"></a>**

Interface for the O&M platform to query the results of link failover/switchback. After the O&M personnel successfully issue an active link failover/switchback command, the results can be queried through this interface.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc SubscribeSwitchNicSignal(SwitchNicRequest) returns (stream SwitchNicResponse) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|SwitchNicRequest|<p>message SwitchNicRequest{</p><p>string jobID;</p>}|**SwitchNicRequest.jobID**: Job ID|

**Return Value Description<a name="section146221236193515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|SwitchNicResponse|message SwitchNicResponse{<p>string jobID;</p><p>string msg;</p>}|<p>**SwitchNicResponse.jobID**: Job ID</p><p>**SwitchNicResponse.msg**: Execution result of the link failover/switchback Command</p>|

## SubscribeNotifySwitch<a name="ZH-CN_TOPIC_0000002511346769"></a>

**Description<a name="section143314311911"></a>**

Receives client subscription requests for link failover/switchback signals. The server assigns a message queue to each job and listens for pending messages in the queue. If messages exist, they are sent to the client via gRPC stream.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc SubscribeNotifySwitch(ClientInfo) returns (stream SwitchRankList) {}
```

**Input Parameters Description<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.</p>|

**Sending Data<a name="section146221236193515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|SwitchRankList|<p>message SwitchRankList{</p><p>repeated string rankID = 1;</p><p>repeated bool op = 2;</p><p>string jobId = 3;</p>}|<p>**SwitchRankList.rankID**: List of device IDs on this node, with the same number of entries as `DeviceList.op`.</p><p>**SwitchRankList.op**: The list of link failover operations to be performed for the devices corresponding to the device IDs on this node. `true` indicates switching to the standby link, `false` indicates using the primary link.</p><p>**SwitchRankList.jobId**: Job ID</p>|

**Return Value <a name="section69806312314"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This interface returns a gRPC stream (the specific data structure of the return value depends on the programming language chosen by the client).</li><li>The client can call the stream's `Receive` method (the specific method name depends on the programming language chosen by the client) to receive data pushed by the server.</li></ul>|

## ReplySwitchNicResult<a name="ZH-CN_TOPIC_0000002479386790"></a>

**Description<a name="section143314311911"></a>**

Interface for the client to return the link failover/switchback result to ClusterD.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc ReplySwitchNicResult(SwitchResult) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|SwitchResult|message SwitchResult{<p>string jobId = 1;</p><p>bool result = 2;</p>}|<p>**SwitchResult.jobId**: Job ID.</p><p>**SwitchResult.result**: Result of command execution: `true` for success, `false` for failure.</p>|

**Return Value<a name="section69806312314"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**: Return code.<ul><li>Value `0`: Normal process</li><li>Other values: Abnormal process</li></ul>**Status.info**: Return information.|
