# TaskD Manager接口<a name="ZH-CN_TOPIC_0000002479386782"></a>

## def init\_taskd\_manager\(config:dict\) -\> bool:<a name="ZH-CN_TOPIC_0000002479386834"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，初始化TaskD Manager。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|config|dict:{str : str}|TaskD Manager配置信息，以键值对形式传入。其中键包括：<ul><li>job_id：string类型，表示任务ID。</li><li>node_nums：int类型，表示节点数量。</li><li>proc_per_node：int类型，表示每节点进程数量。</li><li>plugin_dir：string类型，表示插件目录。</li><li>fault_recover：string类型，表示故障恢复策略。</li><li>taskd_enable：string类型，表示TaskD进程级恢复功能开关。</li><li>cluster_infos：dict类型，表示集群信息。cluster_infos的key分别为ip（当前节点的IP地址）、port（服务器端口）、name（服务器名称）、role（服务器角色），均为string类型。</li></ul>|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<ul><li>True：初始化成功。</li><li>False：初始化失败。</li></ul>|

## def start\_taskd\_manager\(\) -\> bool:<a name="ZH-CN_TOPIC_0000002479226810"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，启动TaskD Manager。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明启动是否成功。<ul><li>True：启动成功。</li><li>False：启动失败。</li></ul>|
