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

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"strconv"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
)

func init() {
	reSchedulerCache = newReSchedulerCache()
}

func newReSchedulerCache() *DealReSchedulerCache {
	return &DealReSchedulerCache{
		FaultNodes: map[string]*FaultNode{},
		FaultJobs:  map[api.JobID]*FaultJob{},
	}
}

// checkGraceDeleteTimeValid used by GetGraceDeleteTime for validity checking
func checkGraceDeleteTimeValid(overTime int64) bool {
	if overTime < minGraceOverTime || overTime > maxGraceOverTime {
		klog.V(util.LogErrorLev).Infof("GraceOverTime value should be range [2, 3600], configured is [%d], "+
			"GraceOverTime will not be changed", overTime)
		return false
	}
	// use user's configuration to set grace over time
	klog.V(util.LogInfoLev).Infof("set GraceOverTime to new value [%d].", overTime)
	return true
}

func getGraceDeleteTime(Conf []config.Configuration) int64 {
	klog.V(util.LogInfoLev).Infof("enter GetGraceDeleteTime ...")
	defer klog.V(util.LogInfoLev).Infof("leave GetGraceDeleteTime ...")
	if len(Conf) == 0 {
		klog.V(util.LogErrorLev).Infof("GetGraceDeleteTime failed: %s, no conf", util.ArgumentError)
		return DefaultGraceOverTime
	}
	// Read configmap
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, Conf)
	if err != nil {
		klog.V(util.LogErrorLev).Info("cannot get configuration, GraceOverTime will not be changed.")
		return DefaultGraceOverTime
	}
	// get grace over time by user configuration
	overTimeStr, ok := configuration.Arguments[GraceOverTimeKey]
	if !ok {
		klog.V(util.LogErrorLev).Info("set GraceOverTime failed and will not be changed, " +
			"key grace-over-time doesn't exists.")
		return DefaultGraceOverTime
	}
	overTime, err := strconv.ParseInt(overTimeStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("set GraceOverTime failed and will not be changed, "+
			"grace-over-time is invalid [%s].", util.SafePrint(overTimeStr))
		return DefaultGraceOverTime
	}
	// check time validity
	if !checkGraceDeleteTimeValid(overTime) {
		return DefaultGraceOverTime
	}
	return overTime
}

// JudgePublicFaultInReason return fTask has public fault
func JudgePublicFaultInReason(fTask *miniFaultTask) bool {
	if fTask == nil {
		return false
	}
	for _, reason := range fTask.Reason {
		if reason.FaultType == PublicFaultType {
			return true
		}
	}
	return false
}
