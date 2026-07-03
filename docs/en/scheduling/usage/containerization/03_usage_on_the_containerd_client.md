# Using Ascend Docker Runtime in the Containerd Client<a name="ZH-CN_TOPIC_0000002511347203"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:22:01.695Z pushedAt=2026-06-30T12:23:24.389Z -->

## Usage Instructions<a name="section0966931165317"></a>

- Ascend Docker Runtime supports mounting physical chips and virtual chips. Before mounting virtual chips, refer to the [Creating vNPU](../virtual_instance/virtual_instance_with_hdk/static_vnpu_scheduling/01_creating_vnpu.md) section to virtualize physical chips. Both static virtualization and dynamic virtualization of physical chips are supported.
- You can query the currently available physical chip IDs using the <b>ls /dev/davinci\*</b> command, and query the currently available virtual chip IDs using the <b>ls /dev/vdavinci\*</b> command.
- If you do not need to mount all content from the default Ascend Docker Runtime configuration file "/etc/ascend-docker-runtime.d/base.list", you can create a custom configuration file (for example, hostlog.list) to reduce the mounted content. For details, refer to the [(Optional) Configuring Custom Mounted Content](./01_configuring_custom_mounted_content.md) section.

## Usage Examples<a name="section148905517122"></a>

The `image-name:tag` in the examples represents the image name and tag, such as "ascend-pytorch:pytorch\_TAG". `containerID` is the container ID. When using ctr to start a container, you must specify a container ID, such as "c1".

- Example 1: Mount the physical chip with chip ID 0 when starting the container.

    ```shell
    ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 {image-name:tag} {containerID} bash
    ```

- Example 2: Mount only NPU and management devices when starting a container. Do not mount driver-related directories.

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_OPTIONS=NODRV {image-name:tag} {containerID} bash
    ```

- Example 3: Mount the physical chip with chip ID 0 when starting the container, and read the mount content from the custom configuration file `hostlog`.

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_MOUNTS=hostlog {image-name:tag} {containerID} bash
    ```

- Example 4: When starting a container, mount the virtual chip with chip ID 100.

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=100 --env ASCEND_RUNTIME_OPTIONS=VIRTUAL {image-name:tag} {containerID} bash
    ```

- Example 5: When starting a container, slice 4 AICores from the physical chip with chip ID 0 as virtual devices and mount them into the container.

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_VNPU_SPECS=vir04 {image-name:tag} {containerID} bash
    ```

- Example 6: When starting a container, mount the physical chip with chip ID 0, and allow soft links in the mounted driver files (applicable only to Atlas 500 A2 Intelligent Station, Atlas 200I A2 Acceleration Module, and Atlas 200I DK A2 Developer Kit).

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True {image-name:tag} {containerID} bash
    ```

The parameters related to the startup command are shown in [Table 1](#table5134121862415).

After the container is started, you can run the following commands to check whether the corresponding devices and drivers are mounted successfully. For the specific mount directories of each model, refer to [Content Mounted by Ascend Docker Runtime](../../references/appendix.md#content-mounted-by-ascend-docker-runtime). The command example is as follows:

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

Possible output results are as follows:

```ColdFusion
davinci0
davinci_manager
devmm_svm
hisi_hdc
include lib64
dcmi
npu-smi
```

>[!NOTE]
>During use, do not redefine or fix environment variables such as `ASCEND_VISIBLE_DEVICES`, `ASCEND_RUNTIME_OPTIONS`, `ASCEND_RUNTIME_MOUNTS`, and `ASCEND_VNPU_SPECS` in the container image.

**Table 1** Ascend Docker Runtime parameters

<a name="table5134121862415"></a>

|Parameter|Description|Example|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>If the job does not require an NPU device, you can set the ASCEND_VISIBLE_DEVICES environment variable to void or leave it empty.</li><li>If the job requires an NPU device, you must use ASCEND_VISIBLE_DEVICES to specify the NPU device to be mounted into the container; otherwise, the NPU device mount will fail. When specifying devices by device index, single and range specifications are supported and can be mixed. When specifying devices by chip name, multiple chip names of the same type can be specified simultaneously.</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void indicates that the mount function of Ascend Docker Runtime is not used, and no NPU devices, drivers, or file directories are mounted. Related mount parameters will also become invalid.</li><li>Mount Physical chip (NPU)<ul><li>ASCEND_VISIBLE_DEVICES=0 indicates that NPU device 0 (/dev/davinci0) is mounted into the container.</li><li>ASCEND_VISIBLE_DEVICES=1,3 indicates that NPU devices 1 and 3 are mounted into the container.</li><li>ASCEND_VISIBLE_DEVICES=0-2 indicates that NPU devices 0 through 2 (including 0 and 2) are mounted into the container, with the same effect as -e ASCEND_VISIBLE_DEVICES=0,1,2.</li><li>ASCEND_VISIBLE_DEVICES=0-2,4 indicates that NPU devices 0 through 2 and device 4 are mounted into the container, with the same effect as -e ASCEND_VISIBLE_DEVICES=0,1,2,4.</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y, where XXX represents the NPU device, with supported values being npu, Ascend910, Ascend310, Ascend310B, and Ascend310P; Y represents the physical NPU device ID.<ul><li>ASCEND_VISIBLE_DEVICES=npu-1 indicates that NPU device 1 is mounted into the container.</li><li>ASCEND_VISIBLE_DEVICES=npu-1,npu-3 indicates that NPU 1 and NPU 3 are mounted into the container.</li></ul></li></ul><div class="note"><span class="notetitle">[!NOTE] Note</span><div class="notebody"><ul><li>When specifying devices by chip name, it is recommended to uniformly use the value npu.</li><li>Specifying both a device index and an NPU name in a single parameter is not supported, i.e., ASCEND_VISIBLE_DEVICES=0,npu-1 is not supported.</li></ul></div></div></li><li>Mount Virtual chip (vNPU)<ul><li>**Static virtualization**: The usage is the same as for physical chips; simply replace the physical chip ID with the virtual chip ID (vNPU ID).</li><li>**Dynamic virtualization**: ASCEND_VISIBLE_DEVICES=0 indicates that a certain number of AICores are partitioned from NPU device 0.<div class="note"><span class="notetitle">[!NOTE] Note</span><div class="notebody"><ul><li>A single dynamic virtualization command can only specify the ID of one physical NPU for dynamic virtualization.</li><li>Must be used together with ASCEND_VNPU_SPECS, which indicates the number of AICores partitioned on the specified NPU.</li><li>Can be used together with ASCEND_RUNTIME_OPTIONS, but the value can only be NODRV, indicating that driver-related directories are not mounted.</li></ul></div></div></li></ul></li></ul>|
|ASCEND_ALLOW_LINK|Whether to allow soft links in the mounted files or directories. This parameter needs to be specified in scenarios involving the Atlas 500 A2 Intelligent Station, Atlas 200I A2 AI Acceleration Module, and Atlas 200I DK A2 Developer Kit.<p>Other devices, such as Atlas training series products, <term>Atlas A2 training series products</term>, and Atlas 200I SoC A1 core board, can use this parameter. However, because their default mount content does not contain soft links, there is no need to specify this parameter additionally.</p>|<ul><li>ASCEND_ALLOW_LINK=True indicates that mounting driver files with soft links is allowed in scenarios involving the Atlas 500 A2 Intelligent Station, Atlas 200I A2 AI Acceleration Module, and Atlas 200I DK A2 Developer Kit.</li><li>If ASCEND_ALLOW_LINK=False or this parameter is not specified, the Atlas 500 A2 Intelligent Station, Atlas 200I A2 AI Acceleration Module, and Atlas 200I DK A2 Developer Kit will be unable to use Ascend Docker Runtime.</li></ul>|
|ASCEND_RUNTIME_OPTIONS|Restricts the chip ID specified in the ASCEND_VISIBLE_DEVICES parameter:<ul><li>NODRV: Indicates that driver-related directories are not mounted.</li><li>VIRTUAL: Indicates that the mounted chip is a virtual chip.</li><li>NODRV,VIRTUAL: Indicates that the mounted chip is a virtual chip, and driver-related directories are not mounted.</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul><div class="note"><span class="notetitle">Note:</span><div class="notebody"><ul><li>In static virtualization scenarios, ASCEND_RUNTIME_OPTIONS is a mandatory parameter, and its value must include VIRTUAL.</li><li>In dynamic virtualization scenarios, if the ASCEND_RUNTIME_OPTIONS parameter is used, its value cannot include VIRTUAL.</li></ul></div></div>|
|ASCEND_RUNTIME_MOUNTS|The configuration file name for the content to be mounted. This file can configure the files and directories to be mounted into the container.|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2<div class="note"><span class="notetitle">Note:</span><div class="notebody"><ul><li>By default, the /etc/ascend-docker-runtime.d/base.list configuration file is read.</li><li>For hostlog.list, modify it according to the actual custom configuration file name.</li><li>Reading multiple custom configuration files is supported.</li><li>File names must be lowercase and cannot contain uppercase letters.</li></ul></div></div></li></ul>|
|ASCEND_VNPU_SPECS|Partitions a certain number of AICores from a physical NPU device and specifies them as a virtual device. For supported values, see the "Virtualization Instance Template" column in Table 1 of [Virtualization Templates](../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md).<ul><li>This parameter can only be used for product forms that support dynamic virtualization.</li><li>Must be used together with the "ASCEND_VISIBLE_DEVICES" parameter, which specifies the physical NPU device used for virtualization.</li></ul>|ASCEND_VNPU_SPECS=vir04 indicates that 4 AICores are partitioned as a virtual device and mounted into the container.|
