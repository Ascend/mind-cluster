# 整卡调度或静态vNPU调度（训练）<a name="ZH-CN_TOPIC_0000002479387138"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002511347093"></a>

**前提条件<a name="section52051339787"></a>**

- 确保环境中有配置相应的存储方案，比如使用NFS（Network File System），用户可以参见[安装NFS](../../common_operations.md#安装nfs)进行操作。
- 在使用整卡调度或静态vNPU调度特性前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../../installation_guide//03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
    - 调度器（Volcano或其他调度器）
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - Ascend Operator
    - ClusterD
    - NodeD

- 对于训练任务类型为acjob，调度器为Volcano的整卡调度，支持批量创建Pod和批量调度功能。
    - 若要使用批量创建Pod功能，安装Ascend Operator组件时需使用openFuyao定制Kubernetes组件。
    - 若要使用批量调度功能，安装Volcano组件时需使用openFuyao定制Kubernetes和volcano-ext组件，并开启批量调度功能。
    - 批量调度功能适用于超大规模集群场景，在此场景下请根据实际需要扩展MindCluster组件分配的CPU和内存资源，防止MindCluster组件出现性能不足或者超出分配内存使用，导致组件被Kubernetes驱逐。

**使用方式<a name="section179431435174811"></a>**

- 通过命令行使用：整卡调度或静态vNPU调度特性需要使用调度器，用户可以选择使用Volcano调度器和其他调度器。无论选择哪种调度器，都需要使用Ascend Operator组件设置资源信息。
- 集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section577625973520"></a>**

- 资源监测可以和训练场景下的所有特性一起使用。
- 集群中同时跑多个训练任务，每个任务使用的特性可以不同。
- 静态vNPU调度特性需要搭配算力虚拟化特性一起使用，关于静态虚拟化的相关说明和操作请参见[静态虚拟化](../virtual_instance/virtual_instance_with_hdk/06_mounting_vnpu.md#静态虚拟化)章节。

**支持的产品形态<a name="section169961844182917"></a>**

- 支持以下产品使用**整卡调度**。
    - Atlas 训练系列产品
    - <term>Atlas A2 训练系列产品</term>
    - <term>Atlas A3 训练系列产品</term>

- 支持以下产品使用**静态vNPU调度**。

    - Atlas 训练系列产品

**使用流程<a name="section5640184231810"></a>**

整卡调度、静态vNPU调度有3种使用场景，分别是通过命令行使用（Volcano）、通过命令行使用（其他调度器）和集成后使用。

通过命令行使用Volcano和其他调度器的使用流程一致。使用其他调度器准备任务YAML需要参考[通过命令行使用（其他调度器）](#通过命令行使用其他调度器)章节创建任务YAML。使用其他调度器的其余操作和使用Volcano一致，可以参考[通过命令行使用（Volcano）](#通过命令行使用volcano)进行操作。

**图 1**  整卡调度和静态vNPU调度使用流程<a name="fig107864120214"></a>  
![](../../../figures/scheduling/整卡调度和静态vNPU调度使用流程.png "整卡调度和静态vNPU调度使用流程")

1. 脚本适配时，用户可根据实际情况选择通过环境变量或文件配置资源信息。
2. 在准备任务YAML时，下发的任务YAML需要根据具体的NPU型号，选择不同的YAML进行修改适配。选择YAML时可以参考[准备任务YAML](#准备任务yaml)，根据实际情况选择合适的YAML。

## 实现原理<a name="ZH-CN_TOPIC_0000002479387150"></a>

根据训练任务类型的不同，特性的原理图略有差异。静态vNPU调度需要使用npu-smi工具提前创建好需要的vNPU。

**acjob任务<a name="section9971431567"></a>**

acjob任务原理图如[图1](#fig5188536014)所示。

**图 1**  acjob任务调度原理图<a name="fig5188536014"></a>  
![](../../../figures/scheduling/acjob任务调度原理图.png "acjob任务调度原理图")

各步骤说明如下：

1. 集群调度组件定期上报节点和芯片信息。
    - kubelet上报节点芯片数量到节点对象（node）中。
    - Ascend Device Plugin定期上报芯片拓扑信息。
        - 上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）、已使用的芯片数量（allocated）和芯片的基础信息（device ip和super\_device\_ip）上报到Node中，用于整卡调度。
        - 上报vNPU相关信息到Node中，用于静态vNPU调度。

    - 当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2. ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3. 用户通过kubectl或者其他深度学习平台下发acjob任务。
4. Ascend Operator为任务创建相应的PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5. Ascend Operator为任务创建相应的Pod，并在容器中注入集合通信所需环境变量。
6. volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
    - 整卡调度写入整卡信息。
    - 静态vNPU调度写入vNPU相关信息。

7. kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin或volcano-scheduler在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源。
8. Ascend Operator读取Pod的annotation信息，将相关信息写入hccl.json。
9. 容器读取环境变量或者hccl.json信息，建立通信通道，开始执行训练任务。

**vcjob任务<a name="section13884164615313"></a>**

vcjob任务的原理图如[图2](#fig8717151315416)所示。

**图 2**  vcjob任务调度原理图<a name="fig8717151315416"></a>  
![](../../../figures/scheduling/vcjob任务调度原理图.png "vcjob任务调度原理图")

各步骤说明如下：

1. 集群调度组件定期上报节点和芯片信息。
    - kubelet上报节点芯片数量到节点对象（node）中。
    - Ascend Device Plugin定期上报芯片拓扑信息。
        - 上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）和已使用的芯片数量（allocated）上报到Node中，用于整卡调度。
        - 上报vNPU相关信息到Node中，用于静态vNPU调度。

    - 当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2. ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3. 用户通过kubectl或者其他深度学习平台下发vcjob任务。
4. volcano-controller为任务创建相应PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5. 当集群资源满足任务要求时，volcano-controller创建任务Pod。
6. volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
    - 整卡调度写入整卡信息。
    - 静态vNPU调度写入vNPU相关信息。

7. kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源，将hccl.json挂载进入容器。
8. Ascend Operator获取每个Pod的annotation信息，写入hccl.json。
9. 容器读取hccl.json信息，建立通信渠道，开始执行训练任务。

**deploy任务<a name="section32752223579"></a>**

deploy任务原理图如[图3](#fig06571541566)所示。

**图 3**  deploy任务调度原理图<a name="fig06571541566"></a>  
![](../../../figures/scheduling/deploy任务调度原理图.png "deploy任务调度原理图")

各步骤说明如下：

1. 集群调度组件定期上报节点和芯片信息。
    - kubelet上报节点芯片数量到节点对象（node）中。
    - Ascend Device Plugin定期上报芯片拓扑信息。
        - 上报整卡信息。将芯片的物理ID上报到device-info-cm中；可调度的芯片总数量（allocatable）和已使用的芯片数量（allocated）上报到Node中，用于整卡调度。
        - 上报vNPU相关信息到Node中，用于静态vNPU调度。

    - 当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2. ClusterD读取device-info-cm和node-info-cm中信息后，将信息写入cluster-info-cm。
3. 用户通过kubectl或者其他深度学习平台下发deploy任务。
4. kube-controller为任务创建相应Pod。
5. volcano-controller创建任务PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
6. volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入选择的芯片信息。
    - 整卡调度写入整卡信息。
    - 静态vNPU调度写入vNPU相关信息。

7. kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin在Pod的annotation上写入芯片信息。Ascend Docker Runtime协助挂载相应资源，将hccl.json挂载进入容器。
8. Ascend Operator获取每个Pod的annotation信息，写入hccl.json。
9. 容器读取hccl.json信息，建立通信渠道，开始执行训练任务。

## 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002479227158"></a>

### 制作镜像<a name="ZH-CN_TOPIC_0000002479227164"></a>

**获取训练镜像<a name="zh-cn_topic_0000001609314597_section971616541059"></a>**

可选择以下方式中的一种来获取训练镜像：

- （推荐）从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据系统架构（ARM/x86\_64）、模型框架（PyTorch、MindSpore）下载配套驱动版本的**训练基础镜像**。基于训练基础镜像进行修改，将容器中默认用户修改为root（21.0.4版本之后训练基础镜像默认用户为非root）。基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。
- 从头开始定制用户自己的训练镜像，制作过程请参考[制作镜像](../../common_operations.md#制作镜像)中制作容器相关章节。

可将下载/制作的训练基础镜像重命名，如：training:v26.0.0。

**加固镜像<a name="zh-cn_topic_0000001609314597_section8425732111611"></a>**

下载或者制作的训练基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../../security_hardening.md#容器安全加固)章节进行操作。

### 脚本适配<a name="ZH-CN_TOPIC_0000002511347097"></a>

#### 通过环境变量配置资源信息<a name="ZH-CN_TOPIC_0000002479387142"></a>

根据模型框架选择对应的指导示例。

- [PyTorch](#zh-cn_topic_0000001558834814_section17760205783316)
- [MindSpore](#zh-cn_topic_0000001558834814_section868111733711)

    >[!NOTE] 
    >- 本节中使用的数据集为[ImageNet2012](https://image-net.org/challenges/LSVRC/2012/2012-downloads.php)数据集（**注：如使用该数据集需遵循数据集提供者的使用规范**）。
    >- 下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
    >- 以下MindSpore示例需使用CANN 8.5.0之前版本。   

**PyTorch<a name="zh-cn_topic_0000001558834814_section17760205783316"></a>**

1. <a name="zh-cn_topic_0000001558834814_li1298552813512"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50\_ID4149\_for\_PyTorch”作为训练代码。
2. 自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3. 管理员用户上传数据集到存储节点。
    1. 进入“/data/atlas\_dls/public”目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

        回显示例如下：

        ```ColdFusion
        /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

    2. 执行**du -sh**命令，查看数据集大小。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# du -sh
        ```

        回显示例如下：

        ```ColdFusion
        11G
        ```

4. 将[步骤1](#zh-cn_topic_0000001558834814_li1298552813512)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50\_ID4149\_for\_PyTorch”目录上传至环境，如“/data/atlas\_dls/public/code/”路径下。
5. 在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch”路径下，注释或删除main.py文件中的加粗字段。

    <pre codetype="Python">
    def main():
        args = parser.parse_args()
        os.environ['MASTER_ADDR'] = args.addr
        <strong>#os.environ['MASTER_PORT'] = '29501'  # 注释或删除该行代码</strong>
        if os.getenv('ALLOW_FP32', False) and os.getenv('ALLOW_HF32', False):
            raise RuntimeError('ALLOW_FP32 and ALLOW_HF32 cannot be set at the same time!')
        elif os.getenv('ALLOW_HF32', False):
            torch.npu.conv.allow_hf32 = True
        elif os.getenv('ALLOW_FP32', False):
            torch.npu.conv.allow_hf32 = False
            torch.npu.matmul.allow_hf32 = False</pre>

6. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/without-ranktable/pytorch”目录中的train\_start.sh，在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch/scripts”路径下，构造如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts#
    scripts/
         ├── train_start.sh
    ```

**MindSpore<a name="zh-cn_topic_0000001558834814_section868111733711"></a>**

1. <a name="zh-cn_topic_0000001558834814_li1141932513379"></a>下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/cv/ResNet)中master分支的“ResNet”代码作为训练代码。
2. 自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3. 管理员用户上传数据集到存储节点。
    1. 进入“/data/atlas\_dls/public”目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/imagenet”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# pwd
        ```

        回显示例如下：

        ```ColdFusion
        /data/atlas_dls/public/dataset/imagenet
        ```

    2. 执行**du -sh**命令，查看数据集大小。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# du -sh
        ```

        回显示例如下：

        ```ColdFusion
        11G
        ```

4. 在本地解压[步骤1](#zh-cn_topic_0000001558834814_li1141932513379)中下载的训练代码，将“models/official/cv/”下的“ResNet”目录重命名为“ResNet50\_for\_MindSpore\_2.0\_code”。后续步骤以“ResNet50\_for\_MindSpore\_2.0\_code”目录为例。
5. 将ResNet50\_for\_MindSpore\_2.0\_code文件上传至环境“/data/atlas\_dls/public/code/”路径下。
6. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/without-ranktable/mindspore”目录中的“train\_start.sh”文件，结合训练代码中“scripts”目录，在host上构造如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_for_MindSpore_2.0_code/scripts/#
    scripts/
    ├── docker_start.sh
    ├── run_standalone_train_gpu.sh
    ├── run_standalone_train.sh
     ...
    └── train_start.sh
    ```

7. 进入“/data/atlas\_dls/public/code/ResNet50\_for\_MindSpore\_2.0\_code/train.py”目录下，修改train.py对应部分，如下所示。

    ```Python
     ...
         if config.run_distribute:
             if target == "Ascend":
               #device_id = int(os.getenv('DEVICE_ID', '0'))   #注释该行代码
               #ms.set_context(device_id=device_id)     #注释该行代码
                 ms.set_auto_parallel_context(device_num=config.device_num, parallel_mode=ms.ParallelMode.DATA_PARALLEL,
                                              gradients_mean=True)
                 set_algo_parameters(elementwise_op_strategy_follow=True)
                 if config.net_name == "resnet50" or config.net_name == "se-resnet50":
                     if config.boost_mode not in ["O1", "O2"]:
                         ms.set_auto_parallel_context(all_reduce_fusion_config=config.all_reduce_fusion_config)
                 elif config.net_name in ["resnet101", "resnet152"]:
                     ms.set_auto_parallel_context(all_reduce_fusion_config=config.all_reduce_fusion_config)
                 init()
             # GPU target
     ...
    ```

#### 通过文件配置资源信息<a name="ZH-CN_TOPIC_0000002479387136"></a>

通过文件变量配置资源信息支持创建以下3种类型的对象：acjob、vcjob及deploy。下面将以vcjob和deploy为例，介绍脚本适配的操作示例。

- [PyTorch](#zh-cn_topic_0000001558834798_section17760205783316)
- [MindSpore](#zh-cn_topic_0000001558834798_section868111733711)

>[!NOTE] 
>
>- 本节中使用的数据集为[ImageNet2012](https://image-net.org/challenges/LSVRC/2012/2012-downloads.php)数据集（**注：如使用该数据集需遵循数据集提供者的使用规范**）。
>- 下文中模型示例代码可能与实际版本存在差异，请以实际版本代码为准。
>- 以下MindSpore示例需使用CANN 8.5.0之前版本。

**PyTorch<a name="zh-cn_topic_0000001558834798_section17760205783316"></a>**

1. <a name="zh-cn_topic_0000001558834798_li1298552813512"></a>下载[PyTorch代码仓](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch)中master分支的“ResNet50\_ID4149\_for\_PyTorch”作为训练代码。
2. 自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3. 管理员用户上传数据集到存储节点。
    1. 进入“/data/atlas\_dls/public”目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/resnet50/imagenet”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

        回显示例如下：

        ```ColdFusion
        /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

    2. 执行**du -sh**命令，查看数据集大小。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# du -sh
        ```

        回显示例如下：

        ```ColdFusion
        11G
        ```

4. 将[步骤1](#zh-cn_topic_0000001558834798_li1298552813512)中下载的训练代码解压到本地，将解压后的训练代码中“ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50\_ID4149\_for\_PyTorch”目录上传至环境，如“/data/atlas\_dls/public/code/”路径下。
5. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/ranktable”目录中的train\_start.sh、rank\_table.sh和utils.sh文件，在“/data/atlas\_dls/public/code/ResNet50\_ID4149\_for\_PyTorch/scripts”路径下，构造如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts#
    scripts/
         ├── train_start.sh
         ├── utils.sh
         └── rank_table.sh
    ```

**MindSpore<a name="zh-cn_topic_0000001558834798_section868111733711"></a>**

1. <a name="zh-cn_topic_0000001558834798_li1141932513379"></a>下载[MindSpore代码仓](https://gitee.com/mindspore/models/tree/master/official/cv/ResNet)中master分支的“ResNet”代码作为训练代码。
2. 自行准备ResNet50对应的数据集，使用时请遵守对应规范。
3. 管理员用户上传数据集到存储节点。
    1. 进入“/data/atlas\_dls/public”目录，将数据集上传到任意位置，如“/data/atlas\_dls/public/dataset/imagenet”。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# pwd
        ```

        回显示例如下：

        ```ColdFusion
        /data/atlas_dls/public/dataset/imagenet
        ```

    2. 执行**du -sh**命令，查看数据集大小。

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/imagenet# du -sh
        ```

        回显示例如下：

        ```ColdFusion
        11G
        ```

4. 在本地解压[步骤1](#zh-cn_topic_0000001558834798_li1141932513379)中下载的训练代码，将“models/official/cv/”下的“ResNet”目录重命名为“ResNet50\_for\_MindSpore\_2.0\_code”。后续步骤以“ResNet50\_for\_MindSpore\_2.0\_code”目录为例。
5. 将ResNet50\_for\_MindSpore\_2.0\_code文件上传至环境“/data/atlas\_dls/public/code/”路径下。
6. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。获取“samples/train/basic-training/ranktable”目录中的“train\_start.sh”、“utils.sh”和“rank\_table.sh”文件，结合训练代码中“scripts”目录，在host上构造如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/ResNet50_for_MindSpore_2.0_code/scripts/#
    scripts/
    ├── cache_util.sh
    ├── docker_start.sh
    ├── run_standalone_train_gpu.sh
    ├── run_standalone_train.sh
     ...
    ├── rank_table.sh
    ├── utils.sh
    └── train_start.sh
    ```

### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479227170"></a>

#### 选择YAML示例<a name="ZH-CN_TOPIC_0000002479227150"></a>

集群调度为用户提供YAML示例，用户需要根据使用的组件、芯片类型和任务类型等，选择相应的YAML示例并根据需求进行相应修改后才可使用。

**通过环境变量配置资源信息场景<a name="section1969664932615"></a>**

- 若当前环境使用的是<term>Atlas A2 训练系列产品</term>，选择[表1](#table529015783811)获取相应的YAML示例。

    根据[表1](#table529015783811)获取示例YAML后，Atlas 800T A2 训练服务器、Atlas 200T A2 Box16 异构子框和A200T A3 Box8 超节点服务器可基于[表1 YAML参数说明](#table159746356276)给出的参数说明进行修改适配。

- 若当前环境使用的是Atlas 训练系列产品，选择[表2](#table18698184918261)获取相应的YAML示例。

    根据[表2](#table18698184918261)获取示例YAML后，服务器（插Atlas 300T 训练卡）可基于Atlas 800 训练服务器的YAML，以及参考[表1 YAML参数说明](#table159746356276)给出的参数说明进行修改适配。

- 若当前环境使用的是<term>Atlas A3 训练系列产品</term>，选择[表3](#table57051049102614)获取相应的YAML示例。

- 若当前环境使用的是Atlas 950 训练系列产品，选择[表4](#table5290157950yaml)获取相应的YAML示例。

**表 1** <term>Atlas A2 训练系列产品</term>支持的YAML

<a name="table529015783811"></a>
<table><thead align="left"><tr id="row52903576386"><th class="cellrowborder" valign="top" width="8.8%" id="mcps1.2.7.1.1"><p id="p129019578385"><a name="p129019578385"></a><a name="p129019578385"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.000000000000002%" id="mcps1.2.7.1.2"><p id="p14290115712387"><a name="p14290115712387"></a><a name="p14290115712387"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.650000000000002%" id="mcps1.2.7.1.3"><p id="p1329015723817"><a name="p1329015723817"></a><a name="p1329015723817"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="30.740000000000006%" id="mcps1.2.7.1.4"><p id="p14291125717389"><a name="p14291125717389"></a><a name="p14291125717389"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="18.810000000000002%" id="mcps1.2.7.1.5"><p id="p1129114571381"><a name="p1129114571381"></a><a name="p1129114571381"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.000000000000002%" id="mcps1.2.7.1.6"><p id="p1229110574387"><a name="p1229110574387"></a><a name="p1229110574387"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row13291757163813"><td class="cellrowborder" rowspan="7" valign="top" width="8.8%" headers="mcps1.2.7.1.1 "><p id="p11291115783810"><a name="p11291115783810"></a><a name="p11291115783810"></a>Ascend Job</p>
<p id="p1629145703816"><a name="p1629145703816"></a><a name="p1629145703816"></a></p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="15.000000000000002%" headers="mcps1.2.7.1.2 "><p id="p14227163913366"><a name="p14227163913366"></a><a name="p14227163913366"></a><span id="ph13291155773812"><a name="ph13291155773812"></a><a name="ph13291155773812"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.650000000000002%" headers="mcps1.2.7.1.3 "><p id="p729135713817"><a name="p729135713817"></a><a name="p729135713817"></a><span id="ph1029120577382"><a name="ph1029120577382"></a><a name="ph1029120577382"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.740000000000006%" headers="mcps1.2.7.1.4 "><p id="p52921057113812"><a name="p52921057113812"></a><a name="p52921057113812"></a>tensorflow_multinodes_acjob_<span id="ph0292957193810"><a name="ph0292957193810"></a><a name="ph0292957193810"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.7.1.5 "><p id="p1729235703814"><a name="p1729235703814"></a><a name="p1729235703814"></a>示例默认为双机2卡任务。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="15.000000000000002%" headers="mcps1.2.7.1.6 "><p id="p17292357133814"><a name="p17292357133814"></a><a name="p17292357133814"></a>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
<div class="note" id="note14933145219586"><a name="note14933145219586"></a><a name="note14933145219586"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1027616512420"><a name="p1027616512420"></a><a name="p1027616512420"></a><span id="ph9014016509"><a name="ph9014016509"></a><a name="ph9014016509"></a>下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row829235719380"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p52921579382"><a name="p52921579382"></a><a name="p52921579382"></a><span id="ph1829255713389"><a name="ph1829255713389"></a><a name="ph1829255713389"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1929295713389"><a name="p1929295713389"></a><a name="p1929295713389"></a>pytorch_multinodes_acjob_<span id="ph7292757133817"><a name="ph7292757133817"></a><a name="ph7292757133817"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1129285753810"><a name="p1129285753810"></a><a name="p1129285753810"></a>示例默认为双机2卡任务。</p>
</td>
</tr>
<tr id="row0292357163814"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1529295723813"><a name="p1529295723813"></a><a name="p1529295723813"></a><span id="ph15292125733819"><a name="ph15292125733819"></a><a name="ph15292125733819"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p182921757103818"><a name="p182921757103818"></a><a name="p182921757103818"></a>mindspore_multinodes_acjob_<span id="ph15292205723815"><a name="ph15292205723815"></a><a name="ph15292205723815"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p52921657103816"><a name="p52921657103816"></a><a name="p52921657103816"></a>示例默认为双机16卡任务。</p>
</td>
</tr>
<tr id="row829335719385"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1293195793816"><a name="p1293195793816"></a><a name="p1293195793816"></a><span id="ph14293145733816"><a name="ph14293145733816"></a><a name="ph14293145733816"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p929305793820"><a name="p929305793820"></a><a name="p929305793820"></a>tensorflow_standalone_acjob_<span id="ph529314572388"><a name="ph529314572388"></a><a name="ph529314572388"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.3 "><p id="p2293205715380"><a name="p2293205715380"></a><a name="p2293205715380"></a>示例默认为单机单卡任务。</p>
<p id="p1924194410282"><a name="p1924194410282"></a><a name="p1924194410282"></a></p>
<p id="p13241444192817"><a name="p13241444192817"></a><a name="p13241444192817"></a></p>
</td>
</tr>
<tr id="row751217295286"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2012563417286"><a name="p2012563417286"></a><a name="p2012563417286"></a><span id="ph1512583402811"><a name="ph1512583402811"></a><a name="ph1512583402811"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p61257345281"><a name="p61257345281"></a><a name="p61257345281"></a>mindspore_standalone_acjob_<span id="ph51251834122810"><a name="ph51251834122810"></a><a name="ph51251834122810"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_4"><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a>{xxx}</em></span>b.yaml</p>
</td>
</tr>
<tr id="row429313576389"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.7.1.1 "><p id="p7293457173815"><a name="p7293457173815"></a><a name="p7293457173815"></a><span id="ph2293125773811"><a name="ph2293125773811"></a><a name="ph2293125773811"></a>PyTorch</span></p>
<p id="p1469421012715"><a name="p1469421012715"></a><a name="p1469421012715"></a></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p8293457173812"><a name="p8293457173812"></a><a name="p8293457173812"></a>pytorch_standalone_acjob_<span id="ph10293195714383"><a name="ph10293195714383"></a><a name="ph10293195714383"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b.yaml</p>
</td>
</tr>
<tr id="row7693111092718"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p10694181017275"><a name="p10694181017275"></a><a name="p10694181017275"></a>pytorch_multinodes_acjob_<span id="ph11166172612911"><a name="ph11166172612911"></a><a name="ph11166172612911"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_6"><a name="zh-cn_topic_0000001519959665_i1489729141619_6"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_6"></a>{xxx}</em></span>b_with_ranktable.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1524254417282"><a name="p1524254417282"></a><a name="p1524254417282"></a>示例默认为单机2卡任务。使用<span id="ph747115394291"><a name="ph747115394291"></a><a name="ph747115394291"></a>Ascend Operator</span>组件生成RankTable文件。</p>
</td>
</tr>
</tbody>
</table>

**表 2** Atlas 训练系列产品支持的YAML

<a name="table18698184918261"></a>
<table><thead align="left"><tr id="row6698849162611"><th class="cellrowborder" valign="top" width="10.000000000000002%" id="mcps1.2.7.1.1"><p id="p15698549192614"><a name="p15698549192614"></a><a name="p15698549192614"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.000000000000002%" id="mcps1.2.7.1.2"><p id="p11698849132612"><a name="p11698849132612"></a><a name="p11698849132612"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="12.000000000000002%" id="mcps1.2.7.1.3"><p id="p2698124919262"><a name="p2698124919262"></a><a name="p2698124919262"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="30.000000000000004%" id="mcps1.2.7.1.4"><p id="p069914491269"><a name="p069914491269"></a><a name="p069914491269"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.000000000000004%" id="mcps1.2.7.1.5"><p id="p66993497268"><a name="p66993497268"></a><a name="p66993497268"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="13.000000000000004%" id="mcps1.2.7.1.6"><p id="p3699649192614"><a name="p3699649192614"></a><a name="p3699649192614"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row206991499261"><td class="cellrowborder" rowspan="6" valign="top" width="10.000000000000002%" headers="mcps1.2.7.1.1 "><p id="p10699249132614"><a name="p10699249132614"></a><a name="p10699249132614"></a>Ascend Job</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="15.000000000000002%" headers="mcps1.2.7.1.2 "><p id="p196992049112613"><a name="p196992049112613"></a><a name="p196992049112613"></a><span id="ph76991749122617"><a name="ph76991749122617"></a><a name="ph76991749122617"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="12.000000000000002%" headers="mcps1.2.7.1.3 "><p id="p146991494269"><a name="p146991494269"></a><a name="p146991494269"></a><span id="ph1669934972611"><a name="ph1669934972611"></a><a name="ph1669934972611"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.000000000000004%" headers="mcps1.2.7.1.4 "><p id="p6699104910262"><a name="p6699104910262"></a><a name="p6699104910262"></a>tensorflow_multinodes_acjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="20.000000000000004%" headers="mcps1.2.7.1.5 "><p id="p369954932614"><a name="p369954932614"></a><a name="p369954932614"></a>示例默认为双机8卡任务。</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="13.000000000000004%" headers="mcps1.2.7.1.6 "><p id="p369974917262"><a name="p369974917262"></a><a name="p369974917262"></a>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row669964932617"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p146994498265"><a name="p146994498265"></a><a name="p146994498265"></a><span id="ph3699154922611"><a name="ph3699154922611"></a><a name="ph3699154922611"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p469944911268"><a name="p469944911268"></a><a name="p469944911268"></a>pytorch_multinodes_acjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p269913494266"><a name="p269913494266"></a><a name="p269913494266"></a>示例默认为双机16卡任务。</p>
</td>
</tr>
<tr id="row670044914266"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p2070054918265"><a name="p2070054918265"></a><a name="p2070054918265"></a><span id="ph177001649182618"><a name="ph177001649182618"></a><a name="ph177001649182618"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p370024915266"><a name="p370024915266"></a><a name="p370024915266"></a>mindspore_multinodes_acjob.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p6700104952618"><a name="p6700104952618"></a><a name="p6700104952618"></a>示例默认为双机8卡任务。</p>
<div class="note" id="note170014493266"><a name="note170014493266"></a><a name="note170014493266"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1370015494264"><a name="p1370015494264"></a><a name="p1370015494264"></a>若下发单机8卡的<span id="ph6700049182610"><a name="ph6700049182610"></a><a name="ph6700049182610"></a>MindSpore</span>任务，需要将mindspore_multinodes_acjob.yaml中minAvailable修改为2，Worker的replicas修改为1。</p>
</div></div>
</td>
</tr>
<tr id="row11700124942615"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1470004911264"><a name="p1470004911264"></a><a name="p1470004911264"></a><span id="ph16700154942618"><a name="ph16700154942618"></a><a name="ph16700154942618"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p15700174913266"><a name="p15700174913266"></a><a name="p15700174913266"></a>tensorflow_standalone_acjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.3 "><p id="p107007497265"><a name="p107007497265"></a><a name="p107007497265"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
<tr id="row15700184992617"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p8700114917265"><a name="p8700114917265"></a><a name="p8700114917265"></a><span id="ph1970044942611"><a name="ph1970044942611"></a><a name="ph1970044942611"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p117007498269"><a name="p117007498269"></a><a name="p117007498269"></a>pytorch_standalone_acjob.yaml</p>
</td>
</tr>
<tr id="row1170074952614"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p1970114911261"><a name="p1970114911261"></a><a name="p1970114911261"></a><span id="ph770124962613"><a name="ph770124962613"></a><a name="ph770124962613"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p770164982610"><a name="p770164982610"></a><a name="p770164982610"></a>mindspore_standalone_acjob.yaml</p>
</td>
</tr>
</tbody>
</table>

**表 3** <term>Atlas A3 训练系列产品</term>支持的YAML

<a name="table57051049102614"></a>
<table><thead align="left"><tr id="row107051249172610"><th class="cellrowborder" valign="top" width="8.799999999999999%" id="mcps1.2.7.1.1"><p id="p8705114972617"><a name="p8705114972617"></a><a name="p8705114972617"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.2"><p id="p1870594972615"><a name="p1870594972615"></a><a name="p1870594972615"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.65%" id="mcps1.2.7.1.3"><p id="p97063498264"><a name="p97063498264"></a><a name="p97063498264"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="39.269999999999996%" id="mcps1.2.7.1.4"><p id="p1706204911262"><a name="p1706204911262"></a><a name="p1706204911262"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="10.280000000000001%" id="mcps1.2.7.1.5"><p id="p970615497266"><a name="p970615497266"></a><a name="p970615497266"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.6"><p id="p170694910264"><a name="p170694910264"></a><a name="p170694910264"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row570610499268"><td class="cellrowborder" rowspan="3" valign="top" width="8.799999999999999%" headers="mcps1.2.7.1.1 "><p id="p1770624902616"><a name="p1770624902616"></a><a name="p1770624902616"></a>Ascend Job</p>
<p id="p167068495269"><a name="p167068495269"></a><a name="p167068495269"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15%" headers="mcps1.2.7.1.2 "><p id="p19706849182618"><a name="p19706849182618"></a><a name="p19706849182618"></a><span id="ph167064499269"><a name="ph167064499269"></a><a name="ph167064499269"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.65%" headers="mcps1.2.7.1.3 "><p id="p6706349192610"><a name="p6706349192610"></a><a name="p6706349192610"></a><span id="ph13706154962618"><a name="ph13706154962618"></a><a name="ph13706154962618"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="39.269999999999996%" headers="mcps1.2.7.1.4 "><p id="p1370624952612"><a name="p1370624952612"></a><a name="p1370624952612"></a>tensorflow_standalone_acjob_super_pod.yaml</p>
</td>
<td class="cellrowborder" valign="top" width="10.280000000000001%" headers="mcps1.2.7.1.5 "><p id="p3707749162616"><a name="p3707749162616"></a><a name="p3707749162616"></a>示例默认为单机单卡任务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15%" headers="mcps1.2.7.1.6 "><p id="p1670794911264"><a name="p1670794911264"></a><a name="p1670794911264"></a>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
<p id="p1770716492268"><a name="p1770716492268"></a><a name="p1770716492268"></a></p>
</td>
</tr>
<tr id="row770724972619"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p0707749172618"><a name="p0707749172618"></a><a name="p0707749172618"></a><span id="ph12707184972613"><a name="ph12707184972613"></a><a name="ph12707184972613"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1370794914264"><a name="p1370794914264"></a><a name="p1370794914264"></a>pytorch_standalone_acjob_super_pod.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p167072049132613"><a name="p167072049132613"></a><a name="p167072049132613"></a>示例默认为单机16卡任务。</p>
</td>
</tr>
<tr id="row7707164912262"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p270754962617"><a name="p270754962617"></a><a name="p270754962617"></a><span id="ph1570754952617"><a name="ph1570754952617"></a><a name="ph1570754952617"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p127081949202617"><a name="p127081949202617"></a><a name="p127081949202617"></a>mindspore_standalone_acjob_super_pod.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p1070817499261"><a name="p1070817499261"></a><a name="p1070817499261"></a>示例默认为双机16卡任务。</p>
</td>
</tr>
</tbody>
</table>

**表 4** Atlas 950 系列产品支持的YAML
<a name="table5290157950yaml"></a>
<table>
    <thead align="left">
        <tr>
            <th class="cellrowborder" valign="top" width="8.799999999999999%" id="mcps1.2.7.1.1"><p id="p8705114972617"><a name="p8705114972617"></a><a name="p8705114972617"></a>任务类型</p></th>
            <th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.2"><p id="p1870594972615"><a name="p1870594972615"></a><a name="p1870594972615"></a>硬件型号</p></th>
            <th class="cellrowborder" valign="top" width="11.65%" id="mcps1.2.7.1.3"><p id="p97063498264"><a name="p97063498264"></a><a name="p97063498264"></a>训练框架</p></th>
            <th class="cellrowborder" valign="top" width="39.269999999999996%" id="mcps1.2.7.1.4"><p id="p1706204911262"><a name="p1706204911262"></a><a name="p1706204911262"></a>YAML文件名称</p></th>
            <th class="cellrowborder" valign="top" width="10.280000000000001%" id="mcps1.2.7.1.5"><p id="p970615497266"><a name="p970615497266"></a><a name="p970615497266"></a>说明</p></th>
            <th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.6"><p id="p170694910264"><a name="p170694910264"></a><a name="p170694910264"></a>获取链接</p></th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td class="cellrowborder" rowspan="2" valign="top" width="8.799999999999999%" headers="mcps1.2.7.1.1 "><p>Ascend Job</p></td>
            <td class="cellrowborder" rowspan="2" valign="top" width="15%" headers="mcps1.2.7.1.2 "><p><span>Atlas 950 SuperPoD</span></p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>PyTorch</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>pytorch_standalone_acjob_950.yaml</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>示例默认为单机8卡任务。</p></td>
            <td class="cellrowborder" rowspan="2" valign="top" width="15%" headers="mcps1.2.7.1.6 ">
                <p>选择相应的训练框架后，<a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/without-ranktable" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
            </td>
        </tr>
        <tr><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>MindSpore</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>mindspore_standalone_acjob_950.yaml</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p>示例默认为单机单卡任务。</p></td>
        </tr>
    </tbody>
</table>

**通过文件配置资源信息场景<a name="section158807920347"></a>**

- 若当前环境使用的是<term>Atlas A2 训练系列产品</term>，选择[表5](#table62591594016)获取相应的YAML示例。

    根据[表5](#table62591594016)获取示例YAML后，Atlas 800T A2 训练服务器、Atlas 200T A2 Box16 异构子框和A200T A3 Box8 超节点服务器可基于[表2 YAML参数说明](#zh-cn_topic_0000001609074269_table1565872494511)给出的参数说明进行修改适配。

- 若当前环境使用的是Atlas 训练系列产品，选择[表6](#table21811158146)获取相应的YAML示例。
- 若当前环境使用的是Atlas 950 训练系列产品，选择[表7](#table950yaml)获取相应的YAML示例。

**表 5** <term>Atlas A2 训练系列产品</term>支持的YAML

<a name="table62591594016"></a>
<table><thead align="left"><tr id="row72551515403"><th class="cellrowborder" valign="top" width="9.35%" id="mcps1.2.7.1.1"><p id="p72510154400"><a name="p72510154400"></a><a name="p72510154400"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="14.99%" id="mcps1.2.7.1.2"><p id="p122531515408"><a name="p122531515408"></a><a name="p122531515408"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.87%" id="mcps1.2.7.1.3"><p id="p1325131584014"><a name="p1325131584014"></a><a name="p1325131584014"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="36.51%" id="mcps1.2.7.1.4"><p id="p225815114016"><a name="p225815114016"></a><a name="p225815114016"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="12.26%" id="mcps1.2.7.1.5"><p id="p2261615184014"><a name="p2261615184014"></a><a name="p2261615184014"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.7.1.6"><p id="p32613153408"><a name="p32613153408"></a><a name="p32613153408"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row1726201544016"><td class="cellrowborder" rowspan="3" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p id="p326111516407"><a name="p326111516407"></a><a name="p326111516407"></a>Volcano Job</p>
<p id="p12475353114815"><a name="p12475353114815"></a><a name="p12475353114815"></a></p>
<p id="p18475175312481"><a name="p18475175312481"></a><a name="p18475175312481"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.99%" headers="mcps1.2.7.1.2 "><p id="p455716252506"><a name="p455716252506"></a><a name="p455716252506"></a><span id="ph1262151402"><a name="ph1262151402"></a><a name="ph1262151402"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.87%" headers="mcps1.2.7.1.3 "><p id="p1126215134018"><a name="p1126215134018"></a><a name="p1126215134018"></a><span id="ph22631519407"><a name="ph22631519407"></a><a name="ph22631519407"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.51%" headers="mcps1.2.7.1.4 "><p id="p1726151594018"><a name="p1726151594018"></a><a name="p1726151594018"></a>a800_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="12.26%" headers="mcps1.2.7.1.5 "><p id="p15261215104017"><a name="p15261215104017"></a><a name="p15261215104017"></a>示例默认为单机16卡任务。</p>
<p id="p8271715184013"><a name="p8271715184013"></a><a name="p8271715184013"></a></p>
<p id="p7271115194015"><a name="p7271115194015"></a><a name="p7271115194015"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15.02%" headers="mcps1.2.7.1.6 "><p id="p142781511408"><a name="p142781511408"></a><a name="p142781511408"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/910b" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
<p id="p71901195499"><a name="p71901195499"></a><a name="p71901195499"></a></p>
<p id="p151911091496"><a name="p151911091496"></a><a name="p151911091496"></a></p>
</td>
</tr>
<tr id="row11271915204017"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p102791534015"><a name="p102791534015"></a><a name="p102791534015"></a><span id="ph15271015144017"><a name="ph15271015144017"></a><a name="ph15271015144017"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p14271815154017"><a name="p14271815154017"></a><a name="p14271815154017"></a>a800_pytorch_vcjob.yaml</p>
</td>
</tr>
<tr id="row14272155406"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p20273150409"><a name="p20273150409"></a><a name="p20273150409"></a><span id="ph62717152401"><a name="ph62717152401"></a><a name="ph62717152401"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p32771519408"><a name="p32771519408"></a><a name="p32771519408"></a>a800_mindspore_vcjob.yaml</p>
</td>
</tr>
<tr id="row728141517408"><td class="cellrowborder" rowspan="3" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p id="p1289158408"><a name="p1289158408"></a><a name="p1289158408"></a>Deployment</p>
<p id="p93517386498"><a name="p93517386498"></a><a name="p93517386498"></a></p>
<p id="p12352113874920"><a name="p12352113874920"></a><a name="p12352113874920"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.99%" headers="mcps1.2.7.1.2 "><p id="p1538185310530"><a name="p1538185310530"></a><a name="p1538185310530"></a><span id="ph2029215114013"><a name="ph2029215114013"></a><a name="ph2029215114013"></a>Atlas 900 A2 PoD 集群基础单元</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.87%" headers="mcps1.2.7.1.3 "><p id="p14296151409"><a name="p14296151409"></a><a name="p14296151409"></a><span id="ph1729111584013"><a name="ph1729111584013"></a><a name="ph1729111584013"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.51%" headers="mcps1.2.7.1.4 "><p id="p15291415194010"><a name="p15291415194010"></a><a name="p15291415194010"></a>a800_tensorflow_deployment.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="12.26%" headers="mcps1.2.7.1.5 "><p id="p142910157401"><a name="p142910157401"></a><a name="p142910157401"></a>示例默认为单机16卡任务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15.02%" headers="mcps1.2.7.1.6 "><p id="p7243709503"><a name="p7243709503"></a><a name="p7243709503"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/910b" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row12914156400"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p172910152406"><a name="p172910152406"></a><a name="p172910152406"></a><span id="ph1029181516406"><a name="ph1029181516406"></a><a name="ph1029181516406"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p2029191584010"><a name="p2029191584010"></a><a name="p2029191584010"></a>a800_pytorch_deployment.yaml</p>
</td>
</tr>
<tr id="row32915158403"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="p7291515164014"><a name="p7291515164014"></a><a name="p7291515164014"></a><span id="ph102941514401"><a name="ph102941514401"></a><a name="ph102941514401"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p202915155405"><a name="p202915155405"></a><a name="p202915155405"></a>a800_mindspore_deployment.yaml</p>
</td>
</tr>
</tbody>
</table>

**表 6** Atlas 训练系列产品支持的YAML

<a name="table21811158146"></a>
<table><thead align="left"><tr id="row10181111518146"><th class="cellrowborder" valign="top" width="9.35%" id="mcps1.2.7.1.1"><p id="p51941552181410"><a name="p51941552181410"></a><a name="p51941552181410"></a>任务类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.2"><p id="p20181111517147"><a name="p20181111517147"></a><a name="p20181111517147"></a>硬件型号</p>
</th>
<th class="cellrowborder" valign="top" width="11.86%" id="mcps1.2.7.1.3"><p id="p5821153911586"><a name="p5821153911586"></a><a name="p5821153911586"></a>训练框架</p>
</th>
<th class="cellrowborder" valign="top" width="36.51%" id="mcps1.2.7.1.4"><p id="p181811156149"><a name="p181811156149"></a><a name="p181811156149"></a>YAML文件名称</p>
</th>
<th class="cellrowborder" valign="top" width="12.280000000000001%" id="mcps1.2.7.1.5"><p id="p86271732132719"><a name="p86271732132719"></a><a name="p86271732132719"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.7.1.6"><p id="p11672113624010"><a name="p11672113624010"></a><a name="p11672113624010"></a>获取链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row71811415111417"><td class="cellrowborder" rowspan="6" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p id="p191941452171418"><a name="p191941452171418"></a><a name="p191941452171418"></a>Volcano Job</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="15%" headers="mcps1.2.7.1.2 "><p id="p218101516149"><a name="p218101516149"></a><a name="p218101516149"></a><span id="ph158146714142"><a name="ph158146714142"></a><a name="ph158146714142"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.86%" headers="mcps1.2.7.1.3 "><p id="zh-cn_topic_0000001609074269_p15865151810597"><a name="zh-cn_topic_0000001609074269_p15865151810597"></a><a name="zh-cn_topic_0000001609074269_p15865151810597"></a><span id="ph12195638125217"><a name="ph12195638125217"></a><a name="ph12195638125217"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.51%" headers="mcps1.2.7.1.4 "><p id="p19182161511148"><a name="p19182161511148"></a><a name="p19182161511148"></a>a800_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="12.280000000000001%" headers="mcps1.2.7.1.5 "><p id="p16627332172713"><a name="p16627332172713"></a><a name="p16627332172713"></a>示例默认为单机8卡任务。</p>
</td>
<td class="cellrowborder" rowspan="12" valign="top" width="15%" headers="mcps1.2.7.1.6 "><p id="p6510121394114"><a name="p6510121394114"></a><a name="p6510121394114"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/910" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
</td>
</tr>
<tr id="row1598044745910"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p208651518105919"><a name="zh-cn_topic_0000001609074269_p208651518105919"></a><a name="zh-cn_topic_0000001609074269_p208651518105919"></a><span id="ph19355165113512"><a name="ph19355165113512"></a><a name="ph19355165113512"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p779714251577"><a name="p779714251577"></a><a name="p779714251577"></a>a800_pytorch_vcjob.yaml</p>
</td>
</tr>
<tr id="row66819525592"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p85061291815"><a name="zh-cn_topic_0000001609074269_p85061291815"></a><a name="zh-cn_topic_0000001609074269_p85061291815"></a><span id="ph13573184092614"><a name="ph13573184092614"></a><a name="ph13573184092614"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p2682524593"><a name="p2682524593"></a><a name="p2682524593"></a>a800_mindspore_vcjob.yaml</p>
</td>
</tr>
<tr id="row181824157147"><td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.1 "><p id="p11182141513140"><a name="p11182141513140"></a><a name="p11182141513140"></a>服务器（插<span id="ph97657495514"><a name="ph97657495514"></a><a name="ph97657495514"></a>Atlas 300T 训练卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="zh-cn_topic_0000001609074269_p764812181714"><a name="zh-cn_topic_0000001609074269_p764812181714"></a><a name="zh-cn_topic_0000001609074269_p764812181714"></a><span id="ph204675452274"><a name="ph204675452274"></a><a name="ph204675452274"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p171821215171419"><a name="p171821215171419"></a><a name="p171821215171419"></a>a300t_tensorflow_vcjob.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.4 "><p id="p5627143215276"><a name="p5627143215276"></a><a name="p5627143215276"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
<tr id="row11788135445919"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p1864871820119"><a name="zh-cn_topic_0000001609074269_p1864871820119"></a><a name="zh-cn_topic_0000001609074269_p1864871820119"></a><span id="ph134441022151619"><a name="ph134441022151619"></a><a name="ph134441022151619"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1798025195718"><a name="p1798025195718"></a><a name="p1798025195718"></a>a300t_pytorch_vcjob.yaml</p>
</td>
</tr>
<tr id="row161351656205911"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p2648618616"><a name="zh-cn_topic_0000001609074269_p2648618616"></a><a name="zh-cn_topic_0000001609074269_p2648618616"></a><span id="ph114081559152716"><a name="ph114081559152716"></a><a name="ph114081559152716"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p9135145619598"><a name="p9135145619598"></a><a name="p9135145619598"></a>a300t_mindspore_vcjob.yaml</p>
</td>
</tr>
<tr id="row1182815141410"><td class="cellrowborder" rowspan="6" valign="top" headers="mcps1.2.7.1.1 "><p id="p1519415221416"><a name="p1519415221416"></a><a name="p1519415221416"></a>Deployment</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.2 "><p id="p151831029101812"><a name="p151831029101812"></a><a name="p151831029101812"></a><span id="ph17662124432"><a name="ph17662124432"></a><a name="ph17662124432"></a>Atlas 800 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="zh-cn_topic_0000001609074269_p12218446435"><a name="zh-cn_topic_0000001609074269_p12218446435"></a><a name="zh-cn_topic_0000001609074269_p12218446435"></a><span id="ph3202748132715"><a name="ph3202748132715"></a><a name="ph3202748132715"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p1118218150140"><a name="p1118218150140"></a><a name="p1118218150140"></a>a800_tensorflow_deployment.yaml</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.5 "><p id="p15627143202718"><a name="p15627143202718"></a><a name="p15627143202718"></a>示例默认为单机8卡任务。</p>
</td>
</tr>
<tr id="row365114361833"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p122181468314"><a name="zh-cn_topic_0000001609074269_p122181468314"></a><a name="zh-cn_topic_0000001609074269_p122181468314"></a><span id="ph724411337162"><a name="ph724411337162"></a><a name="ph724411337162"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p96517361736"><a name="p96517361736"></a><a name="p96517361736"></a>a800_pytorch_deployment.yaml</p>
</td>
</tr>
<tr id="row23490341239"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p142187461335"><a name="zh-cn_topic_0000001609074269_p142187461335"></a><a name="zh-cn_topic_0000001609074269_p142187461335"></a><span id="ph541921262813"><a name="ph541921262813"></a><a name="ph541921262813"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p1734918342033"><a name="p1734918342033"></a><a name="p1734918342033"></a>a800_mindspore_deployment.yaml</p>
</td>
</tr>
<tr id="row11821815111419"><td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.7.1.1 "><p id="p166661021172117"><a name="p166661021172117"></a><a name="p166661021172117"></a>服务器（插<span id="ph39359582495"><a name="ph39359582495"></a><a name="ph39359582495"></a>Atlas 300T 训练卡</span>）</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="zh-cn_topic_0000001609074269_p1160218474318"><a name="zh-cn_topic_0000001609074269_p1160218474318"></a><a name="zh-cn_topic_0000001609074269_p1160218474318"></a><span id="ph168741250162713"><a name="ph168741250162713"></a><a name="ph168741250162713"></a>TensorFlow</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p16182131581416"><a name="p16182131581416"></a><a name="p16182131581416"></a>a300t_tensorflow_deployment.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.4 "><p id="p462719324273"><a name="p462719324273"></a><a name="p462719324273"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
<tr id="row768144119312"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p160284718315"><a name="zh-cn_topic_0000001609074269_p160284718315"></a><a name="zh-cn_topic_0000001609074269_p160284718315"></a><span id="ph162683501617"><a name="ph162683501617"></a><a name="ph162683501617"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p768164120310"><a name="p768164120310"></a><a name="p768164120310"></a>a300t_pytorch_deployment.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p656851019573"><a name="p656851019573"></a><a name="p656851019573"></a>示例默认为单机8卡任务。</p>
</td>
</tr>
<tr id="row3166392032"><td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p id="zh-cn_topic_0000001609074269_p18602047631"><a name="zh-cn_topic_0000001609074269_p18602047631"></a><a name="zh-cn_topic_0000001609074269_p18602047631"></a><span id="ph5731131512820"><a name="ph5731131512820"></a><a name="ph5731131512820"></a>MindSpore</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p id="p958514331647"><a name="p958514331647"></a><a name="p958514331647"></a>a300t_mindspore_deployment.yaml</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.7.1.3 "><p id="p185698104572"><a name="p185698104572"></a><a name="p185698104572"></a>示例默认为单机单卡任务。</p>
</td>
</tr>
</tbody>
</table>

**表 7** Atlas 950 系列产品支持的YAML
<a name="table950yaml"></a>
<table>
    <thead align="left">
        <tr>
            <th class="cellrowborder" valign="top" width="9.35%" id="mcps1.2.7.1.1"><p>任务类型</p></th>
            <th class="cellrowborder" valign="top" width="14.99%" id="mcps1.2.7.1.2"><p>硬件型号</p></th>
            <th class="cellrowborder" valign="top" width="11.87%" id="mcps1.2.7.1.3"><p>训练框架</p></th>
            <th class="cellrowborder" valign="top" width="36.51%" id="mcps1.2.7.1.4"><p>YAML文件名称</p></th>
            <th class="cellrowborder" valign="top" width="12.26%" id="mcps1.2.7.1.5"><p>说明</p></th>
            <th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.7.1.6"><p>获取链接</p></th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td class="cellrowborder" rowspan="2" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p>Volcano Job</p></td>
            <td class="cellrowborder" rowspan="2" valign="top" width="14.99%" headers="mcps1.2.7.1.2 "><p>Atlas 950 SuperPoD</p><p>Atlas 850 系列硬件产品（超节点）</p><p>Atlas 350 标卡</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>PyTorch</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>atlas_950_pytorch_vcjob.yaml</p></td>
            <td class="cellrowborder" rowspan="4" valign="top" width="12.26%" headers="mcps1.2.7.1.5 "><p>示例默认为单机8卡任务。</p></td>
            <td class="cellrowborder" rowspan="4" valign="top" width="15.02%" headers="mcps1.2.7.1.6 ">
                <p><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/950" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
            </td>
        </tr>
        <tr>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>MindSpore</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>atlas_950_mindspore_vcjob.yaml</p></td>
        </tr>
        <tr>
            <td class="cellrowborder" rowspan="2" valign="top" width="9.35%" headers="mcps1.2.7.1.1 "><p>Deployment</p></td>
            <td class="cellrowborder" rowspan="2" valign="top" width="14.99%" headers="mcps1.2.7.1.2 "><p>Atlas 950 SuperPoD</p><p>Atlas 850 系列硬件产品（超节点）</p><p>Atlas 350 标卡</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>PyTorch</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>atlas_950_pytorch_deployment.yaml</p></td>
        </tr>
        <tr>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.1 "><p>MindSpore</p></td>
            <td class="cellrowborder" valign="top" headers="mcps1.2.7.1.2 "><p>atlas_950_mindspore_deployment.yaml</p></td>
        </tr>
    </tbody>
</table>

#### YAML参数说明<a name="ZH-CN_TOPIC_0000002511347099"></a>

本章节提供使用整卡调度或静态vNPU调度配置YAML的操作示例。在操作前，用户需要了解YAML示例的参数说明，再进行操作。

- 使用Ascend Job的用户，请参考[表1](#table159746356276)。
- 使用Volcano Job的用户，请参考[表2](#zh-cn_topic_0000001609074269_table1565872494511)。

**YAML参数说明（acjob任务）<a name="section3507205517910"></a>**

在acjob训练任务中，可使用的YAML参数说明如下表所示。

每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../../api/ascend_job.md)。

**表 1**  YAML参数说明

<a name="table159746356276"></a>
<table><thead align="left"><tr id="row69742355274"><th class="cellrowborder" valign="top" width="27.21%" id="mcps1.2.4.1.1"><p id="p199750359270"><a name="p199750359270"></a><a name="p199750359270"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.18%" id="mcps1.2.4.1.2"><p id="p99751135152710"><a name="p99751135152710"></a><a name="p99751135152710"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.61%" id="mcps1.2.4.1.3"><p id="p39754354276"><a name="p39754354276"></a><a name="p39754354276"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row99751335102713"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p189751735172716"><a name="p189751735172716"></a><a name="p189751735172716"></a>framework</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul4975113512712"></a><a name="ul4975113512712"></a><ul id="ul4975113512712"><li>mindspore</li><li>pytorch</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p79751235102712"><a name="p79751235102712"></a><a name="p79751235102712"></a>框架类型，目前只支持两种。</p>
</td>
</tr>
<tr id="row2097633513272"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p6976153532710"><a name="p6976153532710"></a><a name="p6976153532710"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul16230203215710"></a><a name="ul16230203215710"></a><ul id="ul16230203215710"><li><span id="ph20976435102713"><a name="ph20976435102713"></a><a name="ph20976435102713"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="ph163483412215"><a name="ph163483412215"></a><a name="ph163483412215"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph11976935122715"><a name="ph11976935122715"></a><a name="ph11976935122715"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>Atlas 800 训练服务器，服务器（插<span id="ph2099203201811"><a name="ph2099203201811"></a><a name="ph2099203201811"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li>（可选）Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD取值为：ascend-npu</li>
</ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p139761835102718"><a name="p139761835102718"></a><a name="p139761835102718"></a>用于区分任务使用的芯片的类型。</p>
</td>
</tr>
<tr id="row1067294111274"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p9672541162718"><a name="p9672541162718"></a><a name="p9672541162718"></a>podgroup-sched-enable</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p6895115683718"><a name="p6895115683718"></a><a name="p6895115683718"></a>"true"</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p73261850175013"><a name="p73261850175013"></a><a name="p73261850175013"></a>仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。</p>
<a name="ul12357145235010"></a><a name="ul12357145235010"></a><ul id="ul12357145235010"><li>取值配置为字符串"true"时，表示开启批量调度功能。</li><li>取值配置为其他字符串时，表示批量调度功能不生效，使用普通调度。</li></ul>
<p id="p16526183318537"><a name="p16526183318537"></a><a name="p16526183318537"></a>若不配置该参数，表示批量调度功能不生效，使用普通调度。</p>
<div class="note" id="note1385644144618"><a name="note1385644144618"></a><a name="note1385644144618"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul1431912124720"></a><a name="ul1431912124720"></a><ul id="ul1431912124720"><li>该参数只支持使用<span id="ph1361434468"><a name="ph1361434468"></a><a name="ph1361434468"></a>Volcano</span>调度器的整卡调度特性。</li><li>仅支持在<span id="ph1731512317424"><a name="ph1731512317424"></a><a name="ph1731512317424"></a>Atlas 900 A3 SuperPoD 超节点</span>和<span id="ph1031518232426"><a name="ph1031518232426"></a><a name="ph1031518232426"></a>Atlas 800T A3 超节点服务器</span>中使用本参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="row6977735132710"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1397753512272"><a name="p1397753512272"></a><a name="p1397753512272"></a>schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p5977163532714"><a name="p5977163532714"></a><a name="p5977163532714"></a>默认值为<span class="parmvalue" id="parmvalue99772035172710"><a name="parmvalue99772035172710"></a><a name="parmvalue99772035172710"></a>“volcano”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1397873502710"><a name="p1397873502710"></a><a name="p1397873502710"></a><span id="ph8978183512271"><a name="ph8978183512271"></a><a name="ph8978183512271"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
</td>
</tr>
<tr id="row169789353270"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p0978123511271"><a name="p0978123511271"></a><a name="p0978123511271"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p2978113514276"><a name="p2978113514276"></a><a name="p2978113514276"></a>默认值为任务总副本数</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p897873582712"><a name="p897873582712"></a><a name="p897873582712"></a><span id="ph129781235182715"><a name="ph129781235182715"></a><a name="ph129781235182715"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="ph197814351275"><a name="ph197814351275"></a><a name="ph197814351275"></a>Volcano</span>时，任务运行总副本数。</p>
</td>
</tr>
<tr id="row997815356274"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p16978193511276"><a name="p16978193511276"></a><a name="p16978193511276"></a>queue</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p10978123512277"><a name="p10978123512277"></a><a name="p10978123512277"></a>默认值为<span class="parmvalue" id="parmvalue199781835112716"><a name="parmvalue199781835112716"></a><a name="parmvalue199781835112716"></a>“default”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1497812357278"><a name="p1497812357278"></a><a name="p1497812357278"></a><span id="ph0978173562717"><a name="ph0978173562717"></a><a name="ph0978173562717"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="ph797816353270"><a name="ph797816353270"></a><a name="ph797816353270"></a>Volcano</span>时，任务所属队列。</p>
</td>
</tr>
<tr id="row1397863562711"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p8979163582710"><a name="p8979163582710"></a><a name="p8979163582710"></a>（可选）successPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul897943582712"></a><a name="ul897943582712"></a><ul id="ul897943582712"><li>默认值为空，若用户不填写该参数，则默认取空值</li><li>AllWorkers</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p11979535122719"><a name="p11979535122719"></a><a name="p11979535122719"></a>表明任务成功的前提。空值代表只需要一个<span id="ph149798358273"><a name="ph149798358273"></a><a name="ph149798358273"></a>Pod</span>成功，整个任务判定为成功。取值为<span class="parmvalue" id="parmvalue1797983518271"><a name="parmvalue1797983518271"></a><a name="parmvalue1797983518271"></a>“AllWorkers”</span>表示所有<span id="ph8979163522716"><a name="ph8979163522716"></a><a name="ph8979163522716"></a>Pod</span>都成功，任务才判定为成功。</p>
</td>
</tr>
<tr id="row16979103514275"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p7979123522712"><a name="p7979123522712"></a><a name="p7979123522712"></a>container.name</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p39791535192717"><a name="p39791535192717"></a><a name="p39791535192717"></a>ascend</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p3979235182716"><a name="p3979235182716"></a><a name="p3979235182716"></a>容器的名称必须是<span class="parmvalue" id="parmvalue12979735192712"><a name="parmvalue12979735192712"></a><a name="parmvalue12979735192712"></a>“ascend”</span>。</p>
</td>
</tr>
<tr id="row20979735172712"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p4979143514276"><a name="p4979143514276"></a><a name="p4979143514276"></a>（可选）ports</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p1297916359270"><a name="p1297916359270"></a><a name="p1297916359270"></a>若用户未进行设置，系统默认填写以下参数：</p>
<a name="ul69804359278"></a><a name="ul69804359278"></a><ul id="ul69804359278"><li>name: ascendjob-port</li><li>containerPort: 2222</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p2980183592716"><a name="p2980183592716"></a><a name="p2980183592716"></a>分布式训练集合通信端口。<span class="parmname" id="parmname1198063542717"><a name="parmname1198063542717"></a><a name="parmname1198063542717"></a>“name”</span>取值只能为<span class="parmvalue" id="parmvalue17980153515270"><a name="parmvalue17980153515270"></a><a name="parmvalue17980153515270"></a>“ascendjob-port”</span>，<span class="parmname" id="parmname8980135102711"><a name="parmname8980135102711"></a><a name="parmname8980135102711"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
</td>
</tr>
<tr id="row12980535152720"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1398023582714"><a name="p1398023582714"></a><a name="p1398023582714"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul498063562713"></a><a name="ul498063562713"></a><ul id="ul498063562713"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1798016353272"><a name="p1798016353272"></a><a name="p1798016353272"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="row1198010356278"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1398113517275"><a name="p1398113517275"></a><a name="p1398113517275"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p209811535192718"><a name="p209811535192718"></a><a name="p209811535192718"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1898143513271"><a name="p1898143513271"></a><a name="p1898143513271"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr id="row1298116359278"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p898193522717"><a name="p898193522717"></a><a name="p898193522717"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p149811335192718"><a name="p149811335192718"></a><a name="p149811335192718"></a><span id="ph8981193532719"><a name="ph8981193532719"></a><a name="ph8981193532719"></a>ARM</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
<p id="p2098193513276"><a name="p2098193513276"></a><a name="p2098193513276"></a><span id="ph2098163502710"><a name="ph2098163502710"></a><a name="ph2098163502710"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p6981203592717"><a name="p6981203592717"></a><a name="p6981203592717"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="p99813357275"><a name="p99813357275"></a><a name="p99813357275"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="row711915310345"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p19119935347"><a name="p19119935347"></a><a name="p19119935347"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p131191634340"><a name="p131191634340"></a><a name="p131191634340"></a>pod：只支持Pod级重调度，不升级为Job级别。（当使用vcjob时，需要配置该策略：policies: -event:PodFailed -action:RestartTask）</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p21194343415"><a name="p21194343415"></a><a name="p21194343415"></a>任务重调度策略。</p>
</td>
</tr>
<tr id="row14366105123419"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p123665543412"><a name="p123665543412"></a><a name="p123665543412"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p636605193410"><a name="p636605193410"></a><a name="p636605193410"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p436675193411"><a name="p436675193411"></a><a name="p436675193411"></a>任务能够调度的最小副本数。</p>
</td>
</tr>
<tr id="row7926131215173"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1377484145110"><a name="p1377484145110"></a><a name="p1377484145110"></a><span id="ph19817116185120"><a name="ph19817116185120"></a><a name="ph19817116185120"></a>huawei.com/schedule_policy</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p1877410425111"><a name="p1877410425111"></a><a name="p1877410425111"></a><span id="ph135426519519"><a name="ph135426519519"></a><a name="ph135426519519"></a>目前支持</span><a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p1671165052112"><a name="p1671165052112"></a><a name="p1671165052112"></a>配置任务需要调度的AI芯片布局形态。<span id="ph204811934163414"><a name="ph204811934163414"></a><a name="ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
</td>
</tr>
<tr id="row2098119351272"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p12982163592711"><a name="p12982163592711"></a><a name="p12982163592711"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p159821035122717"><a name="p159821035122717"></a><a name="p159821035122717"></a>指定逻辑超节点芯片数量。</p>
<a name="ul10451144414619"></a><a name="ul10451144414619"></a><ul id="ul10451144414619"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p398216353271"><a name="p398216353271"></a><a name="p398216353271"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="ph521204025916"><a name="ph521204025916"></a><a name="ph521204025916"></a>若用户未指定该字段，</span><span id="ph172121408590"><a name="ph172121408590"></a><a name="ph172121408590"></a>Volcano</span><span id="ph192121140135911"><a name="ph192121140135911"></a><a name="ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19975815131410"><a name="p19975815131410"></a><a name="p19975815131410"></a>了解详细说明请参见<a href="./01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note1998233513279"><a name="note1998233513279"></a><a name="note1998233513279"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul546892712569"></a><a name="ul546892712569"></a><ul id="ul546892712569"><li>仅支持在Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点服务器、Atlas 800I A3 超节点服务器中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row14982435112715"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p18982635202720"><a name="p18982635202720"></a><a name="p18982635202720"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul109821735182715"></a><a name="ul109821735182715"></a><ul id="ul109821735182715"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="note7983133512715"><a name="note7983133512715"></a><a name="note7983133512715"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p16983103515274"><a name="p16983103515274"></a><a name="p16983103515274"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p2983173572716"><a name="p2983173572716"></a><a name="p2983173572716"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p><ul id="ul59831535122714"><li>交换机亲和性调度1.0版本支持<span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="ph168598363399"><a name="ph168598363399"></a><a name="ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917_2"><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph4181625925"><a name="ph4181625925"></a><a name="ph4181625925"></a>PyTorch</span>和<span id="ph61882510210"><a name="ph61882510210"></a><a name="ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph311717506401"><a name="ph311717506401"></a><a name="ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_3"><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph619244413568"><a name="ph619244413568"></a><a name="ph619244413568"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>
</td>
</tr>
<tr id="row9984235132714"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p18984535132720"><a name="p18984535132720"></a><a name="p18984535132720"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p16729326173616"><a name="p16729326173616"></a><a name="p16729326173616"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul139845353279"></a><a name="ul139845353279"></a><ul id="ul139845353279"><li><span id="ph169841935102711"><a name="ph169841935102711"></a><a name="ph169841935102711"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="ph10984193517273"><a name="ph10984193517273"></a><a name="ph10984193517273"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="ph169855357273"><a name="ph169855357273"></a><a name="ph169855357273"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span>和<span id="ph14985135162710"><a name="ph14985135162710"></a><a name="ph14985135162710"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="ph1898523510277"><a name="ph1898523510277"></a><a name="ph1898523510277"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li><li><span id="ph19985163514279"><a name="ph19985163514279"></a><a name="ph19985163514279"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="ph4985183516277"><a name="ph4985183516277"></a><a name="ph4985183516277"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-16</li><li><span id="ph1514953013253"><a name="ph1514953013253"></a><a name="ph1514953013253"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph8619174411286"><a name="ph8619174411286"></a><a name="ph8619174411286"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签。</li><li><span id="ph261924414289"><a name="ph261924414289"></a><a name="ph261924414289"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>（可选）Atlas 350 标卡</span>：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>（可选）Atlas 850 系列硬件产品：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>（可选）Atlas 950 SuperPoD：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p5986235142714"><a name="p5986235142714"></a><a name="p5986235142714"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
<div class="note" id="note898773512719"><a name="note898773512719"></a><a name="note898773512719"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1027616512420"><a name="p1027616512420"></a><a name="p1027616512420"></a><span id="ph9014016509"><a name="ph9014016509"></a><a name="ph9014016509"></a>下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row1598720359275"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p2987143519279"><a name="p2987143519279"></a><a name="p2987143519279"></a>requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p1198733572715"><a name="p1198733572715"></a><a name="p1198733572715"></a><strong id="b2098733513275"><a name="b2098733513275"></a><a name="b2098733513275"></a>整卡调度：</strong></p>
<p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为huawei.com/npu: <em>x</em></p>
<p id="p998783518273"><a name="p998783518273"></a><a name="p998783518273"></a>其他产品配置为huawei.com/Ascend910: <em id="i6988935192715"><a name="i6988935192715"></a><a name="i6988935192715"></a>x</em></p>
<p id="p39887359272"><a name="p39887359272"></a><a name="p39887359272"></a>根据所使用芯片类型不同，x取值如下：</p>
<a name="ul1798823522713"></a><a name="ul1798823522713"></a><ul id="ul1798823522713"><li><span id="ph1598853514277"><a name="ph1598853514277"></a><a name="ph1598853514277"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="ul1998810354278"></a><a name="ul1998810354278"></a><ul id="ul1998810354278"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul>
</li><li><span id="ph2988153582716"><a name="ph2988153582716"></a><a name="ph2988153582716"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="ul159887358276"></a><a name="ul159887358276"></a><ul id="ul159887358276"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul>
</li><li>服务器（插<span id="ph15989835132717"><a name="ph15989835132717"></a><a name="ph15989835132717"></a>Atlas 300T 训练卡</span>）：<a name="ul18989235102719"></a><a name="ul18989235102719"></a><ul id="ul18989235102719"><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul>
</li><li><span id="ph52261657163417"><a name="ph52261657163417"></a><a name="ph52261657163417"></a>Atlas 800T A2 训练服务器</span>和<span id="ph999063517276"><a name="ph999063517276"></a><a name="ph999063517276"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="ul199020351273"></a><a name="ul199020351273"></a><ul id="ul199020351273"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="ph18991173510273"><a name="ph18991173510273"></a><a name="ph18991173510273"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="ul69913354274"></a><a name="ul69913354274"></a><ul id="ul69913354274"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li><li><span id="ph18991153532717"><a name="ph18991153532717"></a><a name="ph18991153532717"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph20855174373110"><a name="ph20855174373110"></a><a name="ph20855174373110"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph27041232195"><a name="ph27041232195"></a><a name="ph27041232195"></a>Atlas 800T A3 超节点服务器</span>：<a name="ul499153515277"></a><a name="ul499153515277"></a><ul id="ul499153515277"><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对<span id="ph798511112819"><a name="ph798511112819"></a><a name="ph798511112819"></a>Atlas 900 A3 SuperPoD 超节点</span>的逻辑超节点亲和任务：16</li></ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内8卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8</li>
        <li>分布式：1、2、3、4、5、6、7、8</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内16卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
        <li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 8卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8</li>
        <li>单机（不保证亲和性）：5、6、7</li>
        <li>分布式（满足亲和性）：1、2、3、4、8</li>
        <li>分布式（不保证亲和性）：5、6、7</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 16卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
        <li>分布式（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（普通集群）</span>：
    <ul>
        <li>单机：1、2、4、8</li>
        <li>分布式：1、2、4、8</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（超节点集群）</span>：
    <ul>
        <li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD 集群</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
</ul>
<p id="p4993163513278"><a name="p4993163513278"></a><a name="p4993163513278"></a><strong id="b099353510277"><a name="b099353510277"></a><a name="b099353510277"></a>静态vNPU调度：</strong></p>
<p id="p39931635142716"><a name="p39931635142716"></a><a name="p39931635142716"></a>huawei.com/Ascend910-<strong id="b12993153514272"><a name="b12993153514272"></a><a name="b12993153514272"></a><em id="i16993153522711"><a name="i16993153522711"></a><a name="i16993153522711"></a>Y</em></strong>: 1</p>
<p id="p1399323512713"><a name="p1399323512713"></a><a name="p1399323512713"></a>取值为1。只能使用一个NPU下的vNPU。</p>
<p id="p119931335142715"><a name="p119931335142715"></a><a name="p119931335142715"></a>如huawei.com/Ascend910-<em id="i20993935152714"><a name="i20993935152714"></a><a name="i20993935152714"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p4994193520276"><a name="p4994193520276"></a><a name="p4994193520276"></a>至少请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<div class="note" id="note6994535162719"><a name="note6994535162719"></a><a name="note6994535162719"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p169941635192717"><a name="p169941635192717"></a><a name="p169941635192717"></a><strong id="b179941535122710"><a name="b179941535122710"></a><a name="b179941535122710"></a><em id="i1999493582720"><a name="i1999493582720"></a><a name="i1999493582720"></a>Y</em></strong>取值可参考<a href="../virtual_instance/virtual_instance_with_hdk/06_mounting_vnpu.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与虚拟设备类型关系表的“vNPU类型”列。</p>
<p id="p17994143511279"><a name="p17994143511279"></a><a name="p17994143511279"></a>以vNPU类型Ascend910-6c.1cpu.16g为例，<strong id="b29941735102716"><a name="b29941735102716"></a><a name="b29941735102716"></a><em id="i11995143582719"><a name="i11995143582719"></a><a name="i11995143582719"></a>Y</em></strong>取值为6c.1cpu.16g，不包括前面的Ascend910。</p>
<p id="p699519351275"><a name="p699519351275"></a><a name="p699519351275"></a>虚拟化模板的更多信息可以参考<a href="../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md">虚拟化模板</a>章节。</p>
</div></div>
</td>
</tr>
<tr id="row1899517352272"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p999523582714"><a name="p999523582714"></a><a name="p999523582714"></a>limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p203039915181"><a name="p203039915181"></a><a name="p203039915181"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<p id="p4739213121713"><a name="p4739213121713"></a><a name="p4739213121713"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
</td>
</tr>
<tr id="row119951035132715"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p6995103513275"><a name="p6995103513275"></a><a name="p6995103513275"></a>metadata.annotations['huawei.com/Ascend<em id="i1599519355274"><a name="i1599519355274"></a><a name="i1599519355274"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p12995163592715"><a name="p12995163592715"></a><a name="p12995163592715"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 ">
    <p id="p17995235152710"><a name="p17995235152710"></a><a name="p17995235152710"></a><span id="ph13995133582718"><a name="ph13995133582718"></a><a name="ph13995133582718"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
    <div class="note" id="note6995123592720"><a name="note6995123592720"></a><a name="note6995123592720"></a><span class="notetitle">[!NOTE] 说明</span>
        <div class="notebody">
            <ul>
                <li>
                    <p id="p1299613592712"><a name="p1299613592712"></a><a name="p1299613592712"></a>该参数只支持使用<span id="ph699616354271"><a name="ph699616354271"></a><a name="ph699616354271"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
                </li>
                <li><p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</p></li>
            </ul>
        </div>
    </div>
</td>
</tr>
<tr id="row158871116111016"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p1062816462407"><a name="p1062816462407"></a><a name="p1062816462407"></a>hostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><a name="ul960434424111"></a><a name="ul960434424111"></a><ul id="ul960434424111"><li>true：使用HostIP创建Pod。<p id="p3565137193114"><a name="p3565137193114"></a><a name="p3565137193114"></a>此种情况下，需要在YAML中同步配置环境变量HCCL_IF_IP为status.hostIP。</p>
</li><li>false：不使用HostIP创建Pod。<p id="p15499134563115"><a name="p15499134563115"></a><a name="p15499134563115"></a>未传入此参数或此参数的值为false时，不需要配置上述环境变量。</p>
<p id="p8537842173118"><a name="p8537842173118"></a><a name="p8537842173118"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><a name="ul5695536155919"></a><a name="ul5695536155919"></a><ul id="ul5695536155919"><li>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</li><li>不传入此参数时，默认不使用HostIP创建Pod。<div class="note" id="note1962023375819"><a name="note1962023375819"></a><a name="note1962023375819"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p21989143551"><a name="p21989143551"></a><a name="p21989143551"></a>当采用HostIp方式创建Pod，依然存在创建Pod速度慢且Pod之间通信速度慢的问题。此时推荐采用挂载RankTable文件的方式，通过解析RankTable文件获得Pod的hostIP，并将其注入到对应框架任务的环境变量中（如ms框架注入到环境变量MS_SCHED_HOST中），实现建链。</p>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row1175517234373"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p2755623183718"><a name="p2755623183718"></a><a name="p2755623183718"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="p20956101093820"><a name="p20956101093820"></a><a name="p20956101093820"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul959814611371"></a><a name="ul959814611371"></a><ul id="ul959814611371"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="p20494202617363"><a name="p20494202617363"></a><a name="p20494202617363"></a>仅支持在<span id="ph7230184917387"><a name="ph7230184917387"></a><a name="ph7230184917387"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</td>
</tr>
<tr id="rowcustomjobkey"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="pcustomjobkey"><a name="pcustomjobkey"></a><a name="pcustomjobkey"></a>customJobKey</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="pcustomjobkeyvalue"><a name="pcustomjobkeyvalue"></a><a name="pcustomjobkeyvalue"></a>用户自定义标签，以二级跳转的方式设置作业唯一标识符，如：<br> customJobKey: tid<br> tid: "123456"</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.61%" headers="mcps1.2.4.1.3 "><p id="pcustomjobkeydesc"><a name="pcustomjobkeydesc"></a><a name="pcustomjobkeydesc"></a>支持通过customJobKey或custom-job-id设置作业唯一标识符，方便用户根据该标识符过滤作业相关的告警、ISSUE等关键信息。<br> 在资源AscendJob的metadata.labels标签中设置。</p>
</td>
</tr>
<tr id="rowcustomjobid"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="pcustomjobid"><a name="pcustomjobid"></a><a name="pcustomjobid"></a>custom-job-id</p>
</td>
<td class="cellrowborder" valign="top" width="36.18%" headers="mcps1.2.4.1.2 "><p id="pcustomjobidvalue"><a name="pcustomjobidvalue"></a><a name="pcustomjobidvalue"></a>用户自定义标签，直接设置作业唯一标识符，如：<br> custom-job-id："123456"</p>
</td>
</tr>
</tbody>
</table>

**YAML参数说明（deploy任务或vcjob任务）<a name="section645410711512"></a>**

在deploy任务或vcjob训练任务中，可使用的YAML参数说明如下表所示。

**表 2**  YAML参数说明

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="22.58%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="40.86%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001609074269_p19658152414451"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001609074269_ul1531417539259"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001609074269_p11302326164814"></a>N为节点个数，Deployment类型的任务不需要该参数，该参数建议与replicas保持一致。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001609074269_ph16676195493717"></a>ARM</span>环境：<span id="ph4569134274515"><a name="ph4569134274515"></a><a name="ph4569134274515"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：<span id="ph7394135434515"><a name="ph7394135434515"></a><a name="ph7394135434515"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="row319913141385"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p17879179384"><a name="p17879179384"></a><a name="p17879179384"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p11787717143811"><a name="p11787717143811"></a><a name="p11787717143811"></a>pod：只支持Pod级重调度，不升级为Job级别。（当使用vcjob时，需要配置该策略：policies: -event:PodFailed -action:RestartTask）</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1278741713381"><a name="p1278741713381"></a><a name="p1278741713381"></a>任务重调度策略。</p>
</td>
</tr>
<tr id="row675991618389"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p778791715380"><a name="p778791715380"></a><a name="p778791715380"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1378781718388"><a name="p1378781718388"></a><a name="p1378781718388"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1378741712380"><a name="p1378741712380"></a><a name="p1378741712380"></a>任务能够调度的最小副本数。</p>
</td>
</tr>
<tr id="row492051125013"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
</td>
</tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>指定逻辑超节点芯片数量。</p>
<a name="ul1150756144219"></a><a name="ul1150756144219"></a><ul id="ul1150756144219"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p1250719624216"><a name="p1250719624216"></a><a name="p1250719624216"></a>了解详细说明请参见<a href="./01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点服务器、Atlas 800I A3 超节点服务器中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row862818313577"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p132726845716"><a name="p132726845716"></a><a name="p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="ul1427218195710"></a><a name="ul1427218195710"></a><ul id="ul1427218195710"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="note32586245294"><a name="note32586245294"></a><a name="note32586245294"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p5258102462916"><a name="p5258102462916"></a><a name="p5258102462916"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p32732087577"><a name="p32732087577"></a><a name="p32732087577"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p><ul id="ul961424647"><li>交换机亲和性调度1.0版本支持<span id="ph63831524184110"><a name="ph63831524184110"></a><a name="ph63831524184110"></a>Atlas 训练系列产品</span>和<span id="ph138318245414"><a name="ph138318245414"></a><a name="ph138318245414"></a><term id="zh-cn_topic_0000001519959665_term57208119917_4"><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph17383182419412"><a name="ph17383182419412"></a><a name="ph17383182419412"></a>PyTorch</span>和<span id="ph1383224134120"><a name="ph1383224134120"></a><a name="ph1383224134120"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="ph438320243412"><a name="ph438320243412"></a><a name="ph438320243412"></a><term id="zh-cn_topic_0000001519959665_term57208119917_5"><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph134821711841"><a name="ph134821711841"></a><a name="ph134821711841"></a>PyTorch</span>框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001609074269_p1449413229314"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p7665323173618"><a name="p7665323173618"></a><a name="p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul14200073713"></a><a name="ul14200073713"></a><ul id="ul14200073713"><li><span id="zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000001609074269_ph1284164912438"><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a><a name="zh-cn_topic_0000001609074269_ph1284164912438"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph4528511506"><a name="zh-cn_topic_0000001609074269_ph4528511506"></a><a name="zh-cn_topic_0000001609074269_ph4528511506"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph486033685311"><a name="ph486033685311"></a><a name="ph486033685311"></a>Atlas 800T A2 训练服务器</span>和<span id="ph1296712308221"><a name="ph1296712308221"></a><a name="ph1296712308221"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="ph4487202241512"><a name="ph4487202241512"></a><a name="ph4487202241512"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-8</li><li><span id="ph1114211211203"><a name="ph1114211211203"></a><a name="ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="ph5811017182112"><a name="ph5811017182112"></a><a name="ph5811017182112"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_4"><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_4"></a>{xxx}</em></span>b-16</li><li><span id="ph115277505269"><a name="ph115277505269"></a><a name="ph115277505269"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph7730165573912"><a name="ph7730165573912"></a><a name="ph7730165573912"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签。</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li>（可选）Atlas 350 标卡：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>（可选）Atlas 850 系列硬件产品：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>（可选）Atlas 950 SuperPoD：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1954213851616"><a name="p1954213851616"></a><a name="p1954213851616"></a>根据需要运行训练任务的节点类型，选取不同的值。</p>
<div class="note" id="note19666163011214"><a name="note19666163011214"></a><a name="note19666163011214"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1105153313533"><a name="p1105153313533"></a><a name="p1105153313533"></a><span id="ph710573305319"><a name="ph710573305319"></a><a name="ph710573305319"></a>下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209_1"><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_1"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1996615912482"><a name="p1996615912482"></a><a name="p1996615912482"></a><strong id="b118963916494"><a name="b118963916494"></a><a name="b118963916494"></a>整卡调度：</strong></p>
<p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为huawei.com/npu: <em>x</em></p>
<p id="p44675644812"><a name="p44675644812"></a><a name="p44675644812"></a><span id="ph1567213131613"><a name="ph1567213131613"></a><a name="ph1567213131613"></a>其他产品配置为huawei.com/Ascend910</span>: <em id="i2478131910511"><a name="i2478131910511"></a><a name="i2478131910511"></a>x</em></p>
<p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>根据所使用芯片类型不同，x取值如下：</p>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a><ul id="ul4403181216571"><li><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</li><li><span id="zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001609074269_ul1713712328597"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</li><li>服务器（插<span id="zh-cn_topic_0000001609074269_ph1223449506"><a name="zh-cn_topic_0000001609074269_ph1223449506"></a><a name="zh-cn_topic_0000001609074269_ph1223449506"></a>Atlas 300T 训练卡</span>）：<a name="ul3519194217372"></a><a name="ul3519194217372"></a><ul id="ul3519194217372"><li>单机单芯片：1</li><li>单机多芯片：2</li><li>分布式：2</li></ul>
</li><li><span id="ph1176216314557"><a name="ph1176216314557"></a><a name="ph1176216314557"></a>Atlas 800T A2 训练服务器</span>和<span id="ph107421743105017"><a name="ph107421743105017"></a><a name="ph107421743105017"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="ul169264817234"></a><a name="ul169264817234"></a><ul id="ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="ph129391532155719"><a name="ph129391532155719"></a><a name="ph129391532155719"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="ul555885820439"></a><a name="ul555885820439"></a><ul id="ul555885820439"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li><li><span id="ph133001904447"><a name="ph133001904447"></a><a name="ph133001904447"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph830011074420"><a name="ph830011074420"></a><a name="ph830011074420"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 超节点服务器</span>：<a name="ul130020074415"></a><a name="ul130020074415"></a><ul id="ul130020074415"><li>单机多芯片：2、4、6、8、10、12、14、16</li><li>分布式：16</li></ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内8卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8</li>
        <li>分布式：1、2、3、4、5、6、7、8</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（无互联节点内16卡）</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
        <li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 8卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8</li>
        <li>单机（不保证亲和性）：5、6、7</li>
        <li>分布式（满足亲和性）：1、2、3、4、8</li>
        <li>分布式（不保证亲和性）：5、6、7</li>
    </ul>
</li>
<li>
    <span>Atlas 350 标卡（4P mesh 16卡）</span>：
    <ul>
        <li>单机（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
        <li>分布式（满足亲和性）：1、2、3、4、8、12、16</li>
        <li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（普通集群）</span>：
    <ul>
        <li>单机：1、2、4、8</li>
        <li>分布式：1、2、4、8</li>
    </ul>
</li>
<li>
    <span>Atlas 850 系列硬件产品（超节点集群）</span>：
    <ul>
        <li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>：
    <ul>
        <li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li>
        <li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li>
    </ul>
</li>
</ul>
<p id="p1498123034911"><a name="p1498123034911"></a><a name="p1498123034911"></a><strong id="b7488133134911"><a name="b7488133134911"></a><a name="b7488133134911"></a>静态vNPU调度：</strong></p>
<p id="p19104113195111"><a name="p19104113195111"></a><a name="p19104113195111"></a>huawei.com/Ascend910-<strong id="b14105734512"><a name="b14105734512"></a><a name="b14105734512"></a><em id="i17105533512"><a name="i17105533512"></a><a name="i17105533512"></a>Y</em></strong>: 1</p>
<p id="p1851116142917"><a name="p1851116142917"></a><a name="p1851116142917"></a>取值为1。只能使用一个NPU下的vNPU。</p>
<p id="p11413153312435"><a name="p11413153312435"></a><a name="p11413153312435"></a>如huawei.com/Ascend910-<em id="i94134332434"><a name="i94134332434"></a><a name="i94134332434"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5498134535310"><a name="p5498134535310"></a><a name="p5498134535310"></a>至少请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<div class="note" id="note1648201912419"><a name="note1648201912419"></a><a name="note1648201912419"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p124851954116"><a name="p124851954116"></a><a name="p124851954116"></a><strong id="b1327175513443"><a name="b1327175513443"></a><a name="b1327175513443"></a><em id="i827655184412"><a name="i827655184412"></a><a name="i827655184412"></a>Y</em></strong>取值可参考<a href="../virtual_instance/virtual_instance_with_hdk/06_mounting_vnpu.md#静态虚拟化">静态虚拟化</a>章节中的虚拟化实例模板与虚拟设备类型关系表的“vNPU类型”列。</p>
<p id="p128388914430"><a name="p128388914430"></a><a name="p128388914430"></a>以vNPU类型Ascend910-6c.1cpu.16g为例，<strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong>取值为6c.1cpu.16g，不包括前面的Ascend910。</p>
<p id="p2491818192318"><a name="p2491818192318"></a><a name="p2491818192318"></a>虚拟化模板的更多信息可以参考<a href="../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md">虚拟化模板</a>章节。</p>
</div></div>
</td>
</tr>
<tr id="row25918533287"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p05117110298"><a name="p05117110298"></a><a name="p05117110298"></a>limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p13683185074711"><a name="p13683185074711"></a><a name="p13683185074711"></a>限制请求的NPU或vNPU类型（只能请求一种类型）、数量，请根据实际修改。</p>
<p id="p16683135019479"><a name="p16683135019479"></a><a name="p16683135019479"></a>limits需要和requests的芯片名称和数量需保持一致。</p>
</td>
</tr>
<tr id="row14747131720228"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/Ascend<em id="i103895254475"><a name="i103895254475"></a><a name="i103895254475"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 ">
    <p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
<div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle">[!NOTE] 说明</span>
    <div class="notebody">
        <ul>
            <li>
                <p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>该参数只支持使用<span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span>调度器的整卡调度特性。使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
            </li>
            <li><p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</p></li>
        </ul>
</div>
</div>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1941725316543"><a name="p1941725316543"></a><a name="p1941725316543"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="ul2750122165318"></a><a name="ul2750122165318"></a><ul id="ul2750122165318"><li>Atlas 800 训练服务器，服务器（插<span id="ph6581133055411"><a name="ph6581133055411"></a><a name="ph6581133055411"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li><span id="ph10656173717129"><a name="ph10656173717129"></a><a name="ph10656173717129"></a><term id="zh-cn_topic_0000001519959665_term57208119917_6"><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a>Atlas A2 训练系列产品</term></span>、<span id="ph1665620377128"><a name="ph1665620377128"></a><a name="ph1665620377128"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph14656337131215"><a name="ph14656337131215"></a><a name="ph14656337131215"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph12656113717123"><a name="ph12656113717123"></a><a name="ph12656113717123"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph1265633714121"><a name="ph1265633714121"></a><a name="ph1265633714121"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b</li><li>（可选）Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD取值为：ascend-npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>用于区分任务使用的芯片的类型。需要在<span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
<div class="note" id="note14282027593"><a name="note14282027593"></a><a name="note14282027593"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1328162720912"><a name="p1328162720912"></a><a name="p1328162720912"></a><span id="ph19729197"><a name="ph19729197"></a><a name="ph19729197"></a>此处的{<em id="zh-cn_topic_0000001519959665_i1914312018209_2"><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="row141124616406"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>超节点任务使用的亲和性调度策略，需要用户在YAML的label中声明。</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft：集群资源不满足超节点亲和性时，任务使用集群中碎片资源继续调度。</li><li>hard：集群资源不满足超节点亲和性时，任务Pending，等待资源。</li><li>其他值或不传入此参数：强制超节点亲和性调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>仅支持在<span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用本参数。</p>
</td>
</tr>
<tr id="rowcustomjobkey2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobkey2"><a name="pcustomjobkey2"></a><a name="pcustomjobkey2"></a>customJobKey</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobkeyvalue2"><a name="pcustomjobkeyvalue2"></a><a name="pcustomjobkeyvalue2"></a>用户自定义标签，以二级跳转的方式设置作业唯一标识符，如：<br> customJobKey: tid<br> tid: "123456"</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="pcustomjobkeydesc2"><a name="pcustomjobkeydesc2"></a><a name="pcustomjobkeydesc2"></a>支持通过customJobKey或custom-job-id设置作业唯一标识符，方便用户根据该标识符过滤作业相关的告警、ISSUE等关键信息。<br> <ul><li>vcjob任务在资源Job的metadata.labels标签中设置。<br></li> <li>deploy任务在资源Deployment的spec.template.metadata.labels标签中设置。</li></ul></p>
</td>
</tr>
<tr id="rowcustomjobid2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobid2"><a name="pcustomjobid2"></a><a name="pcustomjobid2"></a>custom-job-id</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobidvalue2"><a name="pcustomjobidvalue2"></a><a name="pcustomjobidvalue2"></a>用户自定义标签，直接设置作业唯一标识符，如：<br> custom-job-id："123456"</p>
</td>
</tr>
</tbody>
</table>

**表 3**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>

|配置|说明|
|--|--|
|chip4-node8|1个节点8张芯片，每4个芯片形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的整模块场景/Atlas 350 标卡共8张卡，每4张卡通过UB扣板连接。|
|chip1-node2|1个节点2张芯片。例如，Atlas 300T 训练卡的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。|
|chip4-node4|1个节点4张芯片，形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的半配场景。|
|chip8-node8|1个节点8张卡，8张卡都在1个互联环上。例如，Atlas 800T A2 训练服务器 /Atlas 850 系列硬件产品。|
|chip8-node16|1个节点16张卡，每8张卡在1个互联环上。例如，Atlas 200T A2 Box16 异构子框。|
|chip2-node8|1个节点8张卡，每2张卡在1个互联环上。|
|chip2-node16|1个节点16张卡，每2张卡在1个互联环上。例如，Atlas 800T A3 超节点服务器。|
|chip2-node8-sp|1个节点8张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 9000 A3 SuperPoD 集群算力系统。|
|chip2-node16-sp|1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 900 A3 SuperPoD 超节点。|
|chip4-node16|1个节点16张卡，每4张卡都在1个互联环上。例如，Atlas 350 标卡共16张卡，每4张卡通过UB扣板连接。|
|chip1-node8|1个节点8张卡，每张卡之间无互联。例如，Atlas 350 标卡共8张卡，每张卡之间无互联。|
|chip1-node16|1个节点16张卡，每张卡之间无互联。例如，Atlas 350 标卡共16张卡，每张卡之间无互联。|
|chip8-node8-sp|1个节点8张卡，8张卡都在1个互联环上，多个服务器形成超节点。例如，Atlas 850 系列硬件产品（超节点服务器）。|
|chip8-node8-ra64-sp|1个节点8张卡，8张卡都在1个互联环上，64个节点组成一个计算框，多个框形成超节点。例如，Atlas 950 SuperPoD。|
|chip1-softShareDev|软切分虚拟化专用调度策略。|
|multilevel|多级调度场景使用，多级调度的详细使用方法请参见[多级调度](./05_multi_level_scheduling.md)。|

#### 配置YAML<a name="ZH-CN_TOPIC_0000002511347101"></a>

本章节指导用户配置整卡调度或静态vNPU调度特性的任务YAML，通过环境变量配置资源信息的用户请参考[通过环境变量配置资源信息场景](#section598118132817)；通过文件配置资源信息的用户请参考[通过文件配置资源信息场景](#section6131855154814)。

**通过环境变量配置资源信息场景<a name="section598118132817"></a>**

>[!NOTE] 
>此场景下，用户需已创建[hccl.json](../../api/hccl.json_file_description.md)文件的具体挂载路径才能执行以下操作，详细操作步骤请参见[步骤4](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)。

1. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    **表 1**  操作参考

    <a name="table9830101615287"></a>
    <table><thead align="left"><tr id="row1183115167289"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p1883131617285"><a name="p1883131617285"></a><a name="p1883131617285"></a>特性名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p118311416122815"><a name="p118311416122815"></a><a name="p118311416122815"></a>操作示例</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1383111642810"><td class="cellrowborder" rowspan="1" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p183111662814"><a name="p183111662814"></a><a name="p183111662814"></a>整卡调度</p>
    <p id="p6831131682816"><a name="p6831131682816"></a><a name="p6831131682816"></a></p>
    <p id="p4831141617286"><a name="p4831141617286"></a><a name="p4831141617286"></a></p>
    <p id="p1783141615287"><a name="p1783141615287"></a><a name="p1783141615287"></a></p>
    <p id="p53986186431"><a name="p53986186431"></a><a name="p53986186431"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1083151622813"><a name="p1083151622813"></a><a name="p1083151622813"></a></p><p id="p128321216132812"><a name="p128321216132812"></a><a name="p128321216132812"></a>若需要使用<span id="ph1183219160285"><a name="ph1183219160285"></a><a name="ph1183219160285"></a>PyTorch</span>或<span id="ph8832121614281"><a name="ph8832121614281"></a><a name="ph8832121614281"></a>MindSpore</span>框架支持的交换机亲和性调度，配置示例请参见<a href="#li583911163280">配置交换机亲和性调度参考示例</a>。</p>
    </td>
    </tr>
    <tr id="row108334168282"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p2833141612811"><a name="p2833141612811"></a><a name="p2833141612811"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p18339161282"><a name="p18339161282"></a><a name="p18339161282"></a><a href="#li1086213163289">Atlas 900 A3 SuperPoD 超节点上创建单机训练任务</a></p>
    </td>
    </tr>
    <tr id="row524620253494"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1324682574913"><a name="p1324682574913"></a><a name="p1324682574913"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1924622524915"><a name="p1924622524915"></a><a name="p1924622524915"></a><a href="#li164321720423">在Atlas&nbsp;800T&nbsp;A2&nbsp;训练服务器上创建训练任务（Scheduler挂载芯片的方式）</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    - <a name="li583911163280"></a>使用**整卡调度**特性，参考本配置。PyTorch和MindSpore框架新增了使用交换机亲和性调度的功能，该功能支持大模型任务和普通任务。以pytorch\_standalone\_acjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用1个芯片，修改示例如下。

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-pytorch
          labels:
            framework: pytorch   # 镜像名称
            tor-affinity: "normal-schema" # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
        spec:
          schedulerName: volcano  # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 1    # 任务总副本数
              queue: default    # 任务所属队列
          successPolicy: AllWorkers   # 任务成功的前提
          replicaSpecs:
            Master:
              replicas: 1      # 任务副本数
              restartPolicy: Never
              template:
                spec:
                  nodeSelector:
                    host-arch: huawei-arm               # 可选值，根据实际情况填写
                    accelerator-type: module         # 节点类型
                  containers:
                  - name: ascend                       # 必须为ascend，不能修改
                  image: PyTorch-test:latest       # 镜像名称
        ...
                  env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                    ports:                          #分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 1    # 任务申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 1   # 与limits取值一致
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

        >[!NOTE] 
        >PyTorch、MindSpore框架中对应的Chief、Master、Scheduler的“replicas”字段不能超过1。单机任务时，PyTorch框架不需要Worker。单卡任务时，MindSpore框架不需要Scheduler。

    - <a name="li1086213163289"></a>使用**整卡调度**特性，参考本配置。以pytorch\_standalone\_acjob\_super\_pod.yaml为例，在一台Atlas 900 A3 SuperPoD 超节点上创建**单机训练**任务，修改示例如下。

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-pytorch
          labels:
            framework: pytorch    # 框架类型
            ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
            podgroup-sched-enable: "true"  # 仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。取值为字符串"true"时，表示开启批量调度功能；取值为其他字符串时，表示批量调度功能不生效，使用普通调度。若不配置该参数，表示批量调度功能不生效，使用普通调度
          annotations:
            sp-block: "16"  # 需要和申请的芯片数量一致
        spec:
          schedulerName: volcano  # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 1     # 任务总副本数
              queue: default  # 任务所属队列
          successPolicy: AllWorkers     # 任务成功的前提
          replicaSpecs:
            Master:
              replicas: 1   # 任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b
                spec:
                  nodeSelector:
                    host-arch: huawei-arm      # 可选值，根据实际情况填写
                    accelerator-type: module-a3-16-super-pod    # 节点类型
                  containers:
                  - name: ascend  # 必须为ascend，不能修改
                    image: pytorch-test:latest      # 训练基础镜像
                    imagePullPolicy: IfNotPresent
                    env:
        ...
                      - name: ASCEND_VISIBLE_DEVICES     # Ascend Docker Runtime会使用该字段
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.annotations['huawei.com/Ascend910']               
        ...
                    ports:                     # 分布式训练集合通信端口   
                      - containerPort: 2222         # determined by user
                        name: ascendjob-port        # do not modify
                    resources:
                      limits:
                        huawei.com/Ascend910: 16   # 任务申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 16   # 与limits取值一致
        ...
        ```

        修改完成后执行[步骤2](#li118885168281)，配置YAML的其他字段。

    - <a name="li164321720423"></a>使用整卡调度特性，参考本配置。以mindspore\_multinodes\_acjob\_\{xxx\}b.yaml为例，在一台Atlas 800T A2 训练服务器上以Scheduler挂载芯片的方式执行2\*8卡训练任务，修改示例如下。

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-mindspore
          labels:
            framework: mindspore     # 训练框架名称
            ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
        spec:
          schedulerName: volcano    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            schedulingPolicy:      # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
              minAvailable: 2  #任务总副本数
              queue: default     # 任务所属队列
          successPolicy: AllWorkers  #任务成功的前提
          replicaSpecs:
            Scheduler:
              replicas: 1   # 任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  hostNetwork: true    # 可选值，根据实际情况填写，true支持hostIP创建Pod，false不支持hostIP创建Pod
                  affinity:                                         # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: job-name
                                operator: In
                                values:
                                  - default-test-mindspore         # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                  nodeSelector:
                    host-arch: huawei-arm              # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8   # 节点类型
                  containers:
                  - name: ascend                                     # 必须为ascend，不能修改
                    image: mindspore-test:latest  #镜像名称
                    imagePullPolicy: IfNotPresent
        ...
                    env:                                    
                      - name: HCCL_IF_IP                    # 可选值，根据实际情况填写
                        valueFrom:                          # 若hostNetwork配置为true，需要同步配置HCCL_IF_IP环境变量
                          fieldRef:                         # 若hostNetwork未配置或配置为false，不可配置HCCL_IF_IP环境变量
                            fieldPath: status.hostIP        # 
        ...            
                    ports:                          # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 8 # 申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8 #与limits取值一致
                    volumeMounts:
        ...            
                  volumes:
        ...            
            Worker:
              replicas: 1   #任务副本数
              restartPolicy: Never
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b   # 标识产品类型
                spec:
                  hostNetwork: true    # 可选值，根据实际情况填写，true支持hostIP创建Pod，false不支持hostIP创建Pod
                  affinity:            # 本段配置表示分布式任务的Pod调度到不同节点
                    podAntiAffinity:
                      requiredDuringSchedulingIgnoredDuringExecution:
                        - labelSelector:
                            matchExpressions:
                              - key: job-name
                                operator: In
                                values:
                                  - default-test-mindspore        # 需要和上面的任务名一致
                          topologyKey: kubernetes.io/hostname
                  nodeSelector:
                    host-arch: huawei-arm              # 可选值，根据实际情况填写
                    accelerator-type: module-{xxx}b-8  # 节点类型
                  containers:
                  - name: ascend                            # 必须为ascend，不能修改
        ...
                    env:                                    
                      - name: HCCL_IF_IP                    # 可选值，根据实际情况填写
                        valueFrom:                          # 若hostNetwork配置为true，需要同步配置HCCL_IF_IP环境变量
                          fieldRef:                         # 若hostNetwork未配置或配置为false，不可配置HCCL_IF_IP环境变量
                            fieldPath: status.hostIP        # 
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime会使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                    ports:                          # 分布式训练集合通信端口
                      - containerPort: 2222         
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 8  # 申请的芯片数量
                      requests:
                        huawei.com/Ascend910: 8  #与limits取值一致
                    volumeMounts:
        ...
                  volumes:
        ...
        ```

    >[!NOTE] 
    >整卡调度或静态vNPU调度特性配置YAML的操作只在步骤1中有区别，整卡调度和静态vNPU调度特性在步骤1之后的操作相同。

2. <a name="li118885168281"></a>若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m            
                  memory: 100Gi      
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

3. <a name="li0303"></a>修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```Yaml
              volumeMounts:
              - name: ascend-server-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ```

4. 如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及PyTorch、MindSpore框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
    - **PyTorch命令参数**

        ```Yaml
        command:
          - /bin/bash
          - -c
        args: ["cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code /job/output main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --epochs=90 --batch-size=512"]
        ...
        ```

    - **MindSpore命令参数**

        ```Yaml
        command:
          - /bin/bash
          - -c
        args: ["cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code/ /job/code/output train.py  --data_path=/job/data/resnet50/imagenet/train --config=/job/code/config/resnet50_imagenet2012_config.yaml"]
        ...
        ```

        >[!NOTE] 
        >以PyTorch命令参数为例。
        >- /job/code/：[步骤3](#li0303)中用户自定义的容器中训练脚本路径。
        >- /job/output/：[步骤3](#li0303)中用户自定义的容器中训练输出路径。
        >- main.py：启动训练脚本路径。

5. YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

    <pre codetype="yaml">
    ...
              volumeMounts:
              - name: ascend-server-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ...
               # 可选，使用组件为训练任务生成RankTable文件，需要新增以下加粗字段，设置容器中hccl.json文件保存路径。该路径不可修改。
              <strong>- name: ranktable</strong>                                 
                <strong>mountPath: /user/serverid/devindex/config</strong>
    ...
            volumes:
    ...
            - name: code
              nfs:
                server: 127.0.0.1        # NFS服务器IP地址
                path: "xxxxxx"           # 配置训练脚本路径
            - name: data
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 配置训练集路径
            - name: output
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
             # 可选，使用组件为PyTorch和MindSpore框架生成RankTable文件，需要新增以下加粗字段，设置hccl.json文件保存路径
            <strong>- name: ranktable           # 请勿修改此参数的默认值，Ascend Operator会用于检查是否开启文件挂载hccl.json</strong>
              <strong>hostPath:                 #请使用hostpath挂载或NFS挂载</strong>
                <strong>path: /user/mindx-dl/ranktable/default.default-test-pytorch   # 共享存储或者本地存储路径，/user/mindx-dl/ranktable/为前缀路径，必须和Ascend Operator挂载的Ranktable根目录保持一致。default.default-test-pytorch为后缀路径，建议改为:namespace.job-name</strong></pre>

**通过文件配置资源信息场景<a name="section6131855154814"></a>**

1. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    **表 2**  操作参考

    <a name="table353271710226"></a>
    <table><thead align="left"><tr id="row8532181713223"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p8532151719228"><a name="p8532151719228"></a><a name="p8532151719228"></a>特性名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p853216179225"><a name="p853216179225"></a><a name="p853216179225"></a>操作示例</p>
    </th>
    </tr>
    </thead>
    <tbody>
    <tr id="row173101655202217"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1366503795516"><a name="p1366503795516"></a><a name="p1366503795516"></a>整卡调度</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p918205224718"><a name="p918205224718"></a><a name="p918205224718"></a>若需要使用<span id="ph181814524472"><a name="ph181814524472"></a><a name="ph181814524472"></a>PyTorch</span>或<span id="ph318155294712"><a name="ph318155294712"></a><a name="ph318155294712"></a>MindSpore</span>框架支持的交换机亲和性调度，配置示例请参见<a href="#li1460553372">配置交换机亲和性调度参考示例</a>。</p>
    </td>
    </tr>
    </tbody>
    </table>

    - <a name="li1460553372"></a>使用**整卡调度**特性，参考本配置。PyTorch和MindSpore框架新增了使用交换机亲和性调度的功能，该功能支持大模型任务和普通任务。以a800\_pytorch\_vcjob.yaml为例，在一台Atlas 800T A2 训练服务器节点创建**分布式训练**任务，任务使用1\*8个芯片，修改示例如下。

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
          namespace: vcjob                      
          labels:
            ring-controller.atlas: ascend-{xxx}b   # 产品类型
        data:
          hccl.json: |
            {
                "status":"initializing"
            }
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改，必须使用Volcano的API
        kind: Job                               # 目前只支持Job类型
        metadata:
        ...                 
          labels:
            ring-controller.atlas: ascend-{xxx}b   # 必须与ConfigMap中的标签保持一致，不可修改
            fault-scheduling: "force"
            tor-affinity: "normal-schema"      # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
        spec:
          minAvailable: 1                       # 此处建议与下面的为节点个数保持一致
          schedulerName: volcano                # 使用Volcano进行调度
        ...
          tasks:
          - name: "default-test"
            replicas: 1                              # 此处为节点个数
            template:
              metadata:
                labels:
                  app: pytorch
                  ring-controller.atlas: ascend-{xxx}b  # 必须与ConfigMap中的标签一致，不可修改
              spec:
                affinity:                           # 本段配置表示分布式任务的Pod调度到不同节点
                  podAntiAffinity:
                    requiredDuringSchedulingIgnoredDuringExecution:
                      - labelSelector:
                          matchExpressions:
                            - key: volcano.sh/job-name
                              operator: In
                              values:
                                - mindx-dls-test
                        topologyKey: kubernetes.io/hostname
                hostNetwork: true
                containers:
                - image: torch:b030               # 训练框架镜像，根据实际情况修改
                  - name: XDL_IP                   # 本段固定不变
                    valueFrom:
                      fieldRef:
                        fieldPath: status.hostIP
                  - name: POD_UID
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.uid
                  - name: framework
                    value: "PyTorch"
        ...
                  - name: ASCEND_VISIBLE_DEVICES
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources.requests保持一致
        ...
                  resources:
                    requests:
                      huawei.com/Ascend910: 8                # 每台Atlas 800T A2 训练服务器芯片数量最多为8
                    limits:
                      huawei.com/Ascend910: 8                 # 每台Atlas 800T A2 训练服务器芯片数量最多为8
         ...
                nodeSelector:
                  host-arch: huawei-x86       # 可选值，根据实际情况填写
                  accelerator-type: module-{xxx}b-8    #调度到Atlas 800T A2 训练服务器节点
        ...
        ```

        >[!NOTE] 
        >其余示例可参考[表5](#table62591594016)和[表6](#table21811158146)，以及YAML对应的参数说明[表2 YAML参数说明](#zh-cn_topic_0000001609074269_table1565872494511)进行适配修改。修改完成后执行[步骤2](#li832632419711)，继续配置YAML的其他字段。

    >[!NOTE] 
    >整卡调度或静态vNPU调度特性配置YAML的操作只在步骤1中有区别，整卡调度和静态vNPU调度特性在步骤1之后的操作相同。

2. <a name="li832632419711"></a>若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m            
                  memory: 100Gi      
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

3. <a name="li112747151117"></a>修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```Yaml
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ```

4. 如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
    - **PyTorch命令参数**

        ```Yaml
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --world-size=1 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024;"
        ...
        ```

    - **MindSpore命令参数**

        ```Yaml
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ train.py  --config_path=/job/code/config/resnet50_imagenet2012_config.yaml --output_dir=/job/output --run_distribute=True --device_num=8 --data_path=/job/data/imagenet/train"
        ...
        ```

        >[!NOTE] 
        >以PyTorch命令参数为例。
        >- /job/code/：[步骤3](#li112747151117)中用户自定义的容器中训练脚本路径。
        >- /job/output/：[步骤3](#li112747151117)中用户自定义的容器中训练输出路径。
        >- main.py：启动训练脚本路径。

5. YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

    ```Yaml
    ...
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ...
            volumes:
    ...
            - name: code
              nfs:
                server: 127.0.0.1        # NFS服务器IP地址
                path: "xxxxxx"           # 配置训练脚本路径
            - name: data
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 配置训练集路径
            - name: output
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
    ```

### 下发任务<a name="ZH-CN_TOPIC_0000002511427065"></a>

**操作步骤<a name="zh-cn_topic_0000001608595413_section15243115731317"></a>**

1. 示例a800\_tensorflow\_vcjob.yaml中，任务部署在vcjob命名空间下，因此需要在管理节点执行以下命令，为训练任务创建命名空间。如果任务创建到非默认的命名空间，则需要根据实际情况创建命名空间。

    ```shell
    kubectl create namespace vcjob
    ```

2. 在管理节点示例YAML所在路径，执行以下命令，使用YAML下发训练任务。

    ```shell
    kubectl apply -f XXX.yaml
    ```

    >[!NOTE] 
    >如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f <i>XXX.yaml</i>命令删除原任务，再重新下发任务。

    - 通过环境变量配置资源信息场景的示例如下：

        ```shell
        kubectl apply -f tensorflow_standalone_acjob.yaml
        ```

        回显示例如下：

        ```ColdFusion
        ascendjob.mindxdl.gitee.com/default-tensorflow-test created
        ```

    - 通过文件配置资源信息场景的示例如下：

        ```shell
        kubectl apply -f a800_tensorflow_vcjob.yaml
        ```

        回显示例如下：

        ```ColdFusion
        configmap/rings-config-mindx-dls-test created
        job.batch.volcano.sh/mindx-dls-test created
        ```

>[!NOTE]
>
>- 若下发训练任务后，任务一直处于Pending状态，可以参见[训练任务处于Pending状态，原因：nodes are unavailable](../../faq.md#训练任务处于pending状态原因nodes-are-unavailable)或者[资源不足时，任务处于Pending状态](../../faq.md#资源不足时任务处于pending状态)章节进行处理。
>- 若成功启动训练任务后，发现训练任务容器内部hccl.json文件处于initializing状态，可以参见[hccl.json文件没有生成](../../faq.md#hccljson文件没有生成)章节进行处理。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002479387130"></a>

**操作步骤<a name="zh-cn_topic_0000001558675462_section15243115731317"></a>**

1. 在管理节点查看任务Pod的状态，需要保证Pod状态为Running。

    执行以下命令，查看Pod运行情况。

    ```shell
    kubectl get pod --all-namespaces -o wide
    ```

    - 单机单芯片训练任务回显示例。

        ```ColdFusion
        NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE           NOMINATED NODE   READINESS GATES
        ...
        vcjob            mindx-dls-test-default-test-0             1/1     Running            0          4m      192.168.243.198   ubuntu         <none>           <none>
        ...
        ```

    - 两个训练节点，执行2\*8芯片分布式训练任务回显示例。

        ```ColdFusion
        NAMESPACE        NAME                                       READY   STATUS              RESTARTS   AGE     IP                NODE           NOMINATED NODE   READINESS GATES
        ...
        vcjob            mindx-dls-test-default-test-0             1/1     Running            0          3m      192.168.243.198   ubuntu         <none>           <none>
        vcjob            mindx-dls-test-default-test-1             1/1     Running            0          3m      192.168.243.199   ubuntu         <none>           <none>
        ...
        ```

2. 查看计算节点的NPU分配情况，在管理节点执行以下命令查看。

    ```shell
    kubectl describe nodes {任务运行节点的节点名}
    ```

    - 使用**整卡调度**特性，单机单芯片训练任务回显示例。

        ```ColdFusion
        Name:               ubuntu
        Roles:              master,worker
        Labels:             accelerator=huawei-Ascend910
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests        Limits
          --------              --------        ------
          cpu                   37250m (19%)    37500m (19%)
          memory                117536Mi (15%)  119236Mi (15%)
          ephemeral-storage     0 (0%)          0 (0%)
          huawei.com/Ascend910  1               1
        Events:                 <none>
        ```

        >[!NOTE] 
        >**Allocated resources**的字段huawei.com/Ascend910的值为1，表明训练使用了一个NPU。

    - 使用**静态vNPU调度**特性，单机单芯片训练任务回显示例。

        ```ColdFusion
        Name:               ubuntu
        Roles:              master,worker
        Labels:             accelerator=huawei-Ascend910
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests        Limits
          --------              --------        ------
          cpu                   37250m (19%)    37500m (19%)
          memory                117536Mi (15%)  119236Mi (15%)
          ephemeral-storage     0 (0%)          0 (0%)
          huawei.com/Ascend910-2c  1               1
        Events:                 <none>
        ```

        >[!NOTE] 
        >**Allocated resources**的字段**huawei.com/Ascend910-2c**的值为1，表明训练使用了一个包含了2个AI Core的vNPU。

    - 两个训练节点，执行2\*8芯片分布式训练任务，查看其中一个节点示例。**静态vNPU调度**不支持分布式训练任务。

        ```ColdFusion
        Name:               ubuntu
        Roles:              master,worker
        Labels:             accelerator=huawei-Ascend910
                            beta.kubernetes.io/arch=arm64
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests        Limits
          --------              --------        ------
          cpu                   37250m (19%)    37500m (19%)
          memory                117536Mi (15%)  119236Mi (15%)
          ephemeral-storage     0 (0%)          0 (0%)
          huawei.com/Ascend910  8               8
        Events:                 <none>
        ```

        >[!NOTE] 
        >**Allocated resources**的字段huawei.com/Ascend910的值为8，表明分布式训练使用了节点上所有的NPU。

3. 查看Pod的NPU使用情况。

    本例中使用**kubectl describe pod mindx-dls-test-default-test-0 -n vcjob**命令查看运行Pod的情况。

    - 单机单芯片训练任务示例，显示如下内容表示正常。

        ```ColdFusion
        root@ubuntu:/home/test/yaml# kubectl describe pod mindx-dls-test-default-test-0 -n vcjob
        Name:         mindx-dls-test-default-test-0
        Namespace:    vcjob
        Priority:     0
        Node:         ubuntu/XXX.XXX.XXX.XXX
        Start Time:   Wed, 30 Sep 2020 15:38:22 +0800
        Labels:       app=tf
                      ring-controller.atlas=ascend-910
                      volcano.sh/job-name=mindx-dls-test
                      volcano.sh/job-namespace=vcjob
        Annotations:  ascend.kubectl.kubernetes.io/ascend-910-configuration:
                        {"pod_name":"0","server_id":"xx-xx-xx-xx","devices":[{"device_id":"3","device_ip":"192.168.20.102"}...
                      cni.projectcalico.org/podIP: 192.168.243.195/32
                      cni.projectcalico.org/podIPs: 192.168.243.195/32
                      huawei.com/Ascend910: Ascend910-3
                      huawei.com/AscendReal: Ascend910-3
                      huawei.com/kltDev: Ascend910-3
                      predicate-time: 18446744073709551615
                      scheduling.k8s.io/group-name: mindx-dls-test
                      volcano.sh/job-name: mindx-dls-test
                      volcano.sh/job-version: 0
                      volcano.sh/task-spec: default-test
        Status:       Running
        ```

    - 两个训练节点，执行2\*8芯片分布式训练任务示例，显示如下内容表示正常。

        ```ColdFusion
        root@ubuntu:/home/test/yaml# kubectl describe pod mindx-dls-test-default-test-0 -n vcjob
        Name:         mindx-dls-test-default-test-0
        Namespace:    vcjob
        Priority:     0
        Node:         ubuntu/XXX.XXX.XXX.XXX
        Start Time:   Wed, 30 Sep 2020 15:38:22 +0800
        Labels:       app=tf
                      ring-controller.atlas=ascend-910
                      volcano.sh/job-name=mindx-dls-test
                      volcano.sh/job-namespace=vcjob
        Annotations:  ascend.kubectl.kubernetes.io/ascend-910-configuration:
                        {"pod_name":"0","server_id":"xx-xx-xx-xx","devices":[{"device_id":"0","device_ip":"192.168.20.100"}...
                      cni.projectcalico.org/podIP: 192.168.243.195/32
                      cni.projectcalico.org/podIPs: 192.168.243.195/32
                      huawei.com/Ascend910: Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7
                      huawei.com/AscendReal: Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7,Ascend910-0
                      huawei.com/kltDev: Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7,Ascend910-0,Ascend910-1,Ascend910-2
                      predicate-time: 18446744073709551615
                      scheduling.k8s.io/group-name: mindx-dls-test
                      volcano.sh/job-name: mindx-dls-test
                      volcano.sh/job-version: 0
                      volcano.sh/task-spec: default-test
        Status:       Running
        ```

### 查看整卡调度或静态vNPU调度结果<a name="ZH-CN_TOPIC_0000002479387140"></a>

**PyTorch<a name="zh-cn_topic_0000001609474257_section15657195014514"></a>**

1. 在执行如下命令，查看训练结果。

    ```shell
    kubectl logs -n  <namespace> <pod-name>
    ```

    如：

    ```shell
    kubectl logs -n vcjob mindx-dls-test-default-test-0
    ```

2. 查看训练日志，如果出现如下内容表示训练成功。

    ```ColdFusion
    [gpu id: 0 ] Test: [77/85]      Time  0.117 ( 0.281)    Loss 1.073741e+01 (1.078090e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.00 (  0.12)
    [gpu id: 0 ] Test: [78/85]      Time  0.114 ( 0.279)    Loss 1.072909e+01 (1.078015e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.00 (  0.12)
    [gpu id: 0 ] Test: [79/85]      Time  0.115 ( 0.277)    Loss 1.073733e+01 (1.077953e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.20 (  0.12)
    [gpu id: 0 ] Test: [80/85]      Time  2.385 ( 0.306)    Loss 1.087646e+01 (1.078090e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.00 (  0.12)
    [gpu id: 0 ] Test: [81/85]      Time  1.139 ( 0.318)    Loss 1.075754e+01 (1.078058e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.39 (  0.12)
    [gpu id: 0 ] Test: [82/85]      Time  0.115 ( 0.315)    Loss 1.068419e+01 (1.077925e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.20 (  0.13)
    [gpu id: 0 ] Test: [83/85]      Time  0.129 ( 0.313)    Loss 1.075079e+01 (1.077887e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.20 (  0.13)
    [gpu id: 0 ] Test: [84/85]      Time  0.134 ( 0.310)    Loss 1.093459e+01 (1.078095e+01)        Acc@1   0.00 (  0.02)   Acc@5   0.39 (  0.13)
    [gpu id: 0 ] [AVG-ACC] * Acc@1 0.016 Acc@5 0.130
    validate acc1 tensor(0.0156, device='npu:0')
    Complete 90 epoch training, take time:1.05h
    ...
    ```

3. 进入模型输出目录，查看生成的模型文件。

    ```ColdFusion
    drwxrwx--- 2 root root      4096 Mar  4 19:28 ./
    drwxrwx--- 4 root root      4096 Mar  4 19:28 ../
    -rw-rw---- 1 root root 102489869 Mar  4 19:28 checkpoint_npu0model_best.pth.tar
    -rw-rw---- 1 root root 102489869 Mar  4 19:28 checkpoint_npu0.pth.tar
    ...
    ```

    可以参考ModelZoo上，PyTorch框架的[ResNet-50](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_for_PyTorch#%E6%A8%A1%E5%9E%8B%E6%8E%A8%E7%90%86)模型中的“模型推理”章节，对生成的模型文件进行模型转换处理。

**MindSpore<a name="zh-cn_topic_0000001609474257_section1533335719514"></a>**

1. 在执行如下命令，查看训练结果。

    ```shell
    kubectl logs -n  <namespace> <pod-name>
    ```

    如：

    ```shell
    kubectl logs -n vcjob mindx-dls-test-default-test-0
    ```

2. 查看训练日志，如果出现如下内容表示训练成功。

    ```ColdFusion
    ...
    2023-06-09 17:55:04,837:INFO:epoch: [70/90] loss: 1.541062, epoch time: 7.563 s, per step time: 157.554 ms
    2023-06-09 17:55:10,540:INFO:epoch: [71/90] loss: 1.544771, epoch time: 5.702 s, per step time: 118.796 ms
    2023-06-09 17:55:16,347:INFO:epoch: [72/90] loss: 1.506525, epoch time: 5.807 s, per step time: 120.979 ms
    2023-06-09 17:55:24,904:INFO:epoch: [73/90] loss: 1.519342, epoch time: 8.556 s, per step time: 178.260 ms
    2023-06-09 17:55:29,887:INFO:epoch: [74/90] loss: 1.387423, epoch time: 4.982 s, per step time: 103.783 ms
    2023-06-09 17:55:39,785:INFO:epoch: [75/90] loss: 1.440862, epoch time: 9.897 s, per step time: 206.194 ms
    2023-06-09 17:55:48,780:INFO:epoch: [76/90] loss: 1.431275, epoch time: 8.995 s, per step time: 187.399 ms
    2023-06-09 17:55:55,764:INFO:epoch: [77/90] loss: 1.411003, epoch time: 6.984 s, per step time: 145.492 ms
    2023-06-09 17:56:03,962:INFO:epoch: [78/90] loss: 1.457689, epoch time: 8.198 s, per step time: 170.783 ms
    2023-06-09 17:56:11,517:INFO:epoch: [79/90] loss: 1.410896, epoch time: 7.554 s, per step time: 157.372 ms
    2023-06-09 17:56:16,643:INFO:epoch: [80/90] loss: 1.517990, epoch time: 5.126 s, per step time: 106.789 ms
    2023-06-09 17:56:23,364:INFO:epoch: [81/90] loss: 1.342399, epoch time: 6.720 s, per step time: 140.005 ms
    2023-06-09 17:56:31,835:INFO:epoch: [82/90] loss: 1.352396, epoch time: 8.471 s, per step time: 176.470 ms
    2023-06-09 17:56:36,971:INFO:epoch: [83/90] loss: 1.358075, epoch time: 5.135 s, per step time: 106.984 ms
    2023-06-09 17:56:44,259:INFO:epoch: [84/90] loss: 1.400720, epoch time: 7.288 s, per step time: 151.838 ms
    2023-06-09 17:56:52,868:INFO:epoch: [85/90] loss: 1.371813, epoch time: 8.608 s, per step time: 179.339 ms
    2023-06-09 17:56:57,613:INFO:epoch: [86/90] loss: 1.303416, epoch time: 4.745 s, per step time: 98.858 ms
    2023-06-09 17:57:04,177:INFO:epoch: [87/90] loss: 1.290425, epoch time: 6.564 s, per step time: 136.744 ms
    2023-06-09 17:57:11,797:INFO:epoch: [88/90] loss: 1.298486, epoch time: 7.619 s, per step time: 158.738 ms
    2023-06-09 17:57:16,807:INFO:epoch: [89/90] loss: 1.297104, epoch time: 5.009 s, per step time: 104.363 ms
    2023-06-09 17:57:25,568:INFO:epoch: [90/90] loss: 1.401816, epoch time: 8.759 s, per step time: 182.486 ms
    ```

3. 进入模型输出目录，查看生成的模型文件。

    ```ColdFusion
    drwx------  2 root root      4096 Dec 21 15:35 ./
    drwxrwxrwx 10 root root      4096 Dec 21 15:26 ../
    -r--------  1 root root 188546464 Dec 21 15:31 resnet50-45_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:31 resnet50-50_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:32 resnet50-55_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:32 resnet50-60_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:33 resnet50-65_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:33 resnet50-70_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:33 resnet50-75_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:34 resnet50-80_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:34 resnet50-85_234.ckpt
    -r--------  1 root root 188546464 Dec 21 15:35 resnet50-90_234.ckpt
    -rw-------  1 root root    769071 Dec 21 15:28 resnet50-graph.meta
    ```

### 删除任务<a name="ZH-CN_TOPIC_0000002479227168"></a>

**操作步骤<a name="zh-cn_topic_0000001609474265_section1595872772813"></a>**

在示例YAML所在路径下，执行以下命令，删除对应的训练任务。

```shell
kubectl delete -f XXX.yaml
```

- 通过环境变量配置资源信息场景的示例如下：

    ```shell
    kubectl delete -f tensorflow_standalone_acjob.yaml
    ```

    回显示例如下：

    ```ColdFusion
    ascendjob.mindxdl.gitee.com "default-test-tensorflow" deleted
    ```

- 通过文件配置资源信息场景的示例如下：

    ```shell
    kubectl delete -f a800_tensorflow_vcjob.yaml
    ```

    回显示例如下：

    ```ColdFusion
    configmap "rings-config-mindx-dls-test" deleted
    job.batch.volcano.sh "mindx-dls-test" deleted
    ```

>[!NOTE] 
>若删除训练任务后，Pod一直处于Terminating状态，可以参见[手动删除vcjob后Pod一直处于Terminating状态](../../faq.md#手动删除vcjob后pod一直处于terminating状态)章节进行处理。

## 通过命令行使用（其他调度器）<a name="ZH-CN_TOPIC_0000002511427069"></a>

通过命令行使用（其他调度器）和通过命令行使用（Volcano）使用流程一致，只有任务YAML有所不同，用户可以准备好相应YAML后参考[通过命令行使用（Volcano）](#通过命令行使用volcano)章节使用。

**操作步骤<a name="section1780564613381"></a>**

1. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    集群调度并未专门提供使用其他调度器的YAML示例，用户可以获取使用Volcano的YAML示例并做如下修改即可使用。

    以tensorflow\_standalone\_acjob.yaml为例，在一台Atlas 800T A2 训练服务器节点创建**单机训练**任务，执行1\*8芯片训练任务，修改示例如下。

    <pre codetype="yaml">
    apiVersion: mindxdl.gitee.com/v1
    kind: AscendJob
    metadata:
      name: default-test-tensorflow
      labels:
        framework: tensorflow
        ring-controller.atlas: ascend-{xxx}b   
    spec:
      schedulerName: volcano        # 使用其他调度器时，删除该字段
      runPolicy:                    # 使用其他调度器时，删除该字段
        schedulingPolicy:           
          minAvailable: 1
          queue: default
      successPolicy: AllWorkers
      replicaSpecs:
        Chief:
          replicas: 1
          restartPolicy: Never
          template:
            metadata:
              labels:
                ring-controller.atlas: ascend-{xxx}b   
            spec:
              nodeSelector:
                host-arch: huawei-arm
                accelerator-type: module-{xxx}b-8
              containers:
              - name: ascend                    
    ...
                env:
    ...
             # 使用其他调度器暂不支持ASCEND_VISIBLE_DEVICES相关字段，需要删除以下加粗字段
              <strong>- name: ASCEND_VISIBLE_DEVICES</strong>                       
                <strong>valueFrom:</strong>
                  <strong>fieldRef:</strong>
                    <strong>fieldPath: metadata.annotations['huawei.com/Ascend910']</strong>               
    ...
              resources: 
                limits:
                 huawei.com/Ascend910: 8 
                requests:
                  huawei.com/Ascend910: 8
              volumeMounts:</pre>

2. 若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m            
                  memory: 100Gi      
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

3. <a name="li1127471511178"></a>修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```Yaml
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ```

4. 如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及框架模型文件）、启动脚本相对代码目录的路径。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。
    - **PyTorch命令参数**

        ```Yaml
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --world-size=1 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024;"
        ...
        ```

    - **MindSpore命令参数**

        ```Yaml
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ train.py  --config_path=/job/code/config/resnet50_imagenet2012_config.yaml --output_dir=/job/output --run_distribute=True --device_num=8 --data_path=/job/data/imagenet/train"
        ...
        ```

    >[!NOTE] 
    >以PyTorch命令参数为例。
    >- /job/code/：[步骤3](#li1127471511178)中用户自定义的容器中训练脚本路径。
    >- /job/output/：[步骤3](#li1127471511178)中用户自定义的容器中训练输出路径。
    >- main.py：启动训练脚本路径。

5. YAML为使用NFS场景，需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

    ```Yaml
    ...
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ...
            volumes:
    ...
            - name: code
              nfs:
                server: 127.0.0.1        # NFS服务器IP地址
                path: "xxxxxx"           # 配置训练脚本路径
            - name: data
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 配置训练集路径
            - name: output
              nfs:
                server: 127.0.0.1
                path: "xxxxxx"           # 设置脚本相关配置模型保存路径
    ...
    ```

## 集成后使用<a name="ZH-CN_TOPIC_0000002511347081"></a>

本章节以**整卡调度**特性为例，介绍如何将整卡调度特性集成在AI平台上的关键操作步骤。在下发训练任务时，平台需要实现获取认证文件，创建客户端，创建Job对象，创建命名空间和调用接口下发训练任务等，将整卡调度特性提供的示例YAML转换成K8s提供的Go编程语言的API对象。

**集成前说明<a name="section16646104012516"></a>**

集成操作中会涉及到很多接口，请用户根据实际情况去相关官网了解接口的详细信息，本文档不再进行二次说明。

- K8s相关接口请根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)了解相关内容。
- Ascend Job可参考[表1](#table159746356276)中的参数说明，了解相关内容。
- Volcano Job相关接口可参见《云容器实例 API参考》中“[创建Volcano Job](https://support.huaweicloud.com/api-cci/createBatchVolcanoShV1alpha1NamespacedJob.html)”章节了解相关内容。
- 下文的\{_xxx_\}即取“910”字符作为芯片型号数值。

**集成操作<a name="section9868584469"></a>**

1. 获取K8s认证文件。

    用户根据实际情况选择合适的[集群认证方式](https://kubernetes.io/zh-cn/docs/concepts/security/controlling-access/)，创建相应的集群配置。使用ServiceAccount创建集群配置（InCluster模式）示例代码如下。

    ```go
           // 使用ServiceAccount创建集群配置（InCluster模式）
           if config, err = rest.InClusterConfig(); err != nil {
                  // 使用KubeConfig文件创建集群配置
                  if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
                         panic(err.Error())
                  }
           }
    ```

2. 创建客户端Clientset。

    ```go
    Client, err := NewForConfig(cfg)
    ```

    >[!NOTE] 
    >NewForConfig\(cfg\)的函数原型为**NewForConfig\(c \*rest.Config\)\(\*Clientset, error\)。**
    >参数说明如下：
    >- **\*rest.Config**：客户端配置文件，由K8s提供的接口生成；包括cluster host、证书等信息。
    >- **\*Clientset**：Client集合，包括AscendJob client（或VolcanoJob client）和discovery client。
    >- **error**：错误信息。

3. 创建Job对象。通过环境变量配置资源信息的用户需要创建Ascend Job对象；通过文件配置资源信息的用户需要创建Volcano Job对象。

    >[!NOTE] 
    >在进行本步骤操作之前，建议用户详细阅读[准备任务YAML](#准备任务yaml)章节，了解示例YAML实现逻辑和关键字段说明，可以更好地帮助用户进行接下来的操作。

    - 创建Ascend Job对象。

        创建Ascend Job对象，初始化Ascend Job相关字段，示例如下。

        ```go
        import (
           v1 "ascend-operator-apis/pkg/apis/batch/v1"
           "ascend-operator-apis/pkg/client/clientset/versioned"
           commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
           corev1 "k8s.io/api/core/v1"
           "k8s.io/apimachinery/pkg/api/resource"
           metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
           "k8s.io/client-go/tools/clientcmd"
        )
        
        func initAcJob() v1.AscendJob {
           job := newAcJob().
              initName("default-test-pytorch"). // 初始化任务名
              initNameSpace("default").         // 初始化命名空间
              initLabels(map[string]string{     // 初始化任务标签
                 "ring-controller.atlas": "ascend-{xxx}b",   // 标识任务使用的芯片的产品类型
                 "framework":             "pytorch",       // 使用的训练框架名
                 "tor-affinity":          "normal-schema", // 是否使用交换亲和性调度。value为large-model-schema，表示使用大模型调度模式；value为normal-schema，使用普通任务调度模式；value为null，表示关闭交换机亲和性调度
                 "podgroup-sched-enable": "true",  // 仅在集群使用openFuyao定制Kubernetes和volcano-ext组件场景下配置。取值为字符串"true"时，表示开启批量调度功能；取值为其他字符串时，表示批量调度功能不生效，使用普通调度。若不配置该参数，表示批量调度功能不生效，使用普通调度
              }).
              initSchedulerName("volcano"). // 初始化调度器名
              initRunPolicy(&maNum).        // 初始化RunPolicy
              initSuccessPolicy().
              addReplicaSpecs("Master", newReplica(). // 初始化Master副本
                             initRcReplicas(&rcNum).                           // 初始化pod副本数
                             initRcRestartPolicy(commonv1.RestartPolicyNever). // 初始化容器重启策略
                             initRcLabels(map[string]string{                   // 初始化Master的标签
                    "ring-controller.atlas": "ascend-{xxx}b", // 标识任务使用的芯片的产品类型
                 }).                                   //
                 initRcNodeSelector(map[string]string{ // 初始化Master的NodeSelector
                    "host-arch":        "huawei-x86",     // 可选字段，用户根据实际需求进行配置
                    "accelerator-type": "module-{xxx}b-16", // 用户根据服务器类型进行配置，取值可以参考YAML参数
                             }).
                             initRcVolumes(). // 初始化挂载项
                             addRcContainers(newContainer().
                                initContainerName("ascend").                                             // 初始化容器名
                                initContainerImage("pt-arm:b120").                                       // 初始化镜像名
                                initContainerImagePullPolicy(corev1.PullIfNotPresent).                   // 初始化镜像拉取策略
                                initContainerEnv().                                                      // 初始化容器环境变量
                                initContainerCommand([]string{"/bin/bash", "-c", "bash train_start.sh ..."}).  // 初始化容器启动命令，具体参数参考示例YAML
                                initContainerArgs([]string{"/bin/bash", "-c", "bash train_start.sh ..."}).  // 初始化容器启动命令，具体参数参考示例YAML
                                initContainerPorts(2222).                                                // 初始化容器端口
                                initContainerLimits("huawei.com/Ascend910", "8").                        // 初始化任务资源，Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需要修改参数名称为huawei.com/npu
                                initContainerRequests("huawei.com/Ascend910", "8").                      // 初始化任务资源，Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需要修改参数名称为huawei.com/npu
                                initContainerVolumeMounts()).                                            // 初始化容器挂载项
                             initReplica()).
              addReplicaSpecs("Worker", newReplica(). // 初始化Worker副本
                             initRcReplicas(&rcNum).                           // 初始化pod副本数
                             initRcRestartPolicy(commonv1.RestartPolicyNever). // 初始化容器重启策略
                             initRcLabels(map[string]string{                   // 初始化Worker的标签
                    "ring-controller.atlas": "ascend-{xxx}b", // 标识任务使用的芯片的产品类型
                 }).
                 initRcAffinity("default-test-pytorch"). // 初始化Worker的反亲和性字段
                 initRcNodeSelector(map[string]string{   // 初始化Worker的NodeSelector
                    "host-arch":        "huawei-x86",     // 用户根据实际架构配置arm架构value为huawei-arm
                    "accelerator-type": "module-{xxx}b-8", // 用户根据服务器类型进行配置，value值可以参考YAML参数
                 }).
                 initRcVolumes().
                 addRcContainers(newContainer().
                    initContainerName("ascend").                                                  // 初始化容器名
                    initContainerImage("pt-arm:b120").                                            // 初始化镜像名
                    initContainerImagePullPolicy(corev1.PullIfNotPresent).                        // 初始化镜像拉取策略
                    initContainerEnv().                                                           // 初始化容器环境变量
                    initContainerCommand([]string{"/bin/bash", "-c", "bash train_start.sh ..."}). // 初始化容器启动命令，具体参数参考示例YAML
                    initContainerArgs([]string{"/bin/bash", "-c", "bash train_start.sh ..."}).    // 初始化容器启动命令，具体参数参考示例YAML
                    initContainerPorts(2222).                                                     // 初始化容器端口
                    initContainerLimits("huawei.com/Ascend910", "8").                             // 初始化任务资源，Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需要修改参数名称为huawei.com/npu
                    initContainerRequests("huawei.com/Ascend910", "8").                           // 初始化任务资源，Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需要修改参数名称为huawei.com/npu
                    initContainerVolumeMounts()).
                 initReplica())
           return v1.AscendJob(job)
        }
        
        type acJob v1.AscendJob
        type Replica commonv1.ReplicaSpec
        type container corev1.Container
        
        func (job acJob) initRunPolicy(n *int32) acJob {
           job.Spec.RunPolicy = commonv1.RunPolicy{SchedulingPolicy: &commonv1.SchedulingPolicy{MinAvailable: n, Queue: "default"}}
           return job
        }
        ...
        func (rc Replica) initRcReplicas(rs *int32) Replica {
           rc.Replicas = rs
           return rc
        }
        ...
        func (ct container) initContainerEnv() container {
           ct.Env = []corev1.EnvVar{
              {
                 Name: "XDL_IP",
                 ValueFrom: &corev1.EnvVarSource{
                    FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.hostIP"},
                 },
              },
           }
           return ct
        }
        ```

    - 创建Volcano Job对象。
        1. 初始化Volcano Job挂载的ConfigMap。初始化ConfigMap相关字段，示例如下。

            ```go
            import "k8s.io/api/core/v1"                                                 
            func newConfigMap(name string) *v1.ConfigMap {
                   cm := &v1.ConfigMap{}                           
                   cm.Name = name
                   cm.Labels = map[string]string{
                          "ring-controller.atlas": "ascend-{xxx}b",  // 标识任务使用的芯片的产品类型
                   }
                   cm.Data = map[string]string{
                          "hccl.json": `{"status": "initializing"}`,  
                   }
                   return cm
            }
            ```

        2. 初始化Volcano Job。创建Volcano Job对象，初始化Volcano Job相关字段，示例如下。

            ```go
            import (
               "k8s.io/api/core/v1"
               metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
               "k8s.io/client-go/kubernetes"
               "k8s.io/client-go/tools/clientcmd"
               "volcano.sh/apis/pkg/apis/batch/v1alpha1"
               "volcano.sh/apis/pkg/client/clientset/versioned"
            )
            
            func initJob() v1alpha1.Job {
               job := newJobBuilder().
                  initNameSpace("vcjob").       // 初始化命名空间
                  initName("mindx-dls-test").   // 初始化任务名
                  initLabels(map[string]string{ // 初始化任务标签
                     "ring-controller.atlas": "ascend-{xxx}b", // 标识任务使用的芯片的产品类型
                  }).
                  initMinAvailable(2).          // minAvailable表示运行该job所要运行的最少Pod数量。只有当job中处于running状态的Pod数量不小于minAvailable时，才认为该job运行正常
                  initSchedulerName("volcano"). // 使用Volcano作为调度，用户可根据实际情况修改
                  initPolicies([]v1alpha1.LifecyclePolicy{{Event: "PodEvicted", Action: "RestartJob"}}).
                  initPlugins(map[string][]string{"ssh": {}, "env": {}, "svc": {}}). // 初始化调度插件
                  initMaxRetry(3).                                                   // maxRetry表示该job可以进行的最大重启次数
                  initQueue("default" ).                                              // queue表示该job所属的队列
                  addTask(v1alpha1.TaskSpec(newTaskSpec().
                     initTaskName("default-test").       // 初始化task名
                     initTaskReplicas(2).                // 初始化task副本数
                     initTaskAffinity("mindx-dls-test"). // 初始化任务反亲和性字段，输入值与任务名相同
                     initTaskLabels(map[string]string{
                        "app":                   "mindspore",   // 固定字段
                        "ring-controller.atlas": "ascend-{xxx}b", //  标识任务使用的芯片的产品类型
                     }).
                     initTaskNodeSelector(map[string]string{
                        "host-arch":        "huawei-x86",     // 可选字段，用户根据实际需求配置
                        "accelerator-type": "module-{xxx}b-8", // 用户根据服务器类型进行配置
                     }).
                     initTaskVolumes(). // 初始化挂载项
                     addTaskContainers(v1.Container(newContainer().
                        initContainerName("mindspore").                                             // 初始化容器名
                        initContainerImage("ms-arm:b120").                                          // 初始化镜像名
                        initContainerImagePullPolicy("IfNotPresent").                               // 初始化镜像拉取策略
                        initContainerLimits("huawei.com/Ascend910", "8").                           // 初始化任务资源，Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需要修改参数名称为huawei.com/npu
                        initContainerRequests("huawei.com/Ascend910", "8").                         // 初始化任务资源，Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需要修改参数名称为huawei.com/npu
                        initContainerVolumeMounts().                                                // 初始化容器挂载项
                        initContainerEnv("MindSpore").                                              // 初始化容器环境变量
                        initContainerCommand([]string{"/bin/bash", "-c", "bash train_start.sh ..."}))))) // 初始化容器启动命令，具体参数参考示例YAML
               return v1alpha1.Job(job)
            }
            
            type vcJob v1alpha1.Job
            type vcTask v1alpha1.TaskSpec
            type container v1.Container
            
            // 初始化任务名
            func (job *vcJob) initName(n string) *vcJob {
               job.Name = n
               return job
            }
            ...
            // 初始化task名
            func (task vcTask) initTaskName(tn string) vcTask {
               task.Name = tn
               return task
            }
            ...
            // 初始化容器环境变量
            func (ct container) initContainerEnv(framework string) container {
               ct.Env = []v1.EnvVar{
                  {
                     Name: "mindx-dls-test", // 任务名称
                     ValueFrom: &v1.EnvVarSource{
                        FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
                     },
                  },
                  {
                     Name: "XDL_IP", // 固定字段
                     ValueFrom: &v1.EnvVarSource{
                        FieldRef: &v1.ObjectFieldSelector{FieldPath: "status.hostIP"},
                     },
                  },
                  {
                     Name:  "framework",
                     Value: framework, // 使用的训练框架名称。支持MindSpore、PyTorch
                  },
               }
               return ct
            }
            ```

4. 创建命名空间，以vcjob为例，示例如下。

    ```go
    clientset.CoreV1().Namespaces().Create(context.TODO(), newNameSpace("vcjob"), metav1.CreateOptions{})
    ```

5. （可选）如果通过文件配置资源信息，还需要创建RankTable的ConfigMap，示例如下。

    ```go
    clientset.CoreV1().ConfigMaps(job.Namespace).Create(context.TODO(), newConfigMap("rings-config-"+job.Name), metav1.CreateOptions{})
    ```

6. 调用Create接口，下发训练任务。
    - Ascend Job

        ```go
        acjobClient.BatchV1().Jobs("default").Create(context.TODO(), &job, metav1.CreateOptions{})
        ```

    - Volcano Job

        ```go
        vcjobClient.BatchV1alpha1().Jobs("vcjob").Create(context.TODO(), &job, metav1.CreateOptions{})
        ```

7. 查看任务进程。调用Get接口查看job是否创建成功。
    - Ascend Job

        ```go
        acjobClient.BatchV1().Jobs("default").Get(context.TODO(), job.Name, metav1.GetOptions{})
        ```

    - Volcano Job

        ```go
        vcjobClient.BatchV1alpha1().Jobs("vcjob").Get(context.TODO(), job.Name, metav1.GetOptions{})
        ```

8. 删除任务。调用Delete接口删除任务。
    - Ascend Job

        ```go
        acjobClient.BatchV1().Jobs("default").Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
        ```

    - Volcano Job

        ```go
        vcjobClient.BatchV1alpha1().Jobs("vcjob").Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
        ```

**集成后使用<a name="section1027912153611"></a>**

1. 制作相应的镜像，可参考[制作镜像](../../common_operations.md#制作镜像)章节进行操作。
2. 完成相应的脚本适配，可参考[脚本适配](#脚本适配)章节进行操作。
3. 创建任务。
4. 运行训练任务。可通过平台配置并创建训练任务，下发任务后查看结果。
