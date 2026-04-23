# 故障服务接口<a name="ZH-CN_TOPIC_0000002479386826"></a>

## Register<a name="ZH-CN_TOPIC_0000002511426773"></a>

**功能说明<a name="section143314311911"></a>**

接收处理客户端的注册请求，为订阅故障信息等功能做初始化准备。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc Register(ClientInfo) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>传入jobId为空时，表示注册集群所有任务。</li><li>传入jobId不为空时，表示注册指定任务。</li></ul></div></div></p>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示注册成功。</li><li>其他值：表示注册失败。</li></ul>**Status.info**：返回信息描述。|

## SubscribeFaultMsgSignal<a name="ZH-CN_TOPIC_0000002511426699"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅故障信息请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

>[!NOTE] 
>
>- 调用此接口前，需先调用[Register接口](#ZH-CN_TOPIC_0000002511426773)。
>- 客户端订阅通算任务的故障信息后，只能收到NodeD故障和K8s节点状态异常故障。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc SubscribeFaultMsgSignal(ClientInfo) returns (stream FaultMsgSignal){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|<p>message ClientInfo{</p><p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。</p><p>**ClientInfo.role**：客户端角色。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>传入jobId为空时，获取的结果为集群内所有job的故障。</li><li>传入jobId不为空时，获取的结果为任务所属节点的故障。</li></ul></div></div>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream|grpc stream|<ul><li>该接口返回gRPC stream（返回值的具体数据结构基于客户端选择的编程语言）。</li><li>客户端可以调用stream的Receive方法（具体方法名基于客户端选择的编程语言）接收服务端推送的数据。</li></ul>|

**发送数据说明<a name="section112224012419"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|FaultMsgSignal|<p>message FaultMsgSignal{</p><p>string uuid = 1;</p><p>string jobId = 2;</p><p>string signalType = 3;</p><p>repeated NodeFaultInfo nodeFaultInfo = 4;</p>}</p><p>message NodeFaultInfo{<p>string nodeName = 1;</p><p>string nodeIP = 2;</p><p>string nodeSN = 3;</p><p>string faultLevel = 4;</p><p>repeated DeviceFaultInfo faultDevice = 5;</p>}</p><p>message DeviceFaultInfo{<p>string deviceId = 1;</p><p>string deviceType = 2;</p><p>repeated string faultCodes = 3;</p><p>string faultLevel = 4;</p><p>repeated string faultType = 5;</p><p>repeated string faultReason = 6;</p><p>repeated SwitchFaultInfo switchFaultInfos = 7;</p><p>repeated string faultLevels = 8;</p>}</p><p>message SwitchFaultInfo{<p>string faultCode = 1;</p><p>string switchChipId = 2;</p><p>string switchPortId = 3;</p><p>string faultTime = 4;</p><p>string faultLevel = 5;</p>}</p>|<p>**FaultMsgSignal.uuid**：消息ID</p><p>**FaultMsgSignal.jobId**：任务ID</p><p>**FaultMsgSignal.signalType**：消息类型，“fault”代表故障发生，“normal”代表无故障或故障恢复</p><p>**FaultMsgSignal.nodeFaultInfo**：节点故障信息</p><p>**NodeFaultInfo.nodeName**：故障节点名称</p><p>**NodeFaultInfo.nodeIP**：节点IP</p><p>**NodeFaultInfo.nodeSN**：节点SN号</p><p>**NodeFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，设置为DeviceFaultInfo.faultLevel中最严重的级别</p><p>**NodeFaultInfo.faultDevice**：设备故障信息</p><p>**DeviceFaultInfo.deviceId**：设备ID。当节点发生总线设备故障和K8s状态异常故障时，deviceId为-1</p><p>**DeviceFaultInfo.deviceType**：设备类型名，包括“Node”、“NPU”、“Storage”、“CPU”、“Network”等</p><p>**DeviceFaultInfo.faultCodes**：故障码列表</p><p>**DeviceFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，严重级别依次递增</p><p>**DeviceFaultInfo.faultType**：故障子系统类型，预留字段</p><p>**DeviceFaultInfo.faultReason**：故障原因，预留字段</p><p>**DeviceFaultInfo.switchFaultInfos**：灵衢故障信息</p><p>**DeviceFaultInfo.faultLevels**：故障等级列表</p><p>**SwitchFaultInfo.faultCode**：灵衢故障码</p><p>**SwitchFaultInfo.switchChipId**：灵衢故障芯片ID</p><p>**SwitchFaultInfo.switchPortId**：灵衢故障端口ID</p><p>**SwitchFaultInfo.faultTime**：灵衢故障发生时间</p><p>**SwitchFaultInfo.faultLevel**：灵衢故障等级</p>|

## GetFaultMsgSignal<a name="ZH-CN_TOPIC_0000002479226874"></a>

**功能说明<a name="section143314311911"></a>**

本接口为故障查询接口。功能主要是接收客户端查询集群、任务故障信息的请求。

>[!NOTE] 
>该接口每秒最多可查询10次，超过10次时会将请求加入等待队列中。总等待数超过50时，再次发送请求会被拒绝。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc GetFaultMsgSignal(ClientInfo) returns (FaultQueryResult){}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ClientInfo|message ClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**ClientInfo.jobId**：任务ID。当jobId传入空值时返回集群范围内的故障信息。若jobId不传入空值，则jobId的合理长度为[8,128]个字符，且不能包含汉字字符。</p><p>**ClientInfo.role**：客户端角色。</p><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>传入jobId为空时，查询的结果为当前集群的全量故障。</li><li>传入jobId不为空时，查询的结果为任务所属节点的故障。</li></ul></div></div>|

**返回值说明<a name="section206103328174"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|FaultQueryResult|<p>message FaultQueryResult{</p><p>int32 code = 1;</p><p>string info = 2;</p><p>FaultMsgSignal faultSignal =3;</p>}|<p>**code**：本次查询的返回码。<ul><li>200：查询正常返回。</li><li>429：服务端限流。</li><li>500：服务端错误。</li></ul></p><p>**info**：本次查询结果的描述信息</p><p>**faultSignal**：故障信息结构体</p><p>**FaultMsgSignal.uuid**：消息id</p><p>**FaultMsgSignal.jobId**：任务id，-1代表集群</p><p>**FaultMsgSignal.signalType**：消息类型，“fault”代表故障发生，“normal”代表无故障或故障恢复</p><p>**FaultMsgSignal.nodeFaultInfo**：节点故障信息</p><p>**NodeFaultInfo.nodeName**：故障节点名称</p><p>**NodeFaultInfo.nodeIP**：节点IP</p><p>**NodeFaultInfo.nodeSN**：节点SN号</p><p>**NodeFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，设置为DeviceFaultInfo.faultLevel中最严重的级别</p><p>**NodeFaultInfo.faultDevice**：设备故障信息</p><p>**DeviceFaultInfo.deviceId**：设备ID</p><p>**DeviceFaultInfo.deviceType**：设备类型名，包括“Node”、“NPU”、“Storage”、“CPU”、“Network”等</p><p>**DeviceFaultInfo.faultCodes**：故障码列表</p><p>**DeviceFaultInfo.faultLevel**：故障类型，包括“Healthy”、“SubHealthy”和“UnHealthy”，严重级别依次递增</p><p>**DeviceFaultInfo.faultType**：故障子系统类型，预留字段</p><p>**DeviceFaultInfo.faultReason**：故障原因，预留字段</p><p>**DeviceFaultInfo.switchFaultInfos**：灵衢故障信息列表</p><p>**DeviceFaultInfo.faultLevels**：故障等级列表</p><p>**SwitchFaultInfo.faultCode**：灵衢故障码</p><p>**SwitchFaultInfo.switchChipId**：灵衢故障芯片ID</p><p>**SwitchFaultInfo.switchPortId**：灵衢故障端口ID</p><p>**SwitchFaultInfo.faultTime**：灵衢故障发生时间</p><p>**SwitchFaultInfo.faultLevel**：灵衢故障等级</p>|
