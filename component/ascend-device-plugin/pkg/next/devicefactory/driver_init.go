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

// Package devicefactory a series of driver manager init function
package devicefactory

import (
	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/deviceswitch"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

func initDevManager() (*devmanager.DeviceManager, *deviceswitch.SwitchDevManager, error) {
	devM, err := devmanager.AutoInit("", common.ParamOption.DeviceResetTimeout)
	if err != nil {
		hwlog.RunLog.Errorf("init devmanager failed, err: %v", err)
		return nil, nil, err
	}
	if devM.GetDevType() == api.Ascend910A5 {
		hwlog.RunLog.Infof("current devType is %s, switch device manager not supported.", api.HuaweiNPU)
		return devM, nil, nil
	}
	switchDevMgr := deviceswitch.NewSwitchDevManager()
	if err := switchDevMgr.InitSwitchDev(); err != nil {
		hwlog.RunLog.Warnf("failed to init switch device manager, will not deal with switch fault, "+
			"err: %s", err.Error())
		return devM, nil, nil
	}
	return devM, switchDevMgr, nil
}
