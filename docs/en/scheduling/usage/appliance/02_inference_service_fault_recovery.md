# Configuring Inference Service Fault Recovery<a name="ZH-CN_TOPIC_0000002511630975"></a>

In scenarios where inference processes are deployed in an integrated appliance or without Kubernetes, there is no effective recovery mechanism after a process failure. This chapter provides an example of automatic recovery after inference service failures. In this example, the startup script acts as the container entrypoint to automatically launch the inference process, monitor its status, and restart it upon failure.

- Support single-node MindIE Server inference.
- Do not support multi-node MindIE Server inference. This is because restarting the inference process in just one container is insufficient to recover the service.

**Procedure<a name="section169801610181818"></a>**

The following uses the Qwen3-1.7B model as an example.

1. Obtain the MindIE container image.
    - Method 1: Go to the [MindIE Image Download](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f) page in the Ascend image repository and download the MindIE image.
    - Method 2: Refer to the "Installing MindIE > [Method 3: Containerized Installation](https://gitcode.com/Ascend/MindIE-LLM/blob/v3.0.0/docs/en/user_guide/install/source/installation_in_containerized.md)" section in the *MindIE Installation Guide* to prepare the image yourself.

2. View the MindIE image on the node.

    ```shell
    docker images |grep mindie
    ```

    The output is as follows:

    ```ColdFusion
    …
    swr.cn-south-1.myhuaweicloud.com/ascendhub/mindie   2.1.RC2-800I-A2-py311-openeuler24.03-lts   a4708118cd12        6 weeks ago         16GB
    …
    ```

3. Obtain the Qwen3-1.7B model weights.

    ```shell
    # Create a directory to save the model weights
    mkdir -p /data/atlas_dls/public/infer/model_weight
    cd /data/atlas_dls/public/infer/model_weight/
    # If git-lfs is not installed, install it first. git-lfs is a Git extension specifically designed for managing large files and binary files.
    yum install -y git-lfs
    # Enable git-lfs
    git lfs install
    # Download weights
    git clone https://www.modelscope.cn/Qwen/Qwen3-1.7B.git
    # Modify weight file permissions
    chmod -R 750 Qwen3-1.7B/
    # (Optional) If a non-root user image is used, the weight path must be owned by the default user 1000 in the image.
    chown -R 1000:1000 Qwen3-1.7B/
    ```

    >[!NOTE]
    >After downloading certain models, weight quantization is also required. For details, see the README of each model in [ModelZoo-PyTorch](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/MindIE/LLM).

4. Copy the configuration file `config.json` from the MindIE container to the node directory.
    1. Create a directory on the node.

        ```shell
        mkdir -p /data/atlas_dls/public/infer/script/Qwen3-1.7B
        ```

    2. Start the container and mount the directory `/data/atlas_dls/public/infer/script/Qwen3-1.7B` into the container.

        ```shell
        docker run --rm -it \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B:/data/atlas_dls/public/infer/script/Qwen3-1.7B \
        <mindie image:tag>  /bin/bash
        ```

        Replace `<mindie image:tag>` with the actual image name and tag.

    3. In the container, copy `config.json` to `/data/atlas_dls/public/infer/script/Qwen3-1.7B`.

        ```shell
        cp  $MIES_INSTALL_PATH/conf/config.json /data/atlas_dls/public/infer/script/Qwen3-1.7B/
        ```

        The environment variable `MIES_INSTALL_PATH` in the container is the installation path of MindIE Server, which defaults to `/usr/local/Ascend/mindie/latest/mindie-service`. Replace it with the actual installation path.

    4. Exit the container.

        ```shell
        exit
        ```

    5. On the node, view the `config.json` file in the `/data/atlas_dls/public/infer/script/Qwen3-1.7B` directory.

        ```shell
        ll
        ```

        Output:

        ```ColdFusion
        …
        -rw-r----- 1 root root 3,920 Nov  8 11:53 config.json
        …
        ```

5. Modify the `config.json` file.
    1. Open the `config.json` file.

        ```shell
        vi /data/atlas_dls/public/infer/script/Qwen3-1.7B/config.json
        ```

    2. Press `i` to enter insert mode, and modify the following parameters as required. For details on the parameters, see the "Core Concepts and Configuration > [Configuration Parameter Description (Serving)](https://www.hiascend.com/document/detail/en/mindie/300/mindiellm/llmdev/user_guide/user_manual/service_parameter_configuration.md)" section in the *MindIE LLM Development Guide*.

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

        Here, `modelWeightPath` is the model weight path mounted to the container.

    3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit.

6. Go to the [mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy) repository, switch to the corresponding version branch according to [mindcluster-deploy Open-Source Repository Version Description](../../references/appendix.md), obtain the startup script `infer_start.sh` from the `samples/inference/without-k8s/` directory, place it in the node directory `/data/atlas_dls/public/infer/script/Qwen3-1.7B/`, and edit the `infer_start.sh` script.

    1. Open the `infer_start.sh` script.

        ```shell
        vi /data/atlas_dls/public/infer/script/Qwen3-1.7B/infer_start.sh
        ```

    2. Press `i` to enter insert mode, and modify the relevant configurations in the script according to the actual situation.

        ```shell
        …
        if [[ -z "${MIES_INSTALL_PATH}" ]]; then
            export MIES_INSTALL_PATH=/usr/local/Ascend/mindie/latest/mindie-service # MindIE Server installation directory in the image. If the installation path is different, modify it accordingly.
        fi
        …
        mkdir -p /job/script/alllog/
        INFER_LOG_PATH=/job/script/alllog/output_$(date +%Y%m%d_%H%M%S).log # Log flushing path

        # config.json
        export MIES_CONFIG_JSON_PATH=/job/script/config.json # Inference job startup configuration file path, which is mounted into the container when you start the container.
        # (Optional) Other user-defined steps
        …
        ```

    3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit editing.
    4. Add executable permission to the script.

        ```shell
        chmod +x infer_start.sh
        ```

    The directory structure of `/data/atlas_dls/public/infer/` is as follows:

    ```shell
    ├── model_weight
    │   └── Qwen3-1.7B
    └── script
        └── Qwen3-1.7B
            ├── config.json
            └── infer_start.sh
    ```

7. Start the container and launch the MindIE task.

    - Use Ascend Docker Runtime to mount chips and devices

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

    - Mount chips and devices without using Ascend Docker Runtime

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

    The preceding configuration is described as follows:
    - `<container-name>` indicates the container name.
    - Replace `<mindie image:tag>` with the actual image name and tag.
    - `<restart_times>` is passed as a parameter to `infer_start.sh`, indicating the number of service restart attempts. It must be replaced with a number. If left blank, the default value is `0`. The container will exit if the number of restart attempts is exceeded.
    - Modify the value of the environment variable `ASCEND\_VISIBLE\_DEVICES` as needed to mount different numbers of chips. The chip IDs must be consistent with the chip IDs contained in the `npuDeviceIds` field in `config.json`.
    - Add or remove the `--device` parameter as needed to mount different numbers of chips and devices. The chip IDs must be consistent with the chip IDs contained in the `npuDeviceIds` field in `config.json`.

   >[!NOTE]
   >After the container is started, if the error "OpenBLAS blas_thread_int: pthread_create failed for thread 1 of 128: Operation not permitted" is reported, it means that OpenBLAS failed to create multiple threads. The likely cause is that seccomp is blocking system calls related to pthread. In this case, add the `--security-opt seccomp=unconfined --security-opt no-new-privileges` parameter to the Docker startup command to resolve the issue.

8. View container logs.

    ```shell
    docker logs -f <container-name>
    ```

    If the following information is displayed, the container has started successfully.

    ```ColdFusion
    …
    Daemon start success!
    …
    ```

9. Open a new terminal window and enter the following command to access the service. If the request returns successfully, the inference service has been deployed successfully.

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
    >- `<model_name>` must be replaced with the modelName field value in config.json.
    >- `<ipAddress>` must be replaced with the `ipAddress` field value in `config.json`.
    >- `<port>` must be replaced with the `port` field value in `config.json`.

10. Test whether the service automatically restarts after a fault occurs.
    1. Construct a service fault on the node.

        ```shell
        # Query the process information on the NPU, including the process ID.
        npu-smi info
        # Kill the process to simulate a fault. Replace <process_id> with the process ID.
        kill -9 <process_id>
        ```

    2. View the container logs.

        ```shell
        docker logs -f <container-name>
        ```

        If the following information is displayed, the restart is successful.

        ```ColdFusion
        Daemon is killing...
        …
        [EntryPoint Script Log]running job failed. exit code: 137
        [EntryPoint Script Log]restart mindie service daemon, cur: 0, max: 1
        …
        Daemon start success!
        ```

11. Stop the container.

    ```shell
    docker stop <container-name>
    ```
