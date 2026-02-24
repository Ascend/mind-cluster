# NPU Exporter<a name="ZH-CN_TOPIC_0000002479386812"></a>

## NPU Exporter主页<a name="ZH-CN_TOPIC_0000002479386854"></a>

**功能说明<a name="zh-cn_topic_0000001497524785_section1617874274411"></a>**

NPU Exporter的基本信息页面。

**URL<a name="zh-cn_topic_0000001497524785_section103113034014"></a>**

GET http://ip:port/

>[!NOTE] 说明 
>-   IP：在容器化部署场景中，使用容器IP；在二进制部署场景中，使用启动NPU Exporter的IP入参。
>-   port：默认为8082，部署时如有修改，使用实际部署时使用的port入参。

**请求参数<a name="zh-cn_topic_0000001497524785_section162719122175"></a>**

无

**响应说明<a name="zh-cn_topic_0000001497524785_section1433551894112"></a>**

返回一个简单的html页面。

```
<html>
   <head><title>NPU-Exporter</title></head>
   <body>
   <h1 align="center">NPU-Exporter</h1>
   <p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is http://ip:8082/metrics: <a href="./metrics">Metrics</a></p>
   </body>
   </html>
```


## Prometheus Metrics接口<a name="ZH-CN_TOPIC_0000002511426743"></a>

**功能说明<a name="zh-cn_topic_0000001446964912_section1617874274411"></a>**

提供Metrics接口，供Prometheus调用和集成。

**URL<a name="zh-cn_topic_0000001446964912_section103113034014"></a>**

GET http://ip:port/metrics

>[!NOTE] 说明 
>NPU Exporter为了安全考虑，默认启用容器级别端口（默认8082），请求IP为Kubernetes容器IP，当K8s网络插件为Calico时，网络策略设置为允许label为app=prometheus的应用访问。

**请求参数<a name="zh-cn_topic_0000001446964912_section162719122175"></a>**

无

**响应说明<a name="zh-cn_topic_0000001446964912_section1433551894112"></a>**

按照Prometheus的专用格式返回数据，相关数据信息如下所示，仅供参考，以实际回显为准。数据信息的详细说明参见下文或[数据信息说明.xlsx](../resource/数据信息说明.xlsx)。Prometheus自带数据信息无需关注，不在此展示说明。有部分数据信息仅支持某种产品形态，具体以实际上报的数据信息为准。

```
...

# HELP machine_npu_nums Amount of npu installed on the machine.
# TYPE machine_npu_nums gauge
machine_npu_nums 8
# HELP npu_chip_info_aicore_current_freq the npu ai core current frequency, unit is 'MHz'
# TYPE npu_chip_info_aicore_current_freq gauge
npu_chip_info_aicore_current_freq{container_name="",id="0",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:61:00.0",pod_name="",vdie_id="185011D4-21104518-A0C4ED94-14CC040A-56102003"} 1000 1723621883587
npu_chip_info_aicore_current_freq{container_name="",id="1",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:DB:00.0",pod_name="",vdie_id="185011D4-21E04718-93B2ED94-14CC040A-BF102003"} 1000 1723621883932
npu_chip_info_aicore_current_freq{container_name="",id="2",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:B2:00.0",pod_name="",vdie_id="185011D4-20C02418-59D4ED94-14CC040A-F9102003"} 1000 1723621884277
npu_chip_info_aicore_current_freq{container_name="",id="3",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:3E:00.0",pod_name="",vdie_id="185011D4-21502C18-0464ED94-14CC040A-6E102003"} 1000 1723621884682
npu_chip_info_aicore_current_freq{container_name="",id="4",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:60:00.0",pod_name="",vdie_id="185011D4-21A02418-64946D94-14CC040A-F8102003"} 1000 1723621885026
npu_chip_info_aicore_current_freq{container_name="",id="5",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:DA:00.0",pod_name="",vdie_id="185011DC-21F02B18-C4B66D94-14CC040A-57102003"} 1000 1723621885385
npu_chip_info_aicore_current_freq{container_name="",id="6",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:B1:00.0",pod_name="",vdie_id="185011D4-20602118-14646D94-14CC040A-8A102003"} 1000 1723621885784
npu_chip_info_aicore_current_freq{container_name="",id="7",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:3D:00.0",pod_name="",vdie_id="185011D4-21504C18-10656D94-14CC040A-6B102003"} 1000 1723621886131
# HELP npu_chip_info_bandwidth_rx the npu interface receive speed, unit is 'MB/s'
# TYPE npu_chip_info_bandwidth_rx gauge
npu_chip_info_bandwidth_rx{container_name="",id="0",model_name="910A-Ascend-V1",namespace="",pcie_bus_info="0000:61:00.0",pod_name="",vdie_id="185011D4-21104518-A0C4ED94-14CC040A-56102003"} 0 1723621883587
...
```

本接口支持查询默认指标组和自定义指标组。自定义指标组的方法详细请参见[自定义指标开发](../appendix.md#自定义指标开发)；默认指标组包含如下几个部分。指标组的采集和上报由配置文件中的开关控制，若开关配置为开启，则对应的指标组会进行采集和上报；若开关配置为关闭，则对应的指标组不会进行采集和上报。

-   [版本数据信息](#section17031652143614)
-   [NPU数据信息](#section1379685784314)
-   [vNPU数据信息](#section81411161343)
-   [Network数据信息](#section630155191018)
-   [DDR数据信息](#section11460736193116)
-   [片上内存数据信息](#section82014427452)
-   [HCCS数据信息](#section9741133815914)
-   [PCIe数据信息](#section124052024182413)
-   [RoCE数据信息](#section2080452819294)
-   [SIO数据信息](#section1773315620217)
-   [光模块数据信息](#section1692536163118)

了解以上各项数据信息中标签的说明，请参见[标签数据信息说明](#section129583351126)。

>[!NOTE] 说明 
>-   如果进程运行在主机上，Pod没有使用NPU，则pod\_name、container\_name和namespace的值将为空。
>-   NPU Exporter是通过调用底层的HDK接口，获取相应的信息。数据信息调用的HDK接口请参考[调用的HDK接口](#section1137512020304)。
>-   若查询某项数据信息时，NPU Exporter组件不支持该设备形态或调用HDK接口失败，则不会上报该数据信息。

**版本数据信息<a name="section17031652143614"></a>**

**表 1**  版本数据信息

<a name="table81981837143713"></a>
<table><thead align="left"><tr id="row319910378373"><th class="cellrowborder" valign="top" width="6.278116565030489%" id="mcps1.2.8.1.1"><p id="p622884917372"><a name="p622884917372"></a><a name="p622884917372"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="19.794061781465558%" id="mcps1.2.8.1.2"><p id="p1322844963710"><a name="p1322844963710"></a><a name="p1322844963710"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="16.7849645106468%" id="mcps1.2.8.1.3"><p id="p822864983718"><a name="p822864983718"></a><a name="p822864983718"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="13.126062181345594%" id="mcps1.2.8.1.4"><p id="p92292490370"><a name="p92292490370"></a><a name="p92292490370"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="7.807657702689193%" id="mcps1.2.8.1.5"><p id="p22291749183713"><a name="p22291749183713"></a><a name="p22291749183713"></a>字段类型</p>
</th>
<th class="cellrowborder" valign="top" width="17.234829551134656%" id="mcps1.2.8.1.6"><p id="p15229649123712"><a name="p15229649123712"></a><a name="p15229649123712"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="18.97430770768769%" id="mcps1.2.8.1.7"><p id="p1023084933716"><a name="p1023084933716"></a><a name="p1023084933716"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row191991937123720"><td class="cellrowborder" valign="top" width="6.278116565030489%" headers="mcps1.2.8.1.1 "><p id="p96501612386"><a name="p96501612386"></a><a name="p96501612386"></a>版本</p>
</td>
<td class="cellrowborder" valign="top" width="19.794061781465558%" headers="mcps1.2.8.1.2 "><p id="p36502183811"><a name="p36502183811"></a><a name="p36502183811"></a>npu_exporter_version_info</p>
</td>
<td class="cellrowborder" valign="top" width="16.7849645106468%" headers="mcps1.2.8.1.3 "><p id="p1665017112386"><a name="p1665017112386"></a><a name="p1665017112386"></a><span id="ph122556224122"><a name="ph122556224122"></a><a name="ph122556224122"></a>NPU Exporter</span>版本信息</p>
</td>
<td class="cellrowborder" valign="top" width="13.126062181345594%" headers="mcps1.2.8.1.4 "><p id="p146501619389"><a name="p146501619389"></a><a name="p146501619389"></a>exporterVersion：当前<span id="ph159901939141619"><a name="ph159901939141619"></a><a name="ph159901939141619"></a>NPU Exporter</span>版本信息</p>
</td>
<td class="cellrowborder" valign="top" width="7.807657702689193%" headers="mcps1.2.8.1.5 "><p id="p1650311389"><a name="p1650311389"></a><a name="p1650311389"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="17.234829551134656%" headers="mcps1.2.8.1.6 "><p id="p1965091173818"><a name="p1965091173818"></a><a name="p1965091173818"></a>1：占位字符，无实际含义</p>
</td>
<td class="cellrowborder" valign="top" width="18.97430770768769%" headers="mcps1.2.8.1.7 "><p id="p765051203819"><a name="p765051203819"></a><a name="p765051203819"></a>Atlas 训练系列产品</p>
<p id="p1065018116386"><a name="p1065018116386"></a><a name="p1065018116386"></a>Atlas A2 训练系列产品</p>
<p id="p106501411383"><a name="p106501411383"></a><a name="p106501411383"></a>Atlas A3 训练系列产品</p>
<p id="p46507143816"><a name="p46507143816"></a><a name="p46507143816"></a>推理服务器（插Atlas 300I 推理卡）</p>
<p id="p106501911386"><a name="p106501911386"></a><a name="p106501911386"></a>Atlas 推理系列产品</p>
<p id="p56500163810"><a name="p56500163810"></a><a name="p56500163810"></a><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span></p>
<p id="p1115563443"><a name="p1115563443"></a><a name="p1115563443"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
<p><span>推理服务器（插Atlas 350 标卡）</span></p>
</td>
</tr>
</tbody>
</table>

**NPU数据信息<a name="section1379685784314"></a>**

**表 2**  NPU数据信息

<a name="table5395205714441"></a>
<table><thead align="left"><tr id="row1039518578442"><th class="cellrowborder" valign="top" width="8.43%" id="mcps1.2.7.1.1"><p id="p5140181116462"><a name="p5140181116462"></a><a name="p5140181116462"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="23.89%" id="mcps1.2.7.1.2"><p id="p18140191112461"><a name="p18140191112461"></a><a name="p18140191112461"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.759999999999998%" id="mcps1.2.7.1.3"><p id="p11401411154619"><a name="p11401411154619"></a><a name="p11401411154619"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.7.1.4"><p id="p2014061144610"><a name="p2014061144610"></a><a name="p2014061144610"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="10.77%" id="mcps1.2.7.1.5"><p id="p13140191144612"><a name="p13140191144612"></a><a name="p13140191144612"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="20.13%" id="mcps1.2.7.1.6"><p id="p3140011184615"><a name="p3140011184615"></a><a name="p3140011184615"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row1650165618504"><td class="cellrowborder" valign="top" width="8.43%" headers="mcps1.2.7.1.1 "><p id="p78751739517"><a name="p78751739517"></a><a name="p78751739517"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="23.89%" headers="mcps1.2.7.1.2 "><p id="p58756317517"><a name="p58756317517"></a><a name="p58756317517"></a>npu_chip_info_overall_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="21.759999999999998%" headers="mcps1.2.7.1.3 "><p id="p38753314515"><a name="p38753314515"></a><a name="p38753314515"></a>昇腾AI处理器整体利用率</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.7.1.4 "><p id="p48751331519"><a name="p48751331519"></a><a name="p48751331519"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10.77%" headers="mcps1.2.7.1.5 "><p id="p198756311511"><a name="p198756311511"></a><a name="p198756311511"></a>单位：%</p>
</td>
<td class="cellrowborder" valign="top" width="20.13%" headers="mcps1.2.7.1.6 "><a name="ul11480102917516"></a><a name="ul11480102917516"></a><ul id="ul11480102917516"><li>Atlas A2 训练系列产品</li></ul>
<a name="ul0480112918517"></a><a name="ul0480112918517"></a><ul id="ul0480112918517"><li>Atlas A3 训练系列产品</li></ul>
<a name="ul648022945116"></a><a name="ul648022945116"></a><ul id="ul648022945116"><li><span id="ph279972618380"><a name="ph279972618380"></a><a name="ph279972618380"></a>Atlas 800I A2 推理服务器</span></li></ul>
<a name="ul148152911516"></a><a name="ul148152911516"></a><ul id="ul148152911516"><li><span id="ph1848152919517"><a name="ph1848152919517"></a><a name="ph1848152919517"></a>A200I A2 Box 异构组件</span></li></ul><ul><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
</td>
</tr>
<tr id="row1855012373517"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p10846935203715"><a name="p10846935203715"></a><a name="p10846935203715"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p78461435113717"><a name="p78461435113717"></a><a name="p78461435113717"></a>npu_chip_info_vector_utilization</p>
<p id="p1784653518379"><a name="p1784653518379"></a><a name="p1784653518379"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p3846133516379"><a name="p3846133516379"></a><a name="p3846133516379"></a>昇腾AI处理器AI Vector利用率</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p118468353375"><a name="p118468353375"></a><a name="p118468353375"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1846173515376"><a name="p1846173515376"></a><a name="p1846173515376"></a>单位：%</p>
</td>
<td><ul><li>Atlas A2 训练系列产品</li></ul>
<ul><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 300I 推理卡）</li></ul>
<ul><li>Atlas 推理系列产品</li><li><span>Atlas 800I A2 推理服务器</span></li></ul>
<ul><li><span>A200I A2 Box 异构组件</span></li></ul><ul><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
</td>
</tr>
<tr id="row14396115717449"><td class="cellrowborder" valign="top" width="8.43%" headers="mcps1.2.7.1.1 "><p id="p1214131124617"><a name="p1214131124617"></a><a name="p1214131124617"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="23.89%" headers="mcps1.2.7.1.2 "><p id="p0141161144613"><a name="p0141161144613"></a><a name="p0141161144613"></a>machine_npu_nums</p>
</td>
<td class="cellrowborder" valign="top" width="21.759999999999998%" headers="mcps1.2.7.1.3 "><p id="p514191164611"><a name="p514191164611"></a><a name="p514191164611"></a><span id="ph1514161113463"><a name="ph1514161113463"></a><a name="ph1514161113463"></a>昇腾AI处理器</span>数目</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.7.1.4 "><p id="p1614151144610"><a name="p1614151144610"></a><a name="p1614151144610"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="10.77%" headers="mcps1.2.7.1.5 "><p id="p114201174619"><a name="p114201174619"></a><a name="p114201174619"></a>单位：个</p>
</td>
<td class="cellrowborder" rowspan="14" valign="top" width="20.13%" headers="mcps1.2.7.1.6 "><a name="ul16872655115116"></a><a name="ul16872655115116"></a><ul id="ul16872655115116"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li></ul>
<a name="ul13196335219"></a><a name="ul13196335219"></a><ul id="ul13196335219"><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 300I 推理卡）</li></ul>
<a name="ul199548911521"></a><a name="ul199548911521"></a><ul id="ul199548911521"><li>Atlas 推理系列产品</li><li><span id="ph328863144019"><a name="ph328863144019"></a><a name="ph328863144019"></a>Atlas 800I A2 推理服务器</span></li></ul>
<a name="ul12630181715213"></a><a name="ul12630181715213"></a><ul id="ul12630181715213"><li><span id="ph14798132394418"><a name="ph14798132394418"></a><a name="ph14798132394418"></a>A200I A2 Box 异构组件</span></li></ul><ul><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
</td>
</tr>
<tr id="row1679397191510"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p3433112131511"><a name="p3433112131511"></a><a name="p3433112131511"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p6433621121518"><a name="p6433621121518"></a><a name="p6433621121518"></a>npu_chip_info_voltage</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p443302141517"><a name="p443302141517"></a><a name="p443302141517"></a><span id="ph543352115158"><a name="ph543352115158"></a><a name="ph543352115158"></a>昇腾AI处理器</span>电压</p>
<p id="p94341021131516"><a name="p94341021131516"></a><a name="p94341021131516"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p44342219152"><a name="p44342219152"></a><a name="p44342219152"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1543402161516"><a name="p1543402161516"></a><a name="p1543402161516"></a>单位：伏特（V）</p>
</td>
</tr>
<tr id="row1941145713443"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p18154111112466"><a name="p18154111112466"></a><a name="p18154111112466"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p11587113171112"><a name="p11587113171112"></a><a name="p11587113171112"></a>第一个错误码为：npu_chip_info_error_code</p>
<p id="p1658317311118"><a name="p1658317311118"></a><a name="p1658317311118"></a>其他错误码：npu_chip_info_error_code_X</p>
<p id="p61541511174619"><a name="p61541511174619"></a><a name="p61541511174619"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p41551011194610"><a name="p41551011194610"></a><a name="p41551011194610"></a><span id="ph1415520117467"><a name="ph1415520117467"></a><a name="ph1415520117467"></a>昇腾AI处理器</span>错误码</p>
<p id="p2011915418141"><a name="p2011915418141"></a><a name="p2011915418141"></a>当昇腾AI处理器上没有错误码时，不会上报该字段</p>
<div class="note" id="note71551511134616"><a name="note71551511134616"></a><a name="note71551511134616"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul01551011174616"></a><a name="ul01551011174616"></a><ul id="ul01551011174616"><li>Prometheus场景：若该<span id="ph4155141154620"><a name="ph4155141154620"></a><a name="ph4155141154620"></a>昇腾AI处理器</span>上同时存在多个错误码，由于Prometheus格式限制，当前只支持上报前十个出现的错误码。X的取值范围：1~9</li><li>Telegraf场景：最多支持上报128个错误码。</li><li>错误码的详细说明，请参见<a href="../appendix.md#芯片故障码参考文档">芯片故障码参考文档</a>章节。</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1222255524811"><a name="p1222255524811"></a><a name="p1222255524811"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p11551111194615"><a name="p11551111194615"></a><a name="p11551111194615"></a>-</p>
</td>
</tr>
<tr id="row54151957134410"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1515861164611"><a name="p1515861164611"></a><a name="p1515861164611"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p17158811104613"><a name="p17158811104613"></a><a name="p17158811104613"></a>npu_chip_info_name</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p191581911144614"><a name="p191581911144614"></a><a name="p191581911144614"></a><span id="ph015881104617"><a name="ph015881104617"></a><a name="ph015881104617"></a>昇腾AI处理器</span>名称和ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p979666193712"><a name="p979666193712"></a><a name="p979666193712"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p5159411164612"><a name="p5159411164612"></a><a name="p5159411164612"></a>-</p>
</td>
</tr>
<tr id="row1841755718445"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p31611311104617"><a name="p31611311104617"></a><a name="p31611311104617"></a>NPU</p>
<p id="p1660815114465"><a name="p1660815114465"></a><a name="p1660815114465"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1916141114611"><a name="p1916141114611"></a><a name="p1916141114611"></a>npu_chip_info_health_status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p316151110465"><a name="p316151110465"></a><a name="p316151110465"></a><span id="ph216161112462"><a name="ph216161112462"></a><a name="ph216161112462"></a>昇腾AI处理器</span>健康状态</p>
<p id="p060821114612"><a name="p060821114612"></a><a name="p060821114612"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1253882563818"><a name="p1253882563818"></a><a name="p1253882563818"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p18162911154611"><a name="p18162911154611"></a><a name="p18162911154611"></a>取值为0或1</p>
<a name="ul1216271124619"></a><a name="ul1216271124619"></a><ul id="ul1216271124619"><li>1：健康</li><li>0：不健康</li></ul>
</td>
</tr>
<tr id="row68995374572"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p4931350175717"><a name="p4931350175717"></a><a name="p4931350175717"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p159311550135719"><a name="p159311550135719"></a><a name="p159311550135719"></a>npu_chip_info_temperature</p>
<p id="p79317505579"><a name="p79317505579"></a><a name="p79317505579"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p19931135017570"><a name="p19931135017570"></a><a name="p19931135017570"></a><span id="ph29318508572"><a name="ph29318508572"></a><a name="ph29318508572"></a>昇腾AI处理器</span>温度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1893265011576"><a name="p1893265011576"></a><a name="p1893265011576"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p3932750205718"><a name="p3932750205718"></a><a name="p3932750205718"></a>单位：摄氏度（℃）</p>
</td>
</tr>
<tr id="row1573713209594"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p557522325912"><a name="p557522325912"></a><a name="p557522325912"></a>NPU</p>
<p id="p7575172316596"><a name="p7575172316596"></a><a name="p7575172316596"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p257512318599"><a name="p257512318599"></a><a name="p257512318599"></a>npu_chip_info_process_info</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p8576122312591"><a name="p8576122312591"></a><a name="p8576122312591"></a><span id="ph757602325915"><a name="ph757602325915"></a><a name="ph757602325915"></a>占用<span id="ph45761023205915"><a name="ph45761023205915"></a><a name="ph45761023205915"></a>昇腾AI处理器</span>进程的信息</span></p>
<a name="ul157652325915"></a><a name="ul157652325915"></a><ul id="ul157652325915"><li><span id="ph057610232595"><a name="ph057610232595"></a><a name="ph057610232595"></a>Prometheus场景：取值为进程使用的内存</span></li><li>Telegraf场景：仅当没有进程占用昇腾AI处理器时上报，值为0</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p557615232599"><a name="p557615232599"></a><a name="p557615232599"></a><a href="#table191895615241">标签2</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p65761923135912"><a name="p65761923135912"></a><a name="p65761923135912"></a>单位：MB</p>
<p id="p9576102313596"><a name="p9576102313596"></a><a name="p9576102313596"></a></p>
</td>
</tr>
<tr id="row193211218116"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p0380154816"><a name="p0380154816"></a><a name="p0380154816"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p038012419110"><a name="p038012419110"></a><a name="p038012419110"></a>npu_chip_info_process_info_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p17380949116"><a name="p17380949116"></a><a name="p17380949116"></a>占用<span id="ph16380541813"><a name="ph16380541813"></a><a name="ph16380541813"></a>昇腾AI处理器</span>的进程数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p73801841116"><a name="p73801841116"></a><a name="p73801841116"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p14380945113"><a name="p14380945113"></a><a name="p14380945113"></a>单位：个</p>
</td>
</tr>
<tr id="row797104620184"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p566019620196"><a name="p566019620196"></a><a name="p566019620196"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p146607681918"><a name="p146607681918"></a><a name="p146607681918"></a>npu_container_info</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1266186161910"><a name="p1266186161910"></a><a name="p1266186161910"></a>NPU容器信息</p>
<div class="note" id="note15661264198"><a name="note15661264198"></a><a name="note15661264198"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p76611668193"><a name="p76611668193"></a><a name="p76611668193"></a>Telegraf不支持上报该指标</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1966106141918"><a name="p1966106141918"></a><a name="p1966106141918"></a><a href="#table191895615241">标签5</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p7661563191"><a name="p7661563191"></a><a name="p7661563191"></a>-</p>
</td>
</tr>
<tr id="row19419657124420"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p516521154612"><a name="p516521154612"></a><a name="p516521154612"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p16165101174617"><a name="p16165101174617"></a><a name="p16165101174617"></a>npu_chip_info_power</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p616601115461"><a name="p616601115461"></a><a name="p616601115461"></a><span id="ph216617115462"><a name="ph216617115462"></a><a name="ph216617115462"></a>昇腾AI处理器</span>功耗</p>
<div class="note" id="note125551721182116"><a name="note125551721182116"></a><a name="note125551721182116"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p2555192112218"><a name="p2555192112218"></a><a name="p2555192112218"></a>只有Atlas 推理系列产品为板卡功耗，其余产品为<span id="ph13555921192113"><a name="ph13555921192113"></a><a name="ph13555921192113"></a>昇腾AI处理器</span>功耗</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1837111312397"><a name="p1837111312397"></a><a name="p1837111312397"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p16167131144610"><a name="p16167131144610"></a><a name="p16167131144610"></a>单位：瓦特（W）</p>
</td>
</tr>
<tr id="row17316516325"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p691711583213"><a name="p691711583213"></a><a name="p691711583213"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p791781503215"><a name="p791781503215"></a><a name="p791781503215"></a>npu_chip_info_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p791831523215"><a name="p791831523215"></a><a name="p791831523215"></a><span id="ph149181315193212"><a name="ph149181315193212"></a><a name="ph149181315193212"></a>昇腾AI处理器</span>AI Core利用率</p>
<p id="p191811520325"><a name="p191811520325"></a><a name="p191811520325"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1191861563217"><a name="p1191861563217"></a><a name="p1191861563217"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1991831510326"><a name="p1991831510326"></a><a name="p1991831510326"></a>单位：%</p>
</td>
</tr>
<tr id="row14668153293218"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p8326245103217"><a name="p8326245103217"></a><a name="p8326245103217"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p2326184543212"><a name="p2326184543212"></a><a name="p2326184543212"></a>npu_chip_info_aicore_current_freq</p>
<p id="p632604563212"><a name="p632604563212"></a><a name="p632604563212"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p5326204533217"><a name="p5326204533217"></a><a name="p5326204533217"></a><span id="ph9326144583218"><a name="ph9326144583218"></a><a name="ph9326144583218"></a>昇腾AI处理器</span>的AI Core当前频率</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p153271145183219"><a name="p153271145183219"></a><a name="p153271145183219"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p632716458324"><a name="p632716458324"></a><a name="p632716458324"></a>单位：MHz</p>
</td>
</tr>
<tr id="row131882916359"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1876434118351"><a name="p1876434118351"></a><a name="p1876434118351"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1576464114356"><a name="p1576464114356"></a><a name="p1576464114356"></a>container_npu_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1076564110356"><a name="p1076564110356"></a><a name="p1076564110356"></a>带有容器信息的NPU的AI Core利用率</p>
<div class="note" id="note15509181573618"><a name="note15509181573618"></a><a name="note15509181573618"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p1851016154367"><a name="p1851016154367"></a><a name="p1851016154367"></a>Telegraf不支持上报该指标</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p9765641143519"><a name="p9765641143519"></a><a name="p9765641143519"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p11765541123510"><a name="p11765541123510"></a><a name="p11765541123510"></a>单位：%</p>
</td>
</tr>
<tr id="row37623275413"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1376232724113"><a name="p1376232724113"></a><a name="p1376232724113"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p197621927104119"><a name="p197621927104119"></a><a name="p197621927104119"></a>npu_chip_info_serial_number</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p11762142754113"><a name="p11762142754113"></a><a name="p11762142754113"></a>昇腾AI处理器序列号</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p192641024134213"><a name="p192641024134213"></a><a name="p192641024134213"></a><a href="#table191895615241">标签6</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p14763162754112"><a name="p14763162754112"></a><a name="p14763162754112"></a>1：占位字符，无实际含义</p>
</td>
</tr>
<tr id="row1143413571446"><td class="cellrowborder" valign="top" width="8.43%" headers="mcps1.2.7.1.1 "><p id="p1346910221177"><a name="p1346910221177"></a><a name="p1346910221177"></a>NPU</p>
<p id="p1469322121711"><a name="p1469322121711"></a><a name="p1469322121711"></a></p>
<p id="p19469122261719"><a name="p19469122261719"></a><a name="p19469122261719"></a></p>
<p id="p6469132281714"><a name="p6469132281714"></a><a name="p6469132281714"></a></p>
</td>
<td class="cellrowborder" valign="top" width="23.89%" headers="mcps1.2.7.1.2 "><p id="p134692223176"><a name="p134692223176"></a><a name="p134692223176"></a>npu_chip_info_network_status</p>
<p id="p154691022121714"><a name="p154691022121714"></a><a name="p154691022121714"></a></p>
<p id="p746952219173"><a name="p746952219173"></a><a name="p746952219173"></a></p>
<p id="p846917229177"><a name="p846917229177"></a><a name="p846917229177"></a></p>
</td>
<td class="cellrowborder" valign="top" width="21.759999999999998%" headers="mcps1.2.7.1.3 "><p id="p14469152218174"><a name="p14469152218174"></a><a name="p14469152218174"></a><span id="ph19469122211176"><a name="ph19469122211176"></a><a name="ph19469122211176"></a>昇腾AI处理器</span>网络健康状态</p>
<p id="p346912281714"><a name="p346912281714"></a><a name="p346912281714"></a></p>
<p id="p3469172219173"><a name="p3469172219173"></a><a name="p3469172219173"></a></p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.7.1.4 "><p id="p8469192218177"><a name="p8469192218177"></a><a name="p8469192218177"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10.77%" headers="mcps1.2.7.1.5 "><p id="p1146982211711"><a name="p1146982211711"></a><a name="p1146982211711"></a>取值为0或1</p>
<a name="ul1469162261710"></a><a name="ul1469162261710"></a><ul id="ul1469162261710"><li>1：健康，可以连通</li><li>0：不健康，无法连通</li></ul>
</td>
<td class="cellrowborder" valign="top" width="20.13%" headers="mcps1.2.7.1.6 "><a name="ul221951194615"></a><a name="ul221951194615"></a><ul id="ul221951194615"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph1067265144615"><a name="ph1067265144615"></a><a name="ph1067265144615"></a>A200I A2 Box 异构组件</span></li><li><span id="ph16900101265315"><a name="ph16900101265315"></a><a name="ph16900101265315"></a>Atlas 800I A2 推理服务器</span></li></ul>
<p id="p1858910113463"><a name="p1858910113463"></a><a name="p1858910113463"></a></p>
</td>
</tr>
<tr id="row9448057204414"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1421521164618"><a name="p1421521164618"></a><a name="p1421521164618"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p13216171154611"><a name="p13216171154611"></a><a name="p13216171154611"></a>container_npu_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p72161211174615"><a name="p72161211174615"></a><a name="p72161211174615"></a>带有容器信息的NPU内存总大小</p>
<div class="note" id="note47464273266"><a name="note47464273266"></a><a name="note47464273266"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p167469276269"><a name="p167469276269"></a><a name="p167469276269"></a>Telegraf不支持上报该指标</p>
</div></div>
<p id="p5592191117464"><a name="p5592191117464"></a><a name="p5592191117464"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p859221114516"><a name="p859221114516"></a><a name="p859221114516"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p621891154618"><a name="p621891154618"></a><a name="p621891154618"></a>单位：MB</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20.13%" headers="mcps1.2.7.1.6 "><ul><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li>A200I A2 Box 异构组件</li><li>Atlas 800I A2 推理服务器</li><li>推理服务器（插Atlas 350 标卡）</li></ul>
<p id="p1858910113463"><a name="p1858910113463"></a><a name="p1858910113463"></a></p>
</td>
</tr>
<tr id="row1450957104413"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p322318116463"><a name="p322318116463"></a><a name="p322318116463"></a>NPU</p>
<p id="p358911112467"><a name="p358911112467"></a><a name="p358911112467"></a></p>
<p id="p958991154618"><a name="p958991154618"></a><a name="p958991154618"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p13224141194611"><a name="p13224141194611"></a><a name="p13224141194611"></a>container_npu_used_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p112259114466"><a name="p112259114466"></a><a name="p112259114466"></a>带有容器信息的NPU已使用内存</p>
<div class="note" id="note1789193552611"><a name="note1789193552611"></a><a name="note1789193552611"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p8891143517268"><a name="p8891143517268"></a><a name="p8891143517268"></a>Telegraf不支持上报该指标</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1369214472452"><a name="p1369214472452"></a><a name="p1369214472452"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p15226191164615"><a name="p15226191164615"></a><a name="p15226191164615"></a>单位：MB</p>
</td>
</tr>
<tr><td><p>NPU</p>
</td>
<td><p>npu_chip_info_product_type</p>
</td>
<td><p>昇腾AI处理器产品形态</p>
</td>
<td><p><a href="#table191895615241">标签7</a></p>
</td>
<td><p>1：占位字符，无实际含义</p>
</td>
<td><p>Atlas 推理系列产品</p>
</td>
</tr>
<tr><td><p>NPU</p>
</td>
<td><p>npu_chip_info_cube_utilization</p>
</td>
<td><p>昇腾AI处理器AI Cube利用率</p>
</td>
<td><p><a href="#table191895615241">标签1</a></p>
</td>
<td><p>单位：%</p>
</td>
<td><ul><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li>Atlas 800I A2 推理服务器</li><li>A200I A2 Box 异构组件</li></ul>
</td>
</tr>
</tbody>
</table>

**vNPU数据信息<a name="section81411161343"></a>**

**表 3**  vNPU数据信息

<a name="table176992573417"></a>
<table><thead align="left"><tr id="row147006579418"><th class="cellrowborder" valign="top" width="8.37%" id="mcps1.2.7.1.1"><p id="p1359454518512"><a name="p1359454518512"></a><a name="p1359454518512"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="24.62%" id="mcps1.2.7.1.2"><p id="p659410454514"><a name="p659410454514"></a><a name="p659410454514"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.7.1.3"><p id="p559418452513"><a name="p559418452513"></a><a name="p559418452513"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="14.899999999999999%" id="mcps1.2.7.1.4"><p id="p145958451556"><a name="p145958451556"></a><a name="p145958451556"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="10.43%" id="mcps1.2.7.1.5"><p id="p1059519451752"><a name="p1059519451752"></a><a name="p1059519451752"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="20.18%" id="mcps1.2.7.1.6"><p id="p5596145452"><a name="p5596145452"></a><a name="p5596145452"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row1470011579414"><td class="cellrowborder" valign="top" width="8.37%" headers="mcps1.2.7.1.1 "><p id="p43597271390"><a name="p43597271390"></a><a name="p43597271390"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.62%" headers="mcps1.2.7.1.2 "><p id="p435920278918"><a name="p435920278918"></a><a name="p435920278918"></a>vnpu_pod_aicore_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.7.1.3 "><p id="p1835982714920"><a name="p1835982714920"></a><a name="p1835982714920"></a>vNPU的AI Core利用率</p>
</td>
<td class="cellrowborder" valign="top" width="14.899999999999999%" headers="mcps1.2.7.1.4 "><p id="p848015281474"><a name="p848015281474"></a><a name="p848015281474"></a><a href="#table191895615241">标签3</a></p>
</td>
<td class="cellrowborder" valign="top" width="10.43%" headers="mcps1.2.7.1.5 "><p id="p1835917279917"><a name="p1835917279917"></a><a name="p1835917279917"></a>单位：%</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="20.18%" headers="mcps1.2.7.1.6 "><p id="p735982715917"><a name="p735982715917"></a><a name="p735982715917"></a><span id="ph19590185162111"><a name="ph19590185162111"></a><a name="ph19590185162111"></a>Atlas 推理系列产品</span></p>
</td>
</tr>
<tr id="row17703155715411"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p136032717910"><a name="p136032717910"></a><a name="p136032717910"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p153601427591"><a name="p153601427591"></a><a name="p153601427591"></a>vnpu_pod_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p136032711910"><a name="p136032711910"></a><a name="p136032711910"></a>vNPU拥有的总内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1852652910818"><a name="p1852652910818"></a><a name="p1852652910818"></a><a href="#table191895615241">标签3</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p4360927892"><a name="p4360927892"></a><a name="p4360927892"></a>单位：KB</p>
</td>
</tr>
<tr id="row20706115712414"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p18361152717916"><a name="p18361152717916"></a><a name="p18361152717916"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p13361127993"><a name="p13361127993"></a><a name="p13361127993"></a>vnpu_pod_used_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1236115276919"><a name="p1236115276919"></a><a name="p1236115276919"></a>vNPU使用中的内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p19249134618106"><a name="p19249134618106"></a><a name="p19249134618106"></a><a href="#table191895615241">标签3</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p163619271590"><a name="p163619271590"></a><a name="p163619271590"></a>单位：KB</p>
</td>
</tr>
</tbody>
</table>

**Network数据信息<a name="section630155191018"></a>**

<a name="table164281059191110"></a>
<table><thead align="left"><tr id="row5428155912116"><th class="cellrowborder" valign="top" width="11.21%" id="mcps1.1.7.1.1"><p id="p2429195912115"><a name="p2429195912115"></a><a name="p2429195912115"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="21.73%" id="mcps1.1.7.1.2"><p id="p64291159101114"><a name="p64291159101114"></a><a name="p64291159101114"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.61%" id="mcps1.1.7.1.3"><p id="p5429759111115"><a name="p5429759111115"></a><a name="p5429759111115"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.540000000000001%" id="mcps1.1.7.1.4"><p id="p942911590111"><a name="p942911590111"></a><a name="p942911590111"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="10%" id="mcps1.1.7.1.5"><p id="p26148601218"><a name="p26148601218"></a><a name="p26148601218"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="19.91%" id="mcps1.1.7.1.6"><p id="p1424452617302"><a name="p1424452617302"></a><a name="p1424452617302"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row0429155915119"><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p id="p547216401235"><a name="p547216401235"></a><a name="p547216401235"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p id="p847284010318"><a name="p847284010318"></a><a name="p847284010318"></a>npu_chip_info_bandwidth_rx</p>
<p id="p184729405313"><a name="p184729405313"></a><a name="p184729405313"></a></p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p id="p5472174019315"><a name="p5472174019315"></a><a name="p5472174019315"></a><span id="ph84728401534"><a name="ph84728401534"></a><a name="ph84728401534"></a>昇腾AI处理器</span>网口实时接收速率。</p>
<p id="p144738404315"><a name="p144738404315"></a><a name="p144738404315"></a></p>
</td>
<td class="cellrowborder" valign="top" width="15.540000000000001%" headers="mcps1.1.7.1.4 "><p id="p22385016353"><a name="p22385016353"></a><a name="p22385016353"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p id="p188904163413"><a name="p188904163413"></a><a name="p188904163413"></a>单位：MB/s</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="19.91%" headers="mcps1.1.7.1.6 "><a name="ul178907161943"></a><a name="ul178907161943"></a><ul id="ul178907161943"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph57012578543"><a name="ph57012578543"></a><a name="ph57012578543"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph1518064711478"><a name="ph1518064711478"></a><a name="ph1518064711478"></a>A200I A2 Box 异构组件</span></li></ul>
</td>
</tr>
<tr id="row11943132171414"><td class="cellrowborder" valign="top" headers="mcps1.1.7.1.1 "><p id="p163872531349"><a name="p163872531349"></a><a name="p163872531349"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.2 "><p id="p173871953046"><a name="p173871953046"></a><a name="p173871953046"></a>npu_chip_info_bandwidth_tx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.3 "><p id="p43871653644"><a name="p43871653644"></a><a name="p43871653644"></a><span id="ph438713531644"><a name="ph438713531644"></a><a name="ph438713531644"></a>昇腾AI处理器</span>网口实时发送速率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.4 "><p id="p17534950171412"><a name="p17534950171412"></a><a name="p17534950171412"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.5 "><p id="p15556656691"><a name="p15556656691"></a><a name="p15556656691"></a>单位：MB/s</p>
<p id="p1655613561890"><a name="p1655613561890"></a><a name="p1655613561890"></a></p>
</td>
</tr>
<tr id="row2097471041420"><td class="cellrowborder" valign="top" headers="mcps1.1.7.1.1 "><p id="p5177695400"><a name="p5177695400"></a><a name="p5177695400"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.2 "><p id="p1469741634018"><a name="p1469741634018"></a><a name="p1469741634018"></a>npu_chip_info_link_status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.3 "><p id="p472685912436"><a name="p472685912436"></a><a name="p472685912436"></a><span id="ph67261259164316"><a name="ph67261259164316"></a><a name="ph67261259164316"></a>昇腾AI处理器</span>网口Link状态。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.4 "><p id="p62411497407"><a name="p62411497407"></a><a name="p62411497407"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.5 "><p id="p18658038144417"><a name="p18658038144417"></a><a name="p18658038144417"></a>取值为0或1</p>
<a name="ul136589389444"></a><a name="ul136589389444"></a><ul id="ul136589389444"><li>1：UP</li><li>0：DOWN</li></ul>
</td>
</tr>
<tr id="row128958179146"><td class="cellrowborder" valign="top" headers="mcps1.1.7.1.1 "><p id="p1895101717145"><a name="p1895101717145"></a><a name="p1895101717145"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.2 "><p id="p116111121174615"><a name="p116111121174615"></a><a name="p116111121174615"></a>npu_chip_link_speed</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.3 "><p id="p6866132914619"><a name="p6866132914619"></a><a name="p6866132914619"></a><span id="ph115291716171817"><a name="ph115291716171817"></a><a name="ph115291716171817"></a>昇腾AI处理器</span>网口默认速率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.4 "><p id="p18906696167"><a name="p18906696167"></a><a name="p18906696167"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.5 "><p id="p728312308482"><a name="p728312308482"></a><a name="p728312308482"></a>单位：MB/s</p>
</td>
</tr>
<tr id="row1063192616140"><td class="cellrowborder" valign="top" headers="mcps1.1.7.1.1 "><p id="p12632265145"><a name="p12632265145"></a><a name="p12632265145"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.2 "><p id="p460342419500"><a name="p460342419500"></a><a name="p460342419500"></a>npu_chip_link_up_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.3 "><p id="p136392641412"><a name="p136392641412"></a><a name="p136392641412"></a><span id="ph18383151912186"><a name="ph18383151912186"></a><a name="ph18383151912186"></a>昇腾AI处理器</span>网口UP的统计次数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.4 "><p id="p559312144517"><a name="p559312144517"></a><a name="p559312144517"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.7.1.5 "><p id="p142831501527"><a name="p142831501527"></a><a name="p142831501527"></a>单位：次</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_status_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口Link状态。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。</p>
</td>
<td class="cellrowborder" valign="top" width="15.540000000000001%" headers="mcps1.1.7.1.4 "><p><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>取值为0或1</p><ul><li>1：UP</li><li>0：DOWN</li></ul>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="19.91%" headers="mcps1.1.7.1.6 "><p><span>推理服务器（插Atlas 350 标卡 4Pmesh互联）</span></p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口实时接收速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。使用的-time参数为100。</p>
</td>
<td class="cellrowborder" valign="top" width="15.540000000000001%" headers="mcps1.1.7.1.4 "><p><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口实时发送速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。使用的-time参数为100。</p>
</td>
<td class="cellrowborder" valign="top" width="15.540000000000001%" headers="mcps1.1.7.1.4 "><p><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_speed_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>物理端口的速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。</p>
</td>
<td class="cellrowborder" valign="top" width="15.540000000000001%" headers="mcps1.1.7.1.4 "><p><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：G</p>
</td>
</tr>
</tbody>
</table>

**DDR数据信息<a name="section11460736193116"></a>**

**表 4**  DDR数据信息

<a name="table1251541123212"></a>
<table><thead align="left"><tr id="row152510419324"><th class="cellrowborder" valign="top" width="9.460946094609461%" id="mcps1.2.7.1.1"><p id="p7637172411347"><a name="p7637172411347"></a><a name="p7637172411347"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="23.812381238123812%" id="mcps1.2.7.1.2"><p id="p126377248349"><a name="p126377248349"></a><a name="p126377248349"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.85208520852085%" id="mcps1.2.7.1.3"><p id="p2643152493410"><a name="p2643152493410"></a><a name="p2643152493410"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="16.251625162516252%" id="mcps1.2.7.1.4"><p id="p1064472493415"><a name="p1064472493415"></a><a name="p1064472493415"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="9.770977097709771%" id="mcps1.2.7.1.5"><p id="p364592493419"><a name="p364592493419"></a><a name="p364592493419"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="19.851985198519852%" id="mcps1.2.7.1.6"><p id="p8645142463412"><a name="p8645142463412"></a><a name="p8645142463412"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row7261741103215"><td class="cellrowborder" valign="top" width="9.460946094609461%" headers="mcps1.2.7.1.1 "><p id="p198470485402"><a name="p198470485402"></a><a name="p198470485402"></a>DDR</p>
</td>
<td class="cellrowborder" valign="top" width="23.812381238123812%" headers="mcps1.2.7.1.2 "><p id="p13847154834019"><a name="p13847154834019"></a><a name="p13847154834019"></a>npu_chip_info_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="20.85208520852085%" headers="mcps1.2.7.1.3 "><p id="p68477486408"><a name="p68477486408"></a><a name="p68477486408"></a><span id="ph16848144816402"><a name="ph16848144816402"></a><a name="ph16848144816402"></a>昇腾AI处理器</span>DDR内存已使用量</p>
</td>
<td class="cellrowborder" valign="top" width="16.251625162516252%" headers="mcps1.2.7.1.4 "><p id="p13852164814011"><a name="p13852164814011"></a><a name="p13852164814011"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="9.770977097709771%" headers="mcps1.2.7.1.5 "><p id="p17849174864015"><a name="p17849174864015"></a><a name="p17849174864015"></a>单位：MB</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="19.851985198519852%" headers="mcps1.2.7.1.6 "><a name="ul12849124816407"></a><a name="ul12849124816407"></a><ul id="ul12849124816407"><li>Atlas 训练系列产品</li><li>推理服务器（插Atlas 300I 推理卡）</li><li>Atlas 推理系列产品</li></ul>
</td>
</tr>
<tr id="row20281241173220"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p178531848194013"><a name="p178531848194013"></a><a name="p178531848194013"></a>DDR</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p13854194819400"><a name="p13854194819400"></a><a name="p13854194819400"></a>npu_chip_info_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1785424824020"><a name="p1785424824020"></a><a name="p1785424824020"></a><span id="ph8854114810407"><a name="ph8854114810407"></a><a name="ph8854114810407"></a>昇腾AI处理器</span>DDR内存总量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1725112158412"><a name="p1725112158412"></a><a name="p1725112158412"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p138551548194015"><a name="p138551548194015"></a><a name="p138551548194015"></a>单位：MB</p>
</td>
</tr>
</tbody>
</table>

**片上内存数据信息<a name="section82014427452"></a>**

**表 5**  片上内存数据信息

<a name="table1989710355466"></a>
<table><thead align="left"><tr id="row10898113515468"><th class="cellrowborder" valign="top" width="7.370000000000002%" id="mcps1.2.7.1.1"><p id="p19552123871112"><a name="p19552123871112"></a><a name="p19552123871112"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="25.910000000000004%" id="mcps1.2.7.1.2"><p id="p105521238171114"><a name="p105521238171114"></a><a name="p105521238171114"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.180000000000003%" id="mcps1.2.7.1.3"><p id="p1455312382119"><a name="p1455312382119"></a><a name="p1455312382119"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.990000000000002%" id="mcps1.2.7.1.4"><p id="p13553153831114"><a name="p13553153831114"></a><a name="p13553153831114"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="10.4%" id="mcps1.2.7.1.5"><p id="p1255412381116"><a name="p1255412381116"></a><a name="p1255412381116"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="20.150000000000006%" id="mcps1.2.7.1.6"><p id="p1055583851117"><a name="p1055583851117"></a><a name="p1055583851117"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row11899143564615"><td class="cellrowborder" valign="top" width="7.370000000000002%" headers="mcps1.2.7.1.1 "><p id="p18851135485"><a name="p18851135485"></a><a name="p18851135485"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" width="25.910000000000004%" headers="mcps1.2.7.1.2 "><p id="p1888613137483"><a name="p1888613137483"></a><a name="p1888613137483"></a>npu_chip_info_hbm_used_memory</p>
<p id="p1350201414811"><a name="p1350201414811"></a><a name="p1350201414811"></a></p>
</td>
<td class="cellrowborder" valign="top" width="20.180000000000003%" headers="mcps1.2.7.1.3 "><p id="p1088621354812"><a name="p1088621354812"></a><a name="p1088621354812"></a><span id="ph6886181316484"><a name="ph6886181316484"></a><a name="ph6886181316484"></a>昇腾AI处理器</span>片上内存已使用量</p>
<p id="p115021414815"><a name="p115021414815"></a><a name="p115021414815"></a></p>
</td>
<td class="cellrowborder" valign="top" width="15.990000000000002%" headers="mcps1.2.7.1.4 "><p id="p1170184010372"><a name="p1170184010372"></a><a name="p1170184010372"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10.4%" headers="mcps1.2.7.1.5 "><p id="p138871313174819"><a name="p138871313174819"></a><a name="p138871313174819"></a>单位：MB</p>
<p id="p649111418488"><a name="p649111418488"></a><a name="p649111418488"></a></p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="20.150000000000006%" headers="mcps1.2.7.1.6 "><a name="ul588814137484"></a><a name="ul588814137484"></a><ul id="ul588814137484"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph043025116483"><a name="ph043025116483"></a><a name="ph043025116483"></a>A200I A2 Box 异构组件</span></li><li><span id="ph19520133125919"><a name="ph19520133125919"></a><a name="ph19520133125919"></a>Atlas 800I A2 推理服务器</span></li></ul><ul><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
<p id="p1393532018545"><a name="p1393532018545"></a><a name="p1393532018545"></a></p>
<p id="p674513211898"><a name="p674513211898"></a><a name="p674513211898"></a></p>
<p id="p25074235513"><a name="p25074235513"></a><a name="p25074235513"></a></p>
<p id="p1943172101118"><a name="p1943172101118"></a><a name="p1943172101118"></a></p>
<p id="p633418586217"><a name="p633418586217"></a><a name="p633418586217"></a></p>
<p id="p18975952161118"><a name="p18975952161118"></a><a name="p18975952161118"></a></p>
<p id="p1511933016172"><a name="p1511933016172"></a><a name="p1511933016172"></a></p>
<p id="p27112544188"><a name="p27112544188"></a><a name="p27112544188"></a></p>
<p id="p43007141914"><a name="p43007141914"></a><a name="p43007141914"></a></p>
</td>
</tr>
<tr id="row16902735174610"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1889119132482"><a name="p1889119132482"></a><a name="p1889119132482"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p989116139485"><a name="p989116139485"></a><a name="p989116139485"></a>npu_chip_info_hbm_total_memory</p>
<p id="p445121413487"><a name="p445121413487"></a><a name="p445121413487"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p2089219136481"><a name="p2089219136481"></a><a name="p2089219136481"></a><span id="ph789211364816"><a name="ph789211364816"></a><a name="ph789211364816"></a>昇腾AI处理器</span>片上总内存</p>
<p id="p1645141464814"><a name="p1645141464814"></a><a name="p1645141464814"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p184021251193712"><a name="p184021251193712"></a><a name="p184021251193712"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1389314135484"><a name="p1389314135484"></a><a name="p1389314135484"></a>单位：MB</p>
</td>
</tr>
<tr id="row1890503513462"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p989731312486"><a name="p989731312486"></a><a name="p989731312486"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1489711132488"><a name="p1489711132488"></a><a name="p1489711132488"></a>npu_chip_info_hbm_ecc_enable_flag</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p168978133484"><a name="p168978133484"></a><a name="p168978133484"></a><span id="ph168971313174819"><a name="ph168971313174819"></a><a name="ph168971313174819"></a>昇腾AI处理器</span>片上内存的ECC使能状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p41235553716"><a name="p41235553716"></a><a name="p41235553716"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1898013114817"><a name="p1898013114817"></a><a name="p1898013114817"></a>取值为1或0</p>
<a name="ul089881311484"></a><a name="ul089881311484"></a><ul id="ul089881311484"><li>0：ECC检测未使能</li><li>1：ECC检测使能</li></ul>
</td>
</tr>
<tr id="row59081535154619"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p99021013104819"><a name="p99021013104819"></a><a name="p99021013104819"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1990201324811"><a name="p1990201324811"></a><a name="p1990201324811"></a>npu_chip_info_hbm_ecc_single_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p090301394810"><a name="p090301394810"></a><a name="p090301394810"></a><span id="ph13903161324818"><a name="ph13903161324818"></a><a name="ph13903161324818"></a>昇腾AI处理器</span>片上内存单比特当前错误计数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1690781316485"><a name="p1690781316485"></a><a name="p1690781316485"></a></p>
<p id="p10663121112385"><a name="p10663121112385"></a><a name="p10663121112385"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p490341314818"><a name="p490341314818"></a><a name="p490341314818"></a>-</p>
</td>
</tr>
<tr id="row179108357465"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p4982016714"><a name="p4982016714"></a><a name="p4982016714"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p17907913194813"><a name="p17907913194813"></a><a name="p17907913194813"></a>npu_chip_info_hbm_ecc_double_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p2907151315481"><a name="p2907151315481"></a><a name="p2907151315481"></a><span id="ph794012452114"><a name="ph794012452114"></a><a name="ph794012452114"></a>昇腾AI处理器</span>片上内存多比特当前错误计数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p158511556116"><a name="p158511556116"></a><a name="p158511556116"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p12329871222"><a name="p12329871222"></a><a name="p12329871222"></a>-</p>
</td>
</tr>
<tr id="row19913335114610"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p109133136481"><a name="p109133136481"></a><a name="p109133136481"></a>片上内存</p>
<p id="p1232171484810"><a name="p1232171484810"></a><a name="p1232171484810"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p5913151318489"><a name="p5913151318489"></a><a name="p5913151318489"></a>npu_chip_info_hbm_ecc_total_single_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p16914141320484"><a name="p16914141320484"></a><a name="p16914141320484"></a><span id="ph29141513154819"><a name="ph29141513154819"></a><a name="ph29141513154819"></a>昇腾AI处理器</span>片上内存生命周期内所有单比特错误数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p191721313487"><a name="p191721313487"></a><a name="p191721313487"></a></p>
<p id="p412703015388"><a name="p412703015388"></a><a name="p412703015388"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p3915101312481"><a name="p3915101312481"></a><a name="p3915101312481"></a>-</p>
</td>
</tr>
<tr id="row991619353468"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1918113104815"><a name="p1918113104815"></a><a name="p1918113104815"></a>片上内存</p>
<p id="p1528201417487"><a name="p1528201417487"></a><a name="p1528201417487"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1891820133481"><a name="p1891820133481"></a><a name="p1891820133481"></a>npu_chip_info_hbm_ecc_total_double_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1991941316488"><a name="p1991941316488"></a><a name="p1991941316488"></a><span id="ph1991981374817"><a name="ph1991981374817"></a><a name="ph1991981374817"></a>昇腾AI处理器</span>片上内存生命周期内所有多比特错误数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1562817341380"><a name="p1562817341380"></a><a name="p1562817341380"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p3920171314810"><a name="p3920171314810"></a><a name="p3920171314810"></a>-</p>
</td>
</tr>
<tr id="row59191035164613"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p144311826115"><a name="p144311826115"></a><a name="p144311826115"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p189251013134815"><a name="p189251013134815"></a><a name="p189251013134815"></a>npu_chip_info_hbm_ecc_single_bit_isolated_pages_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p2092620134489"><a name="p2092620134489"></a><a name="p2092620134489"></a><span id="ph189261413134817"><a name="ph189261413134817"></a><a name="ph189261413134817"></a>昇腾AI处理器</span>片上内存单比特错误隔离内存页数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p10195144313814"><a name="p10195144313814"></a><a name="p10195144313814"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p19271413134815"><a name="p19271413134815"></a><a name="p19271413134815"></a>-</p>
</td>
</tr>
<tr id="row139221435134616"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p593301324814"><a name="p593301324814"></a><a name="p593301324814"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p7933111311485"><a name="p7933111311485"></a><a name="p7933111311485"></a>npu_chip_info_hbm_ecc_double_bit_isolated_pages_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p8934111311488"><a name="p8934111311488"></a><a name="p8934111311488"></a><span id="ph1993417133487"><a name="ph1993417133487"></a><a name="ph1993417133487"></a>昇腾AI处理器</span>片上内存多比特错误隔离内存页数量。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1565510515387"><a name="p1565510515387"></a><a name="p1565510515387"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p139351113204819"><a name="p139351113204819"></a><a name="p139351113204819"></a>-</p>
</td>
</tr>
<tr id="row15165225122814"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p6165132519287"><a name="p6165132519287"></a><a name="p6165132519287"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p716520252283"><a name="p716520252283"></a><a name="p716520252283"></a>npu_chip_info_hbm_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p17115131573117"><a name="p17115131573117"></a><a name="p17115131573117"></a>昇腾AI处理器的片上内存利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1762442113719"><a name="p1762442113719"></a><a name="p1762442113719"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p85175417305"><a name="p85175417305"></a><a name="p85175417305"></a>单位：%</p>
</td>
</tr>
<tr id="row471165412187"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p16912264198"><a name="p16912264198"></a><a name="p16912264198"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p5691826121920"><a name="p5691826121920"></a><a name="p5691826121920"></a>npu_chip_info_hbm_temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p12691112601911"><a name="p12691112601911"></a><a name="p12691112601911"></a>昇腾AI处理器片上内存的温度。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p0711454121814"><a name="p0711454121814"></a><a name="p0711454121814"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1871185411812"><a name="p1871185411812"></a><a name="p1871185411812"></a>单位：&deg;C</p>
</td>
</tr>
<tr id="row123017710196"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p36914261197"><a name="p36914261197"></a><a name="p36914261197"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p5691142631920"><a name="p5691142631920"></a><a name="p5691142631920"></a>npu_chip_info_hbm_bandwidth_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p86911726101915"><a name="p86911726101915"></a><a name="p86911726101915"></a>昇腾AI处理器片上内存的带宽利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p7308715195"><a name="p7308715195"></a><a name="p7308715195"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1130779198"><a name="p1130779198"></a><a name="p1130779198"></a>单位：%</p>
</td>
</tr>
</tbody>
</table>

**HCCS数据信息<a name="section9741133815914"></a>**

**表 6**  HCCS数据信息

<a name="table812924831013"></a>
<table><thead align="left"><tr id="row6130748141017"><th class="cellrowborder" valign="top" width="7.76%" id="mcps1.2.7.1.1"><p id="p93074164111"><a name="p93074164111"></a><a name="p93074164111"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="26.200000000000003%" id="mcps1.2.7.1.2"><p id="p83072164112"><a name="p83072164112"></a><a name="p83072164112"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="25.53%" id="mcps1.2.7.1.3"><p id="p10308141621110"><a name="p10308141621110"></a><a name="p10308141621110"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="12.72%" id="mcps1.2.7.1.4"><p id="p203091716191115"><a name="p203091716191115"></a><a name="p203091716191115"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="8.88%" id="mcps1.2.7.1.5"><p id="p1131011166113"><a name="p1131011166113"></a><a name="p1131011166113"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="18.91%" id="mcps1.2.7.1.6"><p id="p731101691113"><a name="p731101691113"></a><a name="p731101691113"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row51311483103"><td class="cellrowborder" valign="top" width="7.76%" headers="mcps1.2.7.1.1 "><p id="p192174718131"><a name="p192174718131"></a><a name="p192174718131"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" width="26.200000000000003%" headers="mcps1.2.7.1.2 "><p id="p1313153991311"><a name="p1313153991311"></a><a name="p1313153991311"></a>npu_chip_info_hccs_statistic_info_tx_cnt_X</p>
<p id="p104597463146"><a name="p104597463146"></a><a name="p104597463146"></a>X范围：1~7（Atlas A2 训练系列产品或Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" width="25.53%" headers="mcps1.2.7.1.3 "><a name="ul64654414321"></a><a name="ul64654414321"></a><ul id="ul64654414321"><li>第X个HDLC链路发送报文数，单位是flit。</li><li>采集失败时上报-1。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="12.72%" headers="mcps1.2.7.1.4 "><p id="p2046102919516"><a name="p2046102919516"></a><a name="p2046102919516"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="8.88%" headers="mcps1.2.7.1.5 "><p id="p1696115661316"><a name="p1696115661316"></a><a name="p1696115661316"></a>-</p>
</td>
<td class="cellrowborder" rowspan="31" valign="top" width="18.91%" headers="mcps1.2.7.1.6 "><a name="ul14353121612171"></a><a name="ul14353121612171"></a><ul id="ul14353121612171"><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li></ul>
</td>
</tr>
<tr id="row1134048201013"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p131591118181"><a name="p131591118181"></a><a name="p131591118181"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p8269148191816"><a name="p8269148191816"></a><a name="p8269148191816"></a>npu_chip_info_hccs_statistic_info_rx_cnt_X</p>
<p id="p19920451111518"><a name="p19920451111518"></a><a name="p19920451111518"></a>X范围：1~7（Atlas A2 训练系列产品或Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><a name="ul14760124163111"></a><a name="ul14760124163111"></a><ul id="ul14760124163111"><li>第X个HDLC链路接收报文数，单位是flit。</li><li>采集失败时上报-1。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p29016402514"><a name="p29016402514"></a><a name="p29016402514"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p14702194714197"><a name="p14702194714197"></a><a name="p14702194714197"></a>-</p>
</td>
</tr>
<tr id="row111434263218"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1582372462510"><a name="p1582372462510"></a><a name="p1582372462510"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p082322414251"><a name="p082322414251"></a><a name="p082322414251"></a>npu_chip_info_hccs_statistic_info_crc_err_cnt_X</p>
<p id="p9478114473615"><a name="p9478114473615"></a><a name="p9478114473615"></a>X范围：1~7（Atlas A2 训练系列产品、<span id="ph1519015228918"><a name="ph1519015228918"></a><a name="ph1519015228918"></a>Atlas 900 A3 SuperPoD 超节点</span>），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><a name="ul5374194612915"></a><a name="ul5374194612915"></a><ul id="ul5374194612915"><li>第X个HDLC链路接收报文crc错误，单位是flit。</li><li>采集失败时上报-1。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p969619232262"><a name="p969619232262"></a><a name="p969619232262"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p20568654143813"><a name="p20568654143813"></a><a name="p20568654143813"></a>-</p>
</td>
</tr>
<tr id="row12081052145718"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p125454065918"><a name="p125454065918"></a><a name="p125454065918"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p0541440145910"><a name="p0541440145910"></a><a name="p0541440145910"></a>npu_chip_info_hccs_bandwidth_info_profiling_time</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p954114095919"><a name="p954114095919"></a><a name="p954114095919"></a>HCCS链路带宽采样时长，取值范围1~1000</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p18627201614536"><a name="p18627201614536"></a><a name="p18627201614536"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p11101030613"><a name="p11101030613"></a><a name="p11101030613"></a>单位：ms</p>
</td>
</tr>
<tr id="row175309715812"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p686416451828"><a name="p686416451828"></a><a name="p686416451828"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p686412451026"><a name="p686412451026"></a><a name="p686412451026"></a>npu_chip_info_hccs_bandwidth_info_total_tx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p178641345920"><a name="p178641345920"></a><a name="p178641345920"></a>HCCS链路总发送数据带宽，采集失败时上报-1</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p123732216538"><a name="p123732216538"></a><a name="p123732216538"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p15101133914319"><a name="p15101133914319"></a><a name="p15101133914319"></a>单位：GB/s</p>
</td>
</tr>
<tr id="row553337135814"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p159518501567"><a name="p159518501567"></a><a name="p159518501567"></a>HCCS</p>
<p id="p19536117135810"><a name="p19536117135810"></a><a name="p19536117135810"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p29513509610"><a name="p29513509610"></a><a name="p29513509610"></a>npu_chip_info_hccs_bandwidth_info_total_rx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p89512502619"><a name="p89512502619"></a><a name="p89512502619"></a>HCCS链路总接收数据带宽，采集失败时上报-1</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p513103314535"><a name="p513103314535"></a><a name="p513103314535"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p19451733781"><a name="p19451733781"></a><a name="p19451733781"></a>单位：GB/s</p>
</td>
</tr>
<tr id="row3533195313714"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p23261414161118"><a name="p23261414161118"></a><a name="p23261414161118"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p9326014101111"><a name="p9326014101111"></a><a name="p9326014101111"></a>npu_chip_info_hccs_bandwidth_info_tx_X</p>
<p id="p1048653410101"><a name="p1048653410101"></a><a name="p1048653410101"></a>X范围：1~7（Atlas A2 训练系列产品、<span id="ph5486133415104"><a name="ph5486133415104"></a><a name="ph5486133415104"></a>Atlas 900 A3 SuperPoD 超节点</span>），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p16326161421112"><a name="p16326161421112"></a><a name="p16326161421112"></a>HCCS单链路发送数据带宽，采集失败时上报-1</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p679117537101"><a name="p679117537101"></a><a name="p679117537101"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p7271611124"><a name="p7271611124"></a><a name="p7271611124"></a>单位：GB/s</p>
</td>
</tr>
<tr id="row15235427121014"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p416181411132"><a name="p416181411132"></a><a name="p416181411132"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1916181418136"><a name="p1916181418136"></a><a name="p1916181418136"></a>npu_chip_info_hccs_bandwidth_info_rx_X</p>
<p id="p1161714161312"><a name="p1161714161312"></a><a name="p1161714161312"></a></p>
<p id="p1110523181119"><a name="p1110523181119"></a><a name="p1110523181119"></a>X范围：1~7（Atlas A2 训练系列产品、<span id="ph1811015231112"><a name="ph1811015231112"></a><a name="ph1811015231112"></a>Atlas 900 A3 SuperPoD 超节点</span>），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p191681411135"><a name="p191681411135"></a><a name="p191681411135"></a>HCCS单链路接收数据带宽，采集失败时上报-1</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p99951417125412"><a name="p99951417125412"></a><a name="p99951417125412"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p8114134441317"><a name="p8114134441317"></a><a name="p8114134441317"></a>单位：GB/s</p>
</td>
</tr>
</tbody>
</table>

**PCIe数据信息<a name="section124052024182413"></a>**

**表 7**  PCIe数据信息

<a name="table1341911380255"></a>
<table><thead align="left"><tr id="row941993842520"><th class="cellrowborder" valign="top" width="7.080000000000002%" id="mcps1.2.7.1.1"><p id="p3550823152717"><a name="p3550823152717"></a><a name="p3550823152717"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="23.580000000000005%" id="mcps1.2.7.1.2"><p id="p10551823102710"><a name="p10551823102710"></a><a name="p10551823102710"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="19.790000000000003%" id="mcps1.2.7.1.3"><p id="p15525231279"><a name="p15525231279"></a><a name="p15525231279"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.200000000000003%" id="mcps1.2.7.1.4"><p id="p105522023172715"><a name="p105522023172715"></a><a name="p105522023172715"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="12.440000000000001%" id="mcps1.2.7.1.5"><p id="p1355312311278"><a name="p1355312311278"></a><a name="p1355312311278"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="21.91%" id="mcps1.2.7.1.6"><p id="p12554423192716"><a name="p12554423192716"></a><a name="p12554423192716"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row174206389252"><td class="cellrowborder" valign="top" width="7.080000000000002%" headers="mcps1.2.7.1.1 "><p id="p9435131620283"><a name="p9435131620283"></a><a name="p9435131620283"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" width="23.580000000000005%" headers="mcps1.2.7.1.2 "><p id="p104351116142814"><a name="p104351116142814"></a><a name="p104351116142814"></a>npu_chip_info_pcie_rx_p_bw</p>
<p id="p259631611288"><a name="p259631611288"></a><a name="p259631611288"></a></p>
</td>
<td class="cellrowborder" valign="top" width="19.790000000000003%" headers="mcps1.2.7.1.3 "><p id="p12368105055510"><a name="p12368105055510"></a><a name="p12368105055510"></a><span id="ph143686506554"><a name="ph143686506554"></a><a name="ph143686506554"></a>昇腾AI处理器</span>接收远端写的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" width="15.200000000000003%" headers="mcps1.2.7.1.4 "><p id="p5444316162814"><a name="p5444316162814"></a><a name="p5444316162814"></a><a href="#table191895615241">标签4</a></p>
</td>
<td class="cellrowborder" valign="top" width="12.440000000000001%" headers="mcps1.2.7.1.5 "><p id="p11437141652810"><a name="p11437141652810"></a><a name="p11437141652810"></a>单位：MB/ms</p>
<p id="p1159519168288"><a name="p1159519168288"></a><a name="p1159519168288"></a></p>
</td>
<td class="cellrowborder" rowspan="47" valign="top" width="21.91%" headers="mcps1.2.7.1.6 "><a name="ul64395165289"></a><a name="ul64395165289"></a><ul id="ul64395165289"><li>Atlas A2 训练系列产品</li><li><span id="ph9506329133010"><a name="ph9506329133010"></a><a name="ph9506329133010"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph83840211502"><a name="ph83840211502"></a><a name="ph83840211502"></a>A200I A2 Box 异构组件</span></li></ul>
</td>
</tr>
<tr id="row34233384255"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p12447181662811"><a name="p12447181662811"></a><a name="p12447181662811"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p54471716162810"><a name="p54471716162810"></a><a name="p54471716162810"></a>npu_chip_info_pcie_rx_np_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1625830105612"><a name="p1625830105612"></a><a name="p1625830105612"></a><span id="ph20258100175620"><a name="ph20258100175620"></a><a name="ph20258100175620"></a>昇腾AI处理器</span>接收远端读的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p16432491272"><a name="p16432491272"></a><a name="p16432491272"></a><a href="#table191895615241">标签4</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p7450131652812"><a name="p7450131652812"></a><a name="p7450131652812"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row54261838112518"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1045717162286"><a name="p1045717162286"></a><a name="p1045717162286"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p345713168285"><a name="p345713168285"></a><a name="p345713168285"></a>npu_chip_info_pcie_rx_cpl_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p177633865611"><a name="p177633865611"></a><a name="p177633865611"></a><span id="ph168271294568"><a name="ph168271294568"></a><a name="ph168271294568"></a>昇腾AI处理器</span>从远端读收到CPL回复的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p14651516112814"><a name="p14651516112814"></a><a name="p14651516112814"></a><a href="#table191895615241">标签4</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p2460916132812"><a name="p2460916132812"></a><a name="p2460916132812"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row184291938102510"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p64665168288"><a name="p64665168288"></a><a name="p64665168288"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p11467141615283"><a name="p11467141615283"></a><a name="p11467141615283"></a>npu_chip_info_pcie_tx_p_bw</p>
<p id="p16582191612820"><a name="p16582191612820"></a><a name="p16582191612820"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p746851613289"><a name="p746851613289"></a><a name="p746851613289"></a><span id="ph1444162375613"><a name="ph1444162375613"></a><a name="ph1444162375613"></a>昇腾AI处理器</span>向远端写PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p187703161481"><a name="p187703161481"></a><a name="p187703161481"></a><a href="#table191895615241">标签4</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1246931672811"><a name="p1246931672811"></a><a name="p1246931672811"></a>单位：MB/ms</p>
<p id="p1758211161285"><a name="p1758211161285"></a><a name="p1758211161285"></a></p>
</td>
</tr>
<tr id="row34321838132519"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p3479516182815"><a name="p3479516182815"></a><a name="p3479516182815"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p12479191642816"><a name="p12479191642816"></a><a name="p12479191642816"></a>npu_chip_info_pcie_tx_np_bw</p>
<p id="p16578181619284"><a name="p16578181619284"></a><a name="p16578181619284"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p048081662818"><a name="p048081662818"></a><a name="p048081662818"></a><span id="ph148731632125613"><a name="ph148731632125613"></a><a name="ph148731632125613"></a>昇腾AI处理器</span>从远端读PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p3130323785"><a name="p3130323785"></a><a name="p3130323785"></a><a href="#table191895615241">标签4</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p848212160286"><a name="p848212160286"></a><a name="p848212160286"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row743773816250"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1148912160281"><a name="p1148912160281"></a><a name="p1148912160281"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p7489416162813"><a name="p7489416162813"></a><a name="p7489416162813"></a>npu_chip_info_pcie_tx_cpl_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p144901216132813"><a name="p144901216132813"></a><a name="p144901216132813"></a><span id="ph1269610482563"><a name="ph1269610482563"></a><a name="ph1269610482563"></a>昇腾AI处理器</span>回复远端读操作CPL的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p968710261484"><a name="p968710261484"></a><a name="p968710261484"></a><a href="#table191895615241">标签4</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p949171612813"><a name="p949171612813"></a><a name="p949171612813"></a>单位：MB/ms</p>
</td>
</tr>
</tbody>
</table>

**RoCE数据信息<a name="section2080452819294"></a>**

**表 8**  RoCE数据信息

<a name="table16943172263012"></a>
<table><thead align="left"><tr id="row11943122133012"><th class="cellrowborder" valign="top" width="8.52%" id="mcps1.2.7.1.1"><p id="p14944105716305"><a name="p14944105716305"></a><a name="p14944105716305"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="21.51%" id="mcps1.2.7.1.2"><p id="p1094435710307"><a name="p1094435710307"></a><a name="p1094435710307"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.08%" id="mcps1.2.7.1.3"><p id="p199451578301"><a name="p199451578301"></a><a name="p199451578301"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="14.299999999999999%" id="mcps1.2.7.1.4"><p id="p0945125712309"><a name="p0945125712309"></a><a name="p0945125712309"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="10.36%" id="mcps1.2.7.1.5"><p id="p14946135793014"><a name="p14946135793014"></a><a name="p14946135793014"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="24.23%" id="mcps1.2.7.1.6"><p id="p159478570306"><a name="p159478570306"></a><a name="p159478570306"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row394462203014"><td class="cellrowborder" valign="top" width="8.52%" headers="mcps1.2.7.1.1 "><p id="p679820203318"><a name="p679820203318"></a><a name="p679820203318"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" width="21.51%" headers="mcps1.2.7.1.2 "><p id="p779870103313"><a name="p779870103313"></a><a name="p779870103313"></a>npu_chip_mac_rx_pause_num</p>
<p id="p94131244717"><a name="p94131244717"></a><a name="p94131244717"></a></p>
</td>
<td class="cellrowborder" valign="top" width="21.08%" headers="mcps1.2.7.1.3 "><p id="p77991300331"><a name="p77991300331"></a><a name="p77991300331"></a>MAC接收的Pause帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" width="14.299999999999999%" headers="mcps1.2.7.1.4 "><p id="p139017337194"><a name="p139017337194"></a><a name="p139017337194"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="10.36%" headers="mcps1.2.7.1.5 "><p id="p78527532154"><a name="p78527532154"></a><a name="p78527532154"></a>-</p>
</td>
<td class="cellrowborder" rowspan="21" valign="top" width="24.23%" headers="mcps1.2.7.1.6 "><a name="ul199447020331"></a><a name="ul199447020331"></a><ul id="ul199447020331"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph5953105733115"><a name="ph5953105733115"></a><a name="ph5953105733115"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph1247784012511"><a name="ph1247784012511"></a><a name="ph1247784012511"></a>A200I A2 Box 异构组件</span></li></ul>
<p id="p152876112332"><a name="p152876112332"></a><a name="p152876112332"></a></p>
</td>
</tr>
<tr id="row69472221300"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p188092006337"><a name="p188092006337"></a><a name="p188092006337"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p168091707339"><a name="p168091707339"></a><a name="p168091707339"></a>npu_chip_mac_tx_pause_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p88109043312"><a name="p88109043312"></a><a name="p88109043312"></a>MAC发送的Pause帧总报文数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p97111720112016"><a name="p97111720112016"></a><a name="p97111720112016"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p280905151513"><a name="p280905151513"></a><a name="p280905151513"></a>-</p>
</td>
</tr>
<tr id="row295292213015"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p28193023312"><a name="p28193023312"></a><a name="p28193023312"></a>RoCE</p>
<p id="p17347611331"><a name="p17347611331"></a><a name="p17347611331"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p15820110123310"><a name="p15820110123310"></a><a name="p15820110123310"></a>npu_chip_mac_rx_pfc_pkt_num</p>
<p id="p1134715113332"><a name="p1134715113332"></a><a name="p1134715113332"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p19821170173317"><a name="p19821170173317"></a><a name="p19821170173317"></a>MAC接收的PFC帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p31848359205"><a name="p31848359205"></a><a name="p31848359205"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p469035112152"><a name="p469035112152"></a><a name="p469035112152"></a>-</p>
</td>
</tr>
<tr id="row495922219309"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p3829904337"><a name="p3829904337"></a><a name="p3829904337"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p20829170173316"><a name="p20829170173316"></a><a name="p20829170173316"></a>npu_chip_mac_tx_pfc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p78305013316"><a name="p78305013316"></a><a name="p78305013316"></a>MAC发送的PFC帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p4822184872013"><a name="p4822184872013"></a><a name="p4822184872013"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1856945115154"><a name="p1856945115154"></a><a name="p1856945115154"></a>-</p>
</td>
</tr>
<tr id="row1896252213015"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p985570183313"><a name="p985570183313"></a><a name="p985570183313"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p128552010331"><a name="p128552010331"></a><a name="p128552010331"></a>npu_chip_mac_rx_bad_pkt_num</p>
<p id="p533913114336"><a name="p533913114336"></a><a name="p533913114336"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p188569017338"><a name="p188569017338"></a><a name="p188569017338"></a>MAC接收的坏包总报文数。</p>
<p id="p1433931193310"><a name="p1433931193310"></a><a name="p1433931193310"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p15621668213"><a name="p15621668213"></a><a name="p15621668213"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p8446851181514"><a name="p8446851181514"></a><a name="p8446851181514"></a>-</p>
</td>
</tr>
<tr id="row13964102210309"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p168681803337"><a name="p168681803337"></a><a name="p168681803337"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p28694019338"><a name="p28694019338"></a><a name="p28694019338"></a>npu_chip_mac_tx_bad_pkt_num</p>
<p id="p73346113333"><a name="p73346113333"></a><a name="p73346113333"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p12870170173314"><a name="p12870170173314"></a><a name="p12870170173314"></a>MAC发送的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p11308152352118"><a name="p11308152352118"></a><a name="p11308152352118"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1787216019335"><a name="p1787216019335"></a><a name="p1787216019335"></a>-</p>
</td>
</tr>
<tr id="row3967322193019"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p58791604332"><a name="p58791604332"></a><a name="p58791604332"></a>RoCE</p>
<p id="p17328181153319"><a name="p17328181153319"></a><a name="p17328181153319"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p888017073318"><a name="p888017073318"></a><a name="p888017073318"></a>npu_chip_mac_tx_bad_oct_num</p>
<p id="p2328117334"><a name="p2328117334"></a><a name="p2328117334"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p2881807333"><a name="p2881807333"></a><a name="p2881807333"></a>MAC发送的坏包总报文字节数。</p>
<p id="p19328919335"><a name="p19328919335"></a><a name="p19328919335"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p8626140112114"><a name="p8626140112114"></a><a name="p8626140112114"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p118845011336"><a name="p118845011336"></a><a name="p118845011336"></a>-</p>
</td>
</tr>
<tr id="row9971142223016"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2089460193310"><a name="p2089460193310"></a><a name="p2089460193310"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p389510013314"><a name="p389510013314"></a><a name="p389510013314"></a>npu_chip_mac_rx_bad_oct_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p68994019333"><a name="p68994019333"></a><a name="p68994019333"></a>MAC接收的坏包总报文字节数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p179161350142115"><a name="p179161350142115"></a><a name="p179161350142115"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p890115012339"><a name="p890115012339"></a><a name="p890115012339"></a>-</p>
</td>
</tr>
<tr id="row14977122193011"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1091110018334"><a name="p1091110018334"></a><a name="p1091110018334"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p9912604333"><a name="p9912604333"></a><a name="p9912604333"></a>npu_chip_roce_rx_all_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1391380143315"><a name="p1391380143315"></a><a name="p1391380143315"></a>RoCE接收的总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p199252002337"><a name="p199252002337"></a><a name="p199252002337"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p159163016332"><a name="p159163016332"></a><a name="p159163016332"></a>-</p>
</td>
</tr>
<tr id="row798011225303"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p392616011330"><a name="p392616011330"></a><a name="p392616011330"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p12926120163315"><a name="p12926120163315"></a><a name="p12926120163315"></a>npu_chip_roce_tx_all_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p7927508336"><a name="p7927508336"></a><a name="p7927508336"></a>RoCE发送的总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p95780355222"><a name="p95780355222"></a><a name="p95780355222"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p109298011339"><a name="p109298011339"></a><a name="p109298011339"></a>-</p>
</td>
</tr>
<tr id="row15983152211300"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p189401302335"><a name="p189401302335"></a><a name="p189401302335"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p209401705334"><a name="p209401705334"></a><a name="p209401705334"></a>npu_chip_roce_rx_err_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1294212023310"><a name="p1294212023310"></a><a name="p1294212023310"></a>RoCE接收的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p411325192215"><a name="p411325192215"></a><a name="p411325192215"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p9943190163310"><a name="p9943190163310"></a><a name="p9943190163310"></a>-</p>
</td>
</tr>
<tr id="row16985132213307"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p59553083313"><a name="p59553083313"></a><a name="p59553083313"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p595513043317"><a name="p595513043317"></a><a name="p595513043317"></a>npu_chip_roce_tx_err_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p09565014332"><a name="p09565014332"></a><a name="p09565014332"></a>RoCE发送的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1280355162319"><a name="p1280355162319"></a><a name="p1280355162319"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p495812033313"><a name="p495812033313"></a><a name="p495812033313"></a>-</p>
</td>
</tr>
<tr id="row59878224307"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p19967206337"><a name="p19967206337"></a><a name="p19967206337"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p16967506333"><a name="p16967506333"></a><a name="p16967506333"></a>npu_chip_roce_rx_cnp_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1496913013316"><a name="p1496913013316"></a><a name="p1496913013316"></a>RoCE接收的CNP类型报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1798013063317"><a name="p1798013063317"></a><a name="p1798013063317"></a></p>
<p id="p1442410232238"><a name="p1442410232238"></a><a name="p1442410232238"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p159710063310"><a name="p159710063310"></a><a name="p159710063310"></a>-</p>
</td>
</tr>
<tr id="row599232216309"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1498150193319"><a name="p1498150193319"></a><a name="p1498150193319"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p9982009337"><a name="p9982009337"></a><a name="p9982009337"></a>npu_chip_roce_tx_cnp_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1998419015335"><a name="p1998419015335"></a><a name="p1998419015335"></a>RoCE发送的CNP类型报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1207163318"><a name="p1207163318"></a><a name="p1207163318"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p16986120173314"><a name="p16986120173314"></a><a name="p16986120173314"></a>-</p>
</td>
</tr>
<tr id="row129941422203012"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p817133319"><a name="p817133319"></a><a name="p817133319"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p91114332"><a name="p91114332"></a><a name="p91114332"></a>npu_chip_roce_new_pkt_rty_num</p>
<p id="p52987183312"><a name="p52987183312"></a><a name="p52987183312"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p42131193320"><a name="p42131193320"></a><a name="p42131193320"></a>RoCE发送的重传的数量统计。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p3607125102618"><a name="p3607125102618"></a><a name="p3607125102618"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p2519163315"><a name="p2519163315"></a><a name="p2519163315"></a>-</p>
</td>
</tr>
<tr id="row15999192216301"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p21318113330"><a name="p21318113330"></a><a name="p21318113330"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p131419123319"><a name="p131419123319"></a><a name="p131419123319"></a>npu_chip_roce_unexpected_ack_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p6153112339"><a name="p6153112339"></a><a name="p6153112339"></a>RoCE接收的非预期ACK报文数，NPU做丢弃处理，不影响业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p42551203313"><a name="p42551203313"></a><a name="p42551203313"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p51751123319"><a name="p51751123319"></a><a name="p51751123319"></a>-</p>
</td>
</tr>
<tr id="row428238309"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p82615111337"><a name="p82615111337"></a><a name="p82615111337"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p62717123315"><a name="p62717123315"></a><a name="p62717123315"></a>npu_chip_roce_out_of_order_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p19281311335"><a name="p19281311335"></a><a name="p19281311335"></a>RoCE接收的PSN&gt;预期PSN的报文，或重复PSN报文数。乱序或丢包，会触发重传。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p153916118334"><a name="p153916118334"></a><a name="p153916118334"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p153117153319"><a name="p153117153319"></a><a name="p153117153319"></a>-</p>
</td>
</tr>
<tr id="row16415237309"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p12413115332"><a name="p12413115332"></a><a name="p12413115332"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p44110115338"><a name="p44110115338"></a><a name="p44110115338"></a>npu_chip_roce_verification_err_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p19437113337"><a name="p19437113337"></a><a name="p19437113337"></a>RoCE接收的域段校验失败的报文数，域段校验的场景包括：ICRC、报文长度、目的端口号等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1778763362711"><a name="p1778763362711"></a><a name="p1778763362711"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p945141173316"><a name="p945141173316"></a><a name="p945141173316"></a>-</p>
</td>
</tr>
<tr id="row669231301"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p15561411331"><a name="p15561411331"></a><a name="p15561411331"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p25614113312"><a name="p25614113312"></a><a name="p25614113312"></a>npu_chip_roce_qp_status_err_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p45812163314"><a name="p45812163314"></a><a name="p45812163314"></a>RoCE接收的QP连接状态异常产生的报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p743193811346"><a name="p743193811346"></a><a name="p743193811346"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p1960014331"><a name="p1960014331"></a><a name="p1960014331"></a>-</p>
</td>
</tr>
<tr id="row17901231204610"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p18550154310478"><a name="p18550154310478"></a><a name="p18550154310478"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p261881623220"><a name="p261881623220"></a><a name="p261881623220"></a>npu_chip_info_rx_ecn_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1619111611326"><a name="p1619111611326"></a><a name="p1619111611326"></a><span id="ph0619181633215"><a name="ph0619181633215"></a><a name="ph0619181633215"></a>昇腾AI处理器</span>网络接收ECN数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p415213402817"><a name="p415213402817"></a><a name="p415213402817"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p11943192854813"><a name="p11943192854813"></a><a name="p11943192854813"></a>-</p>
</td>
</tr>
<tr id="row19921163819465"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p12559172615114"><a name="p12559172615114"></a><a name="p12559172615114"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p313510495010"><a name="p313510495010"></a><a name="p313510495010"></a>npu_chip_info_rx_fcs_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p14135164175013"><a name="p14135164175013"></a><a name="p14135164175013"></a><span id="ph313511465016"><a name="ph313511465016"></a><a name="ph313511465016"></a>昇腾AI处理器</span>网络接收FCS数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p17343195212818"><a name="p17343195212818"></a><a name="p17343195212818"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p5495553185019"><a name="p5495553185019"></a><a name="p5495553185019"></a>-</p>
</td>
</tr>
</tbody>
</table>

**SIO数据信息<a name="section1773315620217"></a>**

**表 9**  SIO数据信息

<a name="table128661257212"></a>
<table><thead align="left"><tr id="row886614251216"><th class="cellrowborder" valign="top" width="8.16%" id="mcps1.2.7.1.1"><p id="p7564115242211"><a name="p7564115242211"></a><a name="p7564115242211"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="22.470000000000002%" id="mcps1.2.7.1.2"><p id="p7564115220221"><a name="p7564115220221"></a><a name="p7564115220221"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="14.7%" id="mcps1.2.7.1.3"><p id="p17565135292219"><a name="p17565135292219"></a><a name="p17565135292219"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="25.8%" id="mcps1.2.7.1.4"><p id="p12566145213227"><a name="p12566145213227"></a><a name="p12566145213227"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="9.180000000000001%" id="mcps1.2.7.1.5"><p id="p85671652132218"><a name="p85671652132218"></a><a name="p85671652132218"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="19.689999999999998%" id="mcps1.2.7.1.6"><p id="p05681652102210"><a name="p05681652102210"></a><a name="p05681652102210"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row1086411514228"><td class="cellrowborder" valign="top" width="8.16%" headers="mcps1.2.7.1.1 "><p id="p3112122822318"><a name="p3112122822318"></a><a name="p3112122822318"></a>SIO</p>
</td>
<td class="cellrowborder" valign="top" width="22.470000000000002%" headers="mcps1.2.7.1.2 "><p id="p14112132820236"><a name="p14112132820236"></a><a name="p14112132820236"></a>npu_chip_info_sio_crc_tx_err_cnt</p>
</td>
<td class="cellrowborder" valign="top" width="14.7%" headers="mcps1.2.7.1.3 "><p id="p1511282872312"><a name="p1511282872312"></a><a name="p1511282872312"></a>SIO发送的错包数</p>
</td>
<td class="cellrowborder" valign="top" width="25.8%" headers="mcps1.2.7.1.4 "><p id="p11362410683"><a name="p11362410683"></a><a name="p11362410683"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="9.180000000000001%" headers="mcps1.2.7.1.5 "><p id="p0112152802316"><a name="p0112152802316"></a><a name="p0112152802316"></a>-</p>
</td>
<td class="cellrowborder" rowspan="13" valign="top" width="19.689999999999998%" headers="mcps1.2.7.1.6 "><ul><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 350 标卡）</li></ul>
<p id="p181568282234"><a name="p181568282234"></a><a name="p181568282234"></a></p>
</td>
</tr>
<tr id="row65141835162217"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p3113152882314"><a name="p3113152882314"></a><a name="p3113152882314"></a>SIO</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p16113182810235"><a name="p16113182810235"></a><a name="p16113182810235"></a>npu_chip_info_sio_crc_rx_err_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p3114112818230"><a name="p3114112818230"></a><a name="p3114112818230"></a>SIO接收的错包数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1536220101385"><a name="p1536220101385"></a><a name="p1536220101385"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p10114122802314"><a name="p10114122802314"></a><a name="p10114122802314"></a>-</p>
</td>
</tr>
</tbody>
</table>

**光模块数据信息<a name="section1692536163118"></a>**

**表 10**  光模块数据信息

<a name="table1845716484313"></a>
<table><thead align="left"><tr id="row17457848123120"><th class="cellrowborder" valign="top" width="8.150815081508151%" id="mcps1.2.7.1.1"><p id="p1360995513110"><a name="p1360995513110"></a><a name="p1360995513110"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="22.852285228522852%" id="mcps1.2.7.1.2"><p id="p16610145510317"><a name="p16610145510317"></a><a name="p16610145510317"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="19.401940194019403%" id="mcps1.2.7.1.3"><p id="p186111255183113"><a name="p186111255183113"></a><a name="p186111255183113"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="14.761476147614763%" id="mcps1.2.7.1.4"><p id="p12611145553111"><a name="p12611145553111"></a><a name="p12611145553111"></a>数据信息标签字段</p>
</th>
<th class="cellrowborder" valign="top" width="11.96119611961196%" id="mcps1.2.7.1.5"><p id="p261315510311"><a name="p261315510311"></a><a name="p261315510311"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="22.872287228722872%" id="mcps1.2.7.1.6"><p id="p116141355133113"><a name="p116141355133113"></a><a name="p116141355133113"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row14581348173118"><td class="cellrowborder" valign="top" width="8.150815081508151%" headers="mcps1.2.7.1.1 "><p id="p812632173512"><a name="p812632173512"></a><a name="p812632173512"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" width="22.852285228522852%" headers="mcps1.2.7.1.2 "><p id="p81266218359"><a name="p81266218359"></a><a name="p81266218359"></a>npu_chip_optical_state</p>
</td>
<td class="cellrowborder" valign="top" width="19.401940194019403%" headers="mcps1.2.7.1.3 "><p id="p212710214352"><a name="p212710214352"></a><a name="p212710214352"></a>光模块在位状态</p>
</td>
<td class="cellrowborder" valign="top" width="14.761476147614763%" headers="mcps1.2.7.1.4 "><p id="p129671862189"><a name="p129671862189"></a><a name="p129671862189"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" width="11.96119611961196%" headers="mcps1.2.7.1.5 "><p id="p20129112123513"><a name="p20129112123513"></a><a name="p20129112123513"></a>取值为0或1</p>
<a name="ul5129202143514"></a><a name="ul5129202143514"></a><ul id="ul5129202143514"><li>0：不在位</li><li>1：在位</li></ul>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="22.872287228722872%" headers="mcps1.2.7.1.6 "><a name="ul151317217352"></a><a name="ul151317217352"></a><ul id="ul151317217352"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li><span id="ph113282153511"><a name="ph113282153511"></a><a name="ph113282153511"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><p id="p1546725019404"><a name="p1546725019404"></a><a name="p1546725019404"></a><span id="ph19390121883919"><a name="ph19390121883919"></a><a name="ph19390121883919"></a>Atlas 800I A2 推理服务器</span></p>
</li><li><span id="ph11463114805219"><a name="ph11463114805219"></a><a name="ph11463114805219"></a>A200I A2 Box 异构组件</span></li></ul>
<p id="p0608021173520"><a name="p0608021173520"></a><a name="p0608021173520"></a></p>
</td>
</tr>
<tr id="row184616483311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2141721143510"><a name="p2141721143510"></a><a name="p2141721143510"></a>光模块</p>
<p id="p12670721113518"><a name="p12670721113518"></a><a name="p12670721113518"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p14141172183510"><a name="p14141172183510"></a><a name="p14141172183510"></a>npu_chip_optical_tx_power_X （X范围为0~3）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1614272163511"><a name="p1614272163511"></a><a name="p1614272163511"></a>光模块发送功率</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p94525931810"><a name="p94525931810"></a><a name="p94525931810"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p0144142193520"><a name="p0144142193520"></a><a name="p0144142193520"></a>单位：mW</p>
</td>
</tr>
<tr id="row1846416482311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p121531221173514"><a name="p121531221173514"></a><a name="p121531221173514"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p5153132118357"><a name="p5153132118357"></a><a name="p5153132118357"></a>npu_chip_optical_rx_power_X （X范围为0~3）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1515522153519"><a name="p1515522153519"></a><a name="p1515522153519"></a>光模块接收功率</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p93211216189"><a name="p93211216189"></a><a name="p93211216189"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p5156421163517"><a name="p5156421163517"></a><a name="p5156421163517"></a>单位：mW</p>
</td>
</tr>
<tr id="row20466104816315"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p616482113359"><a name="p616482113359"></a><a name="p616482113359"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p216452116359"><a name="p216452116359"></a><a name="p216452116359"></a>npu_chip_optical_vcc</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p416520214353"><a name="p416520214353"></a><a name="p416520214353"></a>光模块电压</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1580415142189"><a name="p1580415142189"></a><a name="p1580415142189"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p18167132113517"><a name="p18167132113517"></a><a name="p18167132113517"></a>单位：mV</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1317652112351"><a name="p1317652112351"></a><a name="p1317652112351"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p91765214359"><a name="p91765214359"></a><a name="p91765214359"></a>npu_chip_optical_temp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p817732119357"><a name="p817732119357"></a><a name="p817732119357"></a>光模块温度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p185571217201819"><a name="p185571217201819"></a><a name="p185571217201819"></a><a href="#table191895615241">标签1</a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p131797216355"><a name="p131797216355"></a><a name="p131797216355"></a>单位：摄氏度（℃）</p>
</td>
</tr>
</tbody>
</table>

**标签数据信息说明<a name="section129583351126"></a>**

关于以上表格中所用到的数据信息标签说明如下。

**表 11**  数据信息标签

<a name="table191895615241"></a>
<table><thead align="left"><tr id="row8191356112418"><th class="cellrowborder" valign="top" width="9.43%" id="mcps1.2.4.1.1"><p id="p1919205692417"><a name="p1919205692417"></a><a name="p1919205692417"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="45.45%" id="mcps1.2.4.1.2"><p id="p819155632413"><a name="p819155632413"></a><a name="p819155632413"></a>字段及说明</p>
</th>
<th class="cellrowborder" valign="top" width="45.12%" id="mcps1.2.4.1.3"><p id="p7213752020"><a name="p7213752020"></a><a name="p7213752020"></a>字段类型</p>
</th>
</tr>
</thead>
<tbody><tr id="row191916565240"><td class="cellrowborder" rowspan="7" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p id="p11251921132717"><a name="p11251921132717"></a><a name="p11251921132717"></a>标签1</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p id="p845273212520"><a name="p845273212520"></a><a name="p845273212520"></a>container_name：容器名</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p id="p445233232518"><a name="p445233232518"></a><a name="p445233232518"></a>string</p>
</td>
</tr>
<tr id="row15694518722"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p34521032132513"><a name="p34521032132513"></a><a name="p34521032132513"></a>id：NPU的ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1469410183214"><a name="p1469410183214"></a><a name="p1469410183214"></a>string</p>
</td>
</tr>
<tr id="row739383919217"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p43931839326"><a name="p43931839326"></a><a name="p43931839326"></a>model_name：昇腾AI处理器名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p143931839928"><a name="p143931839928"></a><a name="p143931839928"></a>string</p>
</td>
</tr>
<tr id="row196257269217"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p113052331628"><a name="p113052331628"></a><a name="p113052331628"></a>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p86252269215"><a name="p86252269215"></a><a name="p86252269215"></a>string</p>
</td>
</tr>
<tr id="row15353933334"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p204521232152511"><a name="p204521232152511"></a><a name="p204521232152511"></a>pcie_bus_info：昇腾AI处理器的PCIe信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p12353133317315"><a name="p12353133317315"></a><a name="p12353133317315"></a>string</p>
</td>
</tr>
<tr id="row155451138137"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p545219327256"><a name="p545219327256"></a><a name="p545219327256"></a>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p175458386311"><a name="p175458386311"></a><a name="p175458386311"></a>string</p>
</td>
</tr>
<tr id="row1235318335319"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1829345012515"><a name="p1829345012515"></a><a name="p1829345012515"></a>vdie_id：昇腾AI处理器唯一标识，可作为NPU的UUID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p173531633538"><a name="p173531633538"></a><a name="p173531633538"></a>string</p>
</td>
</tr>
<tr id="row3907925453"><td class="cellrowborder" rowspan="9" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p id="p610614317312"><a name="p610614317312"></a><a name="p610614317312"></a>标签2</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p id="p189077251559"><a name="p189077251559"></a><a name="p189077251559"></a>container_name：容器名，输出格式为“Pod Namespace_Pod名_容器名”。如果进程运行在宿主机上，该值为空。</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p id="p179072251510"><a name="p179072251510"></a><a name="p179072251510"></a>string</p>
</td>
</tr>
<tr id="row119071125657"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p2907225856"><a name="p2907225856"></a><a name="p2907225856"></a>container_id：容器ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p690720255518"><a name="p690720255518"></a><a name="p690720255518"></a>string</p>
</td>
</tr>
<tr id="row1290762517511"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p16907425855"><a name="p16907425855"></a><a name="p16907425855"></a>id：NPU的ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p17907225953"><a name="p17907225953"></a><a name="p17907225953"></a>string</p>
</td>
</tr>
<tr id="row16907162510513"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p890718251958"><a name="p890718251958"></a><a name="p890718251958"></a>model_name：<span id="ph151081914202916"><a name="ph151081914202916"></a><a name="ph151081914202916"></a>昇腾AI处理器</span>名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p390718251755"><a name="p390718251755"></a><a name="p390718251755"></a>string</p>
</td>
</tr>
<tr id="row1890714251357"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p79071725258"><a name="p79071725258"></a><a name="p79071725258"></a>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p590722513516"><a name="p590722513516"></a><a name="p590722513516"></a>string</p>
</td>
</tr>
<tr id="row149071925558"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p169073252051"><a name="p169073252051"></a><a name="p169073252051"></a>pcie_bus_info：<span id="ph1510917147299"><a name="ph1510917147299"></a><a name="ph1510917147299"></a>昇腾AI处理器</span>的PCIe信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p39074253512"><a name="p39074253512"></a><a name="p39074253512"></a>string</p>
</td>
</tr>
<tr id="row169072252520"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1390718251455"><a name="p1390718251455"></a><a name="p1390718251455"></a>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p189075251752"><a name="p189075251752"></a><a name="p189075251752"></a>string</p>
</td>
</tr>
<tr id="row1510613431434"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p41063431337"><a name="p41063431337"></a><a name="p41063431337"></a>vdie_id：<span id="ph3109101422915"><a name="ph3109101422915"></a><a name="ph3109101422915"></a>昇腾AI处理器</span>唯一标识，可作为NPU的UUID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p110615431735"><a name="p110615431735"></a><a name="p110615431735"></a>string</p>
</td>
</tr>
<tr id="row18906135864"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><div class="p" id="p191091814162910"><a name="p191091814162910"></a><a name="p191091814162910"></a>process_id:<a name="ul8109141402915"></a><a name="ul8109141402915"></a><ul id="ul8109141402915"><li><span id="ph4109014172916"><a name="ph4109014172916"></a><a name="ph4109014172916"></a>NPU Exporter</span>是特权容器下root用户启动的，查询到的PID是进程在宿主机上的PID</li><li>宿主机场景下启动，查询到的PID是进程在宿主机上的PID</li><li>其他容器场景，查询到的PID是进程在当前容器内的PID</li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p590711351168"><a name="p590711351168"></a><a name="p590711351168"></a>string</p>
</td>
</tr>
<tr id="row1988185619710"><td class="cellrowborder" rowspan="8" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p id="p434964731010"><a name="p434964731010"></a><a name="p434964731010"></a>标签3</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p id="p1798955617720"><a name="p1798955617720"></a><a name="p1798955617720"></a>aicore_count：vNPU核数</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p id="p20206191913160"><a name="p20206191913160"></a><a name="p20206191913160"></a>float</p>
</td>
</tr>
<tr id="row159891956371"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p298995620718"><a name="p298995620718"></a><a name="p298995620718"></a>container_name：容器名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p120611931615"><a name="p120611931615"></a><a name="p120611931615"></a>string</p>
</td>
</tr>
<tr id="row59891756178"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1298920561479"><a name="p1298920561479"></a><a name="p1298920561479"></a>id：NPU的ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p320616197163"><a name="p320616197163"></a><a name="p320616197163"></a>string</p>
</td>
</tr>
<tr id="row9989656471"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p14989556078"><a name="p14989556078"></a><a name="p14989556078"></a>model_name：昇腾AI处理器名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11206819191615"><a name="p11206819191615"></a><a name="p11206819191615"></a>string</p>
</td>
</tr>
<tr id="row1398911561279"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p598916561078"><a name="p598916561078"></a><a name="p598916561078"></a>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p620631910164"><a name="p620631910164"></a><a name="p620631910164"></a>string</p>
</td>
</tr>
<tr id="row13989956371"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p098985614716"><a name="p098985614716"></a><a name="p098985614716"></a>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1720631921611"><a name="p1720631921611"></a><a name="p1720631921611"></a>string</p>
</td>
</tr>
<tr id="row169894568712"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p64191230135414"><a name="p64191230135414"></a><a name="p64191230135414"></a>v_dev_id：vNPU唯一标识</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p62061419161618"><a name="p62061419161618"></a><a name="p62061419161618"></a>string</p>
</td>
</tr>
<tr id="row14907133516618"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1441953095414"><a name="p1441953095414"></a><a name="p1441953095414"></a>is_virtual：是否是虚拟设备</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1420741911167"><a name="p1420741911167"></a><a name="p1420741911167"></a>bool</p>
</td>
</tr>
<tr id="row10217101831"><td class="cellrowborder" rowspan="8" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p id="p19218160939"><a name="p19218160939"></a><a name="p19218160939"></a>标签4</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p id="p103704551136"><a name="p103704551136"></a><a name="p103704551136"></a>container_name：容器名</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p id="p237025512315"><a name="p237025512315"></a><a name="p237025512315"></a>string</p>
</td>
</tr>
<tr id="row162182007317"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p173707555318"><a name="p173707555318"></a><a name="p173707555318"></a>id：NPU的ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p113712551439"><a name="p113712551439"></a><a name="p113712551439"></a>string</p>
</td>
</tr>
<tr id="row12181801433"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p17371205512310"><a name="p17371205512310"></a><a name="p17371205512310"></a>model_name：<span id="ph1237110558312"><a name="ph1237110558312"></a><a name="ph1237110558312"></a>昇腾AI处理器</span>名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p6371125516312"><a name="p6371125516312"></a><a name="p6371125516312"></a>string</p>
</td>
</tr>
<tr id="row17218106313"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p6371055238"><a name="p6371055238"></a><a name="p6371055238"></a>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p6371175512318"><a name="p6371175512318"></a><a name="p6371175512318"></a>string</p>
</td>
</tr>
<tr id="row22181901733"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p15372195518312"><a name="p15372195518312"></a><a name="p15372195518312"></a>pcie_bus_info：<span id="ph73724551314"><a name="ph73724551314"></a><a name="ph73724551314"></a>昇腾AI处理器</span>的PCIe信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p437220551737"><a name="p437220551737"></a><a name="p437220551737"></a>string</p>
</td>
</tr>
<tr id="row1218110432"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p03721255839"><a name="p03721255839"></a><a name="p03721255839"></a>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p53720552037"><a name="p53720552037"></a><a name="p53720552037"></a>string</p>
</td>
</tr>
<tr id="row42181701639"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p2372955231"><a name="p2372955231"></a><a name="p2372955231"></a>vdie_id：<span id="ph1372195512316"><a name="ph1372195512316"></a><a name="ph1372195512316"></a>昇腾AI处理器</span>唯一标识，可作为NPU的UUID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p173738556310"><a name="p173738556310"></a><a name="p173738556310"></a>string</p>
</td>
</tr>
<tr id="row221813016314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><div class="p" id="p243419475489"><a name="p243419475489"></a><a name="p243419475489"></a>pcie_bw_type：向远端写PCIe带宽的统计值<a name="ul16373185519314"></a><a name="ul16373185519314"></a><ul id="ul16373185519314"><li>minPcieBw：最小值</li><li>maxPcieBw：最大值</li><li>avgPcieBw：平均值</li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p133731556314"><a name="p133731556314"></a><a name="p133731556314"></a>string</p>
</td>
</tr>
<tr id="row97376712319"><td class="cellrowborder" rowspan="9" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p id="p167371271038"><a name="p167371271038"></a><a name="p167371271038"></a>标签5</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p id="p13670533839"><a name="p13670533839"></a><a name="p13670533839"></a>container_name：容器名</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p id="p129282543417"><a name="p129282543417"></a><a name="p129282543417"></a>string</p>
</td>
</tr>
<tr id="row17737678310"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p15670153314318"><a name="p15670153314318"></a><a name="p15670153314318"></a>containerName：容器名，输出格式为“Pod Namespace_Pod名_容器名”</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1792810517343"><a name="p1792810517343"></a><a name="p1792810517343"></a>string</p>
</td>
</tr>
<tr id="row1473718719312"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p667018331312"><a name="p667018331312"></a><a name="p667018331312"></a>containerID：容器ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p179296511349"><a name="p179296511349"></a><a name="p179296511349"></a>string</p>
</td>
</tr>
<tr id="row1173713719317"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1467083311319"><a name="p1467083311319"></a><a name="p1467083311319"></a>npuID：NPU ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p199299511345"><a name="p199299511345"></a><a name="p199299511345"></a>string</p>
</td>
</tr>
<tr id="row127379715315"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p14670733636"><a name="p14670733636"></a><a name="p14670733636"></a>model_name：昇腾AI处理器名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p10930125193414"><a name="p10930125193414"></a><a name="p10930125193414"></a>string</p>
</td>
</tr>
<tr id="row4737871036"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p467073312313"><a name="p467073312313"></a><a name="p467073312313"></a>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p10930175143412"><a name="p10930175143412"></a><a name="p10930175143412"></a>string</p>
</td>
</tr>
<tr id="row20737197535"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1567018331316"><a name="p1567018331316"></a><a name="p1567018331316"></a>pcie_bus_info：昇腾AI处理器的PCIe信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5931953347"><a name="p5931953347"></a><a name="p5931953347"></a>string</p>
</td>
</tr>
<tr id="row6737671133"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p76708332313"><a name="p76708332313"></a><a name="p76708332313"></a>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p109311556343"><a name="p109311556343"></a><a name="p109311556343"></a>string</p>
</td>
</tr>
<tr id="row14170441549"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1467093319311"><a name="p1467093319311"></a><a name="p1467093319311"></a>vdie_id：昇腾AI处理器唯一标识，可作为NPU的UUID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p693218593417"><a name="p693218593417"></a><a name="p693218593417"></a>string</p>
</td>
</tr>
<tr id="row11396133454012"><td class="cellrowborder" rowspan="8" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p id="p127054914110"><a name="p127054914110"></a><a name="p127054914110"></a>标签6</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p id="p1170559184111"><a name="p1170559184111"></a><a name="p1170559184111"></a>container_name：容器名</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p id="p17705291418"><a name="p17705291418"></a><a name="p17705291418"></a>string</p>
</td>
</tr>
<tr id="row626813874011"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p18705093416"><a name="p18705093416"></a><a name="p18705093416"></a>id：NPU的ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1570549114116"><a name="p1570549114116"></a><a name="p1570549114116"></a>string</p>
</td>
</tr>
<tr id="row113218442408"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p6706394412"><a name="p6706394412"></a><a name="p6706394412"></a>model_name：昇腾AI处理器名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1670611944113"><a name="p1670611944113"></a><a name="p1670611944113"></a>string</p>
</td>
</tr>
<tr id="row17132944154018"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p9706139164110"><a name="p9706139164110"></a><a name="p9706139164110"></a>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1770610910412"><a name="p1770610910412"></a><a name="p1770610910412"></a>string</p>
</td>
</tr>
<tr id="row16436174954010"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p10706119144110"><a name="p10706119144110"></a><a name="p10706119144110"></a>pcie_bus_info：昇腾AI处理器的PCIe信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p0706169134114"><a name="p0706169134114"></a><a name="p0706169134114"></a>string</p>
</td>
</tr>
<tr id="row7436124934012"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p670619974112"><a name="p670619974112"></a><a name="p670619974112"></a>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p97065974117"><a name="p97065974117"></a><a name="p97065974117"></a>string</p>
</td>
</tr>
<tr id="row14371493400"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p4706129194114"><a name="p4706129194114"></a><a name="p4706129194114"></a>serial_number：昇腾AI处理器序列号，未配置时上报NA</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p12706159184120"><a name="p12706159184120"></a><a name="p12706159184120"></a>string</p>
</td>
</tr>
<tr id="row204371949104015"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1170617912416"><a name="p1170617912416"></a><a name="p1170617912416"></a>vdie_id：昇腾AI处理器唯一标识，可作为NPU的UUID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p137061791412"><a name="p137061791412"></a><a name="p137061791412"></a>string</p>
</td>
</tr>
<tr><td class="cellrowborder" rowspan="8" valign="top" width="9.43%" headers="mcps1.2.4.1.1 "><p>标签7</p>
</td>
<td class="cellrowborder" valign="top" width="45.45%" headers="mcps1.2.4.1.2 "><p>container_name：容器名</p>
</td>
<td class="cellrowborder" valign="top" width="45.12%" headers="mcps1.2.4.1.3 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>id：NPU的ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>model_name：昇腾AI处理器名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>namespace：命名空间名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>pcie_bus_info：昇腾AI处理器的PCIe信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>pod_name：Pod名</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>product_type：昇腾AI处理器产品形态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p>vdie_id：昇腾AI处理器唯一标识，可作为NPU的UUID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p>string</p>
</td>
</tr>
</tbody>
</table>

**调用的HDK接口<a name="section1137512020304"></a>**

NPU Exporter是通过调用底层的HDK接口，获取相应的信息。数据信息调用的HDK接口请参考[NPU Exporter调用的HDK接口.xlsx](../resource/NPU%20Exporter调用的HDK接口.xlsx)。查找数据信息对应的HDK接口，可参考如下步骤。

1.  登录[昇腾计算文档](https://support.huawei.com/enterprise/zh/category/ascend-computing-pid-1557196528909?submodel=doc)中心，选择单击对应产品名称，进入文档界面。例如Atlas 800I A2 推理服务器产品的用户，单击“Atlas 800I A2”。
2.  在左侧导航栏找到“二次开发”，根据接口的类型选择对应文档。
    -   DCMI接口选择“API参考”，单击进入《[DCMI API参考](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-hdk-pid-252764743?category=developer-documents&subcategory=api-reference)》。
    -   HCCN Tool接口选择“接口参考”，单击进入《[Atlas A2 中心推理和训练硬件 25.5.0 HCCN Tool 接口参考](https://support.huawei.com/enterprise/zh/doc/EDOC1100540101/426cffd9)》。

3.  在文档首页搜索栏中，直接搜索对应的接口名称或者关键词，获取接口的相关信息。

**状态码<a name="zh-cn_topic_0000001446964912_section1287016166169"></a>**

**表 12**  状态码

<a name="zh-cn_topic_0000001446964912_zh-cn_topic_0000001104815128_table10702170191419"></a>
|状态码|含义|
|--|--|
|200|正常状态。|
|307|临时跳转。|
|500|服务器内部错误。|


## Telegraf数据信息说明<a name="ZH-CN_TOPIC_0000002511426775"></a>

运行Telegraf后，会显示监测的昇腾AI处理器的数据信息，回显示例如下，仅供参考，以实际回显为准。数据信息的详细说明参见下文或[数据信息说明.xlsx](../resource/数据信息说明.xlsx)。

```
...
Ascend910-0,host=xxx  npu_chip_link_speed=104857600000i,npu_chip_roce_rx_cnp_pkt_num=0i,npu_chip_roce_unexpected_ack_num=0i,npu_chip_optical_vcc=3245.1,npu_chip_optical_rx_power_1=0.8585,npu_chip_info_hbm_used_memory=0i,npu_chip_mac_rx_pause_num=0i,npu_chip_roce_tx_all_pkt_num=0i,npu_chip_roce_tx_cnp_pkt_num=0i,npu_chip_info_temperature=46,npu_chip_mac_rx_bad_pkt_num=0i,npu_chip_roce_tx_err_pkt_num=0i,npu_chip_optical_rx_power_3=0.8466,npu_chip_optical_rx_power_0=0.7933,npu_chip_info_network_status=0i,npu_chip_mac_rx_pfc_pkt_num=0i,npu_chip_mac_tx_bad_pkt_num=0i,npu_chip_roce_rx_all_pkt_num=0i,npu_chip_mac_rx_bad_oct_num=0i,npu_chip_optical_tx_power_1=0.9162,npu_chip_info_utilization=0,npu_chip_info_power=73.9000015258789,npu_chip_info_link_status=1i,npu_chip_info_bandwidth_rx=0,npu_chip_mac_tx_pfc_pkt_num=0i,npu_chip_roce_rx_err_pkt_num=0i,npu_chip_roce_verification_err_num=0i,npu_chip_optical_state=1i,npu_chip_info_bandwidth_tx=0,npu_chip_mac_tx_bad_oct_num=0i,npu_chip_roce_out_of_order_num=0i,npu_chip_roce_qp_status_err_num=0i,npu_chip_optical_rx_power_2=0.855,npu_chip_optical_tx_power_0=0.9095,npu_chip_info_hbm_utilization=0,npu_chip_link_up_num=2i,npu_chip_info_health_status=1i,npu_chip_mac_tx_pause_num=0i,npu_chip_roce_new_pkt_rty_num=0i,npu_chip_optical_temp=53,npu_chip_optical_tx_power_2=1.0342,npu_chip_optical_tx_power_3=0.9715 1694772754612200641,npu_chip_info_process_info_num=0i
```

本接口支持查询默认指标组和自定义指标组。自定义指标组的方法详细请参见[自定义指标开发](../appendix.md#自定义指标开发)；默认指标组包含如下几个部分。指标组的采集和上报由配置文件中的开关控制，若开关配置为开启，则对应的指标组会进行采集和上报；若开关配置为关闭，则对应的指标组不会进行采集和上报。

-   [版本数据信息](#section170316521436141)
-   [NPU数据信息](#section1442282202316)
-   [vNPU数据信息](#section814111613432)
-   [Network数据信息](#section1358881214551)
-   [片上内存数据信息](#section177232045203114)
-   [HCCS数据信息](#section039816240252)
-   [PCIe数据信息](#section1240520241824136)
-   [RoCE数据信息](#section184516450323)
-   [SIO数据信息](#section7109037161515)
-   [光模块数据信息](#section1517163183510)
-   [DDR数据信息](#section114607361931169)

>[!NOTE] 说明 
>-   NPU Exporter是通过调用底层的HDK接口，获取相应的信息。数据信息调用的HDK接口请参考[调用的HDK接口](#section345820153363)。
>-   若查询某个数据信息时，NPU Exporter组件不支持该设备形态或调用HDK接口失败，则不会上报该数据信息。

**版本数据信息<a name="section170316521436141"></a>**

**表 1**  版本数据信息

<a name="table81981837143713"></a>
<table><thead align="left"><tr id="row319910378373"><th class="cellrowborder" valign="top" width="7.9399999999999995%" id="mcps1.2.6.1.1"><p id="p622884917372"><a name="p622884917372"></a><a name="p622884917372"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="25.03%" id="mcps1.2.6.1.2"><p id="p1322844963710"><a name="p1322844963710"></a><a name="p1322844963710"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.23%" id="mcps1.2.6.1.3"><p id="p822864983718"><a name="p822864983718"></a><a name="p822864983718"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="21.8%" id="mcps1.2.6.1.4"><p id="p15229649123712"><a name="p15229649123712"></a><a name="p15229649123712"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="24%" id="mcps1.2.6.1.5"><p id="p1023084933716"><a name="p1023084933716"></a><a name="p1023084933716"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row191991937123720"><td class="cellrowborder" valign="top" width="7.9399999999999995%" headers="mcps1.2.6.1.1 "><p id="p96501612386"><a name="p96501612386"></a><a name="p96501612386"></a>版本</p>
</td>
<td class="cellrowborder" valign="top" width="25.03%" headers="mcps1.2.6.1.2 "><p id="p36502183811"><a name="p36502183811"></a><a name="p36502183811"></a>npu_exporter_version_info</p>
</td>
<td class="cellrowborder" valign="top" width="21.23%" headers="mcps1.2.6.1.3 "><p id="p1665017112386"><a name="p1665017112386"></a><a name="p1665017112386"></a><span id="ph122556224122"><a name="ph122556224122"></a><a name="ph122556224122"></a>NPU Exporter</span>版本信息</p>
</td>
<td class="cellrowborder" valign="top" width="21.8%" headers="mcps1.2.6.1.4 "><p id="p11641734202218"><a name="p11641734202218"></a><a name="p11641734202218"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24%" headers="mcps1.2.6.1.5 "><p id="p10892195514317"><a name="p10892195514317"></a><a name="p10892195514317"></a>Atlas 训练系列产品</p>
<p id="p148921155154320"><a name="p148921155154320"></a><a name="p148921155154320"></a>Atlas A2 训练系列产品</p>
<p id="p1589214558432"><a name="p1589214558432"></a><a name="p1589214558432"></a>Atlas A3 训练系列产品</p>
<p id="p13892755104310"><a name="p13892755104310"></a><a name="p13892755104310"></a>推理服务器（插Atlas 300I 推理卡）</p>
<p id="p17892655184314"><a name="p17892655184314"></a><a name="p17892655184314"></a>Atlas 推理系列产品</p>
<p id="p08921555204319"><a name="p08921555204319"></a><a name="p08921555204319"></a><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span></p>
<p id="p1115563443"><a name="p1115563443"></a><a name="p1115563443"></a><span id="ph9892125584318"><a name="ph9892125584318"></a><a name="ph9892125584318"></a>A200I A2 Box 异构组件</span></p><p><span>推理服务器（插Atlas 350 标卡）</span></p>
</td>
</tr>
</tbody>
</table>

**NPU数据信息<a name="section1442282202316"></a>**

**表 2**  NPU数据信息

<a name="table18223172210289"></a>
<table><thead align="left"><tr id="row8223722122814"><th class="cellrowborder" valign="top" width="9.8%" id="mcps1.2.6.1.1"><p id="p14952185417289"><a name="p14952185417289"></a><a name="p14952185417289"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="24.279999999999998%" id="mcps1.2.6.1.2"><p id="p4953105472810"><a name="p4953105472810"></a><a name="p4953105472810"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.6.1.3"><p id="p11953954142813"><a name="p11953954142813"></a><a name="p11953954142813"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="19.68%" id="mcps1.2.6.1.4"><p id="p149531054142812"><a name="p149531054142812"></a><a name="p149531054142812"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="23.94%" id="mcps1.2.6.1.5"><p id="p14953115412284"><a name="p14953115412284"></a><a name="p14953115412284"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row1999144418467"><td class="cellrowborder" valign="top" width="9.8%" headers="mcps1.2.6.1.1 "><p id="p1549615614460"><a name="p1549615614460"></a><a name="p1549615614460"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.279999999999998%" headers="mcps1.2.6.1.2 "><p id="p949618565464"><a name="p949618565464"></a><a name="p949618565464"></a>machine_npu_nums</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p12496205674616"><a name="p12496205674616"></a><a name="p12496205674616"></a>昇腾AI处理器数目</p>
</td>
<td class="cellrowborder" valign="top" width="19.68%" headers="mcps1.2.6.1.4 "><p id="p19999944144617"><a name="p19999944144617"></a><a name="p19999944144617"></a>单位：个</p>
</td>
<td class="cellrowborder" rowspan="13" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><a name="ul1142611144613"></a><a name="ul1142611144613"></a><ul id="ul1142611144613"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 300I 推理卡）</li><li>Atlas 推理系列产品</li><li><span id="ph279972618380"><a name="ph279972618380"></a><a name="ph279972618380"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph1823654413571"><a name="ph1823654413571"></a><a name="ph1823654413571"></a>A200I A2 Box 异构组件</span></li><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
<p id="p72006546535"><a name="p72006546535"></a><a name="p72006546535"></a></p>
<p id="p4385659175311"><a name="p4385659175311"></a><a name="p4385659175311"></a></p>
<p id="p11457148195318"><a name="p11457148195318"></a><a name="p11457148195318"></a></p>
<p id="p17569728105411"><a name="p17569728105411"></a><a name="p17569728105411"></a></p>
<p id="p1523083720553"><a name="p1523083720553"></a><a name="p1523083720553"></a></p>
<p id="p1723003712552"><a name="p1723003712552"></a><a name="p1723003712552"></a></p>
<p id="p13230937195511"><a name="p13230937195511"></a><a name="p13230937195511"></a></p>
<p id="p152301237155517"><a name="p152301237155517"></a><a name="p152301237155517"></a></p>
<p id="p328564335513"><a name="p328564335513"></a><a name="p328564335513"></a></p>
<p id="p1913317507563"><a name="p1913317507563"></a><a name="p1913317507563"></a></p>
<p id="p1660195725612"><a name="p1660195725612"></a><a name="p1660195725612"></a></p>
<p id="p15286124311550"><a name="p15286124311550"></a><a name="p15286124311550"></a></p>
<p id="p18452227195416"><a name="p18452227195416"></a><a name="p18452227195416"></a></p>
<p id="p745242713549"><a name="p745242713549"></a><a name="p745242713549"></a></p>
<p id="p34521527205414"><a name="p34521527205414"></a><a name="p34521527205414"></a></p>
<p id="p19452027175413"><a name="p19452027175413"></a><a name="p19452027175413"></a></p>
<p id="p1945262715417"><a name="p1945262715417"></a><a name="p1945262715417"></a></p>
<p id="p156813447544"><a name="p156813447544"></a><a name="p156813447544"></a></p>
<p id="p589672265514"><a name="p589672265514"></a><a name="p589672265514"></a></p>
</td>
</tr>
<tr id="row1883144115448"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p18433255105719"><a name="p18433255105719"></a><a name="p18433255105719"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p243318551574"><a name="p243318551574"></a><a name="p243318551574"></a>npu_chip_info_name</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p843414559571"><a name="p843414559571"></a><a name="p843414559571"></a><span id="ph543465525716"><a name="ph543465525716"></a><a name="ph543465525716"></a>昇腾AI处理器</span>名称和ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1831841114417"><a name="p1831841114417"></a><a name="p1831841114417"></a>-</p>
</td>
</tr>
<tr id="row1320035495310"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12169113820544"><a name="p12169113820544"></a><a name="p12169113820544"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p111691338155414"><a name="p111691338155414"></a><a name="p111691338155414"></a>npu_chip_info_health_status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p20169193819546"><a name="p20169193819546"></a><a name="p20169193819546"></a><span id="ph1169163818542"><a name="ph1169163818542"></a><a name="ph1169163818542"></a><span id="ph11169203811544"><a name="ph11169203811544"></a><a name="ph11169203811544"></a>昇腾AI处理器</span></span>健康状态。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p316933865416"><a name="p316933865416"></a><a name="p316933865416"></a>取值为0或</p>
<a name="ul1016913885417"></a><a name="ul1016913885417"></a><ul id="ul1016913885417"><li>1：健康</li><li>0：不健康</li></ul>
</td>
</tr>
<tr id="row10385175935311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p316943810545"><a name="p316943810545"></a><a name="p316943810545"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p116913381548"><a name="p116913381548"></a><a name="p116913381548"></a>npu_chip_info_power</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3169638165419"><a name="p3169638165419"></a><a name="p3169638165419"></a><span id="ph10169183819548"><a name="ph10169183819548"></a><a name="ph10169183819548"></a><span id="ph1616953865415"><a name="ph1616953865415"></a><a name="ph1616953865415"></a>昇腾AI处理器</span></span>功耗。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1516963825410"><a name="p1516963825410"></a><a name="p1516963825410"></a>单位：瓦特（W）</p>
</td>
</tr>
<tr id="row556922812544"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p13169838125413"><a name="p13169838125413"></a><a name="p13169838125413"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p71693385547"><a name="p71693385547"></a><a name="p71693385547"></a>npu_chip_info_temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11169113815412"><a name="p11169113815412"></a><a name="p11169113815412"></a><span id="ph216910388546"><a name="ph216910388546"></a><a name="ph216910388546"></a><span id="ph1716913384544"><a name="ph1716913384544"></a><a name="ph1716913384544"></a>昇腾AI处理器</span></span>温度。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1717023895414"><a name="p1717023895414"></a><a name="p1717023895414"></a>单位：摄氏度（℃）</p>
</td>
</tr>
<tr id="row922903714555"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p8470111915614"><a name="p8470111915614"></a><a name="p8470111915614"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p9470151955618"><a name="p9470151955618"></a><a name="p9470151955618"></a>第一个错误码为：npu_chip_info_error_code</p>
<p id="p1658317311118"><a name="p1658317311118"></a><a name="p1658317311118"></a>其他错误码：npu_chip_info_error_code_X</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p547061925610"><a name="p547061925610"></a><a name="p547061925610"></a><span id="ph1415520117467"><a name="ph1415520117467"></a><a name="ph1415520117467"></a>昇腾AI处理器</span>错误码。</p>
<p id="p2011915418141"><a name="p2011915418141"></a><a name="p2011915418141"></a>当昇腾AI处理器上没有错误码时，不会上报该字段。</p>
<div class="note" id="note71551511134616"><a name="note71551511134616"></a><a name="note71551511134616"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="ul01551011174616"></a><a name="ul01551011174616"></a><ul id="ul01551011174616"><li>Prometheus场景：若该<span id="ph4155141154620"><a name="ph4155141154620"></a><a name="ph4155141154620"></a>昇腾AI处理器</span>上同时存在多个错误码，由于Prometheus格式限制，当前只支持上报前十个出现的错误码。X的取值范围：1~9</li><li>Telegraf场景：最多支持上报128个错误码。</li><li>错误码的详细说明，请参见<a href="../appendix.md#芯片故障码参考文档">芯片故障码参考文档</a>章节。</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p2470171915560"><a name="p2470171915560"></a><a name="p2470171915560"></a>-</p>
</td>
</tr>
<tr id="row19230153714559"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p74701519145618"><a name="p74701519145618"></a><a name="p74701519145618"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14708196563"><a name="p14708196563"></a><a name="p14708196563"></a>npu_chip_info_process_info_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10470101985611"><a name="p10470101985611"></a><a name="p10470101985611"></a>占用<span id="ph13471719205614"><a name="ph13471719205614"></a><a name="ph13471719205614"></a>昇腾AI处理器</span>的进程数量。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17471219115612"><a name="p17471219115612"></a><a name="p17471219115612"></a>-</p>
</td>
</tr>
<tr id="row19230203715555"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4471191985612"><a name="p4471191985612"></a><a name="p4471191985612"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1047181985617"><a name="p1047181985617"></a><a name="p1047181985617"></a>npu_chip_info_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p447112192567"><a name="p447112192567"></a><a name="p447112192567"></a><span id="ph7471111975618"><a name="ph7471111975618"></a><a name="ph7471111975618"></a><span id="ph3471141910562"><a name="ph3471141910562"></a><a name="ph3471141910562"></a>昇腾AI处理器</span></span>的AI Core利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p247131915566"><a name="p247131915566"></a><a name="p247131915566"></a>单位：%</p>
</td>
</tr>
<tr id="row15230837165519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4471171935616"><a name="p4471171935616"></a><a name="p4471171935616"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p15471161918560"><a name="p15471161918560"></a><a name="p15471161918560"></a>npu_chip_info_aicore_current_freq</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p19471819155615"><a name="p19471819155615"></a><a name="p19471819155615"></a>昇腾AI处理器的AI Core当前频率</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p104711319115610"><a name="p104711319115610"></a><a name="p104711319115610"></a>单位：MHz</p>
</td>
</tr>
<tr id="row13285243115515"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1047112197565"><a name="p1047112197565"></a><a name="p1047112197565"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8471161912564"><a name="p8471161912564"></a><a name="p8471161912564"></a>npu_chip_info_process_info</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p847161945616"><a name="p847161945616"></a><a name="p847161945616"></a>占用昇腾AI处理器进程的信息，</p>
<p id="p647171965610"><a name="p647171965610"></a><a name="p647171965610"></a>仅当没有进程占用昇腾AI处理器时上报，值为0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16471519205612"><a name="p16471519205612"></a><a name="p16471519205612"></a>单位：MB</p>
</td>
</tr>
<tr id="row1013365095611"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1238710217578"><a name="p1238710217578"></a><a name="p1238710217578"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1838716285712"><a name="p1838716285712"></a><a name="p1838716285712"></a>npu_chip_info_process_info_PID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1938722125716"><a name="p1938722125716"></a><a name="p1938722125716"></a>占用昇腾AI处理器进程的信息，其中PID为进程在宿主机上的PID；取值为进程使用的内存。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p113876216576"><a name="p113876216576"></a><a name="p113876216576"></a>单位：MB</p>
</td>
</tr>
<tr id="row18601657145611"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1738710215716"><a name="p1738710215716"></a><a name="p1738710215716"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p16387025573"><a name="p16387025573"></a><a name="p16387025573"></a>npu_chip_info_voltage</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p93871624575"><a name="p93871624575"></a><a name="p93871624575"></a>昇腾AI处理器电压</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p83871222579"><a name="p83871222579"></a><a name="p83871222579"></a>单位：伏特（V）</p>
</td>
</tr>
<tr id="row18962022155516"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1376232724113"><a name="p1376232724113"></a><a name="p1376232724113"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p197621927104119"><a name="p197621927104119"></a><a name="p197621927104119"></a>npu_chip_info_serial_number</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11762142754113"><a name="p11762142754113"></a><a name="p11762142754113"></a>昇腾AI处理器序列号</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p158961122195514"><a name="p158961122195514"></a><a name="p158961122195514"></a>-</p>
</td>
</tr>
<tr id="row9224102218282"><td class="cellrowborder" valign="top" width="9.8%" headers="mcps1.2.6.1.1 "><p id="p1095345410280"><a name="p1095345410280"></a><a name="p1095345410280"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.279999999999998%" headers="mcps1.2.6.1.2 "><p id="p209531154112810"><a name="p209531154112810"></a><a name="p209531154112810"></a>npu_chip_info_network_status</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p395385413282"><a name="p395385413282"></a><a name="p395385413282"></a><span id="ph1395316544284"><a name="ph1395316544284"></a><a name="ph1395316544284"></a><span id="ph3953105462815"><a name="ph3953105462815"></a><a name="ph3953105462815"></a>昇腾AI处理器</span>的网络健康状态</span>。</p>
</td>
<td class="cellrowborder" valign="top" width="19.68%" headers="mcps1.2.6.1.4 "><p id="p169536543281"><a name="p169536543281"></a><a name="p169536543281"></a>取值为0或1</p>
<a name="ul1695355411281"></a><a name="ul1695355411281"></a><ul id="ul1695355411281"><li>1：健康，可以连通</li><li>0：不健康，无法连通</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><a name="ul195418540284"></a><a name="ul195418540284"></a><ul id="ul195418540284"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph79376229499"><a name="ph79376229499"></a><a name="ph79376229499"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph288515314573"><a name="ph288515314573"></a><a name="ph288515314573"></a>A200I A2 Box 异构组件</span></li></ul>
</td>
</tr>
<tr id="row44536494526"><td class="cellrowborder" valign="top" width="9.8%" headers="mcps1.2.6.1.1 "><p id="p96351458115217"><a name="p96351458115217"></a><a name="p96351458115217"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.279999999999998%" headers="mcps1.2.6.1.2 "><p id="p8635358125213"><a name="p8635358125213"></a><a name="p8635358125213"></a>npu_chip_info_overall_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p13635155814528"><a name="p13635155814528"></a><a name="p13635155814528"></a>昇腾AI处理器整体利用率</p>
</td>
<td class="cellrowborder" valign="top" width="19.68%" headers="mcps1.2.6.1.4 "><p id="p36351558135211"><a name="p36351558135211"></a><a name="p36351558135211"></a>单位：%</p>
</td>
<td class="cellrowborder" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><a name="ul11480102917516"></a><a name="ul11480102917516"></a><ul id="ul11480102917516"><li>Atlas A2 训练系列产品</li></ul>
<a name="ul0480112918517"></a><a name="ul0480112918517"></a><ul id="ul0480112918517"><li>Atlas A3 训练系列产品</li></ul>
<a name="ul648022945116"></a><a name="ul648022945116"></a><ul id="ul648022945116"><li><span id="ph178711115212"><a name="ph178711115212"></a><a name="ph178711115212"></a>Atlas 800I A2 推理服务器</span></li></ul>
<a name="ul148152911516"></a><a name="ul148152911516"></a><ul id="ul148152911516"><li><span id="ph1848152919517"><a name="ph1848152919517"></a><a name="ph1848152919517"></a>A200I A2 Box 异构组件</span></li><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
</td>
</tr>
<tr id="row0457848145313"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p122885821014"><a name="p122885821014"></a><a name="p122885821014"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1628915881020"><a name="p1628915881020"></a><a name="p1628915881020"></a>npu_chip_info_vector_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1716943814545"><a name="p1716943814545"></a><a name="p1716943814545"></a><span id="ph155021541122118"><a name="ph155021541122118"></a><a name="ph155021541122118"></a>昇腾AI处理器</span>AI Vector利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p150244119218"><a name="p150244119218"></a><a name="p150244119218"></a>单位：%</p>
</td>
<td><ul><li>Atlas A2 训练系列产品</li></ul>
<ul><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 300I 推理卡）</li></ul>
<ul><li>Atlas 推理系列产品</li><li><span>Atlas 800I A2 推理服务器</span></li></ul>
<ul><li><span>A200I A2 Box 异构组件</span></li><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
</td>
</tr>
<tr><td><p>NPU</p>
</td>
<td><p>npu_chip_info_product_type</p>
</td>
<td><p>昇腾AI处理器产品形态</p>
</td>
<td><p>1：占位字符，无实际含义</p>
</td>
<td><p>Atlas 推理系列产品</p>
</td>
</tr>
<tr><td><p>NPU</p>
</td>
<td><p>npu_chip_info_cube_utilization</p>
</td>
<td><p>昇腾AI处理器AI Cube利用率</p>
</td>
<td><p>单位：%</p>
</td>
<td><ul><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li>Atlas 800I A2 推理服务器</li><li>A200I A2 Box 异构组件</li></ul>
</td>
</tr>
</tbody>
</table>

**vNPU数据信息<a name="section814111613432"></a>**

**表 3**  vNPU数据信息

<a name="table176992573417"></a>
<table><thead align="left"><tr id="row147006579418"><th class="cellrowborder" valign="top" width="9.8009800980098%" id="mcps1.2.6.1.1"><p id="p1380213250210"><a name="p1380213250210"></a><a name="p1380213250210"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="28.822882288228826%" id="mcps1.2.6.1.2"><p id="p580352532112"><a name="p580352532112"></a><a name="p580352532112"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="25.182518251825186%" id="mcps1.2.6.1.3"><p id="p280372562117"><a name="p280372562117"></a><a name="p280372562117"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="12.55125512551255%" id="mcps1.2.6.1.4"><p id="p1580317257217"><a name="p1580317257217"></a><a name="p1580317257217"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="23.642364236423642%" id="mcps1.2.6.1.5"><p id="p0803132522113"><a name="p0803132522113"></a><a name="p0803132522113"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row1470011579414"><td class="cellrowborder" valign="top" width="9.8009800980098%" headers="mcps1.2.6.1.1 "><p id="p0803142511215"><a name="p0803142511215"></a><a name="p0803142511215"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" width="28.822882288228826%" headers="mcps1.2.6.1.2 "><p id="p78031825162113"><a name="p78031825162113"></a><a name="p78031825162113"></a>vnpu_pod_aicore_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="25.182518251825186%" headers="mcps1.2.6.1.3 "><p id="p7803225132114"><a name="p7803225132114"></a><a name="p7803225132114"></a>vNPU的AI Core利用率</p>
</td>
<td class="cellrowborder" valign="top" width="12.55125512551255%" headers="mcps1.2.6.1.4 "><p id="p480452542116"><a name="p480452542116"></a><a name="p480452542116"></a>单位：%</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.642364236423642%" headers="mcps1.2.6.1.5 "><p id="p553617193312"><a name="p553617193312"></a><a name="p553617193312"></a><span id="ph19590185162111"><a name="ph19590185162111"></a><a name="ph19590185162111"></a>Atlas 推理系列产品</span></p>
<p id="p5979185811548"><a name="p5979185811548"></a><a name="p5979185811548"></a></p>
<p id="p1272129165519"><a name="p1272129165519"></a><a name="p1272129165519"></a></p>
<p id="p1666712017555"><a name="p1666712017555"></a><a name="p1666712017555"></a></p>
<p id="p17841125753817"><a name="p17841125753817"></a><a name="p17841125753817"></a></p>
</td>
</tr>
<tr id="row17703155715411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7804152519214"><a name="p7804152519214"></a><a name="p7804152519214"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p198041725182114"><a name="p198041725182114"></a><a name="p198041725182114"></a>vnpu_pod_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p380482517215"><a name="p380482517215"></a><a name="p380482517215"></a>vNPU拥有的总内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1880432511212"><a name="p1880432511212"></a><a name="p1880432511212"></a>单位：KB</p>
</td>
</tr>
<tr id="row20706115712414"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1180492520213"><a name="p1180492520213"></a><a name="p1180492520213"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5805925152119"><a name="p5805925152119"></a><a name="p5805925152119"></a>vnpu_pod_used_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p128051725192114"><a name="p128051725192114"></a><a name="p128051725192114"></a>vNPU使用中的内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p108050258215"><a name="p108050258215"></a><a name="p108050258215"></a>单位：KB</p>
</td>
</tr>
</tbody>
</table>

**Network数据信息<a name="section1358881214551"></a>**

**表 4**  Network数据信息

<a name="table133306180110"></a>
<table><thead align="left"><tr id="row1033041814119"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.1"><p id="p9187428181111"><a name="p9187428181111"></a><a name="p9187428181111"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.2"><p id="p18187728181116"><a name="p18187728181116"></a><a name="p18187728181116"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.3"><p id="p1118742891112"><a name="p1118742891112"></a><a name="p1118742891112"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.4"><p id="p9187152818113"><a name="p9187152818113"></a><a name="p9187152818113"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.5"><p id="p124519421113"><a name="p124519421113"></a><a name="p124519421113"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row103308189112"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.1 "><p id="p141871428111110"><a name="p141871428111110"></a><a name="p141871428111110"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.2 "><p id="p718722816112"><a name="p718722816112"></a><a name="p718722816112"></a>npu_chip_info_bandwidth_rx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="p14187162841116"><a name="p14187162841116"></a><a name="p14187162841116"></a><span id="ph7187192814119"><a name="ph7187192814119"></a><a name="ph7187192814119"></a><span id="ph4187182811114"><a name="ph4187182811114"></a><a name="ph4187182811114"></a>昇腾AI处理器</span></span>的网口实时接收速率。</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.4 "><p id="p19187928171115"><a name="p19187928171115"></a><a name="p19187928171115"></a>单位：MB/s</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="20%" headers="mcps1.2.6.1.5 "><a name="ul19245194241111"></a><a name="ul19245194241111"></a><ul id="ul19245194241111"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph206765471804"><a name="ph206765471804"></a><a name="ph206765471804"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph6245242151119"><a name="ph6245242151119"></a><a name="ph6245242151119"></a>A200I A2 Box 异构组件</span></li></ul>
</td>
</tr>
<tr id="row93303180110"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p19187142811114"><a name="p19187142811114"></a><a name="p19187142811114"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p01871528121114"><a name="p01871528121114"></a><a name="p01871528121114"></a>npu_chip_info_bandwidth_tx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p21872283116"><a name="p21872283116"></a><a name="p21872283116"></a><span id="ph318772811113"><a name="ph318772811113"></a><a name="ph318772811113"></a><span id="ph141873288114"><a name="ph141873288114"></a><a name="ph141873288114"></a>昇腾AI处理器</span></span>的网口实时发送速率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p191872283111"><a name="p191872283111"></a><a name="p191872283111"></a>单位：MB/s</p>
</td>
</tr>
<tr id="row133011181117"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p181872028151119"><a name="p181872028151119"></a><a name="p181872028151119"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12187122801111"><a name="p12187122801111"></a><a name="p12187122801111"></a>npu_chip_info_link_status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1187142891114"><a name="p1187142891114"></a><a name="p1187142891114"></a><span id="ph91871728141110"><a name="ph91871728141110"></a><a name="ph91871728141110"></a><span id="ph7187132813117"><a name="ph7187132813117"></a><a name="ph7187132813117"></a>昇腾AI处理器</span></span>的网口Link状态。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1118714283116"><a name="p1118714283116"></a><a name="p1118714283116"></a>取值为0或1</p>
<a name="ul5187128201110"></a><a name="ul5187128201110"></a><ul id="ul5187128201110"><li>1：UP</li><li>0：DOWN</li></ul>
</td>
</tr>
<tr id="row7330418111118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p10187122819116"><a name="p10187122819116"></a><a name="p10187122819116"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p218782813114"><a name="p218782813114"></a><a name="p218782813114"></a>npu_chip_link_speed</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p618712861114"><a name="p618712861114"></a><a name="p618712861114"></a><span id="ph15187192831111"><a name="ph15187192831111"></a><a name="ph15187192831111"></a><span id="ph1187102851116"><a name="ph1187102851116"></a><a name="ph1187102851116"></a>昇腾AI处理器</span></span>网口默认速率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p61871028181113"><a name="p61871028181113"></a><a name="p61871028181113"></a>单位：MB/s</p>
</td>
</tr>
<tr id="row1133111810118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p11188142815112"><a name="p11188142815112"></a><a name="p11188142815112"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p418852851110"><a name="p418852851110"></a><a name="p418852851110"></a>npu_chip_link_up_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7188142817112"><a name="p7188142817112"></a><a name="p7188142817112"></a><span id="ph1618852818116"><a name="ph1618852818116"></a><a name="ph1618852818116"></a><span id="ph141881428161117"><a name="ph141881428161117"></a><a name="ph141881428161117"></a>昇腾AI处理器</span></span>网口UP的统计次数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p018822815114"><a name="p018822815114"></a><a name="p018822815114"></a>单位：次</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_status_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口Link状态。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>取值为0或1</p><ul><li>1：UP</li><li>0：DOWN</li></ul>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="19.91%" headers="mcps1.1.7.1.6 "><p><span>推理服务器（插Atlas 350 标卡 4Pmesh互联）</span></p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口实时接收速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。使用的-time参数为100。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口实时发送速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。使用的-time参数为100。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetWorkA5</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_speed_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>物理端口的速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：G</p>
</td>
</tr>
</tbody>
</table>

**片上内存数据信息<a name="section177232045203114"></a>**

**表 5**  片上内存数据信息

<a name="table728745315300"></a>
<table><thead align="left"><tr id="row72881853103019"><th class="cellrowborder" valign="top" width="9.21%" id="mcps1.2.6.1.1"><p id="p11126227312"><a name="p11126227312"></a><a name="p11126227312"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="26.97%" id="mcps1.2.6.1.2"><p id="p21121225312"><a name="p21121225312"></a><a name="p21121225312"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.6.1.3"><p id="p11123222316"><a name="p11123222316"></a><a name="p11123222316"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="14.91%" id="mcps1.2.6.1.4"><p id="p17112182293120"><a name="p17112182293120"></a><a name="p17112182293120"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="26.61%" id="mcps1.2.6.1.5"><p id="p71121822153119"><a name="p71121822153119"></a><a name="p71121822153119"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row1288953163016"><td class="cellrowborder" valign="top" width="9.21%" headers="mcps1.2.6.1.1 "><p id="p537581413115"><a name="p537581413115"></a><a name="p537581413115"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" width="26.97%" headers="mcps1.2.6.1.2 "><p id="p1637571420317"><a name="p1637571420317"></a><a name="p1637571420317"></a>npu_chip_info_hbm_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p1837671483116"><a name="p1837671483116"></a><a name="p1837671483116"></a><span id="ph837613143312"><a name="ph837613143312"></a><a name="ph837613143312"></a><span id="ph3376131453120"><a name="ph3376131453120"></a><a name="ph3376131453120"></a>昇腾AI处理器</span></span>的片上内存已使用量。</p>
</td>
<td class="cellrowborder" valign="top" width="14.91%" headers="mcps1.2.6.1.4 "><p id="p1376214173120"><a name="p1376214173120"></a><a name="p1376214173120"></a>单位：MB</p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="26.61%" headers="mcps1.2.6.1.5 "><a name="ul1737721403120"></a><a name="ul1737721403120"></a><ul id="ul1737721403120"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph043025116483"><a name="ph043025116483"></a><a name="ph043025116483"></a>A200I A2 Box 异构组件</span></li><li><span id="ph1201913534"><a name="ph1201913534"></a><a name="ph1201913534"></a>Atlas 800I A2 推理服务器</span></li></ul><ul><li><span>推理服务器（插Atlas 350 标卡）</span></li></ul>
<p id="p744164714207"><a name="p744164714207"></a><a name="p744164714207"></a></p>
<p id="p1642204722019"><a name="p1642204722019"></a><a name="p1642204722019"></a></p>
</td>
</tr>
<tr id="row8376151714335"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p930919134154"><a name="p930919134154"></a><a name="p930919134154"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13687292332"><a name="p13687292332"></a><a name="p13687292332"></a>npu_chip_info_hbm_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15368729183317"><a name="p15368729183317"></a><a name="p15368729183317"></a>昇腾AI处理器片上总内存。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1537617171338"><a name="p1537617171338"></a><a name="p1537617171338"></a>单位：MB</p>
</td>
</tr>
<tr id="row152881353163011"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1937671413114"><a name="p1937671413114"></a><a name="p1937671413114"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p137614149311"><a name="p137614149311"></a><a name="p137614149311"></a>npu_chip_info_hbm_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1737631433115"><a name="p1737631433115"></a><a name="p1737631433115"></a><span id="ph143778147311"><a name="ph143778147311"></a><a name="ph143778147311"></a><span id="ph10377161417316"><a name="ph10377161417316"></a><a name="ph10377161417316"></a>昇腾AI处理器</span></span>的片上内存利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1337761412315"><a name="p1337761412315"></a><a name="p1337761412315"></a>单位：%</p>
</td>
</tr>
<tr id="row1944471872714"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1899921617152"><a name="p1899921617152"></a><a name="p1899921617152"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p208411785402"><a name="p208411785402"></a><a name="p208411785402"></a>npu_chip_info_hbm_ecc_enable_flag</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p12841489405"><a name="p12841489405"></a><a name="p12841489405"></a>昇腾AI处理器片上内存的ECC使能状态。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1898013114817"><a name="p1898013114817"></a><a name="p1898013114817"></a>取值为1或0</p>
<a name="ul089881311484"></a><a name="ul089881311484"></a><ul id="ul089881311484"><li>0：ECC检测未使能</li><li>1：ECC检测使能</li></ul>
</td>
</tr>
<tr id="row1713612195407"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p11591181911516"><a name="p11591181911516"></a><a name="p11591181911516"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1149010582404"><a name="p1149010582404"></a><a name="p1149010582404"></a>npu_chip_info_hbm_ecc_single_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1449045884011"><a name="p1449045884011"></a><a name="p1449045884011"></a>昇腾AI处理器片上内存单比特当前错误计数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p21361419184019"><a name="p21361419184019"></a><a name="p21361419184019"></a>-</p>
</td>
</tr>
<tr id="row51361219184015"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5700142221516"><a name="p5700142221516"></a><a name="p5700142221516"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5608101654112"><a name="p5608101654112"></a><a name="p5608101654112"></a>npu_chip_info_hbm_ecc_double_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p160871694114"><a name="p160871694114"></a><a name="p160871694114"></a>昇腾AI处理器片上内存多比特当前错误计数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9136181915405"><a name="p9136181915405"></a><a name="p9136181915405"></a>-</p>
</td>
</tr>
<tr id="row41361919134011"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12199426201519"><a name="p12199426201519"></a><a name="p12199426201519"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p65349266419"><a name="p65349266419"></a><a name="p65349266419"></a>npu_chip_info_hbm_ecc_total_single_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p753442694119"><a name="p753442694119"></a><a name="p753442694119"></a>昇腾AI处理器片上内存生命周期内所有单比特错误数量。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p413618196403"><a name="p413618196403"></a><a name="p413618196403"></a>-</p>
</td>
</tr>
<tr id="row14137719124010"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p101371819194013"><a name="p101371819194013"></a><a name="p101371819194013"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p101041130121610"><a name="p101041130121610"></a><a name="p101041130121610"></a>npu_chip_info_hbm_ecc_total_double_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1810412305163"><a name="p1810412305163"></a><a name="p1810412305163"></a>昇腾AI处理器片上内存生命周期内所有多比特错误数量。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p161377192404"><a name="p161377192404"></a><a name="p161377192404"></a>-</p>
</td>
</tr>
<tr id="row65841633144013"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p155841733204015"><a name="p155841733204015"></a><a name="p155841733204015"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p37841238171612"><a name="p37841238171612"></a><a name="p37841238171612"></a>npu_chip_info_hbm_ecc_single_bit_isolated_pages_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9784173861615"><a name="p9784173861615"></a><a name="p9784173861615"></a>昇腾AI处理器片上内存单比特错误隔离内存页数量。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p14584333134019"><a name="p14584333134019"></a><a name="p14584333134019"></a>-</p>
</td>
</tr>
<tr id="row15584173374015"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p11194724184"><a name="p11194724184"></a><a name="p11194724184"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p41947271811"><a name="p41947271811"></a><a name="p41947271811"></a>npu_chip_info_hbm_ecc_double_bit_isolated_pages_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p819417291813"><a name="p819417291813"></a><a name="p819417291813"></a>昇腾AI处理器片上内存多比特错误隔离内存页数量。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p6584633144018"><a name="p6584633144018"></a><a name="p6584633144018"></a>-</p>
</td>
</tr>
<tr id="row838319296204"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16912264198"><a name="p16912264198"></a><a name="p16912264198"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5691826121920"><a name="p5691826121920"></a><a name="p5691826121920"></a>npu_chip_info_hbm_temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p12691112601911"><a name="p12691112601911"></a><a name="p12691112601911"></a>昇腾AI处理器片上内存的温度。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1871185411812"><a name="p1871185411812"></a><a name="p1871185411812"></a>单位：&deg;C</p>
</td>
</tr>
<tr id="row6486636112012"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p36914261197"><a name="p36914261197"></a><a name="p36914261197"></a>片上内存</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5691142631920"><a name="p5691142631920"></a><a name="p5691142631920"></a>npu_chip_info_hbm_bandwidth_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p86911726101915"><a name="p86911726101915"></a><a name="p86911726101915"></a>昇腾AI处理器片上内存的带宽利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1130779198"><a name="p1130779198"></a><a name="p1130779198"></a>单位：%</p>
</td>
</tr>
</tbody>
</table>

**HCCS数据信息<a name="section039816240252"></a>**

**表 6**  HCCS数据信息

<a name="table9399845122516"></a>
<table><thead align="left"><tr id="row153998454253"><th class="cellrowborder" valign="top" width="8.950000000000001%" id="mcps1.2.6.1.1"><p id="p94879162267"><a name="p94879162267"></a><a name="p94879162267"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="28.190000000000005%" id="mcps1.2.6.1.2"><p id="p10487151652611"><a name="p10487151652611"></a><a name="p10487151652611"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.990000000000006%" id="mcps1.2.6.1.3"><p id="p1648711682615"><a name="p1648711682615"></a><a name="p1648711682615"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.570000000000004%" id="mcps1.2.6.1.4"><p id="p174873165265"><a name="p174873165265"></a><a name="p174873165265"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="25.300000000000004%" id="mcps1.2.6.1.5"><p id="p0487161620269"><a name="p0487161620269"></a><a name="p0487161620269"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row2040184552518"><td class="cellrowborder" valign="top" width="8.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p14401145172516"><a name="p14401145172516"></a><a name="p14401145172516"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" width="28.190000000000005%" headers="mcps1.2.6.1.2 "><p id="p338114141712"><a name="p338114141712"></a><a name="p338114141712"></a>npu_chip_info_hccs_statistic_info_tx_cnt_X</p>
<p id="p1024171512455"><a name="p1024171512455"></a><a name="p1024171512455"></a>X范围：1~7（Atlas A2 训练系列产品或Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" width="21.990000000000006%" headers="mcps1.2.6.1.3 "><a name="ul1424913612438"></a><a name="ul1424913612438"></a><ul id="ul1424913612438"><li>第X个HDLC链路发送报文数，单位是flit。</li><li>采集失败时上报-1。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="15.570000000000004%" headers="mcps1.2.6.1.4 "><p id="p840184516254"><a name="p840184516254"></a><a name="p840184516254"></a>-</p>
</td>
<td class="cellrowborder" rowspan="8" valign="top" width="25.300000000000004%" headers="mcps1.2.6.1.5 "><a name="ul11925372813"></a><a name="ul11925372813"></a><ul id="ul11925372813"><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li></ul>
</td>
</tr>
<tr id="row1140184517258"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1240115459259"><a name="p1240115459259"></a><a name="p1240115459259"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p186021875020"><a name="p186021875020"></a><a name="p186021875020"></a>npu_chip_info_hccs_statistic_info_rx_cnt_X</p>
<p id="p11562871013"><a name="p11562871013"></a><a name="p11562871013"></a>X范围：1~7（Atlas A2 训练系列产品或Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1234520167435"></a><a name="ul1234520167435"></a><ul id="ul1234520167435"><li>第X个HDLC链路接收报文数，单位是flit。</li><li>采集失败时上报-1。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1668141520283"><a name="p1668141520283"></a><a name="p1668141520283"></a>-</p>
</td>
</tr>
<tr id="row1240254522514"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5402245122513"><a name="p5402245122513"></a><a name="p5402245122513"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1139102394715"><a name="p1139102394715"></a><a name="p1139102394715"></a>npu_chip_info_hccs_statistic_info_crc_err_cnt_X</p>
<p id="p1429534131018"><a name="p1429534131018"></a><a name="p1429534131018"></a>X范围：1~7（Atlas A2 训练系列产品或Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul112792234316"></a><a name="ul112792234316"></a><ul id="ul112792234316"><li>第X个HDLC链路接收报文crc错误，单位是flit。</li><li>采集失败时上报-1。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p712111832818"><a name="p712111832818"></a><a name="p712111832818"></a>-</p>
</td>
</tr>
<tr id="row1461410598130"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p193164531143"><a name="p193164531143"></a><a name="p193164531143"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p549382011151"><a name="p549382011151"></a><a name="p549382011151"></a>npu_chip_info_hccs_bandwidth_info_profiling_time</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p549382012152"><a name="p549382012152"></a><a name="p549382012152"></a>HCCS链路带宽采样时长，取值范围1~1000。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p04931220171514"><a name="p04931220171514"></a><a name="p04931220171514"></a>单位：ms</p>
</td>
</tr>
<tr id="row1620614518148"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p18316753201414"><a name="p18316753201414"></a><a name="p18316753201414"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7493220191514"><a name="p7493220191514"></a><a name="p7493220191514"></a>npu_chip_info_hccs_bandwidth_info_total_tx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11493120161511"><a name="p11493120161511"></a><a name="p11493120161511"></a>HCCS链路总发送数据带宽，采集失败时上报-1。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p204931620131517"><a name="p204931620131517"></a><a name="p204931620131517"></a>单位：GB/s</p>
</td>
</tr>
<tr id="row122341714181411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3316185361418"><a name="p3316185361418"></a><a name="p3316185361418"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p149302010159"><a name="p149302010159"></a><a name="p149302010159"></a>npu_chip_info_hccs_bandwidth_info_total_rx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7493820131511"><a name="p7493820131511"></a><a name="p7493820131511"></a>HCCS链路总接收数据带宽，采集失败时上报-1。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p164936208158"><a name="p164936208158"></a><a name="p164936208158"></a>单位：GB/s</p>
</td>
</tr>
<tr id="row1853162231416"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p153161253161412"><a name="p153161253161412"></a><a name="p153161253161412"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p549352091516"><a name="p549352091516"></a><a name="p549352091516"></a>npu_chip_info_hccs_bandwidth_info_tx_X</p>
<p id="p2493172021511"><a name="p2493172021511"></a><a name="p2493172021511"></a>X范围：1~7（Atlas A2 训练系列产品、Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3493152051511"><a name="p3493152051511"></a><a name="p3493152051511"></a>HCCS单链路发送数据带宽，采集失败时上报-1。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p194946203153"><a name="p194946203153"></a><a name="p194946203153"></a>单位：GB/s</p>
</td>
</tr>
<tr id="row18299930131419"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1031665361416"><a name="p1031665361416"></a><a name="p1031665361416"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7494122091513"><a name="p7494122091513"></a><a name="p7494122091513"></a>npu_chip_info_hccs_bandwidth_info_rx_X</p>
<p id="p10494720131520"><a name="p10494720131520"></a><a name="p10494720131520"></a>X范围：1~7（Atlas A2 训练系列产品、Atlas 900 A3 SuperPoD），2~7（Atlas 9000 A3 SuperPoD）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p449492011517"><a name="p449492011517"></a><a name="p449492011517"></a>HCCS单链路接收数据带宽，采集失败时上报-1。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1649442041518"><a name="p1649442041518"></a><a name="p1649442041518"></a>单位：GB/s</p>
</td>
</tr>
</tbody>
</table>

**PCIe数据信息<a name="section1240520241824136"></a>**

**表 7**  PCIe数据信息

<a name="table1341911380255"></a>
<table><thead align="left"><tr id="row941993842520"><th class="cellrowborder" valign="top" width="9.85%" id="mcps1.2.6.1.1"><p id="p5917122513215"><a name="p5917122513215"></a><a name="p5917122513215"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="21.37%" id="mcps1.2.6.1.2"><p id="p14918182519219"><a name="p14918182519219"></a><a name="p14918182519219"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="26.279999999999998%" id="mcps1.2.6.1.3"><p id="p6406358175813"><a name="p6406358175813"></a><a name="p6406358175813"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="11.59%" id="mcps1.2.6.1.4"><p id="p79189258216"><a name="p79189258216"></a><a name="p79189258216"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="30.91%" id="mcps1.2.6.1.5"><p id="p169182258212"><a name="p169182258212"></a><a name="p169182258212"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row174206389252"><td class="cellrowborder" valign="top" width="9.85%" headers="mcps1.2.6.1.1 "><p id="p79181025142113"><a name="p79181025142113"></a><a name="p79181025142113"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" width="21.37%" headers="mcps1.2.6.1.2 "><p id="p59181925142118"><a name="p59181925142118"></a><a name="p59181925142118"></a>npu_chip_info_pcie_rx_p_bw</p>
</td>
<td class="cellrowborder" valign="top" width="26.279999999999998%" headers="mcps1.2.6.1.3 "><p id="p243612169285"><a name="p243612169285"></a><a name="p243612169285"></a><span id="ph643661692816"><a name="ph643661692816"></a><a name="ph643661692816"></a>昇腾AI处理器</span>接收远端写的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" width="11.59%" headers="mcps1.2.6.1.4 "><p id="p1191992510210"><a name="p1191992510210"></a><a name="p1191992510210"></a>单位：MB/ms</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="30.91%" headers="mcps1.2.6.1.5 "><a name="ul64395165289"></a><a name="ul64395165289"></a><ul id="ul64395165289"><li><p id="li2043917168282p0"><a name="li2043917168282p0"></a><a name="li2043917168282p0"></a>Atlas A2 训练系列产品</p>
</li><li><p id="li114396166286p0"><a name="li114396166286p0"></a><a name="li114396166286p0"></a><span id="ph1722042181618"><a name="ph1722042181618"></a><a name="ph1722042181618"></a>Atlas 800I A2 推理服务器</span></p>
</li><li><p id="p258313228429"><a name="p258313228429"></a><a name="p258313228429"></a><span id="ph18642192315427"><a name="ph18642192315427"></a><a name="ph18642192315427"></a>A200I A2 Box 异构组件</span></p>
</li></ul>
</td>
</tr>
<tr id="row34233384255"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7919182532111"><a name="p7919182532111"></a><a name="p7919182532111"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1091942532114"><a name="p1091942532114"></a><a name="p1091942532114"></a>npu_chip_info_pcie_rx_np_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p174481716202815"><a name="p174481716202815"></a><a name="p174481716202815"></a><span id="ph10448141613285"><a name="ph10448141613285"></a><a name="ph10448141613285"></a>昇腾AI处理器</span>接收远端读的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p20919225172118"><a name="p20919225172118"></a><a name="p20919225172118"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row54261838112518"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p15919122502119"><a name="p15919122502119"></a><a name="p15919122502119"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p139191025142118"><a name="p139191025142118"></a><a name="p139191025142118"></a>npu_chip_info_pcie_rx_cpl_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1945841616288"><a name="p1945841616288"></a><a name="p1945841616288"></a><span id="ph1745831612816"><a name="ph1745831612816"></a><a name="ph1745831612816"></a>昇腾AI处理器</span>从远端读收到CPL回复的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16920122520215"><a name="p16920122520215"></a><a name="p16920122520215"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row184291938102510"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p992032582117"><a name="p992032582117"></a><a name="p992032582117"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p29201625102111"><a name="p29201625102111"></a><a name="p29201625102111"></a>npu_chip_info_pcie_tx_p_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p746851613289"><a name="p746851613289"></a><a name="p746851613289"></a><span id="ph10468616112812"><a name="ph10468616112812"></a><a name="ph10468616112812"></a>昇腾AI处理器</span>向远端写PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17920152552119"><a name="p17920152552119"></a><a name="p17920152552119"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row34321838132519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7921125152120"><a name="p7921125152120"></a><a name="p7921125152120"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1692115258214"><a name="p1692115258214"></a><a name="p1692115258214"></a>npu_chip_info_pcie_tx_np_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9406658175816"><a name="p9406658175816"></a><a name="p9406658175816"></a><span id="ph1148018161281"><a name="ph1148018161281"></a><a name="ph1148018161281"></a>昇腾AI处理器</span>从远端读PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p169211251214"><a name="p169211251214"></a><a name="p169211251214"></a>单位：MB/ms</p>
</td>
</tr>
<tr id="row743773816250"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6921325192115"><a name="p6921325192115"></a><a name="p6921325192115"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p692118258210"><a name="p692118258210"></a><a name="p692118258210"></a>npu_chip_info_pcie_tx_cpl_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p144901216132813"><a name="p144901216132813"></a><a name="p144901216132813"></a><span id="ph84901416192811"><a name="ph84901416192811"></a><a name="ph84901416192811"></a>昇腾AI处理器</span>回复远端读操作CPL的PCIe带宽。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p119211225102110"><a name="p119211225102110"></a><a name="p119211225102110"></a>单位：MB/ms</p>
</td>
</tr>
</tbody>
</table>

**RoCE数据信息<a name="section184516450323"></a>**

**表 8**  RoCE数据信息

<a name="table1562691116332"></a>
<table><thead align="left"><tr id="row126261011133318"><th class="cellrowborder" valign="top" width="9.01%" id="mcps1.2.6.1.1"><p id="p149115619355"><a name="p149115619355"></a><a name="p149115619355"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="27.07%" id="mcps1.2.6.1.2"><p id="p15911466357"><a name="p15911466357"></a><a name="p15911466357"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="22.1%" id="mcps1.2.6.1.3"><p id="p14911667358"><a name="p14911667358"></a><a name="p14911667358"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="13.52%" id="mcps1.2.6.1.4"><p id="p59116653520"><a name="p59116653520"></a><a name="p59116653520"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="28.299999999999997%" id="mcps1.2.6.1.5"><p id="p199115623510"><a name="p199115623510"></a><a name="p199115623510"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row862751110336"><td class="cellrowborder" valign="top" width="9.01%" headers="mcps1.2.6.1.1 "><p id="p6133145318341"><a name="p6133145318341"></a><a name="p6133145318341"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" width="27.07%" headers="mcps1.2.6.1.2 "><p id="p17133115323411"><a name="p17133115323411"></a><a name="p17133115323411"></a>npu_chip_mac_rx_pause_num</p>
</td>
<td class="cellrowborder" valign="top" width="22.1%" headers="mcps1.2.6.1.3 "><p id="p5134105313348"><a name="p5134105313348"></a><a name="p5134105313348"></a>MAC接收的Pause帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" width="13.52%" headers="mcps1.2.6.1.4 "><p id="p1413585319348"><a name="p1413585319348"></a><a name="p1413585319348"></a>-</p>
</td>
<td class="cellrowborder" rowspan="21" valign="top" width="28.299999999999997%" headers="mcps1.2.6.1.5 "><a name="ul3135253123412"></a><a name="ul3135253123412"></a><ul id="ul3135253123412"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph1241332842611"><a name="ph1241332842611"></a><a name="ph1241332842611"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph6496152317452"><a name="ph6496152317452"></a><a name="ph6496152317452"></a>A200I A2 Box 异构组件</span></li></ul>
</td>
</tr>
<tr id="row762714116330"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p13135135314348"><a name="p13135135314348"></a><a name="p13135135314348"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p9135125317341"><a name="p9135125317341"></a><a name="p9135125317341"></a>npu_chip_mac_tx_pause_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p4136353183411"><a name="p4136353183411"></a><a name="p4136353183411"></a>MAC发送的Pause帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3136185317341"><a name="p3136185317341"></a><a name="p3136185317341"></a>-</p>
</td>
</tr>
<tr id="row1562751114333"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17136155363415"><a name="p17136155363415"></a><a name="p17136155363415"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1513735318347"><a name="p1513735318347"></a><a name="p1513735318347"></a>npu_chip_mac_rx_pfc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p113710536343"><a name="p113710536343"></a><a name="p113710536343"></a>MAC接收的PFC帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1213755314346"><a name="p1213755314346"></a><a name="p1213755314346"></a>-</p>
</td>
</tr>
<tr id="row14628211153320"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p181389536343"><a name="p181389536343"></a><a name="p181389536343"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p18138145313416"><a name="p18138145313416"></a><a name="p18138145313416"></a>npu_chip_mac_tx_pfc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p91385532343"><a name="p91385532343"></a><a name="p91385532343"></a>MAC发送的PFC帧总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p8138205373419"><a name="p8138205373419"></a><a name="p8138205373419"></a>-</p>
</td>
</tr>
<tr id="row1562881110335"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1113965343414"><a name="p1113965343414"></a><a name="p1113965343414"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p171391253143419"><a name="p171391253143419"></a><a name="p171391253143419"></a>npu_chip_mac_rx_bad_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10139175315340"><a name="p10139175315340"></a><a name="p10139175315340"></a>MAC接收的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16139125393416"><a name="p16139125393416"></a><a name="p16139125393416"></a>-</p>
</td>
</tr>
<tr id="row862816116333"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p151391753173410"><a name="p151391753173410"></a><a name="p151391753173410"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p171407538349"><a name="p171407538349"></a><a name="p171407538349"></a>npu_chip_mac_tx_bad_oct_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p14140853173416"><a name="p14140853173416"></a><a name="p14140853173416"></a>MAC发送的坏包总报文字节数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11401153113412"><a name="p11401153113412"></a><a name="p11401153113412"></a>-</p>
</td>
</tr>
<tr id="row1162961119337"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1714014537344"><a name="p1714014537344"></a><a name="p1714014537344"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p41401253143415"><a name="p41401253143415"></a><a name="p41401253143415"></a>npu_chip_mac_rx_bad_oct_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18141155311348"><a name="p18141155311348"></a><a name="p18141155311348"></a>MAC接收的坏包总报文字节数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p51411053173410"><a name="p51411053173410"></a><a name="p51411053173410"></a>-</p>
</td>
</tr>
<tr id="row693213413419"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p15141753183414"><a name="p15141753183414"></a><a name="p15141753183414"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p171411533341"><a name="p171411533341"></a><a name="p171411533341"></a>npu_chip_mac_tx_bad_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p114213531347"><a name="p114213531347"></a><a name="p114213531347"></a>MAC发送的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p314295312342"><a name="p314295312342"></a><a name="p314295312342"></a>-</p>
</td>
</tr>
<tr id="row17933124113410"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12142185318343"><a name="p12142185318343"></a><a name="p12142185318343"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p614214538344"><a name="p614214538344"></a><a name="p614214538344"></a>npu_chip_roce_rx_all_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p5143125311343"><a name="p5143125311343"></a><a name="p5143125311343"></a>RoCE接收的总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1114305343415"><a name="p1114305343415"></a><a name="p1114305343415"></a>-</p>
</td>
</tr>
<tr id="row1293474173411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14143185363416"><a name="p14143185363416"></a><a name="p14143185363416"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p814320538348"><a name="p814320538348"></a><a name="p814320538348"></a>npu_chip_roce_tx_all_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p61441953123412"><a name="p61441953123412"></a><a name="p61441953123412"></a>RoCE发送的总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p10144135310342"><a name="p10144135310342"></a><a name="p10144135310342"></a>-</p>
</td>
</tr>
<tr id="row109345413348"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1114513534341"><a name="p1114513534341"></a><a name="p1114513534341"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8145353193418"><a name="p8145353193418"></a><a name="p8145353193418"></a>npu_chip_roce_rx_err_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p814665311347"><a name="p814665311347"></a><a name="p814665311347"></a>RoCE接收的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p18146155316347"><a name="p18146155316347"></a><a name="p18146155316347"></a>-</p>
</td>
</tr>
<tr id="row4935124133411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p71476534343"><a name="p71476534343"></a><a name="p71476534343"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p10147185333416"><a name="p10147185333416"></a><a name="p10147185333416"></a>npu_chip_roce_tx_err_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1214775393414"><a name="p1214775393414"></a><a name="p1214775393414"></a>RoCE发送的坏包总报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p201477531346"><a name="p201477531346"></a><a name="p201477531346"></a>-</p>
</td>
</tr>
<tr id="row99353413412"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p614811539346"><a name="p614811539346"></a><a name="p614811539346"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p614995363419"><a name="p614995363419"></a><a name="p614995363419"></a>npu_chip_roce_rx_cnp_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1114915363411"><a name="p1114915363411"></a><a name="p1114915363411"></a>RoCE接收的CNP类型报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p2149115323411"><a name="p2149115323411"></a><a name="p2149115323411"></a>-</p>
</td>
</tr>
<tr id="row693504203411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12150125313417"><a name="p12150125313417"></a><a name="p12150125313417"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1115118535345"><a name="p1115118535345"></a><a name="p1115118535345"></a>npu_chip_roce_tx_cnp_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p111516539341"><a name="p111516539341"></a><a name="p111516539341"></a>RoCE发送的CNP类型报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p81514535345"><a name="p81514535345"></a><a name="p81514535345"></a>-</p>
</td>
</tr>
<tr id="row149361242344"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p315275317344"><a name="p315275317344"></a><a name="p315275317344"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8152653183414"><a name="p8152653183414"></a><a name="p8152653183414"></a>npu_chip_roce_new_pkt_rty_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p21521153143412"><a name="p21521153143412"></a><a name="p21521153143412"></a>RoCE发送的重传的数量统计。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p5152353173416"><a name="p5152353173416"></a><a name="p5152353173416"></a>-</p>
</td>
</tr>
<tr id="row1893410440349"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14153205323412"><a name="p14153205323412"></a><a name="p14153205323412"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p18154253193418"><a name="p18154253193418"></a><a name="p18154253193418"></a>npu_chip_roce_unexpected_ack_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p31543535343"><a name="p31543535343"></a><a name="p31543535343"></a>RoCE接收的非预期ACK报文数，NPU做丢弃处理，不影响业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p151541353153412"><a name="p151541353153412"></a><a name="p151541353153412"></a>-</p>
</td>
</tr>
<tr id="row0935164412342"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p31551553153414"><a name="p31551553153414"></a><a name="p31551553153414"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p915565393419"><a name="p915565393419"></a><a name="p915565393419"></a>npu_chip_roce_out_of_order_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2155155311342"><a name="p2155155311342"></a><a name="p2155155311342"></a>RoCE接收的PSN &gt; 预期PSN的报文，或重复PSN报文数。乱序或丢包，会触发重传。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p20155125314344"><a name="p20155125314344"></a><a name="p20155125314344"></a>-</p>
</td>
</tr>
<tr id="row1936144463416"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p915613539340"><a name="p915613539340"></a><a name="p915613539340"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p31561153123417"><a name="p31561153123417"></a><a name="p31561153123417"></a>npu_chip_roce_verification_err_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2156175353417"><a name="p2156175353417"></a><a name="p2156175353417"></a>RoCE接收的域段校验失败的报文数，域段校验的场景包括：ICRC、报文长度、目的端口号等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p215675314348"><a name="p215675314348"></a><a name="p215675314348"></a>-</p>
</td>
</tr>
<tr id="row793704413343"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p10157155314349"><a name="p10157155314349"></a><a name="p10157155314349"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p71577532341"><a name="p71577532341"></a><a name="p71577532341"></a>npu_chip_roce_qp_status_err_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p91571453163410"><a name="p91571453163410"></a><a name="p91571453163410"></a>RoCE接收的QP连接状态异常产生的报文数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p415795323418"><a name="p415795323418"></a><a name="p415795323418"></a>-</p>
</td>
</tr>
<tr id="row1334918433510"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1349443653"><a name="p1349443653"></a><a name="p1349443653"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p123491043457"><a name="p123491043457"></a><a name="p123491043457"></a>npu_chip_info_rx_ecn_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p103491431157"><a name="p103491431157"></a><a name="p103491431157"></a>昇腾AI处理器网络接收ECN数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p103499431552"><a name="p103499431552"></a><a name="p103499431552"></a>-</p>
</td>
</tr>
<tr id="row1433394916519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p833313493511"><a name="p833313493511"></a><a name="p833313493511"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14334174915514"><a name="p14334174915514"></a><a name="p14334174915514"></a>npu_chip_info_rx_fcs_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p6334349155"><a name="p6334349155"></a><a name="p6334349155"></a>昇腾AI处理器网络接收FCS数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p123341491158"><a name="p123341491158"></a><a name="p123341491158"></a>-</p>
</td>
</tr>
</tbody>
</table>

**SIO数据信息<a name="section7109037161515"></a>**

**表 9**  SIO数据信息

<a name="table1910972371718"></a>
<table><thead align="left"><tr id="row10109122321710"><th class="cellrowborder" valign="top" width="9.509999999999998%" id="mcps1.2.6.1.1"><p id="p18120184111178"><a name="p18120184111178"></a><a name="p18120184111178"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="28.559999999999995%" id="mcps1.2.6.1.2"><p id="p1121144111714"><a name="p1121144111714"></a><a name="p1121144111714"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="21.109999999999996%" id="mcps1.2.6.1.3"><p id="p312116415171"><a name="p312116415171"></a><a name="p312116415171"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="17.509999999999998%" id="mcps1.2.6.1.4"><p id="p1212112418173"><a name="p1212112418173"></a><a name="p1212112418173"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="23.309999999999995%" id="mcps1.2.6.1.5"><p id="p12121204113171"><a name="p12121204113171"></a><a name="p12121204113171"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row12110123121718"><td class="cellrowborder" valign="top" width="9.509999999999998%" headers="mcps1.2.6.1.1 "><p id="p6113184916172"><a name="p6113184916172"></a><a name="p6113184916172"></a>SIO</p>
</td>
<td class="cellrowborder" valign="top" width="28.559999999999995%" headers="mcps1.2.6.1.2 "><p id="p211316490178"><a name="p211316490178"></a><a name="p211316490178"></a>npu_chip_info_sio_crc_tx_err_cnt</p>
</td>
<td class="cellrowborder" valign="top" width="21.109999999999996%" headers="mcps1.2.6.1.3 "><p id="p6113194914170"><a name="p6113194914170"></a><a name="p6113194914170"></a>SIO发送的错包数</p>
</td>
<td class="cellrowborder" valign="top" width="17.509999999999998%" headers="mcps1.2.6.1.4 "><p id="p10113449181711"><a name="p10113449181711"></a><a name="p10113449181711"></a>-</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="23.309999999999995%" headers="mcps1.2.6.1.5 "><ul><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 350 标卡）</li></ul>
</td>
</tr>
<tr id="row1111082310171"><td class="cellrowborder" valign="top" width="9.509999999999998%" headers="mcps1.2.6.1.1 "><p id="p10114204910174"><a name="p10114204910174"></a><a name="p10114204910174"></a>SIO</p>
</td>
<td class="cellrowborder" valign="top" width="28.559999999999995%" headers="mcps1.2.6.1.2 "><p id="p411484911177"><a name="p411484911177"></a><a name="p411484911177"></a>npu_chip_info_sio_crc_rx_err_cnt</p>
</td>
<td class="cellrowborder" valign="top" width="21.109999999999996%" headers="mcps1.2.6.1.3 "><p id="p1911454914173"><a name="p1911454914173"></a><a name="p1911454914173"></a>SIO接收的错包数</p>
</td>
<td class="cellrowborder" valign="top" width="17.509999999999998%" headers="mcps1.2.6.1.4 "><p id="p9114104916179"><a name="p9114104916179"></a><a name="p9114104916179"></a>-</p>
</td>
</tr>
</tbody>
</table>

**光模块数据信息<a name="section1517163183510"></a>**

**表 10**  光模块数据信息

<a name="table1379935213357"></a>
<table><thead align="left"><tr id="row1279915521353"><th class="cellrowborder" valign="top" width="9.08%" id="mcps1.2.6.1.1"><p id="p570012555368"><a name="p570012555368"></a><a name="p570012555368"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="27.27%" id="mcps1.2.6.1.2"><p id="p670145511363"><a name="p670145511363"></a><a name="p670145511363"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="19.24%" id="mcps1.2.6.1.3"><p id="p14701955113619"><a name="p14701955113619"></a><a name="p14701955113619"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="17.52%" id="mcps1.2.6.1.4"><p id="p5701115514361"><a name="p5701115514361"></a><a name="p5701115514361"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="26.889999999999997%" id="mcps1.2.6.1.5"><p id="p970175515360"><a name="p970175515360"></a><a name="p970175515360"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row28001952103515"><td class="cellrowborder" valign="top" width="9.08%" headers="mcps1.2.6.1.1 "><p id="p176808377361"><a name="p176808377361"></a><a name="p176808377361"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" width="27.27%" headers="mcps1.2.6.1.2 "><p id="p368103793614"><a name="p368103793614"></a><a name="p368103793614"></a>npu_chip_optical_state</p>
</td>
<td class="cellrowborder" valign="top" width="19.24%" headers="mcps1.2.6.1.3 "><p id="p13681143793613"><a name="p13681143793613"></a><a name="p13681143793613"></a>光模块在位状态。</p>
</td>
<td class="cellrowborder" valign="top" width="17.52%" headers="mcps1.2.6.1.4 "><p id="p12681133718365"><a name="p12681133718365"></a><a name="p12681133718365"></a>取值为0或1</p>
<a name="ul14681837183617"></a><a name="ul14681837183617"></a><ul id="ul14681837183617"><li>0：不在位</li><li>1：在位</li></ul>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="26.889999999999997%" headers="mcps1.2.6.1.5 "><a name="ul1868114372365"></a><a name="ul1868114372365"></a><ul id="ul1868114372365"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li><span id="ph1768263703612"><a name="ph1768263703612"></a><a name="ph1768263703612"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph16373157182715"><a name="ph16373157182715"></a><a name="ph16373157182715"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph103551958184611"><a name="ph103551958184611"></a><a name="ph103551958184611"></a>A200I A2 Box 异构组件</span></li></ul>
</td>
</tr>
<tr id="row780035203510"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7682143773611"><a name="p7682143773611"></a><a name="p7682143773611"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p668263773619"><a name="p668263773619"></a><a name="p668263773619"></a>npu_chip_optical_tx_power_X（X范围：0~3）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1682133783619"><a name="p1682133783619"></a><a name="p1682133783619"></a>光模块发送功率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p968363753614"><a name="p968363753614"></a><a name="p968363753614"></a>单位：mW</p>
</td>
</tr>
<tr id="row5175131619363"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p19685173716367"><a name="p19685173716367"></a><a name="p19685173716367"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p16851337173614"><a name="p16851337173614"></a><a name="p16851337173614"></a>npu_chip_optical_rx_power_X（X范围：0~3）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15685173713364"><a name="p15685173713364"></a><a name="p15685173713364"></a>光模块接收功率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p268533719363"><a name="p268533719363"></a><a name="p268533719363"></a>单位：mW</p>
</td>
</tr>
<tr id="row16175111616365"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4686837133615"><a name="p4686837133615"></a><a name="p4686837133615"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p136861437133613"><a name="p136861437133613"></a><a name="p136861437133613"></a>npu_chip_optical_vcc</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1868663753610"><a name="p1868663753610"></a><a name="p1868663753610"></a>光模块电压。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p166863375361"><a name="p166863375361"></a><a name="p166863375361"></a>单位：mV</p>
</td>
</tr>
<tr id="row5800105220358"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p2068715375363"><a name="p2068715375363"></a><a name="p2068715375363"></a>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1368793719362"><a name="p1368793719362"></a><a name="p1368793719362"></a>npu_chip_optical_temp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p20687183763617"><a name="p20687183763617"></a><a name="p20687183763617"></a>光模块温度。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p106882371361"><a name="p106882371361"></a><a name="p106882371361"></a>单位：摄氏度（℃）</p>
</td>
</tr>
</tbody>
</table>

**DDR数据信息<a name="section114607361931169"></a>**

**表 11**  DDR数据信息

<a name="table1251541123212"></a>
<table><thead align="left"><tr id="row152510419324"><th class="cellrowborder" valign="top" width="11.288871112888712%" id="mcps1.2.6.1.1"><p id="p191542026122120"><a name="p191542026122120"></a><a name="p191542026122120"></a>类别</p>
</th>
<th class="cellrowborder" valign="top" width="28.43715628437156%" id="mcps1.2.6.1.2"><p id="p121551826102119"><a name="p121551826102119"></a><a name="p121551826102119"></a>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.897510248975102%" id="mcps1.2.6.1.3"><p id="p2015511264214"><a name="p2015511264214"></a><a name="p2015511264214"></a>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="11.668833116688331%" id="mcps1.2.6.1.4"><p id="p515513269215"><a name="p515513269215"></a><a name="p515513269215"></a>单位</p>
</th>
<th class="cellrowborder" valign="top" width="23.707629237076294%" id="mcps1.2.6.1.5"><p id="p191558266210"><a name="p191558266210"></a><a name="p191558266210"></a>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr id="row7261741103215"><td class="cellrowborder" valign="top" width="11.288871112888712%" headers="mcps1.2.6.1.1 "><p id="p101556265217"><a name="p101556265217"></a><a name="p101556265217"></a>DDR</p>
</td>
<td class="cellrowborder" valign="top" width="28.43715628437156%" headers="mcps1.2.6.1.2 "><p id="p1915532642112"><a name="p1915532642112"></a><a name="p1915532642112"></a>npu_chip_info_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="24.897510248975102%" headers="mcps1.2.6.1.3 "><p id="p2155182642119"><a name="p2155182642119"></a><a name="p2155182642119"></a><span id="ph16848144816402"><a name="ph16848144816402"></a><a name="ph16848144816402"></a>昇腾AI处理器</span>DDR内存已使用量</p>
</td>
<td class="cellrowborder" valign="top" width="11.668833116688331%" headers="mcps1.2.6.1.4 "><p id="p151555262211"><a name="p151555262211"></a><a name="p151555262211"></a>单位：MB</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="23.707629237076294%" headers="mcps1.2.6.1.5 "><a name="ul12849124816407"></a><a name="ul12849124816407"></a><ul id="ul12849124816407"><li><p id="li16850648204019p0"><a name="li16850648204019p0"></a><a name="li16850648204019p0"></a>Atlas 训练系列产品</p>
</li><li><p id="li1785018488403p0"><a name="li1785018488403p0"></a><a name="li1785018488403p0"></a>推理服务器（插Atlas 300I 推理卡）</p>
</li><li><p id="li9850948124020p0"><a name="li9850948124020p0"></a><a name="li9850948124020p0"></a>Atlas 推理系列产品</p>
</li></ul>
</td>
</tr>
<tr id="row20281241173220"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p315632620214"><a name="p315632620214"></a><a name="p315632620214"></a>DDR</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p615682622115"><a name="p615682622115"></a><a name="p615682622115"></a>npu_chip_info_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1156626172118"><a name="p1156626172118"></a><a name="p1156626172118"></a><span id="ph8854114810407"><a name="ph8854114810407"></a><a name="ph8854114810407"></a>昇腾AI处理器</span>DDR内存总量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1215682615214"><a name="p1215682615214"></a><a name="p1215682615214"></a>单位：MB</p>
</td>
</tr>
</tbody>
</table>

**调用的HDK接口<a name="section345820153363"></a>**

NPU Exporter是通过调用底层的HDK接口，获取相应的信息。数据信息调用的HDK接口请参考[NPU Exporter调用的HDK接口.xlsx](../resource/NPU%20Exporter调用的HDK接口.xlsx)。查找数据信息对应的HDK接口，可参考如下步骤。

1.  登录[昇腾计算文档](https://support.huawei.com/enterprise/zh/category/ascend-computing-pid-1557196528909?submodel=doc)中心，选择单击对应产品名称，进入文档界面。例如Atlas 800I A2 推理服务器产品的用户，单击“Atlas 800I A2”。
2.  在左侧导航栏找到“二次开发”，根据接口的类型选择对应文档。
    -   DCMI接口选择“API参考”，单击进入《[DCMI API参考](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-hdk-pid-252764743?category=developer-documents&subcategory=api-reference)》。
    -   HCCN Tool接口选择“接口参考”，单击进入《[Atlas A2 中心推理和训练硬件 25.5.0 HCCN Tool 接口参考](https://support.huawei.com/enterprise/zh/doc/EDOC1100540101/426cffd9)》。

3.  在文档首页搜索栏中，直接搜索对应的接口名称或者关键词，获取接口的相关信息。


## 自定义指标文件<a name="ZH-CN_TOPIC_0000002501343480"></a>

**字段说明<a name="section42791696421"></a>**

自定义指标文件中的字段说明如[表1](#ZH-CN_TOPIC_0000002501343480_table5395205714441)所示。通过自定义指标文件可开发自定义指标，详情请参见[自定义指标开发](../appendix.md#自定义指标开发)。

**表 1**  自定义指标文件字段说明

<a name="ZH-CN_TOPIC_0000002501343480_table5395205714441"></a>
|字段名称|类型|说明|
|--|--|--|
|version|string|固定值：1.0。|
|name|string|指标名称。不能为空，长度不能超过128。|
|desc|string|指标的详情介绍。不能为空，长度不能超过1024。|
|timestamp|timestamp|指标的更新时间戳，单位为us。|
|data_list|list|非空数组，长度不能超过128。|
|-value|float|指标的值。|
|-label|json|指标的标签，JSON的key和value都必须是字符串类型，JSON的子元素个数不能超过10个。|


**约束说明<a name="section342514924413"></a>**

-   不支持软链接。
-   不支持指定一个目录。
-   支持指定的文件路径不能超过10个，多个文件之间以逗号分隔。
-   当指定多个文件时，若其中部分文件不存在或为空文件，合并等待1分钟后仍不满足条件，则取消对应文件的指标采集。
-   文件中字段的格式需满足[表1](#ZH-CN_TOPIC_0000002501343480_table5395205714441)的要求。格式不正确会取消对应文件的指标采集。
-   自动获取指标的label时，以data\_list中的第一条数据为准。
-   自定义指标文件的属组需为npu-exporter进程属组，且具有读权限，不具有任何执行权限。
-   程序运行中不支持修改文件的name、desc和version。
-   程序运行中若修改label名称，程序不会感知，仍按初始化时的label名称上报。
-   自定义指标文件大小限制为100KB。
-   容器场景部署时，需确保指标文件正确挂载到容器中。
-   指标采集过程中，若某批次数据不满足约束，则忽略本批次数据，按缓存数据进行上报。

**指标文件样例**

```
{
  "version": "1.0",
  "name": "hccs_bandwidth",
  "desc": "hccs bandwidth info, unit is 'MB/s'.",
  "timestamp": 1766456419845127,
  "data_list": [
    {
      "value": 190.02,
      "label": {
        "numa": "2",
        "device": "hisi_sicl10_pa0",
        "link": "0",
        "direction": "in",
        "path": "P0->P1"
      }
    },
    {
      "value": 143.09,
      "label": {
        "numa": "2",
        "device": "hisi_sicl10_pa0",
        "link": "1",
        "direction": "in",
        "path": "P2->P1"
      }
    }
  ]
}
```

**Telegraf中自定义指标上报样例**

```
/tmp/data/data.json,device=hisi_sicl10_pa0,direction=in,host=ubuntu20,link=0,numa=2,path=P0->P1 hccs_bandwidth=190.02 1766456419845127000
/tmp/data/data.json,device=hisi_sicl10_pa0,direction=in,host=ubuntu20,link=1,numa=2,path=P2->P1 hccs_bandwidth=143.09 1766456419845127000
```
