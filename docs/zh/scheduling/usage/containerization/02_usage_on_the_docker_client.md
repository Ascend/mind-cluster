# 在Docker客户端使用<a name="ZH-CN_TOPIC_0000002479387248"></a>

**使用说明<a name="section0966931165317"></a>**

- Ascend Docker Runtime支持挂载物理芯片，同时支持挂载虚拟芯片。挂载虚拟芯片前需要参考[创建vNPU](../virtual_instance/virtual_instance_with_hdk/04_creating_vnpu.md)章节，对物理芯片进行虚拟化操作，支持对物理芯片进行静态虚拟化和动态虚拟化。
- 可通过<b>ls /dev/davinci\*</b>命令查询当前可用的物理芯片ID；通过<b>ls /dev/vdavinci\*</b>命令查询当前可用的虚拟芯片ID。
- 若用户不需要挂载Ascend Docker Runtime的默认配置文件“/etc/ascend-docker-runtime.d/base.list”中所有内容，可创建自定义配置文件（例如hostlog.list），减少挂载内容，具体操作请参考[（可选）配置自定义挂载内容](./01_configuring_custom_mounted_content.md)章节。

**使用Ascend Docker Runtime挂载芯片<a name="section11917171014591"></a>**

示例中的image-name:tag为镜像名称与标签，其他参数说明请参见[表1](#table3488191614328)。

- 示例1：启动容器时，挂载物理芯片ID为0的芯片。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 image-name:tag /bin/bash
    ```

- 示例2：启动容器时，仅挂载NPU设备和管理设备，不挂载驱动相关目录。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_OPTIONS=NODRV image-name:tag /bin/bash
    ```

- 示例3：启动容器时，挂载物理芯片ID为0的芯片，读取自定义配置文件hostlog.list中的挂载内容。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag /bin/bash
    ```

- 示例4：启动容器时，挂载虚拟芯片ID为100的芯片。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL image-name:tag /bin/bash
    ```

- 示例5：启动容器时，从物理芯片ID为0的芯片上，切分出4个AI Core作为虚拟设备并挂载至容器中。

    ```shell
    docker run -it --rm -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_VNPU_SPECS=vir04 image-name:tag /bin/bash
    ```

- 示例6：启动容器时，挂载物理芯片ID为0的芯片，并且允许挂载的驱动文件中存在软链接（仅适用于Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景）：

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True  image-name:tag /bin/bash
    ```

容器启动后，在容器内外执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录参考[Ascend Docker Runtime默认挂载内容](../../appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

**使用Ascend Docker Runtime挂载芯片和其他设备<a name="section111912299472"></a>**

使用Ascend Docker Runtime支持容器运行训练、推理或其他任务。

- 以Atlas 200I SoC A1 核心板运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表1](#table3488191614328)和[表2](#table46513386334)所示。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/vpc:rwm --device=/dev/log_drv:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /usr/lib64/libtensorflow.so:/usr/lib64/libtensorflow.so:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE]  
    >- 如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）及之前的版本，则需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。
    >- 如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）之后的版本，则不需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。

- 以Atlas 500 A2 智能小站运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表1](#table3488191614328)和[表2](#table46513386334)所示。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True  workload-image:v1.0 /bin/bash
    ```

**不使用Ascend Docker Runtime挂载芯片和其他设备<a name="section1212516610490"></a>**

- 以Atlas 200I SoC A1 核心板运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表2](#table46513386334)所示。

    ```shell
    docker run -it --device=/dev/davinci0:rwm --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/venc:rwm --device=/dev/vpc:rwm --device=/dev/davinci_manager:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /etc/sys_version.conf:/etc/sys_version.conf:ro -v /usr/local/bin/npu-smi:/usr/local/bin/npu-smi:ro -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /etc/hdcBasic.cfg:/etc/hdcBasic.cfg:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /usr/lib64/libtensorflow.so:/usr/lib64/libtensorflow.so:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE] 
    >- 如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）及之前的版本，则需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。
    >- 如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）之后的版本，则不需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。

- Atlas 500 A2 智能小站不使用Ascend Docker Runtime运行推理任务，可参考《Atlas 500 A2 智能小站 昇腾软件安装指南》的“部署昇腾软件（定制系统场景）\> 容器部署 \>  [制作容器镜像](https://support.huawei.com/enterprise/zh/doc/EDOC1100438187/97777e51?idPath=23710424|251366513|254884019|261408772|258915651)”章节中的“启动容器”，运行推理任务。

**参数说明<a name="section131432039144912"></a>**

**表 1** Ascend Docker Runtime运行参数解释

<a name="table3488191614328"></a>

|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>如果任务不需要使用NPU设备，可以设置ASCEND_VISIBLE_DEVICES环境变量的取值为void或为空。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void表示不使用Ascend Docker Runtime的挂载功能，不挂载NPU设备、驱动和文件目录。相关挂载参数也会失效。</li><li>挂载物理芯片（NPU）</li><ul><li>ASCEND_VISIBLE_DEVICES=0时，表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3时，表示将1、3号NPU设备挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=0-2时，表示将0号至2号NPU设备（包含0号和2号）挂载入容器中，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2。</li><li>ASCEND_VISIBLE_DEVICES=0-2,4时，表示将0号至2号以及4号NPU设备挂载入容器，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2,4。</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y，其中XXX表示NPU设备，支持的取值为npu、Ascend910、Ascend310、Ascend310B和Ascend310P；Y表示物理NPU设备ID。</li><ul><li>ASCEND_VISIBLE_DEVICES=npu-1，表示把1号NPU设备挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=npu-1,npu-3，表示把1号NPU和3号NPU挂载进容器。</li></ul></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用芯片名称指定设备时，建议统一取值npu。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=0，npu-1。</li></ul></div></div><li>挂载虚拟芯片（vNPU）<ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AI Core。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS使用，表示在指定的NPU上划分出的AI Core数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul></div></div></li></ul></li></ul>|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景下必须指定该参数。<p>其他设备如Atlas 训练系列产品、Atlas A2 训练系列产品和Atlas 200I SoC A1 核心板等产品可以使用该参数，但因其默认挂载内容中不存在软链接，所以无需额外指定该参数。</p>|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件将无法使用Ascend Docker Runtime。</li></ul>|
|ASCEND_RUNTIME_OPTIONS|对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：<ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul>|
|ASCEND_RUNTIME_MOUNTS|待挂载内容的配置文件名，该文件中可配置需要挂载到容器内的文件及目录。|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>默认读取/etc/ascend-docker-runtime.d/base.list配置文件。</li><li>hostlog.list请根据实际自定义配置文件名修改。</li><li>支持读取多个自定义配置文件。</li><li>文件名必须小写，不能包含大写字母，包含大写字母的文件名可能导致配置文件无法生效。</li></ul></div></div></li></ul>|
|ASCEND_VNPU_SPECS|从物理NPU设备中切分出一定数量的AI Core，指定为虚拟设备。支持的取值请参见[表1](../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md)。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li><li>参数ASCEND_RUNTIME_OPTIONS的取值包含VIRTUAL时，ASCEND_VNPU_SPECS参数将不再生效。</li></ul>|ASCEND_VNPU_SPECS=vir04表示切分4个AI Core作为虚拟设备，挂载至容器中。|

**表 2**  其他参数解释

<a name="table46513386334"></a>

|参数|参数说明|
|--|--|
|/dev/xsmem_dev|将内存设备管理挂载到容器。|
|/dev/event_sched|将事件调度的设备挂载到容器。|
|/dev/ts_aisle|将aicpudrv驱动对应的设备挂载到容器。|
|/dev/svm0|将内存管理的设备挂载到容器。|
|/dev/sys|将dvpp相关的设备挂载到容器。|
|/dev/vdec|将dvpp相关的设备挂载到容器。|
|/dev/vpc|将dvpp相关的设备挂载到容器。|
|/dev/log_drv|将日志记录相关的设备挂载到容器。|
|/dev/upgrade|将获取昇腾系统相关配置、固件设备挂载到容器。|
|/dev/spi_smbus|将设备带外spi通信相关的设备挂载到容器。|
|/dev/user_config|将管理用户配置相关的设备挂载到容器。|
|/dev/memory_bandwidth|将内存带宽相关的设备挂载到容器。|
|-v /var/slogd:/var/slogd|将宿主机日志进程文件以只读方式挂载到容器。|
|-v /var/dmp_daemon:/var/dmp_daemon|将dmp守护进程挂载到容器。|
|-v /var/log/npu/conf/slog:/var/log/npu/conf/slog|将npu日志模块挂载到容器。|
|-v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro|将宿主机libyaml的.so文件挂载到容器中。|
|-v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools|将驱动相关工具目录“/usr/local/Ascend/driver/tools”挂载到容器中。|
|-v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64|将驱动依赖动态库目录“/usr/local/Ascend/driver/lib64”挂载到容器中。|
|-v /usr/lib64/libtensorflow.so:/usr/lib64/libtensorflow.so|将TensorFlow的aicpu算子库文件“/usr/lib64/libtensorflow.so”挂载到容器中。|
|-v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels|将aicpu lib库目录“/usr/lib64/aicpu_kernels”挂载到容器。|
|-v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro|将宿主机查询内存占用率所需依赖目录“/sys/fs/cgroup/memory”以只读方式挂载到容器中。|
|-v /etc/ascend_install.info:/etc/ascend_install.info|将宿主机安装信息文件“/etc/ascend_install.info”挂载到容器中。|
|-v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info|将宿主机版本信息文件“/usr/local/Ascend/driver/version.info”挂载到容器中，请根据实际情况修改。|
|workload-image:v1.0|生成的镜像文件。|
|/bin/bash|在容器内启动交互式的终端Bash Shell。|
