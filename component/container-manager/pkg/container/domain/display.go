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

// Package domain function for displaying container status
package domain

import (
	"encoding/json"
	"os"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	pausingDuration  = 20
	pausedDuration   = 400
	resumingDuration = 20

	defaultFileSize             = 100
	defaultFilePerm os.FileMode = 0640
)

var ctrStatusInfos = map[string]struct {
	statusDuration int64
	description    string
}{
	common.StatusPausing: {
		pausingDuration,
		"Container pause may fail. Please manually delete the container",
	},
	common.StatusPaused: {
		pausedDuration,
		"Device hot reset may fail. Please check of device status and recovery are required",
	},
	common.StatusResuming: {
		resumingDuration,
		"The device has been recovered, but the container failed to be resumed. Please manually pull up the container",
	},
}

func (cc *CtrCache) updateStatusFile() {
	var contexts []common.CtrStatusInfo
	for id, info := range cc.ctrInfoMap {
		contexts = append(contexts, common.CtrStatusInfo{
			CtrId:           id,
			Status:          info.Status,
			StatusStartTime: info.StatusStartTime,
			Description:     cc.getDesc(info.Status, info.StatusStartTime),
		})
	}
	bytes, err := json.Marshal(contexts)
	if err != nil {
		hwlog.RunLog.Errorf("marshal status info failed, error: %v", err)
		return
	}
	if utils.IsExist(common.StatusInfoFile) {
		_, err = utils.RealFileChecker(common.StatusInfoFile, false, false, defaultFileSize)
		if err != nil {
			hwlog.RunLog.Errorf("check file %s failed, error: %v", common.StatusInfoFile, err)
			return
		}
	}
	if err = os.WriteFile(common.StatusInfoFile, bytes, defaultFilePerm); err != nil {
		hwlog.RunLog.Errorf("write status info to file failed, error: %v", err)
		return
	}
}

func (cc *CtrCache) getDesc(status string, startTime int64) string {
	if status == common.StatusRunning {
		return common.DescNormal
	}
	infos, ok := ctrStatusInfos[status]
	if !ok {
		return common.DescUnknown
	}
	if time.Now().Unix()-startTime > infos.statusDuration {
		return infos.description
	}
	return common.DescNormal
}
