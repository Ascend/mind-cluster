# Job Information<a name="ZH-CN_TOPIC_0000002511426731"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:47:11.499Z pushedAt=2026-06-09T02:05:50.701Z -->

## Description<a name="ZH-CN_TOPIC_0000002479386816"></a>

This module provides external subscription interfaces for obtaining cluster job status and basic information. It consists of three interfaces: `Register`, `SubscribeJobSummarySignal`, and `SubscribeJobSummarySignalList`.

**Call Sequence <a name="section171351329174616"></a>**

Before calling the `SubscribeJobSummarySignal` and `SubscribeJobSummarySignalList` interfaces, users must first obtain a valid client ID through the `Register` interface, and then use that ID to call the subscription interfaces.

The subscription interfaces automatically close if there is no activity within two minutes by default.

## Register<a name="ZH-CN_TOPIC_0000002479226804"></a>

**Description <a name="section14645125754213"></a>**

Receives registration requests from clients in the scenario of subscribing to job information.

If a client needs to subscribe to cluster job information, it must first call this interface to obtain the returned UUID, and then use that ID to call `SubscribeJobSummarySignal` and `SubscribeJobSummarySignalList` to obtain cluster job information.

>[!NOTE]
>A maximum of 80 active subscription links can exist in a cluster, and each client role can create a maximum of 20 active subscription links.

**FPrototype<a name="section4140960433"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**Input Parameters<a name="section1317321424310"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string role = 1;</p><p>string clientId = 3;</p>}|**ClientInfo.role**: Client role. Currently, only the following client roles are supported. If other values are passed in, registration will fail.<ul><li>CCAgent</li><li>DefaultUser1</li><li>DefaultUser2</li><li>FdAgent</li></ul>**ClientInfo.clientId**: Client ID.|

**Return Value<a name="section4839929184717"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p><p>string clientId = 3;</p>}|**Status.code**: The status code of this call result. Currently divided into the following: <ul><li>200: Query returned normally.</li><li>429: Server-side rate limiting.</li><li>500: Server-side error.</li></ul><p>**Status.info**: The description information of this call result.</p><p>**Status.clientId**: The UUID returned by the registration interface.</p>|

## SubscribeJobSummarySignal<a name="ZH-CN_TOPIC_0000002511426723"></a>

**Description<a name="section85381247165120"></a>**

Receives task information change subscriptions from the client. When the client initially subscribes to the interface, it pushes the information of all tasks in the current cluster one by one. When the task status changes, it broadcasts and pushes to the registered clients. When there is no message and no heartbeat within a connection for two minutes, the server actively disconnects the connection and releases the subscription.

>[!NOTE]
>
>- This interface has a rate limiting mechanism, allowing a maximum of 20 accesses per second.
>- A maximum of 80 active subscription links can exist within the cluster, and each client role supports a maximum of 20 active subscription links.

**Prototype<a name="section1199205575113"></a>**

```proto
rpc SubscribeJobSummarySignal(ClientInfo) returns (stream JobSummarySignal){}
```

**Input Parameters<a name="section6291133165212"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string role = 1;</p><p>string clientId = 3;</p>}|**ClientInfo.role**: Client role. Currently, only the following client roles are supported. If other values are passed, the registration will fail.<ul><li>CCAgent</li><li>DefaultUser1</li><li>DefaultUser2</li><li>FdAgent</li></ul>**ClientInfo.clientId**: Client ID.|

**Return Value<a name="section1883821810542"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This interface returns a gRPC stream (the specific data structure of the return value depends on the programming language selected by the client).</li><li>The client can call the stream's `Receive` method (the specific method name depends on the programming language selected by the client) to receive data pushed by the server.</li></ul>|

**Sending Data<a name="section10140143475520"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|JobSummarySignal|<p>message JobSummarySignal{</p><p>string uuid = 1;</p><p>string jobId = 2;</p><p>string jobName = 3;</p><p>string namespace =4;</p><p>string frameWork = 5;</p><p>string jobStatus = 6;</p><p>string time = 7;</p><p>string cmIndex = 8;</p><p>string total = 9;</p><p>string HcclJson = 10;</p><p>string deleteTime = 11;</p><p>string sharedTorIp = 12;</p><p>string masterAddr = 13;</p><p>string operator = 14;</p><p>string sid = 15;</p>}|<p>**uuid**: ID of this message</p><p>**jobId**: K8s ID information of the job</p><p>**jobName**: Name of the current job</p><p>**namespace**: Namespace to which the job belongs</p><p>**frameWork**: Job framework</p>**jobStatus**: Job status, which can be one of the following: <ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul><p>**time**: Job start time</p><p>**cmIndex**: Sequence number</p><p>**total**: Total number of jobsummary ConfigMaps corresponding to the job</p><p>**HcclJson**: Chip communication information used by the job. If the number of NPUs scheduled for the job exceeds 40,000, HcclJson in the reported information received by the client will be set to empty.<p>It can be escaped into JSON format, with the following field descriptions:</p><ul><li>status: Whether the job RankTable has been generated</li><li>initializing: Devices are still being allocated for the job, and the RankTable has not been generated</li><li>complete: When the RankTable is generated, the status immediately changes to complete, and other fields such as server_list appear synchronously</li><li>server_list: Job device allocation</li><li>device: Records NPU allocation, NPU IP, and rank_id information</li><li>server_id: AI Server identifier, globally unique</li><li>server_name: Node name</li><li>server_sn: SN of the node. The device SN must exist. If it does not exist, contact Huawei technical support</li><li>server_count: Number of nodes used by the job</li><li>version: Version information</li></ul></p><p>**deleteTime**: Time when the job was deleted</p><p>**sharedTorIp**: Shared switch information used by the job</p><p>**masterAddr**: MASTER_ADDR value specified during PyTorch training</p><p>**operator**: Status updates to add after receiving a job addition command; status updates to delete after receiving a job deletion command</p><p>**sid**: Unique job identifier</p>|

## SubscribeJobSummarySignalList

**Description**

Receives job information change subscriptions from the client. When the client subscribes to the interface for the first time, information of all jobs in the current cluster is pushed. When the job status changes, a broadcast push is sent to the registered clients. If there are no messages and no heartbeat within two minutes of connection, the server actively disconnects and releases the subscription.

>[!NOTE]
>
>- This API has a rate limiting mechanism, allowing a maximum of 20 accesses per second.
>- A cluster can have up to 80 active subscription links, with each client role supporting a maximum of 20 active subscription links.

**Prototype**

```proto
rpc SubscribeJobSummarySignalList(ClientInfo) returns (stream JobSummarySignalList){}
```

**Input Parameters**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string role = 1;</p><p>string clientId = 3;</p>}|**ClientInfo.role**: Client role. Currently, only the following client roles are supported. Passing other values will cause registration to fail.<ul><li>CCAgent</li><li>DefaultUser1</li><li>DefaultUser2</li><li>FdAgent</li></ul><p>**ClientInfo.clientId**: Client ID.</p>|

**Return Value**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This interface returns a gRPC stream (the specific data structure of the return value depends on the programming language selected by the client).</li><li>The client can call the stream's `Receive` method (the specific method name depends on the programming language selected by the client) to receive data pushed by the server.</li></ul>|

**Sending Data**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|JobSummarySignalList|<p>message JobSummarySignalList{</p><p>repeated JobSummarySignal jobSummarySignals = 1;</p><p>string ReportTime = 2;</p><p>int32 JobTotalNum = 3;</p>}<p>message JobSummarySignal{<p>string uuid = 1;</p><p>string jobId = 2;</p><p>string jobName = 3;</p><p>string namespace =4;</p><p>string frameWork = 5;</p><p>string jobStatus = 6;</p><p>string time = 7;</p><p>string cmIndex = 8;</p><p>string total = 9;</p><p>string HcclJson = 10;</p><p>string deleteTime = 11;</p><p>string sharedTorIp = 12;</p><p>string masterAddr = 13;</p><p>string operator = 14;</p><p>string sid = 15;</p>}</p>|<p>**jobSummarySignals**: Job information list</p><p>**ReportTime**: Time when the current batch is reported</p><p>**JobTotalNum**: Total number of jobs reported in the same batch</p><p>**uuid**: ID of this message</p><p>**jobId**: K8s ID information of the job</p><p>**jobName**: Name of the current job</p><p>**namespace**: Namespace to which the job belongs</p><p>**frameWork**: Job framework</p>**jobStatus**: Job status, which can be one of the following:<ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul><p>**time**: Job start time</p><p>**cmIndex**: Sequence number</p><p>**total**: Total number of jobsummary ConfigMaps corresponding to the job</p><p>**HcclJson**: Chip communication information used by the job. It can be escaped into JSON format. Field descriptions are as follows:<ul><li>status: Whether the job RankTable has been generated</li><li>initializing: Devices are still being allocated for the job, and the RankTable has not been generated</li><li>complete: When the RankTable is generated, the status immediately changes to complete, and other fields such as server_list appear synchronously</li><li>server_list: Job device allocation details</li><li>device: Records NPU allocation, NPU IP, and rank_id information</li><li>server_id: AI Server identifier, globally unique</li><li>server_name: Node name</li><li>server_sn: SN of the node. Ensure that the device SN exists. If it does not exist, contact Huawei technical support</li><li>server_count: Number of nodes used by the job</li><li>version: Version information</li></ul></p><div class="note"><span class="notetitle">NOTE</span><div class="notebody"><ul><li>If the number of NPUs used by a single job exceeds 40,000, HcclJson in the reported job information will be set to empty.</li><li>When the client initially subscribes to the interface, if the total number of NPUs used by multiple jobs exceeds 40,000, the reported information will be paginated to ensure that the total number of NPUs in each reported message does not exceed 40,000.</li></ul></div></div><p>**deleteTime**: Time when the job was deleted</p><p>**sharedTorIp**: Shared switch information used by the job</p><p>**masterAddr**: MASTER_ADDR value specified during PyTorch training</p><p>**operator**: Status updates to add after receiving an add job command; status updates to delete after receiving a delete job command</p><p>**sid**: Job unique identifier</p>|
