# 快速入门<a name="ZH-CN_TOPIC_0000002511346939"></a>

本文档提供两种快速入门场景，帮助您快速上手Ascend NPU集群调度：

- **10分钟快速入门**：仅部署Ascend Device Plugin和Ascend Docker Runtime，使用Kubernetes原生调度器调度普通Pod，快速验证NPU资源调度能力，适合初学者快速体验。
- **训练业务快速入门**：部署完整的集群调度组件（NodeD、Ascend Device Plugin、Ascend Docker Runtime、Volcano、ClusterD、Ascend Operator），以PyTorch训练任务为例，体验端到端的训练流程。

您可以根据实际需求选择合适的入门路径。

## 环境准备<a name="section159013591917"></a>

入门示例需要确保集群环境已经搭建完成。

- 所有节点已安装Kubernetes，支持的版本为1.17.x\~1.34.x。（如需安装Volcano组件，请安装1.19.x及以上版本的Kubernetes，具体Kubernetes版本请参见[Volcano官网中对应的Kubernetes版本](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility)）。如需获取软件包，请参见[Kubernetes社区](https://kubernetes.io/zh-cn/docs/setup/)。
- 所有节点已安装Docker，支持的版本为18.09.x\~28.5.1。如需获取软件包，请参见[Docker社区或官网](https://docs.docker.com/engine/install/)。
- 所有节点已经安装配套的固件与驱动。Atlas 800T A2 训练服务器固件和驱动安装步骤请参见《[Atlas A2 中心推理和训练硬件 NPU驱动和固件安装指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100568434/426cffd9)》。
- 检查主机上[npu-smi](https://support.huawei.com/enterprise/zh/doc/EDOC1100568421/426cffd9)以及[hccn_tool工具](https://support.huawei.com/enterprise/zh/doc/EDOC1100568362/426cffd9)是否可正常运行。
- 拉取镜像，下载组件安装包等可能需要网络环境，请自行确保网络正常或自行准备相关离线镜像包、组件安装包等。

  >[!NOTE]
  >
  >- 参见[《Ascend Training Solution 版本配套表》](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-training-solution-pid-258915853/software)，确认固件与驱动的版本与集群调度组件是否配套。
  >- NPU驱动和固件版本可通过**npu-smi info -t board -i** <i>NPU ID</i>命令查询。回显信息中的“Software Version”字段值表示NPU驱动版本，“Firmware Version”字段值表示NPU固件版本。

## 10分钟快速入门

本教程将指导您在 **10分钟内** 完成最简化的Ascend NPU集群调度环境搭建，仅使用：

- **Ascend Device Plugin**：NPU设备发现与资源上报
- **Ascend Docker Runtime**：NPU设备等资源挂载能力
- **Kubernetes原生调度器**：无需额外调度组件
- **普通Pod**：快速验证NPU调度能力

### 安装组件

下面以计算节点为Atlas 800T A2 训练服务器、CPU架构为AArch64为例。

1. 检查NPU状态，确保与服务器配套的NPU驱动已正确安装。

    ```shell
    npu-smi info
    ```

   若显示芯片信息，说明NPU驱动已正确安装。

2. 为NPU节点添加标签。

    执行以下命令，为**计算节点**创建节点标签（如节点名称为“worker01”）。

    ```shell
    kubectl label nodes worker01 workerselector=dls-worker-node
    ```

   >[!NOTE]
   >
   > `workerselector=dls-worker-node`标签用于标识计算节点，供集群调度组件（如NodeD、Ascend Device Plugin）识别并管理NPU资源。

3. 部署Ascend Docker Runtime和Ascend Device Plugin组件。

   >[!NOTE]
   >
   > `VERSION` 环境变量用于指定Ascend组件版本，本文档以`26.1.0`为例。每个独立的代码块中均需设置此变量。

    1. 部署Ascend Docker Runtime。

        ```shell
        VERSION=26.1.0
        mkdir -p /tmp/Ascend-docker-runtime
        cd /tmp/Ascend-docker-runtime
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        chmod +x Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        echo Y | ./Ascend-docker-runtime_${VERSION}_linux-aarch64.run --install
        systemctl daemon-reload && systemctl restart docker
        ```

       回显示例如下，表示安装成功。

        ```output
        Uncompressing ascend-docker-runtime  100%
        Please read the End User License Agreement carefully. Your use of the Huawei Software
        will be deemed as your acceptance of the constraints mentioned in the Agreement.
        The full text of the EULA is available at:
        https://www.hiascend.com/zh/legal/softlicense

        Do you accept the EULA to install Ascend-docker-runtime? [y/N]
        [INFO] user accepted EULA
        [INFO] installing ascend-docker-runtime
        ...
        [INFO] ascend-docker-runtime install success
        ```

    2. 拉取Ascend Device Plugin镜像。

        ```shell
        VERSION=26.1.0
        # 从华为云镜像仓拉取Ascend Device Plugin镜像
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v${VERSION}

        # 为镜像添加本地标签
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v${VERSION} ascend-k8sdeviceplugin:v${VERSION}
        ```

    3. 部署Ascend Device Plugin。

        ```shell
        VERSION=26.1.0
        # 拉取配置文件
        mkdir -p /tmp/devicePlugin
        cd /tmp/devicePlugin
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-mindxdl-device-plugin_${VERSION}_linux-aarch64.zip
        unzip Ascend-mindxdl-device-plugin_${VERSION}_linux-aarch64.zip

        # 部署Device Plugin，若VERSION低于26.1.0版本，yaml文件为device-plugin-910-v${VERSION}.yaml
        kubectl apply -f device-plugin-v${VERSION}.yaml
        ```

       查看Ascend Device Plugin Pod的状态。

        ```shell
        kubectl get pod -n kube-system
        ```

       回显示例如下，表示状态正常。

        ```output
        NAME                                  READY   STATUS    RESTARTS   AGE
        ...
        ascend-device-plugin-daemonset-d5ctz  1/1     Running   0          11s
        ...
        ```

    4. 查看节点的NPU资源。

        ```shell
        kubectl describe node -A | grep "huawei.com/Ascend910"
        ```

       回显示例如下，正常显示可用的NPU数量。

        ```output
        huawei.com/Ascend910:     8
        huawei.com/Ascend910:     8
        ```

### 调度NPU Pod

1. 创建测试Pod配置文件 `npu-test-pod.yaml`。

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: npu-test
    spec:
      containers:
      - name: npu-container
        image: ubuntu:22.04          # 测试pod镜像，可以自定义
        command: ["/bin/bash", "-c", "sleep 3600"]
        resources:
          limits:
            huawei.com/Ascend910: 1  # 请求1个NPU卡
          requests:
            huawei.com/Ascend910: 1
    ```

2. 部署测试Pod。

    ```shell
    kubectl apply -f npu-test-pod.yaml
    ```

3. 验证Pod调度。

    ```shell
    # 查看Pod状态
    kubectl get pods npu-test -o wide

    # 预期输出（STATUS为Running表示调度成功）
    NAME      READY   STATUS    RESTARTS   AGE   IP           NODE      NOMINATED NODE
    npu-test  1/1     Running   0          10s   10.244.1.2   worker01  <none>
    ```

4. 验证NPU访问。

    ```shell
    # 进入容器验证NPU可用性
    kubectl exec -it npu-test  -- /bin/bash

    # 在容器内执行npu-smi info命令正确显示芯片信息
    export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:${LD_LIBRARY_PATH}
    npu-smi info
    ```

5. 清理测试资源。

    ```shell
    VERSION=26.1.0
    # 删除测试Pod
    kubectl delete pod npu-test

    # 删除Ascend Device Plugin，若VERSION低于26.1.0版本，yaml文件为device-plugin-910-v${VERSION}.yaml
    kubectl delete -f device-plugin-v${VERSION}.yaml
    ```

**常见问题**

| 问题 | 原因 | 解决方法 |
|------|------|---------|
| Pod一直Pending | NPU资源不足或节点标签不匹配 | 检查`kubectl describe pod`和节点标签 |
| Ascend Device Plugin启动失败 | 驱动路径不正确 | 检查`/usr/local/Ascend/driver`是否存在 |

## 训练业务快速入门

本章节依然以一台Atlas 800T A2 训练服务器、CPU架构为AArch64为例，指导开发者快速完成NodeD、Ascend Device Plugin、Ascend Docker Runtime、Volcano、ClusterD、Ascend Operator组件的安装及使用整卡调度特性快速下发训练任务。

### 操作说明<a name="section17940333114314"></a>

**表 1**  关键步骤说明

|操作步骤|操作说明|更多参考|
|--|--|--|
|[安装组件](#section1837511531098)|以Atlas 800T A2 训练服务器为例，手把手带您在昇腾设备上快速安装集群调度组件。|更多安装集群调度组件的参数说明和操作步骤，请参考[安装部署](../03_installation_guide/02_installation/00_helm_installation.md)章节。|
|[下发训练任务](#section106493419399)|以一个简单的PyTorch训练任务为例，让您快速了解训练任务下发的操作流程。|更多下发训练任务的参数说明和操作步骤，请参考[基础调度](../04_usage/03_basic_scheduling/00_feature_description.md)章节。|

### 安装组件<a name="section1837511531098"></a>

以下步骤命令均以一台Atlas 800T A2 训练服务器为例，如需了解所有组件的详细安装步骤和参数说明请参见[安装部署](../03_installation_guide/02_installation/00_helm_installation.md)。

1. 创建节点标签。

    执行以下命令，为**计算节点**创建节点标签（如节点名称为“worker01”）。

    ```shell
    kubectl label nodes worker01 node-role.kubernetes.io/worker=worker workerselector=dls-worker-node masterselector=dls-master-node --overwrite
    ```

2. 安装组件。以AArch64架构为例，用户需根据实际情况下载对应架构的软件包。
    >[!NOTE]
    >
    >快速入门以Helm快捷部署为例，要求MindCluster版本为26.1.0及以上，可以参考安装部署章节的[使用Helm安装](../03_installation_guide/02_installation/00_helm_installation.md)。

    1. 安装Ascend Docker Runtime。

        ```shell
        VERSION=26.1.0
        mkdir -p /tmp/Ascend-docker-runtime
        cd /tmp/Ascend-docker-runtime
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        chmod +x Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        echo Y | ./Ascend-docker-runtime_${VERSION}_linux-aarch64.run --install
        systemctl daemon-reload && systemctl restart docker
        ```

    2. 通过Helm安装NodeD、Ascend Device Plugin、Volcano、ClusterD、Ascend Operator组件。

        ```shell
        VERSION=26.1.0
        mkdir /tmp/helm
        cd /tmp/helm
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-helm-deploy-tool_${VERSION}_linux.zip
        unzip Ascend-helm-deploy-tool_${VERSION}_linux.zip
        helm install mindcluster-crds mindcluster-crds-deploy-tool-*.tgz
        helm install mindcluster mindcluster-deploy-tool-*.tgz
        ```

        回显示例如下，表示安装成功。

        ```output
        Release "mindcluster-crds" does not exist. Installing it now.
        NAME: mindcluster-crds
        LAST DEPLOYED: ...
        NAMESPACE: mindx-dl
        STATUS: deployed
        REVISION: 1
        TEST SUITE: None
        ```

        ```output
        Release "mindcluster" does not exist. Installing it now.
        NAME: mindcluster
        LAST DEPLOYED: ...
        NAMESPACE: mindx-dl
        STATUS: deployed
        REVISION: 1
        TEST SUITE: None
        ```

    3. 验证组件是否正常运行，以NodeD组件为例。

        ```shell
        kubectl get pod -n mindx-dl
        ```

        回显示例如下，表示NodeD组件运行正常。

        ```shell
        NAME                                  READY   STATUS    RESTARTS   AGE
        ...
        noded-694474f599-54w6b                1/1     Running   0          11s
        ...
        ```

### 下发训练任务<a name="section106493419399"></a>

1. 准备镜像。

   从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)下载24.0.X版本的ascend-pytorch训练镜像。镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

   >[!NOTE]
   >
   > 本示例使用的镜像版本为24.0.0-A2-2.1.0。如需获取最新版本镜像，请访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)查看可用版本列表，或联系华为技术支持获取版本配套信息。

    ```shell
    docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-pytorch:24.0.0-A2-2.1.0-ubuntu20.04
    docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-pytorch:24.0.0-A2-2.1.0-ubuntu20.04 ascend-pytorch:24.0.0-A2-2.1.0-ubuntu20.04
    ```

2. 准备训练任务。

    1. 执行以下命令，下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50_ID4149_for_PyTorch”作为训练代码，并解压到“/data/atlas_dls/public/code/”路径下。

        ```shell
        mkdir -p /data/atlas_dls/public/code/
        cd /data/atlas_dls/public/code/
        wget https://raw.gitcode.com/Ascend/ModelZoo-PyTorch/archive/refs/heads/master.zip?path=PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch -O ResNet50_ID4149_for_PyTorch.zip
        unzip ResNet50_ID4149_for_PyTorch.zip
        mv ModelZoo-PyTorch-master-PyTorch-built-in-cv-classification-ResNet50_ID4149_for_PyTorch/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch ResNet50_ID4149_for_PyTorch
        ```

    2. 执行以下命令，获取[MindCluster-Samples](https://gitcode.com/Ascend/mindcluster-deploy)仓库的“samples/train/basic-training/without-ranktable/pytorch”目录中的train_start.sh，放在“/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts”路径下。

        ```shell
        mkdir /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts
        cd /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts
        wget https://raw.gitcode.com/Ascend/mindcluster-deploy/raw/master/samples/train/basic-training/without-ranktable/pytorch/train_start.sh
        ```

    3. 执行以下命令，获取[MindCluster-Samples](https://gitcode.com/Ascend/mindcluster-deploy)仓库“samples/train/basic-training/without-ranktable/pytorch”目录下的“pytorch_standalone_acjob_quickstart.yaml”文件。示例默认为单机单卡任务。

        ```shell
        cd /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts
        wget https://raw.gitcode.com/Ascend/mindcluster-deploy/raw/master/samples/train/basic-training/without-ranktable/pytorch/pytorch_standalone_acjob_quickstart.yaml
        ```

    4. （可选）准备数据集。pytorch_standalone_acjob_quickstart.yaml中默认设置了`--dummy`参数，能自动为训练任务生成随机数据集，无需真实数据集即可启动训练任务。若用户需要使用真实数据集，请删掉此yaml文件中的`--dummy`参数，然后自行准备ResNet-50对应的数据集，使用时请遵守对应规范，将数据集上传到”/data/atlas_dls/public/dataset/resnet50/imagenet“。

        ```shell
        mkdir /data/atlas_dls/public/dataset/resnet50/imagenet
        cd /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

3. 下发单机单卡任务。

    ```shell
    kubectl apply -f /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts/pytorch_standalone_acjob_quickstart.yaml
    ```

4. 查看Pod运行情况。

    ```shell
    kubectl get pod -A -o wide
    ```

   回显示例如下，出现Running表示任务正常运行。

   >[!NOTE]
   >
   > 回显中`192.168.244.xxx`为Pod分配的实际IP地址，`worker01`为实际节点名称，请以实际回显为准。

    ```output
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE     IP                NODE      NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          6s      192.168.244.xxx   worker01   <none>           <none>
    ```

   >[!NOTE]
   >
   >若下发训练任务后，任务一直处于Pending状态，可以参见[训练任务处于Pending状态，原因：nodes are unavailable](https://gitcode.com/Ascend/mind-cluster/issues/352)或者[资源不足时，任务处于Pending状态](https://gitcode.com/Ascend/mind-cluster/issues/355)章节进行处理。

5. 查看训练结果。

    1. 在任意节点执行如下命令，查看训练结果。

        ```shell
        kubectl logs -n default default-test-pytorch-master-0
        ```

    2. 查看训练日志，如果出现如下内容表示训练成功。

       >[!NOTE]
       >
       > 回显中`10.106.227.xxx`为集群分配的实际IP地址，请以实际回显为准。

        ```output
        [20260724-11:16:23] [MindXDL Service Log]Training start at 2026-07-24-11:16:23
        /usr/local/python3.9.2/lib/python3.9/site-packages/torchvision/io/image.py:13: UserWarning: Failed to load image Python extension: 'libc10_cuda.so: cannot open shared object file: No such file or directory'If you don't plan on using image functionality from `torchvision.io`, you can ignore this warning. Otherwise, there might be something wrong with your environment. Did you have `libjpeg` or `libpng` installed before building `torchvision` from source?
          warn(
        /job/code/main.py:215: UserWarning: You have chosen to seed training. This will turn on the CUDNN deterministic setting, which can slow down your training considerably! You may see unexpected behavior when restarting from checkpoints.
          warnings.warn('You have chosen to seed training. '
        /job/code/main.py:222: UserWarning: You have chosen a specific GPU. This will completely disable data parallelism.
          warnings.warn('You have chosen a specific GPU. This will completely '
        Use GPU: 0 for training
        => creating model 'resnet50'
        ```

6. 清理测试资源。

    ```shell
    # 删除训练任务Pod
    kubectl delete -f /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts/pytorch_standalone_acjob_quickstart.yaml

    # 卸载Helm部署的组件（可选）
    helm uninstall mindcluster
    helm uninstall mindcluster-crds
    ```
