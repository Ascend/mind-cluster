// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package jobv2 a series of message processing function
package jobv2

import (
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/kube"
)

func addJob(jobKey string) {
	podGroupCache := podgroup.GetPodGroup(jobKey)
	// if both pod and podGroup exist, skip to update flow
	if podGroupCache.Name != "" && len(pod.GetPodByJobId(jobKey)) > 0 {
		uniqueQueue.Store(jobKey, queueOperatorUpdate)
		return
	}
	oldJobInfo := job.GetJobByNameSpaceAndName(podgroup.GetJobNameByPG(&podGroupCache), podGroupCache.Namespace)
	if oldJobInfo.Name != "" && oldJobInfo.IsPreDelete && oldJobInfo.Key != jobKey {
		// if old job is pre delete, and new job is add, delete old job cache
		job.DeleteJobCache(oldJobInfo.Key)
	}
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		job.InitCmAndCache(podGroupCache)
		return
	}
	if jobInfo.Status == job.StatusJobPending {
		return
	}
	// if cache exists, add update-message to queue
	uniqueQueue.Store(jobKey, queueOperatorUpdate)
}

func updateJob(jobKey string) {
	pg := podgroup.GetPodGroup(jobKey)
	// jobInfo status is empty if jobInfo is not exists
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok && pg.Name == "" {
		hwlog.RunLog.Debugf("job cache is empty and podGroup is empty, skip %s message", jobKey)
		return
	}
	podsInJob := pod.GetPodByJobId(jobKey)
	isPreDelete, status := getStatusByCache(pg, podsInJob, jobInfo)
	if ok && jobInfo.Status == status && jobInfo.IsPreDelete == isPreDelete {
		hwlog.RunLog.Debugf("the job %s cache is consistent with pod and podGroup cache", jobInfo.Name)
		return
	}
	if isPreDelete {
		// updateJob to preDelete
		if !jobInfo.IsPreDelete {
			hwlog.RunLog.Debugf("job %s updateJob to preDeleteJob", jobInfo.Name)
			job.PreDeleteCmAndCache(jobKey)
		}
		return
	}
	// updateJob to addJob
	if status == job.StatusJobPending && jobInfo.Status != job.StatusJobPending {
		hwlog.RunLog.Debugf("job %s updateJob to addJob", jobInfo.Name)
		job.InitCmAndCache(pg)
		return
	}
	// update job to running or completed or failed
	if jobInfo.Status != status {
		job.UpdateCmAndCache(status, jobInfo, pg, podsInJob)
		return
	}
	hwlog.RunLog.Warnf("this logic branch is unreachable, there must have been some issues with the code."+
		"isPreDelete: %v, status: %s, job name: %s, job.isPreDelete: %v, job.status: %s",
		isPreDelete, status, jobInfo.Name, jobInfo.IsPreDelete, jobInfo.Status)
}

func getStatusByCache(podGroup v1beta1.PodGroup, podsInJob map[string]v1.Pod, jobInfo constant.JobInfo) (bool, string) {
	if podGroup.Name == "" && len(podsInJob) == 0 {
		return true, getStatusByOldStatus(jobInfo)
	}
	if len(podsInJob) == 0 {
		return false, getStatusByOldStatus(jobInfo)
	}
	isFailed := false
	isSuccess := true
	isRunning := true
	for _, p := range podsInJob {
		if p.Status.Phase == v1.PodFailed {
			isFailed = true
			isSuccess = false
			isRunning = false
			break
		}
		if p.Status.Phase != v1.PodSucceeded {
			isSuccess = false
		}
		// pod is running and device is allocated, then rank table can be completed
		if (p.Status.Phase != v1.PodRunning && p.Status.Phase != v1.PodSucceeded) ||
			!pod.DeviceAllocateIsCompleted(p) {
			isRunning = false
		}
	}
	if isFailed {
		return false, job.StatusJobFail
	}
	if isSuccess && len(podsInJob) >= int(podGroup.Spec.MinMember) {
		return false, job.StatusJobCompleted
	}
	if isRunning && len(podsInJob) >= int(podGroup.Spec.MinMember) {
		return false, job.StatusJobRunning
	}
	return false, getStatusByOldStatus(jobInfo)
}

func getStatusByOldStatus(jobInfo constant.JobInfo) string {
	switch jobInfo.Status {
	case job.StatusJobRunning:
		return job.StatusJobFail
	case job.StatusJobCompleted:
		return job.StatusJobCompleted
	case job.StatusJobFail:
		return job.StatusJobFail
	default:
		return job.StatusJobPending
	}
}

func preDeleteJob(jobKey string) {
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		return
	}
	if jobInfo.IsPreDelete {
		return
	}
	podsInJob := pod.GetPodByJobId(jobKey)
	if len(podsInJob) > 0 {
		uniqueQueue.Store(jobKey, queueOperatorUpdate)
		return
	}
	job.PreDeleteCmAndCache(jobKey)
	kube.RecoverFaultJobInfoCm(jobKey)
}

func deleteJob(joKey string) {
	hwlog.RunLog.Debugf("delete job %s", joKey)
	job.DeleteCmAndCache(joKey)
}
