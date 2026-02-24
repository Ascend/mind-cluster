# 容器化支持特性指南<a name="ZH-CN_TOPIC_0000002479227062"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002511427169"></a>

容器化支持是一种将应用程序及其依赖项打包到一个独立、可移植的环境（容器）中的技术支持。了解容器化支持所依赖组件、使用说明等详细介绍，请参见[容器化支持](../introduction.md#容器化支持)章节。

**前提条件<a name="section1632062465010"></a>**

在使用容器化支持特性前，需要确保Ascend Docker Runtime组件已经安装，若没有安装，可以参考[安装部署](../installation_guide.md#安装部署)章节进行操作。

**使用说明<a name="section44381612353"></a>**

-   容器化支持可以和训练场景下的所有特性一起使用，也可以和推理场景的所有特性一起使用。
-   若使用Volcano进行任务调度，则不建议通过Docker或Containerd指令创建/挂载NPU卡的容器，并在容器内跑任务。否则可能会触发Volcano调度问题。

**支持的产品形态<a name="section169961844182917"></a>**

支持以下产品使用容器化支持。

-   Atlas 训练系列产品
-   Atlas A2 训练系列产品
-   Atlas A3 训练系列产品
-   推理服务器（插Atlas 300I 推理卡）
-   Atlas 200/300/500 推理产品
-   Atlas 200I/500 A2 推理产品
-   Atlas 推理系列产品
-   Atlas 800I A2 推理服务器
-   A200I A2 Box 异构组件
-   Atlas 800I A3 超节点服务器
-   推理服务器（插Atlas 350 标卡）

**使用场景<a name="section124697813416"></a>**

Ascend Docker Runtime组件支持在以下4种场景下使用容器化支持功能。

-   [在Docker客户端使用](#在Docker客户端使用)
-   [K8s集成Docker使用](#K8s集成Docker使用)
-   [在Containerd客户端使用](#在Containerd客户端使用)
-   [在K8s集成Containerd使用](#在K8s集成Containerd使用)


## （可选）配置自定义挂载内容<a name="ZH-CN_TOPIC_0000002511427171"></a>

Ascend Docker Runtime会为用户默认挂载驱动以及基础配置文件“/etc/ascend-docker-runtime.d/base.list”中的全部内容。若用户需要挂载文件里的全部路径，则跳过本小节；若用户不需要挂载基础配置文件base.list中的全部内容时，可新增自定义配置文件，减少挂载的内容。自定义配置文件挂载内容须基于base.list文件，操作如下：

1.  进入配置文件目录。

    ```
    cd /etc/ascend-docker-runtime.d/
    ```

    该目录下已存在基础配置文件base.list，内容即Ascend Docker Runtime默认挂载内容，具体可参见[Ascend Docker Runtime默认挂载内容](../appendix.md#ascend-docker-runtime默认挂载内容)，原则上不允许用户修改base.list文件。

2.  创建新的配置文件，文件名可自定义（如hostlog.list）。

    ```
    vi hostlog.list
    ```

3.  将需要挂载的文件或目录写入hostlog.list，保存并退出。
4.  执行命令，使自定义配置文件hostlog.list生效。示例如下：

    ```
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag /bin/bash
    ```

    >[!NOTE] 说明 
    >ASCEND\_VISIBLE\_DEVICES和ASCEND\_RUNTIME\_MOUNTS参数说明，请参见[表1](#table3488191614328)。


## 在Docker客户端使用<a name="ZH-CN_TOPIC_0000002479387248"></a>

**使用说明<a name="section0966931165317"></a>**

-   Ascend Docker Runtime支持挂载物理芯片，同时支持挂载虚拟芯片。挂载虚拟芯片前需要参考[创建vNPU](./virtual_instance.md#创建vnpu)章节，对物理芯片进行虚拟化操作，支持对物理芯片进行静态虚拟化和动态虚拟化。
-   可通过<b>ls /dev/davinci\*</b>命令查询当前可用的物理芯片ID；通过<b>ls /dev/vdavinci\*</b>命令查询当前可用的虚拟芯片ID。
-   若用户不需要挂载Ascend Docker Runtime的默认配置文件“/etc/ascend-docker-runtime.d/base.list”中所有内容，可创建自定义配置文件（例如hostlog.list），减少挂载内容，具体操作请参考[（可选）配置自定义挂载内容](#可选配置自定义挂载内容)章节。

**使用Ascend Docker Runtime挂载芯片<a name="section11917171014591"></a>**

示例中的image-name:tag为镜像名称与标签，其他参数说明请参见[表1](#table3488191614328)。

-   示例1：启动容器时，挂载物理芯片ID为0的芯片。

    ```
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 image-name:tag /bin/bash
    ```

-   示例2：启动容器时，仅挂载NPU设备和管理设备，不挂载驱动相关目录。

    ```
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_OPTIONS=NODRV image-name:tag /bin/bash
    ```

-   示例3：启动容器时，挂载物理芯片ID为0的芯片，读取自定义配置文件hostlog.list中的挂载内容。

    ```
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag /bin/bash
    ```

-   示例4：启动容器时，挂载虚拟芯片ID为100的芯片。

    ```
    docker run -it -e ASCEND_VISIBLE_DEVICES=100 -e ASCEND_RUNTIME_OPTIONS=VIRTUAL image-name:tag /bin/bash
    ```

-   示例5：启动容器时，从物理芯片ID为0的芯片上，切分出4个AI Core作为虚拟设备并挂载至容器中。

    ```
    docker run -it --rm -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_VNPU_SPECS=vir04 image-name:tag /bin/bash
    ```

-   示例6：启动容器时，挂载物理芯片ID为0的芯片，并且允许挂载的驱动文件中存在软链接（仅适用于Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景）：

    ```
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True  image-name:tag /bin/bash
    ```

容器启动后，在容器内外执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录参考[Ascend Docker Runtime默认挂载内容](../appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

**使用Ascend Docker Runtime挂载芯片和其他设备<a name="section111912299472"></a>**

使用Ascend Docker Runtime支持容器运行训练、推理或其他任务。

-   以Atlas 200I SoC A1 核心板运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表1](#table3488191614328)和[表2](#table46513386334)所示。

    ```
    docker run -it -e ASCEND_VISIBLE_DEVICES=0 --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/vpc:rwm --device=/dev/log_drv:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /usr/lib64/libtensorflow.so:/usr/lib64/libtensorflow.so:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE] 说明 
    >-   如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）及之前的版本，则需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。
    >-   如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）之后的版本，则不需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。

-   以Atlas 500 A2 智能小站运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表1](#table3488191614328)和[表2](#table46513386334)所示。

    ```
    docker run --rm -it -e ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True  workload-image:v1.0 /bin/bash
    ```

**不使用Ascend Docker Runtime挂载芯片和其他设备<a name="section1212516610490"></a>**

-   以Atlas 200I SoC A1 核心板运行推理容器为例，用户请根据实际情况修改。示例如下，相关参数如[表2](#table46513386334)所示。

    ```
    docker run -it --device=/dev/davinci0:rwm --device=/dev/xsmem_dev:rwm --device=/dev/event_sched:rwm --device=/dev/svm0:rwm --device=/dev/sys:rwm --device=/dev/vdec:rwm --device=/dev/venc:rwm --device=/dev/vpc:rwm --device=/dev/davinci_manager:rwm --device=/dev/spi_smbus:rwm --device=/dev/upgrade:rwm --device=/dev/user_config:rwm --device=/dev/ts_aisle:rwm --device=/dev/memory_bandwidth:rwm -v /etc/sys_version.conf:/etc/sys_version.conf:ro -v /usr/local/bin/npu-smi:/usr/local/bin/npu-smi:ro -v /var/dmp_daemon:/var/dmp_daemon:ro -v /var/slogd:/var/slogd:ro -v /var/log/npu/conf/slog/slog.conf:/var/log/npu/conf/slog/slog.conf:ro -v /etc/hdcBasic.cfg:/etc/hdcBasic.cfg:ro -v /usr/local/Ascend/driver/tools:/usr/local/Ascend/driver/tools -v /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64 -v /usr/lib64/aicpu_kernels:/usr/lib64/aicpu_kernels:ro -v /usr/lib64/libtensorflow.so:/usr/lib64/libtensorflow.so:ro -v /sys/fs/cgroup/memory:/sys/fs/cgroup/memory:ro -v /usr/lib64/libyaml-0.so.2:/usr/lib64/libyaml-0.so.2:ro -v /etc/ascend_install.info:/etc/ascend_install.info -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info workload-image:v1.0 /bin/bash
    ```

    >[!NOTE] 说明 
    >-   如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）及之前的版本，则需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。
    >-   如果Atlas 200I SoC A1 核心板的驱动是1.0.0（Ascend HDK 22.0.0）之后的版本，则不需要挂载/dev/xsmem\_dev和/dev/event\_sched这两个设备。

-   Atlas 500 A2 智能小站不使用Ascend Docker Runtime运行推理任务，可参考《Atlas 500 A2 智能小站 昇腾软件安装指南》的“部署昇腾软件（定制系统场景）\> 容器部署 \>  [制作容器镜像](https://support.huawei.com/enterprise/zh/doc/EDOC1100438187/97777e51?idPath=23710424|251366513|254884019|261408772|258915651)”章节中的“启动容器”，运行推理任务。

**参数说明<a name="section131432039144912"></a>**

**表 1** Ascend Docker Runtime运行参数解释

<a name="table3488191614328"></a>
|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>如果任务不需要使用NPU设备，可以设置ASCEND_VISIBLE_DEVICES环境变量的取值为void或为空。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void表示不使用Ascend Docker Runtime的挂载功能，不挂载NPU设备、驱动和文件目录。相关挂载参数也会失效。</li><li>挂载物理芯片（NPU）</li><ul><li>ASCEND_VISIBLE_DEVICES=0时，表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3时，表示将1、3号NPU设备挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=0-2时，表示将0号至2号NPU设备（包含0号和2号）挂载入容器中，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2。</li><li>ASCEND_VISIBLE_DEVICES=0-2,4时，表示将0号至2号以及4号NPU设备挂载入容器，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2,4。</li></ul><li>挂载虚拟芯片（vNPU）<ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AI Core。<div class="note"><span>[!NOTE] 说明</span><div class="notebody"><ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS使用，表示在指定的NPU上划分出的AI Core数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul>|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景下必须指定该参数。<p>其他设备如Atlas 训练系列产品、Atlas A2 训练系列产品和Atlas 200I SoC A1 核心板等产品可以使用该参数，但因其默认挂载内容中不存在软链接，所以无需额外指定该参数。</p>|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件将无法使用Ascend Docker Runtime。</li></ul>|
|ASCEND_RUNTIME_OPTIONS|对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：<ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul>|
|ASCEND_RUNTIME_MOUNTS|待挂载内容的配置文件名，该文件中可配置需要挂载到容器内的文件及目录。|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2<div class="note"><span>[!NOTE] 说明</span><div class="notebody"><ul><li>默认读取/etc/ascend-docker-runtime.d/base.list配置文件。</li><li>hostlog.list请根据实际自定义配置文件名修改。</li><li>支持读取多个自定义配置文件。</li><li>文件名必须小写，不能包含大写字母，包含大写字母的文件名可能导致配置文件无法生效。</li></ul></li></ul>|
|ASCEND_VNPU_SPECS|从物理NPU设备中切分出一定数量的AI Core，指定为虚拟设备。支持的取值请参见<a href="./virtual_instance.md#虚拟化规则">表1</a>。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li><li>参数ASCEND_RUNTIME_OPTIONS的取值包含VIRTUAL时，ASCEND_VNPU_SPECS参数将不再生效。</li></ul>|ASCEND_VNPU_SPECS=vir04表示切分4个AI Core作为虚拟设备，挂载至容器中。|


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
|-v** **/var/slogd:/var/slogd|将宿主机日志进程文件以只读方式挂载到容器。|
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



## K8s集成Docker使用<a name="ZH-CN_TOPIC_0000002511347209"></a>

K8s集成Docker场景下，用户需要安装Ascend Docker Runtime。

-   **（二选一）申请NPU资源情况下**。使用任务YAML下发训练或推理任务时，NPU芯片的分配由Volcano和Ascend Device Plugin组件自动完成；NPU芯片及相关文件目录的挂载由Ascend Docker Runtime组件自动完成。示例如下。

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # 镜像名称根据实际情况修改
    ...
                resources:
                  limits:
                    huawei.com/Ascend910: 1     # 资源名称和数量根据实际情况修改
                  requests:
                    huawei.com/Ascend910: 1     #  资源名称和数量根据实际情况修改
    ...
    ```

-   **（二选一）如果未申请NPU资源**，由集成平台写入ASCEND\_VISIBLE\_DEVICES=void环境变量。示例如下。

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # 镜像名称根据实际情况修改
    ...
                env:
                - name: ASCEND_VISIBLE_DEVICES     # 未使用resources申请NPU资源时需增加此配置
                   value: "void"
    ...
    ```


## 在Containerd客户端使用<a name="ZH-CN_TOPIC_0000002511347203"></a>

**使用说明<a name="section0966931165317"></a>**

-   在Containerd客户端使用Ascend Docker Runtime挂载之前，需要确认当前cgroup的版本。执行<b>stat -fc %T /sys/fs/cgroup/</b>命令，若显示为tmpfs，表示当前为cgroup v1版本；若显示为cgroup2fs，表示当前为cgroup v2版本。
-   Ascend Docker Runtime支持挂载物理芯片，同时支持挂载虚拟芯片。挂载虚拟芯片前需要参考[创建vNPU](./virtual_instance.md#创建vnpu)章节，对物理芯片进行虚拟化操作，支持对物理芯片进行静态虚拟化和动态虚拟化。
-   可通过<b>ls /dev/davinci\*</b>命令查询当前可用的物理芯片ID；通过<b>ls /dev/vdavinci\*</b>命令查询当前可用的虚拟芯片ID。
-   若用户不需要挂载Ascend Docker Runtime的默认配置文件“/etc/ascend-docker-runtime.d/base.list”中所有内容，可创建自定义配置文件（例如hostlog.list），减少挂载内容，具体操作请参考[（可选）配置自定义挂载内容](#可选配置自定义挂载内容)章节。

**使用示例<a name="section148905517122"></a>**

示例中的image-name:tag为镜像名称与标签，如“ascend-tensorflow:tensorflow\_TAG”。containerID为容器ID，使用ctr启动容器需要指定容器ID，如“c1”。

-   示例1：启动容器时，挂载物理芯片ID为0的芯片。
    -   cgroup v1

        ```
        ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=0 image-name:tag containerID bash
        ```

    -   cgroup v2

        ```
        ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 image-name:tag containerID bash
        ```

-   示例2：启动容器时，仅挂载NPU设备和管理设备，不挂载驱动相关目录。
    -   cgroup v1

        ```
        ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_OPTIONS=NODRV image-name:tag containerID bash
        ```

    -   cgroup v2

        ```
        ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_OPTIONS=NODRV image-name:tag containerID bash
        ```

-   示例3：启动容器时，挂载物理芯片ID为0的芯片，读取自定义配置文件hostlog中的挂载内容。
    -   cgroup v1

        ```
        ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag containerID bash
        ```

    -   cgroup v2

        ```
        ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag containerID bash
        ```

-   示例4：启动容器时，挂载虚拟芯片ID为100的芯片。
    -   cgroup v1

        ```
        ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=100 --env ASCEND_RUNTIME_OPTIONS=VIRTUAL image-name:tag containerID bash
        ```

    -   cgroup v2

        ```
        ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=100 --env ASCEND_RUNTIME_OPTIONS=VIRTUAL image-name:tag containerID bash
        ```

-   示例5：启动容器时，从物理芯片ID为0的芯片上，切分出4个AI Core作为虚拟设备并挂载至容器中。
    -   cgroup v1

        ```
        ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_VNPU_SPECS=vir04 image-name:tag containerID bash
        ```

    -   cgroup v2

        ```
        ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_VNPU_SPECS=vir04 image-name:tag containerID bash
        ```

-   示例6：启动容器时，挂载物理芯片ID为0的芯片，并且允许挂载的驱动文件中存在软链接（仅适用于Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景）。
    -   cgroup v1

        ```
        ctr run --runtime io.containerd.runtime.v1.linux -t --env ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True image-name:tag containerID bash
        ```

    -   cgroup v2

        ```
        ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime -t --env ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True image-name:tag containerID bash
        ```

启动命令相关参数如[表1](#table5134121862415)所示。

容器启动后，可执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录参考[Ascend Docker Runtime默认挂载内容](../appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

>[!NOTE] 说明 
>用户在使用过程中，请勿重复定义和在容器镜像中固定ASCEND\_VISIBLE\_DEVICES、ASCEND\_RUNTIME\_OPTIONS、ASCEND\_RUNTIME\_MOUNTS和ASCEND\_VNPU\_SPECS等环境变量。

**表 1** Ascend Docker Runtime运行参数解释

<a name="table5134121862415"></a>
|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>如果任务不需要使用NPU设备，可以设置ASCEND_VISIBLE_DEVICES环境变量的取值为void或为空。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void表示不使用Ascend Docker Runtime的挂载功能，不挂载NPU设备、驱动和文件目录。相关挂载参数也会失效。</li><li>挂载物理芯片（NPU）<ul><li>ASCEND_VISIBLE_DEVICES=0时，表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3时，表示将1、3号NPU设备挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=0-2时，表示将0号至2号NPU设备（包含0号和2号）挂载入容器中，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2。</li><li>ASCEND_VISIBLE_DEVICES=0-2,4时，表示将0号至2号以及4号NPU设备挂载入容器，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2,4。</li><li>ASCEND_VISIBLE_DEVICES=AscendXXX-Y，其中XXX表示NPU设备的型号，支持的取值为910，310、310B和310P；Y表示物理NPU设备ID。<ul><li>ASCEND_VISIBLE_DEVICES=Ascend910-1，表示把1号NPU设备挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=Ascend910-1,Ascend910-3，表示把1号NPU和3号NPU挂载进容器。</li></ul><div class="note"><span>[!NOTE] 说明</span><div class="notebody"><ul><li>NPU类型需要和实际环境的NPU类型保持一致，否则将会挂载失败。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=0，Ascend910-1。</li></ul></ul><li>挂载虚拟芯片（vNPU）<ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AI Core。<div class="note"><span>[!NOTE] 说明</span><div class="notebody"><ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS，表示在指定的NPU上划分出的AI Core数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul>|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2 智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2 开发者套件场景下需要指定该参数。<p>其他设备如Atlas 训练系列产品、Atlas A2 训练系列产品和Atlas 200I SoC A1 核心板等产品可以使用该参数，但因其默认挂载内容中不存在软链接，所以无需额外指定该参数。</p>|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2 智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2 开发者套件场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2 智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2 开发者套件将无法使用Ascend Docker Runtime。</li></ul>|
|ASCEND_RUNTIME_OPTIONS|对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：<ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul>|
|ASCEND_RUNTIME_MOUNTS|待挂载内容的配置文件名，该文件中可配置需要挂载到容器内的文件及目录。|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2<div class="note"><span>[!NOTE] 说明</span><div class="notebody"><ul><li>默认读取/etc/ascend-docker-runtime.d/base.list配置文件。</li><li>hostlog.list请根据实际自定义配置文件名修改。</li><li>支持读取多个自定义配置文件。</li><li>文件名必须小写，不能包含大写字母。</li></ul></li></ul>|
|ASCEND_VNPU_SPECS|从物理NPU设备中切分出一定数量的AI Core，指定为虚拟设备。支持的取值请参见<a href="./virtual_instance.md#虚拟化规则">表1</a>。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li></ul>|ASCEND_VNPU_SPECS=vir04表示切分4个AI Core作为虚拟设备，挂载至容器中。|



## 在K8s集成Containerd使用<a name="ZH-CN_TOPIC_0000002479227280"></a>

K8s集成Containerd场景下，用户需要安装Ascend Docker Runtime。

-   **（二选一）申请NPU卡资源的情况下**。使用任务YAML下发训练或推理任务时，NPU芯片的分配由Volcano和Ascend Device Plugin组件自动完成；NPU芯片及相关文件目录的挂载由Ascend Docker Runtime组件自动完成。示例如下。

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # 镜像名称根据实际情况修改
    ...
                resources:
                  limits:
                    huawei.com/Ascend910: 1     # 资源名称和数量根据实际情况修改
                  requests:
                    huawei.com/Ascend910: 1     #  资源名称和数量根据实际情况修改
    ...
    ```

-   **（二选一）如果未申请NPU资源**，由集成平台写入ASCEND\_VISIBLE\_DEVICES=void环境变量。示例如下。

    ```
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # 镜像名称根据实际情况修改
    ...
                env:
                - name: ASCEND_VISIBLE_DEVICES     # 未使用resources申请NPU资源时需增加此配置
                   value: "void"
    ...
    ```


