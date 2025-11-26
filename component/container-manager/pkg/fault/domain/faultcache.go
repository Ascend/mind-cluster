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

// Package domain fault cache function
package domain

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const mockFaultAttr = -1

var faultCache *FaultCache
var initOnce sync.Once

// FaultCache fault events from dcmi interface
type FaultCache struct {
	// key: phy id, value: {fault code : {module type + module id + submodule type + submodule id : fault info}}
	faults map[int32]map[int64]map[string]*common.DevFaultInfo
	// UpdateChan for changed fault events
	UpdateChan chan struct{}
	mutex      sync.Mutex
}

// GetFaultCache new fault cache
func GetFaultCache() *FaultCache {
	initOnce.Do(
		func() {
			faultCache = &FaultCache{
				faults:     make(map[int32]map[int64]map[string]*common.DevFaultInfo),
				UpdateChan: make(chan struct{}, 1),
				mutex:      sync.Mutex{},
			}
		},
	)
	return faultCache
}

// AddFault add fault to faults in cache
func (fc *FaultCache) AddFault(newFault common.DevFaultInfo) {
	moduleLayerKey := fmt.Sprintf("%d%d%d%d", newFault.ModuleType, newFault.ModuleID,
		newFault.SubModuleType, newFault.SubModuleID)
	switch newFault.Assertion {
	case common.FaultOccur:
		fc.dealFaultOccur(newFault, moduleLayerKey)
	case common.FaultRecover:
		fc.dealFaultRecover(newFault, moduleLayerKey)
	default:
		// ignore once fault event
		return
	}
}

func (fc *FaultCache) dealFaultOccur(newFault common.DevFaultInfo, moduleLayerKey string) {
	fc.mutex.Lock()
	defer func() {
		fc.mutex.Unlock()
		fc.printFaults()
	}()
	codeLayer, ok := fc.faults[newFault.PhyID]
	if !ok {
		fc.faults[newFault.PhyID] = map[int64]map[string]*common.DevFaultInfo{
			newFault.EventID: {moduleLayerKey: &newFault},
		}
		return
	}
	moduleLayer, ok := codeLayer[newFault.EventID]
	if !ok {
		codeLayer[newFault.EventID] = map[string]*common.DevFaultInfo{moduleLayerKey: &newFault}
		return
	}
	_, ok = moduleLayer[moduleLayerKey]
	if !ok {
		moduleLayer[moduleLayerKey] = &newFault
		return
	}
}

func (fc *FaultCache) dealFaultRecover(newFault common.DevFaultInfo, moduleLayerKey string) {
	fc.mutex.Lock()
	defer func() {
		fc.mutex.Unlock()
		fc.printFaults()
	}()
	codeLayer, ok := fc.faults[newFault.PhyID]
	if !ok {
		return
	}
	moduleLayer, ok := codeLayer[newFault.EventID]
	if !ok {
		return
	}
	_, ok = moduleLayer[moduleLayerKey]
	if !ok {
		return
	}
	delete(moduleLayer, moduleLayerKey)
	if len(moduleLayer) == 0 {
		delete(codeLayer, newFault.EventID)
		if len(codeLayer) == 0 {
			delete(fc.faults, newFault.PhyID)
		}
	}
}

func (fc *FaultCache) printFaults() {
	if len(fc.faults) == 0 {
		hwlog.RunLog.Debug("no faults")
		return
	}
	hwlog.RunLog.Debug("begin record faults")
	for id, codeLayer := range fc.faults {
		for code, moduleLayer := range codeLayer {
			for moduleLayerKey, fault := range moduleLayer {
				hwlog.RunLog.Debugf("id: %d, code: %s, module: %s, fault: %+v",
					id, strconv.FormatInt(code, common.Hex), moduleLayerKey, fault)
			}
		}
	}
	hwlog.RunLog.Debug("record faults end")
}

// Notify fault events updated, notify reset manager to process
func (fc *FaultCache) Notify() {
	if len(fc.UpdateChan) == 0 {
		fc.UpdateChan <- struct{}{}
	}
}

// DeepCopy deep copy faults
func (fc *FaultCache) DeepCopy() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	result := new(map[int32]map[int64]map[string]*common.DevFaultInfo)
	if err := common.DeepCopy(result, fc.faults); err != nil {
		return nil, err
	}
	return *result, nil
}

// UpdateFaultsOnDev update faults on dev
func (fc *FaultCache) UpdateFaultsOnDev(id int32, faultCodes []int64) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	if len(faultCodes) == 0 {
		fc.faults[id] = make(map[int64]map[string]*common.DevFaultInfo)
		return
	}

	// if the recover message is lost, the data in the cache needs to be deleted
	// directly deleting may result in out of range
	var newCodeLayer = make(map[int64]map[string]*common.DevFaultInfo)
	for code, codeLayer := range fc.faults[id] {
		if utils.Contains(faultCodes, code) {
			newCodeLayer[code] = codeLayer
		}
	}
	fc.faults[id] = newCodeLayer
}

// ConstructMockModuleFault construct mock module fault
// module type / module id / submodule type / submodule id mock value is 0
func ConstructMockModuleFault(phyId int32, faultCode int64) *common.DevFaultInfo {
	return &common.DevFaultInfo{
		EventID:       faultCode,
		LogicID:       mockFaultAttr,
		ModuleType:    mockFaultAttr,
		ModuleID:      mockFaultAttr,
		SubModuleType: mockFaultAttr,
		SubModuleID:   mockFaultAttr,
		Assertion:     common.FaultOccur,
		PhyID:         phyId,
		FaultLevel:    GetFaultLevelByCode([]int64{faultCode}),
		ReceiveTime:   time.Now().Unix(),
	}
}
