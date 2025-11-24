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

// Package domain fault config function
package domain

import (
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

type faultCodeCfgCache struct {
	NotHandleFaultCodes  map[int64]struct{}
	RestartRequestCodes  map[int64]struct{}
	RestartBusinessCodes map[int64]struct{}
	RestartNPUCodes      map[int64]struct{}
	FreeRestartNPUCodes  map[int64]struct{}
	SeparateNPUCodes     map[int64]struct{}
}

// FaultCodeFromFile fault codes from file
type FaultCodeFromFile struct {
	NotHandleFaultCodes  []string
	RestartRequestCodes  []string
	RestartBusinessCodes []string
	RestartNPUCodes      []string
	FreeRestartNPUCodes  []string
	SeparateNPUCodes     []string
}

var faultCodeCfg faultCodeCfgCache

func init() {
	faultCodeCfg = faultCodeCfgCache{
		NotHandleFaultCodes:  make(map[int64]struct{}),
		RestartRequestCodes:  make(map[int64]struct{}),
		RestartBusinessCodes: make(map[int64]struct{}),
		RestartNPUCodes:      make(map[int64]struct{}),
		FreeRestartNPUCodes:  make(map[int64]struct{}),
		SeparateNPUCodes:     make(map[int64]struct{}),
	}
}

// SaveFaultCodesToCache save fault codes to cache
func SaveFaultCodesToCache(faultCodes FaultCodeFromFile) {
	faultCodeCfg = faultCodeCfgCache{
		NotHandleFaultCodes:  utils.StringTool.HexStringToInt(utils.RemoveDuplicates(faultCodes.NotHandleFaultCodes)),
		RestartRequestCodes:  utils.StringTool.HexStringToInt(utils.RemoveDuplicates(faultCodes.RestartRequestCodes)),
		RestartBusinessCodes: utils.StringTool.HexStringToInt(utils.RemoveDuplicates(faultCodes.RestartBusinessCodes)),
		RestartNPUCodes:      utils.StringTool.HexStringToInt(utils.RemoveDuplicates(faultCodes.RestartNPUCodes)),
		FreeRestartNPUCodes:  utils.StringTool.HexStringToInt(utils.RemoveDuplicates(faultCodes.FreeRestartNPUCodes)),
		SeparateNPUCodes:     utils.StringTool.HexStringToInt(utils.RemoveDuplicates(faultCodes.SeparateNPUCodes)),
	}
}

// GetFaultLevelByCode get fault level by fault code
func GetFaultLevelByCode(faultCodes []int64) string {
	if len(faultCodes) == 0 {
		return common.NormalNPU
	}
	switch {
	case utils.SameElementInMap(faultCodeCfg.SeparateNPUCodes, faultCodes):
		return common.SeparateNPU
	case utils.SameElementInMap(faultCodeCfg.RestartNPUCodes, faultCodes):
		return common.RestartNPU
	case utils.SameElementInMap(faultCodeCfg.FreeRestartNPUCodes, faultCodes):
		return common.FreeRestartNPU
	case utils.SameElementInMap(faultCodeCfg.RestartBusinessCodes, faultCodes):
		return common.RestartBusiness
	case utils.SameElementInMap(faultCodeCfg.RestartRequestCodes, faultCodes):
		return common.RestartRequest
	case utils.SameElementInMap(faultCodeCfg.NotHandleFaultCodes, faultCodes):
		return common.NotHandleFault
	default:
		return common.UnknownLevel
	}
}
