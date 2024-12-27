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

const (
	// jobRestartingReason is added in an ascendjob when it is restart.
	jobRestartingReason = "jobRestarting"
	// FailedDeleteJobReason is added in an ascendjob when it is deleted failed.
	FailedDeleteJobReason = "FailedDeleteJob"
	// SuccessfulDeleteJobReason is added in an ascendjob when it is deleted successful.
	SuccessfulDeleteJobReason = "SuccessfulDeleteJob"

	controllerName = "ascendjob-controller"

	// volcanoTaskSpecKey task spec key used in pod annotation when EnableGangScheduling is true
	volcanoTaskSpecKey = "volcano.sh/task-spec"

	// gang scheduler name.
	gangSchedulerName = "volcano"

	// exitedWithCodeReason is the normal reason when the pod is exited because of the exit code.
	exitedWithCodeReason = "ExitedWithCode"
	// podTemplateRestartPolicyReason is the warning reason when the restart
	// policy is set in pod template.
	podTemplateRestartPolicyReason = "SettedPodTemplateRestartPolicy"
	// jobSchedulerNameReason is the warning reason when other scheduler name is set in job with gang-scheduling enabled
	jobSchedulerNameReason = "SettedJobSchedulerName"
	// podTemplateSchedulerNameReason is the warning reason when other scheduler name is set
	// in pod templates with gang-scheduling enabled
	podTemplateSchedulerNameReason = "SettedPodTemplateSchedulerName"
	// gangSchedulingPodGroupAnnotation is the annotation key used by batch schedulers
	gangSchedulingPodGroupAnnotation = "scheduling.k8s.io/group-name"
	// for ascend-volcano-plugin rescheduling
	rankIndexKey = "hccl/rankIndex"
	// prefix of request npu name
	npuPrefix   = "huawei.com/"
	npuCoreName = "huawei.com/npu-core"

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
	hcclCtrName      = "hccl-controller"
	atlasTaskKey     = "ring-controller.atlas"
	// the status of mount chips for non-worker Pods
	nonWorkerPodMountChipStatus = "nonWorkerPodMountChipStatus"
)

const (
	msServerNum     = "MS_SERVER_NUM"
	msWorkerNum     = "MS_WORKER_NUM"
	msLocalWorker   = "MS_LOCAL_WORKER"
	msSchedHost     = "MS_SCHED_HOST"
	msSchedPort     = "MS_SCHED_PORT"
	msRole          = "MS_ROLE"
	msNodeRank      = "MS_NODE_RANK"
	msSchedulerRole = "MS_SCHED"
	msWorkerRole    = "MS_WORKER"

	ptMasterAddr     = "MASTER_ADDR"
	ptMasterPort     = "MASTER_PORT"
	ptWorldSize      = "WORLD_SIZE"
	ptRank           = "RANK"
	ptLocalWorldSize = "LOCAL_WORLD_SIZE"
	ptLocalRank      = "LOCAL_RANK"

	tfChiefIP     = "CM_CHIEF_IP"
	tfChiefPort   = "CM_CHIEF_PORT"
	tfChiefDevice = "CM_CHIEF_DEVICE"
	tfWorkerSize  = "CM_WORKER_SIZE"
	tfLocalWorker = "CM_LOCAL_WORKER"
	tfWorkerIP    = "CM_WORKER_IP"
	tfRank        = "CM_RANK"

	hostNetwork = "HostNetwork"
	npuPod      = "NPU_POD"

	mindxServerIPEnv         = "MINDX_SERVER_IP"                              // clusterd grpc service env name
	mindxServiceName         = "clusterd-grpc-svc"                            // clusterd grpc service name
	mindxServiceNamespace    = "mindx-dl"                                     // clusterd grpc service namespace
	mindxDefaultServerDomain = "clusterd-grpc-svc.mindx-dl.svc.cluster.local" // clusterd grpc service domain
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
)
