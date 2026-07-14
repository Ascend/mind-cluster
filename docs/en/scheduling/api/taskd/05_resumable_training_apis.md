# Resumable Training APIs<a name="ZH-CN_TOPIC_0000002479226856"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:44:07.212Z pushedAt=2026-06-09T02:05:50.661Z -->

## taskd.python.toolkit.recover_module.recover_manager. DLRecoverManager (Internal, Do Not Call)<a name="ZH-CN_TOPIC_0000002479386778"></a>

**Function Description<a name="section95016253292"></a>**

The `DLRecoverManager` class provides APIs related to process-level recovery and process-level online recovery. The client imports it into the client code as a Python package.

>[!NOTE]
>The APIs provided by the `DLRecoverManager` class may throw exceptions. The caller is responsible for catching and handling these exceptions.

**__init__(self, info: pb.ClientInfo, server_addr: str)<a name="section93535281517"></a>**

Constructs a `DLRecoverManager` for subsequent communication.

**Table 1** Parameters

|Parameter|Type|Description|
|--|--|--|
|info|pb.ClientInfo|<p>info.jobId: str type, the job ID.</p><p>info.role: str type, the client role.</p>|
|server_addr|str|Server address|

**register(self, request: pb.ClientInfo) -> pb.Status<a name="section92911329181515"></a>**

Registers a client. The server performs pre-recovery initialization operations for the job specified by the request.

**Table 2**  Parameters

|Parameter|Type|Description|
|--|--|--|
|request|pb.ClientInfo|<p>request.jobId: str type, job ID.</p><p>request.role: str type, client role.</p>|

**Table 3**  Return value description

|Return Value Type|Description|
|--|--|
|Status|<p>Status.info: str type, return information description.</p><p>Status.code: int type. `0` indicates success; other values indicate failure. For details about return codes, see [Return Codes](./07_return_codes.md).</p>|

**def start_subscribe(self, frame: str = "pytorch")<a name="section5051271214"></a>**

The client and server establish a gRPC persistent connection, through which the server communicates unidirectionally with the client. For example, when a fault occurs, the server sends the client information such as a training stop signal and global faulty process rank.

**Table 4**  Parameters

|Parameter|Type|Description|
|--|--|--|
|frame|str|Indicates the AI framework used by the job.|

**init_clusterd(self)<a name="section18270133519256"></a>**

Initializes the ClusterD server status on the client side to ensure normal registration and connection establishment for subsequent jobs.

## report_stop_complete(code: int, msg: str, fault_ranks: dict) -> int<a name="ZH-CN_TOPIC_0000002479386796"></a>

**Function Description<a name="section1620210127300"></a>**

The client reports to the server that the job process has stopped. Typically, after the client receives a stop training signal from the server, the client stops the training job process and then reports to the server that the job process has stopped.

**Input Parameters<a name="section1793816299304"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|
|--|--|--|
|code|int|Status code|
|msg|str|Return information|
|fault_ranks|dict|Rank fo the faulty process|

**Return Value<a name="section924216017310"></a>**

**Table 2** Return value description

|Return Value Type|Description|
|--|--|
|int|`0` indicates success, and other return values indicate failure. For details about return codes, see [Return Codes](./07_return_codes.md).|

## report_recover_strategy(fault_ranks: dict, strategy_list: list) -> int<a name="ZH-CN_TOPIC_0000002479386838"></a>

**Function Description<a name="section350336124214"></a>**

The client provides the server with the recovery strategies it supports, so that the server can select the optimal recovery strategy. The server then sends it to the client through the persistent connection established by `start_subscribe`.

**Input Parameters<a name="section91358261429"></a>**

**Table 1** Parameter Description

| Parameter | Type | Description |
|--|--|--|
| fault_ranks | dict | Rank of the faulty process |
| strategy_list | list | Recovery strategy list |

**Return Value<a name="section1365711594319"></a>**

**Table 2** Return value description

| Return Value Type | Description |
|--|--|
| int | `0` indicates success, and other return values indicate failure. For details about return codes, see [Return Codes](./07_return_codes.md). |

## report_recover_status(code: int, msg: str, fault_ranks: dict, strategy: str) -> int<a name="ZH-CN_TOPIC_0000002479226842"></a>

**Function Description<a name="section9417169184510"></a>**

The client reports the job recovery status to the server.

**Input Parameters<a name="section7968321124510"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|
|--|--|--|
|code|int|Status code|
|msg|str|Return information|
|fault_ranks|dict|Rank of the faulty process|
|strategy|str|Recovery strategy|

**Return Value<a name="section1365711594319"></a>**

**Table 2** Return value description

|Return Value Type|Description|
|--|--|
|int|`0` indicates success, and other return values indicate failure. For details about return codes, see [Return Codes](./07_return_codes.md).|

## report_process_fault(fault_ranks: dict) -> int<a name="ZH-CN_TOPIC_0000002511426703"></a>

**Function Description<a name="section3468140175411"></a>**

The client reports a service-plane fault in the job process. When the client detects a fault first, it reports the information of the rank where the service-plane fault occurred to the server.

**Input Parameters<a name="section1177311115553"></a>**

**Table 1**  Parameter description

|Parameter|Type|Description|
|--|--|--|
|fault_ranks|dict|Rank of the faulty process|

**Return Value<a name="section4468173015517"></a>**

**Table 2** Return value description

|Return Value Type|Description|
|--|--|
|int|`0` indicates success, and other return values indicate failure. For details about return codes, see [Return Codes](./07_return_codes.md).|

## taskd.python.framework.agent.ms_mgr.msrun_plugin. MSRunPlugin<a name="ZH-CN_TOPIC_0000002511426749"></a>

The `MSRunPlugin` class provides MindSpore process management functions. It is called by MindSpore and integrated into the MindSpore package.

## register_callbacks(self, operator, func)<a name="ZH-CN_TOPIC_0000002511346731"></a>

**Function Description<a name="section19441242061"></a>**

Registers a process management function with TaskD for subsequent use in managing the process lifecycle.

**Input Parameters<a name="section42271142719"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|
|--|--|--|
|operator|string|The type of callback currently being injected.<ul><li>`KILL_WORKER`: Registers a stop method for the MindSpore process to stop a specific training process.</li><li>START_ALL_WORKER: Registers a start method for the MindSpore process to start all processes on the current node.</li><li>`MONITOR`: Registers a monitoring method for the MindSpore process to return information about each rank process on the current node.</li><li>`START_WORKER_LIST`: Registers a start method for the MindSpore process to start some processes on the current node.</li></ul>|
|func|Function|The function callback for the currently registered function|

## start(self)<a name="ZH-CN_TOPIC_0000002479226816"></a>

Calls the `MSRunPlugin` start method to allow TaskD to take over MindSpore training process management.

## __init__(self)<a name="ZH-CN_TOPIC_0000002511346791"></a>

Constructs the `MSRunPlugin` class for subsequent instantiation and invocation by the user.
