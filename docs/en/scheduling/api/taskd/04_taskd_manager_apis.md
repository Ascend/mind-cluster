# TaskD Manager Interface<a name="ZH-CN_TOPIC_0000002479386782"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:43:32.836Z pushedAt=2026-06-09T02:05:50.654Z -->

## def init_taskd_manager(config:dict) -> bool:<a name="ZH-CN_TOPIC_0000002479386834"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to initializeTaskD Manager.

**Input Parameters<a name="section1177311115553"></a>**

**Table 1**  Parameter description

|Parameter|Type|Description|
|--|--|--|
|config|dict:{str : str}|TaskD Manager configuration information, passed in as key-value pairs. The keys include:<ul><li>`job_id`: string type, indicating the job ID.</li><li>`node_nums`: int type, indicating the number of nodes.</li><li>`proc_per_node`: int type, indicating the number of processes per node.</li><li>`plugin_dir`: string type, indicating the plugin directory.</li><li>`fault_recover`: string type, indicating the fault recovery policy.</li><li>`taskd_enable`: string type, indicating the switch for the TaskD process-level recovery feature.</li><li>`cluster_infos`: dict type, indicating cluster information. The keys of `cluster_infos` are `ip` (the IP address of the current node), `port` (the server port), `name` (the server name), and `role` (the server role), all of which are string types.</li></ul>|

**Return Value<a name="section4468173015517"></a>**

**Table 2**  Return value description

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the initialization is successful.<ul><li>`True`: Initialization succeeded.</li><li>`False`: Initialization failed.</li></ul>|

## def start_taskd_manager() -> bool:<a name="ZH-CN_TOPIC_0000002479226810"></a>

**Function Description<a name="section3468140175411"></a>**

The user-side code calls this function to start TaskD Manager.

**Input Parameters<a name="section1177311115553"></a>**

None

**Return Value <a name="section4468173015517"></a>**

**Table 1**  Return value description

|Return Value Type|Description|
|--|--|
|bool|Indicates whether the startup was successful.<ul><li>`True`: Startup succeed.</li><li>`False`: Startup failed.</li></ul>|
