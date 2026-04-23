# TaskD Agent接口<a name="ZH-CN_TOPIC_0000002479226872"></a>

## def init\_taskd\_agent\(config : dict = \{\}, cls = None\) -\> bool<a name="ZH-CN_TOPIC_0000002511426763"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，初始化TaskD Agent。

**输入参数说明<a name="section1177311115553"></a>**

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|config|dict:{str : str}|Agent配置信息，包括Agent配置与网络配置。其中键包括：<ul><li>Framework：Agent框架，当前支持PyTorch和MindSpore</li><li>UpstreamAddr：网络侧上游IP地址</li><li>UpstreamPort：网络侧上游端口</li><li>ServerRank：Agent rank号</li></ul>|
|cls|具体实例类型|该入参在PyTorch框架下使用，为SimpleElasticAgent实例。其他框架无需传入。|

**返回值说明<a name="section4468173015517"></a>**

**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|bool|表明初始化是否成功。<ul><li>True：初始化成功。</li><li>False：初始化失败。</li></ul>|

## def start\_taskd\_agent\(\):<a name="ZH-CN_TOPIC_0000002479226808"></a>

**功能说明<a name="section3468140175411"></a>**

用户侧调用此函数，启动TaskD Agent。

**输入参数说明<a name="section1177311115553"></a>**

无

**返回值说明<a name="section4468173015517"></a>**

**表 1**  返回值说明

|返回值类型|说明|
|--|--|
|不固定|返回结果由框架下Agent中主要运行逻辑决定，不同框架下的Agent启动后会有不同的返回结果。例如，PyTorch框架下，SimpleElasticAgent run()会返回训练结果。|

## def register\_func\(operator, func\) -\> bool:<a name="ZH-CN_TOPIC_0000002511426733"></a>

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
|bool|表明注册是否成功。<ul><li>True：注册成功。</li><li>False：注册失败。</li></ul>|
