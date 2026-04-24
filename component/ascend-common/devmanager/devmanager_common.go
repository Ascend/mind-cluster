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

package devmanager

import (
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
	"errors"
	"fmt"
	"time"
)

type DeviceCommonSetInterface interface {
	DeviceInterface
	SetValidMainBoardInfo() error
	SetDcManger(dcMgr interface{}) error
	SetDevType(devType string)
	GetDcManager() DeviceInterface
	SetAllProductType() error
	GetDcmiApiVersion() string
	SetDcmiVersion()
}

const (
	DcmiApiV1 = "dcmi"
	// DcmiApiV2 for the dcmiv2_xxx api
	DcmiApiV2 = "dcmiv2"
)

var deviceCommonSetManagerList = []DeviceCommonSetInterface{
	&deviceCommonInitManager{
		DeviceManager: DeviceManager{
			DcMgr:          &dcmi.DcManager{},
			dcmiApiVersion: DcmiApiV1,
		},
	},
	&deviceCommonInitManagerV2{
		DeviceManagerV2: DeviceManagerV2{
			DcMgr:          &dcmi.DcV2Manager{},
			dcmiApiVersion: DcmiApiV2,
		},
	},
}

// DetectDcmiApiVersion for detect dcmi dynamic library interface api version, such as dcmi_xxx or dcmiv2_xxx,
// and return a common device set manager to set all common param within resetTimeout
func DetectDcmiApiVersion(resetTimeout int) (DeviceCommonSetInterface, error) {
	for start, retryCnt := 0, 1; start < resetTimeout; retryCnt, start = retryCnt+1, start+defaultRetryDelay {
		hwlog.RunLog.Infof("timeout is %ds, dcmi version detection at %d times: ", resetTimeout, retryCnt)
		for _, devCommonSetMgr := range deviceCommonSetManagerList {
			hwlog.RunLog.Infof("try dcmi api version: %v", devCommonSetMgr.GetDcmiApiVersion())
			if err := devCommonSetMgr.Init(); err == nil {
				if err := devCommonSetMgr.ShutDown(); err != nil {
					hwlog.RunLog.Warnf("dcmi shutdown failed, err: %v", err)
					// ignore error
				}
				hwlog.RunLog.Infof("dcmi api version is %v", devCommonSetMgr.GetDcmiApiVersion())
				return devCommonSetMgr, nil
			} else {
				hwlog.RunLog.Warnf("dcmi api version: %v, init err: %v", devCommonSetMgr.GetDcmiApiVersion(), err)
			}
		}
		time.Sleep(defaultRetryDelay)
	}
	return nil, errors.New(fmt.Sprintf("after %ds, can not find an available dcmi version", resetTimeout))
}

func getDeviceInfoForInit(commonDevMgr DeviceInterface) (common.ChipInfo, common.BoardInfo, error) {
	var err error
	chipInfo, err := commonDevMgr.GetValidChipInfo()
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.ChipInfo{}, common.BoardInfo{}, err
	}
	boardInfo, err := commonDevMgr.GetValidBoardInfo()
	if err != nil {
		hwlog.RunLog.Error(err)
		return chipInfo, common.BoardInfo{}, err
	}

	return chipInfo, boardInfo, nil
}

// AutoInit auto detect npu chip type and return the corresponding processing object
func AutoInit(dType string, resetTimeout int) (DeviceInterface, error) {
	var devMgr DeviceInterface
	devCommonSetMgr, err := DetectDcmiApiVersion(resetTimeout)
	if err != nil {
		return nil, fmt.Errorf("detect dcmi version failed, err: %s", err)
	}
	// reduce interface range
	devMgr = devCommonSetMgr.GetDcManager()
	devMgr.WaitDeviceOnline(resetTimeout)
	chipInfo, boardInfo, err := getDeviceInfoForInit(devCommonSetMgr)
	if err != nil {
		return nil, fmt.Errorf("auto init failed when get device info, err: %s", err)
	}
	devCommonSetMgr.SetDcmiVersion()
	hwlog.RunLog.Infof("the dcmi version is %s", devMgr.GetDcmiVersion())
	err = devCommonSetMgr.SetValidMainBoardInfo()
	if err != nil {
		// Non-blocking when the main board ID is not found
		hwlog.RunLog.Warn(err)
	}
	var devType = common.GetDevType(chipInfo.Name, boardInfo.BoardId)
	switch devType {
	case api.Ascend910A5:
		err = devCommonSetMgr.SetDcManger(&A950Manager{})
	case api.Ascend910A, api.Ascend910B, api.Ascend910A3:
		err = devCommonSetMgr.SetDcManger(&A910Manager{})
	case api.Ascend310P:
		err = devCommonSetMgr.SetDcManger(&A310PManager{})
	case api.Ascend310, api.Ascend310B:
		err = devCommonSetMgr.SetDcManger(&A310Manager{})
	default:
		return nil, fmt.Errorf("unsupported device type (%s)", devType)
	}
	if err != nil {
		return nil, fmt.Errorf("set dcManager failed, err: %s", err)
	}
	if devType == api.Ascend910A5 {
		hwlog.RunLog.Infof("chipName: %v, devType: npu", chipInfo.Name)
	} else {
		hwlog.RunLog.Infof("chipName: %v, devType: %v", chipInfo.Name, devType)
	}
	if dType != "" && devType != dType {
		return nil, fmt.Errorf("the value of dType(%s) is inconsistent with the actual chip type(%s)",
			dType, devType)
	}
	devCommonSetMgr.SetDevType(devType)
	if err := devCommonSetMgr.SetIsTrainingCard(); err != nil {
		hwlog.RunLog.Errorf("auto recognize training card failed, err: %s", err)
	}
	err = devCommonSetMgr.SetAllProductType()
	if err != nil {
		hwlog.RunLog.Debugf("auto init product types failed, err: %s", err)
	}
	return devMgr, nil
}
