# Telegraf Data Information Description<a name="ZH-CN_TOPIC_0000002511426775"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-26T11:48:51.406Z pushedAt=2026-06-27T00:32:25.625Z -->

After Telegraf runs, the monitored data information of the Ascend AI Processor will be displayed. The following information is for reference only, and the actual display shall prevail. For detailed description of the data information, see below or [Data Information Description.xlsx](../../../resource/Data Information说明.xlsx).

```ColdFusion
...
Ascend910-0,host=xxx  npu_chip_link_speed=104857600000i,npu_chip_roce_rx_cnp_pkt_num=0i,npu_chip_roce_unexpected_ack_num=0i,npu_chip_optical_vcc=3245.1,npu_chip_optical_rx_power_1=0.8585,npu_chip_info_hbm_used_memory=0i,npu_chip_mac_rx_pause_num=0i,npu_chip_roce_tx_all_pkt_num=0i,npu_chip_roce_tx_cnp_pkt_num=0i,npu_chip_info_temperature=46,npu_chip_mac_rx_bad_pkt_num=0i,npu_chip_roce_tx_err_pkt_num=0i,npu_chip_optical_rx_power_3=0.8466,npu_chip_optical_rx_power_0=0.7933,npu_chip_info_network_status=0i,npu_chip_mac_rx_pfc_pkt_num=0i,npu_chip_mac_tx_bad_pkt_num=0i,npu_chip_roce_rx_all_pkt_num=0i,npu_chip_mac_rx_bad_oct_num=0i,npu_chip_optical_tx_power_1=0.9162,npu_chip_info_utilization=0,npu_chip_info_power=73.9000015258789,npu_chip_info_link_status=1i,npu_chip_info_bandwidth_rx=0,npu_chip_mac_tx_pfc_pkt_num=0i,npu_chip_roce_rx_err_pkt_num=0i,npu_chip_roce_verification_err_num=0i,npu_chip_optical_state=1i,npu_chip_info_bandwidth_tx=0,npu_chip_mac_tx_bad_oct_num=0i,npu_chip_roce_out_of_order_num=0i,npu_chip_roce_qp_status_err_num=0i,npu_chip_optical_rx_power_2=0.855,npu_chip_optical_tx_power_0=0.9095,npu_chip_info_hbm_utilization=0,npu_chip_link_up_num=2i,npu_chip_info_health_status=1i,npu_chip_mac_tx_pause_num=0i,npu_chip_roce_new_pkt_rty_num=0i,npu_chip_optical_temp=53,npu_chip_optical_tx_power_2=1.0342,npu_chip_optical_tx_power_3=0.9715 1694772754612200641,npu_chip_info_process_info_num=0i
```

This interface supports querying default metric groups and custom metric groups. For details on how to customize metric groups, see [Custom Metric Development](../../references/appendix.md#custom-metric-development). The default metric groups include the following sections. The collection and reporting of metric groups are controlled by switches in the configuration file. If a switch is configured as enabled, the corresponding metric group will be collected and reported; if a switch is configured as disabled, the corresponding metric group will not be collected and reported.

- [Telegraf Data Information Description](#telegraf-data-information-description)
  - [Version Data Information](#version-data-information)
  - [Basic Node Information](#basic-node-information)
  - [NPU Data Information](#npu-data-information)
  - [Utilization Data Information](#utilization-data-information)
  - [vNPU Data Information](#vnpu-data-information)
  - [Network Data Information](#network-data-information)
  - [On-Chip Memory Data Information](#on-chip-memory-data-information)
  - [HCCS Data Information](#hccs-data-information)
  - [PCIe Data Information](#pcie-data-information)
  - [RoCE Data Information](#roce-data-information)
  - [SIO Data Information](#sio-data-information)
  - [Optical Module Data Information](#optical-module-data-information)
  - [DDR Data Information](#ddr-data-information)
  - [UB Data Information](#ub-data-information)
  - [HDK Interface Call](#hdk-interface-call)

>[!NOTE]
>
>- NPU Exporter obtains the corresponding information by calling the underlying HDK interface. For the HDK interfaces called for data information, see [HDK Interface Call](#section345820153363).
>- If NPU Exporter does not support the product form or fails to call the HDK interface when querying a certain data information item, this data information will not be reported.

## Version Data Information<a name="section170316521436141"></a>

**Table 1** Version data information

<a name="table81981837143713"></a>
<table><thead align="left"><tr id="row319910378373"><th class="cellrowborder" valign="top" width="7.9399999999999995%" id="mcps1.2.6.1.1"><p id="p622884917372"><a name="p622884917372"></a><a name="p622884917372"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="25.03%" id="mcps1.2.6.1.2"><p id="p1322844963710"><a name="p1322844963710"></a><a name="p1322844963710"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="21.23%" id="mcps1.2.6.1.3"><p id="p822864983718"><a name="p822864983718"></a><a name="p822864983718"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="21.8%" id="mcps1.2.6.1.4"><p id="p15229649123712"><a name="p15229649123712"></a><a name="p15229649123712"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="24%" id="mcps1.2.6.1.5"><p id="p1023084933716"><a name="p1023084933716"></a><a name="p1023084933716"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row191991937123720"><td class="cellrowborder" valign="top" width="7.9399999999999995%" headers="mcps1.2.6.1.1 "><p id="p96501612386"><a name="p96501612386"></a><a name="p96501612386"></a>Version</p>
</td>
<td class="cellrowborder" valign="top" width="25.03%" headers="mcps1.2.6.1.2 "><p id="p36502183811"><a name="p36502183811"></a><a name="p36502183811"></a>npu_exporter_version_info</p>
</td>
<td class="cellrowborder" valign="top" width="21.23%" headers="mcps1.2.6.1.3 "><p id="p1665017112386"><a name="p1665017112386"></a><a name="p1665017112386"></a><span id="ph122556224122"><a name="ph122556224122"></a><a name="ph122556224122"></a>NPU Exporter</span> version information</p>
</td>
<td class="cellrowborder" valign="top" width="21.8%" headers="mcps1.2.6.1.4 "><p id="p11641734202218"><a name="p11641734202218"></a><a name="p11641734202218"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24%" headers="mcps1.2.6.1.5 "><ul><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Inference Server (with Atlas 300I Inference Card)</li><li>Atlas Inference Series Products</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Subrack</li><li>Atlas 350 PCIe Card</li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>
</td>
</tr>
</tbody>
</table>

## Basic Node Information<a name="section170316521436120"></a>

**Table 2**  Basic node information

|Category|Data Information Name|Data Information Description|Unit|Supported Product Forms|
|------|-------------|-------------|------|---------------|
|nodeBase|node_base_info|Basic node information, including: <ul><li>exporterVersion: Current NPU Exporter version information</li><li>driverVersion: Driver version information</li></ul>| 1: Placeholder character, having no actual meaning. |<ul><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Inference Server (with Atlas 300I Inference Card)</li><li>Atlas Inference Series Products</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Subrack</li><li>Atlas 350 PCIe Card</li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>|

## NPU Data Information<a name="section1442282202316"></a>

**Table 3** NPU data information

<a name="table18223172210289"></a>
<table><thead align="left"><tr id="row8223722122814"><th class="cellrowborder" valign="top" width="9.8%" id="mcps1.2.6.1.1"><p id="p14952185417289"><a name="p14952185417289"></a><a name="p14952185417289"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="24.279999999999998%" id="mcps1.2.6.1.2"><p id="p4953105472810"><a name="p4953105472810"></a><a name="p4953105472810"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.6.1.3"><p id="p11953954142813"><a name="p11953954142813"></a><a name="p11953954142813"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="19.68%" id="mcps1.2.6.1.4"><p id="p149531054142812"><a name="p149531054142812"></a><a name="p149531054142812"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="23.94%" id="mcps1.2.6.1.5"><p id="p14953115412284"><a name="p14953115412284"></a><a name="p14953115412284"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody>
<tr id="row_machine_card_nums"><td class="cellrowborder" valign="top" width="9.8%" headers="mcps1.2.6.1.1 "><p id="p_machine_card_nums_cat">NPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.279999999999998%" headers="mcps1.2.6.1.2 "><p id="p_machine_card_nums_name">machine_card_nums</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p_machine_card_nums_desc"><span>Number of Ascend AI Processor</span> modules</p>
</td>
<td class="cellrowborder" valign="top" width="19.68%" headers="mcps1.2.6.1.4 "><p id="p_machine_card_nums_unit">Unit: pcs</p>
</td>
<td class="cellrowborder" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><ul><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Inference Server (with Atlas 300I Inference Card)</li><li>Atlas Inference Series Products</li><li><span>Atlas 800I A2 Inference Server</span></li><li><span>A200I A2 Box Heterogeneous Subrack</span></li></ul>
</td>
</tr>
<tr id="row1999144418467"><td class="cellrowborder" valign="top" width="9.8%" headers="mcps1.2.6.1.1 "><p id="p1549615614460"><a name="p1549615614460"></a><a name="p1549615614460"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.279999999999998%" headers="mcps1.2.6.1.2 "><p id="p949618565464"><a name="p949618565464"></a><a name="p949618565464"></a>machine_npu_nums</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p12496205674616"><a name="p12496205674616"></a><a name="p12496205674616"></a>Number of Ascend AI Processors</p>
</td>
<td class="cellrowborder" valign="top" width="19.68%" headers="mcps1.2.6.1.4 "><p id="p19999944144617"><a name="p19999944144617"></a><a name="p19999944144617"></a>Unit: pcs</p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><a name="ul1142611144613"></a><a name="ul1142611144613"></a><ul id="ul1142611144613"><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Inference Server (with Atlas 300I Inference Card)</li><li>Atlas Inference Series Products</li><li><span id="ph279972618380"><a name="ph279972618380"></a><a name="ph279972618380"></a>Atlas 800I A2 Inference Server</span></li><li><span id="ph1823654413571"><a name="ph1823654413571"></a><a name="ph1823654413571"></a>A200I A2 Box Heterogeneous Subrack</span></li><li><span>Atlas 350 PCIe Card</span></li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>
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
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p843414559571"><a name="p843414559571"></a><a name="p843414559571"></a><span id="ph543465525716"><a name="ph543465525716"></a><a name="ph543465525716"></a>Ascend AI Processor</span> name and ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1831841114417"><a name="p1831841114417"></a><a name="p1831841114417"></a>-</p>
</td>
</tr>
<tr id="row1320035495310"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12169113820544"><a name="p12169113820544"></a><a name="p12169113820544"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p111691338155414"><a name="p111691338155414"></a><a name="p111691338155414"></a>npu_chip_info_health_status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p20169193819546"><a name="p20169193819546"></a><a name="p20169193819546"></a><span id="ph1169163818542"><a name="ph1169163818542"></a><a name="ph1169163818542"></a><span id="ph11169203811544"><a name="ph11169203811544"></a><a name="ph11169203811544"></a>Ascend AI Processor</span></span> health status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p316933865416"><a name="p316933865416"></a><a name="p316933865416"></a>The value is 0, 1, or -1.</p>
<a name="ul1016913885417"></a><a name="ul1016913885417"></a><ul id="ul1016913885417"><li>1: Healthy</li><li>0: Unhealthy</li><li>-1: Unknown (DCMI call failed)</li></ul>
</td>
</tr>
<tr id="row10385175935311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p316943810545"><a name="p316943810545"></a><a name="p316943810545"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p116913381548"><a name="p116913381548"></a><a name="p116913381548"></a>npu_chip_info_power</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3169638165419"><a name="p3169638165419"></a><a name="p3169638165419"></a><span id="ph10169183819548"><a name="ph10169183819548"></a><a name="ph10169183819548"></a><span id="ph1616953865415"><a name="ph1616953865415"></a><a name="ph1616953865415"></a>Ascend AI Processor</span></span> power consumption.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1516963825410"><a name="p1516963825410"></a><a name="p1516963825410"></a>Unit: Watt (W)</p>
</td>
</tr>
<tr id="row556922812544"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p13169838125413"><a name="p13169838125413"></a><a name="p13169838125413"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p71693385547"><a name="p71693385547"></a><a name="p71693385547"></a>npu_chip_info_temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11169113815412"><a name="p11169113815412"></a><a name="p11169113815412"></a><span id="ph216910388546"><a name="ph216910388546"></a><a name="ph216910388546"></a><span id="ph1716913384544"><a name="ph1716913384544"></a><a name="ph1716913384544"></a>Ascend AI Processor</span></span> temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1717023895414"><a name="p1717023895414"></a><a name="p1717023895414"></a>Unit: Celsius (°C)</p>
</td>
</tr>
<tr id="row922903714555"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p8470111915614"><a name="p8470111915614"></a><a name="p8470111915614"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p9470151955618"><a name="p9470151955618"></a><a name="p9470151955618"></a>First error code: npu_chip_info_error_code</p>
<p id="p1658317311118"><a name="p1658317311118"></a><a name="p1658317311118"></a>Other error codes: npu_chip_info_error_code_X</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p547061925610"><a name="p547061925610"></a><a name="p547061925610"></a><span id="ph1415520117467"><a name="ph1415520117467"></a><a name="ph1415520117467"></a>Ascend AI Processor</span> error code</p>
<p id="p2011915418141"><a name="p2011915418141"></a><a name="p2011915418141"></a>When there is no error code on the Ascend AI Processor, this field will not be reported.</p>
<div class="note" id="note71551511134616"><a name="note71551511134616"></a><a name="note71551511134616"></a><span class="notetitle">Note:</span><div class="notebody"><a name="ul01551011174616"></a><a name="ul01551011174616"></a><ul id="ul01551011174616"><li>Prometheus scenario: If multiple error codes exist on this <span id="ph4155141154620"><a name="ph4155141154620"></a><a name="ph4155141154620"></a>Ascend AI Processor</span> simultaneously, due to Prometheus format limitations, only the first ten error codes that appear are currently supported for reporting. The value range of X: 1~9</li><li>Telegraf scenario: Supports reporting up to 128 error codes.</li><li>For detailed descriptions of error codes, see the <a href="../../references/appendix.md#chip-fault-code-references">Chip Fault Code References</a> section.</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p2470171915560"><a name="p2470171915560"></a><a name="p2470171915560"></a>-</p>
</td>
</tr>
<tr id="row19230153714559"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p74701519145618"><a name="p74701519145618"></a><a name="p74701519145618"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14708196563"><a name="p14708196563"></a><a name="p14708196563"></a>npu_chip_info_process_info_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10470101985611"><a name="p10470101985611"></a><a name="p10470101985611"></a>Number of processes occupying the <span id="ph13471719205614"><a name="ph13471719205614"></a><a name="ph13471719205614"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17471219115612"><a name="p17471219115612"></a><a name="p17471219115612"></a>-</p>
</td>
</tr>
<tr id="row15230837165519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4471171935616"><a name="p4471171935616"></a><a name="p4471171935616"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p15471161918560"><a name="p15471161918560"></a><a name="p15471161918560"></a>npu_chip_info_aicore_current_freq</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p19471819155615"><a name="p19471819155615"></a><a name="p19471819155615"></a>Current frequency of the AICore of the Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p104711319115610"><a name="p104711319115610"></a><a name="p104711319115610"></a>Unit: MHz</p>
</td>
</tr>
<tr id="row13285243115515"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1047112197565"><a name="p1047112197565"></a><a name="p1047112197565"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8471161912564"><a name="p8471161912564"></a><a name="p8471161912564"></a>npu_chip_info_process_info</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p847161945616"><a name="p847161945616"></a><a name="p847161945616"></a>Information about the process occupying the Ascend AI Processor.</p>
<p id="p647171965610"><a name="p647171965610"></a><a name="p647171965610"></a>Reported only when no process is occupying the Ascend AI Processor, with a value of 0.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16471519205612"><a name="p16471519205612"></a><a name="p16471519205612"></a>Unit: MB</p>
</td>
</tr>
<tr id="row1013365095611"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1238710217578"><a name="p1238710217578"></a><a name="p1238710217578"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1838716285712"><a name="p1838716285712"></a><a name="p1838716285712"></a>npu_chip_info_process_info_PID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1938722125716"><a name="p1938722125716"></a><a name="p1938722125716"></a>Information about the process occupying the Ascend AI Processor, where PID is the process ID on the host; the value is the memory used by the process.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p113876216576"><a name="p113876216576"></a><a name="p113876216576"></a>Unit: MB</p>
</td>
</tr>
<tr id="row18601657145611"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1738710215716"><a name="p1738710215716"></a><a name="p1738710215716"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p16387025573"><a name="p16387025573"></a><a name="p16387025573"></a>npu_chip_info_voltage</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p93871624575"><a name="p93871624575"></a><a name="p93871624575"></a>Ascend AI Processor voltage</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p83871222579"><a name="p83871222579"></a><a name="p83871222579"></a>Unit: Volt (V)</p>
</td>
</tr>
<tr id="row18962022155516"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1376232724113"><a name="p1376232724113"></a><a name="p1376232724113"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p197621927104119"><a name="p197621927104119"></a><a name="p197621927104119"></a>npu_chip_info_serial_number</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11762142754113"><a name="p11762142754113"></a><a name="p11762142754113"></a>Ascend AI Processor serial number</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p158961122195514"><a name="p158961122195514"></a><a name="p158961122195514"></a>-</p>
</td>
</tr>
<tr id="row9224102218282"><td class="cellrowborder" valign="top" width="9.8%" headers="mcps1.2.6.1.1 "><p id="p1095345410280"><a name="p1095345410280"></a><a name="p1095345410280"></a>NPU</p>
</td>
<td class="cellrowborder" valign="top" width="24.279999999999998%" headers="mcps1.2.6.1.2 "><p id="p209531154112810"><a name="p209531154112810"></a><a name="p209531154112810"></a>npu_chip_info_network_status</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p395385413282"><a name="p395385413282"></a><a name="p395385413282"></a><span id="ph1395316544284"><a name="ph1395316544284"></a><a name="ph1395316544284"></a><span id="ph3953105462815"><a name="ph3953105462815"></a><a name="ph3953105462815"></a>Ascend AI Processor</span> network health status</span></p>
</td>
<td class="cellrowborder" valign="top" width="19.68%" headers="mcps1.2.6.1.4 "><p id="p169536543281"><a name="p169536543281"></a><a name="p169536543281"></a>The value is 0, 1, or -1.</p>
<a name="ul1695355411281"></a><a name="ul1695355411281"></a><ul id="ul1695355411281"><li>1: Healthy, reachable</li><li>0: Unhealthy, unreachable</li><li>-1: Unknown (DCMI call failed)</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.94%" headers="mcps1.2.6.1.5 "><a name="ul195418540284"></a><a name="ul195418540284"></a><ul id="ul195418540284"><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li><span id="ph79376229499"><a name="ph79376229499"></a><a name="ph79376229499"></a>Atlas 800I A2 Inference Server</span></li><li><span id="ph288515314573"><a name="ph288515314573"></a><a name="ph288515314573"></a>A200I A2 Box Heterogeneous Subrack</span></li></ul>
</td>
</tr>
<tr><td><p>NPU</p>
</td>
<td><p>npu_chip_info_product_type</p>
</td>
<td><p>Ascend AI Processor product type</p>
</td>
<td><p>1: Placeholder character, having no actual meaning.</p>
</td>
<td><p>Atlas Inference Series Products</p>
</td>
</tr>
</tbody>
</table>

## Utilization Data Information<a name="section1379685784315"></a>

**Table 4** Utilization data information

| Category | Data Information Name | Data Information Description | Unit | Supported Product Forms |
| --- | --- | --- | --- | --- |
| utilization | npu_chip_info_utilization | AICore utilization of the Ascend AI Processor | % |<ul><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Inference Server (with Atlas 300I Inference Card)</li><li>Atlas Inference Series Products</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Subrack</li><li>Atlas 350 PCIe Card</li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>|
| utilization | npu_chip_info_vector_utilization | AIVector utilization of the Ascend AI Processor | % |<ul><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Inference Server (with Atlas 300I Inference Card)</li><li>Atlas Inference Series Products</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Subrack</li><li>Atlas 350 PCIe Card</li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>|
| utilization | npu_chip_info_cube_utilization | AICube utilization of the Ascend AI Processor| % |<ul><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Subrack</li><li>Atlas 350 PCIe Card</li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>|
| utilization | npu_chip_info_overall_utilization | Overall utilization of the Ascend AI Processor | % |<ul><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Subrack</li><li>Atlas 350 PCIe Card</li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>|

## vNPU Data Information<a name="section814111613432"></a>

**Table 5** vNPU data information

<a name="table176992573417"></a>
<table><thead align="left"><tr id="row147006579418"><th class="cellrowborder" valign="top" width="9.8009800980098%" id="mcps1.2.6.1.1"><p id="p1380213250210"><a name="p1380213250210"></a><a name="p1380213250210"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="28.822882288228826%" id="mcps1.2.6.1.2"><p id="p580352532112"><a name="p580352532112"></a><a name="p580352532112"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="25.182518251825186%" id="mcps1.2.6.1.3"><p id="p280372562117"><a name="p280372562117"></a><a name="p280372562117"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="12.55125512551255%" id="mcps1.2.6.1.4"><p id="p1580317257217"><a name="p1580317257217"></a><a name="p1580317257217"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="23.642364236423642%" id="mcps1.2.6.1.5"><p id="p0803132522113"><a name="p0803132522113"></a><a name="p0803132522113"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row1470011579414"><td class="cellrowborder" valign="top" width="9.8009800980098%" headers="mcps1.2.6.1.1 "><p id="p0803142511215"><a name="p0803142511215"></a><a name="p0803142511215"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" width="28.822882288228826%" headers="mcps1.2.6.1.2 "><p id="p78031825162113"><a name="p78031825162113"></a><a name="p78031825162113"></a>vnpu_pod_aicore_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="25.182518251825186%" headers="mcps1.2.6.1.3 "><p id="p7803225132114"><a name="p7803225132114"></a><a name="p7803225132114"></a>AICore utilization of vNPU</p>
</td>
<td class="cellrowborder" valign="top" width="12.55125512551255%" headers="mcps1.2.6.1.4 "><p id="p480452542116"><a name="p480452542116"></a><a name="p480452542116"></a>Unit: %</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.642364236423642%" headers="mcps1.2.6.1.5 "><p id="p553617193312"><a name="p553617193312"></a><a name="p553617193312"></a><span id="ph19590185162111"><a name="ph19590185162111"></a><a name="ph19590185162111"></a>Atlas Inference Series Products</span></p>
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
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p380482517215"><a name="p380482517215"></a><a name="p380482517215"></a>Total vNPU memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1880432511212"><a name="p1880432511212"></a><a name="p1880432511212"></a>Unit: KB</p>
</td>
</tr>
<tr id="row20706115712414"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1180492520213"><a name="p1180492520213"></a><a name="p1180492520213"></a>vNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5805925152119"><a name="p5805925152119"></a><a name="p5805925152119"></a>vnpu_pod_used_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p128051725192114"><a name="p128051725192114"></a><a name="p128051725192114"></a>vNPU memory in use</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p108050258215"><a name="p108050258215"></a><a name="p108050258215"></a>Unit: KB</p>
</td>
</tr>
</tbody>
</table>

## Network Data Information<a name="section1358881214551"></a>

**Table 6** Network data information

<a name="table133306180110"></a>
<table><thead align="left"><tr id="row1033041814119"><td class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.1"><p id="p9187428181111"><a name="p9187428181111"></a><a name="p9187428181111"></a>Category</p>
</td>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.2"><p id="p18187728181116"><a name="p18187728181116"></a><a name="p18187728181116"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.3"><p id="p1118742891112"><a name="p1118742891112"></a><a name="p1118742891112"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.4"><p id="p9187152818113"><a name="p9187152818113"></a><a name="p9187152818113"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.5"><p id="p124519421113"><a name="p124519421113"></a><a name="p124519421113"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row103308189112"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.1 "><p id="p141871428111110"><a name="p141871428111110"></a><a name="p141871428111110"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.2 "><p id="p718722816112"><a name="p718722816112"></a><a name="p718722816112"></a>npu_chip_info_bandwidth_rx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="p14187162841116"><a name="p14187162841116"></a><a name="p14187162841116"></a><span id="ph7187192814119"><a name="ph7187192814119"></a><a name="ph7187192814119"></a><span id="ph4187182811114"><a name="ph4187182811114"></a><a name="ph4187182811114"></a>Ascend AI Processor's</span></span> real-time network port RX rate</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.4 "><p id="p19187928171115"><a name="p19187928171115"></a><a name="p19187928171115"></a>Unit: MB/s</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="20%" headers="mcps1.2.6.1.5 "><a name="ul19245194241111"></a><a name="ul19245194241111"></a><ul id="ul19245194241111"><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li><span id="ph206765471804"><a name="ph206765471804"></a><a name="ph206765471804"></a>Atlas 800I A2 Inference Server</span></li><li><span id="ph6245242151119"><a name="ph6245242151119"></a><a name="ph6245242151119"></a>A200I A2 Box Heterogeneous Subrack</span></li></ul>
</td>
</tr>
<tr id="row93303180110"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p19187142811114"><a name="p19187142811114"></a><a name="p19187142811114"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p01871528121114"><a name="p01871528121114"></a><a name="p01871528121114"></a>npu_chip_info_bandwidth_tx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p21872283116"><a name="p21872283116"></a><a name="p21872283116"></a><span id="ph318772811113"><a name="ph318772811113"></a><a name="ph318772811113"></a><span id="ph141873288114"><a name="ph141873288114"></a><a name="ph141873288114"></a>Ascend AI Processor's</span></span> real-time network port TX rate</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p191872283111"><a name="p191872283111"></a><a name="p191872283111"></a>Unit: MB/s</p>
</td>
</tr>
<tr id="row133011181117"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p181872028151119"><a name="p181872028151119"></a><a name="p181872028151119"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12187122801111"><a name="p12187122801111"></a><a name="p12187122801111"></a>npu_chip_info_link_status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1187142891114"><a name="p1187142891114"></a><a name="p1187142891114"></a><span id="ph91871728141110"><a name="ph91871728141110"></a><a name="ph91871728141110"></a><span id="ph7187132813117"><a name="ph7187132813117"></a><a name="ph7187132813117"></a>Ascend AI Processor's</span></span> network port link status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1118714283116"><a name="p1118714283116"></a><a name="p1118714283116"></a>Value is 0, 1, or -1.</p>
<a name="ul5187128201110"></a><a name="ul5187128201110"></a><ul id="ul5187128201110"><li>1: UP</li><li>0: DOWN</li><li>-1: Unknown (hccn_tool call failed)</li></ul>
</td>
</tr>
<tr id="row7330418111118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p10187122819116"><a name="p10187122819116"></a><a name="p10187122819116"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p218782813114"><a name="p218782813114"></a><a name="p218782813114"></a>npu_chip_link_speed</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p618712861114"><a name="p618712861114"></a><a name="p618712861114"></a><span id="ph15187192831111"><a name="ph15187192831111"></a><a name="ph15187192831111"></a><span id="ph1187102851116"><a name="ph1187102851116"></a><a name="ph1187102851116"></a>Ascend AI Processor's</span></span> default network port speed</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p61871028181113"><a name="p61871028181113"></a><a name="p61871028181113"></a>Unit: MB/s</p>
</td>
</tr>
<tr id="row1133111810118"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p11188142815112"><a name="p11188142815112"></a><a name="p11188142815112"></a>Network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p418852851110"><a name="p418852851110"></a><a name="p418852851110"></a>npu_chip_link_up_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7188142817112"><a name="p7188142817112"></a><a name="p7188142817112"></a><span id="ph1618852818116"><a name="ph1618852818116"></a><a name="ph1618852818116"></a><span id="ph141881428161117"><a name="ph141881428161117"></a><a name="ph141881428161117"></a>Ascend AI Processor's</span></span> network port UP statistics</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p018822815114"><a name="p018822815114"></a><a name="p018822815114"></a>Unit: times</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>Network</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_status_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>Link status of the Ascend AI Processor port.</p><p> X is the Udie ID and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>The value is 0, 1, or -1.</p><ul><li>1: UP</li><li>0: DOWN</li><li>-1: Unknown (hccn_tool call failed)</li></ul>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="19.91%" headers="mcps1.1.7.1.6 "><ul><li>Atlas 350 PCIe Card (4-Processor mesh)</li><li>Atlas 850 Series Hardware</li><li>Atlas 950 SuperPoD</li></ul>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>Network</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>Real-time RX rate of the Ascend AI Processor port.</p><p>X is the Udie ID and Y is the Port ID. The -time parameter used is 100.</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>Unit: MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>Network</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_bandwidth_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>Real-time TX rate of the Ascend AI Processor port.</p><p>X is the Udie ID and Y is the Port ID. The -time parameter used is 100.</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>Unit: MB/s</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="11.21%" headers="mcps1.1.7.1.1 "><p>Network</p>
</td>
<td class="cellrowborder" valign="top" width="21.73%" headers="mcps1.1.7.1.2 "><p>npu_chip_info_link_speed_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="21.61%" headers="mcps1.1.7.1.3 "><p>Speed of the physical port.</p><p> X is the Udie ID and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.1.7.1.5 "><p>Unit: G</p>
</td>
</tr>
</tbody>
</table>

## On-Chip Memory Data Information<a name="section177232045203114"></a>

**Table 7**  On-Chip memory data information

<a name="table728745315300"></a>
<table><thead align="left"><tr id="row72881853103019"><th class="cellrowborder" valign="top" width="9.21%" id="mcps1.2.6.1.1"><p id="p11126227312"><a name="p11126227312"></a><a name="p11126227312"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="26.97%" id="mcps1.2.6.1.2"><p id="p21121225312"><a name="p21121225312"></a><a name="p21121225312"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.6.1.3"><p id="p11123222316"><a name="p11123222316"></a><a name="p11123222316"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="14.91%" id="mcps1.2.6.1.4"><p id="p17112182293120"><a name="p17112182293120"></a><a name="p17112182293120"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="26.61%" id="mcps1.2.6.1.5"><p id="p71121822153119"><a name="p71121822153119"></a><a name="p71121822153119"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row1288953163016"><td class="cellrowborder" valign="top" width="9.21%" headers="mcps1.2.6.1.1 "><p id="p537581413115"><a name="p537581413115"></a><a name="p537581413115"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" width="26.97%" headers="mcps1.2.6.1.2 "><p id="p1637571420317"><a name="p1637571420317"></a><a name="p1637571420317"></a>npu_chip_info_hbm_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.6.1.3 "><p id="p1837671483116"><a name="p1837671483116"></a><a name="p1837671483116"></a><span id="ph837613143312"><a name="ph837613143312"></a><a name="ph837613143312"></a><span id="ph3376131453120"><a name="ph3376131453120"></a><a name="ph3376131453120"></a>Used on-chip memory of the</span></span> Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" width="14.91%" headers="mcps1.2.6.1.4 "><p id="p1376214173120"><a name="p1376214173120"></a><a name="p1376214173120"></a>Unit: MB</p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="26.61%" headers="mcps1.2.6.1.5 "><a name="ul1737721403120"></a><a name="ul1737721403120"></a><ul id="ul1737721403120"><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li>Atlas A3 Training Series Products</li><li><span id="ph043025116483"><a name="ph043025116483"></a><a name="ph043025116483"></a>A200I A2 Box Heterogeneous Subrack</span></li><li><span id="ph1201913534"><a name="ph1201913534"></a><a name="ph1201913534"></a>Atlas 800I A2 Inference Server</span></li></ul><ul><li><span>Atlas 350 PCIe Card</span></li><li>Atlas 850 Series Hardware Products</li><li>Atlas 950 SuperPoD</li></ul>
<p id="p744164714207"><a name="p744164714207"></a><a name="p744164714207"></a></p>
<p id="p1642204722019"><a name="p1642204722019"></a><a name="p1642204722019"></a></p>
</td>
</tr>
<tr id="row8376151714335"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p930919134154"><a name="p930919134154"></a><a name="p930919134154"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13687292332"><a name="p13687292332"></a><a name="p13687292332"></a>npu_chip_info_hbm_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15368729183317"><a name="p15368729183317"></a><a name="p15368729183317"></a>Total on-chip memory of the Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1537617171338"><a name="p1537617171338"></a><a name="p1537617171338"></a>Unit: MB</p>
</td>
</tr>
<tr id="row152881353163011"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1937671413114"><a name="p1937671413114"></a><a name="p1937671413114"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p137614149311"><a name="p137614149311"></a><a name="p137614149311"></a>npu_chip_info_hbm_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1737631433115"><a name="p1737631433115"></a><a name="p1737631433115"></a><span id="ph143778147311"><a name="ph143778147311"></a><a name="ph143778147311"></a><span id="ph10377161417316"><a name="ph10377161417316"></a><a name="ph10377161417316"></a>On-chip memory utilization of the</span></span> Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1337761412315"><a name="p1337761412315"></a><a name="p1337761412315"></a>Unit: %</p>
</td>
</tr>
<tr id="row1944471872714"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1899921617152"><a name="p1899921617152"></a><a name="p1899921617152"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p208411785402"><a name="p208411785402"></a><a name="p208411785402"></a>npu_chip_info_hbm_ecc_enable_flag</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p12841489405"><a name="p12841489405"></a><a name="p12841489405"></a>On-chip memory ECC enabling status of the Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1898013114817"><a name="p1898013114817"></a><a name="p1898013114817"></a>The value is 1 or 0.</p>
<a name="ul089881311484"></a><a name="ul089881311484"></a><ul id="ul089881311484"><li>0: ECC detection not enabled.</li><li>1: ECC detection enabled.</li></ul>
</td>
</tr>
<tr id="row1713612195407"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p11591181911516"><a name="p11591181911516"></a><a name="p11591181911516"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1149010582404"><a name="p1149010582404"></a><a name="p1149010582404"></a>npu_chip_info_hbm_ecc_single_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1449045884011"><a name="p1449045884011"></a><a name="p1449045884011"></a>Single‑bit error count of the Ascend AI Processor's on‑chip memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p21361419184019"><a name="p21361419184019"></a><a name="p21361419184019"></a>-</p>
</td>
</tr>
<tr id="row51361219184015"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5700142221516"><a name="p5700142221516"></a><a name="p5700142221516"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5608101654112"><a name="p5608101654112"></a><a name="p5608101654112"></a>npu_chip_info_hbm_ecc_double_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p160871694114"><a name="p160871694114"></a><a name="p160871694114"></a>Double-bit error count of the Ascend AI processor's on-chip memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9136181915405"><a name="p9136181915405"></a><a name="p9136181915405"></a>-</p>
</td>
</tr>
<tr id="row41361919134011"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12199426201519"><a name="p12199426201519"></a><a name="p12199426201519"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p65349266419"><a name="p65349266419"></a><a name="p65349266419"></a>npu_chip_info_hbm_ecc_total_single_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p753442694119"><a name="p753442694119"></a><a name="p753442694119"></a>Total number of all single‑bit errors over the lifetime of the Ascend AI Processor's on‑chip memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p413618196403"><a name="p413618196403"></a><a name="p413618196403"></a>-</p>
</td>
</tr>
<tr id="row14137719124010"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p101371819194013"><a name="p101371819194013"></a><a name="p101371819194013"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p101041130121610"><a name="p101041130121610"></a><a name="p101041130121610"></a>npu_chip_info_hbm_ecc_total_double_bit_error_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1810412305163"><a name="p1810412305163"></a><a name="p1810412305163"></a>Total number of all double‑bit errors over the lifetime of the Ascend AI Processor's on‑chip memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p161377192404"><a name="p161377192404"></a><a name="p161377192404"></a>-</p>
</td>
</tr>
<tr id="row65841633144013"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p155841733204015"><a name="p155841733204015"></a><a name="p155841733204015"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p37841238171612"><a name="p37841238171612"></a><a name="p37841238171612"></a>npu_chip_info_hbm_ecc_single_bit_isolated_pages_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9784173861615"><a name="p9784173861615"></a><a name="p9784173861615"></a>Number of memory pages isolated due to single‑bit errors in the Ascend AI Processor's on‑chip memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p14584333134019"><a name="p14584333134019"></a><a name="p14584333134019"></a>-</p>
</td>
</tr>
<tr id="row15584173374015"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p11194724184"><a name="p11194724184"></a><a name="p11194724184"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p41947271811"><a name="p41947271811"></a><a name="p41947271811"></a>npu_chip_info_hbm_ecc_double_bit_isolated_pages_cnt</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p819417291813"><a name="p819417291813"></a><a name="p819417291813"></a>Number of memory pages isolated due to double‑bit errors in the Ascend AI Processor's on‑chip memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p6584633144018"><a name="p6584633144018"></a><a name="p6584633144018"></a>-</p>
</td>
</tr>
<tr id="row838319296204"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16912264198"><a name="p16912264198"></a><a name="p16912264198"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5691826121920"><a name="p5691826121920"></a><a name="p5691826121920"></a>npu_chip_info_hbm_temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p12691112601911"><a name="p12691112601911"></a><a name="p12691112601911"></a>On-chip memory temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1871185411812"><a name="p1871185411812"></a><a name="p1871185411812"></a>Unit: &deg;C</p>
</td>
</tr>
<tr id="row6486636112012"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p36914261197"><a name="p36914261197"></a><a name="p36914261197"></a>On-Chip Memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5691142631920"><a name="p5691142631920"></a><a name="p5691142631920"></a>npu_chip_info_hbm_bandwidth_utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p86911726101915"><a name="p86911726101915"></a><a name="p86911726101915"></a>On-chip memory bandwidth utilization</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1130779198"><a name="p1130779198"></a><a name="p1130779198"></a>Unit: %</p>
</td>
</tr>
</tbody>
</table>

## HCCS Data Information<a name="section039816240252"></a>

**Table 8**  HCCS data information

<a name="table9399845122516"></a>
<table><thead align="left"><tr id="row153998454253"><th class="cellrowborder" valign="top" width="8.950000000000001%" id="mcps1.2.6.1.1"><p id="p94879162267"><a name="p94879162267"></a><a name="p94879162267"></a>Category</p>
</th>
<td class="cellrowborder" valign="top" width="28.190000000000005%" id="mcps1.2.6.1.2"><p id="p10487151652611"><a name="p10487151652611"></a><a name="p10487151652611"></a>Data Information Name</p>
</td>
<th class="cellrowborder" valign="top" width="21.990000000000006%" id="mcps1.2.6.1.3"><p id="p1648711682615"><a name="p1648711682615"></a><a name="p1648711682615"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="15.570000000000004%" id="mcps1.2.6.1.4"><p id="p174873165265"><a name="p174873165265"></a><a name="p174873165265"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="25.300000000000004%" id="mcps1.2.6.1.5"><p id="p0487161620269"><a name="p0487161620269"></a><a name="p0487161620269"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row2040184552518"><td class="cellrowborder" valign="top" width="8.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p14401145172516"><a name="p14401145172516"></a><a name="p14401145172516"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" width="28.190000000000005%" headers="mcps1.2.6.1.2 "><p id="p338114141712"><a name="p338114141712"></a><a name="p338114141712"></a>npu_chip_info_hccs_statistic_info_tx_cnt_X</p>
<p id="p1024171512455"><a name="p1024171512455"></a><a name="p1024171512455"></a>X range: 1~7 (Atlas A2 Training Series products or Atlas 900 A3 SuperPoD), 2~7 (Atlas 9000 A3 SuperPoD)</p>
</td>
<td class="cellrowborder" valign="top" width="21.990000000000006%" headers="mcps1.2.6.1.3 "><a name="ul1424913612438"></a><a name="ul1424913612438"></a><ul id="ul1424913612438"><li>Number of packets sent on the X-th HDLC link, unit is flit.</li><li>-1 is reported if collection fails.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="15.570000000000004%" headers="mcps1.2.6.1.4 "><p id="p840184516254"><a name="p840184516254"></a><a name="p840184516254"></a>-</p>
</td>
<td class="cellrowborder" rowspan="8" valign="top" width="25.300000000000004%" headers="mcps1.2.6.1.5 "><a name="ul11925372813"></a><a name="ul11925372813"></a><ul id="ul11925372813"><li>Atlas A2 Training Series products</li><li>Atlas A3 Training Series products</li></ul>
</td>
</tr>
<tr id="row1140184517258"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1240115459259"><a name="p1240115459259"></a><a name="p1240115459259"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p186021875020"><a name="p186021875020"></a><a name="p186021875020"></a>npu_chip_info_hccs_statistic_info_rx_cnt_X</p>
<p id="p11562871013"><a name="p11562871013"></a><a name="p11562871013"></a>X range: 1~7 (Atlas A2 Training Series products or Atlas 900 A3 SuperPoD), 2~7 (Atlas 9000 A3 SuperPoD)</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1234520167435"></a><a name="ul1234520167435"></a><ul id="ul1234520167435"><li>Number of packets received on the X-th HDLC link, unit is flit.</li><li>-1 is reported if collection fails.</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1668141520283"><a name="p1668141520283"></a><a name="p1668141520283"></a>-</p>
</td>
</tr>
<tr id="row1240254522514"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5402245122513"><a name="p5402245122513"></a><a name="p5402245122513"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1139102394715"><a name="p1139102394715"></a><a name="p1139102394715"></a>npu_chip_info_hccs_statistic_info_crc_err_cnt_X</p>
<p id="p1429534131018"><a name="p1429534131018"></a><a name="p1429534131018"></a>X range: 1~7 (Atlas A2 Training Series products or Atlas 900 A3 SuperPoD), 2~7 (Atlas 9000 A3 SuperPoD)</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul112792234316"></a><a name="ul112792234316"></a><ul id="ul112792234316"><li>CRC error in received packets on the X‑th HDLC link, unit is flit.</li><li>-1 is reported if collection fails.</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p712111832818"><a name="p712111832818"></a><a name="p712111832818"></a>-</p>
</td>
</tr>
<tr id="row1461410598130"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p193164531143"><a name="p193164531143"></a><a name="p193164531143"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p549382011151"><a name="p549382011151"></a><a name="p549382011151"></a>npu_chip_info_hccs_bandwidth_info_profiling_time</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p549382012152"><a name="p549382012152"></a><a name="p549382012152"></a>HCCS link bandwidth sampling duration, ranging from 1 to 1000.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p04931220171514"><a name="p04931220171514"></a><a name="p04931220171514"></a>Unit: ms</p>
</td>
</tr>
<tr id="row1620614518148"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p18316753201414"><a name="p18316753201414"></a><a name="p18316753201414"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7493220191514"><a name="p7493220191514"></a><a name="p7493220191514"></a>npu_chip_info_hccs_bandwidth_info_total_tx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11493120161511"><a name="p11493120161511"></a><a name="p11493120161511"></a>Total TX bandwidth of the HCCS link. -1 is reported if collection fails.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p204931620131517"><a name="p204931620131517"></a><a name="p204931620131517"></a>Unit: GB/s</p>
</td>
</tr>
<tr id="row122341714181411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3316185361418"><a name="p3316185361418"></a><a name="p3316185361418"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p149302010159"><a name="p149302010159"></a><a name="p149302010159"></a>npu_chip_info_hccs_bandwidth_info_total_rx</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p7493820131511"><a name="p7493820131511"></a><a name="p7493820131511"></a>Total RX bandwidth of the HCCS link. -1 is reported if collection fails.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p164936208158"><a name="p164936208158"></a><a name="p164936208158"></a>Unit: GB/s</p>
</td>
</tr>
<tr id="row1853162231416"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p153161253161412"><a name="p153161253161412"></a><a name="p153161253161412"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p549352091516"><a name="p549352091516"></a><a name="p549352091516"></a>npu_chip_info_hccs_bandwidth_info_tx_X</p>
<p id="p2493172021511"><a name="p2493172021511"></a><a name="p2493172021511"></a>X range: 1~7 (Atlas A2 Training Series products, Atlas 900 A3 SuperPoD), 2~7 (Atlas 9000 A3 SuperPoD).</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3493152051511"><a name="p3493152051511"></a><a name="p3493152051511"></a>TX data bandwidth for a single HCCS link. -1 is reported if collection fails.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p194946203153"><a name="p194946203153"></a><a name="p194946203153"></a>Unit: GB/s</p>
</td>
</tr>
<tr id="row18299930131419"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1031665361416"><a name="p1031665361416"></a><a name="p1031665361416"></a>HCCS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7494122091513"><a name="p7494122091513"></a><a name="p7494122091513"></a>npu_chip_info_hccs_bandwidth_info_rx_X</p>
<p id="p10494720131520"><a name="p10494720131520"></a><a name="p10494720131520"></a>X range: 1–7 (Atlas A2 Training Series products, Atlas 900 A3 SuperPoD), 2–7 (Atlas 9000 A3 SuperPoD)</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p449492011517"><a name="p449492011517"></a><a name="p449492011517"></a>RX data bandwidth for a single HCCS link. -1 is reported if collection fails.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1649442041518"><a name="p1649442041518"></a><a name="p1649442041518"></a>Unit: GB/s</p>
</td>
</tr>
</tbody>
</table>

## PCIe Data Information<a name="section1240520241824136"></a>

**Table 9** PCIe data information

<a name="table1341911380255"></a>
<table><thead align="left"><tr id="row941993842520"><th class="cellrowborder" valign="top" width="9.85%" id="mcps1.2.6.1.1"><p id="p5917122513215"><a name="p5917122513215"></a><a name="p5917122513215"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="21.37%" id="mcps1.2.6.1.2"><p id="p14918182519219"><a name="p14918182519219"></a><a name="p14918182519219"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="26.279999999999998%" id="mcps1.2.6.1.3"><p id="p6406358175813"><a name="p6406358175813"></a><a name="p6406358175813"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="11.59%" id="mcps1.2.6.1.4"><p id="p79189258216"><a name="p79189258216"></a><a name="p79189258216"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="30.91%" id="mcps1.2.6.1.5"><p id="p169182258212"><a name="p169182258212"></a><a name="p169182258212"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row174206389252"><td class="cellrowborder" valign="top" width="9.85%" headers="mcps1.2.6.1.1 "><p id="p79181025142113"><a name="p79181025142113"></a><a name="p79181025142113"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" width="21.37%" headers="mcps1.2.6.1.2 "><p id="p59181925142118"><a name="p59181925142118"></a><a name="p59181925142118"></a>npu_chip_info_pcie_rx_p_bw</p>
</td>
<td class="cellrowborder" valign="top" width="26.279999999999998%" headers="mcps1.2.6.1.3 "><p id="p243612169285"><a name="p243612169285"></a><a name="p243612169285"></a><span id="ph643661692816"><a name="ph643661692816"></a><a name="ph643661692816"></a>PCIe bandwidth for remote writes</span> received by the Ascend AI Processor.</p>
</td>
<td class="cellrowborder" valign="top" width="11.59%" headers="mcps1.2.6.1.4 "><p id="p1191992510210"><a name="p1191992510210"></a><a name="p1191992510210"></a>Unit: MB/ms</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="30.91%" headers="mcps1.2.6.1.5 "><a name="ul64395165289"></a><a name="ul64395165289"></a><ul id="ul64395165289"><li><p id="li2043917168282p0"><a name="li2043917168282p0"></a><a name="li2043917168282p0"></a>Atlas A2 Training Series products</p>
</li><li><p id="li114396166286p0"><a name="li114396166286p0"></a><a name="li114396166286p0"></a><span id="ph1722042181618"><a name="ph1722042181618"></a><a name="ph1722042181618"></a>Atlas 800I A2 Inference Server</span></p>
</li><li><p id="p258313228429"><a name="p258313228429"></a><a name="p258313228429"></a><span id="ph18642192315427"><a name="ph18642192315427"></a><a name="ph18642192315427"></a>A200I A2 Box Heterogeneous Subrack</span></p>
</li></ul>
</td>
</tr>
<tr id="row34233384255"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7919182532111"><a name="p7919182532111"></a><a name="p7919182532111"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1091942532114"><a name="p1091942532114"></a><a name="p1091942532114"></a>npu_chip_info_pcie_rx_np_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p174481716202815"><a name="p174481716202815"></a><a name="p174481716202815"></a><span id="ph10448141613285"><a name="ph10448141613285"></a><a name="ph10448141613285"></a>PCIe bandwidth for remote reads</span> received by the Ascend AI Processor.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p20919225172118"><a name="p20919225172118"></a><a name="p20919225172118"></a>Unit: MB/ms</p>
</td>
</tr>
<tr id="row54261838112518"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p15919122502119"><a name="p15919122502119"></a><a name="p15919122502119"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p139191025142118"><a name="p139191025142118"></a><a name="p139191025142118"></a>npu_chip_info_pcie_rx_cpl_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1945841616288"><a name="p1945841616288"></a><a name="p1945841616288"></a><span id="ph1745831612816"><a name="ph1745831612816"></a><a name="ph1745831612816"></a>PCIe bandwidth for receiving CPL responses to remote reads</span> by the Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16920122520215"><a name="p16920122520215"></a><a name="p16920122520215"></a>Unit: MB/ms</p>
</td>
</tr>
<tr id="row184291938102510"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p992032582117"><a name="p992032582117"></a><a name="p992032582117"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p29201625102111"><a name="p29201625102111"></a><a name="p29201625102111"></a>npu_chip_info_pcie_tx_p_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p746851613289"><a name="p746851613289"></a><a name="p746851613289"></a><span id="ph10468616112812"><a name="ph10468616112812"></a><a name="ph10468616112812"></a>PCIe bandwidth for remote writes</span> from the Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17920152552119"><a name="p17920152552119"></a><a name="p17920152552119"></a>Unit: MB/ms</p>
</td>
</tr>
<tr id="row34321838132519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7921125152120"><a name="p7921125152120"></a><a name="p7921125152120"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1692115258214"><a name="p1692115258214"></a><a name="p1692115258214"></a>npu_chip_info_pcie_tx_np_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9406658175816"><a name="p9406658175816"></a><a name="p9406658175816"></a><span id="ph1148018161281"><a name="ph1148018161281"></a><a name="ph1148018161281"></a>PCIe bandwidth for remote reads</span> from the Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p169211251214"><a name="p169211251214"></a><a name="p169211251214"></a>Unit: MB/ms</p>
</td>
</tr>
<tr id="row743773816250"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6921325192115"><a name="p6921325192115"></a><a name="p6921325192115"></a>PCIe</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p692118258210"><a name="p692118258210"></a><a name="p692118258210"></a>npu_chip_info_pcie_tx_cpl_bw</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p144901216132813"><a name="p144901216132813"></a><a name="p144901216132813"></a><span id="ph84901416192811"><a name="ph84901416192811"></a><a name="ph84901416192811"></a>PCIe bandwidth for the Ascend AI Processor to</span> send CPL replies to remote read operations</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p119211225102110"><a name="p119211225102110"></a><a name="p119211225102110"></a>Unit: MB/ms</p>
</td>
</tr>
</tbody>
</table>

## RoCE Data Information<a name="section184516450323"></a>

**Table 10**  RoCE data information

<a name="table1562691116332"></a>
<table><thead align="left"><tr id="row126261011133318"><th class="cellrowborder" valign="top" width="9.01%" id="mcps1.2.6.1.1"><p id="p149115619355"><a name="p149115619355"></a><a name="p149115619355"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="27.07%" id="mcps1.2.6.1.2"><p id="p15911466357"><a name="p15911466357"></a><a name="p15911466357"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="22.1%" id="mcps1.2.6.1.3"><p id="p14911667358"><a name="p14911667358"></a><a name="p14911667358"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="13.52%" id="mcps1.2.6.1.4"><p id="p59116653520"><a name="p59116653520"></a><a name="p59116653520"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="28.299999999999997%" id="mcps1.2.6.1.5"><p id="p199115623510"><a name="p199115623510"></a><a name="p199115623510"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row862751110336"><td class="cellrowborder" valign="top" width="9.01%" headers="mcps1.2.6.1.1 "><p id="p6133145318341"><a name="p6133145318341"></a><a name="p6133145318341"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" width="27.07%" headers="mcps1.2.6.1.2 "><p id="p17133115323411"><a name="p17133115323411"></a><a name="p17133115323411"></a>npu_chip_mac_rx_pause_num</p>
</td>
<td class="cellrowborder" valign="top" width="22.1%" headers="mcps1.2.6.1.3 "><p id="p5134105313348"><a name="p5134105313348"></a><a name="p5134105313348"></a>Total number of Pause frames received by MAC</p>
</td>
<td class="cellrowborder" valign="top" width="13.52%" headers="mcps1.2.6.1.4 "><p id="p1413585319348"><a name="p1413585319348"></a><a name="p1413585319348"></a>-</p>
</td>
<td class="cellrowborder" rowspan="21" valign="top" width="28.299999999999997%" headers="mcps1.2.6.1.5 "><a name="ul3135253123412"></a><a name="ul3135253123412"></a><ul id="ul3135253123412"><li>Atlas Training Series products</li><li>Atlas A2 Training Series products</li><li>Atlas A3 Training Series products</li><li><span id="ph1241332842611"><a name="ph1241332842611"></a><a name="ph1241332842611"></a>Atlas 800I A2 Inference Server</span></li><li><span id="ph6496152317452"><a name="ph6496152317452"></a><a name="ph6496152317452"></a>A200I A2 Box Heterogeneous Subrack</span></li></ul>
</td>
</tr>
<tr id="row762714116330"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p13135135314348"><a name="p13135135314348"></a><a name="p13135135314348"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p9135125317341"><a name="p9135125317341"></a><a name="p9135125317341"></a>npu_chip_mac_tx_pause_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p4136353183411"><a name="p4136353183411"></a><a name="p4136353183411"></a>Total number of Pause frames sent by MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3136185317341"><a name="p3136185317341"></a><a name="p3136185317341"></a>-</p>
</td>
</tr>
<tr id="row1562751114333"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17136155363415"><a name="p17136155363415"></a><a name="p17136155363415"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1513735318347"><a name="p1513735318347"></a><a name="p1513735318347"></a>npu_chip_mac_rx_pfc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p113710536343"><a name="p113710536343"></a><a name="p113710536343"></a>Total number of PFC frames received by MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1213755314346"><a name="p1213755314346"></a><a name="p1213755314346"></a>-</p>
</td>
</tr>
<tr id="row14628211153320"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p181389536343"><a name="p181389536343"></a><a name="p181389536343"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p18138145313416"><a name="p18138145313416"></a><a name="p18138145313416"></a>npu_chip_mac_tx_pfc_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p91385532343"><a name="p91385532343"></a><a name="p91385532343"></a>Total number of PFC frames sent by MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p8138205373419"><a name="p8138205373419"></a><a name="p8138205373419"></a>-</p>
</td>
</tr>
<tr id="row1562881110335"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1113965343414"><a name="p1113965343414"></a><a name="p1113965343414"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p171391253143419"><a name="p171391253143419"></a><a name="p171391253143419"></a>npu_chip_mac_rx_bad_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10139175315340"><a name="p10139175315340"></a><a name="p10139175315340"></a>Total number of bad packets received by MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16139125393416"><a name="p16139125393416"></a><a name="p16139125393416"></a>-</p>
</td>
</tr>
<tr id="row862816116333"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p151391753173410"><a name="p151391753173410"></a><a name="p151391753173410"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p171407538349"><a name="p171407538349"></a><a name="p171407538349"></a>npu_chip_mac_tx_bad_oct_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p14140853173416"><a name="p14140853173416"></a><a name="p14140853173416"></a>Total number of bad packet bytes sent MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11401153113412"><a name="p11401153113412"></a><a name="p11401153113412"></a>-</p>
</td>
</tr>
<tr id="row1162961119337"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1714014537344"><a name="p1714014537344"></a><a name="p1714014537344"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p41401253143415"><a name="p41401253143415"></a><a name="p41401253143415"></a>npu_chip_mac_rx_bad_oct_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18141155311348"><a name="p18141155311348"></a><a name="p18141155311348"></a>Total number of bad packet bytes received by MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p51411053173410"><a name="p51411053173410"></a><a name="p51411053173410"></a>-</p>
</td>
</tr>
<tr id="row693213413419"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p15141753183414"><a name="p15141753183414"></a><a name="p15141753183414"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p171411533341"><a name="p171411533341"></a><a name="p171411533341"></a>npu_chip_mac_tx_bad_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p114213531347"><a name="p114213531347"></a><a name="p114213531347"></a>Total number of bad packets sent by MAC</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p314295312342"><a name="p314295312342"></a><a name="p314295312342"></a>-</p>
</td>
</tr>
<tr id="row17933124113410"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12142185318343"><a name="p12142185318343"></a><a name="p12142185318343"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p614214538344"><a name="p614214538344"></a><a name="p614214538344"></a>npu_chip_roce_rx_all_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p5143125311343"><a name="p5143125311343"></a><a name="p5143125311343"></a>Total number of packets received by RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1114305343415"><a name="p1114305343415"></a><a name="p1114305343415"></a>-</p>
</td>
</tr>
<tr id="row1293474173411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14143185363416"><a name="p14143185363416"></a><a name="p14143185363416"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p814320538348"><a name="p814320538348"></a><a name="p814320538348"></a>npu_chip_roce_tx_all_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p61441953123412"><a name="p61441953123412"></a><a name="p61441953123412"></a>Total number of packets sent by RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p10144135310342"><a name="p10144135310342"></a><a name="p10144135310342"></a>-</p>
</td>
</tr>
<tr id="row109345413348"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1114513534341"><a name="p1114513534341"></a><a name="p1114513534341"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8145353193418"><a name="p8145353193418"></a><a name="p8145353193418"></a>npu_chip_roce_rx_err_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p814665311347"><a name="p814665311347"></a><a name="p814665311347"></a>Total number of bad packets received by RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p18146155316347"><a name="p18146155316347"></a><a name="p18146155316347"></a>-</p>
</td>
</tr>
<tr id="row4935124133411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p71476534343"><a name="p71476534343"></a><a name="p71476534343"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p10147185333416"><a name="p10147185333416"></a><a name="p10147185333416"></a>npu_chip_roce_tx_err_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1214775393414"><a name="p1214775393414"></a><a name="p1214775393414"></a>Total number of bad packets sent by RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p201477531346"><a name="p201477531346"></a><a name="p201477531346"></a>-</p>
</td>
</tr>
<tr id="row99353413412"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p614811539346"><a name="p614811539346"></a><a name="p614811539346"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p614995363419"><a name="p614995363419"></a><a name="p614995363419"></a>npu_chip_roce_rx_cnp_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1114915363411"><a name="p1114915363411"></a><a name="p1114915363411"></a>Number of CNP type packets received by RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p2149115323411"><a name="p2149115323411"></a><a name="p2149115323411"></a>-</p>
</td>
</tr>
<tr id="row693504203411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p12150125313417"><a name="p12150125313417"></a><a name="p12150125313417"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1115118535345"><a name="p1115118535345"></a><a name="p1115118535345"></a>npu_chip_roce_tx_cnp_pkt_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p111516539341"><a name="p111516539341"></a><a name="p111516539341"></a>Number of CNP type packets sent by RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p81514535345"><a name="p81514535345"></a><a name="p81514535345"></a>-</p>
</td>
</tr>
<tr id="row149361242344"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p315275317344"><a name="p315275317344"></a><a name="p315275317344"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8152653183414"><a name="p8152653183414"></a><a name="p8152653183414"></a>npu_chip_roce_new_pkt_rty_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p21521153143412"><a name="p21521153143412"></a><a name="p21521153143412"></a>Statistics on the number of retries sent over RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p5152353173416"><a name="p5152353173416"></a><a name="p5152353173416"></a>-</p>
</td>
</tr>
<tr id="row1893410440349"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14153205323412"><a name="p14153205323412"></a><a name="p14153205323412"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p18154253193418"><a name="p18154253193418"></a><a name="p18154253193418"></a>npu_chip_roce_unexpected_ack_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p31543535343"><a name="p31543535343"></a><a name="p31543535343"></a>Number of unexpected ACK packets received by RoCE. The NPU discards them, which does not affect services.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p151541353153412"><a name="p151541353153412"></a><a name="p151541353153412"></a>-</p>
</td>
</tr>
<tr id="row0935164412342"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p31551553153414"><a name="p31551553153414"></a><a name="p31551553153414"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p915565393419"><a name="p915565393419"></a><a name="p915565393419"></a>npu_chip_roce_out_of_order_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2155155311342"><a name="p2155155311342"></a><a name="p2155155311342"></a>Number of packets received by RoCE with a PSN greater than the expected PSN, or duplicate PSN packets. Out‑of‑order delivery or packet loss triggers transmission retries.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p20155125314344"><a name="p20155125314344"></a><a name="p20155125314344"></a>-</p>
</td>
</tr>
<tr id="row1936144463416"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p915613539340"><a name="p915613539340"></a><a name="p915613539340"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p31561153123417"><a name="p31561153123417"></a><a name="p31561153123417"></a>npu_chip_roce_verification_err_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2156175353417"><a name="p2156175353417"></a><a name="p2156175353417"></a>Number of packets received by RoCE that failed field verification. Field verification scenarios include: ICRC, packet length, destination port number, etc.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p215675314348"><a name="p215675314348"></a><a name="p215675314348"></a>-</p>
</td>
</tr>
<tr id="row793704413343"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p10157155314349"><a name="p10157155314349"></a><a name="p10157155314349"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p71577532341"><a name="p71577532341"></a><a name="p71577532341"></a>npu_chip_roce_qp_status_err_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p91571453163410"><a name="p91571453163410"></a><a name="p91571453163410"></a>Number of packets received by RoCE that are generated due to abnormal QP connection status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p415795323418"><a name="p415795323418"></a><a name="p415795323418"></a>-</p>
</td>
</tr>
<tr id="row1334918433510"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1349443653"><a name="p1349443653"></a><a name="p1349443653"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p123491043457"><a name="p123491043457"></a><a name="p123491043457"></a>npu_chip_info_rx_ecn_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p103491431157"><a name="p103491431157"></a><a name="p103491431157"></a>Number of ECN marks received by the Ascend AI Processor network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p103499431552"><a name="p103499431552"></a><a name="p103499431552"></a>-</p>
</td>
</tr>
<tr id="row1433394916519"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p833313493511"><a name="p833313493511"></a><a name="p833313493511"></a>RoCE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14334174915514"><a name="p14334174915514"></a><a name="p14334174915514"></a>npu_chip_info_rx_fcs_num</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p6334349155"><a name="p6334349155"></a><a name="p6334349155"></a>Number of FCS marks received by the Ascend AI processor network</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p123341491158"><a name="p123341491158"></a><a name="p123341491158"></a>-</p>
</td>
</tr>
</tbody>
</table>

## SIO Data Information<a name="section7109037161515"></a>

**Table 11**  SIO data information

<a name="table1910972371718"></a>
<table><thead align="left"><tr id="row10109122321710"><th class="cellrowborder" valign="top" width="9.509999999999998%" id="mcps1.2.6.1.1"><p id="p18120184111178"><a name="p18120184111178"></a><a name="p18120184111178"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="28.559999999999995%" id="mcps1.2.6.1.2"><p id="p1121144111714"><a name="p1121144111714"></a><a name="p1121144111714"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="21.109999999999996%" id="mcps1.2.6.1.3"><p id="p312116415171"><a name="p312116415171"></a><a name="p312116415171"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="17.509999999999998%" id="mcps1.2.6.1.4"><p id="p1212112418173"><a name="p1212112418173"></a><a name="p1212112418173"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="23.309999999999995%" id="mcps1.2.6.1.5"><p id="p12121204113171"><a name="p12121204113171"></a><a name="p12121204113171"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row12110123121718"><td class="cellrowborder" valign="top" width="9.509999999999998%" headers="mcps1.2.6.1.1 "><p id="p6113184916172"><a name="p6113184916172"></a><a name="p6113184916172"></a>SIO</p>
</td>
<td class="cellrowborder" valign="top" width="28.559999999999995%" headers="mcps1.2.6.1.2 "><p id="p211316490178"><a name="p211316490178"></a><a name="p211316490178"></a>npu_chip_info_sio_crc_tx_err_cnt</p>
</td>
<td class="cellrowborder" valign="top" width="21.109999999999996%" headers="mcps1.2.6.1.3 "><p id="p6113194914170"><a name="p6113194914170"></a><a name="p6113194914170"></a>Number of bad packets sent by SIO</p>
</td>
<td class="cellrowborder" valign="top" width="17.509999999999998%" headers="mcps1.2.6.1.4 "><p id="p10113449181711"><a name="p10113449181711"></a><a name="p10113449181711"></a>-</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="23.309999999999995%" headers="mcps1.2.6.1.5 "><ul><li>Atlas A3 Training Series products</li><li>Atlas 350 standard card</li><li>Atlas 850 Series Hardware products</li><li>Atlas 950 SuperPoD</li></ul>
</td>
</tr>
<tr id="row1111082310171"><td class="cellrowborder" valign="top" width="9.509999999999998%" headers="mcps1.2.6.1.1 "><p id="p10114204910174"><a name="p10114204910174"></a><a name="p10114204910174"></a>SIO</p>
</td>
<td class="cellrowborder" valign="top" width="28.559999999999995%" headers="mcps1.2.6.1.2 "><p id="p411484911177"><a name="p411484911177"></a><a name="p411484911177"></a>npu_chip_info_sio_crc_rx_err_cnt</p>
</td>
<td class="cellrowborder" valign="top" width="21.109999999999996%" headers="mcps1.2.6.1.3 "><p id="p1911454914173"><a name="p1911454914173"></a><a name="p1911454914173"></a>Number of bad packets received by SIO</p>
</td>
<td class="cellrowborder" valign="top" width="17.509999999999998%" headers="mcps1.2.6.1.4 "><p id="p9114104916179"><a name="p9114104916179"></a><a name="p9114104916179"></a>-</p>
</td>
</tr>
</tbody>
</table>

## Optical Module Data Information<a name="section1517163183510"></a>

**Table 12**  Optical module data information

<a name="table1379935213357"></a>
<table><thead align="left"><tr id="row1279915521353"><th class="cellrowborder" valign="top" width="9.08%" id="mcps1.2.6.1.1"><p id="p570012555368"><a name="p570012555368"></a><a name="p570012555368"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="27.27%" id="mcps1.2.6.1.2"><p id="p670145511363"><a name="p670145511363"></a><a name="p670145511363"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="19.24%" id="mcps1.2.6.1.3"><p id="p14701955113619"><a name="p14701955113619"></a><a name="p14701955113619"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="17.52%" id="mcps1.2.6.1.4"><p id="p5701115514361"><a name="p5701115514361"></a><a name="p5701115514361"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="26.889999999999997%" id="mcps1.2.6.1.5"><p id="p970175515360"><a name="p970175515360"></a><a name="p970175515360"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row28001952103515"><td class="cellrowborder" valign="top" width="9.08%" headers="mcps1.2.6.1.1 "><p id="p176808377361"><a name="p176808377361"></a><a name="p176808377361"></a>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" width="27.27%" headers="mcps1.2.6.1.2 "><p id="p368103793614"><a name="p368103793614"></a><a name="p368103793614"></a>npu_chip_optical_state</p>
</td>
<td class="cellrowborder" valign="top" width="19.24%" headers="mcps1.2.6.1.3 "><p id="p13681143793613"><a name="p13681143793613"></a><a name="p13681143793613"></a>Optical module presence status</p>
</td>
<td class="cellrowborder" valign="top" width="17.52%" headers="mcps1.2.6.1.4 "><p id="p12681133718365"><a name="p12681133718365"></a><a name="p12681133718365"></a>The value is 0 or 1.</p>
<a name="ul14681837183617"></a><a name="ul14681837183617"></a><ul id="ul14681837183617"><li>0: Not present</li><li>1: Present</li></ul>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="26.889999999999997%" headers="mcps1.2.6.1.5 "><a name="ul1868114372365"></a><a name="ul1868114372365"></a><ul id="ul1868114372365"><li>Atlas Training Series Products</li><li>Atlas A2 Training Series Products</li><li><span id="ph1768263703612"><a name="ph1768263703612"></a><a name="ph1768263703612"></a>Atlas 900 A3 SuperPoD Super Node</span></li><li><span id="ph16373157182715"><a name="ph16373157182715"></a><a name="ph16373157182715"></a>Atlas 800I A2 Inference Server</span></li><li><span id="ph103551958184611"><a name="ph103551958184611"></a><a name="ph103551958184611"></a>A200I A2 Box Heterogeneous Subrack</span></li></ul>
</td>
</tr>
<tr id="row780035203510"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7682143773611"><a name="p7682143773611"></a><a name="p7682143773611"></a>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p668263773619"><a name="p668263773619"></a><a name="p668263773619"></a>npu_chip_optical_tx_power_X (X range: 0~3)</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1682133783619"><a name="p1682133783619"></a><a name="p1682133783619"></a>Optical module TX power</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p968363753614"><a name="p968363753614"></a><a name="p968363753614"></a>Unit: mW</p>
</td>
</tr>
<tr id="row5175131619363"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p19685173716367"><a name="p19685173716367"></a><a name="p19685173716367"></a>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p16851337173614"><a name="p16851337173614"></a><a name="p16851337173614"></a>npu_chip_optical_rx_power_X (X range: 0~3)</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15685173713364"><a name="p15685173713364"></a><a name="p15685173713364"></a>Optical module RX power</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p268533719363"><a name="p268533719363"></a><a name="p268533719363"></a>Unit: mW</p>
</td>
</tr>
<tr id="row16175111616365"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4686837133615"><a name="p4686837133615"></a><a name="p4686837133615"></a>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p136861437133613"><a name="p136861437133613"></a><a name="p136861437133613"></a>npu_chip_optical_vcc</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1868663753610"><a name="p1868663753610"></a><a name="p1868663753610"></a>Optical module voltage</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p166863375361"><a name="p166863375361"></a><a name="p166863375361"></a>Unit: mV</p>
</td>
</tr>
<tr id="row5800105220358"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p2068715375363"><a name="p2068715375363"></a><a name="p2068715375363"></a>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1368793719362"><a name="p1368793719362"></a><a name="p1368793719362"></a>npu_chip_optical_temp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p20687183763617"><a name="p20687183763617"></a><a name="p20687183763617"></a>Optical module temperature</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p106882371361"><a name="p106882371361"></a><a name="p106882371361"></a>Unit: Celsius (℃)</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="8.150815081508151%" headers="mcps1.2.7.1.1 "><p>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" width="22.852285228522852%" headers="mcps1.2.7.1.2 "><p>npu_chip_info_optical_index_num_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="19.401940194019403%" headers="mcps1.2.7.1.3 "><p>Number of optical module lanes connected to the chip Udie Port. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" width="11.96119611961196%" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="22.872287228722872%" headers="mcps1.2.7.1.6 "><p>Atlas 850 Series Hardware Products</p>
</td>
</tr>
<tr id="row184616483311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_optical_tx_power_Z_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Optical module TX power. Z is the lane index with value of [0:3], X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>Unit: mW</p>
</td>
</tr>
<tr id="row1846416482311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>Optical Module</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_optical_rx_power_Z_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Optical module RX power. Z is the lane index with value of [0:3], X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>Unit: mW</p>
</td>
</tr>
</tbody>
</table>

## DDR Data Information<a name="section114607361931169"></a>

**Table 13** DDR data information

<a name="table1251541123212"></a>
<table><thead align="left"><tr id="row152510419324"><th class="cellrowborder" valign="top" width="11.288871112888712%" id="mcps1.2.6.1.1"><p id="p191542026122120"><a name="p191542026122120"></a><a name="p191542026122120"></a>Category</p>
</th>
<th class="cellrowborder" valign="top" width="28.43715628437156%" id="mcps1.2.6.1.2"><p id="p121551826102119"><a name="p121551826102119"></a><a name="p121551826102119"></a>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="24.897510248975102%" id="mcps1.2.6.1.3"><p id="p2015511264214"><a name="p2015511264214"></a><a name="p2015511264214"></a>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="11.668833116688331%" id="mcps1.2.6.1.4"><p id="p515513269215"><a name="p515513269215"></a><a name="p515513269215"></a>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="23.707629237076294%" id="mcps1.2.6.1.5"><p id="p191558266210"><a name="p191558266210"></a><a name="p191558266210"></a>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr id="row7261741103215"><td class="cellrowborder" valign="top" width="11.288871112888712%" headers="mcps1.2.6.1.1 "><p id="p101556265217"><a name="p101556265217"></a><a name="p101556265217"></a>DDR</p>
</td>
<td class="cellrowborder" valign="top" width="28.43715628437156%" headers="mcps1.2.6.1.2 "><p id="p1915532642112"><a name="p1915532642112"></a><a name="p1915532642112"></a>npu_chip_info_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="24.897510248975102%" headers="mcps1.2.6.1.3 "><p id="p2155182642119"><a name="p2155182642119"></a><a name="p2155182642119"></a><span id="ph16848144816402"><a name="ph16848144816402"></a><a name="ph16848144816402"></a>Used DDR memory of the</span> Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" width="11.668833116688331%" headers="mcps1.2.6.1.4 "><p id="p151555262211"><a name="p151555262211"></a><a name="p151555262211"></a>Unit: MB</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="23.707629237076294%" headers="mcps1.2.6.1.5 "><a name="ul12849124816407"></a><a name="ul12849124816407"></a><ul id="ul12849124816407"><li><p id="li16850648204019p0"><a name="li16850648204019p0"></a><a name="li16850648204019p0"></a>Atlas Training Series Products</p>
</li><li><p id="li1785018488403p0"><a name="li1785018488403p0"></a><a name="li1785018488403p0"></a>Inference Server (with Atlas 300I Inference Card)</p>
</li><li><p id="li9850948124020p0"><a name="li9850948124020p0"></a><a name="li9850948124020p0"></a>Atlas Inference Series Products</p>
</li></ul>
</td>
</tr>
<tr id="row20281241173220"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p315632620214"><a name="p315632620214"></a><a name="p315632620214"></a>DDR</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p615682622115"><a name="p615682622115"></a><a name="p615682622115"></a>npu_chip_info_total_memory</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1156626172118"><a name="p1156626172118"></a><a name="p1156626172118"></a><span id="ph8854114810407"><a name="ph8854114810407"></a><a name="ph8854114810407"></a>Total DDR memory of the</span> Ascend AI Processor</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1215682615214"><a name="p1215682615214"></a><a name="p1215682615214"></a>Unit: MB</p>
</td>
</tr>
</tbody>
</table>

## UB Data Information<a name="section998877563214"></a>

**Table 14** UB Data Information

<a name="table998877563214"></a>
<table><thead align="left"><tr><th class="cellrowborder" valign="top" width="8.150815081508151%"><p>Category</p>
</th>
<th class="cellrowborder" valign="top" width="22.852285228522852%"><p>Data Information Name</p>
</th>
<th class="cellrowborder" valign="top" width="19.401940194019403%"><p>Data Information Description</p>
</th>
<th class="cellrowborder" valign="top" width="11.96119611961196%"><p>Unit</p>
</th>
<th class="cellrowborder" valign="top" width="22.872287228722872%"><p>Supported Product Forms</p>
</th>
</tr>
</thead>
<tbody><tr><td class="cellrowborder" valign="top" width="8.150815081508151%" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" width="22.852285228522852%" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv4_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" width="19.401940194019403%" headers="mcps1.2.7.1.3 "><p>Number of IPv4 UB packets received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" width="11.96119611961196%" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
<td class="cellrowborder" rowspan="48" valign="top" width="22.872287228722872%" headers="mcps1.2.7.1.6 "><ul><li>Atlas 350 PCIe Card (4-Processor Mesh)</li><li>Atlas 850 Series Hardware</li><li>Atlas 950 SuperPoD</li></ul>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv6_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of IPv6 UB packets received on the RX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row1846416482311"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv4_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1515522153519"><a name="p1515522153519"></a><a name="p1515522153519"></a>Number of IPv4 UNIC packets received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row20466104816315"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv6_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of IPv6 UNIC packets received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_compact_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of CFG6 packets received on the RX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ctph_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of CFG7 CLAN packets received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ntph_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of CFG7 non-CLAN packets received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_mem_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of UB mem packets received on the RX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unknown_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of unknown packets received on the RX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_drop_ind_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of packets with drop_ind received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_err_ind_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of ERR packets received on the RX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_host_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of received packets on the RX side that are delivered after routing (excluding enumeration configuration and management packets). X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_imp_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of enumeration configuration and management packets received on the RX side that are delivered after routing. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_mar_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of UB memory packets received on the RX side that are delivered after routing. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_link_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of packets received on the RX side that are forwarded to the TX side of the same port after routing. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_to_noc_pkt_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of P2P packets received on the RX side after routing. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_route_err_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of packets received on the RX side that encounter routing table lookup errors after routing. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_out_err_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of bad packets received on the RX side after verification. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_length_err_cnt_rx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>>Number of packets received on the RX side that are identified as length‑error packets after validation. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_rx_busi_flit_num_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of RX packet bytes. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_rx_send_ack_flit_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of packet bytes returned by RX to the peer. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv4_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of IPv4 UB packets sent by the TX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_ipv6_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of IPv6 UB packets sent by the TX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv4_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of IPv4 UNIC packets sent by the TX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unic_ipv6_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of IPv6 UNIC packets sent by the TX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_compact_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of CFG6 packets sent by the TX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ctph_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of CFG7 CLAN packets sent by the TX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_umoc_ntph_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of CFG7 non-CLAN packets sent by the TX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_ub_mem_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of UB mem packets sent by the TX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_unknown_pkt_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of unknown packets sent by the TX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_drop_ind_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of packets with drop_ind sent by the TX side. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_err_ind_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of ERR packets sent by the TX side. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_lpbk_ind_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of packets sent by the TX side that are looped back via NL. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_out_err_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of bad packets sent by the TX side after verification. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_length_err_cnt_tx_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of bad packets sent by the TX side that are identified as length‑error packets after validation. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_tx_busi_flit_num_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of TX packet bytes. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_tx_recv_ack_flit_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of bytes in response packets received by the TX side from the peer. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_retry_req_sum_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of transmission retries. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_retry_ack_sum_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Number of response retries. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_crc_error_sum_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>>Number of CRC validation errors. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxpausepkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of RX pause frames. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txpausepkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of TX pause frames. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxpfcpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of RX PFC frames. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txpfcpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of TX PFC frames. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxbadpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of RX bad packets. X is the Udie ID, and Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txbadpkts_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of TX bad packets. X is the Udie ID, Y is the Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_rxbadoctets_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of RX bad packet bytes. X is Udie ID, Y is Port ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
<tr id="row11470648143118"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>UB</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>npu_chip_info_core_mib_txbadoctets_X_Y</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>Total number of TX bad packet bytes. X is the Udie ID, and Y is the Port ID.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p>-</p>
</td>
</tr>
</tbody>
</table>

## HDK Interface Call<a name="section345820153363"></a>

NPU Exporter obtains corresponding information by calling the underlying HDK interfaces. For more details, refer to HDK Interfaces Called by NPU Exporter.xlsx. To find the HDK interface corresponding to the data information, refer to the following steps.

1. Log in to the [Ascend Computing Documentation](https://support.huawei.com/enterprise/en/category/ascend-computing-pid-1557196528909?submodel=doc) center, select and click the corresponding product name to enter the documentation page. For example, to search for the Atlas 800I A2 Inference Server documentation, click "Atlas 800I A2".
2. Find "Secondary Development" in the left navigation bar, and select the corresponding document based on the interface type.
    - For DCMIs, select "API Reference" and click *[DCMI API Reference](https://support.huawei.com/enterprise/en/ascend-computing/ascend-hdk-pid-252764743?category=developer-documents&subcategory=api-reference)*.
    - For HCCN Tool, select "API Reference" and click *[Atlas A2 Center Inference and Training Hardware HCCN Tool API Reference](https://support.huawei.com/enterprise/en/doc/EDOC1100568362/426cffd9)*.

3. In the search bar on the document homepage, directly search for the corresponding interface name or keyword to obtain relevant information about the interface.
