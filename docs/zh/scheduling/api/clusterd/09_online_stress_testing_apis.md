# 在线压测接口<a name="ZH-CN_TOPIC_0000002479226858"></a>

## StressTest<a name="ZH-CN_TOPIC_0000002511426729"></a>

**功能说明<a name="section143314311911"></a>**

接收运维平台的在线压测请求，将指定训练任务的指定节点下发压测操作，该接口需要等待训练任务已经成功运行，出迭代以后再调用，保证任务已经注册到ClusterD。在线压测接口属于人工运维操作，调用接口前请先确保服务器环境正常。

>[!NOTE] 
>请在训练正常迭代后，再进行在线压测指令的下发。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc StressTest(StressTestParam) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTest|<p>message StressTestParam {</p><p>string jobID = 1;</p><p>map<string, StressOpList> stressParam = 2;</p><p>repeated int64 allNodesOps = 3;</p>}<p>message StressOpList {<p>repeated int64 ops = 1;</p>}</p>|<p>**StressTestParam.jobID**：任务ID。</p><p>**StressTestParam.stressParam**：用户下发压测指令的节点与操作。key为node name，value为该节点要执行的压测操作。</p><p>**StressTestParam.allNodesOps**：若用户要对任务的所有节点进行压测，则该字段表示所有节点要执行的压测操作。allNodesOps字段优先级高于stressParam。其中，0表示“aic”压测；1表示“p2p”压测。</p><p>**StressOpList.ops**：该节点要执行的压测操作。0表示“aic”压测；1表示“p2p”压测。</p>|

**返回值说明<a name="section146221236193515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info = 2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示下发指令成功。</li><li>其他值：表示下发失败。</li></ul>**Status.info**：返回信息描述。|

## SubscribeStressTestResponse<a name="ZH-CN_TOPIC_0000002511346789"></a>

**功能说明<a name="section143314311911"></a>**

运维平台查询压测结果的接口。当运维人员下发在线压测指令成功后，可通过该接口查询结果。

**函数原型<a name="section3958124212115"></a>**

```proto
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

## SubscribeNotifyExecStressTest<a name="ZH-CN_TOPIC_0000002479386800"></a>

**功能说明<a name="section143314311911"></a>**

接收客户端订阅在线压测信号请求。服务端为每一个任务分配一个消息队列，并侦听消息队列是否存在待发送的消息，若存在则通过gRPC stream发送给客户端。

**函数原型<a name="section3958124212115"></a>**

```proto
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

## ReplyStressTestResult<a name="ZH-CN_TOPIC_0000002511346775"></a>

**功能说明<a name="section143314311911"></a>**

客户端向ClusterD返回在线压测结果的接口。

**函数原型<a name="section3958124212115"></a>**

```proto
rpc ReplyStressTestResult(StressTestResult) returns (Status) {}
```

**输入参数说明<a name="section14344145451114"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|StressTestResult|<p>message StressTestResult {</p><p>string jobId = 1;</p><p>map<string, StressTestRankResult> stressResult = 2;</p>}<p>message StressTestRankResult {<p>map<string, StressTestOpResult> rankResult= 1;</p>}</p><p>message StressTestOpResult {<p>string code = 1;</p><p>string result = 2;</p>}</p>|<p>**StressTestResult.jobId**：任务ID。</p><p>**StressTestResult.stressResult**：指令执行的结果。key为执行压测的global rankID；value为执行压测的结果。</p><p>**StressTestRankResult.rankResult**：某张卡执行压测的结果。key为压测的操作，0表示“aic”压测；1表示“p2p”压测。value为对应的结果。</p><p>**StressTestOpResult.code**：压测结果的错误码。<ul><li>0表示执行成功，无故障</li><li>1表示压测失败，可正常恢复训练</li><li>2表示发现压测故障，需要隔离对应节点</li><li>3表示压测超时，该节点任务退出重启</li><li>4表示压测电压未恢复，该节点任务退出重启</li></ul></p><p>**StressTestOpResult.result**：压测结果的描述信息。</p>|

**返回值说明<a name="section69806312314"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Status|<p>message Status{</p><p>int32 code = 1;</p><p>string info =2;</p>}|**Status.code**：返回码。<ul><li>取值为0：表示流程正常</li><li>其他值：表示流程异常</li></ul>**Status.info**：返回信息描述。|
