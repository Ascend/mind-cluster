# 附录

## 边缘容器日志输出指导<a name="ZH-CN_TOPIC_0000002479226424"></a>

**使用背景<a name="zh-cn_topic_0000001589264561_section15670165114555"></a>**

由于边缘设备（如Atlas 500 A2 智能小站）存储空间有限，并且边缘设备多采用eMMC等flash作为存储介质，该介质存在使用寿命的限制。为避免存储空间过快被写满从而影响业务或存储介质过快达到使用寿命，用户可以参考本章节边缘容器日志的输出建议，使边缘容器以合适的方式输出日志。

**输出方式<a name="zh-cn_topic_0000001589264561_section5556162785617"></a>**

当前Atlas硬件上运行的边缘容器应用一般是通过K8s兼容的边缘管理平台来进行管理，如华为云IEF或基于KubeEdge搭建的第三方边缘平台等。在该平台下，容器日志的输出方式主要分为以下三种：

- 容器控制台标准输出（STDOUT和STDERR）方式
- （推荐）挂载到主机目录方式
- 容器日志直接输出到日志服务

>[!NOTE] 
>如果系统中有日志服务器，建议直接在容器中将日志输出到日志服务中；如果没有，建议采用挂载到主机目录的方式输出日志，减少日志对硬件和其他业务影响的风险。

**容器控制台标准输出方式<a name="zh-cn_topic_0000001589264561_section8645749571"></a>**

在这种方式下，应用将容器的日志输出到标准输出。缺省情况下，Docker引擎捕捉所有容器的标准输出，使用JSON格式写入到文件里，该文件会保存到主机的/var/lib/docker/containers/<i>\{containerid\}</i>目录下，如[图1](#zh-cn_topic_0000001589264561_zh-cn_topic_0000001182332559_zh-cn_topic_0000001092717454_fig167489420139)所示。

**图 1** _\{containerid\}_-json.log文件所在路径示例<a name="zh-cn_topic_0000001589264561_zh-cn_topic_0000001182332559_zh-cn_topic_0000001092717454_fig167489420139"></a>  
![](../figures/scheduling/containerid--json-log文件所在路径示例.png "containerid--json-log文件所在路径示例")

>[!NOTE] 
>如果边缘管理平台不支持该目录下日志文件的绕接或日志绕接配置错误，会导致<b>/var/lib/docker</b>被占满，从而影响新容器的部署及其他容器业务的正常运行。故不建议采用该方式。

**（推荐）挂载到主机目录方式<a name="zh-cn_topic_0000001589264561_section139871046185718"></a>**

该方式下边缘平台日志收集的方式如[图2](#zh-cn_topic_0000001589264561_zh-cn_topic_0000001182452463_zh-cn_topic_0000001140102079_fig13294175199)所示。

**图 2**  方案架构<a name="zh-cn_topic_0000001589264561_zh-cn_topic_0000001182452463_zh-cn_topic_0000001140102079_fig13294175199"></a>  
![](../figures/scheduling/方案架构.png "方案架构")

应用将容器日志挂载到边缘主机上。边缘管理平台提供主机上日志收集能力，并将主机文件日志进行绕接。

>[!NOTE] 
>
>- 应用可以将容器日志挂载到主机上的非关键大容量目录，建议不要挂载到eMMC等存储介质上，避免影响硬件整体寿命。
>- 边缘容器管理平台一般会支持该能力，以减少对系统目录**var/lib/docker**的影响。基于安全性考虑，该配置需要符合所在组织的安全要求。

**容器日志直接输出到日志服务<a name="zh-cn_topic_0000001589264561_section195870131582"></a>**

如[图3](#zh-cn_topic_0000001589264561_zh-cn_topic_0000001136212966_zh-cn_topic_0000001093005606_fig8724931363)所示，应用环境里如果有日志服务器，可以将日志直接输出到外部日志服务器，使日志不在边缘环境里落盘，最大限度减少对硬件和其他业务的影响。

**图 3**  方案架构<a name="zh-cn_topic_0000001589264561_zh-cn_topic_0000001136212966_zh-cn_topic_0000001093005606_fig8724931363"></a>  
![](../figures/scheduling/方案架构-0.png "方案架构-0")

## Ascend Docker Runtime默认挂载内容<a name="ZH-CN_TOPIC_0000002511346331"></a>

Ascend Docker Runtime会根据实际环境情况默认以只读方式挂载以下目录和文件到容器中。

**表 1**  默认挂载目录和文件（Atlas 200 AI加速模块（RC场景））

<a name="zh-cn_topic_0000001538584750_table11867194212594"></a>

|路径|说明|
|--|--|
|/dev/davinci*X*|NPU设备，X是ID号。例如：davinci0。|
|/dev/davinci_manager|管理设备。|
|/usr/local/Ascend/driver/tools|目录，驱动提供的工具包。|
|/usr/local/Ascend/driver/lib64|目录，驱动提供的用户态库。|
|/usr/local/sbin/npu-smi|文件，NPU-SMI工具。|
|/etc/hdcBasic.cfg|文件，hdc基础文件。|
|/etc/sys_version.conf|文件，驱动的版本信息。|
|/dev/dvpp_cmdlist|设备文件，支撑推理业务。|
|/var/queue_schedule|管理FlowGW调度框架。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>挂载此目录需同时满足以下条件：</p><ul><li>MindCluster组件版本≥6.0.0。</li><li>HDK版本≥24.1.RC2。</li></ul></div></div>|

**表 2**  默认挂载目录和文件（Atlas 200I SoC A1 核心板）

<a name="zh-cn_topic_0000001538584750_table2868154235914"></a>

|路径|说明|
|--|--|
|/dev/davinci*X*|NPU设备，X是ID号。例如：davinci0。|
|/dev/davinci_manager|davinci相关的设备管理设备。|
|/usr/local/bin/npu-smi|文件，NPU-SMI工具。|
|/etc/hdcBasic.cfg|文件，hdc基础文件。|
|/etc/sys_version.conf|文件，驱动的版本信息。|
|/dev/dvpp_cmdlist|设备文件，支撑推理业务。|
|/var/queue_schedule|管理FlowGW调度框架。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>挂载此目录需同时满足以下条件：</p><ul><li>MindCluster组件版本≥6.0.0。</li><li>HDK版本≥24.1.RC2。</li></ul></div></div>|

**表 3**  默认挂载目录和文件（Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件）

<a name="zh-cn_topic_0000001538584750_table1986129115"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001538584750_row158718919114"><th class="cellrowborder" valign="top" width="42.86%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000001538584750_p2871497112"><a name="zh-cn_topic_0000001538584750_p2871497112"></a><a name="zh-cn_topic_0000001538584750_p2871497112"></a>路径</p>
</th>
<th class="cellrowborder" valign="top" width="57.14%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000001538584750_p148716919110"><a name="zh-cn_topic_0000001538584750_p148716919110"></a><a name="zh-cn_topic_0000001538584750_p148716919110"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001538584750_row887398115"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p3873913114"><a name="zh-cn_topic_0000001538584750_p3873913114"></a><a name="zh-cn_topic_0000001538584750_p3873913114"></a>/dev/davinciX</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p987149212"><a name="zh-cn_topic_0000001538584750_p987149212"></a><a name="zh-cn_topic_0000001538584750_p987149212"></a>NPU设备，X是ID号。例如：davinci0。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row88720918119"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p587191016"><a name="zh-cn_topic_0000001538584750_p587191016"></a><a name="zh-cn_topic_0000001538584750_p587191016"></a>/dev/davinci_manager</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p18878915118"><a name="zh-cn_topic_0000001538584750_p18878915118"></a><a name="zh-cn_topic_0000001538584750_p18878915118"></a>davinci相关的设备管理设备。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row17871991715"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p5871596110"><a name="zh-cn_topic_0000001538584750_p5871596110"></a><a name="zh-cn_topic_0000001538584750_p5871596110"></a>/dev/svm0</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p4874911116"><a name="zh-cn_topic_0000001538584750_p4874911116"></a><a name="zh-cn_topic_0000001538584750_p4874911116"></a>内存管理的设备。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row12871991110"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p118714910112"><a name="zh-cn_topic_0000001538584750_p118714910112"></a><a name="zh-cn_topic_0000001538584750_p118714910112"></a>/dev/ts_aisle</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p14871191816"><a name="zh-cn_topic_0000001538584750_p14871191816"></a><a name="zh-cn_topic_0000001538584750_p14871191816"></a>aicpudrv驱动设备，为任务调度提供事件驱动的渠道接口。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row387694111"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p158717920111"><a name="zh-cn_topic_0000001538584750_p158717920111"></a><a name="zh-cn_topic_0000001538584750_p158717920111"></a>/dev/upgrade</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p787109313"><a name="zh-cn_topic_0000001538584750_p787109313"></a><a name="zh-cn_topic_0000001538584750_p787109313"></a>驱动设备。</p>
<p id="zh-cn_topic_0000001538584750_p488199419"><a name="zh-cn_topic_0000001538584750_p488199419"></a><a name="zh-cn_topic_0000001538584750_p488199419"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row1888792116"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p2881297112"><a name="zh-cn_topic_0000001538584750_p2881297112"></a><a name="zh-cn_topic_0000001538584750_p2881297112"></a>/dev/sys</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row16881191116"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p3881291514"><a name="zh-cn_topic_0000001538584750_p3881291514"></a><a name="zh-cn_topic_0000001538584750_p3881291514"></a>/dev/vdec</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p19885910111"><a name="zh-cn_topic_0000001538584750_p19885910111"></a><a name="zh-cn_topic_0000001538584750_p19885910111"></a>设备文件，支撑推理业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row2088891710"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1088691118"><a name="zh-cn_topic_0000001538584750_p1088691118"></a><a name="zh-cn_topic_0000001538584750_p1088691118"></a>/dev/vpc</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row158819919112"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1188891316"><a name="zh-cn_topic_0000001538584750_p1188891316"></a><a name="zh-cn_topic_0000001538584750_p1188891316"></a>/dev/pngd</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row588179715"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p168829515"><a name="zh-cn_topic_0000001538584750_p168829515"></a><a name="zh-cn_topic_0000001538584750_p168829515"></a>/dev/venc</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row488199215"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p38889319"><a name="zh-cn_topic_0000001538584750_p38889319"></a><a name="zh-cn_topic_0000001538584750_p38889319"></a>/dev/dvpp_cmdlist</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row5881891118"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p38815916118"><a name="zh-cn_topic_0000001538584750_p38815916118"></a><a name="zh-cn_topic_0000001538584750_p38815916118"></a>/dev/log_drv</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p108811910116"><a name="zh-cn_topic_0000001538584750_p108811910116"></a><a name="zh-cn_topic_0000001538584750_p108811910116"></a>日志驱动设备。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row188829510"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p165082055743"><a name="zh-cn_topic_0000001538584750_p165082055743"></a><a name="zh-cn_topic_0000001538584750_p165082055743"></a>/etc/sys_version.conf</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p1488791019"><a name="zh-cn_topic_0000001538584750_p1488791019"></a><a name="zh-cn_topic_0000001538584750_p1488791019"></a>文件，驱动的版本信息。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row1788391510"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p205071551147"><a name="zh-cn_topic_0000001538584750_p205071551147"></a><a name="zh-cn_topic_0000001538584750_p205071551147"></a>/etc/hdcBasic.cfg</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p207918322811"><a name="zh-cn_topic_0000001538584750_p207918322811"></a><a name="zh-cn_topic_0000001538584750_p207918322811"></a>文件，hdc基础文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row4405101113211"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p767371282114"><a name="zh-cn_topic_0000001538584750_p767371282114"></a><a name="zh-cn_topic_0000001538584750_p767371282114"></a>/usr/local/sbin/npu-smi</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p8406151192117"><a name="zh-cn_topic_0000001538584750_p8406151192117"></a><a name="zh-cn_topic_0000001538584750_p8406151192117"></a>文件，NPU-SMI工具。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row191323162119"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p202501407219"><a name="zh-cn_topic_0000001538584750_p202501407219"></a><a name="zh-cn_topic_0000001538584750_p202501407219"></a>/usr/local/Ascend/driver/lib64</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p14913103162113"><a name="zh-cn_topic_0000001538584750_p14913103162113"></a><a name="zh-cn_topic_0000001538584750_p14913103162113"></a>目录，驱动提供的用户态库。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row591373112319"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p52527011210"><a name="zh-cn_topic_0000001538584750_p52527011210"></a><a name="zh-cn_topic_0000001538584750_p52527011210"></a>/usr/lib64/aicpu_kernels/</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row71535348212"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p92491501323"><a name="zh-cn_topic_0000001538584750_p92491501323"></a><a name="zh-cn_topic_0000001538584750_p92491501323"></a>/var/slogd</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p189302374220"><a name="zh-cn_topic_0000001538584750_p189302374220"></a><a name="zh-cn_topic_0000001538584750_p189302374220"></a>文件，日志组件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row15553144182114"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1455394452117"><a name="zh-cn_topic_0000001538584750_p1455394452117"></a><a name="zh-cn_topic_0000001538584750_p1455394452117"></a>/var/dmp_daemon</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p3553444112118"><a name="zh-cn_topic_0000001538584750_p3553444112118"></a><a name="zh-cn_topic_0000001538584750_p3553444112118"></a>文件，dmp守护进程。</p>
</td>
</tr>
<tr id="row4773105231920"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="p37731952181911"><a name="p37731952181911"></a><a name="p37731952181911"></a>/usr/lib64/libcrypto.so.1.1</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="p1354341203715"><a name="p1354341203715"></a><a name="p1354341203715"></a>文件，驱动所需动态库。</p>
<p id="p10927155142111"><a name="p10927155142111"></a><a name="p10927155142111"></a><span id="ph9796114014252"><a name="ph9796114014252"></a><a name="ph9796114014252"></a>openEuler</span> 22.03需要挂载该路径。</p>
</td>
</tr>
<tr id="row2418501193"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p144145021912"><a name="p144145021912"></a><a name="p144145021912"></a>/usr/lib64/libyaml-0.so.2</p>
</td>
</tr>
<tr id="row1826901502019"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="p226991514205"><a name="p226991514205"></a><a name="p226991514205"></a>/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="p19161449374"><a name="p19161449374"></a><a name="p19161449374"></a>文件，驱动所需动态库。</p>
<p id="p540120472227"><a name="p540120472227"></a><a name="p540120472227"></a><span id="ph1052114212228"><a name="ph1052114212228"></a><a name="ph1052114212228"></a>Ubuntu</span> 22.04需要挂载该路径。</p>
</td>
</tr>
<tr id="row211711176202"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1111821752013"><a name="p1111821752013"></a><a name="p1111821752013"></a>/usr/lib/aarch64-linux-gnu/libyaml-0.so.2</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row989119916"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p11506655744"><a name="zh-cn_topic_0000001538584750_p11506655744"></a><a name="zh-cn_topic_0000001538584750_p11506655744"></a>/usr/lib64/libaicpu_processer.so</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p1189591513"><a name="zh-cn_topic_0000001538584750_p1189591513"></a><a name="zh-cn_topic_0000001538584750_p1189591513"></a>文件，驱动所需动态库。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row108919919120"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p175057555415"><a name="zh-cn_topic_0000001538584750_p175057555415"></a><a name="zh-cn_topic_0000001538584750_p175057555415"></a>/usr/lib64/libaicpu_prof.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row108912918120"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p450413551641"><a name="zh-cn_topic_0000001538584750_p450413551641"></a><a name="zh-cn_topic_0000001538584750_p450413551641"></a>/usr/lib64/libaicpu_sharder.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row389149511"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1625520017218"><a name="zh-cn_topic_0000001538584750_p1625520017218"></a><a name="zh-cn_topic_0000001538584750_p1625520017218"></a>/usr/lib64/libadump.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row989199714"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p425514020216"><a name="zh-cn_topic_0000001538584750_p425514020216"></a><a name="zh-cn_topic_0000001538584750_p425514020216"></a>/usr/lib64/libtsd_eventclient.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row78919917112"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p5254304219"><a name="zh-cn_topic_0000001538584750_p5254304219"></a><a name="zh-cn_topic_0000001538584750_p5254304219"></a>/usr/lib64/libaicpu_scheduler.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row58918913114"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p9946242610"><a name="zh-cn_topic_0000001538584750_p9946242610"></a><a name="zh-cn_topic_0000001538584750_p9946242610"></a>/usr/lib64/libdcmi.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row19901097112"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p625330423"><a name="zh-cn_topic_0000001538584750_p625330423"></a><a name="zh-cn_topic_0000001538584750_p625330423"></a>/usr/lib64/libmpi_dvpp_adapter.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row14901996110"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p62511103210"><a name="zh-cn_topic_0000001538584750_p62511103210"></a><a name="zh-cn_topic_0000001538584750_p62511103210"></a>/usr/lib64/libstackcore.so</p>
</td>
</tr>
<tr id="row88572085920"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="p17250112212918"><a name="p17250112212918"></a><a name="p17250112212918"></a>/var/queue_schedule</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="p192505227918"><a name="p192505227918"></a><a name="p192505227918"></a>管理FlowGW调度框架。</p>
<div class="note" id="note62503223913"><a name="note62503223913"></a><div class="notebody"><p id="p325017223919"><a name="p325017223919"></a><a name="p325017223919"></a>挂载此目录需同时满足以下条件：</p>
<a name="ul112517221897"></a><a name="ul112517221897"></a><ul id="ul112517221897"><li>MindCluster组件版本≥6.0.0。</li><li>HDK版本≥24.1.RC2。</li></ul>
</div></div>
</td>
</tr>
</tbody>
</table>

**表 4**  默认挂载目录和文件（Atlas 500 智能小站（型号 3000））

<a name="zh-cn_topic_0000001538584750_table13873642175917"></a>

|路径|说明|
|--|--|
|/dev/davinci*X*|NPU设备，X是ID号。例如：davinci0。|
|/dev/davinci_manager|管理设备。|
|/dev/hisi_hdc|管理设备。|
|/dev/devmm_svm|管理设备。|
|/home/data/miniD/driver/lib64|目录，驱动提供的用户态库。|
|/usr/local/dcmi|目录，DCMI头文件和库。|
|/usr/local/lib/libdcmi.so|文件，DCMI动态库。|
|/usr/local/bin/npu-smi|文件，NPU-SMI工具。|
|/dev/dvpp_cmdlist|设备文件，支撑推理业务。|
|/var/queue_schedule|管理FlowGW调度框架。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>挂载此目录需同时满足以下条件：</p><ul><li>MindCluster组件版本≥6.0.0。</li><li>HDK版本≥24.1.RC2。</li></ul></div></div>|

**表 5**  默认挂载目录和文件（Atlas 500 A2 智能小站）

<a name="zh-cn_topic_0000001538584750_table1023983110534"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001538584750_row11240193115538"><th class="cellrowborder" valign="top" width="42.86%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000001538584750_p16240731145317"><a name="zh-cn_topic_0000001538584750_p16240731145317"></a><a name="zh-cn_topic_0000001538584750_p16240731145317"></a>路径</p>
</th>
<th class="cellrowborder" valign="top" width="57.14%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000001538584750_p32401731185310"><a name="zh-cn_topic_0000001538584750_p32401731185310"></a><a name="zh-cn_topic_0000001538584750_p32401731185310"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001538584750_row424018316537"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p91141812145416"><a name="zh-cn_topic_0000001538584750_p91141812145416"></a><a name="zh-cn_topic_0000001538584750_p91141812145416"></a>/dev/davinciX</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p1733622015418"><a name="zh-cn_topic_0000001538584750_p1733622015418"></a><a name="zh-cn_topic_0000001538584750_p1733622015418"></a>NPU设备，X是ID号。例如：davinci0。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row1724013155312"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p3785534175420"><a name="zh-cn_topic_0000001538584750_p3785534175420"></a><a name="zh-cn_topic_0000001538584750_p3785534175420"></a>/dev/davinci_manager</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p19759114295414"><a name="zh-cn_topic_0000001538584750_p19759114295414"></a><a name="zh-cn_topic_0000001538584750_p19759114295414"></a>davinci相关的设备管理设备。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row175343390145"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p8535939101418"><a name="zh-cn_topic_0000001538584750_p8535939101418"></a><a name="zh-cn_topic_0000001538584750_p8535939101418"></a>/dev/svm0</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p0535113961415"><a name="zh-cn_topic_0000001538584750_p0535113961415"></a><a name="zh-cn_topic_0000001538584750_p0535113961415"></a>内存管理的设备。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row081724113149"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p7817154161415"><a name="zh-cn_topic_0000001538584750_p7817154161415"></a><a name="zh-cn_topic_0000001538584750_p7817154161415"></a>/dev/ts_aisle</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p08177412149"><a name="zh-cn_topic_0000001538584750_p08177412149"></a><a name="zh-cn_topic_0000001538584750_p08177412149"></a>aicpudrv驱动设备，为任务调度提供事件驱动的渠道接口。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row97701421617"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1677111219613"><a name="zh-cn_topic_0000001538584750_p1677111219613"></a><a name="zh-cn_topic_0000001538584750_p1677111219613"></a>/dev/upgrade</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p4771821368"><a name="zh-cn_topic_0000001538584750_p4771821368"></a><a name="zh-cn_topic_0000001538584750_p4771821368"></a>驱动设备。</p>
<p id="zh-cn_topic_0000001538584750_p139858917612"><a name="zh-cn_topic_0000001538584750_p139858917612"></a><a name="zh-cn_topic_0000001538584750_p139858917612"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row19985159863"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p398514910612"><a name="zh-cn_topic_0000001538584750_p398514910612"></a><a name="zh-cn_topic_0000001538584750_p398514910612"></a>/dev/sys</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row68010161568"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1280616867"><a name="zh-cn_topic_0000001538584750_p1280616867"></a><a name="zh-cn_topic_0000001538584750_p1280616867"></a>/dev/vdec</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p1572782173418"><a name="zh-cn_topic_0000001538584750_p1572782173418"></a><a name="zh-cn_topic_0000001538584750_p1572782173418"></a>设备文件，支撑推理业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row512410151477"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1412417155713"><a name="zh-cn_topic_0000001538584750_p1412417155713"></a><a name="zh-cn_topic_0000001538584750_p1412417155713"></a>/dev/vpc</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row96243616713"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p362143610714"><a name="zh-cn_topic_0000001538584750_p362143610714"></a><a name="zh-cn_topic_0000001538584750_p362143610714"></a>/dev/pngd</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row196382414717"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1763815414719"><a name="zh-cn_topic_0000001538584750_p1763815414719"></a><a name="zh-cn_topic_0000001538584750_p1763815414719"></a>/dev/venc</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row16599816203215"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1659910166323"><a name="zh-cn_topic_0000001538584750_p1659910166323"></a><a name="zh-cn_topic_0000001538584750_p1659910166323"></a>/dev/dvpp_cmdlist</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row2279321123214"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p3279192113211"><a name="zh-cn_topic_0000001538584750_p3279192113211"></a><a name="zh-cn_topic_0000001538584750_p3279192113211"></a>/dev/log_drv</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p7279192111324"><a name="zh-cn_topic_0000001538584750_p7279192111324"></a><a name="zh-cn_topic_0000001538584750_p7279192111324"></a>日志驱动设备。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row9232145311019"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p223317531407"><a name="zh-cn_topic_0000001538584750_p223317531407"></a><a name="zh-cn_topic_0000001538584750_p223317531407"></a>/usr/local/Ascend/driver/lib64</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p62338531306"><a name="zh-cn_topic_0000001538584750_p62338531306"></a><a name="zh-cn_topic_0000001538584750_p62338531306"></a>目录，驱动提供的用户态库。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row1172341822420"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p7205142018244"><a name="zh-cn_topic_0000001538584750_p7205142018244"></a><a name="zh-cn_topic_0000001538584750_p7205142018244"></a>/usr/lib64/aicpu_kernels</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row127775519018"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p8777559012"><a name="zh-cn_topic_0000001538584750_p8777559012"></a><a name="zh-cn_topic_0000001538584750_p8777559012"></a>/usr/local/sbin/npu-smi</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p1377195514015"><a name="zh-cn_topic_0000001538584750_p1377195514015"></a><a name="zh-cn_topic_0000001538584750_p1377195514015"></a>文件，NPU-SMI工具。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row4981195619016"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p206061719115"><a name="zh-cn_topic_0000001538584750_p206061719115"></a><a name="zh-cn_topic_0000001538584750_p206061719115"></a>/etc/sys_version.conf</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p10601117412"><a name="zh-cn_topic_0000001538584750_p10601117412"></a><a name="zh-cn_topic_0000001538584750_p10601117412"></a>文件，驱动的版本信息。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row974117581204"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p165913171314"><a name="zh-cn_topic_0000001538584750_p165913171314"></a><a name="zh-cn_topic_0000001538584750_p165913171314"></a>/etc/ld.so.conf.d/mind_so.conf</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p658111714114"><a name="zh-cn_topic_0000001538584750_p658111714114"></a><a name="zh-cn_topic_0000001538584750_p658111714114"></a>驱动动态库路径配置文件</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row16158163414"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p20158935113"><a name="zh-cn_topic_0000001538584750_p20158935113"></a><a name="zh-cn_topic_0000001538584750_p20158935113"></a>/etc/hdcBasic.cfg</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p1915812317110"><a name="zh-cn_topic_0000001538584750_p1915812317110"></a><a name="zh-cn_topic_0000001538584750_p1915812317110"></a>文件，hdc基础文件。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row84221482011"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p124221581918"><a name="zh-cn_topic_0000001538584750_p124221581918"></a><a name="zh-cn_topic_0000001538584750_p124221581918"></a>/var/dmp_daemon</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p154226813114"><a name="zh-cn_topic_0000001538584750_p154226813114"></a><a name="zh-cn_topic_0000001538584750_p154226813114"></a>文件，dmp守护进程。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row116051118118"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p560531113116"><a name="zh-cn_topic_0000001538584750_p560531113116"></a><a name="zh-cn_topic_0000001538584750_p560531113116"></a>/var/slogd</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p46056111118"><a name="zh-cn_topic_0000001538584750_p46056111118"></a><a name="zh-cn_topic_0000001538584750_p46056111118"></a>文件，日志组件。</p>
</td>
</tr>
<tr id="row08099714240"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="p1759334492418"><a name="p1759334492418"></a><a name="p1759334492418"></a>/usr/lib64/libcrypto.so.1.1</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="p41371512123710"><a name="p41371512123710"></a><a name="p41371512123710"></a>文件，驱动所需动态库。</p>
<p id="p195937445243"><a name="p195937445243"></a><a name="p195937445243"></a><span id="ph3593184412244"><a name="ph3593184412244"></a><a name="ph3593184412244"></a>openEuler</span> 22.03或<span id="ph959324412415"><a name="ph959324412415"></a><a name="ph959324412415"></a>EulerOS</span> 2.11及以上需要挂载该路径。</p>
</td>
</tr>
<tr id="row5838189192417"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p2059315446244"><a name="p2059315446244"></a><a name="p2059315446244"></a>/usr/lib64/libyaml-0.so.2</p>
</td>
</tr>
<tr id="row325513118245"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="p959316445241"><a name="p959316445241"></a><a name="p959316445241"></a>/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="p469285716395"><a name="p469285716395"></a><a name="p469285716395"></a>文件，驱动所需动态库。</p>
<p id="p165931844152419"><a name="p165931844152419"></a><a name="p165931844152419"></a><span id="ph15593164414247"><a name="ph15593164414247"></a><a name="ph15593164414247"></a>Ubuntu</span> 22.04需要挂载该路径。</p>
<p id="p453474217249"><a name="p453474217249"></a><a name="p453474217249"></a></p>
</td>
</tr>
<tr id="row3534194217242"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p953494282412"><a name="p953494282412"></a><a name="p953494282412"></a>/usr/lib/aarch64-linux-gnu/libyaml-0.so.2</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row1666318232319"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p135803395414"><a name="zh-cn_topic_0000001538584750_p135803395414"></a><a name="zh-cn_topic_0000001538584750_p135803395414"></a>/usr/lib64/libsemanage.so.2</p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001538584750_p28911240182511"><a name="zh-cn_topic_0000001538584750_p28911240182511"></a><a name="zh-cn_topic_0000001538584750_p28911240182511"></a>文件，驱动所需动态库。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row71821847239"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p85796395418"><a name="zh-cn_topic_0000001538584750_p85796395418"></a><a name="zh-cn_topic_0000001538584750_p85796395418"></a>/usr/lib64/libmmpa.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row111170712310"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p25771539843"><a name="zh-cn_topic_0000001538584750_p25771539843"></a><a name="zh-cn_topic_0000001538584750_p25771539843"></a>/usr/lib64/libdrvdsmi.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row166051381237"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p3576113911410"><a name="zh-cn_topic_0000001538584750_p3576113911410"></a><a name="zh-cn_topic_0000001538584750_p3576113911410"></a>/usr/lib64/libdcmi.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row19801191230"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1057510391544"><a name="zh-cn_topic_0000001538584750_p1057510391544"></a><a name="zh-cn_topic_0000001538584750_p1057510391544"></a>/usr/lib64/libstackcore.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row472511122310"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1957413391646"><a name="zh-cn_topic_0000001538584750_p1957413391646"></a><a name="zh-cn_topic_0000001538584750_p1957413391646"></a>/usr/lib64/libmpi_dvpp_adapter.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row72932134239"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p65733390417"><a name="zh-cn_topic_0000001538584750_p65733390417"></a><a name="zh-cn_topic_0000001538584750_p65733390417"></a>/usr/lib64/libaicpu_scheduler.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row127171614192320"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p057283911412"><a name="zh-cn_topic_0000001538584750_p057283911412"></a><a name="zh-cn_topic_0000001538584750_p057283911412"></a>/usr/lib64/libaicpu_processer.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row970931410111"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p1357123911416"><a name="zh-cn_topic_0000001538584750_p1357123911416"></a><a name="zh-cn_topic_0000001538584750_p1357123911416"></a>/usr/lib64/libaicpu_prof.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row129961716612"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p5569133918412"><a name="zh-cn_topic_0000001538584750_p5569133918412"></a><a name="zh-cn_topic_0000001538584750_p5569133918412"></a>/usr/lib64/libaicpu_sharder.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row198131201110"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p12568739540"><a name="zh-cn_topic_0000001538584750_p12568739540"></a><a name="zh-cn_topic_0000001538584750_p12568739540"></a>/usr/lib64/libadump.so</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538584750_row2038118222116"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001538584750_p756723910412"><a name="zh-cn_topic_0000001538584750_p756723910412"></a><a name="zh-cn_topic_0000001538584750_p756723910412"></a>/usr/lib64/libtsd_eventclient.so</p>
</td>
</tr>
<tr id="row620342498"><td class="cellrowborder" valign="top" width="42.86%" headers="mcps1.2.3.1.1 "><p id="p157612531895"><a name="p157612531895"></a><a name="p157612531895"></a>/var/queue_schedule</p>
</td>
<td class="cellrowborder" valign="top" width="57.14%" headers="mcps1.2.3.1.2 "><p id="p2076218531290"><a name="p2076218531290"></a><a name="p2076218531290"></a>管理FlowGW调度框架。</p>
<div class="note" id="note177621753592"><a name="note177621753592"></a><div class="notebody"><p id="p476220537918"><a name="p476220537918"></a><a name="p476220537918"></a>挂载此目录需同时满足以下条件：</p>
<a name="ul1276295317916"></a><a name="ul1276295317916"></a><ul id="ul1276295317916"><li>MindCluster组件版本≥6.0.0。</li><li>HDK版本≥24.1.RC2。</li></ul>
</div></div>
</td>
</tr>
</tbody>
</table>

**表 6**  默认挂载目录和文件（Atlas 350 标卡）

|路径|说明|
|--|--|
|/dev/davinci*X*|NPU设备，X是ID号。例如：davinci0。|
|/dev/davinci_manager|管理设备。|
|/dev/hisi_hdc|管理设备。|
|/dev/uburma|管理设备，支持UB协议。当不支持UB协议时，不会挂载此设备。|
|/dev/ummu|管理设备，支持UB协议。当不支持UB协议时，不会挂载此设备。|
|/usr/local/Ascend/driver/lib64|目录，驱动提供的用户态库。|
|/usr/local/Ascend/driver/include|目录，驱动提供的头文件。|
|/usr/local/dcmi|目录，DCMI头文件和库。|
|/usr/local/bin/npu-smi|文件，NPU-SMI工具。|
|/etc/hccl_rootinfo.json|mindcluster-tools生成的rootinfo文件，该文件非必需挂载项。|
|/usr/local/Ascend/driver/topo|拓扑目录。|

**表 7**  默认挂载目录和文件（其他设备）

<a name="zh-cn_topic_0000001538584750_table3875124214592"></a>

|路径|说明|
|--|--|
|/dev/davinci*X*|NPU设备，X是ID号。例如：davinci0。|
|/dev/davinci_manager|管理设备。|
|/dev/hisi_hdc|管理设备。|
|/dev/devmm_svm|管理设备。|
|/usr/local/Ascend/driver/lib64|目录，驱动提供的用户态库。|
|/usr/local/Ascend/driver/include|目录，驱动提供的头文件。|
|/usr/local/dcmi|目录，DCMI头文件和库。|
|/usr/local/bin/npu-smi|文件，NPU-SMI工具。|
|/dev/dvpp_cmdlist|设备文件，支撑数字视觉预处理功能。|
|/var/queue_schedule|管理FlowGW调度框架。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>挂载此目录需同时满足以下条件：</p><ul><li>MindCluster组件版本≥6.0.0。</li><li>HDK版本≥24.1.RC2。</li></ul></div></div>|

## Ascend Docker Runtime默认挂载白名单

Ascend Docker Runtime通过ASCEND_RUNTIME_MOUNTS参数，提供了配置自定义默认挂载项的特性，具体操作请参考[（可选）配置自定义挂载内容](./usage/containerization.md#可选配置自定义挂载内容)章节。Ascend Docker Runtime的默认挂载项受白名单限制，具体白名单列表如[表1](#runtime_mount_white_list)所示。

**表 1**  默认挂载白名单列表

<a name="runtime_mount_white_list"></a>
<table>
  <thead>
    <tr>
      <th>路径</th>
      <th>说明</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>/usr/local/Ascend/driver/lib64</td>
      <td>目录，驱动提供的用户态库。</td>
    </tr>
    <tr>
      <td>/usr/local/Ascend/driver/include</td>
      <td>目录，驱动提供的头文件。</td>
    </tr>
    <tr>
      <td>/usr/local/dcmi</td>
      <td>目录，DCMI头文件和库。</td>
    </tr>
    <tr>
      <td>/usr/local/bin/npu-smi</td>
      <td>文件，NPU-SMI工具。</td>
    </tr>
    <tr>
      <td>/home/data/miniD/driver/lib64</td>
      <td rowspan="2">目录，驱动提供的用户态库。</td>
    </tr>
    <tr>
      <td>/usr/lib64/aicpu_kernels</td>
    </tr>
    <tr>
      <td>/usr/local/sbin/npu-smi</td>
      <td>文件，NPU-SMI工具。</td>
    </tr>
    <tr>
      <td>/usr/local/Ascend/driver/tools</td>
      <td>目录，驱动提供的工具包。</td>
    </tr>
    <tr>
      <td>/etc/hdcBasic.cfg</td>
      <td>文件，hdc基础文件。</td>
    </tr>
    <tr>
      <td>/etc/sys_version.conf</td>
      <td>文件，驱动的版本信息。</td>
    </tr>
    <tr>
      <td>/etc/ld.so.conf.d/mind_so.conf</td>
      <td>驱动动态库路径配置文件。</td>
    </tr>
    <tr>
      <td>/etc/slog.conf</td>
      <td>日志配置文件。</td>
    </tr>
    <tr>
      <td>/var/dmp_daemon</td>
      <td>文件，dmp守护进程。</td>
    </tr>
    <tr>
      <td>/var/slogd</td>
      <td>文件，日志组件。</td>
    </tr>
    <tr>
      <td>/usr/lib64/libsemanage.so.2</td>
      <td rowspan="16">文件，驱动所需动态库。</td>
    </tr>
    <tr>
      <td>/usr/lib64/libmmpa.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libcrypto.so.1.1</td>
    </tr>
    <tr>
      <td>/usr/lib64/libdrvdsmi.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libdcmi.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libstackcore.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libmpi_dvpp_adapter.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libaicpu_scheduler.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libaicpu_processer.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libaicpu_prof.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libaicpu_sharder.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libadump.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libtsd_eventclient.so</td>
    </tr>
    <tr>
      <td>/usr/lib64/libyaml-0.so.2</td>
      <td></td>
    </tr>
    <tr>
      <td>/usr/lib/aarch64-linux-gnu/libyaml-0.so.2</td>
      <td></td>
    </tr>
    <tr>
      <td>/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1</td>
      <td></td>
    </tr>
    <tr>
      <td>/var/queue_schedule</td>
      <td>管理FlowGW调度框架。</td>
    </tr>
    <tr>
      <td>/etc/hccl_rootinfo.json</td>
      <td>mindcluster-tools生成的rootinfo文件。</td>
    </tr>
    <tr>
      <td>/usr/local/Ascend/driver/topo</td>
      <td>拓扑目录。</td>
    </tr>
  </tbody>
</table>

## Ascend Docker Runtime命令说明<a name="ZH-CN_TOPIC_0000002511346347"></a>

Ascend Docker Runtime安装后，会在安装目录生成可执行工具，涉及到的指令为内部命令，用户请勿直接使用，相关指令如[表1](#zh-cn_topic_0000001538744718_table0615184315110)所示。

**表 1**  命令说明

<a name="zh-cn_topic_0000001538744718_table0615184315110"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001538744718_row061664319112"><th class="cellrowborder" valign="top" width="19.97%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001538744718_p46161543414"><a name="zh-cn_topic_0000001538744718_p46161543414"></a><a name="zh-cn_topic_0000001538744718_p46161543414"></a>工具名</p>
</th>
<th class="cellrowborder" valign="top" width="20.03%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001538744718_p9616343615"><a name="zh-cn_topic_0000001538744718_p9616343615"></a><a name="zh-cn_topic_0000001538744718_p9616343615"></a>短指令</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001538744718_p1561644313114"><a name="zh-cn_topic_0000001538744718_p1561644313114"></a><a name="zh-cn_topic_0000001538744718_p1561644313114"></a>长指令</p>
</th>
<th class="cellrowborder" valign="top" width="19.97%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001538744718_p17616443112"><a name="zh-cn_topic_0000001538744718_p17616443112"></a><a name="zh-cn_topic_0000001538744718_p17616443112"></a>其他参数类型</p>
</th>
<th class="cellrowborder" valign="top" width="20.03%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001538744718_p14616943811"><a name="zh-cn_topic_0000001538744718_p14616943811"></a><a name="zh-cn_topic_0000001538744718_p14616943811"></a>其他参数位置</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001538744718_row6616743117"><td class="cellrowborder" rowspan="6" valign="top" width="19.97%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p8811226233"><a name="zh-cn_topic_0000001538744718_p8811226233"></a><a name="zh-cn_topic_0000001538744718_p8811226233"></a>ascend-docker-cli</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p742716165314"><a name="zh-cn_topic_0000001538744718_p742716165314"></a><a name="zh-cn_topic_0000001538744718_p742716165314"></a>p</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p20427646339"><a name="zh-cn_topic_0000001538744718_p20427646339"></a><a name="zh-cn_topic_0000001538744718_p20427646339"></a>pid</p>
</td>
<td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p66161143214"><a name="zh-cn_topic_0000001538744718_p66161143214"></a><a name="zh-cn_topic_0000001538744718_p66161143214"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001538744718_p126165434112"><a name="zh-cn_topic_0000001538744718_p126165434112"></a><a name="zh-cn_topic_0000001538744718_p126165434112"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row106162432116"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p9427121619312"><a name="zh-cn_topic_0000001538744718_p9427121619312"></a><a name="zh-cn_topic_0000001538744718_p9427121619312"></a>r</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p54271046937"><a name="zh-cn_topic_0000001538744718_p54271046937"></a><a name="zh-cn_topic_0000001538744718_p54271046937"></a>rootfs</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p961684319116"><a name="zh-cn_topic_0000001538744718_p961684319116"></a><a name="zh-cn_topic_0000001538744718_p961684319116"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p136160431212"><a name="zh-cn_topic_0000001538744718_p136160431212"></a><a name="zh-cn_topic_0000001538744718_p136160431212"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row1616164318118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p194270161639"><a name="zh-cn_topic_0000001538744718_p194270161639"></a><a name="zh-cn_topic_0000001538744718_p194270161639"></a>o</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p104278460313"><a name="zh-cn_topic_0000001538744718_p104278460313"></a><a name="zh-cn_topic_0000001538744718_p104278460313"></a>options</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p961712432112"><a name="zh-cn_topic_0000001538744718_p961712432112"></a><a name="zh-cn_topic_0000001538744718_p961712432112"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p8617943419"><a name="zh-cn_topic_0000001538744718_p8617943419"></a><a name="zh-cn_topic_0000001538744718_p8617943419"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row166179431016"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p842771610316"><a name="zh-cn_topic_0000001538744718_p842771610316"></a><a name="zh-cn_topic_0000001538744718_p842771610316"></a>f</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p17427134619318"><a name="zh-cn_topic_0000001538744718_p17427134619318"></a><a name="zh-cn_topic_0000001538744718_p17427134619318"></a>mount-file</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p6617543816"><a name="zh-cn_topic_0000001538744718_p6617543816"></a><a name="zh-cn_topic_0000001538744718_p6617543816"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p136179438110"><a name="zh-cn_topic_0000001538744718_p136179438110"></a><a name="zh-cn_topic_0000001538744718_p136179438110"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row461724318120"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p144271816836"><a name="zh-cn_topic_0000001538744718_p144271816836"></a><a name="zh-cn_topic_0000001538744718_p144271816836"></a>l</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p1942774619319"><a name="zh-cn_topic_0000001538744718_p1942774619319"></a><a name="zh-cn_topic_0000001538744718_p1942774619319"></a>allow-link</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p1561712431611"><a name="zh-cn_topic_0000001538744718_p1561712431611"></a><a name="zh-cn_topic_0000001538744718_p1561712431611"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p14617194314116"><a name="zh-cn_topic_0000001538744718_p14617194314116"></a><a name="zh-cn_topic_0000001538744718_p14617194314116"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row116174431311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p1242791617310"><a name="zh-cn_topic_0000001538744718_p1242791617310"></a><a name="zh-cn_topic_0000001538744718_p1242791617310"></a>i</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p342712461838"><a name="zh-cn_topic_0000001538744718_p342712461838"></a><a name="zh-cn_topic_0000001538744718_p342712461838"></a>mount-dir</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p1961713431416"><a name="zh-cn_topic_0000001538744718_p1961713431416"></a><a name="zh-cn_topic_0000001538744718_p1961713431416"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p176171436115"><a name="zh-cn_topic_0000001538744718_p176171436115"></a><a name="zh-cn_topic_0000001538744718_p176171436115"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row186172438110"><td class="cellrowborder" rowspan="11" valign="top" width="19.97%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p34219818414"><a name="zh-cn_topic_0000001538744718_p34219818414"></a><a name="zh-cn_topic_0000001538744718_p34219818414"></a>ascend-docker-plugin-install-helper</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p19617943318"><a name="zh-cn_topic_0000001538744718_p19617943318"></a><a name="zh-cn_topic_0000001538744718_p19617943318"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p297862116415"><a name="zh-cn_topic_0000001538744718_p297862116415"></a><a name="zh-cn_topic_0000001538744718_p297862116415"></a>add</p>
</td>
<td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p14617204314111"><a name="zh-cn_topic_0000001538744718_p14617204314111"></a><a name="zh-cn_topic_0000001538744718_p14617204314111"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001538744718_p16618124312113"><a name="zh-cn_topic_0000001538744718_p16618124312113"></a><a name="zh-cn_topic_0000001538744718_p16618124312113"></a>1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row176188435113"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p46181643818"><a name="zh-cn_topic_0000001538744718_p46181643818"></a><a name="zh-cn_topic_0000001538744718_p46181643818"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p397811211640"><a name="zh-cn_topic_0000001538744718_p397811211640"></a><a name="zh-cn_topic_0000001538744718_p397811211640"></a>rm</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p1861819432013"><a name="zh-cn_topic_0000001538744718_p1861819432013"></a><a name="zh-cn_topic_0000001538744718_p1861819432013"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p161817436117"><a name="zh-cn_topic_0000001538744718_p161817436117"></a><a name="zh-cn_topic_0000001538744718_p161817436117"></a>1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row1261810431416"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p761810431616"><a name="zh-cn_topic_0000001538744718_p761810431616"></a><a name="zh-cn_topic_0000001538744718_p761810431616"></a>h</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p166187431316"><a name="zh-cn_topic_0000001538744718_p166187431316"></a><a name="zh-cn_topic_0000001538744718_p166187431316"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p81741521654"><a name="zh-cn_topic_0000001538744718_p81741521654"></a><a name="zh-cn_topic_0000001538744718_p81741521654"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p1261816431817"><a name="zh-cn_topic_0000001538744718_p1261816431817"></a><a name="zh-cn_topic_0000001538744718_p1261816431817"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row1061816436110"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p196181843512"><a name="zh-cn_topic_0000001538744718_p196181843512"></a><a name="zh-cn_topic_0000001538744718_p196181843512"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p126188439117"><a name="zh-cn_topic_0000001538744718_p126188439117"></a><a name="zh-cn_topic_0000001538744718_p126188439117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p10192061850"><a name="zh-cn_topic_0000001538744718_p10192061850"></a><a name="zh-cn_topic_0000001538744718_p10192061850"></a>destPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p76193433119"><a name="zh-cn_topic_0000001538744718_p76193433119"></a><a name="zh-cn_topic_0000001538744718_p76193433119"></a>2</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row76190431817"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p162019431814"><a name="zh-cn_topic_0000001538744718_p162019431814"></a><a name="zh-cn_topic_0000001538744718_p162019431814"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p7620114314118"><a name="zh-cn_topic_0000001538744718_p7620114314118"></a><a name="zh-cn_topic_0000001538744718_p7620114314118"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p1619136957"><a name="zh-cn_topic_0000001538744718_p1619136957"></a><a name="zh-cn_topic_0000001538744718_p1619136957"></a>srcPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p16620184311116"><a name="zh-cn_topic_0000001538744718_p16620184311116"></a><a name="zh-cn_topic_0000001538744718_p16620184311116"></a>3</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row8620124310116"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p186203431713"><a name="zh-cn_topic_0000001538744718_p186203431713"></a><a name="zh-cn_topic_0000001538744718_p186203431713"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p362012432111"><a name="zh-cn_topic_0000001538744718_p362012432111"></a><a name="zh-cn_topic_0000001538744718_p362012432111"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p1119206158"><a name="zh-cn_topic_0000001538744718_p1119206158"></a><a name="zh-cn_topic_0000001538744718_p1119206158"></a>installPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p1362004319110"><a name="zh-cn_topic_0000001538744718_p1362004319110"></a><a name="zh-cn_topic_0000001538744718_p1362004319110"></a>安装时为4</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row562014432017"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p362064312112"><a name="zh-cn_topic_0000001538744718_p362064312112"></a><a name="zh-cn_topic_0000001538744718_p362064312112"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p262018439118"><a name="zh-cn_topic_0000001538744718_p262018439118"></a><a name="zh-cn_topic_0000001538744718_p262018439118"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p61956955"><a name="zh-cn_topic_0000001538744718_p61956955"></a><a name="zh-cn_topic_0000001538744718_p61956955"></a>reserveDefault</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p1662017431915"><a name="zh-cn_topic_0000001538744718_p1662017431915"></a><a name="zh-cn_topic_0000001538744718_p1662017431915"></a>安装时为5，卸载时为4</p>
</td>
</tr>
<tr id="row29416591138"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p19942135921317"><a name="p19942135921317"></a><a name="p19942135921317"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12942105951317"><a name="p12942105951317"></a><a name="p12942105951317"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p420017310239"><a name="p420017310239"></a><a name="p420017310239"></a>installScene</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p82001435233"><a name="p82001435233"></a><a name="p82001435233"></a>安装时为6，卸载时为5</p>
</td>
</tr>
<tr id="row4729103441412"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1572913347142"><a name="p1572913347142"></a><a name="p1572913347142"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p972912348144"><a name="p972912348144"></a><a name="p972912348144"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p520033172318"><a name="p520033172318"></a><a name="p520033172318"></a>cgroupInfo</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p19200837239"><a name="p19200837239"></a><a name="p19200837239"></a>安装时为7，卸载时为6</p>
</td>
</tr>
<tr id="row1548213111410"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14828315142"><a name="p14828315142"></a><a name="p14828315142"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8482131181418"><a name="p8482131181418"></a><a name="p8482131181418"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15200637233"><a name="p15200637233"></a><a name="p15200637233"></a>osName</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p82017312312"><a name="p82017312312"></a><a name="p82017312312"></a>安装时为8，卸载时为7</p>
</td>
</tr>
<tr id="row885713161410"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p58571138148"><a name="p58571138148"></a><a name="p58571138148"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1685717351413"><a name="p1685717351413"></a><a name="p1685717351413"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1520163172314"><a name="p1520163172314"></a><a name="p1520163172314"></a>osVersion</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p0201736235"><a name="p0201736235"></a><a name="p0201736235"></a>安装时为9，卸载时为8</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row146209438117"><td class="cellrowborder" rowspan="2" valign="top" width="19.97%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p19762345867"><a name="zh-cn_topic_0000001538744718_p19762345867"></a><a name="zh-cn_topic_0000001538744718_p19762345867"></a>ascend-docker-runtime</p>
<p id="zh-cn_topic_0000001538744718_p156203435115"><a name="zh-cn_topic_0000001538744718_p156203435115"></a><a name="zh-cn_topic_0000001538744718_p156203435115"></a></p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p13620243516"><a name="zh-cn_topic_0000001538744718_p13620243516"></a><a name="zh-cn_topic_0000001538744718_p13620243516"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p24111859667"><a name="zh-cn_topic_0000001538744718_p24111859667"></a><a name="zh-cn_topic_0000001538744718_p24111859667"></a>create</p>
</td>
<td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p1362074311118"><a name="zh-cn_topic_0000001538744718_p1362074311118"></a><a name="zh-cn_topic_0000001538744718_p1362074311118"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001538744718_p8620204318119"><a name="zh-cn_topic_0000001538744718_p8620204318119"></a><a name="zh-cn_topic_0000001538744718_p8620204318119"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row26206433119"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p16620143414"><a name="zh-cn_topic_0000001538744718_p16620143414"></a><a name="zh-cn_topic_0000001538744718_p16620143414"></a>b</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p154110591664"><a name="zh-cn_topic_0000001538744718_p154110591664"></a><a name="zh-cn_topic_0000001538744718_p154110591664"></a>bundle</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p5620104318116"><a name="zh-cn_topic_0000001538744718_p5620104318116"></a><a name="zh-cn_topic_0000001538744718_p5620104318116"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p1862034312117"><a name="zh-cn_topic_0000001538744718_p1862034312117"></a><a name="zh-cn_topic_0000001538744718_p1862034312117"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row1962114431418"><td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p75417201577"><a name="zh-cn_topic_0000001538744718_p75417201577"></a><a name="zh-cn_topic_0000001538744718_p75417201577"></a>ascend-docker-destroy</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p66211443512"><a name="zh-cn_topic_0000001538744718_p66211443512"></a><a name="zh-cn_topic_0000001538744718_p66211443512"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p186217438110"><a name="zh-cn_topic_0000001538744718_p186217438110"></a><a name="zh-cn_topic_0000001538744718_p186217438110"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p4929659973"><a name="zh-cn_topic_0000001538744718_p4929659973"></a><a name="zh-cn_topic_0000001538744718_p4929659973"></a>cardId</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001538744718_p362104320116"><a name="zh-cn_topic_0000001538744718_p362104320116"></a><a name="zh-cn_topic_0000001538744718_p362104320116"></a>1</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row5621114319120"><td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p12621743815"><a name="zh-cn_topic_0000001538744718_p12621743815"></a><a name="zh-cn_topic_0000001538744718_p12621743815"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p66211143918"><a name="zh-cn_topic_0000001538744718_p66211143918"></a><a name="zh-cn_topic_0000001538744718_p66211143918"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p862116431113"><a name="zh-cn_topic_0000001538744718_p862116431113"></a><a name="zh-cn_topic_0000001538744718_p862116431113"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p2092913591074"><a name="zh-cn_topic_0000001538744718_p2092913591074"></a><a name="zh-cn_topic_0000001538744718_p2092913591074"></a>deviceId</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001538744718_p462194314112"><a name="zh-cn_topic_0000001538744718_p462194314112"></a><a name="zh-cn_topic_0000001538744718_p462194314112"></a>2</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001538744718_row06218435116"><td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001538744718_p1062113433110"><a name="zh-cn_topic_0000001538744718_p1062113433110"></a><a name="zh-cn_topic_0000001538744718_p1062113433110"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001538744718_p362124317120"><a name="zh-cn_topic_0000001538744718_p362124317120"></a><a name="zh-cn_topic_0000001538744718_p362124317120"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001538744718_p662114431119"><a name="zh-cn_topic_0000001538744718_p662114431119"></a><a name="zh-cn_topic_0000001538744718_p662114431119"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="19.97%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001538744718_p992910591075"><a name="zh-cn_topic_0000001538744718_p992910591075"></a><a name="zh-cn_topic_0000001538744718_p992910591075"></a>vDeviceId</p>
</td>
<td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001538744718_p36213439117"><a name="zh-cn_topic_0000001538744718_p36213439117"></a><a name="zh-cn_topic_0000001538744718_p36213439117"></a>3</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 
>
>- 因为Ascend Docker Runtime会将输入参数直接传递至runc或者docker-runc，所以runc/docker-runc的相关参数也会被Ascend Docker Runtime接受，用户请自行参考所在环境的runc/docker-runc的命令行选项使用相关参数。
>- ascend-docker-hook工具会忽略参数运行，运行时会接受标准输入。

## 容器和二进制部署差异<a name="ZH-CN_TOPIC_0000002479386442"></a>

**表 1**  组件两种安装方式差异

<a name="zh-cn_topic_0000001447284928_table19317191474517"></a>

|安装方式|差异|
|--|--|
|二进制|<ul><li>以系统服务的方式部署在物理机上。</li><li>配置Capability之后可以使用普通用户（hwMindX）运行。</li></ul>|
|容器|K8s作为调度管理平台，需要使用特权容器和root用户。|

## 使用ServiceAccount和KubeConfig差异<a name="ZH-CN_TOPIC_0000002511346357"></a>

**表 1** K8s认证授权方式差异

<a name="zh-cn_topic_0000001497205377_table75257815113"></a>

|认证凭据|组件|差异|
|--|--|--|
|ServiceAccount|<ul><li>Ascend Operator</li><li>Ascend Device Plugin</li><li>NodeD</li><li>Volcano</li><li>ClusterD</li></ul>|ServiceAccount的token文件内容会明文挂载到物理机上，有暴露风险。|
|导入的KubeConfig文件|Resilience Controller|通过集群调度组件提供的加密工具导入后为密文落盘，工具不提供解密导出功能，安全性较高。如果既配置了ServiceAccount，也导入了KubeConfig文件，后者优先级更高。|

## 高可用集群中的调度组件<a name="ZH-CN_TOPIC_0000002479226440"></a>

生产环境中，Kubernetes集群通常会部署多个管理节点，以避免单个管理节点故障导致整个集群不可用。Kubernetes官方提供了两种高可用的集群搭建方案，请参见[Kubernetes文档](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/ha-topology/)中高可用拓扑选项。集群调度组件基于官方“**Stacked etcd topology**”方案进行验证，各组件能够在多个管理节点场景下正常运行，且功能正常。

多管理节点场景需要保证所有管理节点配置一致，如集群调度组件镜像、日志目录、运行用户、节点标签等配置需一致。多管理节点下集群调度组件的安装请参见[安装部署](./installation_guide.md#安装部署)章节。多管理节点的安装请参见K8s官方文档[利用kubeadm创建高可用集群](https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/high-availability/)。

## 使用Containerd的集群调度组件<a name="ZH-CN_TOPIC_0000002479226400"></a>

Kubernetes  1.20版本之后，将不再支持使用Docker作为CRI（container runtime interface）。生产环境中，如果对使用的K8s有高版本要求时，需要考虑改用其它CRI。集群调度组件基于主流CRI  Containerd的1.4.4版本进行安装和验证，各组件能够在Containerd  +  Kubernetes场景下正常运行，且功能正常。

Containerd安装流程请参见[官方资料](https://github.com/containerd/containerd/blob/master/BUILDING.md)。集群调度组件安装时默认使用Ascend Docker Runtime，可以配置Containerd使用Ascend Docker Runtime替代runc，用于在启动容器时自动挂载设备，对Containerd需要做的配置请参见[安装部署](./installation_guide.md#安装部署)章节的Containerd场景下安装Ascend Docker Runtime。

## 模型训练任务说明<a name="ZH-CN_TOPIC_0000002479226458"></a>

使用其他调度器时，根据服务器类型，对训练任务的约束如下。当使用集群调度组件的Volcano作为调度器时，调度任务时已经满足如下使用约束。

**表 1**  训练任务使用说明

<a name="table2251851172715"></a>
<table><thead align="left"><tr id="row1426351202717"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.1"><p id="p1626105117274"><a name="p1626105117274"></a><a name="p1626105117274"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.2"><p id="p1226351182715"><a name="p1226351182715"></a><a name="p1226351182715"></a>训练场景</p>
</th>
<th class="cellrowborder" valign="top" width="60%" id="mcps1.2.4.1.3"><p id="p1926115110279"><a name="p1926115110279"></a><a name="p1926115110279"></a>使用说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row326155119271"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p32614515273"><a name="p32614515273"></a><a name="p32614515273"></a><span id="ph1181011812299"><a name="ph1181011812299"></a><a name="ph1181011812299"></a>Atlas 800 训练服务器（NPU满配）</span></p>
<p id="p102616512270"><a name="p102616512270"></a><a name="p102616512270"></a></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p5261151172716"><a name="p5261151172716"></a><a name="p5261151172716"></a>单机场景</p>
</td>
<td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p77141494309"><a name="p77141494309"></a><a name="p77141494309"></a>可申请NPU的数目为1、2、4、8。</p>
<p id="p185881543163017"><a name="p185881543163017"></a><a name="p185881543163017"></a>当申请NPU数目为2、4时，根据亲和性约束分配的NPU只能在同一台服务器同一个环内（0~3号NPU为一个环，4~7号NPU为一个环）。</p>
<p id="p102655116277"><a name="p102655116277"></a><a name="p102655116277"></a>例如申请了2个NPU进行训练，则分配的2个NPU要么都在同一台服务器的0~3号上或者都在4~7号上。不能出现一个在0~3号上，另一个在4~7号上。</p>
</td>
</tr>
<tr id="row52618519270"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1526175115279"><a name="p1526175115279"></a><a name="p1526175115279"></a>分布式场景</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p178849347316"><a name="p178849347316"></a><a name="p178849347316"></a>可申请NPU数目为1N、2N、4N、8N。</p>
<p id="p172611510279"><a name="p172611510279"></a><a name="p172611510279"></a>N表示节点个数，其中每个节点的NPU调度约束同单机场景。</p>
</td>
</tr>
<tr id="row1826951172711"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p826165112711"><a name="p826165112711"></a><a name="p826165112711"></a><span id="ph33362565317"><a name="ph33362565317"></a><a name="ph33362565317"></a>Atlas 800 训练服务器（NPU半配）</span></p>
<p id="p182705119278"><a name="p182705119278"></a><a name="p182705119278"></a></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p16616135918317"><a name="p16616135918317"></a><a name="p16616135918317"></a>单机场景</p>
</td>
<td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p182725132716"><a name="p182725132716"></a><a name="p182725132716"></a>可申请NPU的数目为1、2、4。</p>
</td>
</tr>
<tr id="row1727551132714"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p26161459183110"><a name="p26161459183110"></a><a name="p26161459183110"></a>分布式场景</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1956123214326"><a name="p1956123214326"></a><a name="p1956123214326"></a>可申请NPU数目为1N、2N、4N。N表示节点个数。</p>
</td>
</tr>
<tr id="row83031728327"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p030419223219"><a name="p030419223219"></a><a name="p030419223219"></a><span id="ph1435231416346"><a name="ph1435231416346"></a><a name="ph1435231416346"></a>Atlas 200T A2 Box16 异构子框</span></p>
<p id="p62201442326"><a name="p62201442326"></a><a name="p62201442326"></a></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p230416293218"><a name="p230416293218"></a><a name="p230416293218"></a>单机场景</p>
</td>
<td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p202281257143415"><a name="p202281257143415"></a><a name="p202281257143415"></a>可申请NPU的数目为1、2、3、4、5、6、7、8、10、12、14、16。</p>
<a name="ul14261149163515"></a><a name="ul14261149163515"></a><ul id="ul14261149163515"><li>当申请NPU数目小于8时，根据亲和性约束分配的NPU只能在同一台服务器同一个环内（0~7号NPU为一个环，8~16号NPU为一个环）。</li><li>当申请NPU数目为10、12、14时，需要将所需的NPU平均分配到两个环，相对的物理地址也一致。例如申请了2个NPU进行训练，则分配的2个NPU要么都在同一台服务器的0~7号上或者都在8~16号上。不能出现一个在0~7号上，另一个在8~16号上。</li></ul>
</td>
</tr>
<tr id="row20219245321"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p95591734153713"><a name="p95591734153713"></a><a name="p95591734153713"></a>分布式场景</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p881212548376"><a name="p881212548376"></a><a name="p881212548376"></a>可申请NPU的数目为1N、2N、3N、4N、5N、6N、7N、8N、10N、12N、14N、16N。</p>
<a name="ul153831715113813"></a><a name="ul153831715113813"></a><ul id="ul153831715113813"><li>N表示节点个数，其中每个节点的NPU调度约束同单机场景。</li><li>申请NPU的数目为10N、12N、14N时，需要将所需的NPU平均分配到两个环，相对的物理地址可以不一致。</li></ul>
</td>
</tr>
<tr id="row8392059113816"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p163965933816"><a name="p163965933816"></a><a name="p163965933816"></a><span id="ph20714203916"><a name="ph20714203916"></a><a name="ph20714203916"></a>Atlas 800T A2 训练服务器</span>或<span id="ph366416144394"><a name="ph366416144394"></a><a name="ph366416144394"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
<p id="p2077250203910"><a name="p2077250203910"></a><a name="p2077250203910"></a></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p16953101913920"><a name="p16953101913920"></a><a name="p16953101913920"></a>单机场景</p>
</td>
<td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p2402599384"><a name="p2402599384"></a><a name="p2402599384"></a>可申请NPU的数目为1、2、3、4、5、6、7、8。</p>
</td>
</tr>
<tr id="row1677130173911"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p19531919123915"><a name="p19531919123915"></a><a name="p19531919123915"></a>分布式场景</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p18772120173911"><a name="p18772120173911"></a><a name="p18772120173911"></a>可申请NPU的数目为1N、2N、3N、4N、5N、6N、7N、8N、16N。N表示节点个数。</p>
</td>
</tr>
<tr id="row140031116473"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p18993101513479"><a name="p18993101513479"></a><a name="p18993101513479"></a><span id="ph11548211143817"><a name="ph11548211143817"></a><a name="ph11548211143817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p425423854716"><a name="p425423854716"></a><a name="p425423854716"></a>单机场景</p>
</td>
<td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p19401181104712"><a name="p19401181104712"></a><a name="p19401181104712"></a>可申请NPU的数目为1、2、4、6、8、10、12、14、16。</p>
</td>
</tr>
<tr id="row17140171317472"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p0254143813479"><a name="p0254143813479"></a><a name="p0254143813479"></a>分布式场景</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p9141413194710"><a name="p9141413194710"></a><a name="p9141413194710"></a>可申请NPU的数目为2、4、6、8、10、12、14、16。若为逻辑超节点亲和任务，即任务YAML中的sp-block字段配置了逻辑超节点大小，则申请NPU的数目只能为16。</p>
</td>
</tr>
<tr id="row22120464408"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p1889218054115"><a name="p1889218054115"></a><a name="p1889218054115"></a>注：</p>
<p id="p4703164155014"><a name="p4703164155014"></a><a name="p4703164155014"></a>对不使用NPU的Pod，不做NPU数量的要求。</p>
</td>
</tr>
</tbody>
</table>

## Ascend Device Plugin通信文件<a name="ZH-CN_TOPIC_0000002479386384"></a>

**Socket文件通信<a name="zh-cn_topic_0000001446805124_section16108175614459"></a>**

在Ascend Device Plugin插件中，会生成用于通信的sock文件，sock文件的类型如下所示。

- npu.sock
- Ascend910.sock
- Ascend310.sock
- Ascend310P.sock
- davinci-mini.sock
- Ascend910-X.sock：X可取值为2c、4c、8c、16c、12c.3cpu.32g、12c.3cpu.32g.dvpp、12c.3cpu.32g.ndvpp、6c.1cpu.16g、3c.0.5cpu.8g、10c.3cpu.16g、10c.3cpu.16g.dvpp、10c.3cpu.16g.ndvpp、5c.1cpu.8g、4c.1cpu.5g、10c.3cpu.32g、10c.3cpu.32g.dvpp、10c.3cpu.32g.ndvpp和5c.1cpu.16g
- Ascend310P-X.sock：X可取值1c、2c、4c、2c.1cpu、4c.3cpu、4c.3cpu.ndvpp、4c.4cpu.dvpp

上述sock文件仅用于和本机的K8s通信。

## 芯片故障码参考文档<a name="ZH-CN_TOPIC_0000002511426289"></a>

各个产品芯片故障码的详细说明，可以参见[表1](#table87909405314)。

**表 1**  产品故障码参考文档

<a name="table87909405314"></a>

|产品形态|参考文档|
|--|--|
|Atlas 训练系列产品|<ul><li><span>《[Atlas 中心训练服务器 25.5.0 健康管理故障定义](https://support.huawei.com/enterprise/zh/doc/EDOC1100540096)》</span></li><li><span>《[Atlas 中心训练服务器 25.5.0 黑匣子错误码信息列表](https://support.huawei.com/enterprise/zh/doc/EDOC1100540113)》</span></li></ul>|
|<term>Atlas A2 训练系列产品</term>|<ul><li><span>《[Atlas A2 中心推理和训练硬件 25.5.0 健康管理故障定义](https://support.huawei.com/enterprise/zh/doc/EDOC1100540097)》</span></li><li><span>《[Atlas A2 中心推理和训练硬件 25.5.0 黑匣子错误码信息列表](https://support.huawei.com/enterprise/zh/doc/EDOC1100540117)》</span></li></ul>|
|<term>Atlas A3 训练系列产品</term>|<ul><li><span>《[Atlas A3 中心推理和训练硬件 25.5.0 健康管理故障定义](https://support.huawei.com/enterprise/zh/doc/EDOC1100540107)》</span></li><li><span>《[Atlas A3 中心推理和训练硬件 25.5.0 黑匣子错误码信息列表](https://support.huawei.com/enterprise/zh/doc/EDOC1100540120)》</span></li></ul>|
|推理服务器（插Atlas 300I 推理卡）|<span>《[Atlas 300I 推理卡 24.1.0 黑匣子错误码信息列表（型号 3000, 3010）](https://support.huawei.com/enterprise/zh/doc/EDOC1100438311)》</span>|
|Atlas 200I SoC A1 核心板|<ul><li><span>《[Atlas 200I SoC A1核心板 25.2.0 健康管理故障定义](https://support.huawei.com/enterprise/zh/doc/EDOC1100493983)》</span></li><li><span>《[Atlas 200I SoC A1核心板 25.2.0 黑匣子错误码信息列表](https://support.huawei.com/enterprise/zh/doc/EDOC1100493985)》</span></li></ul>|
|Atlas 推理系列产品（不包含Atlas 200I SoC A1 核心板）|<ul><li><span>《[Atlas 中心推理卡 25.5.0 健康管理故障定义](https://support.huawei.com/enterprise/zh/doc/EDOC1100540102)》</span></li><li><span>《[Atlas 中心推理卡 25.5.0 黑匣子错误码信息列表](https://support.huawei.com/enterprise/zh/doc/EDOC1100540099)》</span></li></ul>|

## 节点故障码参考文档<a name="ZH-CN_TOPIC_0000002479386430"></a>

各个产品节点故障码的详细说明，可以参见[表1](#table879094053145)。

**表 1**  节点故障码参考文档

<a name="table879094053145"></a>

|产品形态|参考文档|
|--|--|
|Atlas 800T A2 训练服务器|《[Atlas 800T A2 训练服务器 iBMC 告警处理](https://support.huawei.com/enterprise/zh/doc/EDOC1100317321)》|
|Atlas 900 A2 PoD 集群基础单元|《[Atlas 900 RCK A2 计算节点 iBMC 告警处理](https://support.huawei.com/enterprise/zh/doc/EDOC1100313926)》|

## 名词说明<a name="ZH-CN_TOPIC_0000002511426337"></a>

本章节介绍了集群调度组件用户指南中部分名词，方便用户更好的理解文档内容和步骤，名词介绍如[表1](#zh-cn_topic_0000001497364769_table163361144124517)所示。

**表 1**  名词解释

<a name="zh-cn_topic_0000001497364769_table163361144124517"></a>

|名词|说明|
|--|--|
|多机多卡训练|同时使用多台训练服务器上的多块芯片进行分布式训练。|
|单机单卡训练|使用1台训练服务器上的1颗芯片进行训练。|

>[!NOTE] 
>更多关于集群调度组件支持的产品形态，请参见[支持的产品形态和OS清单](./introduction.md#支持的产品形态和os清单)章节。

## 公网地址<a name="ZH-CN_TOPIC_0000002511346387"></a>

包含的公网地址请参见[MindCluster 7.3.0 公网地址.xlsx](../resource/MindCluster%207.3.0%20公网地址.xlsx)。

**表 1** 集群调度组件代码非公网地址说明

<a name="zh-cn_topic_0000001447124752_table52574541269"></a>

|url|说明|
|--|--|
|huawei.com/Ascend910|Atlas 训练系列产品资源名称，非网址，不访问。|
|huawei.com/Ascend310P|Atlas 推理系列产品资源名称，非网址，不访问。|
|huawei.com/Ascend310|Atlas 200/300/500 推理产品资源名称，非网址，不访问。|
|huawei.com/Ascend*|Ascend*切分芯片资源名称，非网址，不访问。|
|<https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.3>|注释参考信息，不访问。|
|huawei.com/Ascend310P-V|Atlas 推理系列产品混插模式：Atlas 300V 视频解析卡资源名称，非网址，不访问。|
|huawei.com/Ascend310P-VPro|Atlas 推理系列产品混插模式：Atlas 300V Pro 视频解析卡资源名称，非网址，不访问。|
|huawei.com/Ascend310P-IPro|Atlas 推理系列产品混插模式：Atlas 300I Pro 推理卡资源名称，非网址，不访问。|

## 安全说明<a name="ZH-CN_TOPIC_0000002479386374"></a>

断点续训展示的代码为开源代码，其中涉及到的脚本（Python以及shell）需要设置相同的用户和用户组。出于安全的考虑，建议用户对其中的输入参数、文件目录、文件路径等信息进行校验。

输入参数校验项目包括但不限于：

- 涉及使用外部变量作为命令的一部分都进行严格的参数校验和防注入措施。
- 从环境变量中获取的外部变量在用于命令拼接之前都要做严格的校验和防注入措施。
- 所有的进程理应最小权限原则，避免由于注入导致严重后果。
- 代码中不存在直接使用外部变量作为命令。
- 遵守各类编程语言安全规范。

文件路径校验项目包括但不限于：

- 路径长度有做限制。
- 路径有做特殊字符过滤和防绕过机制。
- 不存在命令注入。
- 进程满足最小权限原则。
- 白名单中不存在高危路径。
- 文件路径真实性有校验，有做抛异常处理。
- 命令注入是可控外部变量导致的非预期行为。
- 恢复策略只支持Python  3.7和Python  3.9版本。
- 脚本适配中，用户需要根据情况对异常进行捕获并按照业务逻辑处理。

## 用户信息列表<a name="ZH-CN_TOPIC_0000002479226450"></a>

请周期性地更新用户的密码，避免长期使用同一个密码带来的风险。

**系统用户<a name="zh-cn_topic_0000001515257736_zh-cn_topic_0000001446965016_section1069715500319"></a>**

|用户|描述|初始密码|密码修改方法|
|--|--|--|--|
|root|-|用户自定义|使用**passwd**命令修改。|
|HwHiAiUser|驱动run包的运行用户。|用户自定义|使用**passwd**命令修改。|
|hwMindX|集群调度组件默认的运行用户，默认设置为nologin。|无|-|
|HwBaseUser|Atlas 200I SoC A1 核心板上驱动相关设备运行用户，安装驱动时由驱动run包或者用户自行创建，默认设置为nologin。|无|-|
|HwDmUser|Atlas 200I SoC A1 核心板上驱动相关设备运行用户，安装驱动时由驱动run包或者用户自行创建，默认设置为nologin。|无|-|

**集群调度组件容器内用户<a name="zh-cn_topic_0000001515257736_zh-cn_topic_0000001446965016_section222461118323"></a>**

|用户|描述|初始密码|密码修改方法|
|--|--|--|--|
|root|-|无|-|
|HwHiAiUser|驱动run包的运行用户，非Atlas 200I SoC A1 核心板上的集群调度组件容器内默认为nologin，该用户不可登录。|无|-|
|hwMindX|集群调度组件容器内默认的运行用户，默认设置为nologin。|无|-|
|HwBaseUser|Atlas 200I SoC A1 核心板上驱动相关设备运行用户，集群调度组件容器内由用户自行创建。|无|-|
|HwDmUser|Atlas 200I SoC A1 核心板上驱动相关设备运行用户，集群调度组件容器内由用户自行创建。|无|-|

**nginx容器内用户（非安全加固场景不涉及）<a name="section1462355162610"></a>**

|用户|描述|初始密码|密码修改方法|
|--|--|--|--|
|nginx|nginx容器运行账户|无|-|

**Dockerfile示例中alpha基础镜像中的用户<a name="zh-cn_topic_0000001515257736_zh-cn_topic_0000001446965016_section16137174143318"></a>**

|用户|初始密码|密码修改方法|
|--|--|--|
|root|无|-|
|bin|无|-|
|daemon|无|-|
|adm|无|-|
|lp|无|-|
|sync|无|-|
|shutdown|无|-|
|halt|无|-|
|mail|无|-|
|news|无|-|
|uucp|无|-|
|operator|无|-|
|man|无|-|
|postmaster|无|-|
|cron|无|-|
|ftp|无|-|
|sshd|无|-|
|at|无|-|
|squid|无|-|
|xfs|无|-|
|games|无|-|
|cyrus|无|-|
|vpopmail|无|-|
|ntp|无|-|
|smmsp|无|-|
|guest|无|-|
|nobody|无|-|

**Dockerfile示例中Ubuntu基础镜像中的用户<a name="zh-cn_topic_0000001515257736_zh-cn_topic_0000001446965016_section158195363315"></a>**

|用户|初始密码|密码修改方法|
|--|--|--|
|root|无|-|
|daemon|无|-|
|bin|无|-|
|sys|无|-|
|sync|无|-|
|games|无|-|
|man|无|-|
|lp|无|-|
|mail|无|-|
|news|无|-|
|uucp|无|-|
|proxy|无|-|
|www-data|无|-|
|backup|无|-|
|list|无|-|
|irc|无|-|
|gnats|无|-|
|nobody|无|-|
|_apt|无|-|

**K8s的ServiceAccount<a name="section0422920124516"></a>**

**表 1**  组件在K8s中创建的ServiceAccount列表

<a name="table7715152119467"></a>

|账号名|说明|
|--|--|
|volcano-controllers|开源Volcano的controller组件在K8s中创建的用户。|
|volcano-scheduler|开源Volcano的scheduler组件在K8s中创建的用户。|
|ascend-device-plugin-sa-910|用YAML启动服务，将会在K8s中创建该用户，不同型号的设备使用的账号名不同。|
|ascend-device-plugin-sa-310p|用YAML启动服务，将会在K8s中创建该用户，不同型号的设备使用的账号名不同。|
|ascend-device-plugin-sa-310|用YAML启动服务，将会在K8s中创建该用户，不同型号的设备使用的账号名不同。|
|ascend-operator-manager|用YAML启动服务，将会在K8s中创建该用户，如：ascend-operator-v{version}.yaml。|
|resilience-controller|建议安全加固启动，使用带without-token的YAML启动服务，在K8s中创建并使用resilience-controller账号，同时为该账号授予适当权限。|
|noded|用YAML启动服务，将会在K8s中创建该用户，如：noded-v{version}.yaml。|
|clusterd|用YAML启动服务，将会在K8s中创建该用户，如：clusterd-v{version}.yaml。|
|default|MindCluster组件或开源Volcano部署时，会在K8s中自动创建的用户。|

## mindcluster-deploy开源仓版本说明<a name="ZH-CN_TOPIC_0000002511426311"></a>

[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)为MindCluster集群调度组件的开源仓库，仓库提供MindCluster示例代码和脚本仅供参考，不能用于生产环境。

代码仓版本配套关系说明如下：

**表 1**  mindcluster-deploy代码仓版本配套说明

<a name="table134454151315"></a>

|MindCluster版本|mindcluster-deploy仓配套分支|
|--|--|
|7.3.0|branch_v7.3.0|
|7.2.RC1|branch_v7.2.RC1|
|7.1.RC1|branch_v7.1.RC1|
|7.0.RC1|branch_v7.0.0-RC1|
|6.0.0|branch_v6.0.0|
|6.0.RC3|branch_v6.0.0-RC3|
|6.0.RC2|branch_v6.0.0-RC2|
|6.0.RC1，5.0.1，5.0.0，3.0.0|branch_v6.0.0-RC1|

## 进程级在线恢复验证<a name="ZH-CN_TOPIC_0000002511426299"></a>

本章节通过在训练代码中打桩构造片上内存的UCE故障，指导用户完成进程级在线恢复验证的适配步骤。

>[!NOTE] 
>
>- 本章节相关修改仅用于指导用户在测试环境下验证进程级在线恢复功能，切勿将此打桩版本上线到生产环境。
>- 配置本章节步骤前，请确保训练能正常拉起并已配置进程级在线恢复。
>- 为保证进程级在线恢复功能的正常使用，请将K8s集群master节点与worker节点的时钟保持一致。
>- 下文中代码可能与实际版本存在差异，请以实际版本代码为准。

### MindCluster适配<a name="ZH-CN_TOPIC_0000002479386410"></a>

1. <a name="li977718409381"></a>拉取MindCluster代码。

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone [https://gitcode.com/Ascend/mind-cluster.git](https://gitcode.com/Ascend/mind-cluster.git)
    cd ./mind-cluster/component/clusterd
    git checkout v7.3.0   # v7.3.0是代码仓版本tag，请自行切换到目标版本
    ```

2. 修改ClusterD代码。
    1. 打开“pkg/application/faultmanager/jobprocess/faultrank/job\_fault\_rank\_processor.go”文件。

        ```shell
        vi pkg/application/faultmanager/jobprocess/faultrank/job_fault_rank_processor.go
        ```

    2. 按“i”进入编辑模式，添加如下代码。

        ```go
        package faultrank
        
        import (
        …
            "clusterd/pkg/domain/faultdomain/collector"
        …
        )
        …
        func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(
        …
                if deviceDetail, ok := processor.retryInBusinessPlane(podInfo.jobId, nodeName, deviceName); ok {
              faultRankList = append(faultRankList, constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid,
                PodRank: podRankStr, FaultCode: faultdomain.GetRetryCodeByFaultType(deviceDetail.FaultType),
                FaultLevel:  constant.RestartBusiness,
                DoStepRetry: processor.canDoStepRetry(podInfo.jobId, nodeName, deviceName),
                DeviceId:    deviceInfo.DeviceID,
              })
              collector.ReportInfoCollector.ReportRetryInfo(podInfo.jobId, deviceInfo.RankID, constant.JobNotRecover, constant.UceFaultType)   // 业务面故障时间设置为无效时间，避免单次故障重复触发进程级在线恢复
            }
        …
        ```

    3. 按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。

3. <a name="li114977117517"></a>编译ClusterD。

    ```shell
    cd ./build/
    chmod +x build.sh && dos2unix build.sh
    sed -i 's|build_version="v[^"]\+"|build_version="xxx"|g' build.sh  # xxx替换为版本号，如v7.3.0
    sed -i 's|export CGO_ENABLED=0|export CGO_ENABLED=1|g' build.sh  # 开启CGO功能
    ./build.sh # 编译ClusterD，需要go 1.21及以上版本，建议使用1.21版本
    ```

    编译成功后，会在“../output/”目录下生成相关文件，可执行如下命令进行查看：

    ```shell
    ll ../output/
    ```

    回显示例如下：

    ```ColdFusion
    -r-x------. 1 root root 45891128 Aug 13 10:52 clusterd
    -r--------. 1 root root     4021 Aug 13 10:52 clusterd-v7.3.0.yaml
    -r--------. 1 root root      946 Aug 13 10:52 Dockerfile
    -r--------. 1 root root      209 Aug 13 10:52 faultDuration.json
    -r--------. 1 root root      207 Aug 13 10:52 fdConfig.yaml
    -r--------. 1 root root      467 Aug 13 10:52 publicFaultConfiguration.json
    -r--------. 1 root root      756 Aug 13 10:52 relationFaultCustomization.json
    ```

4. <a name="li89701053589"></a>进入output目录，制作ClusterD镜像。

    ```shell
    cd ../output/
    docker build --no-cache -t clusterd:{tag} ./  # {tag}与步骤3中build_version="xxx"的取值保持一致
    ```

5. （可选）保存镜像，并将保存后的镜像文件和clusterd-\{tag\}.yaml文件上传到主节点。若[步骤1](#li977718409381)到[步骤4](#li89701053589)在主节点执行，可跳过该步骤。

    ```shell
    docker save -o clusterd.tar clusterd:{tag}  #保存镜像
    docker load -i clusterd.tar  #在主节点导入镜像
    ```

6. 在主节点重新拉起ClusterD。

    ```shell
    kubectl delete -f  clusterd-{tag}.yaml  # 删除旧ClusterD容器
    kubectl apply -f  clusterd-{tag}.yaml  # 拉起新容器
    ```

### 脚本适配<a name="ZH-CN_TOPIC_0000002479226412"></a>

#### PyTorch场景适配示例（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002511426361"></a>

1. 搭建训练环境，拉起训练，详细请参见[PyTorch场景适配示例（基于MindSpeed-LLM）](./usage/resumable_training.md#适配示例)。
2. 开启进程级在线恢复，详细请参见[配置进程级在线恢复](./usage/resumable_training.md#配置进程级在线恢复)。
3. 在“QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py”代码中增加如下加粗内容，打桩注入故障，新增代码根据环境变量“RAISE\_UCE\_ERROR\_STEP\_AND\_RANK”获取注入故障迭代位置和故障rank信息。

    ```Python
    import os
    import ast
    …
    GLB_CNT = 0
     
    def train(forward_step_func, model, optimizer, opt_param_scheduler,
              train_data_iterator, valid_data_iterator,
              process_non_loss_data_func, config):
        """Train the model function."""
        args = get_args()
    timers = get_timers()
    …
        while iteration < args.train_iters:
            …
            num_microbatches = get_num_microbatches()
            update_num_microbatches(args.consumed_train_samples, consistency_check=True)
     
            global GLB_CNT
            cur_rank = torch.distributed.get_rank()
            uce_env = os.getenv("RAISE_UCE_ERROR_STEP_AND_RANK", "{}")
            uce_step_rank = ast.literal_eval(uce_env)
            if iteration in uce_step_rank and cur_rank == uce_step_rank[iteration] and GLB_CNT < iteration:
                GLB_CNT = iteration
                print(f"############# rank:{cur_rank} start UCE error #############")        
                raise RuntimeError('UCE ERROR')
     
            args.curr_iteration = iteration
            …
    ```

4. 修改启动脚本“QWEN3\_for\_PyTorch\_2.7\_code/scripts/train\_start.sh”。

    ```shell
    …
    export RAISE_UCE_ERROR_STEP_AND_RANK="{3:1,10:2}"  # 配置故障注入的迭代和卡号，在第3个迭代的rank 1卡和第10个迭代的rank 2卡上注入UCE故障
    sed -i 's/check_memory_result = torch_npu.npu.check_uce_in_memory(device)/check_memory_result = ha_constant.UCE_HIGH_LEVEL/g' /job/code/mindspeed_llm/core/high_availability/tft_stop_clean.py #修改PTA接口返回值，将训练代码抛出的异常识别为UCE故障
    …
    ```

#### MindSpore场景适配示例（基于MindFormers）<a name="ZH-CN_TOPIC_0000002511346369"></a>

1. 搭建训练环境，拉起训练，详细请参见[MindSpore场景适配示例（基于MindFormers）](./usage/resumable_training.md#适配示例)。
2. 开启进程级在线恢复，详细请参见[配置进程级在线恢复](./usage/resumable_training.md#配置进程级在线恢复)。
3. 在“QWEN3\_for\_MS\_code/mindformers/core/callback/callback.py”代码中增加如下内容，打桩注入故障。

    ```Python
    import json
    import os
    ...
    import ast
    GLB_CNT = 0
    EPOCH_CNT = 0
    ...
        def print_output_info(self, cb_params, cur_epoch_num, origin_epochs, throughput,
                              cur_step_num, steps_per_epoch, loss, per_step_seconds,
                              overflow, scaling_sens, time_remain, percent, global_norm):
            """print output information."""
            ...
            logger.info("  %4.1f%% %s %.5f samples/s/p  %s }", percent, show_str, throughput,
                        datetime.timedelta(seconds=int(time_remain)))
            global GLB_CNT
            global EPOCH_CNT
            if EPOCH_CNT < cur_epoch_num:
               GLB_CNT = 0
               EPOCH_CNT = cur_epoch_num
            uce_env = os.getenv("RAISE_UCE_ERROR_STEP_AND_RANK", "{}")
            uce_step_rank = ast.literal_eval(uce_env)
            if cur_step_num in uce_step_rank and get_rank() == uce_step_rank[cur_step_num] and GLB_CNT < cur_step_num:
               GLB_CNT = cur_step_num
               print(f"############# rank:{get_rank()} start UCE error #############")
               raise RuntimeError('UCEError occured.')
            if self.tensor_writer is not None:
                ...
    ```

4. 修改启动脚本“QWEN3\_for\_MS\_code/scripts/msrun\_launcher.sh”。

    ```shell
    …
    export RAISE_UCE_ERROR_STEP_AND_RANK="{3:1,10:2}"  # 配置故障注入的迭代和卡号，在第3个迭代的rank 1卡和第10个迭代的rank 2卡上注入UCE故障
    sed -i 's/err_strategy = _get_uce_process_strategy()/err_strategy = "RS_UCE_LOWLEVEL"/g' $(pip3 show mindspore | grep Location | awk -F ' ' '{print $2}')/mindspore/train/callback/_train_fault_tolerance.py #修改UCE处理策略
    …
    ```

## K8s集群基础性能调优<a name="ZH-CN_TOPIC_0000002511346319"></a>

MindCluster集群调度组件是基于K8s生态的功能组件，因此训练任务调度基于K8s平台时才支持使用断点续训。断点续训支持的K8s版本与MindCluster集群调度组件一致，当前为1.17.x\~1.34.x（推荐使用1.19.x及以上版本）。

>[!NOTE] 
>以下配置为万卡集群的推荐配置，实际配置时，请根据集群的规模进行调整。

**表 1**  配置说明

<a name="table47841609234"></a>
<table><thead align="left"><tr id="row197846082318"><th class="cellrowborder" valign="top" width="17.578242175782425%" id="mcps1.2.5.1.1"><p id="p47851805231"><a name="p47851805231"></a><a name="p47851805231"></a>配置项</p>
</th>
<th class="cellrowborder" valign="top" width="41.675832416758325%" id="mcps1.2.5.1.2"><p id="p47851703237"><a name="p47851703237"></a><a name="p47851703237"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25.51744825517448%" id="mcps1.2.5.1.3"><p id="p1078518052319"><a name="p1078518052319"></a><a name="p1078518052319"></a>推荐配置</p>
</th>
<th class="cellrowborder" valign="top" width="15.22847715228477%" id="mcps1.2.5.1.4"><p id="p9532854142416"><a name="p9532854142416"></a><a name="p9532854142416"></a>参考文件路径</p>
</th>
</tr>
</thead>
<tbody><tr id="row678511012234"><td class="cellrowborder" rowspan="2" valign="top" width="17.578242175782425%" headers="mcps1.2.5.1.1 "><p id="p1146512301552"><a name="p1146512301552"></a><a name="p1146512301552"></a>修改API Server</p>
<p id="p175123717238"><a name="p175123717238"></a><a name="p175123717238"></a>启动参数</p>
<p id="p3785160112316"><a name="p3785160112316"></a><a name="p3785160112316"></a></p>
</td>
<td class="cellrowborder" valign="top" width="41.675832416758325%" headers="mcps1.2.5.1.2 "><p id="p207408581462"><a name="p207408581462"></a><a name="p207408581462"></a>--max-request-inflight和--max-mutating-requests-inflight参数表示在给定时间内限制并行处理读写请求的最大数量限制。</p>
<p id="p277719526010"><a name="p277719526010"></a><a name="p277719526010"></a>若配置过低会出现请求超限错误，若配置过高会出现占用过多内存。</p>
</td>
<td class="cellrowborder" valign="top" width="25.51744825517448%" headers="mcps1.2.5.1.3 "><pre class="screen" id="screen19731459194413"><a name="screen19731459194413"></a><a name="screen19731459194413"></a>--max-request-inflight=20000
--max-mutating-requests-inflight=2000</pre>
</td>
<td class="cellrowborder" valign="top" width="15.22847715228477%" headers="mcps1.2.5.1.4 "><p id="p787221917258"><a name="p787221917258"></a><a name="p787221917258"></a>/etc/kubernetes/manifests/kube-apiserver.yaml</p>
</td>
</tr>
<tr id="row137854022316"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p8785406238"><a name="p8785406238"></a><a name="p8785406238"></a>--watch-cache和--watch-cache-sizes参数表示API Server的缓存量大小。</p>
<p id="p7515525461"><a name="p7515525461"></a><a name="p7515525461"></a>API Server获取etcd对象时，会优先访问本地cache，当cache中没有需要的信息时再访问etcd，并将etcd数据存入cache。若cache达到上限则覆盖cache，配置合理的cache大小可以提升etcd获取效率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><pre class="screen" id="screen16516174921615"><a name="screen16516174921615"></a><a name="screen16516174921615"></a>--watch-cache=true 
--watch-cache-sizes=node#1000,pod#2000,event#200,namespace#100,service#200</pre>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p8516749111613"><a name="p8516749111613"></a><a name="p8516749111613"></a>/etc/kubernetes/manifests/kube-apiserver.yaml</p>
<p id="p1053235417243"><a name="p1053235417243"></a><a name="p1053235417243"></a></p>
</td>
</tr>
<tr id="row47851014235"><td class="cellrowborder" valign="top" width="17.578242175782425%" headers="mcps1.2.5.1.1 "><p id="p1178512017239"><a name="p1178512017239"></a><a name="p1178512017239"></a>修改API Server资源配置</p>
</td>
<td class="cellrowborder" valign="top" width="41.675832416758325%" headers="mcps1.2.5.1.2 "><p id="p97856052315"><a name="p97856052315"></a><a name="p97856052315"></a>API Server配置的CPU资源将影响API Server的处理能力。</p>
</td>
<td class="cellrowborder" valign="top" width="25.51744825517448%" headers="mcps1.2.5.1.3 "><p id="p52391218132810"><a name="p52391218132810"></a><a name="p52391218132810"></a>API Server request的CPU资源上限调整为35核。</p>
<pre class="screen" id="screen1066354754917"><a name="screen1066354754917"></a><a name="screen1066354754917"></a>resources:
  requests:
    cpu: 35000m</pre>
<div class="note" id="note1956143192814"><a name="note1956143192814"></a><a name="note1956143192814"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p55611931112817"><a name="p55611931112817"></a><a name="p55611931112817"></a>API Server整体的CPU占用率不受此参数限制。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="15.22847715228477%" headers="mcps1.2.5.1.4 "><p id="p62971115275"><a name="p62971115275"></a><a name="p62971115275"></a>/etc/kubernetes/manifests/kube-apiserver.yaml</p>
</td>
</tr>
<tr id="row3786150112317"><td class="cellrowborder" rowspan="2" valign="top" width="17.578242175782425%" headers="mcps1.2.5.1.1 "><p id="p1678690152319"><a name="p1678690152319"></a><a name="p1678690152319"></a>修改etcd启动参数</p>
<p id="p4773205225"><a name="p4773205225"></a><a name="p4773205225"></a></p>
</td>
<td class="cellrowborder" valign="top" width="41.675832416758325%" headers="mcps1.2.5.1.2 "><p id="p7151173693015"><a name="p7151173693015"></a><a name="p7151173693015"></a>--quota-backend-bytes参数为etcd的存储上限，默认为2G。</p>
</td>
<td class="cellrowborder" valign="top" width="25.51744825517448%" headers="mcps1.2.5.1.3 "><p id="p7795115813293"><a name="p7795115813293"></a><a name="p7795115813293"></a>修改为8G。</p>
<pre class="screen" id="screen59032533500"><a name="screen59032533500"></a><a name="screen59032533500"></a>--quota-backend-bytes=8589934590</pre>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="15.22847715228477%" headers="mcps1.2.5.1.4 "><p id="p1052214919162"><a name="p1052214919162"></a><a name="p1052214919162"></a>/etc/kubernetes/manifests/etcd.yaml</p>
<p id="p25331854162415"><a name="p25331854162415"></a><a name="p25331854162415"></a></p>
</td>
</tr>
<tr id="row4773135823"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p2773351428"><a name="p2773351428"></a><a name="p2773351428"></a>--auto-compaction-retention：进行自动压缩，降低资源占用。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p570371410337"><a name="p570371410337"></a><a name="p570371410337"></a>进行碎片整理，降低资源占用。</p>
<pre class="screen" id="screen15970531113311"><a name="screen15970531113311"></a><a name="screen15970531113311"></a>--auto-compaction-retention</pre>
<div class="note" id="note989123811420"><a name="note989123811420"></a><a name="note989123811420"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p188918381649"><a name="p188918381649"></a><a name="p188918381649"></a>--auto-compaction-retention不会实际释放空间，需要用户手动配合使用etcdctl compact和etcd defrag清理空间。</p>
</div></div>
</td>
</tr>
<tr id="row0820640122718"><td class="cellrowborder" valign="top" width="17.578242175782425%" headers="mcps1.2.5.1.1 "><p id="p138203403275"><a name="p138203403275"></a><a name="p138203403275"></a>修改etcd资源配置</p>
</td>
<td class="cellrowborder" valign="top" width="41.675832416758325%" headers="mcps1.2.5.1.2 "><p id="p7820140192711"><a name="p7820140192711"></a><a name="p7820140192711"></a>etcd配置的CPU和内存资源将影响etcd的处理能力。</p>
</td>
<td class="cellrowborder" valign="top" width="25.51744825517448%" headers="mcps1.2.5.1.3 "><p id="p196391628173210"><a name="p196391628173210"></a><a name="p196391628173210"></a>etcd request的CPU资源上限调整为20核，memory资源上限调整为10G。</p>
<pre class="screen" id="screen4260163245215"><a name="screen4260163245215"></a><a name="screen4260163245215"></a>resources:
  requests:
    cpu: 20000m
    memory: 10000Mi</pre>
</td>
<td class="cellrowborder" valign="top" width="15.22847715228477%" headers="mcps1.2.5.1.4 "><p id="li281135811710p0"><a name="li281135811710p0"></a><a name="li281135811710p0"></a>/etc/kubernetes/manifests/etcd.yaml</p>
<p id="p38202040102713"><a name="p38202040102713"></a><a name="p38202040102713"></a></p>
</td>
</tr>
<tr id="row19820114015278"><td class="cellrowborder" valign="top" width="17.578242175782425%" headers="mcps1.2.5.1.1 "><p id="p18967164583211"><a name="p18967164583211"></a><a name="p18967164583211"></a>修改Volcano资源配置</p>
</td>
<td class="cellrowborder" valign="top" width="41.675832416758325%" headers="mcps1.2.5.1.2 "><p id="p18820154062717"><a name="p18820154062717"></a><a name="p18820154062717"></a>Volcano配置的CPU和内存资源将影响Volcano的处理能力。</p>
</td>
<td class="cellrowborder" valign="top" width="25.51744825517448%" headers="mcps1.2.5.1.3 "><p id="p3821140192710"><a name="p3821140192710"></a><a name="p3821140192710"></a>Volcano request的CPU资源上限调整为20核，memory资源上限调整为8G。</p>
<pre class="screen" id="screen6620102175310"><a name="screen6620102175310"></a><a name="screen6620102175310"></a>resources:
  requests:
    cpu: 20000m
    memory: 4Gi</pre>
</td>
<td class="cellrowborder" valign="top" width="15.22847715228477%" headers="mcps1.2.5.1.4 "><div class="p" id="p18444555165312"><a name="p18444555165312"></a><a name="p18444555165312"></a>参考配置命令：<pre class="screen" id="screen3754120691"><a name="screen3754120691"></a><a name="screen3754120691"></a>kubectl edit deployment -n volcano-system  volcano-scheduler</pre>
</div>
<p id="p1282144032717"><a name="p1282144032717"></a><a name="p1282144032717"></a></p>
</td>
</tr>
</tbody>
</table>

## 自定义指标开发<a name="ZH-CN_TOPIC_0000002512192053"></a>

支持通过如下两种方式开发自定义指标。

- 通过文件方式开发自定义指标

    用户根据[自定义指标文件](./api/npu_exporter.md#自定义指标文件)，创建符合要求的自定义指标文件。启动NPU Exporter时，配置“-textMetricsFilePath”参数，指定该自定义指标文件的路径。详情请参见[NPU Exporter](./installation_guide.md#npu-exporter)中“NPU Exporter启动参数”表。NPU Exporter会在每个数据采集周期读取自定义指标文件，并将文件内容上报给Prometheus或Telegraf。

    开发示例如下：

    使用NPU Exporter集成并采集Devkit工具生成的hccs\_bandwidth指标，详情请参见[NPU Exporter集成Devkit部署指南](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/utils/npu-exporter)。关于hccs\_bandwidth指标信息的说明请参见[HCCS带宽监控](https://www.hikunpeng.com/document/detail/zh/kunpengdevps/userguide/cliuserguide/KunpengDevKitCli_0251.html)。

- 通过插件方式开发自定义指标

    用户可通过编写插件的方式自定义指标，使用该插件前，开发者需要自行学习了解cgo、go相关语言特性，并阅读[README](https://gitcode.com/Ascend/mind-cluster/blob/master/component/npu-exporter/plugins/README.md)了解使用方法。

>[!NOTICE] 
>
>- 自定义的指标不能与已有的指标名重复。
>- 开发者需对自定义插件的稳定性负责，确保不引入运行时panic等问题。
>- 开发者需要对自定义指标文件格式的正确性负责。

## 修订记录<a name="ZH-CN_TOPIC_0000002479386422"></a>

<a name="table11921168962"></a>

|发布日期|修订说明|
|--|--|
|2026-01-19|第一次正式发布|
