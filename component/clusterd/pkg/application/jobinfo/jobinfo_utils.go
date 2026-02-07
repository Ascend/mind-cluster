/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package jobinfo is used to return job info by subscribe

package jobinfo

import (
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/job"
)

const (
	ptFramework        = "pytorch"
	jobUpdateChanCache = 10
	// StatusJobFail is the failed job status
	StatusJobFail = "failed"
	// StatusJobCompleted is the complete job status
	StatusJobCompleted = "complete"
)

var jobUpdateChan = make(chan job.JobSummarySignal, jobUpdateChanCache)

// SendJobInfoSignal send jobInfo signal
func SendJobInfoSignal(jobSignal job.JobSummarySignal) {
	select {
	case jobUpdateChan <- jobSignal:
	default:
		hwlog.RunLog.Infof("msg %#v has been dropped", jobSignal)
	}
}

func calcJobNPUNum(jobInfo constant.JobInfo) int {
	jobNPUNum := 0
	for _, serverHccl := range jobInfo.JobRankTable.ServerList {
		jobNPUNum += len(serverHccl.DeviceList)
	}
	return jobNPUNum
}

// BuildJobSignalFromJobInfo build jobSignal from jobInfo
func BuildJobSignalFromJobInfo(jobInfo constant.JobInfo, hccl, operator string) job.JobSummarySignal {
	jobSignal := job.JobSummarySignal{
		Uuid:      string(uuid.NewUUID()),
		JobId:     jobInfo.Key,
		JobName:   jobInfo.Name,
		Namespace: jobInfo.NameSpace,
		FrameWork: jobInfo.Framework,
		JobStatus: jobInfo.Status,
		Time:      strconv.Itoa(int(jobInfo.AddTime)),
		CmIndex:   "0",
		Operator:  operator,
		Total:     strconv.Itoa(jobInfo.TotalCmNum),
		HcclJson:  hccl,
		Sid:       jobInfo.Sid,
	}
	if jobInfo.Framework == ptFramework {
		jobSignal.SharedTorIp = jobInfo.SharedTorIp
		jobSignal.MasterAddr = jobInfo.MasterAddr
	}
	if jobInfo.Status == StatusJobFail || jobInfo.Status == StatusJobCompleted {
		jobSignal.DeleteTime = strconv.Itoa(int(time.Now().Unix()))
	}
	return jobSignal
}
