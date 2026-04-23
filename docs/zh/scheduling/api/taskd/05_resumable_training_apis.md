# 断点续训相关接口<a name="ZH-CN_TOPIC_0000002479226856"></a>

## taskd.python.toolkit.recover\_module.recover\_manager. DLRecoverManager（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479386778"></a>

**功能说明<a name="section95016253292"></a>**

DLRecoverManager类提供进程级恢复和进程级在线恢复相关接口。客户端以Python包形式import到客户端代码中。

>[!NOTE] 
>DLRecoverManager类提供的接口可能抛出Exception异常，调用方自行捕获异常、处理异常。

**\_\_init\_\_\(self, info: pb.ClientInfo, server\_addr: str\)<a name="section93535281517"></a>**

构造DLRecoverManager，用于后续的通信。

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|info|pb.ClientInfo|<p>info.jobId：str类型，任务ID。</p><p>info.role：str类型，客户端角色。</p>|
|server_addr|str|服务端地址|

**register\(self, request: pb.ClientInfo\) -\> pb.Status<a name="section92911329181515"></a>**

注册客户端，服务端为request指定的任务做恢复前的初始化操作。

**表 2**  参数说明

|参数|类型|说明|
|--|--|--|
|request|pb.ClientInfo|<p>request.jobId：str类型，任务ID。</p><p>request.role：str类型，客户端角色。</p>|

**表 3**  返回值说明

|返回值类型|说明|
|--|--|
|Status|<p>Status.info：str类型，返回信息描述</p><p>Status.code：int类型，0表示成功，其他值表示失败。关于返回码的详细说明请参见[返回码说明](./07_return_codes.md)。</p>|

**def start\_subscribe\(self, frame: str = "pytorch"\)<a name="section5051271214"></a>**

客户端和服务端建立gRPC长链接，服务端将通过该长链接与客户端单向通信。比如发生故障时，服务端给客户端发送停止训练、全局故障rank信息等。

**表 4**  参数说明

|参数|类型|说明|
|--|--|--|
|frame|str|表示任务所使用的AI框架。|

**init\_clusterd\(self\)<a name="section18270133519256"></a>**

客户端初始化ClusterD服务端状态，保证后续任务正常注册、建立链接。

## report\_stop\_complete\(code: int, msg: str, fault\_ranks: dict\) -\> int<a name="ZH-CN_TOPIC_0000002479386796"></a>

**功能说明<a name="section1620210127300"></a>**

客户端给服务端上报任务的进程停止完成。一般是在客户端收到服务端的停止训练信号后，客户端停止训练任务的进程，然后给服务端上报任务的进程停止完成。

**输入参数说明<a name="section1793816299304"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|code|int|状态码|
|msg|str|返回信息|
|fault_ranks|dict|故障进程Rank|

**返回值说明<a name="section924216017310"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见[返回码说明](./07_return_codes.md)。|

## report\_recover\_strategy\(fault\_ranks: dict, strategy\_list: list\) -\> int<a name="ZH-CN_TOPIC_0000002479386838"></a>

**功能说明<a name="section350336124214"></a>**

客户端给服务端上报客户端支持的恢复策略，供服务端选择最佳恢复策略，服务端再通过start\_subscribe构建的长链接下发给客户端。

**输入参数说明<a name="section91358261429"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|fault_ranks|dict|故障进程Rank|
|strategy_list|list|恢复策略列表|

**返回值说明<a name="section1365711594319"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见[返回码说明](./07_return_codes.md)。|

## report\_recover\_status\(code: int, msg: str, fault\_ranks: dict, strategy: str\) -\> int<a name="ZH-CN_TOPIC_0000002479226842"></a>

**功能说明<a name="section9417169184510"></a>**

客户端给服务端上报任务恢复状态。

**输入参数说明<a name="section7968321124510"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|code|int类型|状态码|
|msg|str|返回信息|
|fault_ranks|dict|故障进程Rank|
|strategy|str|修复策略|

**返回值说明<a name="section1365711594319"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见[返回码说明](./07_return_codes.md)。|

## report\_process\_fault\(fault\_ranks: dict\) -\> int<a name="ZH-CN_TOPIC_0000002511426703"></a>

**功能说明<a name="section3468140175411"></a>**

客户端上报任务进程业务面故障。客户端先发现故障时，给服务端上报业务面故障所在的rank的信息。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|fault_ranks|dict|故障进程Rank|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见[返回码说明](./07_return_codes.md)。|

## taskd.python.framework.agent.ms\_mgr.msrun\_plugin. MSRunPlugin<a name="ZH-CN_TOPIC_0000002511426749"></a>

MSRunPlugin类提供MindSpore进程管理功能，由MindSpore调用，集成到MindSpore包内部。

## register\_callbacks\(self, operator, func\)<a name="ZH-CN_TOPIC_0000002511346731"></a>

**功能说明<a name="section19441242061"></a>**

向TaskD注册进程管理函数，用于后续在管理进程生命周期过程中使用。

**输入参数说明<a name="section42271142719"></a>**

**表 1**  输入参数说明

|参数|类型|说明|
|--|--|--|
|operator|string|当前注入的回调类型。<ul><li>KILL_WORKER：注册MindSpore进程的停止方法，停止特定训练进程。</li><li>START_ALL_WORKER：注册MindSpore进程的启动方法，启动当前节点所有的进程。</li><li>MONITOR：注册MindSpore进程的监测方法，返回当前本节点各rank进程信息。</li><li>START_WORKER_LIST：注册MindSpore进程的启动方法，启动当前节点的部分进程。</li></ul>|
|func|函数|当前注册的功能的函数回调|

## start\(self\)<a name="ZH-CN_TOPIC_0000002479226816"></a>

调用MSRunPlugin start方法使TaskD接管MindSpore训练进程管理。

## \_\_init\_\_\(self\)<a name="ZH-CN_TOPIC_0000002511346791"></a>

构造MSRunPlugin类，用户后续实例化调用。
