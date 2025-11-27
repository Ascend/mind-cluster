// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package job a series of job test function
package job

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/pod"
)

const (
	vcJobKey      = "job"
	nodeName1     = "node1"
	podName1      = "pod1"
	podName2      = "pod2"
	podNameSpace1 = "default"
	podUid1       = "123"
	podUid2       = "456"

	pgMinMember2    = 2
	pgMinMember2Str = "2"
	masterIp        = "127.0.0.1"
	mindIeJobId     = "mindie-ms"
)

func TestPreDeleteCmAndCache(t *testing.T) {
	convey.Convey("test PreDeleteCmAndCache", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test job cache is nil", func() {
			PreDeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not nil, preDeleteCM success. job status should be delete", func() {
			mockGetHcclSlice := gomonkey.ApplyFunc(getHcclSlice,
				func(table constant.RankTable) []string {
					return []string{"123"}
				})
			defer mockGetHcclSlice.Reset()
			mockPreDeleteCM := gomonkey.ApplyFunc(preDeleteCM,
				func(jobInfo constant.JobInfo, hccls []string) bool {
					return true
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			PreDeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, StatusJobFail)
		})
		convey.Convey("when job cache is not nil, job status is completed, preDeleteCM failed. "+
			"job status should be completed", func() {
			mockGetHcclSlice := gomonkey.ApplyFunc(getHcclSlice,
				func(table constant.RankTable) []string {
					return []string{"123"}
				})
			defer mockGetHcclSlice.Reset()
			mockPreDeleteCM := gomonkey.ApplyFunc(preDeleteCM,
				func(jobInfo constant.JobInfo, hccls []string) bool {
					return false
				})
			defer mockPreDeleteCM.Reset()
			jobInfo.Status = StatusJobCompleted
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			PreDeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, StatusJobCompleted)
		})
	})
}

func TestDeleteCmAndCache(t *testing.T) {
	convey.Convey("test DeleteCmAndCache", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		jobInfo.IsPreDelete = true
		convey.Convey("when job cache is nil", func() {
			DeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not nil, deleteCm failed. job should be exists", func() {
			mockPreDeleteCM := gomonkey.ApplyFunc(deleteCm,
				func(jobInfo constant.JobInfo) bool {
					return false
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			DeleteCmAndCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
		})
		convey.Convey("when job cache is not nil, deleteCm success. job should not be exists", func() {
			mockPreDeleteCM := gomonkey.ApplyFunc(deleteCm,
				func(jobInfo constant.JobInfo) bool {
					return true
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			DeleteCmAndCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
		convey.Convey("when job cache is not nil, deleteCm failed but have job with same name. "+
			"job should not be exists", func() {
			mockPreDeleteCM := gomonkey.ApplyFunc(deleteCm,
				func(jobInfo constant.JobInfo) bool {
					return false
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfo2 := getDemoJob(jobName1, jobNameSpace, jobUid2)
			SaveJobCache(jobUid2, jobInfo2)
			defer DeleteJobCache(jobUid2)
			DeleteCmAndCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestInitCmAndCache(t *testing.T) {
	convey.Convey("test InitCmAndCache", t, func() {
		newPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when pg name is nil, job cache should be nil", func() {
			newPGInfo.Name = ""
			InitCmAndCache(*newPGInfo, nil)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
		convey.Convey("when pg name is not nil, initCm success. job cache should not be nil", func() {
			mockInitCM := gomonkey.ApplyFunc(initCM,
				func(jobInfo constant.JobInfo) bool {
					return true
				})
			defer mockInitCM.Reset()
			InitCmAndCache(*newPGInfo, nil)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("when pg name is not nil, initCm failed. job cache should be nil", func() {
			mockInitCM := gomonkey.ApplyFunc(initCM,
				func(jobInfo constant.JobInfo) bool {
					return false
				})
			defer mockInitCM.Reset()
			InitCmAndCache(*newPGInfo, nil)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
	})
}

func TestGetJobBasicInfoByPodGroup(t *testing.T) {
	convey.Convey("test getJobBasicInfoByPodGroup success", t, func() {
		newPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		jobInfo := getJobBasicInfoByPG(*newPGInfo, nil)
		convey.So(jobInfo.Name, convey.ShouldEqual, jobName1)
	})
}

func TestUpdateCmAndCache(t *testing.T) {
	pgDemo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
	jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
	mockInitRankTableByPod := gomonkey.ApplyFunc(pod.ConstructRankTableByPod, getDemoRankTable)
	defer mockInitRankTableByPod.Reset()
	convey.Convey("test UpdateCmAndCache Job status", t, func() {
		convey.Convey("when jobInfo is nil, updateCM success. job status should be running", func() {
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return true
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, constant.JobInfo{}, *pgDemo, map[string]v1.Pod{})
			defer DeleteJobCache(jobUid1)
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.Status, convey.ShouldEqual, StatusJobRunning)
		})
		convey.Convey("when jobInfo is not nil, updateCM success. job status should be running", func() {
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return true
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, jobDemo, *pgDemo, map[string]v1.Pod{})
			defer DeleteJobCache(jobUid1)
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.Status, convey.ShouldEqual, StatusJobRunning)
		})
		convey.Convey("when jobInfo is not nil, updateCM failed. job should be nil", func() {
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return false
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, jobDemo, *pgDemo, map[string]v1.Pod{})
			defer DeleteJobCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestInitJobShareTorInfo(t *testing.T) {
	podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
	podDemo2 := getDemoPod(podName2, podNameSpace1, podUid2)
	podMapDemo := map[string]v1.Pod{
		podUid1: *podDemo1,
		podUid2: *podDemo2,
	}
	convey.Convey("test initJobShareTorInfo", t, func() {
		convey.Convey("when job framework is not pytorch. job masterAddr be nil", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			initJobShareTorInfo(&jobDemo, podMapDemo)
			convey.So(jobDemo.MasterAddr, convey.ShouldEqual, "")
		})
		convey.Convey("when job framework is pytorch. job masterAddr is masterIp", func() {
			mockGetEnvByPod := gomonkey.ApplyFunc(pod.GetEnvByPod,
				func(_ map[string]v1.Pod, _ string) string {
					return masterIp
				})
			defer mockGetEnvByPod.Reset()
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobDemo.Framework = ptFramework
			initJobShareTorInfo(&jobDemo, podMapDemo)
			convey.So(jobDemo.MasterAddr, convey.ShouldEqual, masterIp)
		})
		convey.Convey("when job framework is pytorch, jobType is vcjob. job masterAddr is masterIp", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobDemo.JobType = vcJobKind
			jobDemo.Framework = ptFramework
			serverHccl := constant.ServerHccl{
				ServerID: masterIp,
			}
			jobDemo.JobRankTable.ServerList = append(jobDemo.JobRankTable.ServerList, serverHccl)
			initJobShareTorInfo(&jobDemo, podMapDemo)
			convey.So(jobDemo.MasterAddr, convey.ShouldEqual, masterIp)
		})
	})
}

func getDemoPod(name, nameSpace, podUid string) *v1.Pod {
	p := &v1.Pod{}
	p.Name = name
	p.Namespace = nameSpace
	p.UID = types.UID(podUid)
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName1,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid1)}
	p.SetOwnerReferences([]metav1.OwnerReference{owner})
	return p
}

func getDemoPodGroup(jobName, nameSpace, jobUid string) *v1beta1.PodGroup {
	podGroupInfo := &v1beta1.PodGroup{}
	podGroupInfo.Name = jobName
	podGroupInfo.Namespace = nameSpace
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid)}
	podGroupInfo.SetOwnerReferences([]metav1.OwnerReference{owner})
	podGroupInfo.Spec.MinMember = pgMinMember2
	podGroupInfo.Spec.MinResources = &v1.ResourceList{"huawei/Ascend910": resource.Quantity{}}
	return podGroupInfo
}

func getDemoRankTable(_ map[string]v1.Pod, _ int) (constant.RankTable, int) {
	rankTable := constant.RankTable{
		Status:      StatusJobCompleted,
		ServerCount: pgMinMember2Str,
		Total:       1,
		ServerList:  []constant.ServerHccl{},
	}
	return rankTable, pgMinMember2
}

func TestGetJobServerInfoMap(t *testing.T) {
	convey.Convey("test GetJobServerInfoMap", t, func() {
		convey.Convey("when job cache is nil, jobServerInfoMap should be nil", func() {
			jobServerInfoMap := GetJobServerInfoMap()
			convey.So(len(jobServerInfoMap.InfoMap), convey.ShouldEqual, 0)
		})
		convey.Convey("when job cache length is 1, jobServerInfoMap length is 1", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			serverHccl := constant.ServerHccl{
				ServerID: masterIp,
			}
			deviceList := constant.Device{
				DeviceID: "",
			}
			serverHccl.DeviceList = append(serverHccl.DeviceList, deviceList)
			jobDemo.PreServerList = append(jobDemo.PreServerList, serverHccl)
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			jobServerInfoMap := GetJobServerInfoMap()
			convey.So(len(jobServerInfoMap.InfoMap), convey.ShouldEqual, 1)
		})
	})
}

func TestGetJobIsRunning(t *testing.T) {
	convey.Convey("test GetJobIsRunning", t, func() {
		jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when job cache is running, return true", func() {
			jobDemo.Status = StatusJobRunning
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			convey.So(GetJobIsRunning(jobUid1), convey.ShouldBeTrue)
		})
		convey.Convey("when job cache is not running, return false", func() {
			jobDemo.Status = StatusJobPending
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			convey.So(GetJobIsRunning(jobUid1), convey.ShouldBeFalse)
		})
	})
}

func TestGetJobIsExists(t *testing.T) {
	convey.Convey("test GetJobIsExists", t, func() {
		convey.Convey("when job cache is not exists, return false", func() {
			convey.So(GetJobIsExists(jobUid1), convey.ShouldBeFalse)
		})
		convey.Convey("when job cache is exists, return true", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			convey.So(GetJobIsExists(jobUid1), convey.ShouldBeTrue)
		})
	})
}

func TestFlushLastUpdateTime(t *testing.T) {
	convey.Convey("test FlushLastUpdateTime", t, func() {
		convey.Convey("when job cache is not exists, flush should be failed", func() {
			FlushLastUpdateTime(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("when job cache is exists, return true", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			FlushLastUpdateTime(jobUid1)
			jonInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(jonInfo.LastUpdatedCmTime, convey.ShouldNotBeZeroValue)
		})
	})
}

func TestIsInferenceJob(t *testing.T) {
	convey.Convey("test IsMindIeServerPod", t, func() {
		podInfo := getDemoPod(jobName1, jobNameSpace, podUid1)
		convey.Convey("when job is not mindie server job, return false", func() {
			convey.So(IsMindIeServerPod(*podInfo), convey.ShouldBeFalse)
		})
		convey.Convey("when job is mindie server job, return true", func() {
			podInfo.Labels = map[string]string{}
			podInfo.Labels[constant.MindIeJobIdLabelKey] = mindIeJobId
			podInfo.Labels[constant.MindIeAppTypeLabelKey] = constant.ServerAppType
			convey.So(IsMindIeServerPod(*podInfo), convey.ShouldBeTrue)
		})
	})
}

func TestGetMindIeServerJobDeviceInfoMap(t *testing.T) {
	convey.Convey("test GetMindIeServerJobAndUsedDeviceInfoMap", t, func() {
		convey.Convey("if there is an mindie server job on the current node, return mindie server jobId", func() {
			demoPod := getDemoPod(jobName1, jobNameSpace, podUid1)
			demoPod.Spec = v1.PodSpec{NodeName: nodeName1}
			patch := gomonkey.ApplyFuncReturn(GetAllJobCache, map[string]constant.JobInfo{
				jobUid1: getDemoJob(jobName1, jobNameSpace, jobUid1),
			}).ApplyFuncReturn(pod.GetPodByJobId, map[string]v1.Pod{
				podUid1: *demoPod,
			}).ApplyFuncReturn(IsMindIeServerPod, true)
			defer patch.Reset()
			jobInfoMap, deviceInfoMap := GetMindIeServerJobAndUsedDeviceInfoMap()
			convey.So(jobInfoMap, convey.ShouldNotBeEmpty)
			convey.So(deviceInfoMap, convey.ShouldNotBeEmpty)
		})
		convey.Convey("if there is no mindie server job on the current node, return mindie server jobId", func() {
			jobInfoMap, deviceInfoMap := GetMindIeServerJobAndUsedDeviceInfoMap()
			convey.So(jobInfoMap, convey.ShouldBeEmpty)
			convey.So(deviceInfoMap, convey.ShouldBeEmpty)
		})
	})
}
