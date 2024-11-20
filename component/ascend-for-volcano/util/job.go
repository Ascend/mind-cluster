/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// For Job type
const (
	// JobTypeWhole whole card  for NPU  Job.
	JobTypeWhole = iota
	// JobTypeDyCut Dynamic force cutting NPU vJob.
	JobTypeDyCut
	// JobTypeStCut Static force segmentation NPU Job.
	JobTypeStCut
	// JobTypeUnknown unknown NPU Job type.
	JobTypeUnknown
)

// VJob for dynamic NPU Job.
type VJob struct {
	// type: JobTypeWhole, JobTypeDycut, JobTypeStcut.
	Type   int
	Status string
}

// NPUJob only npu vcJob have.
type NPUJob struct {
	// the mapKey is taskID, not Name.
	Tasks         map[api.TaskID]NPUTask
	SelectServers string
	NPUTaskNum    int
	ReqNPUName    string
	ReqNPUNum     int
	*VJob
}

// ComJob all vcJob has.
type ComJob struct {
	Name          api.JobID
	ReferenceName string
	NameSpace     string
	Annotation    map[string]string
	Selector      map[string]string
	Label         map[string]string
}

// SchedulerJobAttr vcJob's attribute.
type SchedulerJobAttr struct {
	ComJob
	*NPUJob
}

// IsSelectorMeetJob check the selectors
func IsSelectorMeetJob(jobSelectors, conf map[string]string) bool {
	for jobKey, jobValue := range jobSelectors {
		confValue, confOk := conf[jobKey]
		if !confOk {
			klog.V(LogErrorLev).Infof("conf has no job selector key:%s.", jobKey)
			return false
		}
		if !strings.Contains(confValue, jobValue) {
			klog.V(LogErrorLev).Infof("conf has no job selector value:%s.", jobValue)
			return false
		}
	}
	return true
}

// IsVJob Determine whether is the NPU virtual job.
// Dynamic segmentation: huawei.com/npu-core.
// static segmentation: huawei.com/Ascend910-Y.
// no segmentation: huawei.com/Ascend910.
func (nJob *NPUJob) IsVJob() bool {
	if len(strings.Split(nJob.ReqNPUName, "-")) > 1 {
		return true
	}
	return false
}

// IsNPUJob Determine whether is the NPU job.
// Dynamic segmentation: huawei.com/npu-core.
// static segmentation: huawei.com/Ascend910-Y.
// no segmentation: huawei.com/Ascend910.
func (nJob *NPUJob) IsNPUJob() bool {
	return strings.Contains(nJob.ReqNPUName, HwPreName)
}

// GetNPUTaskNumInJob get the NPU task number in one job. for some task has no NPU.
func (nJob *NPUJob) GetNPUTaskNumInJob() int {
	if nJob == nil || !nJob.IsNPUJob() {
		return 0
	}
	taskNum := 0
	for _, task := range nJob.Tasks {
		if task.IsNPUTask() {
			taskNum++
		}
	}

	return taskNum
}

// GetVTaskNumInVJob get the NPU task number in one job. for some task has no NPU.
func (nJob *NPUJob) GetVTaskNumInVJob() int {
	if nJob == nil || !nJob.IsVJob() {
		return 0
	}
	taskNum := 0
	for _, task := range nJob.Tasks {
		if task.IsVNPUTask() {
			taskNum++
		}
	}

	return taskNum
}

func (nJob *NPUJob) setJobType() {
	tmpTypes := make(map[int]struct{}, MapInitNum)
	for _, nTask := range nJob.Tasks {
		if nTask.ReqNPUNum == 0 {
			continue
		}
		value := nTask.ComputeTaskType()
		nJob.VJob.Type = value
		tmpTypes[value] = struct{}{}
	}

	// all NPU Task type in job must be same.
	if len(tmpTypes) != 1 {
		nJob.VJob.Type = JobTypeUnknown
		return
	}
}

// SetJobType set virtual job type
// must be used after isVJob.
// for all chips: 910, 310P
func (nJob *NPUJob) SetJobType() {
	if nJob == nil {
		return
	}
	nJob.setJobType()
}

// SetJobStatusByInf set vJob status by podGroup.
func (nJob *NPUJob) SetJobStatusByInf(vcJob *api.JobInfo) {
	if nJob == nil {
		return
	}
	nJob.VJob.Status = string(vcJob.PodGroup.Status.Phase)
}

func ReferenceNameOfJob(job *api.JobInfo) string {
	if job != nil && job.PodGroup != nil && len(job.PodGroup.OwnerReferences) > 0 {
		return job.PodGroup.OwnerReferences[0].Name
	}
	return ""
}

func UuidOfJob(job *api.JobInfo) types.UID {
	if job != nil && job.PodGroup != nil && len(job.PodGroup.OwnerReferences) > 0 {
		return job.PodGroup.OwnerReferences[0].UID
	}
	return ""
}
