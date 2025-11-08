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

// Package devmanager this for device driver manager mock
package devmanager

import "ascend-common/devmanager/common"

// GetUrmaDeviceCount get urma device count for A5
func (d *DeviceManagerMock) GetUrmaDeviceCount(cardID int32, deviceID int32) (int32, error) {
	return 0, nil
}

// GetUrmaDevEidList get urma device index EID info for A5
func (d *DeviceManagerMock) GetUrmaDevEidList(cardID int32, deviceID int32, index int32) (*common.UrmaDeviceInfo,
	error) {
	return &common.UrmaDeviceInfo{}, nil
}

// GetUrmaDevEidListAll get urma device EID info for A5
func (d *DeviceManagerMock) GetUrmaDevEidListAll(cardID int32, deviceID int32) ([]common.UrmaDeviceInfo, error) {
	return []common.UrmaDeviceInfo{
		{
			EidCount: 1,
			EidInfos: []common.UrmaEidInfo{
				{Eid: common.Eid{Raw: [common.EidByteSize]byte{}}},
			},
		},
	}, nil
}
