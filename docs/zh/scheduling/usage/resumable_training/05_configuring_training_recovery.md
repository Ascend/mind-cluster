# 配置恢复加速<a name="ZH-CN_TOPIC_0000002479386506"></a>

## 配置周期性CKPT保存<a name="ZH-CN_TOPIC_0000002479226552"></a>

本章节将指导用户了解周期性CKPT保存的关键步骤。周期性CKPT保存的特性介绍请参见[周期性CKPT保存](./01_solutions_principles.md#周期性ckpt保存)。

**配置存储CKPT加载<a name="zh-cn_topic_0000002111866386_section1296017551704"></a>**

从存储加载CKPT可基于AI框架提供的加载接口进行加载，用户需要传入需要加载的文件路径到AI框架中。以MindSpeed-LLM框架为例，用户如果需要配置从存储加载CKPT功能，可参考以下示例。

在任务YAML中，新增--load /data/ckpt/XXX \参数，开启存储CKPT加载。其中“--load”是训练进程恢复的统一开关，打开后训练进程恢复才生效。

```Yaml
...
spec:
  replicaSpecs:
    Master:
      template:
        spec:
          containers:
          - name: ascend # do not modify
            args:
              - | 
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                  --load /data/ckpt/XXX \  # 填写CKPT所在存储的路径   
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
                --load /data/ckpt/XXX \    # 填写CKPT所在存储的路径   
                  ...
...
```

## 配置临终CKPT保存<a name="ZH-CN_TOPIC_0000002479226544"></a>

本章节将指导用户了解临终CKPT保存的关键步骤。临终CKPT保存的特性介绍请参见[临终CKPT保存](./01_solutions_principles.md#临终ckpt保存)。

**构建镜像<a name="zh-cn_topic_0000002112026142_section26738428458"></a>**

使用Dockerfile构建容器镜像，新增启动命令。

```shell
... 
# MindCluster无损失断点续训适配脚本
RUN pip3 install $TASKD_WHL 
RUN pip3 install $MINDIO_TTP_PKG 

# 可选，使用优雅容错、Pod级别重调度或进程级别重调度时必须配置以下命令
RUN sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
```

**准备任务YAML<a name="zh-cn_topic_0000002112026142_section2671124124612"></a>**

在训练任务YAML中，新增以下字段，开启进程级别恢复。recover-strategy是训练进程恢复使用的策略，其中的dump代表开启临终CKPT。在ports中增加ttp-port为8000，增加TaskD通信使用的端口9601。

临终CKPT保存可以作为进程级别恢复流程中的一个策略，名为“dump”策略，设置到recover-strategy中。示例如下。

<pre codetype="yaml">
... 
metadata:  
   labels:  
     ...  
 ... 
...  
   annotations:  
     ...  
     <strong>recover-strategy: "dump"       # 任务可用恢复策略为保存临终遗言</strong>
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

**适配训练脚本<a name="zh-cn_topic_0000002112026142_section058501610462"></a>**

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

    2. 在训练脚本中增加以下代码拉起TaskD Manager。

        ```shell
        export TASKD_PROCESS_ENABLE="on" 
         
        # PyTorch框架下
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP} 
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
        fi 
              
        torchrun ...
        ```

2. 在启动脚本（例如train\_start.sh）中，新增--max\_restarts参数，示例如下。

    <pre codetype="shell">
    ... 
       logger "server id is: ""${server_id}" 
       if [ "${framework}" == "PyTorch" ]; then 
         get_env_for_pytorch_multi_node_job 
         <strong>DISTRIBUTED_ARGS</strong>="--nproc_per_node $GPUS_PER_NODE --nnodes $NNODES --node_rank $NODE_RANK --master_addr $MASTER_ADDR --master_port $MASTER_PORT  <strong>--max_restarts 32767</strong>" 
     ...</pre>

      其中，--max_restarts表示配置容器内最大允许触发的故障次数，取值为整数。超出次数后PyTorch训练进程会直接退出训练，不配置该参数时默认为32767次。

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

## 配置参数面传参恢复<a name="ZH-CN_TOPIC_0000002479386502"></a>

当前仅支持在进程级别重调度和进程级在线恢复特性中使用该能力，按照[配置进程级别重调度](./04_configuring_fault_handling_policies.md#配置进程级别重调度)和[配置进程级在线恢复](./04_configuring_fault_handling_policies.md#配置进程级在线恢复)特性适配后默认开启该能力。

**（可选）关闭参数面传参恢复<a name="zh-cn_topic_0000002181310402_section199132050405"></a>**

在进程级别重调度和进程级在线恢复特性中，如果用户想要关闭该功能，修改为从存储CKPT加载参数恢复，需修改任务YAML。以使用进程级重调度且关闭参数面传参恢复为例，示例如下。

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
     <strong>recover-strategy: "recover"   # 任务可用恢复策略为进程级别重调度</strong> 
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
          - name: ascend # do not modify
            ...
            args:
              - |
                ...
                bash scripts/train_start.sh /job/code /job/output pretrain_gpt.py \
                  ...
                  <strong>--distributed-optimizer-no-replica \</strong>
                  ...
...</pre>

**distributed-optimizer-no-replica**：数据修复支持周期CKPT功能开关，默认关闭，配置后副本优化器无副本，减小内存占用，在进程级别重调度和进程级在线恢复场景下，使用周期CKPT进行修复。本开关需开启进程级别重调度或进程级在线恢复。

## 配置集成时间优化方案<a name="ZH-CN_TOPIC_0000002479386526"></a>

### 恢复时间优化（PyTorch）<a name="ZH-CN_TOPIC_0000002479386516"></a>

本章节介绍在PyTorch框架上使用断点续训特性时，用户可以选择使用的缩短断点续训时间的相关功能，包括[故障检测时间优化](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section195202179141)、[集合通信初始化时间优化](#zh-cn_topic_0000002163883997_section725312412292)、[训练回滚及加载Checkpoint时间优化](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section71731855173720)和[算子编译时间优化](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section14599171031417)。

**故障检测时间优化<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section195202179141"></a>**

由于集群中出现的参数面网络故障不一定会影响训练任务，因此集群调度组件不会强制中断任务；当参数面网络故障影响训练任务时，会触发集合通信的网络超时等待机制，在等待时间（默认为30分钟）后，集群调度组件才能感知到该故障，从而触发断点续训。针对该问题，PyTorch  Adapter插件（torch\_npu）提供**watchdog故障检测**功能，可用于检测训练任务是否受到影响，缩短故障检测时间，该功能的详细说明请参见[表1](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table4822175901415)。

**表 1** watchdog故障检测功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table4822175901415"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row9823145931412"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p188231359141419"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p128231659131413"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12943926103611"></a>watchdog</span>故障检测。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row58231859181412"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1882355910149"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p98238590143"></a>训练启动时，同时启动一个监测线程不断获取通信异常以及task执行异常。监测到故障发生后，快速抛出异常并终止训练任务进程，触发重调度流程。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row138235598144"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15823155941416"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1982365912149"></a>仅支持<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1810104910187"></a>PyTorch</span> 1.11.0、2.1.0及以上版本；<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1758915488355"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC1。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row11823195941410"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p17823959121418"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p132201841114917"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14998597340"></a>PyTorch</span> 2.1.0及以上版本默认开启<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph6991859163419"></a>watchdog</span>故障检测，<strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b28088598507"></a>无需手动配置环境变量</strong>。</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1099185913342"></a>（可选）如需关闭<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph355416217352"></a>watchdog</span>故障检测，需在训练的shell启动脚本（例如train_start.sh）中，修改以下环境变量。</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen129905913414"></a>...
# env for breakpoint ckpt
export RESUME_MODE_ENABLE=1
<br>
export HCCL_ASYNC_ERROR_HANDLING=0  <strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b177584103519"></a>          </strong># 该环境变量的详细说明请参见<a href="../../api/environment_variable_description.md#taskd环境变量说明">TaskD环境变量说明</a></pre>
</td>
</tr>
</tbody>
</table>

**集合通信初始化时间优化<a name="zh-cn_topic_0000002163883997_section725312412292"></a>**

Parallel Store多线程建链优化：PyTorch框架创建通信组时，使用TCP Store进行信息交换。随着任务规模变大会影响原生TCP Store的信息处理性能，导致创建通信组时间过长。针对该问题，PyTorch Adapter插件支持使用原生TCP Store的优化版本Parallel Store，详细说明请参见[表2](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table14133757143220)。

**表 2**  Parallel Store功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table14133757143220"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row15133115723218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p191333574329"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16133257163210"></a>Parallel Store</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row2013316574328"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p71332057113215"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31331957193214"></a>多线程处理建链请求，减少建链请求队列等待时间，降低总体建链时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1913318574324"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133175711322"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p5133165713217"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1493841033420"></a>PyTorch</span> 1.11.0版本：<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph14923134593417"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC1。</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1599664014012"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1199694044010"></a>PyTorch</span> 2.1.0及以上版本：<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13996140164020"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC3。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16133957183217"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p15134175723214"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p3415192817368"></a>将启动训练的shell脚本（例如train_start.sh）中，torchrun启动命令修改为torch_npu_run。</p>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p150716434420"></a>比如将</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1330411394310"></a><strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b1230514391831"></a>torchrun train.py --train_parameter=xxx ....</strong></pre>
<p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2732193816313"></a>修改为</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen974017323469"></a><strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b692841611410"></a>torch_npu_run train.py --train_parameter=xxx ....</strong></pre>
</td>
</tr>
</tbody>
</table>

- 原生HCCL建链性能优化：PyTorch框架在NPU侧交换集合通信信息后进行NPU间连接建链。随任务规模变大，导致建链时间大幅度增加。针对该问题，CANN对原生HCCL建链进行了性能优化，详细说明请参见[表3](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table10637950133911)。

    **表 3**  原生HCCL建链性能优化功能说明

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table10637950133911"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row763710506398"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p163710508395"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2637135011397"></a>原生HCCL建链性能优化。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row963765019395"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p16638195017394"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p62631347145019"></a>多线程异步完成集合通信信息协商，减少通信信息协商时间，降低总体建链时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row263845043913"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1963825014391"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1638950113918"></a>仅支持CANN 8.0.RC2及以上版本。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1563845013912"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763816507391"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1321615531212"></a>无。</p>
    </td>
    </tr>
    </tbody>
    </table>

- RankTable模式建链优化：集群调度Ascend Operator组件为PyTorch框架提供生成集合通信配置文件（RankTable File，也叫hccl.json文件）功能，可以通过RankTable模式建链，缩短集群通信建链时间，详细说明请参见[表4](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1749892464019)。

    **表 4**  集合通信使用RankTable模式建链

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1749892464019"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row84981324184016"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p24982024194017"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p124981924164011"></a>RankTable模式建链。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16498162484017"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p849802464013"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p20499624194016"></a>使用<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph173171749164715"></a>Ascend Operator</span>为PyTorch任务生成集合通信配置文件，缩短集群通信建链时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row2499424194015"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p2499182413406"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p44991248407"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph494239144114"></a>PyTorch</span> Adapter插件（torch_npu）版本必须高于6.0.RC3。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row4499524124018"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8499122454018"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"></a><ol id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol83564504415"><li>启动YAML中已经默认挂载了hccl.json文件的父目录，用户可以根据实际情况进行修改。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen188130104213"></a>volumes:
           - name: ranktable-dir
             hostPath:
               path: /user/mindx-dl/ranktable  # 该宿主机目录需要在共享目录下
               type: DirectoryOrCreate</pre>
    <div class="p" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p982140154213"></a>执行以下命令，在宿主机目录下创建hccl.json文件的具体挂载路径，并修改所属用户。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen148290174212"></a>mkdir -m 777 /user/mindx-dl/ranktable/任务运行的命名空间.任务名称
    chown 9000:9000 /user/mindx-dl/ranktable/default.pytorch-test</pre>
    </div>
    <div class="p" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p4827024210"></a>例如：<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1382306426"></a>mkdir -m 777 /user/mindx-dl/ranktable/default.pytorch-test
    chown 9000:9000 /user/mindx-dl/ranktable/default.pytorch-test</pre>
    </div>
    </li><li>修改训练脚本，添加如下环境变量。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1782170124216"></a>export RANK_TABLE_FILE=/user/mindx-dl/ranktable/hccl.json</pre>
    </li><li>修改训练YAML，添加如下设置。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen19101124310423"></a>yaml
          volumeMounts:
          - name: ranktable
            mountPath: /user/mindx-dl/ranktable

           volumes:
           - name: ranktable
             hostPath:
               path: /user/mindx-dl/ranktable/任务运行的命名空间.任务名称  # 宿主机目录下hccl.json文件的实际路径
    </pre>
    </li></ol>
    </td>
    </tr>
    </tbody>
    </table>

**训练回滚及加载Checkpoint时间优化<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section71731855173720"></a>**

- 异步保存Checkpoint：训练任务会定期保存Checkpoint文件，用于保存参数信息，故障恢复需要从上一次保存的Checkpoint回滚恢复训练。由于每次保存Checkpoint文件均会浪费一定的训练时间，为了保证训练效率，保存Checkpoint的时间间隔通常较大，而保存间隔越大，每次故障时训练回滚浪费的时间就会越长。针对该问题，集群调度组件支持通过MindIO ACP异步保存Checkpoint，详细说明请参见[表5](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table5173115519372)。

    **表 5**  异步保存Checkpoint功能说明

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table5173115519372"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row717435514373"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p31749558370"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1417445523712"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph197701281571"></a>MindIO ACP</span>异步保存Checkpoint。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row10174115583714"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174105513720"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1017415503717"></a>从NPU中获取Checkpoint后，异步写入存储中，降低每次保存Checkpoint的训练损失和保存周期，从而降低训练回滚时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row6174655153715"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1174115513711"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6174145503713"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph620813251731"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row171741155133719"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p131751155143719"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14233123161113"></a>安装和使用<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph7901201365813"></a>MindIO</span>组件，请参见<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph12306119151316"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Checkpoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

- 高效恢复Checkpoint：回滚恢复训练时，通常需要从存储中加载保存的Checkpoint，由于Checkpoint数据量较大，直接从存储读取加载Checkpoint的耗时较长。针对该问题，集群调度组件支持通过MindIO ACP进行Checkpoint高效恢复，详细说明请参见[表6](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1163115196618)。

    **表 6**  Checkpoint高效恢复功能说明

    <a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table1163115196618"></a>
    <table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row763114191366"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763113190615"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8488163331214"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph10665545175814"></a>MindIO</span> Checkpoint高效恢复。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row14631191914615"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p963119191369"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p206311198612"></a><span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph185951849175817"></a>MindIO</span>将最新的Checkpoint存储到内存中，故障恢复时可直接从内存中读取Checkpoint，降低Checkpoint读取时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row26321219766"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10632619164"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1423217297012"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph13333125495813"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row9632219868"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1763218197619"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1223164710116"></a>安装和使用<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph648515775810"></a>MindIO</span>组件，请参见<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph758864121412"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Checkpoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

**算子编译时间优化<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_section14599171031417"></a>**

断点续训过程中拉起训练需要重新执行算子时，算子编译需要消耗大量时间。针对该问题，可选择算子二进制或算子编译缓存降低编译时间，详细说明请参见[表7](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table8599191019143)和[表8](#zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table2193759172110)。

>[!NOTE] 
>算子二进制和算子编译缓存二者不兼容，请选择其中之一进行使用。

**表 7**  算子二进制功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table8599191019143"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1599111016145"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p459951061414"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p1859961011147"></a>使用算子二进制。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1059931012143"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p135991910151414"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p8599110181417"></a>算子编译时提前加载预置的算子二进制，直接免编译执行算子。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row16599161015147"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p86003102149"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p7600191071419"></a>仅支持CANN 8.0.RC2及以上版本。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row7600610181419"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p106006109149"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p118736402313"></a>在<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph1541893220206"></a>Python</span>启动脚本中，添加算子二进制配置命令，开启算子二进制。</p>
<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen11271685535"></a>torch.npu.set_compile_mode(jit_compile=False)</pre>
</td>
</tr>
</tbody>
</table>

**表 8**  算子编译缓存功能说明

<a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_table2193759172110"></a>
<table><tbody><tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1819335920218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p141931659192119"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p01931659192118"></a>算子编译缓存。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row11193185913215"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p181935592211"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p18193759202120"></a>算子编译时加载存储中保存的算子编译缓存文件，加载后可降低编译时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1719310593218"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p14193125917212"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p6193165952117"></a>仅支持CANN 8.0.RC2及以上版本。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_row1193195962112"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_p10194105932115"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"></a><ol id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ol9110131582413"><li>在<span id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_ph17194359132118"></a>Python</span>启动脚本中，添加算子编译缓存配置命令，开启算子编译缓存。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen5194145992110"></a>torch.npu.set_compile_mode(jit_compile=True)</pre>
</li><li>在训练的shell启动脚本中（例如train_start.sh），添加如下环境变量。<pre class="screen" id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_screen1495913117539"></a>export ASCEND_CACHE_PATH<strong id="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"></a><a name="zh-cn_topic_0000002163883997_zh-cn_topic_0000002017918296_b861203214264"></a>=xxx</strong>   # 添加共享存储路径
export ASCEND_MAX_OP_CACHE_SIZE=-1    # 使用共享存储时建议开启，可解决多节点读取共享存储缓存资源争抢严重问题</pre>
</li></ol>
</td>
</tr>
</tbody>
</table>

### 恢复时间优化（MindSpore）<a name="ZH-CN_TOPIC_0000002511346499"></a>

断点续训特性在使用MindSpore框架场景时，可以使用以下功能，缩短断点续训整体恢复时间，包括[故障检测时间优化](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section517194154019)、[训练回滚及加载Checkpoint时间优化](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section2743164217401)和[编译缓存时间优化](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section1139444324019)。

**故障检测时间优化<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section517194154019"></a>**

由于集群中出现的参数面网络故障不一定会影响训练任务，因此集群调度组件不会强制中断任务；当参数面网络故障影响训练任务时，会触发集合通信的网络超时等待机制，在等待时间（默认为30分钟）后，集群调度组件才能感知到该故障，从而触发断点续训。针对该问题，MindSpore提供**watchdog故障检测**功能，可用于检测训练任务是否受到影响，缩短故障检测时间，该功能的详细说明请参见[表1](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table17897155873217)。

**表 1** watchdog故障检测功能说明

<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table17897155873217"></a>
<table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row289715810326"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p15897135815323"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p148971958143217"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph389717582323"></a>watchdog</span>故障检测。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1589716585326"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1689713588324"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9897658123215"></a>训练启动时，同时启动一个监测线程不断获取通信异常以及task执行异常。监测到故障发生后，快速抛出异常并终止训练任务进程，触发重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row189775853220"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7897105873219"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p789713583324"></a>仅支持MindSpore 2.4版本以上</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row5898058143215"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p19898458173211"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6898145818329"></a>MindSpore<strong id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b171861451155620"></a>默认开启</strong><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph289835820329"></a>watchdog</span>故障检测，<strong id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_b2843133219564"></a>无需手动配置</strong>。如果需要关闭<span id="ph1052517411176"><a name="ph1052517411176"></a><a name="ph1052517411176"></a>watchdog</span>故障检测，请在模型配置文件中新增如下加粗字段。</p>
<pre class="screen" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen1898205823219"></a>...
context:
  <strong id="b393317297113"><a name="b393317297113"></a><a name="b393317297113"></a>ascend_config:</strong>
    <strong id="b12660461696"><a name="b12660461696"></a><a name="b12660461696"></a>hccl_watchdog: False</strong>    
...</pre>
</td>
</tr>
</tbody>
</table>

**训练回滚及加载Checkpoint时间优化<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section2743164217401"></a>**

- 异步保存Checkpoint：训练任务会定期保存Checkpoint文件，用于保存参数信息，故障恢复需要从上一次保存的Checkpoint回滚恢复训练。由于每次保存Checkpoint文件均会浪费一定的训练时间，为了保证训练效率，保存Checkpoint的时间间隔通常较大，而保存间隔越大，每次故障时训练回滚浪费的时间就会越长。针对该问题，集群调度组件支持通过MindIO ACP异步保存Checkpoint，详细说明请参见[表2](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table56063271212)。

    **表 2**  异步保存Checkpoint功能说明

    <a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table56063271212"></a>
    <table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row4606162713214"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1960610272217"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76068273210"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph036712110595"></a>MindIO ACP</span>异步保存Checkpoint。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row260619272216"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p12606112716213"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p146061227102118"></a>从NPU中获取Checkpoint后，异步写入存储中，降低每次保存Checkpoint的训练损失和保存周期，从而降低训练回滚时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row126061827152113"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2060672782119"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1460622714218"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph620813251731"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row136069278219"><th class="firstcol" valign="top" width="19.98%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p5606527102113"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80.02%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18606182742112"></a>安装和使用<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph9523194885915"></a>MindIO</span>组件，请参见<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph12306119151316"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Checkpoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

- 高效恢复Checkpoint：回滚恢复训练时，通常需要从存储中加载保存的Checkpoint，由于Checkpoint数据量较大，直接从存储读取加载Checkpoint的耗时较长。针对该问题，集群调度组件支持通过MindIO ACP进行Checkpoint高效恢复，详细说明请参见[表3](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table66066274216)。

    **表 3**  Checkpoint高效恢复功能说明

    <a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table66066274216"></a>
    <table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row106071271216"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6607327132112"></a>功能名称</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p76071727152117"></a><span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph141313012210"></a>MindIO</span> Checkpoint高效恢复。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row360715276216"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p760712772111"></a>功能特点</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p9607627182111"></a>将最新的Checkpoint存储到内存中，故障恢复时可直接从内存中读取Checkpoint，降低Checkpoint读取时间。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1860772716217"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1760715273215"></a>使用说明</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p2336153411594"></a>仅支持6.0.RC2及以上版本的集群调度组件和<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph14507135416591"></a>MindIO</span>组件。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row196071127102110"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p7607327172119"></a>关键操作</p>
    </th>
    <td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p18336173418590"></a>安装和使用<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph19467185545916"></a>MindIO</span>组件，请参见<span id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_ph235818092220"></a><a href="../../optimizing_saving_and_loading_checkpoints/01_product_description.md">Checkpoint保存与加载优化</a></span>。</p>
    </td>
    </tr>
    </tbody>
    </table>

**编译缓存时间优化<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_section1139444324019"></a>**

断点续训过程中拉起训练时需要构建计算图，在大模型场景下，构建计算图并编译需要消耗大量时间。针对该问题，MindSpore支持在首次编译时将编译缓存文件进行存储，进行故障恢复时可以直接读取存储中的图编译缓存，降低图编译时间，详细说明请参见[表4](#zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table175224282139)。

**表 4**  图编译缓存功能说明

<a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_table175224282139"></a>
<table><tbody><tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row135238284132"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p1952352818133"></a>功能名称</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p6523828191317"></a>图编译缓存。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row1052322818133"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.2.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p652313283132"></a>功能特点</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.2.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p0523328101313"></a>图编译时加载存储中保存的图编译缓存文件，加载后可降低图编译时间。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row5523628191316"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.3.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p125231028151311"></a>使用说明</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.3.1 "><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p252318280136"></a>仅支持MindSpore2.3.0及以上版本。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_row352313282136"><th class="firstcol" valign="top" width="20%" id="mcps1.2.3.4.1"><p id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p852314283139"></a>关键操作</p>
</th>
<td class="cellrowborder" valign="top" width="80%" headers="mcps1.2.3.4.1 "><div class="p" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_p111971223112917"></a>在训练的shell启动脚本中（例如train_start.sh），添加如下环境变量。<pre class="screen" id="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"></a><a name="zh-cn_topic_0000002128524426_zh-cn_topic_0000002053878705_screen115231428101313"></a>export MS_COMPILER_CACHE_ENABLE=1  # 开启图编译缓存
export MS_COMPILER_CACHE_PATH=xxx  # 设置图编译缓存路径</pre>
</div>
</td>
</tr>
</tbody>
</table>

## 配置HCCL主动触发建链<a name="ZH-CN_TOPIC_0000002511346489"></a>

当故障发生在HCCL建链阶段时，会导致进程级别重调度或进程级别在线恢复失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考以下步骤进行提前建链，防止故障出现在HCCL建链阶段。

**PyTorch单算子场景<a name="section145466566911"></a>**

PyTorch单算子场景HCCL建链为懒加载模式，当建立Torch通信组后，该通信组下发的第一个算子将触发HCCL通信域的创建，创建后完成卡间建链。因此，如果需要在训练初始化阶段完成所有通信域的建链，只需要在初始化阶段给每个通信组下发一个通信算子。

以下为创建通信组主动创建的示例：

```Python
rank = 0 # 设置本进程rank
sub_ranks = [0, 1, 2]  # 假设为一个包含0、1、2的通信组
groupX = torch.distributed.new_group(ranks=sub_ranks,...) # 创建通信组X
test_tensor = torch.ones(1).to(f'npu:{rank}') * (rank + 1)  # 构建一个测试数据tensor
torch.distributed.all_reduce(test_tensor, op=dist.ReduceOp.SUM, group=groupX)  # 在通信组X执行all reduce算子
```
