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
	"nodeD/pkg/common"
	"nodeD/pkg/device"
	"nodeD/pkg/grpcclient/pubfault"
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
		faults = append(faults,
			constructDpcError(dpcProcessError, common.DpcProcessErrorId, common.DpcProcessFaultCode))
	}
	newDpcMemoryStatus := getNewDpcMemoryStatus(newDpcStatusMap)
	if newDpcMemoryStatus != dpcMemoryError || isFirst {
		dpcMemoryError = newDpcMemoryStatus
		faults = append(faults, constructDpcError(dpcMemoryError, common.DpcMemoryErrorId, common.DpcMemoryFaultCode))
	}
	isFirst = false
	if len(faults) == 0 {
		return faultDevInfo
	}
	publicFault := &pubfault.PublicFaultRequest{
		Version:   common.PublicFaultVersion,
		Id:        string(uuid.NewUUID()),
		Timestamp: time.Now().UnixMilli(),
		Resource:  common.DpcFaultResource,
	}
	publicFault.Faults = faults
	faultDevInfo.PubFaultInfo = append(faultDevInfo.PubFaultInfo, publicFault)
	return faultDevInfo
}

func getNewDpcMemoryStatus(newStatusMap map[int]common.DpcStatus) bool {
	// if old status is true(error), should all new status is false and keep ten minutes, then return false(healthy)
	if dpcMemoryError {
		for _, newStatus := range newStatusMap {
			if newStatus.MemoryError || newStatus.MemoryErrorTime == 0 ||
				time.Since(time.UnixMilli(newStatus.MemoryErrorTime)) < common.MemoryErrorTimeOut {
				return true
			}
		}
		return false
	} else {
		// old status is false(healthy), should any new status is true and keep ten minutes, then return true(error)
		for _, newStatus := range newStatusMap {
			if newStatus.MemoryError && newStatus.MemoryErrorTime != 0 &&
				time.Since(time.UnixMilli(newStatus.MemoryErrorTime)) >= common.MemoryErrorTimeOut {
				return true
			}
		}
		return false
	}
}

func constructDpcError(errorStatus bool, id string, faultCode string) *pubfault.Fault {
	assertion := common.FaultAssertionRecover
	if errorStatus {
		assertion = common.FaultAssertionOccur
	}
	nodeName := os.Getenv(api.NodeNameEnv)
	return &pubfault.Fault{
		Assertion:     assertion,
		FaultId:       common.GenerateFaultID(nodeName, id),
		FaultType:     common.FaultType,
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
	dm := device.GetDeviceManager()
	cardNum, cardList, err := dm.GetCardList()
	if err != nil {
		hwlog.RunLog.Errorf("get card list error %v", err)
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
