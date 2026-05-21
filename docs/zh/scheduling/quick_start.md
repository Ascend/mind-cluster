# 快速入门<a name="ZH-CN_TOPIC_0000002511346939"></a>
本文档提供两种快速入门场景，帮助您快速上手Ascend NPU集群调度：

- **10分钟极简入门**：仅部署Ascend Device Plugin，使用Kubernetes原生调度器调度普通Pod，快速验证NPU资源调度能力，适合初学者快速体验。
- **完整训练业务入门**：部署完整的集群调度组件（NodeD、Ascend Device Plugin、Ascend Docker Runtime、Volcano、ClusterD、Ascend Operator），以PyTorch训练任务为例，体验端到端的分布式训练流程。

您可以根据实际需求选择合适的入门路径。

## 环境准备<a name="section159013591917"></a>

入门示例需要确保集群环境已经搭建完成。

- 所有节点已安装Kubernetes，支持的版本为1.17.x\~1.34.x。（如需安装Volcano组件，请安装1.19.x及以上版本的Kubernetes，具体Kubernetes版本请参见[Volcano官网中对应的Kubernetes版本](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility)）。如需获取软件包，请参见[Kubernetes社区](https://kubernetes.io/zh-cn/docs/setup/)。
- 所有节点已安装Docker，支持的版本为18.09.x\~28.5.1。如需获取软件包，请参见[Docker社区或官网](https://docs.docker.com/engine/install/)。
- 所有节点已经安装配套的固件与驱动。Atlas 800T A2 训练服务器固件和驱动安装步骤请参见《[Atlas A2 中心推理和训练硬件 25.5.0 NPU驱动和固件安装指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100540370/426cffd9)》。
- 检查主机上[npu-smi](https://support.huawei.com/enterprise/zh/doc/EDOC1100540371/426cffd9?idPath=23710424|251366513|22892968|252309113|254184887)以及[hccn_tool工具](https://support.huawei.com/enterprise/zh/doc/EDOC1100540101/426cffd9?idPath=23710424|251366513|254884019|261408772|261457531)是否可正常运行。

    >[!NOTE]
    >
    >- 参见[《Ascend Training Solution 版本配套表》](https://support.huawei.com/enterprise/zh/ascend-computing/ascend-training-solution-pid-258915853/software)，确认固件与驱动的版本与集群调度组件是否配套。
    >- NPU驱动和固件版本可通过**npu-smi info -t board -i** <i>NPU ID</i>命令查询。回显信息中的“Software Version”字段值表示NPU驱动版本，“Firmware Version”字段值表示NPU固件版本。
    >- 下文的<i>\{xxx\}</i>即取“910”字符作为芯片型号数值。


## 10分钟快速入门

### 概述

本教程将指导您在 **10分钟内** 完成最简化的Ascend NPU集群调度环境搭建，仅使用：
- **Ascend Device Plugin** - NPU设备发现与资源上报
- **Kubernetes原生调度器** - 无需额外调度组件
- **普通Pod** - 快速验证NPU调度能力

### 环境要求

| 要求 | 说明                |
|------|-------------------|
| 计算节点 | 以Altlas 800T A2 arm64训练服务器为例    |
| 驱动版本 | 配套服务器的Ascend驱动已安装 |


### 前置检查

在开始前，请确保NPU驱动已正确安装：

```shell
# 检查NPU状态，预期输出示例（显示芯片信息）
npu-smi info
```

### 为NPU节点添加标签

```shell
# 获取节点名称
kubectl get nodes

# 为NPU节点添加必要标签（ 将worker01替换为实际节点名）
kubectl label nodes worker01 workerselector=dls-worker-node
kubectl label nodes worker01 accelerator=huawei-Ascend910
```

### 部署Ascend Device Plugin

#### 1. 拉取Device Plugin镜像

```shell
# 从华为云镜像仓拉取Device Plugin镜像
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0

# 为镜像添加本地标签
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0 ascend-k8sdeviceplugin:v26.0.0
```

#### 2 部署Device Plugin

```shell
# 拉取配置文件
mkdir /tmp/devicePlugin
cd /tmp/devicePlugin
wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip
unzip Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip

# 部署Device Plugin
kubectl apply -f device-plugin-910-v26.0.0.yaml
```

#### 3 验证部署

```shell
# 查看Device Plugin Pod状态
kubectl get pod -n kube-system

# 预期输出
NAME                                  READY   STATUS    RESTARTS   AGE
...
ascend-device-plugin-daemonset-d5ctz  1/1     Running   0          11s
...
```

### 验证NPU资源上报

```shell
# 查看节点的NPU资源
kubectl describe node worker01 | grep -A 10 "huawei.com/Ascend910"

# 预期输出（显示可用的NPU数量）
huawei.com/Ascend910:     8
huawei.com/Ascend910:     8
```

### 调度NPU Pod

#### 1 创建测试Pod配置文件

创建 `npu-test-pod.yaml`：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: npu-test
spec:
  nodeSelector:
    workerselector: dls-worker-node
  containers:
  - name: npu-container
    image: ubuntu:22.04
    command: ["/bin/bash", "-c", "sleep 3600"]
    resources:
      limits:
        huawei.com/Ascend910: 1  # 请求1个NPU卡
      requests:
        huawei.com/Ascend910: 1
    volumeMounts:
    - name: ascend-driver
      mountPath: /usr/local/Ascend/driver
      readOnly: true
  volumes:
  - name: ascend-driver
    hostPath:
      path: /usr/local/Ascend/driver
```

#### 2 部署测试Pod

```shell
kubectl apply -f npu-test-pod.yaml
```

#### 3 验证Pod调度

```shell
# 查看Pod状态
kubectl get pods npu-test -o wide

# 预期输出（STATUS为Running表示调度成功）
NAME      READY   STATUS    RESTARTS   AGE   IP           NODE      NOMINATED NODE
npu-test  1/1     Running   0          10s   10.244.1.2   worker01  <none>
```

### 验证NPU访问

```shell
# 进入容器验证NPU可用性
kubectl exec -it npu-test  -- /bin/bash

# 在容器内执行npu-smi info正确回显卡信息
export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:${LD_LIBRARY_PATH}
npu-smi info
```

### 清理测试资源

```shell
# 删除测试Pod
kubectl delete pod npu-test

# 删除Device Plugin（如需）
kubectl delete -f device-plugin-910-v26.0.0.yaml
```

### 常见问题

| 问题 | 原因 | 解决方法 |
|------|------|---------|
| Pod一直Pending | NPU资源不足或节点标签不匹配 | 检查`kubectl describe pod`和节点标签 |
| Device Plugin启动失败 | 驱动路径不正确 | 检查`/usr/local/Ascend/driver`是否存在 |


## 训练业务快速入门
本章节以待安装设备为两台Atlas 800T A2 训练服务器（一台作为管理节点、一台作为计算节点）为例，指导开发者快速完成NodeD、Ascend Device Plugin、Ascend Docker Runtime、Volcano、ClusterD、Ascend Operator组件的安装及使用整卡调度特性快速下发训练任务。

### 操作说明<a name="section17940333114314"></a>

**表 1**  关键步骤说明

|操作步骤|操作说明|更多参考|
|--|--|--|
|[安装组件](#section1837511531098)|以Atlas 800T A2 训练服务器为例，手把手带您在昇腾设备上快速安装集群调度组件。|更多安装集群调度组件的参数说明和操作步骤，请参考[安装部署](installation_guide/02_installation/manual_installation/00_obtaining_software_packages.md)章节。|
|[下发训练任务](#section106493419399)|以一个简单的PyTorch训练任务为例，让您快速了解训练任务下发的操作流程。|更多下发训练任务的参数说明和操作步骤，请参考[基础调度](./usage/basic_scheduling/00_feature_description.md)章节。|


### 安装组件<a name="section1837511531098"></a>

以下步骤命令均以待安装设备Atlas 800T A2 训练服务器为例，如需了解所有组件的详细安装步骤和参数说明请参见[安装](installation_guide/02_installation/manual_installation/00_obtaining_software_packages.md)。

1. 以root用户登录计算或管理节点，创建组件安装目录。
    1. 依次执行以下命令，在**计算节点**创建安装目录。以下目录仅为示例，请以实际为准。

        ```shell
        mkdir /tmp/noded
        mkdir /tmp/devicePlugin
        mkdir /tmp/Ascend-docker-runtime
        ```

    2. 依次执行以下命令，在**管理节点**创建安装目录。以下目录仅为示例，请以实际为准。

        ```shell
        mkdir /tmp/ascend-volcano
        mkdir /tmp/ascend-operator
        mkdir /tmp/clusterd
        ```

2. 下载软件包。以AArch64架构为例，用户需根据实际情况下载对应架构的软件包。
    1. 依次执行以下命令，在**计算节点**获取NodeD、Ascend Device Plugin和Ascend Docker Runtime组件安装包并解压。

        ```shell
        cd /tmp/noded
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip

        cd /tmp/devicePlugin
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip

        cd /tmp/Ascend-docker-runtime
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-docker-runtime_26.0.0_linux-aarch64.run
        ```

    2. 在**管理节点**依次执行以下命令，获取Volcano、ClusterD和Ascend Operator组件安装包。

        ```shell
        cd /tmp/ascend-volcano
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip

        cd /tmp/ascend-operator
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64.zip

        cd /tmp/clusterd
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-clusterd_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-clusterd_26.0.0_linux-aarch64.zip
        ```

3. 拉取组件镜像。
   1. 依次执行以下命令，在**计算节点**拉取组件镜像。

        ```shell
        cd /tmp/noded
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0 noded:v26.0.0

        cd /tmp/devicePlugin
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0 ascend-k8sdeviceplugin:v26.0.0
        ```

    2. 依次执行以下命令，在**管理节点**制作组件镜像。

        ```shell
        cd /tmp/ascend-volcano/volcano-v1.7.0
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0 volcanosh/vc-scheduler:v1.7.0-v26.0.0

        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0 volcanosh/vc-controller-manager:v1.7.0-v26.0.0

        cd /tmp/ascend-operator
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0 ascend-operator:v26.0.0

        cd /tmp/clusterd
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0 clusterd:v26.0.0
        ```

4. 创建节点标签。
    >[!NOTE]
    >
    >若执行创建节点标签命令时提示"already has a value ... and --overwrite is false"，说明该标签已存在，可使用--overwrite覆盖。
    1. 在K8s管理节点执行以下命令，查询节点名称。

        ```shell
        kubectl get node
        ```

        回显示例如下：

        ```ColdFusion
        NAME       STATUS   ROLES           AGE   VERSION
        worker01   Ready    worker    23h   v1.17.3
        ```

    2. 依次执行以下命令，为**计算节点**创建节点标签（如节点名称为“worker01”）。

        ```shell
        kubectl label nodes worker01 node-role.kubernetes.io/worker=worker
        kubectl label nodes worker01 workerselector=dls-worker-node
        kubectl label nodes worker01 host-arch=huawei-arm
        kubectl label nodes worker01 accelerator=huawei-Ascend910
        kubectl label nodes worker01 accelerator-type=module-{xxx}b-8     #填写芯片型号数值
        kubectl label nodes worker01 nodeDEnable=on
        ```

    3. 执行以下命令，为**管理节点**创建节点标签（如节点名称为“master01”）。

        ```shell
        kubectl label nodes master01 masterselector=dls-master-node
        ```

5. 创建用户。
    >[!NOTE]
    >
    >创建用户时需要sudo权限。
    1. 依次执行以下命令，在**计算节点**创建用户名。

        ```shell
        useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
        usermod -a -G HwHiAiUser hwMindX
        ```

    2. 执行以下命令，在**管理节点**创建用户名。

        ```shell
        useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
        ```

6. 创建日志目录。不支持用户自定义日志目录。
    >[!NOTE]
    >
    >创建日志目录时需要sudo权限。
    1. 依次执行以下命令，在**计算节点**创建日志目录。

        ```shell
        mkdir -m 755 /var/log/mindx-dl
        chown root:root /var/log/mindx-dl
        mkdir -m 750 /var/log/mindx-dl/devicePlugin
        chown root:root /var/log/mindx-dl/devicePlugin
        mkdir -m 750 /var/log/mindx-dl/noded
        chown hwMindX:hwMindX /var/log/mindx-dl/noded
        ```

    2. 依次执行以下命令，在**管理节点**创建日志目录。

        ```shell
        mkdir -m 755 /var/log/mindx-dl
        chown root:root /var/log/mindx-dl
        mkdir -m 750 /var/log/mindx-dl/volcano-controller
        chown hwMindX:hwMindX /var/log/mindx-dl/volcano-controller
        mkdir -m 750 /var/log/mindx-dl/volcano-scheduler
        chown hwMindX:hwMindX /var/log/mindx-dl/volcano-scheduler
        mkdir -m 750 /var/log/mindx-dl/ascend-operator
        chown hwMindX:hwMindX /var/log/mindx-dl/ascend-operator
        mkdir -m 750 /var/log/mindx-dl/clusterd
        chown hwMindX:hwMindX /var/log/mindx-dl/clusterd
        ```

7. 在任意节点执行以下命令，创建命名空间。

    ```shell
    kubectl create ns mindx-dl
    ```

8. 安装组件。
    1. 依次执行以下命令，在计算节点的宿主机上安装Ascend Docker Runtime。

        ```shell
        cd /tmp/Ascend-docker-runtime
        chmod u+x Ascend-docker-runtime_26.0.0_linux-aarch64.run
        ./Ascend-docker-runtime_26.0.0_linux-aarch64.run --install
        systemctl daemon-reload && systemctl restart docker
        ```
    2. 在**计算节点**，依次执行以下命令，安装组件。
        ```shell
        cd /tmp/noded
        kubectl apply -f noded-v26.0.0.yaml

        cd /tmp/devicePlugin
        kubectl apply -f device-plugin-volcano-v26.0.0.yaml
        ```

    2. 在**管理节点**，依次执行以下命令，安装组件。

        ```shell
        cd /tmp/ascend-operator
        kubectl apply -f ascend-operator-v26.0.0.yaml

        cd /tmp/ascend-volcano/volcano-v1.7.0  # 使用1.9.0版本Volcano需要修改为v1.9.0
        kubectl apply -f volcano-v1.7.0.yaml

        cd /tmp/clusterd
        kubectl apply -f clusterd-v26.0.0.yaml
        ```

    3. 执行以下命令，查看组件是否启动成功。

        ```shell
        kubectl get pod -A
        ```

        以clusterd组件为例，回显示例如下，出现**Running**表示组件启动成功。

        ```ColdFusion
        NAME                              READY   STATUS    RESTARTS   AGE
        ...
        clusterd-fd6t8                       1/1     Running   0          74s
        ...
        ```

### 下发训练任务<a name="section106493419399"></a>

1. 准备镜像。

    从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据系统架构（ARM/x86_64），下载24.0.X版本的ascend-pytorch训练镜像。基于训练基础镜像进行修改，将容器中默认用户修改为root。镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

2. 脚本适配。
    1. <a name="zh-cn_topic_0000001558834814_li1298552813512"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50_ID4149_for_PyTorch”作为训练代码。
    2. 自行准备ResNet-50对应的数据集，使用时请遵守对应规范。
    3. 管理员用户上传数据集到存储节点。进入“/data/atlas_dls/public”目录，将数据集上传到任意位置，如“/data/atlas_dls/public/dataset/resnet50/imagenet”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

    4. 将[步骤1](#zh-cn_topic_0000001558834814_li1298552813512)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch”目录上传至环境，如“/data/atlas_dls/public/code/”路径下。
    5. 在“/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch”路径下，注释掉main.py中的以下代码。

        ```Python
        def main():
            args = parser.parse_args()
            os.environ['MASTER_ADDR'] = args.addr
            #os.environ['MASTER_PORT'] = '29501'  # 注释该行代码
            if os.getenv('ALLOW_FP32', False) and os.getenv('ALLOW_HF32', False):
                raise RuntimeError('ALLOW_FP32 and ALLOW_HF32 cannot be set at the same time!')
            elif os.getenv('ALLOW_HF32', False):
                torch.npu.conv.allow_hf32 = True
            elif os.getenv('ALLOW_FP32', False):
                torch.npu.conv.allow_hf32 = False
                torch.npu.matmul.allow_hf32 = False
        ```

    6. 进入[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)仓库，根据[mindcluster-deploy开源仓版本说明](./appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/basic-training/without-ranktable/pytorch”目录中的train_start.sh，在“/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts”路径下，构造如下的目录结构。

        ```text
        root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts#
        scripts/
             ├── train_start.sh
        ```

3. 准备任务YAML。
    1. 进入[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)仓库，根据[mindcluster-deploy开源仓版本说明](./appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/basic-training/without-ranktable/pytorch”目录下的“pytorch_standalone_acjob_<i>\{xxx\}</i>.yaml”文件（<i>\{xxx\}</i>表示芯片型号的数值）。示例默认为单机单卡任务。
    2. 修改示例YAML，修改完成后将其上传至任意文件路径。下述YAML中各参数的详细说明详见[表1](./api/ascend_operator.md)。

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        ...
        spec:
        ...
          replicaSpecs:
            Master:
        ...
                spec:
                  nodeSelector:
                    host-arch: huawei-arm
                    accelerator-type: module-{xxx}b-8   # 由原来的card-{xxx}b-2修改为module-{xxx}b-8，{xxx}表示芯片型号的数值
                  containers:
                  - name: ascend
                    image: pytorch-test:latest     # 修改为步骤1中获取的镜像名称
        ...
                    resources:
                      limits:
                        huawei.com/Ascend910: 1
                      requests:
                        huawei.com/Ascend910: 1
        ...
                  volumes:
                  - name: code
                    nfs:      #如没有安装nfs服务，需要将nfs改为hostPath，并且删掉server: 127.0.0.1
                      server: 127.0.0.1
                      path: "/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/"
                  - name: data
                    nfs:     #如没有安装nfs服务，需要将nfs改为hostPath，并且删掉server: 127.0.0.1
                      server: 127.0.0.1
                      path: "/data/atlas_dls/public/dataset/"
                  - name: output
                    nfs:     #如没有安装nfs服务，需要将nfs改为hostPath，并且删掉server: 127.0.0.1
                      server: 127.0.0.1
                      path: "/data/atlas_dls/output/"
        ...
        ```

4. 执行以下命令，下发单机单卡任务。

    ```shell
    kubectl apply -f pytorch_standalone_acjob_{xxx}.yaml
    ```

5. 执行以下命令，查看Pod运行情况。

    ```shell
    kubectl get pod --all-namespaces -o wide
    ```

    回显示例如下，出现Running表示任务正常运行。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE     IP                NODE      NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          6s      192.168.244.xxx   worker01   <none>           <none>
    ```

    >[!NOTE]
    >
    >若下发训练任务后，任务一直处于Pending状态，可以参见[训练任务处于Pending状态，原因：nodes are unavailable](https://gitcode.com/Ascend/mind-cluster/issues/352)或者[资源不足时，任务处于Pending状态](https://gitcode.com/Ascend/mind-cluster/issues/355)章节进行处理。

6. 查看训练结果。
    1. 在任意节点执行如下命令，查看训练结果。

        ```shell
        kubectl logs -n  命名空间名称 pod名称
        ```

        如：

        ```shell
        kubectl logs -n default default-test-pytorch-master-0
        ```

    2. 查看训练日志，如果出现如下内容表示训练成功。

        ```ColdFusion
        [20251218-20:31:57] [MindXDL Service Log]server id is: 0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=7 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=6 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=5 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=4 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=3 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=2 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=1 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/lib/python3.10/site-packages/torchvision/io/image.py:13: UserWarning: Failed to load image Python extension: ''If you don't plan on using image functionality from `torchvision.io`, you can ignore this warning. Otherwise, there might be something wrong with your environment. Did you have `libjpeg` or `libpng` installed before building `torchvision` from source?
          warn(
        [2025-12-18 20:32:02] [WARNING] [470] profiler.py: Invalid parameter export_type: None, reset it to text.
        /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py:201: UserWarning: You have chosen to seed training. This will turn on the CUDNN deterministic setting, which can slow down your training considerably! You may see unexpected behavior when restarting from checkpoints.
        warnings.warn('You have chosen to seed training. '
        /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py:208: UserWarning: You have chosen a specific GPU. This will completely disable data parallelism.
        warnings.warn('You have chosen a specific GPU. This will completely '
        Use GPU: 0 for training
        => creating model 'resnet50'
        ```
