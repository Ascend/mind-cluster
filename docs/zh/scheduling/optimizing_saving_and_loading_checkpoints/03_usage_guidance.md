# 使用指导

> [!NOTE]说明
>
> - MindIO ACP SDK端支持宿主机和容器内部署。
> - 容器场景的镜像制作、镜像部署、镜像安全加固等由用户负责。
> - 只支持DeepSpeed框架、X1框架、MindSpeed-LLM、K8s的固定版本。
> - 在使用MindIO ACP服务时，启动训练任务的用户需要和启动MindIO ACP守护进程的用户属于同一个主组。

安装MindIO ACP SDK之后，为了使用MindIO ACP的缓存加速能力，将训练模型中使用到Python文件中的Torch的load/save函数，替换为MindIO ACP SDK的load/save函数。

- 支持将同一份数据保存到多个路径，将训练模型中循环保存同一份数据的torch.save函数，替换为MindIO ACP SDK的mindio\_acp.multi\_save函数。
- MindIO ACP SDK提供 `register_checker(callback, check_dict, user_context, timeout_sec)` 接口，支持将需要观察的文件夹和文件夹下的普通文件个数作为 `check_dict` 的元素注册到MindIO ACP。MindIO ACP会在 `timeout_sec` 时间内检查这些文件夹下的文件个数，并检查其与 `check_dict` 元素指定的文件个数是否相同，通过注册的callback函数回调应用程序，`user_context` 为callback函数的第二个参数，支持用户设置callback函数中需要调用的参数；`timeout_sec` 为注册超时时间，当超过超时时间仍然检查到不符合要求，则会在回调函数中报告错误。用户可以根据检查结果处理后续业务逻辑。

## Torch对接DeepSpeed框架

1. 使用业务用户登录到计算节点。

    > [!NOTE]说明
    > 业务用户不是 *{MindIO-install-user}*、HwHiAiUser、hwMindX用户，由用户根据实际情况决定。

2. 进入DeepSpeed安装目录。

    ```bash
    cd {deepspeed安装目录}/runtime 
    ```

3. <a id="step_acp_li001"></a>修改engine.py文件。
    1. 打开engine.py文件。

        ```bash
        vim engine.py
        ```

    2. <a id="step_acp_li002"></a>按“i”进入编辑模式，修改如下内容。
        - 在文件首行加入以下内容。

            ```python
            import mindio_acp
            ```

        - 将torch.load函数替换为mindio\_acp.load函数。

            替换前：

            ```python
            optim_checkpoint = torch.load(optim_load_path,
                                          map_location=torch.device('cpu'))
            ```

            替换后：

            ```python
            optim_checkpoint = mindio_acp.load(optim_load_path, map_location='cpu')
            ```

        - 将torch.save函数替换为mindio\_acp.save函数。

            替换前：

            ```python
            torch.save(state, save_path)
            ```

            替换后：

            ```python
            mindio_acp.save(state, save_path)
            ```

        - 将包含torch.save函数的with open语句整体替换为mindio\_acp.save函数。

            替换前：

            ```python
            with open(self._get_optimizer_ckpt_name(save_dir, tag, expp_rank), 'wb') as fd:
                torch.save(optimizer_state, fd)
                fd.flush()
            ```

            替换后：

            ```python
            mindio_acp.save(optimizer_state, self._get_optimizer_ckpt_name(save_dir, tag, expp_rank))
            ```

        - 替换DeepSpeedEngine.\_get\_expert\_ckpt\_name函数。

            替换前：

            ```python
                            expert_state_dict = torch.load(DeepSpeedEngine._get_expert_ckpt_name(
                                checkpoint_path,
                                -1, # -1 means ignore layer_id
                                global_expert_id,
                                tag,
                                mpu),
                                map_location=torch.device('cpu'))
            ```

            替换后：

            ```python
                            expert_state_dict = mindio_acp.load(DeepSpeedEngine._get_expert_ckpt_name(
                                checkpoint_path,
                                -1, # -1 means ignore layer_id
                                global_expert_id,
                                tag,
                                mpu),
                                map_location='cpu')
            ```

    3. <a id="step_acp_li003"></a>按“Esc”键，输入 **:wq!** ，按“Enter”保存并退出编辑。

4. 修改module.py文件。
    1. 打开module.py文件。

        ```bash
        vim pipe/module.py
        ```

    2. 替换torch.save和torch.load，替换方式参见[步骤3.2](#step_acp_li002)  \~ [步骤3.3](#step_acp_li003)。

5. <a id="step_acp_li004"></a>修改state\_dict\_factory.py文件。
    1. 打开state\_dict\_factory.py文件。

        ```bash
        vim state_dict_factory.py
        ```

    2. 替换torch.save和torch.load，替换方式参见[步骤3.2](#step_acp_li002)  \~ [步骤3.3](#step_acp_li003)。

6. 完成[步骤3](#step_acp_li001)  \~ [步骤5](#step_acp_li004)的.py文件修改，DeepSpeed即可使用MindIO ACP服务。

## Torch对接X1框架

1. 登录到计算节点。
2. 进入X1安装目录。

    ```bash
    cd {X1安装目录}/Megatron-LM/megatron
    ```

3. 修改checkpointing.py文件。
    1. 打开checkpointing.py文件。

        ```bash
        vim checkpointing.py
        ```

    2. 按“i”进入编辑模式，修改如下内容。
        - 在文件首行加入以下内容。

            ```python
            import mindio_acp
            ```

        - 将torch.load函数替换为mindio\_acp.load函数。

            替换前：

            ```python
            optim_checkpoint = torch.load(optim_load_path,
                                          map_location=torch.device('cpu'))
            ```

            替换后：

            ```python
            optim_checkpoint = mindio_acp.load(optim_load_path, map_location='cpu')
            ```

        - 将torch.save函数替换为mindio\_acp.save函数。

            替换前：

            ```python
            torch.save(state, save_path)
            ```

            替换后：

            ```python
            mindio_acp.save(state, save_path)
            ```

        - 将包含torch.save函数的with open语句整体替换为mindio\_acp.save函数。

            替换前：

            ```python
            with open(self._get_optimizer_ckpt_name(save_dir, tag, expp_rank), 'wb') as fd:
                torch.save(optimizer_state, fd)
                fd.flush()
            ```

            替换后：

            ```python
            mindio_acp.save(optimizer_state, self._get_optimizer_ckpt_name(save_dir, tag, expp_rank))
            ```

    3. 按“Esc”键，输入 **:wq!** ，按“Enter”保存并退出编辑。

## Torch对接MindSpeed-LLM框架

**前提条件**

- 使用前请先了解MindIO ACP特性的[约束限制](./02_installation_and_deployment.md#约束限制)章节。
- MindSpeed-LLM框架准备请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。匹配的Megatron-LM版本为 **core\_v0.12.1**。

> [!NOTE]说明
> 本次发布包配套MindSpeed-LLM的 **2.3.0** 分支，环境、代码、数据集准备请用户参考MindSpeed-LLM仓库的相关指导说明，并确保其安全性。

**操作步骤**

1. 使用业务用户登录到计算节点。

    > [!NOTE]说明
    > 业务用户不是 *{MindIO-install-user}*、HwHiAiUser、hwMindX用户，由用户根据实际情况决定。

2. 进入MindSpeed-LLM安装目录。

    ```bash
    cd MindSpeed-LLM/
    ```

3. <a id="step_acp_li005"></a>修改pretrain\_gpt.py文件。
    1. 打开pretrain\_gpt.py文件。

        ```bash
        vim pretrain_gpt.py
        ```

    2. 按“i”进入编辑模式，在文件头部找到`from mindspeed_llm import megatron_adaptor`，换行增加 `import mindio_acp`。

        ```python
        from mindspeed_llm import megatron_adaptor
        import mindio_acp
        ```

    3. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

4. <a id="step_acp_li006"></a>编辑预训练脚本（仅供参考）。

    此处以编辑“examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh”脚本为例。

    1. 打开“examples/mcore/llama2/pretrain\_llama2\_7b\_ptd.sh”脚本。

        ```bash
        vim examples/mcore/llama2/pretrain_llama2_7b_ptd.sh
        ```

    2. 按“i”进入编辑模式，在脚本中增加如下内容以启用周期性Checkpoint加速功能。
        
        ```bash
        export MINDIO_AUTO_PATCH_MEGATRON=true
        export GLOO_SOCKET_IFNAME=enp189s0f0
        export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:$LD_LIBRARY_PATH
        source /usr/local/Ascend/cann/set_env.sh
        ```
        
        修改后的pretrain_llama2_7b_ptd.sh脚本示例如下：

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

        周期性Checkpoint加速功能相关参数说明如下：

        - **MINDIO\_AUTO\_PATCH\_MEGATRON**：MindIO ACP框架自动patch Megatron的源码，用于启用加速周期性Checkpoint特性。
        - **GLOO\_SOCKET\_IFNAME**：根据主节点高速网卡实际情况进行配置。
        - **LD\_LIBRARY\_PATH**：CANN包驱动的so库地址，请根据CANN实际的安装路径进行修改。
        - **set\_env.sh文件路径**：请根据CANN实际的安装路径进行修改。

    3. 按“Esc”键，输入 **:wq!** ，按“Enter”保存并退出编辑。

5. 完成[步骤3](#step_acp_li005)  \~ [步骤4](#step_acp_li006)的.py文件修改，MindSpeed-LLM即可使用MindIO ACP加速周期性Checkpoint特性。

## 对接K8s

在容器中使用MindIO ACP加速服务时，需要将SDK安装到对应的容器中。

1. 修改创建Pod的yaml文件，下面以“/home/testuser/mygpt.yaml”文件为例，增加映射卷配置。
    1. 打开mygpt.yaml文件。

        ```bash
        vim /home/testuser/mygpt.yaml
        ```

    2. 按“i”进入编辑模式，修改mygpt.yaml文件。

        > [!NOTE]说明
        > - 如果volumeMounts和volumes不存在，直接在文件中添加全部内容。
        > - 如果volumeMounts和volumes已存在，只需在volumeMounts和volumes内部添加其后面的内容。

        - （可选）如果环境中[使用了DPC访问存储](./07_appendixes.md#可选使用dpc文件访问存储加速checkpoint加载)，增加卷在容器中映射路径，内容如下：

            ```yaml
            volumeMounts:
                - mountPath: /opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so
                  name: mindio-dpc-nds
                  readOnly: false
            ```

            > [!NOTE]说明
            > “/opt/oceanstor/dataturbo/sdk/lib/libdpc\_nds.so”不可随意更改。

        - （可选）如果环境中[使用了DPC访问存储](./07_appendixes.md#可选使用dpc文件访问存储加速checkpoint加载)，增加宿主机需要映射的卷声明，增加内容如下：

            ```yaml
            volumes:
              - name: mindio-dpc-nds
                hostPath:
                  path: /opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so
                  type: File
            ```

    3. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。

2. 使用修改后的yaml文件，创建Pod。

    ```bash
    kubectl apply -f mygpt.yaml
    ```

3. 进入到创建好的Pod，以命名空间“test-mindio”下名称为“mygptdd”的Pod为例。

    ```bash
    kubectl exec -it mygptdd -n test-mindio /bin/bash
    ```

4. 将MindIO ACP SDK上传到Pod中，并参见[在计算节点安装MindIO ACP SDK](./02_installation_and_deployment.md#在计算节点安装mindio-acp-sdk)完成SDK安装。

## Checkpoint文件格式转换示例（Torch）

对于使用PyTorch框架的用户，在大模型训练结束后，Checkpoint文件需要用于推理。这里举例说明，如何将MindIO ACP保存的Checkpoint文件转换成Torch原生格式的文件。

> [!NOTE]说明
>
> - **load\_dir**：替换为真实的Checkpoint保存目录。
> - **new\_dir**：替换为Checkpoint转换后新保存的目录，建议为空目录。
> - **iteration**：指定转换这个iteration迭代周期的所有Checkpoint文件，会和 **load\_dir** 进行拼接。

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
