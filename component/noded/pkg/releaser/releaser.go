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

// Package releaser the device fault releaser
package releaser

import (
	"encoding/json"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	devcom "ascend-common/devmanager/common"
	"nodeD/pkg/common"
	"nodeD/pkg/device"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/watcher/configmap"
)

// Releaser the device fault releaser
type Releaser struct {
	devManager *devmanager.DeviceManager
	chips      []*devcom.ChipBaseInfo
}

var (
	faultInfoCm = make(map[string]map[int]api.SuperPodFaultInfos)
	releaser    = &Releaser{}
)

// InitReleaser init the device fault releaser
func InitReleaser() {
	releaser.devManager = device.GetDeviceManager()
	releaser.initWatcher()
	chips, err := releaser.devManager.GetChipBaseInfos()
	if err != nil {
		hwlog.RunLog.Errorf("get chips failed, err: %v", err)
		return
	}
	releaser.chips = chips
}

func (r *Releaser) initWatcher() {
	var opts []configmap.Option
	opts = append(opts, configmap.WithNamedHandlers(
		configmap.NamedHandler{Name: api.FaultJobCmName, Handle: r.handleFaultJobEvent},
	))
	configmap.DoCMWatcherWithOptions(opts...)
}

func (r *Releaser) handleFaultJobEvent(cm *v1.ConfigMap) {
	faultJobs := getFaultJobInfosFromCm(cm)
	needResetJob, needRecoverId := getReleaseAndRecoverMap(faultJobs)
	if len(needResetJob) != 0 {
		r.handleFaultJobRelease(needResetJob)
	}
	if len(needRecoverId) != 0 {
		r.handleNodeStatusChange(needRecoverId, common.NormalStatus)
	}

	faultInfoCm = faultJobs
}

func (r *Releaser) handleFaultJobRelease(needResetJob map[string]api.SuperPodFaultInfos) {
	sdids := make(sets.String)
	for _, jobInfo := range needResetJob {
		for _, sdid := range jobInfo.SdIds {
			sdids.Insert(sdid)
		}
	}
	if len(sdids) != 0 {
		r.handleNodeStatusChange(sdids, common.AbnormalStatus)
	}
}

func (r *Releaser) handleNodeStatusChange(sdids sets.String, nodeStatus int) {
	for sdid := range sdids {
		tmpSdId, err := strconv.ParseUint(sdid, common.Decimal, common.Bit32Size)
		if err != nil {
			hwlog.RunLog.Errorf("parse sdid failed, sdid: %s, err: %v", sdid, err)
			continue
		}
		for _, dev := range r.chips {
			status, err := r.devManager.DcGetSuperPodStatus(dev.CardID, dev.DeviceID, uint32(tmpSdId))
			if err != nil {
				hwlog.RunLog.Errorf("get super pod status failed, err: %v", err)
				continue
			}

			if status == nodeStatus {
				hwlog.RunLog.Warnf("super pod status is normal, skip status <%v> resource cardId:%v,"+
					"devicei:%v,sdid:%v", nodeStatus, dev.CardID, dev.DeviceID, tmpSdId)
				continue
			}
			hwlog.RunLog.Infof("start exec fault job resource status <%v> cardId:%v,devicei:%v,"+
				"sdid:%v", nodeStatus, dev.CardID, dev.DeviceID, tmpSdId)

			if err := r.devManager.DcSetSuperPodStatus(dev.CardID, dev.DeviceID, uint32(tmpSdId),
				uint32(nodeStatus)); err != nil {
				hwlog.RunLog.Errorf("set super pod status failed, err: %v", err)
				continue
			}
			hwlog.RunLog.Infof("set fault job resource status <%v> success device %v", nodeStatus, dev.LogicID)
		}
	}
}

func getFaultJobInfosFromCm(cm *v1.ConfigMap) map[string]map[int]api.SuperPodFaultInfos {
	faultJobs := make(map[string]map[int]api.SuperPodFaultInfos, len(cm.Data))
	for jobId, faultInfo := range cm.Data {
		var tmpFaultJobInfo map[int]api.SuperPodFaultInfos
		err := json.Unmarshal([]byte(faultInfo), &tmpFaultJobInfo)
		if err != nil {
			hwlog.RunLog.Errorf("unmarshal fault time failed, err: %v", err)
			continue
		}
		faultJobs[jobId] = tmpFaultJobInfo
	}
	return faultJobs
}

func getReleaseAndRecoverMap(faultInfos map[string]map[int]api.SuperPodFaultInfos) (
	map[string]api.SuperPodFaultInfos, sets.String) {
	needResetJob := make(map[string]api.SuperPodFaultInfos)
	for jobId, superFaultInfos := range faultInfos {
		for spId, superFaultInfo := range superFaultInfos {
			oldFaultJobInfo := faultInfoCm[jobId][spId]
			if superFaultInfo.FaultTimes == oldFaultJobInfo.FaultTimes {
				continue
			}
			insertNeedResetJobInfo(needResetJob, superFaultInfo)
		}
	}
	needRecoverId := make(sets.String)
	for jobId, oldFJob := range faultInfoCm {
		if _, ok := faultInfos[jobId]; !ok {
			hwlog.RunLog.Infof("job %s need recover", jobId)
			for _, superFaultInfo := range oldFJob {
				insertSdidsByNodeName(needRecoverId, superFaultInfo)
			}
		}
	}
	return needResetJob, needRecoverId
}

func insertNeedResetJobInfo(needResetJob map[string]api.SuperPodFaultInfos, superFaultInfo api.SuperPodFaultInfos) {
	for _, nodeName := range superFaultInfo.NodeNames {
		if nodeName == kubeclient.GetK8sClient().NodeName {
			needResetJob[superFaultInfo.JobId] = superFaultInfo
			hwlog.RunLog.Infof("job %s need reset", superFaultInfo.JobId)
		}
	}
}

func insertSdidsByNodeName(s sets.String, superFaultInfo api.SuperPodFaultInfos) {
	for _, nodeName := range superFaultInfo.NodeNames {
		if nodeName != kubeclient.GetK8sClient().NodeName {
			continue
		}
		for _, sdid := range superFaultInfo.SdIds {
			s.Insert(sdid)
		}
	}
}
