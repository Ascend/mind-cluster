# Online Stress Testing<a name="ZH-CN_TOPIC_0000002479226858"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:47:20.059Z pushedAt=2026-06-09T02:05:50.706Z -->

## StressTest<a name="ZH-CN_TOPIC_0000002511426729"></a>

**Description<a name="section143314311911"></a>**

Receives online stress testing requests from the O&M platform and delivers stress testing operations to the specified nodes of a given training job. This interface requires that the training job has already run successfully and produced iterations before being called, ensuring that the job has been registered with ClusterD. The online stress testing interface is a manual O&M operation. Please ensure that the server environment is normal before calling the interface.

**NOTE**:
Please deliver the online stress testing command only after the training has produced normal iterations.

*Prototype<a name="section3958124212115"></a>**

```proto
rpc StressTest(StressTestParam) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|StressTest|<p>message StressTestParam {</p><p>string jobID = 1;</p><p>map<string, StressOpList> stressParam = 2;</p><p>repeated int64 allNodesOps = 3;</p>}<p>message StressOpList {<p>repeated int64 ops = 1;</p>}</p>|<p>**StressTestParam.jobID**: Job ID.</p><p>**StressTestParam.stressParam**: Nodes and operations for which the user issues stress test commands. The key is the node name, and the value is the stress testing operation to be executed on that node.</p><p>**StressTestParam.allNodesOps**: If the user wants to perform stress testing on all nodes of the job, this field indicates the stress testing operation to be executed on all nodes. The `allNodesOps` field has a higher priority than `stressParam`. `0` indicates "AIC" stress testing; `1` indicates "P2P" stress testing.</p><p>**StressOpList.ops**: Stress testing operations to be executed on this node. `0` indicates "AIC" stress testing; `1` indicates "P2P" stress testing.</p>|

**Return Value <a name="section146221236193515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p>}|**Status.code**: Return code.<ul><li>Value `0`: the command was issued successfully.</li><li>Other values: the command  issuance failed.</li></ul>**Status.info**: Return information.|

## SubscribeStressTestResponse<a name="ZH-CN_TOPIC_0000002511346789"></a>

**Description<a name="section143314311911"></a>**

Interface for the O&M platform to query stress testing results. After the O&M personnel successfully issue an online stress testing command, the results can be queried through this interface.

*Prototype<a name="section3958124212115"></a>**

```proto
rpc SubscribeStressTestResponse(StressTestRequest) returns (stream StressTestResponse) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|StressTestRequest|message StressTestRequest{<p>string jobID = 1;</p>}|**StressTestRequest.jobID**: Job ID.|

**Return Value Description<a name="section146221236193515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|StressTestResponse|<p>message StressTestResponse {</p><p>string jobID;</p><p>string msg;</p>}|<p>**StressTestResponse.jobID**: Job ID.</p><p>**StressTestResponse.msg**: Execution result of stress testing.</p>|

## SubscribeNotifyExecStressTest<a name="ZH-CN_TOPIC_0000002479386800"></a>

**Description<a name="section143314311911"></a>**

Receives the client's subscription request for online stress test signals. The server assigns a message queue to each job and listens for pending messages in the queue. If a message exists, it is sent to the client via gRPC stream.

**Prototype <a name="section3958124212115"></a>**

```proto
rpc SubscribeNotifyExecStressTest(ClientInfo) returns (stream StressTestRankParams) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.</p>|

**Sending Data Description<a name="section146221236193515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|StressTestRankParams|<p>message StressTestRankParams {</p><p>map<string, StressOpList> stressParam = 1;</p><p>string jobId = 2;</p>}|<p>**StressTestRankParams.stressParam**: The key is the global rank ID of the node on which the stress testing is to be executed, and the value is the corresponding stress testing operation, where `0` indicates "AIC" stress testing and `1` indicates "p2p" stress testing.</p><p>**StressTestRankParams.jobId**: Job ID.</p>|

**Return Value<a name="section69806312314"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This interface returns a gRPC stream (the specific data structure of the return value depends on the programming language chosen by the client).</li><li>The client can call the stream's `Receive` method (the specific method name depends on the programming language chosen by the client) to receive data pushed by the server.</li></ul>|

## ReplyStressTestResult<a name="ZH-CN_TOPIC_0000002511346775"></a>

**Description<a name="section143314311911"></a>**

Interface for the client to return online stress testing results to ClusterD.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc ReplyStressTestResult(StressTestResult) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|StressTestResult|<p>message StressTestResult {</p><p>string jobId = 1;</p><p>map<string, StressTestRankResult> stressResult = 2;</p>}<p>message StressTestRankResult {<p>map<string, StressTestOpResult> rankResult= 1;</p>}</p><p>message StressTestOpResult {<p>string code = 1;</p><p>string result = 2;</p>}</p>|<p>**StressTestResult.jobId**: Job ID.</p><p>**StressTestResult.stressResult**: Result of the instruction execution. The key is the global rank ID that executed the stress testing; the value is the execution result of stress testing.</p><p>**StressTestRankResult.rankResult**: Result of the stress testing executed on a specific card. The key is the stress testing operation, where `0` indicates "AIC" stress testing and `1` indicates "P2P" stress testing.</p><p>**StressTestOpResult.code**: Error code of the stress testing result.<ul><li>`0` indicates successful execution; no fault</li><li>`1` indicates stress testing failure; training can be resumed normally</li><li>`2` indicates a stress testing fault detected; the corresponding node needs to be isolated</li><li>`3` indicates stress testing timeout; the job exits and restarts</li><li>`4` indicates stress testing voltage not recovered; the job exits and restarts</li></ul></p><p>**StressTestOpResult.result**: Description of the stress testing result.</p>|

**Return Value <a name="section69806312314"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**: Return code.<ul><li>Value `0`: normal process</li><li>Other values: abnormal process</li></ul>**Status.info**: Return information.|
