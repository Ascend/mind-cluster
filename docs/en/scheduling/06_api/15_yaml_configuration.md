# Job YAML Configuration<a name="yaml_configuration"></a>

## acjob YAML Parameter Description<a name="acjob"></a>

The YAML parameters available for training acjob are described in the following table.

**Table 1** Key field description for acjob

|Field Path|Type|Format|Description|
|--|--|--|--|
|`apiVersion`|String|-|Defines the versioned schema of the object representation. The server converts it to the latest internal value and rejects unrecognized versions. For more information, see [Types](https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds).|
|`kind`|String|-|Indicates the REST resource type corresponding to this object. The value is inferred from the endpoint, cannot be updated, and uses camelCase naming. For more information, see [Resources](https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources).|
|`metadata`|Object|-|Kubernetes metadata (such as namespace, labels, etc.). For more information, see [Metadata](https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata).|
|`metadata.labels.app`|String|-|<p>Indicates the role of the MindIE Motor job in acjob. Valid values include `mindie-ms-controller`, `mindie-ms-coordinator`, and `mindie-ms-server`.</p><ul><li>When the acjob YAML contains both the `jobID` and `app` fields, Ascend Operator automatically injects the environment variables `MINDX_TASK_ID`, `APP_TYPE`, `MINDX_SERVER_IP`, and `MINDX_SERVER_DOMAIN`, and identifies it as a MindIE inference job.</li><li>For details about the preceding environment variables, see [Training Environment Variables Injected by Ascend Operator](./13_environment_variable_description.md#ascend-operator-environment-variables).</li><li>This parameter is supported only on Atlas 800I A3 SuperPoD server and Atlas 800I A2 inference server.</li></ul>|
|`metadata.labels.mind-cluster/scaling-rule: scaling-rule`|String|-|Marks the ConfigMap name corresponding to the scaling rule. This parameter is supported only for MindIE Motor inference jobs on Atlas 800I A3 SuperPoD server and Atlas 800I A2 inference server.|
|`metadata.labels.mind-cluster/group-name: group0`|String|-|Marks the group name corresponding to the scaling rule. This parameter is supported only for MindIE Motor inference jobs on Atlas 800I A3 SuperPoD server and Atlas 800I A2 inference server.|
|`metadata.labels.framework`|String|-|AI framework type. Valid values are `pytorch` and `mindspore`.|
|`metadata.labels.jobID`|String|-|The unique ID of the current MindIE Motor job in the cluster. Users can configure it based on actual conditions. This parameter is supported only on Atlas 800I A3 SuperPoD server and Atlas 800I A2 inference server.|
|`metadata.labels.pod-rescheduling`|String|-|<p>Pod-level rescheduling. Indicates that after a job fault occurs, not all job Pods are deleted. Instead, the faulty Pod is deleted, and a new Pod is created for rescheduling.</p><ul><li>`on`: Enables Pod-level rescheduling.</li><li>Other values or when this field is not used: Disables Pod-level rescheduling.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>The default rescheduling mode is Job-level rescheduling. To enable Pod-level rescheduling, add this field.</li></ul></div></div>|
|`metadata.labels.process-recover-enable`|String|-|<p>Ascend Operator automatically adds the `process-recover-enable=on` label to the job based on the recover-strategy configured by the user. No manual specification is required.</p><ul><li>`on`: Enables process-level rescheduling and process-level online recovery.<p>Process-level rescheduling and graceful fault tolerance cannot be enabled at the same time. If both are enabled, resumable training continues through Job-level rescheduling.</p></li><li>`pause`: Temporarily disables process-level rescheduling and process-level online recovery.</li><li>`off` or when this field is not used: Disables process-level rescheduling and process-level online recovery.</li></ul>|
|`metadata.annotations.recover-strategy`|String|-|<p>Available recovery strategies for the job. `recover-strategy` is configured under annotations in the job YAML. The value is any combination of the six strategies, separated by commas.</p><ul><li>`retry`: Process-level online recovery.</li><li>`recover`: Process-level rescheduling.</li><li>`recover-in-place`: Process-level in-place recovery.</li><li>`elastic-training`: Elastic training.</li><li>`dump`: Save dying gasp.</li><li>`exit`: Exit training.</li></ul>|
|`metadata.labels.subHealthyStrategy`|String|-|<p>Processing strategy for nodes whose status is `SubHealthy`.</p><ul><li>`ignore`: Ignores the subhealthy node. Subsequent jobs do not preferentially schedule to this node in affinity scheduling.</li><li>`graceExit`: Does not use the subhealthy node, saves the dying gasp checkpoint file, and then performs rescheduling. Subsequent jobs are not scheduled to this node.</li><li>`forceExit`: Does not use the subhealthy node, exits the job directly without saving, and performs rescheduling. Subsequent jobs are not scheduled to this node.</li><li>`hotSwitch`: Performs hot switching. After pulling up the backup Pod, pauses the training job and uses a new node to restart training.</li><li>The default value is ignore.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When using the graceExit strategy, ensure that the job has enabled the last CKPT saving function.</li><li>For usage constraints of the `hotSwitch` strategy, see [Usage Constraints](../04_usage/04_resumable_training/01_solutions_principles.md#hot-switching).</li></ul></div></div>|
|`metadata.labels.fault-scheduling`|String|-|<ul><li>`grace`: Configures the job to use graceful deletion mode. During the process, the original Pod is gracefully deleted first. If the deletion is not successful after 15 minutes, the original Pod is forcibly deleted. For process-level rescheduling and process-level online recovery scenarios, set this parameter to `grace`.</li><li>`force`: Configures the job to use forced deletion mode. During the process, the original Pod is forcibly deleted.</li><li>`off`: The job does not use the resumable training feature, and the K8s maxRetry still takes effect.</li><li>None (no `fault-scheduling` field): The job does not use the resumable training feature, and the K8s maxRetry still takes effect.</li><li>Other values: The job does not use the resumable training feature, and the K8s maxRetry still takes effect.</li></ul>|
|`metadata.labels.fault-retry-times`|Integer|int32|<p>To handle service plane faults, the number of unconditional retries on the service plane must be configured.</p><ul><li>0 &lt; `fault-retry-times`: To handle service plane faults, the number of unconditional retries on the service plane must be configured.</li><li>None (no `fault-retry-times`) or 0: The job does not use the unconditional retry function and cannot detect service plane faults. The vcjob maxRetry still takes effect.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>To use the unconditional retry function, ensure that a training process exception causes the container to exit abnormally. If the container does not exit abnormally, the retry cannot succeed.</li><li>Currently, only Atlas 800T A2 training server and Atlas 900 A2 PoD cluster basic unit support the unconditional retry function.</li><li>During process-level recovery, service plane faults are triggered. To use process-level recovery, this parameter must be configured.</li></ul></div></div>|
|`metadata.labels.ring-controller.atlas`|String|-|Used to distinguish the chip type used by the job.<ul><li><term>Atlas A2 training products</term>, A200T A3 Box8 SuperPoD server, Atlas 900 A3 SuperPoD, and Atlas 800T A3 SuperPoD server: `ascend-{xxx}b`</li><li>Atlas 800 training server (with Atlas 300T training card): `ascend-910`</li><li>(Optional) Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E servers, and Atlas 950 SuperPoD: `ascend-npu`</li></ul>|
|`metadata.labels.podgroup-sched-enable`|String|-|<p>Configured only when the cluster uses the openFuyao-customized Kubernetes and volcano-ext components.</p><ul><li>When the value is set to `true`, the batch scheduling function is enabled.</li><li>When the value is set to other strings, the batch scheduling function does not take effect, and normal scheduling is used.</li></ul><p>If this parameter is not configured, the batch scheduling function does not take effect, and normal scheduling is used.</p><span class="notetitle">Note</span><div class="notebody"><ul><li>This parameter supports only the full-NPU scheduling feature of the Volcano scheduler.</li><li>This parameter is supported only on Atlas 900 A3 SuperPoD and Atlas 800T A3 SuperPoD server.</li></ul></div>|
|`metadata.labels.tor-affinity`|String|-|<p>The default value is `null`, indicating that switch affinity scheduling is not used.</p><ul><li>`large-model-schema`: Large model tasks or padding tasks</li><li>`normal-schema`: Normal tasks</li><li>`null`: Switch affinity scheduling is not used</li></ul><span class="notetitle">Note</span><div class="notebody">Users need to select the task type based on the number of task replicas. A task with fewer than 4 replicas is a padding task. A task with 4 or more replicas is a large model task. Normal tasks have no restriction on the number of task replicas.</div><p>Users need to configure based on the task type.</p><ul><li>Switch affinity scheduling 1.0 supports <term>Atlas training products</term> and <term>Atlas A2 training products</term>; supports the PyTorch and MindSpore frameworks.</li><li>Switch affinity scheduling 2.0 supports <term>Atlas A2 training products</term>; supports the PyTorch framework.</li><li>Only full NPUs support switch affinity scheduling. Static vNPU does not support switch affinity scheduling.</li></ul>|
|`metadata.annotations['sp-block']`|String|-|<p>Specifies the `sp-block` field. The cluster scheduling components divide the physical SuperPoD into logical SuperPoDs based on the splitting policy for affinity scheduling of training jobs. If the user does not specify this field, the logical SuperPoD size of a job is set to the total number of NPUs configured for the job during scheduling.</p><ul><li>For single-node jobs, it must be consistent with the number of chips requested by the job.</li><li>For distributed jobs, it must be an integer multiple of the number of chips on a node, and the total number of chips of the job must be an integer multiple of it.</li></ul><p>For details, see [UnifiedBus Device Node Network Description](../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod).</p><span class="notetitle">Note</span><div class="notebody"><ul><li>This field is supported only on Atlas 900 A3 SuperPoD, Atlas 800T A3 SuperPoD server, Atlas 800I A3 SuperPoD server, Atlas 850E SuperPoD, and Atlas 950 SuperPoD.</li><li>After this field is used, the `tor-affinity` field does not need to be configured additionally.</li><li>FAQ: [Total number of chips requested by the job is 32, sp-block set to 32 allows normal training, sp-block set to 16 fails to complete training, training container error report indicates initialization connection failed](https://gitcode.com/Ascend/mind-cluster/issues/377)</li></ul></div>|
|`metadata.annotations['ra-block']`|String|-|<p>Identifier for rack affinity scheduling. Specifies the `ra-block` field. On the premise of supporting dynamic ratio, a single rack with 64 cards is divided into 4 OSs. Each OS is considered a node in the K8s cluster. The intra-rack communication latency is lower than the inter-rack communication latency. This field is configured for rack affinity scheduling of training jobs.</p><p>The value range is 0 to 64 and must be a power of 2.</p><p>By enumeration, the values of `ra-block` are {1, 2, 4, 8, 16, 32, 64}.</p><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody">This field is supported only on Atlas 950 SuperPoD.</div></div>|
|`metadata.annotations.huawei.com/schedule_policy`|String|-|Configures the AI chip layout form that the job needs to schedule. Volcano selects an appropriate scheduling policy based on this field. Currently, the configurations in [huawei.com/schedule_policy Configuration Description](#huaweicomschedule_policy-configuration-description) are supported.|
|`huawei.com/affinity-config`|String|-|<p>Configures the affinity levels of multi-level scheduling for the job.</p><p>Value: `level1=x`,`level2=y`,...</p><p>Where x,y... are the subtask sizes of the corresponding network levels.</p><p>The value must be a concatenation of strings in the format "leveli=ni", separated by English commas. Here, `i` is the network level sequence number, and `ni` is the number of replicas of the subtask at that network level. For example, for a task with a total of 8 replicas, "level1=2,level2=4" indicates that every 2 Pods in the task Pods are assigned to nodes with the same level1 label, and every 4 Pods are assigned to nodes with the same level2 label.</p><p>The network level configuration must meet the following requirements:<ul><li>When the task has more than 1 level, the value of level n must be an integer multiple of n-1.</li><li>The total number of task replicas must be an integer multiple of all levels.</li><li>The task level configuration must start from level1 and be continuous from small to large.</li></ul></p>|
|`spec`|Object|-|Specification description of the desired state of acjob. `replicaSpecs` must be configured.|
|`spec.template.metadata.annotations.huawei.com/recover_policy_path`|String|-|Job rescheduling policy. When the value is `pod`, only Pod-level rescheduling is supported and it is not upgraded to Job-level rescheduling.|
|`spec.template.metadata.annotations.huawei.com/schedule_minAvailable`|Integer|-|The default value is the total number of task replicas. When Ascend Operator enables "gang" scheduling and the scheduler is Volcano, this is the total number of replicas for task running.|
|`metadata.annotations.wait-reschedule-timeout`|Integer|int32|Timeout for waiting for faulty node rescheduling during process-level rescheduling, in seconds. The default value is `270`. The value range is 30 to 270.|
|`spec.replicaSpecs`|Object|-|Mapping from `ReplicaType` to `ReplicaSpec`, specifying the MS cluster configuration. Example: { "Scheduler": ReplicaSpec, "Worker": ReplicaSpec }.|
|`spec.replicaSpecs.[ReplicaType]`|Object|-|Description of the replica.|
|`spec.replicaSpecs.[ReplicaType].replicas`|Integer|int32|Number of replicas, indicating the number of replicas required for the given template. The default value is 1.|
|`spec.replicaSpecs.[ReplicaType].restartPolicy`|String|-|<p>Container restart policy. The default value is `Never`. When unconditional retry for service plane faults is configured, the container restart policy must be set to `Never`.</p><ul><li>`Never`: Never restart</li><li>`Always`: Always restart</li><li>`OnFailure`: Restart on failure</li><li>`ExitCode`: Decide whether to restart the Pod based on the process exit code. When the error code is 1 to 127, the Pod is not restarted. When the error code is 128 to 255, the Pod is restarted.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody">vcjob-type training jobs do not support `ExitCode`.</div></div>|
|`spec.replicaSpecs.[ReplicaType].template`|Object|-|Kubernetes Pod template. For more information, see [Kubernetes Pod Template](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-template-v1/).|
|`spec.replicaSpecs.[ReplicaType].template.spec.hostNetwork`|String|-|<ul><li>`true`: Creates the Pod using `HostIP`. In this case, the environment variable `HCCL_IF_IP` must be configured as `status.hostIP` in the YAML. When the cluster is large (more than 1000 nodes), it is recommended to create Pods using `HostIP`.</li><li>`false`: Does not create the Pod using `HostIP`. When this parameter is not passed or its value is `false`, the preceding environment variable does not need to be configured.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody">When Pods are created using `HostIP`, the problems of slow Pod creation and slow communication between Pods still exist. In this case, it is recommended to mount the RankTable file, obtain the `HostIP` of the Pod by parsing the RankTable file, and inject it into the environment variables of the corresponding framework (for example, MindSpore injects it into `MS_SCHED_HOST`) to establish the link.</div></div>|
|`spec.replicaSpecs.[ReplicaType].template.spec.containers.name`|String|-|Container name. Currently, it must be `ascend`.|
|`spec.replicaSpecs.[ReplicaType].template.spec.containers.image`|String|-|Training image name. Modify it based on actual conditions (the image name created by the user in the image creation section).|
|`spec.replicaSpecs.[ReplicaType].template.spec.containers.ports`|Object|-|Distributed training collective communication port. The value of `name` can only be `ascendjob-port`. The `containerPort` can be set by the user based on actual conditions. If it is not set, the default port 2222 is used.|
|<ul><li>`spec.replicaSpecs.[ReplicaType].template.spec.containers.resources.requests`</li><li>`spec.replicaSpecs.[ReplicaType].template.spec.containers.resources.limits`</li></ul>|Object|-|<p>Limits the requested NPU or vNPU type (only one type can be requested) and quantity. Modify them based on actual conditions. `limits` must be consistent with the chip name and quantity of `requests`.</p><p><strong>Full-NPU scheduling:</strong></p><ul><li>Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD: Configure as `huawei.com/npu`: <em>x</em></li><li>Inference server (with Atlas 300I inference card): Configure as `huawei.com/Ascend310`: <em>x</em></li><li><term>Atlas inference products</term> non-mixed mode: Configure as `huawei.com/Ascend310P`: <em>x</em></li><li><term>Atlas inference products</term> mixed mode:<ul><li>Configure as `huawei.com/Ascend310P-V`: <em>x</em></li><li>Configure as `huawei.com/Ascend310P-VPro`: <em>x</em></li><li>Configure as `huawei.com/Ascend310P-IPro`: <em>x</em></li></ul></li><li>Other products: Configure as `huawei.com/Ascend910`: <em>x</em></li></ul><p>Depending on the chip type used, the value of `x` is as follows:</p><ul><li>Atlas 800 training server (fully populated with NPUs):<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4, 8</li><li>Distributed: 1, 2, 4, 8</li></ul></li><li>Atlas 800 training server (half populated with NPUs):<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4</li><li>Distributed: 1, 2, 4</li></ul></li><li>Server (with Atlas 300T training card):<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2</li><li>Distributed: 2</li></ul></li><li>Atlas 800T A2 training server and Atlas 900 A2 PoD cluster basic unit:<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul></li><li>Atlas 200T A2 Box16 heterogeneous subrack:<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li></ul></li><li>Atlas 900 A3 SuperPoD, A200T A3 Box8 SuperPoD server, and Atlas 800T A3 SuperPoD server:<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4, 6, 8, 10, 12, 14, 16</li><li>Distributed: 2, 4, 6, 8, 10, 12, 14, 16</li><li>For logical SuperPoD affinity tasks on Atlas 900 A3 SuperPoD: 16</li></ul></li><li>Server (with Atlas 350 accelerator card) (8 processors in a node without interconnect):<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul></li><li>Servers (with Atlas 350 accelerator card) (16 processors in a node without interconnect):<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li></ul></li><li>Servers (with Atlas 350 accelerator cards) (4 processor-meshed, 8 processors):<ul><li>Single-node (affinity satisfied): 1, 2, 3, 4, 8</li><li>Single-node (affinity not guaranteed): 5, 6, 7</li><li>Distributed (affinity satisfied): 1, 2, 3, 4, 8</li><li>Distributed (affinity not guaranteed): 5, 6, 7</li></ul></li><li>Servers (with Atlas 350 accelerator cards) (4 processor-meshed, 16 processors):<ul><li>Single-node (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li><li>Single-node (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li><li>Distributed (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li><li>Distributed (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li></ul></li><li>Atlas 650E server:<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul></li><li>Atlas 850E SuperPoD:<ul><li>Single-node: 1, 2, 4, 8 (the `sp-block` value must be consistent with it)</li><li>Distributed: 8 (the `sp-block` value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not be greater than the physical SuperPoD size)</li></ul></li><li>Atlas 950 SuperPoD:<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8 (the `sp-block` value must be consistent with it)</li><li>Distributed: 8 (the `sp-block` value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not be greater than the physical SuperPoD size)</li></ul></li></ul><p><strong>Static vNPU scheduling:</strong></p><p>`huawei.com/Ascend910-Y`: 1</p><p>The value is 1. Only vNPU under one NPU can be used, </p><p>for example, `huawei.com/Ascend910-6c.1cpu.16g: 1`</p>|
|`spec.replicaSpecs.{Master\|Scheduler\|Worker}.template.spec.containers[0].env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath`| String|-|<p>The value is `metadata.annotations['huawei.com/AscendXXX']`, where *XXX* indicates the chip model. Supported values are 910, 310, and 310P. The value must be consistent with the actual chip type in the environment.</p><p>Ascend Docker Runtime obtains this parameter value to mount the corresponding type of NPU to the container.</p><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>This parameter supports only the full-NPU scheduling feature of the Volcano scheduler. Users who use static vNPU scheduling or other schedulers need to delete the fields related to this parameter from the sample YAML.</li><li>For Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD, configure this parameter as `metadata.annotations['huawei.com/npu']`.</li></ul></div></div>|
|`spec.replicaSpecs.{Master\|Scheduler\|Worker}.template.spec.terminationGracePeriodSeconds`| Integer|0 &lt; `terminationGracePeriodSeconds` &lt; `grace-over-time`|<p>The time from when the container receives SIGTERM to when it is forcibly stopped by K8s. This time must be greater than 0 and less than the value of the `grace-over-time` in the `volcano-v{version}.yaml` file. It must also be sufficient to save the checkpoint file. Modify it based on the actual situation. For details, see [Container Lifecycle Hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/) on the Kubernetes official website.</p><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody">This field takes effect only when `fault-scheduling` is set to `grace`. When `fault-scheduling` is set to `force`, this field is invalid.</div></div>|
|`spec.runPolicy`|Object|-|Encapsulates the runtime policies (such as resource cleanup and active time) of a distributed training job.|
|`spec.runPolicy.backoffLimit`|Integer|int32|The number of retries allowed before a job fails (optional).<ul><li>0 &lt; `backoffLimit`: the number of times a job can be rescheduled. When a job fails, this is the number of times it can be rescheduled. When the number of rescheduling already performed equals the value of `backoffLimit`, the job will no longer be rescheduled.</li><li>None (no `backoffLimit`) or `backoffLimit` ≤ 0: the total number of rescheduling is not limited.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>When both `backoffLimit` and `fault-retry-times` are configured, rescheduling stops when the number of rescheduling already performed equals either the value of `backoffLimit` or the value of `fault-retry-times`.</p><p>If `backoffLimit` is not configured but `fault-retry-times` is configured, the number of rescheduling specified by `fault-retry-times` is used.</p></div></div>|
|`spec.runPolicy.activeDeadlineSeconds`|Integer|int64|The maximum time (in seconds) that a job remains active. The value must be a positive integer. Currently meaningless and will be removed in a later version.|
|`spec.runPolicy.cleanPodPolicy`|String|-|The policy for cleaning up Pods after a job completes. The default value is `Running`. Currently meaningless and will be removed in a later version.|
|`spec.runPolicy.ttlSecondsAfterFinished`|Integer|int32|Time to live (TTL) after a job completes. The default value is `unlimited`, and actual deletion may be delayed. Currently meaningless and will be removed in a later version.|
|`spec.runPolicy.schedulingPolicy`|Object|-|The scheduling policy (such as gang-scheduling).|
|`spec.runPolicy.schedulingPolicy.minAvailable`|Integer|int32|The minimum number of available resources. The default value is the total number of task replicas. When Ascend Operator enables "gang" scheduling and the scheduler is Volcano, this is the total number of replicas for task execution.|
|`spec.runPolicy.schedulingPolicy.minResources`|Object|-|The minimum set of resources allocated by resource name (supports integer or string format).|
|`spec.runPolicy.schedulingPolicy.priorityClass`|String|-|The name of the priority class.|
|`spec.runPolicy.schedulingPolicy.queue`|String|-|The name of the scheduling queue. The default value is `default`. Users need to fill in the value based on their own situation. When Ascend Operator enables "gang" scheduling and the scheduler is Volcano, this is the queue to which the job belongs.|
|`spec.schedulerName`|String|-|The scheduler selected when Ascend Operator enables "gang" scheduling. The default value is `volcano`. Users need to fill in the value based on their own situation.|
|`spec.successPolicy`|String|-|The criterion for marking AscendJob as successful. Currently meaningless. A job is determined as successful only when all Pods succeed. This field will be removed in a later version.|
|`status`|Object|-|The latest observed status of AscendJob (read-only). Required fields: `conditions`, `replicaStatuses`.|
|`status.completionTime`|String|date-time|The job completion time (RFC3339 format, UTC).|
|`status.conditions`|Array|-|The array of current job conditions.|
|`status.conditions[type]`|String|-|The type of the job condition (for example, "Complete").|
|`status.conditions[status]`|String|-|The condition status: `True`, `False`, `Unknown`.|
|`status.conditions[lastTransitionTime]`|String|date-time|The time when the condition status transitioned.|
|`status.conditions[lastUpdateTime]`|String|date-time|The final time after the condition was updated.|
|`status.conditions[message]`|String|-|The detailed description of the condition.|
|`status.conditions[reason]`|String|-|The reason for the condition transition.|
|`status.lastReconcileTime`|String|date-time|The time when the job was last reconciled (RFC3339 format, UTC).|
|`status.replicaStatuses`|Object|-|The mapping from replica types to replica statuses.|
|`status.replicaStatuses.[ReplicaType].active`|Integer|int32|The number of running Pods.|
|`status.replicaStatuses.[ReplicaType].failed`|Integer|int32|The number of failed Pods.|
|`status.replicaStatuses.[ReplicaType].succeeded`|Integer|int32|The number of successful Pods.|
|`status.replicaStatuses.[ReplicaType].labelSelector`|Object|-|The Pod label selector (defines how Pods are filtered).|
|`status.replicaStatuses.[ReplicaType].labelSelector.matchExpressions`|Array|-|The label matching rules (supports operators such as In, NotIn, Exists, and DoesNotExist).|
|`status.replicaStatuses.[ReplicaType].labelSelector.matchLabels`|Object|-|The key-value pairs for label matching (equivalent to matchExpressions conditions).|
|`status.startTime`|String|date-time|The job start time (RFC3339 format, UTC).|
|`metadata.annotations['huawei.com/AscendXXX']`|String|-|*XXX* indicates the chip model. Supported values are 910, 310, and 310P. The value must be consistent with the actual chip type in the environment. Ascend Docker Runtime obtains this parameter value to mount the corresponding type of NPU to the container.<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>This parameter supports only the full-NPU scheduling feature of the Volcano scheduler. Users who use static vNPU scheduling or other schedulers need to delete the fields related to this parameter from the sample YAML.</li><li>For Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD, configure this parameter as `metadata.annotations['huawei.com/npu']`.</li></ul></div></div>|
|`huawei.com/Ascend910`|Number|-|The number of NPUs requested. Modify it based on the actual situation.<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody">For Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD, configure this parameter as `metadata.annotations['huawei.com/npu']`.</div></div>Atlas 800 training server (fully populated with NPUs):<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4, 8</li><li>Distributed: 1, 2, 4, 8</li></ul>Atlas 800 training server (half populated with NPUs):<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4</li><li>Distributed: 1, 2, 4</li></ul>Server (with Atlas 300T training card):<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2</li><li>Distributed: 2</li></ul>Atlas 800T A2 training server and Atlas 900 A2 PoD cluster basic unit:<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul>Atlas 200T A2 Box16 heterogeneous subrack and Atlas 200I A2 Box16 heterogeneous subrack:<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li></ul>Atlas 900 A3 SuperPoD, A200T A3 Box8 SuperPoD server, and Atlas 800T A3 SuperPoD server:<ul><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4, 6, 8, 10, 12, 14, 16</li><li>Distributed: 2, 4, 6, 8, 10, 12, 14, 16</li><li>Logical SuperPoD affinity task for Atlas 900 A3 SuperPoD: 16</li></ul>Server (with Atlas 350 accelerator card) (8 processor in a node without interconnect):<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul>Server (with Atlas 350 accelerator cards) (16 processors in a node without interconnect):<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li></ul>Server (with Atlas 350 accelerator card) (4 processor-meshed, 8 processors):<ul><li>Single-node (affinity satisfied): 1, 2, 3, 4, 8</li><li>Single-node (affinity not guaranteed): 5, 6, 7</li><li>Distributed (affinity satisfied): 1, 2, 3, 4, 8</li><li>Distributed (affinity not guaranteed): 5, 6, 7</li></ul>Server (with Atlas 350 accelerator cards) (4 processor-meshed, 16 processors):<ul><li>Single-node (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li><li>Single-node (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li><li>Distributed (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li><li>Distributed (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li></ul>Atlas 650E server:<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul>Atlas 850E SuperPoD:<ul><li>Single-node: 1, 2, 4, 8 (the value of `sp-block` must be consistent with it)</li><li>Distributed: 8 (the value of `sp-block` must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not be greater than the physical SuperPoD size)</li></ul>Atlas 950 SuperPoD:<ul><li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8 (the value of `sp-block` must be consistent with it)</li><li>Distributed: 8 (the value of `sp-block` must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not be greater than the physical SuperPoD size)</li></ul>|
|`super-pod-affinity`|String|-|<p>This parameter can be used only on Atlas 900 A3 SuperPoD. It is the affinity scheduling policy used by SuperPoD jobs, and users need to declare it in the label of the YAML.</p><ul><li>`soft`: when cluster resources do not satisfy SuperPoD affinity, the job continues to be scheduled using fragmented resources in the cluster.</li><li>`hard`: when cluster resources do not satisfy SuperPoD affinity, the job remains Pending and waits for resources.</li><li>Other values or when this parameter is not passed: forced SuperPoD affinity scheduling.</li></ul>|
|<ul><li>`customJobKey`</li><li>`custom-job-id`</li></ul>|String|-|<p>Supports setting a unique job identifier through `customJobKey` or `custom-job-id`, making it convenient for users to filter key information such as alarms and ISSUEs related to the job based on this identifier. Set it in the `metadata.labels` label of the AscendJob resource.</p><ul><li>`customJobKey`: a user-defined label that sets the unique job identifier through a two-level jump, for example:<p>`customJobKey: tid`</p><p>`tid: "123456"`</p></li><li>`custom-job-id`: a user-defined label that directly sets the unique job identifier, for example:<p>`custom-job-id: "123456"`</p></li></ul>|
|`huawei.com/scheduler.softShareDev.aicoreQuota`|String|-|The requested AICore percentage. The value range is [1, 100].|
|`huawei.com/scheduler.softShareDev.hbmQuota`|String|-|<p>The requested amount of high-bandwidth memory. The value range is [1, maxHBM], in MB.</p><p>`maxHBM` is the `HBM` value in `HBM-Usage(MB)` queried using the <b>npu-smi info</b> command.</p>|
|`huawei.com/scheduler.softShareDev.policy`|String|-|<p>The soft partitioning policy. Supported values:</p><ul><li>`fixed-share`</li><li>`elastic`</li><li>`best-effort`</li></ul>|
|`podAffinity`|String|-|<p>Indicates that a logical SuperPoD is scheduled to a physical SuperPoD with more affinity Pods.</p><p>This parameter can be used only on Atlas 800I A3 SuperPoD servers for MindIE Motor inference jobs.</p>|
|`sp-fit`|String|-|<p>The SuperPoD scheduling policy. This parameter can be used only on Atlas 800I A3 SuperPoD server for MindIE Motor inference jobs.</p><ul><li>`idlest`: the logical SuperPoD is scheduled to a more idle physical SuperPoD.</li><li>`Non-idlest`: the logical SuperPoD preferentially fills up a physical SuperPoD.</li></ul>|
|`metadata.labels['duo']`|String|-|<p>A parameter supported only by inference servers (with Atlas 300I Duo inference cards).</p><ul><li>`true`: use the Atlas 300I Duo inference card.</li><li>`false`: do not use the Atlas 300I Duo inference card.</li></ul>|
|`metadata.labels['npu-310-strategy']`|String|-|<p>A parameter supported only by inference servers (with Atlas 300I Duo inference cards).</p><ul><li>`card`: schedule by inference card. The number of Ascend AI Processors requested does not exceed 2, and Ascend AI Processors on the same Atlas 300I Duo inference card are used.</li><li>`chip`: schedule by Ascend AI Processor. The number of Ascend AI Processors requested does not exceed the maximum value of a single node.</li></ul>|
|`metadata.labels['distributed']`|String|-|<p>Whether to use distributed inference. A parameter supported only by inference servers (with Atlas 300I Duo inference cards).</p><ul><li>`true`: use distributed inference. When the chip mode is used, the job must be scheduled to an entire Atlas 300I Duo inference card. If the number of Ascend AI Processors required by the job is an odd number, the part using a single Ascend AI Processor is preferentially scheduled to an Atlas 300I Duo inference card with 1 remaining Ascend AI Processor.</li><li>`false`: use non-distributed inference. When the chip mode is used, the number of Ascend AI Processors requested does not exceed the maximum value of a single node.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>Regardless of whether distributed inference is used, the scheduling policy of the `card` mode remains unchanged.</li><li>When `distributed` is `true`, only single-node multi-chip is supported; when `distributed` is `false`, only multi-node multi-chip is supported.</li><li>When `distributed` is `true`, deploy jobs are not supported.</li></ul></div></div>|

## vcjob YAML Parameter Description<a name="vcjob"></a>

The YAML parameters that can be used in a vcjob are described in the following table.

**Table 2** Key fields of a vcjob

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="22.58%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="40.86%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001609074269_p19658152414451"></a>spec.minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001609074269_ul1531417539259"><li>Single-node: 1</li><li>Distributed: N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001609074269_p11302326164814"></a>N is the number of nodes. This parameter is not required for deploy jobs. It is recommended that this parameter be consistent with replicas.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>spec.tasks[].replicas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>Standalone: 1</li><li>Distributed: N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N is the number of task replicas.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1458223119296"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p7582183112296"><a name="zh-cn_topic_0000001951418201_p7582183112296"></a><a name="zh-cn_topic_0000001951418201_p7582183112296"></a>spec.maxRetry</p>
<p id="zh-cn_topic_0000001951418201_p1758196165112"><a name="zh-cn_topic_0000001951418201_p1758196165112"></a><a name="zh-cn_topic_0000001951418201_p1758196165112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1026835111"><a name="zh-cn_topic_0000001951418201_p1026835111"></a><a name="zh-cn_topic_0000001951418201_p1026835111"></a>0&lt; maxRetry</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p2216813515"><a name="zh-cn_topic_0000001951418201_p2216813515"></a><a name="zh-cn_topic_0000001951418201_p2216813515"></a>Number of job rescheduling attempts. When a job fails, this is the number of times it can be rescheduled. When the number of rescheduling attempts already performed equals the value of maxRetry, the job will no longer be rescheduled.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note394611531302"><a name="zh-cn_topic_0000001951418201_note394611531302"></a><a name="zh-cn_topic_0000001951418201_note394611531302"></a><span class="notetitle">Note</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p0947553193013"><a name="zh-cn_topic_0000001951418201_p0947553193013"></a><a name="zh-cn_topic_0000001951418201_p0947553193013"></a>When both maxRetry and fault-retry-times are configured, the job will no longer be rescheduled once the number of rescheduling attempts already performed equals the value of either maxRetry or fault-retry-times.</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11581962517"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p13882123719515"><a name="zh-cn_topic_0000001951418201_p13882123719515"></a><a name="zh-cn_topic_0000001951418201_p13882123719515"></a>None (no maxRetry) or maxRetry equals 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1637895110"><a name="zh-cn_topic_0000001951418201_p1637895110"></a><a name="zh-cn_topic_0000001951418201_p1637895110"></a>If maxRetry is not configured or is set to 0, the system performs 3 rescheduling attempts by default.</p>
</td>
</tr>
<tr id="row917012162413"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p871672217415"><a name="p871672217415"></a><a name="p871672217415"></a>minReplicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p14170516144111"><a name="p14170516144111"></a><a name="p14170516144111"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p317081614417"><a name="p317081614417"></a><a name="p317081614417"></a>Minimum number of replicas. Set this to the minimum number of nodes required by the job.</p>
</td>
</tr>
<tr>
<td rowspan="2">metadata.labels.fault-scheduling</td>
<td>grace</td>
<td>Configures the job to use graceful deletion mode. During the process, the original Pod is gracefully deleted first. If the deletion is not successful after 15 minutes, the original Pod is forcibly deleted. For process-level rescheduling and process-level online recovery scenarios, set this parameter to grace.</td>
</tr>
<tr>
<td>force</td>
<td>Configures the job to use forced deletion mode, in which the original Pod is forcibly deleted during the process.</td>
</tr>
<tr id="row128861384219"><td class="cellrowborder" valign="top" width="27.21%" headers="mcps1.2.4.1.1 "><p id="p11288121310421"><a name="p11288121310421"></a><a name="p11288121310421"></a>metadata.labels.elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.230000000000004%" headers="mcps1.2.4.1.2 "><p id="p7288191354217"><a name="p7288191354217"></a><a name="p7288191354217"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1628816134422"><a name="p1628816134422"></a><a name="p1628816134422"></a>Enables elastic training.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>spec.tasks[0].template.spec.containers[0].image</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>Training image name. Modify it based on the actual situation (see the image name created in the Image Creation section).</p>
</td>
</tr>
<tr id="row319913141385"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p17879179384"><a name="p17879179384"></a><a name="p17879179384"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p11787717143811"><a name="p11787717143811"></a><a name="p11787717143811"></a>pod: Only Pod-level rescheduling is supported, and it is not upgraded to the Job level. (When vcjob is used, configure this policy: policies: -event:PodFailed -action:RestartTask)</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1278741713381"><a name="p1278741713381"></a><a name="p1278741713381"></a>Job rescheduling policy.</p>
</td>
</tr>
<tr id="row675991618389"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p778791715380"><a name="p778791715380"></a><a name="p778791715380"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1378781718388"><a name="p1378781718388"></a><a name="p1378781718388"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1378741712380"><a name="p1378741712380"></a><a name="p1378741712380"></a>Minimum number of replicas that can be scheduled for the job.</p>
</td>
</tr>
<tr id="row492051125013"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>Currently supports the configurations in <a href="#schedule_policy">huawei.com/schedule_policy Configuration Description</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>Configures the AI chip layout form that the job needs to schedule.<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>selects an appropriate scheduling policy based on this field.</p>
</td>
</tr>
<tr><td>servertype</td><td><ul><li>npu-{number of AI cores}</li><li>soc</li><li>Ascend910-{number of AI cores}</li><li>Ascend310P-{number of AI cores}</li></ul></td><td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p202093166576"><a name="zh-cn_topic_0000001609074213_p202093166576"></a><a name="zh-cn_topic_0000001609074213_p202093166576"></a>Server type.</p>
    <a name="zh-cn_topic_0000001609074213_ul87677178911"></a><a name="zh-cn_topic_0000001609074213_ul87677178911"></a><ul id="zh-cn_topic_0000001609074213_ul87677178911"><li>soc: Schedules to the <span id="zh-cn_topic_0000001609074213_ph126801133164916"><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a>Atlas 200I SoC A1 core board</span> node. This configuration must be added, and directory mounting must be performed by referring to the <span class="filepath" id="zh-cn_topic_0000001609074213_filepath127811055718"><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a>"infer-310p-1usoc.yaml"</span> file.</li><li>This parameter is not required for nodes of other types.</li></ul>
    </td></tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>metadata.annotations['sp-block']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>Specifies the number of chips in a logical SuperPoD.</p>
<a name="ul1150756144219"></a><a name="ul1150756144219"></a><ul id="ul1150756144219"><li>For a single node, it must be consistent with the number of chips requested by the job.</li><li>For a distributed deploy job, it must be an integer multiple of the number of chips per node, and the total number of chips requested by the job must be an integer multiple of it.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>Specifies the sp-block field. The cluster scheduling components divide the physical SuperPoD into logical SuperPoDs based on the splitting policy for affinity scheduling of training jobs.<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>If the user does not specify this field,</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>sets the logical SuperPoD size of this job to the total number of NPUs configured for the job during scheduling.</span></p>
<p id="p1250719624216"><a name="p1250719624216"></a><a name="p1250719624216"></a>For details, see <a href="../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod">UnifiedBus Device Node Network Description</a>.</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>This field can be used only on Atlas 900 A3 SuperPoD, Atlas 800T A3 SuperPoD server, Atlas 800I A3 SuperPoD server, Atlas 850E SuperPoD, and Atlas 950 SuperPoD.</li><li>After this field is used, the tor-affinity field does not need to be configured additionally.</li><li>FAQ: <a href="https://gitcode.com/Ascend/mind-cluster/issues/377">Total number of chips requested by the job is 32, sp-block set to 32 allows normal training, sp-block set to 16 fails to complete training, training container error report indicates initialization connection failed.</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>metadata.annotations['ra-block']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>Identifier for rack affinity scheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>Specifies the ra-block field. On the premise of supporting dynamic ratio configuration, 64 cards in a single rack are divided into 4 OSs, and each OS is considered a node in the K8s cluster. The intra-rack communication latency is lower than the inter-rack communication latency. This field is configured for rack affinity scheduling of training jobs.</p><p id="p175075613422"><a name="p175075613422"></a><a name="ul1150756144219"></a><a name="ul1150756144219"></a>The value range is 0 to 64 and must be a power of 2.</p><p id="p175075613422"><a name="p175075613422"></a><a name="ul1150756144219"></a><a name="ul1150756144219"></a>By enumeration, the value of ra-block is {1, 2, 4, 8, 16, 32, 64}.</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a>This field can be used only on Atlas 950 SuperPoD.
</div></div>
</td>
</tr>
<tr id="row862818313577"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p132726845716"><a name="p132726845716"></a><a name="p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="ul1427218195710"></a><a name="ul1427218195710"></a><ul id="ul1427218195710"><li>large-model-schema: large model task or padding task</li><li>normal-schema: normal task</li><li>null: switch affinity scheduling is not used<div class="note" id="note32586245294"><a name="note32586245294"></a><a name="note32586245294"></a><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><p id="p5258102462916"><a name="p5258102462916"></a><a name="p5258102462916"></a>Users need to select the task type based on the number of task replicas. If the number of task replicas is less than 4, it is a padding task. If the number of task replicas is greater than or equal to 4, it is a large model task. Normal tasks do not restrict the number of task replicas.</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p32732087577"><a name="p32732087577"></a><a name="p32732087577"></a>The default value is null, indicating that switch affinity scheduling is not used. Users need to configure it based on the task type.</p><ul id="ul961424647"><li>Switch affinity scheduling 1.0 supports <span id="ph63831524184110"><a name="ph63831524184110"></a><a name="ph63831524184110"></a><term>Atlas training products</term></span> and <span id="ph138318245414"><a name="ph138318245414"></a><a name="ph138318245414"></a><term id="zh-cn_topic_0000001519959665_term57208119917_4"><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a>Atlas A2 training products</term></span>; supports <span id="ph17383182419412"><a name="ph17383182419412"></a><a name="ph17383182419412"></a>PyTorch</span> and <span id="ph1383224134120"><a name="ph1383224134120"></a><a name="ph1383224134120"></a>MindSpore</span> frameworks.</li><li>Switch affinity scheduling 2.0 supports <span id="ph438320243412"><a name="ph438320243412"></a><a name="ph438320243412"></a><term id="zh-cn_topic_0000001519959665_term57208119917_5"><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a>Atlas A2 training products</term></span>; supports <span id="ph134821711841"><a name="ph134821711841"></a><a name="ph134821711841"></a>PyTorch</span> framework.</li><li>Switch affinity scheduling is supported only for full NPU. Static vNPU is not supported for switch affinity scheduling.</li></ul>
</td>
</tr>

<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>spec.tasks[0].template.spec.containers[0].resources.requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1996615912482"><a name="p1996615912482"></a><a name="p1996615912482"></a><strong id="b118963916494"><a name="b118963916494"></a><a name="b118963916494"></a>Full-NPU scheduling:</strong></p>
<ul><li>Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, Atlas 950 SuperPoD:<ul><li>Configure as huawei.com/npu: <em>x</em></li></ul></li><li>Inference server (with Atlas 300I inference card):<ul><li>Configure as huawei.com/Ascend310: <em>x</em></li></ul></li><li><term>Atlas inference products</term> non-mixed mode:<ul><li>Configure as huawei.com/Ascend310P: <em>x</em></li></ul></li><li><term>Atlas inference products</term> mixed mode:<ul><li>Configure as huawei.com/Ascend310P-V: <em>x</em></li><li>Configure as huawei.com/Ascend310P-VPro: <em>x</em></li><li>Configure as huawei.com/Ascend310P-IPro: <em>x</em></li></ul></li><li>For other products, configure as huawei.com/Ascend910: <em>x</em></li></ul>
<p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>Depending on the chip type used, the value of x is as follows:</p>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a><ul id="ul4403181216571"><li><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 training server (fully populated with NPUs)</span>:<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>Single-node, single-chip: 1</li><li>Single-node, multi-chip: 2, 4, 8</li><li>Distributed: 1, 2, 4, 8</li></ul>
</li><li><span id="zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 training server (half populated with NPUs)</span>:<a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001609074269_ul1713712328597"><li>Single-node, single-chip: 1</li><li>Single-node, multi-chip: 2, 4</li><li>Distributed: 1, 2, 4</li></ul>
</li><li>Server (with <span id="zh-cn_topic_0000001609074269_ph1223449506"><a name="zh-cn_topic_0000001609074269_ph1223449506"></a><a name="zh-cn_topic_0000001609074269_ph1223449506"></a>Atlas 300T training card</span>):<a name="ul3519194217372"></a><a name="ul3519194217372"></a><ul id="ul3519194217372"><li>Single-node, single-chip: 1</li><li>Single-node, multi-chip 2</li><li>Distributed: 2</li></ul>
</li><li><span id="ph1176216314557"><a name="ph1176216314557"></a><a name="ph1176216314557"></a>Atlas 800T A2 training server</span> and <span id="ph107421743105017"><a name="ph107421743105017"></a><a name="ph107421743105017"></a>Atlas 900 A2 PoD cluster basic unit</span>:<a name="ul169264817234"></a><a name="ul169264817234"></a><ul id="ul169264817234"><li>Single-node, single-chip: 1</li><li>Single-node, multi-chip: 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul>
</li><li><span id="ph129391532155719"><a name="ph129391532155719"></a><a name="ph129391532155719"></a>Atlas 200T A2 Box16 heterogeneous subrack</span>:<a name="ul555885820439"></a><a name="ul555885820439"></a><ul id="ul555885820439"><li>Single-node, single-chip: 1</li><li>Single-node, multi-chip: 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li></ul>
</li><li><span id="ph133001904447"><a name="ph133001904447"></a><a name="ph133001904447"></a>Atlas 900 A3 SuperPoD</span>, <span id="ph830011074420"><a name="ph830011074420"></a><a name="ph830011074420"></a>A200T A3 Box8 SuperPoD server</span>, <span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 SuperPoD server</span>:<a name="ul130020074415"></a><a name="ul130020074415"></a><ul id="ul130020074415"><li>Single-node, multi-chip: 2, 4, 6, 8, 10, 12, 14, 16</li><li>Distributed: 16</li></ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (8 processors in a node without interconnect)</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (16 processors in a node without interconnect)</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (4 processor-meshed, 8 processors)</span>:
    <ul>
        <li>Single-node (affinity satisfied): 1, 2, 3, 4, 8</li>
        <li>Single-node (affinity not guaranteed): 5, 6, 7</li>
        <li>Distributed (affinity satisfied): 1, 2, 3, 4, 8</li>
        <li>Distributed (affinity not guaranteed): 5, 6, 7</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (4 processor-meshed, 16 processors)</span>:
    <ul>
        <li>Single-node (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li>
        <li>Single-node (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li>
        <li>Distributed (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li>
        <li>Distributed (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li>
    </ul>
</li>
<li>
    <span>Atlas 650E server</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li>
    </ul>
</li>
<li>
    <span>Atlas 850E SuperPoD</span>:
    <ul>
        <li>Single-node: 1, 2, 4, 8 (the sp-block value must be consistent with it)</li>
        <li>Distributed: 8 (the sp-block value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not exceed the physical SuperPoD size)</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8 (the sp-block value must be consistent with it)</li>
        <li>Distributed: 8 (the sp-block value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not exceed the physical SuperPoD size)</li>
    </ul>
</li>
</ul>
<p id="p1498123034911"><a name="p1498123034911"></a><a name="p1498123034911"></a><strong id="b7488133134911"><a name="b7488133134911"></a><a name="b7488133134911"></a>Static vNPU scheduling:</strong></p>
<p id="p19104113195111"><a name="p19104113195111"></a><a name="p19104113195111"></a>huawei.com/Ascend910-<strong id="b14105734512"><a name="b14105734512"></a><a name="b14105734512"></a><em id="i17105533512"><a name="i17105533512"></a><a name="i17105533512"></a>Y</em></strong>: 1</p>
<p id="p1851116142917"><a name="p1851116142917"></a><a name="p1851116142917"></a>The value is 1. Only vNPU under one NPU can be used.</p>
<p id="p11413153312435"><a name="p11413153312435"></a><a name="p11413153312435"></a>For example, huawei.com/Ascend910-<em id="i94134332434"><a name="i94134332434"></a><a name="i94134332434"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5498134535310"><a name="p5498134535310"></a><a name="p5498134535310"></a>Type (only one type can be requested) and quantity of the requested NPU or vNPU. Modify them based on the actual situation.</p>
<ul id="ul10782193418818"><li>Only <span id="ph1038285416813"><a name="ph1038285416813"></a><a name="ph1038285416813"></a><term>Atlas inference products</term></span> in non-mixed-insertion mode supports static vNPU scheduling.</li><li>Inference servers (with <span id="ph1990710374611"><a name="ph1990710374611"></a><a name="ph1990710374611"></a>Atlas 300I inference cards</span>) and <span id="ph629210161695"><a name="ph629210161695"></a><a name="ph629210161695"></a><term>Atlas inference products</term></span> in mixed-insertion mode do not support static vNPU scheduling.</li><li>For the value of <strong id="b179331118122318"><a name="b179331118122318"></a><a name="b179331118122318"></a><em id="i14933131862318"><a name="i14933131862318"></a><a name="i14933131862318"></a>Y</em></strong>, refer to the "vNPU Type" column of the corresponding product in the virtualization instance template and vNPU type relationship table in the <a href="../04_usage/02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/02_mounting_vnpu_static.md#method-3-mounting-vnpus-on-kubernetes">Static Virtualization</a> section.<p id="p208621211164518"><a name="p208621211164518"></a><a name="p208621211164518"></a>For example, for the vNPU type <em id="i412654718449"><a name="i412654718449"></a><a name="i412654718449"></a>Ascend310P-4c.3cpu</em>, the value of <strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong> is 4c.3cpu, excluding the preceding Ascend310P.</p>
    </li></ul>
</td>
</tr>
<tr id="row25918533287"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p05117110298"><a name="p05117110298"></a><a name="p05117110298"></a>spec.tasks[0].template.spec.containers[0].resources.limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p13683185074711"><a name="p13683185074711"></a><a name="p13683185074711"></a>Limits the type (only one type can be requested) and quantity of the requested NPU or vNPU. Modify them based on the actual situation.</p>
<p id="p16683135019479"><a name="p16683135019479"></a><a name="p16683135019479"></a>The chip name and quantity in limits must be consistent with those in requests.</p>
</td>
</tr>
<tr id="row14747131720228"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/Ascend<em id="i103895254475"><a name="i103895254475"></a><a name="i103895254475"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX indicates the chip model. The supported values are 910, 310, and 310P. The value must be consistent with the actual chip type in the environment.</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 ">
    <p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span> obtains this parameter value and uses it to mount the corresponding type of NPU to the container.</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>In distributed jobs, ensure that the nodes running the training job have the same architecture.</p>
<div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle">Note</span>
    <div class="notebody">
        <ul>
            <li>
                <p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>This parameter supports only the full-NPU scheduling feature of the <span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span> scheduler. Users who use static vNPU scheduling or other schedulers need to delete the fields related to this parameter from the sample YAML.</p>
            </li>
            <li><p>For Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD, configure this parameter as metadata.annotations['huawei.com/npu'].</p></li>
        </ul>
</div>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"></a>spec.tasks[0].template.spec.containers[0].resources.{requests|limits}['huawei.com/Ascend910']</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"></a>The value varies depending on the chip type used:</p>
<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 training server (fully populated with NPUs)</span>: <a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4, 8</li><li>Distributed: 1, 2, 4, 8</li></ul>
</li><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 training server (half configured with NPUs)</span>: <a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4</li><li>Distributed: 1, 2, 4</li></ul>
</li><li><span id="ph157984201135"><a name="ph157984201135"></a><a name="ph157984201135"></a>Atlas 800T A2 training server</span> and <span id="zh-cn_topic_0000001951418201_ph745323894316"><a name="zh-cn_topic_0000001951418201_ph745323894316"></a><a name="zh-cn_topic_0000001951418201_ph745323894316"></a>Atlas 900 A2 PoD cluster basic unit</span><a name="zh-cn_topic_0000001951418201_ul169264817234"></a><a name="zh-cn_topic_0000001951418201_ul169264817234"></a><ul id="zh-cn_topic_0000001951418201_ul169264817234"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul>
</li><li><span id="zh-cn_topic_0000001951418201_ph419517625020"><a name="zh-cn_topic_0000001951418201_ph419517625020"></a><a name="zh-cn_topic_0000001951418201_ph419517625020"></a>Atlas 200T A2 Box16 heterogeneous subrack</span><span id="ph1891953184717"><a name="ph1891953184717"></a><a name="ph1891953184717"></a> and </span><span id="ph1149713543472"><a name="ph1149713543472"></a><a name="ph1149713543472"></a>Atlas 200I A2 Box16 heterogeneous subrack</span>: <a name="zh-cn_topic_0000001951418201_ul191955617509"></a><a name="zh-cn_topic_0000001951418201_ul191955617509"></a><ul id="zh-cn_topic_0000001951418201_ul191955617509"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li></ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (8 processors in a node without interconnect)</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (16 processors in a node without interconnect)</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (4 processor-meshed, 8 processors)</span>:
    <ul>
        <li>Single-node (affinity satisfied): 1, 2, 3, 4, 8</li>
        <li>Single-node (affinity not guaranteed): 5, 6, 7</li>
        <li>Distributed (affinity satisfied): 1, 2, 3, 4, 8</li>
        <li>Distributed (affinity not guaranteed): 5, 6, 7</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (4 processor-meshed, 16 processors)</span>:
    <ul>
        <li>Single-node (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li>
        <li>Single-node (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li>
        <li>Distributed (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li>
        <li>Distributed (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li>
    </ul>
</li>
<li>
    <span>Atlas 650E server</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li>
    </ul>
</li>
<li>
    <span>Atlas 850E SuperPoD</span>:
    <ul>
        <li>Single-node: 1, 2, 4, 8 (the sp-block value must be consistent with it)</li>
        <li>Distributed: 8 (the sp-block value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not exceed the physical SuperPoD size)</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8 (the sp-block value must be consistent with it)</li>
        <li>Distributed: 8 (the sp-block value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not exceed the physical SuperPoD size)</li>
    </ul>
</li>
</ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"></a>Number of NPUs requested. Modify it based on the actual situation. When requesting whole cards, you cannot request vNPUs at the same time.</p>
<div class="note" id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"></a><span class="notetitle">Note</span>
    <div class="notebody"><a name="zh-cn_topic_0000001951418201_ul54321224184319"></a><a name="zh-cn_topic_0000001951418201_ul54321224184319"></a>
        <ul id="zh-cn_topic_0000001951418201_ul54321224184319">
            <li>
                <strong id="zh-cn_topic_0000001951418201_b16213840172320"><a name="zh-cn_topic_0000001951418201_b16213840172320"></a><a name="zh-cn_topic_0000001951418201_b16213840172320"></a>Graceful fault tolerance mode</strong> supports <span id="zh-cn_topic_0000001951418201_ph158146714142"><a name="zh-cn_topic_0000001951418201_ph158146714142"></a><a name="zh-cn_topic_0000001951418201_ph158146714142"></a>Atlas 800 training server</span>, and the resource request quantity can only be 4N or 8N, where N is the number of training nodes.
            </li>
            <li>
                <strong id="zh-cn_topic_0000001951418201_b1091614581433"><a name="zh-cn_topic_0000001951418201_b1091614581433"></a><a name="zh-cn_topic_0000001951418201_b1091614581433"></a>Graceful fault tolerance mode</strong> supports <span id="ph184881417142314"><a name="ph184881417142314"></a><a name="ph184881417142314"></a>Atlas 800T A2 training server</span> or <span id="zh-cn_topic_0000001951418201_ph9246916444"><a name="zh-cn_topic_0000001951418201_ph9246916444"></a><a name="zh-cn_topic_0000001951418201_ph9246916444"></a>Atlas 900 A2 PoD cluster basic unit</span>, and the resource request quantity can only be 8N, where N is the number of training nodes.
            </li>
            <li>
                <p>For Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD, change the parameter name to huawei.com/npu.</p>
            </li>
        </ul>
    </div>
</div>
</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>{metadata, spec.tasks[0].template.metadata}.labels['ring-controller.atlas']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1941725316543"><a name="p1941725316543"></a><a name="p1941725316543"></a>The value varies depending on the chip type used:</p>
<a name="ul2750122165318"></a><a name="ul2750122165318"></a><ul id="ul2750122165318"><li>Inference server (with <span id="ph3690191194813"><a name="ph3690191194813"></a><a name="ph3690191194813"></a>Atlas 300I inference card</span>): ascend-310</li><li><span id="ph56912120486"><a name="ph56912120486"></a><a name="ph56912120486"></a><term>Atlas inference products</term></span>: ascend-310P</li><li>Atlas 800 training server, server (with <span id="ph6581133055411"><a name="ph6581133055411"></a><a name="ph6581133055411"></a>Atlas 300T training card</span>): ascend-910</li><li><span id="ph10656173717129"><a name="ph10656173717129"></a><a name="ph10656173717129"></a><term id="zh-cn_topic_0000001519959665_term57208119917_6"><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a>Atlas A2 training products</term></span>, <span id="ph1665620377128"><a name="ph1665620377128"></a><a name="ph1665620377128"></a>A200T A3 Box8 SuperPoD server</span>, <span id="ph14656337131215"><a name="ph14656337131215"></a><a name="ph14656337131215"></a>Atlas 900 A3 SuperPoD</span>, <span id="ph12656113717123"><a name="ph12656113717123"></a><a name="ph12656113717123"></a>Atlas 800T A3 SuperPoD server</span>: ascend-<span id="ph1265633714121"><a name="ph1265633714121"></a><a name="ph1265633714121"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b</li><li>(Optional) Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, Atlas 950 SuperPoD: ascend-npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>Used to distinguish the chip type used by the job. It needs to be configured in both the <span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span> and the job.</p>
<div class="note" id="note14282027593"><a name="note14282027593"></a><a name="note14282027593"></a><span class="notetitle">Note</span><div class="notebody"><p id="p1328162720912"><a name="p1328162720912"></a><a name="p1328162720912"></a><span id="ph19729197"><a name="ph19729197"></a><a name="ph19729197"></a>Here, {<em id="zh-cn_topic_0000001519959665_i1914312018209_2"><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a>xxx</em>} refers to using "910" as the chip model value.</span></p>
</div></div>
</td>
</tr>
<tr id="row141124616406"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>Affinity scheduling policy used by SuperPoD jobs. It needs to be declared in the label of the YAML.</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft: When cluster resources do not satisfy SuperPoD affinity, the job continues to be scheduled using fragmented resources in the cluster.</li><li>hard: When cluster resources do not satisfy SuperPoD affinity, the job remains Pending and waits for resources.</li><li>Other values or when this parameter is not passed: forced SuperPoD affinity scheduling</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>This parameter can be used only in the <span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD</span>.</p>
</td>
</tr>
<tr id="rowcustomjobkey2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobkey2"><a name="pcustomjobkey2"></a><a name="pcustomjobkey2"></a>customJobKey</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobkeyvalue2"><a name="pcustomjobkeyvalue2"></a><a name="pcustomjobkeyvalue2"></a>User-defined label that sets the unique job identifier through a two-level reference, for example:<br> customJobKey: tid<br> tid: "123456"</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="pcustomjobkeydesc2"><a name="pcustomjobkeydesc2"></a><a name="pcustomjobkeydesc2"></a>Supports setting the unique job identifier through customJobKey or custom-job-id, making it convenient for users to filter key information such as alarms and ISSUEs related to the job based on this identifier.<br> <ul><li>For vcjob, set it in the metadata.labels label of the Job resource.<br></li> <li>For deploy jobs, set it in the spec.template.metadata.labels label of the Deployment resource.</li></ul></p>
</td>
</tr>
<tr id="rowcustomjobid2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobid2"><a name="pcustomjobid2"></a><a name="pcustomjobid2"></a>custom-job-id</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobidvalue2"><a name="pcustomjobidvalue2"></a><a name="pcustomjobidvalue2"></a>User-defined label that directly sets the unique job identifier, for example:<br> custom-job-id: "123456"</p>
</td>
</tr>
<tr id="row136201528182116"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p56210289215"><a name="p56210289215"></a><a name="p56210289215"></a>spec.tasks[0].template.metadata.labels['vnpu-level']</p>
    <p id="p262172815213"><a name="p262172815213"></a><a name="p262172815213"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p562182842111"><a name="p562182842111"></a><a name="p562182842111"></a>low</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662112892120"><a name="p662112892120"></a><a name="p662112892120"></a>Low configuration, the default value. Selects the virtualized instance template with the lowest configuration.</p>
    </td>
    </tr>
    <tr id="row196219286214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p146219285218"><a name="p146219285218"></a><a name="p146219285218"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p19621528112118"><a name="p19621528112118"></a><a name="p19621528112118"></a>Performance first.</p>
    <p id="p6621152812214"><a name="p6621152812214"></a><a name="p6621152812214"></a>When cluster resources are sufficient, the virtualized instance template with the highest possible configuration is selected. When the overall cluster resources have been excessively used, for example, when most physical NPUs are already in use and each physical NPU has only a small portion of AICores remaining, which is insufficient to satisfy the high-configuration virtualized instance template, a lower-configuration template with the same number of AICores is used instead. For specific selection, refer to the <a href="../04_usage/02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md">Virtualization Templates</a> section.</p>
    </td>
    </tr>
    <tr id="row1762192862114"><td class="cellrowborder" rowspan="3" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p462112842110"><a name="p462112842110"></a><a name="p462112842110"></a>spec.tasks[0].template.metadata.labels['vnpu-dvpp']</p>
    <p id="p362120286216"><a name="p362120286216"></a><a name="p362120286216"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621122816219"><a name="p8621122816219"></a><a name="p8621122816219"></a>yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662162819213"><a name="p662162819213"></a><a name="p662162819213"></a>The <span id="ph1762113285210"><a name="ph1762113285210"></a><a name="ph1762113285210"></a>Pod</span> uses DVPP.</p>
    </td>
    </tr>
    <tr id="row1762172862117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p46214285213"><a name="p46214285213"></a><a name="p46214285213"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5621162812213"><a name="p5621162812213"></a><a name="p5621162812213"></a>The <span id="ph1362102815215"><a name="ph1362102815215"></a><a name="ph1362102815215"></a>Pod</span> does not use DVPP.</p>
    </td>
    </tr>
    <tr id="row1262122852117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p462192852111"><a name="p462192852111"></a><a name="p462192852111"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11621102818211"><a name="p11621102818211"></a><a name="p11621102818211"></a>Default value. Whether DVPP is used is not concerned.</p>
    </td>
    </tr>
<tr id="zh-cn_topic_0000001951418201_row4635558201210"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1499116019135"><a name="zh-cn_topic_0000001951418201_p1499116019135"></a><a name="zh-cn_topic_0000001951418201_p1499116019135"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p599118017133"><a name="zh-cn_topic_0000001951418201_p599118017133"></a><a name="zh-cn_topic_0000001951418201_p599118017133"></a>Recovery strategies available for the job.</p>
<a name="zh-cn_topic_0000001951418201_ul139911803137"></a><a name="zh-cn_topic_0000001951418201_ul139911803137"></a><ul id="zh-cn_topic_0000001951418201_ul139911803137"><li>retry: process-level online recovery.</li><li>recover: process-level rescheduling.</li><li>recover-in-place: process-level in-place recovery.</li><li>dump: save dying gasp words.</li><li>exit: exit training.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul169911906135"></a><a name="zh-cn_topic_0000001951418201_ul169911906135"></a>recover-strategy is configured under the annotations of the job YAML. Its value is any combination of the five strategies, separated by commas.
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row10152132415157"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10821192541514"><a name="zh-cn_topic_0000001951418201_p10821192541514"></a><a name="zh-cn_topic_0000001951418201_p10821192541514"></a>metadata.labels.pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><ul id="zh-cn_topic_0000001951418201_ul5821162501510"><li>on: Enables Pod-level rescheduling.</li><li>Other values or when this field is not used: Disables Pod-level rescheduling.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p78221125201514"><a name="zh-cn_topic_0000001951418201_p78221125201514"></a><a name="zh-cn_topic_0000001951418201_p78221125201514"></a>Pod-level rescheduling means that after a job fault occurs, not all Pods are deleted. Instead, only the faulty Pod is deleted, and a new Pod is created for rescheduling.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note5822925151516"><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><ul id="zh-cn_topic_0000001951418201_ul17822112517158"><li>The default rescheduling mode is Job-level rescheduling. To enable Pod-level rescheduling, add this field.</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row576132216324"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1772202423212"><a name="zh-cn_topic_0000001951418201_p1772202423212"></a><a name="zh-cn_topic_0000001951418201_p1772202423212"></a>metadata.labels.subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><ul id="zh-cn_topic_0000001951418201_ul972624133214"><li>ignore: Ignores the sub-healthy node. Subsequent jobs do not prioritize this node in affinity scheduling.</li><li>graceExit: Does not use the sub-healthy node, saves the dying gasp checkpoint file, and then performs rescheduling. Subsequent jobs are not scheduled to this node.</li><li>forceExit: Does not use the sub-healthy node, exits the job directly without saving, and then performs rescheduling. Subsequent jobs are not scheduled to this node.</li><li>The default value is ignore.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1973102463218"><a name="zh-cn_topic_0000001951418201_p1973102463218"></a><a name="zh-cn_topic_0000001951418201_p1973102463218"></a>Processing strategy for nodes whose status is sub-healthy (SubHealthy).</p>
<div class="note" id="zh-cn_topic_0000001951418201_note173271703519"><a name="zh-cn_topic_0000001951418201_note173271703519"></a><a name="zh-cn_topic_0000001951418201_note173271703519"></a><span class="notetitle">Note</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p163271901355"><a name="zh-cn_topic_0000001951418201_p163271901355"></a><a name="zh-cn_topic_0000001951418201_p163271901355"></a>When using the graceExit strategy, ensure that the dying gasp checkpoint saving function is enabled for the job.</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1314311835012"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p123205151739"><a name="zh-cn_topic_0000001951418201_p123205151739"></a><a name="zh-cn_topic_0000001951418201_p123205151739"></a>metadata.labels.fault-retry-times</p>
<p id="zh-cn_topic_0000001951418201_p196969196112"><a name="zh-cn_topic_0000001951418201_p196969196112"></a><a name="zh-cn_topic_0000001951418201_p196969196112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1192310597344"><a name="zh-cn_topic_0000001951418201_p1192310597344"></a><a name="zh-cn_topic_0000001951418201_p1192310597344"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p109232597342"><a name="zh-cn_topic_0000001951418201_p109232597342"></a><a name="zh-cn_topic_0000001951418201_p109232597342"></a>To handle service-plane faults, you must configure the number of times the service plane can be retried unconditionally.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note15571815115017"><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><ul id="zh-cn_topic_0000001951418201_ul15238182410364"><li>To use the unconditional retry function, ensure that an abnormal training process causes the container to exit abnormally. If the container does not exit abnormally, the retry cannot succeed.</li><li>Currently, only the <span id="ph1377171612516"><a name="ph1377171612516"></a><a name="ph1377171612516"></a>Atlas 800T A2 training server</span> and the <span id="zh-cn_topic_0000001951418201_ph14104952376"><a name="zh-cn_topic_0000001951418201_ph14104952376"></a><a name="zh-cn_topic_0000001951418201_ph14104952376"></a>Atlas 900 A2 PoD cluster basic unit</span> support the unconditional retry function.</li><li>When process-level recovery is performed, a service-plane fault is triggered. To use process-level recovery, you must configure this parameter.</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row260912190502"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2966613113520"><a name="zh-cn_topic_0000001951418201_p2966613113520"></a><a name="zh-cn_topic_0000001951418201_p2966613113520"></a>None (no fault-retry-times) or 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p2096618130353"><a name="zh-cn_topic_0000001951418201_p2096618130353"></a><a name="zh-cn_topic_0000001951418201_p2096618130353"></a>This job does not use the unconditional retry feature and cannot detect service-plane faults. The maxRetry setting of vcjob still takes effect.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row128551542131510"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10285161985210"><a name="zh-cn_topic_0000001951418201_p10285161985210"></a><a name="zh-cn_topic_0000001951418201_p10285161985210"></a>spec.policies</p>
<p id="zh-cn_topic_0000001951418201_p490916512164"><a name="zh-cn_topic_0000001951418201_p490916512164"></a><a name="zh-cn_topic_0000001951418201_p490916512164"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p056810252162"><a name="zh-cn_topic_0000001951418201_p056810252162"></a><a name="zh-cn_topic_0000001951418201_p056810252162"></a>event, with the following values:</p>
<a name="zh-cn_topic_0000001951418201_ul1781384818238"></a><a name="zh-cn_topic_0000001951418201_ul1781384818238"></a><ul id="zh-cn_topic_0000001951418201_ul1781384818238"><li>PodFailed: The Pod fails.</li><li>PodEvicted: The Pod is evicted.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p180717598243"><a name="zh-cn_topic_0000001951418201_p180717598243"></a><a name="zh-cn_topic_0000001951418201_p180717598243"></a>Pod status. Used together with the action field to indicate the processing policy of <span id="zh-cn_topic_0000001951418201_ph525518226126"><a name="zh-cn_topic_0000001951418201_ph525518226126"></a><a name="zh-cn_topic_0000001951418201_ph525518226126"></a>Volcano</span> when a Pod is in a certain status. The default value is PodEvicted.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1390814541612"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1590911581611"><a name="zh-cn_topic_0000001951418201_p1590911581611"></a><a name="zh-cn_topic_0000001951418201_p1590911581611"></a>action, with the following values:</p>
<a name="zh-cn_topic_0000001951418201_ul17824133752420"></a><a name="zh-cn_topic_0000001951418201_ul17824133752420"></a><ul id="zh-cn_topic_0000001951418201_ul17824133752420"><li>RestartJob: Restarts the training task.</li><li>Ignore: <span id="zh-cn_topic_0000001951418201_ph141051824104819"><a name="zh-cn_topic_0000001951418201_ph141051824104819"></a><a name="zh-cn_topic_0000001951418201_ph141051824104819"></a>Ignore. The open-source Volcano</span> does not perform any processing, and the <span id="zh-cn_topic_0000001951418201_ph631119334409"><a name="zh-cn_topic_0000001951418201_ph631119334409"></a><a name="zh-cn_topic_0000001951418201_ph631119334409"></a>Ascend-volcano-plugin</span> performs the processing.</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p20698124111427"><a name="zh-cn_topic_0000001951418201_p20698124111427"></a><a name="zh-cn_topic_0000001951418201_p20698124111427"></a>The processing policy of <span id="zh-cn_topic_0000001951418201_ph10699341154214"><a name="zh-cn_topic_0000001951418201_ph10699341154214"></a><a name="zh-cn_topic_0000001951418201_ph10699341154214"></a>Volcano</span> for Pods in a certain status. The default value is RestartJob.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note128691230174312"><a name="zh-cn_topic_0000001951418201_note128691230174312"></a><a name="zh-cn_topic_0000001951418201_note128691230174312"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul13166894452"></a><a name="zh-cn_topic_0000001951418201_ul13166894452"></a><ul id="zh-cn_topic_0000001951418201_ul13166894452"><li>To enable Pod-level rescheduling, delete policies and its sub-parameters event and action.</li><li>When using unconditional retry for service-plane faults (or using both Pod-level rescheduling and unconditional retry for service-plane faults), set event to PodFailed and action to Ignore.</li><li>If the cluster scheduling components <span id="zh-cn_topic_0000001951418201_ph8224175173014"><a name="zh-cn_topic_0000001951418201_ph8224175173014"></a><a name="zh-cn_topic_0000001951418201_ph8224175173014"></a>Volcano</span> are not used, or the open-source <span id="zh-cn_topic_0000001951418201_ph83286473313"><a name="zh-cn_topic_0000001951418201_ph83286473313"></a><a name="zh-cn_topic_0000001951418201_ph83286473313"></a>Volcano</span> does not integrate <span id="zh-cn_topic_0000001951418201_ph617718254597"><a name="zh-cn_topic_0000001951418201_ph617718254597"></a><a name="zh-cn_topic_0000001951418201_ph617718254597"></a>Ascend-volcano-plugin</span>, refer to <a href="https://gitcode.com/Ascend/mind-cluster/issues/362">In the scenario where Volcano and Ascend Operator are used, the status of all Pods of a job with a service-plane fault changes to Failed, and the job cannot trigger unconditional retry rescheduling</a> to modify the open-source Volcano code.</li><li>The open-source Volcano also provides other values for policies. It is not recommended that users change them to other values, as this may affect the normal use of the resumable training feature.</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11217021145014"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1929464718814"><a name="zh-cn_topic_0000001951418201_p1929464718814"></a><a name="zh-cn_topic_0000001951418201_p1929464718814"></a>spec.tasks[0].template.spec.restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><ul id="zh-cn_topic_0000001951418201_ul193373071216"><li>Never: Never restart.</li><li>Always: Always restart.</li><li>OnFailure: Restart on failure.</li><li><del>ExitCode: Determine whether to restart the Pod based on the process exit code. The Pod is not restarted when the error code is 1 to 127, and is restarted when the error code is 128 to 255.</del><div class="note" id="zh-cn_topic_0000001951418201_note278954373014"><a name="zh-cn_topic_0000001951418201_note278954373014"></a><a name="zh-cn_topic_0000001951418201_note278954373014"></a><span class="notetitle">Note</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p14789194311309"><a name="zh-cn_topic_0000001951418201_p14789194311309"></a><a name="zh-cn_topic_0000001951418201_p14789194311309"></a>ExitCode is not supported for training vcjob.</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1129434710811"><a name="zh-cn_topic_0000001951418201_p1129434710811"></a><a name="zh-cn_topic_0000001951418201_p1129434710811"></a>Container restart policy. When unconditional retry for service-plane faults is configured, the container restart policy must be set to <span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue182751614652"><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a>"Never"</span>.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1116371844811"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a>spec.tasks[0].template.spec.terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a>0 &lt; terminationGracePeriodSeconds &lt;<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a> grace-over-time</strong> parameter value</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a>The time elapsed from when the container receives SIGTERM to when it is forcibly stopped by <span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a>K8s</span>. This value must be greater than 0 and less than the value of the "<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a>grace-over-time</strong>" parameter in the volcano-v<em id="zh-cn_topic_0000001951418201_i1645121221719"><a name="zh-cn_topic_0000001951418201_i1645121221719"></a><a name="zh-cn_topic_0000001951418201_i1645121221719"></a>{version}</em>.yaml file. In addition, ensure that CKPT files can be saved. Modify the value based on the actual situation. For details, see <a href="https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">Container Lifecycle Hooks</a> on the official <span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a>K8s</span> website.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note17641176363"><a name="zh-cn_topic_0000001951418201_note17641176363"></a><a name="zh-cn_topic_0000001951418201_note17641176363"></a><div class="notebody"><p id="zh-cn_topic_0000001951418201_p97641517103616"><a name="zh-cn_topic_0000001951418201_p97641517103616"></a><a name="zh-cn_topic_0000001951418201_p97641517103616"></a>This field takes effect only when fault-scheduling is set to grace. When fault-scheduling is set to force, this field is invalid.</p>
</div></div>
</td>
</tr>
</tbody>
</table>

## deploy Job YAML Parameter Description<a name="deploy"></a>

The YAML parameters that can be used in a deploy job are described in the following table.

**Table 3** Key fields of a deploy job

<a name="zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="22.58%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001609074269_p13658124194513"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="40.86%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001609074269_p4658152420459"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001609074269_p8302202619484"></a>Description</p>
</th>
</tr>
</thead>
<tbody>
<tr id="zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001609074269_p5658142413455"></a>spec.replicas</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001609074269_ul122461585257"><li>Single-node: 1</li><li>Distributed: N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001609074269_p3302102644813"></a>N is the number of task replicas.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001609074269_p12658132454515"></a>spec.template.spec.containers[0].image</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001609074269_p1930210269483"></a>Training image name. Modify it based on the actual situation (see the image name created in the Image Creation section).</p>
</td>
</tr>
<tr>
<td rowspan="2">spec.template.metadata.labels['fault-scheduling']</td>
<td>grace</td>
<td>Configures the job to use graceful deletion mode. During the process, the original Pod is gracefully deleted first. If the deletion is not successful after 15 minutes, the original Pod is forcibly deleted.</td>
</tr>
<tr>
<td>force</td>
<td>Configures the job to use forced deletion mode. During the process, the original Pod is forcibly deleted.</td>
</tr>
<tr id="row171754462391"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p15220101916253"><a name="p15220101916253"></a><a name="p15220101916253"></a>{metadata, spec.template.metadata}.labels['ring-controller.atlas']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1941725316543"><a name="p1941725316543"></a><a name="p1941725316543"></a>The value varies depending on the chip type used:</p>
<a name="ul2750122165318"></a><a name="ul2750122165318"></a><ul id="ul2750122165318"><li>Inference server (with <span id="ph3690191194813"><a name="ph3690191194813"></a><a name="ph3690191194813"></a>Atlas 300I inference card</span>): ascend-310</li><li><span id="ph56912120486"><a name="ph56912120486"></a><a name="ph56912120486"></a><term>Atlas inference products</term></span>: ascend-310P</li><li>Atlas 800 training server, server (with <span id="ph6581133055411"><a name="ph6581133055411"></a><a name="ph6581133055411"></a>Atlas 300T training card</span>): ascend-910</li><li><span id="ph10656173717129"><a name="ph10656173717129"></a><a name="ph10656173717129"></a><term id="zh-cn_topic_0000001519959665_term57208119917_6"><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a><a name="zh-cn_topic_0000001519959665_term57208119917_6"></a>Atlas A2 training products</term></span>, <span id="ph1665620377128"><a name="ph1665620377128"></a><a name="ph1665620377128"></a>A200T A3 Box8 SuperPoD server</span>, <span id="ph14656337131215"><a name="ph14656337131215"></a><a name="ph14656337131215"></a>Atlas 900 A3 SuperPoD</span>, <span id="ph12656113717123"><a name="ph12656113717123"></a><a name="ph12656113717123"></a>Atlas 800T A3 SuperPoD server</span>: ascend-<span id="ph1265633714121"><a name="ph1265633714121"></a><a name="ph1265633714121"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_5"><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_5"></a>{xxx}</em></span>b</li><li>(Optional) Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, Atlas 950 SuperPoD: ascend-npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p19220131902512"><a name="p19220131902512"></a><a name="p19220131902512"></a>Used to distinguish the chip type used by the job. It must be configured in both the <span id="ph12290749162911"><a name="ph12290749162911"></a><a name="ph12290749162911"></a>ConfigMap</span> and task.</p>
<div class="note" id="note14282027593"><a name="note14282027593"></a><a name="note14282027593"></a><span class="notetitle">Note</span><div class="notebody"><p id="p1328162720912"><a name="p1328162720912"></a><a name="p1328162720912"></a><span id="ph19729197"><a name="ph19729197"></a><a name="ph19729197"></a>The {<em id="zh-cn_topic_0000001519959665_i1914312018209_2"><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a><a name="zh-cn_topic_0000001519959665_i1914312018209_2"></a>xxx</em>} here takes "910" as the chip model.</span></p>
</div></div>
</td>
</tr>
<tr id="row136201528182116"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p56210289215"><a name="p56210289215"></a><a name="p56210289215"></a>spec.template.metadata.labels['vnpu-level']</p>
    <p id="p262172815213"><a name="p262172815213"></a><a name="p262172815213"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p562182842111"><a name="p562182842111"></a><a name="p562182842111"></a>low</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662112892120"><a name="p662112892120"></a><a name="p662112892120"></a>Low configuration, the default value. Selects the virtualized instance template with the lowest configuration.</p>
    </td>
    </tr>
    <tr id="row196219286214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p146219285218"><a name="p146219285218"></a><a name="p146219285218"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p19621528112118"><a name="p19621528112118"></a><a name="p19621528112118"></a>Performance first.</p>
    <p id="p6621152812214"><a name="p6621152812214"></a><a name="p6621152812214"></a>When cluster resources are sufficient, the virtualized instance template with the highest possible configuration is selected. When the overall cluster resources are already heavily used, for example, when most physical NPUs are already in use and each physical NPU has only a small portion of AI Cores remaining, which is insufficient to satisfy the high-configuration virtualized instance template, a lower-configuration template with the same number of AI Cores is used instead. For specific selection, refer to the <a href="../04_usage/02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md">Virtualization Templates</a> chapter.</p>
    </td>
    </tr>
    <tr id="row1762192862114"><td class="cellrowborder" rowspan="3" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p462112842110"><a name="p462112842110"></a><a name="p462112842110"></a>spec.template.metadata.labels['vnpu-dvpp']</p>
    <p id="p362120286216"><a name="p362120286216"></a><a name="p362120286216"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621122816219"><a name="p8621122816219"></a><a name="p8621122816219"></a>yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662162819213"><a name="p662162819213"></a><a name="p662162819213"></a>This <span id="ph1762113285210"><a name="ph1762113285210"></a><a name="ph1762113285210"></a>Pod</span> uses DVPP.</p>
    </td>
    </tr>
    <tr id="row1762172862117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p46214285213"><a name="p46214285213"></a><a name="p46214285213"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5621162812213"><a name="p5621162812213"></a><a name="p5621162812213"></a>This <span id="ph1362102815215"><a name="ph1362102815215"></a><a name="ph1362102815215"></a>Pod</span> does not use DVPP.</p>
    </td>
    </tr>
    <tr id="row1262122852117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p462192852111"><a name="p462192852111"></a><a name="p462192852111"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11621102818211"><a name="p11621102818211"></a><a name="p11621102818211"></a>Default value. Whether DVPP is used is not a concern.</p>
    </td>
    </tr>
<tr id="row33457612918"><td class="cellrowborder" valign="top" width="26.12261226122612%" headers="mcps1.2.4.1.1 "><p id="p137845934610"><a name="p137845934610"></a><a name="p137845934610"></a>spec.template.metadata.labels['npu-310-strategy']</p>
</td>
<td class="cellrowborder" valign="top" width="36.16361636163616%" headers="mcps1.2.4.1.2 "><a name="ul1967514291118"></a><a name="ul1967514291118"></a><p>Parameter supported only by inference servers (with Atlas 300I inference cards installed)</p><ul id="ul1967514291118"><li>card: Schedules by inference card. The number of <span id="ph1978781173013"><a name="ph1978781173013"></a><a name="ph1978781173013"></a>Ascend AI Processors</span> requested does not exceed 4, and the <span id="ph933971152917"><a name="ph933971152917"></a><a name="ph933971152917"></a>Ascend AI Processors</span> on the same <span id="ph77331623132919"><a name="ph77331623132919"></a><a name="ph77331623132919"></a>Atlas 300I inference card</span> are used.</li><li>chip: Schedules by <span id="ph14705121219305"><a name="ph14705121219305"></a><a name="ph14705121219305"></a>Ascend AI Processor</span>. The number of chips requested does not exceed the maximum value of a single node.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="p2799246194816"><a name="p2799246194816"></a><a name="p2799246194816"></a>-</p>
</td>
</tr>
<tr id="row319913141385"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p17879179384"><a name="p17879179384"></a><a name="p17879179384"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p11787717143811"><a name="p11787717143811"></a><a name="p11787717143811"></a>pod: Only Pod-level rescheduling is supported, and it is not escalated to the Job level. (When vcjob is used, configure this policy: policies: -event:PodFailed -action:RestartTask)</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>It must be written into the annotation or label of the pod.</p><p id="p1278741713381"><a name="p1278741713381"></a><a name="p1278741713381"></a>Job rescheduling policy.</p>
</td>
</tr>
<tr id="row675991618389"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p778791715380"><a name="p778791715380"></a><a name="p778791715380"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1378781718388"><a name="p1378781718388"></a><a name="p1378781718388"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>It needs to be written into the annotation or label of the pod.</p><p id="p1378741712380"><a name="p1378741712380"></a><a name="p1378741712380"></a>Minimum number of replicas that the task can schedule.</p>
</td>
</tr>
<tr id="row492051125013"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p1430323175013"><a name="p1430323175013"></a><a name="p1430323175013"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p930320315500"><a name="p930320315500"></a><a name="p930320315500"></a>Currently supports the configurations in <a href="#schedule_policy">huawei.com/schedule_policy Configuration Description</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p>It needs to be written into the annotation or label of the pod.</p><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>Configures the AI chip layout form that the task needs to schedule. <span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span> selects an appropriate scheduling policy based on this field.</p>
</td>
</tr>
<tr><td>spec.template.spec.nodeSelector['servertype']</td>
<td><ul><li>npu-{number of AI cores}</li><li>soc</li><li>Ascend910-{number of AI cores}</li><li>Ascend310P-{number of AI cores}</li></ul></td>
<td class="cellrowborder" valign="top" width="37.71377137713771%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001609074213_p202093166576"><a name="zh-cn_topic_0000001609074213_p202093166576"></a><a name="zh-cn_topic_0000001609074213_p202093166576"></a>Server type.</p>
    <a name="zh-cn_topic_0000001609074213_ul87677178911"></a><a name="zh-cn_topic_0000001609074213_ul87677178911"></a><ul id="zh-cn_topic_0000001609074213_ul87677178911"><li>soc: Schedules to the <span id="zh-cn_topic_0000001609074213_ph126801133164916"><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a><a name="zh-cn_topic_0000001609074213_ph126801133164916"></a>Atlas 200I SoC A1 core board</span> node. This configuration must be added, and directory mounting must be performed by referring to the <span class="filepath" id="zh-cn_topic_0000001609074213_filepath127811055718"><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a><a name="zh-cn_topic_0000001609074213_filepath127811055718"></a>"infer-310p-1usoc.yaml"</span> file.</li><li>This parameter is not required for nodes of other types.</li></ul>
    </td></tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>metadata.annotations['sp-block']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>Specifies the number of chips in a logical SuperPoD.</p>
<a name="ul1150756144219"></a><a name="ul1150756144219"></a><ul id="ul1150756144219"><li>For a single node, it must be consistent with the number of chips requested by the job.</li><li>For distributed deploy job, it must be an integer multiple of the number of chips per node, and the total number of chips requested by the job must be an integer multiple of it.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>Specify the sp-block field, and the cluster scheduling component divides the physical SuperPoD into logical SuperPoDs based on the splitting policy for affinity scheduling of training jobs.<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>If this field is not specified,</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>sets the logical SuperPoD size of this job to the total number of NPUs configured for the job during scheduling.</span></p>
<p id="p1250719624216"><a name="p1250719624216"></a><a name="p1250719624216"></a>For details, see <a href="../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod">UnifiedBus Device Network Description</a>.</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>This field can be used only on Atlas 900 A3 SuperPoDs, Atlas 800T A3 SuperPoD servers, Atlas 800I A3 SuperPoD servers, Atlas 850E SuperPoDs, and Atlas 950 SuperPoDs.</li><li>After this field is used, the tor-affinity field does not need to be configured separately.</li><li>FAQ: <a href="https://gitcode.com/Ascend/mind-cluster/issues/377">Total number of chips requested by the job is 32, sp-block set to 32 allows normal training, sp-block set to 16 fails to complete training, training container error report indicates initialization connection failed.</a></li></ul>
</div></div>
</td>
</tr>
<tr id="row16235354174110"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p950710610422"><a name="p950710610422"></a><a name="p950710610422"></a>metadata.annotations['ra-block']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p550719674212"><a name="p550719674212"></a><a name="p550719674212"></a>Identifier for rack affinity scheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p175075613422"><a name="p175075613422"></a><a name="p175075613422"></a>Specify the ra-block field. With dynamic ratio support, a single rack of 64 cards is divided into 4 OSs, each of which is considered a node in the K8s cluster. The intra-rack communication latency is lower than the inter-rack communication latency. Configure this field for rack affinity scheduling of training jobs.</p><p id="p175075613422"><a name="p175075613422"></a><a name="ul1150756144219"></a><a name="ul1150756144219"></a>The value range is 0 to 64 and must be a power of 2.</p><p id="p175075613422"><a name="p175075613422"></a><a name="ul1150756144219"></a><a name="ul1150756144219"></a>By enumeration, the value of ra-block can be {1, 2, 4, 8, 16, 32, 64}.</p>
<div class="note" id="note550714615429"><a name="note550714615429"></a><a name="note550714615429"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a>This field can be used only on Atlas 950 SuperPoD.
</div></div>
</td>
</tr>
<tr id="row14747131720228"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p10781181822210"><a name="p10781181822210"></a><a name="p10781181822210"></a>metadata.annotations['huawei.com/Ascend<em id="i103895254475"><a name="i103895254475"></a><a name="i103895254475"></a>XXX</em>']</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p178151812224"><a name="p178151812224"></a><a name="p178151812224"></a>XXX indicates the chip model. Supported values are 910, 310, and 310P. The value must be consistent with the actual chip type in the environment.</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 ">
    <p id="p5781181818226"><a name="p5781181818226"></a><a name="p5781181818226"></a><span id="ph1378141872210"><a name="ph1378141872210"></a><a name="ph1378141872210"></a>Ascend Docker Runtime</span>obtains this parameter value to mount the corresponding type of NPU to the container.</p>
<p id="zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001609074269_p173021526124817"></a>In distributed jobs, ensure that the nodes running the training job have the same architecture.</p>
<div class="note" id="note269473654014"><a name="note269473654014"></a><a name="note269473654014"></a><span class="notetitle">Note</span>
    <div class="notebody">
        <ul>
            <li>
                <p id="p66941536154018"><a name="p66941536154018"></a><a name="p66941536154018"></a>This parameter supports only the full-NPU scheduling feature of the <span id="ph4213155617124"><a name="ph4213155617124"></a><a name="ph4213155617124"></a>Volcano</span>scheduler. Users who use static vNPU scheduling or other schedulers need to delete the fields related to this parameter in the sample YAML.</p>
            </li>
            <li><p>For Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD, configure this parameter as metadata.annotations['huawei.com/npu'].</p></li>
        </ul>
</div>
</div>
</td>
</tr>
<tr id="row862818313577"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p132726845716"><a name="p132726845716"></a><a name="p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><a name="ul1427218195710"></a><a name="ul1427218195710"></a><ul id="ul1427218195710"><li>large-model-schema: large model task or padding task</li><li>normal-schema: normal job</li><li>null: switch affinity scheduling is not used<div class="note" id="note32586245294"><a name="note32586245294"></a><a name="note32586245294"></a><span class="notetitle">Note</span><div class="notebody"><p id="p5258102462916"><a name="p5258102462916"></a><a name="p5258102462916"></a>Users need to select the task type based on the number of task replicas. If the number of task replicas is less than 4, it is a padding task. If the number of task replicas is greater than or equal to 4, it is a large model task. Normal tasks have no restriction on the number of task replicas.</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p32732087577"><a name="p32732087577"></a><a name="p32732087577"></a>The default value is null, indicating that switch affinity scheduling is not used. Users need to configure it based on the task type.</p><ul id="ul961424647"><li>Switch affinity scheduling 1.0 supports <span id="ph63831524184110"><a name="ph63831524184110"></a><a name="ph63831524184110"></a><term>Atlas training products</term></span>and <span id="ph138318245414"><a name="ph138318245414"></a><a name="ph138318245414"></a><term id="zh-cn_topic_0000001519959665_term57208119917_4"><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a><a name="zh-cn_topic_0000001519959665_term57208119917_4"></a>Atlas A2 training products</term></span>; supports <span id="ph17383182419412"><a name="ph17383182419412"></a><a name="ph17383182419412"></a>PyTorch</span>and <span id="ph1383224134120"><a name="ph1383224134120"></a><a name="ph1383224134120"></a>MindSpore</span>frameworks.</li><li>Switch affinity scheduling 2.0 supports <span id="ph438320243412"><a name="ph438320243412"></a><a name="ph438320243412"></a><term id="zh-cn_topic_0000001519959665_term57208119917_5"><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a><a name="zh-cn_topic_0000001519959665_term57208119917_5"></a>Atlas A2 training products</term></span>; supports <span id="ph134821711841"><a name="ph134821711841"></a><a name="ph134821711841"></a>PyTorch</span>framework.</li><li>Switch affinity scheduling is supported only for full NPU. Static vNPU does not support switch affinity scheduling.</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001609074269_p15256112124619"></a>spec.template.spec.containers[0].resources.requests</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1996615912482"><a name="p1996615912482"></a><a name="p1996615912482"></a><strong id="b118963916494"><a name="b118963916494"></a><a name="b118963916494"></a>Full-NPU scheduling:</strong></p>
<ul><li>Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, Atlas 950 SuperPoD:<ul><li>Configure as huawei.com/npu: <em>x</em></li></ul></li><li>Inference server (with Atlas 300I inference card):<ul><li>Configure as huawei.com/Ascend310: <em>x</em></li></ul></li><li><term>Atlas inference products</term> non-mixed mode:<ul><li>Configure as huawei.com/Ascend310P: <em>x</em></li></ul></li><li><term>Atlas inference products</term> mixed mode:<ul><li>Configure as huawei.com/Ascend310P-V: <em>x</em></li><li>Configure as huawei.com/Ascend310P-VPro: <em>x</em></li><li>Configure as huawei.com/Ascend310P-IPro: <em>x</em></li></ul></li><li>For other products, configure as huawei.com/Ascend910: <em>x</em></li></ul>
<p id="p370843110385"><a name="p370843110385"></a><a name="p370843110385"></a>Depending on the chip type used, the value of x is as follows:</p>
<a name="ul4403181216571"></a><a name="ul4403181216571"></a><ul id="ul4403181216571"><li><span id="zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 training server (fully populated with NPUs)</span>:<a name="zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001609074269_ul169264817234"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4, 8</li><li>Distributed: 1, 2, 4, 8</li></ul>
</li><li><span id="zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 training server (half populated with NPUs)</span>:<a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001609074269_ul1713712328597"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 4</li><li>Distributed: 1, 2, 4</li></ul>
</li><li>Server (with <span id="zh-cn_topic_0000001609074269_ph1223449506"><a name="zh-cn_topic_0000001609074269_ph1223449506"></a><a name="zh-cn_topic_0000001609074269_ph1223449506"></a>Atlas 300T training card</span>):<a name="ul3519194217372"></a><a name="ul3519194217372"></a><ul id="ul3519194217372"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2</li><li>Distributed: 2</li></ul>
</li><li><span id="ph1176216314557"><a name="ph1176216314557"></a><a name="ph1176216314557"></a>Atlas 800T A2 training server</span> and <span id="ph107421743105017"><a name="ph107421743105017"></a><a name="ph107421743105017"></a>Atlas 900 A2 PoD cluster basic unit</span>:<a name="ul169264817234"></a><a name="ul169264817234"></a><ul id="ul169264817234"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li></ul>
</li><li><span id="ph129391532155719"><a name="ph129391532155719"></a><a name="ph129391532155719"></a>Atlas 200T A2 Box16 heterogeneous subrack</span>:<a name="ul555885820439"></a><a name="ul555885820439"></a><ul id="ul555885820439"><li>Single-node single-chip: 1</li><li>Single-node multi-chip: 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li><li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16</li></ul>
</li><li><span id="ph133001904447"><a name="ph133001904447"></a><a name="ph133001904447"></a>Atlas 900 A3 SuperPoD</span>, <span id="ph830011074420"><a name="ph830011074420"></a><a name="ph830011074420"></a>A200T A3 Box8 SuperPoD server</span>, <span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 SuperPoD server</span>:<a name="ul130020074415"></a><a name="ul130020074415"></a><ul id="ul130020074415"><li>Single-node multi-chip: 2, 4, 6, 8, 10, 12, 14, 16</li><li>Distributed: 16</li></ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (8 processors in node without interconnection)</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (16 processors in node without interconnection)</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (4 processor-meshed, 8 processors)</span>:
    <ul>
        <li>Single-node (affinity satisfied): 1, 2, 3, 4, 8</li>
        <li>Single-node (affinity not guaranteed): 5, 6, 7</li>
        <li>Distributed (affinity satisfied): 1, 2, 3, 4, 8</li>
        <li>Distributed (affinity not guaranteed): 5, 6, 7</li>
    </ul>
</li>
<li>
    <span>Server (with Atlas 350 accelerator card) (4 processor-meshed, 16 processors)</span>:
    <ul>
        <li>Single-node (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li>
        <li>Single-node (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li>
        <li>Distributed (affinity satisfied): 1, 2, 3, 4, 8, 12, 16</li>
        <li>Distributed (affinity not guaranteed): 5, 6, 7, 9, 10, 11, 13, 14, 15</li>
    </ul>
</li>
<li>
    <span>Atlas 650E server</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8</li>
        <li>Distributed: 1, 2, 3, 4, 5, 6, 7, 8</li>
    </ul>
</li>
<li>
    <span>Atlas 850E SuperPoD</span>:
    <ul>
        <li>Single-node: 1, 2, 4, 8 (the sp-block value must be consistent with it)</li>
        <li>Distributed: 8 (the sp-block value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not exceed the physical SuperPoD size)</li>
    </ul>
</li>
<li>
    <span>Atlas 950 SuperPoD</span>:
    <ul>
        <li>Single-node: 1, 2, 3, 4, 5, 6, 7, 8 (the sp-block value must be consistent with it)</li>
        <li>Distributed: 8 (the sp-block value must be 8 or a multiple of 8, must be divisible by the total number of cards required by the job, and must not exceed the physical SuperPoD size)</li>
    </ul>
</li>
</ul>
<p id="p1498123034911"><a name="p1498123034911"></a><a name="p1498123034911"></a><strong id="b7488133134911"><a name="b7488133134911"></a><a name="b7488133134911"></a>Static vNPU scheduling:</strong></p>
<p id="p19104113195111"><a name="p19104113195111"></a><a name="p19104113195111"></a>huawei.com/Ascend910-<strong id="b14105734512"><a name="b14105734512"></a><a name="b14105734512"></a><em id="i17105533512"><a name="i17105533512"></a><a name="i17105533512"></a>Y</em></strong>: 1</p>
<p id="p1851116142917"><a name="p1851116142917"></a><a name="p1851116142917"></a>The value is 1. Only vNPUs under one NPU can be used.</p>
<p id="p11413153312435"><a name="p11413153312435"></a><a name="p11413153312435"></a>For example, huawei.com/Ascend910-<em id="i94134332434"><a name="i94134332434"></a><a name="i94134332434"></a>6c.1cpu.16g</em>: 1</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5498134535310"><a name="p5498134535310"></a><a name="p5498134535310"></a>The requested NPU or vNPU type (only one type can be requested) and quantity. Modify them based on the actual situation.</p>
<ul id="ul10782193418818"><li>Only the <span id="ph1038285416813"><a name="ph1038285416813"></a><a name="ph1038285416813"></a><term>Atlas inference products</term></span> non-mixed mode supports static vNPU scheduling.</li><li>The inference server (with <span id="ph1990710374611"><a name="ph1990710374611"></a><a name="ph1990710374611"></a>Atlas 300I inference card</span>) and the <span id="ph629210161695"><a name="ph629210161695"></a><a name="ph629210161695"></a><term>Atlas inference products</term></span> mixed mode do not support static vNPU scheduling.</li><li>For the value of <strong id="b179331118122318"><a name="b179331118122318"></a><a name="b179331118122318"></a><em id="i14933131862318"><a name="i14933131862318"></a><a name="i14933131862318"></a>Y</em></strong>, refer to the "vNPU Type" column of the corresponding product in the virtualization instance template and virtual device type relationship table in the <a href="../04_usage/02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/02_mounting_vnpu_static.md#method-3-mounting-vnpus-on-kubernetes">Static Virtualization</a> section.<p id="p208621211164518"><a name="p208621211164518"></a><a name="p208621211164518"></a>Taking the vNPU type <em id="i412654718449"><a name="i412654718449"></a><a name="i412654718449"></a>Ascend310P-4c.3cpu</em> as an example, the value of <strong id="b1835616104433"><a name="b1835616104433"></a><a name="b1835616104433"></a><em id="i135681014319"><a name="i135681014319"></a><a name="i135681014319"></a>Y</em></strong> is 4c.3cpu, excluding the preceding Ascend310P.</p>
    </li></ul>
</td>
</tr>
<tr id="row25918533287"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p05117110298"><a name="p05117110298"></a><a name="p05117110298"></a>spec.template.spec.containers[0].resources.limits</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p13683185074711"><a name="p13683185074711"></a><a name="p13683185074711"></a>The limited NPU or vNPU type (only one type can be requested) and quantity. Modify them based on the actual situation.</p>
<p id="p16683135019479"><a name="p16683135019479"></a><a name="p16683135019479"></a>The chip name and quantity in limits must be consistent with those in requests.</p>
</td>
</tr>
<tr id="row141124616406"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="p9313107114010"><a name="p9313107114010"></a><a name="p9313107114010"></a>super-pod-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="p1531312713409"><a name="p1531312713409"></a><a name="p1531312713409"></a>Affinity scheduling policy used by super-pod jobs. It must be declared in the label of the YAML file.</p>
<a name="ul231337194020"></a><a name="ul231337194020"></a><ul id="ul231337194020"><li>soft: When cluster resources do not meet the super-pod affinity requirement, the job continues to be scheduled using fragmented resources in the cluster.</li><li>hard: When cluster resources do not meet the super-pod affinity requirement, the job remains Pending and waits for resources.</li><li>Other values or when this parameter is not passed: Super-pod affinity scheduling is enforced.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p2313117194012"><a name="p2313117194012"></a><a name="p2313117194012"></a>This parameter can be used only in the <span id="ph133130710403"><a name="ph133130710403"></a><a name="ph133130710403"></a>Atlas 900 A3 SuperPoD</span>.</p>
</td>
</tr>
<tr id="rowcustomjobkey2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobkey2"><a name="pcustomjobkey2"></a><a name="pcustomjobkey2"></a>customJobKey</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobkeyvalue2"><a name="pcustomjobkeyvalue2"></a><a name="pcustomjobkeyvalue2"></a>User-defined label used to set the unique job identifier through two-level indirection, for example:<br> customJobKey: tid<br> tid: "123456"</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="pcustomjobkeydesc2"><a name="pcustomjobkeydesc2"></a><a name="pcustomjobkeydesc2"></a>You can use customJobKey or custom-job-id to set a unique job identifier, making it easier to filter key information such as alarms and ISSUEs related to the job.<br> <ul><li>For vcjob, set it in the metadata.labels label of the Job resource.<br></li> <li>For deploy jobs, set it in the spec.template.metadata.labels label of the Deployment resource.</li></ul></p>
</td>
</tr>
<tr id="rowcustomjobid2"><td class="cellrowborder" valign="top" width="22.58%" headers="mcps1.2.4.1.1 "><p id="pcustomjobid2"><a name="pcustomjobid2"></a><a name="pcustomjobid2"></a>custom-job-id</p>
</td>
<td class="cellrowborder" valign="top" width="40.86%" headers="mcps1.2.4.1.2 "><p id="pcustomjobidvalue2"><a name="pcustomjobidvalue2"></a><a name="pcustomjobidvalue2"></a>A user-defined label that directly sets the unique job identifier, for example:<br> custom-job-id: "123456"</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row4635558201210"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1499116019135"><a name="zh-cn_topic_0000001951418201_p1499116019135"></a><a name="zh-cn_topic_0000001951418201_p1499116019135"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p599118017133"><a name="zh-cn_topic_0000001951418201_p599118017133"></a><a name="zh-cn_topic_0000001951418201_p599118017133"></a>Recovery strategies available for the job.</p>
<a name="zh-cn_topic_0000001951418201_ul139911803137"></a><a name="zh-cn_topic_0000001951418201_ul139911803137"></a><ul id="zh-cn_topic_0000001951418201_ul139911803137"><li>retry: process-level online recovery.</li><li>recover: process-level rescheduling.</li><li>recover-in-place: process-level in-place recovery.</li><li>dump: save dying gasp.</li><li>exit: exit training.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul169911906135"></a><a name="zh-cn_topic_0000001951418201_ul169911906135"></a>recover-strategy is configured under the annotations of the job YAML. Its value is any combination of the five strategies, separated by commas.
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row10152132415157"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10821192541514"><a name="zh-cn_topic_0000001951418201_p10821192541514"></a><a name="zh-cn_topic_0000001951418201_p10821192541514"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><ul id="zh-cn_topic_0000001951418201_ul5821162501510"><li>on: enable Pod-level rescheduling.</li><li>Other values or not using this field: disable Pod-level rescheduling.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p78221125201514"><a name="zh-cn_topic_0000001951418201_p78221125201514"></a><a name="zh-cn_topic_0000001951418201_p78221125201514"></a>Pod-level rescheduling means that after a job failure occurs, not all Pods are deleted. Instead, only the failed Pod is deleted, and a new Pod is created for rescheduling.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note5822925151516"><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><ul id="zh-cn_topic_0000001951418201_ul17822112517158"><li>The default rescheduling mode is Job-level rescheduling. To enable Pod-level rescheduling, add this field.</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row576132216324"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1772202423212"><a name="zh-cn_topic_0000001951418201_p1772202423212"></a><a name="zh-cn_topic_0000001951418201_p1772202423212"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><ul id="zh-cn_topic_0000001951418201_ul972624133214"><li>ignore: Ignore the sub-healthy node. Subsequent jobs do not prioritize this node in affinity scheduling.</li><li>graceExit: Do not use the sub-healthy node. Save the dying gasp checkpoint file, then perform rescheduling. Subsequent jobs are not scheduled to this node.</li><li>forceExit: Do not use the sub-healthy node. Exit the job directly without saving, then perform rescheduling. Subsequent jobs are not scheduled to this node.</li><li>The default value is ignore.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1973102463218"><a name="zh-cn_topic_0000001951418201_p1973102463218"></a><a name="zh-cn_topic_0000001951418201_p1973102463218"></a>Processing policy for nodes whose status is SubHealthy.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note173271703519"><a name="zh-cn_topic_0000001951418201_note173271703519"></a><a name="zh-cn_topic_0000001951418201_note173271703519"></a><span class="notetitle">Note</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p163271901355"><a name="zh-cn_topic_0000001951418201_p163271901355"></a><a name="zh-cn_topic_0000001951418201_p163271901355"></a>When using the graceExit policy, ensure that the dying gasp checkpoint saving function is enabled for the job.</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1314311835012"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p123205151739"><a name="zh-cn_topic_0000001951418201_p123205151739"></a><a name="zh-cn_topic_0000001951418201_p123205151739"></a>fault-retry-times</p>
<p id="zh-cn_topic_0000001951418201_p196969196112"><a name="zh-cn_topic_0000001951418201_p196969196112"></a><a name="zh-cn_topic_0000001951418201_p196969196112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1192310597344"><a name="zh-cn_topic_0000001951418201_p1192310597344"></a><a name="zh-cn_topic_0000001951418201_p1192310597344"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p109232597342"><a name="zh-cn_topic_0000001951418201_p109232597342"></a><a name="zh-cn_topic_0000001951418201_p109232597342"></a>To handle service-plane faults, you must configure the number of times the service plane can be retried unconditionally.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note15571815115017"><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><span class="notetitle">Note</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><ul id="zh-cn_topic_0000001951418201_ul15238182410364"><li>To use the unconditional retry function, ensure that an abnormal training process causes the container to exit abnormally. If the container does not exit abnormally, the retry cannot succeed.</li><li>Currently, only the <span id="ph1377171612516"><a name="ph1377171612516"></a><a name="ph1377171612516"></a>Atlas 800T A2 training server</span> and the <span id="zh-cn_topic_0000001951418201_ph14104952376"><a name="zh-cn_topic_0000001951418201_ph14104952376"></a><a name="zh-cn_topic_0000001951418201_ph14104952376"></a>Atlas 900 A2 PoD cluster basic unit</span> support the unconditional retry function.</li><li>Process-level recovery triggers a service-plane fault. To use process-level recovery, you must configure this parameter.</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row260912190502"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2966613113520"><a name="zh-cn_topic_0000001951418201_p2966613113520"></a><a name="zh-cn_topic_0000001951418201_p2966613113520"></a>None (no fault-retry-times) or 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p2096618130353"><a name="zh-cn_topic_0000001951418201_p2096618130353"></a><a name="zh-cn_topic_0000001951418201_p2096618130353"></a>The job does not use the unconditional retry function and cannot detect service-plane faults. The maxRetry of vcjob still takes effect.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11217021145014"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1929464718814"><a name="zh-cn_topic_0000001951418201_p1929464718814"></a><a name="zh-cn_topic_0000001951418201_p1929464718814"></a>restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><ul id="zh-cn_topic_0000001951418201_ul193373071216"><li>Never: Never restart</li><li>Always: Always restart</li><li>OnFailure: Restart on failure</li><li>ExitCode: Determine whether to restart the Pod based on the process exit code. If the error code is 1 to 127, the Pod is not restarted. If the error code is 128 to 255, the Pod is restarted.<div class="note" id="zh-cn_topic_0000001951418201_note278954373014"><a name="zh-cn_topic_0000001951418201_note278954373014"></a><a name="zh-cn_topic_0000001951418201_note278954373014"></a><span class="notetitle">Note</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p14789194311309"><a name="zh-cn_topic_0000001951418201_p14789194311309"></a><a name="zh-cn_topic_0000001951418201_p14789194311309"></a><del>ExitCode</del> is not supported for vcjob.</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1129434710811"><a name="zh-cn_topic_0000001951418201_p1129434710811"></a><a name="zh-cn_topic_0000001951418201_p1129434710811"></a>Container restart policy. When unconditional retry for service-plane faults is configured, the container restart policy must be set to <span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue182751614652"><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a>"Never"</span>.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1116371844811"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a>terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a>0 &lt; terminationGracePeriodSeconds &lt;<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a> grace-over-time</strong> parameter value</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a>The time elapsed from when the container receives SIGTERM to when it is forcibly stopped by <span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a>K8s</span>. This value must be greater than 0 and less than the value of the "<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a>grace-over-time</strong>" parameter in the volcano-v<em id="zh-cn_topic_0000001951418201_i1645121221719"><a name="zh-cn_topic_0000001951418201_i1645121221719"></a><a name="zh-cn_topic_0000001951418201_i1645121221719"></a>{version}</em>.yaml file. In addition, it must be long enough to save the CKPT file. Modify it based on the actual situation. For details, see <a href="https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">Container Lifecycle Hooks</a> on the official <span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a>K8s</span> website.</p>
<div class="note" id="zh-cn_topic_0000001951418201_note17641176363"><a name="zh-cn_topic_0000001951418201_note17641176363"></a><a name="zh-cn_topic_0000001951418201_note17641176363"></a><div class="notebody"><p id="zh-cn_topic_0000001951418201_p97641517103616"><a name="zh-cn_topic_0000001951418201_p97641517103616"></a><a name="zh-cn_topic_0000001951418201_p97641517103616"></a>This field takes effect only when fault-scheduling is set to grace. When fault-scheduling is set to force, this field is invalid.</p>
</div></div>
</td>
</tr>
</tbody>
</table>

## Description of Other Parameters

### huawei.com/schedule_policy Configuration Description<a name="schedule_policy"></a>

**Table 4** `huawei.com/schedule_policy` configuration description

|Configuration|Description|Form example|
|--|--|--|
|`chip4-node8`|One node with 8 chips, where every 4 chips form one interconnect ring.|Atlas 800 training server (model 9000)/Atlas 800 training server (model 9010) with fully populated with chips/Atlas 350 accelerator card (8 chips in total, where every 4 chips are connected through a UB mezzanine board)|
|`chip1-node2`|One node with 2 chips.|Atlas 300T training card. One card supports at most 1 chip and one node supports at most 2 cards.|
|`chip4-node4`|One node with 4 chips forming one interconnect ring.|Atlas 800 training server (model 9000)/Atlas 800 training server (model 9010) with half populated with chips|
|`chip8-node8`|One node with 8 chips, where all 8 chips are on one interconnect ring.|Atlas 800T A2 training server/Atlas 850E SuperPoD/Atlas 650E server|
|`chip8-node16`|One node with 16 chips, where every 8 chips are on one interconnect ring.|Atlas 200T A2 Box16 heterogeneous subrack|
|`chip2-node8`|One node with 8 chips, where every 2 chips are on one interconnect ring.|Atlas 9000 A3 SuperPoD cluster computing system|
|`chip2-node16`|One node with 16 chips, where every 2 chips are on one interconnect ring.|Atlas 800T A3 SuperPoD server|
|`chip2-node8-sp`|One node with 8 chips, where every 2 chips are on one interconnect ring, and multiple servers form a SuperPoD.|Atlas 9000 A3 SuperPoD cluster computing system|
|`chip2-node16-sp`|One node with 16 chips, where every 2 chips are on one interconnect ring, and multiple servers form a SuperPoD.|Atlas 900 A3 SuperPoD|
|`chip4-node16`|One node with 16 chips, where every 4 chips are on one interconnect ring.|Atlas 350 accelerator card with 16 chips in total, where every 4 chips are connected through a UB mezzanine board|
|`chip1-node8`|One node with 8 chip, where there is no interconnect between chips.|Atlas 350 accelerator card with 8 chips in total, where there is no interconnect between chips|
|`chip1-node16`|One node with 16 chips, where there is no interconnect between chips.|Atlas 350 accelerator card with 16 chips in total, where there is no interconnect between chips.|
|`chip8-node8-sp`|One node with 8 chip, where all 8 chips are on one interconnect ring, and multiple servers form a SuperPoD.|Atlas 850E SuperPoD|
|`chip8-node8-ra64-sp`|One node with 8 chips, where all 8 chips are on one interconnect ring, 64 nodes form one computing rack, and multiple racks form a SuperPoD.|Atlas 950 SuperPoD|
|`chip1-softShareDev`|Dedicated scheduling policy for soft partitioning virtualization.|Atlas 800I A2, Atlas 800I A3, Atlas 350 accelerator card|
|`multilevel`|Used in multi-level scheduling scenarios. For details about how to use multi-level scheduling, see [Multi-level Scheduling](../04_usage/03_basic_scheduling/04_multi_level_scheduling.md).|Atlas 900 A3 SuperPoD, Atlas 950 SuperPoD|
