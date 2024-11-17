/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func getContainerExitCode(pod *corev1.Pod) int32 {
	var exitCode int32 = 0xbeef // magic number
	for _, status := range pod.Status.ContainerStatuses {
		state := status.State
		if status.Name == mindxdlv1.DefaultContainerName && state.Terminated != nil {
			exitCode = state.Terminated.ExitCode
		}
	}
	return exitCode
}

func setRestartPolicy(podTemplateSpec *corev1.PodTemplateSpec, spec *commonv1.ReplicaSpec) {
	// This is necessary since restartPolicyExitCode is not supported in v1.PodTemplateSpec
	if spec.RestartPolicy == commonv1.RestartPolicyExitCode {
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicyNever
	} else {
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicy(spec.RestartPolicy)
	}
}

func getTotalTrainReplicas(job *mindxdlv1.AscendJob) int32 {
	count := int32(0)
	for _, spec := range job.Spec.ReplicaSpecs {
		count += *spec.Replicas
	}
	return count
}

// initializeReplicaStatuses initializes the ReplicaStatuses for replica.
func initializeReplicaStatuses(jobStatus *commonv1.JobStatus, rtype commonv1.ReplicaType) {
	if jobStatus.ReplicaStatuses == nil {
		jobStatus.ReplicaStatuses = make(map[commonv1.ReplicaType]*commonv1.ReplicaStatus)
	}

	jobStatus.ReplicaStatuses[rtype] = &commonv1.ReplicaStatus{}
}

// updateJobReplicaStatuses updates the JobReplicaStatuses according to the pod.
func updateJobReplicaStatuses(jobStatus *commonv1.JobStatus, rtype commonv1.ReplicaType, pod *corev1.Pod) {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		jobStatus.ReplicaStatuses[rtype].Active++
	case corev1.PodSucceeded:
		jobStatus.ReplicaStatuses[rtype].Succeeded++
	case corev1.PodFailed:
		if pod.DeletionTimestamp != nil {
			hwlog.RunLog.Infof("pod<%s> is deleting, so it can not be treat as failed", pod.Name)
			return
		}
		jobStatus.ReplicaStatuses[rtype].Failed++
	}
}

func ContainsChiefOrMasterSpec(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) bool {
	if _, ok := replicas[mindxdlv1.TensorflowReplicaTypeChief]; ok {
		return true
	}
	if _, ok := replicas[mindxdlv1.PytorchReplicaTypeMaster]; ok {
		return true
	}
	return false
}

func getContainerResourceReq(ct corev1.Container) int {
	for rName, rNum := range ct.Resources.Requests {
		if strings.Contains(string(rName), npuPrefix) {
			return int(rNum.Value())
		}
	}
	return 0
}

func getNpuReqPerPod(job *mindxdlv1.AscendJob) int {
	npuWorker := getNpuWorkerSpec(job)
	if npuWorker == nil {
		return 0
	}

	for _, ct := range npuWorker.Template.Spec.Containers {
		if ct.Name == mindxdlv1.DefaultContainerName {
			return getContainerResourceReq(ct)
		}
	}
	return 0
}

func getNpuWorkerSpec(job *mindxdlv1.AscendJob) *commonv1.ReplicaSpec {
	workerSpec, ok := job.Spec.ReplicaSpecs[mindxdlv1.ReplicaTypeWorker]
	if ok {
		return workerSpec
	}

	for rt, spec := range job.Spec.ReplicaSpecs {
		if rt != mindxdlv1.MindSporeReplicaTypeScheduler {
			return spec
		}
	}

	return nil
}

func localRankStr(req int) string {
	rankStr := ""
	for i := 0; i < req-1; i++ {
		rankStr += strconv.Itoa(i) + ","
	}
	rankStr += strconv.Itoa(req - 1)
	return rankStr
}

func getTotalNpuReplicas(job *mindxdlv1.AscendJob) int {
	jobReplicas := int32(0)
	for rtype, spec := range job.Spec.ReplicaSpecs {
		if rtype == mindxdlv1.MindSporeReplicaTypeScheduler {
			continue
		}
		jobReplicas += *spec.Replicas
	}
	return int(jobReplicas)
}
