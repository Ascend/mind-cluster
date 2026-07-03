# Usage on the Docker Client<a name="ZH-CN_TOPIC_0000002479387248"></a>

## Usage Instructions<a name="section0966931165317"></a>

- Ascend Docker Runtime supports mounting physical and virtual chips. Before mounting virtual chips, refer to [Creating vNPUs](../virtual_instance/virtual_instance_with_hdk/static_vnpu_scheduling/01_creating_vnpu.md) to virtualize physical chips. Both static virtualization and dynamic virtualization of physical chips are supported.
- You can query the currently available physical chip IDs by running the `ls /dev/davinci*` command, and query the currently available virtual chip IDs by running the `ls /dev/vdavinci*` command.
- If you do not need to mount all the content in the default configuration file `/etc/ascend-docker-runtime.d/base.list` of Ascend Docker Runtime, create a custom configuration file (for example, `hostlog.list`) to reduce the mounted content. For details, see [(Optional) Configuring Custom Mounted Content](./01_configuring_custom_mounted_content.md).

## Mounting Chips Using Ascend Docker Runtime<a name="section11917171014591"></a>

In the examples, `image-name:tag` represents the image name and tag. For details about other parameters, see [Table 1](#table3488191614328).

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

After the container is started, run the following commands inside and outside the container to check whether the corresponding devices and drivers are mounted successfully. For the specific mount directory of each model, refer to [Content Mounted by Ascend Docker Runtime](../../references/appendix.md#content-mounted-by-ascend-docker-runtime). Example commands are as follows:

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

- Taking the Atlas 200I SoC A1 core board running an inference container as an example, modify the command according to the actual situation. The example is as follows, and the related parameters are shown in [Table 2](#table46513386334).

    ```shell
    docker run -it --device=/dev/davinci0:rwm --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/venc:rwm --device=/dev/vpc:rwm --device=/dev/davinci_manager:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /etc/sys_version.conf:/etc/sys_version.conf:ro -v /usr/local/bin/npu-smi:/usr/local/bin/npu-smi:ro -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /etc/hdcBasic.cfg:/etc/hdcBasic.cfg:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE]
    >- If the driver of the Atlas 200I SoC A1 core board is version 1.0.0 (Ascend HDK 22.0.0) or earlier, you need to mount `/dev/xsmem_dev` and `/dev/event_sched`.
    >- If the driver of the Atlas 200I SoC A1 core board is a version later than 1.0.0 (Ascend HDK 22.0.0), you do not need to mount `/dev/xsmem_dev` and `/dev/event_sched`.

- To run inference tasks on the Atlas 500 A2 intelligent station without using Ascend Docker Runtime, refer to the "Starting the Container" section in "Deploying Ascend Software (Customized System Scenario) > Container Deployment > [Creating a Container Image](https://support.huawei.com/enterprise/en/doc/EDOC1100438187/97777e51?idPath=23710424|251366513|254884019|261408772|258915651)" in the *Atlas 500 A2 Intelligent Station Ascend Software Installation Guide*.

## Parameter Description<a name="section131432039144912"></a>

**Table 1** Ascend Docker Runtime running parameters

<a name="table3488191614328"></a>

|Parameter|Description|Example|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>If the task does not require an NPU device, you can set the ASCEND_VISIBLE_DEVICES environment variable to void or leave it empty.</li><li>If the task requires an NPU device, you must use ASCEND_VISIBLE_DEVICES to specify the NPU device to be mounted to the container; otherwise, the NPU device mount will fail. When specifying devices by device ID, single IDs, ranges, and a mix of both are supported.</li><li>If the task requires an NPU device, you must use ASCEND_VISIBLE_DEVICES to specify the NPU device to be mounted to the container; otherwise, the NPU device mount will fail. When specifying devices by device ID, single IDs, ranges, and a mix of both are supported; when specifying devices by chip name, multiple chip names of the same type can be specified simultaneously.</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void indicates that the mount function of Ascend Docker Runtime is not used, and no NPU device, driver, or file directory is mounted. Related mount parameters will also become invalid.</li><li>Mount physical chip (NPU)</li><ul><li>ASCEND_VISIBLE_DEVICES=0 indicates that NPU device 0 (/dev/davinci0) is mounted to the container.</li><li>ASCEND_VISIBLE_DEVICES=1,3 indicates that NPU devices 1 and 3 are mounted to the container.</li><li>ASCEND_VISIBLE_DEVICES=0-2 indicates that NPU devices 0 through 2 (including 0 and 2) are mounted to the container, with the same effect as -e ASCEND_VISIBLE_DEVICES=0,1,2.</li><li>ASCEND_VISIBLE_DEVICES=0-2,4 indicates that NPU devices 0 through 2 and device 4 are mounted to the container, with the same effect as -e ASCEND_VISIBLE_DEVICES=0,1,2,4.</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y, where XXX represents the NPU device, with supported values being npu, Ascend910, Ascend310, Ascend310B, and Ascend310P; Y represents the physical NPU device ID.</li><ul><li>ASCEND_VISIBLE_DEVICES=npu-1 indicates that NPU device 1 is mounted to the container.</li><li>ASCEND_VISIBLE_DEVICES=npu-1,npu-3 indicates that NPU 1 and NPU 3 are mounted to the container.</li></ul></ul><div class="note"><span class="notetitle">**NOTE**</span><div class="notebody"><ul><li>When specifying devices by chip name, it is recommended to use the value npu uniformly.</li><li>Specifying both a device ID and an NPU name in a single parameter is not supported, meaning ASCEND_VISIBLE_DEVICES=0,npu-1 is not supported.</li></ul></div></div><li>Mount virtual chip (vNPU)<ul><li>**Static virtualization**: The usage is the same as for physical chips; simply replace the physical chip ID with the virtual chip ID (vNPU ID).</li><li>**Dynamic virtualization**: ASCEND_VISIBLE_DEVICES=0 indicates that a certain number of AICores are partitioned from NPU device 0.<div class="note"><span class="notetitle">**NOTE**</span><div class="notebody"><ul><li>A single dynamic virtualization command can only specify the ID of one physical NPU for dynamic virtualization.</li><li>Must be used together with ASCEND_VNPU_SPECS, which indicates the number of AICores partitioned on the specified NPU.</li><li>Can be used together with ASCEND_RUNTIME_OPTIONS, but only the value NODRV is allowed, indicating that driver-related directories are not mounted.</li></ul></div></div></li></ul></li></ul>|
|ASCEND_ALLOW_LINK|Specifies whether soft links are allowed in the mounted files or directories. This parameter must be specified in scenarios involving the Atlas 500 A2 Smart Station, Atlas 200I A2 acceleration module, and Atlas 200I DK A2 developer kit.<p>Other devices, such as Atlas training series products, <term>Atlas A2 training series products</term>, and the Atlas 200I SoC A1 core board, can use this parameter, but since soft links do not exist in their default mount content, specifying this parameter is unnecessary.</p>|<ul><li>ASCEND_ALLOW_LINK=True indicates that mounting driver files with soft links is allowed in scenarios involving the Atlas 500 A2 Smart Station, Atlas 200I A2 acceleration module, and Atlas 200I DK A2 developer kit.</li><li>If ASCEND_ALLOW_LINK=False or this parameter is not specified, Ascend Docker Runtime cannot be used on the Atlas 500 A2 Smart Station, Atlas 200I A2 acceleration module, and Atlas 200I DK A2 developer kit.</li></ul>|
|ASCEND_RUNTIME_OPTIONS|Restricts the chip ID specified in the ASCEND_VISIBLE_DEVICES parameter:<ul><li>NODRV: Indicates that driver-related directories are not mounted.</li><li>VIRTUAL: Indicates that a virtual chip is mounted.</li><li>NODRV,VIRTUAL: Indicates that a virtual chip is mounted and driver-related directories are not mounted.</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul><div class="note"><span class="notetitle">**NOTE**</span><div class="notebody"><ul><li>In static virtualization scenarios, ASCEND_RUNTIME_OPTIONS is a required parameter, and its value must include VIRTUAL.</li><li>In dynamic virtualization scenarios, if the ASCEND_RUNTIME_OPTIONS parameter is used, its value cannot include VIRTUAL.</li></ul></div></div>|
|ASCEND_RUNTIME_MOUNTS|Specifies the configuration file name for the content to be mounted. This file can configure the files and directories to be mounted to the container.|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2<div class="note"><span class="notetitle">**NOTE**</span><div class="notebody"><ul><li>By default, the /etc/ascend-docker-runtime.d/base.list configuration file is read.</li><li>For hostlog.list, modify it according to the actual custom configuration file name.</li><li>Reading multiple custom configuration files is supported.</li><li>File names must be lowercase and cannot contain uppercase letters. File names containing uppercase letters may cause the configuration file to fail to take effect.</li></ul></div></div></li></ul>|
|ASCEND_VNPU_SPECS|Partitions a certain number of AICores from a physical NPU device and specifies them as virtual devices. For supported values, see the "Virtualization Instance Template" column in Table 1 of [Virtualization Templates](../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md).<ul><li>This parameter can only be used for product forms that support dynamic virtualization.</li><li>Must be used together with the "ASCEND_VISIBLE_DEVICES" parameter, which specifies the physical NPU device used for virtualization.</li><li>When the value of the ASCEND_RUNTIME_OPTIONS parameter includes VIRTUAL, the ASCEND_VNPU_SPECS parameter will no longer take effect.</li></ul>|ASCEND_VNPU_SPECS=vir04 indicates that 4 AICores are partitioned as virtual devices and mounted to the container.|

**Table 2** Explanation of other parameters

<a name="table46513386334"></a>

|Parameter|Description|
|--|--|
|/dev/xsmem_dev|Mounts the memory device management to the container.|
|/dev/event_sched|Mounts the event scheduling device to the container.|
|/dev/ts_aisle|Mounts the device corresponding to the aicpudrv driver to the container.|
|/dev/svm0|Mounts the memory management device to the container.|
|/dev/sys|Mounts dvpp-related devices to the container.|
|/dev/vdec|Mounts dvpp-related devices to the container.|
|/dev/vpc|Mounts dvpp-related devices to the container.|
|/dev/log_drv|Mounts the logging-related device to the container.|
|/dev/upgrade|Mounts the device for obtaining Ascend system-related configurations and firmware to the container.|
|/dev/spi_smbus|Mounts the device related to out-of-band SPI communication to the container.|
|/dev/user_config|Mounts the device for managing user configurations to the container.|
|/dev/memory_bandwidth|Mounts the memory bandwidth-related device to the container.|
|-v /var/slogd:/var/slogd|Mounts the host machine log process file to the container in read-only mode.|
|-v /var/dmp_daemon:/var/dmp_daemon|Mounts the dmp daemon to the container.|
|-v /var/log/npu/conf/slog:/var/log/npu/conf/slog|Mounts the NPU log module to the container.|
|-v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro|Mounts the host machine libyaml .so file to the container.|
|-v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools|Mounts the driver-related tools directory "/usr/local/Ascend/driver/tools" to the container.|
|-v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64|Mounts the driver-dependent dynamic library directory "/usr/local/Ascend/driver/lib64" to the container.|
|-v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels|Mounts the aicpu lib library directory "/usr/lib64/aicpu_kernels" to the container.|
|-v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro|Mounts the dependency directory "/sys/fs/cgroup/memory" required for querying memory usage on the host machine to the container in read-only mode.|
|-v /etc/ascend_install.info:/etc/ascend_install.info|Mounts the host machine installation information file "/etc/ascend_install.info" to the container.|
|-v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info|Mounts the host machine version information file "/usr/local/Ascend/driver/version.info" to the container. Modify it based on the actual situation.|
|workload-image:v1.0|The generated image file.|
|/bin/bash|Starts an interactive terminal Bash Shell in the container.|
