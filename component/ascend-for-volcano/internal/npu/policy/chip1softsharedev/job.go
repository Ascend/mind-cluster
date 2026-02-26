/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"fmt"
	"strconv"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func (tp *chip1softsharedev) getSoftShareDevResource() (softShareDevResource, error) {
	aicoreQuotaStr, hasAicoreQuota := tp.Label[util.SchedulerSoftShareDevAicoreQuotaKey]
	hbmQuotaStr, hasHbmQuota := tp.Label[util.SchedulerSoftShareDevHbmQuotaKey]
	schedulingPolicy, hasSchedulingPolicy := tp.Label[util.SchedulerSoftShareDevPolicyKey]
	if !hasAicoreQuota || !hasHbmQuota || !hasSchedulingPolicy {
		err := fmt.Errorf("%s check share device job(%s) valid failed, hasAicoreQuota: %v, hasHbmQuota: %v, "+
			"hasSchedulingPolicy: %v", tp.GetPluginName(), tp.Name, hasAicoreQuota, hasHbmQuota, hasSchedulingPolicy)
		return softShareDevResource{}, err
	}
	aicoreQuota, err := strconv.Atoi(aicoreQuotaStr)
	if err != nil {
		err := fmt.Errorf("%s check share device job(%s) valid failed, aicoreQuota: %s convert to int err: %v",
			tp.GetPluginName(), tp.Name, aicoreQuotaStr, err)
		return softShareDevResource{}, err
	}
	hbmQuota, err := strconv.Atoi(hbmQuotaStr)
	if err != nil {
		err := fmt.Errorf("%s check share device job(%s) valid failed, hbmQuota: %s convert to int err: %v",
			tp.GetPluginName(), tp.Name, hbmQuotaStr, err)
		return softShareDevResource{}, err
	}
	return softShareDevResource{
		aicoreQuota:      aicoreQuota,
		hbmQuota:         hbmQuota,
		schedulingPolicy: schedulingPolicy,
	}, nil
}

func (tp *chip1softsharedev) checkSoftShareDevResource(reqResource softShareDevResource) error {
	if reqResource.aicoreQuota > util.MaxAicoreQuota || reqResource.aicoreQuota < util.MinAicoreQuota {
		return fmt.Errorf("%s check share device job(%s) valid failed, aicoreQuota: %v not in range [1,100]",
			tp.GetPluginName(), tp.Name, reqResource.aicoreQuota)
	}
	if tp.ReqNPUNum/tp.NPUTaskNum != reqResource.aicoreQuota {
		return fmt.Errorf("%s check share device job(%s) valid failed, aicoreQuota: %v not equal to "+
			"tp.ReqNPUNum/tp.NPUTaskNum: %v", tp.GetPluginName(), tp.Name, reqResource.aicoreQuota,
			tp.ReqNPUNum/tp.NPUTaskNum)
	}
	if reqResource.hbmQuota < util.MinHbmQuota {
		return fmt.Errorf("%s check share device job(%s) valid failed, hbmQuota: %v less than 1",
			tp.GetPluginName(), tp.Name, reqResource.hbmQuota)
	}
	if reqResource.schedulingPolicy != util.SoftShareDevPolicyFixedShare &&
		reqResource.schedulingPolicy != util.SoftShareDevPolicyElastic &&
		reqResource.schedulingPolicy != util.SoftShareDevPolicyBestEffort {
		return fmt.Errorf("%s check share device job(%s) valid failed, schedulingPolicy: %v is invalid",
			tp.GetPluginName(), tp.Name, reqResource.schedulingPolicy)
	}
	return nil
}

func (tp *chip1softsharedev) validNPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = util.InvalidResourceRequestReason
			vResult.Message = vErr.Error()
		}
	}()
	// check parameter
	if tp == nil {
		vErr = fmt.Errorf("nil plugin")
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}
	reqResource, vErr := tp.getSoftShareDevResource()
	if vErr != nil {
		klog.V(util.LogErrorLev).Infof("getSoftShareDevResource err: %s.", vErr)
		return vResult
	}
	if vErr = tp.checkSoftShareDevResource(reqResource); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkSoftShareDevResource err: %s.", vErr)
		return vResult
	}
	return nil
}
