# 安装前必读<a name="ZH-CN_TOPIC_0000002511426285"></a>

在安装组件前，用户需详细阅读[简介](./introduction.md#概述)章节，了解集群调度各组件功能详细的说明，并根据要使用的特性选择安装相应的组件。

Elastic Agent和TaskD组件需部署在容器内，详细安装步骤请参见[制作镜像](./usage/resumable_training.md#制作镜像)。

>[!NOTE] 说明 
>Resilience Controller和Elastic Agent组件已经日落，Resilience Controller相关内容将于2026年的8.2.RC1版本删除；Elastic Agent相关内容将于2026年的8.3.0版本删除。

**使用约束<a name="section933252483715"></a>**

-   请确保根目录有足够的磁盘空间，根目录的磁盘空间利用率高于85%会触发kubelet的资源驱逐机制，将导致服务不可用。磁盘空间要求说明请参见[表1](#软硬件规格要求)；驱逐策略请查看[Kubernetes官方文档](https://kubernetes.io/zh-cn/docs/concepts/scheduling-eviction/node-pressure-eviction/)。
-   为保证MindCluster集群调度组件的正常安装及使用，同一集群下，不同训练服务器的系统时间请保持一致。
-   ARM架构和x86\_64架构使用的集群调度组件镜像不能相互兼容。
-   K8s默认的证书有效期为365天，到期前需要用户自行更新。

**组件部署说明<a name="section1563217510232"></a>**

安装部署集群调度组件时，可以参考[图1](#fig87391254145620)，将相应的集群调度组件或其他第三方软件安装到相应的节点上。大部分组件都使用容器化方式部署；Ascend Docker Runtime使用二进制方式部署；只有NPU Exporter组件既可以使用容器化方式部署，又可以使用二进制方式部署。

**图 1**  组件安装部署<a name="fig87391254145620"></a>  
![](../figures/scheduling/组件安装部署.png "组件安装部署")

>[!NOTE] 说明 
>MindCluster提供Volcano组件，该组件在开源Volcano上集成了昇腾插件Ascend-volcano-plugin。

**日志路径说明<a name="section4837236204914"></a>**

-   Ascend Docker Runtime日志路径为“/var/log/ascend-docker-runtime/“。
-   其他集群调度组件日志路径可参考[创建日志目录](#创建日志目录)章节。

# 支持的产品形态和OS清单<a name="ZH-CN_TOPIC_0000002511346411"></a>

集群场景下的管理节点、计算节点和存储节点支持的产品形态各不相同；其中计算节点支持的产品形态和单机场景支持的产品形态一致。

**集群场景<a name="section108952813220"></a>**

-   管理节点：支持多种类型服务器，如Taishan 200服务器（型号2280）、FusionServer Pro 2288H V5等。
-   计算节点：支持的产品形态和单机场景支持的产品一致，请参见下表。
-   存储节点：存储服务器。

**单机场景（训练）<a name="section18541427121413"></a>**

单机训练场景下，支持的产品形态和OS如下表所示。

**表 1**  支持的产品形态和OS

<a name="table7314423114217"></a>
<table><thead align="left"><tr id="row83141238425"><th class="cellrowborder" valign="top" width="15.93%" id="mcps1.2.4.1.1"><p id="p1731452318420"><a name="p1731452318420"></a><a name="p1731452318420"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.67%" id="mcps1.2.4.1.2"><p id="p183141923124210"><a name="p183141923124210"></a><a name="p183141923124210"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="50.4%" id="mcps1.2.4.1.3"><p id="p1367835115533"><a name="p1367835115533"></a><a name="p1367835115533"></a>操作系统</p>
</th>
</tr>
</thead>
<tbody><tr id="row193141923124213"><td class="cellrowborder" rowspan="5" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p13314132394215"><a name="p13314132394215"></a><a name="p13314132394215"></a><span id="ph1331492318423"><a name="ph1331492318423"></a><a name="ph1331492318423"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p123141723124213"><a name="p123141723124213"></a><a name="p123141723124213"></a>训练服务器（插<span id="ph113141423144220"><a name="ph113141423144220"></a><a name="ph113141423144220"></a>Atlas 300T 训练卡（型号 9000）</span>）</p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul183141847103713"></a><a name="ul183141847103713"></a><ul id="ul183141847103713"><li>CentOS 7.6 for x86</li><li>Kylin V10 SP1 for x86</li><li>openEuler 20.03 for x86</li><li>openEuler 22.03 for x86</li><li>Ubuntu 18.04.1 for x86</li><li>Ubuntu 20.04 for x86</li></ul>
</td>
</tr>
<tr id="row1231412319429"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p0314723124211"><a name="p0314723124211"></a><a name="p0314723124211"></a>训练服务器（插<span id="ph10314723164217"><a name="ph10314723164217"></a><a name="ph10314723164217"></a>Atlas 300T Pro 训练卡（型号 9000）</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul12789124011525"></a><a name="ul12789124011525"></a><ul id="ul12789124011525"><li>CentOS 7.6 for x86</li><li>Kylin V10 SP1 for x86</li><li>openEuler 20.03 for x86</li><li>openEuler 22.03 for x86</li><li>Ubuntu 18.04.1 for x86</li><li>Ubuntu 20.04 for x86</li></ul>
</td>
</tr>
<tr id="row113141823114217"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1631416238425"><a name="p1631416238425"></a><a name="p1631416238425"></a><span id="ph18314192319429"><a name="ph18314192319429"></a><a name="ph18314192319429"></a>Atlas 800 训练服务器（型号 9000）</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul388031125610"></a><a name="ul388031125610"></a><ul id="ul388031125610"><li>CentOS 7.6 for ARM</li><li>Kylin V10 SP2 for ARM</li><li>openEuler 20.03 for ARM</li><li>openEuler 22.03 for ARM</li><li>Ubuntu 20.04  + 5.15.0-25-generic kernel for ARM</li><li>Ubuntu 20.04  + 5.4.0-26-generic kernel for ARM</li><li>UOS V20 1020e for ARM</li></ul>
</td>
</tr>
<tr id="row1631417235426"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1331442313425"><a name="p1331442313425"></a><a name="p1331442313425"></a><span id="ph631452384213"><a name="ph631452384213"></a><a name="ph631452384213"></a>Atlas 800 训练服务器（型号 9010）</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul1083283212561"></a><a name="ul1083283212561"></a><ul id="ul1083283212561"><li>CentOS 7.6 for x86</li><li>Kylin V10 (OpenEuler) SP1 for x86</li><li>openEuler 20.03 for x86</li><li>Ubuntu 18.04.1 for x86</li><li>Ubuntu 18.04.5 for x86</li><li>Ubuntu 20.04 for x86</li></ul>
</td>
</tr>
<tr id="row1731412319425"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1931492374219"><a name="p1931492374219"></a><a name="p1931492374219"></a><span id="ph12314923104215"><a name="ph12314923104215"></a><a name="ph12314923104215"></a>Atlas 900 PoD（型号 9000）</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul1486510460560"></a><a name="ul1486510460560"></a><ul id="ul1486510460560"><li>CentOS 7.6 for ARM</li><li>Kylin V10 SP2 for ARM</li><li>openEuler 20.03 for ARM</li><li>openEuler 22.03 for ARM</li><li>Ubuntu 20.04  for ARM</li><li>UOS V20 1020e for ARM</li></ul>
</td>
</tr>
<tr id="row5314823154211"><td class="cellrowborder" rowspan="4" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p73141323184218"><a name="p73141323184218"></a><a name="p73141323184218"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p231472304210"><a name="p231472304210"></a><a name="p231472304210"></a></p>
<p id="p1531492319428"><a name="p1531492319428"></a><a name="p1531492319428"></a></p>
<p id="p9314152374214"><a name="p9314152374214"></a><a name="p9314152374214"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p5314102315424"><a name="p5314102315424"></a><a name="p5314102315424"></a><span id="ph1831422311424"><a name="ph1831422311424"></a><a name="ph1831422311424"></a>Atlas 200T A2 Box16 异构子框</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul66426318579"></a><a name="ul66426318579"></a><ul id="ul66426318579"><li>Debian 10.0 for x86</li><li>Debian 11.7 for x86<span id="ph1116132419442"><a name="ph1116132419442"></a><a name="ph1116132419442"></a> </span>(kernel 5.10.103)</li><li>Debian 12 for x86 (kernel 5.15.152.ve.10)</li><li>Ubuntu 22.04 for x86</li><li>Ubuntu 20.04.1 for x86</li><li>Ubuntu 22.04.1 for x86(5.16.20-051620-generic)</li><li>Tlinux 3.1 for x86</li><li>Tlinux 3.2 for x86</li><li>Tlinux 4.0 for x86<span id="ph1569872854619"><a name="ph1569872854619"></a><a name="ph1569872854619"></a> (</span>6.6内核)</li><li>openEuler 22.03 LTS SP4 for x86</li><li>openEuler 24.03 LTS for x86</li><li>openEuler 24.03 LTS SP1 for x86</li></ul>
</td>
</tr>
<tr id="row531414237429"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p831482324215"><a name="p831482324215"></a><a name="p831482324215"></a><span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul20696163813553"></a><a name="ul20696163813553"></a><ul id="ul20696163813553"><li>CentOS 7.6 for ARM</li><li>CTYunOS 22.06 for ARM</li><li>CTYunOS 23.01 for ARM</li><li>Kylin V10 SP2 for ARM</li><li>Kylin V10 SP3 for ARM</li><li>Kylin V11 for ARM</li><li>UOS V20 1050u2e for ARM</li><li>Ubuntu 22.04 for ARM</li><li>Ubuntu 22.04.4 LTS<span id="ph198698211464"><a name="ph198698211464"></a><a name="ph198698211464"></a> </span>(Linux 6.5.0-18-generic) for ARM</li><li>Ubuntu 24.04 LTS for ARM</li><li>openEuler 22.03 for ARM</li><li>openEuler 22.03 LTS SP2 for ARM</li><li>openEuler 22.03 LTS SP4 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>BC-Linux_21.10 U4  for ARM</li></ul>
</td>
</tr>
<tr id="row5314923164214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p11314172324210"><a name="p11314172324210"></a><a name="p11314172324210"></a><span id="ph1531415237425"><a name="ph1531415237425"></a><a name="ph1531415237425"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul0428918809"></a><a name="ul0428918809"></a><ul id="ul0428918809"><li>BC-Linux-for-Euler-21.10 for ARM</li><li>BC-Linux_21.10 U4  for ARM</li><li>Kylin V10 SP2 for ARM</li><li>Kylin V11 for ARM</li><li>CTYunOS 22.06 for ARM</li><li>openEuler 22.03 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>Ubuntu 22.04 for ARM</li><li>Ubuntu 24.04 LTS for ARM</li><li>HCE 2.0  for ARM</li></ul>
</td>
</tr>
<tr id="row13314162394218"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p14314142374220"><a name="p14314142374220"></a><a name="p14314142374220"></a><span id="ph1431418234423"><a name="ph1431418234423"></a><a name="ph1431418234423"></a>Atlas 900 A2 PoDc 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul1192418295016"></a><a name="ul1192418295016"></a><ul id="ul1192418295016"><li>openEuler 22.03 LTS SP4 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>CTYunOS 23.01 for ARM</li><li>Ubuntu 24.04 LTS for ARM</li><li>BC-Linux_21.10 U4  for ARM</li><li>Kylin V11 for ARM</li></ul>
</td>
</tr>
<tr id="row153141923154210"><td class="cellrowborder" rowspan="2" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p931415234425"><a name="p931415234425"></a><a name="p931415234425"></a><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
<p id="p153141423174218"><a name="p153141423174218"></a><a name="p153141423174218"></a></p>
<p id="p73152023194218"><a name="p73152023194218"></a><a name="p73152023194218"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p1315152304215"><a name="p1315152304215"></a><a name="p1315152304215"></a><span id="ph1731512317424"><a name="ph1731512317424"></a><a name="ph1731512317424"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul1330454618016"></a><a name="ul1330454618016"></a><ul id="ul1330454618016"><li>HCE 2.0  for ARM</li><li>Debian 10.2   for ARM</li><li>BC-Linux_21.10 U4  for ARM</li><li>MTOS for ARM</li><li>openEuler 24.03 LTS SP1  for ARM</li><li>openEuler 22.03 SP1 OS外围 + openEuler 24.03 SP1 6.6内核 for ARM</li><li>CentOS 7.5  for ARM</li><li>Linux Kernel 6.6  for ARM</li><li>CTYunOS 23.01 for ARM</li><li><span id="ph103053469018"><a name="ph103053469018"></a><a name="ph103053469018"></a>AntOS</span> 6.6.47 for ARM</li><li>Kylin V10 SP3 2403 for ARM</li><li>Kylin V11 for ARM</li><li>Velinux 2.0 for ARM</li><li>VesselOS 2.0 for ARM</li></ul>
</td>
</tr>
<tr id="row73155239422"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1231542318428"><a name="p1231542318428"></a><a name="p1231542318428"></a><span id="ph1031518232426"><a name="ph1031518232426"></a><a name="ph1031518232426"></a>Atlas 800T A3 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul19518634212"></a><a name="ul19518634212"></a><ul id="ul19518634212"><li>openEuler 22.03 LTS SP4 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>CUlinux 3.0 for ARM</li><li>HCE 2.0.2506 for ARM</li><li>Velinux 2.0 for ARM</li><li>Kylin V11 for ARM</li></ul>
</td>
</tr>
<tr id="row16507172322215"><td class="cellrowborder" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p3927183818221"><a name="p3927183818221"></a><a name="p3927183818221"></a><span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p166420278225"><a name="p166420278225"></a><a name="p166420278225"></a><span id="ph064214271224"><a name="ph064214271224"></a><a name="ph064214271224"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul164242719228"></a><a name="ul164242719228"></a><ul id="ul164242719228"><li>Tlinux 3.1 for x86</li><li>Tlinux 4.0 (6.6内核) for x86</li><li>Velinux 1.4 for x86</li></ul>
</td>
</tr>
<tr id="row7316192354212"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p331612304218"><a name="p331612304218"></a><a name="p331612304218"></a>注：</p>
<p id="p1431613237421"><a name="p1431613237421"></a><a name="p1431613237421"></a>6.0.RC2及以上版本支持在<span id="ph831682314426"><a name="ph831682314426"></a><a name="ph831682314426"></a>Atlas 900 A3 SuperPoD 超节点</span>上使用<span id="ph183161923104214"><a name="ph183161923104214"></a><a name="ph183161923104214"></a>Ascend Operator</span>组件的资源监测、整卡调度和断点续训特性。</p>
</td>
</tr>
</tbody>
</table>

**单机场景（推理）<a name="section105511161028"></a>**

单机推理场景下，支持的产品形态和OS如下表所示。

**表 2**  支持的产品形态和OS

<a name="table107471445138"></a>
<table><thead align="left"><tr id="row207471745039"><th class="cellrowborder" valign="top" width="15.93%" id="mcps1.2.4.1.1"><p id="p074714451932"><a name="p074714451932"></a><a name="p074714451932"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.67%" id="mcps1.2.4.1.2"><p id="p1374714519315"><a name="p1374714519315"></a><a name="p1374714519315"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="50.4%" id="mcps1.2.4.1.3"><p id="p1174714458319"><a name="p1174714458319"></a><a name="p1174714458319"></a>操作系统</p>
</th>
</tr>
</thead>
<tbody><tr id="row87471045632"><td class="cellrowborder" rowspan="9" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p174718453315"><a name="p174718453315"></a><a name="p174718453315"></a><span id="ph19590185162111"><a name="ph19590185162111"></a><a name="ph19590185162111"></a>Atlas 推理系列产品</span></p>
<p id="p374720458313"><a name="p374720458313"></a><a name="p374720458313"></a></p>
<p id="p87473453312"><a name="p87473453312"></a><a name="p87473453312"></a></p>
<p id="p167479451233"><a name="p167479451233"></a><a name="p167479451233"></a></p>
<p id="p147478457318"><a name="p147478457318"></a><a name="p147478457318"></a></p>
<p id="p187471745234"><a name="p187471745234"></a><a name="p187471745234"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p945141161810"><a name="p945141161810"></a><a name="p945141161810"></a><span id="ph111514573319"><a name="ph111514573319"></a><a name="ph111514573319"></a>Atlas 800 推理服务器（型号 3000）</span>（插<span id="ph84512411188"><a name="ph84512411188"></a><a name="ph84512411188"></a>Atlas 300I Pro 推理卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul511011394313"></a><a name="ul511011394313"></a><ul id="ul511011394313"><li>CentOS 7.6 for ARM</li><li>CTYunOS 23.01 for ARM</li><li>Kylin V10 SP1 for ARM</li><li>Kylin V10 SP2 for ARM</li><li>Kylin V10 SP3 2403 for ARM</li><li>Kylin V11 for ARM</li><li>openEuler 20.03 for ARM</li><li>openEuler 22.03 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>Ubuntu 20.04 for ARM</li><li>Euler 2.12 for ARM</li><li>Euler 2.13 for ARM</li><li>Euler 2.15 for ARM</li><li>Debian 10.2 for ARM</li><li>HCS 8.5.0 (Euler 2.12 ARM) for ARM</li><li>HCE 2.0 for ARM</li><li>HCE 2.0.2503 for ARM</li></ul>
</td>
</tr>
<tr id="row1574784513319"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p43708501184"><a name="p43708501184"></a><a name="p43708501184"></a><span id="ph629518508411"><a name="ph629518508411"></a><a name="ph629518508411"></a>Atlas 800 推理服务器（型号 3000）</span>（插<span id="ph11370155014182"><a name="ph11370155014182"></a><a name="ph11370155014182"></a>Atlas 300V Pro 视频解析卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul885522434318"></a><a name="ul885522434318"></a><ul id="ul885522434318"><li>CentOS 7.6 for ARM</li><li>CTYunOS 23.01 for ARM</li><li>Kylin V10 SP1 for ARM</li><li>Kylin V10 SP2 for ARM</li><li>Kylin V10 SP3 2403 for ARM</li><li>Kylin V11 for ARM</li><li>openEuler 20.03 for ARM</li><li>openEuler 22.03 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>Ubuntu 20.04 for ARM</li><li>Euler 2.13 for ARM</li><li>Euler 2.15 for ARM</li><li>HCE 2.0.2503 for ARM</li></ul>
</td>
</tr>
<tr id="row374710451637"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p9846181011916"><a name="p9846181011916"></a><a name="p9846181011916"></a><span id="ph032572219310"><a name="ph032572219310"></a><a name="ph032572219310"></a>Atlas 800 推理服务器（型号 3000）</span>（插<span id="ph38461010121913"><a name="ph38461010121913"></a><a name="ph38461010121913"></a>Atlas 300I Duo 推理卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul12521284310"></a><a name="ul12521284310"></a><ul id="ul12521284310"><li>Ubuntu 20.04 for ARM</li><li>Euler 2.12 for ARM</li><li>Euler 2.13 for ARM</li><li>Euler 2.15 for ARM</li><li>Debian 10.2 for ARM</li><li>HCS 8.5.0 (Euler 2.12 ARM) for ARM</li><li>HCE 2.0 for ARM</li><li>HCE 2.0.2503 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>CTYunOS 23.01 for ARM</li><li>BC-Linux_21.10  for ARM</li><li>Kylin V10 SP3 2403 for ARM</li><li>Kylin V11 for ARM</li></ul>
</td>
</tr>
<tr id="row177472451314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p135998206190"><a name="p135998206190"></a><a name="p135998206190"></a><span id="ph14285172325711"><a name="ph14285172325711"></a><a name="ph14285172325711"></a>Atlas 800 推理服务器（型号 3000）</span>（插<span id="ph12599520201912"><a name="ph12599520201912"></a><a name="ph12599520201912"></a>Atlas 300V 视频解析卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul1858503815433"></a><a name="ul1858503815433"></a><ul id="ul1858503815433"><li>openEuler 22.03 for ARM</li><li>Ubuntu 20.04 for ARM</li><li>Euler 2.13 for ARM</li><li>HCE 2.0.2503 for ARM</li></ul>
</td>
</tr>
<tr id="row0274181305"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p929135235814"><a name="p929135235814"></a><a name="p929135235814"></a><span id="ph1622217204595"><a name="ph1622217204595"></a><a name="ph1622217204595"></a>Atlas 800 推理服务器（型号 3000）</span>（插<span id="ph17882195911594"><a name="ph17882195911594"></a><a name="ph17882195911594"></a>Atlas 300I 推理卡（型号 3000）</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul1367818913590"></a><a name="ul1367818913590"></a><ul id="ul1367818913590"><li>CentOS 7.6 for ARM</li><li>Kylin V10 (openEuler) SP1 for ARM</li><li>openEuler 20.03 for ARM</li><li>openEuler 22.03 for ARM</li><li>Ubuntu 20.04 for ARM</li><li>UOS V20 1020e for ARM</li></ul>
</td>
</tr>
<tr id="row1093610269447"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1614017370445"><a name="p1614017370445"></a><a name="p1614017370445"></a><span id="ph1589132515575"><a name="ph1589132515575"></a><a name="ph1589132515575"></a>Atlas 800 推理服务器（型号 3010）</span>（插<span id="ph37641518154514"><a name="ph37641518154514"></a><a name="ph37641518154514"></a>Atlas 300V 视频解析卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul18831155584417"></a><a name="ul18831155584417"></a><ul id="ul18831155584417"><li>openEuler 22.03 for x86</li><li>Ubuntu 20.04 for x86</li></ul>
</td>
</tr>
<tr id="row150520455587"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p2254120609"><a name="p2254120609"></a><a name="p2254120609"></a><span id="ph1469113419015"><a name="ph1469113419015"></a><a name="ph1469113419015"></a>Atlas 800 推理服务器（型号 3010）</span>（插<span id="ph88034120118"><a name="ph88034120118"></a><a name="ph88034120118"></a>Atlas 300I 推理卡（型号 3010）</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul122874255182"></a><a name="ul122874255182"></a><ul id="ul122874255182"><li>CentOS 7.6 for x86</li><li>Kylin V10 SP1 for x86</li><li>openEuler 20.03 for x86</li><li>openEuler 22.03 for x86</li><li>Ubuntu 18.04.1 for x86</li><li>Ubuntu 18.04.5 for x86</li><li>Ubuntu 20.04 for x86</li></ul>
</td>
</tr>
<tr id="row5407193918119"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p21151131101813"><a name="p21151131101813"></a><a name="p21151131101813"></a><span id="ph2689195113175"><a name="ph2689195113175"></a><a name="ph2689195113175"></a>Atlas 800 推理服务器（型号 3010）</span>（插<span id="ph8455182411180"><a name="ph8455182411180"></a><a name="ph8455182411180"></a>Atlas 300I Pro 推理卡</span>）</p>
<p id="p167891541191818"><a name="p167891541191818"></a><a name="p167891541191818"></a><span id="ph18789204113183"><a name="ph18789204113183"></a><a name="ph18789204113183"></a>Atlas 800 推理服务器（型号 3010）</span>（插<span id="ph87621118131917"><a name="ph87621118131917"></a><a name="ph87621118131917"></a>Atlas 300V Pro 视频解析卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul388382719122"></a><a name="ul388382719122"></a><ul id="ul388382719122"><li>CentOS 7.6 for x86</li><li>CTYunOS 23.01 for x86</li><li>Kylin V10 SP1 for x86</li><li>Kylin V10 SP3 2403 for x86</li><li>Kylin V11 for x86</li><li>openEuler 20.03 for x86</li><li>openEuler 22.03 for x86</li><li>openEuler 24.03 LTS SP1 for x86</li><li>Ubuntu 20.04 for x86</li></ul>
</td>
</tr>
<tr id="row1374813451136"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p4468152312519"><a name="p4468152312519"></a><a name="p4468152312519"></a>Atlas 200I SoC A1核心板</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p143050517345"><a name="p143050517345"></a><a name="p143050517345"></a>openEuler 20.03 for ARM</p>
</td>
</tr>
<tr id="row12748144512318"><td class="cellrowborder" rowspan="3" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p19470381257"><a name="p19470381257"></a><a name="p19470381257"></a><span id="ph996833614580"><a name="ph996833614580"></a><a name="ph996833614580"></a><term id="zh-cn_topic_0000001094307702_term99602034117"><a name="zh-cn_topic_0000001094307702_term99602034117"></a><a name="zh-cn_topic_0000001094307702_term99602034117"></a>Atlas A2 推理系列产品</term></span></p>
<p id="p722736189"><a name="p722736189"></a><a name="p722736189"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p159463383518"><a name="p159463383518"></a><a name="p159463383518"></a><span id="ph16179151202"><a name="ph16179151202"></a><a name="ph16179151202"></a>Atlas 800I A2 推理服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul1491814439342"></a><a name="ul1491814439342"></a><ul id="ul1491814439342"><li>BC-Linux-for-Euler-21.10 for ARM</li><li>BC-Linux_21.10 U4  for ARM</li><li>Euler 2.12 for ARM</li><li>Euler 2.13 for ARM</li><li>CentOS 7.6 for ARM</li><li>CTYunOS 22.06 for ARM</li><li>CTYunOS 23.01 for ARM</li><li>Kylin V10 SP2 for ARM</li><li>Kylin V10 SP3 for ARM</li><li>Kylin V11 for ARM</li><li>openEuler 22.03 LTS for ARM</li><li>openEuler 22.03 LTS SP4 for ARM</li><li>openEuler 24.03 LTS SP1 for ARM</li><li>UOS V20 1050u2e for ARM</li><li>Ubuntu 22.04 for ARM</li><li>Ubuntu 24.04 LTS for ARM</li></ul>
</td>
</tr>
<tr id="row4711679201"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1171275208"><a name="p1171275208"></a><a name="p1171275208"></a><span id="ph9484121219204"><a name="ph9484121219204"></a><a name="ph9484121219204"></a>A200I A2 Box 异构组件</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul115263163208"></a><a name="ul115263163208"></a><ul id="ul115263163208"><li>Velinux 1.2 for x86</li><li>Ubuntu 22.04 LTS for x86</li><li>openEuler 22.03 LTS for x86</li></ul>
</td>
</tr>
<tr id="row022717618816"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1927840182513"><a name="p1927840182513"></a><a name="p1927840182513"></a><span id="ph10949202261219"><a name="ph10949202261219"></a><a name="ph10949202261219"></a>Atlas 200I A2 Box16 异构子框</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><a name="ul147601946131620"></a><a name="ul147601946131620"></a><ul id="ul147601946131620"><li>Debian 10.0 for x86</li><li>Debian 11.7 for x86 (kernel 5.10.103)</li><li>Debian 12 for x86 (kernel 5.15.152.ve.10)</li><li>Ubuntu 22.04 for x86</li><li>Ubuntu 20.04.1 for x86</li><li>Ubuntu 22.04.1 for x86 (5.16.20-051620-generic)</li><li>Tlinux 3.1 for x86</li><li>Tlinux 3.2 for x86</li><li>Tlinux 4.0 for x86 (6.6内核)</li><li>openEuler 22.03 LTS SP4 for x86</li><li>openEuler 24.03 LTS  for x86</li><li>openEuler 24.03 LTS SP1 for x86</li></ul>
</td>
</tr>
<tr id="row15438111011218"><td class="cellrowborder" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p168412537204"><a name="p168412537204"></a><a name="p168412537204"></a><span id="ph791742714211"><a name="ph791742714211"></a><a name="ph791742714211"></a><term id="zh-cn_topic_0000001519959665_term176419491615"><a name="zh-cn_topic_0000001519959665_term176419491615"></a><a name="zh-cn_topic_0000001519959665_term176419491615"></a>Atlas A3 推理系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p143871017213"><a name="p143871017213"></a><a name="p143871017213"></a><span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 800I A3 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><a name="ul877545583412"></a><a name="ul877545583412"></a><ul id="ul877545583412"><li>openEuler 22.03 LTS SP4 for ARM</li><li>Euler 2.13 for ARM</li><li>CUlinux 3.0 for ARM</li><li>HCE 2.0.2506 for ARM</li><li>Velinux 2.0 for ARM</li><li>Kylin V10 SP3 2403 for ARM</li><li>Kylin V11 for ARM</li></ul>
</td>
</tr>
<tr id="row64419316256"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><a name="ul17949191917256"></a><a name="ul17949191917256"></a><ul id="ul17949191917256"><li><strong id="b189491819122517"><a name="b189491819122517"></a><a name="b189491819122517"></a>单机场景下：以下硬件产品，仅支持安装<span id="ph10949131918259"><a name="ph10949131918259"></a><a name="ph10949131918259"></a>Ascend Docker Runtime</span>组件。</strong></li><li><strong id="b094919197251"><a name="b094919197251"></a><a name="b094919197251"></a>集群场景下：以下硬件产品，仅支持安装<span id="ph1294912191256"><a name="ph1294912191256"></a><a name="ph1294912191256"></a>Ascend Docker Runtime</span>、<span id="ph1794971919254"><a name="ph1794971919254"></a><a name="ph1794971919254"></a>Ascend Device Plugin</span>组件。</strong></li></ul>
</td>
</tr>
<tr id="row1774814451831"><td class="cellrowborder" rowspan="4" valign="top" width="15.93%" headers="mcps1.2.4.1.1 "><p id="p19447381516"><a name="p19447381516"></a><a name="p19447381516"></a><span id="ph66631140182316"><a name="ph66631140182316"></a><a name="ph66631140182316"></a><term id="zh-cn_topic_0000001519959665_term169221139190"><a name="zh-cn_topic_0000001519959665_term169221139190"></a><a name="zh-cn_topic_0000001519959665_term169221139190"></a>Atlas 200/300/500 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="33.67%" headers="mcps1.2.4.1.2 "><p id="p18478175248"><a name="p18478175248"></a><a name="p18478175248"></a><span id="ph847101717244"><a name="ph847101717244"></a><a name="ph847101717244"></a>Atlas 200 AI加速模块（RC场景）</span></p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="50.4%" headers="mcps1.2.4.1.3 "><p id="p16081439161317"><a name="p16081439161317"></a><a name="p16081439161317"></a>支持的操作系统以硬件产品本身为准。</p>
</td>
</tr>
<tr id="row5748104517319"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p16935163818514"><a name="p16935163818514"></a><a name="p16935163818514"></a><span id="ph1170032414244"><a name="ph1170032414244"></a><a name="ph1170032414244"></a>Atlas 300I 推理卡（型号 3000）</span></p>
</td>
</tr>
<tr id="row1074994512314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p14782173122413"><a name="p14782173122413"></a><a name="p14782173122413"></a><span id="ph1878213182417"><a name="ph1878213182417"></a><a name="ph1878213182417"></a>Atlas 300I 推理卡（型号 3010）</span></p>
</td>
</tr>
<tr id="row137496452314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p129244381458"><a name="p129244381458"></a><a name="p129244381458"></a><span id="ph11405839112417"><a name="ph11405839112417"></a><a name="ph11405839112417"></a>Atlas 500 智能小站（型号 3000）</span></p>
</td>
</tr>
<tr id="row197496451832"><td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.1 "><p id="p69213381451"><a name="p69213381451"></a><a name="p69213381451"></a><span id="ph17875123113012"><a name="ph17875123113012"></a><a name="ph17875123113012"></a><term id="zh-cn_topic_0000001519959665_term7466858493"><a name="zh-cn_topic_0000001519959665_term7466858493"></a><a name="zh-cn_topic_0000001519959665_term7466858493"></a>Atlas 200I/500 A2 推理产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p31781131153019"><a name="p31781131153019"></a><a name="p31781131153019"></a><span id="ph8178173117309"><a name="ph8178173117309"></a><a name="ph8178173117309"></a>Atlas 200I A2 加速模块</span></p>
</td>
</tr>
<tr id="row12749144516310"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1191393818517"><a name="p1191393818517"></a><a name="p1191393818517"></a><span id="ph374453873014"><a name="ph374453873014"></a><a name="ph374453873014"></a>Atlas 200I DK A2 开发者套件</span></p>
</td>
</tr>
<tr id="row1774913451236"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p17935174814307"><a name="p17935174814307"></a><a name="p17935174814307"></a><span id="ph99355488303"><a name="ph99355488303"></a><a name="ph99355488303"></a>Atlas 500 A2 智能小站</span></p>
</td>
</tr>
</tbody>
</table>

# 环境依赖<a name="ZH-CN_TOPIC_0000002511346315"></a>




## 软件依赖<a name="ZH-CN_TOPIC_0000002479226378"></a>

**Ascend Docker Runtime<a name="section14779174114012"></a>**

-   当前环境的Docker版本需要为18.09及以上版本。
-   宿主机已安装驱动和固件，详情请参见《CANN 软件安装指南》中的“<a href="https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian">安装NPU驱动和固件</a>”章节（商用版）或“<a href="https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=openEuler">安装NPU驱动和固件</a>”章节（社区版）。
-   Atlas 500 A2 智能小站安装Ascend Docker Runtime需要修改Docker配置。执行**vi /etc/sysconfig/docker**命令，将--config-file=""参数删除；并执行**systemctl restart docker**使配置生效。
-   Atlas 500 A2 智能小站预置的MEF服务会对Docker进行安全加固配置，Ascend Docker Runtime不支持在安全加固后的Docker环境下使用。若需要使用Ascend Docker Runtime，请手动卸载MEF服务，参考《MindEdge Framework  用户指南》中的“[卸载MEF Edge](https://www.hiascend.com/document/detail/zh/mindedge/730/mef/mefug/mefug_0034.html)”章节进行操作。

    >[!NOTE] 说明 
    >执行**systemctl status docker**命令，如果返回信息里包含“/docker\_entrypoint.sh”字段，则为MEF服务安全加固后的Docker。

**其他集群调度组件<a name="section172351929104018"></a>**

ARM架构和x86\_64架构对应的依赖不一样，请根据系统架构选择。集群调度组件支持IPv4和IPv6，默认使用IPv4。

**表 1**  软件环境

<a name="table20235172944010"></a>
|软件名称|支持的版本|安装位置|说明|
|--|--|--|--|
|（可选）Kubernetes|1.17.x~1.34.x（推荐使用1.19.x及以上版本）<ul><li>建议选择最新的bugfix版本。</li><li>如需安装Volcano组件，请安装1.19.x及以上版本的Kubernetes，具体Kubernetes版本请参见<a href="https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility">Volcano官网中对应的Kubernetes版本</a>。</li></ul>|所有节点|了解K8s的使用请参见<a href="https://kubernetes.io/zh-cn/docs/">Kubernetes文档</a>。|
|（可选）Docker|18.09.x~28.5.1|所有节点|可从<a href="https://docs.docker.com/engine/install/">Docker社区或官网</a>获取。使用的Docker版本需要与Kubernetes配套，配套关系可参考Kubernetes的<a href="https://github.com/kubernetes/kubernetes/tree/master/CHANGELOG">说明</a>，或者从Kubernetes社区获取。建议选择最新的bugfix版本。|
|（可选）Containerd|1.4.x~2.1.4（推荐使用1.6.x版本）|所有节点|可从Containerd的<a href="https://containerd.io/downloads/">官网</a>或者<a href="https://github.com/containerd/containerd/blob/main/docs/getting-started.md#installing-containerd">社区</a>获取，建议选择最新的bugfix版本。请关注配套Kubernetes使用的<a href="https://kubernetes.io/zh-cn/docs/setup/production-environment/container-runtimes/#cri-versions">CRI接口版本</a>。|
|昇腾AI处理器驱动和固件|请参见<a href="https://support.huawei.com/enterprise/zh/ascend-computing/ascend-training-solution-pid-258915853/software">版本配套表</a>（训练）或<a href="https://support.huawei.com/enterprise/zh/ascend-computing/ascend-inference-solution-pid-258915651/software">版本配套表</a>（推理），根据实际硬件设备型号选择与MindCluster配套的驱动、固件。|计算节点|请参见各硬件产品中<a href="https://support.huawei.com/enterprise/zh/ascend-computing/ascend-hdk-pid-252764743">驱动和固件安装升级指南</a>获取对应的指导。<p>为保证NPU Exporter以二进制部署时可使用非root用户安装（如hwMindX），请在安装驱动时使用--install-for-all参数。示例如下。</p><pre class="screen">./Ascend-hdk-&lt;chip_type&gt;-npu-driver_&lt;version&gt;_linux-&lt;arch&gt;.run --full --install-for-all</pre>|
|（可选）CANN|只安装集群调度组件的情况下可不安装CANN，用户可根据实际需要选择安装所需的CANN软件包，可参见版本配套表安装对应的软件包。|计算节点或者训练推理容器内|在宿主机上安装CANN软件包，请参见《<a href="https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0000.html?Mode=PmIns&InstallType=netconda&OS=openEuler">CANN 软件安装指南</a>》（商用版）或《<a href="https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0000.html?Mode=PmIns&InstallType=netconda&OS=openEuler">CANN 软件安装指南件</a>》（社区版）。|
|Python|3.8~3.12|训练或推理容器内|使用时Python版本请以具体AI框架为准。|

>[!NOTE] 说明 
>-   请根据业务的实际使用场景，选择安装Docker或者Containerd。
>-   Atlas 服务器产品安装操作系统可以参见[安装指导书](https://support.huawei.com/enterprise/zh/ascend-computing/a800-9000-pid-250702818?category=installation-upgrade&subcategory=software-deployment-guide)（ARM）和[安装指导书](https://support.huawei.com/enterprise/zh/ascend-computing/a800-9010-pid-250702809?category=installation-upgrade&subcategory=software-deployment-guide)（x86\_64），安装指导书并不包含上述所有操作系统，仅供参考。
>-   Atlas A2 训练系列产品在虚拟机场景下对操作系统的要求不同，具体的操作系统约束请参见《Atlas A2 中心推理和训练硬件 25.0.RC1 NPU驱动和固件安装指南》中的“[虚拟机安装与卸载](https://support.huawei.com/enterprise/zh/doc/EDOC1100468900/cb91d9dc)”章节。

## 组网要求<a name="ZH-CN_TOPIC_0000002479386452"></a>

由于集群调度的核心调度组件Volcano目前是部署在K8s（即Kubernetes）的管理节点，为保证业务健康稳定，部署管理节点根据K8s的部署要求作出如下建议，客户可根据自身业务特点作出调整。

-   管理节点与计算节点、存储节点分离，建议使用单独服务器部署。
-   若集群规模较大或者对业务可靠性要求较高，管理节点需使用多节点方式。

**部署逻辑示意图<a name="section10677192773320"></a>**

**图 1**  部署逻辑示意图<a name="zh-cn_topic_0000001382921066_fig1081627298"></a>  
![](../figures/scheduling/部署逻辑示意图-3.png "部署逻辑示意图")

数据中心集群中的节点类型一般分为以下三种：

-   管理节点（即Master节点）：管理集群，负责分发训练、推理任务到各个计算节点执行，可安装与Master节点相关联的集群调度组件。
-   计算节点（即Worker节点）：实际执行训练和推理任务，可安装与Worker节点相关联的集群调度组件。
-   存储节点：存储数据集、训练输出的模型等数据。

用户需要将网络平面划分为：

-   业务面：用于K8s集群业务管理。
-   存储面：用于从存储节点读取训练用的数据集。因为对带宽有要求，所以建议使用单独的网络平面和网络端口，将训练节点（管理节点或计算节点）和存储节点连通。
-   参数面：用于分布式训练时训练节点之间的参数交换，可参考以下组网说明。
    -   《[Ascend Training Solution 23.0.RC1 组网指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100302398/3a822881)》：提供华为训练计算设备（包括Atlas 800 训练服务器、Atlas 900 PoD（型号 9000）等）搭建组网的相关说明。
    -   [《Ascend Training Solution 25.0.RC1 组网指南（Atlas A2训练产品）》](https://support.huawei.com/enterprise/zh/doc/EDOC1100471246?idPath=23710424|251366513|22892968|252309113|258915853)：提供华为训练计算设备（包括Atlas 800T A2 训练服务器、Atlas 900 A2 PoD 集群基础单元、集成Atlas 200T A2 Box16 异构子框的训练服务器）搭建组网的相关说明。

## 软硬件规格要求<a name="ZH-CN_TOPIC_0000002479386424"></a>

**操作系统磁盘分区<a name="section13457101811533"></a>**

操作系统磁盘分区推荐如[表1](#table147711423499)所示。

**表 1**  磁盘空间规划

<a name="table147711423499"></a>
|分区|说明|大小|启动标志|
|--|--|--|--|
|/boot|启动分区。|500 MB|on|
|/var|软件运行所产生的数据存放分区，如日志、缓存等。|> 300 GB|off|
|/var/lib/docker|Docker镜像与容器存放分区。<p>Docker镜像和容器默认放在/var/lib/docker分区下，如果/var/lib/docker分区使用率大于85%，K8s会启动资源驱逐机制，使用时请确保/var/lib/docker分区使用率在85%以下。</p>|> 300 GB|off|
|/etc/mindx-dl|该分区会存放导入的证书、KubeConfig等文件。建议配置100MB，可根据实际情况调整。|100 MB|off|
|/|主分区。|> 300 GB|off|

**硬件规格要求<a name="section8991121132815"></a>**

硬件产品需要满足如下要求。

**表 2**  资源要求

<a name="table292311420386"></a>
|名称|要求|
|--|--|
|CPU|管理节点CPU＞32核|
|内存|管理节点内存＞64GB|
|磁盘空间|＞1TB磁盘空间规划请参见<a href="#table147711423499">表1</a>|
|网络|<ul><li>带外管理（BMC）：≥1Gbit/s</li><li>带内管理（SSH）：≥1Gbit/s</li><li>业务面：≥10Gbit/s</li><li>存储面：≥25Gbit/s</li><li>参数面：100Gbit/s或200Gbit/s</li></ul>|

**集群调度组件资源配置要求<a name="section168471642185618"></a>**

集群调度组件资源配置需要满足如下要求：

**表 3**  管理节点组件资源配置要求

<a name="table1259491717587"></a>
<table><thead align="left"><tr id="row13594171755810"><th class="cellrowborder" rowspan="2" valign="top" id="mcps1.2.8.1.1"><p id="p16594201795815"><a name="p16594201795815"></a><a name="p16594201795815"></a>组件名称</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.2"><p id="p115944179581"><a name="p115944179581"></a><a name="p115944179581"></a>100节点以内</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.3"><p id="p12594217115811"><a name="p12594217115811"></a><a name="p12594217115811"></a>500节点</p>
</th>
<th class="cellrowborder" colspan="2" valign="top" id="mcps1.2.8.1.4"><p id="p1293144765912"><a name="p1293144765912"></a><a name="p1293144765912"></a>1000节点</p>
</th>
</tr>
<tr id="row124371832162218"><th class="cellrowborder" valign="top" id="mcps1.2.8.2.1"><p id="p243733210220"><a name="p243733210220"></a><a name="p243733210220"></a>CPU（单位：核）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.2"><p id="p1437193212222"><a name="p1437193212222"></a><a name="p1437193212222"></a>内存（单位：GB）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.3"><p id="p543773216224"><a name="p543773216224"></a><a name="p543773216224"></a>CPU（单位：核）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.4"><p id="p1260710142233"><a name="p1260710142233"></a><a name="p1260710142233"></a>内存（单位：GB）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.5"><p id="p54375321221"><a name="p54375321221"></a><a name="p54375321221"></a>CPU（单位：核）</p>
</th>
<th class="cellrowborder" valign="top" id="mcps1.2.8.2.6"><p id="p1965417425238"><a name="p1965417425238"></a><a name="p1965417425238"></a>内存（单位：GB）</p>
</th>
</tr>
</thead>
<tbody><tr id="row10594191713589"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p75941817155816"><a name="p75941817155816"></a><a name="p75941817155816"></a><span id="ph525518226126"><a name="ph525518226126"></a><a name="ph525518226126"></a>Volcano</span> Scheduler</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p16594161785820"><a name="p16594161785820"></a><a name="p16594161785820"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p75661624202111"><a name="p75661624202111"></a><a name="p75661624202111"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p359415171584"><a name="p359415171584"></a><a name="p359415171584"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p960771462311"><a name="p960771462311"></a><a name="p960771462311"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p5931164705914"><a name="p5931164705914"></a><a name="p5931164705914"></a>5.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p14654342142315"><a name="p14654342142315"></a><a name="p14654342142315"></a>8</p>
</td>
</tr>
<tr id="row859401719586"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p185941517145811"><a name="p185941517145811"></a><a name="p185941517145811"></a><span id="ph154111939182412"><a name="ph154111939182412"></a><a name="ph154111939182412"></a>Volcano</span> Controller</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p19594101705811"><a name="p19594101705811"></a><a name="p19594101705811"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p18566172414211"><a name="p18566172414211"></a><a name="p18566172414211"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p15941817195810"><a name="p15941817195810"></a><a name="p15941817195810"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p19607101417238"><a name="p19607101417238"></a><a name="p19607101417238"></a>3</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p9931144725915"><a name="p9931144725915"></a><a name="p9931144725915"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p10654542122312"><a name="p10654542122312"></a><a name="p10654542122312"></a>4</p>
</td>
</tr>
<tr id="row19828191113591"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p182871175916"><a name="p182871175916"></a><a name="p182871175916"></a>Ascend Operator</p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p2082881195911"><a name="p2082881195911"></a><a name="p2082881195911"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p12567122472117"><a name="p12567122472117"></a><a name="p12567122472117"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p15828111145917"><a name="p15828111145917"></a><a name="p15828111145917"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p460741414230"><a name="p460741414230"></a><a name="p460741414230"></a>3</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p1993154735916"><a name="p1993154735916"></a><a name="p1993154735916"></a>2.5</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p265434215235"><a name="p265434215235"></a><a name="p265434215235"></a>4</p>
</td>
</tr>
<tr id="row138951522135910"><td class="cellrowborder" valign="top" width="14.34%" headers="mcps1.2.8.1.1 mcps1.2.8.2.1 "><p id="p9895152214593"><a name="p9895152214593"></a><a name="p9895152214593"></a><span id="ph189871659101117"><a name="ph189871659101117"></a><a name="ph189871659101117"></a>ClusterD</span></p>
</td>
<td class="cellrowborder" valign="top" width="14.729999999999999%" headers="mcps1.2.8.1.2 mcps1.2.8.2.2 "><p id="p10895152216594"><a name="p10895152216594"></a><a name="p10895152216594"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="14.05%" headers="mcps1.2.8.1.2 mcps1.2.8.2.3 "><p id="p175671244212"><a name="p175671244212"></a><a name="p175671244212"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.3 mcps1.2.8.2.4 "><p id="p158951922165910"><a name="p158951922165910"></a><a name="p158951922165910"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="13.780000000000001%" headers="mcps1.2.8.1.3 mcps1.2.8.2.5 "><p id="p1460717147234"><a name="p1460717147234"></a><a name="p1460717147234"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="14.549999999999999%" headers="mcps1.2.8.1.4 mcps1.2.8.2.6 "><p id="p129318474598"><a name="p129318474598"></a><a name="p129318474598"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="14.000000000000002%" headers="mcps1.2.8.1.4 "><p id="p1265416427235"><a name="p1265416427235"></a><a name="p1265416427235"></a>8</p>
</td>
</tr>
</tbody>
</table>

**表 4**  计算节点组件资源配置要求

<a name="table8522160193317"></a>
|组件名称|CPU（单位：核）|内存（单位：GB）|
|--|--|--|
|Ascend Device Plugin|0.5|0.5|
|NodeD|0.5|0.3|
|NPU Exporter|1|1|
|Ascend Docker Runtime|Docker的业务插件，无需单独的CPU和内存空间|

# 准备安装环境<a name="ZH-CN_TOPIC_0000002479386402"></a>

**安装Kubernetes须知<a name="section1188585815256"></a>**

-   Kubernetes使用Calico作为集群网络插件时，默认使用node-to-node mesh的网络配置；当集群规模较大时，该配置可能造成业务交换机网络负载过大，建议配置成reflector模式，具体操作请参考[Calico官方文档](https://docs.tigera.io/calico-enterprise/latest/networking/configuring/bgp#disable-the-default-bgp-node-to-node-mesh)。
-   在CentOS  7.6系统上安装Kubernetes，并且使用v3.24版本Calico作为集群网络插件，可能会安装失败，可参考[系统要求](https://docs.tigera.io/calico/3.24/getting-started/kubernetes/requirements)查看相关约束。
-   Kubernetes  1.24及以上版本，Dockershim已从Kubernetes项目中移除。如果用户还想继续使用Docker作为Kubernetes的容器引擎，需要再安装cri-dockerd，可参考[使用1.24及以上版本的Kubernetes时，使用Docker失败](./faq.md#使用124及以上版本的kubernetes时使用docker失败)章节进行操作。
-   Kubernetes  1.25.10及以上版本，不支持虚拟化的vNPU的恢复使能功能。

**安装开源系统<a name="section1849016313266"></a>**

在安装集群调度组件前，用户需确保完成以下基础环境的准备：

-   安装Docker，支持18.09.x\~28.5.1版本，具体操作请参见[安装Docker](https://docs.docker.com/engine/install/)。
-   安装Containerd，支持1.4.x\~2.1.4版本，具体操作请参见[安装Containerd](https://github.com/containerd/containerd/blob/main/docs/getting-started.md)。
-   安装Kubernetes，支持1.17.x\~1.34.x版本的Kubernetes（推荐使用1.19.x及以上版本），具体操作请参见[安装Kubernetes](https://kubernetes.io/zh/docs/setup/production-environment/tools/)推荐[使用Kubeadm创建集群](https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)，集群初始化过程中的部分问题可参考[初始化Kubernetes失败](./faq.md#初始化kubernetes失败)。且需解除管理节点隔离。如需解除管理节点隔离，命令示例如下。
    -   解除单节点隔离。

        ```
        kubectl taint nodes <hostname> node-role.kubernetes.io/master-
        ```

    -   解除所有节点隔离。

        ```
        kubectl taint nodes --all node-role.kubernetes.io/master-
        ```

        >[!NOTE] 说明 
        >通过解除管理节点隔离可移除主节点的污点，以允许Pod被调度到主节点上。

# 安装部署<a name="ZH-CN_TOPIC_0000002511346373"></a>



## 手动安装<a name="ZH-CN_TOPIC_0000002511426383"></a>












### 获取软件包<a name="ZH-CN_TOPIC_0000002479386476"></a>

获取相应的软件可参见[下载软件包](#section10979172103311)；获取相应软件包的源码可参见[开源组件源码](#section149534517468)进行操作。

**下载软件包<a name="section10979172103311"></a>**

下载本软件即表示您同意[华为企业业务最终用户许可协议（EULA）](https://e.huawei.com/cn/about/eula)的条款和条件。

>[!NOTE] 说明 
><i>\{version\}</i>表示软件版本号，<i>\{arch\}</i>表示CPU架构。

**表 1**  各组件软件包

<a name="table13465342493"></a>
<table><thead align="left"><tr id="row64656424913"><th class="cellrowborder" valign="top" width="30.440000000000005%" id="mcps1.2.5.1.1"><p id="p16465154174918"><a name="p16465154174918"></a><a name="p16465154174918"></a>组件名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.310000000000002%" id="mcps1.2.5.1.2"><p id="p11465249493"><a name="p11465249493"></a><a name="p11465249493"></a>包内文件列表</p>
</th>
<th class="cellrowborder" valign="top" width="35.25%" id="mcps1.2.5.1.3"><p id="p146504134919"><a name="p146504134919"></a><a name="p146504134919"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="10.000000000000002%" id="mcps1.2.5.1.4"><p id="p6780144171019"><a name="p6780144171019"></a><a name="p6780144171019"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row10954194381716"><td class="cellrowborder" rowspan="10" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p16445379177"><a name="p16445379177"></a><a name="p16445379177"></a><span id="ph54404369331"><a name="ph54404369331"></a><a name="ph54404369331"></a>Ascend Docker Runtime</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1264443711175"><a name="p1264443711175"></a><a name="p1264443711175"></a>ascend-docker-cli</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p18270161061517"><a name="p18270161061517"></a><a name="p18270161061517"></a><span id="ph622819010286"><a name="ph622819010286"></a><a name="ph622819010286"></a>Ascend Docker Runtime</span>运行所必需的可执行程序，不建议用户直接运行。</p>
</td>
<td class="cellrowborder" rowspan="10" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p1620912301185"><a name="p1620912301185"></a><a name="p1620912301185"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row13637104114174"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6945174713815"><a name="p6945174713815"></a><a name="p6945174713815"></a>ascend-docker-destroy</p>
</td>
</tr>
<tr id="row81412392177"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1128015545814"><a name="p1128015545814"></a><a name="p1128015545814"></a>ascend-docker-hook</p>
</td>
</tr>
<tr id="row1402153611178"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1368012586816"><a name="p1368012586816"></a><a name="p1368012586816"></a>ascend-docker-plugin-install-helper</p>
</td>
</tr>
<tr id="row2070753441718"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p57693341196"><a name="p57693341196"></a><a name="p57693341196"></a>ascend-docker-runtime</p>
</td>
</tr>
<tr id="row0100183371713"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p890412183216"><a name="p890412183216"></a><a name="p890412183216"></a>assets</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p990482113216"><a name="p990482113216"></a><a name="p990482113216"></a>说明资料的图片资源。</p>
</td>
</tr>
<tr id="row19558133191710"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11904202115328"><a name="p11904202115328"></a><a name="p11904202115328"></a>base.list*</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1490462112323"><a name="p1490462112323"></a><a name="p1490462112323"></a>默认的挂载列表，安装时，程序会根据install-type，安装不同的挂载列表。</p>
</td>
</tr>
<tr id="row108361029181712"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p4904142119321"><a name="p4904142119321"></a><a name="p4904142119321"></a>run_main.sh</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1190482112322"><a name="p1190482112322"></a><a name="p1190482112322"></a>安装脚本，不建议用户直接使用。</p>
</td>
</tr>
<tr id="row930912871717"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p4904721113214"><a name="p4904721113214"></a><a name="p4904721113214"></a>uninstall.sh</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p6904721183220"><a name="p6904721183220"></a><a name="p6904721183220"></a>卸载脚本，不建议用户直接使用。</p>
</td>
</tr>
<tr id="row19769725201719"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p3904202153210"><a name="p3904202153210"></a><a name="p3904202153210"></a>README.md</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p14904321133219"><a name="p14904321133219"></a><a name="p14904321133219"></a><span id="ph179041721123211"><a name="ph179041721123211"></a><a name="ph179041721123211"></a>Ascend Docker Runtime</span>说明资料，包含设计原理。</p>
</td>
</tr>
<tr id="row103211151198"><td class="cellrowborder" rowspan="8" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p8202152511199"><a name="p8202152511199"></a><a name="p8202152511199"></a><span id="ph19685554163314"><a name="ph19685554163314"></a><a name="ph19685554163314"></a>NPU Exporter</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p815122517194"><a name="p815122517194"></a><a name="p815122517194"></a>npu-exporter</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p1415172514193"><a name="p1415172514193"></a><a name="p1415172514193"></a><span id="ph11151122561916"><a name="ph11151122561916"></a><a name="ph11151122561916"></a>NPU Exporter</span>二进制文件。</p>
</td>
<td class="cellrowborder" rowspan="8" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p8603637171910"><a name="p8603637171910"></a><a name="p8603637171910"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row89908641918"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p181511250198"><a name="p181511250198"></a><a name="p181511250198"></a><span id="ph61511425111914"><a name="ph61511425111914"></a><a name="ph61511425111914"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1915116253194"><a name="p1915116253194"></a><a name="p1915116253194"></a><span id="ph1415116253194"><a name="ph1415116253194"></a><a name="ph1415116253194"></a>NPU Exporter</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row25751084192"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9151132581916"><a name="p9151132581916"></a><a name="p9151132581916"></a>Dockerfile-310P-1usoc</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1815192519196"><a name="p1815192519196"></a><a name="p1815192519196"></a><span id="ph19151102512197"><a name="ph19151102512197"></a><a name="ph19151102512197"></a>Atlas 200I SoC A1 核心板</span>上<span id="ph7151725131914"><a name="ph7151725131914"></a><a name="ph7151725131914"></a>NPU Exporter</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row698299141912"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p14151325131910"><a name="p14151325131910"></a><a name="p14151325131910"></a>run_for_310P_1usoc.sh</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1515118259195"><a name="p1515118259195"></a><a name="p1515118259195"></a><span id="ph191511825101916"><a name="ph191511825101916"></a><a name="ph191511825101916"></a>Atlas 200I SoC A1 核心板</span>上<span id="ph13152182518199"><a name="ph13152182518199"></a><a name="ph13152182518199"></a>NPU Exporter</span>镜像中启动组件的脚本。</p>
</td>
</tr>
<tr id="row20440611191912"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p14152625161915"><a name="p14152625161915"></a><a name="p14152625161915"></a>npu-exporter-v<em id="i715216259199"><a name="i715216259199"></a><a name="i715216259199"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p21521025131920"><a name="p21521025131920"></a><a name="p21521025131920"></a><span id="ph315212252196"><a name="ph315212252196"></a><a name="ph315212252196"></a>NPU Exporter</span>的启动配置文件。</p>
</td>
</tr>
<tr id="row122001013141917"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p181521425201919"><a name="p181521425201919"></a><a name="p181521425201919"></a>npu-exporter-310P-1usoc-<em id="i115218256198"><a name="i115218256198"></a><a name="i115218256198"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1815242581918"><a name="p1815242581918"></a><a name="p1815242581918"></a><span id="ph0152425111919"><a name="ph0152425111919"></a><a name="ph0152425111919"></a>Atlas 200I SoC A1 核心板</span>上<span id="ph131521125161915"><a name="ph131521125161915"></a><a name="ph131521125161915"></a>NPU Exporter</span>的启动配置文件。</p>
</td>
</tr>
<tr id="row1070615221286"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p586243615816"><a name="p586243615816"></a><a name="p586243615816"></a>metricConfiguration.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p57072022483"><a name="p57072022483"></a><a name="p57072022483"></a>默认指标组配置文件。</p>
</td>
</tr>
<tr id="row179079245810"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p144987432813"><a name="p144987432813"></a><a name="p144987432813"></a>pluginConfiguration.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2907182417816"><a name="p2907182417816"></a><a name="p2907182417816"></a>自定义指标组配置文件。</p>
</td>
</tr>
<tr id="row10465947499"><td class="cellrowborder" rowspan="16" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p349233264112"><a name="p349233264112"></a><a name="p349233264112"></a><span id="ph522114212719"><a name="ph522114212719"></a><a name="ph522114212719"></a>Ascend Device Plugin</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p246510417491"><a name="p246510417491"></a><a name="p246510417491"></a>device-plugin</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p34651145494"><a name="p34651145494"></a><a name="p34651145494"></a><span id="ph1024012311247"><a name="ph1024012311247"></a><a name="ph1024012311247"></a>Ascend Device Plugin</span>二进制文件。</p>
</td>
<td class="cellrowborder" rowspan="16" align="left" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p578044431010"><a name="p578044431010"></a><a name="p578044431010"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<p id="p17810448103"><a name="p17810448103"></a><a name="p17810448103"></a></p>
<p id="p178144410101"><a name="p178144410101"></a><a name="p178144410101"></a></p>
</td>
</tr>
<tr id="row6309142253316"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p18649310173610"><a name="p18649310173610"></a><a name="p18649310173610"></a><span id="ph945392991719"><a name="ph945392991719"></a><a name="ph945392991719"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p86498100367"><a name="p86498100367"></a><a name="p86498100367"></a><span id="ph59665356281"><a name="ph59665356281"></a><a name="ph59665356281"></a>Ascend Device Plugin</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row1792262132218"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1892218210228"><a name="p1892218210228"></a><a name="p1892218210228"></a>Dockerfile-310P-1usoc</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p5922721122210"><a name="p5922721122210"></a><a name="p5922721122210"></a><span id="ph2783189473"><a name="ph2783189473"></a><a name="ph2783189473"></a>Atlas 200I SoC A1 核心板</span>上<span id="ph758710382283"><a name="ph758710382283"></a><a name="ph758710382283"></a>Ascend Device Plugin</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row9471113116227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p64721131142217"><a name="p64721131142217"></a><a name="p64721131142217"></a>run_for_310P_1usoc.sh</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p44721931152213"><a name="p44721931152213"></a><a name="p44721931152213"></a><span id="ph158291210134715"><a name="ph158291210134715"></a><a name="ph158291210134715"></a>Atlas 200I SoC A1 核心板</span>上<span id="ph1168811404280"><a name="ph1168811404280"></a><a name="ph1168811404280"></a>Ascend Device Plugin</span>镜像中启动组件的脚本。</p>
</td>
</tr>
<tr id="row552132183519"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p12522320359"><a name="p12522320359"></a><a name="p12522320359"></a>faultCode.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p252193233517"><a name="p252193233517"></a><a name="p252193233517"></a>记录芯片故障码与其故障恢复方式的对应关系。</p>
<div class="notice" id="note44814184129"><a name="note44814184129"></a><a name="note44814184129"></a><span class="noticetitle"> 须知： </span><div class="noticebody"><p id="p104819187122"><a name="p104819187122"></a><a name="p104819187122"></a>系统配置文件，请勿随意修改，否则可能会导致系统故障处理功能出错。</p>
</div></div>
</td>
</tr>
<tr id="row889915010197"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p148996504199"><a name="p148996504199"></a><a name="p148996504199"></a>SwitchFaultCode.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1017559104813"><a name="p1017559104813"></a><a name="p1017559104813"></a>记录灵衢总线设备故障码与其故障恢复方式的对应关系。</p>
<div class="notice" id="note111761099487"><a name="note111761099487"></a><a name="note111761099487"></a><span class="noticetitle"> 须知： </span><div class="noticebody"><p id="p141761695482"><a name="p141761695482"></a><a name="p141761695482"></a>系统配置文件，请勿随意修改，否则可能会导致系统故障处理功能出错。</p>
</div></div>
</td>
</tr>
<tr id="row149041193485"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p5721410142710"><a name="p5721410142710"></a><a name="p5721410142710"></a>faultCustomization.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p07231013278"><a name="p07231013278"></a><a name="p07231013278"></a>芯片故障频率及时长默认配置文件。</p>
<div class="notice" id="note154671054144417"><a name="note154671054144417"></a><a name="note154671054144417"></a><span class="noticetitle"> 须知： </span><div class="noticebody"><p id="p14467125413448"><a name="p14467125413448"></a><a name="p14467125413448"></a>系统配置文件，请勿随意修改，否则可能会导致系统故障处理功能出错。</p>
</div></div>
</td>
</tr>
<tr id="row20130191445811"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p141301014165820"><a name="p141301014165820"></a><a name="p141301014165820"></a>deviceNameCustomization.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2013151475812"><a name="p2013151475812"></a><a name="p2013151475812"></a>自定义设备名称配置文件。</p>
<div class="notice" id="note3110378817"><a name="note3110378817"></a><a name="note3110378817"></a><span class="noticetitle"> 须知： </span><div class="noticebody"><p id="p211163719817"><a name="p211163719817"></a><a name="p211163719817"></a>系统配置文件，请勿随意修改，否则可能会导致系统故障处理、设备纳管功能出错。</p>
</div></div>
</td>
</tr>
<tr id="row15301532123217"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p364901073616"><a name="p364901073616"></a><a name="p364901073616"></a>device-plugin-310-v<em id="i10364134551714"><a name="i10364134551714"></a><a name="i10364134551714"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1032715441283"><a name="p1032715441283"></a><a name="p1032715441283"></a>推理服务器（插<span id="ph163696166292"><a name="ph163696166292"></a><a name="ph163696166292"></a>Atlas 300I 推理卡</span>）上不使用<span id="ph183921109162"><a name="ph183921109162"></a><a name="ph183921109162"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row1317496692"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p18536134319468"><a name="p18536134319468"></a><a name="p18536134319468"></a>device-plugin-310-volcano-<em id="i6536134324615"><a name="i6536134324615"></a><a name="i6536134324615"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11536543194616"><a name="p11536543194616"></a><a name="p11536543194616"></a>推理服务器（插<span id="ph0536443144614"><a name="ph0536443144614"></a><a name="ph0536443144614"></a>Atlas 300I 推理卡</span>）上使用<span id="ph1353694310467"><a name="ph1353694310467"></a><a name="ph1353694310467"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row1100173517329"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p168332200511"><a name="p168332200511"></a><a name="p168332200511"></a>device-plugin-310P-v<em id="i3556348201713"><a name="i3556348201713"></a><a name="i3556348201713"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p075418583017"><a name="p075418583017"></a><a name="p075418583017"></a><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>设备上不使用<span id="ph1868725712201"><a name="ph1868725712201"></a><a name="ph1868725712201"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row6435172313911"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1310865514615"><a name="p1310865514615"></a><a name="p1310865514615"></a>device-plugin-310P-volcano-<em id="i141081555114614"><a name="i141081555114614"></a><a name="i141081555114614"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p610885574617"><a name="p610885574617"></a><a name="p610885574617"></a><span id="ph16108155134618"><a name="ph16108155134618"></a><a name="ph16108155134618"></a>Atlas 推理系列产品</span>设备上使用<span id="ph1110885524616"><a name="ph1110885524616"></a><a name="ph1110885524616"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row349515251394"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p111089552466"><a name="p111089552466"></a><a name="p111089552466"></a>device-plugin-310P-1usoc-<em id="i210811554462"><a name="i210811554462"></a><a name="i210811554462"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p3108105513461"><a name="p3108105513461"></a><a name="p3108105513461"></a><span id="ph141081955204614"><a name="ph141081955204614"></a><a name="ph141081955204614"></a>Atlas 200I SoC A1 核心板</span>上不使用<span id="ph1510945511461"><a name="ph1510945511461"></a><a name="ph1510945511461"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row79263271099"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p138727813478"><a name="p138727813478"></a><a name="p138727813478"></a>device-plugin-310P-1usoc-volcano-v<em id="i1987218834713"><a name="i1987218834713"></a><a name="i1987218834713"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1687216824716"><a name="p1687216824716"></a><a name="p1687216824716"></a><span id="ph18721785476"><a name="ph18721785476"></a><a name="ph18721785476"></a>Atlas 200I SoC A1 核心板</span>上使用<span id="ph787228144715"><a name="ph787228144715"></a><a name="ph787228144715"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row1761512372323"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p3890224115110"><a name="p3890224115110"></a><a name="p3890224115110"></a>device-plugin-910-v<em id="i14676173911711"><a name="i14676173911711"></a><a name="i14676173911711"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p690919143714"><a name="p690919143714"></a><a name="p690919143714"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a>Atlas 训练系列产品</span>或<span id="ph155178916436"><a name="ph155178916436"></a><a name="ph155178916436"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>上不使用<span id="ph12340165912011"><a name="ph12340165912011"></a><a name="ph12340165912011"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row3124540173214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p891215422516"><a name="p891215422516"></a><a name="p891215422516"></a>device-plugin-volcano-<em id="i201964115579"><a name="i201964115579"></a><a name="i201964115579"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p723164185113"><a name="p723164185113"></a><a name="p723164185113"></a><span id="ph5477154313217"><a name="ph5477154313217"></a><a name="ph5477154313217"></a>Atlas 训练系列产品</span>或<span id="ph1962393802018"><a name="ph1962393802018"></a><a name="ph1962393802018"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>上使用<span id="ph988012132113"><a name="ph988012132113"></a><a name="ph988012132113"></a>Volcano</span>的配置文件。</p>
</td>
</tr>
<tr id="row15450141514209"><td class="cellrowborder" rowspan="7" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p10503112473412"><a name="p10503112473412"></a><a name="p10503112473412"></a><span id="ph5971324183414"><a name="ph5971324183414"></a><a name="ph5971324183414"></a>Volcano</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p178531826142015"><a name="p178531826142015"></a><a name="p178531826142015"></a>volcano-npu_<em id="i7853172682017"><a name="i7853172682017"></a><a name="i7853172682017"></a>{version}</em>_linux-<em id="i285417261206"><a name="i285417261206"></a><a name="i285417261206"></a>{arch}.</em>so</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p16854142616202"><a name="p16854142616202"></a><a name="p16854142616202"></a><span id="ph3854152619207"><a name="ph3854152619207"></a><a name="ph3854152619207"></a>Volcano</span>华为NPU调度插件动态链接库。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p18607137230"><a name="p18607137230"></a><a name="p18607137230"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row7321176208"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1585422618209"><a name="p1585422618209"></a><a name="p1585422618209"></a>Dockerfile-scheduler</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p285413268207"><a name="p285413268207"></a><a name="p285413268207"></a>Volcano scheduler镜像构建文本文件。</p>
</td>
</tr>
<tr id="row1095771814208"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p19854132682016"><a name="p19854132682016"></a><a name="p19854132682016"></a>Dockerfile-controller</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1585432622020"><a name="p1585432622020"></a><a name="p1585432622020"></a>Volcano controller镜像构建文本文件。</p>
</td>
</tr>
<tr id="row168141320182019"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10854526102018"><a name="p10854526102018"></a><a name="p10854526102018"></a>volcano-v<em id="i1985411263201"><a name="i1985411263201"></a><a name="i1985411263201"></a>{version}.yaml</em></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p285419263209"><a name="p285419263209"></a><a name="p285419263209"></a><span id="ph1685412619201"><a name="ph1685412619201"></a><a name="ph1685412619201"></a>Volcano</span>的启动配置文件。</p>
</td>
</tr>
<tr id="row5308422172014"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9855202672013"><a name="p9855202672013"></a><a name="p9855202672013"></a>vc-scheduler</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p885532611201"><a name="p885532611201"></a><a name="p885532611201"></a>volcano-scheduler组件二进制文件。</p>
</td>
</tr>
<tr id="row87499237203"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p585552652013"><a name="p585552652013"></a><a name="p585552652013"></a>vc-controller-manager</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1785572611207"><a name="p1785572611207"></a><a name="p1785572611207"></a>volcano-controller组件二进制文件。</p>
</td>
</tr>
<tr id="row8271202512020"><td class="cellrowborder" colspan="2" valign="top" headers="mcps1.2.5.1.2 mcps1.2.5.1.3 "><div class="note" id="note118552026192017"><a name="note118552026192017"></a><a name="note118552026192017"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p89689544619"><a name="p89689544619"></a><a name="p89689544619"></a>请根据<span id="ph1280481810116"><a name="ph1280481810116"></a><a name="ph1280481810116"></a>K8s</span>和开源<span id="ph145451647133910"><a name="ph145451647133910"></a><a name="ph145451647133910"></a>Volcano</span>的兼容性选择合适的版本进行安装<span id="ph14710141210128"><a name="ph14710141210128"></a><a name="ph14710141210128"></a>，具体</span><span id="ph271071215129"><a name="ph271071215129"></a><a name="ph271071215129"></a>版本请参见</span><a href="https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility" target="_blank" rel="noopener noreferrer">Volcano官网中对应的Kubernetes版本</a>。</p>
<a name="ul422518573615"></a><a name="ul422518573615"></a><ul id="ul422518573615"><li><span id="ph148661501863"><a name="ph148661501863"></a><a name="ph148661501863"></a>Volcano</span> v1.7.0兼容的<span id="ph11728942124314"><a name="ph11728942124314"></a><a name="ph11728942124314"></a>K8s</span>版本范围为1.19.x~1.28.x。</li><li><span id="ph1945014372610"><a name="ph1945014372610"></a><a name="ph1945014372610"></a>Volcano</span> v1.9.0兼容的<span id="ph127281742154313"><a name="ph127281742154313"></a><a name="ph127281742154313"></a>K8s</span>版本范围为1.21.x~1.28.x。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1823125217"><td class="cellrowborder" rowspan="3" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p20784133583419"><a name="p20784133583419"></a><a name="p20784133583419"></a><span id="ph12208193653414"><a name="ph12208193653414"></a><a name="ph12208193653414"></a>Ascend Operator</span></p>
<p id="p958252015218"><a name="p958252015218"></a><a name="p958252015218"></a></p>
<p id="p145801320142115"><a name="p145801320142115"></a><a name="p145801320142115"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1355502032117"><a name="p1355502032117"></a><a name="p1355502032117"></a>ascend-operator</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p12556132012216"><a name="p12556132012216"></a><a name="p12556132012216"></a><span id="ph85561520142113"><a name="ph85561520142113"></a><a name="ph85561520142113"></a>Ascend Operator</span>二进制文件。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p166316160318"><a name="p166316160318"></a><a name="p166316160318"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row4950313152111"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6556182042112"><a name="p6556182042112"></a><a name="p6556182042112"></a><span id="ph1655611208214"><a name="ph1655611208214"></a><a name="ph1655611208214"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p17556172062117"><a name="p17556172062117"></a><a name="p17556172062117"></a><span id="ph355672072120"><a name="ph355672072120"></a><a name="ph355672072120"></a>Ascend Operator</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row466481562111"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p655682032116"><a name="p655682032116"></a><a name="p655682032116"></a>ascend-operator-v<em id="i165564208218"><a name="i165564208218"></a><a name="i165564208218"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p18556172072113"><a name="p18556172072113"></a><a name="p18556172072113"></a><span id="ph355672072119"><a name="ph355672072119"></a><a name="ph355672072119"></a>Ascend Operator</span>的启动配置文件。</p>
</td>
</tr>
<tr id="row1221137141018"><td class="cellrowborder" rowspan="7" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p8312184419344"><a name="p8312184419344"></a><a name="p8312184419344"></a><span id="ph573619445343"><a name="ph573619445343"></a><a name="ph573619445343"></a>NodeD</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p14306112011011"><a name="p14306112011011"></a><a name="p14306112011011"></a>noded</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p1430618209109"><a name="p1430618209109"></a><a name="p1430618209109"></a><span id="ph133061920101019"><a name="ph133061920101019"></a><a name="ph133061920101019"></a>NodeD</span>二进制文件。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p2078214442104"><a name="p2078214442104"></a><a name="p2078214442104"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row2527181014104"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1930716203101"><a name="p1930716203101"></a><a name="p1930716203101"></a>noded-v<em id="i1645121221719"><a name="i1645121221719"></a><a name="i1645121221719"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p930719208107"><a name="p930719208107"></a><a name="p930719208107"></a><span id="ph18307820161016"><a name="ph18307820161016"></a><a name="ph18307820161016"></a>NodeD</span>的启动配置文件。</p>
</td>
</tr>
<tr id="row687220595818"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p088099175818"><a name="p088099175818"></a><a name="p088099175818"></a>noded-dpc-v<em id="i15881291587"><a name="i15881291587"></a><a name="i15881291587"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1988119185816"><a name="p1988119185816"></a><a name="p1988119185816"></a>如需使用<a href="./usage/resumable_training.md#节点故障">dpc故障检测功能</a>，使用本配置文件启动<span id="ph143091154131112"><a name="ph143091154131112"></a><a name="ph143091154131112"></a>NodeD</span>。</p>
</td>
</tr>
<tr id="row33201227121219"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p832011273124"><a name="p832011273124"></a><a name="p832011273124"></a>NodeDConfiguration.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p43201727181211"><a name="p43201727181211"></a><a name="p43201727181211"></a>记录节点硬件故障码与其故障恢复方式的对应关系。</p>
</td>
</tr>
<tr id="row149024613565"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11917462565"><a name="p11917462565"></a><a name="p11917462565"></a>pingmesh-config.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p891646165617"><a name="p891646165617"></a><a name="p891646165617"></a>pingmesh配置文件。</p>
</td>
</tr>
<tr id="row91232975612"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17313403583"><a name="p17313403583"></a><a name="p17313403583"></a>fdConfig.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p0311340125819"><a name="p0311340125819"></a><a name="p0311340125819"></a>故障诊断配置文件。</p>
</td>
</tr>
<tr id="row0727191313105"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17307162011106"><a name="p17307162011106"></a><a name="p17307162011106"></a><span id="ph530762013105"><a name="ph530762013105"></a><a name="ph530762013105"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p14307142081012"><a name="p14307142081012"></a><a name="p14307142081012"></a><span id="ph830712203103"><a name="ph830712203103"></a><a name="ph830712203103"></a>NodeD</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row3837449156"><td class="cellrowborder" rowspan="7" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p3777145313419"><a name="p3777145313419"></a><a name="p3777145313419"></a><span id="ph12251185443414"><a name="ph12251185443414"></a><a name="ph12251185443414"></a>ClusterD</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p12435111214612"><a name="p12435111214612"></a><a name="p12435111214612"></a>clusterd</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p1843512121165"><a name="p1843512121165"></a><a name="p1843512121165"></a><span id="ph1043518127610"><a name="ph1043518127610"></a><a name="ph1043518127610"></a>ClusterD</span>二进制文件。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p866312341360"><a name="p866312341360"></a><a name="p866312341360"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row6376165513514"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15435111214610"><a name="p15435111214610"></a><a name="p15435111214610"></a>clusterd-v<em id="i74356121167"><a name="i74356121167"></a><a name="i74356121167"></a>{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1143521216617"><a name="p1143521216617"></a><a name="p1143521216617"></a><span id="ph84351912566"><a name="ph84351912566"></a><a name="ph84351912566"></a>ClusterD</span>的启动配置文件。</p>
</td>
</tr>
<tr id="row5481122714581"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1481727185818"><a name="p1481727185818"></a><a name="p1481727185818"></a>fdConfig.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p94819271584"><a name="p94819271584"></a><a name="p94819271584"></a>故障诊断配置文件。</p>
</td>
</tr>
<tr id="row191165115618"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p124358124611"><a name="p124358124611"></a><a name="p124358124611"></a><span id="ph1435121219617"><a name="ph1435121219617"></a><a name="ph1435121219617"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p6436151215618"><a name="p6436151215618"></a><a name="p6436151215618"></a><span id="ph743671216617"><a name="ph743671216617"></a><a name="ph743671216617"></a>ClusterD</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row1741494011129"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p14414340141218"><a name="p14414340141218"></a><a name="p14414340141218"></a>faultDuration.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p241413408120"><a name="p241413408120"></a><a name="p241413408120"></a>关联故障处理时长配置文件。</p>
</td>
</tr>
<tr id="row185422455125"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1368494941817"><a name="p1368494941817"></a><a name="p1368494941817"></a>relationFaultCustomization.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7542545171219"><a name="p7542545171219"></a><a name="p7542545171219"></a>关联故障处理策略配置文件。</p>
</td>
</tr>
<tr id="row11798195071219"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p177981950151213"><a name="p177981950151213"></a><a name="p177981950151213"></a>publicFaultConfiguration.json</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11798115021217"><a name="p11798115021217"></a><a name="p11798115021217"></a>公共故障配置文件。</p>
</td>
</tr>
<tr id="row662110549571"><td class="cellrowborder" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p947818561570"><a name="p947818561570"></a><a name="p947818561570"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p14478165613571"><a name="p14478165613571"></a><a name="p14478165613571"></a>taskd-{version}-py3-none-linux_{arch}.whl</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p1347835685717"><a name="p1347835685717"></a><a name="p1347835685717"></a>断点续训特性二进制文件。</p>
</td>
<td class="cellrowborder" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p523641511441"><a name="p523641511441"></a><a name="p523641511441"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
<tr id="row1726355217106"><td class="cellrowborder" rowspan="6" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p13571143931110"><a name="p13571143931110"></a><a name="p13571143931110"></a><span id="ph1857163911120"><a name="ph1857163911120"></a><a name="ph1857163911120"></a>Resilience Controller</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p165711539161110"><a name="p165711539161110"></a><a name="p165711539161110"></a>resilience-controller</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p1557119396111"><a name="p1557119396111"></a><a name="p1557119396111"></a><span id="ph05711396119"><a name="ph05711396119"></a><a name="ph05711396119"></a>Resilience Controller</span>二进制文件。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p1557113981113"><a name="p1557113981113"></a><a name="p1557113981113"></a><a href="https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann" target="_blank" rel="noopener noreferrer">获取链接</a></p>
<div class="note" id="note05631653219"><a name="note05631653219"></a><a name="note05631653219"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p17563155317115"><a name="p17563155317115"></a><a name="p17563155317115"></a>7.3.0版本<span id="ph18649516741"><a name="ph18649516741"></a><a name="ph18649516741"></a>Resilience Controller</span>和<span id="ph7366141610619"><a name="ph7366141610619"></a><a name="ph7366141610619"></a>Elastic Agent</span>组件已经日落，请获取7.3.0之前版本的软件包。</p>
</div></div>
</td>
</tr>
<tr id="row0585150141113"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p13571133921117"><a name="p13571133921117"></a><a name="p13571133921117"></a>cert-importer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p85718398117"><a name="p85718398117"></a><a name="p85718398117"></a>证书导入工具二进制文件。</p>
</td>
</tr>
<tr id="row5585601111"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p20571339101114"><a name="p20571339101114"></a><a name="p20571339101114"></a><span id="ph1057117390113"><a name="ph1057117390113"></a><a name="ph1057117390113"></a>Dockerfile</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p4571239131114"><a name="p4571239131114"></a><a name="p4571239131114"></a><span id="ph0571839171112"><a name="ph0571839171112"></a><a name="ph0571839171112"></a>Resilience Controller</span>镜像构建文本文件。</p>
</td>
</tr>
<tr id="row0608866112"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p757143961115"><a name="p757143961115"></a><a name="p757143961115"></a>resilience-controller-<em id="i11571103918111"><a name="i11571103918111"></a><a name="i11571103918111"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p057123918119"><a name="p057123918119"></a><a name="p057123918119"></a><span id="ph1857193951114"><a name="ph1857193951114"></a><a name="ph1857193951114"></a>Resilience Controller</span>的启动配置文件（不需要用户导入KubeConfig文件）。</p>
</td>
</tr>
<tr id="row16086617110"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p13571183931112"><a name="p13571183931112"></a><a name="p13571183931112"></a>resilience-controller-without-token-<em id="i12571039101112"><a name="i12571039101112"></a><a name="i12571039101112"></a>v{version}</em>.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p17571123917113"><a name="p17571123917113"></a><a name="p17571123917113"></a><span id="ph357103971112"><a name="ph357103971112"></a><a name="ph357103971112"></a>Resilience Controller</span>的启动配置文件（需要用户导入KubeConfig文件）。</p>
</td>
</tr>
<tr id="row14608136151118"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p35711939181112"><a name="p35711939181112"></a><a name="p35711939181112"></a>lib</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p16571193918119"><a name="p16571193918119"></a><a name="p16571193918119"></a>加密组件依赖的动态库文件。</p>
</td>
</tr>
<tr id="row1911313791120"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p8571163920116"><a name="p8571163920116"></a><a name="p8571163920116"></a><span id="ph657183941119"><a name="ph657183941119"></a><a name="ph657183941119"></a>Elastic Agent</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p185710391111"><a name="p185710391111"></a><a name="p185710391111"></a>mindx_elastic-<em id="i857133921110"><a name="i857133921110"></a><a name="i857133921110"></a>{version}</em>-py3-none-linux_<em id="i15571153961114"><a name="i15571153961114"></a><a name="i15571153961114"></a>{arch}</em>.whl</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p19571193910118"><a name="p19571193910118"></a><a name="p19571193910118"></a>断点续训特性二进制文件。</p>
</td>
</tr>
<tr id="row1646165211919"><td class="cellrowborder" valign="top" width="30.440000000000005%" headers="mcps1.2.5.1.1 "><p id="p1646275210917"><a name="p1646275210917"></a><a name="p1646275210917"></a><span id="ph65811069214"><a name="ph65811069214"></a><a name="ph65811069214"></a>Container Manager</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p10462352896"><a name="p10462352896"></a><a name="p10462352896"></a>container-manager</p>
</td>
<td class="cellrowborder" valign="top" width="35.25%" headers="mcps1.2.5.1.3 "><p id="p1846275218911"><a name="p1846275218911"></a><a name="p1846275218911"></a><span id="ph762521420387"><a name="ph762521420387"></a><a name="ph762521420387"></a>Container Manager</span>组件二进制文件。</p>
</td>
<td class="cellrowborder" valign="top" width="10.000000000000002%" headers="mcps1.2.5.1.4 "><p id="p12618122171614"><a name="p12618122171614"></a><a name="p12618122171614"></a><a href="https://gitcode.com/Ascend/mind-cluster/releases" target="_blank" rel="noopener noreferrer">获取链接</a></p>
</td>
</tr>
</tbody>
</table>


**软件数字签名验证<a name="section51703441649"></a>**

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参考《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请不要使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[https://support.huawei.com/carrier/digitalSignatureAction](https://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

**开源组件源码<a name="section149534517468"></a>**

集群调度提供Ascend Docker Runtime、NPU Exporter、Ascend Device Plugin、Volcano、Ascend Operator、NodeD和ClusterD等开源组件。如果用户需要了解源码或定制开发组件，则可根据[表2](#table978944123012)获取相应组件源码。

**表 2**  获取组件源码

<a name="table978944123012"></a>
|组件名|源码地址|
|--|--|
|Ascend Docker Runtime|https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-docker-runtime|
|NPU Exporter|https://gitcode.com/Ascend/mind-cluster/tree/master/component/npu-exporter|
|Ascend Device Plugin|https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-device-plugin|
|Volcano|https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-for-volcano|
|Ascend Operator|https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-operator|
|NodeD|https://gitcode.com/Ascend/mind-cluster/tree/master/component/noded|
|ClusterD|https://gitcode.com/Ascend/mind-cluster/tree/master/component/clusterd|
|TaskD|https://gitcode.com/Ascend/mind-cluster/tree/master/component/taskd|
|Container Manager|https://gitcode.com/Ascend/mind-cluster/tree/master/component/container-manager|

### 安装前准备<a name="ZH-CN_TOPIC_0000002479386432"></a>






#### 准备镜像<a name="ZH-CN_TOPIC_0000002479226488"></a>

用户可通过以下两种方式准备镜像，获取镜像后依次为安装的相应组件创建节点标签、创建用户、创建日志目录和创建命名空间。

-   （推荐）[制作镜像](#section106851195114)。本章节以Ascend Operator为例，介绍了制作集群调度组件容器部署时所需镜像的操作步骤。软件包中的Dockerfile仅作为参考，用户可基于本示例制作定制化镜像。

-   [从昇腾镜像仓库拉取镜像](#section133861705416)。用户可以从镜像仓库获取制作好的集群调度各组件的镜像。

>[!NOTE] 说明 
>-   拉取或者制作镜像完成后，请及时进行安全加固，如修复基础镜像的漏洞、安装第三方依赖导致的漏洞等。
>-   在K8s所使用的容器运行时中导入镜像。如K8s  1.24以上版本默认使用Containerd作为容器运行时，拉取或者制作完镜像后需要将镜像导入到Containerd中。
>-   NPU Exporter和Ascend Device Plugin的运行用户为root，在对应的Dockerfile中配置了LD\_LIBRARY\_PATH环境变量，其中的值包含了驱动库的相关路径。组件运行时会使用到其中的文件，建议驱动安装时指定的运行用户为root，避免用户不一致带来的提权风险。

**制作镜像<a name="section106851195114"></a>**

1.  在[获取软件包](#获取软件包)章节，获取需要安装的集群调度组件软件包。
2.  将软件包解压后，上传到制作镜像服务器的任意目录。以Ascend Operator为例，放到“/home/ascend-operator”目录，目录结构如下。

    ```
    root@node:/home/ascend-operator# ll
    total 41388
    drwxr-xr-x 2 root root     4096 Aug 26 20:20 ./
    drwxr-xr-x 6 root root     4096 Aug 26 20:20 ../
    -r-x------ 1 root root 41992192 Aug 26 02:02 ascend-operator*
    -r-------- 1 root root   372291 Aug 26 02:02 ascend-operator-v{version}.yaml
    -r-------- 1 root root      482 Aug 26 02:02 Dockerfile
    ```

    >[!NOTE] 说明 
    >NPU Exporter和Ascend Device Plugin若以容器化的形式部署在Atlas 200I SoC A1 核心板上，需要进行如下操作。
    >1.  在制作镜像时检查宿主机HwHiAiUser、HwDmUser、HwBaseUser用户的UID和GID，并记录该GID和UID的取值。
    >2.  查看在Dockerfile-310P-1usoc中创建HwHiAiUser、HwDmUser、HwBaseUser用户时指定的GID和UID是否与宿主机的一致。如果一致则不做修改；如果不一致，请手动修改Dockerfile-310P-1usoc文件使其保持一致，同时需要保证每台宿主机上HwHiAiUser、HwDmUser、HwBaseUser用户的GID和UID的取值一致。

3.  检查制作集群调度组件镜像的节点是否存在如下基础镜像。

    -   执行**docker images | grep ubuntu**命令检查Ubuntu镜像，ARM架构和x86\_64架构镜像大小有差异。

        ```
        ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
        ```

    -   如果需要安装Volcano，则需要检查alpine镜像是否存在。执行**docker images | grep alpine**命令检查，回显示例如下，ARM架构和x86\_64架构镜像大小有差异。

        ```
        alpine            latest              a24bb4013296        2 years ago         5.57MB
        ```

    若上述基础镜像不存在，使用[表1](#table17241135718196)中相关命令拉取基础镜像（拉取镜像需要服务器能访问互联网）。

    **表 1**  获取基础镜像命令

    <a name="table17241135718196"></a>
    |基础镜像|拉取镜像命令|说明|
    |--|--|--|
    |ubuntu:22.04|<pre class="screen">docker pull ubuntu:22.04</pre>|拉取时自动识别系统架构。|
    |alpine:latest|<ul><li><span>x86_64</span>架构<pre class="screen">docker pull alpine:latest</pre></li><li><span>ARM</span>架构<pre class="screen"><p>docker pull arm64v8/alpine:latest</p><p>docker tag arm64v8/alpine:latest alpine:latest</p></pre></li></ul>|-|

4.  进入组件解压目录，执行**docker build**命令制作镜像，命令参考如下[表2](#table998719467243)。

    **表 2**  各组件镜像制作命令

    <a name="table998719467243"></a>
    <table><thead align="left"><tr id="row4988174618246"><th class="cellrowborder" valign="top" width="12.941294129412938%" id="mcps1.2.5.1.1"><p id="p14926203952810"><a name="p14926203952810"></a><a name="p14926203952810"></a>节点产品类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="13.081308130813083%" id="mcps1.2.5.1.2"><p id="p09883468245"><a name="p09883468245"></a><a name="p09883468245"></a>组件名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.76547654765477%" id="mcps1.2.5.1.3"><p id="p998884619247"><a name="p998884619247"></a><a name="p998884619247"></a>镜像制作命令</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.21192119211921%" id="mcps1.2.5.1.4"><p id="p438416952520"><a name="p438416952520"></a><a name="p438416952520"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row2098819467246"><td class="cellrowborder" valign="top" width="12.941294129412938%" headers="mcps1.2.5.1.1 "><p id="p179024214293"><a name="p179024214293"></a><a name="p179024214293"></a>其他产品</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="13.081308130813083%" headers="mcps1.2.5.1.2 "><p id="p34169197258"><a name="p34169197258"></a><a name="p34169197258"></a><span id="ph36246385212"><a name="ph36246385212"></a><a name="ph36246385212"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="54.76547654765477%" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen3237730141519"><a name="screen3237730141519"></a><a name="screen3237730141519"></a>docker build --no-cache -t ascend-k8sdeviceplugin:<em id="i02419301157"><a name="i02419301157"></a><a name="i02419301157"></a>{</em><em id="i133991029173612"><a name="i133991029173612"></a><a name="i133991029173612"></a>tag}</em> ./</pre>
    </td>
    <td class="cellrowborder" rowspan="8" valign="top" width="19.21192119211921%" headers="mcps1.2.5.1.4 "><p id="p10280193431010"><a name="p10280193431010"></a><a name="p10280193431010"></a><em id="i472612293915"><a name="i472612293915"></a><a name="i472612293915"></a>{tag}</em>需要参考软件包上的版本。如：软件包上版本为<span id="ph18653133316811"><a name="ph18653133316811"></a><a name="ph18653133316811"></a>7.3.0</span>，则<em id="i1572610273910"><a name="i1572610273910"></a><a name="i1572610273910"></a>{tag}</em>为v<span id="ph205239348813"><a name="ph205239348813"></a><a name="ph205239348813"></a>7.3.0</span>。</p>
    <div class="note" id="note1217913258443"><a name="note1217913258443"></a><a name="note1217913258443"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p11793259444"><a name="p11793259444"></a><a name="p11793259444"></a>请确保Dockerfile-310P-1usoc中HwDmUser和HwBaseUser的<span id="ph18833164913291"><a name="ph18833164913291"></a><a name="ph18833164913291"></a>GID</span>和<span id="ph5530185193011"><a name="ph5530185193011"></a><a name="ph5530185193011"></a>UID</span>与物理机上的保持一致。</p>
    </div></div>
    <p id="p7733142881719"><a name="p7733142881719"></a><a name="p7733142881719"></a></p>
    </td>
    </tr>
    <tr id="row11961911142910"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1519601142915"><a name="p1519601142915"></a><a name="p1519601142915"></a><span id="ph138789131469"><a name="ph138789131469"></a><a name="ph138789131469"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen11251535101518"><a name="screen11251535101518"></a><a name="screen11251535101518"></a>docker build --no-cache -t<strong id="b412563510158"><a name="b412563510158"></a><a name="b412563510158"></a> </strong>ascend-k8sdeviceplugin:<em id="i14896103963618"><a name="i14896103963618"></a><a name="i14896103963618"></a>{</em><em id="i108961395368"><a name="i108961395368"></a><a name="i108961395368"></a>tag}</em> -f Dockerfile-310P-1usoc ./</pre>
    </td>
    </tr>
    <tr id="row098844612415"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p3927139182817"><a name="p3927139182817"></a><a name="p3927139182817"></a>其他产品</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.5.1.2 "><p id="p114161919102520"><a name="p114161919102520"></a><a name="p114161919102520"></a><span id="ph5113121424115"><a name="ph5113121424115"></a><a name="ph5113121424115"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen194843931520"><a name="screen194843931520"></a><a name="screen194843931520"></a>docker build --no-cache -t npu-exporter:<em id="i1233412449361"><a name="i1233412449361"></a><a name="i1233412449361"></a>{</em><em id="i16334174433615"><a name="i16334174433615"></a><a name="i16334174433615"></a>tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row435991410290"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6359161411292"><a name="p6359161411292"></a><a name="p6359161411292"></a><span id="ph1257419163460"><a name="ph1257419163460"></a><a name="ph1257419163460"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen18159134401518"><a name="screen18159134401518"></a><a name="screen18159134401518"></a>docker build --no-cache -t<strong id="b416024416154"><a name="b416024416154"></a><a name="b416024416154"></a> </strong>npu-exporter:<em id="i1316184923612"><a name="i1316184923612"></a><a name="i1316184923612"></a>{</em><em id="i21616493369"><a name="i21616493369"></a><a name="i21616493369"></a>tag}</em> -f Dockerfile-310P-1usoc ./</pre>
    </td>
    </tr>
    <tr id="row16602529173910"><td class="cellrowborder" rowspan="5" valign="top" headers="mcps1.2.5.1.1 "><p id="p119247391094"><a name="p119247391094"></a><a name="p119247391094"></a>其他产品</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p4603162993920"><a name="p4603162993920"></a><a name="p4603162993920"></a><span id="ph2247144612408"><a name="ph2247144612408"></a><a name="ph2247144612408"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen118201953161519"><a name="screen118201953161519"></a><a name="screen118201953161519"></a>docker build --no-cache -t ascend-operator:<em id="i1582195311159"><a name="i1582195311159"></a><a name="i1582195311159"></a>{tag} </em>./</pre>
    </td>
    </tr>
    <tr id="row17988246152414"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1741731972511"><a name="p1741731972511"></a><a name="p1741731972511"></a><span id="ph16157133165316"><a name="ph16157133165316"></a><a name="ph16157133165316"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen2020115813153"><a name="screen2020115813153"></a><a name="screen2020115813153"></a>docker build --no-cache -t resilience-controller:<em id="i1078611616374"><a name="i1078611616374"></a><a name="i1078611616374"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row139888467245"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15417131916251"><a name="p15417131916251"></a><a name="p15417131916251"></a><span id="ph78731053479"><a name="ph78731053479"></a><a name="ph78731053479"></a>NodeD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen27324211618"><a name="screen27324211618"></a><a name="screen27324211618"></a>docker build --no-cache -t noded:<em id="i693671211372"><a name="i693671211372"></a><a name="i693671211372"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row273319281179"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1973362871712"><a name="p1973362871712"></a><a name="p1973362871712"></a><span id="ph143563971716"><a name="ph143563971716"></a><a name="ph143563971716"></a>ClusterD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen134421047161717"><a name="screen134421047161717"></a><a name="screen134421047161717"></a>docker build --no-cache -t clusterd:<em id="i1344219474175"><a name="i1344219474175"></a><a name="i1344219474175"></a>{tag}</em> ./</pre>
    </td>
    </tr>
    <tr id="row1498819461243"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p7417181910258"><a name="p7417181910258"></a><a name="p7417181910258"></a><span id="ph1841103815159"><a name="ph1841103815159"></a><a name="ph1841103815159"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7611881466"><a name="p7611881466"></a><a name="p7611881466"></a>进入<span id="ph11611128154615"><a name="ph11611128154615"></a><a name="ph11611128154615"></a>Volcano</span>组件解压目录，选择以下版本路径并进入。</p>
    <a name="ul1193395714453"></a><a name="ul1193395714453"></a><ul id="ul1193395714453"><li>v1.7.0版本执行以下命令。<pre class="screen" id="screen73221362140"><a name="screen73221362140"></a><a name="screen73221362140"></a>docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f ./Dockerfile-scheduler
    docker build --no-cache -t volcanosh/vc-controller-manager:v1.7.0 ./ -f ./Dockerfile-controller</pre>
    </li><li>v1.9.0版本执行以下命令。<pre class="screen" id="screen20630163032915"><a name="screen20630163032915"></a><a name="screen20630163032915"></a>docker build --no-cache -t volcanosh/vc-scheduler:v1.9.0 ./ -f ./Dockerfile-scheduler
    docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./ -f ./Dockerfile-controller</pre>
    </li></ul>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p966311264620"><a name="p966311264620"></a><a name="p966311264620"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

    以Ascend Operator组件的镜像制作为例，执行<b>docker build --no-cache -t ascend-operator:v\{version\} .</b>命令进行制作，回显示例如下。

    ```
    DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
                Install the buildx component to build images with BuildKit:
                https://docs.docker.com/go/buildx/
    Sending build context to Docker daemon  42.37MB
    Step 1/5 : FROM ubuntu:22.04 as build
     ---> 1f37bb13f08a
    Step 2/5 : RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&     usermod root -s /usr/sbin/nologin
     ---> Running in d43f1927b1fd
    Removing intermediate container d43f1927b1fd
     ---> 9f1d64e06ee6
    Step 3/5 : COPY ./ascend-operator  /usr/local/bin/
     ---> 5022b58c516e
    Step 4/5 : RUN chown -R hwMindX:hwMindX /usr/local/bin/ascend-operator  &&    chmod 500 /usr/local/bin/ascend-operator &&    chmod 750 /home/hwMindX &&    echo 'umask 027' >> /etc/profile &&     echo 'source /etc/profile' >> /home/hwMindX/.bashrc
     ---> Running in a781bde3dc56
    Removing intermediate container a781bde3dc56
     ---> 3d7e2ee7a3bd
    Step 5/5 : USER hwMindX
     ---> Running in 338954be8d99
    Removing intermediate container 338954be8d99
     ---> 103f6a2b43a5
    Successfully built 103f6a2b43a5
    Successfully tagged ascend-operator:v{version}
    ```

5.  满足以下场景可以跳过本步骤。

    -   已将制作好的集群调度组件镜像上传到私有镜像仓库，各节点可以通过私有镜像仓库拉取集群调度组件的镜像。
    -   已在安装集群调度组件各节点制作好了组件对应的镜像。

    如不满足上述场景，则需要手动分发各组件镜像到各个节点。以NodeD组件为例，使用离线镜像包的方式，分发镜像到其他节点。

    1.  将制作完成的镜像保存成离线镜像。

        ```
        docker save noded:v{version} > noded-v{version}-linux-aarch64.tar
        ```

    2.  将镜像拷贝到其他节点。

        ```
        scp noded-v{version}-linux-aarch64.tar root@{目标节点IP地址}:保存路径
        ```

    3.  以root用户登录各个节点载入离线镜像。

        ```
        docker load < noded-v{version}-linux-aarch64.tar
        ```

6.  （可选）导入离线镜像到Containerd中。本步骤适用于容器运行时为Containerd场景，其他场景下可跳过。

    以NodeD组件为例，使用离线镜像包的方式，执行以下命令。

    ```
    ctr -n k8s.io images import noded-v{version}-linux-aarch64.tar
    ```

**从昇腾镜像仓库拉取镜像<a name="section133861705416"></a>**

1.  确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2.  <a name="li1381232414410"></a>在左侧导航栏选择“MindCluster”，然后根据下表选择组件对应的镜像。拉取的镜像需要重命名后才能使用组件启动YAML进行部署，可参考[步骤3](#li14816124549)。

    **表 3**  镜像列表

    <a name="table981217243412"></a>
    <table><thead align="left"><tr id="row1781262416419"><th class="cellrowborder" valign="top" width="28.689999999999998%" id="mcps1.2.5.1.1"><p id="p168129241348"><a name="p168129241348"></a><a name="p168129241348"></a>组件</p>
    </th>
    <th class="cellrowborder" valign="top" width="34.43%" id="mcps1.2.5.1.2"><p id="p581214248413"><a name="p581214248413"></a><a name="p581214248413"></a>镜像名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="17.21%" id="mcps1.2.5.1.3"><p id="p12812122410414"><a name="p12812122410414"></a><a name="p12812122410414"></a>镜像tag</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.67%" id="mcps1.2.5.1.4"><p id="p28136241144"><a name="p28136241144"></a><a name="p28136241144"></a>拉取镜像的节点</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1581362414416"><td class="cellrowborder" valign="top" width="28.689999999999998%" headers="mcps1.2.5.1.1 "><p id="p1881318241247"><a name="p1881318241247"></a><a name="p1881318241247"></a><span id="ph11813182415418"><a name="ph11813182415418"></a><a name="ph11813182415418"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="34.43%" headers="mcps1.2.5.1.2 "><p id="p2081312411418"><a name="p2081312411418"></a><a name="p2081312411418"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/a04c486c9d7c41f1a9b9d21d929d8903" target="_blank" rel="noopener noreferrer">resilience-controller</a></p>
    </td>
    <td class="cellrowborder" valign="top" width="17.21%" headers="mcps1.2.5.1.3 "><p id="p198132242418"><a name="p198132242418"></a><a name="p198132242418"></a>v<span id="ph54141371785"><a name="ph54141371785"></a><a name="ph54141371785"></a>7.3.0</span></p>
    </td>
    <td class="cellrowborder" rowspan="4" valign="top" width="19.67%" headers="mcps1.2.5.1.4 "><p id="p18131924748"><a name="p18131924748"></a><a name="p18131924748"></a>管理节点</p>
    <p id="p1081314241741"><a name="p1081314241741"></a><a name="p1081314241741"></a></p>
    </td>
    </tr>
    <tr id="row38132241945"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p138133241142"><a name="p138133241142"></a><a name="p138133241142"></a><span id="ph88139247418"><a name="ph88139247418"></a><a name="ph88139247418"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><a name="ul158133245418"></a><a name="ul158133245418"></a><ul id="ul158133245418"><li><a href="https://www.hiascend.com/developer/ascendhub/detail/54545fa4ff9f446e914bf44b85efdb61" target="_blank" rel="noopener noreferrer">volcanosh/vc-scheduler</a></li><li><a href="https://www.hiascend.com/developer/ascendhub/detail/16f17a3c95d54f9da710a9c51bfceaa3" target="_blank" rel="noopener noreferrer">volcanosh/vc-controller-manager</a></li></ul>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p38142241846"><a name="p38142241846"></a><a name="p38142241846"></a>根据需要选择镜像：</p>
    <p id="p1814102416419"><a name="p1814102416419"></a><a name="p1814102416419"></a>v1.7.0-v<span id="ph616117387810"><a name="ph616117387810"></a><a name="ph616117387810"></a>7.3.0</span></p>
    <p id="p9814824342"><a name="p9814824342"></a><a name="p9814824342"></a>v1.9.0-v<span id="ph57147381283"><a name="ph57147381283"></a><a name="ph57147381283"></a>7.3.0</span></p>
    </td>
    </tr>
    <tr id="row38143241742"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p128147241147"><a name="p128147241147"></a><a name="p128147241147"></a><span id="ph168144244410"><a name="ph168144244410"></a><a name="ph168144244410"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1381415241342"><a name="p1381415241342"></a><a name="p1381415241342"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/a066319600634cf6a1e522856a63a1c5" target="_blank" rel="noopener noreferrer">ascend-operator</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1881412416419"><a name="p1881412416419"></a><a name="p1881412416419"></a>v<span id="ph19259839285"><a name="ph19259839285"></a><a name="ph19259839285"></a>7.3.0</span></p>
    </td>
    </tr>
    <tr id="row1381419241342"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1814324740"><a name="p1814324740"></a><a name="p1814324740"></a><span id="ph88147247419"><a name="ph88147247419"></a><a name="ph88147247419"></a>ClusterD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p98151024149"><a name="p98151024149"></a><a name="p98151024149"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/b554929b470747448924bc786b5ab95d" target="_blank" rel="noopener noreferrer">clusterd</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1481592418419"><a name="p1481592418419"></a><a name="p1481592418419"></a>v<span id="ph9804039087"><a name="ph9804039087"></a><a name="ph9804039087"></a>7.3.0</span></p>
    </td>
    </tr>
    <tr id="row138151249410"><td class="cellrowborder" valign="top" width="28.689999999999998%" headers="mcps1.2.5.1.1 "><p id="p1881520248414"><a name="p1881520248414"></a><a name="p1881520248414"></a><span id="ph081511241449"><a name="ph081511241449"></a><a name="ph081511241449"></a>NodeD</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="34.43%" headers="mcps1.2.5.1.2 "><p id="p1681572413418"><a name="p1681572413418"></a><a name="p1681572413418"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/cc7e6c0a10834f1888d790174fba4bc5" target="_blank" rel="noopener noreferrer">noded</a></p>
    </td>
    <td class="cellrowborder" valign="top" width="17.21%" headers="mcps1.2.5.1.3 "><p id="p108159249411"><a name="p108159249411"></a><a name="p108159249411"></a>v<span id="ph19289104014814"><a name="ph19289104014814"></a><a name="ph19289104014814"></a>7.3.0</span></p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="19.67%" headers="mcps1.2.5.1.4 "><p id="p128156248413"><a name="p128156248413"></a><a name="p128156248413"></a>计算节点</p>
    </td>
    </tr>
    <tr id="row08151024548"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p281518242412"><a name="p281518242412"></a><a name="p281518242412"></a><span id="ph481514241548"><a name="ph481514241548"></a><a name="ph481514241548"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p481512243413"><a name="p481512243413"></a><a name="p481512243413"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/1b1a8c3cc1ff4710bdb0222514a8a7a3" target="_blank" rel="noopener noreferrer">npu-exporter</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p081515241546"><a name="p081515241546"></a><a name="p081515241546"></a>v<span id="ph1878517407813"><a name="ph1878517407813"></a><a name="ph1878517407813"></a>7.3.0</span></p>
    </td>
    </tr>
    <tr id="row1781532410415"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p78163241644"><a name="p78163241644"></a><a name="p78163241644"></a><span id="ph148168241849"><a name="ph148168241849"></a><a name="ph148168241849"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1081612418413"><a name="p1081612418413"></a><a name="p1081612418413"></a><a href="https://www.hiascend.com/developer/ascendhub/detail/a592da7bd2ab4dffa8864abd4eac5068" target="_blank" rel="noopener noreferrer">ascend-k8sdeviceplugin</a></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p19816132417413"><a name="p19816132417413"></a><a name="p19816132417413"></a>v<span id="ph210911425819"><a name="ph210911425819"></a><a name="ph210911425819"></a>7.3.0</span></p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 说明 
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。

3.  <a name="li14816124549"></a>昇腾镜像仓库中拉取的集群调度镜像与组件启动YAML中的名字不一致，需要重命名拉取的镜像后才能启动。根据以下步骤将[2](#li1381232414410)中获取的镜像重新命名，同时建议删除原始名字的镜像。具体操作如下。
    1.  执行以下命令，重命名镜像（用户需根据所使用的组件，选取对应命令执行）。

        ```
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/resilience-controller:v7.3.0 resilience-controller:v7.3.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v7.3.0 ascend-operator:v7.3.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:v7.3.0 npu-exporter:v7.3.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v7.3.0 ascend-k8sdeviceplugin:v7.3.0
        
        # 使用1.9.0版本的Volcano，需要将镜像tag修改为v1.9.0-v7.3.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v7.3.0 volcanosh/vc-controller-manager:v1.7.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v7.3.0 volcanosh/vc-scheduler:v1.7.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v7.3.0 noded:v7.3.0
        
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v7.3.0 clusterd:v7.3.0
        ```

    2.  （可选）执行以下命令，删除原始名字镜像（用户需根据所使用的组件，选取对应命令执行）。

        ```
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/resilience-controller:v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v7.3.0
        # 使用1.9.0版本的Volcano，需要将镜像tag修改为v1.9.0-v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v7.3.0
        docker rmi swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v7.3.0
        ```

4.  （可选）导入离线镜像到Containerd中。本步骤适用于容器运行时为Containerd场景，其他场景下可跳过。

    以NodeD组件为例，使用离线镜像包的方式，执行以下步骤。

    1.  将制作完成的镜像保存成离线镜像。

        ```
        docker save noded:v{version} > noded-v{version}-linux-aarch64.tar
        ```

    2.  将离线镜像导入Containerd中。

        ```
        ctr -n k8s.io images import noded-v{version}-linux-aarch64.tar
        ```

#### 创建节点标签<a name="ZH-CN_TOPIC_0000002511426279"></a>

K8s集群中，如果将包含昇腾AI处理器的节点作为K8s的管理节点，此时该节点既是管理节点又是计算节点，除了需要管理节点对应的标签外，还需要根据节点的昇腾AI处理器类型，打上计算节点的相关标签。生产环境中，管理节点一般为通用服务器，不包含昇腾AI处理器。

**操作步骤<a name="section847765415564"></a>**

1.  在任意节点执行以下命令，查询节点名称。

    ```
    kubectl get node
    ```

    回显示例如下：

    ```
    NAME       STATUS   ROLES           AGE   VERSION
    ubuntu     Ready    worker          23h   v1.17.3
    ```

2.  按照[表1](#table202738181704)的标签信息，为对应节点打标签，方便集群调度组件在各种不同形态的工作节点之间进行调度。为节点打标签的命令参考如下。

    ```
    kubectl label nodes 主机名称 标签
    ```

    以主机名称“ubuntu”，标签“masterselector=dls-master-node”为例，命令参考如下。

    ```
    kubectl label nodes ubuntu masterselector=dls-master-node
    ```

    回显示例如下，表示操作成功。

    ```
    node/ubuntu labeled
    ```

    >[!NOTE] 说明 
    >-   [表1](#table202738181704)中各节点标签的详细说明请参见[K8s原生对象说明](./api/k8s.md)章节。
    >-   请按[表1](#table202738181704)，根据节点类型和产品类型，配置所列出的所有标签。
    >-   芯片型号的数值可通过**npu-smi info**命令查询，返回的“Name”字段对应信息为芯片型号，下文的\{_xxx_\}即取“910”字符作为芯片型号数值。

    **表 1**  节点对应的标签信息

    <a name="table202738181704"></a>
    <table><thead align="left"><tr id="row627331819017"><th class="cellrowborder" valign="top" width="31.840000000000003%" id="mcps1.2.4.1.1"><p id="p19273918201"><a name="p19273918201"></a><a name="p19273918201"></a>节点类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="25.96%" id="mcps1.2.4.1.2"><p id="p3273218803"><a name="p3273218803"></a><a name="p3273218803"></a>产品类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="42.199999999999996%" id="mcps1.2.4.1.3"><p id="p19273118301"><a name="p19273118301"></a><a name="p19273118301"></a>标签</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row227451815011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p142747189017"><a name="p142747189017"></a><a name="p142747189017"></a>管理节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p102741181908"><a name="p102741181908"></a><a name="p102741181908"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><p id="p1227417181004"><a name="p1227417181004"></a><a name="p1227417181004"></a>masterselector=dls-master-node</p>
    </td>
    </tr>
    <tr id="row127412189015"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p14274118905"><a name="p14274118905"></a><a name="p14274118905"></a>计算节点</p>
    <p id="p203704324914"><a name="p203704324914"></a><a name="p203704324914"></a></p>
    <p id="p4371534493"><a name="p4371534493"></a><a name="p4371534493"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p627418181808"><a name="p627418181808"></a><a name="p627418181808"></a><span id="ph42747181102"><a name="ph42747181102"></a><a name="ph42747181102"></a>Atlas 800 训练服务器（NPU满配）</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul727421813014"></a><a name="ul727421813014"></a><ul id="ul727421813014"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row19274318806"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p742615141511"><a name="p742615141511"></a><a name="p742615141511"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p027411181309"><a name="p027411181309"></a><a name="p027411181309"></a><span id="ph127517181101"><a name="ph127517181101"></a><a name="ph127517181101"></a>Atlas 800 训练服务器（NPU半配）</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul22751618203"></a><a name="ul22751618203"></a><ul id="ul22751618203"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=half</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row92751018202"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p554271313169"><a name="p554271313169"></a><a name="p554271313169"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p527551818016"><a name="p527551818016"></a><a name="p527551818016"></a><span id="ph1427511188015"><a name="ph1427511188015"></a><a name="ph1427511188015"></a>Atlas 800T A2 训练服务器</span>或<span id="ph102750181803"><a name="ph102750181803"></a><a name="ph102750181803"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul32752181202"></a><a name="ul32752181202"></a><ul id="ul32752181202"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph12761718301"><a name="ph12761718301"></a><a name="ph12761718301"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b-8</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row8394133819129"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p1237115354918"><a name="p1237115354918"></a><a name="p1237115354918"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p2039613891219"><a name="p2039613891219"></a><a name="p2039613891219"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
    <p id="p12463112181614"><a name="p12463112181614"></a><a name="p12463112181614"></a><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul3874134511121"></a><a name="ul3874134511121"></a><ul id="ul3874134511121"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-a3-16-super-pod</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row69181319336"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p738423163315"><a name="p738423163315"></a><a name="p738423163315"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p1584884715522"><a name="p1584884715522"></a><a name="p1584884715522"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul537611425289"></a><a name="ul537611425289"></a><ul id="ul537611425289"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul>
    <a name="ul13263154872811"></a><a name="ul13263154872811"></a><ul id="ul13263154872811"><li>host-arch=huawei-x86或host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul>
    <a name="ul17911532280"></a><a name="ul17911532280"></a><ul id="ul17911532280"><li>accelerator-type=module-a3-16</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row271845218270"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p188095589274"><a name="p188095589274"></a><a name="p188095589274"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p164951627162819"><a name="p164951627162819"></a><a name="p164951627162819"></a><span id="ph19495127162814"><a name="ph19495127162814"></a><a name="ph19495127162814"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul16834964293"></a><a name="ul16834964293"></a><ul id="ul16834964293"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li></ul>
    <a name="ul128341660299"></a><a name="ul128341660299"></a><ul id="ul128341660299"><li>host-arch=huawei-x86或host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li></ul>
    <a name="ul168341764299"></a><a name="ul168341764299"></a><ul id="ul168341764299"><li>accelerator-type=module-a3-16</li><li>server-usage=infer</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row42763185011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p16530201015713"><a name="p16530201015713"></a><a name="p16530201015713"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p19276111815011"><a name="p19276111815011"></a><a name="p19276111815011"></a><span id="ph152766181106"><a name="ph152766181106"></a><a name="ph152766181106"></a>Atlas 800I A2 推理服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul72766183018"></a><a name="ul72766183018"></a><ul id="ul72766183018"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph2027661812017"><a name="ph2027661812017"></a><a name="ph2027661812017"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li><li>server-usage=infer</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row1468510421395"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p868624283911"><a name="p868624283911"></a><a name="p868624283911"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p534220145119"><a name="p534220145119"></a><a name="p534220145119"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul19511133318489"></a><a name="ul19511133318489"></a><ul id="ul19511133318489"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph175351194911"><a name="ph175351194911"></a><a name="ph175351194911"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-8</li><li>server-usage=infer</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row13277101813019"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p356115645715"><a name="p356115645715"></a><a name="p356115645715"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p122778182014"><a name="p122778182014"></a><a name="p122778182014"></a><span id="ph3277518801"><a name="ph3277518801"></a><a name="ph3277518801"></a>Atlas 200T A2 Box16 异构子框</span></p>
    <p id="p1993115373112"><a name="p1993115373112"></a><a name="p1993115373112"></a><span id="ph10949202261219"><a name="ph10949202261219"></a><a name="ph10949202261219"></a>Atlas 200I A2 Box16 异构子框</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul15277318601"></a><a name="ul15277318601"></a><ul id="ul15277318601"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=module-<span id="ph52776181604"><a name="ph52776181604"></a><a name="ph52776181604"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-16</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row1627716183019"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p556216614577"><a name="p556216614577"></a><a name="p556216614577"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p32771718506"><a name="p32771718506"></a><a name="p32771718506"></a><span id="ph162771318306"><a name="ph162771318306"></a><a name="ph162771318306"></a>训练服务器（插<span id="ph4277131818016"><a name="ph4277131818016"></a><a name="ph4277131818016"></a>Atlas 300T 训练卡</span>）</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul72771181601"></a><a name="ul72771181601"></a><ul id="ul72771181601"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>accelerator-type=card</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row62791418607"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p45625617576"><a name="p45625617576"></a><a name="p45625617576"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p122793182008"><a name="p122793182008"></a><a name="p122793182008"></a>推理服务器（插<span id="ph19279181811010"><a name="ph19279181811010"></a><a name="ph19279181811010"></a>Atlas 300I 推理卡</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul127919181101"></a><a name="ul127919181101"></a><ul id="ul127919181101"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row72822181005"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p165621264571"><a name="p165621264571"></a><a name="p165621264571"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p16282118603"><a name="p16282118603"></a><a name="p16282118603"></a><span id="ph182828181802"><a name="ph182828181802"></a><a name="ph182828181802"></a>Atlas 推理系列产品</span>（除<span id="ph828261816012"><a name="ph828261816012"></a><a name="ph828261816012"></a>Atlas 200I SoC A1 核心板</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul162825182010"></a><a name="ul162825182010"></a><ul id="ul162825182010"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310P</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row328212184011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p20562266579"><a name="p20562266579"></a><a name="p20562266579"></a>计算节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p228281818011"><a name="p228281818011"></a><a name="p228281818011"></a><span id="ph928241810010"><a name="ph928241810010"></a><a name="ph928241810010"></a><span id="ph122828181609"><a name="ph122828181609"></a><a name="ph122828181609"></a>Atlas 200I SoC A1 核心板</span></span></p>
    </td>
    <td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul202825181508"></a><a name="ul202825181508"></a><ul id="ul202825181508"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend310P</li><li>servertype=soc</li><li>（可选）nodeDEnable=on</li></ul>
    </td>
    </tr>
    <tr id="row328212184011"><td class="cellrowborder" valign="top" width="31.840000000000003%" headers="mcps1.2.4.1.1 "><p id="p20562266579"><a name="p20562266579"></a><a name="p20562266579"></a>计算节点</p>
 	</td>
 	<td class="cellrowborder" valign="top" width="25.96%" headers="mcps1.2.4.1.2 "><p id="p228281818011"><a name="p228281818011"></a><a name="p228281818011"></a><span id="ph928241810010"><a name="ph928241810010"></a><a name="ph928241810010"></a><span id="ph122828181609"><a name="ph122828181609"></a><a name="ph122828181609"></a>Atlas 350 标卡</span></span></p>
 	</td>
 	<td class="cellrowborder" valign="top" width="42.199999999999996%" headers="mcps1.2.4.1.3 "><a name="ul202825181508"></a><a name="ul202825181508"></a><ul id="ul202825181508"><li>node-role.kubernetes.io/worker=worker</li><li>workerselector=dls-worker-node</li><li>host-arch=huawei-arm或host-arch=huawei-x86</li><li>accelerator=huawei-Ascend910</li><li>servertype=soc</li><li>（可选）nodeDEnable=on</li><li>（可选）accelerator-type</li></ul>
 	</td>
 	</tr>
    </tbody>
    </table>

#### 创建用户<a name="ZH-CN_TOPIC_0000002511346353"></a>

在对应组件安装的节点上执行以下命令创建用户。

-   <a name="li1069651515405"></a>Ubuntu操作系统

    ```
    useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
    usermod -a -G HwHiAiUser hwMindX
    ```

-   <a name="li19202165424015"></a>CentOS操作系统

    ```
    useradd -d /home/hwMindX -u 9000 -m -s /sbin/nologin hwMindX
    usermod -a -G HwHiAiUser hwMindX
    ```

>[!NOTE] 说明 
>-   其余操作系统创建用户：
>     -   基于Ubuntu操作系统开发的操作系统，参考[Ubuntu操作系统](#li1069651515405)。
>     -   基于CentOS操作系统开发的操作系统，参考[CentOS操作系统](#li19202165424015)。
>-   HwHiAiUser是驱动或CANN软件包需要的软件运行用户。
>-   执行**getent passwd**命令，查看所有物理机（存储节点、管理节点、计算节点）和容器内，HwHiAiUser的UID和GID是否一致，且都为1000。如果被占用可能会导致服务不可用，可以参见[用户UID或GID被占用](./faq.md#用户uid或gid被占用)章节进行处理。

**表 1**  组件用户说明

<a name="table125971501113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0299839362_row86431704617"><th class="cellrowborder" valign="top" width="20.962096209620963%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0299839362_p464201754614"><a name="zh-cn_topic_0299839362_p464201754614"></a><a name="zh-cn_topic_0299839362_p464201754614"></a>组件</p>
</th>
<th class="cellrowborder" valign="top" width="34.13341334133413%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0299839362_p11647172468"><a name="zh-cn_topic_0299839362_p11647172468"></a><a name="zh-cn_topic_0299839362_p11647172468"></a>启动用户</p>
</th>
<th class="cellrowborder" valign="top" width="44.90449044904491%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0299839362_p56451734620"><a name="zh-cn_topic_0299839362_p56451734620"></a><a name="zh-cn_topic_0299839362_p56451734620"></a>是否使用特权容器</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0299839362_row3641172465"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p671453716107"><a name="p671453716107"></a><a name="p671453716107"></a><span id="ph14925450192719"><a name="ph14925450192719"></a><a name="ph14925450192719"></a>NPU Exporter</span></p>
</td>
<td class="cellrowborder" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><a name="ul124012695512"></a><a name="ul124012695512"></a><ul id="ul124012695512"><li>二进制运行：hwMindX</li><li>容器运行：root</li></ul>
</td>
<td class="cellrowborder" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><a name="ul8401830195518"></a><a name="ul8401830195518"></a><ul id="ul8401830195518"><li>二进制运行：不涉及。</li><li>容器运行：需要使用特权容器，建议用户使用二进制运行。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0299839362_row1064121764612"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0299839362_p16641317134612"><a name="zh-cn_topic_0299839362_p16641317134612"></a><a name="zh-cn_topic_0299839362_p16641317134612"></a><span id="ph522114212719"><a name="ph522114212719"></a><a name="ph522114212719"></a>Ascend Device Plugin</span></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p53735269103"><a name="p53735269103"></a><a name="p53735269103"></a>root</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p29286561106"><a name="p29286561106"></a><a name="p29286561106"></a>需要使用特权容器。</p>
</td>
</tr>
<tr id="row10935147171519"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1935947181513"><a name="p1935947181513"></a><a name="p1935947181513"></a><span id="ph5551115391513"><a name="ph5551115391513"></a><a name="ph5551115391513"></a>NodeD</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0299839362_row664817164615"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0299839362_p0649177466"><a name="zh-cn_topic_0299839362_p0649177466"></a><a name="zh-cn_topic_0299839362_p0649177466"></a><span id="ph175881448132716"><a name="ph175881448132716"></a><a name="ph175881448132716"></a>Volcano</span></p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p153424813128"><a name="p153424813128"></a><a name="p153424813128"></a>hwMindX</p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p17327314131212"><a name="p17327314131212"></a><a name="p17327314131212"></a>不涉及。</p>
</td>
</tr>
<tr id="row24141825191817"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1941515259187"><a name="p1941515259187"></a><a name="p1941515259187"></a><span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span></p>
</td>
</tr>
<tr id="row29051413163917"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p390551333913"><a name="p390551333913"></a><a name="p390551333913"></a><span id="ph829115811272"><a name="ph829115811272"></a><a name="ph829115811272"></a>Resilience Controller</span></p>
</td>
</tr>
<tr id="row1674814434406"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p97491434407"><a name="p97491434407"></a><a name="p97491434407"></a><span id="ph1566531814589"><a name="ph1566531814589"></a><a name="ph1566531814589"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row6784854202610"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p11621711181811"><a name="p11621711181811"></a><a name="p11621711181811"></a><span id="ph1790811553279"><a name="ph1790811553279"></a><a name="ph1790811553279"></a>Elastic Agent</span></p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p161622011121819"><a name="p161622011121819"></a><a name="p161622011121819"></a>由用户自行决定，建议使用非root用户。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p1916271131815"><a name="p1916271131815"></a><a name="p1916271131815"></a>由用户自行决定，建议不使用特权容器。</p>
</td>
</tr>
<tr id="row315419369301"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p715593611302"><a name="p715593611302"></a><a name="p715593611302"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
</td>
</tr>
<tr id="row3502131311115"><td class="cellrowborder" valign="top" width="20.962096209620963%" headers="mcps1.2.4.1.1 "><p id="p175021513201117"><a name="p175021513201117"></a><a name="p175021513201117"></a><span id="ph16988102112717"><a name="ph16988102112717"></a><a name="ph16988102112717"></a>Container Manager</span></p>
</td>
<td class="cellrowborder" valign="top" width="34.13341334133413%" headers="mcps1.2.4.1.2 "><p id="p1450212134110"><a name="p1450212134110"></a><a name="p1450212134110"></a>root</p>
</td>
<td class="cellrowborder" valign="top" width="44.90449044904491%" headers="mcps1.2.4.1.3 "><p id="p6502191318116"><a name="p6502191318116"></a><a name="p6502191318116"></a>不涉及。</p>
</td>
</tr>
</tbody>
</table>

#### 创建日志目录<a name="ZH-CN_TOPIC_0000002511346417"></a>

在对应节点创建组件日志父目录和各组件的日志目录，并设置目录对应属主和权限。

**操作步骤<a name="section124928122416"></a>**

1.  执行以下命令，按照[表1 集群调度组件日志路径列表](#table957112617314)，在各节点创建组件日志父目录。

    ```
    mkdir -m 755 /var/log/mindx-dl
    chown root:root /var/log/mindx-dl
    ```

2.  根据所使用组件的具体情况，创建相应的日志目录。

    **表 1** 集群调度组件日志路径列表

    <a name="table957112617314"></a>
    <table><thead align="left"><tr id="row2057210616310"><th class="cellrowborder" valign="top" width="21.93%" id="mcps1.2.5.1.1"><p id="p10572761231"><a name="p10572761231"></a><a name="p10572761231"></a>组件</p>
    </th>
    <th class="cellrowborder" valign="top" width="41.91%" id="mcps1.2.5.1.2"><p id="p11572156430"><a name="p11572156430"></a><a name="p11572156430"></a>创建日志目录命令</p>
    </th>
    <th class="cellrowborder" valign="top" width="17.05%" id="mcps1.2.5.1.3"><p id="p25721364319"><a name="p25721364319"></a><a name="p25721364319"></a>日志路径创建节点</p>
    </th>
    <th class="cellrowborder" valign="top" width="19.11%" id="mcps1.2.5.1.4"><p id="p16572661320"><a name="p16572661320"></a><a name="p16572661320"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row457296131"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p1572469315"><a name="p1572469315"></a><a name="p1572469315"></a><span id="ph9572196532"><a name="ph9572196532"></a><a name="ph9572196532"></a>Ascend Device Plugin</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1657216638"><a name="screen1657216638"></a><a name="screen1657216638"></a>mkdir -m 750 /var/log/mindx-dl/devicePlugin
    chown root:root /var/log/mindx-dl/devicePlugin</pre>
    </td>
    <td class="cellrowborder" rowspan="5" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p11572661536"><a name="p11572661536"></a><a name="p11572661536"></a>计算节点</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p557592110325"><a name="p557592110325"></a><a name="p557592110325"></a>-</p>
    </td>
    </tr>
    <tr id="row95721761536"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p125721269315"><a name="p125721269315"></a><a name="p125721269315"></a><span id="ph14572161034"><a name="ph14572161034"></a><a name="ph14572161034"></a>NPU Exporter</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen457213611313"><a name="screen457213611313"></a><a name="screen457213611313"></a>mkdir -m 750 /var/log/mindx-dl/npu-exporter
    chown root:root /var/log/mindx-dl/npu-exporter</pre>
    </td>
    </tr>
    <tr id="row105739620318"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p195731868318"><a name="p195731868318"></a><a name="p195731868318"></a><span id="ph11573862310"><a name="ph11573862310"></a><a name="ph11573862310"></a>NodeD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1957396735"><a name="screen1957396735"></a><a name="screen1957396735"></a>mkdir -m 750 /var/log/mindx-dl/noded
    chown root:root /var/log/mindx-dl/noded</pre>
    </td>
    </tr>
    <tr id="row55731961237"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15573961314"><a name="p15573961314"></a><a name="p15573961314"></a><span id="ph13573106431"><a name="ph13573106431"></a><a name="ph13573106431"></a>Elastic Agent</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen55735616314"><a name="screen55735616314"></a><a name="screen55735616314"></a>mkdir -m 750 /var/log/mindx-dl/elastic
    chown <em id="i15731661134"><a name="i15731661134"></a><a name="i15731661134"></a>由用户自行定义</em> /var/log/mindx-dl/elastic</pre>
    <div class="note" id="note3573061032"><a name="note3573061032"></a><a name="note3573061032"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p2057310617318"><a name="p2057310617318"></a><a name="p2057310617318"></a>将<span id="ph1472342453512"><a name="ph1472342453512"></a><a name="ph1472342453512"></a>Elastic Agent</span>日志目录挂载到容器内，详见<a href="./usage/resumable_training.md#任务yaml配置示例">任务YAML配置示例</a>章节中"修改训练脚本、代码的挂载路径"步骤。</p>
    </div></div>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><a name="ul958614153510"></a><a name="ul958614153510"></a><ul id="ul958614153510"><li>目录属主由用户自定义。注意：安装<span id="ph67093892615"><a name="ph67093892615"></a><a name="ph67093892615"></a>Elastic Agent</span>的用户属组、调用<span id="ph1642075902418"><a name="ph1642075902418"></a><a name="ph1642075902418"></a>Elastic Agent</span>的运行用户属组、挂载宿主机的目录属组请保持一致。</li><li>用户可自定义<span id="ph1790811553279"><a name="ph1790811553279"></a><a name="ph1790811553279"></a>Elastic Agent</span>的运行日志的落盘路径，在该路径下，用户可查看<span id="ph1529820279122"><a name="ph1529820279122"></a><a name="ph1529820279122"></a>Elastic Agent</span>所有节点日志，无需逐一登录每个节点查看。</li></ul>
    </td>
    </tr>
    <tr id="row189638410329"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p7963164113217"><a name="p7963164113217"></a><a name="p7963164113217"></a><span id="ph11742444163719"><a name="ph11742444163719"></a><a name="ph11742444163719"></a>TaskD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen929012103313"><a name="screen929012103313"></a><a name="screen929012103313"></a>mkdir  -m 750  <em id="i15660102313617"><a name="i15660102313617"></a><a name="i15660102313617"></a>训练脚本目录</em>/taskd_log
    chown <em id="i4956143053617"><a name="i4956143053617"></a><a name="i4956143053617"></a>由用户自行定义</em> <em id="i6187123720366"><a name="i6187123720366"></a><a name="i6187123720366"></a>训练脚本目录</em>/taskd_log </pre>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><a name="ul9461980353"></a><a name="ul9461980353"></a><ul id="ul9461980353"><li>目录属主由用户自定义。</li><li><span id="ph1524182517352"><a name="ph1524182517352"></a><a name="ph1524182517352"></a>TaskD</span>在运行过程中可以自动创建对应日志目录，日志目录前缀一般为任务YAML中执行<strong id="b5881131073711"><a name="b5881131073711"></a><a name="b5881131073711"></a>bash命令</strong>或拉起训练时所在目录。</li></ul>
    </td>
    </tr>
    <tr id="row65749616319"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p8574136838"><a name="p8574136838"></a><a name="p8574136838"></a><span id="ph13574365316"><a name="ph13574365316"></a><a name="ph13574365316"></a>Ascend Operator</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen05746613313"><a name="screen05746613313"></a><a name="screen05746613313"></a>mkdir -m 750 /var/log/mindx-dl/ascend-operator
    chown hwMindX:hwMindX /var/log/mindx-dl/ascend-operator</pre>
    </td>
    <td class="cellrowborder" rowspan="5" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p65611868135"><a name="p65611868135"></a><a name="p65611868135"></a>管理节点</p>
    </td>
    <td class="cellrowborder" rowspan="5" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p11355115061313"><a name="p11355115061313"></a><a name="p11355115061313"></a>-</p>
    </td>
    </tr>
    <tr id="row45741461130"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p18574466314"><a name="p18574466314"></a><a name="p18574466314"></a><span id="ph13574176736"><a name="ph13574176736"></a><a name="ph13574176736"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen1574064313"><a name="screen1574064313"></a><a name="screen1574064313"></a>mkdir -m 750 /var/log/mindx-dl/resilience-controller
    chown hwMindX:hwMindX /var/log/mindx-dl/resilience-controller</pre>
    </td>
    </tr>
    <tr id="row68981954111810"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p28991454191811"><a name="p28991454191811"></a><a name="p28991454191811"></a><span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen161652618196"><a name="screen161652618196"></a><a name="screen161652618196"></a>mkdir -m 750 /var/log/mindx-dl/clusterd
    chown hwMindX:hwMindX /var/log/mindx-dl/clusterd</pre>
    </td>
    </tr>
    <tr id="row957413616315"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.5.1.1 "><p id="p1657414618311"><a name="p1657414618311"></a><a name="p1657414618311"></a><span id="ph185741164311"><a name="ph185741164311"></a><a name="ph185741164311"></a>Volcano</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen145741661036"><a name="screen145741661036"></a><a name="screen145741661036"></a>mkdir -m 750 /var/log/mindx-dl/volcano-controller
    chown hwMindX:hwMindX /var/log/mindx-dl/volcano-controller</pre>
    </td>
    </tr>
    <tr id="row18574568314"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><pre class="screen" id="screen1257416635"><a name="screen1257416635"></a><a name="screen1257416635"></a>mkdir -m 750 /var/log/mindx-dl/volcano-scheduler
    chown hwMindX:hwMindX /var/log/mindx-dl/volcano-scheduler</pre>
    </td>
    </tr>
    <tr id="row14307175681213"><td class="cellrowborder" valign="top" width="21.93%" headers="mcps1.2.5.1.1 "><p id="p1030717560124"><a name="p1030717560124"></a><a name="p1030717560124"></a><span id="ph172417011305"><a name="ph172417011305"></a><a name="ph172417011305"></a>Container Manager</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="41.91%" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen44681417291"><a name="screen44681417291"></a><a name="screen44681417291"></a>mkdir -m 750 /var/log/mindx-dl/container-manager
    chown root:root /var/log/mindx-dl/container-manager</pre>
    </td>
    <td class="cellrowborder" valign="top" width="17.05%" headers="mcps1.2.5.1.3 "><p id="p53074565125"><a name="p53074565125"></a><a name="p53074565125"></a>需要使用容器恢复特性的节点</p>
    </td>
    <td class="cellrowborder" valign="top" width="19.11%" headers="mcps1.2.5.1.4 "><p id="p1518124119135"><a name="p1518124119135"></a><a name="p1518124119135"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

#### 创建命名空间<a name="ZH-CN_TOPIC_0000002479226384"></a>

-   集群调度的NodeD、Resilience Controller、ClusterD和Ascend Operator组件会运行在K8s的mindx-dl命名空间下，请在K8s的管理节点执行如下命令，创建对应的命名空间。

    ```
    kubectl create ns mindx-dl
    ```

-   MindCluster上报超节点信息、pingmesh配置信息、公共故障信息需手动创建名为cluster-system命名空间。请在K8s的管理节点执行如下命令。

    ```
    kubectl create ns cluster-system
    ```

-   NPU Exporter的命名空间为npu-exporter；Volcano的命名空间为volcano-system；Ascend Device Plugin的命名空间为kube-system，上述组件的命名空间由系统创建，用户无需再次创建。

### Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002479226434"></a>

-   使用容器化支持、整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须安装Ascend Docker Runtime。
-   仅使用资源监测的用户，可以不安装Ascend Docker Runtime，请直接跳过本章节。

**前提条件<a name="section137058405153"></a>**

安装前，请确保runc文件的用户ID为0。

**确认安装场景<a name="zh-cn_topic_0000001930317932_section1235447163310"></a>**

目前仅支持root用户安装Ascend Docker Runtime，请根据实际情况选择对应的安装方式。

1.  在K8s管理节点执行以下命令，查询节点名称。

    ```
    kubectl get node
    ```

    回显示例如下：

    ```
    NAME       STATUS   ROLES           AGE   VERSION
    ubuntu     Ready    worker          23h   v1.17.3
    ```

2.  查看当前节点的容器运行时。其中node-name为节点名称。
    -   不使用K8s场景：在任意节点执行以下命令。

        ```
        docker --version      # Docker
        containerd --version     # Containerd
        ```

        -   若回显为Docker的版本信息，表示当前是[Docker场景](#zh-cn_topic_0000001930317932_section1443063532919)。
        -   若回显为Containerd的版本信息，表示当前是[Containerd场景](#zh-cn_topic_0000001930317932_section196591123133116)。
        -   若同时有Docker和Containerd的版本信息，请用户自行确定任务所要使用的容器运行时。

    -   K8s集成容器运行时场景：在管理节点执行以下命令。

        ```
        kubectl describe node <node-name> | grep -i runtime
        ```

        -   若回显中有Docker信息，表示当前是[K8s集成Docker场景](#zh-cn_topic_0000001930317932_section1443063532919)。
        -   若回显中有Containerd信息，表示当前是[K8s集成Containerd场景](#zh-cn_topic_0000001930317932_section14600174633116)。

**Docker场景下安装Ascend Docker Runtime<a name="zh-cn_topic_0000001930317932_section1443063532919"></a>**

K8s集成Docker场景安装Ascend Docker Runtime，与Docker场景下安装Ascend Docker Runtime操作一致。

1.  安装包下载完成后，在所有计算节点，进入安装包（run包）所在路径。

    ```
    cd <path to run package>
    ```

2.  执行以下命令，为软件包添加可执行权限。

    ```
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3.  执行如下命令，校验软件包安装文件的一致性和完整性。

    ```
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    回显示例如下：

    ```
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4.  可通过以下命令安装Ascend Docker Runtime。

    -   安装到默认路径下，执行以下命令。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install
        ```

    -   安装到指定路径下，执行以下命令，“--install-path”参数为指定的安装路径。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-path=<path>
        ```

    >[!NOTE] 说明 
    >-   指定安装路径时必须使用绝对路径。
    >-   Docker配置文件路径不是默认的“/etc/docker/daemon.json”时，需要新增--config-file-path参数，用于指定该配置文件路径。

    回显示例如下，表示安装成功。

    ```
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5.  执行以下命令，使Ascend Docker Runtime生效。

    ```
    systemctl daemon-reload && systemctl restart docker
    ```

    Ascend Device Plugin在启动时会自动检测Ascend Docker Runtime是否存在，所以需要先启动Ascend Docker Runtime，再启动Ascend Device Plugin。若先启动Ascend Device Plugin，再启动Ascend Docker Runtime，需要参见[Ascend Device Plugin](#ascend-device-plugin)章节重新启动Ascend Device Plugin。

**Containerd场景下安装Ascend Docker Runtime<a name="zh-cn_topic_0000001930317932_section196591123133116"></a>**

1.  安装包下载完成后，首先进入安装包（run包）所在路径。

    ```
    cd <path to run package>
    ```

2.  执行以下命令，为软件包添加可执行权限。

    ```
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3.  执行如下命令，校验软件包安装文件的一致性和完整性。

    ```
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    回显示例如下：

    ```
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4.  可通过以下命令安装Ascend Docker Runtime。

    -   安装到默认路径下。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd
        ```

    -   安装到指定路径下，执行以下命令，“--install-path”参数为指定的安装路径。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd --install-path=<path>
        ```

        >[!NOTE] 说明 
        >-   指定安装路径时必须使用绝对路径。
        >-   Containerd的配置文件路径不是默认的“/etc/containerd/config.toml”时，需要新增--config-file-path参数，用于指定该配置文件路径。

    回显示例如下，表示安装成功。

    ```
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5.  （可选）如果安装失败，可参照以下步骤修改Containerd配置文件。
    1.  修改配置文件。
        -   **Containerd无默认配置文件场景**：依次执行以下命令，创建并修改配置文件。

            ```
            mkdir /etc/containerd
            containerd config default > /etc/containerd/config.toml
            vim /etc/containerd/config.toml
            ```

        -   **Containerd已有配置文件场景**：打开并修改配置文件。

            ```
            vim /etc/containerd/config.toml
            ```

    2.  执行以下命令查询当前cgroup的版本。

        ```
        stat -fc %T /sys/fs/cgroup/
        ```

        -   若回显为tmpfs，表示当前为cgroup v1版本。
        -   若回显为cgroup2fs，表示当前为cgroup v2版本。

    3.  根据cgroup的版本修改runtime\_type字段，并修改Ascend Docker Runtime安装路径，示例如下所示。
        -   cgroup v1

            >[!NOTE] 说明 
            >若为openEuler 24.03操作系统，需将cgroup v1版本中的runtime\_type修改为io.containerd.runc.v2。

            ```
                [plugins."io.containerd.grpc.v1.cri".containerd.runtimes] 
                   [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]  
                     runtime_type = "io.containerd.runtime.v1.linux" 
                     runtime_engine = "" 
                     runtime_root = "" 
                     privileged_without_host_devices = false 
                     base_runtime_spec = "" 
                     [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options] 
               [plugins."io.containerd.grpc.v1.cri".cni] 
                 bin_dir = "/opt/cni/bin" 
                 conf_dir = "/etc/cni/net.d" 
                 max_conf_num = 1 
                 conf_template = "" 
             [plugins."io.containerd.grpc.v1.cri".registry] 
                 [plugins."io.containerd.grpc.v1.cri".registry.mirrors] 
                   [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"] 
                     endpoint = ["https://registry-1.docker.io"] 
             [plugins."io.containerd.grpc.v1.cri".image_decryption] 
                 key_model = "" 
             
            ...
             [plugins."io.containerd.monitor.v1.cgroups"] 
               no_prometheus = false 
             [plugins."io.containerd.runtime.v1.linux"] 
               shim = "containerd-shim" 
               runtime = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"   
               runtime_root = "" 
               no_shim = false 
               shim_debug = false 
             [plugins."io.containerd.runtime.v2.task"] 
               platforms = ["linux/amd64"] 
            ...
            ```

        -   cgroup v2

            ```
                    [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]
                  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
                    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
                      base_runtime_spec = ""
                      cni_conf_dir = ""
                      cni_max_conf_num = 0
                      container_annotations = []
                      pod_annotations = []
                      privileged_without_host_devices = false
                      runtime_engine = ""
                      runtime_path = ""
                      runtime_root = ""
                      runtime_type = "io.containerd.runc.v2"
                      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
                        BinaryName = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"
                        CriuImagePath = ""
                        CriuPath = ""
                        CriuWorkPath = ""
                        IoGid = 0
                        IoUid = 0
                        NoNewKeyring = false
                        NoPivotRoot = false
                        Root = ""
                        ShimCgroup = ""
                        SystemdCgroup = true
            ...
            ```

6.  执行以下命令，重启Containerd。

    ```
    systemctl daemon-reload && systemctl restart containerd
    ```

**K8s集成Containerd场景下安装Ascend Docker Runtime<a name="zh-cn_topic_0000001930317932_section14600174633116"></a>**

1.  安装包下载完成后，首先进入安装包（run包）所在路径。

    ```
    cd <path to run package>
    ```

2.  执行以下命令，为软件包添加可执行权限。

    ```
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3.  执行如下命令，校验软件包安装文件的一致性和完整性。

    ```
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    回显示例如下：

    ```
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4.  可通过以下命令安装Ascend Docker Runtime。

    -   安装到默认路径下。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd
        ```

    -   安装到指定路径下，执行以下命令，“--install-path”参数为指定的安装路径。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd --install-path=<path>
        ```

        >[!NOTE] 说明 
        >指定安装路径时必须使用绝对路径。

    回显示例如下，表示安装成功。

    ```
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5.  （可选）如果安装失败，可参照以下步骤修改Containerd配置文件。
    1.  修改配置文件。
        -   **Containerd无默认配置文件场景**：依次执行以下命令，创建并修改配置文件。

            ```
            mkdir /etc/containerd
            containerd config default > /etc/containerd/config.toml
            vim /etc/containerd/config.toml
            ```

        -   **Containerd已有配置文件场景**：打开并修改配置文件。

            ```
            vim /etc/containerd/config.toml
            ```

    2.  执行以下命令查询当前cgroup的版本。

        ```
        stat -fc %T /sys/fs/cgroup/
        ```

        -   若回显为tmpfs，表示当前为cgroup v1版本。
        -   若回显为cgroup2fs，表示当前为cgroup v2版本。

    3.  根据cgroup的版本修改runtime\_type字段，并修改Ascend Docker Runtime安装路径，示例如下所示。
        -   cgroup v1

            >[!NOTE] 说明 
            >若为openEuler 24.03操作系统，需将cgroup v1版本中的runtime\_type修改为io.containerd.runc.v2。

            ```
                [plugins."io.containerd.grpc.v1.cri".containerd.runtimes] 
                   [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]  
                     runtime_type = "io.containerd.runtime.v1.linux" 
                     runtime_engine = "" 
                     runtime_root = "" 
                     privileged_without_host_devices = false 
                     base_runtime_spec = "" 
                     [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options] 
               [plugins."io.containerd.grpc.v1.cri".cni] 
                 bin_dir = "/opt/cni/bin" 
                 conf_dir = "/etc/cni/net.d" 
                 max_conf_num = 1 
                 conf_template = "" 
             [plugins."io.containerd.grpc.v1.cri".registry] 
                 [plugins."io.containerd.grpc.v1.cri".registry.mirrors] 
                   [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"] 
                     endpoint = ["https://registry-1.docker.io"] 
             [plugins."io.containerd.grpc.v1.cri".image_decryption] 
                 key_model = "" 
             
            ...
             [plugins."io.containerd.monitor.v1.cgroups"] 
               no_prometheus = false 
             [plugins."io.containerd.runtime.v1.linux"] 
               shim = "containerd-shim" 
               runtime = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"   
               runtime_root = "" 
               no_shim = false 
               shim_debug = false 
             [plugins."io.containerd.runtime.v2.task"] 
               platforms = ["linux/amd64"] 
            ...
            ```

        -   cgroup v2

            ```
                    [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]
                  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
                    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
                      base_runtime_spec = ""
                      cni_conf_dir = ""
                      cni_max_conf_num = 0
                      container_annotations = []
                      pod_annotations = []
                      privileged_without_host_devices = false
                      runtime_engine = ""
                      runtime_path = ""
                      runtime_root = ""
                      runtime_type = "io.containerd.runc.v2"
                      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
                        BinaryName = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"
                        CriuImagePath = ""
                        CriuPath = ""
                        CriuWorkPath = ""
                        IoGid = 0
                        IoUid = 0
                        NoNewKeyring = false
                        NoPivotRoot = false
                        Root = ""
                        ShimCgroup = ""
                        SystemdCgroup = true
            ...
            ```

6.  如需将节点上的容器运行时从Docker更改为Containerd，需要修改节点上kubelet的配置文件kubeadm-flags.env。详情请参见[K8s官方文档](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/change-runtime-containerd/)。
7.  如果存在Docker服务，请执行以下命令停止对应服务。

    ```
    systemctl stop docker
    ```

8.  执行命令，重启Containerd和kubelet，示例如下。

    ```
    systemctl daemon-reload && systemctl restart containerd kubelet
    ```

**Ascend Docker Runtime安装包命令行参数说明<a name="zh-cn_topic_0000001930317932_section425619177219"></a>**

参数说明如[表1](#zh-cn_topic_0000001930317932_table35676204212)所示。

**表 1**  安装包支持的参数说明

<a name="zh-cn_topic_0000001930317932_table35676204212"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001930317932_row1856732017219"><th class="cellrowborder" valign="top" width="32.43%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000001930317932_p155677203214"><a name="zh-cn_topic_0000001930317932_p155677203214"></a><a name="zh-cn_topic_0000001930317932_p155677203214"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="67.57%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000001930317932_p1456712016216"><a name="zh-cn_topic_0000001930317932_p1456712016216"></a><a name="zh-cn_topic_0000001930317932_p1456712016216"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001930317932_row2568112072119"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p05681620192117"><a name="zh-cn_topic_0000001930317932_p05681620192117"></a><a name="zh-cn_topic_0000001930317932_p05681620192117"></a>--help | -h</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p1356892011218"><a name="zh-cn_topic_0000001930317932_p1356892011218"></a><a name="zh-cn_topic_0000001930317932_p1356892011218"></a>查询帮助信息。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row165681520112117"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p3568122042118"><a name="zh-cn_topic_0000001930317932_p3568122042118"></a><a name="zh-cn_topic_0000001930317932_p3568122042118"></a>--info</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p15568720122112"><a name="zh-cn_topic_0000001930317932_p15568720122112"></a><a name="zh-cn_topic_0000001930317932_p15568720122112"></a>查询软件包构建信息。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row756832062117"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p6568142052120"><a name="zh-cn_topic_0000001930317932_p6568142052120"></a><a name="zh-cn_topic_0000001930317932_p6568142052120"></a>--list</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p4568182018212"><a name="zh-cn_topic_0000001930317932_p4568182018212"></a><a name="zh-cn_topic_0000001930317932_p4568182018212"></a>查询软件包文件列表。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row2568520172112"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p856882092110"><a name="zh-cn_topic_0000001930317932_p856882092110"></a><a name="zh-cn_topic_0000001930317932_p856882092110"></a>--check</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p185681720182113"><a name="zh-cn_topic_0000001930317932_p185681720182113"></a><a name="zh-cn_topic_0000001930317932_p185681720182113"></a>检查软件包的一致性和完整性。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row165681920202119"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p15568122012120"><a name="zh-cn_topic_0000001930317932_p15568122012120"></a><a name="zh-cn_topic_0000001930317932_p15568122012120"></a>--quiet</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p256818204217"><a name="zh-cn_topic_0000001930317932_p256818204217"></a><a name="zh-cn_topic_0000001930317932_p256818204217"></a>静默安装，跳过交互式信息，需要配合install、uninstall或者upgrade使用。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row19568182011213"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p55691220202114"><a name="zh-cn_topic_0000001930317932_p55691220202114"></a><a name="zh-cn_topic_0000001930317932_p55691220202114"></a>--tar arg1 [arg2 ...]</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p25691320162114"><a name="zh-cn_topic_0000001930317932_p25691320162114"></a><a name="zh-cn_topic_0000001930317932_p25691320162114"></a>对软件包执行tar命令，使用tar后面的参数作为命令的参数。例如执行<strong id="zh-cn_topic_0000001930317932_b656982016214"><a name="zh-cn_topic_0000001930317932_b656982016214"></a><a name="zh-cn_topic_0000001930317932_b656982016214"></a>--tar xvf</strong>命令，解压run安装包的内容到当前目录。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row156942092116"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p75697203214"><a name="zh-cn_topic_0000001930317932_p75697203214"></a><a name="zh-cn_topic_0000001930317932_p75697203214"></a>--install</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p1357015208213"><a name="zh-cn_topic_0000001930317932_p1357015208213"></a><a name="zh-cn_topic_0000001930317932_p1357015208213"></a>安装软件包。可以指定安装路径--install-path=&lt;path&gt;，也可以不指定安装路径，直接安装到默认路径下。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row19570122010217"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p15570172014213"><a name="zh-cn_topic_0000001930317932_p15570172014213"></a><a name="zh-cn_topic_0000001930317932_p15570172014213"></a>--install-path=&lt;path&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="p369633161410"><a name="p369633161410"></a><a name="p369633161410"></a>指定安装路径。</p>
<a name="zh-cn_topic_0000001930317932_ul29611936455"></a><a name="zh-cn_topic_0000001930317932_ul29611936455"></a><ul id="zh-cn_topic_0000001930317932_ul29611936455"><li>必须使用绝对路径作为安装路径。</li><li>当环境上存在全局配置文件“ascend_docker_runtime_install.info”时，指定的安装路径必须与全局配置文件中保存的安装路径保持一致。</li><li>如用户想更换安装路径，需先卸载原路径下的<span id="zh-cn_topic_0000001930317932_ph1528115352583"><a name="zh-cn_topic_0000001930317932_ph1528115352583"></a><a name="zh-cn_topic_0000001930317932_ph1528115352583"></a>Ascend Docker Runtime</span>软件包并确保全局配置文件“ascend_docker_runtime_install.info”已被删除。</li><li>若5.0.RC1版本之前的<span id="zh-cn_topic_0000001930317932_ph93781522588"><a name="zh-cn_topic_0000001930317932_ph93781522588"></a><a name="zh-cn_topic_0000001930317932_ph93781522588"></a>Ascend Docker Runtime</span>是通过ToolBox安装包安装的，则该文件不存在，不需要删除。</li><li>若不指定安装路径，将安装到默认路径<span class="filepath" id="zh-cn_topic_0000001930317932_filepath7570102017212"><a name="zh-cn_topic_0000001930317932_filepath7570102017212"></a><a name="zh-cn_topic_0000001930317932_filepath7570102017212"></a>“/usr/local/Ascend”</span>。</li><li>若通过该参数指定了安装目录，运行用户需要对指定的安装路径有读写权限。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1444404185013"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1144584125019"><a name="zh-cn_topic_0000001930317932_p1144584125019"></a><a name="zh-cn_topic_0000001930317932_p1144584125019"></a>--install-scene=&lt;scene&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p153510190174"><a name="zh-cn_topic_0000001930317932_p153510190174"></a><a name="zh-cn_topic_0000001930317932_p153510190174"></a><span id="zh-cn_topic_0000001930317932_ph1308455195116"><a name="zh-cn_topic_0000001930317932_ph1308455195116"></a><a name="zh-cn_topic_0000001930317932_ph1308455195116"></a>Ascend Docker Runtime</span>安装场景。<span id="zh-cn_topic_0000001930317932_ph1641213426170"><a name="zh-cn_topic_0000001930317932_ph1641213426170"></a><a name="zh-cn_topic_0000001930317932_ph1641213426170"></a>默认值为</span><span id="zh-cn_topic_0000001930317932_ph8821719135318"><a name="zh-cn_topic_0000001930317932_ph8821719135318"></a><a name="zh-cn_topic_0000001930317932_ph8821719135318"></a>docker，</span>取值说明如下。</p>
<a name="zh-cn_topic_0000001930317932_ul8352122811918"></a><a name="zh-cn_topic_0000001930317932_ul8352122811918"></a><ul id="zh-cn_topic_0000001930317932_ul8352122811918"><li><span id="zh-cn_topic_0000001930317932_ph3371331161710"><a name="zh-cn_topic_0000001930317932_ph3371331161710"></a><a name="zh-cn_topic_0000001930317932_ph3371331161710"></a>docker</span>：表示在<span id="zh-cn_topic_0000001930317932_ph1159416519530"><a name="zh-cn_topic_0000001930317932_ph1159416519530"></a><a name="zh-cn_topic_0000001930317932_ph1159416519530"></a>Docker</span>（或<span id="zh-cn_topic_0000001930317932_ph5391475179"><a name="zh-cn_topic_0000001930317932_ph5391475179"></a><a name="zh-cn_topic_0000001930317932_ph5391475179"></a>K8s集成Docker</span>）场景安装。</li><li><span id="zh-cn_topic_0000001930317932_ph7743733115213"><a name="zh-cn_topic_0000001930317932_ph7743733115213"></a><a name="zh-cn_topic_0000001930317932_ph7743733115213"></a>c</span><span id="zh-cn_topic_0000001930317932_ph1274373385212"><a name="zh-cn_topic_0000001930317932_ph1274373385212"></a><a name="zh-cn_topic_0000001930317932_ph1274373385212"></a>ontainerd：表示在</span>Containerd（或K8s集成Containerd）场景安装。</li><li>isula：表示在iSula容器引擎场景下安装。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row16570162013216"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1457092012114"><a name="zh-cn_topic_0000001930317932_p1457092012114"></a><a name="zh-cn_topic_0000001930317932_p1457092012114"></a>--uninstall</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p35701320182115"><a name="zh-cn_topic_0000001930317932_p35701320182115"></a><a name="zh-cn_topic_0000001930317932_p35701320182115"></a>卸载软件。如果安装时指定了安装路径，那么卸载时也需要指定安装路径，安装路径的参数为--install-path=&lt;path&gt;。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row757019209212"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p11570122092117"><a name="zh-cn_topic_0000001930317932_p11570122092117"></a><a name="zh-cn_topic_0000001930317932_p11570122092117"></a>--upgrade</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p5570720152111"><a name="zh-cn_topic_0000001930317932_p5570720152111"></a><a name="zh-cn_topic_0000001930317932_p5570720152111"></a>升级软件。如果安装时指定了安装路径，那么升级时也需要指定安装路径，安装路径的参数为--install-path=&lt;path&gt;。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row106534178110"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p17661618012"><a name="zh-cn_topic_0000001930317932_p17661618012"></a><a name="zh-cn_topic_0000001930317932_p17661618012"></a>--config-file-path=&lt;path&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p18661121811111"><a name="zh-cn_topic_0000001930317932_p18661121811111"></a><a name="zh-cn_topic_0000001930317932_p18661121811111"></a><span id="zh-cn_topic_0000001930317932_ph86621218919"><a name="zh-cn_topic_0000001930317932_ph86621218919"></a><a name="zh-cn_topic_0000001930317932_ph86621218919"></a>Docker</span>或<span id="zh-cn_topic_0000001930317932_ph196625181110"><a name="zh-cn_topic_0000001930317932_ph196625181110"></a><a name="zh-cn_topic_0000001930317932_ph196625181110"></a>Containerd</span>的配置文件路径。不指定该参数时默认使用以下路径。</p>
<a name="zh-cn_topic_0000001930317932_ul1666216181816"></a><a name="zh-cn_topic_0000001930317932_ul1666216181816"></a><ul id="zh-cn_topic_0000001930317932_ul1666216181816"><li><span id="zh-cn_topic_0000001930317932_ph146627186110"><a name="zh-cn_topic_0000001930317932_ph146627186110"></a><a name="zh-cn_topic_0000001930317932_ph146627186110"></a>Docker</span>: /etc/docker/daemon.json</li><li><span id="zh-cn_topic_0000001930317932_ph4662118513"><a name="zh-cn_topic_0000001930317932_ph4662118513"></a><a name="zh-cn_topic_0000001930317932_ph4662118513"></a>Containerd</span>: /etc/containerd/config.toml</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1857082012216"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p65701620122117"><a name="zh-cn_topic_0000001930317932_p65701620122117"></a><a name="zh-cn_topic_0000001930317932_p65701620122117"></a>--install-type=&lt;type&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><div class="p" id="zh-cn_topic_0000001930317932_p155774343616"><a name="zh-cn_topic_0000001930317932_p155774343616"></a><a name="zh-cn_topic_0000001930317932_p155774343616"></a>仅支持在以下产品安装或升级<span id="zh-cn_topic_0000001930317932_ph1796213135594"><a name="zh-cn_topic_0000001930317932_ph1796213135594"></a><a name="zh-cn_topic_0000001930317932_ph1796213135594"></a>Ascend Docker Runtime</span>时使用该参数：<a name="zh-cn_topic_0000001930317932_ul760551653710"></a><a name="zh-cn_topic_0000001930317932_ul760551653710"></a><ul id="zh-cn_topic_0000001930317932_ul760551653710"><li><span id="zh-cn_topic_0000001930317932_ph87811154145311"><a name="zh-cn_topic_0000001930317932_ph87811154145311"></a><a name="zh-cn_topic_0000001930317932_ph87811154145311"></a>Atlas 200 AI加速模块（RC场景）</span></li><li><span id="zh-cn_topic_0000001930317932_ph1851111042012"><a name="zh-cn_topic_0000001930317932_ph1851111042012"></a><a name="zh-cn_topic_0000001930317932_ph1851111042012"></a>Atlas 200I A2 加速模块</span>（RC场景）</li><li><span id="zh-cn_topic_0000001930317932_ph225916251208"><a name="zh-cn_topic_0000001930317932_ph225916251208"></a><a name="zh-cn_topic_0000001930317932_ph225916251208"></a>Atlas 200I DK A2 开发者套件</span></li><li><span id="zh-cn_topic_0000001930317932_ph271718714435"><a name="zh-cn_topic_0000001930317932_ph271718714435"></a><a name="zh-cn_topic_0000001930317932_ph271718714435"></a>Atlas 200I SoC A1 核心板</span></li><li><span id="zh-cn_topic_0000001930317932_ph12573124613552"><a name="zh-cn_topic_0000001930317932_ph12573124613552"></a><a name="zh-cn_topic_0000001930317932_ph12573124613552"></a>Atlas 500 智能小站（型号 3000）</span></li><li><span id="zh-cn_topic_0000001930317932_ph11710328131520"><a name="zh-cn_topic_0000001930317932_ph11710328131520"></a><a name="zh-cn_topic_0000001930317932_ph11710328131520"></a>Atlas 500 A2 智能小站</span></li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001930317932_p157201431201014"><a name="zh-cn_topic_0000001930317932_p157201431201014"></a><a name="zh-cn_topic_0000001930317932_p157201431201014"></a>该参数用于设置<span id="zh-cn_topic_0000001930317932_ph118353873517"><a name="zh-cn_topic_0000001930317932_ph118353873517"></a><a name="zh-cn_topic_0000001930317932_ph118353873517"></a>Ascend Docker Runtime</span>的默认挂载内容，且需要配合“--install”一起使用，格式为--install --install-type=&lt;type&gt;。&lt;type&gt;可选值为：<a name="zh-cn_topic_0000001930317932_ul848511715115"></a><a name="zh-cn_topic_0000001930317932_ul848511715115"></a><ul id="zh-cn_topic_0000001930317932_ul848511715115"><li>A200</li><li>A200ISoC</li><li>A200IA2（支持<span id="zh-cn_topic_0000001930317932_ph1323354011201"><a name="zh-cn_topic_0000001930317932_ph1323354011201"></a><a name="zh-cn_topic_0000001930317932_ph1323354011201"></a>Atlas 200I A2 加速模块</span>（RC场景）和<span id="zh-cn_topic_0000001930317932_ph192331940102018"><a name="zh-cn_topic_0000001930317932_ph192331940102018"></a><a name="zh-cn_topic_0000001930317932_ph192331940102018"></a>Atlas 200I DK A2 开发者套件</span>）</li><li>A500</li><li>A500A2</li></ul>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row14570162052115"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1857042012112"><a name="zh-cn_topic_0000001930317932_p1857042012112"></a><a name="zh-cn_topic_0000001930317932_p1857042012112"></a>--ce=&lt;ce&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><a name="zh-cn_topic_0000001930317932_ul4752351238"></a><a name="zh-cn_topic_0000001930317932_ul4752351238"></a><ul id="zh-cn_topic_0000001930317932_ul4752351238"><li>仅在使用<span id="zh-cn_topic_0000001930317932_ph137882109239"><a name="zh-cn_topic_0000001930317932_ph137882109239"></a><a name="zh-cn_topic_0000001930317932_ph137882109239"></a>iSula</span>启动容器时需要指定该参数，参数值为isula。并且需要配合--install或者--uninstall一起使用，不能单独使用。</li><li>不支持和--install-scene同时使用。建议使用--install-scene替代--ce参数。后续--ce会废弃。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1633572102619"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p733611211268"><a name="zh-cn_topic_0000001930317932_p733611211268"></a><a name="zh-cn_topic_0000001930317932_p733611211268"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p83361215264"><a name="zh-cn_topic_0000001930317932_p83361215264"></a><a name="zh-cn_topic_0000001930317932_p83361215264"></a>查询<span id="zh-cn_topic_0000001930317932_ph7723132765210"><a name="zh-cn_topic_0000001930317932_ph7723132765210"></a><a name="zh-cn_topic_0000001930317932_ph7723132765210"></a>Ascend Docker Runtime</span>版本。</p>
</td>
</tr>
</tbody>
</table>

### NPU Exporter<a name="ZH-CN_TOPIC_0000002511426331"></a>

-   使用**资源监测**时，必须安装NPU Exporter，该组件支持对接Prometheus或Telegraf。
    -   对接Prometheus时，支持通过容器和二进制两种方式部署NPU Exporter，部署差异可参考[容器和二进制部署差异](./appendix.md#容器和二进制部署差异)。
    -   对接Telegraf时，参考[通过Telegraf使用](./usage/resource_monitoring.md#通过telegraf使用)章节，安装NPU Exporter和Telegraf。

-   不使用**资源监测**的用户，可以不安装NPU Exporter，请直接跳过本章节。

**使用约束<a name="section1362795652416"></a>**

在安装NPU Exporter前，需要提前了解相关约束，具体说明请参见[表1](#table105071852271)。

**表 1**  约束说明

<a name="table105071852271"></a>
<table><thead align="left"><tr id="row2050719520272"><th class="cellrowborder" valign="top" width="29.970000000000002%" id="mcps1.2.3.1.1"><p id="p1950795152711"><a name="p1950795152711"></a><a name="p1950795152711"></a>约束场景</p>
</th>
<th class="cellrowborder" valign="top" width="70.03%" id="mcps1.2.3.1.2"><p id="p75071151277"><a name="p75071151277"></a><a name="p75071151277"></a>约束说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row115077513271"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p17925222411"><a name="p17925222411"></a><a name="p17925222411"></a>NPU驱动</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p450745142712"><a name="p450745142712"></a><a name="p450745142712"></a><span id="ph10112356112713"><a name="ph10112356112713"></a><a name="ph10112356112713"></a>NPU Exporter</span>会周期性调用NPU驱动的相关接口以检测NPU状态。如果要升级驱动，请先停止业务任务，再停止<span id="ph154413248375"><a name="ph154413248375"></a><a name="ph154413248375"></a>NPU Exporter</span>容器服务。</p>
<div class="note" id="note1993172317415"><a name="note1993172317415"></a><a name="note1993172317415"></a><span class="notetitle"> 说明： </span><div class="notebody"><div class="p" id="zh-cn_topic_0000002479226378_p18934232419"><a name="zh-cn_topic_0000002479226378_p18934232419"></a><a name="zh-cn_topic_0000002479226378_p18934232419"></a>为保证<span id="zh-cn_topic_0000002479226378_ph7206429154119"><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a>NPU Exporter</span>以二进制部署时可使用非root用户安装（如hwMindX），请在安装驱动时使用--install-for-all参数。示例如下。<pre class="screen" id="zh-cn_topic_0000002479226378_screen15239164112445"><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a>./Ascend-hdk-&lt;chip_type&gt;-npu-driver_&lt;version&gt;_linux-&lt;arch&gt;.run --full --install-for-all</pre>
</div>
</div></div>
</td>
</tr>
<tr id="row54685525282"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p5249201634114"><a name="p5249201634114"></a><a name="p5249201634114"></a><span id="ph1461172794116"><a name="ph1461172794116"></a><a name="ph1461172794116"></a>K8s</span>版本</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p5468852142813"><a name="p5468852142813"></a><a name="p5468852142813"></a>使用<span id="ph98079531286"><a name="ph98079531286"></a><a name="ph98079531286"></a>NPU Exporter</span>前需要确保环境的<span id="ph18807253152810"><a name="ph18807253152810"></a><a name="ph18807253152810"></a>K8s</span>版本，若<span id="ph6808453102813"><a name="ph6808453102813"></a><a name="ph6808453102813"></a>K8s</span>版本在1.24.x及以上版本，需要用户自行<a href="https://github.com/mirantis/cri-dockerd#build-and-install" target="_blank" rel="noopener noreferrer">安装cri-dockerd</a>依赖。</p>
</td>
</tr>
<tr id="row7507135142716"><td class="cellrowborder" rowspan="3" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p45071516276"><a name="p45071516276"></a><a name="p45071516276"></a>DCMI动态库</p>
<p id="p14507145152714"><a name="p14507145152714"></a><a name="p14507145152714"></a></p>
<p id="p9507651272"><a name="p9507651272"></a><a name="p9507651272"></a></p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p6555101612381"><a name="p6555101612381"></a><a name="p6555101612381"></a>DCMI动态库目录权限要求如下：</p>
<p id="p950745102715"><a name="p950745102715"></a><a name="p950745102715"></a><span id="ph1496251019288"><a name="ph1496251019288"></a><a name="ph1496251019288"></a>NPU Exporter</span>调用的DCMI动态库其所有父目录，需要满足属主为root，其他属主程序无法运行；同时，这些文件及其目录需满足group和other不具备写权限。</p>
</td>
</tr>
<tr id="row1650710572715"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p195079518272"><a name="p195079518272"></a><a name="p195079518272"></a>DCMI动态库路径深度必须小于20。</p>
</td>
</tr>
<tr id="row35071553276"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18507205192711"><a name="p18507205192711"></a><a name="p18507205192711"></a>如果通过设置LD_LIBRARY_PATH设置动态库路径，LD_LIBRARY_PATH环境变量总长度不能超过1024。</p>
</td>
</tr>
<tr id="row75074519275"><td class="cellrowborder" rowspan="2" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p050719519271"><a name="p050719519271"></a><a name="p050719519271"></a><span id="ph13135203152812"><a name="ph13135203152812"></a><a name="ph13135203152812"></a>Atlas 200I SoC A1 核心板</span></p>
<p id="p35076552719"><a name="p35076552719"></a><a name="p35076552719"></a></p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p209012054192411"><a name="p209012054192411"></a><a name="p209012054192411"></a><span id="ph56561935182816"><a name="ph56561935182816"></a><a name="ph56561935182816"></a>Atlas 200I SoC A1 核心板</span>使用<span id="ph1865633562811"><a name="ph1865633562811"></a><a name="ph1865633562811"></a>NPU Exporter</span>组件，需要确保<span id="ph10656153513282"><a name="ph10656153513282"></a><a name="ph10656153513282"></a>Atlas 200I SoC A1 核心板</span>的NPU驱动在23.0.RC2及以上版本。升级NPU驱动可参考<span id="ph19001377278"><a name="ph19001377278"></a><a name="ph19001377278"></a>《Atlas 200I SoC A1 核心板 25.0.RC1 NPU驱动和固件升级指导书》中“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100468879/dd05eaa6" target="_blank" rel="noopener noreferrer">升级驱动</a>”章节</span>进行操作。</p>
</td>
</tr>
<tr id="row165073518272"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p95251515257"><a name="p95251515257"></a><a name="p95251515257"></a><span id="ph19614124172819"><a name="ph19614124172819"></a><a name="ph19614124172819"></a>Atlas 200I SoC A1 核心板</span>节点上使用容器化部署<span id="ph136141041142813"><a name="ph136141041142813"></a><a name="ph136141041142813"></a>NPU Exporter</span>，需要配置多容器共享模式，具体请参考<span id="ph3957123242310"><a name="ph3957123242310"></a><a name="ph3957123242310"></a>《Atlas 200I SoC A1 核心板 25.0.RC1 NPU驱动和固件安装指南》中“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100468901/55e9d968" target="_blank" rel="noopener noreferrer">容器内运行</a>”章节</span>。</p>
</td>
</tr>
<tr id="row1044710113298"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p1144701142912"><a name="p1144701142912"></a><a name="p1144701142912"></a>虚拟机场景</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p14473110297"><a name="p14473110297"></a><a name="p14473110297"></a>如果在虚拟机场景下部署<span id="ph6368151492319"><a name="ph6368151492319"></a><a name="ph6368151492319"></a>NPU Exporter</span>，需要在<span id="ph24388313372"><a name="ph24388313372"></a><a name="ph24388313372"></a>NPU Exporter</span>的镜像中安装systemd，推荐在Dockerfile中加入<strong id="b14813193310547"><a name="b14813193310547"></a><a name="b14813193310547"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong>命令进行安装。</p>
</td>
</tr>
</tbody>
</table>

**操作步骤<a name="section83111543151612"></a>**

NPU Exporter支持两种安装方式，用户可根据实际情况选择其中一种进行安装。该组件仅提供HTTP服务，如需使用更为安全的HTTPS服务，请自行修改源码进行适配。

-   （推荐）以容器化方式运行，安装步骤参见[容器化方式运行](#section2035402135914)。
-   当安全要求较高时，建议在物理机上以二进制方式运行，安装步骤参见[二进制方式运行](#section103551921135917)。

**容器化方式运行<a name="section2035402135914"></a>**

1.  以root用户登录各计算节点。
2.  （可选）修改metricConfiguration.json或pluginConfiguration.json文件，配置默认指标组或自定义指标组采集和上报的开关。
    1.  进入NPU Exporter软件包解压目录。
    2.  <a name="li11364381194"></a>打开metricConfiguration.json文件。

        ```
        vi metricConfiguration.json
        ```

    3.  按“i”进入编辑模式，根据实际需要配置默认指标组采集和上报的开关。

        <a name="table192202574406"></a>
        <table><thead align="left"><tr id="row152204575408"><th class="cellrowborder" valign="top" width="30.12%" id="mcps1.1.3.1.1"><p id="p1220125712404"><a name="p1220125712404"></a><a name="p1220125712404"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="69.88%" id="mcps1.1.3.1.2"><p id="p622019575401"><a name="p622019575401"></a><a name="p622019575401"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="row182201357164014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="p152201573404"><a name="p152201573404"></a><a name="p152201573404"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="p222035704018"><a name="p222035704018"></a><a name="p222035704018"></a>默认指标组名称。</p>
        <a name="ul222055714012"></a><a name="ul222055714012"></a><ul id="ul222055714012"><li>ddr：DDR数据信息</li><li>hccs：HCCS数据信息</li><li>npu：NPU数据信息</li><li>network：Network数据信息</li><li>pcie：PCIe数据信息</li><li>roce：ReCE数据信息</li><li>sio：SIO数据信息</li><li>vnpu：vNPU数据信息</li><li>version：版本数据信息</li><li>optical：光模块数据信息</li><li>hbm：片上内存数据信息</li></ul>
        </td>
        </tr>
        <tr id="row5220257114014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="p182201657134015"><a name="p182201657134015"></a><a name="p182201657134015"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="p722015718403"><a name="p722015718403"></a><a name="p722015718403"></a>指标组采集和上报的开关。默认值为ON。</p>
        <a name="ul14220557134016"></a><a name="ul14220557134016"></a><ul id="ul14220557134016"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    4.  <a name="li151815494115"></a>按“Esc”键，输入:wq!保存并退出。
    5.  参考[2.b](#li11364381194)到[2.d](#li151815494115)，修改pluginConfiguration.json文件，根据实际需要配置自定义指标组采集和上报的开关。

        <a name="table970154420512"></a>
        <table><thead align="left"><tr id="row157015443510"><th class="cellrowborder" valign="top" width="23.14%" id="mcps1.1.3.1.1"><p id="p15701444553"><a name="p15701444553"></a><a name="p15701444553"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="76.86%" id="mcps1.1.3.1.2"><p id="p117011441156"><a name="p117011441156"></a><a name="p117011441156"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="row47010440518"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="p170118446517"><a name="p170118446517"></a><a name="p170118446517"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="p9568105120719"><a name="p9568105120719"></a><a name="p9568105120719"></a>向<span id="ph18671058476"><a name="ph18671058476"></a><a name="ph18671058476"></a>NPU Exporter</span>注册的自定义指标组名称。自定义指标的方法详细请参见<a href="./appendix.md#自定义指标开发">自定义指标开发</a>。</p>
        </td>
        </tr>
        <tr id="row157010441654"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="p157021644755"><a name="p157021644755"></a><a name="p157021644755"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="p2148170172513"><a name="p2148170172513"></a><a name="p2148170172513"></a>指标组采集和上报的开关。默认值为OFF。</p>
        <a name="ul1870217441514"></a><a name="ul1870217441514"></a><ul id="ul1870217441514"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    6.  若通过插件方式开发了自定义指标，需重新构建编译二进制文件。
    7.  参见[准备镜像](#准备镜像)，重新进行镜像制作和分发，然后执行[4](#li0640635114211)。

3.  查看NPU Exporter镜像和版本号是否正确。
    -   **Docker场景**：执行如下命令。

        ```
        docker images | grep npu-exporter
        ```

        回显示例如下。

        ```
        npu-exporter                         v7.3.0              20185c45f1bc        About an hour ago         90.1MB
        ```

    -   **Containerd场景**：执行如下命令。

        ```
        ctr -n k8s.io c ls | grep npu-exporter
        ```

        回显示例如下。

        ```
        docker.io/library/npu-exporter:v7.3.0                                                         application/vnd.docker.distribution.manifest.v2+json      sha256:38fd69ee9f5753e73a55a216d039f6ed4ea8a5de15c0e6b3bb503022db470c7b 91.5 MiB  linux/arm64 
        ```

    -   是，执行[4](#li0640635114211)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作和分发。

4.  <a name="li0640635114211"></a>将NPU Exporter软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
5.  请根据实际使用的容器化方式，选择执行以下步骤。
    -   **Containerd场景**：需要将containerMode设置为containerd，并对以下代码进行修改。

        如果使用默认的NPU Exporter启动参数“-containerMode=docker”时，可跳过本步骤。

        ```
        apiVersion: apps/v1
        kind: DaemonSet
        metadata:
          name: npu-exporter
          namespace: npu-exporter
        spec:
          selector:
            matchLabels:
              app: npu-exporter
        ...
            spec:
        ...
              args: [ "umask 027;npu-exporter -port=8082 -ip=0.0.0.0  -updateTime=5
                         -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log -logLevel=0 -containerMode=containerd" ]
        ...
              volumeMounts:
        ...
                - name: docker-shim                                       
                  mountPath: /var/run/dockershim.sock
                  readOnly: true
                - name: docker                                       # 仅使用containerd时删除
                  mountPath: /var/run/docker
                  readOnly: true
                - name: cri-dockerd                                 
                  mountPath: /var/run/cri-dockerd.sock
                  readOnly: true
                - name: containerd                             
                  mountPath: /run/containerd
                  readOnly: true
                - name: isulad                                
                  mountPath: /run/isulad.sock
                  readOnly: true
        ...
              volumes:
        ...
                - name: docker-shim                             
                  hostPath:
                    path: /var/run/dockershim.sock
                - name: docker                                # 仅使用containerd时删除
                  hostPath:
                    path: /var/run/docker
                - name: cri-dockerd                           
                  hostPath:
                    path: /var/run/cri-dockerd.sock
                - name: containerd                            
                  hostPath:
                    path: /run/containerd
                - name: isulad                               
                  hostPath:
                    path: /run/isulad.sock
        
        ...
        ```

    -   **Docker场景**：删除原有容器运行时的挂载文件，新增dockershim.sock文件的挂载目录，并对以下代码进行修改。

        如果使用的NPU Exporter启动参数“-containerMode=containerd”，可跳过本步骤。

        >[!NOTICE] 须知 
        >该步骤可有效解决kubelet重启后，造成的NPU Exporter数据丢失问题。新增挂载目录后，会同时新增很多挂载文件，如docker.sock，有容器逃逸的风险。

        ```
        ...
                volumeMounts:
                  - name: log-npu-exporter
        ...
                  - name: sys
                    mountPath: /sys
                    readOnly: true
                  - name: docker-shim                        # 删除以下字段
                    mountPath: /var/run/dockershim.sock
                    readOnly: true
                  - name: docker 
                    mountPath: /var/run/docker
                    readOnly: true
                  - name: cri-dockerd 
                    mountPath: /var/run/cri-dockerd.sock
                    readOnly: true
                  - name: sock                   # 新增以下字段
                    mountPath: /var/run        # 以实际的dockershim.sock文件目录为准
                  - name: containerd  
                    mountPath: /run/containerd
        ...
              volumes:
                - name: log-npu-exporter
        ...
                - name: sys
                  hostPath:
                    path: /sys
                - name: docker-shim                    # 删除以下字段
                  hostPath:   
                    path: /var/run/dockershim.sock
                - name: docker 
                  hostPath:
                    path: /var/run/docker
                - name: cri-dockerd 
                  hostPath:
                    path: /var/run/cri-dockerd.sock
                - name: sock                 # 新增以下字段
                  hostPath:
                    path: /var/run                    # 以实际的dockershim.sock文件目录为准
                - name: containerd  
                  hostPath:
                    path: /run/containerd
         ...
        ```

6.  如不修改组件的其他启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中NPU Exporter的启动参数。启动参数如[表2](#table872410431914)所示，也可执行<b>./npu-exporter -h</b>查看参数说明。
7.  在管理节点的YAML所在路径，执行以下命令，启动NPU Exporter。

    -   K8s集群中使用Atlas 200I SoC A1 核心板节点，执行以下命令。

        ```
        kubectl apply -f npu-exporter-310P-1usoc-v{version}.yaml
        ```

    -   K8s集群中使用除Atlas 200I SoC A1 核心板外的其他类型节点，执行以下命令。

        ```
        kubectl apply -f npu-exporter-v{version}.yaml
        ```

    启动示例如下：

    ```
    namespace/npu-exporter created
    networkpolicy.networking.K8s.io/exporter-network-policy created
    daemonset.apps/npu-exporter created
    ```

8.  在任意节点执行以下命令，查看组件是否启动成功。

    ```
    kubectl get pod -n npu-exporter
    ```

    回显示例如下，出现**Running**表示组件启动成功。若状态为**CrashLoopBackOff**，可能是因为目录权限不正确导致，可以参见[NPU Exporter检查动态路径失败，日志出现check uid or mode failed](./faq.md#npu-exporter检查动态路径失败日志出现check-uid-or-mode-failed)章节进行处理。

    ```
    NAME                            READY   STATUS    RESTARTS   AGE
    ...
    npu-exporter-hqpxl        1/1    Running   0        11s
    ```

    >![](public_sys-resources/icon-note.gif) **说明：** 
    >-   NPU Exporter的使用对进程环境有要求，以容器形式运行时，请确保“/sys“目录和容器运行时通信socket文件挂载至NPU Exporter容器中。若通过调用NPU Exporter的Metrics接口，没有获取到NPU容器的相关信息，该问题可能是因为socket文件路径不正确导致，可以参见[日志出现connecting to container runtime failed](./faq.md#日志出现connecting-to-container-runtime-failed)章节进行处理。
    >-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
    >-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
    >-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
    >-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**二进制方式运行<a name="section103551921135917"></a>**

NPU Exporter组件以容器化方式运行时需使用特权容器、root用户和挂载了docker-shim或Containerd的socket文件，如果容器被人恶意利用，有容器逃逸风险。当安全性要求较高时，可直接在物理机上通过二进制方式运行。

>[!NOTE] 说明 
>-   以二进制方式部署NPU Exporter时，可以使用非root用户（例如hwMindX）进行部署。请将日志目录权限修改为hwMindX，命令示例如下：**chown  _hwMindX:hwMindX_  /var/log/mindx-dl/npu-exporter**。
>-   下文步骤中的用户均为hwMindX。

1.  使用root用户登录服务器。
2.  将NPU Exporter软件包上传至服务器的任意目录（如“/home/ascend-npu-exporter”）并进行解压操作。
3.  将NPU Exporter软件包解压目录下的metricConfiguration.json和pluginConfiguration.json文件，拷贝到“/usr/local”目录下。
4.  （可选）修改metricConfiguration.json或pluginConfiguration.json文件，配置默认指标组或自定义指标组采集和上报的开关。
    1.  进入“/usr/local”目录。
    2.  <a name="li1445835411478"></a>打开metricConfiguration.json文件。

        ```
        vi metricConfiguration.json
        ```

    3.  按“i”进入编辑模式，根据实际需要配置默认指标组采集和上报的开关。

        <a name="zh-cn_topic_0000002511426331_table192202574406"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002511426331_row152204575408"><th class="cellrowborder" valign="top" width="30.12%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000002511426331_p1220125712404"><a name="zh-cn_topic_0000002511426331_p1220125712404"></a><a name="zh-cn_topic_0000002511426331_p1220125712404"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="69.88%" id="mcps1.1.3.1.2"><p id="zh-cn_topic_0000002511426331_p622019575401"><a name="zh-cn_topic_0000002511426331_p622019575401"></a><a name="zh-cn_topic_0000002511426331_p622019575401"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002511426331_row182201357164014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p152201573404"><a name="zh-cn_topic_0000002511426331_p152201573404"></a><a name="zh-cn_topic_0000002511426331_p152201573404"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p222035704018"><a name="zh-cn_topic_0000002511426331_p222035704018"></a><a name="zh-cn_topic_0000002511426331_p222035704018"></a>默认指标组名称。</p>
        <a name="zh-cn_topic_0000002511426331_ul222055714012"></a><a name="zh-cn_topic_0000002511426331_ul222055714012"></a><ul id="zh-cn_topic_0000002511426331_ul222055714012"><li>ddr：DDR数据信息</li><li>hccs：HCCS数据信息</li><li>npu：NPU数据信息</li><li>network：Network数据信息</li><li>pcie：PCIe数据信息</li><li>roce：ReCE数据信息</li><li>sio：SIO数据信息</li><li>vnpu：vNPU数据信息</li><li>version：版本数据信息</li><li>optical：光模块数据信息</li><li>hbm：片上内存数据信息</li></ul>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_row5220257114014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p182201657134015"><a name="zh-cn_topic_0000002511426331_p182201657134015"></a><a name="zh-cn_topic_0000002511426331_p182201657134015"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p722015718403"><a name="zh-cn_topic_0000002511426331_p722015718403"></a><a name="zh-cn_topic_0000002511426331_p722015718403"></a>指标组采集和上报的开关。默认值为ON。</p>
        <a name="zh-cn_topic_0000002511426331_ul14220557134016"></a><a name="zh-cn_topic_0000002511426331_ul14220557134016"></a><ul id="zh-cn_topic_0000002511426331_ul14220557134016"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    4.  <a name="li18459954104718"></a>按“Esc”键，输入:wq!保存并退出。
    5.  参考[4.b](#li1445835411478)到[4.d](#li18459954104718)，修改pluginConfiguration.json文件，根据实际需要配置自定义指标组采集和上报的开关。

        <a name="table16459165464719"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002511426331_row157015443510"><th class="cellrowborder" valign="top" width="23.14%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000002511426331_p15701444553"><a name="zh-cn_topic_0000002511426331_p15701444553"></a><a name="zh-cn_topic_0000002511426331_p15701444553"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="76.86%" id="mcps1.1.3.1.2"><p id="zh-cn_topic_0000002511426331_p117011441156"><a name="zh-cn_topic_0000002511426331_p117011441156"></a><a name="zh-cn_topic_0000002511426331_p117011441156"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002511426331_row47010440518"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p170118446517"><a name="zh-cn_topic_0000002511426331_p170118446517"></a><a name="zh-cn_topic_0000002511426331_p170118446517"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p9568105120719"><a name="zh-cn_topic_0000002511426331_p9568105120719"></a><a name="zh-cn_topic_0000002511426331_p9568105120719"></a>向<span id="zh-cn_topic_0000002511426331_ph18671058476"><a name="zh-cn_topic_0000002511426331_ph18671058476"></a><a name="zh-cn_topic_0000002511426331_ph18671058476"></a>NPU Exporter</span>注册的自定义指标组名称。自定义指标的方法详细请参见<a href="./appendix.md#自定义指标开发">自定义指标开发</a>。</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_row157010441654"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p157021644755"><a name="zh-cn_topic_0000002511426331_p157021644755"></a><a name="zh-cn_topic_0000002511426331_p157021644755"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p2148170172513"><a name="zh-cn_topic_0000002511426331_p2148170172513"></a><a name="zh-cn_topic_0000002511426331_p2148170172513"></a>指标组采集和上报的开关。默认值为OFF。</p>
        <a name="zh-cn_topic_0000002511426331_ul1870217441514"></a><a name="zh-cn_topic_0000002511426331_ul1870217441514"></a><ul id="zh-cn_topic_0000002511426331_ul1870217441514"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    6.  若通过插件方式开发了自定义指标，需重新构建编译二进制文件。

5.  创建并编辑npu-exporter.service文件。
    1.  执行以下命令，创建npu-exporter.service文件。

        ```
        vi /home/ascend-npu-exporter/npu-exporter.service
        ```

    2.  参考如下内容，写入npu-exporter.service文件中。

        ```
        [Unit]
        Description=Ascend npu exporter
        Documentation=hiascend.com
        
        [Service]
        ExecStart=/bin/bash -c "/usr/local/bin/npu-exporter -ip=127.0.0.1 -port=8082 -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log>/dev/null  2>&1 &"
        Restart=always
        RestartSec=2
        KillMode=process
        Environment="GOGC=50"
        Environment="GOMAXPROCS=2"
        Environment="GODEBUG=madvdontneed=1"
        Type=forking
        User=hwMindX
        Group=hwMindX
        
        [Install]
        WantedBy=multi-user.target
        ```
        NPU Exporter默认情况只侦听127.0.0.1，可通过修改的启动参数“-ip”和“npu-exporter.service”文件的“ExecStart”字段修改需要侦听的IP地址。

    3.  按“Esc”键，输入:wq!保存并退出。

6.  创建并编辑npu-exporter.timer文件。通过配置timer延时启动，可保证NPU Exporter启动时NPU卡已就位。
    1.  执行以下命令，创建npu-exporter.timer文件。

        ```
         vi /home/ascend-npu-exporter/npu-exporter.timer
        ```

    2.  参考以下示例，并将其写入npu-exporter.timer文件中。

        ```
        [Unit]
        Description=Timer for NPU Exporter Service
        
        [Timer]
        OnBootSec=60s            # 设置NPU Exporter延时启动时间，请根据实际情况调整
        Unit=npu-exporter.service
        
        [Install]
        WantedBy=timers.target
        ```

    3.  按“Esc”键，输入:wq!保存并退出。

7.  若部署节点为Atlas 200I SoC A1 核心板，请依次执行以下命令，在节点上将hwMindX用户加入到HwBaseUser、HwDmUser用户组中。非Atlas 200I SoC A1 核心板用户，可跳过本步骤。

    ```
    usermod -a -G HwBaseUser hwMindX
    usermod -a -G HwDmUser hwMindX
    ```

8.  依次执行以下命令，启用NPU Exporter服务。

    ```
    cd /home/ascend-npu-exporter
    cp npu-exporter /usr/local/bin
    cp npu-exporter.service /etc/systemd/system
    chattr +i /etc/systemd/system/npu-exporter.service
    cp npu-exporter.timer /etc/systemd/system     
    chattr +i /etc/systemd/system/npu-exporter.timer      
    chmod 500 /usr/local/bin/npu-exporter
    chown hwMindX:hwMindX /usr/local/bin/npu-exporter
    chattr +i /usr/local/bin/npu-exporter
    systemctl enable npu-exporter.timer 
    systemctl start npu-exporter
    systemctl start npu-exporter.timer
    ```

    > [!NOTE] 说明
    >如果需要获取容器相关数据信息，NPU Exporter需要临时提权以便于和CRI、OCI的socket建立连接，需要执行以下命令。
    >```
    >chattr -i /usr/local/bin/npu-exporter
    >setcap cap_setuid+ep /usr/local/bin/npu-exporter
    >chattr +i /usr/local/bin/npu-exporter
    >systemctl restart npu-exporter
    >```

**参数说明<a name="section2042611570392"></a>**

**表 2** NPU Exporter启动参数

<a name="table872410431914"></a>
|参数|类型|默认值|说明|
|--|--|--|--|
|-port|int|8082|侦听端口，取值范围为1025~40000。|
|-updateTime|int|5|信息更新周期1~60秒。如果设置的时间过长，一些生存时间小于更新周期的容器可能无法上报。|
|-ip|string|无|参数无默认值，必须配置。<p>侦听IP地址，在多网卡主机上不建议配置成0.0.0.0。</p>|
|-version|bool|false|是否查询NPU Exporter版本号。<ul><li>true：查询。</li><li>false：不查询。</li></ul>|
|-concurrency|int|5|HTTP服务的限流大小，默认5个并发，取值范围为1~512。|
|-logLevel|int|0|日志级别：<ul><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>|
|-maxAge|int|7|日志备份时间，取值范围为7~700，单位为天。|
|-logFile|string|/var/log/mindx-dl/npu-exporter/npu-exporter.log|日志文件。<p>单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：npu-exporter-触发转储的时间.log，如：npu-exporter-2023-10-07T03-38-24.402.log。</p>|
|-maxBackups|int|30|转储后日志文件保留个数上限，取值范围为1~30，单位为个。|
|-containerMode|string|docker|设置容器运行时类型。<ul><li>设置为docker表示当前环境使用Docker作为容器运行时。</li><li>设置为containerd表示当前环境使用Containerd作为容器运行时。</li><li>设置为“isula”表示当前环境使用iSula作为容器运行时。</li></ul>|
|-containerd|string|<ul><li>(Docker)unix：/run/docker/containerd/docker-containerd.sock</li><li>(Containerd)unix：///run/containerd/containerd.sock</li><li>(iSula)unix：///run/isulad.sock</li></ul>|containerd daemon进程endpoint，用于与Containerd通信。<ul><li>若containerMode=docker，则默认值为/run/docker/containerd/docker-containerd.sock；连接失败后，自动尝试连接：unix：///run/containerd/containerd.sock和unix:///run/docker/containerd/containerd.sock。</li><li>若containerMode=containerd，则默认值为/run/containerd/containerd.sock。</li><li>若containerMode=isula，则默认值为/run/isulad.sock。</li></ul><p>一般情况下使用默认值即可。若用户自行修改了Containerd的sock文件路径则需要进行相应路径的修改。</p><p>可通过**ps aux \| grep containerd**命令查询Containerd的sock文件路径是否修改。</p>|
|-endpoint|string|<ul><li>(Docker)unix：///var/run/dockershim.sock</li><li>(Containerd)unix：///run/containerd/containerd.sock</li><li>(iSula)unix：///run/isulad.sock</li></ul>|CRI server的sock地址：<ul><li>若containerMode=docker，将连接到Dockershim获取容器列表，默认值/var/run/dockershim.sock；</li><li>若containerMode=containerd，默认值/run/containerd/containerd.sock。</li><li>若containerMode=isula，则默认值为/run/isulad.sock。</li></ul><p>一般情况下使用默认值即可，除非用户自行修改了Dockershim或者Containerd的sock文件路径。</p><p>连接失败后，自动尝试连接unix:///run/cri-dockerd.sock</p>|
|-limitIPConn|int|5|每个IP的TCP限制数的取值范围为1~128。|
|-limitTotalConn|int|20|程序总共的TCP限制数的取值范围为1~512。|
|-limitIPReq|string|20/1|每个IP的请求限制数，20/1表示1秒限制20个请求，“/”两侧最大只支持三位数。|
|-cacheSize|int|102400|缓存key的数量限制，取值范围为1~1024000。|
|-h或者-help|无|无|显示帮助信息。|
|-platform|string|Prometheus|指定对接平台。<ul><li>Prometheus：对接Prometheus</li><li>Telegraf：对接Telegraf</li></ul>|
|-poll_interval|duration(int)|1|Telegraf数据上报的间隔时间，单位：秒。此参数在对接Telegraf平台时才起作用，即需要指定-platform=Telegraf时才生效，否则该参数不生效。|
|-profilingTime|int|200|配置采集PCIe带宽时间，单位：毫秒，取值范围为1~2000。|
|-hccsBWProfilingTime|int|200|HCCS链路带宽采样时长，取值范围1~1000，单位：毫秒。|
|-deviceResetTimeout|int|60|组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600。<ul><li>Atlas A2 训练系列产品、Atlas 800I A2 推理服务器、A200I A2 Box 异构组件：建议配置为150秒。</li><li>Atlas A3 训练系列产品、A200T A3 Box8 超节点服务器、Atlas 800I A3 超节点服务器：建议配置为360秒。</li><li>推理服务器（插Atlas 350 标卡）、Atlas 850 服务器、Atlas 950 SuperPoD 超节点/集群：建议配置为600秒。</li></ul>|
|-textMetricsFilePath|string|无|指定自定义指标文件的路径，其约束说明详细请参见<a href="./api/npu_exporter.md#自定义指标文件">约束说明</a>。|


### Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426341"></a>

-   使用整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须在计算节点安装Ascend Device Plugin。
-   仅使用容器化支持和资源监测的用户，可以不安装Ascend Device Plugin，请直接跳过本章节。

**使用约束<a name="section1362795652416"></a>**

在安装Ascend Device Plugin前，需要提前了解相关约束，具体说明请参见[表1](#table113813012140)。

**表 1**  约束说明

<a name="table113813012140"></a>
<table><thead align="left"><tr id="row193815031414"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.3.1.1"><p id="p13383051411"><a name="p13383051411"></a><a name="p13383051411"></a>约束场景</p>
</th>
<th class="cellrowborder" valign="top" width="75%" id="mcps1.2.3.1.2"><p id="p73814015146"><a name="p73814015146"></a><a name="p73814015146"></a>约束说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row738802142"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p13388019145"><a name="p13388019145"></a><a name="p13388019145"></a>NPU驱动</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p73819019145"><a name="p73819019145"></a><a name="p73819019145"></a><span id="ph11461134318147"><a name="ph11461134318147"></a><a name="ph11461134318147"></a>Ascend Device Plugin</span>会周期性调用NPU驱动的相关接口。如果要升级驱动，请先停止业务任务，再停止<span id="ph1546116433149"><a name="ph1546116433149"></a><a name="ph1546116433149"></a>Ascend Device Plugin</span>容器服务。</p>
</td>
</tr>
<tr id="row5531349229"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p6691413112218"><a name="p6691413112218"></a><a name="p6691413112218"></a>配合<span id="ph14695135229"><a name="ph14695135229"></a><a name="ph14695135229"></a>Ascend Docker Runtime</span>使用</p>
<p id="p6920163951110"><a name="p6920163951110"></a><a name="p6920163951110"></a></p>
<p id="p1920153951110"><a name="p1920153951110"></a><a name="p1920153951110"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p175159351335"><a name="p175159351335"></a><a name="p175159351335"></a>组件安装顺序要求如下：</p>
<p id="p1745811135313"><a name="p1745811135313"></a><a name="p1745811135313"></a><span id="ph197011318223"><a name="ph197011318223"></a><a name="ph197011318223"></a>Ascend Device Plugin</span>容器化运行时会自动识别是否安装了<span id="ph18701713102214"><a name="ph18701713102214"></a><a name="ph18701713102214"></a>Ascend Docker Runtime</span>，需要优先安装<span id="ph167081312219"><a name="ph167081312219"></a><a name="ph167081312219"></a>Ascend Docker Runtime</span>后<span id="ph207041311225"><a name="ph207041311225"></a><a name="ph207041311225"></a>Ascend Device Plugin</span>才能正确识别<span id="ph11701013152214"><a name="ph11701013152214"></a><a name="ph11701013152214"></a>Ascend Docker Runtime</span>的安装情况。</p>
<p id="p019819298377"><a name="p019819298377"></a><a name="p019819298377"></a><span id="ph06381714397"><a name="ph06381714397"></a><a name="ph06381714397"></a>Ascend Device Plugin</span>若部署在<span id="ph66321793918"><a name="ph66321793918"></a><a name="ph66321793918"></a>Atlas 200I SoC A1 核心板</span>上，无需安装<span id="ph116721035125612"><a name="ph116721035125612"></a><a name="ph116721035125612"></a>Ascend Docker Runtime</span>。</p>
</td>
</tr>
<tr id="row1648094416218"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p9484175210212"><a name="p9484175210212"></a><a name="p9484175210212"></a>组件版本要求如下：</p>
<p id="p44813447213"><a name="p44813447213"></a><a name="p44813447213"></a>该功能要求<span id="ph196135501025"><a name="ph196135501025"></a><a name="ph196135501025"></a>Ascend Docker Runtime</span>与<span id="ph1161319502212"><a name="ph1161319502212"></a><a name="ph1161319502212"></a>Ascend Device Plugin</span>版本保持一致且需要为5.0.RC1及以上版本，安装或卸载<span id="ph1361319501123"><a name="ph1361319501123"></a><a name="ph1361319501123"></a>Ascend Docker Runtime</span>之后需要重启容器引擎才能使<span id="ph11613850625"><a name="ph11613850625"></a><a name="ph11613850625"></a>Ascend Device Plugin</span>正确识别。</p>
</td>
</tr>
<tr id="row1449218752210"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p14704133226"><a name="p14704133226"></a><a name="p14704133226"></a>以下2种场景不支持<span id="ph1371171332212"><a name="ph1371171332212"></a><a name="ph1371171332212"></a>Ascend Device Plugin</span>和<span id="ph071513132214"><a name="ph071513132214"></a><a name="ph071513132214"></a>Ascend Docker Runtime</span>配合使用。<a name="ul1771141362211"></a><a name="ul1771141362211"></a><ul id="ul1771141362211"><li>混插场景。</li><li><span id="ph1471111314226"><a name="ph1471111314226"></a><a name="ph1471111314226"></a>Atlas 200I SoC A1 核心板</span>。</li></ul>
</div>
</td>
</tr>
<tr id="row5381205148"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p16384020141"><a name="p16384020141"></a><a name="p16384020141"></a>DCMI动态库</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p67821743113213"><a name="p67821743113213"></a><a name="p67821743113213"></a>DCMI动态库目录权限要求如下：</p>
<p id="p1238120191413"><a name="p1238120191413"></a><a name="p1238120191413"></a><span id="ph285261461515"><a name="ph285261461515"></a><a name="ph285261461515"></a>Ascend Device Plugin</span>调用的DCMI动态库及其所有父目录，需要满足属主为root，其他属主程序无法运行；同时，这些文件及其目录需满足group和other不具备写权限。</p>
</td>
</tr>
<tr id="row1138160191419"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1138180101418"><a name="p1138180101418"></a><a name="p1138180101418"></a>DCMI动态库路径深度必须小于20。</p>
</td>
</tr>
<tr id="row338407145"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1739170161413"><a name="p1739170161413"></a><a name="p1739170161413"></a>如果通过设置LD_LIBRARY_PATH设置动态库路径，LD_LIBRARY_PATH环境变量总长度不能超过1024。</p>
</td>
</tr>
<tr id="row11391707149"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p133919013143"><a name="p133919013143"></a><a name="p133919013143"></a><span id="ph1078193611515"><a name="ph1078193611515"></a><a name="ph1078193611515"></a>Atlas 200I SoC A1 核心板</span></p>
<p id="p1918223205014"><a name="p1918223205014"></a><a name="p1918223205014"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p786843510309"><a name="p786843510309"></a><a name="p786843510309"></a><span id="ph1480005781518"><a name="ph1480005781518"></a><a name="ph1480005781518"></a>Atlas 200I SoC A1 核心板</span>节点上如果使用容器化部署<span id="ph080185715158"><a name="ph080185715158"></a><a name="ph080185715158"></a>Ascend Device Plugin</span>，需要配置多容器共享模式，具体请参考<span id="ph3957123242310"><a name="ph3957123242310"></a><a name="ph3957123242310"></a>《Atlas 200I SoC A1 核心板 25.0.RC1 NPU驱动和固件安装指南》中“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100468901/55e9d968" target="_blank" rel="noopener noreferrer">容器内运行</a>”章节</span>。</p>
</td>
</tr>
<tr id="row4248144116153"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p5840775161"><a name="p5840775161"></a><a name="p5840775161"></a><span id="ph697712515161"><a name="ph697712515161"></a><a name="ph697712515161"></a>Atlas 200I SoC A1 核心板</span>使用<span id="ph99771752169"><a name="ph99771752169"></a><a name="ph99771752169"></a>Ascend Device Plugin</span>组件，需要遵循以下配套关系：<a name="ul2977251161"></a><a name="ul2977251161"></a><ul id="ul2977251161"><li>5.0.RC2版本的<span id="ph49779571614"><a name="ph49779571614"></a><a name="ph49779571614"></a>Ascend Device Plugin</span>需要配合<span id="ph5977135101614"><a name="ph5977135101614"></a><a name="ph5977135101614"></a>Atlas 200I SoC A1 核心板</span>的23.0.RC2及其之后的驱动一起使用。</li><li>5.0.RC2之前版本的<span id="ph59771512164"><a name="ph59771512164"></a><a name="ph59771512164"></a>Ascend Device Plugin</span>只能和<span id="ph1977115181612"><a name="ph1977115181612"></a><a name="ph1977115181612"></a>Atlas 200I SoC A1 核心板</span>的23.0.RC2之前的驱动一起使用。</li></ul>
</div>
</td>
</tr>
<tr id="row14538194431511"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p45382449151"><a name="p45382449151"></a><a name="p45382449151"></a>虚拟机场景</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p8538144420153"><a name="p8538144420153"></a><a name="p8538144420153"></a>如果在虚拟机场景下部署<span id="ph142915347164"><a name="ph142915347164"></a><a name="ph142915347164"></a>Ascend Device Plugin</span>，需要在<span id="ph0429634121617"><a name="ph0429634121617"></a><a name="ph0429634121617"></a>Ascend Device Plugin</span>的镜像中安装systemd，推荐在Dockerfile中加入<strong id="b93339419563"><a name="b93339419563"></a><a name="b93339419563"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong>命令进行安装。</p>
</td>
</tr>
<tr id="row1150514563377"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p450675616371"><a name="p450675616371"></a><a name="p450675616371"></a>重启场景</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p105070566371"><a name="p105070566371"></a><a name="p105070566371"></a>若用户在安装<span id="ph444301153912"><a name="ph444301153912"></a><a name="ph444301153912"></a>Ascend Device Plugin</span>后，又重新修改了NPU的基础信息，例如修改了device ip，则需要重启<span id="ph52417305424"><a name="ph52417305424"></a><a name="ph52417305424"></a>Ascend Device Plugin</span>，否则<span id="ph23611038174213"><a name="ph23611038174213"></a><a name="ph23611038174213"></a>Ascend Device Plugin</span>不能正确识别NPU的相关信息。</p>
</td>
</tr>
</tbody>
</table>

**操作步骤<a name="section71204451253"></a>**

1.  以root用户登录各计算节点，并执行以下命令查看镜像和版本号是否正确。

    ```
    docker images | grep k8sdeviceplugin
    ```

    回显示例如下：

    ```
    ascend-k8sdeviceplugin               v7.3.0              29eec79eb693        About an hour ago   105MB
    ```

    -   是，执行[步骤2](#zh-cn_topic_0000001497364849_li922154411117)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作和分发。

2.  <a name="zh-cn_topic_0000001497364849_li922154411117"></a>将Ascend Device Plugin软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。请注意此处需使用适配具体处理器型号的YAML文件，并且为了避免自动识别Ascend Docker Runtime功能出现异常，请勿修改YAML文件中DaemonSet.metadata.name字段，详见下表。

    **表 2** Ascend Device Plugin的YAML文件列表

    <a name="zh-cn_topic_0000001497364849_table58619457211"></a>
    |YAML文件列表|说明|
    |--|--|
    |device-plugin-310-v*{version}*.yaml|推理服务器（插Atlas 300I 推理卡）上不使用Volcano的配置文件。|
    |device-plugin-310-volcano-v*{version}*.yaml|推理服务器（插Atlas 300I 推理卡）上使用Volcano的配置文件。|
    |device-plugin-310P-1usoc-*v{version}*.yaml|Atlas 200I SoC A1 核心板上不使用Volcano的配置文件。|
    |device-plugin-310P-1usoc-volcano-v*{version}*.yaml|Atlas 200I SoC A1 核心板上使用Volcano的配置文件。|
    |device-plugin-310P-*v{version}*.yaml|Atlas 推理系列产品上不使用Volcano的配置文件。|
    |device-plugin-310P-volcano-*v{version}*.yaml|Atlas 推理系列产品上使用Volcano的配置文件。|
    |device-plugin-910-v*{version}*.yaml|Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件上不使用Volcano的配置文件。|
    |device-plugin-volcano-*v{version}*.yaml|Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件上使用Volcano的配置文件。|

3.  如不修改组件启动参数，可跳过本步骤。否则，根据实际情况修改Ascend Device Plugin的启动参数。启动参数请参见[表3](#table1064314568229)，可执行<b>./device-plugin -h</b>查看参数说明。
    -   在Atlas 200I SoC A1 核心板节点上，修改启动脚本“run\_for\_310P\_1usoc.sh”中Ascend Device Plugin的启动参数。修改完后需在所有Atlas 200I SoC A1 核心板节点上重新制作镜像，或者将本节点镜像重新制作后分发到其余所有Atlas 200I SoC A1 核心板节点。

        >[!NOTE] 说明 
        >如果不使用Volcano作为调度器，在启动Ascend Device Plugin的时候，需要修改“run\_for\_310P\_1usoc.sh”中Ascend Device Plugin的启动参数，将“-volcanoType”参数设置为false。

    -   其他类型节点，修改对应启动YAML文件中Ascend Device Plugin的启动参数。

4.  （可选）使用**断点续训**（包括进程级恢复）或**弹性训练**时，根据需要使用的故障处理模式，修改Ascend Device Plugin组件的启动YAML。

    ```
    ...
          containers:
          - image: ascend-k8sdeviceplugin:v7.3.0
            name: device-plugin-01
            resources:
              requests:
                memory: 500Mi
                cpu: 500m
              limits:
                memory: 500Mi
                cpu: 500m
            command: [ "/bin/bash", "-c", "--"]
            args: [ "device-plugin  
                     -useAscendDocker=true 
                     -volcanoType=true                    # 重调度场景下必须使用Volcano
                     -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                     -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]，单位为秒
                     -hotReset=2 # 使用进程级恢复时，请将hotReset参数值设置为2，开启离线恢复模式
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                     -logLevel=0" ]
            securityContext:
              privileged: true
              readOnlyRootFilesystem: true
    ...
    ```

5.  （可选）使用推理卡故障恢复时，需要配置热复位功能。

    ```
          containers:
          - image: ascend-k8sdeviceplugin:v7.3.0
            name: device-plugin-01
            resources:
              requests:
                memory: 500Mi
                cpu: 500m
              limits:
                memory: 500Mi
                cpu: 500m
            command: [ "/bin/bash", "-c", "--"]
            args: [ "device-plugin  
    ...
                     -hotReset=0 # 使用推理卡故障恢复时，开启热复位功能
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                     -logLevel=0" ]
    ...
    ```

6.  （可选）如需更改kubelet的默认端口，则需要修改Ascend Device Plugin组件的启动YAML。示例如下。

    ```
      env:
         - name: NODE_NAME
           valueFrom:
             fieldRef:
               fieldPath: spec.nodeName
         - name: HOST_IP
           valueFrom:
             fieldRef:
               fieldPath: status.hostIP
         - name: KUBELET_PORT   # 通知Ascend Device Plugin组件当前节点kubelet默认端口号，若未自定义kubelet默认端口号则无需传入本字段
           value: "10251"      
    volumes:
       - name: device-plugin
         hostPath:
           path: /var/lib/kubelet/device-plugins
    ...
    ```

7.  在K8s管理节点上各YAML对应路径下执行以下命令，启动Ascend Device Plugin。

    -   K8s集群中存在使用Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件的节点（配合Volcano使用，支持虚拟化实例，YAML默认开启静态虚拟化）。

        ```
        kubectl apply -f device-plugin-volcano-v{version}.yaml
        ```

    -   K8s集群中存在使用Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件的节点（Ascend Device Plugin独立工作，不配合Volcano使用）。

        ```
        kubectl apply -f device-plugin-910-v{version}.yaml
        ```

    -   K8s集群中存在使用推理服务器（插Atlas 300I 推理卡）的节点（使用Volcano调度器）。

        ```
        kubectl apply -f device-plugin-310-volcano-v{version}.yaml
        ```

    -   K8s集群中存在使用推理服务器（插Atlas 300I 推理卡）的节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

        ```
        kubectl apply -f device-plugin-310-v{version}.yaml
        ```

    -   K8s集群中存在使用Atlas 推理系列产品的节点（使用Volcano调度器，支持虚拟化实例，YAML默认开启静态虚拟化）。

        ```
        kubectl apply -f device-plugin-310P-volcano-v{version}.yaml
        ```

    -   K8s集群中存在使用Atlas 推理系列产品的节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

        ```
        kubectl apply -f device-plugin-310P-v{version}.yaml
        ```

    -   K8s集群中存在使用Atlas 200I SoC A1 核心板的节点（使用Volcano调度器）。

        ```
        kubectl apply -f device-plugin-310P-1usoc-volcano-v{version}.yaml
        ```

    -   K8s集群中存在使用Atlas 200I SoC A1 核心板的节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

        ```
        kubectl apply -f device-plugin-310P-1usoc-v{version}.yaml
        ```

    >[!NOTE] 说明 
    >如果K8s集群使用了多种类型的昇腾AI处理器，请分别执行对应命令。

    启动示例如下：

    ```
    serviceaccount/ascend-device-plugin-sa created
    clusterrole.rbac.authorization.K8s.io/pods-node-ascend-device-plugin-role created
    clusterrolebinding.rbac.authorization.K8s.io/pods-node-ascend-device-plugin-rolebinding created
    daemonset.apps/ascend-device-plugin-daemonset created
    ```

8.  在任意节点执行以下命令，查看组件是否启动成功。

    ```
    kubectl get pod -n kube-system
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```
    NAME                                        READY   STATUS    RESTARTS   AGE
    ...
    ascend-device-plugin-daemonset-d5ctz  1/1   Running   0        11s
    ...
    ```

>[!NOTE] 说明 
>-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
>-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section479917441223"></a>**

**表 3** Ascend Device Plugin启动参数

<a name="table1064314568229"></a>
|参数|类型|默认值|说明|
|--|--|--|--|
|-fdFlag|bool|false|边缘场景标志，是否使用FusionDirector系统来管理设备。<ul><li>true：使用FusionDirector。</li><li>false：不使用FusionDirector。</li></ul>|
|-shareDevCount|uint|1|共享设备特性开关，取值范围为1~100。<p>默认值为1，代表不开启共享设备；取值为2~100，表示单颗芯片虚拟化出来的共享设备个数。</p><p>支持以下设备，其余设备该参数无效，不影响组件正常启动。</p><ul><li>Atlas 500 A2 智能小站</li><li>Atlas 200I A2 加速模块</li><li>Atlas 200I DK A2 开发者套件</li><li>Atlas 300I Pro 推理卡</li><li>Atlas 300V 视频解析卡</li><li>Atlas 300V Pro 视频解析卡</li></ul><p>若用户使用的是以上支持的Atlas 推理系列产品，需要注意以下问题：</p><ul><li>不支持在使用静态vNPU调度、动态vNPU调度、推理卡故障恢复和推理卡故障重调度等特性下使用共享设备功能。</li><li>单任务的请求资源数必须为1，不支持分配多芯片和跨芯片使用的场景。</li><li>依赖驱动开启共享模式，设置device-share为true，详细操作步骤和说明请参见《Atlas 中心推理卡 25.5.0 npu-smi 命令参考》中的“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100540373/af78d7e5">设置指定设备的指定芯片的容器共享模式</a>”章节。</li></ul>|
|-edgeLogFile|string|/var/alog/AtlasEdge_log/devicePlugin.log|边缘场景日志文件。fdFlag设置为true时生效。<p>单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。</p>|
|-useAscendDocker|bool|true|默认为true，容器引擎是否使用Ascend Docker Runtime。开启K8s的CPU绑核功能时，需要卸载Ascend Docker Runtime并重启容器引擎。取值说明如下：<ul><li>true：使用Ascend Docker Runtime。</li><li>false：不使用Ascend Docker Runtime。</li></ul><span> 说明： </span><p>MindCluster 5.0.RC1及以上版本只支持自动获取运行模式，不接受指定。</p>|
|-use310PMixedInsert|bool|false|是否使用混插模式。<ul><li>true：使用混插模式。</li><li>false：不使用混插模式。</li></ul><span> 说明： </span><ul><li>仅支持服务器混插Atlas 300I Pro 推理卡、Atlas 300V 视频解析卡、Atlas 300V Pro 视频解析卡。</li><li>服务器混插模式下不支持Volcano调度模式。</li><li>服务器混插模式不支持虚拟化实例。</li><li>服务器混插模式不支持故障重调度场景。</li><li>服务器混插模式不支持Ascend Docker Runtime。</li><li>非混插模式下，上报给K8s资源名称不变。<ul><li>非混插模式上报的资源名称格式为huawei.com/Ascend310P。</li><li>混插模式上报的资源名称格式为：huawei.com/Ascend310P-V、huawei.com/Ascend310P-VPro和huawei.com/Ascend310P-IPro。</li></ul></li></ul>|
|-volcanoType|bool|false|是否使用Volcano进行调度，当前已支持Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas 推理系列产品和推理服务器（插Atlas 300I 推理卡）芯片。<ul><li>true：使用Volcano。</li><li>false：不使用Volcano。</li></ul>|
|-presetVirtualDevice|bool|true|虚拟化功能开关。<ul><li>设置为true时，表示使用静态虚拟化。</li><li>设置为false时，表示使用动态虚拟化。需要同步开启Volcano，即设置-volcanoType参数为true。</li></ul>|
|-version|bool|false|是否查看当前Ascend Device Plugin的版本号。<ul><li>true：查询。</li><li>false：不查询。</li></ul>|
|-listWatchPeriod|int|5|<p>设置健康状态检查周期，取值范围为[3,1800]，单位为秒。</p><span> 说明： </span><p>每个周期内会进行如下检查，并将检查结果写入ConfigMap中。</p><ul><li>如果设备信息没有变化且距离上次更新ConfigMap未超过5min，则不会更新ConfigMap。</li><li>如果距离上次更新ConfigMap超过5min，则无论设备信息是否发生变化，都会更新ConfigMap。</li></ul>|
|-autoStowing|bool|true|是否自动纳管已修复设备，volcanoType为true时生效。<ul><li>true：自动纳管。</li><li>false：不会自动纳管。</li></ul><span> 说明： </span><p>设备故障后，会自动从K8s里面隔离。如果设备恢复正常，默认会自动加入K8s集群资源池。如果设备不稳定，可以设置为false，此时需要手动纳管。</p><ul><li>用户可以使用以下命令，将健康状态由unhealthy恢复为healthy的芯片重新放入资源池。<p>kubectl label nodes *node_name* huawei.com/Ascend910-Recover-</p></li><li>用户可以使用以下命令，将参数面网络健康状态由unhealthy恢复为healthy的芯片重新放入资源池。<p>kubectl label nodes *node_name* huawei.com/Ascend910-NetworkRecover-</p></li></ul>|
|-logLevel|int|0|日志级别：<ul><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>|
|-maxAge|int|7|日志备份时间限制，取值范围为7~700，单位为天。|
|-logFile|string|/var/log/mindx-dl/devicePlugin/devicePlugin.log|非边缘场景日志文件。fdFlag设置为false时生效。<p>单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：devicePlugin-触发转储的时间.log，如：devicePlugin-2023-10-07T03-38-24.402.log。</p>|
|-hotReset|int|-1|设备热复位功能参数。开启此功能，芯片发生故障后，Ascend Device Plugin会进行热复位操作，使芯片恢复健康。<ul><li>-1：关闭芯片复位功能</li><li>0：开启推理设备复位功能</li><li>1：开启训练设备在线复位功能</li><li>2：开启训练/推理设备离线复位功能</li></ul><span> 说明： </span><p>取值为1对应的功能已经日落，请配置其他取值。</p><p>该参数支持的训练设备：</p><ul><li>Atlas 800 训练服务器（型号 9000）（NPU满配）</li><li>Atlas 800 训练服务器（型号 9010）（NPU满配）</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD（型号 9000）</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A2 PoD 集群基础单元</li><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 800T A3 超节点服务器</li></ul><p>该参数支持的推理设备：</p><ul><li>Atlas 300I Pro 推理卡</li><li>Atlas 300V 视频解析卡</li><li>Atlas 300V Pro 视频解析卡</li><li>Atlas 300I Duo 推理卡</li><li>Atlas 300I 推理卡（型号 3000）（整卡）</li><li>Atlas 300I 推理卡（型号 3010）</li><li>Atlas 800I A2 推理服务器</li><li>A200I A2 Box 异构组件</li><li>Atlas 800I A3 超节点服务器</li></ul><span> 说明： </span><ul><li>针对Atlas 300I Duo 推理卡形态硬件，仅支持按卡复位，即两颗芯片会同时复位。</li><li>Atlas 800I A2 推理服务器存在以下两种热复位方式，一台Atlas 800I A2 推理服务器只能使用一种热复位方式，由集群调度组件自动识别使用哪种热复位方式。<ul><li>方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。</li><li>方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。</li></ul></li></ul>|
|-linkdownTimeout|int|30|网络linkdown超时时间，单位秒，取值范围为1~30。<p>该参数取值建议与用户在训练脚本中配置的HCCL_RDMA_TIMEOUT时间一致。如果是多任务，建议设置为多任务中HCCL_RDMA_TIMEOUT的最小值。</p>|
|-enableSlowNode|bool|false|是否启用慢节点检测（劣化诊断）功能。<ul><li>true：开启。</li><li>false：关闭。</li></ul><span> 说明： </span><p>关于劣化诊断的详细说明请参见《iMaster CCAE 产品文档》的“<a href="https://support.huawei.com/hedex/hdx.do?docid=EDOC1100445519&amp;id=ZH-CN_TOPIC_0000002147436540">劣化诊断</a>”章节。</p>|
|-dealWatchHandler|bool|false|当informer链接因异常结束时，是否需要刷新本地的Pod informer缓存。<ul><li>true：刷新Pod informer缓存。</li><li>false：不刷新Pod informer缓存。</li></ul>|
|-checkCachedPods|bool|true|是否定期检查缓存中的Pod。默认取值为true，当缓存中的Pod超过1小时没有被更新，Ascend Device Plugin将会主动请求api-server查看Pod情况。<ul><li>true：检查。</li><li>false：不检查。</li></ul>|
|-maxBackups|int|30|转储后日志文件保留个数上限，取值范围为1~30，单位为个。|
|-thirdPartyScanDelay|int|300|<p>Ascend Device Plugin组件启动重新扫描的等待时长。</p><p>Ascend Device Plugin自动复位芯片失败后，会将失败信息写到节点annotation上，三方平台可以根据该信息复位失败的芯片。Ascend Device Plugin组件根据本参数设置的等待时长，等待一段时间后，重新扫描设备。</p><p>仅Atlas 800T A3 超节点服务器支持使用本参数。</p><p>单位：秒。</p>|
|-deviceResetTimeout|int|60|组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600。<ul><li>Atlas A2 训练系列产品、Atlas 800I A2 推理服务器、A200I A2 Box 异构组件：建议配置为150秒。</li><li>Atlas A3 训练系列产品、A200T A3 Box8 超节点服务器、Atlas 800I A3 超节点服务器：建议配置为360秒。</li><li>推理服务器（插Atlas 350 标卡）、Atlas 850 服务器、Atlas 950 SuperPoD 超节点/集群：建议配置为600秒。</li></ul>|
|-h或者-help|无|无|显示帮助信息。|

### Volcano<a name="ZH-CN_TOPIC_0000002479226394"></a>




#### 安装Volcano<a name="ZH-CN_TOPIC_0000002511426351"></a>

-   使用整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须在管理节点安装**调度器**，该调度器可以是Volcano或其他调度器。
-   若使用Volcano进行任务调度，则不建议通过Docker或Containerd指令创建/挂载NPU卡的容器，并在容器内跑任务。否则可能会触发Volcano调度问题。
-   仅使用容器化支持和资源监测的用户，可以不安装Volcano，请直接跳过本章节。

    本章为集群调度提供Volcano组件（vc-scheduler和vc-controller-manager）的安装指导。如需使用开源Volcano的其他组件，请用户自行安装，并保证其安全性。

    >[!NOTE] 说明 
    >-   本文档中Volcano默认为集群调度组件提供的Volcano组件。若需要更高版本的Volcano或者其他基于开源Volcano的调度器，可通过[（可选）集成昇腾插件扩展开源Volcano](#可选集成昇腾插件扩展开源volcano)章节，集成集群调度组件为开发者提供的Ascend-volcano-plugin插件，实现NPU调度相关功能。
    >-   6.0.RC1及以上版本NodeD与老版本Volcano不兼容，若使用6.0.RC1及以上版本的NodeD，需要配套使用6.0.RC1及以上版本的Volcano。
    >-   6.0.RC2及以上版本使用Volcano调度器时，默认必须安装ClusterD组件，若不安装ClusterD，则必须修改Volcano的启动参数，否则Volcano将无法正常调度任务。

**操作步骤<a name="section57241227172819"></a>**

1.  以root用户登录K8s管理节点，并执行以下命令，查看Volcano镜像和版本号是否正确。

    ```
    docker images | grep volcanosh
    ```

    回显示例如下。

    ```
    volcanosh/vc-controller-manager      v1.7.0              84c73128cc55        3 days ago          44.5MB
    volcanosh/vc-scheduler               v1.7.0              e90c114c75b1        3 days ago          188MB
    ```

    -   是，执行[步骤2](#li823273914318)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作。

2.  <a name="li823273914318"></a>将Volcano软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3.  如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改对应启动YAML文件中Volcano的启动参数。常用启动参数请参见[表4](#table5305150122116)和[表5](#table203077022111)。
4.  配置Volcano日志转储。

    安装过程中，Volcano日志将挂载到磁盘空间（“/var/log/mindx-dl”）。默认情况下单日日志写入达到1.8G后，Volcano将清空日志文件。为防止空间被占满，请为Volcano配置日志转储，配置项信息参见[表1](#table1123141112311)，或选择更频繁的日志转储策略，避免日志丢失。

    1.  在管理节点“/etc/logrotate.d”目录下，执行以下命令，创建日志转储配置文件。

        ```
        vi /etc/logrotate.d/文件名
        ```

        例如：

        ```
        vi /etc/logrotate.d/volcano
        ```

        写入以下内容，然后执行<b>:wq</b>命令保存。

        ```
        /var/log/mindx-dl/volcano-*/*.log{    
             daily     
             rotate 8     
             size 50M     
             compress     
             dateext     
             missingok     
             notifempty     
             copytruncate     
             create 0640 hwMindX hwMindX     
             sharedscripts     
             postrotate         
                 chmod 640 /var/log/mindx-dl/volcano-*/*.log                
                 chmod 440 /var/log/mindx-dl/volcano-*/*.log-*            
             endscript 
        }
        ```

    2.  依次执行以下命令，设置配置文件权限为640，属主为root。

        ```
        chmod 640 /etc/logrotate.d/文件名
        chown root /etc/logrotate.d/文件名
        ```

        例如：

        ```
        chmod 640 /etc/logrotate.d/volcano
        chown root /etc/logrotate.d/volcano
        ```

    **表 1** Volcano日志转储文件配置项

    <a name="table1123141112311"></a>
    <table><thead align="left"><tr id="row412371119316"><th class="cellrowborder" valign="top" width="20.352035203520348%" id="mcps1.2.4.1.1"><p id="p12123811163113"><a name="p12123811163113"></a><a name="p12123811163113"></a>配置项</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.82338233823382%" id="mcps1.2.4.1.2"><p id="p8123141118315"><a name="p8123141118315"></a><a name="p8123141118315"></a>说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="45.82458245824582%" id="mcps1.2.4.1.3"><p id="p16123121111319"><a name="p16123121111319"></a><a name="p16123121111319"></a>可选值</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row612391119318"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p1012481112315"><a name="p1012481112315"></a><a name="p1012481112315"></a>daily</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p16124101111317"><a name="p16124101111317"></a><a name="p16124101111317"></a>日志转储频率。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul1512431163118"></a><a name="ul1512431163118"></a><ul id="ul1512431163118"><li>daily：每日进行一次转储检查。</li><li>weekly：每周进行一次转储检查。</li><li>monthly：每月进行一次转储检查。</li><li>yearly：每年进行一次转储检查。</li></ul>
    </td>
    </tr>
    <tr id="row912511118314"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p111261711103117"><a name="p111261711103117"></a><a name="p111261711103117"></a>rotate <em id="i20126171193115"><a name="i20126171193115"></a><a name="i20126171193115"></a>x</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p13126191153117"><a name="p13126191153117"></a><a name="p13126191153117"></a>日志文件删除之前转储的次数。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><p id="p161261911153111"><a name="p161261911153111"></a><a name="p161261911153111"></a><em id="i161266119317"><a name="i161266119317"></a><a name="i161266119317"></a>x</em>为备份次数。</p>
    <p id="p19126101103110"><a name="p19126101103110"></a><a name="p19126101103110"></a>例如：</p>
    <a name="ul151261211103117"></a><a name="ul151261211103117"></a><ul id="ul151261211103117"><li>rotate 0：没有备份。</li><li>rotate 8：保留8次备份。</li></ul>
    </td>
    </tr>
    <tr id="row1912641115318"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p15127111153115"><a name="p15127111153115"></a><a name="p15127111153115"></a>size <em id="i912731113110"><a name="i912731113110"></a><a name="i912731113110"></a>xx</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p412751115312"><a name="p412751115312"></a><a name="p412751115312"></a>日志文件到达指定的大小时才转储。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><p id="p131273112314"><a name="p131273112314"></a><a name="p131273112314"></a>size单位可以指定：</p>
    <a name="ul1012771118311"></a><a name="ul1012771118311"></a><ul id="ul1012771118311"><li>byte（缺省）</li><li>K</li><li>M</li></ul>
    <p id="p1112761173115"><a name="p1112761173115"></a><a name="p1112761173115"></a>例如size 50M指日志文件达到50 MB时转储。</p>
    <div class="note" id="note191277111311"><a name="note191277111311"></a><a name="note191277111311"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p17127191193120"><a name="p17127191193120"></a><a name="p17127191193120"></a>logrotate会根据配置的转储频率，定期检查日志文件大小，检查时大小超过size的文件才会触发转储。</p>
    <p id="p112771153111"><a name="p112771153111"></a><a name="p112771153111"></a>这意味着，logrotate并不会在日志文件达到大小限制时立刻将其转储。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row4127111173111"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p161282115316"><a name="p161282115316"></a><a name="p161282115316"></a>compress</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p112851110317"><a name="p112851110317"></a><a name="p112851110317"></a>是否通过gzip压缩转储以后的日志。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul712814118312"></a><a name="ul712814118312"></a><ul id="ul712814118312"><li>compress：使用gzip压缩。</li><li>nocompress：不使用gzip压缩。</li></ul>
    </td>
    </tr>
    <tr id="row18128511203117"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p412801153117"><a name="p412801153117"></a><a name="p412801153117"></a>notifempty</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p9128141123113"><a name="p9128141123113"></a><a name="p9128141123113"></a>空文件是否转储。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul31281611163118"></a><a name="ul31281611163118"></a><ul id="ul31281611163118"><li>ifempty：空文件也转储。</li><li>notifempty：空文件不触发转储。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

5.  （可选）在“volcano-v<i>\{version\}</i>.yaml“中，配置Volcano所需的CPU和内存。CPU和内存推荐值可以参见[开源Volcano官方文档](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html)中volcano-controller和volcano-scheduler表格的建议值。

    ```
    ...
    kind: Deployment
    ...
      labels:
        app: volcano-scheduler
    spec:
      replicas: 1
    ...
        spec:
    ...
              imagePullPolicy: "IfNotPresent"
              resources:
                requests:
                  memory: 4Gi
                  cpu: 5500m
                limits:
                  memory: 8Gi
                  cpu: 5500m
    ...
    kind: Deployment
    ...
      labels:
        app: volcano-controller
    spec:
    ...
        spec:
    ...
              resources:
                requests:
                  memory: 3Gi
                  cpu: 2000m
                limits:
                  memory: 3Gi
                  cpu: 2000m
    ...
    ```

6.  （可选）调度时间性能调优。支持在“volcano-v<i>\{version\}</i>.yaml“中，配置Volcano所使用的插件。请参见[开源Volcano官方文档](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html)中Volcano高级配置参数说明和支持的Plugins列表的表格说明进行操作。

    ```
    ...
    data:
      volcano-scheduler.conf: |
        actions: "enqueue, allocate, backfill"
        tiers:
        - plugins:
          - name: priority
            enableNodeOrder: false
          - name: gang
            enableNodeOrder: false
          - name: conformance
            enableNodeOrder: false
          - name: volcano-npu_v7.3.0_linux-aarch64   # 其中v7.3.0为MindCluster的版本号，根据不同版本，该处取值不同
        - plugins:
          - name: drf
            enableNodeOrder: false
          - name: predicates
            enableNodeOrder: false
            arguments:
              predicate.GPUSharingEnable: false
              predicate.GPUNumberEnable: false
          - name: proportion
            enableNodeOrder: false
          - name: nodeorder
          - name: binpack
            enableNodeOrder: false
    ....
    ```

7.  （可选）在“volcano-_v\{version\}_.yaml”中，配置开启Volcano健康检查接口和Prometheus信息收集接口。

    ```
    ...
    kind: Deployment
    metadata:
      name: volcano-scheduler
      namespace: volcano-system
      labels:
        app: volcano-scheduler
    spec:
      ...
      template:
    ...
            - name: volcano-scheduler
              image: volcanosh/vc-scheduler:v1.7.0
              args: [ ...
                  ...
                  --enable-healthz=true   # 为保证可正常访问Volcano健康检查端口，本参数取值需为"true"
                  --enable-metrics=true   # 为保证可正常访问Prometheus信息收集端口，本参数取值需为"true"
                  ...
    ...
    ```

    **表 2** 集群调度Volcano组件开放接口说明

    <a name="zh-cn_topic_0000001446965056_table173071368477"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001446965056_row153077618473"><th class="cellrowborder" valign="top" width="34.68346834683469%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001446965056_p3307116134715"><a name="zh-cn_topic_0000001446965056_p3307116134715"></a><a name="zh-cn_topic_0000001446965056_p3307116134715"></a>访问方式</p>
    </th>
    <th class="cellrowborder" valign="top" width="6.0906090609060906%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001446965056_p1525244211493"><a name="zh-cn_topic_0000001446965056_p1525244211493"></a><a name="zh-cn_topic_0000001446965056_p1525244211493"></a>协议</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.741174117411742%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001446965056_p04543391867"><a name="zh-cn_topic_0000001446965056_p04543391867"></a><a name="zh-cn_topic_0000001446965056_p04543391867"></a>方法</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.89168916891689%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001446965056_p23071468473"><a name="zh-cn_topic_0000001446965056_p23071468473"></a><a name="zh-cn_topic_0000001446965056_p23071468473"></a>作用</p>
    </th>
    <th class="cellrowborder" valign="top" width="30.59305930593059%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001446965056_p730796134713"><a name="zh-cn_topic_0000001446965056_p730796134713"></a><a name="zh-cn_topic_0000001446965056_p730796134713"></a>所属组件</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001446965056_row23070613479"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p1730717615477"><a name="zh-cn_topic_0000001446965056_p1730717615477"></a><a name="zh-cn_topic_0000001446965056_p1730717615477"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252142154917"><a name="zh-cn_topic_0000001446965056_p10252142154917"></a><a name="zh-cn_topic_0000001446965056_p10252142154917"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p64546391612"><a name="zh-cn_topic_0000001446965056_p64546391612"></a><a name="zh-cn_topic_0000001446965056_p64546391612"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p151727316530"><a name="zh-cn_topic_0000001446965056_p151727316530"></a><a name="zh-cn_topic_0000001446965056_p151727316530"></a>健康检查端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p103089610475"><a name="zh-cn_topic_0000001446965056_p103089610475"></a><a name="zh-cn_topic_0000001446965056_p103089610475"></a>volcano-controller</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001446965056_row1308176144715"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p2308865475"><a name="zh-cn_topic_0000001446965056_p2308865475"></a><a name="zh-cn_topic_0000001446965056_p2308865475"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p162523428499"><a name="zh-cn_topic_0000001446965056_p162523428499"></a><a name="zh-cn_topic_0000001446965056_p162523428499"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p245433916610"><a name="zh-cn_topic_0000001446965056_p245433916610"></a><a name="zh-cn_topic_0000001446965056_p245433916610"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p53084617475"><a name="zh-cn_topic_0000001446965056_p53084617475"></a><a name="zh-cn_topic_0000001446965056_p53084617475"></a>健康检查端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p14308176104718"><a name="zh-cn_topic_0000001446965056_p14308176104718"></a><a name="zh-cn_topic_0000001446965056_p14308176104718"></a>volcano-scheduler</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001446965056_row830812614472"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p19308116104714"><a name="zh-cn_topic_0000001446965056_p19308116104714"></a><a name="zh-cn_topic_0000001446965056_p19308116104714"></a>http://volcano-scheduler-serviceIP:8080/metrics</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252104224912"><a name="zh-cn_topic_0000001446965056_p10252104224912"></a><a name="zh-cn_topic_0000001446965056_p10252104224912"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p134543391765"><a name="zh-cn_topic_0000001446965056_p134543391765"></a><a name="zh-cn_topic_0000001446965056_p134543391765"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p193087624718"><a name="zh-cn_topic_0000001446965056_p193087624718"></a><a name="zh-cn_topic_0000001446965056_p193087624718"></a>Prometheus信息收集端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p3308166154716"><a name="zh-cn_topic_0000001446965056_p3308166154716"></a><a name="zh-cn_topic_0000001446965056_p3308166154716"></a>volcano-scheduler</p>
    </td>
    </tr>
    </tbody>
    </table>

8.  （可选）在“volcano-v<i>\{version\}</i>.yaml“中，配置Volcano使用的集群调度组件为用户提供的重调度时删除Pod的模式、虚拟化方式、交换机亲和性调度、是否自维护可用芯片状态等。

    ```
    ...
    data:
      volcano-scheduler.conf: |
    ...
        configurations:
          - name: init-params
            arguments: {"grace-over-time":"900","presetVirtualDevice":"true","nslb-version":"1.0","shared-tor-num":"2","useClusterInfoManager":"false","self-maintain-available-card":"true","super-pod-size": "48","reserve-nodes": "2","forceEnqueue":"true"}
    ...
    ```

    **表 3**  参数说明

    <a name="table208981646194315"></a>
    <table><thead align="left"><tr id="row08991746174316"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.1"><p id="p132621494445"><a name="p132621494445"></a><a name="p132621494445"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.2"><p id="p194862061467"><a name="p194862061467"></a><a name="p194862061467"></a>默认值</p>
    </th>
    <th class="cellrowborder" valign="top" width="60%" id="mcps1.2.4.1.3"><p id="p18991846144317"><a name="p18991846144317"></a><a name="p18991846144317"></a>参数说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1788817373541"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p1888143725417"><a name="p1888143725417"></a><a name="p1888143725417"></a>grace-over-time</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p7888103725412"><a name="p7888103725412"></a><a name="p7888103725412"></a>900</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p145262285146"><a name="p145262285146"></a><a name="p145262285146"></a>重调度优雅删除模式下删除Pod所需最大时间，单位为秒，取值范围2~3600。配置该字段表示使用重调度的优雅删除模式。优雅删除是指在重调度过程中，会等待<span id="ph8305245165813"><a name="ph8305245165813"></a><a name="ph8305245165813"></a>Volcano</span>执行相关善后工作，900秒后若Pod还未删除成功，再直接强制删除Pod，不做善后。</p>
    </td>
    </tr>
    <tr id="row95211735125411"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p1352203555411"><a name="p1352203555411"></a><a name="p1352203555411"></a>presetVirtualDevice</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p552293515412"><a name="p552293515412"></a><a name="p552293515412"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p135221235105419"><a name="p135221235105419"></a><a name="p135221235105419"></a>采用的虚拟化方式。</p>
    <a name="ul206451443111219"></a><a name="ul206451443111219"></a><ul id="ul206451443111219"><li>true：静态虚拟化</li><li>false：动态虚拟化</li></ul>
    </td>
    </tr>
    <tr id="row1589974674320"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p6899114619435"><a name="p6899114619435"></a><a name="p6899114619435"></a>nslb-version</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p6484146114619"><a name="p6484146114619"></a><a name="p6484146114619"></a>1.0</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p7830165414514"><a name="p7830165414514"></a><a name="p7830165414514"></a>交换机亲和性调度的版本，可以取值为1.0和2.0。</p>
    <div class="note" id="note882315541054"><a name="note882315541054"></a><a name="note882315541054"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul59831535122714"></a><a name="ul59831535122714"></a><ul id="ul59831535122714"><li>交换机亲和性调度1.0版本支持<span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="ph168598363399"><a name="ph168598363399"></a><a name="ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph4181625925"><a name="ph4181625925"></a><a name="ph4181625925"></a>PyTorch</span>和<span id="ph61882510210"><a name="ph61882510210"></a><a name="ph61882510210"></a>MindSpore</span>。</li><li>交换机亲和性调度2.0版本支持<span id="ph311717506401"><a name="ph311717506401"></a><a name="ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph619244413568"><a name="ph619244413568"></a><a name="ph619244413568"></a>PyTorch</span>框架。</li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="row8899946174318"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p168998463434"><a name="p168998463434"></a><a name="p168998463434"></a>shared-tor-num</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p989910464439"><a name="p989910464439"></a><a name="p989910464439"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p925113215214"><a name="p925113215214"></a><a name="p925113215214"></a>交换机亲和性调度2.0中单个任务可使用的最大共享交换机数量，可取值为1或2。仅在nslb-version取值为2.0时生效。</p>
    <p id="p1856962434719"><a name="p1856962434719"></a><a name="p1856962434719"></a>交换机亲和性调度（1.0或2.0）说明可以参见<a href="./references.md#基于节点的亲和性">基于节点的亲和性</a>章节。</p>
    </td>
    </tr>
    <tr id="row797916276295"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p9621013114312"><a name="p9621013114312"></a><a name="p9621013114312"></a>useClusterInfoManager</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p1024875418187"><a name="p1024875418187"></a><a name="p1024875418187"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p1797510751015"><a name="p1797510751015"></a><a name="p1797510751015"></a><span id="ph18393155819297"><a name="ph18393155819297"></a><a name="ph18393155819297"></a>Volcano</span>获取集群信息的方式。取值说明如下：</p>
    <a name="ul675021361014"></a><a name="ul675021361014"></a><ul id="ul675021361014"><li>true：读取<span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span>上报的<span id="ph1921415457302"><a name="ph1921415457302"></a><a name="ph1921415457302"></a>ConfigMap</span>。</li><li>false：分别读取<span id="ph19274234236"><a name="ph19274234236"></a><a name="ph19274234236"></a>Ascend Device Plugin</span>和<span id="ph144095321390"><a name="ph144095321390"></a><a name="ph144095321390"></a>NodeD</span>上报的<span id="ph039324431114"><a name="ph039324431114"></a><a name="ph039324431114"></a>ConfigMap</span>。</li></ul>
    <div class="note" id="note1466414341216"><a name="note1466414341216"></a><a name="note1466414341216"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p1166463181214"><a name="p1166463181214"></a><a name="p1166463181214"></a>默认使用读取<span id="ph19101421151220"><a name="ph19101421151220"></a><a name="ph19101421151220"></a>ClusterD</span>组件上报的<span id="ph139579361121"><a name="ph139579361121"></a><a name="ph139579361121"></a>ConfigMap</span>。后续版本将不支持读取<span id="ph3588183951516"><a name="ph3588183951516"></a><a name="ph3588183951516"></a>Ascend Device Plugin</span>和<span id="ph1758893981514"><a name="ph1758893981514"></a><a name="ph1758893981514"></a>NodeD</span>上报的<span id="ph75887392157"><a name="ph75887392157"></a><a name="ph75887392157"></a>ConfigMap</span>。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row1913114164518"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p91454144514"><a name="p91454144514"></a><a name="p91454144514"></a>self-maintain-available-card</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p814241174517"><a name="p814241174517"></a><a name="p814241174517"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p151414164516"><a name="p151414164516"></a><a name="p151414164516"></a>Volcano是否自维护可用芯片状态。取值说明如下：</p>
    <a name="ul299044019472"></a><a name="ul299044019472"></a><ul id="ul299044019472"><li>true：Volcano自维护可用芯片状态。</li><li>false：Volcano根据ClusterD或<span id="ph98552414486"><a name="ph98552414486"></a><a name="ph98552414486"></a>Ascend Device Plugin</span>上报的<span id="ph1185824104819"><a name="ph1185824104819"></a><a name="ph1185824104819"></a>ConfigMap</span>获取可用芯片状态。</li></ul>
    </td>
    </tr>
    <tr id="row4612538250"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p26130381510"><a name="p26130381510"></a><a name="p26130381510"></a>super-pod-size</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p6613173818516"><a name="p6613173818516"></a><a name="p6613173818516"></a>48</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p461323812519"><a name="p461323812519"></a><a name="p461323812519"></a><span id="ph128111331314"><a name="ph128111331314"></a><a name="ph128111331314"></a>Atlas 900 A3 SuperPoD 超节点</span>中一个超节点的节点数量。</p>
    </td>
    </tr>
    <tr id="row9561856657"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p956215565514"><a name="p956215565514"></a><a name="p956215565514"></a>reserve-nodes</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p125626567516"><a name="p125626567516"></a><a name="p125626567516"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p25637568515"><a name="p25637568515"></a><a name="p25637568515"></a><span id="ph915032251212"><a name="ph915032251212"></a><a name="ph915032251212"></a>Atlas 900 A3 SuperPoD 超节点</span>中一个超节点中预留节点数量。</p>
    <div class="note" id="note1514175285210"><a name="note1514175285210"></a><a name="note1514175285210"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p96481321115510"><a name="p96481321115510"></a><a name="p96481321115510"></a>若设置的reserve-nodes大于super-pod-size时，存在以下场景。</p>
    <a name="ul13842528165510"></a><a name="ul13842528165510"></a><ul id="ul13842528165510"><li>super-pod-size大于2，则默认重置reserve-nodes取值为2</li><li>super-pod-size小于或等于2，则默认重置reserve-nodes取值为0。</li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="row1890722719501"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p590882716507"><a name="p590882716507"></a><a name="p590882716507"></a><span id="ph19180940145012"><a name="ph19180940145012"></a><a name="ph19180940145012"></a>forceEnqueue</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p17908162765017"><a name="p17908162765017"></a><a name="p17908162765017"></a><span id="ph16315161885115"><a name="ph16315161885115"></a><a name="ph16315161885115"></a>true</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p1790852711505"><a name="p1790852711505"></a><a name="p1790852711505"></a><span id="ph188121729115116"><a name="ph188121729115116"></a><a name="ph188121729115116"></a>任务在集群NPU资源满足的情况下是否强制</span><span id="ph280124455118"><a name="ph280124455118"></a><a name="ph280124455118"></a>进入待调度队列</span><span id="ph1179814511514"><a name="ph1179814511514"></a><a name="ph1179814511514"></a>。</span><span id="ph2278145215114"><a name="ph2278145215114"></a><a name="ph2278145215114"></a>取值说明如下：</span></p>
    <a name="ul12820554135117"></a><a name="ul12820554135117"></a><ul id="ul12820554135117"><li>true：Volcano开启<span id="ph11766123385220"><a name="ph11766123385220"></a><a name="ph11766123385220"></a>Enqueue</span>这个action时，若集群NPU资源满足当前任务，则任务会<span id="ph3349644125319"><a name="ph3349644125319"></a><a name="ph3349644125319"></a>强制</span><span id="ph22191237135316"><a name="ph22191237135316"></a><a name="ph22191237135316"></a>进入待调度队列</span>，不会关心其他资源是否充足。如果当前任务长时间在待调度队列中，会预占用资源，从而可能导致其他任务无法入队。</li><li>其他值：当集群NPU资源不足时，拒绝任务<span id="ph6205121155415"><a name="ph6205121155415"></a><a name="ph6205121155415"></a>进入待调度队列。若</span>NPU资源满足当前任务，则由所有插件共同决定是否<span id="ph370210413554"><a name="ph370210413554"></a><a name="ph370210413554"></a>进入待调度队列</span>。</li></ul>
    <p id="p12691948115614"><a name="p12691948115614"></a><a name="p12691948115614"></a>关于该参数的详细说明请参见<a href="https://volcano.sh/en/docs/v1-12-0/actions/" target="_blank" rel="noopener noreferrer">Volcano Actions</a>。</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 说明 
    >-   更多关于开源Volcano的配置，可以参见[开源Volcano官方文档](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html)进行操作。
    >-   K8s支持使用nodeAffinity字段进行节点亲和性调度，该字段的详细说明请参见[Kubernetes官方文档](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/)；Volcano也支持使用该字段，操作指导请参见[调度配置](./common_operations.md#调度配置)章节。

9.  （可选）调度时间性能调优。支持Volcano将单任务（训练的vcjob或acjob任务）的4000或5000个Pod调度到4000或5000个节点上的调度时间优化到5分钟左右，若用户想要使用该调度性能，需要在“volcano-v<i>\{version\}</i>.yaml”上做如下修改。
 
    -   若要达到5分钟左右的参考时间，需要保证CPU的频率至少为2.60GHz，APIServer时延不超过80毫秒。
    -   如果不使用K8s原生的nodeAffinity和podAntiAffinity进行调度，可以关闭nodeorder插件，进一步减少调度时间。

    ```
    data:
      volcano-scheduler.conf: |
    
    ...
          - name: proportion
            enableNodeOrder: false
          - name: nodeorder
            enableNodeOrder: false     # 可选，不使用nodeAffinity和podAntiAffinity调度时，可关闭nodeorder插件
    ...
          containers:
            - name: volcano-scheduler
              image: volcanosh/vc-scheduler:v1.7.0
              command: ["/bin/ash"]
              args: ["-c", "umask 027; GOMEMLIMIT=15000000000 GOGC=off /vc-scheduler      # 新增GOMEMLIMIT=15000000000和GOGC=off字段
                      --scheduler-conf=/volcano.scheduler/volcano-scheduler.conf
                      --plugins-dir=plugins
                      --logtostderr=false
                      --log_dir=/var/log/mindx-dl/volcano-scheduler
                      --log_file=/var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log
                      -v=2 2>&1"]
              imagePullPolicy: "IfNotPresent"
              resources:
                requests:
                  memory: 10000Mi                                                                # 将4Gi修改为10000Mi
                  cpu: 5500m
                limits:
                  memory: 15000Mi                                                       # 将8Gi修改为15000Mi
                  cpu: 5500m
    ...
    ```

10. 在管理节点的YAML所在路径，执行以下命令，启动Volcano。

    ```
    kubectl apply -f volcano-v{version}.yaml
    ```

    启动示例如下：

    ```
    namespace/volcano-system created
    namespace/volcano-monitoring created
    configmap/volcano-scheduler-configmap created
    serviceaccount/volcano-scheduler created
    clusterrole.rbac.authorization.K8s.io/volcano-scheduler created
    clusterrolebinding.rbac.authorization.K8s.io/volcano-scheduler-role created
    deployment.apps/volcano-scheduler created
    service/volcano-scheduler-service created
    serviceaccount/volcano-controllers created
    clusterrole.rbac.authorization.K8s.io/volcano-controllers created
    clusterrolebinding.rbac.authorization.K8s.io/volcano-controllers-role created
    deployment.apps/volcano-controllers created
    customresourcedefinition.apiextensions.K8s.io/jobs.batch.volcano.sh created
    customresourcedefinition.apiextensions.K8s.io/commands.bus.volcano.sh created
    customresourcedefinition.apiextensions.K8s.io/podgroups.scheduling.volcano.sh created
    customresourcedefinition.apiextensions.K8s.io/queues.scheduling.volcano.sh created
    customresourcedefinition.apiextensions.K8s.io/numatopologies.nodeinfo.volcano.sh created
    ```

11. 执行以下命令，查看组件状态。

    ```
    kubectl get pod -n volcano-system
    ```

    回显示例如下，出现**Running**表示组件启动成功：

    ```
    NAME                                          READY    STATUS     RESTARTS     AGE
    volcano-controllers-5cf8d788d5-qdpzq   1/1     Running   0          1m
    volcano-scheduler-6cffd555c9-45k7c     1/1     Running   0          1m
    ```

    >[!NOTE] 说明 
    >-   若Volcano的Pod状态为CrashLoopBackOff，可以参见[手动安装Volcano后，Pod状态为：CrashLoopBackOff](./faq.md#手动安装volcano后pod状态为crashloopbackoff)章节进行处理。
    >-   若volcano-scheduler-6cffd555c9-45k7c状态为Running，但是调度异常，可以参见[Volcano组件工作异常，日志出现Failed to get plugin](./faq.md#volcano组件工作异常日志出现failed-to-get-plugin)章节进行处理。
    >-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
    >-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
    >-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
    >-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section1317934882010"></a>**

**表 4**  volcano-scheduler启动参数

<a name="table5305150122116"></a>
<table><thead align="left"><tr id="row63052016218"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p133052042113"><a name="p133052042113"></a><a name="p133052042113"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15.310000000000002%" id="mcps1.2.5.1.2"><p id="p330560162111"><a name="p330560162111"></a><a name="p330560162111"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="29.69%" id="mcps1.2.5.1.3"><p id="p3305600215"><a name="p3305600215"></a><a name="p3305600215"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p63067062115"><a name="p63067062115"></a><a name="p63067062115"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row12306160112118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p14306100102116"><a name="p14306100102116"></a><a name="p14306100102116"></a>--log-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p103066014211"><a name="p103066014211"></a><a name="p103066014211"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p13306150192120"><a name="p13306150192120"></a><a name="p13306150192120"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p2030616002117"><a name="p2030616002117"></a><a name="p2030616002117"></a>日志目录，组件启动YAML中默认值为/var/log/mindx-dl/volcano-scheduler。</p>
</td>
</tr>
<tr id="row230620102115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p143064002118"><a name="p143064002118"></a><a name="p143064002118"></a>--log-file</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p173067012119"><a name="p173067012119"></a><a name="p173067012119"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p430620132116"><a name="p430620132116"></a><a name="p430620132116"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p930615020218"><a name="p930615020218"></a><a name="p930615020218"></a>日志文件名称，组件启动YAML中默认值为/var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log。</p>
<div class="note" id="note19596191219291"><a name="note19596191219291"></a><a name="note19596191219291"></a><div class="notebody"><p id="p10596012112919"><a name="p10596012112919"></a><a name="p10596012112919"></a>转储后文件的命名格式为：volcano-scheduler.log-触发转储的时间.gz，如：volcano-scheduler.log-20230926.gz。</p>
</div></div>
</td>
</tr>
<tr id="row17922126205817"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p139228267582"><a name="p139228267582"></a><a name="p139228267582"></a>--scheduler-conf</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p17490165855810"><a name="p17490165855810"></a><a name="p17490165855810"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p192312261580"><a name="p192312261580"></a><a name="p192312261580"></a>/volcano.scheduler/volcano-scheduler.conf</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1192372695818"><a name="p1192372695818"></a><a name="p1192372695818"></a>调度组件配置文件的绝对路径。</p>
</td>
</tr>
<tr id="row630618042113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p173061701214"><a name="p173061701214"></a><a name="p173061701214"></a>--logtostderr</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p730613011217"><a name="p730613011217"></a><a name="p730613011217"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p0306170112117"><a name="p0306170112117"></a><a name="p0306170112117"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p430613012117"><a name="p430613012117"></a><a name="p430613012117"></a>日志是否打印到标准输出。</p>
<a name="ul582374031615"></a><a name="ul582374031615"></a><ul id="ul582374031615"><li>true：打印。</li><li>false：不打印。</li></ul>
</td>
</tr>
<tr id="row53063062118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p330618022113"><a name="p330618022113"></a><a name="p330618022113"></a>-v</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p133061010218"><a name="p133061010218"></a><a name="p133061010218"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p23068042118"><a name="p23068042118"></a><a name="p23068042118"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p43067042115"><a name="p43067042115"></a><a name="p43067042115"></a>日志输出级别：</p>
<a name="ul03064012212"></a><a name="ul03064012212"></a><ul id="ul03064012212"><li>取值为1：error</li><li>取值为2：warning</li><li>取值为3：info</li><li>取值为4：debug</li></ul>
</td>
</tr>
<tr id="row11306140152113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p730614015211"><a name="p730614015211"></a><a name="p730614015211"></a>--plugins-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1130614013214"><a name="p1130614013214"></a><a name="p1130614013214"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p12307200192115"><a name="p12307200192115"></a><a name="p12307200192115"></a>plugins</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p173071603217"><a name="p173071603217"></a><a name="p173071603217"></a>scheduler插件加载路径。</p>
</td>
</tr>
<tr id="row113072012113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p9307140142120"><a name="p9307140142120"></a><a name="p9307140142120"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p03072016212"><a name="p03072016212"></a><a name="p03072016212"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p330712011215"><a name="p330712011215"></a><a name="p330712011215"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p53071209215"><a name="p53071209215"></a><a name="p53071209215"></a>是否查询volcano-scheduler二进制版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row62114943417"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p182349173416"><a name="p182349173416"></a><a name="p182349173416"></a>--log_file_max_size</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1221849193415"><a name="p1221849193415"></a><a name="p1221849193415"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p1021949203420"><a name="p1021949203420"></a><a name="p1021949203420"></a>1800</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1321749193419"><a name="p1321749193419"></a><a name="p1321749193419"></a>日志文件最大存储大小（单位为M）。</p>
<div class="note" id="note1919311416364"><a name="note1919311416364"></a><a name="note1919311416364"></a><div class="notebody"><p id="p7193444361"><a name="p7193444361"></a><a name="p7193444361"></a>当日志文件大小超过阈值时，日志内容会被清空。</p>
</div></div>
</td>
</tr>
<tr id="row159867311462"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10986173174613"><a name="p10986173174613"></a><a name="p10986173174613"></a>--leader-elect</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1098619374617"><a name="p1098619374617"></a><a name="p1098619374617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p19866311462"><a name="p19866311462"></a><a name="p19866311462"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p4986143184611"><a name="p4986143184611"></a><a name="p4986143184611"></a>多副本启动时启动选主模式。</p>
</td>
</tr>
<tr id="row1253065634617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1453015644610"><a name="p1453015644610"></a><a name="p1453015644610"></a>--percentage-nodes-to-find</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p145301156194612"><a name="p145301156194612"></a><a name="p145301156194612"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p16530175615462"><a name="p16530175615462"></a><a name="p16530175615462"></a>100</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p11530165644617"><a name="p11530165644617"></a><a name="p11530165644617"></a>任务调度时选取可用节点占集群总节点的百分比。</p>
</td>
</tr>
</tbody>
</table>

**表 5**  volcano-controller启动参数

<a name="table203077022111"></a>
<table><thead align="left"><tr id="row18307705217"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p193071001218"><a name="p193071001218"></a><a name="p193071001218"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p13307208218"><a name="p13307208218"></a><a name="p13307208218"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p123078062120"><a name="p123078062120"></a><a name="p123078062120"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p4307100172120"><a name="p4307100172120"></a><a name="p4307100172120"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row173077014210"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p43078015211"><a name="p43078015211"></a><a name="p43078015211"></a>--log-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p173071104213"><a name="p173071104213"></a><a name="p173071104213"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p113071302218"><a name="p113071302218"></a><a name="p113071302218"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p330718019213"><a name="p330718019213"></a><a name="p330718019213"></a>日志目录，组件启动YAML中默认值为/var/log/mindx-dl/volcano-controller。</p>
</td>
</tr>
<tr id="row1307170112113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p17307130112117"><a name="p17307130112117"></a><a name="p17307130112117"></a>--log-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1630780182118"><a name="p1630780182118"></a><a name="p1630780182118"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1930714062115"><a name="p1930714062115"></a><a name="p1930714062115"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p143077018217"><a name="p143077018217"></a><a name="p143077018217"></a>日志文件名称，组件启动YAML中默认值为/var/log/mindx-dl/volcano-controller/volcano-controller.log。</p>
<div class="note" id="note215144410296"><a name="note215144410296"></a><a name="note215144410296"></a><div class="notebody"><p id="p715144132910"><a name="p715144132910"></a><a name="p715144132910"></a>转储后文件的命名格式为：volcano-controller.log-触发转储的时间.gz，如：volcano-controller.log-20230926.gz。</p>
</div></div>
</td>
</tr>
<tr id="row730760202120"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p93071805219"><a name="p93071805219"></a><a name="p93071805219"></a>--logtostderr</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p730812011211"><a name="p730812011211"></a><a name="p730812011211"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p2308140142118"><a name="p2308140142118"></a><a name="p2308140142118"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p2308170172116"><a name="p2308170172116"></a><a name="p2308170172116"></a>日志是否打印到标准输出。</p>
<a name="ul142362048125710"></a><a name="ul142362048125710"></a><ul id="ul142362048125710"><li>true：打印。</li><li>false：不打印。</li></ul>
</td>
</tr>
<tr id="row930819012214"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p193088092115"><a name="p193088092115"></a><a name="p193088092115"></a>-v</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1930812016213"><a name="p1930812016213"></a><a name="p1930812016213"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p123081003218"><a name="p123081003218"></a><a name="p123081003218"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p330830162118"><a name="p330830162118"></a><a name="p330830162118"></a>日志输出级别：</p>
<a name="ul6308150112119"></a><a name="ul6308150112119"></a><ul id="ul6308150112119"><li>1：error</li><li>2：warning</li><li>3：info</li><li>4：debug</li></ul>
</td>
</tr>
<tr id="row1330813015217"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p133085052120"><a name="p133085052120"></a><a name="p133085052120"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p030814011212"><a name="p030814011212"></a><a name="p030814011212"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p10308140122115"><a name="p10308140122115"></a><a name="p10308140122115"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p130818011219"><a name="p130818011219"></a><a name="p130818011219"></a>volcano-controller二进制版本号。</p>
</td>
</tr>
<tr id="row926534763719"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1413064912376"><a name="p1413064912376"></a><a name="p1413064912376"></a>--log_file_max_size</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p313074910373"><a name="p313074910373"></a><a name="p313074910373"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p31301349183714"><a name="p31301349183714"></a><a name="p31301349183714"></a>1800</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p111301349113715"><a name="p111301349113715"></a><a name="p111301349113715"></a>日志文件最大存储大小（单位为M）。</p>
<div class="note" id="note1513064943719"><a name="note1513064943719"></a><a name="note1513064943719"></a><div class="notebody"><p id="p111317492373"><a name="p111317492373"></a><a name="p111317492373"></a>当日志文件大小超过阈值时，日志内容会被清空。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>Volcano为开源软件，启动参数只罗列目前使用的常见参数，其他详细的参数请参见开源软件的说明。

#### （可选）使用Volcano交换机亲和性调度<a name="ZH-CN_TOPIC_0000002479226480"></a>

Volcano组件支持交换机的亲和性调度。使用该功能需要上传交换机与服务器节点的对应关系以供Volcano使用，操作步骤如下。

>[!NOTE] 说明 
>当前只支持训练和推理任务进行整卡的交换机亲和性调度，不支持静态或动态vNPU调度。

**操作步骤<a name="section7172163412209"></a>**

1.  <a name="li6319161364017"></a>准备部署环境的网络设计LLD文档，将其上传到K8s管理节点的任意目录（以“/home/tor-affinity”为例）。

    >[!NOTE] 说明 
    >LLD文件名需要是lld.xlsx。

2.  获取LLD文档解析脚本。

    进入[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)仓库，根据[mindcluster-deploy开源仓版本说明](./appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。下载“samples/utils”目录中的lld\_to\_cm.py文件，将该文件上传到管理节点[步骤1](#li6319161364017)中的目录下。

3.  执行以下命令，启动“lld\_to\_cm.py”脚本。

    ```
    python ./lld_to_cm.py --num 32
    ```

    >[!NOTE] 说明 
    >-   使用--num（或-n）子命令指定一个交换机下的节点个数，不指定该参数时默认取值为4。
    >-   使用--level（或-l）子命令指定交换机组网类型，不指定该参数时默认取值为double\_layer，取值说明如下。
    >    -   single\_layer：使用单层交换机组网。
    >    -   double\_layer：使用双层交换机组网。
    >-   该脚本需要使用到openpyxl模块，如果安装环境缺少该模块，可以使用**pip install openpyxl**命令进行安装。

4.  执行以下命令，检查ConfigMap是否创建成功。

    ```
    kubectl get cm -n kube-system basic-tor-node-cm
    ```

    回显示例如下，表示创建成功。

    ```
    NAME                DATA   AGE
    basic-tor-node-cm   1      8s
    ```

**配置交换机亲和性调度<a name="section125904488511"></a>**

配置交换机的亲和性调度需要在任务YAML中配置tor-affinity参数，tor-affinity的位置和配置说明如下表所示。

**表 1**  YAML参数说明

<a name="table325141716575"></a>
|参数|取值|说明|
|--|--|--|
|(.kind=="AscendJob").metadata.labels.tor-affinity|<ul><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note"><span class="notetitle"> 说明： </span><div class="notebody"><p>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p></div></div></li></ul>|<p>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p><ul><li>交换机亲和性调度1.0版本支持Atlas 训练系列产品和<term>Atlas A2 训练系列产品</term>；支持PyTorch和MindSpore框架。</li><li>交换机亲和性调度2.0版本支持<term>Atlas A2 训练系列产品</term>；支持PyTorch框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>|

#### （可选）集成昇腾插件扩展开源Volcano<a name="ZH-CN_TOPIC_0000002511426365"></a>

集群调度提供的Volcano组件是在开源Volcano的基础上新增了关于NPU调度相关的功能，该功能可通过集成集群调度为开发者提供的Ascend-volcano-plugin插件实现。开源[Volcano](https://volcano.sh/zh/#home_slider)框架支持插件机制供用户注册调度插件，实现不同的调度策略。

>[!NOTE] 说明 
>Ascend-volcano-plugin目前支持开源Volcano  v1.7.0、v1.9.0、v1.10.0、v1.11.0和v1.12.0版本，且未对开源Volcano框架做修改。

**操作步骤<a name="section2672154791712"></a>**

1.  依次执行以下命令，在“$GOPATH/src/volcano.sh/”目录下拉取Volcano版本（以v1.7为例）官方开源代码。

    ```
    mkdir -p $GOPATH/src/volcano.sh/
    cd $GOPATH/src/volcano.sh/ 
    git clone -b release-1.7 https://github.com/volcano-sh/volcano.git
    ```

2.  将获取的[ascend-for-volcano](https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-for-volcano)源码重命名为ascend-volcano-plugin，并上传至开源Volcano官方开源代码的插件路径下（“_$GOPATH_/src/volcano.sh/volcano/pkg/scheduler/plugins/”）。
3.  <a name="li627818212613"></a>依次执行以下命令，编译开源Volcano二进制文件和华为NPU调度插件so文件。根据开源代码版本，为build.sh脚本选择对应的参数，如v1.7.0。

    ```
    cd $GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build
    chmod +x build.sh
    ./build.sh v1.7.0
    ```

    >[!NOTE] 说明 
    >编译出的二进制文件和动态链接库文件在“$GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output”目录下。

    编译后的文件列表见[表1](#table5623201371819)。

    **表 1**  output路径下的文件

    <a name="table5623201371819"></a>
    |文件名|说明|
    |--|--|
    |volcano-npu-<em>{version}</em>.so|华为NPU调度插件动态链接库|
    |Dockerfile-scheduler|volcano-scheduler镜像构建文本文件|
    |Dockerfile-controller|volcano-controller镜像构建文本文件|
    |volcano-<em>v{version}</em>.yaml|Volcano的启动配置文件|
    |vc-scheduler|volcano-scheduler组件二进制文件|
    |vc-controller-manager|volcano-controller组件二进制文件|

4.  选择以下两种方式之一，启动volcano-scheduler组件。
    -   使用集群调度组件提供的启动YAML，启动volcano-scheduler组件。
        1.  执行以下命令，制作Volcano镜像。根据开源代码版本，为镜像选择对应的参数，如v1.7.0。

            ```
            docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f ./Dockerfile-scheduler
            ```

        2.  执行以下命令，启动volcano-scheduler组件。

            ```
            kubectl apply -f volcano-v{version}.yaml
            ```

            启动示例如下。

            ```
            namespace/volcano-system created
            namespace/volcano-monitoring created
            configmap/volcano-scheduler-configmap createdserviceaccount/volcano-scheduler created
            clusterrole.rbac.authorization.k8s.io/volcano-scheduler created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-scheduler-role created
            deployment.apps/volcano-scheduler createdservice/volcano-scheduler-service created
            serviceaccount/volcano-controllers created
            clusterrole.rbac.authorization.k8s.io/volcano-controllers createdclusterrolebinding.rbac.authorization.k8s.io/volcano-controllers-role created
            deployment.apps/volcano-controllers created
            customresourcedefinition.apiextensions.k8s.io/jobs.batch.volcano.sh createdcustomresourcedefinition.apiextensions.k8s.io/commands.bus.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/podgroups.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/queues.scheduling.volcano.shcreated
            customresourcedefinition.apiextensions.k8s.io/numatopologies.nodeinfo.volcano.sh created
            ```

    -   使用开源Volcano的启动YAML，启动volcano-scheduler组件。
        1.  将步骤[3](#li627818212613)中编译出的volcano-npu-_\{version\}_.so文件拷贝到开源Volcano的“$GOPATH/src/volcano.sh/volcano”目录下；在开源Volcano的Dockerfile（路径为“$GOPATH/src/volcano.sh/volcano/installer/dockerfile/scheduler/Dockerfile”）中添加如下命令。

            ```
            FROM golang:1.19.1 AS builder
            WORKDIR /go/src/volcano.sh/
            ADD . volcano
            RUN cd volcano && make vc-scheduler
            FROM alpine:latest
            COPY --from=builder /go/src/volcano.sh/volcano/_output/bin/vc-scheduler /vc-scheduler
            COPY volcano-npu_*.so plugins/     #新增
            ENTRYPOINT ["/vc-scheduler"]
            ```

        2.  依次执行以下命令，制作Volcano镜像。根据开源代码版本，为镜像选择对应的参数，如v1.7.0。

            ```
            cd $GOPATH/src/volcano.sh/volcano
            docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f installer/dockerfile/scheduler/Dockerfile
            ```

        3.  修改volcano-development.yaml，该文件路径为“$GOPATH/src/volcano.sh/volcano/installer/volcano-development.yaml”。

            ```
            apiVersion: v1
            kind: ConfigMap
            metadata: 
              name: volcano-scheduler-configmap 
              namespace: volcano-system
            data:
               volcano-scheduler.conf: |
                 actions: "enqueue, allocate, backfill"
                 tiers:
                 - plugins:
                   - name: priority
                   - name: gang
                     enablePreemptable: false
                   - name: conformance
                   - name: volcano-npu_v7.3.0_linux-x86_64    # 在ConfigMap中的新增自定义调度插件，请注意保持组件的版本配套关系
                 - plugins:
                   - name: overcommit
                   - name: drf
                     enablePreemptable: false
                   - name: predicates
                   - name: proportion
                   - name: nodeorder
                   - name: binpack
                configurations:           # 新增以下字段，该字段为Volcano配置字段
                  - name: init-params
                    arguments: {"grace-over-time":"900","presetVirtualDevice":"true","nslb-version":"1.0","shared-tor-num":"2","useClusterInfoManager":"false","super-pod-size": "48","reserve-nodes": "2"}
            ...
            kind: Deployment
            apiVersion: apps/v1
            metadata:
              name: volcano-scheduler
              namespace: volcano-system
              labels:
                app: volcano-scheduler
            spec:
              ...
              template:
            ...
                    - name: volcano-scheduler
                      image: volcanosh/vc-scheduler:v1.7.0
                      args:
                        - --logtostderr
                        - --scheduler-conf=/volcano.scheduler/volcano-scheduler.conf
                        - --enable-healthz=true   
                        - --enable-metrics=true      # v1.12.0版本必须使用该参数
                        - --plugins-dir=plugins       # 在volcano-scheduler启动命令中加载自定义插件
                        - -v=3
                        - 2>&1
            ---
            # Source: volcano/templates/scheduler.yaml
            kind: ClusterRole
            apiVersion: rbac.authorization.k8s.io/v1
            metadata:
              name: volcano-scheduler
            rules:
            ...
              - apiGroups: ["nodeinfo.volcano.sh"]
                resources: ["numatopologies"]
                verbs: ["get", "list", "watch", "delete"]
              - apiGroups: [""]                          # 新增services的get权限  
                resources: ["services"]
                verbs: ["get"]
              - apiGroups: [""]
                resources: ["configmaps"]
                verbs: ["get", "create", "delete", "update","list","watch"]    # 新增ConfigMap的list和watch权限
              - apiGroups: ["apps"]
                resources: ["daemonsets", "replicasets", "statefulsets"]
                verbs: ["list", "watch", "get"]
            ...
            ```

        4.  执行以下命令，启动volcano-scheduler组件。

            ```
            kubectl apply -f installer/volcano-development.yaml
            ```

            回显示例如下。

            ```
            namespace/volcano-system created
            namespace/volcano-monitoring created
            serviceaccount/volcano-admission created
            configmap/volcano-admission-configmap created
            clusterrole.rbac.authorization.k8s.io/volcano-admission created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-admission-role created
            service/volcano-admission-service createddeployment.apps/volcano-admission created
            job.batch/volcano-admission-init created
            customresourcedefinition.apiextensions.k8s.io/jobs.batch.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/commands.bus.volcano.sh created
            serviceaccount/volcano-controllers created
            clusterrole.rbac.authorization.k8s.io/volcano-controllers created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-controllers-role created
            deployment.apps/volcano-controllers created
            serviceaccount/volcano-scheduler createdconfigmap/volcano-scheduler-configmap created
            clusterrole.rbac.authorization.k8s.io/volcano-scheduler created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-scheduler-role createdservice/volcano-scheduler-service created
            deployment.apps/volcano-scheduler created
            customresourcedefinition.apiextensions.k8s.io/podgroups.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/queues.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/numatopologies.nodeinfo.volcano.sh created
            mutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-pods-mutate createdmutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-queues-mutate createdmutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-podgroups-mutate createdmutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-jobs-mutate createdvalidatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-jobs-validate createdvalidatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-pods-validate createdvalidatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-queues-validate created
            ```

### ClusterD<a name="ZH-CN_TOPIC_0000002511346341"></a>

-   使用整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须安装ClusterD。集群中同时存在Ascend Device Plugin和NodeD组件时，ClusterD才能提供全量的信息收集服务。
-   在安装ClusterD时，建议提前安装Volcano。若ClusterD先于Volcano安装，ClusterD所在的Pod可能会CrashLoopBackOff，需等待Volcano的Pod启动后，ClusterD才会恢复正常。
-   仅使用容器化支持和资源监测的用户，可以不安装ClusterD，请直接跳过本章节。
-   使用慢节点&慢网络故障功能前，需安装ClusterD，详细说明请参见[慢节点&慢网络故障](./usage/resumable_training.md#慢节点慢网络故障)。

**操作步骤<a name="section20114193212615"></a>**

1.  以root用户登录K8s管理节点，并执行以下命令，查看ClusterD镜像和版本号是否正确。

    ```
    docker images | grep clusterd
    ```

    回显示例如下：

    ```
    clusterd                   v7.3.0              c532e9d0889c        About an hour ago         126MB
    ```

    -   是，执行[步骤2](#li615118054419)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作和分发。

2.  <a name="li615118054419"></a>将ClusterD软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3.  如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中ClusterD的启动参数。启动参数请参见[表2](#table11614104894617)，可以在ClusterD二进制包的目录下执行<b>./clusterd -h</b>查看参数说明。
4.  （可选）在“clusterd-v<i>\{version\}</i>.yaml”中，配置人工隔离芯片检测开关及故障频率、解除隔离时间等。

    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: clusterd-config-cm
      namespace: cluster-system
    data:
      manually_separate_policy.conf: |
        enabled: true
        separate:
          fault_window_hours: 24
          fault_threshold: 3
        release:
          fault_free_hours: 48

    ```

    **表 1**  manually_separate_policy.conf的参数说明

    <a name="table208901"></a>
    |一级参数|二级参数|类型|说明|
    |--|--|--|--|
    |enabled|-|bool|人工隔离芯片的检测开关。取值包括：<ul><li>true：开启人工隔离芯片检测功能。</li><li>false：关闭人工隔离芯片检测功能。</li></ul><p>默认值为true。若关闭该开关，会将所有ClusterD人工隔离的芯片及相关缓存都清除。</p>|
    |separate|fault_window_hours|int|人工隔离芯片的时间。在该时间内，同一个故障码的故障次数达到fault_threshold取值，ClusterD会将故障芯片进行人工隔离。取值范围为[1, 720]，默认值为24，单位为h（小时）。|
    |-|fault_threshold|int|人工隔离芯片的阈值。取值范围为[1, 50]，默认值为3，单位为次。|
    |release|fault_free_hours|int|解除隔离的时间，表示距离最后一次达到频率进行隔离的时间，超过该时间会解除隔离。取值范围为[1, 240]或-1，默认值为48，单位为h（小时）。<ul><li>最后一次达到频率的时间即为clusterd-manual-info-cm中的LastSeparateTime。clusterd-manual-info-cm的说明请参见[clusterd-manual-info-cm](./api/clusterd.md#集群资源)。</li><li>配置为-1，表示关闭解除隔离功能。</li><li>达到解除隔离时间进行自动解除隔离时，无论故障是否恢复，都会解除。</li></ul>|
5.  在管理节点的YAML所在路径，执行以下命令，启动ClusterD。

    ```
    kubectl apply -f clusterd-v{version}.yaml
    ```

    启动示例如下：

    ```
    clusterrolebinding.rbac.authorization.k8s.io/pods-clusterd-rolebinding created
    lease.coordination.k8s.io/cluster-info-collector created
    deployment.apps/clusterd created
    service/clusterd-grpc-svc created
    ```

6.  执行以下命令，查看组件是否启动成功。

    ```
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```
    NAME                          READY   STATUS              RESTARTS   AGE
    clusterd-7844cb867d-fwcj7     0/1     Running            0          45s
    ```

>[!NOTE] 说明 
>-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
>-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section1250239182212"></a>**

**表 2** ClusterD启动参数

<a name="table11614104894617"></a>
<table><thead align="left"><tr id="row2614114884616"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p961416489463"><a name="p961416489463"></a><a name="p961416489463"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="14.979999999999999%" id="mcps1.2.5.1.2"><p id="p6614174812464"><a name="p6614174812464"></a><a name="p6614174812464"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.5.1.3"><p id="p12614194844618"><a name="p12614194844618"></a><a name="p12614194844618"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p261454810466"><a name="p261454810466"></a><a name="p261454810466"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row14614134874619"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p86145488460"><a name="p86145488460"></a><a name="p86145488460"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p20614848194617"><a name="p20614848194617"></a><a name="p20614848194617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p26141489467"><a name="p26141489467"></a><a name="p26141489467"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5614648134615"><a name="p5614648134615"></a><a name="p5614648134615"></a>查询<span id="ph1950137183918"><a name="ph1950137183918"></a><a name="ph1950137183918"></a>ClusterD</span>版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row56148488467"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p661410488460"><a name="p661410488460"></a><a name="p661410488460"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p561484814469"><a name="p561484814469"></a><a name="p561484814469"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p8614134874616"><a name="p8614134874616"></a><a name="p8614134874616"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6614174884615"><a name="p6614174884615"></a><a name="p6614174884615"></a>日志级别：</p>
<a name="ul76142481467"></a><a name="ul76142481467"></a><ul id="ul76142481467"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>
</td>
</tr>
<tr id="row1961574813469"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14615748144617"><a name="p14615748144617"></a><a name="p14615748144617"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p76151484466"><a name="p76151484466"></a><a name="p76151484466"></a>日志备份时间，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row1061520484468"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p161512481467"><a name="p161512481467"></a><a name="p161512481467"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16615048134619"><a name="p16615048134619"></a><a name="p16615048134619"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1668892293119"><a name="p1668892293119"></a><a name="p1668892293119"></a>/var/log/mindx-dl/clusterd/clusterd.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：clusterd-触发转储的时间.log，如：clusterd-2024-06-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row8615248184611"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p361516481465"><a name="p361516481465"></a><a name="p361516481465"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row147481810102010"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p15748191011204"><a name="p15748191011204"></a><a name="p15748191011204"></a>-useProxy</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p17830536152010"><a name="p17830536152010"></a><a name="p17830536152010"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13748141013205"><a name="p13748141013205"></a><a name="p13748141013205"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p2748131042020"><a name="p2748131042020"></a><a name="p2748131042020"></a>是否使用代理转发gRPC请求。</p>
<a name="ul71770166215"></a><a name="ul71770166215"></a><ul id="ul71770166215"><li>true：是</li><li>false：否<div class="note" id="note12300045132119"><a name="note12300045132119"></a><a name="note12300045132119"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p17300245162118"><a name="p17300245162118"></a><a name="p17300245162118"></a>建议在启动YAML中将本参数取值配置为“true”，并对ClusterD进行安全加固，详细说明请参见<a href="./references.md#clusterd安全加固">ClusterD安全加固</a>章节。</p>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row2615144813463"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1061594884617"><a name="p1061594884617"></a><a name="p1061594884617"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16151748144614"><a name="p16151748144614"></a><a name="p16151748144614"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13615048184615"><a name="p13615048184615"></a><a name="p13615048184615"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16616174834615"><a name="p16616174834615"></a><a name="p16616174834615"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>

### Ascend Operator<a name="ZH-CN_TOPIC_0000002479386414"></a>

-   使用整卡调度（训练）、静态vNPU调度（训练）、断点续训或弹性训练的用户，必须安装Ascend Operator组件。如果使用Volcano组件作为调度器，需要先安装Volcano组件，否则Ascend Operator会启动失败。
-   使用整卡调度（推理）和推理卡故障重调度特性，下发acjob类型的分布式推理任务，必须安装Ascend Operator。
-   仅使用容器化支持和资源监测、推理卡故障恢复或推理卡故障重调度（单机任务）的用户，可以不安装Ascend Operator，请直接跳过本章节。

>[!NOTE] 说明 
>Ascend Operator组件允许创建的单个AscendJob任务的最大副本数量为20000。

**操作步骤<a name="section209273712583"></a>**

1.  以root用户登录K8s管理节点，并执行以下命令，查看Ascend Operator镜像和版本号是否正确。

    ```
    docker images | grep ascend-operator
    ```

    回显示例如下：

    ```
    ascend-operator                      v7.3.0              c532e9d0889c        About an hour ago         137MB
    ```

    -   是，执行[步骤2](#li19793191914420)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作和分发。

2.  <a name="li19793191914420"></a>将Ascend Operator软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3.  如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Ascend Operator的启动参数。启动参数请参见[表1](#table11614104894617)，可执行<b>./ascend-operator -h</b>查看参数说明。
4.  （可选）使用Ascend Operator为PyTorch和MindSpore框架下的训练任务生成集合通信配置文件（RankTable File，也叫[hccl.json](./appendix.md#hccljson文件说明)文件），缩短集群通信建链时间。使用其他框架的用户，可跳过本步骤。
    1.  启动YAML中已经默认挂载了hccl.json文件的父目录，用户可以根据实际情况进行修改。

        ```
        ...
                - name: ranktable-dir
                  mountPath: /user/mindx-dl/ranktable        # 容器内路径，不可修改
        ...
              volumes:
                - name: ascend-operator-log
                  hostPath:
                    path: /var/log/mindx-dl/ascend-operator
                    type: Directory
                - name: ranktable-dir
                  hostPath:
                    path: /user/mindx-dl/ranktable    # 宿主机路径，任务YAML中hccl.json文件保存路径的根目录必须和宿主机路径保持一致
                    type: DirectoryOrCreate                                      # 用于检查给定文件夹是否存在，若不存在，则会创建空文件夹。
        ...
        ```

        >[!NOTE] 说明 
        >-   容器内RankTable根目录路径不可修改，宿主机路径可以修改。用户部署任务时，任务YAML中hccl.json文件保存路径的**根目录**必须和宿主机路径保持一致。
        >-   RankTable根目录文件夹权限，必须满足以下任意一个条件。
        >    -   所属的用户和用户组为hwMindX（集群调度组件默认的运行用户）。
        >    -   RankTable根目录文件夹权限为777。

    2.  执行以下命令，在父目录下创建hccl.json文件的具体挂载路径。

        ```
        mkdir -m 777 /user/mindx-dl/ranktable/{具体挂载路径}
        ```

5.  在管理节点的YAML所在路径，执行以下命令，启动Ascend Operator。

    ```
    kubectl apply -f ascend-operator-v{version}.yaml
    ```

    启动示例如下：

    ```
    deployment.apps/ascend-operator-manager created
    serviceaccount/ascend-operator-manager created
    clusterrole.rbac.authorization.k8s.io/ascend-operator-manager-role created
    clusterrolebinding.rbac.authorization.k8s.io/ascend-operator-manager-rolebinding created
    customresourcedefinition.apiextensions.k8s.io/ascendjobs.mindxdl.gitee.com created
    ...
    ```

6.  执行以下命令，查看组件是否启动成功。

    ```
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```
    NAME                                         READY   STATUS    RESTARTS   AGE
    ...
    ascend-operator-7667495b6b-hwmjw      1/1    Running  0         11s
    ```

>[!NOTE] 说明 
>-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
>-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section91521925121114"></a>**

**表 1** Ascend Operator启动参数

<a name="table11614104894617"></a>
<table><thead align="left"><tr id="row2614114884616"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p961416489463"><a name="p961416489463"></a><a name="p961416489463"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="14.979999999999999%" id="mcps1.2.5.1.2"><p id="p6614174812464"><a name="p6614174812464"></a><a name="p6614174812464"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.5.1.3"><p id="p12614194844618"><a name="p12614194844618"></a><a name="p12614194844618"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p261454810466"><a name="p261454810466"></a><a name="p261454810466"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row14614134874619"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p86145488460"><a name="p86145488460"></a><a name="p86145488460"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p20614848194617"><a name="p20614848194617"></a><a name="p20614848194617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p26141489467"><a name="p26141489467"></a><a name="p26141489467"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5614648134615"><a name="p5614648134615"></a><a name="p5614648134615"></a>是否查询<span id="ph446121313413"><a name="ph446121313413"></a><a name="ph446121313413"></a>Ascend Operator</span>版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row56148488467"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p661410488460"><a name="p661410488460"></a><a name="p661410488460"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p561484814469"><a name="p561484814469"></a><a name="p561484814469"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p8614134874616"><a name="p8614134874616"></a><a name="p8614134874616"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p2312104517312"><a name="p2312104517312"></a><a name="p2312104517312"></a>日志级别支持如下几种取值：</p>
<a name="ul76142481467"></a><a name="ul76142481467"></a><ul id="ul76142481467"><li>取值为-1：debug</li><li>取值为0：info</li><li>取值为1：warning</li><li>取值为2：error</li><li>取值为3：critical</li></ul>
</td>
</tr>
<tr id="row1961574813469"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14615748144617"><a name="p14615748144617"></a><a name="p14615748144617"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p76151484466"><a name="p76151484466"></a><a name="p76151484466"></a>日志备份时间限制，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row1061520484468"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p161512481467"><a name="p161512481467"></a><a name="p161512481467"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16615048134619"><a name="p16615048134619"></a><a name="p16615048134619"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p56159486469"><a name="p56159486469"></a><a name="p56159486469"></a>/var/log/mindx-dl/ascend-operator/ascend-operator.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：ascend-operator-触发转储的时间.log，如：ascend-operator-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row8615248184611"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p361516481465"><a name="p361516481465"></a><a name="p361516481465"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row25282845417"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p155314286546"><a name="p155314286546"></a><a name="p155314286546"></a>-enableGangScheduling</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p19531128135415"><a name="p19531128135415"></a><a name="p19531128135415"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p55362825414"><a name="p55362825414"></a><a name="p55362825414"></a>true</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p3537285549"><a name="p3537285549"></a><a name="p3537285549"></a>是否启用“gang”策略调度，默认开启。开启时根据任务指定的调度器进行任务调度。“gang”策略调度说明请参见<a href="https://volcano.sh/zh/docs/v1-7-0/plugins/" target="_blank" rel="noopener noreferrer">开源Volcano官方文档</a>。</p>
<a name="ul1161205685015"></a><a name="ul1161205685015"></a><ul id="ul1161205685015"><li>true：启用“gang”策略调度。<p id="p1469315258274"><a name="p1469315258274"></a><a name="p1469315258274"></a>使用Job级别弹性扩缩容功能时，需将本字段的取值设置为true。</p>
</li><li>false：不启用“gang”策略调度。</li></ul>
</td>
</tr>
<tr id="row1758314497918"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1231919131198"><a name="p1231919131198"></a><a name="p1231919131198"></a>-isCompress</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p631912134197"><a name="p631912134197"></a><a name="p631912134197"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p031911314198"><a name="p031911314198"></a><a name="p031911314198"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1131914134197"><a name="p1131914134197"></a><a name="p1131914134197"></a>当日志文件大小达到转储阈值时，是否对日志文件进行压缩转储（该参数后面将会弃用）。</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true：压缩转储。</li><li>false：不压缩转储。</li></ul>
</td>
</tr>
<tr id="row1636910277610"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p836962710617"><a name="p836962710617"></a><a name="p836962710617"></a>-kubeconfig</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p1536942715620"><a name="p1536942715620"></a><a name="p1536942715620"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p7628122811369"><a name="p7628122811369"></a><a name="p7628122811369"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1231013103910"><a name="p1231013103910"></a><a name="p1231013103910"></a>kubeconfig的路径，当程序运行于集群外时必须配置。</p>
</td>
</tr>
<tr id="row57381540134219"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14739164054215"><a name="p14739164054215"></a><a name="p14739164054215"></a>-kubeApiBurst</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p0739184012425"><a name="p0739184012425"></a><a name="p0739184012425"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1273915409420"><a name="p1273915409420"></a><a name="p1273915409420"></a>100</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5739114017421"><a name="p5739114017421"></a><a name="p5739114017421"></a>与K8s通信时使用的突发流量。取值范围为（0,10000]，不在取值范围内使用默认值100。</p>
</td>
</tr>
<tr id="row182053596442"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p4205165917447"><a name="p4205165917447"></a><a name="p4205165917447"></a>-kubeApiQps</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p42051059174419"><a name="p42051059174419"></a><a name="p42051059174419"></a>float32</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17205159154412"><a name="p17205159154412"></a><a name="p17205159154412"></a>50</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p172054590444"><a name="p172054590444"></a><a name="p172054590444"></a>与K8s通信时使用的QPS（每秒请求率）。取值范围为（0,10000]，不在取值范围内使用默认值50。</p>
</td>
</tr>
<tr id="row2615144813463"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1061594884617"><a name="p1061594884617"></a><a name="p1061594884617"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16151748144614"><a name="p16151748144614"></a><a name="p16151748144614"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13615048184615"><a name="p13615048184615"></a><a name="p13615048184615"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16616174834615"><a name="p16616174834615"></a><a name="p16616174834615"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>

### NodeD<a name="ZH-CN_TOPIC_0000002479226406"></a>

-   使用整卡调度、静态vNPU调度、动态vNPU调度、推理卡故障恢复、推理卡故障重调度、断点续训或弹性训练时，必须安装NodeD。
-   仅使用容器化支持和资源监测的用户，可以不安装NodeD，请直接跳过本章节。
-   使用慢节点&慢网络故障功能前，需安装NodeD，详细说明请参见[慢节点&慢网络故障](./usage/resumable_training.md#慢节点慢网络故障)。

**操作步骤<a name="section135381552125414"></a>**

1.  以root用户登录各计算节点，并执行以下命令查看镜像和版本号是否正确。

    ```
    docker images | grep noded
    ```

    回显示例如下：

    ```
    noded                               v7.3.0              ef801847acd2        29 minutes ago      133MB
    ```

    -   是，执行[步骤2](#li26221447455)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作和分发。

2.  <a name="li26221447455"></a>将NodeD软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3.  如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中NodeD的启动参数。启动参数请参见[表1](#table1862682843614)，可执行<b>./noded -h</b>查看参数说明。
4.  （可选）使用**断点续训**或者**弹性训练**时，需要配置节点状态上报间隔。在NodeD启动YAML文件的“args”行增加“-reportInterval”参数，如下所示：

    ```
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 -reportInterval=5" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: true
              volumeMounts:
                - name: log-noded
    ...
    ```

    >[!NOTE] 说明 
    >-   K8s[默认40秒未收到节点响应时](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)将该节点置为NotReady。
    >-   当K8s APIServer请求压力变大时，可根据实际情况增大间隔时间，以减轻APIServer压力。

5.  在管理节点的YAML所在路径，执行以下命令，启动NodeD。
    -   不使用[dpc故障检测](./usage/resumable_training.md#节点故障)功能，请执行以下命令。

        ```
        kubectl apply -f noded-v{version}.yaml
        ```

    -   如果环境已部署Scale-Out Storage DPC 24.2.0及以上版本，并且使用[dpc故障检测](./usage/resumable_training.md#节点故障)功能，则执行以下命令，启动NodeD。

        ```
        kubectl apply -f noded-dpc-v{version}.yaml
        ```

        启动示例如下：

        ```
        serviceaccount/noded created
        clusterrole.rbac.authorization.k8s.io/pods-noded-role created
        clusterrolebinding.rbac.authorization.k8s.io/pods-noded-rolebinding created
        daemonset.apps/noded created
        ```

6.  执行以下命令，查看组件是否启动成功。

    ```
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```
    NAME                              READY   STATUS    RESTARTS   AGE
    ...
    noded-fd6t8                  1/1    Running  0        74s
    ...
    ```

>[!NOTE] 说明 
>-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
>-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section1851191618362"></a>**

**表 1** NodeD启动参数

<a name="table1862682843614"></a>
<table><thead align="left"><tr id="row462602873614"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p14626028143611"><a name="p14626028143611"></a><a name="p14626028143611"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p136269286369"><a name="p136269286369"></a><a name="p136269286369"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p126271528193618"><a name="p126271528193618"></a><a name="p126271528193618"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="45%" id="mcps1.2.5.1.4"><p id="p13627192820361"><a name="p13627192820361"></a><a name="p13627192820361"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row162762819362"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p126271328193610"><a name="p126271328193610"></a><a name="p126271328193610"></a>-reportInterval</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p2062718289366"><a name="p2062718289366"></a><a name="p2062718289366"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1962732833610"><a name="p1962732833610"></a><a name="p1962732833610"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><a name="ul49338283283"></a><a name="ul49338283283"></a><ul id="ul49338283283"><li>上报节点故障信息的最小间隔，如果节点状态有变化，那么在5s内就会上报，如果节点状态持续没有变化，那么上报周期为30分钟。</li><li>取值范围为1~300，单位为秒。</li><li>当K8s APIServer请求压力变大时，可根据实际情况增大间隔时间，以减轻APIServer压力。</li></ul>
</td>
</tr>
<tr id="row1240181274312"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1691522724316"><a name="p1691522724316"></a><a name="p1691522724316"></a>-monitorPeriod</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p17916227194316"><a name="p17916227194316"></a><a name="p17916227194316"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1491652715431"><a name="p1491652715431"></a><a name="p1491652715431"></a>60</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p139161227154317"><a name="p139161227154317"></a><a name="p139161227154317"></a>节点硬件故障的轮询检测周期，取值范围为60~600，单位为秒。</p>
</td>
</tr>
<tr id="row562722803619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p862732803617"><a name="p862732803617"></a><a name="p862732803617"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p166271328153612"><a name="p166271328153612"></a><a name="p166271328153612"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p176271728143613"><a name="p176271728143613"></a><a name="p176271728143613"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p146279281367"><a name="p146279281367"></a><a name="p146279281367"></a>是否查询当前<span id="ph1437310218483"><a name="ph1437310218483"></a><a name="ph1437310218483"></a>NodeD</span>的版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row15627928153617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1627328103615"><a name="p1627328103615"></a><a name="p1627328103615"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p56272028193610"><a name="p56272028193610"></a><a name="p56272028193610"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p4627172833615"><a name="p4627172833615"></a><a name="p4627172833615"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p13627628113614"><a name="p13627628113614"></a><a name="p13627628113614"></a>日志级别：</p>
<a name="ul262712284361"></a><a name="ul262712284361"></a><ul id="ul262712284361"><li>取值为-1：debug</li><li>取值为0：info</li><li>取值为1：warning</li><li>取值为2：error</li><li>取值为3：critical</li></ul>
</td>
</tr>
<tr id="row126271928143618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13627132863613"><a name="p13627132863613"></a><a name="p13627132863613"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p662782817368"><a name="p662782817368"></a><a name="p662782817368"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p062752813611"><a name="p062752813611"></a><a name="p062752813611"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p126271289369"><a name="p126271289369"></a><a name="p126271289369"></a>日志备份时间，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row0896102832513"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p178963287252"><a name="p178963287252"></a><a name="p178963287252"></a>-resultMaxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p18896202818250"><a name="p18896202818250"></a><a name="p18896202818250"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p48961228192511"><a name="p48961228192511"></a><a name="p48961228192511"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p198961128162516"><a name="p198961128162516"></a><a name="p198961128162516"></a>pingmesh结果备份文件保留的天数。取值范围为[7, 700]，单位为天。</p>
<div class="note" id="note1058610517274"><a name="note1058610517274"></a><a name="note1058610517274"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p946415413280"><a name="p946415413280"></a><a name="p946415413280"></a>该参数仅支持在<span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span>上使用。且所使用的驱动版本需≥24.1.RC1。</p>
</div></div>
</td>
</tr>
<tr id="row86273287368"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1962772813618"><a name="p1962772813618"></a><a name="p1962772813618"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p162772823618"><a name="p162772823618"></a><a name="p162772823618"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p962817282367"><a name="p962817282367"></a><a name="p962817282367"></a>/var/log/mindx-dl/noded/noded.log</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1862816283365"><a name="p1862816283365"></a><a name="p1862816283365"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：noded-触发转储的时间.log，如：noded-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row1862892813363"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10628202814365"><a name="p10628202814365"></a><a name="p10628202814365"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p4628828173616"><a name="p4628828173616"></a><a name="p4628828173616"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p16628182814362"><a name="p16628182814362"></a><a name="p16628182814362"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1062817287368"><a name="p1062817287368"></a><a name="p1062817287368"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row68317556187"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101519"><a name="p0894319101519"></a><a name="p0894319101519"></a><span id="ph96781327191516"><a name="ph96781327191516"></a><a name="ph96781327191516"></a>-deviceResetTimeout</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151514"><a name="p108941719151514"></a><a name="p108941719151514"></a><span id="ph1899563312153"><a name="ph1899563312153"></a><a name="ph1899563312153"></a>int</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961512"><a name="p19894131961512"></a><a name="p19894131961512"></a><span id="ph67327379151"><a name="ph67327379151"></a><a name="ph67327379151"></a>60</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971510"><a name="p589551971510"></a><a name="p589551971510"></a><span id="ph4556742141516"><a name="ph4556742141516"></a><a name="ph4556742141516"></a>组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600</span><span id="ph124041056151513"><a name="ph124041056151513"></a><a name="ph124041056151513"></a>。</span></p>
<a name="ul1354220213192"></a><a name="ul1354220213192"></a><ul id="ul1354220213192"><li><span id="ph278017516257"><a name="ph278017516257"></a><a name="ph278017516257"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="ph13163257131918"><a name="ph13163257131918"></a><a name="ph13163257131918"></a>Atlas 800I A2 推理服务器</span>、<span id="ph10930753142211"><a name="ph10930753142211"></a><a name="ph10930753142211"></a>A200I A2 Box 异构组件</span>：建议配置为150秒。</li><li><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>、<span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 800I A3 超节点服务器</span>：建议配置为360秒。</li><li><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>推理服务器（插Atlas 350 标卡）</term></span>、<span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>Atlas 850 服务器</span>、<span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 950 SuperPoD 超节点/集群</span>：建议配置为600秒。</li></ul>
</td>
</tr>
<tr id="row10282191492316"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p4283714172316"><a name="p4283714172316"></a><a name="p4283714172316"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p82838147233"><a name="p82838147233"></a><a name="p82838147233"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482316"><a name="p828341482316"></a><a name="p828341482316"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432318"><a name="p828311432318"></a><a name="p828311432318"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>

### Resilience Controller<a name="ZH-CN_TOPIC_0000002511426375"></a>



#### （可选）导入证书和KubeConfig<a name="ZH-CN_TOPIC_0000002479226468"></a>

**使用前必读<a name="section18169249192720"></a>**

导入工具cert-importer在组件的软件包中。

-   使用之前请先查看[导入工具说明](#section890515124614)，根据实际情况选择对应的导入步骤。
-   导入KubeConfig文件参见[导入KubeConfig文件](#section1538945217341)。

**导入工具说明<a name="section890515124614"></a>**

-   导入文件的说明请参考[表1](#table66513321527)，详细命令参数请参考[表4](#table18529165716504)。

    **表 1**  组件导入文件说明

    <a name="table66513321527"></a>
    <table><thead align="left"><tr id="row866113218219"><th class="cellrowborder" valign="top" width="19.59195919591959%" id="mcps1.2.5.1.1"><p id="p19661432425"><a name="p19661432425"></a><a name="p19661432425"></a>组件</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.88168816881688%" id="mcps1.2.5.1.2"><p id="p5118134235115"><a name="p5118134235115"></a><a name="p5118134235115"></a>导入文件类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="26.95269526952695%" id="mcps1.2.5.1.3"><p id="p99612619162"><a name="p99612619162"></a><a name="p99612619162"></a>导入命令示例</p>
    </th>
    <th class="cellrowborder" valign="top" width="36.57365736573657%" id="mcps1.2.5.1.4"><p id="p176262101716"><a name="p176262101716"></a><a name="p176262101716"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row12463182714316"><td class="cellrowborder" valign="top" width="19.59195919591959%" headers="mcps1.2.5.1.1 "><p id="p72311217103014"><a name="p72311217103014"></a><a name="p72311217103014"></a><span id="ph14361567178"><a name="ph14361567178"></a><a name="ph14361567178"></a>Resilience Controller</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="16.88168816881688%" headers="mcps1.2.5.1.2 "><p id="p33991858232"><a name="p33991858232"></a><a name="p33991858232"></a>连接<span id="ph4808918506"><a name="ph4808918506"></a><a name="ph4808918506"></a>K8s</span>的KubeConfig文件</p>
    <p id="p331133914167"><a name="p331133914167"></a><a name="p331133914167"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="26.95269526952695%" headers="mcps1.2.5.1.3 "><p id="p16682153041618"><a name="p16682153041618"></a><a name="p16682153041618"></a>./cert-importer -kubeConfig=<em id="i28511515200"><a name="i28511515200"></a><a name="i28511515200"></a>{kubeFile}</em>  -cpt=<em id="i11887152317202"><a name="i11887152317202"></a><a name="i11887152317202"></a>{component}</em></p>
    <p id="p115141742151614"><a name="p115141742151614"></a><a name="p115141742151614"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="36.57365736573657%" headers="mcps1.2.5.1.4 "><p id="p2833102831511"><a name="p2833102831511"></a><a name="p2833102831511"></a>由于<span id="ph88891493615"><a name="ph88891493615"></a><a name="ph88891493615"></a>K8s</span>自带的ServiceAccount的token文件会挂载到物理机上，有暴露风险，可通过外部导入加密KubeConfig文件替换ServiceAccount进行安全加固。</p>
    <p id="p18105124517162"><a name="p18105124517162"></a><a name="p18105124517162"></a></p>
    </td>
    </tr>
    </tbody>
    </table>

-   工具支持的操作如[表2](#table13221181211509)。

    **表 2**  操作说明

    <a name="table13221181211509"></a>
    <table><thead align="left"><tr id="row4222141214502"><th class="cellrowborder" valign="top" width="15.709999999999999%" id="mcps1.2.3.1.1"><p id="p6222131285015"><a name="p6222131285015"></a><a name="p6222131285015"></a>操作</p>
    </th>
    <th class="cellrowborder" valign="top" width="84.28999999999999%" id="mcps1.2.3.1.2"><p id="p1222181295014"><a name="p1222181295014"></a><a name="p1222181295014"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row022271217502"><td class="cellrowborder" valign="top" width="15.709999999999999%" headers="mcps1.2.3.1.1 "><p id="p1822219129505"><a name="p1822219129505"></a><a name="p1822219129505"></a>新增</p>
    </td>
    <td class="cellrowborder" valign="top" width="84.28999999999999%" headers="mcps1.2.3.1.2 "><p id="p622220128501"><a name="p622220128501"></a><a name="p622220128501"></a>导入KubeConfig等文件。</p>
    </td>
    </tr>
    <tr id="row1622231295011"><td class="cellrowborder" valign="top" width="15.709999999999999%" headers="mcps1.2.3.1.1 "><p id="p132221512105017"><a name="p132221512105017"></a><a name="p132221512105017"></a>更新</p>
    </td>
    <td class="cellrowborder" valign="top" width="84.28999999999999%" headers="mcps1.2.3.1.2 "><p id="p147469919538"><a name="p147469919538"></a><a name="p147469919538"></a>导入新的KubeConfig等文件，替换旧的文件。</p>
    <p id="p922217125500"><a name="p922217125500"></a><a name="p922217125500"></a>重新导入后，需要重启业务组件才生效。请提前规划证书的有效期，有效期要匹配产品生命周期，不能过长或者过短，避免业务组件重启导致业务中断。</p>
    </td>
    </tr>
    </tbody>
    </table>

-   默认情况下，导入成功后，工具会自动删除KubeConfig授权文件，用户可通过<b>-n</b>参数停用自动删除功能。如果不自动删除，用户应妥善保管相关配置文件，如果决定不再使用相关文件，请立即删除，防止意外泄露。
-   导入的文件会被重新加密并存入“/etc/mindx-dl”目录中，具体参考[表3](#table252713572507)。
-   如果从3.0.RC3及以后版本降级到3.0.RC3之前的旧版本，需在手动删除“/etc/mindx-dl/”目录下的文件后，重新使用旧版cert-importer工具导入。
-   导入工具加密需要系统有足够的熵池（random pool）。如果熵池不够，程序可能阻塞，可以安装haveged组件来进行补熵。

    安装命令可参考：

    -   类似CentOS操作系统执行**yum install haveged -y**命令进行安装，并执行**systemctl start haveged**命令启动haveged组件。
    -   类似Ubuntu操作系统执行**apt install haveged -y**命令进行安装，并执行**systemctl start haveged**命令启动haveged组件。

**导入KubeConfig文件<a name="section1538945217341"></a>**

1.  登录K8s管理节点。
2.  创建“/etc/kubernetes/mindxdl”文件夹，权限设置为750。

    ```
    rm -rf /etc/kubernetes/mindxdl
    mkdir /etc/kubernetes/mindxdl
    chmod 750 /etc/kubernetes/mindxdl
    ```

3.  参考[Kubernetes相关指导](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)自行创建名为resilience-controller-cfg.conf的KubeConfig文件，其中KubeConfig文件中的“user”字段为“resilience-controller”。将KubeConfig文件放到“/etc/kubernetes/mindxdl/”路径下。
4.  进入Resilience Controller安装包解压路径，将lib文件夹设置到当前窗口的环境变量LD\_LIBRARY\_PATH中，不需要持久化或继承给其他用户（证书导入工具需要配置自带的加密组件相关的so包路径）。
    1.  执行以下命令，将环境变量进行备份。

        ```
        export LD_LIBRARY_PATH_BAK=${LD_LIBRARY_PATH}
        ```

    2.  执行以下命令，将lib文件夹设置到当前环境变量LD\_LIBRARY\_PATH中。

        ```
        export LD_LIBRARY_PATH=`pwd`/lib/:${LD_LIBRARY_PATH}
        ```

5.  执行以下命令，为Resilience Controller组件导入KubeConfig文件。

    ```
    ./cert-importer -kubeConfig=/etc/kubernetes/mindxdl/resilience-controller-cfg.conf  -cpt=rc
    ```

    回显示例如下，请以实际回显为准，出现以下字段表示导入成功。

    ```
    encrypt kubeConfig successfully
    start to write data to disk
    [OP]import kubeConfig successfully
    change owner and set file mode successfully
    ```

    >[!NOTE] 说明 
    >-   已经导入了KubeConfig配置文件，但是组件还是出现连接K8s异常的场景，可以参见[集群调度组件连接K8s异常](./faq.md#集群调度组件连接k8s异常)章节进行处理。
    >-   导入证书时，导入工具cert-importer会自动创建“/var/log/mindx-dl/cert-importer”目录，目录权限750，属主为root:root。

6.  执行以下命令，将备份的环境变量还原。

    ```
    export LD_LIBRARY_PATH=${LD_LIBRARY_PATH_BAK}
    ```

**表 3** 集群调度组件证书配置文件表

<a name="table252713572507"></a>
<table><thead align="left"><tr id="row4527257145015"><th class="cellrowborder" valign="top" width="17.5982401759824%" id="mcps1.2.5.1.1"><p id="p14528165725013"><a name="p14528165725013"></a><a name="p14528165725013"></a>组件</p>
</th>
<th class="cellrowborder" valign="top" width="19.24807519248075%" id="mcps1.2.5.1.2"><p id="p14528105765013"><a name="p14528105765013"></a><a name="p14528105765013"></a>证书等配置文件路径</p>
</th>
<th class="cellrowborder" valign="top" width="11.08889111088891%" id="mcps1.2.5.1.3"><p id="p105282572501"><a name="p105282572501"></a><a name="p105282572501"></a>目录及其文件属主</p>
</th>
<th class="cellrowborder" valign="top" width="52.064793520647946%" id="mcps1.2.5.1.4"><p id="p11528155755016"><a name="p11528155755016"></a><a name="p11528155755016"></a>配置文件说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row9528155785012"><td class="cellrowborder" valign="top" width="17.5982401759824%" headers="mcps1.2.5.1.1 "><p id="p1528155715501"><a name="p1528155715501"></a><a name="p1528155715501"></a><span id="ph1488142812262"><a name="ph1488142812262"></a><a name="ph1488142812262"></a>集群调度组件</span>证书相关根目录</p>
</td>
<td class="cellrowborder" valign="top" width="19.24807519248075%" headers="mcps1.2.5.1.2 "><p id="p7528357125019"><a name="p7528357125019"></a><a name="p7528357125019"></a>/etc/mindx-dl/</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="11.08889111088891%" headers="mcps1.2.5.1.3 "><p id="p1528457195011"><a name="p1528457195011"></a><a name="p1528457195011"></a>hwMindX:hwMindX</p>
<p id="p17618514195"><a name="p17618514195"></a><a name="p17618514195"></a></p>
<p id="p27716513196"><a name="p27716513196"></a><a name="p27716513196"></a></p>
<p id="p1532775483511"><a name="p1532775483511"></a><a name="p1532775483511"></a></p>
</td>
<td class="cellrowborder" valign="top" width="52.064793520647946%" headers="mcps1.2.5.1.4 "><p id="p9528857175013"><a name="p9528857175013"></a><a name="p9528857175013"></a>kmc_primary_store/master.ks：自动生成的主密钥，请勿删除。</p>
<p id="p1152811579509"><a name="p1152811579509"></a><a name="p1152811579509"></a>.config/backup.ks：自动生成的备份密钥，请勿删除。</p>
</td>
</tr>
<tr id="row207702393454"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p127701395458"><a name="p127701395458"></a><a name="p127701395458"></a><span id="ph1287272539"><a name="ph1287272539"></a><a name="ph1287272539"></a>Resilience Controller</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1977073944518"><a name="p1977073944518"></a><a name="p1977073944518"></a>/etc/mindx-dl/resilience-controller/</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p176150132191"><a name="p176150132191"></a><a name="p176150132191"></a>.config/config6：导入的加密<span id="ph10615131313194"><a name="ph10615131313194"></a><a name="ph10615131313194"></a>K8s</span> KubeConfig文件，连接<span id="ph761518136190"><a name="ph761518136190"></a><a name="ph761518136190"></a>K8s</span>使用。</p>
<p id="p16156132195"><a name="p16156132195"></a><a name="p16156132195"></a>.config6：导入的加密<span id="ph761517138198"><a name="ph761517138198"></a><a name="ph761517138198"></a>K8s</span> KubeConfig文件备份。</p>
</td>
</tr>
</tbody>
</table>

**表 4**  导入工具参数说明

<a name="table18529165716504"></a>
<table><thead align="left"><tr id="row1852914572501"><th class="cellrowborder" valign="top" width="17.349999999999998%" id="mcps1.2.5.1.1"><p id="p5529175745012"><a name="p5529175745012"></a><a name="p5529175745012"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="19.41%" id="mcps1.2.5.1.2"><p id="p17529185775019"><a name="p17529185775019"></a><a name="p17529185775019"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="11.01%" id="mcps1.2.5.1.3"><p id="p1352935715507"><a name="p1352935715507"></a><a name="p1352935715507"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="52.23%" id="mcps1.2.5.1.4"><p id="p1552925711509"><a name="p1552925711509"></a><a name="p1552925711509"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row55021443133913"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p117491127105717"><a name="p117491127105717"></a><a name="p117491127105717"></a>-kubeConfig</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p18750132718575"><a name="p18750132718575"></a><a name="p18750132718575"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p975010276572"><a name="p975010276572"></a><a name="p975010276572"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p9750162720574"><a name="p9750162720574"></a><a name="p9750162720574"></a>待导入的KubeConfig文件的路径。</p>
</td>
</tr>
<tr id="row45301657115017"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p8530165715011"><a name="p8530165715011"></a><a name="p8530165715011"></a>-cpt</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1353085745016"><a name="p1353085745016"></a><a name="p1353085745016"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p18558162664212"><a name="p18558162664212"></a><a name="p18558162664212"></a>rc</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p9931932155418"><a name="p9931932155418"></a><a name="p9931932155418"></a>导入证书的组件名称为rc，表示<span id="ph131541756961"><a name="ph131541756961"></a><a name="ph131541756961"></a>Resilience Controller</span>。</p>
</td>
</tr>
<tr id="row953045718504"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p5530195785020"><a name="p5530195785020"></a><a name="p5530195785020"></a>-encryptAlgorithm</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p12530125719509"><a name="p12530125719509"></a><a name="p12530125719509"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p35312571506"><a name="p35312571506"></a><a name="p35312571506"></a>9</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p135316571501"><a name="p135316571501"></a><a name="p135316571501"></a>私钥口令加密算法：</p>
<a name="ul145317578507"></a><a name="ul145317578507"></a><ul id="ul145317578507"><li>8：AES128GCM</li><li>9：AES256GCM</li></ul>
<div class="note" id="note05311457165012"><a name="note05311457165012"></a><a name="note05311457165012"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p18531125718501"><a name="p18531125718501"></a><a name="p18531125718501"></a>无效参数值会被重置为默认值。</p>
</div></div>
</td>
</tr>
<tr id="row18531135717506"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p1253175785015"><a name="p1253175785015"></a><a name="p1253175785015"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p75318575501"><a name="p75318575501"></a><a name="p75318575501"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p853175715507"><a name="p853175715507"></a><a name="p853175715507"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p35317578505"><a name="p35317578505"></a><a name="p35317578505"></a>查询<span id="ph19991165205214"><a name="ph19991165205214"></a><a name="ph19991165205214"></a>Resilience Controller</span>版本号。</p>
</td>
</tr>
<tr id="row2573635141612"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p138495616250"><a name="p138495616250"></a><a name="p138495616250"></a>-n</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p6384135614257"><a name="p6384135614257"></a><a name="p6384135614257"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p03848562252"><a name="p03848562252"></a><a name="p03848562252"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p2384145614255"><a name="p2384145614255"></a><a name="p2384145614255"></a>导入成功后是否删除<span id="ph418094814555"><a name="ph418094814555"></a><a name="ph418094814555"></a>KubeConfig</span>文件。</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true：导入成功后不删除<span id="ph39020528557"><a name="ph39020528557"></a><a name="ph39020528557"></a>KubeConfig</span>文件。</li><li>false：导入成功后删除<span id="ph7200135465511"><a name="ph7200135465511"></a><a name="ph7200135465511"></a>KubeConfig</span>文件。</li></ul>
</td>
</tr>
<tr id="row5485341194020"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p6232132695813"><a name="p6232132695813"></a><a name="p6232132695813"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p623242655813"><a name="p623242655813"></a><a name="p623242655813"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1232182685820"><a name="p1232182685820"></a><a name="p1232182685820"></a>/var/log/mindx-dl/cert-importer/cert-importer.log</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p102331826175814"><a name="p102331826175814"></a><a name="p102331826175814"></a>工具运行日志文件。转储后文件的命名格式为：cert-importer-触发转储的时间.log，如：cert-importer-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row8384164173412"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p1138412411345"><a name="p1138412411345"></a><a name="p1138412411345"></a>-updateMk</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1738494133414"><a name="p1738494133414"></a><a name="p1738494133414"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1238434133420"><a name="p1238434133420"></a><a name="p1238434133420"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p17211144255612"><a name="p17211144255612"></a><a name="p17211144255612"></a>是否更新KMC加密组件的主密钥。</p>
<a name="ul154871314165520"></a><a name="ul154871314165520"></a><ul id="ul154871314165520"><li>true：更新主密钥。</li><li>false：不更新主密钥。</li></ul>
</td>
</tr>
<tr id="row1397106345"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p53986014348"><a name="p53986014348"></a><a name="p53986014348"></a>-updateRk</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p139811020345"><a name="p139811020345"></a><a name="p139811020345"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p139850143413"><a name="p139850143413"></a><a name="p139850143413"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p741952510563"><a name="p741952510563"></a><a name="p741952510563"></a>是否更新KMC加密组件的根密钥。</p>
<a name="ul14451957145511"></a><a name="ul14451957145511"></a><ul id="ul14451957145511"><li>true：更新根密钥。</li><li>false：不更新根密钥。</li></ul>
</td>
</tr>
<tr id="row050462052716"><td class="cellrowborder" valign="top" width="17.349999999999998%" headers="mcps1.2.5.1.1 "><p id="p13504720112715"><a name="p13504720112715"></a><a name="p13504720112715"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="19.41%" headers="mcps1.2.5.1.2 "><p id="p1350422002713"><a name="p1350422002713"></a><a name="p1350422002713"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="11.01%" headers="mcps1.2.5.1.3 "><p id="p1650420209273"><a name="p1650420209273"></a><a name="p1650420209273"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="52.23%" headers="mcps1.2.5.1.4 "><p id="p4505820152717"><a name="p4505820152717"></a><a name="p4505820152717"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>

#### 安装Resilience Controller<a name="ZH-CN_TOPIC_0000002479226460"></a>

-   使用**弹性训练**时，必须安装Resilience Controller。Resilience Controller连接K8s时，可以选择使用ServiceAccount或KubeConfig文件进行认证，两种方式差异可参考[使用ServiceAccount和KubeConfig差异](./appendix.md#使用serviceaccount和kubeconfig差异)。
-   不使用**弹性训练**的用户，可以不安装Resilience Controller，请直接跳过本章节。

**操作步骤<a name="section0531457718"></a>**

1.  以root用户登录K8s管理节点，并执行以下命令，查看Resilience Controller镜像和版本号是否正确。

    ```
    docker images | grep resilience-controller
    ```

    回显示例如下：

    ```
    resilience-controller                      v7.3.0              c532e9d0889c        About an hour ago         142MB
    ```

    -   是，执行[步骤2](#li10743192474541)。
    -   否，请参见[准备镜像](#准备镜像)，完成镜像制作和分发。

2.  <a name="li10743192474541"></a>将Resilience Controller软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3.  如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Resilience Controller的启动参数。启动参数的说明请参见[表1](#table195504370194)，也可执行<b>./resilience-controller -h</b>查看参数说明。
4.  在管理节点的YAML所在路径，执行以下命令，启动Resilience Controller。

    -   如果没有导入KubeConfig证书，执行如下命令。

        ```
        kubectl apply -f resilience-controller-v{version}.yaml
        ```

    -   如果导入了KubeConfig证书，执行如下命令。

        ```
        kubectl apply -f resilience-controller-without-token-v{version}.yaml
        ```

    启动示例如下：

    ```
    root@ubuntu:/home/ascend-resilience-controller# kubectl apply -f resilience-controller-v7.3.0.yaml 
    serviceaccount/resilience-controller createdclusterrole.rbac.authorization.k8s.io/pods-resilience-controller-role createdclusterrolebinding.rbac.authorization.k8s.io/resilience-controller-rolebinding createddeployment.apps/resilience-controller created
    [root@localhost resilience-controller]# kubectl apply -f resilience-controller-without-token-v7.3.0.yaml 
    deployment.apps/resilience-controller created
    ```

5.  执行以下命令，查看组件是否安装成功。

    ```
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```
    NAME                                            READY    STATUS      RESTARTS   AGE
    ...
    resilience-controller-7667495b6b-hwmjw   1/1     Running   0         11s
    ...
    ```

>![](public_sys-resources/icon-note.gif) **说明：** 
>-   安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](./faq.md#组件pod状态不为running)章节进行处理。
>-   安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](./faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>-   启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](./faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>-   组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](./faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section1868556161717"></a>**

**表 1** Resilience Controller启动参数

<a name="table195504370194"></a>
<table><thead align="left"><tr id="row10550173721915"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p1855053711192"><a name="p1855053711192"></a><a name="p1855053711192"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p355063710197"><a name="p355063710197"></a><a name="p355063710197"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p055073781916"><a name="p055073781916"></a><a name="p055073781916"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3550237171920"><a name="p3550237171920"></a><a name="p3550237171920"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row3551143715196"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p65517376197"><a name="p65517376197"></a><a name="p65517376197"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p19551153781918"><a name="p19551153781918"></a><a name="p19551153781918"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p15511378194"><a name="p15511378194"></a><a name="p15511378194"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p18551173791915"><a name="p18551173791915"></a><a name="p18551173791915"></a>是否查询<span id="ph151418415511"><a name="ph151418415511"></a><a name="ph151418415511"></a>Resilience Controller</span>版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row8551137161913"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1155183715199"><a name="p1155183715199"></a><a name="p1155183715199"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p105511137141920"><a name="p105511137141920"></a><a name="p105511137141920"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p755113373192"><a name="p755113373192"></a><a name="p755113373192"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6551123716195"><a name="p6551123716195"></a><a name="p6551123716195"></a>日志级别：</p>
<a name="ul655113715194"></a><a name="ul655113715194"></a><ul id="ul655113715194"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>
</td>
</tr>
<tr id="row1455163771915"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p10551143710191"><a name="p10551143710191"></a><a name="p10551143710191"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p14551193781920"><a name="p14551193781920"></a><a name="p14551193781920"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1655193715195"><a name="p1655193715195"></a><a name="p1655193715195"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p7551183716190"><a name="p7551183716190"></a><a name="p7551183716190"></a>日志备份时间限制，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row175527378195"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p455223751910"><a name="p455223751910"></a><a name="p455223751910"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p195521937131913"><a name="p195521937131913"></a><a name="p195521937131913"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p15552137111920"><a name="p15552137111920"></a><a name="p15552137111920"></a>/var/log/mindx-dl/resilience-controller/run.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p18552143713199"><a name="p18552143713199"></a><a name="p18552143713199"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：run-触发转储的时间.log，如run-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row1655213379191"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p11552163741920"><a name="p11552163741920"></a><a name="p11552163741920"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1552137171918"><a name="p1552137171918"></a><a name="p1552137171918"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p7552193711192"><a name="p7552193711192"></a><a name="p7552193711192"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p555233718199"><a name="p555233718199"></a><a name="p555233718199"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row33119022219"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1532160192215"><a name="p1532160192215"></a><a name="p1532160192215"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p123213019227"><a name="p123213019227"></a><a name="p123213019227"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1832100102210"><a name="p1832100102210"></a><a name="p1832100102210"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1328016224"><a name="p1328016224"></a><a name="p1328016224"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>

### Container Manager<a name="ZH-CN_TOPIC_0000002524428759"></a>

Container Manager组件直接在物理机上通过二进制方式运行。

1.  使用root用户登录服务器。
2.  将获取到的Container Manager软件包上传至服务器的任意目录（以下以“/home/container-manager”目录为例）。
3.  进入“/home/container-manager”目录并进行解压操作。

    ```
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

    >[!NOTE] 说明 
    ><i><version\></i>为软件包的版本号；<i><arch\></i>为CPU架构。

4.  （可选）创建自定义故障码配置文件，自定义故障码处理级别。配置及使用详情请参见[（可选）配置芯片故障级别](./usage/appliance.md#可选配置芯片故障级别)，以下步骤不体现该文件。
5.  创建并编辑container-manager.service文件。
    1.  执行以下命令，创建container-manager.service文件。

        ```
        vi container-manager.service
        ```

    2.  参考如下内容，写入container-manager.service文件中。“ExecStart”字段中加粗的内容为启动命令，启动参数说明请参见[表1](#table872410431914)，用户可以根据实际需要进行修改。

        ```
        [Unit]
        Description=Ascend container manager
        Documentation=hiascend.com
        
        [Service]
        ExecStart=/bin/bash -c "container-manager run -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null  2>&1 &"
        Restart=always
        RestartSec=2
        KillMode=process
        Environment="GOGC=50"
        Environment="GOMAXPROCS=2"
        Environment="GODEBUG=madvdontneed=1"
        Type=forking
        User=root
        Group=root
        
        [Install]
        WantedBy=multi-user.target
        ```

    3.  按“Esc”键，输入:wq!保存并退出。

6.  创建并编辑container-manager.timer文件。通过配置timer延时启动，可保证Container Manager启动时NPU卡已就位。
    1.  执行以下命令，创建container-manager.timer文件。

        ```
        vi container-manager.timer
        ```

    2.  参考以下示例，并将其写入container-manager.timer文件中。

        ```
        [Unit]
        Description=Timer for container manager Service
        
        [Timer]
        # 设置Container Manager延时启动时间，请根据实际情况调整
        OnBootSec=60s 
        Unit=container-manager.service
        
        [Install]
        WantedBy=timers.target
        ```

    3.  按“Esc”键，输入:wq!保存并退出。

7.  依次执行以下命令，启用Container Manager服务。

    ```
    # 准备Container Manager二进制文件到PATH
    cp container-manager /usr/local/bin
    chmod 500 /usr/local/bin/container-manager
    
    # 准备Container Manager系统服务文件
    cp container-manager.service /etc/systemd/system
    cp container-manager.timer /etc/systemd/system      
    
    # 启动Container Manager系统服务
    systemctl enable container-manager.service 
    systemctl enable container-manager.timer 
    systemctl start container-manager.service
    systemctl start container-manager.timer
    ```

**参数说明<a name="section2042611570392"></a>**

**表 1** Container Manager启动参数

<a name="table872410431914"></a>
<table><thead align="left"><tr id="row57241434113"><th class="cellrowborder" valign="top" width="10.801080108010803%" id="mcps1.2.6.1.1"><p id="p1272416432118"><a name="p1272416432118"></a><a name="p1272416432118"></a>命令</p>
</th>
<th class="cellrowborder" valign="top" width="16.291629162916294%" id="mcps1.2.6.1.2"><p id="p18138161362918"><a name="p18138161362918"></a><a name="p18138161362918"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.561156115611562%" id="mcps1.2.6.1.3"><p id="p1072419431419"><a name="p1072419431419"></a><a name="p1072419431419"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="23.342334233423344%" id="mcps1.2.6.1.4"><p id="p1372464316111"><a name="p1372464316111"></a><a name="p1372464316111"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="38.00380038003801%" id="mcps1.2.6.1.5"><p id="p772517434117"><a name="p772517434117"></a><a name="p772517434117"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row1450614311118"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p5507143131115"><a name="p5507143131115"></a><a name="p5507143131115"></a>help</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p15138141392917"><a name="p15138141392917"></a><a name="p15138141392917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p623516353012"><a name="p623516353012"></a><a name="p623516353012"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p3507243131112"><a name="p3507243131112"></a><a name="p3507243131112"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p15507184331111"><a name="p15507184331111"></a><a name="p15507184331111"></a>查看帮助信息。</p>
</td>
</tr>
<tr id="row1494284312299"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p19942104322911"><a name="p19942104322911"></a><a name="p19942104322911"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p1942743162912"><a name="p1942743162912"></a><a name="p1942743162912"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p894234312917"><a name="p894234312917"></a><a name="p894234312917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p39421343132915"><a name="p39421343132915"></a><a name="p39421343132915"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p129421643102918"><a name="p129421643102918"></a><a name="p129421643102918"></a>查看<span id="ph1220617322468"><a name="ph1220617322468"></a><a name="ph1220617322468"></a>Container Manager</span>的版本信息。</p>
</td>
</tr>
<tr id="row19151746182920"><td class="cellrowborder" rowspan="8" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p215164602914"><a name="p215164602914"></a><a name="p215164602914"></a>run</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p41514652911"><a name="p41514652911"></a><a name="p41514652911"></a>-logPath</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p106467567226"><a name="p106467567226"></a><a name="p106467567226"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p1364685612219"><a name="p1364685612219"></a><a name="p1364685612219"></a>/var/log/mindx-dl/container-manager/container-manager.log</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p46466565223"><a name="p46466565223"></a><a name="p46466565223"></a>日志文件。单个日志文件超过20MB时，会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为container-manager-触发转储的时间.log，例如：container-manager-2025-11-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row17214348192911"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3645125662216"><a name="p3645125662216"></a><a name="p3645125662216"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13645175613228"><a name="p13645175613228"></a><a name="p13645175613228"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9645105618222"><a name="p9645105618222"></a><a name="p9645105618222"></a>0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1926353023718"><a name="p1926353023718"></a><a name="p1926353023718"></a>日志级别：</p>
<a name="ul15263163018377"></a><a name="ul15263163018377"></a><ul id="ul15263163018377"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>
</td>
</tr>
<tr id="row14307145012915"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p33071750112914"><a name="p33071750112914"></a><a name="p33071750112914"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p335715188373"><a name="p335715188373"></a><a name="p335715188373"></a>日志备份时间，取值范围为[7, 700]，单位为天。</p>
</td>
</tr>
<tr id="row535865213293"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7358952182915"><a name="p7358952182915"></a><a name="p7358952182915"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>转储后日志文件保留个数上限，取值范围为(0, 30]，单位为个。</p>
</td>
</tr>
<tr id="row8414634133110"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p241417348316"><a name="p241417348316"></a><a name="p241417348316"></a>-ctrStrategy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13414234183112"><a name="p13414234183112"></a><a name="p13414234183112"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p134147348319"><a name="p134147348319"></a><a name="p134147348319"></a>never</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9414134153113"><a name="p9414134153113"></a><a name="p9414134153113"></a>故障容器启停策略：</p>
<a name="ul17352545173818"></a><a name="ul17352545173818"></a><ul id="ul17352545173818"><li>never：不进行容器启停。</li><li>singleRecover：仅启停挂载单个故障芯片的容器。故障产生时，停止容器；故障恢复后，将容器重新拉起。</li><li>ringRecover：启停挂载故障芯片所关联的所有芯片的容器。故障产生时，停止容器；故障恢复后，将容器重新拉起。</li></ul>
<div class="note" id="note16897891164"><a name="note16897891164"></a><a name="note16897891164"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul370062752110"></a><a name="ul370062752110"></a><ul id="ul370062752110"><li><span id="ph646865823518"><a name="ph646865823518"></a><a name="ph646865823518"></a>Container Manager</span>在感知到芯片处于RestartRequest、RestartBusiness、FreeRestartNPU和RestartNPU类型故障时，才会进行容器启停操作。故障类型说明请参见<a href="./usage/appliance.md#故障配置说明">故障配置说明</a>中"故障码级别说明"。</li><li>当故障容器启停策略配置为singleRecover或者ringRecover时，不支持用户通过容器运行时自动重启容器，二者选其一即可。</li><li>若用户手动干预导致容器停止，可能会造成<span id="ph93985387580"><a name="ph93985387580"></a><a name="ph93985387580"></a>Container Manager</span>内存数据混乱，导致容器状态异常。</li></ul>
</div></div>
</td>
</tr>
<tr id="row16901536173117"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1069033663113"><a name="p1069033663113"></a><a name="p1069033663113"></a>-sockPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p969043633119"><a name="p969043633119"></a><a name="p969043633119"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13690153610315"><a name="p13690153610315"></a><a name="p13690153610315"></a>/run/docker.sock</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9690143653110"><a name="p9690143653110"></a><a name="p9690143653110"></a>容器运行时的sock文件，该路径不允许为软链接。</p>
</td>
</tr>
<tr id="row11407174710314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1407174713310"><a name="p1407174713310"></a><a name="p1407174713310"></a>-runtimeType</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14407247203112"><a name="p14407247203112"></a><a name="p14407247203112"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p140711477312"><a name="p140711477312"></a><a name="p140711477312"></a>docker</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p6407647193117"><a name="p6407647193117"></a><a name="p6407647193117"></a>容器运行时类型：</p>
<a name="ul8283112164115"></a><a name="ul8283112164115"></a><ul id="ul8283112164115"><li>docker：容器运行时为docker。</li><li>containerd：容器运行时为containerd。<div class="note" id="note1244216377415"><a name="note1244216377415"></a><a name="note1244216377415"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul7130194664718"></a><a name="ul7130194664718"></a><ul id="ul7130194664718"><li><span id="ph14779959144911"><a name="ph14779959144911"></a><a name="ph14779959144911"></a>Container Manager</span>仅支持管理一种容器运行时启动的容器。</li><li>当容器运行时为containerd时，仅支持管理命名空间不为moby的容器。当多个命名空间下有相同名称的容器，容器管理功能可能会出现异常。</li></ul>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row44581192384"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p945879163814"><a name="p945879163814"></a><a name="p945879163814"></a>-faultConfigPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p6458139183820"><a name="p6458139183820"></a><a name="p6458139183820"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3949155543819"><a name="p3949155543819"></a><a name="p3949155543819"></a>""</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16458189133819"><a name="p16458189133819"></a><a name="p16458189133819"></a>自定义故障配置文件路径。若不配置，则使用默认的故障码配置。自定义故障配置文件详情请参见<a href="./usage/appliance.md#故障级别配置">故障级别配置</a>。</p>
<div class="note" id="note116910214413"><a name="note116910214413"></a><a name="note116910214413"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul1246612216016"></a><a name="ul1246612216016"></a><ul id="ul1246612216016"><li>该路径不允许为软链接。</li><li>该文件权限需不高于640。</li></ul>
</div></div>
</td>
</tr>
<tr id="row441711302328"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p0417030143218"><a name="p0417030143218"></a><a name="p0417030143218"></a>status</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p4417103012320"><a name="p4417103012320"></a><a name="p4417103012320"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p041703019329"><a name="p041703019329"></a><a name="p041703019329"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p541719308323"><a name="p541719308323"></a><a name="p541719308323"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p541718306324"><a name="p541718306324"></a><a name="p541718306324"></a>查询容器恢复进度，包括容器ID、状态、状态开始时间及描述信息。容器的状态定义及变化规则详细请参见<a href="./usage/appliance.md#容器恢复">容器恢复</a>。</p>
<div class="note" id="note18966355162717"><a name="note18966355162717"></a><a name="note18966355162717"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p179661455192711"><a name="p179661455192711"></a><a name="p179661455192711"></a>如果status查询到的容器信息有误，需确认run服务是否已经终止，或者环境上启动了一个以上的<span id="ph47887203387"><a name="ph47887203387"></a><a name="ph47887203387"></a>Container Manager</span>。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>Container Manager服务已经启动后，若需要修改Container Manager的启动参数，请修改服务配置文件中的启动参数后，执行以下命令，重启Container Manager系统服务。
>```
>systemctl daemon-reload && systemctl restart container-manager
>```

## 使用工具安装<a name="ZH-CN_TOPIC_0000002479386368"></a>

借助Ascend Deployer工具可以批量安装集群调度组件，大幅度简化手动安装过程中繁琐的配置操作，简化安装流程，适用于集群场景下批量安装组件。

Ascend Deployer工具现支持的硬件产品、OS清单、安装场景请参见《MindCluster Ascend Deployer 用户指南》中的“<a href="https://gitcode.com/Ascend/ascend-deployer/blob/dev/docs/zh/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%92%8Cos%E6%B8%85%E5%8D%95">支持的产品和OS清单</a>”章节，请根据“支持部署”列的支持情况，选择是否使用Ascend Deployer工具。

如需使用Ascend Deployer工具安装，请参考《MindCluster Ascend Deployer 用户指南》中的“<a href="https://gitcode.com/Ascend/ascend-deployer/blob/dev/docs/zh/installation_guide.md#%E5%AE%89%E8%A3%85%E6%98%87%E8%85%BE%E8%BD%AF%E4%BB%B6">安装昇腾软件</a>”章节。

>[!NOTE] 说明 
>-   建议用户在使用工具安装前先了解[手动安装](#手动安装)章节中相应组件的使用约束和启动参数，可以更好地帮助用户理解组件的使用场景和功能。
>-   工具版本需要与集群调度组件的版本一致，不同版本之间不可混用。

# 组件状态确认<a name="ZH-CN_TOPIC_0000002479386390"></a>










## Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002511426307"></a>

如已安装Ascend Docker Runtime，请在所有安装了该组件的节点上执行如下步骤确认Ascend Docker Runtime的状态。

**操作步骤<a name="section44081649104318"></a>**

1.  执行以下命令，查看是否存在基础镜像。

    ```
    docker images | grep ubuntu
    ```

    回显示例如下，表示存在基础镜像ubuntu:22.04。若不存在基础镜像，可以执行**docker pull ubuntu:22.04**命令，拉取基础镜像。

    ```
    ubuntu              22.04               6526a1858e5d        2 years ago         64.2MB
    ```

2.  执行以下命令，使用Ascend Docker Runtime挂载物理芯片ID为0的芯片。

    -   Docker（或K8s集成Docker场景）。

        ```
        docker run -it -e ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 /bin/bash
        ```

    -   Containerd（或K8s集成Containerd场景）。

        执行以下命令，查看当前cgroup的版本。

        ```
        stat -fc %T /sys/fs/cgroup/
        ```

        -   若回显为tmpfs，表示当前为cgroup v1版本，执行以下命令挂载物理芯片。

            ```
            ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 containerID
            ```

        -   若回显为cgroup2fs，表示当前为cgroup v2版本，执行以下命令挂载物理芯片。

            ```
            ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 ubuntu:22.04 containerID
            ```

    >[!NOTE] 说明 
    >-   ASCEND\_VISIBLE\_DEVICES参数表示挂载的芯片ID。
    >-   containerID为用户自定义的容器ID。

3.  执行以下命令，查询芯片是否挂载成功。

    ```
    ls /dev
    ```

    若回显中存在**davinci0**字段，表示芯片挂载成功，安装Ascend Docker Runtime成功且组件功能正常。

## NPU Exporter<a name="ZH-CN_TOPIC_0000002511346363"></a>

本章节以对接Prometheus，上报Prometheus数据为例，确认NPU Exporter组件是否正常运行。

**NPU Exporter使用容器部署<a name="section1595201114126"></a>**

请在任意节点执行以下步骤验证NPU Exporter的安装状态。

1.  通过如下命令查看K8s集群中NPU Exporter的Pod，需要满足Pod的STATUS为Running，READY为1/1。如果集群中有多个节点安装了NPU Exporter，需要逐个确认。

    ```
    kubectl get pods -n npu-exporter -o wide | grep npu-exporter
    ```

    回显示例：

    ```
    npu-exporter-4ln8w   1/1     Running   0          36m   192.168.102.109   ubuntu       <none>           <none>
    ```

2.  通过如下命令查看K8s集群中NPU Exporter的日志。

    ```
    kubectl logs -n npu-exporter {npu-exporter组件的Pod名字}
    ```

    回显示例：

    ```
    [INFO]     2023/12/08 07:38:56.551173 1       hwlog/api.go:108    npu-exporter.log's logger init success
    [INFO]     2023/12/08 07:38:56.551275 1       npu-exporter/main.go:205    listen on: 0.0.0.0
    [INFO]     2023/12/08 07:38:56.551369 1       npu-exporter/main.go:325    npu exporter starting and the version is v7.3.0_linux-x86_64
    [WARN]     2023/12/08 07:38:56.684424 1       npu-exporter/main.go:339    enable unsafe http server
    [WARN]     2023/12/08 07:39:01.686205 98      container/runtime_ops.go:150    failed to get OCI connection: context deadline exceeded
    [WARN]     2023/12/08 07:39:01.686311 98      container/runtime_ops.go:152    use backup address to try again
    [INFO]     2023/12/08 07:39:01.687444 98      collector/npu_collector.go:418    Starting update cache every 5 seconds
    [WARN]     2023/12/08 07:39:01.688039 157     collector/npu_collector.go:463    get info of npu-exporter-network-info failed: no value found, so use initial net info
    [INFO]     2023/12/08 07:39:01.744739 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:01.852413 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:05.055247 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:06.688352 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:06.750876 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:09.843914 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:11.688505 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:11.701081 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:14.859243 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    ...
    ```

**NPU Exporter使用二进制部署<a name="zh-cn_topic_0000001497205429_section2976165515363"></a>**

请在安装NPU Exporter的节点执行以下步骤验证组件的安装状态。

1.  登录部署NPU Exporter的节点，使用如下命令，查看组件服务的状态，需要满足组件状态为active \(running\)。

    ```
    systemctl status npu-exporter
    ```

    回显示例：

    ```
    root@ubuntu:~# systemctl status npu-exporter
    ● npu-exporter.service - Ascend npu exporter
       Loaded: loaded (/etc/systemd/system/npu-exporter.service; enabled; vendor preset: enabled)
       Active: active (running) since Thu 2022-11-17 16:24:41 CST; 3 days ago
     Main PID: 25121 (npu-exporter)
        Tasks: 8 (limit: 7372)
       CGroup: /system.slice/npu-exporter.service
               └─25121 /usr/local/bin/npu-exporter -ip=127.0.0.1 -port=8082 -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log
    ...
    ```

2.  查看组件日志。

    ```
    cat /var/log/mindx-dl/npu-exporter/npu-exporter.log
    ```

    回显示例：

    ```
    [INFO]     2023/12/08 07:38:56.551173 1       hwlog/api.go:108    npu-exporter.log's logger init success
    [INFO]     2023/12/08 07:38:56.551275 1       npu-exporter/main.go:205    listen on: 0.0.0.0
    [INFO]     2023/12/08 07:38:56.551369 1       npu-exporter/main.go:325    npu exporter starting and the version is v7.3.0_linux-x86_64
    [WARN]     2023/12/08 07:38:56.684424 1       npu-exporter/main.go:339    enable unsafe http server
    [WARN]     2023/12/08 07:39:01.686205 98      container/runtime_ops.go:150    failed to get OCI connection: context deadline exceeded
    [WARN]     2023/12/08 07:39:01.686311 98      container/runtime_ops.go:152    use backup address to try again
    [INFO]     2023/12/08 07:39:01.687444 98      collector/npu_collector.go:418    Starting update cache every 5 seconds
    [WARN]     2023/12/08 07:39:01.688039 157     collector/npu_collector.go:463    get info of npu-exporter-network-info failed: no value found, so use initial net info
    [INFO]     2023/12/08 07:39:01.744739 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:01.852413 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:05.055247 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:06.688352 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:06.750876 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:09.843914 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    [INFO]     2023/12/08 07:39:11.688505 157     collector/npu_collector.go:476    update cache,key is npu-exporter-network-info
    [INFO]     2023/12/08 07:39:11.701081 158     collector/npu_collector.go:499    update cache,key is npu-exporter-containers-devices
    [INFO]     2023/12/08 07:39:14.859243 148     collector/npu_collector.go:442    update cache,key is npu-exporter-npu-list
    ...
    ```

## Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426319"></a>

请在任意节点执行以下步骤验证Ascend Device Plugin的安装状态。

**操作步骤<a name="zh-cn_topic_0000001497205413_section197491249115016"></a>**

1.  通过如下命令查看K8s集群中Ascend Device Plugin的Pod，需要满足Pod的“STATUS”为Running，READY为1/1。如果集群中有多个节点安装了Ascend Device Plugin，每一个节点都需要确认。

    ```
    kubectl get pods -n kube-system -o wide | grep device-plugin
    ```

    回显示例：

    ```
    ascend-device-plugin-daemonset-910-85p9v   1/1     Running   0          19h     192.168.185.251   ubuntu       <none>           <none>
    ```

2.  通过如下命令查看K8s集群中Ascend Device Plugin的日志。

    ```
    kubectl logs -n kube-system Ascend Device Plugin组件的Pod名字
    ```

    回显示例如下，表示组件正常。

    ```
    root@ubuntu:~# kubectl logs -n kube-system ascend-device-plugin-daemonset-910-85p9v 
    [INFO]     2022/11/21 11:20:04.534992 1       hwlog@v0.0.0/api.go:96    devicePlugin.log's logger init success
    [INFO]     2022/11/21 11:20:04.535750 1       main.go:127    ascend device plugin starting and the version is xxx_linux-x86_64
    [INFO]     2022/11/21 11:20:05.992823 1       K8stool@v0.0.0/self_K8s_client.go:116    start to decrypt cfg
    [INFO]     2022/11/21 11:20:06.002773 1       K8stool@v0.0.0/self_K8s_client.go:125    Config loaded from file: ****tc/mindx-dl/device-plugin/.config/config6
    [INFO]     2022/11/21 11:20:06.003751 1       main.go:153    init kube client success 
    [INFO]     2022/11/21 11:20:06.003923 1       device/ascendcommon.go:104    Found Huawei Ascend, deviceType: Ascend910, deviceName: Ascend910-4
    [INFO]     2022/11/21 11:20:06.003970 1       main.go:160    init device manager success
    [INFO]     2022/11/21 11:20:06.004157 21      device/manager.go:125    starting the listen device
    [INFO]     2022/11/21 11:20:06.004285 7       device/manager.go:206    Serve start
    [INFO]     2022/11/21 11:20:06.004970 7       server/server.go:88    device plugin (Ascend910) start serving.
    [INFO]     2022/11/21 11:20:06.007285 7       server/server.go:36    register Ascend910 to kubelet success.
    [INFO]     2022/11/21 11:20:06.007521 7       server/pod_resource.go:44    pod resource client init success.
    [INFO]     2022/11/21 11:20:06.007755 35      server/plugin.go:87    ListAndWatch resp devices: Ascend910-4 Healthy# 上报K8s的芯片，请以实际为准
    [INFO]     2022/11/21 11:20:11.063218 21      kubeclient/client_server.go:123    reset annotation success
    ...
    ```

3.  通过如下命令查看K8s中节点的详细情况。如果节点详情中的“Capacity”字段和“Allocatable”字段出现了昇腾AI处理器的相关信息，表示Ascend Device Plugin给K8s上报芯片正常，组件运行正常。

    ```
    kubectl describe node K8s中的节点名
    ```

    -   以Atlas 800 训练服务器为例，回显示例如下：

        ```
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend910
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                      72
          ephemeral-storage:        479567536Ki
          huawei.com/Ascend910:     8# K8s已感知到该节点总共有8个NPU
        ...
        Allocatable:
          cpu:                      72
          ephemeral-storage:        441969440446
          huawei.com/Ascend910:     8  # K8s已感知到该节点可供分配的NPU总个数为8
        ...
        ```

    -   以服务器（插Atlas 300I 推理卡）为例，回显示例如下，节点上芯片个数请以实际为准。

        ```
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend310
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                       72
          ephemeral-storage:         163760Mi
          huawei.com/Ascend310:      4
        ...
        Allocatable:
          cpu:                       72
          ephemeral-storage:         154543324929
          huawei.com/Ascend310:      4
        ...
        ```

    -   以服务器（插Atlas 300I Pro 推理卡）为例。非混插模式，节点包含Atlas 推理系列产品，回显示例如下，节点上芯片个数请以实际为准。

        ```
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend310
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                      96
          ephemeral-storage:        95596964Ki
          huawei.com/Ascend310P:    3
        ...
        Allocatable:
          cpu:                      96
          ephemeral-storage:        88102161877
          huawei.com/Ascend310P:    3
        ...
        ```

    -   以服务器（插Atlas 300I Pro 推理卡）为例。混插模式，节点包含Atlas 推理系列产品，回显示例如下，节点上芯片个数请以实际为准。

        ```
        root@ubuntu:~# kubectl describe node ubuntu
        Name:               ubuntu
        Roles:              worker
        Labels:             accelerator=huawei-Ascend310
                            beta.kubernetes.io/arch=amd64
        ...
        CreationTimestamp:  Wed, 22 Dec 2021 20:10:04 +0800
        Taints:             <none>
        Unschedulable:      false
        ...
        Capacity:
          cpu:                      96
          ephemeral-storage:        95596964Ki
          huawei.com/Ascend310P-IPro:    1
          huawei.com/Ascend310P-V:       1
          huawei.com/Ascend310P-VPro:    1
        ...
        Allocatable:
          cpu:                      96
          ephemeral-storage:        88102161877
          huawei.com/Ascend310P-IPro:    1
          huawei.com/Ascend310P-V:       1
          huawei.com/Ascend310P-VPro:    1
        ...
        ```

## Volcano<a name="ZH-CN_TOPIC_0000002511346325"></a>

1.  通过如下命令查看K8s集群中Volcano的两个Pod，需要满足Pod的STATUS都为Running，READY都为1/1。

    ```
    kubectl get pods -n volcano-system -o wide | grep volcano
    ```

    回显示例：

    ```
    volcano-controllers-758b6d8bdd-b7g89   1/1     Running   2          166m   192.168.102.69   ubuntu       <none>           <none>
    volcano-scheduler-86775f88f-w649w      1/1     Running   2          166m   192.168.102.91   ubuntu       <none>           <none>
    ```

2.  登录Volcano Pod运行的节点，使用如下命令查看Volcano组件日志。
    -   查看volcano-controllers的日志。

        ```
        cat /var/log/mindx-dl/volcano-controller/volcano-controller.log
        ```

        回显示例如下，表示组件正常运行。

        ```
        Log file created at: 2022/10/14 11:22:32
        Running on machine: volcano-controllers-758b6d8bdd-wc49r
        Binary: Built with gc go1.17.8-htrunk4 for linux/arm64
        Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
        I1014 11:22:32.070656       1 garbagecollector.go:91] Starting garbage collector
        I1014 11:22:32.072772       1 queue_controller.go:171] Starting queue controller.
        I1014 11:22:32.652887       1 queue_controller.go:238] Begin execute SyncQueue action for queue default, current status
        I1014 11:22:32.653026       1 queue_controller_action.go:36] Begin to sync queue default.
        I1014 11:22:32.756216       1 queue_controller_action.go:82] End sync queue default.
        I1014 11:22:32.756254       1 queue_controller.go:220] Finished syncing queue default (103.399375ms).
        I1014 11:22:32.972001       1 pg_controller.go:109] PodgroupController is running ......
        I1014 11:22:32.972396       1 job_controller.go:252] JobController is running ......
        I1014 11:22:32.972423       1 job_controller.go:256] worker 1 start ......
        I1014 11:22:32.972426       1 job_controller.go:256] worker 0 start ......
        I1014 11:22:32.972426       1 job_controller.go:256] worker 2 start ......
        ...
        ```

    -   查看volcano-scheduler的日志。

        ```
        cat /var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log
        ```

        回显示例如下，表示组件运行正常。

        ```
        Log file created at: 2022/10/14 11:22:32
        Running on machine: volcano-scheduler-86775f88f-6dtqf
        Binary: Built with gc go1.17.8-htrunk4 for linux/arm64
        Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
        ...
        ```

## ClusterD<a name="ZH-CN_TOPIC_0000002479386380"></a>

请在任意节点执行以下步骤验证ClusterD的安装状态。

1.  通过如下命令查看K8s集群中ClusterD的Pod，需要满足Pod的数量为1，STATUS为Running，READY为1/1。

    ```
    kubectl get pods -n mindx-dl -o wide | grep clusterd
    ```

    回显示例：

    ```
    clusterd-7844cb867d-fwcj7   1/1     Running   0          2m14s   <none>   node133   <none>           <none>
    ```

2.  执行以下命令，查询ClusterD的Pod日志。

    ```
    kubectl logs -f -n mindx-dl {ClusterD组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```
    [INFO]     2024/07/24 13:58:30.602051 CST 1       hwlog@v0.10.12/api.go:105    cluster-info.log's logger init success
    W0724 13:58:30.602197       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.603416 CST 1       grpc/grpc_init.go:57    cluster info server start listen
    ...
    W0724 13:58:30.621433       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    [INFO]     2024/07/24 13:58:30.621911 CST 258     job/factory.go:172    delete job summary cm goroutine started
    ```

## Ascend Operator<a name="ZH-CN_TOPIC_0000002479386462"></a>

请在任意节点执行以下步骤验证Ascend Operator的安装状态。

1.  通过如下命令查看K8s集群中Ascend Operator的Pod，需要满足Pod的STATUS为Running，READY为1/1。

    ```
    kubectl get pods -n mindx-dl -o wide | grep ascend-operator
    ```

    回显示例：

    ```
    ascend-operator-manager-b59774f7-8l5gn         1/1     Running   0          6m52s   192.168.102.67   ubuntu       <none>           <none>
    ```

2.  通过如下命令查看K8s集群中Ascend Operator的日志。

    ```
    kubectl logs -n mindx-dl {Ascend Operator组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```
    root@ubuntu:~# kubectl logs -n mindx-dl ascend-operator-manager-b59774f7-8l5gn 
    [INFO]     2023/03/20 17:48:34.308373 1       hwlog/api.go:108    ascend-operator.log's logger init success
    [INFO]     2023/03/20 17:48:34.308469 1       ascend-operator/main.go:86    ascend-operator starting and the version is xxx
    [INFO]     2023/03/20 17:48:34.964296 1       ascend-operator/main.go:101    starting manager
    ...
    ```

## NodeD<a name="ZH-CN_TOPIC_0000002479386440"></a>

请在任意节点执行以下步骤验证NodeD的安装状态。

1.  通过如下命令查看K8s集群中NodeD的Pod，需要满足Pod的STATUS为Running，READY为1/1。如果集群中有多个节点安装了NodeD，每个节点都需要确认。

    ```
    kubectl get pods -n mindx-dl -o wide | grep noded
    ```

    回显示例：

    ```
    noded-bnmwt                        1/1     Running   10         40d    192.168.41.28     ubuntu       <none>           <none>
    ```

2.  通过如下命令查看NodeD组件日志。

    ```
    kubectl logs -n mindx-dl {NodeD组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```
    [INFO] 2025/05/25 15:24:19.897280 1 hwlog/api.go:108 noded.log's logger init success
    [INFO] 2025/05/25 15:24:19.897392 1 noded/main.go:93 noded starting and the version is v7.3.0_linux-x86_64
    W0525 15:24:19.897410 1 client_config.go:617] Neither --kubeconfig nor --master was specified. Using the inClusterConfig. This might not work.
    [INFO] 2025/05/25 15:24:19.994306 1 devmanager/devmanager.go:123 the dcmi version is 24.1.rc3.b060
    [INFO] 2025/05/25 15:24:19.994360 1 devmanager/devmanager.go:1071 get chip base info, cardID: 0, deviceID: 0, logicID: 0, physicID: 0
    [INFO] 2025/05/25 15:24:19.994386 1 devmanager/devmanager.go:1071 get chip base info, cardID: 1, deviceID: 0, logicID: 1, physicID: 1
    [INFO] 2025/05/25 15:24:19.994408 1 devmanager/devmanager.go:1071 get chip base info, cardID: 2, deviceID: 0, logicID: 2, physicID: 2
    [INFO] 2025/05/25 15:24:19.994430 1 devmanager/devmanager.go:1071 get chip base info, cardID: 3, deviceID: 0, logicID: 3, physicID: 3
    [INFO] 2025/05/25 15:24:19.994449 1 devmanager/devmanager.go:1071 get chip base info, cardID: 4, deviceID: 0, logicID: 4, physicID: 4
    [INFO] 2025/05/25 15:24:19.994476 1 devmanager/devmanager.go:1071 get chip base info, cardID: 5, deviceID: 0, logicID: 5, physicID: 5
    [INFO] 2025/05/25 15:24:19.994505 1 devmanager/devmanager.go:1071 get chip base info, cardID: 6, deviceID: 0, logicID: 6, physicID: 6
    [INFO] 2025/05/25 15:24:19.994528 1 devmanager/devmanager.go:1071 get chip base info, cardID: 7, deviceID: 0, logicID: 7, physicID: 7
    [WARN] 2025/05/25 15:24:19.994564 1 executor/dev_manager.go:71 deviceManager get hccsPingMeshState failed, err: dcmi get hccs ping mesh state failed cardID(0) deviceID(0) error code: -99998
    [ERROR] 2025/05/25 15:24:19.994588 1 pingmesh/controller.go:68 new device manager failed, err: dcmi get hccs ping mesh state failed cardID(0) deviceID(0) error code: -99998
    [INFO] 2025/05/25 15:24:19.999314 1 config/configurator.go:98 update fault config success
    [INFO] 2025/05/25 15:24:19.999350 1 config/configurator.go:231 init fault config from config map success
    [INFO] 2025/05/25 15:24:39.037815 1 control/controller.go:220 get node SN success, add SN(HS20200764) to node annotation
    ...
    ```

## Resilience Controller<a name="ZH-CN_TOPIC_0000002511426295"></a>

请在任意节点执行以下步骤验证Resilience Controller的安装状态。

1.  通过如下命令查看K8s集群中Resilience Controller的Pod，需要满足Pod的STATUS为Running，READY为1/1。

    ```
    kubectl get pods -n mindx-dl -o wide | grep resilience-controller
    ```

    回显示例：

    ```
    resilience-controller-76f4476bb5-fs986         1/1     Running   0          6m52s   192.168.102.67   ubuntu       <none>           <none>
    ```

2.  通过如下命令查看K8s集群中Resilience Controller的日志。

    ```
    kubectl logs -n mindx-dl {Resilience组件的Pod名字}
    ```

    回显示例如下，表示组件正常运行。

    ```
    root@ubuntu:~# kubectl logs -n mindx-dl resilience-controller-76f4476bb5-fs986 
    [INFO]     2022/11/17 17:18:46.697010 1       hwlog@v0.0.0/api.go:96    run.log's logger init success
    [INFO]     2022/11/17 17:18:46.697139 1       cmd/main.go:57    resilience-controller starting and the version is xxx_linux-x86_64
    [INFO]     2022/11/17 17:18:47.227913 1       K8stool@v0.0.0/self_K8s_client.go:116    start to decrypt cfg
    [INFO]     2022/11/17 17:18:47.297559 1       K8stool@v0.0.0/self_K8s_client.go:125    Config loaded from file: ****tc/mindx-dl/resilience-controller/.config/config6
    [INFO]     2022/11/17 17:18:47.300066 1       elastic/controller.go:45    Setting up elastic event handlers
    [INFO]     2022/11/17 17:18:47.300179 1       elastic/controller.go:63    Starting elastic controller, waiting for informer caches to sync
    [INFO]     2022/11/17 17:18:47.401246 1       cmd/main.go:80    elastic controller started
    ...
    ```

## Container Manager<a name="ZH-CN_TOPIC_0000002492269056"></a>

请在Container Manager组件部署的节点上执行以下步骤验证Container Manager组件的安装状态。

1.  查看组件服务的状态，需要满足组件状态为active \(running\)。

    ```
    systemctl status container-manager.service
    ```

    回显示例：

    ```
    ● container-manager.service - Ascend container manager
         Loaded: loaded (/etc/systemd/system/container-manager.service; disabled; vendor preset: enabled)
         Active: active (running) since Wed 2025-11-26 20:56:50 UTC; 16s ago
        Process: 41459 ExecStart=/bin/bash -c container-manager run  -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
       Main PID: 41464 (container-manag)
          Tasks: 10 (limit: 629145)
         Memory: 13.3M
         CGroup: /system.slice/container-manager.service
                 └─41464 /home/container-manager/container-manager run -ctrStrategy ringRecover
    ...
    ```

    >[!NOTE] 说明 
    >若回显中出现类似如下信息，可忽略，不影响实际功能，可能原因是未配置RoCE网卡IP地址和子网掩码。若不想打印该信息，可参见《Atlas A2 中心推理和训练硬件 25.5.0 HCCN Tool 接口参考》的“[配置功能\>配置RoCE网卡IP地址和子网掩码](https://support.huawei.com/enterprise/zh/doc/EDOC1100540101/44299f2a)”章节配置。
    >```
    >[dsmi_common_interface.c:1017][ascend][curpid:244135,244135][drv][dmp][dsmi_get_device_ip_address]devid 0 dsmi_cmd_get_device_ip_address return 1 error!
    >```

2.  查看组件日志。

    ```
    cat /var/log/mindx-dl/container-manager/container-manager.log
    ```

    回显以Atlas 800I A3 超节点服务器为例：

    ```
    [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
    [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
    [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
    [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
    [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
    [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
    ...
    ```

    如果出现如下打印信息，表示组件运行正常。

    ```
    ...
    [INFO]     2025/11/25 22:46:59.289352 1       devmgr/workflow.go:57    init module <hwDev manager> success
    [INFO]     2025/11/25 22:46:59.293773 1       app/config.go:40    load fault config from faultCode.json success
    [INFO]     2025/11/25 22:46:59.293866 1       app/workflow.go:50    init module <fault manager> success
    [INFO]     2025/11/25 22:46:59.293901 1       app/workflow.go:76    init module <container controller> success
    [INFO]     2025/11/25 22:46:59.293930 1       app/workflow.go:64    init module <reset-manager> success
    [INFO]     2025/11/25 22:46:59.315101 378     devmgr/hwdevmgr.go:365    subscribe device fault event success
    ...
    ```

# 升级<a name="ZH-CN_TOPIC_0000002479226452"></a>







## 升级说明<a name="ZH-CN_TOPIC_0000002511346381"></a>

本章节旨在指导用户将MindCluster集群调度组件升级到新版本。MindCluster集群调度组件的升级支持以下2种方式。

-   全量升级：此种升级方式不仅会升级各组件的二进制镜像文件，而且升级后可对组件的配置文件进行修改。此种升级方式支持跨版本升级，例如，用户可从5.0.x版本升级到7.0.x版本。
-   升级镜像：此种升级方式仅升级各组件的二进制文件，不支持修改权限、启动参数等，无需进行升级前环境检查。此种升级方式仅支持在同一个版本内进行升级。

    **表 1**  升级方式说明

    <a name="table1527494117524"></a>
    <table><thead align="left"><tr id="row327404115216"><th class="cellrowborder" valign="top" width="17.5%" id="mcps1.2.5.1.1"><p id="p627494165216"><a name="p627494165216"></a><a name="p627494165216"></a>升级方式</p>
    </th>
    <th class="cellrowborder" valign="top" width="25.990000000000002%" id="mcps1.2.5.1.2"><p id="p92749419529"><a name="p92749419529"></a><a name="p92749419529"></a>是否支持跨版本升级</p>
    </th>
    <th class="cellrowborder" valign="top" width="30.240000000000002%" id="mcps1.2.5.1.3"><p id="p19274134120522"><a name="p19274134120522"></a><a name="p19274134120522"></a>是否需要停止训练/推理任务</p>
    </th>
    <th class="cellrowborder" valign="top" width="26.27%" id="mcps1.2.5.1.4"><p id="p15533184405419"><a name="p15533184405419"></a><a name="p15533184405419"></a>参考章节</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1727434112526"><td class="cellrowborder" valign="top" width="17.5%" headers="mcps1.2.5.1.1 "><p id="p1027414185220"><a name="p1027414185220"></a><a name="p1027414185220"></a>全量升级</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.990000000000002%" headers="mcps1.2.5.1.2 "><p id="p3274841105220"><a name="p3274841105220"></a><a name="p3274841105220"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.240000000000002%" headers="mcps1.2.5.1.3 "><p id="p927454111524"><a name="p927454111524"></a><a name="p927454111524"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="26.27%" headers="mcps1.2.5.1.4 "><p id="p6533944195419"><a name="p6533944195419"></a><a name="p6533944195419"></a><a href="#升级说明">升级说明</a>-<a href="#升级其他组件">升级其他组件</a>章节</p>
    </td>
    </tr>
    <tr id="row8274241115212"><td class="cellrowborder" valign="top" width="17.5%" headers="mcps1.2.5.1.1 "><p id="p202747416524"><a name="p202747416524"></a><a name="p202747416524"></a>升级镜像</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.990000000000002%" headers="mcps1.2.5.1.2 "><p id="p1327413412527"><a name="p1327413412527"></a><a name="p1327413412527"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.240000000000002%" headers="mcps1.2.5.1.3 "><p id="p3274144175214"><a name="p3274144175214"></a><a name="p3274144175214"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="26.27%" headers="mcps1.2.5.1.4 "><p id="p25334441543"><a name="p25334441543"></a><a name="p25334441543"></a><a href="#升级镜像">升级镜像</a>章节</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 说明 
    >本章节不适用的场景：用户对旧版本MindCluster集群调度组件的源代码（不含配置文件）进行了修改，请分析版本代码差异后再进行升级。

**升级环境检查<a name="section19242859587"></a>**

在进行各组件的升级步骤前，请根据实际安装场景，选择相应的组件进行检查。

1.  检查是否有正在运行的任务。若用户正在执行的任务，请等待任务执行完成或提前停止任务后，再升级MindCluster组件。
    1.  请执行以下命令检查是否有正在运行的任务。

        ```
        kubectl get pods -A
        ```

        回显示例如下。

        ```
        NAMESPACE        NAME                                       READY   STATUS    RESTARTS         AGE
        default          ubuntu-pod                                 1/1     Running   32 (118m ago)    3d18h ...  
        ```

    2.  进入任务YAML所在路径，执行以下命令停止任务。

        ```
        kubectl delete -f  xxx.yaml              # xxx表示任务YAML的名称，请根据实际情况填写    
        ```

2.  （可选）检查pingmesh灵衢网络检测开关是否已关闭。
    1.  登录环境，进入NodeD解压目录。
    2.  执行以下命令编辑pingmesh-config文件。

        ```
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

        如果回显如下所示，表示pingmesh灵衢网络检测开关已关闭。无需执行[步骤3](#li1427143773119)。

        ```
        Error from server (NotFound): configmaps "pingmesh-config" not found
        ```

    3.  <a name="li1427143773119"></a>（可选）修改activate字段的取值。
        -   如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        -   如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            -   在配置文件中新增该超节点信息，并将activate为off。
            -   删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

3.  检查已安装的MindCluster组件。
    -   （可选）**检查TaskD组件**。执行以下命令进入容器内部，查看TaskD组件安装状态。

        ```
        docker run -it  {训练镜像名称}:tag /bin/bash
        pip show taskd
        ```

        回显如下，表示镜像中已安装TaskD组件。

        ```
        Name: taskd
        Version: x.x.x
        Summary: Ascend MindCluster taskd is a new library for training management
        Home-page: UNKNOWN
        Author: 
        Author-email: 
        License: UNKNOWN
        Location: /usr/local/python3/lib/python3.10/site-packages
        Requires: grpcio, protobuf, pyOpenSSL, torch, torch-npu
        Required-by:
        ```

    -   （可选）**检查其他组件**。参考[组件状态确认](#组件状态确认)，确认集群中节点是否安装了相应组件。

4.  （可选）若尚未安装MindCluster集群调度组件，请参考[安装部署](#安装部署)章节先安装组件，TaskD的安装步骤请参考[制作镜像](./usage/resumable_training.md#制作镜像)章节。

## 升级Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002479226420"></a>

仅Ascend Docker Runtime支持通过命令行进行升级，其他集群调度组件可通过卸载后重新安装进行升级。

目前只支持root用户升级Ascend Docker Runtime。

**前提条件<a name="section176591058124515"></a>**

已完成[升级环境检查](#升级说明)。

**升级步骤<a name="section520182224617"></a>**

1.  下载新版本组件安装包，详情请参见参考[获取软件包](#获取软件包)章节。
2.  <a name="li12599722163212"></a>进入安装包（run包）所在路径，在该路径下执行以下命令为软件包添加可执行权限。

    ```
    cd <path to run package>
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3.  通过以下命令升级Ascend Docker Runtime。
    -   （可选）在默认路径下升级Ascend Docker Runtime，需要依次执行以下命令。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --upgrade
        ```

    -   （可选）在指定路径下升级Ascend Docker Runtime，需要依次执行以下命令。“--install-path”参数为指定的升级路径。

        ```
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --upgrade --install-path=<path>
        ```

        回显示例如下，表示升级成功。

        ```
        Uncompressing ascend-docker-runtime  100%
        ...
        [INFO] ascend-docker-runtime upgrade success
        ```

4.  （可选）执行以下命令重启容器，使新版Ascend Docker Runtime生效。如不涉及安装路径、安装参数变更，可跳过本步骤。
    -   Docker场景（或K8s集成Docker场景）

        ```
        systemctl daemon-reload && systemctl restart docker
        ```

    -   Containerd场景（或K8s集成Containerd场景）

        ```
        systemctl daemon-reload && systemctl restart containerd
        ```

5.  <a name="li76002022113215"></a>参考[组件状态确认](#组件状态确认)章节，检查新版本Ascend Docker Runtime是否升级成功状态。
6.  （可选）恢复旧版本。下载旧版本安装包，依次重新执行[步骤2](#li12599722163212)到[步骤5](#li76002022113215)。

## 升级TaskD<a name="ZH-CN_TOPIC_0000002479226444"></a>

TaskD组件安装在训练镜像内部，在训练镜像内部重新安装该whl包即可完成升级。

**前提条件<a name="section18616132394915"></a>**

已完成[升级环境检查](#升级说明)。

**升级步骤<a name="section1720814439492"></a>**

1.  参考[获取软件包](#获取软件包)章节，下载新版本组件安装包。
2.  下载完成后，进入安装包所在路径并解压安装包。
3.  执行**ls -l**命令，回显示例如下。

    ```
    -rw-r--r-- 1 root root 1493228 Mar 14 02:09 Ascend-mindxdl-taskd_{version}_linux-aarch64.zip
    -r-------- 1 root root 1506842 Mar 12 18:07 taskd-{version}-py3-none-linux_aarch64.whl
    ```

4.  基于已有的训练镜像，安装新版本TaskD组件。
    1.  执行以下命令运行训练镜像。

        ```
        docker run -it  -v /host/packagepath:/container/packagepath training_image:latest bash
        ```

    2.  执行以下命令卸载已安装的TaskD组件。

        ```
        pip uninstall taskd -y
        ```

        回显示例如下表示卸载成功。

        ```
        Successfully uninstalled taskd-{version}
        ```

    3.  执行以下命令安装新版本TaskD。

        ```
        pip install taskd-{version}-py3-none-linux_aarch64.whl
        ```

        回显如下。

        ```
        Successfully installed taskd-{version}
        ```

    4.  安装了新版本TaskD后，将容器保存为新镜像。

        ```
        docker ps
        ```

        回显示例如下。

        ```
        CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS     NAMES
        8b70390775f2   fd6acb527bad           "/bin/bash -c 'sleep…"   2 hours ago    Up 2 hours              k8s_ascend_default-last-test-deepseek2-60b
        ```

        将该容器提交为新版本训练容器镜像，注意新镜像的tag与旧镜像不一致。示例如下。

        ```
        docker commit 8b70390775f2 newimage:latest
        ```

5.  检查新版TaskD是否升级完成，参考[检查TaskD组件](#升级说明)章节，检查组件状态是否正常。
6.  （可选）回退老版本。若旧版镜像仍然存在，无需回退操作；若不存在则按上述步骤，重新安装旧版本TaskD软件包即可。

## 升级Container Manager<a name="ZH-CN_TOPIC_0000002524548731"></a>

在物理机上直接替换Container Manager二进制升级组件。

1.  以root用户登录Container Manager组件部署的节点。
2.  将获取到的Container Manager软件包上传至服务器的任意目录（如“/tmp/container-manager”）。
3.  进入“/tmp/container-manager”目录并进行解压操作。

    ```
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

    >[!NOTE] 说明 
    ><i><version\></i>为软件包的版本号；<i><arch\></i>为CPU架构。

4.  依次执行以下命令，升级Container Manager组件。

    ```
    # 停止Container Manager系统服务，并删除对应Container Manager二进制文件
    systemctl stop container-manager.service
    chattr -i /usr/local/bin/container-manager
    rm -f /usr/local/bin/container-manager
    
    # 从解压文件中获取新二进制文件，替换旧Container Manager二进制文件
    cp /tmp/container-manager/container-manager /usr/local/bin
    chmod 500 /usr/local/bin/container-manager
    
    # 重启Container Manager系统服务
    systemctl daemon-reload
    systemctl start container-manager.service
    ```

5.  验证Container Manager组件的升级状态。
    1.  查看组件服务的状态，需要满足组件状态为active \(running\)。

        ```
        systemctl status container-manager.service
        ```

        回显示例：

        ```
        ● container-manager.service - Ascend container manager
             Loaded: loaded (/etc/systemd/system/container-manager.service; disabled; vendor preset: enabled)
             Active: active (running) since Wed 2025-11-26 20:56:50 UTC; 16s ago
            Process: 41459 ExecStart=/bin/bash -c container-manager run  -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
           Main PID: 41464 (container-manag)
              Tasks: 10 (limit: 629145)
             Memory: 13.3M
             CGroup: /system.slice/container-manager.service
                     └─41464 /home/container-manager/container-manager run -ctrStrategy ringRecover
        ...
        ```

    2.  查看组件日志。

        ```
        cat /var/log/mindx-dl/container-manager/container-manager.log
        ```

        回显以Atlas 800I A3 超节点服务器为例：

        ```
        [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
        [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
        [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
        [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
        [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
        [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
        ...
        ```

        如果出现如下打印信息，表示组件运行正常。

        ```
        ...
        [INFO]     2025/11/25 22:46:59.289352 1       devmgr/workflow.go:57    init module <hwDev manager> success
        [INFO]     2025/11/25 22:46:59.293773 1       app/config.go:40    load fault config from /home/faultCode.json success
        [INFO]     2025/11/25 22:46:59.293866 1       app/workflow.go:50    init module <fault manager> success
        [INFO]     2025/11/25 22:46:59.293901 1       app/workflow.go:76    init module <container controller> success
        [INFO]     2025/11/25 22:46:59.293930 1       app/workflow.go:64    init module <reset-manager> success
        [INFO]     2025/11/25 22:46:59.315101 378     devmgr/hwdevmgr.go:365    subscribe device fault event success
        ...
        ```

## 升级其他组件<a name="ZH-CN_TOPIC_0000002511346401"></a>

**前提条件<a name="section176591058124515"></a>**

-   已完成[升级环境检查](#升级说明)。

-   如需升级NPU Exporter、Ascend Device Plugin、Volcano、ClusterD、Ascend Operator、NodeD和Resilience Controller组件，需卸载旧版本后，再执行新版本的安装步骤。

**升级步骤<a name="section65996266718"></a>**

1.  卸载MindCluster旧版本组件。详情请参见[卸载其他组件](#卸载)中"卸载组件"步骤。
2.  参考[获取软件包](#获取软件包)章节，下载新版本组件安装包。
3.  （可选）准备MindCluster集群调度组件新版本镜像。若新版本组件采用二进制方式安装，可跳过本步骤。

    参考[准备镜像](#准备镜像)章节，从昇腾镜像仓库拉取新版本镜像或者制作新版本镜像。注意新版本组件镜像tag要与旧版本组件镜像tag不一致，避免覆盖旧版本组件镜像。

4.  <a name="li147194506333"></a>请根据要升级的组件，重新执行手动安装步骤。详细步骤请参见[安装MindCluster新版本组件](#npu-exporter)。
5.  （可选）如需回退老版本，依次执行[卸载](#卸载)中"卸载其他组件"中卸载组件步骤和[步骤4](#li147194506333)，卸载新版本组件后安装旧版本组件即可。

## 升级镜像<a name="ZH-CN_TOPIC_0000002511346311"></a>

本章节仅指导用户在同一个版本内对容器镜像中二进制文件版本进行升级，升级过程中不会修改权限及启动参数。如需了解关于升级方式的更详细说明，请参见[升级说明](#升级说明)。

-   如需升级Volcano、ClusterD、Ascend Operator、Resilience Controller组件的镜像，可参考[升级管理节点组件](#section1292111716589)。
-   如需升级NPU Exporter、Ascend Device Plugin和NodeD组件镜像，可参考[升级计算节点组件](#section231311416588)。
-   TaskD暂不支持此种升级方式。

**升级管理节点组件<a name="section1292111716589"></a>**

1.  参考[准备镜像](#准备镜像)章节，使用新的软件包制作镜像。

    >[!NOTE] 说明
    >请保持镜像名称一致，否则可能导致原配置文件无法拉起Pod。

2.  执行以下命令，查询旧版本Deployment配置。

    ```
    kubectl get deployment -A|grep {组件名称}
    ```

    以ClusterD组件为例，回显示例如下。

    ```
    mindx-dl         clusterd        1/1     1      1       45h
    ```

3.  执行以下命令，重启Deployment。

    ```
    kubectl rollout restart deployment -n {命名空间名称} {deployment名称}
    ```

    以ClusterD组件为例，回显示例如下。

    ```
    deployment.apps/clusterd restarted
    ```

4.  检查新版本Pod是否已拉起。

    ```
    kubectl get pod -A|grep {组件名称}
    ```

    以ClusterD为例，回显示例如下，表示Pod成功拉起。

    ```
    mindx-dl   clusterd-99f8795c8-drqb4  1/1  Running 0       1m
    ```

**升级计算节点组件<a name="section231311416588"></a>**

1.  参考[准备镜像](#准备镜像)章节，使用新的软件包制作镜像。

    >[!NOTE] 说明 
    >请保持镜像名称一致，否则可能导致原配置文件无法拉起Pod。

2.  执行以下命令，查询旧版本DaemonSet配置。

    ```
    kubectl get ds -A|grep {组件名称}
    ```

    以NodeD组件为例，回显示例如下。

    ```
    mindx-dl         noded        1/1     1      1       45h
    ```

3.  执行以下命令，重启DaemonSet。

    ```
    kubectl rollout restart ds -n {命名空间名称} {ds名称}
    ```

    以NodeD组件为例，回显示例如下。

    ```
    daemonsets.apps/noded restarted
    ```

4.  检查新版本Pod是否已拉起。

    ```
    kubectl get pod -A|grep {组件名称}
    ```

    以NodeD为例，回显示例如下，表示Pod已拉起。

    ```
    mindx-dl   noded- m4j4r  1/1  Running 0     1m
    ```

# 卸载<a name="ZH-CN_TOPIC_0000002511426389"></a>

-   卸载Ascend Docker Runtime组件，请参见[卸载Ascend Docker Runtime](#section6134163311244)进行操作。
-   卸载Container Manager组件，请参见[卸载Container Manager组件](#section1461059103619)进行操作。
-   卸载NPU Exporter、Ascend Device Plugin、Volcano、ClusterD、Ascend Operator、NodeD和Resilience Controller，请参见[卸载其他组件](#section6361146202520)。

**卸载Ascend Docker Runtime<a name="section6134163311244"></a>**

-   情况一：使用不同安装路径。

    用户在卸载Ascend Docker Runtime时需要针对不同容器引擎，根据[步骤2](#li345320287225)进行两次卸载操作，每次卸载需要指定相应的安装路径，即--install-path参数。

-   情况二：使用相同安装路径。

    用户在卸载Ascend Docker Runtime时，只需根据[步骤2](#li345320287225)进行一次卸载操作。卸载完成之后需要手动将另一引擎的daemon.json文件还原为Ascend Docker Runtime安装之前的内容。

若用户需要保留其中一个容器引擎，需要在Ascend Docker Runtime卸载之后，针对相应场景进行重新安装。

1.  （可选）关闭pingmesh灵衢网络检测。
    1.  登录环境，进入NodeD解压目录。
    2.  执行以下命令编辑pingmesh-config文件。

        ```
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

    3.  修改activate字段的取值。
        -   如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        -   如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            -   在配置文件中新增该超节点信息，并将activate为off。
            -   删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

2.  <a name="li345320287225"></a>可以选择以下方式中的一种卸载Ascend Docker Runtime软件。
    -   方式一：（推荐）使用软件包卸载
        1.  首先进入安装包（run包）所在路径。

            ```
            cd <path to run package>
            ```

        2.  执行以下卸载命令，在**默认安装路径**下卸载Ascend Docker Runtime。

            -   Docker场景（或K8s集成Docker场景）

                ```
                ./Ascend-docker-runtime_{version}_linux-{arch}.run --uninstall
                ```

            -   Containerd场景（或K8s集成Containerd场景）

                ```
                ./Ascend-docker-runtime_{version}_linux-{arch}.run --uninstall --install-scene=containerd
                ```

            >[!NOTE] 说明 
            >-   Docker配置文件路径不是默认的“/etc/docker/daemon.json”时，需要新增“--config-file-path”参数，用于指定该配置文件路径。
            >-   Containerd的配置文件路径不是默认的“/etc/containerd/config.toml”时，需要新增“--config-file-path”参数，用于指定该配置文件路径。
            >-   如需要卸载指定安装路径下的Ascend Docker Runtime，需要在卸载命令中新增“--install-path=<path\>”参数。

            回显示例如下，表示卸载成功。

            ```
            Uncompressing ascend-docker-runtime  100%
            ...
            [INFO] ascend-docker-runtime uninstall success
            ```

    -   方式二：使用脚本卸载

        1.  首先进入Ascend Docker Runtime的安装路径下的“script”目录（默认安装路径为：“/usr/local/Ascend/Ascend-Docker-Runtime”）：

            ```
            cd /usr/local/Ascend/Ascend-Docker-Runtime/script
            ```

        2.  运行卸载的脚本进行卸载。

            -   Docker场景（或K8s集成Docker场景）

                ```
                uninstall.sh docker docker <daemon.json文件路径>
                ```

            -   Containerd场景（或K8s集成Containerd场景）

                ```
                uninstall.sh containerd containerd <config.toml文件路径>
                ```

            >[!NOTE] 说明 
            >-   可以不指定Docker的配置文件daemon.json路径，不指定时默认使用“/etc/docker/daemon.json”。
            >-   可以不指定Containerd的配置文件config.toml路径，不指定时默认使用“/etc/containerd/config.toml”。

        回显示例如下，表示卸载成功。

        ```
        [INFO]: You will recover Docker's daemon
        ...
        [INFO] uninstall.sh exec success
        ```

3.  （可选）在K8s集成Containerd的场景下，如果需要还原修改的kubeadm-flags.env，请参见[K8s官方文档](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/change-runtime-containerd/)，还原配置文件kubeadm-flags.env。其他场景可跳过该步骤。
4.  重启服务。
    -   Docker场景（或K8s集成Docker场景）

        ```
        systemctl daemon-reload && systemctl restart docker
        ```

    -   Containerd场景（或K8s集成Containerd场景）

        ```
        systemctl daemon-reload && systemctl restart containerd
        ```

**卸载Container Manager组件<a name="section1461059103619"></a>**

1.  以root用户登录Container Manager组件部署的节点。
2.  依次执行以下命令，卸载Container Manager组件系统服务。

    ```
    # 停止Container Manager系统服务
    systemctl stop container-manager.timer
    systemctl disable container-manager.timer
    systemctl stop container-manager.service
    systemctl disable container-manager.service
    
    # 删除Container Manager系统服务
    rm -f /etc/systemd/system/container-manager.service
    rm -f /etc/systemd/system/container-manager.timer
    systemctl daemon-reload
    systemctl reset-failed
    
    # 删除对应Container Manager二进制文件
    chattr -i /usr/local/bin/container-manager
    rm -f /usr/local/bin/container-manager
    ```

3.  删除日志文件，请确认实际路径后再删除。

    ```
    rm -rf /var/log/mindx-dl/container-manager
    ```

**卸载其他组件<a name="section6361146202520"></a>**

支持卸载集群调度组件，用户可以卸载组件后重新安装最新版本组件。通过逐一卸载各组件，并删除对应的命名空间、日志目录、配置文件等，请根据安装方式选择对应的卸载方式。

1.  （可选）关闭pingmesh灵衢网络检测。
    1.  登录环境，进入NodeD解压目录。
    2.  执行以下命令编辑pingmesh-config文件。

        ```
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

    3.  修改activate字段的取值。
        -   如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        -   如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            -   在配置文件中新增该超节点信息，并将activate为off。
            -   删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

2.  卸载组件。根据组件的安装方式，选择以下对应的卸载方式。
    -   通过容器方式卸载。各组件卸载方法类似，均为进入该组件配置文件YAML所在目录，并执行删除操作实现，此操作需要在K8s的管理节点操作。以卸载Ascend Device Plugin为例说明，请用户自行完成其余组件卸载。

        1.  以root用户登录管理节点。
        2.  进入Ascend Device PluginYAML配置文件所在目录（如：“/home/ascend-device-plugin”）。

            ```
            cd /home/ascend-device-plugin
            ```

        3.  在Ascend Device Plugin组件安装环境下，执行以下命令，卸载Ascend Device Plugin。

            ```
            kubectl delete -f device-plugin-volcano-v{version}.yaml
            ```

            回显示例如下：

            ```
            serviceaccount "ascend-device-plugin-sa-910" deleted
            clusterrole.rbac.authorization.k8s.io "pods-node-ascend-device-plugin-role-910" deleted
            clusterrolebinding.rbac.authorization.k8s.io "pods-node-ascend-device-plugin-rolebinding-910" deleted
            deployment.apps "ascend-device-plugin-daemonset-910" deleted
            ```

        >[!NOTE] 说明 
        >Ascend Device Plugin配合Volcano使用时，会创建ConfigMap，执行如下命令进行删除。
        >```
        >kubectl delete cm mindx-dl-deviceinfo-<node-name> -n kube-system
        >```

    -   通过二进制方式卸载。以卸载NPU Exporter为例说明，请用户自行完成其余组件卸载。
        1.  以root用户登录组件部署的节点。
        2.  在NPU Exporter组件安装环境下，依次执行如下命令卸载NPU Exporter组件。

            ```
            systemctl stop npu-exporter.service
            systemctl disable npu-exporter.service
            chattr -i /etc/systemd/system/npu-exporter.service
            rm -f /etc/systemd/system/npu-exporter.service
            systemctl daemon-reload
            systemctl reset-failed
            chattr -i /usr/local/bin/npu-exporter
            rm -f /usr/local/bin/npu-exporter
            ```

3.  删除命名空间。NPU Exporter的命名空间npu-exporter和Volcano的命名空间volcano-system在卸载组件时就已经同步删除，用户可以跳过本步骤。

    执行如下命令，卸载安装集群调度组件时创建的namespace。删除namespace会删除该namespace下的所有资源，请确认后再执行。

    ```
    kubectl delete ns mindx-dl
    ```

    回显示例如下：

    ```
    namespace "mindx-dl" deleted
    ```

4.  删除日志文件。参考[创建日志目录](#创建日志目录)章节，在对应节点上删除集群调度组件的日志目录。以ClusterD为例，请确认后再删除。

    ```
    rm -rf /var/log/mindx-dl/clusterd
    ```

5.  （可选）卸载Resilience Controller时，若导入了证书和KubeConfig文件，则需要删除证书和KubeConfig文件，请确认后再删除。

    ```
    rm -rf /etc/mindx-dl/resilience-controller
    ```

# 使用TaskD替换Elastic Agent<a name="ZH-CN_TOPIC_0000002515202401"></a>

Elastic Agent组件将会在后续版本日落，本章节提供使用TaskD替换Elastic Agent的操作指导。

**前提条件<a name="section565512391204"></a>**

-   已完成升级环境检查。
-   训练镜像已安装Elastic Agent。

**操作步骤<a name="section1643711813"></a>**

1.  参考[获取软件包](#获取软件包)章节，下载新版本TaskD组件安装包。
2.  下载完成后，进入安装包所在路径并解压安装包。
3.  执行**ls -l**命令，回显示例如下。

    ```
    -rw-r--r-- 1 root root 6134726 Nov 10 10:32 Ascend-mindxdl-taskd_{version}_linux-aarch64.zip
    -r-------- 1 root root 6205642 Nov  5 23:38 taskd-{version}-py3-none-linux_aarch64.whl
    ```

4.  基于已有的训练镜像，卸载Elastic Agent并安装新版本TaskD。
    1.  运行训练镜像。示例如下：

        ```
        docker run -it  -v /host/packagepath:/container/packagepath training_image:latest /bin/bash
        ```

    2.  卸载已安装的Elastic Agent组件。

        ```
        pip uninstall mindx-elastic -y
        ```

        回显示例如下，表示卸载成功。

        ```
        Successfully uninstalled mindx_elastic-{version}
        ```

    3.  删除Elastic Agent使能代码。

        ```
        sed -i '/mindx_elastic.api/d' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

        （可选）执行以下命令，查看对应文件是否已经删除Elastic Agent嵌入代码。

        ```
        vi $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

    4.  安装新版本TaskD。

        ```
        pip install taskd-{version}-py3-none-linux_aarch64.whl
        ```

        回显示例如下，表示安装成功。

        ```
        Successfully installed taskd-{version}
        ```

        执行以下命令，使能TaskD。

        ```
        sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

    5.  安装了新版本TaskD后，将容器保存为新镜像。

        ```
        docker ps
        ```

        回显示例如下。

        ```
        CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS     NAMES 
        bb118ca00041    f76142d63d3a           "/bin/bash -c 'sleep…"   2 hours ago    Up 2 hours              k8s_ascend_default-last-test-deepseek2-60b
        ```

        将该容器提交为新版本训练容器镜像，注意新镜像的tag与旧镜像不一致。示例如下：

        ```
        docker commit bb118ca00041 newimage:latest
        ```

5.  检查TaskD是否替换完成。参考[检查TaskD](#升级说明)章节，检查组件状态是否正常。
6.  修改训练脚本（例如train\_start.sh）和任务YAML。
    1.  创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os 
         
        job_id=os.getenv("MINDX_TASK_ID") 
        node_nums=XX         # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
          
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node}) 
        start_taskd_manager()
        ```

        >[!NOTE] 说明 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](./api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2.  在训练脚本中增加以下代码拉起TaskD  Manager。

        ```
        export TASKD_PROCESS_ENABLE="on" 
        # 以PyTorch框架为例
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP} 
            python manager.py &           # 具体执行路径由当前路径决定
        fi 
              
        torchrun ...
        ```

    3.  在任务YAML中修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

        ```
        ...
                spec:
        ...
                   containers:
        ...
                     ports:                          
                       - containerPort: 9601              
                         name: taskd-port
        ...
        ```

