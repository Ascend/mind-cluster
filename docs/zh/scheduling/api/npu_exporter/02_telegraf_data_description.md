# Telegraf数据信息说明<a name="ZH-CN_TOPIC_0000002511426775"></a>

运行Telegraf后，会显示监测的昇腾AI处理器的数据信息，回显示例如下，仅供参考，以实际回显为准。数据信息的详细说明参见下文或[数据信息说明.xlsx](../../../resource/数据信息说明.xlsx)。

```ColdFusion
...
Ascend910-0,host=xxx  npu_chip_link_speed=104857600000i,npu_chip_roce_rx_cnp_pkt_num=0i,npu_chip_roce_unexpected_ack_num=0i,npu_chip_optical_vcc=3245.1,npu_chip_optical_rx_power_1=0.8585,npu_chip_info_hbm_used_memory=0i,npu_chip_mac_rx_pause_num=0i,npu_chip_roce_tx_all_pkt_num=0i,npu_chip_roce_tx_cnp_pkt_num=0i,npu_chip_info_temperature=46,npu_chip_mac_rx_bad_pkt_num=0i,npu_chip_roce_tx_err_pkt_num=0i,npu_chip_optical_rx_power_3=0.8466,npu_chip_optical_rx_power_0=0.7933,npu_chip_info_network_status=0i,npu_chip_mac_rx_pfc_pkt_num=0i,npu_chip_mac_tx_bad_pkt_num=0i,npu_chip_roce_rx_all_pkt_num=0i,npu_chip_mac_rx_bad_oct_num=0i,npu_chip_optical_tx_power_1=0.9162,npu_chip_info_utilization=0,npu_chip_info_power=73.9000015258789,npu_chip_info_link_status=1i,npu_chip_info_bandwidth_rx=0,npu_chip_mac_tx_pfc_pkt_num=0i,npu_chip_roce_rx_err_pkt_num=0i,npu_chip_roce_verification_err_num=0i,npu_chip_optical_state=1i,npu_chip_info_bandwidth_tx=0,npu_chip_mac_tx_bad_oct_num=0i,npu_chip_roce_out_of_order_num=0i,npu_chip_roce_qp_status_err_num=0i,npu_chip_optical_rx_power_2=0.855,npu_chip_optical_tx_power_0=0.9095,npu_chip_info_hbm_utilization=0,npu_chip_link_up_num=2i,npu_chip_info_health_status=1i,npu_chip_mac_tx_pause_num=0i,npu_chip_roce_new_pkt_rty_num=0i,npu_chip_optical_temp=53,npu_chip_optical_tx_power_2=1.0342,npu_chip_optical_tx_power_3=0.9715 1694772754612200641,npu_chip_info_process_info_num=0i
```

本接口支持查询默认指标组和自定义指标组。自定义指标组的方法详细请参见[自定义指标开发](../../appendix.md#自定义指标开发)；默认指标组包含如下几个部分。指标组的采集和上报由配置文件中的开关控制，若开关配置为开启，则对应的指标组会进行采集和上报；若开关配置为关闭，则对应的指标组不会进行采集和上报。

- [版本数据信息](#section170316521436141)
- [NPU数据信息](#section1442282202316)
- [vNPU数据信息](#section814111613432)
- [Network数据信息](#section1358881214551)
- [片上内存数据信息](#section177232045203114)
- [HCCS数据信息](#section039816240252)
- [PCIe数据信息](#section1240520241824136)
- [RoCE数据信息](#section184516450323)
- [SIO数据信息](#section7109037161515)
- [光模块数据信息](#section1517163183510)
- [DDR数据信息](#section114607361931169)
- [UB数据信息](#section998877563214)

>[!NOTE]
>
>- NPU Exporter是通过调用底层的HDK接口，获取相应的信息。数据信息调用的HDK接口请参考[调用的HDK接口](#section345820153363)。
>- 若查询某个数据信息时，NPU Exporter组件不支持该设备形态或调用HDK接口失败，则不会上报该数据信息。

## 版本数据信息<a name="section170316521436141"></a>

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
<p id="p1115563443"><a name="p1115563443"></a><a name="p1115563443"></a><span id="ph9892125584318"><a name="ph9892125584318"></a><a name="ph9892125584318"></a>A200I A2 Box 异构组件</span></p><p><span>Atlas 350 标卡</span></p><p><span>Atlas 950 SuperPoD</span></p>
</td>
</tr>
</tbody>
</table>

## NPU数据信息<a name="section1442282202316"></a>

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
<td class="cellrowborder" rowspan="13" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><a name="ul1142611144613"></a><a name="ul1142611144613"></a><ul id="ul1142611144613"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li>推理服务器（插Atlas 300I 推理卡）</li><li>Atlas 推理系列产品</li><li><span id="ph279972618380"><a name="ph279972618380"></a><a name="ph279972618380"></a>Atlas 800I A2 推理服务器</span></li><li><span id="ph1823654413571"><a name="ph1823654413571"></a><a name="ph1823654413571"></a>A200I A2 Box 异构组件</span></li><li><span>Atlas 350 标卡</span></li></ul>
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
<div class="note" id="note71551511134616"><a name="note71551511134616"></a><a name="note71551511134616"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul01551011174616"></a><a name="ul01551011174616"></a><ul id="ul01551011174616"><li>Prometheus场景：若该<span id="ph4155141154620"><a name="ph4155141154620"></a><a name="ph4155141154620"></a>昇腾AI处理器</span>上同时存在多个错误码，由于Prometheus格式限制，当前只支持上报前十个出现的错误码。X的取值范围：1~9</li><li>Telegraf场景：最多支持上报128个错误码。</li><li>错误码的详细说明，请参见<a href="../../appendix.md#芯片故障码参考文档">芯片故障码参考文档</a>章节。</li></ul>
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
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p447112192567"><a name="p447112192567"></a><a name="p447112192567"></a><span id="ph7471111975618"><a name="ph7471111975618"></a><a name="ph7471111975618"></a><span id="ph3471141910562"><a name="ph3471141910562"></a><a name="ph3471141910562"></a>昇腾AI处理器</span></span>的AICore利用率。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p247131915566"><a name="p247131915566"></a><a name="p247131915566"></a>单位：%</p>
</td>
</tr>
<tr id="row15230837165519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4471171935616"><a name="p4471171935616"></a><a name="p4471171935616"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p15471161918560"><a name="p15471161918560"></a><a name="p15471161918560"></a>npu_chip_info_aicore_current_freq</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p19471819155615"><a name="p19471819155615"></a><a name="p19471819155615"></a>昇腾AI处理器的AICore当前频率</p>
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
<a name="ul148152911516"></a><a name="ul148152911516"></a><ul id="ul148152911516"><li><span id="ph1848152919517"><a name="ph1848152919517"></a><a name="ph1848152919517"></a>A200I A2 Box 异构组件</span></li><li><span>Atlas 350 标卡</span></li></ul>
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
<ul><li><span>A200I A2 Box 异构组件</span></li><li><span>Atlas 350 标卡</span></li></ul>
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

## vNPU数据信息<a name="section814111613432"></a>

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
<td class="cellrowborder" valign="top" width="25.182518251825186%" headers="mcps1.2.6.1.3 "><p id="p7803225132114"><a name="p7803225132114"></a><a name="p7803225132114"></a>vNPU的AICore利用率</p>
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

## Network数据信息<a name="section1358881214551"></a>

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
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetworkNpu</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_status_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口Link状态。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>取值为0或1</p><ul><li>1：UP</li><li>0：DOWN</li></ul>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="19.91%" headers="mcps1.1.7.1.6 "><p><span>Atlas 350 标卡（4Pmesh互联）</span></p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetworkNpu</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口实时接收速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。使用的-time参数为100。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetworkNpu</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>昇腾AI处理器端口实时发送速率。</p><p>其中，X为Udie ID，Y为Port ID，在Atlas 350 标卡中，Udie ID为0，Port ID取值范围为[4, 6]。使用的-time参数为100。</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>单位：MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>NetworkNpu</p>
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

## 片上内存数据信息<a name="section177232045203114"></a>

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
<td class="cellrowborder" rowspan="12" valign="top" width="26.61%" headers="mcps1.2.6.1.5 "><a name="ul1737721403120"></a><a name="ul1737721403120"></a><ul id="ul1737721403120"><li>Atlas 训练系列产品</li><li>Atlas A2 训练系列产品</li><li>Atlas A3 训练系列产品</li><li><span id="ph043025116483"><a name="ph043025116483"></a><a name="ph043025116483"></a>A200I A2 Box 异构组件</span></li><li><span id="ph1201913534"><a name="ph1201913534"></a><a name="ph1201913534"></a>Atlas 800I A2 推理服务器</span></li></ul><ul><li><span>Atlas 350 标卡</span></li></ul>
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

## HCCS数据信息<a name="section039816240252"></a>

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

## PCIe数据信息<a name="section1240520241824136"></a>

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

## RoCE数据信息<a name="section184516450323"></a>

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

## SIO数据信息<a name="section7109037161515"></a>

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
<td class="cellrowborder" rowspan="2" valign="top" width="23.309999999999995%" headers="mcps1.2.6.1.5 "><ul><li>Atlas A3 训练系列产品</li><li>Atlas 350 标卡</li></ul>
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

## 光模块数据信息<a name="section1517163183510"></a>

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
<tr><td class="cellrowborder" valign="top" width="8.150815081508151%" headers="mcps1.2.7.1.1 "><p>光模块</p>
</td>
<td class="cellrowborder" valign="top" width="22.852285228522852%" headers="mcps1.2.7.1.2 "><p>npu_chip_info_optical_index_num_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="19.401940194019403%" headers="mcps1.2.7.1.3 "><p>芯片Udie Port连接的光模块Lane数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" width="11.96119611961196%" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="22.872287228722872%" headers="mcps1.2.7.1.6 "><p>Atlas 850 系列硬件产品</p>
</td>
</tr>
<tr id="row184616483311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_optical_tx_power_Z_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>光模块发送功率。Z为Lane的index，取值为[0:3]，X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>单位：mW</p>
</td>
</tr>
<tr id="row1846416482311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>光模块</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_optical_rx_power_Z_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>光模块接收功率。Z为Lane的index，取值为[0:3]，X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>单位：mW</p>
</td>
</tr>
</tbody>
</table>

## DDR数据信息<a name="section114607361931169"></a>

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

## UB数据信息<a name="section998877563214"></a>

**表 12**  UB数据信息

<a name="table998877563214"></a>
<table><thead align="left"><tr><th class="cellrowborder" valign="top" width="8.150815081508151%"><p>类别</p>
</th>
<th class="cellrowborder" valign="top" width="22.852285228522852%"><p>数据信息名称</p>
</th>
<th class="cellrowborder" valign="top" width="19.401940194019403%"><p>数据信息说明</p>
</th>
<th class="cellrowborder" valign="top" width="11.96119611961196%"><p>单位</p>
</th>
<th class="cellrowborder" valign="top" width="22.872287228722872%"><p>支持的产品形态</p>
</th>
</tr>
</thead>
<tbody><tr><td class="cellrowborder" valign="top" width="8.150815081508151%" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" width="22.852285228522852%" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv4_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="19.401940194019403%" headers="mcps1.2.7.1.3 "><p>RX侧接收到的IPv4 UB报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" width="11.96119611961196%" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
<td class="cellrowborder" rowspan="48" valign="top" width="22.872287228722872%" headers="mcps1.2.7.1.6 "><ul><li>Atlas 350 标卡</li><li>Atlas 850 系列硬件产品</li><li>Atlas 950 SuperPoD</li></ul>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv6_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的IPv6 UB报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row1846416482311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv4_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1515522153519"><a name="p1515522153519"></a><a name="p1515522153519"></a>RX侧接收到的IPv4 UNIC报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row20466104816315"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv6_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的IPv6 UNIC报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_compact_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的CFG6报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ctph_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的CFG7 CLAN报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ntph_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的CFG7 非CLAN报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_mem_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的UB mem报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unknown_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的未知报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_drop_ind_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的带drop_ind的报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_err_ind_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的ERR报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_host_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过路由后的落地报文数量（不包含枚举配置和管理报文）。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_imp_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过路由后落地的枚举配置和管理报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_mar_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过路由后落地的UB memory报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_link_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过路由后转发到同Port的TX侧的报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_noc_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过路由后的P2P报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_route_err_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过路由后出现路由查表错误的报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_out_err_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过校验后的错误报文总数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_length_err_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX侧接收到的报文经过校验后的长度错误报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_rx_busi_flit_num_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX报文字节数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_rx_send_ack_flit_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX向对端返回响应的报文字节数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv4_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的IPv4 UB报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv6_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的IPv6 UB报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv4_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的IPv4 UNIC报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv6_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的IPv6 UNIC报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_compact_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的CFG6报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ctph_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的CFG7 CLAN报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ntph_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的CFG7 非CLAN报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_mem_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的UB mem报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unknown_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的未知报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_drop_ind_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的带drop_ind的报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_err_ind_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的ERR报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_lpbk_ind_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送报文在NL环回的报文个数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_out_err_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的报文经过校验后的错误报文总数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_length_err_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX侧发送的报文经过校验后的长度错误报文数量。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_tx_busi_flit_num_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX报文字节数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_tx_recv_ack_flit_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX收到对端响应的报文字节数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_retry_req_sum_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>发起重传次数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_retry_ack_sum_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>响应重传次数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_crc_error_sum_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>CRC校验错误次数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxpausepkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX pause帧总报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txpausepkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX pause帧总报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxpfcpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX PFC帧总报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txpfcpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX PFC帧总报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxbadpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX坏包总报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txbadpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX坏包总报文数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxbadoctets_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>RX坏包总报文字节数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txbadoctets_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>TX坏包总报文字节数。X为Udie ID，Y为Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
</tbody>
</table>

## 调用的HDK接口<a name="section345820153363"></a>

NPU Exporter是通过调用底层的HDK接口，获取相应的信息。数据信息调用的HDK接口请参考[NPU Exporter调用的HDK接口.xlsx](../../../resource/NPU-Exporter调用的HDK接口.xlsx)。查找数据信息对应的HDK接口，可参考如下步骤。

1. 登录[昇腾计算文档](https://support.huawei.com/enterprise/zh/category/ascend-computing-pid-1557196528909?submodel=doc)中心，选择单击对应产品名称，进入文档界面。例如Atlas 800I A2 推理服务器产品的用户，单击“Atlas 800I A2”。
2. 在左侧导航栏找到“二次开发”，根据接口的类型选择对应文档。
    - DCMI接口选择“API参考”，单击进入《[DCMI API参考](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-hdk-pid-252764743?category=developer-documents&subcategory=api-reference)》。
    - HCCN Tool接口选择“接口参考”，单击进入《[Atlas A2 中心推理和训练硬件 25.5.0 HCCN Tool 接口参考](https://support.huawei.com/enterprise/zh/doc/EDOC1100540101/426cffd9)》。

3. 在文档首页搜索栏中，直接搜索对应的接口名称或者关键词，获取接口的相关信息。
