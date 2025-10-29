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
	"fmt"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

const (
	// some keys relevent to the job_summary
	add = "add"
	del = "delete"
)

func jobSummaryProcessor(jobSummary *model.JobSummary) {
	if jobSummary == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]job summary is nil")
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got job summary data, operator: %s, data: %+v", jobSummary.Operator, jobSummary)
	if err := common.JobIdValidator(jobSummary.JobId); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]invalid jobId: %s, err: %v", jobSummary.JobId, err)
		return
	}
	// query context from local contextMap
	var key = fmt.Sprintf("%s/%s", jobSummary.Namespace, jobSummary.JobName)
	ctx, ok := slownodejob.GetJobCtxMap().Get(key)
	if !ok {
		hwlog.RunLog.Warnf("[FD-OL SLOWNODE]no slow node context found, key: %s", key)
		return
	}
	if ctx == nil || ctx.Job == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]invalid nil context or job, key: %s", key)
		return
	}
	if ctx.Job.JobId != jobSummary.JobId {
		// case 1: no jobId in ctx, update it -> start slow node job
		hwlog.RunLog.Infof("%s detected jobId updated, update it to: %s ", ctx.LogPrefix(), jobSummary.JobId)
		ctx.Job.JobId = jobSummary.JobId
	}
	ctx.UpdateTrainingJobStatus(jobSummary.JobStatus)
	var j = jobProcessor{ctx: ctx, job: ctx.Job}
	switch jobSummary.Operator {
	case add:
		jobStatusProcessor(ctx, jobSummary)
	case del:
		hwlog.RunLog.Infof("%s job summary is deleted, stopping slow node job", ctx.LogPrefix())
		j.delete()
	default:
		return
	}
}

func serversGenerator(hcclJson model.HcclJson) []slownode.Server {
	servers := make([]slownode.Server, len(hcclJson.ServerList))
	for i, server := range hcclJson.ServerList {
		var rankIds = make([]string, len(server.Device))
		for j, device := range server.Device {
			rankIds[j] = device.RankId
		}
		servers[i] = slownode.Server{
			Sn:      server.ServerSn,
			Ip:      server.ServerId,
			RankIds: rankIds,
		}
	}
	return servers
}

func jobStatusProcessor(ctx *slownodejob.JobContext, jobSummary *model.JobSummary) {
	servers := serversGenerator(jobSummary.HcclJson)
	var newJob = &slownode.Job{
		SlowNode: ctx.Job.SlowNode,
		Servers:  servers,
	}
	if len(ctx.Job.Servers) == 0 {
		ctx.Update(newJob)
	}
	var j = jobProcessor{ctx: ctx, job: ctx.Job}
	switch jobSummary.JobStatus {
	case enum.IsPending:
		hwlog.RunLog.Infof("%s detected training job is pending", ctx.LogPrefix())
		// case: job_status is pending -> update and stop
		ctx.Update(newJob)
		j.stop()
	case enum.IsFailed:
		hwlog.RunLog.Infof("%s detected training job is failed, stop job", ctx.LogPrefix())
		// case: job_status is failed -> stop job
		j.stop()
	case enum.IsRunning:
		if !ctx.IsRunning() {
			hwlog.RunLog.Infof("%s detected training job is running, but job is not running", ctx.LogPrefix())
			// case: job_status is running, job is not running -> update job, start depends on SlowNode
			ctx.Update(newJob)
			j.start()
			return
		}
		// case: job_status is running, job is running, rankIds changes -> stop then start job
		if !common.AreServersEqual(ctx.Job.Servers, servers) {
			hwlog.RunLog.Infof("%s detected training job is running, rankIds changed, stop then start job",
				ctx.LogPrefix())
			ctx.Update(newJob)
			j.stop()
			j.start()
		}
	// case: job_status is complete -> delete job
	case enum.IsCompleted:
		hwlog.RunLog.Infof("%s detected training job is complete, delete job", ctx.LogPrefix())
		j.delete()
	default:
		return
	}
}
