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

// Package app fault manager struct
package app

import (
	"context"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"ascend-common/common-utils/hwlog"
	ascommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/domain"
)

const (
	processFaultDuration = 500 * time.Millisecond
	faultLimit           = 10000
	checkTimeout         = 300
)

var limiter = rate.NewLimiter(rate.Every(1*time.Minute/faultLimit), faultLimit)

// ProcessDCMIFault process fault from dcmi interface
func (fm *FaultMgr) ProcessDCMIFault(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-fm.faultInfo.UpdateChan:
			if !ok {
				hwlog.RunLog.Info("catch update signal channel closed")
				return
			}
			hwlog.RunLog.Infof("receive reset device success signal, check the fault cache")
			fm.doCheck()
		default:
			if QueueCache.Len() == 0 {
				time.Sleep(processFaultDuration)
				continue
			}
			needDeal := QueueCache.Pop()
			fm.faultInfo.AddFault(needDeal)
			domain.SharedFaultCache.AddFault(&needDeal)
		}
	}
}

func (fm *FaultMgr) checkMoreThanFiveMinFaults(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fm.doCheck()
		}
	}
}

func (fm *FaultMgr) doCheck() {
	faultInfos, err := fm.faultInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault info failed, error: %v", err)
		return
	}
	var idsNeedGetCodes []int32
	for id, codeLayer := range faultInfos {
		if fm.needGetAllCodes(codeLayer) {
			idsNeedGetCodes = append(idsNeedGetCodes, id)
		}
	}
	for _, id := range idsNeedGetCodes {
		_, codes, err := devmgr.DevMgr.GetDeviceErrCode(id)
		if err != nil {
			hwlog.RunLog.Errorf("get device %d error code failed, error: %v", id, err)
			return
		}
		// only consider the situation of losing recover messages
		fm.faultInfo.UpdateFaultsOnDev(id, codes)
	}
}

// needGetAllCodes as long as there is a fault code timeout on a card, it will return true directly,
// and the fault code of the card will be queried uniformly below
func (fm *FaultMgr) needGetAllCodes(codeLayer map[int64]map[string]*common.DevFaultInfo) bool {
	for _, moduleLayer := range codeLayer {
		for _, faultInfo := range moduleLayer {
			receiveTime := faultInfo.ReceiveTime
			if time.Now().Unix()-receiveTime <= checkTimeout {
				continue
			}
			return true
		}
	}
	return false
}

func (fm *FaultMgr) getAllFaultInfo() {
	idCodesMap := devmgr.DevMgr.GetFaultCodesMap()
	for id, codes := range idCodesMap {
		for _, code := range codes {
			QueueCache.Push(*domain.ConstructMockModuleFault(id, code))
		}
	}
	return
}

func saveDevFaultInfo(devFaultInfo ascommon.DevFaultInfo) {
	if !limiter.Allow() {
		hwlog.RunLog.Errorf("fault exceeds the upper limit from subscribe interface, fault [%+v] will be discard", devFaultInfo)
		return
	}
	hwlog.RunLog.Infof("receive devFaultInfo: %+v, hex fault code: %v", devFaultInfo,
		strconv.FormatInt(devFaultInfo.EventID, common.Hex))
	if devFaultInfo.EventID == 0 {
		return
	}
	QueueCache.Push(convertDevFaultInfo(devFaultInfo))
}

func convertDevFaultInfo(devFaultInfo ascommon.DevFaultInfo) common.DevFaultInfo {
	return common.DevFaultInfo{
		EventID:       devFaultInfo.EventID,
		LogicID:       devFaultInfo.LogicID,
		ModuleType:    devFaultInfo.ModuleType,
		ModuleID:      devFaultInfo.ModuleID,
		SubModuleType: devFaultInfo.SubModuleType,
		SubModuleID:   devFaultInfo.SubModuleID,
		Assertion:     devFaultInfo.Assertion,
		PhyID:         devmgr.DevMgr.GetPhyIdByLogicId(devFaultInfo.LogicID),
		FaultLevel:    domain.GetFaultLevelByCode([]int64{devFaultInfo.EventID}),
		ReceiveTime:   time.Now().Unix(),
	}
}
