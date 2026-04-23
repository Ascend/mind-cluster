# 借轨回切接口<a name="ZH-CN_TOPIC_0000002511426725"></a>

## SwitchNicTrack<a name="ZH-CN_TOPIC_0000002511346727"></a>

**功能说明<a name="section143314311911"></a>**

接收运维平台的借轨请求，将训练任务的指定节点的Device下发借轨/回切操作，该接口需要等待训练任务已经成功运行，出迭代以后再调用，保证任务已经注册到ClusterD。借轨/回切接口属于人工运维操作，对于反复切换场景，若每次切换都失败，会导致频繁保存CKPT，存在磁盘爆盘的风险。

>[!NOTE] 
>请在训练正常迭代后，再进行借轨或回切指令的下发。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc SwitchNicTrack(SwitchNics) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|SwitchNics|<p>message SwitchNics{</p><p>string jobID;</p><p>map<string, DeviceList> nicOps;</p>}<p>message DeviceList {<p>repeated string dev;</p><p>repeated bool op;</p>}</p>|<p>**SwitchNics.jobID**：任务ID。</p><p>**SwitchNics.nicOps**：用户下发借轨/回切指令的设备与操作。key为node name，value为该节点要操作的Device。</p><p>**DeviceList.dev**：该节点上的DeviceID列表，与DeviceList.op数量保持一致。</p><p>**DeviceList.op**：该节点的DeviceID对应设备要执行的借轨操作列表。true表示切换到备用链路，false表示使用主链路。</p>|

**返回值说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示下发指令成功。</li><li>其他值：表示下发失败。</li></ul>**Status.info**：返回信息描述。|

## SubscribeSwitchNicSignal<a name="ZH-CN_TOPIC_0000002479226844"></a>

**功能说明<a name="section143314311911"></a>**

运维平台查询借轨/回切结果的接口。当运维人员下发主动借轨/回切指令成功后，可通过该接口查询结果。

**函数原型<a name="section3958124212115"></a>**

```proto
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

## SubscribeNotifySwitch<a name="ZH-CN_TOPIC_0000002511346769"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅借轨/回切信号请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```proto
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

## ReplySwitchNicResult<a name="ZH-CN_TOPIC_0000002479386790"></a>

**功能说明<a name="section143314311911"></a>**

客户端向ClusterD返回借轨/回切结果的接口。

**函数原型<a name="section3958124212115"></a>**

```proto
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
