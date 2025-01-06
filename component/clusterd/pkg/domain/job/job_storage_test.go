// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package job a series of job test function
package job

import (
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

const (
	jobName1     = "job1"
	jobName2     = "job2"
	jobNameSpace = "default"
	jobUid1      = "123"
	jobUid2      = "456"
	two = 2
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
