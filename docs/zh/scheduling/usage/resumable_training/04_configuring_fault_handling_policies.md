# 配置故障处理<a name="ZH-CN_TOPIC_0000002479386478"></a>

## 配置Job级别重调度<a name="ZH-CN_TOPIC_0000002479226580"></a>

Job级别重调度默认开启，用户只需完成制作镜像的步骤及准备任务YAML的步骤即可。Job级别重调度的特性介绍、使用约束、支持的产品型号及原理请参见[Job级别重调度](./01_solutions_principles.md#job级别重调度)。

**准备任务YAML<a name="zh-cn_topic_0000002098814658_section463203519254"></a>**

在任务YAML中，新增以下字段，开启Job级别重调度。

```Yaml
... 
metadata:  
   labels:  
     ...  
     fault-scheduling: "force"
```

## 配置Pod级别重调度<a name="ZH-CN_TOPIC_0000002479226508"></a>

本章节将指导用户了解配置Pod级别重调度的关键步骤。Pod级别重调度的特性介绍、使用约束、支持的产品型号及原理请参见[Pod级别重调度](./01_solutions_principles.md#pod级别重调度)。

**构建镜像<a name="zh-cn_topic_0000002098654822_section11751140165911"></a>**

使用Dockerfile构建容器镜像，新增启动命令。示例如下。

```shell
# MindCluster断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，请根据实际情况填写   
# 可选，PyTorch框架下，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN pip install $TASKD_WHL 
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用Pod级别重调度需配置以下命令
RUN pip install $TASKD_WHL
```

**准备任务YAML<a name="zh-cn_topic_0000002098654822_section027517423166"></a>**

在任务YAML中，新增以下字段，开启Pod级别重调度，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

<pre codetype="yaml">
... 
metadata:  
   labels:  
     ...  
     <strong>pod-rescheduling: "on"</strong>
     <strong>fault-scheduling: "force"   # 可以根据实际情况选择force或者grace, 配置为force时Pod不能配置使用主机网络</strong>
...
        spec:
...
           containers:
...
             <strong>ports:</strong>                          
               <strong>- containerPort: 9601</strong>              
                 <strong>name: taskd-port</strong>
...</pre>

**适配训练脚本<a name="zh-cn_topic_0000002098654822_section17330181621710"></a>**

1. 在启动脚本（例如train\_start.sh）中，新增以下加粗字段，示例如下。

    <pre codetype="shell">
    ...
    <strong>export MS_ENABLE_TFT="{RSC:1}"    # MindSpore场景下配置此字段开启Pod级别重调度</strong>
    ...
    # 可选，PyTorch场景下，设置容器内重启次数和训练进程监控间隔
       logger "server id is: ""${server_id}" 
       if [ "${framework}" == "PyTorch" ]; then 
         get_env_for_pytorch_multi_node_job 
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT <strong>--max_restarts 32767</strong>" </pre>

     其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

2. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

    2. 在训练脚本（例如train\_start.sh）中增加以下代码，拉起TaskD Manager。在以下代码中：

        - TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。
        - TASKD\_PROCESS\_ENABLE环境变量配置说明：若任务YAML中“recover-strategy”未配置恢复策略且未使能亚健康热切，需要配置**export TASKD\_PROCESS\_ENABLE="off"**；若“recover-strategy”配置了恢复策略或使能了亚健康热切，则无需配置**export TASKD\_PROCESS\_ENABLE="off"**。

        ```shell
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="off" 
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
        # MindSpore框架下
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
        ```

## 配置进程级别重调度<a name="ZH-CN_TOPIC_0000002511426407"></a>

本章节将指导用户了解配置进程级别重调度的关键步骤。进程级别重调度的特性介绍、使用约束、支持的产品型号及原理请参见[进程级别重调度](./01_solutions_principles.md#进程级别重调度)。

**构建镜像<a name="zh-cn_topic_0000002134293721_section18253151810133"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```shell
# MindCluster无损失断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，MINDIO_TTP_PKG为MindIO的whl安装包的路径，请根据实际情况填写  
# 可选，PyTorch框架下，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN pip3 install $TASKD_WHL   
RUN pip3 install $MINDIO_TTP_PKG 
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用进程级别重调度需配置以下命令
RUN pip3 install $MINDIO_TTP_PKG
RUN pip3 install $TASKD_WHL
```

**准备任务YAML<a name="zh-cn_topic_0000002134293721_section2492121411271"></a>**

在任务YAML中，修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

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

在任务YAML中，新增以下字段，开启进程级别重调度。recover-strategy是训练进程恢复使用的策略，其中的recover代表开启进程级别恢复。

目前进程级别重调度支持以下2种方式，用户可根据实际使用场景，选择其中一种方式进行使用。

- 方式一：故障后迁移故障Pod到健康节点

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
         <strong>recover-strategy: "recover"   # 任务可用恢复策略（retry：进程级在线恢复；recover：进程级别重调度；recover-in-place: 进程级原地恢复；elastic-training：弹性训练；dump：保存临终遗言；exit：退出训练），6种策略可随意组合，策略之间由逗号分割</strong>
     ... 
    ...
    spec:
      replicaSpecs:
        Master:
          template:
            spec:
              containers:
              - name: ascend       # do not modify
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
              - name: ascend # do not modify
                ...
                args:
                  - |
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
    ...</pre>

- 方式二：故障后不迁移故障Pod，仅重启故障进程

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
         <strong>recover-strategy: "recover-in-place"   # 任务可用恢复策略（retry：进程级在线恢复；recover：进程级别重调度；recover-in-place: 进程级原地恢复；elastic-training：弹性训练；dump：保存临终遗言；exit：退出训练），6种策略可随意组合，策略之间由逗号分割</strong>
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
        Worker:
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
    ...</pre>

**适配训练脚本<a name="zh-cn_topic_0000002134293721_section1829103214273"></a>**

1. （可选）在启动脚本（例如train\_start.sh）中，配置--max_restarts参数，示例如下。

    <pre codetype="shell">
    # PyTorch场景下，设置训练进程监控间隔
    ...
       logger "server id is: ""${server_id}"
       if [ "${framework}" == "PyTorch" ]; then
         get_env_for_pytorch_multi_node_job
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT  <strong>--max_restarts 32767</strong>"
    ...</pre>

     其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

2. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

    2. 在训练脚本（例如train\_start.sh）中增加以下代码，拉起TaskD Manager。在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

        ```shell
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="on"
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
           export MASTER_ADDR=${POD_IP}
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
        # MindSpore框架下
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
        ```

## 配置进程级在线恢复<a name="ZH-CN_TOPIC_0000002479386492"></a>

本章节将指导用户了解配置进程级在线恢复的关键步骤。进程级在线恢复的特性介绍、使用约束、支持的产品型号及原理请参见[进程级在线恢复](./01_solutions_principles.md#进程级在线恢复)。

**构建镜像<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```shell
# MindCluster断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，MINDIO_TTP_PKG为MindIO的whl安装包的路径，请根据实际情况填写
# 可选，PyTorch框架下，使用优雅容错、Pod级别重调度、进程级别重调度或进程级在线恢复时必须配置以下命令
RUN pip3 install $TASKD_WHL 
RUN pip3 install $MINDIO_TTP_PKG 
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用进程级在线恢复需配置以下命令
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
```

**准备任务YAML<a name="zh-cn_topic_0000002134174097_section98717593512"></a>**

在任务YAML中，新增以下加粗字段，开启进程级别恢复，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

<pre codetype="yaml">
...  
   labels:  
     ...  
     <strong>fault-scheduling: "grace"</strong>
 ... 
...  
   annotations:  
     ...  
     <strong>recover-strategy: "retry"    # 任务可用恢复策略（retry：进程级在线恢复；recover：进程级别重调度；recover-in-place: 进程级原地恢复；elastic-training：弹性训练；dump：保存临终遗言；exit：退出训练），6种策略可随意组合，策略之间由逗号分割</strong>
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
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
            <strong>ports:</strong>                          
               <strong>- containerPort: 9601</strong>              
                 <strong>name: taskd-port</strong>
...</pre>

MindSpore场景下，用户需修改模型参数配置YAML。打开QWEN3\_for\_MS\_code/configs/qwen3/pretrain\_qwen3\_32b\_4k.yaml文件，在代码中增加以下加粗字段。

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

**适配训练脚本<a name="zh-cn_topic_0000002134174097_section189248183358"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os
         
        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX          # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
         
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE] 
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

    2. 在训练脚本（例如train\_start.sh）中增加以下代码，拉起TaskD Manager。在以下代码中，TASKD\_SO\_PATH和export LD\_PRELOAD两条语句的作用是将安装TaskD后libtaskd.so的路径配置到环境变量LD\_PRELOAD中。如果这两条语句配置不成功，可通过手动执行pip show taskd命令获取Location的值拼接上/taskd/python/cython\_api/libs/libtaskd.so，然后通过export设置。

        ```shell
        TASKD_SO_PATH="$(pip show taskd | awk '/^Location: / {print $2"/taskd/python/cython_api/libs/libtaskd.so"}')"
        export LD_PRELOAD=$TASKD_SO_PATH:$LD_PRELOAD
        export TASKD_PROCESS_ENABLE="on"
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
           export MASTER_ADDR=${POD_IP}
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
        # MindSpore框架下
        if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
           python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
        ```

2. （可选）在启动脚本（例如train\_start.sh）中，新增--max\_restarts参数，示例如下。

    <pre codetype="shell">
    ... 
       logger "server id is: ""${server_id}" 
       if [ "${framework}" == "PyTorch" ]; then 
         get_env_for_pytorch_multi_node_job 
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT <strong>--max_restarts 32767</strong>" </pre>

        其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

    - MindSpeed场景下，用户需修改训练启动脚本train\_start.sh，在代码中增加如下字段，示例如下。

        ```shell
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"   # 开启HCCL算子的重执行特性（算子级在线恢复）。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
        export HCCL_ASYNC_ERROR_HANDLING=0
        ```

    - MindFormers场景下，用户需修改训练启动脚本msrun\_launcher.sh文件，在代码中增加如下字段，示例如下。

        ```shell
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  # 此环境变量用于配置是否开启HCCL算子的重执行特性。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
        ```

>[!NOTE] 
>用户若要测试进程级在线恢复功能，可参考[进程级在线恢复验证](../../appendix.md#进程级在线恢复验证)进行配置。

## 配置算子级在线恢复<a name="ZH-CN_TOPIC_0000002511426477"></a>

本章节将指导用户了解配置算子级在线恢复的关键步骤。算子级在线恢复的特性介绍、使用约束、支持的产品型号及原理请参见[算子级在线恢复](./01_solutions_principles.md#算子级在线恢复)。

**配置环境变量<a name="section12610013287"></a>**

使用算子级在线恢复前，用户需在启动训练的脚本中配置环境变量HCCL\_OP\_RETRY\_ENABLE和HCCL\_OP\_RETRY\_PARAMS。关于该环境变量的详细说明请参见《[CANN 环境变量参考](https://www.hiascend.com/document/detail/zh/canncommercial/850/maintenref/envvar/envref_07_0001.html)》。配置示例如下。

```shell
export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"     # 是否开启HCCL算子的重执行特性
export HCCL_OP_RETRY_PARAMS="MaxCnt:3, HoldTime:5000, IntervalTime:1000"    # 配置HCCL算子重执行的具体参数，包括最大重执行次数、第一次重执行的等待时间以及两次重执行的间隔时间
```

## 配置借轨通信任务暂停与回切<a name="ZH-CN_TOPIC_0000002511346495"></a>

### PyTorch场景（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002511426445"></a>

本章节将指导用户了解配置借轨通信任务暂停与回切的关键步骤。借轨通信任务暂停与回切的特性介绍、使用约束、支持的产品型号及原理请参见[借轨通信任务暂停与回切](./01_solutions_principles.md#借轨通信任务暂停与回切)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- 在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../../installation_guide/03_installation/manual_installation/02_ascend_docker_runtime.md)、[Ascend Operator](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)、[ClusterD](../../installation_guide/03_installation/manual_installation/06_clusterd.md)、[Ascend Device Plugin](../../installation_guide/03_installation/manual_installation/04_ascend_device_plugin.md)和[Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装[torch\_npu](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（8.2.RC1及以上版本）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)和[MindIO](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）

**操作步骤<a name="section188080175496"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在训练进程内部拉起TaskD  Worker。

    1. 拉起TaskD  Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

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
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码，拉起TaskD  Manager。

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
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

2. 修改训练框架代码。
    1. 进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3”目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

        ```text
        root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
        scripts/
        └── train_start.sh
        ```

    2. 配置训练启动脚本train\_start.sh，在代码中增加如下字段。

        ```shell
        # 开启HCCL算子的重执行特性。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
        export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  
        ```

3. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启进程级在线恢复，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

    <pre codetype="yaml">
    ...  
        labels:  
          ...  
          <strong>fault-scheduling: "grace"</strong>
       ... 
    ...  
        annotations:  
          ...  
          <strong>recover-strategy: "retry"    # 任务可用恢复策略，取值为retry，表示开启进程级在线恢复</strong>
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
                 <strong>ports: </strong>                         
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
                     ...
                     bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                       ...
                 <strong>ports:</strong>                          
                   <strong>- containerPort: 9601</strong>              
                     <strong>name: taskd-port</strong>
    ...</pre>

### MindSpore场景（基于MindFormers）<a name="ZH-CN_TOPIC_0000002511346443"></a>

本章节将指导用户了解配置借轨通信任务暂停与回切的关键步骤。借轨通信任务暂停与回切的特性介绍、使用约束、支持的产品型号及原理请参见[借轨通信任务暂停与回切](./01_solutions_principles.md#借轨通信任务暂停与回切)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- 在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../../installation_guide/03_installation/manual_installation/02_ascend_docker_runtime.md)、[Ascend Operator](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)、[ClusterD](../../installation_guide/03_installation/manual_installation/06_clusterd.md)、[Ascend Device Plugin](../../installation_guide/03_installation/manual_installation/04_ascend_device_plugin.md)和[Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装MindSpore（2.7.0及以上版本）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindformers训练镜像mindspore框架)（8.2.RC1及以上版本）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindformers训练镜像mindspore框架)和[MindIO](./07_using_resumable_training_on_the_cli.md#制作mindformers训练镜像mindspore框架)（7.1.RC1及以上版本）

**操作步骤<a name="section9479182019317"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
    1. 拉起TaskD  Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

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
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码拉起TaskD  Manager。

            ```shell
            export TASKD_PROCESS_ENABLE="on"
            if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then
                python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
            fi
                
            msrun ...
            ```

    2. 拉起TaskD  Worker。修改./mindformers/trainer/base\_trainer.py文件，增加如下加粗字段。

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

2. 修改训练框架代码，打开借轨开关。

    编辑启动脚本QWEN3\_for\_MS\_code/scripts/msrun\_launcher.sh文件，在代码中增加如下字段。

    ```shell
    export MS_ENABLE_TFT="{TTP:1,TSP:1}"           # 开启临终遗言和借轨回切
    export HCCL_OP_RETRY_ENABLE="L0:0, L1:1, L2:1"  # 此环境变量用于配置是否开启HCCL算子的重执行特性。重执行是指当执行通信算子时报SDMA或者RDMA CQE类型的错误，HCCL会尝试重新执行此通信算子。
    ```

    >[!NOTE] 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

3. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启进程级在线恢复，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

    <pre codetype="yaml">
    ...  
        labels:  
          ...  
          <strong>fault-scheduling: "grace"</strong>
      ... 
    ...  
        annotations:  
          ...  
          <strong>recover-strategy: "retry"    # 任务可用恢复策略，取值为retry，表示开启进程级在线恢复</strong>
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
                    ...
                    bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                      ...
                <strong>ports:</strong>                          
                  <strong>- containerPort: 9601</strong>              
                    <strong>name: taskd-port</strong>
    ...</pre>

## 配置优雅容错<a name="ZH-CN_TOPIC_0000002511346501"></a>

>[!NOTE] 
>该功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。

本章节将指导用户了解配置优雅容错的关键步骤。优雅容错的特性介绍、使用约束、支持的产品型号及原理请参见[（可选）优雅容错](./01_solutions_principles.md#可选优雅容错)。

**构建镜像<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```shell
# MindCluster断点续训适配脚本，MINDIO_TTP_PKG为MindIO的whl安装包的路径，请根据实际情况填写
RUN pip3 install $MINDIO_TTP_PKG 
```

**适配训练脚本<a name="section731511818483"></a>**

在启动脚本（例如train\_start.sh）中，新增以下字段，示例如下。

```shell
...
export MS_ENABLE_TFT="{RSC:1}"      # MindSpore场景下配置此字段开启优雅容错
...
```

**配置启动YAML<a name="zh-cn_topic_0000002138594553_section18371651403"></a>**

修改Ascend Device Plugin组件的启动YAML，设置-hotReset=1开启热复位，使用优雅容错模式。**注意：优雅容错和进程级别重调度、进程级在线恢复不可同时开启。**

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
                 -volcanoType=true                    # 重调度场景下必须使用Volcano
                 -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                 -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]；单位为秒
                 -hotReset=1      # 使用断点续训时，可以在Job级或Pod级重调度的基础上，开启热复位功能，使用优雅容错模式
                 -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                 -logLevel=0" ]
        securityContext:
          privileged: true
          readOnlyRootFilesystem: true
...
```

## 配置在线压测<a name="ZH-CN_TOPIC_0000002511426487"></a>

### PyTorch场景（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002479386572"></a>

本章节将指导用户了解配置在线压测的关键步骤。在线压测的特性介绍、使用约束、支持的产品型号等请参见[在线压测](./01_solutions_principles.md#在线压测)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- 在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../../installation_guide/03_installation/manual_installation/02_ascend_docker_runtime.md)、[Ascend Operator](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)、[ClusterD](../../installation_guide/03_installation/manual_installation/06_clusterd.md)、[Ascend Device Plugin](../../installation_guide/03_installation/manual_installation/04_ascend_device_plugin.md)和[Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装[torch\_npu](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（8.2.RC1及以上版本）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)和[MindIO](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（7.2.RC1及以上版本）

**操作步骤<a name="section188080175496"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在训练进程内部拉起TaskD  Worker。

    1. 拉起TaskD  Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

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
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码，拉起TaskD  Manager。

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
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

2. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启进程级别重调度，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

      <pre codetype="yaml">
        ...  
           labels:  
             ...  
             <strong>fault-scheduling: "grace"</strong>
         ... 
        ...  
           annotations:  
             ...  
             <strong>recover-strategy: "recover"    # 任务可用恢复策略，取值为recover，表示开启进程级别重调度</strong>
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

### MindSpore场景（基于MindFormers）<a name="ZH-CN_TOPIC_0000002479226554"></a>

本章节将指导用户了解配置在线压测的关键步骤。在线压测的特性介绍、使用约束、支持的产品型号等请参见[在线压测](./01_solutions_principles.md#在线压测)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- 在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../../installation_guide/03_installation/manual_installation/02_ascend_docker_runtime.md)、[Ascend Operator](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)、[ClusterD](../../installation_guide/03_installation/manual_installation/06_clusterd.md)、[Ascend Device Plugin](../../installation_guide/03_installation/manual_installation/04_ascend_device_plugin.md)和[Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装MindSpore（2.7.0及以上版本）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindformers训练镜像mindspore框架)（8.2.RC1及以上版本）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindformers训练镜像mindspore框架)和[MindIO](./07_using_resumable_training_on_the_cli.md#制作mindformers训练镜像mindspore框架)（7.2.RC1及以上版本）

**操作步骤<a name="section9479182019317"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
    1. 拉起TaskD  Manager。
        1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

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
            >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

        2. 在训练脚本中增加以下代码拉起TaskD  Manager。

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
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

3. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启进程级别重调度，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

      <pre codetype="yaml">
        ...  
           labels:  
             ...  
             <strong>fault-scheduling: "grace"</strong>
         ... 
        ...  
           annotations:  
             ...  
             <strong>recover-strategy: "recover"    # 任务可用恢复策略，取值为recover，表示开启进程级别重调度</strong>
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
                    command:                           # training command, which can be modified
                      - /bin/bash
                      - -c
                      - |
                       cd /job/code/;bash scripts/msrun_launcher.sh "run_mindformer.py --config configs/qwen3/pretrain_qwen3_32b_4k.yaml --auto_trans_ckpt False --use_parallel True --run_mode train"
                    <strong>ports:</strong>                          
                      <strong>- containerPort: 9601</strong>              
                        <strong>name: taskd-port</strong>
        ...</pre>

## 配置亚健康热切<a name="ZH-CN_TOPIC_0000002511426471"></a>

本章节将指导用户了解配置亚健康热切的关键步骤。亚健康热切的特性介绍、使用约束、支持的产品型号及原理请参见[亚健康热切](./01_solutions_principles.md#亚健康热切)。

**构建镜像<a name="zh-cn_topic_0000002134174097_section178178450348"></a>**

使用Dockerfile构建容器镜像，新增启动命令。示例如下。

```shell
# MindCluster断点续训适配脚本，TASKD_WHL为TaskD whl安装包的路径，MINDIO_TTP_PKG为MindIO的whl安装包的路径，MINDSPORE_WHL为MindSpore的whl安装包的路径，请根据实际情况填写
# 可选，PyTorch框架下，使用亚健康热切时必须配置以下命令
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDIO_TTP_PKG
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py

# 可选，MindSpore框架下，使用亚健康热切需配置以下命令
RUN pip3 install $MINDIO_TTP_PKG
RUN pip3 install $TASKD_WHL
RUN pip3 install $MINDSPORE_WHL
```

**准备任务YAML<a name="zh-cn_topic_0000002134174097_section98717593512"></a>**

在任务YAML中，新增以下字段，配置亚健康热切，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

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

**适配训练脚本<a name="zh-cn_topic_0000002098654822_section17330181621710"></a>**

在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。

1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

    ```Python
    from taskd.api import init_taskd_manager, start_taskd_manager
    import os
     
    job_id=os.getenv("MINDX_TASK_ID")
    node_nums=XX          # 用户填入任务节点总数
    proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
     
    init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
    start_taskd_manager()
    ```

    >[!NOTE] 
    >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

2. 在训练脚本中增加以下代码拉起TaskD  Manager。

    ```shell
    export TASKD_PROCESS_ENABLE="on" 
     
    # PyTorch框架下
    if [[ "${RANK}" == 0 ]]; then
        export MASTER_ADDR=${POD_IP} 
        python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
    fi 
          
    torchrun ...
     
    # MindSpore框架下
    if [[ "${MS_SCHED_HOST}" == "${POD_IP}" ]]; then 
        python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &   # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
    fi 
          
    msrun ...
    ```

    >[!NOTE] 
    >如果训练中出现报错“the libtaskd.so has not been loaded”，则需在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。
    >
    >```shell
    >export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/lib/python3.10/dist-packages/taskd/python/cython_api/libs/libtaskd.so
    >```
    >
    >- libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。
    >- libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。
    >
    >     ```shell
    >     pip show taskd
    >     ```

## 配置弹性训练<a name="ZH-CN_TOPIC_0000002511346471"></a>

本章节将指导用户了解配置弹性训练的关键步骤。弹性训练的特性介绍、使用约束、支持的产品型号及原理请参见[弹性训练](./01_solutions_principles.md#弹性训练)。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- 在相应节点上完成以下组件的安装：[Ascend Docker Runtime](../../installation_guide/03_installation/manual_installation/02_ascend_docker_runtime.md)、[Ascend Operator](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)、[ClusterD](../../installation_guide/03_installation/manual_installation/06_clusterd.md)、[Ascend Device Plugin](../../installation_guide/03_installation/manual_installation/04_ascend_device_plugin.md)和[Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装[torch\_npu](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（7.1.RC1及以上版本）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（8.2.RC1及以上版本）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)和[MindIO](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（7.2.RC1及以上版本）

**操作步骤<a name="section188080175496"></a>**

1. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager。
    1. 创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

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
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

    2. 在训练脚本中增加以下代码，拉起TaskD  Manager。

        ```shell
        sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        export TASKD_PROCESS_ENABLE="on"
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi
            
        torchrun ...
        ```

2. 修改任务YAML。

    在任务YAML中新增以下加粗字段，开启弹性训练，并修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

      <pre codetype="yaml">
        ...  
           labels:  
             ...  
             <strong>fault-scheduling: "grace"</strong>
         ... 
        ...  
           annotations:  
             ...
             <strong>wait-reschedule-timeout: "270" # 进程级恢复等待故障节点重调度的超时时间，默认为270秒，取值范围为30~270。进程级恢复和弹性训练均开启时，等待此时间后若故障节点调度成功，则进行进程级恢复，否则触发弹性训练</strong>
             <strong>recover-strategy: "elastic-training"    # 任务可用恢复策略，取值为elastic-training，表示开启弹性训练</strong>
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
                      <strong>- name: MINDIO_WAIT_MINDX_TIME         # 未开启进程级恢复，开启弹性训练场景下建议配置60以上</strong>
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
                      <strong>- name: MINDIO_WAIT_MINDX_TIME         # 未开启进程级恢复，开启弹性训练场景下建议配置60以上</strong>
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

3. 修改训练框架代码。

    进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/train/resumable-training/fault-tolerance/without-ranktable/pytorch/Qwen3”目录下的train\_start.sh文件，在管理节点构造成如下的目录结构。

    ```text
    root@ubuntu:/data/atlas_dls/public/code/QWEN3_for_PyTorch_2.7_code/scripts#
    scripts/
    └── train_start.sh
    ```

## 参数说明<a name="ZH-CN_TOPIC_0000002511346491"></a>

不同的故障处理模式需要配置的参数各不相同，如[表1](#table1247342123814)所示，每个参数所表示的含义及填写说明详见[表2](#zh-cn_topic_0000002163392281_table1474820818115)。Ascend Operator在进程级别重调度、进程级在线恢复、进程级原地恢复和弹性训练场景下，会根据用户配置的recover-strategy和pod-rescheduling注入不同的环境变量，自动给任务打上process-recover-enable=on标签开启进程级恢复开关，无需用户手动指定。具体注入的环境变量如[表3](#table10283161512105)所示。

**表 1**  故障处理所需参数

<a name="table1247342123814"></a>
<table><tbody><tr id="row624717420389"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1724711429388"><a name="p1724711429388"></a><a name="p1724711429388"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="13%"><p id="p882622715436"><a name="p882622715436"></a><a name="p882622715436"></a>Job级别重调度</p>
</td>
<td class="cellrowborder" valign="top" width="11.92%"><p id="p117701032194319"><a name="p117701032194319"></a><a name="p117701032194319"></a>Pod级别重调度</p>
</td>
<td class="cellrowborder" valign="top" width="13.07%"><p id="p182471842183813"><a name="p182471842183813"></a><a name="p182471842183813"></a>进程级别重调度（recover策略）</p>
</td>
<td class="cellrowborder" valign="top" width="13.87%"><p id="p35155320438"><a name="p35155320438"></a><a name="p35155320438"></a>进程级别原地恢复</p>
<p id="p1259532434"><a name="p1259532434"></a><a name="p1259532434"></a>（recover-in-place策略）</p>
</td>
<td class="cellrowborder" valign="top" width="11.87%"><p id="p37521612448"><a name="p37521612448"></a><a name="p37521612448"></a>进程级在线恢复</p>
</td>
<td class="cellrowborder" valign="top" width="11.32%"><p id="p14247442143812"><a name="p14247442143812"></a><a name="p14247442143812"></a>优雅容错</p>
</td>
<td class="cellrowborder" valign="top" width="9.71%"><p id="p84114121442"><a name="p84114121442"></a><a name="p84114121442"></a>弹性训练</p>
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
<tr id="row448045664010"><td class="cellrowborder" valign="top" width="15.24%"><p id="p1848011565403"><a name="p1848011565403"></a><a name="p1848011565403"></a>--enable-high-availability（MindSpeed-LLM侧参数）</p>
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
<tr id="row112463954119"><td class="cellrowborder" valign="top" width="15.24%"><p id="p76389163416"><a name="p76389163416"></a><a name="p76389163416"></a>--enable-hbmfault-repair（MindSpeed-LLM侧参数）</p>
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
<tr id="row3150821154117"><td class="cellrowborder" valign="top" width="15.24%"><p id="p7151102124114"><a name="p7151102124114"></a><a name="p7151102124114"></a>--enable-worker-reboot（MindSpeed-LLM侧参数）</p>
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
<tr id="row4799183364111"><td class="cellrowborder" valign="top" width="15.24%"><p id="p7799233154115"><a name="p7799233154115"></a><a name="p7799233154115"></a>--enable-elastic-training（MindSpeed-LLM侧参数）</p>
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

**表 2**  参数填写说明

<a name="zh-cn_topic_0000002163392281_table1474820818115"></a>

|参数名称|参数位置| 参数说明                                                                                                                                                                                                                                                                                 |
|--|--|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|hotReset|Ascend Device Plugin组件的启动YAML| 优雅容错功能开关。<ul><li>取值为1：使用断点续训时，可以在Job级或Pod级重调度的基础上，开启热复位功能，使用优雅容错模式；</li><li>取值为2：使用进程级恢复时，请将hotReset参数值设置为2，开启离线恢复模式。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>取值为1对应的功能已经日落，请配置其他取值。</p></div></div>                            |
|pod-rescheduling|训练任务YAML的metadata.labels| <ul><li>on：开启Pod级别重调度。</li><li>其他值或不使用该字段：关闭Pod级别重调度。</li></ul>                                                                                                                                                                                                                      |
|fault-scheduling|训练任务YAML的metadata.labels| 重调度开关。                                                                                                                                                                                                                                                                               |
|process-recover-enable|训练任务YAML的metadata.labels| <ul><li>on：开启进程级别重调度及进程级在线恢复。进程级别重调度和优雅容错不能同时开启，若同时开启，断点续训将通过job级重调度恢复训练。</li><li>pause：暂时关闭进程级别重调度及进程级在线恢复。</li><li>off或不使用该字段：关闭进程级别重调度及进程级在线恢复。</li></ul>                                                                                                                         |
|recover-strategy|训练任务YAML的metadata.annotations| 任务可用恢复策略。<ul><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>elastic-training：弹性训练。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>                                                                                                          |
|PROCESS_RECOVER|训练任务YAML的spec.replicaSpecs.{ Master \|Scheduler\| Worker}.template.spec.containers.env| 进程级别重调度及进程级在线恢复Elastic Agent/TaskD侧总开关。<ul><li>on：开启。</li><li>off：关闭。</li></ul>                                                                                                                                                                                                      |
|ELASTIC_PROCESS_RECOVER_ENABLE|启动训练YAML的spec.replicaSpecs.{ Master\|Scheduler\| Worker}. template.spec.containers.args| Elastic Agent侧进程级别重调度、进程级在线恢复、临终CKPT恢复功能开关。<ul><li>取值为1：开启本功能。</li><li>其他值：关闭本功能。<p>关闭本功能时，MindIO侧相关功能需同时关闭。</p></li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>Elastic Agent组件已经日落，相关资料将于2026年12月30日的版本删除。该环境变量会随之删除。</p></div></div> |
|ENABLE_RESTART_FAULT_PROCESS|启动训练YAML的spec.replicaSpecs.{ Master\|Scheduler\| Worker}. template.spec.containers.args| Elastic Agent/TaskD组件开启故障进程原地恢复功能的开关。<ul><li>on：开启本功能；</li><li>其他值：关闭本功能</li></ul>                                                                                                                                                                                                   |
|--enable-high-availability|训练脚本pretrain_gpt.py的启动参数| 故障快速恢复特性开关，默认关闭，配置后即开启临终遗言功能。                                                                                                                                                                                                                                                        |
|--enable-hbmfault-repair|训练脚本pretrain_gpt.py的启动参数| 进程级在线恢复功能开关，默认关闭，配置后对片上内存进行故障检测，并完成在线修复。需同时开启enable-high-availability。                                                                                                                                                                                                               |
|--enable-worker-reboot|训练脚本pretrain_gpt.py的启动参数| 进程级别重调度功能开关，默认关闭。配置后在发生一般性故障时，进行进程级别调度。需同时开启enable-high-availability。                                                                                                                                                                                                                |
|--enable-elastic-training|训练脚本pretrain_gpt.py的启动参数| 弹性训练功能开关，默认关闭。                                                                                                                                                                                                                                                                       |
|max_restarts|启动训练的shell脚本（例如train_start.sh）中| 配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。                                                                                                                                                                                                                     |
|monitor_interval|启动训练的shell脚本（例如train_start.sh）中| 配置监测训练进程状态的时间间隔，单位为秒，取值为整数。不配置该参数时默认为5秒。                                                                                                                                                                                                                                             |
|HIGH_AVAILABILITY|Ascend Operator注入容器的环境变量中| Ascend Operator根据任务类型自动注入该环境变量，使用2.3.0版本MindSpeed-LLM会自动读取该环境变量，无需在train_start.sh中手动添加--enable-high-availability、--enable-hbmfault-repair、--enable-worker-reboot和--enable-elastic-training参数开启对应功能。                                                                                  |

**表 3** Ascend Operator注入的环境变量

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
