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

// Package devmanager this for device driver manager
package devmanager

import (
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
)

// GetUrmaDeviceCount for A5
func (d *DeviceManager) GetUrmaDeviceCount(logicID int32) (int32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForLogicIdErr, logicID,
			"failed to get cardId and deviceId by logicID(%d), error: %v", logicID, err)
		return common.RetError, err
	}
	return d.DcMgr.DcGetUrmaDeviceCount(cardID, deviceID)
}

// GetUrmaDevEidList for A5
func (d *DeviceManager) GetUrmaDevEidList(logicID int32, index int32) (*common.UrmaDeviceInfo, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForLogicIdErr, logicID,
			"failed to get cardId and deviceId by logicID(%d), error: %v", logicID, err)
		return nil, err
	}
	return d.DcMgr.DcGetUrmaDevEidList(cardID, deviceID, index)
}

// GetUrmaDevEidListAll for A5
func (d *DeviceManager) GetUrmaDevEidListAll(logicID int32) ([]common.UrmaDeviceInfo, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForLogicIdErr, logicID,
			"failed to get cardId and deviceId by logicID(%d), error: %v", logicID, err)
		return nil, err
	}
	return d.DcMgr.DcGetUrmaDevEidListAll(cardID, deviceID)
}

// GetUrmaDeviceCount get urma device count by dcmiv2
func (d *DeviceManagerV2) GetUrmaDeviceCount(logicID int32) (int32, error) {
	return d.DcMgr.DcGetUrmaDeviceCount(logicID)
}

// GetUrmaDevEidList get urma device eid list with index by dcmiv2
func (d *DeviceManagerV2) GetUrmaDevEidList(logicID int32, index int32) (*common.UrmaDeviceInfo, error) {
	return d.DcMgr.DcGetUrmaDevEidList(logicID, index)
}

// GetUrmaDevEidListAll get urma device eid list all by dcmiv2
func (d *DeviceManagerV2) GetUrmaDevEidListAll(logicID int32) ([]common.UrmaDeviceInfo, error) {
	return d.DcMgr.DcGetUrmaDevEidListAll(logicID)
}
