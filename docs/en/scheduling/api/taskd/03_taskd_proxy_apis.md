# TaskD Proxy APIs<a name="ZH-CN_TOPIC_0000002479386846"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:43:27.292Z pushedAt=2026-06-09T02:05:50.644Z -->

## def init_taskd_proxy(config : dict) -> bool:<a name="ZH-CN_TOPIC_0000002479226870"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to initialize TaskD Proxy.

**Input Parameters<a name="section1177311115553"></a>**

**Table 1**  Parameter description

|Parameter|Type|Description|
|--|--|--|
|config|dict:{str : str}|TaskD Proxy configuration information, including TaskD Proxy configuration and network configuration.<ul><li>`ListenAddr`: TaskD Proxy listening IP</li><li>`ListenPort`: TaskD Proxy listening port</li><li>`UpstreamAddr`: Upstream IP address on the network side</li><li>`UpstreamPort`: Upstream port on the network side</li><li>`ServerRank`: TaskD Proxy rank number</li></ul>|

**Return Value<a name="section4468173015517"></a>**

**Table 2**  Return value description

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the initialization is successful.<ul><li>`True`: The initialization is successful.</li><li>`False`: The initialization fails.</li></ul>|

## def destroy_taskd_proxy() -> bool: <a name="ZH-CN_TOPIC_0000002479226806"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to destroy TaskD Proxy. This function must be used after [init_taskd_proxy](#ZH-CN_TOPIC_0000002479226870) is called.

**Input Parameters<a name="section1177311115553"></a>**

None

**Return Value Description<a name="section4468173015517"></a>**

**Table 1**  Return value description

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the destruction is successful.<ul><li>`True`: The destruction is successful.</li><li>`False`: The destruction fails.</li></ul>|
