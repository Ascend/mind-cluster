/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package cluster provides some func to process the callback data from algo and data parse in cluster
package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

const (
	maxDegradationCount int = 5

	// slow node report
	faultCode         = "110001010"
	recoveryFaultCode = "100001011"
)

// AlgoCallbackProcessor process the callback data from slow node algo deployed in cluster
func AlgoCallbackProcessor(message string) {
	var data = map[string]map[string]any{}
	if err := json.Unmarshal([]byte(message), &data); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json unmarshal data: %s failed: %v", message, err)
		return
	}
	var result = slownode.ClusterAlgoResult{}
	if err := common.ConvertMaptoStruct(data, &result); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]convert map data: %s to struct failed: %v", data, err)
		return
	}
	ctx, ok := slownodejob.GetJobCtxMap().GetByJobId(result.JobId)
	if !ok || !ctx.IsRunning() {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not exit or not running, exit slow node algo callback process",
			result.JobName, result.JobId)
		return
	}
	profilingDataProcessor(ctx, &result)
}

// ParallelGroupInfoCallbackProcessor process the callback data from data parse
func ParallelGroupInfoCallbackProcessor(message string) {
	var result = slownode.DataParseResult{}
	if err := json.Unmarshal([]byte(message), &result); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]error parsing dataParseResult JSON: %v", err)
		return
	}
	ctx, ok := slownodejob.GetJobCtxMap().GetByJobId(result.JobId)
	if !ok || !ctx.IsRunning() {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not exit or not running, exit slow node algo callback process",
			result.JobName, result.JobId)
		return
	}
	// ClusterStep2 means merge parallel group info
	if result.IsFinished && ctx.Step() == slownodejob.ClusterStep2 {
		algo.NewController(ctx).Start()
	}
}

func profilingDataProcessor(ctx *slownodejob.JobContext, result *slownode.ClusterAlgoResult) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, jobId=%s)", ctx.Job.JobName, ctx.Job.JobId)
	if result.IsSlow != constants.IsDegradation {
		if ctx.IsDegradation {
			// this means the degradation is recovery after the heavy profiling
			hwlog.RunLog.Infof("%s detected node is not degradation, set the sign of degradation is false", logPrefix)
			ctx.IsDegradation = false
			reportSlowNode(ctx, result)
		}
		if ctx.IsStartedHeavyProfiling() {
			// this means algo result is not slow during the heavy profiling
			hwlog.RunLog.Infof("%s detected node is not degradation, stop heavy profiling", logPrefix)
			ctx.StopHeavyProfiling()
		}
		return
	}
	if ctx.IsDegradation {
		// this means got the degradation result, but the ctx has been record it
		hwlog.RunLog.Infof("%s in cluster got degradation, no need to start heavy profiling", logPrefix)
		return
	}

	ctx.AddAlgoRecord(result)
	if !ctx.IsStartedHeavyProfiling() {
		hwlog.RunLog.Infof("%s detected node is degradation, start heavy profiling", logPrefix)
		ctx.StartHeavyProfiling()
		return
	}
	degradation := slicetool.Filter(ctx.AlgoRes, func(record *slownode.ClusterAlgoResult) bool {
		return record.IsSlow == constants.IsDegradation
	})
	if len(degradation) >= maxDegradationCount {
		hwlog.RunLog.Infof("%s in cluster got degradation", logPrefix)
		ctx.IsDegradation = true
		reportSlowNode(ctx, result)
		ctx.StopHeavyProfiling()
	}
}

func reportSlowNode(ctx *slownodejob.JobContext, result *slownode.ClusterAlgoResult) {
	grpcClient, err := grpc.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]get grpc client failed: %s", err)
		return
	}
	var fc = recoveryFaultCode
	if result.IsSlow == constants.IsDegradation {
		fc = faultCode
	}
	hostname, err := os.Hostname()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) got hostname failed: %s",
			ctx.Job.JobName, ctx.Job.JobId, err)
	}
	var deviceIds = make([]int32, len(result.SlowCalculateRanks))
	for i, v := range result.SlowCalculateRanks {
		deviceIds[i] = int32(v)
	}
	var fault = []*pubfault.Fault{
		{
			FaultId:       ctx.Job.JobId,
			FaultType:     "Node",
			FaultCode:     fc,
			FaultTime:     time.Now().UnixMilli(),
			Assertion:     "once",
			FaultLocation: map[string]string{},
			Influence: []*pubfault.PubFaultInfo{
				{
					NodeName:  hostname,
					DeviceIds: deviceIds,
				},
			},
			Description: "",
		}}
	if err := grpcClient.ReportFault(fault); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) report slow node detection failed: %s",
			ctx.Job.JobName, ctx.Job.JobId, err)
	} else {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) report slow node detection: %+v successfully",
			ctx.Job.JobName, ctx.Job.JobId, fault)
	}
}
