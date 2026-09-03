# k8s-rdma-shared-dev-plugin

# 支持的产品形态

支持以下设备使用资源监测

- UB 类型的 RDMA 网卡

# 组件介绍

设备管理插件拥有以下功能：

- 设备发现：支持发现计算节点上的 UB RDMA 设备，将发现的设备上报到 Kubernetes 系统中。支持共享与独占两种工作模式，共享模式下将节点上的 RDMA 设备作为共享资源上报；独占模式下按 NPU 与 DPU 的映射关系为 Pod 分配节点上发现的真实 UB 设备。
- 健康检查：支持检测 UB 设备的健康状态，当设备处于不健康状态时，将故障信息写入 Kubernetes ConfigMap，不健康设备不会被分配给 Pod。故障检测周期可通过 `faultDetectPeriod` 参数配置。
- 设备分配：支持在 Kubernetes 系统中分配 RDMA 设备；独占模式下按 NPU 与 DPU 的映射关系为 Pod 分配设备，并将分配结果写入 `k8s.v1.cni.cncf.io/device-status` 注解，供 Multus CNI 与 UB Host Device CNI 使用完成设备挂载。

# 编译

1. 通过git拉取源码，并切换master分支，获得k8s-rdma-shared-dev-plugin。

    示例：源码放在/home/mind-cluster/component/k8s-rdma-shared-dev-plugin目录下

2. 执行以下命令，进入构建目录，执行构建脚本，在“output”目录下生成二进制组件包、yaml文件、Dockerfile等文件。

    **cd** _/home/mind-cluster/component/_**k8s-rdma-shared-dev-plugin/build/**

    ```bash
    chmod +x build.sh
    ./build.sh
    ```

3. 执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/mind-cluster/component/_**k8s-rdma-shared-dev-plugin/output**

    ```text
    k8s-rdma-shared-dp
    k8s-rdma-shared-dp-v26.1.0.yaml
    config.json
    Dockerfile
    Dockerfile.openeuler
    agreement.txt
    fault_code.json
    fault_detection.sh
    npu-nic-mapping.json
    ```

4. 执行以下命令，根据基础镜像选择对应的Dockerfile构建设备插件镜像（在**output**目录下执行）。

   - 使用Ubuntu 22.04基础镜像

      ```bash
      docker build -f Dockerfile -t k8s-rdma-shared-dev-plugin:ubuntu22.04 .
      ```

   - 使用openEuler 24.03基础镜像

      ```bash
      docker build -f Dockerfile.openeuler -t k8s-rdma-shared-dev-plugin:openeuler24.03 .
      ```

   **说明：**
        1、Dockerfile 使用 Ubuntu 22.04 基础镜像，Dockerfile.openeuler 使用 openEuler 24.03 基础镜像。
        2、镜像构建前需先执行build.sh，生成k8s-rdma-shared-dp文件。

# 说明

当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount，该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。
