# Configuring Fault Handling Policies<a name="ZH-CN_TOPIC_0000002479386478"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T08:01:34.739Z pushedAt=2026-06-09T09:02:55.493Z -->

## Configuring Job-Level Rescheduling<a name="ZH-CN_TOPIC_0000002479226580"></a>

Job-level Rescheduling is enabled by default. You only need to complete the steps for building an image and preparing the job YAML. For details about the feature introduction, usage constraints, supported product models, and principles of Job-level rescheduling, see [Job-Level Rescheduling](./01_solutions_principles.md#job-level-rescheduling).

**Preparing the Job YAML<a name="zh-cn_topic_0000002098814658_section463203519254"></a>**

In the job YAML, add the following fields to enable Job-level rescheduling.

```Yaml
...
metadata:
   labels:
     ...
     fault-scheduling: "force"
```

## Configuring Pod-Level Rescheduling

This part guides you through the key steps for configuring Pod-level rescheduling. For details on the features, usage constraints, supported product models, and principles of Pod-level rescheduling, see [Pod-Level Rescheduling](./01_solutions_principles.md#pod-level-rescheduling).

**Building an Image <a name="zh-cn_topic_0000002098654822_section11751140165911"></a>**

Use a Dockerfile to build a container image and add a startup command. An example is shown below.

```shell
# Adaptation script to MindCluster resumable training. TASKD_WHL is the path to the TaskD whl installation package. Fill in the actual path accordingly.
# Optional. Under the PyTorch framework, the following commands must be configured when using graceful fault tolerance, Pod-level rescheduling, or process-level rescheduling.
RUN pip install $TASKD_WHL
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# Optional. Under the MindSpore framework, the following commands must be configured when using Pod-level rescheduling.
RUN pip install $TASKD_WHL
```

**Preparing the Job YAML<a name="zh-cn_topic_0000002098654822_section027517423166a"></a>**

In the job YAML, add the following fields to enable Pod-level rescheduling, modify the container port, and add port 9601 for TaskD communication under all Pods.

<pre codetype="yaml">
...
metadata:
   labels:
     ...
     <strong>pod-rescheduling: "on"</strong>
     <strong>fault-scheduling: "force"   # You can choose force or grace based on the actual situation. When configured as force, the Pod cannot use the host network.</strong>
...
        spec:
...
           containers:
...
             <strong>ports:</strong>
               <strong>- containerPort: 9601</strong>
                 <strong>name: taskd-port</strong>
...</pre>

**Adapting the Training Script<a name="zh-cn_topic_0000002098654822_section17330181621710"></a>**

1. In the startup script (for example, `train_start.sh`), add the following bold fields as shown in the example below.

    <pre codetype="shell">
    ...
    <strong>export MS_ENABLE_TFT="{RSC:1}"    # Configure this field to enable Pod-level rescheduling in MindSpore scenarios</strong>
    ...
    # Optional. In PyTorch scenarios, set the number of restarts within the container and the training process monitoring interval.
       logger "server id is: ""${server_id}"
       if [ "${framework}" == "PyTorch" ]; then
         get_env_for_pytorch_multi_node_job
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT <strong>--max_restarts 32767</strong>" </pre>

    Where, `--max_restarts` specifies the maximum number of fault triggers allowed within the container, expressed as an integer. If this limit is exceeded, the PyTorch training process will exit immediately. If this parameter is not configured, the default value is `32767`.

2. After the distributed environment initialization is complete and the global rank is obtained, modify the training script to start TaskD Manager in the training script.
    1. Create a `manager.py` file and place it in the current directory when calling the training script. The content of the `manager.py` file is as follows.

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os

        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # Total number of task nodes
        proc_per_node=XX     # Number of training processes per node

        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE]
        >For detailed parameter descriptions in the `manager.py` file, see [def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

    2. Add the following code to the training script (for example, `train_start.sh`) to start TaskD Manager. In the following code:

        - The two statements `TASKD_SO_PATH` and export `LD_PRELOAD` are used to configure the path of `libtaskd.so` (from the TaskD installation) into the environment variable `LD_PRELOAD`. If these two statements are not configured successfully, you can manually run the `pip show taskd` command to obtain the `Location` value, append `/taskd/python/cython_api/libs/libtaskd.so`, and then set it via `export`.
        - `TASKD_PROCESS_ENABLE` configuration instructions: If `recover-strategy` in the job YAML does not configure a recovery policy and does not enable hot switching, you need to configure `export TASKD_PROCESS_ENABLE="off`"; if `recover-strategy` is configured or hot switching is enabled, you do not need to configure `export TASKD\_PROCESS\_ENABLE="off`.

        ```shell
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="off"
        # Under the PyTorch framework
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path. The error.log log path must be created in advance.
        fi
        # Under the MindSpore framework
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path. The error.log log path must be created in advance.
        fi
        ```

## Configuring Process-Level Rescheduling<a name="ZH-CN_TOPIC_0000002511426407"></a>

This part describes the key steps for configuring process-level rescheduling. For details about the feature introduction, usage constraints, supported product models, and principles of process-level rescheduling, see [Process-Level Rescheduling](./01_solutions_principles.md#process-level-rescheduling).

**Building an Image <a name="zh-cn_topic_0000002134293721_section18253151810133"></a>**

Use a Dockerfile to build a container image and add a startup command.

```shell
# Adaptation script to MindCluster lossless resumable training. TASKD_WHL is the path to the TaskD whl installation package, and MINDIO_TTP_PKG is the path to the MindIO whl installation package. Fill them in according to the actual situation.
# Optional. Under the PyTorch framework, the following commands must be configured when using graceful fault tolerance, Pod-level rescheduling, or process-level rescheduling.
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# Optional. Under the MindSpore framework, the following commands must be configured when using process-level rescheduling.
RUN pip3 install $MINDIO_TTP_PKG
RUN pip3 install $TASKD_WHL
```

**Preparing the Job YAML<a name="zh-cn_topic_0000002134293721_section2492121411271"></a>**

In the job YAML, modify the container port and add port 9601 for TaskD communication under all Pods.

```Yaml
...
        spec:
...
           containers:
...
             ports:
               - containerPort: 9601
                 name: taskd-port
...
```

In the job YAML, add the following fields to enable process-level rescheduling. `recover-strategy` is the strategy used for training process recovery, where `recover` indicates enabling process-level recovery.

Currently, process-level rescheduling supports the following two scenarios. Choose one based on the actual usage scenario.

- Scenario 1: Migrate the faulty Pod to a healthy node after a fault occurs

    <pre codetype="yaml">
    ...
    metadata:
       labels:
         ...
         <strong>fault-scheduling: "grace"</strong>
     ...
    ...
       annotations:
         ...
         <strong>recover-strategy: "recover"   # Recovery strategies (retry: process-level online recovery; recover: process-level rescheduling; recover-in-place: process-level in-place recovery; elastic-training: elastic training; dump: save dying gasps; exit: exit training). Six strategies can be combined arbitrarily, separated by commas.</strong>
     ...
    ...
    spec:
      replicaSpecs:
        Master:
          template:
            spec:
              containers:
              - name: ascend       # Do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
        Worker:
          template:
            spec:
              containers:
              - name: ascend # Do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
    ...</pre>

- Scenario 2: Do not migrate the faulty Pod after a fault; only restart the faulty process

    <pre codetype="yaml">
    ...
    metadata:
       labels:
         ...
         <strong>fault-scheduling: "grace"</strong>
     ...
    ...
       annotations:
         ...
         <strong>recover-strategy: "recover-in-place"   # Recovery strategies (retry: process-level online recovery; recover: process-level rescheduling; recover-in-place: process-level in-place recovery; elastic-training: elastic training; dump: save last words; exit: exit training). Six strategies can be combined arbitrarily, separated by commas.</strong>
     ...
    ...
    spec:
      replicaSpecs:
        Master:
          template:
            spec:
              containers:
              - name: ascend # Do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
        Worker:
          template:
            spec:
              containers:
              - name: ascend # Do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
    ...</pre>

**Adapting the Training Script<a name="zh-cn_topic_0000002134293721_section1829103214273"></a>**

1. (Optional) In the startup script (for example, `train_start.sh`), configure the `--max_restarts` parameter. An example is shown below.

    <pre codetype="shell">
    # In PyTorch scenarios, set the training process monitoring interval.
    ...
       logger "server id is: ""${server_id}"
       if [ "${framework}" == "PyTorch" ]; then
         get_env_for_pytorch_multi_node_job
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT  <strong>--max_restarts 32767</strong>"
    ...</pre>

     Here, `--max_restarts` specifies the maximum number of fault triggers allowed within the container, expressed as an integer. If this limit is exceeded, the PyTorch training process will exit immediately. If this parameter is not configured, the default value is 32767.

2. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the training script.
    1. Create a `manager.py` file in the current directory where the training script is invoked. The content of the `manager.py` file is as follows.

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os

        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # Total number of nodes
        proc_per_node=XX     # Number of training processes per node

        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE]
        >For details about the parameters in the manager.py file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

3. Add the following code to the training script (for example, `train_start.sh`) to start TaskD Manager. In the following code, the two statements `TASKD_SO_PATH` and `export LD_PRELOAD` configure the path of `libtaskd.so` (from the TaskD installation) into the environment variable `LD_PRELOAD`. If these two statements fail to configure successfully, you can manually run the `pip show taskd` command to obtain the value of Location, append `/taskd/python/cython_api/libs/libtaskd.so` to it, and then set it via `export`.

        ```shell
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="on"
        # For PyTorch Framework
        if [[ "${RANK}" == 0 ]]; then
           export MASTER_ADDR=${POD_IP}
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of `manager.py` is determined by the current path, and the `error.log` path must be created in advance.
        fi
        # For MindSpore Framework
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of `manager.py` is determined by the current path, and the `error.log` path must be created in advance.
        fi
        ```

## Configuring Process-Level Online Recovery<a name="ZH-CN_TOPIC_0000002479386492"></a>

This part describes the key steps for configuring process-level online recovery. For details about the features, usage constraints, supported product models, and principles of process-level online recovery, see [Process-Level Online Recovery](./01_solutions_principles.md#process-level-online-recovery).

**Build Image<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

Use a Dockerfile to build a container image and add a startup command.

```shell
# Adaptation script to MindCluster resumable training. TASKD_WHL is the path to the TaskD whl installation package, and MINDIO_TTP_PKG is the path to the MindIO whl installation package. Fill them in according to the actual situation.
# Optional. Under the PyTorch framework, you must configure the following commands when using graceful fault tolerance, Pod-level rescheduling, process-level rescheduling, or process-level online recovery.
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# Optional. Under the MindSpore framework, you must configure the following commands when using process-level online recovery.
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
```

**Preparing the Job YAML<a name="zh-cn_topic_0000002134174097_section98717593512"></a>**

In the job YAML, add the following bold fields to enable process-level recovery, modify the container port, and add port 9601 for TaskD communication under all Pods.

<pre codetype="yaml">
...
   labels:
     ...
     <strong>fault-scheduling: "grace"</strong>
 ...
...
   annotations:
     ...
     <strong>recover-strategy: "retry"    # Recovery strategies (retry: process-level online recovery; recover: process-level rescheduling; recover-in-place: process-level in-place recovery; elastic-training: elastic training; dump: save last words; exit: exit training). Six strategies can be combined arbitrarily, separated by commas.</strong>
 ...
...
spec:
  replicaSpecs:
    Master:
      template:
        spec:
          containers:
          - name: ascend # Do not modify
            ...
            args:
              - |
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
            <strong>ports:</strong>
               <strong>- containerPort: 9601</strong>
                 <strong>name: taskd-port</strong>
...
    Worker:
      template:
        spec:
          containers:
          - name: ascend # Do not modify
            ...
            args:
              - |
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
            <strong>ports:</strong>
               <strong>- containerPort: 9601</strong>
                 <strong>name: taskd-port</strong>
...</pre>

In the MindSpore scenario, you need to modify the model parameter configuration YAML. Open the `QWEN3_for_MS_code/configs/qwen3/pretrain_qwen3_32b_4k.yaml` file and add the following bold fields.

<pre codetype="yaml">
# mindspore context init config
context:
  mode: 0  #0--Graph Mode; 1-Pynative Mode
  device_target: "Ascend"
  graph_kernel_flags: "--disable_pass=cluster.floatstatus_fusion,preprocess.depend_elimination"
  max_call_depth: 10000
  max_device_memory: "59GB"
  mempool_block_size: "59GB"
  save_graphs: True
  save_graphs_path: "./graph"
  device_id: 0
  jit_config:
    jit_level: "O1"
  memory_optimize_level: "00"
  <strong>ascend_config:</strong>
    <strong>hccl_watchdog: False</strong></pre>

**Adapting the Training Script<a name="zh-cn_topic_0000002134174097_section189248183358"></a>**

1. After the distributed environment initialization is complete and the global rank is obtained, modify the training script to start the TaskD Manager within the training script.
    1. Create a `manager.py` file in the current directory where the training script is invoked. The content of the `manager.py` file is as follows.

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os

        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # Total number of nodes
        proc_per_node=XX     # Number of training processes per node

        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE]
        >For detailed parameter descriptions in the manager.py file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

    2. Add the following code to the training script (for example, `train_start.sh`) to start TaskD Manager. In the following code, the two statements `TASKD_SO_PATH` and `export LD_PRELOAD` are used to configure the path of `libtaskd.so` (from TaskD installation) into the environment variable `LD_PRELOAD`. If these two statements fail to configure successfully, you can manually run the `pip show taskd` command to get the `Location` value, append `/taskd/python/cython_api/libs/libtaskd.so`, and then set it via `export`.

        ```shell
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="on"
        # Under the PyTorch framework
        if [[ "${RANK}" == 0 ]]; then
           export MASTER_ADDR=${POD_IP}
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path, and the error.log log path must be created in advance.
        fi
        # Under the MindSpore framework
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path, and the error.log log path must be created in advance.
        fi
        ```

2. (Optional) In the startup script (for example, `train_start.sh`), add the `--max_restarts` parameter. An example is shown below.

    <pre codetype="shell">
    ...
       logger "server id is: ""${server_id}"
       if [ "${framework}" == "PyTorch" ]; then
         get_env_for_pytorch_multi_node_job
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT <strong>--max_restarts 32767</strong>" </pre>

Where, `--max_restarts` specifies the maximum number of fault triggers allowed within the container, expressed as an integer. If this limit is exceeded, the PyTorch training process will exit immediately. If this parameter is not configured, the default value is `32767`.

- In the MindSpeed scenario, you need to modify the `train_start.sh` script and add the following fields in the code. An example is shown below.

        ```shell
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"   # Enable the re-execution feature of HCCL operators (operator-level online recovery). Re-execution means that when an SDMA or RDMA CQE type error is reported during the execution of a communication operator, HCCL will attempt to re-execute this communication operator.
        export HCCL_ASYNC_ERROR_HANDLING=0
        ```

- In the MindFormers scenario, you need to modify the `msrun_launcher.sh` script and add the following fields in the code. An example is shown below.

        ```shell
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  # This environment variable is used to configure whether to enable the re-execution feature of HCCL operators. Re-execution means that when an SDMA or RDMA CQE type error is reported during the execution of a communication operator, HCCL will attempt to re-execute this communication operator.
        ```

>[!NOTE]
>To test the process-level online recovery feature, configure it by referring to [Process-level Online Recovery Verification](../../references/appendix.md).

## Configuring Operator-level Online Recovery<a name="ZH-CN_TOPIC_0000002511426477"></a>

This part describes key steps for configuring operator-level online recovery. For details on the feature introduction, usage constraints, supported product models, and principles of operator-level online recovery, see [Operator-level Online Recovery](./01_solutions_principles.md#operator-level-online-recovery).

**Configuring Environment Variables<a name="section12610013287a"></a>**

Before using operator-level online recovery, configure the environment variables `HCCL_OP_RETRY_ENABLE` and `HCCL_OP_RETRY_PARAMS` in the training startup script. For detailed descriptions of these environment variables, see the *[CANN Environment Variable Reference](https://www.hiascend.com/document/detail/en/canncommercial/900/maintenref/envvar/envref_07_0001.html)*. A configuration example is shown below.

```shell
export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"     # Whether to enable the re-execution feature of HCCL operators
export HCCL_OP_RETRY_PARAMS="MaxCnt:3, HoldTime:5000, IntervalTime:1000"    # Configures specific parameters for HCCL operator re-execution, including the maximum number of re-execution attempts, the wait time before the first re-execution, and the interval between two re-executions
```

## Configuring Suspension and Switchback for Link Failover Communication<a name="ZH-CN_TOPIC_0000002511346495"></a>

### PyTorch Scenario (Based on MindSpeed-LLM)<a name="ZH-CN_TOPIC_0000002511426445"></a>

This section  describes how to configure suspension and switchback of link failover communication. For details about its features, restrictions, supported products, and working principles, see [Suspension and Switchback for Link Failover Communication](./01_solutions_principles.md).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- Complete the installation of the following components on the corresponding nodes: [Ascend Docker Runtime](../../developer_guide/installation_deployment/manual_installation/02_ascend_docker_runtime.md), [Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md), [ClusterD](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md), [Ascend Device Plugin](../../developer_guide/installation_deployment/manual_installation/04_ascend_device_plugin.md), and [Volcano](../../developer_guide/installation_deployment/manual_installation/05_volcano.md) (The versions of the above MindCluster components must be compatible with TaskD.)
- Install the following components in the container: [torch_npu](./07_using_resumable_training_on_the_cli.md) (7.1.RC1 or later), [CANN](./07_using_resumable_training_on_the_cli.md) (8.2.RC1 or later), [TaskD](./07_using_resumable_training_on_the_cli.md), and [MindIO](./07_using_resumable_training_on_the_cli.md#) (7.1.RC1 or later)

**Procedure<a name="section188080175496"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the script and start TaskD Worker within the training process.

    1. Start TaskD Manager.
        1. Create a `manager.py` file in the current directory where the training script is invoked. The content of the `manager.py` file is as follows.

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
            >For detailed parameter descriptions in the manager.py file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path. The error.log path must be created in advance.
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
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload specified .so files. An example is shown below.
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- `libmspti.so`: This .so file is provided by MindStudio and integrated within the CANN package. Default installation path: `/usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This .so file is provided by TaskD, and the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. The TaskD installation path can be queried using the following command. The `Location` field in the output is the TaskD installation path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

2. Modify the training framework code.
    1. Go to the "[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)" repository, switch to the corresponding version branch according to [mindcluster-deploy Open Source Repository Version Description](../../references/appendix.md#mindcluster-deploy-open-source-repository-version-description), obtain the `train_start.sh` file from the `samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3` directory, and construct the following directory structure on the management node.

        ```text
        root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
        scripts/
        └── train_start.sh
        ```

    2. Configure the training startup script `train_start.sh`, and add the following fields to the code.

        ```shell
        # Enable the re-execution feature for HCCL operators. Re-execution means that when SDMA or RDMA CQE type errors are reported during the execution of a communication operator, HCCL will attempt to re-execute this communication operator.
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"
        ```

3. Modify the job YAML.

    Add the following bold fields to the job YAML to enable process-level online recovery, and modify the container port by adding port 9601 for TaskD communication under all Pods.

    <pre codetype="yaml">
    ...
        labels:
          ...
          <strong>fault-scheduling: "grace"</strong>
       ...
    ...
        annotations:
          ...
          <strong>recover-strategy: "retry"    # Recovery strategy. The value retry indicates that process-level online recovery is enabled.</strong>
       ...
    ...
    spec:
       replicaSpecs:
         Master:
           template:
             spec:
               containers:
               - name: ascend # Do not modify
                 ...
                 args:
                   - |
                     ...
                     bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                       ...
                 <strong>ports: </strong>
                   <strong>- containerPort: 9601</strong>
                     <strong>name: taskd-port</strong>
    ...
         Worker:
           template:
             spec:
               containers:
               - name: ascend # Do not modify
                 ...
                 args:
                   - |
                     ...
                     bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                       ...
                 <strong>ports:</strong>
                   <strong>- containerPort: 9601</strong>
                     <strong>name: taskd-port</strong>
    ...</pre>

### MindSpore Scenario (Based on MindFormers)<a name="ZH-CN_TOPIC_0000002511346443"></a>

This section describes how to configure suspension and switchback of link failover communication. For details about its features, restrictions, supported products, and working principles, see [Suspension and Switchback of Link Failover communication](./01_solutions_principles.md).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- Install the following components on the corresponding nodes: [Ascend Docker Runtime](../../developer_guide/installation_deployment/manual_installation/02_ascend_docker_runtime.md), [Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md), [ClusterD](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md), [Ascend Device Plugin](../../developer_guide/installation_deployment/manual_installation/04_ascend_device_plugin.md), and [Volcano](../../developer_guide/installation_deployment/manual_installation/05_volcano.md) (the versions of the above MindCluster components must be compatible with TaskD)
- Install the following components in the container: [torch_npu](./07_using_resumable_training_on_the_cli.md) (7.1.RC1 or later), [CANN](./07_using_resumable_training_on_the_cli.md) (8.2.RC1 or later), [TaskD](./07_using_resumable_training_on_the_cli.md), and [MindIO](./07_using_resumable_training_on_the_cli.md#) (7.1.RC1 or later)

**Procedure<a name="section9479182019317"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the training script, start TaskD Proxy in the management process, and start TaskD Worker in the training process.
    1. Start TaskD Manager.
        1. Create `a manager.py` file in the current directory when calling the training script. The content of the `manager.py` file is as follows.

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
            >For details about the parameters in the `manager.py` file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            ```shell
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path. The error.log path must be created in advance.
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
                ……</pre>

                logger.info(".........Starting Training Model..........")
                if get_real_rank() % 8 == 0:
                    pprint(config)
                logger.info(".........Model Compiling, Please Wait a Moment...........")
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

2. Modify the training framework code to enable the track borrowing switch.

    Edit `QWEN3_for_MS_code/scripts/msrun_launcher.sh` and add the following fields to the code.

    ```shell
    export MS_ENABLE_TFT="{TTP:1,TSP:1}"           # Enable dying gasp and link failover.
    export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  # This environment variable is used to configure whether to enable the re-execution feature of HCCL operators. Re-execution means that when an SDMA or RDMA CQE type error is reported during the execution of a communication operator, HCCL will attempt to re-execute this communication operator.
    ```

    >[!NOTE]
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload the specified so file. An example is as follows.
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- `libmspti.so`: This .so file is provided by MindStudio and integrated in the CANN package. The default installation path is `/usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This .so file is provided by TaskD. After the whl package is installed, the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. You can run the following command to query the path where TaskD is located. The `Location` field in the command output is the target path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

3. Modify the job YAML.

    Add the following bold fields in the job YAML to enable process-level online recovery, and modify the container port by adding port 9601 for TaskD communication under all Pods.

    <pre codetype="yaml">
    ...
        labels:
          ...
          <strong>fault-scheduling: "grace"</strong>
      ...
    ...
        annotations:
          ...
          <strong>recover-strategy: "retry"    # Recovery strategy. The value retry indicates enabling process-level online recovery.</strong>
      ...
    ...
    spec:
      replicaSpecs:
        Master:
          template:
            spec:
              containers:
              - name: ascend # Do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
                <strong>ports:</strong>
                  <strong>- containerPort: 9601</strong>
                    <strong>name: taskd-port</strong>
    ...
        Worker:
          template:
            spec:
              containers:
              - name: ascend # Do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
                <strong>ports:</strong>
                  <strong>- containerPort: 9601</strong>
                    <strong>name: taskd-port</strong>
    ...</pre>

## Configuring Graceful Fault Tolerance<a name="ZH-CN_TOPIC_0000002511346501"></a>

>[!NOTE]
>This function has been deprecated. It will not be supported in PyTorch versions beyond 7.2.RC1 and MindSpore versions beyond 7.1.RC1.

This section describes how to configure graceful fault tolerance. For details about its features, restrictions, supported products, and working principles, see [(Optional) Graceful Fault Tolerance](./01_solutions_principles.md#optional-graceful-fault-tolerance).

**Building an Image<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

Use a Dockerfile to build a container image and add a startup command.

```shell
# Adaptation script to MindCluster resuamble training. MINDIO_TTP_PKG is the path to the MindIO whl installation package. Fill it in according to the actual situation.
RUN pip3 install $MINDIO_TTP_PKG
```

**Adapting the Training Script<a name="section731511818483"></a>**

In the startup script (for example, `train_start.sh`), add the following fields. An example is shown below.

```shell
...
export MS_ENABLE_TFT="{RSC:1}"      # Configure this field in MindSpore scenarios to enable graceful fault tolerance.
...
```

**Configuring the Startup YAML File<a name="zh-cn_topic_0000002138594553_section18371651403"></a>**

Modify the startup YAML of Ascend Device Plugin, set `-hotReset=1` to enable hot reset, and use graceful fault tolerance mode. **Note: Graceful fault tolerance cannot be enabled simultaneously with process-level rescheduling or process-level online recovery.**

```Yaml
...
      containers:
      - image: ascend-k8sdeviceplugin:v{version}
        name: device-plugin-01
        resources:
          requests:
            memory: 500Mi
            cpu: 500m
          limits:
            memory: 500Mi
            cpu: 500m
        command: [ "/bin/bash", "-c", "--"]
        args: [ "device-plugin
                 -useAscendDocker=true
                 -volcanoType=true                    # Volcano must be used in rescheduling scenarios.
                 -autoStowing=true                    # Whether to enable automatic management. The default value is true. If this parameter is set to false, automatic management is disabled. In this case, after the processor health status changes from unhealthy to healthy, or the network fault on the processor parameter plane is recovered, the processor will not be automatically added to the schedulable resource pool. This parameter applies only to Atlas training products.
                 -listWatchPeriod=5                   # Sets the health status check period, in seconds. Range: [3,1800]
                 -hotReset=1      # Enable the hot reset function and use graceful fault tolerance mode on top of Job-level or Pod-level rescheduling.
                 -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                 -logLevel=0" ]
        securityContext:
          privileged: true
          readOnlyRootFilesystem: true
...
```

## Configuring Online Stress Testing <a name="ZH-CN_TOPIC_0000002511426487"></a>

### PyTorch Scenario (Based on MindSpeed-LLM)<a name="ZH-CN_TOPIC_0000002479386572"></a>

This section guides users through the key steps for configuring online stress testing. For details on the feature introduction, usage constraints, and supported product models of online stress testing, see [Online Stress Testing](./01_solutions_principles.md#online-stress-testing).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- Complete the installation of the following components on the corresponding nodes: [Ascend Docker Runtime](../../developer_guide/installation_deployment/manual_installation/02_ascend_docker_runtime.md), [Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md), [ClusterD](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md), [Ascend Device Plugin](../../developer_guide/installation_deployment/manual_installation/04_ascend_device_plugin.md), and [Volcano](../../developer_guide/installation_deployment/manual_installation/05_volcano.md) (The versions of the above MindCluster components must be compatible with TaskD.)
- Install the following components in the container: [torch_npu](./07_using_resumable_training_on_the_cli.md) (7.1.RC1 or later), [CANN](./07_using_resumable_training_on_the_cli.md) (8.2.RC1 or later), [TaskD](./07_using_resumable_training_on_the_cli.md), and [MindIO](./07_using_resumable_training_on_the_cli.md#) (7.1.RC1 or later)

**Procedure<a name="section188080175496"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager within the training script and start TaskD Worker inside the training process.

    1. Start the TaskD Manager.
        1. Create a `manager.py` file and place it in the current directory where the training script is invoked. The content of the `manager.py` file is as follows.

            ```Python
            from taskd.api import init_taskd_manager, start_taskd_manager
            import os

            job_id=os.getenv("MINDX_TASK_ID")
            node_nums=XX         #Total number of nodes
            proc_per_node=XX     # Number of training processes per node

            init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
            start_taskd_manager()
            ```

            >[!NOTE]
            >For details about the parameters in the manager.py file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            ```shell
            sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${RANK}" == 0 ]]; then
                export MASTER_ADDR=${POD_IP}
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path. The error.log path must be created in advance.
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
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload the specified so files. An example is as follows.
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- `libmspti.so`: This .so file is provided by MindStudio and integrated in the CANN package. The default installation path is `/usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This so is provided by TaskD. After the whl package is installed, the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. You can run the following command to query the path where TaskD is located. The `Location` field in the command output is the target path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

2. Modify the job YAML.

    Add the following bold fields in the job YAML to enable process-level rescheduling and add port 9601 for TaskD communication under all Pods.

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>recover-strategy: "recover"    # Recovery strategy. The value recover indicates enabling process-level rescheduling.</strong>
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
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601</strong>
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
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>

### MindSpore Scenario (Based on MindFormers) <a name="ZH-CN_TOPIC_0000002479226554"></a>

This section describes how to configure online stress testing. For details about its features, restrictions, supported products, and working principles, see [Online Stress Testing](./01_solutions_principles.md#online-stress-testing).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- Install the following components on the corresponding nodes: [Ascend Docker Runtime](../../developer_guide/installation_deployment/manual_installation/02_ascend_docker_runtime.md), [Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md), [ClusterD](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md), [Ascend Device Plugin](../../developer_guide/installation_deployment/manual_installation/04_ascend_device_plugin.md), and [Volcano](../../developer_guide/installation_deployment/manual_installation/05_volcano.md) (the versions of the above MindCluster components must be compatible with TaskD)
- Install MindSpore (version 2.7.0 or later), [CANN](./07_using_resumable_training_on_the_cli.md) (version 8.2.RC1 or later), [TaskD](./07_using_resumable_training_on_the_cli.md), and [MindIO](./07_using_resumable_training_on_the_cli.md) (version 7.2.RC1 or later) in the container.

**Procedure<a name="section9479182019317"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the training script, start TaskD Proxy in the management process, and start TaskD Worker inside the training process.
    1. Start TaskD Manager.
        1. Create a `manager.py` file in the current directory where the training script is invoked. The content of the `manager.py` file is as follows.

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
            >For detailed parameter descriptions in the manager.py file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

        2. Add the following code to the training script to start TaskD Manager.

            ```shell
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path. The error.log path must be created in advance.
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

2. Modify the training framework code and enable online stress testing.

    Edit the startup script `QWEN3_for_MS_code/scripts/msrun_launcher.sh` file and add the following fields to the code.

    ```shell
    export MS_ENABLE_TFT="{TTP:1,TSP:1}"           # Enable dying gasp and online stress testing.
    ```

    >[!NOTE]
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload the specified .so file. An example is as follows.
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- `libmspti.so`: This .so file is provided by MindStudio and integrated in the CANN package. The default installation path is /`usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This .so file is provided by TaskD. After the whl package is installed, the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. You can run the following command to query the path where TaskD is located. The `Location` field in the command output is the target path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

3. Modify the job YAML.

    Add the following bold fields to the job YAML to enable process-level rescheduling, and modify the container port by adding port 9601 for TaskD communication under all Pods.

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>recover-strategy: "recover"    # Recovery policy. The value is recover, indicating that process-level rescheduling is enabled.</strong>
         ...
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # Do not modify
                    ...
                    command:                           # Training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601</strong>
                        <strong>name: taskd-port</strong>
        ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # Do not modify
                    ...
                    command:                           # Training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>

## Configuring Hot Switching<a name="ZH-CN_TOPIC_0000002511426471"></a>

This section describes how to configure hot switching. For details about its features, restrictions, supported products, and working principles, see [Hot Switching](./01_solutions_principles.md#hot-switching).

**Building an Image<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

Use a Dockerfile to build a container image and add a startup command. An example is shown below.

```shell
# Adaptation script to MindCluster resumable training. TASKD_WHL is the path to the TaskD whl installation package, MINDIO_TTP_PKG is the path to the MindIO whl installation package, and MINDSPORE_WHL is the path to the MindSpore whl installation package. Please fill in the paths according to your actual situation.
# Optional. Under the PyTorch framework, you must configure the following command when using hot switching.
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# Optional. Under the MindSpore framework, you must configure the following command when using hot switching.
RUN pip3 install $MINDIO_TTP_PKG
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDSPORE_WHL
```

**Prepararing the Job YAML<a name="zh-cn_topic_0000002134174097_section98717593512"></a>**

In the job YAML, add the following fields to enable hot switching, modify the container port, and add port 9601 for TaskD communication under all Pods.

```Yaml
...
metadata:
   labels:
     ...
     subHealthyStrategy: "hotSwitch"
...
        spec:
...
           containers:
...
             ports:
               - containerPort: 9601
                 name: taskd-port
...
```

**Adapting the Training Script<a name="zh-cn_topic_0000002098654822_section17330181621710"></a>**

After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager within the training script.

1. Create a `manager.py` file and place it in the current directory when invoking the training script. The content of the `manager.py` file is as follows.

    ```Python
    from taskd.api import init_taskd_manager, start_taskd_manager
    import os

    job_id=os.getenv("MINDX_TASK_ID")
    node_nums=XX          # Total number of nodes
    proc_per_node=XX     # Training processes per node

    init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
    start_taskd_manager()
    ```

    >[!NOTE]
    >For detailed parameter descriptions in the manager.py file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

2. Add the following code to the training script to start TaskD Manager.

    ```shell
    export TASKD_PROCESS_ENABLE="on"

    # Under the PyTorch framework
    if [[ "${RANK}" == 0 ]]; then
        export MASTER_ADDR=${POD_IP}
        python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path. The error.log log path must be created in advance.
    fi

    torchrun ...

    # Under the MindSpore framework
    if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
        python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # The specific execution path of manager.py is determined by the current path. The error.log log path must be created in advance.
    fi

    msrun ...
    ```

    >[!NOTE]
    >If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the `LD_PRELOAD` environment variable in the training script. This environment variable allows the system to preload specified .so files. An example is as follows:
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- `libmspti.so`: This .so file is provided by MindStudio and integrated in the CANN package. The default installation path is `/usr/local/Ascend/cann/lib64/libmspti.so`.
    >- `libtaskd.so`: This .so file is provided by TaskD. After the whl package is installed, the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. You can run the following command to query the path where TaskD is located. The `Location` field in the command output is the target path.
    >
    >     ```shell
    >     pip show taskd
    >     ```

## Configuring Elastic Training<a name="ZH-CN_TOPIC_0000002511346471"></a>

This section describes how to configure elastic training. For details about its features, restrictions, supported products, and working principles, see [Elastic Training](./01_solutions_principles.md#elastic-training).

**Prerequisites<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- Complete the installation of the following components on the corresponding nodes: [Ascend Docker Runtime](../../developer_guide/installation_deployment/manual_installation/02_ascend_docker_runtime.md), [Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md), [ClusterD](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md), [Ascend Device Plugin](../../developer_guide/installation_deployment/manual_installation/04_ascend_device_plugin.md), and [Volcano](../../developer_guide/installation_deployment/manual_installation/05_volcano.md) (The versions of the above MindCluster components must be compatible with TaskD.)
- Install the following components in the container: [torch_npu](./07_using_resumable_training_on_the_cli.md) (7.1.RC1 or later), [CANN](./07_using_resumable_training_on_the_cli.md) (8.2.RC1 or later), [TaskD](./07_using_resumable_training_on_the_cli.md), and [MindIO](./07_using_resumable_training_on_the_cli.md#) (7.1.RC1 or later)

**Procedure<a name="section188080175496"></a>**

1. After the distributed environment is initialized and the global rank is obtained, modify the training script to start TaskD Manager in the training script.
    1. Create a `manager.py` file in the current directory where the training script is invoked. The content of the `manager.py` file is as follows.

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
        >For detailed parameter descriptions in the manager.py file, see [def init_taskd_manager(config:dict) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

    2. Add the following code to the training script to start TaskD Manager.

        ```shell
        sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        export TASKD_PROCESS_ENABLE="on"
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path. The error.log path must be created in advance.
        fi

        torchrun ...
        ```

2. Modify the job YAML.

    Add the following bold fields to the job YAML to enable elastic training, and modify the container port by adding port 9601 for TaskD communication under all Pods.

      <pre codetype="yaml">
        ...
           labels:
             ...
             <strong>fault-scheduling: "grace"</strong>
         ...
        ...
           annotations:
             ...
             <strong>wait-reschedule-timeout: "270" # Timeout for waiting for the faulty node to be rescheduled during process-level recovery. The default is 270 seconds, with a valid range of 30 to 270. When both process-level recovery and elastic training are enabled, if the faulty node is successfully scheduled within this time, process-level recovery is performed; otherwise, elastic training is triggered.</strong>
             <strong>recover-strategy: "elastic-training"    # Available recovery strategy. The value is elastic-training, indicating that elastic training is enabled.</strong>
         ...
        ...
        spec:
          replicaSpecs:
            Master:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    env:
                      <strong>- name: MINDIO_WAIT_MINDX_TIME         # It is recommended to configure this to 60 or above when process-level recovery is not enabled and elastic training is enabled.</strong>
                        <strong>value: "60"</strong>
                    args:
                      - |
                        ...
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601</strong>
                        <strong>name: taskd-port</strong>
        ...
            Worker:
              template:
                spec:
                  containers:
                  - name: ascend # do not modify
                    env:
                      <strong>- name: MINDIO_WAIT_MINDX_TIME         # Recommended to set to 60 or above when process-level recovery is disabled and elastic training is enabled</strong>
                        <strong>value: "60"</strong>
                    args:
                      - |
                        ...
                        bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                          ...
                    <strong>ports:</strong>
                      <strong>- containerPort: 9601</strong>
                        <strong>name: taskd-port</strong>
        ...</pre>

3. Modify the training framework code.

    Go to the "[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)" repository, switch to the corresponding version branch according to the [mindcluster-deploy Open Source Repository Version Description](../../references/appendix.md#mindcluster-deploy-open-source-repository-version-description), obtain the `train_start.sh` file from the `samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3` directory, and construct the following directory structure on the management node.

    ```text
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```

## Parameter Description<a name="ZH-CN_TOPIC_0000002511346491"></a>

Different fault handling modes require different parameters to be configured, as shown in [Table 1](#table1247342123814). For details about the meaning and filling instructions of each parameter, see [Table 2](#zh-cn_topic_0000002163392281_table1474820818115).In scenarios such as process-level rescheduling, process-level online recovery, process-level in-place recovery, and elastic training, Ascend Operator injects different environment variables based on the user-configured `recover-strategy` and `pod-rescheduling`, and automatically labels the job with `process-recover-enable=on` to enable process-level recovery, without requiring manual specification by the user. The specific injected environment variables are shown in [Table 3](#table10283161512105).

**Table 1**  Parameters required for fault handling

<a name="table1247342123814"></a>
<table><tbody><tr id="row624717420389"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1724711429388"><a name="p1724711429388"></a><a name="p1724711429388"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p882622715436"><a name="p882622715436"></a><a name="p882622715436"></a>Job-Level Rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p117701032194319"><a name="p117701032194319"></a><a name="p117701032194319"></a>Pod-Level Rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p182471842183813"><a name="p182471842183813"></a><a name="p182471842183813"></a>Process-Level  Rescheduling (recover)</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p35155320438"><a name="p35155320438"></a><a name="p35155320438"></a>Process-level In-place Recovery</p>
<p id="p1259532434"><a name="p1259532434"></a><a name="p1259532434"></a>(recover-in-place)</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p37521612448"><a name="p37521612448"></a><a name="p37521612448"></a>Process-level Online Recovery</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p14247442143812"><a name="p14247442143812"></a><a name="p14247442143812"></a>Graceful Fault Tolerance</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p84114121442"><a name="p84114121442"></a><a name="p84114121442"></a>Elastic Training</p>
</td>
</tr>
<tr id="row7247154215383"><td class="cellrowborder" valign="top" width="15.24%"><p id="p22316366391"><a name="p22316366391"></a><a name="p22316366391"></a>hotReset</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p32233618390"><a name="p32233618390"></a><a name="p32233618390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1422133683919"><a name="p1422133683919"></a><a name="p1422133683919"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p12221636143918"><a name="p12221636143918"></a><a name="p12221636143918"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p52116364390"><a name="p52116364390"></a><a name="p52116364390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p221173643911"><a name="p221173643911"></a><a name="p221173643911"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p0211836143914"><a name="p0211836143914"></a><a name="p0211836143914"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1321203633918"><a name="p1321203633918"></a><a name="p1321203633918"></a>-</p>
</td>
</tr>
<tr id="row1024894243810"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1218113612390"><a name="p1218113612390"></a><a name="p1218113612390"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p518736123916"><a name="p518736123916"></a><a name="p518736123916"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1917936133912"><a name="p1917936133912"></a><a name="p1917936133912"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p21753673912"><a name="p21753673912"></a><a name="p21753673912"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1171736143916"><a name="p1171736143916"></a><a name="p1171736143916"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p9171236113916"><a name="p9171236113916"></a><a name="p9171236113916"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p12171236163917"><a name="p12171236163917"></a><a name="p12171236163917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p10161836143914"><a name="p10161836143914"></a><a name="p10161836143914"></a>√</p>
</td>
</tr>
<tr id="row1824884293812"><td class="cellrowborder" valign="top" width="15.24%"><p id="p91533663919"><a name="p91533663919"></a><a name="p91533663919"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p5147367393"><a name="p5147367393"></a><a name="p5147367393"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p151433693911"><a name="p151433693911"></a><a name="p151433693911"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1714183618396"><a name="p1714183618396"></a><a name="p1714183618396"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p111413620399"><a name="p111413620399"></a><a name="p111413620399"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p713136113913"><a name="p713136113913"></a><a name="p713136113913"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p6134364397"><a name="p6134364397"></a><a name="p6134364397"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1813203617390"><a name="p1813203617390"></a><a name="p1813203617390"></a>-</p>
</td>
</tr>
<tr id="row2248144273815"><td class="cellrowborder" valign="top" width="15.24%"><p id="p15112368396"><a name="p15112368396"></a><a name="p15112368396"></a>process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p511163613915"><a name="p511163613915"></a><a name="p511163613915"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p2011203616395"><a name="p2011203616395"></a><a name="p2011203616395"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1610133603920"><a name="p1610133603920"></a><a name="p1610133603920"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p310103643919"><a name="p310103643919"></a><a name="p310103643919"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p91015365399"><a name="p91015365399"></a><a name="p91015365399"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p6105364395"><a name="p6105364395"></a><a name="p6105364395"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p191836133916"><a name="p191836133916"></a><a name="p191836133916"></a>√</p>
</td>
</tr>
<tr id="row2248154243818"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1814364394"><a name="p1814364394"></a><a name="p1814364394"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p12773693917"><a name="p12773693917"></a><a name="p12773693917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p147836193919"><a name="p147836193919"></a><a name="p147836193919"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p13712364397"><a name="p13712364397"></a><a name="p13712364397"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p127193633919"><a name="p127193633919"></a><a name="p127193633919"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p86173683915"><a name="p86173683915"></a><a name="p86173683915"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p9673673919"><a name="p9673673919"></a><a name="p9673673919"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p66123611391"><a name="p66123611391"></a><a name="p66123611391"></a>√</p>
</td>
</tr>
<tr id="row424864214389"><td class="cellrowborder" valign="top" width="15.24%"><p id="p134936133911"><a name="p134936133911"></a><a name="p134936133911"></a>PROCESS_RECOVER</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p184173618395"><a name="p184173618395"></a><a name="p184173618395"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p164123616391"><a name="p164123616391"></a><a name="p164123616391"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p63193643915"><a name="p63193643915"></a><a name="p63193643915"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p83113663920"><a name="p83113663920"></a><a name="p83113663920"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p4323643914"><a name="p4323643914"></a><a name="p4323643914"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p736362397"><a name="p736362397"></a><a name="p736362397"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p42163683917"><a name="p42163683917"></a><a name="p42163683917"></a>√</p>
</td>
</tr>
<tr id="row1924904210386"><td class="cellrowborder" valign="top" width="15.24%"><p id="p212036183913"><a name="p212036183913"></a><a name="p212036183913"></a>ENABLE_RESTART_FAULT_PROCESS</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p170133620393"><a name="p170133620393"></a><a name="p170133620393"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p140173613390"><a name="p140173613390"></a><a name="p140173613390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p799983573910"><a name="p799983573910"></a><a name="p799983573910"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1899973516399"><a name="p1899973516399"></a><a name="p1899973516399"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p1199873573913"><a name="p1199873573913"></a><a name="p1199873573913"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p49985358399"><a name="p49985358399"></a><a name="p49985358399"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1799893511398"><a name="p1799893511398"></a><a name="p1799893511398"></a>-</p>
</td>
</tr>
<tr id="row19391344114010"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1794014446408"><a name="p1794014446408"></a><a name="p1794014446408"></a>ELASTIC_PROCESS_RECOVER_ENABLE</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p1494064419401"><a name="p1494064419401"></a><a name="p1494064419401"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p19940944134019"><a name="p19940944134019"></a><a name="p19940944134019"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p2094054414013"><a name="p2094054414013"></a><a name="p2094054414013"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p494084464017"><a name="p494084464017"></a><a name="p494084464017"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p69402446400"><a name="p69402446400"></a><a name="p69402446400"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p894014464011"><a name="p894014464011"></a><a name="p894014464011"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p199403441401"><a name="p199403441401"></a><a name="p199403441401"></a>-</p>
</td>
</tr>
<tr id="row448045664010"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1848011565403"><a name="p1848011565403"></a><a name="p1848011565403"></a>--enable-high-availability (MindSpeed-LLM side parameter)</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p948075615409"><a name="p948075615409"></a><a name="p948075615409"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1748045664018"><a name="p1748045664018"></a><a name="p1748045664018"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p548025618403"><a name="p548025618403"></a><a name="p548025618403"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p448016564402"><a name="p448016564402"></a><a name="p448016564402"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p9480145674013"><a name="p9480145674013"></a><a name="p9480145674013"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p1848055694012"><a name="p1848055694012"></a><a name="p1848055694012"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p124809565404"><a name="p124809565404"></a><a name="p124809565404"></a>√</p>
</td>
</tr>
<tr id="row112463954119"><td class="cellrowborder" valign="top" width="15.24%"><p id="p76389163416"><a name="p76389163416"></a><a name="p76389163416"></a>--enable-hbmfault-repair (MindSpeed-LLM side parameter)</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p1124699204118"><a name="p1124699204118"></a><a name="p1124699204118"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p192469924110"><a name="p192469924110"></a><a name="p192469924110"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1824619916411"><a name="p1824619916411"></a><a name="p1824619916411"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p724719904114"><a name="p724719904114"></a><a name="p724719904114"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p1424720924114"><a name="p1424720924114"></a><a name="p1424720924114"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p152478914111"><a name="p152478914111"></a><a name="p152478914111"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p32471934114"><a name="p32471934114"></a><a name="p32471934114"></a>-</p>
</td>
</tr>
<tr id="row3150821154117"><td class="cellrowborder" valign="top" width="15.24%"><p id="p7151102124114"><a name="p7151102124114"></a><a name="p7151102124114"></a>--enable-worker-reboot (MindSpeed-LLM side parameter)</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p815192144117"><a name="p815192144117"></a><a name="p815192144117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p315110217417"><a name="p315110217417"></a><a name="p315110217417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p31511721114110"><a name="p31511721114110"></a><a name="p31511721114110"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p115118211413"><a name="p115118211413"></a><a name="p115118211413"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p10151142164119"><a name="p10151142164119"></a><a name="p10151142164119"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p161517211417"><a name="p161517211417"></a><a name="p161517211417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p10151221164119"><a name="p10151221164119"></a><a name="p10151221164119"></a>-</p>
</td>
</tr>
<tr id="row4799183364111"><td class="cellrowborder" valign="top" width="15.24%"><p id="p7799233154115"><a name="p7799233154115"></a><a name="p7799233154115"></a>--enable-elastic-training (MindSpeed-LLM side parameter)</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p2799133324116"><a name="p2799133324116"></a><a name="p2799133324116"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1079915338414"><a name="p1079915338414"></a><a name="p1079915338414"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p127999338410"><a name="p127999338410"></a><a name="p127999338410"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1280083317415"><a name="p1280083317415"></a><a name="p1280083317415"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p11800143320417"><a name="p11800143320417"></a><a name="p11800143320417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p9800153364112"><a name="p9800153364112"></a><a name="p9800153364112"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1780063324111"><a name="p1780063324111"></a><a name="p1780063324111"></a>√</p>
</td>
</tr>
<tr id="row1551285114419"><td class="cellrowborder" valign="top" width="15.24%"><p id="p12512175154117"><a name="p12512175154117"></a><a name="p12512175154117"></a>max_restarts</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p7512125115416"><a name="p7512125115416"></a><a name="p7512125115416"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p13512145114412"><a name="p13512145114412"></a><a name="p13512145114412"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p151211510417"><a name="p151211510417"></a><a name="p151211510417"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p1751215515414"><a name="p1751215515414"></a><a name="p1751215515414"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p10512125104113"><a name="p10512125104113"></a><a name="p10512125104113"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p751285111417"><a name="p751285111417"></a><a name="p751285111417"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p051245113418"><a name="p051245113418"></a><a name="p051245113418"></a>-</p>
</td>
</tr>
<tr id="row1810414334211"><td class="cellrowborder" valign="top" width="15.24%"><p id="p171048313421"><a name="p171048313421"></a><a name="p171048313421"></a>monitor_interval</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p1910414312422"><a name="p1910414312422"></a><a name="p1910414312422"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1710483114210"><a name="p1710483114210"></a><a name="p1710483114210"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p1810415310420"><a name="p1810415310420"></a><a name="p1810415310420"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p201044364218"><a name="p201044364218"></a><a name="p201044364218"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p9104335421"><a name="p9104335421"></a><a name="p9104335421"></a>√</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p1104163194212"><a name="p1104163194212"></a><a name="p1104163194212"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p4104183114212"><a name="p4104183114212"></a><a name="p4104183114212"></a>-</p>
</td>
</tr>
<tr id="row1260817211339"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1960910211831"><a name="p1960910211831"></a><a name="p1960910211831"></a><span id="ph48451032338"><a name="ph48451032338"></a><a name="ph48451032338"></a>fault-retry-times</span></p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p96093216315"><a name="p96093216315"></a><a name="p96093216315"></a><span id="ph116361991246"><a name="ph116361991246"></a><a name="ph116361991246"></a>√</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p1660911211319"><a name="p1660911211319"></a><a name="p1660911211319"></a><span id="ph151830137416"><a name="ph151830137416"></a><a name="ph151830137416"></a>√</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p18609192118312"><a name="p18609192118312"></a><a name="p18609192118312"></a><span id="ph5658101813418"><a name="ph5658101813418"></a><a name="ph5658101813418"></a>√</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p160912214316"><a name="p160912214316"></a><a name="p160912214316"></a><span id="ph115481430045"><a name="ph115481430045"></a><a name="ph115481430045"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p960913211138"><a name="p960913211138"></a><a name="p960913211138"></a><span id="ph8478143111413"><a name="ph8478143111413"></a><a name="ph8478143111413"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p12609102110312"><a name="p12609102110312"></a><a name="p12609102110312"></a><span id="ph1129711326411"><a name="ph1129711326411"></a><a name="ph1129711326411"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p1609421731"><a name="p1609421731"></a><a name="p1609421731"></a><span id="ph1301142311415"><a name="ph1301142311415"></a><a name="ph1301142311415"></a>√</span></p>
</td>
</tr>
</tbody>
</table>

**Table 2** Parameter description

<a name="zh-cn_topic_0000002163392281_table1474820818115"></a>

|Parameter Name|Parameter Location| Parameter Description                                                                                                                                                                                                                                                                                 |
|--|--|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|hotReset|Startup YAML of Ascend Device Plugin | Graceful fault tolerance switch.<ul><li>Value 1: When using resumable training, you can enable the hot reset feature on top of Job-level or Pod-level rescheduling to use graceful fault tolerance mode;</li><li>Value 2: When using process-level recovery, set the hotReset parameter value to 2 to enable offline recovery mode.</li></ul><div class="note"><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><p>The feature corresponding to value 1 has been sunset. Please configure other values.</p></div></div>                            |
|pod-rescheduling|metadata.labels of the training job YAML| <ul><li>on: Enable Pod-level rescheduling.</li><li>Other values or not using this field: Disable Pod-level rescheduling.</li></ul>                                                                                                                                                                                                                      |
|fault-scheduling|metadata.labels of the training job YAML| Rescheduling switch.                                                                                                                                                                                                                                                                               |
|process-recover-enable|metadata.labels of the training job YAML| <ul><li>on: Enable process-level rescheduling and process-level online recovery. Process-level rescheduling and graceful fault tolerance cannot be enabled simultaneously. If both are enabled, checkpoint restart will resume training through Job-level rescheduling.</li><li>pause: Temporarily disable process-level rescheduling and process-level online recovery.</li><li>off or not using this field: Disable process-level rescheduling and process-level online recovery.</li></ul>                                                                                                                         |
|recover-strategy|metadata.annotations of the training job YAML| Available recovery strategies for the job.<ul><li>retry: Process-level online recovery.</li><li>recover: Process-level rescheduling.</li><li>recover-in-place: Process-level in-place recovery.</li><li>elastic-training: Elastic training.</li><li>dump: Save last words.</li><li>exit: Exit training.</li></ul>                                                                                                          |
|PROCESS_RECOVER|spec.replicaSpecs.{ Master \|Scheduler\| Worker}.template.spec.containers.env of the training job YAML| Master switch on the Elastic Agent/TaskD side for process-level rescheduling and process-level online recovery.<ul><li>on: Enable.</li><li>off: Disable.</li></ul>                                                                                                                                                                                                      |
|ELASTIC_PROCESS_RECOVER_ENABLE|spec.replicaSpecs.{ Master\|Scheduler\| Worker}. template.spec.containers.args of the startup training YAML| Switch on the Elastic Agent side for process-level rescheduling, process-level online recovery, and last CKPT recovery features.<ul><li>Value 1: Enable this feature.</li><li>Other values: Disable this feature.<p>When disabling this feature, the related features on the MindIO side must be disabled simultaneously.</p></li></ul><div class="note"><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><p>The Elastic Agent component has been sunset, and related materials will be removed in the version released on December 30, 2026. This environment variable will be removed accordingly.</p></div></div> |
|ENABLE_RESTART_FAULT_PROCESS|spec.replicaSpecs.{ Master\|Scheduler\| Worker}. template.spec.containers.args of the startup training YAML| Switch for the Elastic Agent/TaskD component to enable the in-place recovery feature for faulty processes.<ul><li>on: Enable this feature;</li><li>Other values: Disable this feature</li></ul>                                                                                                                                                                                                   |
|--enable-high-availability|Startup parameter of the training script pretrain_gpt.py| Fault fast recovery feature switch, disabled by default. When configured, the last words feature is enabled.                                                                                                                                                                                                                                                        |
|--enable-hbmfault-repair|Startup parameter of the training script pretrain_gpt.py| Process-level online recovery feature switch, disabled by default. When configured, fault detection is performed on on-chip memory and online repair is completed. Must be enabled together with enable-high-availability.                                                                                                                                                                                                               |
|--enable-worker-reboot|Startup parameter of the training script pretrain_gpt.py| Process-level rescheduling feature switch, disabled by default. When configured, process-level scheduling is performed when a general fault occurs. Must be enabled together with enable-high-availability.                                                                                                                                                                                                                |
|--enable-elastic-training|Startup parameter of the training script pretrain_gpt.py| Elastic training feature switch, disabled by default.                                                                                                                                                                                                                                                                       |
|max_restarts|In the shell script for starting training (e.g., train_start.sh)| Configures the maximum number of fault triggers allowed within the container, with an integer value. If this number is exceeded, the PyTorch training process will exit training directly. The default value is 32767 if this parameter is not configured.                                                                                                                                                                                                                     |
|monitor_interval|In the shell script for starting training (e.g., train_start.sh)| Configures the time interval for monitoring the training process status, in seconds, with an integer value. The default value is 5 seconds if this parameter is not configured.                                                                                                                                                                                                                                             |
|HIGH_AVAILABILITY|In the environment variables injected into the container by Ascend Operator| Ascend Operator automatically injects this environment variable based on the job type. When using MindSpeed-LLM version 2.3.0, this environment variable is automatically read, eliminating the need to manually add the --enable-high-availability, --enable-hbmfault-repair, --enable-worker-reboot, and --enable-elastic-training parameters in train_start.sh to enable the corresponding features.                                                                                  |

**Table 3** Environment variables injected by Ascend Operator

<a name="table10283161512105"></a>
<table><tbody><tr id="row928321541018"><td class="cellrowborder" valign="top" width="7.580000000000002%"><p id="p7676133111213"><a name="p7676133111213"></a><a name="p7676133111213"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="15.500000000000004%"><p id="p98460430123"><a name="p98460430123"></a><a name="p98460430123"></a>recover</p>
</td>
<td class="cellrowborder" valign="top" width="14.830000000000002%"><p id="p58461043151210"><a name="p58461043151210"></a><a name="p58461043151210"></a>retry</p>
</td>
<td class="cellrowborder" valign="top" width="19.39%"><p id="p20846164319122"><a name="p20846164319122"></a><a name="p20846164319122"></a>recover-in-place</p>
</td>
<td class="cellrowborder" valign="top" width="9.920000000000002%"><p id="p14247569165"><a name="p14247569165"></a><a name="p14247569165"></a>elastic-training</p>
</td>
<td class="cellrowborder" valign="top" width="16.200000000000003%"><p id="p158461443201210"><a name="p158461443201210"></a><a name="p158461443201210"></a>dump</p>
</td>
<td class="cellrowborder" valign="top" width="7.000000000000002%"><p id="p148465435121"><a name="p148465435121"></a><a name="p148465435121"></a>exit</p>
</td>
<td class="cellrowborder" valign="top" width="9.580000000000002%"><p id="p584614351218"><a name="p584614351218"></a><a name="p584614351218"></a>pod-rescheduling</p>
</td>
</tr>
<tr id="row10283115171018"><td class="cellrowborder" valign="top" width="7.580000000000002%"><p id="p1621320131319"><a name="p1621320131319"></a><a name="p1621320131319"></a><span id="ph1551815244211"><a name="ph1551815244211"></a><a name="ph1551815244211"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.500000000000004%"><a name="ul22620111313"></a><a name="ul22620111313"></a><ul id="ul22620111313"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>HIGH_AVAILABILITY=recover</li></ul>
</td>
<td class="cellrowborder" valign="top" width="14.830000000000002%"><a name="ul62102015138"></a><a name="ul62102015138"></a><ul id="ul62102015138"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>HIGH_AVAILABILITY=retry<p id="p102420141318"><a name="p102420141318"></a><a name="p102420141318"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="19.39%"><a name="ul721720131318"></a><a name="ul721720131318"></a><ul id="ul721720131318"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>ENABLE_RESTART_FAULT_PROCESS=on</li><li>HIGH_AVAILABILITY=recover</li></ul>
</td>
<td class="cellrowborder" valign="top" width="9.920000000000002%"><a name="ul4462124162110"></a><a name="ul4462124162110"></a><ul id="ul4462124162110"><li>PROCESS_RECOVER=on</li><li>HIGH_AVAILABILITY=elastic-training</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.200000000000003%"><a name="ul182142017135"></a><a name="ul182142017135"></a><ul id="ul182142017135"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>HIGH_AVAILABILITY=dump<p id="p5216201131"><a name="p5216201131"></a><a name="p5216201131"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="7.000000000000002%"><p id="p102102012138"><a name="p102102012138"></a><a name="p102102012138"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.580000000000002%"><p id="p163162071320"><a name="p163162071320"></a><a name="p163162071320"></a>-</p>
</td>
</tr>
<tr id="row1628391516107"><td class="cellrowborder" valign="top" width="7.580000000000002%"><p id="p139556322136"><a name="p139556322136"></a><a name="p139556322136"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="15.500000000000004%"><a name="ul295563211139"></a><a name="ul295563211139"></a><ul id="ul295563211139"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT={ ARF:1}<p id="p16955153219134"><a name="p16955153219134"></a><a name="p16955153219134"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="14.830000000000002%"><a name="ul109551632111320"></a><a name="ul109551632111320"></a><ul id="ul109551632111320"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT={ UCE:1, HCCE:1}<p id="p1595593241313"><a name="p1595593241313"></a><a name="p1595593241313"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="19.39%"><a name="ul195583215139"></a><a name="ul195583215139"></a><ul id="ul195583215139"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>ENABLE_RESTART_FAULT_PROCESS=on</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT={ ARF:1}<p id="p1095514325133"><a name="p1095514325133"></a><a name="p1095514325133"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="9.920000000000002%"><p id="p324776161612"><a name="p324776161612"></a><a name="p324776161612"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="16.200000000000003%"><a name="ul1595513218133"></a><a name="ul1595513218133"></a><ul id="ul1595513218133"><li>PROCESS_RECOVER=on</li><li>ELASTIC_PROCESS_RECOVER_ENABLE=1</li><li>MINDIO_FOR_MINDSPORE=1</li><li>MS_ENABLE_TFT={ TTP:1}<p id="p495553210131"><a name="p495553210131"></a><a name="p495553210131"></a></p>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="7.000000000000002%"><p id="p7955193220131"><a name="p7955193220131"></a><a name="p7955193220131"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="9.580000000000002%"><p id="p275141842218"><a name="p275141842218"></a><a name="p275141842218"></a>MS_ENABLE_TFT={ RSC:1}</p>
</td>
</tr>
</tbody>
</table>
