# TaskD Worker接口<a name="ZH-CN_TOPIC_0000002479386850"></a>

## def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool<a name="ZH-CN_TOPIC_0000002479226866"></a>

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
|bool|表明初始化是否成功。<ul><li>True：初始化成功。</li><li>False：初始化失败。</li></ul>|

## def start\_taskd\_worker\(\) -\> bool<a name="ZH-CN_TOPIC_0000002511346737"></a>

**功能说明<a name="section1458863753514"></a>**

用户侧代码调用此函数，启动TaskD Worker。

**输入参数说明<a name="section1574654643513"></a>**

无输入参数。

**返回值说明<a name="section1871411618361"></a>**

|参数|说明|
|--|--|
|bool|表明初始化是否成功。<ul><li>True：初始化成功。</li><li>False：初始化失败。</li></ul>|

## def destroy\_taskd\_worker\(\) -\> bool:<a name="ZH-CN_TOPIC_0000002511426721"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，销毁TaskD worker通信资源。此函数需要在[init\_taskd\_worker](#ZH-CN_TOPIC_0000002479226866)接口后使用。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明销毁是否成功。<ul><li>True：销毁成功。</li><li>False：销毁失败。</li></ul>|
