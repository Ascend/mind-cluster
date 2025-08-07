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

// Package sdk provides node parse
package sdk

import (
	"errors"
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

// JobPipeline Job顺序执行
type JobPipeline struct {
	jobMap sync.Map
	mutex  sync.Mutex
}

// StartJob 启动Job
func (jp *JobPipeline) StartJob(cg config.DataParseModel) {
	jp.mutex.Lock()
	defer jp.mutex.Unlock()

	if existJob, err := jp.GetJobInfo(cg.JobId); err == nil {
		if existJob.JobStatus == constants.JobStoppingStatus {
			hwlog.RunLog.Warnf("[SLOWNODE PARSE]Job %s is in the %s state and waiting to be stopped",
				cg.JobId, existJob.JobStatus)
			jp.mutex.Unlock()
			waitStop(existJob)
			jp.mutex.Lock()
		} else if existJob.JobStatus == constants.JobRunStatus {
			hwlog.RunLog.Warnf("[SLOWNODE PARSE]Job %s is in the %s state and cannot be executed repeatedly",
				cg.JobId, existJob.JobStatus)
			return
		}
	}

	jobInfo := &model.ParseJobInfo{
		JobName:       cg.JobName,
		JobId:         cg.JobId,
		JobStatus:     constants.JobRunStatus,
		StopParseFlag: make(chan struct{}),
		JobWg:         &sync.WaitGroup{},
		StopWg:        &sync.WaitGroup{},
		TimeStamp:     cg.JobStartTime,
	}
	jp.jobMap.Store(cg.JobId, jobInfo)
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Succeeded in initializing parse job: %s", cg.JobId)

	jobInfo.StopWg.Add(1)
	go DealParseJob(cg, jobInfo)
}

// StopJob 停止Job
func (jp *JobPipeline) StopJob(cg config.DataParseModel) {
	jp.mutex.Lock()
	defer jp.mutex.Unlock()

	jobInfo, err := jp.GetJobInfo(cg.JobId)
	if err != nil {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]Get stop job failed, can not stop job: %v", err)
		return
	}

	if jobInfo.JobStatus != constants.JobRunStatus {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]The same job %s cannot be stopped repeatedly, job status: %s",
			jobInfo.JobId, jobInfo.JobStatus)
		return
	}

	// 发送job停止信号
	close(jobInfo.StopParseFlag)
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Successed in closing chan, job id is: %s", jobInfo.JobId)
	jobInfo.JobStatus = constants.JobStoppingStatus

	// 释放锁，让任务有机会执行完成
	jp.mutex.Unlock()
	// 等待任务真正结束
	jobInfo.StopWg.Wait()
	jobInfo.JobStatus = constants.JobStopStatus
	// 重新获取锁以更新状态
	jp.mutex.Lock()

}

// RestartJob 重启指定ID的Job
func (jp *JobPipeline) RestartJob(cg config.DataParseModel) {
	jp.mutex.Lock()
	if _, err := jp.GetJobInfo(cg.JobId); err != nil {
		hwlog.RunLog.Infof("[SLOWNODE PARSE]The job doesn't exist, restart the job: %s", cg.JobId)
		jp.mutex.Unlock()
		StartParse(cg)
		return
	}
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Stopping running job: %s", cg.JobId)
	jp.mutex.Unlock()
	StopParse(cg)
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Successed in stopping job and restarting the job: %s", cg.JobId)
	StartParse(cg)
}

// GetJobInfo 获取对应jobId的任务信息
func (jp *JobPipeline) GetJobInfo(jobId string) (*model.ParseJobInfo, error) {
	data, exists := jp.jobMap.Load(jobId)
	if !exists {
		return nil, fmt.Errorf("failed to get job id: %s", jobId)
	}
	jobInfo, ok := data.(*model.ParseJobInfo)
	if !ok {
		return nil, errors.New("job is not of type *model.ParseJobInfo")
	}
	return jobInfo, nil
}

// GetStopChan 获取对应jobId的停止信息
func (jp *JobPipeline) GetStopChan(jobId string) chan struct{} {
	jobInfo, err := jp.GetJobInfo(jobId)
	if err != nil {
		hwlog.RunLog.Warnf("[SLOWNODE PARSE]Get stop chan failed: %v", err)
		return nil
	}
	return jobInfo.StopParseFlag
}

func waitStop(existJob *model.ParseJobInfo) {
	stopFunc := func() (bool, error) {
		return existJob.JobStatus == constants.JobStopStatus, nil
	}

	err := utils.Poller(stopFunc, constants.StopPoll, constants.WaitStopTime, nil)
	if err != nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]Failed to stop job:", err)
	} else {
		hwlog.RunLog.Info("[SLOWNODE PARSE]Succeeded in stopping the executing job, start to execute a new job")
	}

}
