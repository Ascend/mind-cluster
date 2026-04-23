# 性能劣化故障接口<a name="ZH-CN_TOPIC_0000002479226802"></a>

## ModifyTrainingDataTraceSwitch<a name="ZH-CN_TOPIC_0000002511426771"></a>

**功能说明<a name="section22878209356"></a>**

外部调用修改各类数据动态打点开关能力。

>[!NOTE] 
>如果通过ClusterD提供的gRPC接口这种方式开启或修改轻量profiling获取落盘数据，创建的data-trace-<任务名称\> ConfigMap的生命周期会随着任务的删除而删除。当任务不存在的时候，该接口会调用失败。

**函数原型<a name="section1472624833519"></a>**

```proto
rpc ModifyTrainingDataTraceSwitch(DataTypeReq) returns (DataTypeRes) {}
```

**输入参数说明<a name="section6782115723515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataTypeReq|<p>message DataTypeReq{<p>string jobNsName = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p>}</p><p>message ProfilingSwitch{<p>string CommunicationOperator = 1;</p><p>string Step = 2;</p><p>string SaveCheckpoint = 3;</p><p>string FP =4;</p><p>string DataLoader =5;</p>}</p>|<p>**jobNsName**：所需修改的任务的命名空间和任务名称，以’/’拼接，如：default/test-pytorch。</p><p>**profilingSwitch**：各类开关详情。</p><ul><li>**CommunicationOperator**：通信算子开关。</li><li>**Step**：Step时延开关。</li><li>**SaveCheckpoint**：SaveCheckpoint耗时开关。</li><li>**FP**：前向传播数据开关。</li><li>**DataLoader**：DataLoader耗时开关。</li></ul>|

**返回值说明<a name="section7920469381"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataTypeRes|message DataTypeRes{<p>string message = 1;</p><p>int32 code = 2;</p>}|<p>**message**：接口调用结果信息。</p><p>**code**：接口调用返回码。</p><ul><li>1：300，入参不合法。</li><li>2：404，无法查询ConfigMap。</li><li>3：500，服务端异常。</li><li>4：200，接口正常返回。</li></ul>|

## GetTrainingDataTraceSwitch<a name="ZH-CN_TOPIC_0000002479386852"></a>

**功能说明<a name="section21882190424"></a>**

外部调用获取各类数据动态打点开关状态。

**函数原型<a name="section1723573217426"></a>**

```proto
rpc GetTrainingDataTraceSwitch(DataStatusReq) returns (DataStatusRes) {}
```

**输入参数说明<a name="section19921040164215"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataStatusReq|message DataStatusReq{<p>string jobNsName = 1;</p>}|**jobNsName**：所需修改的任务的命名空间和任务名称，以’/’拼接，如：default/test-pytorch。|

**返回值说明<a name="section93011951104217"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataStatusRes|message DataStatusRes{<p>string message = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p><p>int32 code = 3;</p>}|<p>**message**：接口调用结果信息。</p><p>**profilingSwitch**：各类开关详情。</p><ul><li>**CommunicationOperator**：通信算子开关。</li><li>**Step**：Step时延开关。</li><li>**SaveCheckpoint**：SaveCheckpoint耗时开关。</li><li>**FP**：前向传播数据开关。</li><li>**DataLoader**：DataLoader耗时开关。</li></ul>**code**：接口调用返回码。<ul><li>1：300，入参不合法。</li><li>2：404，无法查询ConfigMap。</li><li>3：500，服务端异常。</li><li>4：200，接口正常返回。</li></ul>|

## SubscribeDataTraceSwitch<a name="ZH-CN_TOPIC_0000002511346751"></a>

**功能说明<a name="section22878209356"></a>**

外部订阅各类数据动态打点开关状态。

**函数原型<a name="section1472624833519"></a>**

```proto
rpc SubscribeDataTraceSwitch(ProfilingClientInfo) returns (stream DataStatusRes) {}
```

**输入参数说明<a name="section6782115723515"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|ProfilingClientInfo|message ProfilingClientInfo{<p>string jobId = 1;</p><p>string role = 2;</p>}|<p>**jobId**：任务ID。</p><p>**role**：客户端角色。</p>|

**返回值说明<a name="section7920469381"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|DataStatusRes|message DataStatusRes{<p>string message = 1;</p><p>ProfilingSwitch profilingSwitch = 2;</p><p>int32 code = 3;</p>}|<p>**message**：接口调用结果信息。</p><p>**profilingSwitch**：各类开关详情。</p><ul><li>**CommunicationOperator**：通信算子开关。</li><li>**Step**：Step时延开关。</li><li>**SaveCheckpoint**：SaveCheckpoint耗时开关。</li><li>**FP**：前向传播数据开关。</li><li>**DataLoader**：DataLoader耗时开关。</li></ul>**code**：接口调用返回码。<ul><li>1：300，入参不合法。</li><li>2：404，无法查询ConfigMap。</li><li>3：500，服务端异常。</li><li>4：200，接口正常返回。</li></ul>|
