# 故障恢复加速

## 产品描述

**产品介绍**

MindCluster MindIO Training Fault Tolerance（下文简称MindIO TFT）包括临终Checkpoint保存、进程级在线恢复和进程级别重调度等功能。

- MindCluster MindIO Try To Persist（下文简称MindIO TTP）功能，旨在针对大模型训练过程中故障恢复加速，MindIO TTP特性通过在训练过程中发生故障后，校验中间状态数据的完整性和一致性，生成一次临终Checkpoint数据，恢复训练时能够通过该Checkpoint数据恢复，减少故障造成的训练迭代损失。
- MindCluster MindIO Uncorrectable Memory Error（下文简称MindIO UCE）功能，旨在针对大模型训练过程中片上内存的UCE故障检测，并完成在线修复，达到Step级重计算。
- MindCluster MindIO Air Refuelling（下文简称MindIO ARF）功能，训练发生异常后，不用重启整个集群，只需以节点为单位进行重启或替换，对于部分故障仅需原地重启单进程，完成修复并继续训练。

**产品价值**

LLM（Large Language Model）是全球当前科技界竞争的焦点，LLM的训练往往需要长达数十天、甚至数月，Checkpoint是模型训练中断后恢复训练的关键点，Checkpoint过程中，整个集群中的训练任务会停滞，为了集群的利用率，Checkpoint的周期都配置得比较长，甚至达到数小时。这导致如果训练任务在即将生成Checkpoint数据的前一刻发生故障，未能生成本次Checkpoint数据，则只能从上一次的Checkpoint数据恢复，上次Checkpoint到故障前一刻的训练迭代需要重新计算，损失较大。MindIO TTP特性，在故障发生后，立即生成一次Checkpoint数据，恢复时也能立即恢复到故障前一刻的状态，减少迭代损失。

与此同时，LLM训练每一次保存Checkpoint数据并加载数据重新迭代训练所需时间同保存和加载周期Checkpoint类似都比较长，MindIO UCE在线修复，当NPU（Neural Processing Unit）发生UCE故障后，首先通过故障清理、故障恢复以及数据回滚等操作实现重新训练，恢复到故障前一刻的状态，节约训练停止重启时间；修复失败后走TTP流程作为保障措施。

**MindIO TFT架构**

![](../figures/scheduling/mindio_ttp架构.png)

MindIO TFT的各个功能集成在一个whl包中对外提供，需要通过import模块的方式，修改MindSpeed-LLM等大模型框架适配并使用对应功能。

MindIO TFT的关键点如下：

- MindIO TTP
    - 通过Controller和Processor模块，检测模型训练状态，并通过心跳定期汇报至Controller模块。一旦检测到故障，就开始临终Checkpoint保存。
    - 大模型训练中业界定期保存Checkpoint的时间间隔长。如果发生故障时，距离上一次保存的时间间隔过长，但又没到下一次保存的时间，此时如果重新训练就会消耗大量时间和资源。MindIO TTP提供了几乎零损时间和资源的重新训练方案，即重新训练从上一次故障处开始。

- MindIO UCE
    - 一旦检测到UCE故障，就开始在线修复。
    - 在大模型训练中，无论是定期保存Checkpoint，还是MindIO TFT的临终Checkpoint保存，重新训练的消耗都是巨大的。UCE提供了训练框架Step级重计算能力，不需要重启进程，同时能保证续训迭代损失，UCE失败后进入TTP流程。

- MindIO ARF
    - 针对更多的故障，不需要模型停止训练，只需通过节点重启或替换，完成修复和模型续训。
    - 对于业务进程异常故障或RestartRequest和RestartBusiness级别的芯片故障，可以通过进程级别原地恢复功能，仅重启故障进程完成修复和模型续训。

**逻辑模型**

- Controller模块：负责分布式任务的协同，内部维护状态机，状态机支持不同场景的流程控制；实时收集各个训练进程的训练状态，当训练发生异常后，结合异常类型，触发状态机运作，将状态机对应的Action发送到Processor模块执行。
- Processor模块：负责与训练框架交互，获取训练进程的训练状态，向Controller汇报，同时负责执行Controller模块下发的对应Action动作。
- Adaptor模块：负责完成训练框架对MindIO TTP、MindIO UCE、MindIO ARF特性的适配。目前MindIO TFT已完成对[MindSpeed-LLM](#对接mindspeed-llm框架)训练框架的适配。对于其他训练框架，需用户参考并自行适配。

**部署形态**

- Controller模块：在整个训练集群中，仅支持存在一个Active Controller，建议部署在集群0号节点上，并自动启动最多两个Backup Controller。
- Processor模块：在整个训练集群中，每个训练进程均需要启动Processor。

## 安装部署

### 安装前必读

#### 免责声明

本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。

用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。

#### 约束限制

- MindIO提供TTP、UCE和ARF三种特性，其中MindIO TTP支持在Atlas 800 训练服务器（型号：9000）上使用，MindIO UCE和MindIO ARF不支持该型号设备。
- 众多大模型框架都支持ZeRO（Zero Redundancy Optimizer，零冗余优化器）来减少对显存的使用，当前MindIO TFT仅支持开启ZeRO-1，支持DP（Data Parallelism，数据并行） Size为偶数，同时使用不同的功能对DP Size有不同的限制：
    - MindIO TTP功能
        - 为了保证故障发生后，有完整的优化器状态数据，要求DP Size能被副本数整除。
        - 开启MoE（Mixture of Experts，混合专家结构）前要求稠密层DP Size大于1；开启MoE后要求稠密层和稀疏层DP Size都大于1。
        - 针对分布式优化器，MindIO TFT在ZeRO-1功能的基础上，通过以算代传，在DP Group上重新切分优化器ZeRO-1范围，实现了优化器数据副本。

    - MindIO UCE和MindIO ARF功能
        - 若要实现从当前Step恢复训练，对DP Size限制与MindIO TTP功能一致。
        - 对于显存有限，不做副本的情况，即DP Size = 1，此时若发生UCE或者节点故障，支持在线从周期性Checkpoint中加载模型权重和优化器参数恢复训练，损失当前Step到上次周期性Checkpoint的Step之间的训练成本。

    - 分布式优化器在开启ZeRO特性后，优化器状态数据全局只有一份，无数据冗余。MindIO TFT通过增加优化器状态冗余数据副本，保证故障场景下优化器状态数据的完整性，但同时该方案会导致片上内存使用增加。在原有的模型配置基础上，直接使用MindIO TFT可能会导致模型训练启动过程中出现片上内存OOM（Out Of Memory，内存不足）异常。在此情况下，需要通过扩容增加训练作业的片上内存总量。

        增加副本对应增加的片上内存大小计算公式：增加片上内存总量（GB） = 模型参数量N（B） \* 12 \* 副本数。其中，模型参数量的单位为B（十亿），通过以上公式，计算出需要增加的片上内存，扩容后，再使用MindIO TFT。

- 训练容错框架中有一个Active Controller与两个Backup Controller，为了包括Active Controller在内多张卡发生故障时，能够顺利切换到Backup Controller完成临终保存，需要状态正常的卡的数量大于world\_size的一半。
- MindIO TFT会对优化器状态数据做副本，MindIO UCE或MindIO ARF修复时，寻找有效副本修复故障卡，当训练集群故障较多，通过副本仍然无法拼凑出一个完整副本时，则从Step在线修复退化为在线加载周期Checkpoint修复。
- MindIO TFT在生成临终Checkpoint数据时，除了考虑一个完整的数据副本，还要校验数据是否一致。如果发生故障后，存在一个OS（Optimizer State，优化器状态）数据Shard长期处于修改状态，或者OS数据不同Shard间训练迭代不一致，都认为是全局数据不一致，无法生成临终Checkpoint数据。
- MindIO TTP不使用MindIO ACP（Async Checkpoint Persistence，异步Checkpoint保存）功能。MindIO TTP完成临终Checkpoint保存后会结束训练进程。为确保在进程退出前，临终Checkpoint已经保存到持久化存储，约束MindIO TTP写数据不使用异步Checkpoint保存方式，而是直接写入到持久化存储。
- MindIO TFT目前不支持级联故障场景。例如：当MindIO TTP正在保存时，如果出现其他故障，就会保存失败。
- MindIO TFT会增加显存占用，详情请参见[表1 原生优化器与开启故障快速恢复特性后优化器参数的理论数值变化](#table_tft_03)。
- 默认开启TLS（Transport Layer Security，传输层安全性协议）安全特性，关闭可能导致伪造Controller连接影响训练进程。
- MindIO ARF需要多个节点（≥2），不支持Controller节点发生故障，不支持级联故障；MindIO ARF修复失败后，由MindCluster控制后续流程。
- 日志保存路径默认在运行脚本同级目录下“logs/ttp\_log.log”文件，可在运行脚本里自行配置，默认日志级别为“INFO”，单日志文件大小限制为10MB，写方式为单个追加写，单日志文件达到大小上限后会新建滚动日志文件，滚动日志文件数量限制为5个，多个文件循环写覆盖旧文件。

### 安装前准备

#### 组网规划

**图 1**  部署逻辑示意图  
![](../figures/scheduling/部署逻辑示意图ttp.png "部署逻辑示意图")

深度学习平台与训练任务相关的节点有计算节点和存储节点。各类节点主要功能如下：

- 计算节点：实际执行训练、推理任务的节点，MindIO TFT仅部署在计算节点。
- 存储节点：存储平台数据和用户数据，如平台日志、用户上传的数据集、训练脚本、训练输出的模型等。

网络平面划分为：

- 业务面：用于管理集群业务。管理节点和计算节点之间连接。
- 存储面：用于访问存储节点。管理节点和计算节点连接到存储节点。
- 参数面：用于分布式训练时，训练节点之间的参数交换和连接。

    > [!NOTE]说明
    > - 逻辑部署示意图展示深度学习平台的完整示意图，MindIO TFT特性只需要在计算节点上部署一个SDK（Software Development Kit），不涉及存储节点的安装部署。
    > - MindIO TFT功能SDK需要在计算节点相互通信，发送心跳报文，需要使用业务面网络，SDK在所有运行大模型训练的计算节点对等部署，部署时不区分管理节点和计算节点。

#### 环境要求

**硬件环境**

安装前，需要检查以下硬件配置，如[表1](#table_tft_01)所示。

**表 1<a id="table_tft_01"></a>**  硬件环境

|类型|配置参考|
|--|--|
|服务器（单机场景）|<ul><li>Atlas 800 训练服务器（型号：9000）：仅支持MindIO TTP功能</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A3 SuperPoD 超节点</li></ul>|
|服务器（集群场景）|计算节点：<ul><li>Atlas 800 训练服务器（型号：9000）：仅支持MindIO TTP功能</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A3 SuperPoD 超节点</li></ul> 存储节点：存储服务器|
|网络|<ul><li>带外管理（BMC）：≥1Gbit/s</li><li>带内管理（SSH）：≥1Gbit/s</li><li>业务面：≥10Gbit/s</li><li>存储面：≥25Gbit/s</li><li>参数面：100Gbit/s</li></ul>|

**软件环境**

安装前，需要完成以下环境的安装，如[表2](#table_tft_02)所示。

**表 2<a id="table_tft_02"></a>**  软件环境

|软件|版本|安装位置|获取方式|
|--|--|--|--|
|操作系统|<ul><li>CentOS 7.6</li><li>Ubuntu 18.04</li><li>Ubuntu 20.04</li><li>Ubuntu 22.04</li></ul>|所有节点|-|
|Python|3.7 ~ 3.11|计算节点|用户安装|
|Torch|2.7.1|计算节点|用户安装|
|torch_npu|7.3.0|计算节点|用户安装|
|CANN|8.5.0|计算节点|用户安装|
|驱动与固件|25.5.0|计算节点|用户安装|

#### 准备软件包

**下载软件包**

下载本软件即表示您同意[华为企业业务最终用户许可协议（EULA）](https://e.huawei.com/cn/about/eula)的条款和条件。

**表 1**  软件下载

|组件名称|软件包|获取地址|
|--|--|--|
|MindIO TFT|内存缓存系统软件包|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases)|

**软件数字签名验证**

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请勿使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[http://support.huawei.com/carrier/digitalSignatureAction](http://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

#### （可选）启动haveged服务

1. 确认系统是否开启了haveged服务（建议一直开启）。

    ```bash
    systemctl status haveged.service
    ```

    或

    ```bash
    ps -ef | grep "haveged" | grep -v "grep"
    ```

2. 启动haveged服务，并将其设置为随系统启动，确保haveged服务一直开启。

    ```bash
    systemctl start haveged.service
    systemctl enable haveged.service
    ```

3. 查看屏幕输出随机数的速度。

    ```bash
    cat /dev/random | od -x
    ```

    查看当前熵值。

    ```bash
    cat /proc/sys/kernel/random/entropy_avail
    ```

    正常情况下，熵值在未启动haveged时是100多，启动haveged之后会增大到1000多甚至2000。

### 在计算节点安装MindIO TFT SDK

在大模型训练框架使用的Python环境中，安装MindIO TFT SDK，可使能训练任务的故障恢复，从而加速训练恢复。

**操作步骤**

1. 以安装用户 *{MindIO-install-user}* 登录安装节点。

    > [!NOTE]说明
    > 安装用户设置的口令需符合口令复杂度要求（请参见[口令复杂度要求](#口令复杂度要求)）。密码有效期为90天，您可以在“/etc/login.defs”文件中修改有效期的天数，或者通过 **chage** 命令来设置用户的有效期，详情请参见[设置用户有效期](#设置用户有效期)。

2. 将内存缓存系统软件包上传至设备上安装用户有权限读写的路径下。

    > [!NOTE]说明 
    > - 内存缓存系统软件包以获取的实际包名为准。
    > - 如果Python环境是共享目录，则在任一计算节点上传即可，否则所有计算节点都需要上传安装包。

3. 进入软件包上传路径，解压内存缓存系统软件包。

    ```bash
    unzip Ascend-mindxdl-mindio_{version}_linux-{arch}.zip
    ```

    **表 1**  解压后文件

    |文件|说明|
    |--|--|
    |mindio_acp-*{mindio_acp_version}*-py3-none-linux_*{arch}*.whl|MindIO ACP安装包。|
    |mindio_ttp-*{mindio_ttp_version}*-py3-none-linux_*{arch}*.whl|MindIO TFT安装包。|

4. 进入上传路径，执行以下命令，安装MindIO TFT SDK。

    此处以mindio_ttp-_*{mindio_ttp_version}*_-py3-none-linux_*{arch}*.whl为例，请根据实际情况进行选择。

    ```bash
    pip3 install mindio_ttp-{mindio_ttp_version}-py3-none-linux_{arch}.whl --force-reinstall --no-index
    ```

    - 首次安装MindIO TFT SDK回显如下，表示安装成功。

        ```bash
        Processing ./mindio_ttp-{mindio_ttp_version}-py3-none-linux_{arch}.whl
        Installing collected packages: mindio_ttp
        Successfully installed mindio_ttp-{mindio_ttp_version}
        ```

    - 非首次安装MindIO TFT SDK回显如下，表示安装成功。

        ```bash
        Processing ./mindio_ttp-{mindio_ttp_version}-py3-none-linux_{arch}.whl
        Installing collected packages: mindio_ttp
          Atempting uninstall: mindio-ttp
            Found existing installation: mindio_ttp {mindio_ttp_version}
            Uninstalling mindio_ttp-{mindio_ttp_version}:
              Successfully uninstalled mindio_ttp-{mindio_ttp_version}
        Successfully installed mindio_ttp-{mindio_ttp_version}
        ```

5. 将软件安装目录内的可执行文件和代码脚本权限更改为550，避免出现非法篡改。

    ```bash
    chmod -R 550 {MindIO TFT SDK安装目录}
    ```

### 卸载MindIO TFT SDK

**操作步骤**

1. 将软件安装目录内的可执行文件和代码脚本权限更改为750。

    ```bash
    chmod -R 750 {MindIO TFT SDK安装目录}
    ```

2. 卸载MindIO TFT SDK。

    ```bash
    pip3 uninstall mindio_ttp
    ```

## 使用指导

### 概述

> [!NOTE]说明
> MindIO TFT以SDK的形式提供服务，支持部署在裸机和容器环境中。

安装MindIO TFT SDK之后，需要在框架中启动MindIO TFT模块，并在训练过程中同步优化器数据更新状态到该模块。

### 对接MindSpeed-LLM框架

**前提条件**

- 使用前请先了解MindIO TFT的[约束限制](#约束限制)。
- MindSpeed-LLM框架准备参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/master)。匹配的Megatron-LM的版本为 **core\_v0.12.1**。

> [!NOTE]说明
>
> - 本次发布包配套MindSpeed-LLM的 **master** 分支，环境、代码、数据集准备请用户参考MindSpeed-LLM仓库的相关指导说明，并确保其安全性。
> - MindIO TFT对接MindSpeed-LLM框架，目前支持MindIO TTP、MindIO UCE和MindIO ARF功能。
> - 对于PyTorch类框架，安装或开启MindCluster后，跳过步骤[1](#step_tft_li001)对torchrun的修改，由MindCluster控制进程退出。

**操作步骤**

1. <a id="step_tft_li001"></a>（可选）编辑“torchrun”文件。
    1. 查找环境中的“torchrun”文件。

        ```bash
        which torchrun
        ```

    2. 打开以上命令显示路径下的“torchrun”文件。

        ```bash
        vim {torchrun文件路径}/torchrun
        ```

    3. 按“i”进入编辑模式，在文件中对应位置增加 **import mindio_ttp.framework_ttp**。

        ```python
        import re
        import sys
        import mindio_ttp.framework_ttp
        from torch.distributed.run import main as torch_main
        ```

    4. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

2. <a id="step_tft_li002"></a>编辑预训练脚本（仅供参考）。

    此处以编辑“examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh”脚本为例。

    1. 打开“examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh”脚本。

        ```bash
        vim examples/mcore/llama2/pretrain_llama2_7b_ptd.sh
        ```

    2. 按“i”进入编辑模式，开启高可用功能需要在脚本中增加如下内容。

        ```bash
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export TTP_ADDR="master node ip"
        source /usr/local/Ascend/cann/set_env.sh

        # 在GPT_ARGS中的--bf16后增加如下内容
            \
            --enable-high-availability \
            --enable-hbmfault-repair \
            --enable-worker-reboot \
            --distributed-optimizer-no-replica \

        ```

        修改后的pretrain_llama2_7b_ptd.sh脚本示例如下：

        ```bash
        #!/bin/bash
        
        export CUDA_DEVICE_MAX_CONNECTIONS=1
        export PYTORCH_NPU_ALLOC_CONF=expandable_segments:True
        
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export TTP_ADDR="master node ip"
        source /usr/local/Ascend/cann/set_env.sh
        
        NPUS_PER_NODE=8
        MASTER_ADDR=localhost
        MASTER_PORT=6000
        NNODES=1
        NODE_RANK=0
        WORLD_SIZE=$(($NPUS_PER_NODE*$NNODES))
        
        CKPT_SAVE_DIR="your model save ckpt path"
        DATA_PATH="your data path"
        TOKENIZER_MODEL="your tokenizer path"
        CKPT_LOAD_DIR="your model ckpt path"
        TP=1
        PP=2
        
        DISTRIBUTED_ARGS="
            --nproc_per_node $NPUS_PER_NODE \
            --nnodes $NNODES \
            --node_rank $NODE_RANK \
            --master_addr $MASTER_ADDR \
            --master_port $MASTER_PORT
        "
        
        GPT_ARGS="
            --use-mcore-models \
            --tensor-model-parallel-size ${TP} \
            --pipeline-model-parallel-size ${PP} \
            --sequence-parallel \
            --num-layers 32 \
            --hidden-size 4096 \
            --ffn-hidden-size 11008 \
            --num-attention-heads 32 \
            --tokenizer-type Llama2Tokenizer \
            --tokenizer-model ${TOKENIZER_MODEL} \
            --seq-length 4096 \
            --max-position-embeddings 4096 \
            --micro-batch-size 1 \
            --global-batch-size 256 \
            --make-vocab-size-divisible-by 1 \
            --lr 1.25e-6 \
            --train-iters 5000 \
            --lr-decay-style cosine \
            --untie-embeddings-and-output-weights \
            --disable-bias-linear \
            --attention-dropout 0.0 \
            --init-method-std 0.01 \
            --hidden-dropout 0.0 \
            --position-embedding-type rope \
            --normalization RMSNorm \
            --use-fused-rmsnorm \
            --swiglu \
            --use-flash-attn \
        
            --no-masked-softmax-fusion \
            --attention-softmax-in-fp32 \
            --min-lr 1.25e-7 \
            --weight-decay 1e-1 \
            --lr-warmup-fraction 0.01 \
            --clip-grad 1.0 \
            --adam-beta1 0.9 \
            --initial-loss-scale 65536 \
            --adam-beta2 0.95 \
            --no-gradient-accumulation-fusion \
            --no-load-optim \
            --no-load-rng \
            --use-distributed-optimizer \
            --use-fused-swiglu \
            --use-fused-rotary-pos-emb \
            --overlap-grad-reduce \
            --bf16 \
            --enable-high-availability \
            --enable-hbmfault-repair \
            --enable-worker-reboot \
            --distributed-optimizer-no-replica \
        "
        
        DATA_ARGS="
            --data-path $DATA_PATH \
            --split 949,50,1
        "
        
        OUTPUT_ARGS="
            --log-interval 1 \
            --save-interval 10000 \
            --eval-interval 1000 \
            --eval-iters 10 \
        "
        
        torchrun $DISTRIBUTED_ARGS pretrain_gpt.py \
            $GPT_ARGS \
            $DATA_ARGS \
            $OUTPUT_ARGS \
            --distributed-backend nccl \
            --load $CKPT_LOAD_DIR \
            --save $CKPT_SAVE_DIR \
            | tee logs/train_llama2_7b.log
        ```

        高可用功能相关参数说明如下：

        - **GLOO\_SOCKET\_IFNAME**：根据主节点高速网卡实际情况进行配置。
        - **TTP\_ADDR**：集群主节点的IP地址,要求为常规IPv4或IPv6格式。参数详情请参见[环境变量](#环境变量)。
        - **set\_env.sh文件路径**：请根据CANN实际的安装路径进行修改。
        - **enable-high-availability**：MindIO TFT总开关，默认关闭，配置后默认开启临终遗言功能。

            开启MindIO TFT开关后，各类优化器显存会发生变化，变化详情请参见[表1](#table_tft_03)。

            对于分布式优化器而言，由于增加了优化器副本，导致静态内存有所增加。但是集群规模越大时，DP Size越大，平均到单卡的显存增加量很小，这样可以避免OOM，因此推荐在大集群中使用。根据显存情况选择开启与否，调节参数。

        - **enable-hbmfault-repair**：MindIO UCE功能开关，默认关闭，配置后对片上内存进行故障检测，并完成在线修复，达到Step级重计算功能。本开关在开启 enable-high-availability 时生效。此特性依赖PyTorch的内存管理机制，仅在PyTorch的环境变量 PYTORCH\_NO\_NPU\_MEMORY\_CACHING 未配置，即开启内存复用机制时，才可使用此特性，若 export PYTORCH\_NO\_NPU\_MEMORY\_CACHING = 1，则无法使用此特性。
        - **enable-worker-reboot**：MindIO ARF功能开关，默认关闭，配置后在发生一般性故障时，进行进程级重启修复，继续训练。本开关在开启 enable-high-availability 时生效。
        - **distributed-optimizer-no-replica**：开启高可用特性后，分布式优化器默认增加优化器副本，会导致片上内存使用增加，开启该开关后，分布式优化器不增加副本内存占用；在MindIO UCE和MindIO ARF场景下，直接使用周期Checkpoint进行在线修复。

        **表 1<a id="table_tft_03"></a>**  原生优化器与使用MindIO TFT后优化器参数的理论数值变化

        |优化器|原生|使用MindIO TFT|说明|
        |--|--|--|--|
        |fp16/bf16|20|20|-|
        |fp32|16|16|-|
        |fp16/bf16 Distributed|4 + 16/d|4 + 16 * N/d|<ul><li>d：DP Group Size</li><li>N：副本数，N < d</li></ul>|

    3. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

### 对接MindCluster

MindIO TFT以SDK形式提供服务，不存在常驻进程。服务随着训练进程的启动而启动。当训练任务结束时，则服务退出。

与MindCluster对接时，MindCluster管理K8s容器，在K8s容器中安装对接过程与裸机安装部署一致。

**操作步骤**

- 当Python环境不是安装在共享存储中时，为了便于大集群使用，可以将MindIO TFT SDK集成到镜像中，通过镜像安装Pod时，已经安装好MindIO TFT SDK。
- MindIO TFT服务Controller模块与Processor模块存在心跳报文，在K8s做网络隔离时，需要将通信端口添加到创建Pod时配置的yaml文件中。

    修改创建Pod时配置的yaml文件。此处以“pod.yaml”为例。

    1. 打开“pod.yaml”文件。

        ```bash
        vim pod.yaml
        ```

    2. 按“i”进入编辑模式，新增以下内容。

        ```yaml
        ports:
          - containerPort: 8000    # 用于MindIO TFT服务Controller与Processor通信端口
            name: ttp-port
        ```

    3. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

- 适配K8s网络，在步骤[2](#step_tft_li002)的预训练脚本基础上做如下修改。

    ```yaml
    # 注释下面两行，该环境变量由MindCluster配置
    # MASTER_ADDR=$(hostname -I | awk '{print $1}')
    # MASTER_PORT=XXXX
    
    # 从K8s获取MASTER_ADDR、MASTER_PORT环境变量（K8s的service网络IP地址）
    CONTROLLER_ADDR=$(hostname -I | awk '{print $1}')
    PROCESSOR_ADDR=${MASTER_ADDR}
    export CONTROLLER_ADDR
    export PROCESSOR_ADDR
    ```

### 对接非MindSpeed-LLM框架

**前提条件**

使用前请先了解MindIO TFT的[约束限制](#约束限制)。

> [!NOTE]说明
>
> - 本次发布包支持类Megatron框架，环境、代码、数据集请用户自行准备，并确保其安全性。
> - 本节内容仅具有适配指导意义，具体实现细节需由用户自行实现。

**特性参考**

相关特性所需的功能适配点如[表1](#table_tft_04)所示，各功能适配点对应的代码参考链接如[表2](#table_tft_05)所示。

**表 1<a id="table_tft_04"></a>**  特性及功能适配点

|特性|需要的功能适配点序号|
|--|--|
|临终遗言|1、2、3、4、5、6、7|
|UCE快恢|1、2、3、4、5、6、8、10、11|
|网络快恢|1、2、5、6、11|
|进程快恢|1、2、3、4、5、6、9、10、11|
|亚健康热切|1、2、3、4、5、9、10、11、12|
|在线压测/借轨回切|1、2、12|

**表 2<a id="table_tft_05"></a>**  相关功能的代码参考链接

|序号|适配功能点|参考代码|
|--|--|--|
|1|初始化启动|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/268f870b10e450feade3c98b603254851e8fa4cd?ref=pre_preparation)|
|2|上报优化器更新状态|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/268f870b10e450feade3c98b603254851e8fa4cd?ref=pre_preparation)|
|3|创建DP副本组|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/df6317e62ef7cefcec25ba8740f25e152eba34e4?ref=create_dp_replica_group)|
|4|优化器副本|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/e3490911407d88f9c6d3ac0c0eb3186f1812d171?ref=replica_optimizer)|
|5|异常捕获装饰器|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/0827869d031303a231a69897c12692fb92d8cf8d?ref=exception_handler)|
|6|算子资源清理|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/45824ee7303c05bce1260f2cab590dd858147767?ref=stop_clean)|
|7|临终Checkpoint|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/0e94a3fcb2643580d151b90deb205e9034adde2a?ref=dump_ckpt)|
|8|UCE模型优化器重建|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/93f599fa480c7f7931c74e782c617e0ebaffceb9?ref=uce_clear_rebuild)|
|9|节点重启及通信重建|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/4b490ff888cea9766e461f6bb53e73712adf097d?ref=node_reboot)|
|10|参数面在线修复|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/9bd17ca7fdda3f8c5f70eef68cf1db4ac2ba738f?ref=online_repair_ckpt)|
|11|状态回滚|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/7835670ec12b2ae5969bd1cd9eec72c882225c18?ref=rollback_callback)|
|12|优雅暂停|[LLM仓参考链接](https://gitcode.com/wlwen/MindSpeed-LLM/commit/db87cc048455f67218f5a8caca626f1b64d35f61?ref=active_pause)|

## 安全管理与加固

### 安全管理

> [!NOTE]说明 
> MindIO TFT暂不支持公有云场景、多租户场景使用，不支持公网直接访问系统。

**防病毒软件例行检查**

定期开展对集群的防病毒扫描，防病毒例行检查会帮助集群免受病毒、恶意代码、间谍软件以及恶意程序侵害，降低系统瘫痪、信息安全问题等风险。可以使用业界主流防病毒软件进行防病毒检查。

**日志管理**

日志管理需要关注以下两点。

- 检查系统是否可以限制单个日志文件的大小。
- 检查日志空间占满后，是否存在机制进行清理。

**漏洞/功能问题修复**

为保证生产环境的安全，降低被攻击的风险，需要定期查看开源社区修复的以下漏洞/功能问题。

- 操作系统漏洞/功能问题。
- 其他相关组件漏洞/功能问题。

### 安全加固

#### 加固须知

本文中列出的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施，必要时可参考业界优秀加固方案和安全专家的建议。

#### 风险提示

Checkpoint序列化过程中使用了torch.load接口，该接口中使用了Python自带的pickle组件，必须确保非授权用户没有存储目录及上层目录的写权限，需保证Checkpoint为可信数据，否则可能造成Checkpoint被篡改引起pickle反序列化注入的风险。

#### 操作系统安全加固

**防火墙配置**

操作系统安装后，若配置普通用户，可以通过在“/etc/login.defs”文件中新增“ALWAYS\_SET\_PATH=yes”配置，防止越权操作。此外，为了防止使用“su”命令切换用户时，将当前用户环境变量带入其他环境造成提权，请使用 **su - [user]** 命令进行用户切换，同时在服务器配置文件“/etc/default/su”中增加配置参数“ALWAYS\_SET\_PATH=yes”防止提权。

**设置umask**

建议用户将服务器的umask设置为027\~777以限制文件权限。

以设置umask为027为例，具体操作如下。

1. 以root用户登录服务器，编辑“/etc/profile”文件。

    ```bash
    vim /etc/profile
    ```

2. 在“/etc/profile”文件末尾加上 **umask 027**，保存并退出。
3. 执行如下命令使配置生效。

    ```bash
    source /etc/profile
    ```

**无属主文件安全加固**

用户可以执行 **find / -nouser -nogroup** 命令，查找容器内或物理机上的无属主文件。根据文件的UID和GID创建相应的用户和用户组，或者修改已有用户的UID、用户组的GID来适配，赋予文件属主，避免无属主文件给系统带来安全隐患。

**端口扫描**

用户需要关注全网侦听的端口和非必要端口，如有非必要端口请及时关闭。建议用户关闭不安全的服务，如Telnet、FTP等，以提升系统安全性。具体操作方法可参考所使用操作系统的官方文档。

**防DoS攻击**

用户可以根据IP地址限制与服务器的连接速率对系统进行防DoS攻击，方法包括但不限于利用Linux系统自带Iptables防火墙进行预防、优化sysctl参数等。具体使用方法，用户可自行查阅相关资料。

**SSH加固**

由于root用户拥有最高权限，出于安全目的，建议取消root用户SSH远程登录服务器的权限，以提升系统安全性。具体操作步骤如下：

1. 登录安装MindIO TFT组件的节点。
2. 打开“/etc/ssh/sshd\_config”文件。

    ```bash
    vim /etc/ssh/sshd_config
    ```

3. 按“i”进入编辑模式，找到“PermitRootLogin”配置项并将其值设置为“no”。

    ```text
    PermitRootLogin no
    ```

4. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。
5. 执行命令使配置生效。

    ```bash
    systemctl restart sshd
    ```

**缓冲区溢出安全保护**

为阻止缓冲区溢出攻击，建议使用ASLR（Address Space Layout Randomization，内存地址随机化机制）技术，通过对堆、栈、共享库映射等线性区布局的随机化，增加攻击者预测目的地址的难度，防止攻击者直接定位攻击代码位置。该技术可作用于堆、栈、内存映射区（mmap基址、shared libraries、vdso页）。

开启方式：

```bash
echo 2 >/proc/sys/kernel/randomize_va_space
```

### 开启TLS认证

#### 说明

- 为了保障MindIO TFT组件内部Controller和Processor之间的通信安全，保护信息不被篡改、仿冒，建议启用TLS加密。
- TLS加密仅用于MindIO TFT内部模块间通信，不对外提供TLS接入、认证功能。
- 因为开启安全认证依赖OpenSSL组件，所以建议用户使用OpenSSL无漏洞版本，需要配套使用GLIBC 2.33或更高版本。

#### 导入TLS证书

- 通过接口tft\_start\_controller、tft\_init\_processor配置TLS密钥证书等，进行TLS安全连接，安全选项默认开启，建议用户开启TLS加密配置，以保证通信安全，如需关闭加密功能，可以使用下面示例，调用接口进行关闭。
- 系统启动后，建议删除本地密钥证书等敏感信息文件。
- 调用该接口时，传入的文件路径应避免包含英文分号、逗号、冒号。
- 支持通过环境变量 **TTP\_ACCLINK\_CHECK\_PERIOD\_HOURS** 和 **TTP\_ACCLINK\_CERT\_CHECK\_AHEAD\_DAYS** 配置证书检查周期与证书过期预警时间。

**配置TLS接口调用示例**

- TLS关闭（**enable\_tls**=False）时，**tls\_info**无效，无需配置。此开关不影响MindIO TFT特性功能。

    ```python
    from mindio_ttp.framework_ttp import tft_start_controller, tft_init_processor
    
    tft_start_controller(bind_ip: str, port: int, enable_tls=False, tls_info='')
    tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=False, tls_info='', enable_uce=True, enable_arf=False)
    ```

    > [!CAUTION]注意
    > - 如果关闭TLS（即**enable\_tls**=False时），会存在较高的网络安全风险。
    > - **tft\_start\_controller** 和 **tft\_init\_processor** 的enable\_tls开关状态需要保持一致。若两个接口enable\_tls开关不同，会造成以下问题：
    >   - 模块间TLS建链失败。
    >   - MindIO TFT无法正常运行，训练任务启动失败。

- TLS开启（**enable\_tls**=True）时，证书相关信息，作为必选参数 **tls\_info** 用于如下接口：

    ```python
    from mindio_ttp.framework_ttp import tft_start_controller, tft_init_processor, tft_register_decrypt_handler
    
    # 在tls_info中，以“;”分隔不同字段,以“,”分隔各个文件
    tls_info = r"(
    tlsCert: /etc/ssl/certs/cert.pem;
    tlsCrlPath: /etc/ssl/crl/;
    tlsCaPath: /etc/ssl/ca/;
    tlsCaFile: ca_cert_1.pem, ca_cert_2.pem;
    tlsCrlFile: crl_1.pem, crl_2.pem;
    tlsPk: private key;
    tlsPkPwd: private key pwd;
    packagePath: /etc/ssl/
    )"
    
    # 若tlsPkPwd口令为密文，则需注册口令解密函数
    tft_register_decrypt_handler(user_decrypt_callback)
    tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info=tls_info)
    tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info=tls_info, enable_uce=True, enable_arf=False)
    ```

**tls\_info中各字段含义**

|字段|含义|Required|
|--|--|--|
|tlsCert|Server证书。|是|
|tlsCaPath|CA证书存储路径。|是|
|tlsCaFile|CA证书列表。|是|
|tlsCrlPath|证书吊销列表存储路径。|否|
|tlsCrlFile|证书吊销列表。|否|
|tlsPk|私钥。|是|
|tlsPkPwd|私钥口令。|是|
|packagePath|OpenSSL库路径|是|

> [!CAUTION]注意
> 证书安全要求：
>
> - 需使用业界公认安全可信的非对称加密算法、密钥交换算法、密钥长度、Hash算法、证书格式等。
> - 应处于有效期内。

#### （可选）证书有效性校验

如果启用TLS认证，则需要关注证书有效期。请合理规划证书有效期和证书更新周期，并在证书过期前及时更新证书，防范安全风险。MindIO TFT提供证书有效期定期巡检功能，默认巡检周期为7天，默认提前告警时间为30天，若发现证书存在过期风险，则会在环境变量 **TTP\_LOG\_PATH** 配置的日志中打印WARNING告警信息，请及时关注并处理。

## API接口参考

### 说明

所有接口参数表和回调函数参数表，默认按照函数参数顺序排列。

### tft\_init\_controller

**接口功能**

初始化MindIO TFT Controller模块。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_init_controller(rank: int, world_size: int, enable_local_copy: bool, enable_arf=False, enable_zit=False)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前执行训练任务的NPU卡号。|int，[-1, world_size)。MindCluster在Torch Agent进程拉起Controller时rank值取-1。|
|world_size|必选|整个集群参与训练任务的卡数。|int，[1, 100000]。|
|enable_local_copy|必选|表示是否启用local copy。优化器更新前，先对优化器做一次备份。|<ul><li>False：关闭</li><li>True：启用</li></ul>|
|enable_arf|可选|MindIO ARF特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|
|enable_zit|可选|MindIO ZIT特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_start\_controller

**接口功能**

在初始化Controller模块成功后，调用该接口以启动MindIO TFT Controller模块服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info='')
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|bind_ip|必选|Controller所在节点IP地址或域名。|符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。|
|port|必选|Controller侦听端口号。|[1024, 65535]|
|enable_tls|可选|TLS加密传输开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为True。|
|tls_info|可选|TLS的证书配置。|默认为空，当开启TLS认证时，需要配置证书信息，具体字段应以键值对形式组织。具体配置指导见[导入TLS证书](#导入tls证书)。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_destroy\_controller

**接口功能**

在训练完成后，调用该接口以关闭MindIO TFT Controller服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_destroy_controller()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_init\_processor

**接口功能**

初始化MindIO TFT Processor模块。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info='', enable_uce=True, enable_arf=False, enable_zit=False)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前执行训练任务NPU卡号。|int，[0, world_size)。|
|world_size|必选|参与训练任务的集群卡数。|int，[1, 100000]。|
|enable_local_copy|必选|是否启用local copy。|<ul><li>False：关闭</li><li>True：启用</li></ul>|
|enable_tls|可选|TLS加密传输开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为True。|
|tls_info|可选|TLS的证书配置。|默认为空，当开启TLS认证时，需要配置证书信息，具体字段应以键值对形式组织。具体配置指导见[导入TLS证书](#导入tls证书)。|
|enable_uce|可选|MindIO UCE特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为True。|
|enable_arf|可选|MindIO ARF特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|
|enable_zit|可选|MindIO ZIT特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_start\_processor

**接口功能**

在初始化Processor模块成功后，调用该接口以启动MindIO TFT Processor模块服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_processor(master_ip: str, port: int, local_ip='')
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|master_ip|必选|Controller所在节点IP地址或域名。|符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。|
|port|必选|Controller侦听端口号。|[1024, 65535]|
|local_ip|可选|K8s中Processor所在节点的Service IP地址或域名。|符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_destroy\_processor

**接口功能**

在训练完成后，调用该接口以关闭MindIO TFT Processor服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_destroy_processor()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_start\_updating\_os

**接口功能**

在优化器状态更新前，调用该接口以更新optimizer state为Updating。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_updating_os(backup_step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|backup_step|必选|备份的step。|-1或自然数，范围[-1, 9223372036854775807)。<ul><li>-1：表示不使用备份step。</li><li>自然数：优化器更新前，备份的优化器状态数据对应的step。</li></ul>|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_start\_copy\_os

**接口功能**

通知Processor开始copy优化器状态。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_copy_os()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_end\_updating\_os

**接口功能**

在优化器状态更新完成后，调用该接口以更新optimizer state为Updated。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_end_updating_os(step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|step|必选|当前的step。|正整数，范围[1, 9223372036854775807)。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_set\_optimizer\_replica

**接口功能**

设置rank对应的优化器状态数据副本关系。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_set_optimizer_replica(rank: int, replica_info: list)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前执行训练任务的NPU卡号。|int，[0, 100000)。|
|replica_info|必选|副本关系list，其中每个元素是一个字典，字典按照ATTENTION（0）、MOE（1）的索引顺序排列。|[<br>{<br>"rank_list":list,   # 对应的一组副本关系rank列表，PyTorch场景为DP组rank list,MindSpore场景为该卡对应的所有副本卡的list <br>"replica_cnt":int,   # 副本数，PyTorch场景为副本数，MindSpore场景为rank_list的长度 <br>"replica_shift":int,  # PyTorch场景有效<br>},<br>]|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_exception\_handler

**接口功能**

装饰器，对MindSpeed-LLM的train方法进行装饰，捕获训练状态异常以及上报处理，对于用户的其他训练框架，本接口仅提供参考示例功能。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_exception_handler(func: Callable)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|函数作为参数。|框架的train方法。|

**返回值**

装饰器返回的func。

### tft\_set\_step\_args

**接口功能**

训练框架设置的参数集合。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，设置功能已经由MindIO TFT完成适配，不需要调用。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_set_step_args(args)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|args|必选|训练框架设置需要保存的参数集合。MindIO TFT在 stop/clean/repair/rollback 等阶段调用注册的回调函数时，将参数集合传回，框架根据参数集合完成相应功能。|由训练框架决定，MindIO TFT不访问也不修改该参数集合，在 stop/clean/repair/rollback 等阶段时调用注册的业务回调将其传回，业务回调负责对取值范围进行校验。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_rename\_handler

**接口功能**

注册框架侧rename回调函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_rename_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rename函数，将保存成功的临终Checkpoint重命名，与原生框架Checkpoint命名规则一致。|回调函数，不为空，回调函数的入参要求请参见[表 1](#table_tft_06)和[表 2](#table_tft_07)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_06"></a>**  MindSpore回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|dump优化器数据时对应的step。|正整数。|
|ctx|回调函数上下文。|由注册方决定。|

**表 2<a id="table_tft_07"></a>**  非MindSpore回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|dump优化器数据时对应的step。|正整数。|
|args|tft_set_step_args设置的参数。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_save\_ckpt\_handler

**接口功能**

注册框架侧dump回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_save_ckpt_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|临终Checkpoint保存函数，完成保存临终Checkpoint的功能。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_08)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_08"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|dump优化器数据时对应的step。|正整数。|
|save_info|不同优化器参与保存临终遗言时的rank list，其中每个元素是一个字典，字典按照ATTENTION（0）、MOE（1）的索引顺序排列。|[<br>{<br>"type": int,   # 优化器类型 <br>"ranks": list, # 参与对应优化器保存临终遗言时的rank列表<br>},<br>]|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_exit\_handler

**接口功能**

向MindIO TFT注册用户自定义退出方法。

> [!NOTE]说明 
> 目前仅针对MindSpore框架提供了注册退出回调的功能，用户需要自行确保回调函数的安全性；其他框架的退出则由MindIO TFT负责。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_exit_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|完成退出的回调函数。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_09)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_09"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_stop\_handler

**接口功能**

在恢复过程中注册停止训练的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_stop_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|停止训练的回调函数，实现停止训练的功能，并抛出FORCE STOP异常将训练主线程控制权交由装饰器接管。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_19)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_19"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_clean\_handler

**接口功能**

在恢复过程中注册清理残留算子执行的回调函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式<**

```python
mindio_ttp.framework_ttp.tft_register_clean_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|清理残留算子执行的回调函数，完成清理残留算子、底层故障的功能。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_10)。约定该回调函数返回值： <ul><li>0：成功。</li><li>1：失败。</li><li>2：UCE场景且无需重建模型优化器。</li></ul>|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_10"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|is_uce_error|表示该卡是否发生UCE故障。|<ul><li>False：未发生UCE故障。</li><li>True：发生UCE故障。</li></ul>|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_rebuild\_group\_handler

**接口功能**

注册MindIO ARF重新建组的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_rebuild_group_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|MindIO ARF重新建组的回调函数，完成正常节点与重启节点清理旧通信组并重建新通信组的功能。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_11)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_11"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|fault_ranks|故障卡集合。|list。|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_repair\_handler

**接口功能**

注册repair回调函数。

> [!NOTE]说明
>
> - 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。
> - MindIO TFT已在回调函数中对模型优化器中的变量进行重建与覆写，用户在框架中自定义的其他参与计算的变量，需在repair中自行实现对其的重建与覆写。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_repair_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|repair回调函数，完成优化器修复等数据修复功能。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_12)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调上下文。|默认为空。|

**表 1<a id="table_tft_12"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|修复时对应的step。|正整数。|
|need_rebuild|-|修复是否需要重建模型和优化器。|<ul><li>False：无需重建。</li><li>True：需要重建。</li></ul>|
|error_ranks|需要修复的故障卡list。|list。|
|repair_info|修复策略dict，其中优化器类型按照ATTENTION（0）、MOE（1）的关系对应。|{<br>"type": int,   # 优化器类型 <br>"repair_type": Enum,   # 枚举类型取值参见[RepairType](#repairtype) <br>"src": list,    # 优化器修复数据的来源卡列表 <br>"dst": list,   # 优化器修复数据的目的卡列表<br>"rank_list": list, # 修复通信组建立所需要的卡列表<br>}|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_rollback\_handler

**接口功能**

注册rollback回滚函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_rollback_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rollback回调函数，完成数据集回滚等重置操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过设置环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_13)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_13"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|回滚到的step。|正整数。|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_stream\_sync\_handler

**接口功能**

注册同步回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_stream_sync_handler(func: Callable, ctx=None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|同步回调函数，完成训练暂停后同步操作。避免在暂停训练后算子队列有残留算子未执行完。|回调函数，不为空。回调函数无参数，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_zit\_upgrade\_rollback\_handler

**接口功能**

训练框架向Processor注册升级流程回滚的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rollback_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rollback回调函数，完成数据集回滚等重置操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_zit\_upgrade\_repair\_handler

**接口功能**

训练框架向Processor注册升级流程修复的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_repair_handler(func: Callable, ctx = None)
```

**接口参数<a id="section34575883518"></a>**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|repair回调函数，完成优化器修复等数据修复操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_zit\_upgrade\_rebuild\_handler

**接口功能**
训练框架向Processor注册升级流程重建通信组的回调函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rebuild_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rebuild回调函数，完成升级流程重建通信组的修复操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_zit\_downgrade\_rebuild\_handler

**接口功能**

训练框架向Processor注册降级流程重建修复的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_downgrade_rebuild_handler(func: Callable, ctx = None)
```

**接口参数=**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rebuild回调函数，完成降级流程重建修复操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_register\_exception\_handler

**接口功能**

注册异常处理程序。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_exception_handler(fault_pattern: str, fault_type: str, fault_handle: Callable)
```

**接口参数=**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|fault_pattern|必选|异常关键字。用于精确匹配异常类型。|异常信息中的关键字字符串。|
|fault_type|必选|异常类型。用于在捕获对应的异常时，与fault_handle的返回值一起在MindIO上报异常信息|字符串，取值范围如下（详情请参见[ReportState](#reportstate)）:<ul><li>RS_NORMAL</li><li>RS_RETRY</li><li>RS_UCE</li><li>RS_UCE_CORRUPTED</li><li>RS_HCCL_FAILED</li><li>RS_INIT_FINISH</li><li>RS_PREREPAIR_FINISH</li><li>RS_STEP_FINISH</li><li>RS_UNKNOWN</li></ul>|
|fault_handle|必选|异常处理方法。用于接收异常信息字符串，并返回一个字符串。该返回值与fault_type一起在上报异常信息时使用|可执行方法，该方法需要接收异常字符串，并且返回值为字符串。|

**返回值**

无返回值。

### tft\_report\_error

**接口功能**

上报错误类型。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_report_error(error_type: ReportState)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|error_type|必选|上报异常类型，用以决定后续修复流程。|实际错误类型。取值范围请参见[ReportState](#reportstate)。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_wait\_next\_action

**接口功能**

修复期间，训练主线程在装饰器中调用该接口等待从线程完成业务数据修复。

> [!NOTE]说明
> 该接口为阻塞接口，在未获取到下一次action前，会一直阻塞。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_wait_next_action()
```

**接口参数**

无

**返回值**

- 0：成功
- 1：失败

### tft\_get\_repair\_step

**接口功能**

查询修复位置的step值。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_get_repair_step()
```

**接口参数**

无

**返回值**

修复使用的step，返回0表示无效值。

### tft\_get\_repair\_type

**接口功能**

提供给MindSpore调用，用于在stop/clean/repair阶段的回调中查询修复类型。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_get_repair_step()
```

**接口参数**

无

**返回值**

str类型。

- retry：执行UCE修复。
- recover：执行ARF修复。
- dump：执行临终遗言。
- unknow：未找到修复类型。

### tft\_is\_reboot\_node

**接口功能**

MindIO ARF功能流程中，判断当前进程是否为故障后重新拉起的节点，仅支持在tft\_start\_processor接口调用成功后立即调用，且仅支持调用一次。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_is_reboot_node()
```

**接口参数**

无

**返回值**

bool值，表示是否为故障后重新拉起的节点。

### tft\_get\_reboot\_type

**接口功能**

提供给MindSpore调用，在故障重新拉起节点后，训练框架从mindio\_ttp获取节点重启场景类型，进程启动后仅支持调用一次。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_get_reboot_type()
```

**接口参数**

无

**返回值**

str类型。

- arf：代表进程重调度。
- hot switch：代表亚健康热切。

### tft\_reset\_limit\_step

**接口功能**

更新Processor中prelock标记为true，并重置limitStep\_为最大值。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_reset_limit_step()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_set\_dp\_group\_info

**接口功能**

训练框架向Processor注册DP组信息。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_set_dp_group_info(rank: int, dp_rank_list: list)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前rank。|大于或等于0。|
|dp_rank_list|可选|DP组信息。|必空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_report\_load\_ckpt\_step

**接口功能**

使用周期Checkpoint修复时，上报从Checkpoint加载的步数。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_report_load_ckpt_step(step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|step|必选|从Checkpoint加载的步数。|非负整数。|

**返回值**

无

### tft\_register\_decrypt\_handler

**接口功能**

如果用户开启TLS加密，则需要使用该接口注册私钥口令解密函数。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_decrypt_handler(decryptor: Callable)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|decryptor|必选|用户自定义的私钥口令解密函数。|通过 tft_start_controller 和 tft_init_processor 配置TLS加密，并且如果口令为密文，则需注册解密函数。具体配置指导见[导入TLS证书](#导入tls证书)。|

**回调函数参数**

|参数|说明|取值要求|
|--|--|--|
|cipherText|需要解密的私钥口令。|由注册方决定。|

**回调函数返回值**为plainText : str，即解密后的私钥口令。

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

### tft\_notify\_controller\_dump

**接口功能**

提供给MindCluster调用，通知MindIO TFT主动停止训练，执行dump后退出训练。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_dump()
```

**接口参数**

无

**返回值**

- 0：调用成功
- 1：调用失败

### tft\_notify\_controller\_stop\_train

**接口功能**

提供给MindCluster调用，通知MindIO TFT主动停止训练，并告知MindIO TFT发生故障的卡信息。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_stop_train(fault_ranks: dict, stop_type: str = "stop", timeout: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|fault_ranks|必选|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号</li><li>errorType为故障类型：</li><ul><li>0：UCE故障</li><li>1：非UCE故障</li></ul></ul>|
|stop_type|可选|停止训练的类型。|字符串，支持以下两种方式：<ul><li>"stop"：暂停训练，taskabort方式。</li><li>"pause"：暂停训练，非taskabort方式。</li></ul>|
|timeout|可选|暂停训练之后等待MindCluster做下一步通知的超时时间。|非负整数，单位：s。|

**返回值**

- 0：调用成功
- 1：调用失败

### tft\_notify\_controller\_on\_global\_rank

**接口功能**

提供给MindCluster调用，通知MindIO TFT全局的故障卡信息。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_on_global_rank(fault_ranks: dict,time:int=1)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|fault_ranks|必选|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|
|time|可选|根据环境变量设置，决定与MindCluster的修复策略交互的最大时间。|int，取值范围：[1, 3600]，默认值：1，单位：s。|

**返回值**

- 0：调用成功
- 1：调用失败

### tft\_notify\_controller\_prepare\_action

**接口功能**

提供给MindCluster调用，通知MindIO TFT要执行的修复策略。

> [!NOTE]说明
> 该修复策略必须在MindCluster和MindIO TFT协商的可选修复策略范围内。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_prepare_action(action: str, fault_ranks: dict = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|action|必选|通知MindIO TFT亚健康迁移热切动作。|str，支持的修复策略如下：<ul><li>hot switch</li><li>stop switch</li></ul>|
|fault_ranks|可选|发生故障的卡信息。|dict，key为rank号，取值范围0\~100000，value为errtype，取值范围0\~2。|

**返回值**

- 0：调用成功
- 1：调用失败

### tft\_notify\_controller\_change\_strategy

**接口功能**

提供给MindCluster调用，通知MindIO TFT要执行的修复策略。

> [!NOTE]说明
> 该修复策略必须在MindCluster和MindIO TFT协商的可选修复策略范围内。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_change_strategy(strategy: str, params: str = "")
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|strategy|必选|通知MindIO TFT修复策略。|str，支持的修复策略如下：<ul><li>retry</li><li>downgrade </li><li>upgrade</li><li>recover</li><li>dump</li><li>continue</li><li>migration</li><li>exit</li></ul>|
|params|<ul><li>降级训练必选</li><li>其他可选</li></ul>|降级训练参数。|str，默认值：""。|

**返回值**

- 0：调用成功
- 1：调用失败

### tft\_register\_mindx\_callback

**接口功能**

提供给MindCluster调用，向MindIO TFT注册修复流程回调函数接口。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_register_mindx_callback(action: str, func: Callable)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|action|必选|回调函数要注册的动作名。|str，支持的动作名如下：<ul><li>report_fault_ranks</li> <li>report_stop_complete</li><li>report_strategies</li><li>report_result</li></ul>|
|func|必选|要注册的函数。|回调函数，不为空，回调函数入参详情请参见[表1](#table_tft_14) ~ [表4](#table_tft_17)。|

**表 1<a id="table_tft_14"></a>**  action为report\_fault\_ranks时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|

**表 2<a id="table_tft_15"></a>**  action为report\_stop\_complete时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|code|action执行结果。|<ul><li>0：成功。</li><li>400：普通错误。</li><li>401：MindCluster task id不存在。</li><li>402：模型错误。</li><li>403：顺序错误。</li><li>404：Processor未全部准备就绪。</li></ul>|
|msg|训练是否停止消息。|str。|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|

**表 3<a id="table_tft_16"></a>**  action为report\_strategies时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|
|strategy_list|基于当前可用的副本信息，MindIO TFT支持的修复策略列表。|list，支持的修复策略可选值如下（str）：<ul><li>retry：执行UCE修复。</li><li>recover：执行ARF修复。</li><li>dump：执行临终遗言。</li><li>exit：退出。</li></ul>|

**表 4<a id="table_tft_17"></a>**  action为report\_result时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|code|action的执行结果。|<ul><li>0：修复成功。</li><li>405：retry修复失败，支持做recover、dump、exit修复策略。</li><li>406：修复失败，支持做dump或exit修复策略。</li><li>499：修复失败，仅支持exit策略。</li></ul>|
|msg|修复成功或失败的消息。|str|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li> <li>1：非UCE故障。</li></ul></ul>|
|curr_strategy|本次修复策略。|str，支持的修复策略取值范围为表3中的strategy_list。|

**返回值**

- 0：调用成功
- 1：调用失败

### tft\_query\_high\_availability\_switch

**接口功能**

提供给MindCluster调用，实时查询是否开启高可用。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_query_high_availability_switch()
```

**接口参数**

无

**返回值**

bool值，是否开启高可用。

### tft\_can\_do\_uce\_repair

**接口功能**

提供给MindSpore调用，根据L2 Cache触发的UCE故障时间和优化器更新前后时间，判断优化器数据在时间维度是否有被污染的可能，进而返回是否能修复的判断结果。

> [!NOTE]说明
> 该接口仅从时间区间交集上判断优化器数据是否有被污染可能，无法根据内存地址判断。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_can_do_uce_repair(hbm_error_time: int, start_time: int = None, end_time: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|hbm_error_time|必选|L2 Cache触发的UCE故障时间。|int|
|start_time|可选|优化器在本地更新前从device获取的时间。|int|
|end_time|可选|优化器在本地更新后从device获取的时间。|int|

**返回值**

bool值，根据时间交集判断是否可以进行UCE快恢的判断结果。

### tft\_set\_update\_start\_time

**接口功能**

设置优化器更新开始时间，用于判断优化器数据在时间维度是否有被污染可能，进而返回是否能修复的判断结果。

**接口格式**

```python
mindio_ttp.utils.tft_set_update_start_time(start_time: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|start_time|可选|优化器在本地更新前从device获取的时间。|int|

**返回值**

无

### tft\_set\_update\_end\_time

**接口功能**

设置优化器更新结束时间，用于判断优化器数据在时间维度是否有被污染可能，进而返回是否能修复的判断结果。

**接口格式**

```python
mindio_ttp.utils.tft_set_update_end_time(end_time: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|end_time|可选|优化器在本地更新后从device获取的时间。|int|

**返回值**

无

### tft\_pause\_train

**接口功能**

将训练暂停在某一个step。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_pause_train(cur_step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|cur_step|必选|当前训练框架执行的步数。|非负整数。|

**返回值**

无

### OptimizerType

**接口功能**

定义优化器类型枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.OptimizerType
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|OptimizerType|必选|区分优化器类型：<ul><li>ATTENTION：注意力机制类型。</li><li>MOE：MOE场景。</li></ul>|<ul><li>ATTENTION：0</li><li>MOE：1</li></ul>|

**返回值**

无

### Action

**接口功能**

主线程上报异常后的动作类型枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.Action
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|Action|必选|区分主线程上报异常后的动作类型，具体如下：<ul><li>RETRY：修复成功后续训。</li><li>EXIT：退出。</li></ul>|<ul><li>RETRY：0</li><li>EXIT：1</li></ul>|

**返回值**

无

### ReportState

**接口功能**

装饰器上报训练状态枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.ReportState
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|ReportState|必选|区分上报的训练状态类型：<ul><li>RS_NORMAL：正常状态。</li><li>RS_UCE：UCE错误。</li><li>RS_UCE_CORRUPTED：片上内存 MULTI BIT ECC故障。</li><li>RS_HCCL_FAILED：HCCL重计算失败。</li><li>RS_UNKNOWN：其他错误。</li><li>RS_INIT_FINISH：在MindSpore框架中，ARF新启动的节点在训练进程完成初始化后抛出的异常。</li><li>RS_PREREPAIR_FINISH：ARF新启动的节点抛出的异常。</li><li>RS_STEP_FINISH：亚健康热切中step级暂停已经完成抛出的异常。</li></ul>|<ul><li>RS_NORMAL.value：ttp_c2python_api.ReportState_RS_NORMAL。</li><li>RS_UCE.value：ttp_c2python_api.ReportState_RS_UCE。</li><li>RS_UCE_CORRUPTED：ttp_c2python_api.ReportState_RS_UCE_CORRUPTED。</li><li>RS_HCCL_FAILED.value: ttp_c2python_api.ReportState_RS_HCCL_FAILED。</li><li>RS_UNKNOWN.value：ttp_c2python_api.ReportState_RS_UNKNOWN。</li><li>RS_INIT_FINISH：ttp_c2python_api.ReportState_RS_INIT_FINISH。</li><li>RS_PREREPAIR_FINISH.value：ttp_c2python_api.ReportState_RS_PREREPAIR_FINISH。</li><li>RS_STEP_FINISH：ttp_c2python_api.ReportState_RS_STEP_FINISH。</li></ul>|

**返回值**

无

### RepairType

**接口功能**

定义修复类型枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.RepairType
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|RepairType|必选|区分修复类型：<ul><li>RT_SEND：备份卡发送数据。</li><li>RT_UCE_HIGHLEVEL：故障卡需要优化器和模型重建。</li><li>RT_UCE_LOWLEVEL：故障卡不需要优化器和模型重建。</li><li>RT_ROLLBACK：回滚数据集。</li><li>RT_RECV_REPAIR：ARF新拉起卡接收数据。</li><li>RT_LOAD_CKPT：周期Checkpoint数据修复。</li><li>RT_LOAD_REBUILD：重建模型优化器周期Checkpoint数据修复。</li></ul>|<ul><li>RT_SEND.value：ttp_c2python_api.RepairType_RT_SEND。</li><li>RT_UCE_HIGHLEVEL.value：ttp_c2python_api.RepairType_RT_UCE_HIGHLEVEL。</li><li>RT_UCE_LOWLEVEL.value：ttp_c2python_api.RepairType_RT_UCE_LOWLEVEL。</li><li>RT_ROLLBACK.value：ttp_c2python_api.RepairType_RT_ROLLBACK。</li><li>RT_RECV_REPAIR.value：ttp_c2python_api.RepairType_RT_RECV_REPAIR。</li><li>RT_LOAD_CKPT.value：ttp_c2python_api.RepairType_RT_LOAD_CKPT。</li><li>RT_LOAD_REBUILD.value：ttp_c2python_api.RepairType_RT_LOAD_REBUILD。</li></ul>|

**返回值**

无

## 附录

### 环境变量

> [!NOTE]说明
> 加粗显示的环境变量为常用环境变量。

|参数名称|参数说明|取值范围|缺省值|
|--|--|--|--|
|**TTP_LOG_PATH**|MindIO TFT日志路径。禁止配置软链接，日志文件名补充为ttp_log.log，建议日志路径中包含日期时间，避免多次训练记录在同一个日志中，造成循环覆写。推荐在训练启动脚本中按如下方式配置日志路径： <br> `date_time=\$(date +%Y-%m-%d-%H_%M_%S)` <br> `export TTP_LOG_PATH=logs/\${date_time}` <br>当使用共享存储时，建议按照节点配置日志路径：<br>`export TTP_LOG_PATH=logs/\${nodeId}`|文件夹路径。|logs|
|**TTP_LOG_LEVEL**|MindIO TFT日志等级。<ul><li>DEBUG：细节信息，仅当诊断问题时适用。</li><li>INFO：确认程序按预期运行。</li><li>WARNING：表明有已经或即将发生的意外。程序仍按预期进行。</li><li>ERROR：由于严重的问题，程序的某些功能已经不能正常执行。</li></ul>|<ul><li>DEBUG</li><li>INFO</li><li>WARNING</li><li>ERROR|INFO</li></ul>|
|TTP_LOG_MODE|MindIO TFT日志模式。<ul><li>ONLY_ONE：所有MindIO TFT进程写一个日志。</li><li>PER_PROC：每个MindIO TFT进程写独立日志，日志文件路径为 {TTP_LOG_PATH}/ttp_log.log.{pid}。</li></ul>|<ul><li>ONLY_ONE</li><li>PER_PROC（若非指定ONLY_ONE，则默认为PER_PROC）</li></ul>|PER_PROC|
|TTP_LOG_STDOUT|MindIO TFT日志记录方式。<ul><li>0：将MindIO TFT运行日志记录到对应的日志文件中。</li><li>1：直接打印MindIO TFT运行日志，不在本地存储。</li></ul>|<ul><li>0</li><li>1</li></ul>|0|
|MASTER_ADDR|训练主节点IP地址或域名。|合法的IPv4或IPv6地址或域名。|-|
|MASTER_PORT|训练主节点通信端口，端口可配。|[1024, 65535]|-|
|TTP_RETRY_TIMES|Processor TCP（Transmission Control Protocol）建链尝试次数。|[1, 300]|10|
|MINDIO_WAIT_MINDX_TIME|Controller等待MindCluster响应的最大时间，单位：s。|[1, 3600]|30|
|TTP_ACCLINK_CHECK_PERIOD_HOURS|开启TLS认证后，MindIO TFT检查证书有效性的周期，单位：h。|[24, 720]|168|
|TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS|开启TLS认证后，MindIO TFT检查证书过期日提前告警的时长，单位：天。需满足证书过期提前告警时长不小于巡检周期，保证及时发现证书过期风险并告警。|[7, 180]，且需满足TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS * 24 ≥ TTP_ACCLINK_CHECK_PERIOD_HOURS。|30|
|TTP_NORMAL_ACTION_TIME_LIMIT|故障恢复流程中，执行rebuild/repair/rollback回调函数的超时时间，单位：s。|[30, 1800]|180|
|MINDIO_FOR_MINDSPORE|表示是否启用MindSpore开关，传入True（不区分大小写）或1时，开启MindSpore开关，其他值关闭MindSpore开关。|<ul><li>True（不区分大小写）或1：启用MindSpore。</li><li>其他：关闭MindSpore。</li></ul>|False|
|MINDX_TASK_ID|MindIO ARF特性使用，MindCluster任务ID，由ClusterD配置，无需用户干预。|字符串。|-|
|TORCHELASTIC_USE_AGENT_STORE|PyTorch环境变量，控制创建TCP Store Server还是Client，MindIO TFT在临终Checkpoint保存且Torch Agent TCP Store Server连接失败场景下使用。|<ul><li>True：创建Client。</li><li>False：创建Server。</li></ul>|-|
|TTP_STOP_CLEAN_BEFORE_DUMP|MindIO TFT特性使用，控制MindIO TTP在保存临终Checkpoint前是否做stop&clean操作。|<ul><li>0：关闭临终前stop&clean操作。</li><li>1：启用临终前stop&clean操作。</li></ul>|0|

### 设置用户有效期

为保证用户的安全性，应设置用户的有效期，使用系统命令 **chage** 来设置用户的有效期。

命令为：

```bash
chage [-m mindays] [-M maxdays] [-d lastday] [-I inactive] [-E expiredate] [-W warndays] user
```

相关参数请参见[表1](#table_tft_18)。

**表 1<a id="table_tft_18"></a>**  设置用户有效期

|参数|参数说明|
|--|--|
|-d<br>--lastday|上一次更改的日期。|
|-E<br>--expiredate|用户到期的日期。超过该日期，此用户将不可用。|
|-h<br>--help|显示命令帮助信息。|
|-i<br>--iso8601|更改用户密码的过期日期并以YYYY-MM-DD格式显示。|
|-I<br>--inactive|停滞时期。过期指定天数后，设定密码为失效状态。|
|-l<br>--list|列出当前的设置。由非特权用户来确定口令或账户何时过期。|
|-m<br>--mindays|口令可更改的最小天数。设置为“0”表示任何时候都可以更改口令。|
|-M<br>--maxdays|口令保持有效的最大天数。设置为“-1”表示可删除这项口令的检测。设置为“99999”，表示无限期。|
|-R<br>--root|将命令执行的根目录设置为指定目录。|
|-W<br>--warndays|用户口令到期前，提前收到警告信息的天数。|

> [!NOTE]说明
>
> - 日期格式为YYYY-MM-DD，如 **chage -E 2017-12-01 _test_** 表示用户 **_test_** 的口令在2017年12月1日过期。
> - user必须填写，填写时请替换为具体用户，默认为root用户。
> - 账号口令应该定期更新，否则容易导致安全风险。

举例说明：修改用户 **_test_** 的有效期为90天。

```bash
chage -M 90 test
```

### 口令复杂度要求

口令至少满足如下要求：

1. 口令长度至少8个字符。
2. 口令必须包含如下至少两种字符的组合：
    - 一个小写字母
    - 一个大写字母
    - 一个数字
    - 一个特殊字符：\`\~!@\#$%^&\*\(\)-\_=+\\|[\{\}];:'",<.\>/?和空格

3. 口令不能和账号一样。

### 账户一览表

|用户|描述|初始密码|密码修改方法|
|--|--|--|--|
| *{MindIO-install-user}* |MindIO TFT安装用户。|用户自定义。|使用 **passwd** 命令修改。|
