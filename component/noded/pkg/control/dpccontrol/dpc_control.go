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

// Package dpccontrol for dpc fault handling
package dpccontrol

import (
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"nodeD/pkg/common"
	"nodeD/pkg/grpcclient/pubfault"
)

const (
	publicFaultVersion    = "1.0"
	faultResource         = "dpcStorage"
	faultAssertionRecover = "recover"
	faultAssertionOccur   = "occur"
	faultType             = "Node"

	dpcMemoryErrorId  = "DPC_MEMORY"
	dpcProcessErrorId = "DPC_PROCESS"

	processFaultCode = "110001020"
	memoryFaultCode  = "110001021"

	memoryErrorTimeOut = 10 * 60 * time.Second
)

var (
	dpcProcessError = false
	dpcMemoryError  = false
	isFirst         = true
)

// Controller control dpc fault on server
type Controller struct {
}

// NewDpcController create a dpc controller
func NewDpcController() *Controller {
	return &Controller{}
}

// Name get dpc control name
func (dc *Controller) Name() string {
	return common.PluginControlFault
}

// Control update fault dpc info
func (dc *Controller) Control(faultDevInfo *common.FaultAndConfigInfo) *common.FaultAndConfigInfo {
	newDpcStatusMap := faultDevInfo.DpcStatusMap
	if newDpcStatusMap == nil {
		return faultDevInfo
	}
	var faults []*pubfault.Fault
	newDpcProcessStatus := getNewDpcProcessStatus(newDpcStatusMap)
	if newDpcProcessStatus != dpcProcessError || isFirst {
		dpcProcessError = newDpcProcessStatus
		faults = append(faults, constructDpcError(dpcProcessError, dpcProcessErrorId, processFaultCode))
	}
	newDpcMemoryStatus := getNewDpcMemoryStatus(newDpcStatusMap)
	if newDpcMemoryStatus != dpcMemoryError || isFirst {
		dpcMemoryError = newDpcMemoryStatus
		faults = append(faults, constructDpcError(dpcMemoryError, dpcMemoryErrorId, memoryFaultCode))
	}
	isFirst = false
	if len(faults) == 0 {
		return faultDevInfo
	}
	if faultDevInfo.PubFaultInfo == nil {
		faultDevInfo.PubFaultInfo = &pubfault.PublicFaultRequest{
			Version:   publicFaultVersion,
			Id:        string(uuid.NewUUID()),
			Timestamp: time.Now().UnixMilli(),
			Resource:  faultResource,
		}
	}
	faultDevInfo.PubFaultInfo.Faults = faults
	return faultDevInfo
}

func getNewDpcMemoryStatus(newStatusMap map[int]common.DpcStatus) bool {
	// if old status is true(error), should all new status is false and keep ten minutes, then return false(healthy)
	if dpcMemoryError {
		for _, newStatus := range newStatusMap {
			if newStatus.MemoryError || newStatus.MemoryErrorTime == 0 ||
				time.Since(time.UnixMilli(newStatus.MemoryErrorTime)) < memoryErrorTimeOut {
				return true
			}
		}
		return false
	} else {
		// old status is false(healthy), should any new status is true and keep ten minutes, then return true(error)
		for _, newStatus := range newStatusMap {
			if newStatus.MemoryError && newStatus.MemoryErrorTime != 0 &&
				time.Since(time.UnixMilli(newStatus.MemoryErrorTime)) >= memoryErrorTimeOut {
				return true
			}
		}
		return false
	}
}

func constructDpcError(errorStatus bool, id string, faultCode string) *pubfault.Fault {
	assertion := faultAssertionRecover
	if errorStatus {
		assertion = faultAssertionOccur
	}
	nodeName := os.Getenv(api.NodeNameEnv)
	return &pubfault.Fault{
		Assertion:     assertion,
		FaultId:       common.GenerateFaultID(nodeName, id),
		FaultType:     faultType,
		FaultCode:     faultCode,
		FaultTime:     time.Now().UnixMilli(),
		FaultLocation: map[string]string{},
		Influence: []*pubfault.PubFaultInfo{
			{
				NodeName:  nodeName,
				DeviceIds: getNodeCardId(),
			},
		},
	}
}

func getNodeCardId() []int32 {
	dm, err := devmanager.GetDeviceManager()
	if err != nil {
		hwlog.RunLog.Errorf("get dev manager failed, err is %v", err)
		return []int32{int32(0)}
	}
	cardNum, cardList, err := dm.GetCardList()
	if err != nil {
		hwlog.RunLog.Error(err)
		return []int32{int32(0)}
	}
	if cardNum == 0 {
		hwlog.RunLog.Errorf("get chip info failed, no card found")
		return []int32{int32(0)}
	}
	return cardList
}

func getNewDpcProcessStatus(newStatusMap map[int]common.DpcStatus) bool {
	newProcessError := false
	for _, status := range newStatusMap {
		if status.ProcessError {
			newProcessError = true
			break
		}
	}
	return newProcessError
}
