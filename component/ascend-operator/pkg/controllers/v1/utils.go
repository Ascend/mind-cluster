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
Package controllers is using for reconcile AscendJob.
*/

package v1

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

// initializeReplicaStatuses initializes the ReplicaStatuses for replica.
func initializeReplicaStatuses(jobStatus *commonv1.JobStatus, rtype commonv1.ReplicaType) {
	if jobStatus.ReplicaStatuses == nil {
		jobStatus.ReplicaStatuses = make(map[commonv1.ReplicaType]*commonv1.ReplicaStatus)
	}

	jobStatus.ReplicaStatuses[rtype] = &commonv1.ReplicaStatus{}
}

// updateJobReplicaStatuses updates the JobReplicaStatuses according to the pod.
func updateJobReplicaStatuses(jobStatus *commonv1.JobStatus, rtype commonv1.ReplicaType, pod *corev1.Pod) {
	hwlog.RunLog.Debugf("before updateJobReplicaStatuses  status<%#v> by pod<%s> phase<%s>",
		jobStatus.ReplicaStatuses[rtype], pod.Name, pod.Status.Phase)
	defer hwlog.RunLog.Debugf("after updateJobReplicaStatuses status<%#v>", jobStatus.ReplicaStatuses[rtype])
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
	default:
	}
}

// ContainsChiefOrMasterSpec check whether replicas having 'Chief' or 'Master'
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

func getContainerNPUResourceNameAndReq(ct corev1.Container) (string, int) {
	for rName, rNum := range ct.Resources.Requests {
		if strings.Contains(string(rName), npuPrefix) {
			return string(rName), int(rNum.Value())
		}
	}
	return "", 0
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

func getNpuReqInfoPerPod(job *mindxdlv1.AscendJob) (string, int) {
	npuWorker := getNpuWorkerSpec(job)
	if npuWorker == nil {
		return "", 0
	}

	for _, ct := range npuWorker.Template.Spec.Containers {
		if ct.Name == mindxdlv1.DefaultContainerName {
			return getContainerNPUResourceNameAndReq(ct)
		}
	}
	return "", 0
}

func getNpuWorkerSpec(job *mindxdlv1.AscendJob) *commonv1.ReplicaSpec {
	status := getNonWorkerPodMountChipStatus(job)
	for rtype, spec := range job.Spec.ReplicaSpecs {
		if status {
			return spec
		}
		if rtype == mindxdlv1.ReplicaTypeWorker {
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
	status := getNonWorkerPodMountChipStatus(job)
	for rtype, spec := range job.Spec.ReplicaSpecs {
		if !status && rtype != mindxdlv1.ReplicaTypeWorker {
			continue
		}
		jobReplicas += *spec.Replicas
	}
	return int(jobReplicas)
}

func getTotalReplicas(job *mindxdlv1.AscendJob) int32 {
	jobReplicas := int32(0)
	for _, spec := range job.Spec.ReplicaSpecs {
		jobReplicas += *spec.Replicas
	}
	return jobReplicas
}

func getRestartCondition(conds []commonv1.JobCondition) *commonv1.JobCondition {
	for _, condition := range conds {
		if condition.Type == commonv1.JobRestarting {
			return &commonv1.JobCondition{
				Reason:  condition.Reason,
				Message: condition.Message,
			}
		}
	}
	return nil
}

func specReplicas(spec *commonv1.ReplicaSpec) int32 {
	if spec.Replicas == nil {
		return int32(1)
	}
	return *spec.Replicas
}

type specInfo struct {
	name   commonv1.ReplicaType
	job    *mindxdlv1.AscendJob
	spec   *commonv1.ReplicaSpec
	status *commonv1.ReplicaStatus
}

type podInfo struct {
	frame           string
	job             *mindxdlv1.AscendJob
	clusterdSvcIp   string
	status          *commonv1.ReplicaStatus
	rtype           commonv1.ReplicaType
	isDynamicCutJob bool
	index           int
	spec            *commonv1.ReplicaSpec
	isMaster        bool
	ip              string
	port            string
	ctReq           int
	npuReplicas     int
	rank            int
}

func (pi *podInfo) DeepCopy() *podInfo {
	return &podInfo{
		isDynamicCutJob: pi.isDynamicCutJob,
		frame:           pi.frame,
		job:             pi.job,
		status:          pi.status,
		rtype:           pi.rtype,
		spec:            pi.spec,
		ip:              pi.ip,
		port:            pi.port,
		ctReq:           pi.ctReq,
		npuReplicas:     pi.npuReplicas,
	}
}

type validateError struct {
	reason  string
	message string
}

func (ve *validateError) Error() string {
	return ve.message
}

func filterPodsByReplicaType(pods []*corev1.Pod, rt string) []*corev1.Pod {
	var filtered []*corev1.Pod
	for _, pod := range pods {
		if pod.Labels[commonv1.ReplicaTypeLabel] == rt {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}

func checkNonWorkerRplMountChips(ji *jobInfo) bool {
	for rtype, spec := range ji.rpls {
		if rtype == mindxdlv1.ReplicaTypeWorker {
			continue
		}
		if checkContainersResourceReq(spec.Template.Spec.Containers) {
			return true
		}
	}
	return false
}

func checkContainersResourceReq(containers []corev1.Container) bool {
	for _, container := range containers {
		if container.Name == mindxdlv1.DefaultContainerName {
			rNum := getContainerResourceReq(container)
			if rNum > 0 {
				return true
			}
		}
	}
	return false
}

func getNonWorkerPodMountChipStatus(job *mindxdlv1.AscendJob) bool {
	annotations := job.GetAnnotations()
	status, ok := annotations[nonWorkerPodMountChipStatus]
	if !ok {
		return false
	}
	return status == "true"

}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
