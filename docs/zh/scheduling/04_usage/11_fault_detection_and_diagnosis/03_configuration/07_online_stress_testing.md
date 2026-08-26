# 配置在线压测<a name="ZH-CN_TOPIC_0000002511426487"></a>

## PyTorch场景（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002479386572"></a>

本章节将指导用户了解配置在线压测的关键步骤。在线压测的特性介绍、使用约束、支持的产品型号等请参见[在线压测](../01_working_principle.md#在线压测)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- 在相应节点上完成以下组件的安装：Ascend Docker Runtime、Ascend Operator、ClusterD、Ascend Device Plugin和Volcano（以上MindCluster组件版本均需与TaskD配套），详细安装步骤请参见[安装部署](../../../03_installation_guide/02_installation/00_helm_installation.md)。
- 在容器内安装TorchNPU（7.1.RC1及以上版本）、CANN（8.2.RC1及以上版本）、TaskD和MindIO（7.2.RC1及以上版本），详情请参见[制作MindSpeed-LLM训练镜像（PyTorch框架）](../../04_resumable_training/04_using_resumable_training_on_the_cli.md#ZH-CN_TOPIC_0000002479386504)。

**操作步骤<a name="section188080175496"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD Manager，在训练进程内部拉起TaskD Worker。

    1. 拉起TaskD Manager。
        1. 创建manager.py文件，放在训练代码目录下（可根据实际情况自行调整，但需注意修改训练脚本）。manager.py文件内容如下所示。

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码，拉起TaskD Manager。

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
            fi

            torchrun ...
            ```

    2. 拉起TaskD Worker。

        修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，增加如下加粗字段。

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
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so:$LD_PRELOAD
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

2. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启进程级别重调度，并修改必要的配置。

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>recover-strategy: "recover"    # 任务可用恢复策略，取值为recover，表示开启进程级别重调度。压测之后有问题的话，可以使用本字段开启进程级重调度。</strong>
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
                      <strong>- name: POD_IP                        # 配置MindIO的通信地址。</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # 在所有的Pod下增加TaskD通信使用的端口9601</strong>
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
                      <strong>- name: POD_IP                        # 配置MindIO的通信地址。</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # 在所有的Pod下增加TaskD通信使用的端口9601</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>

## MindSpore场景（基于MindFormers）<a name="ZH-CN_TOPIC_0000002479226554"></a>

本章节将指导用户了解配置在线压测的关键步骤。在线压测的特性介绍、使用约束、支持的产品型号等请参见[在线压测](../01_working_principle.md#在线压测)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533-duplicate-2"></a>**

- 在相应节点上完成以下组件的安装：Ascend Docker Runtime、Ascend Operator、ClusterD、Ascend Device Plugin和Volcano（以上MindCluster组件版本均需与TaskD配套），详细安装步骤请参见[安装部署](../../../03_installation_guide/02_installation/00_helm_installation.md)。
- 在容器内安装MindSpore（2.7.0及以上版本）、CANN（8.2.RC1及以上版本）、TaskD和MindIO（7.2.RC1及以上版本），详情请参见[制作MindFormers训练镜像（MindSpore框架）](../../04_resumable_training/04_using_resumable_training_on_the_cli.md#ZH-CN_TOPIC_0000002511426451)。

**操作步骤<a name="section9479182019317"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD Worker。
    1. 拉起TaskD Manager。
        1. 创建manager.py文件，放在训练代码目录下（可根据实际情况自行调整，但需注意修改训练脚本），manager.py文件内容如下所示。

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         # 用户填入任务节点总数
            proc_per_node=XX     # 用户填入任务每个节点的训练进程数量

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码拉起TaskD Manager。

            ```shell
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
            fi

            msrun ...
            ```

    2. 拉起TaskD Worker。修改./mindformers/trainer/base\_trainer.py文件，增加如下加粗字段。

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

2. 修改训练框架代码，打开在线压测开关。

    编辑启动脚本QWEN3\_for\_MS\_code/scripts/msrun\_launcher.sh文件，在代码中增加如下字段。

    ```shell
    export MS_ENABLE_TFT="{TTP:1,TSP:1}"           # 开启临终遗言和在线压测
    ```

    >[!NOTE]
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so:$LD_PRELOAD
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

3. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启进程级别重调度，并修改必要的配置。

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>recover-strategy: "recover"    # 任务可用恢复策略，取值为recover，表示开启进程级别重调度。压测之后有问题的话，可以使用本字段开启进程级重调度。</strong>
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
                      <strong>- name: POD_IP                        # 配置MindIO的通信地址。</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # 在所有的Pod下增加TaskD通信使用的端口9601</strong>
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
                      <strong>- name: POD_IP                        # 配置MindIO的通信地址。</strong>
                        <strong>valueFrom:</strong>
                          <strong>fieldRef:</strong>
                            <strong>fieldPath: status.podIP         # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。</strong>
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601                 # 在所有的Pod下增加TaskD通信使用的端口9601</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>
