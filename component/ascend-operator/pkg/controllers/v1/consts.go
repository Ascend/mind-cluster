/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

const (
	// msJobRestarting is added in an ascendjob when it is restarting.
	jobRestartingReason = "jobRestarting"
	// FailedDeleteJobReason is added in an ascendjob when it is deleted failed.
	FailedDeleteJobReason = "FailedDeleteJob"
	// SuccessfulDeleteJobReason is added in an ascendjob when it is deleted successful.
	SuccessfulDeleteJobReason = "SuccessfulDeleteJob"
	// jobRestartingReason is added in an ascendjob when it is restart.

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
	npuPrefix = "huawei.com/Ascend"

	statusPodIPDownwardAPI = "status.podIP"
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
	//If the value is n (n > 0), the pod is rescheduled for the nth time.
	podVersionLabel = "version"
	// defaultPodVersion is the default version of pod.
	defaultPodVersion = 0
	// decimal stands for base-10.
	decimal = 10
	// labelFaultRetryTimes represents the key of label fault-retry-times.
	labelFaultRetryTimes = "fault-retry-times"
)
