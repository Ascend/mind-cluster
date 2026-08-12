# 在CRI-O客户端使用<a name="ZH-CN_TOPIC_0000002511347204"></a>

## 使用说明<a name="section0966931165318"></a>

- Ascend Docker Runtime 支持以 CRI-O 为容器引擎，配合 Kubernetes（或 crictl）使能 NPU 容器化。安装与卸载请参考 [Ascend Docker Runtime 手动安装](../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md#zh-cn_topic_0000001930317932_section_crio_install) 中的“CRI-O场景下安装”章节与 [手动卸载](../../05_developer_guide/00_installation_deployment/02_uninstallation.md#section6134163311244) 章节。
- 支持挂载物理芯片与虚拟芯片。挂载虚拟芯片前请参考 [创建vNPU](../02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md) 章节完成虚拟化操作。
- 可通过 **ls /dev/davinci\*** 查询当前可用的物理芯片 ID；通过 **ls /dev/vdavinci\*** 查询当前可用的虚拟芯片 ID。
- 启动参数（`ASCEND_VISIBLE_DEVICES`、`ASCEND_RUNTIME_OPTIONS`、`ASCEND_RUNTIME_MOUNTS`、`ASCEND_VNPU_SPECS` 等）的取值与 Docker/Containerd 场景完全一致，参数解释请参考 [在Containerd客户端使用](./03_usage_on_the_containerd_client.md#table5134121862415) 中的表1。

## 使用示例<a name="section148905517123"></a>

CRI-O 作为守护进程运行，容器由 kubelet 经 CRI 接口下发；NPU 由 Ascend Device Plugin 经 `ASCEND_VISIBLE_DEVICES` 注入。示例中的 image-name:tag 为镜像名称与标签，如“ascend-pytorch:pytorch\_TAG”。

- 示例1：确认 CRI-O 已注册 ascend 运行时。

    ```shell
    crio status | grep -i runtime
    crictl info | grep -i runtime
    ```

    回显中应包含 `default_runtime = "ascend"` 且 `runtime_path` 指向 ascend-docker-runtime 可执行文件。

- 示例2：Kubernetes Pod 调度时挂载物理芯片 ID 为 0 的 NPU。

    通过 Ascend Device Plugin 暴露的 NPU 资源（如 `huawei.com/Ascend910`，资源名以 Device Plugin 实际配置为准）请求设备，Device Plugin 会自动注入 `ASCEND_VISIBLE_DEVICES=0`：

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: npu-pod
    spec:
      containers:
      - name: npu-container
        image: {image-name:tag}
        command: ["bash"]
        resources:
          limits:
            huawei.com/Ascend910: 1
    ```

- 示例3：仅挂载 NPU 设备和管理设备，不挂载驱动相关目录。

    在 Pod 环境变量中设置 `ASCEND_RUNTIME_OPTIONS=NODRV`（由 Device Plugin 透传），效果与 Containerd 场景一致。

- 示例4：挂载虚拟芯片 ID 为 100 的芯片。

    通过 Device Plugin 注入 `ASCEND_VISIBLE_DEVICES=100 --env ASCEND_RUNTIME_OPTIONS=VIRTUAL`，取值规则与 Docker/Containerd 场景一致。

容器启动后，可执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录参考 [Ascend Docker Runtime默认挂载内容](../../07_references/05_appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

可能的输出结果如下：

```text
davinci0
davinci_manager
devmm_svm
hisi_hdc
include lib64
dcmi
npu-smi
```

>[!NOTE]
> 用户在使用过程中，请勿重复定义或在容器镜像中固定 `ASCEND_VISIBLE_DEVICES`、`ASCEND_RUNTIME_OPTIONS`、`ASCEND_RUNTIME_MOUNTS` 和 `ASCEND_VNPU_SPECS` 等环境变量。
> 断点续训（checkpoint/restore）当前依赖 containerd snapshotter 导出 rootfs 差异，CRI-O 场景暂不支持，规划在后续版本提供。

启动命令相关参数如 [表1](./03_usage_on_the_containerd_client.md#table5134121862415) 所示。
