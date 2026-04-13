# 在Containerd客户端使用<a name="ZH-CN_TOPIC_0000002511347203"></a>

**使用说明<a name="section0966931165317"></a>**

- Ascend Docker Runtime支持挂载物理芯片，同时支持挂载虚拟芯片。挂载虚拟芯片前需要参考[创建vNPU](../virtual_instance/virtual_instance_with_hdk/04_creating_vnpu.md)章节，对物理芯片进行虚拟化操作，支持对物理芯片进行静态虚拟化和动态虚拟化。
- 可通过<b>ls /dev/davinci\*</b>命令查询当前可用的物理芯片ID；通过<b>ls /dev/vdavinci\*</b>命令查询当前可用的虚拟芯片ID。
- 若用户不需要挂载Ascend Docker Runtime的默认配置文件“/etc/ascend-docker-runtime.d/base.list”中所有内容，可创建自定义配置文件（例如hostlog.list），减少挂载内容，具体操作请参考[（可选）配置自定义挂载内容](./01_configuring_custom_mounted_content.md)章节。

**使用示例<a name="section148905517122"></a>**

示例中的image-name:tag为镜像名称与标签，如“ascend-tensorflow:tensorflow\_TAG”。containerID为容器ID，使用ctr启动容器需要指定容器ID，如“c1”。

- 示例1：启动容器时，挂载物理芯片ID为0的芯片。

    ```shell
    ctr run --runtime io.containerd.runc.v2 --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime-t --env ASCEND_VISIBLE_DEVICES=0 image-name:tag containerID bash
    ```

- 示例2：启动容器时，仅挂载NPU设备和管理设备，不挂载驱动相关目录。

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime-t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_OPTIONS=NODRV image-name:tag containerID bash
    ```

- 示例3：启动容器时，挂载物理芯片ID为0的芯片，读取自定义配置文件hostlog中的挂载内容。

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime-t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_RUNTIME_MOUNTS=hostlog image-name:tag containerID bash
    ```

- 示例4：启动容器时，挂载虚拟芯片ID为100的芯片。

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime-t --env ASCEND_VISIBLE_DEVICES=100 --env ASCEND_RUNTIME_OPTIONS=VIRTUAL image-name:tag containerID bash
    ```

- 示例5：启动容器时，从物理芯片ID为0的芯片上，切分出4个AI Core作为虚拟设备并挂载至容器中。

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime-t --env ASCEND_VISIBLE_DEVICES=0 --env ASCEND_VNPU_SPECS=vir04 image-name:tag containerID bash
    ```

- 示例6：启动容器时，挂载物理芯片ID为0的芯片，并且允许挂载的驱动文件中存在软链接（仅适用于Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景）。

    ```shell
    ctr run --runc-binary /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime-t --env ASCEND_VISIBLE_DEVICES=0 -e ASCEND_ALLOW_LINK=True image-name:tag containerID bash
    ```

启动命令相关参数如[表1](#table5134121862415)所示。

容器启动后，可执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录参考[Ascend Docker Runtime默认挂载内容](../../appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

>[!NOTE] 
>用户在使用过程中，请勿重复定义和在容器镜像中固定ASCEND\_VISIBLE\_DEVICES、ASCEND\_RUNTIME\_OPTIONS、ASCEND\_RUNTIME\_MOUNTS和ASCEND\_VNPU\_SPECS等环境变量。

**表 1** Ascend Docker Runtime运行参数解释

<a name="table5134121862415"></a>

|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>如果任务不需要使用NPU设备，可以设置ASCEND_VISIBLE_DEVICES环境变量的取值为void或为空。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void表示不使用Ascend Docker Runtime的挂载功能，不挂载NPU设备、驱动和文件目录。相关挂载参数也会失效。</li><li>挂载物理芯片（NPU）<ul><li>ASCEND_VISIBLE_DEVICES=0时，表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3时，表示将1、3号NPU设备挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=0-2时，表示将0号至2号NPU设备（包含0号和2号）挂载入容器中，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2。</li><li>ASCEND_VISIBLE_DEVICES=0-2,4时，表示将0号至2号以及4号NPU设备挂载入容器，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2,4。</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y，其中XXX表示NPU设备，支持的取值为npu、Ascend910、Ascend310、Ascend310B和Ascend310P；Y表示物理NPU设备ID。<ul><li>ASCEND_VISIBLE_DEVICES=npu-1，表示把1号NPU设备挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=npu-1,npu-3，表示把1号NPU和3号NPU挂载进容器。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用芯片名称指定设备时，建议统一取值npu。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=0，npu-1。</li></ul></div></div></li></ul></li></ul><li>挂载虚拟芯片（vNPU）<ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AI Core。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS，表示在指定的NPU上划分出的AI Core数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul></div></div></li></ul></li>|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2 智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2 开发者套件场景下需要指定该参数。<p>其他设备如Atlas 训练系列产品、Atlas A2 训练系列产品和Atlas 200I SoC A1 核心板等产品可以使用该参数，但因其默认挂载内容中不存在软链接，所以无需额外指定该参数。</p>|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2 智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2 开发者套件场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2 智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2 开发者套件将无法使用Ascend Docker Runtime。</li></ul>|
|ASCEND_RUNTIME_OPTIONS|对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：<ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul>|
|ASCEND_RUNTIME_MOUNTS|待挂载内容的配置文件名，该文件中可配置需要挂载到容器内的文件及目录。|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>默认读取/etc/ascend-docker-runtime.d/base.list配置文件。</li><li>hostlog.list请根据实际自定义配置文件名修改。</li><li>支持读取多个自定义配置文件。</li><li>文件名必须小写，不能包含大写字母。</li></ul></div></div></li></ul>|
|ASCEND_VNPU_SPECS|从物理NPU设备中切分出一定数量的AI Core，指定为虚拟设备。支持的取值请参见[表1](../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md)。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li></ul>|ASCEND_VNPU_SPECS=vir04表示切分4个AI Core作为虚拟设备，挂载至容器中。|
