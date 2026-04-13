# 挂载vNPU<a name="ZH-CN_TOPIC_0000002479386388"></a>

## 基于原生Docker挂载vNPU<a name="ZH-CN_TOPIC_0000002479226416"></a>

原生Docker场景下（未部署MindCluster集群调度组件），需要使用npu-smi工具创建vNPU后，将vNPU挂载到容器。具体操作请参见《Atlas 中心训练服务器 25.5.0 NPU驱动和固件安装指南》的“昇腾虚拟化实例（AVI）容器场景下的安装与卸载\>[多容器场景下安装](https://support.huawei.com/enterprise/zh/doc/EDOC1100540363/5b32515a)”章节，该章节指导用户安装Docker和将vNPU挂载进容器。

## 基于MindCluster组件挂载vNPU<a name="ZH-CN_TOPIC_0000002511346329"></a>

### 方式一：Ascend Docker Runtime挂载vNPU<a name="ZH-CN_TOPIC_0000002479386376"></a>

单独结合Ascend Docker Runtime（容器引擎插件）使用，将vNPU挂载到容器。

**使用前提<a name="section18128140645"></a>**

需要先获取Ascend-docker-runtime\__\{version\}_\_linux-_\{arch\}_.run，并安装容器引擎插件，方法可参见[Ascend Docker Runtime](../../../installation_guide/03_installation.md#ascend-docker-runtime)。

**Ascend Docker Runtime使用vNPU方法<a name="section514441719341"></a>**

选择以下两种方式之一进行使用：

- 静态虚拟化：用户已通过npu-smi工具创建vNPU，在拉起容器时执行以下命令将vNPU挂载至容器中。以下命令表示用户在拉起容器时，挂载虚拟芯片ID为100的芯片。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL image-name:tag /bin/bash
    ```

- 动态虚拟化：用户在拉起容器时，执行以下命令虚拟化资源，以下命令表示从物理芯片ID为0的芯片上，切分出4个AI Core作为vNPU并挂载至容器。以此方式拉起的容器，在结束容器进程时，虚拟设备会自动销毁。

    ```shell
    docker run -it --rm -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_VNPU_SPECS=vir04 image-name:tag /bin/bash
    ```

>[!NOTE] 
>
>- 使用动态虚拟化时，需要关闭vNPU的恢复使能功能，该功能的详细说明和操作指导请参考《Atlas 中心推理卡  25.5.0 npu-smi 命令参考》中的“昇腾虚拟化实例（AVI）相关命令\>[设置vNPU的配置恢复使能状态](https://support.huawei.com/enterprise/zh/doc/EDOC1100540373/fa2a6907)”章节。
>- 可用的芯片ID可通过如下方式查询确认：
>   - 物理芯片ID：
>
>      ```shell
>      ls /dev/davinci*
>      ```
>
>   - 虚拟芯片ID：
>
>     ```shell
>     ls /dev/vdavinci*
>     ```
>
>- image-name:tag：镜像名称与标签，请根据实际情况修改。如“ascend-tensorflow:tensorflow\_TAG”。
>- 用户在使用过程中，请勿重复定义和在容器镜像中固定ASCEND\_VISIBLE\_DEVICES、ASCEND\_RUNTIME\_OPTIONS和ASCEND\_VNPU\_SPECS环境变量。
>- 使用动态虚拟化时，若发生服务器重启，则此场景下无法自动销毁vnpu，需用户自己手动销毁。

**表 1**  参数解释

<a name="zh-cn_topic_0000001136053188_table19948947144812"></a>

|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|必须使用ASCEND_VISIBLE_DEVICES环境变量指定被挂载至容器中的NPU设备，否则挂载NPU设备失败；使用NPU设备序号指定设备，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。|<ul><li>静态虚拟化：<ul><li>ASCEND_VISIBLE_DEVICES=100表示将100号vNPU挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=101,103表示将101、103号vNPU挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=100-102表示将100号至102号vNPU（包含100号和102号）挂载入容器中，效果同ASCEND_VISIBLE_DEVICES=100,101,102。</li><li>ASCEND_VISIBLE_DEVICES=100-102,104表示将100号至102号以及104号vNPU挂载入容器，效果同ASCEND_VISIBLE_DEVICES=100,101,102,104。</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y，其中XXX表示NPU设备，支持的取值为npu、Ascend910、Ascend310、Ascend310B和Ascend310P；Y表示物理NPU设备ID。<ul><li>ASCEND_VISIBLE_DEVICES=npu-101，表示把101号vNPU挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=npu-101,npu-103，表示把101号NPU和103号vNPU挂载进容器。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用芯片名称指定设备时，建议统一取值npu。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=101，npu-103。</li><li>必须搭配ASCEND_RUNTIME_OPTIONS，取值必须包含VIRTUAL，表示挂载的是vNPU。</li></ul></div></div></li><li>动态虚拟化：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AI Core。<ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS，表示在指定的NPU上划分出的AI Core数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul></li></ul></li></ul>|
|ASCEND_RUNTIME_OPTIONS|<p>对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：</p><ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul>|
|ASCEND_VNPU_SPECS|从物理NPU设备中划分出一定数量的AI Core，指定为虚拟设备。支持的取值请参见[表1](./03_virtualization_templates.md)。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li></ul>|ASCEND_VNPU_SPECS=vir04表示划分4个AI Core作为vNPU，挂载至容器。|

### 方式二：Kubernetes挂载vNPU<a name="ZH-CN_TOPIC_0000002511346321"></a>

#### 使用vNPU说明<a name="ZH-CN_TOPIC_0000002511426303"></a>

在Kubernetes场景，当用户需要使用vNPU资源时，需要通过结合集群调度组件Ascend Device Plugin的使用，使Kubernetes可以管理昇腾处理器资源。使用方式又按照是否需要提前切分好vNPU，划分为静态虚拟化和动态虚拟化两种，且两种模式不能混用，也不能和之前章节提到的Ascend Docker Runtime使用方式混合使用。昇腾虚拟化实例特性需要的集群调度组件如下表所示，支持的产品型号情况请参见[表1 产品支持情况说明](./01_description.md)。

**表 1**  虚拟化需要的集群调度组件

<a name="table19103194217329"></a>
<table><thead align="left"><th class="cellrowborder" valign="top" width="11.677219849801206%" id="mcps1.2.5.1.1"><p id="p2103642143218"><a name="p2103642143218"></a><a name="p2103642143218"></a>特性</p>
</th>
<th class="cellrowborder" valign="top" width="24.82697688116625%" id="mcps1.2.5.1.2"><p id="p619110456115"><a name="p619110456115"></a><a name="p619110456115"></a>需要的集群调度组件</p>
</th>
</thead>
<tbody><tr id="row61035425322"><td class="cellrowborder" rowspan="4" valign="top" width="11.677219849801206%" headers="mcps1.2.5.1.1 "><p id="p310384263219"><a name="p310384263219"></a><a name="p310384263219"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p4191645116"><a name="p4191645116"></a><a name="p4191645116"></a><span id="ph1795411794410"><a name="ph1795411794410"></a><a name="ph1795411794410"></a>Ascend Device Plugin</span></p>
</td>
</tr>
<tr id="row1844495022714"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p574771602812"><a name="p574771602812"></a><a name="p574771602812"></a>（可选）<span id="ph1610211588167">Volcano</span></p>
</td>
</tr>
<tr id="row18230132874912"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11381824102511"><a name="p11381824102511"></a><a name="p11381824102511"></a>（可选）<span id="ph1566531814589">Ascend Operator</span></p>
</td>
</tr>
<tr><td><p>（可选）<span>ClusterD</span></p>
</td>
</tr>
<tr id="row610314214324"><td class="cellrowborder" rowspan="4" valign="top" width="11.677219849801206%" headers="mcps1.2.5.1.1 "><p id="p11036426328"><a name="p11036426328"></a><a name="p11036426328"></a>动态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p1219211451715"><a name="p1219211451715"></a><a name="p1219211451715"></a><span id="ph12922181924413"><a name="ph12922181924413"></a><a name="ph12922181924413"></a>Ascend Device Plugin</span></p>
</td>
</tr>
<tr><td><p><span>Volcano</span></p>
</td>
</tr>
<tr><td><p>（可选）<span>Ascend Operator</span></p>
</td>
</tr>
<tr><td><p>（可选）<span>ClusterD</span></p>
</td>
</tr>
</tbody>
</table>

>[!NOTE]  
>Ascend Device Plugin组件的安装请参见[Ascend Device Plugin](../../../installation_guide/03_installation.md#ascend-device-plugin)。
>在静态虚拟化场景下，组件的可选性说明如下。
>
>- Volcano：用户若使用自己的调度组件，需要进行参数配置，请参见[表2](#table1064314568229)；用户也可直接使用该组件进行任务调度。
>- Ascend Operator：当使用训练系列产品时才需要选择该组件；使用推理系列产品时可不选择。
>- ClusterD：当使用Volcano时才需要选择该组件，详细请参见[安装Volcano](../../../installation_guide/03_installation.md#安装volcano)。

#### 静态虚拟化<a name="ZH-CN_TOPIC_0000002479226392"></a>

**使用限制<a name="section785220396317"></a>**

- 当前vNPU仅支持单个vNPU单容器任务，不支持创建多副本任务。
- 任务运行过程中，不支持卸载Volcano。
- 目前任务的每个Pod请求的NPU设备数量规则如下：

    使用切分后的vNPU，则仅支持1个。

- 静态虚拟化场景，如果创建或者销毁vNPU，需要重启Ascend Device Plugin。
- 静态虚拟化任务，不支持故障重调度。

**表 1**  虚拟化实例模板与虚拟设备类型关系表

<a name="table47415104403"></a>
<table><thead align="left"><tr id="row67416101402"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p117491014400"><a name="p117491014400"></a><a name="p117491014400"></a>NPU类型</p>
</th>
<th class="cellrowborder" valign="top" width="19.96%" id="mcps1.2.5.1.2"><p id="p177431064013"><a name="p177431064013"></a><a name="p177431064013"></a>虚拟化实例模板</p>
</th>
<th class="cellrowborder" valign="top" width="20.04%" id="mcps1.2.5.1.3"><p id="p1374210134015"><a name="p1374210134015"></a><a name="p1374210134015"></a>vNPU类型</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p1041963771317"><a name="p1041963771317"></a><a name="p1041963771317"></a>具体虚拟设备名称（以vNPU ID100、物理卡ID0为例）</p>
</th>
</tr>
</thead>
<tbody><tr id="row5741710164014"><td class="cellrowborder" rowspan="4" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p074181014408"><a name="p074181014408"></a><a name="p074181014408"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a>Atlas 训练系列产品</span>（30或32个AI Core）</p>
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
</tr>
<tr id="row20758109404"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1375910194012"><a name="p1375910194012"></a><a name="p1375910194012"></a>vir16</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p075131044012"><a name="p075131044012"></a><a name="p075131044012"></a>Ascend910-16c</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p188588202135"><a name="p188588202135"></a><a name="p188588202135"></a>Ascend910-16c-100-0</p>
</td>
</tr>
<tr><td rowspan="4" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A2 训练系列产品</span>（20或24或25个AI Core）</p>
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
<tr><td><p>vir12_3c_32g</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A3 训练系列产品</span>（48个AI Core）</p>
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
<tr id="row84911853114212"><td class="cellrowborder" rowspan="7" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>（8个AI Core）</p>
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
<tr><td rowspan="6" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A2 推理系列产品</span>（20个AI Core）</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A3 推理系列产品</span>（40个AI Core）</p>
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

**前提条件<a name="section18128140645"></a>**

1. 需要先获取“Ascend-docker-runtime\_\{version\}\_linux-\{arch\}.run”，安装容器引擎插件。
2. 参见[安装部署](../../../installation_guide/03_installation.md)章节，完成各组件的安装。

    虚拟化实例涉及到需要修改相关参数的集群调度组件为Volcano和Ascend Device Plugin，请按如下要求修改并使用对应的YAML安装部署：

    - 亲和性场景：需要安装Volcano。
    - 非亲和性场景：不需要安装Volcano，只会上报设备数量给节点的K8s。

    1. Ascend Device Plugin参数修改及启动说明：

        虚拟化实例启动参数说明如下：

        **表 2** Ascend Device Plugin启动参数

        <a name="table1064314568229"></a>

        |参数|类型|默认值|说明|
        |--|--|--|--|
        |-volcanoType|bool|false|是否使用Volcano进行调度，如使用动态虚拟化，需要设置为true。|
        |-presetVirtualDevice|bool|true|静态虚拟化功能开关，值只能为true。<p>如使用动态虚拟化，需要设置为false，并需要同步开启Volcano，即设置“-volcanoType”参数为true。</p>|

        YAML启动说明如下：

        - K8s集群中存在使用Atlas 推理系列产品节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

            ```shell
            kubectl apply -f device-plugin-310P-v{version}.yaml
            ```

        - K8s集群中存在使用Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品、Atlas A2 推理系列产品、Atlas A3 推理系列产品节点（Ascend Device Plugin独立工作，不配合Volcano和Ascend Operator使用）。

            ```shell
            kubectl apply -f device-plugin-910-v{version}.yaml
            ```

        - K8s集群中存在使用Atlas 推理系列产品节点（使用Volcano调度器，支持NPU虚拟化，YAML默认关闭动态虚拟化）。

            ```shell
            kubectl apply -f device-plugin-310P-volcano-v{version}.yaml
            ```

        - K8s集群中存在使用Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品、Atlas A2 推理系列产品、Atlas A3 推理系列产品节点（配合Volcano和Ascend Operator使用，支持NPU虚拟化，YAML默认关闭动态虚拟化）。

            ```shell
            kubectl apply -f device-plugin-volcano-v{version}.yaml
            ```

        如果K8s集群使用了多种类型的昇腾AI处理器，请分别执行对应命令。

    2. Volcano参数修改及启动说明：

        在Volcano部署文件“volcano-v\{version\}.yaml”中，需要配置“presetVirtualDevice”且值只能为“true”。

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
              - name: volcano-npu-v26.0.0_linux-aarch64    # 其中26.0.0为MindCluster的版本号，根据不同版本，该处取值不同
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

**使用方法<a name="section514441719341"></a>**

- 创建训练任务时，需要在创建YAML文件时，修改如下配置。以Atlas 训练系列产品使用为例。

    resources中设定的requests和limits资源类型，应修改为huawei.com/Ascend910-_Y_，其中<i>Y</i>值和vNPU类型相关，具体取值参考[表 虚拟化实例模板与虚拟设备类型关系表](#table47415104403)中的虚拟类型。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend910-Y: 1          # 请求的vNPU数量，最大值为1。
                limits:
                  huawei.com/Ascend910-Y: 1          # 数值与请求数量一致。
    ...
    ```

- 创建推理任务时，需要在创建YAML文件时，修改如下配置。以Atlas 推理系列产品使用为例。

    resources中设定的requests和limits资源类型，应修改为huawei.com/Ascend310P-_Y_，其中<i>Y</i>值和vNPU类型相关，具体取值参考[表 虚拟化实例模板与虚拟设备类型关系表](#table47415104403)中的虚拟类型。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend310P-Y: 1          # 请求的vNPU数量，最大值为1。
                limits:
                  huawei.com/Ascend310P-Y: 1          # 数值与请求数量一致。
    ...
    ```

#### 动态虚拟化<a name="ZH-CN_TOPIC_0000002511426291"></a>

使用动态虚拟化前，需要提前了解[表1](#table625511844619)中的相关使用说明。

**使用说明<a name="section1576110260450"></a>**

**表 1**  场景说明

<a name="table625511844619"></a>
<table><thead align="left"><tr id="row9255148204610"><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p id="p4381442125317"><a name="p4381442125317"></a><a name="p4381442125317"></a>场景</p>
</th>
<th class="cellrowborder" valign="top" width="80.02%" id="mcps1.2.3.1.2"><p id="p2255984464"><a name="p2255984464"></a><a name="p2255984464"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row132012115910"><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p1950512911598"><a name="p1950512911598"></a><a name="p1950512911598"></a>通用说明</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p450516910592"><a name="p450516910592"></a><a name="p450516910592"></a>分配的芯片信息会在Pod的annotation中体现出来，关于Pod annotation的详细说明请参见<a href="../../../api/k8s.md">Pod annotation</a>中的huawei.com/npu-core、huawei.com/AscendReal参数。</p>
</td>
</tr>
<tr id="row48061646595"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1749665239"><a name="p1749665239"></a><a name="p1749665239"></a>同一时刻，只能下发相同<a href="./03_virtualization_templates.md">虚拟化模板</a>的任务。</p>
</td>
</tr>
<tr id="row18542176195917"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p450559185914"><a name="p450559185914"></a><a name="p450559185914"></a>动态分配vNPU时，经<span id="ph19255162231216"><a name="ph19255162231216"></a><a name="ph19255162231216"></a>MindCluster</span>调度，将优先占满剩余算力最少的物理NPU。</p>
</td>
</tr>
<tr id="row11648825917"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p19505796596"><a name="p19505796596"></a><a name="p19505796596"></a>目前任务的每个Pod请求的NPU数量为1个。</p>
</td>
</tr>
<tr id="row32567817461"><td class="cellrowborder" rowspan="3" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p1325613818460"><a name="p1325613818460"></a><a name="p1325613818460"></a>特性支持的场景</p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p32561983469"><a name="p32561983469"></a><a name="p32561983469"></a>支持多副本，但多副本中的每个pod都必须使用vNPU。</p>
</td>
</tr>
<tr id="row5256198134612"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p72561586465"><a name="p72561586465"></a><a name="p72561586465"></a>支持K8s的机制，如亲和性等。</p>
</td>
</tr>
<tr id="row825611817468"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p2795151384913"><a name="p2795151384913"></a><a name="p2795151384913"></a>支持芯片故障和节点故障的重调度。具体参考<span id="ph1389215534914"><a name="ph1389215534914"></a><a name="ph1389215534914"></a><a href="../../basic_scheduling.md#推理卡故障恢复">推理卡故障恢复</a></span>和<a href="../../basic_scheduling.md#推理卡故障重调度">推理卡故障重调度</a>章节。</p>
</td>
</tr>
<tr id="row237762345420"><td class="cellrowborder" rowspan="4" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p840574125511"><a name="p840574125511"></a><a name="p840574125511"></a>特性不支持的场景</p>
<p id="p17835104672517"><a name="p17835104672517"></a><a name="p17835104672517"></a></p>
<p id="p36763525314"><a name="p36763525314"></a><a name="p36763525314"></a></p>
<p id="p767616565314"><a name="p767616565314"></a><a name="p767616565314"></a></p>
<p id="p667616595317"><a name="p667616595317"></a><a name="p667616595317"></a></p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p14377152385414"><a name="p14377152385414"></a><a name="p14377152385414"></a>不支持不同芯片在一个任务内混用。</p>
</td>
</tr>
<tr id="row1625614818462"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p32566874611"><a name="p32566874611"></a><a name="p32566874611"></a>任务运行过程中，不支持卸载<span id="ph42462611516"><a name="ph42462611516"></a><a name="ph42462611516"></a>Volcano</span>。</p>
</td>
</tr>
<tr id="row1854910515540"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p12256108124616"><a name="p12256108124616"></a><a name="p12256108124616"></a>K8s场景会自动创建与销毁vNPU，不能与Docker场景的操作混用。</p>
</td>
</tr>
<tr id="row151011624135113"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18102182414515"><a name="p18102182414515"></a><a name="p18102182414515"></a>进行动态虚拟化的节点不能对芯片的CPU进行设置。详情请参考<span id="ph373734654014"><a name="ph373734654014"></a><a name="ph373734654014"></a>《Atlas 中心推理卡  25.5.0 npu-smi 命令参考》中的“信息查询&gt;<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100540373/6faea171" target="_blank" rel="noopener noreferrer">查询所有芯片的AI CPU、control CPU和data CPU数量</a>”</span>章节。</p>
</td>
</tr>
<tr id="row192561854613"><td class="cellrowborder" rowspan="3" valign="top" width="19.98%" headers="mcps1.2.3.1.1 "><p id="p1125610854611"><a name="p1125610854611"></a><a name="p1125610854611"></a><span id="ph10445185418466"><a name="ph10445185418466"></a><a name="ph10445185418466"></a>Atlas 推理系列产品</span>（8个AI Core）使用说明</p>
<p id="p1173133213564"><a name="p1173133213564"></a><a name="p1173133213564"></a></p>
</td>
<td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.2 "><p id="p02561481463"><a name="p02561481463"></a><a name="p02561481463"></a>任务请求的AI Core数量，为vNPU时，按实际填写1、2、4；整张物理NPU时，需要为8以及8的倍数。</p>
</td>
</tr>
<tr id="row11782173617479"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18782936144718"><a name="p18782936144718"></a><a name="p18782936144718"></a>默认需要容器以root用户启动，若需要以普通用户运行推理任务，需要参考<a href="../../../faq.md#使用动态虚拟化时以普通用户运行推理业务失败">使用动态虚拟化时，以普通用户运行推理业务容器失败</a>章节进行操作。</p>
</td>
</tr>
<tr id="row117233216566"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18081933105617"><a name="p18081933105617"></a><a name="p18081933105617"></a>vNPU动态创建和销毁仅在<span id="ph20808153335610"><a name="ph20808153335610"></a><a name="ph20808153335610"></a>Atlas 推理系列产品</span>上有效，并且需要配套<span id="ph13808233145619"><a name="ph13808233145619"></a><a name="ph13808233145619"></a>Volcano</span>使用。</p>
</td>
</tr>
</tbody>
</table>

**表 2**  虚拟化实例模板与虚拟设备类型关系表

<a name="table47415104403"></a>
<table><thead align="left"><tr id="row67416101402"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p117491014400"><a name="p117491014400"></a><a name="p117491014400"></a>NPU类型</p>
</th>
<th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.5.1.2"><p id="p177431064013"><a name="p177431064013"></a><a name="p177431064013"></a>虚拟化实例模板</p>
</th>
<th class="cellrowborder" valign="top" width="20.02%" id="mcps1.2.5.1.3"><p id="p1374210134015"><a name="p1374210134015"></a><a name="p1374210134015"></a>vNPU类型</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p1041963771317"><a name="p1041963771317"></a><a name="p1041963771317"></a>具体虚拟设备名称（以vNPU ID100、物理卡ID0为例）</p>
</th>
</tr>
</thead>
<tbody><tr id="row84911853114212"><td class="cellrowborder" rowspan="7" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph1534112451967"><a name="ph1534112451967"></a><a name="ph1534112451967"></a>Atlas 推理系列产品</span>（8个AI Core）</p>
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

**前提条件<a name="section18128140645"></a>**

1. 需要先获取“Ascend-docker-runtime\_\{version\}\_linux-\{arch\}.run”，安装容器引擎插件。
2. 参见[安装部署](../../../installation_guide/03_installation.md)章节，完成各组件的安装。

    虚拟化实例涉及到需要修改相关参数的集群调度组件为Volcano和Ascend Device Plugin，请按如下要求修改并使用对应的YAML安装部署：

    1. Ascend Device Plugin参数修改及启动说明。

        虚拟化实例启动参数说明如下：

        **表 3** Ascend Device Plugin启动参数

        <a name="table1064314568229"></a>

        |参数|类型|默认值|说明|
        |--|--|--|--|
        |-volcanoType|bool|false|是否使用Volcano进行调度，如使用动态虚拟化，需要设置为true。|
        |-presetVirtualDevice|bool|true|静态虚拟化功能开关，值只能为true。<p>如使用动态虚拟化，需要设置为false，并需要同步开启Volcano，即设置“-volcanoType”参数为true。</p>|

        YAML启动说明如下：

        K8s集群中存在使用Atlas 推理系列产品的节点，需要在device-plugin-310P-volcano-v\{version\}中将“presetVirtualDevice”字段修改为“false”（协同Volcano使用，支持NPU虚拟化，YAML默认关闭动态虚拟化）。

        ```Yaml
        ...
        args: [ "device-plugin  -useAscendDocker=true -volcanoType=true -presetVirtualDevice=false
                   -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=0" ]
        ...
        ```

    2. Volcano参数修改及启动说明。

        在Volcano部署文件“volcano-v_\{version\}_.yaml”中，需要配置“presetVirtualDevice”的值为“false”。

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
                arguments: {"grace-over-time":"900","presetVirtualDevice":"false"}  # 开启动态虚拟化，presetVirtualDevice的值需要设置为false
        ...
        ```

**使用方法<a name="section514441719341"></a>**

创建推理任务时，需要在创建YAML文件时，修改如下配置。以Atlas 推理系列产品使用为例。

resources中设定的requests和limits资源类型，申请一个AI Core，应修改为huawei.com/npu-core。以deployment部署方式为例：

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
        ring-controller.atlas: ascend-310P  # 参见表4
        fault-scheduling: "grace"           # 重调度所使用的label
        vnpu-dvpp: "yes"                    # 参见表4
        vnpu-level: "low"                   # 参见表4
    spec:
      schedulerName: volcano  # 需要使用MindCluster的调度器Volcano
      nodeSelector:
        host-arch: huawei-arm
      containers:
        - image: ubuntu:22.04   # 示例镜像
          imagePullPolicy: IfNotPresent
          name: tf
          command:
          - "/bin/bash"
          - "-c"
          args: [ "客户自己的运行脚本"  ]
          resources:
            requests:
              huawei.com/npu-core: 1        # 使用vir01模板动态虚拟化NPU
            limits:
              huawei.com/npu-core: 1        # 数值与requests一致。
 ....
```

**表 4**  虚拟化实例任务YAML中label说明

<a name="table1084325844716"></a>
<table><thead align="left"><tr id="row13843105815479"><th class="cellrowborder" valign="top" width="17.88178817881788%" id="mcps1.2.4.1.1"><p id="p1879944394819"><a name="p1879944394819"></a><a name="p1879944394819"></a>key</p>
</th>
<th class="cellrowborder" valign="top" width="31.053105310531055%" id="mcps1.2.4.1.2"><p id="p6307191712494"><a name="p6307191712494"></a><a name="p6307191712494"></a>value</p>
</th>
<th class="cellrowborder" valign="top" width="51.06510651065107%" id="mcps1.2.4.1.3"><p id="p571812231496"><a name="p571812231496"></a><a name="p571812231496"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row11843135814719"><td class="cellrowborder" rowspan="2" valign="top" width="17.88178817881788%" headers="mcps1.2.4.1.1 "><p id="p11799943114814"><a name="p11799943114814"></a><a name="p11799943114814"></a>vnpu-level</p>
<p id="p127511550154811"><a name="p127511550154811"></a><a name="p127511550154811"></a></p>
</td>
<td class="cellrowborder" valign="top" width="31.053105310531055%" headers="mcps1.2.4.1.2 "><p id="p73071317144911"><a name="p73071317144911"></a><a name="p73071317144911"></a>low</p>
</td>
<td class="cellrowborder" valign="top" width="51.06510651065107%" headers="mcps1.2.4.1.3 "><p id="p20719152316493"><a name="p20719152316493"></a><a name="p20719152316493"></a>低配，默认值，选择最低配置的“虚拟化实例模板”。</p>
</td>
</tr>
<tr id="row1475114503484"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p12307151724910"><a name="p12307151724910"></a><a name="p12307151724910"></a>high</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1271902314916"><a name="p1271902314916"></a><a name="p1271902314916"></a>性能优先。</p>
<p id="p071922312490"><a name="p071922312490"></a><a name="p071922312490"></a>在集群资源充足的情况下，将选择尽量高配的虚拟化实例模板；在整个集群资源已使用过多的情况下，如大部分物理NPU都已使用，每个物理NPU只剩下小部分AI Core，不足以满足高配虚拟化实例模板时，将使用相同AI Core数量下较低配置的其他模板。具体选择请参考<a href="#table83781115185619">表5</a>。</p>
</td>
</tr>
<tr id="row8843145854711"><td class="cellrowborder" rowspan="3" valign="top" width="17.88178817881788%" headers="mcps1.2.4.1.1 "><p id="p168872618492"><a name="p168872618492"></a><a name="p168872618492"></a>vnpu-dvpp</p>
</td>
<td class="cellrowborder" valign="top" width="31.053105310531055%" headers="mcps1.2.4.1.2 "><p id="p2030751719499"><a name="p2030751719499"></a><a name="p2030751719499"></a>yes</p>
</td>
<td class="cellrowborder" valign="top" width="51.06510651065107%" headers="mcps1.2.4.1.3 "><p id="p971972316498"><a name="p971972316498"></a><a name="p971972316498"></a>该Pod使用DVPP。</p>
</td>
</tr>
<tr id="row165811357114820"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1630820172490"><a name="p1630820172490"></a><a name="p1630820172490"></a>no</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5719152304920"><a name="p5719152304920"></a><a name="p5719152304920"></a>该Pod不使用DVPP。</p>
</td>
</tr>
<tr id="row173650119495"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p12308131744912"><a name="p12308131744912"></a><a name="p12308131744912"></a>null</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1871982374915"><a name="p1871982374915"></a><a name="p1871982374915"></a>默认值。不关注是否使用DVPP。</p>
</td>
</tr>
<tr id="row184385814710"><td class="cellrowborder" valign="top" width="17.88178817881788%" headers="mcps1.2.4.1.1 "><p id="p1680094354812"><a name="p1680094354812"></a><a name="p1680094354812"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="31.053105310531055%" headers="mcps1.2.4.1.2 "><p id="p530851794913"><a name="p530851794913"></a><a name="p530851794913"></a>ascend-310P</p>
</td>
<td class="cellrowborder" valign="top" width="51.06510651065107%" headers="mcps1.2.4.1.3 "><p id="p1871918233494"><a name="p1871918233494"></a><a name="p1871918233494"></a>任务使用<span id="ph968626194020"><a name="ph968626194020"></a><a name="ph968626194020"></a>Atlas 推理系列产品</span>的标识。</p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 
>vnpu-level和vnpu-dvpp的选择结果，具体请参见[表5](#table83781115185619)。
>
>- 表中“降级”表示AI Core满足的情况下，其他资源不够（如AI CPU）时，模板会选择同AI Core下的其他满足资源要求的模板。如在只剩一颗芯片上只有2个AI Core，1个AI CPU时，vir02模板会降级为vir02\_1c。
>- 表中“选择模板”中的值来源于[虚拟化模板](./03_virtualization_templates.md)中Atlas 推理系列产品、“虚拟化实例模板”列的取值。
>- 表中“vnpu-level”列的“其他值”表示除去“low”和“high”后的任意取值。
>- 整卡（core的请求数量为8的倍数）场景下vnpu-dvpp与vnpu-level可以取任意值。

**表 5**  dvpp和level作用结果表

<a name="table83781115185619"></a>
<table><thead align="left"><tr id="row1837817157565"><th class="cellrowborder" valign="top" width="17.2982701729827%" id="mcps1.2.7.1.1"><p id="p11560216112"><a name="p11560216112"></a><a name="p11560216112"></a>产品型号</p>
</th>
<th class="cellrowborder" valign="top" width="16.42835716428357%" id="mcps1.2.7.1.2"><p id="p1024717408463"><a name="p1024717408463"></a><a name="p1024717408463"></a>AI Core请求数量</p>
</th>
<th class="cellrowborder" valign="top" width="15.768423157684234%" id="mcps1.2.7.1.3"><p id="p192479402463"><a name="p192479402463"></a><a name="p192479402463"></a>vnpu-dvpp</p>
</th>
<th class="cellrowborder" valign="top" width="20.987901209879013%" id="mcps1.2.7.1.4"><p id="p1024716402460"><a name="p1024716402460"></a><a name="p1024716402460"></a>vnpu-level</p>
</th>
<th class="cellrowborder" valign="top" width="8.52914708529147%" id="mcps1.2.7.1.5"><p id="p8247440174613"><a name="p8247440174613"></a><a name="p8247440174613"></a>是否降级</p>
</th>
<th class="cellrowborder" valign="top" width="20.987901209879013%" id="mcps1.2.7.1.6"><p id="p0247164034611"><a name="p0247164034611"></a><a name="p0247164034611"></a>选择模板</p>
</th>
</tr>
</thead>
<tbody><tr id="row1517703912018"><td class="cellrowborder" rowspan="12" valign="top" width="17.2982701729827%" headers="mcps1.2.7.1.1 "><p id="p8916171416125"><a name="p8916171416125"></a><a name="p8916171416125"></a><span id="ph1856391311016"><a name="ph1856391311016"></a><a name="ph1856391311016"></a>Atlas 推理系列产品</span>（8个AI Core）</p>
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
<td class="cellrowborder" valign="top" width="20.987901209879013%" headers="mcps1.2.7.1.4 "><p id="p1385717396538"><a name="p1385717396538"></a><a name="p1385717396538"></a>任意值</p>
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
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p14619832145315"><a name="p14619832145315"></a><a name="p14619832145315"></a>low/其他值</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p126198326538"><a name="p126198326538"></a><a name="p126198326538"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p3248164094613"><a name="p3248164094613"></a><a name="p3248164094613"></a>vir02_1c</p>
</td>
</tr>
<tr id="row117818394016"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.7.1.1 "><p id="p162489402463"><a name="p162489402463"></a><a name="p162489402463"></a>high</p>
<p id="p143218450593"><a name="p143218450593"></a><a name="p143218450593"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p22482040124615"><a name="p22482040124615"></a><a name="p22482040124615"></a>否</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p182481740174611"><a name="p182481740174611"></a><a name="p182481740174611"></a>vir02</p>
</td>
</tr>
<tr id="row16943192222113"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1324834017468"><a name="p1324834017468"></a><a name="p1324834017468"></a>是</p>
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
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.3 "><p id="p3248184024610"><a name="p3248184024610"></a><a name="p3248184024610"></a>low/其他值</p>
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
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p10249124011467"><a name="p10249124011467"></a><a name="p10249124011467"></a>否</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p324964074618"><a name="p324964074618"></a><a name="p324964074618"></a>vir04</p>
</td>
</tr>
<tr id="row16813143918117"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2249340144615"><a name="p2249340144615"></a><a name="p2249340144615"></a>是</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p924924064613"><a name="p924924064613"></a><a name="p924924064613"></a>vir04_3c</p>
</td>
</tr>
<tr id="row1781312397116"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p102497405465"><a name="p102497405465"></a><a name="p102497405465"></a>8或8的倍数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p42491440174615"><a name="p42491440174615"></a><a name="p42491440174615"></a>任意值</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p5249114074614"><a name="p5249114074614"></a><a name="p5249114074614"></a>任意值</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1224920403467"><a name="p1224920403467"></a><a name="p1224920403467"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.5 "><p id="p55031522345"><a name="p55031522345"></a><a name="p55031522345"></a>-</p>
</td>
</tr>
<tr id="row74471126913"><td class="cellrowborder" colspan="6" valign="top" headers="mcps1.2.7.1.1 mcps1.2.7.1.2 mcps1.2.7.1.3 mcps1.2.7.1.4 mcps1.2.7.1.5 mcps1.2.7.1.6 "><p id="p627014191100"><a name="p627014191100"></a><a name="p627014191100"></a>注：</p>
<p id="p9942971914"><a name="p9942971914"></a><a name="p9942971914"></a>如果是<span id="ph884102218100"><a name="ph884102218100"></a><a name="ph884102218100"></a>Atlas 推理系列产品</span>（8个AI Core），必须申请AI Core数量为8或8的倍数。</p>
</td>
</tr>
</tbody>
</table>

>[!NOTICE] 
>上表中对于芯片虚拟化（非整卡），vnpu-dvpp的值只能为表中对应的值，其他值会导致任务不能下发。
