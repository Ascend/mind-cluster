# Container Manager<a name="ZH-CN_TOPIC_0000002525600001"></a>

The Container Manager component provides container lifecycle management, fault detection, and fault recovery. In multi-node distributed job (back-to-back appliance) scenarios without K8s, Container Manager uses a Leader/normal node collaboration mode to coordinate the start and stop of distributed job containers across nodes, ensuring consistent recovery of the entire distributed job. For details about the distributed coordination principle, see [Distributed Job Container Recovery](../04_usage/06_appliance/01_npu_hardware_fault_detection_and_rectification.md#distributed-job-container-recovery).

## Job Information<a name="section_cm_task_info"></a>

When Container Manager is started, it parses container labels and classifies containers into normal containers and distributed job containers based on the label information. Distributed coordination takes effect only for distributed job containers configured with `huawei.com/job.id`. The supported job labels are as follows.

**Table 1** Job labels

| Label | Description |
|-------|------|
| huawei.com/job.id | Distributed job identifier. If this label is not configured, the container is treated as a normal container. |
| huawei.com/job.replica | Number of distributed job replicas. You are advised to set this parameter to the actual number of replicas (greater than 1) of the distributed job to improve the coordination success rate. If this parameter is set to 1, the container is treated as a normal container and does not participate in distributed coordination. |
| huawei.com/job.enableRecover | Job recovery switch. It takes effect only when huawei.com/job.id is configured. If it is not explicitly configured, the switch is enabled by default. If the value is false, the switch is disabled. |

>[!NOTE]
>
>- A container configured with `huawei.com/job.id` and with the number of replicas not 1 is treated as a distributed job container, and its start and stop are coordinated by the Leader node.
>- When `huawei.com/job.replica` is set to 1, the container is treated as a normal container which does not participate in distributed coordination and does not initiate coordination requests to the leader node.
>- To improve the coordination success rate of distributed job containers, set `huawei.com/job.replica` to the actual number of replicas of the distributed job, so that the leader node can verify that the number of containers reported by each node is consistent with the number of job replicas.
>- Whether a container (including normal containers and distributed job containers) participates in recovery is determined by `huawei.com/job.enableRecover` and the startup parameter `-ctrStrategy` together. Only when both are enabled does the container participate in cross-node coordination self-healing.

## gRPC Interfaces of the Coordination Service<a name="section_cm_grpc_desc"></a>

When distributed coordination is enabled, a normal node communicates with the leader node through gRPC interfaces. The gRPC service on the leader node listens on port 8890 by default, which can be changed through the `-leaderPort` startup parameter. For details, see [Container Manager Startup Parameters](../05_developer_guide/00_installation_deployment/00_manual_installation/10_container-manager.md#parameter-description). For details about the gRPC service interfaces, see [Table 2](#zh-cn_topic_0000002525600001_table2).

**Table 2** gRPC service interfaces

<a name="zh-cn_topic_0000002525600001_table2"></a>

|Interface|Invocation Direction|Type|Description|
|--|--|--|--|
|[SyncData](#section_cm_sync_data)|Normal node -> Leader node|Unary call|A normal node reports the complete information about the distributed job containers on this node to the leader node, so that the leader node can maintain the global container status.|
|[Coordinate](#section_cm_coordinate)|Faulty node -> leader node|Unary call|When a distributed job container needs to be started or stopped, the node initiates a coordination request to the leader node. After the verification is successful, the leader node delivers the request to the associated nodes through the broadcast stream and waits synchronously until all nodes have completed the execution.|
|[InitBroadcastStream](#section_cm_broadcast_stream)|Normal node -> Leader node|Bidirectional stream|A persistent bidirectional broadcast stream established by a normal node to the leader node. The leader node proactively delivers start/stop coordination requests to the normal node through this stream, and the normal node returns the execution result through this stream.|

>[!NOTE]
>
>- The preceding interfaces take effect only when distributed coordination is enabled (the startup parameter `-leaderAddrs` is configured).
>- A normal node maintains a persistent connection and broadcast stream for each non-local leader node, and automatically reconnects after the connection is disconnected.
>- When a node is a leader node (the startup parameter `-leaderIp` is not empty) and the listening address of this node (`-leaderIp:-leaderPort`) is configured in `-leaderAddrs`, the node does not connect to itself through gRPC, and coordination requests are handled as local calls.

### SyncData (Data Synchronization)<a name="section_cm_sync_data"></a>

**Function Description<a name="section_cm_sync_data_func"></a>**

A normal node reports the complete information about the distributed job containers on this node to the leader node. Based on the information, the leader node maintains a global container status snapshot for subsequent coordination verification.

**Function Prototype<a name="section_cm_sync_data_proto"></a>**

```proto
rpc SyncData(SyncDataReq) returns (Response) {}
```

**Input Parameters<a name="section_cm_sync_data_req"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|SyncDataReq|<p>message ContainerInfo{<p>string container_id = 1;</p><p>string status = 2;</p><p>repeated int32 phy_ids = 3;</p><p>string job_id = 4;</p><p>int32 job_replica = 5;</p><p>bool enable_recover = 6;</p><p>string paused_by_peer = 7;</p>}</p><p>message SyncDataReq{<p>string uuid = 1;</p><p>string node_id = 2;</p><p>repeated ContainerInfo containers = 3;</p><p>int64 sync_time = 4;</p>}</p>|<p>**ContainerInfo.container_id**: String type. Container ID.</p><p>**ContainerInfo.status**: String type. Container status.</p><p>**ContainerInfo.phy_ids**: int32 array. List of NPU physical IDs used by the container.</p><p>**ContainerInfo.job_id**: String type. Identifier of the distributed job to which the container belongs.</p><p>**ContainerInfo.job_replica**: int32 type. Total number of distributed job replicas.</p><p>**ContainerInfo.enable_recover**: bool type. Whether to enable recovery. true indicates enabled. When the startup parameter `-ctrStrategy` is never, the value is always false.</p><p>**ContainerInfo.paused_by_peer**: String type. Node that initiated the pause of the container. If the value is empty, the container is not paused by a remote node.</p><p>**SyncDataReq.uuid**: String type. Unique request identifier.</p><p>**SyncDataReq.node_id**: String type. Node identifier of the reporting node (hostname or the value configured by the startup parameter `-nodeID`).</p><p>**SyncDataReq.containers**: repeated ContainerInfo type. Information about all distributed job containers on this node, up to 128.</p><p>**SyncDataReq.sync_time**: int64 type. Synchronization timestamp.</p>|

**Return Values<a name="section_cm_sync_data_resp"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Response|message Response{<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>string message = 3;</p>}|<p>**uuid**: String type. Unique identifier of the corresponding request.</p><p>**code**: uint32 type. Return code. The value 0 indicates that the synchronization is successful, and other values indicate error codes.</p><p>**message**: String type. Error description information.</p>|

### Coordinate (Coordination Request)<a name="section_cm_coordinate"></a>

**Function Description<a name="section_cm_coordinate_func"></a>**

When a distributed job container needs to be started or stopped, the node initiates a coordination request to the leader node. After the verification is successful, the leader node delivers the request to all nodes associated with the job through the broadcast stream and waits synchronously until all nodes have completed the execution before returning the result. If all leader nodes fail to coordinate, this coordination fails.

**Function Prototype<a name="section_cm_coordinate_proto"></a>**

```proto
rpc Coordinate(CoordinateReq) returns (Response) {}
```

**Input Parameters<a name="section_cm_coordinate_req"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|CoordinateReq|message CoordinateReq{<p>string uuid = 1;</p><p>string node_id = 2;</p><p>repeated string job_ids = 3;</p><p>repeated string ctr_ids = 4;</p><p>string action = 5;</p>}|<p>**uuid**: String type. Unique request identifier, used for request deduplication.</p><p>**node_id**: String type. Identifier of the node that initiates the coordination request.</p><p>**job_ids**: repeated string. List of IDs of the distributed jobs to be started or stopped (multiple jobs may be mounted on one chip or ring).</p><p>**ctr_ids**: repeated string. List of IDs of the containers that initiate the request.</p><p>**action**: String type. Action type. The value is stop or start.</p>|

**Return Values<a name="section_cm_coordinate_resp"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Response|message Response{<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>string message = 3;</p>}|<p>**uuid**: String type. Unique identifier of the corresponding request.</p><p>**code**: uint32 type. Return code. The value 0 indicates that the coordination is successful, and other values indicate error codes.</p><p>**message**: String type. Error description information.</p>|

### InitBroadcastStream (Broadcast Stream)<a name="section_cm_broadcast_stream"></a>

**Function Description<a name="section_cm_broadcast_stream_func"></a>**

A normal node establishes a persistent bidirectional broadcast stream to the leader node. The leader node proactively delivers start/stop coordination requests to the normal node through this stream, and the normal node returns the execution result through this stream after execution. After the connection is disconnected, the normal node automatically reconnects.

**Function Prototype<a name="section_cm_broadcast_stream_proto"></a>**

```proto
rpc InitBroadcastStream(stream Response) returns (stream CoordinateReq) {}
```

**Input Parameters<a name="section_cm_broadcast_stream_req"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|stream Response|gRPC stream. The message type is Response|<p>The normal node returns the execution result through this stream. For details about the fields of **Response**, see <a href="#section_cm_coordinate_resp">Coordinate Return Values</a>.</p>|

**Return Values<a name="section_cm_broadcast_stream_resp"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream CoordinateReq|gRPC stream. The message type is CoordinateReq|<ul><li>The leader node proactively delivers start/stop coordination requests to the normal node through this stream.</li><li>For details about the fields of **CoordinateReq**, see <a href="#section_cm_coordinate_req">Coordinate Input Parameters</a>.</li></ul>|
