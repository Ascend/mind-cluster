# Mounting vNPUs (Static Virtualization)<a name="ZH-CN_TOPIC_0000002479386388"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-07-01T06:20:25.612Z pushedAt=2026-07-01T06:25:34.110Z -->

## vNPU Usage Notes<a name="ZH-CN_TOPIC_0000002511426303"></a>

In Kubernetes scenarios, when you need to use vNPU resources, you must combine Ascend Device Plugin to enable Kubernetes to manage Ascend processor resources. Static virtualization scenarios cannot be mixed with dynamic virtualization. The cluster scheduling components required for the Ascend virtual instance feature are shown in the following table. For supported product models, see "Table 1 Supported products" in [Feature Description](../01_description.md).

**Table 1** Cluster scheduling components required for virtualization

<a name="table19103194217329"></a>
<table><thead align="left"><th class="cellrowborder" valign="top" width="11.677219849801206%" id="mcps1.2.5.1.1"><p id="p2103642143218"><a name="p2103642143218"></a><a name="p2103642143218"></a>Feature</p>
</th>
<th class="cellrowborder" valign="top" width="24.82697688116625%" id="mcps1.2.5.1.2"><p id="p619110456115"><a name="p619110456115"></a><a name="p619110456115"></a>Required Cluster Scheduling Component</p>
</th>
</thead>
<tbody><tr id="row61035425322"><td class="cellrowborder" rowspan="5" valign="top" width="11.677219849801206%" headers="mcps1.2.5.1.1 "><p id="p310384263219"><a name="p310384263219"></a><a name="p310384263219"></a>Static virtualization</p>
</td>
<td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p4191645116"><a name="p4191645116"></a><a name="p4191645116"></a><span id="ph1795411794410"><a name="ph1795411794410"></a><a name="ph1795411794410"></a>Ascend Docker Runtime</span></p>
</td>
</tr>
<tr id="row1844495022714"><td class="cellrowborder" valign="top" width="24.82697688116625%" headers="mcps1.2.5.1.2 "><p id="p4191645116"><a name="p4191645116"></a><a name="p4191645116"></a><span id="ph1795411794410"><a name="ph1795411794410"></a><a name="ph1795411794410"></a>Ascend Device Plugin</span></p>
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
</tbody>
</table>

>[!NOTE]
>In the static virtualization scenario, the optionality of components is described as follows.
>
>- Volcano: If you use your own scheduling component, parameter configuration is required. See [Table 3](#table1064314568229). You can also directly use this component for job scheduling.
>- Ascend Operator: This component is required only when training series products are used; it is optional when inference series products are supported.
>- ClusterD: This component is required only when Volcano is used. For details, see [Installing Volcano](../../../../developer_guide/installation_deployment/manual_installation/05_volcano.md).

## Static Virtualization<a name="ZH-CN_TOPIC_0000002479226392"></a>

**Restrictions<a name="section785220396317"></a>**

- Uninstalling Volcano is not supported while a job is running.
- Currently, the rules for the number of NPU devices requested per job Pod are as follows:

    If partitioned vNPUs are used, only 1 is supported.

- In static virtualization scenarios, if you create or destroy a vNPU, you need to restart Ascend Device Plugin.
- Static virtualization tasks do not support fault rescheduling.
- Switch affinity scheduling is not supported for static vNPUs.
- Static vNPU scheduling does not currently support fields related to `ASCEND_VISIBLE_DEVICES`. If the following fields exist, delete them:

```yaml
...
                env:
                - name: ASCEND_VISIBLE_DEVICES
                  valueFrom:
                    fieldRef:
                      fieldPath: metadata.annotations['huawei.com/Ascend310P']
...
```

**Table 2** Virtual instance templates and vNPU types

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
2. See the [Installation and Deployment](../../../../developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md) section to complete the installation of each component.

    The cluster scheduling components involved in modifying relevant parameters for virtual instances are Volcano and Ascend Device Plugin. Please modify and use the corresponding YAML for installation and deployment as required below:

    - Affinity scenario: Volcano must be installed.
    - Non-affinity scenario: Volcano does not need to be installed; only the device count is reported to the node's K8s.

    1. Ascend Device Plugin parameter modification and startup instructions:

        Virtual instance startup parameter description:

        **Table 3** Ascend Device Plugin startup parameters

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

1. Volcano parameter modification and startup instructions:

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

    The resource type specified in `requests` and `limits` under `resources` should be changed to `huawei.com/Ascend910-_Y_`, where the <i>Y</i> value is related to the vNPU type. For specific values, refer to the "vNPU Type" column in [Table 2 Virtual instance templates and vNPU types](#table47415104403).

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

The resource type specified in `requests` and `limits` under `resources` should be changed to `huawei.com/Ascend310P-_Y_`, where the <i>Y</i> value is related to the vNPU type. For specific values, refer to the "vNPU Type" column in [Table 2 Virtual instance templates and vNPU types](#table47415104403).

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend310P-Y: 1          # Number of vNPUs requested. The maximum value is 1.
                limits:
                  huawei.com/Ascend310P-Y: 1          # The value must match the requested quantity.
    ...
    ```
