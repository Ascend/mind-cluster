# Usage Guide

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:27:52.332Z pushedAt=2026-06-09T07:15:15.736Z -->

> [!NOTE]
>
> - MindIO ACP SDK supports deployment on both the host and inside containers.
> - You are responsible for image creation, image deployment, and image security hardening in container scenarios.
> - Only fixed versions of the DeepSpeed framework, X1 framework, MindSpeed-LLM, and Kubernetes are supported.
> - To use the MindIO ACP service, the user starting the training job must belong to the same primary group as the user starting the MindIO ACP daemon.

After installing MindIO ACP SDK, to leverage the cache acceleration capabilities of MindIO ACP, replace the Torch load/save functions used in the Python files of the training model with the load/save functions of MindIO ACP SDK.

- Saving the same data to multiple paths is supported. Replace the `torch.save` function used for loop saving the same data in the training model with the `mindio_acp.multi_save` function of MindIO ACP SDK.
- MindIO ACP SDK provides the `register_checker(callback, check_dict, user_context, timeout_sec)` interface, which supports registering the folders to be observed and the number of regular files in those folders as elements of `check_dict` with MindIO ACP. MindIO ACP will check the number of files in these folders within the `timeout_sec` period and verify whether they match the file count specified by the `check_dict` elements. It will then call back the application through the registered callback function. `user_context` is the second parameter of the callback function, allowing you to set the parameters to be invoked within the callback function. `timeout_sec` is the registration timeout period; if the check still fails to meet the requirements after the timeout period, an error will be reported in the callback function. Handle subsequent business logic based on the check results.

## Torch-DeepSpeed Integration

1. Log in to the compute node as the service user.

    > [!NOTE]NOTE
    > The service user is not the `{MindIO-install-user}`, `HwHiAiUser`, or `hwMindX` user. It is determined based on the actual situation.

2. Go to the DeepSpeed installation directory.

    ```bash
    cd {DeepSpeed installation directory}/runtime
    ```

3. <a id="step_acp_li001"></a>Modify the `engine.py` file.
    1. Open the `engine.py` file.

        ```bash
        vim engine.py
        ```

    2. <a id="step_acp_li002"></a>Press `i` to enter insert mode, and modify the following content.
        - Add the following content to the first line of the file.

            ```python
            import mindio_acp
            ```

        - Replace the `torch.load` function with the `mindio_acp.load` function.

            Before:

            ```python
            optim_checkpoint = torch.load(optim_load_path,
                                          map_location=torch.device('cpu'))
            ```

            After:

            ```python
            optim_checkpoint = mindio_acp.load(optim_load_path, map_location='cpu')
            ```

        - Replace the `torch.save` function with the `mindio_acp.save` function.

            Before:

            ```python
            torch.save(state, save_path)
            ```

            After:

            ```python
            mindio_acp.save(state, save_path)
            ```

        - Replace the entire `with open` statement containing the `torch.save` function with the `mindio_acp.save` function.

            Before:

            ```python
            with open(self._get_optimizer_ckpt_name(save_dir, tag, expp_rank), 'wb') as fd:
                torch.save(optimizer_state, fd)
                fd.flush()
            ```

            After:

            ```python
            mindio_acp.save(optimizer_state, self._get_optimizer_ckpt_name(save_dir, tag, expp_rank))
            ```

        - Replace the `DeepSpeedEngine._get_expert_ckpt_name` function.

            Before:

            ```python
                            expert_state_dict = torch.load(DeepSpeedEngine._get_expert_ckpt_name(
                                checkpoint_path,
                                -1, # -1 means ignore layer_id
                                global_expert_id,
                                tag,
                                mpu),
                                map_location=torch.device('cpu'))
            ```

            After:

            ```python
                            expert_state_dict = mindio_acp.load(DeepSpeedEngine._get_expert_ckpt_name(
                                checkpoint_path,
                                -1, # -1 means ignore layer_id
                                global_expert_id,
                                tag,
                                mpu),
                                map_location='cpu')
            ```

    3. <a id="step_acp_li003"></a>Press `Esc`, type *`:wq!`, and press `Enter` to save and exit insert mode.

4. Modify the `module.py` file.
    1. Open the `module.py` file.

        ```bash
        vim pipe/module.py
        ```

    2. Replace `torch.save` and `torch.load`. For the replacement method, see [Step 3.2](#step_acp_li002) to [Step 3.3](#step_acp_li003).

5. <a id="step_acp_li004"></a>Modify the `state_dict_factory.py` file.
    1. Open the `state_dict_factory.py` file.

        ```bash
        vim state_dict_factory.py
        ```

    2. Replace `torch.save` and `torch.load`. For details on the replacement method, see [Step 3.2](#step_acp_li002) to [Step 3.3](#step_acp_li003).

6. After completing the `.py` file modifications from [Step 3](#step_acp_li001) to [Step 5](#step_acp_li004), DeepSpeed can use the MindIO ACP service.

## Torch-X1 Integration

1. Log in to the compute node.
2. Go to the X1 installation directory.

    ```bash
    cd {X1 installation directory}/Megatron-LM/megatron
    ```

3. Modify the `checkpointing.py` file.
    1. Open the `checkpointing.py` file.

        ```bash
        vim checkpointing.py
        ```

    2. Press `i to enter edit mode and modify the following content.
        - Add the following content to the first line of the file.

            ```python
            import mindio_acp
            ```

        - Replace the `torch.load` function with the `mindio_acp.load` function.

            Before:

            ```python
            optim_checkpoint = torch.load(optim_load_path,
                                          map_location=torch.device('cpu'))
            ```

            After:

            ```python
            optim_checkpoint = mindio_acp.load(optim_load_path, map_location='cpu')
            ```

        - Replace the `torch.save` function with the `mindio_acp.save` function.

            Before:

            ```python
            torch.save(state, save_path)
            ```

            After:

            ```python
            mindio_acp.save(state, save_path)
            ```

        - Replace the entire `with open` statement containing the `torch.save` function with the `mindio_acp.save` function.

            Before:

            ```python
            with open(self._get_optimizer_ckpt_name(save_dir, tag, expp_rank), 'wb') as fd:
                torch.save(optimizer_state, fd)
                fd.flush()
            ```

            After:

            ```python
            mindio_acp.save(optimizer_state, self._get_optimizer_ckpt_name(save_dir, tag, expp_rank))
            ```

    3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit insert mode.

## Torch-MindSpeed-LLM Integration

**Prerequisites**

- Before use, familiarize yourself with the [Constraints](./02_installation_and_deployment.md#constraints) section of the MindIO ACP feature.
- For MindSpeed-LLM framework preparation, see [MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0). The matching Megatron-LM version is **core_v0.12.1**.

> [!NOTE]
> This release package is compatible with the **2.3.0** branch of MindSpeed-LLM. For environment, code, and dataset preparation, refer to the relevant guidance in the MindSpeed-LLM repository and ensure their security.

**Procedure**

1. Log in to the compute node as a service user.

    > [!NOTE]Note
    > The service user is not the `{MindIO-install-user}`, `HwHiAiUser`, or `hwMindX` user. It is determined based on the actual situation.

2. Go to the MindSpeed-LLM installation directory.

    ```bash
    cd MindSpeed-LLM/
    ```

3. <a id="step_acp_li005"></a>Modify the `pretrain_gpt.py` file.
    1. Open the `pretrain_gpt.py` file.

        ```bash
        vim pretrain_gpt.py
        ```

    2. Press `i` to enter edit mode, locate `from mindspeed_llm import megatron_adaptor` at the top of the file, and add `import mindio_acp` on a new line.

        ```python
        from mindspeed_llm import megatron_adaptor
        import mindio_acp
        ```

    3. Press the `Esc` key, type `:wq!`, and press `Enter` to save and exit insert mode.

4. <a id="step_acp_li006"></a>Edit the pre-training script (for reference only).

    This example uses the `"examples/mcore/llama2/pretrain_llama2_7b_ptd.sh"` script.

    1. Open the `"examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh"` script.

        ```bash
        vim examples/mcore/llama2/pretrain_llama2_7b_ptd.sh
        ```

    2. Press `i` to enter insert mode, and add the following content to the script to enable the periodic checkpoint acceleration feature.

        ```bash
        export MINDIO_AUTO_PATCH_MEGATRON=true
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:$LD_LIBRARY_PATH
        source /usr/local/Ascend/cann/set_env.sh
        ```

        The modified `pretrain_llama2_7b_ptd.sh` script example is as follows:

        ```bash
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

        The parameters related to the periodic checkpoint acceleration feature are described as follows:

        - `MINDIO_AUTO_PATCH_MEGATRON`: The MindIO ACP framework automatically patches the Megatron source code to enable the periodic checkpoint acceleration feature.
        - `GLOO_SOCKET_IFNAME`: Configure this based on the actual high-speed NIC of the master node.
        - `LD_LIBRARY_PATH`: The address of the CANN driver's `.so` library. Modify this based on the actual installation path of CANN.
        - `set_env.sh file path`: Modify this based on the actual installation path of CANN.

    3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit insert mode.

5. After completing the `.py` file modifications in [Step 3](#step_acp_li005)  \~ [Step 4](#step_acp_li006), MindSpeed-LLM can use MindIO ACP's periodic checkpoint acceleration feature.

## Integrating with Kubernetes

When using the MindIO ACP acceleration service in a container, you need to install the SDK into the corresponding container.

1. Modify the yaml file for pod creation. The following uses the `"/home/testuser/mygpt.yaml"` file as an example to add mapped volume configuration.
    1. Open the `mygpt.yaml` file.

        ```bash
        vim /home/testuser/mygpt.yaml
        ```

    2. Press `i` to enter edit mode and modify the `mygpt.yaml` file.

        > [!NOTE]
        > - If `volumeMounts` and `volumes` do not exist, add all the content directly to the file.
        > - If `volumeMounts` and `volumes` already exist, only add the content that follows within their respective blocks.

        - (Optional) If [DPC is used to access storage](./07_appendixes.md) in the environment, add the volume mapping path in the container. The content is as follows:

            ```yaml
            volumeMounts:
                - mountPath: /opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so
                  name: mindio-dpc-nds
                  readOnly: false
            ```

            > [!NOTE]
            > `"/opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so"` cannot be changed arbitrarily.

        - (Optional) If [DPC is used to access storage](./07_appendixes.md) in the environment, add the volume claim that needs to be mapped on the host. The content to add is as follows:

            ```yaml
            volumes:
              - name: mindio-dpc-nds
                hostPath:
                  path: /opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so
                  type: File
            ```

    3. Press `Esc`, enter `:wq!`, and press `Enter` to save and exit insert mode.

2. Use the modified yaml file to create a pod.

    ```bash
    kubectl apply -f mygpt.yaml
    ```

3. Enter the created pod. The following example uses a pod named `"mygptdd"` in the `"test-mindio"` namespace.

    ```bash
    kubectl exec -it mygptdd -n test-mindio /bin/bash
    ```

4. Upload the MindIO ACP SDK to the pod, and complete the SDK installation by referring to [Installing MindIO ACP SDK on a Compute Node](./02_installation_and_deployment.md).

## Checkpoint File Format Conversion Example (Torch)

For users of the PyTorch framework, after large model training is complete, the checkpoint files need to be used for inference. Here is an example of how to convert checkpoint files saved by MindIO ACP into files in the native Torch format.

> [!NOTE] NOTE
>
> - `load_dir`: Replace with the actual Checkpoint save directory.
> - `new_dir`: Replace with the new directory for saving the converted Checkpoint. An empty directory is recommended.
> - `iteration`: Specifies the conversion of all checkpoint files for this iteration cycle, which will be concatenated with `load_dir`.

```python
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
