/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package v1 is using for reconcile AscendJob.
*/
package v1

import "time"

const (
	// jobRestartingReason is added in an ascendjob when it is restart.
	jobRestartingReason = "jobRestarting"
	// FailedDeleteJobReason is added in an ascendjob when it is deleted failed.
	FailedDeleteJobReason = "FailedDeleteJob"
	// SuccessfulDeleteJobReason is added in an ascendjob when it is deleted successful.
	SuccessfulDeleteJobReason = "SuccessfulDeleteJob"
	// controllerName is the name of controller,used in log.
	controllerName = "ascendjob-controller"
	// volcanoTaskSpecKey volcano.sh/task-spec key used in pod annotation when EnableGangScheduling is true
	volcanoTaskSpecKey = "volcano.sh/task-spec"
	// gangSchedulerName gang scheduler name.
	gangSchedulerName = "volcano"
	// exitedWithCodeReason is the reason of a job that exited with a non-zero code.
	exitedWithCodeReason = "ExitedWithCode"
	// podTemplateRestartPolicyReason is the reason of a job that set podTemplate restartPolicy.
	podTemplateRestartPolicyReason = "SettedPodTemplateRestartPolicy"
	// jobSchedulerNameReason is the warning reason when other scheduler name is set in job with gang-scheduling enabled
	jobSchedulerNameReason = "SettedJobSchedulerName"
	// podTemplateSchedulerNameReason is the warning reason when other scheduler name is set
	// in pod templates with gang-scheduling enabled
	podTemplateSchedulerNameReason = "SettedPodTemplateSchedulerName"
	// gangSchedulingPodGroupAnnotation is the annotation key used by batch schedulers
	gangSchedulingPodGroupAnnotation = "scheduling.k8s.io/group-name"
	npuCoreName                      = "huawei.com/npu-core"

	statusPodIPDownwardAPI = "status.podIP"

	cmRetryTime      = 3
	configmapPrefix  = "rings-config-"
	acjobKind        = "AscendJob"
	vcjobKind        = "Job"
	vcjobLabelKey    = "volcano.sh/job-name"
	deployKind       = "Deployment"
	deployLabelKey   = "deploy-name"
	configmapKey     = "hccl.json"
	configmapVersion = "version"
	// the status of mount chips for non-worker Pods
	nonWorkerPodMountChipStatus = "nonWorkerPodMountChipStatus"
)

const (
	msServerNum     = "MS_SERVER_NUM"
	msSchedHost     = "MS_SCHED_HOST"
	msSchedPort     = "MS_SCHED_PORT"
	msRole          = "MS_ROLE"
	msNodeRank      = "MS_NODE_RANK"
	msSchedulerRole = "MS_SCHED"
	msWorkerRole    = "MS_WORKER"

	ptMasterAddr = "MASTER_ADDR"
	ptMasterPort = "MASTER_PORT"
	ptRank       = "RANK"

	tfChiefIP     = "CM_CHIEF_IP"
	tfChiefPort   = "CM_CHIEF_PORT"
	tfChiefDevice = "CM_CHIEF_DEVICE"
	tfWorkerIP    = "CM_WORKER_IP"
	tfRank        = "CM_RANK"

	hostNetwork = "HostNetwork"
	npuPod      = "NPU_POD"

	mindxServerIPEnv         = "MINDX_SERVER_IP"                              // clusterd grpc service env name
	mindxServiceName         = "clusterd-grpc-svc"                            // clusterd grpc service name
	mindxDefaultServerDomain = "clusterd-grpc-svc.mindx-dl.svc.cluster.local" // clusterd grpc service domain

	// hcclSuperPodLogicId is the logic id of the superpod, ascend container env name
	hcclSuperPodLogicId = "HCCL_LOGIC_SUPERPOD_ID"
	// ascendVisibleDevicesEnv represents the env of ASCEND_VISIBLE_DEVICES
	ascendVisibleDevicesEnv = "ASCEND_VISIBLE_DEVICES"
)

const (
	// vcRescheduleCMName Name of ReSchedulerConfigmap
	vcRescheduleCMName = "vcjob-fault-npu-cm"
	// vcNamespace Namespace of ReSchedulerConfigmap
	vcNamespace = "volcano-system"
	// unconditionalRetryLabelKey label key of unconditional retry job
	unconditionalRetryLabelKey = "fault-retry-times"
	// cmJobRemainRetryTimes judging node fault needs heartbeat info from former session, so should be recorded
	cmJobRemainRetryTimes = "remain-retry-times"
)

const (
	// default min member number of replica
	defaultMinMember = 1
	// batch create pods parameter for k8s api-server
	batchCreateParam = "collectionCreate"
	// batch create pods default size
	batchCreatePodsDefaultSize = 1000
	// batch create default interval
	defaultBatchCreateFailInterval = time.Hour
	// unsetBackoffLimits default Re-scheduling Times of job, it stands for Unlimited.
	unsetBackoffLimits = -1
	// podVersionLabel version of the current pod, if the value is 0, the pod is created for the first time.
	// If the value is n (n > 0), the pod is rescheduled for the nth time.
	podVersionLabel = "version"
	// defaultPodVersion is the default version of pod.
	defaultPodVersion = 0
	// decimal stands for base-10.
	decimal = 10
	// labelFaultRetryTimes represents the key of label fault-retry-times.
	labelFaultRetryTimes = "fault-retry-times"
	// maxReplicas
	maxReplicas = 15000
)

const (
	// NPU310CardName represents the name for 310 npu resource.
	NPU310CardName = "huawei.com/Ascend310"
	// NPU310PCardName represents the name for 310P npu resource.
	NPU310PCardName = "huawei.com/Ascend310P"
	// NPU910CardName represents the name for 910 npu resource.
	NPU910CardName = "huawei.com/Ascend910"
)

const (
	workQueueBaseDelay = 5 * time.Millisecond
	workQueueMaxDelay  = 20 * time.Second
	workQueueQps       = 10
	workQueueBurst     = 100
)
