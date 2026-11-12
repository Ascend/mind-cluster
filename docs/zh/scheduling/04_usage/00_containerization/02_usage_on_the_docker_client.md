# 在Docker客户端使用<a name="ZH-CN_TOPIC_0000002479387248"></a>

## 使用说明<a name="section0966931165317"></a>

- Ascend Docker Runtime支持挂载物理芯片，同时支持挂载虚拟芯片。挂载虚拟芯片前需要参考[创建vNPU](../02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md)章节，对物理芯片进行虚拟化操作，支持对物理芯片进行静态虚拟化和动态虚拟化。
- 可通过<b>ls /dev/davinci\*</b>命令查询当前可用的物理芯片ID；通过<b>ls /dev/vdavinci\*</b>命令查询当前可用的虚拟芯片ID。
- 若用户不需要挂载Ascend Docker Runtime的默认配置文件“/etc/ascend-docker-runtime.d/base.list”中所有内容，可创建自定义配置文件（例如hostlog.list），减少挂载内容，具体操作请参考[（可选）配置自定义挂载内容](./01_configuring_custom_mounted_content.md)章节。
<!-- npu="950" id2 -->
- 在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下，支持在Legacy模式下通过`ASCEND_RDMA_DEVICES`指定UB网卡在“/sys/class/infiniband/”下的完整HCA名称（可通过<b>ls /sys/class/infiniband</b>命令查询），将对应的RDMA字符设备与驱动文件挂载至容器；前提条件和默认驱动挂载说明请参见[使用前必读](./00_before_you_start.md)。
<!-- end id2 -->

## 使用Ascend Docker Runtime挂载芯片<a name="section11917171014591"></a>

示例中的\{image-name:tag\}为镜像名称与标签，其他参数说明请参见[表1](#table3488191614328)。

- 示例1：启动容器时，挂载物理芯片ID为0的芯片。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 {image-name:tag} /bin/bash
    ```

- 示例2：启动容器时，仅挂载NPU设备和管理设备，不挂载驱动相关目录。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_OPTIONS=NODRV {image-name:tag} /bin/bash
    ```

- 示例3：启动容器时，挂载物理芯片ID为0的芯片，读取自定义配置文件hostlog.list中的挂载内容。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog {image-name:tag} /bin/bash
    ```

- 示例4：启动容器时，挂载虚拟芯片ID为100的芯片。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL {image-name:tag} /bin/bash
    ```

- 示例5：启动容器时，从物理芯片ID为0的芯片上，切分出4个AICore作为虚拟设备并挂载至容器中。

    ```shell
    docker run -it --rm -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_VNPU_SPECS=vir04 {image-name:tag} /bin/bash
    ```

<!-- npu="310b" id1 -->
- 示例6：启动容器时，挂载物理芯片ID为0的芯片，并且允许挂载的驱动文件中存在软链接（仅适用于Atlas 500 A2智能小站、Atlas 200I A2加速模块和Atlas 200I DK A2开发者套件场景）：

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True {image-name:tag} /bin/bash
    ```

<!-- end id1 -->
容器启动后，在容器内外执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录参考[Ascend Docker Runtime默认挂载内容](../../07_references/05_appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

可能的输出结果如下：

```ColdFusion
davinci0
davinci_manager
devmm_svm
hisi_hdc
include lib64
dcmi
npu-smi
```

<!-- npu="950" id3 -->

## 使用Ascend Docker Runtime挂载UB网卡设备及驱动<a name="section_rdma_docker_usage"></a>

在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下，可通过`ASCEND_RDMA_DEVICES`挂载UB网卡的RDMA字符设备与驱动文件。示例中的\{image-name:tag\}为镜像名称与标签，HCA名称与NPU编号请根据实际环境修改，相关参数说明请参见[表1](#table3488191614328)。

- 示例1：启动容器时，同时挂载0号NPU设备与`hrn5_0`、`hrn5_1`对应的RDMA设备，并自动挂载NPU与RDMA驱动文件。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RDMA_DEVICES=hrn5_0,hrn5_1 {image-name:tag} /bin/bash
    ```

- 示例2：启动容器时，仅挂载`hrn5_0`对应的RDMA设备及驱动文件，不挂载NPU设备及NPU驱动。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=void -e ASCEND_RDMA_DEVICES=hrn5_0 {image-name:tag} /bin/bash
    ```

- 示例3：启动容器时，仅挂载NPU和RDMA字符设备，不自动挂载驱动文件（由容器镜像自带驱动，或通过容器卷按需挂载所需驱动文件）。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RDMA_DEVICES=hrn5_0 -e ASCEND_RUNTIME_OPTIONS=NODRV {image-name:tag} /bin/bash
    ```

    >[!NOTE]
    >
    >- 若镜像内已包含RDMA驱动、仅需从宿主机挂载NPU驱动目录，可在上述命令中增加`-v /usr/local/Ascend/driver:/usr/local/Ascend/driver:ro`（路径按实际环境调整）。
    >- 若镜像内已包含NPU驱动、仅需从宿主机挂载RDMA驱动文件，可在上述命令中按`rdma_driver.list`中的路径逐一增加只读挂载（不挂载整个“/usr/lib64”目录）。以下示例假设宿主机不存在“/usr/lib64/libibverbs/libhrn5-rdmav34.so”；若宿主机存在该文件，需一并增加该路径的只读挂载。
    >
    >    ```shell
    >    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RDMA_DEVICES=hrn5_0 -e ASCEND_RUNTIME_OPTIONS=NODRV \
    >      --mount type=bind,src=/etc/libibverbs.d/hrn5.driver,dst=/etc/libibverbs.d/hrn5.driver,readonly \
    >      --mount type=bind,src=/usr/lib64/libhrn5-rdmav34.so,dst=/usr/lib64/libhrn5-rdmav34.so,readonly \
    >      --mount type=bind,src=/usr/lib64/libhrn5.so,dst=/usr/lib64/libhrn5.so,readonly \
    >      --mount type=bind,src=/usr/lib64/libhrn5.so.1,dst=/usr/lib64/libhrn5.so.1,readonly \
    >      --mount type=bind,src=/usr/lib64/libibv_extend.so,dst=/usr/lib64/libibv_extend.so,readonly \
    >      --mount type=bind,src=/usr/lib64/libibv_extend.so.4,dst=/usr/lib64/libibv_extend.so.4,readonly \
    >      --mount type=bind,src=/usr/lib64/libibv_extend.so.4.2.0,dst=/usr/lib64/libibv_extend.so.4.2.0,readonly \
    >      --mount type=bind,src=/usr/lib64/libc_sec.so,dst=/usr/lib64/libc_sec.so,readonly \
    >      {image-name:tag} /bin/bash
    >    ```

容器启动后，可在容器内执行以下命令检查RDMA字符设备与UB网卡设备是否识别成功：

```shell
ls -l /dev/infiniband && ibv_devices
```

<!-- end id3 -->

## 使用Ascend Docker Runtime挂载芯片和其他设备<a name="section111912299472"></a>

使用Ascend Docker Runtime支持容器运行训练、推理或其他任务。

- 以Atlas 200I SoC A1核心板运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表1](#table3488191614328)和[表2](#table46513386334)所示。

    ```shell
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/vpc:rwm --device=/dev/log_drv:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE]
    >- 如果Atlas 200I SoC A1核心板的驱动是1.0.0（Ascend HDK 22.0.0）及之前的版本，则需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。
    >- 如果Atlas 200I SoC A1核心板的驱动是1.0.0（Ascend HDK 22.0.0）之后的版本，则不需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。

- 以Atlas 500 A2智能小站运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表1](#table3488191614328)和[表2](#table46513386334)所示。

    ```shell
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True workload-image:v1.0 /bin/bash
    ```

## 不使用Ascend Docker Runtime挂载芯片和其他设备<a name="section1212516610490"></a>

- 以Atlas 200I SoC A1核心板运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表2](#table46513386334)所示。

    ```shell
    docker run -it --device=/dev/davinci0:rwm --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/venc:rwm --device=/dev/vpc:rwm --device=/dev/davinci_manager:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /etc/sys_version.conf:/etc/sys_version.conf:ro -v /usr/local/bin/npu-smi:/usr/local/bin/npu-smi:ro -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /etc/hdcBasic.cfg:/etc/hdcBasic.cfg:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE]
    >- 如果Atlas 200I SoC A1核心板的驱动是1.0.0（Ascend HDK 22.0.0）及之前的版本，则需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。
    >- 如果Atlas 200I SoC A1核心板的驱动是1.0.0（Ascend HDK 22.0.0）之后的版本，则不需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。

- Atlas 500 A2智能小站不使用Ascend Docker Runtime运行推理任务，可参考《Atlas 500 A2智能小站 昇腾软件安装指南》的“部署昇腾软件（定制系统场景）\> 容器部署 \>  [制作容器镜像](https://support.huawei.com/enterprise/zh/doc/EDOC1100438187/97777e51?idPath=23710424|251366513|254884019|261408772|258915651)”章节中的“启动容器”，运行推理任务。

## 参数说明<a name="section131432039144912"></a>

**表 1** Ascend Docker Runtime运行参数解释

<a name="table3488191614328"></a>

|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>如果任务不需要使用NPU设备，可以设置ASCEND_VISIBLE_DEVICES环境变量的取值为void或为空。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void表示不挂载NPU设备及NPU驱动。Legacy模式下，若未指定ASCEND_RDMA_DEVICES，不执行Hook挂载流程；若指定了ASCEND_RDMA_DEVICES，仅挂载所选RDMA设备及对应的RDMA驱动文件。若同时指定ASCEND_RUNTIME_OPTIONS=NODRV，跳过驱动文件的自动挂载，仅注入所选RDMA字符设备并校验宿主机RDMA依赖。</li><li>挂载物理芯片（NPU）</li><ul><li>ASCEND_VISIBLE_DEVICES=0时，表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3时，表示将1、3号NPU设备挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=0-2时，表示将0号至2号NPU设备（包含0号和2号）挂载入容器中，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2。</li><li>ASCEND_VISIBLE_DEVICES=0-2,4时，表示将0号至2号以及4号NPU设备挂载入容器，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2,4。</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y，其中XXX表示NPU设备，支持的取值为npu、Ascend910、Ascend310、Ascend310B和Ascend310P；Y表示物理NPU设备ID。</li><ul><li>ASCEND_VISIBLE_DEVICES=npu-1，表示把1号NPU设备挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=npu-1,npu-3，表示把1号NPU和3号NPU挂载进容器。</li></ul></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用芯片名称指定设备时，建议统一取值npu。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=0,npu-1。</li></ul></div></div><li>挂载虚拟芯片（vNPU）<ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AICore。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS使用，表示在指定的NPU上划分出的AICore数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul></div></div></li></ul></li></ul>|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2智能小站、Atlas 200I A2加速模块和Atlas 200I DK A2开发者套件场景下必须指定该参数。<p>其他设备如<term>Atlas训练系列产品</term>、<term>Atlas A2训练系列产品</term>和Atlas 200I SoC A1核心板等产品可以使用该参数，但因其默认挂载内容中不存在软链接，所以无需额外指定该参数。</p>|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2智能小站、Atlas 200I A2加速模块、Atlas 200I DK A2开发者套件、Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2智能小站、Atlas 200I A2加速模块和Atlas 200I DK A2开发者套件将无法使用Ascend Docker Runtime。</li><li>挂载所选RDMA设备的驱动文件时，自动允许通过属主、权限和路径校验的软链接，无需指定ASCEND_ALLOW_LINK=True；设置ASCEND_ALLOW_LINK=False时，该流程仍使用True。</li></ul>|
|ASCEND_UB_DRV_MOUNT|是否挂载UnifiedBus驱动相关文件，在Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点等需要支持UnifiedBus通信协议的场景下，通过指定该参数决定是否挂载宿主机的UnifiedBus驱动相关文件到容器中。<p>其他设备如<term>Atlas训练系列产品</term>、<term>Atlas A2训练系列产品</term>和<term>Atlas A3训练系列产品</term>等可以使用该参数，但因其不依赖UnifiedBus驱动相关文件，所以无需额外指定该参数。</p>|<ul><li>ASCEND_UB_DRV_MOUNT=False，表示在Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点场景下不挂载UnifiedBus驱动相关文件，此时可通过在容器内安装UnifiedBus驱动用户态软件包的方式支持相关功能。</li><li>ASCEND_UB_DRV_MOUNT=True或者不指定该参数，Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点将挂载UnifiedBus驱动相关文件。由于UnifiedBus驱动相关文件存在软链接，为了成功挂载，会自动跳过软链接校验，无需指定ASCEND_ALLOW_LINK=True环境变量。需要挂载ub_driver.list的场景，指定ASCEND_ALLOW_LINK=False会无效。</li></ul>|
|ASCEND_RUNTIME_OPTIONS|配置容器运行时的挂载与虚拟化选项：<ul><li>NODRV：不自动挂载NPU与RDMA驱动相关文件及目录（若指定了ASCEND_RDMA_DEVICES，仍注入RDMA字符设备并校验宿主机RDMA设备、设备表、清单及必需文件）。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不自动挂载NPU与RDMA驱动相关文件及目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>静态虚拟化场景下，ASCEND_RUNTIME_OPTIONS为必选参数，且取值必须包含VIRTUAL。</li><li>动态虚拟化场景下，若使用ASCEND_RUNTIME_OPTIONS参数，则其取值不能包含VIRTUAL。</li></ul></div></div>|
|ASCEND_RUNTIME_MOUNTS|待挂载内容的配置文件名，该文件中可配置需要挂载到容器内的文件及目录。|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>默认读取/etc/ascend-docker-runtime.d/base.list配置文件。</li><li>hostlog.list请根据实际自定义配置文件名修改。</li><li>支持读取多个自定义配置文件。</li><li>文件名必须小写，不能包含大写字母，包含大写字母的文件名可能导致配置文件无法生效。</li></ul></div></div>|
|ASCEND_VNPU_SPECS|从物理NPU设备中切分出一定数量的AICore，指定为虚拟设备。支持的取值请参见[虚拟化模板](../02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md)中表1的“虚拟化实例模板”列。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li><li>参数ASCEND_RUNTIME_OPTIONS的取值包含VIRTUAL时，ASCEND_VNPU_SPECS参数将不再生效。</li></ul>|ASCEND_VNPU_SPECS=vir04表示切分4个AICore作为虚拟设备，挂载至容器中。|
|ASCEND_RDMA_DEVICES|指定待挂载的RDMA设备，参数取值为UB网卡在“/sys/class/infiniband/”下的完整HCA名称，在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下使用，仅在Legacy注入模式下支持；CDI模式会忽略ASCEND_RDMA_DEVICES，不会根据该参数自动注入RDMA设备及挂载对应驱动文件。<ul><li>ASCEND_RDMA_DEVICES未设置、取值为空或void时，不读取RDMA设备表，不自动发现和注入RDMA设备及驱动。</li><li>仅支持使用英文逗号分隔的HCA设备全称，去重后最多支持512个。名称仅支持英文字母、数字和下划线，不支持纯数字或“0-7”等范围格式，且必须以英文字母开头。</li></ul><p>该参数与ASCEND_VISIBLE_DEVICES相互独立，前提条件和默认驱动挂载说明请参见[使用前必读](./00_before_you_start.md)。</p>|<ul><li>ASCEND_RDMA_DEVICES=void，表示不自动发现和注入RDMA设备及驱动；若仍申请NPU，ASCEND_RUNTIME_MOUNTS手动指定的集合仍按普通规则挂载。NPU和RDMA都未申请时，不执行Hook挂载流程。</li><li>ASCEND_RDMA_DEVICES=hrn5_0，表示挂载hrn5_0对应的“/dev/infiniband/uverbsX”字符设备（宿主机存在“/dev/infiniband/rdma_cm”时一并挂载）及匹配的RDMA驱动文件。</li><li>ASCEND_RDMA_DEVICES=hrn5_0,hrn5_1，表示同时挂载hrn5_0和hrn5_1对应的RDMA字符设备及驱动文件。</li></ul>|

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
|-v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels|将aicpu lib库目录“/usr/lib64/aicpu_kernels”挂载到容器。|
|-v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro|将宿主机查询内存占用率所需依赖目录“/sys/fs/cgroup/memory”以只读方式挂载到容器中。|
|-v /etc/ascend_install.info:/etc/ascend_install.info|将宿主机安装信息文件“/etc/ascend_install.info”挂载到容器中。|
|-v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info|将宿主机版本信息文件“/usr/local/Ascend/driver/version.info”挂载到容器中，请根据实际情况修改。|
|workload-image:v1.0|生成的镜像文件。|
|/bin/bash|在容器内启动交互式的终端Bash Shell。|
