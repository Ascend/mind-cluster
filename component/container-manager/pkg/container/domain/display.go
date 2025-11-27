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

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	defaultFileSize             = 100
	defaultFilePerm os.FileMode = 0640
)

func (cc *CtrCache) updateStatusFile() {
	var contexts []common.CtrStatusInfo
	for id, info := range cc.ctrInfoMap {
		contexts = append(contexts, common.CtrStatusInfo{
			CtrId:           id,
			Status:          info.Status,
			StatusStartTime: info.StatusStartTime,
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
