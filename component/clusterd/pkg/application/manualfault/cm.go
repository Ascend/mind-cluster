/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manualfault process manual separate npu info
package manualfault

import (
	"context"
	"time"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
)

// ProcessManuSep process manually separate npu info
func ProcessManuSep(ctx context.Context) {
	const updateCmInterval = 15 * time.Second
	ticker := time.NewTicker(updateCmInterval)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Infof("received stop signal: %v", ctx.Err())
			return
		case <-ticker.C:
			if manualfault.FaultCmInfo.Len() == 0 {
				manualfault.DeleteManualCm()
				continue
			}
			checkDiffAndDelete()
			release()
			manualfault.UpdateOrCreateManualCm()
		}
	}
}

// checkDiffAndDelete check diff from cache and cm. delete the dev deleted from cm form the cache synchronously
func checkDiffAndDelete() {
	// get manual info cm failed, don't check diff
	cm, err := manualfault.TryGetManualCm()
	if err != nil {
		return
	}
	manualDeleted := getManualDeletedDev(cm)
	for nodeName, info := range manualDeleted {
		for _, devId := range info {
			hwlog.RunLog.Errorf("node: %s, dev: %s is manually delete from cm, so delete from cache synchronously",
				nodeName, devId)
			manualfault.FaultCmInfo.DeleteSeparateDev(nodeName, devId)
		}
	}
}

// getManualDeletedDev delete manually separate npu in cm, data resource is cache
func getManualDeletedDev(cm *v1.ConfigMap) map[string][]string {
	lastSep := manualfault.GetSepNPUByLastCmInfo()
	currentSep := manualfault.GetSepNPUByCurrentCmInfo(cm)
	return utils.GetItemInANotInB(lastSep, currentSep)
}

func release() {
	if !conf.IsReleaseEnable() {
		return
	}

	nodeInfo, err := manualfault.FaultCmInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault cm info failed, error: %v", err)
		return
	}
	for nodeName, info := range nodeInfo {
		for dev, devInfo := range info.Detail {
			doRelease(nodeName, dev, devInfo)
		}
	}
}

func doRelease(nodeName, dev string, devInfo []manualfault.DevCmInfo) {
	for _, cmInfo := range devInfo {
		if time.Now().UnixMilli()-cmInfo.LastSeparateTime >= conf.GetReleaseDuration() {
			hwlog.RunLog.Infof("node: %s, dev: %s, code: %s has been reached release time, released it",
				nodeName, dev, cmInfo.FaultCode)
			manualfault.Counter.ClearDevFault(nodeName, dev, cmInfo.FaultCode)
			manualfault.FaultCmInfo.DeleteDevCode(nodeName, dev, cmInfo.FaultCode)
			continue
		}
	}
}

// LoadManualCmInfo load manually separate npu info from configmap
func LoadManualCmInfo() {
	cm, err := manualfault.TryGetManualCm()
	if err != nil {
		hwlog.RunLog.Errorf("load cm <%s/%s> err: %v", api.ClusterNS, constant.ConfigCmName, err)
		return
	}
	if cm == nil {
		hwlog.RunLog.Infof("manually separate npu cm <%s/%s> is not found", api.ClusterNS, constant.ConfigCmName)
		return
	}
	cmInfo, err := manualfault.ParseManualCm(cm)
	if err != nil {
		hwlog.RunLog.Errorf("parse separate npu cm failed, error: %v", err)
		return
	}
	manualfault.FaultCmInfo.SetNodeInfo(cmInfo)
	hwlog.RunLog.Info("save manually separate npu info to cache success")
}
