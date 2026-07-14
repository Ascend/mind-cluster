# Configuring Training Recovery<a name="ZH-CN_TOPIC_0000002479386506"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T08:01:30.246Z pushedAt=2026-06-09T09:02:55.489Z -->

## Configuring Periodic Checkpoint Saving<a name="ZH-CN_TOPIC_0000002479226552"></a>

This section describes key steps for periodic checkpoint saving. For details on the features of periodic checkpoint saving, see [Periodic Checkpoint Saving](./01_solutions_principles.md).

**Configuring Storage Checkpoint Loading<a name="zh-c_topic_0000002111866386_section1296017551704"></a>**

Loading checkpnoints from storage can be performed using the loading interface provided by the AI framework. You need to pass the file path to be loaded into the AI framework. Taking the MindSpeed-LLM framework as an example, you can refer to the following example if you need to configure the storage checkpoint loading function.

In the job YAML, add the `--load /data/ckpt/XXX \` parameter to enable storage checkpoint loading. `--load` is the unified switch for training process recovery; training process recovery takes effect only after this switch is turned on.

```Yaml
...
spec:
  replicaSpecs:
    Master:
      template:
        spec:
          containers:
          - name: ascend # Do not modify
            args:
              - |
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                  --load /data/ckpt/XXX \  # Checkpoint storage path
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
                --load /data/ckpt/XXX \    # Checkpoint storage path
                  ...
...
```

## Configuring Dying Gasp Checkpoint Saving<a name="ZH-CN_TOPIC_0000002479226544"></a>

This section provides key steps for dying gasp checkpoint saving. For details, see [Dying Gasp Checkpoint Saving](./01_solutions_principles.md).

**Building an Image<a name="zh-cn_topic_0000002112026142_section26738428458"></a>**

Use a Dockerfile to build a container image and add the startup command.

```shell
...
# Adaptation Script to MindCluster lossless resumable training
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG

# Optional. The following commands must be configured when using graceful fault tolerance, Pod-level rescheduling, or process-level rescheduling.
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
```

**Preparing the Job YAML<a name="zh-cn_topic_0000002112026142_section2671124124612"></a>**

In the training job YAML, add the following fields to enable process-level recovery. `recover-strategy` is the strategy used for training process recovery, where `dump` indicates dying gasp checkpoint saving. Under `ports`, add `ttp-port 8000` and `port 9601` for TaskD communication.

Saving dying gasp checkpoint can be used as a policy named `dump` of `recover-strategy` for process-level recovery. An example is shown below.

<pre codetype="yaml">
...
metadata:
   labels:
     ...
 ...
...
   annotations:
     ...
     <strong>recover-strategy: "dump"       # Ding gasp checkpoint saving</strong>
 ...

...
spec:
   replicaSpecs:
      Master:
         template:
            spec:
              containers:
                 env:
                   <strong>- name: TTP_PORT</strong>
                     <strong>value: "8000"</strong>
                 args: […]
                 ports:
                   <strong>- containerPort: 8000</strong>
                     <strong>name: ttp-port</strong>
                   <strong>- containerPort: 9601</strong>
                     <strong>name: taskd-port</strong>
     ...
     Worker:
        template:
          spec:
            containers:
               env:
                 <strong>- name: TTP_PORT</strong>
                   <strong>value: "8000"</strong>
               args: […]
               ports:
                 <strong>- containerPort: 8000</strong>
                   <strong>name: ttp-port</strong>
                 <strong>- containerPort: 9601</strong>
                   <strong>name: taskd-port</strong>
  ...</pre>

**Adapting the Training Script<a name="zh-cn_topic_0000002112026142_section058501610462"></a>**

1. After the distributed environment initialization is complete and the global rank is obtained, modify the training script to launch TaskD Manager within the training script.
    1. Create a `manager.py` and save it to the directory where the training script is called. The content of the `manager.py` file is as follows.

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
        >For detailed parameter descriptions in the `manager.py` file, see [def init_taskd_manager(config:dict) -> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

    2. Add the following code to the training script to start TaskD Manager.

        ```shell
        export TASKD_PROCESS_ENABLE="on"

        # Under the PyTorch framework
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path, and the error.log log path must be created in advance.
        fi

        torchrun ...
        ```

2. In the startup script (e.g., `train_start.sh`), add the `--max_restarts` parameter. An example is shown below.

    <pre codetype="shell">
    ...
       logger "server id is: ""${server_id}"
       if [ "${framework}" == "PyTorch" ]; then
         get_env_for_pytorch_multi_node_job
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT  <strong>--max_restarts 32767</strong>"
     ...</pre>

      Here, --max_restarts specifies the maximum number of fault triggers allowed within the container, expressed as an integer. If this limit is exceeded, the PyTorch training process will exit immediately. If this parameter is not configured, the default value is `32767`.

>[!NOTE]
>If the error "the libtaskd.so has not been loaded" occurs during training, you need to import the L`D_PRELOAD` environment variable in the training script. This environment variable allows the system to preload specified .so files. An example is shown below.
>
>```shell
>export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
>```
>
>- `libmspti.so`: This .so file is provided by MindStudio and integrated in the CANN package. The default installation path is /`usr/local/Ascend/cann/lib64/libmspti.so`.
>- `libtaskd.so`: This .so file is provided by TaskD. After the whl package is installed, the path is `TaskD installation path/taskd/python/cython_api/libs/libtaskd.so`. You can run the following command to query the path where TaskD is located. The `Location` field in the command output is the target path.
>
>     ```shell
>     pip show taskd
>     ```

## Restoring Parameter Passing on the Parameter Plane<a name="ZH-CN_TOPIC_0000002479386502"></a>

Currently, this capability is only supported in the process-level rescheduling and process-level online recovery features. It is enabled by default after adapting according to the [Configuring Process-Level Rescheduling](./04_configuring_fault_handling_policies.md#configuring-process-level-rescheduling) and [Configuring Process-Level Online Recovery](./04_configuring_fault_handling_policies.md#configuring-process-level-online-recovery) features.

**(Optional) Disabling Parameter Passing Recovery on the Parameter Plane<a name="zh-cn_topic_0000002181310402_section199132050405"></a>**

For the process-level rescheduling and process-level online recovery features, if you want to disable this function and load parameters from the storage checkpoint, you need to modify the job YAML file. The following is an example of using process-level rescheduling and disabling parameter passing recovery on the parameter plane.

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
     <strong>recover-strategy: "recover"   # Recovery strategy: process-level rescheduling</strong>
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
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                  <strong>--distributed-optimizer-no-replica \</strong>
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
                  <strong>--distributed-optimizer-no-replica \</strong>
                  ...
...</pre>

**distributed-optimizer-no-replica** indicates whether to support periodic checkpoints for data repair, which is disabled by default. After this function is enabled, the replica optimizer does not have replicas, reducing memory usage. In process-level rescheduling and process-level online recovery scenarios, periodic checkpoints are used for repair. This function must be enabled only when process-level rescheduling or process-level online recovery is enabled.

## Optimizing the Integration Time<a name="ZH-CN_TOPIC_0000002479386526"></a>

### Recovery Time Optimization (PyTorch)<a name="ZH-CN_TOPIC_0000002479386516"></a>

This section describes the related features that you can choose to shorten the resumable training time on the PyTorch framework, including [Fault Detection Time](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section195202179141), [Collective Communication Initialization Time](#zh-cn_topic_0000002163883997_section725312412292), [Training Rollback and Checkpoint Loading Time](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section71731855173720), and [Operator Compilation Time](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section14599171031417).

**Fault Detection Time<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section195202179141"></a>**

A parameter plane network fault in a cluster may not affect a training job. Therefore, the cluster scheduling components do not forcibly interrupt the job. When the parameter plane network fault affects a training job, the network timeout mechanism of collective communication is triggered. After a default waiting period of 30 minutes, the cluster scheduling components can detect the fault and trigger resumable training. To solve this problem, the PyTorch Adapter plugin (torch_npu) provides a watchdog fault detection function to determine if training jobs are affected and to reduce fault detection time. For details, see [Table 1](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table4822175901415).

**Table 1** Watchdog fault detection

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table4822175901415"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row9823145931412"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"></a>Function Name</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"></a>Watchdog</span> fault detection</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row58231859181412"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"></a>Feature Description</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"></a>When training starts, a monitoring thread is simultaneously started to continuously capture communication exceptions and job execution exceptions. After a fault is detected, an exception is quickly thrown and the training job process is terminated, triggering the rescheduling process.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row138235598144"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"></a>Usage Notes</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"></a>Only supports <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"></a>PyTorch</span> 1.11.0, 2.1.0 and later versions; the <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"></a>PyTorch</span> Adapter plugin (torch_npu) version must be higher than 6.0.RC1.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row11823195941410"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"></a>Key Operations</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"></a>In PyTorch</span> 2.1.0 and later versions, <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"></a>watchdog</span> fault detection is enabled by default, <strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"></a>without the need to manually configure environment variables</strong>.</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"></a>(Optional) To disable <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"></a>watchdog</span> fault detection, modify the following environment variables in the training shell startup script (e.g., train_start.sh).</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"></a>...
# env for breakpoint ckpt
export RESUME_MODE_ENABLE=1
<br>
export HCCL_ASYNC_ERROR_HANDLING=0  <strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"></a>          </strong># For details about this environment variable, see <a href="../../api/environment_variable_description.md#taskd-environment-variables">TaskD Environment Variable Description</a></pre>
</td>
</tr>
</tbody>
</table>

**Collective Communication Initialization Time<a name="zh-cn_topic_0000002163883997_section725312412292"></a>**

Parallel Store multi-thread link setup optimization: When PyTorch creates communication groups, TCP Store is used for information exchange. As the job scale increases, the information processing performance of the native TCP Store degrades, leading to prolonged times for creating communication groups. To solve this problem, torch_npu supports the optimized Parallel Store built on the native TCP Store. For details, see [Table 2](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table14133757143220).

**Table 2**  Parallel Store

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table14133757143220"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row15133115723218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"></a>Function Name</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"></a>Parallel Store</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row2013316574328"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"></a>Feature Description</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"></a>During multi-thread link setup, this function can reduce both the waiting time of the link setup request queue and the overall link setup time.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1913318574324"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"></a>Instructions</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"></a>PyTorch</span> 1.11.0: The version of <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"></a>torch_npu</span> must be higher than 6.0.RC1.</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"></a>PyTorch</span> 2.1.0 and later: The version of <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"></a>torch_npu</span> must be higher than 6.0.RC3.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16133957183217"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"></a>Key Operations</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"></a>In the shell script used to start training (for example, train_start.sh), change the torchrun launch command to torch_npu_run.</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"></a>For example, change</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"></a><strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"></a>torchrun train.py --train_parameter=xxx ....</strong></pre>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"></a>to</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"></a><strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"></a>torch_npu_run train.py --train_parameter=xxx ....</strong></pre>
</td>
</tr>
</tbody>
</table>

- Performance optimization of native HCCL link setup: PyTorch sets up a link between NPUs after the collective communication information is exchanged on the NPU. As the job scale increases, the link setup time increases significantly. To solve this problem, CANN is introduced to optimize the performance of the native HCCL link setup. For details, see [Table 3](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table10637950133911).

    **Table 3**  Native HCCL link setup performance optimization

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table10637950133911"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row763710506398"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"></a>Function Name</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"></a>Native HCCL link setup performance optimization</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row963765019395"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"></a>Feature Description</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"></a>By asynchronously completing collective communication information negotiation, multiple threads reduce both the negotiation time and the overall link setup time.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row263845043913"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"></a>Instructions</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"></a>Only CANN 8.0.RC2 and later versions are supported.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1563845013912"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"></a>Key Operations</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"></a>None</p>
    </td>
    </tr>
    </tbody>
    </table>

- Link setup optimization in RankTable mode: Ascend Operator provides the function of generating a collective communication configuration file (RankTable file, also called `hccl.json`) for PyTorch. Links can be set up in RankTable mode to shorten cluster communication link setup time. For details,, see [Table 4](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1749892464019).

    **Table 4** Link setup for collective communication in RankTable mode

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1749892464019"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row84981324184016"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"></a>Function Name</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"></a>Link setup in RankTable mode</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16498162484017"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"></a>Feature Description</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"></a>Uses <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"></a>Ascend Operator</span> is used to generate the collective communication configuration file for PyTorch tasks, reducing the cluster communication link setup time.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row2499424194015"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"></a>Instructions</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"></a>The</span> version of torch_npu must be higher than 6.0.RC3.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row4499524124018"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"></a>Key Operations</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"></a><ol id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"><li>The parent directory of the hccl.json file is already mounted by default in the startup YAML. You can change it as required.<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"></a>volumes:
           - name: ranktable-dir
             hostPath:
               path: /user/mindx-dl/ranktable  # This host directory must be under a shared directory
               type: DirectoryOrCreate</pre>
    <div class="p" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"></a>Run the following commands to create the specific mount path for the hccl.json file in the host directory and modify the owner.<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"></a>mkdir -m 777 /user/mindx-dl/ranktable/Task_Running_Namespace.Task_Name
    chown 9000:9000 /user/mindx-dl/ranktable/default.pytorch-test</pre>
    </div>
    <div class="p" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"></a>For example:<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"></a>mkdir -m 777 /user/mindx-dl/ranktable/default.pytorch-test
    chown 9000:9000 /user/mindx-dl/ranktable/default.pytorch-test</pre>
    </div>
    </li><li>Modify the training script and add the following environment variable.<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"></a>export RANK_TABLE_FILE=/user/mindx-dl/ranktable/hccl.json</pre>
    </li><li>Modify the training YAML and add the following settings.<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"></a>yaml
          volumeMounts:
          - name: ranktable
            mountPath: /user/mindx-dl/ranktable

           volumes:
           - name: ranktable
             hostPath:
               path: /user/mindx-dl/ranktable/namespace_of_the_running_task.task_name  # Actual path of the hccl.json file in the host directory
    </pre>
    </li></ol>
    </td>
    </tr>
    </tbody>
    </table>

**Training Rollback and Checkpoint Loading Time<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section71731855173720"></a>**

- Asynchronous checkpoint saving: A training job periodically saves checkpoint files to save parameter information. Once a fault is rectified, training is rolled back from the most recently saved checkpoint file for recovery. Each time a checkpoint file is saved, a specific training period is wasted. To ensure training efficiency, the interval for saving checkpoint files is usually large. However, a larger saving interval indicates longer time wasted for training rollback upon each fault. To solve this problem, MindIO ACP is introduced to asynchronously save checkpoints. For details, see [Table 5](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table5173115519372).

    **Table 5** Asynchronous checkpoint saving

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table5173115519372"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row717435514373"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"></a>Function Name</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"></a>Asynchronous</span> checkpoint saving</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row10174115583714"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"></a>Feature Description</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"></a>After checkpoints are obtained from the NPU, they are asynchronously written to storage to minimize training loss and the storage period for each checkpoint saving, thereby reducing the training rollback time.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row6174655153715"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"></a> Instructions</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"></a>Only  cluster scheduling components and <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"></a>MindIO</span> components of version 6.0.RC2 or later are supported.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row171741155133719"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"></a>Key Operations</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"></a>To install and use <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"></a>MindIO</span>, see <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Optimizing Checkpoint Saving and Loading</a></span>.</p>
    </td>
    </tr>
    </tbody>
    </table>

- Efficient checkpoint recovery: During training rollback and recovery, checkpoints must be loaded from storage. Due to the large volume of checkpoint data, directly reading and loading checkpoints from storage takes considerable time. To solve this problem, MindIO ACP is introduced for efficient checkpoint recovery. For details, see [Table 6](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1163115196618).

    **Table 6** Efficient checkpoint recovery

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1163115196618"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row763114191366"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"></a>Function Name</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"></a>Efficient</span> checkpoint recovery</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row14631191914615"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"></a>Feature Description</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"></a>MindIO</span> stores the latest checkpoint in memory, allowing it to be read directly from memory during fault recovery, thereby reducing checkpoint read time.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row26321219766"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"></a>Instructions</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"></a>Only cluster scheduling components and <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"></a>MindIO</span> components of version 6.0.RC2 or later are supported.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row9632219868"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"></a>Key Operations</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"></a>To install and use <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"></a>MindIO</span>, see <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Optimizing Checkpoint Saving and Loading</a></span>.</p>
    </td>
    </tr>
    </tbody>
    </table>

**Operator Compilation Time<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section14599171031417"></a>**

If an operator needs to be re-executed during resumable training, building the operator takes a long time. To solve this problem, you can select the operator binary or operator building cache to reduce the building time. For details, see [Table 7](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table8599191019143) and [Table 8](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table2193759172110).

>[!NOTE]
>The operator binary and operator compliation cache are incompatible. Please choose one of them to use.

**Table 7**  Operator Binary Function Description

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table8599191019143"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1599111016145"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"></a>Function Name</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"></a>Operator binary</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1059931012143"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"></a>Feature Description</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"></a>During operator compilation, the preset operator binary is loaded in advance so that the operator can be executed without compilation.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16599161015147"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"></a>Instructions</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"></a>Only CANN 8.0.RC2 and later versions are supported.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row7600610181419"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"></a>Key Operations</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"></a>In the <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"></a>Python</span> startup script, add the operator binary configuration command to enable the operator binary.</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"></a>torch.npu.set_compile_mode(jit_compile=False)</pre>
</td>
</tr>
</tbody>
</table>

**Table 8** Operator compilation cache

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table2193759172110"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1819335920218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"></a>Function Name</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"></a>Operator compilation cache</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row11193185913215"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"></a>Feature Description</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"></a>Load the operator compilation cache file saved in storage during operator compilation, reducing compilation time after loading.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1719310593218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"></a>Instructions</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"></a>Only CANN 8.0.RC2 and later versions are supported.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1193195962112"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"></a>Key Operations</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"></a><ol id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"><li>In the <span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"></a>Python</span> startup script, add the operator compilation cache configuration command to enable operator compilation cache.<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"></a>torch.npu.set_compile_mode(jit_compile=True)</pre>
</li><li>In the training shell startup script (e.g., train_start.sh), add the following environment variables.<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"></a>export ASCEND_CACHE_PATH<strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"></a>=xxx</strong>   # Add shared storage path
export ASCEND_MAX_OP_CACHE_SIZE=-1    # Recommended when using shared storage; resolves resource contention issues when multiple nodes read shared storage cache</pre>
</li></ol>
</td>
</tr>
</tbody>
</table>

### Recovery Time (MindSpore)<a name="ZH-CN_TOPIC_0000002511346499"></a>

This section describes the optimization items that can be used to shorten the resumable training time on MindSpore, including [Fault Detection Time](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section517194154019), [Training Rollback and Checkpoint Loading Time](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section2743164217401), and [Compilation Cache Time](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section1139444324019).

**Fault Detection Time<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section517194154019"></a>**

SA parameter plane network fault in a cluster may not affect a training job. Therefore, the cluster scheduling components do not forcibly interrupt the job. When the parameter plane network fault affects a training job, the network timeout mechanism of collective communication is triggered. After a default waiting period of 30 minutes, the cluster scheduling components can detect the fault and trigger resumable training. To solve this problem, MindSpore provides a watchdog fault detection function to determine if training jobs are affected and to reduce fault detection time. For details, see [Table 1](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table17897155873217).

**Table 1** Watchdog fault detection

<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table17897155873217"></a>
<table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row289715810326"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"></a>Function Name</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"></a>Watchdog</span> fault detection</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1589716585326"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"></a>Feature Description</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"></a>When training is started, a monitoring thread is started at the same time to continuously obtain communication exceptions and task execution exceptions. After a fault is detected, an exception is quickly thrown, the training process is terminated, and rescheduling is triggered.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row189775853220"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"></a>Instructions</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"></a>Only MindSpore 2.4 and later versions are supported.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row5898058143215"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"></a>Key Operations</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"></a>MindSpore <strong id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"></a>enables</strong><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"></a> watchdog</span> fault detection <strong id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"></a>by default, requiring no manual configuration</strong>. If you need to disable <span id="ph1052517411176"><a name="ph1052517411176"></a><a name="ph1052517411176"></a>this</span> function, add the following bold fields to the model configuration file.</p>
<pre class="screen" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"></a>...
context:
  <strong id="b393317297113"><a name="b393317297113"></a><a name="b393317297113"></a>ascend_config:</strong>
    <strong id="b12660461696"><a name="b12660461696"></a><a name="b12660461696"></a>hccl_watchdog: False</strong>
...</pre>
</td>
</tr>
</tbody>
</table>

**Training Rollback and Checkpoint Loading Time<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section2743164217401"></a>**

- Asynchronous checkpoint saving: A training job periodically saves checkpoint files to save parameter information. Once a fault is rectified, training is rolled back from the most recently saved checkpoint file for recovery. Each time a checkpoint file is saved, a specific training period is wasted. To ensure training efficiency, the interval for saving checkpoint files is usually large. However, a larger saving interval indicates longer time wasted for training rollback upon each fault. To solve this problem, MindIO ACP is introduced to asynchronously save checkpoints. For details, see [Table 2](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table56063271212).

    **Table 2** Asynchronous checkpoint saving

    <a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table56063271212"></a>
    <table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row4606162713214"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"></a>Function Name</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"></a>Asynchronous</span> checkpoint saving</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row260619272216"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"></a>Feature Description</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"></a>After checkpoints are obtained from the NPU, they are asynchronously written to storage to minimize training loss and the storage period for each checkpoint saving, thereby reducing the training rollback time.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row126061827152113"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"></a>Instructions</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"></a>Only cluster scheduling components and <span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"></a>MindIO</span> components of version 6.0.RC2 and later are suppported.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row136069278219"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"></a>Key Operations</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"></a>To install and use <span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"></a>MindIO</span>, see <span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Optimizing Checkpoint Saving and Loading</a></span>.</p>
    </td>
    </tr>
    </tbody>
    </table>

- Efficient checkpoint recovery: During training rollback and recovery, checkpoints must be loaded from storage. Due to the large volume of checkpoint data, directly reading and loading checkpoints from storage takes considerable time. To solve this problem, MindIO ACP is introduced for efficient checkpoint recovery. For details, see [Table 3](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table66066274216).

    **Table 3** Efficient checkpoint recovery

    <a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table66066274216"></a>
    <table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row106071271216"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"></a>Function Name</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"></a>Efficient</span> checkpoint recovery</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row360715276216"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"></a>Feature Description</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"></a>Store the latest checkpoint in memory, allowing direct read from memory during fault recovery to reduce checkpoint read time.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1860772716217"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"></a>Instructions</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"></a>Only cluster scheduling components and <span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"></a>MindIO</span> components of version 6.0.RC2 and later are supported.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row196071127102110"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"></a>Key Operations</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"></a>To install and use <span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"></a>MindIO</span>, see <span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Opyimizing Checkpoint Saving and Loading</a></span>.</p>
    </td>
    </tr>
    </tbody>
    </table>

**Compilation Cache Time<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section1139444324019"></a>**

During resumable training, a computational graph needs to be built. However, this process takes a long time in foundation model scenarios. To solve this problem, MindSpore can store a building cache file during the first building. During fault recovery, the graph building cache in storage can be directly read to reduce the graph building time. For details, see [Table 4](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table175224282139).

**Table 4**  Graph compilation cache

<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table175224282139"></a>
<table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row135238284132"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"></a>Function Name</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"></a>Graph compilation cache</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1052322818133"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"></a>Feature Description</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"></a>During graph compilation, the cache file stored on the storage device is loaded to help reduce compilation time.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row5523628191316"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"></a>Instructions</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"></a>Only MindSpore 2.3.0 and later are supported.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row352313282136"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"></a>Key Operations</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><div class="p" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"></a>In the training shell Startup Script (e.g., train_start.sh), add the following environment variables.<pre class="screen" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"></a>export MS_COMPILER_CACHE_ENABLE=1  # Enable graph compilation cache
export MS_COMPILER_CACHE_PATH=xxx  # Set the graph compilation cache path</pre>
</div>
</td>
</tr>
</tbody>
</table>

## Configuring Proactive HCCL Link Setup<a name="ZH-CN_TOPIC_0000002511346489"></a>

If faults occur in the HCCL link setup phase, process-level rescheduling or process-level online recovery will fail. If HCCL link setup is required in other training phases in addition to the training initialization phase, you can set up the link in advance to avoid faults during the setup process.

**PyTorch Single-Operator Scenario<a name="section145466566911"></a>**

n the PyTorch single-operator scenario, HCCL links are set up in lazy loading mode. After a Torch communication group is set up, its first operator triggers the creation of the HCCL communicator. After the creation, the inter-rank link is set up. Therefore, to ensure all communicators are linked during training initialization, a communication operator must be dispatched to each group at that stage.

The following is an example of actively creating a communication group:

```Python
rank = 0 # Set the rank of this process
sub_ranks = [0, 1, 2]  # Assume a communication group containing 0, 1, and 2
groupX = torch.distributed.new_group(ranks=sub_ranks,...) # Create communication group X
test_tensor = torch.ones(1).to(f'npu:{rank}') * (rank + 1)  # Construct a test data tensor
torch.distributed.all_reduce(test_tensor, op=dist.ReduceOp.SUM, group=groupX)  # Execute the all reduce operator in communication group X
```
