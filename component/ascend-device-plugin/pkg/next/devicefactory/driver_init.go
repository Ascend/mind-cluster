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
	"errors"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/deviceswitch"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
)

// initDevManager inits the mandatory device manager, a non-nil error is fatal
func initDevManager() (devmanager.DeviceInterface, error) {
	devM, err := devmanager.AutoInit("", common.ParamOption.DeviceResetTimeout)
	if err != nil {
		hwlog.RunLog.Errorf("init devmanager failed, err: %v", err)
		return devM, err
	}
	return devM, nil
}

// initSwitchDevManager tries to init the optional switch fault capability.
// It never fails: a nil return means the capability is unavailable, and the reason has been
// logged once already, so callers only need to drop it and keep the rest of the plugin running.
//
// NOTE: only 910A5 is skipped here. Other device types still try to load the library on purpose,
// so the capability stays ready on any deployment that provides it. Do not narrow this condition
// to 910A3 without confirming that every consumer (server/manager.go and common/device.go) no
// longer needs the manager on non-A3 platforms.
func initSwitchDevManager(devM devmanager.DeviceInterface) *deviceswitch.SwitchDevManager {
	devType := devM.GetDevType()
	if devType == api.Ascend910A5 {
		hwlog.RunLog.Infof("switch fault detection: disabled (reason=unsupported-device, devType=%s)",
			utils.MaskDevType(devType))
		return nil
	}
	switchDevMgr := deviceswitch.NewSwitchDevManager()
	initErr := switchDevMgr.InitSwitchDev()
	switch {
	case initErr == nil:
		hwlog.RunLog.Infof("switch fault detection: initialized (devType=%s), feature only active on %s",
			utils.MaskDevType(devType), api.Ascend910A3)
		return switchDevMgr
	case errors.Is(initErr, deviceswitch.ErrSwitchLibNotFound):
		hwlog.RunLog.Warnf("switch fault detection: disabled (reason=lib-not-found, lib=%s), "+
			"switch faults will not be reported, err: %v", deviceswitch.SwitchLibName, initErr)
	case errors.Is(initErr, deviceswitch.ErrSwitchLibInitFailed):
		hwlog.RunLog.Errorf("switch fault detection: disabled (reason=init-failed, lib=%s), "+
			"switch faults will not be reported, err: %v", deviceswitch.SwitchLibName, initErr)
	default:
		hwlog.RunLog.Errorf("switch fault detection: disabled (reason=lib-load-failed, lib=%s), "+
			"switch faults will not be reported, err: %v", deviceswitch.SwitchLibName, initErr)
	}
	return nil
}
