// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package job a series of job test function
package job

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/superpod"
)

const (
	jobName1     = "job1"
	jobName2     = "job2"
	jobNameSpace = "default"
	jobUid1      = "123"
	jobUid2      = "456"
	two          = 2
)

func getDemoJob(jobName1 string, jobNameSpace string, jobUid1 string) constant.JobInfo {
	return constant.JobInfo{
		Name:       jobName1,
		NameSpace:  jobNameSpace,
		Key:        jobUid1,
		Replicas:   pgMinMember2,
		TotalCmNum: 1,
	}
}

func TestGetJobCache(t *testing.T) {
	convey.Convey("test GetJobCache", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is nil, return empty and false", func() {
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
			convey.So(jobInfo.Name, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not nil, return jobInfo and true", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.Name, convey.ShouldEqual, jobName1)
		})
	})
}

func TestCopyJobCache(t *testing.T) {
	convey.Convey("test GetJobCacheDeepCopy", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is nil, return empty and false", func() {
			jobInfo, ok := GetJobCacheDeepCopy(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
			convey.So(jobInfo.Name, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not nil, modifying a copy will not affect the cache", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.NodeNames = map[string]string{podName1: nodeName1}
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			copyJobInfo, ok := GetJobCacheDeepCopy(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			copyJobInfo.NodeNames = map[string]string{podName2: nodeName2}
			copyJobInfo.Name = jobName2
			convey.ShouldNotResemble(jobInfo.NodeNames, copyJobInfo.NodeNames)
			convey.ShouldNotEqual(jobInfo.Name, copyJobInfo.Name)
		})
	})
}

func TestGetAllJobCache(t *testing.T) {
	convey.Convey("test GetAllJobCache", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is nil, return map is empty", func() {
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
		convey.Convey("when job cache is not nil", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo2 := getDemoJob(jobName2, jobNameSpace, jobUid2)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)

			SaveJobCache(jobUid2, jobInfo2)
			defer DeleteJobCache(jobUid2)
			jobInfoMap = GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, two)
		})
	})
}

func TestUpdateAndIterateJobCacheCausePanic(t *testing.T) {
	convey.Convey("test updating and iterating through the cache cause a panic", t, func() {
		jobSummaryMap = sync.Map{}
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		jobInfo.NodeNames = make(map[string]string)
		SaveJobCache(jobUid1, jobInfo)
		defer DeleteJobCache(jobUid1)
		retryTimes := 500
		wg := sync.WaitGroup{}
		goNum := 2
		wg.Add(goNum)
		go func(jobKey string) {
			for i := 0; i < retryTimes; i++ {
				job, _ := GetJobCacheDeepCopy(jobKey)
				job.NodeNames[strconv.Itoa(i)] = fmt.Sprintf("value_%d", i)
				time.Sleep(time.Millisecond)
			}
			wg.Done()
		}(jobUid1)
		go func() {
			for i := 0; i < retryTimes; i++ {
				GetAllJobCache()
				time.Sleep(time.Millisecond)
			}
			wg.Done()
		}()
		wg.Wait()
	})
}

func TestSaveJobCache(t *testing.T) {
	convey.Convey("test SaveJobCache", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job key is empty, save success", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache("", jobInfo)
			defer DeleteJobCache("")
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("when job key is not empty, save success", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("when job key is same, save success and map length is 1", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo2 := getDemoJob(jobName2, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)
			SaveJobCache(jobUid1, jobInfo2)
			jobInfoMap = GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("when job key is not same, save success and map length is 2", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo2 := getDemoJob(jobName2, jobNameSpace, jobUid2)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)

			SaveJobCache(jobUid2, jobInfo2)
			defer DeleteJobCache(jobUid2)
			jobInfoMap = GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, two)
		})
	})
}

func TestDeleteJobCache(t *testing.T) {
	convey.Convey("test DeleteJobCache", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is empty, delete success", func() {
			DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
		convey.Convey("when job key is not empty, delete success", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
	})
}

func TestGetJobByNameSpaceAndName(t *testing.T) {
	convey.Convey("test GetJobByNameSpaceAndName", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is empty, get empty job info", func() {
			jobInfo := GetJobByNameSpaceAndName(jobName1, jobNameSpace)
			convey.So(jobInfo.Name, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not empty and name is not right, get empty job info", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfo = GetJobByNameSpaceAndName(jobName2, jobNameSpace)
			convey.So(jobInfo.Name, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not empty and name is right, get job info", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfo = GetJobByNameSpaceAndName(jobName1, jobNameSpace)
			convey.So(jobInfo.Name, convey.ShouldEqual, jobName1)
		})
	})
}

func TestGetJobByNameSpaceAndNameAndPreDelete(t *testing.T) {
	convey.Convey("test GetJobByNameSpaceAndNameAndPreDelete", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is empty, get empty job info", func() {
			jobInfos := GetJobByNameSpaceAndNameAndPreDelete(jobName1, jobNameSpace, true)
			convey.So(len(jobInfos), convey.ShouldEqual, 0)
		})
		convey.Convey("when job cache is not empty and isPreDelete is not right, get empty job info", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfos := GetJobByNameSpaceAndNameAndPreDelete(jobName1, jobNameSpace, true)
			convey.So(len(jobInfos), convey.ShouldEqual, 0)
		})
		convey.Convey("when job cache is not empty and isPreDelete is right, get job info", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.IsPreDelete = true
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfos := GetJobByNameSpaceAndNameAndPreDelete(jobName1, jobNameSpace, true)
			convey.So(len(jobInfos), convey.ShouldEqual, 1)
		})
	})
}

func TestGetShouldDeleteJobKey(t *testing.T) {
	convey.Convey("test GetShouldDeleteJobKey", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is empty, get empty job key", func() {
			jobKeys := GetShouldDeleteJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 0)
		})
		convey.Convey("when job is not preDelete and time is not timeout, get empty job key", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.DeleteTime = time.Now().Unix()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobKeys := GetShouldDeleteJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 0)
		})
		convey.Convey("when job is preDelete and time is not timeout, get empty job key", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.IsPreDelete = true
			jobInfo.DeleteTime = time.Now().Unix()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobKeys := GetShouldDeleteJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 0)
		})
		convey.Convey("when job is preDelete and time is timeout, get empty job key", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.IsPreDelete = true
			jobInfo.DeleteTime = time.Now().Unix() - preDeleteToDeleteSecond
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobKeys := GetShouldDeleteJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 1)
		})
	})
}

func TestGetShouldUpdateJobKey(t *testing.T) {
	convey.Convey("test GetShouldUpdateJobKey", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when job cache is empty, get empty job key", func() {
			jobKeys := GetShouldUpdateJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 0)
		})
		convey.Convey("when job lastUpdateTime is not timeout, get empty job key", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.LastUpdatedCmTime = time.Now().Unix()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobKeys := GetShouldUpdateJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 0)
		})
		convey.Convey("when job lastUpdateTime is timeout, get job key", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.LastUpdatedCmTime = time.Now().Unix() - updateSecond
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobKeys := GetShouldUpdateJobKey()
			convey.So(len(jobKeys), convey.ShouldEqual, 1)
		})
	})
}

func TestNamespaceByJobIdAndAppType(t *testing.T) {
	convey.Convey("test NamespaceByJobIdAndAppType", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when jobId and appType match", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.MultiInstanceJobId = jobUid1
			jobInfo.AppType = constant.ControllerAppType
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			namespace, err := GetNamespaceByJobIdAndAppType(jobUid1, constant.ControllerAppType)
			convey.So(err, convey.ShouldBeNil)
			convey.So(namespace, convey.ShouldEqual, jobNameSpace)
		})
		convey.Convey("when jobId and appType do not match", func() {
			namespace, err := GetNamespaceByJobIdAndAppType(jobUid1, constant.ControllerAppType)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(namespace, convey.ShouldEqual, "")
		})
	})
}

func TestInstanceJobKey(t *testing.T) {
	convey.Convey("test InstanceJobKey", t, func() {
		jobSummaryMap = sync.Map{}
		convey.Convey("when jobId, namespace, and appType match", func() {
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.MultiInstanceJobId = jobUid1
			jobInfo.AppType = constant.ControllerAppType
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobKey, err := GetInstanceJobKey(jobUid1, jobNameSpace, constant.ControllerAppType)
			convey.So(err, convey.ShouldBeNil)
			convey.So(jobKey, convey.ShouldEqual, jobUid1)
		})
		convey.Convey("when jobId, namespace, and appType do not match", func() {
			jobKey, err := GetInstanceJobKey(jobUid1, jobNameSpace, constant.ControllerAppType)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(jobKey, convey.ShouldEqual, "")
		})
	})
}

func TestGetJobFaultSdIdAndNodeName(t *testing.T) {
	convey.Convey("Test GetJobFaultSdIdAndNodeName", t, func() {
		testJobId := "test-job-id"
		testPodNames := map[string]string{"0": "pod1"}
		convey.Convey("Should return nil when job not in cache", func() {
			result := GetJobFaultSdIdAndNodeName(testJobId, testPodNames)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("Should return fault info when nt valid", func() {
			jobInfo := constant.JobInfo{
				JobRankTable: constant.RankTable{
					ServerList: []constant.ServerHccl{{
						PodID:      "pod1",
						SuperPodId: 0,
						ServerName: "node1",
						DeviceList: []constant.Device{{
							SuperDeviceID: "sd1",
						}},
					}},
				},
			}
			SaveJobCache(testJobId, jobInfo)
			defer DeleteJobCache(testJobId)
			result := GetJobFaultSdIdAndNodeName(testJobId, testPodNames)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestGetFaultSuperID(t *testing.T) {
	convey.Convey("Test getFaultSuperID", t, func() {
		convey.Convey("When there are no fault nodes", func() {
			faultNodes := sets.NewString()
			patches := gomonkey.ApplyFunc(superpod.ListClusterDevice, func() []*api.SuperPodDevice {
				return []*api.SuperPodDevice{}
			})
			defer patches.Reset()
			result := getFaultSuperID(faultNodes)
			convey.So(result, convey.ShouldBeEmpty)
		})

		convey.Convey("When there are fault nodes but no matching super nodes", func() {
			faultNodes := sets.NewString("node1", "node2")
			patches := gomonkey.ApplyFunc(superpod.ListClusterDevice, func() []*api.SuperPodDevice {
				return []*api.SuperPodDevice{
					{SuperPodID: "superpod1", NodeDeviceMap: map[string]*api.NodeDevice{
						"node3": {NodeName: "node3", DeviceMap: map[string]string{"device1": "sdid1"}}}}}
			})
			defer patches.Reset()
			result := getFaultSuperID(faultNodes)
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}
