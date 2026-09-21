# Mounting vNPUs (Static Virtualization)<a name="ZH-CN_TOPIC_0000002479386388"></a>

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-28T09:22:25.753Z pushedAt=2026-08-28T09:25:39.049Z -->

## Method 1: Mounting vNPUs Based on Native Docker

In the native Docker scenario (where MindCluster cluster scheduling components are not deployed), you need to use the npu-smi tool to create vNPUs and then mount the vNPUs into containers.

## Method 2: Mounting vNPUs Using Ascend Docker Runtime

Used alone with Ascend Docker Runtime (a container engine plugin) to mount vNPUs into containers.

### Prerequisites

Install Ascend Docker Runtime by referring to [Installing Ascend Docker Runtime](../../../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md).

### Usage

Assume you have created a vNPU using the npu-smi tool. When starting a container, run the following command to mount it into the container at container startup. The example below mounts the vNPU with ID 100.

```shell
docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL {image-name:tag} /bin/bash
```

>[!NOTE]
>
>- The available chip IDs can be queried and confirmed as follows:
>   - Physical chip ID:
>
>      ```shell
>      ls /dev/davinci*
>      ```
>
>   - Virtual chip ID:
>
>     ```shell
>     ls /dev/vdavinci*
>     ```
>
>- `image-name:tag`: image name and tag. Modify them based on the actual situation.
>- During use, do not repeatedly define the `ASCEND_VISIBLE_DEVICES` and `ASCEND_RUNTIME_OPTIONS` environment variables or fix them in the container image.

**Table 1**  Parameter description

|Parameter|Description|Example|
|--|--|--|
|`ASCEND_VISIBLE_DEVICES`|Used to specify the NPU devices to be mounted into the container. Otherwise, mounting the NPU devices fails. Use NPU device ID to specify devices. A single device, a range of devices, or a combination of both is supported. When using chip names to specify devices, multiple chip names of the same type can be specified at the same time.|<ul><li>`ASCEND_VISIBLE_DEVICES=100` indicates that vNPU 100 is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=101,103` indicates that vNPUs 101 and 103 are mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=100-102` indicates that vNPUs 100 to 102 (including 100 and 102) are mounted into the container, which has the same effect as `ASCEND_VISIBLE_DEVICES=100,101,102`.</li><li>`ASCEND_VISIBLE_DEVICES=100-102,104` indicates that vNPUs 100 to 102 and vNPU 104 are mounted into the container, which has the same effect as `ASCEND_VISIBLE_DEVICES=100,101,102,104`.</li><li>`ASCEND_VISIBLE_DEVICES=XXX-Y`, where `XXX` indicates the NPU device and the supported values are `npu`, `Ascend910`, `Ascend310`, `Ascend310B`, and `Ascend310P`; `Y` indicates the physical NPU device ID.<ul><li>`ASCEND_VISIBLE_DEVICES=npu-101` indicates that vNPU 101 is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=npu-101,npu-103` indicates that NPU 101 and vNPU 103 are mounted into the container.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When using chip names to specify devices, it is recommended to use `npu` uniformly.</li><li>Specifying both a device ID and an NPU name in one parameter is not supported. That is, `ASCEND_VISIBLE_DEVICES=101,npu-103` is not supported.</li><li>`ASCEND_RUNTIME_OPTIONS` must be used together, and its value must contain `VIRTUAL`, indicating that the mounted devices are vNPUs.</li></ul></div></div></li></ul>|
|`ASCEND_RUNTIME_OPTIONS`|<p>Restricts the chip IDs specified in `ASCEND_VISIBLE_DEVICES`:</p><ul><li>`NODRV`: indicates that driver-related directories are not mounted.</li><li>`VIRTUAL`: indicates that the mounted devices are virtual chips.</li><li>`NODRV,VIRTUAL`: indicates that the mounted devices are virtual chips and driver-related directories are not mounted.</li></ul>|<ul><li>`ASCEND_RUNTIME_OPTIONS=NODRV`</li><li>`ASCEND_RUNTIME_OPTIONS=VIRTUAL`</li><li>`ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>In the static virtualization scenario, `ASCEND_RUNTIME_OPTIONS` is mandatory, and its value must contain `VIRTUAL`.</li></ul></div></div>|

## Method 3: Mounting vNPUs on Kubernetes

### Usage<a name="ZH-CN_TOPIC_0000002511426303"></a>

In Kubernetes scenarios, when you need to use vNPU resources, use Ascend Device Plugin together so that Kubernetes can manage Ascend processor resources. In static virtualization scenarios, it cannot be mixed with dynamic virtualization. The cluster scheduling components required by the Ascend virtual instance feature are as follows. For supported products, see "Table 1 Supported product" in [Feature Description](../01_description.md).

- Ascend Docker Runtime
- Ascend Device Plugin
- (Optional) Volcano
- (Optional) Ascend Operator
- (Optional) ClusterD

>[!NOTE]
>For optional components in static virtualization scenarios:
>
>- Volcano: If you use your own scheduling component, you need to configure parameters. For details, see [Table 3](#table1064314568229). You can also directly use this component for task scheduling.
>- Ascend Operator: This component needs to be selected only when training products are used; it can be omitted when inference products are used.
>- ClusterD: This component needs to be selected only when Volcano is used. For details, see [Installing Volcano](../../../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md).

### Usage Restrictions<a name="ZH-CN_TOPIC_0000002479226392"></a>

- Uninstalling Volcano is not supported while a task is running.
- Currently, the rules for the number of NPU devices requested by each Pod of a task are as follows:

    If a split vNPU is used, only one is supported.

- In static virtualization scenarios, if a vNPU is created or destroyed, Ascend Device Plugin must be restarted.
- Static virtualization tasks do not support fault rescheduling.
- Switch affinity scheduling is not supported for static vNPUs.
- Static vNPU scheduling does not support fields related to `ASCEND_VISIBLE_DEVICES`. If the following fields exist, delete them:

  ```yaml
  ...
                  env:
                  - name: ASCEND_VISIBLE_DEVICES
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend310P']
  ...
  ```

**Table 2**  Relationship between virtual instance templates and vNPU types

<a name="table47415104403"></a>
<table><thead align="left"><tr id="row67416101402"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p117491014400"><a name="p117491014400"></a><a name="p117491014400"></a>NPU Type</p>
</th>
<th class="cellrowborder" valign="top" width="19.96%" id="mcps1.2.5.1.2"><p id="p177431064013"><a name="p177431064013"></a><a name="p177431064013"></a>Virtual Instance Template</p>
</th>
<th class="cellrowborder" valign="top" width="20.04%" id="mcps1.2.5.1.3"><p id="p1374210134015"><a name="p1374210134015"></a><a name="p1374210134015"></a>vNPU Type</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p1041963771317"><a name="p1041963771317"></a><a name="p1041963771317"></a>Specific virtual device name (using vNPU ID 100 and physical chip ID 0 as an example)</p>
</th>
</tr>
</thead>
<tbody><tr id="row5741710164014"><td class="cellrowborder" rowspan="4" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p074181014408"><a name="p074181014408"></a><a name="p074181014408"></a><span id="ph327965117217"><a name="ph327965117217"></a><a name="ph327965117217"></a><term>Atlas training products</term></span> (30 or 32 AICores)</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span><term>Atlas A2 training products</term></span> (24 AICores)</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span><term>Atlas A3 training products</term></span> (48 AICores)</p>
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
<tr id="row84911853114212"><td class="cellrowborder" rowspan="7" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1868751772016"><a name="p1868751772016"></a><a name="p1868751772016"></a><span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a><term>Atlas inference products</term></span> (8 AICores)</p>
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
<tr><td rowspan="6" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span><term>Atlas A2 inference products</term></span> (20 AICores)</p>
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
<tr><td rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p><span><term>Atlas A3 inference products</term></span> (40 AICores)</p>
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

### Prerequisites

1. Obtain `Ascend-docker-runtime_{version}_linux-{arch}.run` first, and install Ascend Docker Runtime.
2. See [Installation and Deployment](../../../../03_installation_guide/02_installation/00_helm_installation.md) to complete the installation of each component.

    For virtual instances, the cluster scheduling components that require modification of related parameters are Volcano and Ascend Device Plugin. Modify and use the corresponding YAML for installation and deployment as follows:

    - Affinity scenario: Volcano needs to be installed.
    - Non-affinity scenario: Volcano is not required. Only the device quantity is reported to the K8s of the node.

    1. Ascend Device Plugin parameter modification and startup description:

        The startup parameters of the virtual instances are described as follows:

        **Table 3** Ascend Device Plugin startup parameters

        <a name="table1064314568229"></a>

        |Name|Type|Default Value|Description|
        |--|--|--|--|
        |`-volcanoType`|bool|false|Whether to use Volcano for scheduling. If dynamic virtualization is used, set this parameter to `true`.|
        |`-presetVirtualDevice`|bool|true|Switch for static virtualization. The value can only be `true`.<p>If dynamic virtualization is used, set this parameter to `false` and enable Volcano at the same time, that is, set the `-volcanoType` parameter to `true`.</p>|

        The YAML startup is as follows:

        - The K8s cluster contains nodes of <term>Atlas inference products</term>, <term>Atlas training products</term>, <term>Atlas A2 training products</term>, <term>Atlas A3 training products</term>, <term>Atlas A2 inference products</term>, and <term>Atlas A3 inference products</term>(Ascend Device Plugin works independently, without Volcano and Ascend Operator).

            ```shell
            kubectl apply -f device-plugin-v{version}.yaml
            ```

        - The K8s cluster contains nodes of <term>Atlas inference products</term>, <term>Atlas training products</term>, <term>Atlas A2 training products</term>, <term>Atlas A3 training products</term>, <term>Atlas A2 inference products</term>, and <term>Atlas A3 inference products</term> (used with Volcano and Ascend Operator, supporting NPU virtualization. Dynamic virtualization is disabled by default in the YAML).

            ```shell
            kubectl apply -f device-plugin-volcano-v{version}.yaml
            ```

        If the K8s cluster uses multiple types of Ascend AI Processors, run the corresponding commands separately.

    2. Volcano parameter modification and startup instructions:

        In the Volcano deployment file `volcano-v{version}.yaml`, set `presetVirtualDevice` to `true`.

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
              - name: volcano-npu-v26.1.0_linux-aarch64    # Where 26.1.0 is the MindCluster version number, and the value varies depending on the version.
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

### Usage

- When creating a training job, modify the following configurations when creating the YAML file. The following uses <term>Atlas training products</term> as an example.

    The resource types specified for `requests` and `limits` in `resources` should be changed to `huawei.com/Ascend910-_Y_`, where the `Y` value is related to the vNPU type. For details about the value, see the "vNPU Type" column in [Table 2 Relationship between virtual instance templates and vNPU types](#table47415104403).

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend910-Y: 1          # Number of requested vNPUs. The maximum value is 1.
                limits:
                  huawei.com/Ascend910-Y: 1          # The value is the same as the requested number.
    ...
    ```

- When creating an inference job, modify the following configuration when creating the YAML file. The following uses the <term>Atlas inference products</term> as an example.

    The `requests` and `limits` resource types set in `resources` should be changed to `huawei.com/Ascend310P-_Y_`, where the `Y` value is related to the vNPU type. For the specific value, refer to the "vNPU Type" column in [Table 2 Relationship between virtual instance templates and vNPU types](#table47415104403).

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend310P-Y: 1          # Number of requested vNPUs. The maximum value is 1.
                limits:
                  huawei.com/Ascend310P-Y: 1          # The value is consistent with the requested quantity.
    ...
    ```
