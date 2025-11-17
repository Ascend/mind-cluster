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

// Package plugin is using for A5 HuaWei Ascend pin affinity schedule frame.
package plugin

import (
	"strconv"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func (sJob *SchedulerJob) setTpBlock() {
	tpBlockStr, ok := sJob.Annotation[util.TpBlockAnnoKey]
	if !ok {
		klog.V(util.LogInfoLev).Infof("cannot find parameter (%s), using default value of tp-block",
			util.TpBlockAnnoKey)
		sJob.TpBlockNPUNum = util.LeastTpBlock
		return
	}
	tpBlockNum, err := strconv.Atoi(tpBlockStr)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("get job %s tp-block %s failed %v", sJob.Name, tpBlockStr, err)
		return
	}
	sJob.TpBlockNPUNum = tpBlockNum
}

func (sJob *SchedulerJob) updateSchedulerForA5Fields(oldFields A5Fields) {
	if sJob == nil {
		klog.V(util.LogErrorLev).Infof("update scheduler job for a5 fields failed: %s.", util.ArgumentError)
		return
	}
	sJob.WhetherBackToVspSchedule = oldFields.WhetherBackToVspSchedule
	sJob.TpBlock = oldFields.TpBlock
}
