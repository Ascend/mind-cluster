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

// Package slownode a series of function relevant to the fd-ol deployed in node
package slownode

import (
	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
)

// NodeProcessSlowNodeJob store the slow node job into the confMap in node
func NodeProcessSlowNodeJob(oldData, newData *slownode.SlowNodeJob, operator string) {
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got job cm data, oldObj: %+v, newObj: %+v, operator: %s",
		oldData, newData, operator)
	if newData.JobId == "" || newData.JobName != getLocalJobName() {
		hwlog.RunLog.Infof(
			"[FD-OL SLOWNODE]jobId is empty or jobName: %s does not match local jobName: %s, ignore it.",
			newData.JobName, getLocalJobName())
		return
	}
	slowNodeCtxMap := sm.GetSlowNodeCtxMap()
	slowNodeCtx, ok := slowNodeCtxMap.Get(newData.JobId)

	switch operator {
	case AddOperator:
		if ok {
			hwlog.RunLog.Warnf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) has been existed in SlowNodeCtxMap, ignore it",
				newData.JobName, newData.JobId)
			return
		}
		slowNodeCtx := sm.NewSlowNodeContext(newData, enum.Node)
		slowNodeCtxMap.Insert(newData.JobId, slowNodeCtx)
		if newData.SlowNode == SlowNodeOn {
			nodeStart(slowNodeCtx)
		}
	case UpdateOperator:
		if !ok {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) does not exist in SlowNodeCtxMap, ignore it.",
				newData.JobName, newData.JobId)
			return
		}
		slowNodeCtx.Update(newData)
		if !slowNodeCtx.IsRunning() && slowNodeCtx.Job.SlowNode == SlowNodeOn {
			nodeStart(slowNodeCtx)
		} else if slowNodeCtx.IsRunning() && slowNodeCtx.Job.SlowNode == SlowNodeOff {
			nodeStop(slowNodeCtx)
		}
	case DeleteOperator:
		if !ok {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) does not exist in SlowNodeCtxMap, ignore it.",
				newData.JobName, newData.JobId)
			return
		}
		if slowNodeCtx.IsRunning() {
			nodeStop(slowNodeCtx)
		}
		slowNodeCtxMap.Delete(newData.JobId)
	default:
		return
	}
}

func getLocalJobName() string {
	labels, err := global.K8sClient.GetLabels()
	if err != nil {
		return ""
	}
	return labels[keyJobName]
}

func nodeStart(slowNodeCtx *sm.SlowNodeContext) {
	slowNodeCtx.Start()
	watchingStartDataParse(slowNodeCtx)
	watchingStopDataParse(slowNodeCtx)
	watchingStartSlowNodeAlgo(slowNodeCtx)
	watchingStopSlowNodeAlgo(slowNodeCtx)
	slowNodeCtx.StartDataParse()
}

func nodeStop(slowNodeCtx *sm.SlowNodeContext) {
	slowNodeCtx.StopDataParse()
	slowNodeCtx.StopSlowNodeAlgo()
	slowNodeCtx.Stop()
}
