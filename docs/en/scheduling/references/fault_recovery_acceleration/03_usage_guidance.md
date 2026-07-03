# Usage Guidance

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:23:59.596Z pushedAt=2026-06-09T07:15:15.706Z -->

> [!NOTE]
> MindIO TFT provides services in the form of an SDK, supporting deployment in both bare-metal and container environments.

After installing the MindIO TFT SDK, you need to start the MindIO TFT module in your framework and synchronize the optimizer data update status to this module during training.

## Integrating with the MindSpeed-LLM Framework

**Prerequisites**

- Before use, please understand the [Constraints](./02_installation_and_deployment.md#constraints) of MindIO TFT.
- For MindSpeed-LLM Framework preparation, see [MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0). The matching Megatron-LM version is **core_v0.12.1**.

> [!NOTE]
>
> - This release package is paired with the **2.3.0** branch of MindSpeed-LLM. For environment, code, and dataset preparation, users should refer to the relevant guidance in the MindSpeed-LLM repository and ensure its security.
> - MindIO TFT- MindSpeed-LLM integration currently supports MindIO TTP, MindIO UCE, and MindIO ARF features.
> - For PyTorch Frameworks, after installing or enabling MindCluster, skip the modification of the `"torchrun"` file in [Step 1](#step_tft_li001), and let MindCluster control the process exit.

**Procedure**

1. <a id="step_tft_li001"></a>(Optional) Edit the `"torchrun"` file.
    1. Locate the `"torchrun"` file in the environment.

        ```bash
        which torchrun
        ```

    2. Open the `"torchrun"` file at the path shown by the command above.

        ```bash
        vim {torchrun file path}/torchrun
        ```

    3. Press `i` to enter insert mode, and add `import mindio_ttp.framework_ttp` at the corresponding position in the file.

        ```python
        import re
        import sys
        import mindio_ttp.framework_ttp
        from torch.distributed.run import main as torch_main
        ```

    4. Press `Esc`, type `:wq!`, and press `Enter` to save and exit insert mode.

2. <a id="step_tft_li002"></a>Edit the pre-training script (for reference only).

    The following uses editing the `"examples/mcore/llama2/pretrain_llama2_7b_ptd.sh"` script as an example.

    1. Open the `"examples/mcore/llama2/pretrain_llama2_7b_ptd.sh"` script.

        ```bash
        vim examples/mcore/llama2/pretrain_llama2_7b_ptd.sh
        ```

    2. Press `i` to enter insert mode. To enable the high availability feature, add the following content to the script.

        ```bash
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export TTP_ADDR="master node ip"
        source /usr/local/Ascend/cann/set_env.sh

        # After --bf16 in GPT_ARGS, add the following content
            \
            --enable-high-availability \
            --enable-hbmfault-repair \
            --enable-worker-reboot \
            --distributed-optimizer-no-replica \

        ```

        The modified `pretrain_llama2_7b_ptd.sh` script example is as follows:

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

        The parameters related to the high availability feature are described as follows:

        - `GLOO_SOCKET_IFNAME`: Configure this based on the actual high-speed NIC of the master node.
        - `TTP_ADDR`: The IP address of the cluster master node, which must conform to the standard IPv4 or IPv6 format. For details, see [Environment Variables](./06_appendixes.md#environment-variables).
        - `set_env.sh file path`: Modify this based on the actual installation path of CANN.
        - `enable-high-availability`: The master switch for MindIO TFT, disabled by default. When configured, the dying gasp feature is enabled by default.

            After the MindIO TFT switch is enabled, the memory usage of various optimizers will change. For details about the changes, see [Table 1](#table_tft_03).

            For distributed optimizers, the static memory increases due to the addition of optimizer replicas. However, as the cluster size increases, the DP Size becomes larger, and the average increase in memory per device is very small, which helps avoid OOM. Therefore, it is recommended to be used in large clusters. Choose whether to enable it and adjust parameters based on the memory situation.

        - `enable-hbmfault-repair`: switch for MindIO UCE, disabled by default. When configured, it performs fault detection on on-chip memory and completes online repair, achieving step-level recomputation functionality. This switch takes effect when `enable-high-availability` is enabled. This feature depends on the memory management mechanism of PyTorch and can only be used when the PyTorch environment variable `PYTORCH_NO_NPU_MEMORY_CACHING` is not configured, meaning the memory reuse mechanism is enabled. If `export PYTORCH_NO_NPU_MEMORY_CACHING = 1`, this feature cannot be used.
        - `enable-worker-reboot`: switch for MindIO ARF, disabled by default. When configured, it performs process-level restart repair to continue training when a general fault occurs. This switch takes effect when `enable-high-availability` is enabled.
        - `distributed-optimizer-no-replica`: After enabling the high availability feature, the distributed optimizer adds optimizer replicas by default, which increases on-chip memory usage. After enabling this switch, the distributed optimizer does not increase replica memory usage. In MindIO UCE and MindIO ARF scenarios, periodic checkpoints are directly used for online repair.

        **Table 1<a id="table_tft_03"></a>**  Theoretical numerical changes in optimizer parameters between the native optimizer and the optimizer when MindIO TFT enabled

        |Optimizer|Native|MindIO TFT Enabled|Description|
        |--|--|--|--|
        |fp16/bf16|20|20|-|
        |fp32|16|16|-|
        |fp16/bf16 Distributed|4 + 16/d|4 + 16 * N/d|<ul><li>d: DP Group Size</li><li>N: Number of replicas, N < d</li></ul>|

    3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit insert mode.

## Integrating with MindCluster

MindIO TFT provides services in the form of an SDK, with no resident processes. The service starts when the training process starts. When the training job ends, the service exits.

When integrating with MindCluster, MindCluster manages Kubernetes containers. The integration process within  Kubernetes containers is consistent with bare-metal installation and deployment.

**Procedure**

- When the Python environment is not installed on shared storage, to facilitate use on large clusters, the MindIO TFT SDK can be integrated into the container image. When a pod is deployed from this image, the MindIO TFT SDK is already installed.
- The MindIO TFT Controller module and Processor module exchange heartbeat messages. When Kubernetes performs network isolation, you need to add the communication port to the yaml file configured when creating the pod.

    Modify the yaml file configured when creating the pod. Here, `"pod.yaml"` is used as an example.

    1. Open the `"pod.yaml"` file.

        ```bash
        vim pod.yaml
        ```

    2. Press `i` to enter insert mode and add the following content.

        ```yaml
        ports:
          - containerPort: 8000 # Port for communication between the MindIO TFT service Controller and Processor
            name: ttp-port
        ```

    3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit insert mode.

- To adapt to the Kubernetes network, make the following modifications based on the pre-training script in [Step 2](#step_tft_li002).

    ```yaml
    # Comment out the following two lines. These environment variables are configured by MindCluster.
    # MASTER_ADDR=$(hostname -I | awk '{print $1}')
    # MASTER_PORT=XXXX

    # Obtain the MASTER_ADDR and MASTER_PORT environment variables (the service network IP address of Kubernetes) from Kubernetes.
    CONTROLLER_ADDR=$(hostname -I | awk '{print $1}')
    PROCESSOR_ADDR=${MASTER_ADDR}
    export CONTROLLER_ADDR
    export PROCESSOR_ADDR
    ```

## Integrating with Non-MindSpeed-LLM Frameworks

**Prerequisites**

Before you proceed, please understand the [Constraints](./02_installation_and_deployment.md#constraints) of MindIO TFT.

> [!NOTE]
>
> - This release package supports Megatron-like frameworks. You are responsible for preparing the environment, code, and datasets, and ensuring their security.
> - The content in this section is for adaptation guidance only. Specific implementation details must be implemented by yourself.

**Feature Reference**

The functional adaptation points required for related features are shown in [Table 1](#table_tft_04), and the code reference links corresponding to each functional adaptation point are shown in [Table 2](#table_tft_05).

**Table 1<a id="table_tft_04"></a>**  Features and functional adaptation points

| Feature |  Adaptation Point Number |
|--|--|
| Dying Gasp | 1, 2, 3, 4, 5, 6, 7 |
| UCE Fast Recovery | 1, 2, 3, 4, 5, 6, 8, 10, 11 |
| Network Fast Recovery | 1, 2, 5, 6, 11 |
| Process Fast Recovery | 1, 2, 3, 4, 5, 6, 9, 10, 11 |
| Hot Switching | 1, 2, 3, 4, 5, 9, 10, 11, 12 |
| Online Stress Testing/Link Failover and Switchback | 1, 2, 12 |

**Table 2<a id="table_tft_05"></a>** Code reference links for related features

| No. | Adapted Feature | Reference Code |
|--|--|--|
| 1 | Boot While Initializing | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/268f870b10e450feade3c98b603254851e8fa4cd?ref=pre_preparation) |
| 2 | Report Optimizer Update Status | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/268f870b10e450feade3c98b603254851e8fa4cd?ref=pre_preparation) |
| 3 | Create DP Replica Group | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/df6317e62ef7cefcec25ba8740f25e152eba34e4?ref=create_dp_replica_group) |
| 4 | Optimizer Replica | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/e3490911407d88f9c6d3ac0c0eb3186f1812d171?ref=replica_optimizer) |
| 5 | Exception Capture Decorator | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/0827869d031303a231a69897c12692fb92d8cf8d?ref=exception_handler) |
| 6 | Operator Resource Cleanup | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/45824ee7303c05bce1260f2cab590dd858147767?ref=stop_clean) |
| 7 | Dying Gasp Checkpoint | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/0e94a3fcb2643580d151b90deb205e9034adde2a?ref=dump_ckpt) |
| 8 | UCE Model Optimizer Rebuild | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/93f599fa480c7f7931c74e782c617e0ebaffceb9?ref=uce_clear_rebuild) |
| 9 | Node Restart and Communication Rebuild | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/4b490ff888cea9766e461f6bb53e73712adf097d?ref=node_reboot) |
| 10 | Parameter-Plane Online Repair | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/9bd17ca7fdda3f8c5f70eef68cf1db4ac2ba738f?ref=online_repair_ckpt) |
| 11 | State Rollback | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/7835670ec12b2ae5969bd1cd9eec72c882225c18?ref=rollback_callback) |
| 12 | Graceful Pause | [LLM Repository Reference Link](https://gitcode.com/wlwen/MindSpeed-LLM/commit/db87cc048455f67218f5a8caca626f1b64d35f61?ref=active_pause) |
