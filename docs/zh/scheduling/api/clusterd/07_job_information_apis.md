# 任务信息接口<a name="ZH-CN_TOPIC_0000002511426731"></a>

## 接口说明<a name="ZH-CN_TOPIC_0000002479386816"></a>

本模块向外部提供订阅获取集群任务状态及基本信息的订阅接口。分为3个接口即：Register、SubscribeJobSummarySignal和SubscribeJobSummarySignalList。

**调用顺序说明<a name="section171351329174616"></a>**

用户在调用SubscribeJobSummarySignal和SubscribeJobSummarySignalList接口前需首先通过Register接口获取合法的客户端ID后，持有该ID调用订阅接口。

订阅接口默认两分钟内无活动信息主动关闭。

## Register<a name="ZH-CN_TOPIC_0000002479226804"></a>

**功能说明<a name="section14645125754213"></a>**

在订阅任务信息的场景下，接收客户端的注册请求。

客户端如需订阅集群任务信息，需先调用本接口获取返回的UUID后，持有该ID调用订阅接口SubscribeJobSummarySignal和SubscribeJobSummarySignalList获取集群的任务信息。

>[!NOTE] 
>集群内最多存在80个活跃订阅链接，每种客户端角色最多支持创建20个活跃订阅链接。

**函数原型<a name="section4140960433"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section1317321424310"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string role = 1;</p><p>string clientId = 3;</p>}|**ClientInfo.role**：客户端角色。当前仅支持以下几种客户端角色。如果传入其他值，会导致注册失败。<ul><li>CCAgent</li><li>DefaultUser1</li><li>DefaultUser2</li><li>FdAgent</li></ul>**ClientInfo.clientId**：客户端ID。|

**返回值说明<a name="section4839929184717"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p><p>string clientId = 3;</p>}|**Status.code**：本次调用结果的状态码。目前分为以下几种：<ul><li>200：查询正常返回。</li><li>429：服务端限流。</li><li>500：服务端错误。</li></ul><p>**Status.info**：本次调用结果的描述信息。</p><p>**Status.clientId**：注册接口返回的UUID。</p>|

## SubscribeJobSummarySignal<a name="ZH-CN_TOPIC_0000002511426723"></a>

**功能说明<a name="section85381247165120"></a>**

接收客户端的任务信息变更订阅。客户端初次订阅接口时，会逐条推送当前集群中所有任务的信息。当任务状态改变时，向注册的客户端广播推送。当连接两分钟内无消息且无心跳时，服务端主动断开该连接，并释放订阅。

>[!NOTE] 
>
>- 本接口具有限流机制，1秒内允许访问的最大次数为20。
>- 集群内最多存在80个活跃订阅链接，每种客户端角色最多支持创建20个活跃订阅链接。

**函数原型<a name="section1199205575113"></a>**

```proto
rpc SubscribeJobSummarySignal(ClientInfo) returns (stream JobSummarySignal){}
```

**输入参数说明<a name="section6291133165212"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string role = 1;</p><p>string clientId = 3;</p>}|**ClientInfo.role**：客户端角色。当前仅支持以下几种客户端角色。如果传入其他值，会导致注册失败。<ul><li>CCAgent</li><li>DefaultUser1</li><li>DefaultUser2</li><li>FdAgent</li></ul>**ClientInfo.clientId**：客户端ID。|

**返回值说明<a name="section1883821810542"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|

**发送数据说明<a name="section10140143475520"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|JobSummarySignal|<p>message JobSummarySignal{</p><p>string uuid = 1;</p><p>string jobId = 2;</p><p>string jobName = 3;</p><p>string namespace =4;</p><p>string frameWork = 5;</p><p>string jobStatus = 6;</p><p>string time = 7;</p><p>string cmIndex = 8;</p><p>string total = 9;</p><p>string HcclJson = 10;</p><p>string deleteTime = 11;</p><p>string sharedTorIp = 12;</p><p>string masterAddr = 13;</p><p>string operator = 14;</p><p>string sid = 15;</p>}|<p>**uuid**：本条消息id</p><p>**jobId**：任务的K8s ID信息</p><p>**jobName**：当前任务的名称</p><p>**namespace**：任务所属命名空间</p><p>**frameWork**：任务框架</p>**jobStatus**：任务状态，存在以下几种状态：<ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul><p>**time**：任务开始时间</p><p>**cmIndex**：序号</p><p>**total**：任务对应的jobsummary ConfigMap的数量总数</p><p>**HcclJson**：任务使用的芯片通信信息。若任务调度的NPU数量超过4万，客户端接收的上报信息中HcclJson会被设置为空。<p>可转义为JSON格式，字段说明如下：</p><ul><li>status：任务RankTable是否已经生成</li><li>initializing：还在为任务分配设备，RankTable未生成</li><li>complete：当RankTable生成后，状态会立即变为complete，同步出现server_list等其他字段</li><li>server_list：任务设备分配情况</li><li>device：记录NPU分配，NPU IP和rank_id信息</li><li>server_id：AI Server标识，全局唯一</li><li>server_name：节点名称</li><li>server_sn：节点的SN号。需要保证设备的SN存在。若不存在，请联系华为技术支持</li><li>server_count：任务使用的节点数量</li><li>version：版本信息</li></ul></p><p>**deleteTime**：任务被删除的时间</p><p>**sharedTorIp**：任务使用的共享交换机信息</p><p>**masterAddr**：PyTorch训练时指定的MASTER_ADDR值</p><p>**operator**：接收到添加任务命令后状态更新为add；接收到删除任务命令后状态更新为delete</p><p>**sid**：作业唯一标识符</p>|

## SubscribeJobSummarySignalList

**功能说明**

接收客户端的任务信息变更订阅。客户端初次订阅接口时，推送当前集群中所有任务的信息。当任务状态改变时，向注册的客户端广播推送。当连接两分钟内无消息且无心跳时，服务端主动断开该连接，并释放订阅。

>[!NOTE]
> 
>- 本接口具有限流机制，1秒内允许访问的最大次数为20。
>- 集群内最多存在80个活跃订阅链接，每种客户端角色最多支持创建20个活跃订阅链接。

**函数原型**

```proto
rpc SubscribeJobSummarySignalList(ClientInfo) returns (stream JobSummarySignalList){}
```

**输入参数说明**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string role = 1;</p><p>string clientId = 3;</p>}|**ClientInfo.role**：客户端角色。当前仅支持以下几种客户端角色。如果传入其他值，会导致注册失败。<ul><li>CCAgent</li><li>DefaultUser1</li><li>DefaultUser2</li><li>FdAgent</li></ul><p>**ClientInfo.clientId**：客户端ID。</p>|

**返回值说明**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|

**发送数据说明**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|JobSummarySignalList|<p>message JobSummarySignalList{</p><p>repeated JobSummarySignal jobSummarySignals = 1;</p><p>string ReportTime = 2;</p><p>int32 JobTotalNum = 3;</p>}<p>message JobSummarySignal{<p>string uuid = 1;</p><p>string jobId = 2;</p><p>string jobName = 3;</p><p>string namespace =4;</p><p>string frameWork = 5;</p><p>string jobStatus = 6;</p><p>string time = 7;</p><p>string cmIndex = 8;</p><p>string total = 9;</p><p>string HcclJson = 10;</p><p>string deleteTime = 11;</p><p>string sharedTorIp = 12;</p><p>string masterAddr = 13;</p><p>string operator = 14;</p><p>string sid = 15;</p>}</p>|<p>**jobSummarySignals**: 任务信息列表</p><p>**ReportTime**：当前批次上报的时间</p><p>**JobTotalNum**：相同批次上报的任务总数</p><p>**uuid**：本条消息id</p><p>**jobId**：任务的K8s ID信息</p><p>**jobName**：当前任务的名称</p><p>**namespace**：任务所属命名空间</p><p>**frameWork**：任务框架</p>**jobStatus**：任务状态，存在以下几种状态：<ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul><p>**time**：任务开始时间</p><p>**cmIndex**：序号</p><p>**total**：任务对应的jobsummary ConfigMap的数量总数</p><p>**HcclJson**：任务使用的芯片通信信息。可转义为JSON格式，字段说明如下：<ul><li>status：任务RankTable是否已经生成</li><li>initializing：还在为任务分配设备，RankTable未生成</li><li>complete：当RankTable生成后，状态会立即变为complete，同步出现server_list等其他字段</li><li>server_list：任务设备分配情况</li><li>device：记录NPU分配，NPU IP和rank_id信息</li><li>server_id：AI Server标识，全局唯一</li><li>server_name：节点名称</li><li>server_sn：节点的SN号。需要保证设备的SN存在。若不存在，请联系华为技术支持</li><li>server_count：任务使用的节点数量</li><li>version：版本信息</li></ul></p><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>若单个任务所使用的NPU数量超过4万，上报的任务信息中HcclJson会被设置为空。</li><li>客户端初次订阅接口时，若多个任务合计使用的NPU数量超过4万，会对上报信息进行分页上报，确保每条上报信息中任务总NPU数不超过4万。</li></ul></div></div><p>**deleteTime**：任务被删除的时间</p><p>**sharedTorIp**：任务使用的共享交换机信息</p><p>**masterAddr**：PyTorch训练时指定的MASTER_ADDR值</p><p>**operator**：接收到添加任务命令后状态更新为add；接收到删除任务命令后状态更新为delete</p><p>**sid**：作业唯一标识符</p>|
