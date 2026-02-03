# ClusterD<a name="ZH-CN_TOPIC_0000002511346793"></a>

## 集群资源<a name="ZH-CN_TOPIC_0000002511346785"></a>

**ConfigMap说明<a name="section17868183824213"></a>**

ClusterD启动后，会创建如下ConfigMap：

-   cluster-info-node-cm，详细说明请参见[表1](#table25031946405)。
-   cluster-info-device-$\{m\}，详细说明请参见[表2](#table915714719368)。m为从0开始递增的整数。集群规模每增加1000个节点，则会新增一个该ConfigMap文件。
-   cluster-info-switch-$\{x\}，详细说明请参见[表3](#table9246232250)。x为从0开始递增的整数。集群规模每增加2000个节点，则会新增一个该ConfigMap文件。

**表 1**  cluster-info-node-cm

<a name="table25031946405"></a>
|参数|说明|
|--|--|
|mindx-dl-nodeinfo-*<kwok-node-0>*|前缀为固定的mindx-dl-nodeinfo，kwok-node-0是节点名称，方便定位故障的具体节点。|
|NodeInfo|节点维度的故障信息。|
|FaultDevList|节点故障设备列表。|
|- DeviceType|故障设备类型。|
|- DeviceId|故障设备ID。|
|- FaultCode|故障码，由英文和数组拼接而成的字符串，字符串表示故障码的十六进制。|
|- FaultLevel|故障处理等级。<ul><li>NotHandleFault：无需处理。</li><li>PreSeparateFault：该节点上有任务则不处理，后续调度时不调度任务到该节点。</li><li>SeparateFault：任务重调度。</li></ul>|
|NodeStatus|节点健康状态，由本节点故障处理等级最严重的设备决定。<ul><li>Healthy：该节点故障处理等级存在且不超过NotHandleFault，该节点为健康节点，可以正常训练。若该节点故障处理等级为PreSeparateFault，且节点有NPU卡正在使用，则该节点为健康节点。任务执行完成后，该节点将变为故障节点。</li><li>UnHealthy：该节点故障处理等级存在SeparateFault，该节点为故障节点，将影响训练任务，立即将任务调离该节点。若该节点故障处理等级为PreSeparateFault，且节点无NPU卡正在使用，则该节点为故障节点，不可将任务调度到该节点。</li></ul>|


**表 2** cluster-info-device-$\{m\}

<a name="table915714719368"></a>
|参数|说明|
|--|--|
|mindx-dl-deviceinfo-*<kwok-node-0>*|前缀为固定的mindx-dl-deviceinfo，kwok-node-0是节点名称，用于定位故障的具体节点。|
|huawei.com/Ascend910|当前节点可用的芯片名称信息，存在多个时用英文逗号拼接。|
|huawei.com/Ascend910-NetworkUnhealthy|当前节点网络不健康的芯片名称信息，存在多个时用英文逗号拼接。|
|huawei.com/Ascend910-Unhealthy|当前芯片不健康的芯片名称信息，存在多个时用英文逗号拼接。|
|huawei.com/Ascend910-Fault|数组对象，对象包含fault_type、npu_name、large_model_fault_level、 fault_level、fault_handling、fault_code和fault_time_and_level_map字段。|
|- fault_type|故障类型。<ul><li>CardUnhealthy：芯片故障</li><li>CardNetworkUnhealthy：参数面网络故障（芯片网络相关故障）</li><li>NodeUnhealthy：节点故障</li><li>PublicFault：公共故障</li></ul>|
|- npu_name|故障的芯片名称，节点故障时为空。|
|<p>- large_model_fault_level</p><p>- fault_level</p><p>- fault_handling</p>|故障处理类型，节点故障时取值为空。<ul><li>NotHandleFault：不做处理</li><li>RestartRequest：推理场景需要重新执行推理请求，训练场景重新执行训练业务</li><li>RestartBusiness：需要重新执行业务</li><li>FreeRestartNPU：影响业务执行，待芯片空闲时需复位芯片</li><li>RestartNPU：直接复位芯片并重新执行业务</li><li>SeparateNPU：隔离芯片</li><li>PreSeparateNPU：预隔离芯片，会根据训练任务实际运行情况判断是否重调度</li></ul><div class="note"><span>说明：</span><ul><li>large_model_fault_level、fault_handling和fault_level参数功能一致，推荐使用fault_handling。</li><li>若推理任务订阅了故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</li></ul>|
|- fault_code|故障码，英文逗号拼接的字符串。|
|- fault_time_and_level_map|故障码、故障发生时间及故障处理等级。|
|SuperPodID|超节点ID。|
|ServerIndex|当前节点在超节点中的相对位置。<ul><li>驱动上报的SuperPodID或ServerIndex的值为0xffffffff时，SuperPodID或ServerIndex的取值为-1。</li><li>存在以下情况，SuperPodID或ServerIndex的取值为-2。</li><ul><li>当前设备不支持查询超节点信息。</li><li>因驱动问题导致获取超节点信息失败。</li></ul></ul>|


**表 3**  cluster-info-switch-$\{x\}

<a name="table9246232250"></a>
|参数|说明|
|--|--|
|FaultCode|当前节点的灵衢总线设备故障码列表。数组对象包含EventType、AssembledFaultCode、PeerPortDevice、PeerPortId、SwitchChipId、SwitchPortId、Severity、Assertion、AlarmRaisedTime等字段。|
|-EventType|告警ID。|
|-AssembledFaultCode|故障码。|
|-PeerPortDevice|对接设备类型。<ul><li>0：CPU</li><li>1：NPU</li><li>2：SW</li><li>0xFFFF：NA</li></ul>|
|-PeerPortId|对接设备ID。|
|-SwitchChipId|灵衢故障芯片ID。从0开始编号。|
|-SwitchPortId|灵衢故障端口ID。从0开始编号。|
|-Severity|故障等级。<ul><li>0：提示</li><li>1：次要</li><li>2：重要</li><li>3：紧急</li></ul>|
|-Assertion|事件类型。<ul><li>0：故障恢复</li><li>1：故障产生</li><li>2：通知类事件</li></ul>|
|-AlarmRaisedTime|故障/事件产生时间。|
|FaultLevel|当前节点故障处理等级。<p>取FaultCode中所有故障中等级最高的故障等级，取值包含：NotHandle、SubHealthFault、Separate和RestartRequest。</p>|
|UpdateTime|故障上报刷新时间。|
|NodeStatus|当前节点健康状态。<p>对应FaultLevel取值，NotHandle:Healthy、SubHealthFault:SubHealthy、Separate:UnHealthy和RestartRequest:UnHealthy。</p>|
|FaultTimeAndLevelMap|故障发生时间及故障处理等级列表。数组对象包含故障码、灵衢故障芯片ID、灵衢故障端口ID、fault_time和fault_level字段。键值为故障码、灵衢故障芯片ID、灵衢故障端口ID，由下划线连接组成。|
|-fault_time|故障发生时间。|
|-fault_level|故障处理等级。|


**statistic-fault-info<a name="section1153232554520"></a>**

该ConfigMap位于用户创建的cluster-system命名空间下，Label为mc-statistic-fault=true。用于展示集群中的故障信息（当前仅展示公共故障信息）。

**表 4**  Data数据信息说明

|参数|说明|
|--|--|
|PublicFaults|公共故障详情。故障数量过大时，不再更新本字段内容。以下各字段的详细说明请参见<a href="#公共故障接口">故障信息说明表</a>。|
|-<i>\<node name></i>|故障节点名称|
|-resource|故障发送方<p>默认配置为CCAE、fd-online、pingmesh、Netmind。</p>|
|-devIds|故障芯片物理ID|
|-faultId|故障实例ID|
|-type|故障类型<ul><li>NPU：芯片故障。</li><li>Node：节点故障。</li><li>Network：网络故障。</li><li>Storage：存储故障。</li></ul>|
|-faultCode|故障码|
|-level|故障级别<ul><li>NotHandleFault：暂不处理。</li><li>SubHealthFault：亚健康。</li><li>SeparateNPU：无法恢复，需要隔离芯片。</li><li>PreSeparateNPU：暂不影响业务，后续不再调度任务到该芯片。</li></ul>|
|-faultTime|故障产生时间|
|FaultNum|故障数量|
|-publicFaultNum|所有节点的公共故障数量之和。|
|Description|公共故障数量过大时的提示信息。|
|<p>说明：</p><ul><li>公共故障对外展示1M数据，大约4500条。</li><li>超过4500条时，部分数据不再对外展示，ConfigMap中会新增Description内容进行提示，内部缓存正常运行。</li></ul>|


**cluster-system super-pod-<super-pod-id\><a name="section53741611135414"></a>**

该ConfigMap位于用户创建的cluster-system命名空间下，Label为app=pingmesh。

**表 5**  cluster-system super-pod-<super-pod-id\>

|参数|说明|
|--|--|
|app|NodeD识别ConfigMap所需的Label key，取值为pingmesh。|
|superPodDevice|超节点信息的key。|
|SuperPodID|超节点ID|
|NodeDeviceMap|超节点中包含的所有节点信息。|
|NodeName|节点名称|
|DeviceMap|节点中的所有NPU信息，格式为physicID: superDeviceID。|


**fault-job-info<a name="section1548342116513"></a>**

该ConfigMap位于用户创建的cluster-system命名空间下。用于展示集群中需要强制释放通信资源的故障任务信息。仅在Atlas 900 A3 SuperPoD 超节点进行进程级别重调度时生效。

**表 6**  fault-job-info

|参数|说明|取值|
|--|--|--|
|SdIds|故障卡的SDID。|字符串序列|
|NodeNames|需要强制释放资源的节点名。|字符串序列|
|FaultTimes|发生故障的时间。|64位整数类型|
|JobId|任务的UID。|字符串|



## 任务信息<a name="ZH-CN_TOPIC_0000002511426769"></a>

**job-summary-<任务名称\><a name="section24017282404"></a>**

**表 1**  job-summary-任务名称 ConfigMap字段说明

|参数|说明|取值|
|--|--|--|
|hccl.json|任务使用的芯片通信信息。可转义为JSON格式，字段说明如下：<ul><li>status：任务RankTable是否已经生成。</li><ul><li>initializing：还在为任务分配设备，RankTable未生成。</li><li>complete：当RankTable生成后，状态会立即变为complete，同步出现server_list等其他字段。</li></ul><li>server_list：任务设备分配情况。</li><ul><li>device：记录NPU分配，NPU IP和rank_id信息。</li><li>server_id：AI Server标识，全局唯一。</li><li>server_name：节点名称。</li><li>server_sn：节点的SN号。需要保证设备的SN存在。若不存在，请联系华为技术支持。</li></ul><li>server_count：任务使用的节点数量。</li><li>version：版本信息。</li></ul>|字符串|
|job_id|任务的K8s ID信息。|字符串|
|operator|<ul><li>add：接收到添加任务命令后状态更新为add。</li><li>delete：接收到删除任务命令后状态更新为delete。</li></ul>|字符串|
|deleteTime|任务被删除的时间。|字符串|
|sharedTorIp|任务使用的共享交换机信息。|字符串|
|masterAddr|PyTorch训练时指定的MASTER_ADDR值。|字符串|
|total|ConfigMap的个数。|整数类型|
|time|任务开始时间。|字符串|
|framework|任务使用的框架。|字符串|
|job_status|任务状态，存在以下几种状态。<ul><li>Pending</li><li>Running</li><li>Complete</li><li>Failed</li></ul>|字符串|
|job_name|任务名称。|字符串|
|cm_index|当前ConfigMap的序号。|字符串|


**current-job-statistic<a name="section39901331194218"></a>**

用于展示集群中当前任务的统计信息，记录在/var/log/mindx-dl/clusterd/event\_job.log日志文件中。由于K8s的ConfigMap容量大小限制，最大支持统计集群任务数量约为1w条。当日志文件达到20M时，触发自动转储，最多保存5份转储日志，转储日志最长保留时间为40天。

|参数|说明|
|--|--|
|data|-|
|- ID|K8s集群分配的Job ID。|
|- customID|用户自定义的Job ID，如果内容为空则不展示。|
|- cardNum|任务使用的卡的数量，如果内容为空则不展示。|
|- podFirstRunTime|任务Pod第一次全部running的时间，如果内容为空则不展示。|
|- stopTime|任务Pod全部complete或者被强行删除的时间，如果内容为空则不展示。|
|- podLastRunTime|任务Pod上一次全部恢复running的时间，如果内容为空则不展示。|
|- podLastFaultTime|任务Pod上一次部分或者全部failed的时间，如果内容为空则不展示。|
|- podFaultTimes|任务故障导致Pod重调度的次数，如果次数为0则不展示。|
|totalJob|当前集群中的总任务数。|



## 进程级恢复接口<a name="ZH-CN_TOPIC_0000002511346765"></a>

### gRPC接口<a name="ZH-CN_TOPIC_0000002511346735"></a>

#### Register（公共接口）<a name="ZH-CN_TOPIC_0000002511346739"></a>

**功能说明<a name="section143314311911"></a>**

接收处理客户端的注册请求，为进程级恢复做初始化准备。作业调用Init成功后，需要等待客户端确认MindIO侧进程级别重调度和进程级在线恢复开关为打开状态后，调用此接口。 Register成功后进程级别重调度和进程级在线恢复功能才会处于可用状态。

**函数原型<a name="section3958124212115"></a>**

```
rpc Register(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID</p><p>**ClientInfo.role**：客户端角色</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<ul><li>取值为0：表示注册成功。</li><li>其他值：表示注册失败。</li></ul></p><p>**Status.info**：返回信息描述。</p>|



#### Init<a name="ZH-CN_TOPIC_0000002479386824"></a>

**功能说明<a name="section83882209338"></a>**

用于初始化进程级别重调度和进程级在线恢复，初始化成功后，进程级别重调度和进程级在线恢复功能将暂时处于不可用状态。

**函数原型<a name="section2049633816332"></a>**

```
rpc Init(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID</p><p>**ClientInfo.role**：客户端角色</p>|


**返回值说明<a name="section1864651893415"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<ul><li>取值为0：表示注册成功。</li><li>其他值：表示注册失败。</li></ul></p><p>**Status.info**：返回信息描述。</p>|



#### SubscribeProcessManageSignal<a name="ZH-CN_TOPIC_0000002511426713"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅进程控制信号请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```
 rpc SubscribeProcessManageSignal(ClientInfo) returns (stream ProcessManageSignal){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p>|


**发送数据说明<a name="section10140143475520"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ProcessManageSignal|<p>message FaultRank{<p>string rankId = 1;</p><p>string faultType = 2;</p>}</p><p>message ProcessManageSignal{<p>string uuid=1;</p><p>string jobId = 2;</p><p>string signalType = 3;</p><p>repeated string actions = 4;</p><p>repeated FaultRank faultRanks = 5;</p><p>string changeStrategy = 6;</p><p>int64 timeout = 7;</p>}</p>|<p>**rankId**：string类型，故障卡ID</p><p>**faultType**：string类型，故障类型</p><p>**uuid**：string类型，本次signal的uuid</p><p>**jobId**：string类型，训练的任务ID</p><p>**signalType**：string类型，signal类型</p><p>**actions**：repeated string，要执行的动作</p><p>**faultRanks**：repeated FaultRank，故障卡信息</p><p>**changeStrategy**：string类型，要执行的恢复策略</p><p>**timeout**：int64类型，超时时间</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|
|nodeRankIds|string数组|故障节点Node Rank ID。|
|extraParams|string|以JSON字符串形式传递扩缩容具体策略信息，通过TaskD透传给MindIO，最终传递给callback回调函数进行解析。|



#### ReportStopComplete<a name="ZH-CN_TOPIC_0000002511426707"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端上报暂停训练进程是否成功。

**函数原型<a name="section3958124212115"></a>**

```
rpc ReportStopComplete(StopCompleteRequest) returns (Status){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StopCompleteRequest|message StopCompleteRequest{<p>string jobId = 1;</p><p>Status status = 2;</p><p>repeated FaultRank faultRankIds = 3;</p>}|<p>**StopCompleteRequest.jobId**：任务ID。</p><p>**StopCompleteRequest.status.code**：返回码，OK表示暂停训练成功，其他值表示暂停训练失败。</p><p>**StopCompleteRequest.status.info**：返回信息描述。</p><p>**StopCompleteRequest.faultRankIds**：故障芯片全局故障Rank列表。FaultRank是一组包含故障信息的键值对，由rankId（全局Rank ID）和faultType（故障类型）组成。faultType取值为0时，代表片上内存故障。取值为1时，表示其他故障。</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{int32 code = 1;string info = 2;}|<p>**Status.code**：返回码。<ul><li>取值为0：表示故障恢复流程正常</li><li>其他值：表示故障恢复流程异常，并触发重调度。</li></ul></p><p>**Status.info**：返回信息描述。</p>|



#### ReportRecoverStrategy<a name="ZH-CN_TOPIC_0000002511346747"></a>

**功能说明<a name="section16150748174520"></a>**

接收客户端上报的当前任务支持的故障恢复策略。

**函数原型<a name="section3958124212115"></a>**

```
 rpc ReportRecoverStrategy(RecoverStrategyRequest) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|RecoverStrategyRequest|message RecoverStrategyRequest{<p>string jobId = 1;</p><p>repeated FaultRank faultRankIds = 2;</p><p>repeated string strategies = 3;</p>}|<p>**RecoverStrategyRequest.jobId**：任务ID</p><p>**RecoverStrategyRequest.faultRankIds**：故障芯片全局故障Rank列表。FaultRank是故障信息的键值对，包含rankId（全局Rank ID）和faultType（故障类型）。faultType取值为0时，代表片上内存故障。取值为1时，表示其他故障。</p><p>**RecoverStrategyRequest.strategies**：当前任务支持的恢复策略。</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<ul><li>0：表示故障恢复流程正常。</li><li>其他值：表示恢复流程异常，并触发重调度。</li></ul></p><p>**Status.info**：返回信息描述。</p>|



#### ReportRecoverStatus<a name="ZH-CN_TOPIC_0000002511346753"></a>

**功能说明<a name="section16150748174520"></a>**

接收客户端上报的当前任务恢复情况。

**函数原型<a name="section3958124212115"></a>**

```
rpc ReportRecoverStatus(RecoverStatusRequest) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|RecoverStatusRequest|message RecoverStatusRequest{<p>string jobId = 1;</p><p>Status status = 2;</p><p>string strategy = 3;</p><p>repeated string isolateRankIds = 4;</p>}|<p>**RecoverStatusRequest.jobId**：任务ID。</p><p>**RecoverStatusRequest.status.code**：任务恢复情况状态码。<ul><li>0：表示任务恢复成功。</li><li>其他值：表示失败。</li></ul></p><p>**RecoverStatusRequest.status.info**：任务恢复情况描述。</p><p>**RecoverStatusRequest.strategy**：恢复策略名称。</p><p>**RecoverStatusRequest.isolateRankIds**：MindIO上报缩容时需要隔离的Rank列表。</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<ul><li>0：表示故障恢复流程正常。</li><li>其他值：表示恢复流程异常，并触发重调度。</li></ul></p><p>**Status.info**：返回信息描述。</p>|



#### ReportProcessFault<a name="ZH-CN_TOPIC_0000002511346729"></a>

**功能说明<a name="section16150748174520"></a>**

接收客户端上报的故障芯片全局Rank信息。

**函数原型<a name="section3958124212115"></a>**

```
rpc ReportProcessFault(ProcessFaultRequest) returns (Status){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ProcessFaultRequest|message ProcessFaultRequest{<p>string jobId = 1;</p><p>repeated FaultRank faultRankIds = 2;</p>}|<p>**ProcessFaultRequest.jobId**：任务ID。</p><p>**ProcessFaultRequest.faultRankIds**：故障芯片全局Rank ID列表。FaultRank是故障信息的键值对，包含rankId（全局Rank ID）和faultType（故障类型）。faultType取值为0时，代表片上内存故障。取值为1时，表示其他故障。</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<li>0：表示恢复流程正常。</li><li>其他值：表示故障恢复流程异常，并触发重调度。</li></p><p>**Status.info**：返回信息描述。</p>|



#### HealthCheck<a name="ZH-CN_TOPIC_0000002511426765"></a>

**功能说明<a name="section16150748174520"></a>**

检查gRPC链接状态。

**函数原型<a name="section3958124212115"></a>**

```
rpc HealthCheck(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<li>0：表示故障恢复流程正常。</li><li>其他值：表示故障恢复流程异常，并触发重调度。</li></p><p>**Status.info**：返回信息描述。</p>|




### 对接第三方AI平台控制相关接口<a name="ZH-CN_TOPIC_0000002511346803"></a>

**功能说明<a name="section3323912134319"></a>**

AI平台可通过Pod Group Annotation控制故障恢复的流程以及恢复策略。例如，平台写入Pod Group Annotation key：ProcessRecoverStrategy，且value为空时，故障恢复会被卡住，直到平台写入具体的恢复策略才会继续走恢复的流程。

**Pod Group Annotation<a name="section1313991113387"></a>**

**表 1**  参数说明

|参数|取值|说明|
|--|--|--|
|ProcessRecoverStrategy|<ul><li>retry</li><li>recover</li><li>dump</li><li>空或none</li><li>字段不存在</li></ul>|<ul><li>retry：平台启动恢复，策略为进程级在线恢复</li><li>recover：平台启动恢复，策略为在线恢复</li><li>dump：平台启动恢复，策略为保存临终遗言</li><li>空或none：等待平台决策</li><li>字段不存在：关闭进程级恢复</li></ul>|
|ProcessConfirmFault|string|ClusterD刷新后的故障键值对列表，格式为“id1:type1,id2:type2”的字符串。id表示全局rankId，type表示故障类型。type为0表示故障卡只有片上内存故障，1表示至少有一个非片上内存故障。|
|ProcessResultFault|string|平台确认的故障键值对列表，格式为“id1:type1,id2:type2”的字符串。id表示全局rankId，type表示故障类型。type为0表示故障卡只有片上内存故障，1表示至少有一个非片上内存故障。|
|RankTableReady|<ul><li>true</li><li>false或其他值</li><li>字段不存在</li></ul>|<ul><li>true：平台已生成完成RankTable</li><li>false或其他值：平台暂未生成完成RankTable</li><li>字段不存在：非RankTable模式</li></ul>|
|ProcessRecoverStatus|<ul><li>retry-success</li><li>retry-failed</li><li>recover-success</li><li>recover-failed</li><li>dump-success</li><li>dump-failed</li><li>exit-completed</li><li>空值或其他值</li></ul>|<ul><li>retry-success：进程级在线恢复成功</li><li>retry-failed：进程级在线恢复失败</li><li>recover-success：在线恢复成功</li><li>recover-failed：在线恢复失败</li><li>dump-success：保存临终遗言成功</li><li>dump-failed：保存临终遗言失败</li><li>exit-completed</li><li>空值或其他值：未恢复完成</li></ul>|




## 公共故障接口<a name="ZH-CN_TOPIC_0000002479226838"></a>

### ConfigMap<a name="ZH-CN_TOPIC_0000002479386788"></a>

**功能说明<a name="section359310211618"></a>**

接收公共故障的ConfigMap信息，接入断点续训流程。

>[!NOTE] 说明 
>-   实际的ConfigMap中的参数如果与定义的取值范围不相符，ClusterD会将故障信息丢弃，不作处理。
>-   通过ConfigMap或者gRPC接口注入的公共故障，所有节点的故障数量之和上限为5w。当故障数量超过5w时，再次注入故障，ClusterD会将故障信息丢弃，不作处理。
>-   ConfigMap的Label需要为mc-consumer-publicfault=true，Data的key需要为PublicFault。
>-   通过ConfigMap方式发送公共故障时，单次数据量不能超过1M大小，否则ConfigMap会更新失败。

**参数说明<a name="section4809204015614"></a>**

具体的参数说明见下表。

**表 1**  故障信息说明

|参数名称|含义|取值|类型|是否必填|
|--|--|--|--|--|
|id|消息唯一标识|8到128个字符的字符串，支持大小写字母、数字、中划线（-）、下划线（_）和点（.），保证唯一性。|string|是|
|timestamp|消息发送的时间戳|时间戳（单位：ms），13位数字，必须在2025-01-01T00:00:00Z之后。|int64|是|
|version|消息版本号|取值为1.0。|string|是|
|resource|故障发送方|默认配置为CCAE、fd-online、pingmesh、Netmind、dpcStorage。<ul><li>公共故障的故障发送方，必须存在于故障配置文件的publicFaultResource中。</li><li>对于新增的故障发送方，需要将其手动配置到故障配置文件中。详细说明请参见<a href="../usage/resumable_training.md#可选配置公共故障的级别和发送方">（可选）配置公共故障的级别和发送方</a>。</li></ul>|string|是|
|faults|故障内容|切片，长度>0且≤100。|[]object, fault|是|


**表 2**  fault字段说明

|参数名称|含义|取值|类型|是否必填|
|--|--|--|--|--|
|faultId|故障实例ID|8到128个字符的字符串，支持大小写字母、数字、中划线（-）、下划线（_）和点（.），保证唯一性。<p>同一个故障实例，faultId需要保证唯一性。</p>|string|是|
|faultType|故障类型|取值为NPU、Node、Network或Storage。<ul><li>NPU：芯片故障。</li><li>Node：节点故障。</li><li>Network：网络故障。</li><li>Storage：存储故障。</li></ul>该字段在cluster-info-cm中展示为“PublicFault”。|string|是|
|faultCode|故障码|用户可以自定义，9位唯一即可。<ul><li>接入断点续训的故障码，必须存在于故障配置文件的publicFaultCode中。</li><li>对于新增的故障码，需要在故障配置文件配置其故障级别。详细说明请参见<a href="../usage/resumable_training.md#可选配置公共故障的级别和发送方">（可选）配置公共故障的级别和发送方</a>。</li><li>故障码建议遵循故障码说明表中的规则定义，方便后续维护。</li><li>若一张NPU先后出现两个相同的故障码，在cluster-info-cm中fault_code字段将同时记录2个相同的故障码。</li></ul>|string|是|
|faultTime|故障产生时间|时间戳（单位：ms），13位数字，必须在2025-01-01T00:00:00Z之后。<ul><li>无论是故障产生还是故障消除，该字段均为故障产生时间。</li><li>该字段在cluster-info-cm中以秒为单位展示。</li></ul>|int64|是|
|assertion|故障状态|取值为occur、recover或once。<ul><li>occur：故障产生。</li><li>recover：故障恢复。</li><li>once：一次性事件。</li></ul><div class="note"><span>[!NOTE] 说明</span><div class="notebody"><ul><li>公共故障消除需要将相应故障的recover事件写入ConfigMap中，不能通过删除ConfigMap的形式实现。</li><li>对于一次性事件，几秒钟之后故障会自动清除。</li></ul>|string|是|
|faultLocation|故障定位信息|故障源信息，长度≤10，map的key长度≤16，value长度≤128。eg. key: npuIp, value: ip|map[string]string|否|
|influence|故障影响的范围|切片，长度>0且≤1000。|[]object, faultInfo|是|
|description|故障描述|0~512个字符。包含非空白字符和空格。|string|否|


**表 3**  faultInfo字段说明

|参数名称|含义|取值|类型|是否必填|
|--|--|--|--|--|
|nodeName|节点名称。可通过**kubectl get nodes -owide**命令查询。|1到253个字符的字符串，支持小写字母、数字、中划线（-）和点（.），必须以字母数字开头和结尾。该字段存在时，就不使用nodeSN。<p>如果节点名称不存在于K8s集群中，ClusterD不会提示节点名称错误，但是不会将该故障信息写入cluster-info-device-cm。</p>|string|nodeName与nodeSN二选一|
|nodeSN|节点SN号|节点的SN号。取值为NodeD写入的节点annotation，key为product-serial-number。<p>若使用该字段而不使用nodeName，需要提前安装NodeD组件。</p>|string|nodeName与nodeSN二选一|
|deviceIds|芯片物理ID|长度(0, 32]，每个元素的取值[0, 32)，且不允许重复。<ul><li>如果无法准确找到故障的芯片，需要填入节点上的所有芯片物理ID。</li><li>如果传入一个节点上不存在的芯片物理ID，ClusterD也会将其展示在cluster-info-device-cm中。</li></ul>|[]int32|是|



### gRPC接口<a name="ZH-CN_TOPIC_0000002479226854"></a>

**功能说明<a name="section125411749115817"></a>**

接收处理gRPC客户端的公共故障发送请求，接入断点续训流程。

> [!NOTE] 说明 
>-   实际的gRPC请求参数如果与定义的取值范围不相符，ClusterD会将故障信息丢弃，不作处理。
>-   通过ConfigMap或者gRPC接口注入的公共故障，所有节点的故障数量之和上限为5w。当故障数量超过5w时，再次注入故障，ClusterD会将故障信息丢弃，不作处理。
>-   公共故障消除需要将相应故障的recover事件通过gRPC接口发送给ClusterD。

**函数原型<a name="section1698941035919"></a>**

```
rpc SendPublicFault(PublicFaultRequest) returns (RespStatus){}
```

**输入参数说明<a name="section52771657118"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|PublicFaultRequest|<p>message PublicFaultRequest{<p>string id = 1;</p><p>int64 timestamp = 2;</p><p>string version = 3;</p><p>string resource = 4;</p><p>repeated Fault faults = 5;</p>}</p><p>message Fault{<p>string faultId = 1;</p><p>string faultType = 2;</p><p>string faultCode = 3;</p><p>int64 faultTime = 4;</p><p>string assertion = 5;</p><p>map<string, string> faultLocation = 6;</p><p>repeated PubFaultInfo influence = 7;</p><p>string description = 8;</p>}</p><p>message PubFaultInfo{<p>string nodeName = 1;</p><p>string nodeSN = 2;</p><p>repeated int32 deviceIds = 3;</p>}</p>|<p>**PublicFaultRequest.id**：消息唯一标识</p><p>**PublicFaultRequest.timestamp**：消息发送的时间戳</p><p>**PublicFaultRequest.version**：消息版本号</p><p>**PublicFaultRequest.resource**：故障发送方</p><p>**PublicFaultRequest.faults**：故障内容</p><p>**Fault.faultId**：故障实例ID</p><p>**Fault.faultType**：故障类型</p><p>**Fault.faultCode**：故障码</p><p>**Fault.faultTime**：故障产生时间</p><p>**Fault.assertion**：故障状态</p><p>**Fault.faultLocation**：故障定位信息</p><p>**Fault.influence**：故障影响的范围</p><p>**Fault.description**：故障描述</p><p>**PubFaultInfo.nodeName**：节点名称</p><p>**PubFaultInfo.nodeSN**：节点SN号</p><p>**PubFaultInfo.deviceIds**：芯片物理ID</p><p>以上参数的详细说明及取值情况请参见<a href="#公共故障接口">ConfigMap</a>。</p>|


**返回值说明<a name="section521319321415"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|RespStatus|message RespStatus{<p>int32 code = 1;</p><p>string info = 2;</p>}|**RespStatus.code**：返回码。<ul><li>取值为0：表示故障发送成功。</li><li>其他值：表示故障发送失败。409表示请求参数有误，410表示消息发送频率超限。</li></ul>**RespStatus.info**：返回信息描述。|




## 性能劣化故障接口<a name="ZH-CN_TOPIC_0000002479226802"></a>

### ModifyTrainingDataTraceSwitch<a name="ZH-CN_TOPIC_0000002511426771"></a>

**功能说明<a name="section22878209356"></a>**

外部调用修改各类数据动态打点开关能力。

>[!NOTE] 说明 
>如果通过ClusterD提供的gRPC接口这种方式开启或修改轻量profiling获取落盘数据，创建的data-trace-<任务名称\> ConfigMap的生命周期会随着任务的删除而删除。当任务不存在的时候，该接口会调用失败。

**函数原型<a name="section1472624833519"></a>**

```
rpc ModifyTrainingDataTraceSwitch (DataTypeReq) returns (DataTypeRes)
```

**输入参数说明<a name="section6782115723515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataTypeReq|<p>message DataTypeReq{<p>string jobNsName = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p>}</p><p>message ProfilingSwitch{<p>string CommunicationOperator = 1;</p><p>string Step = 2;</p><p>string SaveCheckpoint = 3;</p><p>string FP =4;</p><p>string DataLoader =5;</p>}</p>|<p>**jobNsName**：所需修改的任务的命名空间和任务名称，以’/’拼接，如：default/test-pytorch。</p><p>**profilingSwitch**：各类开关详情。<ul><li>**CommunicationOperator**：通信算子开关。</li><li>**Step**：Step时延开关。</li><li>**SaveCheckpoint**：SaveCheckpoint耗时开关。</li><li>**FP**：前向传播数据开关。</li><li>**DataLoader**：DataLoader耗时开关。</li></ul>|


**返回值说明<a name="section7920469381"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataTypeRes|message DataTypeRes{<p>string message = 1;</p><p>int32 code = 2;</p>}|<p>**message**：接口调用结果信息。</p><p>**code**：接口调用返回码。</p><ul><li>1：300，入参不合法。</li><li>2：404，无法查询ConfigMap。</li><li>3：500，服务端异常。</li><li>4：200，接口正常返回。</li></ul>|



### GetTrainingDataTraceSwitch<a name="ZH-CN_TOPIC_0000002479386852"></a>

**功能说明<a name="section21882190424"></a>**

外部调用获取各类数据动态打点开关状态。

**函数原型<a name="section1723573217426"></a>**

```
rpc GetTrainingDataTraceSwitch (DataStatusReq) returns (DataStatusRes)
```

**输入参数说明<a name="section19921040164215"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataStatusReq|message DataStatusReq{<p>string jobNsName = 1;</p>}|**jobNsName**：所需修改的任务的命名空间和任务名称，以’/’拼接，如：default/test-pytorch。|


**返回值说明<a name="section93011951104217"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataStatusRes|message DataStatusRes{<p>string message = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p><p>int32 code = 3;</p>}|<p>**message**：接口调用结果信息。</p><p>**profilingSwitch**：各类开关详情。</p><ul><li>**CommunicationOperator**：通信算子开关。</li><li>**Step**：Step时延开关。</li><li>**SaveCheckpoint**：SaveCheckpoint耗时开关。</li><li>**FP**：前向传播数据开关。</li><li>**DataLoader**：DataLoader耗时开关。</li></ul>**code**：接口调用返回码。<ul><li>1：300，入参不合法。</li><li>2：404，无法查询ConfigMap。</li><li>3：500，服务端异常。</li><li>4：200，接口正常返回。</li></ul>|



### SubscribeDataTraceSwitch<a name="ZH-CN_TOPIC_0000002511346751"></a>

**功能说明<a name="section22878209356"></a>**

外部订阅各类数据动态打点开关状态。

**函数原型<a name="section1472624833519"></a>**

```
rpc SubscribeDataTraceSwitch (ProfilingClientInfo) returns (stream DataStatusRes)
```

**输入参数说明<a name="section6782115723515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ProfilingClientInfo|message ProfilingClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**jobId**：任务ID。</p><p>**role**：客户端角色。</p>|


**返回值说明<a name="section7920469381"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataStatusRes|message DataStatusRes{<p>string message = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p><p>int32 code = 3;</p>}|<p>**message**：接口调用结果信息。</p><p>**profilingSwitch**：各类开关详情。</p><ul><li>**CommunicationOperator**：通信算子开关。</li><li>**Step**：Step时延开关。</li><li>**SaveCheckpoint**：SaveCheckpoint耗时开关。</li><li>**FP**：前向传播数据开关。</li><li>**DataLoader**：DataLoader耗时开关。</li></ul>**code**：接口调用返回码。<ul><li>1：300，入参不合法。</li><li>2：404，无法查询ConfigMap。</li><li>3：500，服务端异常。</li><li>4：200，接口正常返回。</li></ul>|




## 业务配置接口<a name="ZH-CN_TOPIC_0000002479226840"></a>

### Register<a name="ZH-CN_TOPIC_0000002511426719"></a>

**功能说明<a name="section143314311911"></a>**

接收处理客户端的注册请求，为订阅相关业务配置做初始化准备。

**函数原型<a name="section3958124212115"></a>**

```
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



### SubscribeRankTable<a name="ZH-CN_TOPIC_0000002511346779"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅RankTable请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```
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
|RankTableStream|<p>message RankTableStream{</p><p>string jobId = 1;</p><p>string rankTable = 2;</p>}|<p>**RankTableStream.jobId**：任务ID。</p><p>**RankTableStream.rankTable**：RankTable信息，各字段的详细说明如<a href="#table5843145110294">表1</a>所示。</p>|


**global-ranktable文件说明<a name="section268935611912"></a>**

ClusterD会生成global-ranktable在RankTable字段作为返回消息。global-ranktable中部分字段来自于hccl.json文件，关于hccl.json文件的详细说明请参见[hccl.json文件说明](../appendix.md#hccljson文件说明)。

-   示例如下。

    ```
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

-   Atlas A3 训练系列产品示例如下。

    ```
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
|super_device_id|Atlas A3 训练系列产品超节点内NPU的唯一标识|
|rank_id|NPU对应的训练rank ID|
|device_logical_id|NPU的逻辑ID|
|super_pod_list|超节点列表|
|super_pod_id|逻辑超节点ID|




## 故障服务接口<a name="ZH-CN_TOPIC_0000002479386826"></a>

### Register<a name="ZH-CN_TOPIC_0000002511426773"></a>

**功能说明<a name="section143314311911"></a>**

接收处理客户端的注册请求，为订阅故障信息等功能做初始化准备。

**函数原型<a name="section3958124212115"></a>**

```
rpc Register(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p><span>说明：</span><ul><li>传入jobId为空时，表示注册集群所有任务。</li><li>传入jobId不为空时，表示注册指定任务。</li></ul>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示注册成功。</li><li>其他值：表示注册失败。</li></ul>**Status.info**：返回信息描述。|



### SubscribeFaultMsgSignal<a name="ZH-CN_TOPIC_0000002511426699"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅故障信息请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

>[!NOTE] 说明 
>-   调用此接口前，需先调用[Register接口](#故障服务接口)。
>-   客户端订阅通算任务的故障信息后，只能收到NodeD故障和K8s节点状态异常故障。

**函数原型<a name="section3958124212115"></a>**

```
rpc SubscribeFaultMsgSignal(ClientInfo) returns (stream FaultMsgSignal){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。<span>说明：</span><ul><li>传入jobId为空时，获取的结果为集群内所有job的故障。</li><li>传入jobId不为空时，获取的结果为任务所属节点的故障。</li></ul>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|


**发送数据说明<a name="section112224012419"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|FaultMsgSignal|<p>message FaultMsgSignal{</p><p>string uuid = 1;</p><p>string jobId = 2;</p><p>string signalType = 3;</p><p>repeated NodeFaultInfo nodeFaultInfo = 4;</p>}</p><p>message NodeFaultInfo{<p>string nodeName = 1;</p><p>string nodeIP = 2;</p><p>string nodeSN = 3;</p><p>string faultLevel = 4;</p><p>repeated DeviceFaultInfo faultDevice = 5;</p>}</p><p>message DeviceFaultInfo{<p>string deviceId = 1;</p><p>string deviceType = 2;</p><p>repeated string faultCodes = 3;</p><p>string faultLevel = 4;</p><p>repeated string faultType = 5;</p><p>repeated string faultReason = 6;</p><p>repeated SwitchFaultInfo switchFaultInfos = 7;</p><p>repeated string faultLevels = 8;</p>}</p><p>message SwitchFaultInfo{<p>string faultCode = 1;</p><p>string switchChipId = 2;</p><p>string switchPortId = 3;</p><p>string faultTime = 4;</p><p>string faultLevel = 5;</p>}</p>|<p>**FaultMsgSignal.uuid**：消息ID</p><p>**FaultMsgSignal.jobId**：任务ID</p><p>**FaultMsgSignal.signalType**：消息类型，“fault”代表故障发生，“normal”代表无故障或故障恢复</p><p>**FaultMsgSignal.nodeFaultInfo**：节点故障信息</p><p>**NodeFaultInfo.nodeName**：故障节点名称</p><p>**NodeFaultInfo.nodeIP**：节点IP</p><p>**NodeFaultInfo.nodeSN**：节点SN号</p><p>**NodeFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，设置为DeviceFaultInfo.faultLevel中最严重的级别</p><p>**NodeFaultInfo.faultDevice**：设备故障信息</p><p>**DeviceFaultInfo.deviceId**：设备ID。当节点发生总线设备故障和K8s状态异常故障时，deviceId为-1</p><p>**DeviceFaultInfo.deviceType**：设备类型名，包括“Node”、“NPU”、“Storage”、“CPU”、“Network”等</p><p>**DeviceFaultInfo.faultCodes**：故障码列表</p><p>**DeviceFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，严重级别依次递增</p><p>**DeviceFaultInfo.faultType**：故障子系统类型，预留字段</p><p>**DeviceFaultInfo.faultReason**：故障原因，预留字段</p><p>**DeviceFaultInfo.switchFaultInfos**：灵衢故障信息</p><p>**DeviceFaultInfo.faultLevels**：故障等级列表</p><p>**SwitchFaultInfo.faultCode**：灵衢故障码</p><p>**SwitchFaultInfo.switchChipId**：灵衢故障芯片ID</p><p>**SwitchFaultInfo.switchPortId**：灵衢故障端口ID</p><p>**SwitchFaultInfo.faultTime**：灵衢故障发生时间</p><p>**SwitchFaultInfo.faultLevel**：灵衢故障等级</p>|



### GetFaultMsgSignal<a name="ZH-CN_TOPIC_0000002479226874"></a>

**功能说明<a name="section143314311911"></a>**

本接口为故障查询接口。功能主要是接收客户端查询集群、任务故障信息的请求。

>[!NOTE] 说明 
>该接口每秒最多可查询10次，超过10次时会将请求加入等待队列中。总等待数超过50时，再次发送请求会被拒绝。

**函数原型<a name="section3958124212115"></a>**

```
rpc GetFaultMsgSignal(ClientInfo) returns(FaultQueryResult){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。当jobId传入空值时返回集群范围内的故障信息。若jobId不传入空值，则jobId的合理长度为[8,128]个字符，且不能包含汉字字符。</p><p>**ClientInfo.role**：客户端角色。</p><span>说明：</span><ul><li>传入jobId为空时，查询的结果为当前集群的全量故障。</li><li>传入jobId不为空时，查询的结果为任务所属节点的故障。</li></ul>|


**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|FaultQueryResult|<p>message FaultQueryResult{</p><p>int32 code = 1;</p><p>string info = 2;</p><p>FaultMsgSignal faultSignal =3;</p>}|<p>**code**：本次查询的返回码。<ul><li>200：查询正常返回。</li><li>429：服务端限流。</li><li>500：服务端错误。</li></ul></p><p>**info**：本次查询结果的描述信息</p><p>**faultSignal**：故障信息结构体</p><p>**FaultMsgSignal.uuid**：消息id</p><p>**FaultMsgSignal.jobId**：任务id，-1代表集群</p><p>**FaultMsgSignal.signalType**：消息类型，“fault”代表故障发生，“normal”代表无故障或故障恢复</p><p>**FaultMsgSignal.nodeFaultInfo**：节点故障信息</p><p>**NodeFaultInfo.nodeName**：故障节点名称</p><p>**NodeFaultInfo.nodeIP**：节点IP</p><p>**NodeFaultInfo.nodeSN**：节点SN号</p><p>**NodeFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，设置为DeviceFaultInfo.faultLevel中最严重的级别</p><p>**NodeFaultInfo.faultDevice**：设备故障信息</p><p>**DeviceFaultInfo.deviceId**：设备ID</p><p>**DeviceFaultInfo.deviceType**：设备类型名，包括“Node”、“NPU”、“Storage”、“CPU”、“Network”等</p><p>**DeviceFaultInfo.faultCodes**：故障码列表</p><p>**DeviceFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，严重级别依次递增</p><p>**DeviceFaultInfo.faultType**：故障子系统类型，预留字段</p><p>**DeviceFaultInfo.faultReason**：故障原因，预留字段</p><p>**DeviceFaultInfo.switchFaultInfos**：灵衢故障信息列表</p><p>**DeviceFaultInfo.faultLevels**：故障等级列表</p><p>**SwitchFaultInfo.faultCode**：灵衢故障码</p><p>**SwitchFaultInfo.switchChipId**：灵衢故障芯片ID</p><p>**SwitchFaultInfo.switchPortId**：灵衢故障端口ID</p><p>**SwitchFaultInfo.faultTime**：灵衢故障发生时间</p><p>**SwitchFaultInfo.faultLevel**：灵衢故障等级</p>|




## 任务信息接口<a name="ZH-CN_TOPIC_0000002511426731"></a>

### 接口说明<a name="ZH-CN_TOPIC_0000002479386816"></a>

本模块向外部提供订阅获取集群任务状态及基本信息的订阅接口。分为3个接口即：Register、SubscribeJobSummarySignal和SubscribeJobSummarySignalList。

**调用顺序说明<a name="section171351329174616"></a>**

用户在调用SubscribeJobSummarySignal和SubscribeJobSummarySignalList接口前需首先通过Register接口获取合法的客户端ID后，持有该ID调用订阅接口。

订阅接口默认两分钟内无活动信息主动关闭。


### Register<a name="ZH-CN_TOPIC_0000002479226804"></a>

**功能说明<a name="section14645125754213"></a>**

在订阅任务信息的场景下，接收客户端的注册请求。

客户端如需订阅集群任务信息，需先调用本接口获取返回的UUID后，持有该ID调用订阅接口SubscribeJobSummarySignal和SubscribeJobSummarySignalList获取集群的任务信息。

>[!NOTE] 说明 
>集群内最多存在80个订阅方，每种客户端角色最多支持创建20个订阅方。

**函数原型<a name="section4140960433"></a>**

```
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



### SubscribeJobSummarySignal<a name="ZH-CN_TOPIC_0000002511426723"></a>

**功能说明<a name="section85381247165120"></a>**

接收客户端的任务信息变更订阅。客户端初次订阅接口时，会逐条推送当前集群中所有任务的信息。当任务状态改变时，向注册的客户端广播推送。当连接两分钟内无消息且无心跳时，服务端主动断开该连接，并释放订阅。

>[!NOTE] 说明 
>本接口具有限流机制，1秒内允许访问的最大次数为20。

**函数原型<a name="section1199205575113"></a>**

```
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
|JobSummarySignal|<p>message JobSummarySignal{</p><p>string uuid = 1;</p><p>string jobId = 2;</p><p>string jobName = 3;</p><p>string namespace =4;</p><p>string frameWork = 5;</p><p>string jobStatus = 6;</p><p>string time = 7;</p><p>string cmIndex = 8;</p><p>string total = 9;</p><p>string HcclJson = 10;</p><p>string deleteTime = 11;</p><p>string sharedTorIp = 12;</p><p>string masterAddr = 13;</p><p>string operator = 14;</p>}|<p>**uuid**：本条消息id</p><p>**jobId**：任务的K8s ID信息</p><p>**jobName**：当前任务的名称</p><p>**namespace**：任务所属命名空间</p><p>**frameWork**：任务框架</p>**jobStatus**：任务状态，存在以下几种状态：<ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul><p>**time**：任务开始时间</p><p>**cmIndex**：序号</p><p>**total**：任务对应的jobsummary ConfigMap的数量总数</p><p>**HcclJson**：任务使用的芯片通信信息。若任务调度的NPU数量超过4万，客户端接收的上报信息中HcclJson会被设置为空。<p>可转义为JSON格式，字段说明如下：</p><ul><li>status：任务RankTable是否已经生成</li><li>initializing：还在为任务分配设备，RankTable未生成</li><li>complete：当RankTable生成后，状态会立即变为complete，同步出现server_list等其他字段</li><li>server_list：任务设备分配情况</li><li>device：记录NPU分配，NPU IP和rank_id信息</li><li>server_id：AI Server标识，全局唯一</li><li>server_name：节点名称</li><li>server_sn：节点的SN号。需要保证设备的SN存在。若不存在，请联系华为技术支持</li><li>server_count：任务使用的节点数量</li><li>version：版本信息</li></ul></p><p>**deleteTime**：任务被删除的时间</p><p>**sharedTorIp**：任务使用的共享交换机信息</p><p>**masterAddr**：PyTorch训练时指定的MASTER_ADDR值</p><p>**operator**：接收到添加任务命令后状态更新为add；接收到删除任务命令后状态更新为delete</p>|

### SubscribeJobSummarySignalList

**功能说明**

接收客户端的任务信息变更订阅。客户端初次订阅接口时，推送当前集群中所有任务的信息。当任务状态改变时，向注册的客户端广播推送。当连接两分钟内无消息且无心跳时，服务端主动断开该连接，并释放订阅。

>[!NOTE] 说明 
>本接口具有限流机制，1秒内允许访问的最大次数为20。

**函数原型**

```
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
|JobSummarySignalList|<p>message JobSummarySignalList{</p><p>repeated JobSummarySignal jobSummarySignals = 1;</p><p>string ReportTime = 2;</p><p>int32 JobTotalNum = 3;</p>}</p><p>message JobSummarySignal{<p>string uuid = 1;</p><p>string jobId = 2;</p><p>string jobName = 3;</p><p>string namespace =4;</p><p>string frameWork = 5;</p><p>string jobStatus = 6;</p><p>string time = 7;</p><p>string cmIndex = 8;</p><p>string total = 9;</p><p>string HcclJson = 10;</p><p>string deleteTime = 11;</p><p>string sharedTorIp = 12;</p><p>string masterAddr = 13;</p><p>string operator = 14;</p>}</p>|<p>**jobSummarySignals**: 任务信息列表</p><p>**ReportTime**：当前批次上报的时间</p><p>**JobTotalNum**：相同批次上报的任务总数</p><p>**uuid**：本条消息id</p><p>**jobId**：任务的K8s ID信息</p><p>**jobName**：当前任务的名称</p><p>**namespace**：任务所属命名空间</p><p>**frameWork**：任务框架</p>**jobStatus**：任务状态，存在以下几种状态：<ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul><p>**time**：任务开始时间</p><p>**cmIndex**：序号</p><p>**total**：任务对应的jobsummary ConfigMap的数量总数</p><p>**HcclJson**：任务使用的芯片通信信息。可转义为JSON格式，字段说明如下：<ul><li>status：任务RankTable是否已经生成</li><li>initializing：还在为任务分配设备，RankTable未生成</li><li>complete：当RankTable生成后，状态会立即变为complete，同步出现server_list等其他字段</li><li>server_list：任务设备分配情况</li><li>device：记录NPU分配，NPU IP和rank_id信息</li><li>server_id：AI Server标识，全局唯一</li><li>server_name：节点名称</li><li>server_sn：节点的SN号。需要保证设备的SN存在。若不存在，请联系华为技术支持</li><li>server_count：任务使用的节点数量</li><li>version：版本信息</li></ul></p><p>说明：</p><ul><li>若单个任务所使用的NPU数量超过4万，上报的任务信息中HcclJson会被设置为空。</li><li>客户端初次订阅接口时，若多个任务合计使用的NPU数量超过4万，会对上报信息进行分页上报，确保每条上报信息中任务总NPU数不超过4万。</li></ul><p>**deleteTime**：任务被删除的时间</p><p>**sharedTorIp**：任务使用的共享交换机信息</p><p>**masterAddr**：PyTorch训练时指定的MASTER_ADDR值</p><p>**operator**：接收到添加任务命令后状态更新为add；接收到删除任务命令后状态更新为delete</p>|



## 借轨回切接口<a name="ZH-CN_TOPIC_0000002511426725"></a>

### SwitchNicTrack<a name="ZH-CN_TOPIC_0000002511346727"></a>

**功能说明<a name="section143314311911"></a>**

接收运维平台的借轨请求，将训练任务的指定节点的Device下发借轨/回切操作，该接口需要等待训练任务已经成功运行，出迭代以后再调用，保证任务已经注册到ClusterD。借轨/回切接口属于人工运维操作，对于反复切换场景，若每次切换都失败，会导致频繁保存CKPT，存在磁盘爆盘的风险。

>[!NOTE] 说明 
>请在训练正常迭代后，再进行借轨或回切指令的下发。

**函数原型<a name="section3958124212115"></a>**

```
rpc SwitchNicTrack(SwitchNics) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SwitchNics|<p>message SwitchNics{</p><p>string jobID;</p><p>map<string, DeviceList> nicOps;</p>}</p><p>message DeviceList {<p>repeated string dev;</p><p>repeated bool op;</p>}</p>|<p>**SwitchNics.jobID**：任务ID。</p><p>**SwitchNics.nicOps**：用户下发借轨/回切指令的设备与操作。key为node name，value为该节点要操作的Device。</p><p>**DeviceList.dev**：该节点上的DeviceID列表，与DeviceList.op数量保持一致。</p><p>**DeviceList.op**：该节点的DeviceID对应设备要执行的借轨操作列表。true表示切换到备用链路，false表示使用主链路。|


**返回值说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示下发指令成功。</li><li>其他值：表示下发失败。</li></ul>**Status.info**：返回信息描述。|



### SubscribeSwitchNicSignal<a name="ZH-CN_TOPIC_0000002479226844"></a>

**功能说明<a name="section143314311911"></a>**

运维平台查询借轨/回切结果的接口。当运维人员下发主动借轨/回切指令成功后，可通过该接口查询结果。

**函数原型<a name="section3958124212115"></a>**

```
rpc SubscribeSwitchNicSignal(SwitchNicRequest) returns (stream SwitchNicResponse) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SwitchNicRequest|<p>message SwitchNicRequest{</p><p>string jobID;</p>}|**SwitchNicRequest.jobID**：任务ID|


**返回值说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SwitchNicResponse|message SwitchNicResponse{<p>string jobID;</p><p>string msg;</p>}|<p>**SwitchNicResponse.jobID**：任务ID</p><p>**SwitchNicResponse.msg**：借轨/回切指令的执行结果</p>|



### SubscribeNotifySwitch<a name="ZH-CN_TOPIC_0000002511346769"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅借轨/回切信号请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```
rpc SubscribeNotifySwitch(ClientInfo) returns (stream SwitchRankList) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p>|


**发送数据说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SwitchRankList|<p>message SwitchRankList{</p><p>repeated string rankID = 1;</p><p>repeated bool op = 2;</p><p>string jobId = 3;</p>}|<p>**SwitchRankList.rankID**：该节点上的DeviceID列表，与DeviceList.op数量保持一致。</p><p>**SwitchRankList.op**：该节点的DeviceID对应设备要执行的借轨操作列表。true表示切换到备用链路，false表示使用主链路。</p><p>**SwitchRankList.jobId**：任务ID</p>|


**返回值说明<a name="section69806312314"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|



### ReplySwitchNicResult<a name="ZH-CN_TOPIC_0000002479386790"></a>

**功能说明<a name="section143314311911"></a>**

客户端向ClusterD返回借轨/回切结果的接口。

**函数原型<a name="section3958124212115"></a>**

```
rpc ReplySwitchNicResult(SwitchResult) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SwitchResult|message SwitchResult{<p>string jobId = 1;</p><p>bool result = 2;</p>}|<p>**SwitchResult.jobId**：任务ID。</p><p>**SwitchResult.result**：指令执行的结果，true为成功，false为失败。</p>|


**返回值说明<a name="section69806312314"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示流程正常</li><li>其他值：表示流程异常</li></ul>**Status.info**：返回信息描述。|




## 在线压测接口<a name="ZH-CN_TOPIC_0000002479226858"></a>

### StressTest<a name="ZH-CN_TOPIC_0000002511426729"></a>

**功能说明<a name="section143314311911"></a>**

接收运维平台的在线压测请求，将指定训练任务的指定节点下发压测操作，该接口需要等待训练任务已经成功运行，出迭代以后再调用，保证任务已经注册到ClusterD。在线压测接口属于人工运维操作，调用接口前请先确保服务器环境正常。

>[!NOTE] 说明 
>请在训练正常迭代后，再进行在线压测指令的下发。

**函数原型<a name="section3958124212115"></a>**

```
rpc StressTest(StressTestParam) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTest|<p>message StressTestParam {</p><p>string jobID = 1;</p><p>map<string, StressOpList> stressParam = 2;</p><p>repeated int64 allNodesOps = 3;</p>}</p><p>message StressOpList {<p>repeated int64 ops = 1;</p>}</p>|<p>**StressTestParam.jobID**：任务ID。</p><p>**StressTestParam.stressParam**：用户下发压测指令的节点与操作。key为node name，value为该节点要执行的压测操作。</p><p>**StressTestParam.allNodesOps**：若用户要对任务的所有节点进行压测，则该字段表示所有节点要执行的压测操作。allNodesOps字段优先级高于stressParam。其中，0表示“aic”压测；1表示“p2p”压测。</p><p>**StressOpList.ops**：该节点要执行的压测操作。0表示“aic”压测；1表示“p2p”压测。</p>|


**返回值说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示下发指令成功。</li><li>其他值：表示下发失败。</li></ul>**Status.info**：返回信息描述。|



### SubscribeStressTestResponse<a name="ZH-CN_TOPIC_0000002511346789"></a>

**功能说明<a name="section143314311911"></a>**

运维平台查询压测结果的接口。当运维人员下发在线压测指令成功后，可通过该接口查询结果。

**函数原型<a name="section3958124212115"></a>**

```
rpc SubscribeStressTestResponse(StressTestRequest) returns (stream StressTestResponse) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTestRequest|message StressTestRequest{<p>string jobID = 1;</p>}|**StressTestRequest.jobID**：任务ID。|


**返回值说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTestResponse|<p>message StressTestResponse {</p><p>string jobID;</p><p>string msg;</p>}|<p>**StressTestResponse.jobID**：任务ID。</p><p>**StressTestResponse.msg**：压测的执行结果。</p>|



### SubscribeNotifyExecStressTest<a name="ZH-CN_TOPIC_0000002479386800"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅在线压测信号请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```
rpc SubscribeNotifyExecStressTest(ClientInfo) returns (stream StressTestRankParams) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。</p>|


**发送数据说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTestRankParams|<p>message StressTestRankParams {</p><p>map<string, StressOpList> stressParam = 1;</p><p>string jobId = 2;</p>}|<p>**StressTestRankParams.stressParam**：key为该节点上要执行压测的global RankID，value为对应的压测操作，0表示“aic”压测；1表示“p2p”压测。</p><p>**StressTestRankParams.jobId**：任务ID。</p>|


**返回值说明<a name="section69806312314"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|



### ReplyStressTestResult<a name="ZH-CN_TOPIC_0000002511346775"></a>

**功能说明<a name="section143314311911"></a>**

客户端向ClusterD返回在线压测结果的接口。

**函数原型<a name="section3958124212115"></a>**

```
rpc ReplyStressTestResult(StressTestResult) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTestResult|<p>message StressTestResult {</p><p>string jobId = 1;</p><p>map<string, StressTestRankResult> stressResult = 2;</p>}</p><p>message StressTestRankResult {<p>map<string, StressTestOpResult> rankResult= 1;</p>}</p><p>message StressTestOpResult {<p>string code = 1;</p><p>string result = 2;</p>}</p>|<p>**StressTestResult.jobId**：任务ID。</p><p>**StressTestResult.stressResult**：指令执行的结果。key为执行压测的global rankID；value为执行压测的结果。</p><p>**StressTestRankResult.rankResult**：某张卡执行压测的结果。key为压测的操作，0表示“aic”压测；1表示“p2p”压测。value为对应的结果。</p><p>**StressTestOpResult.code**：压测结果的错误码。<ul><li>0表示执行成功，无故障</li><li>1表示压测失败，可正常恢复训练</li><li>2表示发现压测故障，需要隔离对应节点</li><li>3表示压测超时，该节点任务退出重启</li><li>4表示压测电压未恢复，该节点任务退出重启</li></ul></p><p>**StressTestOpResult.result**：压测结果的描述信息。</p>|


**返回值说明<a name="section69806312314"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示流程正常</li><li>其他值：表示流程异常</li></ul>**Status.info**：返回信息描述。|




