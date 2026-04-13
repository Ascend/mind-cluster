# 配置推理业务故障恢复<a name="ZH-CN_TOPIC_0000002511630975"></a>

在一体机混部或无K8s的场景下，推理进程异常后，没有有效的恢复手段。本章节提供了推理业务故障后自动恢复的示例。示例中启动脚本作为容器entrypoint，自动拉起推理进程，监控推理进程状态，并在异常后重新拉起推理进程。

- 支持MindIE Server单机推理。
- 不支持MindIE Server多机推理。原因是仅重启其中一个容器中的推理进程，业务无法恢复。

**操作步骤<a name="section169801610181818"></a>**

以下配置过程以Qwen3-1.7B模型为例。

1. 获取MindIE容器镜像。
    - 方式一：进入昇腾镜像仓库的[MindIE镜像下载](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)页面，下载MindIE镜像。
    - 方式二：参考《MindIE安装指南》中的“安装MindIE \> [方式三：容器安装方式](https://www.hiascend.com/document/detail/zh/mindie/230/envdeployment/instg/mindie_instg_0022.html)”章节，自行准备镜像。

2. 在节点上查看MindIE镜像。

    ```shell
    docker images |grep mindie
    ```

    回显示例如下：

    ```ColdFusion
    …
    swr.cn-south-1.myhuaweicloud.com/ascendhub/mindie   2.1.RC2-800I-A2-py311-openeuler24.03-lts   a4708118cd12        6 weeks ago         16GB
    …
    ```

3. 获取Qwen3-1.7B模型权重。

    ```shell
    # 创建模型权重保存目录
    mkdir -p /data/atlas_dls/public/infer/model_weight
    cd /data/atlas_dls/public/infer/model_weight/
    # 若未安装git-lfs，需要先安装git-lfs。git-lfs是一个Git扩展，专门用于管理大文件和二进制文件
    yum install -y git-lfs 
    # git启用lfs
    git lfs install
    # 权重下载
    git clone https://www.modelscope.cn/Qwen/Qwen3-1.7B.git
    # 修改权重文件权限
    chmod -R 750 Qwen3-1.7B/
    # （可选）如果使用普通用户镜像，权重路径所属应为镜像内默认的1000用户
    chown -R 1000:1000 Qwen3-1.7B/
    ```

    >[!NOTE] 
    >某些模型下载后，还需进行权重量化，详细请参见[ModelZoo-PyTorch](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/MindIE/LLM)中各类模型的README.md。

4. 从MindIE容器内复制配置文件config.json到节点目录。
    1. 在节点上创建目录。

        ```shell
        mkdir -p /data/atlas_dls/public/infer/script/Qwen3-1.7B
        ```

    2. 启动容器，将目录“/data/atlas\_dls/public/infer/script/Qwen3-1.7B”挂载到容器中。

        ```shell
        docker run --rm -it \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B:/data/atlas_dls/public/infer/script/Qwen3-1.7B \
        <mindie image:tag>  /bin/bash
        ```

        请用户将\<mindie image:tag\>替换为实际镜像名和tag。

    3. 在容器内，将config.json复制到“/data/atlas\_dls/public/infer/script/Qwen3-1.7B”中。

        ```shell
        cp  $MIES_INSTALL_PATH/conf/config.json /data/atlas_dls/public/infer/script/Qwen3-1.7B/
        ```

        容器内环境变量MIES\_INSTALL\_PATH为MindIE Server的安装路径，默认为“/usr/local/Ascend/mindie/latest/mindie-service”，请用户替换为实际安装路径。

    4. 退出容器。

        ```shell
        exit
        ```

    5. 在节点的“/data/atlas\_dls/public/infer/script/Qwen3-1.7B”目录中查看config.json文件。

        ```shell
        ll
        ```

        回显示例如下：

        ```ColdFusion
        …
        -rw-r----- 1 root root 3920 11月  8 11:53 config.json
        …
        ```

5. 修改config.json文件。
    1. 打开config.json文件。

        ```shell
        vi /data/atlas_dls/public/infer/script/Qwen3-1.7B/config.json
        ```

    2. 按“i”进入编辑模式，按实际使用情况修改如下参数。参数说明详细请参见《MindIE LLM开发指南》中的“核心概念与配置 \> [配置参数说明（服务化）](https://www.hiascend.com/document/detail/zh/mindie/230/mindiellm/llmdev/mindie_service0285.html)”章节。

        ```json
        {
            …
            "ServerConfig" :
        {
                "ipAddress" : "127.0.0.1",
                "managementIpAddress" : "127.0.0.2",
                "port" : 1025,
                "managementPort" : 1026,
                "metricsPort" : 1027,
                …
                "httpsEnabled" : false,
                …
            },
         
        "BackendConfig" : {
            …
                "npuDeviceIds" : [[0,1]],
                …
                "ModelDeployConfig" :
                {
                    …
                    "truncation" : false,
                    "ModelConfig" : [
                        {
                            …
                            "modelName" : "qwen3",
                            "modelWeightPath" : "/job/model_weight/",
                            "worldSize" : 2,
                            …
                        }
                    ]
                },
                …
            }
        }
        ```

        其中，modelWeightPath为挂载到容器中的模型权重路径。

        >[!NOTICE] 
        >"httpsEnabled"表示是否开启HTTPS协议。设为"true"表示开启HTTPS协议，此时需要配置双向认证证书；设为"false"表示不开启HTTPS协议。推荐开启HTTPS协议，并参见《MindIE Motor开发指南》中的“配套工具 \> MindIE Service Tools \> [CertTools](https://www.hiascend.com/document/detail/zh/mindie/230/mindiemotor/motordev/mindie_service0312.html)”章节，配置开启HTTPS通信所需服务证书、私钥等证书文件。

    3. 按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。

6. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/inference/without-k8s/”目录下的启动脚本infer\_start.sh，放在节点目录“/data/atlas\_dls/public/infer/script/Qwen3-1.7B/”下，并对infer\_start.sh脚本进行编辑。

    1. 打开infer\_start.sh脚本。

        ```shell
        vi /data/atlas_dls/public/infer/script/Qwen3-1.7B/infer_start.sh
        ```

    2. 按“i”进入编辑模式，按实际情况修改脚本中的相关配置。

        ```shell
        …
        if [[ -z "${MIES_INSTALL_PATH}" ]]; then
            export MIES_INSTALL_PATH=/usr/local/Ascend/mindie/latest/mindie-service # 镜像中MindIE Server安装目录，若安装路径不一致，请用户自行修改
        fi
        …
        mkdir -p /job/script/alllog/
        INFER_LOG_PATH=/job/script/alllog/output_$(date +%Y%m%d_%H%M%S).log # 日志落盘路径
         
        # config.json
        export MIES_CONFIG_JSON_PATH=/job/script/config.json # 推理任务启动配置文件路径，容器启动时挂载进容器
        # （可选）其他用户自定义步骤
        …
        ```

    3. 按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。
    4. 增加脚本可执行权限。

        ```shell
        chmod +x infer_start.sh
        ```

    “/data/atlas\_dls/public/infer/“的目录结构如下：

    ```shell
    ├── model_weight
    │   └── Qwen3-1.7B
    └── script
        └── Qwen3-1.7B
            ├── config.json
            └── infer_start.sh
    ```

7. 启动容器，拉起MindIE任务。

    - 使用Ascend Docker Runtime挂载芯片和设备

        ```shell
        docker run -it -d --net=host --shm-size=1g \ 
        --name <container-name> \
        -e ASCEND_VISIBLE_DEVICES=0,1 \
        -v /usr/local/Ascend/driver:/usr/local/Ascend/driver:ro \
        -v /usr/local/sbin:/usr/local/sbin:ro \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B/:/job/script/ \
        -v /data/atlas_dls/public/infer/model_weight/Qwen3-1.7B/:/job/model_weight/ \
        --entrypoint /job/script/infer_start.sh  <mindie image:tag>  <restart_times>
        ```

    - 不使用Ascend Docker Runtime挂载芯片和设备

        ```shell
        docker run -it -d --net=host --shm-size=1g \
        --name <container-name> \
        --device=/dev/davinci0:rwm \
        --device=/dev/davinci1:rwm \
        --device=/dev/davinci_manager:rwm \
        --device=/dev/devmm_svm:rwm \
        --device=/dev/hisi_hdc:rwm \
         -v /usr/local/sbin/npu-smi:/usr/local/sbin/npu-smi \
        -v /usr/local/Ascend/driver:/usr/local/Ascend/driver:ro \
        -v /usr/local/sbin:/usr/local/sbin:ro \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B/:/job/script/ \
        -v /data/atlas_dls/public/infer/model_weight/Qwen3-1.7B/:/job/model_weight/ \
        --entrypoint /job/script/infer_start.sh  <mindie image:tag>  <restart_times>
        ```

       上述配置说明如下：
      >- \<container-name\>表示容器名称。
      >- 请用户将\<mindie image:tag\>替换为实际镜像名和tag。
      >- \<restart\_times\>作为参数传入infer\_start.sh中，表示服务重启次数，需替换为数字，不填默认为0。超过重启次数会退出容器。
      >- 请用户按需自行修改环境变量ASCEND\_VISIBLE\_DEVICES的值，以挂载不同数量芯片。芯片ID需要与config.json中npuDeviceIds字段包含的芯片ID保持一致。
      >- 请用户自行增删“--device”参数，以挂载不同数量芯片和设备。芯片ID需要与config.json中npuDeviceIds字段包含的芯片ID保持一致。
     
   >[!NOTE] 
   >启动容器后，若报错"OpenBLAS blas_thread_int: pthread_create failed for thread 1 of 128: Operation not permitted"，即OpenBLAS尝试创建多线程失败，可能原因是seccomp阻止了pthread相关系统的调用，此时可以在Docker启动命令中增加“--security-opt seccomp=unconfined --security-opt no-new-privileges”参数解决。

8. 查看容器日志。

    ```shell
    docker logs -f <container-name>
    ```

    若显示如下信息，说明容器启动成功。

    ```ColdFusion
    …
    Daemon start success!
    …
    ```

9. 新建终端窗口，输入以下命令，访问服务。若请求成功返回，表示推理服务部署成功。

    ```shell
    curl -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -X POST -d '{
        "model": "<model_name>", 
    "messages": [ 
            {"role": "system", "content": "you are a helpful assistant."},
            { "role": "user", "content": "How many r are in the word \"strawberry\"" } 
        ], 
        "max_tokens": 256, 
        "stream": false,
        "do_sample": true,
        "ignore_eos": true, 
        "temperature": 0.6,
        "top_p": 0.95,
        "top_k": 20,
        "stream": false }' \
    http://<ipAddress>:<port>/v1/chat/completions
    ```

    >[!NOTE] 
    >- \<model\_name\>需替换为config.json中modelName字段的值。
    >- \<ipAddress\>需替换为config.json中ipAddress字段的值。
    >- \<port\>需替换为config.json中port字段的值。

10. 测试服务故障后是否自动重启。
    1. 在节点上构造服务故障。

        ```shell
        # 查询NPU卡上的进程信息，包含进程号
        npu-smi info
        # 杀进程，构造故障，请将<process_id>替换为进程号
        kill -9 <process_id>
        ```

    2. 查看容器日志。

        ```shell
        docker logs -f <container-name>
        ```

        若显示如下信息，说明重启成功。

        ```ColdFusion
        Daemon is killing...
        …
        [EntryPoint Script Log]running job failed. exit code: 137
        [EntryPoint Script Log]restart mindie service daemon, cur: 0, max: 1
        …
        Daemon start success!
        ```

11. 停止容器。

    ```shell
    docker stop <container-name>
    ```
