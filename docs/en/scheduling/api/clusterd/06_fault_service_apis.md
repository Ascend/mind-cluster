# Fault Service<a name="ZH-CN_TOPIC_0000002479386826"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:47:11.779Z pushedAt=2026-06-09T02:05:50.703Z -->

## Register<a name="ZH-CN_TOPIC_0000002511426773"></a>

**Description<a name="section143314311911"></a>**

Receives and processes client registration requests, performing initialization preparations for functions such as subscribing to fault information.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**Input Parameters <a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When the passed-in `jobId` is null, it indicates registration of all jobs in the cluster.</li><li>When the passed-in `jobId` is not null, it indicates registration of the specified job.</li></ul></div></div></p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**: Return code.<ul><li>Value `0`: Indicates successful registration.</li><li>Other values: Indicates registration failure.</li></ul>**Status.info**: Return information.|

## SubscribeFaultMsgSignal<a name="ZH-CN_TOPIC_0000002511426699"></a>

**Description<a name="section143314311911"></a>**

Receives a client request to subscribe to fault information. The server allocates a message queue for each job and listens for pending messages in the queue. If messages exist, they are sent to the client via gRPC stream.

>[!NOTE]
>
>- Before calling this API, you must first call [Register](#ZH-CN_TOPIC_0000002511426773).
>- After a client subscribes to fault information for a computing task, it can only receive NodeD faults and K8s node status anomaly faults.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc SubscribeFaultMsgSignal(ClientInfo) returns (stream FaultMsgSignal){}
```

**Input Parameters <a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.<div class="note"><span class="notetitle">Note</span><div class="notebody"><ul><li>When the passed-in `jobId` is null, the obtained result is the faults of all jobs in the cluster.</li><li>When the passed-in `jobId` is not null, the obtained result is the fault of the node to which the job belongs.</li></ul></div></div>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This interface returns a gRPC stream (the specific data structure of the return value depends on the programming language chosen by the client).</li><li>The client can call the stream's `Receive` method (the specific method name depends on the programming language chosen by the client) to receive data pushed by the server.</li></ul>|

**Sending Data<a name="section112224012419"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|FaultMsgSignal|<p>message FaultMsgSignal{</p><p>string uuid = 1;</p><p>string jobId = 2;</p><p>string signalType = 3;</p><p>repeated NodeFaultInfo nodeFaultInfo = 4;</p>}</p><p>message NodeFaultInfo{<p>string nodeName = 1;</p><p>string nodeIP = 2;</p><p>string nodeSN = 3;</p><p>string faultLevel = 4;</p><p>repeated DeviceFaultInfo faultDevice = 5;</p>}</p><p>message DeviceFaultInfo{<p>string deviceId = 1;</p><p>string deviceType = 2;</p><p>repeated string faultCodes = 3;</p><p>string faultLevel = 4;</p><p>repeated string faultType = 5;</p><p>repeated string faultReason = 6;</p><p>repeated SwitchFaultInfo switchFaultInfos = 7;</p><p>repeated string faultLevels = 8;</p>}</p><p>message SwitchFaultInfo{<p>string faultCode = 1;</p><p>string switchChipId = 2;</p><p>string switchPortId = 3;</p><p>string faultTime = 4;</p><p>string faultLevel = 5;</p>}</p>|<p>**FaultMsgSignal.uuid**: Message ID</p><p>**FaultMsgSignal.jobId**: Job ID</p><p>**FaultMsgSignal.signalType**: Message type, "fault" indicates a fault occurrence, "normal" indicates no fault or fault recovery</p><p>**FaultMsgSignal.nodeFaultInfo**: Node fault information</p><p>**NodeFaultInfo.nodeName**: Faulty node name</p><p>**NodeFaultInfo.nodeIP**: Node IP</p><p>**NodeFaultInfo.nodeSN**: Node SN</p><p>**NodeFaultInfo.faultLevel**: Fault type, including "Healthy", "SubHealthy", and "UnHealthy", set to the most severe level in DeviceFaultInfo.faultLevel</p><p>**NodeFaultInfo.faultDevice**: Device fault information</p><p>**DeviceFaultInfo.deviceId**: Device ID. When a bus device fault or K8s status anomaly fault occurs on the node, deviceId is -1</p><p>**DeviceFaultInfo.deviceType**: Device type name, including "Node", "NPU", "Storage", "CPU", "Network", etc.</p><p>**DeviceFaultInfo.faultCodes**: Fault code list</p><p>**DeviceFaultInfo.faultLevel**: Fault type, including "Healthy", "SubHealthy", and "UnHealthy", with severity increasing in order</p><p>**DeviceFaultInfo.faultType**: Fault subsystem type, reserved field</p><p>**DeviceFaultInfo.faultReason**: Fault reason, reserved field</p><p>**DeviceFaultInfo.switchFaultInfos**: UnifiedBus fault information</p><p>**DeviceFaultInfo.faultLevels**: Fault level list</p><p>**SwitchFaultInfo.faultCode**: UnifiedBus fault code</p><p>**SwitchFaultInfo.switchChipId**: UnifiedBus fault chip ID</p><p>**SwitchFaultInfo.switchPortId**: UnifiedBus fault port ID</p><p>**SwitchFaultInfo.faultTime**: UnifiedBus fault occurrence time</p><p>**SwitchFaultInfo.faultLevel**: UnifiedBus fault level</p>|

## GetFaultMsgSignal<a name="ZH-CN_TOPIC_0000002479226874"></a>

**Description<a name="section143314311911"></a>**

This interface is a fault query interface. Its main function is to receive requests from clients to query cluster and task fault information.

>**NOTE**
>This interface can be queried up to 10 times per second. When the limit is exceeded, requests are added to a waiting queue. When the total number of waiting requests exceeds 50, subsequent requests will be rejected.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc GetFaultMsgSignal(ClientInfo) returns (FaultQueryResult){}
```

**Input Parameters <a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Task ID. When jobId is passed in as null, cluster-wide fault information is returned. If jobId is not passed in as null, the valid length of jobId is [8,128] characters, and it cannot contain Chinese characters.</p><p>**ClientInfo.role**: Client role.</p><div class="note"><span class="notetitle">**NOTE** Description</span><div class="notebody"><ul><li>When the passed-in jobId is null, the query result is all faults in the current cluster.</li><li>When the passed-in jobId is not null, the query result is the fault of the node to which the task belongs.</li></ul></div></div>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|FaultQueryResult|<p>message FaultQueryResult{</p><p>int32 code = 1;</p><p>string info = 2;</p><p>FaultMsgSignal faultSignal =3;</p>}|<p>**code**: Return code of this query.<ul><li>200: Query returned normally.</li><li>429: Server-side rate limiting.</li><li>500: Server-side error.</li></ul></p><p>**info**: Description information of this query result</p><p>**faultSignal**: Fault information structure</p><p>**FaultMsgSignal.uuid**: Message ID</p><p>**FaultMsgSignal.jobId**: Task ID, -1 represents the cluster</p><p>**FaultMsgSignal.signalType**: Message type, "fault" represents a fault occurrence, "normal" represents no fault or fault recovery</p><p>**FaultMsgSignal.nodeFaultInfo**: Node fault information</p><p>**NodeFaultInfo.nodeName**: Faulty node name</p><p>**NodeFaultInfo.nodeIP**: Node IP</p><p>**NodeFaultInfo.nodeSN**: Node SN</p><p>**NodeFaultInfo.faultLevel**: Fault type, including "Healthy", "SubHealthy", and "UnHealthy", set to the most severe level in DeviceFaultInfo.faultLevel</p><p>**NodeFaultInfo.faultDevice**: Device fault information</p><p>**DeviceFaultInfo.deviceId**: Device ID</p><p>**DeviceFaultInfo.deviceType**: Device type name, including "Node", "NPU", "Storage", "CPU", "Network", etc.</p><p>**DeviceFaultInfo.faultCodes**: Fault code list</p><p>**DeviceFaultInfo.faultLevel**: Fault type, including "Healthy", "SubHealthy", and "UnHealthy", with severity levels increasing in order</p><p>**DeviceFaultInfo.faultType**: Fault subsystem type, reserved field</p><p>**DeviceFaultInfo.faultReason**: Fault reason, reserved field</p><p>**DeviceFaultInfo.switchFaultInfos**: UnifiedBus fault information list</p><p>**DeviceFaultInfo.faultLevels**: Fault level list</p><p>**SwitchFaultInfo.faultCode**: UnifiedBus fault code</p><p>**SwitchFaultInfo.switchChipId**: UnifiedBus fault chip ID</p><p>**SwitchFaultInfo.switchPortId**: UnifiedBus fault port ID</p><p>**SwitchFaultInfo.faultTime**: UnifiedBus fault occurrence time</p><p>**SwitchFaultInfo.faultLevel**: UnifiedBus fault level</p>|
