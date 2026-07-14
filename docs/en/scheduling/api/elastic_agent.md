# Elastic-Agent (Interfaces Related to Resumable Training) <a name="ZH-CN_TOPIC_0000002479386784"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:41:31.793Z pushedAt=2026-06-09T02:05:50.620Z -->

>[!NOTE]
>
>- Elastic Agent has reached the EOL, and related materials will be removed in the version of December 30, 2026.
>- Internal interfaces. Do not call.

## mindx_elastic.__version__<a name="ZH-CN_TOPIC_0000002511346763"></a>

Obtains the Elastic Agent version number.

Input: Empty

Return Value: Elastic Agent version number

Usage example:

```Python
import mindx_elastic
mindx_elastic.__version__
```

## mindx_elastic.api.patch_torch_methods (Internal)<a name="ZH-CN_TOPIC_0000002479226826"></a>

**Function<a name="section1222112260226"></a>**

When building a training image and installing Elastic Agent, after using the command `sed -i '/import os/i import mindx_elastic.api' $(pip3.7 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py`, Elastic Agent executes the `mindx_elastic.api.patch_torch_methods` interface when importing Torch's Elastic module, so that the patch takes effect automatically.

Elastic Agent patches methods such as `torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run`, `torch.distributed.launcher.api.launch_agent`, and `torch.distributed.elastic.agent.server.api.SimpleElasticAgent._initialize_workers` to additionally provide fault detection and recovery capabilities for Ascend NPU devices.

## mindx_elastic.recover_manager.DLRecoverManager (Internal)<a name="ZH-CN_TOPIC_0000002511346787"></a>

The `DLRecoverManager` class provides interfaces for process-level recovery and process-level online recovery. The client imports it into the client code as a Python package.

>[!NOTE]
>The interfaces provided by the `DLRecoverManager` class may throw Exception. The caller is responsible for catching and handling the exceptions.

**__init__(self, info: pb.ClientInfo, server_addr: str, secure_conn: bool = True, cert_path: str = "")<a name="section93535281517"></a>**

Constructs a `DLRecoverManager` for subsequent communication.

**Table 1** Parameters

|Parameter|Type|Description|
|--|--|--|
|info|pb.ClientInfo|<p>info.ip: str type, client IP (Reserved).</p><p>info.port: str type, client port (Reserved).</p><p>info.taskId: str type, task ID.</p><p>info.role: str type, client role.</p>|
|server_addr|str|Server address|

**register(self, request: pb.ClientInfo) -> pb.Status<a name="section92911329181515"></a>**

Registers the client. The server performs pre-recovery initialization for the job specified by `request`.

**Table 2** Parameters

|Parameter|Type|Description|
|--|--|--|
|request|pb.ClientInfo|<p>request.ip: str type, client IP (Reserved).</p><p>request.port: str type, client port (Reserved).</p><p>request.taskId: str type, task ID.</p><p>request.role: str type, client role.</p>|

**Table 3** Return values

|Return Value Type|Description|
|--|--|
|Status|<p>Status.info: str type, return information description</p><p>Status.code: int type, 0 indicates success, other values indicate failure. For details about return codes, see [Return Code Description](#return-code-description).</p>|

**start_subscribe(self)<a name="section5051271214"></a>**

The client and server establish a gRPC persistent connection, through which the server communicates unidirectionally with the client. For example, when a fault occurs, the server sends the client information such as stop training signal and global fault process rank.

**init_clusterd(self)<a name="section18270133519256"></a>**

The client initializes the ClusterD server status to ensure that subsequent jobs can be registered and connections established normally.

## report_stop_complete(code: int, msg: str, fault_ranks: dict) -> int<a name="ZH-CN_TOPIC_0000002511426697"></a>

The client reports to the server that the job process has stopped. Typically, after receiving a training stop signal from the server, the client stops the training job process and then reports the process stop completion to the server.

**Table 1** Parameters

|Parameter|Type|Description|
|--|--|--|
|code|int|Status code|
|msg|str|Return information|
|fault_ranks|dict|Fault process rank|

**Table 2** Return values

|Return Value Type|Description|
|--|--|
|int|0 indicates success, and other return values indicate failure. For details about return codes, see [Return Code Description](#return-code-description).|

## report_recover_strategy(fault_ranks: dict, strategy_list: list) -> int<a name="ZH-CN_TOPIC_0000002511346757"></a>

The client reports the recovery policies it supports to the server, so that the server can select the optimal recovery policy. The server then sends it to the client through the persistent connection established by `start_subscribe`.

**Table 1** Parameters

|Parameter|Type|Description|
|--|--|--|
|fault_ranks|dict|Fault process rank|
|strategy_list|list|List of recovery policies|

**Table 2** Return values

|Return Value Type|Description|
|--|--|
|int|0 indicates success, and other return values indicate failure. For details about return codes, see [Return Code Description](#return-code-description).|

## report_recover_status(code: int, msg: str, fault_ranks: dict, strategy: str) -> int<a name="ZH-CN_TOPIC_0000002511426757"></a>

The client reports the job recovery status to the server.

**Table 1**  Parameters

|Parameter|Type|Description|
|--|--|--|
|code|int|Status code|
|msg|str|Return information|
|fault_ranks|dict|Fault process rank|
|strategy|str|Recovery policy|

**Table 2**  Return values

|Return Value Type|Description|
|--|--|
|int|0 indicates success, and other return values indicate failure. For details about the return codes, see [Return Code Description](#return-code-description).|

## report_process_fault(fault_ranks: dict) -> int<a name="ZH-CN_TOPIC_0000002479386856"></a>

The client reports a service-plane fault of the job process. When the client detects a fault first, it reports the information about the rank where the service-plane fault is located to the server.

**Table 1**  Parameters

|Parameter|Type|Description|
|--|--|--|
|fault_ranks|dict|Fault process rank|

**Table 2** Return values

|Return Value Type|Description|
|--|--|
|int|0 indicates success, and other return values indicate failure. For details, see [Return Code Description](#return-code-description).|

## Return Code Description<a name="ZH-CN_TOPIC_0000002511426709"></a>

The Elastic-Agent return codes are shown in [Table 1](#table1248859202914).

**Table 1** Elastic-Agent return codes

<a name="table1248859202914"></a>

| Return Code | Value | Meaning |
|--|--|--|
| OK | 0 | The API call is normal. |
| UnRegistry | 400 | The Job ID is not registered. |
| OrderMix | 401 | The request does not conform to the state machine sequence. |
| JobNotExist | 402 | The Job ID does not exist. |
| ProcessRescheduleOff | 403 | The process-level recovery switch is not turned on. |
| ProcessNotReady | 404 | The training process is not started. |
| RecoverableRetryError | 405 | Recovery failed due to a clean device failure. |
| UnRecoverableRetryError | 406 | Recovery failed due to a stop device failure. |
| DumpError | 407 | Failed to save the last words. |
| UnInit | 408 | Initialization is not called. |
| ClientError | 499 | Other failure causes. |
| OutOfMaxServeJobs | 500 | The maximum number of service jobs has been exceeded. |
| OperateConfigMapError | 501 | Failed to operate the ConfigMap. |
| OperatePodGroupError | 502 | Failed to operate the PodGroup. |
| ScheduleTimeout | 503 | Pod scheduling timed out. |
| SignalQueueBusy | 504 | Failed to enqueue the control signal. |
| EventQueueBusy | 505 | Failed to enqueue the state machine event. |
| ControllerEventCancel | 506 | The state machine has exited. |
| WaitReportTimeout | 507 | Timed out waiting for the client to call the API. |
| WaitPlatStrategyTimeout | 508 | Timed out waiting for the AI platform to prepare the recovery policy. |
| WriteConfirmFaultOrWaitPlatResultFault | 509 | AI platform fault information error. |
| ServerInnerError | 599 | Internal server error. |
