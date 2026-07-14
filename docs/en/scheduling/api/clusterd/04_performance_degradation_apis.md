# Performance Degradation<a name="ZH-CN_TOPIC_0000002479226802"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:46:30.161Z pushedAt=2026-06-09T02:05:50.687Z -->

## ModifyTrainingDataTraceSwitch<a name="ZH-CN_TOPIC_0000002511426771"></a>

**Description<a name="section22878209356"></a>**

External call to modify the capability of various dynamic data tracing.

>[!NOTE]
>If lightweight profiling is enabled or modified through the gRPC interface provided by ClusterD to obtain data written to disk, the lifecycle of the created `data-trace-<Job Name>` ConfigMap will be deleted along with the job. When the job does not exist, this interface call will fail.

**Prototype<a name="section1472624833519"></a>**

```proto
rpc ModifyTrainingDataTraceSwitch(DataTypeReq) returns (DataTypeRes) {}
```

**Input Parameters<a name="section6782115723515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|DataTypeReq|<p>message DataTypeReq{<p>string jobNsName = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p>}</p><p>message ProfilingSwitch{<p>string CommunicationOperator = 1;</p><p>string Step = 2;</p><p>string SaveCheckpoint = 3;</p><p>string FP =4;</p><p>string DataLoader =5;</p>}</p>|<p>**jobNsName**: The namespace and name of the job to be modified, concatenated with '/', for example: `default/test-pytorch`.</p><p>**profilingSwitch**: Details of various switches.</p><ul><li>**CommunicationOperator**: Communication operators.</li><li>**Step**: Step latency.</li><li>**SaveCheckpoint**: SaveCheckpoint time consumption.</li><li>**FP**: Forward propagation.</li><li>**DataLoader**: DataLoader time consumption.</li></ul>|

**Return Value<a name="section7920469381"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|DataTypeRes|message DataTypeRes{<p>string message = 1;</p><p>int32 code = 2;</p>}|<p>**message**: Interface call result.</p><p>**code**: Interface call return code.</p><ul><li>`300`: invalid input parameters.</li><li>`404`: unable to query ConfigMap.</li><li>`500`: server exception.</li><li>`200`: interface returned normally.</li></ul>|

## GetTrainingDataTraceSwitch<a name="ZH-CN_TOPIC_0000002479386852"></a>

**Description<a name="section21882190424"></a>**

External call to obtain the status of various dynamic data tracing.

**Prototype<a name="section1723573217426"></a>**

```proto
rpc GetTrainingDataTraceSwitch(DataStatusReq) returns (DataStatusRes) {}
```

**Input Parameters<a name="section19921040164215"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|DataStatusReq|message DataStatusReq{<p>string jobNsName = 1;</p>}|**jobNsName**: The namespace and name of the job to be modified, concatenated with '/', e.g., `default/test-pytorch`.|

**Return Value Description<a name="section93011951104217"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|DataStatusRes|message DataStatusRes{<p>string message = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p><p>int32 code = 3;</p>}|<p>**message**: Result information of the interface call.</p><p>**profilingSwitch**: Details of various switches.</p><ul><li>**CommunicationOperator**: Communication operator.</li><li>**Step**: Step latency.</li><li>**SaveCheckpoint**: SaveCheckpoint time consumption.</li><li>**FP**: Forward propagation.</li><li>**DataLoader**: DataLoader time consumption.</li></ul>**code**: Return code of the interface call.<ul><li>`300`: invalid input parameters.</li><li>`404`: unable to query ConfigMap.</li><li>`500`: server exception.</li><li>`200`: interface returned normally.</li></ul>|

## SubscribeDataTraceSwitch<a name="ZH-CN_TOPIC_0000002511346751"></a>

**Description<a name="section22878209356"></a>**

Externally subscribe to the status of various dynamic data tracing switches.

**Prototype<a name="section1472624833519"></a>**

```proto
rpc SubscribeDataTraceSwitch(ProfilingClientInfo) returns (stream DataStatusRes) {}
```

**Input Parameters<a name="section6782115723515"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ProfilingClientInfo|message ProfilingClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**jobId**: Job ID.</p><p>**role**: Client role.</p>|

**Return Value Description<a name="section7920469381"></a>**

| Parameter | Type (Protobuf Definition) | Description |
|--|--|--|
| DataStatusRes | message DataStatusRes{<p>string message = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p><p>int32 code = 3;</p>} | <p>**message**: Information about the interface call result.</p><p>**profilingSwitch**: Details of various switches.</p><ul><li>**CommunicationOperator**: Communication operator.</li><li>**Step**: Step latency.</li><li>**SaveCheckpoint**: SaveCheckpoint time consumption.</li><li>**FP**: Forward propagation.</li><li>**DataLoader**: DataLoader time consumption.</li></ul>**code**: Return code of the interface call.<ul><li>`300`: invalid input parameters.</li><li>`404`: unable to query ConfigMap.</li><li>`500`: server exception.</li><li>`200`: interface returned normally.</li></ul> |
