# Usage on the Docker Client<a name="ZH-CN_TOPIC_0000002479387248"></a>

## Usage Instructions<a name="section0966931165317"></a>

- Ascend Docker Runtime supports mounting physical and virtual chips. Before mounting virtual chips, refer to [Creating vNPUs](../02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md) to virtualize physical chips. Both static virtualization and dynamic virtualization of physical chips are supported.
- You can query the currently available physical chip IDs by running the `ls /dev/davinci*` command, and query the currently available virtual chip IDs by running the `ls /dev/vdavinci*` command.
- If you do not need to mount all the content in the default configuration file `/etc/ascend-docker-runtime.d/base.list` of Ascend Docker Runtime, create a custom configuration file (for example, `hostlog.list`) to reduce the mounted content. For details, see [(Optional) Configuring Custom Mounted Content](./01_configuring_custom_mounted_content.md).

## Mounting Chips Using Ascend Docker Runtime<a name="section11917171014591"></a>

In the examples, `{image-name:tag}` represents the image name and tag. For details about other parameters, see [Table 1](#table3488191614328).

- Example 1: Mount the physical chip with chip ID 0 when starting the container.

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 {image-name:tag} /bin/bash
    ```

- Example 2: Mount only the NPU device and management device when starting the container, without mounting driver-related directories.

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_OPTIONS=NODRV {image-name:tag} /bin/bash
    ```

- Example 3: Mount the physical chip with chip ID 0 when starting the container, and read the mount content from the custom configuration file `hostlog.list`.

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog {image-name:tag} /bin/bash
    ```

- Example 4: Mount the chip with virtual chip ID 100 when starting the container.

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL {image-name:tag} /bin/bash
    ```

- Example 5: When starting the container, slice 4 AICores from the chip with physical chip ID 0 as virtual devices and mount them to the container.

    ```shell
    docker run -it --rm -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_VNPU_SPECS=vir04 {image-name:tag} /bin/bash
    ```

- Example 6: When starting the container, mount the chip with physical chip ID 0, and allow soft links in the mounted driver files (applicable only to Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, and Atlas 200I DK A2):

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True {image-name:tag} /bin/bash
    ```

After the container is started, run the following commands inside and outside the container to check whether the corresponding devices and drivers are mounted successfully. For the specific mount directory of each model, refer to [Content Mounted by Ascend Docker Runtime](../../07_references/05_appendix.md#content-mounted-by-ascend-docker-runtime). Example commands are as follows:

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

Possible outputs:

```ColdFusion
davinci0
davinci_manager
devmm_svm
hisi_hdc
include lib64
dcmi
npu-smi
```

## Using Ascend Docker Runtime to Mount Chips and Other Devices<a name="section111912299472"></a>

Use Ascend Docker Runtime to support containers running training, inference, or other tasks.

- Taking the Atlas 200I SoC A1 core board running an inference container as an example, modify the command according to the actual situation. The example is as follows, and the related parameters are shown in [Table 1](#table3488191614328) and [Table 2](#table46513386334).

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/vpc:rwm --device=/dev/log_drv:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >**NOTE**
    >- If the driver of the Atlas 200I SoC A1 core board is version 1.0.0 (Ascend HDK 22.0.0) or earlier, you need to mount `/dev/xsmem_dev` and `/dev/event_sched`.
    >- If the driver of the Atlas 200I SoC A1 core board is later than version 1.0.0 (Ascend HDK 22.0.0), you do not need to mount  `/dev/xsmem_dev` and `/dev/event_sched`.

- Taking the Atlas 500 A2 intelligent station running an inference container as an example, modify the command according to the actual situation. The example is as follows, and the related parameters are shown in [Table 1](#table3488191614328) and [Table 2](#table46513386334).

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True workload-image:v1.0 /bin/bash
    ```

## Mounting Chips and Other Devices Without Ascend Docker Runtime<a name="section1212516610490"></a>

Taking the Atlas 200I SoC A1 core board running an inference container as an example, modify the command according to the actual situation. The example is as follows, and the related parameters are shown in [Table 2](#table46513386334).

```shell
docker run -it --device=/dev/davinci0:rwm --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/venc:rwm --device=/dev/vpc:rwm --device=/dev/davinci_manager:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /etc/sys_version.conf:/etc/sys_version.conf:ro -v /usr/local/bin/npu-smi:/usr/local/bin/npu-smi:ro -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /etc/hdcBasic.cfg:/etc/hdcBasic.cfg:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
```

>[!NOTE]
>
>- If the driver of the Atlas 200I SoC A1 core board is version 1.0.0 (Ascend HDK 22.0.0) or earlier, you need to mount `/dev/xsmem_dev` and `/dev/event_sched`.
>- If the driver of the Atlas 200I SoC A1 core board is a version later than 1.0.0 (Ascend HDK 22.0.0), you do not need to mount `/dev/xsmem_dev` and `/dev/event_sched`.

## Parameter Description<a name="section131432039144912"></a>

**Table 1** Ascend Docker Runtime running parameters

<a name="table3488191614328"></a>

|Parameter|Description|Example|
|--|--|--|
| `ASCEND_VISIBLE_DEVICES` | <ul><li>If the task does not require NPU devices, set `ASCEND_VISIBLE_DEVICES` to `void` or leave it empty.</li><li>If the task requires NPU devices, you must use `ASCEND_VISIBLE_DEVICES` to specify the NPU devices to be mounted into the container; otherwise, NPU device mounting will fail. When specifying devices by device ID, one or more devices can be specified (mixing specification is supported). When specifying devices by chip name, multiple chip names of the same type can be specified simultaneously.</li></ul> | <ul><li>`ASCEND_VISIBLE_DEVICES=void` means the Ascend Docker Runtime mounting feature is not used, and NPU devices, drivers, and file directories are not mounted. Related mount parameters will also become ineffective.</li><li>Mounting physical chips (NPUs)</li><ul><li>`ASCEND_VISIBLE_DEVICES=0` means NPU device 0 (`/dev/davinci0`) is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=1,3` means NPU devices 1 and 3 are mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=0-2` means NPU devices 0 through 2 (inclusive) are mounted, equivalent to `-e ASCEND_VISIBLE_DEVICES=0,1,2`.</li><li>`ASCEND_VISIBLE_DEVICES=0-2,4` means NPU devices 0 through 2 and device 4 are mounted, equivalent to `-e ASCEND_VISIBLE_DEVICES=0,1,2,4`.</li><li>`ASCEND_VISIBLE_DEVICES=XXX-Y`, where `XXX` represents the NPU device type (supported values: `npu`, `Ascend910`, `Ascend310`, `Ascend310B`, `Ascend310P`) and `Y` represents the physical NPU device ID.</li><ul><li>`ASCEND_VISIBLE_DEVICES=npu-1` means NPU device 1 is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=npu-1,npu-3` means NPU devices 1 and 3 are mounted into the container.</li></ul></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When specifying devices by chip name, it is recommended to use `npu` uniformly.</li><li>Mixing device IDs and NPU names in the same parameter is not supported. For example, `ASCEND_VISIBLE_DEVICES=0,npu-1` is not supported.</li></ul></div></div><li>Mounting virtual chips (vNPU)<ul><li>**Static virtualization**: The usage is the same as for physical chips; simply replace the physical chip ID with the virtual chip ID (vNPU ID).</li><li>**Dynamic virtualization**: `ASCEND_VISIBLE_DEVICES=0` means allocating a certain number of AICores from NPU device 0.<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>A single dynamic virtualization command can only specify one physical NPU ID for.</li><li>`ASCEND_VNPU_SPECS` must be used together with this parameter to specify the number of AICores to allocate from the specified NPU.</li><li>`ASCEND_RUNTIME_OPTIONS` can also be used with the value `NODRV`, meaning driver-related directories are not mounted.</li></ul></div></div></li></ul></li></ul> |
| `ASCEND_ALLOW_LINK` | Whether to allow symbolic links in mounted files or directories. This parameter must be specified for Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, and Atlas 200I DK A2. <p>Other devices such as <term>Atlas training products</term>, <term>Atlas A2 training products</term>, and Atlas 200I SoC A1 core board can also use this parameter, but since their default mounted contents do not contain symbolic links, it is not necessary to specify this parameter.</p> | <ul><li>`ASCEND_ALLOW_LINK=True` allows mounting driver files with symbolic links on Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, and Atlas 200I DK A2.</li><li>`ASCEND_ALLOW_LINK=False` or omitting this parameter means Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, and Atlas 200I DK A2 cannot use Ascend Docker Runtime.</li></ul> |
| `ASCEND_RUNTIME_OPTIONS` | Restricts the chip ID specified in `ASCEND_VISIBLE_DEVICES`: <ul><li>`NODRV`: Driver-related directories are not mounted.</li><li>`VIRTUAL`: A virtual chip is mounted.</li><li>`NODRV,VIRTUAL`: A virtual chip is mounted, and driver-related directories are not mounted.</li></ul> | <ul><li>`ASCEND_RUNTIME_OPTIONS=NODRV`</li><li>`ASCEND_RUNTIME_OPTIONS=VIRTUAL`</li><li>`ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>In static virtualization scenarios, `ASCEND_RUNTIME_OPTIONS` is mandatory and must include `VIRTUAL`.</li><li>In dynamic virtualization scenarios, if `ASCEND_RUNTIME_OPTIONS` is used, its value cannot include `VIRTUAL`.</li></ul></div></div> |
| `ASCEND_RUNTIME_MOUNTS` | The configuration file name(s) for content to be mounted, which lists files and directories to be mounted into the container. | <ul><li>`ASCEND_RUNTIME_MOUNTS=base`</li><li>`ASCEND_RUNTIME_MOUNTS=hostlog`</li><li>`ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>By default, the configuration file `/etc/ascend-docker-runtime.d/base.list` is read.</li><li>Modify `hostlog.list` according to the actual custom configuration file name.</li><li>Multiple custom configuration files can be read.</li><li>File names must be lowercase and cannot contain uppercase letters; file names with uppercase letters may cause the configuration file to fail.</li></ul></div></div> |
| `ASCEND_VNPU_SPECS` | Allocates a certain number of AICores from a physical NPU device and designates them as virtual devices. For supported values, see the "Virtualization Instance Template" column in Table 1 of [Virtualization Templates](../02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md). <ul><li>This parameter can only be used on product models that support dynamic virtualization.</li><li>Must be used together with `ASCEND_VISIBLE_DEVICES`, which specifies the physical NPU device for virtualization.</li><li>When `ASCEND_RUNTIME_OPTIONS` includes `VIRTUAL`, `ASCEND_VNPU_SPECS` will no longer take effect.</li></ul> | `ASCEND_VNPU_SPECS=vir04` means allocating 4 AICores as a virtual device and mounting it into the container. |

**Table 2** Explanation of other parameters

<a name="table46513386334"></a>

|Parameter|Description|
|--|--|
| `/dev/xsmem_dev` | Mounts the memory device management into the container. |
| `/dev/event_sched` | Mounts the event scheduling device into the container. |
| `/dev/ts_aisle` | Mounts the device corresponding to `aicpudrv` into the container. |
| `/dev/svm0` | Mounts the memory management device into the container. |
| `/dev/sys` | Mounts DVPP-related devices into the container. |
| `/dev/vdec` | Mounts DVPP-related devices into the container. |
| `/dev/vpc` | Mounts DVPP-related devices into the container. |
| `/dev/log_drv` | Mounts log recording-related devices into the container. |
| `/dev/upgrade` | Mounts devices for obtaining Ascend system configuration and firmware into the container. |
| `/dev/spi_smbus` | Mounts devices related to out-of-band SPI communication into the container. |
| `/dev/user_config` | Mounts devices for managing user configuration into the container. |
| `/dev/memory_bandwidth` | Mounts memory bandwidth-related devices into the container. |
| `-v /var/slogd:/var/slogd` | Mounts the host log process files into the container in read-only mode. |
| `-v /var/dmp_daemon:/var/dmp_daemon` | Mounts the dmp daemon into the container. |
| `-v /var/log/npu/conf/slog:/var/log/npu/conf/slog` | Mounts the NPU log module into the container. |
| `-v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro` | Mounts the host `libyaml` `.so` file into the container. |
| `-v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools` | Mounts the driver tools directory `/usr/local/Ascend/driver/tools` into the container. |
| `-v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64` | Mounts the driver dynamic library directory `/usr/local/Ascend/driver/lib64` into the container. |
| `-v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels` | Mounts the aicpu library directory `/usr/lib64/aicpu_kernels` into the container. |
| `-v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro` | Mounts the memory usage directory `/sys/fs/cgroup/memory` from the host into the container in read-only mode. |
| `-v /etc/ascend_install.info:/etc/ascend_install.info` | Mounts the host installation information file `/etc/ascend_install.info` into the container. |
| `-v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info` | Mounts the host version information file `/usr/local/Ascend/driver/version.info` into the container. Please modify the path according to the actual situation. |
| `workload-image:v1.0` | The generated image file. |
| `/bin/bash` | Starts an interactive terminal Bash shell inside the container. |
