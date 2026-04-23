# TaskD Proxy接口<a name="ZH-CN_TOPIC_0000002479386846"></a>

## def init\_taskd\_proxy\(config : dict\) -\> bool:<a name="ZH-CN_TOPIC_0000002479226870"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，初始化TaskD Proxy。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|config|dict:{str : str}|TaskD Proxy配置信息，包括TaskD Proxy配置及网络配置。<ul><li>ListenAddr：TaskD Proxy侦听IP</li><li>ListenPort：TaskD Proxy侦听端口</li><li>UpstreamAddr：网络侧上游IP地址</li><li>UpstreamPort：网络侧上游端口</li><li>ServerRank：TaskD Proxy rank号</li></ul>|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<ul><li>True：初始化成功。</li><li>False：初始化失败。</li></ul>|

## def destroy\_taskd\_proxy\(\) -\> bool:<a name="ZH-CN_TOPIC_0000002479226806"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，销毁TaskD Proxy。此函数需要在[init\_taskd\_proxy](#def-init_taskd_proxyconfig--dict---bool)接口后使用。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明销毁是否成功。<ul><li>True：销毁成功。</li><li>False：销毁失败。</li></ul>|
