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

// Package app fault config function
package app

import (
	"encoding/json"
	"fmt"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	"container-manager/pkg/fault/domain"
)

func loadFaultCodeFromFile() error {
	filePath := common.ParamOption.FaultCfgPath
	if filePath == "" {
		return nil
	}
	faultCodeBytes, err := utils.LoadFile(filePath)
	if err != nil {
		return err
	}
	var faultCodes domain.FaultCodeFromFile
	if err = json.Unmarshal(faultCodeBytes, &faultCodes); err != nil {
		return fmt.Errorf("unmarshal custom fault code byte failed: %v", err)
	}
	domain.SaveFaultCodesToCache(faultCodes)
	hwlog.RunLog.Infof("load custom fault config file from %s success", filePath)
	return nil
}
