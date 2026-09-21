# Verifying Fault Handling

<!-- md-trans-meta sourceCommit=c4693e618818875051b730a07eb0ae8f922b409d translatedAt=2026-08-31T03:56:24.067Z pushedAt=2026-08-31T05:47:15.260Z -->

This document provides verification methods for various fault handling policies. You can quickly jump to the corresponding verification section based on the fault handling policy actually configured.

**Quick Navigation**

- [Verifying Fault Handling](#verifying-fault-handling)
  - [Verifying Job-Level Rescheduling](#verifying-job-level-rescheduling)
  - [Verifying Pod-Level Rescheduling](#verifying-pod-level-rescheduling)
  - [Verifying Process-Level Rescheduling](#verifying-process-level-rescheduling)
  - [Verifying Process-Level Online Recovery](#verifying-process-level-online-recovery)
    - [MindCluster adaptation](#mindcluster-adaptation)
    - [Script Adaptation](#script-adaptation)
      - [PyTorch Adaptation Example (Based on MindSpeed-LLM)](#pytorch-adaptation-example-based-on-mindspeed-llm)
      - [MindSpore Adaptation Example (Based on MindFormers)](#mindspore-adaptation-example-based-on-mindformers)
    - [Verification Process](#verification-process)

## Verifying Job-Level Rescheduling<a name="verifying-job-level-rescheduling"></a>

**Prerequisites**

In the job YAML for basic scheduling, add the configuration for Job-level rescheduling. For configuration details, see [Configuring Job-Level Rescheduling](03_configuration/02_configuring_fault_handling_policies.md#configuring-job-level-rescheduling). For principles, see [Job-Level Rescheduling](01_solutions_principles.md#job-level-rescheduling).

**Procedure**

1. Dispatch a job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file.
   > - The job Pod names and namespace vary according to the configuration in the job YAML. The `taskmgr-npu-020-default-test-` and `trjob` values that appear below are example values, and the actual values vary according to the configuration in the job YAML.

2. View the job status and UID.

   1. View the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The output example is as follows. A `STATUS` field of `Running` indicates that the job is in normal operation.

      <pre codetype="ColdFusion">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>    0          2s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>    0          3s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. View the UIDs of the two Pods:

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      The output example is as follows:

      ```bash
      7286faf8-f029-450a-b302-5e6e94d4346c
      997add9e-6115-456c-9e8e-e05e4b70bb12
      ```

3. Inject a fault.

   1. Query the job process.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      The output example is as follows:

      ```ColdFusion
      2398104
      2398105
      2398107
      ```

   2. Terminate the process to inject a fault.

      ```bash
      kill -9 2398104
      ```

4. Observe the rescheduling process.

   Monitor the status changes of the 2 Pods of this Job:

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   The historical statuses of the 2 Pods of this Job are as follows. By observing the changes in the bold fields, you can see that the 2 Pods of this Job go through the `Terminating` → `Pending` → `ContainerCreating` → `Running` stages and then resume normal operation, indicating that Job rescheduling succeeds:

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running             0          2s      xx.xx.xx.xx       node173                 &lt;none&gt;           &lt;none&gt;
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running             0          3s      xx.xx.xx.xx       localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   // ===================== Inject fault ======================
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          73s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          85s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          85s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;                 localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;                 localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0          1s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0          1s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Running</strong>             0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Running</strong>             0          2s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   </pre>

5. View the job status and UID.

   1. View the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The output example is as follows. A `STATUS` field value of `Running` indicates that the job is in normal operation.

      <pre codetype="ColdFusion">
      NAMESPACE        NAME                                            READY   STATUS    RESTARTS   AGE     IP                NODE                    NOMINATED NODE   READINESS GATES
      ...              ...                                             ...     ...       ...        ...     ...               ...                     ...              ...
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          2s      xx.xx.xx.xx      node173   &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          33s     xx.xx.xx.xx      node173   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. View the UIDs of the two Pods:

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      The output example is as follows. The UIDs of both Pods of this Job have changed, indicating that both Pods have undergone rescheduling, that is, Job-level rescheduling is triggered:

      ```ColdFusion
      2a24eee8-88f1-4107-bc9d-dabcfb09dea9
      074f9f9c-35f1-4b9e-9298-5b2bcf3759e7
      ```

## Verifying Pod-Level Rescheduling<a name="verifying-pod-level-rescheduling"></a>

**Prerequisites**

In the job YAML for basic scheduling, add the configuration for Pod-level rescheduling. For configuration description, see [Configuring Pod-Level Rescheduling](03_configuration/02_configuring_fault_handling_policies.md#configuring-pod-level-rescheduling). For principles, see [Pod-Level Rescheduling](01_solutions_principles.md#pod-level-rescheduling).

**Procedure**

1. Dispatch a job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file.
   > - The job Pod names and namespace vary according to the configuration in the job YAML. The following `taskmgr-npu-020-default-test-` and `trjob` are example values, and the actual values vary according to the configuration in the job YAML.

2. View the job status and UID.

   1. View the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The output example is as follows. The presence of `Running` indicates that the job is in normal operation.

      <pre codetype="ColdFusion">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>             0          6s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. View the UIDs of the two Pods.

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      The output example is as follows:

      ```ColdFusion
      de1f8848-ed88-4e18-abda-7abc8dbede87
      47291595-85b0-47ff-8393-c922d0e2dfb2
      ```

3. Inject a fault.

   1. Query the job process.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      The output example is as follows:

      ```ColdFusion
      2398132
      2398144
      2398158
      ```

   2. Terminate the process to inject a fault.

      ```bash
      kill -9 2398144
      ```

4. Observe the rescheduling process.

   Monitor the status changes of the 2 Pods of this Job.

   ```bash
   kubectl get pod -A -o wide -w | grep trjob
   ```

   The historical statuses of the 2 Pods of this Job are as follows. By observing the changes in the bold fields, you can see that the faulty Pod (`taskmgr-npu-020-default-test-1`) goes through the `Error` → `Terminating` → `Pending` → `ContainerCreating` → `Running` stages and then resumes normal operation, indicating that Pod rescheduling succeeds:

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running              0          6s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            taskmgr-npu-020-default-test-1                  1/1     Running              0          6s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   // ===================== Inject fault ======================
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Error</strong>               0          34s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          35s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Terminating</strong>         0          35s     xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           1s      &lt;none&gt;                localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>Pending</strong>             0           1s      &lt;none&gt;                localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0           1s      xx.xx.xx.xx     localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 0/1     <strong>ContainerCreating</strong>   0           1s      xx.xx.xx.xx     localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-1</strong>                 1/1     <strong>Running</strong>             0           2s      xx.xx.xx.xx     localhost.localdomain   &lt;none&gt;           &lt;none&gt;
   </pre>

5. View the job status and UID.

   1. View the job status.

      ```bash
      kubectl get pod -A -o wide
      ```

      The output example is as follows. The presence of `Running` indicates that the job is in normal operation.

      <pre codetype="ColdFusion">
      trjob            taskmgr-npu-020-default-test-0                  1/1     <strong>Running</strong>   0          66s      xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
      trjob            taskmgr-npu-020-default-test-1                  1/1     <strong>Running</strong>   0          31s      xx.xx.xx.xx      localhost.localdomain   &lt;none&gt;           &lt;none&gt;
      </pre>

   2. View the UIDs of the two Pods again.

      ```bash
      kubectl get pod taskmgr-npu-020-default-test-0  -n trjob -o jsonpath='{.metadata.uid}'
      kubectl get pod taskmgr-npu-020-default-test-1  -n trjob -o jsonpath='{.metadata.uid}'
      ```

      The output example is as follows: the UID of the `taskmgr-npu-020-default-test-0`Pod remains unchanged, while the UID of the `taskmgr-npu-020-default-test-1` Pod has changed, indicating that only the faulty Pod (`taskmgr-npu-020-default-test-1`) underwent rescheduling, that is, Pod-level rescheduling was triggered:

      ```ColdFusion
      de1f8848-ed88-4e18-abda-7abc8dbede87
      6eb3c217-3b63-457a-9010-9d236d281634
      ```

## Verifying Process-Level Rescheduling<a name="verifying-process-level-rescheduling"></a>

**Prerequisites**

In the job YAML for basic scheduling, add the configuration for process-level rescheduling. For configuration description, see [Configuring Process-Level Rescheduling](03_configuration/02_configuring_fault_handling_policies.md#configuring-process-level-rescheduling). For principles, see [Process-Level Rescheduling](01_solutions_principles.md#process-level-rescheduling).

**Procedure**

1. Dispatch a job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file.
   > - The job Pod name and namespace vary according to the configuration in the job YAML. The `process-reschedule-function-` and `trjob` values shown below are example values, and the actual values vary according to the configuration in the job YAML.

2. View the job status.

   ```bash
   kubectl get pod -A -o wide
   ```

   The output example is as follows. Running indicates that the job is in normal operation:
   <pre codetype="ColdFusion">
   trjob            process-reschedule-function-master-0   1/1     Running   0               14s   xx.xx.xx.xx     master-69-117   &lt;none&gt;           &lt;none&gt;
   trjob            process-reschedule-function-worker-0   1/1     Running   0               14s   xx.xx.xx.xx     work-69-115     &lt;none&gt;           &lt;none&gt;
   </pre>

3. View the iteration step count in the training logs to confirm that training has iterated normally.

   ```bash
   kubectl logs -n trjob process-reschedule-function-worker-0|grep -Po '] iteration [[:space:]]*'|wc -l
   ```

   The output example is as follows:

   ```bash
   50
   ```

4. View the process ID and inject a fault.

   1. View the process ID.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      The output example is as follows:

      ```ColdFusion
      635755
      635756
      635760
      635770
      635777
      635784
      635791
      635795
      ```

   2. Terminate one of the processes to inject a fault.

      ```bash
      kill -9 635777
      ```

5. Observe the training logs.

   ```bash
   kubectl logs -n trjob process-reschedule-function-master-0
   ```

   The output example is as follows:

   ```ColdFusion
   # The following information indicates that the ARF process starts to be triggered.
   Mindx calling notify do ARF repair

   # The following information indicates that ARF succeeds.
   ... Mindio do repair operation ok ...
   ```

6. Check whether the ConfigMap `job-reschedule-reason` contains job information.

   ```bash
   kubectl describe cm -n mindx-dl job-reschedule-reason |grep process-reschedule-function
   ```

   The output example is as follows, which includes the rescheduling time, the pod, node, and rank that triggered the rescheduling, and the current rescheduling count of this job:

   ```ColdFusion
   {"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07":{"JobID":"trjob/process-reschedule-function-ebfbc149-5312-4232-a021-453db0d4ce07","TotalRescheduleTimes":1,"RescheduleRecords":[{"LogFileFormatTime":"I0603 05:16:52","RescheduleTimeStamp":1780435012,"ReasonOfTask":[{"RescheduleReason":"pod-failed","PodName":"process-reschedule-function-worker-0","NodeName":"work-69-115","NodeRankIndex":"1"}]}]}}
   ```

## Verifying Process-Level Online Recovery<a name="verifying-process-level-online-recovery"></a>

This section guides users through the adaptation steps for verifying process-level online recovery by injecting an on-chip memory UCE fault into the training code.

>[!NOTE]
>
>- The modifications described in this section are intended only to guide users in verifying the process-level online recovery feature in a test environment. Do not deploy this fault-injection version to a production environment.
>- Before configuring the steps in this section, ensure that training can start normally and that process-level online recovery has been configured.
>- To ensure the normal operation of process-level online recovery, keep the clocks of the K8s cluster master nodes and worker nodes synchronized.
>- The code in the following content may differ from the actual version. Use the code in the actual version as the authoritative reference.

### MindCluster adaptation<a name="ZH-CN_TOPIC_0000002479386410"></a>

1. <a name="li977718409381"></a>Pull the MindCluster code.

    ```shell
    mkdir -p /data/atlas_dls/public/code
    cd /data/atlas_dls/public/code
    git clone https://gitcode.com/Ascend/mind-cluster.git
    cd ./mind-cluster/component/clusterd
    git checkout branch_v26.1.0   # branch_v26.1.0 is the repository version branch. Switch to the target branch as needed.
    ```

2. Modify the ClusterD code.
   1. Open the `pkg/application/faultmanager/jobprocess/faultrank/job_fault_rank_processor.go` file.

      ```shell
      vi pkg/application/faultmanager/jobprocess/faultrank/job_fault_rank_processor.go
      ```

   2. Press `i` to enter edit mode and add the following bold code.

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
               <strong>collector.ReportInfoCollector.ReportRetryInfo(podInfo.jobId, deviceInfo.RankID, constant.JobNotRecover, constant.UceFaultType)  // Set the service-plane fault time to an invalid value to prevent a single fault from repeatedly triggering process-level online recovery.</strong>
            }
        …
      </pre>

   3. Press `Esc`, type `:wq!`, and press `Enter` to save and exit editing.

3. <a name="li114977117517"></a>Compile ClusterD.

   ```shell
   cd ./build/
   chmod +x build.sh && dos2unix build.sh
   sed -i 's|build_version="v[^"]\+"|build_version="xxx"|g' build.sh  # Replace xxx with the version number, for example, v26.1.0.
   sed -i 's|export CGO_ENABLED=0|export CGO_ENABLED=1|g' build.sh  # Enable the CGO feature.
   ./build.sh # Compile ClusterD. Install the Go SDK in advance. For the specific version, refer to the go.mod file in the ClusterD component code.
   ```

   After successful compilation, related files are generated in the `../output/` directory. Run the following command to view them:

   ```shell
   ll ../output/
   ```

   The output example is as follows:

   ```bash
   -r-x------. 1 root root 45891128 Aug 13 10:52 clusterd
   -r--------. 1 root root     4021 Aug 13 10:52 clusterd-v26.1.0.yaml
   -r--------. 1 root root      946 Aug 13 10:52 Dockerfile
   -r--------. 1 root root      209 Aug 13 10:52 faultDuration.json
   -r--------. 1 root root      207 Aug 13 10:52 fdConfig.yaml
   -r--------. 1 root root      467 Aug 13 10:52 publicFaultConfiguration.json
   -r--------. 1 root root      756 Aug 13 10:52 relationFaultCustomization.json
   ```

4. <a name="li89701053589"></a>Enter the `output` directory and build the ClusterD image.

   ```shell
   cd ../output/
   docker build --no-cache -t clusterd:{tag} ./  # {tag} must be consistent with the value of build_version="xxx" in step 3.
   ```

5. (Optional) Save the image, and upload the saved image file and the `clusterd-{tag}.yaml` file to the primary node. If [step 1](#li977718409381) through [step 4](#li89701053589) are performed on the primary node, you can skip this step.

   ```shell
   docker save -o clusterd.tar clusterd:{tag}  # Save the image.
   docker load -i clusterd.tar  # Import the image on the primary node.
   ```

6. Restart ClusterD on the primary node.

   ```shell
   kubectl delete -f  clusterd-{tag}.yaml  # Delete the old ClusterD container.
   kubectl apply -f  clusterd-{tag}.yaml  # Start the new container.
   ```

### Script Adaptation<a name="ZH-CN_TOPIC_0000002479226412"></a>

#### PyTorch Adaptation Example (Based on MindSpeed-LLM)<a name="ZH-CN_TOPIC_0000002511426361"></a>

1. Set up the training environment and start training. For details, see [PyTorch Adaptation Example (Based on MindSpeed-LLM)](04_using_resumable_training_on_the_cli.md#zh-cn_topic_0000002003180016_section412442472511).
2. Enable process-level online recovery. For details, see [Configuring Process-Level Online Recovery](03_configuration/02_configuring_fault_handling_policies.md#configuring-process-level-online-recovery).
3. Add the following bold content to the `QWEN3\_for\_PyTorch\_2.7\_code/mindspeed_llm/training/training.py` code to inject a fault through instrumentation. The newly added code obtains the fault injection iteration position and the fault rank information based on the environment variable `RAISE_UCE_ERROR_STEP_AND_RANK`.

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

4. Modify the startup script `QWEN3_for_PyTorch_2.7_code/scripts/train_start.sh`.

   ```shell
   …
   export RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"  # Configure the iteration and rank for fault injection. Inject UCE faults on rank 8 in the 3rd iteration and rank 9 in the 10th iteration.
   sed -i 's/check_memory_result = torch_npu.npu.check_uce_in_memory(device)/check_memory_result = ha_constant.UCE_HIGH_LEVEL/g' /job/code/mindspeed_llm/core/high_availability/tft_stop_clean.py # Modify the TorchNPU interface return value to identify the exception thrown by the training code as a UCE fault.
   …
   ```

#### MindSpore Adaptation Example (Based on MindFormers)<a name="ZH-CN_TOPIC_0000002511346369"></a>

1. Set up the training environment and start training. For details, see [MindSpore Adaptation Example (Based on MindFormers)](04_using_resumable_training_on_the_cli.md#zh-cn_topic_0000002003180016_section718243883518).
2. Enable process-level online recovery. For details, see [Configuring Process-Level Online Recovery](03_configuration/02_configuring_fault_handling_policies.md#configuring-process-level-online-recovery).
3. Add the following bold content to the `QWEN3_for_MS_code/mindformers/core/callback/callback.py` code to inject a fault through instrumentation.

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

4. Modify the startup script `QWEN3_for_MS_code/scripts/msrun_launcher.sh`.

   ```shell
   …
   export RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"  # Configure the iteration and rank for fault injection. Inject UCE faults on rank 8 at iteration 3 and rank 9 at iteration 10.
   sed -i 's/err_strategy = _get_uce_process_strategy()/err_strategy = "RS_UCE_LOWLEVEL"/g' $(pip3 show mindspore | grep Location | awk -F ' ' '{print $2}')/mindspore/train/callback/_train_fault_tolerance.py # Modify the UCE handling strategy.
   …
   ```

### Verification Process

The following example is based on a **two-node 16-processor** environment (8 processors per node, Master rank 0–7, Worker rank 8–15), consistent with the `RAISE_UCE_ERROR_STEP_AND_RANK="{3:8,10:9}"` configuration in [Script Adaptation](#ZH-CN_TOPIC_0000002479226412). If you use a single node or another topology, adjust the environment variables and the rank and Pod names in the `grep` commands below accordingly.

**Prerequisites**

- In the job YAML for basic scheduling, add the configuration for process-level online recovery. For configuration details, see [Configuring Process-level Online Recovery](03_configuration/02_configuring_fault_handling_policies.md#configuring-process-level-online-recovery). For principles, see [Process-level Online Recovery](01_solutions_principles.md#process-level-online-recovery).
- MindCluster adaptation and script adaptation have been completed. The `RAISE_UCE_ERROR_STEP_AND_RANK` in the startup script must be consistent with the rank and iteration step in the verification commands below.

**Procedure**

1. Dispatch a job.

   ```bash
   kubectl apply -f trjob.yaml
   ```

   >[!NOTE]
   > - Replace `trjob.yaml` with the actual job YAML file. If you adapted the QWEN3 script as described above, use the corresponding job YAML and Pod name.
   > - The job Pod name and namespace vary according to the configuration in the job YAML. The `process-online-recovery-` and `trjob` values shown below are examples.

2. View the job status.

   ```bash
   kubectl get pod -A -o wide
   ```

   The output example is as follows. A status of `Running` indicates that the job is in normal operation.

   ```ColdFusion
   trjob            process-online-recovery-master-0                   1/1     Running   0                 14s     xx.xx.xx.xx     master-x   <none>           <none>
   trjob            process-online-recovery-worker-0                   1/1     Running   0                 14s     xx.xx.xx.xx     worker-x   <none>           <none>
   ```

3. Monitor the training logs.
   1. Monitor the training logs to check whether a UCE fault is triggered.

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "status error, rank:8"
      ```

      >[!NOTE]
      > In this example, the fault is injected into rank 8 in step 3. The rank in the `grep` keyword must be consistent with the global rank configured in the environment variable.

      The output example is as follows, indicating that a UCE fault is triggered.

      ```ColdFusion
      2026-06-04 09:24:31.767278 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 0
      2026-06-04 09:24:33.767422 warn 3062106 [TTP controller.cpp:2510] status error, rank:8 step: 3 npu_status: 2 run_status: 0 data_aval: 0 data_status: 0 diff_time : 1417
      ```

      >[!NOTE]
      > In the log, `step: 3` indicates that the fault is triggered at the third training iteration step. `npu_status: 2` indicates that the MindIO/TTP side has entered the UCE processing state; in this stubbing scenario, it is triggered by the software simulation path and does not represent a real on-chip memory fault in the hardware.

   2. Check the recovery result of the fault at step 3. If the output on either the Master or Worker Pod is greater than or equal to 1, the repair is successful.

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 --all-containers=true | grep -Fa "(0, 'Mindio do repair operation ok', {}, 'retry')"|wc -l
      ```

4. Check whether the iteration is normal.

   1. View the job status:

      ```bash
      kubectl get pod -A -o wide
      ```

      The output example is as follows:

      ```ColdFusion
      trjob            process-online-recovery-master-0                   1/1     Running   0                 110s    xx.xx.xx.xx     master-x   <none>           <none>
      trjob            process-online-recovery-worker-0                   1/1     Running   0                 110s    xx.xx.xx.xx     worker-x   <none>           <none>
      ```

      >[!NOTE]
      > At this point, check the `RESTARTS` column. This value must remain `0`, which proves that the Pod never restarted throughout the entire UCE fault triggering and repair process.

   2. View the training iteration count.

      ```bash
      kubectl logs -n trjob process-online-recovery-master-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # Returns: 0
      kubectl logs -n trjob process-online-recovery-worker-0 | grep -Po "] iteration [[:space:]]*4"|wc -l
      # Returns: 11
      ```

      >[!NOTE]
      > - In the commands above, the iteration count used by `grep` (such as `iteration 4`) must be adjusted according to the step at which the fault is actually injected. If the fault is injected at step `N`, training should resume from step `N+1` after recovery, so the pattern should be `grep iteration [[:space:]]*{N+1}`. In this example, the fault at step 3 corresponds to `iteration 4`, and the fault at step 10 corresponds to `iteration 11`.
      > - In distributed multi-node training, due to the log redirection mechanism of the training framework, the iteration logs of each rank may be output only to the stdout of some nodes, or redirected to local physical log files.
      > - In this example, the Master node returns 0 and the Worker node returns 11. As long as any node has a count greater than 0, it proves that training has crossed the corresponding fault step and continued after hot repair.
