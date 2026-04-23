# 使用指导

> [!NOTE]说明
> MindIO TFT以SDK的形式提供服务，支持部署在裸机和容器环境中。

安装MindIO TFT SDK之后，需要在框架中启动MindIO TFT模块，并在训练过程中同步优化器数据更新状态到该模块。

## 对接MindSpeed-LLM框架

**前提条件**

- 使用前请先了解MindIO TFT的[约束限制](./02_installation_and_deployment.md#约束限制)。
- MindSpeed-LLM框架准备参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。匹配的Megatron-LM的版本为 **core\_v0.12.1**。

> [!NOTE]说明
>
> - 本次发布包配套MindSpeed-LLM的 **2.3.0** 分支，环境、代码、数据集准备请用户参考MindSpeed-LLM仓库的相关指导说明，并确保其安全性。
> - MindIO TFT对接MindSpeed-LLM框架，目前支持MindIO TTP、MindIO UCE和MindIO ARF功能。
> - 对于PyTorch类框架，安装或开启MindCluster后，跳过[步骤1](#step_tft_li001)对“torchrun”文件的修改，由MindCluster控制进程退出。

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
        - **TTP\_ADDR**：集群主节点的IP地址，须符合IPv4或IPv6标准格式。参数详情请参见[环境变量](./06_appendixes.md#环境变量)。
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

## 对接MindCluster

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
          - containerPort: 8000 # 用于MindIO TFT服务Controller与Processor通信端口
            name: ttp-port
        ```

    3. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

- 适配K8s网络，在[步骤2](#step_tft_li002)的预训练脚本基础上做如下修改。

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

## 对接非MindSpeed-LLM框架

**前提条件**

使用前请先了解MindIO TFT的[约束限制](./02_installation_and_deployment.md#约束限制)。

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
