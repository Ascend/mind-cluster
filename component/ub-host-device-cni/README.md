# ub-host-device-cni

# 支持的产品形态

- 支持以下产品使用本插件：
    - 华为 DPU（UB 总线）网卡

# 组件介绍

ub-host-device-cni 是一个基于 CNI（Container Network Interface）规范的主机设备挂载插件，在 host-device 插件基础上扩展了对华为 DPU（UB 总线）设备的支持，拥有以下功能：

- 设备挂载：将主机网络设备（如 UB 总线上的 DPU 网卡）挂载进容器网络命名空间，容器内直接使用宿主设备，性能无损。
- 设备识别：
    - UB 模式下支持两种设备来源，优先级为 **NAD `device`（主机网卡名） > `runtimeConfig.deviceID`（kubelet/Multus 注入的 UB 设备地址）**。
- IP 继承：开启 `inheritHostIP` 后，挂载的 UB 接口保留宿主机原有 IP 地址，不额外申请新 IP。
- IPAM 分配：未开启 IP 继承时，支持配置 IPAM 插件为挂载的接口分配 IP 地址、路由。
- 资源调度联动：配合 device-plugin 在 NAD 中声明 `huawei.com/ub_rdma` 资源，由 Multus 通过 `capabilities: { "deviceID": true }` 将分配的设备地址注入 `runtimeConfig.deviceID`。

# 配置说明

本插件通过 Kubernetes 的 NetworkAttachmentDefinition（NAD）进行配置，典型示例如下（UB 模式、继承宿主机 IP）：

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: roce-network
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: "huawei.com/ub_rdma"
spec:
  config: |
    {
      "cniVersion": "0.3.1",
      "name": "roce-network",
      "type": "ub-host-device",
      "ubMode": true,
      "inheritHostIP": true,
      "capabilities": { "deviceID": true }
    }
```

## 参数说明

| 参数 | 是否必填 | 说明 |
| --- | --- | --- |
| `type` | 是 | 插件类型，固定为 `ub-host-device` |
| `ubMode` | 否 | 是否开启 UB（DPU）模式，默认 false；开启后设备来源为 NAD `device` 或 `runtimeConfig.deviceID` |
| `inheritHostIP` | 否 | 是否继承宿主机网卡 IP，默认 false；开启后不再调用 IPAM 分配 IP |
| `capabilities` | 否 | 声明插件支持的运行时能力，`deviceID: true` 表示接受 Multus 注入的设备地址，作为未配置 NAD `device` 时的回退设备来源 |
| `device` | 否 | UB 模式下的主机网卡名（如 `eth0`），或非 UB 模式下的目标设备名 |
| `ipam` | 否 | IPAM 插件配置，用于为挂载的接口分配 IP/路由；开启 `inheritHostIP` 时无需配置 |

> **说明：**
>
> 1. 需要 Multus CNI 注入设备地址时，NAD 中必须声明 `capabilities: { "deviceID": true }`，并在 NAD 注解中声明 `k8s.v1.cni.cncf.io/resourceName`。
> 2. UB 模式下设备优先级：NAD `device` > `runtimeConfig.deviceID`；两者均未配置时，ADD 流程将报错。

# 编译

1. 通过 git 拉取源码，并切换 master 分支，获得 ub-host-device-cni。

    示例：源码放在 `/home/mind-cluster/component/ub-host-device-cni` 目录下

2. 执行以下命令，进入构建目录，执行构建脚本，在 `output` 目录下生成二进制 ub-host-device。

    **cd** _/home/mind-cluster/component/_**ub-host-device-cni/build/**

    ```shell
    chmod +x build.sh

    ./build.sh
    ```

3. 执行以下命令，查看 **output** 生成的软件列表。

    **ll** _/home/mind-cluster/component/_**ub-host-device-cni/output**

    ```output
    drwxr-xr-x 2 root root     4096  4月 29 09:28 ./
    drwxr-xr-x 6 root root     4096  4月 29 09:28 ../
    -r-x------ 1 root root 59349656  4月 29 09:28 ub-host-device*
    ```

4. 执行以下命令，进行单元测试与覆盖率分析，测试结果输出到 `test` 目录。

    **cd** _/home/mind-cluster/component/_**ub-host-device-cni/build/**

    ```shell
    chmod +x test.sh

    ./test.sh
    ```

    > **说明：**
    >
    > 1. `test.sh` 仅支持在 Linux 环境下执行（插件依赖 netlink 与 netns）。
    > 2. 覆盖率结果小于 50% 时，测试失败；可通过 `test/api.html` 查看代码覆盖明细。

# 说明

1. 本插件为 CNI 网络插件，需通过 CNI 配置（如 NAD/Multus）加载，非独立部署组件。
2. 插件仅支持 Linux 操作系统，依赖 netlink 与网络命名空间（netns）能力。
3. UB 模式下，`runtimeConfig.deviceID` 由 kubelet/Multus 依据 device-plugin 的资源分配结果注入。
