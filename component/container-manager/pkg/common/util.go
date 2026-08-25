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

// Package common a series common function
package common

import (
	"bytes"
	"encoding/gob"
	"os"
	"os/signal"

	"ascend-common/api"
	"ascend-common/common-utils/utils"
)

// NewSignWatcher new sign watcher
func NewSignWatcher(osSigns ...os.Signal) chan os.Signal {
	// create signs chan
	signChan := make(chan os.Signal, 1)
	for _, sign := range osSigns {
		signal.Notify(signChan, sign)
	}
	return signChan
}

// GetDevNumPerRing get reset device num at a time.
// 910 and 910A2 device reset by ring, 910A3 reset related devices, 910A5 reset single device.
func GetDevNumPerRing(devType string, deviceNum int, boardId uint32) int {
	switch devType {
	case api.Ascend910A:
		return Ascend910RingsNum
	case api.Ascend910B:
		if boardId == A300IA2BoardId || boardId == A300IA2GB64BoardId ||
			boardId == A800IA2NoneHccsBoardId || boardId == A800IA2NoneHccsBoardIdOld {
			return NoRingNum
		}
		if deviceNum > Ascend910BRingsNumTrain {
			return A200TA2RingsNum
		}
		return Ascend910BRingsNumTrain
	case api.Ascend910A3:
		return Ascend910A3RingsNum
	case api.Ascend910A5:
		return Ascend910A5RingsNum
	default:
		// use initial value, do nothing
	}
	return NoRingNum
}

// GetNeedPauseCtrFaultLevels need pause ctr fault levels
func GetNeedPauseCtrFaultLevels() []string {
	return []string{
		RestartRequest,
		RestartBusiness,
		FreeRestartNPU,
		RestartNPU,
	}
}

// GetDevStatus get dev status
func GetDevStatus(faults []*DevFaultInfo) string {
	for _, fault := range faults {
		if utils.Contains(GetNeedPauseCtrFaultLevels(), fault.FaultLevel) {
			return StatusNeedPause
		}
	}
	return StatusIgnorePause
}

// DeepCopy for object using gob
// DeepCopy has performance problem, cannot use in Time-sensitive scenario
func DeepCopy(dst, src interface{}) error {
	if src == nil {
		return nil
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
