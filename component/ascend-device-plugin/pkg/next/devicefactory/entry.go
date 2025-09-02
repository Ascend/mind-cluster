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

// Package devicefactory a series of entry function
package devicefactory

import (
	"fmt"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/server"
	"ascend-common/common-utils/hwlog"
)

// InitFunction init function
func InitFunction() (*server.HwDevManager, error) {
	initPublicNameConfig()
	devM, switchDevM, err := initDevManager()
	if err != nil {
		hwlog.RunLog.Errorf("init dev manager failed, err: %v", err)
		return nil, err
	}
	hdm := server.NewHwDevManager(devM)
	if hdm == nil {
		hwlog.RunLog.Error("init device manager failed")
		return nil, fmt.Errorf("init device manager failed")
	}
	hwlog.RunLog.Info("init device manager success")
	common.ParamOption.EnableSwitchFault = false
	if switchDevM != nil {
		hdm.SwitchDevManager = switchDevM
		common.ParamOption.EnableSwitchFault = true
	}
	return hdm, nil
}
