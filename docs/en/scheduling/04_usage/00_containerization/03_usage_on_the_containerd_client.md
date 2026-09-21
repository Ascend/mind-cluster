# Using Ascend Docker Runtime in the Containerd Client<a name="ZH-CN_TOPIC_0000002511347203"></a>

## Usage Instructions<a name="section0966931165317"></a>

- Ascend Docker Runtime supports mounting physical chips and virtual chips. Before mounting virtual chips, refer to the [Creating vNPU](../02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md) section to virtualize physical chips. Both static virtualization and dynamic virtualization of physical chips are supported.
- You can query the currently available physical chip IDs using the <b>ls /dev/davinci\*</b> command, and query the currently available virtual chip IDs using the <b>ls /dev/vdavinci\*</b> command.
- If you do not need to mount all content from the default Ascend Docker Runtime configuration file `/etc/ascend-docker-runtime.d/base.list`, you can create a custom configuration file (for example, `hostlog.list`) to reduce the mounted content. For details, refer to the [(Optional) Configuring Custom Mounted Content](./01_configuring_custom_mounted_content.md) section.

## Usage Examples<a name="section148905517122"></a>

The `{image-name:tag}` in the examples represents the image name and tag, such as "ascend-pytorch:pytorch\_TAG". `containerID` is the container ID. When using ctr to start a container, you must specify a container ID, such as "c1".

- Example 1: Mount the physical chip with chip ID 0 when starting the container.

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 {image-name:tag} {containerID} bash
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

- Example 6: When starting a container, mount the physical chip with chip ID 0, and allow soft links in the mounted driver files (applicable only to Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, and Atlas 200I DK A2 Developer Kit).

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True {image-name:tag} {containerID} bash
    ```

The parameters related to the startup command are shown in [Table 1](#table5134121862415).

After the container is started, you can run the following commands to check whether the corresponding devices and drivers are mounted successfully. For the specific mount directories of each model, refer to [Content Mounted by Ascend Docker Runtime](../../07_references/05_appendix.md#content-mounted-by-ascend-docker-runtime). The command example is as follows:

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
| `ASCEND_VISIBLE_DEVICES` | <ul><li>If the task does not require NPU devices, set `ASCEND_VISIBLE_DEVICES` to `void` or leave it empty.</li><li>If the task requires NPU devices, you must use `ASCEND_VISIBLE_DEVICES` to specify the NPU devices to be mounted into the container; otherwise, NPU device mounting will fail. When specifying devices by device ID, one or more devices can be specified (mixed specification is supported). When specifying devices by chip name, multiple chip names of the same type can be specified simultaneously.</li></ul> | <ul><li>`ASCEND_VISIBLE_DEVICES=void` means the Ascend Docker Runtime mounting feature is not used, and NPU devices, drivers, and file directories are not mounted. Related mount parameters will also become ineffective.</li>\<li>Mounting physical chips (NPUs)<ul><li>`ASCEND_VISIBLE_DEVICES=0` means NPU device 0 (`/dev/davinci0`) is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=1,3` means NPU devices 1 and 3 are mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=0-2` means NPU devices 0 through 2 (inclusive) are mounted, equivalent to `-e ASCEND_VISIBLE_DEVICES=0,1,2`.</li><li>`ASCEND_VISIBLE_DEVICES=0-2,4` means NPU devices 0 through 2 and device 4 are mounted, equivalent to `-e ASCEND_VISIBLE_DEVICES=0,1,2,4`.</li><li>`ASCEND_VISIBLE_DEVICES=XXX-Y`, where `XXX` represents the NPU device type (supported values: `npu`, `Ascend910`, `Ascend310`, `Ascend310B`, `Ascend310P`) and `Y` represents the physical NPU device ID.<ul><li>`ASCEND_VISIBLE_DEVICES=npu-1` means NPU device 1 is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=npu-1,npu-3` means NPU devices 1 and 3 are mounted into the container.</li></ul></li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When specifying devices by chip name, it is recommended to use `npu` uniformly.</li><li>Mixing device IDs and NPU names in the same parameter is not supported. For example, `ASCEND_VISIBLE_DEVICES=0,npu-1` is not supported.</li></ul></div></div><li>Mounting virtual chips (vNPU)<ul><li>**Static virtualization**: The usage is the same as for physical chips; simply replace the physical chip ID with the virtual chip ID (vNPU ID).</li><li>**Dynamic virtualization**: `ASCEND_VISIBLE_DEVICES=0` means allocating a certain number of AICores from NPU device 0.<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>A single dynamic virtualization command can only specify one physical NPU ID for dynamic virtualization.</li><li>`ASCEND_VNPU_SPECS` must be used together with this parameter to specify the number of AICores to allocate from the specified NPU.</li><li>`ASCEND_RUNTIME_OPTIONS` can also be used with the value `NODRV`, meaning driver-related directories are not mounted.</li></ul></div></div></li></ul></li></ul> |
| `ASCEND_ALLOW_LINK` | Whether to allow symbolic links in mounted files or directories. This parameter must be specified for Atlas 500 A2 intelligent station, Atlas 200I A2 AI accelerator module, and Atlas 200I DK A2. <p>Other devices such as <term>Atlas training products</term>, <term>Atlas A2 training products</term>, and Atlas 200I SoC A1 core board can also use this parameter, but since their default mounted contents do not contain symbolic links, it is not necessary to specify this parameter.</p> | <ul><li>`ASCEND_ALLOW_LINK=True` allows mounting driver files with symbolic links on Atlas 500 A2 intelligent station, Atlas 200I A2 AI accelerator module, and Atlas 200I DK A2.</li><li>`ASCEND_ALLOW_LINK=False` or omitting this parameter means Atlas 500 A2 intelligent station, Atlas 200I A2 AI accelerator module, and Atlas 200I DK A2 cannot use Ascend Docker Runtime.</li></ul> |
| `ASCEND_RUNTIME_OPTIONS` | Restricts the chip ID specified in `ASCEND_VISIBLE_DEVICES`: <ul><li>`NODRV`: Driver-related directories are not mounted.</li><li>`VIRTUAL`: A virtual chip is mounted.</li><li>`NODRV,VIRTUAL`: A virtual chip is mounted, and driver-related directories are not mounted.</li></ul> | <ul><li>`ASCEND_RUNTIME_OPTIONS=NODRV`</li><li>`ASCEND_RUNTIME_OPTIONS=VIRTUAL`</li><li>`ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>In static virtualization scenarios, `ASCEND_RUNTIME_OPTIONS` is mandatory and must include `VIRTUAL`.</li><li>In dynamic virtualization scenarios, if `ASCEND_RUNTIME_OPTIONS` is used, its value cannot include `VIRTUAL`.</li></ul></div></div> |
| `ASCEND_RUNTIME_MOUNTS` | The configuration file name(s) for content to be mounted, which lists files and directories to be mounted into the container. | <ul><li>`ASCEND_RUNTIME_MOUNTS=base`</li><li>`ASCEND_RUNTIME_MOUNTS=hostlog`</li><li>`ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>By default, the configuration file `/etc/ascend-docker-runtime.d/base.list` is read.</li><li>Modify `hostlog.list` according to the actual custom configuration file name.</li><li>Multiple custom configuration files can be read.</li><li>File names must be lowercase and cannot contain uppercase letters.</li></ul></div></div> |
| `ASCEND_VNPU_SPECS` | Allocates a certain number of AICores from a physical NPU device and designates them as virtual devices. For supported values, see the "Virtualization Instance Template" column in Table 1 of [Virtualization Templates](../02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md). <ul><li>This parameter can only be used on product models that support dynamic virtualization.</li><li>Must be used together with `ASCEND_VISIBLE_DEVICES`, which specifies the physical NPU device for virtualization.</li></ul> | `ASCEND_VNPU_SPECS=vir04` means allocating 4 AICores as a virtual device and mounting it into the container. |
