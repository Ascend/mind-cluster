# NPU Exporter<a name="ZH-CN_TOPIC_0000002511426331"></a>

- 使用**资源监测**时，必须安装NPU Exporter，该组件支持对接Prometheus或Telegraf。
    - 对接Prometheus时，支持通过镜像和二进制两种方式部署NPU Exporter，部署差异可参考[镜像和二进制部署差异](../../../07_references/appendix.md#镜像和二进制部署差异)。
    - 对接Telegraf时，参考[通过Telegraf使用](../../../04_usage/resource_monitoring/03_working_with_telegraf.md)章节，安装NPU Exporter和Telegraf。

- 不使用**资源监测**的用户，可以不安装NPU Exporter，请直接跳过本章节。

## 使用约束<a name="section1362795652416"></a>

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
<div class="note" id="note1993172317415"><a name="note1993172317415"></a><a name="note1993172317415"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><div class="p" id="zh-cn_topic_0000002479226378_p18934232419"><a name="zh-cn_topic_0000002479226378_p18934232419"></a><a name="zh-cn_topic_0000002479226378_p18934232419"></a>为保证<span id="zh-cn_topic_0000002479226378_ph7206429154119"><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a>NPU Exporter</span>以二进制部署时可使用非root用户安装（如hwMindX），请在安装驱动时使用--install-for-all参数。示例如下。<pre class="screen" id="zh-cn_topic_0000002479226378_screen15239164112445"><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a>./Ascend-hdk-&lt;chip_type&gt;-npu-driver_&lt;version&gt;_linux-&lt;arch&gt;.run --full --install-for-all</pre>
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
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p209012054192411"><a name="p209012054192411"></a><a name="p209012054192411"></a><span id="ph56561935182816"><a name="ph56561935182816"></a><a name="ph56561935182816"></a>Atlas 200I SoC A1 核心板</span>使用<span id="ph1865633562811"><a name="ph1865633562811"></a><a name="ph1865633562811"></a>NPU Exporter</span>组件，需要确保<span id="ph10656153513282"><a name="ph10656153513282"></a><a name="ph10656153513282"></a>Atlas 200I SoC A1 核心板</span>的NPU驱动在23.0.RC2及以上版本。升级NPU驱动可参考<span id="ph19001377278"><a name="ph19001377278"></a><a name="ph19001377278"></a>《Atlas 200I SoC A1 核心板 NPU驱动和固件升级指导书》中“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100493481/b35e85d1" target="_blank" rel="noopener noreferrer">升级驱动</a>”章节</span>进行操作。</p>
</td>
</tr>
<tr id="row165073518272"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p95251515257"><a name="p95251515257"></a><a name="p95251515257"></a><span id="ph19614124172819"><a name="ph19614124172819"></a><a name="ph19614124172819"></a>Atlas 200I SoC A1 核心板</span>节点上使用镜像方式部署<span id="ph136141041142813"><a name="ph136141041142813"></a><a name="ph136141041142813"></a>NPU Exporter</span>，需要配置多容器共享模式，具体请参考<span id="ph3957123242310"><a name="ph3957123242310"></a><a name="ph3957123242310"></a>《Atlas 200I SoC A1 核心板 NPU驱动和固件安装指南》中“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100493510/55e9d968" target="_blank" rel="noopener noreferrer">容器内运行</a>”章节</span>。</p>
</td>
</tr>
<tr id="row1044710113298"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p1144701142912"><a name="p1144701142912"></a><a name="p1144701142912"></a>虚拟机场景</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p14473110297"><a name="p14473110297"></a><a name="p14473110297"></a>如果在虚拟机场景下部署<span id="ph6368151492319"><a name="ph6368151492319"></a><a name="ph6368151492319"></a>NPU Exporter</span>，需要在<span id="ph24388313372"><a name="ph24388313372"></a><a name="ph24388313372"></a>NPU Exporter</span>的镜像中安装systemd，推荐在Dockerfile中加入<strong id="b14813193310547"><a name="b14813193310547"></a><a name="b14813193310547"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong>命令进行安装。</p>
</td>
</tr>
<tr id="row_container_metrics"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p_container_metrics_cat">容器相关指标</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><ol id="ol_container_metrics"><li>仅支持Prometheus场景下，NPU卡挂载到K8s的Pod中时展示相关指标，不支持Telegraf场景展示容器相关指标。</li><li>Prometheus场景下，若NPU卡未挂载到K8s的Pod中，则各指标label中的container_name、namespace、pod_name为空。</li><li>解析Pod挂载NPU卡信息时，优先解析容器中是否存在ASCEND_VISIBLE_DEVICE环境变量，请确保镜像中未预置该环境变量，避免干扰。</li><li>请勿使用特权模式启动业务容器，如使用特权启动业务容器，可能会出现业务进程实际使用的NPU卡和MindCluster组件分配的NPU卡不一致的情况。</li></ol>
</td>
</tr>
</tbody>
</table>

## 兼容性说明

#### 26.1.0 兼容性说明

- 配置文件路径变更，`metricConfiguration.json`和`pluginConfiguration.json`路径由`/usr/local/`变更为`/user/mind-cluster/npu-exporter-config`。

- 如下指标增加`-1`状态，表示dcmi接口/hccn_tool工具调用失败，具体指标信息请参考[NPU数据信息](../../../06_api/npu_exporter/01_prometheus_metrics_api.md#section1379685784314)：
    - `npu_chip_info_health_status`
    - `npu_chip_info_network_status`
    - `npu_chip_info_link_status`
    - `npu_chip_info_link_status_X_Y`

- 将如下指标由npu指标组迁移到utilization指标组，默认采集周期为1s，具体指标信息请参考[利用率数据信息](../../../06_api/npu_exporter/01_prometheus_metrics_api.md#section1379685784315)：
    - `npu_chip_info_utilization`
    - `container_npu_utilization`
    - `npu_chip_info_vector_utilization`
    - `npu_chip_info_cube_utilization`
    - `npu_chip_info_overall_utilization`

- 取消默认的`-updateTime=5`参数，由配置文件中各指标组的`intervalSeconds`控制对应指标组的采集间隔。
    - 如仍然配置了`-updateTime`参数且参数范围有效，则优先使用该参数值。

- 调整各指标组默认采集周期，由原来的全局5s调整为如下：

  | 采集周期       | 指标组                                                    |
    |------------|--------------------------------------------------------|
  | 只采集一次      | version                                                |
  | 1s         | utilization                                            |
  | 5s         | npu                                                    |
  | 10s        | ddr                                                    |
  | 60s        | sio, hbm, hccs, pcie, vnpu, roce, optical, network, ub |
  | 86400s(一天) | nodeBase                                               |


## 操作步骤<a name="section83111543151612"></a>

NPU Exporter支持两种安装方式，用户可根据实际情况选择其中一种进行安装。该组件仅提供HTTP服务，如需使用更为安全的HTTPS服务，请自行修改源码进行适配。

- （推荐）以镜像方式运行，安装步骤参见[镜像方式运行](#section2035402135914)。
- 当安全要求较高时，建议在物理机上以二进制方式运行，安装步骤参见[二进制方式运行](#section103551921135917)。

## 镜像方式运行<a name="section2035402135914"></a>

1. 以root用户登录各计算节点。
2. （可选）修改metricConfiguration.json或pluginConfiguration.json文件，配置默认指标组或自定义指标组的采集开关和采集周期。
    1. 进入NPU Exporter软件包解压目录。
    2. <a name="li11364381194"></a>打开metricConfiguration.json文件。

        ```shell
        vi metricConfiguration.json
        ```

    3. 按“i”进入编辑模式，根据实际需要配置默认指标组的采集开关和采集周期。

        配置文件示例如下：

        ```json
            [
                {"metricsGroup": "version", "state": "ON", "intervalSeconds": -1},
                {"metricsGroup": "utilization", "state": "ON", "intervalSeconds": 1},
                {"metricsGroup": "npu", "state": "ON", "intervalSeconds": 5},
                {"metricsGroup": "ddr", "state": "ON", "intervalSeconds": 10},
                {"metricsGroup": "sio", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hbm", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hccs", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "pcie", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "vnpu", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "nodeBase", "state": "ON", "intervalSeconds": 86400},

                {"metricsGroup": "roce", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "optical", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "network", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "ub", "state": "ON", "intervalSeconds": 60}
            ]
        ```

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
        <a name="ul222055714012"></a><a name="ul222055714012"></a>
           <ul id="zh-cn_topic_0000002511426331_ul222055714012">
               通过dcmi采集：
               <li>version：版本数据信息</li>
               <li>utilization：使用率数据信息</li>
               <li>npu：NPU数据信息</li>
               <li>ddr：DDR数据信息</li>
               <li>sio：SIO数据信息</li>
               <li>hbm：片上内存数据信息</li>
               <li>hccs：HCCS数据信息</li>
               <li>pcie：PCIe数据信息</li>
               <li>vnpu：vNPU数据信息</li>
               <li>nodeBase：节点基本信息</li>
           </ul>
           <br/>
           <ul>
               通过hccn_tool采集：
               <li>roce：RoCE数据信息</li>
               <li>optical：光模块数据信息</li>
               <li>network：Network数据信息</li>
               <li>ub：NPU UB数据信息</li>
           </ul>
        </td>
        </tr>
        <tr id="row5220257114014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="p182201657134015"><a name="p182201657134015"></a><a name="p182201657134015"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="p722015718403"><a name="p722015718403"></a><a name="p722015718403"></a>指标组采集和上报的开关。默认值为ON。</p>
        <a name="ul14220557134016"></a><a name="ul14220557134016"></a><ul id="ul14220557134016"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        <tr id="rowinterval"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="pinterval"><a name="pinterval"></a><a name="pinterval"></a>intervalSeconds</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="pintervaldesc"><a name="pintervaldesc"></a><a name="pintervaldesc"></a>指标组采集周期，单位为秒。</p>
        <a name="ulinterval"></a><a name="ulinterval"></a><ul id="ulinterval"><li>必须配置为整数值</li><li>取值范围为-1、1~86400秒</li><li>若缺失该配置项，使用默认值60秒</li><li>若配置为-1，表示该指标组只采集一次，不再重复采集</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    4. <a name="li151815494115"></a>按“Esc”键，输入:wq!保存并退出。
    5. 参考[2.b](#li11364381194)到[2.d](#li151815494115)，修改pluginConfiguration.json文件，根据实际需要配置自定义指标组的采集开关和采集周期。

        <a name="table970154420512"></a>
        <table><thead align="left"><tr id="row157015443510"><th class="cellrowborder" valign="top" width="23.14%" id="mcps1.1.3.1.1"><p id="p15701444553"><a name="p15701444553"></a><a name="p15701444553"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="76.86%" id="mcps1.1.3.1.2"><p id="p117011441156"><a name="p117011441156"></a><a name="p117011441156"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="row47010440518"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="p170118446517"><a name="p170118446517"></a><a name="p170118446517"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="p9568105120719"><a name="p9568105120719"></a><a name="p9568105120719"></a>向<span id="ph18671058476"><a name="ph18671058476"></a><a name="ph18671058476"></a>NPU Exporter</span>注册的自定义指标组名称。自定义指标的方法详细请参见<a href="../../custom_metrics_development.md">自定义指标开发</a>。</p>
        </td>
        </tr>
        <tr id="row157010441654"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="p157021644755"><a name="p157021644755"></a><a name="p157021644755"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="p2148170172513"><a name="p2148170172513"></a><a name="p2148170172513"></a>指标组采集和上报的开关。默认值为OFF。</p>
        <a name="ul1870217441514"></a><a name="ul1870217441514"></a><ul id="ul1870217441514"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        <tr id="rowplugininterval"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="pplugininterval"><a name="pplugininterval"></a><a name="pplugininterval"></a>intervalSeconds</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="ppluginintervaldesc"><a name="ppluginintervaldesc"></a><a name="ppluginintervaldesc"></a>自定义指标组采集周期，单位为秒。</p>
        <a name="ulplugininterval"></a><a name="ulplugininterval"></a><ul id="ulplugininterval"><li>必须配置为整数值</li><li>取值范围为-1、1~86400秒</li><li>若缺失该配置项，使用默认值60秒</li><li>若配置为-1，表示该指标组只采集一次，不再重复采集</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    6. 若通过插件方式开发了自定义指标，需重新构建编译二进制文件。
    7. 参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，重新进行镜像制作和分发。

3. （可选）配置文件挂载说明。

    配置文件挂载情况有如下三种，具体的配置方法请参考[动态配置加载说明](#动态配置加载说明)章节。

    - **默认情况**：默认不挂载配置文件到宿主机，使用镜像内置的默认配置，**无法实现动态配置修改**。
    - **HostPath挂载**：将宿主机上的配置文件挂载到容器中，路径为`/user/mind-cluster/npu-exporter-config/metricConfiguration.json`和`/user/mind-cluster/npu-exporter-config/pluginConfiguration.json`。此方式支持每个节点独立配置，也可通过共享目录实现全局统一配置。
    - **ConfigMap挂载**：通过K8s ConfigMap统一管理配置文件，所有节点使用相同配置。

4. 查看NPU Exporter镜像和版本号是否正确。
    - **Docker场景**：执行如下命令。

        ```shell
        docker images | grep npu-exporter
        ```

        回显示例如下。

        ```ColdFusion
        npu-exporter                         v26.0.0              20185c45f1bc        About an hour ago         90.1MB
        ```

    - **Containerd场景**：执行如下命令。

        ```shell
        ctr -n k8s.io c ls | grep npu-exporter
        ```

        回显示例如下。

        ```ColdFusion
        docker.io/library/npu-exporter:v26.0.0                                                         application/vnd.docker.distribution.manifest.v2+json      sha256:38fd69ee9f5753e73a55a216d039f6ed4ea8a5de15c0e6b3bb503022db470c7b 91.5 MiB  linux/arm64
        ```

    - 是，执行[步骤5](#li0640635114211)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

5. <a name="li0640635114211"></a>将NPU Exporter软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
6. 请根据实际使用的镜像方式，选择执行以下步骤。
    - **Containerd场景**：需要将containerMode设置为containerd，并对以下加粗代码进行修改。

        如果使用默认的NPU Exporter启动参数“-containerMode=docker”时，可跳过本步骤。

        <pre codetype="yaml">
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
                         -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log -logLevel=0 <strong>-containerMode=containerd</strong>" ]
        ...
              volumeMounts:
        ...
                - name: docker-shim
                  mountPath: /var/run/dockershim.sock
                  readOnly: true
                <strong>- name: docker                                       # 仅使用containerd时删除</strong>
                  <strong>mountPath: /var/run/docker</strong>
                  <strong>readOnly: true</strong>
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
                <strong>- name: docker                                # 仅使用containerd时删除</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/docker</strong>
                - name: cri-dockerd
                  hostPath:
                    path: /var/run/cri-dockerd.sock
                - name: containerd
                  hostPath:
                    path: /run/containerd
                - name: isulad
                  hostPath:
                    path: /run/isulad.sock

        ...</pre>

    - **Docker场景**：删除原有容器运行时的挂载文件，新增dockershim.sock文件的挂载目录，并对以下加粗代码进行修改。

        如果使用的NPU Exporter启动参数“-containerMode=containerd”，可跳过本步骤。

        >[!NOTICE]
        >该步骤可有效解决kubelet重启后，造成的NPU Exporter数据丢失问题。新增挂载目录后，会同时新增很多挂载文件，如docker.sock，有容器逃逸的风险。

        <pre codetype="yaml">
        ...
                volumeMounts:
                  - name: log-npu-exporter
        ...
                  - name: sys
                    mountPath: /sys
                    readOnly: true
                  <strong>- name: docker-shim                        # 删除以下加粗字段</strong>
                    <strong>mountPath: /var/run/dockershim.sock</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: docker</strong>
                    <strong>mountPath: /var/run/docker</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: cri-dockerd</strong>
                    <strong>mountPath: /var/run/cri-dockerd.sock</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: sock                   # 新增以下加粗字段</strong>
                    <strong>mountPath: /var/run        # 以实际的dockershim.sock文件目录为准</strong>
                  - name: containerd
                    mountPath: /run/containerd
        ...
              volumes:
                - name: log-npu-exporter
        ...
                - name: sys
                  hostPath:
                    path: /sys
                <strong>- name: docker-shim                    # 删除以下加粗字段</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/dockershim.sock</strong>
                <strong>- name: docker</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/docker</strong>
                <strong>- name: cri-dockerd</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/cri-dockerd.sock</strong>
                <strong>- name: sock                 # 新增以下加粗字段</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run                    # 以实际的dockershim.sock文件目录为准</strong>
                - name: containerd
                  hostPath:
                    path: /run/containerd
         ...</pre>

7. 如不修改组件的其他启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中NPU Exporter的启动参数。启动参数如[表2](#table872410431914)所示，也可执行<b>./npu-exporter -h</b>查看参数说明。
8. 在管理节点的YAML所在路径，执行以下命令，启动NPU Exporter。

    - K8s集群中使用Atlas 200I SoC A1 核心板节点，执行以下命令。

        ```shell
        kubectl apply -f npu-exporter-310P-1usoc-v{version}.yaml
        ```

    - K8s集群中使用除Atlas 200I SoC A1 核心板外的其他类型节点，执行以下命令。

        ```shell
        kubectl apply -f npu-exporter-v{version}.yaml
        ```

    启动示例如下：

    ```ColdFusion
    namespace/npu-exporter created
    networkpolicy.networking.K8s.io/exporter-network-policy created
    daemonset.apps/npu-exporter created
    ```

    >[!NOTE]
    >启动NPU Exporter时，若出现报错“Error from server (NotFound): error when creating "npu-exporter-<i>x.x.x</i>.yaml":namespaces "npu-exporter" not found”，说明NPU Exporter的命名空间未创建成功，需执行以下命令手动创建。
    >
    >```shell
    >kubectl create ns npu-exporter
    >```

9. 在任意节点执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n npu-exporter
    ```

    回显示例如下，出现**Running**表示组件启动成功。若状态为**CrashLoopBackOff**，可能是因为目录权限不正确导致，可以参见[NPU Exporter检查动态路径失败，日志出现check uid or mode failed](https://gitcode.com/Ascend/mind-cluster/issues/350)章节进行处理。

    ```ColdFusion
    NAME                            READY   STATUS    RESTARTS   AGE
    ...
    npu-exporter-hqpxl        1/1    Running   0        11s
    ```

    >[!NOTE]
    >
    >- NPU Exporter的使用对进程环境有要求，以镜像方式运行时，请确保“/sys”目录和容器运行时通信socket文件挂载至NPU Exporter容器中。若通过调用NPU Exporter的Metrics接口，没有获取到NPU容器的相关信息，该问题可能是因为socket文件路径不正确导致，可以参见[日志出现connecting to container runtime failed](https://gitcode.com/Ascend/mind-cluster/issues/346)章节进行处理。
    >- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](https://gitcode.com/Ascend/mind-cluster/issues/342)章节进行处理。
    >- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](https://gitcode.com/Ascend/mind-cluster/issues/343)章节进行处理。
    >- 启动组件失败，可参考[启动集群调度组件失败，日志打印"get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390)章节信息。
    >- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](https://gitcode.com/Ascend/mind-cluster/issues/345)章节信息。

## 二进制方式运行<a name="section103551921135917"></a>

NPU Exporter组件以镜像方式运行时需使用特权容器、root用户和挂载了docker-shim或Containerd的socket文件，如果容器被人恶意利用，有容器逃逸风险。当安全性要求较高时，可直接在物理机上通过二进制方式运行。

>[!NOTE]
>
>- 以二进制方式部署NPU Exporter时，可以使用非root用户（例如hwMindX）进行部署。请将日志目录权限修改为hwMindX，命令示例如下：**chown <i>hwMindX:hwMindX</i> /var/log/mindx-dl/npu-exporter**。
>- 下文步骤中的用户均为hwMindX。

1. 使用root用户登录服务器。
2. 将NPU Exporter软件包上传至服务器的任意目录（如“/home/ascend-npu-exporter”）并进行解压操作。
3. 将NPU Exporter软件包解压目录下的metricConfiguration.json和pluginConfiguration.json文件，拷贝到“/usr/local”目录下。
4. （可选）修改metricConfiguration.json或pluginConfiguration.json文件，配置默认指标组或自定义指标组的采集开关和采集周期。
    1. 进入“/usr/local”目录。
    2. <a name="li1445835411478"></a>打开metricConfiguration.json文件。

        ```shell
        vi metricConfiguration.json
        ```

    3. 按“i”进入编辑模式，根据实际需要配置默认指标组的采集开关和采集周期。

        配置文件示例如下：

        ```json
            [
                {"metricsGroup": "version", "state": "ON", "intervalSeconds": -1},
                {"metricsGroup": "utilization", "state": "ON", "intervalSeconds": 1},
                {"metricsGroup": "npu", "state": "ON", "intervalSeconds": 5},
                {"metricsGroup": "ddr", "state": "ON", "intervalSeconds": 10},
                {"metricsGroup": "sio", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hbm", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hccs", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "pcie", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "vnpu", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "nodeBase", "state": "ON", "intervalSeconds": 86400},

                {"metricsGroup": "roce", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "optical", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "network", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "ub", "state": "ON", "intervalSeconds": 60}
            ]
        ```

        <a name="zh-cn_topic_0000002511426331_table192202574406"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002511426331_row152204575408"><th class="cellrowborder" valign="top" width="30.12%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000002511426331_p1220125712404"><a name="zh-cn_topic_0000002511426331_p1220125712404"></a><a name="zh-cn_topic_0000002511426331_p1220125712404"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="69.88%" id="mcps1.1.3.1.2"><p id="zh-cn_topic_0000002511426331_p622019575401"><a name="zh-cn_topic_0000002511426331_p622019575401"></a><a name="zh-cn_topic_0000002511426331_p622019575401"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002511426331_row182201357164014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="p152201573404"><a name="p152201573404"></a><a name="p152201573404"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="p222035704018"><a name="p222035704018"></a><a name="p222035704018"></a>默认指标组名称。</p>
        <a name="ul222055714012"></a><a name="ul222055714012"></a>
           <ul id="zh-cn_topic_0000002511426331_ul222055714012">
               通过dcmi采集：
               <li>version：版本数据信息</li>
               <li>utilization：使用率数据信息</li>
               <li>npu：NPU数据信息</li>
               <li>ddr：DDR数据信息</li>
               <li>sio：SIO数据信息</li>
               <li>hbm：片上内存数据信息</li>
               <li>hccs：HCCS数据信息</li>
               <li>pcie：PCIe数据信息</li>
               <li>vnpu：vNPU数据信息</li>
               <li>nodeBase：节点基本信息</li>
           </ul>
           <br/>
           <ul>
               通过hccn_tool采集：
               <li>roce：RoCE数据信息</li>
               <li>optical：光模块数据信息</li>
               <li>network：Network数据信息</li>
               <li>ub：NPU UB数据信息</li>
           </ul>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_row5220257114014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p182201657134015"><a name="zh-cn_topic_0000002511426331_p182201657134015"></a><a name="zh-cn_topic_0000002511426331_p182201657134015"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p722015718403"><a name="zh-cn_topic_0000002511426331_p722015718403"></a><a name="zh-cn_topic_0000002511426331_p722015718403"></a>指标组采集和上报的开关。默认值为ON。</p>
        <a name="zh-cn_topic_0000002511426331_ul14220557134016"></a><a name="zh-cn_topic_0000002511426331_ul14220557134016"></a><ul id="zh-cn_topic_0000002511426331_ul14220557134016"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_rowinterval"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_pinterval"><a name="zh-cn_topic_0000002511426331_pinterval"></a><a name="zh-cn_topic_0000002511426331_pinterval"></a>intervalSeconds</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_pintervaldesc"><a name="zh-cn_topic_0000002511426331_pintervaldesc"></a><a name="zh-cn_topic_0000002511426331_pintervaldesc"></a>指标组采集周期，单位为秒。</p>
        <a name="zh-cn_topic_0000002511426331_ulinterval"></a><a name="zh-cn_topic_0000002511426331_ulinterval"></a><ul id="zh-cn_topic_0000002511426331_ulinterval"><li>必须配置为整数值</li><li>取值范围为-1、1~86400秒</li><li>若缺失该配置项，使用默认值60秒</li><li>若配置为-1，表示该指标组只采集一次，不再重复采集</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    4. <a name="li18459954104718"></a>按“Esc”键，输入:wq!保存并退出。
    5. 参考[4.b](#li1445835411478)到[4.d](#li18459954104718)，修改pluginConfiguration.json文件，根据实际需要配置自定义指标组的采集开关和采集周期。

        <a name="table16459165464719"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002511426331_row157015443510"><th class="cellrowborder" valign="top" width="23.14%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000002511426331_p15701444553"><a name="zh-cn_topic_0000002511426331_p15701444553"></a><a name="zh-cn_topic_0000002511426331_p15701444553"></a>参数</p>
        </th>
        <th class="cellrowborder" valign="top" width="76.86%" id="mcps1.1.3.1.2"><p id="zh-cn_topic_0000002511426331_p117011441156"><a name="zh-cn_topic_0000002511426331_p117011441156"></a><a name="zh-cn_topic_0000002511426331_p117011441156"></a>说明</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002511426331_row47010440518"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p170118446517"><a name="zh-cn_topic_0000002511426331_p170118446517"></a><a name="zh-cn_topic_0000002511426331_p170118446517"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p9568105120719"><a name="zh-cn_topic_0000002511426331_p9568105120719"></a><a name="zh-cn_topic_0000002511426331_p9568105120719"></a>向<span id="zh-cn_topic_0000002511426331_ph18671058476"><a name="zh-cn_topic_0000002511426331_ph18671058476"></a><a name="zh-cn_topic_0000002511426331_ph18671058476"></a>NPU Exporter</span>注册的自定义指标组名称。自定义指标的方法详细请参见<a href="../../custom_metrics_development.md">自定义指标开发</a>。</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_row157010441654"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p157021644755"><a name="zh-cn_topic_0000002511426331_p157021644755"></a><a name="zh-cn_topic_0000002511426331_p157021644755"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p2148170172513"><a name="zh-cn_topic_0000002511426331_p2148170172513"></a><a name="zh-cn_topic_0000002511426331_p2148170172513"></a>指标组采集和上报的开关。默认值为OFF。</p>
        <a name="zh-cn_topic_0000002511426331_ul1870217441514"></a><a name="zh-cn_topic_0000002511426331_ul1870217441514"></a><ul id="zh-cn_topic_0000002511426331_ul1870217441514"><li>ON：表示开启。开启对应指标组的开关后，会采集和上报该指标组的指标。</li><li>OFF：表示关闭。关闭对应指标组的开关后，不会采集和上报该指标组的指标。</li></ul>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_rowplugininterval"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_pplugininterval"><a name="zh-cn_topic_0000002511426331_pplugininterval"></a><a name="zh-cn_topic_0000002511426331_pplugininterval"></a>intervalSeconds</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_ppluginintervaldesc"><a name="zh-cn_topic_0000002511426331_ppluginintervaldesc"></a><a name="zh-cn_topic_0000002511426331_ppluginintervaldesc"></a>自定义指标组采集周期，单位为秒。</p>
        <a name="zh-cn_topic_0000002511426331_ulplugininterval"></a><a name="zh-cn_topic_0000002511426331_ulplugininterval"></a><ul id="zh-cn_topic_0000002511426331_ulplugininterval"><li>必须配置为整数值</li><li>取值范围为1~86400秒</li><li>若不配置该参数，使用默认值5秒</li><li>若配置为-1，表示该指标组只采集一次，不再重复采集</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    6. 若通过插件方式开发了自定义指标，需重新构建编译二进制文件。

5. 创建并编辑npu-exporter.service文件。
    1. 执行以下命令，创建npu-exporter.service文件。

        ```shell
        vi /home/ascend-npu-exporter/npu-exporter.service
        ```

    2. 参考如下内容，写入npu-exporter.service文件中。

        <pre>
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
        WantedBy=multi-user.target</pre>

        NPU Exporter默认情况只侦听127.0.0.1，可通过修改的启动参数“-ip”和“npu-exporter.service”文件的“ExecStart”字段修改需要侦听的IP地址。

    3. 按“Esc”键，输入:wq!保存并退出。

6. 创建并编辑npu-exporter.timer文件。通过配置timer延时启动，可保证NPU Exporter启动时NPU卡已就位。
    1. 执行以下命令，创建npu-exporter.timer文件。

        ```shell
         vi /home/ascend-npu-exporter/npu-exporter.timer
        ```

    2. 参考以下示例，并将其写入npu-exporter.timer文件中。

        <pre>
        [Unit]
        Description=Timer for NPU Exporter Service

        [Timer]
        OnBootSec=60s            # 设置NPU Exporter延时启动时间，请根据实际情况调整
        Unit=npu-exporter.service

        [Install]
        WantedBy=timers.target</pre>

    3. 按“Esc”键，输入:wq!保存并退出。

7. 若部署节点为Atlas 200I SoC A1 核心板，请依次执行以下命令，在节点上将hwMindX用户加入到HwBaseUser、HwDmUser用户组中。非Atlas 200I SoC A1 核心板用户，可跳过本步骤。

    ```shell
    usermod -a -G HwBaseUser hwMindX
    usermod -a -G HwDmUser hwMindX
    ```

8. 依次执行以下命令，启用NPU Exporter服务。

    ```shell
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

    > [!NOTE]
    >如果需要获取容器相关数据信息，NPU Exporter需要临时提权以便于和CRI、OCI的socket建立连接，需要执行以下命令。
    >
    >```shell
    >chattr -i /usr/local/bin/npu-exporter
    >setcap cap_setuid+ep /usr/local/bin/npu-exporter
    >chattr +i /usr/local/bin/npu-exporter
    >systemctl restart npu-exporter
    >```

## 参数说明<a name="section2042611570392"></a>

**表 2** NPU Exporter启动参数

<a name="table872410431914"></a>

|参数|类型|默认值|说明|
|--|--|--|--|
|-port|int|8082|侦听端口，取值范围为1025~40000。|
|-updateTime|int|无|**即将日落，不推荐使用**。全局配置指标更新周期，取值范围为1~60秒。推荐按分组配置指标更新周期，详见[配置文件](#section103551921135917)。<p>兼容性说明：如果配置了updateTime参数，该参数仍然有效，且优先级高于metricConfiguration.json/pluginConfiguration.json配置文件中的intervalSeconds。</p>|
|-ip|string|无|参数无默认值，必须配置。<p>侦听IP地址，要求为合法的IPv4或IPv6格式，在多网卡主机上不建议配置成0.0.0.0。</p>|
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
|--enable-healthz|bool|false|是否启用健康检查服务。K8s部署时由组件YAML配置启用（true）。<ul><li>true：启用。</li><li>false：禁用。</li></ul>|
|--healthz-address|string|11251|健康检查服务侦听端口号，取值范围 1025~65535。K8s部署时由组件YAML配置为11256。若指定端口被占用，组件启动失败。|
|--tls-cert-file|string|""|HTTPS 证书文件路径。为空则使用 HTTP 协议。与 --tls-private-key-file 必须同时配置或同时为空。配置方法及安全注意事项请参见<a href="../../../07_references/security_hardening.md#健康探针安全加固">健康探针安全加固</a>。|
|--tls-private-key-file|string|""|HTTPS 私钥文件路径。为空则使用 HTTP 协议。与 --tls-cert-file 必须同时配置或同时为空。|
|-h或者-help|无|无|显示帮助信息。|
|-platform|string|Prometheus|指定对接平台。<ul><li>Prometheus：对接Prometheus</li><li>Telegraf：对接Telegraf</li></ul>|
|-poll_interval|duration(int)|1|Telegraf数据上报的间隔时间，单位：秒。此参数在对接Telegraf平台时才起作用，即需要指定-platform=Telegraf时才生效，否则该参数不生效。|
|-profilingTime|int|200|配置采集PCIe带宽时间，单位：毫秒，取值范围为1~2000。|
|-hccsBWProfilingTime|int|200|HCCS链路带宽采样时长，取值范围1~1000，单位：毫秒。|
|-deviceResetTimeout|int|60|组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600。<ul><li>Atlas A2 训练系列产品、Atlas 800I A2 推理服务器、A200I A2 Box 异构组件：建议配置为150秒。</li><li>Atlas A3 训练系列产品、A200T A3 Box8 超节点服务器、Atlas 800I A3 超节点服务器：建议配置为360秒。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD：建议配置为600秒。</li></ul>|
|-textMetricsFilePath|string|无|指定自定义指标文件的路径，其约束说明详细请参见[约束说明](../../../06_api/npu_exporter/03_custom_metrics_file.md#约束说明)。|

## 动态加载配置说明<a name="动态配置加载说明"></a>

NPU Exporter支持动态加载配置文件，无需重启组件即可使配置变更生效。

### 二进制部署场景

配置文件路径为`/user/mind-cluster/npu-exporter-config/metricConfiguration.json`和`/user/mind-cluster/npu-exporter-config/pluginConfiguration.json`。

直接修改这两个文件即可，NPU Exporter会自动检测文件变更并重新加载配置。

### K8s HostPath挂载场景

在部署YAML中配置HostPath挂载，将宿主机上的配置文件挂载到容器的`/user/mind-cluster/npu-exporter-config`路径：

**步骤 1：在节点上准备配置文件**

进入NPU Exporter软件包解压目录。

```bash
# 在每个节点上创建配置目录，目录可自定义，和yaml中挂载路径保持一致即可
mkdir -p /user/mind-cluster/npu-exporter-config
cp metricConfiguration.json /user/mind-cluster/npu-exporter-config
cp pluginConfiguration.json /user/mind-cluster/npu-exporter-config
```

**步骤 2：挂载到 Pod**

```yaml
volumeMounts:
  - name: npu-config
    mountPath: /user/mind-cluster/npu-exporter-config
    readOnly: true

volumes:
  - name: npu-config
    hostPath:
      path: /user/mind-cluster/npu-exporter-config
      type: DirectoryOrCreate
```

修改宿主机上的配置文件即可，NPU Exporter会自动检测文件变更并重新加载配置。

HostPath挂载的优点：

- 配置修改后立即生效
- 每个节点可以独立配置，也可结合共享目录实现全局统一配置

HostPath挂载的缺点：

- 配置变更不易追踪

### K8s ConfigMap挂载场景

**步骤 1：创建 ConfigMap**

进入NPU Exporter软件包解压目录。

创建 ConfigMap（从文件创建）：

```bash
kubectl create ns npu-exporter
kubectl create cm -n npu-exporter npu-exporter-metric-config \
  --from-file=metricConfiguration.json=./metricConfiguration.json \
  --from-file=pluginConfiguration.json=./pluginConfiguration.json
```

**步骤 2：挂载到 Pod**

```yaml
volumeMounts:
  - name: npu-config
    mountPath: /user/mind-cluster/npu-exporter-config
    readOnly: true

volumes:
  - name: npu-config
    configMap:
      name: npu-exporter-metric-config
```

> **重要**：直接通过NPU Exporter软件包中的yaml进行组件卸载时，npu-exporter-metric-config ConfigMap会被删除，如需保留配置，请提前备份：
>
> ```bash
> kubectl get cm -n npu-exporter npu-exporter-metric-config -o yaml > npu-exporter-metric-config.yaml
> ```

> **重要**：必须直接挂载 ConfigMap 到目录，**不能使用`subPath`**：
> - 使用 `subPath`：会导致ConfigMap更新后无法自动同步到容器内，需要重启才能生效
> - 修改 ConfigMap 后，容器内文件不能实时更新，需要等待一定时间（K8s 机制，最长约 10 分钟）才能感知到文件变化

更新ConfigMap后，K8s会自动更新容器中的配置文件，NPU Exporter会自动检测并重新加载配置。

ConfigMap挂载的优点：

- 统一管理所有节点的配置，支持一键更新所有节点配置
- 配置变更可追踪和版本控制

ConfigMap挂载的缺点：

- 所有节点使用相同配置，无法独立配置单个节点
- 配置生效有一定延迟（K8s更新ConfigMap到容器的时间）

### 配置变更验证

配置变更后，可以通过查看NPU Exporter的日志来确认配置是否成功加载：

```shell
# 需先登录到对应服务器
# prometheus场景
tail -100f /var/log/mindx-dl/npu-exporter/npu-exporter.log

# telegraf场景
tail -100f /var/log/mindx-dl/npu-exporter/npu-plugin.log
```

成功加载配置会打印类似日志：

```text
detected config change: ...
reloading configuration...
```
