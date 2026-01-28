# TaskD<a name="ZH-CN_TOPIC_0000002511426759"></a>

## taskd.\_\_version\_\_<a name="ZH-CN_TOPIC_0000002479226868"></a>

**功能说明<a name="section8423617134017"></a>**

获取TaskD组件的版本号。

**输入参数说明<a name="section44514266413"></a>**

输入值：空。

**返回值说明<a name="section11284124817411"></a>**

返回值：TaskD组件的版本号。

使用样例如下：

```
import taskd
taskd.__version__
```


## TaskD Worker接口<a name="ZH-CN_TOPIC_0000002479386850"></a>

### def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool<a name="ZH-CN_TOPIC_0000002479226866"></a>

**功能说明<a name="section1931361114330"></a>**

用户侧代码调用此函数，初始化TaskD Worker。

**输入参数说明<a name="section126587317332"></a>**

**表 1**  输入参数说明

|参数|类型|说明|
|--|--|--|
|rank_id|int|当前训练进程的global rank号。|
|upper_limit_of_disk_in_mb|int|所有训练进程能使用的profiling文件夹存储空间上限，实际大小在此阈值上下波动，单位为MB，非负值，默认5000。|
|framework|str|表示任务所使用的AI框架。|


**返回值说明<a name="section134891539193315"></a>**

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<li>True：初始化成功。</li><li>False：初始化失败。</li>|



### def start\_taskd\_worker\(\) -\> bool<a name="ZH-CN_TOPIC_0000002511346737"></a>

**功能说明<a name="section1458863753514"></a>**

用户侧代码调用此函数，启动Taskd Worker。

**输入参数说明<a name="section1574654643513"></a>**

无输入参数。

**返回值说明<a name="section1871411618361"></a>**

|参数|说明|
|--|--|
|bool|表明初始化是否成功。<li>True：初始化成功。</li><li>False：初始化失败。</li>|



### def destroy\_taskd\_worker\(\) -\> bool:<a name="ZH-CN_TOPIC_0000002511426721"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，销毁TaskD worker通信资源。此函数需要在[init\_taskd\_worker](#def%20init_taskd_worker(rank_id:%20int,%20upper_limit_of_disk_in_mb:%20int%20=%205000,%20framework:%20str%20=%20"pt")%20->%20bool)接口后使用。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明销毁是否成功。<li>True：销毁成功。</li><li>False：销毁失败。</li>|




## TaskD Agent接口<a name="ZH-CN_TOPIC_0000002479226872"></a>

### def init\_taskd\_agent\(config : dict = \{\}, cls = None\) -\> bool<a name="ZH-CN_TOPIC_0000002511426763"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，初始化TaskD Agent。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|config|dict:{str : str}|Agent配置信息，包括Agent配置与网络配置。其中键包括：<li>Framework：Agent框架，当前支持PyTorch和MindSpore</li><li>UpstreamAddr：网络侧上游IP地址</li><li>UpstreamPort：网络侧上游端口</li><li>ServerRank：Agent rank号</li>|
|cls|具体实例类型|该入参在PyTorch框架下使用，为SimpleElasticAgent实例。其他框架无需传入。|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<li>True：初始化成功。</li><li>False：初始化失败。</li>|



### def start\_taskd\_agent\(\):<a name="ZH-CN_TOPIC_0000002479226808"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，启动TaskD Agent。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|不固定|返回结果由框架下Agent中主要运行逻辑决定，不同框架下的Agent启动后会有不同的返回结果。例如，PyTorch框架下，SimpleElasticAgent run()会返回训练结果。|



### def register\_func\(operator, func\) -\> bool:<a name="ZH-CN_TOPIC_0000002511426733"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，用于注册TaskD Agent回调函数。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|operator|str|注册回调函数键值，如START_ALL_WORKER。|
|func|callable|对应回调函数。|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明注册是否成功。<li>True：注册成功。</li><li>False：注册失败。</li>|




## TaskD Proxy接口<a name="ZH-CN_TOPIC_0000002479386846"></a>

### def init\_taskd\_proxy\(config : dict\) -\> bool:<a name="ZH-CN_TOPIC_0000002479226870"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，初始化TaskD Proxy。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|config|dict:{str : str}|TaskD Proxy配置信息，包括TaskD Proxy配置及网络配置。<li>ListenAddr：TaskD Proxy侦听IP</li><li>ListenPort：TaskD Proxy侦听端口</li><li>UpstreamAddr：网络侧上游IP地址</li><li>UpstreamPort：网络侧上游端口</li><li>ServerRank：TaskD Proxy rank号</li>|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<li>True：初始化成功。</li><li>False：初始化失败。</li>|



### def destroy\_taskd\_proxy\(\) -\> bool:<a name="ZH-CN_TOPIC_0000002479226806"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，销毁TaskD Proxy。此函数需要在[init\_taskd\_proxy](#def%20init_taskd_proxy(config%20:%20dict)%20->%20bool:)接口后使用。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明销毁是否成功。<li>True：销毁成功。</li><li>False：销毁失败。</li>|




## TaskD Manager接口<a name="ZH-CN_TOPIC_0000002479386782"></a>

### def init\_taskd\_manager\(config:dict\) -\> bool:<a name="ZH-CN_TOPIC_0000002479386834"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，初始化TaskD Manager。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|config|dict:{str : str}|TaskD Manager配置信息，以键值对形式传入。其中键包括：<li>job_id：string类型，表示任务ID。</li><li>node_nums：int类型，表示节点数量。</li><li>proc_per_node：int类型，表示每节点进程数量。</li><li>plugin_dir：string类型，表示插件目录。</li><li>fault_recover：string类型，表示故障恢复策略。</li><li>taskd_enable：string类型，表示TaskD进程级恢复功能开关。</li><li>cluster_infos：dict类型，表示集群信息。cluster_infos的key分别为ip（当前节点的IP地址）、port（服务器端口）、name（服务器名称）、role（服务器角色），均为string类型。</li>|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<li>True：初始化成功。</li><li>False：初始化失败。</li>|



### def start\_taskd\_manager\(\) -\> bool:<a name="ZH-CN_TOPIC_0000002479226810"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，启动TaskD Manager。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明启动是否成功。<li>True：启动成功。</li><li>False：启动失败。</li>|




## 断点续训相关接口<a name="ZH-CN_TOPIC_0000002479226856"></a>

### taskd.python.toolkit.recover\_module.recover\_manager. DLRecoverManager（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479386778"></a>

**功能说明<a name="section95016253292"></a>**

DLRecoverManager类提供进程级恢复和进程级在线恢复相关接口。客户端以Python包形式import到客户端代码中。

>[!NOTE] 说明 
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
|Status|<p>Status.info：str类型，返回信息描述</p><p>Status.code：int类型，0表示成功，其他值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。</p>|


**def start\_subscribe\(self, frame: str = "pytorch"\)<a name="section5051271214"></a>**

客户端和服务端建立gRPC长链接，服务端将通过该长链接与客户端单向通信。比如发生故障时，服务端给客户端发送停止训练、全局故障rank信息等。

**表 4**  参数说明

|参数|类型|说明|
|--|--|--|
|frame|str|表示任务所使用的AI框架。|


**init\_clusterd\(self\)<a name="section18270133519256"></a>**

客户端初始化ClusterD服务端状态，保证后续任务正常注册、建立链接。


### report\_stop\_complete\(code: int, msg: str, fault\_ranks: dict\) -\> int<a name="ZH-CN_TOPIC_0000002479386796"></a>

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
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



### report\_recover\_strategy\(fault\_ranks: dict, strategy\_list: list\) -\> int<a name="ZH-CN_TOPIC_0000002479386838"></a>

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
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



### report\_recover\_status\(code: int, msg: str, fault\_ranks: dict, strategy: str\) -\> int<a name="ZH-CN_TOPIC_0000002479226842"></a>

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
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



### report\_process\_fault\(fault\_ranks: dict\) -\> int<a name="ZH-CN_TOPIC_0000002511426703"></a>

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
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



### taskd.python.framework.agent.ms\_mgr.msrun\_plugin. MSRunPlugin<a name="ZH-CN_TOPIC_0000002511426749"></a>

MSRunPlugin类提供MindSpore进程管理功能，由MindSpore调用，集成到MindSpore包内部。


### register\_callbacks\(self, operator, func\)<a name="ZH-CN_TOPIC_0000002511346731"></a>

**功能说明<a name="section19441242061"></a>**

向TaskD注册进程管理函数，用于后续在管理进程生命周期过程中使用。

**输入参数说明<a name="section42271142719"></a>**

**表 1**  输入参数说明

|参数|类型|说明|
|--|--|--|
|operator|string|当前注入的回调类型。<li>KILL_WORKER：注册MindSpore进程的停止方法，停止特定训练进程。</li><li>START_ALL_WORKER：注册MindSpore进程的启动方法，启动当前节点所有的进程。</li><li>MONITOR：注册MindSpore进程的监测方法，返回当前本节点各rank进程信息。</li><li>START_WORKER_LIST：注册MindSpore进程的启动方法，启动当前节点的部分进程。</li>|
|func|函数|当前注册的功能的函数回调|



### start\(self\)<a name="ZH-CN_TOPIC_0000002479226816"></a>

调用MSRunPlugin start方法使TaskD接管MindSpore训练进程管理。


### \_\_init\_\_\(self\)<a name="ZH-CN_TOPIC_0000002511346791"></a>

构造MSRunPlugin类，用户后续实例化调用。



## TaskD内部接口<a name="ZH-CN_TOPIC_0000002479386822"></a>

### Register接口（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226852"></a>

**功能说明<a name="section3468140175411"></a>**

注册角色。

**函数原型<a name="section1818889191813"></a>**

```
rpc Register(RegisterReq) returns (Ack)
```

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|RegisterReq|<p>message RegisterReq {<p>string  uuid = 1;</p><p>Position pos = 2;}</p><p>message Position {<p>string role = 1;</p><p>string serverRank = 2;</p><p>string processRank = 3;</p>}</p>|<p>**uuid**：注册消息UUID</p><p>**pos**：注册消息来源</p><p>**role**：注册的角色：如Proxy，Worker，Agent，Mgr</p><p>**serverRank**：角色所在server Rank信息</p><p>**processRank**：角色所在进程Rank信息，包含如下几种类型：Proxy、Agent、Mgr不涉及此信息。这三类角色该字段统一填-1</p>|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与注册消息UUID一致</p><p>**code**：返回码<li>取值为0：注册成功</li><li>其他值：注册失败</li></p><p>**src**：Ack确认消息返回方角色位置信息|



### PathDiscovery（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226818"></a>

**功能说明<a name="section3468140175411"></a>**

路径发现。

**函数原型<a name="section1818889191813"></a>**

```
rpc PathDiscovery(PathDiscoveryReq) returns (Ack)
```

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|PathDiscoveryReq|message PathDiscoveryReq {<p>string  uuid = 1;</p><p>Position proxyPos = 2;</p><p>repeated Position path = 3;</p>}|<p>**uuid**：消息UUID</p><p>**proxyPos**：PathDiscovery请求发起角色的位置信息</p><p>**path**：PathDiscovery请求经过的角色位置信息列表</p>|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与PathDiscovery消息UUID一致</p><p>**code**：返回码<li>取值为0：PathDiscovery接口调用成功</li><li>其他值：PathDiscovery接口调用失败</li></p><p>**src**：Ack确认消息返回方角色位置信息|



### TransferMessage（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226848"></a>

**功能说明<a name="section3468140175411"></a>**

发送消息。

**函数原型<a name="section1818889191813"></a>**

```
rpc TransferMessage(Message) returns (Ack)
```

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Message|<p>message MessageHeader {<p>string uuid = 1;</p><p>string mtype = 2;</p><p>bool sync = 3;</p><p>Position src = 4;</p><p>Position dst = 5;</p><p>int64 createTime = 6;</p>}</p><p>message Message {<p>MessageHeader header = 1;</p><p>string body = 2;</p>}</p>|<p>**uuid**：消息UUID</p><p>**mtype**：消息类型</p><p>**sync**：是否同步发送</p><p>**src**：消息来源信息</p><p>**dst**：消息目的信息</p><p>**createTime**：消息创建时间戳</p><p>**header**：消息头</p><p>**body**：消息体</p>|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与MessageHeader中的消息UUID一致</p><p>**code**：返回码<li>取值为0：消息发送成功</li><li>其他值：消息发送失败</li></p><p>**src**：Ack确认消息返回方角色位置信息</p>|



### InitServerDownStream（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002511346741"></a>

**功能说明<a name="section3468140175411"></a>**

从服务端订阅消息。

**函数原型<a name="section1818889191813"></a>**

```
rpc InitServerDownStream(stream Ack) returns (stream Message)
```

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|stream Ack|<p>message MessageHeader {<p>string uuid = 1;</p><p>string mtype = 2;</p><p>bool sync = 3;</p><p>Position src = 4;</p><p>Position dst = 5;</p><p>int64 createTime = 6;</p>}</p><p>message Message {<p>MessageHeader header = 1;</p><p>string body = 2;</p>}</p>|<p>**uuid**：消息UUID</p><p>**mtype**：消息类型</p><p>**sync**：是否同步发送</p><p>**src**：消息来源信息</p><p>**dst**：消息目的信息</p><p>**createTime**：消息创建时间戳</p><p>**header**：消息头</p><p>**body**：消息体</p>|


**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream Message|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与Message.uuid一致</p><p>**code**：返回码<li>取值为0：消息发送成功</li><li>其他值：消息发送失败</li></p><p>**src**：Ack确认消息返回方角色位置信息</p>|



### run\_log（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479226820"></a>

**功能说明<a name="section3468140175411"></a>**

TaskD日志对象。


### Validator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479386808"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### FileValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511346777"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### StringValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479226846"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### DirectoryValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511346743"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### IntValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479226828"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### MapValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479386828"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### RankSizeValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511426745"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### ClassValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511346755"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。


### 返回码<a name="ZH-CN_TOPIC_0000002511426777"></a>

TaskD内部接口返回码如下表所示。

**表 1** TaskD内部接口返回码

|返回码|值|含义|
|--|--|--|
|NilMessage|4000|消息为空|
|NilHeader|4001|消息头为空|
|NilPosition|4002|位置信息为空|
|DstRoleIllegal|4003|目的角色非法|
|DstSrvRankIllegal|4004|目的角色的server Rank非法|
|DstProcessRankIllegal|4005|目的角色的进程Rank非法|
|DstTypeIllegal|4006|目的角色类型非法|
|ClientErr|4999|TaskD客户端错误|
|RecvBufNil|5000|接收缓冲区为空|
|RecvBufBusy|5001|接收缓冲区阻塞|
|NoRoute|5002|没有路由路径|
|ExceedMaxRegistryNum|5003|TaskD网络注册超过最大限制|
|ServerErr|5999|TaskD服务端错误|
|NetworkSendLost|6000|消息发送丢失|
|NetworkAckLost|6001|消息ACK丢失|
|NetStreamNotInited|6002|gRPC流未初始化|
|NetErr|6999|TaskD网络错误|




## 返回码说明<a name="ZH-CN_TOPIC_0000002511426711"></a>

TaskD返回码如下表所示。

**表 1** TaskD返回码

|返回码|值|含义|
|--|--|--|
|OK|0|接口调用正常。|
|UnRegistry|400|Job ID未注册。|
|OrderMix|401|请求不符合状态机顺序。|
|JobNotExist|402|Job ID不存在。|
|ProcessRescheduleOff|403|未打开进程级恢复开关。|
|ProcessNotReady|404|训练进程未拉起。|
|RecoverableRetryError|405|恢复失败，失败原因为clean device失败。|
|UnRecoverableRetryError|406|恢复失败，失败原因为stop device失败。|
|DumpError|407|临终遗言保存失败。|
|UnInit|408|未调用初始化。|
|ClientError|499|其他失败原因。|
|OutOfMaxServeJobs|500|超过最大服务任务数。|
|OperateConfigMapError|501|操作ConfigMap失败。|
|OperatePodGroupError|502|操作PodGroup失败。|
|ScheduleTimeout|503|Pod调度超时。|
|SignalQueueBusy|504|控制信号入队失败。|
|EventQueueBusy|505|状态机事件入队失败。|
|ControllerEventCancel|506|状态机已退出。|
|WaitReportTimeout|507|等待客户端调用接口超时。|
|WaitPlatStrategyTimeout|508|等待AI平台准备恢复策略超时。|
|WriteConfirmFaultOrWaitPlatResultFault|509|AI平台故障信息错误。|
|ServerInnerError|599|服务端内部错误。|



