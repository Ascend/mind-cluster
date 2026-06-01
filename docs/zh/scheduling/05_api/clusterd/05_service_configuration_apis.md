# 业务配置接口<a name="ZH-CN_TOPIC_0000002479226840"></a>

## Register<a name="ZH-CN_TOPIC_0000002511426719"></a>

**功能说明<a name="section143314311911"></a>**

接收处理客户端的注册请求，为订阅相关业务配置做初始化准备。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|<p>**Status.code**：返回码。<ul><li>取值为0：表示注册成功。</li><li>其他值：表示注册失败。</li></ul></p><p>**Status.info**：返回信息描述。</p>|

## SubscribeRankTable<a name="ZH-CN_TOPIC_0000002511346779"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅RankTable请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc SubscribeRankTable(ClientInfo) returns (stream RankTableStream) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|

**发送数据说明<a name="section8539121202217"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|RankTableStream|<p>message RankTableStream{</p><p>string jobId = 1;</p><p>string rankTable = 2;</p>}|<p>**RankTableStream.jobId**：任务ID。</p><p>**RankTableStream.rankTable**：RankTable信息，各字段的详细说明如[表1](#table5843145110294)所示。</p>|

**global-ranktable文件说明<a name="section268935611912"></a>**

ClusterD会生成global-ranktable在RankTable字段作为返回消息。global-ranktable中部分字段来自于hccl.json文件，关于hccl.json文件的详细说明请参见[hccl.json文件说明](../hccl.json_file_description.md)。

- 示例如下。

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

- <term>Atlas A3 训练系列产品</term>示例如下。

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

**表 1**  global-ranktable字段说明

<a name="table5843145110294"></a>

|字段|说明|
|--|--|
|version|版本|
|status|状态|
|server_group_list|服务组列表|
|group_id|任务组编号|
|server_count|服务器数量|
|server_list|服务器列表|
|server_id|AI Server标识，全局唯一|
|server_ip|Pod IP|
|device_id|NPU的设备ID|
|device_ip|NPU的设备IP|
|super_device_id|<term>Atlas A3 训练系列产品</term>超节点内NPU的唯一标识|
|rank_id|NPU对应的训练rank ID|
|device_logical_id|NPU的逻辑ID|
|super_pod_list|超节点列表|
|super_pod_id|逻辑超节点ID|
