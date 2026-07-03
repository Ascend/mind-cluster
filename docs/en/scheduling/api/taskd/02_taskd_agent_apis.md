# TaskD Agent APIs<a name="ZH-CN_TOPIC_0000002479226872"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:43:05.254Z pushedAt=2026-06-09T02:05:50.642Z -->

## def init_taskd_agent(config : dict = {}, cls = None) -> bool<a name="ZH-CN_TOPIC_0000002511426763"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to initialize TaskD Agent.

**Input Parameters<a name="section1177311115553"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|
|--|--|--|
|config|dict:{str : str}|Agent configuration information, including Agent configuration and network configuration. The keys include:<ul><li>`Framework`: Agent framework, currently supports PyTorch and MindSpore</li><li>`UpstreamAddr`: Upstream IP address on the network side</li><li>`UpstreamPort`: Upstream port on the network side</li><li>ServerRank: Agent rank number</li></ul>|
|cls|Specific instance type|This input parameter is used under the PyTorch framework and is a `SimpleElasticAgent` instance. It is not required for other frameworks.|

**Return Value<a name="section4468173015517"></a>**

**Table 2** Return value description

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the initialization is successful.<ul><li>`True`: Initialization succeeded.</li><li>`False`: Initialization failed.</li></ul>|

## def start_taskd_agent():<a name="ZH-CN_TOPIC_0000002479226808"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to start TaskD Agent.

**Input Parameters<a name="section1177311115553"></a>**

None

**Return Value<a name="section4468173015517"></a>**

**Table 1** Return value description

|Return Value Type|Description|
|--|--|
|Varies|The return result is determined by the main execution logic of the Agent under the framework. Agents under different frameworks will have different return results after startup. For example, under the PyTorch framework, `SimpleElasticAgent run()` will return the training result.|

## def register_func(operator, func) -> bool:<a name="ZH-CN_TOPIC_0000002511426733"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to register a TaskD Agent callback function.

**Input Parameters<a name="section1177311115553"></a>**

**Table 1** Parameters

|Parameter|Type|Description|
|--|--|--|
|operator|str|Key for registering the callback function, such as `START_ALL_WORKER`.|
|func|callable|The corresponding callback function.|

**Return Value <a name="section4468173015517"></a>**

**Table 2**  Return value description

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the registration is successful.<ul><li>`True`: Registration succeed.</li><li>`False`: Registration failed.</li></ul>|
