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

// Package cluster is a series of function to process the data in job_summary
package cluster

import (
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

const (
	// some keys relevent to the job_summary
	keyJobId     = "job_id"
	keyJobName   = "job_name"
	keyJobStatus = "job_status"
	keyHccl      = "hccl.json"
)

func processJobSummaryData(data any, operator watch.EventType) {
	var jobSummary, err = convertCMToJobSummary(data)
	if err != nil {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]convert cm data: %v to job summary failed: %v", data, err)
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got job summary data, operator: %s, data: %+v", operator, jobSummary)
	// query context from local contextMap
	var key = fmt.Sprintf("%s/%s", jobSummary.Namespace, jobSummary.JobName)
	ctx, ok := slownodejob.GetJobCtxMap().Get(key)
	if !ok {
		hwlog.RunLog.Warnf("[FD-OL SLOWNODE]no slow node context found, key: %s", key)
		return
	}
	if ctx.Job.JobId != jobSummary.JobId {
		// case 1: no jobId in ctx, update it -> start slow node job
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) detected jobId updated, update it to: %s ",
			ctx.Job.JobName, ctx.Job.JobId, jobSummary.JobId)
		ctx.Job.JobId = jobSummary.JobId
	}
	ctx.UpdateTrainingJobStatus(jobSummary.JobStatus)
	var j = jobProcessor{ctx: ctx, job: ctx.Job}
	switch operator {
	case watch.Added:
		ctx.Update(&slownode.Job{
			SlowNode: ctx.Job.SlowNode,
			Servers:  jobSummary.Servers,
		})
		if jobSummary.JobStatus == enum.IsRunning {
			// case 1: no jobId in ctx, update it -> start slow node job
			hwlog.RunLog.Infof(
				"[FD-OL SLOWNODE]training job status is %s, starting slow node job(name=%s, jobId=%s)",
				enum.IsRunning, jobSummary.JobName, jobSummary.JobId)
			j.start()
		}
	case watch.Modified:
		processJobSummaryUpdate(ctx, jobSummary)
	case watch.Deleted:
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]job summary is deleted, stopping slow node job(name= %s, jobId= %s)",
			jobSummary.JobName, jobSummary.JobId)
		j.stop()
	default:
		return
	}
}

func processJobSummaryUpdate(ctx *slownodejob.JobContext, jobSummary *slownode.JobSummary) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, jobId=%s)", jobSummary.JobName, jobSummary.JobId)
	var newJob = &slownode.Job{
		SlowNode: ctx.Job.SlowNode,
		Servers:  jobSummary.Servers,
	}
	var j = jobProcessor{ctx: ctx, job: ctx.Job}
	switch jobSummary.JobStatus {
	case enum.IsPending:
		hwlog.RunLog.Infof("%s detected training job is pending", logPrefix)
		// case: job_status is pending -> update and stop
		ctx.Update(newJob)
		j.stop()
	case enum.IsFailed:
		hwlog.RunLog.Infof("%s detected training job is failed, stop job", logPrefix)
		// case: job_status is failed -> stop job
		j.stop()
	case enum.IsRunning:
		if !ctx.IsRunning() {
			hwlog.RunLog.Infof("%s detected training job is running, but job is not running", logPrefix)
			// case: job_status is running, job is not running -> update job, start depends on SlowNode
			ctx.Update(newJob)
			j.start()
			return
		}
		// case: job_status is running, job is running, rankIds changes -> stop then start job
		if !common.AreServersEqual(ctx.Job.Servers, jobSummary.Servers) {
			hwlog.RunLog.Infof("%s detected training job is running, rankIds changed, stop then start job", logPrefix)
			ctx.Update(newJob)
			j.stop()
			j.start()
		}
	// case: job_status is complete -> delete job
	case enum.IsCompleted:
		hwlog.RunLog.Infof("%s detected training job is complete, delete job", logPrefix)
		j.delete()
	default:
		return
	}
}

// convertCMToJobSummary convert config map data to job summary
func convertCMToJobSummary(data any) (*slownode.JobSummary, error) {
	cm, ok := data.(*corev1.ConfigMap)
	if !ok {
		return nil, errors.New("convert to ConfigMap object failed")
	}
	var jobSummary = &slownode.JobSummary{Namespace: cm.Namespace}
	errMsg := fmt.Sprintf("ConfigMap %s/%s does not contain", cm.Namespace, cm.Name)
	if cm.Data[keyJobId] == "" {
		return jobSummary, fmt.Errorf("%s %s", errMsg, keyJobId)
	}
	jobSummary.JobId = cm.Data[keyJobId]
	if cm.Data[keyJobName] == "" {
		return jobSummary, fmt.Errorf("%s %s", errMsg, keyJobName)
	}
	jobSummary.JobName = cm.Data[keyJobName]
	if cm.Data[keyJobStatus] == "" {
		return jobSummary, fmt.Errorf("%s %s", errMsg, keyJobStatus)
	}
	jobSummary.JobStatus = cm.Data[keyJobStatus]
	if cm.Data[keyHccl] == "" {
		return jobSummary, fmt.Errorf("%s %s", errMsg, keyHccl)
	}
	// Unmarshal the HCCL data
	var hcclData = struct {
		ServerList []struct {
			ServerId string `json:"server_id"`
			ServerSn string `json:"server_sn"`
			Device   []struct {
				RankId string `json:"rank_id"`
			} `json:"device"`
		} `json:"server_list"`
	}{}
	if err := json.Unmarshal([]byte(cm.Data[keyHccl]), &hcclData); err != nil {
		return jobSummary, fmt.Errorf("failed to unmarshal HCCL data: %v", err)
	}
	jobSummary.Servers = make([]slownode.Server, len(hcclData.ServerList))
	for i, server := range hcclData.ServerList {
		var rankIds = make([]string, len(server.Device))
		for j, device := range server.Device {
			rankIds[j] = device.RankId
		}
		jobSummary.Servers[i] = slownode.Server{
			Sn:      server.ServerSn,
			Ip:      server.ServerId,
			RankIds: rankIds,
		}
	}
	return jobSummary, nil
}
