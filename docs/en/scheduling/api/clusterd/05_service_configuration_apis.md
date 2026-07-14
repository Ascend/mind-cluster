# Service Configuration<a name="ZH-CN_TOPIC_0000002479226840"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:46:45.467Z pushedAt=2026-06-09T02:05:50.692Z -->

## Register<a name="ZH-CN_TOPIC_0000002511426719"></a>

**Description<a name="section143314311911"></a>**

Receives and processes client registration requests, performing initialization preparations for subscribing to related service configurations.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|<p>**Status.code**: Return code.<ul><li>`0`: successful registration.</li><li>Other values: registration failure.</li></ul></p><p>**Status.info**: Return information.</p>|

## SubscribeRankTable<a name="ZH-CN_TOPIC_0000002511346779"></a>

**Description<a name="section143314311911"></a>**

Receives a client's request to subscribe to the RankTable. The server assigns a message queue to each job and listens for pending messages in the queue. If messages are present, they are sent to the client via a gRPC stream.

**Prototype<a name="section3958124212115"></a>**

```proto
rpc SubscribeRankTable(ClientInfo) returns (stream RankTableStream) {}
```

**Input Parameters<a name="section14344145451114"></a>**

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**: Job ID.</p><p>**ClientInfo.role**: Client role.</p>|

**Return Value<a name="section206103328174"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream|grpc stream|<ul><li>This API returns a gRPC stream (the specific data structure of the return value depends on the programming language selected by the client).</li><li>The client can call the stream's `Receive` method (the specific method name depends on the programming language selected by the client) to receive data pushed by the server.</li></ul>|

**Sending Data Description<a name="section8539121202217"></a>**

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|RankTableStream|<p>message RankTableStream{</p><p>string jobId = 1;</p><p>string rankTable = 2;</p>}|<p>**RankTableStream.jobId**: Job ID.</p><p>**RankTableStream.rankTable**: RankTable information. For detailed descriptions of each field, see [Table 1](#table5843145110294).</p>|

**global-ranktable Description<a name="section268935611912"></a>**

ClusterD generates `global-ranktable` in the RankTable field as the return message. Some fields in `global-ranktable` come from the `hccl.json` file. For details about the `hccl.json` file, see [hccl.json File Description](../hccl.json_file_description.md).

- Example:

    ```json
    {
        "version": "1.0",
        "status": "completed",
        "server_group_list": [
            {
                "group_id": "2",
                "deploy_server": "0",
                "server_count": "1",
                "server_list": [
                    {
                        "device": [
                            {
                                "device_id": "x",
                                "device_ip": "xx.xx.xx.xx",
                                "device_logical_id": "x",
                                "rank_id": "x"
                            }
                        ],
                        "server_id": "xx.xx.xx.xx",
                        "server_ip": "xx.xx.xx.xx"
                    }
                ]
            }
        ]
    }
    ```

- Example (Atlas A3 Training Series Products):

    ```json
    {
        "version": "1.2",
        "status": "completed",
        "server_group_list": [
            {
                "group_id": "2",
                "deploy_server": "1",
                "server_count": "1",
                "server_list": [
                    {
                        "device": [
                            {
                                "device_id": "0",
                                "device_ip": "xx.xx.xx.xx",
                                "super_device_id": "xxxxx",
                                "device_logical_id": "0",
                                "rank_id": "0"
                            }
                        ],
                        "server_id": "xx.xx.xx.xx",
                        "server_ip": "xx.xx.xx.xx"
                    }
                ],
                "super_pod_list": [
                    {
                        "super_pod_id": "0",
                        "server_list": [
                            {
                                "server_id": "xx.xx.xx.xx"
                            }
                        ]
                    }
                ]
            }
        ]
    }
    ```

**Table 1** global-ranktable description

<a name="table5843145110294"></a>

|Field|Description|
|--|--|
|version|Version|
|status|Status|
|server_group_list|Server group list|
|group_id|Job group ID|
|server_count|Number of servers|
|server_list|Server list|
|server_id|AI Server identifier, globally unique|
|server_ip|Pod IP|
|device_id|NPU device ID|
|device_ip|NPU device IP|
|super_device_id|Unique identifier of the NPU within the Atlas A3 Training Series Products/SuperPoD|
|rank_id|Training rank ID corresponding to the NPU|
|device_logical_id|Logical ID of the NPU|
|super_pod_list|SuperPoD list|
|super_pod_id|Logical SuperPoD ID|
