# Internal TaskD APIs<a name="ZH-CN_TOPIC_0000002479386822"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:44:11.302Z pushedAt=2026-06-09T02:05:50.664Z -->

## Register (Internal, Do Not Call)<a name="ZH-CN_TOPIC_0000002479226852"></a>

**Function Description<a name="section3468140175411"></a>**

Registers a role.

**Prototype<a name="section1818889191813"></a>**

<pre>
rpc Register(RegisterReq) returns (Ack)</pre>

**Input Parameters<a name="section1177311115553"></a>**

**Table 1**  Parameter description

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|RegisterReq|message RegisterReq {<p>string  uuid = 1;</p><p>Position pos = 2;</p>}<p>message Position {<p>string role = 1;</p><p>string serverRank = 2;</p><p>string processRank = 3;</p>}</p>|<p>**uuid**: Registration message UUID</p><p>**pos**: Registration message source</p><p>**role**: Registered role: such as Proxy, Worker, Agent, Mgr</p><p>**serverRank**: Server Rank information of the role</p><p>**processRank**: Process rank information of the role. Required for Worker role; Proxy, Agent, Mgr roles do not involve this information, uniformly fill in -1</p>|

**Return Value<a name="section4468173015517"></a>**

**Table 2**  Return value description

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**: Consistent with the UUID of the registration message</p><p>**code**: Return Code<li>Value `0`: Registration success</li><li>Other Values: Registration Failure</li></p><p>**src**: Position information of the role returning the Ack confirmation message</p>|

## PathDiscovery (Internal, Do Not Call)<a name="ZH-CN_TOPIC_0000002479226818"></a>

**Function Description<a name="section3468140175411"></a>**

Discovers paths.

**Prototype<a name="section1818889191813"></a>**

<pre>
rpc PathDiscovery(PathDiscoveryReq) returns (Ack)</pre>

**Input Parameters<a name="section1177311115553"></a>**

**Table 1** Parameter description

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|PathDiscoveryReq|message PathDiscoveryReq {<p>string  uuid = 1;</p><p>Position proxyPos = 2;</p><p>repeated Position path = 3;</p>}|<p>**uuid**: Message UUID</p><p>**proxyPos**: Position information of the role that initiates the `PathDiscovery` request</p><p>**path**: List of position information of roles that the `PathDiscovery` request passes through</p>|

**Return Value<a name="section4468173015517"></a>**

**Table 2**  Return value description

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|Ack|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**: Consistent with the UUID of the PathDiscovery Message</p><p>**code**: Return Code<ul><li>`0`: PathDiscovery API call success</li><li>Other Values: PathDiscovery API call Failure</li></ul></p><p>**src**: Destination Role position Information of the Ack confirmation message return</p>|

## TransferMessage (Internal, Do Not Call)<a name="ZH-CN_TOPIC_0000002479226848"></a>

**Function Description<a name="section3468140175411"></a>**

Sends a message.

**Prototype<a name="section1818889191813"></a>**

<pre>
rpc TransferMessage(Message) returns (Ack)</pre>

**Input Parameters<a name="section1177311115553"></a>**

**Table 1**  Parameter description

| Parameter | Type (Protobuf Definition) | Description |
|--|--|--|
| Message | <p>message MessageHeader {<p>string uuid = 1;</p><p>string mtype = 2;</p><p>bool sync = 3;</p><p>Position src = 4;</p><p>Position dst = 5;</p><p>int64 createTime = 6;</p>}</p><p>message Message {<p>MessageHeader header = 1;</p><p>string body = 2;</p>}</p> | <p>**uuid**: Message UUID</p><p>**mtype**: Message type</p><p>**sync**: Whether to send synchronously</p><p>**src**: Message source information</p><p>**dst**: Message destination information</p><p>**createTime**: Message creation timestamp</p><p>**header**: Message header</p><p>**body**: Message body</p> |

**Return Value<a name="section4468173015517"></a>**

**Table 2**  Return value description

| Return Value | Type (Protobuf Definition) | Description |
|--|--|--|
| Ack | message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>} | <p>**uuid**: Consistent with the message UUID in MessageHeader</p><p>**code**: Return Code<ul><li>`0`: Message sending success</li><li>Other Values: Message sending failure</li></ul></p><p>**src**: Role position information of the Ack confirmation message returner</p> |

## InitServerDownStream (Internal, Do Not Call)<a name="ZH-CN_TOPIC_0000002511346741"></a>

**Function Description<a name="section3468140175411"></a>**

Subscribes to messages from the server.

**Prototype<a name="section1818889191813"></a>**

<pre>
rpc InitServerDownStream(stream Ack) returns (stream Message)</pre>

**Input Parameters<a name="section1177311115553"></a>**

**Table 1** Parameter description

|Parameter|Type (Protobuf Definition)|Description|
|--|--|--|
|stream Ack|<p>message MessageHeader {<p>string uuid = 1;</p><p>string mtype = 2;</p><p>bool sync = 3;</p><p>Position src = 4;</p><p>Position dst = 5;</p><p>int64 createTime = 6;</p>}</p><p>message Message {<p>MessageHeader header = 1;</p><p>string body = 2;</p>}</p>|<p>**uuid**: Message UUID</p><p>**mtype**: Message type</p><p>**sync**: Whether to send synchronously</p><p>**src**: Message source information</p><p>**dst**: Message destination information</p><p>**createTime**: Message creation timestamp</p><p>**header**: Message header</p><p>**body**: Message body</p>|

**Return Value<a name="section4468173015517"></a>**

**Table 2** Return value description

|Return Value|Type (Protobuf Definition)|Description|
|--|--|--|
|stream Message|message Ack {<p>string uuid = 1;</p><p>uint32 code = 2;</p><p>Position src = 3;</p>}|<p>**uuid**: Consistent with Message.uuid</p><p>**code**: Return code<ul><li>`0`: Message sent successfully</li><li>Other values: Message sending failed</li></ul></p><p>**src**: Role position information of the Ack confirmation message sender</p>|

## run_log (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002479226820"></a>

**Function Description<a name="section3468140175411"></a>**

TaskD log object.

## Validator (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002479386808"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## FileValidator (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002511346777"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## StringValidator (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002479226846"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## DirectoryValidator (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002511346743"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## IntValidator (Internal, Do Not Modify or Call) <a name="ZH-CN_TOPIC_0000002479226828"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## MapValidator (Internal, Do Not Modify or Call) <a name="ZH-CN_TOPIC_0000002479386828"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## RankSizeValidator (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002511426745"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## ClassValidator (Internal, Do Not Modify or Call)<a name="ZH-CN_TOPIC_0000002511346755"></a>

**Function Description<a name="section3468140175411"></a>**

External parameter verification class.

## Return Codes<a name="ZH-CN_TOPIC_0000002511426777"></a>

The return codes of internal TaskD APIs are shown in the following table.

**Table 1** Return codes

|Return Code|Value|Meaning|
|--|--|--|
|NilMessage|4000|Message is empty|
|NilHeader|4001|Message header is empty|
|NilPosition|4002|Position information is empty|
|DstRoleIllegal|4003|Destination role is invalid|
|DstSrvRankIllegal|4004|Destination role's server rank is invalid|
|DstProcessRankIllegal|4005|Destination role's process rank is invalid|
|DstTypeIllegal|4006|Destination role type is invalid|
|ClientErr|4999|TaskD client error|
|RecvBufNil|5000|Receive buffer is empty|
|RecvBufBusy|5001|Receive buffer is blocked|
|NoRoute|5002|No routing path|
|ExceedMaxRegistryNum|5003|TaskD network registration exceeds the maximum limit|
|ServerErr|5999|TaskD server error|
|NetworkSendLost|6000|Message sending lost|
|NetworkAckLost|6001|Message ACK lost|
|NetStreamNotInited|6002|gRPC stream not initialized|
|NetErr|6999|TaskD network error|
