# 挂载vNPU（静态虚拟化）<a name="ZH-CN_TOPIC_0000002479386388"></a>

## 方式一：基于原生Docker挂载vNPU

原生Docker场景下（未部署MindCluster集群调度组件），需要使用npu-smi工具创建vNPU后，将vNPU挂载到容器。具体操作请参见《Atlas 中心训练服务器 NPU驱动和固件安装指南》的“昇腾虚拟化实例（AVI）容器场景下的安装与卸载>[多容器场景下安装](https://support.huawei.com/enterprise/zh/doc/EDOC1100568429/5b32515a)”章节，该章节指导用户安装Docker和将vNPU挂载进容器。

## 方式二：Ascend Docker Runtime挂载vNPU

单独结合Ascend Docker Runtime（容器引擎插件）使用，将vNPU挂载到容器。

### 前提条件

需要先安装容器引擎插件Ascend Docker Runtime，方法可参见[安装Ascend Docker Runtime](../../../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md)。

### 使用方法

用户已通过npu-smi工具创建vNPU，在拉起容器时执行以下命令将vNPU挂载至容器中。以下命令表示用户在拉起容器时，挂载虚拟芯片ID为100的芯片。

```shell
docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL {image-name:tag} /bin/bash
```

>[!NOTE]
>
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
>- image-name:tag：镜像名称与标签，请根据实际情况修改。
>- 用户在使用过程中，请勿重复定义和在容器镜像中固定ASCEND_VISIBLE_DEVICES和ASCEND_RUNTIME_OPTIONS环境变量。

**表 1**  参数解释

|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|必须使用ASCEND_VISIBLE_DEVICES环境变量指定被挂载至容器中的NPU设备，否则挂载NPU设备失败；使用NPU设备序号指定设备，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。|<ul><li>ASCEND_VISIBLE_DEVICES=100表示将100号vNPU挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=101,103表示将101、103号vNPU挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=100-102表示将100号至102号vNPU（包含100号和102号）挂载入容器中，效果同ASCEND_VISIBLE_DEVICES=100,101,102。</li><li>ASCEND_VISIBLE_DEVICES=100-102,104表示将100号至102号以及104号vNPU挂载入容器，效果同ASCEND_VISIBLE_DEVICES=100,101,102,104。</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y，其中XXX表示NPU设备，支持的取值为npu、Ascend910、Ascend310、Ascend310B和Ascend310P；Y表示物理NPU设备ID。<ul><li>ASCEND_VISIBLE_DEVICES=npu-101，表示把101号vNPU挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=npu-101,npu-103，表示把101号NPU和103号vNPU挂载进容器。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用芯片名称指定设备时，建议统一取值npu。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=101，npu-103。</li><li>必须搭配ASCEND_RUNTIME_OPTIONS，取值必须包含VIRTUAL，表示挂载的是vNPU。</li></ul></div></div></li></ul>|
|ASCEND_RUNTIME_OPTIONS|<p>对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：</p><ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>静态虚拟化场景下，ASCEND_RUNTIME_OPTIONS为必选参数，且取值必须包含VIRTUAL。</li></ul></div></div>|

## 方式三：Kubernetes挂载vNPU

### 使用说明<a name="ZH-CN_TOPIC_0000002511426303"></a>

在Kubernetes场景，当用户需要使用vNPU资源时，需要通过结合集群调度组件Ascend Device Plugin的使用，使Kubernetes可以管理昇腾处理器资源。静态虚拟化场景使用时，不能与动态虚拟化混合使用。昇腾虚拟化实例特性需要的集群调度组件如下表所示，支持的产品型号情况请参见[特性说明](../01_description.md)中的“表1 产品支持情况说明”。

**表 1**  虚拟化需要的集群调度组件

<a name="table19103194217329"></a>
<table><thead align="left"><th class="cellrowborder" valign="top" width="11.677219849801206%" id="mcps1.2.5.1.1"><p id="p2103642143218"><a name="p2103642143218"></a><a name="p2103642143218"></a>特性</p>
</th>
<th class="cellrowborder" valign="top" width="24.82697688116625%" id="mcps1.2.5.1.2"><p id="p619110456115"><a name="p619110456115"></a><a name="p619110456115"></a>需要的集群调度组件</p>
</th>
</thead>
<tbody><tr id="row61035425322"><td class="cellrowborder" rowspan="5" valign="top" width="11.677219849801206%" headers="mcps1.2.5.1.1 "><p id="p310384263219"><a name="p310384263219"></a><a name="p310384263219"></a>静态虚拟化</p>
</td>
<td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p4191645116"><a name="p4191645116"></a><a name="p4191645116"></a><span id="ph1795411794410"><a name="ph1795411794410"></a><a name="ph1795411794410"></a>Ascend Docker Runtime</span></p>
</td>
</tr>
<tr id="row1844495022714"><td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p4191645116"><a name="p4191645116"></a><a name="p4191645116"></a><span id="ph1795411794410"><a name="ph1795411794410"></a><a name="ph1795411794410"></a>Ascend Device Plugin</span></p>
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
</tbody>
</table>

>[!NOTE]
>在静态虚拟化场景下，组件的可选性说明如下。
>
>- Volcano：用户若使用自己的调度组件，需要进行参数配置，请参见[表3](#table1064314568229)；用户也可直接使用该组件进行任务调度。
>- Ascend Operator：当使用训练系列产品时才需要选择该组件；使用推理系列产品时可不选择。
>- ClusterD：当使用Volcano时才需要选择该组件，详细请参见[安装Volcano](../../../../05_developer_guide/00_installation_deployment/00_manual_installation/05_volcano.md#安装volcano)。

### 使用限制<a name="ZH-CN_TOPIC_0000002479226392"></a>

- 任务运行过程中，不支持卸载Volcano。
- 目前任务的每个Pod请求的NPU设备数量规则如下：

    使用切分后的vNPU，则仅支持1个。

- 静态虚拟化场景，如果创建或者销毁vNPU，需要重启Ascend Device Plugin。
- 静态虚拟化任务，不支持故障重调度。
- 不支持静态vNPU进行交换机亲和性调度。
- 静态vNPU调度暂不支持ASCEND_VISIBLE_DEVICES相关字段，如存在以下字段，需要删除：

```yaml
...
                env:
                - name: ASCEND_VISIBLE_DEVICES
                  valueFrom:
                    fieldRef:
                      fieldPath: metadata.annotations['huawei.com/Ascend310P']
...
```

**表 2**  虚拟化实例模板与vNPU类型关系表

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
<tbody><tr id="row5741710164014"><td class="cellrowborder" rowspan="4" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p074181014408"><a name="p074181014408"></a><a name="p074181014408"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a>Atlas 训练系列产品</span>（30或32个AICore）</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A2 训练系列产品</span>（24个AICore）</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A3 训练系列产品</span>（48个AICore）</p>
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
<tr id="row84911853114212"><td class="cellrowborder" rowspan="7" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>（8个AICore）</p>
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
<tr><td rowspan="6" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A2 推理系列产品</span>（20个AICore）</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span>Atlas A3 推理系列产品</span>（40个AICore）</p>
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

### 前提条件

1. 需要先获取"Ascend-docker-runtime\_\{version\}\_linux-\{arch\}.run"，安装容器引擎插件。
2. 参见[安装部署](../../../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md)章节，完成各组件的安装。

    虚拟化实例涉及到需要修改相关参数的集群调度组件为Volcano和Ascend Device Plugin，请按如下要求修改并使用对应的YAML安装部署：

    - 亲和性场景：需要安装Volcano。
    - 非亲和性场景：不需要安装Volcano，只会上报设备数量给节点的K8s。

    1. Ascend Device Plugin参数修改及启动说明：

        虚拟化实例启动参数说明如下：

        **表 3** Ascend Device Plugin启动参数

        <a name="table1064314568229"></a>

        |参数|类型|默认值|说明|
        |--|--|--|--|
        |-volcanoType|bool|false|是否使用Volcano进行调度，如使用动态虚拟化，需要设置为true。|
        |-presetVirtualDevice|bool|true|静态虚拟化功能开关，值只能为true。<p>如使用动态虚拟化，需要设置为false，并需要同步开启Volcano，即设置"-volcanoType"参数为true。</p>|

        YAML启动说明如下：

        - K8s集群中存在使用Atlas 推理系列产品节点、Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品、Atlas A2 推理系列产品、Atlas A3 推理系列产品节点（Ascend Device Plugin独立工作，不配合Volcano和Ascend Operator使用）。

            ```shell
            kubectl apply -f device-plugin-v{version}.yaml
            ```

        - K8s集群中存在使用Atlas 推理系列产品节点、Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品、Atlas A2 推理系列产品、Atlas A3 推理系列产品节点（配合Volcano和Ascend Operator使用，支持NPU虚拟化，YAML默认关闭动态虚拟化）。

            ```shell
            kubectl apply -f device-plugin-volcano-v{version}.yaml
            ```

        如果K8s集群使用了多种类型的昇腾AI处理器，请分别执行对应命令。

    2. Volcano参数修改及启动说明：

        在Volcano部署文件"volcano-v<i>\{version\}</i>.yaml"中，需要配置"presetVirtualDevice"且值只能为"true"。

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
              - name: volcano-npu-v26.1.0_linux-aarch64    # 其中26.1.0为MindCluster的版本号，根据不同版本，该处取值不同
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

### 使用方法

- 创建训练任务时，需要在创建YAML文件时，修改如下配置。以Atlas 训练系列产品使用为例。

    resources中设定的requests和limits资源类型，应修改为huawei.com/Ascend910-_Y_，其中<i>Y</i>值和vNPU类型相关，具体取值参考[表2 虚拟化实例模板与vNPU类型关系表](#table47415104403)中的“vNPU类型”列。

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

    resources中设定的requests和limits资源类型，应修改为huawei.com/Ascend310P-_Y_，其中<i>Y</i>值和vNPU类型相关，具体取值参考[表2 虚拟化实例模板与vNPU类型关系表](#table47415104403)中的“vNPU类型”列。

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend310P-Y: 1          # 请求的vNPU数量，最大值为1。
                limits:
                  huawei.com/Ascend310P-Y: 1          # 数值与请求数量一致。
    ...
    ```
