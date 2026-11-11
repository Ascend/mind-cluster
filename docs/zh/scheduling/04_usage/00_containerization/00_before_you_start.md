# 使用前必读<a name="ZH-CN_TOPIC_0000002511427169"></a>

容器化支持是一种将应用程序及其依赖项打包到一个独立、可移植的环境（容器）中的技术支持。

## 前提条件<a name="section1632062465010"></a>

- 在使用容器化支持特性前，需要确保Ascend Docker Runtime组件已经安装，若没有安装，可以参考[安装部署](../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md)章节进行操作。
<!-- npu="950" id2 -->
- （可选）在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下，如需挂载UB网卡设备及驱动，还需满足以下条件：
  - 已安装Ascend Docker Runtime，并使用Legacy注入模式。安装方法请参见[Ascend Docker Runtime](../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md)。
  - 宿主机搭载UB网卡，已安装并加载UB网卡驱动；已安装版本为openEuler 24.03 LTS SP4的操作系统。
  - 确保容器内`glibc`版本不低于宿主机版本，`rdma-core`版本与宿主机一致。在宿主机和容器内执行`rpm -qa | grep rdma-core`命令查询`rdma-core`的版本，不满足要求时，需先准备兼容镜像。
<!-- end id2 -->

## 使用说明<a name="section44381612353"></a>

- 容器化支持可以和训练场景下的所有特性一起使用，也可以和推理场景的所有特性一起使用。
- 若使用Volcano进行任务调度，则不建议通过Docker或Containerd指令创建/挂载NPU卡的容器，否则可能会触发Volcano调度问题。
- 容器化支持会自动挂载宿主机的二进制、动态库文件到容器中，可能会涉及glibc依赖，此时容器OS的glibc版本需>=宿主机OS的glibc版本，以便挂载到容器内可以正常运行。
<!-- npu="950" id1 -->
- 在Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器、Atlas 950 SuperPoD超节点等存在UB总线的产品形态下，会默认挂载ub_driver.list中的UB驱动相关内容。ub_driver.list中的文件依赖宿主机的glibc版本，目前支持的OS版本和glibc版本如下：
  - 宿主机：openEuler24.03（glibc2.38）
  - 容器内：openEuler24.03（glibc2.38）、Ubuntu24.04（glibc2.39）。
- 在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下，支持通过环境变量`ASCEND_RDMA_DEVICES`指定待挂载RDMA设备的HCA名称（取值为UB网卡在“/sys/class/infiniband/”下的完整HCA名称，如`hrn5_0,hrn5_1`），将对应的UB网卡字符设备和驱动文件挂载到容器中。`ASCEND_RDMA_DEVICES`与`ASCEND_VISIBLE_DEVICES`独立配置，支持仅挂载NPU设备、仅挂载RDMA设备，或同时挂载NPU和RDMA设备：
  - 仅支持Legacy模式；CDI模式会忽略`ASCEND_RDMA_DEVICES`，不会根据该参数自动注入RDMA设备及挂载对应驱动文件。在CDI模式下，该参数去除首尾空白后既非空也非`void`时，会记录警告，不会因此终止容器创建。
  - 每个所选HCA的“device/infiniband_verbs/”目录中，必须有一个`uverbsX`的`ibdev`内容与该HCA名称一致；同一目录可以包含对应不同HCA的多个`uverbsX`。Ascend Docker Runtime据此挂载对应的“/dev/infiniband/uverbsX”字符设备（宿主机存在“/dev/infiniband/rdma_cm”时一并挂载），不会挂载整个“/dev/infiniband”目录。
  - 默认读取“/etc/ascend-docker-runtime.d/rdma_driver.list”，将如下UB网卡驱动配置文件和动态库以只读方式挂载至容器中（默认配置文件无需用户修改）：

    ```text
    /etc/libibverbs.d/hrn5.driver
    /usr/lib64/libibverbs/libhrn5-rdmav34.so
    /usr/lib64/libhrn5-rdmav34.so
    /usr/lib64/libhrn5.so
    /usr/lib64/libhrn5.so.1
    /usr/lib64/libibv_extend.so
    /usr/lib64/libibv_extend.so.4
    /usr/lib64/libibv_extend.so.4.2.0
    ```

    >[!NOTE]
    >
    >- 默认清单中，除“/usr/lib64/libibverbs/libhrn5-rdmav34.so”在宿主机不存在时会跳过挂载外，其余驱动文件缺失、软链接无效，或所选HCA及其`uverbsX`设备节点不存在、校验失败时，均会报错并终止容器创建。
    >- 挂载RDMA驱动文件时会自动允许合法软链接（软链接及其目标文件属主须为root），无需额外指定`ASCEND_ALLOW_LINK=True`。
    >- 指定`ASCEND_RUNTIME_OPTIONS=NODRV`时，跳过普通集合及RDMA集合中文件和目录的自动挂载（仍注入所选字符设备并校验宿主机RDMA依赖），用户可通过容器卷（如Docker `--mount`/`-v`、Containerd `--mount`、CRI-O `mounts`或Kubernetes `volumeMounts`）按需挂载驱动文件。

<!-- end id1 -->

## 支持的产品形态<a name="section169961844182917"></a>

支持以下产品使用容器化支持。

- <term>Atlas训练系列产品</term>
- <term>Atlas A2训练系列产品</term>
- <term>Atlas A3训练系列产品</term>
- 推理服务器（插Atlas 300I推理卡）
- <term>Atlas 200/300/500推理产品</term>
- <term>Atlas 200I/500 A2推理产品</term>
- <term>Atlas推理系列产品</term>
- Atlas 800I A2推理服务器
- A200I A2 Box异构组件
- Atlas 800I A3超节点服务器
- <term>Ascend 950PR&950DT系列产品</term>

## 使用场景<a name="section124697813416"></a>

Ascend Docker Runtime组件支持在以下3种场景下使用容器化支持功能。

- [在Docker客户端使用](./02_usage_on_the_docker_client.md)
- [在Containerd客户端使用](./03_usage_on_the_containerd_client.md)
- [通过crictl命令行工具使用CRI-O](./04_usage_on_the_crio_client.md)
