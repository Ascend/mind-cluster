# Container Manager<a name="ZH-CN_TOPIC_0000002525600001"></a>

Container Manager组件提供容器生命周期管理、故障检测与恢复功能。在无K8s场景下的多机分布式任务（背靠背一体机）场景中，Container Manager通过Leader/普通节点协同方式，对跨节点的分布式任务容器进行统一启停协调，保证整个分布式任务的一致性恢复。分布式协调的详细原理请参见[分布式任务恢复](../04_usage/05_appliance/01_npu_hardware_fault_detection_and_rectification.md#分布式任务恢复)。

## 任务信息<a name="section_cm_task_info"></a>

Container Manager启动时解析容器label，根据label信息将容器划分为普通容器和分布式任务容器。分布式协调仅对配置了`huawei.com/job.id`的分布式任务容器生效，支持的任务label如下。

**表 1**  任务label

| label | 说明 |
|-------|------|
| huawei.com/job.id | 分布式任务标识，不配置时该容器视为普通容器。 |
| huawei.com/job.replica | 分布式任务副本数，建议配置为该分布式任务的实际副本数（大于1），以提升协调成功率。配置为1时，该容器按普通容器对待，不参与分布式协调。 |
| huawei.com/job.enableRecover | 任务恢复开关，仅配置了`huawei.com/job.id`时生效，未显式配置时默认开启，取值为false时关闭。 |

>[!NOTE]
>
>- 配置了`huawei.com/job.id`且副本数不为1的容器视为分布式任务容器，其启停由Leader节点统一协调。
>- 当`huawei.com/job.replica`配置为1时，该容器按普通容器对待，不参与分布式协调，不向Leader节点发起协调请求。
>- 为提升分布式任务容器协调成功率，请将`huawei.com/job.replica`配置为该分布式任务的实际副本数，便于Leader节点校验各节点上报的容器数量与任务副本数一致。
>- 容器（包括普通容器和分布式任务容器）是否参与恢复由`huawei.com/job.enableRecover`和启动参数`-ctrStrategy`共同决定，仅当两者均开启时，容器才参与跨节点协调自愈。

## 协调服务gRPC接口<a name="section_cm_grpc_desc"></a>

启用分布式协调时，普通节点与Leader节点之间通过gRPC接口通信。Leader节点gRPC服务默认监听端口为8890，可通过启动参数`-leaderPort`修改，详细请参见[Container Manager启动参数](../05_developer_guide/00_installation_deployment/00_manual_installation/11_container-manager.md#参数说明)。gRPC服务接口说明请参见[表2](#zh-cn_topic_0000002525600001_table2)。

**表 2**  gRPC服务接口

<a name="zh-cn_topic_0000002525600001_table2"></a>

|接口|调用方向|类型|说明|
|--|--|--|--|
|[SyncData](#section_cm_sync_data)|普通节点 -> Leader节点|一元调用|普通节点向Leader节点全量上报本节点的分布式任务容器信息，用于Leader节点维护全局容器状态。|
|[Coordinate](#section_cm_coordinate)|故障节点 -> Leader节点|一元调用|分布式任务容器需要启停时，节点向Leader节点发起协调请求；Leader节点校验通过后通过广播流将请求下发到关联节点，并同步等待所有节点执行完成。|
|[InitBroadcastStream](#section_cm_broadcast_stream)|普通节点 -> Leader节点|双向流|普通节点向Leader节点建立的常驻双向广播流。Leader节点通过该流主动向普通节点下发启停协调请求，普通节点通过该流返回执行结果。|

>[!NOTE]
>
>- 上述接口仅在启用分布式协调（配置了启动参数`-leaderAddrs`）时生效。
>- 普通节点对每个非本机Leader节点维护常驻连接和广播流，连接断开后自动重连。
>- 当节点为Leader节点（启动参数`-leaderIp`不为空）且`-leaderAddrs`中配置了本节点的监听地址（`-leaderIp:-leaderPort`）时，该节点不会通过gRPC连接自身，协调请求使用本地调用方式处理。

### SyncData（数据同步）<a name="section_cm_sync_data"></a>

**功能说明<a name="section_cm_sync_data_func"></a>**

普通节点向Leader节点全量上报本节点的分布式任务容器信息，Leader节点据此维护全局容器状态快照，供后续协调校验使用。

**函数原型<a name="section_cm_sync_data_proto"></a>**

```proto
rpc SyncData(SyncDataReq) returns (Response) {}
```

**输入参数说明<a name="section_cm_sync_data_req"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SyncDataReq|<p>message ContainerInfo{<p>string container_id = 1;</p><p>string status = 2;</p><p>repeated int32 phy_ids = 3;</p><p>string job_id = 4;</p><p>int32 job_replica = 5;</p><p>bool enable_recover = 6;</p><p>string paused_by_peer = 7;</p>}</p><p>message SyncDataReq{<p>string uuid = 1;</p><p>string node_id = 2;</p><p>repeated ContainerInfo containers = 3;</p><p>int64 sync_time = 4;</p>}</p>|<p>**ContainerInfo.container_id**：string类型，容器ID。</p><p>**ContainerInfo.status**：string类型，容器状态。</p><p>**ContainerInfo.phy_ids**：int32数组，容器使用的NPU物理ID列表。</p><p>**ContainerInfo.job_id**：string类型，容器所属的分布式任务标识。</p><p>**ContainerInfo.job_replica**：int32类型，分布式任务副本总数。</p><p>**ContainerInfo.enable_recover**：bool类型，是否开启恢复，true表示开启，启动参数`-ctrStrategy`为never时恒为false。</p><p>**ContainerInfo.paused_by_peer**：string类型，暂停该容器的发起节点，为空表示未被远端节点暂停。</p><p>**SyncDataReq.uuid**：string类型，请求唯一标识。</p><p>**SyncDataReq.node_id**：string类型，上报节点的节点标识（hostname或启动参数`-nodeID`配置值）。</p><p>**SyncDataReq.containers**：repeated ContainerInfo类型，本节点所有分布式任务容器信息，最多128个。</p><p>**SyncDataReq.sync_time**：int64类型，同步时间戳。</p>|

**返回值说明<a name="section_cm_sync_data_resp"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Response|message Response{<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>string message = 3;</p>}|<p>**uuid**：string类型，对应请求的唯一标识。</p><p>**code**：uint32类型，返回码，取值为0表示同步成功，其他值为错误码。</p><p>**message**：string类型，错误描述信息。</p>|

### Coordinate（协调请求）<a name="section_cm_coordinate"></a>

**功能说明<a name="section_cm_coordinate_func"></a>**

分布式任务容器需要启停时，节点向Leader节点发起协调请求。Leader节点校验通过后，通过广播流将请求下发到该任务关联的所有节点，并同步等待所有节点执行完成后返回结果。若全部Leader节点均协调失败，则本次协调失败。

**函数原型<a name="section_cm_coordinate_proto"></a>**

```proto
rpc Coordinate(CoordinateReq) returns (Response) {}
```

**输入参数说明<a name="section_cm_coordinate_req"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|CoordinateReq|message CoordinateReq{<p>string uuid = 1;</p><p>string node_id = 2;</p><p>repeated string job_ids = 3;</p><p>repeated string ctr_ids = 4;</p><p>string action = 5;</p>}|<p>**uuid**：string类型，请求唯一标识，用于请求去重。</p><p>**node_id**：string类型，发起协调请求的节点标识。</p><p>**job_ids**：repeated string，需要启停的分布式任务ID列表（一个芯片或环上可能挂载多个任务）。</p><p>**ctr_ids**：repeated string，发起请求的容器ID列表。</p><p>**action**：string类型，动作类型，取值为stop（停止）或start（启动）。</p>|

**返回值说明<a name="section_cm_coordinate_resp"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Response|message Response{<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>string message = 3;</p>}|<p>**uuid**：string类型，对应请求的唯一标识。</p><p>**code**：uint32类型，返回码，取值为0表示协调成功，其他值为错误码。</p><p>**message**：string类型，错误描述信息。</p>|

### InitBroadcastStream（广播流）<a name="section_cm_broadcast_stream"></a>

**功能说明<a name="section_cm_broadcast_stream_func"></a>**

普通节点向Leader节点建立常驻双向广播流。Leader节点通过该流主动向普通节点下发启停协调请求，普通节点执行后通过该流返回执行结果。连接断开后，普通节点会自动重连。

**函数原型<a name="section_cm_broadcast_stream_proto"></a>**

```proto
rpc InitBroadcastStream(stream Response) returns (stream CoordinateReq) {}
```

**输入参数说明<a name="section_cm_broadcast_stream_req"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|stream Response|gRPC stream，消息类型为Response|<p>普通节点通过该流返回执行结果。**Response**的具体字段说明请参见<a href="#section_cm_coordinate_resp">Coordinate返回值说明</a>。</p>|

**返回值说明<a name="section_cm_broadcast_stream_resp"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream CoordinateReq|gRPC stream，消息类型为CoordinateReq|<ul><li>Leader节点通过该流主动向普通节点下发启停协调请求。</li><li>**CoordinateReq**的具体字段说明请参见<a href="#section_cm_coordinate_req">Coordinate输入参数说明</a>。</li></ul>|
