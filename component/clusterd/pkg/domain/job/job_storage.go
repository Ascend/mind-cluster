// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job cache function
package job

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	preDeleteToDeleteSecond = 60
	updateSecond            = 3600
)

var jobSummaryMap sync.Map

// GetJobCache get job cache info
func GetJobCache(jobKey string) (constant.JobInfo, bool) {
	jobInfo, ok := jobSummaryMap.Load(jobKey)
	if !ok {
		return constant.JobInfo{}, ok
	}
	return jobInfo.(constant.JobInfo), ok
}

// GetAllJobCache get all job cache info
func GetAllJobCache() map[string]constant.JobInfo {
	allJob := map[string]constant.JobInfo{}
	jobSummaryMap.Range(func(key, value any) bool {
		jobKey, ok := key.(string)
		if !ok {
			return true
		}
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		allJob[jobKey] = jobInfo
		return true
	})
	newJob := new(map[string]constant.JobInfo)
	hwlog.RunLog.Infof("get all job cache, allJob: %v", allJob)
	err := util.DeepCopy(newJob, allJob)
	if err != nil {
		hwlog.RunLog.Errorf("copy job failed, err: %v", err)
	}
	return *newJob
}

// SaveJobCache save job cache info
func SaveJobCache(jobKey string, jobInfo constant.JobInfo) {
	jobSummaryMap.Store(jobKey, jobInfo)
}

// DeleteJobCache delete job cache info
func DeleteJobCache(jobKey string) {
	hwlog.RunLog.Infof("delete job cache, jobKey: %v", jobKey)
	jobSummaryMap.Delete(jobKey)
}

// GetJobByNameSpaceAndName get job by job name and nameSpace
func GetJobByNameSpaceAndName(name, nameSpace string) constant.JobInfo {
	ji := constant.JobInfo{}
	jobSummaryMap.Range(func(_, value any) bool {
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if jobInfo.Name == name && jobInfo.NameSpace == nameSpace {
			ji = jobInfo
			return false
		}
		return true
	})
	return ji
}

// GetShouldDeleteJobKey get should delete job key
func GetShouldDeleteJobKey() []string {
	var allJob = make([]string, 0)
	nowTime := time.Now().Unix()
	jobSummaryMap.Range(func(key, value any) bool {
		jobKey, ok := key.(string)
		if !ok {
			return true
		}
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if jobInfo.IsPreDelete && nowTime-jobInfo.DeleteTime >= preDeleteToDeleteSecond {
			allJob = append(allJob, jobKey)
		}
		return true
	})
	return allJob
}

// GetShouldUpdateJobKey  get should update job key
func GetShouldUpdateJobKey() []string {
	allJobKey := make([]string, 0)
	nowTime := time.Now().Unix()
	jobSummaryMap.Range(func(key, value any) bool {
		jobKey, ok := key.(string)
		if !ok {
			return true
		}
		jobInfo, ok := value.(constant.JobInfo)
		if !ok {
			return true
		}
		if nowTime-jobInfo.LastUpdatedCmTime >= updateSecond {
			allJobKey = append(allJobKey, jobKey)
		}
		return true
	})
	return allJobKey
}
