# 进程级恢复接口<a name="ZH-CN_TOPIC_0000002511346765"></a>

## gRPC接口<a name="ZH-CN_TOPIC_0000002511346735"></a>

### Register（公共接口）<a name="ZH-CN_TOPIC_0000002511346739"></a>

**功能说明<a name="section143314311911"></a>**

接收处理客户端的注册请求，为进程级恢复做初始化准备。作业调用Init成功后，需要等待客户端确认MindIO侧进程级别重调度和进程级在线恢复开关为打开状态后，调用此接口。Register成功后进程级别重调度和进程级在线恢复功能才会处于可用状态。

**函数原型<a name="section3958124212115"></a>**

```proto
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

### Init<a name="ZH-CN_TOPIC_0000002479386824"></a>

**功能说明<a name="section83882209338"></a>**

用于初始化进程级别重调度和进程级在线恢复，初始化成功后，进程级别重调度和进程级在线恢复功能将暂时处于不可用状态。

**函数原型<a name="section2049633816332"></a>**

```proto
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

### SubscribeProcessManageSignal<a name="ZH-CN_TOPIC_0000002511426713"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅进程控制信号请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```proto
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

### ReportStopComplete<a name="ZH-CN_TOPIC_0000002511426707"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端上报暂停训练进程是否成功。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc ReportStopComplete(StopCompleteRequest) returns (Status){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StopCompleteRequest|message StopCompleteRequest{<p>string jobId = 1;</p><p>Status status = 2;</p><p>repeated FaultRank faultRankIds = 3;</p>}|<p>**StopCompleteRequest.jobId**：任务ID。</p><p>**StopCompleteRequest.status.code**：返回码，OK表示暂停训练成功，其他值表示暂停训练失败。</p><p>**StopCompleteRequest.status.info**：返回信息描述。</p><p>**StopCompleteRequest.faultRankIds**：故障芯片全局故障Rank列表。FaultRank是一组包含故障信息的键值对，由rankId（全局Rank ID）和faultType（故障类型）组成。faultType取值为0时，表示片上内存故障；取值为1时，表示其他故障；取值为2时，表示网络故障。</p>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{int32 code = 1;string info = 2;}|<p>**Status.code**：返回码。<ul><li>取值为0：表示故障恢复流程正常</li><li>其他值：表示故障恢复流程异常，并触发重调度。</li></ul></p><p>**Status.info**：返回信息描述。</p>|

### ReportRecoverStrategy<a name="ZH-CN_TOPIC_0000002511346747"></a>

**功能说明<a name="section16150748174520"></a>**

接收客户端上报的当前任务支持的故障恢复策略。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc ReportRecoverStrategy(RecoverStrategyRequest) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|RecoverStrategyRequest|message RecoverStrategyRequest{<p>string jobId = 1;</p><p>repeated FaultRank faultRankIds = 2;</p><p>repeated string strategies = 3;</p>}|<p>**RecoverStrategyRequest.jobId**：任务ID</p><p>**RecoverStrategyRequest.faultRankIds**：故障芯片全局故障Rank列表。FaultRank是故障信息的键值对，包含rankId（全局Rank ID）和faultType（故障类型）。faultType取值为0时，表示片上内存故障；取值为1时，表示其他故障；取值为2时，表示网络故障。</p><p>**RecoverStrategyRequest.strategies**：当前任务支持的恢复策略。</p>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<ul><li>0：表示故障恢复流程正常。</li><li>其他值：表示恢复流程异常，并触发重调度。</li></ul></p><p>**Status.info**：返回信息描述。</p>|

### ReportRecoverStatus<a name="ZH-CN_TOPIC_0000002511346753"></a>

**功能说明<a name="section16150748174520"></a>**

接收客户端上报的当前任务恢复情况。

**函数原型<a name="section3958124212115"></a>**

```proto
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

### ReportProcessFault<a name="ZH-CN_TOPIC_0000002511346729"></a>

**功能说明<a name="section16150748174520"></a>**

接收客户端上报的故障芯片全局Rank信息。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc ReportProcessFault(ProcessFaultRequest) returns (Status){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ProcessFaultRequest|message ProcessFaultRequest{<p>string jobId = 1;</p><p>repeated FaultRank faultRankIds = 2;</p>}|<p>**ProcessFaultRequest.jobId**：任务ID。</p><p>**ProcessFaultRequest.faultRankIds**：故障芯片全局Rank ID列表。FaultRank是故障信息的键值对，包含rankId（全局Rank ID）和faultType（故障类型）。faultType取值为0时，表示片上内存故障；取值为1时，表示其他故障；取值为2时，表示网络故障。</p>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|message Status{<p>int32 code = 1;</p><p>string info = 2;</p>}|<p>**Status.code**：返回码。<li>0：表示恢复流程正常。</li><li>其他值：表示故障恢复流程异常，并触发重调度。</li></p><p>**Status.info**：返回信息描述。</p>|

### HealthCheck<a name="ZH-CN_TOPIC_0000002511426765"></a>

**功能说明<a name="section16150748174520"></a>**

检查gRPC链接状态。

**函数原型<a name="section3958124212115"></a>**

```proto
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

## 对接第三方AI平台控制相关接口<a name="ZH-CN_TOPIC_0000002511346803"></a>

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
