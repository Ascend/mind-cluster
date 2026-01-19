// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package jobv2 a series of job test function
package jobv2

import (
	"context"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/kube"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestAddJob(t *testing.T) {
	convey.Convey("test addJob", t, func() {
		uniqueQueue = sync.Map{}
		mockInitCmAndCache := gomonkey.ApplyFunc(job.InitCmAndCache, func(podGroup v1beta1.PodGroup) {
		})
		defer mockInitCmAndCache.Reset()
		convey.Convey("test podGroup dose not exist, pod dose not exist, job cache dose not exist", func() {
			addJob(jobUid1)
			_, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
		convey.Convey("test podGroup exists, pod dose not exist, job cache dose not exist", func() {
			newPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podgroup.SavePodGroup(newPGInfo)
			addJob(jobUid1)
			_, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
		convey.Convey("test podGroup exists, pod exists, job cache dose not exist", func() {
			newPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			pod.SavePod(newPod)
			addJob(jobUid1)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
		convey.Convey("test podGroup exists, pod not exists, job exists", func() {
			newPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			pod.DeletePod(newPod)
			newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
			job.SaveJobCache(jobUid1, newJob)
			addJob(jobUid1)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
	})
}

func getDemoJob(jobName1 string, jobNameSpace string, jobUid1 string) constant.JobInfo {
	return constant.JobInfo{
		Name:      jobName1,
		NameSpace: jobNameSpace,
		Key:       jobUid1,
	}
}

func TestDeleteJob(t *testing.T) {
	convey.Convey("test deleteJob", t, func() {
		mockDeleteCmAndCache := gomonkey.ApplyFunc(kube.DeleteConfigMap, func(cmName, cmNamespace string) error {
			return nil
		})
		defer mockDeleteCmAndCache.Reset()
		convey.Convey("test job dose not exist", func() {
			deleteJob(jobUid1)
			_, ok := job.GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
		convey.Convey("test job exists", func() {
			newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
			job.SaveJobCache(jobUid1, newJob)
			deleteJob(jobUid1)
			_, ok := job.GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestPreDeleteJob(t *testing.T) {
	convey.Convey("test preDeleteJob", t, func() {
		mockUpdateCmAndCache := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
			func(cmName, cmNamespace string, data, label map[string]string) error {
				return nil
			})
		defer mockUpdateCmAndCache.Reset()
		convey.Convey("test job dose not exist", func() {
			preDeleteJob(jobUid1)
			_, ok := job.GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
		convey.Convey("test job exists and job is not preDelete, pod exists", func() {
			newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
			newJob.IsPreDelete = false
			job.SaveJobCache(jobUid1, newJob)
			newPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			pod.SavePod(newPod)
			preDeleteJob(jobUid1)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
			jobInfo, ok := job.GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.IsPreDelete, convey.ShouldEqual, false)
		})
		convey.Convey("test job exists and job is not preDelete, pod dose not exist", func() {
			newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
			newJob.IsPreDelete = false
			job.SaveJobCache(jobUid1, newJob)
			newPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			pod.DeletePod(newPod)
			preDeleteJob(jobUid1)
			jobInfo, ok := job.GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.IsPreDelete, convey.ShouldEqual, true)
		})
		convey.Convey("test job exists and job is preDelete", func() {
			preDeleteJob(jobUid1)
			jobInfo, ok := job.GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.IsPreDelete, convey.ShouldEqual, true)
		})
	})
}

func TestGetStatusByOldStatus(t *testing.T) {
	convey.Convey("test getStatusByOldStatus", t, func() {
		newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test job is running", func() {
			newJob.Status = job.StatusJobRunning
			convey.So(getStatusByOldStatus(newJob), convey.ShouldEqual, job.StatusJobFail)
		})
		convey.Convey("test job is completed", func() {
			newJob.Status = job.StatusJobCompleted
			convey.So(getStatusByOldStatus(newJob), convey.ShouldEqual, job.StatusJobCompleted)
		})
		convey.Convey("test job is failed", func() {
			newJob.Status = job.StatusJobFail
			convey.So(getStatusByOldStatus(newJob), convey.ShouldEqual, job.StatusJobFail)
		})
		convey.Convey("test job is other status", func() {
			newJob.Status = ""
			convey.So(getStatusByOldStatus(newJob), convey.ShouldEqual, job.StatusJobPending)
		})
	})
}

func TestGetStatusByCacheForPending(t *testing.T) {
	mockDeviceAllocateIsCompleted := gomonkey.ApplyFunc(pod.DeviceAllocateIsCompleted, func(p v1.Pod) bool {
		return true
	})
	defer mockDeviceAllocateIsCompleted.Reset()
	convey.Convey("test getStatusByCache for pending", t, func() {
		newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test pod is empty and podGroup is empty", func() {
			var podGroup v1beta1.PodGroup
			var podsInJob map[string]v1.Pod
			isPreDelete, status := getStatusByCache(podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeTrue)
			convey.So(status, convey.ShouldEqual, job.StatusJobPending)
		})
		convey.Convey("test pod is empty and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			var podsInJob map[string]v1.Pod
			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobPending)
		})
		convey.Convey("test pod is pending and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podsInJob := make(map[string]v1.Pod)
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodPending)
			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobPending)
		})
		convey.Convey("test pod is part running and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 2

			podsInJob := make(map[string]v1.Pod)
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodRunning)

			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobPending)
		})
		convey.Convey("test pod is part complete and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 2

			podsInJob := make(map[string]v1.Pod)
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodSucceeded)
			podsInJob[podUid2] = getDemoPodWithStatus(podName2, podNameSpace1, podUid2, v1.PodPending)

			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobPending)
		})
	})
}

func TestGetStatusByCacheForRunning(t *testing.T) {
	mockDeviceAllocateIsCompleted := gomonkey.ApplyFunc(pod.DeviceAllocateIsCompleted, func(p v1.Pod) bool {
		return true
	})
	defer mockDeviceAllocateIsCompleted.Reset()
	convey.Convey("test getStatusByCache for running", t, func() {
		newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test pod is running and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 1
			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodRunning)
			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobRunning)
		})
		convey.Convey("test pod is all running and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 2

			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodRunning)
			podsInJob[podUid2] = getDemoPodWithStatus(podName2, podNameSpace1, podUid2, v1.PodRunning)

			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobRunning)
		})
	})
}

func TestGetStatusByCacheForComplete(t *testing.T) {
	mockDeviceAllocateIsCompleted := gomonkey.ApplyFunc(pod.DeviceAllocateIsCompleted, func(p v1.Pod) bool {
		return true
	})
	defer mockDeviceAllocateIsCompleted.Reset()
	convey.Convey("test getStatusByCache for complete", t, func() {
		newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test pod is complete and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 1
			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodSucceeded)
			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobCompleted)
		})
		convey.Convey("test pod is all complete and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 2

			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodSucceeded)
			podsInJob[podUid2] = getDemoPodWithStatus(podName2, podNameSpace1, podUid2, v1.PodSucceeded)

			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobCompleted)
		})
	})
}

func TestGetStatusByCacheForFailed(t *testing.T) {
	mockDeviceAllocateIsCompleted := gomonkey.ApplyFunc(pod.DeviceAllocateIsCompleted, func(p v1.Pod) bool {
		return true
	})
	defer mockDeviceAllocateIsCompleted.Reset()
	convey.Convey("test getStatusByCache for failed", t, func() {
		newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test pod is failed and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 1
			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodFailed)
			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobFail)
		})
		convey.Convey("test pod is all failed and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 2

			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodFailed)
			podsInJob[podUid2] = getDemoPodWithStatus(podName2, podNameSpace1, podUid2, v1.PodFailed)

			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobFail)
		})
		convey.Convey("test pod is part failed and podGroup is not empty", func() {
			podGroup := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			podGroup.Spec.MinMember = 2

			podsInJob := map[string]v1.Pod{}
			podsInJob[podUid1] = getDemoPodWithStatus(podName1, podNameSpace1, podUid1, v1.PodFailed)
			podsInJob[podUid2] = getDemoPodWithStatus(podName2, podNameSpace1, podUid2, v1.PodRunning)

			isPreDelete, status := getStatusByCache(*podGroup, podsInJob, newJob)
			convey.So(isPreDelete, convey.ShouldBeFalse)
			convey.So(status, convey.ShouldEqual, job.StatusJobFail)
		})
	})
}

func TestUpdateJob(t *testing.T) {
	convey.Convey("test updateJob", t, func() {
		newJob := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test podGroup is empty and job is empty", func() {
			updateJob(jobUid1)
			convey.So(job.GetJobIsExists(jobUid1), convey.ShouldBeTrue)
		})
		convey.Convey("test isPreDelete and status is same with job cache", func() {
			newJob.IsPreDelete = false
			newJob.Status = job.StatusJobFail
			mockGetStatusByCache := gomonkey.ApplyFunc(getStatusByCache,
				func(podGroup v1beta1.PodGroup, podsInJob map[string]v1.Pod, jobInfo constant.JobInfo) (bool, string) {
					return false, job.StatusJobFail
				})
			defer mockGetStatusByCache.Reset()
			updateJob(jobUid1)
			convey.So(job.GetJobIsExists(jobUid1), convey.ShouldBeTrue)
		})
	})
}
