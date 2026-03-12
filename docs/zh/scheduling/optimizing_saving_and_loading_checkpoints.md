# Checkpoint保存与加载优化

## 产品描述

**产品介绍**

MindCluster MindIO Async Checkpoint Persistence（下文简称MindIO ACP）加速大模型Checkpoint功能主要针对大模型训练中的Checkpoint的保存及加载进行加速，Checkpoint的数据先写入训练服务器的内存系统中，再异步写入后端的可靠性存储设备中。本文档主要介绍纵向加速部分，包含Checkpoint在本系统中的写入及读取过程。

**产品价值**

LLM（Large Language Model，大语言模型）是全球当前科技界竞争的焦点，LLM模型的训练往往需要长达数十天、甚至数月。Checkpoint是模型中断训练后恢复的关键点，Checkpoint的密集程度、保存和恢复的性能较为关键，它可以提高训练系统的有效吞吐率。MindIO ACP针对Checkpoint的加速方案，支持昇腾产品在LLM模型领域扩展市场空间。

该方案提升昇腾平台上LLM模型的训练吞吐量，性能超越[Microsoft Azure Nebula方案](https://learn.microsoft.com/zh-cn/azure/machine-learning/reference-checkpoint-performance-for-large-models?view=azureml-api-2&tabs=PYTORCH)。

**MindIO ACP架构**

![](../figures/scheduling/mindio_acp架构.png)

MindIO ACP加速LLM Checkpoint保存和加载的4个关键点如下：

-   异步持久化。训练框架通过mindio\_acp的save/load接口或MindSpore框架将Checkpoint保存到MindIO ACP后，直接返回继续训练，该时间为秒级；MindIO ACP会异步将Checkpoint写入持久化的分布式存储，该过程为分钟级。
-   高性能MemFS（Memory File System，内存文件系统）。MindIO ACP为实现Checkpoint极速写入，实现了全用户态的以内存为介质的文件系统；消除各种标准文件系统的系统调用和用户态到内核态的内存拷贝。
-   高效Checkpoint保存和加载。MindIO ACP为实现Checkpoint极速写入和恢复，研发了高效Checkpoint保存、加载方式。
-   MindIO ACP具备自动容错能力。当MindIO ACP服务异常导致数据读写失败、超时等异常时，能自动切换到原生数据存储方式，保证业务不中断。

    > [!CAUTION]注意
    > MindIO ACP仅保存训练过程中的Checkpoint数据，暂不支持敏感数据的保存和处理。若涉及敏感数据存储，请在前序流程完成相关脱敏操作，避免造成信息安全问题。




## 安装部署


### 安装前必读


#### 免责声明

本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。

用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。



#### 约束限制

-   训练[故障快速恢复](./fault_recovery_acceleration.md)框架正在向MindIO ACP保存Checkpoint时，如果遇到Checkpoint保存失败，当前正在保存的Checkpoint不能作为训练恢复点，训练框架需要从上一次完整的Checkpoint点进行恢复。
-   在训练过程中发生MindIO ACP故障，已经下发的业务，MindIO ACP SDK会重试3次连接，3次都失败则对接原生存储方式，重试最长等待60s；在训练开始前发生MindIO ACP故障，MindIO ACP SDK则会跳过对接MindIO ACP，Checkpoint的数据直接对接原生数据存储方式。
-   本特性与MindIO TFT[故障快速恢复](./fault_recovery_acceleration.md)特性不兼容。
-   本特性不配套MindSpore 2.7.0之前的版本，功能无法使用。



### 安装前准备


#### 组网规划

**图 1**  部署逻辑示意图  
![](../figures/scheduling/部署逻辑示意图acp.png "部署逻辑示意图")

深度学习平台与训练任务相关的节点有计算节点和存储节点。各类节点主要功能如下：

-   计算节点：实际执行训练、推理任务的节点，MindIO ACP仅部署在计算节点。
-   存储节点：存储平台数据和用户数据，如平台日志、用户上传的数据集、训练脚本、训练输出的模型等。

网络平面划分为：

-   业务面：用于管理集群业务。管理节点和计算节点之间连接。
-   存储面：用于访问存储节点。管理节点和计算节点连接到存储节点。
-   参数面：用于分布式训练时，训练节点之间的参数交换和连接。

> [!NOTE]说明
> - 逻辑部署示意图展示深度学习平台的完整示意图，MindIO ACP作为计算节点上部署的一个组件，不涉及管理节点和存储节点的安装部署。
> - MindIO ACP是单节点内存缓存系统，训练Checkpoint数据通过共享内存方式访问MindIO ACP，不涉及网络平面划分。



#### 环境要求

**硬件环境**

安装前，需要检查以下硬件配置，如[表1](#table_acp_01)所示。

**表 1 <a id="table_acp_01"></a>**  硬件环境

|类型|配置参考|
|--|--|
|服务器（单机场景）|Atlas 800 训练服务器（型号：9000）|
|服务器（集群场景）|<ul><li>计算节点：Atlas 800 训练服务器（型号：9000）</li> <li>存储节点：存储服务器</li></ul>|
|内存|<ul><li>推荐配置：≥64GB</li> <li>最低配置：≥32GB</li></ul>|
|磁盘空间|≥1TB <br> 磁盘空间规划请参见[表3](#table_acp_03)|
|网络|<ul><li>带外管理（BMC）：≥1Gbit/s <li>带内管理（SSH）：≥1Gbit/s <li>业务面：≥10Gbit/s <li>存储面：≥25Gbit/s <li>参数面：100Gbit/s </ul>|


**软件环境**

安装前，需要完成以下环境的安装，如[表2](#table_acp_02)所示。

**表 2 <a id="table_acp_02"></a>**  软件环境

|软件|版本|安装位置|获取方式|
|--|--|--|--|
|操作系统|<ul> <li>CentOS 7.6 Arm<li>CentOS 7.6 x86<li>openEuler 20.03 Arm<li>openEuler 20.03 x86<li>openEuler 22.03 Arm<li>openEuler 22.03 x86<li>Ubuntu 20.04 Arm<li>Ubuntu 20.04 x86<li>Ubuntu 18.04.5 Arm<li>Ubuntu 18.04.5 x86<li>Ubuntu 18.04.1 Arm<li>Ubuntu 18.04.1 x86<li>Kylin V10 SP2 Arm<li>Kylin V10 SP2 x86<li>UOS20 1020e Arm </ul>|所有节点|-|
|Python|3.7或更高版本|计算节点|用户安装|
|Torch|2.7.1|计算节点|用户安装|
|MindSpore|2.7.0或更高版本|计算节点|用户安装|


**操作系统磁盘分区**

操作系统磁盘分区推荐如[表3](#table_acp_03)所示。

**表 3 <a id="table_acp_03"></a>**  磁盘分区

|分区|说明|大小|bootable flag|
|--|--|--|--|
|/boot|启动分区|500MB|on|
|/var|软件运行所产生的数据存放分区，如日志、缓存等|>300GB|off|
|/|主分区|>300GB|off|



#### 准备软件包

**下载软件包**

下载本软件即表示您同意[华为企业业务最终用户许可协议（EULA）](https://e.huawei.com/cn/about/eula)的条款和条件。

**表 1**  软件下载

|组件名称|软件包|获取地址|
|--|--|--|
|MindIO ACP|内存缓存系统软件包|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases)|


**软件数字签名验证**

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请勿使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[http://support.huawei.com/carrier/digitalSignatureAction](http://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)



### 在计算节点安装MindIO ACP SDK

通过使用MindIO ACP SDK对接Torch和MindSpore，加速Torch和MindSpore训练Checkpoint save和load操作。

**操作步骤**

1.  以安装用户 *{MindIO-install-user}* 登录安装节点。

    >[!NOTE]说明
    >安装用户设置的口令需符合口令复杂度要求（请参见[口令复杂度要求](#口令复杂度要求)）。密码有效期为90天，您可以在“/etc/login.defs“文件中修改有效期的天数，或者通过 **chage** 命令来设置用户的有效期，详情请参见[设置用户有效期](#设置用户有效期)。

2.  将内存缓存系统软件包上传至设备中安装用户有权限读写的路径下。

    > [!NOTE]说明
    > - 内存缓存系统软件包以获取的实际包名为准。
    > - 如果Python环境是共享目录，则在任一计算节点上传即可，否则所有计算节点都需要上传安装包。

3.  进入软件包上传路径，解压内存缓存系统软件包。

    ```
    unzip Ascend-mindxdl-mindio_{version}_linux-{arch}.zip
    ```

    **表 1**  解压后文件

    |文件|说明|
    |--|--|
    |mindio_acp-*{mindio_acp_version}*-py3-none-linux_*{arch}*.whl|MindIO ACP安装包。|
    |mindio_ttp-*{mindio_ttp_version}*-py3-none-linux_*{arch}*.whl|MindIO TFT安装包。|


4.  进入上传路径，安装MindIO ACP SDK。

    ```
    pip3 install mindio_acp-{mindio_acp_version}-py3-none-linux_{arch}.whl --force-reinstall
    ```

    -   首次安装MindIO ACP SDK回显如下，表示安装成功。

        ```
        Processing ./mindio_acp-{mindio_acp_version}-py3-none-linux_{arch}.whl
        Installing collected packages: mindio_acp
        Successfully installed mindio_acp-{version}
        ```

    -   非首次安装MindIO ACP SDK回显如下，表示安装成功。

        ```
        Processing ./mindio_acp-{mindio_acp_version}-py3-none-linux_{arch}.whl
         Installing collected packages: mindio_acp
           Attempting uninstall: mindio_acp
             Found existing installation: mindio_acp{version}
             Uninstalling mindio_acp{version}:
               Successfully uninstalled mindio_acp{version}
         Successfully installed mindio_acp{version}
        ```

5.  将软件安装目录内的可执行文件和代码脚本权限更改为550，避免出现非法篡改。

    ```
    chmod -R 550 {MindIO ACP SDK安装目录}
    ```


### 卸载MindIO ACP SDK

**操作步骤**

1.  将软件安装目录内的可执行文件和代码脚本权限更改为750。

    ```
    chmod -R 750 {MindIO ACP SDK安装目录}
    ```

2.  卸载MindIO ACP SDK。

    ```
    pip3 uninstall mindio_acp
    ```


## 使用指导



### 概述

> [!NOTE]说明
> - MindIO ACP SDK端支持宿主机和容器内部署。
> - 容器场景的镜像制作、镜像部署、镜像安全加固等由用户保证。
> - 只支持DeepSpeed框架、X1框架、MindSpeed-LLM、K8s的固定版本。
> - 在使用MindIO ACP服务时，启动训练任务的用户需要和启动MindIO ACP守护进程的用户属于同一个主组。

安装MindIO ACP SDK之后，为了使用MindIO ACP的缓存加速能力，将训练模型中使用到Python文件中的Torch的load/save函数，替换为MindIO ACP SDK的load/save函数。

-   支持将同一份数据保存到多个路径，将训练模型中循环保存同一份数据的torch.save函数，替换为MindIO ACP SDK的mindio\_acp.multi\_save函数。
-   MindIO ACP SDK提供 **register\_checker(callback, check\_dict, user\_context, timeout\_sec)** 接口，支持将需要观察的文件夹和文件夹下的普通文件个数作为check\_dict的元素注册到MindIO ACP，MindIO ACP会在timeout\_sec时间内检查这些文件夹下的文件个数，并检查其与check\_dict元素指定的文件个数是否相同，通过注册的callback函数回调应用程序，user\_context为callback函数的第二个参数，支持用户设置callback函数中需要调用的参数；timeout\_sec为注册事件超时时间，当超过超时时间仍然检查到不符合要求，则会在回调函数中报告错误。用户可以根据检查结果处理后续业务逻辑。



### Torch对接DeepSpeed框架

1.  使用业务用户登录到计算节点。

    > [!NOTE]说明
    > 业务用户不是 *{MindIO-install-user}*、HwHiAiUser、hwMindX用户，由用户根据实际情况决定。

2.  进入DeepSpeed安装目录。

    ```
    cd {deepspeed安装目录}/runtime 
    ```

3.  <a id="step_acp_li001"></a>修改engine.py文件。
    1.  打开engine.py文件。

        ```
        vim engine.py
        ```

    2.  <a id="step_acp_li002"></a>按“i”进入编辑模式，修改如下内容。
        -   在文件首行加入以下内容。

            ```
            import mindio_acp
            ```

        -   将torch.load函数替换为mindio\_acp.load函数。

            替换前：

            ```
            optim_checkpoint = torch.load(optim_load_path,
                                          map_location=torch.device('cpu'))
            ```

            替换后：

            ```
            optim_checkpoint = mindio_acp.load(optim_load_path, map_location='cpu')
            ```

        -   将torch.save函数替换为mindio\_acp.save函数。

            替换前：

            ```
            torch.save(state, save_path)
            ```

            替换后：

            ```
            mindio_acp.save(state, save_path)
            ```

        -   将包含torch.save函数的with open语句整体替换为mindio\_acp.save函数。

            替换前：

            ```
            with open(self._get_optimizer_ckpt_name(save_dir, tag, expp_rank), 'wb') as fd:
                torch.save(optimizer_state, fd)
                fd.flush()
            ```

            替换后：

            ```
            mindio_acp.save(optimizer_state, self._get_optimizer_ckpt_name(save_dir, tag, expp_rank))
            ```

        -   替换DeepSpeedEngine.\_get\_expert\_ckpt\_name函数。

            替换前：

            ```
                            expert_state_dict = torch.load(DeepSpeedEngine._get_expert_ckpt_name(
                                checkpoint_path,
                                -1, # -1 means ignore layer_id
                                global_expert_id,
                                tag,
                                mpu),
                                map_location=torch.device('cpu'))
            ```

            替换后：

            ```
                            expert_state_dict = mindio_acp.load(DeepSpeedEngine._get_expert_ckpt_name(
                                checkpoint_path,
                                -1, # -1 means ignore layer_id
                                global_expert_id,
                                tag,
                                mpu),
                                map_location='cpu')
            ```

    3.  <a id="step_acp_li003"></a>按“Esc”键，输入 **:wq!** ，按“Enter”保存并退出编辑。

4.  修改module.py文件。
    1.  打开module.py文件。

        ```
        vim pipe/module.py
        ```

    2.  替换torch.save和torch.load，替换方式参见步骤[3.b](#step_acp_li002)  \~ 步骤[3.c](#step_acp_li003)。

5.  <a id="step_acp_li004"></a>修改state\_dict\_factory.py文件。
    1.  打开state\_dict\_factory.py文件。

        ```
        vim state_dict_factory.py
        ```

    2.  替换torch.save和torch.load，替换方式参见步骤[3.b](#step_acp_li002)  \~步骤  [3.c](#step_acp_li003)。

6.  完成步骤[3](#step_acp_li001)  \~ 步骤[5](#step_acp_li004)的.py文件修改，DeepSpeed即可使用MindIO ACP服务。



### Torch对接X1框架

1.  登录到计算节点。
2.  进入X1安装目录。

    ```
    cd {X1安装目录}/Megatron-LM/megatron
    ```

3.  修改checkpointing.py文件。
    1.  打开checkpointing.py文件。

        ```
        vim checkpointing.py
        ```

    2.  按“i”进入编辑模式，修改如下内容。
        -   在文件首行加入以下内容。

            ```
            import mindio_acp
            ```

        -   将torch.load函数替换为mindio\_acp.load函数。

            替换前：

            ```
            optim_checkpoint = torch.load(optim_load_path,
                                          map_location=torch.device('cpu'))
            ```

            替换后：

            ```
            optim_checkpoint = mindio_acp.load(optim_load_path, map_location='cpu')
            ```

        -   将torch.save函数替换为mindio\_acp.save函数。

            替换前：

            ```
            torch.save(state, save_path)
            ```

            替换后：

            ```
            mindio_acp.save(state, save_path)
            ```

        -   将包含torch.save函数的with open语句整体替换为mindio\_acp.save函数。

            替换前：

            ```
            with open(self._get_optimizer_ckpt_name(save_dir, tag, expp_rank), 'wb') as fd:
                torch.save(optimizer_state, fd)
                fd.flush()
            ```

            替换后：

            ```
            mindio_acp.save(optimizer_state, self._get_optimizer_ckpt_name(save_dir, tag, expp_rank))
            ```

    3.  按“Esc”键，输入 **:wq!** ，按“Enter”保存并退出编辑。



### Torch对接MindSpeed-LLM框架

**前提条件**

-   使用前请先了解MindIO ACP特性的[约束限制](#约束限制)章节。
-   MindSpeed-LLM框架准备请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.2.0)。匹配的Megatron-LM版本为 **core\_v0.12.1**。

> [!NOTE]说明
> 本次发布包配套MindSpeed-LLM的 **2.2.0** 分支，环境、代码、数据集准备请用户参考MindSpeed-LLM仓库的相关指导说明，并确保其安全性。

**操作步骤**

1.  使用业务用户登录到计算节点。

    > [!NOTE]说明
    > 业务用户不是 *{MindIO-install-user}*、HwHiAiUser、hwMindX用户，由用户根据实际情况决定。

2.  进入MindSpeed-LLM安装目录。

    ```
    cd MindSpeed-LLM/
    ```

3.  <a id="step_acp_li005"></a>修改pretrain\_gpt.py文件。
    1.  打开pretrain\_gpt.py文件。

        ```
        vim pretrain_gpt.py
        ```

    2.  按“i”进入编辑模式，在文件头部找到from mindspeed\_llm import megatron\_adaptor，换行增加 **import mindio_acp**。

        ```
        from mindspeed_llm import megatron_adaptor
        import mindio_acp
        ```

    3.  按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

4.  <a id="step_acp_li006"></a>编辑预训练脚本（仅供参考）。

    此处以编辑“examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh”脚本为例。

    1.  打开“examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh”脚本。

        ```
        vim examples/mcore/llama2/pretrain_llama2_7b_ptd.sh
        ```

    2.  按“i”进入编辑模式，在脚本中增加如下内容以开启周期性Checkpoint加速功能。
        ```
        export MINDIO_AUTO_PATCH_MEGATRON=true
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:$LD_LIBRARY_PATH
        source /usr/local/Ascend/cann/set_env.sh
        ```
        
        修改后的pretrain_llama2_7b_ptd.sh脚本示例如下：

        ```
        #!/bin/bash
        
        export CUDA_DEVICE_MAX_CONNECTIONS=1
        export PYTORCH_NPU_ALLOC_CONF=expandable_segments:True
        
        export MINDIO_AUTO_PATCH_MEGATRON=true
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:$LD_LIBRARY_PATH
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
            --bf16
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

        周期性Checkpoint加速功能相关参数说明如下：

        -   **MINDIO\_AUTO\_PATCH\_MEGATRON**：MindIO ACP框架自动patch Megatron的源码，用于开启加速周期性Checkpoint特性。
        -   **GLOO\_SOCKET\_IFNAME**：根据主节点高速网卡实际情况进行配置。
        -   **LD\_LIBRARY\_PATH**：CANN包驱动的so库地址，请根据CANN实际的安装路径进行修改。
        -   **set\_env.sh文件路径**：请根据CANN实际的安装路径进行修改。

    3.  按“Esc”键，输入 **:wq!** ，按“Enter”保存并退出编辑。

5.  完成步骤[3](#step_acp_li005)  \~ 步骤[4](#step_acp_li006)的.py文件修改，MindSpeed-LLM即可使用MindIO ACP加速周期性Checkpoint特性。



### 对接K8s

在容器中使用MindIO ACP加速服务时，需要将SDK安装到对应的容器中。

1.  修改创建Pod的yaml文件，下面以“/home/testuser/mygpt.yaml”文件为例，增加映射卷配置。
    1.  打开mygpt.yaml文件。

        ```
        vim /home/testuser/mygpt.yaml
        ```

    2.  按“i”进入编辑模式，修改mygpt.yaml文件。

        > [!NOTE]说明
        > - 如果volumeMounts和volumes不存在，直接在文件中添加全部内容。
        > - 如果volumeMounts和volumes已存在，只需在volumeMounts和volumes内部添加其后面的内容。

        -   （可选）如果环境中[使用了DPC访问存储](#可选使用dpc文件访问存储加速checkpoint加载)，增加卷在容器中映射路径，内容如下：

            ```
            volumeMounts:
                - mountPath: /opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so
                  name: mindio-dpc-nds
                  readOnly: false
            ```

            > [!NOTE]说明
            > “/opt/oceanstor/dataturbo/sdk/lib/libdpc\_nds.so”不可随意更改。

        -   （可选）如果环境中[使用了DPC访问存储](#可选使用dpc文件访问存储加速checkpoint加载)，增加宿主机需要映射的卷声明，增加内容如下：

            ```
            volumes:
              - name: mindio-dpc-nds
                hostPath:
                  path: /opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so
                  type: File
            ```

    3.  按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

2.  使用修改后的yaml文件，创建Pod。

    ```
    kubectl apply -f mygpt.yaml
    ```

3.  登录到创建好的Pod，以命名空间“test-mindio”下名称为“mygptdd”的Pod为例。

    ```
    kubectl exec -it mygptdd -n test-mindio /bin/bash
    ```

4.  将MindIO ACP SDK上传到Pod中，并参见[在计算节点安装MindIO ACP SDK](#在计算节点安装mindio-acp-sdk)完成SDK安装。



### Checkpoint文件格式转换示例（Torch）

对于使用PyTorch框架的用户，在大模型训练结束后，Checkpoint文件需要用于推理。这里举例说明，如何将MindIO ACP保存的Checkpoint文件转换成Torch原生格式的文件。

> [!NOTE]说明
> - **load\_dir**：替换为真实的Checkpoint保存目录。
> - **new\_dir**：替换为Checkpoint转换后新保存的目录，建议为空目录。
> - **iteration**：指定转换这个iteration迭代周期的所有Checkpoint文件，会和 **load\_dir** 进行拼接。

```
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import os
import mindio_acp


def main():
    load_dir = ""  # Replace with the actual checkpoint directory path
    new_dir = ""  # Replace with the actual new directory path
    iteration = 2000  # Replace with the actual iteration number

    directory = 'iter_{:07d}'.format(iteration)
    common_path = os.path.join(load_dir, directory)

    if not os.path.exists(common_path):
        print(f"Source directory {common_path} does not exist.")
        return

    if not os.path.exists(new_dir):
        os.makedirs(new_dir)

    for root, _, files in os.walk(common_path):
        # Compute the relative path and target directory
        relative_path = os.path.relpath(root, common_path)
        target_dir = os.path.join(new_dir, relative_path)

        # Create directories in the target directory
        if not os.path.exists(target_dir):
            os.makedirs(target_dir)

        # Convert all files in the current directory
        for file in files:
            src_file = os.path.join(root, file)
            dst_file = os.path.join(target_dir, file)
            res = mindio_acp.convert(src_file, dst_file)
            print(f"Convert {src_file} to {dst_file}, result: {res}")


if __name__ == '__main__':
    main()
```


## 安全管理与加固



### 安全管理

> [!NOTE]说明
> MindIO ACP暂不支持公有云场景、多租户场景使用，不支持公网直接访问系统。

**防病毒软件例行检查**

定期开展对集群的防病毒扫描，防病毒例行检查会帮助集群免受病毒、恶意代码、间谍软件以及程序侵害，降低系统瘫痪、信息安全问题等风险。可以使用业界主流防病毒软件进行防病毒检查。

**日志管理**

日志管理需要关注以下两点。

-   检查系统是否可以限制单个日志文件的大小。
-   检查日志空间占满后，是否存在机制进行清理。

**漏洞/功能问题修复**

为保证生产环境的安全，降低被攻击的风险，需要定期查看开源社区修复的以下漏洞/功能问题。

-   操作系统漏洞/功能问题。
-   其他相关组件漏洞/功能问题。



### 安全加固



#### 加固须知

本文中列出的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施，必要时可参考业界优秀加固方案和安全专家的建议。


#### 风险提醒

Checkpoint序列化使用了Python自带的pickle组件，必须确保非授权用户没有存储目录及上层目录的写权限，否则可能造成Checkpoint被篡改引起pickle反序列化注入的风险。


#### 操作系统安全加固

**防火墙配置**

操作系统安装后，若配置普通用户，可以通过在“/etc/login.defs”文件中新增“ALWAYS\_SET\_PATH=yes”配置，防止越权操作。此外，为了防止使用“su”命令切换用户时，将当前用户环境变量带入其他环境造成提权，请使用 **su - [user]** 命令进行用户切换，同时在服务器配置文件“/etc/default/su”中增加配置参数“ALWAYS\_SET\_PATH=yes”防止提权。

**设置umask**

建议用户将服务器的umask设置为027 \~ 777以限制文件权限。

以设置umask为027为例，具体操作如下。

1.  以root用户登录服务器，编辑“/etc/profile”文件。

    ```
    vim /etc/profile
    ```

2.  在“/etc/profile”文件末尾加上 **umask 027**，保存并退出。
3.  执行如下命令使配置生效。

    ```
    source /etc/profile
    ```

**无属主文件安全加固**

用户可以执行 **find / -nouser -nogroup** 命令，查找容器内或物理机上的无属主文件。根据文件的UID和GID创建相应的用户和用户组，或者修改已有用户的UID、用户组的GID来适配，赋予文件属主，避免无属主文件给系统带来安全隐患。

**端口扫描**

用户需要关注全网侦听的端口和非必要端口，如果发现非必要端口则应立即关闭。建议用户关闭不安全的服务，如Telnet、FTP等，以提升系统安全性。具体操作方法可参考所使用操作系统的官方文档。

**防DoS攻击**

用户可以根据IP地址限制与服务器的连接的速率对系统进行防DoS攻击，方法包括但不限于利用Linux系统自带Iptables防火墙进行预防、优化sysctl参数等。具体使用方法，用户可自行查阅相关资料。

**SSH加固**

由于root用户拥有最高权限，出于安全目的，建议取消root用户SSH远程登录服务器的权限，以提升系统安全性。具体操作步骤如下：

1.  登录安装MindIO ACP组件的节点。
2.  打开“/etc/ssh/sshd\_config”文件。

    ```
    vim /etc/ssh/sshd_config
    ```

3.  按“i”进入编辑模式，找到“PermitRootLogin”配置项并将其值设置为“no”。

    ```
    PermitRootLogin no
    ```

4.  按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。
5.  执行命令使配置生效。

    ```
    systemctl restart sshd
    ```


## API接口参考



### initialize接口

**接口功能**

初始化MindIO ACP  Client。

**接口格式**

```
mindio_acp.initialize(server_info: Dict[str, str] = None) -> int
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|server_info|可选|自启动的Server进程需要配置参数信息。若不传入该参数，则全部使用默认值。|有效参数集合或None。|


**表 1**  server\_info参数说明

|参数key|默认参数value|是否必选|说明|取值范围|
|--|--|--|--|--|
|'memfs.data_block_pool_capacity_in_gb'|'128'|可选|MindIO ACP文件系统内存分配大小，单位：GB，根据服务器内存大小来配置，建议不超过系统总内存的25%。|[1, 1024]|
|'memfs.data_block_size_in_mb'|'128'|可选|文件数据块分配最小粒度，单位：MB，根据使用场景中大多数文件的size决定配置，建议平均每个文件的数据块大小不超过128MB。|[1, 1024]|
|'memfs.write.parallel.enabled'|'true'|可选|MindIO ACP并发读写性能优化开关配置，用户需结合业务数据模型特征决定是否打开本配置。|<ul><li>false：关闭<li>true：开启 </ul>|
|'memfs.write.parallel.thread_num'|'16'|可选|MindIO ACP并发读写性能优化并发数。|[2, 96]|
|'memfs.write.parallel.slice_in_mb'|'16'|可选|MindIO ACP并发写性能优化数据切分粒度，单位：MB。|[1, 1024]|
|'background.backup.thread_num'|'32'|可选|备份线程数量。|[1, 256]|


> [!NOTE]说明
> mindio\_acp.initialize如果不传入server\_info参数，则按照表中默认参数启动Server。

**使用样例1**

```
>>> # Initialize with default param
>>> mindio_acp.initialize()
```

**使用样例2**

```
>>> # Initialize with server_info
>>> server_info = {
        'memfs.data_block_pool_capacity_in_gb': '200',
    }
>>> mindio_acp.initialize(server_info=server_info)
```

**返回值**

-   0：成功
-   -1：失败



### save接口

**接口功能**

将数据保存到指定的路径下。

**接口格式**

```
mindio_acp.save(obj, path, open_way='memfs')
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|obj|必选|需要保存的对象。|有效数据对象。|
|path|必选|数据保存路径。|有效文件路径。|
|open_way|可选|保存方式。<ul><li>memfs：使用MindIO ACP的高性能MemFS保存数据。<li>fopen：调用C标准库中的文件操作函数保存数据，通常作为memfs方式的备份存在。</ul>默认值：memfs。|<ul><li>memfs <li>fopen</ul>|


**使用样例**

```
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> mindio_acp.save(x, '/mnt/dpc01/tensor.pt')
```

**返回值**

-   -1：保存失败。
-   0：通过原生torch.save方式实现保存。
-   1：通过memfs方式实现保存。
-   2：通过fopen方式实现保存。



### multi\_save接口

**接口功能**

将同一个数据保存到多个文件中。

**接口格式**

```
mindio_acp.multi_save(obj, path_list)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|obj|必选|需要保存的对象。|有效数据对象。|
|path_list|必选|数据保存路径列表。|有效文件路径列表。|


**使用样例**

```
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> path_list = ["/mnt/dpc01/dir1/rank_1.pt","/mnt/dpc01/dir2/rank_1.pt"]
>>> mindio_acp.multi_save(x, path_list)
```

**返回值**

-   None：失败。
-   0：通过原生torch.save方式实现保存。
-   1：通过memfs方式实现保存。
-   2：通过fopen方式实现保存。



### register\_checker接口

**接口功能**

注册异步回调函数。

**接口格式**

```
mindio_acp.register_checker(callback, check_dict, user_context, timeout_sec)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|callback|必选|回调函数（第一个参数result为数据完整性校验的结果，0为成功，其他为失败；第二个参数为user_context）。|有效函数名。|
|check_dict|必选|数据完整性校验条件，类型dict，用来校验指定path下的文件个数是否符合要求。|<ul><li>key：path，数据路径。<li>value：对应key路径下的文件个数。</ul>|
|user_context|必选|回调函数的第二个参数。|-|
|timeout_sec|必选|回调超时时间，单位：秒。<br>如果训练客户端日志中提示："watching checkpoint failed"，则需要调大该参数。代码在mindio_acp实际安装路径（mindio_acp/acc_checkpoint/framework_acp.py）下的async_write_tracker_file函数中。|[1, 3600]|


**使用样例**

```
>>> def callback(result, user_context):
>>>    if result == 0:
>>>        print("success")
>>>    else:
>>>        print("fail")
>>> context_obj = None
>>> check_dict = {'/mnt/dpc01/checkpoint-last': 4}
>>> mindio_acp.register_checker(callback, check_dict, context_obj, 1000)
```

**返回值**

-   None：失败。
-   1：成功。



### load接口

**接口功能**

从文件中加载save/multi\_save接口持久化的对象。

**接口格式**

```
mindio_acp.load(path, open_way='memfs', map_location=None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|加载路径。|有效文件路径。|
|open_way|可选|加载方式。<ul><li>memfs：使用MindIO ACP的高性能MemFS保存数据。<li>fopen：调用C标准库中的文件操作函数保存数据，通常作为memfs方式的备份存在。</ul>默认值：memfs。|<ul><li>memfs <li>fopen</ul>|
|map_location|可选|加载时需要映射到的设备。默认值：None。|<ul><li>None <li>cpu</ul>|


**使用样例**

```
>>> # load from file
>>> mindio_acp.load('/mnt/dpc01/checkpoint/rank-0.pt')
```

**返回值**

Any

> [!CAUTION]注意
> 如同PyTorch的load接口，本接口内部也使用pickle模块，有被恶意构造的数据在unpickle期间攻击的风险。需要保证被加载的数据来源是安全存储的，仅可以load可信的数据。



### convert接口

**接口功能**

将MindIO ACP格式的Checkpoint文件转换为Torch原生保存的格式。

**接口格式**

```
mindio_acp.convert(src, dst)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|src|必选|待转换的源路径或源文件，源路径或源文件必须存在。|有效文件路径，不能包含软链接。|
|dst|必选|待转换的目标路径或目标文件。指定路径的父目录必须存在，如果文件已存在，则会被覆盖。|有效文件路径，不能包含软链接。|


**使用样例**

```
>>> mindio_acp.convert('/mnt/dpc01/iter_0000050/mp_rank_00/distrib_optim.pt', '/mnt/dpc02/iter_0000050/mp_rank_00/distrib_optim.pt')
```

**返回值**

-   0：转换成功。
-   -1：转换失败。



### preload接口

**接口功能**

从文件中预加载使用torch保存的数据对象，并将其保存为MindIO ACP的高性能MemFS数据。

**接口格式**

```
mindio_acp.preload(*path)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|预加载的源文件，源文件必须存在。|有效文件路径或路径集合。|


**使用样例**

```
>>> # preload from file
>>> mindio_acp.preload('/mnt/dpc01/checkpoint/rank-0.pt')
```

**返回值**

-   0：预加载成功。
-   1：预加载失败。



### flush接口

**接口功能**

等待后台异步刷盘任务全部执行成功。

**接口格式**

```
mindio_acp.flush()
```

**接口参数**

无

**使用样例**

```
>>> # flush all data to disk
>>> mindio_acp.flush()
```

**返回值**

-   0：刷盘成功。
-   1：刷盘失败。



### open\_file接口

本接口只支持MindSpore框架。

**接口功能**

使用with调用open\_file接口，以只读的方式打开文件，并返回对应的\_ReadableFileWrapper实例。该实例提供read\(\)和close\(\)方法。

-   read：读取文件内容。

    ```
    read(self, offset=0, count=-1)
    ```

    |参数|是否必选|说明|取值要求|
    |--|--|--|--|
    |offset|可选|读取文件的偏移位置。需满足count + offset <= file_size|[0, file_size)|
    |count|可选|读取文件的大小。需满足count + offset <= file_size|<ul><li>-1：读取整个文件。 <li>(0, file_size]</ul>|


-   close：关闭文件。

    该方法在with退出上下文的时候自动调用。

    ```
    close(self)
    ```

**接口格式**

```
mindio_acp.open_file(path: str)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|加载路径。|有效文件路径。|


**使用样例**

```
>>> with mindio_acp.open_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     read_data = f.read()
```

**返回值**

\_ReadableFileWrapper实例。

> [!NOTE]说明
> 接口详情请参见[MindSpore文档](https://www.mindspore.cn/docs/zh-CN/master/api_python/mindspore/mindspore.load_checkpoint.html#mindspore.load_checkpoint)。



### create\_file接口

本接口只支持MindSpore框架。

**接口功能**

使用with调用create\_file接口，用于创建文件，并返回对应的\_WriteableFileWrapper实例。该实例提供write\(\)、drop\(\)和close\(\)方法。

-   write：向文件中写入数据。

    ```
    write(self, data: bytes)
    ```

    |参数|是否必选|说明|取值要求|
    |--|--|--|--|
    |data|必选|需要写入的对象。|bytes对象。|


-   drop：删除文件。

    ```
    drop(self)
    ```

-   close：关闭文件。

    该方法在with退出上下文的时候自动调用。

    ```
    close(self)
    ```

**接口格式**

```
mindio_acp.create_file(path: str, mode: int = 0o600)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|path|必选|数据保存路径。|有效文件路径。|
|mode|可选|文件创建权限。|[0o000, 0o777]|


**使用样例**

```
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> with mindio_acp.create_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     write_result = f.write(x)
```

**返回值**

\_WriteableFileWrapper实例。

> [!NOTE]说明
> 接口详情请参见[MindSpore文档](https://www.mindspore.cn/docs/zh-CN/master/api_python/mindspore/mindspore.save_checkpoint.html#mindspore.save_checkpoint)。



## 告警参考


### ALM-0x1001001  MindIO ACP持久化检查点数据异常

**告警解释**

当后端存储系统故障时，产生该告警。

当后端存储系统恢复后，告警恢复。

**告警属性**

|告警ID|告警级别|是否可自动清除|
|--|--|--|
|0x1001001|重要|是|


**告警参数**

无

**对系统的影响**

MindIO ACP组件不可服务，在用户侧转为直接操作后端存储。

**可能原因**

-   后端存储系统故障。
-   操作后端存储文件的权限不足。

**处理步骤**

1.  检查后端存储状态是否正常。
    -   状态正常，执行[2](#step_acp_li007)。
    -   状态异常，执行[3](#step_acp_li008)。

2.  <a id="step_acp_li007"></a>检查后端存储内的文件归属的用户名和属组权限，与客户端进程的权限是否一致。
    -   权限一致，告警会自动清除。
    -   权限不一致，执行[3](#step_acp_li008)。

3.  <a id="step_acp_li008"></a>搜集故障或者日志信息，联系技术支持处理。

**参考信息**

不涉及

**告警清除**

此告警修复后，系统会自动清除此告警，无需手工清除。



## 附录


### （可选）使用DPC文件访问存储，加速Checkpoint加载

检查是否满足如下条件：

-   是否使用DPC（Distributed Parallel Client，分布式并行客户端）文件系统访问存储。
-   是否成功安装NDS 1.0软件包（/opt/oceanstor/dataturbo/sdk/lib/libdpc\_nds.so）。
-   训练进程（如果在容器内）能否访问此so。

如果以上条件全部满足，则自动启用NDS 1.0直通读功能，加速加载Checkpoint。

成功加载NDS 1.0的判断依据是查看日志是否出现如下字样：

```
"initial and open nds file driver success"
```

NDS 1.0更多信息请参见[《OceanStor DataTurbo 25.x.x DTFS用户指南》](https://support.huawei.com/enterprise/zh/doc/EDOC1100539415/3f076df0)。

> [!CAUTION]注意
> 如果使用DPC文件系统访问存储，成功安装NDS 1.0软件包，安装地址为“/opt/oceanstor/dataturbo/sdk/lib/libdpc\_nds.so”，权限设置为444即可保证功能正常，启动训练前，请用户谨慎设置此文件的权限。



### 环境变量

|参数名称|参数说明|取值范围|缺省值|
|--|--|--|--|
|MINDIO_AUTO_PATCH_MEGATRON|是否在import mindio_acp的时候自动patch Megatron框架的源代码中的Checkpoint相关函数。|<ul><li>true或者1：开启 <li>其他值：关闭</ul>|false|
|HCOM_FILE_PATH_PREFIX|HCOM生成的文件路径的前缀，通过前缀保证文件只会在当前路径下（此路径需要已存在）创建和删除。|路径参数|${install_path}|



### 设置用户有效期

为保证用户的安全性，应设置用户的有效期，使用系统命令chage来设置用户的有效期。

命令为：

```
chage [-m mindays] [-M maxdays] [-d lastday] [-I inactive] [-E expiredate] [-W warndays] user
```

相关参数请参见[表1](#table_acp_04)。

**表 1<a id="table_acp_04"></a>**  设置用户有效期

|参数|参数说明|
|--|--|
|-d<br>--lastday|上一次更改的日期。|
|-E<br>--expiredate|用户到期的日期。超过该日期，此用户将不可用。|
|-h<br>--help|显示命令帮助信息。|
|-i<br>--iso8601|更改用户密码的过期日期并以YYYY-MM-DD格式显示。|
|-I<br>--inactive|停滞时期。超过指定天数后，设定密码为失效状态。|
|-l<br>--list|列出当前的设置。由非特权用户来确定口令或账户何时过期。|
|-m<br>--mindays|口令可更改的最小天数。设置为“0”表示任何时候都可以更改口令。|
|-M<br>--maxdays|口令保持有效的最大天数。设置为“-1”表示可删除这项口令的检测。设置为“99999”，表示无限期。|
|-R<br>--root|将命令执行的根目录设置为指定目录。|
|-W<br>--warndays|用户口令到期前，提前收到警告信息的天数。|


> [!NOTE]说明
> - 日期格式为YYYY-MM-DD，如 **chage -E 2017-12-01  _test_** 表示用户 **_test_**的口令在2017年12月1日过期。
> - user必须填写，填写时请替换为具体用户，默认为root用户。
> - 账号口令应该定期更新，否则容易导致安全风险。

举例说明：修改用户 **_test_** 的有效期为90天。

```
chage -M 90 test
```


### 口令复杂度要求

口令至少满足如下要求：

1.  口令长度至少8个字符。
2.  口令必须包含如下至少两种字符的组合：
    -   一个小写字母
    -   一个大写字母
    -   一个数字
    -   一个特殊字符：\`\~!@\#$%^&\*\(\)-\_=+\\|[\{\}];:'",<.\>/?和空格

3.  口令不能和账号一样。



### 账户一览表

|用户|描述|初始密码|密码修改方法|
|--|--|--|--|
|*{MindIO-install-user}*|MindIO ACP安装用户。|用户自定义。|使用 **passwd** 命令修改。|


> [!CAUTION]注意
> 为了保护密码安全性建议用户定期修改密码。



### 公网网址说明

以下表格中列出了产品中包含的公网网址，没有安全风险。

|网址|说明|
|--|--|
|http://www.apache.org/licenses/LICENSE-2.0|该网站是开源许可证的发布地址，用于代码版权信息声明。由于系统中无对外交互场景，因此无安全风险。|
