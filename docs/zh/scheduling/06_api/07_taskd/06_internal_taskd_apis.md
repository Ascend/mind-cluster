# TaskD内部接口<a name="ZH-CN_TOPIC_0000002479386822"></a>

## Register接口（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226852"></a>

**功能说明<a name="section3468140175411"></a>**

注册角色。

**函数原型<a name="section1818889191813"></a>**

<pre>
rpc Register(RegisterReq) returns (Ack)</pre>

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|RegisterReq|message RegisterReq {<p>string  uuid = 1;</p><p>Position pos = 2;</p>}<p>message Position {<p>string role = 1;</p><p>string serverRank = 2;</p><p>string processRank = 3;</p>}</p>|<p>**uuid**：注册消息UUID</p><p>**pos**：注册消息来源</p><p>**role**：注册的角色：如Proxy，Worker，Agent，Mgr</p><p>**serverRank**：角色所在server Rank信息</p><p>**processRank**：角色所在进程Rank信息。Worker角色需要填写；Proxy、Agent、Mgr角色不涉及此信息，统一填写-1</p>|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与注册消息UUID一致</p><p>**code**：返回码<li>取值为0：注册成功</li><li>其他值：注册失败</li></p><p>**src**：Ack确认消息返回方角色位置信息</p>|

## PathDiscovery（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226818"></a>

**功能说明<a name="section3468140175411"></a>**

路径发现。

**函数原型<a name="section1818889191813"></a>**

<pre>
rpc PathDiscovery(PathDiscoveryReq) returns (Ack)</pre>

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|PathDiscoveryReq|message PathDiscoveryReq {<p>string  uuid = 1;</p><p>Position proxyPos = 2;</p><p>repeated Position path = 3;</p>}|<p>**uuid**：消息UUID</p><p>**proxyPos**：PathDiscovery请求发起角色的位置信息</p><p>**path**：PathDiscovery请求经过的角色位置信息列表</p>|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与PathDiscovery消息UUID一致</p><p>**code**：返回码<ul><li>取值为0：PathDiscovery接口调用成功</li><li>其他值：PathDiscovery接口调用失败</li></ul></p><p>**src**：Ack确认消息返回方角色位置信息</p>|

## TransferMessage（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226848"></a>

**功能说明<a name="section3468140175411"></a>**

发送消息。

**函数原型<a name="section1818889191813"></a>**

<pre>
rpc TransferMessage(Message) returns (Ack)</pre>

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|Message|<p>message MessageHeader {<p>string uuid = 1;</p><p>string mtype = 2;</p><p>bool sync = 3;</p><p>Position src = 4;</p><p>Position dst = 5;</p><p>int64 createTime = 6;</p>}</p><p>message Message {<p>MessageHeader header = 1;</p><p>string body = 2;</p>}</p>|<p>**uuid**：消息UUID</p><p>**mtype**：消息类型</p><p>**sync**：是否同步发送</p><p>**src**：消息来源信息</p><p>**dst**：消息目的信息</p><p>**createTime**：消息创建时间戳</p><p>**header**：消息头</p><p>**body**：消息体</p>|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与MessageHeader中的消息UUID一致</p><p>**code**：返回码<ul><li>取值为0：消息发送成功</li><li>其他值：消息发送失败</li></ul></p><p>**src**：Ack确认消息返回方角色位置信息</p>|

## InitServerDownStream（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002511346741"></a>

**功能说明<a name="section3468140175411"></a>**

从服务端订阅消息。

**函数原型<a name="section1818889191813"></a>**

<pre>
rpc InitServerDownStream(stream Ack) returns (stream Message)</pre>

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|stream Ack|<p>message MessageHeader {<p>string uuid = 1;</p><p>string mtype = 2;</p><p>bool sync = 3;</p><p>Position src = 4;</p><p>Position dst = 5;</p><p>int64 createTime = 6;</p>}</p><p>message Message {<p>MessageHeader header = 1;</p><p>string body = 2;</p>}</p>|<p>**uuid**：消息UUID</p><p>**mtype**：消息类型</p><p>**sync**：是否同步发送</p><p>**src**：消息来源信息</p><p>**dst**：消息目的信息</p><p>**createTime**：消息创建时间戳</p><p>**header**：消息头</p><p>**body**：消息体</p>|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|stream Message|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**：与Message.uuid一致</p><p>**code**：返回码<ul><li>取值为0：消息发送成功</li><li>其他值：消息发送失败</li></ul></p><p>**src**：Ack确认消息返回方角色位置信息</p>|

## run\_log（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479226820"></a>

**功能说明<a name="section3468140175411"></a>**

TaskD日志对象。

## Validator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479386808"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## FileValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511346777"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## StringValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479226846"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## DirectoryValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511346743"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## IntValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479226828"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## MapValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002479386828"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## RankSizeValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511426745"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## ClassValidator（内部接口，严禁修改调用）<a name="ZH-CN_TOPIC_0000002511346755"></a>

**功能说明<a name="section3468140175411"></a>**

外部参数校验类。

## 返回码<a name="ZH-CN_TOPIC_0000002511426777"></a>

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
