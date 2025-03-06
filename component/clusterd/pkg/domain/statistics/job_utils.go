// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"time"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

// UpdateStatistic update job statistic info
func UpdateStatistic(jobStc constant.JobStatistic, jobInfo constant.JobInfo) constant.JobStatistic {
	jobStc.Status = jobInfo.Status
	switch jobInfo.Status {
	case job.StatusJobPending:
		// job scheduling
		return jobStc

	case job.StatusJobRunning:
		nowTime := time.Now().Unix()
		jobStc.StopTime = 0
		// job start running success
		if jobStc.PodFirstRunningTime == 0 {
			jobStc.PodFirstRunningTime = nowTime
			cardNum := 0
			for _, serverList := range jobInfo.PreServerList {
				cardNum += len(serverList.DeviceList)
			}
			jobStc.CardNums = int64(cardNum)
			return jobStc
		}
		// job recover success
		if jobStc.PodLastFaultTime > jobStc.PodLastRunningTime {
			jobStc.PodLastRunningTime = nowTime
			return jobStc
		}

	case job.StatusJobCompleted:
		jobStc.StopTime = time.Now().Unix()
		return jobStc

	case job.StatusJobFail:
		if jobStc.PodLastRunningTime >= jobStc.PodLastFaultTime {
			jobStc.PodLastFaultTime = time.Now().Unix()
			jobStc.PodFaultTimes += 1
			if jobInfo.IsPreDelete {
				jobStc.StopTime = jobStc.PodLastFaultTime
			}
			return jobStc
		}
	default:
		return jobStc
	}
	return jobStc
}

// InitStatistic init statistic
func InitStatistic(jobInfo constant.JobInfo, jobID string) constant.JobStatistic {
	jobStc := constant.JobStatistic{}
	jobStc.CustomJobID = jobInfo.CustomJobID
	jobStc.K8sJobID = jobID
	jobStc.Status = jobInfo.Status
	jobStc.CardNums = 0
	jobStc.PodFirstRunningTime = 0
	jobStc.StopTime = 0
	jobStc.PodLastRunningTime = 0
	jobStc.PodLastFaultTime = 0
	jobStc.PodFaultTimes = 0
	jobStc.Name = jobInfo.Name
	jobStc.NameSpace = jobInfo.NameSpace
	return jobStc
}
