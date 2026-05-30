/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package dtfscontrol for dtfs fault handling
package dtfscontrol

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
	dtfsProcessError = false
	dtfsLinkError    = false
	isFirst          = true
)

// Controller control dtfs fault on server
type Controller struct {
}

// NewDtfsController create a dtfs controller
func NewDtfsController() *Controller {
	return &Controller{}
}

// Name get dtfs control name
func (dc *Controller) Name() string {
	return common.PluginControlDtfs
}

// Control update fault dtfs info
func (dc *Controller) Control(faultDevInfo *common.FaultAndConfigInfo) *common.FaultAndConfigInfo {
	newDtfsStatus := faultDevInfo.DtfsStatus
	var faults []*pubfault.Fault

	if newDtfsStatus.ProcessError != dtfsProcessError || isFirst {
		dtfsProcessError = newDtfsStatus.ProcessError
		faults = append(faults,
			constructDtfsError(dtfsProcessError, common.DtfsProcessErrorId, common.DtfsProcessFaultCode))
	}

	if newDtfsStatus.LinkError != dtfsLinkError || isFirst {
		dtfsLinkError = newDtfsStatus.LinkError
		faults = append(faults, constructDtfsError(dtfsLinkError, common.DtfsLinkErrorId, common.DtfsLinkFaultCode))
	}

	isFirst = false
	if len(faults) == 0 {
		return faultDevInfo
	}

	publicFault := &pubfault.PublicFaultRequest{
		Version:   common.PublicFaultVersion,
		Id:        string(uuid.NewUUID()),
		Timestamp: time.Now().UnixMilli(),
		Resource:  common.DtfsFaultResource,
	}
	publicFault.Faults = faults
	faultDevInfo.PubFaultInfo = append(faultDevInfo.PubFaultInfo, publicFault)
	return faultDevInfo
}

func constructDtfsError(errorStatus bool, id string, faultCode string) *pubfault.Fault {
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
		hwlog.RunLog.Error("get chip info failed, no card found")
		return []int32{int32(0)}
	}
	return cardList
}
