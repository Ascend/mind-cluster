# Process-Level Recovery<a name="ZH-CN_TOPIC_0000002511346765"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:46:04.941Z pushedAt=2026-06-09T02:05:50.684Z -->

## gRPC Interface<a name="ZH-CN_TOPIC_0000002511346735"></a>

### Register (Public Interface)<a name="ZH-CN_TOPIC_0000002511346739"></a>

**Description<a name="section143314311911"></a>**

Receives and processes client registration requests, and prepares initialization for process-level recovery. After the job successfully calls `Init`, it must wait for the client to confirm that the process-level rescheduling and process-level online recovery on the MindIO side are enabled before calling this interface. After `Register` succeeds, the process-level rescheduling and process-level online recovery functions become available.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID</p><p>**ClientInfo.role**: Client role</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return code.<ul><li>`0`: Registration successful.</li><li>Other values: Registration failed.</li></ul></p><p>**Status.info**: Return information.</p>|

### Init<a name="ZH-CN_TOPIC_0000002479386824"></a>

**Description<a name="section83882209338"></a>**

Used to initialize process-level rescheduling and process-level online recovery. After successful initialization, the process-level rescheduling and process-level online recovery functions will be temporarily unavailable.

**Prototype<a name="section2049633816332"></a>**

```proto
rpc Init(ClientInfo) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID</p><p>**ClientInfo.role**: Client role</p>|

**Return Value<a name="section1864651893415"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return code.<ul><li>`0`: Registration successful.</li><li>Other values: Registration failed.</li></ul></p><p>**Status.info**: Return information.</p>|

### SubscribeProcessManageSignal<a name="ZH-CN_TOPIC_0000002511426713"></a>

**Description<a name="section143314311911"></a>**

Receives client subscription requests for process control signals. The server allocates a message queue for each job and listens for pending messages in the queue. If a message exists, it is sent to the client via the gRPC stream.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc SubscribeProcessManageSignal(ClientInfo) returns (stream ProcessManageSignal){}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.</p>|

**Sending Data Description<a name="section10140143475520"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ProcessManageSignal|<p>message FaultRank{<p>string rankId = 1;</p><p>string faultType = 2;</p>}</p><p>message ProcessManageSignal{<p>string uuid=1;</p><p>string jobId = 2;</p><p>string signalType = 3;</p><p>repeated string actions = 4;</p><p>repeated FaultRank faultRanks = 5;</p><p>string changeStrategy = 6;</p><p>int64 timeout = 7;</p>}</p>|<p>**rankId**: String, ID of the faulty rank</p><p>**faultType**: String, fault type</p><p>**uuid**: String, UUID of this signal</p><p>**jobId**: String, training job ID</p><p>**signalType**: String, signal type</p><p>**actions**: Repeated string, actions to be executed</p><p>**faultRanks**: Repeated FaultRank, information about faulty ranks</p><p>**changeStrategy**: String, recovery strategy to be executed</p><p>**timeout**: int64, timeout duration</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This interface returns a gRPC stream (the specific data structure of the return value depends on the programming language chosen by the client).</li><li>The client can call the stream's Receive method (the specific method name depends on the programming language chosen by the client) to receive data pushed by the server.</li></ul>|
|nodeRankIds|string array|Node Rank ID of the faulty node.|
|extraParams|string|Passes specific scaling policy information in the form of a JSON string, which is transparently transmitted to MindIO via TaskD and ultimately passed to the callback function for parsing.|

### ReportStopComplete<a name="ZH-CN_TOPIC_0000002511426707"></a>

**Description<a name="section143314311911"></a>**

Receives the report from the client on whether the training process was successfully paused.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc ReportStopComplete(StopCompleteRequest) returns (Status){}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|StopCompleteRequest|message StopCompleteRequest{<p>string jobId = 1;</p><p>Status status = 2;</p><p>repeated FaultRank faultRankIds = 3;</p>}|<p>**StopCompleteRequest.jobId**: Job ID.</p><p>**StopCompleteRequest.status.code**: Return code. `OK` indicates that the training process was paused successfully; other values indicate that the pause failed.</p><p>**StopCompleteRequest.status.info**: Return information.</p><p>**StopCompleteRequest.faultRankIds**: List of global fault ranks for the faulty chips. `FaultRank` is a set of key-value pairs containing fault information, consisting of `rankId` (global rank ID) and `faultType` (fault type). `faultType = 0` indicates an on-chip Memory fault; `faultType = 1`, it indicates other faults; `faultType = 2` indicates a network fault.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return code.<ul><li>`0`: the fault recovery process is normal.</li><li>Other values: the fault recovery process is abnormal, triggering rescheduling.</li></ul></p><p>**Status.info**: Return information.</p>|

### ReportRecoverStrategy<a name="ZH-CN_TOPIC_0000002511346747"></a>

**Description<a name="section16150748174520"></a>**

Receives the fault recovery strategy supported by the current job reported by the client.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc ReportRecoverStrategy(RecoverStrategyRequest) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|RecoverStrategyRequest|message RecoverStrategyRequest{<p>string jobId = 1;</p><p>repeated FaultRank faultRankIds = 2;</p><p>repeated string strategies = 3;</p>}|<p>**RecoverStrategyRequest.jobId**: Job ID</p><p>**RecoverStrategyRequest.faultRankIds**: List of global fault ranks for faulty chips. `FaultRank` is a set of key-value pairs containing fault information, consisting of `rankId` (global rank ID) and `faultType` (fault type). `faultType = 0` indicates an on-chip Memory fault; `faultType = 1`, it indicates other faults; `faultType = 2` indicates a network fault.</p><p>**RecoverStrategyRequest.strategies**: Recovery strategies supported by the Current job.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return Code.<ul><li>`0`: the fault recovery process is normal.</li><li>Other values: the recovery process is abnormal, triggering rescheduling.</li></ul></p><p>**Status.info**: Return information.</p>|

### ReportRecoverStatus<a name="ZH-CN_TOPIC_0000002511346753"></a>

**Description<a name="section16150748174520"></a>**

Receives the current job recovery status reported by the client.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc ReportRecoverStatus(RecoverStatusRequest) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|RecoverStatusRequest|message RecoverStatusRequest{<p>string jobId = 1;</p><p>Status status = 2;</p><p>string strategy = 3;</p><p>repeated string isolateRankIds = 4;</p>}|<p>**RecoverStatusRequest.jobId**: Job ID.</p><p>**RecoverStatusRequest.status.code**: Job recovery status code.<ul><li>`0`: job recovery is successful.</li><li>Other values: job recovery failed.</li></ul></p><p>**RecoverStatusRequest.status.info**: Job recovery status.</p><p>**RecoverStatusRequest.strategy**: Recovery strategy name.</p><p>**RecoverStatusRequest.isolateRankIds**: List of ranks to be isolated when MindIO reports scale-in messages.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return code.<ul><li>`0`: the fault recovery process is normal.</li><li>Other values: the recovery process is abnormal, triggering rescheduling.</li></ul></p><p>**Status.info**: Return information.</p>|

### ReportProcessFault<a name="ZH-CN_TOPIC_0000002511346729"></a>

**Description<a name="section16150748174520"></a>**

Receives the global rank information of the faulty chips reported by the client.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc ReportProcessFault(ProcessFaultRequest) returns (Status){}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ProcessFaultRequest|message ProcessFaultRequest{<p>string jobId = 1;</p><p>repeated FaultRank faultRankIds = 2;</p>}|<p>**ProcessFaultRequest.jobId**: Job ID.</p><p>**ProcessFaultRequest.faultRankIds**: List of global rank IDs of faulty chips. `FaultRank` is a set of key-value pairs containing fault information, consisting of `rankId` (global rank ID) and `faultType` (fault type). `faultType = 0` indicates an on-chip Memory fault; `faultType = 1`, it indicates other faults; `faultType = 2` indicates a network fault.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return code.<li>`0`: the recovery process is normal.</li><li>Other values: the fault recovery process is abnormal, triggering rescheduling.</li></p><p>**Status.info**: Return information.</p>|

### HealthCheck<a name="ZH-CN_TOPIC_0000002511426765"></a>

**Description<a name="section16150748174520"></a>**

Checks the gRPC connection status.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc HealthCheck(ClientInfo) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**: Return code.<li>`0`: the fault recovery process is normal.</li><li>Other values: the fault recovery process is abnormal, triggering rescheduling.</li></p><p>**Status.info**: Return information.</p>|

## Interfaces for Integrating with Third-Party AI Platforms<a name="ZH-CN_TOPIC_0000002511346803"></a>

**Description<a name="section3323912134319"></a>**

An AI platform can control the fault recovery process and recovery strategy through Pod Group Annotations. For example, when the platform writes the Pod Group Annotation key: `ProcessRecoverStrategy` with an empty value, the fault recovery will be blocked until the platform writes a specific recovery strategy to continue the recovery process.

**Pod Group Annotation<a name="section1313991113387a"></a>**

**Table 1**  Parameters

|Parameter|Value|Description|
|--|--|--|
|ProcessRecoverStrategy|<ul><li>retry</li><li>recover</li><li>dump</li><li>Empty or none</li><li>Field does not exist</li></ul>|<ul><li>`retry`: The platform initiates recovery, with the strategy being process-level online recovery</li><li>`recover`: The platform initiates recovery, with the strategy being online recovery</li><li>`dump`: The platform initiates recovery, with the strategy being saving dying gasps</li><li>Empty or none: Waiting for platform decision</li><li>Field does not exist: Disable process-level recovery</li></ul>|
|ProcessConfirmFault|string|A list of fault key-value pairs refreshed by ClusterD, formatted as a string of "id1:type1,id2:type2". `id` represents the global rank ID, and `type` represents the fault type. `type = 0` indicates that the faulty chip only has on-chip memory faults, and `type = 1` indicates at least one non-on-chip memory fault.|
|ProcessResultFault|string|A list of fault key-value pairs confirmed by the platform, formatted as a string of "id1:type1,id2:type2". `id` represents the global rank ID, and `type` represents the fault type. `type = 0` indicates that the faulty chip only has on-chip memory faults, and `type = 1` indicates at least one non-on-chip memory fault.|
|RankTableReady|<ul><li>true</li><li>false or other values</li><li>Field does not exist</li></ul>|<ul><li>`true`: The platform has generated the RankTable</li><li>`false` or other values: The platform has not yet generated the RankTable</li><li>Field does not exist: Non-RankTable mode</li></ul>|
|ProcessRecoverStatus|<ul><li>retry-success</li><li>retry-failed</li><li>recover-success</li><li>recover-failed</li><li>dump-success</li><li>dump-failed</li><li>exit-completed</li><li>Empty or other values</li></ul>|<ul><li>`retry-success`: Process-level online recovery succeeded</li><li>`retry-failed`: Process-level online recovery failed</li><li>`recover-success`: Online recovery succeeded</li><li>`recover-failed`: Online recovery failed</li><li>`dump-success`: Saving dying gasps succeeded</li><li>`dump-failed`: Saving dying gasps failed</li><li>`exit-completed`</li><li>Empty or other values: Recovery not completed</li></ul>|
