# Configuring Online Stress Testing<a name="ZH-CN_TOPIC_0000002511426487"></a>

## PyTorch Scenario (Based on MindSpeed-LLM)<a name="ZH-CN_TOPIC_0000002479386572"></a>

This part describes the key steps for configuring online stress testing. For details about the features, usage constraints, and supported product models of online stress testing, see [Online Stress Testing](../01_working_principle.md#ZH-CN_TOPIC_0000002479226572).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- Complete the installation of the following components on the corresponding nodes: Ascend Docker Runtime, Ascend Operator, ClusterD, Ascend Device Plugin, and Volcano (the versions of the preceding MindCluster components must match the TaskD version). For details about the installation procedure, see [Installation and Deployment](../../../03_installation_guide/02_installation/00_helm_installation.md).
- Install TorchNPU (7.1.RC1 or later), CANN (8.2.RC1 or later), TaskD, and MindIO (7.2.RC1 or later) in the container. For details, see [Building a MindSpeed-LLM Training Image (PyTorch)](../../05_fault_recovery/01_resumable_training/03_examples_and_verification/01_pytorch_examples_and_verification.md#ZH-CN_TOPIC_0000002511426469).

**Procedure<a name="section188080175496"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the script and start TaskD Worker within the training process.

    1. Start TaskD Manager.
        1. Create a `manager.py` file and place it in the training code directory (you can adjust it based on the actual situation, but note that the training script must be modified accordingly). The content of the `manager.py` file is as follows.

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # Total number of nodes
            proc_per_node=XX     # Number of training processes per node

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >For detailed parameter descriptions in the `manager.py` file, see [def init\_taskd\_manager\(config:dict\) -\> bool:](../../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path, and the error.log path must be created in advance.
            fi

            torchrun ...
            ```

    2. Start TaskD Worker.

        Modify the `QWEN3_for_PyTorch_2.7_code/mindspeed_llm/training/training.py` file and add the following bold fields.

        <pre codetype="Python">
        def pretrain(train_valid_test_dataset_provider,
                     model_provider,
                     model_type,
                     forward_step_func,
                     process_non_loss_data_func=None,
                     extra_args_provider=None,
                     args_defaults={}):
            print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                time.time() - _TRAIN_START_TIME))
            print_datetime('after megatron is initialized')
            <strong>import torch.distributed as dist</strong>
            <strong>if dist.is_initialized():</strong>
               <strong>rank = dist.get_rank()</strong>
               <strong>from taskd.api.taskd_worker_api import init_taskd_worker</strong>
               <strong>from taskd.api.taskd_worker_api import start_taskd_worker</strong>
               <strong>init_taskd_worker(rank,5000,"pt")</strong>
               <strong>start_taskd_worker()</strong>
            app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
            one_logger_utils.on_pretrain_start()</pre>

    >[!NOTE]
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload specified `.so` files. An example is shown below.
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so:$LD_PRELOAD
    >```
    >
    >- `libmspti.so`: This `.so` file is provided by MindStudio and integrated within the CANN package. Default installation path: `/usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This `.so` file is provided by TaskD, and the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. The TaskD installation path can be queried using the following command. The `Location` field in the output is the TaskD installation path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

2. Modify the job YAML.

    Add the following bold fields to the job YAML to enable process-level rescheduling, and modify the necessary configurations.

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>recover-strategy: "recover"    # Available recovery strategy for the job. The value recover indicates that process-level rescheduling is enabled. If a problem occurs after stress testing, this field can be used to enable process-level rescheduling.</strong>
         ...
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - |
                        cd /job/code;
                        chmod +x scripts/train_start.sh;
                        bash scripts/train_start.sh
                    env:
                      <strong>- name: POD_IP                        # Configures the communication address for MindIO.</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # Used for MindIO communication. If this parameter is not configured, it will affect the normal startup of training jobs.</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # Adds port 9601 used for TaskD communication under all Pods.</strong>
                        <strong>name: taskd-port</strong>
        ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    args:
                      - |
                        cd /job/code;
                        chmod +x scripts/train_start.sh;
                        bash scripts/train_start.sh
                    env:
                      <strong>- name: POD_IP                        # Configures the communication address for MindIO.</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # Used for MindIO communication. If this parameter is not configured, it will affect the normal startup of training jobs.</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # Adds port 9601 used for TaskD communication under all Pods.</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>

## MindSpore Scenario (Based on MindFormers)<a name="ZH-CN_TOPIC_0000002479226554"></a>

This part describes the key steps for configuring online stress testing. For details about the features, usage constraints, and supported product models of online stress testing, see [Online Stress Testing](../01_working_principle.md#ZH-CN_TOPIC_0000002479226572).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533-duplicate-2"></a>**

- Complete the installation of the following components on the corresponding nodes: Ascend Docker Runtime, Ascend Operator, ClusterD, Ascend Device Plugin, and Volcano (the versions of the preceding MindCluster components must match TaskD). For details about the installation procedure, see [Installation and Deployment](../../../03_installation_guide/02_installation/00_helm_installation.md).
- Install MindSpore (2.7.0 or later), CANN (8.2.RC1 or later), TaskD, and MindIO (7.2.RC1 or later) in the container. For details, see [Building a MindFormers Training Image (MindSpore)](../../05_fault_recovery/01_resumable_training/03_examples_and_verification/02_mindspore_examples_and_verification.md#ZH-CN_TOPIC_0000002511426469-mindspore).

**Procedure<a name="section9479182019317"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the training script, start TaskD Proxy in the management process, and start TaskD Worker in the training process.
    1. Start TaskD Manager.
        1. Create a `manager.py` file and place it in the training code directory (you can adjust it based on the actual situation, but note that the training script must be modified accordingly). The content of the `manager.py` file is as follows.

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # Total number of nodes
            proc_per_node=XX     # Number of training processes per node

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >For detailed parameter descriptions in the `manager.py` file, see [def init\_taskd\_manager\(config:dict\) -\> bool:](../../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            ```shell
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path, and the error.log path must be created in advance.
            fi

            msrun ...
            ```

    2. Start TaskD Worker. Modify the `./mindformers/trainer/base_trainer.py` file and add the following bold fields.

        <pre codetype="Python">
        def training_process(
                self,
                config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                network: Optional[Union[Cell, PreTrainedModel]] = None,
                dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                optimizer: Optional[Optimizer] = None,
                callbacks: Optional[Union[Callback, List[Callback]]] = None,
                compute_metrics: Optional[Union[dict, set]] = None,
                **kwargs):
            ……
            ……
            logger.info(".........Starting Training Model..........")
            if get_real_rank() % 8 == 0:
                pprint(config)
            logger.info(".........Model Compiling, Please Wait a Moment...........")
            ......
            <strong>try:</strong>
                <strong>rank = get_rank()</strong>
                <strong>from taskd.api.taskd_worker_api import init_taskd_worker</strong>
                <strong>from taskd.api.taskd_worker_api import start_taskd_worker</strong>
                <strong>init_taskd_worker(rank,5000,"ms")</strong>
                <strong>start_taskd_worker()</strong>
            <strong>except Exception as e:</strong>
                <strong>print("failed to call mindcluster taskd")</strong>
            model.train(config.runner_config.epochs, dataset,
                        callbacks=callbacks,
                        dataset_sink_mode=config.runner_config.sink_mode,
                        sink_size=config.runner_config.sink_size,
                        initial_epoch=config.runner_config.initial_epoch)</pre>

2. Modify the training framework code to enable the online stress testing switch.

    Edit the `QWEN3_for_MS_code/scripts/msrun_launcher.sh` startup script and add the following field to the code.

    ```shell
    export MS_ENABLE_TFT="{TTP:1,TSP:1}"           # Enable dying gasp and online stress testing.
    ```

    >[!NOTE]
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload specified `.so` files. An example is shown below.
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so:$LD_PRELOAD
    >```
    >
    >- `libmspti.so`: This `.so` file is provided by MindStudio and integrated within the CANN package. Default installation path: `/usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This `.so` file is provided by TaskD, and the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. The TaskD installation path can be queried using the following command. The `Location` field in the output is the TaskD installation path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

3. Modify the job YAML.

    Add the following bold fields to the job YAML to enable process-level rescheduling, and modify the necessary configurations.

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>recover-strategy: "recover"    # Available recovery strategy for the job. The value recover indicates that process-level rescheduling is enabled. If a problem occurs after stress testing, this field can be used to enable process-level rescheduling.</strong>
         ...
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    command:                           # training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
                    env:
                      <strong>- name: POD_IP                        # Configures the communication address for MindIO.</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # Used for MindIO communication. If this parameter is not configured, it will affect the normal startup of training jobs.</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # Adds port 9601 used for TaskD communication under all Pods.</strong>
                        <strong>name: taskd-port</strong>
        ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    ...
                    command:                           # training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
                    env:
                      <strong>- name: POD_IP                        # Configures the communication address for MindIO.</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # Used for MindIO communication. If this parameter is not configured, it will affect the normal startup of training jobs.</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # Adds port 9601 used for TaskD communication under all Pods.</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>
