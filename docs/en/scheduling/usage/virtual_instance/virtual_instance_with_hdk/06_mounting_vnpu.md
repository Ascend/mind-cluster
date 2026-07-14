# Mounting vNPUs

## Mounting vNPUs with Native Docker

In the native Docker scenario (MindCluster cluster scheduling components not deployed), the npu-smi tool is used to create vNPUs and mount the created vNPUs to a container.

## Mounting vNPUs with MindCluster Components

### Method 1: Mounting vNPUs Using Ascend Docker Runtime

This section describes how to use Ascend Docker Runtime (container engine plugin) to mount vNPUs to a container.

**Prerequisites**

Obtain the Ascend-docker-runtime\__\{version\}_\_linux-_\{arch\}_.run package and install it by referring to [Ascend Docker Runtime](../../../installation_guide/03_installation/manual_installation/02_ascend_docker_runtime.md).

**Usage on Ascend Docker Runtime**

Use either of the following methods:

- Static virtualization: After creating a vNPU using the npu-smi tool, run the following command to mount the vNPU to a container when starting the container. The following command means to mount the vNPU whose ID is 100 when the container is started:

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL {image-name:tag} /bin/bash
    ```

- Dynamic virtualization: When starting a container, run the following command to split four AI Cores from the physical device (ID = 0) as vNPUs and mount them to the container. If a container is started in this way, the virtual device is automatically destroyed when the container process is ended.

    ```shell
    docker run -it --rm -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_VNPU_SPECS=vir04 {image-name:tag} /bin/bash
    ```

>[!NOTE] 
>
>- To use dynamic virtualization, disable the vNPU restoration function.
>- You can query the available processor IDs as follows:
>   - Physical processor ID:
>
>      ```shell
>      ls /dev/davinci*
>      ```
>
>   - Virtual processor ID:
>
>     ```shell
>     ls /dev/vdavinci*
>     ```
>
>- image-name:tag: image name and tag, for example, ascend-tensorflow:tensorflow\_TAG.
>- Do not repeatedly define or fix environment variables such as ASCEND\_VISIBLE\_DEVICES, ASCEND\_RUNTIME\_OPTIONS, and ASCEND\_VNPU\_SPECS in the container image.
>- When dynamic virtualization is used, vNPUs cannot be automatically destroyed if a server reboots. You must manually destroy them in this scenario.

**Table 1**  Parameter description

<a name="zh-cn_topic_0000001136053188_table19948947144812"></a>

|Parameter|Description|Example|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|ASCEND_VISIBLE_DEVICES must be used to specify the NPU device to be mounted to the container. Otherwise, the NPU device fails to be mounted. If the NPU device ID is used to specify devices, one or more devices can be specified, and devices can be used together. If the NPU name is used to specify devices, multiple NPU names of the same type can be specified at the same time.|<ul><li>Static virtualization:<ul><li>ASCEND_VISIBLE_DEVICES=100 indicates that vNPU 100 is mounted to the container.</li><li>ASCEND_VISIBLE_DEVICES=101,103 indicates that vNPUs 101 and 103 are mounted to the container.</li><li>ASCEND_VISIBLE_DEVICES=100-102 indicates that vNPUs 100 to 102 (including vNPUs 100 and 102) are mounted to the container. The effect is the same as that of ASCEND_VISIBLE_DEVICES=100,101,102.</li><li>ASCEND_VISIBLE_DEVICES=100-102,104 indicates that vNPUs 100 to 102 and vNPU 104 are mounted to the container. The effect is the same as that of ASCEND_VISIBLE_DEVICES=100,101,102,104.</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y, where XXX represents the NPU device, with supported values being npu, Ascend910, Ascend310, Ascend310B, and Ascend310P; Y represents the physical NPU device ID.<ul><li>ASCEND_VISIBLE_DEVICES=npu-101 indicates that vNPU 101 is mounted to the container.</li><li>ASCEND_VISIBLE_DEVICES=npu-101,npu-103 indicates that NPU 101 and vNPU 103 are mounted to the container.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When specifying devices by chip name, it is recommended to use the value npu uniformly.</li><li>Specifying both a device ID and an NPU name in a single parameter is not supported, meaning ASCEND_VISIBLE_DEVICES=101, npu-103 is not supported.</li><li>It must be used together with ASCEND_RUNTIME_OPTIONS, and the value must contain VIRTUAL, indicating that the vNPU is mounted.</li></ul></div></div></li></ul></li><li>Dynamic virtualization: ASCEND_VISIBLE_DEVICES=0 indicates that a certain number of AI Cores are allocated from NPU device 0.<ul><li>A dynamic virtualization command can specify only one physical NPU ID for dynamic virtualization.</li><li>It must be used together with ASCEND_VNPU_SPECS, to specify the number of AI Cores split on a specified NPU.</li><li>It can be used together with ASCEND_RUNTIME_OPTIONS, but the value can only be NODRV, indicating that the driver-related directory is not mounted.</li></ul></li></ul>|
|ASCEND_RUNTIME_OPTIONS|<p>Restricts the processor ID specified by ASCEND_VISIBLE_DEVICES.</p><ul><li>NODRV indicates that driver-related directories are not mounted.</li><li>VIRTUAL indicates that the virtual processor is mounted.</li><li>NODRV,VIRTUAL indicates that the virtual processor is mounted while driver-related directories are not mounted.</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>In static virtualization scenarios, ASCEND_RUNTIME_OPTIONS is a required parameter, and its value must include VIRTUAL.</li><li>In dynamic virtualization scenarios, if the ASCEND_RUNTIME_OPTIONS parameter is used, its value cannot include VIRTUAL.</li></ul></div></div>|
|ASCEND_VNPU_SPECS|Splits a certain number of AI Cores from a physical NPU device as virtual devices. For supported values, see the "Virtualization Instance Template" column in Table 1 of [Virtualization Templates](./03_virtualization_templates.md).<ul><li>This parameter can only be used for product forms that support dynamic virtualization.</li><li>Must be used together with the "ASCEND_VISIBLE_DEVICES" parameter, which specifies the physical NPU device used for virtualization.</li></ul>|ASCEND_VNPU_SPECS=vir04 indicates that 4 AICores are partitioned as virtual devices and mounted to the container.|

### Method 2: Mounting vNPUs Using Kubernetes

#### vNPU Usage Notes

In Kubernetes scenarios, when you need to use vNPU resources, you must combine Ascend Device Plugin to enable Kubernetes to manage Ascend processor resources. You can use Static virtualization and dynamic virtualization based on whether vNPUs need to be created in advance. These two virtualization modes cannot be used together nor used with the using method on Ascend Docker Runtime mentioned previously. The cluster scheduling components required for the Ascend virtual instance feature are shown in the following table. For supported product models, see "Table 1 Supported products" in [Feature Description](./01_description.md).

**Table 1** Cluster scheduling components required for virtualization

<a name="table19103194217329"></a>
<table><thead align="left"><th class="cellrowborder" valign="top" width="11.677219849801206%" id="mcps1.2.5.1.1"><p id="p2103642143218"><a name="p2103642143218"></a><a name="p2103642143218"></a>Feature</p>
</th>
<th class="cellrowborder" valign="top" width="24.82697688116625%" id="mcps1.2.5.1.2"><p id="p619110456115"><a name="p619110456115"></a><a name="p619110456115"></a>Required Cluster Scheduling Component</p>
</th>
</thead>
<tbody><tr id="row61035425322"><td class="cellrowborder" rowspan="4" valign="top" width="11.677219849801206%" headers="mcps1.2.5.1.1 "><p id="p310384263219"><a name="p310384263219"></a><a name="p310384263219"></a>Static virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p4191645116"><a name="p4191645116"></a><a name="p4191645116"></a><span id="ph1795411794410"><a name="ph1795411794410"></a><a name="ph1795411794410"></a>Ascend Device Plugin</span></p>
</td>
</tr>
<tr id="row1844495022714"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p574771602812"><a name="p574771602812"></a><a name="p574771602812"></a>(Optional) <span id="ph1610211588167">Volcano</span></p>
</td>
</tr>
<tr id="row18230132874912"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11381824102511"><a name="p11381824102511"></a><a name="p11381824102511"></a>(Optional) <span id="ph1566531814589">Ascend Operator</span></p>
</td>
</tr>
<tr><td><p>(Optional) <span>ClusterD</span></p>
</td>
</tr>
<tr id="row610314214324"><td class="cellrowborder" rowspan="4" valign="top" width="11.677219849801206%" headers="mcps1.2.5.1.1 "><p id="p11036426328"><a name="p11036426328"></a><a name="p11036426328"></a>Dynamic Virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p1219211451715"><a name="p1219211451715"></a><a name="p1219211451715"></a><span id="ph12922181924413"><a name="ph12922181924413"></a><a name="ph12922181924413"></a>Ascend Device Plugin</span></p>
</td>
</tr>
<tr><td><p><span>Volcano</span></p>
</td>
</tr>
<tr><td><p>(Optional) <span>Ascend Operator</span></p>
</td>
</tr>
<tr><td><p>(Optional) <span>ClusterD</span></p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>For details about how to install Ascend Device Plugin, see [Ascend Device Plugin](../../../installation_guide/03_installation/manual_installation/04_ascend_device_plugin.md). In the static virtualization scenario, the optionality of components is described as follows.
>
>- Volcano: If you use your own scheduling component, parameter configuration is required. See [Table 2](#table1064314568229). You can also directly use this component for job scheduling.
>- Ascend Operator: This component is required only when training series products are used; it is optional when inference series products are supported.
>- ClusterD: This component is required only when Volcano is used. For details, see [Installing Volcano](../../../installation_guide/03_installation/manual_installation/05_volcano.md).

#### Static Virtualization<a name="ZH-CN_TOPIC_0000002479226392"></a>

**Restrictions<a name="section785220396317"></a>**

- Only single vNPU for single container tasks are supported. Creating copies is not supported.
- Uninstalling Volcano is not supported while a job is running.
- Currently, the rules for the number of NPU devices requested per job Pod are as follows:

    If partitioned vNPUs are used, only 1 is supported.

- In static virtualization scenarios, if you create or destroy a vNPU, you need to restart Ascend Device Plugin.
- Static virtualization tasks do not support fault rescheduling.

**Table 1** Virtual instance templates and vNPU types

<a name="table47415104403"></a>
<table><thead align="left"><tr id="row67416101402"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p117491014400"><a name="p117491014400"></a><a name="p117491014400"></a>NPU Type</p>
</th>
<th class="cellrowborder" valign="top" width="19.96%" id="mcps1.2.5.1.2"><p id="p177431064013"><a name="p177431064013"></a><a name="p177431064013"></a>Virtual Instance Template</p>
</th>
<th class="cellrowborder" valign="top" width="20.04%" id="mcps1.2.5.1.3"><p id="p1374210134015"><a name="p1374210134015"></a><a name="p1374210134015"></a>vNPU Type</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p1041963771317"><a name="p1041963771317"></a><a name="p1041963771317"></a>Virtual Device Name (Example: vNPU 100 and Physical Chip 0)</p>
</th>
</tr>
</thead>
<tbody><tr id="row5741710164014"><td class="cellrowborder" rowspan="4" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p074181014408"><a name="p074181014408"></a><a name="p074181014408"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a>Atlas training series products</span> (30 or 32 AICores)</p>
</td>
<td class="cellrowborder" valign="top" width="19.96%" headers="mcps1.2.5.1.2 "><p id="p974510184017"><a name="p974510184017"></a><a name="p974510184017"></a>vir02</p>
</td>
<td class="cellrowborder" valign="top" width="20.04%" headers="mcps1.2.5.1.3 "><p id="p1575171019404"><a name="p1575171019404"></a><a name="p1575171019404"></a>Ascend910-2c</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1285818202139"><a name="p1285818202139"></a><a name="p1285818202139"></a>Ascend910-2c-100-0</p>
</td>
</tr>
<tr id="row12751210194016"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p177517101404"><a name="p177517101404"></a><a name="p177517101404"></a>vir04</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p47513108403"><a name="p47513108403"></a><a name="p47513108403"></a>Ascend910-4c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p17858172017137"><a name="p17858172017137"></a><a name="p17858172017137"></a>Ascend910-4c-100-0</p>
</td>
</tr>
<tr id="row375141064019"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p197501044011"><a name="p197501044011"></a><a name="p197501044011"></a>vir08</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1275161004018"><a name="p1275161004018"></a><a name="p1275161004018"></a>Ascend910-8c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p168581220181315"><a name="p168581220181315"></a><a name="p168581220181315"></a>Ascend910-8c-100-0</p>
</td>
</tr><tr id="row20758109404"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1375910194012"><a name="p1375910194012"></a><a name="p1375910194012"></a>vir16</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p075131044012"><a name="p075131044012"></a><a name="p075131044012"></a>Ascend910-16c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p188588202135"><a name="p188588202135"></a><a name="p188588202135"></a>Ascend910-16c-100-0</p>
</td>
</tr>
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A2 training series products</span> (24 AICores)</p>
</td>
<td><p>vir12_3c_32g</p>
</td>
<td><p>Ascend910-12c.3cpu.32g</p>
</td>
<td><p>Ascend910-12c.3cpu.32g-100-0</p>
</td>
</tr>
<tr><td><p>vir06_1c_16g</p>
</td>
<td><p>Ascend910-6c.1cpu.16g</p>
</td>
<td><p>Ascend910-6c.1cpu.16g-100-0</p>
</td>
</tr>
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A3 training series products</span> (48 AICores)</p>
</td>
<td><p>vir12_3c_32g</p>
</td>
<td><p>Ascend910-12c.3cpu.32g</p>
</td>
<td><p>Ascend910-12c.3cpu.32g-100-0</p>
</td>
</tr>
<tr><td><p>vir06_1c_16g</p>
</td>
<td><p>Ascend910-6c.1cpu.16g</p>
</td>
<td><p>Ascend910-6c.1cpu.16g-100-0</p>
</td>
</tr>
<tr id="row84911853114212"><td class="cellrowborder" rowspan="7" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas inference series products</span> (8 AICores)</p>
<p id="p12827141603014"><a name="p12827141603014"></a><a name="p12827141603014"></a></p>
</td>
<td class="cellrowborder" valign="top" width="19.96%" headers="mcps1.2.5.1.2 "><p id="p11312190431"><a name="p11312190431"></a><a name="p11312190431"></a>vir01</p>
</td>
<td class="cellrowborder" valign="top" width="20.04%" headers="mcps1.2.5.1.3 "><p id="p9491185334212"><a name="p9491185334212"></a><a name="p9491185334212"></a>Ascend310P-1c</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p785817208133"><a name="p785817208133"></a><a name="p785817208133"></a>Ascend310P-1c-100-0</p>
</td>
</tr>
<tr id="row025285715427"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p42104229438"><a name="p42104229438"></a><a name="p42104229438"></a>vir02</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15252157204214"><a name="p15252157204214"></a><a name="p15252157204214"></a>Ascend310P-2c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p5858122031313"><a name="p5858122031313"></a><a name="p5858122031313"></a>Ascend310P-2c-100-0</p>
</td>
</tr>
<tr id="row97276094310"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p21621623154317"><a name="p21621623154317"></a><a name="p21621623154317"></a>vir04</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7727808436"><a name="p7727808436"></a><a name="p7727808436"></a>Ascend310P-4c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p88588203133"><a name="p88588203133"></a><a name="p88588203133"></a>Ascend310P-4c-100-0</p>
</td>
</tr>
<tr id="row1924012424312"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p864822594315"><a name="p864822594315"></a><a name="p864822594315"></a>vir02_1c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p9240174124315"><a name="p9240174124315"></a><a name="p9240174124315"></a>Ascend310P-2c.1cpu</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p7858122011317"><a name="p7858122011317"></a><a name="p7858122011317"></a>Ascend310P-2c.1cpu-100-0</p>
</td>
</tr>
<tr id="row15871137104318"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17120529164318"><a name="p17120529164318"></a><a name="p17120529164318"></a>vir04_3c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1287219754318"><a name="p1287219754318"></a><a name="p1287219754318"></a>Ascend310P-4c.3cpu</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p2858132091317"><a name="p2858132091317"></a><a name="p2858132091317"></a>Ascend310P-4c.3cpu-100-0</p>
</td>
</tr>
<tr id="row33716311573"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p03711631778"><a name="p03711631778"></a><a name="p03711631778"></a>vir04_3c_ndvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p237116311471"><a name="p237116311471"></a><a name="p237116311471"></a>Ascend310P-4c.3cpu.ndvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p23716311171"><a name="p23716311171"></a><a name="p23716311171"></a>Ascend310P-4c.3cpu.ndvpp-100-0</p>
</td>
</tr>
<tr id="row595773615716"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p119572361679"><a name="p119572361679"></a><a name="p119572361679"></a>vir04_4c_dvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p995718366710"><a name="p995718366710"></a><a name="p995718366710"></a>Ascend310P-4c.4cpu.dvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p9957636276"><a name="p9957636276"></a><a name="p9957636276"></a>Ascend310P-4c.4cpu.dvpp-100-0</p>
</td>
</tr>
<tr><td rowspan="6" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A2 inference series products</span> (20 AICores)</p>
</td>
<td><p>vir10_3c_16g</p>
</td>
<td><p>Ascend910-10c.3cpu.16g</p>
</td>
<td><p>Ascend910-10c.3cpu.16g-100-0</p>
</td>
</tr>
<tr><td><p>vir10_3c_16g_nm</p>
</td>
<td><p>Ascend910-10c.3cpu.16g.ndvpp</p>
</td>
<td><p>Ascend910-10c.3cpu.16g.ndvpp-100-0</p>
</td>
</tr>
<tr><td><p>vir10_4c_16g_m</p>
</td>
<td><p>Ascend910-10c.4cpu.16g.dvpp</p>
</td>
<td><p>Ascend910-10c.4cpu.16g.dvpp-100-0</p>
</td>
</tr>
<tr><td><p>vir05_1c_8g</p>
</td>
<td><p>Ascend910-5c.1cpu.8g</p>
</td>
<td><p>Ascend910-5c.1cpu.8g-100-0</p>
</td>
</tr>
<tr><td><p>vir10_3c_32g</p>
</td>
<td><p>Ascend910-10c.3cpu.32g</p>
</td>
<td><p>Ascend910-10c.3cpu.32g-100-0</p>
</td>
</tr>
<tr><td><p>vir05_1c_16g</p>
</td>
<td><p>Ascend910-5c.1cpu.16g</p>
</td>
<td><p>Ascend910-5c.1cpu.16g-100-0</p>
</td>
</tr>
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A3 inference series products</span> (40 AICores)</p>
</td>
<td><p>vir10_3c_32g</p>
</td>
<td><p>Ascend910-10c.3cpu.32g</p>
</td>
<td><p>Ascend910-10c.3cpu.32g-100-0</p>
</td>
</tr>
<tr><td><p>vir05_1c_16g</p>
</td>
<td><p>Ascend910-5c.1cpu.16g</p>
</td>
<td><p>Ascend910-5c.1cpu.16g-100-0</p>
</td>
</tr>
</tbody>
</table>

**Prerequisites<a name="section18128140645"></a>**

1. You need to first obtain `Ascend-docker-runtime_{version}_linux-{arch}.run` and install the container engine plugin.
2. See the [Installation and Deployment](../../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) section to complete the installation of each component.

    The cluster scheduling components involved in modifying relevant parameters for virtual instances are Volcano and Ascend Device Plugin. Please modify and use the corresponding YAML for installation and deployment as required below:

    - Affinity scenario: Volcano must be installed.
    - Non-affinity scenario: Volcano does not need to be installed; only the device count is reported to the node's K8s.

    1. Ascend Device Plugin parameter modification and startup instructions:

        Virtual instance startup parameter description:

        **Table 2** Ascend Device Plugin startup parameters

        <a name="table1064314568229"></a>

        |Name|Type|Default Value|Description|
        |--|--|--|--|
        |-volcanoType|bool|false|Whether to use Volcano for scheduling. If dynamic virtualization is used, set this to true.|
        |-presetVirtualDevice|bool|true|Static virtualization feature switch. The value can only be true.<p>If dynamic virtualization is used, set this to false and enable Volcano synchronously, that is, set the "-volcanoType" parameter to true.</p>|

        The YAML startup instructions are as follows:

        - Atlas inference series products used in the K8s cluster (Ascend Device Plugin works independently, without using Volcano)

            ```shell
            kubectl apply -f device-plugin-310P-v{version}.yaml
            ```

        - Atlas training series products, Atlas A2 training series products, Atlas A3 training series products, Atlas A2 inference series products, and Atlas A3 inference series products used in the K8s cluster (Ascend Device Plugin works independently, without cooperating with Volcano and Ascend Operator)

            ```shell
            kubectl apply -f device-plugin-910-v{version}.yaml
            ```

        - Atlas inference series products used in the K8s cluster (Volcano used, supporting NPU virtualization, with dynamic virtualization disabled by default in the YAML)

            ```shell
            kubectl apply -f device-plugin-310P-volcano-v{version}.yaml
            ```

        - Atlas training series products, Atlas A2 training series products, Atlas A3 training series products, Atlas A2 inference series products, and Atlas A3 inference series products used in the K8s cluster (cooperating with Volcano and Ascend Operator, supporting NPU virtualization, with dynamic virtualization disabled by default in the YAML).

            ```shell
            kubectl apply -f device-plugin-volcano-v{version}.yaml
            ```

       If the K8s cluster uses multiple types of Ascend AI Processors, run the corresponding commands separately.

    2. Volcano parameter modification and startup instructions:

       In the Volcano deployment file `volcano-v{version}.yaml`", you need to configure `presetVirtualDevice` and the value can only be `true`.

        ```Yaml
        ...
        data:
          volcano-scheduler.conf: |
            actions: "enqueue, allocate, backfill"
            tiers:
            - plugins:
              - name: priority
              - name: gang
              - name: conformance
              - name: volcano-npu-v26.0.0_linux-aarch64    # Where 26.0.0 is the version number of MindCluster. The value varies depending on the version.
            - plugins:
              - name: drf
              - name: predicates
              - name: proportion
              - name: nodeorder
              - name: binpack
            configurations:
             ...
              - name: init-params
                arguments: {"grace-over-time":"900","presetVirtualDevice":"true"}
        ...
        ```

**Usage<a name="section514441719341"></a>**

- When creating a training job, modify the following configuration in the YAML file. This example uses Atlas training series products.

    The resource type specified in `requests` and `limits` under `resources` should be changed to `huawei.com/Ascend910-Y`, where the <i>Y</i> value is related to the vNPU type. For specific values, refer to the "vNPU Type" column in [Table 2 Virtual instance templates and vNPU types](#table47415104403).

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend910-Y: 1          # Number of vNPUs requested. The maximum value is 1.
                limits:
                  huawei.com/Ascend910-Y: 1          # The value must match the requested quantity.
    ...
    ```

- When creating an inference job, modify the following configuration in the YAML file. This example uses Atlas inference series products.

  The resource type specified in `requests` and `limits` under `resources` should be changed to `huawei.com/Ascend310P-Y`, where the <i>Y</i> value is related to the vNPU type. For specific values, refer to the "vNPU Type" column in [Table 2 Virtual instance templates and vNPU types](#table47415104403).

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend310P-Y: 1          # Number of vNPUs requested. The maximum value is 1.
                limits:
                  huawei.com/Ascend310P-Y: 1          # The value must match the requested quantity.
    ...
    ```

#### Dynamic Virtualization

Before using dynamic virtualization, read [Table1 Scenario description](#table625511844619).

**Usage Notes<a name="section1576110260450"></a>**

**Table 1** Scenario description

<a name="table625511844619"></a>
<table><thead align="left"><tr id="row9255148204610"><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p id="p4381442125317"><a name="p4381442125317"></a><a name="p4381442125317"></a>Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="80.02%" id="mcps1.2.3.1.2"><p id="p2255984464"><a name="p2255984464"></a><a name="p2255984464"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row132012115910"><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p1950512911598"><a name="p1950512911598"></a><a name="p1950512911598"></a>General Description</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p450516910592"><a name="p450516910592"></a><a name="p450516910592"></a>The allocated chip information is reflected in the Pod's annotation. For detailed description of Pod annotation, see the huawei.com/npu-core and huawei.com/AscendReal parameters in <a href="../../../api/k8s.md">Pod annotation</a>.</p>
</td>
</tr>
<tr id="row48061646595"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1749665239"><a name="p1749665239"></a><a name="p1749665239"></a>At any given time, only jobs with the same <a href="./03_virtualization_templates.md">virtualization template</a> can be submitted.</p>
</td>
</tr>
<tr id="row18542176195917"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p450559185914"><a name="p450559185914"></a><a name="p450559185914"></a>When dynamically allocating vNPUs, <span id="ph19255162231216"><a name="ph19255162231216"></a><a name="ph19255162231216"></a>MindCluster</span> scheduling will prioritize occupying the physical NPU with the least remaining computing power.</p>
</td>
</tr>
<tr id="row11648825917"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p19505796596"><a name="p19505796596"></a><a name="p19505796596"></a>Currently, the number of NPUs requested by each Pod of a job is 1.</p>
</td>
</tr>
<tr id="row32567817461"><td class="cellrowborder" rowspan="3" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p1325613818460"><a name="p1325613818460"></a><a name="p1325613818460"></a>Scenarios supported</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p32561983469"><a name="p32561983469"></a><a name="p32561983469"></a>Multiple replicas are supported, but each Pod in the replicas must use vNPU.</p>
</td>
</tr>
<tr id="row5256198134612"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p72561586465"><a name="p72561586465"></a><a name="p72561586465"></a>The Kubernetes mechanism is supported, such as affinity.</p>
</td>
</tr>
<tr id="row825611817468"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p2795151384913"><a name="p2795151384913"></a><a name="p2795151384913"></a>Rescheduling upon chip faults and node faults is supported. For details, see <span id="ph1389215534914"><a name="ph1389215534914"></a><a name="ph1389215534914"></a><a href="../../basic_scheduling/10_recovery_of_inference_card_faults.md">Recovery of Inference Card Faults</a></span> and <a href="../../basic_scheduling/09_rescheduling_upon_inference_card_faults.md">Rescheduling upon Inference Card Faults</a>.</p>
</td>
</tr>
<tr id="row237762345420"><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p840574125511"><a name="p840574125511"></a><a name="p840574125511"></a>Scenarios not supported</p>
<p id="p17835104672517"><a name="p17835104672517"></a><a name="p17835104672517"></a></p>
<p id="p36763525314"><a name="p36763525314"></a><a name="p36763525314"></a></p>
<p id="p767616565314"><a name="p767616565314"></a><a name="p767616565314"></a></p>
<p id="p667616595317"><a name="p667616595317"></a><a name="p667616595317"></a></p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p14377152385414"><a name="p14377152385414"></a><a name="p14377152385414"></a>Mixing different chips within a single job is not supported.</p>
</td>
</tr>
<tr id="row1625614818462"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p32566874611"><a name="p32566874611"></a><a name="p32566874611"></a>Uninstalling <span id="ph42462611516"><a name="ph42462611516"></a><a name="ph42462611516"></a>Volcano</span> during job execution is not supported.</p>
</td>
</tr>
<tr id="row1854910515540"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p12256108124616"><a name="p12256108124616"></a><a name="p12256108124616"></a>In K8s scenarios, vNPUs are automatically created and destroyed. Do not mix operations with Docker scenarios.</p>
</td>
</tr>
<tr id="row151011624135113"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18102182414515"><a name="p18102182414515"></a><a name="p18102182414515"></a>For nodes undergoing dynamic virtualization, you cannot configure the chip CPU. </p>
</td>
</tr>
<tr id="row192561854613"><td class="cellrowborder" rowspan="3" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p1125610854611"><a name="p1125610854611"></a><a name="p1125610854611"></a><span id="ph10445185418466"><a name="ph10445185418466"></a><a name="ph10445185418466"></a>Atlas inference product</span> (8 AI Cores)</p>
<p id="p1173133213564"><a name="p1173133213564"></a><a name="p1173133213564"></a></p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p02561481463"><a name="p02561481463"></a><a name="p02561481463"></a>When vNPUs are used, the number of AI Cores that can be requested by a job is 1, 2, or 4. When the physical NPU is used, that number must be 8 or a multiple of 8.</p>
</td>
</tr>
<tr id="row11782173617479"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18782936144718"><a name="p18782936144718"></a><a name="p18782936144718"></a>The container is started by the root user. If you need to run an inference job as a common user, refer to <a href="https://gitcode.com/Ascend/mind-cluster/issues/359">Inference service fails when running as a regular user with dynamic virtualization</a>.</p>
</td>
</tr>
<tr id="row117233216566"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18081933105617"><a name="p18081933105617"></a><a name="p18081933105617"></a>Dynamic vNPU creation and destruction are valid only on Atlas inference product and must be used with Volcano.</p>
</td>
</tr>
</tbody>
</table>

**Table 2** Relationship between virtual instance templates and vNPU types

<a name="table47415104403"></a>
<table><thead align="left"><tr id="row67416101402"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p117491014400"><a name="p117491014400"></a><a name="p117491014400"></a>NPU Type</p>
</th>
<th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.5.1.2"><p id="p177431064013"><a name="p177431064013"></a><a name="p177431064013"></a>Virtual Instance Template</p>
</th>
<th class="cellrowborder" valign="top" width="20.02%" id="mcps1.2.5.1.3"><p id="p1374210134015"><a name="p1374210134015"></a><a name="p1374210134015"></a>vNPU Type</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p1041963771317"><a name="p1041963771317"></a><a name="p1041963771317"></a>Specific virtual device name (using vNPU ID 100 and physical chip ID 0 as an example)</p>
</th>
</tr>
</thead>
<tbody><tr id="row84911853114212"><td class="cellrowborder" rowspan="7" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph1534112451967"><a name="ph1534112451967"></a><a name="ph1534112451967"></a>Atlas inference series products</span> (8 AICores)</p>
</td>
<td class="cellrowborder" valign="top" width="19.98%" headers="mcps1.2.5.1.2 "><p id="p11312190431"><a name="p11312190431"></a><a name="p11312190431"></a>vir01</p>
</td>
<td class="cellrowborder" valign="top" width="20.02%" headers="mcps1.2.5.1.3 "><p id="p9491185334212"><a name="p9491185334212"></a><a name="p9491185334212"></a>Ascend310P-1c</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p785817208133"><a name="p785817208133"></a><a name="p785817208133"></a>Ascend310P-1c-100-0</p>
</td>
</tr>
<tr id="row025285715427"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p42104229438"><a name="p42104229438"></a><a name="p42104229438"></a>vir02</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15252157204214"><a name="p15252157204214"></a><a name="p15252157204214"></a>Ascend310P-2c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p5858122031313"><a name="p5858122031313"></a><a name="p5858122031313"></a>Ascend310P-2c-100-0</p>
</td>
</tr>
<tr id="row97276094310"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p21621623154317"><a name="p21621623154317"></a><a name="p21621623154317"></a>vir04</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p7727808436"><a name="p7727808436"></a><a name="p7727808436"></a>Ascend310P-4c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p88588203133"><a name="p88588203133"></a><a name="p88588203133"></a>Ascend310P-4c-100-0</p>
</td>
</tr>
<tr id="row1924012424312"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p864822594315"><a name="p864822594315"></a><a name="p864822594315"></a>vir02_1c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p9240174124315"><a name="p9240174124315"></a><a name="p9240174124315"></a>Ascend310P-2c.1cpu</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p7858122011317"><a name="p7858122011317"></a><a name="p7858122011317"></a>Ascend310P-2c.1cpu-100-0</p>
</td>
</tr>
<tr id="row15871137104318"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17120529164318"><a name="p17120529164318"></a><a name="p17120529164318"></a>vir04_3c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1287219754318"><a name="p1287219754318"></a><a name="p1287219754318"></a>Ascend310P-4c.3cpu</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p2858132091317"><a name="p2858132091317"></a><a name="p2858132091317"></a>Ascend310P-4c.3cpu-100-0</p>
</td>
</tr>
<tr id="row33716311573"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p03711631778"><a name="p03711631778"></a><a name="p03711631778"></a>vir04_3c_ndvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p237116311471"><a name="p237116311471"></a><a name="p237116311471"></a>Ascend310P-4c.3cpu.ndvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p23716311171"><a name="p23716311171"></a><a name="p23716311171"></a>Ascend310P-4c.3cpu.ndvpp-100-0</p>
</td>
</tr>
<tr id="row595773615716"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p119572361679"><a name="p119572361679"></a><a name="p119572361679"></a>vir04_4c_dvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p995718366710"><a name="p995718366710"></a><a name="p995718366710"></a>Ascend310P-4c.4cpu.dvpp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p9957636276"><a name="p9957636276"></a><a name="p9957636276"></a>Ascend310P-4c.4cpu.dvpp-100-0</p>
</td>
</tr>
</tbody>
</table>

**Prerequisites**

1. You need to first obtain `Ascend-docker-runtime_{version}_linux-{arch}.run` and install the container engine plugin.
2. Refer to the [Installation and Deployment](../../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) chapter to complete the installation of each component.

   The cluster scheduling components involved for the virtual instance configuration are Volcano and Ascend Device Plugin. Please modify and use the corresponding YAML for installation and deployment as required:

   1. Ascend Device Plugin parameter modification and startup instructions

      The startup parameters for the virtual instance are described as follows:

      **Table 3** Ascend Device Plugin startup parameters

      <a name="table1064314568229"></a>

      |Name|Type|Default Value|Description|
      |--|--|--|--|
      |-volcanoType|bool|false|Whether to use Volcano for scheduling. If dynamic virtualization is used, this must be set to true.|
      |-presetVirtualDevice|bool|true|Static virtualization feature switch.<p>If dynamic virtualization is used, this must be set to false, and Volcano must be enabled simultaneously, i.e., set the "-volcanoType" parameter to true.</p>|

      The YAML startup instructions are as follows:

      In a K8s cluster with nodes using Atlas inference series products, the `"presetVirtualDevice"` field must be modified to `"false"` in `device-plugin-310P-volcano-v{version}` (used in conjunction with Volcano to support NPU virtualization; Dynamic virtualization is disabled by default in the YAML).

       ```Yaml
       ...
       args: [ "device-plugin  -useAscendDocker=true -volcanoType=true -presetVirtualDevice=false
                  -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=0" ]
       ...
       ```

   2. Volcano parameter modification and startup instructions

      In the Volcano deployment file "`volcano-v{version}.yaml`", you need to configure the value of "`presetVirtualDevice`" to "`false`".

       ```Yaml
       ...
       data:
         volcano-scheduler.conf: |
           actions: "enqueue, allocate, backfill"
           tiers:
           - plugins:
             - name: priority
             - name: gang
             - name: conformance
             - name: volcano-npu-v{version}_linux-aarch64
           - plugins:
             - name: drf
             - name: predicates
             - name: proportion
             - name: nodeorder
             - name: binpack
           configurations:
            ...
             - name: init-params
               arguments: {"grace-over-time":"900","presetVirtualDevice":"false"}  # Enable dynamic virtualization, the value of presetVirtualDevice needs to be set to false
       ...
       ```

**Instructions**

Modify the following configuration when creating a YAML file upon inference job creation (Atlas inference product as an example).

To allocate an AI Core, the requests and limits types set in resources need to be changed to huawei.com/npu-core. The following uses Deployment as an example:

```Yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy-with-volcano
  labels:
    app: tf
  namespace: vnpu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tf
  template:
    metadata:
      labels:
        app: tf
        ring-controller.atlas: ascend-310P  # See Table 4
        fault-scheduling: "grace"           # Label used for rescheduling.
        vnpu-dvpp: "yes"                    # See Table 4
        vnpu-level: "low"                   # See Table 4
    spec:
      schedulerName: volcano  # MindCluster Volcano is required.
      nodeSelector:
        host-arch: huawei-arm
      containers:
        - image: ubuntu:22.04   # Example image
          imagePullPolicy: IfNotPresent
          name: tf
          command:
          - "/bin/bash"
          - "-c"
          args: ["Customer's own running script"]
          resources:
            requests:
              huawei.com/npu-core: 1        # Use the vir01 template to dynamically virtualize NPUs.
            limits:
              huawei.com/npu-core: 1        # The value is the same as that in requests.
 ....
```

**Table 4**  Virtual instance labels in the YAML file

<a name="table1084325844716"></a>
<table><thead align="left"><tr id="row13843105815479"><th class="cellrowborder" valign="top" width="17.88178817881788%" id="mcps1.2.4.1.1"><p id="p1879944394819"><a name="p1879944394819"></a><a name="p1879944394819"></a>key</p>
</th>
<th class="cellrowborder" valign="top" width="31.053105310531055%" id="mcps1.2.4.1.2"><p id="p6307191712494"><a name="p6307191712494"></a><a name="p6307191712494"></a>value</p>
</th>
<th class="cellrowborder" valign="top" width="51.06510651065107%" id="mcps1.2.4.1.3"><p id="p571812231496"><a name="p571812231496"></a><a name="p571812231496"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row11843135814719"><td class="cellrowborder" rowspan="2" valign="top" width="17.88178817881788%" headers="mcps1.2.4.1.1 "><p id="p11799943114814"><a name="p11799943114814"></a><a name="p11799943114814"></a>vnpu-level</p>
<p id="p127511550154811"><a name="p127511550154811"></a><a name="p127511550154811"></a></p>
</td>
<td class="cellrowborder" valign="top" width="31.053105310531055%" headers="mcps1.2.4.1.2 "><p id="p73071317144911"><a name="p73071317144911"></a><a name="p73071317144911"></a>low</p>
</td>
<td class="cellrowborder" valign="top" width="51.06510651065107%" headers="mcps1.2.4.1.3 "><p id="p20719152316493"><a name="p20719152316493"></a><a name="p20719152316493"></a>Low configuration. This is the default value. Select the virtual instance template with the minimum configuration.</p>
</td>
</tr>
<tr id="row1475114503484"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p12307151724910"><a name="p12307151724910"></a><a name="p12307151724910"></a>high</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1271902314916"><a name="p1271902314916"></a><a name="p1271902314916"></a>Performance comes in the first place.</p>
<p id="p071922312490"><a name="p071922312490"></a><a name="p071922312490"></a>If there are enough cluster resources, select a virtual instance template with the highest configuration. If most of the cluster resources are used, for example, most physical NPUs are used and only a small number of AI Cores are left on each physical NPU, other templates with lower configurations with the same number of AI Cores are used. For details, see <a href="#table83781115185619">Table 5</a>.</p>
</td>
</tr>
<tr id="row8843145854711"><td class="cellrowborder" rowspan="3" valign="top" width="17.88178817881788%" headers="mcps1.2.4.1.1 "><p id="p168872618492"><a name="p168872618492"></a><a name="p168872618492"></a>vnpu-dvpp</p>
</td>
<td class="cellrowborder" valign="top" width="31.053105310531055%" headers="mcps1.2.4.1.2 "><p id="p2030751719499"><a name="p2030751719499"></a><a name="p2030751719499"></a>yes</p>
</td>
<td class="cellrowborder" valign="top" width="51.06510651065107%" headers="mcps1.2.4.1.3 "><p id="p971972316498"><a name="p971972316498"></a><a name="p971972316498"></a>This pod uses DVPP.</p>
</td>
</tr>
<tr id="row165811357114820"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1630820172490"><a name="p1630820172490"></a><a name="p1630820172490"></a>no</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5719152304920"><a name="p5719152304920"></a><a name="p5719152304920"></a>This pod does not use DVPP.</p>
</td>
</tr>
<tr id="row173650119495"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p12308131744912"><a name="p12308131744912"></a><a name="p12308131744912"></a>null</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1871982374915"><a name="p1871982374915"></a><a name="p1871982374915"></a>This is the default value. Whether the DVPP is used is not concerned.</p>
</td>
</tr>
<tr id="row184385814710"><td class="cellrowborder" valign="top" width="17.88178817881788%" headers="mcps1.2.4.1.1 "><p id="p1680094354812"><a name="p1680094354812"></a><a name="p1680094354812"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="31.053105310531055%" headers="mcps1.2.4.1.2 "><p id="p530851794913"><a name="p530851794913"></a><a name="p530851794913"></a>ascend-310P</p>
</td>
<td class="cellrowborder" valign="top" width="51.06510651065107%" headers="mcps1.2.4.1.3 "><p id="p1871918233494"><a name="p1871918233494"></a><a name="p1871918233494"></a>Flag indicates that Atlas inference product is used.</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 
>Selection result of vnpu-level and vnpu-dvpp. For details, see [Table 5](#table83781115185619).
>
>- Degrade in the table indicates that when the number of AI Cores meets the requirement, but other resources (such as the AI CPUs) are insufficient, another template that has the same number of AI Cores but different other resources will be selected. If only one processor is left with two AI Cores and one AI CPU, the vir02 template is degraded to vir02\_1c.
>- The values listed under Template correspond to those listed under Virtual Instance Template of Atlas inference product, in [Virtualization Templates](./03_virtualization_templates.md).
>- In the <b>vnpu-level</b> column of the table, Other indicate any value except <b>low</b> and <b>high</b>.
>- If an entire processor (with 8 cores or a multiple of 8 cores) will be used, vnpu-dvpp and vnpu-level can be set to any value.

**Table 5**  DVPP and levels

<a name="table83781115185619"></a>
<table><thead align="left"><tr id="row1837817157565"><th class="cellrowborder" valign="top" width="17.2982701729827%" id="mcps1.2.7.1.1"><p id="p11560216112"><a name="p11560216112"></a><a name="p11560216112"></a>Product Model</p>
</th>
<th class="cellrowborder" valign="top" width="16.42835716428357%" id="mcps1.2.7.1.2"><p id="p1024717408463"><a name="p1024717408463"></a><a name="p1024717408463"></a>Number of Requested AI Cores</p>
</th>
<th class="cellrowborder" valign="top" width="15.768423157684234%" id="mcps1.2.7.1.3"><p id="p192479402463"><a name="p192479402463"></a><a name="p192479402463"></a>vnpu-dvpp</p>
</th>
<th class="cellrowborder" valign="top" width="20.987901209879013%" id="mcps1.2.7.1.4"><p id="p1024716402460"><a name="p1024716402460"></a><a name="p1024716402460"></a>vnpu-level</p>
</th>
<th class="cellrowborder" valign="top" width="8.52914708529147%" id="mcps1.2.7.1.5"><p id="p8247440174613"><a name="p8247440174613"></a><a name="p8247440174613"></a>Degrade (Y/N)</p>
</th>
<th class="cellrowborder" valign="top" width="20.987901209879013%" id="mcps1.2.7.1.6"><p id="p0247164034611"><a name="p0247164034611"></a><a name="p0247164034611"></a>Template</p>
</th>
</tr>
</thead>
<tbody><tr id="row1517703912018"><td class="cellrowborder" rowspan="12" valign="top" width="17.2982701729827%" headers="mcps1.2.7.1.1 "><p id="p8916171416125"><a name="p8916171416125"></a><a name="p8916171416125"></a>Atlas inference product (8 AI Cores)</p>
<p id="p317720394019"><a name="p317720394019"></a><a name="p317720394019"></a></p>
<p id="p717811391508"><a name="p717811391508"></a><a name="p717811391508"></a></p>
<p id="p16324345105912"><a name="p16324345105912"></a><a name="p16324345105912"></a></p>
<p id="p5934321617"><a name="p5934321617"></a><a name="p5934321617"></a></p>
<p id="p209341921210"><a name="p209341921210"></a><a name="p209341921210"></a></p>
<p id="p59341821618"><a name="p59341821618"></a><a name="p59341821618"></a></p>
<p id="p9797183210114"><a name="p9797183210114"></a><a name="p9797183210114"></a></p>
<p id="p19813153915118"><a name="p19813153915118"></a><a name="p19813153915118"></a></p>
<p id="p1481383919117"><a name="p1481383919117"></a><a name="p1481383919117"></a></p>
</td>
<td class="cellrowborder" valign="top" width="16.42835716428357%" headers="mcps1.2.7.1.2 "><p id="p191771939903"><a name="p191771939903"></a><a name="p191771939903"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="15.768423157684234%" headers="mcps1.2.7.1.3 "><p id="p14248174010469"><a name="p14248174010469"></a><a name="p14248174010469"></a>null</p>
</td>
<td class="cellrowborder" valign="top" width="20.987901209879013%" headers="mcps1.2.7.1.4 "><p id="p1385717396538"><a name="p1385717396538"></a><a name="p1385717396538"></a>Any value</p>
</td>
<td class="cellrowborder" valign="top" width="8.52914708529147%" headers="mcps1.2.7.1.5 "><p id="p38575391531"><a name="p38575391531"></a><a name="p38575391531"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="20.987901209879013%" headers="mcps1.2.7.1.6 "><p id="p385603935319"><a name="p385603935319"></a><a name="p385603935319"></a>vir01</p>
</td>
</tr>
<tr id="row11177839600"><td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.1 "><p id="p1317733915013"><a name="p1317733915013"></a><a name="p1317733915013"></a>2</p>
<p id="p8178439503"><a name="p8178439503"></a><a name="p8178439503"></a></p>
<p id="p1732216453596"><a name="p1732216453596"></a><a name="p1732216453596"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.2 "><p id="p1248174014614"><a name="p1248174014614"></a><a name="p1248174014614"></a>null</p>
<p id="p13302164084616"><a name="p13302164084616"></a><a name="p13302164084616"></a></p>
<p id="p1448013112212"><a name="p1448013112212"></a><a name="p1448013112212"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p14619832145315"><a name="p14619832145315"></a><a name="p14619832145315"></a>low/other</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p126198326538"><a name="p126198326538"></a><a name="p126198326538"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p3248164094613"><a name="p3248164094613"></a><a name="p3248164094613"></a>vir02_1c</p>
</td>
</tr>
<tr id="row117818394016"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.7.1.1 "><p id="p162489402463"><a name="p162489402463"></a><a name="p162489402463"></a>high</p>
<p id="p143218450593"><a name="p143218450593"></a><a name="p143218450593"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p22482040124615"><a name="p22482040124615"></a><a name="p22482040124615"></a>no</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p182481740174611"><a name="p182481740174611"></a><a name="p182481740174611"></a>vir02</p>
</td>
</tr>
<tr id="row16943192222113"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1324834017468"><a name="p1324834017468"></a><a name="p1324834017468"></a>yes</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p16248840154619"><a name="p16248840154619"></a><a name="p16248840154619"></a>vir02_1c</p>
</td>
</tr>
<tr id="row15502725152112"><td class="cellrowborder" rowspan="7" valign="top" headers="mcps1.2.7.1.1 "><p id="p1531894575910"><a name="p1531894575910"></a><a name="p1531894575910"></a>4</p>
<p id="p231434585920"><a name="p231434585920"></a><a name="p231434585920"></a></p>
<p id="p793462111111"><a name="p793462111111"></a><a name="p793462111111"></a></p>
<p id="p1793418218114"><a name="p1793418218114"></a><a name="p1793418218114"></a></p>
<p id="p16934112119119"><a name="p16934112119119"></a><a name="p16934112119119"></a></p>
<p id="p1879713323111"><a name="p1879713323111"></a><a name="p1879713323111"></a></p>
<p id="p68138391419"><a name="p68138391419"></a><a name="p68138391419"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p10248164012460"><a name="p10248164012460"></a><a name="p10248164012460"></a>yes</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.3 "><p id="p3248184024610"><a name="p3248184024610"></a><a name="p3248184024610"></a>low/other</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.4 "><p id="p4249114074618"><a name="p4249114074618"></a><a name="p4249114074618"></a>-</p>
<p id="p1631211451596"><a name="p1631211451596"></a><a name="p1631211451596"></a></p>
<p id="p189347217116"><a name="p189347217116"></a><a name="p189347217116"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p8249540164619"><a name="p8249540164619"></a><a name="p8249540164619"></a>vir04_4c_dvpp</p>
</td>
</tr>
<tr id="row1631142722119"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p192491540164619"><a name="p192491540164619"></a><a name="p192491540164619"></a>no</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p5249124011467"><a name="p5249124011467"></a><a name="p5249124011467"></a>vir04_3c_ndvpp</p>
</td>
</tr>
<tr id="row493411217111"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p424914004612"><a name="p424914004612"></a><a name="p424914004612"></a>null</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p192493409466"><a name="p192493409466"></a><a name="p192493409466"></a>vir04_3c</p>
</td>
</tr>
<tr id="row139342211813"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p924924018462"><a name="p924924018462"></a><a name="p924924018462"></a>yes</p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" headers="mcps1.2.7.1.2 "><p id="p2249440184619"><a name="p2249440184619"></a><a name="p2249440184619"></a>high</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.7.1.3 "><p id="p14272035114811"><a name="p14272035114811"></a><a name="p14272035114811"></a>-</p>
<p id="p021482217814"><a name="p021482217814"></a><a name="p021482217814"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1324984017461"><a name="p1324984017461"></a><a name="p1324984017461"></a>vir04_4c_dvpp</p>
</td>
</tr>
<tr id="row1993412116119"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p824916403462"><a name="p824916403462"></a><a name="p824916403462"></a>no</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p15249440164616"><a name="p15249440164616"></a><a name="p15249440164616"></a>vir04_3c_ndvpp</p>
</td>
</tr>
<tr id="row2797113219118"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.7.1.1 "><p id="p1824974014620"><a name="p1824974014620"></a><a name="p1824974014620"></a>null</p>
<p id="p1681315391419"><a name="p1681315391419"></a><a name="p1681315391419"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p10249124011467"><a name="p10249124011467"></a><a name="p10249124011467"></a>no</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p324964074618"><a name="p324964074618"></a><a name="p324964074618"></a>vir04</p>
</td>
</tr>
<tr id="row16813143918117"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2249340144615"><a name="p2249340144615"></a><a name="p2249340144615"></a>yes</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p924924064613"><a name="p924924064613"></a><a name="p924924064613"></a>vir04_3c</p>
</td>
</tr>
<tr id="row1781312397116"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p102497405465"><a name="p102497405465"></a><a name="p102497405465"></a>8 or a multiple of 8</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p42491440174615"><a name="p42491440174615"></a><a name="p42491440174615"></a>Any value</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p5249114074614"><a name="p5249114074614"></a><a name="p5249114074614"></a>Any value</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1224920403467"><a name="p1224920403467"></a><a name="p1224920403467"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p55031522345"><a name="p55031522345"></a><a name="p55031522345"></a>-</p>
</td>
</tr>
<tr id="row74471126913"><td class="cellrowborder" colspan="6" valign="top" headers="mcps1.2.7.1.1 mcps1.2.7.1.2 mcps1.2.7.1.3 mcps1.2.7.1.4 mcps1.2.7.1.5 mcps1.2.7.1.6 "><p id="p627014191100"><a name="p627014191100"></a><a name="p627014191100"></a>Notes:</p>
<p id="p9942971914"><a name="p9942971914"></a><a name="p9942971914"></a>For Atlas inference product (with eight AI Cores), the number of AI Cores to be allocated must be 8 or a multiple of 8.</p>
</td>
</tr>
</tbody>
</table>

>[!NOTICE] 
>In the preceding table, for vNPUs, the value of vnpu-dvpp need to be consistent with that listed in the table. Otherwise, jobs cannot be delivered.
