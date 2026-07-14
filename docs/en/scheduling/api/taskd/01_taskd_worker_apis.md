# TaskD Worker APIs<a name="ZH-CN_TOPIC_0000002479386850"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:43:27.477Z pushedAt=2026-06-09T02:05:50.647Z -->

## def init_taskd_worker(rank_id: int, upper_limit_of_disk_in_mb: int = 5000, framework: str = "pt") -> bool<a name="ZH-CN_TOPIC_0000002479226866"></a>

**Function Description<a name="section1931361114330"></a>**

The user-side code calls this function to initialize TaskD Worker.

**Parameters<a name="section126587317332"></a>**

**Table 1**  Input parameters

|Parameter|Type|Description|
|--|--|--|
|rank_id|int|Global rank ID of the current training process.|
|upper_limit_of_disk_in_mb|int|Upper limit of storage space in the profiling folder that all training processes can use. The actual size fluctuates around this threshold, in MB. A non-negative value, defaulted to `5000`.|
|framework|str|AI framework used by the job.|

**Return Value<a name="section134891539193315"></a>**

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the initialization is successful.<ul><li>`True:` Initialization succeed.</li><li>`False`: Initialization failed.</li></ul>|

## def start_taskd_worker() -> bool<a name="ZH-CN_TOPIC_0000002511346737"></a>

**Function Description<a name="section1458863753514"></a>**

The user-side code calls this function to start TaskD Worker.

**Parameters<a name="section1574654643513"></a>**

No input parameters.

**Return Value<a name="section1871411618361"></a>**

|Parameter|Description|
|--|--|
|bool|Indicates whether the initialization is successful.<ul><li>`True:` Initialization succeed.</li><li>`False`: Initialization failed.</li></ul>|

## def destroy_taskd_worker() -> bool:<a name="ZH-CN_TOPIC_0000002511426721"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to destroy the TaskD Worker communication resources. This function must be used after [init_taskd_worker](#ZH-CN_TOPIC_0000002479226866) is called.

**Parameters<a name="section1177311115553"></a>**

None

**Return Value<a name="section4468173015517"></a>**

**Table 1**  Return value

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the destruction was successful.<ul><li>True: Destruction succeeded.</li><li>False: Destruction failed.</li></ul>|
