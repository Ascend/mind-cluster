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

// Package device the device manager package
package device

import (
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

var dm *devmanager.DeviceManager

// InitDeviceManager init device manager
func InitDeviceManager() error {
	var err error
	dm, err = devmanager.GetDeviceManager()
	if err != nil {
		hwlog.RunLog.Errorf("init device manager failed:%v", err)
	}
	return err
}

// GetDeviceManager get device manager
func GetDeviceManager() *devmanager.DeviceManager {
	return dm
}
