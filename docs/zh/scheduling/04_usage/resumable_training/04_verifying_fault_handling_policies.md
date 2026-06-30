# 验证故障处理

## 验证Job级别重调度

**前提条件**

在基础调度的任务YAML中，添加Job级别重调度的配置，配置说明可参考[配置Job级别重调度](../configuration/02_configuring_fault_handling_policies.md#zh-cn_topic_0000002098814658_section463203519254)，原理可参考[Job级别重调度](../01_solutions_principles.md#ZH-CN_TOPIC_0000002479226586)。

**操作步骤**

1. 下发任务

   执行以下命令下发任务：

   ```bash
   kubectl apply -f trjob.yaml
   ```
   >[!NOTE]
   > - 请将`trjob.yaml`替换为实际的任务YAML文件。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的`taskmgr-npu-020-default-test-`和`trjob`都是示例值，实际值会根据任务YAML中的配置而变化。

2. 查看任务状态和UID

   1. 执行以下命令查看任务状态：

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，出现Running表示任务正常运行：

      <pre codetype="bash">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>    0          2s     xx.xx.xx.xx      node173                 <none>           <none>
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>    0          3s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
      </pre>


   2. 执行以下命令查看2个Pod的UID：
      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下：
      ```bash
      7286faf8-f029-450a-b302-5e6e94d4346c
      997add9e-6115-456c-9e8e-e05e4b70bb12
      ```

3. 构造故障

   执行以下命令查询任务进程：

   ```bash
   npu-smi info|grep python|awk '{print $5}'
   ```

   回显示例如下：

   ```bash
   2398104
   2398105
   2398107
   ```

   执行以下命令将进程终止模拟故障：

   ```bash
   kill -9 2398104
   ```

4. 观察重调度过程

   执行以下命令监控该Job的2个Pod状态变化：

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   该Job的2个Pod历史状态如下，观察加粗字段的变化可以发现该Job的2个Pod会经历Terminating→Pending→ContainerCreating→Running阶段，然后正常运行，表示Job重调度成功：

   <pre codetype="bash">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running             0          2s      xx.xx.xx.xx       node173                 <none>           <none>
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running             0          3s      xx.xx.xx.xx       localhost.localdomain   <none>           <none>
   // ===================== 注入故障 ======================
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          0s      <none>            <none>                  <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          1s      <none>            <none>                  <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          73s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          85s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          85s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          0s      <none>            <none>                  <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          1s      <none>                 localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     <none>                 node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     <none>                 node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          1s      <none>                 localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0          1s      xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0          1s      xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Running</strong>             0          43s     xx.xx.xx.xx      node173                 <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Running</strong>             0          2s      xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   </pre>

5. 查看任务状态和UID

   1. 执行以下命令查看任务状态：

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，出现Running表示任务正常运行：

      <pre codetype="bash">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          2s      xx.xx.xx.xx      node173   <none>           <none>
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          33s     xx.xx.xx.xx      node173   <none>           <none>
      </pre>

   2. 执行以下命令查看2个Pod的UID：
      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下，该Job的2个Pod的UID均发生变化，说明2个Pod都经历了重调度，即触发Job级别重调度：

      ```bash
      2a24eee8-88f1-4107-bc9d-dabcfb09dea9
      074f9f9c-35f1-4b9e-9298-5b2bcf3759e7
      ```

## 验证Pod级别重调度

**前提条件**

在基础调度的任务YAML中，添加Pod级别重调度的配置，配置说明可参考[配置Pod级别重调度](../configuration/02_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002479226508)，原理可参考[Pod级别重调度](../01_solutions_principles.md#ZH-CN_TOPIC_0000002511346429)。

**操作步骤**

1. 下发任务

   执行以下命令下发任务：

   ```bash
   kubectl apply -f trjob.yaml
   ```
   >[!NOTE]
   > - 请将`trjob.yaml`替换为实际的任务YAML文件。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的`taskmgr-npu-020-default-test-`和`trjob`都是示例值，实际值会根据任务YAML中的配置而变化。

2. 查看任务状态和UID

   1. 执行以下命令查看任务状态：

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，出现Running表示任务正常运行：

      <pre codetype="bash">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      node173                 <none>           <none>
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      localhost.localdomain   <none>           <none>
      </pre>

   2. 执行以下命令查看2个Pod的UID：

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下：

      ```bash
      de1f8848-ed88-4e18-abda-7abc8dbede87
      47291595-85b0-47ff-8393-c922d0e2dfb2
      ```

3. 构造故障

   执行以下命令查询任务进程：

   ```bash
   npu-smi info|grep python|awk '{print $5}'
   ```

   回显示例如下：

   ```bash
   2398132
   2398144
   2398158
   ```

   执行以下命令将进程终止模拟故障：

   ```bash
   kill -9 2398144
   ```

4. 观察重调度过程

   执行以下命令监控该Job的2个Pod状态变化：

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   该Job的2个Pod历史状态如下，观察加粗字段的变化可以发现故障Pod（taskmgr-npu-020-default-test-1）会经历Error→Terminating→Pending→ContainerCreating→Running阶段，然后正常运行，表示Pod重调度成功：

   <pre codetype="bash">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running              0          6s      xx.xx.xx.xx      node173                 <none>           <none>
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running              0          6s      xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   // ===================== 注入故障 ======================
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Error</strong>               0          34s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          35s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          35s     xx.xx.xx.xx      localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           0s      <none>            <none>                  <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           1s      <none>                localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           1s      <none>                localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0           1s      xx.xx.xx.xx     localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0           1s      xx.xx.xx.xx     localhost.localdomain   <none>           <none>
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Running</strong>             0           2s      xx.xx.xx.xx     localhost.localdomain   <none>           <none>
   </pre>

5. 查看任务状态和UID

   1. 执行以下命令查看任务状态：

      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下，出现Running表示任务正常运行：

      <pre codetype="bash">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          66s      xx.xx.xx.xx      node173                 <none>           <none>
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          31s      xx.xx.xx.xx      localhost.localdomain   <none>           <none>
      </pre>

   2. 执行以下命令再次查看2个Pod的UID：

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      回显示例如下，taskmgr-npu-020-default-test-0 Pod的UID未发生变化，taskmgr-npu-020-default-test-1 Pod的UID发生变化，说明只有发生故障的Pod（taskmgr-npu-020-default-test-1）经历了重调度，即触发Pod级别重调度：

      ```bash
      de1f8848-ed88-4e18-abda-7abc8dbede87
      6eb3c217-3b63-457a-9010-9d236d281634
      ```

## 验证进程级别重调度

**前提条件**

在基础调度的任务YAML中，添加进程级别重调度的配置，配置说明可参考[配置进程级别重调度](../configuration/02_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002511426407)，原理可参考[进程级别重调度](../01_solutions_principles.md#ZH-CN_TOPIC_0000002511346457)。

**操作步骤**

1. 下发任务

   执行以下命令下发任务：

   ```bash
   kubectl apply -f trjob.yaml
   ```
   >[!NOTE]
   > - 请将`trjob.yaml`替换为实际的任务YAML文件。
   > - 任务Pod的名称、命名空间会根据任务YAML中的配置而变化，以下出现的`process-reschedule-function-`和`trjob`都是示例值，实际值会根据任务YAML中的配置而变化。

2. 查看任务状态

   执行以下命令查看任务状态：

   ```bash
   kubectl get pod -A -o wide
   ```

   回显示例如下，出现Running表示任务正常运行：
   <pre codetype="bash">
   trjob            process-reschedule-function-master-0   1/1     Running   0               14s   xx.xx.xx.xx     master-69-117   <none>           <none>
   trjob            process-reschedule-function-worker-0   1/1     Running   0               14s   xx.xx.xx.xx     work-69-115     <none>           <none>
   </pre>


4. 查看训练日志迭代步数

   执行以下命令查看迭代步数，确认训练已正常迭代：

   ```bash
   kubectl logs -n trjob process-reschedule-function-worker-0|grep -Po '] iteration [[:space:]]*'|wc -l
   ```

   回显示例如下：
   ```bash
   50
   ```

5. 查看进程ID，并构造故障

   执行以下命令查看进程ID：
   ```bash
   npu-smi info|grep python|awk '{print $5}'
   ```

   回显示例如下：

   ```bash
   635755
   635756
   635760
   635770
   635777
   635784
   635791
   635795
   ```

   终止其中一个进程来模拟故障发生：

   ```bash
   kill -9 635777
   ```

6. 观察训练日志

   执行以下命令监控训练日志：

   ```bash
   kubectl logs -n trjob process-reschedule-function-master-0
   ```

   回显示例如下：

   ```bash
   # 出现以下信息说明开始触发ARF流程
   Mindx calling notify do ARF repair

   # 出现以下信息说明ARF成功
   ... Mindio do repair operation ok ...
   ```

7. 观察job-reschedule-reason内容是否准确

   执行以下命令查看ConfigMap job-reschedule-reason中是否有任务信息：
   ```bash
   kubectl describe cm -n mindx-dl job-reschedule-reason |grep process-reschedule-function
   ```

   回显示例如下，其中包含重调度的时间，触发重调度的pod、node、rank，本任务当前重调度次数等信息：
   ```bash
   {"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07":{"JobID":"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0603 05:16:52","RescheduleTimeStamp":1780435012,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"process-reschedule-function-worker-0","NodeName":"work-69-115","NodeRankIndex":"1"}]}]}}
   ```

## 验证进程级别在线恢复

本章节通过在训练代码中打桩构造片上内存的UCE故障，指导用户完成进程级在线恢复验证的适配步骤。

>[!NOTE]
>
>- 本章节相关修改仅用于指导用户在测试环境下验证进程级在线恢复功能，切勿将此打桩版本上线到生产环境。
>- 配置本章节步骤前，请确保训练能正常拉起并已配置进程级在线恢复。
>- 为保证进程级在线恢复功能的正常使用，请将K8s集群master节点与worker节点的时钟保持一致。
>- 下文中代码可能与实际版本存在差异，请以实际版本代码为准。

### MindCluster适配<a name="ZH-CN_TOPIC_0000002479386410"></a>

1. <a name="li977718409381"></a>拉取MindCluster代码。

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/mind-cluster.git
    cd ./mind-cluster/component/clusterd
    git checkout v26.0.0   # v26.0.0是代码仓版本tag，请自行切换到目标版本
    ```

2. 修改ClusterD代码。
   1. 打开“pkg/application/faultmanager/jobprocess/faultrank/job\_fault\_rank\_processor.go”文件。

      ```shell
      vi pkg/application/faultmanager/jobprocess/faultrank/job_fault_rank_processor.go
      ```

   2. 按“i”进入编辑模式，添加如下加粗代码。

      <pre codetype="go">
         package faultrank

         import (
         …
            <strong>"clusterd/pkg/domain/faultdomain/collector"</strong>
         …
         )
         …
         func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(
         …
               if deviceDetail, ok := processor.retryInBusinessPlane(podInfo.jobId, nodeName, deviceName); ok {
                  faultRankList = append(faultRankList, constant.FaultRank{RankId: deviceInfo.RankID, PodUid: podUid,
                     PodRank: podRankStr, FaultCode: faultdomain.GetRetryCodeByFaultType(deviceDetail.FaultType),
                     FaultLevel:  constant.RestartBusiness,
                     DoStepRetry: processor.canDoStepRetry(podInfo.jobId, nodeName, deviceName),
                     DeviceId:    deviceInfo.DeviceID,
               })
               <strong>collector.ReportInfoCollector.ReportRetryInfo(podInfo.jobId, deviceInfo.RankID, constant.JobNotRecover, constant.UceFaultType)   // 业务面故障时间设置为无效时间，避免单次故障重复触发进程级在线恢复</strong>
            }
        …
      </pre>

   3. 按“Esc”键，输入:wq!，按“Enter”保存并退出编辑。

3. <a name="li114977117517"></a>编译ClusterD。

   ```shell
   cd ./build/
   chmod +x build.sh && dos2unix build.sh
   sed -i 's|build_version="v[^"]\+"|build_version="xxx"|g' build.sh  # xxx替换为版本号，如v26.0.0
   sed -i 's|export CGO_ENABLED=0|export CGO_ENABLED=1|g' build.sh  # 开启CGO功能
   ./build.sh # 编译ClusterD，需要go 1.26及以上版本，建议使用1.26版本
   ```

   编译成功后，会在“../output/”目录下生成相关文件，可执行如下命令进行查看：

   ```shell
   ll ../output/
   ```

   回显示例如下：

   ```bash
   -r-x------. 1 root root 45891128 Aug 13 10:52 clusterd
   -r--------. 1 root root     4021 Aug 13 10:52 clusterd-v26.0.0.yaml
   -r--------. 1 root root      946 Aug 13 10:52 Dockerfile
   -r--------. 1 root root      209 Aug 13 10:52 faultDuration.json
   -r--------. 1 root root      207 Aug 13 10:52 fdConfig.yaml
   -r--------. 1 root root      467 Aug 13 10:52 publicFaultConfiguration.json
   -r--------. 1 root root      756 Aug 13 10:52 relationFaultCustomization.json
   ```

4. <a name="li89701053589"></a>进入output目录，制作ClusterD镜像。

   ```shell
   cd ../output/
   docker build --no-cache -t clusterd:{tag} ./  # {tag}与步骤3中build_version="xxx"的取值保持一致
   ```

5. （可选）保存镜像，并将保存后的镜像文件和clusterd-\{tag\}.yaml文件上传到主节点。若[步骤1](#li977718409381)到[步骤4](#li89701053589)在主节点执行，可跳过该步骤。

   ```shell
   docker save -o clusterd.tar clusterd:{tag}  # 保存镜像
   docker load -i clusterd.tar  # 在主节点导入镜像
   ```

6. 在主节点重新拉起ClusterD。

   ```shell
   kubectl delete -f  clusterd-{tag}.yaml  # 删除旧ClusterD容器
   kubectl apply -f  clusterd-{tag}.yaml  # 拉起新容器
   ```

### 脚本适配<a name="ZH-CN_TOPIC_0000002479226412"></a>

#### PyTorch场景适配示例（基于MindSpeed-LLM）<a name="ZH-CN_TOPIC_0000002511426361"></a>

1. 搭建训练环境，拉起训练，详细请参见[PyTorch场景适配示例（基于MindSpeed-LLM）](../03_using_resumable_training_on_the_cli.md#适配示例)。
2. 开启进程级在线恢复，详细请参见[配置进程级在线恢复](../configuration/02_configuring_fault_handling_policies.md#配置进程级在线恢复)。
3. 在“QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py”代码中增加如下加粗内容，打桩注入故障，新增代码根据环境变量“RAISE\_UCE\_ERROR\_STEP\_AND\_RANK”获取注入故障迭代位置和故障rank信息。

   <pre codetype="Python">
      <strong>import os</strong>
      <strong>import ast</strong>
      <strong>…</strong>
      <strong>GLB_CNT = 0</strong>
      def train(forward_step_func, model, optimizer, opt_param_scheduler,
              train_data_iterator, valid_data_iterator,
              process_non_loss_data_func, config):
         """Train the model function."""
         args = get_args()
         timers = get_timers()
         …
         while iteration < args.train_iters:
            …
            num_microbatches = get_num_microbatches()
            update_num_microbatches(args.consumed_train_samples, consistency_check=True)
            <strong>global GLB_CNT</strong>
            <strong>cur_rank = torch.distributed.get_rank()</strong>
            <strong>uce_env = os.getenv("RAISE_UCE_ERROR_STEP_AND_RANK", "{}")</strong>
            <strong>uce_step_rank = ast.literal_eval(uce_env)</strong>
            <strong>if iteration in uce_step_rank and cur_rank == uce_step_rank[iteration] and GLB_CNT < iteration:</strong>
               <strong>GLB_CNT = iteration</strong>
               <strong>print(f"############# rank:{cur_rank} start UCE error #############")</strong>
               <strong>raise RuntimeError('UCE ERROR')</strong>
            args.curr_iteration = iteration
            …
   </pre>

4. 修改启动脚本“QWEN3\_for\_PyTorch\_2.7\_code/scripts/train\_start.sh”。

   ```shell
   …
   export RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"  # 配置故障注入的迭代和卡号，在第3个迭代的rank 8卡和第10个迭代的rank 9卡上注入UCE故障
   sed -i 's/check_memory_result = torch_npu.npu.check_uce_in_memory(device)/check_memory_result = ha_constant.UCE_HIGH_LEVEL/g' /job/code/mindspeed_llm/core/high_availability/tft_stop_clean.py #修改PTA接口返回值，将训练代码抛出的异常识别为UCE故障
   …
   ```

#### MindSpore场景适配示例（基于MindFormers）<a name="ZH-CN_TOPIC_0000002511346369"></a>

1. 搭建训练环境，拉起训练，详细请参见[MindSpore场景适配示例（基于MindFormers）](../03_using_resumable_training_on_the_cli.md#适配示例)。
2. 开启进程级在线恢复，详细请参见[配置进程级在线恢复](../configuration/02_configuring_fault_handling_policies.md#配置进程级在线恢复)。
3. 在“QWEN3\_for\_MS\_code/mindformers/core/callback/callback.py”代码中增加如下加粗内容，打桩注入故障。

   <pre codetype="Python">
      import json
      import os
      ...
      <strong>import ast</strong>
      <strong>GLB_CNT = 0</strong>
      <strong>EPOCH_CNT = 0</strong>
      ...
         def print_output_info(self, cb_params, cur_epoch_num, origin_epochs, throughput,
                              cur_step_num, steps_per_epoch, loss, per_step_seconds,
                              overflow, scaling_sens, time_remain, percent, global_norm):
            """print output information."""
            ...
            logger.info("  %4.1f%% %s %.5f samples/s/p  %s }", percent, show_str, throughput,
                        datetime.timedelta(seconds=int(time_remain)))
            <strong>global GLB_CNT</strong>
            <strong>global EPOCH_CNT</strong>
            <strong>if EPOCH_CNT < cur_epoch_num: </strong>
               <strong>GLB_CNT = 0</strong>
               <strong>EPOCH_CNT = cur_epoch_num</strong>
            <strong>uce_env = os.getenv("RAISE_UCE_ERROR_STEP_AND_RANK", "{}")</strong>
            <strong>uce_step_rank = ast.literal_eval(uce_env)</strong>
            <strong>if cur_step_num in uce_step_rank and get_rank() == uce_step_rank[cur_step_num] and GLB_CNT < cur_step_num: </strong>
               <strong>GLB_CNT = cur_step_num</strong>
               <strong>print(f"############# rank:{get_rank()} start UCE error #############")</strong>
               <strong>raise RuntimeError('UCEError occurred.')</strong>
            if self.tensor_writer is not None:
               ...
   </pre>

4. 修改启动脚本“QWEN3\_for\_MS\_code/scripts/msrun\_launcher.sh”。

   ```shell
   …
   export RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"  # 配置故障注入的迭代和卡号，在第3个迭代的rank 8卡和第10个迭代的rank 9卡上注入UCE故障
   sed -i 's/err_strategy = _get_uce_process_strategy()/err_strategy = "RS_UCE_LOWLEVEL"/g' $(pip3 show mindspore | grep Location | awk -F ' ' '{print $2}')/mindspore/train/callback/_train_fault_tolerance.py #修改UCE处理策略
   …
   ```

### 验证流程

以下示例基于**双机 16 卡**（单机 8 卡，Master rank 0–7、Worker rank 8–15）环境，与[脚本适配](#ZH-CN_TOPIC_0000002479226412)中 `RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"` 的配置一致。若使用单机或其他拓扑，请同步调整环境变量与下文 `grep` 中的 rank、Pod 名称。

**前提条件**

- 在基础调度的任务 YAML 中，添加进程级在线恢复的配置，配置说明可参考[配置进程级在线恢复](../configuration/02_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002479386492)，原理可参考[进程级在线恢复](../01_solutions_principles.md#ZH-CN_TOPIC_0000002479386460)。
- 已完成 MindCluster 适配和脚本适配；启动脚本中的 `RAISE_UCE_ERROR_STEP_AND_RANK` 与下文验证命令中的 rank、迭代步保持一致。

**操作步骤**

1. 下发任务

   执行以下命令下发任务：

   ```bash
   kubectl apply -f trjob.yaml
   ```
   >[!NOTE]
   > - 请将 `trjob.yaml` 替换为实际的任务 YAML 文件；若按上文 QWEN3 脚本适配，请使用对应的任务 YAML 与 Pod 名称。
   > - 任务 Pod 的名称、命名空间会根据任务 YAML 中的配置而变化，以下出现的 `process-online-recovery-` 和 `trjob` 均为示例值。

2. 查看任务状态

   1. 执行以下命令查看任务状态：

      ```bash
      kubectl get pod -A -o wide
      ```

   2. 回显示例如下，出现Running表示任务正常运行：

      <pre codetype="bash">
      trjob            process-online-recovery-master-0                   1/1     Running   0                 14s     192.168.75.202   master-69-117   <none>           <none>
      trjob            process-online-recovery-worker-0                   1/1     Running   0                 14s     192.168.6.13     work-69-115     <none>           <none>
      </pre>

3. 监控训练日志

   1. 执行以下命令，监控训练日志检查是否触发UCE故障：

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "status error, rank:8"
      ```

      >[!NOTE]
      > 本示例在第3步将故障注入在rank 8。`grep` 关键字中的rank须与环境变量中配置的全局rank一致。

      回显示例如下，触发UCE故障：

      ```bash
      2026-06-04 09:24:31.767278 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 0
      2026-06-04 09:24:33.767422 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 1417
      ```

      >[!NOTE]
      > 日志中的 `step: 3` 表示故障在第 3 个训练迭代步触发。`npu_status: 2` 表示 MindIO/TTP 侧已进入 UCE 处理状态；在本打桩场景下由软件模拟路径触发，不代表真实硬件片上内存故障。

   2. 执行以下命令，检查第 3 步故障的恢复结果。在 master 或 worker 任一 Pod 上输出大于等于 1，即说明修复成功：
      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "(0, 'Mindio do repair operation ok', {}, 'retry')"|wc -l
      ```

4. 检查迭代是否正常

   1. 执行以下命令查看任务状态：
      ```bash
      kubectl get pod -A -o wide
      ```

      回显示例如下：
      ```bash
      trjob            process-online-recovery-master-0                   1/1     Running   0                 110s    192.168.75.202   master-69-117   <none>           <none>
      trjob            process-online-recovery-worker-0                   1/1     Running   0                 110s    192.168.6.13     work-69-115     <none>           <none>
      ```

      >[!NOTE]
      > 此时请检查 RESTARTS 列，该数值必须保持为 0。证明在整个 UCE 故障及修复过程中，Pod 容器从未发生过重启。

   2. 执行以下命令查看训练迭代步数：

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # 返回：0
      kubectl logs -n trjob process-online-recovery-worker-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # 返回：11
      ```

      >[!NOTE]
      > - 以上命令中 `grep` 的迭代步数（如 `iteration 4`）需根据实际注入故障的步数调整。若故障注入在第 `N` 步，恢复后应从第 `N+1` 步继续训练，因此应 `grep iteration [[:space:]]*{N+1}`。本示例中第 3 步故障对应 `iteration 4`，第 10 步故障对应 `iteration 11`。
      > - 在分布式多机训练中，受训练框架的日志重定向机制影响，各 Rank 的迭代日志可能仅输出在部分节点的 stdout 中，或被重定向至本地物理日志文件。
      > - 本示例中 Master 节点返回 0、Worker 节点返回 11，只要任一节点能搜出大于 0 的计数，即证明热修复后训练已跨越对应故障步数并继续。
